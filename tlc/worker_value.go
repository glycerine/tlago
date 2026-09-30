package tlc

import "sync"

var currentWorkerScope = struct {
	sync.Mutex
	id int
	ok bool
}{}

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
	if !mutable || workers <= 1 {
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
	currentWorkerScope.Lock()
	oldID := currentWorkerScope.id
	oldOK := currentWorkerScope.ok
	currentWorkerScope.id = workerID
	currentWorkerScope.ok = true
	currentWorkerScope.Unlock()
	return func() {
		currentWorkerScope.Lock()
		currentWorkerScope.id = oldID
		currentWorkerScope.ok = oldOK
		currentWorkerScope.Unlock()
	}
}

func CurrentWorkerID() (int, bool) {
	currentWorkerScope.Lock()
	defer currentWorkerScope.Unlock()
	return currentWorkerScope.id, currentWorkerScope.ok
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
