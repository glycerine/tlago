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
	DFIDWorkers     []*DFIDWorker
	StatesGenerated int64
}

const dfidInitialSetOfStatesCapacity = 16

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
	if NumWorkers() != 1 {
		return ECGeneral, newTLCError(ECGeneral, "Depth-First Iterative Deepening mode does not support multiple workers. Please run TLC with a single worker.")
	}
	if mc.CheckLiveness {
		return ECGeneral, newTLCError(ECGeneral, "Depth-First Iterative Deepening mode does not support checking liveness properties. Please check liveness properties in Breadth-First-Search mode.")
	}
	if CoverageEnabled() {
		CreateCoverageCostModels(mc.Tool)
		defer ReportCoverage(mc.Tool, mc.StartTime)
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
		worker := NewDFIDWorker(0, level, mc)
		mc.DFIDWorkers = []*DFIDWorker{worker}
		worker.Run()
		mc.Done = false
		if worker.IsTerminated() {
			if worker.Result != NoError || worker.Err != nil {
				return worker.Result, worker.Err
			}
			if mc.ErrorCode != NoError {
				return mc.ErrorCode, nil
			}
			return ECGeneral, worker.Err
		}
		if !worker.HasMoreLevel() {
			return mc.Tool.CheckPostCondition(), nil
		}
		if DoCheckPoint() {
			if err := mc.Checkpoint(); err != nil {
				return ECSystemCheckpointRecoveryCorrupt, err
			}
		}
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) SetErrState(curState *TLCStateMut, succState *TLCStateMut, keepCallStack bool, errorCode int) bool {
	if mc == nil || mc.AbstractChecker == nil {
		return false
	}
	ok := mc.AbstractChecker.SetErrState(curState, succState, keepCallStack, errorCode)
	if ok {
		mc.SetStop(2)
	}
	return ok
}

func (mc *DFIDModelChecker) SetStop(code int) {
	if mc == nil {
		return
	}
	for _, worker := range mc.DFIDWorkers {
		if worker != nil {
			worker.SetStop(code)
		}
	}
}

func (mc *DFIDModelChecker) DoNextInto(cur *TLCStateMut, cfp uint64, isLeaf bool, states *StateVec, fps *LongVec) (bool, int, error) {
	deadlocked := true
	allSuccDone := true
	allSuccNonLeaf := true
	var liveNextStates *SetOfStates
	if mc.CheckLiveness && isLeaf {
		liveNextStates = NewSetOfStates(dfidInitialSetOfStatesCapacity)
	}
	for _, action := range mc.Tool.GetActions() {
		nextStates, err := mc.Tool.GetNextStates(action, cur)
		if err != nil {
			return allSuccNonLeaf, ECGeneral, err
		}
		size := 0
		if nextStates != nil {
			size = nextStates.Size()
		}
		mc.StatesGenerated += int64(size)
		deadlocked = deadlocked && size == 0
		for i := 0; i < size; i++ {
			succ := nextStates.At(i).SetPredecessor(cur).SetAction(action)
			if action != nil && action.CM.node != nil {
				action.CM.IncInvocations()
			}
			if !mc.Tool.IsGoodState(succ) {
				mc.SetErrState(cur, succ, false, ECTLCStateNotCompletelySpecifiedNext)
				return allSuccNonLeaf, ECTLCStateNotCompletelySpecifiedNext, nil
			}
			inModel, err := mc.Tool.IsInModel(succ)
			if err != nil {
				return allSuccNonLeaf, ECGeneral, err
			}
			if inModel {
				inActions, err := mc.Tool.IsInActions(cur, succ)
				if err != nil {
					return allSuccNonLeaf, ECGeneral, err
				}
				inModel = inActions
			}
			status := FPIntStatusNew
			if inModel {
				fp := succ.FingerPrint()
				status = mc.FPSet.SetStatus(fp, FPIntStatusNew)
				allSuccDone = allSuccDone && FPIntSetIsDone(status)
				allSuccNonLeaf = allSuccNonLeaf && !FPIntSetIsLeaf(status)
				if mc.AllStateWriter != nil {
					writeStatus := StateVisitSeen
					if status == FPIntStatusNew {
						writeStatus = StateVisitUnseen
					}
					if err := mc.AllStateWriter.WriteTransition(cur, succ, writeStatus, action); err != nil {
						return allSuccNonLeaf, ECGeneral, err
					}
				}
				if !FPIntSetIsCompleted(status) {
					states.Add(succ)
					fps.AddElement(int64(fp))
				}
				if liveNextStates != nil {
					liveNextStates.PutFP(fp, succ)
				}
			}
			if status == FPIntStatusNew {
				for _, invariant := range mc.Tool.GetInvariants() {
					valid, err := mc.Tool.IsValidState(invariant, succ)
					if err != nil {
						mc.SetErrState(cur, succ, true, ECTLCInvariantEvaluationFailed)
						return allSuccNonLeaf, ECTLCInvariantEvaluationFailed, err
					}
					if !valid {
						mc.SetErrState(cur, succ, false, ECTLCInvariantViolatedBehavior)
						return allSuccNonLeaf, ECTLCInvariantViolatedBehavior, nil
					}
				}
			}
			for _, implied := range mc.Tool.GetImpliedActions() {
				valid, err := mc.Tool.IsValidTransition(implied, cur, succ)
				if err != nil {
					mc.SetErrState(cur, succ, true, ECTLCActionPropertyEvaluationFailed)
					return allSuccNonLeaf, ECTLCActionPropertyEvaluationFailed, err
				}
				if !valid {
					mc.SetErrState(cur, succ, false, ECTLCActionPropertyViolatedBehavior)
					return allSuccNonLeaf, ECTLCActionPropertyViolatedBehavior, nil
				}
			}
		}
	}
	if deadlocked && mc.CheckDeadlock {
		mc.SetErrState(cur, nil, false, ECTLCDeadlockReached)
		return allSuccNonLeaf, ECTLCDeadlockReached, nil
	}
	if liveNextStates != nil {
		liveNextStates.PutFP(cfp, cur)
		if mc.LiveCheck != nil {
			if err := mc.LiveCheck.AddNextState(mc.Tool, cur, cfp, liveNextStates); err != nil {
				return allSuccNonLeaf, ECGeneral, err
			}
		}
	}
	if allSuccDone || (isLeaf && allSuccNonLeaf) {
		mc.FPSet.SetStatus(cfp, FPIntStatusDone)
	}
	return allSuccNonLeaf, NoError, nil
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
		mc.InitStates = append(mc.InitStates, state)
		mc.InitFPs = append(mc.InitFPs, fp)
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
		for _, implied := range mc.Tool.GetImpliedInits() {
			valid, err := mc.Tool.IsValidState(implied, state)
			if err != nil {
				return ECTLCPropertyViolatedInitial, err
			}
			if !valid {
				mc.SetErrState(state, nil, false, ECTLCPropertyViolatedInitial)
				return ECTLCPropertyViolatedInitial, nil
			}
		}
	}
	if vec.Size() == 0 {
		return ECTLCNoStatesSatisfyingInit, nil
	}
	if len(mc.InitStates) == 0 {
		return ECTLCNoStatesSatisfyingInitAndConstraint, nil
	}
	return NoError, nil
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
