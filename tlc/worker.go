package tlc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

var workerCoverageStatics struct {
	once     sync.Once
	action   bool
	variable bool
}

// InitializeWorkerStatics represents a fresh Java classloader for Worker's
// coverage flags. All workers in one runtime share the first captured values.
func InitializeWorkerStatics() {
	workerCoverageStatics.once = sync.Once{}
	workerCoverageStatics.action = false
	workerCoverageStatics.variable = false
}

func initializeWorkerStatics() {
	workerCoverageStatics.once.Do(func() {
		workerCoverageStatics.action = CoverageActionEnabled()
		workerCoverageStatics.variable = CoverageVariableEnabled()
	})
}

type Worker struct {
	mu                    sync.Mutex
	registerMu            sync.Mutex
	ID                    int
	LocalValues           []Value
	NamedRegisters        *NamedRegisterMap
	SetOfStates           *SetOfStates
	SetOfStatesMultiplier int32
	OutDegree             *FixedSizedBucketStatistics
	statesGenerated       atomic.Int64
	Checker               *ModelChecker
	Tool                  *Tool
	Halted                bool
	maxLevel              atomic.Int32
	started               atomic.Bool
	UnseenSuccessorStates int
	done                  chan struct{}
	Err                   error
	traceFileBase         string
	traceRAF              *BufferedRandomAccessFile
	lastPtr               int64
	traceErr              error
	DisableTraceMirror    bool
	checkDeadlock         bool
	checkLiveness         bool
	mode                  ToolMode
	stateQueue            StateQueue
	fingerprintSet        FPSet
	stateWriter           IStateWriter
	actionCoverage        bool
	variableCoverage      bool
	identityCaptured      bool
	idOwner               int
	namedRegistersOwner   *NamedRegisterMap
	runSettingsCaptured   bool
	checkerOwner          *ModelChecker
	toolOwner             *Tool
	outDegreeOwner        *FixedSizedBucketStatistics
}

type WorkerWrappingRuntimeException struct {
	throwableTrace
	Err   error
	State *TLCStateMut

	ownersCaptured bool
	errOwner       error
	stateOwner     *TLCStateMut
}

func NewWorkerWrappingRuntimeException(err error, state *TLCStateMut) error {
	return &WorkerWrappingRuntimeException{
		throwableTrace: captureThrowableTrace(), Err: err, State: state,
		ownersCaptured: true, errOwner: err, stateOwner: state,
	}
}

func (e *WorkerWrappingRuntimeException) Error() string { return javaThrowableClassName(e) }

func (e *WorkerWrappingRuntimeException) GetMessage() *string {
	e.require()
	return nil
}

func (e *WorkerWrappingRuntimeException) GetCause() error {
	e.require()
	return nil
}

func (e *WorkerWrappingRuntimeException) UnwrapState() *TLCStateMut {
	e.require()
	if !e.ownersCaptured {
		return e.State
	}
	return e.stateOwner
}

func (e *WorkerWrappingRuntimeException) UnwrapExp() error {
	e.require()
	if !e.ownersCaptured {
		return e.Err
	}
	return e.errOwner
}

func (e *WorkerWrappingRuntimeException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func NewWorker(id int) *Worker {
	id = int(int32(id))
	initializeWorkerStatics()
	worker := &Worker{
		ID:                    id,
		LocalValues:           make([]Value, 4),
		NamedRegisters:        NewNamedRegisterMap(),
		SetOfStatesMultiplier: 1,
		OutDegree:             NewFixedSizedBucketStatistics(fmt.Sprintf("TLC Worker %d", id), workerOutDegreeBucketCount),
		actionCoverage:        workerCoverageStatics.action,
		variableCoverage:      workerCoverageStatics.variable,
	}
	worker.idOwner = worker.ID
	worker.namedRegistersOwner = worker.NamedRegisters
	worker.identityCaptured = true
	return worker
}

func (w *Worker) ownedID() int {
	if w.identityCaptured {
		return w.idOwner
	}
	return w.ID
}

func (w *Worker) ownedNamedRegisters() *NamedRegisterMap {
	if w.identityCaptured {
		return w.namedRegistersOwner
	}
	return w.NamedRegisters
}

// Java's maxLevel is volatile: reporting reads it while this worker runs.
func (w *Worker) GetMaxLevel() int {
	w.require()
	return int(w.maxLevel.Load())
}

func (w *Worker) SetLevel(level int) {
	w.require()
	w.maxLevel.Store(int32(level))
}

func NewModelCheckingWorker(id int, checker *ModelChecker, tool *Tool) *Worker {
	id = int(int32(id))
	if checker == nil || checker.AbstractChecker == nil || tool == nil {
		panic(NewNullPointerException())
	}
	worker := NewWorker(id)
	worker.Checker = checker
	worker.Tool = tool
	worker.captureRunSettings()
	worker.configureTrace()
	if err := worker.ensureTraceRAF(); err != nil {
		panic(err)
	}
	registered := false
	defer func() {
		if !registered {
			// A rejected native registration must not leak its newly opened file.
			_ = worker.CloseTrace()
		}
	}()
	checker.ownedConcurrentTrace().AddWorker(worker)
	for len(checker.Workers) <= id {
		checker.Workers = append(checker.Workers, nil)
	}
	checker.Workers[id] = worker
	registered = true
	return worker
}

func (w *Worker) captureRunSettings() {
	w.checkerOwner = w.Checker
	w.toolOwner = w.Tool
	w.outDegreeOwner = w.OutDegree
	w.checkDeadlock = w.checkerOwner.ownedCheckDeadlock()
	w.checkLiveness = w.checkerOwner.ownedCheckLiveness()
	w.mode = w.toolOwner.GetMode()
	w.stateQueue = w.checkerOwner.StateQueue
	w.fingerprintSet = w.checkerOwner.FPSet
	w.stateWriter = w.checkerOwner.ownedAllStateWriter()
	w.runSettingsCaptured = true
}

func (w *Worker) MyGetID() int {
	w.require()
	return w.ownedID()
}

func (w *Worker) Start() {
	w.require()
	if !w.started.CompareAndSwap(false, true) {
		panic(NewIllegalThreadStateException())
	}
	w.Halted = false
	w.Err = nil
	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		restoreWorker := pushCurrentCheckerWorker(w.ownedID(), w)
		defer restoreWorker()
		w.Err = w.Run()
	}()
}

func (w *Worker) Join() error {
	w.require()
	if w.done == nil {
		return nil
	}
	<-w.done
	return w.Err
}

func (w *Worker) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *Worker) requireChecker() *ModelChecker {
	w.require()
	checker := w.ownedChecker()
	if checker == nil {
		panic(NewNullPointerException())
	}
	return checker
}

func (w *Worker) requireTool() *Tool {
	w.require()
	tool := w.ownedWorkerTool()
	if tool == nil {
		panic(NewNullPointerException())
	}
	return tool
}

func (w *Worker) ownedChecker() *ModelChecker {
	if w.runSettingsCaptured {
		return w.checkerOwner
	}
	return w.Checker
}

func (w *Worker) ownedWorkerTool() *Tool {
	if w.runSettingsCaptured {
		return w.toolOwner
	}
	return w.Tool
}

func (w *Worker) ownedOutDegree() *FixedSizedBucketStatistics {
	if w.runSettingsCaptured {
		return w.outDegreeOwner
	}
	return w.OutDegree
}

func (w *Worker) requireStateQueue() StateQueue {
	w.require()
	if isNil(w.stateQueue) {
		panic(NewNullPointerException())
	}
	return w.stateQueue
}

func (w *Worker) Run() (err error) {
	w.require()
	checker := w.requireChecker()
	w.requireTool()
	stateQueue := w.requireStateQueue()
	var curState *TLCStateMut
	defer func() {
		if recovered := recover(); recovered != nil {
			err = panicValueAsError(recovered)
			checker.mu.Lock()
			defer checker.mu.Unlock()
			if checker.SetErrState(curState, nil, true, ECGeneral) {
				PrintError(ECGeneral, generalErrorParams("", err)...)
			}
			stateQueue.FinishAll()
			checker.notifyCoordinatorLocked()
		}
	}()
	for {
		curState = stateQueue.SDequeue()
		if curState == nil {
			checker.mu.Lock()
			checker.SetDone()
			checker.notifyCoordinatorLocked()
			checker.mu.Unlock()
			stateQueue.FinishAll()
			return nil
		}
		stop, runErr := w.DoNext(curState)
		if runErr != nil {
			stateQueue.FinishAll()
			return runErr
		}
		if stop {
			stateQueue.FinishAll()
			return nil
		}
	}
}

func (w *Worker) NextStateFunctor() *NextStateFunctor {
	w.require()
	return &NextStateFunctor{
		StateFunctor: StateFunctor{
			AddElementFunc: w.AddElement,
			GetStatesFunc:  w.GetStates,
		},
		AddNextElementFunc:         w.AddNextElement,
		IncrementStatesGeneratedFn: w.IncrementStatesGenerated,
		ShouldHaltFunc:             w.Halt,
		AddUnsatisfiedNextStateFn:  w.AddUnsatisfiedState,
	}
}

func (w *Worker) AddElement(_ *TLCStateMut) (any, error) {
	w.require()
	return nil, NewWrongInvocationException("tlc2.tool.Worker.addElement(TLCState) should not be called")
}

func (w *Worker) DoNext(curState *TLCStateMut) (bool, error) {
	w.require()
	tool := w.requireTool()
	checker := w.requireChecker()
	if w.checkLiveness || w.mode == ModeDebugger {
		w.SetOfStates = w.CreateSetOfStates()
	}
	restoreWorkerID := pushUnownedWorkerID(w.ownedID())
	defer restoreWorkerID()
	restoreCurrentState := PushCurrentState(curState)
	defer restoreCurrentState()
	restoreRandomState := PushRandomEnumerableState(curState)
	defer restoreRandomState()
	preNext := w.GetStatesGenerated()
	recordedOutcome := false
	// Worker.run has an inner catch(Throwable) around getNextStates, separate
	// from its outer loop catch. Preserve the throwable before doNextFailed.
	halt, err := func() (halt bool, err error) {
		defer func() {
			if failure := recover(); failure != nil {
				halt, err = false, panicValueAsError(failure)
			}
		}()
		return tool.GetNextStatesWithFunctor(w.NextStateFunctor(), curState)
	}()
	if err != nil {
		if wrapped, ok := err.(*WorkerWrappingRuntimeException); ok && wrapped != nil {
			checker.doNextFailed(curState, wrapped.UnwrapState(), wrapped.UnwrapExp())
			recordedOutcome = true
		} else {
			checker.doNextFailed(curState, nil, err)
			recordedOutcome = true
		}
	}
	if (halt || w.Halted) && !recordedOutcome {
		if checker.ErrorCode != NoError {
			recordedOutcome = true
		} else {
			return true, nil
		}
	}
	// Java Worker.run catches expected next-state failures inside the
	// iteration, records them through doNextFailed/doNextSetErr, and still
	// executes the deadlock/liveness/out-degree tail before the next dequeue
	// observes finishAll().
	if w.checkDeadlock && preNext == w.GetStatesGenerated() {
		checker.doNextSetErrWithPostConditionTool(tool, curState, nil, false, ECTLCDeadlockReached, "")
		recordedOutcome = true
	}
	if w.checkLiveness {
		if err := w.CheckLiveness(curState); err != nil {
			if IsInvariantViolatedException(err) {
				w.requireStateQueue().FinishAll()
				return true, nil
			}
			checker.doNextFailed(curState, nil, err)
			recordedOutcome = true
		}
	}
	w.RecordOutDegree()
	return false, nil
}

func (w *Worker) CheckLiveness(curState *TLCStateMut) error {
	if w == nil || curState == nil {
		panic(NewNullPointerException())
	}
	curFP := curState.FingerPrint()
	if w.SetOfStates == nil {
		panic(NewNullPointerException())
	}
	workerTool := w.requireTool()
	checker := w.requireChecker()
	w.SetOfStates.PutFP(curFP, curState, workerTool)
	if isNil(checker.ownedAllStateWriter()) {
		panic(NewNullPointerException())
	}
	if err := checker.ownedAllStateWriter().WriteTransitionVisual(curState, curState, StateVisitUnseen, nil, StateVisualizationStuttering); err != nil {
		return err
	}
	if checker.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	tool := checker.ownedTool().NoDebug()
	if checker.ownedLiveCheck() == nil {
		panic(NewNullPointerException())
	}
	err := checker.ownedLiveCheck().AddNextState(tool, curState, curFP, w.SetOfStates)
	if err == nil {
		if w.SetOfStates.Capacity() > int(w.SetOfStatesMultiplier*workerSetOfStatesInitialCapacity) {
			w.SetOfStatesMultiplier++
		}
		return nil
	}
	if !livenessErrorNeedsCallStackReplay(err) {
		return err
	}
	// Java holds the ModelChecker monitor from the winning error-stack claim
	// through the complete call-stack replay.
	checker.mu.Lock()
	defer checker.mu.Unlock()
	if !w.claimLivenessErrorStackPrinterLocked() {
		return nil
	}
	w.SetOfStates.ResetNext()
	if checker.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	// NewCallStackTool selects noDebug once while copying the checker tool.
	callStackTool := NewCallStackTool(checker.ownedTool())
	if checker.ownedLiveCheck() == nil {
		panic(NewNullPointerException())
	}
	rerunErr := checker.ownedLiveCheck().AddNextState(callStackTool, curState, curFP, w.SetOfStates)
	if rerunErr == nil {
		return NewTLCRuntimeExceptionWithCause(ECGeneral, err)
	}
	if !livenessErrorNeedsCallStackReplay(rerunErr) {
		return rerunErr
	}
	if callStackTool.HasCallStack() {
		checker.mu.Lock()
		checker.KeepCallStack = false
		checker.mu.Unlock()
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
	}
	return err
}

func livenessErrorNeedsCallStackReplay(err error) bool {
	// Java catches EvalException and TLCRuntimeException at this boundary.
	// A cause of either type does not make its enclosing exception eligible.
	return isJavaEvalOrRuntimeException(err)
}

func (w *Worker) claimLivenessErrorStackPrinterLocked() bool {
	if w == nil || w.ownedChecker() == nil {
		return false
	}
	checker := w.ownedChecker()
	checker.mu.Lock()
	defer checker.mu.Unlock()
	if checker.PrintedLivenessErrorStack {
		return false
	}
	checker.PrintedLivenessErrorStack = true
	return true
}

func (w *Worker) AddNextElement(curState *TLCStateMut, action *Action, succState *TLCStateMut) (result any, err error) {
	w.require()
	if w.actionCoverage {
		if action == nil {
			panic(NewNullPointerException())
		}
		action.CM.IncInvocations()
	}
	w.statesGenerated.Add(1)
	// Worker.addElement catches Exception, wraps its cause with the successor,
	// and leaves Java Error subclasses for run's surrounding throwable catch.
	defer func() {
		if failure := recover(); failure != nil {
			cause := panicValueAsError(failure)
			if isJavaError(cause) {
				panic(failure)
			}
			result, err = nil, NewWorkerWrappingRuntimeException(cause, succState)
		}
	}()
	checker := w.requireChecker()
	// Java Worker.addElement throws INextStateFunctor.InvariantViolatedException
	// after doNextSetErr records stop-worthy successor failures.
	if w.Halted {
		return nil, NewWorkerWrappingRuntimeException(NewInvariantViolatedException(), succState)
	}
	stop, queued, err := checker.processSuccessorForWorker(w, curState, succState, action, w.SetOfStates)
	if stop || err != nil {
		w.Halted = true
	}
	if err != nil {
		return nil, NewWorkerWrappingRuntimeException(err, succState)
	}
	if stop {
		return nil, NewWorkerWrappingRuntimeException(NewInvariantViolatedException(), succState)
	}
	if queued && succState != nil {
		if succState.Level() > w.GetMaxLevel() {
			w.SetLevel(succState.Level())
		}
	}
	return w, nil
}

func (w *Worker) AddUnsatisfiedState(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	w.require()
	if isNil(w.stateWriter) {
		panic(NewNullPointerException())
	}
	if w.stateWriter.IsConstrained() {
		if err := w.stateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action, pred); err != nil {
			panic(err)
		}
	}
	return succState
}

// AddUnsatisfiedNextState retains the earlier native spelling.
func (w *Worker) AddUnsatisfiedNextState(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	return w.AddUnsatisfiedState(curState, action, succState, pred, con)
}

func (w *Worker) IncrementStatesGenerated(count int64) {
	w.require()
	w.statesGenerated.Add(count)
}

func (w *Worker) Halt() bool {
	w.require()
	return w.Halted
}

func (w *Worker) GetStatesGenerated() int64 {
	w.require()
	return w.statesGenerated.Load()
}

func (w *Worker) GetOutDegree() *FixedSizedBucketStatistics {
	w.require()
	return w.ownedOutDegree()
}

const workerSetOfStatesInitialCapacity = 16
const workerOutDegreeBucketCount = 32

func (w *Worker) CreateSetOfStates() *SetOfStates {
	w.require()
	return NewSetOfStates(int(w.SetOfStatesMultiplier * workerSetOfStatesInitialCapacity))
}

func (w *Worker) HasStates() bool {
	w.require()
	return w.SetOfStates != nil
}

func (w *Worker) GetStates() *SetOfStates {
	w.require()
	return w.SetOfStates
}

func (w *Worker) RecordOutDegree() {
	w.require()
	outDegree := w.ownedOutDegree()
	if outDegree == nil {
		panic(NewNullPointerException())
	}
	outDegree.AddSample(w.UnseenSuccessorStates)
	w.UnseenSuccessorStates = 0
}

const tlcTraceExt = ".st"

func (w *Worker) configureTrace() {
	if w == nil || w.ownedChecker() == nil || w.ownedChecker().AbstractChecker == nil || w.ownedWorkerTool() == nil {
		panic(NewNullPointerException())
	}
	metadir := w.ownedChecker().ownedMetadir()
	rootName := w.ownedWorkerTool().GetRootName()
	w.traceFileBase = workerTraceFileBase(metadir, rootName, w.ownedID())
}

func workerTraceFileBase(metadir string, rootName string, id int) string {
	// Worker retains FileUtil separator concatenation, including dot components.
	return metadir + string(os.PathSeparator) + fmt.Sprintf("%s-%d", rootName, id)
}

func (w *Worker) SetTraceContext(metadir string, rootName string) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	base := workerTraceFileBase(metadir, rootName, w.ownedID())
	if base == w.traceFileBase {
		return
	}
	if w.traceRAF != nil {
		if err := w.traceRAF.Close(); err != nil {
			w.traceErr = err
			w.traceRAF = nil
			w.traceFileBase = base
			return
		}
		w.traceRAF = nil
	}
	w.traceErr = nil
	w.traceFileBase = base
}

func (w *Worker) ensureTraceRAF() error {
	if w == nil {
		return newTLCError(ECGeneral, "worker is nil")
	}
	if w.traceErr != nil {
		return w.traceErr
	}
	if w.traceRAF != nil {
		return nil
	}
	if w.traceFileBase == "" {
		w.configureTrace()
	}
	raf, err := NewBufferedRandomAccessFile(w.traceFileBase+tlcTraceExt, "rw")
	if err != nil {
		w.traceErr = err
		return err
	}
	w.traceRAF = raf
	return nil
}

func (w *Worker) WriteState(state *TLCStateMut, fp uint64, successors ...*TLCStateMut) error {
	if w == nil {
		panic(NewNullPointerException())
	}
	if len(successors) == 0 {
		return w.writeInitState(state, fp)
	}
	if len(successors) == 1 {
		return w.writeNextStateWithAction(state, fp, successors[0], nil)
	}
	panic(NewIllegalArgumentException("WriteState accepts at most one successor"))
}

func (w *Worker) WriteInitState(initialState *TLCStateMut, fp uint64) error {
	return w.WriteState(initialState, fp)
}

func (w *Worker) writeInitState(initialState *TLCStateMut, fp uint64) error {
	if w == nil {
		panic(NewNullPointerException())
	}
	if err := w.writeInitStateRecord(initialState, fp); err != nil {
		return err
	}
	// The native mirror takes the shared trace monitor. Release the worker's
	// file monitor first: reconstruction holds the trace monitor while reading
	// worker records, so nesting these locks in reverse order can deadlock.
	checker := w.ownedChecker()
	if !w.DisableTraceMirror && checker != nil && checker.ownedTrace() != nil {
		checker.ownedTrace().MirrorInitStateForWorker(w.ownedID(), initialState, fp, initialState.UID)
	}
	return nil
}

func (w *Worker) writeInitStateRecord(initialState *TLCStateMut, fp uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	ptr, err := w.traceRAF.GetFilePointer()
	if err != nil {
		return err
	}
	w.lastPtr = ptr
	if err := w.traceRAF.WriteLongNat(1); err != nil {
		return err
	}
	if err := w.traceRAF.WriteShortNat(w.ownedID()); err != nil {
		return err
	}
	if err := w.traceRAF.WriteLong(int64(fp)); err != nil {
		return err
	}
	if initialState == nil {
		panic(NewNullPointerException())
	}
	initialState.WorkerID = int16(w.ownedID())
	initialState.UID = ptr
	return nil
}

func (w *Worker) WriteNextState(curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action) error {
	return w.writeNextStateWithAction(curState, succFP, succState, action)
}

func (w *Worker) writeNextStateWithAction(curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action) error {
	if w == nil {
		panic(NewNullPointerException())
	}
	if err := w.writeNextStateRecord(curState, succFP, succState); err != nil {
		return err
	}
	checker := w.ownedChecker()
	if !w.DisableTraceMirror && checker != nil && checker.ownedTrace() != nil {
		checker.ownedTrace().MirrorNextStateForWorker(w.ownedID(), curState, succFP, succState, action, succState.UID)
	}
	return nil
}

func (w *Worker) writeNextStateRecord(curState *TLCStateMut, succFP uint64, succState *TLCStateMut) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if curState == nil {
		panic(NewNullPointerException())
	}
	if level := int(int32(curState.Level()) + int32(1)); level > w.GetMaxLevel() {
		w.SetLevel(level)
	}
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	prevUID := curState.UID
	prevWorker := curState.WorkerID
	ptr, err := w.traceRAF.GetFilePointer()
	if err != nil {
		return err
	}
	w.lastPtr = ptr
	if err := w.traceRAF.WriteLongNat(prevUID); err != nil {
		return err
	}
	if err := w.traceRAF.WriteShortNat(int(prevWorker)); err != nil {
		return err
	}
	if err := w.traceRAF.WriteLong(int64(succFP)); err != nil {
		return err
	}
	if succState == nil {
		panic(NewNullPointerException())
	}
	succState.WorkerID = int16(w.ownedID())
	succState.UID = ptr
	succState.SetPredecessor(curState)
	w.UnseenSuccessorStates++
	return nil
}

func (w *Worker) ReadStateRecord(ptr int64) (ConcurrentTraceRecord, error) {
	w.require()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	w.traceRAF.Mark()
	if err := w.traceRAF.Seek(ptr); err != nil {
		return ConcurrentTraceRecord{}, err
	}
	prev, err := w.traceRAF.ReadLongNat()
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	if ptr < 0 {
		panic(NewAssertionError())
	}
	workerID, err := w.traceRAF.ReadShortNat()
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	checker := w.ownedChecker()
	if checker == nil || checker.AbstractChecker == nil || checker.Workers == nil {
		panic(NewNullPointerException())
	}
	if workerID < 0 || workerID >= len(checker.Workers) {
		panic(NewAssertionError())
	}
	fp, err := w.traceRAF.ReadLong()
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	if checker.FPSet == nil {
		panic(NewNullPointerException())
	}
	if !checker.FPSet.Contains(uint64(fp)) {
		panic(NewAssertionError())
	}
	if err := w.traceRAF.Seek(w.traceRAF.GetMark()); err != nil {
		return ConcurrentTraceRecord{}, err
	}
	return NewConcurrentTraceRecord(prev, workerID, uint64(fp)), nil
}

func (w *Worker) BeginChkpt() error {
	w.require()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	if err := w.traceRAF.Flush(); err != nil {
		return err
	}
	file, err := openFileOutputStream(w.traceFileBase + ".tmp")
	if err != nil {
		return err
	}
	filePtr, err := w.traceRAF.GetFilePointer()
	if err != nil {
		return err
	}
	if err := binary.Write(file, binary.BigEndian, filePtr); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	if err := binary.Write(file, binary.BigEndian, w.lastPtr); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	return bufferedRandomAccessFileIOError(file.Close())
}

func (w *Worker) CommitChkpt() error {
	w.require()
	w.mu.Lock()
	defer w.mu.Unlock()
	oldChkpt := w.traceFileBase + ".chkpt"
	newChkpt := w.traceFileBase + ".tmp"
	if _, err := os.Stat(oldChkpt); err == nil {
		if err := os.Remove(oldChkpt); err != nil {
			return NewIOException(fmt.Sprintf("Trace.commitChkpt: cannot delete %s", oldChkpt))
		}
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException(fmt.Sprintf("Trace.commitChkpt: cannot delete %s", oldChkpt))
	}
	return nil
}

func (w *Worker) Recover() error {
	w.require()
	file, err := openFileInputStream(w.traceFileBase + ".chkpt")
	if err != nil {
		return err
	}
	var filePos int64
	if err := binary.Read(file, binary.BigEndian, &filePos); err != nil {
		return bufferedRandomAccessFileIOError(valueStreamReadError(err))
	}
	var lastPtr int64
	if err := binary.Read(file, binary.BigEndian, &lastPtr); err != nil {
		return bufferedRandomAccessFileIOError(valueStreamReadError(err))
	}
	w.lastPtr = lastPtr
	if err := file.Close(); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	return w.traceRAF.Seek(filePos)
}

// RecoverTrace retains the earlier native spelling.
func (w *Worker) RecoverTrace() error { return w.Recover() }

func (w *Worker) CloseTrace() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.traceRAF == nil {
		return nil
	}
	return w.traceRAF.Close()
}

func (w *Worker) DeleteTrace() error {
	if w == nil || w.traceFileBase == "" {
		return nil
	}
	var err error
	for _, suffix := range []string{tlcTraceExt, ".tmp", ".chkpt"} {
		if removeErr := os.Remove(w.traceFileBase + suffix); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) && err == nil {
			err = removeErr
		}
	}
	return err
}

func (w *Worker) Elements() (*WorkerTraceEnumerator, error) {
	if w == nil {
		panic(NewNullPointerException())
	}
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	length, err := w.traceRAF.GetFilePointer()
	if err != nil {
		return nil, err
	}
	raf, err := NewBufferedRandomAccessFile(w.traceFileBase+tlcTraceExt, "r")
	if err != nil {
		return nil, err
	}
	return &WorkerTraceEnumerator{length: length, raf: raf}, nil
}

type WorkerTraceEnumerator struct {
	length int64
	raf    *BufferedRandomAccessFile
}

func (e *WorkerTraceEnumerator) HasMoreFP() (bool, error) {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.raf == nil {
		return false, NewNullPointerException()
	}
	pos, err := e.raf.GetFilePointer()
	if err != nil {
		return false, err
	}
	return pos < e.length, nil
}

// HasMoreFPWithError retains the earlier native spelling.
func (e *WorkerTraceEnumerator) HasMoreFPWithError() (bool, error) {
	return e.HasMoreFP()
}

func (e *WorkerTraceEnumerator) NextFP() (uint64, error) {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.raf == nil {
		return 0, NewNullPointerException()
	}
	if _, err := e.raf.ReadLongNat(); err != nil {
		return 0, err
	}
	if _, err := e.raf.ReadShortNat(); err != nil {
		return 0, err
	}
	fp, err := e.raf.ReadLong()
	return uint64(fp), err
}

func (e *WorkerTraceEnumerator) Close() error {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.raf == nil {
		return NewNullPointerException()
	}
	return e.raf.Close()
}
