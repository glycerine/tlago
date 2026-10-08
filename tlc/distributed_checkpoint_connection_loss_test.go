package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type gatedFingerprintCheckpoint struct {
	*LocalFingerprintEndpoint
	phase                       string
	entered, release, finished  chan struct{}
	begin, commit, recoverCalls atomic.Int32
}

func (e *gatedFingerprintCheckpoint) run(phase string, operation func() error) error {
	if phase == e.phase {
		close(e.entered)
		defer close(e.finished)
		<-e.release
	}
	return operation()
}

func (e *gatedFingerprintCheckpoint) BeginChkptFile(name string) error {
	e.begin.Add(1)
	return e.run("begin", func() error { return e.LocalFingerprintEndpoint.BeginChkptFile(name) })
}
func (e *gatedFingerprintCheckpoint) CommitChkptFile(name string) error {
	e.commit.Add(1)
	return e.run("commit", func() error { return e.LocalFingerprintEndpoint.CommitChkptFile(name) })
}
func (e *gatedFingerprintCheckpoint) RecoverFile(name string) error {
	e.recoverCalls.Add(1)
	return e.run("recover", func() error { return e.LocalFingerprintEndpoint.RecoverFile(name) })
}

// No enabled upstream test covers connection loss during an accepted checkpoint
// call. FPSetManager.Checkpoint.run catches I/O failures and continues without
// reassignment. Closing native transport must not cancel owned storage work.
func TestFingerprintRPCInflightCheckpointConnectionLoss(t *testing.T) {
	for _, phase := range []string{"begin", "commit", "recover"} {
		t.Run(phase, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			oldWorkers := NumWorkers()
			SetNumWorkers(0)
			t.Cleanup(func() { SetNumWorkers(oldWorkers) })
			primaryDir, healthyDir := t.TempDir(), t.TempDir()
			primary, healthy := NewMemFPSet(), NewMemFPSet()
			primary.Init(1, primaryDir, "Spec")
			healthy.Init(1, healthyDir, "Spec")
			primary.Put(61)
			healthy.Put(72)
			if phase == "recover" {
				for _, set := range []*MemFPSet{primary, healthy} {
					if err := set.BeginChkptFile("Spec"); err != nil {
						t.Fatal(err)
					}
					if err := set.CommitChkptFile("Spec"); err != nil {
						t.Fatal(err)
					}
				}
				primary, healthy = NewMemFPSet(), NewMemFPSet()
				primary.Init(1, primaryDir, "Spec")
				healthy.Init(1, healthyDir, "Spec")
			}
			endpoint := &gatedFingerprintCheckpoint{
				LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(primary), phase: phase,
				entered: make(chan struct{}), release: make(chan struct{}), finished: make(chan struct{}),
			}
			host, failed := startFingerprintRPC(t, endpoint)
			_, available := startFingerprintRPC(t, NewLocalFingerprintEndpoint(healthy))
			manager := NewDistributedFPSetManager(failed, available)
			manager.fpSets[0].hostname = "disconnected-checkpoint"
			first, second := manager.entry(0), manager.entry(1)
			metadir := t.TempDir()
			queue := NewDiskStateQueue(metadir)
			t.Cleanup(queue.FinishAll)
			trace := NewTLCTrace(metadir, "Spec")
			t.Cleanup(func() { _ = trace.Close() })
			server := NewTLCServer("Spec", "Spec", metadir, manager, queue, trace)
			done := make(chan error, 1)
			joined := make(chan struct{})
			var release sync.Once
			entered := false
			t.Cleanup(func() {
				_ = host.Close()
				release.Do(func() { close(endpoint.release) })
				<-joined
				if entered {
					<-endpoint.finished
				}
			})
			go func() {
				defer close(joined)
				if phase == "recover" {
					done <- manager.Recover("Spec")
				} else {
					done <- server.Checkpoint()
				}
			}()
			select {
			case <-endpoint.entered:
				entered = true
			case <-time.After(5 * time.Second):
				t.Fatal("checkpoint handler did not reach gate")
			}
			if err := host.Close(); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("source I/O catch did not continue: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("checkpoint did not continue after connection loss")
			}
			if manager.NumOfAliveServers() != 2 || manager.entry(0) != first || manager.entry(1) != second {
				t.Fatal("checkpoint failure reassigned or marked a fingerprint registration unavailable")
			}
			warning := "Error: Failed to checkpoint the fingerprint server at disconnected-checkpoint. This server might be down."
			warnings := 0
			for _, message := range ToolIOGetAllMessages() {
				if strings.Contains(message, warning) {
					warnings++
				}
			}
			if warnings != 1 {
				t.Fatalf("checkpoint warning count %d, want 1", warnings)
			}
			if phase == "recover" {
				if primary.Size() != 0 || healthy.Size() != 1 || !healthy.Contains(72) || endpoint.recoverCalls.Load() != 1 {
					t.Fatal("recovery did not continue to the healthy store before old handler completed")
				}
			} else {
				wantCommit := int32(0)
				if phase == "commit" {
					wantCommit = 1
				}
				if endpoint.begin.Load() != 1 || endpoint.commit.Load() != wantCommit {
					t.Fatal("checkpoint retried the disconnected request or committed after failed begin")
				}
				queue.mu.Lock()
				stopped := queue.stop
				queue.mu.Unlock()
				if stopped {
					t.Fatal("coordinator did not resume its queue after caught fingerprint I/O failure")
				}
				recoveredQueue := NewDiskStateQueue(metadir)
				t.Cleanup(recoveredQueue.FinishAll)
				if err := recoveredQueue.Recover(); err != nil {
					t.Fatalf("coordinator queue checkpoint did not commit: %v", err)
				}
				if recoveredQueue.Size() != 0 {
					t.Fatal("empty coordinator frontier changed during checkpoint")
				}
				recoveredTrace := NewTLCTrace(metadir, "Spec")
				t.Cleanup(func() { _ = recoveredTrace.Close() })
				if err := recoveredTrace.Recover(); err != nil {
					t.Fatalf("coordinator trace checkpoint did not commit: %v", err)
				}
				reopened := NewMemFPSet()
				reopened.Init(1, healthyDir, "Spec")
				if err := reopened.RecoverFile("Spec"); err != nil {
					t.Fatal(err)
				}
				if reopened.Size() != 1 || !reopened.Contains(72) || reopened.Contains(61) {
					t.Fatal("healthy partition checkpoint was lost or merged with disconnected partition")
				}
				if _, err := os.Stat(filepath.Join(primaryDir, "Spec.fp.chkpt")); !os.IsNotExist(err) {
					t.Fatalf("paused primary unexpectedly has a committed checkpoint: %v", err)
				}
			}
			release.Do(func() { close(endpoint.release) })
			<-endpoint.finished
			if phase == "recover" {
				if primary.Size() != 1 || !primary.Contains(61) {
					t.Fatal("transport closure cancelled accepted recovery")
				}
			} else {
				path := filepath.Join(primaryDir, "Spec.fp.tmp")
				if phase == "commit" {
					path = filepath.Join(primaryDir, "Spec.fp.chkpt")
				}
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("accepted storage operation did not finish: %v", err)
				}
			}
		})
	}
}
