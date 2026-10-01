package tlc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
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
	if r == nil {
		return 0
	}
	return r.StatesComputed - int64(len(r.NextStates))
}

func (r *NextStateResult) GetComputationTime() int64 {
	if r == nil {
		return 0
	}
	return r.ComputationTime
}

func (r *NextStateResult) GetNextFingerprints() []*LongVec {
	if r == nil {
		return nil
	}
	return r.NextFingerprints
}

func (r *NextStateResult) GetNextStates() []*StateVec {
	if r == nil {
		return nil
	}
	return r.NextStates
}

type DistributedFPSetManager struct {
	Sets               []FPSet
	Mask               uint64
	StatesSeen         atomic.Uint64
	Description        string
	ExpectedNumServers int
	checkpointFilename string
}

func NewDistributedFPSetManager(sets ...FPSet) *DistributedFPSetManager {
	out := &DistributedFPSetManager{Sets: append([]FPSet(nil), sets...), Mask: math.MaxInt64}
	out.normalize()
	return out
}

func NewDynamicDistributedFPSetManager(expectedNumOfServers int) *DistributedFPSetManager {
	if expectedNumOfServers <= 0 {
		panic("expected number of FPSet servers must be positive")
	}
	manager := NewDistributedFPSetManager()
	manager.ExpectedNumServers = expectedNumOfServers
	log := 0
	for expectedNumOfServers > 0 {
		expectedNumOfServers /= 2
		log++
	}
	manager.Mask = (uint64(1) << log) - 1
	return manager
}

func NewDistributedFPSetManagerFromFPSet(set FPSet) *DistributedFPSetManager {
	if multi, ok := set.(*MultiFPSet); ok && multi != nil {
		return &DistributedFPSetManager{
			Sets: append([]FPSet(nil), multi.Sets...),
			Mask: math.MaxInt64,
		}
	}
	return NewDistributedFPSetManager(set)
}

func (m *DistributedFPSetManager) normalize() {
	if m == nil {
		return
	}
	if m.Mask == 0 {
		m.Mask = math.MaxInt64
	}
}

type distributedFPSetIdentity struct {
	typ reflect.Type
	ptr uintptr
}

func fpSetIdentity(set FPSet) (distributedFPSetIdentity, bool) {
	if set == nil {
		return distributedFPSetIdentity{}, false
	}
	value := reflect.ValueOf(set)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		if value.IsNil() {
			return distributedFPSetIdentity{}, false
		}
		return distributedFPSetIdentity{typ: value.Type(), ptr: value.Pointer()}, true
	default:
		return distributedFPSetIdentity{}, false
	}
}

func (m *DistributedFPSetManager) distinctFPSets() []FPSet {
	if m == nil {
		return nil
	}
	sets := make([]FPSet, 0, len(m.Sets))
	seen := make(map[distributedFPSetIdentity]struct{}, len(m.Sets))
	for _, set := range m.Sets {
		if set == nil {
			continue
		}
		if key, ok := fpSetIdentity(set); ok {
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
		}
		sets = append(sets, set)
	}
	return sets
}

func (m *DistributedFPSetManager) NumOfServers() int {
	if m == nil || len(m.Sets) == 0 {
		return 1
	}
	return len(m.Sets)
}

func (m *DistributedFPSetManager) NumOfAliveServers() int {
	if m == nil {
		return 0
	}
	return len(m.distinctFPSets())
}

func (m *DistributedFPSetManager) RegisterFPSet(set FPSet, hostname ...string) error {
	if m == nil || set == nil {
		return nil
	}
	if m.ExpectedNumServers > 0 && len(m.Sets) >= m.ExpectedNumServers {
		return fmt.Errorf("Limit for FPset servers reached (%d). Cannot handle additional servers", m.ExpectedNumServers)
	}
	m.Sets = append(m.Sets, set)
	if len(hostname) > 0 && hostname[0] != "" {
		if m.Description == "" {
			m.Description = hostname[0]
		} else {
			m.Description += "," + hostname[0]
		}
	}
	m.normalize()
	return nil
}

func (m *DistributedFPSetManager) GetMask() uint64 {
	if m == nil {
		return uint64(math.MaxInt64)
	}
	m.normalize()
	return m.Mask
}

func (m *DistributedFPSetManager) GetHostName() string {
	return distributedServerHost()
}

func (m *DistributedFPSetManager) GetFPSetIndex(fp uint64) int {
	if m == nil || len(m.Sets) <= 1 {
		return 0
	}
	mask := m.Mask
	if mask == 0 {
		mask = math.MaxInt64
	}
	return int((fp & mask) % uint64(len(m.Sets)))
}

func (m *DistributedFPSetManager) ContainsBlock(fingerprints []*LongVec) []*BitVector {
	out := make([]*BitVector, len(fingerprints))
	for i, fpv := range fingerprints {
		size := 0
		if fpv != nil {
			size = fpv.Size()
		}
		if m == nil || i >= len(m.Sets) || m.Sets[i] == nil {
			out[i] = newAllTrueBitVector(size)
			m.addStatesSeen(uint64(size))
			continue
		}
		out[i] = m.Sets[i].ContainsBlock(fpv)
	}
	return out
}

func (m *DistributedFPSetManager) Put(fp uint64) bool {
	if m == nil || len(m.Sets) == 0 {
		m.addStatesSeen(1)
		return false
	}
	index := m.GetFPSetIndex(fp)
	if index < 0 || index >= len(m.Sets) || m.Sets[index] == nil {
		m.addStatesSeen(1)
		return false
	}
	return m.Sets[index].Put(fp)
}

func (m *DistributedFPSetManager) Contains(fp uint64) bool {
	if m == nil || len(m.Sets) == 0 {
		m.addStatesSeen(1)
		return false
	}
	index := m.GetFPSetIndex(fp)
	if index < 0 || index >= len(m.Sets) || m.Sets[index] == nil {
		m.addStatesSeen(1)
		return false
	}
	return m.Sets[index].Contains(fp)
}

func (m *DistributedFPSetManager) PutBlock(fingerprints []*LongVec) []*BitVector {
	out := make([]*BitVector, len(fingerprints))
	for i, fpv := range fingerprints {
		size := 0
		if fpv != nil {
			size = fpv.Size()
		}
		if m == nil || i >= len(m.Sets) || m.Sets[i] == nil {
			out[i] = newAllTrueBitVector(size)
			m.addStatesSeen(uint64(size))
			continue
		}
		out[i] = m.Sets[i].PutBlock(fpv)
	}
	return out
}

func (m *DistributedFPSetManager) Size() uint64 {
	if m == nil {
		return 0
	}
	var size uint64
	for _, set := range m.Sets {
		if set != nil {
			size += set.Size()
		}
	}
	return size
}

func (m *DistributedFPSetManager) CheckFPs() uint64 {
	if m == nil || len(m.Sets) == 0 {
		return 0
	}
	actualDistance := uint64(math.MaxUint64)
	checked := false
	for _, set := range m.Sets {
		if set == nil {
			continue
		}
		actualDistance = minUint64(actualDistance, set.CheckFPs())
		checked = true
	}
	if !checked {
		return 0
	}
	return actualDistance
}

func (m *DistributedFPSetManager) CheckInvariant(expectFPs ...uint64) bool {
	if m == nil {
		return true
	}
	for _, set := range m.Sets {
		if set != nil && !set.CheckInvariant(expectFPs...) {
			return false
		}
	}
	return true
}

func (m *DistributedFPSetManager) GetStatesSeen() uint64 {
	if m == nil {
		return 0
	}
	total := uint64(1) + m.StatesSeen.Load()
	for _, set := range m.Sets {
		if set != nil {
			total += set.GetStatesSeen()
		}
	}
	return total
}

func (m *DistributedFPSetManager) Close(cleanup bool) error {
	if m == nil {
		return nil
	}
	for _, set := range m.distinctFPSets() {
		if err := set.Exit(cleanup); err != nil {
			return err
		}
	}
	return nil
}

func (m *DistributedFPSetManager) Checkpoint(filename string) error {
	if m == nil {
		return nil
	}
	m.checkpointFilename = ""
	for _, set := range m.distinctFPSets() {
		if filename != "" {
			if err := set.BeginChkptFile(filename); err != nil {
				return err
			}
			if err := set.CommitChkptFile(filename); err != nil {
				return err
			}
		} else {
			if err := set.BeginChkpt(); err != nil {
				return err
			}
			if err := set.CommitChkpt(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) CommitCheckpoint() error {
	if m != nil {
		m.checkpointFilename = ""
	}
	return nil
}

func (m *DistributedFPSetManager) Recover(filename string) error {
	if m == nil {
		return nil
	}
	for _, set := range m.distinctFPSets() {
		if err := set.RecoverFile(filename); err != nil {
			return err
		}
	}
	return nil
}

func (m *DistributedFPSetManager) addStatesSeen(delta uint64) {
	if m != nil && delta != 0 {
		m.StatesSeen.Add(delta)
	}
}

type DistributedWorker struct {
	ID                    int
	Tool                  *Tool
	FPSetManager          *DistributedFPSetManager
	Cache                 *SimpleCache
	CheckDeadlock         bool
	URI                   string
	Computing             atomic.Bool
	LastInvocation        time.Time
	OverallStatesComputed int64
	CheckStateFunc        func(predecessor *TLCStateMut, successor *TLCStateMut) error
	IsInModelFunc         func(state *TLCStateMut) (bool, error)
	IsInActionsFunc       func(predecessor *TLCStateMut, successor *TLCStateMut) (bool, error)
	NetworkOverhead       float64
}

type TLCServer struct {
	FPSetManager                *DistributedFPSetManager
	StateQueue                  StateQueue
	Trace                       *TLCTrace
	Tool                        *Tool
	Metadir                     string
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
	NumberOfInitialStates       int64
	Workers                     *InsMap[string, *DistributedWorker]
	ServerThreads               *InsMap[string, *TLCServerThread]
	BlockSelector               *BlockSelector
	FinalNumberOfDistinctStates int64
}

func NewTLCServer(fileName string, configName string, metadir string, manager *DistributedFPSetManager, queue StateQueue, trace *TLCTrace) *TLCServer {
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
		FPSetManager:                manager,
		StateQueue:                  queue,
		Trace:                       trace,
		Metadir:                     metadir,
		FileName:                    fileName,
		ConfigName:                  configName,
		Workers:                     NewInsMap[string, *DistributedWorker](),
		ServerThreads:               NewInsMap[string, *TLCServerThread](),
		FinalNumberOfDistinctStates: -1,
	}
	server.BlockSelector = NewBlockSelectorFromProperties(server)
	return server
}

func (s *TLCServer) SetTool(tool *Tool) *TLCServer {
	if s != nil {
		s.Tool = tool
		if s.Trace != nil {
			s.Trace.SetTool(tool)
		}
	}
	return s
}

func (s *TLCServer) HasNoErrors() bool {
	return s != nil && s.ErrState == nil && s.LastError == nil
}

func (s *TLCServer) Checkpoint() error {
	if s == nil || s.StateQueue == nil {
		return nil
	}
	if !s.StateQueue.SuspendAll() {
		return nil
	}
	PrintMessage(ECTLCCheckpointStart, s.Metadir)
	if err := s.StateQueue.BeginChkpt(); err != nil {
		s.StateQueue.ResumeAll()
		return err
	}
	if s.Trace != nil {
		if err := s.Trace.BeginChkpt(); err != nil {
			s.StateQueue.ResumeAll()
			return err
		}
	}
	if s.FPSetManager != nil {
		if err := s.FPSetManager.Checkpoint(s.FileName); err != nil {
			s.StateQueue.ResumeAll()
			return err
		}
	}
	s.StateQueue.ResumeAll()
	if err := s.StateQueue.CommitChkpt(); err != nil {
		return err
	}
	if s.Trace != nil {
		if err := s.Trace.CommitChkpt(); err != nil {
			return err
		}
	}
	if s.FPSetManager != nil {
		if err := s.FPSetManager.CommitCheckpoint(); err != nil {
			return err
		}
	}
	PrintMessage(ECTLCCheckpointEnd)
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
		if err := s.FPSetManager.Recover(s.FileName); err != nil {
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
		if s.StateQueue != nil {
			if err := s.StateQueue.Delete(); err != nil {
				return err
			}
		}
		if s.Trace != nil {
			if err := s.Trace.Delete(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *TLCServer) DoInit(tool ...*Tool) (int, error) {
	if s == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server is nil")
	}
	if len(tool) > 0 {
		s.Tool = tool[0]
	}
	if s.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server has no tool")
	}
	functor := &distributedDoInitFunctor{server: s, tool: s.Tool, returnValue: NoError}
	err := s.Tool.GetInitStates(NewStateFunctor(functor.AddElement))
	if errors.Is(err, errInvariantViolated) {
		s.ErrState = functor.errState
		return functor.returnValue, nil
	}
	if err != nil {
		if functor.errState != nil {
			s.ErrState = functor.errState
		}
		if functor.returnValue != NoError {
			return functor.returnValue, err
		}
		return ECGeneral, err
	}
	if functor.errState != nil {
		s.ErrState = functor.errState
		if functor.err != nil {
			return functor.returnValue, functor.err
		}
	}
	return functor.returnValue, nil
}

func (s *TLCServer) ModelCheck(tool ...*Tool) (int, error) {
	if s == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server is nil")
	}
	startTime := time.Now()
	if len(tool) > 0 {
		s.Tool = tool[0]
	}
	if s.Tool == nil {
		return ECGeneral, newTLCError(ECGeneral, "distributed TLC server has no tool")
	}
	PrintMessage(ECTLCComputingInit)
	result, err := s.DoInit()
	if err != nil {
		if result == NoError {
			result = ECGeneral
		}
		s.PrintSummary(1, 0, s.GetNewStates(), s.fpSetSize(), false)
		PrintMessage(ECTLCFinished, humanReadableTLCRuntime(time.Since(startTime)))
		_ = s.Close(false)
		return result, err
	}
	if result != NoError || !s.HasNoErrors() {
		if result == NoError {
			result = ECGeneral
		}
		s.PrintSummary(1, 0, s.GetNewStates(), s.fpSetSize(), false)
		PrintMessage(ECTLCFinished, humanReadableTLCRuntime(time.Since(startTime)))
		_ = s.Close(false)
		return result, nil
	}
	s.PrintInitGenerated()
	PrintMessage(ECTLCDistributedServerRunning, distributedServerHost())
	if len(s.Tool.GetActions()) == 0 {
		if s.StateQueue != nil && !s.StateQueue.IsEmpty() {
			PrintError(ECTLCStatesAndNoNextAction)
			s.PrintSummary(1, s.GetStatesGenerated(), s.GetNewStates(), s.fpSetSize(), false)
			_ = s.Close(false)
			return ECTLCStatesAndNoNextAction, nil
		}
		s.ReportSuccess()
		s.PrintSummary(1, s.GetStatesGenerated(), s.GetNewStates(), s.fpSetSize(), true)
		PrintMessage(ECTLCFinished, humanReadableTLCRuntime(time.Since(startTime)))
		_ = s.Close(true)
		return NoError, nil
	}
	if s.GetWorkerCount() == 0 {
		err := newTLCError(ECGeneral, "distributed TLC server has no registered workers")
		s.LastError = err
		s.SetDone()
		s.PrintSummary(1, s.GetStatesGenerated(), s.GetNewStates(), s.fpSetSize(), false)
		PrintMessage(ECTLCFinished, humanReadableTLCRuntime(time.Since(startTime)))
		_ = s.Close(false)
		return ECGeneral, err
	}
	for _, thread := range s.ServerThreads.All() {
		if thread != nil {
			thread.Start()
		}
	}
	s.waitForDistributedCompletion(startTime)
	for _, thread := range s.ServerThreads.All() {
		if thread == nil {
			continue
		}
		thread.Join()
		if thread.Worker != nil {
			_ = thread.Worker.Exit()
		}
		cacheRatio := "n/a"
		if thread.GetCacheRateRatio() >= 0 {
			cacheRatio = fmt.Sprintf("%.2f", thread.GetCacheRateRatio())
		}
		PrintMessage(ECTLCDistributedWorkerStats,
			thread.GetURI(),
			fmtInt(thread.GetSentStates()),
			fmtInt(thread.GetReceivedStates()),
			cacheRatio,
		)
	}
	s.FinalNumberOfDistinctStates = int64(s.fpSetSize())
	statesGenerated := s.GetStatesGenerated()
	statesLeft := s.GetNewStates()
	level := 1
	if s.Trace != nil {
		level = s.Trace.GetLevelForReporting()
	}
	s.StatesPerMinute = 0
	s.DistinctStatesPerMinute = 0
	if s.HasNoErrors() {
		s.ReportSuccess()
	} else if s.KeepCallStack {
		// The concrete call-stack replay is handled by Tool/CallStackTool in the
		// local checker. Distributed TLC records the intent here for callers.
		s.KeepCallStack = true
	}
	s.PrintSummary(level, statesGenerated, statesLeft, s.fpSetSize(), s.HasNoErrors())
	PrintMessage(ECTLCFinished, humanReadableTLCRuntime(time.Since(startTime)))
	if err := s.Close(s.HasNoErrors()); err != nil && s.HasNoErrors() {
		return ECGeneral, err
	}
	if s.HasNoErrors() {
		return NoError, nil
	}
	if s.ErrorCode != NoError {
		return s.ErrorCode, s.LastError
	}
	return ECGeneral, s.LastError
}

func (s *TLCServer) waitForDistributedCompletion(startTime time.Time) {
	progress := ProgressInterval()
	if progress <= 0 {
		progress = time.Duration(DefaultProgressIntervalMillis) * time.Millisecond
	}
	ticker := time.NewTicker(progress)
	defer ticker.Stop()
	oldGenerated := int64(0)
	oldDistinct := uint64(0)
	for !s.IsDone() {
		<-ticker.C
		if DoCheckPoint() {
			if err := s.Checkpoint(); err != nil {
				s.LastError = err
				s.SetErrState(nil, nil, true, ECGeneral)
				return
			}
		}
		if s.IsDone() {
			return
		}
		s.PrintProgressStats(startTime, &oldGenerated, &oldDistinct)
	}
}

func (s *TLCServer) PrintProgressStats(startTime time.Time, oldGenerated *int64, oldDistinct *uint64) {
	if s == nil {
		return
	}
	generated := s.GetStatesGenerated()
	distinct := s.fpSetSize()
	factor := ProgressInterval().Minutes()
	if factor <= 0 {
		factor = 1
	}
	if oldGenerated != nil {
		s.StatesPerMinute = int64(float64(generated-*oldGenerated) / factor)
		*oldGenerated = generated
	}
	if oldDistinct != nil {
		var distinctDelta uint64
		if distinct >= *oldDistinct {
			distinctDelta = distinct - *oldDistinct
		}
		s.DistinctStatesPerMinute = int64(float64(distinctDelta) / factor)
		*oldDistinct = distinct
	}
	level := 1
	if s.Trace != nil {
		level = s.Trace.GetLevelForReporting()
	}
	PrintMessage(ECTLCProgressStats,
		fmtInt(level),
		fmtInt64(generated),
		fmtUint64(distinct),
		fmtInt64(s.GetNewStates()),
		fmtInt64(s.StatesPerMinute),
		fmtInt64(s.DistinctStatesPerMinute),
	)
	_ = startTime
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
		PrintMessage(ECTLCProgressStats, fmtInt(level), fmtInt64(statesGenerated), fmtUint64(distinctStates), fmtInt64(statesLeftInQueue), "0", "0")
	}
	PrintMessage(ECTLCStats, fmtInt64(statesGenerated), fmtUint64(distinctStates), fmtInt64(statesLeftInQueue))
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

func (s *TLCServer) RegisterWorker(worker *DistributedWorker) {
	key := s.registerWorkerOnly(worker)
	if key == "" {
		return
	}
	if s.ServerThreads != nil && s.ServerThreads.Get(key) != nil {
		return
	}
	thread := NewTLCServerThread(worker, key, s, s.BlockSelector)
	thread.Start()
	PrintMessage(ECTLCDistributedWorkerRegistered, key)
}

func (s *TLCServer) registerWorkerOnly(worker *DistributedWorker) string {
	if s == nil || worker == nil {
		return ""
	}
	if s.Workers == nil {
		s.Workers = NewInsMap[string, *DistributedWorker]()
	}
	if s.StateQueue != nil {
		s.StateQueue.ResumeAllStuck()
	}
	key := distributedWorkerKey(worker)
	s.Workers.Set(key, worker)
	return key
}

func (s *TLCServer) RegisterTLCServerThread(thread *TLCServerThread) {
	if s == nil || thread == nil {
		return
	}
	if s.ServerThreads == nil {
		s.ServerThreads = NewInsMap[string, *TLCServerThread]()
	}
	if thread.Worker != nil && thread.Worker.Worker != nil {
		s.registerWorkerOnly(thread.Worker.Worker)
	}
	s.ServerThreads.Set(thread.GetURI(), thread)
}

func (s *TLCServer) RemoveTLCServerThread(thread *TLCServerThread) *TLCServerThread {
	if s == nil || thread == nil {
		return nil
	}
	if thread.Worker != nil && thread.Worker.Worker != nil {
		s.RemoveWorker(thread.Worker.Worker)
	}
	if s.ServerThreads == nil {
		return nil
	}
	removed := s.ServerThreads.Get(thread.GetURI())
	s.ServerThreads.Delkey(thread.GetURI())
	if removed != nil {
		PrintMessage(ECTLCDistributedWorkerDeregistered, thread.GetURI())
	}
	return removed
}

func (s *TLCServer) RemoveWorker(worker *DistributedWorker) *DistributedWorker {
	if s == nil || s.Workers == nil || worker == nil {
		return nil
	}
	key := distributedWorkerKey(worker)
	removed := s.Workers.Get(key)
	s.Workers.Delkey(key)
	return removed
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
	if !s.Done.CompareAndSwap(false, true) {
		return false
	}
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
	server      *TLCServer
	tool        *Tool
	errState    *TLCStateMut
	err         error
	returnValue int
}

func (f *distributedDoInitFunctor) AddElement(curState *TLCStateMut) (any, error) {
	if f == nil || f.server == nil {
		return nil, newTLCError(ECGeneral, "distributed init functor has no server")
	}
	if isPowerOfTwo(f.server.NumberOfInitialStates) && f.server.NumberOfInitialStates > 1 {
		PrintMessage(ECTLCComputingInitProgress, fmt.Sprintf("%d", f.server.NumberOfInitialStates))
	}
	f.server.NumberOfInitialStates++
	if f.errState != nil {
		if f.returnValue == NoError {
			f.returnValue = ECTLCInitialState
		}
		return f.returnValue, nil
	}
	if f.tool == nil {
		f.err = newTLCError(ECGeneral, "distributed init functor has no tool")
		return ECGeneral, f.err
	}
	inModel, err := f.tool.IsInModel(curState)
	if err != nil {
		f.errState = curState
		f.err = err
		_ = f.server.SetErrState(curState, nil, true, ECGeneral)
		return f.returnValue, err
	}
	seen := false
	if inModel {
		fp := curState.FingerPrint()
		if f.server.FPSetManager != nil {
			seen = f.server.FPSetManager.Put(fp)
		}
		if !seen {
			if f.server.Trace != nil {
				if _, err := f.server.Trace.WriteState(nil, fp, curState, curState.GetAction()); err != nil {
					f.errState = curState
					f.err = err
					_ = f.server.SetErrState(curState, nil, true, ECGeneral)
					return f.returnValue, err
				}
			}
			if f.server.StateQueue != nil {
				f.server.StateQueue.SEnqueue(curState)
			}
		}
	}
	if !inModel || !seen {
		if err := distributedCheckState(f.tool, nil, curState); err != nil {
			if f.server.SetErrState(curState, nil, true, ECGeneral) {
				f.errState = curState
				f.err = err
				f.returnValue = ECGeneral
			}
			return f.returnValue, nil
		}
	}
	return f.returnValue, nil
}

func distributedVetoCleanup() bool {
	if value, ok := tlcLookupSystemProperty(tlcServerVetoCleanup); ok {
		return javaBooleanProperty(value)
	}
	return false
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
	var size int64
	if s.StateQueue != nil {
		size += s.StateQueue.Size()
	}
	if s.Workers != nil {
		for _, worker := range s.Workers.All() {
			if worker != nil && worker.IsComputing() {
				size++
			}
		}
	}
	return size
}

func (s *TLCServer) GetStatesGeneratedPerMinute() int64 {
	if s == nil {
		return 0
	}
	return s.StatesPerMinute
}

func (s *TLCServer) GetDistinctStatesGeneratedPerMinute() int64 {
	if s == nil {
		return 0
	}
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
	if s == nil || s.Workers == nil {
		return 0
	}
	return s.Workers.Len()
}

func (s *TLCServer) GetFPSetManager() *DistributedFPSetManager {
	if s == nil {
		return nil
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

type TLCServerThread struct {
	ID                int
	ReceivedStates    int
	SentStates        int
	CacheRateHitRatio float64
	Selector          *BlockSelector
	States            []*TLCStateMut
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

func NewTLCServerThread(worker *DistributedWorker, uri string, server *TLCServer, selector *BlockSelector) *TLCServerThread {
	if uri == "" && worker != nil {
		uri = distributedWorkerKey(worker)
	}
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
		States:            []*TLCStateMut{},
		Worker:            NewDistributedWorkerSmartProxy(worker),
		Server:            server,
		URI:               uri,
		keepAliveDone:     make(chan struct{}),
		runDone:           make(chan struct{}),
	}
	thread.cleanupGlobals.Store(true)
	thread.TimerTask = &TLCTimerTask{Thread: thread}
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
	t.startKeepAlive()
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
	go func() {
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				t.TimerTask.Run()
				timer.Reset(60 * time.Second)
			case <-t.keepAliveDone:
				return
			}
		}
	}()
}

func (t *TLCServerThread) Run() {
	if t == nil || t.Server == nil {
		return
	}
	IncNumWorkers(1)
	stateQueue := t.Server.StateQueue
	defer func() {
		if recovered := recover(); recovered != nil {
			t.handleRunError(fmt.Errorf("panic in TLCServerThread: %v", recovered), stateQueue)
		}
		t.readCacheRateRatio()
		t.cancelKeepAlive()
		t.States = []*TLCStateMut{}
	}()
	for {
		if t.Selector == nil {
			t.Selector = t.Server.BlockSelector
		}
		if t.Selector == nil {
			t.Selector = NewBlockSelectorFromProperties(t.Server)
		}
		t.States = t.Selector.GetBlocks(stateQueue, t.underlyingWorker())
		if t.States == nil {
			t.Server.SetDone()
			if stateQueue != nil {
				stateQueue.FinishAll()
			}
			return
		}
		if len(t.States) == 0 {
			continue
		}
		t.SentStates += len(t.States)

		res, ok := t.computeBlock(stateQueue)
		if !ok {
			return
		}
		if res == nil {
			continue
		}
		newStates := res.GetNextStates()
		newFps := res.GetNextFingerprints()
		if len(newStates) > 0 && newStates[0] != nil {
			t.ReceivedStates += newStates[0].Size()
		}
		if t.TimerTask != nil {
			t.TimerTask.SetLastInvocation(time.Now())
		}
		t.Server.AddStatesGeneratedDelta(res.GetStatesComputedDelta())
		t.publishBlock(stateQueue, newStates, newFps)
	}
}

func (t *TLCServerThread) computeBlock(stateQueue StateQueue) (*NextStateResult, bool) {
	for {
		res, err := t.Worker.GetNextStates(t.States)
		if err == nil {
			return res, true
		}
		var workerErr *WorkerException
		if errors.As(err, &workerErr) {
			t.handleRunError(workerErr, stateQueue)
			return nil, false
		}
		if isRecoverableDistributedError(err) && len(t.States) > 1 {
			if stateQueue != nil {
				stateQueue.SEnqueueAll(t.States)
			}
			if t.Selector != nil {
				t.Selector.SetMaxTXSize(len(t.States) / 2)
			}
			PrintMessage(ECTLCDistributedExceedBlocksize, fmtInt(len(t.States)/2))
			return nil, true
		}
		PrintMessage(ECTLCDistributedWorkerLost, t.GetURI())
		t.HandleRemoteWorkerLost(stateQueue)
		return nil, false
	}
}

func (t *TLCServerThread) publishBlock(stateQueue StateQueue, newStates []*StateVec, newFps []*LongVec) {
	if t == nil || t.Server == nil || t.Server.FPSetManager == nil {
		return
	}
	visited := t.Server.FPSetManager.PutBlock(newFps)
	for i, vector := range visited {
		if i >= len(newStates) || i >= len(newFps) || newStates[i] == nil || newFps[i] == nil {
			continue
		}
		iter := NewBitVectorIter(vector)
		for {
			index := iter.Next()
			if index == -1 {
				break
			}
			state := newStates[i].At(index)
			fp := uint64(newFps[i].ElementAt(index))
			if t.Server.Trace != nil {
				if _, err := t.Server.Trace.WriteState(state.Predecessor(), fp, state, state.GetAction()); err != nil {
					t.handleRunError(err, stateQueue)
					return
				}
			}
			if stateQueue != nil {
				stateQueue.SEnqueue(state)
			}
		}
	}
}

func (t *TLCServerThread) handleRunError(err error, stateQueue StateQueue) {
	if t == nil || t.Server == nil {
		return
	}
	t.Server.LastError = err
	var workerErr *WorkerException
	if errors.As(err, &workerErr) {
		if t.Server.SetErrState(workerErr.State1, nil, true, ECGeneral) {
			if workerErr.State1 != nil {
				if t.Server.Trace != nil {
					t.Server.Trace.PrintTrace(workerErr.State1, workerErr.State2)
				} else {
					PrintError(ECGeneral, generalErrorParams("", err)...)
				}
			} else {
				PrintError(ECGeneral, generalErrorParams("", err)...)
			}
			if stateQueue != nil {
				stateQueue.FinishAll()
			}
		}
		return
	}
	if t.Server.SetErrState(nil, nil, true, ECGeneral) && stateQueue != nil {
		stateQueue.FinishAll()
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
		stateQueue.SEnqueueAll(t.States)
	}
	t.States = []*TLCStateMut{}
	if stateQueue != nil {
		stateQueue.ResumeAllStuck()
	}
	DecNumWorkers()
}

func (t *TLCServerThread) GetCurrentSize() int {
	if t == nil {
		return 0
	}
	return len(t.States)
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
	t.CacheRateHitRatio = t.Worker.GetCacheRateRatio()
}

func (t *TLCServerThread) cancelKeepAlive() {
	if t == nil || t.keepAliveDone == nil {
		return
	}
	if t.keepAliveStopped.CompareAndSwap(false, true) {
		close(t.keepAliveDone)
	}
}

func (t *TLCServerThread) underlyingWorker() *DistributedWorker {
	if t == nil || t.Worker == nil {
		return nil
	}
	return t.Worker.Worker
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
		if t.Thread.Worker == nil || !t.Thread.Worker.IsAlive() {
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

type DistributedRecoverableError struct {
	Err error
}

func NewDistributedRecoverableError(err error) *DistributedRecoverableError {
	return &DistributedRecoverableError{Err: err}
}

func (e *DistributedRecoverableError) Error() string {
	if e == nil || e.Err == nil {
		return "recoverable distributed worker error"
	}
	return e.Err.Error()
}

func (e *DistributedRecoverableError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func isRecoverableDistributedError(err error) bool {
	var recoverable *DistributedRecoverableError
	return errors.As(err, &recoverable)
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
	AverageBlockCnt      int64
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
	selector.AverageBlockCnt = int64(selector.StaticBlockSize)
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
	return javaIntProperty(value)
}

func (b *BlockSelector) GetBlocks(stateQueue StateQueue, worker *DistributedWorker) []*TLCStateMut {
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
	if b != nil {
		b.Maximum = maximum
	}
}

func (b *BlockSelector) GetAverageBlockCnt() int64 {
	if b == nil {
		return 0
	}
	return b.AverageBlockCnt
}

func (b *BlockSelector) getBlockSize(size int64, worker *DistributedWorker) int64 {
	if b == nil {
		return 1
	}
	switch b.Mode {
	case BlockSelectorStatic:
		return int64(b.StaticBlockSize)
	case BlockSelectorStatistical:
		if worker != nil && worker.NetworkOverhead != 0 {
			limit := b.NetworkOverheadLimit
			if limit == 0 {
				limit = blockSelectorDefaultNetworkOverheadLimit
			}
			maximum := b.Maximum
			if maximum <= 0 {
				maximum = blockSelectorDefaultMaximum
			}
			blockSize := math.Abs(math.Ceil(float64(size) * (worker.NetworkOverhead / limit)))
			blockSize = math.Min(math.Max(blockSize, 1), float64(maximum))
			return int64(blockSize)
		}
		fallthrough
	case BlockSelectorLimiting:
		blockSize := b.proportionalBlockSize(size)
		maximum := b.Maximum
		if maximum <= 0 {
			maximum = blockSelectorDefaultMaximum
		}
		if blockSize > int64(maximum) {
			return int64(maximum)
		}
		return blockSize
	default:
		return b.proportionalBlockSize(size)
	}
}

func (b *BlockSelector) proportionalBlockSize(size int64) int64 {
	workerCount := 1
	if b != nil && b.Server != nil {
		workerCount = b.Server.GetWorkerCount()
	}
	if workerCount <= 0 {
		workerCount = 1
	}
	return int64(math.Ceil(float64(size) * (1.0 / float64(workerCount))))
}

func (b *BlockSelector) setAverageBlockCnt(blockCnt int64) {
	if b == nil || b.Mode == BlockSelectorStatic {
		return
	}
	if b.AverageBlockCnt > 0 {
		b.AverageBlockCnt = (blockCnt + b.AverageBlockCnt) / 2
	} else {
		b.AverageBlockCnt = blockCnt
	}
}

func NewDistributedWorker(id int, tool *Tool, fpSetManager *DistributedFPSetManager) *DistributedWorker {
	if fpSetManager == nil {
		fpSetManager = NewDistributedFPSetManager()
	}
	checkDeadlock := true
	if tool != nil && tool.GetModelConfig() != nil {
		checkDeadlock = tool.GetModelConfig().GetCheckDeadlock()
	}
	return &DistributedWorker{
		ID:              id,
		Tool:            tool,
		FPSetManager:    fpSetManager,
		Cache:           NewSimpleCache(),
		CheckDeadlock:   checkDeadlock,
		NetworkOverhead: math.MaxFloat64,
	}
}

func distributedWorkerKey(worker *DistributedWorker) string {
	if worker == nil {
		return ""
	}
	if worker.URI != "" {
		return worker.URI
	}
	return strconv.Itoa(worker.ID)
}

type DistributedWorkerSmartProxy struct {
	Worker          *DistributedWorker
	NetworkOverhead float64
}

func NewDistributedWorkerSmartProxy(worker *DistributedWorker) *DistributedWorkerSmartProxy {
	return &DistributedWorkerSmartProxy{Worker: worker, NetworkOverhead: math.MaxFloat64}
}

func (p *DistributedWorkerSmartProxy) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	if p == nil || p.Worker == nil {
		return nil, newTLCError(ECGeneral, "distributed worker proxy has no worker")
	}
	start := time.Now()
	nextStates, err := p.Worker.GetNextStates(states)
	if err != nil {
		return nil, err
	}
	roundTripTime := time.Since(start).Milliseconds() + 1
	computationTime := sanitizeDistributedComputationTime(nextStates.GetComputationTime())
	networkTime := math.Max(float64(roundTripTime-computationTime), 0.00001)
	percentageNetworkOverhead := networkTime / float64(roundTripTime)
	stateCount := len(states)
	if stateCount <= 0 {
		stateCount = 1
	}
	p.NetworkOverhead = percentageNetworkOverhead / float64(stateCount)
	p.Worker.NetworkOverhead = p.NetworkOverhead
	return nextStates, nil
}

func (p *DistributedWorkerSmartProxy) GetNetworkOverhead() float64 {
	if p == nil {
		return math.MaxFloat64
	}
	return p.NetworkOverhead
}

func (p *DistributedWorkerSmartProxy) Exit() error {
	return nil
}

func (p *DistributedWorkerSmartProxy) GetURI() string {
	if p == nil || p.Worker == nil {
		return ""
	}
	return p.Worker.URI
}

func (p *DistributedWorkerSmartProxy) IsAlive() bool {
	return p != nil && p.Worker != nil && p.Worker.IsAlive()
}

func (p *DistributedWorkerSmartProxy) GetCacheRateRatio() float64 {
	if p == nil || p.Worker == nil {
		return 0
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

func (w *DistributedWorker) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	if w == nil {
		return nil, newTLCError(ECGeneral, "distributed worker is nil")
	}
	w.Computing.Store(true)
	w.LastInvocation = time.Now()
	var statesComputed int64
	var state1, state2 *TLCStateMut
	defer w.Computing.Store(false)

	holdersByFP := make(map[uint64]distributedStateHolder, len(states))
	orderedFPs := make([]uint64, 0, len(states))
	for _, state := range states {
		state1 = state
		nextStates, err := w.computeNextStates(state)
		if err != nil {
			var workerErr *WorkerException
			if errors.As(err, &workerErr) {
				return nil, workerErr
			}
			return nil, NewWorkerExceptionWithCause(err.Error(), err, state1, state2, true)
		}
		statesComputed += int64(nextStates.Size())
		for i := 0; i < nextStates.Size(); i++ {
			state2 = nextStates.At(i)
			fp := state2.FingerPrint()
			if w.Cache != nil && w.Cache.Hit(fp) {
				continue
			}
			if _, ok := holdersByFP[fp]; ok {
				continue
			}
			holdersByFP[fp] = distributedStateHolder{Fingerprint: fp, Successor: state2, Predecessor: state1}
			orderedFPs = append(orderedFPs, fp)
		}
	}
	w.OverallStatesComputed += statesComputed

	sort.Slice(orderedFPs, func(i, j int) bool { return int64(orderedFPs[i]) < int64(orderedFPs[j]) })
	serverCount := w.FPSetManager.NumOfServers()
	predecessors := make([]*StateVec, serverCount)
	successors := make([]*StateVec, serverCount)
	fingerprints := make([]*LongVec, serverCount)
	for i := 0; i < serverCount; i++ {
		predecessors[i] = NewStateVec(0)
		successors[i] = NewStateVec(0)
		fingerprints[i] = NewLongVec()
	}
	for _, fp := range orderedFPs {
		holder := holdersByFP[fp]
		fpIndex := w.FPSetManager.GetFPSetIndex(fp)
		predecessors[fpIndex].Add(holder.Predecessor)
		successors[fpIndex].Add(holder.Successor)
		fingerprints[fpIndex].AddElement(int64(fp))
	}

	visited := w.FPSetManager.ContainsBlock(fingerprints)
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
				var workerErr *WorkerException
				if errors.As(err, &workerErr) {
					return nil, workerErr
				}
				return nil, NewWorkerExceptionWithCause(err.Error(), err, state1, state2, true)
			}
			inModel, err := w.IsInModel(state2)
			if err != nil {
				return nil, NewWorkerExceptionWithCause(err.Error(), err, state1, state2, true)
			}
			inActions, err := w.IsInActions(state1, state2)
			if err != nil {
				return nil, NewWorkerExceptionWithCause(err.Error(), err, state1, state2, true)
			}
			if inModel && inActions {
				state2.UID = state1.UID
				newStates[i].Add(state2)
				newFingerprints[i].AddElement(fingerprints[i].ElementAt(index))
			}
		}
	}

	elapsed := time.Since(w.LastInvocation).Milliseconds()
	return NewNextStateResult(newStates, newFingerprints, elapsed, statesComputed), nil
}

func (w *DistributedWorker) computeNextStates(state *TLCStateMut) (*StateVec, error) {
	out := NewStateVec(0)
	if w == nil || w.Tool == nil {
		return out, nil
	}
	functor := NewNextStateFunctor(func(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
		if succState != nil {
			succState.attachTraceMetadata(curState, action)
			out.Add(succState)
		}
		return out, nil
	})
	_, err := w.Tool.GetNextStatesWithFunctor(functor, state)
	if err != nil {
		return out, err
	}
	if out.Size() == 0 && w.CheckDeadlock {
		return out, NewWorkerException("Error: deadlock reached.", state, nil, false)
	}
	for i := 0; i < out.Size(); i++ {
		succState := out.At(i)
		if !w.Tool.IsGoodState(succState) {
			return out, NewWorkerException("Error: Successor state is not completely specified by the next-state action.", state, succState, false)
		}
	}
	return out, nil
}

func (w *DistributedWorker) CheckState(predecessor *TLCStateMut, successor *TLCStateMut) error {
	if w != nil && w.CheckStateFunc != nil {
		return w.CheckStateFunc(predecessor, successor)
	}
	if w == nil {
		return nil
	}
	return distributedCheckState(w.Tool, predecessor, successor)
}

func distributedCheckState(tool *Tool, predecessor *TLCStateMut, successor *TLCStateMut) error {
	if tool == nil || successor == nil {
		return nil
	}
	for i, invariant := range tool.GetInvariants() {
		valid, err := tool.IsValidState(invariant, successor)
		if err != nil {
			return err
		}
		if !valid {
			return NewWorkerException(fmt.Sprintf("Error: Invariant %s is violated.", nameAt(tool.GetInvNames(), i)), predecessor, successor, false)
		}
	}
	if predecessor == nil {
		for i, implied := range tool.GetImpliedInits() {
			valid, err := tool.IsValidState(implied, successor)
			if err != nil {
				return err
			}
			if !valid {
				return NewWorkerException(fmt.Sprintf("Error: Implied-init %s is violated.", nameAt(tool.GetImpliedInitNames(), i)), predecessor, successor, false)
			}
		}
		return nil
	}
	for i, implied := range tool.GetImpliedActions() {
		valid, err := tool.IsValidTransition(implied, predecessor, successor)
		if err != nil {
			return err
		}
		if !valid {
			return NewWorkerException(fmt.Sprintf("Error: Implied-action %s is violated.", nameAt(tool.GetImpliedActNames(), i)), predecessor, successor, false)
		}
	}
	return nil
}

func (w *DistributedWorker) IsInModel(state *TLCStateMut) (bool, error) {
	if w != nil && w.IsInModelFunc != nil {
		return w.IsInModelFunc(state)
	}
	if w == nil || w.Tool == nil {
		return true, nil
	}
	return w.Tool.IsInModel(state)
}

func (w *DistributedWorker) IsInActions(predecessor *TLCStateMut, successor *TLCStateMut) (bool, error) {
	if w != nil && w.IsInActionsFunc != nil {
		return w.IsInActionsFunc(predecessor, successor)
	}
	if w == nil || w.Tool == nil {
		return true, nil
	}
	return w.Tool.IsInActions(predecessor, successor)
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

func newAllTrueBitVector(size int) *BitVector {
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		bv.Set(i)
	}
	return bv
}
