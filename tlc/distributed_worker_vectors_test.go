package tlc

import "testing"

// Upstream has no direct TLCStateVec tests. Distributed worker partitions use
// this unbounded collection, unlike the tool's SetBound-limited StateVec.
func distributedPartitionWorker(t *testing.T) (*DistributedWorker, *TLCStateMut) {
	t.Helper()
	previousTool := stateTool
	stateTool = nil
	t.Cleanup(func() { stateTool = previousTool })
	Globals.Lock()
	previousBound := Globals.SetBound
	Globals.SetBound = 1
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.SetBound = previousBound
		Globals.Unlock()
	})
	successors := NewStateVec(11)
	for i := 1; i <= 11; i++ {
		successors.Add(&TLCStateMut{values: []Value{NewIntValue(int32(i))}})
	}
	tool := &Tool{Actions: []*Action{{}},
		GetNextStatesFunc: func(*Tool, *Action, *TLCStateMut) (*StateVec, error) { return successors, nil },
		IsGoodStateFunc:   func(*Tool, *TLCStateMut) bool { return true },
	}
	worker := NewDistributedWorker(0, tool, NewDistributedFPSetManagerFromFPSet(NewMemFPSet()))
	t.Cleanup(worker.Runtime.executor.Shutdown)
	return worker, &TLCStateMut{UID: 37}
}

func TestDistributedWorkerPartitionsIgnoreToolSetBound(t *testing.T) {
	worker, predecessor := distributedPartitionWorker(t)
	result, err := worker.GetNextStates([]*TLCStateMut{predecessor})
	if err != nil || result == nil {
		t.Fatalf("tool SetBound incorrectly limited distributed partitions: %v", err)
	}
	if result.StatesComputed != 11 || worker.OverallStatesComputed.Load() != 11 || worker.Computing.Load() || len(result.NextStates) != 1 || result.NextStates[0].Size() != 11 || result.NextFingerprints[0].Size() != 11 {
		t.Fatal("worker lost successor counts or retained computation state")
	}
	partition := result.NextStates[0]
	if cap(partition.states) != 20 || partition.At(19) != nil {
		t.Fatal("worker partition lost default capacity/doubling or unused-slot access")
	}
	for _, state := range partition.states {
		if state.(*TLCStateMut).UID != predecessor.UID {
			t.Fatal("worker did not retain predecessor trace identity")
		}
	}
	decoded := resultPayloadRoundTrip(t, result).NextStates[0]
	if cap(decoded.states) != 11 {
		t.Fatal("native result transferred spare vector slots")
	}
	decoded.Add(predecessor)
	if decoded.Size() != 12 || cap(decoded.states) != 22 || decoded.At(11) != predecessor || decoded.At(21) != nil {
		t.Fatal("decoded distributed vector lost unbounded growth or unused slots")
	}
}

type workerSelectedFingerprintEndpoint struct {
	*LocalFingerprintEndpoint
	index int
}

func (e *workerSelectedFingerprintEndpoint) ContainsBlock(*LongVec) (*BitVector, error) {
	answer := NewBitVector(0)
	answer.Set(e.index)
	return answer, nil
}

func TestDistributedWorkerFingerprintSelectionUsesBackingCapacity(t *testing.T) {
	for _, index := range []int{9, 10} {
		endpoint := &workerSelectedFingerprintEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), index: index}
		tool := &Tool{Actions: []*Action{{}}, GetNextStatesFunc: func(*Tool, *Action, *TLCStateMut) (*StateVec, error) { return NewStateVec(0), nil }}
		worker := NewDistributedWorker(0, tool, NewDistributedFPSetManager(endpoint))
		worker.App.checkDeadlock = false
		defer worker.Runtime.executor.Shutdown()
		checkCalls := 0
		failure := NewRuntimeException("selected unused state")
		worker.CheckStateFunc = func(predecessor, successor *TLCStateMut) error {
			checkCalls++
			if predecessor != nil || successor != nil {
				t.Fatal("unused selected slots did not contain null states")
			}
			return failure
		}
		predecessor := &TLCStateMut{UID: 23}
		result, err := worker.GetNextStates([]*TLCStateMut{predecessor})
		wrapped, ok := err.(*WorkerException)
		if result != nil || !ok || !wrapped.KeepCallStack || wrapped.State2 != nil || worker.Computing.Load() {
			t.Fatalf("selection %d lost failure context: %v", index, err)
		}
		if index == 9 {
			if checkCalls != 1 || wrapped.Cause != failure || wrapped.State1 != nil {
				t.Fatalf("unused slot failed before source state checks: %#v", wrapped)
			}
		} else if _, ok := wrapped.Cause.(*ArrayIndexOutOfBoundsException); !ok || checkCalls != 0 || wrapped.State1 != predecessor {
			t.Fatalf("out-of-capacity selection changed source failure: %#v", wrapped)
		}
	}
}

func TestWorkerRPCPartitionsIgnoreToolSetBound(t *testing.T) {
	worker, predecessor := distributedPartitionWorker(t)
	_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
	result, err := client.GetNextStates([]*TLCStateMut{predecessor})
	if err != nil || result == nil || len(result.NextStates) != 1 || result.NextStates[0].Size() != 11 || result.NextFingerprints[0].Size() != 11 || result.StatesComputed != 11 {
		t.Fatalf("native worker partition result = %v/%v", result, err)
	}
	partition := result.NextStates[0]
	if cap(partition.states) != 11 {
		t.Fatal("TCP result retained spare partition capacity")
	}
	for _, state := range partition.states {
		if state.(*TLCStateMut).UID != predecessor.UID {
			t.Fatal("TCP result lost predecessor trace identity")
		}
	}
	partition.Add(predecessor)
	if partition.Size() != 12 || cap(partition.states) != 22 || partition.At(21) != nil {
		t.Fatal("TCP result vector regained the tool SetBound limit")
	}
}

func TestDistributedPublicationRetainsUnusedPartitionNull(t *testing.T) {
	manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
	trace := NewTLCTrace()
	queue := NewMemStateQueue()
	thread := &TLCServerThread{Server: &TLCServer{Trace: trace, FPSetManager: manager}}
	// An in-process result retains capacity ten; a received result has only
	// active capacity. FP insertion still precedes the selected state's failure.
	failure := invokeDistributedServerOperation(func() error {
		return thread.publishBlock(queue, []*StateVec{newDistributedStateVec(10)}, []*LongVec{NewLongVecFrom([]int64{71})})
	})
	if _, ok := failure.(*NullPointerException); !ok {
		t.Fatalf("selected unused partition entry = %T/%v, want null failure", failure, failure)
	}
	if manager.Size() != 1 || queue.Size() != 0 || len(trace.Records()) != 0 {
		t.Fatal("partition failure changed FP-before-trace/queue publication order")
	}
}
