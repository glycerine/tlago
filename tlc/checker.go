package tlc

import (
	"errors"
	"fmt"
	"math"
	"os"
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

func NewModelChecker(tool *Tool, metadir string, deadlock bool, opts ...ModelCheckerOption) *ModelChecker {
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
		LiveCheck:       NewNoOpLiveCheck(tool, metadir),
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
	if mc.ConcurrentTrace == nil {
		mc.ConcurrentTrace = NewConcurrentTLCTrace(metadir, rootName)
	}
	mc.ConcurrentTrace.TLCTrace = mc.Trace
	mc.ConcurrentTrace.SetTool(tool)
	if mc.LiveCheck == nil {
		mc.LiveCheck = NewNoOpLiveCheck(tool, metadir)
	}
	mc.initWorkers()
	SetMainChecker(mc)
	return mc
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
		intValueFromDurationSince(TLCStartTime()),
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

func (mc *ModelChecker) ModelCheck() (int, error) {
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(mc.Tool)
		defer ReportCoverage(mc.Tool, mc.StartTime)
	}
	if result := mc.CheckAssumptions(); result != NoError {
		return result, nil
	}
	recovered, err := mc.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	result := NoError
	if !recovered {
		PrintMessage(ECTLCComputingInit)
		result, err = mc.DoInit(false)
		if err != nil || result != NoError {
			if result != NoError {
				mc.checkPostConditionAfterInitFailure()
			}
			mc.PrintSummary(false)
			return result, err
		}
		mc.PrintInitGenerated()
	}
	if len(mc.Tool.GetActions()) == 0 {
		if !mc.StateQueue.IsEmpty() {
			PrintError(ECTLCStatesAndNoNextAction)
			mc.PrintSummary(false)
			return ECTLCStatesAndNoNextAction, nil
		}
		ReportSuccess(mc.FPSet, mc.GetStatesGenerated())
		mc.PrintSummary(true)
		return NoError, nil
	}
	result, err = mc.RunTLC(0)
	if err != nil || result != NoError {
		mc.PrintSummary(false)
		return result, err
	}
	if mc.CheckLiveness && mc.LiveCheck != nil {
		result, err = mc.LiveCheck.FinalCheck(mc.Tool)
		if err != nil || result != NoError {
			mc.PrintSummary(false)
			return result, err
		}
	}
	result = mc.Tool.CheckPostCondition()
	if result == NoError {
		ReportSuccess(mc.FPSet, mc.GetStatesGenerated())
	}
	mc.PrintSummary(result == NoError)
	return result, nil
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
	if err := RecoverUniqueStrings(mc.FromCheckpoint); err != nil {
		return false, err
	}
	if mc.Trace != nil {
		mc.Trace.SetCheckpointContext(mc.FromCheckpoint, mc.Tool.GetRootName())
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
	_ = success
	if mc == nil {
		return nil
	}
	var err error
	vetoCleanup := modelCheckerVetoCleanup()
	if cleanup && CheckpointExplicitlyEnabled() && mc.StateQueue != nil && !mc.StateQueue.IsEmpty() && (mc.ErrState != nil || mc.TimeBound) {
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
		if mc.StateQueue != nil {
			if deleteErr := mc.StateQueue.Delete(); deleteErr != nil && err == nil {
				err = deleteErr
			}
		}
		if mc.ConcurrentTrace != nil {
			if deleteErr := mc.ConcurrentTrace.Delete(); deleteErr != nil && err == nil {
				err = deleteErr
			}
		} else if mc.Trace != nil {
			if deleteErr := mc.Trace.Delete(); deleteErr != nil && err == nil {
				err = deleteErr
			}
			for _, worker := range mc.Workers {
				if worker != nil {
					if deleteErr := worker.DeleteTrace(); deleteErr != nil && err == nil {
						err = deleteErr
					}
				}
			}
		}
		if mc.Metadir != "" {
			if deleteErr := os.RemoveAll(mc.Metadir); deleteErr != nil && err == nil {
				err = deleteErr
			}
		}
	}
	return err
}

func modelCheckerVetoCleanup() bool {
	for _, key := range []string{"tlc2.tool.ModelChecker.vetoCleanup", "TLAGO_MODEL_CHECKER_VETO_CLEANUP"} {
		if value, ok := os.LookupEnv(key); ok {
			return strings.EqualFold(value, "true")
		}
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
			return mc.ErrorCode, joinErr
		}
		return ECGeneral, joinErr
	}
	return mc.ErrorCode, nil
}

func (mc *ModelChecker) waitForWorkersWithPeriodicWork(maxDepth int) (int, error) {
	done := make(chan error, 1)
	go func() {
		done <- mc.joinWorkers()
	}()
	interval := ProgressInterval()
	if interval <= 0 {
		return NoError, <-done
	}
	coverageCountdown := periodicCoverageCountdown(interval)
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case err := <-done:
			return NoError, err
		case <-timer.C:
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
			if maxDepth > 0 && mc.GetProgress() > int64(maxDepth) {
				if mc.StateQueue != nil {
					mc.StateQueue.FinishAll()
				}
				return NoError, <-done
			}
			if coverageCountdown == 0 {
				mc.reportPeriodicCoverage()
				coverageCountdown = periodicCoverageCountdown(interval)
			} else if coverageCountdown > 0 {
				coverageCountdown--
			}
			timer.Reset(interval)
		}
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
	if mc == nil || mc.Tool == nil || !CoverageAnyEnabled() || len(mc.Tool.GetActions()) == 0 {
		return
	}
	reportCoverage(mc.Tool)
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
	mc.PrintProgressStats(time.Time{}, false)
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
			return result, err
		}
	} else if mc.RuntimeRatio > LivenessRatio() {
		mc.UpdateRuntimeRatio(0)
	}
	if periodic != nil {
		value, err := mc.Tool.NoDebug().Eval(periodic)
		if err != nil {
			return ECTLCAssumptionEvaluationError, err
		}
		if boolValue, ok := value.(*BoolValue); ok && !boolValue.Val {
			return ECTLCAssumptionFalse, nil
		}
	}
	if createCheckpoint {
		if err := mc.Checkpoint(); err != nil {
			return ECSystemCheckpointRecoveryCorrupt, err
		}
		resume = false
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
	if mc.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "model checker has no tool")
	}
	functor := &doInitFunctor{mc: mc, tool: mc.Tool, forceChecks: ignoreCancel, returnValue: NoError}
	err := mc.Tool.GetInitStates(NewStateFunctor(functor.AddElement))
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

func (mc *ModelChecker) DoNext(curState *TLCStateMut) (bool, error) {
	if mc.Tool == nil {
		return true, newTLCError(ECGeneral, "model checker has no tool")
	}
	restoreRandomState := PushRandomEnumerableState(curState)
	defer restoreRandomState()
	restoreCurrentState := PushCurrentState(curState)
	defer restoreCurrentState()
	deadLocked := true
	var succState *TLCStateMut
	for _, action := range mc.Tool.GetActions() {
		nextStates, err := mc.Tool.GetNextStates(action, curState)
		if err != nil {
			mc.doNextFailed(curState, succState, err)
			return true, err
		}
		size := 0
		if nextStates != nil {
			size = nextStates.Size()
		}
		if action != nil && CoverageEnabled() {
			action.CM.IncInvocations(int64(size))
		}
		mc.NextStatesGenerated += int64(size)
		deadLocked = deadLocked && size == 0
		for i := 0; i < size; i++ {
			succState = nextStates.At(i)
			stop, _, err := mc.processSuccessorForWorker(0, curState, succState, action, nil)
			if stop || err != nil {
				return stop, err
			}
		}
		succState = nil
	}
	if deadLocked && mc.CheckDeadlock {
		return mc.doNextSetErr(curState, nil, false, ECTLCDeadlockReached, ""), nil
	}
	return false, nil
}

func (mc *ModelChecker) processSuccessorForWorker(workerID int, curState *TLCStateMut, succState *TLCStateMut, action *Action, collectedStates *SetOfStates) (bool, bool, error) {
	if succState != nil {
		succState.SetPredecessor(curState).SetAction(action)
	}
	if !mc.Tool.IsGoodState(succState) {
		return mc.doNextSetErrParams(curState, succState, false, ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(mc.Tool, action, succState)...), false, nil
	}
	inModel, err := mc.Tool.IsInModel(succState)
	if err != nil {
		mc.doNextEvalFailed(curState, succState, ECGeneral, "", err)
		return true, false, err
	}
	if inModel {
		inActions, err := mc.Tool.IsInActions(curState, succState)
		if err != nil {
			mc.doNextEvalFailed(curState, succState, ECGeneral, "", err)
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
	wrote := false
	for _, constraint := range mc.Tool.GetModelConstraints() {
		ok, err := mc.Tool.IsInModelForConstraint(constraint, succState)
		if err != nil {
			return err
		}
		if !ok {
			wrote = true
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
			wrote = true
			if err := mc.AllStateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action, constraint); err != nil {
				return err
			}
		}
	}
	if !wrote {
		return mc.AllStateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action)
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
		if mc.Tool != nil {
			mc.Tool.RememberState(succState)
		}
		if worker := mc.workerAt(workerID); worker != nil {
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
	invariants := mc.Tool.GetInvariants()
	names := mc.Tool.GetInvNames()
	for i, invariant := range invariants {
		valid, err := mc.Tool.IsValidState(invariant, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCInvariantEvaluationFailed, nameAt(names, i), err)
		}
		if !valid {
			if continuationEnabled() {
				PrintError(ECTLCInvariantViolatedBehavior, nameAt(names, i))
				mc.printBehaviorTrace(curState, succState)
				continue
			}
			return mc.doNextSetErr(curState, succState, false, ECTLCInvariantViolatedBehavior, nameAt(names, i)), nil
		}
	}
	return false, nil
}

func (mc *ModelChecker) doNextCheckImplied(curState *TLCStateMut, succState *TLCStateMut) (bool, error) {
	implied := mc.Tool.GetImpliedActions()
	names := mc.Tool.GetImpliedActNames()
	for i, action := range implied {
		valid, err := mc.Tool.IsValidTransition(action, curState, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCActionPropertyEvaluationFailed, nameAt(names, i), err)
		}
		if !valid {
			if continuationEnabled() {
				PrintError(ECTLCActionPropertyViolatedBehavior, nameAt(names, i))
				mc.printBehaviorTrace(curState, succState)
				continue
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
	isConsole := !mc.Done
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
	mc.checkPostConditionWithErrorTrace(curState, succState, isConsole)
	return true
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
		var previous *TLCStateMut
		if i > 0 && trace[i-1] != nil {
			previous = trace[i-1].OriginalState()
		}
		PrintInvariantViolationStateTraceState(info, previous, i+1, i == len(trace)-1)
	}
}

func (mc *ModelChecker) errorTraceInfo(curState *TLCStateMut, succState *TLCStateMut) []*TLCStateInfo {
	if succState != nil {
		if curState != nil {
			return appendTraceStateIfMissing(mc.traceInfoPrefix(curState), succState)
		}
		return appendTraceStateIfMissing(mc.traceInfoPrefix(succState), succState)
	}
	if curState != nil {
		return appendTraceStateIfMissing(mc.traceInfoPrefix(curState), curState)
	}
	return nil
}

func (mc *ModelChecker) traceInfoPrefix(state *TLCStateMut) []*TLCStateInfo {
	if state == nil {
		return nil
	}
	if mc != nil && mc.ConcurrentTrace != nil {
		if trace := mc.ConcurrentTrace.GetTraceFromState(state); len(trace) > 0 {
			return trace
		}
	}
	if mc != nil && mc.Trace != nil {
		return mc.Trace.GetTrace(state)
	}
	return NewTLCTrace().GetTrace(state)
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
	ec := ECGeneral
	if tlcErr, ok := err.(*TLCError); ok {
		ec = tlcErr.Code
	}
	if mc.SetErrState(curState, succState, true, ec) {
		PrintError(ec, err.Error())
		mc.printBehaviorTrace(curState, succState)
		if mc.StateQueue != nil {
			mc.StateQueue.FinishAll()
		}
	}
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
		f.errState = curState
		f.err = err
		return f.returnValue, err
	}
	seen := false
	if inModel {
		fp := curState.FingerPrint()
		seen = f.mc.FPSet.Put(fp)
		if !seen {
			if f.tool != nil {
				f.tool.RememberState(curState)
			}
			if err := f.mc.AllStateWriter.WriteInitState(curState); err != nil {
				f.errState = curState
				f.err = err
				return f.returnValue, err
			}
			if worker := f.mc.workerAt(0); worker != nil {
				if err := worker.WriteInitState(curState, fp); err != nil {
					f.errState = curState
					f.err = err
					return f.returnValue, err
				}
			} else if f.mc.Trace != nil {
				if err := f.mc.Trace.WriteInitState(curState, fp); err != nil {
					f.errState = curState
					f.err = err
					return f.returnValue, err
				}
			}
			if f.mc.CheckLiveness && f.mc.LiveCheck != nil {
				if err := f.mc.LiveCheck.AddInitState(f.tool, curState, fp); err != nil {
					f.errState = curState
					f.err = err
					return f.returnValue, err
				}
			}
			f.mc.StateQueue.Enqueue(curState)
		}
	}
	if !seen || f.forceChecks {
		for i, invariant := range f.tool.GetInvariants() {
			valid, err := f.tool.IsValidState(invariant, curState)
			if err != nil {
				f.errState = curState
				f.err = err
				return f.returnValue, err
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
				f.errState = curState
				f.err = err
				return f.returnValue, err
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
