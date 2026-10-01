package tlc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var errInvariantViolated = errors.New("tlc invariant violated")

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
	AllStateWriter            *StateWriter
	Workers                   []*Worker
	PrintedLivenessErrorStack bool
	Config                    Value
	StartTime                 time.Time
}

func NewAbstractChecker(tool *Tool, metadir string, stateWriter *StateWriter, deadlock bool, fromCheckpoint string, startTime time.Time) *AbstractChecker {
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
	checker.Config = checker.createConfig()
	return checker
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
	NumberOfInitialStates   int64
	FPSet                   FPSet
	StateQueue              StateQueue
	Trace                   *TLCTrace
	ConcurrentTrace         *ConcurrentTLCTrace
	LiveCheck               *LiveCheck
	NextStatesGenerated     int64
	StatesPerMinute         int64
	DistinctStatesPerMinute int64
	OldNumOfGenStates       int64
	OldFPSetSize            uint64
	RuntimeRatio            float64
	ForceLiveCheck          bool
	TimeBound               bool
	LiveCheckInitErr        error
	CleanupEnabled          bool
	cleanupDone             bool
}

type ModelCheckerOption func(*ModelChecker)

func WithModelCheckerFPSet(fpSet FPSet) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.FPSet = fpSet
	}
}

func WithModelCheckerStateQueue(queue StateQueue) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.StateQueue = queue
	}
}

func WithModelCheckerStateWriter(writer *StateWriter) ModelCheckerOption {
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
	concurrentTrace := NewConcurrentTLCTrace(metadir, rootName)
	concurrentTrace.SetTool(tool)
	mc := &ModelChecker{
		AbstractChecker: NewAbstractChecker(tool, metadir, NewNoopStateWriter(), checkDeadlock, "", time.Now()),
		FPSet:           NewFPSet(NewFPSetConfiguration()).Init(NumWorkers(), metadir, rootName),
		StateQueue:      NewStateQueue(metadir),
		Trace:           concurrentTrace.TLCTrace,
		ConcurrentTrace: concurrentTrace,
		CleanupEnabled:  true,
	}
	for _, opt := range opts {
		opt(mc)
	}
	if mc.FPSet == nil {
		mc.FPSet = NewFPSet(NewFPSetConfiguration())
	}
	if !fpSetInitialized(mc.FPSet) {
		mc.FPSet = mc.FPSet.Init(NumWorkers(), metadir, rootName)
	}
	if mc.StateQueue == nil {
		mc.StateQueue = NewStateQueue(metadir)
	}
	if !stateQueueInitialized(mc.StateQueue) {
		setStateQueueDir(mc.StateQueue, metadir)
	}
	if mc.Trace == nil {
		mc.Trace = NewTLCTrace(metadir, rootName)
	} else {
		mc.Trace.SetCheckpointContext(metadir, rootName)
	}
	mc.Trace.SetTool(tool)
	if mc.ConcurrentTrace == nil {
		mc.ConcurrentTrace = NewConcurrentTLCTrace(metadir, rootName)
	}
	mc.ConcurrentTrace.TLCTrace = mc.Trace
	mc.ConcurrentTrace.SetTool(tool)
	if mc.LiveCheck == nil {
		if mc.CheckLiveness {
			if tool != nil && tool.HasSymmetry() {
				PrintWarning(ECTLCFeatureUnsupportedLivenessSymmetry)
			}
			if livenessTestingImplementationEnabled() {
				mc.LiveCheck, mc.LiveCheckInitErr = NewAddAndCheckLiveCheckFromTool(tool, metadir)
			} else {
				mc.LiveCheck, mc.LiveCheckInitErr = NewLiveCheckFromTool(tool, metadir, mc.AllStateWriter)
			}
		} else {
			mc.LiveCheck = NewNoOpLiveCheck(tool, metadir)
		}
	}
	if mc.LiveCheck == nil {
		mc.LiveCheck = NewNoOpLiveCheck(tool, metadir)
	}
	mc.initWorkers()
	if scheduleStopAfterFromJavaProperty(mc.Stop) {
		mc.TimeBound = true
	}
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
		return 0
	}
	if mc.ConcurrentTrace != nil {
		return int64(mc.ConcurrentTrace.GetLevelForReporting())
	}
	if mc.Trace == nil {
		return 0
	}
	return int64(mc.Trace.GetLevelForReporting())
}

func (mc *ModelChecker) GetStatistics() Value {
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
	defer func() {
		if mc == nil {
			return
		}
		if cleanupErr := mc.Cleanup(result == NoError, mc.CleanupEnabled); err == nil {
			err = cleanupErr
		}
	}()
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	if mc.LiveCheckInitErr != nil {
		return ECGeneral, mc.LiveCheckInitErr
	}
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(mc.Tool)
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
		PrintMessage(ECTLCComputingInit)
		result, err = mc.DoInit(false)
		if err != nil {
			result = mc.reportInitException(result, err)
			mc.PrintSummary(false)
			return result, err
		}
		if result != NoError {
			mc.checkPostConditionAfterInitFailure()
			return result, nil
		}
		mc.PrintInitGenerated()
	}
	if len(mc.Tool.GetActions()) == 0 {
		if !mc.StateQueue.IsEmpty() {
			PrintError(ECTLCStatesAndNoNextAction)
			result = ECTLCStatesAndNoNextAction
			if cleanupErr := mc.Cleanup(true, mc.CleanupEnabled); cleanupErr != nil {
				err = cleanupErr
			}
			return result, err
		}
		ReportSuccess(mc.FPSet, mc.GetStatesGenerated())
		mc.PrintSummary(true)
		return NoError, nil
	}
	result, err = mc.RunTLC(0)
	if err != nil || result != NoError {
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
			result, err = mc.LiveCheck.FinalCheck(mc.Tool.NoDebug())
			if err != nil || result != NoError {
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
	if err := RecoverUniqueStrings(mc.FromCheckpoint); err != nil {
		return false, err
	}
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
		if err := mc.FPSet.RecoverTrace(mc.Trace); err != nil {
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
		// the metadir here.  The Go port still closes owned resources below;
		// preserving recovery artifacts must not preserve open handles too.
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
		// on success. Keep failure artifacts for inspection/recovery, but close
		// every owned resource above so the failure path does not leak handles.
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
	inDegree, err := mc.LiveCheck.CalculateInDegreeDiskGraphs(NewBucketStatistics("Histogram vertex in-degree"))
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

func printLivenessStatistics(inDegree *BucketStatistics, outDegree *BucketStatistics) {
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
	if factor <= 0 {
		factor = 1
	}
	generated := mc.GetStatesGenerated()
	mc.StatesPerMinute = int64(float64(generated-mc.OldNumOfGenStates) / factor)
	mc.OldNumOfGenStates = generated
	var distinctDelta uint64
	if fpSetSize >= mc.OldFPSetSize {
		distinctDelta = fpSetSize - mc.OldFPSetSize
	}
	mc.DistinctStatesPerMinute = int64(float64(distinctDelta) / factor)
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
	reportCoverage(mc.Tool)
}

func checkerCoverageEnabled(tool *Tool) bool {
	return tool != nil && CoverageAnyEnabled() && len(tool.GetActions()) > 0
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
		periodic = mc.Tool.Periodic
	}
	forceLiveCheck := mc.CheckLiveness && mc.LiveCheck != nil && mc.ForceLiveCheck
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
		mc.ForceLiveCheck = false
		mc.UpdateRuntimeRatio(time.Since(start))
		if err != nil || result != NoError {
			resume = false
			return result, err
		}
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
	if denominator <= 0 {
		mc.RuntimeRatio = 0
		return
	}
	mc.RuntimeRatio = (float64(delta) + absLivenessRuntime) / denominator
}

func (mc *ModelChecker) ForceLivenessCheck() {
	if mc == nil {
		return
	}
	mc.ForceLiveCheck = true
	if mc.LiveCheck != nil {
		mc.LiveCheck.ForceCheck()
	}
}

func (mc *ModelChecker) DoInit(ignoreCancel bool) (int, error) {
	return mc.doInitWithTool(mc.Tool, ignoreCancel)
}

func (mc *ModelChecker) doInitWithTool(tool *Tool, ignoreCancel bool) (int, error) {
	if tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	functor := &doInitFunctor{mc: mc, tool: tool, forceChecks: ignoreCancel, returnValue: NoError}
	err := tool.GetInitStates(NewStateFunctor(functor.AddElement))
	if errors.Is(err, errInvariantViolated) {
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
	if result == NoError {
		result = initExceptionCode(err)
	}
	message := err.Error()
	if message == "" {
		message = fmt.Sprintf("%T", err)
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
	var eval *EvalException
	if errors.As(err, &eval) && eval != nil {
		return eval.GetErrorCode()
	}
	var tlcErr *TLCError
	if errors.As(err, &tlcErr) && tlcErr != nil {
		return tlcErr.Code
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
		var fpErr *FingerprintException
		if errors.As(err, &fpErr) && fpErr != nil {
			trace := fpErr.GetTrace()
			if callStackTool.HasCallStack() {
				trace = callStackTool.CallStackString()
			}
			rootMessage := ""
			if root := fpErr.GetRootCause(); root != nil {
				rootMessage = root.Error()
			}
			PrintError(ECTLCFingerprintException, trace, rootMessage)
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
	rootName := "Spec"
	if mc.Tool.GetRootName() != "" {
		rootName = mc.Tool.GetRootName()
	}
	replayWorker.SetTraceContext(mc.Metadir, rootName)
	if _, err := mc.doNextWithTool(callStackTool, mc.PredErrState, liveNextStates, replayWorker); err != nil {
		var fpErr *FingerprintException
		if errors.As(err, &fpErr) && fpErr != nil {
			trace := fpErr.GetTrace()
			if callStackTool.HasCallStack() {
				trace = callStackTool.CallStackString()
			}
			rootMessage := ""
			if root := fpErr.GetRootCause(); root != nil {
				rootMessage = root.Error()
			}
			PrintError(ECTLCFingerprintException, trace, rootMessage)
			return ECTLCFingerprintException
		}
		var eval *EvalException
		if errors.As(err, &eval) && eval != nil {
			PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
			return eval.GetErrorCode()
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

func (mc *ModelChecker) doNextWithTool(tool *Tool, curState *TLCStateMut, liveNextStates *SetOfStates, worker *Worker) (bool, error) {
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
				seen, err := mc.isSeenStateUsingWorker(workerIDForReplayWorker(worker), worker, curState, succState, action)
				if err != nil {
					mc.doNextFailed(curState, succState, err)
					return true, err
				}
				unseen = !seen
				if liveNextStates != nil {
					liveNextStates.PutFP(succState.FingerPrint(), succState)
				}
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

func (mc *ModelChecker) processSuccessorForWorker(workerID int, curState *TLCStateMut, succState *TLCStateMut, action *Action, collectedStates *SetOfStates) (bool, bool, error) {
	if !mc.Tool.IsGoodState(succState) {
		return mc.doNextSetErrParamsWithPostCondition(curState, succState, false, ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(mc.Tool, action, succState)...), false, nil
	}
	if succState != nil {
		succState.SetPredecessor(curState).SetAction(action)
	}
	inModel, err := mc.Tool.IsInModel(succState)
	if err != nil {
		return true, false, err
	}
	if inModel {
		inActions, err := mc.Tool.IsInActions(curState, succState)
		if err != nil {
			return true, false, err
		}
		inModel = inActions
	}
	unseen := true
	if inModel {
		seen, err := mc.isSeenState(workerID, curState, succState, action)
		if err != nil {
			return true, false, err
		}
		if collectedStates != nil {
			collectedStates.PutFP(succState.FingerPrint(), succState)
		}
		unseen = !seen
	} else if mc.AllStateWriter != nil && mc.AllStateWriter.IsConstrained() {
		if err := mc.writeConstrainedTransitionReasons(curState, succState, action); err != nil {
			return true, false, err
		}
	}
	if unseen {
		stop, err := mc.doNextCheckInvariants(curState, succState)
		if stop || err != nil {
			return stop, false, err
		}
	}
	stop, err := mc.doNextCheckImplied(curState, succState)
	if stop || err != nil {
		return stop, false, err
	}
	if inModel && unseen {
		mc.StateQueue.SEnqueue(succState)
		CountStateVariableCoverage(succState)
		return false, true, nil
	}
	return false, false, nil
}

func (mc *ModelChecker) writeConstrainedTransitionReasons(curState *TLCStateMut, succState *TLCStateMut, action *Action) error {
	if mc == nil || mc.Tool == nil || mc.AllStateWriter == nil {
		return nil
	}
	for _, constraint := range mc.Tool.GetModelConstraints() {
		ok, err := mc.Tool.IsInModelForConstraint(constraint, succState)
		if err != nil {
			return err
		}
		if !ok {
			if err := mc.AllStateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action, constraint); err != nil {
				return err
			}
		}
	}
	for _, constraint := range mc.Tool.GetActionConstraints() {
		ok, err := mc.Tool.IsInActionsForConstraint(constraint, curState, succState)
		if err != nil {
			return err
		}
		if !ok {
			if err := mc.AllStateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action, constraint); err != nil {
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
			total += worker.StatesGenerated
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

func (mc *ModelChecker) isSeenState(workerID int, curState *TLCStateMut, succState *TLCStateMut, action *Action) (bool, error) {
	return mc.isSeenStateUsingWorker(workerID, mc.workerAt(workerID), curState, succState, action)
}

func (mc *ModelChecker) isSeenStateUsingWorker(workerID int, worker *Worker, curState *TLCStateMut, succState *TLCStateMut, action *Action) (bool, error) {
	fp := succState.FingerPrint()
	seen := mc.FPSet.Put(fp)
	if mc.AllStateWriter != nil {
		status := StateVisitUnseen
		if seen {
			status = StateVisitSeen
		}
		if err := mc.AllStateWriter.WriteTransition(curState, succState, status, action); err != nil {
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
				PrintError(ECTLCInvariantViolatedBehavior, nameAt(names, i))
				mc.printBehaviorTrace(curState, succState)
				return false, nil
			}
			if withPostCondition {
				return mc.doNextSetErrWithPostCondition(curState, succState, false, ECTLCInvariantViolatedBehavior, nameAt(names, i)), nil
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
				PrintError(ECTLCActionPropertyViolatedBehavior, nameAt(names, i))
				mc.printBehaviorTrace(curState, succState)
				return false, nil
			}
			if withPostCondition {
				return mc.doNextSetErrWithPostCondition(curState, succState, false, ECTLCActionPropertyViolatedBehavior, nameAt(names, i)), nil
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
	if param == "" {
		return mc.doNextSetErrParamsWithPostCondition(curState, succState, keep, ec)
	}
	return mc.doNextSetErrParamsWithPostCondition(curState, succState, keep, ec, param)
}

func (mc *ModelChecker) doNextSetErrParamsWithPostCondition(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, params ...string) bool {
	isConsole := !mc.Done
	result := mc.doNextSetErrParams(curState, succState, keep, ec, params...)
	mc.checkPostConditionWithErrorTrace(curState, succState, isConsole)
	return result
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
	if mc == nil || mc.Tool == nil {
		return
	}
	trace := mc.errorTraceInfo(curState, succState)
	if len(trace) == 0 {
		mc.Tool.CheckPostCondition()
		return
	}
	trace = mc.aliasErrorTrace(trace)
	mc.Tool.CheckPostConditionWithCounterExample(NewCounterExample(trace, UnknownAction, 0, isConsole))
}

func (mc *ModelChecker) printBehaviorTrace(curState *TLCStateMut, succState *TLCStateMut) {
	trace := mc.errorTraceInfo(curState, succState)
	if len(trace) == 0 {
		return
	}
	PrintError(ECTLCBehaviorUpToThisPoint)
	trace = mc.aliasErrorTrace(trace)
	for i, info := range trace {
		previous := javaTracePrintPredecessor(trace, i, curState, succState)
		PrintInvariantViolationStateTraceState(info, previous, i+1, i == len(trace)-1)
	}
}

func javaTracePrintPredecessor(trace []*TLCStateInfo, index int, curState *TLCStateMut, succState *TLCStateMut) *TLCStateMut {
	if index < 0 || index >= len(trace) {
		return nil
	}
	if curState != nil && curState.IsInitial() {
		if succState != nil {
			return curState
		}
		return nil
	}
	if index == len(trace)-1 {
		return nil
	}
	if index > 0 && trace[index-1] != nil {
		return trace[index-1].OriginalState()
	}
	return nil
}

func (mc *ModelChecker) errorTraceInfo(curState *TLCStateMut, succState *TLCStateMut) []*TLCStateInfo {
	if curState == nil {
		if succState == nil {
			return nil
		}
		trace := mc.traceInfoPrefix(succState)
		return append(trace, mc.stateInfoForState(succState, nil))
	}
	if succState == nil {
		if curState.IsInitial() {
			return []*TLCStateInfo{mc.stateInfoForState(curState, nil)}
		}
		trace := mc.traceInfoPrefix(curState)
		return append(trace, mc.stateInfoForState(curState, lastTraceState(trace)))
	}
	if succState.AllAssigned() && succState.WorkerID == TLCStateInitWorkerID {
		if curState.IsInitial() {
			return []*TLCStateInfo{
				mc.stateInfoForState(curState, nil),
				mc.stateInfoForTransition(succState, curState),
			}
		}
		trace := mc.traceInfoPrefix(curState)
		trace = append(trace, mc.stateInfoForState(curState, lastTraceState(trace)))
		return append(trace, mc.stateInfoForTransition(succState, curState))
	}
	trace := mc.traceInfoPrefix(succState)
	return append(trace, mc.stateInfoForTransition(succState, curState))
}

func lastTraceState(trace []*TLCStateInfo) *TLCStateMut {
	if len(trace) == 0 || trace[len(trace)-1] == nil {
		return nil
	}
	return trace[len(trace)-1].State
}

func (mc *ModelChecker) stateInfoForState(state *TLCStateMut, predecessor *TLCStateMut) *TLCStateInfo {
	if state == nil {
		return nil
	}
	if mc != nil && mc.Tool != nil {
		var (
			info *TLCStateInfo
			err  error
		)
		fp := state.FingerPrint()
		if predecessor == nil {
			info, err = mc.Tool.GetState(fp)
		} else {
			info, err = mc.Tool.GetState(fp, predecessor)
		}
		if err == nil && info != nil && info.State != nil {
			info.State.WorkerID = state.WorkerID
			info.State.UID = state.UID
			return info
		}
	}
	info := NewTLCStateInfo(state)
	fp := state.FingerPrint()
	info.FP = &fp
	return info
}

func (mc *ModelChecker) stateInfoForTransition(state *TLCStateMut, predecessor *TLCStateMut) *TLCStateInfo {
	if state == nil {
		return nil
	}
	if mc != nil && mc.Tool != nil && predecessor != nil {
		info, err := mc.Tool.GetStateForTransition(state, predecessor)
		if err == nil && info != nil && info.State != nil {
			info.State.WorkerID = state.WorkerID
			info.State.UID = state.UID
			return info
		}
	}
	info := NewTLCStateInfo(state)
	fp := state.FingerPrint()
	info.FP = &fp
	return info
}

func trimTraceState(trace []*TLCStateInfo, state *TLCStateMut) []*TLCStateInfo {
	if len(trace) == 0 || state == nil {
		return trace
	}
	last := trace[len(trace)-1]
	if last != nil && last.State != nil && (last.State == state || last.State.Equal(state)) {
		return trace[:len(trace)-1]
	}
	return trace
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
	if state == nil || state.IsInitial() {
		return nil
	}
	if mc != nil && mc.ConcurrentTrace != nil {
		if trace := mc.ConcurrentTrace.GetTraceFromState(state); len(trace) > 0 {
			return trimTraceState(trace, state)
		}
	}
	if mc != nil && mc.Trace != nil {
		return trimTraceState(mc.Trace.GetTrace(state), state)
	}
	return trimTraceState(NewTLCTrace().GetTrace(state), state)
}

func (mc *ModelChecker) aliasErrorTrace(trace []*TLCStateInfo) []*TLCStateInfo {
	if mc == nil || mc.Tool == nil || len(trace) == 0 {
		return trace
	}
	aliased := append([]*TLCStateInfo(nil), trace...)
	for i, current := range aliased {
		successor := current.OriginalState()
		if i+1 < len(aliased) {
			successor = aliased[i+1].OriginalState()
		}
		alias, err := mc.Tool.EvalAliasInfo(current, successor, func() []*TLCStateInfo {
			return append([]*TLCStateInfo(nil), aliased[:i]...)
		})
		if err == nil && alias != nil {
			aliased[i] = alias
		}
	}
	return aliased
}

func (mc *ModelChecker) doNextEvalFailed(curState *TLCStateMut, succState *TLCStateMut, ec int, param string, err error) error {
	if mc.SetErrState(curState, succState, true, ec) {
		if param == "" {
			PrintError(ec, err.Error())
		} else {
			PrintError(ec, param, err.Error())
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
	ec, params, keepCallStack := doNextFailureMessage(err)
	if mc.SetErrState(curState, succState, keepCallStack, ec) {
		if params != nil && len(params) > 0 {
			PrintError(ec, params...)
		} else if params != nil {
			PrintError(ec)
		}
		mc.printBehaviorTrace(curState, succState)
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
}

func doNextFailureMessage(err error) (int, []string, bool) {
	ec := ECGeneral
	params := []string{err.Error()}
	keepCallStack := true

	var eval *EvalException
	if errors.As(err, &eval) && eval != nil {
		ec = eval.GetErrorCode()
		if eval.HasParameters() {
			params = eval.GetParameters()
		} else if ec == ECGeneral {
			params = generalErrorParams("", err)
		}
		return ec, params, keepCallStack
	}

	var tlcErr *TLCError
	if errors.As(err, &tlcErr) && tlcErr != nil {
		if tlcErr.Code == ECSystemStackOverflow || tlcErr.Code == ECSystemOutOfMemory || tlcErr.Code == ECTLCBug {
			return tlcErr.Code, params, false
		}
	}

	return ECGeneral, generalErrorParams("", err), keepCallStack
}

func generalErrorParams(cause string, err error) []string {
	if err == nil || err.Error() == "" {
		return nil
	}
	return []string{javaGeneralErrorMessage(cause, err)}
}

type doInitFunctor struct {
	mc          *ModelChecker
	tool        *Tool
	forceChecks bool
	errState    *TLCStateMut
	err         error
	returnValue int
}

func (f *doInitFunctor) AddElement(curState *TLCStateMut) (any, error) {
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
	if !f.tool.IsGoodState(curState) {
		PrintError(ECTLCInitialState, "current state is not a legal state", curState.String())
		f.errState = curState
		f.returnValue = ECTLCInitialState
		return f.returnValue, errInvariantViolated
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
					return f.returnValue, errInvariantViolated
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
				return f.returnValue, errInvariantViolated
			}
		}
	}
	return f.returnValue, nil
}

func (f *doInitFunctor) handleInitError(curState *TLCStateMut, err error) (any, error) {
	f.errState = curState
	f.err = err
	if isJavaAbortingInitError(err) {
		return f.returnValue, err
	}
	return f.returnValue, nil
}

func isJavaAbortingInitError(err error) bool {
	if err == nil {
		return false
	}
	// Java DoInitFunctor immediately rethrows only invariant, Assert, and eval
	// failures. Other throwables are recorded and surfaced after getInitStates.
	if errors.Is(err, errInvariantViolated) {
		return true
	}
	var eval *EvalException
	if errors.As(err, &eval) && eval != nil {
		return true
	}
	var tlcErr *TLCError
	if errors.As(err, &tlcErr) && tlcErr != nil {
		return true
	}
	return false
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
