package tlc

import (
	"context"
	"math"
	"time"
)

type RunMode int

const (
	RunModeModelCheck RunMode = iota
	RunModeSimulate
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
	NoSeed                    bool
	Seed                      int64
	Aril                      int64
	TraceDepth                int
	TraceNum                  int64
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
	GenerateTraceSpec         bool
	ForceGenerateTraceSpec    bool
	GenerateTraceSpecBinary   bool
	GenerateTraceSpecMonolith bool
	TraceSpecOutputDir        string
	ToolMode                  bool
	DebugPort                 int
	DebugSuspend              bool
	DebugHalt                 bool
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
	if opts.TraceNum == 0 {
		opts.TraceNum = math.MaxInt64
	}
	if opts.Workers <= 0 {
		opts.Workers = 1
	}
	if opts.FPSetConfiguration == nil {
		opts.FPSetConfiguration = NewFPSetConfiguration()
	}
	if opts.DebugPort == 0 {
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
		return &Result{ExitStatus: ExitStatusError, ErrorCode: ECGeneral}, newTLCError(ECGeneral, "TLC runner has no tool")
	}

	t.applyGlobals()
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
	switch t.Mode {
	case RunModeSimulate:
		result, err = t.processSimulation()
	default:
		result, err = t.processModelChecking()
	}
	if result == nil {
		result = &Result{ErrorCode: ECGeneral}
	}
	if err != nil && result.ErrorCode == NoError {
		result.ErrorCode = ECGeneral
	}
	if result.ErrorCode == NoError {
		PrintMessage(ECTLCSuccess)
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
	PrintMessage(ECTLCFinished)
	recorder.mu.Lock()
	result.Messages = append([]Message(nil), recorder.Messages...)
	recorder.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, err
}

func (t *TLC) applyGlobals() {
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
	if t.Seed != 0 {
		SetRandomEnumerableSeed(t.Seed + t.Aril)
	}
}

func (t *TLC) processModelChecking() (*Result, error) {
	PrintMessage(ECTLCModeMC)
	if t.DFIDDepth > 0 {
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
		if cleanupErr := checker.Cleanup(code == NoError, t.Cleanup); err == nil {
			err = cleanupErr
		}
		return result, err
	}

	opts := make([]ModelCheckerOption, 0, 5)
	if t.FPSet != nil {
		opts = append(opts, WithModelCheckerFPSet(t.FPSet))
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
	code, err := checker.ModelCheck()
	result := &Result{
		ErrorCode:       code,
		StatesGenerated: checker.GetStatesGenerated(),
		DistinctStates:  checker.GetDistinctStatesGenerated(),
		InitialStates:   checker.GetInitialStatesGenerated(),
		QueueSize:       checker.GetStateQueueSize(),
		SearchDepth:     checker.GetProgress(),
	}
	if cleanupErr := checker.Cleanup(code == NoError, t.Cleanup); err == nil {
		err = cleanupErr
	}
	return result, err
}

func (t *TLC) processSimulation() (*Result, error) {
	PrintMessage(ECTLCModeSimu)
	seed := t.Seed
	if seed != 0 {
		seed += t.Aril
	}
	simulator := NewSimulator(t.Tool, t.Deadlock, t.TraceDepth, t.TraceNum, seed,
		WithSimulatorTraceFile(t.TraceFile),
		WithSimulatorTraceActions(t.TraceActions),
		WithSimulatorSchedule(t.SimulationSchedule),
		WithSimulatorLiveCheck(t.LiveCheck),
	)
	code, err := simulator.Simulate()
	return &Result{
		ErrorCode:       code,
		StatesGenerated: simulator.StatesGenerated,
		TraceCount:      simulator.TracesGenerated,
		SearchDepth:     int64(t.TraceDepth),
	}, err
}
