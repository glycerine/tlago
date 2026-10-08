package tlc

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type heldWorkerReply struct {
	*LocalWorkerEndpoint
	computed, release, finished chan struct{}
	calls                       atomic.Int64
	result                      *NextStateResult
	err                         error
}

func (e *heldWorkerReply) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	e.calls.Add(1)
	e.result, e.err = e.LocalWorkerEndpoint.GetNextStates(states)
	close(e.computed)
	<-e.release
	defer close(e.finished)
	return e.result, e.err
}

// No original Java method directly covers lost native replies. Source EOF retry
// and worker-loss cleanup must retain assigned work without publishing a result
// computed by an endpoint whose connection has disappeared.
func TestWorkerRPCLostComputedReplyRetainsAssignedWork(t *testing.T) {
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	for _, count := range []int{1, 2} {
		t.Run(fmtInt(count), func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			SetNumWorkers(0)
			storage := NewMemFPSet()
			manager := NewDistributedFPSetManager(NewLocalFingerprintEndpoint(storage))
			worker, generated := managerOwnershipWorker(t, 2, manager)
			endpoint := &heldWorkerReply{LocalWorkerEndpoint: NewLocalWorkerEndpoint(worker),
				computed: make(chan struct{}), release: make(chan struct{}), finished: make(chan struct{})}
			host, client := startWorkerRPC(t, endpoint)
			queue := NewMemStateQueue()
			states := []*TLCStateMut{{UID: 31, level: 4}, {UID: 32, level: 5}}[:count]
			queue.SEnqueueAll(states)
			trace := NewTLCTrace(t.TempDir(), "Spec")
			t.Cleanup(func() { _ = trace.Close() })
			server := &TLCServer{FPSetManager: manager, StateQueue: queue, Trace: trace}
			server.WorkerStatesGenerated.Store(7)
			selector := &BlockSelector{Server: server, Mode: BlockSelectorLimiting, Maximum: count}
			thread := NewTLCServerThread(client, "tcp://worker/primary", server, selector)
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			t.Cleanup(func() { RemoveMessageRecorder(recorder) })
			var release sync.Once
			entered := false
			t.Cleanup(func() {
				_ = host.Close()
				release.Do(func() { close(endpoint.release) })
				thread.Join()
				if entered {
					<-endpoint.finished
				}
				if err := host.CloseGracefully(); err != nil {
					t.Error(err)
				}
			})
			thread.Start()
			select {
			case <-endpoint.computed:
				entered = true
			case <-time.After(5 * time.Second):
				t.Fatal("worker computation watchdog expired")
			}
			if endpoint.err != nil || endpoint.result == nil || endpoint.result.StatesComputed != int64(2*count) || worker.Computing.Load() || *generated != count {
				t.Fatalf("worker did not complete a real result before reply loss: %v/%v", endpoint.result, endpoint.err)
			}
			if queue.Size() != 0 || thread.GetCurrentSize() != count || storage.Size() != 0 || len(trace.records) != 0 {
				t.Fatal("unreceived result changed coordinator publication")
			}
			if err := host.Close(); err != nil {
				t.Fatal(err)
			}
			select {
			case <-thread.runDone:
			case <-time.After(5 * time.Second):
				t.Fatal("coordinator did not retire disconnected worker")
			}
			if NumWorkers() != 0 || thread.GetCurrentSize() != 0 || queue.Size() != int64(count) || len(server.GetServerThreads()) != 0 || !thread.keepAliveStopped.Load() {
				t.Fatal("reply loss did not requeue and deregister exactly once")
			}
			if selector.getMaximum() != 1 || thread.SentStates != 2*count-1 || thread.ReceivedStates != 0 || thread.TimerTask.LastInvocation.Load() != 0 || server.WorkerStatesGenerated.Load() != 7 || server.LastError != nil || server.IsDone() {
				t.Fatal("reply loss changed source retry/statistics/completion behavior")
			}
			if storage.Size() != 0 || len(trace.records) != 0 || endpoint.calls.Load() != 1 || !worker.IsAlive() || worker.Runtime.executor.IsShutdown() {
				t.Fatal("lost reply was replayed, published, or terminated its worker runtime")
			}
			wantRetry := 0
			if count == 2 {
				wantRetry = 1
			}
			if len(recorder.Records(ECTLCDistributedExceedBlocksize)) != wantRetry || len(recorder.Records(ECTLCDistributedWorkerLost)) != 1 || len(recorder.Records(ECTLCDistributedWorkerDeregistered)) != 1 || len(recorder.Records(ECGeneral)) != 1 {
				t.Fatal("source retry/loss/cache-warning event counts changed")
			}
			warning := recorder.Records(ECGeneral)[0]
			if warning.Severity != SeverityWarning || len(warning.Params) != 1 || warning.Params[0] != "Failed to read remote worker cache statistic (Expect to see a negative chache hit rate. Does not invalidate model checking results)" {
				t.Fatalf("unexpected GENERAL diagnostic: %+v", warning)
			}
			thread.HandleRemoteWorkerLost(queue)
			thread.TimerTask.Run()
			if queue.Size() != int64(count) || NumWorkers() != 0 || len(recorder.Records(ECTLCDistributedWorkerDeregistered)) != 1 {
				t.Fatal("repeated loss reports duplicated queued work or deregistration")
			}
			// The smaller-block retry rotates the head state to the tail.
			for index := range count {
				want := states[index]
				if count == 2 {
					want = states[1-index]
				}
				if got := queue.SDequeue(); got != want {
					t.Fatalf("pending source queue order: got %p, want %p", got, want)
				}
			}
			release.Do(func() { close(endpoint.release) })
			<-endpoint.finished
			if err := host.CloseGracefully(); err != nil {
				t.Fatal(err)
			}
			if storage.Size() != 0 || len(trace.records) != 0 || server.WorkerStatesGenerated.Load() != 7 {
				t.Fatal("late reply published after worker retirement")
			}
		})
	}
}
