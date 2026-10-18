package tlc

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

type finalizerCacheWorker struct {
	*rpcTestWorker
	failure error
	panics  bool
}

func (w *finalizerCacheWorker) GetCacheRateRatio() (float64, error) {
	if w.panics {
		panic(w.failure)
	}
	return 0, w.failure
}

// No original method covers final cache-read failure or uncaught server-thread
// failure. Native return and panic forms must preserve the same catch boundary.
func TestDistributedThreadFinalCacheRemoteFailure(t *testing.T) {
	for _, panics := range []bool{false, true} {
		t.Run(map[bool]string{false: "returned", true: "panicked"}[panics], func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			thread := &TLCServerThread{CacheRateHitRatio: -1,
				Worker: NewDistributedWorkerSmartProxy(&finalizerCacheWorker{
					rpcTestWorker: &rpcTestWorker{}, failure: distributedTestRemoteFailure("cache unavailable"), panics: panics,
				})}
			err := invokeDistributedServerOperation(func() error { thread.readCacheRateRatio(); return nil })
			if err != nil || thread.CacheRateHitRatio != -1 {
				t.Fatalf("remote cache catch = %v, ratio %v", err, thread.CacheRateHitRatio)
			}
			if records := recorder.Records(ECGeneral); len(records) != 1 || records[0].Severity != SeverityWarning {
				t.Fatal("remote cache failure did not print exactly one warning")
			}
		})
	}
}

func TestDistributedThreadFinalizerFailureProcess(t *testing.T) {
	if family := os.Getenv("TLAGO_THREAD_FINALIZER_FAILURE"); family != "" {
		checkDistributedThreadFinalizerFailure(t, family)
		return
	}
	for _, family := range []string{"runtime_return", "runtime_panic", "fatal_return", "fatal_panic", "handler_fatal"} {
		t.Run(family, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestDistributedThreadFinalizerFailureProcess$")
			command.Env = append(os.Environ(), "TLAGO_THREAD_FINALIZER_FAILURE="+family)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("native finalizer process failed: %v\n%s", err, output)
			}
		})
	}
}

func checkDistributedThreadFinalizerFailure(t *testing.T, family string) {
	captureFailoverToolIO(t, ToolIOTool)
	SetNumWorkers(0)
	output, err := os.CreateTemp(t.TempDir(), "finalizer-stderr")
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = output
	defer func() { os.Stderr = oldStderr; _ = output.Close() }()
	queue := NewMemStateQueue()
	server := &TLCServer{StateQueue: queue}
	var failure error = NewRuntimeException("final cache failure")
	if strings.HasPrefix(family, "fatal") {
		failure = NewAssertionError("final cache failure")
	}
	worker := &finalizerCacheWorker{rpcTestWorker: &rpcTestWorker{}, failure: failure, panics: strings.HasSuffix(family, "panic")}
	if family == "handler_fatal" {
		state := &TLCStateMut{UID: 7, level: 2}
		queue.Enqueue(state)
		worker.failure = distributedTestRemoteFailure("final cache unavailable")
		worker.next = func([]*TLCStateMut) (*NextStateResult, error) {
			return nil, NewWorkerExceptionWithCause("worker failure", nil, state, nil, true)
		}
		server.Trace = NewTLCTrace()
		server.Trace.Tool = &Tool{GetStateFunc: func(*Tool, uint64, ...any) (*TLCStateInfo, error) {
			panic(NewAssertionError("error handler failure"))
		}}
	} else {
		queue.FinishAll()
	}
	thread := &TLCServerThread{Server: server, Worker: NewDistributedWorkerSmartProxy(worker),
		Selector: NewBlockSelectorFromProperties(server), CacheRateHitRatio: -1,
		runDone: make(chan struct{}), keepAliveDone: make(chan struct{})}
	thread.TimerTask = &TLCTimerTask{Thread: thread}
	thread.cleanupGlobals.Store(true)
	thread.Start()
	thread.Join()
	defer thread.cancelKeepAlive()
	if !server.IsDone() || NumWorkers() != 1 || thread.CacheRateHitRatio != -1 || !thread.cleanupGlobals.Load() {
		t.Fatal("uncaught thread failure changed model result, worker count or lost-worker cleanup")
	}
	wantMessage := "final cache failure"
	if family == "handler_fatal" {
		wantMessage = "error handler failure"
		if queue.finish.Load() {
			t.Fatal("fatal trace printing continued into the catch's queue finish")
		}
		if !thread.keepAliveStopped.Load() || thread.currentStates() == nil || len(thread.currentStates()) != 0 {
			t.Fatal("error handler failure skipped the source finally operations")
		}
		if len(ToolIOGetAllMessages()) < 2 {
			t.Fatal("error handler failure skipped the remote cache warning")
		}
	} else if thread.keepAliveStopped.Load() || thread.currentStates() != nil {
		t.Fatal("uncaught final cache failure continued into later finally operations")
	}
	data, err := os.ReadFile(output.Name())
	if err != nil || !strings.Contains(string(data), wantMessage) {
		t.Fatalf("uncaught thread diagnostic missing: %q/%v", data, err)
	}
}
