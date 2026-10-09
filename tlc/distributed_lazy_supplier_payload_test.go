package tlc

import (
	"strings"
	"sync/atomic"
	"testing"
)

// LazySupplierValue.getValue consults its supplier regardless of cached val.
// It must not become an ordinary lazy value during native payload transfer.
func TestDistributedLazySupplierNullEvaluation(t *testing.T) {
	for _, cached := range []Value{nil, ValUndef, NewIntValue(7)} {
		value := NewLazySupplierValue(nil, nil)
		value.Val = cached
		func() {
			defer func() {
				if failure := recover(); failure == nil {
					t.Fatal("nil supplier silently returned a value")
				} else if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("nil supplier failure = %T/%v", failure, failure)
				}
			}()
			_, _ = value.GetValue(nil, nil, nil, EvalClear)
		}()
	}
}

func lazySupplierPayloadStates() []*TLCStateMut {
	cached := NewIntValue(7)
	value := NewLazySupplierValue(nil, nil)
	value.Val = cached
	return []*TLCStateMut{{level: 1, values: []Value{value, value, cached}}}
}

func checkLazySupplierPayload(t *testing.T, states []*TLCStateMut) {
	t.Helper()
	values := states[0].values
	value, ok := values[0].(*LazySupplierValue)
	if !ok || values[1] != value || value.Val != values[2] || value.Supplier != nil || value.Expr != nil {
		t.Fatalf("supplier type or graph identity lost: %T", values[0])
	}
	defer func() {
		if failure := recover(); failure == nil {
			t.Fatal("received nil supplier silently evaluated its cached value")
		} else if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("received nil supplier failure = %T/%v", failure, failure)
		}
	}()
	_, _ = value.GetValue(nil, nil, nil, EvalClear)
}

func TestDistributedLazySupplierPayload(t *testing.T) {
	states := lazySupplierPayloadStates()
	got := distributedPayloadRoundTrip(t, states)
	checkLazySupplierPayload(t, got)
	got[0].values[2].(*IntValue).Val = 9
	if states[0].values[2].(*IntValue).Val != 7 {
		t.Fatal("received supplier cache aliases sender")
	}
}

func TestWorkerRPCLazySupplierPayload(t *testing.T) {
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if _, ok := states[0].values[0].(*LazySupplierValue); !ok {
			return nil, NewRuntimeException("request lost supplier type")
		}
		if states[0].UID == 1 {
			return nil, NewWorkerException("supplier context", states[0], states[0], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVec()}, 1, 0), nil
	}})
	states := lazySupplierPayloadStates()
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	checkLazySupplierPayload(t, []*TLCStateMut{result.NextStates[0].At(0)})
	states[0].UID = 1
	result, err = client.GetNextStates(states)
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("supplier failure context = %v/%v", result, err)
	}
	checkLazySupplierPayload(t, []*TLCStateMut{failure.State1})
}

func TestWorkerRPCRejectsExecutableLazySupplier(t *testing.T) {
	var supplierCalls, workerCalls atomic.Int32
	value := NewLazySupplierValue(nil, func() Value { supplierCalls.Add(1); return NewIntValue(99) })
	value.Val = NewIntValue(7)
	states := []*TLCStateMut{{level: 1, values: []Value{value}}}
	if _, err := EncodeDistributedStates(states); err == nil || !strings.Contains(err.Error(), "unsupported network lazy supplier") {
		t.Fatalf("executable supplier flattened into cached value: %v", err)
	}
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func([]*TLCStateMut) (*NextStateResult, error) {
		workerCalls.Add(1)
		return nil, nil
	}})
	result, err := client.GetNextStates(states)
	if result != nil || !isDistributedRemoteFailure(err) || isRecoverableDistributedError(err) || javaThrowableCause(err) == nil || !strings.Contains(javaThrowableCause(err).Error(), "unsupported network lazy supplier") || supplierCalls.Load() != 0 || workerCalls.Load() != 0 {
		t.Fatalf("supplier rejection evaluated or dispatched code: %v/%v, calls %d/%d", result, err, supplierCalls.Load(), workerCalls.Load())
	}
}

func TestDistributedLazySupplierPayloadInvalid(t *testing.T) {
	for _, refs := range [][]int{nil, {0, 0}, {2}} {
		payload := &DistributedStatePayload{Values: []DistributedValueNode{{Kind: "lazySupplier", References: refs}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid supplier cache references accepted: %v", refs)
		}
	}
}
