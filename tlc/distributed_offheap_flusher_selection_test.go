package tlc

import (
	"runtime"
	"testing"
	"time"
)

// Selector boundaries remain applicable after upstream issue #1470 was fixed.
func TestDistributedOffHeapFlusherFallsBackWhenPartitionsTooSmall(t *testing.T) {
	for _, threads := range []int{4, 48} {
		t.Run(intToDecimal(threads), func(t *testing.T) {
			set := &OffHeapDiskFPSet{NonCheckpointableDiskFPSet: &NonCheckpointableDiskFPSet{DiskFPSet: &DiskFPSet{}},
				array: NewLongArray(8192), probeLimit: 1024}
			previous := set.selectOffHeapConcurrentFlusher(1)
			if previous == nil {
				t.Fatal("eligible partition did not select a concurrent flusher")
			}
			previous.shutdown = true
			if selected := set.selectOffHeapConcurrentFlusher(threads); selected != nil {
				t.Fatal("small partitions retained a parallel flusher instead of selecting sequential preparation")
			}
			fresh := set.selectOffHeapConcurrentFlusher(3)
			if fresh == nil || fresh == previous || fresh.shutdown {
				t.Fatal("eligible selection did not create a fresh executor")
			}
			for _, failure := range fresh.invokeAll(func(int) {}) {
				if failure != nil {
					t.Fatal(failure)
				}
			}
		})
	}
}

// Ported from OffHeapDiskFPSetFlusherReuseTest at upstream 210df3af8.
func newEvictedOffHeapFPSet(t *testing.T) *OffHeapDiskFPSet {
	t.Helper()
	set := javaOffHeapInitialized(t, 8192)
	for _, fp := range []uint64{41, 97} {
		if set.Put(fp) {
			t.Fatal("new fingerprint already present")
		}
	}
	set.Evict()
	for _, fp := range []uint64{131, 197} {
		if set.Put(fp) {
			t.Fatal("new fingerprint already present after eviction")
		}
	}
	return set
}

func assertOffHeapFlusherFingerprints(t *testing.T, set *OffHeapDiskFPSet) {
	t.Helper()
	for _, fp := range []uint64{41, 97, 131, 197} {
		if !set.Contains(fp) {
			t.Fatalf("lost fingerprint %d", fp)
		}
	}
}

func TestOffHeapFlusherReuseCheckFPsAfterConcurrentEviction(t *testing.T) {
	old := NumWorkers()
	defer SetNumWorkers(old)
	set := newEvictedOffHeapFPSet(t)
	SetNumWorkers(48)
	set.CheckFPs()
	assertOffHeapFlusherFingerprints(t, set)
}

func TestOffHeapFlusherReuseCheckFPsWithConcurrentFlusherAfterConcurrentEviction(t *testing.T) {
	old := NumWorkers()
	defer SetNumWorkers(old)
	set := newEvictedOffHeapFPSet(t)
	SetNumWorkers(2)
	set.CheckFPs()
	assertOffHeapFlusherFingerprints(t, set)
}

// The Go set no longer has a flusher field: the default is structurally
// sequential. Exercise the inherited default dispatch after concurrent eviction.
func TestOffHeapFlusherReuseDefaultFlusherSurvivesConcurrentEviction(t *testing.T) {
	set := newEvictedOffHeapFPSet(t)
	if err := set.flushOffHeapTable(); err != nil {
		t.Fatal(err)
	}
	assertOffHeapFlusherFingerprints(t, set)
}

func TestOffHeapFlusherReuseCheckInvariantAfterConcurrentEviction(t *testing.T) {
	set := newEvictedOffHeapFPSet(t)
	if !set.CheckInvariant(4) {
		t.Fatal("invariant failed")
	}
	assertOffHeapFlusherFingerprints(t, set)
}

func TestOffHeapFlusherReuseSecondEvictionAfterConcurrentEviction(t *testing.T) {
	set := newEvictedOffHeapFPSet(t)
	set.Evict()
	assertOffHeapFlusherFingerprints(t, set)
}

func TestOffHeapFlusherReuseCheckFPsLeavesNoThreadsBehind(t *testing.T) {
	set := newEvictedOffHeapFPSet(t)
	before := runtime.NumGoroutine()
	set.CheckFPs()
	// Go task workers are joined by invokeAll, rather than persistent JVM
	// executor threads. Allow unrelated runtime goroutines to settle.
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before {
		if time.Now().After(deadline) {
			t.Fatal("final check left goroutines behind")
		}
		runtime.Gosched()
	}
}
