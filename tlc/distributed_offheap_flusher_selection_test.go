package tlc

import "testing"

// OffHeapDiskFPSet.getFlusher returns the previous flusher when partitions
// would be too small, even if that flusher's executor has already shut down.
// This explains the upstream-disabled distributed model's final CheckFPs
// failure; resetting the flusher here would change the source algorithm.
func TestDistributedOffHeapFlusherRetainsIneligibleSelection(t *testing.T) {
	for _, threads := range []int{4, 48} {
		t.Run(intToDecimal(threads), func(t *testing.T) {
			set := &OffHeapDiskFPSet{NonCheckpointableDiskFPSet: &NonCheckpointableDiskFPSet{DiskFPSet: &DiskFPSet{}},
				array: NewLongArray(8192), probeLimit: 1024}
			previous := set.selectOffHeapConcurrentFlusher(1)
			if previous == nil {
				t.Fatal("eligible partition did not select a concurrent flusher")
			}
			previous.shutdown = true
			if selected := set.selectOffHeapConcurrentFlusher(threads); selected != previous {
				t.Fatal("ineligible partition replaced the source's retained flusher")
			}
			failure := invokeDistributedServerOperation(func() error {
				previous.invokeAll(func(int) { t.Error("closed executor ran a task") })
				return nil
			})
			if _, ok := failure.(*RejectedExecutionException); !ok {
				t.Fatalf("retained closed executor: %T %v", failure, failure)
			}
			fresh := set.selectOffHeapConcurrentFlusher(3)
			if fresh == nil || fresh == previous || fresh.shutdown {
				t.Fatal("eligible selection did not replace the old executor")
			}
			for _, failure := range fresh.invokeAll(func(int) {}) {
				if failure != nil {
					t.Fatal(failure)
				}
			}
		})
	}
}
