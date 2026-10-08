package tlc

import (
	"strconv"
	"strings"
	"testing"
)

// Upstream has no direct test of worker resource failures. The coordinator's
// smaller-block decision must survive both local and native TCP invocation.
func TestDistributedWorkerComputationFailure(t *testing.T) {
	for _, transport := range []string{"local", "tcp"} {
		for _, family := range []string{"memory", "rejected"} {
			for _, panics := range []bool{false, true} {
				t.Run(transport+"/"+family+"/"+strconv.FormatBool(panics), func(t *testing.T) {
					var cause error = NewOutOfMemoryError("worker memory exhausted")
					prefix := "OutOfMemoryError occurred at worker: "
					if family == "rejected" {
						cause = NewRejectedExecutionException(javaString("worker stopped accepting tasks"), nil)
						prefix = "Executor rejected task at worker: "
					}
					tool := &Tool{Actions: []*Action{{}}, GetNextStatesFunc: func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
						if panics {
							panic(cause)
						}
						return nil, cause
					}}
					worker := NewDistributedWorker(7, tool, nil, DistributedWorkerAddress{Hostname: "worker", Port: 1234})
					t.Cleanup(worker.Runtime.executor.Shutdown)
					worker.OverallStatesComputed.Store(17)
					var endpoint DistributedWorkerEndpoint = NewLocalWorkerEndpoint(worker)
					if transport == "tcp" {
						_, endpoint = startWorkerRPC(t, endpoint)
					}
					result, err := endpoint.GetNextStates([]*TLCStateMut{{UID: 31}})
					failure, ok := err.(*DistributedOperationError)
					if result != nil || !ok {
						t.Fatalf("worker failure must use Go operation categories: %v, %T/%v", result, err, err)
					}
					if !failure.Remote || !failure.IO || failure.Recoverable != (family == "memory") || !failure.ExitIgnorable || failure.WorkerUnavailable {
						t.Fatalf("worker failure changed coordinator decisions: %+v", failure)
					}
					if failure.Message == nil || *failure.Message != prefix+worker.GetASCIIURI() || strings.Contains(failure.Class, "java.rmi") {
						t.Fatalf("worker diagnostic lost context or impersonates RMI: %+v", failure)
					}
					if transport == "local" && failure.Cause != cause {
						t.Fatal("local failure replaced its original cause")
					}
					if failure.Cause == nil || failure.Cause.Error() != cause.Error() {
						t.Fatalf("failure lost cause: %v", failure.Cause)
					}
					if !strings.Contains(javaThrowableStackTrace(failure), "distributed.go") {
						t.Fatal("worker failure lost its originating Go computation stack")
					}
					if worker.Computing.Load() || worker.LastInvocation.Load() == 0 || worker.OverallStatesComputed.Load() != 17 {
						t.Fatal("resource failure changed computation lifecycle or counters")
					}
					if alive, err := endpoint.IsAlive(); err != nil || !alive {
						t.Fatalf("resource failure removed worker: %v/%v", alive, err)
					}
					if family == "memory" {
						captureFailoverToolIO(t, ToolIOTool)
						selector := NewLimitingBlockSelector(&TLCServer{}, 100)
						thread := &TLCServerThread{Selector: selector, Worker: NewDistributedWorkerSmartProxy(endpoint)}
						states := []*TLCStateMut{{UID: 31}, {UID: 32}}
						thread.setStates(states)
						queue := NewMemStateQueue()
						result, proceed := thread.computeBlock(queue)
						if result != nil || !proceed || queue.Size() != 2 || queue.SDequeue() != states[0] || queue.SDequeue() != states[1] || selector.getMaximum() != 1 {
							t.Fatal("native worker failure lost smaller-batch retry or requeued states")
						}
					}
				})
			}
		}
	}
}
