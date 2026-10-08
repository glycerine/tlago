package tlc

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type activeCheckpointQueue struct {
	*DiskStateQueue
	begin chan int64
}

func (q *activeCheckpointQueue) BeginChkpt() error {
	q.begin <- q.Size()
	return q.DiskStateQueue.BeginChkpt()
}

// No enabled direct Java test covers this boundary. Use the production server
// thread, TCP worker endpoint, disk queue, trace and local/remote FP recovery. Only
// remote evaluation is gated to make the assigned-block ordering observable.
func TestDistributedCheckpointWaitsForAssignedBlock(t *testing.T) {
	for _, remote := range []bool{false, true} {
		name := "local_fingerprints"
		if remote {
			name = "tcp_fingerprints"
		}
		t.Run(name, func(t *testing.T) { checkDistributedCheckpointWithAssignedBlock(t, remote) })
	}
}

func checkDistributedCheckpointWithAssignedBlock(t *testing.T, remote bool) {
	t.Helper()
	oldWorkers, oldVariables, oldVarCount, oldEmpty := NumWorkers(), stateVariables, UniqueStringVariableCount(), EmptyState
	t.Cleanup(func() {
		SetNumWorkers(oldWorkers)
		stateVariables = oldVariables
		SetUniqueStringVariableCount(oldVarCount)
		EmptyState = oldEmpty
	})
	SetNumWorkers(0)
	stateVariables = nil
	SetUniqueStringVariableCount(0)
	EmptyState = nil
	metadir := t.TempDir()
	queue := &activeCheckpointQueue{DiskStateQueue: NewDiskStateQueue(metadir), begin: make(chan int64, 1)}
	trace := NewTLCTrace(metadir, "Spec")
	set := NewMemFPSet()
	set.Init(1, metadir, "Spec")
	manager := NewNonDistributedFPSetManager(set, "local", trace)
	if remote {
		_, endpoint := startFingerprintRPC(t, NewLocalFingerprintEndpoint(set))
		manager = NewDistributedFPSetManager(endpoint)
	}
	server := NewTLCServer("Spec", "Spec", metadir, manager, queue, trace)
	initial := &TLCStateMut{UID: TLCStateInitUID, level: 1}
	if err := trace.WriteInitState(initial, 61); err != nil {
		t.Fatal(err)
	}
	manager.Put(61)
	queue.Enqueue(initial)
	assigned := make(chan int64, 1)
	continued := make(chan struct{}, 1)
	releaseFirst, releaseSecond := make(chan struct{}), make(chan struct{})
	var firstOnce, secondOnce sync.Once
	var calls atomic.Int32
	worker := &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if calls.Add(1) == 1 {
			assigned <- states[0].UID
			<-releaseFirst
			successor := &TLCStateMut{UID: states[0].UID, level: 2}
			return NewNextStateResult([]*StateVec{NewStateVecFrom([]*TLCStateMut{successor})}, []*LongVec{NewLongVecFrom([]int64{71})}, 0, 1), nil
		}
		continued <- struct{}{}
		<-releaseSecond
		return NewNextStateResult([]*StateVec{NewStateVec(0)}, []*LongVec{NewLongVec()}, 0, 0), nil
	}}
	_, endpoint := startWorkerRPC(t, worker)
	thread := &TLCServerThread{Server: server, Worker: NewDistributedWorkerSmartProxy(endpoint), Selector: NewStaticBlockSelector(server, 1)}
	threadDone := make(chan struct{})
	go func() { defer close(threadDone); thread.Run() }()
	var checkpointDone chan error
	t.Cleanup(func() {
		queue.FinishAll()
		firstOnce.Do(func() { close(releaseFirst) })
		secondOnce.Do(func() { close(releaseSecond) })
		<-threadDone
		if checkpointDone != nil {
			<-checkpointDone
		}
		_ = trace.Close()
	})
	select {
	case uid := <-assigned:
		if uid != initial.UID || queue.Size() != 0 || manager.Size() != 1 {
			t.Fatal("first block was not assigned from the original frontier")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("worker assignment watchdog expired")
	}
	checkpointDone = make(chan error, 1)
	go func() { checkpointDone <- server.Checkpoint() }()
	// Observe the actual queue stop flag, not a guessed evaluation delay.
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		queue.mu.Lock()
		stopped := queue.stop
		queue.mu.Unlock()
		if stopped {
			break
		}
		select {
		case <-ticker.C:
		case <-deadline.C:
			t.Fatal("queue suspension watchdog expired")
		}
	}
	select {
	case <-queue.begin:
		t.Fatal("checkpoint began while the remote block was still assigned")
	case <-checkpointDone:
		checkpointDone = nil
		t.Fatal("checkpoint returned before the remote block completed")
	default:
	}
	firstOnce.Do(func() { close(releaseFirst) })
	select {
	case frontier := <-queue.begin:
		if frontier != 1 || manager.Size() != 2 || len(trace.Records()) != 2 {
			t.Fatal("checkpoint did not wait for successor publication")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("checkpoint did not reach the suspended frontier")
	}
	select {
	case err := <-checkpointDone:
		checkpointDone = nil
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("checkpoint commit watchdog expired")
	}
	select {
	case <-continued:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not resume after the checkpoint")
	}
	queue.FinishAll()
	secondOnce.Do(func() { close(releaseSecond) })
	<-threadDone
	if server.LastError != nil || calls.Load() != 2 {
		t.Fatalf("worker continuation failed: %v, calls %d", server.LastError, calls.Load())
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	SetNumWorkers(0)
	recoveredQueue := NewDiskStateQueue(metadir)
	t.Cleanup(recoveredQueue.FinishAll)
	recoveredTrace := NewTLCTrace(metadir, "Spec")
	t.Cleanup(func() { _ = recoveredTrace.Close() })
	recoveredSet := NewMemFPSet()
	recoveredSet.Init(1, metadir, "Spec")
	recoveredManager := NewNonDistributedFPSetManager(recoveredSet, "local", recoveredTrace)
	if remote {
		// Reopen owned storage behind a new host/client: recovery must use the
		// committed remote checkpoint file, not the old in-memory table.
		_, endpoint := startFingerprintRPC(t, NewLocalFingerprintEndpoint(recoveredSet))
		recoveredManager = NewDistributedFPSetManager(endpoint)
	}
	if recoveredManager.Size() != 0 {
		t.Fatal("recovery must start with newly opened empty fingerprint storage")
	}
	recovered := NewTLCServer("Spec", "Spec", metadir, recoveredManager, recoveredQueue, recoveredTrace)
	if err := recovered.Recover(); err != nil {
		t.Fatal(err)
	}
	if recoveredQueue.Size() != 1 || recoveredManager.Size() != 2 || !recoveredManager.Contains(61) || !recoveredManager.Contains(71) {
		t.Fatal("recovery lost the published fingerprint/frontier relationship")
	}
	successor := recoveredQueue.Dequeue()
	if successor == nil || successor.UID != trace.Records()[1].State.UID || successor.Level() != 2 {
		t.Fatal("recovered frontier does not identify the committed successor trace")
	}
	enumerator := recoveredTrace.Elements()
	defer enumerator.Close()
	if enumerator.NextPos() != initial.UID || enumerator.NextFP() != 61 || enumerator.NextPos() != successor.UID || enumerator.NextFP() != 71 || enumerator.NextPos() != -1 {
		t.Fatal("recovered disk trace does not match the committed frontier and fingerprints")
	}
}
