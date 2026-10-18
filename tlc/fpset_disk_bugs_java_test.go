// Copyright (c) 2011 Microsoft Corporation. All rights reserved.
package tlc

import (
	"math"
	"testing"
	"time"
)

// Original test-side DummyDiskFPSet only broadens visibility of the index.
type javaDummyDiskFPSet struct{ *LSBDiskFPSet }

func (s *javaDummyDiskFPSet) setIndex(index []uint64) { s.index = index }

// Whole original Bug210DiskFPSetTest.testDiskLookupWithOverflow.
func TestJavaBug210DiskLookupWithOverflow(t *testing.T) {
	t.Logf("Test started at %s", time.Now())
	t.Cleanup(func() { t.Logf("Test finished at %s", time.Now()) })
	size := math.MaxInt32/diskFPSetNumEntriesPerPage + 8
	index := make([]uint64, size)
	index[size-2] = math.MaxInt64 - 3
	index[size-1] = math.MaxInt64 - 1
	set := &javaDummyDiskFPSet{NewLSBDiskFPSet(NewFPSetConfiguration())}
	set.Init(1, t.TempDir(), "FPSetTestTest")
	t.Cleanup(set.Close)
	set.setIndex(index)
	found, err := set.diskLookup(math.MaxInt64 - 2)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("nonexistent fingerprint found")
	}
}

// Bug242DiskFPSetTest.getFPSet(int) retains source configuration and constructor;
// its dummy is intentionally not initialized in these five original methods.
func javaBug242FPSet(memory int64) *javaDummyDiskFPSet {
	configuration := NewFPSetConfiguration()
	configuration.SetMemory(memory)
	configuration.SetRatio(1)
	return &javaDummyDiskFPSet{NewLSBDiskFPSet(configuration)}
}

// Source catch(Exception) must not swallow Java Error families.
func javaFPSetCatchException(failure any) (error, bool) {
	exception, ok := failure.(error)
	return exception, ok && !isJavaError(exception)
}

func javaBug242HighMemory(t *testing.T, memory int64) {
	t.Helper()
	t.Logf("Test started at %s", time.Now())
	t.Cleanup(func() { t.Logf("Test finished at %s", time.Now()) })
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(*OutOfMemoryError); ok {
				return
			}
			if exception, ok := javaFPSetCatchException(failure); ok {
				t.Fatal(exception)
			}
			panic(failure)
		}
	}()
	javaBug242FPSet(memory)
}

// Whole original Bug242DiskFPSetTest.testDiskFPSetWithHighMem.
func TestJavaBug242DiskFPSetWithHighMem(t *testing.T) { javaBug242HighMemory(t, 2097153638) }

// Whole original Bug242DiskFPSetTest.testDiskFPSetIntMaxValue.
func TestJavaBug242DiskFPSetIntMaxValue(t *testing.T) { javaBug242HighMemory(t, math.MaxInt32) }

// Whole original Bug242DiskFPSetTest.testDiskFPSetIntMinValue.
func TestJavaBug242DiskFPSetIntMinValue(t *testing.T) {
	t.Logf("Test started at %s", time.Now())
	t.Cleanup(func() { t.Logf("Test finished at %s", time.Now()) })
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := javaFPSetCatchException(failure); ok {
				return
			}
			panic(failure)
		}
	}()
	javaBug242FPSet(math.MinInt32)
	t.Fatal("negative memory did not throw an Exception")
}

func javaBug242SmallMemory(t *testing.T, memory int64) {
	t.Helper()
	t.Logf("Test started at %s", time.Now())
	t.Cleanup(func() { t.Logf("Test finished at %s", time.Now()) })
	defer func() {
		if failure := recover(); failure != nil {
			if exception, ok := javaFPSetCatchException(failure); ok {
				t.Fatal(exception)
			}
			panic(failure)
		}
	}()
	javaBug242FPSet(memory)
}

// Whole original Bug242DiskFPSetTest.testDiskFPSetZero.
func TestJavaBug242DiskFPSetZero(t *testing.T) { javaBug242SmallMemory(t, 0) }

// Whole original Bug242DiskFPSetTest.testDiskFPSetOne.
func TestJavaBug242DiskFPSetOne(t *testing.T) { javaBug242SmallMemory(t, 1) }
