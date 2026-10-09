package tlc

import (
	"testing"
	"time"
)

// Native scheduler coverage: Java has no enabled direct test for TLCTimerTask.
// Keep its actual ten-second initial delay; do not replace the timer with a
// manual run invocation or shorten the production activity timeout.
func TestNativeWorkerTimerDetectsCachedCoordinatorLoss(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() { RemoveMessageRecorder(recorder) })
	host, coordinator := startCoordinatorRPC(t, NewLocalServerEndpoint(&TLCServer{}))
	network, err := NewDistributedWorkerNetwork("127.0.0.1:0", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := network.Close(); err != nil {
			t.Error(err)
		}
	})
	location := "tcp://" + coordinator.Address + "/main"
	// Establish and use the cached discovery connection before losing the host.
	view, err := network.Discovery.Lookup(location)
	if err != nil {
		t.Fatal(err)
	}
	if done, err := view.IsDone(); err != nil || done {
		t.Fatalf("live coordinator status = %v/%v", done, err)
	}
	workers := []*DistributedWorker{
		NewDistributedWorker(0, nil, NewDistributedFPSetManager(), network.workerAddress),
		NewDistributedWorker(1, nil, NewDistributedFPSetManager(), network.workerAddress),
	}
	runtime := NewDistributedWorkerRuntime(workers...)
	callbacks := make([]*NetworkWorkerEndpoint, len(workers))
	env := network.Environment(DistributedWorkerEnvironment{})
	for i, worker := range workers {
		if err := env.PublishWorker(worker); err != nil {
			t.Fatal(err)
		}
		callbacks[i], err = DialWorkerEndpoint(network.Address, worker.networkReference.Object)
		if err != nil {
			t.Fatal(err)
		}
		callback := callbacks[i]
		t.Cleanup(func() { _ = callback.CloseConnection() })
		if alive, err := callback.IsAlive(); err != nil || !alive {
			t.Fatalf("published worker %d = %v/%v", i, alive, err)
		}
	}
	if err := host.Close(); err != nil {
		t.Fatal(err)
	}
	lookupStarted := make(chan time.Time, 1)
	lookup := NewTLCServerStatusLookup(network.Discovery.Lookup)
	task := &distributedWorkerKeepAlive{
		workers: workers, serverURL: location, done: make(chan struct{}),
		timeout: int64(DistributedWorkerKeepAliveTimeoutMillis()),
		lookup: func(url string) (bool, error) {
			lookupStarted <- time.Now()
			return lookup(url)
		},
	}
	// Run the same scheduler used by StartKeepAlive in a joinable fixture
	// goroutine. Worker.exit must cancel this published task itself.
	runtime.keepAlive = task
	joined := make(chan struct{})
	started := time.Now()
	go func() { defer close(joined); task.runTimer() }()
	t.Cleanup(func() { task.cancel(); <-joined })
	t.Log("coordinator closed; waiting for the unchanged ten-second worker timer")
	select {
	case invoked := <-lookupStarted:
		if elapsed := invoked.Sub(started); elapsed < 10*time.Second {
			t.Fatalf("scheduler probed coordinator early: %s", elapsed)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("production scheduler did not probe the lost coordinator")
	}
	select {
	case <-joined:
	case <-time.After(5 * time.Second):
		t.Fatal("coordinator loss did not terminate the timer")
	}
	select {
	case <-runtime.latch.Load().done:
	default:
		t.Fatal("coordinator loss did not release all worker exits")
	}
	if !runtime.executor.IsShutdown() {
		t.Fatal("coordinator loss retained the shared worker executor")
	}
	for i, callback := range callbacks {
		if !workers[i].unexported.Load() {
			t.Fatalf("worker %d was not removed", i)
		}
		if _, err := callback.IsAlive(); !isDistributedWorkerEndpointRemoved(err) {
			t.Fatalf("worker %d callback survived timer exit: %v", i, err)
		}
	}
	if records := recorder.Records(ECTLCDistributedServerNotRunning); len(records) != 1 {
		t.Fatalf("lost-coordinator diagnostic count = %d, want 1", len(records))
	}
	if len(recorder.Records(ECTLCDistributedServerFinished)) != 0 || len(recorder.Records(ECGeneral)) != 0 {
		t.Fatal("coordinator loss was reported as success or a general failure")
	}
	t.Log("timer joined; both worker callbacks removed and shared completion latch released")
}
