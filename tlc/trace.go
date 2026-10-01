package tlc

import (
	"errors"
	"fmt"
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
	mu            sync.Mutex
	Tool          *Tool
	records       []TraceRecord
	level         int
	previousLevel int
	diskdir       string
	rootName      string
	raf           *BufferedRandomAccessFile
	lastPtr       int64
	traceErr      error
}

func NewTLCTrace(metaDir ...string) *TLCTrace {
	trace := &TLCTrace{lastPtr: 1}
	if len(metaDir) > 0 {
		trace.diskdir = metaDir[0]
	}
	if len(metaDir) > 1 {
		trace.rootName = metaDir[1]
	}
	if trace.diskdir != "" {
		trace.traceErr = trace.ensureTraceRAFLocked()
	}
	return trace
}

func (t *TLCTrace) SetTool(tool *Tool) {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.Tool = tool
	t.mu.Unlock()
}

func (t *TLCTrace) SetCheckpointContext(metadir string, rootName string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	oldName := ""
	if t.diskdir != "" {
		oldName = t.traceFileName()
	}
	if metadir != "" {
		t.diskdir = metadir
	}
	if rootName != "" {
		t.rootName = rootName
	}
	newName := ""
	if t.diskdir != "" {
		newName = t.traceFileName()
	}
	if oldName != newName {
		if t.raf != nil {
			if err := t.raf.Close(); err != nil {
				t.traceErr = err
				t.raf = nil
				return
			}
			t.raf = nil
		}
		t.traceErr = nil
	}
	t.traceErr = t.ensureTraceRAFLocked()
}

func (t *TLCTrace) ensureTraceRAFLocked() error {
	if t == nil || t.raf != nil {
		return nil
	}
	if t.traceErr != nil {
		return t.traceErr
	}
	if t.diskdir == "" {
		return nil
	}
	if err := os.MkdirAll(t.diskdir, 0o755); err != nil {
		t.traceErr = err
		return err
	}
	raf, err := NewBufferedRandomAccessFile(t.traceFileName(), "rw")
	if err != nil {
		t.traceErr = err
		return err
	}
	t.raf = raf
	return nil
}

func (t *TLCTrace) traceFileName() string {
	rootName := t.rootName
	if rootName == "" {
		rootName = "Spec"
	}
	return filepath.Join(t.diskdir, rootName+tlcTraceExt)
}

func (t *TLCTrace) WriteInitState(state *TLCStateMut, fp uint64) error {
	_, err := t.WriteState(nil, fp, state, nil)
	return err
}

func (t *TLCTrace) WriteState(predecessor *TLCStateMut, fp uint64, state *TLCStateMut, action *Action) (int64, error) {
	if t == nil {
		return TLCStateInitUID, nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	uid := int64(len(t.records))
	prevUID := int64(1)
	if predecessor != nil {
		prevUID = predecessor.UID
	}
	if err := t.ensureTraceRAFLocked(); err != nil {
		return TLCStateInitUID, err
	}
	if t.raf != nil {
		ptr, err := t.raf.GetFilePointer()
		if err != nil {
			return TLCStateInitUID, err
		}
		uid = ptr
		if err := t.raf.WriteLongNat(prevUID); err != nil {
			return TLCStateInitUID, err
		}
		if err := t.raf.WriteLong(int64(fp)); err != nil {
			return TLCStateInitUID, err
		}
		t.lastPtr = ptr
	}
	t.records = append(t.records, TraceRecord{
		PreviousUID: prevUID,
		WorkerID:    0,
		FP:          fp,
		State:       state,
		Action:      action,
	})
	if state != nil {
		state.WorkerID = 0
		state.UID = uid
		state.SetPredecessor(predecessor)
		state.SetAction(action)
		if state.Level() > t.level {
			t.level = state.Level()
		}
	}
	return uid, nil
}

func (t *TLCTrace) MirrorInitStateForWorker(workerID int, state *TLCStateMut, fp uint64, uid int64) {
	if t == nil || state == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	traceWorkerID := int16(workerID)
	if workerID < 0 || workerID > int(TLCStateInitWorkerID) {
		traceWorkerID = TLCStateInitWorkerID
	}
	t.records = append(t.records, TraceRecord{
		PreviousUID: 1,
		WorkerID:    traceWorkerID,
		FP:          fp,
		State:       state,
	})
	state.WorkerID = traceWorkerID
	state.UID = uid
	if state.Level() > t.level {
		t.level = state.Level()
	}
}

func (t *TLCTrace) WriteNextState(curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action) error {
	return t.WriteNextStateForWorker(0, curState, succFP, succState, action)
}

func (t *TLCTrace) WriteNextStateForWorker(workerID int, curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action) error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	uid := int64(len(t.records))
	prevUID := TLCStateInitUID
	predecessorWorkerID := int16(0)
	if curState != nil {
		prevUID = curState.UID
		predecessorWorkerID = curState.WorkerID
	}
	if err := t.ensureTraceRAFLocked(); err != nil {
		return err
	}
	if t.raf != nil {
		ptr, err := t.raf.GetFilePointer()
		if err != nil {
			return err
		}
		uid = ptr
		if err := t.raf.WriteLongNat(prevUID); err != nil {
			return err
		}
		if err := t.raf.WriteLong(int64(succFP)); err != nil {
			return err
		}
		t.lastPtr = ptr
	}
	generatedWorkerID := int16(workerID)
	if workerID < 0 || workerID > int(TLCStateInitWorkerID) {
		generatedWorkerID = TLCStateInitWorkerID
	}
	t.records = append(t.records, TraceRecord{
		PreviousUID: prevUID,
		WorkerID:    predecessorWorkerID,
		FP:          succFP,
		State:       succState,
		Action:      action,
	})
	succState.WorkerID = generatedWorkerID
	succState.UID = uid
	succState.SetPredecessor(curState)
	succState.SetAction(action)
	if succState.Level() > t.level {
		t.level = succState.Level()
	}
	return nil
}

func (t *TLCTrace) MirrorNextStateForWorker(workerID int, curState *TLCStateMut, succFP uint64, succState *TLCStateMut, action *Action, uid int64) {
	if t == nil || succState == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	prevUID := TLCStateInitUID
	predecessorWorkerID := int16(0)
	if curState != nil {
		prevUID = curState.UID
		predecessorWorkerID = curState.WorkerID
	}
	generatedWorkerID := int16(workerID)
	if workerID < 0 || workerID > int(TLCStateInitWorkerID) {
		generatedWorkerID = TLCStateInitWorkerID
	}
	t.records = append(t.records, TraceRecord{
		PreviousUID: prevUID,
		WorkerID:    predecessorWorkerID,
		FP:          succFP,
		State:       succState,
		Action:      action,
	})
	succState.WorkerID = generatedWorkerID
	succState.UID = uid
	succState.SetPredecessor(curState)
	succState.SetAction(action)
	if succState.Level() > t.level {
		t.level = succState.Level()
	}
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
	for _, record := range t.records {
		if record.State == state {
			return record, true
		}
	}
	return TraceRecord{}, false
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
	trace, recovered, err := t.getTraceAtFromDisk(pos, included)
	if err != nil {
		panic(err)
	}
	if recovered {
		return trace
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	var state *TLCStateMut
	if pos >= 0 && pos < int64(len(t.records)) {
		state = t.records[pos].State
	}
	if state == nil {
		if idx := t.recordIndexByUIDLocked(pos); idx != -1 {
			state = t.records[idx].State
		}
	}
	if state == nil {
		return nil
	}
	if !included && state != nil {
		state = state.Predecessor()
	}
	return traceFromState(state)
}

func (t *TLCTrace) getTraceAtFromDisk(pos int64, included bool) ([]*TLCStateInfo, bool, error) {
	if t == nil || t.Tool == nil {
		return nil, false, nil
	}
	fps, err := t.traceFPsFromDisk(pos, included)
	if err != nil {
		return nil, false, err
	}
	if fps == nil {
		return nil, false, nil
	}
	if len(fps) == 0 {
		return []*TLCStateInfo{}, true, nil
	}
	trace, err := t.recoverTraceFromFPs(nil, fps)
	return trace, err == nil, err
}

func (t *TLCTrace) traceFPsFromDisk(pos int64, included bool) ([]uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.ensureTraceRAFLocked(); err != nil {
		return nil, err
	}
	if t.raf == nil {
		return nil, nil
	}
	current, err := t.raf.GetFilePointer()
	if err != nil {
		return nil, err
	}
	defer func() { _ = t.raf.Seek(current) }()
	loc := pos
	if !included {
		loc, err = t.getPrevFromDiskLocked(pos)
		if err != nil {
			return nil, err
		}
	}
	fps := make([]uint64, 0)
	for predecessorLoc := loc; predecessorLoc != 1; {
		fp, err := t.getFPFromDiskLocked(predecessorLoc)
		if err != nil {
			return nil, err
		}
		fps = append(fps, fp)
		next, err := t.getPrevFromDiskLocked(predecessorLoc)
		if err != nil {
			return nil, err
		}
		if next == predecessorLoc {
			break
		}
		predecessorLoc = next
	}
	return fps, nil
}

func (t *TLCTrace) recoverTraceFromFPs(sinfo *TLCStateInfo, fps []uint64) ([]*TLCStateInfo, error) {
	if t == nil || t.Tool == nil || len(fps) == 0 {
		return nil, nil
	}
	snapshot := ResetRandomEnumerableValues()
	defer SetRandomEnumerableGenerator(snapshot)
	out := make([]*TLCStateInfo, 0, len(fps))
	if sinfo == nil {
		fp := fps[len(fps)-1]
		info, err := t.Tool.GetState(fp)
		if err != nil {
			return nil, err
		}
		if info == nil {
			return nil, newTLCError(ECTLCFailedToRecoverInit, "initial state fingerprint %d could not be regenerated", fp)
		}
		info.FP = &fp
		sinfo = info
	}
	out = append(out, sinfo)
	for i := len(fps) - 2; i >= 0; i-- {
		fp := fps[i]
		info, err := t.Tool.GetState(fp, sinfo)
		if err != nil {
			return nil, err
		}
		if info == nil {
			return nil, newTLCError(ECTLCFailedToRecoverNext, "successor fingerprint %d could not be regenerated", fp)
		}
		info.FP = &fp
		out = append(out, info)
		sinfo = info
	}
	return out, nil
}

func (t *TLCTrace) GetLevelForReporting() int {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.raf != nil && t.lastPtr != 1 {
		level, err := t.getLevelFromDiskLocked(t.lastPtr)
		if err == nil && level > t.previousLevel {
			t.previousLevel = level
		}
		return t.previousLevel
	}
	return t.level
}

func (t *TLCTrace) GetLevel(startUID int64) int {
	if t == nil || startUID < 0 {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.raf != nil {
		level, err := t.getLevelFromDiskLocked(startUID)
		if err == nil {
			return level
		}
	}
	return t.getLevelLocked(startUID)
}

func (t *TLCTrace) GetLevelForState(state *TLCStateMut) int {
	if t == nil || state == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for i := range t.records {
		if t.records[i].State == state {
			return t.getLevelLocked(t.records[i].State.UID)
		}
	}
	return 0
}

func (t *TLCTrace) getLevelLocked(startUID int64) int {
	level := 0
	for uid := startUID; uid >= 0; {
		idx := t.recordIndexByUIDLocked(uid)
		if idx == -1 {
			break
		}
		level++
		prev := t.records[idx].PreviousUID
		if prev == 1 {
			break
		}
		if prev < 0 || prev == uid {
			break
		}
		uid = prev
	}
	return level
}

func (t *TLCTrace) getLevelFromDiskLocked(startLoc int64) (int, error) {
	if t == nil || t.raf == nil {
		return 0, nil
	}
	current, err := t.raf.GetFilePointer()
	if err != nil {
		return 0, err
	}
	defer func() { _ = t.raf.Seek(current) }()
	level := 0
	for predecessorLoc := startLoc; predecessorLoc != 1; {
		level++
		if err := t.raf.Seek(predecessorLoc); err != nil {
			return 0, err
		}
		next, err := t.raf.ReadLongNat()
		if err != nil {
			return 0, err
		}
		if next == predecessorLoc {
			break
		}
		predecessorLoc = next
	}
	return level, nil
}

func (t *TLCTrace) getPrevFromDiskLocked(loc int64) (int64, error) {
	if t == nil || t.raf == nil {
		return 0, nil
	}
	if err := t.raf.Seek(loc); err != nil {
		return 0, err
	}
	return t.raf.ReadLongNat()
}

func (t *TLCTrace) getFPFromDiskLocked(loc int64) (uint64, error) {
	if t == nil || t.raf == nil {
		return 0, nil
	}
	if err := t.raf.Seek(loc); err != nil {
		return 0, err
	}
	if _, err := t.raf.ReadLongNat(); err != nil {
		return 0, err
	}
	fp, err := t.raf.ReadLong()
	if err != nil {
		return 0, err
	}
	return uint64(fp), nil
}

func (t *TLCTrace) recordIndexByUIDLocked(uid int64) int {
	for i := range t.records {
		if t.records[i].State != nil && t.records[i].State.UID == uid {
			return i
		}
	}
	return -1
}

func (t *TLCTrace) Elements() *TLCTraceEnumerator {
	if t == nil {
		return &TLCTraceEnumerator{}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.ensureTraceRAFLocked(); err == nil && t.raf != nil {
		if err := t.raf.Flush(); err == nil {
			length, lenErr := t.raf.Length()
			if lenErr == nil {
				enumRAF, openErr := NewBufferedRandomAccessFile(t.traceFileName(), "r")
				if openErr == nil {
					return &TLCTraceEnumerator{length: length, raf: enumRAF, path: t.traceFileName()}
				}
			}
		}
	}
	out := make([]TraceRecord, len(t.records))
	copy(out, t.records)
	return &TLCTraceEnumerator{records: out}
}

func (t *TLCTrace) BeginChkpt() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.diskdir != "" {
		if err := t.ensureTraceRAFLocked(); err != nil {
			return err
		}
		if t.raf != nil {
			if err := t.raf.Flush(); err != nil {
				return err
			}
			file, err := os.Create(t.chkptName("tmp"))
			if err != nil {
				return err
			}
			out := NewValueOutputStream(file)
			filePtr, err := t.raf.GetFilePointer()
			if err != nil {
				_ = out.Close()
				return err
			}
			if err := out.WriteLong(filePtr); err != nil {
				_ = out.Close()
				return err
			}
			if err := out.WriteLong(t.lastPtr); err != nil {
				_ = out.Close()
				return err
			}
			return out.Close()
		}
	}
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
		return fmt.Errorf("Trace.commitChkpt: cannot delete %s", oldChkpt)
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return fmt.Errorf("Trace.commitChkpt: cannot delete %s", oldChkpt)
	}
	return nil
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
	if err := t.ensureTraceRAFLocked(); err != nil {
		return err
	}
	file, err := os.Open(t.chkptName("chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	if t.raf != nil {
		filePos, err := in.ReadLong()
		if err != nil {
			_ = in.Close()
			return err
		}
		lastPtr, err := in.ReadLong()
		if err != nil {
			_ = in.Close()
			return err
		}
		if err := in.Close(); err != nil {
			return err
		}
		t.lastPtr = lastPtr
		return t.raf.Seek(filePos)
	}
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
			name, err := readJavaUniqueString(in)
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
	if t.raf != nil {
		if err := t.raf.Close(); err != nil {
			return err
		}
		t.raf = nil
	}
	if err := os.Remove(t.traceFileName()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
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
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.raf == nil {
		return nil
	}
	err := t.raf.Close()
	t.raf = nil
	return err
}

type TLCTraceEnumerator struct {
	records []TraceRecord
	index   int
	length  int64
	raf     *BufferedRandomAccessFile
	path    string
}

func (e *TLCTraceEnumerator) NextPos() int64 {
	if e != nil && e.raf != nil {
		pos, err := e.raf.GetFilePointer()
		if err != nil || pos >= e.length {
			return -1
		}
		return pos
	}
	if e == nil || e.index >= len(e.records) {
		return -1
	}
	return int64(e.index)
}

func (e *TLCTraceEnumerator) NextFP() uint64 {
	if e != nil && e.raf != nil {
		if _, err := e.raf.ReadLongNat(); err != nil {
			return 0
		}
		fp, err := e.raf.ReadLong()
		if err != nil {
			return 0
		}
		return uint64(fp)
	}
	if e == nil || e.index >= len(e.records) {
		return 0
	}
	fp := e.records[e.index].FP
	e.index++
	return fp
}

func (e *TLCTraceEnumerator) Close() error {
	if e != nil && e.raf != nil {
		err := e.raf.Close()
		e.raf = nil
		return err
	}
	return nil
}

func (e *TLCTraceEnumerator) Reset(pos int64) {
	if e == nil {
		return
	}
	if e.raf != nil {
		if pos == -1 {
			pos, _ = e.raf.GetFilePointer()
		}
		path := e.path
		length := e.length
		_ = e.raf.Close()
		raf, err := NewBufferedRandomAccessFile(path, "r")
		if err != nil {
			e.raf = nil
			e.length = 0
			e.index = 0
			return
		}
		e.raf = raf
		e.length = length
		_ = e.raf.Seek(pos)
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

func TLCTraceWriteBehavior(fileName string, state *TLCStateMut, stateTrace *StateVec) error {
	_ = state
	dir := filepath.Dir(fileName)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	out := NewValueOutputStreamWithCompression(file, true)
	size := 0
	if stateTrace != nil {
		size = stateTrace.Size()
	}
	values := make([]Value, size)
	for i := 0; i < size; i++ {
		traceState := stateTrace.At(i)
		if traceState == nil {
			_ = out.Close()
			return newTLCError(ECSystemDiskIOErrorForFile, "%s", fileName)
		}
		values[i] = NewRecordValueFromInsMap(traceState.Values())
	}
	if err := out.WriteExternal(NewTupleValue(values)); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
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
