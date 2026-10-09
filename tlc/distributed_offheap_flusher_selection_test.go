package tlc

import "testing"

// Deliberate divergence from Java's stale-flusher reuse; see JAVA_BUG_FOUND.md.
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
			if selected := set.selectOffHeapConcurrentFlusher(threads); selected != nil || set.concurrentFlusher != nil {
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

// Exercise the actual merge that closes the previous executor, then the final
// check with small partitions. No synthetic shutdown flag is used here.
func TestDistributedOffHeapFinalCheckAfterConcurrentFlush(t *testing.T) {
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	set := javaOffHeapInitialized(t, 8192)
	for _, fp := range []uint64{41, 97} {
		if set.Put(fp) {
			t.Fatal("new fingerprint already present")
		}
	}
	if err := set.evict(); err != nil {
		t.Fatal(err)
	}
	previous := set.concurrentFlusher
	if previous == nil || !previous.shutdown || set.GetFileCnt() != 2 {
		t.Fatal("actual concurrent merge did not commit and shut down its executor")
	}
	for _, fp := range []uint64{131, 197} {
		if set.Put(fp) {
			t.Fatal("new fingerprint already present after merge")
		}
	}
	SetNumWorkers(48)
	if distance := set.CheckFPs(); distance == 0 || distance == 1<<63-1 {
		t.Fatalf("final fingerprint distance = %d", distance)
	}
	if set.concurrentFlusher != nil || set.Size() != 4 {
		t.Fatal("final check retained an ineligible executor or changed membership count")
	}
	for _, fp := range []uint64{41, 97, 131, 197} {
		if !set.Contains(fp) {
			t.Fatalf("final check lost fingerprint %d", fp)
		}
	}
}
