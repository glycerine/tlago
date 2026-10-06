// Copyright (c) 2016 Microsoft Research. All rights reserved.
// Copyright (c) 2024, Oracle and/or its affiliates.
package tlc

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
)

// Whole original OffHeapDiskFPSetTest, including its 64-bit setup assumption.
func TestJavaOffHeapDiskFPSet(t *testing.T) {
	// Ant forks each original test class; constructor-only factory tests from
	// earlier classes must not remain in this class's singleton barrier.
	InitializeOffHeapDiskFPSetStatics()
	run := func(name string, body func(*testing.T)) {
		t.Run(name, func(t *testing.T) {
			if strconv.IntSize != 64 {
				t.Skip("source Assume: TLCRuntime architecture BIT_64")
			}
			body(t)
		})
	}
	for i, input := range []struct{ seed, length int64 }{
		{1473793977852, 87}, {1473793976137, 87}, {1473839150698, 46}, {1473839150698, 46},
		{1473839322351, 23}, {1473839380539, 23}, {1473839422899, 11}, {1473839543883, 11},
		{1473871461079, 64}, {1473871462765, 64}, {1473871522834, 32}, {1473871526136, 32},
		{1473873732723, 47}, {1473871294851, 93}, {1473871365625, 93}, {1473871209569, 157},
	} {
		run(fmt.Sprintf("testInsertAndEvict%d", i+1), func(t *testing.T) { javaOffHeapInsertAndEvict(t, input.seed, input.length) })
	}
	for _, pages := range []int64{1, 3, 5, 9} {
		run(fmt.Sprintf("testOffset%dPage", pages), func(t *testing.T) { javaOffHeapOffset(t, pages*diskFPSetNumEntriesPerPage, 1474536306841) })
	}
	run("testWriteIndex", func(t *testing.T) {
		cfg := newJavaDummyFPSetConfiguration()
		cfg.setMemoryInFingerprintCnt(1)
		set := NewOffHeapDiskFPSet(cfg.FPSetConfiguration)
		t.Cleanup(set.Close)
		const length int32 = 99999999
		// Source assertion deliberately evaluates signed int multiplication first.
		product := length
		product *= int32(diskFPSetNumEntriesPerPage)
		if product >= 1 {
			t.Fatal("length * NumEntriesPerPage did not overflow")
		}
		raf, err := NewRandomAccessFile(filepath.Join(t.TempDir(), "foobar"), "rw")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = raf.Close() })
		t.Logf("Writing the original %d-entry index with a zero-filling RandomAccessFile", length)
		if err := set.writeIndex(make([]uint64, int(length)), &javaOffHeapDummyRandomAccessFile{raf}, math.MaxInt32); err != nil {
			t.Fatal(err)
		}
	})
	run("testMergeDuplicate", func(t *testing.T) { javaOffHeapMerge(t, true) })
	run("testMergeDistinct", func(t *testing.T) { javaOffHeapMerge(t, false) })
}

func javaOffHeapFingerprint(random *JavaRandom) uint64 {
	return uint64(int64(random.NextIntN(math.MaxInt32-1))+1)<<32 | uint64(uint32(random.NextInt()))
}

func javaOffHeapInitialized(t *testing.T, length int64) *OffHeapDiskFPSet {
	t.Helper()
	cfg := newJavaDummyFPSetConfiguration()
	cfg.setMemoryInFingerprintCnt(length)
	set := NewOffHeapDiskFPSet(cfg.FPSetConfiguration)
	t.Cleanup(set.Close)
	set.Init(1, t.TempDir(), "OffHeapDiskFPSetTest")
	return set
}

// Whole original doTest: all insertion, special-position, disk membership,
// invariant, and memory-only membership assertions retain their original order.
func javaOffHeapInsertAndEvict(t *testing.T, seed, length int64) {
	t.Helper()
	set := javaOffHeapInitialized(t, length)
	random := NewJavaRandom(seed)
	for i := int64(0); i < length/2; i++ {
		if set.Put(javaOffHeapFingerprint(random)) {
			t.Fatal("first put unexpectedly duplicated fingerprint")
		}
	}
	expected := LongArraysToArray(set.array)
	if set.GetGrowDiskMark() != 0 {
		t.Fatal("initial growDiskMark is not zero")
	}
	set.ForceFlush()
	set.Contains(1)
	if set.GetGrowDiskMark() != 1 {
		t.Fatal("growDiskMark did not increment")
	}
	actual := set.array
	for i, value := range expected {
		if value == 0 {
			if got := actual.Get(int64(i)); got != 0 {
				t.Fatalf("Expected empty position with seed %dL and length %d: [%d]=%d; expected=%v", seed, length, i, got, expected)
			}
		} else if value < 0 {
			if got := actual.Get(int64(i)); got != 0 {
				t.Fatalf("Expected negative position with seed %dL and length %d: [%d]=%d; expected=%v", seed, length, i, got, expected)
			}
		}
	}
	random = NewJavaRandom(seed)
	for i := int64(0); i < length/2; i++ {
		fp := javaOffHeapFingerprint(random)
		if !set.Contains(fp) {
			t.Fatalf("Failed to find fp %d/%d with seed %dL and length %d; expected=%v", fp, int64(fp|diskFPSetMarkFlushed), seed, length, expected)
		}
	}
	if !set.CheckInvariant() {
		t.Fatalf("Invariant violated with seed %dL and length %d; expected=%v", seed, length, expected)
	}
	set.index = nil
	random = NewJavaRandom(seed)
	for i := int64(0); i < length/2; i++ {
		fp := javaOffHeapFingerprint(random)
		if !set.Contains(fp) {
			t.Fatalf("Failed to find memory-only fp %d/%d with seed %dL and length %d; expected=%v", fp, int64(fp|diskFPSetMarkFlushed), seed, length, expected)
		}
	}
	set.Close()
}

// Whole original doTestOffset. Its loop compares i to the shrinking TreeSet
// size; preserve that source condition instead of checking a larger loop.
func javaOffHeapOffset(t *testing.T, length, seed int64) {
	t.Helper()
	set := javaOffHeapInitialized(t, length)
	longs := make([]uint64, 0, length/2)
	random := NewJavaRandom(seed)
	for i := int64(0); i < length/2; i++ {
		fp := javaOffHeapFingerprint(random)
		if set.Put(fp) {
			t.Fatal("first put unexpectedly duplicated fingerprint")
		}
		longs = append(longs, fp)
	}
	sort.Slice(longs, func(i, j int) bool { return longs[i] < longs[j] })
	// TreeSet removes duplicate fingerprints, though the preceding put assertion
	// already rejects duplicates for these exact seeded inputs.
	unique := longs[:0]
	for _, fp := range longs {
		if len(unique) == 0 || unique[len(unique)-1] != fp {
			unique = append(unique, fp)
		}
	}
	longs = unique
	set.ForceFlush()
	if set.Contains(1) {
		t.Fatal("contains(1) unexpectedly true after flush")
	}
	for i := int64(0); i < int64(len(longs)); i++ {
		fp := longs[0]
		got, err := set.getDiskOffset(0, fp+1)
		if err != nil {
			t.Fatal(err)
		}
		if got != i+1 {
			t.Fatalf("Length: %d with seed: %d, offset=%d, want %d", length, seed, got, i+1)
		}
		longs = longs[1:]
	}
}

type javaOffHeapDummyRandomAccessFile struct{ *RandomAccessFile }

func (f *javaOffHeapDummyRandomAccessFile) Read(p []byte) (int, error)  { clear(p); return len(p), nil }
func (f *javaOffHeapDummyRandomAccessFile) ReadByteValue() (int, error) { return 0, nil }

type javaOffHeapDummyIterator struct {
	*offHeapIterator
	i int64
}

func (i *javaOffHeapDummyIterator) markNext() (int64, bool) { v := i.i; i.i++; return v, true }
func (i *javaOffHeapDummyIterator) hasNext() bool           { return false }

func javaOffHeapMerge(t *testing.T, duplicate bool) {
	t.Helper()
	var recorder *MemoryRecorder
	if duplicate {
		recorder = &MemoryRecorder{}
		AddMessageRecorder(recorder)
		t.Cleanup(func() { RemoveMessageRecorder(recorder) })
	}
	cfg := newJavaDummyFPSetConfiguration()
	cfg.setMemoryInFingerprintCnt(1)
	set := NewOffHeapDiskFPSet(cfg.FPSetConfiguration)
	t.Cleanup(set.Close)
	dir := t.TempDir()
	in, err := NewBufferedRandomAccessFile(filepath.Join(dir, "OffHeapDiskFPSetTest_test.bin"), "rw")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = in.Close() })
	if err := in.SetLength(8 * fpSetLongSize); err != nil {
		t.Fatal(err)
	}
	input := []int64{3, 4, 5, 6, 7, 8, 10, 11}
	elements := int64(2)
	if duplicate {
		input = []int64{1, 2, 4, 6, 7, 8, 10, 11}
		elements = 8
	}
	for _, fp := range input {
		if err := in.WriteLong(fp); err != nil {
			t.Fatal(err)
		}
	}
	if err := in.Seek(0); err != nil {
		t.Fatal(err)
	}
	out, err := NewRandomAccessFile(filepath.Join(dir, "OffHeapDiskFPSetTest_test.out"), "rw")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = out.Close() })
	itr := &javaOffHeapDummyIterator{offHeapIterator: newOffHeapIterator(nil, elements, 0, nil, true), i: 1}
	if err := set.mergeOffHeapEntries(in, out, offHeapMergeIterator{itr.elements, itr.markNext, itr.hasNext}, 8); err != nil {
		t.Fatal(err)
	}
	info, err := out.file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 10*fpSetLongSize {
		t.Fatalf("output length=%d, want %d", info.Size(), 10*fpSetLongSize)
	}
	if err := out.Seek(0); err != nil {
		t.Fatal(err)
	}
	for _, want := range []int64{1, 2, 3, 4, 5, 6, 7, 8, 10, 11} {
		got, err := ReadLong(out)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("output fingerprint=%d, want %d", got, want)
		}
	}
	if duplicate {
		records := recorder.Records(ECTLCFPValueAlreadyOnDisk)
		if len(records) != 6 {
			t.Fatalf("duplicate records=%d, want 6", len(records))
		}
		found := make(map[int]bool)
		for _, record := range records {
			fp, err := strconv.Atoi(record.Params[0])
			if err != nil {
				t.Fatal(err)
			}
			found[fp] = true
		}
		if len(found) != 6 {
			t.Fatalf("warning fingerprint set=%v", found)
		}
		for _, want := range []int{1, 2, 4, 6, 7, 8} {
			if !found[want] {
				t.Fatalf("missing duplicate warning %d", want)
			}
		}
	}
}
