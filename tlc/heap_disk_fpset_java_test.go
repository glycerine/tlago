// Copyright (c) 2012 Markus Alexander Kuppe. All rights reserved.
package tlc

import "testing"

// Original DummyFPSetConfiguration overrides getMemoryInBytes only. Calls from
// the inherited getMemoryInFingerprintCnt must dispatch through that override.
type javaDummyFPSetConfiguration struct{ *FPSetConfiguration }

func newJavaDummyFPSetConfiguration() *javaDummyFPSetConfiguration {
	config := &javaDummyFPSetConfiguration{NewFPSetConfiguration()}
	config.GetMemoryInBytesOverride = func() int64 { return config.MemoryInBytes }
	return config
}

func (c *javaDummyFPSetConfiguration) setMemoryInFingerprintCnt(length int64) {
	c.MemoryInBytes = length * fpSetLongSize
}

// Whole inherited AbstractHeapBasedDiskFPSetTest.doTest helper.
func javaHeapDiskFPSetConstructorCheck(t *testing.T, physicalMemoryInBytes int64, newSet func(*FPSetConfiguration) *DiskFPSet) {
	t.Helper()
	config := newJavaDummyFPSetConfiguration()
	config.SetMemory(physicalMemoryInBytes)
	set := newSet(config.FPSetConfiguration)
	maxTblCntInBytes := set.GetMaxTblCnt() * fpSetLongSize
	if physicalMemoryInBytes < maxTblCntInBytes {
		t.Fatalf("Internal storage exceeds allocated memory: %d < %d", physicalMemoryInBytes, maxTblCntInBytes)
	}
	if maxTblCntInBytes <= 0 {
		t.Fatal("Internal storage underflow allocated memory")
	}
	lowerLimit := float64(physicalMemoryInBytes/2) / set.GetAuxiliaryStorageRequirement()
	if lowerLimit > float64(maxTblCntInBytes) {
		t.Fatalf("Internal storage falls short lower allocation limit: %g > %d", lowerLimit, maxTblCntInBytes)
	}
}

// All twelve original constructor methods, including subclass-specific bounds.
func javaHeapDiskFPSetConstructorMethods(t *testing.T, lower, upper int64, newSet func(*FPSetConfiguration) *DiskFPSet) {
	t.Helper()
	for _, input := range []struct {
		name   string
		memory int64
	}{
		{"testCtorLLMinus1", lower - 1}, {"testCtorLL", lower}, {"testCtorLLPlus1", lower + 1}, {"testCtorLLNextPow2Min1", (lower << 1) - 1},
		{"testCtorPow16Minus1", (1 << 16) - 1}, {"testCtorPow16", 1 << 16}, {"testCtorPow16Plus1", (1 << 16) + 1}, {"testCtorPow16NextPow2Min1", ((1 << 16) << 1) - 1},
		{"testCtorULMinus1", upper - 1}, {"testCtorUL", upper}, {"testCtorULPlus1", upper + 1}, {"testCtorULNextPow2Min1", (upper << 1) - 1},
	} {
		t.Run(input.name, func(t *testing.T) { javaHeapDiskFPSetConstructorCheck(t, input.memory, newSet) })
	}
}

// Whole original three inherited low-level fingerprint recovery methods. These
// retain their original ranges; no checkpoint/recovery model is added here.
func javaHeapDiskFPSetRecoveryMethods(t *testing.T, className string, newSet func(*FPSetConfiguration) *DiskFPSet) {
	t.Helper()
	t.Run("testFPSetRecovery", func(t *testing.T) {
		const limit = 99999
		metadir := t.TempDir()
		trace := NewTLCTrace(metadir, className)
		t.Cleanup(func() { _ = trace.Close() })
		predecessor := NewEmptyState()
		predecessor.UID = 1
		if err := trace.WriteInitState(nil, uint64(predecessor.UID)); err != nil {
			t.Fatal(err)
		}
		for fp := predecessor.UID + 1; fp < limit; fp++ {
			if _, err := trace.WriteStateRecord(predecessor, uint64(fp), nil); err != nil {
				t.Fatal(err)
			}
			predecessor.UID = fp
		}
		if err := trace.BeginChkpt(); err != nil {
			t.Fatal(err)
		}
		if err := trace.CommitChkpt(); err != nil {
			t.Fatal(err)
		}
		set := newSet(NewFPSetConfiguration())
		set.Init(1, metadir, className)
		t.Cleanup(set.Close)
		if err := set.RecoverTrace(trace); err != nil {
			t.Fatal(err)
		}
		if got := set.Size(); got != limit-1 {
			t.Fatalf("size=%d, want %d", got, limit-1)
		}
		for fp := uint64(1); fp < limit; fp++ {
			if !set.Contains(fp) {
				t.Fatalf("missing recovered fingerprint %d", fp)
			}
		}
	})
	t.Run("testFPSetRecovery2", func(t *testing.T) {
		set := newSet(NewFPSetConfiguration())
		set.Init(1, t.TempDir(), className+"testFPSetRecovery2")
		t.Cleanup(set.Close)
		set.ForceFlush()
		for fp := uint64(1); fp <= 1024; fp++ {
			if err := set.RecoverFP(fp); err != nil {
				t.Fatal(err)
			}
		}
	})
	t.Run("testFPSetRecoveryDuplicate", func(t *testing.T) {
		set := newSet(NewFPSetConfiguration())
		set.Init(1, t.TempDir(), className+"testFPSetRecoveryDuplicate")
		t.Cleanup(set.Close)
		set.ForceFlush()
		t.Setenv(DiskFPSetError2WarningProperty, "true")
		if err := set.RecoverFP(1); err != nil {
			t.Fatal(err)
		}
		if err := set.RecoverFP(1); err != nil {
			t.Fatal(err)
		}
	})
}

// Whole original LSBDiskFPsetTest: all fifteen inherited methods.
func TestJavaLSBDiskFPset(t *testing.T) {
	newSet := func(config *FPSetConfiguration) *DiskFPSet { return NewLSBDiskFPSet(config).DiskFPSet }
	javaHeapDiskFPSetConstructorMethods(t, 1<<9, 1<<31, newSet)
	javaHeapDiskFPSetRecoveryMethods(t, "tlc2.tool.fp.LSBDiskFPsetTest", newSet)
}
