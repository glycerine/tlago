// Copyright (c) 2011 Microsoft Corporation. All rights reserved.
package tlc

import (
	"fmt"
	"math"
	"testing"
	"time"
)

var javaShortDiskFPSetCounter int32

// Whole original ShortDiskFPSetTest and its AbstractFPSetTest setup/helpers.
func TestJavaShortDiskFPSet(t *testing.T) {
	property, _ := tlcLookupSystemProperty("tlc2.tool.fp.ShortDiskFPSetTest.runKnown")
	runKnownFailures := javaBooleanProperty(property)
	run := func(name string, body func(*testing.T, string)) {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			t.Logf("Test started at %s", time.Now())
			t.Cleanup(func() { t.Logf("Test finished at %s", time.Now()) })
			body(t, dir)
		})
	}
	getSet := func(t *testing.T, dir string, config *FPSetConfiguration) *javaDummyDiskFPSet {
		t.Helper()
		set := &javaDummyDiskFPSet{NewLSBDiskFPSet(config)}
		filename := fmt.Sprintf("FPSetTestTest%d", javaShortDiskFPSetCounter)
		javaShortDiskFPSetCounter++
		set.Init(1, dir, filename)
		t.Cleanup(set.Close)
		return set
	}
	getInitialized := func(t *testing.T, dir string) *javaDummyDiskFPSet {
		t.Helper()
		set := getSet(t, dir, NewFPSetConfiguration())
		// AbstractFPSetTest initializes again after the subclass getFPSet returns.
		set.Init(1, dir, "FPSetTestTest")
		t.Logf("Maximum FPSet table count is: %d (approx: %d GiB)", set.GetMaxTblCnt(), set.GetMaxTblCnt()*fpSetLongSize>>20)
		t.Logf("FPSet lock count is: %d", set.GetLockCnt())
		t.Logf("FPSet bucket count is: %d", set.GetTblCapacity())
		t.Log("Testing tlc2.tool.fp.DummyDiskFPSet")
		return set
	}
	diskLookup := func(t *testing.T, set *javaDummyDiskFPSet, fp uint64) bool {
		t.Helper()
		found, err := set.diskLookup(fp)
		if err != nil {
			t.Fatal(err)
		}
		return found
	}
	flush := func(t *testing.T, set *javaDummyDiskFPSet) {
		t.Helper()
		if err := set.flushTable(); err != nil {
			t.Fatal(err)
		}
	}
	known := func(t *testing.T) bool {
		t.Helper()
		if !runKnownFailures {
			t.Log("Skipping test failing due to Bug #213 in general/bugzilla/index.html")
			return false
		}
		return true
	}
	for _, input := range []struct {
		name string
		fp   uint64
	}{
		{"testWithoutZeroFP", 0}, {"testWithoutMinFP", uint64(1) << 63}, {"testWithoutMaxFP", math.MaxInt64},
	} {
		run(input.name, func(t *testing.T, dir string) {
			set := getSet(t, dir, NewFPSetConfiguration())
			if set.Contains(input.fp) {
				t.Fatal("Succeeded to look up 0 fp")
			}
		})
	}
	for _, input := range []struct {
		name        string
		fp          uint64
		conditional bool
	}{
		{"testZeroFP", 0, true}, {"testMinFP", uint64(1) << 63, true},
		{"testMinMin1FP", math.MaxInt64, false}, {"testNeg1FP", math.MaxUint64, false},
		{"testPos1FP", 1, false}, {"testMaxFP", math.MaxInt64, false},
	} {
		run(input.name, func(t *testing.T, dir string) {
			if input.conditional && !known(t) {
				return
			}
			set := getSet(t, dir, NewFPSetConfiguration())
			if set.Put(input.fp) {
				t.Fatal("first put unexpectedly duplicated fingerprint")
			}
			if !set.Contains(input.fp) {
				t.Fatal("Failed to look up fp")
			}
		})
	}
	run("testValues", func(t *testing.T, dir string) {
		set := getSet(t, dir, NewFPSetConfiguration())
		values := []int64{0, 1, math.MaxInt64 / 2, math.MaxInt64 - 1, math.MaxInt64}
		entries := []int64{0, 1, int64(math.MaxInt32) * 1024}
		for _, loVal := range values {
			for _, hiVal := range values {
				for _, fp := range values {
					for _, loEntry := range entries {
						for _, hiEntry := range entries {
							javaShortDiskCalculateMidEntry(t, set, loVal, hiVal, fp, loEntry, hiEntry)
						}
					}
				}
			}
		}
	})
	run("testDiskLookupWithFpOnLoPage", func(t *testing.T, dir string) {
		config := newJavaDummyFPSetConfiguration()
		config.SetRatio(1)
		config.SetMemory(1000)
		set := getSet(t, dir, config.FPSetConfiguration)
		const fp uint64 = 1
		for i := uint64(0); i < 1024*3; i++ {
			if set.Put(fp + i) {
				t.Fatal("first put unexpectedly duplicated fingerprint")
			}
			if !set.Contains(fp + i) {
				t.Fatal("missing inserted fingerprint")
			}
		}
		if !diskLookup(t, set, fp) {
			t.Fatal("Failed to lookup fp on first page")
		}
	})
	for _, input := range []struct {
		name        string
		fp          uint64
		conditional bool
	}{
		{"testMemLookupWithZeros", 0, true}, {"testMemLookupWithMin", 0, true}, {"testMemLookupWithMax", math.MaxInt64, false},
	} {
		run(input.name, func(t *testing.T, dir string) {
			if input.conditional && !known(t) {
				return
			}
			set := getSet(t, dir, NewFPSetConfigurationWithRatio(.75))
			if set.memInsert(input.fp) {
				t.Fatal("memInsert unexpectedly duplicated fingerprint")
			}
			if diskLookup(t, set, input.fp) {
				t.Fatal("unexpected fingerprint on disk")
			}
			if !set.memLookup(input.fp) {
				t.Fatal("missing fingerprint in memory")
			}
		})
	}
	run("testDiskLookupWithZeros", func(t *testing.T, dir string) {
		const fp uint64 = 0
		set := getSet(t, dir, NewFPSetConfigurationWithRatio(.75))
		if set.Put(fp) {
			t.Fatal("first put unexpectedly duplicated fingerprint")
		}
		if set.memLookup(fp) {
			t.Fatal("zero unexpectedly visible in memory")
		}
		if diskLookup(t, set, fp) {
			t.Fatal("zero unexpectedly visible on disk")
		}
		if set.Contains(fp) {
			t.Fatal("zero unexpectedly contained before flush")
		}
		flush(t, set)
		if set.memLookup(fp) {
			t.Fatal("zero unexpectedly visible in memory after flush")
		}
		if !diskLookup(t, set, fp) {
			t.Fatal("zero absent on disk after flush")
		}
		if !set.Contains(fp) {
			t.Fatal("zero absent after flush")
		}
	})
	run("testDiskLookupWithMin", func(t *testing.T, dir string) {
		const fp uint64 = 0 // Long.MIN_VALUE & 0x7FFFFFFFFFFFFFFF
		set := getSet(t, dir, NewFPSetConfiguration())
		if set.memInsert(fp) {
			t.Fatal("memInsert unexpectedly duplicated fingerprint")
		}
		if diskLookup(t, set, fp) {
			t.Fatal("unexpected fingerprint on disk")
		}
		flush(t, set)
		if !diskLookup(t, set, fp) {
			t.Fatal("missing fingerprint on disk after flush")
		}
	})
	run("testDiskLookupWithMax", func(t *testing.T, dir string) {
		const fp uint64 = math.MaxInt64
		set := getSet(t, dir, NewFPSetConfiguration())
		if set.memInsert(fp) {
			t.Fatal("memInsert unexpectedly duplicated fingerprint")
		}
		if diskLookup(t, set, fp) {
			t.Fatal("unexpected fingerprint on disk")
		}
		flush(t, set)
		if !diskLookup(t, set, fp) {
			t.Fatal("missing fingerprint on disk after flush")
		}
		if !set.memLookup(fp) {
			t.Fatal("missing fingerprint in memory after flush")
		}
	})
	for _, input := range []struct {
		name        string
		fp          uint64
		conditional bool
	}{
		{"testDiskLookupWithMaxOnPage", math.MaxInt64, false},
		{"testDiskLookupWithZerosOnPage", 0, true},
		{"testDiskLookupWithLongMinValueOnPage", uint64(1) << 63, true},
	} {
		run(input.name, func(t *testing.T, dir string) {
			if input.conditional && !known(t) {
				return
			}
			config := NewFPSetConfiguration()
			config.SetRatio(1)
			config.SetMemory(1000)
			set := getSet(t, dir, config)
			if set.Put(input.fp) {
				t.Fatal("first put unexpectedly duplicated fingerprint")
			}
			for i := int64(1); i < 1024*2; i++ {
				if !set.Put(input.fp) {
					t.Fatal("Failed to add fingerprint")
				}
				if !set.Contains(input.fp) {
					t.Fatal("missing fingerprint")
				}
			}
			fp0 := input.fp & diskFPSetFlushedMask
			if !set.memLookup(input.fp) {
				t.Fatal("missing fingerprint in memory")
			}
			if diskLookup(t, set, fp0) {
				t.Fatal("unexpected fingerprint on disk")
			}
		})
	}
	run("testComparePutAndPutBlock", func(t *testing.T, dir string) {
		putSet := getInitialized(t, dir)
		blockSet := getInitialized(t, dir)
		const fp int64 = 1
		vector := NewLongVec()
		vector.AddElement(fp)
		blockResult := !blockSet.PutBlock(vector).Get(0)
		if got := putSet.Put(uint64(fp)); got != blockResult {
			t.Fatalf("put=%v, putBlock=%v", got, blockResult)
		}
	})
	run("testCompareContainsAndContainsBlock", func(t *testing.T, dir string) {
		containsSet := getInitialized(t, dir)
		blockSet := getInitialized(t, dir)
		const fp int64 = 1
		vector := NewLongVec()
		vector.AddElement(fp)
		blockResult := !blockSet.ContainsBlock(vector).Get(0)
		if got := containsSet.Contains(uint64(fp)); got != blockResult {
			t.Fatalf("contains=%v, containsBlock=%v", got, blockResult)
		}
	})
	run("testContainsBlock", func(t *testing.T, dir string) {
		set := getInitialized(t, dir)
		const fp int64 = 1
		vector := NewLongVec()
		vector.AddElement(fp)
		if !set.ContainsBlock(vector).Get(0) {
			t.Fatal("missing absent-fingerprint bit")
		}
		set.Put(uint64(fp))
		if set.ContainsBlock(vector).Get(0) {
			t.Fatal("present-fingerprint bit unexpectedly set")
		}
	})
	run("testPutBlock", func(t *testing.T, dir string) {
		set := getInitialized(t, dir)
		vector := NewLongVec()
		vector.AddElement(1)
		if !set.PutBlock(vector).Get(0) {
			t.Fatal("first putBlock bit not set")
		}
		if set.PutBlock(vector).Get(0) {
			t.Fatal("second putBlock bit unexpectedly set")
		}
	})
}

// Whole original testCalculateMidEntry/isInvalidInput helpers. All1125 tuples
// reach this helper; preserve the source invalid-input filter and four assertions.
func javaShortDiskCalculateMidEntry(t *testing.T, set *javaDummyDiskFPSet, loVal, hiVal, fp, loEntry, hiEntry int64) {
	t.Helper()
	if loVal > hiVal || loVal > fp || hiVal < fp || loEntry >= hiEntry {
		return
	}
	defer func() {
		if failure := recover(); failure != nil {
			switch exception := failure.(type) {
			case *RuntimeException:
				t.Fatalf("failed to calculate for valid input (%d,%d,%d,%d,%d): %v", loVal, hiVal, fp, loEntry, hiEntry, exception)
			case *TLCError:
				if exception.Runtime {
					t.Fatalf("failed to calculate for valid input (%d,%d,%d,%d,%d): %v", loVal, hiVal, fp, loEntry, hiEntry, exception)
				}
			}
			panic(failure)
		}
	}()
	midEntry := set.calculateMidEntry(uint64(loVal), uint64(hiVal), float64(fp), loEntry, hiEntry)
	message := fmt.Sprintf("(loVal, hiVal, fp, loEntry, hiEntry, midEntry): %d, %d, %d, %d, %d, %d", loVal, hiVal, fp, loEntry, hiEntry, midEntry)
	if midEntry < 0 {
		t.Fatal("Negative mid entry " + message)
	}
	if midEntry < loEntry {
		t.Fatal("Not within lower bound " + message)
	}
	if midEntry > hiEntry {
		t.Fatal("Not within upper bound " + message)
	}
	if midEntry*8 < 0 {
		t.Fatal("midEntry turned negative " + message)
	}
}
