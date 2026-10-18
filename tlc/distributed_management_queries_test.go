package tlc

import (
	"math"
	"testing"
)

// No direct upstream methods test distributed management queries. Missing
// ownership is a failure; only the source's explicit inactive sentinels apply.
func distributedManagementQueries(w *TLCServerMXWrapper) map[string]func() {
	return map[string]func(){
		"generated": func() { w.GetStatesGenerated() }, "distinct": func() { w.GetDistinctStatesGenerated() },
		"queue": func() { w.GetStateQueueSize() }, "generatedRate": func() { w.GetStatesGeneratedPerMinute() },
		"distinctRate": func() { w.GetDistinctStatesGeneratedPerMinute() }, "depth": func() { w.GetProgress() },
		"workers": func() { w.GetWorkerCount() }, "blockAverage": func() { w.GetAverageBlockCnt() },
		"current": func() { w.GetCurrentState() }, "spec": func() { w.GetSpecName() }, "model": func() { w.GetModelName() },
	}
}

func TestDistributedManagementQueryMissingOwner(t *testing.T) {
	for _, wrapper := range []*TLCServerMXWrapper{nil, {}} {
		for name, query := range distributedManagementQueries(wrapper) {
			t.Run(name, func(t *testing.T) {
				requireCoordinatorNullFailure(t, func() error { query(); return nil })
			})
		}
	}
	server := &TLCServer{}
	wrapper := &TLCServerMXWrapper{Server: server}
	for name, query := range map[string]func(){
		"generated":    func() { wrapper.GetStatesGenerated() },
		"queue":        func() { wrapper.GetStateQueueSize() },
		"depth":        func() { wrapper.GetProgress() },
		"current":      func() { wrapper.GetCurrentState() },
		"blockAverage": func() { wrapper.GetAverageBlockCnt() },
	} {
		t.Run("component/"+name, func(t *testing.T) {
			requireCoordinatorNullFailure(t, func() error { query(); return nil })
		})
	}
	// Source explicitly tolerates a missing FP manager only for distinct count.
	if got := wrapper.GetDistinctStatesGenerated(); got != -1 {
		t.Fatalf("missing distinct manager = %d, want -1", got)
	}
	server.SetDone()
	if wrapper.GetStatesGenerated() != -1 || wrapper.GetDistinctStatesGenerated() != -1 || wrapper.GetProgress() != -1 || wrapper.GetSpecName() != "N/A" || wrapper.GetModelName() != "N/A" {
		t.Fatal("inactive queries accessed missing running components")
	}
	// Queue/current-state and selector queries remain unconditional when done.
	for _, query := range []func(){func() { wrapper.GetStateQueueSize() }, func() { wrapper.GetCurrentState() }, func() { wrapper.GetAverageBlockCnt() }} {
		requireCoordinatorNullFailure(t, func() error { query(); return nil })
	}
}

func TestDistributedManagementQueryCounters(t *testing.T) {
	queue := NewMemStateQueue()
	state := NewEmptyState()
	queue.Enqueue(state)
	set := NewMemFPSet()
	set.statesSeen = math.MaxInt64
	manager := NewDistributedFPSetManager(NewLocalFingerprintEndpoint(set))
	server := &TLCServer{StateQueue: queue, FPSetManager: manager, FileName: "Spec", ConfigName: "Config", StatesPerMinute: -7, DistinctStatesPerMinute: 11}
	server.WorkerStatesGenerated.Store(2)
	server.BlockSelector = NewStaticBlockSelector(server, 13)
	thread := &TLCServerThread{Worker: NewDistributedWorkerSmartProxy(&rpcTestWorker{})}
	thread.setStates([]*TLCStateMut{state, state})
	server.RegisterTLCServerThread(thread)
	wrapper := &TLCServerMXWrapper{Server: server}
	if got := wrapper.GetStatesGenerated(); got != math.MinInt64+2 {
		t.Fatalf("signed generated overflow = %d, want %d", got, int64(math.MinInt64+2))
	}
	if wrapper.GetStateQueueSize() != 3 || wrapper.GetWorkerCount() != 1 || wrapper.GetAverageBlockCnt() != 13 || wrapper.GetStatesGeneratedPerMinute() != -7 || wrapper.GetDistinctStatesGeneratedPerMinute() != 11 {
		t.Fatal("management counters lost queue/assigned work, selector or rate values")
	}
	if wrapper.GetSpecName() != "Spec" || wrapper.GetModelName() != "Config" {
		t.Fatal("active management names changed")
	}
	if wrapper.GetCurrentState() != state.String() || queue.Size() != 1 {
		t.Fatal("current state query changed the queue")
	}
	queue.Dequeue()
	if wrapper.GetCurrentState() != "N/A" || wrapper.GetStateQueueSize() != 2 {
		t.Fatal("empty queue lost the explicit current-state sentinel or assigned work")
	}
	server.SetDone()
	if wrapper.GetStateQueueSize() != 2 || wrapper.GetWorkerCount() != 1 || wrapper.GetAverageBlockCnt() != 13 || wrapper.GetStatesGeneratedPerMinute() != -7 || wrapper.GetDistinctStatesGeneratedPerMinute() != 11 {
		t.Fatal("inactive management query replaced unconditional counters")
	}
}
