package tlc

import "testing"

// No original method tests an absent local recovery trace. Disk and nested
// stores replay that trace; MemFPSet instead reads its named checkpoint.
func TestDistributedFingerprintRecoveryRequiresTraceForReplayStores(t *testing.T) {
	for _, test := range []struct {
		name, implementation string
		bits                 int
		needsTrace           bool
	}{
		{"mem", "tlc2.tool.fp.MemFPSet", 0, false},
		{"lsb", "tlc2.tool.fp.LSBDiskFPSet", 0, true},
		{"msb", "tlc2.tool.fp.MSBDiskFPSet", 0, true},
		{"nested-mem", "tlc2.tool.fp.MemFPSet", 1, true},
		{"nested-lsb", "tlc2.tool.fp.LSBDiskFPSet", 1, true},
		{"nested-msb", "tlc2.tool.fp.MSBDiskFPSet", 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			newStore := func() FPSet {
				config := NewFPSetConfigurationWithRatioAndImplementation(1, test.implementation)
				config.SetMemory(1 << 20)
				config.SetFPBits(test.bits)
				var store FPSet
				if test.bits == 0 && test.implementation == "tlc2.tool.fp.LSBDiskFPSet" {
					store = NewLSBDiskFPSet(config)
				} else if test.bits == 0 && test.implementation == "tlc2.tool.fp.MSBDiskFPSet" {
					store = NewMSBDiskFPSet(config)
				} else {
					store = NewFPSet(config)
				}
				return store.Init(1, directory, "Spec")
			}
			store := newStore()
			fingerprints := []uint64{41}
			if test.bits != 0 {
				// Source disk recovery rejects an empty child checkpoint.
				// Populate both high-bit partitions for this valid-file control.
				fingerprints = append(fingerprints, uint64(1)<<63|43)
			}
			for _, fp := range fingerprints {
				if store.Put(fp) {
					t.Fatal("initial fingerprint already present")
				}
			}
			if err := store.BeginChkptFile("Spec"); err != nil {
				t.Fatal(err)
			}
			if err := store.CommitChkptFile("Spec"); err != nil {
				t.Fatal(err)
			}
			store.Close()
			store = newStore()
			defer store.Close()
			manager := NewNonDistributedFPSetManager(store, "local", nil)
			err := invokeDistributedServerOperation(func() error { return manager.Recover("unused") })
			if test.needsTrace {
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing trace: %T/%v", err, err)
				}
				if store.Size() != 0 || store.Contains(41) {
					t.Fatal("missing trace fell back to committed file membership")
				}
				// The files really are recoverable: failure must come from the
				// required trace, rather than an unavailable snapshot.
				if err := store.RecoverFile("Spec"); err != nil {
					t.Fatal(err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if store.Size() != uint64(len(fingerprints)) {
				t.Fatal("named checkpoint did not retain committed membership")
			}
			for _, fp := range fingerprints {
				if !store.Contains(fp) {
					t.Fatalf("named checkpoint lost fingerprint %d", fp)
				}
			}
		})
	}
	t.Run("offheap-trace-only", func(t *testing.T) {
		// Source OffHeap inherits NonCheckpointableDiskFPSet: named file
		// checkpointing is unsupported, but trace replay still requires a trace.
		config := NewFPSetConfigurationWithRatioAndImplementation(1, "tlc2.tool.fp.OffHeapDiskFPSet")
		config.SetMemory(1 << 20)
		config.SetFPBits(0)
		store := NewOffHeapDiskFPSet(config).Init(1, t.TempDir(), "Spec")
		defer store.Close()
		manager := NewNonDistributedFPSetManager(store, "local", nil)
		err := invokeDistributedServerOperation(func() error { return manager.Recover("unused") })
		if _, ok := err.(*NullPointerException); !ok {
			t.Fatalf("missing offheap trace: %T/%v", err, err)
		}
		if store.Size() != 0 {
			t.Fatal("missing trace mutated offheap membership")
		}
	})
}
