package tlc

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const javaIntegerMaxValue = int(1<<31 - 1)

var modelCheckerStatics struct {
	once           sync.Once
	vetoCleanup    bool
	actionCoverage bool
}

var abstractCheckerStatics struct {
	once                          sync.Once
	livenessTestingImplementation bool
	livenessStats                 bool
}

const (
	livenessTestingImplementationProperty = "tlc2.tool.liveness.ILiveCheck.testing"
	livenessStatsProperty                 = "tlc2.tool.liveness.statistics"
)

// InitializeAbstractCheckerStatics represents a fresh Java classloader for
// the property-backed initial values of AbstractChecker's static fields.
func InitializeAbstractCheckerStatics() {
	abstractCheckerStatics.once = sync.Once{}
	abstractCheckerStatics.livenessTestingImplementation = false
	abstractCheckerStatics.livenessStats = false
}

func initializeAbstractCheckerStatics() {
	abstractCheckerStatics.once.Do(func() {
		if value, ok := tlcLookupSystemProperty(livenessTestingImplementationProperty); ok {
			abstractCheckerStatics.livenessTestingImplementation = javaBooleanProperty(value)
		}
		if value, ok := tlcLookupSystemProperty(livenessStatsProperty); ok {
			abstractCheckerStatics.livenessStats = javaBooleanProperty(value)
		}
	})
}

// InitializeModelCheckerStatics represents a fresh Java classloader. Ordinary
// model checker construction retains the cleanup veto captured at class load.
func InitializeModelCheckerStatics() {
	modelCheckerStatics.once = sync.Once{}
	modelCheckerStatics.vetoCleanup = false
	modelCheckerStatics.actionCoverage = false
}

func initializeModelCheckerStatics() {
	modelCheckerStatics.once.Do(func() {
		modelCheckerStatics.actionCoverage = CoverageActionEnabled()
		if value, ok := tlcLookupSystemProperty(modelCheckerVetoProperty); ok {
			modelCheckerStatics.vetoCleanup = javaBooleanProperty(value)
			return
		}
		if value, ok := os.LookupEnv("TLAGO_MODEL_CHECKER_VETO_CLEANUP"); ok {
			modelCheckerStatics.vetoCleanup = javaBooleanProperty(value)
		}
	})
}

type AbstractChecker struct {
	mu                        distributedServerMonitor
	PredErrState              *TLCStateMut
	ErrState                  *TLCStateMut
	ErrorCode                 int
	Done                      atomic.Bool
	KeepCallStack             bool
	CheckDeadlock             bool
	CheckLiveness             bool
	FromCheckpoint            *string
	Metadir                   string
	Tool                      *Tool
	AllStateWriter            IStateWriter
	Workers                   []*Worker
	dfidWorkerSlots           *[]*DFIDWorker
	PrintedLivenessErrorStack bool
	Config                    Value
	LiveCheck                 *LiveCheck
	StartTime                 time.Time

	configurationCaptured bool
	checkDeadlockOwner    bool
	checkLivenessOwner    bool
	fromCheckpointOwner   *string
	metadirOwner          string
	toolOwner             *Tool
	allStateWriterOwner   IStateWriter
	liveCheckOwner        *LiveCheck
	configOwner           Value
	startTimeOwner        time.Time
}

func (c *AbstractChecker) require() {
	if c == nil {
		panic(NewNullPointerException())
	}
}

func (c *AbstractChecker) captureConfiguration() {
	c.checkDeadlockOwner = c.CheckDeadlock
	c.checkLivenessOwner = c.CheckLiveness
	if c.FromCheckpoint != nil {
		fromCheckpoint := *c.FromCheckpoint
		c.fromCheckpointOwner = &fromCheckpoint
	}
	c.metadirOwner = c.Metadir
	c.toolOwner = c.Tool
	c.allStateWriterOwner = c.AllStateWriter
	c.liveCheckOwner = c.LiveCheck
	c.configOwner = c.Config
	c.startTimeOwner = c.StartTime
	c.configurationCaptured = true
}

func (c *AbstractChecker) ownedCheckDeadlock() bool {
	if c.configurationCaptured {
		return c.checkDeadlockOwner
	}
	return c.CheckDeadlock
}

func (c *AbstractChecker) ownedCheckLiveness() bool {
	if c.configurationCaptured {
		return c.checkLivenessOwner
	}
	return c.CheckLiveness
}

func (c *AbstractChecker) ownedFromCheckpoint() *string {
	if c.configurationCaptured {
		return c.fromCheckpointOwner
	}
	return c.FromCheckpoint
}

func (c *AbstractChecker) ownedMetadir() string {
	if c.configurationCaptured {
		return c.metadirOwner
	}
	return c.Metadir
}

func (c *AbstractChecker) ownedTool() *Tool {
	if c.configurationCaptured {
		return c.toolOwner
	}
	return c.Tool
}

func (c *AbstractChecker) ownedAllStateWriter() IStateWriter {
	if c.configurationCaptured {
		return c.allStateWriterOwner
	}
	return c.AllStateWriter
}

func (c *AbstractChecker) ownedLiveCheck() *LiveCheck {
	if c.configurationCaptured {
		return c.liveCheckOwner
	}
	return c.LiveCheck
}

func (c *AbstractChecker) ownedConfig() Value {
	if c.configurationCaptured {
		return c.configOwner
	}
	return c.Config
}

func (c *AbstractChecker) ownedStartTime() time.Time {
	if c.configurationCaptured {
		return c.startTimeOwner
	}
	return c.StartTime
}

func NewAbstractChecker(tool *Tool, metadir string, stateWriter IStateWriter, deadlock bool, fromCheckpoint *string, startTime time.Time) *AbstractChecker {
	checker := newAbstractCheckerFields(tool, metadir, stateWriter, deadlock, fromCheckpoint, startTime)
	checker.initialize(checker.Stop)
	return checker
}

func newAbstractCheckerFields(tool *Tool, metadir string, stateWriter IStateWriter, deadlock bool, fromCheckpoint *string, startTime time.Time) *AbstractChecker {
	initializeAbstractCheckerStatics()
	if tool == nil {
		panic(NewNullPointerException())
	}
	checkLiveness := !tool.LivenessIsTrue()
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
	c.require()
	if CoverageAnyEnabled() {
		CreateCoverageCostModels(c.Tool)
	}
	if c.CheckLiveness && c.Tool != nil && c.Tool.HasSymmetry() {
		PrintWarning(ECTLCFeatureUnsupportedLivenessSymmetry, []string{}...)
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
	c.captureConfiguration()
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
	c.require()
	panic(NewUnsupportedOperationException("stop not implemented"))
}

func (c *AbstractChecker) Suspend() {
	c.require()
	panic(NewUnsupportedOperationException("suspend not implemented"))
}

func (c *AbstractChecker) Resume() {
	c.require()
	panic(NewUnsupportedOperationException("resume not implemented"))
}

func (c *AbstractChecker) GetProgress() int64 {
	c.require()
	return -1
}

func (c *AbstractChecker) IsRecovery() bool {
	c.require()
	return c.ownedFromCheckpoint() != nil
}

func (c *AbstractChecker) IsTimeBound() bool {
	c.require()
	_, timeBound := stopAfterFromJavaProperty()
	return timeBound
}

func (c *AbstractChecker) ReportCoverage(workers ...any) {
	c.require()
	if len(workers) > 1 {
		panic(NewIllegalArgumentException("ReportCoverage accepts at most one worker array"))
	}
	if CoverageEnabled() && len(c.ownedTool().GetActions()) > 0 {
		ReportCoverage(c.ownedTool(), c.ownedStartTime())
	}
}

func (c *AbstractChecker) Report(message string) {
	c.require()
	DebugPrintMessage(message)
}

func (c *AbstractChecker) GetTraceInfo(_ *TLCStateMut) []*TLCStateInfo {
	c.require()
	panic(NewUnsupportedOperationException("getTraceInfo(TLCState) not implemented for this AbstractChecker"))
}

func (c *AbstractChecker) GetTraceInfoBetween(_ *TLCStateMut, _ *TLCStateMut) []*TLCStateInfo {
	c.require()
	panic(NewUnsupportedOperationException("getTraceInfo(TLCState, TLCState) not implemented for this AbstractChecker"))
}

func (c *AbstractChecker) GetStateQueueSize() int64 {
	c.require()
	return -1
}

func (c *AbstractChecker) GetDistinctStatesGenerated() int64 {
	c.require()
	return -1
}

func (c *AbstractChecker) GetStatesGenerated() int64 {
	c.require()
	return -1
}

func (c *AbstractChecker) GetInitialStatesGenerated() int64 {
	c.require()
	return -1
}

func (c *AbstractChecker) GetStatistics() Value {
	c.require()
	return runtimeCheckerStatistics(c)
}

func (c *AbstractChecker) GetConfig() Value {
	c.require()
	return c.ownedConfig()
}

func (c *AbstractChecker) distinctStateCount() int64 {
	c.require()
	return c.GetDistinctStatesGenerated()
}

func (c *AbstractChecker) traceInfoPrefix(state *TLCStateMut) []*TLCStateInfo {
	c.require()
	return c.GetTraceInfo(state)
}

func (c *AbstractChecker) SetDone() bool {
	c.require()
	old := c.Done.Load()
	c.Done.Store(true)
	return old
}

func (c *AbstractChecker) SetErrState(curState *TLCStateMut, succState *TLCStateMut, keepCallStack bool, errorCode int) bool {
	c.require()
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.setErrStateLocked(curState, succState, keepCallStack, errorCode)
}

func (c *AbstractChecker) setErrStateLocked(curState *TLCStateMut, succState *TLCStateMut, keepCallStack bool, errorCode int) bool {
	c.require()
	if !continuationEnabled() && c.Done.Load() {
		return false
	}
	ResetCurrentState()
	c.PredErrState = curState
	if succState == nil {
		c.ErrState = curState
	} else {
		c.ErrState = succState
	}
	c.ErrorCode = int(int32(errorCode))
	c.Done.Store(true)
	c.KeepCallStack = keepCallStack
	return true
}

func (c *AbstractChecker) SetError(keepCallStack bool, errorCode int) {
	c.require()
	c.mu.Lock()
	defer c.mu.Unlock()
	ResetCurrentState()
	c.ErrorCode = int(int32(errorCode))
	c.Done.Store(true)
	c.KeepCallStack = keepCallStack
}

func continuationEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.Continuation
}

// checkerRegisterWorker is the register portion of Java's IWorker/IdThread.
type checkerRegisterWorker interface {
	GetLocalValue(int) Value
	SetLocalValue(int, Value)
	GetLocalValues() []Value
	GetNamedRegister(*UniqueString) Value
	SetNamedRegister(*UniqueString, Value)
	GetNamedRegisters() *NamedRegisterMap
	namedRegisterKeys() []*UniqueString
}

func (c *AbstractChecker) registerWorkerCount() int {
	c.require()
	if c.dfidWorkerSlots != nil {
		return len(*c.dfidWorkerSlots)
	}
	return len(c.Workers)
}
func (c *AbstractChecker) registerWorkerAt(i int) checkerRegisterWorker {
	c.require()
	i = int(int32(i))
	count := c.registerWorkerCount()
	if i < 0 || i >= count {
		panic(NewArrayIndexOutOfBoundsException(i, count))
	}
	if c.dfidWorkerSlots != nil {
		w := (*c.dfidWorkerSlots)[i]
		if w == nil {
			panic(NewNullPointerException())
		}
		return w
	}
	w := c.Workers[i]
	if w == nil {
		panic(NewNullPointerException())
	}
	return w
}
func (c *AbstractChecker) GetValue(workerID int, idx int) Value {
	c.require()
	return c.registerWorkerAt(workerID).GetLocalValue(idx)
}
func (c *AbstractChecker) SetAllValues(idx int, value Value) {
	c.require()
	for i := 0; i < c.registerWorkerCount(); i++ {
		c.registerWorkerAt(i).SetLocalValue(idx, value)
	}
}
func (c *AbstractChecker) SetValue(workerID int, idx int, value Value) {
	c.require()
	c.registerWorkerAt(workerID).SetLocalValue(idx, value)
}
func (c *AbstractChecker) SetWorkerValues(idx int, values []Value) {
	c.require()
	for i, value := range values {
		c.registerWorkerAt(i).SetLocalValue(idx, value)
	}
}
func (c *AbstractChecker) GetAllValue(idx int) []Value {
	c.require()
	values := make([]Value, c.registerWorkerCount())
	for i := range values {
		values[i] = c.registerWorkerAt(i).GetLocalValue(idx)
	}
	return values
}
func (c *AbstractChecker) GetAllValues() Value {
	c.require()
	locals := c.registerWorkerAt(0).GetLocalValues()
	pairs := newTLCValueHashMapWithCapacity[Value](len(locals))
	for idx := range locals {
		if c.registerWorkerAt(0).GetLocalValue(idx) == nil {
			continue
		}
		workerValues := make([]Value, c.registerWorkerCount())
		for i := range workerValues {
			workerValues[i] = c.registerWorkerAt(i).GetLocalValue(idx)
		}
		pairs.Set(NewIntValue(int32(idx)), NewTupleValue(workerValues))
	}
	return NewFcnRcdValueFromMap(pairs)
}
func (c *AbstractChecker) GetNamedValue(workerID int, key *UniqueString) Value {
	c.require()
	return c.registerWorkerAt(workerID).GetNamedRegister(key)
}
func (c *AbstractChecker) SetAllNamedValues(key *UniqueString, value Value) {
	c.require()
	for i := 0; i < c.registerWorkerCount(); i++ {
		c.registerWorkerAt(i).SetNamedRegister(key, value)
	}
}
func (c *AbstractChecker) SetNamedValue(workerID int, key *UniqueString, value Value) {
	c.require()
	c.registerWorkerAt(workerID).SetNamedRegister(key, value)
}
func (c *AbstractChecker) SetAllNamedWorkerValues(key *UniqueString, values []Value) {
	c.require()
	for i, value := range values {
		c.registerWorkerAt(i).SetNamedRegister(key, value)
	}
}
func (c *AbstractChecker) GetAllNamedRegisterValues() Value {
	c.require()
	keys := c.registerWorkerAt(0).namedRegisterKeys()
	pairs := newTLCValueHashMap[Value]()
	for _, key := range keys {
		workerValues := make([]Value, c.registerWorkerCount())
		for i := range workerValues {
			workerValues[i] = c.registerWorkerAt(i).GetNamedRegister(key)
		}
		pairs.Set(NewStringValueFromUnique(key), NewTupleValue(workerValues))
	}
	return NewFcnRcdValueFromMap(pairs)
}
func (c *AbstractChecker) GetAllNamedValues(key *UniqueString) []Value {
	c.require()
	values := make([]Value, c.registerWorkerCount())
	for i := range values {
		values[i] = c.registerWorkerAt(i).GetNamedRegister(key)
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
	coordinatorWake         chan struct{} // guarded by AbstractChecker.mu
	NumberOfInitialStates   int64
	FPSet                   FPSet
	FPSetFuture             *FPSetFuture
	fpSetFutureSupplied     bool
	FPSetConfiguration      *FPSetConfiguration
	StateQueue              StateQueue
	Trace                   *TLCTrace
	ConcurrentTrace         *ConcurrentTLCTrace
	NextStatesGenerated     int64
	StatesPerMinute         int64
	DistinctStatesPerMinute int64
	OldNumOfGenStates       int64
	OldFPSetSize            uint64
	runtimeRatio            atomic.Uint64 // Java's private double, stored as float bits.
	forceLiveCheck          atomic.Bool
	TimeBound               bool
	CleanupEnabled          bool
	actionCoverage          bool
	traceCaptured           bool
	concurrentTraceOwner    *ConcurrentTLCTrace
}

func (mc *ModelChecker) require() {
	if mc == nil {
		panic(NewNullPointerException())
	}
}

func (mc *ModelChecker) captureTrace() {
	mc.concurrentTraceOwner = mc.ConcurrentTrace
	mc.traceCaptured = true
}

func (mc *ModelChecker) ownedConcurrentTrace() *ConcurrentTLCTrace {
	if mc.traceCaptured {
		return mc.concurrentTraceOwner
	}
	return mc.ConcurrentTrace
}

func (mc *ModelChecker) ownedTrace() *TLCTrace {
	if mc.traceCaptured {
		if mc.concurrentTraceOwner == nil {
			return nil
		}
		return mc.concurrentTraceOwner.TLCTrace
	}
	return mc.Trace
}

type ModelCheckerOption func(*ModelChecker)

func WithModelCheckerFPSet(fpSet FPSet) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.FPSet = fpSet
	}
}

func WithModelCheckerFPSetFuture(future *FPSetFuture) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.FPSetFuture = future
		mc.fpSetFutureSupplied = true
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
		if isNil(writer) {
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

func WithModelCheckerFromCheckpoint(fromCheckpoint *string) ModelCheckerOption {
	return func(mc *ModelChecker) {
		if mc.AbstractChecker != nil {
			mc.FromCheckpoint = fromCheckpoint
		}
	}
}

func WithModelCheckerStartTime(startTime time.Time) ModelCheckerOption {
	return func(mc *ModelChecker) {
		if mc.AbstractChecker != nil {
			mc.StartTime = startTime
		}
	}
}

func WithModelCheckerCleanup(cleanup bool) ModelCheckerOption {
	return func(mc *ModelChecker) {
		mc.CleanupEnabled = cleanup
	}
}

func NewModelChecker(tool *Tool, metadir string, deadlock bool, opts ...ModelCheckerOption) *ModelChecker {
	// Java initializes AbstractChecker before its ModelChecker subclass.
	initializeAbstractCheckerStatics()
	initializeModelCheckerStatics()
	if tool == nil {
		panic(NewNullPointerException())
	}
	SetTLCStateTool(tool)
	rootName := tool.GetRootName()
	mc := &ModelChecker{
		AbstractChecker: newAbstractCheckerFields(tool, metadir, NewNoopStateWriter(), deadlock, nil, time.Now()),
		CleanupEnabled:  true,
		actionCoverage:  modelCheckerStatics.actionCoverage,
	}
	for _, opt := range opts {
		opt(mc)
	}
	mc.AbstractChecker.initialize(mc.Stop)
	// ModelChecker's private constructor creates the queue, then trace. Its
	// public constructor initializes the selected fingerprint set and workers.
	if isNil(mc.StateQueue) {
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
	mc.captureTrace()

	if mc.fpSetFutureSupplied {
		if mc.FPSetFuture == nil {
			panic(NewNullPointerException())
		}
		set, err := mc.FPSetFuture.Get()
		if err != nil {
			panic(err)
		}
		mc.FPSet = set
	} else if isNil(mc.FPSet) {
		config := mc.FPSetConfiguration
		if config == nil {
			config = NewFPSetConfiguration()
		}
		mc.FPSet = NewFPSet(config)
		if isNil(mc.FPSet) {
			panic(NewNullPointerException())
		}
	}
	if !mc.fpSetFutureSupplied && !fpSetInitialized(mc.FPSet) {
		mc.FPSet = mc.FPSet.Init(NumWorkers(), metadir, rootName)
	}
	mc.initWorkers()
	return mc
}

func livenessTestingImplementationEnabled() bool {
	initializeAbstractCheckerStatics()
	return abstractCheckerStatics.livenessTestingImplementation
}

func (mc *ModelChecker) initWorkers() {
	mc.require()
	if len(mc.Workers) > 0 {
		return
	}
	workerCount := NumWorkers()
	if workerCount < 0 {
		panic(NewNegativeArraySizeException(fmtInt(workerCount)))
	}
	mc.Workers = make([]*Worker, workerCount)
	for id := 0; id < workerCount; id++ {
		workerTool := mc.ownedTool()
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
	if mc == nil || mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if isNil(mc.StateQueue) {
		panic(NewNullPointerException())
	}
	mc.StateQueue.FinishAll()
	mc.notifyCoordinatorLocked()
}

func (mc *ModelChecker) GetTraceInfo(state *TLCStateMut) []*TLCStateInfo {
	if mc == nil || state == nil {
		panic(NewNullPointerException())
	}
	if mc.ownedConcurrentTrace() != nil {
		return mc.ownedConcurrentTrace().GetTrace(state)
	}
	if mc.ownedTrace() != nil {
		return mc.ownedTrace().GetTrace(state)
	}
	panic(NewNullPointerException())
}

func (mc *ModelChecker) GetTraceInfoBetween(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	if mc == nil || to == nil {
		panic(NewNullPointerException())
	}
	if mc.ownedConcurrentTrace() != nil {
		return mc.ownedConcurrentTrace().GetTraceBetween(from, to)
	}
	if mc.ownedTrace() != nil {
		return mc.ownedTrace().GetTraceBetween(from, to)
	}
	panic(NewNullPointerException())
}

func (mc *ModelChecker) Suspend() {
	if mc == nil || mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if isNil(mc.StateQueue) {
		panic(NewNullPointerException())
	}
	mc.StateQueue.SuspendAll()
	mc.notifyCoordinatorLocked()
}

func (mc *ModelChecker) Resume() {
	if mc == nil || mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if isNil(mc.StateQueue) {
		panic(NewNullPointerException())
	}
	mc.StateQueue.ResumeAll()
	mc.notifyCoordinatorLocked()
}

func (mc *ModelChecker) notifyCoordinatorLocked() {
	if mc.coordinatorWake != nil {
		select {
		case mc.coordinatorWake <- struct{}{}:
		default:
		}
	}
}

func (mc *ModelChecker) beginCoordinatorWait() <-chan struct{} {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.coordinatorWake = make(chan struct{}, 1)
	if mc.Done.Load() {
		mc.notifyCoordinatorLocked()
	}
	return mc.coordinatorWake
}

func (mc *ModelChecker) endCoordinatorWait() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.coordinatorWake = nil
}

func (mc *ModelChecker) GetProgress() int64 {
	if mc == nil {
		panic(NewNullPointerException())
	}
	if mc.ownedConcurrentTrace() != nil {
		return int64(mc.ownedConcurrentTrace().GetLevelForReporting())
	}
	if mc.ownedTrace() == nil {
		panic(NewNullPointerException())
	}
	level, err := mc.ownedTrace().GetLevelForReportingWithError()
	if err != nil {
		if isJavaIOException(err) {
			return -1
		}
		panic(err)
	}
	return int64(level)
}

func (mc *ModelChecker) GetStatistics() Value {
	if mc == nil {
		panic(NewNullPointerException())
	}
	return runtimeCheckerStatistics(mc)
}

func (mc *ModelChecker) GetConfig() Value {
	if mc == nil {
		panic(NewNullPointerException())
	}
	return mc.ownedConfig()
}

func (c *AbstractChecker) createConfig() Value {
	c.require()
	ensureTLCGetSetUniqueStrings()
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
		BoolValueOf(c.CheckDeadlock),
		NewIntValue(int32(NumWorkers())),
		NewStringValue(fmt.Sprintf("%d", RandomEnumerableSeed())),
		NewStringValue(fmt.Sprintf("%d", int64(FP64GetIrredPoly()))),
		NewStringValue(TLCInstallLocation()),
	}
	return NewRecordValue(names, values, false)
}

func (mc *ModelChecker) CheckAssumptions() int {
	if mc == nil || mc.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	return mc.ownedTool().CheckAssumptions()
}

// GetModuleFiles returns the specification modules through the supplied name
// resolver, matching ModelChecker.getModuleFiles in Java.
func (mc *ModelChecker) GetModuleFiles(resolver FilenameToStream) []*TLAFile {
	if mc == nil {
		panic(NewNullPointerException())
	}
	return mc.ownedTool().GetModuleFiles(resolver)
}

// ModelCheck preserves AbstractChecker.modelCheck's final wrapper: the concrete
// implementation runs first, and a stored checker error replaces only a
// successful return code.
func (mc *ModelChecker) ModelCheck() (result int, err error) {
	mc.require()
	result, err = mc.ModelCheckImpl()
	if result == NoError && mc.ErrorCode != NoError {
		result = mc.ErrorCode
	}
	return result, err
}

// ModelCheckImpl is the protected ModelChecker.modelCheckImpl boundary.
func (mc *ModelChecker) ModelCheckImpl() (result int, err error) {
	mc.require()
	result = NoError
	mc.Report("entering modelCheck()")
	debugCleanupExit := ""
	cleanupRequired := false
	searchFinalization := false
	cleanupSuccessOverride := false
	hasCleanupSuccessOverride := false
	defer func() {
		if cleanupRequired {
			cleanupSuccess := result == NoError
			if hasCleanupSuccessOverride {
				cleanupSuccess = cleanupSuccessOverride
			}
			if searchFinalization {
				mc.PrintSummary(cleanupSuccess)
				if statsErr := mc.PrintLivenessStatistics(); statsErr != nil {
					err = statsErr
					return
				}
			}
			cleanupErr := mc.Cleanup(cleanupSuccess, mc.CleanupEnabled)
			if cleanupErr != nil {
				err = cleanupErr
			}
			if cleanupErr == nil && debugCleanupExit != "" {
				mc.Report(debugCleanupExit)
			}
		}
	}()
	recovered, err := mc.Recover()
	if err != nil {
		return ECSystemCheckpointRecoveryCorrupt, err
	}
	if !recovered {
		if mc.ownedCheckLiveness() && mc.ownedLiveCheck().GetNumChecker() == 0 {
			PrintError(ECTLCLiveFormulaTautology)
			return ECTLCLiveFormulaTautology, nil
		}
		if result := mc.CheckAssumptions(); result != NoError {
			return result, nil
		}
		proceed := false
		result, err, proceed = mc.processInitialStates()
		if err != nil {
			mc.Report("exception in init")
			DebugPrintThrowable(err)
			result = mc.reportInitException(result, err)
			mc.PrintSummary(false)
			// Java's initialization catch(Throwable) consumes the failure after
			// diagnostics and replay; only cleanup failures escape this boundary.
			cleanupErr := mc.Cleanup(false, mc.CleanupEnabled)
			if cleanupErr == nil {
				mc.Report("exiting, because init failed with exception")
			}
			return result, cleanupErr
		}
		if !proceed {
			return result, nil
		}
		mc.PrintInitGenerated()
	}
	mc.Report("init processed")
	if len(mc.ownedTool().GetActions()) == 0 {
		debugCleanupExit = "exiting with actions.length == 0"
		if !mc.requireStateQueue().IsEmpty() {
			PrintError(ECTLCStatesAndNoNextAction)
			result = ECTLCStatesAndNoNextAction
			cleanupSuccessOverride = true
			hasCleanupSuccessOverride = true
			cleanupRequired = true
			return result, err
		}
		ReportSuccess(mc.requireFPSet(), mc.GetStatesGenerated())
		mc.PrintSummary(true)
		cleanupRequired = true
		return NoError, nil
	}
	debugCleanupExit = "exiting modelCheck()"
	// Java enters its exploration try/finally only after recovery and initial
	// state processing. Earlier returns retain their initialized resources.
	cleanupRequired = true
	searchFinalization = true
	result = ECGeneral
	defer func() {
		if recovered := recover(); recovered != nil {
			result, err = mc.modelCheckSearchException(result, panicValueAsError(recovered))
		}
	}()
	mc.Report("running TLC")
	runResult, runErr := mc.RunTLC(javaIntegerMaxValue)
	if runErr != nil {
		return mc.modelCheckSearchException(result, runErr)
	}
	result = runResult
	if result != NoError {
		mc.Report("TLC terminated with error")
		return result, nil
	}
	if mc.ErrState == nil {
		if mc.ownedCheckLiveness() {
			PrintMessage(ECTLCProgressStats,
				fmtInt64(mc.GetProgress()),
				MessageNumberFormat(mc.GetStatesGenerated()),
				MessageNumberFormat(int64(mc.GetDistinctStatesGenerated())),
				MessageNumberFormat(mc.GetStateQueueSize()),
			)
			mc.Report("checking liveness")
			liveResult, liveErr := mc.ownedLiveCheck().FinalCheck(mc.ownedTool().NoDebug())
			if liveErr != nil {
				return mc.modelCheckSearchException(result, liveErr)
			}
			result = liveResult
			mc.Report("liveness check complete")
			if result != NoError {
				mc.Report("exiting error status on liveness check")
				return result, nil
			}
		}
		result = mc.ownedTool().CheckPostCondition()
		if result == NoError {
			ReportSuccess(mc.requireFPSet(), mc.GetStatesGenerated())
		}
	} else if mc.KeepCallStack {
		result = mc.replayNextErrorCallStack()
	}
	return result, err
}

// processInitialStates is ModelChecker.modelCheckImpl's initialization
// try/catch(Throwable) body. Recovery and fresh-start validation precede it.
func (mc *ModelChecker) processInitialStates() (result int, failure error, proceed bool) {
	mc.require()
	result = NoError
	defer func() {
		if recovered := recover(); recovered != nil {
			failure = panicValueAsError(recovered)
			proceed = false
		}
	}()
	mc.Report("doInit(false)")
	PrintMessage(ECTLCComputingInit)
	result, failure = mc.DoInit(false)
	if failure != nil {
		return result, failure, false
	}
	if result != NoError {
		mc.Report("exiting, because init failed")
		mc.checkPostConditionAfterInitFailure()
		return result, nil, false
	}
	return result, nil, true
}

// modelCheckSearchException applies ModelChecker.modelCheckImpl's outer
// catch(Exception). Java Error families bypass the catch and remain returned.
func (mc *ModelChecker) modelCheckSearchException(result int, err error) (int, error) {
	if err == nil {
		return result, nil
	}
	if isJavaError(err) {
		return result, err
	}
	mc.Report("TLC terminated with error")
	return PrintError(ECGeneral, generalErrorParams("", err)...), nil
}

func (mc *ModelChecker) Checkpoint() error {
	mc.require()
	PrintMessage(ECTLCCheckpointStart, mc.ownedMetadir())
	if err := mc.requireStateQueue().BeginChkpt(); err != nil {
		return err
	}
	if err := mc.BeginTraceChkpt(); err != nil {
		return err
	}
	if err := mc.requireFPSet().BeginChkpt(); err != nil {
		return err
	}
	if err := BeginChkptUniqueStrings(mc.ownedMetadir()); err != nil {
		return err
	}
	if mc.ownedCheckLiveness() {
		if err := mc.ownedLiveCheck().BeginChkpt(); err != nil {
			return err
		}
	}
	mc.requireStateQueue().ResumeAll()
	if err := mc.requireStateQueue().CommitChkpt(); err != nil {
		return err
	}
	if err := mc.CommitTraceChkpt(); err != nil {
		return err
	}
	if err := mc.requireFPSet().CommitChkpt(); err != nil {
		return err
	}
	if err := CommitChkptUniqueStrings(mc.ownedMetadir()); err != nil {
		return err
	}
	if mc.ownedCheckLiveness() {
		if err := mc.ownedLiveCheck().CommitChkpt(); err != nil {
			return err
		}
	}
	PrintMessage(ECTLCCheckpointEnd)
	return nil
}

func (mc *ModelChecker) Recover() (bool, error) {
	mc.require()
	if mc.ownedFromCheckpoint() == nil {
		return false, nil
	}
	PrintMessage(ECTLCCheckpointRecoverStart, *mc.ownedFromCheckpoint())
	// The flag enables recovery; each initialized store owns its checkpoint
	// location, including fingerprint sets supplied by the caller.
	if err := mc.RecoverTrace(); err != nil {
		return false, err
	}
	if err := mc.requireStateQueue().Recover(); err != nil {
		return false, err
	}
	if err := recoverFPSetFromCheckerTrace(mc.requireFPSet(), mc.ownedTrace(), mc.ownedConcurrentTrace()); err != nil {
		return false, err
	}
	if mc.ownedCheckLiveness() {
		err := mc.ownedTool().GetInitStates(NewStateFunctor(func(state *TLCStateMut) (any, error) {
			if err := mc.ownedLiveCheck().AddInitState(mc.ownedTool(), state, state.FingerPrint()); err != nil {
				return nil, err
			}
			return true, nil
		}))
		if err != nil {
			return false, err
		}
		if err := mc.ownedLiveCheck().Recover(); err != nil {
			return false, err
		}
	}
	fpSize := mc.requireFPSet().Size()
	queueSize := mc.requireStateQueue().Size()
	PrintMessage(ECTLCCheckpointRecoverEnd, fmtInt64(int64(fpSize)), fmtInt64(queueSize))
	atomic.StoreInt64(&mc.NumberOfInitialStates, int64(mc.requireFPSet().Size()))
	return true, nil
}

func recoverFPSetFromCheckerTrace(fpSet FPSet, trace *TLCTrace, concurrentTrace *ConcurrentTLCTrace) error {
	if isNil(fpSet) {
		panic(NewNullPointerException())
	}
	if concurrentTrace != nil {
		return fpSet.Recover(concurrentTrace)
	}
	return fpSet.Recover(trace)
}

func (mc *ModelChecker) requireStateQueue() StateQueue {
	mc.require()
	if isNil(mc.StateQueue) {
		panic(NewNullPointerException())
	}
	return mc.StateQueue
}

func (mc *ModelChecker) requireFPSet() FPSet {
	mc.require()
	if isNil(mc.FPSet) {
		panic(NewNullPointerException())
	}
	return mc.FPSet
}

func (mc *ModelChecker) requireStateWriter() IStateWriter {
	mc.require()
	if isNil(mc.ownedAllStateWriter()) {
		panic(NewNullPointerException())
	}
	return mc.ownedAllStateWriter()
}

func (mc *ModelChecker) Cleanup(success bool, cleanup bool) error {
	mc.require()
	vetoCleanup := modelCheckerVetoCleanup()
	if cleanup && CheckpointExplicitlyEnabled() && !mc.requireStateQueue().IsEmpty() && (mc.ErrState != nil || mc.TimeBound || mc.IsTimeBound()) {
		// Important Java compatibility: this branch is load-bearing.  Java TLC
		// snapshots an interrupted/error run with queued work and then vetoes
		// metadata deletion so users can recover and continue.  Do not "tidy"
		// the metadir here. This local checkpoint veto still closes resources;
		// only the explicit VETO_CLEANUP property retains the liveness graph.
		if checkpointErr := mc.Checkpoint(); checkpointErr != nil {
			return checkpointErr
		}
		vetoCleanup = true
	}
	mc.requireFPSet().Close()
	if mc.ownedConcurrentTrace() != nil {
		if closeErr := mc.ownedConcurrentTrace().Close(); closeErr != nil {
			return closeErr
		}
	} else if mc.ownedTrace() != nil {
		if closeErr := mc.ownedTrace().Close(); closeErr != nil {
			return closeErr
		}
		for _, worker := range mc.Workers {
			if worker != nil {
				if closeErr := worker.CloseTrace(); closeErr != nil {
					return closeErr
				}
			}
		}
	} else {
		panic(NewNullPointerException())
	}
	if mc.ownedCheckLiveness() {
		if closeErr := mc.ownedLiveCheck().Close(); closeErr != nil {
			return closeErr
		}
	}
	if closeErr := mc.requireStateWriter().Close(); closeErr != nil {
		return closeErr
	}
	if cleanup && !vetoCleanup {
		// Java's FileUtil.deleteDir(path, success) is deliberately recursive only
		// on success. Keep failure artifacts for inspection/recovery. Resources
		// close above, except graphs retained by the explicit VETO_CLEANUP property.
		deleteDirLikeJava(mc.ownedMetadir(), success)
	}
	return nil
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
	initializeModelCheckerStatics()
	return modelCheckerStatics.vetoCleanup
}

func (mc *ModelChecker) BeginTraceChkpt() error {
	mc.require()
	if mc.ownedConcurrentTrace() != nil {
		return mc.ownedConcurrentTrace().BeginChkpt()
	}
	for _, worker := range mc.Workers {
		if worker != nil {
			if err := worker.BeginChkpt(); err != nil {
				return err
			}
		}
	}
	if mc.ownedTrace() != nil {
		return mc.ownedTrace().BeginChkpt()
	}
	panic(NewNullPointerException())
}

func (mc *ModelChecker) CommitTraceChkpt() error {
	mc.require()
	if mc.ownedConcurrentTrace() != nil {
		return mc.ownedConcurrentTrace().CommitChkpt()
	}
	for _, worker := range mc.Workers {
		if worker != nil {
			if err := worker.CommitChkpt(); err != nil {
				return err
			}
		}
	}
	if mc.ownedTrace() != nil {
		return mc.ownedTrace().CommitChkpt()
	}
	panic(NewNullPointerException())
}

func (mc *ModelChecker) RecoverTrace() error {
	mc.require()
	if mc.ownedConcurrentTrace() != nil {
		return mc.ownedConcurrentTrace().Recover()
	}
	for _, worker := range mc.Workers {
		if worker != nil {
			if err := worker.Recover(); err != nil {
				return err
			}
		}
	}
	if mc.ownedTrace() != nil {
		return mc.ownedTrace().Recover()
	}
	panic(NewNullPointerException())
}

func (mc *ModelChecker) PrintSummary(success bool, startMillis ...int64) {
	if mc == nil {
		panic(NewNullPointerException())
	}
	if len(startMillis) > 1 {
		panic(NewIllegalArgumentException("PrintSummary accepts at most one start time"))
	}
	startTime := mc.ownedStartTime()
	if len(startMillis) == 1 {
		if startMillis[0] < 0 {
			startTime = time.Time{}
		} else {
			startTime = time.UnixMilli(startMillis[0])
		}
	}
	mc.AbstractChecker.ReportCoverage(mc.Workers)
	if toolMode() {
		mc.PrintProgressStats(startTime, true)
	}
	PrintMessage(ECTLCStats,
		fmtInt64(mc.GetStatesGenerated()),
		fmtInt64(int64(mc.GetDistinctStatesGenerated())),
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
	mc.require()
	if !mc.ownedCheckLiveness() || !livenessStatsEnabled() {
		return nil
	}
	runtime.GC()
	inDegree, err := mc.ownedLiveCheck().CalculateInDegreeDiskGraphs(
		NewBucketStatisticsWithMX("Histogram vertex in-degree", javaLivenessPackageName, "DiskGraphsInDegree"),
	)
	if err != nil {
		return err
	}
	printLivenessStatistics(inDegree, mc.ownedLiveCheck().GetOutDegreeStatistics())
	return nil
}

func livenessStatsEnabled() bool {
	initializeAbstractCheckerStatics()
	return abstractCheckerStatics.livenessStats
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
	mc.require()
	statesGenerated := mc.GetStatesGenerated()
	plural := ""
	if statesGenerated != 1 {
		plural = "s"
	}
	if uint64(statesGenerated) == mc.GetDistinctStatesGenerated() {
		PrintMessage(ECTLCInitGenerated1, fmtInt64(statesGenerated), plural)
		return
	}
	PrintMessage(ECTLCInitGenerated2, fmtInt64(statesGenerated), plural, fmtInt64(int64(mc.GetDistinctStatesGenerated())))
}

func (mc *ModelChecker) PrintProgressStats(startTime time.Time, isFinal bool) {
	_ = isFinal
	mc.require()
	fpSetSize := mc.requireFPSet().Size()
	factor := float64(ProgressInterval().Milliseconds()) / 60000
	if !startTime.IsZero() {
		mc.OldNumOfGenStates = 0
		mc.OldFPSetSize = 0
		factor = float64(time.Now().UnixMilli()-startTime.UnixMilli()) / 60000
	}
	generated := mc.GetStatesGenerated()
	atomic.StoreInt64(&mc.StatesPerMinute, javaDoubleToLong(float64(generated-mc.OldNumOfGenStates)/factor))
	mc.OldNumOfGenStates = generated
	distinctDelta := int64(fpSetSize) - int64(mc.OldFPSetSize)
	atomic.StoreInt64(&mc.DistinctStatesPerMinute, javaDoubleToLong(float64(distinctDelta)/factor))
	mc.OldFPSetSize = fpSetSize
	PrintMessage(ECTLCProgressStats,
		fmtInt64(mc.GetProgress()),
		MessageNumberFormat(generated),
		MessageNumberFormat(int64(fpSetSize)),
		MessageNumberFormat(mc.GetStateQueueSize()),
		MessageNumberFormat(atomic.LoadInt64(&mc.StatesPerMinute)),
		MessageNumberFormat(atomic.LoadInt64(&mc.DistinctStatesPerMinute)),
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

func javaMathRound(value float64) int64 {
	return javaDoubleToLong(math.Floor(value + 0.5))
}

func (mc *ModelChecker) PrintOutDegreeSummary() {
	mc.require()
	agg := NewBucketStatistics("State Graph OutDegree")
	for _, worker := range mc.Workers {
		if worker == nil || worker.ownedOutDegree() == nil {
			panic(NewNullPointerException())
		}
		agg.Add(worker.ownedOutDegree())
	}
	if agg.Observations() == 0 {
		return
	}
	PrintMessage(ECTLCStateGraphOutdegree,
		fmtInt(agg.Min()),
		fmtInt64(javaMathRound(agg.Mean())),
		fmtInt64(javaMathRound(agg.Percentile(.95))),
		fmtInt(agg.Max()),
	)
}

func toolMode() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.Tool
}

// StartWorkers publishes the worker count to the fingerprint set and starts
// the breadth-first workers. Java's checker and checkIndex parameters are not
// used by ModelChecker.startWorkers.
func (mc *ModelChecker) StartWorkers(_ *AbstractChecker, _ int) []*Worker {
	mc.require()
	mc.initWorkers()
	mc.requireFPSet().IncWorkers(len(mc.Workers))
	for _, worker := range mc.Workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		worker.Start()
	}
	return mc.Workers
}

func (mc *ModelChecker) RunTLC(maxDepth int) (int, error) {
	if mc == nil {
		panic(NewNullPointerException())
	}
	maxDepth = int(int32(maxDepth))
	if maxDepth < 2 {
		return NoError, nil
	}
	mc.StartWorkers(mc.AbstractChecker, 0)
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

func (mc *ModelChecker) waitForWorkersWithPeriodicWork(maxDepth int) (result int, failure error) {
	defer mc.endCoordinatorWait()
	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- mc.joinWorkers()
	}()
	defer func() {
		if recovered := recover(); recovered != nil {
			mc.Stop()
			<-done
			result, failure = ECGeneral, panicValueAsError(recovered)
		}
	}()
	interval := ProgressInterval()
	coverageCountdown := periodicCoverageCountdown(interval)
	wakeup := mc.beginCoordinatorWait()
	select {
	case err := <-done:
		result, periodicErr := mc.DoPeriodicWork()
		if periodicErr != nil || result != NoError {
			return result, periodicErr
		}
		return NoError, err
	case <-time.After(3 * time.Second):
	case <-wakeup:
	}
	mc.endCoordinatorWait()
	for {
		result, err := mc.DoPeriodicWork()
		if err != nil || result != NoError {
			mc.requireStateQueue().FinishAll()
			joinErr := <-done
			if err != nil {
				return result, err
			}
			return result, joinErr
		}
		// Java holds the checker monitor across the done test, progress report,
		// depth check, and coverage report. Its wait then releases that monitor.
		shouldWait, completed := func() (bool, bool) {
			mc.mu.Lock()
			defer mc.mu.Unlock()
			if mc.isModelCheckerDone() {
				return false, true
			}
			// Register before reporting progress so a control notification cannot
			// be lost between continuation and the timed wait.
			wakeup = mc.beginCoordinatorWait()
			shouldWait := mc.RunTLCContinueDoing(coverageCountdown, maxDepth)
			if shouldWait && interval < 0 {
				panic(NewIllegalArgumentException("timeout value is negative"))
			}
			return shouldWait, false
		}()
		if completed {
			return NoError, <-done
		}
		var joinErr error
		workersFinished := false
		if shouldWait && !mc.isModelCheckerDone() {
			timer := time.NewTimer(interval)
			select {
			case joinErr = <-done:
				workersFinished = true
			case <-timer.C:
			case <-wakeup:
			}
			timer.Stop()
		}
		completed = func() bool {
			mc.mu.Lock()
			defer mc.mu.Unlock()
			mc.endCoordinatorWait()
			if coverageCountdown == 0 {
				coverageCountdown = periodicCoverageCountdown(interval)
			} else {
				coverageCountdown = int(int32(coverageCountdown) - 1)
			}
			return mc.isModelCheckerDone()
		}()
		if workersFinished {
			return NoError, joinErr
		}
		if completed {
			return NoError, <-done
		}
	}
}

func (mc *ModelChecker) isModelCheckerDone() bool {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	return mc.Done.Load()
}

// RunTLCContinueDoing reports progress, applies the depth bound and emits
// periodic coverage. The return value tells the native coordinator whether the
// Java monitor wait that follows this method is still required.
func (mc *ModelChecker) RunTLCContinueDoing(count int, depth int) bool {
	if mc == nil {
		panic(NewNullPointerException())
	}
	count = int(int32(count))
	depth = int(int32(depth))
	trace := mc.ownedConcurrentTrace()
	if trace == nil {
		panic(NewNullPointerException())
	}
	level := trace.GetLevel()
	mc.PrintProgressStats(time.Time{}, false)
	if level > depth {
		mc.requireStateQueue().FinishAll()
		mc.SetDone()
		return false
	}
	if count == 0 {
		mc.reportPeriodicCoverage()
	}
	return true
}

func periodicCoverageCountdown(interval time.Duration) int {
	Globals.Lock()
	coverageMillis := Globals.CoverageInterval
	Globals.Unlock()
	progressMillis := int32(interval / time.Millisecond)
	if progressMillis == 0 {
		panic(NewArithmeticException("/ by zero"))
	}
	return int(int32(coverageMillis) / progressMillis)
}

func (mc *ModelChecker) reportPeriodicCoverage() {
	mc.require()
	if mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	mc.AbstractChecker.ReportCoverage(mc.Workers)
}

func (mc *ModelChecker) joinWorkers() error {
	mc.require()
	var joinErr error
	for _, worker := range mc.Workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if err := worker.Join(); err != nil && joinErr == nil {
			joinErr = err
		}
	}
	return joinErr
}

func (mc *ModelChecker) DoPeriodicWork() (int, error) {
	mc.require()
	createCheckpoint := DoCheckPoint()
	if mc.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	tool := mc.ownedTool()
	periodic := tool.GetPeriodic()
	liveCheckDue := mc.ownedCheckLiveness() && !(mc.GetRuntimeRatio() > LivenessRatio()) && mc.ownedLiveCheck().DoLiveCheck()
	if !liveCheckDue && !mc.forceLiveCheck.Load() && !createCheckpoint && periodic == nil {
		mc.UpdateRuntimeRatio(0)
		return NoError, nil
	}
	if !mc.requireStateQueue().SuspendAll() {
		return NoError, nil
	}
	if mc.ownedCheckLiveness() && (mc.GetRuntimeRatio() < LivenessRatio() || mc.forceLiveCheck.Load()) {
		startMillis := time.Now().UnixMilli()
		result, err := mc.ownedLiveCheck().Check(tool.NoDebug(), mc.forceLiveCheck.Load())
		if err != nil || result != NoError {
			return result, err
		}
		mc.forceLiveCheck.Store(false)
		mc.updateRuntimeRatioMillis(time.Now().UnixMilli() - startMillis)
	} else if mc.GetRuntimeRatio() > LivenessRatio() {
		mc.UpdateRuntimeRatio(0)
	}
	if periodic != nil {
		value, err := tool.NoDebug().Eval(periodic)
		if err != nil {
			return ECGeneral, err
		}
		isFalse, err := BoolFalse.Equal(value)
		if err != nil {
			return ECGeneral, err
		}
		if isFalse {
			return ECTLCAssumptionFalse, nil
		}
	}
	if createCheckpoint {
		if err := mc.Checkpoint(); err != nil {
			return ECSystemCheckpointRecoveryCorrupt, err
		}
	} else {
		mc.requireStateQueue().ResumeAll()
	}
	return NoError, nil
}

func (mc *ModelChecker) UpdateRuntimeRatio(delta time.Duration) {
	if mc == nil {
		panic(NewNullPointerException())
	}
	if delta < 0 {
		panic(NewAssertionError())
	}
	mc.updateRuntimeRatioMillis(delta.Milliseconds())
}

func (mc *ModelChecker) updateRuntimeRatioMillis(deltaMillis int64) {
	mc.require()
	if deltaMillis < 0 {
		panic(NewAssertionError())
	}
	totalRuntimeMillis := time.Now().UnixMilli() - mc.ownedStartTime().UnixMilli()
	totalRuntimeMillis -= ProgressInterval().Milliseconds()
	totalRuntimeMillis -= deltaMillis
	absLivenessRuntime := float64(totalRuntimeMillis) * mc.GetRuntimeRatio()
	if absLivenessRuntime < 0 {
		absLivenessRuntime = 0
	}
	denominator := float64(totalRuntimeMillis + ProgressInterval().Milliseconds() + deltaMillis)
	mc.runtimeRatio.Store(math.Float64bits((float64(deltaMillis) + absLivenessRuntime) / denominator))
}

func (mc *ModelChecker) GetRuntimeRatio() float64 {
	if mc == nil {
		panic(NewNullPointerException())
	}
	return math.Float64frombits(mc.runtimeRatio.Load())
}

func (mc *ModelChecker) ForceLiveCheck() {
	if mc == nil {
		panic(NewNullPointerException())
	}
	mc.forceLiveCheck.Store(true)
}

func (mc *ModelChecker) ForceLivenessCheck() { mc.ForceLiveCheck() }

func (mc *ModelChecker) DoInit(ignoreCancel bool) (int, error) {
	mc.require()
	return mc.doInitWithTool(mc.ownedTool(), ignoreCancel)
}

func (mc *ModelChecker) doInitWithTool(tool *Tool, ignoreCancel bool) (result int, err error) {
	mc.require()
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
			result = initExceptionCode(err)
		}
	}()
	if tool == nil {
		panic(NewNullPointerException())
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
		return failure.ownedCode()
	}
	return ECGeneral
}

func (mc *ModelChecker) replayInitErrorCallStack(fallback int) int {
	mc.require()
	callStackTool := NewCallStackTool(mc.ownedTool())
	atomic.StoreInt64(&mc.NumberOfInitialStates, 0)
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
	mc.require()
	return mc.doNextWithTool(mc.ownedTool(), curState, nil, nil)
}

func (mc *ModelChecker) replayNextErrorCallStack() int {
	mc.require()
	callStackTool := NewCallStackTool(mc.ownedTool())
	var liveNextStates *SetOfStates
	if mc.ownedCheckLiveness() {
		liveNextStates = NewSetOfStates()
	}
	replayWorker := NewWorker(4223)
	replayWorker.Checker = mc
	replayWorker.Tool = mc.ownedTool()
	replayWorker.captureRunSettings()
	replayWorker.actionCoverage = mc.actionCoverage
	replayWorker.DisableTraceMirror = true
	rootName := "Spec"
	if mc.ownedTool().GetRootName() != "" {
		rootName = mc.ownedTool().GetRootName()
	}
	replayWorker.SetTraceContext(mc.ownedMetadir(), rootName)
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
			return err.(*TLCError).ownedCode()
		}
		PrintError(ECTLCNestedExpression, callStackTool.CallStackString())
		return ECTLCNestedExpression
	}
	return NoError
}

func (mc *ModelChecker) doNextWithTool(tool *Tool, curState *TLCStateMut, liveNextStates *SetOfStates, worker *Worker) (stop bool, err error) {
	mc.require()
	var succState *TLCStateMut
	defer func() {
		if failure := recover(); failure != nil {
			stop, err = true, panicValueAsError(failure)
			mc.doNextFailed(curState, succState, err)
		}
	}()
	if tool == nil {
		err := NewNullPointerException()
		mc.doNextFailed(curState, nil, err)
		return true, err
	}
	restoreRandomState := PushRandomEnumerableState(curState)
	defer restoreRandomState()
	restoreCurrentState := PushCurrentState(curState)
	defer restoreCurrentState()
	deadLocked := true
	for _, action := range tool.GetActions() {
		nextStates, err := tool.GetNextStates(action, curState)
		if err != nil {
			mc.doNextFailed(curState, succState, err)
			return true, err
		}
		if nextStates == nil {
			err := NewNullPointerException()
			mc.doNextFailed(curState, succState, err)
			return true, err
		}
		size := nextStates.Size()
		if worker != nil {
			worker.IncrementStatesGenerated(int64(size))
		} else {
			atomic.AddInt64(&mc.NextStatesGenerated, int64(size))
		}
		deadLocked = deadLocked && size == 0
		for i := 0; i < size; i++ {
			succState = nextStates.At(i)
			if !tool.IsGoodState(succState) {
				return mc.doNextSetErrIncomplete(curState, succState, action), nil
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
			if inModel && unseen {
				mc.requireStateQueue().SEnqueue(succState)
			}
		}
		succState = nil
	}
	if deadLocked && mc.ownedCheckDeadlock() {
		return mc.doNextSetErr(curState, nil, false, ECTLCDeadlockReached, ""), nil
	}
	return false, nil
}

func workerIDForReplayWorker(worker *Worker) int {
	if worker == nil {
		return 0
	}
	return worker.ownedID()
}

func (mc *ModelChecker) processSuccessorForWorker(worker *Worker, curState *TLCStateMut, succState *TLCStateMut, action *Action, collectedStates *SetOfStates) (bool, bool, error) {
	tool := worker.ownedWorkerTool()
	if tool == nil {
		panic(NewNullPointerException())
	}
	if !tool.IsGoodState(succState) {
		return mc.doNextSetErrIncompleteWithPostConditionTool(tool, curState, succState, action), false, nil
	}
	succState.SetPredecessor(curState).SetAction(action)
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
		seen, err := mc.isSeenStateUsingWorker(worker.ownedID(), worker, curState, succState, action, collectedStates)
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
		if worker.variableCoverage {
			countStateVariableCoverage(succState)
		}
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
	if mc == nil || mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	total := atomic.LoadInt64(&mc.NumberOfInitialStates) + atomic.LoadInt64(&mc.NextStatesGenerated)
	for _, worker := range mc.Workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		total += worker.GetStatesGenerated()
	}
	return total
}

func (mc *ModelChecker) GetInitialStatesGenerated() int64 {
	if mc == nil || mc.AbstractChecker == nil {
		panic(NewNullPointerException())
	}
	return atomic.LoadInt64(&mc.NumberOfInitialStates)
}

func (mc *ModelChecker) GetStateQueueSize() int64 {
	if mc == nil || mc.AbstractChecker == nil || isNil(mc.StateQueue) {
		panic(NewNullPointerException())
	}
	return mc.StateQueue.Size()
}

func (mc *ModelChecker) GetDistinctStatesGenerated() uint64 {
	if mc == nil || mc.AbstractChecker == nil || isNil(mc.FPSet) {
		panic(NewNullPointerException())
	}
	return mc.FPSet.Size()
}

func (mc *ModelChecker) isSeenStateUsingWorker(workerID int, worker *Worker, curState *TLCStateMut, succState *TLCStateMut, action *Action, collectedStates *SetOfStates) (bool, error) {
	tool := mc.ownedTool()
	if worker != nil {
		tool = worker.ownedWorkerTool()
		if tool == nil {
			panic(NewNullPointerException())
		}
	}
	fp := succState.FingerPrintWithTool(tool)
	fingerprintSet, writer := mc.FPSet, mc.ownedAllStateWriter()
	if worker != nil {
		fingerprintSet, writer = worker.fingerprintSet, worker.stateWriter
	}
	if isNil(fingerprintSet) {
		panic(NewNullPointerException())
	}
	seen := fingerprintSet.Put(fp)
	if isNil(writer) {
		panic(NewNullPointerException())
	}
	status := StateVisitUnseen
	if seen {
		status = StateVisitSeen
	}
	if err := writer.WriteTransition(curState, succState, status, action); err != nil {
		return seen, err
	}
	if !seen {
		if worker != nil {
			if err := worker.WriteNextState(curState, fp, succState, action); err != nil {
				return seen, err
			}
		} else {
			if mc.ownedTrace() == nil {
				panic(NewNullPointerException())
			}
			if err := mc.ownedTrace().WriteNextStateForWorker(workerID, curState, fp, succState, action); err != nil {
				return seen, err
			}
		}
		actionCoverage := mc.actionCoverage
		if worker != nil {
			actionCoverage = worker.actionCoverage
		}
		if actionCoverage {
			if action == nil {
				panic(NewNullPointerException())
			}
			action.CM.IncSecondary()
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
	return mc.doNextCheckInvariantsWithTool(mc.ownedTool(), curState, succState, true)
}

func (mc *ModelChecker) doNextCheckInvariantsWithTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, withPostCondition bool) (stop bool, failure error) {
	i := 0
	defer func() {
		if recovered := recover(); recovered != nil {
			err, ok := recovered.(error)
			if !ok || isJavaError(err) {
				panic(recovered)
			}
			stop, failure = true, mc.doNextEvalFailed(curState, succState, ECTLCInvariantEvaluationFailed, tool.propertyNameAt(tool.GetInvNames(), i), err)
		}
	}()
	for ; i < len(tool.requireActionArray(tool.GetInvariants())); i++ {
		invariant := tool.requireActionArray(tool.GetInvariants())[i]
		valid, err := tool.IsValidState(invariant, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCInvariantEvaluationFailed, tool.propertyNameAt(tool.GetInvNames(), i), err)
		}
		if !valid {
			if continuationEnabled() {
				mc.printContinuationViolation(curState, succState, ECTLCInvariantViolatedBehavior, tool.propertyNameAt(tool.GetInvNames(), i))
				return false, nil
			}
			if withPostCondition {
				return mc.doNextSetErrParamsWithPostConditionTool(tool, curState, succState, false, ECTLCInvariantViolatedBehavior, tool.propertyNameAt(tool.GetInvNames(), i)), nil
			}
			return mc.doNextSetErrParams(curState, succState, false, ECTLCInvariantViolatedBehavior, tool.propertyNameAt(tool.GetInvNames(), i)), nil
		}
	}
	return false, nil
}

func (mc *ModelChecker) doNextCheckImplied(curState *TLCStateMut, succState *TLCStateMut) (bool, error) {
	return mc.doNextCheckImpliedWithTool(mc.ownedTool(), curState, succState, true)
}

func (mc *ModelChecker) doNextCheckImpliedWithTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, withPostCondition bool) (stop bool, failure error) {
	i := 0
	defer func() {
		if recovered := recover(); recovered != nil {
			err, ok := recovered.(error)
			if !ok || isJavaError(err) {
				panic(recovered)
			}
			stop, failure = true, mc.doNextEvalFailed(curState, succState, ECTLCActionPropertyEvaluationFailed, tool.propertyNameAt(tool.GetImpliedActNames(), i), err)
		}
	}()
	for ; i < len(tool.requireActionArray(tool.GetImpliedActions())); i++ {
		action := tool.requireActionArray(tool.GetImpliedActions())[i]
		valid, err := tool.IsValidTransition(action, curState, succState)
		if err != nil {
			return true, mc.doNextEvalFailed(curState, succState, ECTLCActionPropertyEvaluationFailed, tool.propertyNameAt(tool.GetImpliedActNames(), i), err)
		}
		if !valid {
			if continuationEnabled() {
				mc.printContinuationViolation(curState, succState, ECTLCActionPropertyViolatedBehavior, tool.propertyNameAt(tool.GetImpliedActNames(), i))
				return false, nil
			}
			if withPostCondition {
				return mc.doNextSetErrParamsWithPostConditionTool(tool, curState, succState, false, ECTLCActionPropertyViolatedBehavior, tool.propertyNameAt(tool.GetImpliedActNames(), i)), nil
			}
			return mc.doNextSetErrParams(curState, succState, false, ECTLCActionPropertyViolatedBehavior, tool.propertyNameAt(tool.GetImpliedActNames(), i)), nil
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
	mc.mu.Lock()
	defer mc.mu.Unlock()
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
		mc.requireStateQueue().FinishAll()
		mc.notifyCoordinatorLocked()
	}
	return true
}

func (mc *ModelChecker) doNextSetErrIncomplete(curState *TLCStateMut, succState *TLCStateMut, action *Action) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.doNextSetErrIncompleteLocked(curState, succState, action)
}

func (mc *ModelChecker) doNextSetErrIncompleteLocked(curState *TLCStateMut, succState *TLCStateMut, action *Action) bool {
	if mc.SetErrState(curState, succState, false, ECTLCStateNotCompletelySpecifiedNext) {
		PrintError(ECTLCStateNotCompletelySpecifiedNext, incompleteNextStateParams(mc.ownedTool(), action, succState)...)
		mc.printBehaviorTrace(curState, succState)
		mc.requireStateQueue().FinishAll()
		mc.notifyCoordinatorLocked()
	}
	return true
}

func (mc *ModelChecker) doNextSetErrIncompleteWithPostConditionTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, action *Action) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	isConsole := !mc.isModelCheckerDone()
	result := mc.doNextSetErrIncompleteLocked(curState, succState, action)
	mc.checkPostConditionWithErrorTraceTool(tool, curState, succState, isConsole)
	return result
}

func (mc *ModelChecker) doNextSetErrWithPostCondition(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, param string) bool {
	return mc.doNextSetErrWithPostConditionTool(mc.ownedTool(), curState, succState, keep, ec, param)
}

func (mc *ModelChecker) doNextSetErrWithPostConditionTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, param string) bool {
	if param == "" {
		return mc.doNextSetErrParamsWithPostConditionTool(tool, curState, succState, keep, ec)
	}
	return mc.doNextSetErrParamsWithPostConditionTool(tool, curState, succState, keep, ec, param)
}

func (mc *ModelChecker) doNextSetErrParamsWithPostCondition(curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, params ...string) bool {
	return mc.doNextSetErrParamsWithPostConditionTool(mc.ownedTool(), curState, succState, keep, ec, params...)
}

func (mc *ModelChecker) doNextSetErrParamsWithPostConditionTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, keep bool, ec int, params ...string) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	isConsole := !mc.isModelCheckerDone()
	result := mc.doNextSetErrParamsLocked(curState, succState, keep, ec, params...)
	mc.checkPostConditionWithErrorTraceTool(tool, curState, succState, isConsole)
	return result
}

func (mc *ModelChecker) printContinuationViolation(curState *TLCStateMut, succState *TLCStateMut, ec int, param string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	PrintError(ec, param)
	mc.printBehaviorTrace(curState, succState)
}

func (mc *ModelChecker) checkPostConditionAfterInitFailure() {
	mc.require()
	if mc.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	tool := mc.ownedTool()
	if mc.ErrState != nil {
		tool.CheckPostConditionWithCounterExample(NewCounterExampleFromInitialState(mc.ErrState))
		return
	}
	tool.CheckPostCondition()
}

func (mc *ModelChecker) checkPostConditionWithErrorTrace(curState *TLCStateMut, succState *TLCStateMut, isConsole bool) {
	mc.checkPostConditionWithErrorTraceTool(mc.ownedTool(), curState, succState, isConsole)
}

func (mc *ModelChecker) checkPostConditionWithErrorTraceTool(tool *Tool, curState *TLCStateMut, succState *TLCStateMut, isConsole bool) {
	trace := mc.errorTraceInfo(tool, curState, succState)
	trace = aliasTraceWithToolPairs(tool, trace)
	tool.CheckPostConditionWithCounterExample(NewCounterExampleFromTraceConsole(trace, isConsole))
}

func (mc *ModelChecker) printBehaviorTrace(curState *TLCStateMut, succState *TLCStateMut) {
	mc.require()
	if mc.ownedConcurrentTrace() != nil {
		mc.ownedConcurrentTrace().PrintTrace(curState, succState)
	} else if mc.ownedTrace() != nil {
		mc.ownedTrace().PrintTrace(curState, succState)
	} else {
		panic(NewNullPointerException())
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
	return trace[len(trace)-1].ownedState()
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

func (mc *ModelChecker) traceInfoPrefix(state *TLCStateMut) []*TLCStateInfo {
	if mc == nil {
		panic(NewNullPointerException())
	}
	if mc.ownedConcurrentTrace() != nil {
		return mc.ownedConcurrentTrace().GetTrace(state)
	}
	if mc.ownedTrace() != nil {
		return mc.ownedTrace().GetTrace(state)
	}
	panic(NewNullPointerException())
}

func aliasTraceWithTool(tool *Tool, trace []*TLCStateInfo) []*TLCStateInfo {
	if tool == nil || len(trace) == 0 {
		return trace
	}
	aliased := append([]*TLCStateInfo(nil), trace...)
	for i, current := range aliased {
		successor := current.GetOriginalState()
		if i+1 < len(aliased) {
			successor = aliased[i+1].GetOriginalState()
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
		successor := next.GetOriginalState()
		alias, err := tool.EvalAliasInfoPair(current, successor)
		if err != nil {
			panic(err)
		}
		aliased[i] = alias
	}
	return aliased
}

func (mc *ModelChecker) doNextEvalFailed(curState *TLCStateMut, succState *TLCStateMut, ec int, param string, err error) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.SetErrState(curState, succState, true, ec) {
		msg := javaThrowableMessage(err)
		PrintError(ec, param, msg)
		mc.printBehaviorTrace(curState, succState)
		mc.requireStateQueue().FinishAll()
		mc.notifyCoordinatorLocked()
	}
	return err
}

func (mc *ModelChecker) doNextFailed(curState *TLCStateMut, succState *TLCStateMut, err error) {
	if err == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	ec, keepCallStack := doNextFailureCode(err)
	if mc.SetErrState(curState, succState, keepCallStack, ec) {
		_, params, _ := doNextFailureMessage(err)
		if params != nil {
			if isDoNextSystemOrAssertionError(err) {
				PrintErrorNullable(ec, javaThrowableDetailMessage(err))
			} else if eval, ok := err.(*EvalException); ok && eval.GetNullableParameters() != nil {
				PrintErrorNullable(ec, eval.GetNullableParameters()...)
			} else if failure, ok := err.(*TLCError); ok && !failure.Runtime && failure.NullableParams != nil {
				PrintErrorNullable(ec, failure.NullableParams...)
			} else {
				PrintError(ec, params...)
			}
		}
		mc.printBehaviorTrace(curState, succState)
		mc.requireStateQueue().FinishAll()
		mc.notifyCoordinatorLocked()
	}
}

func doNextFailureMessage(err error) (int, []string, bool) {
	ec, keepCallStack := doNextFailureCode(err)
	// Java checks the throwable message only after winning setErrState.
	if ec == ECGeneral && javaThrowableDetailMessage(err) == nil && javaSystemFailureCode(err) == NoError {
		return ec, nil, keepCallStack
	}
	// Java tests the thrown exception itself, not its nested causes.
	switch err.(type) {
	case *StackOverflowError:
		return ec, []string{}, keepCallStack
	case *OutOfMemoryError:
		return ec, []string{}, keepCallStack
	case *AssertionError:
		return ec, []string{}, keepCallStack
	}
	params := []string{err.Error()}

	if eval, ok := err.(*EvalException); ok && eval != nil {
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
		if tlcErr.ownedCode() != ECGeneral {
			if tlcErr.Params == nil {
				return tlcErr.ownedCode(), []string{}, keepCallStack
			}
			return tlcErr.ownedCode(), tlcErr.Params, keepCallStack
		}
	}

	return ECGeneral, generalErrorParams("", err), keepCallStack
}

func doNextFailureCode(err error) (int, bool) {
	// This is the source's pre-claim instanceof selection. Message and parameter
	// inspection belongs to the winning reporter after setErrState.
	switch err.(type) {
	case *StackOverflowError:
		return ECSystemStackOverflow, false
	case *OutOfMemoryError:
		return ECSystemOutOfMemory, false
	case *AssertionError:
		return ECTLCBug, false
	}
	if eval, ok := err.(*EvalException); ok && eval != nil {
		return eval.GetErrorCode(), true
	}
	if failure, ok := err.(*TLCError); ok && failure != nil && !failure.Runtime {
		keep := true
		if code := javaSystemFailureCode(err); code == ECSystemStackOverflow || code == ECSystemOutOfMemory || code == ECTLCBug {
			keep = false
		}
		return failure.ownedCode(), keep
	}
	return ECGeneral, true
}

func isDoNextSystemOrAssertionError(err error) bool {
	switch err.(type) {
	case *StackOverflowError, *OutOfMemoryError, *AssertionError:
		return true
	default:
		return false
	}
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
	if isPowerOfTwo(atomic.LoadInt64(&f.mc.NumberOfInitialStates)) && atomic.LoadInt64(&f.mc.NumberOfInitialStates) > 1 {
		PrintMessage(ECTLCComputingInitProgress, fmt.Sprintf("%d", atomic.LoadInt64(&f.mc.NumberOfInitialStates)))
	}
	atomic.AddInt64(&f.mc.NumberOfInitialStates, 1)
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
		return f.handleInitError(curState, NewDoInitInvariantViolatedException())
	}
	inModel, err := f.tool.IsInModel(curState)
	if err != nil {
		return f.handleInitError(curState, err)
	}
	seen := false
	if inModel {
		fp := curState.FingerPrint()
		seen = f.mc.requireFPSet().Put(fp)
		if !seen {
			if err := f.mc.requireStateWriter().WriteInitState(curState); err != nil {
				return f.handleInitError(curState, err)
			}
			if worker := f.mc.workerAt(0); worker != nil {
				if err := worker.WriteInitState(curState, fp); err != nil {
					return f.handleInitError(curState, err)
				}
			} else {
				if f.mc.ownedTrace() == nil {
					panic(NewNullPointerException())
				}
				if err := f.mc.ownedTrace().WriteInitState(curState, fp); err != nil {
					return f.handleInitError(curState, err)
				}
			}
			f.mc.requireStateQueue().Enqueue(curState)
			if f.mc.ownedCheckLiveness() {
				if err := f.mc.ownedLiveCheck().AddInitState(f.tool.NoDebug(), curState, fp); err != nil {
					return f.handleInitError(curState, err)
				}
			}
		}
	}
	if !seen || f.forceChecks {
		for i := 0; i < len(f.tool.requireActionArray(f.tool.GetInvariants())); i++ {
			invariant := f.tool.requireActionArray(f.tool.GetInvariants())[i]
			valid, err := f.tool.IsValidState(invariant, curState)
			if err != nil {
				return f.handleInitError(curState, err)
			}
			if !valid {
				alias := curState
				if f.tool != nil {
					alias = f.tool.EvalAlias(curState, curState)
				}
				PrintError(ECTLCInvariantViolatedInitial, f.tool.propertyNameAt(f.tool.GetInvNames(), i), alias.String())
				if !continuationEnabled() {
					f.errState = curState
					f.returnValue = ECTLCInvariantViolatedInitial
					return f.handleInitError(curState, NewDoInitInvariantViolatedException())
				}
			}
		}
		for i := 0; i < len(f.tool.requireActionArray(f.tool.GetImpliedInits())); i++ {
			implied := f.tool.requireActionArray(f.tool.GetImpliedInits())[i]
			valid, err := f.tool.IsValidState(implied, curState)
			if err != nil {
				return f.handleInitError(curState, err)
			}
			if !valid {
				alias := curState
				if f.tool != nil {
					alias = f.tool.EvalAlias(curState, curState)
				}
				PrintError(ECTLCPropertyViolatedInitial, f.tool.propertyNameAt(f.tool.GetImpliedInitNames(), i), alias.String())
				f.errState = curState
				f.returnValue = ECTLCPropertyViolatedInitial
				return f.handleInitError(curState, NewDoInitInvariantViolatedException())
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
	return action.nameForJavaFormat()
}
