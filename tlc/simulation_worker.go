package tlc

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var simulationWorkerCoverageStatics struct {
	once   sync.Once
	action bool
}

// InitializeSimulationWorkerStatics represents a fresh Java classloader for
// SimulationWorker's coverage flag.
func InitializeSimulationWorkerStatics() {
	simulationWorkerCoverageStatics.once = sync.Once{}
	simulationWorkerCoverageStatics.action = false
}

func initializeSimulationWorkerStatics() {
	simulationWorkerCoverageStatics.once.Do(func() {
		simulationWorkerCoverageStatics.action = CoverageActionEnabled()
	})
}

type SimulationWorkerError struct {
	*InvariantViolatedException
	Code           int
	Params         []string
	NullableParams []*string
	StateTrace     *StateVec
	Err            error
	Tool           *Tool

	ownersCaptured      bool
	paramsOwner         []string
	nullableParamsOwner []*string
	stateTraceOwner     *StateVec
	errOwner            error
}

func NewSimulationWorkerError(code int, params []string, stateTrace *StateVec, exception ...error) *SimulationWorkerError {
	if len(exception) > 1 {
		panic(NewIllegalArgumentException("NewSimulationWorkerError accepts at most one exception"))
	}
	code = int(int32(code))
	failure := &SimulationWorkerError{
		InvariantViolatedException: NewInvariantViolatedException(),
		Code:                       code, Params: params, StateTrace: stateTrace,
		ownersCaptured: true, paramsOwner: params, stateTraceOwner: stateTrace,
	}
	if len(exception) > 0 {
		failure.Err, failure.errOwner = exception[0], exception[0]
	}
	return failure
}

func (e *SimulationWorkerError) ownedParams() []string {
	if e.ownersCaptured {
		return e.paramsOwner
	}
	return e.Params
}

func (e *SimulationWorkerError) ownedNullableParams() []*string {
	if e.ownersCaptured {
		return e.nullableParamsOwner
	}
	return e.NullableParams
}

func (e *SimulationWorkerError) ownedStateTrace() *StateVec {
	if e.ownersCaptured {
		return e.stateTraceOwner
	}
	return e.StateTrace
}

func (e *SimulationWorkerError) ownedErr() error {
	if e.ownersCaptured {
		return e.errOwner
	}
	return e.Err
}

func (e *SimulationWorkerError) Error() string {
	if e == nil {
		return ""
	}
	return *e.GetMessage()
}

func (e *SimulationWorkerError) HasTrace() bool {
	if e == nil {
		panic(NewNullPointerException())
	}
	trace := e.ownedStateTrace()
	return trace != nil && trace.Size() > 0
}

func (e *SimulationWorkerError) GetCounterExample() *CounterExample {
	if e == nil {
		panic(NewNullPointerException())
	}
	if failure, ok := e.ownedErr().(*LiveCounterExampleException); ok && failure != nil {
		return failure.ownedCounterExample()
	}
	return NewCounterExampleFromTrace(e.GetTrace())
}

func (e *SimulationWorkerError) GetTrace() []*TLCStateInfo {
	if e == nil || e.ownedStateTrace() == nil {
		panic(NewNullPointerException())
	}
	stateTrace := e.ownedStateTrace()
	trace := make([]*TLCStateInfo, 0, stateTrace.Size())
	for i := 0; i < stateTrace.Size(); i++ {
		state := stateTrace.At(i)
		successor := state
		if i+1 < stateTrace.Size() {
			successor = stateTrace.At(i + 1)
		}
		if state == nil {
			panic(NewNullPointerException())
		}
		info := NewTLCStateInfo(state, state.GetAction())
		if e.Tool == nil {
			panic(NewNullPointerException())
		}
		alias, err := e.Tool.EvalAliasInfoPair(info, successor)
		if err != nil {
			panic(err)
		}
		trace = append(trace, alias)
	}
	return trace
}

func (e *SimulationWorkerError) TraceInfo() []*TLCStateInfo { return e.GetTrace() }

type SimulationWorkerResult struct {
	workerID    int
	workerError *SimulationWorkerError
}

func SimulationWorkerOK(workerID int) SimulationWorkerResult {
	return SimulationWorkerResult{workerID: int(int32(workerID))}
}

func SimulationWorkerFailed(workerID int, err *SimulationWorkerError) SimulationWorkerResult {
	if err == nil {
		panic(NewNullPointerException())
	}
	return SimulationWorkerResult{workerID: int(int32(workerID)), workerError: err}
}

func (r SimulationWorkerResult) IsError() bool {
	return r.workerError != nil
}

func (r SimulationWorkerResult) Error() *SimulationWorkerError {
	if r.workerError == nil {
		panic(NewNoSuchElementException("No value present"))
	}
	return r.workerError
}

func (r SimulationWorkerResult) WorkerID() int { return r.workerID }

type SimulationWorkerMode int

const (
	SimulationWorkerStandard SimulationWorkerMode = iota
	SimulationWorkerRL
	SimulationWorkerRLAction
)

type SimulationWorkerStatistics struct {
	TraceActions   *string // Only nil disables the action matrix; empty still enables it.
	NumGenStates   *atomic.Int64
	NumGenTraces   *atomic.Int64
	WelfordM2Mean  *atomic.Int64
	ActionStats    [][]int64
	NextRetries    int64
	DistinctStates int64
	DistinctValues *InsMap[*UniqueString, int64]
	ActionCounts   *InsMap[*UniqueString, int64]
	Extended       bool
	TraceID        int64
	tool           *Tool
	distinctStates *CountDistinct

	// Captured once; source access indexes it with current variable locations.
	variableCounters []*CountDistinct
	worker           *SimulationWorker

	ownersCaptured      bool
	traceActionsOwner   *string
	numGenStatesOwner   *atomic.Int64
	numGenTracesOwner   *atomic.Int64
	welfordM2MeanOwner  *atomic.Int64
	actionStatsOwner    [][]int64
	distinctValuesOwner *InsMap[*UniqueString, int64]
	actionCountsOwner   *InsMap[*UniqueString, int64]
	extendedOwner       bool
}

func (s *SimulationWorkerStatistics) captureOwners() {
	if s.TraceActions != nil {
		traceActions := *s.TraceActions
		s.traceActionsOwner = &traceActions
	}
	s.numGenStatesOwner = s.NumGenStates
	s.numGenTracesOwner = s.NumGenTraces
	s.welfordM2MeanOwner = s.WelfordM2Mean
	s.actionStatsOwner = s.ActionStats
	s.distinctValuesOwner = s.DistinctValues
	s.actionCountsOwner = s.ActionCounts
	s.extendedOwner = s.Extended
	s.ownersCaptured = true
}

func (s *SimulationWorkerStatistics) ownedTraceActions() *string {
	if s.ownersCaptured {
		return s.traceActionsOwner
	}
	return s.TraceActions
}

func (s *SimulationWorkerStatistics) ownedNumGenStates() *atomic.Int64 {
	if s.ownersCaptured {
		return s.numGenStatesOwner
	}
	return s.NumGenStates
}

func (s *SimulationWorkerStatistics) ownedNumGenTraces() *atomic.Int64 {
	if s.ownersCaptured {
		return s.numGenTracesOwner
	}
	return s.NumGenTraces
}

func (s *SimulationWorkerStatistics) ownedWelfordM2Mean() *atomic.Int64 {
	if s.ownersCaptured {
		return s.welfordM2MeanOwner
	}
	return s.WelfordM2Mean
}

func (s *SimulationWorkerStatistics) ownedActionStats() [][]int64 {
	if s.ownersCaptured {
		return s.actionStatsOwner
	}
	return s.ActionStats
}

func (s *SimulationWorkerStatistics) ownedDistinctValues() *InsMap[*UniqueString, int64] {
	if s.ownersCaptured {
		return s.distinctValuesOwner
	}
	return s.DistinctValues
}

func (s *SimulationWorkerStatistics) ownedActionCounts() *InsMap[*UniqueString, int64] {
	if s.ownersCaptured {
		return s.actionCountsOwner
	}
	return s.ActionCounts
}

func (s *SimulationWorkerStatistics) ownedExtended() bool {
	if s.ownersCaptured {
		return s.extendedOwner
	}
	return s.Extended
}

func NewSimulationWorkerStatistics(tool *Tool, traceActions *string, states *atomic.Int64, traces *atomic.Int64, m2Mean *atomic.Int64) *SimulationWorkerStatistics {
	initializeSimulatorStatics()
	size := 1
	if traceActions != nil {
		if tool == nil {
			panic(NewNullPointerException())
		}
		size = len(tool.GetSpecActions())
	}
	stats := &SimulationWorkerStatistics{
		TraceActions:   traceActions,
		NumGenStates:   states,
		NumGenTraces:   traces,
		WelfordM2Mean:  m2Mean,
		ActionStats:    make([][]int64, size),
		DistinctValues: NewInsMap[*UniqueString, int64](),
		ActionCounts:   newInsMapAllowNilKeys[*UniqueString, int64](),
		tool:           tool,
		Extended:       simulatorStatics.extendedStatistics,
	}
	if stats.ownedExtended() {
		stats.distinctStates = newSimulationCountDistinct(8)
		stats.variableCounters = make([]*CountDistinct, simulationStateVariableCount())
		for i := 0; i < simulationStateVariableCount(); i++ {
			index := stats.variableCounterIndex(simulationStateVariableName(i))
			stats.variableCounters[index] = newSimulationCountDistinct(10)
		}
	}
	for i := range stats.ownedActionStats() {
		stats.ownedActionStats()[i] = make([]int64, size)
	}
	stats.captureOwners()
	return stats
}

func (s *SimulationWorkerStatistics) CollectPreSuccessor(state *TLCStateMut, action *Action, next *TLCStateMut) {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.ownedNumGenStates() == nil {
		panic(NewNullPointerException())
	}
	s.ownedNumGenStates().Add(1)
	if s.ownedExtended() {
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
			s.ownedDistinctValues().Set(name, counter.Count())
		}
		if next == nil || s.distinctStates == nil {
			panic(NewNullPointerException())
		}
		s.distinctStates.AddState(next)
		s.DistinctStates = s.distinctStates.Count()
		if state == nil || state.GetAction() == nil {
			panic(NewNullPointerException())
		}
		key := state.GetAction().GetNameUnique()
		s.ownedActionCounts().Set(key, int64(int32(s.ownedActionCounts().Get(key))+1))
		return
	}
}

func newSimulationCountDistinct(bits int) *CountDistinct {
	initializeSimulatorStatics()
	if simulatorStatics.extendedStatisticsNaive {
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.ownedTraceActions() == nil {
		return
	}
	if state == nil || state.GetAction() == nil {
		panic(NewNullPointerException())
	}
	from := state.GetAction().GetID()
	row := simulationActionStatsRow(s.ownedActionStats(), from)
	if next == nil || next.GetAction() == nil {
		panic(NewNullPointerException())
	}
	to := next.GetAction().GetID()
	simulationActionStatsCell(row, to)
	row[to]++
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
		panic(NewNullPointerException())
	}
	if s.ownedNumGenTraces() == nil {
		panic(NewNullPointerException())
	}
	return s.ownedNumGenTraces().Add(1)
}

func (s *SimulationWorkerStatistics) CollectNextRetries() {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.ownedExtended() {
		s.NextRetries++
	}
}

func (s *SimulationWorkerStatistics) CollectPostTrace(state *TLCStateMut, maxTraceDepthOverride ...int) {
	if s == nil || state == nil {
		panic(NewNullPointerException())
	}
	if len(maxTraceDepthOverride) > 1 {
		panic(NewIllegalArgumentException("CollectPostTrace accepts at most one trace depth"))
	}
	maxTraceDepth := 0
	if s.worker != nil {
		maxTraceDepth = s.worker.ownedMaxTraceDepth()
	} else if len(maxTraceDepthOverride) == 1 {
		maxTraceDepth = maxTraceDepthOverride[0]
	} else {
		panic(NewNullPointerException())
	}
	traceLen := int64(state.Level())
	if traceLen > int64(maxTraceDepth) {
		traceLen = int64(maxTraceDepth)
	}
	if s.ownedWelfordM2Mean() == nil {
		panic(NewNullPointerException())
	}
	for {
		old := s.ownedWelfordM2Mean().Load()
		mean := int64(int32(old & 0xffffffff))
		m2 := int64(uint64(old) >> 32)
		if s.ownedNumGenTraces() == nil {
			panic(NewNullPointerException())
		}
		count := s.ownedNumGenTraces().Load()
		if count < 1 {
			count = 1
		}
		delta := traceLen - mean
		nextMean := mean + delta/count
		nextM2 := m2 + delta*(traceLen-nextMean)
		next := (nextM2 << 32) | (nextMean & 0xffffffff)
		if s.ownedWelfordM2Mean().CompareAndSwap(old, next) {
			return
		}
	}
}

func (s *SimulationWorkerStatistics) GetNextRetries() Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	if !s.ownedExtended() {
		return NewIntValue(-1)
	}
	return intValueFromInt64(s.NextRetries)
}

func (s *SimulationWorkerStatistics) GetTraceStatistics(state *TLCStateMut) Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	ensureTLCGetSetUniqueStrings()
	actionCounts := newInsMapAllowNilKeys[*UniqueString, Value]()
	for cur := state; cur != nil && !cur.IsInitial(); cur = cur.Predecessor() {
		action := cur.GetAction()
		if action == nil {
			panic(NewNullPointerException())
		}
		actionKey := action.GetNameUnique()
		count := int32(1)
		if old, ok := actionCounts.Get(actionKey).(*IntValue); ok {
			count = old.ownedVal() + 1
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if !s.ownedExtended() {
		return NewIntValue(-1)
	}
	if s.distinctStates != nil {
		return intValueFromInt64(s.distinctStates.Count())
	}
	return intValueFromInt64(s.DistinctStates)
}

func (s *SimulationWorkerStatistics) GetDistinctValues() Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	if !s.ownedExtended() {
		return NewIntValue(-1)
	}
	values := NewInsMap[*UniqueString, Value]()
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

func (s *SimulationWorkerStatistics) GetActions() Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	if !s.ownedExtended() {
		return EmptyRecord
	}
	if s.tool == nil {
		panic(NewNullPointerException())
	}
	values := newInsMapAllowNilKeys[*UniqueString, Value]()
	for _, action := range s.tool.GetSpecActions() {
		if action == nil {
			panic(NewNullPointerException())
		}
		key := action.GetNameUnique()
		values.Set(key, intValueFromInt64(s.ownedActionCounts().Get(key)))
	}
	return NewRecordValueFromInsMap(values)
}

func (s *SimulationWorkerStatistics) traceCount() int64 {
	if s == nil {
		return 0
	}
	if s.worker != nil {
		return s.worker.GlobalTrace
	}
	return s.TraceID
}

type SimulationWorker struct {
	traceMu        sync.Mutex
	registerMu     sync.Mutex
	ID             int
	Tool           *Tool
	Rand           *JavaRandom
	CurState       *TLCStateMut
	InitStates     *StateVec
	LocalValues    []Value
	NamedRegisters *NamedRegisterMap
	ResultQueue    *SimulationWorkerResultQueue
	TraceCnt       int64
	GlobalTrace    int64
	MaxTraceNum    int64
	MaxTraceDepth  int
	CheckDeadlock  bool
	debug          bool
	Mode           SimulationWorkerMode
	TraceFile      *string
	LiveCheck      *LiveCheck
	LiveCheck1     *LiveCheck1
	Statistics     *SimulationWorkerStatistics
	Stopped        atomic.Bool
	Halted         atomic.Bool
	started        atomic.Bool
	done           chan struct{}
	NextStates     *StateVec
	RLAlpha        float64
	RLGamma        float64
	RLReward       float64
	RLEnabledOnly  bool
	RLQ            *InsMap[*Action, *InsMap[int64, float64]]
	actionCoverage bool

	configurationCaptured bool
	resultQueueOwner      *SimulationWorkerResultQueue
	toolOwner             *Tool
	randOwner             *JavaRandom
	maxTraceNumOwner      int64
	maxTraceDepthOwner    int
	checkDeadlockOwner    bool
	traceFileOwner        *string
	liveCheckOwner        *LiveCheck
	liveCheck1Owner       *LiveCheck1
	statisticsOwner       *SimulationWorkerStatistics
	nextStatesOwner       *StateVec
	idOwner               int
	namedRegistersOwner   *NamedRegisterMap
	rlQCaptured           bool
	rlQOwner              *InsMap[*Action, *InsMap[int64, float64]]
	rlPropertiesCaptured  bool
	rlPropertiesOwner     rlSimulationWorkerProperties
	modeCaptured          bool
	modeOwner             SimulationWorkerMode
}

func (w *SimulationWorker) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *SimulationWorker) ownedResultQueue() *SimulationWorkerResultQueue {
	if w.configurationCaptured {
		return w.resultQueueOwner
	}
	return w.ResultQueue
}

func (w *SimulationWorker) ownedTool() *Tool {
	if w.configurationCaptured {
		return w.toolOwner
	}
	return w.Tool
}

func (w *SimulationWorker) ownedRand() *JavaRandom {
	if w.configurationCaptured {
		return w.randOwner
	}
	return w.Rand
}

func (w *SimulationWorker) ownedMaxTraceNum() int64 {
	if w.configurationCaptured {
		return w.maxTraceNumOwner
	}
	return w.MaxTraceNum
}

func (w *SimulationWorker) ownedMaxTraceDepth() int {
	if w.configurationCaptured {
		return w.maxTraceDepthOwner
	}
	return w.MaxTraceDepth
}

func (w *SimulationWorker) ownedCheckDeadlock() bool {
	if w.configurationCaptured {
		return w.checkDeadlockOwner
	}
	return w.CheckDeadlock
}

func (w *SimulationWorker) ownedTraceFile() *string {
	if w.configurationCaptured {
		return w.traceFileOwner
	}
	return w.TraceFile
}

func (w *SimulationWorker) ownedLiveCheck() *LiveCheck {
	if w.configurationCaptured {
		return w.liveCheckOwner
	}
	return w.LiveCheck
}

func (w *SimulationWorker) ownedLiveCheck1() *LiveCheck1 {
	if w.configurationCaptured {
		return w.liveCheck1Owner
	}
	return w.LiveCheck1
}

func (w *SimulationWorker) ownedStatistics() *SimulationWorkerStatistics {
	if w.configurationCaptured {
		return w.statisticsOwner
	}
	return w.Statistics
}

func (w *SimulationWorker) ownedNextStates() *StateVec {
	if w.configurationCaptured {
		return w.nextStatesOwner
	}
	return w.NextStates
}

func (w *SimulationWorker) ownedID() int {
	if w.configurationCaptured {
		return w.idOwner
	}
	return w.ID
}

func (w *SimulationWorker) ownedNamedRegisters() *NamedRegisterMap {
	if w.configurationCaptured {
		return w.namedRegistersOwner
	}
	return w.NamedRegisters
}

// setLiveCheck1 completes the Go adapter for Java's final ILiveCheck field.
// The source stores either checker implementation in one constructor field.
func (w *SimulationWorker) setLiveCheck1(liveCheck *LiveCheck1) {
	w.LiveCheck1 = liveCheck
	if w.configurationCaptured && liveCheck != nil {
		w.liveCheckOwner = nil
		w.liveCheck1Owner = liveCheck
	}
}

// NewSimulationWorkerWithDefaultStatistics implements the source short constructor,
// which allocates its own counters. The full constructor retains supplied counters.
func NewSimulationWorkerWithDefaultStatistics(id int, tool *Tool, results *SimulationWorkerResultQueue, seed int64, maxTraceDepth int, maxTraceNum int64, checkDeadlock bool, debug bool, traceFile *string, liveCheck *LiveCheck) *SimulationWorker {
	return NewSimulationWorker(id, tool, results, seed, maxTraceDepth, maxTraceNum, nil, checkDeadlock, debug, traceFile, liveCheck, &atomic.Int64{}, &atomic.Int64{}, &atomic.Int64{})
}

func NewSimulationWorker(id int, tool *Tool, results *SimulationWorkerResultQueue, seed int64, maxTraceDepth int, maxTraceNum int64, traceActions *string, checkDeadlock bool, debug bool, traceFile *string, liveCheck *LiveCheck, states *atomic.Int64, traces *atomic.Int64, m2Mean *atomic.Int64) *SimulationWorker {
	id = int(int32(id))
	maxTraceDepth = int(int32(maxTraceDepth))
	initializeSimulationWorkerStatics()
	worker := &SimulationWorker{
		ID:             id,
		Tool:           tool,
		Rand:           NewJavaRandom(seed),
		LocalValues:    make([]Value, 4),
		NamedRegisters: NewNamedRegisterMap(),
		ResultQueue:    results,
		MaxTraceNum:    maxTraceNum,
		MaxTraceDepth:  maxTraceDepth,
		CheckDeadlock:  checkDeadlock,
		debug:          debug,
		Mode:           SimulationWorkerStandard,
		TraceFile:      traceFile,
		LiveCheck:      liveCheck,
		Statistics:     NewSimulationWorkerStatistics(tool, traceActions, states, traces, m2Mean),
		done:           make(chan struct{}),
		NextStates:     NewStateVec(1),
		RLAlpha:        0.3,
		RLGamma:        0.7,
		RLReward:       -10,
		actionCoverage: simulationWorkerCoverageStatics.action,
	}
	worker.Statistics.worker = worker
	worker.resultQueueOwner = worker.ResultQueue
	worker.toolOwner = worker.Tool
	worker.randOwner = worker.Rand
	worker.maxTraceNumOwner = worker.MaxTraceNum
	worker.maxTraceDepthOwner = worker.MaxTraceDepth
	worker.checkDeadlockOwner = worker.CheckDeadlock
	if worker.TraceFile != nil {
		traceFile := *worker.TraceFile
		worker.traceFileOwner = &traceFile
	}
	worker.liveCheckOwner = worker.LiveCheck
	worker.liveCheck1Owner = worker.LiveCheck1
	worker.statisticsOwner = worker.Statistics
	worker.nextStatesOwner = worker.NextStates
	worker.idOwner = worker.ID
	worker.namedRegistersOwner = worker.NamedRegisters
	worker.configurationCaptured = true
	return worker
}

func (w *SimulationWorker) Stop() {
	w.require()
	w.SetStopped()
	if queue := w.ownedResultQueue(); queue != nil {
		queue.signalInterruptedProducer()
	}
}

func (w *SimulationWorker) SetStopped() {
	w.require()
	w.Stopped.Store(true)
}

func (w *SimulationWorker) Halt() bool {
	w.require()
	old := w.Halted.Load()
	w.Halted.Store(true)
	return old
}

func (w *SimulationWorker) Start(initStates *StateVec) {
	w.require()
	w.SetInitialStates(initStates)
	if !w.started.CompareAndSwap(false, true) {
		panic(NewIllegalThreadStateException())
	}
	go func() {
		defer close(w.done)
		w.Run()
	}()
}

func (w *SimulationWorker) Run() {
	w.require()
	for {
		if !w.SimulateAndReport() {
			return
		}
	}
}

func (w *SimulationWorker) Join(timeout time.Duration) bool {
	w.require()
	if timeout < 0 {
		panic(NewIllegalArgumentException("timeout value is negative"))
	}
	if !w.started.Load() {
		return true
	}
	if timeout == 0 {
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
	w.require()
	if !w.started.Load() {
		return false
	}
	select {
	case <-w.done:
		return false
	default:
		return true
	}
}

func (w *SimulationWorker) CheckForInterrupt() {
	w.require()
	if w.Stopped.Load() {
		panic(NewInterruptedException())
	}
}

func (w *SimulationWorker) putResult(result SimulationWorkerResult) {
	w.require()
	// LinkedBlockingQueue.put acquires its lock interruptibly, even when empty.
	if err := w.ownedResultQueue().putInterruptibly(result, &w.Stopped); err != nil {
		panic(err)
	}
}

// Worker errors capture their tool immediately, like Java's inner error class.
func (w *SimulationWorker) newError(code int, params []string, trace *StateVec, exception ...error) *SimulationWorkerError {
	w.require()
	failure := NewSimulationWorkerError(code, params, trace, exception...)
	failure.Tool = w.ownedTool()
	return failure
}

func (w *SimulationWorker) newNullableError(code int, params []*string, trace *StateVec) *SimulationWorkerError {
	w.require()
	failure := newSimulationWorkerErrorNullable(code, params, trace)
	failure.Tool = w.ownedTool()
	return failure
}

func (w *SimulationWorker) SimulateAndReport() (keepRunning bool) {
	w.require()
	restoreWorkerID := pushCurrentSimulationWorker(w)
	defer restoreWorkerID()
	defer ResetCurrentState()
	defer func() {
		if recovered := recover(); recovered != nil {
			exception, ok := recovered.(error)
			if !ok || isJavaError(exception) {
				panic(recovered)
			}
			if _, interrupted := exception.(*InterruptedException); interrupted {
				w.ownedResultQueue().Offer(SimulationWorkerOK(w.ownedID()))
				keepRunning = false
				return
			}
			workerErr := w.newError(NoError, nil, w.GetTrace(w.CurState), exception)
			workerErr.Tool = w.ownedTool()
			w.ownedResultQueue().Offer(SimulationWorkerFailed(w.ownedID(), workerErr))
			keepRunning = false
		}
	}()
	w.GlobalTrace = w.ownedStatistics().CollectPreTrace()
	w.ownedStatistics().TraceID = w.GlobalTrace
	err := w.SimulateRandomTrace()
	w.TraceCnt++
	if err != nil {
		w.attachToolToError(err)
		w.putResult(SimulationWorkerFailed(w.ownedID(), err))
	}
	if w.TraceCnt >= w.ownedMaxTraceNum() {
		w.putResult(SimulationWorkerOK(w.ownedID()))
		return false
	}
	return true
}

func (w *SimulationWorker) attachToolToError(err *SimulationWorkerError) {
	w.require()
	if err != nil && err.Tool == nil {
		err.Tool = w.ownedTool()
	}
}

func (w *SimulationWorker) RandomState(states *StateVec) *TLCStateMut {
	w.require()
	return w.RandomStateWithRNG(w.ownedRand(), states)
}

func (w *SimulationWorker) RandomStateWithRNG(rng *JavaRandom, states *StateVec) *TLCStateMut {
	w.require()
	if states.Size() == 0 {
		return nil
	}
	if rng == nil {
		panic(NewNullPointerException())
	}
	index := int(math.Floor(rng.NextDouble() * float64(states.Size())))
	return states.At(index)
}

func (w *SimulationWorker) GetNextActionIndex(actions []*Action, curState *TLCStateMut) int {
	w.require()
	return w.GetNextActionIndexWithRNG(w.ownedRand(), actions, curState)
}

func (w *SimulationWorker) GetNextActionIndexWithRNG(_ *JavaRandom, actions []*Action, curState *TLCStateMut) int {
	w.require()
	if w.IsRLMode() {
		return w.GetRLNextActionIndex(actions, curState)
	}
	// Java evaluates the RNG draw before dereferencing the action array.
	draw := w.ownedRand().NextDouble()
	if actions == nil {
		panic(NewNullPointerException())
	}
	return int(math.Floor(draw * float64(len(actions))))
}

func (w *SimulationWorker) GetNextActionAltIndex(index int, p int, actions []*Action, curState *TLCStateMut) int {
	w.require()
	if w.IsRLMode() {
		w.UpdateRLDisabledAction(index, actions, curState)
	}
	w.ownedStatistics().CollectNextRetries()
	if actions == nil {
		panic(NewNullPointerException())
	}
	if len(actions) == 0 {
		panic(NewArithmeticException("/ by zero"))
	}
	return int(int32(index)+int32(p)) % len(actions)
}

func (w *SimulationWorker) SimulateRandomTrace() *SimulationWorkerError {
	w.require()
	if w.debug {
		return w.SimulateExplorationTrace()
	}
	w.CurState = w.RandomState(w.InitStates)
	SetCurrentState(w.CurState)
	if w.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	allActions := w.ownedTool().GetActions()
	for traceIdx := 0; traceIdx < w.ownedMaxTraceDepth(); traceIdx++ {
		w.CheckForInterrupt()
		w.ownedNextStates().Clear()
		actions, workerErr := w.FilterActions(allActions, w.CurState)
		if workerErr != nil {
			return workerErr
		}
		index := w.GetNextActionIndex(actions, w.CurState)
		step := w.ownedRand().NextPrime()
		for i := 0; i < len(actions); i++ {
			action := actions[index]
			var err error
			func() {
				restoreCurrentState := PushCurrentState(w.CurState)
				defer restoreCurrentState()
				_, err = w.ownedTool().GetNextStatesForAction(w.NextStateFunctor(), w.CurState, action)
			}()
			if err != nil {
				if workerErr, ok := err.(*SimulationWorkerError); ok {
					return workerErr
				}
				panic(err)
			}
			if !w.ownedNextStates().IsEmpty() {
				break
			}
			index = w.GetNextActionAltIndex(index, step, actions, w.CurState)
		}
		if w.ownedNextStates().IsEmpty() {
			if w.ownedCheckDeadlock() {
				return w.newError(ECTLCDeadlockReached, nil, w.GetTrace(w.CurState))
			}
			break
		}
		next := w.RandomState(w.ownedNextStates())
		if _, err := next.ExecCallable(); err != nil {
			panic(err)
		}
		w.ownedStatistics().CollectPostSuccessor(w.CurState, actions[index], next)
		w.CurState = next
		SetCurrentState(w.CurState)
	}
	w.CheckForInterrupt()
	if workerErr := w.CheckLivenessTrace(); workerErr != nil {
		return workerErr
	}
	w.ownedStatistics().CollectPostTrace(w.CurState)
	if w.ownedTraceFile() != nil {
		if err := w.WriteTraceFile(); err != nil {
			panic(err)
		}
	}
	w.PostTrace(w.CurState)
	return nil
}

func (w *SimulationWorker) CheckLivenessTrace() *SimulationWorkerError {
	w.require()
	if w.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedLiveCheck1() != nil {
		if err := w.ownedLiveCheck1().CheckTrace(w.ownedTool().NoDebug(), func() *StateVec { return w.GetTrace(w.CurState) }); err != nil {
			panic(err)
		}
		return nil
	}
	if w.ownedLiveCheck() == nil {
		panic(NewNullPointerException())
	}
	if err := w.ownedLiveCheck().CheckTrace(w.ownedTool().NoDebug(), func() *StateVec { return w.GetTrace(w.CurState) }); err != nil {
		panic(err)
	}
	return nil
}

func (w *SimulationWorker) FilterActions(actions []*Action, curState *TLCStateMut) ([]*Action, *SimulationWorkerError) {
	w.require()
	if w.IsRLMode() && w.ownedRLProperties().enabledOnly {
		return w.FilterEnabledRLActions(actions, curState)
	}
	return actions, nil
}

func (w *SimulationWorker) PostTrace(finalState *TLCStateMut) bool {
	w.require()
	if w.ownedMode() == SimulationWorkerRL {
		return w.PostRLTrace(finalState)
	}
	return true
}

func (w *SimulationWorker) SetElement(state *TLCStateMut) any {
	w.require()
	w.ownedNextStates().Clear()
	w.ownedNextStates().Add(state)
	return w
}

func (w *SimulationWorker) AddElement(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
	w.require()
	if workerErr := w.AddGeneratedSuccessor(curState, action, succState); workerErr != nil {
		return nil, workerErr
	}
	return w, nil
}

func (w *SimulationWorker) HasStates() bool {
	w.require()
	if !IsProbabilistic() {
		panic(NewAssertionError())
	}
	return !w.ownedNextStates().IsEmpty()
}

func (w *SimulationWorker) GetStates() *SetOfStates {
	w.require()
	return NewSetOfStatesFromVec(w.ownedNextStates())
}

func (w *SimulationWorker) AddGeneratedSuccessor(curState *TLCStateMut, action *Action, succ *TLCStateMut) *SimulationWorkerError {
	w.require()
	if w.actionCoverage {
		if action == nil {
			panic(NewNullPointerException())
		}
		action.CM.IncInvocations()
	}
	succ.SetPredecessor(curState).SetAction(action)
	if w.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	if !w.ownedTool().IsGoodState(succ) {
		return w.newError(ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(w.ownedTool(), action, succ), w.GetTrace(succ))
	}
	w.ownedStatistics().CollectPreSuccessor(curState, action, succ)
	if workerErr := w.CheckInvariants(succ); workerErr != nil {
		return workerErr
	}
	if workerErr := w.CheckImpliedActions(succ); workerErr != nil {
		return workerErr
	}
	inModel, err := w.ownedTool().IsInModel(succ)
	if err != nil {
		panic(err)
	}
	if inModel {
		inActions, err := w.ownedTool().IsInActions(curState, succ)
		if err != nil {
			panic(err)
		}
		inModel = inActions
	}
	if inModel {
		if w.actionCoverage {
			if action == nil {
				panic(NewNullPointerException())
			}
			action.CM.IncSecondary()
		}
		w.ownedNextStates().Add(succ)
	}
	return nil
}

func (w *SimulationWorker) CheckInvariants(state *TLCStateMut) (failure *SimulationWorkerError) {
	w.require()
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
			failure = w.newNullableError(ECTLCInvariantEvaluationFailed, []*string{javaString(w.ownedTool().propertyNameAt(w.ownedTool().GetInvNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
	}()
	for ; i < len(w.ownedTool().requireActionArray(w.ownedTool().GetInvariants())); i++ {
		invariant := w.ownedTool().requireActionArray(w.ownedTool().GetInvariants())[i]
		valid, err := w.ownedTool().IsValidState(invariant, state)
		if err != nil {
			return w.newNullableError(ECTLCInvariantEvaluationFailed, []*string{javaString(w.ownedTool().propertyNameAt(w.ownedTool().GetInvNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
		if !valid {
			return w.newError(ECTLCInvariantViolatedBehavior, []string{w.ownedTool().propertyNameAt(w.ownedTool().GetInvNames(), i)}, w.GetTrace(state))
		}
	}
	return nil
}

func (w *SimulationWorker) CheckImpliedActions(state *TLCStateMut) (failure *SimulationWorkerError) {
	w.require()
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
			failure = w.newNullableError(ECTLCActionPropertyEvaluationFailed, []*string{javaString(w.ownedTool().propertyNameAt(w.ownedTool().GetImpliedActNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
	}()
	for ; i < len(w.ownedTool().requireActionArray(w.ownedTool().GetImpliedActions())); i++ {
		action := w.ownedTool().requireActionArray(w.ownedTool().GetImpliedActions())[i]
		valid, err := w.ownedTool().IsValidTransition(action, w.CurState, state)
		if err != nil {
			return w.newNullableError(ECTLCActionPropertyEvaluationFailed, []*string{javaString(w.ownedTool().propertyNameAt(w.ownedTool().GetImpliedActNames(), i)), javaThrowableDetailMessage(err)}, w.GetTrace(state))
		}
		if !valid {
			return w.newError(ECTLCActionPropertyViolatedBehavior, []string{w.ownedTool().propertyNameAt(w.ownedTool().GetImpliedActNames(), i)}, w.GetTrace(state))
		}
	}
	return nil
}

func (w *SimulationWorker) GetTrace(state TLCPredecessorState) *StateVec {
	w.require()
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
	w.require()
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
	w.require()
	w.InitStates = initStates
}

func (w *SimulationWorker) GetTraceCnt() int64 {
	w.require()
	return w.TraceCnt + 1
}

func (w *SimulationWorker) GetRNG() *JavaRandom {
	w.require()
	return w.ownedRand()
}

func (w *SimulationWorker) WriteTraceFile() error {
	w.require()
	name := fmt.Sprintf("%s_%d_%d", *w.ownedTraceFile(), w.ownedID(), w.TraceCnt)
	file, err := openPoolFile(name, true)
	if err != nil {
		ToolIOPrintln("Error: Unable to write to file " + name)
		return err
	}
	// PrintWriter(OutputStream) adds its own character buffer and a
	// default-charset encoder while suppressing output errors after open.
	writer := &statePrintWriter{buffer: NewJavaBufferedWriter(file)}
	lineSeparator := javaFormatLineSeparator()
	_, _ = fmt.Fprintf(writer, "---------------- MODULE %s -----------------%s", name, lineSeparator)
	trace := w.GetTrace(w.CurState)
	for i := 0; i < trace.Size(); i++ {
		state := trace.At(i)
		if state.HasAction() {
			_, _ = fmt.Fprintf(writer, "\\* %s%s", state.GetAction().GetLocation(), lineSeparator)
		}
		_, _ = fmt.Fprintf(writer, "STATE_%d == %s%s\n%s", i+1, lineSeparator, state.String(), lineSeparator)
	}
	_, _ = writer.WriteString("=================================================" + lineSeparator)
	writer.Flush()
	writer.closed = true
	_ = file.Close()
	return nil
}

func newSimulationWorkerErrorNullable(code int, params []*string, trace *StateVec) *SimulationWorkerError {
	failure := NewSimulationWorkerError(code, messageParameterStrings(params), trace)
	failure.NullableParams, failure.nullableParamsOwner = params, params
	return failure
}

func (e *SimulationWorkerError) GetMessage() *string {
	e.require()
	if params := e.ownedNullableParams(); params != nil {
		return javaString(formatNullableMessage(e.Code, params))
	}
	return javaString(formatMessage(e.Code, e.ownedParams()))
}

func (e *SimulationWorkerError) GetCause() error {
	e.require()
	return e.InvariantViolatedException.GetCause()
}

func (e *SimulationWorkerError) SetKnown() bool {
	e.require()
	return e.InvariantViolatedException.SetKnown()
}

func (e *SimulationWorkerError) IsKnown() bool {
	e.require()
	return e.InvariantViolatedException.IsKnown()
}

func (e *SimulationWorkerError) require() {
	if e == nil || e.InvariantViolatedException == nil {
		panic(NewNullPointerException())
	}
}
