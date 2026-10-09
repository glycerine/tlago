//go:build tlc_fp_stress

// Copyright (c) 2012 Markus Alexander Kuppe. All rights reserved.
package tlc

import "testing"

// Complete OffHeapDiskFPSetLongTest.testMultipleFlushes. Preserve the source
// ratio-1.0 factory, single reader, Java RNG seed, four rounds, runtime-derived
// insertion bound and exact invariant count. Run normally, never with -race.
func TestJavaOffHeapDiskFPSetLong_testMultipleFlushes(t *testing.T) {
	set := NewOffHeapDiskFPSet(NewFPSetConfigurationWithRatioAndImplementation(1.0, "tlc2.tool.fp.OffHeapDiskFPSet"))
	set.Init(1, t.TempDir(), "FPSetTestTest")
	t.Cleanup(set.Close)
	rnd := NewJavaRandom(15041980)
	freeMemoryInFPs := tlcRuntimeNonHeapPhysicalMemory() / fpSetLongSize
	t.Logf("four original rounds of %d insertions", freeMemoryInFPs)
	for flushes := int64(0); flushes < 4; flushes++ {
		for i := int64(0); i < freeMemoryInFPs; i++ {
			if set.Put(uint64(rnd.NextLong())) {
				t.Fatalf("put unexpectedly present in round %d at insertion %d", flushes+1, i)
			}
		}
		if !set.CheckInvariant(uint64((flushes + 1) * freeMemoryInFPs)) {
			t.Fatalf("checkInvariant failed after round %d", flushes+1)
		}
		t.Logf("round %d invariant passed: %d fingerprints", flushes+1, (flushes+1)*freeMemoryInFPs)
	}
}
