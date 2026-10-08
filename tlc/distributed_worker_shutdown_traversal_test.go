package tlc

import (
	"bytes"
	"sync"
	"testing"
)

type shutdownPublicationWriter struct {
	bytes.Buffer
	publish func()
}

func (w *shutdownPublicationWriter) Write(data []byte) (int, error) {
	if w.publish != nil {
		publish := w.publish
		w.publish = nil
		publish()
	}
	return w.Buffer.Write(data)
}

func shutdownTraversalRuntime(t *testing.T) (*DistributedWorkerRuntime, *DistributedWorker, *DistributedWorker) {
	t.Helper()
	first := NewDistributedWorker(0, nil, nil, DistributedWorkerAddress{Hostname: "first"})
	second := NewDistributedWorker(1, nil, nil, DistributedWorkerAddress{Hostname: "second"})
	runtime := NewDistributedWorkerRuntime(first, second)
	runtime.keepAlive = &distributedWorkerKeepAlive{done: make(chan struct{})}
	runtime.runnables = []*DistributedWorkerRunnable{{}, {}}
	runtime.runnables[0].worker.Store(first)
	return runtime, first, second
}

// No original method tests startup publication during shutdown. Source shutdown
// reads each runnable's worker after exiting the preceding worker.
func TestDistributedWorkerShutdownObservesLaterPublication(t *testing.T) {
	for _, concurrent := range []bool{false, true} {
		t.Run(map[bool]string{false: "callback", true: "startup-goroutine"}[concurrent], func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOSystem)
			runtime, first, second := shutdownTraversalRuntime(t)
			publish := func() { runtime.runnables[1].worker.Store(second) }
			if concurrent {
				runnable := runtime.runnables[1]
				start, done := make(chan struct{}), make(chan struct{})
				var release sync.Once
				go func() {
					<-start
					runnable.worker.Store(second)
					close(done)
				}()
				publish = func() { release.Do(func() { close(start) }); <-done }
				t.Cleanup(publish)
			}
			output := &shutdownPublicationWriter{publish: publish}
			defer ToolIOSetSystemStreams(output, output)()
			if err := runtime.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if !first.unexported.Load() || !second.unexported.Load() || runtime.latch.Load().remaining != 0 {
				t.Fatal("shutdown skipped a worker published during an earlier exit")
			}
			if len(runtime.runnables) != 0 || len(runtime.workers) != 0 || !runtime.executor.IsShutdown() {
				t.Fatal("completed shutdown did not retain executor/array lifecycle")
			}
			select {
			case <-runtime.keepAlive.done:
			default:
				t.Fatal("shutdown did not cancel timer")
			}
		})
	}
}

func TestDistributedWorkerShutdownSkipsUnpublishedWorker(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	runtime, first, second := shutdownTraversalRuntime(t)
	if err := runtime.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if !first.unexported.Load() || second.unexported.Load() || runtime.latch.Load().remaining != 1 || len(runtime.runnables) != 0 {
		t.Fatal("shutdown changed source null-worker skip or completion latch")
	}
}

func TestDistributedWorkerShutdownNullRunnableRetainsEarlierExit(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	runtime, first, second := shutdownTraversalRuntime(t)
	runtime.runnables[1] = nil
	defer func() {
		failure := recover()
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("null runnable failure %T/%v", failure, failure)
		}
		if !runtime.keepAliveMu.TryLock() {
			t.Fatal("null runnable failure retained the runtime lifecycle lock")
		}
		runtime.keepAliveMu.Unlock()
		if !first.unexported.Load() || second.unexported.Load() || runtime.latch.Load().remaining != 1 || len(runtime.runnables) != 2 || len(runtime.workers) != 2 {
			t.Fatal("null runnable changed prior exits or cleared runtime arrays")
		}
	}()
	_ = runtime.Shutdown()
}
