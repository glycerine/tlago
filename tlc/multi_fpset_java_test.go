package tlc

import (
	"strconv"
	"testing"
)

// Translated after implementation from tlc2.tool.fp.MultiFPSetTest.
// Assertions and fingerprint inputs follow the upstream methods; its binary
// print helper has an empty body. Temporary files use Go's test directory.

func TestJavaMultiFPSetPutMax(t *testing.T) {
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testPutMax")
	mfps.Put(uint64(1<<63 - 1))
}

func TestJavaMultiFPSetPutMin(t *testing.T) {
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testPutMin")
	mfps.Put(uint64(1 << 63))
}

func TestJavaMultiFPSetPutZero(t *testing.T) {
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testPutZero")
	mfps.Put(0)
}

func TestJavaMultiFPSetGetFPSet(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSet")
	var a uint64 = (1 << 62) + 1
	var b uint64 = 1
	aFPSet := mfps.fpSet(a)
	javaMultiFPAssertBool(t, aFPSet == mfps.fpSet(b), true, "aFPSet == mfps.fpSet(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), false, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), true, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, aFPSet.Contains(b), true, "aFPSet.Contains(b)")
	javaMultiFPAssertInt(t, aFPSet.Size(), 2, "aFPSet.Size()")
	fpSets := mfps.Sets
	s := NewInsMap[FPSet, struct{}]()
	for i := range fpSets {
		s.Set(fpSets[i], struct{}{})
	}
	s.Delkey(aFPSet)
	bFPSet := s.Cached()[0].key
	javaMultiFPAssertBool(t, bFPSet.Contains(a), false, "bFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), false, "bFPSet.Contains(b)")
	javaMultiFPAssertInt(t, bFPSet.Size(), 0, "bFPSet.Size()")
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSet0(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSet0")
	var a uint64 = (1 << 63) + 1
	var b uint64 = 1
	var c uint64 = (1 << 62) + 1
	var d uint64 = (3 << 62) + 1
	aFPSet := mfps.fpSet(a)
	bFPSet := mfps.fpSet(b)
	javaMultiFPAssertBool(t, aFPSet != bFPSet, true, "aFPSet != bFPSet")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), false, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), false, "bFPSet.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(c), false, "mfps.Put(c)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(d), false, "mfps.Put(d)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), true, "mfps.Contains(d)")
	for _, fpSet := range mfps.Sets {
		javaMultiFPAssertInt(t, fpSet.Size(), 2, "fpSet.Size()")
		javaMultiFPAssertInt(t, fpSet.(interface{ GetTblLoad() int64 }).GetTblLoad(), 2, "fpSet.(interface { GetTblLoad() int64 }).GetTblLoad()")
	}
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSet1(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(2)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSet1")
	var a uint64 = 1
	var b uint64 = (1 << 62) + 1
	var c uint64 = (1 << 63) + 1
	var d uint64 = (3 << 62) + 1
	s := NewInsMap[FPSet, struct{}]()
	aFPSet := mfps.fpSet(a)
	s.Set(aFPSet, struct{}{})
	bFPSet := mfps.fpSet(b)
	s.Set(bFPSet, struct{}{})
	cFPSet := mfps.fpSet(c)
	s.Set(cFPSet, struct{}{})
	dFPSet := mfps.fpSet(d)
	s.Set(dFPSet, struct{}{})
	javaMultiFPAssertInt(t, s.Len(), 4, "s.Len()")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(c), false, "mfps.Put(c)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(d), false, "mfps.Put(d)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), true, "mfps.Contains(d)")
	for fpSet := range s.All() {
		javaMultiFPAssertInt(t, fpSet.Size(), 1, "fpSet.Size()")
		javaMultiFPAssertInt(t, fpSet.(interface{ GetTblLoad() int64 }).GetTblLoad(), 1, "fpSet.(interface { GetTblLoad() int64 }).GetTblLoad()")
	}
	javaMultiFPAssertBool(t, aFPSet.Contains(a), true, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, aFPSet.Contains(b), false, "aFPSet.Contains(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(c), true, "aFPSet.Contains(c)")
	javaMultiFPAssertBool(t, aFPSet.Contains(d), false, "aFPSet.Contains(d)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), true, "bFPSet.Contains(b)")
	javaMultiFPAssertBool(t, bFPSet.Contains(a), false, "bFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(c), false, "bFPSet.Contains(c)")
	javaMultiFPAssertBool(t, bFPSet.Contains(d), true, "bFPSet.Contains(d)")
	javaMultiFPAssertBool(t, cFPSet.Contains(c), true, "cFPSet.Contains(c)")
	javaMultiFPAssertBool(t, cFPSet.Contains(b), false, "cFPSet.Contains(b)")
	javaMultiFPAssertBool(t, cFPSet.Contains(a), true, "cFPSet.Contains(a)")
	javaMultiFPAssertBool(t, cFPSet.Contains(d), false, "cFPSet.Contains(d)")
	javaMultiFPAssertBool(t, dFPSet.Contains(d), true, "dFPSet.Contains(d)")
	javaMultiFPAssertBool(t, dFPSet.Contains(b), true, "dFPSet.Contains(b)")
	javaMultiFPAssertBool(t, dFPSet.Contains(c), false, "dFPSet.Contains(c)")
	javaMultiFPAssertBool(t, dFPSet.Contains(a), false, "dFPSet.Contains(a)")
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSetL(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSetL")
	var a uint64 = (1 << 62) + 1
	var b uint64 = 1
	aFPSet := mfps.fpSet(a)
	javaMultiFPAssertBool(t, aFPSet == mfps.fpSet(b), true, "aFPSet == mfps.fpSet(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), false, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), true, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, aFPSet.Contains(b), true, "aFPSet.Contains(b)")
	javaMultiFPAssertInt(t, aFPSet.Size(), 2, "aFPSet.Size()")
	fpSets := mfps.Sets
	s := NewInsMap[FPSet, struct{}]()
	for i := range fpSets {
		s.Set(fpSets[i], struct{}{})
	}
	s.Delkey(aFPSet)
	bFPSet := s.Cached()[0].key
	javaMultiFPAssertBool(t, bFPSet.Contains(a), false, "bFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), false, "bFPSet.Contains(b)")
	javaMultiFPAssertInt(t, bFPSet.Size(), 0, "bFPSet.Size()")
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSet0L(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSet0")
	var a uint64 = (1 << 63) + 1
	var b uint64 = 1
	aFPSet := mfps.fpSet(a)
	bFPSet := mfps.fpSet(b)
	javaMultiFPAssertBool(t, aFPSet != bFPSet, true, "aFPSet != bFPSet")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), false, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), false, "bFPSet.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSet1L(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(2)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSet1")
	var a uint64 = 1
	var b uint64 = (1 << 62) + 1
	var c uint64 = (1 << 63) + 1
	var d uint64 = (3 << 62) + 1
	s := NewInsMap[FPSet, struct{}]()
	aFPSet := mfps.fpSet(a)
	s.Set(aFPSet, struct{}{})
	bFPSet := mfps.fpSet(b)
	s.Set(bFPSet, struct{}{})
	cFPSet := mfps.fpSet(c)
	s.Set(cFPSet, struct{}{})
	dFPSet := mfps.fpSet(d)
	s.Set(dFPSet, struct{}{})
	javaMultiFPAssertInt(t, s.Len(), 4, "s.Len()")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(c), false, "mfps.Put(c)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(d), false, "mfps.Put(d)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), true, "mfps.Contains(d)")
	for fpSet := range s.All() {
		javaMultiFPAssertInt(t, fpSet.Size(), 1, "fpSet.Size()")
	}
	javaMultiFPAssertBool(t, aFPSet.Contains(a), true, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, aFPSet.Contains(b), false, "aFPSet.Contains(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(c), true, "aFPSet.Contains(c)")
	javaMultiFPAssertBool(t, aFPSet.Contains(d), false, "aFPSet.Contains(d)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), true, "bFPSet.Contains(b)")
	javaMultiFPAssertBool(t, bFPSet.Contains(a), false, "bFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(c), false, "bFPSet.Contains(c)")
	javaMultiFPAssertBool(t, bFPSet.Contains(d), true, "bFPSet.Contains(d)")
	javaMultiFPAssertBool(t, cFPSet.Contains(c), true, "cFPSet.Contains(c)")
	javaMultiFPAssertBool(t, cFPSet.Contains(b), false, "cFPSet.Contains(b)")
	javaMultiFPAssertBool(t, cFPSet.Contains(a), true, "cFPSet.Contains(a)")
	javaMultiFPAssertBool(t, cFPSet.Contains(d), false, "cFPSet.Contains(d)")
	javaMultiFPAssertBool(t, dFPSet.Contains(d), true, "dFPSet.Contains(d)")
	javaMultiFPAssertBool(t, dFPSet.Contains(b), true, "dFPSet.Contains(b)")
	javaMultiFPAssertBool(t, dFPSet.Contains(c), false, "dFPSet.Contains(c)")
	javaMultiFPAssertBool(t, dFPSet.Contains(a), false, "dFPSet.Contains(a)")
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSetOffHeap(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("upstream 64-bit architecture guard")
	}
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.OffHeapDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSetOffHeap")
	var a uint64 = (1 << 62) + 1
	var b uint64 = 1
	aFPSet := mfps.fpSet(a)
	javaMultiFPAssertBool(t, aFPSet == mfps.fpSet(b), true, "aFPSet == mfps.fpSet(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), false, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), true, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, aFPSet.Contains(b), true, "aFPSet.Contains(b)")
	javaMultiFPAssertInt(t, aFPSet.Size(), 2, "aFPSet.Size()")
	fpSets := mfps.Sets
	s := NewInsMap[FPSet, struct{}]()
	for i := range fpSets {
		s.Set(fpSets[i], struct{}{})
	}
	s.Delkey(aFPSet)
	bFPSet := s.Cached()[0].key
	javaMultiFPAssertBool(t, bFPSet.Contains(a), false, "bFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), false, "bFPSet.Contains(b)")
	javaMultiFPAssertInt(t, bFPSet.Size(), 0, "bFPSet.Size()")
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSetOffHeap0(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("upstream 64-bit architecture guard")
	}
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.OffHeapDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSetOffHeap0")
	var a uint64 = (1 << 63) + 1
	var b uint64 = 1
	var c uint64 = (1 << 62) + 1
	var d uint64 = (3 << 62) + 1
	aFPSet := mfps.fpSet(a)
	bFPSet := mfps.fpSet(b)
	javaMultiFPAssertBool(t, aFPSet != bFPSet, true, "aFPSet != bFPSet")
	javaMultiFPAssertBool(t, aFPSet.Contains(a), false, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), false, "bFPSet.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(c), false, "mfps.Put(c)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(d), false, "mfps.Put(d)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), true, "mfps.Contains(d)")
	for _, fpSet := range mfps.Sets {
		javaMultiFPAssertInt(t, fpSet.Size(), 2, "fpSet.Size()")
		javaMultiFPAssertInt(t, fpSet.(interface{ GetTblLoad() int64 }).GetTblLoad(), 2, "fpSet.(interface { GetTblLoad() int64 }).GetTblLoad()")
	}
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetGetFPSetOffHeap1(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("upstream 64-bit architecture guard")
	}
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.OffHeapDiskFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(2)
	mfps := NewMultiFPSet(conf)
	t.Cleanup(mfps.Close)
	mfps.Init(1, t.TempDir(), "testGetFPSetOffHeap1")
	var a uint64 = 1
	var b uint64 = (1 << 62) + 1
	var c uint64 = (1 << 63) + 1
	var d uint64 = (3 << 62) + 1
	s := NewInsMap[FPSet, struct{}]()
	aFPSet := mfps.fpSet(a)
	s.Set(aFPSet, struct{}{})
	bFPSet := mfps.fpSet(b)
	s.Set(bFPSet, struct{}{})
	cFPSet := mfps.fpSet(c)
	s.Set(cFPSet, struct{}{})
	dFPSet := mfps.fpSet(d)
	s.Set(dFPSet, struct{}{})
	javaMultiFPAssertInt(t, s.Len(), 4, "s.Len()")
	javaMultiFPAssertBool(t, mfps.Contains(a), false, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(a), false, "mfps.Put(a)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), false, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(b), false, "mfps.Put(b)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), false, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(c), false, "mfps.Put(c)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), false, "mfps.Contains(d)")
	javaMultiFPAssertBool(t, mfps.Put(d), false, "mfps.Put(d)")
	javaMultiFPAssertBool(t, mfps.Contains(a), true, "mfps.Contains(a)")
	javaMultiFPAssertBool(t, mfps.Contains(b), true, "mfps.Contains(b)")
	javaMultiFPAssertBool(t, mfps.Contains(c), true, "mfps.Contains(c)")
	javaMultiFPAssertBool(t, mfps.Contains(d), true, "mfps.Contains(d)")
	for fpSet := range s.All() {
		javaMultiFPAssertInt(t, fpSet.Size(), 1, "fpSet.Size()")
		javaMultiFPAssertInt(t, fpSet.(interface{ GetTblLoad() int64 }).GetTblLoad(), 1, "fpSet.(interface { GetTblLoad() int64 }).GetTblLoad()")
	}
	javaMultiFPAssertBool(t, aFPSet.Contains(a), true, "aFPSet.Contains(a)")
	javaMultiFPAssertBool(t, aFPSet.Contains(b), false, "aFPSet.Contains(b)")
	javaMultiFPAssertBool(t, aFPSet.Contains(c), true, "aFPSet.Contains(c)")
	javaMultiFPAssertBool(t, aFPSet.Contains(d), false, "aFPSet.Contains(d)")
	javaMultiFPAssertBool(t, bFPSet.Contains(b), true, "bFPSet.Contains(b)")
	javaMultiFPAssertBool(t, bFPSet.Contains(a), false, "bFPSet.Contains(a)")
	javaMultiFPAssertBool(t, bFPSet.Contains(c), false, "bFPSet.Contains(c)")
	javaMultiFPAssertBool(t, bFPSet.Contains(d), true, "bFPSet.Contains(d)")
	javaMultiFPAssertBool(t, cFPSet.Contains(c), true, "cFPSet.Contains(c)")
	javaMultiFPAssertBool(t, cFPSet.Contains(b), false, "cFPSet.Contains(b)")
	javaMultiFPAssertBool(t, cFPSet.Contains(a), true, "cFPSet.Contains(a)")
	javaMultiFPAssertBool(t, cFPSet.Contains(d), false, "cFPSet.Contains(d)")
	javaMultiFPAssertBool(t, dFPSet.Contains(d), true, "dFPSet.Contains(d)")
	javaMultiFPAssertBool(t, dFPSet.Contains(b), true, "dFPSet.Contains(b)")
	javaMultiFPAssertBool(t, dFPSet.Contains(c), false, "dFPSet.Contains(c)")
	javaMultiFPAssertBool(t, dFPSet.Contains(a), false, "dFPSet.Contains(a)")
	javaMultiFPAssertBool(t, mfps.CheckInvariant(), true, "mfps.CheckInvariant()")
}

func TestJavaMultiFPSetCTorLowerMin(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MemFPSet")
	conf := NewFPSetConfiguration()
	conf.SetFPBits(0)
	defer func() {
		failure, ok := recover().(*TLCError)
		if !ok || failure == nil || javaRuntimeException(failure) == nil {
			t.Fatal("upstream constructor must reject zero bits with a runtime exception")
		}
	}()
	NewMultiFPSet(conf)
}

func TestJavaMultiFPSetCTorMin(t *testing.T) {
	conf := NewFPSetConfiguration()
	conf.SetFPBits(1)
	NewMultiFPSet(conf)
}

func TestJavaMultiFPSetCTorMax(t *testing.T) {
	conf := NewFPSetConfiguration()
	conf.SetFPBits(30)
	defer func() {
		if failure := recover(); failure != nil {
			if e, ok := failure.(*IllegalArgumentException); ok && e.GetMessage() != nil && *e.GetMessage() == "Given fpSetConfig results in zero or negative fp count." {
				return
			}
			if e, ok := failure.(error); ok && javaSystemFailureCode(e) == ECSystemOutOfMemory {
				return
			}
			t.Fatalf("upstream maximum constructor failure = %v", failure)
		}
	}()
	NewMultiFPSet(conf)
}

func TestJavaMultiFPSetCTorHigherMax(t *testing.T) {
	defer func() {
		failure, ok := recover().(*TLCError)
		if !ok || failure == nil || javaRuntimeException(failure) == nil {
			t.Fatal("upstream constructor must reject 31 bits with a runtime exception")
		}
	}()
	conf := NewFPSetConfiguration()
	conf.SetFPBits(31)
	NewMultiFPSet(conf)
}

func javaMultiFPAssertBool(t *testing.T, got, want bool, expression string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %t, want %t", expression, got, want)
	}
}

func javaMultiFPAssertInt[T int | int64 | uint64](t *testing.T, got, want T, expression string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %d, want %d", expression, got, want)
	}
}
