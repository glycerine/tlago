package tlc

import (
	"fmt"
	"time"
)

type DFIDModelChecker struct {
	*AbstractChecker
	InitStates      []*TLCStateMut
	InitFPs         []uint64
	FPSet           *MemFPIntSet
	LiveCheck       *LiveCheck
	StatesGenerated int64
}

type DFIDModelCheckerOption func(*DFIDModelChecker)

func WithDFIDFromCheckpoint(fromCheckpoint string) DFIDModelCheckerOption {
	return func(mc *DFIDModelChecker) {
		if mc.AbstractChecker != nil {
			mc.FromCheckpoint = fromCheckpoint
		}
	}
}

func WithDFIDFPSet(fpSet *MemFPIntSet) DFIDModelCheckerOption {
	return func(mc *DFIDModelChecker) {
		mc.FPSet = fpSet
	}
}

func WithDFIDLiveCheck(liveCheck *LiveCheck) DFIDModelCheckerOption {
	return func(mc *DFIDModelChecker) {
		mc.LiveCheck = liveCheck
	}
}

func NewDFIDModelChecker(tool *Tool, metadir string, deadlock bool, opts ...DFIDModelCheckerOption) *DFIDModelChecker {
	rootName := "Spec"
	if tool != nil {
		rootName = tool.GetRootName()
	}
	checkLiveness := false
	if tool != nil {
		checkLiveness = !tool.LivenessIsTrue()
	}
	mc := &DFIDModelChecker{
		AbstractChecker: NewAbstractChecker(tool, metadir, NewNoopStateWriter(), deadlock, "", time.Now()),
		FPSet:           NewMemFPIntSet().Init(NumWorkers(), metadir, rootName),
	}
	mc.CheckLiveness = checkLiveness
	mc.LiveCheck = NewNoOpLiveCheck(tool, metadir)
	for _, opt := range opts {
		opt(mc)
	}
	if mc.FPSet == nil {
		mc.FPSet = NewMemFPIntSet()
	}
	if mc.FPSet.metadir == "" || mc.FPSet.filename == "" {
		mc.FPSet.Init(NumWorkers(), metadir, rootName)
	}
	if mc.LiveCheck == nil {
		mc.LiveCheck = NewNoOpLiveCheck(tool, metadir)
	}
	return mc
}

func (mc *DFIDModelChecker) ModelCheck() (int, error) {
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "DFID model checker has no tool")
	}
	recovered, err := mc.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	if result := mc.Tool.CheckAssumptions(); result != NoError {
		return result, nil
	}
	result, err := mc.DoInit(false)
	if err != nil || result != NoError {
		return result, err
	}
	if recovered {
		PrintMessage(ECTLCInitGenerated3, fmtInt64(mc.StatesGenerated), fmtInt(len(mc.InitStates)))
	} else {
		PrintMessage(ECTLCInitGenerated4, fmtInt64(mc.StatesGenerated), fmtInt(len(mc.InitStates)))
	}
	if len(mc.Tool.GetActions()) == 0 {
		return mc.Tool.CheckPostCondition(), nil
	}
	max := Globals.DFIDMax
	if max < 0 {
		max = int(^uint(0) >> 1)
	}
	for level := 2; level <= max; level++ {
		PrintMessage(ECTLCProgressStartStatsDFID, fmtInt(level), fmtInt64(mc.StatesGenerated), fmtUint64(mc.FPSet.Size()))
		FPIntSetIncLevel()
		more := false
		for _, init := range mc.InitStates {
			hasMore, result, err := mc.doNext(init, init.FingerPrint(), 1, level)
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
		if DoCheckPoint() {
			if err := mc.Checkpoint(); err != nil {
				return ECSystemCheckpointRecoveryCorrupt, err
			}
		}
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) DoPeriodicWork() (int, error) {
	if mc == nil {
		return NoError, nil
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		result, err := mc.LiveCheck.Check(mc.Tool, false)
		if err != nil || result != NoError {
			return result, err
		}
	}
	if DoCheckPoint() {
		if err := mc.Checkpoint(); err != nil {
			return ECSystemCheckpointRecoveryCorrupt, err
		}
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
		status := mc.FPSet.SetStatus(fp, FPIntStatusNew)
		if status != FPIntStatusNew {
			continue
		}
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
		if mc.AllStateWriter != nil {
			if err := mc.AllStateWriter.WriteInitState(state); err != nil {
				return ECGeneral, err
			}
		}
		if mc.CheckLiveness && mc.LiveCheck != nil {
			if err := mc.LiveCheck.AddInitState(mc.Tool, state, fp); err != nil {
				return ECGeneral, err
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

func (mc *DFIDModelChecker) doNext(cur *TLCStateMut, cfp uint64, depth int, maxDepth int) (bool, int, error) {
	if depth >= maxDepth {
		if mc.FPSet != nil {
			mc.FPSet.SetLeveled(cfp)
		}
		return true, NoError, nil
	}
	status := mc.FPSet.GetStatus(cfp)
	if FPIntSetIsCompleted(status) {
		return false, NoError, nil
	}

	deadlocked := true
	more := false
	allSuccDone := true
	allSuccNonLeaf := true
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
			fp := succ.FingerPrint()
			status := mc.FPSet.SetStatus(fp, FPIntStatusNew)
			allSuccDone = allSuccDone && FPIntSetIsDone(status)
			allSuccNonLeaf = allSuccNonLeaf && !FPIntSetIsLeaf(status)
			if mc.AllStateWriter != nil {
				writeStatus := StateVisitSeen
				if status == FPIntStatusNew {
					writeStatus = StateVisitUnseen
				}
				if err := mc.AllStateWriter.WriteTransition(cur, succ, writeStatus, action); err != nil {
					return false, ECGeneral, err
				}
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
			childMore, result, err := mc.doNext(succ, fp, depth+1, maxDepth)
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
	if allSuccDone || (depth == maxDepth-1 && allSuccNonLeaf) {
		mc.FPSet.SetStatus(cfp, FPIntStatusDone)
	}
	if !more {
		mc.FPSet.SetLeveled(cfp)
	}
	return more, NoError, nil
}

func (mc *DFIDModelChecker) Checkpoint() error {
	if mc == nil {
		return nil
	}
	PrintMessage(ECTLCCheckpointStart, mc.Metadir)
	if mc.FPSet != nil {
		if err := mc.FPSet.BeginChkpt(); err != nil {
			return err
		}
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		if err := mc.LiveCheck.BeginChkpt(); err != nil {
			return err
		}
	}
	if err := BeginChkptUniqueStrings(mc.Metadir); err != nil {
		return err
	}
	if mc.FPSet != nil {
		if err := mc.FPSet.CommitChkpt(); err != nil {
			return err
		}
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		if err := mc.LiveCheck.CommitChkpt(); err != nil {
			return err
		}
	}
	if err := CommitChkptUniqueStrings(mc.Metadir); err != nil {
		return err
	}
	PrintMessage(ECTLCCheckpointEnd)
	return nil
}

func (mc *DFIDModelChecker) Recover() (bool, error) {
	if mc == nil || mc.FromCheckpoint == "" {
		return false, nil
	}
	PrintMessage(ECTLCCheckpointRecoverStart, mc.FromCheckpoint)
	if err := RecoverUniqueStrings(mc.FromCheckpoint); err != nil {
		return false, err
	}
	if mc.FPSet != nil {
		rootName := "Spec"
		if mc.Tool != nil {
			rootName = mc.Tool.GetRootName()
		}
		mc.FPSet.Init(NumWorkers(), mc.FromCheckpoint, rootName)
		if err := mc.FPSet.Recover(); err != nil {
			return false, err
		}
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		if err := mc.LiveCheck.Recover(); err != nil {
			return false, err
		}
	}
	size := uint64(0)
	if mc.FPSet != nil {
		size = mc.FPSet.Size()
	}
	PrintMessage(ECTLCCheckpointRecoverEndDFID, fmtUint64(size))
	mc.StatesGenerated = int64(size)
	return true, nil
}

func fmtInt(value int) string {
	return fmt.Sprintf("%d", value)
}

func fmtInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}

func fmtUint64(value uint64) string {
	return fmt.Sprintf("%d", value)
}
