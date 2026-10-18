package tlc

import (
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type nestedCheckChild struct {
	*NoopFPSet
	id       int
	entered  chan int
	release  <-chan struct{}
	failure  error
	distance uint64
	valid    bool
}

func (s *nestedCheckChild) check() {
	if s.entered != nil {
		s.entered <- s.id
		<-s.release
	}
	if s.failure != nil {
		panic(s.failure)
	}
}
func (s *nestedCheckChild) CheckFPs() uint64              { s.check(); return s.distance }
func (s *nestedCheckChild) CheckInvariant(...uint64) bool { s.check(); return s.valid }

// No original MultiFPSet method checks the parallel check/failure boundary.
func TestDistributedNestedChecksConcurrentAndJoined(t *testing.T) {
	for _, operation := range []string{"fingerprints", "invariant"} {
		for _, mode := range []string{"success", "io_failure", "unchecked_failure"} {
			t.Run(operation+"/"+mode, func(t *testing.T) {
				entered := make(chan int, 2)
				release := make(chan struct{})
				ready := make(chan struct{})
				first := &nestedCheckChild{NoopFPSet: NewNoopFPSet(nil), id: 0, entered: entered, release: release, distance: 41, valid: true}
				second := &nestedCheckChild{NoopFPSet: NewNoopFPSet(nil), id: 1, entered: entered, release: ready, distance: uint64(1) << 63, valid: false}
				if mode == "io_failure" {
					second.failure = io.ErrUnexpectedEOF
				} else if mode == "unchecked_failure" {
					second.failure = errors.New("child check failed")
				}
				set := &MultiFPSet{Sets: []FPSet{first, second}}
				type result struct {
					distance uint64
					valid    bool
					failure  any
				}
				done := make(chan result, 1)
				joined := make(chan struct{})
				var unblock sync.Once
				var secondUnblock sync.Once
				t.Cleanup(func() {
					unblock.Do(func() { close(release) })
					secondUnblock.Do(func() { close(ready) })
					<-joined
				})
				go func() {
					defer close(joined)
					var r result
					defer func() { r.failure = recover(); done <- r }()
					if operation == "fingerprints" {
						r.distance = set.CheckFPs()
					} else {
						r.valid = set.CheckInvariant()
					}
				}()
				seen := map[int]bool{}
				for range 2 {
					select {
					case id := <-entered:
						seen[id] = true
					case <-time.After(time.Second):
						t.Fatal("child checks did not start concurrently")
					}
				}
				if !seen[0] || !seen[1] {
					t.Fatal("child identity lost")
				}
				secondUnblock.Do(func() { close(ready) })
				select {
				case r := <-done:
					t.Fatalf("check returned before held child finished: %+v", r)
				default:
				}
				unblock.Do(func() { close(release) })
				r := <-done
				if mode == "success" {
					if r.failure != nil || operation == "fingerprints" && r.distance != uint64(1)<<63 || operation == "invariant" && r.valid {
						t.Fatalf("reduction = %+v", r)
					}
				} else if mode == "unchecked_failure" {
					if r.failure != second.failure {
						t.Fatalf("unchecked failure = %v", r.failure)
					}
				} else {
					err, ok := r.failure.(error)
					if !ok || !errors.Is(err, second.failure) || isJavaIOException(err) {
						t.Fatalf("child I/O failure lost wrapping/cause: %v", r.failure)
					}
				}
			})
		}
	}
	if set := (&MultiFPSet{}); set.CheckFPs() != math.MaxInt64 || !set.CheckInvariant() {
		t.Fatal("empty check reduction changed")
	}
}

func TestDistributedNestedCheckIOFailureOverTCP(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	output, err := os.CreateTemp(t.TempDir(), "check-stderr-")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stderr
	os.Stderr = output
	t.Cleanup(func() { os.Stderr = previous; _ = output.Close() })
	child := &nestedCheckChild{NoopFPSet: NewNoopFPSet(nil), failure: io.ErrUnexpectedEOF}
	_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(&MultiFPSet{Sets: []FPSet{child}}))
	if _, err := client.CheckInvariant(); err == nil || isJavaIOException(err) {
		t.Fatalf("nested check exposed ignorable I/O: %v", err)
	}
	manager := NewDistributedFPSetManager(client)
	if manager.CheckFPs() != math.MaxInt64 || !manager.CheckInvariant() {
		t.Fatal("nested failure entered the callable I/O fallback")
	}
	if messages := ToolIOGetAllMessages(); len(messages) != 0 {
		t.Fatalf("nested failure emitted GENERAL: %v", messages)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil || strings.Count(string(data), "java.util.concurrent.ExecutionException:") != 2 || !strings.Contains(string(data), "unexpected EOF") {
		t.Fatalf("failed completion diagnostics: %s/%v", data, err)
	}
}
