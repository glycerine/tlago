package tlc

import "time"

type DFIDModelChecker struct {
	*AbstractChecker
	InitStates      []*TLCStateMut
	InitFPs         []uint64
	StatesGenerated int64
}

func NewDFIDModelChecker(tool *Tool, metadir string, deadlock bool) *DFIDModelChecker {
	return &DFIDModelChecker{
		AbstractChecker: NewAbstractChecker(tool, metadir, NewNoopStateWriter(), deadlock, "", time.Now()),
	}
}

func (mc *DFIDModelChecker) ModelCheck() (int, error) {
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "DFID model checker has no tool")
	}
	if result := mc.Tool.CheckAssumptions(); result != NoError {
		return result, nil
	}
	result, err := mc.DoInit(false)
	if err != nil || result != NoError {
		return result, err
	}
	if len(mc.Tool.GetActions()) == 0 {
		return mc.Tool.CheckPostCondition(), nil
	}
	max := Globals.DFIDMax
	if max < 0 {
		max = int(^uint(0) >> 1)
	}
	for level := 2; level <= max; level++ {
		more := false
		for _, init := range mc.InitStates {
			hasMore, result, err := mc.doNext(init, 1, level, make(map[uint64]struct{}))
			if err != nil || result != NoError {
				return result, err
			}
			more = more || hasMore
			if mc.Done {
				return mc.ErrorCode, nil
			}
		}
		if !more {
			return mc.Tool.CheckPostCondition(), nil
		}
		mc.Done = false
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) DoInit(ignoreCancel bool) (int, error) {
	_ = ignoreCancel
	vec := NewStateVec(0)
	if err := mc.Tool.GetInitStates(NewStateFunctor(vec.AddElement)); err != nil {
		return ECGeneral, err
	}
	mc.StatesGenerated += int64(vec.Size())
	mc.InitStates = make([]*TLCStateMut, 0, vec.Size())
	mc.InitFPs = make([]uint64, 0, vec.Size())
	seen := make(map[uint64]struct{})
	for i := 0; i < vec.Size(); i++ {
		state := vec.At(i)
		if !mc.Tool.IsGoodState(state) {
			mc.SetErrState(state, nil, false, ECTLCStateNotCompletelySpecifiedInitial)
			return ECTLCStateNotCompletelySpecifiedInitial, nil
		}
		inModel, err := mc.Tool.IsInModel(state)
		if err != nil {
			return ECGeneral, err
		}
		if !inModel {
			continue
		}
		fp := state.FingerPrint()
		if _, ok := seen[fp]; ok {
			continue
		}
		seen[fp] = struct{}{}
		for _, invariant := range mc.Tool.GetInvariants() {
			valid, err := mc.Tool.IsValidState(invariant, state)
			if err != nil {
				return ECTLCInvariantEvaluationFailed, err
			}
			if !valid {
				mc.SetErrState(state, nil, false, ECTLCInvariantViolatedInitial)
				return ECTLCInvariantViolatedInitial, nil
			}
		}
		mc.InitStates = append(mc.InitStates, state)
		mc.InitFPs = append(mc.InitFPs, fp)
	}
	if vec.Size() == 0 {
		return ECTLCNoStatesSatisfyingInit, nil
	}
	if len(mc.InitStates) == 0 {
		return ECTLCNoStatesSatisfyingInitAndConstraint, nil
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) doNext(cur *TLCStateMut, depth int, maxDepth int, seen map[uint64]struct{}) (bool, int, error) {
	if depth >= maxDepth {
		return true, NoError, nil
	}
	fp := cur.FingerPrint()
	if _, ok := seen[fp]; ok {
		return false, NoError, nil
	}
	seen[fp] = struct{}{}
	defer delete(seen, fp)

	deadlocked := true
	more := false
	for _, action := range mc.Tool.GetActions() {
		nextStates, err := mc.Tool.GetNextStates(action, cur)
		if err != nil {
			return false, ECGeneral, err
		}
		if nextStates == nil || nextStates.Size() == 0 {
			continue
		}
		deadlocked = false
		mc.StatesGenerated += int64(nextStates.Size())
		for i := 0; i < nextStates.Size(); i++ {
			succ := nextStates.At(i).SetPredecessor(cur).SetAction(action)
			if !mc.Tool.IsGoodState(succ) {
				mc.SetErrState(cur, succ, false, ECTLCStateNotCompletelySpecifiedNext)
				return false, ECTLCStateNotCompletelySpecifiedNext, nil
			}
			inModel, err := mc.Tool.IsInModel(succ)
			if err != nil {
				return false, ECGeneral, err
			}
			if inModel {
				inActions, err := mc.Tool.IsInActions(cur, succ)
				if err != nil {
					return false, ECGeneral, err
				}
				inModel = inActions
			}
			if !inModel {
				continue
			}
			for _, invariant := range mc.Tool.GetInvariants() {
				valid, err := mc.Tool.IsValidState(invariant, succ)
				if err != nil {
					return false, ECTLCInvariantEvaluationFailed, err
				}
				if !valid {
					mc.SetErrState(cur, succ, false, ECTLCInvariantViolatedBehavior)
					return false, ECTLCInvariantViolatedBehavior, nil
				}
			}
			childMore, result, err := mc.doNext(succ, depth+1, maxDepth, seen)
			if err != nil || result != NoError {
				return childMore, result, err
			}
			more = more || childMore
		}
	}
	if deadlocked && mc.CheckDeadlock {
		mc.SetErrState(cur, nil, false, ECTLCDeadlockReached)
		return false, ECTLCDeadlockReached, nil
	}
	return more, NoError, nil
}
