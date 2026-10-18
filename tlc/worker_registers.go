package tlc

import "fmt"

func registerLocalGet(values []Value, index int) Value {
	index = int(int32(index))
	if index < 0 {
		panic(NewArrayIndexOutOfBoundsException(index, len(values)))
	}
	if index >= len(values) {
		return nil
	}
	return values[index]
}
func registerLocalSet(values *[]Value, index int, value Value) {
	index = int(int32(index))
	if index < 0 {
		panic(NewArrayIndexOutOfBoundsException(index, len(*values)))
	}
	if index >= len(*values) {
		size := int32(index) + 1
		if size < 0 {
			panic(NewNegativeArraySizeException(fmt.Sprint(size)))
		}
		expanded := make([]Value, int(size))
		copy(expanded, *values)
		*values = expanded
	}
	(*values)[index] = value
}

func (w *Worker) GetLocalValue(index int) Value {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return registerLocalGet(w.LocalValues, index)
}
func (w *Worker) SetLocalValue(index int, value Value) {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registerLocalSet(&w.LocalValues, index, value)
}
func (w *Worker) GetLocalValues() []Value {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return w.LocalValues
}
func (w *Worker) GetNamedRegister(key *UniqueString) Value {
	if w == nil || key == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registers := w.ownedNamedRegisters()
	if registers == nil {
		return nil
	}
	return registers.Get(key)
}
func (w *Worker) SetNamedRegister(key *UniqueString, value Value) {
	if w == nil || key == nil || isNil(value) {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registers := w.ownedNamedRegisters()
	if registers == nil {
		registers = NewNamedRegisterMap()
		w.NamedRegisters = registers
	}
	registers.Set(key, value)
}
func (w *Worker) GetNamedRegisters() *NamedRegisterMap {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return w.ownedNamedRegisters()
}
func (w *Worker) namedRegisterKeys() []*UniqueString {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	keys := []*UniqueString{}
	if registers := w.ownedNamedRegisters(); registers != nil {
		for key := range registers.All() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (w *DFIDWorker) GetLocalValue(index int) Value {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return registerLocalGet(w.LocalValues, index)
}
func (w *DFIDWorker) SetLocalValue(index int, value Value) {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registerLocalSet(&w.LocalValues, index, value)
}
func (w *DFIDWorker) GetLocalValues() []Value {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return w.LocalValues
}
func (w *DFIDWorker) GetNamedRegister(key *UniqueString) Value {
	if w == nil || key == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registers := w.ownedNamedRegisters()
	if registers == nil {
		return nil
	}
	return registers.Get(key)
}
func (w *DFIDWorker) SetNamedRegister(key *UniqueString, value Value) {
	if w == nil || key == nil || isNil(value) {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registers := w.ownedNamedRegisters()
	if registers == nil {
		registers = NewNamedRegisterMap()
		w.NamedRegisters = registers
	}
	registers.Set(key, value)
}
func (w *DFIDWorker) GetNamedRegisters() *NamedRegisterMap {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return w.ownedNamedRegisters()
}
func (w *DFIDWorker) namedRegisterKeys() []*UniqueString {
	if w == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	keys := []*UniqueString{}
	if registers := w.ownedNamedRegisters(); registers != nil {
		for key := range registers.All() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (w *SimulationWorker) GetLocalValue(index int) Value {
	w.require()
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return registerLocalGet(w.LocalValues, index)
}
func (w *SimulationWorker) SetLocalValue(index int, value Value) {
	w.require()
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registerLocalSet(&w.LocalValues, index, value)
}
func (w *SimulationWorker) GetLocalValues() []Value {
	w.require()
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return w.LocalValues
}
func (w *SimulationWorker) GetNamedRegister(key *UniqueString) Value {
	w.require()
	if key == nil {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registers := w.ownedNamedRegisters()
	if registers == nil {
		return nil
	}
	return registers.Get(key)
}
func (w *SimulationWorker) SetNamedRegister(key *UniqueString, value Value) {
	w.require()
	if key == nil || isNil(value) {
		panic(NewNullPointerException())
	}
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	registers := w.ownedNamedRegisters()
	if registers == nil {
		w.NamedRegisters = NewNamedRegisterMap()
		registers = w.NamedRegisters
	}
	registers.Set(key, value)
}
func (w *SimulationWorker) GetNamedRegisters() *NamedRegisterMap {
	w.require()
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	return w.ownedNamedRegisters()
}
func (w *SimulationWorker) namedRegisterKeys() []*UniqueString {
	w.require()
	w.registerMu.Lock()
	defer w.registerMu.Unlock()
	keys := []*UniqueString{}
	if registers := w.ownedNamedRegisters(); registers != nil {
		for key := range registers.All() {
			keys = append(keys, key)
		}
	}
	return keys
}
