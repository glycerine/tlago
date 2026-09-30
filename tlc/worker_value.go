package tlc

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
)

var currentWorkerScope = struct {
	sync.Mutex
	stack map[uint64][]int
}{stack: make(map[uint64][]int)}

var currentStateScope = struct {
	sync.Mutex
	stack map[uint64][]*TLCStateMut
}{stack: make(map[uint64][]*TLCStateMut)}

type WorkerValue struct {
	values []Value
}

func NewWorkerValue(values []Value) *WorkerValue {
	out := make([]Value, len(values))
	copy(out, values)
	return &WorkerValue{values: out}
}

func DemuxWorkerValue(evaluate func() (Value, error), mutable bool, workers int) (any, error) {
	if evaluate == nil {
		return nil, nil
	}
	value, err := evaluate()
	if err != nil {
		return nil, err
	}
	if value != nil {
		value.DeepNormalize()
	}
	if !mutable || !ValueMutates(value) || workers <= 1 {
		return value, nil
	}
	values := make([]Value, workers)
	values[0] = value
	seed := RandomEnumerableSeed()
	for i := 1; i < workers; i++ {
		SetRandomEnumerableSeed(seed)
		v, err := evaluate()
		if err != nil {
			return nil, err
		}
		if v != nil {
			v.DeepNormalize()
		}
		values[i] = v
	}
	return NewWorkerValue(values), nil
}

func ValueMutates(value Value) bool {
	switch value.(type) {
	case nil, *BoolValue, *IntValue, *StringValue, *ModelValue, *IntervalValue, *UndefValue:
		return false
	default:
		return true
	}
}

func MuxWorkerValue(value any, workerID int) Value {
	switch v := value.(type) {
	case nil:
		return nil
	case Value:
		return v
	case *WorkerValue:
		return v.ValueForWorker(workerID)
	default:
		return nil
	}
}

func PushCurrentWorkerID(workerID int) func() {
	gid := currentGoroutineID()
	currentWorkerScope.Lock()
	currentWorkerScope.stack[gid] = append(currentWorkerScope.stack[gid], workerID)
	currentWorkerScope.Unlock()
	return func() {
		currentWorkerScope.Lock()
		stack := currentWorkerScope.stack[gid]
		if len(stack) <= 1 {
			delete(currentWorkerScope.stack, gid)
		} else {
			currentWorkerScope.stack[gid] = stack[:len(stack)-1]
		}
		currentWorkerScope.Unlock()
	}
}

func CurrentWorkerID() (int, bool) {
	gid := currentGoroutineID()
	currentWorkerScope.Lock()
	defer currentWorkerScope.Unlock()
	stack := currentWorkerScope.stack[gid]
	if len(stack) == 0 {
		return 0, false
	}
	return stack[len(stack)-1], true
}

func CurrentThreadIDOr(otherID int) int {
	if id, ok := CurrentWorkerID(); ok {
		return id
	}
	return otherID
}

func PushCurrentState(state *TLCStateMut) func() {
	gid := currentGoroutineID()
	currentStateScope.Lock()
	currentStateScope.stack[gid] = append(currentStateScope.stack[gid], state)
	currentStateScope.Unlock()
	return func() {
		currentStateScope.Lock()
		stack := currentStateScope.stack[gid]
		if len(stack) <= 1 {
			delete(currentStateScope.stack, gid)
		} else {
			currentStateScope.stack[gid] = stack[:len(stack)-1]
		}
		currentStateScope.Unlock()
	}
}

func CurrentState() (*TLCStateMut, bool) {
	gid := currentGoroutineID()
	currentStateScope.Lock()
	defer currentStateScope.Unlock()
	stack := currentStateScope.stack[gid]
	if len(stack) == 0 {
		return nil, false
	}
	return stack[len(stack)-1], true
}

func ResetCurrentState() *TLCStateMut {
	gid := currentGoroutineID()
	currentStateScope.Lock()
	defer currentStateScope.Unlock()
	stack := currentStateScope.stack[gid]
	if len(stack) == 0 {
		return nil
	}
	state := stack[len(stack)-1]
	delete(currentStateScope.stack, gid)
	return state
}

func currentGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	line := buf[:n]
	prefix := []byte("goroutine ")
	if !bytes.HasPrefix(line, prefix) {
		return 0
	}
	start := len(prefix)
	end := start
	for end < len(line) && line[end] >= '0' && line[end] <= '9' {
		end++
	}
	id, err := strconv.ParseUint(string(line[start:end]), 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func (v *WorkerValue) ValueForWorker(workerID int) Value {
	if v == nil || len(v.values) == 0 {
		return nil
	}
	if workerID < 0 || workerID >= len(v.values) {
		workerID = 0
	}
	return v.values[workerID]
}

func workerValueTuple(value any) Value {
	size := NumWorkers()
	if size < 1 {
		size = 1
	}
	switch v := value.(type) {
	case nil:
		return tupleOfUndefined(size)
	case Value:
		values := make([]Value, size)
		for i := range values {
			values[i] = v
		}
		return NewTupleValue(values)
	case *WorkerValue:
		values := make([]Value, size)
		for i := range values {
			if i < len(v.values) && v.values[i] != nil {
				values[i] = v.values[i]
			} else {
				values[i] = ValUndef
			}
		}
		return NewTupleValue(values)
	default:
		return tupleOfUndefined(size)
	}
}

func tupleOfUndefined(size int) Value {
	values := make([]Value, size)
	for i := range values {
		values[i] = ValUndef
	}
	return NewTupleValue(values)
}

func workerIDFromState(state *TLCStateMut) int {
	if state == nil || state.WorkerID == TLCStateInitWorkerID || state.WorkerID < 0 {
		return 0
	}
	return int(state.WorkerID)
}
