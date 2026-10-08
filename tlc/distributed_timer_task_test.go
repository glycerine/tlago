package tlc

import (
	"testing"
	"time"
)

type timerStatusWorker struct {
	*rpcTestWorker
	failure error
	panics  bool
	alive   bool
	calls   int
}

func (w *timerStatusWorker) IsAlive() (bool, error) {
	w.calls++
	if w.panics {
		panic(w.failure)
	}
	return w.alive, w.failure
}

// No original method directly covers coordinator timer-task failure categories.
// The task body is synchronous; no timer delay or long workload is selected.
func TestDistributedTimerTaskRemoteFailureForms(t *testing.T) {
	for _, phase := range []string{"returned_remote", "panicked_remote", "not_alive"} {
		t.Run(phase, func(t *testing.T) {
			worker := &timerStatusWorker{rpcTestWorker: &rpcTestWorker{}, panics: phase == "panicked_remote"}
			if phase != "not_alive" {
				worker.failure = NewRemoteException(javaString("worker unavailable"), nil)
			}
			checkTimerTaskWorkerLoss(t, worker)
			if worker.calls != 2 {
				t.Fatal("repeated timer invocation did not repeat its liveness call")
			}
		})
	}
}

func checkTimerTaskWorkerLoss(t *testing.T, worker DistributedWorkerEndpoint) {
	t.Helper()
	captureFailoverToolIO(t, ToolIOTool)
	oldWorkers := NumWorkers()
	SetNumWorkers(2)
	defer SetNumWorkers(oldWorkers)
	queue := NewMemStateQueue()
	server := &TLCServer{StateQueue: queue}
	thread := &TLCServerThread{Server: server, Worker: NewDistributedWorkerSmartProxy(worker),
		URI: "tcp://worker:1234/primary", keepAliveDone: make(chan struct{})}
	thread.cleanupGlobals.Store(true)
	states := []*TLCStateMut{{UID: 7}, {UID: 11}}
	thread.setStates(states)
	server.RegisterTLCServerThread(thread)
	task := &TLCTimerTask{Thread: thread}
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	for range 2 {
		if err := invokeDistributedServerOperation(func() error { task.Run(); return nil }); err != nil {
			t.Fatalf("remote timer failure escaped the worker-loss catch: %T/%v", err, err)
		}
	}
	if server.IsDone() || server.GetWorkerCount() != 0 || NumWorkers() != 1 || thread.cleanupGlobals.Load() || !thread.keepAliveStopped.Load() || thread.GetCurrentSize() != 0 || queue.Size() != 2 {
		t.Fatal("timer failure changed worker-loss cleanup, queueing or completion")
	}
	if queue.SDequeue() != states[0] || queue.SDequeue() != states[1] {
		t.Fatal("timer failure lost, duplicated or reordered assigned work")
	}
	if len(recorder.Records(ECTLCDistributedWorkerDeregistered)) != 1 || recorder.Recorded(ECTLCDistributedWorkerLost) || recorder.Recorded(ECGeneral) {
		t.Fatal("timer task changed the source deregistration-only reporting")
	}
}

func TestDistributedTimerTaskUncheckedFailureAndActivity(t *testing.T) {
	for _, failure := range []error{NewRuntimeException("status runtime"), NewAssertionError("status fatal")} {
		for _, panics := range []bool{false, true} {
			worker := &timerStatusWorker{rpcTestWorker: &rpcTestWorker{}, failure: failure, panics: panics}
			queue := NewMemStateQueue()
			thread := &TLCServerThread{Server: &TLCServer{StateQueue: queue}, Worker: NewDistributedWorkerSmartProxy(worker), keepAliveDone: make(chan struct{})}
			thread.cleanupGlobals.Store(true)
			thread.setStates([]*TLCStateMut{{UID: 7}})
			task := &TLCTimerTask{Thread: thread}
			for _, when := range []time.Time{time.Now(), time.Now().Add(time.Minute)} {
				task.SetLastInvocation(when)
				task.Run()
			}
			if worker.calls != 0 {
				t.Fatal("recent or future activity did not suppress the status call")
			}
			task.SetLastInvocation(time.Now().Add(-2 * time.Minute))
			err := invokeDistributedServerOperation(func() error { task.Run(); return nil })
			if err != failure || worker.calls != 1 || thread.Server.IsDone() || queue.Size() != 0 || !thread.cleanupGlobals.Load() || thread.keepAliveStopped.Load() || thread.GetCurrentSize() != 1 {
				t.Fatal("unchecked timer failure entered the remote catch or changed ownership")
			}
		}
	}
}

func TestDistributedTimerTaskFailuresOverTCP(t *testing.T) {
	for _, panics := range []bool{false, true} {
		_, client := startWorkerRPC(t, &timerStatusWorker{rpcTestWorker: &rpcTestWorker{}, failure: NewAssertionError("remote status fatal"), panics: panics})
		checkTimerTaskWorkerLoss(t, client)
		if _, err := client.GetURI(); err != nil {
			t.Fatal("status failure stopped the worker host", err)
		}
	}
}
