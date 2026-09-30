package tlc

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

func (v *WorkerValue) ValueForWorker(workerID int) Value {
	if v == nil || len(v.values) == 0 {
		return nil
	}
	if workerID < 0 || workerID >= len(v.values) {
		workerID = 0
	}
	return v.values[workerID]
}

func workerIDFromState(state *TLCStateMut) int {
	if state == nil || state.WorkerID == TLCStateInitWorkerID || state.WorkerID < 0 {
		return 0
	}
	return int(state.WorkerID)
}
