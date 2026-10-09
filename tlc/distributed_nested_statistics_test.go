package tlc

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

type nestedSizeChild struct {
	*NoopFPSet
	id      int
	entered chan int
	release <-chan struct{}
	size    uint64
	failure error
}

func (s *nestedSizeChild) Size() uint64 {
	s.entered <- s.id
	<-s.release
	if s.failure != nil {
		panic(s.failure)
	}
	return s.size
}

// No source method directly checks parallel size traversal or its failure join.
func TestDistributedNestedSizeConcurrentAndJoined(t *testing.T) {
	for _, fails := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fails], func(t *testing.T) {
			entered := make(chan int, 2)
			release := make(chan struct{})
			ready := make(chan struct{})
			close(ready)
			first := &nestedSizeChild{NoopFPSet: NewNoopFPSet(nil), id: 0, entered: entered, release: release, size: math.MaxUint64}
			second := &nestedSizeChild{NoopFPSet: NewNoopFPSet(nil), id: 1, entered: entered, release: ready, size: 7}
			if fails {
				second.failure = errors.New("size failed")
			}
			set := &MultiFPSet{Sets: []FPSet{first, second}}
			type result struct {
				size    uint64
				failure any
			}
			done := make(chan result, 1)
			joined := make(chan struct{})
			var unblock sync.Once
			t.Cleanup(func() { unblock.Do(func() { close(release) }); <-joined })
			go func() {
				defer close(joined)
				var r result
				defer func() { r.failure = recover(); done <- r }()
				r.size = set.Size()
			}()
			seen := map[int]bool{}
			for range 2 {
				select {
				case id := <-entered:
					seen[id] = true
				case <-time.After(time.Second):
					t.Fatal("size calls did not start concurrently")
				}
			}
			if !seen[0] || !seen[1] {
				t.Fatal("size traversal lost a child")
			}
			select {
			case r := <-done:
				t.Fatalf("size returned before held child finished: %+v", r)
			default:
			}
			unblock.Do(func() { close(release) })
			r := <-done
			if fails {
				if r.failure != second.failure {
					t.Fatalf("failure identity changed: %v", r.failure)
				}
			} else if r.failure != nil || r.size != 6 {
				t.Fatalf("long sum overflow changed: %+v", r)
			}
		})
	}
	if (&MultiFPSet{}).Size() != 0 {
		t.Fatal("empty size sum changed")
	}
}

func TestDistributedNestedStatesSeenUsesParentCounter(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "tcp"}[remote], func(t *testing.T) {
			first, second := NewMemFPSet(), NewMemFPSet()
			set := &MultiFPSet{Sets: []FPSet{first, second}, Shift: 63}
			block := NewLongVec()
			block.AddElement(41)
			block.AddElement(43)
			first.ContainsBlock(block)
			block.AddElement(47)
			second.ContainsBlock(block)
			block.AddElement(math.MinInt64 + 49)
			set.ContainsBlock(block)
			if first.GetStatesSeen() != 2 || second.GetStatesSeen() != 3 {
				t.Fatal("fixture child counters changed during parent lookup")
			}
			var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(set)
			if remote {
				_, endpoint = startFingerprintRPC(t, endpoint)
			}
			if seen, err := endpoint.GetStatesSeen(); err != nil || seen != 4 {
				t.Fatalf("parent states seen = %d/%v, want 4", seen, err)
			}
			manager := NewDistributedFPSetManager(endpoint)
			if seen := manager.GetStatesSeen(); seen != 5 {
				t.Fatalf("manager states seen = %d, want source 1 + parent 4", seen)
			}
			if size := manager.Size(); size != 0 {
				t.Fatalf("lookups changed membership: %d", size)
			}
		})
	}
}
