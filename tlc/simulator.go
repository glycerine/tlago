package tlc

import (
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Simulator struct {
	Tool             *Tool
	invariants       []*Action
	CheckDeadlock    bool
	checkLiveness    bool
	TraceDepth       int
	TraceNum         int64
	TraceFile        *string
	TraceActions     *string
	MetaDir          string
	StartTime        time.Time
	Rand             *JavaRandom
	Seed             int64
	Aril             int64
	Config           Value
	ResultQueue      *SimulationWorkerResultQueue
	Workers          []*SimulationWorker
	workerCount      int
	metaDirSet       bool
	WorkerMode       SimulationWorkerMode
	workerModeSet    bool
	LiveCheck        *LiveCheck
	LiveCheck1Errors atomic.Bool
	NumGenStates     *atomic.Int64
	NumGenTraces     *atomic.Int64
	WelfordM2Mean    *atomic.Int64

	StatesGenerated int64
	TracesGenerated int64
	DisabledRetries int64
	Stopped         bool

	configurationCaptured bool
	toolOwner             *Tool
	checkDeadlockOwner    bool
	checkLivenessOwner    bool
	traceDepthOwner       int
	traceNumOwner         int64
	traceFileOwner        *string
	traceActionsOwner     *string
	randOwner             *JavaRandom
	seedOwner             int64
	configOwner           Value
	resultQueueOwner      *SimulationWorkerResultQueue
	workersOwner          []*SimulationWorker
	startTimeOwner        time.Time
	numGenStatesOwner     *atomic.Int64
	numGenTracesOwner     *atomic.Int64
	welfordM2MeanOwner    *atomic.Int64
}

func (s *Simulator) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *Simulator) captureConfiguration() {
	s.toolOwner = s.Tool
	s.checkDeadlockOwner = s.CheckDeadlock
	s.checkLivenessOwner = s.checkLiveness
	s.traceDepthOwner = s.TraceDepth
	s.traceNumOwner = s.TraceNum
	if s.TraceFile != nil {
		traceFile := *s.TraceFile
		s.traceFileOwner = &traceFile
	}
	if s.TraceActions != nil {
		traceActions := *s.TraceActions
		s.traceActionsOwner = &traceActions
	}
	s.randOwner = s.Rand
	s.seedOwner = s.Seed
	s.configOwner = s.Config
	s.resultQueueOwner = s.ResultQueue
	s.workersOwner = s.Workers
	s.startTimeOwner = s.StartTime
	s.numGenStatesOwner = s.NumGenStates
	s.numGenTracesOwner = s.NumGenTraces
	s.welfordM2MeanOwner = s.WelfordM2Mean
	s.configurationCaptured = true
}

func (s *Simulator) ownedTool() *Tool {
	if s.configurationCaptured {
		return s.toolOwner
	}
	return s.Tool
}

func (s *Simulator) ownedCheckDeadlock() bool {
	if s.configurationCaptured {
		return s.checkDeadlockOwner
	}
	return s.CheckDeadlock
}

func (s *Simulator) ownedCheckLiveness() bool {
	if s.configurationCaptured {
		return s.checkLivenessOwner
	}
	return s.checkLiveness
}

func (s *Simulator) ownedTraceDepth() int {
	if s.configurationCaptured {
		return s.traceDepthOwner
	}
	return s.TraceDepth
}

func (s *Simulator) ownedTraceNum() int64 {
	if s.configurationCaptured {
		return s.traceNumOwner
	}
	return s.TraceNum
}

func (s *Simulator) ownedTraceFile() *string {
	if s.configurationCaptured {
		return s.traceFileOwner
	}
	return s.TraceFile
}

func (s *Simulator) ownedTraceActions() *string {
	if s.configurationCaptured {
		return s.traceActionsOwner
	}
	return s.TraceActions
}

func (s *Simulator) ownedRand() *JavaRandom {
	if s.configurationCaptured {
		return s.randOwner
	}
	return s.Rand
}

func (s *Simulator) ownedSeed() int64 {
	if s.configurationCaptured {
		return s.seedOwner
	}
	return s.Seed
}

func (s *Simulator) ownedConfig() Value {
	if s.configurationCaptured {
		return s.configOwner
	}
	return s.Config
}

func (s *Simulator) ownedResultQueue() *SimulationWorkerResultQueue {
	if s.configurationCaptured {
		return s.resultQueueOwner
	}
	return s.ResultQueue
}

func (s *Simulator) ownedWorkers() []*SimulationWorker {
	if s.configurationCaptured {
		return s.workersOwner
	}
	return s.Workers
}

func (s *Simulator) ownedStartTime() time.Time {
	if s.configurationCaptured {
		return s.startTimeOwner
	}
	return s.StartTime
}

func (s *Simulator) ownedNumGenStates() *atomic.Int64 {
	if s.configurationCaptured {
		return s.numGenStatesOwner
	}
	if s.NumGenStates == nil {
		s.NumGenStates = &atomic.Int64{}
	}
	return s.NumGenStates
}

func (s *Simulator) ownedNumGenTraces() *atomic.Int64 {
	if s.configurationCaptured {
		return s.numGenTracesOwner
	}
	if s.NumGenTraces == nil {
		s.NumGenTraces = &atomic.Int64{}
	}
	return s.NumGenTraces
}

func (s *Simulator) ownedWelfordM2Mean() *atomic.Int64 {
	if s.configurationCaptured {
		return s.welfordM2MeanOwner
	}
	if s.WelfordM2Mean == nil {
		s.WelfordM2Mean = &atomic.Int64{}
	}
	return s.WelfordM2Mean
}

const simulatorUnboundedTraceDepth = math.MaxInt32

const (
	simulatorRLProperty       = "tlc2.tool.Simulator.rl"
	simulatorRLActionProperty = "tlc2.tool.Simulator.rlaction"
)

var simulatorStatics struct {
	once                    sync.Once
	extendedStatistics      bool
	extendedStatisticsNaive bool
	experimentalLiveness    bool
}

// InitializeSimulatorStatics represents a fresh Java classloader for
// Simulator's property-backed static fields.
func InitializeSimulatorStatics() {
	simulatorStatics.once = sync.Once{}
	simulatorStatics.extendedStatistics = false
	simulatorStatics.extendedStatisticsNaive = false
	simulatorStatics.experimentalLiveness = false
}

func initializeSimulatorStatics() {
	simulatorStatics.once.Do(func() {
		simulatorStatics.extendedStatistics = simulatorPropertyBool(
			"tlc2.tool.Simulator.extendedStatistics",
			"TLAGO_SIMULATOR_EXTENDED_STATISTICS",
		)
		simulatorStatics.extendedStatisticsNaive = simulatorPropertyBool(
			"tlc2.tool.Simulator.extendedStatistics.naive",
			"TLAGO_SIMULATOR_EXTENDED_STATISTICS_NAIVE",
		)
		simulatorStatics.experimentalLiveness = simulatorPropertyBool(
			"tlc2.tool.Simulator.experimentalLiveness",
			"TLAGO_SIMULATOR_EXPERIMENTAL_LIVENESS",
		)
	})
}

// SetSimulatorExtendedStatistics represents assignment to Java's public
// Simulator.EXTENDED_STATISTICS static field.
func SetSimulatorExtendedStatistics(enabled bool) {
	initializeSimulatorStatics()
	simulatorStatics.extendedStatistics = enabled
}

// SetSimulatorExtendedStatisticsNaive represents assignment to Java's public
// Simulator.EXTENDED_STATISTICS_NAIVE static field.
func SetSimulatorExtendedStatisticsNaive(enabled bool) {
	initializeSimulatorStatics()
	simulatorStatics.extendedStatisticsNaive = enabled
}

// SetSimulatorExperimentalLiveness represents assignment to Java's public
// Simulator.EXPERIMENTAL_LIVENESS_SIMULATION static field. Assignment first
// initializes the class, then replaces its property-backed initial value.
func SetSimulatorExperimentalLiveness(enabled bool) {
	initializeSimulatorStatics()
	simulatorStatics.experimentalLiveness = enabled
}

type SimulatorOption func(*Simulator)

func WithSimulatorTraceFile(traceFile *string) SimulatorOption {
	return func(s *Simulator) {
		s.TraceFile = traceFile
	}
}

func WithSimulatorTraceActions(traceActions *string) SimulatorOption {
	return func(s *Simulator) {
		s.TraceActions = traceActions
	}
}

func WithSimulatorMetaDir(metadir string) SimulatorOption {
	return func(s *Simulator) {
		s.MetaDir = metadir
		s.metaDirSet = true
	}
}

func WithSimulatorSchedule(schedule SimulationSchedule) SimulatorOption {
	return func(s *Simulator) {
		switch schedule {
		case SimulationScheduleRL:
			s.WorkerMode = SimulationWorkerRL
			s.workerModeSet = true
		case SimulationScheduleRLAction:
			s.WorkerMode = SimulationWorkerRLAction
			s.workerModeSet = true
		}
	}
}

func WithSimulatorAril(aril int64) SimulatorOption {
	return func(s *Simulator) {
		s.Aril = aril
	}
}

// WithSimulatorRandomGenerator carries the source constructor's caller-owned
// RandomGenerator, including its current seed position.
func WithSimulatorRandomGenerator(rng *JavaRandom) SimulatorOption {
	return func(s *Simulator) { s.Rand = rng }
}

func WithSimulatorLiveCheck(liveCheck *LiveCheck) SimulatorOption {
	return func(s *Simulator) {
		s.LiveCheck = liveCheck
	}
}

// WithSimulatorWorkerCount carries the explicit source constructor argument,
// including zero workers for simulators used only to print a behavior.
func WithSimulatorWorkerCount(count int) SimulatorOption {
	return func(s *Simulator) { s.workerCount = int(int32(count)) }
}

func NewSimulator(tool *Tool, deadlock bool, traceDepth int, traceNum int64, seed int64, opts ...SimulatorOption) *Simulator {
	initializeSimulatorStatics()
	// Java initializes this instance field before entering the constructor body.
	startTime := time.Now()
	if tool == nil {
		panic(NewNullPointerException())
	}
	SetTLCStateTool(tool)
	traceDepth = int(int32(traceDepth))
	if traceDepth == -1 {
		traceDepth = simulatorUnboundedTraceDepth
	}
	checkDeadlock := false
	if deadlock {
		checkDeadlock = tool.GetModelConfig().GetCheckDeadlock()
	}
	checkLiveness := !tool.LivenessIsTrue()
	simulator := &Simulator{
		Tool:          tool,
		CheckDeadlock: checkDeadlock,
		checkLiveness: checkLiveness,
		TraceDepth:    traceDepth,
		TraceNum:      traceNum,
		Seed:          seed,
		MetaDir:       "states",
		StartTime:     startTime,
		ResultQueue:   NewSimulationWorkerResultQueue(),
		NumGenStates:  &atomic.Int64{},
		NumGenTraces:  &atomic.Int64{},
		WelfordM2Mean: &atomic.Int64{},
		workerCount:   max(NumWorkers(), 1),
		Rand:          NewJavaRandom(seed),
	}
	if tool != nil {
		simulator.invariants = tool.GetInvariants()
	}
	simulator.WorkerMode = simulator.selectWorkerMode()
	for _, opt := range opts {
		if opt != nil {
			opt(simulator)
		}
	}
	if simulator.MetaDir == "" && !simulator.metaDirSet {
		simulator.MetaDir = "states"
	}
	if simulator.Aril > 0 {
		if simulator.Rand == nil {
			panic(NewNullPointerException())
		}
		simulator.Rand.SetSeedWithAril(simulator.Seed, simulator.Aril)
	}
	// Java keeps the reported aril zero until simulation startup, independently
	// of the RNG's replay position. Config captures this construction-time value.
	simulator.Aril = 0
	// Java shares one no-op checker, retaining the original tool even when
	// later debugger workers use its noDebug tool for state generation.
	if simulator.LiveCheck == nil && !simulator.checkLiveness {
		simulator.LiveCheck = NewNoOpLiveCheck(tool, simulator.MetaDir)
	}
	if simulator.workerCount < 0 {
		panic(NewIllegalArgumentException(fmt.Sprintf("Illegal Capacity: %d", simulator.workerCount)))
	}
	for i := 0; i < simulator.workerCount; i++ {
		simulator.Workers = append(simulator.Workers, simulator.newSimulationWorker(i))
	}
	simulator.Config = simulator.createConfig()
	simulator.captureConfiguration()
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(tool)
	}
	scheduleStopAfterFromJavaProperty(simulator.Stop)
	return simulator
}

func (s *Simulator) Simulate() (int, error) {
	s.require()
	tool := s.ownedTool()
	if tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "simulator has no tool")
	}
	if result := tool.CheckAssumptions(); result != NoError {
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
	s.Aril = s.ownedRand().GetAril()
	workerResult := s.SimulateWithStates(initStates)
	s.StatesGenerated = s.ownedNumGenStates().Load()
	s.TracesGenerated = s.ownedNumGenTraces().Load()
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
	s.require()
	s.ReportCoverage()
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				exception, ok := recovered.(error)
				if !ok || isJavaError(exception) {
					panic(recovered)
				}
				PrintTLCBug(ECTLCReporterDied)
			}
		}()
		if err := s.writeActionFlowGraph(); err != nil {
			PrintTLCBug(ECTLCReporterDied)
		}
	}()
	if toolMode() {
		PrintMessage(ECTLCProgressSimu,
			fmtInt64(s.ownedNumGenStates().Load()),
			fmtInt64(s.ownedNumGenTraces().Load()),
		)
	}
	PrintMessage(ECTLCStatsSimu,
		fmtInt64(s.ownedNumGenStates().Load()),
		fmtInt64(s.ownedSeed()),
		fmtInt64(s.Aril),
	)
}

func (s *Simulator) ReportCoverage() {
	s.require()
	if CoverageEnabled() {
		ReportCoverage(s.ownedTool(), s.ownedStartTime())
	}
}

func (s *Simulator) GetTool() *Tool {
	s.require()
	return s.ownedTool()
}

func (s *Simulator) startProgressReporter() func() {
	s.require()
	tool := s.ownedTool()
	if tool == nil {
		return func() {}
	}
	interval := ProgressInterval()
	stop := make(chan struct{})
	done := make(chan struct{})
	ready := make(chan struct{})
	go func() {
		defer close(done)
		readyAnnounced := false
		defer func() {
			if !readyAnnounced {
				close(ready)
			}
		}()
		defer func() {
			if recovered := recover(); recovered != nil {
				exception, ok := recovered.(error)
				if !ok || isJavaError(exception) {
					panic(recovered)
				}
				PrintTLCBug(ECTLCReporterDied)
			}
		}()
		periodic := tool.GetPeriodic()
		coverageCountdown := periodicCoverageCountdown(interval)
		close(ready)
		readyAnnounced = true
		if interval < 0 {
			panic(NewIllegalArgumentException("timeout value is negative"))
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				switch s.reportSimulationProgress(&coverageCountdown, interval, periodic) {
				case simulatorProgressStop:
					s.ownedResultQueue().Put(SimulationWorkerOK(-1))
				case simulatorProgressReporterDone:
					return
				}
				// Java checks isRunning after this report and exits without
				// processing a pending completion notification a second time.
				select {
				case <-stop:
					return
				default:
				}
			case <-stop:
				if s.reportSimulationProgress(&coverageCountdown, interval, periodic) == simulatorProgressStop {
					s.ownedResultQueue().Put(SimulationWorkerOK(-1))
				}
				return
			}
		}
	}()
	<-ready
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

func (s *Simulator) reportSimulationProgress(coverageCountdown *int, interval time.Duration, periodic SemanticNode) simulatorProgressStatus {
	s.require()
	genTrace := s.ownedNumGenTraces().Load()
	m2AndMean := s.ownedWelfordM2Mean().Load()
	mean := int64(m2AndMean & 0xffffffff)
	m2 := uint64(m2AndMean) >> 32
	PrintMessage(ECTLCProgressSimu,
		fmtInt64(s.ownedNumGenStates().Load()),
		fmtInt64(genTrace),
		fmtInt64(mean),
		fmtInt64(javaMathRound(float64(m2)/(float64(genTrace)+1))),
		fmtInt64(javaMathRound(math.Sqrt(float64(m2)/(float64(genTrace)+1)))),
	)
	if coverageCountdown != nil {
		if *coverageCountdown > 1 {
			(*coverageCountdown)--
		} else {
			if CoverageEnabled() {
				ReportCoverage(s.ownedTool(), s.ownedStartTime())
			}
			*coverageCountdown = periodicCoverageCountdown(interval)
		}
	}
	if err := s.writeActionFlowGraph(); err != nil {
		if isJavaError(err) {
			panic(err)
		}
		PrintTLCBug(ECTLCReporterDied)
		return simulatorProgressReporterDone
	}
	if tool := s.ownedTool(); tool != nil && !semanticExploreNull(periodic) {
		value, err := tool.NoDebug().Eval(periodic)
		if err != nil {
			if isJavaError(err) {
				panic(err)
			}
			PrintTLCBug(ECTLCReporterDied)
			return simulatorProgressReporterDone
		}
		isFalse, err := BoolFalse.Equal(value)
		if err != nil {
			if isJavaError(err) {
				panic(err)
			}
			PrintTLCBug(ECTLCReporterDied)
			return simulatorProgressReporterDone
		}
		if isFalse {
			PrintError(ECTLCAssumptionFalse, SemanticString(periodic))
			return simulatorProgressStop
		}
	}
	return simulatorProgressContinue
}

func (s *Simulator) Stop() {
	s.require()
	s.Stopped = true
	for _, worker := range s.ownedWorkers() {
		worker.Stop()
	}
}

func (s *Simulator) shutdownAndJoinWorkers(workers []*SimulationWorker) {
	s.require()
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
	s.require()
	workers := s.ownedWorkers()
	if len(workers) == 0 {
		return nil
	}
	return workers[0].GetLocalValue(idx)
}

func (s *Simulator) SetAllValues(idx int, value Value) {
	s.require()
	for _, worker := range s.ownedWorkers() {
		worker.SetLocalValue(idx, value)
	}
}

// GetAllValuesAt is Java's indexed getAllValues overload.
func (s *Simulator) GetAllValuesAt(index int) []Value {
	s.require()
	workers := s.ownedWorkers()
	values := make([]Value, len(workers))
	for i, worker := range workers {
		values[i] = worker.GetLocalValue(index)
	}
	return values
}

func (s *Simulator) GetAllValues() Value {
	s.require()
	worker := s.firstRegisterWorker()
	workers := s.ownedWorkers()
	locals := worker.GetLocalValues()
	pairs := newTLCValueHashMapWithCapacity[Value](len(locals))
	for idx := range locals {
		if worker.GetLocalValue(idx) == nil {
			continue
		}
		workerValues := make([]Value, len(workers))
		for i, w := range workers {
			workerValues[i] = w.GetLocalValue(idx)
		}
		pairs.Set(NewIntValue(int32(idx)), NewTupleValue(workerValues))
	}
	return NewFcnRcdValueFromMap(pairs)
}

func (s *Simulator) GetLocalNamedValue(key *UniqueString) Value {
	s.require()
	workers := s.ownedWorkers()
	if len(workers) == 0 {
		return nil
	}
	return workers[0].GetNamedRegister(key)
}

func (s *Simulator) SetAllNamedValues(key *UniqueString, value Value) {
	s.require()
	for _, worker := range s.ownedWorkers() {
		worker.SetNamedRegister(key, value)
	}
}

func (s *Simulator) GetAllNamedRegisterValues() Value {
	s.require()
	keys := s.firstRegisterWorker().namedRegisterKeys()
	workers := s.ownedWorkers()
	pairs := newTLCValueHashMap[Value]()
	for _, key := range keys {
		workerValues := make([]Value, len(workers))
		for i, worker := range workers {
			workerValues[i] = worker.GetNamedRegister(key)
		}
		pairs.Set(NewStringValueFromUnique(key), NewTupleValue(workerValues))
	}
	return NewFcnRcdValueFromMap(pairs)
}

func (s *Simulator) GetAllNamedValues(key *UniqueString) []Value {
	s.require()
	workers := s.ownedWorkers()
	values := make([]Value, len(workers))
	for i, worker := range workers {
		values[i] = worker.GetNamedRegister(key)
	}
	return values
}

func (s *Simulator) firstRegisterWorker() *SimulationWorker {
	s.require()
	workers := s.ownedWorkers()
	if len(workers) == 0 {
		panic(NewIndexOutOfBoundsException(0, 0))
	}
	if workers[0] == nil {
		panic(NewNullPointerException())
	}
	return workers[0]
}

func (s *Simulator) GetStatistics(state *TLCStateMut) Value {
	s.require()
	ensureTLCGetSetUniqueStrings()
	traces := s.ownedNumGenTraces().Load()
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
	values := make([]Value, len(names))
	values[0] = intValueFromInt64(traces)
	values[1] = intValueFromDurationSince(s.ownedStartTime())
	values[2] = intValueFromInt64(s.ownedNumGenStates().Load())
	values[3] = s.GetWorkerStatistics().GetTraceStatistics(state)
	values[4] = NewIntValue(int32(currentWorkerIDOrZero()))
	values[5] = s.GetWorkerStatistics().GetDistinctStates()
	values[6] = s.GetWorkerStatistics().GetDistinctValues()
	values[7] = s.GetWorkerStatistics().GetNextRetries()
	values[8] = s.GetWorkerStatistics().GetActions()
	m2AndMean := s.ownedWelfordM2Mean().Load()
	mean := int64(m2AndMean & 0xffffffff)
	m2 := uint64(m2AndMean) >> 32
	values[9] = intValueFromInt64(mean)
	values[10] = intValueFromInt64(javaMathRound(float64(m2) / (float64(traces) + 1)))
	return NewRecordValue(names, values, false)
}

func (s *Simulator) GetTrace(state *TLCStateMut) *StateVec {
	s.require()
	return s.traceWorker().GetTrace(state)
}

func (s *Simulator) GetUncompressedTrace(state *TLCStateMut) *StateVec {
	s.require()
	return s.traceWorker().GetUncompressedTrace(state)
}

func (s *Simulator) traceWorker() *SimulationWorker {
	s.require()
	if worker := currentSimulationWorker(); worker != nil {
		return worker
	}
	workers := s.ownedWorkers()
	if s.workerCount != 1 || len(workers) != s.workerCount {
		panic(NewAssertionError())
	}
	if workers[0] == nil {
		panic(NewNullPointerException())
	}
	return workers[0]
}

func (s *Simulator) GetRNG() *JavaRandom {
	s.require()
	if worker := currentSimulationWorker(); worker != nil {
		return worker.GetRNG()
	}
	return s.ownedRand()
}

func (s *Simulator) GetTraceDepth() int {
	s.require()
	return s.ownedTraceDepth()
}

func (s *Simulator) GetConfig() Value {
	s.require()
	return s.ownedConfig()
}

func (s *Simulator) createConfig() Value {
	s.require()
	ensureTLCGetSetUniqueStrings()
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
	depth := int32(s.ownedTraceDepth())
	if s.ownedTraceDepth() == simulatorUnboundedTraceDepth {
		depth = -1
	}
	workerCount := s.workerCount
	mode := "simulate"
	if toolProbabilisticEnabled() {
		mode = "generate"
	}
	values := []Value{
		NewStringValue(mode),
		NewIntValue(depth),
		NewIntValue(int32(int64(workerCount) * s.ownedTraceNum())),
		BoolValueOf(s.ownedCheckDeadlock()),
		NewStringValue(fmt.Sprintf("%d", s.ownedSeed())),
		NewStringValue("0"),
		NewIntValue(int32(workerCount)),
		NewStringValue(TLCInstallLocation()),
		NewStringValue(s.schedulerName()),
	}
	return NewRecordValue(names, values, false)
}

// GetWorkerStatistics selects the actual SimulationWorker thread, or worker zero.
func (s *Simulator) GetWorkerStatistics() *SimulationWorkerStatistics {
	s.require()
	if worker := currentSimulationWorker(); worker != nil {
		return worker.ownedStatistics()
	}
	return s.firstRegisterWorker().ownedStatistics()
}

func currentWorkerIDOrZero() int {
	if id, ok := CurrentWorkerID(); ok {
		return id
	}
	return 0
}

func (s *Simulator) schedulerName() string {
	mode := s.selectWorkerMode()
	if s != nil && s.workerModeSet {
		mode = s.WorkerMode
	}
	switch mode {
	case SimulationWorkerRL:
		return "rl"
	case SimulationWorkerRLAction:
		return "rlaction"
	default:
		return "random"
	}
}

func (s *Simulator) initialStates() (*StateVec, int, error) {
	s.require()
	filtered, code := s.collectInitialStates()
	if code != NoError {
		return nil, code, nil
	}
	if s.ownedNumGenStates().Load() == 0 {
		return nil, PrintError(ECTLCNoStatesSatisfyingInit), nil
	}
	if filtered.IsEmpty() {
		return nil, PrintError(ECTLCNoStatesSatisfyingInitAndConstraint), nil
	}
	return filtered, NoError, nil
}

// collectInitialStates retains Simulator.simulate's try/catch(Exception)
// around initial generation, validity and model constraints. Java Error
// subclasses escape this catch; handled exceptions return a diagnostic code.
func (s *Simulator) collectInitialStates() (filtered *StateVec, code int) {
	s.require()
	tool := s.ownedTool()
	var curState *TLCStateMut
	defer func() {
		if recovered := recover(); recovered != nil {
			exception, ok := recovered.(error)
			if !ok || isJavaError(exception) {
				panic(recovered)
			}
			code = s.printInitialStateException(curState, exception)
			filtered = nil
		}
	}()
	all := NewStateVec(0)
	if err := tool.GetInitStates(NewStateFunctor(func(state *TLCStateMut) (any, error) {
		all.Add(state)
		return all, nil
	})); err != nil {
		panic(err)
	}
	filtered = NewStateVec(all.Size())
	numGenStates := s.ownedNumGenStates()
	if numGenStates.Load() != 0 {
		panic(NewAssertionError())
	}
	s.StatesGenerated += int64(all.Size())
	numGenStates.Add(int64(all.Size()))
	PrintMessage(ECTLCComputingInitProgress, fmtInt64(numGenStates.Load()))
	for i := 0; i < all.Size(); i++ {
		curState = all.At(i)
		if !tool.IsGoodState(curState) {
			return nil, PrintError(ECTLCStateNotCompletelySpecifiedInitial, curState.String())
		}
		for j, invariant := range tool.requireActionArray(s.invariants) {
			valid, err := tool.IsValidState(invariant, curState)
			if err != nil {
				panic(err)
			}
			if !valid {
				alias := tool.EvalAlias(curState, curState)
				result := PrintError(ECTLCInvariantViolatedInitial, tool.propertyNameAt(tool.GetInvNames(), j), alias.String())
				tool.CheckPostConditionWithCounterExample(NewCounterExampleFromInitialState(curState))
				return nil, result
			}
		}
		inModel, err := tool.IsInModel(curState)
		if err != nil {
			panic(err)
		}
		if inModel {
			filtered.Add(curState)
		}
	}
	return filtered, NoError
}

func (s *Simulator) printInitialStateException(state *TLCStateMut, err error) int {
	s.require()
	message := javaThrowableMessage(err)
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

func (s *Simulator) SimulateWithStates(initStates *StateVec) SimulationWorkerResult {
	s.require()
	workers := s.ownedWorkers()
	running := make(map[int]bool, len(workers))
	runningCount := 0
	for i, worker := range workers {
		worker.Start(initStates)
		running[i] = true
		runningCount++
	}
	var result SimulationWorkerResult
	for {
		result = s.ownedResultQueue().Take()
		if result.WorkerID() == -1 {
			break
		}
		if result.IsError() {
			s.printSimulationWorkerError(result.Error())
			if s.simulationErrorStops(result.Error()) {
				break
			}
			continue
		}
		if running[result.WorkerID()] {
			delete(running, result.WorkerID())
			runningCount--
		}
		if runningCount == 0 {
			break
		}
	}
	s.shutdownAndJoinWorkers(workers)
	return result
}

// simulate retains the earlier native spelling.
func (s *Simulator) simulate(initStates *StateVec) SimulationWorkerResult {
	s.require()
	return s.SimulateWithStates(initStates)
}

func (s *Simulator) printSimulationWorkerError(err *SimulationWorkerError) {
	s.require()
	if err == nil {
		return
	}
	if exception := err.ownedErr(); exception != nil {
		if live := javaLiveException(exception); live != nil {
			err.Code = live.ownedErrorCode()
			s.PrintSummary()
			return
		}
		if failure := javaRuntimeException(exception); failure != nil {
			err.Code = failure.ownedCode()
			s.PrintBehaviorException(failure, err.ownedStateTrace())
			return
		}
		err.Code = ECGeneral
		s.PrintBehavior(err.Code, generalErrorParams("", exception), err.ownedStateTrace())
		return
	}
	if params := err.ownedNullableParams(); params != nil {
		PrintErrorNullable(err.Code, params...)
		s.printBehaviorTrace(err.ownedStateTrace())
		s.PrintSummary()
	} else {
		s.PrintBehaviorWorkerError(err)
	}
}

// PrintBehaviorException ports Simulator.printBehavior(TLCRuntimeException,
// StateVec). Unlike the coded-error overload, this overload has no summary.
func (s *Simulator) PrintBehaviorException(exception *TLCError, stateTrace *StateVec) {
	s.require()
	PrintTLCRuntimeException(exception)
	s.printBehaviorTrace(stateTrace)
}

func (s *Simulator) PrintBehaviorWorkerError(err *SimulationWorkerError) {
	s.require()
	if err == nil {
		panic(NewNullPointerException())
	}
	s.PrintBehavior(err.Code, err.ownedParams(), err.ownedStateTrace())
}

func (s *Simulator) PrintBehavior(errorCode int, params []string, stateTrace *StateVec) {
	s.require()
	PrintError(errorCode, params...)
	s.printBehaviorTrace(stateTrace)
	s.PrintSummary()
}

// printBehavior retains the earlier native spelling.
func (s *Simulator) printBehavior(errorCode int, params []string, stateTrace *StateVec) {
	s.require()
	s.PrintBehavior(errorCode, params, stateTrace)
}

func (s *Simulator) printBehaviorTrace(stateTrace *StateVec) {
	s.require()
	if s.ownedTraceDepth() == simulatorUnboundedTraceDepth {
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
		if curState == nil {
			panic(NewNullPointerException())
		}
		info := NewTLCStateInfo(curState)
		if lastState != nil && printDiffsOnly() && curState.FingerPrint() == lastState.FingerPrint() {
			omitted++
			lastState = curState
			continue
		}
		tool := s.ownedTool()
		if tool == nil {
			panic(NewNullPointerException())
		}
		aliased, err := tool.EvalAliasInfoPair(info, sucState)
		if err != nil {
			panic(err)
		}
		PrintInvariantViolationStateTraceState(aliased, lastState, curState.Level(), i+1 == stateTrace.Size())
		lastState = curState
	}
	if omitted > 0 {
		PrintMessage(ECGeneral, fmt.Sprintf("difftrace requested: Shortened behavior by omitting finite stuttering (%d states), which is an artifact of simulation mode.\n", omitted))
	}
}

func (s *Simulator) IsNonContinuableError(code int) bool {
	s.require()
	code = int(int32(code))
	return code == ECTLCInvariantEvaluationFailed ||
		code == ECTLCActionPropertyEvaluationFailed ||
		code == ECTLCStateNotCompletelySpecifiedNext
}

// isNonContinuableError retains the earlier native spelling.
func (s *Simulator) isNonContinuableError(code int) bool {
	s.require()
	return s.IsNonContinuableError(code)
}

func (s *Simulator) simulationErrorStops(err *SimulationWorkerError) bool {
	s.require()
	if err == nil {
		return false
	}
	if exception := err.ownedErr(); exception != nil {
		if live := javaLiveException(exception); live != nil {
			err.Code = live.ownedErrorCode()
		} else if failure := javaRuntimeException(exception); failure != nil {
			err.Code = failure.ownedCode()
		} else {
			err.Code = ECGeneral
		}
		return true
	}
	if s.IsNonContinuableError(err.Code) {
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
	s.require()
	errorCode := NoError
	if workerResult.IsError() {
		errorCode = workerResult.Error().Code
	}
	pcErrorCode := NoError
	if tool := s.ownedTool(); tool != nil {
		if workerResult.IsError() && workerResult.Error().HasTrace() {
			pcErrorCode = tool.CheckPostConditionWithCounterExample(workerResult.Error().GetCounterExample())
		} else {
			pcErrorCode = tool.CheckPostCondition()
		}
	}
	if ExitStatusForErrorCode(pcErrorCode) > ExitStatusForErrorCode(errorCode) {
		return pcErrorCode
	}
	return errorCode
}

func (s *Simulator) newSimulationWorker(id int) *SimulationWorker {
	s.require()
	toolOwner := s.ownedTool()
	liveCheck := s.LiveCheck
	var liveCheck1 *LiveCheck1
	if liveCheck == nil {
		if s.ownedCheckLiveness() && simulatorStatics.experimentalLiveness {
			liveCheck = s.newWorkerLiveCheck(toolOwner, id)
		} else if s.ownedCheckLiveness() {
			liveCheck1 = s.newWorkerLiveCheck1(toolOwner, id)
		} else {
			liveCheck = NewNoOpLiveCheck(toolOwner, s.MetaDir)
		}
	}
	tool := toolOwner
	if id != 0 && tool != nil {
		tool = tool.NoDebug()
	}
	debugger := toolOwner != nil && toolOwner.IsDebugger()
	workerMode := SimulationWorkerStandard
	if !debugger {
		workerMode = s.WorkerMode
		if !s.workerModeSet {
			workerMode = s.selectWorkerMode()
		}
	}
	rng := s.ownedRand()
	if rng == nil {
		panic(NewNullPointerException())
	}
	worker := NewSimulationWorker(
		id,
		tool,
		s.ownedResultQueue(),
		rng.NextLong(),
		s.ownedTraceDepth(),
		s.ownedTraceNum(),
		s.ownedTraceActions(),
		s.ownedCheckDeadlock(),
		debugger,
		s.ownedTraceFile(),
		liveCheck,
		s.ownedNumGenStates(),
		s.ownedNumGenTraces(),
		s.ownedWelfordM2Mean(),
	)
	worker.setLiveCheck1(liveCheck1)
	worker.SetMode(workerMode)
	return worker
}

func (s *Simulator) newWorkerLiveCheck(tool *Tool, workerID int) *LiveCheck {
	s.require()
	// Simulator owns the temporary directory, as Files.createTempDirectory in
	// its Java constructor does. Disk graphs never create missing parents.
	metadir, err := os.MkdirTemp("", fmt.Sprintf("tlc-simulator-%d-", workerID))
	if err != nil {
		if failure, ok := err.(*os.PathError); ok {
			panic(ioUtilsTXTPathError(failure.Path, failure.Err))
		}
		panic(err)
	}
	check, err := NewLiveCheckFromTool(tool.NoDebug(), metadir, nil)
	if err != nil {
		panic(err)
	}
	return check
}

func (s *Simulator) newWorkerLiveCheck1(tool *Tool, workerID int) *LiveCheck1 {
	s.require()
	check, err := NewLiveCheck1WithError(tool.NoDebug(), &s.LiveCheck1Errors, workerID != 0)
	if err != nil {
		panic(err)
	}
	return check
}

func (s *Simulator) selectWorkerMode() SimulationWorkerMode {
	if simulatorPropertyBool(simulatorRLProperty, "TLAGO_SIMULATOR_RL") {
		return SimulationWorkerRL
	}
	if simulatorPropertyBool(simulatorRLActionProperty, "TLAGO_SIMULATOR_RL_ACTION") {
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
		if parsed, err := parseJavaDoubleProperty(value); err == nil {
			return parsed
		}
		panic(NewNumberFormatException(value))
	}
	for _, key := range []string{alias} {
		if value, ok := os.LookupEnv(key); ok {
			if parsed, err := parseJavaDoubleProperty(value); err == nil {
				return parsed
			}
			panic(NewNumberFormatException(value))
		}
	}
	return fallback
}

func (s *Simulator) checkInvariants(state *TLCStateMut, initial bool) (int, error) {
	s.require()
	tool := s.ownedTool()
	for _, invariant := range tool.GetInvariants() {
		valid, err := tool.IsValidState(invariant, state)
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
	s.require()
	traceActions := s.ownedTraceActions()
	if traceActions == nil {
		return nil
	}
	switch *traceActions {
	case "BASIC":
		return s.writeActionFlowGraphBasic()
	case "FULL":
		return s.writeActionFlowGraphFull()
	default:
		return nil
	}
}

func (s *Simulator) getActionFlowGraphSnapshot(contexts actionFlowGraphContexts) *actionFlowGraphSnapshot {
	s.require()
	tool := s.ownedTool()
	if tool == nil {
		panic(NewNullPointerException())
	}
	actions := tool.GetSpecActions()
	length := len(actions)
	aggregate := make([][]int64, length)
	for i := range aggregate {
		aggregate[i] = make([]int64, length)
	}
	for _, worker := range s.ownedWorkers() {
		if worker == nil || worker.ownedStatistics() == nil {
			panic(NewNullPointerException())
		}
		workerStats := worker.ownedStatistics().ownedActionStats()
		for i := 0; i < length; i++ {
			for j := 0; j < length; j++ {
				aggregate[i][j] += simulationActionStatsCell(simulationActionStatsRow(workerStats, i), j)
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
		if action == nil || isNil(action.ownedPred()) {
			panic(NewNullPointerException())
		}
		definition := action.GetDefinition()
		id, found := actionToID.Get2(definition)
		if !found {
			id = len(reducedActions)
			actionToID.Set(definition, id)
			reducedActions = append(reducedActions, action)
		}
		actionID := action.GetID()
		if actionID < 0 || actionID >= len(actionsToDistinctActions) {
			panic(NewArrayIndexOutOfBoundsException(actionID, len(actionsToDistinctActions)))
		}
		actionsToDistinctActions[actionID] = id
	}
	reducedStats := make([][]int64, len(reducedActions))
	for i := range reducedStats {
		reducedStats[i] = make([]int64, len(reducedActions))
	}
	for i := 0; i < length; i++ {
		originID := actionsToDistinctActions[i]
		for j := 0; j < length; j++ {
			nextID := actionsToDistinctActions[j]
			reducedStats[originID][nextID] += aggregate[i][j]
		}
	}
	return &actionFlowGraphSnapshot{actions: reducedActions, actionStats: reducedStats}
}

func (s *Simulator) writeActionFlowGraphFull() error {
	s.require()
	snapshot := s.getActionFlowGraphSnapshot(actionFlowGraphKeep)
	clusters := NewInsMap[string, []int]()
	for id, action := range snapshot.actions {
		if action == nil || action.ownedContext() == nil {
			panic(NewNullPointerException())
		}
		context := action.ownedContext().String()
		clusters.Set(context, append(clusters.Get(context), id))
	}
	writer, err := NewDotActionWriter(s.ownedTool().GetRootName()+"_actions.dot", "")
	if err != nil {
		return err
	}
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
	if err := writeActionFlowGraphEdges(writer, snapshot); err != nil {
		return err
	}
	return writer.Close()
}

func (s *Simulator) writeActionFlowGraphBasic() error {
	s.require()
	snapshot := s.getActionFlowGraphSnapshot(actionFlowGraphReduce)
	writer, err := NewDotActionWriter(s.ownedTool().GetRootName()+"_actions.dot", "")
	if err != nil {
		return err
	}
	for id, action := range snapshot.actions {
		if err := writer.WriteAction(action, id); err != nil {
			return err
		}
	}
	if err := writeActionFlowGraphEdges(writer, snapshot); err != nil {
		return err
	}
	return writer.Close()
}

func writeActionFlowGraphEdges(writer *DotActionWriter, snapshot *actionFlowGraphSnapshot) error {
	length := len(snapshot.actions)
	for i := 0; i < length; i++ {
		for j := 0; j < length; j++ {
			count := simulationActionStatsCell(simulationActionStatsRow(snapshot.actionStats, i), j)
			if count > 0 {
				if err := writer.WriteEdge(i, j, actionFlowGraphWeight(count)); err != nil {
					return err
				}
			} else {
				if snapshot.actions[j] == nil {
					panic(NewNullPointerException())
				}
				if snapshot.actions[j].IsInitPredicate() {
					continue
				}
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
