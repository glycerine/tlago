package tlc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RunMode int

const (
	RunModeModelCheck RunMode = iota
	RunModeSimulate
)

func (m RunMode) String() string {
	switch m {
	case RunModeModelCheck:
		return "MODEL_CHECK"
	case RunModeSimulate:
		return "SIMULATE"
	default:
		return fmt.Sprintf("TLC.RunMode(%d)", m)
	}
}

const (
	tlcNoSuspendProperty        = "tlc2.TLC.nosuspend"
	tlcNoHaltProperty           = "tlc2.TLC.nohalt"
	tlcStopAfterProperty        = "tlc2.TLC.stopAfter"
	modelCheckerVetoProperty    = "tlc2.tool.ModelChecker.vetoCleanup"
	modelCheckerBAQueueProperty = "tlc2.tool.ModelChecker.BAQueue"
)

type Options struct {
	Tool                      *Tool
	LoadTool                  func() (*Tool, error)
	Resolver                  FilenameToStream
	SpecFile                  string
	ConfigFile                string
	PackagedModel             *ModelInJar
	MetaDir                   string
	FromCheckpoint            *string
	Mode                      RunMode
	Workers                   int
	Deadlock                  bool
	NoDeadlock                bool
	Cleanup                   bool
	CleanupPrecleanDone       bool
	NoSeed                    bool
	Seed                      int64
	SeedSet                   bool
	Aril                      int64
	TraceDepth                int
	TraceDepthSet             bool
	TraceNum                  int64
	TraceNumSet               bool
	TraceFile                 *string
	TraceActions              *string
	Probabilistic             bool
	SimulationSchedule        SimulationSchedule
	DFIDMode                  bool
	DFIDModeSet               bool
	DFIDDepth                 int
	CheckpointDurationMillis  int64
	CheckpointDurationSet     bool
	StateWriter               IStateWriter
	StateWriterSet            bool
	FPSet                     FPSet
	FPIndex                   int
	FPIndexSet                bool
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
	TraceSpecModuleName       string
	ToolMode                  bool
	ToolModeSet               bool
	DebugPort                 int
	DebugPortSet              bool
	DebugSuspend              bool
	DebugSuspendSet           bool
	DebugHalt                 bool
	DebugHaltSet              bool
	UserOutput                *os.File
	userOutputOwner           *bufferedUserOutput
	RuntimeParams             RuntimeParameters
}

type bufferedUserOutput struct {
	mu     sync.Mutex
	file   *os.File
	writer *JavaBufferedWriter
	closed bool
}

func newBufferedUserOutput(file *os.File) *bufferedUserOutput {
	return &bufferedUserOutput{
		file:   file,
		writer: NewJavaBufferedWriter(file),
	}
}

func (o *bufferedUserOutput) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return 0, NewIOException("Stream closed")
	}
	written, err := o.writer.Write(p)
	return written, bufferedRandomAccessFileIOError(err)
}

func (o *bufferedUserOutput) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return nil
	}
	flushErr := bufferedRandomAccessFileIOError(o.writer.Flush())
	closeErr := bufferedRandomAccessFileIOError(o.file.Close())
	o.closed = true
	if flushErr != nil {
		javaSuppressCloseError(flushErr, closeErr)
		return flushErr
	}
	return closeErr
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
	if opts.TraceDepth == 0 && !opts.TraceDepthSet {
		opts.TraceDepth = 100
	}
	if opts.TraceNum == 0 && !opts.TraceNumSet {
		opts.TraceNum = math.MaxInt64
	}
	if !opts.SeedSet && !opts.NoSeed && opts.Seed == 0 {
		opts.NoSeed = true
	}
	if opts.Workers <= 0 {
		opts.Workers = NumWorkers()
	}
	if opts.FPSetConfiguration == nil {
		opts.FPSetConfiguration = NewFPSetConfiguration()
	}
	if !opts.FPIndexSet && opts.FPIndex == 0 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		opts.FPIndex = rng.Intn(len(FP64Polys))
	}
	if isNil(opts.StateWriter) && !opts.StateWriterSet {
		opts.StateWriter = NewNoopStateWriter()
	}
	if opts.DFIDDepth > 0 {
		opts.DFIDMode = true
	}
	if !opts.DFIDModeSet && !opts.DFIDMode {
		Globals.Lock()
		depth := Globals.DFIDMax
		Globals.Unlock()
		if depth >= 0 {
			opts.DFIDMode = true
			opts.DFIDDepth = depth
		}
	}
	if opts.DebugPort == 0 && !opts.DebugPortSet {
		opts.DebugPort = -1
	}
	if !opts.DebugSuspendSet {
		opts.DebugSuspend = defaultTLCDebugSuspend()
	}
	if !opts.DebugHaltSet {
		opts.DebugHalt = defaultTLCDebugHalt()
	}
	if opts.NoDeadlock {
		opts.Deadlock = false
	} else if !opts.Deadlock {
		opts.Deadlock = true
	}
	if opts.StopAfter == 0 {
		opts.StopAfter = stopAfterDurationFromEnv()
	}
	if !opts.ToolModeSet {
		Globals.Lock()
		opts.ToolMode = Globals.Tool
		Globals.Unlock()
	}
	runner := &TLC{Options: opts}
	if !isNil(opts.Resolver) {
		runner.SetResolver(opts.Resolver)
	}
	return runner
}

func (t *TLC) GetSpecName() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	return tlcGetSystemProperty("specName", t.SpecFile)
}

func (t *TLC) GetModelName() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	return tlcGetSystemProperty("modelName", t.SpecFile)
}

func (t *TLC) SetResolver(resolver FilenameToStream) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.Resolver = resolver
	ToolIOSetDefaultResolver(resolver)
}

func (t *TLC) GetResolver() FilenameToStream {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.Resolver
}

func (t *TLC) SetStateWriter(writer IStateWriter) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.StateWriter = writer
	t.StateWriterSet = true
}

func (t *TLC) GetStateWriter() IStateWriter {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.StateWriter
}

func (t *TLC) GetRunMode() RunMode {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.Mode
}

func (t *TLC) GetMainFile() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.SpecFile
}

func (t *TLC) GetStartTime() int64 {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.StartTime.IsZero() {
		return 0
	}
	return t.StartTime.UnixMilli()
}

func (t *TLC) GetModuleFiles() []*TLAFile {
	if t == nil {
		panic(NewNullPointerException())
	}
	files := []*TLAFile{}
	if checker, ok := asModelChecker(MainChecker()); ok {
		files = append(files, checker.GetModuleFiles(t.Resolver)...)
		if t.PackagedModel != nil && t.PackagedModel.HasCfg() {
			files = append(files, t.PackagedModel.GetCfg())
		}
	}
	return files
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
	if t == nil {
		panic(NewNullPointerException())
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if t.Tool == nil && t.LoadTool == nil {
		_ = t.closeUserOutput()
		return &Result{ExitStatus: ExitStatusError, ErrorCode: ECGeneral}, newTLCError(ECGeneral, "TLC runner has no tool")
	}
	if t.StartTime.IsZero() {
		t.StartTime = time.Now()
	}
	t.prepareToolMode()
	t.attachDebuggerIfRequested()
	closeUserOutput := t.installUserOutput()
	var modelCheckerMXWrapper interface{ Unregister() bool } = NewNullTLCStandardMBean()

	if t.Cleanup && !t.CleanupPrecleanDone && t.FromCheckpoint == nil {
		deleteDirLikeJava(t.MetaDir, true)
	}
	// Java's handleParameters owns metadata creation before process. The
	// native Options API with an injected Tool bypasses that boundary, so
	// this adapter owns its supplied metadata directory after precleaning.
	// Parsed commands and recovery retain their already prepared paths.
	if t.Tool != nil && t.LoadTool == nil && t.FromCheckpoint == nil {
		if err := os.MkdirAll(t.MetaDir, 0o755); err != nil {
			closeUserOutput()
			return &Result{ExitStatus: ExitStatusError, ErrorCode: ECGeneral}, err
		}
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

	var result *Result
	var err error
	processExited := false
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = panicValueAsError(recovered)
				if exit, ok := err.(*ProcessExit); ok {
					processExited = true
					result = &Result{ErrorCode: exit.ErrorCode}
					return
				}
				code, params := tlcProcessFailureMessage(err)
				if failure := javaRuntimeException(err); failure != nil {
					PrintTLCRuntimeException(failure)
				} else {
					PrintError(code, params...)
				}
				result = &Result{ErrorCode: code}
			}
		}()
		// TLC.process restores intern identities before constructing FastTool.
		// Parsed strings and recovered values must share the checkpoint tokens.
		if t.FromCheckpoint != nil {
			if err := RecoverUniqueStrings(*t.FromCheckpoint); err != nil {
				panic(err)
			}
		}
		// Recovery must succeed before TLC initializes FP64 or random values.
		t.applyGlobals()
		// TLC.process prints the mode banner before constructing FastTool,
		// including when configuration-time constant evaluation fails.
		if t.Mode == RunModeSimulate {
			runtimeValues := t.simulationRuntimeValues()
			PrintMessage(ECTLCModeSimu, simulationRuntimeParams(runtimeValues)...)
			t.collectExecutionStatistics("simulation", runtimeValues)
		} else {
			runtimeValues := t.modelCheckingRuntimeValues()
			if t.DFIDMode {
				PrintMessage(ECTLCModeMCDFS, modelCheckingRuntimeParams(runtimeValues)...)
				t.collectExecutionStatistics("dfs", runtimeValues)
			} else {
				PrintMessage(ECTLCModeMC, modelCheckingRuntimeParams(runtimeValues)...)
				t.collectExecutionStatistics("bfs", runtimeValues)
			}
		}
		if t.Tool == nil {
			mode := ModeMC
			if t.Mode == RunModeSimulate {
				mode = ModeSimulation
			}
			if t.DebugPortSet || t.DebugPort >= 0 {
				mode = ModeDebugger
			}
			prepareTLCStateRuntimeMode(mode)
			t.Tool, err = t.LoadTool()
			if err != nil {
				panic(err)
			}
			if t.Tool == nil {
				panic(newTLCError(ECGeneral, "TLC runner has no tool"))
			}
			t.prepareToolMode()
			t.attachDebuggerIfRequested()
		}
		switch t.Mode {
		case RunModeSimulate:
			result, err = t.processSimulation()
		default:
			result, err = t.processModelChecking(func(wrapper *ModelCheckerMXWrapper) {
				modelCheckerMXWrapper = wrapper
			})
		}
	}()
	if result == nil {
		result = &Result{ErrorCode: ECGeneral}
	}
	if err != nil && result.ErrorCode == NoError {
		result.ErrorCode = ECGeneral
	}
	if processExited {
		// Source System.exit does not unwind TLC.process's finally: no OUTPUT
		// cleanup, finished message or trace generation follows the diagnostic.
		result.ExitStatus = ExitStatusForErrorCode(result.ErrorCode)
		recorder.mu.Lock()
		result.Messages = append([]Message(nil), recorder.Messages...)
		recorder.mu.Unlock()
		return result, err
	}
	// TLC.process's finally ignores OUTPUT flush/close IOException, prints the
	// finished message, then generates a trace spec without changing its result.
	_ = closeUserOutput()
	modelCheckerMXWrapper.Unregister()
	PrintMessage(ECTLCFinished, t.finishedRuntime())
	if traceRecorder != nil && t.Tool != nil {
		if mcError, ok := traceRecorder.GetMCErrorTrace(); ok {
			outputDir := t.TraceSpecOutputDir
			if outputDir == "" {
				outputDir = "."
			}
			teSpec := NewTraceExplorationSpec(outputDir, t.StartTime, t.Tool.GetRootName())
			if t.TraceSpecModuleName != "" {
				teSpec = NewTraceExplorationSpecNamed(outputDir, t.TraceSpecModuleName, t.Tool.GetRootName())
			}
			_, _ = teSpec.Generate(t.Tool, mcError)
		}
	}
	result.ExitStatus = ExitStatusForErrorCode(result.ErrorCode)
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

func (t *TLC) prepareToolMode() {
	if t == nil || t.Tool == nil || t.Tool.GetMode() == ModeDebugger || t.Tool.GetMode() == ModeExecutor {
		return
	}
	mode := ModeMC
	if t.Mode == RunModeSimulate {
		mode = ModeSimulation
	}
	t.Tool.SetMode(mode)
}

func (t *TLC) installUserOutput() func() error {
	if t == nil || t.UserOutput == nil {
		return func() error { return nil }
	}
	var output io.Writer = t.UserOutput
	closeOutput := t.UserOutput.Close
	if owner := t.userOutputOwner; owner != nil && owner.file == t.UserOutput {
		output = owner
		closeOutput = owner.Close
	}
	previous := TLCOutput
	previousUserFile := TLCOutputToUserFile
	TLCOutput = output
	TLCOutputToUserFile = true
	closed := false
	return func() error {
		if closed {
			return nil
		}
		closed = true
		if TLCOutput == output {
			TLCOutput = previous
		}
		TLCOutputToUserFile = previousUserFile
		return closeOutput()
	}
}

func (t *TLC) closeUserOutput() error {
	if t == nil || t.UserOutput == nil {
		return nil
	}
	if owner := t.userOutputOwner; owner != nil && owner.file == t.UserOutput {
		return owner.Close()
	}
	return t.UserOutput.Close()
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
	if code := javaSystemFailureCode(err); code == ECSystemStackOverflow || code == ECSystemOutOfMemory {
		return code, nil
	}
	if failure := javaRuntimeException(err); failure != nil {
		return javaRuntimeFailureMessage(failure)
	}
	// TLC.process does not have an EvalException catch; it reports other
	// RuntimeExceptions (including fingerprint failures) through GENERAL.
	return ECGeneral, generalErrorParams("", err)
}

func (t *TLC) applyGlobals() {
	FP64InitIndex(t.FPIndex)
	Globals.Lock()
	Globals.NumWorkers = t.Workers
	Globals.StartTime = t.StartTime
	Globals.MetaDir = t.MetaDir
	Globals.Tool = t.ToolMode
	if t.CheckpointDurationSet || t.CheckpointDurationMillis > 0 {
		Globals.CheckpointDurationMillis = t.CheckpointDurationMillis
	}
	if t.DFIDMode {
		Globals.DFIDMax = t.DFIDDepth
	} else if t.DFIDModeSet {
		Globals.DFIDMax = -1
	}
	Globals.Unlock()
	t.prepareRandomSeed()
	SetRandomEnumerableSeed(t.Seed)
}

func (t *TLC) prepareRandomSeed() {
	if t == nil || !t.NoSeed {
		return
	}
	rng := NewJavaRandomDefault()
	t.Seed = rng.NextLong()
}

func (t *TLC) processModelChecking(registerMXWrapper func(*ModelCheckerMXWrapper)) (*Result, error) {
	checkDeadlock := t.Deadlock
	if checkDeadlock {
		checkDeadlock = t.Tool.GetModelConfig().GetCheckDeadlock()
	}
	if t.DFIDMode {
		opts := make([]DFIDModelCheckerOption, 0, 4)
		opts = append(opts, WithDFIDStartTime(t.StartTime))
		if t.FromCheckpoint != nil {
			opts = append(opts, WithDFIDFromCheckpoint(t.FromCheckpoint))
		}
		if t.StateWriterSet && isNil(t.StateWriter) {
			opts = append(opts, func(checker *DFIDModelChecker) {
				checker.AllStateWriter = t.StateWriter
			})
		} else if !isNil(t.StateWriter) {
			opts = append(opts, WithDFIDStateWriter(t.StateWriter))
		}
		if t.LiveCheck != nil {
			opts = append(opts, WithDFIDLiveCheck(t.LiveCheck))
		}
		checker := NewDFIDModelChecker(t.Tool, t.MetaDir, checkDeadlock, opts...)
		SetMainChecker(checker)
		code, err := checker.ModelCheck()
		result := &Result{
			ErrorCode:       code,
			StatesGenerated: checker.generatedStates(),
			DistinctStates:  checker.FPSet.Size(),
			InitialStates:   int64(len(checker.InitStates)),
		}
		return result, err
	}

	opts := make([]ModelCheckerOption, 0, 6)
	opts = append(opts, WithModelCheckerStartTime(t.StartTime))
	if !isNil(t.FPSet) {
		opts = append(opts, WithModelCheckerFPSet(t.FPSet))
	} else if t.FPSetConfiguration != nil {
		opts = append(opts, WithModelCheckerFPSetConfiguration(t.FPSetConfiguration))
	}
	if !isNil(t.StateQueue) {
		opts = append(opts, WithModelCheckerStateQueue(t.StateQueue))
	}
	if t.StateWriterSet && isNil(t.StateWriter) {
		opts = append(opts, func(checker *ModelChecker) {
			checker.AllStateWriter = t.StateWriter
		})
	} else if !isNil(t.StateWriter) {
		opts = append(opts, WithModelCheckerStateWriter(t.StateWriter))
	}
	if t.Trace != nil {
		opts = append(opts, WithModelCheckerTrace(t.Trace))
	}
	if t.LiveCheck != nil {
		opts = append(opts, WithModelCheckerLiveCheck(t.LiveCheck))
	}
	if t.FromCheckpoint != nil {
		opts = append(opts, WithModelCheckerFromCheckpoint(t.FromCheckpoint))
	}
	checker := NewModelChecker(t.Tool, t.MetaDir, checkDeadlock, opts...)
	SetMainChecker(checker)
	registerMXWrapper(NewModelCheckerMXWrapper(checker, t))
	if t.StopAfter > 0 {
		checker.TimeBound = true
	}
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
	if value := os.Getenv("TLAGO_STOP_AFTER"); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 0
}

func stopAfterFromJavaProperty() (int64, bool) {
	value, ok := tlcLookupSystemProperty(tlcStopAfterProperty)
	if !ok {
		return 0, false
	}
	seconds, ok := javaDecodeLongProperty(value)
	if !ok {
		return 0, false
	}
	timeBound := seconds != -1
	if seconds <= 0 {
		return 0, timeBound
	}
	return seconds, timeBound
}

func scheduleStopAfterFromJavaProperty(stop func()) bool {
	seconds, timeBound := stopAfterFromJavaProperty()
	if seconds <= 0 {
		return timeBound
	}

	// AbstractChecker.scheduleTermination multiplies in signed Java-long
	// milliseconds before Timer validates the delay and absolute execution time.
	delayMillis := seconds * 1000
	if delayMillis < 0 {
		panic(NewIllegalArgumentException("Negative delay."))
	}
	executionMillis := time.Now().UnixMilli() + delayMillis
	if executionMillis < 0 {
		panic(NewIllegalArgumentException("Illegal execution time."))
	}
	if stop != nil {
		scheduleJavaTimerAt(executionMillis, stop)
	}
	return timeBound
}

func scheduleJavaTimerAt(executionMillis int64, run func()) {
	const maxTimerMillis = int64(math.MaxInt64 / int64(time.Millisecond))
	var schedule func()
	schedule = func() {
		delayMillis := executionMillis - time.Now().UnixMilli()
		if delayMillis <= 0 {
			time.AfterFunc(0, run)
			return
		}
		if delayMillis > maxTimerMillis {
			delayMillis = maxTimerMillis
		}
		time.AfterFunc(time.Duration(delayMillis)*time.Millisecond, func() {
			if time.Now().UnixMilli() >= executionMillis {
				run()
				return
			}
			schedule()
		})
	}
	schedule()
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
	// With no explicit seed, source uses setSeed(seed) without advancing aril,
	// while retaining the parsed aril field for later parameter handling.
	aril := t.Aril
	if t.NoSeed {
		aril = 0
	}
	simulator := NewSimulator(t.Tool, t.Deadlock, t.TraceDepth, t.TraceNum, t.Seed,
		WithSimulatorTraceFile(t.TraceFile),
		WithSimulatorTraceActions(t.TraceActions),
		WithSimulatorMetaDir(t.MetaDir),
		WithSimulatorSchedule(t.SimulationSchedule),
		WithSimulatorAril(aril),
		WithSimulatorWorkerCount(t.Workers),
		WithSimulatorLiveCheck(t.LiveCheck),
	)
	SetSimulator(simulator)
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
	elapsedMillis := time.Now().UnixMilli() - t.StartTime.UnixMilli()
	if t.ToolMode || javaBooleanPropertyValue(tlcSystemPropertyOrEnv("tlc2.TLC.asMilliSeconds")) {
		return fmt.Sprintf("%dms", elapsedMillis)
	}
	return ConvertRuntimeToHumanReadable(elapsedMillis)
}

// ConvertRuntimeToHumanReadable retains TLC's strict millisecond boundaries
// and SimpleDateFormat UTC calendar fields (day-of-year, rather than elapsed
// days). The duplicate day branch in Java selects the same pattern.
func ConvertRuntimeToHumanReadable(runtimeMillis int64) string {
	const day = int64(86400000)
	millis := runtimeMillis
	if runtimeMillis > day {
		millis -= day
	}
	date := time.UnixMilli(millis).UTC()
	var result string
	switch {
	case runtimeMillis > day:
		result = fmt.Sprintf("%dd %02dh", date.YearDay(), date.Hour())
	case runtimeMillis > 3600000:
		result = fmt.Sprintf("%02dh %02dmin", date.Hour(), date.Minute())
	case runtimeMillis > 60000:
		result = fmt.Sprintf("%02dmin %02ds", date.Minute(), date.Second())
	default:
		result = fmt.Sprintf("%02ds", date.Second())
	}
	// DateFormat's numeric fields use the process FORMAT locale's zero digit.
	MessageNumberFormat(0) // Initialize the shared source locale symbols.
	if mpNumberSymbols.zero != '0' {
		result = strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return mpNumberSymbols.zero + r - '0'
			}
			return r
		}, result)
	}
	return result
}

func GetSimulationRuntime(seed int64) map[string]string {
	workers := NumWorkers()
	schedule := "Random"
	if javaBooleanPropertyValue(tlcSystemPropertyOrEnv(simulatorRLProperty)) {
		schedule = "RL"
	}
	return map[string]string{
		"seed":          fmt.Sprintf("%d", seed),
		"workers":       fmt.Sprintf("%d", workers),
		"plural":        pluralSuffix(workers),
		"cores":         fmt.Sprintf("%d", runtime.NumCPU()),
		"osName":        runtime.GOOS,
		"osVersion":     "",
		"osArch":        runtime.GOARCH,
		"jvmVendor":     "Go",
		"jvmVersion":    runtime.Version(),
		"jvmArch":       runtime.GOARCH,
		"jvmHeapMem":    fmt.Sprintf("%d", runtimeHeapMB()),
		"jvmOffHeapMem": fmt.Sprintf("%d", tlcRuntimeNonHeapPhysicalMemory()/1024/1024),
		"jvmPid":        fmt.Sprintf("%d", os.Getpid()),
		"sched":         schedule,
	}
}

func GetModelCheckingRuntime(fpIndex int, cfg *FPSetConfiguration) map[string]string {
	if cfg == nil {
		panic(NewNullPointerException())
	}
	workers := NumWorkers()
	return map[string]string{
		"workers":       fmt.Sprintf("%d", workers),
		"plural":        pluralSuffix(workers),
		"cores":         fmt.Sprintf("%d", runtime.NumCPU()),
		"osName":        runtime.GOOS,
		"osVersion":     "",
		"osArch":        runtime.GOARCH,
		"jvmVendor":     "Go",
		"jvmVersion":    runtime.Version(),
		"jvmArch":       runtime.GOARCH,
		"jvmHeapMem":    fmt.Sprintf("%d", runtimeHeapMB()),
		"jvmOffHeapMem": fmt.Sprintf("%d", tlcRuntimeNonHeapPhysicalMemory()/1024/1024),
		"seed":          fmt.Sprintf("%d", RandomEnumerableSeed()),
		"fpidx":         fmt.Sprintf("%d", fpIndex),
		"jvmPid":        fmt.Sprintf("%d", os.Getpid()),
		"fpset":         simpleJavaName(cfg.GetImplementation()),
		"queue":         GetStateQueueName(),
	}
}

func (t *TLC) modelCheckingRuntimeParams() []string {
	return modelCheckingRuntimeParams(t.modelCheckingRuntimeValues())
}

func (t *TLC) modelCheckingRuntimeValues() map[string]string {
	cfg := t.FPSetConfiguration
	if cfg == nil {
		cfg = NewFPSetConfiguration()
	}
	return GetModelCheckingRuntime(t.FPIndex, cfg)
}

func modelCheckingRuntimeParams(runtimeValues map[string]string) []string {
	return []string{
		runtimeValues["workers"], runtimeValues["plural"], runtimeValues["cores"],
		runtimeValues["osName"], runtimeValues["osVersion"], runtimeValues["osArch"],
		runtimeValues["jvmVendor"], runtimeValues["jvmVersion"], runtimeValues["jvmArch"],
		runtimeValues["jvmHeapMem"], runtimeValues["jvmOffHeapMem"], runtimeValues["seed"],
		runtimeValues["fpidx"], runtimeValues["jvmPid"], runtimeValues["fpset"], runtimeValues["queue"],
	}
}

func (t *TLC) simulationRuntimeParams() []string {
	return simulationRuntimeParams(t.simulationRuntimeValues())
}

func (t *TLC) simulationRuntimeValues() map[string]string {
	runtimeValues := GetSimulationRuntime(t.Seed)
	runtimeValues["sched"] = t.simulationScheduleName()
	return runtimeValues
}

func simulationRuntimeParams(runtimeValues map[string]string) []string {
	return []string{
		runtimeValues["seed"], runtimeValues["workers"], runtimeValues["plural"], runtimeValues["cores"],
		runtimeValues["osName"], runtimeValues["osVersion"], runtimeValues["osArch"],
		runtimeValues["jvmVendor"], runtimeValues["jvmVersion"], runtimeValues["jvmArch"],
		runtimeValues["jvmHeapMem"], runtimeValues["jvmOffHeapMem"], runtimeValues["jvmPid"], runtimeValues["sched"],
	}
}

func (t *TLC) collectExecutionStatistics(mode string, runtimeValues map[string]string) {
	NewDefaultExecutionStatisticsCollector().Collect(t.executionStatisticsParameters(mode, runtimeValues))
}

func (t *TLC) executionStatisticsParameters(mode string, runtimeValues map[string]string) *InsMap[string, string] {
	parameters := NewInsMap[string, string]()
	parameters.Set("ver", TLCRevisionOrDev())
	parameters.Set("mode", mode)
	for _, key := range []string{
		"workers", "cores", "osName", "osVersion", "osArch", "jvmVendor",
		"jvmVersion", "jvmArch", "jvmHeapMem", "jvmOffHeapMem", "sched",
		"fpset", "queue",
	} {
		if value, ok := runtimeValues[key]; ok {
			parameters.Set(key, value)
		}
	}
	tool := toolMode()
	parameters.Set("toolbox", strconv.FormatBool(tool))
	ide := "cli"
	if tool {
		ide = "toolbox"
	}
	parameters.Set("ide", tlcGetSystemProperty("tlc2.TLC.ide", ide))
	return parameters
}

func CollectREPLExecutionStatistics() {
	runtimeValues := GetSimulationRuntime(0)
	delete(runtimeValues, "sched")
	NewDefaultExecutionStatisticsCollector().Collect((&TLC{}).executionStatisticsParameters("repl", runtimeValues))
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
	return uint64(tlcRuntimeMaxHeapMemoryBytes() / 1024 / 1024)
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
