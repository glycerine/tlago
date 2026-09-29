package tlc

import (
	"strings"
)

const (
	TLCStateInitUID   int64 = -1
	TLCStateInitLevel int   = 1
)

type StateVariable struct {
	Name *UniqueString
}

var (
	stateVariables []StateVariable
	EmptyState     *TLCStateMut
)

func SetStateVariables(names []string) {
	stateVariables = make([]StateVariable, len(names))
	SetUniqueStringVariableCount(len(names))
	for i, name := range names {
		us := UniqueStringOf(name)
		us.SetLoc(i)
		stateVariables[i] = StateVariable{Name: us}
	}
	EmptyState = NewEmptyState()
}

func StateVariables() []StateVariable {
	out := make([]StateVariable, len(stateVariables))
	copy(out, stateVariables)
	return out
}

type TLCStateMut struct {
	WorkerID int16
	UID      int64
	level    int
	values   []Value
	pred     *TLCStateMut
	action   *Action
}

func NewEmptyState() *TLCStateMut {
	return &TLCStateMut{
		WorkerID: int16(^uint16(0) >> 1),
		UID:      TLCStateInitUID,
		level:    TLCStateInitLevel,
		values:   make([]Value, len(stateVariables)),
	}
}

func (s *TLCStateMut) CreateEmpty() *TLCStateMut {
	return NewEmptyState()
}

func (s *TLCStateMut) Bind(name *UniqueString, value Value) *TLCStateMut {
	loc := name.VarLoc()
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = value
	}
	return s
}

func (s *TLCStateMut) Unbind(name *UniqueString) *TLCStateMut {
	loc := name.VarLoc()
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = nil
	}
	return s
}

func (s *TLCStateMut) Lookup(name *UniqueString) Value {
	loc := name.VarLoc()
	if loc < 0 || loc >= len(s.values) {
		return nil
	}
	return s.values[loc]
}

func (s *TLCStateMut) ContainsKey(name *UniqueString) bool {
	return s.Lookup(name) != nil
}

func (s *TLCStateMut) Copy() *TLCStateMut {
	values := make([]Value, len(s.values))
	copy(values, s.values)
	return &TLCStateMut{
		WorkerID: s.WorkerID,
		UID:      s.UID,
		level:    s.level,
		values:   values,
		pred:     s.pred,
		action:   s.action,
	}
}

func (s *TLCStateMut) DeepCopy() *TLCStateMut {
	values := make([]Value, len(s.values))
	for i, value := range s.values {
		if value != nil {
			values[i] = value.DeepCopy()
		}
	}
	return &TLCStateMut{
		WorkerID: s.WorkerID,
		UID:      s.UID,
		level:    s.level,
		values:   values,
		pred:     s.pred,
		action:   s.action,
	}
}

func (s *TLCStateMut) DeepNormalize() {
	for _, value := range s.values {
		if value != nil {
			value.DeepNormalize()
		}
	}
}

func (s *TLCStateMut) FingerPrint() uint64 {
	fp := FP64New()
	for _, value := range s.values {
		if value != nil {
			fp = value.FingerPrint(fp)
		}
	}
	return fp
}

func (s *TLCStateMut) AllAssigned() bool {
	for _, value := range s.values {
		if value == nil {
			return false
		}
	}
	return true
}

func (s *TLCStateMut) NoneAssigned() bool {
	for _, value := range s.values {
		if value != nil {
			return false
		}
	}
	return true
}

func (s *TLCStateMut) Unassigned() []StateVariable {
	var out []StateVariable
	for i, value := range s.values {
		if value == nil && i < len(stateVariables) {
			out = append(out, stateVariables[i])
		}
	}
	return out
}

func (s *TLCStateMut) Values() *InsMap[*UniqueString, Value] {
	out := NewInsMap[*UniqueString, Value]()
	for i, variable := range stateVariables {
		if i < len(s.values) {
			out.Set(variable.Name, s.values[i])
		}
	}
	return out
}

func (s *TLCStateMut) SetPredecessor(pred *TLCStateMut) *TLCStateMut {
	if pred != nil {
		s.pred = pred
		s.level = pred.level + 1
	}
	return s
}

func (s *TLCStateMut) UnsetPredecessor() *TLCStateMut {
	s.pred = nil
	return s
}

func (s *TLCStateMut) Predecessor() *TLCStateMut {
	return s.pred
}

func (s *TLCStateMut) Level() int {
	return s.level
}

func (s *TLCStateMut) IsInitial() bool {
	return s.level == TLCStateInitLevel
}

func (s *TLCStateMut) SetAction(action *Action) *TLCStateMut {
	s.action = action
	return s
}

func (s *TLCStateMut) HasAction() bool {
	return s != nil && s.action != nil
}

func (s *TLCStateMut) GetAction() *Action {
	if s == nil || s.action == nil {
		return UnknownAction
	}
	return s.action
}

func (s *TLCStateMut) CopyWith(prototype *TLCStateMut) *TLCStateMut {
	out := NewEmptyState()
	for i, variable := range stateVariables {
		if prototype != nil && prototype.Lookup(variable.Name) != nil {
			out.values[i] = s.Lookup(variable.Name)
		}
	}
	out.level = s.level
	return out
}

func (s *TLCStateMut) Equal(other *TLCStateMut) bool {
	if s == nil || other == nil {
		return s == other
	}
	if len(s.values) != len(other.values) {
		return false
	}
	for i := range s.values {
		if !stateValuesEqual(s.values[i], other.values[i]) {
			return false
		}
	}
	return true
}

func (s *TLCStateMut) String() string {
	return s.StringForVariables(nil)
}

func (s *TLCStateMut) StringForVariables(last *TLCStateMut, vars ...*UniqueString) string {
	if len(vars) == 0 {
		vars = make([]*UniqueString, len(stateVariables))
		for i, variable := range stateVariables {
			vars[i] = variable.Name
		}
	}

	var b strings.Builder
	if len(vars) == 1 {
		key := vars[0]
		value := s.Lookup(key)
		if last == nil || !stateValuesEqual(last.Lookup(key), value) {
			b.WriteString(key.String())
			b.WriteString(" = ")
			b.WriteString(valueString(value))
			b.WriteString("\n")
		}
		return b.String()
	}

	for _, key := range vars {
		value := s.Lookup(key)
		if last != nil && stateValuesEqual(last.Lookup(key), value) {
			continue
		}
		b.WriteString("/\\ ")
		b.WriteString(key.String())
		b.WriteString(" = ")
		b.WriteString(valueString(value))
		b.WriteString("\n")
	}
	return b.String()
}

func stateValuesEqual(a, b Value) bool {
	if a == nil || b == nil {
		return a == b
	}
	eq, err := a.Equal(b)
	return err == nil && eq
}

func valueString(v Value) string {
	if v == nil {
		return "<nil>"
	}
	return v.String()
}

type StateVec struct {
	states []*TLCStateMut
}

func NewStateVec(capacity int) *StateVec {
	if capacity < 0 {
		capacity = 0
	}
	return &StateVec{states: make([]*TLCStateMut, 0, capacity)}
}

func NewStateVecFrom(states []*TLCStateMut) *StateVec {
	out := make([]*TLCStateMut, len(states))
	copy(out, states)
	return &StateVec{states: out}
}

func (v *StateVec) Empty() bool           { return len(v.states) == 0 }
func (v *StateVec) IsEmpty() bool         { return len(v.states) == 0 }
func (v *StateVec) Size() int             { return len(v.states) }
func (v *StateVec) At(i int) *TLCStateMut { return v.states[i] }
func (v *StateVec) First() *TLCStateMut   { return v.states[0] }
func (v *StateVec) Last() *TLCStateMut    { return v.states[len(v.states)-1] }
func (v *StateVec) Clear()                { v.states = v.states[:0] }

func (v *StateVec) Add(state *TLCStateMut) *StateVec {
	v.states = append(v.states, state)
	return v
}

func (v *StateVec) AddElement(state *TLCStateMut) (any, error) {
	v.Add(state)
	return v, nil
}

func (v *StateVec) SetElement(state *TLCStateMut) (any, error) {
	v.Clear()
	v.Add(state)
	return v, nil
}

func (v *StateVec) HasStates() bool {
	return !v.IsEmpty()
}

func (v *StateVec) AddWithPredecessor(pred *TLCStateMut, state *TLCStateMut) *StateVec {
	return v.Add(state.SetPredecessor(pred))
}

func (v *StateVec) AddNextElement(pred *TLCStateMut, action *Action, state *TLCStateMut) (any, error) {
	if state != nil {
		state.SetPredecessor(pred).SetAction(action)
	}
	return v.Add(state), nil
}

func (v *StateVec) AddElements(other *StateVec) *StateVec {
	if v == nil {
		return other
	}
	if other == nil {
		return v
	}
	target := v
	source := other
	if source.Size() > target.Size() {
		target, source = source, target
	}
	target.states = append(target.states, source.states...)
	return target
}

func (v *StateVec) AddAll(other *StateVec) *StateVec {
	return v.AddElements(other)
}

func (v *StateVec) Remove(index int) {
	v.states[index] = v.states[len(v.states)-1]
	v.states = v.states[:len(v.states)-1]
}

func (v *StateVec) Replace(index int, state *TLCStateMut) {
	v.states[index] = state
}

func (v *StateVec) Copy() *StateVec {
	out := NewStateVec(len(v.states))
	for _, state := range v.states {
		out.Add(state.Copy())
	}
	return out
}

func (v *StateVec) DeepCopy() *StateVec {
	out := NewStateVec(len(v.states))
	for _, state := range v.states {
		out.Add(state.DeepCopy())
	}
	return out
}

func (v *StateVec) DeepNormalize() {
	for _, state := range v.states {
		state.DeepNormalize()
	}
}

func (v *StateVec) Contains(state *TLCStateMut) bool {
	fp := state.FingerPrint()
	for _, candidate := range v.states {
		if candidate.FingerPrint() == fp {
			return true
		}
	}
	return false
}

func (v *StateVec) ToSlice() []*TLCStateMut {
	out := make([]*TLCStateMut, len(v.states))
	copy(out, v.states)
	return out
}

func (v *StateVec) String() string {
	parts := make([]string, len(v.states))
	for i, state := range v.states {
		parts[i] = state.String()
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
