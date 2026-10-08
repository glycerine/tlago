package tlc

import (
	"strings"
	"sync/atomic"
	"testing"
)

// No enabled upstream method directly tests native codec failures. These must
// remain remote I/O failures for TLCServerThread's catch boundary, rather than
// becoming worker evaluation failures. The transport does not retry them.
func TestWorkerRPCStateCodecFailures(t *testing.T) {
	for _, phase := range []string{"request_encode", "request_decode", "result_encode"} {
		t.Run(phase, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			var calls atomic.Int32
			invalid := &TLCStateMut{level: 1, values: []Value{&ModelValue{Data: struct{ Name string }{"opaque"}}}}
			_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
				calls.Add(1)
				if phase == "result_encode" {
					return NewNextStateResult([]*StateVec{NewStateVecFrom([]*TLCStateMut{invalid})}, []*LongVec{NewLongVec()}, 0, 1), nil
				}
				return NewNextStateResult([]*StateVec{NewStateVec(0)}, []*LongVec{NewLongVec()}, 0, 0), nil
			}})
			var failure error
			if phase == "request_decode" {
				// Exercise the real server decoder with an invalid graph, not a
				// fabricated endpoint failure or broken TCP stream.
				payload := &DistributedStatePayload{Roots: []int{2}, States: []DistributedStateNode{{Level: 1}}}
				_, failure = client.call(DistributedWorkerRequest{Operation: "next", States: payload})
			} else {
				states := []*TLCStateMut{{level: 1}}
				if phase == "request_encode" {
					states = []*TLCStateMut{invalid}
				}
				result, err := client.GetNextStates(states)
				if result != nil {
					t.Fatal("codec failure returned a result")
				}
				failure = err
			}
			if !isDistributedRemoteFailure(failure) || !isJavaIOException(failure) || isRecoverableDistributedError(failure) {
				t.Fatalf("codec failure lost remote I/O category or became recoverable: %T/%v", failure, failure)
			}
			if _, evaluation := failure.(*WorkerException); evaluation || javaThrowableCause(failure) == nil {
				t.Fatalf("codec failure changed evaluation category or dropped cause: %T/%v", failure, failure)
			}
			want := "unsupported network model-value data"
			wantCalls := int32(0)
			if phase == "request_decode" {
				want = "invalid distributed state reference"
			}
			if phase == "result_encode" {
				wantCalls = 1
			}
			if !strings.Contains(javaThrowableCause(failure).Error(), want) || calls.Load() != wantCalls {
				t.Fatalf("codec cause/calls = %v/%d, want %q/%d", javaThrowableCause(failure), calls.Load(), want, wantCalls)
			}
			if alive, err := client.IsAlive(); err != nil || !alive {
				t.Fatalf("codec failure stopped the native host: %v/%v", alive, err)
			}
			if phase != "request_decode" {
				oldWorkers := NumWorkers()
				SetNumWorkers(2)
				t.Cleanup(func() { SetNumWorkers(oldWorkers) })
				recorder := &MemoryRecorder{}
				AddMessageRecorder(recorder)
				t.Cleanup(func() { RemoveMessageRecorder(recorder) })
				queue := NewMemStateQueue()
				server := &TLCServer{StateQueue: queue}
				thread := &TLCServerThread{Server: server, Worker: NewDistributedWorkerSmartProxy(client), URI: "native-codec-worker"}
				thread.cleanupGlobals.Store(true)
				states := []*TLCStateMut{{UID: 11, level: 1}, {UID: 22, level: 1}}
				if phase == "request_encode" {
					states[0] = invalid
				}
				thread.setStates(states)
				server.RegisterTLCServerThread(thread)
				if result, continuing := thread.computeBlock(queue); result != nil || continuing {
					t.Fatal("coordinator continued after nonrecoverable codec failure")
				}
				if server.IsDone() || queue.Size() != 2 || thread.GetCurrentSize() != 0 || server.GetWorkerCount() != 0 || NumWorkers() != 1 {
					t.Fatal("codec failure did not preserve and requeue the assigned block through worker-loss cleanup")
				}
				if queue.SDequeue() != states[0] || queue.SDequeue() != states[1] {
					t.Fatal("codec failure reordered or replaced assigned states")
				}
				if len(recorder.Records(ECTLCDistributedWorkerLost)) != 1 || len(recorder.Records(ECTLCDistributedWorkerDeregistered)) != 1 || recorder.Recorded(ECGeneral) {
					t.Fatalf("codec failure entered application-error handling: %v", recorder.Messages)
				}
			}
		})
	}
}

func TestWorkerRPCUnevaluatedLazyFailure(t *testing.T) {
	for _, value := range []Value{&LazyValue{}, &LazyValue{Val: ValUndef}, &LazySupplierValue{LazyValue: &LazyValue{}}} {
		var calls atomic.Int32
		_, client := startWorkerRPC(t, &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) {
			calls.Add(1)
			return nil, nil
		}})
		_, err := client.GetNextStates([]*TLCStateMut{{level: 1, values: []Value{value}}})
		if javaRuntimeException(err) == nil || isDistributedRemoteFailure(err) || err.Error() != "Error(TLC): Attempted to serialize lazy value." || calls.Load() != 0 {
			t.Fatalf("unevaluated lazy failure changed source category/detail or dispatched: %T/%v/%d", err, err, calls.Load())
		}
	}
}
