package tlc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
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
	for _, backend := range []string{"mem", "lsb", "msb"} {
		for _, phase := range []string{"begin", "commit", "recover"} {
			t.Run(backend+"/"+phase, func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				oldWorkers := NumWorkers()
				SetNumWorkers(0)
				t.Cleanup(func() { SetNumWorkers(oldWorkers) })
				primaryDir, healthyDir := t.TempDir(), t.TempDir()
				newStore := func(directory string) FPSet {
					var set FPSet
					if backend == "mem" {
						set = NewMemFPSet()
					} else {
						implementation := "tlc2.tool.fp.LSBDiskFPSet"
						if backend == "msb" {
							implementation = "tlc2.tool.fp.MSBDiskFPSet"
						}
						config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
						config.SetFPBits(1)
						config.SetMemory(1 << 20)
						set = NewFPSet(config)
					}
					set.Init(1, directory, "Spec")
					t.Cleanup(set.Close)
					return set
				}
				primary, healthy := newStore(primaryDir), newStore(healthyDir)
				primary.Put(61)
				healthy.Put(72)
				children, count := 1, uint64(1)
				if backend != "mem" {
					children, count = 2, 2
					primary.Put(uint64(1)<<63 | 63)
					healthy.Put(uint64(1)<<63 | 74)
				}
				checkMembership := func(set FPSet, first, second uint64) {
					t.Helper()
					if set.Size() != count || !set.Contains(first) || set.Contains(97) {
						t.Fatal("checkpoint lost exact store membership")
					}
					if children == 2 && !set.Contains(uint64(1)<<63|second) {
						t.Fatal("checkpoint lost upper partition membership")
					}
				}
				checkpointPath := func(child int, extension string) string {
					name := "Spec"
					if children == 2 {
						name += "_" + strconv.Itoa(child)
					}
					return filepath.Join(primaryDir, name+".fp."+extension)
				}
				if phase == "recover" {
					for _, set := range []FPSet{primary, healthy} {
						if err := set.BeginChkptFile("Spec"); err != nil {
							t.Fatal(err)
						}
						if err := set.CommitChkptFile("Spec"); err != nil {
							t.Fatal(err)
						}
					}
					primary.Close()
					healthy.Close()
					primary, healthy = newStore(primaryDir), newStore(healthyDir)
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
					if primary.Size() != 0 || endpoint.recoverCalls.Load() != 1 {
						t.Fatal("recovery did not continue to the healthy store before old handler completed")
					}
					checkMembership(healthy, 72, 74)
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
					healthy.Close()
					reopened := newStore(healthyDir)
					if err := reopened.RecoverFile("Spec"); err != nil {
						t.Fatal(err)
					}
					checkMembership(reopened, 72, 74)
					if reopened.Contains(61) {
						t.Fatal("healthy partition checkpoint was lost or merged with disconnected partition")
					}
					for child := 0; child < children; child++ {
						if _, err := os.Stat(checkpointPath(child, "chkpt")); !os.IsNotExist(err) {
							t.Fatalf("paused primary unexpectedly has a committed checkpoint: %v", err)
						}
					}
				}
				release.Do(func() { close(endpoint.release) })
				<-endpoint.finished
				if phase == "recover" {
					checkMembership(primary, 61, 63)
				} else {
					extension := "tmp"
					if phase == "commit" {
						extension = "chkpt"
					}
					for child := 0; child < children; child++ {
						data, err := os.ReadFile(checkpointPath(child, extension))
						want := []uint64{61, 63}[child]
						if err != nil || len(data) != 8 || binary.BigEndian.Uint64(data) != want {
							t.Fatalf("accepted storage operation lost partition %d snapshot: %x/%v", child, data, err)
						}
						if phase == "begin" {
							if _, err := os.Stat(checkpointPath(child, "chkpt")); !os.IsNotExist(err) {
								t.Fatal("accepted begin promoted pending snapshot", err)
							}
						} else if _, err := os.Stat(checkpointPath(child, "tmp")); !os.IsNotExist(err) {
							t.Fatal("accepted commit did not consume pending snapshot", err)
						}
					}
					if phase == "commit" {
						primary.Close()
						reopened := newStore(primaryDir)
						if err := reopened.RecoverFile("Spec"); err != nil {
							t.Fatal("accepted commit did not produce a recoverable snapshot", err)
						}
						checkMembership(reopened, 61, 63)
					}
				}
			})
		}
	}
}
