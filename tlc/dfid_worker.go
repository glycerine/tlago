package tlc

import (
	"math"
)

type DFIDWorker struct {
	ID             int
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
	MoreLevel      bool
	Result         int
	Err            error
}

func NewDFIDWorker(id int, toLevel int, checker *DFIDModelChecker) *DFIDWorker {
	maxDepth := Globals.DFIDMax
	if maxDepth < toLevel {
		maxDepth = toLevel
	}
	if maxDepth < 1 {
		maxDepth = 1
	}
	seedSource := NewJavaRandomDefault()
	rng := NewJavaRandom(seedSource.NextLong())
	worker := &DFIDWorker{
		ID:             id,
		Checker:        checker,
		Rand:           rng,
		StateStack:     make([]*TLCStateMut, maxDepth),
		FPStack:        make([]uint64, maxDepth),
		SuccStateStack: make([]*StateVec, maxDepth),
		SuccFPStack:    make([]*LongVec, maxDepth),
		ToLevel:        toLevel,
		Result:         NoError,
	}
	if checker != nil {
		worker.FPSet = checker.FPSet
		worker.InitLen = len(checker.InitStates)
		worker.InitStates = make([]*TLCStateMut, worker.InitLen)
		worker.InitFPs = make([]uint64, worker.InitLen)
		copy(worker.InitStates, checker.InitStates)
		copy(worker.InitFPs, checker.InitFPs)
	}
	for i := 0; i < maxDepth; i++ {
		worker.SuccStateStack[i] = NewStateVec(1)
		worker.SuccFPStack[i] = NewLongVecWithCapacity(1)
	}
	return worker
}

func (w *DFIDWorker) SetStop(code int) {
	w.StopCode = code
}

func (w *DFIDWorker) IsTerminated() bool {
	return w != nil && w.StopCode == 2
}

func (w *DFIDWorker) HasMoreLevel() bool {
	return w != nil && w.MoreLevel
}

func (w *DFIDWorker) getInit() int {
	for w.InitLen > 0 {
		index := int(math.Floor(w.Rand.NextDouble() * float64(w.InitLen)))
		fp := w.InitFPs[index]
		status := w.FPSet.GetStatus(fp)
		if !FPIntSetIsCompleted(status) {
			return index
		}
		w.InitLen--
		w.InitStates[index] = w.InitStates[w.InitLen]
		w.InitFPs[index] = w.InitFPs[w.InitLen]
	}
	return -1
}

func (w *DFIDWorker) getNext(curState *TLCStateMut, cfp uint64) int {
	_ = curState
	_ = cfp
	succStates := w.SuccStateStack[w.CurLevel-1]
	succFPs := w.SuccFPStack[w.CurLevel-1]
	length := succFPs.Size()
	for length > 0 {
		index := int(math.Floor(w.Rand.NextDouble() * float64(length)))
		fp := uint64(succFPs.ElementAt(index))
		status := w.FPSet.GetStatus(fp)
		if !FPIntSetIsCompleted(status) && int32(w.CurLevel) < FPIntSetLevelOf(status) {
			return index
		}
		succStates.Remove(index)
		succFPs.RemoveElement(index)
		length--
	}
	return -1
}

func (w *DFIDWorker) Run() {
	var curState *TLCStateMut
	for w.StopCode == 0 {
		index := w.getInit()
		if index == -1 {
			w.StopCode = 1
			if w.Checker != nil {
				w.Checker.SetDone()
			}
			return
		}
		curState = w.InitStates[index]
		cfp := w.InitFPs[index]
		w.StateStack[0] = curState
		w.FPStack[0] = cfp
		w.SuccStateStack[0].Clear()
		w.SuccFPStack[0].Reset()
		isLeaf := w.ToLevel < 2
		noLeaf, result, err := w.Checker.DoNextInto(curState, cfp, isLeaf, w.SuccStateStack[0], w.SuccFPStack[0])
		if w.stopOnResult(result, err) {
			return
		}
		w.MoreLevel = w.MoreLevel || !noLeaf
		w.CurLevel = 1
		for !isLeaf && w.StopCode == 0 {
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
			w.StateStack[w.CurLevel] = curState
			w.FPStack[w.CurLevel] = cfp
			w.SuccStateStack[w.CurLevel].Clear()
			w.SuccFPStack[w.CurLevel].Reset()
			isLeaf = w.CurLevel >= w.ToLevel-1
			noLeaf, result, err = w.Checker.DoNextInto(curState, cfp, isLeaf, w.SuccStateStack[w.CurLevel], w.SuccFPStack[w.CurLevel])
			if w.stopOnResult(result, err) {
				return
			}
			w.MoreLevel = w.MoreLevel || !noLeaf
			w.CurLevel++
		}
	}
}

func (w *DFIDWorker) stopOnResult(result int, err error) bool {
	if err == nil && result == NoError {
		return false
	}
	w.StopCode = 2
	w.Result = result
	w.Err = err
	if w.Checker != nil {
		w.Checker.SetDone()
	}
	return true
}
