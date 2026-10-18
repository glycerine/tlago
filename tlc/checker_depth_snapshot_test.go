package tlc

import "testing"

type levelAdvancingFPSet struct {
	FPSet
	advance func()
}

func (s *levelAdvancingFPSet) Size() uint64 {
	s.advance()
	return s.FPSet.Size()
}

func TestModelCheckerDepthUsesLevelBeforeProgressReporting(t *testing.T) {
	worker := &Worker{}
	worker.SetLevel(1)
	queue := NewMemStateQueue()
	checker := &ModelChecker{
		AbstractChecker: &AbstractChecker{},
		StateQueue:      queue,
		ConcurrentTrace: &ConcurrentTLCTrace{TLCTrace: NewTLCTrace(), Workers: []*Worker{worker}},
		FPSet: &levelAdvancingFPSet{
			FPSet: NewMemFPSet(),
			advance: func() {
				worker.SetLevel(3)
			},
		},
	}
	if shouldWait := checker.RunTLCContinueDoing(1, 2); !shouldWait {
		t.Fatal("depth check used the level advanced during progress reporting")
	}
	if queue.finish.Load() || checker.Done.Load() || worker.GetMaxLevel() != 3 {
		t.Fatal("progress reporting did not preserve the pre-report depth decision")
	}
}
