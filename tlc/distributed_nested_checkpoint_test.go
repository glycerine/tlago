package tlc

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
)

// MultiFPSetTest has no named checkpoint methods. These native checks cover
// the parallel traversal and failure boundary used by remote fingerprint hosts.
type nestedCheckpointChild struct {
	*MemFPSet
	entered chan string
	release <-chan struct{}
	failure error
	panics  bool
}

func (s *nestedCheckpointChild) checkpoint(phase, name string) error {
	s.entered <- phase + ":" + name
	<-s.release
	if s.panics {
		panic(s.failure)
	}
	return s.failure
}
func (s *nestedCheckpointChild) BeginChkptFile(name string) error { return s.checkpoint("begin", name) }
func (s *nestedCheckpointChild) CommitChkptFile(name string) error {
	return s.checkpoint("commit", name)
}
func (s *nestedCheckpointChild) RecoverFile(name string) error { return s.checkpoint("recover", name) }

func TestDistributedNestedNamedCheckpointConcurrentAndJoined(t *testing.T) {
	for _, phase := range []string{"begin", "commit", "recover"} {
		for _, mode := range []string{"success", "io", "panic-io", "panic"} {
			t.Run(phase+"/"+mode, func(t *testing.T) {
				entered := make(chan string, 2)
				release := make(chan struct{})
				defer close(release)
				first := &nestedCheckpointChild{MemFPSet: NewMemFPSet(), entered: entered, release: release}
				secondRelease := make(chan struct{})
				close(secondRelease)
				second := &nestedCheckpointChild{MemFPSet: NewMemFPSet(), entered: entered, release: secondRelease}
				if mode != "success" {
					second.failure = io.ErrUnexpectedEOF
				}
				if mode == "panic" {
					second.failure = errors.New("child failed")
				}
				second.panics = mode == "panic" || mode == "panic-io"
				set := &MultiFPSet{Sets: []FPSet{first, second}}
				type result struct {
					err        error
					panicValue any
				}
				done := make(chan result, 1)
				go func() {
					var r result
					defer func() { r.panicValue = recover(); done <- r }()
					switch phase {
					case "begin":
						r.err = set.BeginChkptFile("job")
					case "commit":
						r.err = set.CommitChkptFile("job")
					case "recover":
						r.err = set.RecoverFile("job")
					}
				}()
				seen := map[string]bool{}
				for i := 0; i < 2; i++ {
					select {
					case name := <-entered:
						seen[name] = true
					case <-time.After(5 * time.Second):
						t.Fatal("children did not enter concurrently")
					}
				}
				for i := 0; i < 2; i++ {
					if !seen[fmt.Sprintf("%s:job_%d", phase, i)] {
						t.Fatalf("child filenames: %v", seen)
					}
				}
				select {
				case r := <-done:
					t.Fatalf("returned before held child completed: %+v", r)
				default:
				}
				release <- struct{}{}
				select {
				case r := <-done:
					if mode == "panic" {
						if r.panicValue != second.failure {
							t.Fatalf("panic = %v", r.panicValue)
						}
					} else if r.panicValue != nil {
						t.Fatalf("unexpected panic: %v", r.panicValue)
					} else if mode == "success" {
						if r.err != nil {
							t.Fatal(r.err)
						}
					} else if !errors.Is(r.err, second.failure) || isJavaIOException(r.err) {
						t.Fatalf("lost cause or classified operation failure as ignorable IO: %v", r.err)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("children were not joined")
				}
			})
		}
	}
}

func TestDistributedNestedNamedCheckpointStorageRecovery(t *testing.T) {
	for _, implementation := range []string{"tlc2.tool.fp.MemFPSet", "tlc2.tool.fp.LSBDiskFPSet", "tlc2.tool.fp.MSBDiskFPSet"} {
		t.Run(implementation, func(t *testing.T) {
			config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
			config.SetMemory(1 << 20)
			config.SetFPBits(1)
			directory := t.TempDir()
			set := NewMultiFPSet(config)
			set.Init(1, directory, "Spec")
			defer func() { set.Close() }()
			fingerprints := []uint64{41, uint64(1)<<63 | 43}
			for _, fp := range fingerprints {
				if set.Put(fp) {
					t.Fatal("new fingerprint reported present")
				}
			}
			if err := set.BeginChkptFile("job"); err != nil {
				t.Fatal(err)
			}
			if err := set.CommitChkptFile("job"); err != nil {
				t.Fatal(err)
			}
			if set.Put(97) {
				t.Fatal("pending fingerprint reported present")
			}
			// Source recovery populates a fresh initialized store; it does not
			// clear the old live table before replaying committed membership.
			set.Close()
			set = NewMultiFPSet(config)
			set.Init(1, directory, "Spec")
			if err := set.RecoverFile("job"); err != nil {
				t.Fatal(err)
			}
			if set.Size() != 2 || set.Sets[0].Size() != 1 || set.Sets[1].Size() != 1 || set.Contains(97) {
				t.Fatal("recovery changed partition membership or retained pending insert")
			}
			for _, fp := range fingerprints {
				if !set.Contains(fp) {
					t.Fatalf("lost fingerprint %d", fp)
				}
			}
			manager := NewDistributedFPSetManagerFromFPSet(set)
			if err := manager.Recover("missing"); err == nil || isJavaIOException(err) {
				t.Fatalf("nested recovery failure was ignored or classified as remote outage: %v", err)
			}
		})
	}
}
