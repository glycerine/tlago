package tlc

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

func TestWorkerTraceFilenamePreservesEmptyComponents(t *testing.T) {
	separator := string(os.PathSeparator)
	for _, inputs := range [][2]string{{"", "Spec"}, {"metadata", ""}, {"", ""}, {"metadata/.", "Spec"}} {
		want := inputs[0] + separator + inputs[1] + "-2"
		if got := workerTraceFileBase(inputs[0], inputs[1], 2); got != want {
			t.Errorf("filename %q, want %q", got, want)
		}
	}
	worker := NewWorker(2)
	worker.Checker = &ModelChecker{AbstractChecker: &AbstractChecker{}}
	worker.Tool = NewTool()
	worker.Tool.RootName = "Explicit"
	worker.configureTrace()
	if worker.traceFileBase != separator+"Explicit-2" {
		t.Fatalf("empty directory skipped configuration: %q", worker.traceFileBase)
	}
}

func TestWorkerTraceOpeningRequiresContext(t *testing.T) {
	defer func() {
		if _, ok := recover().(*NullPointerException); !ok {
			t.Fatal("unconfigured trace opening reported success")
		}
	}()
	NewWorker(0).ensureTraceRAF()
}

// Isolate literal .tmp/.chkpt paths; no filesystem-root filename is opened.
func TestWorkerCheckpointUsesCachedEmptyFilename(t *testing.T) {
	for _, scenario := range []string{"begin", "begin-closed", "commit", "commit-missing", "recover", "recover-missing-owner"} {
		t.Run(scenario, func(t *testing.T) {
			t.Chdir(t.TempDir())
			worker := NewWorker(0)
			if scenario != "recover-missing-owner" {
				owner, err := NewBufferedRandomAccessFile("owner.st", "rw")
				if err != nil {
					t.Fatal(err)
				}
				worker.traceRAF = owner
				defer worker.CloseTrace()
				if _, err := owner.Write(make([]byte, 13)); err != nil {
					t.Fatal(err)
				}
			}
			prior := []byte("prior checkpoint")
			metadata := make([]byte, 16)
			binary.BigEndian.PutUint64(metadata, 3)
			binary.BigEndian.PutUint64(metadata[8:], 91)
			if err := os.WriteFile(".chkpt", metadata, 0600); err != nil {
				t.Fatal(err)
			}
			if scenario != "commit-missing" {
				if err := os.WriteFile(".tmp", prior, 0600); err != nil {
					t.Fatal(err)
				}
			}
			worker.lastPtr = 73
			switch scenario {
			case "begin", "begin-closed":
				if scenario == "begin-closed" {
					if err := worker.CloseTrace(); err != nil {
						t.Fatal(err)
					}
				}
				err := invokeDistributedServerOperation(worker.BeginChkpt)
				got, readErr := os.ReadFile(".tmp")
				if scenario == "begin-closed" {
					if !isJavaIOException(err) || readErr != nil || !bytes.Equal(got, prior) {
						t.Fatalf("flush failure skipped: %v, temporary %x/%v", err, got, readErr)
					}
				} else {
					want := make([]byte, 16)
					binary.BigEndian.PutUint64(want, 13)
					binary.BigEndian.PutUint64(want[8:], 73)
					if err != nil || readErr != nil || !bytes.Equal(got, want) {
						t.Fatalf("empty filename begin skipped: %v, temporary %x/%v", err, got, readErr)
					}
				}
			case "commit", "commit-missing":
				err := worker.CommitChkpt()
				got, readErr := os.ReadFile(".chkpt")
				if scenario == "commit-missing" {
					if !isJavaIOException(err) || !os.IsNotExist(readErr) {
						t.Fatalf("commit did not delete before promotion failure: %v/%v", err, readErr)
					}
				} else if err != nil || readErr != nil || !bytes.Equal(got, prior) {
					t.Fatalf("empty filename commit skipped: %v, checkpoint %x/%v", err, got, readErr)
				}
			case "recover", "recover-missing-owner":
				err := invokeDistributedServerOperation(worker.RecoverTrace)
				if worker.lastPtr != 91 || worker.traceFileBase != "" {
					t.Fatalf("cached recovery context changed or skipped: %d/%q", worker.lastPtr, worker.traceFileBase)
				}
				if scenario == "recover-missing-owner" {
					if _, ok := err.(*NullPointerException); !ok {
						t.Fatalf("missing recovery owner: %T/%v", err, err)
					}
				} else {
					pos, posErr := worker.traceRAF.GetFilePointer()
					if err != nil || posErr != nil || pos != 3 {
						t.Fatalf("empty filename recovery skipped seek: %v, pointer %d/%v", err, pos, posErr)
					}
				}
			}
		})
	}
}
