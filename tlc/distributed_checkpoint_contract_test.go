package tlc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// No enabled original manager test exercises Checkpoint.run's IOException
// catch. Preserve its synchronous ordering and unchecked-failure boundary.
type checkpointContractEndpoint struct {
	*LocalFingerprintEndpoint
	name, phase string
	calls       *[]string
	failure     error
	panics      bool
}

func (e *checkpointContractEndpoint) operation(phase string) error {
	*e.calls = append(*e.calls, e.name+"."+phase)
	if phase == e.phase {
		if e.panics {
			panic(e.failure)
		}
		return e.failure
	}
	return nil
}
func (e *checkpointContractEndpoint) BeginChkptFile(string) error  { return e.operation("begin") }
func (e *checkpointContractEndpoint) CommitChkptFile(string) error { return e.operation("commit") }
func (e *checkpointContractEndpoint) RecoverFile(string) error     { return e.operation("recover") }

func TestDistributedCheckpointCatchesOnlyIOFailures(t *testing.T) {
	for _, phase := range []string{"begin", "commit", "recover"} {
		for _, ioFailure := range []bool{true, false} {
			for _, panics := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/io=%v/panic=%v", phase, ioFailure, panics), func(t *testing.T) {
					var output bytes.Buffer
					oldMode := ToolIOGetMode()
					ToolIOSetMode(ToolIOSystem)
					defer ToolIOSetMode(oldMode)
					defer ToolIOSetSystemStreams(&output, &output)()
					var calls []string
					var failure error = NewIllegalArgumentException("bad checkpoint")
					if ioFailure {
						failure = NewIOException("disk unavailable")
					}
					first := &checkpointContractEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), "first", phase, &calls, failure, panics}
					second := &checkpointContractEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), name: "second", calls: &calls}
					manager := NewDistributedFPSetManager(first, second)
					manager.fpSets[0].hostname = "first-host"
					var err error
					if phase == "recover" {
						err = manager.Recover("job")
					} else {
						err = manager.Checkpoint("job")
					}
					want := []string{"first.begin"}
					if phase == "commit" {
						want = append(want, "first.commit")
					}
					if phase == "recover" {
						want = []string{"first.recover"}
					}
					if ioFailure {
						if phase == "recover" {
							want = append(want, "second.recover")
						} else {
							want = append(want, "second.begin", "second.commit")
						}
						if err != nil || output.String() != "Error: Failed to checkpoint the fingerprint server at first-host. This server might be down.\n" {
							t.Fatalf("error %v, ToolIO %q", err, output.String())
						}
					} else if err != failure || output.Len() != 0 {
						t.Fatalf("unchecked failure swallowed: %v, output %q", err, output.String())
					}
					if !reflect.DeepEqual(calls, want) {
						t.Fatalf("calls %v, want %v", calls, want)
					}
				})
			}
		}
	}
}

func TestDistributedCheckpointPreservesFatalFailure(t *testing.T) {
	var calls []string
	failure := NewAssertionError("fatal checkpoint")
	endpoint := &checkpointContractEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), "first", "begin", &calls, failure, true}
	defer func() {
		if got := recover(); got != failure {
			t.Fatalf("fatal failure %v, want %v", got, failure)
		}
	}()
	_ = NewDistributedFPSetManager(endpoint).Checkpoint("job")
}

func TestMemCheckpointCommitIOAndTruncatedRecovery(t *testing.T) {
	directory := t.TempDir()
	for _, set := range []FPSet{NewMemFPSet().Init(1, directory, "mem"), (&MemFPSet1{metadir: directory}), (&MemFPSet2{metadir: directory})} {
		if err := set.CommitChkptFile("missing"); !isJavaIOException(err) || !strings.Contains(err.Error(), "cannot delete") {
			t.Fatalf("commit failure lost source IO category: %v", err)
		}
	}
	for length := 0; length < 8; length++ {
		path := filepath.Join(directory, "job.fp.chkpt")
		if err := os.WriteFile(path, make([]byte, length), 0600); err != nil {
			t.Fatal(err)
		}
		base := NewMemFPSet()
		base.Init(1, directory, "base")
		err := base.RecoverFile("job")
		if length == 0 {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		failure, ok := err.(*TLCError)
		if !ok || failure.Code != ECSystemDiskIOErrorForFile || !failure.Runtime || !reflect.DeepEqual(failure.Params, []string{"checkpoints"}) {
			t.Fatalf("partial record %d accepted or misclassified: %v", length, err)
		}
		// MemFPSet2's original EOF catch instead throws IOException. No table
		// is needed because a partial record fails before insertion.
		if err := (&MemFPSet2{metadir: directory}).RecoverFile("job"); !isJavaIOException(err) || !strings.Contains(err.Error(), "MemFPSet2.recover: failed.") {
			t.Fatalf("MemFPSet2 partial record %d: %v", length, err)
		}
	}
}

func TestFingerprintRPCCorruptCheckpointStopsRecovery(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "job.fp.chkpt"), []byte{1}, 0600); err != nil {
		t.Fatal(err)
	}
	storage := NewMemFPSet()
	storage.Init(1, directory, "primary")
	_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
	manager := NewDistributedFPSetManager(client)
	if err := manager.Recover("job"); err == nil || isJavaIOException(err) {
		t.Fatalf("corrupt checkpoint was accepted or swallowed as a down server: %v", err)
	}
}

// No enabled original Java method checks this duplicate-file RPC boundary.
func TestFingerprintRPCDuplicateCheckpointRemainsRuntimeFailure(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	directory := t.TempDir()
	var data [24]byte
	for i, fp := range []uint64{41, 41, 97} {
		binary.BigEndian.PutUint64(data[i*8:], fp)
	}
	if err := os.WriteFile(filepath.Join(directory, "job.fp.chkpt"), data[:], 0600); err != nil {
		t.Fatal(err)
	}
	storage := NewMemFPSet()
	storage.Init(1, directory, "primary")
	_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
	manager := NewDistributedFPSetManager(client)
	err := manager.Recover("job")
	failure, ok := err.(*DistributedOperationError)
	if !ok || failure.Class != "util.Assert$TLCRuntimeException" || failure.Error() != "The fingerprint is not in set." || failure.IO || failure.Remote || failure.Recoverable {
		t.Fatalf("duplicate checkpoint runtime failure = %#v", err)
	}
	if storage.Size() != 1 || !storage.Contains(41) || storage.Contains(97) {
		t.Fatal("duplicate recovery did not retain partial insertion and stop")
	}
	if len(ToolIOGetAllMessages()) != 0 {
		t.Fatalf("runtime failure was swallowed as unavailable storage: %v", ToolIOGetAllMessages())
	}
}
