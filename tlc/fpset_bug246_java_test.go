// Copyright (c) Jan 9, 2012 Microsoft Corporation. All rights reserved.
package tlc

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"time"
)

// Whole original Bug246DiskFPSetTest.testLinearFillup. The loop uses the actual
// runtime-derived capacity, without a substituted fixed memory budget or sample.
func TestJavaBug246LinearFillup(t *testing.T) {
	vmMaxMemory := tlcRuntimeMaxHeapMemoryBytes()
	maxMemoryInBytes := tlcRuntimeFPMemSize(0.5)
	if vmMaxMemory <= maxMemoryInBytes {
		t.Fatal("Not enough memory dedicated to JVM, increase -Vmx value")
	}
	configuration := NewFPSetConfiguration()
	configuration.SetMemory(maxMemoryInBytes)
	configuration.SetRatio(1)
	fpSet := &javaDummyDiskFPSet{NewLSBDiskFPSet(configuration)}
	fpSet.Init(0, t.TempDir(), fmt.Sprintf("tlc2.tool.fp.Bug246DiskFPSetTest%d", time.Now().UnixMilli()))
	t.Cleanup(func() {
		if fpSet != nil {
			fpSet.Close()
		}
	})
	t.Logf("Runtime heap limit=%d, fingerprint memory=%d, table capacity=%d", vmMaxMemory, maxMemoryInBytes, fpSet.GetTblCapacity())
	var bucketCapacity, tblCapacity, tblLoad, tblCnt int64
	growDiskMark := 0
	func() {
		defer func() {
			if failure := recover(); failure != nil {
				if _, ok := failure.(*OutOfMemoryError); !ok {
					panic(failure)
				}
				fpSet = nil
				runtime.GC()
				if growDiskMark != 0 {
					t.Fatal("Expect not to have flushed to disk")
				}
				t.Fatalf("OOM occurred (not flush to disk) Bucket Capacity: %dTbl Capacity: %dTbl Load: %dTbl Cnt: %d", bucketCapacity, tblCapacity, tblLoad, tblCnt)
			}
		}()
		for i := 0; int64(i) < fpSet.GetTblCapacity()-1; i++ {
			fp := uint64(math.MaxInt64 - int64(i))
			if fpSet.Put(fp) {
				t.Fatalf("Unexpected duplicated: %d", fp)
			}
			bucketCapacity = fpSet.GetBucketCapacity()
			tblLoad = fpSet.GetTblLoad()
			tblCapacity = fpSet.GetTblCapacity()
			tblCnt = fpSet.GetTblCnt()
			growDiskMark = fpSet.GetGrowDiskMark()
		}
	}()
}

// Whole original Bug246DiskFPSetTest.testFlushDiskFPSet has no active statements:
// all its original body is commented out. Do not invent assertions or flushes.
func TestJavaBug246FlushDiskFPSet(t *testing.T) {}
