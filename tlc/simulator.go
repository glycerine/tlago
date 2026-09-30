package tlc

import (
	"math/rand"
	"time"
)

type Simulator struct {
	Tool          *Tool
	CheckDeadlock bool
	TraceDepth    int
	TraceNum      int64
	Rand          *rand.Rand
	Seed          int64

	StatesGenerated int64
	TracesGenerated int64
	DisabledRetries int64
	Stopped         bool
	Values          *InsMap[int, Value]
	NamedValues     *InsMap[*UniqueString, Value]
}

func NewSimulator(tool *Tool, deadlock bool, traceDepth int, traceNum int64, seed int64) *Simulator {
	if traceDepth < 0 {
		traceDepth = int(^uint(0) >> 1)
	}
	if traceNum <= 0 {
		traceNum = 1
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	checkDeadlock := deadlock
	if tool != nil && tool.GetModelConfig() != nil {
		checkDeadlock = deadlock && tool.GetModelConfig().GetCheckDeadlock()
	}
	simulator := &Simulator{
		Tool:          tool,
		CheckDeadlock: checkDeadlock,
		TraceDepth:    traceDepth,
		TraceNum:      traceNum,
		Seed:          seed,
		Rand:          rand.New(rand.NewSource(seed)),
		Values:        NewInsMap[int, Value](),
		NamedValues:   NewInsMap[*UniqueString, Value](),
	}
	SetSimulator(simulator)
	return simulator
}

func (s *Simulator) Simulate() (int, error) {
	if s.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "simulator has no tool")
	}
	if result := s.Tool.CheckAssumptions(); result != NoError {
		return result, nil
	}
	initStates, result, err := s.initialStates()
	if err != nil || result != NoError {
		return result, err
	}
	if initStates.IsEmpty() {
		return ECTLCNoStatesSatisfyingInit, nil
	}
	initStates.DeepNormalize()
	for trace := int64(0); trace < s.TraceNum && !s.Stopped; trace++ {
		cur := initStates.At(s.Rand.Intn(initStates.Size())).DeepCopy()
		for depth := 0; depth < s.TraceDepth && !s.Stopped; depth++ {
			next, result, err := s.randomSuccessor(cur)
			if err != nil || result != NoError {
				return result, err
			}
			if next == nil {
				break
			}
			cur = next
		}
		s.TracesGenerated++
	}
	return NoError, nil
}

func (s *Simulator) Stop() {
	if s != nil {
		s.Stopped = true
	}
}

func (s *Simulator) GetLocalValue(idx int) Value {
	if s == nil || s.Values == nil {
		return nil
	}
	return s.Values.Get(idx)
}

func (s *Simulator) SetAllValues(idx int, value Value) {
	if s == nil {
		return
	}
	if s.Values == nil {
		s.Values = NewInsMap[int, Value]()
	}
	s.Values.Set(idx, value)
}

func (s *Simulator) GetAllValues() Value {
	if s == nil || s.Values == nil {
		return EmptyFcn
	}
	domain := make([]Value, 0, s.Values.Len())
	values := make([]Value, 0, s.Values.Len())
	for idx, value := range s.Values.All() {
		domain = append(domain, NewIntValue(int32(idx)))
		values = append(values, value)
	}
	return NewFcnRcdValue(domain, values, false)
}

func (s *Simulator) GetLocalNamedValue(key *UniqueString) Value {
	if s == nil || s.NamedValues == nil {
		return nil
	}
	return s.NamedValues.Get(key)
}

func (s *Simulator) SetAllNamedValues(key *UniqueString, value Value) {
	if s == nil || key == nil {
		return
	}
	if s.NamedValues == nil {
		s.NamedValues = NewInsMap[*UniqueString, Value]()
	}
	s.NamedValues.Set(key, value)
}

func (s *Simulator) GetAllNamedRegisterValues() Value {
	if s == nil || s.NamedValues == nil {
		return EmptyFcn
	}
	domain := make([]Value, 0, s.NamedValues.Len())
	values := make([]Value, 0, s.NamedValues.Len())
	for key, value := range s.NamedValues.All() {
		domain = append(domain, NewStringValueFromUnique(key))
		values = append(values, value)
	}
	return NewFcnRcdValue(domain, values, false)
}

func (s *Simulator) GetAllNamedValues(key *UniqueString) []Value {
	value := s.GetLocalNamedValue(key)
	if value == nil {
		return nil
	}
	return []Value{value}
}

func (s *Simulator) GetStatistics(state *TLCStateMut) Value {
	level := int32(0)
	if state != nil {
		level = int32(state.Level())
	}
	names := []*UniqueString{
		tlcGetGenerated,
		tlcGetDiameter,
		tlcGetRetries,
		tlcGetLevel,
		tlcGetDuration,
	}
	values := []Value{
		intValueFromInt64(s.StatesGenerated),
		intValueFromInt64(s.TracesGenerated),
		intValueFromInt64(s.DisabledRetries),
		NewIntValue(level),
		intValueFromDurationSince(TLCStartTime()),
	}
	return NewRecordValue(names, values, false)
}

func (s *Simulator) GetConfig() Value {
	if s == nil {
		return EmptyRecord
	}
	names := []*UniqueString{
		tlcGetMode,
		tlcGetDeadlock,
		tlcGetSeed,
		tlcGetDepth,
		tlcGetTraces,
		tlcGetWorker,
	}
	values := []Value{
		NewStringValue("simulation"),
		NewBoolValue(s.CheckDeadlock),
		intValueFromInt64(s.Seed),
		NewIntValue(int32(s.TraceDepth)),
		intValueFromInt64(s.TraceNum),
		NewIntValue(1),
	}
	return NewRecordValue(names, values, false)
}

func (s *Simulator) initialStates() (*StateVec, int, error) {
	all := NewStateVec(0)
	err := s.Tool.GetInitStates(NewStateFunctor(func(state *TLCStateMut) (any, error) {
		all.Add(state)
		return all, nil
	}))
	if err != nil {
		return nil, ECGeneral, err
	}
	s.StatesGenerated += int64(all.Size())
	filtered := NewStateVec(all.Size())
	for i := 0; i < all.Size(); i++ {
		state := all.At(i)
		if !s.Tool.IsGoodState(state) {
			return nil, ECTLCStateNotCompletelySpecifiedInitial, nil
		}
		if result, err := s.checkInvariants(state, true); result != NoError || err != nil {
			return nil, result, err
		}
		inModel, err := s.Tool.IsInModel(state)
		if err != nil {
			return nil, ECGeneral, err
		}
		if inModel {
			filtered.Add(state)
		}
	}
	if all.Size() > 0 && filtered.IsEmpty() {
		return nil, ECTLCNoStatesSatisfyingInitAndConstraint, nil
	}
	return filtered, NoError, nil
}

func (s *Simulator) randomSuccessor(cur *TLCStateMut) (*TLCStateMut, int, error) {
	successors := make([]*TLCStateMut, 0)
	actions := make([]*Action, 0)
	for _, action := range s.Tool.GetActions() {
		nextStates, err := s.Tool.GetNextStates(action, cur)
		if err != nil {
			return nil, ECGeneral, err
		}
		if nextStates == nil {
			continue
		}
		s.StatesGenerated += int64(nextStates.Size())
		for i := 0; i < nextStates.Size(); i++ {
			succ := nextStates.At(i)
			if !s.Tool.IsGoodState(succ) {
				return nil, ECTLCStateNotCompletelySpecifiedNext, nil
			}
			succ.SetPredecessor(cur).SetAction(action)
			inModel, err := s.Tool.IsInModel(succ)
			if err != nil {
				return nil, ECGeneral, err
			}
			if inModel {
				inActions, err := s.Tool.IsInActions(cur, succ)
				if err != nil {
					return nil, ECGeneral, err
				}
				inModel = inActions
			}
			if !inModel {
				continue
			}
			if result, err := s.checkInvariants(succ, false); result != NoError || err != nil {
				return nil, result, err
			}
			successors = append(successors, succ)
			actions = append(actions, action)
		}
	}
	if len(successors) == 0 {
		if s.CheckDeadlock {
			return nil, ECTLCDeadlockReached, nil
		}
		return nil, NoError, nil
	}
	idx := s.Rand.Intn(len(successors))
	return successors[idx].SetPredecessor(cur).SetAction(actions[idx]), NoError, nil
}

func (s *Simulator) checkInvariants(state *TLCStateMut, initial bool) (int, error) {
	for _, invariant := range s.Tool.GetInvariants() {
		valid, err := s.Tool.IsValidState(invariant, state)
		if err != nil {
			if initial {
				return ECTLCInvariantEvaluationFailed, err
			}
			return ECTLCInvariantEvaluationFailed, err
		}
		if !valid {
			if initial {
				return ECTLCInvariantViolatedInitial, nil
			}
			return ECTLCInvariantViolatedBehavior, nil
		}
	}
	return NoError, nil
}
