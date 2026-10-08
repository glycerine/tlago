package tlc

import (
	"math"
	"testing"
)

// Upstream has no selector test methods. These checks retain the source
// constructor, arithmetic and limit contracts at the native Go boundary.
func TestDistributedBlockSelectorRequiresServer(t *testing.T) {
	for _, constructor := range []func(*TLCServer) *BlockSelector{
		NewBlockSelector, NewProportionalBlockSelector, NewStatisticalBlockSelector,
		func(s *TLCServer) *BlockSelector { return NewLimitingBlockSelector(s) },
		func(s *TLCServer) *BlockSelector { return NewStaticBlockSelector(s) },
	} {
		func() {
			defer func() {
				failure, ok := recover().(*TLCError)
				if !ok || !failure.Runtime || failure.Error() != "TLC found a null TLCServer" {
					t.Fatalf("null server assertion = %v", failure)
				}
			}()
			constructor(nil)
		}()
	}
}

func TestDistributedBlockSelectorArithmetic(t *testing.T) {
	server := &TLCServer{}
	proportional := NewProportionalBlockSelector(server)
	for _, test := range []struct{ size, want int64 }{
		{0, 0}, {1, math.MaxInt64}, {-1, math.MinInt64}, {math.MaxInt64, math.MaxInt64},
	} {
		if got := proportional.getBlockSize(test.size, nil); got != test.want {
			t.Errorf("zero-worker selection(%d) = %d, want %d", test.size, got, test.want)
		}
	}
	server.threadsToWorkers = NewInsMap[*TLCServerThread, DistributedWorkerEndpoint]()
	server.threadsToWorkers.Set(&TLCServerThread{}, nil)
	server.threadsToWorkers.Set(&TLCServerThread{}, nil)
	if got := proportional.getBlockSize(5, nil); got != 3 {
		t.Errorf("two-worker selection = %d, want 3", got)
	}
	if got := proportional.getBlockSize(math.MaxInt64, nil); got != 1<<62 {
		t.Errorf("large selection = %d, want %d", got, int64(1<<62))
	}
	statistical := NewStatisticalBlockSelector(server)
	for _, test := range []struct {
		size     int64
		overhead float64
		want     int64
	}{
		{0, math.Inf(1), 0}, {10, math.NaN(), 0}, {10, math.Inf(1), 8192},
		{10, math.Inf(-1), 8192}, {10, 0, 1}, {10, 0.025, 10},
		{10, -0.025, 10}, {math.MaxInt64, math.MaxFloat64, 8192},
	} {
		worker := &DistributedWorkerSmartProxy{NetworkOverhead: test.overhead}
		if got := statistical.getBlockSize(test.size, worker); got != test.want {
			t.Errorf("statistical selection(%d, %v) = %d, want %d", test.size, test.overhead, got, test.want)
		}
	}
	if got := statistical.getBlockSize(5, nil); got != 3 {
		t.Errorf("statistical fallback = %d, want 3", got)
	}
}

func TestDistributedBlockSelectorLimits(t *testing.T) {
	server := &TLCServer{}
	for _, selector := range []*BlockSelector{NewProportionalBlockSelector(server), NewStaticBlockSelector(server)} {
		before := selector.Maximum
		selector.SetMaxTXSize(7)
		if selector.Maximum != before {
			t.Fatal("non-limiting selector accepted a transfer limit")
		}
	}
	for _, selector := range []*BlockSelector{NewLimitingBlockSelector(server), NewStatisticalBlockSelector(server)} {
		selector.SetMaxTXSize(7)
		if got := selector.getBlockSize(100, nil); got != 7 {
			t.Fatalf("limited zero-worker selection = %d, want 7", got)
		}
	}
}

type distributedSelectorQueue struct {
	StateQueue
	size      int64
	requested int
	result    []*TLCStateMut
}

func (q *distributedSelectorQueue) Size() int64 { return q.size }
func (q *distributedSelectorQueue) SDequeueMany(count int) []*TLCStateMut {
	q.requested = count
	return q.result
}

func TestDistributedBlockSelectorQueueBounds(t *testing.T) {
	server := &TLCServer{}
	for _, test := range []struct {
		size     int64
		overhead float64
		want     int
	}{
		{0, math.Inf(1), 1}, {20, math.NaN(), 1}, {20, math.Inf(1), 20},
		{math.MaxInt64, math.Inf(1), 8192},
	} {
		selector := NewStatisticalBlockSelector(server)
		queue := &distributedSelectorQueue{size: test.size, result: []*TLCStateMut{{}, {}}}
		got := selector.GetBlocks(queue, &DistributedWorkerSmartProxy{NetworkOverhead: test.overhead})
		if queue.requested != test.want || len(got) != 2 || selector.GetAverageBlockCnt() != 2 {
			t.Fatalf("queue selection(%d, %v) requested %d, returned %d, average %d", test.size, test.overhead, queue.requested, len(got), selector.GetAverageBlockCnt())
		}
		queue.result = nil
		if selector.GetBlocks(queue, nil) != nil || selector.GetAverageBlockCnt() != 1 {
			t.Fatal("null dequeue did not update actual block-size average")
		}
	}
	proportional := NewProportionalBlockSelector(server)
	queue := &distributedSelectorQueue{size: math.MaxInt64}
	proportional.GetBlocks(queue, nil)
	if queue.requested != math.MaxInt32 {
		t.Fatalf("queue bound = %d, want MaxInt32", queue.requested)
	}
	static := NewStaticBlockSelector(server, 7)
	queue.size = 2
	static.GetBlocks(queue, nil)
	if queue.requested != 7 || static.GetAverageBlockCnt() != 7 {
		t.Fatal("static selector did not retain its fixed request and average")
	}
}
