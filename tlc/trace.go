package tlc

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type TraceRecord struct {
	PreviousUID int64
	WorkerID    int16
	FP          uint64
	State       *TLCStateMut
	Action      *Action
}

type TLCTrace struct {
	mu       sync.Mutex
	records  []TraceRecord
	level    int
	diskdir  string
	rootName string
}

func NewTLCTrace(metaDir ...string) *TLCTrace {
	trace := &TLCTrace{}
	if len(metaDir) > 0 {
		trace.diskdir = metaDir[0]
	}
	if len(metaDir) > 1 {
		trace.rootName = metaDir[1]
	}
	return trace
}

func (t *TLCTrace) SetCheckpointContext(metadir string, rootName string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if metadir != "" {
		t.diskdir = metadir
	}
	if rootName != "" {
		t.rootName = rootName
	}
}

func (t *TLCTrace) WriteInitState(state *TLCStateMut, fp uint64) error {
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

func (t *TLCTrace) WriteNextState(curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action) error {
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

func (t *TLCTrace) Records() []TraceRecord {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TraceRecord, len(t.records))
	copy(out, t.records)
	return out
}

func (t *TLCTrace) RecordFor(state *TLCStateMut) (TraceRecord, bool) {
	if t == nil || state == nil || state.UID < 0 {
		return TraceRecord{}, false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if state.UID >= int64(len(t.records)) {
		return TraceRecord{}, false
	}
	record := t.records[state.UID]
	return record, record.State == state
}

func (t *TLCTrace) GetTrace(state *TLCStateMut) []*TLCStateInfo {
	if state == nil {
		return nil
	}
	var reversed []*TLCStateInfo
	for cur := state; cur != nil; cur = cur.Predecessor() {
		info := NewTLCStateInfo(cur)
		fp := cur.FingerPrint()
		info.FP = &fp
		reversed = append(reversed, info)
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (t *TLCTrace) GetTraceBetween(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	if to == nil {
		return nil
	}
	if from == nil || from.Equal(to) {
		return t.GetTrace(to)
	}
	var reversed []*TLCStateInfo
	for cur := to; cur != nil; cur = cur.Predecessor() {
		info := NewTLCStateInfo(cur)
		fp := cur.FingerPrint()
		info.FP = &fp
		reversed = append(reversed, info)
		if cur.Equal(from) {
			break
		}
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (t *TLCTrace) GetTraceAt(pos int64, included bool) []*TLCStateInfo {
	if t == nil || pos < 0 {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if pos >= int64(len(t.records)) {
		return nil
	}
	state := t.records[pos].State
	if !included && state != nil {
		state = state.Predecessor()
	}
	return traceFromState(state)
}

func (t *TLCTrace) GetLevelForReporting() int {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.level
}

func (t *TLCTrace) Elements() *TLCTraceEnumerator {
	return &TLCTraceEnumerator{records: t.Records()}
}

func (t *TLCTrace) BeginChkpt() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.diskdir == "" {
		dir, err := os.MkdirTemp("", "TLCTrace")
		if err != nil {
			return err
		}
		t.diskdir = dir
	}
	path := t.chkptName("tmp")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	if err := out.WriteInt(int32(len(t.records))); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.WriteInt(int32(t.level)); err != nil {
		_ = out.Close()
		return err
	}
	for _, record := range t.records {
		if err := out.WriteLong(record.PreviousUID); err != nil {
			_ = out.Close()
			return err
		}
		if err := out.WriteShort(record.WorkerID); err != nil {
			_ = out.Close()
			return err
		}
		if err := out.WriteLong(int64(record.FP)); err != nil {
			_ = out.Close()
			return err
		}
		if record.State == nil {
			if err := out.WriteBool(false); err != nil {
				_ = out.Close()
				return err
			}
		} else {
			if err := out.WriteBool(true); err != nil {
				_ = out.Close()
				return err
			}
			if err := record.State.Write(out); err != nil {
				_ = out.Close()
				return err
			}
		}
		if record.Action == nil {
			if err := out.WriteBool(false); err != nil {
				_ = out.Close()
				return err
			}
		} else {
			if err := out.WriteBool(true); err != nil {
				_ = out.Close()
				return err
			}
			if err := out.WriteUniqueString(UniqueStringOf(record.Action.GetName())); err != nil {
				_ = out.Close()
				return err
			}
		}
	}
	return out.Close()
}

func (t *TLCTrace) CommitChkpt() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.diskdir == "" {
		return nil
	}
	oldChkpt := t.chkptName("chkpt")
	newChkpt := t.chkptName("tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(newChkpt, oldChkpt)
}

func (t *TLCTrace) Recover() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.diskdir == "" {
		return nil
	}
	file, err := os.Open(t.chkptName("chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	length, err := in.ReadInt()
	if err != nil {
		_ = in.Close()
		return err
	}
	level, err := in.ReadInt()
	if err != nil {
		_ = in.Close()
		return err
	}
	records := make([]TraceRecord, int(length))
	for i := range records {
		prev, err := in.ReadLong()
		if err != nil {
			_ = in.Close()
			return err
		}
		workerID, err := in.ReadShort()
		if err != nil {
			_ = in.Close()
			return err
		}
		fp, err := in.ReadLong()
		if err != nil {
			_ = in.Close()
			return err
		}
		hasState, err := in.ReadBool()
		if err != nil {
			_ = in.Close()
			return err
		}
		var state *TLCStateMut
		if hasState {
			state = NewEmptyState()
			if err := state.Read(in); err != nil {
				_ = in.Close()
				return err
			}
			state.UID = int64(i)
			state.WorkerID = workerID
		}
		hasAction, err := in.ReadBool()
		if err != nil {
			_ = in.Close()
			return err
		}
		var action *Action
		if hasAction {
			name, err := in.readExternalUniqueString()
			if err != nil {
				_ = in.Close()
				return err
			}
			action = &Action{Name: name.String()}
			if state != nil {
				state.SetAction(action)
			}
		}
		records[i] = TraceRecord{
			PreviousUID: prev,
			WorkerID:    workerID,
			FP:          uint64(fp),
			State:       state,
			Action:      action,
		}
	}
	if err := in.Close(); err != nil {
		return err
	}
	for i := range records {
		record := &records[i]
		if record.State == nil || record.PreviousUID < 0 || record.PreviousUID >= int64(len(records)) || record.PreviousUID == int64(i) {
			continue
		}
		pred := records[record.PreviousUID].State
		if pred != nil && record.State.Level() > pred.Level() {
			record.State.pred = pred
		}
	}
	t.records = records
	t.level = int(level)
	return nil
}

func (t *TLCTrace) Delete() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.diskdir == "" {
		return nil
	}
	if err := os.Remove(t.chkptName("tmp")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Remove(t.chkptName("chkpt")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (t *TLCTrace) chkptName(ext string) string {
	rootName := t.rootName
	if rootName == "" {
		rootName = "Spec"
	}
	return filepath.Join(t.diskdir, rootName+".st."+ext)
}

func (t *TLCTrace) Close() error {
	return nil
}

type TLCTraceEnumerator struct {
	records []TraceRecord
	index   int
}

func (e *TLCTraceEnumerator) NextPos() int64 {
	if e == nil || e.index >= len(e.records) {
		return -1
	}
	return int64(e.index)
}

func (e *TLCTraceEnumerator) NextFP() uint64 {
	if e == nil || e.index >= len(e.records) {
		return 0
	}
	fp := e.records[e.index].FP
	e.index++
	return fp
}

func (e *TLCTraceEnumerator) Close() error {
	return nil
}

func (e *TLCTraceEnumerator) Reset(pos int64) {
	if e == nil {
		return
	}
	if pos < 0 {
		e.index = 0
		return
	}
	if pos > int64(len(e.records)) {
		pos = int64(len(e.records))
	}
	e.index = int(pos)
}

func traceFromState(state *TLCStateMut) []*TLCStateInfo {
	if state == nil {
		return nil
	}
	var reversed []*TLCStateInfo
	for cur := state; cur != nil; cur = cur.Predecessor() {
		info := NewTLCStateInfo(cur)
		fp := cur.FingerPrint()
		info.FP = &fp
		reversed = append(reversed, info)
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}
