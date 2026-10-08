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
	TLCServerThreadNamePrefix = "TLCServerThread-"
	TLCWorkerThreadNamePrefix = "TLCWorkerThread-"
	TLCServerDefaultPort      = 10997
	tlcServerVetoCleanup      = "tlc2.tool.distributed.TLCServer.vetoCleanup"
)

type NextStateResult struct {
	ComputationTime  int64
	StatesComputed   int64
	NextStates       []*StateVec
	NextFingerprints []*LongVec
}

func NewNextStateResult(nextStates []*StateVec, nextFingerprints []*LongVec, computationTime int64, statesComputed int64) *NextStateResult {
	return &NextStateResult{
		NextStates:       nextStates,
		NextFingerprints: nextFingerprints,
		ComputationTime:  computationTime,
		StatesComputed:   statesComputed,
	}
}

func (r *NextStateResult) GetStatesComputedDelta() int64 {
	if r == nil || r.NextStates == nil {
		panic(NewNullPointerException())
	}
	return r.StatesComputed - int64(len(r.NextStates))
}

func (r *NextStateResult) GetComputationTime() int64 {
	if r == nil {
		panic(NewNullPointerException())
	}
	return r.ComputationTime
}

func (r *NextStateResult) GetNextFingerprints() []*LongVec {
	if r == nil {
		panic(NewNullPointerException())
	}
	return r.NextFingerprints
}

func (r *NextStateResult) GetNextStates() []*StateVec {
	if r == nil {
		panic(NewNullPointerException())
	}
	return r.NextStates
}

type DistributedWorker struct {
	nextStatesMu          sync.Mutex
	ID                    int
	App                   *TLCApp
	unsorted              bool
	FPSetManager          *DistributedFPSetManager
	Cache                 *SimpleCache
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
	AverageBlockCnt             int64
	monitor                     distributedServerMonitor
	completionWaiter            chan struct{}
	threadsMu                   sync.Mutex
	threadsToWorkers            *InsMap[*TLCServerThread, DistributedWorkerEndpoint]
	executor                    DistributedExecutor
	BlockSelector               *BlockSelector
	FinalNumberOfDistinctStates int64
}

func NewTLCServer(fileName string, configName string, metadir string, manager *DistributedFPSetManager, queue StateQueue, trace *TLCTrace) *TLCServer {
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
	server.BlockSelector = NewBlockSelectorFromProperties(server)
	return server
}

func (s *TLCServer) SetTool(tool *Tool) *TLCServer {
	if s != nil {
		s.Tool = tool
		s.app = nil
		if tool != nil {
			s.app = NewTLCApp(tool, s.GetCheckDeadlock())
		}
		s.serverInternTable()
		if s.Trace != nil {
			s.Trace.SetTool(tool)
		}
	}
	return s
}

func (s *TLCServer) HasNoErrors() bool {
	if s == nil {
		return false
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	return s.ErrState == nil
}

func (s *TLCServer) application() *TLCApp {
	if s.app == nil {
		s.app = NewTLCApp(s.Tool, s.GetCheckDeadlock())
	}
	return s.app
}

func (s *TLCServer) SetCheckDeadlock(check bool) *TLCServer {
	s.checkDeadlock = &check
	if s.app != nil {
		s.app.checkDeadlock = check
	}
	return s
}

func (s *TLCServer) GetCheckDeadlock() bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.checkDeadlock != nil {
		return *s.checkDeadlock
	}
	// TLCApp uses the command-line flag, independently of CHECK_DEADLOCK in
	// ModelConfig. Its command-line default is true.
	return true
}

func (s *TLCServer) GetPreprocess() bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	return true // TLCApp's final preprocess field.
}

func (s *TLCServer) GetIrredPolyForFP() uint64 {
	if s == nil {
		panic(NewNullPointerException())
	}
	return FP64IrredPoly()
}

func (s *TLCServer) Checkpoint() error {
	if s == nil || s.StateQueue == nil {
		return nil
	}
	if !s.StateQueue.SuspendAll() {
		return nil
	}
	PrintMessage(ECTLCCheckpointStart, "-- Checkpointing of run "+s.Metadir+" compl")
	if err := s.StateQueue.BeginChkpt(); err != nil {
		return err
	}
	if s.Trace != nil {
		if err := s.Trace.BeginChkpt(); err != nil {
			return err
		}
	}
	if s.FPSetManager != nil {
		if err := s.FPSetManager.Checkpoint(s.checkpointFileName()); err != nil {
			return err
		}
	}
	s.StateQueue.ResumeAll()
	if err := s.serverInternTable().beginChkptWithVarCount(s.Metadir, s.serverInternTable().varCount); err != nil {
		return err
	}
	if err := s.StateQueue.CommitChkpt(); err != nil {
		return err
	}
	if s.Trace != nil {
		if err := s.Trace.CommitChkpt(); err != nil {
			return err
		}
	}
	if err := s.serverInternTable().CommitChkpt(s.Metadir); err != nil {
		return err
	}
	if s.FPSetManager != nil {
		if err := s.FPSetManager.CommitCheckpoint(); err != nil {
			return err
		}
	}
	PrintMessage(ECTLCCheckpointEnd, "eted.")
	return nil
}

func (s *TLCServer) Recover() error {
	if s == nil {
		return nil
	}
	if s.Trace != nil {
		if err := s.Trace.Recover(); err != nil {
			return err
		}
	}
	if s.StateQueue != nil {
		if err := s.StateQueue.Recover(); err != nil {
			return err
		}
	}
	if s.FPSetManager != nil {
		if err := s.FPSetManager.Recover(s.checkpointFileName()); err != nil {
			return err
		}
	}
	return nil
}

func (s *TLCServer) Close(cleanup bool) error {
	if s == nil {
		return nil
	}
	if s.Trace != nil {
		if err := s.Trace.Close(); err != nil {
			return err
		}
	}
	if s.FPSetManager != nil {
		if err := s.FPSetManager.Close(cleanup); err != nil {
			return err
		}
	}
	if cleanup && !distributedVetoCleanup() {
		deleteDirLikeJava(s.Metadir, true)
	}
	return nil
}

func (s *TLCServer) DoInit(tool ...*Tool) (int, error) {
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
		PrintMessage(ECTLCCheckpointRecoverStart, s.Metadir)
		if err := s.Recover(); err != nil {
			return ECGeneral, err // Recovery is outside the init catch(Throwable).
		}
		PrintMessage(ECTLCCheckpointRecoverEnd, fmtUint64(s.fpSetSize()), fmtInt64(s.StateQueue.Size()))
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
		PrintMessage(ECTLCComputingInit)
		if err := s.runDistributedInit(); err != nil {
			s.SetDone()
			result = s.reportDistributedInitFailure(app, err)
		} else {
			PrintMessage(ECTLCInitGenerated1, fmtInt64(s.StateQueue.Size()), "(s)")
		}
	}
	if s.IsDone() {
		s.PrintSummary(1, 0, s.StateQueue.Size(), s.fpSetSize(), false)
		PrintMessage(ECTLCFinished, humanReadableTLCRuntime(time.Since(startTime)))
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
	if s.HasNoErrors() && !s.StateQueue.IsEmpty() {
		return ECGeneral, NewTLCRuntimeException(ECGeneral)
	}
	for _, thread := range s.GetServerThreads() {
		if thread == nil {
			continue
		}
		thread.Join()
		cacheRatio := "n/a"
		if thread.GetCacheRateRatio() >= 0 {
			cacheRatio = groupDecimalIntegerPart(fmt.Sprintf("%.2f", thread.GetCacheRateRatio()))
		}
		PrintMessage(ECTLCDistributedWorkerStats,
			thread.GetURI(),
			fmtInt(thread.GetSentStates()),
			fmtInt(thread.GetReceivedStates()),
			cacheRatio,
		)
		exitErr := func() error {
			defer s.removeServerThreadOnly(thread)
			if err := thread.Worker.Exit(); err != nil {
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
	level := 1
	if s.Trace != nil {
		level = s.Trace.GetLevelForReporting()
	}
	s.monitor.Lock()
	s.StatesPerMinute = 0
	s.DistinctStatesPerMinute = 0
	s.monitor.Unlock()
	if s.HasNoErrors() {
		actualDistance := s.FPSetManager.CheckFPs()
		statesSeen := s.FPSetManager.GetStatesSeen()
		ReportSuccessCountsDistance(uint64(TLCServerFinalNumberOfDistinctStates()), actualDistance, int64(statesSeen))
	} else if s.KeepCallStack {
		app.SetCallStack()
		s.Tool = app.Tool
		if s.Trace != nil {
			s.Trace.SetTool(app.Tool)
		}
	}
	s.PrintSummary(level, statesGenerated, statesLeft, uint64(TLCServerFinalNumberOfDistinctStates()), s.HasNoErrors())
	PrintMessage(ECTLCFinished, humanReadableTLCRuntime(time.Since(startTime)))
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
	if s == nil {
		return
	}
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
	PrintMessage(ECTLCInitGenerated2, fmtInt64(statesGenerated), plural, fmtUint64(distinct))
}

func (s *TLCServer) PrintSummary(level int, statesGenerated int64, statesLeftInQueue int64, distinctStates uint64, success bool) {
	if toolMode() {
		PrintMessage(ECTLCProgressStats, fmtInt(level), groupDecimalIntegerPart(fmtInt64(statesGenerated)), groupDecimalIntegerPart(fmtInt64(int64(distinctStates))), groupDecimalIntegerPart(fmtInt64(statesLeftInQueue)), "0", "0")
	}
	PrintMessage(ECTLCStats, fmtInt64(statesGenerated), fmtInt64(int64(distinctStates)), fmtInt64(statesLeftInQueue))
	if success {
		PrintMessage(ECTLCSearchDepth, fmtInt(level))
	}
}

func (s *TLCServer) ReportSuccess() {
	if s == nil {
		ReportSuccessCountsDistance(0, 0, 0)
		return
	}
	distinct := s.fpSetSize()
	actualDistance := uint64(0)
	statesSeen := uint64(0)
	if s.FPSetManager != nil {
		actualDistance = s.FPSetManager.CheckFPs()
		statesSeen = s.FPSetManager.GetStatesSeen()
	}
	ReportSuccessCountsDistance(distinct, actualDistance, int64(statesSeen))
}

func (s *TLCServer) fpSetSize() uint64 {
	if s == nil || s.FPSetManager == nil {
		return 0
	}
	return s.FPSetManager.Size()
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
	if s.StateQueue != nil {
		s.StateQueue.ResumeAllStuck()
	}
	if worker == nil {
		return NewNullPointerException()
	}
	uri, err := worker.GetURI()
	if err != nil {
		return err
	}
	thread := NewTLCServerThread(worker, uri, s, s.BlockSelector)
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
	if s == nil || thread == nil {
		return
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
	if s == nil || thread == nil {
		return nil
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
	if s == nil {
		return nil
	}
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
	if s != nil {
		s.Done.Store(true)
	}
}

func (s *TLCServer) SetErrState(curState *TLCStateMut, succState *TLCStateMut, keepCallStack bool, errorCode ...int) bool {
	if s == nil {
		return false
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
		s.ErrorCode = errorCode[0]
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
		seen = f.server.FPSetManager.Put(fp)
		if !seen {
			if _, err := f.server.Trace.WriteState(nil, fp, curState, nil); err != nil {
				return curState, err
			}
			f.server.StateQueue.Enqueue(curState)
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
	return s != nil && !s.Done.Load()
}

func (s *TLCServer) IsDone() bool {
	return s == nil || s.Done.Load()
}

func (s *TLCServer) AddStatesGeneratedDelta(delta int64) {
	if s != nil && delta != 0 {
		s.WorkerStatesGenerated.Add(delta)
	}
}

func (s *TLCServer) GetStatesGenerated() int64 {
	if s == nil {
		return 0
	}
	total := s.WorkerStatesGenerated.Load()
	if s.FPSetManager != nil {
		total += int64(s.FPSetManager.GetStatesSeen())
	}
	return total
}

func (s *TLCServer) GetNewStates() int64 {
	if s == nil {
		return 0
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	return s.getNewStatesLocked()
}

func (s *TLCServer) getNewStatesLocked() int64 {
	var size int64
	if s.StateQueue != nil {
		size = s.StateQueue.Size()
	}
	for _, thread := range s.GetServerThreads() {
		size += int64(thread.GetCurrentSize())
	}
	return size
}

func (s *TLCServer) GetStatesGeneratedPerMinute() int64 {
	if s == nil {
		return 0
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	return s.StatesPerMinute
}

func (s *TLCServer) GetDistinctStatesGeneratedPerMinute() int64 {
	if s == nil {
		return 0
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	return s.DistinctStatesPerMinute
}

func (s *TLCServer) GetAverageBlockCnt() int64 {
	if s == nil {
		return 0
	}
	if s.BlockSelector != nil {
		return s.BlockSelector.GetAverageBlockCnt()
	}
	return s.AverageBlockCnt
}

func (s *TLCServer) GetWorkerCount() int {
	if s == nil {
		return 0
	}
	s.threadsMu.Lock()
	defer s.threadsMu.Unlock()
	if s.threadsToWorkers == nil {
		return 0
	}
	return s.threadsToWorkers.Len()
}

func (s *TLCServer) GetFPSetManager() *DistributedFPSetManager {
	if s == nil {
		return nil
	}
	if s.fpRegistration != nil {
		<-s.fpRegistration.done
	}
	return s.FPSetManager
}

func (s *TLCServer) GetSpecFileName() string {
	if s == nil {
		return ""
	}
	return s.FileName
}

func (s *TLCServer) GetConfigFileName() string {
	if s == nil {
		return ""
	}
	return s.ConfigName
}

var tlcServerThreadCount atomic.Int64

type distributedStateBlock struct {
	states []*TLCStateMut
}

type TLCServerThread struct {
	ID                int
	ReceivedStates    int
	SentStates        int
	CacheRateHitRatio float64
	Selector          *BlockSelector
	states            atomic.Pointer[distributedStateBlock]
	TimerTask         *TLCTimerTask
	Worker            *DistributedWorkerSmartProxy
	Server            *TLCServer
	URI               string
	cleanupGlobals    atomic.Bool
	started           atomic.Bool
	keepAliveStopped  atomic.Bool
	keepAliveDone     chan struct{}
	runDone           chan struct{}
}

func NewTLCServerThread(worker DistributedWorkerEndpoint, uri string, server *TLCServer, selector *BlockSelector) *TLCServerThread {
	if selector == nil && server != nil {
		selector = server.BlockSelector
	}
	if selector == nil {
		selector = NewBlockSelectorFromProperties(server)
	}
	thread := &TLCServerThread{
		ID:                int(tlcServerThreadCount.Add(1) - 1),
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
	thread.TimerTask = &TLCTimerTask{Thread: thread}
	// Java schedules keepalive during construction, before Thread.start().
	thread.startKeepAlive()
	if server != nil {
		server.RegisterTLCServerThread(thread)
	}
	return thread
}

func (t *TLCServerThread) Name() string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%s%03d-[%s]", TLCServerThreadNamePrefix, t.ID, t.URI)
}

func (t *TLCServerThread) Start() {
	if t == nil || !t.started.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer close(t.runDone)
		t.Run()
	}()
}

func (t *TLCServerThread) Join() {
	if t == nil || t.runDone == nil {
		return
	}
	<-t.runDone
}

func (t *TLCServerThread) startKeepAlive() {
	if t == nil || t.TimerTask == nil || t.keepAliveDone == nil {
		return
	}
	go t.runKeepAlive()
}

func (t *TLCServerThread) runKeepAlive() {
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
			t.TimerTask.Run()
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
	if t == nil || t.Server == nil {
		return
	}
	IncNumWorkers(1)
	stateQueue := t.Server.StateQueue
	defer func() {
		if recovered := recover(); recovered != nil {
			t.handleRunError(panicValueAsError(recovered), stateQueue)
		}
		t.readCacheRateRatio()
		t.cancelKeepAlive()
		t.setStates([]*TLCStateMut{})
	}()
	for {
		if t.Selector == nil {
			t.Selector = t.Server.BlockSelector
		}
		if t.Selector == nil {
			t.Selector = NewBlockSelectorFromProperties(t.Server)
		}
		t.setStates(t.Selector.GetBlocks(stateQueue, t.Worker))
		if t.currentStates() == nil {
			t.Server.monitor.Lock()
			t.Server.SetDone()
			t.Server.notifyCompletionLocked()
			t.Server.monitor.Unlock()
			if stateQueue != nil {
				stateQueue.FinishAll()
			}
			return
		}
		if len(t.currentStates()) == 0 {
			continue
		}
		t.SentStates = int(int32(t.SentStates) + int32(len(t.currentStates())))

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
			t.handleRunError(err, stateQueue)
			return
		}
	}
}

// Java's inner remote/NPE catch includes result dereferences, statistics,
// the keepalive timestamp and generated-state delta, before publishing the block.
func (t *TLCServerThread) computeBlockAttempt() (res *NextStateResult, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			res = nil
			err = panicValueAsError(failure)
		}
	}()
	res, err = invokeDistributedWorker(t.Worker, t.currentStates())
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
	t.ReceivedStates = int(int32(t.ReceivedStates) + int32(newStates[0].Size()))
	if t.TimerTask != nil {
		t.TimerTask.SetLastInvocation(time.Now())
	}
	t.Server.AddStatesGeneratedDelta(res.GetStatesComputedDelta())
	return res, nil
}

func (t *TLCServerThread) computeBlock(stateQueue StateQueue) (*NextStateResult, bool) {
	res, err := t.computeBlockAttempt()
	if err == nil {
		return res, true
	}
	if isDistributedRemoteFailure(err) {
		if isRecoverableDistributedError(err) && len(t.currentStates()) > 1 {
			PrintMessage(ECTLCDistributedExceedBlocksize, fmtInt(len(t.currentStates())/2))
			if stateQueue != nil {
				stateQueue.SEnqueueAll(t.currentStates())
			}
			if t.Selector != nil {
				t.Selector.SetMaxTXSize(len(t.currentStates()) / 2)
			}
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
	t.handleRunError(err, stateQueue)
	return nil, false
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
	if t == nil || t.Server == nil || t.Server.FPSetManager == nil {
		panic(NewNullPointerException())
	}
	visited := t.Server.FPSetManager.PutBlock(newFps, &t.Server.executor)
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
			if index >= len(newStates[i].states) {
				panic(NewArrayIndexOutOfBoundsException(index, len(newStates[i].states)))
			}
			state := newStates[i].At(index)
			fingerprints := fpBlockAt(newFps, i)
			if fingerprints == nil {
				panic(NewNullPointerException())
			}
			fp := uint64(fingerprints.ElementAt(index))
			if state == nil || t.Server.Trace == nil {
				panic(NewNullPointerException())
			}
			// TLCWorker returns a successor whose UID still identifies its
			// predecessor. Write that UID before replacing the trace location.
			if _, err := t.Server.Trace.WriteStateRecord(state, fp, state); err != nil {
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
	if t == nil || t.Server == nil {
		return
	}
	var state1, state2 *TLCStateMut
	if failure, ok := err.(*WorkerException); ok && failure != nil {
		state1, state2 = failure.State1, failure.State2
	}
	if t.Server.SetErrState(state1, nil, true, ECGeneral) {
		t.Server.LastError = err
		if state1 != nil && t.Server.Trace != nil {
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
				t.Server.Trace.PrintTrace(state1, state2)
			}()
		} else {
			PrintError(ECGeneral, generalErrorParams("", err)...)
		}
		if stateQueue != nil {
			stateQueue.FinishAll()
		}
		t.Server.monitor.Lock()
		t.Server.notifyCompletionLocked()
		t.Server.monitor.Unlock()
	}
}

func (t *TLCServerThread) HandleRemoteWorkerLost(stateQueue StateQueue) {
	if t == nil {
		return
	}
	t.cancelKeepAlive()
	if !t.cleanupGlobals.CompareAndSwap(true, false) {
		return
	}
	if t.Server != nil {
		t.Server.RemoveTLCServerThread(t)
	}
	if stateQueue != nil {
		stateQueue.SEnqueueAll(t.currentStates())
	}
	t.setStates([]*TLCStateMut{})
	if stateQueue != nil {
		stateQueue.ResumeAllStuck()
	}
	DecNumWorkers()
}

func (t *TLCServerThread) setStates(states []*TLCStateMut) {
	t.states.Store(&distributedStateBlock{states: states})
}

func (t *TLCServerThread) currentStates() []*TLCStateMut {
	block := t.states.Load()
	if block == nil {
		return nil
	}
	return block.states
}

func (t *TLCServerThread) GetCurrentSize() int {
	if t == nil {
		return 0
	}
	states := t.currentStates()
	if states == nil {
		panic(NewNullPointerException())
	}
	return len(states)
}

func (t *TLCServerThread) GetURI() string {
	if t == nil {
		return ""
	}
	return t.URI
}

func (t *TLCServerThread) GetReceivedStates() int {
	if t == nil {
		return 0
	}
	return t.ReceivedStates
}

func (t *TLCServerThread) GetSentStates() int {
	if t == nil {
		return 0
	}
	return t.SentStates
}

func (t *TLCServerThread) GetCacheRateRatio() float64 {
	if t == nil {
		return -1
	}
	return t.CacheRateHitRatio
}

func (t *TLCServerThread) readCacheRateRatio() {
	if t == nil || t.Worker == nil {
		return
	}
	ratio, err := t.Worker.GetCacheRateRatio()
	if err != nil {
		if !isDistributedRemoteFailure(err) {
			panic(err)
		}
		PrintWarning(ECGeneral, "Failed to read remote worker cache statistic (Expect to see a negative chache hit rate. Does not invalidate model checking results)")
		return
	}
	t.CacheRateHitRatio = ratio
}

func (t *TLCServerThread) cancelKeepAlive() {
	if t == nil || t.keepAliveDone == nil {
		return
	}
	if t.keepAliveStopped.CompareAndSwap(false, true) {
		close(t.keepAliveDone)
	}
}

type TLCTimerTask struct {
	Thread         *TLCServerThread
	LastInvocation atomic.Int64
}

func (t *TLCTimerTask) Run() {
	if t == nil || t.Thread == nil {
		return
	}
	now := time.Now().UnixMilli()
	last := t.LastInvocation.Load()
	if last == 0 || now-last > int64(time.Minute/time.Millisecond) {
		alive, err := t.Thread.Worker.IsAlive()
		if err != nil && !isDistributedRemoteFailure(err) {
			panic(err)
		}
		if err != nil || !alive {
			queue := StateQueue(nil)
			if t.Thread.Server != nil {
				queue = t.Thread.Server.StateQueue
			}
			t.Thread.HandleRemoteWorkerLost(queue)
		}
	}
}

func (t *TLCTimerTask) SetLastInvocation(when time.Time) {
	if t == nil {
		return
	}
	t.LastInvocation.Store(when.UnixMilli())
}

// Java TLCServerThread.isRecoverable inspects exactly one cause, then the
// direct cause of a nested RemoteException. It does not search a cause chain.
func isRecoverableDistributedError(err error) bool {
	if failure, ok := err.(*DistributedOperationError); ok {
		return failure.Remote && failure.Recoverable
	}
	remote := javaRemoteException(err)
	if remote == nil {
		return false
	}
	cause := remote.GetCause()
	if eof, ok := cause.(*EOFException); ok && eof != nil && eof.GetMessage() == nil {
		return true
	}
	if nested := javaRemoteException(cause); nested != nil {
		return isJavaOutOfMemoryError(nested.GetCause())
	}
	return false
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
	Server               *TLCServer
	Mode                 BlockSelectorMode
	Maximum              int
	StaticBlockSize      int
	NetworkOverheadLimit float64
	averageBlockCnt      atomic.Int64
	maximumMu            sync.RWMutex
}

func NewBlockSelector(server *TLCServer) *BlockSelector {
	return NewBlockSelectorFromProperties(server)
}

func NewBlockSelectorFromProperties(server *TLCServer) *BlockSelector {
	switch {
	case distributedBooleanProperty(distributedSelectorStaticProperty):
		return NewStaticBlockSelector(server)
	case distributedBooleanProperty(distributedSelectorUnlimitingProperty):
		return NewProportionalBlockSelector(server)
	case distributedBooleanProperty(distributedSelectorLimitingProperty):
		return NewLimitingBlockSelector(server)
	default:
		return NewStatisticalBlockSelector(server)
	}
}

func NewProportionalBlockSelector(server *TLCServer) *BlockSelector {
	if server == nil {
		panic(NewTLCRuntimeExceptionMessage("TLC found a null TLCServer"))
	}
	return &BlockSelector{
		Server:               server,
		Mode:                 BlockSelectorProportional,
		Maximum:              blockSelectorDefaultMaximum,
		StaticBlockSize:      blockSelectorDefaultStaticSize,
		NetworkOverheadLimit: blockSelectorDefaultNetworkOverheadLimit,
	}
}

func NewLimitingBlockSelector(server *TLCServer, maximum ...int) *BlockSelector {
	selector := NewProportionalBlockSelector(server)
	selector.Mode = BlockSelectorLimiting
	if len(maximum) > 0 {
		selector.Maximum = maximum[0]
	}
	return selector
}

func NewStatisticalBlockSelector(server *TLCServer) *BlockSelector {
	selector := NewLimitingBlockSelector(server)
	selector.Mode = BlockSelectorStatistical
	return selector
}

func NewStaticBlockSelector(server *TLCServer, blockSize ...int) *BlockSelector {
	selector := NewProportionalBlockSelector(server)
	selector.Mode = BlockSelectorStatic
	if len(blockSize) > 0 {
		selector.StaticBlockSize = blockSize[0]
	} else if value, ok := distributedIntProperty(distributedStaticBlockSizeProperty); ok {
		selector.StaticBlockSize = value
	}
	selector.averageBlockCnt.Store(int64(selector.StaticBlockSize))
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
	if b == nil || stateQueue == nil {
		return nil
	}
	if b.Mode == BlockSelectorStatic {
		return stateQueue.SDequeueMany(b.StaticBlockSize)
	}
	amountOfStates := stateQueue.Size()
	blockSize := b.getBlockSize(amountOfStates, worker)
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
	b.setAverageBlockCnt(int64(len(states)))
	return states
}

func (b *BlockSelector) SetMaxTXSize(maximum int) {
	if b != nil && (b.Mode == BlockSelectorLimiting || b.Mode == BlockSelectorStatistical) {
		b.maximumMu.Lock()
		b.Maximum = maximum
		b.maximumMu.Unlock()
	}
}

func (b *BlockSelector) getMaximum() int {
	b.maximumMu.RLock()
	defer b.maximumMu.RUnlock()
	return b.Maximum
}

func (b *BlockSelector) GetAverageBlockCnt() int64 {
	if b == nil {
		return 0
	}
	return b.averageBlockCnt.Load()
}

func (b *BlockSelector) getBlockSize(size int64, worker *DistributedWorkerSmartProxy) int64 {
	if b == nil {
		return 1
	}
	switch b.Mode {
	case BlockSelectorStatic:
		return int64(b.StaticBlockSize)
	case BlockSelectorStatistical:
		if worker != nil {
			limit := b.NetworkOverheadLimit
			if limit == 0 {
				limit = blockSelectorDefaultNetworkOverheadLimit
			}
			blockSize := math.Abs(math.Ceil(float64(size) * (worker.GetNetworkOverhead() / limit)))
			blockSize = math.Min(math.Max(blockSize, 1), float64(b.getMaximum()))
			return int64(javaDoubleToInt(blockSize))
		}
		fallthrough
	case BlockSelectorLimiting:
		blockSize := b.proportionalBlockSize(size)
		if blockSize > int64(b.getMaximum()) {
			return int64(b.getMaximum())
		}
		return blockSize
	default:
		return b.proportionalBlockSize(size)
	}
}

func (b *BlockSelector) proportionalBlockSize(size int64) int64 {
	if b == nil || b.Server == nil {
		panic(NewNullPointerException())
	}
	workerCount := b.Server.GetWorkerCount()
	return javaDoubleToLong(math.Ceil(float64(size) * (1.0 / float64(workerCount))))
}

func (b *BlockSelector) setAverageBlockCnt(blockCnt int64) {
	if b == nil || b.Mode == BlockSelectorStatic {
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

func initializeDistributedWorkerProperties() {
	distributedWorkerSetMode.Do(func() {
		distributedWorkerSetMode.unsorted = distributedBooleanProperty("tlc2.tool.distributed.TLCWorker.unsorted")
	})
}

func NewDistributedWorker(id int, tool *Tool, fpSetManager *DistributedFPSetManager, address ...DistributedWorkerAddress) *DistributedWorker {
	initializeDistributedWorkerProperties()
	if fpSetManager == nil {
		fpSetManager = NewDistributedFPSetManager()
	}
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
	NewDistributedWorkerRuntime(worker)
	return worker
}

func distributedWorkerURI(worker *DistributedWorker) string {
	return worker.GetURI()
}

func (w *DistributedWorker) GetURI() string {
	if w == nil || w.uri == nil {
		panic(NewNullPointerException())
	}
	return w.uri.raw
}

func (w *DistributedWorker) GetASCIIURI() string {
	if w == nil || w.uri == nil {
		panic(NewNullPointerException())
	}
	return w.uri.asciiString()
}

type DistributedWorkerSmartProxy struct {
	Worker          DistributedWorkerEndpoint
	NetworkOverhead float64
}

func NewDistributedWorkerSmartProxy(worker DistributedWorkerEndpoint) *DistributedWorkerSmartProxy {
	return &DistributedWorkerSmartProxy{Worker: worker, NetworkOverhead: math.MaxFloat64}
}

func (p *DistributedWorkerSmartProxy) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	start := time.Now().UnixMilli()
	if p == nil || p.Worker == nil {
		return nil, NewNullPointerException()
	}
	nextStates, err := p.Worker.GetNextStates(states)
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
		return math.MaxFloat64
	}
	return p.NetworkOverhead
}

func (p *DistributedWorkerSmartProxy) Exit() error {
	if p == nil || p.Worker == nil {
		return NewNullPointerException()
	}
	return p.Worker.Exit()
}

func (p *DistributedWorkerSmartProxy) GetURI() (string, error) {
	if p == nil || p.Worker == nil {
		return "", NewNullPointerException()
	}
	return p.Worker.GetURI()
}

func (p *DistributedWorkerSmartProxy) IsAlive() (bool, error) {
	if p == nil || p.Worker == nil {
		return false, NewNullPointerException()
	}
	return p.Worker.IsAlive()
}

func (p *DistributedWorkerSmartProxy) GetCacheRateRatio() (float64, error) {
	if p == nil || p.Worker == nil {
		return 0, NewNullPointerException()
	}
	return p.Worker.GetCacheRateRatio()
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
				err = NewRemoteException(javaString("OutOfMemoryError occurred at worker: "+w.GetASCIIURI()), err)
				return
			}
			if failure, ok := err.(*RejectedExecutionException); ok && failure != nil {
				err = NewRemoteException(javaString("Executor rejected task at worker: "+w.GetASCIIURI()), err)
				return
			}
			err = newWorkerExceptionFromThrowable(err, state1, state2, true)
		}
	}()

	if states == nil {
		panic(NewNullPointerException())
	}
	holdersByFP := make(map[uint64]distributedStateHolder, len(states))
	orderedFPs := make([]uint64, 0, len(states))
	identityHolders := NewInsMap[*distributedStateHolder, bool]()
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
			if w.Cache == nil {
				panic(NewNullPointerException())
			}
			if w.Cache.Hit(fp) {
				continue
			}
			if w.unsorted {
				// Holder inherits Object.equals/hashCode. Distinct holder objects
				// remain distinct even when their fingerprints match. InsMap gives
				// deterministic iteration in place of Java's VM-dependent HashSet.
				identityHolders.Set(&distributedStateHolder{Fingerprint: fp, Successor: successor, Predecessor: state1}, true)
				continue
			}
			if _, ok := holdersByFP[fp]; ok {
				continue
			}
			holdersByFP[fp] = distributedStateHolder{Fingerprint: fp, Successor: successor, Predecessor: state1}
			orderedFPs = append(orderedFPs, fp)
		}
	}
	w.OverallStatesComputed.Add(statesComputed)

	sort.Slice(orderedFPs, func(i, j int) bool { return int64(orderedFPs[i]) < int64(orderedFPs[j]) })
	holders := make([]distributedStateHolder, 0, len(orderedFPs))
	if w.unsorted {
		for holder := range identityHolders.All() {
			holders = append(holders, *holder)
		}
	} else {
		for _, fp := range orderedFPs {
			holders = append(holders, holdersByFP[fp])
		}
	}
	serverCount := w.FPSetManager.NumOfServers()
	predecessors := make([]*StateVec, serverCount)
	successors := make([]*StateVec, serverCount)
	fingerprints := make([]*LongVec, serverCount)
	for i := 0; i < serverCount; i++ {
		predecessors[i] = NewStateVec(0)
		successors[i] = NewStateVec(0)
		fingerprints[i] = NewLongVec()
	}
	last := int64(math.MinInt64)
	for _, holder := range holders {
		fp := holder.Fingerprint
		if int64(fp) <= last {
			failure := newTLCErrorCode(ECGeneral)
			failure.Runtime = true
			panic(failure)
		}
		last = int64(fp)
		fpIndex := w.FPSetManager.GetFPSetIndex(fp)
		predecessors[fpIndex].Add(holder.Predecessor)
		successors[fpIndex].Add(holder.Successor)
		fingerprints[fpIndex].AddElement(int64(fp))
	}

	if w.Runtime == nil {
		panic(NewNullPointerException())
	}
	visited := w.FPSetManager.ContainsBlock(fingerprints, &w.Runtime.executor)
	newStates := make([]*StateVec, serverCount)
	newFingerprints := make([]*LongVec, serverCount)
	for i := 0; i < serverCount; i++ {
		newStates[i] = NewStateVec(0)
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
	if w == nil || w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.GetNextStates(state)
}

func (w *DistributedWorker) CheckState(predecessor *TLCStateMut, successor *TLCStateMut) error {
	if w != nil && w.CheckStateFunc != nil {
		return w.CheckStateFunc(predecessor, successor)
	}
	if w == nil || w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.CheckState(predecessor, successor)
}

func (w *DistributedWorker) IsInModel(state *TLCStateMut) (bool, error) {
	if w != nil && w.IsInModelFunc != nil {
		return w.IsInModelFunc(state)
	}
	if w == nil || w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.IsInModel(state)
}

func (w *DistributedWorker) IsInActions(predecessor *TLCStateMut, successor *TLCStateMut) (bool, error) {
	if w != nil && w.IsInActionsFunc != nil {
		return w.IsInActionsFunc(predecessor, successor)
	}
	if w == nil || w.App == nil {
		panic(NewNullPointerException())
	}
	return w.App.IsInActions(predecessor, successor)
}

func (w *DistributedWorker) IsAlive() bool {
	return w != nil
}

func (w *DistributedWorker) GetCacheRateRatio() float64 {
	if w == nil || w.Cache == nil {
		return 0
	}
	return w.Cache.GetHitRatio()
}

func (w *DistributedWorker) IsComputing() bool {
	return w != nil && w.Computing.Load()
}

type distributedStateHolder struct {
	Fingerprint uint64
	Successor   *TLCStateMut
	Predecessor *TLCStateMut
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
