package tlc

import (
	"errors"
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
	CleanupEnabled  bool
	cleanupDone     bool
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

func WithDFIDCleanup(cleanup bool) DFIDModelCheckerOption {
	return func(mc *DFIDModelChecker) {
		mc.CleanupEnabled = cleanup
	}
}

func NewDFIDModelChecker(tool *Tool, metadir string, deadlock bool, opts ...DFIDModelCheckerOption) *DFIDModelChecker {
	if tool != nil && tool.GetMode() != ModeDebugger && tool.GetMode() != ModeExecutor {
		tool.SetMode(ModeMCDFS)
	}
	SetTLCStateTool(tool)
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
		CleanupEnabled:  true,
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

func (mc *DFIDModelChecker) ModelCheck() (result int, err error) {
	result = NoError
	defer func() {
		if mc == nil {
			return
		}
		if cleanupErr := mc.Cleanup(result == NoError, mc.CleanupEnabled); err == nil {
			err = cleanupErr
		}
	}()
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "DFID model checker has no tool")
	}
	if NumWorkers() != 1 {
		return ECGeneral, newTLCError(ECGeneral, "Depth-First Iterative Deepening mode does not support multiple workers. Please run TLC with a single worker.")
	}
	if mc.CheckLiveness {
		return ECGeneral, newTLCError(ECGeneral, "Depth-First Iterative Deepening mode does not support checking liveness properties (https://github.com/tlaplus/tlaplus/issues/548).  Please check liveness properties in Breadth-First-Search mode.")
	}
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(mc.Tool)
	}
	recovered, err := mc.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	if mc.CheckLiveness && mc.LiveCheck != nil && mc.LiveCheck.NumChecker() == 0 {
		PrintError(ECTLCLiveFormulaTautology)
		return ECTLCLiveFormulaTautology, nil
	}
	if result := mc.Tool.CheckAssumptions(); result != NoError {
		return result, nil
	}
	result, err = mc.DoInit(false)
	if err != nil {
		result = mc.reportInitException(result, err)
		mc.PrintSummary(false)
		return result, err
	}
	if result != NoError {
		return result, nil
	}
	if recovered {
		PrintMessage(ECTLCInitGenerated3, fmtInt64(mc.StatesGenerated), fmtInt(len(mc.InitStates)))
	} else {
		PrintMessage(ECTLCInitGenerated4, fmtInt64(mc.StatesGenerated), fmtInt(len(mc.InitStates)))
	}
	if len(mc.Tool.GetActions()) == 0 {
		ReportSuccessCountsDistance(mc.FPSet.Size(), mc.FPSet.CheckFPs(), mc.StatesGenerated)
		mc.PrintSummary(true)
		return NoError, nil
	}
	max := Globals.DFIDMax
	if max < 0 {
		max = int(^uint(0) >> 1)
	}
	terminated := false
	for level := 2; level <= max; level++ {
		if terminated {
			return mc.finishTerminatedDFID()
		}
		PrintMessage(ECTLCProgressStartStatsDFID, fmtInt(level), fmtInt64(mc.StatesGenerated), fmtUint64(mc.FPSet.Size()))
		FPIntSetIncLevel()
		result, err = mc.RunTLC(level)
		mc.Done = false
		if result != NoError || err != nil {
			mc.PrintSummary(false)
			return result, err
		}
		worker := mc.dfidWorkerAt(0)
		if worker != nil && worker.IsTerminated() {
			terminated = true
		}
		moreLevel := false
		if worker != nil && worker.HasMoreLevel() {
			moreLevel = true
		}
		terminated = terminated || !moreLevel
	}
	result = NoError
	mc.PrintSummary(true)
	return result, nil
}

func (mc *DFIDModelChecker) RunTLC(depth int) (int, error) {
	if depth < 2 {
		return NoError, nil
	}
	worker := NewDFIDWorker(0, depth, mc)
	mc.DFIDWorkers = []*DFIDWorker{worker}
	done := make(chan struct{}, 1)
	go func() {
		worker.Run()
		done <- struct{}{}
	}()
	select {
	case <-done:
		return mc.dfidWorkerResult(worker)
	case <-time.After(3 * time.Second):
	}
	interval := ProgressInterval()
	coverageCountdown := periodicCoverageCountdown(interval)
	for {
		result, err := mc.DoPeriodicWork()
		if err != nil || result != NoError {
			return result, err
		}
		if mc.isDFIDDone() {
			<-done
			return mc.dfidWorkerResult(worker)
		}
		mc.runTLCContinueDoing(coverageCountdown, depth)
		if coverageCountdown == 0 {
			coverageCountdown = periodicCoverageCountdown(interval)
		} else if coverageCountdown > 0 {
			coverageCountdown--
		}
		if mc.isDFIDDone() {
			<-done
			return mc.dfidWorkerResult(worker)
		}
		if interval <= 0 {
			select {
			case <-done:
				return mc.dfidWorkerResult(worker)
			default:
				continue
			}
		}
		timer := time.NewTimer(interval)
		select {
		case <-done:
			timer.Stop()
			return mc.dfidWorkerResult(worker)
		case <-timer.C:
		}
	}
}

func (mc *DFIDModelChecker) dfidWorkerAt(index int) *DFIDWorker {
	if mc == nil || index < 0 || index >= len(mc.DFIDWorkers) {
		return nil
	}
	return mc.DFIDWorkers[index]
}

func (mc *DFIDModelChecker) dfidWorkerResult(worker *DFIDWorker) (int, error) {
	if worker != nil && (worker.Result != NoError || worker.Err != nil) {
		if mc != nil && mc.KeepCallStack {
			return NoError, nil
		}
		return worker.Result, worker.Err
	}
	if mc != nil && !mc.KeepCallStack && mc.ErrorCode != NoError {
		return mc.ErrorCode, nil
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) isDFIDDone() bool {
	if mc == nil || mc.AbstractChecker == nil {
		return true
	}
	mc.mu.Lock()
	done := mc.Done
	mc.mu.Unlock()
	return done
}

func (mc *DFIDModelChecker) runTLCContinueDoing(count int, depth int) {
	_ = depth
	PrintMessage(ECTLCProgressStatsDFID, fmtInt64(mc.StatesGenerated), fmtUint64(mc.FPSet.Size()))
	if count == 0 {
		if mc != nil && mc.Tool != nil && CoverageAnyEnabled() && len(mc.Tool.GetActions()) > 0 {
			reportCoverage(mc.Tool)
		}
	}
}

func (mc *DFIDModelChecker) finishTerminatedDFID() (int, error) {
	if mc == nil {
		return ECGeneral, nil
	}
	result := NoError
	if mc.ErrState == nil {
		if mc.CheckLiveness && mc.LiveCheck != nil {
			PrintMessage(ECTLCProgressStatsDFID, fmtInt64(mc.StatesGenerated), fmtUint64(mc.FPSet.Size()))
			result, err := mc.LiveCheck.FinalCheck(mc.Tool)
			if err != nil || result != NoError {
				mc.PrintSummary(false)
				return result, err
			}
		}
		ReportSuccessCountsDistance(mc.FPSet.Size(), mc.FPSet.CheckFPs(), mc.StatesGenerated)
		mc.PrintSummary(true)
		return NoError, nil
	}
	if mc.KeepCallStack {
		result = mc.replayDFIDNextErrorCallStack()
	} else if mc.ErrorCode != NoError {
		result = mc.ErrorCode
	} else {
		result = ECGeneral
	}
	mc.PrintSummary(result == NoError)
	return result, nil
}

func (mc *DFIDModelChecker) PrintSummary(success bool) {
	_ = success
	if mc == nil {
		return
	}
	if CoverageAnyEnabled() {
		ReportCoverage(mc.Tool, mc.StartTime)
	}
	fpSize := uint64(0)
	if mc.FPSet != nil {
		fpSize = mc.FPSet.Size()
	}
	if toolMode() {
		PrintMessage(ECTLCProgressStatsDFID, fmtInt64(mc.StatesGenerated), fmtUint64(fpSize))
	}
	PrintMessage(ECTLCStatsDFID, fmtInt64(mc.StatesGenerated), fmtUint64(fpSize))
}

func (mc *DFIDModelChecker) Cleanup(success bool, cleanup bool) error {
	_ = success
	if mc == nil {
		return nil
	}
	if mc.cleanupDone {
		return nil
	}
	mc.cleanupDone = true
	var err error
	if mc.FPSet != nil {
		mc.FPSet.Close()
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		if closeErr := mc.LiveCheck.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	if mc.AllStateWriter != nil {
		if closeErr := mc.AllStateWriter.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	if cleanup {
		deleteDirLikeJava(mc.Metadir, success)
	}
	return err
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
	return mc.doNextIntoWithTool(mc.Tool, cur, cfp, isLeaf, states, fps)
}

func (mc *DFIDModelChecker) replayDFIDNextErrorCallStack() int {
	if mc == nil || mc.Tool == nil || mc.PredErrState == nil {
		if mc != nil && mc.ErrorCode != NoError {
			return mc.ErrorCode
		}
		return ECGeneral
	}
	callStackTool := NewCallStackTool(mc.Tool)
	_, result, err := mc.doNextIntoWithTool(callStackTool, mc.PredErrState, mc.PredErrState.FingerPrint(), true, NewStateVec(1), NewLongVecWithCapacity(1))
	if err != nil || result != NoError {
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
		if result != NoError {
			return result
		}
		return ECTLCNestedExpression
	}
	return NoError
}

func (mc *DFIDModelChecker) doNextIntoWithTool(tool *Tool, cur *TLCStateMut, cfp uint64, isLeaf bool, states *StateVec, fps *LongVec) (bool, int, error) {
	if tool == nil {
		err := newTLCError(ECGeneral, "DFID model checker has no tool")
		return true, ECGeneral, err
	}
	deadlocked := true
	allSuccDone := true
	allSuccNonLeaf := true
	var liveNextStates *SetOfStates
	if mc.CheckLiveness && isLeaf {
		liveNextStates = NewSetOfStates(dfidInitialSetOfStatesCapacity)
	}
	restoreCurrentState := PushCurrentState(cur)
	defer restoreCurrentState()
	for _, action := range tool.GetActions() {
		nextStates, err := tool.GetNextStates(action, cur)
		if err != nil {
			return allSuccNonLeaf, mc.dfidNextFailed(cur, nil, err), err
		}
		size := 0
		if nextStates != nil {
			size = nextStates.Size()
		}
		mc.StatesGenerated += int64(size)
		deadlocked = deadlocked && size == 0
		for i := 0; i < size; i++ {
			succ := nextStates.At(i)
			if !tool.IsGoodState(succ) {
				if mc.SetErrState(cur, succ, false, ECTLCStateNotCompletelySpecifiedNext) {
					mc.printTrace(ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(tool, action, succ), cur, succ)
				}
				return allSuccNonLeaf, ECTLCStateNotCompletelySpecifiedNext, nil
			}
			inModel, err := tool.IsInModel(succ)
			if err != nil {
				return allSuccNonLeaf, mc.dfidNextFailed(cur, succ, err), err
			}
			if inModel {
				inActions, err := tool.IsInActions(cur, succ)
				if err != nil {
					return allSuccNonLeaf, mc.dfidNextFailed(cur, succ, err), err
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
					if err := mc.AllStateWriter.WriteTransition(cur, succ, writeStatus, nil); err != nil {
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
				invariantViolated := false
				invariantNames := tool.GetInvNames()
				for k, invariant := range tool.GetInvariants() {
					valid, err := tool.IsValidState(invariant, succ)
					if err != nil {
						if mc.SetErrState(cur, succ, true, ECTLCInvariantEvaluationFailed) {
							mc.printTrace(ECTLCInvariantEvaluationFailed, []string{nameAt(invariantNames, k)}, cur, succ)
						}
						return allSuccNonLeaf, ECTLCInvariantEvaluationFailed, err
					}
					if !valid {
						if continuationEnabled() {
							mc.printTrace(ECTLCInvariantViolatedBehavior, []string{nameAt(invariantNames, k)}, cur, succ)
							invariantViolated = true
							break
						}
						if mc.SetErrState(cur, succ, false, ECTLCInvariantViolatedBehavior) {
							mc.printTrace(ECTLCInvariantViolatedBehavior, []string{nameAt(invariantNames, k)}, cur, succ)
						}
						return allSuccNonLeaf, ECTLCInvariantViolatedBehavior, nil
					}
				}
				if invariantViolated {
					continue
				}
			}
			impliedViolated := false
			impliedNames := tool.GetImpliedActNames()
			for k, implied := range tool.GetImpliedActions() {
				valid, err := tool.IsValidTransition(implied, cur, succ)
				if err != nil {
					if mc.SetErrState(cur, succ, true, ECTLCActionPropertyEvaluationFailed) {
						mc.printTrace(ECTLCActionPropertyEvaluationFailed, []string{nameAt(impliedNames, k)}, cur, succ)
					}
					return allSuccNonLeaf, ECTLCActionPropertyEvaluationFailed, err
				}
				if !valid {
					if continuationEnabled() {
						mc.printTrace(ECTLCActionPropertyViolatedBehavior, []string{nameAt(impliedNames, k)}, cur, succ)
						impliedViolated = true
						break
					}
					if mc.SetErrState(cur, succ, false, ECTLCActionPropertyViolatedBehavior) {
						mc.printTrace(ECTLCActionPropertyViolatedBehavior, []string{nameAt(impliedNames, k)}, cur, succ)
					}
					return allSuccNonLeaf, ECTLCActionPropertyViolatedBehavior, nil
				}
			}
			if impliedViolated {
				continue
			}
		}
	}
	if deadlocked && mc.CheckDeadlock {
		if mc.SetErrState(cur, nil, false, ECTLCDeadlockReached) {
			mc.printTrace(ECTLCDeadlockReached, nil, cur, nil)
		}
		return allSuccNonLeaf, ECTLCDeadlockReached, nil
	}
	if liveNextStates != nil {
		liveNextStates.PutFP(cfp, cur)
		if mc.AllStateWriter != nil {
			if err := mc.AllStateWriter.WriteTransitionVisual(cur, cur, StateVisitUnseen, nil, StateVisualizationStuttering); err != nil {
				return allSuccNonLeaf, ECGeneral, err
			}
		}
		if mc.LiveCheck != nil {
			if err := mc.LiveCheck.AddNextState(tool, cur, cfp, liveNextStates); err != nil {
				return allSuccNonLeaf, mc.dfidNextFailed(cur, nil, err), err
			}
		}
	}
	if allSuccDone || (isLeaf && allSuccNonLeaf) {
		mc.FPSet.SetStatus(cfp, FPIntStatusDone)
	}
	return allSuccNonLeaf, NoError, nil
}

func (mc *DFIDModelChecker) dfidNextFailed(curState *TLCStateMut, succState *TLCStateMut, err error) int {
	if err == nil {
		return NoError
	}
	ec, params, keepCallStack := dfidNextFailureMessage(err)
	if mc.SetErrState(curState, succState, keepCallStack, ec) {
		mc.printTrace(ec, params, curState, succState)
	}
	return ec
}

func dfidNextFailureMessage(err error) (int, []string, bool) {
	var tlcErr *TLCError
	if errors.As(err, &tlcErr) && tlcErr != nil && (tlcErr.Code == ECSystemStackOverflow || tlcErr.Code == ECSystemOutOfMemory) {
		return tlcErr.Code, nil, false
	}
	var evalErr *EvalException
	if errors.As(err, &evalErr) && evalErr != nil && (evalErr.GetErrorCode() == ECSystemStackOverflow || evalErr.GetErrorCode() == ECSystemOutOfMemory) {
		return evalErr.GetErrorCode(), nil, false
	}
	return ECGeneral, []string{javaGeneralErrorMessage("computing the set of next states", err)}, true
}

func javaGeneralErrorMessage(cause string, err error) string {
	msg := "TLC threw an unexpected exception."
	msg += "\nThis was probably caused by an error in the spec or model."
	if cause == "" {
		msg += "\nSee the User Output or TLC Console for clues to what happened."
	} else {
		msg += "\nThe error occurred when TLC was " + cause + "."
	}
	if err == nil {
		msg += "\nThe exception was a <nil>\n"
		return msg
	}
	msg += fmt.Sprintf("\nThe exception was a %T\n", err)
	if err.Error() != "" {
		msg += ": " + err.Error()
	}
	return msg
}

func (mc *DFIDModelChecker) printTrace(errorCode int, params []string, curState *TLCStateMut, succState *TLCStateMut) {
	PrintError(errorCode, params...)
	traceEnd := succState
	if traceEnd == nil {
		traceEnd = curState
	}
	trace := traceFromState(traceEnd)
	if len(trace) == 0 {
		return
	}
	PrintError(ECTLCBehaviorUpToThisPoint)
	for i, info := range trace {
		var previous *TLCStateMut
		if i > 0 && trace[i-1] != nil {
			previous = trace[i-1].OriginalState()
		}
		PrintInvariantViolationStateTraceState(info, previous, i+1, i == len(trace)-1)
	}
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
	return mc.doInitWithTool(mc.Tool)
}

func (mc *DFIDModelChecker) doInitWithTool(tool *Tool) (int, error) {
	vec := NewStateVec(0)
	if tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "DFID model checker has no tool")
	}
	if err := tool.GetInitStates(NewStateFunctor(vec.AddElement)); err != nil {
		return ECGeneral, err
	}
	mc.StatesGenerated = int64(vec.Size())
	mc.InitStates = make([]*TLCStateMut, 0, vec.Size())
	mc.InitFPs = make([]uint64, 0, vec.Size())
	for i := 0; i < vec.Size(); i++ {
		state := vec.At(i)
		if !tool.IsGoodState(state) {
			if mc.SetErrState(state, nil, false, ECTLCStateNotCompletelySpecifiedInitial) {
				PrintError(ECTLCStateNotCompletelySpecifiedInitial, state.String())
			}
			return ECTLCStateNotCompletelySpecifiedInitial, nil
		}
		status := FPIntStatusNew
		inModel, err := tool.IsInModel(state)
		if err != nil {
			mc.ErrState = state
			return ECGeneral, err
		}
		if inModel {
			fp := state.FingerPrint()
			status = mc.FPSet.SetStatus(fp, FPIntStatusNew)
			if status == FPIntStatusNew {
				mc.InitStates = append(mc.InitStates, state)
				mc.InitFPs = append(mc.InitFPs, fp)
				if mc.AllStateWriter != nil {
					if err := mc.AllStateWriter.WriteInitState(state); err != nil {
						return ECGeneral, err
					}
				}
				if mc.CheckLiveness && mc.LiveCheck != nil {
					if err := mc.LiveCheck.AddInitState(tool, state, fp); err != nil {
						mc.ErrState = state
						return ECGeneral, err
					}
				}
			}
		}
		if status != FPIntStatusNew {
			continue
		}
		for k, invariant := range tool.GetInvariants() {
			valid, err := tool.IsValidState(invariant, state)
			if err != nil {
				mc.ErrState = state
				return ECTLCInvariantEvaluationFailed, err
			}
			if !valid {
				alias := state
				if tool != nil {
					alias = tool.EvalAlias(state, state)
				}
				PrintError(ECTLCInvariantViolatedInitial, nameAt(tool.GetInvNames(), k), alias.String())
				if continuationEnabled() {
					continue
				}
				mc.SetErrState(state, nil, false, ECTLCInvariantViolatedInitial)
				return ECTLCInvariantViolatedInitial, nil
			}
		}
		for k, implied := range tool.GetImpliedInits() {
			valid, err := tool.IsValidState(implied, state)
			if err != nil {
				mc.ErrState = state
				return ECTLCPropertyViolatedInitial, err
			}
			if !valid {
				alias := state
				if tool != nil {
					alias = tool.EvalAlias(state, state)
				}
				PrintError(ECTLCPropertyViolatedInitial, nameAt(tool.GetImpliedInitNames(), k), alias.String())
				mc.SetErrState(state, nil, false, ECTLCPropertyViolatedInitial)
				return ECTLCPropertyViolatedInitial, nil
			}
		}
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) reportInitException(result int, err error) int {
	if result == NoError {
		result = initExceptionCode(err)
	}
	message := ""
	if err != nil {
		message = err.Error()
	}
	if message == "" {
		message = fmt.Sprintf("%T", err)
	}
	if mc.ErrState != nil {
		PrintError(ECTLCInitialState, message, mc.ErrState.String())
	} else {
		PrintError(ECGeneral, javaGeneralErrorMessage("computing initial states", err))
	}
	if replayResult := mc.replayInitErrorCallStack(result); replayResult != NoError {
		result = replayResult
	}
	return result
}

func (mc *DFIDModelChecker) replayInitErrorCallStack(fallback int) int {
	if mc == nil || mc.Tool == nil {
		return fallback
	}
	mc.StatesGenerated = 0
	callStackTool := NewCallStackTool(mc.Tool)
	if _, err := mc.doInitWithTool(callStackTool); err != nil {
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
		return ECTLCNestedExpression
	}
	return fallback
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
	mc.Metadir = mc.FromCheckpoint
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
