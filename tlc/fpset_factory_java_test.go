// Copyright (c) 2012 Markus Alexander Kuppe. All rights reserved.
package tlc

import (
	"strconv"
	"testing"
)

const javaFPSetFactoryMemory int64 = 64 * 1024 * 1024

// Original FPSetFactoryTest.testGetDiskFPSet, including all class classifications.
func TestJavaFPSetFactoryGetDiskFPSet(t *testing.T) {
	for _, name := range []string{"DiskFPSet", "HeapBasedDiskFPSet", "OffHeapDiskFPSet", "LSBDiskFPSet", "MSBDiskFPSet"} {
		if !isDiskFPSetImplementation("tlc2.tool.fp." + name) {
			t.Fatalf("%s must be a DiskFPSet", name)
		}
	}
	for _, name := range []string{"tlc2.tool.fp.FPSet", "tlc2.tool.distributed.fp.FPSetRMI", "tlc2.tool.fp.MultiFPSet", "tlc2.tool.fp.MemFPSet", "tlc2.tool.fp.MemFPSet1", "tlc2.tool.fp.MemFPSet2"} {
		if isDiskFPSetImplementation(name) {
			t.Fatalf("%s must not be a DiskFPSet", name)
		}
	}
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	config := NewFPSetConfiguration()
	if !config.AllowsNesting() {
		t.Fatal("configuration must allow nesting")
	}
}

// Original FPSetFactoryTest.testGetFPSetMSB.
func TestJavaFPSetFactoryGetFPSetMSB(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	config := NewFPSetConfiguration()
	javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MSBDiskFPSet", config)
}

// Original FPSetFactoryTest.testGetFPSetLSB.
func TestJavaFPSetFactoryGetFPSetLSB(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	config := NewFPSetConfiguration()
	javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.LSBDiskFPSet", config)
}

// Original FPSetFactoryTest.testGetFPSetOffHeap.
func TestJavaFPSetFactoryGetFPSetOffHeap(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("original Assume: BIT_64")
	}
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.OffHeapDiskFPSet")
	config := NewFPSetConfiguration()
	javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.OffHeapDiskFPSet", config)
}

// Original FPSetFactoryTest.testGetFPSetMSBWithMem.
func TestJavaFPSetFactoryGetFPSetMSBWithMem(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetMemory(javaFPSetFactoryMemory)
	config.SetRatio(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MSBDiskFPSet", config)
	if set.GetConfiguration().GetMemoryInBytes() != javaFPSetFactoryMemory {
		t.Fatalf("memory=%d, want %d", set.GetConfiguration().GetMemoryInBytes(), javaFPSetFactoryMemory)
	}
	javaFactoryDoTestNested(t, "tlc2.tool.fp.MSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetLSBWithMem.
func TestJavaFPSetFactoryGetFPSetLSBWithMem(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetMemory(javaFPSetFactoryMemory)
	config.SetRatio(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.LSBDiskFPSet", config)
	if set.GetConfiguration().GetMemoryInBytes() != javaFPSetFactoryMemory {
		t.Fatalf("memory=%d, want %d", set.GetConfiguration().GetMemoryInBytes(), javaFPSetFactoryMemory)
	}
	javaFactoryDoTestNested(t, "tlc2.tool.fp.LSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetMSBWithMemAndRatio.
func TestJavaFPSetFactoryGetFPSetMSBWithMemAndRatio(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetMemory(javaFPSetFactoryMemory)
	config.SetRatio(0.5)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MSBDiskFPSet", config)
	if set.GetConfiguration().GetMemoryInBytes() != javaFPSetFactoryMemory/2 {
		t.Fatalf("memory=%d, want %d", set.GetConfiguration().GetMemoryInBytes(), javaFPSetFactoryMemory/2)
	}
	javaFactoryDoTestNested(t, "tlc2.tool.fp.MSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetLSBWithMemAndRatio.
func TestJavaFPSetFactoryGetFPSetLSBWithMemAndRatio(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetMemory(javaFPSetFactoryMemory)
	config.SetRatio(0.5)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.LSBDiskFPSet", config)
	if set.GetConfiguration().GetMemoryInBytes() != javaFPSetFactoryMemory/2 {
		t.Fatalf("memory=%d, want %d", set.GetConfiguration().GetMemoryInBytes(), javaFPSetFactoryMemory/2)
	}
	javaFactoryDoTestNested(t, "tlc2.tool.fp.LSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetMultiFPSet.
func TestJavaFPSetFactoryGetFPSetMultiFPSet(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetFPBits(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MultiFPSet", config)
	javaFactoryDoTestNested(t, "tlc2.tool.fp.MSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetLSBMultiFPSet.
func TestJavaFPSetFactoryGetFPSetLSBMultiFPSet(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetFPBits(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MultiFPSet", config)
	javaFactoryDoTestNested(t, "tlc2.tool.fp.LSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetOffHeapMultiFPSet.
func TestJavaFPSetFactoryGetFPSetOffHeapMultiFPSet(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("original Assume: BIT_64")
	}
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.OffHeapDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetFPBits(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MultiFPSet", config)
	javaFactoryDoTestNested(t, "tlc2.tool.fp.OffHeapDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetMultiFPSetWithMem.
func TestJavaFPSetFactoryGetFPSetMultiFPSetWithMem(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetMemory(javaFPSetFactoryMemory)
	config.SetFPBits(1)
	config.SetRatio(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MultiFPSet", config)
	javaFactoryDoTestNested(t, "tlc2.tool.fp.MSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetLSBMultiFPSetWithMem.
func TestJavaFPSetFactoryGetFPSetLSBMultiFPSetWithMem(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.LSBDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetMemory(javaFPSetFactoryMemory)
	config.SetFPBits(1)
	config.SetRatio(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MultiFPSet", config)
	javaFactoryDoTestNested(t, "tlc2.tool.fp.LSBDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetOffHeapMultiFPSetWithMem.
func TestJavaFPSetFactoryGetFPSetOffHeapMultiFPSetWithMem(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("original Assume: BIT_64")
	}
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.OffHeapDiskFPSet")
	config := NewFPSetConfiguration()
	config.SetMemory(javaFPSetFactoryMemory)
	config.SetFPBits(1)
	config.SetRatio(1)
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MultiFPSet", config)
	javaFactoryDoTestNested(t, "tlc2.tool.fp.OffHeapDiskFPSet", config, set.(*MultiFPSet))
}

// Original FPSetFactoryTest.testGetFPSetOffHeapMultiFPSet42 (no architecture assumption).
func TestJavaFPSetFactoryGetFPSetOffHeapMultiFPSet42(t *testing.T) {
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.OffHeapDiskFPSet")
	nonHeapPhysicalMemory := tlcRuntimeNonHeapPhysicalMemory()
	config := NewFPSetConfiguration()
	javaFactoryEqual(t, nonHeapPhysicalMemory, config.GetMemoryInBytes())
	javaFactoryEqual(t, nonHeapPhysicalMemory/fpSetLongSize, config.GetMemoryInFingerprintCnt())
	set := javaFactoryDoTestGetFPSet(t, "tlc2.tool.fp.MultiFPSet", config).(*MultiFPSet)
	multiConfig := set.GetConfiguration()
	if multiConfig.GetImplementation() != "tlc2.tool.fp.OffHeapDiskFPSet" {
		t.Fatal("wrong parent implementation")
	}
	javaFactoryEqual(t, 1, int64(multiConfig.GetFPBits()))
	javaFactoryEqual(t, 2, int64(multiConfig.GetMultiFPSetCnt()))
	javaFactoryEqual(t, nonHeapPhysicalMemory, multiConfig.GetMemoryInBytes())
	javaFactoryEqual(t, nonHeapPhysicalMemory/fpSetLongSize, multiConfig.GetMemoryInFingerprintCnt())
	javaFactoryEqual(t, 2, int64(len(set.Sets)))
	for _, nested := range set.Sets {
		off := nested.(*OffHeapDiskFPSet)
		child := off.GetConfiguration()
		if child.GetImplementation() != "tlc2.tool.fp.OffHeapDiskFPSet" {
			t.Fatal("wrong nested implementation")
		}
		javaFactoryEqual(t, 1, int64(child.GetFPBits()))
		javaFactoryEqual(t, 2, int64(child.GetMultiFPSetCnt()))
		javaFactoryEqual(t, nonHeapPhysicalMemory/2, child.GetMemoryInBytes())
		javaFactoryEqual(t, (nonHeapPhysicalMemory/fpSetLongSize)/2, child.GetMemoryInFingerprintCnt())
		javaFactoryEqual(t, child.GetMemoryInBytes()/fpSetLongSize, child.GetMemoryInFingerprintCnt())
	}
}

// Original doTestGetFPSet only asserts assignability for a non-disk expected class.
func javaFactoryDoTestGetFPSet(t *testing.T, expected string, config *FPSetConfiguration) FPSet {
	t.Helper()
	set := NewFPSet(config)
	if !isDiskFPSetImplementation(expected) && !javaFactoryAssignable(expected, set) {
		t.Fatalf("%T not assignable to %s", set, expected)
	}
	return set
}

func javaFactoryDoTestNested(t *testing.T, expected string, config *FPSetConfiguration, set *MultiFPSet) {
	t.Helper()
	javaFactoryEqual(t, int64(config.GetMultiFPSetCnt()), int64(len(set.Sets)))
	var memoryInBytes int64
	for _, child := range set.Sets {
		if !javaFactoryAssignable(expected, child) {
			t.Fatalf("%T not assignable to %s", child, expected)
		}
		memoryInBytes += child.GetConfiguration().GetMemoryInBytes()
		if stats, ok := child.(interface{ GetMaxTblCnt() int64 }); ok {
			maxTblCnt := stats.GetMaxTblCnt()
			if child.GetConfiguration().GetMemoryInFingerprintCnt() < maxTblCnt {
				t.Fatal("Nested FPSet has over-allocated memory.")
			}
			if child.GetConfiguration().GetMemoryInFingerprintCnt() < maxTblCnt {
				t.Fatal("fingerprint budget below maxTblCnt")
			}
		}
	}
	javaFactoryEqual(t, set.GetConfiguration().GetMemoryInBytes(), memoryInBytes)
}

func javaFactoryEqual(t *testing.T, want, got int64) {
	t.Helper()
	if got != want {
		t.Fatalf("got=%d, want=%d", got, want)
	}
}

// These are the concrete source classes requested by the original methods.
func javaFactoryAssignable(expected string, set FPSet) bool {
	switch expected {
	case "tlc2.tool.fp.MultiFPSet":
		_, ok := set.(*MultiFPSet)
		return ok
	case "tlc2.tool.fp.MSBDiskFPSet":
		_, ok := set.(*MSBDiskFPSet)
		return ok
	case "tlc2.tool.fp.LSBDiskFPSet":
		_, ok := set.(*LSBDiskFPSet)
		return ok
	case "tlc2.tool.fp.OffHeapDiskFPSet":
		_, ok := set.(*OffHeapDiskFPSet)
		return ok
	default:
		panic("unexpected original factory class: " + expected)
	}
}
