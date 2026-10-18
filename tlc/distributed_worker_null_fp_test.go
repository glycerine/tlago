package tlc

import (
	"sync/atomic"
	"testing"
)

type workerNullFingerprintEndpoint struct {
	*LocalFingerprintEndpoint
	calls atomic.Int32
}

func (e *workerNullFingerprintEndpoint) ContainsBlock(*LongVec) (*BitVector, error) {
	e.calls.Add(1)
	return nil, nil
}

// The original worker wraps a null containsBlock answer in WorkerException,
// with the current predecessor and KeepCallStack. It cannot return success
// with an empty successor list. No enabled direct Java test covers this path.
func nullFingerprintAnswerWorker(t *testing.T) (*DistributedWorker, *workerNullFingerprintEndpoint) {
	t.Helper()
	endpoint := &workerNullFingerprintEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet())}
	tool := &Tool{
		Actions: []*Action{{}},
		GetNextStatesFunc: func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
			return NewStateVec(0), nil
		},
	}
	worker := NewDistributedWorker(0, tool, NewDistributedFPSetManager(endpoint), DistributedWorkerAddress{Hostname: "127.0.0.1", Port: 10997})
	worker.App.checkDeadlock = false
	t.Cleanup(worker.Runtime.executor.Shutdown)
	return worker, endpoint
}

func TestDistributedWorkerRejectsNullFingerprintAnswer(t *testing.T) {
	worker, endpoint := nullFingerprintAnswerWorker(t)
	predecessor := &TLCStateMut{UID: 31, level: 2}
	result, err := worker.GetNextStates([]*TLCStateMut{predecessor})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok {
		t.Fatalf("null FP answer produced success or the wrong failure: result %v, error %T %v", result, err, err)
	}
	if _, ok := failure.Cause.(*NullPointerException); !ok || failure.State1 != predecessor || failure.State2 != nil || !failure.KeepCallStack {
		t.Fatalf("worker lost source failure/context: %#v", failure)
	}
	if endpoint.calls.Load() != 1 || worker.Computing.Load() || worker.OverallStatesComputed.Load() != 0 {
		t.Fatal("worker retried the answer or failed to clear computation state")
	}
}

func TestWorkerRPCNullFingerprintAnswer(t *testing.T) {
	worker, endpoint := nullFingerprintAnswerWorker(t)
	_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
	predecessor := &TLCStateMut{UID: 31, level: 2}
	result, err := client.GetNextStates([]*TLCStateMut{predecessor})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok {
		t.Fatalf("native null FP answer failure: result %v, error %T %v", result, err, err)
	}
	// Native failures carry operation traits rather than reconstructing JVM
	// exception objects. Require the original null-failure classification.
	if !isDistributedNullFailure(failure.Cause) || failure.State1 == nil || failure.State1 == predecessor || failure.State1.UID != 31 || failure.State1.Level() != 2 || failure.State2 != nil || !failure.KeepCallStack {
		t.Fatalf("native worker lost source failure/context: %#v", failure)
	}
	if endpoint.calls.Load() != 1 || worker.Computing.Load() {
		t.Fatal("native call retried the missing fingerprint answer or retained computation state")
	}
}
