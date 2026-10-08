package tlc

import (
	"os"
	"strings"
	"testing"
	"time"
)

// No original method covers uncaught coordinator timer failures. Run the real
// timer loop with its original ten-second initial delay, without -race.
func TestDistributedCoordinatorTimerFailureStopsOnlyTimer(t *testing.T) {
	output, err := os.CreateTemp(t.TempDir(), "timer-stderr-")
	if err != nil {
		t.Fatal(err)
	}
	previousStderr := os.Stderr
	os.Stderr = output
	t.Cleanup(func() { os.Stderr = previousStderr; _ = output.Close() })
	var threads []*TLCServerThread
	var completions []chan struct{}
	for _, failure := range []struct {
		err    error
		panics bool
	}{
		{NewRuntimeException("coordinator timer runtime failure"), false},
		{NewAssertionError("coordinator timer fatal failure"), true},
	} {
		queue := NewMemStateQueue()
		thread := &TLCServerThread{
			Worker:        NewDistributedWorkerSmartProxy(&fatalRPCWorker{rpcTestWorker: &rpcTestWorker{}, failure: failure.err, panics: failure.panics}),
			Server:        &TLCServer{StateQueue: queue},
			keepAliveDone: make(chan struct{}),
		}
		thread.TimerTask = &TLCTimerTask{Thread: thread}
		thread.cleanupGlobals.Store(true)
		thread.setStates([]*TLCStateMut{{UID: 7}})
		threads = append(threads, thread)
		done := make(chan struct{})
		completions = append(completions, done)
		t.Cleanup(func() { thread.cancelKeepAlive(); <-done })
		go func() { defer close(done); thread.runKeepAlive() }()
	}
	t.Log("Waiting for the original ten-second coordinator keepalive delay")
	watchdog := time.NewTimer(20 * time.Second)
	defer watchdog.Stop()
	for i, done := range completions {
		select {
		case <-done:
		case <-watchdog.C:
			t.Fatal("coordinator timer did not terminate after uncaught failure")
		}
		thread := threads[i]
		if thread.Server.Done.Load() || !thread.cleanupGlobals.Load() || thread.GetCurrentSize() != 1 || thread.Server.StateQueue.Size() != 0 || thread.keepAliveStopped.Load() {
			t.Fatal("uncaught timer failure changed coordinator or worker-loss state")
		}
	}
	os.Stderr = previousStderr
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range []string{"coordinator timer runtime failure", "coordinator timer fatal failure"} {
		if !strings.Contains(string(data), message) {
			t.Fatalf("uncaught timer diagnostic omitted %q: %s", message, data)
		}
	}
}
