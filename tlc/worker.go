package tlc

type Worker struct {
	ID              int
	LocalValues     []Value
	NamedRegisters  *InsMap[*UniqueString, Value]
	StatesGenerated int64
}

func NewWorker(id int) *Worker {
	return &Worker{
		ID:             id,
		NamedRegisters: NewInsMap[*UniqueString, Value](),
	}
}

func (w *Worker) MyGetID() int {
	if w == nil {
		return 0
	}
	return w.ID
}

func (w *Worker) Start() {}

func (w *Worker) Join() error {
	return nil
}

func (w *Worker) GetLocalValue(index int) Value {
	if w == nil || index < 0 || index >= len(w.LocalValues) {
		return nil
	}
	return w.LocalValues[index]
}

func (w *Worker) SetLocalValue(index int, value Value) {
	if w == nil || index < 0 {
		return
	}
	for len(w.LocalValues) <= index {
		w.LocalValues = append(w.LocalValues, nil)
	}
	w.LocalValues[index] = value
}

func (w *Worker) GetNamedRegister(name *UniqueString) Value {
	if w == nil || w.NamedRegisters == nil {
		return nil
	}
	return w.NamedRegisters.Get(name)
}

func (w *Worker) SetNamedRegister(name *UniqueString, value Value) {
	if w == nil {
		return
	}
	if w.NamedRegisters == nil {
		w.NamedRegisters = NewInsMap[*UniqueString, Value]()
	}
	w.NamedRegisters.Set(name, value)
}
