//go:build tlc_fp_stress

package tlc

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// These original test-long heap methods take hours. Java customBuild.xml
// excludes their classes even from test-dist-long; this explicit target keeps
// their complete bodies available without adding them to ordinary Go runs.
// Run normally, with -tags=tlc_fp_stress and -timeout=0, never with -race.

// Complete FPSetTest.testMaxFPSetSizeRnd, inherited by the disk subclasses.
func javaLongFPSetRandomFull(t *testing.T, implementation string, factory func(*FPSetConfiguration) FPSet) {
	t.Helper()
	previousTime := time.Now()
	previousSize := uint64(0)
	var endTimeStamp time.Time
	t.Logf("Test started at %s", previousTime.Format("Mon Jan 02 15:04:05 MST 2006"))
	t.Cleanup(func() {
		if endTimeStamp.IsZero() {
			endTimeStamp = time.Now()
		}
		t.Logf("Test finished at %s", endTimeStamp.Format("Mon Jan 02 15:04:05 MST 2006"))
	})
	rnd := NewJavaRandom(15041980)
	set := factory(NewFPSetConfiguration())
	set.Init(1, t.TempDir(), "FPSetTestTest")
	t.Cleanup(set.Close)
	stats, hasStats := set.(interface {
		GetMaxTblCnt() int64
		GetLockCnt() int
		GetTblCapacity() int64
		GetLoadFactor() float64
	})
	if hasStats {
		t.Logf("Maximum FPSet table count is: %s (approx: %s GiB)", javaLongFPSetInteger(stats.GetMaxTblCnt()), javaLongFPSetInteger(stats.GetMaxTblCnt()*fpSetLongSize>>20))
		t.Logf("FPSet lock count is: %d", stats.GetLockCnt())
		t.Logf("FPSet bucket count is: %d", stats.GetTblCapacity())
	}
	t.Logf("Testing %s; full original random workload: 2147483648 iterations", implementation)
	predecessor := uint64(0)
	const limit = int64(2147483649)
	for i := int64(1); i < limit; i++ {
		if predecessor != 0 && !set.Contains(predecessor) {
			t.Fatalf("predecessor missing at iteration %d", i)
		}
		predecessor = uint64(rnd.NextLong())
		if set.Put(predecessor) {
			t.Fatalf("put unexpectedly present at iteration %d fingerprint %016x", i, predecessor)
		}
		currentSize := set.Size()
		if uint64(i) != currentSize {
			t.Fatalf("size=%d, want %d", currentSize, i)
		}
		now := time.Now()
		factor := float64(now.Sub(previousTime).Milliseconds()) / 60000
		if factor >= 1 {
			currentSize = set.Size()
			insertions := javaDoubleToLong(float64(int64(currentSize)-int64(previousSize)) * factor)
			if hasStats {
				load := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(stats.GetLoadFactor(), 'f', 2, 64), "0"), ".")
				t.Logf("%d s (epoch); %s insertions/min; %s load factor; %d/%d iterations", time.Now().UnixMilli(), javaLongFPSetInteger(insertions), load, i, limit-1)
			} else {
				t.Logf("%d s (epoch); %s insertions/min; %d/%d iterations", time.Now().UnixMilli(), javaLongFPSetInteger(insertions), i, limit-1)
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
	endTimeStamp = time.Now()
	if !set.CheckInvariant() {
		t.Fatal("checkInvariant failed")
	}
	if size := set.Size(); size != uint64(limit-1) {
		t.Fatalf("size=%d, want %d", size, limit-1)
	}
}

func TestJavaLongLSBDiskFPSet_testMaxFPSetSizeRnd(t *testing.T) {
	javaLongFPSetRandomFull(t, "tlc2.tool.fp.LSBDiskFPSet", func(config *FPSetConfiguration) FPSet {
		set := NewLSBDiskFPSet(config)
		t.Logf("DiskFPSet approx. consumes MiB: %d", set.GetMaxTblCnt()*fpSetLongSize>>20)
		return set
	})
}

func TestJavaLongMSBDiskFPSet_testMaxFPSetSizeRnd(t *testing.T) {
	javaLongFPSetRandomFull(t, "tlc2.tool.fp.MSBDiskFPSet", func(config *FPSetConfiguration) FPSet {
		return NewMSBDiskFPSet(config)
	})
}

// AbstractFPSetTest uses DecimalFormat("###,###.###") for these integer counts.
func javaLongFPSetInteger(value int64) string {
	text := strconv.FormatInt(value, 10)
	start := 0
	if text[0] == '-' {
		start = 1
	}
	for at := len(text) - 3; at > start; at -= 3 {
		text = text[:at] + "," + text[at:]
	}
	return text
}
