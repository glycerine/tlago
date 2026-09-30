package tlc

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var errInvariantViolated = errors.New("tlc invariant violated")

type AbstractChecker struct {
	mu                        sync.Mutex
	PredErrState              *TLCStateMut
	ErrState                  *TLCStateMut
	ErrorCode                 int
	Done                      bool
	KeepCallStack             bool
	CheckDeadlock             bool
	CheckLiveness             bool
	FromCheckpoint            string
	Metadir                   string
	Tool                      *Tool
	AllStateWriter            *StateWriter
	Workers                   []*Worker
	PrintedLivenessErrorStack bool
	StartTime                 time.Time
	Values                    *InsMap[int, Value]
	NamedValues               *InsMap[*UniqueString, Value]
}

func NewAbstractChecker(tool *Tool, metadir string, stateWriter *StateWriter, deadlock bool, fromCheckpoint string, startTime time.Time) *AbstractChecker {
	if stateWriter == nil {
		stateWriter = NewNoopStateWriter()
	}
	if startTime.IsZero() {
		startTime = time.Now()
	}
	checkLiveness := false
	if tool != nil {
		checkLiveness = !tool.LivenessIsTrue()
	}
	return &AbstractChecker{
		ErrorCode:      NoError,
		CheckDeadlock:  deadlock,
		CheckLiveness:  checkLiveness,
		FromCheckpoint: fromCheckpoint,
		Metadir:        metadir,
		Tool:           tool,
		AllStateWriter: stateWriter,
		StartTime:      startTime,
		Values:         NewInsMap[int, Value](),
		NamedValues:    NewInsMap[*UniqueString, Value](),
	}
}

func (c *AbstractChecker) SetDone() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	old := c.Done
	c.Done = true
	return old
}

func (c *AbstractChecker) SetErrState(curState *TLCStateMut, succState *TLCStateMut, keepCallStack bool, errorCode int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !continuationEnabled() && c.Done {
		return false
	}
	c.PredErrState = curState
	if succState == nil {
		c.ErrState = curState
	} else {
		c.ErrState = succState
	}
	c.ErrorCode = errorCode
	c.Done = true
	c.KeepCallStack = keepCallStack
	return true
}

func (c *AbstractChecker) SetError(keepCallStack bool, errorCode int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ErrorCode = errorCode
	c.Done = true
	c.KeepCallStack = keepCallStack
}

func continuationEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.Continuation
}

func (c *AbstractChecker) GetValue(workerID int, idx int) Value {
	_ = workerID
	if c == nil || c.Values == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Values.Get(idx)
}

func (c *AbstractChecker) SetAllValues(idx int, value Value) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Values == nil {
		c.Values = NewInsMap[int, Value]()
	}
	c.Values.Set(idx, value)
}

func (c *AbstractChecker) GetAllValues() Value {
	if c == nil || c.Values == nil {
		return EmptyFcn
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	domain := make([]Value, 0, c.Values.Len())
	values := make([]Value, 0, c.Values.Len())
	for idx, value := range c.Values.All() {
		domain = append(domain, NewIntValue(int32(idx)))
		values = append(values, value)
	}
	return NewFcnRcdValue(domain, values, false)
}

func (c *AbstractChecker) GetNamedValue(workerID int, key *UniqueString) Value {
	_ = workerID
	if c == nil || c.NamedValues == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.NamedValues.Get(key)
}

func (c *AbstractChecker) SetAllNamedValues(key *UniqueString, value Value) {
	if c == nil || key == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.NamedValues == nil {
		c.NamedValues = NewInsMap[*UniqueString, Value]()
	}
	c.NamedValues.Set(key, value)
}

func (c *AbstractChecker) GetAllNamedRegisterValues() Value {
	if c == nil || c.NamedValues == nil {
		return EmptyFcn
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	domain := make([]Value, 0, c.NamedValues.Len())
	values := make([]Value, 0, c.NamedValues.Len())
	for key, value := range c.NamedValues.All() {
		domain = append(domain, NewStringValueFromUnique(key))
		values = append(values, value)
	}
	return NewFcnRcdValue(domain, values, false)
}

func (c *AbstractChecker) GetAllNamedValues(key *UniqueString) []Value {
	value := c.GetNamedValue(0, key)
	if value == nil {
		return nil
	}
	return []Value{value}
}

type ModelChecker struct {
	*AbstractChecker
	NumberOfInitialStates int64
	FPSet                 *MemFPSet
	StateQueue            *MemStateQueue
	Trace                 *MemoryTrace
	NextStatesGenerated   int64
}

type ModelCheckerOption func(*ModelChecker)

func WithModelCheckerFPSet(fpSet *MemFPSet) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.FPSet = fpSet
	}
}

func WithModelCheckerStateQueue(queue *MemStateQueue) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.StateQueue = queue
	}
}

func WithModelCheckerStateWriter(writer *StateWriter) ModelCheckerOption {
	return func(mc *ModelChecker) {
		if writer == nil {
			writer = NewNoopStateWriter()
		}
		mc.AllStateWriter = writer
	}
}

func WithModelCheckerTrace(trace *MemoryTrace) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.Trace = trace
	}
}

func NewModelChecker(tool *Tool, metadir string, deadlock bool, opts ...ModelCheckerOption) *ModelChecker {
	checkDeadlock := deadlock
	if tool != nil {
		if config := tool.GetModelConfig(); config != nil {
			checkDeadlock = deadlock && config.GetCheckDeadlock()
		}
	}
	mc := &ModelChecker{
		AbstractChecker: NewAbstractChecker(tool, metadir, NewNoopStateWriter(), checkDeadlock, "", time.Now()),
		FPSet:           NewMemFPSet(),
		StateQueue:      NewMemStateQueue(),
		Trace:           NewMemoryTrace(),
	}
	for _, opt := range opts {
		opt(mc)
	}
	if mc.FPSet == nil {
		mc.FPSet = NewMemFPSet()
	}
	if mc.StateQueue == nil {
		mc.StateQueue = NewMemStateQueue()
	}
	if mc.Trace == nil {
		mc.Trace = NewMemoryTrace()
	}
	SetMainChecker(mc)
	return mc
}

func (mc *ModelChecker) Stop() {
	if mc != nil && mc.AbstractChecker != nil {
		mc.SetDone()
	}
}

func (mc *ModelChecker) GetProgress() int64 {
	if mc == nil || mc.Trace == nil {
		return 0
	}
	return int64(mc.Trace.GetLevelForReporting())
}

func (mc *ModelChecker) GetStatistics() Value {
	if mc == nil {
		return EmptyRecord
	}
	names := []*UniqueString{
		tlcGetGenerated,
		tlcGetDistinct,
		tlcGetInitial,
		tlcGetQueue,
		tlcGetDiameter,
		tlcGetDuration,
	}
	values := []Value{
		intValueFromInt64(mc.GetStatesGenerated()),
		intValueFromUint64(mc.GetDistinctStatesGenerated()),
		intValueFromInt64(mc.GetInitialStatesGenerated()),
		intValueFromInt64(mc.GetStateQueueSize()),
		intValueFromInt64(mc.GetProgress()),
		intValueFromDurationSince(TLCStartTime()),
	}
	return NewRecordValue(names, values, false)
}

func (mc *ModelChecker) GetConfig() Value {
	if mc == nil {
		return EmptyRecord
	}
	depth := int32(0)
	names := []*UniqueString{
		tlcGetMode,
		tlcGetDeadlock,
		tlcGetWorker,
		tlcGetDepth,
	}
	values := []Value{
		NewStringValue("model-checking"),
		NewBoolValue(mc.CheckDeadlock),
		NewIntValue(int32(NumWorkers())),
		NewIntValue(depth),
	}
	return NewRecordValue(names, values, false)
}

func (mc *ModelChecker) CheckAssumptions() int {
	if mc.Tool == nil {
		return NoError
	}
	return mc.Tool.CheckAssumptions()
}

func (mc *ModelChecker) ModelCheck() (int, error) {
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	if result := mc.CheckAssumptions(); result != NoError {
		return result, nil
	}
	result, err := mc.DoInit(false)
	if err != nil || result != NoError {
		return result, err
	}
	if len(mc.Tool.GetActions()) == 0 {
		if !mc.StateQueue.IsEmpty() {
			PrintError(ECTLCStatesAndNoNextAction)
			return ECTLCStatesAndNoNextAction, nil
		}
		return mc.Tool.CheckPostCondition(), nil
	}
	result, err = mc.RunTLC(0)
	if err != nil || result != NoError {
		return result, err
	}
	return mc.Tool.CheckPostCondition(), nil
}

func (mc *ModelChecker) RunTLC(maxDepth int) (int, error) {
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	worker := NewModelCheckingWorker(len(mc.Workers), mc, mc.Tool)
	for {
		if mc.Done {
			return mc.ErrorCode, nil
		}
		curState := mc.StateQueue.Dequeue()
		if curState == nil {
			mc.SetDone()
			return mc.ErrorCode, nil
		}
		if maxDepth > 0 && curState.Level() >= maxDepth {
			continue
		}
		stop, err := worker.DoNext(curState)
		if err != nil {
			if mc.ErrorCode != NoError {
				return mc.ErrorCode, err
			}
			return ECGeneral, err
		}
		if stop {
			return mc.ErrorCode, nil
		}
	}
}

func (mc *ModelChecker) DoInit(ignoreCancel bool) (int, error) {
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	functor := &doInitFunctor{mc: mc, tool: mc.Tool, forceChecks: ignoreCancel, returnValue: NoError}
	err := mc.Tool.GetInitStates(NewStateFunctor(functor.AddElement))
	if errors.Is(err, errInvariantViolated) {
		mc.ErrState = functor.errState
		return functor.returnValue, nil
	}
	if err != nil {
		if functor.errState != nil {
			mc.ErrState = functor.errState
		}
		if functor.returnValue != NoError {
			return functor.returnValue, err
		}
		return ECGeneral, err
	}
	if functor.errState != nil {
		mc.ErrState = functor.errState
		if functor.err != nil {
			return functor.returnValue, functor.err
		}
	}
	return functor.returnValue, nil
}

func (mc *ModelChecker) DoNext(curState *TLCStateMut) (bool, error) {
	if mc.Tool == nil {
		return true, newTLCError(ECGeneral, "model checker has no tool")
	}
	deadLocked := true
	var succState *TLCStateMut
	for _, action := range mc.Tool.GetActions() {
		nextStates, err := mc.Tool.GetNextStates(action, curState)
		if err != nil {
			mc.doNextFailed(curState, succState, err)
			return true, err
		}
		size := 0
		if nextStates != nil {
			size = nextStates.Size()
		}
		mc.NextStatesGenerated += int64(size)
		deadLocked = deadLocked && size == 0
		for i := 0; i < size; i++ {
			succState = nextStates.At(i)
			stop, _, err := mc.processSuccessor(curState, succState, action, nil)
			if stop || err != nil {
				return stop, err
			}
		}
		succState = nil
	}
	if deadLocked && mc.CheckDeadlock {
		return mc.doNextSetErr(curState, nil, false, ECTLCDeadlockReached, ""), nil
	}
	return false, nil
}

func (mc *ModelChecker) processSuccessor(curState *TLCStateMut, succState *TLCStateMut, action *Action, collectedStates *SetOfStates) (bool, bool, error) {
	if !mc.Tool.IsGoodState(succState) {
		return mc.doNextSetErr(curState, succState, false, ECTLCStateNotCompletelySpecifiedNext, actionName(action)), false, nil
	}
	succState.SetPredecessor(curState).SetAction(action)
	inModel, err := mc.Tool.IsInModel(succState)
	if err != nil {
		mc.doNextEvalFailed(curState, succState, ECGeneral, "", err)
		return true, false, err
	}
	if inModel {
		inActions, err := mc.Tool.IsInActions(curState, succState)
		if err != nil {
			mc.doNextEvalFailed(curState, succState, ECGeneral, "", err)
			return true, false, err
		}
		inModel = inActions
	}
	unseen := true
	if inModel {
		seen, err := mc.isSeenState(curState, succState, action)
		if err != nil {
			return true, false, err
		}
		if collectedStates != nil {
			collectedStates.PutFP(succState.FingerPrint(), succState)
		}
		unseen = !seen
	} else if mc.AllStateWriter != nil && mc.AllStateWriter.IsConstrained() {
		if err := mc.AllStateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action); err != nil {
			return true, false, err
		}
	}
	if unseen {
		stop, err := mc.doNextCheckInvariants(curState, succState)
		if stop || err != nil {
			return stop, false, err
		}
	}
	stop, err := mc.doNextCheckImplied(curState, succState)
	if stop || err != nil {
		return stop, false, err
	}
	if inModel && unseen {
		mc.StateQueue.SEnqueue(succState)
		return false, true, nil
	}
	return false, false, nil
}

func (mc *ModelChecker) GetStatesGenerated() int64 {
	total := mc.NumberOfInitialStates + mc.NextStatesGenerated
	for _, worker := range mc.Workers {
		if worker != nil {
			total += worker.StatesGenerated
		}
	}
	return total
}

func (mc *ModelChecker) GetInitialStatesGenerated() int64 {
	return mc.NumberOfInitialStates
}

func (mc *ModelChecker) GetStateQueueSize() int64 {
	if mc.StateQueue == nil {
		return 0
	}
	return mc.StateQueue.Size()
}

func (mc *ModelChecker) GetDistinctStatesGenerated() uint64 {
	if mc.FPSet == nil {
		return 0
	}
	return mc.FPSet.Size()
}

func (mc *ModelChecker) isSeenState(curState *TLCStateMut, succState *TLCStateMut, action *Action) (bool, error) {
	fp := succState.FingerPrint()
	seen := mc.FPSet.Put(fp)
	if mc.AllStateWriter != nil {
		status := StateVisitUnseen
		if seen {
			status = StateVisitSeen
		}
		if err := mc.AllStateWriter.WriteTransition(curState, succState, status, action); err != nil {
			return seen, err
		}
	}
	if !seen {
		if mc.Tool != nil {
			mc.Tool.RememberState(succState)
		}
		if err := mc.Trace.WriteNextState(curState, fp, succState, action); err != nil {
			return seen, err
		}
	}
	return seen, nil
}

func (mc *ModelChecker) doNextCheckInvariants(curState *TLCStateMut, succState *TLCStateMut) (bool, error) {
	invariants := mc.Tool.GetInvariants()
	names := mc.Tool.GetInvNames()
	for i, invariant := range invariants {
		valid, err := mc.Tool.IsValidState(invariant, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCInvariantEvaluationFailed, nameAt(names, i), err)
		}
		if !valid {
			if continuationEnabled() {
				PrintError(ECTLCInvariantViolatedBehavior, nameAt(names, i))
				continue
			}
			return mc.doNextSetErr(curState, succState, false, ECTLCInvariantViolatedBehavior, nameAt(names, i)), nil
		}
	}
	return false, nil
}

func (mc *ModelChecker) doNextCheckImplied(curState *TLCStateMut, succState *TLCStateMut) (bool, error) {
	implied := mc.Tool.GetImpliedActions()
	names := mc.Tool.GetImpliedActNames()
	for i, action := range implied {
		valid, err := mc.Tool.IsValidTransition(action, curState, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCActionPropertyEvaluationFailed, nameAt(names, i), err)
		}
		if !valid {
			if continuationEnabled() {
				PrintError(ECTLCActionPropertyViolatedBehavior, nameAt(names, i))
				continue
			}
			return mc.doNextSetErr(curState, succState, false, ECTLCActionPropertyViolatedBehavior, nameAt(names, i)), nil
		}
	}
	return false, nil
}

func (mc *ModelChecker) doNextSetErr(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, param string) bool {
	if mc.SetErrState(curState, succState, keep, ec) {
		if param == "" {
			PrintError(ec)
		} else {
			PrintError(ec, param)
		}
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
	return true
}

func (mc *ModelChecker) doNextEvalFailed(curState *TLCStateMut, succState *TLCStateMut, ec int, param string, err error) error {
	mc.SetErrState(curState, succState, true, ec)
	if param == "" {
		PrintError(ec, err.Error())
	} else {
		PrintError(ec, param, err.Error())
	}
	if mc.StateQueue != nil {
		mc.StateQueue.FinishAll()
	}
	return err
}

func (mc *ModelChecker) doNextFailed(curState *TLCStateMut, succState *TLCStateMut, err error) {
	if err == nil {
		return
	}
	ec := ECGeneral
	if tlcErr, ok := err.(*TLCError); ok {
		ec = tlcErr.Code
	}
	if mc.SetErrState(curState, succState, false, ec) {
		PrintError(ec, err.Error())
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
}

type doInitFunctor struct {
	mc          *ModelChecker
	tool        *Tool
	forceChecks bool
	errState    *TLCStateMut
	err         error
	returnValue int
}

func (f *doInitFunctor) AddElement(curState *TLCStateMut) (any, error) {
	if isPowerOfTwo(f.mc.NumberOfInitialStates) && f.mc.NumberOfInitialStates > 1 {
		PrintMessage(ECTLCComputingInitProgress, fmt.Sprintf("%d", f.mc.NumberOfInitialStates))
	}
	f.mc.NumberOfInitialStates++
	if f.errState != nil {
		if f.returnValue == NoError {
			f.returnValue = ECTLCInitialState
		}
		return f.returnValue, nil
	}
	if !f.tool.IsGoodState(curState) {
		PrintError(ECTLCInitialState, "current state is not a legal state", curState.String())
		f.errState = curState
		f.returnValue = ECTLCInitialState
		return f.returnValue, errInvariantViolated
	}
	inModel, err := f.tool.IsInModel(curState)
	if err != nil {
		f.errState = curState
		f.err = err
		return f.returnValue, err
	}
	seen := false
	if inModel {
		fp := curState.FingerPrint()
		seen = f.mc.FPSet.Put(fp)
		if !seen {
			if f.tool != nil {
				f.tool.RememberState(curState)
			}
			if err := f.mc.AllStateWriter.WriteInitState(curState); err != nil {
				f.errState = curState
				f.err = err
				return f.returnValue, err
			}
			if err := f.mc.Trace.WriteInitState(curState, fp); err != nil {
				f.errState = curState
				f.err = err
				return f.returnValue, err
			}
			f.mc.StateQueue.Enqueue(curState)
		}
	}
	if !seen || f.forceChecks {
		for i, invariant := range f.tool.GetInvariants() {
			valid, err := f.tool.IsValidState(invariant, curState)
			if err != nil {
				f.errState = curState
				f.err = err
				return f.returnValue, err
			}
			if !valid {
				alias := curState
				if f.tool != nil {
					alias = f.tool.EvalAlias(curState, curState)
				}
				PrintError(ECTLCInvariantViolatedInitial, nameAt(f.tool.GetInvNames(), i), alias.String())
				if !continuationEnabled() {
					f.errState = curState
					f.returnValue = ECTLCInvariantViolatedInitial
					return f.returnValue, errInvariantViolated
				}
			}
		}
		for i, implied := range f.tool.GetImpliedInits() {
			valid, err := f.tool.IsValidState(implied, curState)
			if err != nil {
				f.errState = curState
				f.err = err
				return f.returnValue, err
			}
			if !valid {
				alias := curState
				if f.tool != nil {
					alias = f.tool.EvalAlias(curState, curState)
				}
				PrintError(ECTLCPropertyViolatedInitial, nameAt(f.tool.GetImpliedInitNames(), i), alias.String())
				f.errState = curState
				f.returnValue = ECTLCPropertyViolatedInitial
				return f.returnValue, errInvariantViolated
			}
		}
	}
	return f.returnValue, nil
}

func isPowerOfTwo(n int64) bool {
	return n > 0 && n&(n-1) == 0
}

func nameAt(names []string, index int) string {
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return ""
}

func actionName(action *Action) string {
	if action == nil {
		return ""
	}
	return action.GetName()
}
