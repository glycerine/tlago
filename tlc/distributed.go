package tlc

import (
	"errors"
	"fmt"
	"math"
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
	Shift              uint
	StatesSeen         atomic.Uint64
	Description        string
	checkpointFilename string
}

func NewDistributedFPSetManager(sets ...FPSet) *DistributedFPSetManager {
	out := &DistributedFPSetManager{Sets: append([]FPSet(nil), sets...)}
	out.normalize()
	return out
}

func NewDistributedFPSetManagerFromFPSet(set FPSet) *DistributedFPSetManager {
	if multi, ok := set.(*MultiFPSet); ok && multi != nil {
		return &DistributedFPSetManager{
			Sets:  append([]FPSet(nil), multi.Sets...),
			Shift: multi.Shift,
		}
	}
	return NewDistributedFPSetManager(set)
}

func (m *DistributedFPSetManager) normalize() {
	if m == nil || len(m.Sets) <= 1 || m.Shift != 0 {
		return
	}
	bits := 0
	for count := len(m.Sets); count > 1; count >>= 1 {
		bits++
	}
	m.Shift = uint(64 - bits)
}

func (m *DistributedFPSetManager) NumOfServers() int {
	if m == nil || len(m.Sets) == 0 {
		return 1
	}
	return len(m.Sets)
}

func (m *DistributedFPSetManager) GetFPSetIndex(fp uint64) int {
	if m == nil || len(m.Sets) <= 1 {
		return 0
	}
	index := int(fp >> m.Shift)
	if index < 0 {
		return 0
	}
	if index >= len(m.Sets) {
		index %= len(m.Sets)
	}
	return index
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

func (m *DistributedFPSetManager) GetStatesSeen() uint64 {
	if m == nil {
		return 0
	}
	total := m.StatesSeen.Load()
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
	for _, set := range m.Sets {
		if set != nil {
			if err := set.Exit(cleanup); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) Checkpoint(filename string) error {
	if m == nil {
		return nil
	}
	m.checkpointFilename = filename
	for _, set := range m.Sets {
		if set != nil {
			if err := set.BeginChkptFile(filename); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) CommitCheckpoint() error {
	if m == nil {
		return nil
	}
	filename := m.checkpointFilename
	for _, set := range m.Sets {
		if set != nil {
			if filename != "" {
				if err := set.CommitChkptFile(filename); err != nil {
					return err
				}
			} else if err := set.CommitChkpt(); err != nil {
				return err
			}
		}
	}
	m.checkpointFilename = ""
	return nil
}

func (m *DistributedFPSetManager) Recover(filename string) error {
	if m == nil {
		return nil
	}
	for _, set := range m.Sets {
		if set != nil {
			if err := set.RecoverFile(filename); err != nil {
				return err
			}
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
	server.BlockSelector = NewStatisticalBlockSelector(server)
	return server
}

func (s *TLCServer) RegisterWorker(worker *DistributedWorker) {
	if s == nil || worker == nil {
		return
	}
	if s.Workers == nil {
		s.Workers = NewInsMap[string, *DistributedWorker]()
	}
	if s.StateQueue != nil {
		s.StateQueue.ResumeAllStuck()
	}
	s.Workers.Set(distributedWorkerKey(worker), worker)
}

func (s *TLCServer) RegisterTLCServerThread(thread *TLCServerThread) {
	if s == nil || thread == nil {
		return
	}
	if s.ServerThreads == nil {
		s.ServerThreads = NewInsMap[string, *TLCServerThread]()
	}
	if thread.Worker != nil && thread.Worker.Worker != nil {
		s.RegisterWorker(thread.Worker.Worker)
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
}

func NewTLCServerThread(worker *DistributedWorker, uri string, server *TLCServer, selector *BlockSelector) *TLCServerThread {
	if uri == "" && worker != nil {
		uri = distributedWorkerKey(worker)
	}
	if selector == nil && server != nil {
		selector = server.BlockSelector
	}
	if selector == nil {
		selector = NewStatisticalBlockSelector(server)
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
	go t.Run()
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
			t.Selector = NewStatisticalBlockSelector(t.Server)
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
		if isRecoverableDistributedError(err) && len(t.States) > 1 {
			if stateQueue != nil {
				stateQueue.SEnqueueAll(t.States)
			}
			if t.Selector != nil {
				t.Selector.SetMaxTXSize(len(t.States) / 2)
			}
			return nil, true
		}
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
				if err := t.Server.Trace.WriteNextState(state.Predecessor(), fp, state, state.GetAction()); err != nil {
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
		if t.Server.SetErrState(workerErr.State1, workerErr.State2, workerErr.KeepCallStack, ECGeneral) && stateQueue != nil {
			stateQueue.FinishAll()
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
	return NewStatisticalBlockSelector(server)
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
	}
	selector.AverageBlockCnt = int64(selector.StaticBlockSize)
	return selector
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
	return &DistributedWorker{
		ID:              id,
		Tool:            tool,
		FPSetManager:    fpSetManager,
		Cache:           NewSimpleCache(),
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

	sort.Slice(orderedFPs, func(i, j int) bool { return orderedFPs[i] < orderedFPs[j] })
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
			succState.SetPredecessor(curState).SetAction(action)
			out.Add(succState)
		}
		return out, nil
	})
	_, err := w.Tool.GetNextStatesWithFunctor(functor, state)
	return out, err
}

func (w *DistributedWorker) CheckState(predecessor *TLCStateMut, successor *TLCStateMut) error {
	if w != nil && w.CheckStateFunc != nil {
		return w.CheckStateFunc(predecessor, successor)
	}
	if w == nil || w.Tool == nil || successor == nil || w.Tool.IsGoodState(successor) {
		return nil
	}
	return newTLCError(ECTLCStateNotCompletelySpecifiedNext, "%s", successor)
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
