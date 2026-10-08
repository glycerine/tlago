package tlc

import (
	"errors"
	"fmt"
	"os"
)

type ConcurrentTLCTrace struct {
	*TLCTrace
	Tool    *Tool
	Workers []*Worker
}

func NewConcurrentTLCTrace(metadir string, specFile string, workerCount ...int) *ConcurrentTLCTrace {
	count := NumWorkers()
	if len(workerCount) > 0 && workerCount[0] > 0 {
		count = workerCount[0]
	}
	if count <= 0 {
		count = 1
	}
	return &ConcurrentTLCTrace{
		TLCTrace: NewTLCTrace(metadir, specFile),
		Workers:  make([]*Worker, count),
	}
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
	for _, worker := range t.Workers {
		if worker != nil {
			worker.SetTraceContext(metadir, rootName)
		}
	}
}

func (t *ConcurrentTLCTrace) AddWorker(worker *Worker) *Worker {
	if t == nil || worker == nil {
		return worker
	}
	id := worker.MyGetID()
	if id < 0 {
		return worker
	}
	for id >= len(t.Workers) {
		t.Workers = append(t.Workers, nil)
	}
	t.Workers[id] = worker
	return worker
}

func (t *ConcurrentTLCTrace) GetLevelForReporting() int {
	return t.GetLevel()
}

func (t *ConcurrentTLCTrace) GetLevel() int {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	maxLevel := 1
	for _, worker := range t.Workers {
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
	if curState == nil {
		return
	}
	var prefix []*TLCStateInfo
	if !curState.IsInitial() {
		prefix = t.GetTraceFromState(curState)
	}
	t.printTraceWithPrefix(curState, succState, prefix)
}

func (t *ConcurrentTLCTrace) GetTraceFromState(state *TLCStateMut) []*TLCStateInfo {
	if state == nil {
		panic(NewNullPointerException())
	}
	if state.IsInitial() {
		return []*TLCStateInfo{NewTLCStateInfo(state)}
	}
	trace, err := t.recoverTrace(state, nil)
	if err != nil {
		panic(err)
	}
	return trace
}

func (t *ConcurrentTLCTrace) GetTraceBetweenStates(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
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
	trace, err := t.recoverTrace(to, from)
	if err != nil {
		panic(err)
	}
	return trace
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
	t.mu.Lock()
	defer t.mu.Unlock()
	records, err := t.collectTraceRecordsLocked(state, from, first)
	if err != nil || len(records) == 0 {
		return nil, err
	}
	if from != nil {
		return t.recoverTraceFromRecords(NewTLCStateInfo(from), records)
	}
	return t.recoverTraceFromRecords(nil, records)
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
		if (from == nil && record.IsInitial()) || (from != nil && record.FP == from.FingerPrint()) {
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
	if t == nil || state == nil {
		panic(NewNullPointerException())
	}
	if state.WorkerID < 0 || int(state.WorkerID) >= len(t.Workers) {
		panic(NewArrayIndexOutOfBoundsException(int(state.WorkerID), len(t.Workers)))
	}
	worker := t.Workers[state.WorkerID]
	if worker == nil {
		panic(NewNullPointerException())
	}
	record, err := worker.ReadStateRecord(state.UID)
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	record.Workers = t.Workers
	return record, nil
}

func (t *ConcurrentTLCTrace) predecessorRecord(record ConcurrentTraceRecord) (ConcurrentTraceRecord, error) {
	worker := record.GetWorker()
	if worker == nil {
		panic(NewNullPointerException())
	}
	pred, err := worker.ReadStateRecord(record.Ptr)
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	pred.Workers = record.Workers
	return pred, nil
}

func (t *ConcurrentTLCTrace) recoverTraceFromRecords(sinfo *TLCStateInfo, records []ConcurrentTraceRecord) ([]*TLCStateInfo, error) {
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
			if t.Tool == nil {
				panic(NewNullPointerException())
			}
			info, err := t.Tool.GetState(records[end].FP)
			if err != nil {
				return nil, err
			}
			sinfo = info
			if sinfo == nil || sinfo.State == nil {
				panic(NewNullPointerException())
			}
			prev := records[end-1]
			sinfo.State.WorkerID = int16(prev.Worker)
			sinfo.State.UID = prev.Ptr
		}
		out = append(out, sinfo)
		for i := end - 2; i >= 0; i-- {
			record := records[i+1]
			if t.Tool == nil {
				panic(NewNullPointerException())
			}
			info, err := t.Tool.GetState(record.FP, sinfo.State)
			if err != nil {
				return nil, err
			}
			if info == nil {
				traceRecoveryExit(fmt.Sprintf("2 %d", int64(record.FP)), nil)
			}
			prev := records[i]
			info.State.WorkerID = int16(prev.Worker)
			info.State.UID = prev.Ptr
			out = append(out, info)
			sinfo = info
		}
	}
	SetRandomEnumerableGenerator(snapshot)
	return out, nil
}

func (t *ConcurrentTLCTrace) CommitChkpt() error {
	if t == nil {
		panic(NewNullPointerException())
	}
	for _, worker := range t.Workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if err := worker.CommitChkpt(); err != nil {
			return err
		}
	}
	if t.TLCTrace == nil || t.diskdir == "" {
		return nil
	}
	file, err := os.OpenFile(t.chkptName("chkpt"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return file.Close()
}

func (t *ConcurrentTLCTrace) BeginChkpt() error {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, worker := range t.Workers {
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
	for _, worker := range t.Workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if err := worker.RecoverTrace(); err != nil {
			return err
		}
	}
	return nil
}

func (t *ConcurrentTLCTrace) Elements() (*ConcurrentTraceEnumerator, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	enums := make([]*WorkerTraceEnumerator, len(t.Workers))
	for i, worker := range t.Workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		enum, err := worker.Elements()
		if err != nil {
			return nil, err
		}
		enums[i] = enum
	}
	return &ConcurrentTraceEnumerator{enums: enums}, nil
}

func (t *ConcurrentTLCTrace) Close() error {
	if t == nil {
		return nil
	}
	var err error
	for _, worker := range t.Workers {
		if worker != nil {
			if closeErr := worker.CloseTrace(); closeErr != nil && err == nil {
				err = closeErr
			}
		}
	}
	if t.TLCTrace != nil {
		if closeErr := t.TLCTrace.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func (t *ConcurrentTLCTrace) Delete() error {
	if t == nil {
		return nil
	}
	var err error
	for _, worker := range t.Workers {
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

func (e *ConcurrentTraceEnumerator) NextPos() int64 {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.idx >= len(e.enums) {
		return -1
	}
	reader, err := e.readerAtCurrentIndex()
	if err != nil {
		panic(err)
	}
	if reader.HasMoreFP() {
		return 42
	}
	if e.idx+1 >= len(e.enums) {
		return -1
	}
	e.idx++
	reader, err = e.readerAtCurrentIndex()
	if err != nil {
		panic(err)
	}
	if reader.HasMoreFP() {
		return 42
	}
	return -1
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
	reader, err := e.readerAtCurrentIndex()
	if err != nil {
		return 0, err
	}
	more, err := reader.HasMoreFPWithError()
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
		return NewNullPointerException()
	}
	for _, enum := range e.enums {
		if err := enum.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (e *ConcurrentTraceEnumerator) Reset(pos int64) {
	if e != nil {
		e.idx = 0
	}
}

type ConcurrentTraceRecord struct {
	Ptr     int64
	Worker  int
	FP      uint64
	Workers []*Worker
}

func NewConcurrentTraceRecord(ptr int64, worker int, fp uint64) ConcurrentTraceRecord {
	return ConcurrentTraceRecord{Ptr: ptr, Worker: worker, FP: fp}
}

func (r ConcurrentTraceRecord) IsInitial() bool {
	return r.Ptr == 1
}

func (r ConcurrentTraceRecord) GetWorker() *Worker {
	if r.Worker < 0 || r.Worker >= len(r.Workers) {
		panic(NewArrayIndexOutOfBoundsException(r.Worker, len(r.Workers)))
	}
	return r.Workers[r.Worker]
}
