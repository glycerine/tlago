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
	for _, scenario := range []struct {
		name        string
		remoteCount int
	}{{"local_fingerprints", 0}, {"tcp_fingerprints", 1}, {"partitioned_tcp_fingerprints", 2}} {
		t.Run(scenario.name, func(t *testing.T) { checkDistributedCheckpointWithAssignedBlock(t, scenario.remoteCount) })
	}
}

func checkDistributedCheckpointWithAssignedBlock(t *testing.T, remoteCount int) {
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
	storeDirectories := []string{metadir}
	if remoteCount == 2 {
		// Real FP servers own separate metadata directories. Their common
		// checkpoint filename must not overwrite another partition's file.
		storeDirectories = []string{t.TempDir(), t.TempDir()}
	}
	openStores := func(recoveryTrace *TLCTrace) (*DistributedFPSetManager, []*MemFPSet) {
		var stores []*MemFPSet
		var endpoints []DistributedFingerprintEndpoint
		for _, directory := range storeDirectories {
			set := NewMemFPSet()
			set.Init(1, directory, "Spec")
			stores = append(stores, set)
			if remoteCount > 0 {
				_, endpoint := startFingerprintRPC(t, NewLocalFingerprintEndpoint(set))
				endpoints = append(endpoints, endpoint)
			}
		}
		if remoteCount == 0 {
			return NewNonDistributedFPSetManager(stores[0], "local", recoveryTrace), stores
		}
		return NewDistributedFPSetManager(endpoints...), stores
	}
	manager, stores := openStores(trace)
	successorFP := int64(71)
	if remoteCount == 2 {
		successorFP = 72 // Initial 61 is in partition 1, successor 72 in 0.
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
			statePartitions := []*StateVec{NewStateVecFrom([]*TLCStateMut{successor})}
			fpPartitions := []*LongVec{NewLongVecFrom([]int64{successorFP})}
			if remoteCount == 2 {
				statePartitions = append(statePartitions, NewStateVec(0))
				fpPartitions = append(fpPartitions, NewLongVec())
			}
			return NewNextStateResult(statePartitions, fpPartitions, 0, 1), nil
		}
		continued <- struct{}{}
		<-releaseSecond
		statePartitions, fpPartitions := []*StateVec{NewStateVec(0)}, []*LongVec{NewLongVec()}
		if remoteCount == 2 {
			statePartitions = append(statePartitions, NewStateVec(0))
			fpPartitions = append(fpPartitions, NewLongVec())
		}
		return NewNextStateResult(statePartitions, fpPartitions, 0, 0), nil
	}}
	_, endpoint := startWorkerRPC(t, worker)
	thread := &TLCServerThread{Server: server, Worker: NewDistributedWorkerSmartProxy(endpoint), Selector: NewStaticBlockSelector(server, 1)}
	thread.TimerTask = &TLCTimerTask{Thread: thread}
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
		if frontier != 1 || manager.Size() != 2 || trace.GetLevelForReporting() != 2 {
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
	// Open fresh tables behind fresh hosts; Recover must read committed files.
	recoveredManager, recoveredStores := openStores(recoveredTrace)
	if recoveredManager.Size() != 0 {
		t.Fatal("recovery must start with newly opened empty fingerprint storage")
	}
	recovered := NewTLCServer("Spec", "Spec", metadir, recoveredManager, recoveredQueue, recoveredTrace)
	if err := recovered.Recover(); err != nil {
		t.Fatal(err)
	}
	if recoveredQueue.Size() != 1 || recoveredManager.Size() != 2 || !recoveredManager.Contains(61) || !recoveredManager.Contains(uint64(successorFP)) {
		t.Fatal("recovery lost the published fingerprint/frontier relationship")
	}
	if remoteCount == 2 {
		for _, partitions := range [][]*MemFPSet{stores, recoveredStores} {
			if partitions[0].Size() != 1 || !partitions[0].Contains(uint64(successorFP)) || partitions[0].Contains(61) || partitions[1].Size() != 1 || !partitions[1].Contains(61) || partitions[1].Contains(uint64(successorFP)) {
				t.Fatal("checkpoint/recovery merged, swapped or lost fingerprint partitions")
			}
		}
	}
	successor := recoveredQueue.Dequeue()
	if successor == nil || successor.UID != trace.lastPtr || successor.Level() != 2 {
		t.Fatal("recovered frontier does not identify the committed successor trace")
	}
	enumerator, err := recoveredTrace.Elements()
	if err != nil {
		t.Fatal(err)
	}
	defer enumerator.Close()
	for _, want := range []struct {
		pos int64
		fp  uint64
	}{{initial.UID, 61}, {successor.UID, uint64(successorFP)}} {
		pos, err := enumerator.NextPos()
		if err != nil || pos != want.pos {
			t.Fatalf("trace position = %d/%v, want %d", pos, err, want.pos)
		}
		fp, err := enumerator.NextFP()
		if err != nil || fp != want.fp {
			t.Fatalf("trace fingerprint = %d/%v, want %d", fp, err, want.fp)
		}
	}
	if pos, err := enumerator.NextPos(); err != nil || pos != -1 {
		t.Fatalf("trace end = %d/%v", pos, err)
	}
}
