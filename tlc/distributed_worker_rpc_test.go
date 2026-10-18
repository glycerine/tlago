package tlc

import (
	"errors"
	"io"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

// The upstream remote model harness is assumption-disabled. These checks
// exercise the new Go transport; they do not substitute for those model tests.
type rpcTestWorker struct {
	next  func([]*TLCStateMut) (*NextStateResult, error)
	exits atomic.Int64
}

func (w *rpcTestWorker) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	return w.next(states)
}
func (w *rpcTestWorker) IsAlive() (bool, error)              { return true, nil }
func (w *rpcTestWorker) Exit() error                         { w.exits.Add(1); return nil }
func (w *rpcTestWorker) GetURI() (string, error)             { return "tcp://worker:1234/primary", nil }
func (w *rpcTestWorker) GetCacheRateRatio() (float64, error) { return math.NaN(), nil }

func startWorkerRPC(t *testing.T, endpoint DistributedWorkerEndpoint) (*DistributedRPCServer, *NetworkWorkerEndpoint) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := NewDistributedRPCServer()
	if err := server.RegisterWorker("primary", endpoint); err != nil {
		_ = listener.Close()
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Error(err)
		}
		if err := <-done; err != nil && !errors.Is(err, net.ErrClosed) {
			t.Error(err)
		}
	})
	client, err := DialWorkerEndpoint(listener.Addr().String(), "primary")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.CloseConnection() })
	return server, client
}

func TestWorkerRPCStateResultAndLifecycle(t *testing.T) {
	state := &TLCStateMut{WorkerID: 7, UID: -1, level: 40000, values: []Value{NewIntValue(42)}}
	op := NewOpRcdValueFrom([][]Value{{state.values[0]}}, []Value{NewStringValue("answer")})
	state.values = append(state.values, op)
	var calls atomic.Int64
	worker := &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		calls.Add(1)
		if states == nil {
			return nil, nil
		}
		if len(states) == 0 {
			return NewNextStateResult([]*StateVec{}, []*LongVec{}, 0, -1), nil
		}
		if len(states) != 3 || states[0] == state || states[0] != states[1] || states[2] != nil || states[0].level != 40000 || states[0].UID != -1 {
			return nil, errors.New("request graph changed")
		}
		operator, ok := states[0].values[1].(*OpRcdValue)
		if !ok || operator == op || operator.Domain[0][0] != states[0].values[0] {
			return nil, errors.New("request constant operator changed")
		}
		if result, err := operator.Eval([]Value{NewIntValue(42)}, 0); err != nil || result.String() != `"answer"` {
			return nil, errors.New("request constant operator cannot be applied")
		}
		vector := NewStateVecFrom(states)
		fps := NewLongVecFrom([]int64{math.MinInt64, -1, math.MaxInt64})
		return NewNextStateResult([]*StateVec{vector, nil, vector}, []*LongVec{fps, nil, fps}, math.MaxInt64, -1), nil
	}}
	server, client := startWorkerRPC(t, worker)
	got, err := client.GetNextStates([]*TLCStateMut{state, state, nil})
	if err != nil {
		t.Fatal(err)
	}
	if got.NextStates[0] != got.NextStates[2] || got.NextStates[0].At(0) != got.NextStates[0].At(1) || got.NextStates[0].At(0) == state || got.NextStates[1] != nil || got.NextStates[0].At(2) != nil {
		t.Fatal("result graph changed")
	}
	if got.NextFingerprints[0] != got.NextFingerprints[2] || got.NextFingerprints[0].At(0) != math.MinInt64 || got.ComputationTime != math.MaxInt64 || got.StatesComputed != -1 {
		t.Fatal("result fingerprints/counters changed")
	}
	returned := got.NextStates[0].At(0)
	if operator := returned.values[1].(*OpRcdValue); operator == op || operator.Domain[0][0] != returned.values[0] {
		t.Fatal("result constant operator changed")
	}
	if got, err := client.GetNextStates(nil); err != nil || got != nil {
		t.Fatalf("null request/result: %v/%v", got, err)
	}
	if got, err := client.GetNextStates([]*TLCStateMut{}); err != nil || got == nil || got.NextStates == nil || len(got.NextStates) != 0 {
		t.Fatalf("empty request/result: %v/%v", got, err)
	}
	if alive, err := client.IsAlive(); err != nil || !alive {
		t.Fatalf("alive: %v/%v", alive, err)
	}
	if uri, err := client.GetURI(); err != nil || uri != "tcp://worker:1234/primary" {
		t.Fatalf("uri: %q/%v", uri, err)
	}
	if ratio, err := client.GetCacheRateRatio(); err != nil || !math.IsNaN(ratio) {
		t.Fatalf("cache: %v/%v", ratio, err)
	}
	if err := server.RegisterWorker("primary", worker); err == nil {
		t.Fatal("duplicate worker accepted")
	}
	if err := client.Exit(); err != nil {
		t.Fatal(err)
	}
	if _, err := client.IsAlive(); !isDistributedRemoteFailure(err) {
		t.Fatalf("unregistered worker: %v", err)
	}
	if calls.Load() != 3 || worker.exits.Load() != 1 {
		t.Fatal("calls were retried or exit skipped")
	}
	if err := client.Exit(); !isIgnorableDistributedWorkerExit(err) {
		t.Fatalf("repeated exit was not recognized as unavailable: %v", err)
	}
}

func TestWorkerRPCFailureStatesAndCauses(t *testing.T) {
	state := &TLCStateMut{UID: -1, level: 40000, values: []Value{NewIntValue(9)}}
	cause := NewRuntimeException()
	failure := NewWorkerExceptionWithCause("evaluation failed", cause, state, state, true)
	failure.addSuppressedError(cause)
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) { return nil, failure }})
	_, err := client.GetNextStates(nil)
	got, ok := err.(*WorkerException)
	if !ok || got.Msg != failure.Msg || !got.KeepCallStack || got.State1 == state || got.State1 != got.State2 || got.State1.UID != -1 || got.State1.level != 40000 {
		t.Fatalf("worker failure changed: %T/%v", err, err)
	}
	if got.Cause == cause || javaThrowableDetailMessage(got.Cause) != nil || javaThrowableClassName(got.Cause) != javaThrowableClassName(cause) {
		t.Fatal("cause diagnostics changed")
	}
	if len(got.GetSuppressed()) != 1 || got.GetSuppressed()[0] != got.Cause {
		t.Fatal("shared suppressed cause changed")
	}
	if javaThrowableStackTrace(got) != javaThrowableStackTrace(failure) {
		t.Fatal("sender stack replaced with decoder stack")
	}
	if isDistributedRemoteFailure(got) {
		t.Fatal("evaluation error became a connection failure")
	}
}

func TestWorkerRPCFailureClassification(t *testing.T) {
	cases := []struct {
		name                      string
		failure                   error
		remote, recoverable, null bool
	}{
		{"truncated", workerConnectionFailure(io.EOF), true, true, false},
		{"eof detail", &DistributedOperationError{Message: javaString("decode"), Cause: NewEOFException("detail"), Remote: true, IO: true}, true, false, false},
		{"worker memory", workerComputationFailure("memory", NewOutOfMemoryError(), true), true, true, false},
		{"direct memory", &DistributedOperationError{Cause: NewOutOfMemoryError(), Remote: true, IO: true}, true, false, false},
		{"connection", workerConnectionFailure(net.ErrClosed), true, false, false},
		{"null", NewNullPointerException(), false, false, true},
		{"runtime", NewRuntimeException("broken"), false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, client := startWorkerRPC(t, &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) { return nil, tc.failure }})
			_, err := client.GetNextStates(nil)
			if err == nil || isDistributedRemoteFailure(err) != tc.remote || isRecoverableDistributedError(err) != tc.recoverable || isDistributedNullFailure(err) != tc.null {
				t.Fatalf("classification changed: %T/%v", err, err)
			}
		})
	}
}

func TestWorkerRPCConcurrentCalls(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	worker := &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) {
		close(entered)
		<-release
		return NewNextStateResult([]*StateVec{}, []*LongVec{}, 0, 0), nil
	}}
	_, client := startWorkerRPC(t, worker)
	var group sync.WaitGroup
	group.Add(1)
	go func() {
		defer group.Done()
		if _, err := client.GetNextStates(nil); err != nil {
			t.Error(err)
		}
	}()
	<-entered
	// Keepalive must not wait behind a long computation on this connection.
	alive, err := client.IsAlive()
	close(release)
	group.Wait()
	if err != nil || !alive {
		t.Fatalf("concurrent keepalive: %v/%v", alive, err)
	}
}

func TestWorkerRPCPanicAndShutdown(t *testing.T) {
	worker := &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) { panic(NewNullPointerException("panic")) }}
	server, client := startWorkerRPC(t, worker)
	if _, err := client.GetNextStates(nil); !isDistributedNullFailure(err) {
		t.Fatalf("panic classification: %v", err)
	}
	if alive, err := client.IsAlive(); err != nil || !alive {
		t.Fatalf("panic killed service: %v/%v", alive, err)
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := client.IsAlive(); !isDistributedRemoteFailure(err) {
		t.Fatalf("closed transport: %v", err)
	}
	if err := client.Exit(); !isIgnorableDistributedWorkerExit(err) {
		t.Fatalf("dead worker exit: %v", err)
	}
	if worker.exits.Load() != 0 {
		t.Fatal("transport close exited runtime")
	}
}

func TestWorkerRPCCoordinatorRetryAndLoss(t *testing.T) {
	previousWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(previousWorkers) })
	for _, recoverable := range []bool{true, false} {
		t.Run(strconv.FormatBool(recoverable), func(t *testing.T) {
			SetNumWorkers(1)
			failure := workerConnectionFailure(io.EOF)
			if !recoverable {
				failure = workerConnectionFailure(net.ErrClosed)
			}
			_, client := startWorkerRPC(t, &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) { return nil, failure }})
			queue := NewDiskStateQueue(t.TempDir())
			t.Cleanup(queue.FinishAll)
			selector := &BlockSelector{Mode: BlockSelectorLimiting, Maximum: 100}
			thread := &TLCServerThread{Server: &TLCServer{StateQueue: queue}, Worker: NewDistributedWorkerSmartProxy(client), Selector: selector, URI: "tcp://worker/primary", keepAliveDone: make(chan struct{})}
			thread.cleanupGlobals.Store(true)
			states := []*TLCStateMut{{values: []Value{NewIntValue(1)}}, {values: []Value{NewIntValue(2)}}}
			thread.setStates(states)
			result, proceed := thread.computeBlock(queue)
			if result != nil || proceed != recoverable || queue.Size() != 2 {
				t.Fatalf("retry/loss result = %v/%v, queue = %d", result, proceed, queue.Size())
			}
			if queue.SDequeue() != states[0] || queue.SDequeue() != states[1] {
				t.Fatal("unfinished work lost or reordered")
			}
			if recoverable {
				if selector.Maximum != 1 || NumWorkers() != 1 || thread.GetCurrentSize() != 2 {
					t.Fatal("retry failed to halve block while retaining worker")
				}
			} else {
				if selector.Maximum != 100 || NumWorkers() != 0 || thread.GetCurrentSize() != 0 {
					t.Fatal("worker loss did not deregister worker")
				}
				thread.HandleRemoteWorkerLost(queue)
				if queue.Size() != 0 || NumWorkers() != 0 {
					t.Fatal("worker loss was not idempotent")
				}
			}
		})
	}
}
