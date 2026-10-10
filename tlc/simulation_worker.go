package tlc

import (
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type SimulationWorkerError struct {
	*InvariantViolatedException
	Code           int
	Params         []string
	NullableParams []*string
	StateTrace     *StateVec
	Err            error
	Tool           *Tool
}

func NewSimulationWorkerError(code int, params []string, stateTrace *StateVec, exception ...error) *SimulationWorkerError {
	failure := &SimulationWorkerError{
		InvariantViolatedException: NewInvariantViolatedException(),
		Code:                       code, Params: params, StateTrace: stateTrace,
	}
	if len(exception) > 0 {
		failure.Err = exception[0]
	}
	return failure
}

func (e *SimulationWorkerError) Error() string {
	if e == nil {
		return ""
	}
	return *e.GetMessage()
}

func (e *SimulationWorkerError) HasTrace() bool {
	return e != nil && e.StateTrace != nil && e.StateTrace.Size() > 0
}

func (e *SimulationWorkerError) GetCounterExample() *CounterExample {
	if e == nil {
		return NewEmptyCounterExample()
	}
	if failure, ok := e.Err.(*LiveCounterExampleException); ok && failure != nil {
		return failure.CounterExample
	}
	return NewCounterExampleFromTrace(e.TraceInfo())
}

func (e *SimulationWorkerError) TraceInfo() []*TLCStateInfo {
	if e == nil || e.StateTrace == nil || e.StateTrace.Size() == 0 {
		return nil
	}
	trace := make([]*TLCStateInfo, 0, e.StateTrace.Size())
	for i := 0; i < e.StateTrace.Size(); i++ {
		state := e.StateTrace.At(i)
		successor := state
		if i+1 < e.StateTrace.Size() {
			successor = e.StateTrace.At(i + 1)
		}
		action := UnknownAction
		if state != nil {
			action = state.GetAction()
		}
		info := NewTLCStateInfo(state, action)
		if e.Tool != nil {
			if alias, err := e.Tool.EvalAliasInfoPair(info, successor); err == nil && alias != nil {
				info = alias
			}
		}
		trace = append(trace, info)
	}
	return trace
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
	tool            *Tool
	sourceTool      bool
	distinctStates  *CountDistinct
	distinctValues  *InsMap[*UniqueString, *CountDistinct]

	// Captured once; source access indexes it with current variable locations.
	variableCounters []*CountDistinct
	worker           *SimulationWorker
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
		size = len(tool.GetSpecActions())
		if size < 1 && tool.SpecProcessor == nil {
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
		tool:            tool,
		sourceTool:      tool != nil && tool.SpecProcessor != nil,
		Extended:        simulatorPropertyBool("tlc2.tool.Simulator.extendedStatistics", "TLAGO_SIMULATOR_EXTENDED_STATISTICS"),
		distinctValues:  NewInsMap[*UniqueString, *CountDistinct](),
	}
	if stats.Extended {
		stats.distinctStates = newSimulationCountDistinct(8)
		if stats.sourceTool {
			stats.variableCounters = make([]*CountDistinct, simulationStateVariableCount())
			for i := 0; i < simulationStateVariableCount(); i++ {
				index := stats.variableCounterIndex(simulationStateVariableName(i))
				stats.variableCounters[index] = newSimulationCountDistinct(10)
			}
		} else {
			for _, variable := range StateVariables() {
				if variable.Name != nil {
					stats.distinctValues.Set(variable.Name, newSimulationCountDistinct(10))
				}
			}
		}
	}
	for i := range stats.ActionStats {
		stats.ActionStats[i] = make([]int64, size)
	}
	if tool != nil && !stats.sourceTool {
		for id, action := range tool.GetSpecActions() {
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
	if s.Extended && s.sourceTool {
		for i := 0; i < simulationStateVariableCount(); i++ {
			name := simulationStateVariableName(i)
			counter := s.variableCounters[s.variableCounterIndex(name)]
			if next == nil {
				panic(NewNullPointerException())
			}
			value := next.Lookup(name)
			if counter == nil || value == nil {
				panic(NewNullPointerException())
			}
			counter.AddValue(value)
			s.DistinctValues.Set(name, counter.Count())
		}
		if next == nil || s.distinctStates == nil {
			panic(NewNullPointerException())
		}
		s.distinctStates.AddState(next)
		s.DistinctStates = s.distinctStates.Count()
		if state == nil || state.GetAction() == nil {
			panic(NewNullPointerException())
		}
		key := UniqueStringOf(state.GetAction().GetName())
		s.ActionCounts.Set(key, int64(int32(s.ActionCounts.Get(key))+1))
		return
	}
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

func simulationStateVariableCount() int {
	if stateVariableDeclarations != nil {
		return len(stateVariableDeclarations)
	}
	return len(stateVariables)
}

func simulationStateVariableName(index int) *UniqueString {
	count := simulationStateVariableCount()
	if index < 0 || index >= count {
		panic(NewArrayIndexOutOfBoundsException(index, count))
	}
	if stateVariableDeclarations != nil {
		declaration := stateVariableDeclarations[index]
		if declaration == nil {
			panic(NewNullPointerException())
		}
		return declaration.Name
	}
	return stateVariables[index].Name
}

func (s *SimulationWorkerStatistics) variableCounterIndex(name *UniqueString) int {
	if name == nil || s.variableCounters == nil {
		panic(NewNullPointerException())
	}
	index := name.VarLoc()
	if index < 0 || index >= len(s.variableCounters) {
		panic(NewArrayIndexOutOfBoundsException(index, len(s.variableCounters)))
	}
	return index
}

func (s *SimulationWorkerStatistics) CollectPostSuccessor(state *TLCStateMut, action *Action, next *TLCStateMut) {
	if s == nil || s.TraceActions == "" {
		return
	}
	if s.sourceTool {
		if state == nil || state.GetAction() == nil {
			panic(NewNullPointerException())
		}
		from := state.GetAction().GetID()
		row := simulationActionStatsRow(s.ActionStats, from)
		if next == nil || next.GetAction() == nil {
			panic(NewNullPointerException())
		}
		to := next.GetAction().GetID()
		simulationActionStatsCell(row, to)
		row[to]++
		return
	}
	if state == nil || next == nil {
		return
	}
	from := actionIDFromStateAction(s, state.GetAction())
	to := actionIDFromStateAction(s, next.GetAction())
	if from >= 0 && from < len(s.ActionStats) && to >= 0 && to < len(s.ActionStats[from]) {
		s.ActionStats[from][to]++
	}
	_ = action
}

func simulationActionStatsRow(stats [][]int64, index int) []int64 {
	if stats == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(stats) {
		panic(NewArrayIndexOutOfBoundsException(index, len(stats)))
	}
	return stats[index]
}

func simulationActionStatsCell(row []int64, index int) int64 {
	if row == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(row) {
		panic(NewArrayIndexOutOfBoundsException(index, len(row)))
	}
	return row[index]
}

func (s *SimulationWorkerStatistics) CollectPreTrace() int64 {
	if s == nil {
		return 0
	}
	return s.NumGenTraces.Add(1)
}

func (s *SimulationWorkerStatistics) CollectNextRetries() {
	if s != nil && (!s.sourceTool || s.Extended) {
		s.NextRetries++
	}
}

func (s *SimulationWorkerStatistics) CollectPostTrace(state *TLCStateMut, maxTraceDepth int) {
	if s == nil {
		return
	}
	if state == nil {
		if s.sourceTool {
			panic(NewNullPointerException())
		}
		return
	}
	if s.sourceTool && s.worker != nil {
		maxTraceDepth = s.worker.MaxTraceDepth
	}
	traceLen := int64(state.Level())
	if (s.sourceTool || maxTraceDepth >= 0) && traceLen > int64(maxTraceDepth) {
		traceLen = int64(maxTraceDepth)
	}
	for {
		old := s.WelfordM2Mean.Load()
		mean := int64(int32(old & 0xffffffff))
		m2 := int64(uint64(old) >> 32)
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
	ensureTLCGetSetUniqueStrings()
	actionCounts := NewInsMap[*UniqueString, Value]()
	for cur := state; cur != nil && !cur.IsInitial(); cur = cur.Predecessor() {
		action := cur.GetAction()
		if s.sourceTool && action == nil {
			panic(NewNullPointerException())
		}
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
	if s.sourceTool {
		for i := 0; i < simulationStateVariableCount(); i++ {
			name := simulationStateVariableName(i)
			counter := s.variableCounters[s.variableCounterIndex(name)]
			if counter == nil {
				panic(NewNullPointerException())
			}
			values.Set(name, NewIntValue(int32(counter.Count())))
		}
		return NewRecordValueFromInsMap(values)
	}
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
	if s.sourceTool {
		for _, action := range s.tool.GetSpecActions() {
			if action == nil {
				panic(NewNullPointerException())
			}
			key := UniqueStringOf(action.GetName())
			values.Set(key, intValueFromInt64(s.ActionCounts.Get(key)))
		}
		return NewRecordValueFromInsMap(values)
	}
	for key := range s.workerActionIDs.All() {
		values.Set(key, intValueFromInt64(s.ActionCounts.Get(key)))
	}
	return NewRecordValueFromInsMap(values)
}

func (s *SimulationWorkerStatistics) traceCount() int64 {
	if s == nil {
		return 0
	}
	if s.sourceTool && s.worker != nil {
		return s.worker.GlobalTrace
	}
	return s.TraceID
}

type SimulationWorker struct {
	traceMu        sync.Mutex
	ID             int
	Tool           *Tool
	Rand           *JavaRandom
	CurState       *TLCStateMut
	InitStates     *StateVec
	LocalValues    []Value
	NamedRegisters *InsMap[*UniqueString, Value]
	ResultQueue    *SimulationWorkerResultQueue
	TraceCnt       int64
	GlobalTrace    int64
	MaxTraceNum    int64
	MaxTraceDepth  int
	CheckDeadlock  bool
	debug          bool
	Mode           SimulationWorkerMode
	TraceFile      string
	LiveCheck      *LiveCheck
	LiveCheck1     *LiveCheck1
	Statistics     *SimulationWorkerStatistics
	Stopped        atomic.Bool
	Halted         atomic.Bool
	done           chan struct{}
	NextStates     *StateVec
	RLAlpha        float64
	RLGamma        float64
	RLReward       float64
	RLEnabledOnly  bool
	RLQ            *InsMap[*Action, *InsMap[int64, float64]]
}

func NewSimulationWorker(id int, tool *Tool, results *SimulationWorkerResultQueue, seed int64, maxTraceDepth int, maxTraceNum int64, traceActions string, checkDeadlock bool, debug bool, traceFile string, liveCheck *LiveCheck, states *atomic.Int64, traces *atomic.Int64, m2Mean *atomic.Int64) *SimulationWorker {
	if results == nil {
		results = NewSimulationWorkerResultQueue()
	}
	worker := &SimulationWorker{
		ID:             id,
		Tool:           tool,
		Rand:           NewJavaRandom(seed),
		NamedRegisters: NewInsMap[*UniqueString, Value](),
		ResultQueue:    results,
		MaxTraceNum:    maxTraceNum,
		MaxTraceDepth:  maxTraceDepth,
		CheckDeadlock:  checkDeadlock,
		debug:          debug,
		Mode:           SimulationWorkerStandard,
		TraceFile:      traceFile,
		LiveCheck:      liveCheck,
		Statistics:     NewSimulationWorkerStatistics(tool, traceActions, states, traces, m2Mean),
		NextStates:     NewStateVec(1),
		RLAlpha:        0.3,
		RLGamma:        0.7,
		RLReward:       -10,
	}
	worker.Statistics.worker = worker
	return worker
}

func (w *SimulationWorker) Stop() {
	if w != nil {
		w.Stopped.Store(true)
		if w.ResultQueue != nil {
			w.ResultQueue.signalInterruptedProducer()
		}
	}
}

func (w *SimulationWorker) Start(initStates *StateVec) {
	if w == nil {
		return
	}
	w.SetInitialStates(initStates)
	w.Stopped.Store(false)
	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		w.Run()
	}()
}

func (w *SimulationWorker) Run() {
	for w != nil {
		if !w.SimulateAndReport() {
			return
		}
	}
}

func (w *SimulationWorker) Join(timeout time.Duration) bool {
	if w == nil || w.done == nil {
		return true
	}
	if timeout <= 0 {
		<-w.done
		return true
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-w.done:
		return true
	case <-timer.C:
		return false
	}
}

func (w *SimulationWorker) IsAlive() bool {
	if w == nil || w.done == nil {
		return false
	}
	select {
	case <-w.done:
		return false
	default:
		return true
	}
}

func (w *SimulationWorker) checkForInterrupt() {
	if w.Stopped.Load() {
		panic(NewInterruptedException())
	}
}

func (w *SimulationWorker) putResult(result SimulationWorkerResult) {
	// LinkedBlockingQueue.put acquires its lock interruptibly, even when empty.
	if err := w.ResultQueue.putInterruptibly(result, &w.Stopped); err != nil {
		panic(err)
	}
}

func (w *SimulationWorker) SimulateAndReport() (keepRunning bool) {
	if w == nil {
		return false
	}
	restoreWorkerID := pushCurrentSimulationWorker(w)
	defer restoreWorkerID()
	defer ResetCurrentState()
	defer func() {
		if recovered := recover(); recovered != nil {
			if _, interrupted := recovered.(*InterruptedException); interrupted {
				w.ResultQueue.Offer(SimulationWorkerOK(w.ID))
				keepRunning = false
				return
			}
			workerErr := NewSimulationWorkerError(NoError, nil, w.GetTrace(w.CurState), recoveredAsError(recovered))
			workerErr.Tool = w.Tool
			w.ResultQueue.Offer(SimulationWorkerFailed(w.ID, workerErr))
			keepRunning = false
		}
	}()
	w.GlobalTrace = w.Statistics.CollectPreTrace()
	w.Statistics.TraceID = w.GlobalTrace
	err := w.SimulateRandomTrace()
	w.TraceCnt++
	if err != nil {
		w.attachToolToError(err)
		w.putResult(SimulationWorkerFailed(w.ID, err))
	}
	if w.TraceCnt >= w.MaxTraceNum {
		w.putResult(SimulationWorkerOK(w.ID))
		return false
	}
	return true
}

func (w *SimulationWorker) attachToolToError(err *SimulationWorkerError) {
	if w != nil && err != nil && err.Tool == nil {
		err.Tool = w.Tool
	}
}

func recoveredAsError(recovered any) error {
	if err, ok := recovered.(error); ok {
		return err
	}
	return fmt.Errorf("%v", recovered)
}

func (w *SimulationWorker) RandomState(states *StateVec) *TLCStateMut {
	if states == nil || states.Size() == 0 {
		return nil
	}
	index := int(math.Floor(w.Rand.NextDouble() * float64(states.Size())))
	return states.At(index)
}

func (w *SimulationWorker) GetNextActionIndex(actions []*Action, curState *TLCStateMut) int {
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
		return NewSimulationWorkerError(ECGeneral, nil, nil, newTLCError(ECGeneral, "simulation worker has no tool"))
	}
	if w.debug {
		return w.SimulateExplorationTrace()
	}
	w.CurState = w.RandomState(w.InitStates)
	SetCurrentState(w.CurState)
	allActions := w.Tool.GetActions()
	for traceIdx := 0; traceIdx < w.MaxTraceDepth; traceIdx++ {
		w.checkForInterrupt()
		w.NextStates.Clear()
		actions, workerErr := w.FilterActions(allActions, w.CurState)
		if workerErr != nil {
			return workerErr
		}
		if len(actions) == 0 {
			if w.CheckDeadlock {
				return NewSimulationWorkerError(ECTLCDeadlockReached, nil, w.GetTrace(w.CurState))
			}
			break
		}
		index := w.GetNextActionIndex(actions, w.CurState)
		step := w.Rand.NextPrime()
		for i := 0; i < len(actions); i++ {
			action := actions[index]
			var err error
			func() {
				restoreCurrentState := PushCurrentState(w.CurState)
				defer restoreCurrentState()
				_, err = w.Tool.GetNextStatesForAction(w.NextStateFunctor(), w.CurState, action)
			}()
			if err != nil {
				if workerErr, ok := err.(*SimulationWorkerError); ok {
					return workerErr
				}
				panic(err)
			}
			if !w.NextStates.IsEmpty() {
				break
			}
			index = w.GetNextActionAltIndex(index, step, actions, w.CurState)
		}
		if w.NextStates.IsEmpty() {
			if w.CheckDeadlock {
				return NewSimulationWorkerError(ECTLCDeadlockReached, nil, w.GetTrace(w.CurState))
			}
			break
		}
		next := w.RandomState(w.NextStates)
		if _, err := next.ExecCallable(); err != nil {
			panic(err)
		}
		w.Statistics.CollectPostSuccessor(w.CurState, next.GetAction(), next)
		w.CurState = next
		SetCurrentState(w.CurState)
	}
	w.checkForInterrupt()
	if workerErr := w.CheckLivenessTrace(); workerErr != nil {
		return workerErr
	}
	w.Statistics.CollectPostTrace(w.CurState, w.MaxTraceDepth)
	if w.TraceFile != "" {
		if err := w.WriteTraceFile(); err != nil {
			panic(err)
		}
	}
	if workerErr := w.PostTrace(w.CurState); workerErr != nil {
		return workerErr
	}
	return nil
}

func (w *SimulationWorker) CheckLivenessTrace() *SimulationWorkerError {
	if w == nil || w.Tool == nil {
		return nil
	}
	if w.LiveCheck1 != nil {
		if err := w.LiveCheck1.CheckTrace(w.Tool.NoDebug(), func() *StateVec { return w.GetTrace(w.CurState) }); err != nil {
			panic(err)
		}
		return nil
	}
	if w.LiveCheck != nil {
		if err := w.LiveCheck.CheckTrace(w.Tool.NoDebug(), func() *StateVec { return w.GetTrace(w.CurState) }); err != nil {
			panic(err)
		}
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

func (w *SimulationWorker) AddGeneratedSuccessor(curState *TLCStateMut, action *Action, succ *TLCStateMut) *SimulationWorkerError {
	if succ == nil {
		return nil
	}
	if action != nil && CoverageActionEnabled() {
		action.CM.IncInvocations()
	}
	succ.SetPredecessor(curState).SetAction(action)
	if !w.Tool.IsGoodState(succ) {
		return NewSimulationWorkerError(ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(w.Tool, action, succ), w.GetTrace(succ))
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
		panic(err)
	}
	if inModel {
		inActions, err := w.Tool.IsInActions(curState, succ)
		if err != nil {
			panic(err)
		}
		inModel = inActions
	}
	if inModel {
		if action != nil && CoverageActionEnabled() {
			action.CM.IncSecondary()
		}
		w.NextStates.Add(succ)
	}
	return nil
}

func (w *SimulationWorker) CheckInvariants(state *TLCStateMut) (failure *SimulationWorkerError) {
	i := 0
	defer func() {
		if recovered := recover(); recovered != nil {
			err, ok := recovered.(error)
			if !ok || isJavaError(err) {
				panic(recovered)
			}
			if workerError, ok := err.(*SimulationWorkerError); ok {
				failure = workerError
				return
			}
			failure = newSimulationWorkerErrorNullable(ECTLCInvariantEvaluationFailed, []*string{javaString(w.Tool.propertyNameAt(w.Tool.GetInvNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
	}()
	for ; i < len(w.Tool.requireActionArray(w.Tool.GetInvariants())); i++ {
		invariant := w.Tool.requireActionArray(w.Tool.GetInvariants())[i]
		valid, err := w.Tool.IsValidState(invariant, state)
		if err != nil {
			return newSimulationWorkerErrorNullable(ECTLCInvariantEvaluationFailed, []*string{javaString(w.Tool.propertyNameAt(w.Tool.GetInvNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
		if !valid {
			return NewSimulationWorkerError(ECTLCInvariantViolatedBehavior, []string{w.Tool.propertyNameAt(w.Tool.GetInvNames(), i)}, w.GetTrace(state))
		}
	}
	return nil
}

func (w *SimulationWorker) CheckImpliedActions(state *TLCStateMut) (failure *SimulationWorkerError) {
	i := 0
	defer func() {
		if recovered := recover(); recovered != nil {
			err, ok := recovered.(error)
			if !ok || isJavaError(err) {
				panic(recovered)
			}
			if workerError, ok := err.(*SimulationWorkerError); ok {
				failure = workerError
				return
			}
			failure = newSimulationWorkerErrorNullable(ECTLCActionPropertyEvaluationFailed, []*string{javaString(w.Tool.propertyNameAt(w.Tool.GetImpliedActNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
	}()
	for ; i < len(w.Tool.requireActionArray(w.Tool.GetImpliedActions())); i++ {
		action := w.Tool.requireActionArray(w.Tool.GetImpliedActions())[i]
		valid, err := w.Tool.IsValidTransition(action, w.CurState, state)
		if err != nil {
			return newSimulationWorkerErrorNullable(ECTLCActionPropertyEvaluationFailed, []*string{javaString(w.Tool.propertyNameAt(w.Tool.GetImpliedActNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
		if !valid {
			return NewSimulationWorkerError(ECTLCActionPropertyViolatedBehavior, []string{w.Tool.propertyNameAt(w.Tool.GetImpliedActNames(), i)}, w.GetTrace(state))
		}
	}
	return nil
}

func (w *SimulationWorker) GetTrace(state TLCPredecessorState) *StateVec {
	w.traceMu.Lock()
	defer w.traceMu.Unlock()
	if mutable, ok := state.(*TLCStateMut); ok && mutable == nil {
		state = nil
	}
	stack := make([]TLCPredecessorState, 0)
	for cur := state; cur != nil; cur = cur.TracePredecessor() {
		pred := cur.TracePredecessor()
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
		trace.ElementAt(i).(TLCPredecessorState).SetTracePredecessor(trace.ElementAt(i - 1).(TLCPredecessorState))
	}
	if !trace.Empty() {
		if !trace.ElementAt(0).(TLCPredecessorState).IsInitial() {
			panic(NewAssertionError())
		}
		for i := 0; i < trace.Size(); i++ {
			if trace.ElementAt(i).(TLCPredecessorState).Level() != i+1 {
				panic(NewAssertionError())
			}
		}
		if trace.ElementAt(trace.Size()-1).(TLCPredecessorState).Level() != trace.Size() {
			panic(NewAssertionError())
		}
	}
	return trace
}

func (w *SimulationWorker) GetUncompressedTrace(state TLCPredecessorState) *StateVec {
	w.traceMu.Lock()
	defer w.traceMu.Unlock()
	if mutable, ok := state.(*TLCStateMut); ok && mutable == nil {
		state = nil
	}
	stack := make([]TLCPredecessorState, 0)
	for cur := state; cur != nil; cur = cur.TracePredecessor() {
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

func (w *SimulationWorker) GetLocalValue(index int) Value {
	if w == nil || index < 0 || index >= len(w.LocalValues) {
		return nil
	}
	return w.LocalValues[index]
}

func (w *SimulationWorker) SetLocalValue(index int, value Value) {
	if w == nil || index < 0 {
		return
	}
	for len(w.LocalValues) <= index {
		w.LocalValues = append(w.LocalValues, nil)
	}
	w.LocalValues[index] = value
}

func (w *SimulationWorker) GetNamedRegister(name *UniqueString) Value {
	if w == nil || w.NamedRegisters == nil {
		return nil
	}
	return w.NamedRegisters.Get(name)
}

func (w *SimulationWorker) SetNamedRegister(name *UniqueString, value Value) {
	if w == nil || name == nil {
		return
	}
	if w.NamedRegisters == nil {
		w.NamedRegisters = NewInsMap[*UniqueString, Value]()
	}
	w.NamedRegisters.Set(name, value)
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

func newSimulationWorkerErrorNullable(code int, params []*string, trace *StateVec) *SimulationWorkerError {
	copied := copyNullableMessageParameters(params)
	failure := NewSimulationWorkerError(code, messageParameterStrings(copied), trace)
	failure.NullableParams = copied
	return failure
}

func (e *SimulationWorkerError) GetMessage() *string {
	if e == nil {
		return nil
	}
	if e.NullableParams != nil {
		return javaString(formatNullableMessage(e.Code, e.NullableParams))
	}
	return javaString(formatMessage(e.Code, e.Params))
}
