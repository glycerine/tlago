package tlc

import (
	"math"
	"strings"
	"testing"
)

// No direct upstream test covers missing error-handler owners. Source still
// attempts trace printing and queue shutdown after recording the model error.
func TestDistributedThreadMissingTraceReportsPrintingFailure(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	queue := NewMemStateQueue()
	server := &TLCServer{StateQueue: queue}
	waiter := make(chan struct{})
	server.completionWaiter = waiter
	predecessor := &TLCStateMut{UID: 7, level: 2}
	original := NewWorkerException("original worker failure", predecessor, nil, true)
	thread := &TLCServerThread{Server: server}
	err := invokeDistributedServerOperation(func() error { thread.handleRunError(original, queue); return nil })
	if err != nil || !server.IsDone() || server.ErrState != predecessor || server.LastError != original || !queue.finish.Load() {
		t.Fatalf("error-handler state changed: %v", err)
	}
	records := recorder.Records(ECGeneral)
	if len(records) != 1 || len(records[0].Params) != 1 || !strings.Contains(records[0].Params[0], "java.lang.NullPointerException") || strings.Contains(records[0].Params[0], "original worker failure") {
		t.Fatalf("trace failure diagnostic %v", records)
	}
	select {
	case <-waiter:
	default:
		t.Fatal("caught trace-printing failure skipped queue completion notification")
	}
}

func TestDistributedThreadMissingQueueSkipsNotification(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	server := &TLCServer{}
	waiter := make(chan struct{})
	server.completionWaiter = waiter
	original := NewRuntimeException("original computation failure")
	thread := &TLCServerThread{Server: server}
	err := invokeDistributedServerOperation(func() error { thread.handleRunError(original, nil); return nil })
	if !isDistributedNullFailure(err) || !server.IsDone() || server.LastError != original || !server.KeepCallStack {
		t.Fatalf("missing queue lost failure or prior model mutation: %T/%v", err, err)
	}
	select {
	case <-waiter:
		t.Fatal("queue failure continued into completion notification")
	default:
	}
	if server.completionWaiter != waiter {
		t.Fatal("failed handler cleared pending notification")
	}
	messages := ToolIOGetAllMessages()
	if len(messages) != 1 || !strings.Contains(messages[0], "original computation failure") {
		t.Fatalf("missing queue changed prior error reporting: %q", messages)
	}
}

func TestDistributedThreadMissingQueueRunsFinally(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	server := &TLCServer{}
	waiter := make(chan struct{})
	server.completionWaiter = waiter
	thread := &TLCServerThread{Server: server,
		Selector:          NewStaticBlockSelector(server, 1),
		Worker:            NewDistributedWorkerSmartProxy(&rpcTestWorker{}),
		CacheRateHitRatio: -1, keepAliveDone: make(chan struct{}),
	}
	thread.setStates([]*TLCStateMut{{UID: 7}})
	err := invokeDistributedServerOperation(func() error { thread.Run(); return nil })
	if !isDistributedNullFailure(err) || !isDistributedNullFailure(server.LastError) || !server.IsDone() || !server.KeepCallStack {
		t.Fatalf("missing queue was treated as normal completion: %T/%v, recorded %v", err, err, server.LastError)
	}
	if !math.IsNaN(thread.CacheRateHitRatio) || !thread.keepAliveStopped.Load() || thread.currentStates() == nil || len(thread.currentStates()) != 0 {
		t.Fatal("handler failure skipped thread finally operations")
	}
	select {
	case <-waiter:
		t.Fatal("missing queue sent completion notification")
	default:
	}
	if NumWorkers() != oldWorkers+1 {
		t.Fatal("error-handler failure entered worker-loss decrement")
	}
}
