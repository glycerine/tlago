package tlc

import (
	"fmt"
	"math"
	"sync"
)

type DFIDWorker struct {
	ID             int
	LocalValues    []Value
	NamedRegisters *NamedRegisterMap
	Checker        *DFIDModelChecker
	Rand           *JavaRandom
	StateStack     []*TLCStateMut
	FPStack        []uint64
	SuccStateStack []*StateVec
	SuccFPStack    []*LongVec
	FPSet          *MemFPIntSet
	InitStates     []*TLCStateMut
	InitFPs        []uint64
	InitLen        int
	ToLevel        int
	CurLevel       int
	StopCode       int
	stopMu         sync.Mutex
	registerMu     sync.Mutex
	MoreLevel      bool
	Result         int
	Err            error

	identityCaptured    bool
	idOwner             int
	namedRegistersOwner *NamedRegisterMap
}

func (w *DFIDWorker) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *DFIDWorker) requireChecker() *DFIDModelChecker {
	w.require()
	if w.Checker == nil {
		panic(NewNullPointerException())
	}
	return w.Checker
}

func NewDFIDWorker(id int, toLevel int, checker *DFIDModelChecker) *DFIDWorker {
	id = int(int32(id))
	toLevel = int(int32(toLevel))
	maxDepth := Globals.DFIDMax
	seedSource := NewJavaRandomDefault()
	rng := NewJavaRandom(seedSource.NextLong())
	if maxDepth < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(maxDepth)))
	}
	worker := &DFIDWorker{
		ID:             id,
		LocalValues:    make([]Value, 4),
		NamedRegisters: NewNamedRegisterMap(),
		Checker:        checker,
		Rand:           rng,
		StateStack:     make([]*TLCStateMut, maxDepth),
		FPStack:        make([]uint64, maxDepth),
		SuccStateStack: make([]*StateVec, maxDepth),
		SuccFPStack:    make([]*LongVec, maxDepth),
		Result:         NoError,
	}
	for i := 0; i < maxDepth; i++ {
		worker.SuccStateStack[i] = NewStateVec(1)
		worker.SuccFPStack[i] = NewLongVecWithCapacity(1)
	}
	if checker == nil {
		panic(NewNullPointerException())
	}
	worker.FPSet = checker.FPSet
	if checker.InitStates == nil {
		panic(NewNullPointerException())
	}
	worker.InitLen = len(checker.InitStates)
	worker.InitStates = make([]*TLCStateMut, worker.InitLen)
	worker.InitFPs = make([]uint64, worker.InitLen)
	copy(worker.InitStates, checker.InitStates)
	if checker.InitFPs == nil {
		panic(NewNullPointerException())
	}
	if len(checker.InitFPs) < worker.InitLen {
		message := fmt.Sprintf("arraycopy: last source index %d out of bounds for long[%d]", worker.InitLen, len(checker.InitFPs))
		panic(&ArrayIndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)})
	}
	copy(worker.InitFPs, checker.InitFPs)
	worker.ToLevel = toLevel
	worker.idOwner = worker.ID
	worker.namedRegistersOwner = worker.NamedRegisters
	worker.identityCaptured = true
	return worker
}

func (w *DFIDWorker) ownedID() int {
	if w.identityCaptured {
		return w.idOwner
	}
	return w.ID
}

func (w *DFIDWorker) ownedNamedRegisters() *NamedRegisterMap {
	if w.identityCaptured {
		return w.namedRegistersOwner
	}
	return w.NamedRegisters
}

func (w *DFIDWorker) SetStop(code int) {
	w.require()
	w.stopMu.Lock()
	defer w.stopMu.Unlock()
	w.StopCode = int(int32(code))
}

func (w *DFIDWorker) stopCode() int {
	w.require()
	w.stopMu.Lock()
	defer w.stopMu.Unlock()
	return w.StopCode
}

func (w *DFIDWorker) IsTerminated() bool {
	w.require()
	return w.stopCode() == 2
}

func (w *DFIDWorker) HasMoreLevel() bool {
	w.require()
	return w.MoreLevel
}

func (w *DFIDWorker) getInit() int {
	w.require()
	for w.InitLen > 0 {
		index := int(math.Floor(w.Rand.NextDouble() * float64(w.InitLen)))
		fp := w.InitFPs[index]
		status := w.FPSet.GetStatus(fp)
		if !IsCompleted(status) {
			return index
		}
		w.InitLen--
		w.InitStates[index] = w.InitStates[w.InitLen]
		w.InitFPs[index] = w.InitFPs[w.InitLen]
	}
	return -1
}

func (w *DFIDWorker) getNext(curState *TLCStateMut, cfp uint64) int {
	w.require()
	_ = curState
	_ = cfp
	succStates := w.SuccStateStack[w.CurLevel-1]
	succFPs := w.SuccFPStack[w.CurLevel-1]
	length := succFPs.Size()
	for length > 0 {
		index := int(math.Floor(w.Rand.NextDouble() * float64(length)))
		fp := uint64(succFPs.ElementAt(index))
		status := w.FPSet.GetStatus(fp)
		if !IsCompleted(status) && int32(w.CurLevel) < GetLevel(status) {
			return index
		}
		succStates.Remove(index)
		succFPs.RemoveElement(index)
		length--
	}
	return -1
}

func (w *DFIDWorker) traceStackState(index int) *TLCStateMut {
	w.require()
	if index < 0 || index >= len(w.StateStack) {
		panic(NewArrayIndexOutOfBoundsException(index, len(w.StateStack)))
	}
	return w.StateStack[index]
}

func (w *DFIDWorker) setStackState(index int, state *TLCStateMut) {
	w.require()
	if index < 0 || index >= len(w.StateStack) {
		panic(NewArrayIndexOutOfBoundsException(index, len(w.StateStack)))
	}
	w.StateStack[index] = state
}

func (w *DFIDWorker) assertTracePredecessor(state *TLCStateMut, index int) {
	w.require()
	// Java evaluates the array argument before checking the equals receiver.
	stackState := w.traceStackState(index)
	if state == nil {
		panic(NewNullPointerException())
	}
	if !state.Equal(stackState) {
		panic(NewAssertionError())
	}
}

func (w *DFIDWorker) PrintErrorTrace(errorCode int, params []string, s1 *TLCStateMut, s2 *TLCStateMut) {
	w.require()
	PrintError(errorCode, params...)
	PrintError(ECTLCBehaviorUpToThisPoint)
	idx := 0
	for idx < w.CurLevel {
		state := w.traceStackState(idx)
		idx++
		PrintRuntimeErrorStateTraceState(state, idx)
	}
	w.assertTracePredecessor(s1, idx)
	PrintRuntimeErrorStateTraceState(s1, idx+1)
	if s2 != nil {
		PrintRuntimeErrorStateTraceState(s2, idx+2)
	}
}

func (w *DFIDWorker) PrintInvariantTrace(errorCode int, params []string, s1 *TLCStateMut, s2 *TLCStateMut) {
	w.require()
	PrintError(errorCode, params...)
	PrintError(ECTLCBehaviorUpToThisPoint)
	idx := 0
	for ; idx <= w.CurLevel; idx++ {
		if idx == w.CurLevel {
			w.assertTracePredecessor(s1, idx)
		}
		ordinal := idx + 1
		info := NewTLCStateInfoWithOrdinal(w.traceStackState(idx), ordinal)
		var previous *TLCStateMut
		if idx > 0 {
			previous = w.traceStackState(idx - 1)
		}
		PrintInvariantViolationStateTraceState(info, previous, ordinal)
	}
	ordinal := idx + 1
	info := NewTLCStateInfoWithOrdinal(s2, ordinal)
	PrintInvariantViolationStateTraceState(info, s1, ordinal, true)
}

func (w *DFIDWorker) Run() {
	w.requireChecker()
	restoreWorkerID := pushUnownedWorkerID(w.ownedID())
	defer restoreWorkerID()
	var curState *TLCStateMut
	defer func() {
		if recovered := recover(); recovered != nil {
			w.Result = ECGeneral
			w.Err = panicValueAsError(recovered)
			checker := w.requireChecker()
			checker.SetStop(2)
			checker.mu.Lock()
			defer checker.mu.Unlock()
			if checker.SetErrState(curState, nil, true, ECGeneral) {
				PrintError(ECGeneral, generalErrorParams("", w.Err)...)
			}
			checker.SetDone()
		}
	}()
	for w.stopCode() == 0 {
		index := w.getInit()
		if index == -1 {
			checker := w.requireChecker()
			checker.SetStop(1)
			checker.SetDone()
			return
		}
		curState = w.InitStates[index]
		cfp := w.InitFPs[index]
		w.setStackState(0, curState)
		w.FPStack[0] = cfp
		w.SuccStateStack[0].Clear()
		w.SuccFPStack[0].Reset()
		isLeaf := w.ToLevel < 2
		noLeaf, result, err := w.Checker.DoNext(curState, cfp, isLeaf, w.SuccStateStack[0], w.SuccFPStack[0])
		if w.stopOnResult(result, err) {
			return
		}
		w.MoreLevel = w.MoreLevel || !noLeaf
		w.CurLevel = 1
		for !isLeaf {
			index = w.getNext(curState, cfp)
			if index == -1 {
				w.FPSet.SetLeveled(cfp)
				if w.CurLevel == 1 {
					break
				}
				w.CurLevel--
				curState = w.StateStack[w.CurLevel-1]
				cfp = w.FPStack[w.CurLevel-1]
				continue
			}
			curState = w.SuccStateStack[w.CurLevel-1].At(index)
			cfp = uint64(w.SuccFPStack[w.CurLevel-1].ElementAt(index))
			w.setStackState(w.CurLevel, curState)
			w.FPStack[w.CurLevel] = cfp
			w.SuccStateStack[w.CurLevel].Clear()
			w.SuccFPStack[w.CurLevel].Reset()
			isLeaf = w.CurLevel >= w.ToLevel-1
			noLeaf, result, err = w.Checker.DoNext(curState, cfp, isLeaf, w.SuccStateStack[w.CurLevel], w.SuccFPStack[w.CurLevel])
			if w.stopOnResult(result, err) {
				return
			}
			w.MoreLevel = w.MoreLevel || !noLeaf
			w.CurLevel++
		}
	}
}

func (w *DFIDWorker) stopOnResult(result int, err error) bool {
	w.require()
	// Java doNext records semantic failures and returns allSuccNonLeaf. The
	// worker still advances/backtracks until its inner depth loop finishes;
	// stopCode is checked only by the outer initial-state loop. Only a thrown
	// error leaves run immediately.
	if result != NoError && w.Result == NoError {
		w.Result = result
	}
	if err == nil {
		return false
	}
	w.SetStop(2)
	w.Result = result
	w.Err = err
	w.requireChecker().SetDone()
	return true
}
