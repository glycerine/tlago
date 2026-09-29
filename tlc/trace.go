package tlc

import "sync"

type TraceRecord struct {
	PreviousUID int64
	WorkerID    int16
	FP          uint64
	State       *TLCStateMut
	Action      *Action
}

type MemoryTrace struct {
	mu      sync.Mutex
	records []TraceRecord
	level   int
}

func NewMemoryTrace() *MemoryTrace {
	return &MemoryTrace{}
}

func (t *MemoryTrace) WriteInitState(state *TLCStateMut, fp uint64) error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	uid := int64(len(t.records))
	t.records = append(t.records, TraceRecord{
		PreviousUID: 1,
		WorkerID:    0,
		FP:          fp,
		State:       state,
	})
	state.WorkerID = 0
	state.UID = uid
	if state.Level() > t.level {
		t.level = state.Level()
	}
	return nil
}

func (t *MemoryTrace) WriteNextState(curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action) error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	uid := int64(len(t.records))
	prevUID := TLCStateInitUID
	workerID := int16(0)
	if curState != nil {
		prevUID = curState.UID
		workerID = curState.WorkerID
	}
	t.records = append(t.records, TraceRecord{
		PreviousUID: prevUID,
		WorkerID:    workerID,
		FP:          succFP,
		State:       succState,
		Action:      action,
	})
	succState.WorkerID = workerID
	succState.UID = uid
	succState.SetPredecessor(curState)
	succState.SetAction(action)
	if succState.Level() > t.level {
		t.level = succState.Level()
	}
	return nil
}

func (t *MemoryTrace) Records() []TraceRecord {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TraceRecord, len(t.records))
	copy(out, t.records)
	return out
}

func (t *MemoryTrace) GetLevelForReporting() int {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.level
}

func (t *MemoryTrace) Close() error {
	return nil
}
