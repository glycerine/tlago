package tlc

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type Worker struct {
	mu                    sync.Mutex
	ID                    int
	LocalValues           []Value
	NamedRegisters        *InsMap[*UniqueString, Value]
	SetOfStates           *SetOfStates
	SetOfStatesMultiplier int
	OutDegree             *FixedSizedBucketStatistics
	statesGenerated       atomic.Int64
	Checker               *ModelChecker
	Tool                  *Tool
	Halted                bool
	maxLevel              atomic.Int32
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
}

type WorkerWrappingRuntimeException struct {
	throwableTrace
	Err   error
	State *TLCStateMut
}

func NewWorkerWrappingRuntimeException(err error, state *TLCStateMut) error {
	return &WorkerWrappingRuntimeException{throwableTrace: captureThrowableTrace(), Err: err, State: state}
}

func (e *WorkerWrappingRuntimeException) Error() string             { return javaThrowableClassName(e) }
func (e *WorkerWrappingRuntimeException) GetMessage() *string       { return nil }
func (e *WorkerWrappingRuntimeException) UnwrapState() *TLCStateMut { return e.State }
func (e *WorkerWrappingRuntimeException) UnwrapExp() error          { return e.Err }

func NewWorker(id int) *Worker {
	return &Worker{
		ID:                    id,
		NamedRegisters:        NewInsMap[*UniqueString, Value](),
		SetOfStatesMultiplier: 1,
		OutDegree:             NewFixedSizedBucketStatistics(fmt.Sprintf("TLCWorkerThread-%03d", id), workerOutDegreeBucketCount),
	}
}

// Java's maxLevel is volatile: reporting reads it while this worker runs.
func (w *Worker) GetMaxLevel() int {
	return int(w.maxLevel.Load())
}

func (w *Worker) SetLevel(level int) {
	w.maxLevel.Store(int32(level))
}

func NewModelCheckingWorker(id int, checker *ModelChecker, tool *Tool) *Worker {
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
	checker.ConcurrentTrace.AddWorker(worker)
	for len(checker.Workers) <= id {
		checker.Workers = append(checker.Workers, nil)
	}
	checker.Workers[id] = worker
	registered = true
	return worker
}

func (w *Worker) captureRunSettings() {
	w.checkDeadlock = w.Checker.CheckDeadlock
	w.checkLiveness = w.Checker.CheckLiveness
	w.mode = w.Tool.GetMode()
	w.stateQueue = w.Checker.StateQueue
}

func (w *Worker) MyGetID() int {
	if w == nil {
		return 0
	}
	return w.ID
}

func (w *Worker) Start() {
	if w == nil {
		return
	}
	if w.done != nil {
		select {
		case <-w.done:
		default:
			return
		}
	}
	w.Halted = false
	w.Err = nil
	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		w.Err = w.Run()
	}()
}

func (w *Worker) Join() error {
	if w == nil || w.done == nil {
		return nil
	}
	<-w.done
	return w.Err
}

func (w *Worker) Run() (err error) {
	if w == nil {
		return newTLCError(ECGeneral, "worker is nil")
	}
	var curState *TLCStateMut
	defer func() {
		if recovered := recover(); recovered != nil {
			err = panicValueAsError(recovered)
			if w.Checker != nil {
				w.Checker.nextErrorMu.Lock()
				defer w.Checker.nextErrorMu.Unlock()
				if w.Checker.SetErrState(curState, nil, true, ECGeneral) {
					PrintError(ECGeneral, generalErrorParams("", err)...)
				}
				if w.stateQueue != nil {
					w.stateQueue.FinishAll()
				}
			}
		}
	}()
	if w.Tool == nil {
		return newTLCError(ECGeneral, "worker has no tool")
	}
	if w.Checker == nil {
		return newTLCError(ECGeneral, "worker has no model checker")
	}
	if w.stateQueue == nil {
		return newTLCError(ECGeneral, "model checker has no state queue")
	}
	for {
		curState = w.stateQueue.SDequeue()
		if curState == nil {
			w.Checker.SetDone()
			w.stateQueue.FinishAll()
			return nil
		}
		stop, runErr := w.DoNext(curState)
		if runErr != nil {
			if w.stateQueue != nil {
				w.stateQueue.FinishAll()
			}
			return runErr
		}
		if stop {
			if w.stateQueue != nil {
				w.stateQueue.FinishAll()
			}
			return nil
		}
	}
}

func (w *Worker) NextStateFunctor() *NextStateFunctor {
	return &NextStateFunctor{
		StateFunctor: StateFunctor{
			HasStatesFunc: w.HasStates,
			GetStatesFunc: w.GetStates,
		},
		AddNextElementFunc:         w.AddNextElement,
		IncrementStatesGeneratedFn: w.IncrementStatesGenerated,
		ShouldHaltFunc:             w.Halt,
		AddUnsatisfiedNextStateFn:  w.AddUnsatisfiedNextState,
	}
}

func (w *Worker) DoNext(curState *TLCStateMut) (bool, error) {
	if w == nil {
		return true, newTLCError(ECGeneral, "worker is nil")
	}
	if w.Tool == nil {
		return true, newTLCError(ECGeneral, "worker has no tool")
	}
	if w.Checker == nil {
		return true, newTLCError(ECGeneral, "worker has no model checker")
	}
	if w.checkLiveness || w.mode == ModeDebugger {
		w.SetOfStates = w.CreateSetOfStates()
	}
	restoreWorkerID := PushCurrentWorkerID(w.ID)
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
		return w.Tool.GetNextStatesWithFunctor(w.NextStateFunctor(), curState)
	}()
	if err != nil {
		if wrapped, ok := err.(*WorkerWrappingRuntimeException); ok && wrapped != nil {
			w.Checker.doNextFailed(curState, wrapped.UnwrapState(), wrapped.UnwrapExp())
			recordedOutcome = true
		} else {
			w.Checker.doNextFailed(curState, nil, err)
			recordedOutcome = true
		}
	}
	if (halt || w.Halted) && !recordedOutcome {
		if w.Checker.ErrorCode != NoError {
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
		w.Checker.doNextSetErrWithPostCondition(curState, nil, false, ECTLCDeadlockReached, "")
		recordedOutcome = true
	}
	if w.checkLiveness {
		if err := w.CheckLiveness(curState); err != nil {
			if IsInvariantViolatedException(err) {
				if w.stateQueue != nil {
					w.stateQueue.FinishAll()
				}
				return true, nil
			}
			w.Checker.doNextFailed(curState, nil, err)
			recordedOutcome = true
		}
	}
	if w.SetOfStates != nil && w.SetOfStates.Capacity() > w.SetOfStatesMultiplier*workerSetOfStatesInitialCapacity {
		w.SetOfStatesMultiplier++
	}
	w.RecordOutDegree()
	return false, nil
}

func (w *Worker) CheckLiveness(curState *TLCStateMut) error {
	if w == nil || w.Checker == nil || w.Checker.LiveCheck == nil || curState == nil {
		return nil
	}
	if w.SetOfStates == nil {
		w.SetOfStates = w.CreateSetOfStates()
	}
	curFP := curState.FingerPrint()
	w.SetOfStates.PutFP(curFP, curState, w.Tool)
	if w.Checker.AllStateWriter != nil {
		if err := w.Checker.AllStateWriter.WriteTransitionVisual(curState, curState, StateVisitUnseen, nil, StateVisualizationStuttering); err != nil {
			return err
		}
	}
	err := w.Checker.LiveCheck.AddNextState(w.Tool.NoDebug(), curState, curFP, w.SetOfStates)
	if err == nil || !livenessErrorNeedsCallStackReplay(err) {
		return err
	}
	if !w.claimLivenessErrorStackPrinter() {
		return nil
	}
	w.SetOfStates.ResetNext()
	callStackTool := NewCallStackTool(w.Tool.NoDebug())
	rerunErr := w.Checker.LiveCheck.AddNextState(callStackTool, curState, curFP, w.SetOfStates)
	if rerunErr == nil {
		return newTLCError(ECGeneral, "%s", err.Error())
	}
	if !livenessErrorNeedsCallStackReplay(rerunErr) {
		return rerunErr
	}
	if callStackTool.HasCallStack() {
		w.Checker.mu.Lock()
		w.Checker.KeepCallStack = false
		w.Checker.mu.Unlock()
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
	}
	return err
}

func livenessErrorNeedsCallStackReplay(err error) bool {
	// Java catches EvalException and TLCRuntimeException at this boundary.
	// A cause of either type does not make its enclosing exception eligible.
	return isJavaEvalOrRuntimeException(err)
}

func (w *Worker) claimLivenessErrorStackPrinter() bool {
	if w == nil || w.Checker == nil {
		return false
	}
	w.Checker.mu.Lock()
	defer w.Checker.mu.Unlock()
	if w.Checker.PrintedLivenessErrorStack {
		return false
	}
	w.Checker.PrintedLivenessErrorStack = true
	return true
}

func (w *Worker) AddNextElement(curState *TLCStateMut, action *Action, succState *TLCStateMut) (result any, err error) {
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
	if w == nil {
		return nil, newTLCError(ECGeneral, "worker is nil")
	}
	if w.Checker == nil {
		return nil, newTLCError(ECGeneral, "worker has no model checker")
	}
	// Java Worker.addElement throws INextStateFunctor.InvariantViolatedException
	// after doNextSetErr records stop-worthy successor failures.
	if w.Halted {
		return nil, NewWorkerWrappingRuntimeException(NewInvariantViolatedException(), succState)
	}
	if action != nil && CoverageActionEnabled() {
		action.CM.IncInvocations()
	}
	w.statesGenerated.Add(1)
	stop, queued, err := w.Checker.processSuccessorForWorker(w, curState, succState, action, w.SetOfStates)
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

func (w *Worker) AddUnsatisfiedNextState(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	if w != nil && w.Checker != nil && w.Checker.AllStateWriter != nil && w.Checker.AllStateWriter.IsConstrained() {
		_ = w.Checker.AllStateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action, pred)
	}
	return succState
}

func (w *Worker) IncrementStatesGenerated(count int64) {
	if w != nil {
		w.statesGenerated.Add(count)
	}
}

func (w *Worker) Halt() bool {
	return w != nil && w.Halted
}

func (w *Worker) GetStatesGenerated() int64 {
	if w == nil {
		return 0
	}
	return w.statesGenerated.Load()
}

const workerSetOfStatesInitialCapacity = 16
const workerOutDegreeBucketCount = 32

func (w *Worker) CreateSetOfStates() *SetOfStates {
	if w == nil {
		return NewSetOfStates(workerSetOfStatesInitialCapacity)
	}
	multiplier := w.SetOfStatesMultiplier
	if multiplier < 1 {
		multiplier = 1
		w.SetOfStatesMultiplier = multiplier
	}
	return NewSetOfStates(multiplier * workerSetOfStatesInitialCapacity)
}

func (w *Worker) HasStates() bool {
	return w != nil && w.SetOfStates != nil
}

func (w *Worker) GetStates() *SetOfStates {
	if w != nil && w.SetOfStates != nil {
		return w.SetOfStates
	}
	return NewSetOfStates(0)
}

func (w *Worker) RecordOutDegree() {
	if w == nil {
		return
	}
	if w.OutDegree == nil {
		w.OutDegree = NewFixedSizedBucketStatistics(fmt.Sprintf("TLCWorkerThread-%03d", w.ID), workerOutDegreeBucketCount)
	}
	w.OutDegree.AddSample(w.UnseenSuccessorStates)
	w.UnseenSuccessorStates = 0
}

const tlcTraceExt = ".st"

func (w *Worker) configureTrace() {
	if w == nil || w.Checker == nil {
		return
	}
	metadir := w.Checker.Metadir
	rootName := "Spec"
	if w.Tool != nil {
		rootName = w.Tool.GetRootName()
	}
	if metadir == "" {
		return
	}
	w.traceFileBase = workerTraceFileBase(metadir, rootName, w.ID)
}

func workerTraceFileBase(metadir string, rootName string, id int) string {
	if metadir == "" {
		return ""
	}
	if rootName == "" {
		rootName = "Spec"
	}
	// Worker retains FileUtil separator concatenation, including dot components.
	return metadir + string(os.PathSeparator) + fmt.Sprintf("%s-%d", rootName, id)
}

func (w *Worker) SetTraceContext(metadir string, rootName string) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	base := workerTraceFileBase(metadir, rootName, w.ID)
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
	if w.traceFileBase == "" {
		return nil
	}
	raf, err := NewBufferedRandomAccessFile(w.traceFileBase+tlcTraceExt, "rw")
	if err != nil {
		w.traceErr = err
		return err
	}
	w.traceRAF = raf
	return nil
}

func (w *Worker) WriteInitState(initialState *TLCStateMut, fp uint64) error {
	if w == nil {
		panic(NewNullPointerException())
	}
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
	if err := w.traceRAF.WriteShortNat(w.ID); err != nil {
		return err
	}
	if err := w.traceRAF.WriteLong(int64(fp)); err != nil {
		return err
	}
	if initialState == nil {
		panic(NewNullPointerException())
	}
	initialState.WorkerID = int16(w.ID)
	initialState.UID = ptr
	if !w.DisableTraceMirror && w.Checker != nil && w.Checker.Trace != nil {
		w.Checker.Trace.MirrorInitStateForWorker(w.ID, initialState, fp, ptr)
	}
	return nil
}

func (w *Worker) WriteNextState(curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action) error {
	if w == nil {
		panic(NewNullPointerException())
	}
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
	succState.WorkerID = int16(w.ID)
	succState.UID = ptr
	succState.SetPredecessor(curState)
	w.UnseenSuccessorStates++
	if !w.DisableTraceMirror && w.Checker != nil && w.Checker.Trace != nil {
		w.Checker.Trace.MirrorNextStateForWorker(w.ID, curState, succFP, succState, action, ptr)
	}
	return nil
}

func (w *Worker) ReadStateRecord(ptr int64) (ConcurrentTraceRecord, error) {
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
	workerID, err := w.traceRAF.ReadShortNat()
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	fp, err := w.traceRAF.ReadLong()
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	if err := w.traceRAF.Seek(w.traceRAF.GetMark()); err != nil {
		return ConcurrentTraceRecord{}, err
	}
	return NewConcurrentTraceRecord(prev, workerID, uint64(fp)), nil
}

func (w *Worker) BeginChkpt() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	if w.traceFileBase == "" {
		return nil
	}
	if err := w.traceRAF.Flush(); err != nil {
		return err
	}
	file, err := os.Create(w.traceFileBase + ".tmp")
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	filePtr, err := w.traceRAF.GetFilePointer()
	if err != nil {
		_ = out.Close()
		return err
	}
	if err := out.WriteLong(filePtr); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.WriteLong(w.lastPtr); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func (w *Worker) CommitChkpt() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.traceFileBase == "" {
		return nil
	}
	oldChkpt := w.traceFileBase + ".chkpt"
	newChkpt := w.traceFileBase + ".tmp"
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return NewIOException(fmt.Sprintf("Trace.commitChkpt: cannot delete %s", oldChkpt))
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException(fmt.Sprintf("Trace.commitChkpt: cannot delete %s", oldChkpt))
	}
	return nil
}

func (w *Worker) RecoverTrace() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.traceFileBase == "" {
		w.configureTrace()
	}
	if w.traceFileBase == "" {
		return nil
	}
	file, err := os.Open(w.traceFileBase + ".chkpt")
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	filePos, err := in.ReadLong()
	if err != nil {
		_ = in.Close()
		return err
	}
	lastPtr, err := in.ReadLong()
	if err != nil {
		_ = in.Close()
		return err
	}
	w.lastPtr = lastPtr
	if err := in.Close(); err != nil {
		return err
	}
	if w.traceRAF == nil {
		panic(NewNullPointerException())
	}
	return w.traceRAF.Seek(filePos)
}

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
	w.mu.Lock()
	defer w.mu.Unlock()
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

func (e *WorkerTraceEnumerator) HasMoreFP() bool {
	more, err := e.HasMoreFPWithError()
	if err != nil {
		panic(err)
	}
	return more
}

func (e *WorkerTraceEnumerator) HasMoreFPWithError() (bool, error) {
	if e == nil || e.raf == nil {
		return false, NewNullPointerException()
	}
	pos, err := e.raf.GetFilePointer()
	if err != nil {
		return false, err
	}
	return pos < e.length, nil
}

func (e *WorkerTraceEnumerator) NextFP() (uint64, error) {
	if e == nil || e.raf == nil {
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
	if e == nil || e.raf == nil {
		return NewNullPointerException()
	}
	return e.raf.Close()
}

func (w *Worker) GetLocalValue(index int) Value {
	if w == nil || index < 0 || index >= len(w.LocalValues) {
		return nil
	}
	return w.LocalValues[index]
}

func (w *Worker) SetLocalValue(index int, value Value) {
	if w == nil || index < 0 {
		return
	}
	for len(w.LocalValues) <= index {
		w.LocalValues = append(w.LocalValues, nil)
	}
	w.LocalValues[index] = value
}

func (w *Worker) GetNamedRegister(name *UniqueString) Value {
	if w == nil || w.NamedRegisters == nil {
		return nil
	}
	return w.NamedRegisters.Get(name)
}

func (w *Worker) SetNamedRegister(name *UniqueString, value Value) {
	if w == nil {
		return
	}
	if w.NamedRegisters == nil {
		w.NamedRegisters = NewInsMap[*UniqueString, Value]()
	}
	w.NamedRegisters.Set(name, value)
}
