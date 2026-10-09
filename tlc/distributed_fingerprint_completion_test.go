package tlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// No enabled source method directly checks FPSet.exit's completion diagnostic.
// Go keeps process lifetime with the native role owner rather than System.exit.
func TestDistributedFingerprintExitReportsCompletion(t *testing.T) {
	for _, backend := range []string{"mem", "mem1", "mem2", "lsb", "msb"} {
		for _, remote := range []bool{false, true} {
			for _, cleanup := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/remote=%v/cleanup=%v", backend, remote, cleanup), func(t *testing.T) {
					captureFailoverToolIO(t, ToolIOTool)
					wasRunning := distributedFPServerRunning.Load()
					t.Cleanup(func() { distributedFPServerRunning.Store(wasRunning) })
					hostname, err := distributedLocalHostName()
					if err != nil {
						t.Fatal(err)
					}
					recorder := &MemoryRecorder{}
					AddMessageRecorder(recorder)
					t.Cleanup(func() { RemoveMessageRecorder(recorder) })
					config := NewFPSetConfiguration()
					config.SetMemory(1 << 20)
					var store FPSet
					switch backend {
					case "mem":
						store = NewMemFPSet()
					case "mem1":
						store = NewMemFPSet1(config)
					case "mem2":
						store = &MemFPSet2{}
					case "lsb":
						store = NewLSBDiskFPSet(config).DiskFPSet
					case "msb":
						store = NewMSBDiskFPSet(config).DiskFPSet
					}
					directory := t.TempDir()
					store.Init(1, directory, "Spec")
					t.Cleanup(store.Close)
					artifact := filepath.Join(directory, "retained")
					if err := os.WriteFile(artifact, []byte("owned"), 0600); err != nil {
						t.Fatal(err)
					}
					var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(store)
					if remote {
						_, endpoint = startFingerprintRPC(t, endpoint)
					}
					if err := endpoint.Exit(cleanup); err != nil {
						t.Fatal(err)
					}
					records := recorder.Records(ECTLCFPCompleted)
					if len(records) != 1 || len(records[0].Params) != 1 || records[0].Params[0] != hostname {
						t.Fatalf("completion records = %v", records)
					}
					message := hostname + ", work completed. Thank you!"
					if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), message) != 1 {
						t.Fatal("missing exact source completion output")
					}
					if _, err := os.Stat(artifact); cleanup && !os.IsNotExist(err) || !cleanup && err != nil {
						t.Fatalf("cleanup/file retention = %v", err)
					}
					if distributedFPServerRunning.Load() {
						t.Fatal("exit did not signal native fingerprint role shutdown")
					}
				})
			}
		}
	}
}

func TestDistributedFingerprintExitContinuesAfterFailedDirectoryRemoval(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	wasRunning := distributedFPServerRunning.Load()
	t.Cleanup(func() { distributedFPServerRunning.Store(wasRunning) })
	directory := t.TempDir()
	t.Chdir(directory)
	// Removing literal "." fails natively, even in an empty directory. Source
	// FileUtil returns false here; the concrete exit ignores that boolean.
	if err := os.RemoveAll("."); err == nil {
		t.Fatal("fixture did not retain a failed directory-removal boundary")
	}
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() { RemoveMessageRecorder(recorder) })
	store := NewMemFPSet()
	store.Init(1, ".", "Spec")
	if err := store.Exit(true); err != nil {
		t.Fatalf("source-ignored directory removal became an exit failure: %v", err)
	}
	if len(recorder.Records(ECTLCFPCompleted)) != 1 || distributedFPServerRunning.Load() {
		t.Fatal("failed removal suppressed completion or native role shutdown")
	}
	if info, err := os.Stat(directory); err != nil || !info.IsDir() {
		t.Fatalf("failed removal lost its containing directory: %v", err)
	}
}
