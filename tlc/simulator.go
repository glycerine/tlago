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
	return &Simulator{
		Tool:          tool,
		CheckDeadlock: checkDeadlock,
		TraceDepth:    traceDepth,
		TraceNum:      traceNum,
		Seed:          seed,
		Rand:          rand.New(rand.NewSource(seed)),
	}
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
	for trace := int64(0); trace < s.TraceNum; trace++ {
		cur := initStates.At(s.Rand.Intn(initStates.Size())).DeepCopy()
		for depth := 0; depth < s.TraceDepth; depth++ {
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
