package tlc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type RunMode int

const (
	RunModeModelCheck RunMode = iota
	RunModeSimulate
)

const (
	tlcNoSuspendProperty        = "tlc2.TLC.nosuspend"
	tlcNoHaltProperty           = "tlc2.TLC.nohalt"
	tlcStopAfterProperty        = "tlc2.TLC.stopAfter"
	modelCheckerVetoProperty    = "tlc2.tool.ModelChecker.vetoCleanup"
	modelCheckerBAQueueProperty = "tlc2.tool.ModelChecker.BAQueue"
)

type Options struct {
	Tool                      *Tool
	SpecFile                  string
	ConfigFile                string
	MetaDir                   string
	FromCheckpoint            string
	Mode                      RunMode
	Workers                   int
	Deadlock                  bool
	NoDeadlock                bool
	Cleanup                   bool
	CleanupPrecleanDone       bool
	NoSeed                    bool
	Seed                      int64
	Aril                      int64
	TraceDepth                int
	TraceNum                  int64
	TraceNumSet               bool
	TraceFile                 string
	TraceActions              string
	Probabilistic             bool
	SimulationSchedule        SimulationSchedule
	DFIDDepth                 int
	CheckpointDurationMillis  int64
	StateWriter               *StateWriter
	FPSet                     FPSet
	FPIndex                   int
	FPSetConfiguration        *FPSetConfiguration
	StateQueue                StateQueue
	Trace                     *TLCTrace
	LiveCheck                 *LiveCheck
	StartTime                 time.Time
	StopAfter                 time.Duration
	GenerateTraceSpec         bool
	ForceGenerateTraceSpec    bool
	GenerateTraceSpecBinary   bool
	GenerateTraceSpecMonolith bool
	TraceSpecOutputDir        string
	ToolMode                  bool
	DebugPort                 int
	DebugPortSet              bool
	DebugSuspend              bool
	DebugHalt                 bool
	UserOutput                *os.File
	RuntimeParams             RuntimeParameters
}

type SimulationSchedule int

const (
	SimulationScheduleRandom SimulationSchedule = iota
	SimulationScheduleRL
	SimulationScheduleRLAction
)

type Result struct {
	ExitStatus      int
	ErrorCode       int
	StatesGenerated int64
	DistinctStates  uint64
	InitialStates   int64
	QueueSize       int64
	SearchDepth     int64
	TraceCount      int64
	Messages        []Message
}

type TLC struct {
	Options
	WelcomePrinted bool
}

func NewTLC(opts Options) *TLC {
	if opts.MetaDir == "" {
		opts.MetaDir = "states"
	}
	if opts.TraceDepth == 0 {
		opts.TraceDepth = 100
	}
	if opts.TraceNum == 0 && !opts.TraceNumSet {
		opts.TraceNum = math.MaxInt64
	}
	if opts.Workers <= 0 {
		opts.Workers = 1
	}
	if opts.FPSetConfiguration == nil {
		opts.FPSetConfiguration = NewFPSetConfiguration()
	}
	if opts.DebugPort == 0 && !opts.DebugPortSet {
		opts.DebugPort = -1
	}
	if opts.NoDeadlock {
		opts.Deadlock = false
	} else if !opts.Deadlock {
		opts.Deadlock = true
	}
	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}
	if opts.StopAfter == 0 {
		opts.StopAfter = stopAfterDurationFromEnv()
	}
	return &TLC{Options: opts}
}

func (t *TLC) GetSpecName() string {
	if t == nil || t.SpecFile == "" {
		return "N/A"
	}
	return t.SpecFile
}

func (t *TLC) GetModelName() string {
	if t == nil || t.ConfigFile == "" {
		return "N/A"
	}
	return t.ConfigFile
}

func ModelCheck(ctx context.Context, opts Options) (*Result, error) {
	opts.Mode = RunModeModelCheck
	return NewTLC(opts).Process(ctx)
}

func Simulate(ctx context.Context, opts Options) (*Result, error) {
	opts.Mode = RunModeSimulate
	return NewTLC(opts).Process(ctx)
}

func (t *TLC) Process(ctx context.Context) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if t == nil {
		t = NewTLC(Options{})
	}
	if t.Tool == nil {
		if t.UserOutput != nil {
			_ = t.UserOutput.Close()
		}
		return &Result{ExitStatus: ExitStatusError, ErrorCode: ECGeneral}, newTLCError(ECGeneral, "TLC runner has no tool")
	}
	t.attachDebuggerIfRequested()
	closeUserOutput := t.installUserOutput()

	t.prepareRandomSeed()
	t.applyGlobals()
	if t.Cleanup && !t.CleanupPrecleanDone && t.FromCheckpoint == "" {
		deleteDirLikeJava(t.MetaDir, true)
	}
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	var traceRecorder *ErrorTraceMessageRecorder
	if t.GenerateTraceSpec {
		traceRecorder = NewErrorTraceMessageRecorder()
		AddMessageRecorder(traceRecorder)
		defer RemoveMessageRecorder(traceRecorder)
	}

	PrintMessage(ECTLCStarting)
	var result *Result
	var err error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = panicValueAsError(recovered)
				code, params := tlcProcessFailureMessage(err)
				PrintError(code, params...)
				result = &Result{ErrorCode: code}
			}
		}()
		switch t.Mode {
		case RunModeSimulate:
			result, err = t.processSimulation()
		default:
			result, err = t.processModelChecking()
		}
	}()
	if result == nil {
		result = &Result{ErrorCode: ECGeneral}
	}
	if err != nil && result.ErrorCode == NoError {
		result.ErrorCode = ECGeneral
	}
	if closeErr := closeUserOutput(); closeErr != nil && err == nil {
		err = closeErr
		if result.ErrorCode == NoError {
			result.ErrorCode = ECGeneral
		}
	}
	if traceRecorder != nil && t.FromCheckpoint == "" {
		if mcError, ok := traceRecorder.MCErrorTrace(); ok {
			outputDir := t.TraceSpecOutputDir
			if outputDir == "" {
				outputDir = "."
			}
			teSpec := NewTraceExplorationSpec(outputDir, time.Now(), t.Tool.GetRootName())
			if _, genErr := teSpec.Generate(t.Tool, mcError); genErr != nil && err == nil {
				err = genErr
				if result.ErrorCode == NoError {
					result.ErrorCode = ECGeneral
				}
			}
		}
	}
	result.ExitStatus = ExitStatusForErrorCode(result.ErrorCode)
	PrintMessage(ECTLCFinished, t.finishedRuntime())
	recorder.mu.Lock()
	result.Messages = append([]Message(nil), recorder.Messages...)
	recorder.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, err
}

func (t *TLC) attachDebuggerIfRequested() {
	if t == nil || t.Tool == nil {
		return
	}
	if !t.DebugPortSet && t.DebugPort < 0 {
		return
	}
	t.Tool.AttachDebugger(t.DebugPort, t.DebugSuspend, t.DebugHalt)
}

func (t *TLC) installUserOutput() func() error {
	if t == nil || t.UserOutput == nil {
		return func() error { return nil }
	}
	file := t.UserOutput
	previous := TLCOutput
	previousUserFile := TLCOutputToUserFile
	TLCOutput = file
	TLCOutputToUserFile = true
	closed := false
	return func() error {
		if closed {
			return nil
		}
		closed = true
		if TLCOutput == file {
			TLCOutput = previous
		}
		TLCOutputToUserFile = previousUserFile
		return file.Close()
	}
}

func panicValueAsError(value any) error {
	switch v := value.(type) {
	case nil:
		return nil
	case error:
		return v
	case string:
		return errors.New(v)
	default:
		return fmt.Errorf("%v", v)
	}
}

func tlcProcessFailureMessage(err error) (int, []string) {
	var eval *EvalException
	if errors.As(err, &eval) && eval != nil {
		code := eval.GetErrorCode()
		if eval.HasParameters() {
			return code, eval.GetParameters()
		}
		if code == ECGeneral {
			return code, generalErrorParams("", err)
		}
		return code, nil
	}
	var tlcErr *TLCError
	if errors.As(err, &tlcErr) && tlcErr != nil {
		switch tlcErr.Code {
		case ECSystemStackOverflow, ECSystemOutOfMemory, ECTLCBug:
			return tlcErr.Code, nil
		}
	}
	return ECGeneral, generalErrorParams("", err)
}

func (t *TLC) applyGlobals() {
	if t.FPIndex >= 0 && t.FPIndex < len(FP64Polys) {
		FP64InitIndex(t.FPIndex)
	}
	Globals.Lock()
	Globals.NumWorkers = t.Workers
	Globals.StartTime = t.StartTime
	Globals.LastCheckpoint = t.StartTime
	Globals.MetaDir = t.MetaDir
	Globals.Tool = t.ToolMode
	if t.CheckpointDurationMillis > 0 {
		Globals.CheckpointDurationMillis = t.CheckpointDurationMillis
	}
	if t.DFIDDepth > 0 {
		Globals.DFIDMax = t.DFIDDepth
	}
	Globals.Unlock()
	SetRandomEnumerableSeed(t.Seed)
}

func (t *TLC) prepareRandomSeed() {
	if t == nil || !t.NoSeed {
		return
	}
	rng := NewJavaRandomDefault()
	t.Seed = rng.NextLong()
	t.Aril = 0
}

func (t *TLC) processModelChecking() (*Result, error) {
	if t.DFIDDepth > 0 {
		PrintMessage(ECTLCModeMCDFS, t.modelCheckingRuntimeParams()...)
		opts := make([]DFIDModelCheckerOption, 0, 2)
		if t.FromCheckpoint != "" {
			opts = append(opts, WithDFIDFromCheckpoint(t.FromCheckpoint))
		}
		if t.LiveCheck != nil {
			opts = append(opts, WithDFIDLiveCheck(t.LiveCheck))
		}
		checker := NewDFIDModelChecker(t.Tool, t.MetaDir, t.Deadlock, opts...)
		code, err := checker.ModelCheck()
		result := &Result{
			ErrorCode:       code,
			StatesGenerated: checker.StatesGenerated,
			DistinctStates:  checker.FPSet.Size(),
			InitialStates:   int64(len(checker.InitStates)),
		}
		return result, err
	}

	PrintMessage(ECTLCModeMC, t.modelCheckingRuntimeParams()...)
	opts := make([]ModelCheckerOption, 0, 5)
	if t.FPSet != nil {
		opts = append(opts, WithModelCheckerFPSet(t.FPSet))
	} else if t.FPSetConfiguration != nil {
		opts = append(opts, WithModelCheckerFPSet(NewFPSet(t.FPSetConfiguration)))
	}
	if t.StateQueue != nil {
		opts = append(opts, WithModelCheckerStateQueue(t.StateQueue))
	}
	if t.StateWriter != nil {
		opts = append(opts, WithModelCheckerStateWriter(t.StateWriter))
	}
	if t.Trace != nil {
		opts = append(opts, WithModelCheckerTrace(t.Trace))
	}
	if t.LiveCheck != nil {
		opts = append(opts, WithModelCheckerLiveCheck(t.LiveCheck))
	}
	if t.FromCheckpoint != "" {
		opts = append(opts, WithModelCheckerFromCheckpoint(t.FromCheckpoint))
	}
	checker := NewModelChecker(t.Tool, t.MetaDir, t.Deadlock, opts...)
	checker.TimeBound = t.StopAfter > 0
	cancelStopAfter := t.scheduleStopAfter(checker.Stop)
	defer cancelStopAfter()
	code, err := checker.ModelCheck()
	result := &Result{
		ErrorCode:       code,
		StatesGenerated: checker.GetStatesGenerated(),
		DistinctStates:  checker.GetDistinctStatesGenerated(),
		InitialStates:   checker.GetInitialStatesGenerated(),
		QueueSize:       checker.GetStateQueueSize(),
		SearchDepth:     checker.GetProgress(),
	}
	return result, err
}

func (t *TLC) scheduleStopAfter(stop func()) func() {
	if t == nil || t.StopAfter <= 0 || stop == nil {
		return func() {}
	}
	timer := time.AfterFunc(t.StopAfter, stop)
	return func() {
		timer.Stop()
	}
}

func stopAfterDurationFromEnv() time.Duration {
	if value, ok := tlcLookupSystemProperty(tlcStopAfterProperty); ok {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
		return 0
	}
	if value := os.Getenv("TLAGO_STOP_AFTER"); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 0
}

func defaultTLCDebugSuspend() bool {
	if value, ok := tlcLookupSystemProperty(tlcNoSuspendProperty); ok {
		return !javaBooleanProperty(value)
	}
	return true
}

func defaultTLCDebugHalt() bool {
	if value, ok := tlcLookupSystemProperty(tlcNoHaltProperty); ok {
		return !javaBooleanProperty(value)
	}
	return true
}

func (t *TLC) processSimulation() (*Result, error) {
	PrintMessage(ECTLCModeSimu, t.simulationRuntimeParams()...)
	simulator := NewSimulator(t.Tool, t.Deadlock, t.TraceDepth, t.TraceNum, t.Seed,
		WithSimulatorTraceFile(t.TraceFile),
		WithSimulatorTraceActions(t.TraceActions),
		WithSimulatorMetaDir(t.MetaDir),
		WithSimulatorSchedule(t.SimulationSchedule),
		WithSimulatorAril(t.Aril),
		WithSimulatorLiveCheck(t.LiveCheck),
	)
	SetRandomEnumerableSeed(simulator.Seed)
	cancelStopAfter := t.scheduleStopAfter(simulator.Stop)
	defer cancelStopAfter()
	code, err := simulator.Simulate()
	return &Result{
		ErrorCode:       code,
		StatesGenerated: simulator.StatesGenerated,
		TraceCount:      simulator.TracesGenerated,
		SearchDepth:     int64(t.TraceDepth),
	}, err
}

func (t *TLC) finishedRuntime() string {
	if t == nil || t.StartTime.IsZero() {
		return "0s"
	}
	elapsed := time.Since(t.StartTime)
	if t.ToolMode || javaBooleanPropertyValue(tlcSystemPropertyOrEnv("tlc2.TLC.asMilliSeconds")) {
		return fmt.Sprintf("%dms", elapsed.Milliseconds())
	}
	return humanReadableTLCRuntime(elapsed)
}

func humanReadableTLCRuntime(elapsed time.Duration) string {
	if elapsed < 0 {
		elapsed = 0
	}
	totalSeconds := int64(elapsed / time.Second)
	days := totalSeconds / 86400
	hours := (totalSeconds / 3600) % 24
	minutes := (totalSeconds / 60) % 60
	seconds := totalSeconds % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %02dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%02dh %02dmin", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%02dmin %02ds", minutes, seconds)
	default:
		return fmt.Sprintf("%02ds", seconds)
	}
}

func (t *TLC) modelCheckingRuntimeParams() []string {
	workers := NumWorkers()
	cfg := t.FPSetConfiguration
	if cfg == nil {
		cfg = NewFPSetConfiguration()
	}
	return []string{
		fmt.Sprintf("%d", workers),
		pluralSuffix(workers),
		fmt.Sprintf("%d", runtime.NumCPU()),
		runtime.GOOS,
		"",
		runtime.GOARCH,
		"Go",
		runtime.Version(),
		runtime.GOARCH,
		fmt.Sprintf("%d", runtimeHeapMB()),
		"0",
		fmt.Sprintf("%d", RandomEnumerableSeed()),
		fmt.Sprintf("%d", t.FPIndex),
		fmt.Sprintf("%d", os.Getpid()),
		simpleJavaName(cfg.GetImplementation()),
		GetStateQueueName(),
	}
}

func (t *TLC) simulationRuntimeParams() []string {
	workers := NumWorkers()
	return []string{
		fmt.Sprintf("%d", t.Seed),
		fmt.Sprintf("%d", workers),
		pluralSuffix(workers),
		fmt.Sprintf("%d", runtime.NumCPU()),
		runtime.GOOS,
		"",
		runtime.GOARCH,
		"Go",
		runtime.Version(),
		runtime.GOARCH,
		fmt.Sprintf("%d", runtimeHeapMB()),
		"0",
		fmt.Sprintf("%d", os.Getpid()),
		t.simulationScheduleName(),
	}
}

func (t *TLC) simulationScheduleName() string {
	if t == nil {
		return "Random"
	}
	switch t.SimulationSchedule {
	case SimulationScheduleRL:
		return "RL"
	default:
		return "Random"
	}
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func runtimeHeapMB() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Sys / 1024 / 1024
}

func simpleJavaName(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

func tlcSystemPropertyOrEnv(name string) string {
	if value, ok := tlcLookupSystemProperty(name); ok {
		return value
	}
	return ""
}

func javaBooleanPropertyValue(value string) bool {
	if value == "" {
		return false
	}
	return javaBooleanProperty(value)
}
