package tlc

import "testing"

// There is no direct upstream endpoint-removal test. Preserve cleanup order and
// completion ownership through native Go endpoint and error boundaries.
func TestNativeWorkerEndpointRemoval(t *testing.T) {
	for _, transport := range []string{"local", "tcp"} {
		t.Run(transport, func(t *testing.T) { checkNativeWorkerEndpointRemoval(t, transport) })
	}
}

func checkNativeWorkerEndpointRemoval(t *testing.T, transport string) {
	t.Helper()
	captureFailoverToolIO(t, ToolIOTool)
	runtime, worker, _ := shutdownTraversalRuntime(t)
	var endpoint DistributedWorkerEndpoint = NewLocalWorkerEndpoint(worker)
	if transport == "tcp" {
		_, endpoint = startWorkerRPC(t, endpoint)
	}
	if err := endpoint.Exit(); err != nil {
		t.Fatal(err)
	}
	if runtime.latch.Load().remaining != 1 || !runtime.executor.IsShutdown() {
		t.Fatal("first exit did not retain completion/executor ordering")
	}
	operations := map[string]func() error{
		"next":               func() error { _, err := endpoint.GetNextStates(nil); return err },
		"alive":              func() error { _, err := endpoint.IsAlive(); return err },
		"uri":                func() error { _, err := endpoint.GetURI(); return err },
		"cache":              func() error { _, err := endpoint.GetCacheRateRatio(); return err },
		"exit":               endpoint.Exit,
		"direct-repeat-exit": worker.Exit,
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			err := operation()
			failure, ok := err.(*DistributedOperationError)
			if !ok || !failure.Remote || !failure.IO || failure.Recoverable || !failure.WorkerUnavailable || !failure.ExitIgnorable {
				t.Fatalf("removed worker must use native unavailable category: %T/%v", err, err)
			}
			if failure.Class != "tlc.WorkerEndpointRemoved" || !isDistributedWorkerEndpointRemoved(failure) || runtime.latch.Load().remaining != 1 {
				t.Fatal("removed endpoint impersonates a Java transport or decrements completion again")
			}
			payload, err := EncodeDistributedFailure(failure)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeDistributedFailure(payload)
			if err != nil || !isDistributedWorkerUnavailable(decoded) || !isIgnorableDistributedWorkerExit(decoded) || isRecoverableDistributedError(decoded) || !isDistributedWorkerEndpointRemoved(decoded) {
				t.Fatalf("removal category lost across native payload: %v/%v", decoded, err)
			}
		})
	}
}

func TestNativeWorkerCleanupContinuesAfterRemoval(t *testing.T) {
	for _, phase := range []string{"shutdown", "keepalive"} {
		t.Run(phase, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			runtime, first, second := shutdownTraversalRuntime(t)
			runtime.runnables[1].worker.Store(second)
			if err := first.Exit(); err != nil {
				t.Fatal(err)
			}
			logged := 0
			runtime.keepAlive.workers = []*DistributedWorker{first, second}
			runtime.keepAlive.logFinest = func(message string, failure error) {
				logged++
				if _, ok := failure.(*DistributedOperationError); !ok || message != "Failed to exit worker" {
					t.Fatalf("removal diagnostic changed: %q, %T/%v", message, failure, failure)
				}
			}
			var err error
			if phase == "shutdown" {
				err = runtime.Shutdown()
			} else {
				err = runtime.keepAlive.exitWorker(nil, 2)
			}
			if err != nil || runtime.latch.Load().remaining != 0 || !second.unexported.Load() {
				t.Fatalf("earlier removal suppressed later exit: %v", err)
			}
			if phase == "keepalive" && logged != 1 || phase == "shutdown" && logged != 0 {
				t.Fatal("removed endpoint changed the cleanup logging boundary")
			}
			select {
			case <-runtime.keepAlive.done:
			default:
				t.Fatal("cleanup did not cancel its timer")
			}
		})
	}
}
