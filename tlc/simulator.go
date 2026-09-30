package tlc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Simulator struct {
	Tool          *Tool
	CheckDeadlock bool
	TraceDepth    int
	TraceNum      int64
	TraceFile     string
	TraceActions  string
	Rand          *JavaRandom
	Seed          int64
	Aril          int64
	Config        Value
	ResultQueue   chan SimulationWorkerResult
	Workers       []*SimulationWorker
	WorkerMode    SimulationWorkerMode
	LiveCheck     *LiveCheck
	NumGenStates  atomic.Int64
	NumGenTraces  atomic.Int64
	WelfordM2Mean atomic.Int64

	StatesGenerated int64
	TracesGenerated int64
	DisabledRetries int64
	Stopped         bool
}

type SimulatorOption func(*Simulator)

func WithSimulatorTraceFile(traceFile string) SimulatorOption {
	return func(s *Simulator) {
		s.TraceFile = traceFile
	}
}

func WithSimulatorTraceActions(traceActions string) SimulatorOption {
	return func(s *Simulator) {
		s.TraceActions = traceActions
	}
}

func WithSimulatorSchedule(schedule SimulationSchedule) SimulatorOption {
	return func(s *Simulator) {
		switch schedule {
		case SimulationScheduleRL:
			s.WorkerMode = SimulationWorkerRL
		case SimulationScheduleRLAction:
			s.WorkerMode = SimulationWorkerRLAction
		default:
			s.WorkerMode = SimulationWorkerStandard
		}
	}
}

func WithSimulatorLiveCheck(liveCheck *LiveCheck) SimulatorOption {
	return func(s *Simulator) {
		s.LiveCheck = liveCheck
	}
}

func NewSimulator(tool *Tool, deadlock bool, traceDepth int, traceNum int64, seed int64, opts ...SimulatorOption) *Simulator {
	SetTLCStateTool(tool)
	if traceDepth < 0 {
		traceDepth = int(^uint(0) >> 1)
	}
	if traceNum <= 0 {
		traceNum = int64(^uint64(0) >> 1)
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	checkDeadlock := deadlock
	if tool != nil && tool.GetModelConfig() != nil {
		checkDeadlock = deadlock && tool.GetModelConfig().GetCheckDeadlock()
	}
	simulator := &Simulator{
		Tool:          tool,
		CheckDeadlock: checkDeadlock,
		TraceDepth:    traceDepth,
		TraceNum:      traceNum,
		Seed:          seed,
		Rand:          NewJavaRandom(seed),
		ResultQueue:   make(chan SimulationWorkerResult, max(NumWorkers(), 1)*2),
	}
	simulator.WorkerMode = simulator.selectWorkerMode()
	for _, opt := range opts {
		if opt != nil {
			opt(simulator)
		}
	}
	workerCount := NumWorkers()
	if workerCount < 1 {
		workerCount = 1
	}
	for i := 0; i < workerCount; i++ {
		simulator.Workers = append(simulator.Workers, simulator.newSimulationWorker(i))
	}
	simulator.Config = simulator.createConfig()
	SetSimulator(simulator)
	return simulator
}

func (s *Simulator) Simulate() (int, error) {
	if s.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "simulator has no tool")
	}
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(s.Tool)
		defer ReportCoverage(s.Tool, TLCStartTime())
	}
	if result := s.Tool.CheckAssumptions(); result != NoError {
		return result, nil
	}
	initStates, result, err := s.initialStates()
	if err != nil || result != NoError {
		return result, err
	}
	if initStates.IsEmpty() {
		return ECTLCNoStatesSatisfyingInit, nil
	}
	initStates.DeepNormalize()
	s.Aril = s.Rand.Aril()
	workerResult := s.simulate(initStates)
	s.StatesGenerated = s.NumGenStates.Load()
	s.TracesGenerated = s.NumGenTraces.Load()
	code := s.postSimulationErrorCode(workerResult)
	if workerResult.IsError() {
		return code, workerResult.Error.Err
	}
	return code, nil
}

func (s *Simulator) Stop() {
	if s != nil {
		s.Stopped = true
		for _, worker := range s.Workers {
			worker.Stop()
		}
	}
}

func (s *Simulator) GetLocalValue(idx int) Value {
	if s == nil || idx < 0 {
		return nil
	}
	if worker := s.currentWorker(); worker != nil {
		return worker.GetLocalValue(idx)
	}
	if len(s.Workers) > 0 && s.Workers[0] != nil {
		return s.Workers[0].GetLocalValue(idx)
	}
	return nil
}

func (s *Simulator) SetAllValues(idx int, value Value) {
	if s == nil || idx < 0 {
		return
	}
	for _, worker := range s.Workers {
		worker.SetLocalValue(idx, value)
	}
}

func (s *Simulator) GetAllValues() Value {
	if s == nil || len(s.Workers) == 0 || s.Workers[0] == nil {
		return EmptyFcn
	}
	localValues := s.Workers[0].LocalValues
	domain := make([]Value, 0, len(localValues))
	values := make([]Value, 0, len(localValues))
	for idx, value := range localValues {
		if value == nil {
			continue
		}
		workerValues := make([]Value, len(s.Workers))
		for i, worker := range s.Workers {
			if worker != nil {
				workerValues[i] = worker.GetLocalValue(idx)
			}
		}
		domain = append(domain, NewIntValue(int32(idx)))
		values = append(values, NewTupleValue(workerValues))
	}
	return NewFcnRcdValue(domain, values, false)
}

func (s *Simulator) GetLocalNamedValue(key *UniqueString) Value {
	if s == nil || key == nil {
		return nil
	}
	if worker := s.currentWorker(); worker != nil {
		return worker.GetNamedRegister(key)
	}
	if len(s.Workers) > 0 && s.Workers[0] != nil {
		return s.Workers[0].GetNamedRegister(key)
	}
	return nil
}

func (s *Simulator) SetAllNamedValues(key *UniqueString, value Value) {
	if s == nil || key == nil {
		return
	}
	for _, worker := range s.Workers {
		worker.SetNamedRegister(key, value)
	}
}

func (s *Simulator) GetAllNamedRegisterValues() Value {
	if s == nil || len(s.Workers) == 0 || s.Workers[0] == nil || s.Workers[0].NamedRegisters == nil {
		return EmptyFcn
	}
	domain := make([]Value, 0, s.Workers[0].NamedRegisters.Len())
	values := make([]Value, 0, s.Workers[0].NamedRegisters.Len())
	for key := range s.Workers[0].NamedRegisters.All() {
		workerValues := make([]Value, len(s.Workers))
		for i, worker := range s.Workers {
			if worker != nil {
				workerValues[i] = worker.GetNamedRegister(key)
			}
		}
		domain = append(domain, NewStringValueFromUnique(key))
		values = append(values, NewTupleValue(workerValues))
	}
	return NewFcnRcdValue(domain, values, false)
}

func (s *Simulator) GetAllNamedValues(key *UniqueString) []Value {
	if s == nil || key == nil {
		return nil
	}
	values := make([]Value, 0, len(s.Workers))
	for _, worker := range s.Workers {
		if worker == nil {
			values = append(values, nil)
			continue
		}
		values = append(values, worker.GetNamedRegister(key))
	}
	return values
}

func (s *Simulator) GetStatistics(state *TLCStateMut) Value {
	stats := s.currentWorkerStatistics()
	m2AndMean := s.WelfordM2Mean.Load()
	mean := int64(m2AndMean & 0xffffffff)
	m2 := uint64(m2AndMean) >> 32
	traces := s.NumGenTraces.Load()
	if traces == 0 && s.TracesGenerated != 0 {
		traces = s.TracesGenerated
	}
	states := s.NumGenStates.Load()
	if states == 0 && s.StatesGenerated != 0 {
		states = s.StatesGenerated
	}
	names := []*UniqueString{
		tlcGetTraces,
		tlcGetDuration,
		tlcGetGenerated,
		tlcGetBehavior,
		tlcGetWorker,
		tlcGetDistinct,
		tlcGetDistinctValues,
		tlcGetRetries,
		tlcSpecActions,
		tlcGetLevelMean,
		tlcGetLevelVar,
	}
	values := []Value{
		intValueFromInt64(traces),
		intValueFromDurationSince(TLCStartTime()),
		intValueFromInt64(states),
		stats.GetTraceStatistics(state),
		NewIntValue(int32(currentWorkerIDOrZero())),
		stats.GetDistinctStates(),
		stats.GetDistinctValues(),
		stats.GetNextRetries(),
		stats.GetActions(),
		intValueFromInt64(mean),
		intValueFromInt64(int64(math.Round(float64(m2) / (float64(traces) + 1)))),
	}
	return NewRecordValue(names, values, false)
}

func (s *Simulator) GetTrace(state *TLCStateMut) *StateVec {
	if s == nil || len(s.Workers) == 0 {
		return NewStateVec(0)
	}
	workerID := currentWorkerIDOrZero()
	if workerID < 0 || workerID >= len(s.Workers) || s.Workers[workerID] == nil {
		workerID = 0
	}
	return s.Workers[workerID].GetTrace(state)
}

func (s *Simulator) GetUncompressedTrace(state *TLCStateMut) *StateVec {
	if s == nil || len(s.Workers) == 0 {
		return NewStateVec(0)
	}
	workerID := currentWorkerIDOrZero()
	if workerID < 0 || workerID >= len(s.Workers) || s.Workers[workerID] == nil {
		workerID = 0
	}
	return s.Workers[workerID].GetUncompressedTrace(state)
}

func (s *Simulator) GetConfig() Value {
	if s == nil {
		return EmptyRecord
	}
	if s.Config != nil {
		return s.Config
	}
	return s.createConfig()
}

func (s *Simulator) createConfig() Value {
	if s == nil {
		return EmptyRecord
	}
	names := []*UniqueString{
		tlcGetMode,
		tlcGetDepth,
		tlcGetTraces,
		tlcGetDeadlock,
		tlcGetSeed,
		tlcGetAril,
		tlcGetWorker,
		tlcGetInstall,
		tlcGetSched,
	}
	depth := int32(s.TraceDepth)
	if s.TraceDepth == int(^uint(0)>>1) {
		depth = -1
	}
	workerCount := len(s.Workers)
	if workerCount < 1 {
		workerCount = 1
	}
	values := []Value{
		NewStringValue("simulate"),
		NewIntValue(depth),
		NewIntValue(int32(int64(workerCount) * s.TraceNum)),
		NewBoolValue(s.CheckDeadlock),
		NewStringValue(fmt.Sprintf("%d", s.Seed)),
		NewStringValue(fmt.Sprintf("%d", s.Aril)),
		NewIntValue(int32(workerCount)),
		NewStringValue(TLCInstallLocation()),
		NewStringValue(s.schedulerName()),
	}
	return NewRecordValue(names, values, false)
}

func (s *Simulator) currentWorkerStatistics() *SimulationWorkerStatistics {
	if s == nil || len(s.Workers) == 0 {
		return NewSimulationWorkerStatistics(nil, "", nil, nil, nil)
	}
	if worker := s.currentWorker(); worker != nil {
		return worker.Statistics
	}
	if s.Workers[0] != nil {
		return s.Workers[0].Statistics
	}
	return NewSimulationWorkerStatistics(nil, "", nil, nil, nil)
}

func (s *Simulator) currentWorker() *SimulationWorker {
	if s == nil || len(s.Workers) == 0 {
		return nil
	}
	workerID, ok := CurrentWorkerID()
	if !ok {
		return nil
	}
	if workerID < 0 || workerID >= len(s.Workers) || s.Workers[workerID] == nil {
		return nil
	}
	return s.Workers[workerID]
}

func currentWorkerIDOrZero() int {
	if id, ok := CurrentWorkerID(); ok {
		return id
	}
	return 0
}

func (s *Simulator) schedulerName() string {
	if s == nil {
		return "random"
	}
	switch s.WorkerMode {
	case SimulationWorkerRL:
		return "rl"
	case SimulationWorkerRLAction:
		return "rlaction"
	default:
		return "random"
	}
}

func (s *Simulator) initialStates() (*StateVec, int, error) {
	all := NewStateVec(0)
	err := s.Tool.GetInitStates(NewStateFunctor(func(state *TLCStateMut) (any, error) {
		all.Add(state)
		return all, nil
	}))
	if err != nil {
		return nil, ECGeneral, err
	}
	s.StatesGenerated += int64(all.Size())
	s.NumGenStates.Add(int64(all.Size()))
	filtered := NewStateVec(all.Size())
	for i := 0; i < all.Size(); i++ {
		state := all.At(i)
		if !s.Tool.IsGoodState(state) {
			return nil, ECTLCStateNotCompletelySpecifiedInitial, nil
		}
		if result, err := s.checkInvariants(state, true); result != NoError || err != nil {
			if result == ECTLCInvariantViolatedInitial {
				s.Tool.CheckPostConditionWithCounterExample(NewCounterExampleFromInitialState(state))
			}
			return nil, result, err
		}
		inModel, err := s.Tool.IsInModel(state)
		if err != nil {
			return nil, ECGeneral, err
		}
		if inModel {
			filtered.Add(state)
		}
	}
	if all.Size() > 0 && filtered.IsEmpty() {
		return nil, ECTLCNoStatesSatisfyingInitAndConstraint, nil
	}
	return filtered, NoError, nil
}

func (s *Simulator) simulate(initStates *StateVec) SimulationWorkerResult {
	if len(s.Workers) == 0 {
		s.Workers = append(s.Workers, s.newSimulationWorker(0))
	}
	running := make(map[int]bool, len(s.Workers))
	runningCount := 0
	for i, worker := range s.Workers {
		worker.Start(initStates)
		running[i] = true
		runningCount++
	}
	var result SimulationWorkerResult
	for runningCount > 0 {
		result = <-s.ResultQueue
		if result.WorkerID == -1 {
			break
		}
		if result.IsError() {
			if s.simulationErrorStops(result.Error) {
				s.Stop()
				return result
			}
			continue
		}
		if running[result.WorkerID] {
			delete(running, result.WorkerID)
			runningCount--
		}
	}
	return result
}

func (s *Simulator) isNonContinuableError(code int) bool {
	return code == ECTLCInvariantEvaluationFailed ||
		code == ECTLCActionPropertyEvaluationFailed ||
		code == ECTLCStateNotCompletelySpecifiedNext
}

func (s *Simulator) simulationErrorStops(err *SimulationWorkerError) bool {
	if err == nil {
		return false
	}
	if err.Err != nil {
		var live *LiveException
		if errors.As(err.Err, &live) && live != nil {
			err.Code = live.ErrorCode
		} else if err.Code == NoError {
			err.Code = ECGeneral
		}
		return true
	}
	if s.isNonContinuableError(err.Code) {
		return true
	}
	if !Globals.Continuation {
		return true
	}
	if err.Code == NoError {
		err.Code = ECGeneral
	}
	return false
}

func (s *Simulator) postSimulationErrorCode(workerResult SimulationWorkerResult) int {
	errorCode := NoError
	if workerResult.IsError() {
		errorCode = workerResult.Error.Code
		if errorCode == 0 {
			errorCode = ECGeneral
		}
	}
	pcErrorCode := NoError
	if s != nil && s.Tool != nil {
		if workerResult.IsError() && workerResult.Error.HasTrace() {
			pcErrorCode = s.Tool.CheckPostConditionWithCounterExample(workerResult.Error.GetCounterExample())
		} else {
			pcErrorCode = s.Tool.CheckPostCondition()
		}
	}
	if ExitStatusForErrorCode(pcErrorCode) > ExitStatusForErrorCode(errorCode) {
		return pcErrorCode
	}
	return errorCode
}

func (s *Simulator) newSimulationWorker(id int) *SimulationWorker {
	debugger := s.Tool != nil && s.Tool.IsDebugger()
	tool := s.Tool
	if id != 0 && tool != nil {
		tool = tool.NoDebug()
	}
	liveCheck := s.LiveCheck
	if liveCheck == nil {
		liveCheck = NewNoOpLiveCheck(tool, "")
	}
	worker := NewSimulationWorker(
		id,
		tool,
		s.ResultQueue,
		s.Rand.NextLong(),
		s.TraceDepth,
		s.TraceNum,
		s.TraceActions,
		s.CheckDeadlock,
		debugger,
		s.TraceFile,
		liveCheck,
		&s.NumGenStates,
		&s.NumGenTraces,
		&s.WelfordM2Mean,
	)
	worker.SetMode(s.WorkerMode)
	worker.RLAlpha = simulatorPropertyFloat("tlc2.tool.Simulator.rl.alpha", "TLAGO_SIMULATOR_RL_ALPHA", 0.3)
	worker.RLGamma = simulatorPropertyFloat("tlc2.tool.Simulator.rl.gamma", "TLAGO_SIMULATOR_RL_GAMMA", 0.7)
	worker.RLReward = simulatorPropertyFloat("tlc2.tool.Simulator.rl.reward", "TLAGO_SIMULATOR_RL_REWARD", -10)
	worker.RLEnabledOnly = simulatorPropertyBool("tlc2.tool.Simulator.rl.enabledOnly", "TLAGO_SIMULATOR_RL_ENABLED_ONLY")
	return worker
}

func (s *Simulator) selectWorkerMode() SimulationWorkerMode {
	if simulatorPropertyBool("tlc2.tool.Simulator.rl", "TLAGO_SIMULATOR_RL") {
		return SimulationWorkerRL
	}
	if simulatorPropertyBool("tlc2.tool.Simulator.rlaction", "TLAGO_SIMULATOR_RL_ACTION") {
		return SimulationWorkerRLAction
	}
	return SimulationWorkerStandard
}

func simulatorPropertyBool(name string, aliases ...string) bool {
	for _, key := range append([]string{name}, aliases...) {
		if value, ok := os.LookupEnv(key); ok {
			return strings.EqualFold(value, "true")
		}
	}
	return false
}

func simulatorPropertyFloat(name string, alias string, fallback float64) float64 {
	for _, key := range []string{name, alias} {
		if value, ok := os.LookupEnv(key); ok {
			if parsed, err := strconv.ParseFloat(strings.TrimSuffix(value, "d"), 64); err == nil {
				return parsed
			}
		}
	}
	return fallback
}

func (s *Simulator) randomSuccessor(cur *TLCStateMut) (*TLCStateMut, int, error) {
	successors := make([]*TLCStateMut, 0)
	actions := make([]*Action, 0)
	restoreCurrentState := PushCurrentState(cur)
	defer restoreCurrentState()
	for _, action := range s.Tool.GetActions() {
		nextStates, err := s.Tool.GetNextStates(action, cur)
		if err != nil {
			return nil, ECGeneral, err
		}
		if nextStates == nil {
			continue
		}
		s.StatesGenerated += int64(nextStates.Size())
		for i := 0; i < nextStates.Size(); i++ {
			succ := nextStates.At(i)
			if !s.Tool.IsGoodState(succ) {
				return nil, ECTLCStateNotCompletelySpecifiedNext, nil
			}
			succ.SetPredecessor(cur).SetAction(action)
			inModel, err := s.Tool.IsInModel(succ)
			if err != nil {
				return nil, ECGeneral, err
			}
			if inModel {
				inActions, err := s.Tool.IsInActions(cur, succ)
				if err != nil {
					return nil, ECGeneral, err
				}
				inModel = inActions
			}
			if !inModel {
				continue
			}
			if result, err := s.checkInvariants(succ, false); result != NoError || err != nil {
				return nil, result, err
			}
			successors = append(successors, succ)
			actions = append(actions, action)
		}
	}
	if len(successors) == 0 {
		if s.CheckDeadlock {
			return nil, ECTLCDeadlockReached, nil
		}
		return nil, NoError, nil
	}
	idx := int(s.Rand.NextIntN(int32(len(successors))))
	return successors[idx].SetPredecessor(cur).SetAction(actions[idx]), NoError, nil
}

func (s *Simulator) checkInvariants(state *TLCStateMut, initial bool) (int, error) {
	for _, invariant := range s.Tool.GetInvariants() {
		valid, err := s.Tool.IsValidState(invariant, state)
		if err != nil {
			if initial {
				return ECTLCInvariantEvaluationFailed, err
			}
			return ECTLCInvariantEvaluationFailed, err
		}
		if !valid {
			if initial {
				return ECTLCInvariantViolatedInitial, nil
			}
			return ECTLCInvariantViolatedBehavior, nil
		}
	}
	return NoError, nil
}
