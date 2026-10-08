package tlc

import (
	"math"
	"sync"
	"testing"
)

func TestDistributedSelectorAverageSemantics(t *testing.T) {
	selector := NewProportionalBlockSelector(&TLCServer{})
	for _, test := range []struct{ count, want int64 }{
		{10, 10}, {0, 5}, {0, 2}, {0, 1}, {0, 0},
		{math.MaxInt64, math.MaxInt64}, {math.MaxInt64, -1}, {7, 7},
	} {
		selector.setAverageBlockCnt(test.count)
		if got := selector.GetAverageBlockCnt(); got != test.want {
			t.Fatalf("average after count %d = %d, want %d", test.count, got, test.want)
		}
	}
	static := NewStaticBlockSelector(&TLCServer{}, 7)
	static.setAverageBlockCnt(100)
	if static.GetAverageBlockCnt() != 7 {
		t.Fatal("static average changed")
	}
}

// No original method checks concurrent selector statistics. Keep this native
// race check short; the source average intentionally permits lost updates.
func TestDistributedSelectorConcurrentStatistics(t *testing.T) {
	selector := NewLimitingBlockSelector(&TLCServer{})
	selector.SetMaxTXSize(64)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			<-start
			for i := 0; i < 1000; i++ {
				selector.setAverageBlockCnt(int64(1 + (i+worker)%64))
				selector.SetMaxTXSize(1 + (i+worker)%64)
				if average := selector.GetAverageBlockCnt(); average < 1 || average > 64 {
					t.Errorf("concurrent average = %d, want 1..64", average)
					return
				}
				if size := selector.getBlockSize(100, nil); size < 1 || size > 64 {
					t.Errorf("concurrent limit = %d, want 1..64", size)
					return
				}
			}
		}(worker)
	}
	close(start)
	workers.Wait()
}
