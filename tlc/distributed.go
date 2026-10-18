package tlc

import (
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TLCServerName             = "TLCServer"
	TLCServerWorkerName       = TLCServerName + "WORKER"
	TLCWorkerThreadNamePrefix = "TLCWorkerThread-"
	TLCServerThreadNamePrefix = TLCWorkerThreadNamePrefix
	TLCServerDefaultPort      = 10997
	tlcServerVetoCleanup      = "tlc2.tool.distributed.TLCServer.vetoCleanup"
)

type NextStateResult struct {
	ComputationTime  int64
	StatesComputed   int64
	NextStates       []*StateVec
	NextFingerprints []*LongVec

	ownersCaptured        bool
	computationTimeOwner  int64
	statesComputedOwner   int64
	nextStatesOwner       []*StateVec
	nextFingerprintsOwner []*LongVec
}

func NewNextStateResult(nextStates []*StateVec, nextFingerprints []*LongVec, computationTime int64, statesComputed int64) *NextStateResult {
	return &NextStateResult{
		NextStates:       nextStates,
		NextFingerprints: nextFingerprints,
		ComputationTime:  computationTime,
		StatesComputed:   statesComputed,
		ownersCaptured:   true, computationTimeOwner: computationTime,
		statesComputedOwner: statesComputed, nextStatesOwner: nextStates,
		nextFingerprintsOwner: nextFingerprints,
	}
}

func (r *NextStateResult) ownedComputationTime() int64 {
	if r.ownersCaptured {
		return r.computationTimeOwner
	}
	return r.ComputationTime
}

func (r *NextStateResult) ownedStatesComputed() int64 {
	if r.ownersCaptured {
		return r.statesComputedOwner
	}
	return r.StatesComputed
}

func (r *NextStateResult) ownedNextStates() []*StateVec {
	if r.ownersCaptured {
		return r.nextStatesOwner
	}
	return r.NextStates
}

func (r *NextStateResult) ownedNextFingerprints() []*LongVec {
	if r.ownersCaptured {
		return r.nextFingerprintsOwner
	}
	return r.NextFingerprints
}

func (r *NextStateResult) GetStatesComputedDelta() int64 {
	if r == nil || r.ownedNextStates() == nil {
		panic(NewNullPointerException())
	}
	return r.ownedStatesComputed() - int64(len(r.ownedNextStates()))
}

func (r *NextStateResult) GetComputationTime() int64 {
	if r == nil {
		panic(NewNullPointerException())
	}
	return r.ownedComputationTime()
}

func (r *NextStateResult) GetNextFingerprints() []*LongVec {
	if r == nil {
		panic(NewNullPointerException())
	}
	return r.ownedNextFingerprints()
}

func (r *NextStateResult) GetNextStates() []*StateVec {
	if r == nil {
		panic(NewNullPointerException())
	}
	return r.ownedNextStates()
}

type DistributedWorker struct {
	nextStatesMu          sync.Mutex
	ID                    int
	App                   *TLCApp
	unsorted              bool
	FPSetManager          *DistributedFPSetManager
	Cache                 *SimpleCache
	cacheCaptured         bool
	cacheOwner            *SimpleCache
	uri                   *distributedWorkerURIValue
	networkReference      *DistributedEndpointReference
	Runtime               *DistributedWorkerRuntime
	unexported            atomic.Bool
	Computing             atomic.Bool
	LastInvocation        atomic.Int64
	OverallStatesComputed atomic.Int64
	CheckStateFunc        func(predecessor *TLCStateMut, successor *TLCStateMut) error
	IsInModelFunc         func(state *TLCStateMut) (bool, error)
	IsInActionsFunc       func(predecessor *TLCStateMut, successor *TLCStateMut) (bool, error)
	NetworkOverhead       float64
}

type TLCServer struct {
	publication                 TLCServerPublication
	unexported                  atomic.Bool
	checkDeadlock               *bool
	internMu                    sync.Mutex
	InternTable                 *InternTable
	Files                       *DistributedServerFiles
	FPSetManager                *DistributedFPSetManager
	StateQueue                  StateQueue
	Trace                       *TLCTrace
	Tool                        *Tool
	app                         *TLCApp
	Metadir                     string
	checkpointName              *string
	fpRegistration              *distributedFPRegistration
	FileName                    string
	ConfigName                  string
	Done                        atomic.Bool
	ErrState                    *TLCStateMut
	PredErrState                *TLCStateMut
	KeepCallStack               bool
	ErrorCode                   int
	LastError                   error
	WorkerStatesGenerated       atomic.Int64
	StatesPerMinute             int64
	DistinctStatesPerMinute     int64
	monitor                     distributedServerMonitor
	completionWaiter            chan struct{}
	threadsMu                   sync.Mutex
	threadsToWorkers            *InsMap[*TLCServerThread, DistributedWorkerEndpoint]
	executor                    DistributedExecutor
	BlockSelector               BlockSelection
	FinalNumberOfDistinctStates int64
	componentsCaptured          bool
	fpSetManagerOwner           *DistributedFPSetManager
	stateQueueOwner             StateQueue
	traceOwner                  *TLCTrace
	metadataCaptured            bool
	metadirOwner                string
	fileNameOwner               string
	configNameOwner             string
	selectorCaptured            bool
	blockSelectorOwner          BlockSelection
}

func NewTLCServer(fileName string, configName string, metadir string, manager *DistributedFPSetManager, queue StateQueue, trace *TLCTrace) *TLCServer {
	server := newTLCServer(fileName, configName, metadir, manager, queue, trace)
	server.BlockSelector = NewBlockSelectorFromProperties(server)
	server.blockSelectorOwner = server.BlockSelector
	server.selectorCaptured = true
	return server
}

// The application constructor attaches its captured fields before invoking
// the selector factory, so extensions observe the same initialized coordinator.
func newTLCServer(fileName string, configName string, metadir string, manager *DistributedFPSetManager, queue StateQueue, trace *TLCTrace) *TLCServer {
	server, err := newTLCServerConfigured(fileName, configName, metadir, manager, queue, trace, nil)
	if err != nil {
		panic(err)
	}
	return server
}

func newTLCServerConfigured(fileName string, configName string, metadir string, manager *DistributedFPSetManager, queue StateQueue, trace *TLCTrace, configure func(*TLCServer) error) (*TLCServer, error) {
	InitializeTLCServerProperties()
	if manager == nil {
		manager = NewDistributedFPSetManager()
	}
	if queue == nil {
		queue = NewStateQueue(metadir)
	}
	if trace == nil {
		trace = NewTLCTrace(metadir, fileName)
	}
	server := &TLCServer{
		InternTable:                 internTable,
		FPSetManager:                manager,
		StateQueue:                  queue,
		Trace:                       trace,
		Metadir:                     metadir,
		FileName:                    fileName,
		ConfigName:                  configName,
		threadsToWorkers:            NewInsMap[*TLCServerThread, DistributedWorkerEndpoint](),
		FinalNumberOfDistinctStates: -1,
	}
	if configure != nil {
		if err := configure(server); err != nil {
			return nil, err
		}
	}
	server.fpSetManagerOwner = server.FPSetManager
	server.stateQueueOwner = server.StateQueue
	server.traceOwner = server.Trace
	server.componentsCaptured = true
	server.metadirOwner = server.Metadir
	server.fileNameOwner = server.FileName
	server.configNameOwner = server.ConfigName
	server.metadataCaptured = true
	return server, nil
}

func (s *TLCServer) ownedFPSetManager() *DistributedFPSetManager {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.componentsCaptured {
		return s.fpSetManagerOwner
	}
	return s.FPSetManager
}

func (s *TLCServer) ownedStateQueue() StateQueue {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.componentsCaptured {
		return s.stateQueueOwner
	}
	return s.StateQueue
}

func (s *TLCServer) ownedTrace() *TLCTrace {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.componentsCaptured {
		return s.traceOwner
	}
	return s.Trace
}

func (s *TLCServer) ownedMetadir() string {
	s.require()
	if s.metadataCaptured {
		return s.metadirOwner
	}
	return s.Metadir
}

func (s *TLCServer) ownedFileName() string {
	s.require()
	if s.metadataCaptured {
		return s.fileNameOwner
	}
	return s.FileName
}

func (s *TLCServer) ownedConfigName() string {
	s.require()
	if s.metadataCaptured {
		return s.configNameOwner
	}
	return s.ConfigName
}

func (s *TLCServer) ownedBlockSelector() BlockSelection {
	s.require()
	if s.selectorCaptured {
		return s.blockSelectorOwner
	}
	return s.BlockSelector
}

func (s *TLCServer) SetTool(tool *Tool) *TLCServer {
	s.require()
	s.Tool = tool
	s.app = nil
	if tool != nil {
		s.app = NewTLCApp(tool, s.GetCheckDeadlock())
	}
	s.serverInternTable()
	if s.ownedTrace() != nil {
		s.ownedTrace().SetTool(tool)
	}
	return s
}

func (s *TLCServer) HasNoErrors() bool {
	s.require()
	s.monitor.Lock()
	defer s.monitor.Unlock()
	return s.ErrState == nil
}

func (s *TLCServer) application() *TLCApp {
	s.require()
	if s.app == nil {
		s.app = NewTLCApp(s.Tool, s.GetCheckDeadlock())
	}
	return s.app
}

func (s *TLCServer) SetCheckDeadlock(check bool) *TLCServer {
	s.require()
	s.checkDeadlock = &check
	if s.app != nil {
		s.app.checkDeadlock = check
	}
	return s
}

func (s *TLCServer) GetCheckDeadlock() bool {
	s.require()
	if s.checkDeadlock != nil {
		return *s.checkDeadlock
	}
	// TLCApp uses the command-line flag, independently of CHECK_DEADLOCK in
	// ModelConfig. Its command-line default is true.
	return true
}

func (s *TLCServer) GetPreprocess() bool {
	s.require()
	return true // TLCApp's final preprocess field.
}

func (s *TLCServer) GetIrredPolyForFP() uint64 {
	s.require()
	return FP64GetIrredPoly()
}

func (s *TLCServer) Checkpoint() error {
	if !s.requireStateQueue().SuspendAll() {
		return nil
	}
	PrintMessage(ECTLCCheckpointStart, "-- Checkpointing of run "+s.ownedMetadir()+" compl")
	if err := s.requireStateQueue().BeginChkpt(); err != nil {
		return err
	}
	if s.ownedTrace() == nil {
		panic(NewNullPointerException())
	}
	if err := s.ownedTrace().BeginChkpt(); err != nil {
		return err
	}
	if s.ownedFPSetManager() == nil {
		panic(NewNullPointerException())
	}
	if err := s.ownedFPSetManager().Checkpoint(s.checkpointFileName()); err != nil {
		return err
	}
	s.requireStateQueue().ResumeAll()
	if err := s.serverInternTable().beginChkptWithVarCount(s.ownedMetadir(), s.serverInternTable().varCount); err != nil {
		return err
	}
	if err := s.requireStateQueue().CommitChkpt(); err != nil {
		return err
	}
	if s.ownedTrace() == nil {
		panic(NewNullPointerException())
	}
	if err := s.ownedTrace().CommitChkpt(); err != nil {
		return err
	}
	if err := s.serverInternTable().CommitChkpt(s.ownedMetadir()); err != nil {
		return err
	}
	if s.ownedFPSetManager() == nil {
		panic(NewNullPointerException())
	}
	if err := s.ownedFPSetManager().CommitCheckpoint(); err != nil {
		return err
	}
	PrintMessage(ECTLCCheckpointEnd, "eted.")
	return nil
}

func (s *TLCServer) Recover() error {
	if s == nil || s.ownedTrace() == nil {
		panic(NewNullPointerException())
	}
	if err := s.ownedTrace().Recover(); err != nil {
		return err
	}
	if err := s.requireStateQueue().Recover(); err != nil {
		return err
	}
	if s.ownedFPSetManager() == nil {
		panic(NewNullPointerException())
	}
	return s.ownedFPSetManager().Recover(s.checkpointFileName())
}

func (s *TLCServer) Close(cleanup bool) error {
	if s == nil || s.ownedTrace() == nil {
		panic(NewNullPointerException())
	}
	if err := s.ownedTrace().Close(); err != nil {
		return err
	}
	if s.ownedFPSetManager() == nil {
		panic(NewNullPointerException())
	}
	if err := s.ownedFPSetManager().Close(cleanup); err != nil {
		return err
	}
	if cleanup && !distributedVetoCleanup() {
		deleteDirLikeJava(s.ownedMetadir(), true)
	}
	return nil
}

func (s *TLCServer) DoInit(tool ...*Tool) (int, error) {
	if len(tool) > 1 {
		panic(NewIllegalArgumentException("DoInit accepts at most one tool"))
	}
	if s == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server is nil")
	}
	if len(tool) > 0 {
		s.SetTool(tool[0])
	}
	if s.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server has no tool")
	}
	functor := &distributedDoInitFunctor{server: s, app: s.application()}
	if err := functor.app.GetInitStates(NewStateFunctor(functor.AddElement)); err != nil {
		return ECGeneral, err
	}
	if functor.err != nil {
		return ECGeneral, functor.err
	}
	return NoError, nil
}

func (s *TLCServer) ModelCheck(tool ...*Tool) (int, error) {
	if len(tool) > 1 {
		panic(NewIllegalArgumentException("ModelCheck accepts at most one tool"))
	}
	if s == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server is nil")
	}
	startTime := time.Now()
	if len(tool) > 0 {
		s.SetTool(tool[0])
	}
	if s.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server has no tool")
	}
	app := s.application()
	recovered := false
	if app.CanRecover() {
		PrintMessage(ECTLCCheckpointRecoverStart, s.ownedMetadir())
		if err := s.Recover(); err != nil {
			return ECGeneral, err // Recovery is outside the init catch(Throwable).
		}
		PrintMessage(ECTLCCheckpointRecoverEnd, fmtInt64(int64(s.fpSetSize())), fmtInt64(s.requireStateQueue().Size()))
		recovered = true
	}
	publication := s.publicationBoundaries()
	hostname, err := invokeRegistryBoundary(publication.LocalHostName)
	if err != nil {
		return ECGeneral, err
	}
	registry, err := invokeRegistryBoundary(func() (*TLCServerRegistry, error) {
		return publication.CreateRegistry(TLCServerPort())
	})
	if err != nil {
		return ECGeneral, err
	}
	if registry == nil {
		return ECGeneral, NewNullPointerException()
	}
	if err := invokeRegistryOperation(func() error {
		if registry.Rebind == nil {
			return NewNullPointerException()
		}
		return registry.Rebind(TLCServerName, s)
	}); err != nil {
		return ECGeneral, err
	}
	s.WaitForFPSetManager()
	result := NoError
	if !recovered {
		if err := s.runDistributedInit(true); err != nil {
			s.SetDone()
			result = s.reportDistributedInitFailure(app, err)
		}
	}
	if s.IsDone() {
		s.PrintSummary(1, 0, s.requireStateQueue().Size(), s.fpSetSize(), false)
		PrintMessage(ECTLCFinished, ConvertRuntimeToHumanReadable(time.Now().UnixMilli()-startTime.UnixMilli()))
		s.executor.Shutdown()
		if err := s.Close(false); err != nil {
			return ECGeneral, err
		}
		return result, nil
	}
	if err := invokeRegistryOperation(func() error {
		if registry.Rebind == nil {
			return NewNullPointerException()
		}
		return registry.Rebind(TLCServerWorkerName, s)
	}); err != nil {
		return ECGeneral, err
	}
	PrintMessage(ECTLCDistributedServerRunning, hostname)
	if err := s.waitForDistributedCompletion(); err != nil {
		return ECGeneral, err
	}
	if s.HasNoErrors() && !s.requireStateQueue().IsEmpty() {
		return ECGeneral, NewTLCRuntimeException(ECGeneral)
	}
	for _, thread := range s.GetServerThreads() {
		if thread == nil {
			continue
		}
		thread.Join()
		cacheRatio := distributedWorkerCacheRatio(thread.GetCacheRateRatio())
		PrintMessage(ECTLCDistributedWorkerStats,
			thread.GetURI(),
			fmtInt(thread.GetSentStates()),
			fmtInt(thread.GetReceivedStates()),
			cacheRatio,
		)
		exitErr := func() error {
			defer s.removeServerThreadOnly(thread)
			if err := invokeRegistryOperation(thread.Worker.Exit); err != nil {
				if isIgnorableDistributedWorkerExit(err) {
					PrintWarning(ECGeneral, "Ignoring attempt to exit dead worker")
				} else {
					return err
				}
			}
			return nil
		}()
		if exitErr != nil {
			return ECGeneral, exitErr
		}
	}
	s.executor.Shutdown()
	s.FinalNumberOfDistinctStates = int64(s.fpSetSize())
	tlcServerFinalDistinctStates.Store(s.FinalNumberOfDistinctStates)
	statesGenerated := s.GetStatesGenerated()
	statesLeft := s.GetNewStates()
	level, err := s.ownedTrace().GetLevelForReportingWithError()
	if err != nil {
		return ECGeneral, err
	}
	atomic.StoreInt64(&s.StatesPerMinute, 0)
	atomic.StoreInt64(&s.DistinctStatesPerMinute, 0)
	if s.HasNoErrors() {
		actualDistance := s.ownedFPSetManager().CheckFPs()
		statesSeen := s.ownedFPSetManager().GetStatesSeen()
		ReportSuccessCountsDistance(uint64(TLCServerFinalNumberOfDistinctStates()), actualDistance, int64(statesSeen))
	} else if s.KeepCallStack {
		app.SetCallStack()
		s.Tool = app.Tool
		if s.ownedTrace() != nil {
			s.ownedTrace().SetTool(app.Tool)
		}
	}
	s.PrintSummary(level, statesGenerated, statesLeft, uint64(TLCServerFinalNumberOfDistinctStates()), s.HasNoErrors())
	PrintMessage(ECTLCFinished, ConvertRuntimeToHumanReadable(time.Now().UnixMilli()-startTime.UnixMilli()))
	if publication.Flush != nil {
		if err := invokeRegistryOperation(func() error { publication.Flush(); return nil }); err != nil {
			return ECGeneral, err
		}
	}
	if err := s.Close(s.HasNoErrors()); err != nil {
		return ECGeneral, err
	}
	for _, name := range []string{TLCServerWorkerName, TLCServerName} {
		if err := invokeRegistryOperation(func() error {
			if registry.Unbind == nil {
				return NewNullPointerException()
			}
			return registry.Unbind(name)
		}); err != nil {
			return ECGeneral, err
		}
	}
	if _, err := invokeRegistryBoundary(func() (bool, error) { return publication.Unexport(s, false) }); err != nil {
		return ECGeneral, err
	}
	if s.HasNoErrors() {
		return NoError, nil
	}
	if s.ErrorCode != NoError {
		return s.ErrorCode, nil // Java reports worker failures and returns normally.
	}
	return ECGeneral, nil
}

func (s *TLCServer) PrintInitGenerated() {
	s.require()
	statesGenerated := s.GetStatesGenerated()
	distinct := s.fpSetSize()
	plural := ""
	if statesGenerated != 1 {
		plural = "s"
	}
	if uint64(statesGenerated) == distinct {
		PrintMessage(ECTLCInitGenerated1, fmtInt64(statesGenerated), plural)
		return
	}
	PrintMessage(ECTLCInitGenerated2, fmtInt64(statesGenerated), plural, fmtInt64(int64(distinct)))
}

func (s *TLCServer) PrintSummary(level int, statesGenerated int64, statesLeftInQueue int64, distinctStates uint64, success bool) {
	level = int(int32(level))
	if toolMode() {
		PrintMessage(ECTLCProgressStats, fmtInt(level), MessageNumberFormat(statesGenerated), MessageNumberFormat(int64(distinctStates)), MessageNumberFormat(statesLeftInQueue), "0", "0")
	}
	PrintMessage(ECTLCStats, fmtInt64(statesGenerated), fmtInt64(int64(distinctStates)), fmtInt64(statesLeftInQueue))
	if success {
		PrintMessage(ECTLCSearchDepth, fmtInt(level))
	}
}

func (s *TLCServer) ReportSuccess() {
	s.require()
	distinct := s.fpSetSize()
	actualDistance := s.ownedFPSetManager().CheckFPs()
	statesSeen := s.ownedFPSetManager().GetStatesSeen()
	ReportSuccessCountsDistance(distinct, actualDistance, int64(statesSeen))
}

func (s *TLCServer) fpSetSize() uint64 {
	s.require()
	if s.ownedFPSetManager() == nil {
		panic(NewNullPointerException())
	}
	return s.ownedFPSetManager().Size()
}

func distributedServerHost() string {
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}
	return "localhost"
}

func (s *TLCServer) RegisterWorker(worker DistributedWorkerEndpoint) error {
	if s == nil {
		return NewNullPointerException()
	}
	// Java serializes registration, including both remote getURI calls.
	s.monitor.Lock()
	defer s.monitor.Unlock()
	if isNil(s.ownedStateQueue()) {
		return NewNullPointerException()
	}
	s.ownedStateQueue().ResumeAllStuck()
	if isNil(worker) {
		return NewNullPointerException()
	}
	uri, err := worker.GetURI()
	if err != nil {
		return err
	}
	thread := NewTLCServerThread(worker, uri, s, s.ownedBlockSelector())
	s.RegisterTLCServerThread(thread)
	thread.Start()
	// The second call is intentionally after registration/start. A failure
	// here does not undo the registered thread in the Java implementation.
	uri, err = worker.GetURI()
	if err != nil {
		return err
	}
	PrintMessage(ECTLCDistributedWorkerRegistered, uri)
	if local, ok := worker.(*LocalWorkerEndpoint); ok && local.Worker != nil {
		runtime := local.Worker.Runtime
		if runtime != nil && !runtime.launchKeepAlive {
			runtime.StartKeepAlive(NewLocalServerEndpoint(s))
		}
	}
	return nil
}

func (s *TLCServer) RegisterTLCServerThread(thread *TLCServerThread) {
	s.require()
	if thread == nil {
		panic(NewNullPointerException())
	}
	if thread.Worker == nil || thread.Worker.Worker == nil {
		panic(NewNullPointerException())
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	s.threadsMu.Lock()
	defer s.threadsMu.Unlock()
	if s.threadsToWorkers == nil {
		s.threadsToWorkers = NewInsMap[*TLCServerThread, DistributedWorkerEndpoint]()
	}
	// URI is display metadata. Even the same remote worker can be registered
	// more than once; Java's map keys are the distinct server-thread objects.
	s.threadsToWorkers.Set(thread, thread.Worker.Worker)
}

func (s *TLCServer) removeServerThreadOnly(thread *TLCServerThread) DistributedWorkerEndpoint {
	s.require()
	if thread == nil {
		panic(NewNullPointerException())
	}
	s.threadsMu.Lock()
	defer s.threadsMu.Unlock()
	if s.threadsToWorkers == nil {
		return nil
	}
	worker := s.threadsToWorkers.Get(thread)
	s.threadsToWorkers.Delkey(thread)
	return worker
}

func (s *TLCServer) RemoveTLCServerThread(thread *TLCServerThread) DistributedWorkerEndpoint {
	worker := s.removeServerThreadOnly(thread)
	if worker != nil {
		PrintMessage(ECTLCDistributedWorkerDeregistered, thread.GetURI())
	}
	return worker
}

// GetServerThreads snapshots Java's concurrent registry in Go insertion order.
// The lock does not cover joins, remote calls, queue waits, or diagnostics.
func (s *TLCServer) GetServerThreads() []*TLCServerThread {
	s.require()
	s.monitor.Lock()
	defer s.monitor.Unlock()
	s.threadsMu.Lock()
	defer s.threadsMu.Unlock()
	if s.threadsToWorkers == nil {
		return nil
	}
	threads := make([]*TLCServerThread, 0, s.threadsToWorkers.Len())
	for thread := range s.threadsToWorkers.All() {
		threads = append(threads, thread)
	}
	return threads
}

func (s *TLCServer) SetDone() {
	s.require()
	s.Done.Store(true)
}

func (s *TLCServer) SetErrState(curState *TLCStateMut, succState *TLCStateMut, keepCallStack bool, errorCode ...int) bool {
	s.require()
	if len(errorCode) > 1 {
		panic(NewIllegalArgumentException("SetErrState accepts at most one error code"))
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	if s.Done.Load() {
		return false
	}
	s.Done.Store(true)
	s.PredErrState = curState
	if succState == nil {
		s.ErrState = curState
	} else {
		s.ErrState = succState
	}
	s.KeepCallStack = keepCallStack
	if len(errorCode) > 0 {
		s.ErrorCode = int(int32(errorCode[0]))
	}
	return true
}

type distributedDoInitFunctor struct {
	server *TLCServer
	app    *TLCApp
	err    error
}

func (f *distributedDoInitFunctor) AddElement(curState *TLCStateMut) (result any, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
		if err != nil {
			if isJavaError(err) {
				panic(err) // DoInitFunctor catches Exception, not Error.
			}
			if f.server.SetErrState(curState, nil, true, ECGeneral) {
				f.err = err
			}
			err = nil
		}
		result = curState
	}()
	if f.err != nil {
		return curState, nil
	}
	inModel, err := f.app.IsInModel(curState)
	if err != nil {
		return curState, err
	}
	seen := false
	if inModel {
		fp := curState.FingerPrint()
		seen = f.server.ownedFPSetManager().Put(fp)
		if !seen {
			if f.server.ownedTrace() == nil {
				panic(NewNullPointerException())
			}
			if _, err := f.server.ownedTrace().WriteStateRecord(nil, fp, curState); err != nil {
				return curState, err
			}
			f.server.requireStateQueue().Enqueue(curState)
		}
	}
	if !inModel || !seen {
		if err := f.app.CheckState(nil, curState); err != nil {
			return curState, err
		}
	}
	return curState, nil
}

func distributedVetoCleanup() bool {
	InitializeTLCServerProperties()
	return tlcServerProperties.vetoCleanup
}

func (s *TLCServer) IsRunning() bool {
	s.require()
	return !s.Done.Load()
}

func (s *TLCServer) IsDone() bool {
	s.require()
	return s.Done.Load()
}

func (s *TLCServer) AddStatesGeneratedDelta(delta int64) {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.WorkerStatesGenerated.Add(delta)
}

func (s *TLCServer) GetStatesGenerated() int64 {
	if s == nil {
		panic(NewNullPointerException())
	}
	total := s.WorkerStatesGenerated.Load()
	if s.ownedFPSetManager() == nil {
		panic(NewNullPointerException())
	}
	return total + int64(s.ownedFPSetManager().GetStatesSeen())
}

func (s *TLCServer) GetNewStates() int64 {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	return s.getNewStatesLocked()
}

func (s *TLCServer) getNewStatesLocked() int64 {
	s.require()
	size := s.requireStateQueue().Size()
	for _, thread := range s.GetServerThreads() {
		size += int64(thread.GetCurrentSize())
	}
	return size
}

func (s *TLCServer) GetStatesGeneratedPerMinute() int64 {
	s.require()
	return atomic.LoadInt64(&s.StatesPerMinute)
}

func (s *TLCServer) GetDistinctStatesGeneratedPerMinute() int64 {
	s.require()
	return atomic.LoadInt64(&s.DistinctStatesPerMinute)
}

func (s *TLCServer) GetAverageBlockCnt() int64 {
	s.require()
	selector := s.ownedBlockSelector()
	if isNil(selector) {
		panic(NewNullPointerException())
	}
	return selector.GetAverageBlockCnt()
}

func (s *TLCServer) GetWorkerCount() int {
	s.require()
	s.threadsMu.Lock()
	defer s.threadsMu.Unlock()
	if s.threadsToWorkers == nil {
		return 0
	}
	return s.threadsToWorkers.Len()
}

func (s *TLCServer) GetFPSetManager() *DistributedFPSetManager {
	s.require()
	if s.fpRegistration != nil {
		<-s.fpRegistration.done
	}
	return s.ownedFPSetManager()
}

func (s *TLCServer) GetSpecFileName() string {
	s.require()
	if s.app != nil {
		return s.app.GetFileName()
	}
	return s.ownedFileName()
}

func (s *TLCServer) GetConfigFileName() string {
	s.require()
	if s.app != nil {
		return s.app.GetConfigName()
	}
	return s.ownedConfigName()
}

func (s *TLCServer) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *TLCServer) requireStateQueue() StateQueue {
	s.require()
	if isNil(s.ownedStateQueue()) {
		panic(NewNullPointerException())
	}
	return s.ownedStateQueue()
}

var tlcServerThreadCount atomic.Int64

type distributedStateBlock struct {
	states []*TLCStateMut
}

type TLCServerThread struct {
	ID int
	// Statistics may be initialized before Run; concurrent observers use
	// the getters. Never hold statisticsMu across a worker/network call.
	statisticsMu       sync.Mutex
	ReceivedStates     int
	SentStates         int
	CacheRateHitRatio  float64
	Selector           BlockSelection
	states             atomic.Pointer[distributedStateBlock]
	TimerTask          *TLCTimerTask
	Worker             *DistributedWorkerSmartProxy
	Server             *TLCServer
	URI                string
	cleanupGlobals     atomic.Bool
	started            atomic.Bool
	keepAliveStopped   atomic.Bool
	keepAliveDone      chan struct{}
	runDone            chan struct{}
	componentsCaptured bool
	selectorOwner      BlockSelection
	timerTaskOwner     *TLCTimerTask
	workerOwner        *DistributedWorkerSmartProxy
	serverOwner        *TLCServer
}

func NewTLCServerThread(worker DistributedWorkerEndpoint, uri string, server *TLCServer, selector BlockSelection) *TLCServerThread {
	thread := &TLCServerThread{
		ID:                int(int32(tlcServerThreadCount.Add(1) - 1)),
		CacheRateHitRatio: -1,
		Selector:          selector,
		Worker:            NewDistributedWorkerSmartProxy(worker),
		Server:            server,
		URI:               uri,
		keepAliveDone:     make(chan struct{}),
		runDone:           make(chan struct{}),
	}
	thread.setStates([]*TLCStateMut{})
	thread.cleanupGlobals.Store(true)
	thread.TimerTask = &TLCTimerTask{Thread: thread, threadCaptured: true, threadOwner: thread}
	thread.selectorOwner = thread.Selector
	thread.timerTaskOwner = thread.TimerTask
	thread.workerOwner = thread.Worker
	thread.serverOwner = thread.Server
	thread.componentsCaptured = true
	// Java schedules keepalive during construction, before Thread.start().
	thread.startKeepAlive()
	return thread
}

func (t *TLCServerThread) Name() string {
	t.require()
	uri := &distributedWorkerURIValue{raw: t.URI}
	return fmt.Sprintf("%s%03d-[%s]", TLCServerThreadNamePrefix, t.ID, uri.asciiString())
}

func (t *TLCServerThread) Start() {
	t.require()
	if !t.started.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer close(t.runDone)
		// Uncaught source thread failures stop this goroutine, after Run's
		// catch/finally sequence, without terminating the process.
		defer func() {
			if failure := recover(); failure != nil {
				fmt.Fprint(os.Stderr, javaThrowableStackTrace(panicValueAsError(failure)))
			}
		}()
		t.Run()
	}()
}

func (t *TLCServerThread) Join() {
	t.require()
	if t.runDone == nil {
		panic(NewNullPointerException())
	}
	<-t.runDone
}

func (t *TLCServerThread) startKeepAlive() {
	t.require()
	if t.ownedTimerTask() == nil || t.keepAliveDone == nil {
		panic(NewNullPointerException())
	}
	go t.runKeepAlive()
}

func (t *TLCServerThread) runKeepAlive() {
	t.require()
	if t.ownedTimerTask() == nil || t.keepAliveDone == nil {
		panic(NewNullPointerException())
	}
	// An uncaught source timer failure stops its thread, not model checking.
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprint(os.Stderr, javaThrowableStackTrace(panicValueAsError(failure)))
		}
	}()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			started := time.Now()
			t.ownedTimerTask().Run()
			select {
			case <-t.keepAliveDone:
				return
			default:
			}
			// Timer.schedule measures from the preceding actual execution start.
			timer.Reset(time.Until(started.Add(60 * time.Second)))
		case <-t.keepAliveDone:
			return
		}
	}
}

func (t *TLCServerThread) Run() {
	t.require()
	IncNumWorkers(1)
	server := t.ownedServer()
	if server == nil {
		panic(NewNullPointerException())
	}
	stateQueue := distributedJavaStateQueue(server.ownedStateQueue())
	// Register finally separately so an exception raised by the catch itself
	// still executes it. A failure within finally skips its remaining steps.
	defer func() {
		t.readCacheRateRatio()
		t.cancelKeepAlive()
		t.setStates([]*TLCStateMut{})
	}()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.handleRunError(panicValueAsError(recovered), stateQueue)
		}
	}()
	for {
		t.setStates(t.requireSelector().GetBlocks(stateQueue, t.ownedWorker()))
		if t.currentStates() == nil {
			server.monitor.Lock()
			server.SetDone()
			server.notifyCompletionLocked()
			server.monitor.Unlock()
			if stateQueue == nil {
				panic(NewNullPointerException())
			}
			stateQueue.FinishAll()
			return
		}
		if len(t.currentStates()) == 0 {
			continue
		}
		t.statisticsMu.Lock()
		t.SentStates = int(int32(t.SentStates) + int32(len(t.currentStates())))
		t.statisticsMu.Unlock()

		res, ok := t.computeBlock(stateQueue)
		if !ok {
			return
		}
		if res == nil {
			continue
		}
		newStates := res.GetNextStates()
		newFps := res.GetNextFingerprints()
		if err := t.publishBlock(stateQueue, newStates, newFps); err != nil {
			panic(err)
		}
	}
}

// Java's inner remote/NPE catch includes result dereferences, statistics,
// the keepalive timestamp and generated-state delta, before publishing the block.
func (t *TLCServerThread) computeBlockAttempt() (res *NextStateResult, err error) {
	t.require()
	defer func() {
		if failure := recover(); failure != nil {
			res = nil
			err = panicValueAsError(failure)
		}
	}()
	res, err = invokeDistributedWorker(t.ownedWorker(), t.currentStates())
	if err != nil {
		return nil, err
	}
	if res == nil {
		panic(NewNullPointerException())
	}
	newStates := res.GetNextStates()
	if newStates == nil {
		panic(NewNullPointerException())
	}
	if len(newStates) == 0 {
		panic(NewArrayIndexOutOfBoundsException(0, 0))
	}
	if newStates[0] == nil {
		panic(NewNullPointerException())
	}
	received := newStates[0].Size()
	t.statisticsMu.Lock()
	t.ReceivedStates = int(int32(t.ReceivedStates) + int32(received))
	t.statisticsMu.Unlock()
	t.ownedTimerTask().SetLastInvocation(time.Now())
	t.ownedServer().AddStatesGeneratedDelta(res.GetStatesComputedDelta())
	return res, nil
}

func (t *TLCServerThread) computeBlock(stateQueue StateQueue) (*NextStateResult, bool) {
	t.require()
	stateQueue = distributedJavaStateQueue(stateQueue)
	res, err := t.computeBlockAttempt()
	if err == nil {
		return res, true
	}
	if isDistributedRemoteFailure(err) {
		if isRecoverableDistributedError(err) && len(t.currentStates()) > 1 {
			PrintMessage(ECTLCDistributedExceedBlocksize, fmtInt(len(t.currentStates())/2))
			if stateQueue == nil {
				panic(NewNullPointerException())
			}
			stateQueue.SEnqueueAll(t.currentStates())
			t.requireSelector().SetMaxTXSize(len(t.currentStates()) / 2)
			return nil, true
		}
		PrintMessage(ECTLCDistributedWorkerLost, t.GetURI())
		t.HandleRemoteWorkerLost(stateQueue)
		return nil, false
	}
	if isDistributedNullFailure(err) {
		PrintMessage(ECTLCDistributedWorkerLost, "\n"+javaThrowableStackTrace(err))
		t.HandleRemoteWorkerLost(stateQueue)
		return nil, false
	}
	// Other exceptions escape Java's inner remote/NPE catches into the
	// server thread's outer Throwable catch, including WorkerException.
	panic(err)
}

func invokeDistributedWorker(worker *DistributedWorkerSmartProxy, states []*TLCStateMut) (result *NextStateResult, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			result = nil
			err = panicValueAsError(failure)
		}
	}()
	return worker.GetNextStates(states)
}

func (t *TLCServerThread) publishBlock(stateQueue StateQueue, newStates []*StateVec, newFps []*LongVec) error {
	t.require()
	stateQueue = distributedJavaStateQueue(stateQueue)
	server := t.ownedServer()
	if server == nil || server.ownedFPSetManager() == nil {
		panic(NewNullPointerException())
	}
	visited := server.ownedFPSetManager().PutBlock(newFps, &server.executor)
	for i, vector := range visited {
		iter := NewBitVectorIter(vector)
		for {
			index := iter.Next()
			if index == -1 {
				break
			}
			// Java dereferences the selected partition and state after FP
			// insertion. A malformed result must reach the outer failure
			// catch; skipping it silently loses work behind a visited FP.
			if newStates == nil {
				panic(NewNullPointerException())
			}
			if i >= len(newStates) {
				panic(NewArrayIndexOutOfBoundsException(i, len(newStates)))
			}
			if newStates[i] == nil {
				panic(NewNullPointerException())
			}
			if !newStates[i].distributed && index >= len(newStates[i].states) {
				panic(NewArrayIndexOutOfBoundsException(index, len(newStates[i].states)))
			}
			state := newStates[i].At(index)
			fingerprints := fpBlockAt(newFps, i)
			if fingerprints == nil {
				panic(NewNullPointerException())
			}
			fp := uint64(fingerprints.ElementAt(index))
			if state == nil || server.ownedTrace() == nil {
				panic(NewNullPointerException())
			}
			// TLCWorker returns a successor whose UID still identifies its
			// predecessor. Write that UID before replacing the trace location.
			if _, err := server.ownedTrace().WriteStateRecord(state, fp, state); err != nil {
				return err
			}
			if stateQueue == nil {
				panic(NewNullPointerException())
			}
			stateQueue.SEnqueue(state)
		}
	}
	return nil
}

func (t *TLCServerThread) handleRunError(err error, stateQueue StateQueue) {
	t.require()
	stateQueue = distributedJavaStateQueue(stateQueue)
	server := t.ownedServer()
	if server == nil {
		panic(NewNullPointerException())
	}
	var state1, state2 *TLCStateMut
	if failure, ok := err.(*WorkerException); ok && failure != nil {
		state1, state2 = failure.State1, failure.State2
	}
	if server.SetErrState(state1, nil, true, ECGeneral) {
		server.LastError = err
		if state1 != nil {
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						failure := panicValueAsError(recovered)
						// Java catches Exception around trace printing, not Error.
						if isJavaError(failure) {
							panic(recovered)
						}
						PrintError(ECGeneral, generalErrorParams("", failure)...)
					}
				}()
				if server.ownedTrace() == nil {
					panic(NewNullPointerException())
				}
				server.ownedTrace().PrintTrace(state1, state2)
			}()
		} else {
			PrintError(ECGeneral, generalErrorParams("", err)...)
		}
		if stateQueue == nil {
			panic(NewNullPointerException())
		}
		stateQueue.FinishAll()
		server.monitor.Lock()
		server.notifyCompletionLocked()
		server.monitor.Unlock()
	}
}

func (t *TLCServerThread) HandleRemoteWorkerLost(stateQueue StateQueue) {
	t.require()
	stateQueue = distributedJavaStateQueue(stateQueue)
	t.cancelKeepAlive()
	if !t.cleanupGlobals.CompareAndSwap(true, false) {
		return
	}
	server := t.ownedServer()
	if server == nil {
		panic(NewNullPointerException())
	}
	server.RemoveTLCServerThread(t)
	if stateQueue != nil {
		states := t.currentStates()
		// Source cleanup substitutes an empty array for an absent work block.
		// The queue itself still rejects nil batches supplied by other callers.
		if states == nil {
			states = []*TLCStateMut{}
		}
		stateQueue.SEnqueueAll(states)
	}
	t.setStates([]*TLCStateMut{})
	if stateQueue != nil {
		stateQueue.WakeAllWaiters()
	}
	DecNumWorkers()
}

func (t *TLCServerThread) setStates(states []*TLCStateMut) {
	t.require()
	t.states.Store(&distributedStateBlock{states: states})
}

func (t *TLCServerThread) currentStates() []*TLCStateMut {
	t.require()
	block := t.states.Load()
	if block == nil {
		return nil
	}
	return block.states
}

func (t *TLCServerThread) GetCurrentSize() int {
	t.require()
	states := t.currentStates()
	if states == nil {
		panic(NewNullPointerException())
	}
	return len(states)
}

func (t *TLCServerThread) GetURI() string {
	t.require()
	return t.URI
}

func (t *TLCServerThread) GetReceivedStates() int {
	t.require()
	t.statisticsMu.Lock()
	defer t.statisticsMu.Unlock()
	return t.ReceivedStates
}

func (t *TLCServerThread) GetSentStates() int {
	t.require()
	t.statisticsMu.Lock()
	defer t.statisticsMu.Unlock()
	return t.SentStates
}

func (t *TLCServerThread) GetCacheRateRatio() float64 {
	t.require()
	t.statisticsMu.Lock()
	defer t.statisticsMu.Unlock()
	return t.CacheRateHitRatio
}

func (t *TLCServerThread) readCacheRateRatio() {
	t.require()
	worker := t.ownedWorker()
	if worker == nil {
		panic(NewNullPointerException())
	}
	ratio, err := invokeRegistryBoundary(worker.GetCacheRateRatio)
	if err != nil {
		if !isDistributedRemoteFailure(err) {
			panic(err)
		}
		PrintWarning(ECGeneral, "Failed to read remote worker cache statistic (Expect to see a negative chache hit rate. Does not invalidate model checking results)")
		return
	}
	t.statisticsMu.Lock()
	t.CacheRateHitRatio = ratio
	t.statisticsMu.Unlock()
}

func (t *TLCServerThread) cancelKeepAlive() {
	t.require()
	if t.keepAliveDone == nil {
		panic(NewNullPointerException())
	}
	if t.keepAliveStopped.CompareAndSwap(false, true) {
		close(t.keepAliveDone)
	}
}

func (t *TLCServerThread) requireSelector() BlockSelection {
	t.require()
	selector := t.Selector
	if t.componentsCaptured {
		selector = t.selectorOwner
	}
	if isNil(selector) {
		panic(NewNullPointerException())
	}
	return selector
}

func (t *TLCServerThread) ownedTimerTask() *TLCTimerTask {
	t.require()
	if t.componentsCaptured {
		return t.timerTaskOwner
	}
	return t.TimerTask
}

func (t *TLCServerThread) ownedWorker() *DistributedWorkerSmartProxy {
	t.require()
	if t.componentsCaptured {
		return t.workerOwner
	}
	return t.Worker
}

func (t *TLCServerThread) ownedServer() *TLCServer {
	t.require()
	if t.componentsCaptured {
		return t.serverOwner
	}
	return t.Server
}

func (t *TLCServerThread) require() {
	if t == nil {
		panic(NewNullPointerException())
	}
}

type TLCTimerTask struct {
	Thread         *TLCServerThread
	LastInvocation atomic.Int64
	threadCaptured bool
	threadOwner    *TLCServerThread
}

func (t *TLCTimerTask) Run() {
	t.require()
	thread := t.Thread
	if t.threadCaptured {
		thread = t.threadOwner
	}
	if thread == nil {
		panic(NewNullPointerException())
	}
	now := time.Now().UnixMilli()
	last := t.LastInvocation.Load()
	if last == 0 || now-last > int64(time.Minute/time.Millisecond) {
		worker := thread.ownedWorker()
		if worker == nil {
			panic(NewNullPointerException())
		}
		alive, err := invokeRegistryBoundary(worker.IsAlive)
		if err != nil && !isDistributedRemoteFailure(err) {
			panic(err)
		}
		if err != nil || !alive {
			server := thread.ownedServer()
			if server == nil {
				panic(NewNullPointerException())
			}
			thread.HandleRemoteWorkerLost(server.ownedStateQueue())
		}
	}
}

func (t *TLCTimerTask) SetLastInvocation(when time.Time) {
	t.require()
	t.LastInvocation.Store(when.UnixMilli())
}

func (t *TLCTimerTask) require() {
	if t == nil {
		panic(NewNullPointerException())
	}
}

// Endpoint adapters classify failures at their origin. Coordinator retry
// decisions use those native traits rather than infer Java exception nesting.
func isRecoverableDistributedError(err error) bool {
	failure, ok := err.(*DistributedOperationError)
	return ok && failure != nil && failure.Remote && failure.Recoverable
}

type BlockSelectorMode int

const (
	BlockSelectorProportional BlockSelectorMode = iota
	BlockSelectorLimiting
	BlockSelectorStatistical
	BlockSelectorStatic
)

const (
	blockSelectorDefaultMaximum              = 8192
	blockSelectorDefaultStaticSize           = 1024
	blockSelectorDefaultNetworkOverheadLimit = 2.5 / 100.0

	distributedSelectorStaticProperty     = "tlc2.tool.distributed.selector.bsf.staticselector"
	distributedSelectorUnlimitingProperty = "tlc2.tool.distributed.selector.bsf.unlimitingselector"
	distributedSelectorLimitingProperty   = "tlc2.tool.distributed.selector.bsf.limitingselector"
	distributedStaticBlockSizeProperty    = "tlc2.tool.distributed.TLCServerThread.BlockSize"
)

type BlockSelector struct {
	Server                *TLCServer
	Mode                  BlockSelectorMode
	Maximum               int
	StaticBlockSize       int
	NetworkOverheadLimit  float64
	averageBlockCnt       atomic.Int64
	maximumMu             sync.RWMutex
	serverCaptured        bool
	serverOwner           *TLCServer
	configurationCaptured bool
	modeOwner             BlockSelectorMode
	staticBlockSizeOwner  int
	networkOverheadOwner  float64
}

func NewBlockSelector(server *TLCServer) BlockSelection {
	return NewBlockSelectorFromProperties(server)
}

var distributedSelectorStartup struct {
	sync.Once
	mode           BlockSelectorMode
	factoryName    string
	factoryPresent bool
}

var distributedStaticSelectorStartup struct {
	sync.Once
	size int
}

func NewBlockSelectorFromProperties(server *TLCServer) BlockSelection {
	distributedSelectorStartup.Do(func() {
		distributedSelectorStartup.factoryName, distributedSelectorStartup.factoryPresent = tlcLookupSystemProperty(distributedSelectorFactoryProperty)
		switch {
		case distributedBooleanProperty(distributedSelectorStaticProperty):
			distributedSelectorStartup.mode = BlockSelectorStatic
		case distributedBooleanProperty(distributedSelectorUnlimitingProperty):
			distributedSelectorStartup.mode = BlockSelectorProportional
		case distributedBooleanProperty(distributedSelectorLimitingProperty):
			distributedSelectorStartup.mode = BlockSelectorLimiting
		default:
			distributedSelectorStartup.mode = BlockSelectorStatistical
		}
	})
	if distributedSelectorStartup.factoryPresent {
		if factory := loadBlockSelectorFactory(distributedSelectorStartup.factoryName); factory != nil {
			return factory(server)
		}
	}
	switch distributedSelectorStartup.mode {
	case BlockSelectorStatic:
		return NewStaticBlockSelector(server)
	case BlockSelectorProportional:
		return NewProportionalBlockSelector(server)
	case BlockSelectorLimiting:
		return NewLimitingBlockSelector(server)
	default:
		return NewStatisticalBlockSelector(server)
	}
}

func NewProportionalBlockSelector(server *TLCServer) *BlockSelector {
	if server == nil {
		panic(NewTLCRuntimeExceptionMessage("TLC found a null TLCServer"))
	}
	selector := &BlockSelector{
		Server:               server,
		Mode:                 BlockSelectorProportional,
		Maximum:              blockSelectorDefaultMaximum,
		StaticBlockSize:      blockSelectorDefaultStaticSize,
		NetworkOverheadLimit: blockSelectorDefaultNetworkOverheadLimit,
		serverCaptured:       true,
		serverOwner:          server,
	}
	selector.captureConfiguration()
	return selector
}

func (b *BlockSelector) ownedServer() *TLCServer {
	if b.serverCaptured {
		return b.serverOwner
	}
	return b.Server
}

func (b *BlockSelector) captureConfiguration() {
	b.modeOwner = b.Mode
	b.staticBlockSizeOwner = b.StaticBlockSize
	b.networkOverheadOwner = b.NetworkOverheadLimit
	b.configurationCaptured = true
}

func (b *BlockSelector) ownedMode() BlockSelectorMode {
	if b.configurationCaptured {
		return b.modeOwner
	}
	return b.Mode
}

func (b *BlockSelector) ownedStaticBlockSize() int {
	if b.configurationCaptured {
		return b.staticBlockSizeOwner
	}
	return b.StaticBlockSize
}

func (b *BlockSelector) ownedNetworkOverheadLimit() float64 {
	if b.configurationCaptured {
		return b.networkOverheadOwner
	}
	return b.NetworkOverheadLimit
}

func NewLimitingBlockSelector(server *TLCServer, maximum ...int) *BlockSelector {
	if len(maximum) > 1 {
		panic(NewIllegalArgumentException("NewLimitingBlockSelector accepts at most one maximum"))
	}
	selector := NewProportionalBlockSelector(server)
	selector.Mode = BlockSelectorLimiting
	if len(maximum) > 0 {
		selector.Maximum = int(int32(maximum[0]))
	}
	selector.captureConfiguration()
	return selector
}

func NewStatisticalBlockSelector(server *TLCServer) *BlockSelector {
	selector := NewLimitingBlockSelector(server)
	selector.Mode = BlockSelectorStatistical
	selector.captureConfiguration()
	return selector
}

func NewStaticBlockSelector(server *TLCServer, blockSize ...int) *BlockSelector {
	if len(blockSize) > 1 {
		panic(NewIllegalArgumentException("NewStaticBlockSelector accepts at most one block size"))
	}
	// Capture this setting independently of factory selection, before the
	// required-server assertion. Non-static selectors do not initialize it.
	distributedStaticSelectorStartup.Do(func() {
		distributedStaticSelectorStartup.size = blockSelectorDefaultStaticSize
		if value, ok := distributedIntProperty(distributedStaticBlockSizeProperty); ok {
			distributedStaticSelectorStartup.size = value
		}
	})
	selector := NewProportionalBlockSelector(server)
	selector.Mode = BlockSelectorStatic
	selector.StaticBlockSize = distributedStaticSelectorStartup.size
	if len(blockSize) > 0 {
		selector.StaticBlockSize = int(int32(blockSize[0]))
	}
	selector.averageBlockCnt.Store(int64(selector.StaticBlockSize))
	selector.captureConfiguration()
	return selector
}

func distributedBooleanProperty(name string) bool {
	value, ok := tlcLookupSystemProperty(name)
	return ok && javaBooleanProperty(value)
}

func distributedIntProperty(name string) (int, bool) {
	value, ok := tlcLookupSystemProperty(name)
	if !ok {
		return 0, false
	}
	return javaDecodeIntProperty(value)
}

func (b *BlockSelector) GetBlocks(stateQueue StateQueue, worker *DistributedWorkerSmartProxy) []*TLCStateMut {
	b.require()
	if isNil(stateQueue) {
		panic(NewNullPointerException())
	}
	if b.ownedMode() == BlockSelectorStatic {
		return stateQueue.SDequeueMany(b.ownedStaticBlockSize())
	}
	amountOfStates := stateQueue.Size()
	blockSize := b.GetBlockSize(amountOfStates, worker)
	if blockSize > amountOfStates {
		blockSize = amountOfStates
	}
	if blockSize < 1 {
		blockSize = 1
	}
	if blockSize > int64(maxJavaInt) {
		blockSize = int64(maxJavaInt)
	}
	states := stateQueue.SDequeueMany(int(blockSize))
	b.SetAverageBlockCnt(int64(len(states)))
	return states
}

func (b *BlockSelector) SetMaxTXSize(maximum int) {
	b.require()
	maximum = int(int32(maximum))
	if b.ownedMode() == BlockSelectorLimiting || b.ownedMode() == BlockSelectorStatistical {
		b.maximumMu.Lock()
		b.Maximum = maximum
		b.maximumMu.Unlock()
	}
}

func (b *BlockSelector) GetMaximum() int {
	b.require()
	b.maximumMu.RLock()
	defer b.maximumMu.RUnlock()
	return b.Maximum
}

func (b *BlockSelector) GetAverageBlockCnt() int64 {
	b.require()
	return b.averageBlockCnt.Load()
}

func (b *BlockSelector) GetBlockSize(size int64, worker *DistributedWorkerSmartProxy) int64 {
	b.require()
	switch b.ownedMode() {
	case BlockSelectorStatic:
		return int64(b.ownedStaticBlockSize())
	case BlockSelectorStatistical:
		if worker != nil {
			limit := b.ownedNetworkOverheadLimit()
			if limit == 0 {
				limit = blockSelectorDefaultNetworkOverheadLimit
			}
			blockSize := math.Abs(math.Ceil(float64(size) * (worker.GetNetworkOverhead() / limit)))
			blockSize = math.Min(math.Max(blockSize, 1), float64(b.GetMaximum()))
			return int64(javaDoubleToInt(blockSize))
		}
		fallthrough
	case BlockSelectorLimiting:
		blockSize := b.proportionalBlockSize(size)
		if blockSize > int64(b.GetMaximum()) {
			return int64(b.GetMaximum())
		}
		return blockSize
	default:
		return b.proportionalBlockSize(size)
	}
}

func (b *BlockSelector) proportionalBlockSize(size int64) int64 {
	b.require()
	if b.ownedServer() == nil {
		panic(NewNullPointerException())
	}
	workerCount := b.ownedServer().GetWorkerCount()
	return javaDoubleToLong(math.Ceil(float64(size) * (1.0 / float64(workerCount))))
}

func (b *BlockSelector) SetAverageBlockCnt(blockCnt int64) {
	b.require()
	if b.ownedMode() == BlockSelectorStatic {
		return
	}
	// Source volatile reads/writes are individually visible; the whole update
	// remains lossy, so do not make it a CAS loop or serialize the calculation.
	if b.averageBlockCnt.Load() > 0 {
		b.averageBlockCnt.Store((blockCnt + b.averageBlockCnt.Load()) / 2)
	} else {
		b.averageBlockCnt.Store(blockCnt)
	}
}

func (b *BlockSelector) require() {
	if b == nil {
		panic(NewNullPointerException())
	}
}

func distributedJavaStateQueue(queue StateQueue) StateQueue {
	if isNil(queue) {
		return nil
	}
	return queue
}

func initializeDistributedWorkerProperties() {
	distributedWorkerSetMode.Do(func() {
		distributedWorkerSetMode.unsorted = distributedBooleanProperty("tlc2.tool.distributed.TLCWorker.unsorted")
	})
}

func NewDistributedWorker(id int, tool *Tool, fpSetManager *DistributedFPSetManager, address ...DistributedWorkerAddress) *DistributedWorker {
	id = int(int32(id))
	if len(address) > 1 {
		panic(NewIllegalArgumentException("NewDistributedWorker accepts at most one address"))
	}
	initializeDistributedWorkerProperties()
	var app *TLCApp
	if tool != nil {
		app = NewTLCApp(tool, true)
	}
	endpoint := DistributedWorkerAddress{Hostname: distributedServerHost()}
	if len(address) > 0 {
		endpoint = address[0]
	}
	worker := &DistributedWorker{
		ID:              id,
		App:             app,
		unsorted:        distributedWorkerSetMode.unsorted,
		FPSetManager:    fpSetManager,
		Cache:           NewSimpleCache(),
		NetworkOverhead: math.MaxFloat64,
		uri:             newDistributedWorkerURI(endpoint, id),
	}
	worker.cacheOwner = worker.Cache
	worker.cacheCaptured = true
	NewDistributedWorkerRuntime(worker)
	return worker
}

func distributedWorkerURI(worker *DistributedWorker) string {
	return worker.GetURI()
}

func (w *DistributedWorker) GetURI() string {
	w.require()
	if w.uri == nil {
		panic(NewNullPointerException())
	}
	return w.uri.raw
}

func (w *DistributedWorker) GetASCIIURI() string {
	w.require()
	if w.uri == nil {
		panic(NewNullPointerException())
	}
	return w.uri.asciiString()
}

type DistributedWorkerSmartProxy struct {
	Worker          DistributedWorkerEndpoint
	NetworkOverhead float64
	workerCaptured  bool
	workerOwner     DistributedWorkerEndpoint
}

func NewDistributedWorkerSmartProxy(worker DistributedWorkerEndpoint) *DistributedWorkerSmartProxy {
	if isNil(worker) {
		worker = nil
	}
	return &DistributedWorkerSmartProxy{Worker: worker, NetworkOverhead: math.MaxFloat64, workerCaptured: true, workerOwner: worker}
}

func (p *DistributedWorkerSmartProxy) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	start := time.Now().UnixMilli()
	worker := p.ownedWorker()
	if isNil(worker) {
		return nil, NewNullPointerException()
	}
	nextStates, err := worker.GetNextStates(states)
	if err != nil {
		return nil, err
	}
	roundTripTime := time.Now().UnixMilli() - start + 1
	if nextStates == nil {
		panic(NewNullPointerException())
	}
	computationTime := sanitizeDistributedComputationTime(nextStates.GetComputationTime())
	networkTime := math.Max(float64(roundTripTime-computationTime), 0.00001)
	percentageNetworkOverhead := networkTime / float64(roundTripTime)
	if states == nil {
		panic(NewNullPointerException())
	}
	p.NetworkOverhead = percentageNetworkOverhead / float64(len(states))
	return nextStates, nil
}

func (p *DistributedWorkerSmartProxy) GetNetworkOverhead() float64 {
	if p == nil {
		panic(NewNullPointerException())
	}
	return p.NetworkOverhead
}

func (p *DistributedWorkerSmartProxy) Exit() error {
	worker := p.ownedWorker()
	if isNil(worker) {
		return NewNullPointerException()
	}
	return worker.Exit()
}

func (p *DistributedWorkerSmartProxy) GetURI() (string, error) {
	worker := p.ownedWorker()
	if isNil(worker) {
		return "", NewNullPointerException()
	}
	return worker.GetURI()
}

func (p *DistributedWorkerSmartProxy) IsAlive() (bool, error) {
	worker := p.ownedWorker()
	if isNil(worker) {
		return false, NewNullPointerException()
	}
	return worker.IsAlive()
}

func (p *DistributedWorkerSmartProxy) GetCacheRateRatio() (float64, error) {
	worker := p.ownedWorker()
	if isNil(worker) {
		return 0, NewNullPointerException()
	}
	return worker.GetCacheRateRatio()
}

func (p *DistributedWorkerSmartProxy) ownedWorker() DistributedWorkerEndpoint {
	if p == nil {
		return nil
	}
	if p.workerCaptured {
		return p.workerOwner
	}
	return p.Worker
}

func sanitizeDistributedComputationTime(computationTime int64) int64 {
	if computationTime < 0 {
		computationTime = -computationTime
	}
	if computationTime < 1 {
		return 1
	}
	return computationTime
}

func (w *DistributedWorker) GetNextStates(states []*TLCStateMut) (result *NextStateResult, err error) {
	if w == nil {
		return nil, NewNullPointerException()
	}
	w.nextStatesMu.Lock()
	defer w.nextStatesMu.Unlock()
	w.Computing.Store(true)
	start := time.Now().UnixMilli()
	w.LastInvocation.Store(start)
	var statesComputed int64
	var state1, state2 *TLCStateMut
	defer w.Computing.Store(false)
	defer func() {
		if failure := recover(); failure != nil {
			result = nil
			err = panicValueAsError(failure)
		}
		if err != nil {
			if failure, ok := err.(*WorkerException); ok && failure != nil {
				return
			}
			if isJavaOutOfMemoryError(err) {
				err = workerComputationFailure("OutOfMemoryError occurred at worker: "+w.GetASCIIURI(), err, true)
				return
			}
			if failure, ok := err.(*RejectedExecutionException); ok && failure != nil {
				err = workerComputationFailure("Executor rejected task at worker: "+w.GetASCIIURI(), err, false)
				return
			}
			err = newWorkerExceptionFromThrowable(err, state1, state2, true)
		}
	}()

	if states == nil {
		panic(NewNullPointerException())
	}
	holdersByFP := make(map[uint64]DistributedStateHolder, len(states))
	orderedFPs := make([]uint64, 0, len(states))
	identityHolders := NewInsMap[*DistributedStateHolder, bool]()
	for _, state := range states {
		state1 = state
		nextStates, err := w.computeNextStates(state)
		if err != nil {
			return nil, err
		}
		statesComputed += int64(nextStates.Size())
		for i := 0; i < nextStates.Size(); i++ {
			successor := nextStates.At(i)
			fp := successor.FingerPrint()
			cache := w.ownedCache()
			if cache == nil {
				panic(NewNullPointerException())
			}
			if cache.Hit(fp) {
				continue
			}
			if w.unsorted {
				// Holder inherits Object.equals/hashCode. Distinct holder objects
				// remain distinct even when their fingerprints match. InsMap gives
				// deterministic iteration in place of Java's VM-dependent HashSet.
				identityHolders.Set(NewDistributedStateHolder(fp, successor, state1), true)
				continue
			}
			if _, ok := holdersByFP[fp]; ok {
				continue
			}
			holdersByFP[fp] = *NewDistributedStateHolder(fp, successor, state1)
			orderedFPs = append(orderedFPs, fp)
		}
	}
	w.OverallStatesComputed.Add(statesComputed)

	sort.Slice(orderedFPs, func(i, j int) bool { return int64(orderedFPs[i]) < int64(orderedFPs[j]) })
	holders := make([]DistributedStateHolder, 0, len(orderedFPs))
	if w.unsorted {
		for holder := range identityHolders.All() {
			holders = append(holders, *holder)
		}
	} else {
		for _, fp := range orderedFPs {
			holders = append(holders, holdersByFP[fp])
		}
	}
	if w.FPSetManager == nil {
		panic(NewNullPointerException())
	}
	serverCount := w.FPSetManager.NumOfServers()
	predecessors := make([]*StateVec, serverCount)
	successors := make([]*StateVec, serverCount)
	fingerprints := make([]*LongVec, serverCount)
	for i := 0; i < serverCount; i++ {
		predecessors[i] = NewDefaultDistributedStateVec()
		successors[i] = NewDefaultDistributedStateVec()
		fingerprints[i] = NewLongVec()
	}
	last := int64(math.MinInt64)
	for _, holder := range holders {
		fp := holder.GetFP()
		if int64(fp) <= last {
			failure := newTLCErrorCode(ECGeneral)
			failure.Runtime = true
			panic(failure)
		}
		last = int64(fp)
		fpIndex := w.FPSetManager.GetFPSetIndex(fp)
		predecessors[fpIndex].Add(holder.GetParentState())
		successors[fpIndex].Add(holder.GetNewState())
		fingerprints[fpIndex].AddElement(int64(fp))
	}

	if w.Runtime == nil {
		panic(NewNullPointerException())
	}
	visited := w.FPSetManager.ContainsBlock(fingerprints, &w.Runtime.executor)
	newStates := make([]*StateVec, serverCount)
	newFingerprints := make([]*LongVec, serverCount)
	for i := 0; i < serverCount; i++ {
		newStates[i] = NewDefaultDistributedStateVec()
		newFingerprints[i] = NewLongVec()
	}
	for i := 0; i < serverCount; i++ {
		iter := NewBitVectorIter(visited[i])
		for {
			index := iter.Next()
			if index == -1 {
				break
			}
			state1 = predecessors[i].At(index)
			state2 = successors[i].At(index)
			if err := w.CheckState(state1, state2); err != nil {
				return nil, err
			}
			inModel, err := w.IsInModel(state2)
			if err != nil {
				return nil, err
			}
			if !inModel {
				continue
			}
			inActions, err := w.IsInActions(state1, state2)
			if err != nil {
				return nil, err
			}
			if inModel && inActions {
				state2.UID = state1.UID
				newStates[i].Add(state2)
				newFingerprints[i].AddElement(fingerprints[i].ElementAt(index))
			}
		}
	}

	elapsed := time.Now().UnixMilli() - start
	return NewNextStateResult(newStates, newFingerprints, elapsed, statesComputed), nil
}

func (w *DistributedWorker) computeNextStates(state *TLCStateMut) (*StateVec, error) {
	w.require()
	if w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.GetNextStates(state)
}

func (w *DistributedWorker) CheckState(predecessor *TLCStateMut, successor *TLCStateMut) error {
	w.require()
	if w.CheckStateFunc != nil {
		return w.CheckStateFunc(predecessor, successor)
	}
	if w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.CheckState(predecessor, successor)
}

func (w *DistributedWorker) IsInModel(state *TLCStateMut) (bool, error) {
	w.require()
	if w.IsInModelFunc != nil {
		return w.IsInModelFunc(state)
	}
	if w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.IsInModel(state)
}

func (w *DistributedWorker) IsInActions(predecessor *TLCStateMut, successor *TLCStateMut) (bool, error) {
	w.require()
	if w.IsInActionsFunc != nil {
		return w.IsInActionsFunc(predecessor, successor)
	}
	if w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.IsInActions(predecessor, successor)
}

func (w *DistributedWorker) IsAlive() bool {
	w.require()
	return true
}

func (w *DistributedWorker) GetCacheRateRatio() float64 {
	w.require()
	cache := w.ownedCache()
	if cache == nil {
		panic(NewNullPointerException())
	}
	return cache.GetHitRatio()
}

func (w *DistributedWorker) ownedCache() *SimpleCache {
	w.require()
	if w.cacheCaptured {
		return w.cacheOwner
	}
	return w.Cache
}

func (w *DistributedWorker) IsComputing() bool {
	w.require()
	return w.Computing.Load()
}

func (w *DistributedWorker) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

type DistributedStateHolder struct {
	fingerprint uint64
	successor   *TLCStateMut
	predecessor *TLCStateMut
}

func NewDistributedStateHolder(fingerprint uint64, successor, predecessor *TLCStateMut) *DistributedStateHolder {
	return &DistributedStateHolder{fingerprint: fingerprint, successor: successor, predecessor: predecessor}
}

func (h *DistributedStateHolder) GetFP() uint64 {
	if h == nil {
		panic(NewNullPointerException())
	}
	return h.fingerprint
}

func (h *DistributedStateHolder) GetNewState() *TLCStateMut {
	if h == nil {
		panic(NewNullPointerException())
	}
	return h.successor
}

func (h *DistributedStateHolder) GetParentState() *TLCStateMut {
	if h == nil {
		panic(NewNullPointerException())
	}
	return h.predecessor
}

func (h *DistributedStateHolder) CompareTo(other *DistributedStateHolder) int {
	if h == nil || other == nil {
		panic(NewNullPointerException())
	}
	left, right := int64(h.fingerprint), int64(other.fingerprint)
	if left < right {
		return -1
	}
	if left == right {
		return 0
	}
	return 1
}

var distributedWorkerSetMode struct {
	sync.Once
	unsorted bool
}

func newAllTrueBitVector(size int) *BitVector {
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		bv.Set(i)
	}
	return bv
}
