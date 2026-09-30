package tlc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sync/atomic"
)

type SimulationWorkerError struct {
	Code       int
	Params     []string
	StateTrace *StateVec
	Err        error
}

func (e *SimulationWorkerError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("simulation worker error %d", e.Code)
}

func (e *SimulationWorkerError) HasTrace() bool {
	return e != nil && e.StateTrace != nil && e.StateTrace.Size() > 0
}

func (e *SimulationWorkerError) GetCounterExample() *CounterExample {
	if e == nil {
		return NewEmptyCounterExample()
	}
	var liveCounterExample *LiveCounterExampleException
	if errors.As(e.Err, &liveCounterExample) && liveCounterExample.CounterExample != nil {
		return liveCounterExample.CounterExample
	}
	var liveException *LiveException
	if errors.As(e.Err, &liveException) && liveException.CounterExample != nil {
		return liveException.CounterExample
	}
	return NewCounterExampleFromStateVec(e.StateTrace)
}

type SimulationWorkerResult struct {
	WorkerID int
	Error    *SimulationWorkerError
}

func SimulationWorkerOK(workerID int) SimulationWorkerResult {
	return SimulationWorkerResult{WorkerID: workerID}
}

func SimulationWorkerFailed(workerID int, err *SimulationWorkerError) SimulationWorkerResult {
	return SimulationWorkerResult{WorkerID: workerID, Error: err}
}

func (r SimulationWorkerResult) IsError() bool {
	return r.Error != nil
}

type SimulationWorkerMode int

const (
	SimulationWorkerStandard SimulationWorkerMode = iota
	SimulationWorkerRL
	SimulationWorkerRLAction
)

type SimulationWorkerStatistics struct {
	TraceActions    string
	NumGenStates    *atomic.Int64
	NumGenTraces    *atomic.Int64
	WelfordM2Mean   *atomic.Int64
	ActionStats     [][]int64
	NextRetries     int64
	DistinctStates  int64
	DistinctValues  *InsMap[*UniqueString, int64]
	ActionCounts    *InsMap[*UniqueString, int64]
	Extended        bool
	TraceID         int64
	workerActionIDs *InsMap[*UniqueString, int]
	distinctStates  *CountDistinct
	distinctValues  *InsMap[*UniqueString, *CountDistinct]
}

func NewSimulationWorkerStatistics(tool *Tool, traceActions string, states *atomic.Int64, traces *atomic.Int64, m2Mean *atomic.Int64) *SimulationWorkerStatistics {
	if states == nil {
		states = &atomic.Int64{}
	}
	if traces == nil {
		traces = &atomic.Int64{}
	}
	if m2Mean == nil {
		m2Mean = &atomic.Int64{}
	}
	size := 1
	if traceActions != "" && tool != nil {
		size = len(tool.GetActions())
		if size < 1 {
			size = 1
		}
	}
	stats := &SimulationWorkerStatistics{
		TraceActions:    traceActions,
		NumGenStates:    states,
		NumGenTraces:    traces,
		WelfordM2Mean:   m2Mean,
		ActionStats:     make([][]int64, size),
		DistinctValues:  NewInsMap[*UniqueString, int64](),
		ActionCounts:    NewInsMap[*UniqueString, int64](),
		workerActionIDs: NewInsMap[*UniqueString, int](),
		Extended:        simulatorPropertyBool("tlc2.tool.Simulator.extendedStatistics", "TLAGO_SIMULATOR_EXTENDED_STATISTICS"),
		distinctValues:  NewInsMap[*UniqueString, *CountDistinct](),
	}
	if stats.Extended {
		stats.distinctStates = newSimulationCountDistinct(8)
		for _, variable := range StateVariables() {
			if variable.Name != nil {
				stats.distinctValues.Set(variable.Name, newSimulationCountDistinct(10))
			}
		}
	}
	for i := range stats.ActionStats {
		stats.ActionStats[i] = make([]int64, size)
	}
	if tool != nil {
		for id, action := range tool.GetActions() {
			if action != nil {
				stats.workerActionIDs.Set(UniqueStringOf(action.GetName()), id)
			}
		}
	}
	return stats
}

func (s *SimulationWorkerStatistics) CollectPreSuccessor(state *TLCStateMut, action *Action, next *TLCStateMut) {
	if s == nil {
		return
	}
	s.NumGenStates.Add(1)
	if s.Extended && next != nil {
		if s.distinctStates != nil {
			s.distinctStates.AddState(next)
			s.DistinctStates = s.distinctStates.Count()
		}
		for _, variable := range StateVariables() {
			if variable.Name == nil {
				continue
			}
			value := next.Lookup(variable.Name)
			if value == nil {
				continue
			}
			counter := s.distinctValues.Get(variable.Name)
			if counter != nil {
				counter.AddValue(value)
				s.DistinctValues.Set(variable.Name, counter.Count())
			}
		}
		actionName := state.GetAction().GetName()
		actionKey := UniqueStringOf(actionName)
		s.ActionCounts.Set(actionKey, s.ActionCounts.Get(actionKey)+1)
	}
}

func newSimulationCountDistinct(bits int) *CountDistinct {
	if simulatorPropertyBool("tlc2.tool.Simulator.extendedStatistics.naive", "TLAGO_SIMULATOR_EXTENDED_STATISTICS_NAIVE") {
		return NewCountDistinctNaive()
	}
	return NewCountDistinctHyperLogLog(bits)
}

func (s *SimulationWorkerStatistics) CollectPostSuccessor(state *TLCStateMut, action *Action, next *TLCStateMut) {
	if s == nil || s.TraceActions == "" || state == nil || next == nil {
		return
	}
	from := actionIDFromStateAction(s, state.GetAction())
	to := actionIDFromStateAction(s, next.GetAction())
	if from >= 0 && from < len(s.ActionStats) && to >= 0 && to < len(s.ActionStats[from]) {
		s.ActionStats[from][to]++
	}
	_ = action
}

func (s *SimulationWorkerStatistics) CollectPreTrace() int64 {
	if s == nil {
		return 0
	}
	return s.NumGenTraces.Add(1)
}

func (s *SimulationWorkerStatistics) CollectNextRetries() {
	if s != nil {
		s.NextRetries++
	}
}

func (s *SimulationWorkerStatistics) CollectPostTrace(state *TLCStateMut, maxTraceDepth int) {
	if s == nil || state == nil {
		return
	}
	traceLen := int64(state.Level())
	if maxTraceDepth >= 0 && traceLen > int64(maxTraceDepth) {
		traceLen = int64(maxTraceDepth)
	}
	for {
		old := s.WelfordM2Mean.Load()
		mean := int64(int32(old & 0xffffffff))
		m2 := old >> 32
		count := s.NumGenTraces.Load()
		if count < 1 {
			count = 1
		}
		delta := traceLen - mean
		nextMean := mean + delta/count
		nextM2 := m2 + delta*(traceLen-nextMean)
		next := (nextM2 << 32) | (nextMean & 0xffffffff)
		if s.WelfordM2Mean.CompareAndSwap(old, next) {
			return
		}
	}
}

func (s *SimulationWorkerStatistics) GetNextRetries() Value {
	if s == nil || !s.Extended {
		return NewIntValue(-1)
	}
	return intValueFromInt64(s.NextRetries)
}

func (s *SimulationWorkerStatistics) GetTraceStatistics(state *TLCStateMut) Value {
	actionCounts := NewInsMap[*UniqueString, Value]()
	for cur := state; cur != nil && !cur.IsInitial(); cur = cur.Predecessor() {
		action := cur.GetAction()
		actionKey := UniqueStringOf(action.GetName())
		count := int32(1)
		if old, ok := actionCounts.Get(actionKey).(*IntValue); ok {
			count = old.Val + 1
		}
		actionCounts.Set(actionKey, NewIntValue(count))
	}
	return NewRecordValue(
		[]*UniqueString{tlcSpecActions, tlcGetID},
		[]Value{NewRecordValueFromInsMap(actionCounts), intValueFromInt64(s.traceCount())},
		false,
	)
}

func (s *SimulationWorkerStatistics) GetDistinctStates() Value {
	if s == nil || !s.Extended {
		return NewIntValue(-1)
	}
	if s.distinctStates != nil {
		return intValueFromInt64(s.distinctStates.Count())
	}
	return intValueFromInt64(s.DistinctStates)
}

func (s *SimulationWorkerStatistics) GetDistinctValues() Value {
	if s == nil || !s.Extended {
		return NewIntValue(-1)
	}
	values := NewInsMap[*UniqueString, Value]()
	for _, variable := range StateVariables() {
		if variable.Name == nil {
			continue
		}
		counter := s.distinctValues.Get(variable.Name)
		count := s.DistinctValues.Get(variable.Name)
		if counter != nil {
			count = counter.Count()
		}
		values.Set(variable.Name, intValueFromInt64(count))
	}
	return NewRecordValueFromInsMap(values)
}

func (s *SimulationWorkerStatistics) GetActions() Value {
	if s == nil || !s.Extended {
		return EmptyRecord
	}
	values := NewInsMap[*UniqueString, Value]()
	for key := range s.workerActionIDs.All() {
		values.Set(key, intValueFromInt64(s.ActionCounts.Get(key)))
	}
	return NewRecordValueFromInsMap(values)
}

func (s *SimulationWorkerStatistics) traceCount() int64 {
	if s == nil {
		return 0
	}
	return s.TraceID
}

type SimulationWorker struct {
	ID            int
	Tool          *Tool
	Rand          *JavaRandom
	CurState      *TLCStateMut
	InitStates    *StateVec
	ResultQueue   chan SimulationWorkerResult
	TraceCnt      int64
	GlobalTrace   int64
	MaxTraceNum   int64
	MaxTraceDepth int
	CheckDeadlock bool
	debug         bool
	Mode          SimulationWorkerMode
	TraceFile     string
	LiveCheck     *LiveCheck
	Statistics    *SimulationWorkerStatistics
	Stopped       atomic.Bool
	Halted        atomic.Bool
	NextStates    *StateVec
	RLAlpha       float64
	RLGamma       float64
	RLReward      float64
	RLEnabledOnly bool
	RLQ           *InsMap[*Action, *InsMap[int64, float64]]
}

func NewSimulationWorker(id int, tool *Tool, results chan SimulationWorkerResult, seed int64, maxTraceDepth int, maxTraceNum int64, traceActions string, checkDeadlock bool, debug bool, traceFile string, liveCheck *LiveCheck, states *atomic.Int64, traces *atomic.Int64, m2Mean *atomic.Int64) *SimulationWorker {
	if results == nil {
		results = make(chan SimulationWorkerResult, 1)
	}
	if maxTraceNum <= 0 {
		maxTraceNum = math.MaxInt64
	}
	return &SimulationWorker{
		ID:            id,
		Tool:          tool,
		Rand:          NewJavaRandom(seed),
		ResultQueue:   results,
		MaxTraceNum:   maxTraceNum,
		MaxTraceDepth: maxTraceDepth,
		CheckDeadlock: checkDeadlock,
		debug:         debug,
		Mode:          SimulationWorkerStandard,
		TraceFile:     traceFile,
		LiveCheck:     liveCheck,
		Statistics:    NewSimulationWorkerStatistics(tool, traceActions, states, traces, m2Mean),
		NextStates:    NewStateVec(1),
		RLAlpha:       0.3,
		RLGamma:       0.7,
		RLReward:      -10,
	}
}

func (w *SimulationWorker) Stop() {
	if w != nil {
		w.Stopped.Store(true)
	}
}

func (w *SimulationWorker) Start(initStates *StateVec) {
	if w == nil {
		return
	}
	w.SetInitialStates(initStates)
	go w.Run()
}

func (w *SimulationWorker) Run() {
	for w != nil && !w.Stopped.Load() {
		if !w.SimulateAndReport() {
			return
		}
	}
}

func (w *SimulationWorker) SimulateAndReport() bool {
	if w == nil {
		return false
	}
	restoreWorkerID := PushCurrentWorkerID(w.ID)
	defer restoreWorkerID()
	w.GlobalTrace = w.Statistics.CollectPreTrace()
	w.Statistics.TraceID = w.GlobalTrace
	err := w.SimulateRandomTrace()
	w.TraceCnt++
	if err != nil {
		w.ResultQueue <- SimulationWorkerFailed(w.ID, err)
	}
	if w.TraceCnt >= w.MaxTraceNum || w.Stopped.Load() {
		w.ResultQueue <- SimulationWorkerOK(w.ID)
		return false
	}
	return true
}

func (w *SimulationWorker) RandomState(states *StateVec) *TLCStateMut {
	if states == nil || states.Size() == 0 {
		return nil
	}
	index := int(math.Floor(w.Rand.NextDouble() * float64(states.Size())))
	return states.At(index)
}

func (w *SimulationWorker) claimInitialState(state *TLCStateMut) *TLCStateMut {
	if state == nil {
		return nil
	}
	return w.markWorkerState(state.DeepCopy())
}

func (w *SimulationWorker) markWorkerState(state *TLCStateMut) *TLCStateMut {
	if w != nil && state != nil && w.ID >= 0 && w.ID < int(TLCStateInitWorkerID) {
		state.WorkerID = int16(w.ID)
	}
	return state
}

func (w *SimulationWorker) GetNextActionIndex(actions []*Action, curState *TLCStateMut) int {
	if len(actions) == 0 {
		return -1
	}
	if w.IsRLMode() {
		return w.GetRLNextActionIndex(actions, curState)
	}
	return int(math.Floor(w.Rand.NextDouble() * float64(len(actions))))
}

func (w *SimulationWorker) GetNextActionAltIndex(index int, p int, actions []*Action, curState *TLCStateMut) int {
	if w.IsRLMode() {
		w.UpdateRLDisabledAction(index, actions, curState)
	}
	if w != nil && w.Statistics != nil {
		w.Statistics.CollectNextRetries()
	}
	if len(actions) == 0 {
		return -1
	}
	return (index + p) % len(actions)
}

func (w *SimulationWorker) SimulateRandomTrace() *SimulationWorkerError {
	if w == nil || w.Tool == nil {
		return &SimulationWorkerError{Code: ECGeneral, Err: newTLCError(ECGeneral, "simulation worker has no tool")}
	}
	if w.debug {
		return w.SimulateExplorationTrace()
	}
	w.CurState = w.claimInitialState(w.RandomState(w.InitStates))
	allActions := w.Tool.GetActions()
	for traceIdx := 0; traceIdx < w.MaxTraceDepth; traceIdx++ {
		if w.Stopped.Load() {
			return nil
		}
		w.NextStates.Clear()
		actions, workerErr := w.FilterActions(allActions, w.CurState)
		if workerErr != nil {
			return workerErr
		}
		if len(actions) == 0 {
			if w.CheckDeadlock {
				return &SimulationWorkerError{Code: ECTLCDeadlockReached, StateTrace: w.GetTrace(w.CurState)}
			}
			break
		}
		index := w.GetNextActionIndex(actions, w.CurState)
		step := w.Rand.NextPrime()
		for i := 0; i < len(actions); i++ {
			action := actions[index]
			nextStates, err := w.Tool.GetNextStates(action, w.CurState)
			if err != nil {
				return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
			}
			if workerErr := w.AddGeneratedSuccessors(action, nextStates); workerErr != nil {
				return workerErr
			}
			if !w.NextStates.IsEmpty() {
				break
			}
			index = w.GetNextActionAltIndex(index, step, actions, w.CurState)
		}
		if w.NextStates.IsEmpty() {
			if w.CheckDeadlock {
				return &SimulationWorkerError{Code: ECTLCDeadlockReached, StateTrace: w.GetTrace(w.CurState)}
			}
			break
		}
		next := w.RandomState(w.NextStates)
		if _, err := next.ExecCallable(); err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(next), Err: err}
		}
		w.Statistics.CollectPostSuccessor(w.CurState, next.GetAction(), next)
		w.CurState = next
	}
	if w.Stopped.Load() {
		return nil
	}
	if w.LiveCheck != nil {
		if err := w.LiveCheck.CheckTrace(w.Tool.NoDebug(), func() *StateVec { return w.GetTrace(w.CurState) }); err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
		}
	}
	w.Statistics.CollectPostTrace(w.CurState, w.MaxTraceDepth)
	if w.TraceFile != "" {
		if err := w.WriteTraceFile(); err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
		}
	}
	if workerErr := w.PostTrace(w.CurState); workerErr != nil {
		return workerErr
	}
	return nil
}

func (w *SimulationWorker) FilterActions(actions []*Action, curState *TLCStateMut) ([]*Action, *SimulationWorkerError) {
	if w.IsRLMode() && w.RLEnabledOnly {
		return w.FilterEnabledRLActions(actions, curState)
	}
	return actions, nil
}

func (w *SimulationWorker) PostTrace(finalState *TLCStateMut) *SimulationWorkerError {
	if w.Mode == SimulationWorkerRL {
		return w.PostRLTrace(finalState)
	}
	return nil
}

func (w *SimulationWorker) AddGeneratedSuccessors(action *Action, nextStates *StateVec) *SimulationWorkerError {
	if nextStates == nil {
		return nil
	}
	for i := 0; i < nextStates.Size(); i++ {
		succ := nextStates.At(i)
		if succ == nil {
			continue
		}
		if workerErr := w.AddGeneratedSuccessor(w.CurState, action, succ); workerErr != nil {
			return workerErr
		}
	}
	return nil
}

func (w *SimulationWorker) AddGeneratedSuccessor(curState *TLCStateMut, action *Action, succ *TLCStateMut) *SimulationWorkerError {
	if succ == nil {
		return nil
	}
	if action != nil && action.CM.node != nil {
		action.CM.IncInvocations()
	}
	w.markWorkerState(succ)
	succ.SetPredecessor(curState).SetAction(action)
	if !w.Tool.IsGoodState(succ) {
		return &SimulationWorkerError{Code: ECTLCStateNotCompletelySpecifiedNext, Params: incompleteNextStateParams(w.Tool, action, succ), StateTrace: w.GetTrace(succ)}
	}
	w.Statistics.CollectPreSuccessor(curState, action, succ)
	if workerErr := w.CheckInvariants(succ); workerErr != nil {
		return workerErr
	}
	if workerErr := w.CheckImpliedActions(succ); workerErr != nil {
		return workerErr
	}
	inModel, err := w.Tool.IsInModel(succ)
	if err != nil {
		return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(succ), Err: err}
	}
	if inModel {
		inActions, err := w.Tool.IsInActions(curState, succ)
		if err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(succ), Err: err}
		}
		inModel = inActions
	}
	if inModel {
		if action != nil && action.CM.node != nil {
			action.CM.IncSecondary()
		}
		w.NextStates.Add(succ)
	}
	return nil
}

func (w *SimulationWorker) CheckInvariants(state *TLCStateMut) *SimulationWorkerError {
	names := w.Tool.GetInvNames()
	for i, invariant := range w.Tool.GetInvariants() {
		valid, err := w.Tool.IsValidState(invariant, state)
		if err != nil {
			return &SimulationWorkerError{Code: ECTLCInvariantEvaluationFailed, Params: []string{nameAt(names, i), err.Error()}, StateTrace: w.GetTrace(state), Err: err}
		}
		if !valid {
			return &SimulationWorkerError{Code: ECTLCInvariantViolatedBehavior, Params: []string{nameAt(names, i)}, StateTrace: w.GetTrace(state)}
		}
	}
	return nil
}

func (w *SimulationWorker) CheckImpliedActions(state *TLCStateMut) *SimulationWorkerError {
	names := w.Tool.GetImpliedActNames()
	for i, action := range w.Tool.GetImpliedActions() {
		valid, err := w.Tool.IsValidTransition(action, w.CurState, state)
		if err != nil {
			return &SimulationWorkerError{Code: ECTLCActionPropertyEvaluationFailed, Params: []string{nameAt(names, i), err.Error()}, StateTrace: w.GetTrace(state), Err: err}
		}
		if !valid {
			return &SimulationWorkerError{Code: ECTLCActionPropertyViolatedBehavior, Params: []string{nameAt(names, i)}, StateTrace: w.GetTrace(state)}
		}
	}
	return nil
}

func (w *SimulationWorker) GetTrace(state *TLCStateMut) *StateVec {
	stack := make([]*TLCStateMut, 0)
	for cur := state; cur != nil; cur = cur.Predecessor() {
		pred := cur.Predecessor()
		if cur.Equal(pred) {
			continue
		}
		stack = append(stack, cur)
	}
	trace := NewStateVec(len(stack))
	for i := len(stack) - 1; i >= 0; i-- {
		trace.Add(stack[i])
	}
	for i := 1; i < trace.Size(); i++ {
		trace.At(i).SetPredecessor(trace.At(i - 1))
	}
	return trace
}

func (w *SimulationWorker) GetUncompressedTrace(state *TLCStateMut) *StateVec {
	stack := make([]*TLCStateMut, 0)
	for cur := state; cur != nil; cur = cur.Predecessor() {
		stack = append(stack, cur)
	}
	trace := NewStateVec(len(stack))
	for i := len(stack) - 1; i >= 0; i-- {
		trace.Add(stack[i])
	}
	return trace
}

func (w *SimulationWorker) SetInitialStates(initStates *StateVec) {
	if w != nil {
		w.InitStates = initStates
	}
}

func (w *SimulationWorker) GetTraceCnt() int64 {
	if w == nil {
		return 0
	}
	return w.TraceCnt + 1
}

func (w *SimulationWorker) GetRNG() *JavaRandom {
	if w == nil {
		return nil
	}
	return w.Rand
}

func (w *SimulationWorker) WriteTraceFile() error {
	trace := w.GetTrace(w.CurState)
	name := fmt.Sprintf("%s_%d_%d", w.TraceFile, w.ID, w.TraceCnt)
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := fmt.Fprintf(file, "---------------- MODULE %s -----------------\n", name); err != nil {
		return err
	}
	for i := 0; i < trace.Size(); i++ {
		state := trace.At(i)
		if state.HasAction() {
			if _, err := fmt.Fprintf(file, "\\* %s\n", state.GetAction().GetLocation()); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(file, "STATE_%d == \n%s\n\n", i+1, state.String()); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(file, "=================================================")
	return err
}

func actionIDFromStateAction(stats *SimulationWorkerStatistics, action *Action) int {
	if action == nil {
		return -1
	}
	if action.ID >= 0 {
		return action.ID
	}
	if stats == nil || stats.workerActionIDs == nil {
		return -1
	}
	id, ok := stats.workerActionIDs.Get2(UniqueStringOf(action.GetName()))
	if !ok {
		return -1
	}
	return id
}
