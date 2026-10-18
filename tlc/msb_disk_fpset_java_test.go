// Copyright (c) 2012 Markus Alexander Kuppe. All rights reserved.
package tlc

import (
	"fmt"
	"testing"
	"time"
)

// Whole original MSBDiskFPSetTest2.getMSBDiskFPSet helper.
func javaMSBDiskFPSetForSourceTest(t *testing.T) *MSBDiskFPSet {
	t.Helper()
	config := newJavaDummyFPSetConfiguration()
	config.setMemoryInFingerprintCnt(100)
	if got := config.GetMemoryInFingerprintCnt(); got != 100 {
		t.Fatalf("fingerprint memory=%d, want 100", got)
	}
	filename := fmt.Sprintf("tlc2.tool.fp.MSBDiskFPSetTest2%d", time.Now().UnixMilli())
	set := NewMSBDiskFPSet(config.FPSetConfiguration)
	set.Init(1, t.TempDir(), filename)
	t.Cleanup(set.Close)
	return set
}

// Whole original class: fifteen inherited methods and four declared methods.
func TestJavaMSBDiskFPSet2(t *testing.T) {
	newSet := func(config *FPSetConfiguration) *DiskFPSet { return NewMSBDiskFPSet(config).DiskFPSet }
	javaHeapDiskFPSetConstructorMethods(t, 1<<8, 1<<31, newSet)
	javaHeapDiskFPSetRecoveryMethods(t, "tlc2.tool.fp.MSBDiskFPSetTest2", newSet)
	t.Run("testGetLast", func(t *testing.T) {
		set := javaMSBDiskFPSetForSourceTest(t)
		const highFP uint64 = 1 << 62
		set.Put(highFP)
		iterator := newMSBDiskIterator(set.tbl)
		got, err := iterator.getLast()
		if err != nil {
			t.Fatal(err)
		}
		if got != highFP {
			t.Fatalf("getLast=%d, want %d", got, highFP)
		}
		if err := set.flushTable(); err != nil {
			t.Fatal(err)
		}
		// The source constructs and discards a new iterator, then calls the OLD one.
		newMSBDiskIterator(set.tbl)
		if _, err := iterator.getLast(); err != nil {
			if _, ok := err.(*NoSuchElementException); !ok {
				t.Fatal(err)
			}
			const lowFP uint64 = 1
			set.Put(lowFP)
			iterator = newMSBDiskIterator(set.tbl)
			got, err := iterator.getLast()
			if err != nil {
				t.Fatal(err)
			}
			if got != lowFP {
				t.Fatalf("getLast=%d, want %d", got, lowFP)
			}
			return
		}
		t.Fatal("getLast did not throw NoSuchElementException after flush")
	})
	for _, input := range []struct {
		name string
		fp   uint64
	}{
		{"testHighFingerprint1", 9223368718049406096},
		{"testHighFingerprint2", 9223335424116589377},
	} {
		t.Run(input.name, func(t *testing.T) {
			set := javaMSBDiskFPSetForSourceTest(t)
			if set.Put(input.fp) {
				t.Fatal("first put unexpectedly duplicated fingerprint")
			}
			if err := set.flushTable(); err != nil {
				t.Fatal(err)
			}
			if !set.Put(input.fp) {
				t.Fatal("second put did not find fingerprint")
			}
			if err := set.flushTable(); err != nil {
				t.Fatal(err)
			}
			if !set.Put(input.fp) {
				t.Fatal("third put did not find fingerprint")
			}
		})
	}
	t.Run("testGetLastNoBuckets", func(t *testing.T) {
		set := javaMSBDiskFPSetForSourceTest(t)
		iterator := newMSBDiskIterator(set.tbl)
		if _, err := iterator.getLast(); err != nil {
			if _, ok := err.(*NoSuchElementException); !ok {
				t.Fatal(err)
			}
			return
		}
		t.Fatal("getLast did not throw NoSuchElementException with no buckets")
	})
}
