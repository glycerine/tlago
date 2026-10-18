package tlc

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type gatedFingerprintBlock struct {
	*LocalFingerprintEndpoint
	entered, release, finished chan struct{}
}

func (e *gatedFingerprintBlock) gate() { close(e.entered); <-e.release }
func (e *gatedFingerprintBlock) PutBlock(fps *LongVec) (*BitVector, error) {
	e.gate()
	defer close(e.finished)
	return e.LocalFingerprintEndpoint.PutBlock(fps)
}
func (e *gatedFingerprintBlock) ContainsBlock(fps *LongVec) (*BitVector, error) {
	e.gate()
	defer close(e.finished)
	return e.LocalFingerprintEndpoint.ContainsBlock(fps)
}

// No enabled upstream test covers native in-flight connection loss. Preserve
// source callable reassignment and partition ordering over real TCP endpoints.
func TestFingerprintRPCInflightBlockFailover(t *testing.T) {
	for _, put := range []bool{false, true} {
		for _, concurrent := range []bool{false, true} {
			t.Run(fmt.Sprintf("put=%v/concurrent=%v", put, concurrent), func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				primary, survivor := NewMemFPSet(), NewMemFPSet()
				if !put {
					primary.Put(0)
					survivor.Put(1)
				}
				endpoint := &gatedFingerprintBlock{NewLocalFingerprintEndpoint(primary), make(chan struct{}), make(chan struct{}), make(chan struct{})}
				host, failed := startFingerprintRPC(t, endpoint)
				_, spare := startFingerprintRPC(t, NewLocalFingerprintEndpoint(survivor))
				manager := NewDistributedFPSetManager(failed, spare)
				manager.fpSets[0].hostname = "lost-primary"
				manager.fpSets[1].hostname = "survivor"
				var executors []*DistributedExecutor
				if concurrent {
					executor := NewDistributedExecutor()
					t.Cleanup(executor.Shutdown)
					executors = append(executors, executor)
				}
				fps := []*LongVec{NewLongVecFrom([]int64{0, 2}), NewLongVecFrom([]int64{1, 3})}
				done := make(chan []*BitVector, 1)
				jobFinished := make(chan struct{})
				var release sync.Once
				entered := false
				t.Cleanup(func() {
					_ = host.Close()
					release.Do(func() { close(endpoint.release) })
					<-jobFinished
					if entered {
						<-endpoint.finished
					}
				})
				go func() {
					defer close(jobFinished)
					if put {
						done <- manager.PutBlock(fps, executors...)
					} else {
						done <- manager.ContainsBlock(fps, executors...)
					}
				}()
				select {
				case <-endpoint.entered:
					entered = true
				case <-time.After(5 * time.Second):
					t.Fatal("in-flight block watchdog expired")
				}
				// This closes transport, not owned storage or the accepted handler. The
				// manager must recover before that handler is allowed to finish.
				if err := host.Close(); err != nil {
					t.Fatal(err)
				}
				var answers []*BitVector
				select {
				case answers = <-done:
				case <-time.After(5 * time.Second):
					t.Fatal("manager did not reassign the disconnected in-flight call")
				}
				if len(answers) != 2 || answers[0] == nil || answers[1] == nil {
					t.Fatalf("partition answers lost: %v", answers)
				}
				want := [][]bool{{true, true}, {false, true}}
				if put {
					want[1][0] = true
				}
				for partition, vector := range answers {
					got := []bool{vector.Get(0), vector.Get(1)}
					if !reflect.DeepEqual(got, want[partition]) || vector.Get(2) {
						t.Fatalf("partition %d = %v, want %v", partition, got, want[partition])
					}
				}
				if manager.NumOfAliveServers() != 1 || manager.entry(0) != manager.entry(1) || manager.entry(0).set != spare {
					t.Fatal("failed partition did not share the surviving registration wrapper")
				}
				messages := ToolIOGetAllMessages()
				if len(messages) != 1 || !strings.Contains(messages[0], "to the fp server at lost-primary.\n") {
					t.Fatalf("in-flight failure warning = %q", messages)
				}
				if put {
					if survivor.Size() != 4 || primary.Size() != 0 {
						t.Fatal("put replay did not target surviving storage before old handler completion")
					}
				} else if survivor.Size() != 1 || survivor.Contains(0) {
					t.Fatal("contains failover copied or inserted failed-store fingerprints")
				}
				release.Do(func() { close(endpoint.release) })
				<-endpoint.finished
				// Close is a transport operation; its old handler/storage remain owned
				// locally until joined. This is connection-loss coverage, not process kill.
				if put && primary.Size() != 2 {
					t.Fatal("transport close unexpectedly cancelled owned storage work")
				}
			})
		}
	}
}
