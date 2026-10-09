// Copyright (c) 2012 Markus Alexander Kuppe. All rights reserved.
package tlc

import "testing"

// Original OffHeapDiskFPSetLongTest collision/position methods and inherited testSimpleFill.
// The full multiple-flush method is in offheap_multiple_flushes_java_test.go
// behind tlc_fp_stress; the full inherited random method is in
// offheap_long_fpset_stress_java_test.go. Its full execution and the inherited
// sequential method remain pending.
// The source factory ignores the supplied config, selecting ratio 1.0 instead.
func TestJavaOffHeapDiskFPSetLong(t *testing.T) {
	newSet := func(t *testing.T) *OffHeapDiskFPSet {
		t.Helper()
		set := NewOffHeapDiskFPSet(NewFPSetConfigurationWithRatioAndImplementation(1.0, "tlc2.tool.fp.OffHeapDiskFPSet"))
		set.Init(1, t.TempDir(), "FPSetTestTest")
		t.Cleanup(set.Close)
		return set
	}
	t.Run("testSimpleFill", func(t *testing.T) {
		set := newSet(t)
		javaLongFPSetSimpleFill(t, set)
	})
	t.Run("testCollisionBucket", func(t *testing.T) {
		set := newSet(t)
		for i := 0; i < diskFPSetInitialBucketCap+1; i++ {
			fp := uint64(i + 1)
			if set.Put(fp) {
				t.Fatalf("put(%d) unexpectedly present", fp)
			}
			if !set.Contains(fp) {
				t.Fatalf("contains(%d) unexpectedly false", fp)
			}
		}
	})
	t.Run("testPosition", func(t *testing.T) {
		set := newSet(t)
		if set.Put(1<<63 - 1) {
			t.Fatal("put(MAX_VALUE) unexpectedly present")
		}
		if !set.Contains(1<<63 - 1) {
			t.Fatal("contains(MAX_VALUE) unexpectedly false")
		}
		if set.Put(1) {
			t.Fatal("put(1) unexpectedly present")
		}
		if !set.Contains(1) {
			t.Fatal("contains(1) unexpectedly false")
		}
	})
}
