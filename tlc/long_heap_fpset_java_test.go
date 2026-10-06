package tlc

import "testing"

// Complete inherited FPSetTest.testSimpleFill; retain all four source put/contains
// pairs. Factories and their configuration semantics stay with each subclass.
func javaLongFPSetSimpleFill(t *testing.T, set FPSet) {
	t.Helper()
	fp := uint64(1)
	if set.Put(fp) {
		t.Fatalf("put(%d) unexpectedly present", fp)
	}
	if !set.Contains(fp) {
		t.Fatalf("contains(%d) unexpectedly false", fp)
	}
	fp++
	if set.Put(fp) {
		t.Fatalf("put(%d) unexpectedly present", fp)
	}
	if !set.Contains(fp) {
		t.Fatalf("contains(%d) unexpectedly false", fp)
	}
	fp++
	if set.Put(fp) {
		t.Fatalf("put(%d) unexpectedly present", fp)
	}
	if !set.Contains(fp) {
		t.Fatalf("contains(%d) unexpectedly false", fp)
	}
	fp++
	if set.Put(fp) {
		t.Fatalf("put(%d) unexpectedly present", fp)
	}
	if !set.Contains(fp) {
		t.Fatalf("contains(%d) unexpectedly false", fp)
	}
	fp++
}

// Original test-long DiskFPSetTest factory constructs LSBDiskFPSet directly.
// Its random method is in the tlc_fp_stress target; the sequential method remains pending.
func TestJavaLongLSBDiskFPSet(t *testing.T) {
	t.Run("testSimpleFill", func(t *testing.T) {
		set := NewLSBDiskFPSet(NewFPSetConfiguration())
		t.Logf("DiskFPSet approx. consumes MiB: %d", (set.GetMaxTblCnt()*fpSetLongSize)>>20)
		set.Init(1, t.TempDir(), "FPSetTestTest")
		t.Cleanup(set.Close)
		javaLongFPSetSimpleFill(t, set)
	})
}

// Original test-long MSBDiskFPSetTest uses its supplied default configuration.
func TestJavaLongMSBDiskFPSet(t *testing.T) {
	t.Run("testSimpleFill", func(t *testing.T) {
		set := NewMSBDiskFPSet(NewFPSetConfiguration())
		set.Init(1, t.TempDir(), "FPSetTestTest")
		t.Cleanup(set.Close)
		javaLongFPSetSimpleFill(t, set)
	})
}
