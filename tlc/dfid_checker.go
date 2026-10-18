package tlc

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type DFIDModelChecker struct {
	*AbstractChecker
	InitStates      []*TLCStateMut
	InitFPs         []uint64
	FPSet           *MemFPIntSet
	DFIDWorkers     []*DFIDWorker
	dfidWorkerDone  []chan struct{}
	coordinatorWake chan struct{} // guarded by AbstractChecker.mu
	StatesGenerated int64
	CleanupEnabled  bool

	statesGeneratedOwner    *atomic.Int64
	liveCapacityMultipliers sync.Map
}

func (mc *DFIDModelChecker) require() {
	if mc == nil {
		panic(NewNullPointerException())
	}
}

func (mc *DFIDModelChecker) generatedStates() int64 {
	mc.require()
	if mc.statesGeneratedOwner != nil {
		return mc.statesGeneratedOwner.Load()
	}
	return mc.StatesGenerated
}

func (mc *DFIDModelChecker) setGeneratedStates(value int64) {
	mc.require()
	if mc.statesGeneratedOwner != nil {
		mc.statesGeneratedOwner.Store(value)
	}
	mc.StatesGenerated = value
}

func (mc *DFIDModelChecker) addGeneratedStates(delta int64) int64 {
	mc.require()
	if mc.statesGeneratedOwner != nil {
		value := mc.statesGeneratedOwner.Add(delta)
		mc.StatesGenerated = value
		return value
	}
	mc.StatesGenerated += delta
	return mc.StatesGenerated
}

func (mc *DFIDModelChecker) requireFPSet() *MemFPIntSet {
	mc.require()
	if mc.FPSet == nil {
		panic(NewNullPointerException())
	}
	return mc.FPSet
}

func (mc *DFIDModelChecker) requireStateWriter() IStateWriter {
	mc.require()
	if isNil(mc.ownedAllStateWriter()) {
		panic(NewNullPointerException())
	}
	return mc.ownedAllStateWriter()
}

func (mc *DFIDModelChecker) requireLiveCheck() *LiveCheck {
	mc.require()
	if mc.ownedLiveCheck() == nil {
		panic(NewNullPointerException())
	}
	return mc.ownedLiveCheck()
}

const dfidInitialSetOfStatesCapacity = 16

func (mc *DFIDModelChecker) liveCapacityMultiplier() int32 {
	if value, ok := mc.liveCapacityMultipliers.Load(currentGoroutineID()); ok {
		return value.(int32)
	}
	return 1
}

type DFIDModelCheckerOption func(*DFIDModelChecker)

func WithDFIDFromCheckpoint(fromCheckpoint *string) DFIDModelCheckerOption {
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

func WithDFIDStartTime(startTime time.Time) DFIDModelCheckerOption {
	return func(mc *DFIDModelChecker) {
		if mc.AbstractChecker != nil {
			mc.StartTime = startTime
		}
	}
}

func WithDFIDStateWriter(writer IStateWriter) DFIDModelCheckerOption {
	return func(mc *DFIDModelChecker) {
		if isNil(writer) {
			writer = NewNoopStateWriter()
		}
		mc.AllStateWriter = writer
	}
}

func WithDFIDCleanup(cleanup bool) DFIDModelCheckerOption {
	return func(mc *DFIDModelChecker) {
		mc.CleanupEnabled = cleanup
	}
}

func NewDFIDModelChecker(tool *Tool, metadir string, deadlock bool, opts ...DFIDModelCheckerOption) *DFIDModelChecker {
	initializeAbstractCheckerStatics()
	if tool == nil {
		panic(NewNullPointerException())
	}
	SetTLCStateTool(tool)
	rootName := tool.GetRootName()
	mc := &DFIDModelChecker{
		AbstractChecker: newAbstractCheckerFields(tool, metadir, NewNoopStateWriter(), deadlock, nil, time.Now()),
		CleanupEnabled:  true,
	}
	for _, opt := range opts {
		opt(mc)
	}
	mc.AbstractChecker.dfidWorkerSlots = &mc.DFIDWorkers
	mc.AbstractChecker.initialize(mc.Stop)
	// Java applies these assertions after the complete parent constructor and
	// before allocating the fingerprint set or worker array.
	if NumWorkers() != 1 {
		panic(NewTLCRuntimeException(ECGeneral, "Depth-First Iterative Deepening mode does not support multiple workers (https://github.com/tlaplus/tlaplus/issues/548).  Please run TLC with a single worker."))
	}
	if mc.ownedCheckLiveness() {
		panic(NewTLCRuntimeException(ECGeneral, "Depth-First Iterative Deepening mode does not support checking liveness properties (https://github.com/tlaplus/tlaplus/issues/548).  Please check liveness properties in Breadth-First-Search mode."))
	}

	if mc.FPSet == nil {
		mc.FPSet = NewMemFPIntSet()
	}
	if mc.FPSet.metadir == "" || mc.FPSet.filename == "" {
		mc.FPSet.Init(NumWorkers(), metadir, rootName)
	}
	mc.DFIDWorkers = make([]*DFIDWorker, NumWorkers())
	mc.statesGeneratedOwner = &atomic.Int64{}
	return mc
}

// ModelCheck preserves AbstractChecker.modelCheck's final wrapper for DFID.
func (mc *DFIDModelChecker) ModelCheck() (result int, err error) {
	mc.require()
	result, err = mc.ModelCheckImpl()
	if result == NoError && mc.ErrorCode != NoError {
		result = mc.ErrorCode
	}
	return result, err
}

// ModelCheckImpl is the protected DFIDModelChecker.modelCheckImpl boundary.
func (mc *DFIDModelChecker) ModelCheckImpl() (result int, err error) {
	mc.require()
	result = NoError
	recovered, err := mc.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	proceed := false
	result, err, proceed = mc.processInitialStates()
	if !proceed {
		return result, err
	}
	if recovered {
		PrintMessage(ECTLCInitGenerated3, fmtInt64(mc.generatedStates()), fmtInt(len(mc.InitStates)))
	} else {
		PrintMessage(ECTLCInitGenerated4, fmtInt64(mc.generatedStates()), fmtInt(len(mc.InitStates)))
	}
	if len(mc.ownedTool().GetActions()) == 0 {
		ReportSuccessCountsDistance(mc.requireFPSet().Size(), mc.requireFPSet().CheckFPs(), mc.generatedStates())
		mc.PrintSummary(true)
		return NoError, mc.Cleanup(true, mc.CleanupEnabled)
	}
	result = ECGeneral
	defer func() {
		mc.PrintSummary(result == NoError)
		if cleanupErr := mc.Cleanup(result == NoError, mc.CleanupEnabled); cleanupErr != nil {
			err = cleanupErr
		}
	}()
	defer func() {
		if recovered := recover(); recovered != nil {
			result, err = dfidSearchResult(result, panicValueAsError(recovered))
		}
	}()
	terminated := false
	for level := 2; level <= Globals.DFIDMax; level++ {
		if terminated {
			return mc.finishTerminatedDFID()
		}
		PrintMessage(ECTLCProgressStartStatsDFID, fmtInt(level), fmtInt64(mc.generatedStates()), fmtInt64(int64(mc.requireFPSet().Size())))
		IncLevel()
		runResult, runErr := mc.RunTLC(level)
		if runErr != nil {
			return dfidSearchResult(result, runErr)
		}
		result = runResult
		// Java resets done under the checker monitor after all workers join.
		mc.mu.Lock()
		mc.Done.Store(false)
		mc.mu.Unlock()
		if result != NoError {
			return result, nil
		}
		terminated = mc.shouldTerminateAfterLevel()
	}
	return result, nil
}

func (mc *DFIDModelChecker) shouldTerminateAfterLevel() bool {
	mc.require()
	terminated := false
	for i := 0; i < len(mc.DFIDWorkers); i++ {
		if mc.dfidWorkerAt(i).IsTerminated() {
			terminated = true
			break
		}
	}
	moreLevel := false
	for i := 0; i < len(mc.DFIDWorkers); i++ {
		if mc.dfidWorkerAt(i).HasMoreLevel() {
			moreLevel = true
			break
		}
	}
	return terminated || !moreLevel
}

// processInitialStates is DFIDModelChecker.modelCheckImpl's first
// try/catch(Throwable), before the separate search try/finally.
func (mc *DFIDModelChecker) processInitialStates() (result int, err error, proceed bool) {
	mc.require()
	result = NoError
	defer func() {
		if recovered := recover(); recovered != nil {
			result, err = mc.finishInitException(panicValueAsError(recovered))
			proceed = false
		}
	}()
	if mc.ownedCheckLiveness() && mc.requireLiveCheck().GetNumChecker() == 0 {
		PrintError(ECTLCLiveFormulaTautology)
		return ECTLCLiveFormulaTautology, nil, false
	}
	if result := mc.CheckAssumptions(); result != NoError {
		return result, nil, false
	}
	result, err = mc.DoInit(false)
	if err != nil {
		result, err = mc.finishInitException(err)
		return result, err, false
	}
	if result != NoError {
		return result, nil, false
	}
	return NoError, nil, true
}

func (mc *DFIDModelChecker) finishInitException(failure error) (int, error) {
	result := mc.reportInitException(failure)
	mc.PrintSummary(false)
	// Java's initialization catch(Throwable) consumes the failure after
	// diagnostics and replay; only cleanup failures escape this boundary.
	return result, mc.Cleanup(false, mc.CleanupEnabled)
}

// dfidSearchResult applies DFIDModelChecker.modelCheckImpl's outer
// catch(Exception), which covers both runTLC and the terminal liveness check.
func dfidSearchResult(result int, err error) (int, error) {
	if err != nil && !isJavaError(err) {
		switch failure := err.(type) {
		case *LiveException:
			result = failure.ownedErrorCode()
		case *LiveCounterExampleException:
			result = failure.ownedLiveException().ownedErrorCode()
		default:
			result = PrintError(ECGeneral, generalErrorParams("", err)...)
		}
		return result, nil
	}
	return result, err
}

// CheckAssumptions is DFIDModelChecker's own assumption loop. Unlike
// ModelChecker, the source calls the Tool's isValid overload directly.
func (mc *DFIDModelChecker) CheckAssumptions() int {
	mc.require()
	tool := mc.ownedTool()
	if tool == nil {
		panic(NewNullPointerException())
	}
	assumptions := tool.GetAssumptions()
	semanticGraphArrayLength(tool.ownedSpecProcessor() != nil, assumptions)
	axioms := tool.GetAssumptionIsAxiom()
	for i, assumption := range assumptions {
		valid, err := checkPropertyPredicate(func() (bool, error) {
			if axioms == nil {
				panic(NewNullPointerException())
			}
			if i >= len(axioms) {
				panic(NewArrayIndexOutOfBoundsException(i, len(axioms)))
			}
			if axioms[i] {
				return true, nil
			}
			return tool.IsValidExpr(assumption)
		})
		if err != nil {
			if semanticExploreNull(assumption) {
				panic(NewNullPointerException())
			}
			return PrintErrorNullable(ECTLCAssumptionEvaluationError, javaString(SemanticString(assumption)), javaThrowableDetailMessage(err))
		}
		if !valid {
			if semanticExploreNull(assumption) {
				panic(NewNullPointerException())
			}
			return PrintError(ECTLCAssumptionFalse, SemanticString(assumption))
		}
	}
	return NoError
}

func (mc *DFIDModelChecker) RunTLC(depth int) (result int, failure error) {
	mc.require()
	depth = int(int32(depth))
	if depth < 2 {
		return NoError, nil
	}
	mc.StartWorkers(mc.AbstractChecker, depth)
	wakeup := mc.beginCoordinatorWait()
	defer mc.endCoordinatorWait()
	defer func() {
		if recovered := recover(); recovered != nil {
			mc.SetStop(2)
			mc.joinDFIDWorkers()
			result, failure = ECGeneral, panicValueAsError(recovered)
		}
	}()
	interval := ProgressInterval()
	coverageCountdown := periodicCoverageCountdown(interval)
	if !mc.isDFIDDone() {
		select {
		case <-wakeup:
		case <-time.After(3 * time.Second):
		}
	}
	for {
		result, err := mc.DoPeriodicWork()
		if err != nil || result != NoError {
			return result, err
		}
		if !mc.reportDFIDProgressWhileActive(coverageCountdown, depth) {
			mc.joinDFIDWorkers()
			return mc.dfidWorkerResult()
		}
		if interval < 0 {
			panic(NewIllegalArgumentException("timeout value is negative"))
		}
		timer := time.NewTimer(interval)
		select {
		case <-wakeup:
		case <-timer.C:
		}
		timer.Stop()
		completed := func() bool {
			mc.mu.Lock()
			defer mc.mu.Unlock()
			if coverageCountdown == 0 {
				coverageCountdown = periodicCoverageCountdown(interval)
			} else {
				coverageCountdown = int(int32(coverageCountdown) - 1)
			}
			return mc.isDFIDDone()
		}()
		if completed {
			mc.joinDFIDWorkers()
			return mc.dfidWorkerResult()
		}
	}
}

// Java holds the checker monitor from its done check through progress and
// coverage reporting. Keep that decision atomic with worker completion.
func (mc *DFIDModelChecker) reportDFIDProgressWhileActive(count int, depth int) bool {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.Done.Load() {
		return false
	}
	mc.RunTLCContinueDoing(count, depth)
	return true
}

func (mc *DFIDModelChecker) SetDone() bool {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	old := mc.Done.Load()
	mc.Done.Store(true)
	mc.notifyCoordinatorLocked()
	return old
}

func (mc *DFIDModelChecker) notifyCoordinatorLocked() {
	if mc.coordinatorWake != nil {
		select {
		case mc.coordinatorWake <- struct{}{}:
		default:
		}
	}
}

func (mc *DFIDModelChecker) beginCoordinatorWait() <-chan struct{} {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.coordinatorWake = make(chan struct{}, 1)
	if mc.Done.Load() {
		mc.coordinatorWake <- struct{}{}
	}
	return mc.coordinatorWake
}

func (mc *DFIDModelChecker) endCoordinatorWait() {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	mc.coordinatorWake = nil
	mc.mu.Unlock()
}

func (mc *DFIDModelChecker) joinDFIDWorkers() {
	mc.require()
	for i := 0; i < len(mc.DFIDWorkers); i++ {
		mc.dfidWorkerAt(i)
		if i >= len(mc.dfidWorkerDone) {
			panic(NewArrayIndexOutOfBoundsException(i, len(mc.dfidWorkerDone)))
		}
		if mc.dfidWorkerDone[i] == nil {
			panic(NewNullPointerException())
		}
		<-mc.dfidWorkerDone[i]
	}
}

func (mc *DFIDModelChecker) dfidWorkerAt(index int) *DFIDWorker {
	mc.require()
	if index < 0 || index >= len(mc.DFIDWorkers) {
		panic(NewArrayIndexOutOfBoundsException(index, len(mc.DFIDWorkers)))
	}
	worker := mc.DFIDWorkers[index]
	if worker == nil {
		panic(NewNullPointerException())
	}
	return worker
}

func (mc *DFIDModelChecker) dfidWorkerResult() (int, error) {
	mc.require()
	// AbstractChecker.runTLC returns the checker's accepted error code after
	// joining workers. A caught worker failure can be rejected by setErrState
	// once done is already true; local worker bookkeeping is not that result.
	if !mc.KeepCallStack {
		return mc.ErrorCode, nil
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) isDFIDDone() bool {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	return mc.Done.Load()
}

// StartWorkers constructs and publishes all depth-first workers before any of
// them begin running, matching DFIDModelChecker.startWorkers in Java.
func (mc *DFIDModelChecker) StartWorkers(_ *AbstractChecker, checkIndex int) []*DFIDWorker {
	mc.require()
	done := make([]chan struct{}, len(mc.DFIDWorkers))
	for i := range mc.DFIDWorkers {
		worker := NewDFIDWorker(i, checkIndex, mc)
		mc.mu.Lock()
		mc.DFIDWorkers[i] = worker
		mc.mu.Unlock()
		done[i] = make(chan struct{})
	}
	mc.dfidWorkerDone = done
	for i, worker := range mc.DFIDWorkers {
		workerDone := done[i]
		go func(worker *DFIDWorker, workerDone chan struct{}) {
			defer close(workerDone)
			restoreWorker := pushCurrentCheckerWorker(worker.ownedID(), worker)
			defer restoreWorker()
			worker.Run()
		}(worker, workerDone)
	}
	return mc.DFIDWorkers
}

// RunTLCContinueDoing reports depth-first progress and periodic coverage. The
// native coordinator performs Java's monitor wait after this method returns.
func (mc *DFIDModelChecker) RunTLCContinueDoing(count int, depth int) {
	mc.require()
	count = int(int32(count))
	depth = int(int32(depth))
	_ = depth
	PrintMessage(ECTLCProgressStatsDFID, fmtInt64(mc.generatedStates()), fmtInt64(int64(mc.requireFPSet().Size())))
	if count == 0 {
		if mc.AbstractChecker == nil {
			panic(NewNullPointerException())
		}
		mc.AbstractChecker.ReportCoverage(mc.DFIDWorkers)
	}
}

func (mc *DFIDModelChecker) finishTerminatedDFID() (int, error) {
	mc.require()
	result := NoError
	if mc.ErrState == nil {
		if mc.ownedCheckLiveness() {
			PrintMessage(ECTLCProgressStatsDFID, fmtInt64(mc.generatedStates()), fmtInt64(int64(mc.requireFPSet().Size())))
			liveResult, liveErr := mc.requireLiveCheck().FinalCheck(mc.ownedTool())
			if liveErr != nil {
				return dfidSearchResult(result, liveErr)
			}
			result = liveResult
			if result != NoError {
				return result, nil
			}
		}
		ReportSuccessCountsDistance(mc.requireFPSet().Size(), mc.requireFPSet().CheckFPs(), mc.generatedStates())
		return NoError, nil
	}
	if mc.KeepCallStack {
		// Replay adds diagnostics without replacing the iteration result.
		// ModelCheck applies the retained checker error after cleanup.
		mc.replayDFIDNextErrorCallStack()
	} else if mc.ErrorCode != NoError {
		result = mc.ErrorCode
	} else {
		result = ECGeneral
	}
	return result, nil
}

func (mc *DFIDModelChecker) PrintSummary(success bool) {
	_ = success
	mc.require()
	mc.AbstractChecker.ReportCoverage(mc.DFIDWorkers)
	if toolMode() {
		PrintMessage(ECTLCProgressStatsDFID, fmtInt64(mc.generatedStates()), fmtInt64(int64(mc.requireFPSet().Size())))
	}
	PrintMessage(ECTLCStatsDFID, fmtInt64(mc.generatedStates()), fmtInt64(int64(mc.requireFPSet().Size())))
}

func (mc *DFIDModelChecker) Cleanup(success bool, cleanupPolicy ...bool) error {
	mc.require()
	if len(cleanupPolicy) > 1 {
		panic(NewIllegalArgumentException("Cleanup accepts at most one cleanup policy"))
	}
	cleanup := true
	if len(cleanupPolicy) == 1 {
		cleanup = cleanupPolicy[0]
	}
	mc.requireFPSet().Close()
	if mc.ownedCheckLiveness() {
		if closeErr := mc.requireLiveCheck().Close(); closeErr != nil {
			return closeErr
		}
	}
	if closeErr := mc.requireStateWriter().Close(); closeErr != nil {
		return closeErr
	}
	if cleanup {
		deleteDirLikeJava(mc.ownedMetadir(), success)
	}
	return nil
}

func (mc *DFIDModelChecker) SetErrState(curState *TLCStateMut, succState *TLCStateMut, keepCallStack bool, errorCode int) bool {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	ok := mc.AbstractChecker.setErrStateLocked(curState, succState, keepCallStack, errorCode)
	if ok {
		mc.setStopLocked(2)
	}
	return ok
}

func (mc *DFIDModelChecker) SetStop(code int) {
	mc.require()
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.setStopLocked(code)
}

func (mc *DFIDModelChecker) setStopLocked(code int) {
	mc.require()
	if mc.DFIDWorkers == nil {
		panic(NewNullPointerException())
	}
	for _, worker := range mc.DFIDWorkers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		worker.SetStop(code)
	}
}

func (mc *DFIDModelChecker) DoNext(cur *TLCStateMut, cfp uint64, isLeaf bool, states *StateVec, fps *LongVec) (bool, int, error) {
	mc.require()
	return mc.DoNextWithTool(mc.ownedTool(), cur, cfp, isLeaf, states, fps)
}

// DoNextInto retains the earlier native spelling.
func (mc *DFIDModelChecker) DoNextInto(cur *TLCStateMut, cfp uint64, isLeaf bool, states *StateVec, fps *LongVec) (bool, int, error) {
	mc.require()
	return mc.DoNext(cur, cfp, isLeaf, states, fps)
}

func (mc *DFIDModelChecker) replayDFIDNextErrorCallStack() {
	mc.require()
	callStackTool := NewCallStackTool(mc.ownedTool())
	// Java catches Throwable around the complete replay call, including its
	// argument evaluation and failures raised while reporting the next error.
	failed := func() (failed bool) {
		defer func() {
			if failure := recover(); failure != nil {
				if _, ok := failure.(error); !ok {
					panic(failure)
				}
				failed = true
			}
		}()
		_, _, err := mc.DoNextWithTool(callStackTool, mc.PredErrState, mc.PredErrState.FingerPrint(), true, NewStateVec(1), NewLongVecWithCapacity(1))
		return err != nil
	}()
	if failed {
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
	}
}

// DoNextWithTool is the Tool-taking source overload of DoNext. Go uses a
// distinct name because methods cannot be overloaded by parameter list.
func (mc *DFIDModelChecker) DoNextWithTool(tool *Tool, cur *TLCStateMut, cfp uint64, isLeaf bool, states *StateVec, fps *LongVec) (allSuccNonLeaf bool, result int, failure error) {
	mc.require()
	deadlocked := true
	allSuccDone := true
	allSuccNonLeaf = true
	var liveNextStates *SetOfStates
	if mc.ownedCheckLiveness() && isLeaf {
		liveNextStates = NewSetOfStates(int(int32(dfidInitialSetOfStatesCapacity) * mc.liveCapacityMultiplier()))
	}
	restoreCurrentState := PushCurrentState(cur)
	defer restoreCurrentState()
	var currentSuccessor *TLCStateMut
	defer func() {
		if recovered := recover(); recovered != nil {
			failure = panicValueAsError(recovered)
			result = mc.dfidNextFailed(cur, currentSuccessor, failure)
		}
	}()
	propertyIndex := 0
	for _, action := range tool.GetActions() {
		nextStates, err := tool.GetNextStates(action, cur)
		if err != nil {
			return allSuccNonLeaf, mc.dfidNextFailed(cur, nil, err), err
		}
		size := nextStates.Size()
		mc.addGeneratedStates(int64(size))
		deadlocked = deadlocked && size == 0
		for i := 0; i < size; i++ {
			succ := nextStates.At(i)
			currentSuccessor = succ
			if !tool.IsGoodState(succ) {
				mc.reportNextError(cur, succ, false, ECTLCStateNotCompletelySpecifiedNext, func() []string {
					return incompleteNextStateParams(tool, action, succ)
				})
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
				status = mc.requireFPSet().SetStatus(fp, FPIntStatusNew)
				allSuccDone = allSuccDone && IsDone(status)
				allSuccNonLeaf = allSuccNonLeaf && !IsLeaf(status)
				writeStatus := StateVisitSeen
				if status == FPIntStatusNew {
					writeStatus = StateVisitUnseen
				}
				if err := mc.requireStateWriter().WriteTransition(cur, succ, writeStatus, nil); err != nil {
					return allSuccNonLeaf, mc.dfidNextFailed(cur, succ, err), err
				}
				if !IsCompleted(status) {
					states.Add(succ)
					fps.AddElement(int64(fp))
				}
				if liveNextStates != nil {
					liveNextStates.PutFP(fp, succ)
				}
			}
			if status == FPIntStatusNew {
				violated, code := mc.checkDFIDProperty(tool, cur, succ, &propertyIndex, false)
				if code != NoError {
					return allSuccNonLeaf, code, nil
				}
				if violated {
					continue
				}
			}
			violated, code := mc.checkDFIDProperty(tool, cur, succ, &propertyIndex, true)
			if code != NoError {
				return allSuccNonLeaf, code, nil
			}
			if violated {
				continue
			}

		}
		currentSuccessor = nil
	}
	if deadlocked && mc.ownedCheckDeadlock() {
		mc.reportNextError(cur, nil, false, ECTLCDeadlockReached, nil)
		return allSuccNonLeaf, ECTLCDeadlockReached, nil
	}
	if liveNextStates != nil {
		curFP := cur.FingerPrint()
		liveNextStates.PutFP(curFP, cur)
		if err := mc.requireStateWriter().WriteTransitionVisual(cur, cur, StateVisitUnseen, nil, StateVisualizationStuttering); err != nil {
			return allSuccNonLeaf, mc.dfidNextFailed(cur, nil, err), err
		}
		if err := mc.requireLiveCheck().AddNextState(tool, cur, curFP, liveNextStates); err != nil {
			return allSuccNonLeaf, mc.dfidNextFailed(cur, nil, err), err
		}
		multiplier := mc.liveCapacityMultiplier()
		if liveNextStates.Capacity() > int(int32(dfidInitialSetOfStatesCapacity)*multiplier) {
			mc.liveCapacityMultipliers.Store(currentGoroutineID(), multiplier+1)
		}
	}
	if allSuccDone || (isLeaf && allSuccNonLeaf) {
		mc.requireFPSet().SetStatus(cfp, FPIntStatusDone)
	}
	return allSuccNonLeaf, NoError, nil
}

// DoNextIntoWithTool retains the earlier native spelling.
func (mc *DFIDModelChecker) DoNextIntoWithTool(tool *Tool, cur *TLCStateMut, cfp uint64, isLeaf bool, states *StateVec, fps *LongVec) (bool, int, error) {
	mc.require()
	return mc.DoNextWithTool(tool, cur, cfp, isLeaf, states, fps)
}

// DFID captures each property-array length, but reads current elements and
// diagnostic names. The index is shared across both families and successors.
func (mc *DFIDModelChecker) checkDFIDProperty(tool *Tool, cur, succ *TLCStateMut, index *int, implied bool) (violated bool, result int) {
	mc.require()
	violationCode, evaluationCode := ECTLCInvariantViolatedBehavior, ECTLCInvariantEvaluationFailed
	if implied {
		violationCode, evaluationCode = ECTLCActionPropertyViolatedBehavior, ECTLCActionPropertyEvaluationFailed
	}
	failEvaluation := func() int {
		mc.reportNextError(cur, succ, true, evaluationCode, func() []string {
			var names []string
			if implied {
				names = tool.GetImpliedActNames()
			} else {
				names = tool.GetInvNames()
			}
			return []string{tool.propertyNameAt(names, *index)}
		})
		return evaluationCode
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err, ok := recovered.(error)
			if !ok || isJavaError(err) {
				panic(recovered)
			}
			violated, result = false, failEvaluation()
		}
	}()
	var actions []*Action
	if implied {
		actions = tool.GetImpliedActions()
	} else {
		actions = tool.GetInvariants()
	}
	length := len(tool.requireActionArray(actions))
	for *index = 0; *index < length; (*index)++ {
		if implied {
			actions = tool.GetImpliedActions()
		} else {
			actions = tool.GetInvariants()
		}
		action := tool.propertyActionAt(actions, *index)
		var valid bool
		var err error
		if implied {
			valid, err = tool.IsValidTransition(action, cur, succ)
		} else {
			valid, err = tool.IsValidState(action, succ)
		}
		if err != nil {
			// Returned evaluator errors enter the same source catch as array exceptions.
			panic(err)
		}
		if !valid {
			return mc.reportPropertyViolation(tool, cur, succ, *index, implied, violationCode)
		}
	}
	return false, NoError
}

func (mc *DFIDModelChecker) reportNextError(cur, succ *TLCStateMut, keep bool, code int, params func() []string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.SetErrState(cur, succ, keep, code) {
		var values []string
		if params != nil {
			values = params()
		}
		mc.printTrace(code, values, cur, succ)
		mc.notifyCoordinatorLocked()
	}
}

func (mc *DFIDModelChecker) reportPropertyViolation(tool *Tool, cur, succ *TLCStateMut, index int, implied bool, code int) (bool, int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	params := func() []string {
		var names []string
		if implied {
			names = tool.GetImpliedActNames()
		} else {
			names = tool.GetInvNames()
		}
		return []string{tool.propertyNameAt(names, index)}
	}
	if continuationEnabled() {
		mc.printTrace(code, params(), cur, succ)
		return true, NoError
	}
	if mc.SetErrState(cur, succ, false, code) {
		mc.printTrace(code, params(), cur, succ)
		mc.notifyCoordinatorLocked()
	}
	return false, code
}

func (mc *DFIDModelChecker) dfidNextFailed(curState *TLCStateMut, succState *TLCStateMut, err error) int {
	mc.require()
	if err == nil {
		return NoError
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	ec := ECGeneral
	if code := javaSystemFailureCode(err); code == ECSystemStackOverflow || code == ECSystemOutOfMemory {
		ec = code
	}
	keepCallStack := ec == ECGeneral
	if mc.SetErrState(curState, succState, keepCallStack, ec) {
		var params []string
		if ec == ECGeneral {
			params = []string{javaGeneralErrorMessage("computing the set of next states", err)}
		}
		mc.printTrace(ec, params, curState, succState)
		mc.notifyCoordinatorLocked()
	}
	return ec
}

func javaGeneralErrorMessage(cause string, err error) string {
	initializeMPStatics()
	msg := "TLC threw an unexpected exception."
	msg += "\nThis was probably caused by an error in the spec or model."
	if cause == "" {
		msg += "\nSee the User Output or TLC Console for clues to what happened."
	} else {
		msg += "\nThe error occurred when TLC was " + cause + "."
	}
	if err == nil {
		panic(NewNullPointerException())
	}
	name := javaThrowableClassName(err)
	if javaRuntimeException(err) != nil {
		name = "java.lang.RuntimeException"
	}
	msg += "\nThe exception was a " + name + "\n"
	message := javaThrowableDetailMessage(err)
	if message != nil {
		msg += ": " + *message
	}
	// Java MP's historical noDebug property enables this general-error stack
	// suffix independently of TLCGlobals.debug. Null messages always include it.
	if message == nil || mpGeneralDebugEnabled() {
		msg += javaThrowableStackTrace(err)
	}
	return msg
}

func (mc *DFIDModelChecker) printTrace(errorCode int, params []string, curState *TLCStateMut, succState *TLCStateMut) {
	mc.require()
	errorCode = int(int32(errorCode))
	workerID, ok := CurrentWorkerID()
	if !ok {
		panic(NewClassCastException("class java.lang.Thread cannot be cast to class tlc2.util.IdThread (java.lang.Thread is in module java.base of loader 'bootstrap'; tlc2.util.IdThread is in unnamed module of loader 'app')"))
	}
	if mc.DFIDWorkers == nil {
		panic(NewNullPointerException())
	}
	if workerID < 0 || workerID >= len(mc.DFIDWorkers) {
		panic(NewArrayIndexOutOfBoundsException(workerID, len(mc.DFIDWorkers)))
	}
	worker := mc.DFIDWorkers[workerID]
	if worker == nil {
		panic(NewNullPointerException())
	}
	if errorCode == ECTLCInvariantViolatedBehavior {
		worker.PrintInvariantTrace(errorCode, params, curState, succState)
	} else {
		worker.PrintErrorTrace(errorCode, params, curState, succState)
	}
}

func (mc *DFIDModelChecker) DoPeriodicWork() (int, error) {
	mc.require()
	fpSet := mc.requireFPSet()
	fpSet.mu.Lock()
	defer fpSet.mu.Unlock()
	return mc.doPeriodicWorkLocked()
}

func (mc *DFIDModelChecker) doPeriodicWorkLocked() (int, error) {
	mc.require()
	if mc.ownedCheckLiveness() {
		result, err := mc.requireLiveCheck().Check(mc.ownedTool(), false)
		if err != nil || result != NoError {
			return result, err
		}
	}
	if DoCheckPoint() {
		if err := mc.checkpointLocked(); err != nil {
			return ECSystemCheckpointRecoveryCorrupt, err
		}
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) DoInit(ignoreCancel bool) (int, error) {
	mc.require()
	_ = ignoreCancel
	return mc.doInitWithTool(mc.ownedTool())
}

func (mc *DFIDModelChecker) doInitWithTool(tool *Tool) (result int, failure error) {
	mc.require()
	var currentState *TLCStateMut
	defer func() {
		if recovered := recover(); recovered != nil {
			result, failure = ECGeneral, panicValueAsError(recovered)
		}
		if failure != nil {
			if javaSystemFailureCode(failure) == ECSystemOutOfMemory {
				result, failure = PrintError(ECSystemOutOfMemoryTooManyInit), nil
				return
			}
			mc.ErrState = currentState
		}
	}()
	vec := NewStateVec(0)
	if tool == nil {
		panic(NewNullPointerException())
	}
	if err := tool.GetInitStates(NewStateFunctor(vec.AddElement)); err != nil {
		return ECGeneral, err
	}
	mc.setGeneratedStates(int64(vec.Size()))
	// Keep the source preallocated arrays intact on early failure; trim on success.
	mc.InitStates = make([]*TLCStateMut, vec.Size())
	mc.InitFPs = make([]uint64, vec.Size())
	initIndex := 0
	for i := 0; i < vec.Size(); i++ {
		state := vec.At(i)
		currentState = state
		if !tool.IsGoodState(state) {
			PrintError(ECTLCStateNotCompletelySpecifiedInitial, state.String())
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
			status = mc.requireFPSet().SetStatus(fp, FPIntStatusNew)
			if status == FPIntStatusNew {
				mc.InitStates[initIndex] = state
				mc.InitFPs[initIndex] = fp
				initIndex++
				if err := mc.requireStateWriter().WriteInitState(state); err != nil {
					return ECGeneral, err
				}
				if mc.ownedCheckLiveness() {
					if err := mc.requireLiveCheck().AddInitState(tool, state, fp); err != nil {
						mc.ErrState = state
						return ECGeneral, err
					}
				}
			}
		}
		if status != FPIntStatusNew {
			continue
		}
		for k := 0; k < len(tool.requireActionArray(tool.GetInvariants())); k++ {
			invariant := tool.propertyActionAt(tool.GetInvariants(), k)
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
				PrintError(ECTLCInvariantViolatedInitial, tool.propertyNameAt(tool.GetInvNames(), k), alias.String())
				if continuationEnabled() {
					continue
				}
				return ECTLCInvariantViolatedInitial, nil
			}
		}
		for k := 0; k < len(tool.requireActionArray(tool.GetImpliedInits())); k++ {
			implied := tool.propertyActionAt(tool.GetImpliedInits(), k)
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
				PrintError(ECTLCPropertyViolatedInitial, tool.propertyNameAt(tool.GetImpliedInitNames(), k), alias.String())
				return ECTLCPropertyViolatedInitial, nil
			}
		}
	}
	if initIndex < vec.Size() {
		states := make([]*TLCStateMut, initIndex)
		copy(states, mc.InitStates)
		fps := make([]uint64, initIndex)
		copy(fps, mc.InitFPs)
		mc.InitStates, mc.InitFPs = states, fps
	}
	return NoError, nil
}

func (mc *DFIDModelChecker) reportInitException(err error) int {
	mc.require()
	if mc.ErrState != nil {
		PrintErrorNullable(ECTLCInitialState, javaThrowableDetailMessage(err), javaString(mc.ErrState.String()))
	} else {
		PrintError(ECGeneral, javaGeneralErrorMessage("computing initial states", err))
	}
	// The source result assignment never completes when doInit throws. It
	// remains NO_ERROR unless replay itself throws a nested-expression error.
	return mc.replayInitErrorCallStack(NoError)
}

func (mc *DFIDModelChecker) replayInitErrorCallStack(fallback int) int {
	mc.require()
	mc.setGeneratedStates(0)
	callStackTool := NewCallStackTool(mc.ownedTool())
	if _, err := mc.doInitWithTool(callStackTool); err != nil {
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
		return ECTLCNestedExpression
	}
	return fallback
}

func (mc *DFIDModelChecker) Checkpoint() error {
	mc.require()
	fpSet := mc.requireFPSet()
	fpSet.mu.Lock()
	defer fpSet.mu.Unlock()
	return mc.checkpointLocked()
}

func (mc *DFIDModelChecker) checkpointLocked() error {
	mc.require()
	PrintMessage(ECTLCCheckpointStart, mc.ownedMetadir())
	if err := mc.requireFPSet().beginChkptLocked(); err != nil {
		return err
	}
	if mc.ownedCheckLiveness() {
		if err := mc.requireLiveCheck().BeginChkpt(); err != nil {
			return err
		}
	}
	if err := BeginChkptUniqueStrings(mc.ownedMetadir()); err != nil {
		return err
	}
	if err := mc.requireFPSet().CommitChkpt(); err != nil {
		return err
	}
	if mc.ownedCheckLiveness() {
		if err := mc.requireLiveCheck().CommitChkpt(); err != nil {
			return err
		}
	}
	if err := CommitChkptUniqueStrings(mc.ownedMetadir()); err != nil {
		return err
	}
	PrintMessage(ECTLCCheckpointEnd)
	return nil
}

func (mc *DFIDModelChecker) Recover() (bool, error) {
	mc.require()
	if mc.ownedFromCheckpoint() == nil {
		return false, nil
	}
	PrintMessage(ECTLCCheckpointRecoverStart, *mc.ownedFromCheckpoint())
	if err := mc.requireFPSet().Recover(); err != nil {
		return false, err
	}
	if mc.ownedCheckLiveness() {
		if err := mc.requireLiveCheck().Recover(); err != nil {
			return false, err
		}
	}
	size := mc.requireFPSet().Size()
	PrintMessage(ECTLCCheckpointRecoverEndDFID, fmtInt64(int64(size)))
	mc.setGeneratedStates(int64(mc.requireFPSet().Size()))
	return true, nil
}

func fmtInt(value int) string {
	return fmt.Sprintf("%d", value)
}

func fmtInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}
