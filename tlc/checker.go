package tlc

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const javaIntegerMaxValue = int(1<<31 - 1)

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
	AllStateWriter            IStateWriter
	Workers                   []*Worker
	PrintedLivenessErrorStack bool
	Config                    Value
	LiveCheck                 *LiveCheck
	StartTime                 time.Time
}

func NewAbstractChecker(tool *Tool, metadir string, stateWriter IStateWriter, deadlock bool, fromCheckpoint string, startTime time.Time) *AbstractChecker {
	checker := newAbstractCheckerFields(tool, metadir, stateWriter, deadlock, fromCheckpoint, startTime)
	checker.initialize(checker.Stop)
	return checker
}

func newAbstractCheckerFields(tool *Tool, metadir string, stateWriter IStateWriter, deadlock bool, fromCheckpoint string, startTime time.Time) *AbstractChecker {
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
	checker := &AbstractChecker{
		ErrorCode:      NoError,
		CheckDeadlock:  deadlock,
		CheckLiveness:  checkLiveness,
		FromCheckpoint: fromCheckpoint,
		Metadir:        metadir,
		Tool:           tool,
		AllStateWriter: stateWriter,
		StartTime:      startTime,
	}
	return checker
}

// initialize completes Java's parent constructor before concrete checker storage.
func (c *AbstractChecker) initialize(stop func()) bool {
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(c.Tool)
	}
	if c.CheckLiveness && c.Tool != nil && c.Tool.HasSymmetry() {
		PrintWarning(ECTLCFeatureUnsupportedLivenessSymmetry)
	}
	if c.LiveCheck == nil {
		if c.CheckLiveness {
			DebugPrintMessage("initializing liveness checking")
			var err error
			if livenessTestingImplementationEnabled() {
				c.LiveCheck, err = NewAddAndCheckLiveCheckFromTool(c.Tool, c.Metadir)
			} else {
				c.LiveCheck, err = NewLiveCheckFromTool(c.Tool, c.Metadir, c.AllStateWriter)
			}
			if err != nil {
				panic(err)
			}
			DebugPrintMessage("liveness checking initialized")
		} else {
			c.LiveCheck = NewNoOpLiveCheck(c.Tool, c.Metadir)
		}
	}
	c.Config = c.createConfig()
	return scheduleStopAfterFromJavaProperty(func() {
		defer func() {
			if failure := recover(); failure != nil {
				// Java's Timer thread reports its uncaught throwable and exits;
				// it does not terminate the model-checker's host process.
				_, _ = fmt.Fprint(os.Stderr, "Exception in thread \"TLCStopAfterTimer\" "+javaThrowableStackTrace(panicValueAsError(failure)))
			}
		}()
		stop()
	})
}

func (c *AbstractChecker) Stop() {
	panic(NewUnsupportedOperationException("stop not implemented"))
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
	ResetCurrentState()
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
	ResetCurrentState()
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
	if c == nil || idx < 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	worker := c.workerAt(workerID)
	if worker == nil {
		return nil
	}
	return worker.GetLocalValue(idx)
}

func (c *AbstractChecker) SetAllValues(idx int, value Value) {
	if c == nil || idx < 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, worker := range c.Workers {
		if worker != nil {
			worker.SetLocalValue(idx, value)
		}
	}
}

func (c *AbstractChecker) SetValue(workerID int, idx int, value Value) {
	if c == nil || idx < 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	worker := c.workerAt(workerID)
	if worker != nil {
		worker.SetLocalValue(idx, value)
	}
}

func (c *AbstractChecker) SetWorkerValues(idx int, values []Value) {
	if c == nil || idx < 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, value := range values {
		if worker := c.workerAt(i); worker != nil {
			worker.SetLocalValue(idx, value)
		}
	}
}

func (c *AbstractChecker) GetAllValue(idx int) []Value {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]Value, len(c.Workers))
	for i, worker := range c.Workers {
		values[i] = worker.GetLocalValue(idx)
	}
	return values
}

func (c *AbstractChecker) GetAllValues() Value {
	if c == nil || len(c.Workers) == 0 || c.Workers[0] == nil {
		return EmptyFcn
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	localValues := c.Workers[0].LocalValues
	domain := make([]Value, 0, len(localValues))
	values := make([]Value, 0, len(localValues))
	for idx, value := range localValues {
		if value == nil {
			continue
		}
		workerValues := make([]Value, len(c.Workers))
		for i, worker := range c.Workers {
			if worker != nil {
				workerValues[i] = worker.GetLocalValue(idx)
			}
		}
		domain = append(domain, NewIntValue(int32(idx)))
		values = append(values, NewTupleValue(workerValues))
	}
	return NewFcnRcdValue(domain, values, false)
}

func (c *AbstractChecker) GetNamedValue(workerID int, key *UniqueString) Value {
	if c == nil || key == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	worker := c.workerAt(workerID)
	if worker == nil {
		return nil
	}
	return worker.GetNamedRegister(key)
}

func (c *AbstractChecker) SetAllNamedValues(key *UniqueString, value Value) {
	if c == nil || key == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, worker := range c.Workers {
		if worker != nil {
			worker.SetNamedRegister(key, value)
		}
	}
}

func (c *AbstractChecker) SetNamedValue(workerID int, key *UniqueString, value Value) {
	if c == nil || key == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	worker := c.workerAt(workerID)
	if worker != nil {
		worker.SetNamedRegister(key, value)
	}
}

func (c *AbstractChecker) SetAllNamedWorkerValues(key *UniqueString, values []Value) {
	if c == nil || key == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, value := range values {
		if worker := c.workerAt(i); worker != nil {
			worker.SetNamedRegister(key, value)
		}
	}
}

func (c *AbstractChecker) GetAllNamedRegisterValues() Value {
	if c == nil || len(c.Workers) == 0 || c.Workers[0] == nil || c.Workers[0].NamedRegisters == nil {
		return EmptyFcn
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	domain := make([]Value, 0, c.Workers[0].NamedRegisters.Len())
	values := make([]Value, 0, c.Workers[0].NamedRegisters.Len())
	for key := range c.Workers[0].NamedRegisters.All() {
		workerValues := make([]Value, len(c.Workers))
		for i, worker := range c.Workers {
			if worker != nil {
				workerValues[i] = worker.GetNamedRegister(key)
			}
		}
		domain = append(domain, NewStringValueFromUnique(key))
		values = append(values, NewTupleValue(workerValues))
	}
	return NewFcnRcdValue(domain, values, false)
}

func (c *AbstractChecker) GetAllNamedValues(key *UniqueString) []Value {
	if c == nil || key == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]Value, 0, len(c.Workers))
	for _, worker := range c.Workers {
		if worker == nil {
			values = append(values, nil)
			continue
		}
		values = append(values, worker.GetNamedRegister(key))
	}
	return values
}

func (c *AbstractChecker) workerAt(workerID int) *Worker {
	if c == nil {
		return nil
	}
	if workerID < 0 {
		workerID = 0
	}
	if workerID < 0 || workerID >= len(c.Workers) {
		return nil
	}
	return c.Workers[workerID]
}

type ModelChecker struct {
	*AbstractChecker
	// Java holds the checker monitor throughout error reporting. Keep that
	// serialization separate from mu so aliases and postconditions can query
	// checker state without requiring Java's reentrant locking.
	nextErrorMu             sync.Mutex
	NumberOfInitialStates   int64
	FPSet                   FPSet
	FPSetConfiguration      *FPSetConfiguration
	StateQueue              StateQueue
	Trace                   *TLCTrace
	ConcurrentTrace         *ConcurrentTLCTrace
	NextStatesGenerated     int64
	StatesPerMinute         int64
	DistinctStatesPerMinute int64
	OldNumOfGenStates       int64
	OldFPSetSize            uint64
	RuntimeRatio            float64
	ForceLiveCheck          bool
	TimeBound               bool
	CleanupEnabled          bool
	cleanupDone             bool
}

type ModelCheckerOption func(*ModelChecker)

func WithModelCheckerFPSet(fpSet FPSet) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.FPSet = fpSet
	}
}

func WithModelCheckerFPSetConfiguration(config *FPSetConfiguration) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.FPSetConfiguration = config
	}
}

func WithModelCheckerStateQueue(queue StateQueue) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.StateQueue = queue
	}
}

func WithModelCheckerStateWriter(writer IStateWriter) ModelCheckerOption {
	return func(mc *ModelChecker) {
		if writer == nil {
			writer = NewNoopStateWriter()
		}
		mc.AllStateWriter = writer
	}
}

func WithModelCheckerTrace(trace *TLCTrace) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.Trace = trace
	}
}

func WithModelCheckerLiveCheck(liveCheck *LiveCheck) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.LiveCheck = liveCheck
	}
}

func WithModelCheckerFromCheckpoint(fromCheckpoint string) ModelCheckerOption {
	return func(mc *ModelChecker) {
		if mc.AbstractChecker != nil {
			mc.FromCheckpoint = fromCheckpoint
		}
	}
}

func WithModelCheckerCleanup(cleanup bool) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.CleanupEnabled = cleanup
	}
}

func NewModelChecker(tool *Tool, metadir string, deadlock bool, opts ...ModelCheckerOption) *ModelChecker {
	if tool != nil && tool.GetMode() != ModeDebugger && tool.GetMode() != ModeExecutor {
		tool.SetMode(ModeMC)
	}
	SetTLCStateTool(tool)
	checkDeadlock := deadlock
	if tool != nil {
		if config := tool.GetModelConfig(); config != nil {
			checkDeadlock = deadlock && config.GetCheckDeadlock()
		}
	}
	rootName := "Spec"
	if tool != nil {
		rootName = tool.GetRootName()
	}
	mc := &ModelChecker{
		AbstractChecker: newAbstractCheckerFields(tool, metadir, NewNoopStateWriter(), checkDeadlock, "", time.Now()),
		CleanupEnabled:  true,
	}
	for _, opt := range opts {
		opt(mc)
	}
	if mc.AbstractChecker.initialize(mc.Stop) {
		mc.TimeBound = true
	}
	// ModelChecker's private constructor creates the queue, then trace. Its
	// public constructor initializes the selected fingerprint set and workers.
	if mc.StateQueue == nil {
		mc.StateQueue = NewStateQueue(metadir)
	}
	if !stateQueueInitialized(mc.StateQueue) {
		setStateQueueDir(mc.StateQueue, metadir)
	}
	mc.ConcurrentTrace = NewConcurrentTLCTrace(metadir, rootName)
	mc.ConcurrentTrace.SetTool(tool)
	if mc.Trace == nil {
		mc.Trace = mc.ConcurrentTrace.TLCTrace
	} else {
		mc.Trace.SetCheckpointContext(metadir, rootName)
	}
	mc.Trace.SetTool(tool)
	mc.ConcurrentTrace.TLCTrace = mc.Trace

	if mc.FPSet == nil {
		config := mc.FPSetConfiguration
		if config == nil {
			config = NewFPSetConfiguration()
		}
		mc.FPSet = NewFPSet(config)
	}
	if !fpSetInitialized(mc.FPSet) {
		mc.FPSet = mc.FPSet.Init(NumWorkers(), metadir, rootName)
	}
	mc.initWorkers()
	SetMainChecker(mc)
	return mc
}

func livenessTestingImplementationEnabled() bool {
	if value, ok := tlcLookupSystemProperty("tlc2.tool.liveness.ILiveCheck.testing"); ok {
		return javaBooleanProperty(value)
	}
	return false
}

func (mc *ModelChecker) initWorkers() {
	if mc == nil || len(mc.Workers) > 0 {
		return
	}
	workerCount := NumWorkers()
	if workerCount < 1 {
		workerCount = 1
	}
	for id := 0; id < workerCount; id++ {
		workerTool := mc.Tool
		if id > 0 && workerTool != nil {
			workerTool = workerTool.NoDebug()
		}
		NewModelCheckingWorker(id, mc, workerTool)
	}
}

func (mc *ModelChecker) workerAt(id int) *Worker {
	if mc == nil || id < 0 || id >= len(mc.Workers) {
		return nil
	}
	return mc.Workers[id]
}

func (mc *ModelChecker) Stop() {
	if mc != nil && mc.AbstractChecker != nil {
		mc.SetDone()
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
}

func (mc *ModelChecker) GetTraceInfo(state *TLCStateMut) []*TLCStateInfo {
	if mc == nil || state == nil {
		return nil
	}
	if mc.ConcurrentTrace != nil {
		return mc.ConcurrentTrace.GetTraceFromState(state)
	}
	if mc.Trace != nil {
		return mc.Trace.GetTrace(state)
	}
	return nil
}

func (mc *ModelChecker) GetTraceInfoBetween(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	if mc == nil || to == nil {
		return nil
	}
	if mc.ConcurrentTrace != nil {
		return mc.ConcurrentTrace.GetTraceBetweenStates(from, to)
	}
	if mc.Trace != nil {
		return mc.Trace.GetTraceBetween(from, to)
	}
	return nil
}

func (mc *ModelChecker) Suspend() {
	if mc != nil && mc.StateQueue != nil {
		mc.StateQueue.SuspendAll()
	}
}

func (mc *ModelChecker) Resume() {
	if mc != nil && mc.StateQueue != nil {
		mc.StateQueue.ResumeAll()
	}
}

func (mc *ModelChecker) GetProgress() int64 {
	if mc == nil {
		return -1
	}
	if mc.ConcurrentTrace != nil {
		return int64(mc.ConcurrentTrace.GetLevelForReporting())
	}
	if mc.Trace == nil {
		return -1
	}
	level, err := mc.Trace.GetLevelForReportingWithError()
	if err != nil {
		if isJavaIOException(err) {
			return -1
		}
		panic(err)
	}
	return int64(level)
}

func (mc *ModelChecker) GetStatistics() Value {
	ensureTLCGetSetUniqueStrings()
	if mc == nil {
		return EmptyRecord
	}
	workerID := int32(0)
	if id, ok := CurrentWorkerID(); ok {
		workerID = int32(id)
	}
	names := []*UniqueString{
		tlcGetQueue,
		tlcGetDistinct,
		tlcGetInitial,
		tlcGetGenerated,
		tlcGetDiameter,
		tlcGetDuration,
		tlcGetWorker,
	}
	values := []Value{
		intValueFromInt64(mc.GetStateQueueSize()),
		intValueFromUint64(mc.GetDistinctStatesGenerated()),
		intValueFromInt64(mc.GetInitialStatesGenerated()),
		intValueFromInt64(mc.GetStatesGenerated()),
		intValueFromInt64(mc.GetProgress()),
		intValueFromDurationSince(mc.StartTime),
		NewIntValue(workerID),
	}
	return NewRecordValue(names, values, false)
}

func (mc *ModelChecker) GetConfig() Value {
	if mc == nil {
		return EmptyRecord
	}
	if mc.Config != nil {
		return mc.Config
	}
	return mc.createConfig()
}

func (c *AbstractChecker) createConfig() Value {
	ensureTLCGetSetUniqueStrings()
	if c == nil {
		return EmptyRecord
	}
	names := []*UniqueString{
		tlcGetMode,
		tlcGetDeadlock,
		tlcGetWorker,
		tlcGetSeed,
		tlcGetFingerprint,
		tlcGetInstall,
	}
	values := []Value{
		NewStringValue("bfs"),
		NewBoolValue(c.CheckDeadlock),
		NewIntValue(int32(NumWorkers())),
		NewStringValue(fmt.Sprintf("%d", RandomEnumerableSeed())),
		NewStringValue(fmt.Sprintf("%d", int64(FP64IrredPoly()))),
		NewStringValue(TLCInstallLocation()),
	}
	return NewRecordValue(names, values, false)
}

func (mc *ModelChecker) CheckAssumptions() int {
	if mc.Tool == nil {
		return NoError
	}
	return mc.Tool.CheckAssumptions()
}

func (mc *ModelChecker) ModelCheck() (result int, err error) {
	result = NoError
	DebugPrintMessage("entering modelCheck()")
	debugCleanupExit := ""
	cleanupSuccessOverride := false
	hasCleanupSuccessOverride := false
	defer func() {
		if mc == nil {
			return
		}
		cleanupSuccess := result == NoError
		if hasCleanupSuccessOverride {
			cleanupSuccess = cleanupSuccessOverride
		}
		cleanupErr := mc.Cleanup(cleanupSuccess, mc.CleanupEnabled)
		if err == nil {
			err = cleanupErr
		}
		if cleanupErr == nil && debugCleanupExit != "" {
			DebugPrintMessage(debugCleanupExit)
		}
		if result == NoError && mc.ErrorCode != NoError {
			result = mc.ErrorCode
		}
	}()
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	recovered, err := mc.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	if !recovered {
		if mc.CheckLiveness && mc.LiveCheck != nil && mc.LiveCheck.NumChecker() == 0 {
			PrintError(ECTLCLiveFormulaTautology)
			return ECTLCLiveFormulaTautology, nil
		}
		if result := mc.CheckAssumptions(); result != NoError {
			return result, nil
		}
		DebugPrintMessage("doInit(false)")
		PrintMessage(ECTLCComputingInit)
		result, err = mc.DoInit(false)
		if err != nil {
			DebugPrintMessage("exception in init")
			DebugPrintThrowable(err)
			debugCleanupExit = "exiting, because init failed with exception"
			result = mc.reportInitException(result, err)
			mc.PrintSummary(false)
			return result, err
		}
		if result != NoError {
			DebugPrintMessage("exiting, because init failed")
			mc.checkPostConditionAfterInitFailure()
			return result, nil
		}
		mc.PrintInitGenerated()
	}
	DebugPrintMessage("init processed")
	if len(mc.Tool.GetActions()) == 0 {
		debugCleanupExit = "exiting with actions.length == 0"
		if !mc.StateQueue.IsEmpty() {
			PrintError(ECTLCStatesAndNoNextAction)
			result = ECTLCStatesAndNoNextAction
			cleanupSuccessOverride = true
			hasCleanupSuccessOverride = true
			return result, err
		}
		ReportSuccess(mc.FPSet, mc.GetStatesGenerated())
		mc.PrintSummary(true)
		return NoError, nil
	}
	debugCleanupExit = "exiting modelCheck()"
	DebugPrintMessage("running TLC")
	result, err = mc.RunTLC(javaIntegerMaxValue)
	if err != nil || result != NoError {
		DebugPrintMessage("TLC terminated with error")
		result = resultOrGeneralForError(result, err)
		mc.PrintSummary(false)
		if statsErr := mc.PrintLivenessStatistics(); err == nil {
			err = statsErr
		}
		return result, err
	}
	if mc.ErrState == nil {
		if mc.CheckLiveness && mc.LiveCheck != nil {
			PrintMessage(ECTLCProgressStats,
				fmtInt64(mc.GetProgress()),
				fmtInt64(mc.GetStatesGenerated()),
				fmtUint64(mc.GetDistinctStatesGenerated()),
				fmtInt64(mc.GetStateQueueSize()),
			)
			DebugPrintMessage("checking liveness")
			result, err = mc.LiveCheck.FinalCheck(mc.Tool.NoDebug())
			if err == nil {
				DebugPrintMessage("liveness check complete")
			}
			if err != nil || result != NoError {
				if err == nil {
					DebugPrintMessage("exiting error status on liveness check")
				} else {
					DebugPrintMessage("TLC terminated with error")
				}
				result = resultOrGeneralForError(result, err)
				mc.PrintSummary(false)
				if statsErr := mc.PrintLivenessStatistics(); err == nil {
					err = statsErr
				}
				return result, err
			}
		}
		result = mc.Tool.CheckPostCondition()
		if result == NoError {
			ReportSuccess(mc.FPSet, mc.GetStatesGenerated())
		}
	} else if mc.KeepCallStack {
		result = mc.replayNextErrorCallStack()
	}
	mc.PrintSummary(result == NoError)
	if statsErr := mc.PrintLivenessStatistics(); err == nil {
		err = statsErr
	}
	return result, err
}

func resultOrGeneralForError(result int, err error) int {
	if err != nil && result == NoError {
		return ECGeneral
	}
	return result
}

func (mc *ModelChecker) Checkpoint() error {
	if mc == nil {
		return nil
	}
	PrintMessage(ECTLCCheckpointStart, mc.Metadir)
	if mc.StateQueue != nil {
		if err := mc.StateQueue.BeginChkpt(); err != nil {
			return err
		}
	}
	if err := mc.BeginTraceChkpt(); err != nil {
		return err
	}
	if mc.FPSet != nil {
		if err := mc.FPSet.BeginChkpt(); err != nil {
			return err
		}
	}
	if err := BeginChkptUniqueStrings(mc.Metadir); err != nil {
		return err
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		if err := mc.LiveCheck.BeginChkpt(); err != nil {
			return err
		}
	}
	if mc.StateQueue != nil {
		mc.StateQueue.ResumeAll()
	}
	if mc.StateQueue != nil {
		if err := mc.StateQueue.CommitChkpt(); err != nil {
			return err
		}
	}
	if err := mc.CommitTraceChkpt(); err != nil {
		return err
	}
	if mc.FPSet != nil {
		if err := mc.FPSet.CommitChkpt(); err != nil {
			return err
		}
	}
	if err := CommitChkptUniqueStrings(mc.Metadir); err != nil {
		return err
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		if err := mc.LiveCheck.CommitChkpt(); err != nil {
			return err
		}
	}
	PrintMessage(ECTLCCheckpointEnd)
	return nil
}

func (mc *ModelChecker) Recover() (bool, error) {
	if mc == nil || mc.FromCheckpoint == "" {
		return false, nil
	}
	PrintMessage(ECTLCCheckpointRecoverStart, mc.FromCheckpoint)
	mc.Metadir = mc.FromCheckpoint
	rootName := "Spec"
	if mc.Tool != nil && mc.Tool.GetRootName() != "" {
		rootName = mc.Tool.GetRootName()
	}
	if mc.ConcurrentTrace != nil {
		mc.ConcurrentTrace.SetCheckpointContext(mc.FromCheckpoint, rootName)
	} else if mc.Trace != nil {
		mc.Trace.SetCheckpointContext(mc.FromCheckpoint, rootName)
	}
	if mc.ConcurrentTrace != nil || mc.Trace != nil {
		if err := mc.RecoverTrace(); err != nil {
			return false, err
		}
	}
	if mc.StateQueue != nil {
		setStateQueueDir(mc.StateQueue, mc.FromCheckpoint)
		if err := mc.StateQueue.Recover(); err != nil {
			return false, err
		}
	}
	if mc.FPSet != nil {
		mc.FPSet.Init(NumWorkers(), mc.FromCheckpoint, mc.Tool.GetRootName())
		if err := recoverFPSetFromCheckerTrace(mc.FPSet, mc.Trace, mc.ConcurrentTrace); err != nil {
			return false, err
		}
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		err := mc.Tool.GetInitStates(NewStateFunctor(func(state *TLCStateMut) (any, error) {
			return nil, mc.LiveCheck.AddInitState(mc.Tool, state, state.FingerPrint())
		}))
		if err != nil {
			return false, err
		}
		if err := mc.LiveCheck.Recover(); err != nil {
			return false, err
		}
	}
	fpSize := uint64(0)
	queueSize := int64(0)
	if mc.FPSet != nil {
		fpSize = mc.FPSet.Size()
	}
	if mc.StateQueue != nil {
		queueSize = mc.StateQueue.Size()
	}
	PrintMessage(ECTLCCheckpointRecoverEnd, fmt.Sprint(fpSize), fmt.Sprint(queueSize))
	mc.NumberOfInitialStates = int64(fpSize)
	return true, nil
}

func recoverFPSetFromCheckerTrace(fpSet FPSet, trace *TLCTrace, concurrentTrace *ConcurrentTLCTrace) error {
	if fpSet == nil {
		return nil
	}
	if concurrentTrace == nil {
		return fpSet.RecoverTrace(trace)
	}
	elements, err := concurrentTrace.Elements()
	if err != nil {
		return err
	}
	defer elements.Close()
	for pos := elements.NextPos(); pos != -1; pos = elements.NextPos() {
		fp, err := elements.NextFP()
		if err != nil {
			return err
		}
		if err := fpSet.RecoverFP(fp); err != nil {
			return err
		}
	}
	return nil
}

func (mc *ModelChecker) Cleanup(success bool, cleanup bool) error {
	if mc == nil {
		return nil
	}
	if mc.cleanupDone {
		return nil
	}
	mc.cleanupDone = true
	var err error
	vetoCleanup := modelCheckerVetoCleanup()
	if cleanup && CheckpointExplicitlyEnabled() && mc.StateQueue != nil && !mc.StateQueue.IsEmpty() && (mc.ErrState != nil || mc.TimeBound) {
		// Important Java compatibility: this branch is load-bearing.  Java TLC
		// snapshots an interrupted/error run with queued work and then vetoes
		// metadata deletion so users can recover and continue.  Do not "tidy"
		// the metadir here. This local checkpoint veto still closes resources;
		// only the explicit VETO_CLEANUP property retains the liveness graph.
		if checkpointErr := mc.Checkpoint(); checkpointErr != nil && err == nil {
			err = checkpointErr
		}
		vetoCleanup = true
	}
	if mc.FPSet != nil {
		mc.FPSet.Close()
	}
	if mc.ConcurrentTrace != nil {
		if closeErr := mc.ConcurrentTrace.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	} else if mc.Trace != nil {
		if closeErr := mc.Trace.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		for _, worker := range mc.Workers {
			if worker != nil {
				if closeErr := worker.CloseTrace(); closeErr != nil && err == nil {
					err = closeErr
				}
			}
		}
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		if closeErr := mc.LiveCheck.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	if mc.AllStateWriter != nil {
		if closeErr := mc.AllStateWriter.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	if cleanup && !vetoCleanup {
		// Java's FileUtil.deleteDir(path, success) is deliberately recursive only
		// on success. Keep failure artifacts for inspection/recovery. Resources
		// close above, except graphs retained by the explicit VETO_CLEANUP property.
		deleteDirLikeJava(mc.Metadir, success)
	}
	return err
}

func deleteDirLikeJava(path string, recurse bool) {
	if path == "" {
		return
	}
	if recurse {
		_ = os.RemoveAll(path)
		return
	}
	_ = os.Remove(path)
}

func modelCheckerVetoCleanup() bool {
	if value, ok := tlcLookupSystemProperty(modelCheckerVetoProperty); ok {
		return javaBooleanProperty(value)
	}
	if value, ok := os.LookupEnv("TLAGO_MODEL_CHECKER_VETO_CLEANUP"); ok {
		return strings.EqualFold(value, "true")
	}
	return false
}

func (mc *ModelChecker) BeginTraceChkpt() error {
	if mc == nil {
		return nil
	}
	if mc.ConcurrentTrace != nil {
		return mc.ConcurrentTrace.BeginChkpt()
	}
	for _, worker := range mc.Workers {
		if worker != nil {
			if err := worker.BeginChkpt(); err != nil {
				return err
			}
		}
	}
	if mc.Trace != nil {
		return mc.Trace.BeginChkpt()
	}
	return nil
}

func (mc *ModelChecker) CommitTraceChkpt() error {
	if mc == nil {
		return nil
	}
	if mc.ConcurrentTrace != nil {
		return mc.ConcurrentTrace.CommitChkpt()
	}
	for _, worker := range mc.Workers {
		if worker != nil {
			if err := worker.CommitChkpt(); err != nil {
				return err
			}
		}
	}
	if mc.Trace != nil {
		return mc.Trace.CommitChkpt()
	}
	return nil
}

func (mc *ModelChecker) RecoverTrace() error {
	if mc == nil {
		return nil
	}
	if mc.ConcurrentTrace != nil {
		return mc.ConcurrentTrace.Recover()
	}
	for _, worker := range mc.Workers {
		if worker != nil {
			if err := worker.RecoverTrace(); err != nil {
				return err
			}
		}
	}
	if mc.Trace != nil {
		return mc.Trace.Recover()
	}
	return nil
}

func (mc *ModelChecker) PrintSummary(success bool) {
	if mc == nil {
		return
	}
	if checkerCoverageEnabled(mc.Tool) {
		ReportCoverage(mc.Tool, mc.StartTime)
	}
	if toolMode() {
		mc.PrintProgressStats(mc.StartTime, true)
	}
	PrintMessage(ECTLCStats,
		fmtInt64(mc.GetStatesGenerated()),
		fmtUint64(mc.GetDistinctStatesGenerated()),
		fmtInt64(mc.GetStateQueueSize()),
	)
	depth := int64(0)
	if mc.GetStatesGenerated() != 0 {
		depth = mc.GetProgress()
	}
	PrintMessage(ECTLCSearchDepth, fmtInt64(depth))
	if success {
		mc.PrintOutDegreeSummary()
	}
}

func (mc *ModelChecker) PrintLivenessStatistics() error {
	if mc == nil || !mc.CheckLiveness || mc.LiveCheck == nil || !livenessStatsEnabled() {
		return nil
	}
	runtime.GC()
	inDegree, err := mc.LiveCheck.CalculateInDegreeDiskGraphs(
		NewBucketStatisticsWithMX("Histogram vertex in-degree", javaLivenessPackageName, "DiskGraphsInDegree"),
	)
	if err != nil {
		return err
	}
	printLivenessStatistics(inDegree, mc.LiveCheck.GetOutDegreeStatistics())
	return nil
}

func livenessStatsEnabled() bool {
	value, ok := tlcLookupSystemProperty("tlc2.tool.liveness.statistics")
	return ok && javaBooleanProperty(value)
}

func printLivenessStatistics(inDegree any, outDegree any) {
	fmt.Println(outDegree)
	fmt.Println(inDegree)
	stats, observations := liveWorkerStatsSnapshot()
	fmt.Println(stats)
	plural := ""
	if observations > 1 {
		plural = "s"
	}
	fmt.Println(fmt.Sprintf("%d SCC%s found during liveness checking.", observations, plural))
}

func (mc *ModelChecker) PrintInitGenerated() {
	if mc == nil {
		return
	}
	statesGenerated := mc.GetStatesGenerated()
	plural := ""
	if statesGenerated != 1 {
		plural = "s"
	}
	distinct := mc.GetDistinctStatesGenerated()
	if uint64(statesGenerated) == distinct {
		PrintMessage(ECTLCInitGenerated1, fmtInt64(statesGenerated), plural)
		return
	}
	PrintMessage(ECTLCInitGenerated2, fmtInt64(statesGenerated), plural, fmtUint64(distinct))
}

func (mc *ModelChecker) PrintProgressStats(startTime time.Time, isFinal bool) {
	_ = isFinal
	if mc == nil {
		return
	}
	fpSetSize := uint64(0)
	if mc.FPSet != nil {
		fpSetSize = mc.FPSet.Size()
	}
	factor := ProgressInterval().Minutes()
	if !startTime.IsZero() {
		mc.OldNumOfGenStates = 0
		mc.OldFPSetSize = 0
		factor = time.Since(startTime).Minutes()
	}
	generated := mc.GetStatesGenerated()
	mc.StatesPerMinute = javaDoubleToLong(float64(generated-mc.OldNumOfGenStates) / factor)
	mc.OldNumOfGenStates = generated
	distinctDelta := int64(fpSetSize) - int64(mc.OldFPSetSize)
	mc.DistinctStatesPerMinute = javaDoubleToLong(float64(distinctDelta) / factor)
	mc.OldFPSetSize = fpSetSize
	PrintMessage(ECTLCProgressStats,
		fmtInt64(mc.GetProgress()),
		fmtInt64(generated),
		fmtUint64(fpSetSize),
		fmtInt64(mc.GetStateQueueSize()),
		fmtInt64(mc.StatesPerMinute),
		fmtInt64(mc.DistinctStatesPerMinute),
	)
}

func javaDoubleToLong(value float64) int64 {
	if math.IsNaN(value) {
		return 0
	}
	if value <= float64(math.MinInt64) {
		return math.MinInt64
	}
	if value >= float64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(value)
}

func (mc *ModelChecker) PrintOutDegreeSummary() {
	if mc == nil {
		return
	}
	agg := NewBucketStatistics("State Graph OutDegree")
	for _, worker := range mc.Workers {
		if worker == nil || worker.OutDegree == nil {
			continue
		}
		for _, sample := range worker.OutDegree.Samples() {
			agg.AddSampleCount(sample.Amount, sample.Count)
		}
	}
	if agg.Observations() == 0 {
		return
	}
	PrintMessage(ECTLCStateGraphOutdegree,
		fmtInt(agg.Min()),
		fmtInt64(int64(math.Round(agg.Mean()))),
		fmtInt64(int64(math.Round(agg.Percentile(.95)))),
		fmtInt(agg.Max()),
	)
}

func toolMode() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.Tool
}

func (mc *ModelChecker) RunTLC(maxDepth int) (int, error) {
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	if mc.StateQueue == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no state queue")
	}
	if maxDepth > 0 && maxDepth < 2 {
		return NoError, nil
	}
	mc.initWorkers()
	if mc.FPSet != nil {
		mc.FPSet.IncWorkers(len(mc.Workers))
	}
	for _, worker := range mc.Workers {
		if worker == nil {
			continue
		}
		worker.Start()
	}
	result, joinErr := mc.waitForWorkersWithPeriodicWork(maxDepth)
	if result != NoError {
		return result, joinErr
	}
	if joinErr != nil {
		if mc.ErrorCode != NoError {
			if mc.KeepCallStack {
				return NoError, nil
			}
			return mc.ErrorCode, nil
		}
		return ECGeneral, joinErr
	}
	if !mc.KeepCallStack && mc.ErrorCode != NoError {
		return mc.ErrorCode, nil
	}
	return NoError, nil
}

func (mc *ModelChecker) waitForWorkersWithPeriodicWork(maxDepth int) (int, error) {
	done := make(chan error, 1)
	go func() {
		done <- mc.joinWorkers()
	}()
	select {
	case err := <-done:
		result, periodicErr := mc.DoPeriodicWork()
		if periodicErr != nil || result != NoError {
			return result, periodicErr
		}
		return NoError, err
	case <-time.After(3 * time.Second):
	}
	interval := ProgressInterval()
	coverageCountdown := periodicCoverageCountdown(interval)
	for {
		result, err := mc.DoPeriodicWork()
		if err != nil || result != NoError {
			if mc.StateQueue != nil {
				mc.StateQueue.FinishAll()
			}
			joinErr := <-done
			if err != nil {
				return result, err
			}
			return result, joinErr
		}
		if mc.isModelCheckerDone() {
			return NoError, <-done
		}
		mc.runTLCContinueDoing(coverageCountdown, maxDepth)
		if coverageCountdown == 0 {
			coverageCountdown = periodicCoverageCountdown(interval)
		} else if coverageCountdown > 0 {
			coverageCountdown--
		}
		if mc.isModelCheckerDone() {
			return NoError, <-done
		}
		if interval <= 0 {
			select {
			case err := <-done:
				return NoError, err
			default:
				continue
			}
		}
		timer := time.NewTimer(interval)
		select {
		case err := <-done:
			timer.Stop()
			return NoError, err
		case <-timer.C:
		}
	}
}

func (mc *ModelChecker) isModelCheckerDone() bool {
	if mc == nil || mc.AbstractChecker == nil {
		return true
	}
	mc.mu.Lock()
	done := mc.Done
	mc.mu.Unlock()
	return done
}

func (mc *ModelChecker) runTLCContinueDoing(count int, depth int) {
	mc.PrintProgressStats(time.Time{}, false)
	if depth > 0 && mc.GetProgress() > int64(depth) {
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
		mc.SetDone()
		return
	}
	if count == 0 {
		mc.reportPeriodicCoverage()
	}
}

func periodicCoverageCountdown(interval time.Duration) int {
	Globals.Lock()
	coverageMillis := Globals.CoverageInterval
	Globals.Unlock()
	if coverageMillis < 0 {
		return -1
	}
	progressMillis := int(interval / time.Millisecond)
	if progressMillis <= 0 {
		return 0
	}
	return coverageMillis / progressMillis
}

func (mc *ModelChecker) reportPeriodicCoverage() {
	if mc == nil || !checkerCoverageEnabled(mc.Tool) {
		return
	}
	ReportCoverage(mc.Tool, mc.StartTime)
}

func checkerCoverageEnabled(tool *Tool) bool {
	return tool != nil && CoverageEnabled() && len(tool.GetActions()) > 0
}

func (mc *ModelChecker) joinWorkers() error {
	var joinErr error
	for _, worker := range mc.Workers {
		if worker == nil {
			continue
		}
		if err := worker.Join(); err != nil && joinErr == nil {
			joinErr = err
		}
	}
	return joinErr
}

func (mc *ModelChecker) DoPeriodicWork() (int, error) {
	if mc == nil {
		return NoError, nil
	}
	createCheckpoint := DoCheckPoint()
	var periodic SemanticNode
	if mc.Tool != nil {
		periodic = mc.Tool.GetPeriodic()
	}
	forceLiveCheck := mc.ForceLiveCheck
	liveCheckDue := mc.CheckLiveness && mc.LiveCheck != nil && mc.RuntimeRatio <= LivenessRatio() && mc.LiveCheck.DoLiveCheck()
	if !liveCheckDue && !forceLiveCheck && !createCheckpoint && periodic == nil {
		mc.UpdateRuntimeRatio(0)
		return NoError, nil
	}
	if mc.StateQueue == nil || !mc.StateQueue.SuspendAll() {
		return NoError, nil
	}
	resume := true
	defer func() {
		if resume && mc.StateQueue != nil {
			mc.StateQueue.ResumeAll()
		}
	}()
	if mc.CheckLiveness && mc.LiveCheck != nil && (mc.RuntimeRatio < LivenessRatio() || forceLiveCheck) {
		start := time.Now()
		result, err := mc.LiveCheck.Check(mc.Tool.NoDebug(), forceLiveCheck)
		if err != nil || result != NoError {
			resume = false
			return result, err
		}
		mc.ForceLiveCheck = false
		mc.UpdateRuntimeRatio(time.Since(start))
	} else if mc.RuntimeRatio > LivenessRatio() {
		mc.UpdateRuntimeRatio(0)
	}
	if periodic != nil {
		value, err := mc.Tool.NoDebug().Eval(periodic)
		if err != nil {
			resume = false
			PrintError(ECGeneral, generalErrorParams("", err)...)
			return ECGeneral, err
		}
		if boolValue, ok := value.(*BoolValue); ok && !boolValue.Val {
			resume = false
			return ECTLCAssumptionFalse, nil
		}
	}
	if createCheckpoint {
		resume = false
		if err := mc.Checkpoint(); err != nil {
			return ECSystemCheckpointRecoveryCorrupt, err
		}
	}
	return NoError, nil
}

func (mc *ModelChecker) UpdateRuntimeRatio(delta time.Duration) {
	if mc == nil {
		return
	}
	if delta < 0 {
		delta = 0
	}
	totalRuntime := time.Since(mc.StartTime) - ProgressInterval() - delta
	absLivenessRuntime := float64(totalRuntime) * mc.RuntimeRatio
	if absLivenessRuntime < 0 {
		absLivenessRuntime = 0
	}
	denominator := float64(totalRuntime + ProgressInterval() + delta)
	mc.RuntimeRatio = (float64(delta) + absLivenessRuntime) / denominator
}

func (mc *ModelChecker) ForceLivenessCheck() {
	if mc == nil {
		return
	}
	mc.ForceLiveCheck = true
}

func (mc *ModelChecker) DoInit(ignoreCancel bool) (int, error) {
	return mc.doInitWithTool(mc.Tool, ignoreCancel)
}

func (mc *ModelChecker) doInitWithTool(tool *Tool, ignoreCancel bool) (result int, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
			result = initExceptionCode(err)
		}
	}()
	if tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	functor := &doInitFunctor{mc: mc, tool: tool, forceChecks: ignoreCancel, returnValue: NoError}
	err = tool.GetInitStates(NewStateFunctor(functor.AddElement))
	if isDoInitInvariantViolatedException(err) {
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

func (mc *ModelChecker) reportInitException(result int, err error) int {
	if err == nil {
		if result != NoError {
			return result
		}
		return ECGeneral
	}
	// Java's catch(Throwable) takes the original exception's code before
	// diagnostic replay, even when doInit returned a generic failure code.
	result = initExceptionCode(err)
	message := javaThrowableMessage(err)
	if javaSystemFailureCode(err) == ECSystemStackOverflow {
		message = formatMessage(ECSystemStackOverflow, nil)
	}
	if mc != nil && mc.ErrState != nil {
		PrintError(ECTLCInitialState, message, mc.ErrState.String())
	} else {
		PrintError(ECGeneral, message)
	}
	if replayResult := mc.replayInitErrorCallStack(result); replayResult != NoError {
		result = replayResult
	}
	return result
}

func initExceptionCode(err error) int {
	if failure := javaRuntimeException(err); failure != nil {
		return failure.Code
	}
	return ECGeneral
}

func (mc *ModelChecker) replayInitErrorCallStack(fallback int) int {
	if mc == nil || mc.Tool == nil {
		if fallback != NoError {
			return fallback
		}
		return ECGeneral
	}
	callStackTool := NewCallStackTool(mc.Tool)
	mc.NumberOfInitialStates = 0
	if _, err := mc.doInitWithTool(callStackTool, true); err != nil {
		if fpErr, ok := err.(*FingerprintException); ok && fpErr != nil {
			trace := fpErr.GetTrace()
			if callStackTool.HasCallStack() {
				trace = callStackTool.CallStackString()
			}
			PrintErrorNullable(ECTLCFingerprintException, javaString(trace), javaThrowableDetailMessage(fpErr.GetRootCause()))
			return ECTLCFingerprintException
		}
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
		return ECTLCNestedExpression
	}
	return fallback
}

func (mc *ModelChecker) DoNext(curState *TLCStateMut) (bool, error) {
	return mc.doNextWithTool(mc.Tool, curState, nil, nil)
}

func (mc *ModelChecker) replayNextErrorCallStack() int {
	if mc == nil || mc.Tool == nil || mc.PredErrState == nil {
		return mcErrorCodeOrGeneral(mc)
	}
	callStackTool := NewCallStackTool(mc.Tool)
	var liveNextStates *SetOfStates
	if mc.CheckLiveness {
		liveNextStates = NewSetOfStates()
	}
	replayWorker := NewWorker(4223)
	replayWorker.Checker = mc
	replayWorker.Tool = mc.Tool
	replayWorker.captureRunSettings()
	replayWorker.DisableTraceMirror = true
	rootName := "Spec"
	if mc.Tool.GetRootName() != "" {
		rootName = mc.Tool.GetRootName()
	}
	replayWorker.SetTraceContext(mc.Metadir, rootName)
	defer replayWorker.CloseTrace()
	err := replayWorker.ensureTraceRAF()
	if err == nil {
		_, err = mc.doNextWithTool(callStackTool, mc.PredErrState, liveNextStates, replayWorker)
	}
	if err != nil {
		if fpErr, ok := err.(*FingerprintException); ok && fpErr != nil {
			trace := fpErr.GetTrace()
			if callStackTool.HasCallStack() {
				trace = callStackTool.CallStackString()
			}
			PrintErrorNullable(ECTLCFingerprintException, javaString(trace), javaThrowableDetailMessage(fpErr.GetRootCause()))
			return ECTLCFingerprintException
		}
		if isValueEvalException(err) && javaSystemFailureCode(err) == NoError {
			PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
			if eval, ok := err.(*EvalException); ok {
				return eval.GetErrorCode()
			}
			return err.(*TLCError).Code
		}
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
		return ECTLCNestedExpression
	}
	return NoError
}

func mcErrorCodeOrGeneral(mc *ModelChecker) int {
	if mc != nil && mc.ErrorCode != NoError {
		return mc.ErrorCode
	}
	return ECGeneral
}

func (mc *ModelChecker) doNextWithTool(tool *Tool, curState *TLCStateMut, liveNextStates *SetOfStates, worker *Worker) (stop bool, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			stop, err = true, panicValueAsError(failure)
		}
	}()
	if mc == nil {
		return true, newTLCError(ECGeneral, "model checker is nil")
	}
	if tool == nil {
		return true, newTLCError(ECGeneral, "model checker has no tool")
	}
	restoreRandomState := PushRandomEnumerableState(curState)
	defer restoreRandomState()
	restoreCurrentState := PushCurrentState(curState)
	defer restoreCurrentState()
	deadLocked := true
	var succState *TLCStateMut
	for _, action := range tool.GetActions() {
		nextStates, err := tool.GetNextStates(action, curState)
		if err != nil {
			mc.doNextFailed(curState, succState, err)
			return true, err
		}
		size := 0
		if nextStates != nil {
			size = nextStates.Size()
		}
		if worker != nil {
			worker.IncrementStatesGenerated(int64(size))
		} else {
			mc.NextStatesGenerated += int64(size)
		}
		deadLocked = deadLocked && size == 0
		for i := 0; i < size; i++ {
			succState = nextStates.At(i)
			if !tool.IsGoodState(succState) {
				return mc.doNextSetErrParams(curState, succState, false, ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(tool, action, succState)...), nil
			}
			inModel, err := tool.IsInModel(succState)
			if err != nil {
				mc.doNextFailed(curState, succState, err)
				return true, err
			}
			if inModel {
				inActions, err := tool.IsInActions(curState, succState)
				if err != nil {
					mc.doNextFailed(curState, succState, err)
					return true, err
				}
				inModel = inActions
			}
			unseen := true
			if inModel {
				seen, err := mc.isSeenStateUsingWorker(workerIDForReplayWorker(worker), worker, curState, succState, action, liveNextStates)
				if err != nil {
					mc.doNextFailed(curState, succState, err)
					return true, err
				}
				unseen = !seen
			}
			if unseen {
				stop, err := mc.doNextCheckInvariantsWithTool(tool, curState, succState, false)
				if stop || err != nil {
					return stop, err
				}
			}
			stop, err := mc.doNextCheckImpliedWithTool(tool, curState, succState, false)
			if stop || err != nil {
				return stop, err
			}
			if inModel && unseen && mc.StateQueue != nil {
				mc.StateQueue.SEnqueue(succState)
			}
		}
		succState = nil
	}
	if deadLocked && mc.CheckDeadlock {
		return mc.doNextSetErr(curState, nil, false, ECTLCDeadlockReached, ""), nil
	}
	return false, nil
}

func workerIDForReplayWorker(worker *Worker) int {
	if worker == nil {
		return 0
	}
	return worker.ID
}

func (mc *ModelChecker) processSuccessorForWorker(worker *Worker, curState *TLCStateMut, succState *TLCStateMut, action *Action, collectedStates *SetOfStates) (bool, bool, error) {
	tool := worker.Tool
	if tool == nil {
		panic(NewNullPointerException())
	}
	if !tool.IsGoodState(succState) {
		return mc.doNextSetErrParamsWithPostConditionTool(tool, curState, succState, false, ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(mc.Tool, action, succState)...), false, nil
	}
	if succState != nil {
		succState.SetPredecessor(curState).SetAction(action)
	}
	inModel, err := tool.IsInModel(succState)
	if err != nil {
		return true, false, err
	}
	if inModel {
		inActions, err := tool.IsInActions(curState, succState)
		if err != nil {
			return true, false, err
		}
		inModel = inActions
	}
	unseen := true
	if inModel {
		seen, err := mc.isSeenStateUsingWorker(worker.ID, worker, curState, succState, action, collectedStates)
		if err != nil {
			return true, false, err
		}
		unseen = !seen
	} else if worker.stateWriter == nil {
		panic(NewNullPointerException())
	} else if worker.stateWriter.IsConstrained() {
		if err := mc.writeConstrainedTransitionReasons(tool, worker.stateWriter, curState, succState, action); err != nil {
			return true, false, err
		}
	}
	if unseen {
		stop, err := mc.doNextCheckInvariantsWithTool(tool, curState, succState, true)
		if stop || err != nil {
			return stop, false, err
		}
	}
	stop, err := mc.doNextCheckImpliedWithTool(tool, curState, succState, true)
	if stop || err != nil {
		return stop, false, err
	}
	if inModel && unseen {
		if worker.stateQueue == nil {
			panic(NewNullPointerException())
		}
		worker.stateQueue.SEnqueue(succState)
		CountStateVariableCoverage(succState)
		return false, true, nil
	}
	return false, false, nil
}

func (mc *ModelChecker) writeConstrainedTransitionReasons(tool *Tool, writer IStateWriter, curState *TLCStateMut, succState *TLCStateMut, action *Action) error {
	for _, constraint := range tool.requireConstraintArray(tool.GetModelConstraints()) {
		ok, err := tool.IsInModelForConstraint(constraint, succState)
		if err != nil {
			return err
		}
		if !ok {
			if err := writer.WriteTransition(curState, succState, StateVisitNotInModel, action, constraint); err != nil {
				return err
			}
		}
	}
	for _, constraint := range tool.requireConstraintArray(tool.GetActionConstraints()) {
		ok, err := tool.IsInActionsForConstraint(constraint, curState, succState)
		if err != nil {
			return err
		}
		if !ok {
			if err := writer.WriteTransition(curState, succState, StateVisitNotInModel, action, constraint); err != nil {
				return err
			}
		}
	}
	return nil
}

func (mc *ModelChecker) GetStatesGenerated() int64 {
	total := mc.NumberOfInitialStates + mc.NextStatesGenerated
	for _, worker := range mc.Workers {
		if worker != nil {
			total += worker.GetStatesGenerated()
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

func (mc *ModelChecker) isSeenStateUsingWorker(workerID int, worker *Worker, curState *TLCStateMut, succState *TLCStateMut, action *Action, collectedStates *SetOfStates) (bool, error) {
	tool := mc.Tool
	if worker != nil {
		tool = worker.Tool
		if tool == nil {
			panic(NewNullPointerException())
		}
	}
	fp := succState.FingerPrintWithTool(tool)
	fingerprintSet, writer := mc.FPSet, mc.AllStateWriter
	if worker != nil {
		fingerprintSet, writer = worker.fingerprintSet, worker.stateWriter
	}
	if fingerprintSet == nil {
		panic(NewNullPointerException())
	}
	seen := fingerprintSet.Put(fp)
	if worker != nil && writer == nil {
		panic(NewNullPointerException())
	}
	if writer != nil {
		status := StateVisitUnseen
		if seen {
			status = StateVisitSeen
		}
		if err := writer.WriteTransition(curState, succState, status, action); err != nil {
			return seen, err
		}
	}
	if !seen {
		if action != nil && CoverageActionEnabled() {
			action.CM.IncSecondary()
		}
		if worker != nil {
			if err := worker.WriteNextState(curState, fp, succState, action); err != nil {
				return seen, err
			}
		} else if mc.Trace != nil {
			if err := mc.Trace.WriteNextStateForWorker(workerID, curState, fp, succState, action); err != nil {
				return seen, err
			}
		}
	}
	if collectedStates != nil {
		if worker != nil {
			collectedStates.PutFP(fp, succState, tool)
		} else {
			collectedStates.PutFP(fp, succState)
		}
	}
	return seen, nil
}

func (mc *ModelChecker) doNextCheckInvariants(curState *TLCStateMut, succState *TLCStateMut) (bool, error) {
	return mc.doNextCheckInvariantsWithTool(mc.Tool, curState, succState, true)
}

func (mc *ModelChecker) doNextCheckInvariantsWithTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, withPostCondition bool) (bool, error) {
	invariants := tool.GetInvariants()
	names := tool.GetInvNames()
	for i, invariant := range invariants {
		valid, err := tool.IsValidState(invariant, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCInvariantEvaluationFailed, nameAt(names, i), err)
		}
		if !valid {
			if continuationEnabled() {
				mc.printContinuationViolation(curState, succState, ECTLCInvariantViolatedBehavior, nameAt(names, i))
				return false, nil
			}
			if withPostCondition {
				return mc.doNextSetErrWithPostConditionTool(tool, curState, succState, false, ECTLCInvariantViolatedBehavior, nameAt(names, i)), nil
			}
			return mc.doNextSetErr(curState, succState, false, ECTLCInvariantViolatedBehavior, nameAt(names, i)), nil
		}
	}
	return false, nil
}

func (mc *ModelChecker) doNextCheckImplied(curState *TLCStateMut, succState *TLCStateMut) (bool, error) {
	return mc.doNextCheckImpliedWithTool(mc.Tool, curState, succState, true)
}

func (mc *ModelChecker) doNextCheckImpliedWithTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, withPostCondition bool) (bool, error) {
	implied := tool.GetImpliedActions()
	names := tool.GetImpliedActNames()
	for i, action := range implied {
		valid, err := tool.IsValidTransition(action, curState, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCActionPropertyEvaluationFailed, nameAt(names, i), err)
		}
		if !valid {
			if continuationEnabled() {
				mc.printContinuationViolation(curState, succState, ECTLCActionPropertyViolatedBehavior, nameAt(names, i))
				return false, nil
			}
			if withPostCondition {
				return mc.doNextSetErrWithPostConditionTool(tool, curState, succState, false, ECTLCActionPropertyViolatedBehavior, nameAt(names, i)), nil
			}
			return mc.doNextSetErr(curState, succState, false, ECTLCActionPropertyViolatedBehavior, nameAt(names, i)), nil
		}
	}
	return false, nil
}

func (mc *ModelChecker) doNextSetErr(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, param string) bool {
	if param == "" {
		return mc.doNextSetErrParams(curState, succState, keep, ec)
	}
	return mc.doNextSetErrParams(curState, succState, keep, ec, param)
}

func (mc *ModelChecker) doNextSetErrParams(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, params ...string) bool {
	mc.nextErrorMu.Lock()
	defer mc.nextErrorMu.Unlock()
	return mc.doNextSetErrParamsLocked(curState, succState, keep, ec, params...)
}

func (mc *ModelChecker) doNextSetErrParamsLocked(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, params ...string) bool {
	if mc.SetErrState(curState, succState, keep, ec) {
		if len(params) == 0 {
			PrintError(ec)
		} else {
			PrintError(ec, params...)
		}
		mc.printBehaviorTrace(curState, succState)
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
	return true
}

func (mc *ModelChecker) doNextSetErrWithPostCondition(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, param string) bool {
	return mc.doNextSetErrWithPostConditionTool(mc.Tool, curState, succState, keep, ec, param)
}

func (mc *ModelChecker) doNextSetErrWithPostConditionTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, param string) bool {
	if param == "" {
		return mc.doNextSetErrParamsWithPostConditionTool(tool, curState, succState, keep, ec)
	}
	return mc.doNextSetErrParamsWithPostConditionTool(tool, curState, succState, keep, ec, param)
}

func (mc *ModelChecker) doNextSetErrParamsWithPostCondition(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, params ...string) bool {
	return mc.doNextSetErrParamsWithPostConditionTool(mc.Tool, curState, succState, keep, ec, params...)
}

func (mc *ModelChecker) doNextSetErrParamsWithPostConditionTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, params ...string) bool {
	mc.nextErrorMu.Lock()
	defer mc.nextErrorMu.Unlock()
	isConsole := !mc.isModelCheckerDone()
	result := mc.doNextSetErrParamsLocked(curState, succState, keep, ec, params...)
	mc.checkPostConditionWithErrorTraceTool(tool, curState, succState, isConsole)
	return result
}

func (mc *ModelChecker) printContinuationViolation(curState *TLCStateMut, succState *TLCStateMut, ec int, param string) {
	mc.nextErrorMu.Lock()
	defer mc.nextErrorMu.Unlock()
	PrintError(ec, param)
	mc.printBehaviorTrace(curState, succState)
}

func (mc *ModelChecker) checkPostConditionAfterInitFailure() {
	if mc == nil || mc.Tool == nil {
		return
	}
	if mc.ErrState != nil {
		mc.Tool.CheckPostConditionWithCounterExample(NewCounterExampleFromInitialState(mc.ErrState))
		return
	}
	mc.Tool.CheckPostCondition()
}

func (mc *ModelChecker) checkPostConditionWithErrorTrace(curState *TLCStateMut, succState *TLCStateMut, isConsole bool) {
	mc.checkPostConditionWithErrorTraceTool(mc.Tool, curState, succState, isConsole)
}

func (mc *ModelChecker) checkPostConditionWithErrorTraceTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, isConsole bool) {
	trace := mc.errorTraceInfo(tool, curState, succState)
	trace = aliasTraceWithToolPairs(tool, trace)
	tool.CheckPostConditionWithCounterExample(NewCounterExample(trace, UnknownAction, 0, isConsole))
}

func (mc *ModelChecker) printBehaviorTrace(curState *TLCStateMut, succState *TLCStateMut) {
	if mc.ConcurrentTrace != nil {
		mc.ConcurrentTrace.PrintTrace(curState, succState)
	} else if mc.Trace != nil {
		mc.Trace.PrintTrace(curState, succState)
	}
}

func (mc *ModelChecker) errorTraceInfo(tool *Tool, curState *TLCStateMut, succState *TLCStateMut) []*TLCStateInfo {
	if curState == nil {
		panic(NewNullPointerException())
	}
	if succState == nil {
		if curState.IsInitial() {
			return []*TLCStateInfo{mc.stateInfoForState(tool, curState, nil)}
		}
		trace := mc.traceInfoPrefix(curState)
		return append(trace, mc.stateInfoForState(tool, curState, lastTraceState(trace)))
	}
	if succState.AllAssigned() && succState.WorkerID == TLCStateInitWorkerID {
		if curState.IsInitial() {
			return []*TLCStateInfo{
				mc.stateInfoForState(tool, curState, nil),
				mc.stateInfoForTransition(tool, succState, curState),
			}
		}
		trace := mc.traceInfoPrefix(curState)
		trace = append(trace, mc.stateInfoForState(tool, curState, lastTraceState(trace)))
		return append(trace, mc.stateInfoForTransition(tool, succState, curState))
	}
	trace := mc.traceInfoPrefix(succState)
	return append(trace, mc.stateInfoForTransition(tool, succState, curState))
}

func lastTraceState(trace []*TLCStateInfo) *TLCStateMut {
	if len(trace) == 0 {
		panic(NewNoSuchElementException())
	}
	if trace[len(trace)-1] == nil {
		panic(NewNullPointerException())
	}
	return trace[len(trace)-1].State
}

func (mc *ModelChecker) stateInfoForState(tool *Tool, state *TLCStateMut, predecessor *TLCStateMut) *TLCStateInfo {
	if state == nil {
		return nil
	}
	if predecessor != nil {
		return mc.stateInfoForTransition(tool, state, predecessor)
	}
	if tool == nil {
		panic(NewNullPointerException())
	}
	info, err := tool.GetState(state.FingerPrint())
	if err != nil {
		panic(err)
	}
	return info
}

func (mc *ModelChecker) stateInfoForTransition(tool *Tool, state *TLCStateMut, predecessor *TLCStateMut) *TLCStateInfo {
	if state == nil {
		return nil
	}
	if tool == nil {
		panic(NewNullPointerException())
	}
	info, err := tool.GetStateForTransition(state, predecessor)
	if err != nil {
		panic(err)
	}
	return info
}

func appendTraceStateIfMissing(trace []*TLCStateInfo, state *TLCStateMut) []*TLCStateInfo {
	if state == nil {
		return trace
	}
	if len(trace) > 0 {
		last := trace[len(trace)-1]
		if last != nil && last.State != nil && (last.State == state || last.State.Equal(state)) {
			return trace
		}
	}
	info := NewTLCStateInfo(state)
	fp := state.FingerPrint()
	info.FP = &fp
	return append(trace, info)
}

func (mc *ModelChecker) traceInfoPrefix(state *TLCStateMut) []*TLCStateInfo {
	if mc == nil {
		panic(NewNullPointerException())
	}
	if mc.ConcurrentTrace != nil {
		return mc.ConcurrentTrace.GetTraceFromState(state)
	}
	if mc.Trace != nil {
		return mc.Trace.GetTrace(state)
	}
	panic(NewNullPointerException())
}

func aliasTraceWithTool(tool *Tool, trace []*TLCStateInfo) []*TLCStateInfo {
	if tool == nil || len(trace) == 0 {
		return trace
	}
	aliased := append([]*TLCStateInfo(nil), trace...)
	for i, current := range aliased {
		successor := current.OriginalState()
		if i+1 < len(aliased) {
			successor = aliased[i+1].OriginalState()
		}
		alias, err := tool.EvalAliasInfo(current, successor, func() []*TLCStateInfo {
			return append([]*TLCStateInfo(nil), aliased[:i]...)
		})
		if err == nil && alias != nil {
			aliased[i] = alias
		}
	}
	return aliased
}

func aliasTraceWithToolPairs(tool *Tool, trace []*TLCStateInfo) []*TLCStateInfo {
	if tool == nil || len(trace) == 0 {
		return trace
	}
	aliased := append([]*TLCStateInfo(nil), trace...)
	for i, current := range aliased {
		next := current
		if i+1 < len(aliased) {
			next = aliased[i+1]
		}
		if next == nil {
			panic(NewNullPointerException())
		}
		successor := next.OriginalState()
		alias, err := tool.EvalAliasInfoPair(current, successor)
		if err != nil {
			panic(err)
		}
		aliased[i] = alias
	}
	return aliased
}

func (mc *ModelChecker) doNextEvalFailed(curState *TLCStateMut, succState *TLCStateMut, ec int, param string, err error) error {
	mc.nextErrorMu.Lock()
	defer mc.nextErrorMu.Unlock()
	if mc.SetErrState(curState, succState, true, ec) {
		msg := javaThrowableMessage(err)
		if param == "" {
			PrintError(ec, msg)
		} else {
			PrintError(ec, param, msg)
		}
		mc.printBehaviorTrace(curState, succState)
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
	return err
}

func (mc *ModelChecker) doNextFailed(curState *TLCStateMut, succState *TLCStateMut, err error) {
	if err == nil {
		return
	}
	mc.nextErrorMu.Lock()
	defer mc.nextErrorMu.Unlock()
	ec, params, keepCallStack := doNextFailureMessage(err)
	if mc.SetErrState(curState, succState, keepCallStack, ec) {
		if params != nil {
			if eval, ok := err.(*EvalException); ok && eval.NullableParameters != nil {
				PrintErrorNullable(ec, eval.NullableParameters...)
			} else if failure, ok := err.(*TLCError); ok && !failure.Runtime && failure.NullableParams != nil {
				PrintErrorNullable(ec, failure.NullableParams...)
			} else {
				PrintError(ec, params...)
			}
		}
		mc.printBehaviorTrace(curState, succState)
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
}

func doNextFailureMessage(err error) (int, []string, bool) {
	// Java tests the thrown exception itself, not its nested causes.
	if javaThrowableDetailMessage(err) == nil && javaSystemFailureCode(err) == NoError {
		return ECGeneral, nil, true
	}
	ec := ECGeneral
	params := []string{err.Error()}
	keepCallStack := true

	if eval, ok := err.(*EvalException); ok && eval != nil {
		ec = eval.GetErrorCode()
		if eval.HasParameters() {
			params = eval.GetParameters()
		} else if ec == ECGeneral {
			params = generalErrorParams("", err)
		} else {
			params = []string{}
		}
		return ec, params, keepCallStack
	}

	if tlcErr, ok := err.(*TLCError); ok && tlcErr != nil {
		if tlcErr.Runtime {
			return ECGeneral, generalErrorParams("", err), true
		}
		keepCallStack = true
		if code := javaSystemFailureCode(err); code == ECSystemStackOverflow || code == ECSystemOutOfMemory || code == ECTLCBug {
			keepCallStack = false
		}
		if tlcErr.Code != ECGeneral {
			if tlcErr.Params == nil {
				return tlcErr.Code, []string{}, keepCallStack
			}
			return tlcErr.Code, tlcErr.Params, keepCallStack
		}
	}

	return ECGeneral, generalErrorParams("", err), keepCallStack
}

func generalErrorParams(cause string, err error) []string {
	if err == nil {
		return nil
	}
	return []string{javaGeneralErrorMessage(cause, err)}
}

func javaThrowableMessage(err error) string {
	if message := javaThrowableDetailMessage(err); message != nil {
		return *message
	}
	return javaThrowableString(err)
}

type doInitFunctor struct {
	mc          *ModelChecker
	tool        *Tool
	forceChecks bool
	errState    *TLCStateMut
	err         error
	returnValue int
}

func (f *doInitFunctor) AddElement(curState *TLCStateMut) (result any, err error) {
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
	// DoInitFunctor catches throwables from fingerprinting and state callbacks
	// as well as failures returned by evaluator methods.
	defer func() {
		if failure := recover(); failure != nil {
			result, err = f.handleInitError(curState, panicValueAsError(failure))
		}
	}()
	if !f.tool.IsGoodState(curState) {
		PrintError(ECTLCInitialState, "current state is not a legal state", curState.String())
		f.errState = curState
		f.returnValue = ECTLCInitialState
		return f.returnValue, NewDoInitInvariantViolatedException()
	}
	inModel, err := f.tool.IsInModel(curState)
	if err != nil {
		return f.handleInitError(curState, err)
	}
	seen := false
	if inModel {
		fp := curState.FingerPrint()
		seen = f.mc.FPSet.Put(fp)
		if !seen {
			if err := f.mc.AllStateWriter.WriteInitState(curState); err != nil {
				return f.handleInitError(curState, err)
			}
			if worker := f.mc.workerAt(0); worker != nil {
				if err := worker.WriteInitState(curState, fp); err != nil {
					return f.handleInitError(curState, err)
				}
			} else if f.mc.Trace != nil {
				if err := f.mc.Trace.WriteInitState(curState, fp); err != nil {
					return f.handleInitError(curState, err)
				}
			}
			f.mc.StateQueue.Enqueue(curState)
			if f.mc.CheckLiveness && f.mc.LiveCheck != nil {
				if err := f.mc.LiveCheck.AddInitState(f.tool.NoDebug(), curState, fp); err != nil {
					return f.handleInitError(curState, err)
				}
			}
		}
	}
	if !seen || f.forceChecks {
		for i, invariant := range f.tool.GetInvariants() {
			valid, err := f.tool.IsValidState(invariant, curState)
			if err != nil {
				return f.handleInitError(curState, err)
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
					return f.returnValue, NewDoInitInvariantViolatedException()
				}
			}
		}
		for i, implied := range f.tool.GetImpliedInits() {
			valid, err := f.tool.IsValidState(implied, curState)
			if err != nil {
				return f.handleInitError(curState, err)
			}
			if !valid {
				alias := curState
				if f.tool != nil {
					alias = f.tool.EvalAlias(curState, curState)
				}
				PrintError(ECTLCPropertyViolatedInitial, nameAt(f.tool.GetImpliedInitNames(), i), alias.String())
				f.errState = curState
				f.returnValue = ECTLCPropertyViolatedInitial
				return f.returnValue, NewDoInitInvariantViolatedException()
			}
		}
	}
	return f.returnValue, nil
}

func (f *doInitFunctor) handleInitError(curState *TLCStateMut, err error) (any, error) {
	if javaSystemFailureCode(err) == ECSystemOutOfMemory {
		PrintError(ECSystemOutOfMemoryTooManyInit)
		f.returnValue = ECSystemOutOfMemoryTooManyInit
		return f.returnValue, nil
	}
	f.errState = curState
	f.err = err
	if isJavaAbortingInitError(err) {
		return f.returnValue, err
	}
	return f.returnValue, nil
}

func isJavaAbortingInitError(err error) bool {
	// Java DoInitFunctor immediately rethrows only invariant, Assert, and eval
	// failures. Other throwables are recorded and surfaced after getInitStates.
	if isDoInitInvariantViolatedException(err) {
		return true
	}
	return isJavaEvalOrRuntimeException(err)
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
