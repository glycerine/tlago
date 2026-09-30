package tlc

import (
	"errors"
	"os"
)

var errConcurrentTraceUnavailable = errors.New("concurrent trace record unavailable")

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
		return 0
	}
	maxLevel := 1
	for _, worker := range t.Workers {
		if worker != nil && worker.MaxLevel > maxLevel {
			maxLevel = worker.MaxLevel
		}
	}
	if maxLevel == 1 && t.TLCTrace != nil {
		if level := t.TLCTrace.GetLevelForReporting(); level > maxLevel {
			maxLevel = level
		}
	}
	return maxLevel
}

func (t *ConcurrentTLCTrace) GetTraceFromState(state *TLCStateMut) []*TLCStateInfo {
	if state == nil {
		return nil
	}
	if trace, err := t.recoverTrace(state, nil); err == nil && len(trace) > 0 {
		return trace
	}
	if t == nil || t.TLCTrace == nil {
		return traceFromState(state)
	}
	return t.TLCTrace.GetTrace(state)
}

func (t *ConcurrentTLCTrace) GetTraceBetweenStates(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	if to == nil {
		return nil
	}
	if trace, err := t.recoverTrace(to, from); err == nil && len(trace) > 0 {
		return trace
	}
	if t == nil || t.TLCTrace == nil {
		return NewTLCTrace().GetTraceBetween(from, to)
	}
	return t.TLCTrace.GetTraceBetween(from, to)
}

func (t *ConcurrentTLCTrace) recoverTrace(state *TLCStateMut, from *TLCStateMut) ([]*TLCStateInfo, error) {
	if t == nil || t.Tool == nil || state == nil || len(t.Workers) == 0 || state.WorkerID < 0 || int(state.WorkerID) >= len(t.Workers) {
		return nil, nil
	}
	if state.IsInitial() {
		return []*TLCStateInfo{NewTLCStateInfo(state)}, nil
	}
	records, err := t.collectTraceRecords(state, from)
	if errors.Is(err, errConcurrentTraceUnavailable) {
		return nil, nil
	}
	if err != nil || len(records) == 0 {
		return nil, err
	}
	if from != nil {
		return t.recoverTraceFromRecords(NewTLCStateInfo(from), records)
	}
	return t.recoverTraceFromRecords(nil, records)
}

func (t *ConcurrentTLCTrace) collectTraceRecords(state *TLCStateMut, from *TLCStateMut) ([]ConcurrentTraceRecord, error) {
	record, err := t.recordForState(state)
	if err != nil {
		return nil, err
	}
	if from != nil && state.Equal(from) {
		return []ConcurrentTraceRecord{record}, nil
	}
	records := []ConcurrentTraceRecord{record}
	for {
		pred, err := t.predecessorRecord(record)
		if err != nil {
			return nil, err
		}
		if pred.IsInitial() {
			records = append(records, pred)
			return records, nil
		}
		records = append(records, pred)
		if from != nil && pred.FP == from.FingerPrint() {
			return records, nil
		}
		record = pred
	}
}

func (t *ConcurrentTLCTrace) recordForState(state *TLCStateMut) (ConcurrentTraceRecord, error) {
	if state == nil || state.WorkerID < 0 || int(state.WorkerID) >= len(t.Workers) || t.Workers[state.WorkerID] == nil {
		return ConcurrentTraceRecord{}, errConcurrentTraceUnavailable
	}
	record, err := t.Workers[state.WorkerID].ReadStateRecord(state.UID)
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	record.Workers = t.Workers
	return record, nil
}

func (t *ConcurrentTLCTrace) predecessorRecord(record ConcurrentTraceRecord) (ConcurrentTraceRecord, error) {
	if record.IsInitial() {
		return record, nil
	}
	worker := record.GetWorker()
	if worker == nil {
		return ConcurrentTraceRecord{}, errConcurrentTraceUnavailable
	}
	pred, err := worker.ReadStateRecord(record.Ptr)
	if err != nil {
		return ConcurrentTraceRecord{}, err
	}
	pred.Workers = record.Workers
	return pred, nil
}

func (t *ConcurrentTLCTrace) recoverTraceFromRecords(sinfo *TLCStateInfo, records []ConcurrentTraceRecord) ([]*TLCStateInfo, error) {
	if t == nil || t.Tool == nil || len(records) == 0 {
		return nil, nil
	}
	end := len(records) - 1
	if sinfo == nil {
		initRecord := records[end]
		info, err := t.Tool.GetState(initRecord.FP)
		if err != nil || info == nil {
			return nil, err
		}
		sinfo = info
		if end > 0 {
			prev := records[end-1]
			sinfo.State.WorkerID = int16(prev.Worker)
			sinfo.State.UID = prev.Ptr
		}
	}
	out := make([]*TLCStateInfo, 0, end+1)
	out = append(out, sinfo)
	for i := end - 2; i >= 0; i-- {
		record := records[i+1]
		info, err := t.Tool.GetState(record.FP, sinfo.State)
		if err != nil || info == nil {
			return nil, err
		}
		prev := records[i]
		info.State.WorkerID = int16(prev.Worker)
		info.State.UID = prev.Ptr
		out = append(out, info)
		sinfo = info
	}
	return out, nil
}

func (t *ConcurrentTLCTrace) CommitChkpt() error {
	if t == nil {
		return nil
	}
	for _, worker := range t.Workers {
		if worker != nil {
			if err := worker.CommitChkpt(); err != nil {
				return err
			}
		}
	}
	if t.TLCTrace != nil {
		if err := t.TLCTrace.CommitChkpt(); err != nil {
			return err
		}
	}
	if t.TLCTrace == nil || t.diskdir == "" {
		return nil
	}
	file, err := os.OpenFile(t.chkptName("chkpt"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}

func (t *ConcurrentTLCTrace) BeginChkpt() error {
	if t == nil {
		return nil
	}
	for _, worker := range t.Workers {
		if worker != nil {
			if err := worker.BeginChkpt(); err != nil {
				return err
			}
		}
	}
	if t.TLCTrace != nil {
		return t.TLCTrace.BeginChkpt()
	}
	return nil
}

func (t *ConcurrentTLCTrace) Recover() error {
	if t == nil {
		return nil
	}
	for _, worker := range t.Workers {
		if worker != nil {
			if err := worker.RecoverTrace(); err != nil {
				return err
			}
		}
	}
	if t.TLCTrace != nil {
		return t.TLCTrace.Recover()
	}
	return nil
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
		return nil
	}
	return r.Workers[r.Worker]
}
