package tlc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Simulator struct {
	Tool             *Tool
	CheckDeadlock    bool
	TraceDepth       int
	TraceNum         int64
	TraceFile        string
	TraceActions     string
	MetaDir          string
	StartTime        time.Time
	Rand             *JavaRandom
	Seed             int64
	Aril             int64
	Config           Value
	ResultQueue      chan SimulationWorkerResult
	Workers          []*SimulationWorker
	WorkerMode       SimulationWorkerMode
	LiveCheck        *LiveCheck
	LiveCheckInitErr error
	LiveCheck1Errors atomic.Bool
	NumGenStates     atomic.Int64
	NumGenTraces     atomic.Int64
	WelfordM2Mean    atomic.Int64

	StatesGenerated int64
	TracesGenerated int64
	DisabledRetries int64
	Stopped         bool
}

const simulatorUnboundedTraceDepth = math.MaxInt32

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

func WithSimulatorMetaDir(metadir string) SimulatorOption {
	return func(s *Simulator) {
		s.MetaDir = metadir
	}
}

func WithSimulatorSchedule(schedule SimulationSchedule) SimulatorOption {
	return func(s *Simulator) {
		switch schedule {
		case SimulationScheduleRL:
			s.WorkerMode = SimulationWorkerRL
		case SimulationScheduleRLAction:
			s.WorkerMode = SimulationWorkerRLAction
		}
	}
}

func WithSimulatorAril(aril int64) SimulatorOption {
	return func(s *Simulator) {
		s.Aril = aril
	}
}

func WithSimulatorLiveCheck(liveCheck *LiveCheck) SimulatorOption {
	return func(s *Simulator) {
		s.LiveCheck = liveCheck
	}
}

func NewSimulator(tool *Tool, deadlock bool, traceDepth int, traceNum int64, seed int64, opts ...SimulatorOption) *Simulator {
	if tool != nil {
		tool.SetMode(ModeSimulation)
	}
	SetTLCStateTool(tool)
	if traceDepth == -1 {
		traceDepth = simulatorUnboundedTraceDepth
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
		MetaDir:       "states",
		StartTime:     time.Now(),
		ResultQueue:   make(chan SimulationWorkerResult, max(NumWorkers(), 1)*2),
	}
	simulator.WorkerMode = simulator.selectWorkerMode()
	for _, opt := range opts {
		if opt != nil {
			opt(simulator)
		}
	}
	if simulator.MetaDir == "" {
		simulator.MetaDir = "states"
	}
	simulator.Rand = NewJavaRandom(simulator.Seed)
	if simulator.Aril > 0 {
		simulator.Rand.SetSeedWithAril(simulator.Seed, simulator.Aril)
	}
	simulator.Aril = simulator.Rand.Aril()
	workerCount := NumWorkers()
	if workerCount < 1 {
		workerCount = 1
	}
	for i := 0; i < workerCount; i++ {
		simulator.Workers = append(simulator.Workers, simulator.newSimulationWorker(i))
	}
	simulator.Config = simulator.createConfig()
	scheduleStopAfterFromJavaProperty(simulator.Stop)
	SetSimulator(simulator)
	return simulator
}

func (s *Simulator) Simulate() (int, error) {
	if s.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "simulator has no tool")
	}
	if s.LiveCheckInitErr != nil {
		return ECGeneral, s.LiveCheckInitErr
	}
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(s.Tool)
	}
	if result := s.Tool.CheckAssumptions(); result != NoError {
		return result, nil
	}
	initStates, result, err := s.initialStates()
	if err != nil || result != NoError {
		return result, err
	}
	if initStates.IsEmpty() {
		return PrintError(ECTLCNoStatesSatisfyingInit), nil
	}
	PrintMessage(ECTLCInitGenerated1, fmtInt(initStates.Size()), "")
	initStates.DeepNormalize()
	stopProgress := s.startProgressReporter()
	workerResult := s.simulate(initStates)
	s.StatesGenerated = s.NumGenStates.Load()
	s.TracesGenerated = s.NumGenTraces.Load()
	code := s.postSimulationErrorCode(workerResult)
	stopProgress()
	if code == NoError {
		s.PrintSummary()
	}
	if workerResult.IsError() {
		return code, nil
	}
	return code, nil
}

func (s *Simulator) PrintSummary() {
	if s == nil {
		return
	}
	if CoverageAnyEnabled() {
		ReportCoverage(s.Tool, s.StartTime)
	}
	if err := s.writeActionFlowGraph(); err != nil {
		PrintTLCBug(ECTLCReporterDied)
	}
	if toolMode() {
		PrintMessage(ECTLCProgressSimu,
			fmtInt64(s.NumGenStates.Load()),
			fmtInt64(s.NumGenTraces.Load()),
		)
	}
	PrintMessage(ECTLCStatsSimu,
		fmtInt64(s.NumGenStates.Load()),
		fmtInt64(s.Seed),
		fmtInt64(s.Aril),
	)
}

func (s *Simulator) startProgressReporter() func() {
	if s == nil || s.Tool == nil {
		return func() {}
	}
	interval := ProgressInterval()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		coverageCountdown := periodicCoverageCountdown(interval)
		for {
			select {
			case <-ticker.C:
				switch s.reportSimulationProgress(&coverageCountdown, interval) {
				case simulatorProgressStop:
					s.ResultQueue <- SimulationWorkerOK(-1)
					return
				case simulatorProgressReporterDone:
					return
				}
			case <-stop:
				s.reportSimulationProgress(&coverageCountdown, interval)
				return
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}

type simulatorProgressStatus int

const (
	simulatorProgressContinue simulatorProgressStatus = iota
	simulatorProgressStop
	simulatorProgressReporterDone
)

func (s *Simulator) reportSimulationProgress(coverageCountdown *int, interval time.Duration) simulatorProgressStatus {
	genTrace := s.NumGenTraces.Load()
	m2AndMean := s.WelfordM2Mean.Load()
	mean := int64(m2AndMean & 0xffffffff)
	m2 := uint64(m2AndMean) >> 32
	PrintMessage(ECTLCProgressSimu,
		fmtInt64(s.NumGenStates.Load()),
		fmtInt64(genTrace),
		fmtInt64(mean),
		fmtInt64(int64(math.Round(float64(m2)/(float64(genTrace)+1)))),
		fmtInt64(int64(math.Round(math.Sqrt(float64(m2)/(float64(genTrace)+1))))),
	)
	if coverageCountdown != nil {
		if *coverageCountdown > 1 {
			(*coverageCountdown)--
		} else {
			if CoverageAnyEnabled() {
				ReportCoverage(s.Tool, s.StartTime)
			}
			*coverageCountdown = periodicCoverageCountdown(interval)
		}
	}
	if err := s.writeActionFlowGraph(); err != nil {
		PrintTLCBug(ECTLCReporterDied)
		return simulatorProgressReporterDone
	}
	if s.Tool != nil && s.Tool.Periodic != nil {
		value, err := s.Tool.NoDebug().Eval(s.Tool.Periodic)
		if err != nil {
			PrintTLCBug(ECTLCReporterDied)
			return simulatorProgressReporterDone
		}
		if boolValue, ok := value.(*BoolValue); ok && !boolValue.Val {
			PrintError(ECTLCAssumptionFalse, SemanticString(s.Tool.Periodic))
			return simulatorProgressStop
		}
	}
	return simulatorProgressContinue
}

func (s *Simulator) Stop() {
	if s != nil {
		s.Stopped = true
		for _, worker := range s.Workers {
			worker.Stop()
		}
	}
}

func (s *Simulator) shutdownAndJoinWorkers(workers []*SimulationWorker) {
	for _, worker := range workers {
		if worker != nil {
			worker.Stop()
		}
	}
	for _, worker := range workers {
		if worker != nil {
			worker.Join(10 * time.Second)
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
	states := s.NumGenStates.Load()
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
		intValueFromDurationSince(s.StartTime),
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
	if s.TraceDepth == simulatorUnboundedTraceDepth {
		depth = -1
	}
	workerCount := len(s.Workers)
	if workerCount < 1 {
		workerCount = 1
	}
	mode := "simulate"
	if toolProbabilisticEnabled() {
		mode = "generate"
	}
	values := []Value{
		NewStringValue(mode),
		NewIntValue(depth),
		NewIntValue(int32(int64(workerCount) * s.TraceNum)),
		NewBoolValue(s.CheckDeadlock),
		NewStringValue(fmt.Sprintf("%d", s.Seed)),
		NewStringValue("0"),
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
		s.printInitialStateException(nil, err)
		return nil, ECGeneral, err
	}
	s.StatesGenerated += int64(all.Size())
	s.NumGenStates.Add(int64(all.Size()))
	PrintMessage(ECTLCComputingInitProgress, fmtInt64(s.NumGenStates.Load()))
	filtered := NewStateVec(all.Size())
	for i := 0; i < all.Size(); i++ {
		state := all.At(i)
		if !s.Tool.IsGoodState(state) {
			PrintError(ECTLCStateNotCompletelySpecifiedInitial, state.String())
			return nil, ECTLCStateNotCompletelySpecifiedInitial, nil
		}
		for j, invariant := range s.Tool.GetInvariants() {
			valid, err := s.Tool.IsValidState(invariant, state)
			if err != nil {
				code := s.printInitialStateException(state, err)
				return nil, code, err
			}
			if !valid {
				alias := s.Tool.EvalAlias(state, state)
				result := PrintError(ECTLCInvariantViolatedInitial, nameAt(s.Tool.GetInvNames(), j), alias.String())
				s.Tool.CheckPostConditionWithCounterExample(NewCounterExampleFromInitialState(state))
				return nil, result, nil
			}
		}
		inModel, err := s.Tool.IsInModel(state)
		if err != nil {
			code := s.printInitialStateException(state, err)
			return nil, code, err
		}
		if inModel {
			filtered.Add(state)
		}
	}
	if all.Size() > 0 && filtered.IsEmpty() {
		return nil, PrintError(ECTLCNoStatesSatisfyingInitAndConstraint), nil
	}
	return filtered, NoError, nil
}

func (s *Simulator) printInitialStateException(state *TLCStateMut, err error) int {
	message := ""
	if err != nil {
		message = err.Error()
	}
	if message == "" {
		message = fmt.Sprintf("%T", err)
	}
	code := ECGeneral
	if state != nil {
		code = PrintError(ECTLCInitialState, message, state.String())
	} else {
		params := generalErrorParams("", err)
		if params == nil {
			code = PrintError(ECGeneral)
		} else {
			code = PrintError(ECGeneral, params...)
		}
	}
	s.PrintSummary()
	return code
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
			s.printSimulationWorkerError(result.Error)
			if s.simulationErrorStops(result.Error) {
				break
			}
			continue
		}
		if running[result.WorkerID] {
			delete(running, result.WorkerID)
			runningCount--
		}
	}
	s.shutdownAndJoinWorkers(s.Workers)
	return result
}

func (s *Simulator) printSimulationWorkerError(err *SimulationWorkerError) {
	if err == nil {
		return
	}
	if err.Err != nil {
		var liveCounterExample *LiveCounterExampleException
		if errors.As(err.Err, &liveCounterExample) && liveCounterExample != nil && liveCounterExample.LiveException != nil {
			err.Code = liveCounterExample.LiveException.ErrorCode
			s.PrintSummary()
			return
		}
		var live *LiveException
		if errors.As(err.Err, &live) && live != nil {
			err.Code = live.ErrorCode
			s.PrintSummary()
			return
		}
		classified := false
		var eval *EvalException
		if errors.As(err.Err, &eval) && eval != nil {
			classified = true
			if err.Code == NoError {
				err.Code = eval.ErrorCode
			}
			if len(err.Params) == 0 {
				err.Params = eval.GetParameters()
			}
		}
		var tlcErr *TLCError
		if errors.As(err.Err, &tlcErr) && tlcErr != nil {
			classified = true
			if err.Code == NoError {
				err.Code = tlcErr.Code
			}
		}
		if err.Code == NoError {
			err.Code = ECGeneral
		}
		if len(err.Params) == 0 {
			if !classified && err.Code == ECGeneral {
				err.Params = generalErrorParams("", err.Err)
			} else {
				err.Params = []string{err.Err.Error()}
			}
		}
		s.printBehavior(err.Code, err.Params, err.StateTrace)
		return
	}
	if err.Code == NoError {
		err.Code = ECGeneral
	}
	s.printBehavior(err.Code, err.Params, err.StateTrace)
}

func (s *Simulator) printBehavior(errorCode int, params []string, stateTrace *StateVec) {
	PrintError(errorCode, params...)
	s.printBehaviorTrace(stateTrace)
	s.PrintSummary()
}

func (s *Simulator) printBehaviorTrace(stateTrace *StateVec) {
	if stateTrace == nil || stateTrace.Size() == 0 {
		return
	}
	if s.TraceDepth == simulatorUnboundedTraceDepth {
		PrintMessage(ECTLCErrorState)
		PrintStandaloneErrorState(stateTrace.Last())
		return
	}
	PrintError(ECTLCBehaviorUpToThisPoint)
	var lastState *TLCStateMut
	omitted := 0
	for i := 0; i < stateTrace.Size(); i++ {
		curState := stateTrace.At(i)
		sucState := stateTrace.At(min(i+1, stateTrace.Size()-1))
		info := NewTLCStateInfo(curState)
		if s.Tool != nil {
			if aliased, err := s.Tool.EvalAliasInfoPair(info, sucState); err == nil && aliased != nil {
				info = aliased
			}
		}
		if lastState != nil && curState != nil && printDiffsOnly() && curState.FingerPrint() == lastState.FingerPrint() {
			omitted++
			lastState = curState
			continue
		}
		level := i + 1
		if curState != nil {
			level = curState.Level()
		}
		PrintInvariantViolationStateTraceState(info, lastState, level, i+1 == stateTrace.Size())
		lastState = curState
	}
	if omitted > 0 {
		PrintMessage(ECGeneral, fmt.Sprintf("difftrace requested: Shortened behavior by omitting finite stuttering (%d states), which is an artifact of simulation mode.\n", omitted))
	}
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
	var liveCheck1 *LiveCheck1
	if liveCheck == nil {
		if simulatorPropertyBool("tlc2.tool.Simulator.experimentalLiveness", "TLAGO_SIMULATOR_EXPERIMENTAL_LIVENESS") {
			liveCheck = s.newWorkerLiveCheck(tool, id)
		} else {
			liveCheck1 = s.newWorkerLiveCheck1(tool, id)
		}
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
	worker.LiveCheck1 = liveCheck1
	workerMode := s.WorkerMode
	if debugger {
		workerMode = SimulationWorkerStandard
	}
	worker.SetMode(workerMode)
	worker.RLAlpha = simulatorPropertyFloat("tlc2.tool.Simulator.rl.alpha", "TLAGO_SIMULATOR_RL_ALPHA", 0.3)
	worker.RLGamma = simulatorPropertyFloat("tlc2.tool.Simulator.rl.gamma", "TLAGO_SIMULATOR_RL_GAMMA", 0.7)
	worker.RLReward = simulatorPropertyFloat("tlc2.tool.Simulator.rl.reward", "TLAGO_SIMULATOR_RL_REWARD", -10)
	worker.RLEnabledOnly = simulatorPropertyBool("tlc2.tool.Simulator.rl.enabledOnly", "TLAGO_SIMULATOR_RL_ENABLED_ONLY")
	return worker
}

func (s *Simulator) newWorkerLiveCheck(tool *Tool, workerID int) *LiveCheck {
	metadir := filepath.Join(s.MetaDir, fmt.Sprintf("simulator_%d", workerID))
	if tool == nil || tool.LivenessIsTrue() {
		return NewNoOpLiveCheck(tool, metadir)
	}
	check, err := NewLiveCheckFromTool(tool.NoDebug(), metadir, nil)
	if err != nil {
		if s.LiveCheckInitErr == nil {
			s.LiveCheckInitErr = err
		}
		return NewNoOpLiveCheck(tool, metadir)
	}
	return check
}

func (s *Simulator) newWorkerLiveCheck1(tool *Tool, workerID int) *LiveCheck1 {
	if tool == nil || tool.LivenessIsTrue() {
		return nil
	}
	check, err := NewLiveCheck1WithError(tool.NoDebug(), &s.LiveCheck1Errors, workerID != 0)
	if err != nil {
		if s.LiveCheckInitErr == nil {
			s.LiveCheckInitErr = err
		}
		return nil
	}
	return check
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
	if value, ok := tlcLookupSystemProperty(name); ok {
		return javaBooleanProperty(value)
	}
	for _, key := range aliases {
		if value, ok := os.LookupEnv(key); ok {
			return javaBooleanProperty(value)
		}
	}
	return false
}

func simulatorPropertyFloat(name string, alias string, fallback float64) float64 {
	if value, ok := tlcLookupSystemProperty(name); ok {
		if parsed, err := strconv.ParseFloat(strings.TrimSuffix(value, "d"), 64); err == nil {
			return parsed
		}
		return fallback
	}
	for _, key := range []string{alias} {
		if value, ok := os.LookupEnv(key); ok {
			if parsed, err := strconv.ParseFloat(strings.TrimSuffix(value, "d"), 64); err == nil {
				return parsed
			}
		}
	}
	return fallback
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

type actionFlowGraphContexts int

const (
	actionFlowGraphKeep actionFlowGraphContexts = iota
	actionFlowGraphReduce
)

type actionFlowGraphSnapshot struct {
	actions     []*Action
	actionStats [][]int64
}

func (s *Simulator) writeActionFlowGraph() error {
	if s == nil {
		return nil
	}
	switch s.TraceActions {
	case "BASIC":
		return s.writeActionFlowGraphBasic()
	case "FULL":
		return s.writeActionFlowGraphFull()
	default:
		return nil
	}
}

func (s *Simulator) getActionFlowGraphSnapshot(contexts actionFlowGraphContexts) *actionFlowGraphSnapshot {
	if s == nil || s.Tool == nil {
		return &actionFlowGraphSnapshot{}
	}
	actions := s.Tool.GetSpecActions()
	length := len(actions)
	aggregate := make([][]int64, length)
	for i := range aggregate {
		aggregate[i] = make([]int64, length)
	}
	for _, worker := range s.Workers {
		if worker == nil || worker.Statistics == nil {
			continue
		}
		workerStats := worker.Statistics.ActionStats
		for i := 0; i < length && i < len(workerStats); i++ {
			for j := 0; j < length && j < len(workerStats[i]); j++ {
				aggregate[i][j] += workerStats[i][j]
			}
		}
	}
	if contexts == actionFlowGraphKeep {
		return &actionFlowGraphSnapshot{actions: actions, actionStats: aggregate}
	}
	reducedActions := make([]*Action, 0, length)
	actionToID := NewInsMap[SourceLocation, int]()
	actionsToDistinctActions := make([]int, length)
	for _, action := range actions {
		if action == nil {
			continue
		}
		definition := action.GetDefinitionLocation()
		id, ok := actionToID.Get2(definition)
		if !ok {
			id = len(reducedActions)
			actionToID.Set(definition, id)
			reducedActions = append(reducedActions, action)
		}
		actionID := action.GetID()
		if actionID >= 0 && actionID < len(actionsToDistinctActions) {
			actionsToDistinctActions[actionID] = id
		}
	}
	reducedStats := make([][]int64, len(reducedActions))
	for i := range reducedStats {
		reducedStats[i] = make([]int64, len(reducedActions))
	}
	for i := 0; i < length; i++ {
		originID := actionsToDistinctActions[i]
		for j := 0; j < length; j++ {
			nextID := actionsToDistinctActions[j]
			if originID >= 0 && originID < len(reducedStats) && nextID >= 0 && nextID < len(reducedStats[originID]) {
				reducedStats[originID][nextID] += aggregate[i][j]
			}
		}
	}
	return &actionFlowGraphSnapshot{actions: reducedActions, actionStats: reducedStats}
}

func (s *Simulator) writeActionFlowGraphFull() error {
	snapshot := s.getActionFlowGraphSnapshot(actionFlowGraphKeep)
	if len(snapshot.actions) == 0 {
		return nil
	}
	clusters := NewInsMap[string, []int]()
	for id, action := range snapshot.actions {
		context := "[]"
		if action != nil && action.Con != nil {
			context = action.Con.String()
		}
		clusters.Set(context, append(clusters.Get(context), id))
	}
	writer, err := NewDotActionWriter(s.Tool.GetRootName()+"_actions.dot", "")
	if err != nil {
		return err
	}
	defer writer.Close()
	for context, ids := range clusters.All() {
		key := javaAbsStringHash(context)
		if err := writer.WriteSubGraphStart(key, context); err != nil {
			return err
		}
		for _, id := range ids {
			if err := writer.WriteAction(snapshot.actions[id], id); err != nil {
				return err
			}
		}
		if err := writer.WriteSubGraphEnd(); err != nil {
			return err
		}
	}
	return writeActionFlowGraphEdges(writer, snapshot)
}

func (s *Simulator) writeActionFlowGraphBasic() error {
	snapshot := s.getActionFlowGraphSnapshot(actionFlowGraphReduce)
	if len(snapshot.actions) == 0 {
		return nil
	}
	writer, err := NewDotActionWriter(s.Tool.GetRootName()+"_actions.dot", "")
	if err != nil {
		return err
	}
	defer writer.Close()
	for id, action := range snapshot.actions {
		if err := writer.WriteAction(action, id); err != nil {
			return err
		}
	}
	return writeActionFlowGraphEdges(writer, snapshot)
}

func writeActionFlowGraphEdges(writer *DotActionWriter, snapshot *actionFlowGraphSnapshot) error {
	length := len(snapshot.actions)
	for i := 0; i < length; i++ {
		for j := 0; j < length; j++ {
			count := int64(0)
			if i < len(snapshot.actionStats) && j < len(snapshot.actionStats[i]) {
				count = snapshot.actionStats[i][j]
			}
			if count > 0 {
				if err := writer.WriteEdge(i, j, actionFlowGraphWeight(count)); err != nil {
					return err
				}
			} else if snapshot.actions[j] == nil || !snapshot.actions[j].IsInitPredicate() {
				if err := writer.WriteEdge(i, j); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func actionFlowGraphWeight(count int64) float64 {
	weight := math.Log10(math.Log10(float64(count)+1) + 1)
	return math.Round(weight*100) / 100
}

func javaAbsStringHash(value string) string {
	hash := javaStringHashCode(value)
	if hash < 0 && hash != math.MinInt32 {
		hash = -hash
	}
	return fmt.Sprintf("%d", hash)
}
