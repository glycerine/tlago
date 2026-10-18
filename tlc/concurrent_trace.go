package tlc

import (
	"fmt"
)

type ConcurrentTLCTrace struct {
	*TLCTrace
	Tool            *Tool
	Workers         []*Worker
	workersCaptured bool
	workersOwner    []*Worker
}

func NewConcurrentTLCTrace(metadir string, specFile string, workerCount ...int) *ConcurrentTLCTrace {
	if len(workerCount) > 1 {
		panic(NewIllegalArgumentException("NewConcurrentTLCTrace accepts at most one worker count"))
	}
	trace := NewTLCTrace(metadir, specFile)
	count := NumWorkers()
	if len(workerCount) > 0 {
		count = int(int32(workerCount[0]))
	}
	if count < 0 {
		panic(NewNegativeArraySizeException(fmtInt(count)))
	}
	result := &ConcurrentTLCTrace{
		TLCTrace: trace,
		Workers:  make([]*Worker, count),
	}
	result.workersOwner = result.Workers
	result.workersCaptured = true
	return result
}

func (t *ConcurrentTLCTrace) ownedWorkers() []*Worker {
	if t.workersCaptured {
		return t.workersOwner
	}
	return t.Workers
}

func (t *ConcurrentTLCTrace) SetTool(tool *Tool) {
	if t != nil {
		t.Tool = tool
		if t.TLCTrace != nil {
			t.TLCTrace.SetTool(tool)
		}
	}
}

func (t *ConcurrentTLCTrace) SetCheckpointContext(metadir string, rootName string) {
	if t == nil {
		return
	}
	if t.TLCTrace != nil {
		t.TLCTrace.SetCheckpointContext(metadir, rootName)
	}
	for _, worker := range t.ownedWorkers() {
		if worker != nil {
			worker.SetTraceContext(metadir, rootName)
		}
	}
}

func (t *ConcurrentTLCTrace) AddWorker(worker *Worker) *Worker {
	if t == nil || worker == nil {
		panic(NewNullPointerException())
	}
	id := worker.MyGetID()
	workers := t.ownedWorkers()
	if id < 0 || id >= len(workers) {
		panic(NewArrayIndexOutOfBoundsException(id, len(workers)))
	}
	workers[id] = worker
	return worker
}

func (t *ConcurrentTLCTrace) GetLevelForReporting() int {
	return t.GetLevel()
}

func (t *ConcurrentTLCTrace) GetLevel() int {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.TLCTrace == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.levelLocked()
}

func (t *ConcurrentTLCTrace) levelLocked() int {
	maxLevel := 1
	for _, worker := range t.ownedWorkers() {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if level := worker.GetMaxLevel(); level > maxLevel {
			maxLevel = level
		}
	}
	return maxLevel
}

func (t *ConcurrentTLCTrace) PrintTrace(curState *TLCStateMut, succState *TLCStateMut) {
	if t == nil || curState == nil {
		panic(NewNullPointerException())
	}
	var prefix []*TLCStateInfo
	if !curState.IsInitial() {
		prefix = t.GetTrace(curState)
	}
	t.PrintTraceWithPrefix(curState, succState, prefix)
}

func (t *ConcurrentTLCTrace) GetTrace(states ...*TLCStateMut) []*TLCStateInfo {
	if t == nil {
		panic(NewNullPointerException())
	}
	if len(states) == 0 {
		if t.TLCTrace == nil {
			panic(NewNullPointerException())
		}
		return t.TLCTrace.GetTrace()
	}
	if len(states) != 1 {
		panic(NewIllegalArgumentException("GetTrace accepts at most one state"))
	}
	state := states[0]
	if state == nil {
		panic(NewNullPointerException())
	}
	if state.IsInitial() {
		return []*TLCStateInfo{NewTLCStateInfo(state)}
	}
	requireConcurrentTraceRecordCapacity(int32(state.Level()))
	trace, err := t.recoverTrace(state, nil)
	if err != nil {
		panic(err)
	}
	return trace
}

// GetTraceFromState retains the earlier native spelling.
func (t *ConcurrentTLCTrace) GetTraceFromState(state *TLCStateMut) []*TLCStateInfo {
	return t.GetTrace(state)
}

func (t *ConcurrentTLCTrace) GetTraceBetween(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	if t == nil {
		panic(NewNullPointerException())
	}
	if to == nil {
		panic(NewNullPointerException())
	}
	if to.IsInitial() {
		return []*TLCStateInfo{NewTLCStateInfo(to)}
	}
	if from == nil {
		panic(NewNullPointerException())
	}
	if from.Equal(to) {
		return []*TLCStateInfo{NewTLCStateInfo(to)}
	}
	requireConcurrentTraceRecordCapacity(int32(to.Level()) - int32(from.Level()))
	trace, err := t.recoverTrace(to, from)
	if err != nil {
		panic(err)
	}
	return trace
}

func requireConcurrentTraceRecordCapacity(capacity int32) {
	if capacity < 0 {
		panic(NewIllegalArgumentException(fmt.Sprintf("Illegal Capacity: %d", capacity)))
	}
}

// GetTraceBetweenStates retains the earlier native spelling.
func (t *ConcurrentTLCTrace) GetTraceBetweenStates(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	return t.GetTraceBetween(from, to)
}

func (t *ConcurrentTLCTrace) recoverTrace(state *TLCStateMut, from *TLCStateMut) ([]*TLCStateInfo, error) {
	if t == nil || state == nil {
		panic(NewNullPointerException())
	}
	if state.IsInitial() {
		return []*TLCStateInfo{NewTLCStateInfo(state)}, nil
	}
	first, err := t.recordForState(state)
	if err != nil {
		return nil, err
	}
	if t.TLCTrace == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	records, err := t.collectTraceRecordsLocked(state, from, first)
	if err != nil || len(records) == 0 {
		return nil, err
	}
	if len(records) > t.levelLocked() {
		panic(NewAssertionError())
	}
	if from != nil {
		return t.GetTraceFromRecords(NewTLCStateInfo(from), records)
	}
	return t.GetTraceFromRecords(nil, records)
}

func (t *ConcurrentTLCTrace) collectTraceRecordsLocked(state *TLCStateMut, from *TLCStateMut, first ConcurrentTraceRecord) ([]ConcurrentTraceRecord, error) {
	records := []ConcurrentTraceRecord{first}
	// Source reads the end record again under the trace monitor to obtain its
	// predecessor, and retains that monitor through state reconstruction.
	record, err := t.recordForState(state)
	if err != nil {
		return nil, err
	}
	record, err = t.predecessorRecord(record)
	if err != nil {
		return nil, err
	}
	for {
		if (from == nil && record.IsInitial()) || (from != nil && record.ownedFP() == from.FingerPrint()) {
			return append(records, record), nil
		}
		records = append(records, record)
		record, err = t.predecessorRecord(record)
		if err != nil {
			return nil, err
		}
	}
}

func (t *ConcurrentTLCTrace) recordForState(state *TLCStateMut) (ConcurrentTraceRecord, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	return concurrentTraceRecordForState(state, t.ownedWorkers())
}

func (t *ConcurrentTLCTrace) predecessorRecord(record ConcurrentTraceRecord) (ConcurrentTraceRecord, error) {
	return record.GetPredecessor()
}

// GetTraceFromRecords is the protected source overload that reconstructs state
// information from a collected predecessor chain.
func (t *ConcurrentTLCTrace) GetTraceFromRecords(sinfo *TLCStateInfo, records []ConcurrentTraceRecord) ([]*TLCStateInfo, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	snapshot := ResetRandomEnumerableValues()
	end := len(records) - 1
	if end < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(end)))
	}
	out := make([]*TLCStateInfo, 0, end)
	if end > 0 {
		if sinfo == nil {
			if !records[end].IsInitial() {
				panic(NewAssertionError())
			}
			if t.Tool == nil {
				panic(NewNullPointerException())
			}
			info, err := t.Tool.GetState(records[end].ownedFP())
			if err != nil {
				return nil, err
			}
			sinfo = info
			if sinfo == nil || sinfo.ownedState() == nil {
				panic(NewNullPointerException())
			}
			prev := records[end-1]
			sinfo.ownedState().WorkerID = int16(prev.ownedWorker())
			sinfo.ownedState().UID = prev.ownedPtr()
		}
		out = append(out, sinfo)
		for i := end - 2; i >= 0; i-- {
			record := records[i+1]
			if t.Tool == nil {
				panic(NewNullPointerException())
			}
			info, err := t.Tool.GetState(record.ownedFP(), sinfo.ownedState())
			if err != nil {
				return nil, err
			}
			if info == nil {
				traceRecoveryExit(fmt.Sprintf("2 %d", int64(record.ownedFP())), nil)
			}
			if info.ownedState() == nil {
				panic(NewNullPointerException())
			}
			prev := records[i]
			info.ownedState().WorkerID = int16(prev.ownedWorker())
			info.ownedState().UID = prev.ownedPtr()
			out = append(out, info)
			sinfo = info
		}
	}
	SetRandomEnumerableGenerator(snapshot)
	return out, nil
}

func (t *ConcurrentTLCTrace) recoverTraceFromRecords(sinfo *TLCStateInfo, records []ConcurrentTraceRecord) ([]*TLCStateInfo, error) {
	return t.GetTraceFromRecords(sinfo, records)
}

func (t *ConcurrentTLCTrace) CommitChkpt() error {
	if t == nil {
		panic(NewNullPointerException())
	}
	for _, worker := range t.ownedWorkers() {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if err := worker.CommitChkpt(); err != nil {
			return err
		}
	}
	if t.TLCTrace == nil {
		panic(NewNullPointerException())
	}
	if t.diskdir == "" {
		return nil
	}
	return touchFile(t.chkptName("chkpt"))
}

func (t *ConcurrentTLCTrace) BeginChkpt() error {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.TLCTrace == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, worker := range t.ownedWorkers() {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if err := worker.BeginChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (t *ConcurrentTLCTrace) Recover() error {
	if t == nil {
		panic(NewNullPointerException())
	}
	for _, worker := range t.ownedWorkers() {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if err := worker.Recover(); err != nil {
			return err
		}
	}
	return nil
}

func (t *ConcurrentTLCTrace) Elements() (*TLCTraceEnumerator, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.TLCTrace == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	workers := t.ownedWorkers()
	enums := make([]*WorkerTraceEnumerator, len(workers))
	for i, worker := range workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		enum, err := worker.Elements()
		if err != nil {
			return nil, err
		}
		enums[i] = enum
	}
	return &TLCTraceEnumerator{concurrent: &ConcurrentTraceEnumerator{enums: enums}}, nil
}

func (t *ConcurrentTLCTrace) Close() error {
	if t == nil {
		panic(NewNullPointerException())
	}
	var err error
	for _, worker := range t.ownedWorkers() {
		if worker != nil {
			if closeErr := worker.CloseTrace(); closeErr != nil && err == nil {
				err = closeErr
			}
		}
	}
	if t.TLCTrace == nil {
		panic(NewNullPointerException())
	}
	if closeErr := t.TLCTrace.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	return err
}

func (t *ConcurrentTLCTrace) Delete() error {
	if t == nil {
		return nil
	}
	var err error
	for _, worker := range t.ownedWorkers() {
		if worker != nil {
			if deleteErr := worker.DeleteTrace(); deleteErr != nil && err == nil {
				err = deleteErr
			}
		}
	}
	if t.TLCTrace != nil {
		if deleteErr := t.TLCTrace.Delete(); deleteErr != nil && err == nil {
			err = deleteErr
		}
	}
	return err
}

type ConcurrentTraceEnumerator struct {
	idx   int
	enums []*WorkerTraceEnumerator
}

func (e *ConcurrentTraceEnumerator) NextPos() (int64, error) {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.idx >= len(e.enums) {
		return -1, nil
	}
	reader, err := e.readerAtCurrentIndex()
	if err != nil {
		return 0, err
	}
	more, err := reader.HasMoreFP()
	if err != nil {
		return 0, err
	}
	if more {
		return 42, nil
	}
	if e.idx+1 >= len(e.enums) {
		return -1, nil
	}
	e.idx++
	reader, err = e.readerAtCurrentIndex()
	if err != nil {
		return 0, err
	}
	more, err = reader.HasMoreFP()
	if err != nil {
		return 0, err
	}
	if more {
		return 42, nil
	}
	return -1, nil
}

func (e *ConcurrentTraceEnumerator) readerAtCurrentIndex() (*WorkerTraceEnumerator, error) {
	if e == nil {
		return nil, NewNullPointerException()
	}
	if e.idx < 0 || e.idx >= len(e.enums) {
		return nil, NewArrayIndexOutOfBoundsException(e.idx, len(e.enums))
	}
	if e.enums[e.idx] == nil {
		return nil, NewNullPointerException()
	}
	return e.enums[e.idx], nil
}

func (e *ConcurrentTraceEnumerator) NextFP() (uint64, error) {
	if e == nil {
		panic(NewNullPointerException())
	}
	reader, err := e.readerAtCurrentIndex()
	if err != nil {
		return 0, err
	}
	more, err := reader.HasMoreFP()
	if err != nil {
		return 0, err
	}
	if !more {
		e.idx++
		reader, err = e.readerAtCurrentIndex()
		if err != nil {
			return 0, err
		}
	}
	return reader.NextFP()
}

func (e *ConcurrentTraceEnumerator) Close() error {
	if e == nil {
		panic(NewNullPointerException())
	}
	for _, enum := range e.enums {
		if err := enum.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (e *ConcurrentTraceEnumerator) Reset(_ int64) error {
	if e == nil {
		panic(NewNullPointerException())
	}
	e.idx = 0
	return nil
}

type ConcurrentTraceRecord struct {
	Ptr     int64
	Worker  int
	FP      uint64
	Workers []*Worker

	ownersCaptured bool
	ptrOwner       int64
	workerOwner    int
	fpOwner        uint64
}

func NewConcurrentTraceRecord(ptr int64, worker int, fp uint64) ConcurrentTraceRecord {
	worker = int(int32(worker))
	return ConcurrentTraceRecord{
		Ptr: ptr, Worker: worker, FP: fp,
		ownersCaptured: true, ptrOwner: ptr, workerOwner: worker, fpOwner: fp,
	}
}

func (r ConcurrentTraceRecord) ownedPtr() int64 {
	if r.ownersCaptured {
		return r.ptrOwner
	}
	return r.Ptr
}

func (r ConcurrentTraceRecord) ownedWorker() int {
	if r.ownersCaptured {
		return r.workerOwner
	}
	return r.Worker
}

func (r ConcurrentTraceRecord) ownedFP() uint64 {
	if r.ownersCaptured {
		return r.fpOwner
	}
	return r.FP
}

func (r ConcurrentTraceRecord) IsInitial() bool {
	return r.ownedPtr() == 1
}

func (r ConcurrentTraceRecord) GetWorker() *Worker {
	if r.Workers == nil {
		panic(NewNullPointerException())
	}
	worker := r.ownedWorker()
	if worker < 0 || worker >= len(r.Workers) {
		panic(NewArrayIndexOutOfBoundsException(worker, len(r.Workers)))
	}
	return r.Workers[worker]
}

func (r ConcurrentTraceRecord) String() string {
	return fmt.Sprintf("Record [ptr=%d, worker=%d, fp=%d, initial=%t]", r.ownedPtr(), r.ownedWorker(), int64(r.ownedFP()), r.IsInitial())
}

func (r ConcurrentTraceRecord) GetPredecessor() (ConcurrentTraceRecord, error) {
	worker := r.GetWorker()
	if worker == nil {
		panic(NewNullPointerException())
	}
	predecessor, err := worker.ReadStateRecord(r.ownedPtr())
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	predecessor.Workers = r.Workers
	return predecessor, nil
}

func concurrentTraceRecordForState(state *TLCStateMut, workers []*Worker) (ConcurrentTraceRecord, error) {
	if state == nil || workers == nil {
		panic(NewNullPointerException())
	}
	workerID := int(state.WorkerID)
	if workerID < 0 || workerID >= len(workers) {
		panic(NewArrayIndexOutOfBoundsException(workerID, len(workers)))
	}
	worker := workers[workerID]
	if worker == nil {
		panic(NewNullPointerException())
	}
	record, err := worker.ReadStateRecord(state.UID)
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	record.Workers = workers
	return record, nil
}

func concurrentTracePredecessorForState(state *TLCStateMut, workers []*Worker) (ConcurrentTraceRecord, error) {
	record, err := concurrentTraceRecordForState(state, workers)
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	return record.GetPredecessor()
}
