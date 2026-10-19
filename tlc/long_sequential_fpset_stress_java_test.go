//go:build tlc_fp_stress

package tlc

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// Complete original test-long FPSetTest.testMaxFPSetSize. All three concrete
// factories inherit this body. Preserve the source's repeated fingerprint and
// final limit assertion, including their known contradictions; no skip or bound
// reduction. This is an explicit multi-billion-entry workload, never a race test.
func javaLongFPSetSequentialFull(t *testing.T, factory func(*FPSetConfiguration) FPSet) {
	t.Helper()
	previousTime := time.Now()
	previousSize := uint64(0)
	t.Logf("Test started at %s", previousTime.Format("Mon Jan 02 15:04:05 MST 2006"))
	t.Cleanup(func() { t.Logf("Test finished at %s", time.Now().Format("Mon Jan 02 15:04:05 MST 2006")) })
	set := factory(NewFPSetConfiguration())
	set.Init(1, t.TempDir(), "FPSetTestTest")
	t.Cleanup(set.Close)
	counter := int64(0)
	const limit = int64(2147483649)
	for i := (limit / 2) * -1; i < limit; i++ {
		value := int64(-1)
		if i%2 != 0 {
			value = limit - i
		} else {
			value = i
		}
		if set.Put(uint64(value)) {
			t.Fatalf("put unexpectedly present at i=%d fingerprint=%d", i, value)
		}
		currentSize := set.Size()
		counter++
		if uint64(counter) != currentSize {
			t.Fatalf("size=%d, want counter=%d at i=%d", currentSize, counter, i)
		}
		now := time.Now()
		factor := float64(now.Sub(previousTime).Milliseconds()) / 60000
		if factor >= 1 {
			currentSize = set.Size()
			insertions := javaDoubleToLong(float64(int64(currentSize)-int64(previousSize)) * factor)
			if stats, ok := set.(interface{ GetLoadFactor() float64 }); ok {
				load := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(stats.GetLoadFactor(), 'f', 2, 64), "0"), ".")
				t.Logf("%d s (epoch); %s insertions/min; %s load factor", time.Now().UnixMilli(), javaLongFPSetInteger(insertions), load)
			} else {
				t.Logf("%d s (epoch); %s insertions/min", time.Now().UnixMilli(), javaLongFPSetInteger(insertions))
			}
			previousTime, previousSize = now, currentSize
		}
	}
	if err := set.BeginChkpt(); err != nil {
		t.Fatal(err)
	}
	if err := set.CommitChkpt(); err != nil {
		t.Fatal(err)
	}
	if size := set.Size(); size != uint64(limit) {
		t.Fatalf("size=%d, want original limit=%d", size, limit)
	}
}

func TestJavaLongLSBDiskFPSet_testMaxFPSetSize(t *testing.T) {
	skipSlowTest(t)

	javaLongFPSetSequentialFull(t, func(config *FPSetConfiguration) FPSet {
		set := NewLSBDiskFPSet(config)
		t.Logf("DiskFPSet approx. consumes MiB: %d", set.GetMaxTblCnt()*fpSetLongSize>>20)
		return set
	})
}
func TestJavaLongMSBDiskFPSet_testMaxFPSetSize(t *testing.T) {
	skipSlowTest(t)

	javaLongFPSetSequentialFull(t, func(config *FPSetConfiguration) FPSet { return NewMSBDiskFPSet(config) })
}
func TestJavaOffHeapDiskFPSetLong_testMaxFPSetSize(t *testing.T) {
	skipSlowTest(t)

	javaLongFPSetSequentialFull(t, func(ignored *FPSetConfiguration) FPSet {
		return NewOffHeapDiskFPSet(NewFPSetConfigurationWithRatioAndImplementation(1.0, "tlc2.tool.fp.OffHeapDiskFPSet"))
	})
}
