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
	rawPaths      bool
	rootName      string
	raf           *BufferedRandomAccessFile
	lastPtr       int64
	traceErr      error
	closed        bool
}

func NewTLCTrace(metaDir ...string) *TLCTrace {
	trace := &TLCTrace{lastPtr: 1}
	if len(metaDir) > 0 {
		trace.diskdir = metaDir[0]
	}
	if len(metaDir) > 1 {
		trace.rootName = metaDir[1]
		trace.rawPaths = true
	}
	if trace.diskdir != "" || trace.rawPaths {
		// TLCTrace opens its RAF during construction; a failed open never
		// produces a usable trace or reaches fingerprint initialization.
		if err := trace.ensureTraceRAFLocked(); err != nil {
			panic(err)
		}
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
	if t.diskdir == "" && !t.rawPaths {
		return nil
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
	if t.rawPaths {
		return t.diskdir + string(os.PathSeparator) + t.rootName + tlcTraceExt
	}
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
	return t.writeState(predecessor, fp, state, action, true)
}

func (t *TLCTrace) WriteStateRecord(predecessor *TLCStateMut, fp uint64, state *TLCStateMut) (int64, error) {
	return t.writeState(predecessor, fp, state, nil, false)
}

func (t *TLCTrace) writeState(predecessor *TLCStateMut, fp uint64, state *TLCStateMut, action *Action, attachMetadata bool) (int64, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	uid := int64(len(t.records))
	prevUID := int64(1)
	if predecessor != nil {
		prevUID = predecessor.UID
	}
	if t.rawPaths {
		if t.raf == nil {
			panic(NewNullPointerException())
		}
	} else {
		if err := t.ensureTraceRAFLocked(); err != nil {
			return TLCStateInitUID, err
		}
	}
	if t.raf != nil {
		ptr, err := t.raf.GetFilePointer()
		if err != nil {
			return TLCStateInitUID, err
		}
		uid = ptr
		t.lastPtr = ptr
		if err := t.raf.WriteLongNat(prevUID); err != nil {
			return TLCStateInitUID, err
		}
		if err := t.raf.WriteLong(int64(fp)); err != nil {
			return TLCStateInitUID, err
		}
	}
	t.records = append(t.records, TraceRecord{
		PreviousUID: prevUID,
		WorkerID:    0,
		FP:          fp,
		State:       state,
		Action:      action,
	})
	if state != nil {
		state.UID = uid
		if attachMetadata {
			state.WorkerID = 0
			state.attachTraceMetadata(predecessor, action)
		}
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
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if curState == nil {
		panic(NewNullPointerException())
	}
	uid := int64(len(t.records))
	prevUID := curState.UID
	predecessorWorkerID := curState.WorkerID
	if t.rawPaths {
		if t.raf == nil {
			panic(NewNullPointerException())
		}
	} else {
		if err := t.ensureTraceRAFLocked(); err != nil {
			return err
		}
	}
	if t.raf != nil {
		ptr, err := t.raf.GetFilePointer()
		if err != nil {
			return err
		}
		uid = ptr
		t.lastPtr = ptr
		if err := t.raf.WriteLongNat(prevUID); err != nil {
			return err
		}
		if err := t.raf.WriteLong(int64(succFP)); err != nil {
			return err
		}
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
	succState.attachTraceMetadata(curState, action)
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
	t.records = append(t.records, TraceRecord{
		PreviousUID: prevUID,
		WorkerID:    predecessorWorkerID,
		FP:          succFP,
		State:       succState,
		Action:      action,
	})
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

func (t *TLCTrace) PrintTrace(curState *TLCStateMut, succState *TLCStateMut) {
	if t == nil || curState == nil {
		panic(NewNullPointerException())
	}
	t.printTraceWithPrefix(curState, succState, t.GetTraceAt(curState.UID, false))
}

// printTraceWithPrefix follows TLCTrace.printTrace(s1, s2, prefix). The printed
// trace is recovered from s1, including when s2 is incomplete and has no UID.
// Worker postcondition reconstruction follows a separate source path.
func (t *TLCTrace) printTraceWithPrefix(curState *TLCStateMut, succState *TLCStateMut, prefix []*TLCStateInfo) {
	if t == nil || curState == nil {
		panic(NewNullPointerException())
	}
	PrintError(ECTLCBehaviorUpToThisPoint)
	if curState.IsInitial() {
		if succState == nil {
			PrintInvariantViolationStateTraceState(NewTLCStateInfo(curState))
		} else {
			info := t.aliasTraceState(NewTLCStateInfo(curState), succState, prefix)
			PrintInvariantViolationStateTraceState(info, curState, 1)
			info = t.stateInfoForTransition(succState, curState)
			info = t.aliasTraceState(info, succState, prefix)
			PrintInvariantViolationStateTraceState(info, curState, 2, true)
		}
		return
	}

	var lastState *TLCStateMut
	idx := 0
	for idx < len(prefix)-1 {
		j := idx + 1
		info := t.aliasTraceState(prefix[idx], prefix[j].State, prefix[:j])
		PrintInvariantViolationStateTraceState(info, lastState, idx+1)
		lastState = prefix[idx].State
		idx++
	}

	var info *TLCStateInfo
	if len(prefix) == 0 {
		info = t.stateInfoForState(curState, nil)
		if info == nil {
			traceRecoveryExit("3", nil)
		}
	} else {
		previous := prefix[len(prefix)-1]
		curState.SetPredecessor(previous.State)
		aliased := t.aliasTraceState(previous, curState, prefix)
		idx++
		PrintInvariantViolationStateTraceState(aliased, lastState, idx)
		info = t.stateInfoForState(curState, previous.State)
		if info == nil {
			traceRecoveryExit("4", curState)
		}
		info.State.UID = curState.UID
		info.State.WorkerID = curState.WorkerID
	}
	successor := succState
	if successor == nil {
		lastState = nil
		successor = info.State
	}
	aliased := t.aliasTraceState(info, successor, prefix, info)
	idx++
	PrintInvariantViolationStateTraceState(aliased, lastState, idx, succState == nil)
	if succState != nil {
		previous := info
		info = t.stateInfoForTransition(succState, curState)
		if info == nil {
			traceRecoveryExit("5", succState)
		}
		info.State.UID = succState.UID
		info.State.WorkerID = succState.WorkerID
		info = t.aliasTraceState(info, succState, prefix, previous, info)
		idx++
		PrintInvariantViolationStateTraceState(info, (*TLCStateMut)(nil), idx, true)
	}
}

// A reconstruction failure terminates the process after source-ordered
// diagnostics. It must bypass the coordinator's ordinary error catch and
// deferred cleanup; returning an error would continue the shutdown protocol.
func traceRecoveryExit(branch string, state *TLCStateMut) {
	PrintError(ECTLCFailedToRecoverInit)
	PrintError(ECTLCBug, branch)
	if state != nil {
		PrintStandaloneErrorState(state)
	}
	os.Exit(1)
}

func (t *TLCTrace) aliasTraceState(info *TLCStateInfo, successor *TLCStateMut, prefix []*TLCStateInfo, suffix ...*TLCStateInfo) *TLCStateInfo {
	if t == nil || t.Tool == nil {
		panic(NewNullPointerException())
	}
	alias, err := t.Tool.EvalAliasInfoPrefixSuffix(info, successor, prefix, suffix...)
	if err != nil {
		panic(err)
	}
	return alias
}

func (t *TLCTrace) stateInfoForState(state *TLCStateMut, predecessor *TLCStateMut) *TLCStateInfo {
	if t == nil || state == nil {
		panic(NewNullPointerException())
	}
	fp := state.FingerPrint()
	if t.Tool == nil {
		panic(NewNullPointerException())
	}
	var info *TLCStateInfo
	var err error
	if predecessor == nil {
		info, err = t.Tool.GetState(fp)
	} else {
		info, err = t.Tool.GetState(fp, predecessor)
	}
	if err != nil {
		panic(err)
	}
	return info
}

func (t *TLCTrace) stateInfoForTransition(state *TLCStateMut, predecessor *TLCStateMut) *TLCStateInfo {
	if t == nil || t.Tool == nil {
		panic(NewNullPointerException())
	}
	info, err := t.Tool.GetStateForTransition(state, predecessor)
	if err != nil {
		panic(err)
	}
	return info
}

func (t *TLCTrace) GetTraceBetween(from *TLCStateMut, to *TLCStateMut) []*TLCStateInfo {
	if to == nil {
		return nil
	}
	if to.IsInitial() || (from != nil && from.Equal(to)) {
		return []*TLCStateInfo{NewTLCStateInfo(to)}
	}
	if from == nil {
		return t.GetTrace(to)
	}
	trace := t.GetTrace(to)
	if len(trace) == 0 {
		return trace
	}
	trace = trace[:len(trace)-1]
	for i, info := range trace {
		if info != nil && info.State != nil && info.State.Equal(from) {
			return trace[i:]
		}
	}
	return trace
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
	if t == nil {
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
		predecessorLoc = next
	}
	if err := t.raf.Seek(current); err != nil {
		return nil, err
	}
	return fps, nil
}

func (t *TLCTrace) recoverTraceFromFPs(sinfo *TLCStateInfo, fps []uint64) ([]*TLCStateInfo, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	snapshot := ResetRandomEnumerableValues()
	out := make([]*TLCStateInfo, 0, len(fps))
	if len(fps) > 0 {
		if sinfo == nil {
			if t.Tool == nil {
				panic(NewNullPointerException())
			}
			info, err := t.Tool.GetState(fps[len(fps)-1])
			if err != nil {
				return nil, err
			}
			sinfo = info
		}
		out = append(out, sinfo)
		for i := len(fps) - 2; i >= 0; i-- {
			fp := fps[i]
			if sinfo == nil {
				panic(NewNullPointerException())
			}
			// Source passes the predecessor state, not the info overload that
			// raises an evaluation error for a missing match.
			if t.Tool == nil {
				panic(NewNullPointerException())
			}
			info, err := t.Tool.GetState(fp, sinfo.State)
			if err != nil {
				return nil, err
			}
			if info == nil {
				traceRecoveryExit(fmt.Sprintf("2 %d", int64(fp)), nil)
			}
			out = append(out, info)
			sinfo = info
		}
	}
	// Source restores the snapshot only after normal completion.
	SetRandomEnumerableGenerator(snapshot)
	return out, nil
}

func (t *TLCTrace) GetLevelForReporting() int {
	level, err := t.GetLevelForReportingWithError()
	if err != nil {
		panic(err)
	}
	return level
}

// GetLevelForReportingWithError preserves the checked I/O boundary used by
// coordinator reporting and management callers with explicit I/O catches.
func (t *TLCTrace) GetLevelForReportingWithError() (int, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, NewIOException("File handle closed")
	}
	if t.rawPaths || t.raf != nil {
		level, err := t.getLevelFromDiskLocked(t.lastPtr)
		if err != nil {
			return 0, err
		}
		if level > t.previousLevel {
			t.previousLevel = level
		}
		return t.previousLevel, nil
	}
	return t.level, nil
}

func (t *TLCTrace) GetLevel(startUID int64) int {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.rawPaths || t.raf != nil {
		level, err := t.getLevelFromDiskLocked(startUID)
		if err != nil {
			panic(err)
		}
		return level
	}
	if startUID < 0 {
		return 0
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
		panic(NewNullPointerException())
	}
	current, err := t.raf.GetFilePointer()
	if err != nil {
		return 0, err
	}
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
		predecessorLoc = next
	}
	if err := t.raf.Seek(current); err != nil {
		return 0, err
	}
	return level, nil
}

func (t *TLCTrace) getPrevFromDiskLocked(loc int64) (int64, error) {
	if t == nil || t.raf == nil {
		panic(NewNullPointerException())
	}
	if err := t.raf.Seek(loc); err != nil {
		return 0, err
	}
	return t.raf.ReadLongNat()
}

func (t *TLCTrace) getFPFromDiskLocked(loc int64) (uint64, error) {
	if t == nil || t.raf == nil {
		panic(NewNullPointerException())
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

func (t *TLCTrace) Elements() (*TLCTraceEnumerator, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil, NewIOException("File handle closed")
	}
	if t.rawPaths {
		if t.raf == nil {
			panic(NewNullPointerException())
		}
	} else {
		if err := t.ensureTraceRAFLocked(); err != nil {
			return nil, err
		}
	}
	if t.raf != nil {
		length, err := t.raf.Length()
		if err != nil {
			return nil, err
		}
		enumRAF, err := NewBufferedRandomAccessFile(t.traceFileName(), "r")
		if err != nil {
			return nil, err
		}
		return &TLCTraceEnumerator{length: length, raf: enumRAF, path: t.traceFileName(), trace: t}, nil
	}
	out := make([]TraceRecord, len(t.records))
	copy(out, t.records)
	return &TLCTraceEnumerator{records: out}, nil
}

func (t *TLCTrace) BeginChkpt() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.rawPaths && t.raf == nil {
		panic(NewNullPointerException())
	}
	if t.diskdir != "" || t.rawPaths {
		if !t.rawPaths {
			if err := t.ensureTraceRAFLocked(); err != nil {
				return err
			}
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
		return NewIOException(fmt.Sprintf("Trace.commitChkpt: cannot delete %s", oldChkpt))
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException(fmt.Sprintf("Trace.commitChkpt: cannot delete %s", oldChkpt))
	}
	return nil
}

func (t *TLCTrace) Recover() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.diskdir == "" && !t.rawPaths {
		return nil
	}
	// A closed source RAF is not reopened by recovery. Metadata reads still
	// precede the seek failure, including publication of the saved last pointer.
	if !t.rawPaths && !t.closed {
		if err := t.ensureTraceRAFLocked(); err != nil {
			return err
		}
	}
	file, err := os.Open(t.chkptName("chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	if t.rawPaths || t.raf != nil || t.closed {
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
		t.lastPtr = lastPtr
		if err := in.Close(); err != nil {
			return err
		}
		if t.rawPaths && t.raf == nil {
			panic(NewNullPointerException())
		}
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
				state.action = action
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
	if t.rawPaths {
		return t.diskdir + string(os.PathSeparator) + t.rootName + ".st." + ext
	}
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
	t.closed = true
	if t.raf == nil {
		return nil
	}
	err := t.raf.Close()
	if !t.rawPaths {
		t.raf = nil
	}
	return err
}

type TLCTraceEnumerator struct {
	records []TraceRecord
	index   int
	length  int64
	raf     *BufferedRandomAccessFile
	path    string
	trace   *TLCTrace
}

func (e *TLCTraceEnumerator) isDisk() bool {
	return e.trace != nil || e.path != "" || e.raf != nil
}

func (e *TLCTraceEnumerator) NextPos() (int64, error) {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.isDisk() {
		if e.raf == nil {
			panic(NewNullPointerException())
		}
		pos, err := e.raf.GetFilePointer()
		if err != nil {
			return 0, err
		}
		if pos >= e.length {
			return -1, nil
		}
		return pos, nil
	}
	if e.index >= len(e.records) {
		return -1, nil
	}
	return int64(e.index), nil
}

func (e *TLCTraceEnumerator) NextFP() (uint64, error) {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.isDisk() {
		if e.raf == nil {
			panic(NewNullPointerException())
		}
		if _, err := e.raf.ReadLongNat(); err != nil {
			return 0, err
		}
		fp, err := e.raf.ReadLong()
		return uint64(fp), err
	}
	if e.index >= len(e.records) {
		return 0, nil
	}
	fp := e.records[e.index].FP
	e.index++
	return fp, nil
}

func (e *TLCTraceEnumerator) Close() error {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.isDisk() {
		if e.raf == nil {
			panic(NewNullPointerException())
		}
		return e.raf.Close()
	}
	return nil
}

func (e *TLCTraceEnumerator) Reset(pos int64) error {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.isDisk() {
		if e.trace == nil {
			panic(NewNullPointerException())
		}
		e.trace.mu.Lock()
		if e.trace.raf == nil {
			e.trace.mu.Unlock()
			panic(NewNullPointerException())
		}
		length, err := e.trace.raf.Length()
		e.trace.mu.Unlock()
		if err != nil {
			return err
		}
		e.length = length
		if pos == -1 {
			if e.raf == nil {
				panic(NewNullPointerException())
			}
			pos, err = e.raf.GetFilePointer()
			if err != nil {
				return err
			}
		}
		raf, err := NewBufferedRandomAccessFile(e.path, "r")
		if err != nil {
			return err
		}
		// Native ownership releases the replaced read-only handle.
		_ = e.raf.Close()
		e.raf = raf
		return e.raf.Seek(pos)
	}
	if pos >= 0 {
		if pos > int64(len(e.records)) {
			pos = int64(len(e.records))
		}
		e.index = int(pos)
	}
	return nil
}

func TLCTraceWriteBehavior(fileName string, state *TLCStateMut, stateTrace *StateVec) error {
	_ = state
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
