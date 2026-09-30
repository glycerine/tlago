package tlc

import "os"

type ConcurrentTLCTrace struct {
	*TLCTrace
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
	if t == nil || t.TLCTrace == nil {
		trace := NewTLCTrace()
		return trace.GetTrace(state)
	}
	return t.TLCTrace.GetTrace(state)
}

func (t *ConcurrentTLCTrace) GetTraceBetweenStates(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	if to == nil {
		return nil
	}
	if t == nil || t.TLCTrace == nil {
		trace := NewTLCTrace()
		return trace.GetTraceBetween(from, to)
	}
	return t.TLCTrace.GetTraceBetween(from, to)
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

type ConcurrentTraceRecord struct {
	Ptr    int64
	Worker int
	FP     uint64
}

func NewConcurrentTraceRecord(ptr int64, worker int, fp uint64) ConcurrentTraceRecord {
	return ConcurrentTraceRecord{Ptr: ptr, Worker: worker, FP: fp}
}

func (r ConcurrentTraceRecord) IsInitial() bool {
	return r.Ptr == 1
}
