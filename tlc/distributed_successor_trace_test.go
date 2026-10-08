package tlc

import "testing"

func TestDistributedAppRetainsToolSuccessorMetadata(t *testing.T) {
	predecessor := &TLCStateMut{UID: 23, level: 4}
	successor := &TLCStateMut{UID: TLCStateInitUID, level: 5}
	action := &Action{}
	tool := &Tool{
		GetNextStatesFunc: func(_ *Tool, gotAction *Action, gotState *TLCStateMut) (*StateVec, error) {
			if gotAction != action || gotState != predecessor {
				t.Fatal("application changed tool arguments")
			}
			states := NewStateVec(1)
			states.Add(successor)
			return states, nil
		},
		IsGoodStateFunc: func(_ *Tool, state *TLCStateMut) bool { return state == successor },
	}
	app := &TLCApp{Tool: tool, Actions: []*Action{action}}
	states, err := app.GetNextStates(predecessor)
	if err != nil || states.Size() != 1 || states.At(0) != successor {
		t.Fatalf("successor result %v, error %v", states, err)
	}
	if successor.TracePredecessor() != nil || successor.GetAction() != nil || successor.Level() != 5 || successor.UID != TLCStateInitUID {
		t.Fatal("TLCApp injected metadata absent in Java TLCApp.getNextStates")
	}
	if _, err := EncodeDistributedStates([]*TLCStateMut{successor}); err != nil {
		t.Fatalf("ordinary tool successor cannot cross the native boundary: %v", err)
	}
}

func TestDistributedPublicationWritesIncomingPredecessorUID(t *testing.T) {
	trace := NewTLCTrace()
	manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
	thread := &TLCServerThread{Server: &TLCServer{Trace: trace, FPSetManager: manager}}
	state := &TLCStateMut{UID: 314, level: 6}
	states := NewStateVec(1)
	states.Add(state)
	fps := NewLongVecFrom([]int64{0x1234})
	queue := NewMemStateQueue()
	if err := thread.publishBlock(queue, []*StateVec{states}, []*LongVec{fps}); err != nil {
		t.Fatal(err)
	}
	if queue.Size() != 1 || queue.SDequeue() != state {
		t.Fatal("published successor was not queued")
	}
	records := trace.Records()
	if len(records) != 1 || records[0].PreviousUID != 314 || records[0].FP != 0x1234 {
		t.Fatalf("trace lost incoming predecessor location: %#v", records)
	}
	if state.UID != 0 || state.Level() != 6 || state.TracePredecessor() != nil || state.GetAction() != nil {
		t.Fatal("publication must replace UID without changing source state metadata")
	}
	if _, err := EncodeDistributedStates([]*TLCStateMut{state}); err != nil {
		t.Fatalf("published state cannot return to a worker: %v", err)
	}
}

// Upstream has no enabled direct publication-failure test. These checks retain
// its fail-fast dereferences after FP insertion instead of silently losing work.
func TestDistributedPublicationRejectsMissingSelectedStates(t *testing.T) {
	for _, test := range []struct {
		name   string
		states []*StateVec
	}{
		{"null_array", nil},
		{"missing_partition", []*StateVec{}},
		{"null_partition", []*StateVec{nil}},
		{"missing_state", []*StateVec{NewStateVecFrom(nil)}},
		{"null_state", []*StateVec{NewStateVecFrom([]*TLCStateMut{nil})}},
	} {
		t.Run(test.name, func(t *testing.T) {
			trace := NewTLCTrace()
			manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
			thread := &TLCServerThread{Server: &TLCServer{Trace: trace, FPSetManager: manager}}
			queue := NewMemStateQueue()
			err := invokeDistributedServerOperation(func() error {
				return thread.publishBlock(queue, test.states, []*LongVec{NewLongVecFrom([]int64{71})})
			})
			if test.name == "missing_partition" || test.name == "missing_state" {
				if _, ok := err.(*ArrayIndexOutOfBoundsException); !ok {
					t.Fatalf("missing selected state failure: %T %v", err, err)
				}
			} else if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("null selected state failure: %T %v", err, err)
			}
			if manager.Size() != 1 || queue.Size() != 0 || len(trace.Records()) != 0 {
				t.Fatal("publication changed source insertion/failure ordering")
			}
		})
	}
	// No selected bits means Java never dereferences the absent state array.
	manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
	manager.Put(71)
	thread := &TLCServerThread{Server: &TLCServer{Trace: NewTLCTrace(), FPSetManager: manager}}
	if err := thread.publishBlock(NewMemStateQueue(), nil, []*LongVec{NewLongVecFrom([]int64{71})}); err != nil {
		t.Fatal(err)
	}
}

type publicationCountingQueue struct {
	*MemStateQueue
	dequeues int
}

type publicationNullBitsEndpoint struct{ *LocalFingerprintEndpoint }

func (*publicationNullBitsEndpoint) PutBlock(*LongVec) (*BitVector, error) {
	return nil, nil
}

func TestDistributedPublicationRejectsNullVisitedVector(t *testing.T) {
	endpoint := &publicationNullBitsEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet())}
	thread := &TLCServerThread{Server: &TLCServer{Trace: NewTLCTrace(), FPSetManager: NewDistributedFPSetManager(endpoint)}}
	err := invokeDistributedServerOperation(func() error {
		return thread.publishBlock(NewMemStateQueue(), nil, []*LongVec{NewLongVecFrom([]int64{71})})
	})
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("null visited vector was treated as no selected states: %T %v", err, err)
	}
}

func (q *publicationCountingQueue) SDequeueMany(count int) []*TLCStateMut {
	q.dequeues++
	return q.MemStateQueue.SDequeueMany(count)
}

func TestDistributedTracePublicationFailureStopsServerThread(t *testing.T) {
	oldWorkers := NumWorkers()
	defer SetNumWorkers(oldWorkers)
	SetNumWorkers(0)
	queue := &publicationCountingQueue{MemStateQueue: NewMemStateQueue()}
	queue.Enqueue(&TLCStateMut{level: 1})
	failure := NewIOException("trace write failed")
	trace := NewTLCTrace()
	trace.traceErr = failure
	manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
	server := NewTLCServer("Spec", "Spec", t.TempDir(), manager, queue, trace)
	worker := &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) {
		return NewNextStateResult([]*StateVec{NewStateVecFrom([]*TLCStateMut{{level: 2}})}, []*LongVec{NewLongVecFrom([]int64{71})}, 0, 1), nil
	}}
	thread := &TLCServerThread{Server: server, Worker: NewDistributedWorkerSmartProxy(worker), Selector: NewStaticBlockSelector(server, 1)}
	thread.TimerTask = &TLCTimerTask{Thread: thread}
	thread.Run()
	if server.LastError != failure || !server.Done.Load() || !queue.finish.Load() || queue.dequeues != 1 {
		t.Fatalf("trace failure did not terminate the server thread: error %v, done %v, finished %v, dequeues %d", server.LastError, server.Done.Load(), queue.finish.Load(), queue.dequeues)
	}
	if queue.Size() != 0 || len(trace.Records()) != 0 || manager.Size() != 1 {
		t.Fatal("trace failure queued or recorded an unpublished successor")
	}
}
