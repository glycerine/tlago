package tlc

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// NonCheckpointableDiskFPSet has no original Java test. Its named operations
// warn and return without saving/restoring membership, including through a
// remote manager whose later checkpoint-capable registration must still run.
func TestNativeRemoteOffHeapNamedCheckpointRetainsSourceLimitation(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	t.Setenv("TLAGO_MAX_DIRECT_MEMORY", "256k")
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() { RemoveMessageRecorder(recorder) })
	offHeapDirectory, memoryDirectory := t.TempDir(), t.TempDir()
	newGeneration := func() (*MultiFPSet, FPSet, *DistributedFPSetManager, func()) {
		config := NewFPSetConfigurationWithRatioAndImplementation(1, "tlc2.tool.fp.OffHeapDiskFPSet")
		config.SetFPBits(1)
		offHeap := NewFPSet(config).Init(1, offHeapDirectory, "Spec").(*MultiFPSet)
		memory := NewMemFPSet().Init(1, memoryDirectory, "Spec")
		t.Cleanup(offHeap.Close)
		t.Cleanup(memory.Close)
		host0, client0 := startFingerprintRPC(t, NewLocalFingerprintEndpoint(offHeap))
		host1, client1 := startFingerprintRPC(t, NewLocalFingerprintEndpoint(memory))
		var once sync.Once
		closeGeneration := func() {
			once.Do(func() {
				for _, client := range []*NetworkFingerprintEndpoint{client0, client1} {
					if err := client.CloseConnection(); err != nil {
						t.Error(err)
					}
				}
				for _, host := range []*DistributedRPCServer{host0, host1} {
					if err := host.CloseGracefully(); err != nil {
						t.Error(err)
					}
				}
				offHeap.Close()
				memory.Close()
			})
		}
		t.Cleanup(closeGeneration)
		return offHeap, memory, NewDistributedFPSetManager(client0, client1), closeGeneration
	}
	offHeap, memory, manager, closeGeneration := newGeneration()
	fingerprints := []uint64{41, uint64(1)<<63 | 43}
	for _, fp := range fingerprints {
		if offHeap.Put(fp) {
			t.Fatal("new off-heap fingerprint reported present")
		}
	}
	memory.Put(97)
	if err := manager.Checkpoint("job"); err != nil {
		t.Fatal(err)
	}
	if offHeap.Size() != 2 || memory.Size() != 1 || len(offHeap.Sets) != 2 {
		t.Fatal("named checkpoint changed live membership or nesting")
	}
	for _, child := range offHeap.Sets {
		if _, ok := child.(*OffHeapDiskFPSet); !ok {
			t.Fatalf("unexpected physical child %T", child)
		}
	}
	entries, err := os.ReadDir(offHeapDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name() != "Spec_0.fp" || entries[1].Name() != "Spec_1.fp" {
		t.Fatalf("unsupported checkpoint created snapshot artifacts: %v", entries)
	}
	if snapshot, err := os.ReadFile(filepath.Join(memoryDirectory, "job.fp.chkpt")); err != nil || len(snapshot) != 8 {
		t.Fatalf("later memory registration did not commit its actual snapshot: %v/%v", snapshot, err)
	}
	closeGeneration()
	offHeap, memory, manager, _ = newGeneration()
	if offHeap.Size() != 0 || memory.Size() != 0 {
		t.Fatal("fresh storage retained old process membership")
	}
	if err := manager.Recover("job"); err != nil {
		t.Fatal(err)
	}
	if offHeap.Size() != 0 || memory.Size() != 1 || !memory.Contains(97) || manager.Size() != 1 {
		t.Fatal("named recovery invented off-heap membership or skipped the healthy store")
	}
	for _, fp := range fingerprints {
		if offHeap.Contains(fp) {
			t.Fatal("unsupported named recovery restored an off-heap fingerprint")
		}
	}
	warnings := recorder.Records(ECGeneral)
	if len(warnings) != 6 { // Two children, each receiving begin, commit and recover.
		t.Fatalf("source named-operation warnings = %d, want 6", len(warnings))
	}
	for _, warning := range warnings {
		if warning.Severity != SeverityWarning || len(warning.Params) != 1 || warning.Params[0] != "Checkpointing is not implemented for tlc2.tool.fp.OffHeapDiskFPSet" {
			t.Fatalf("source warning changed: %+v", warning)
		}
	}
	if manager.NumOfAliveServers() != 2 {
		t.Fatal("unsupported named operations triggered fingerprint failover")
	}
}
