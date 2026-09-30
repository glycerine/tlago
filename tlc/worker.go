package tlc

import "fmt"

type Worker struct {
	ID                    int
	LocalValues           []Value
	NamedRegisters        *InsMap[*UniqueString, Value]
	SetOfStates           *SetOfStates
	SetOfStatesMultiplier int
	OutDegree             *FixedSizedBucketStatistics
	StatesGenerated       int64
	Checker               *ModelChecker
	Tool                  *Tool
	Halted                bool
	MaxLevel              int
	UnseenSuccessorStates int
}

func NewWorker(id int) *Worker {
	return &Worker{
		ID:                    id,
		NamedRegisters:        NewInsMap[*UniqueString, Value](),
		SetOfStatesMultiplier: 1,
		OutDegree:             NewFixedSizedBucketStatistics(fmt.Sprintf("TLCWorkerThread-%03d", id), workerOutDegreeBucketCount),
	}
}

func NewModelCheckingWorker(id int, checker *ModelChecker, tool *Tool) *Worker {
	if tool == nil && checker != nil {
		tool = checker.Tool
	}
	worker := NewWorker(id)
	worker.Checker = checker
	worker.Tool = tool
	if checker != nil {
		checker.Workers = append(checker.Workers, worker)
	}
	return worker
}

func (w *Worker) MyGetID() int {
	if w == nil {
		return 0
	}
	return w.ID
}

func (w *Worker) Start() {}

func (w *Worker) Join() error {
	return nil
}

func (w *Worker) NextStateFunctor() *NextStateFunctor {
	return &NextStateFunctor{
		StateFunctor: StateFunctor{
			HasStatesFunc: w.HasStates,
			GetStatesFunc: w.GetStates,
		},
		AddNextElementFunc:         w.AddNextElement,
		IncrementStatesGeneratedFn: w.IncrementStatesGenerated,
		HaltFunc:                   w.Halt,
		AddUnsatisfiedNextStateFn:  w.AddUnsatisfiedNextState,
	}
}

func (w *Worker) DoNext(curState *TLCStateMut) (bool, error) {
	if w == nil {
		return true, newTLCError(ECGeneral, "worker is nil")
	}
	if w.Tool == nil {
		return true, newTLCError(ECGeneral, "worker has no tool")
	}
	if w.Checker == nil {
		return true, newTLCError(ECGeneral, "worker has no model checker")
	}
	if w.Checker.CheckLiveness || w.Tool.GetMode() == ModeDebugger {
		w.SetOfStates = w.CreateSetOfStates()
	}
	preNext := w.StatesGenerated
	halt, err := w.Tool.GetNextStatesWithFunctor(w.NextStateFunctor(), curState)
	if err != nil {
		w.Checker.doNextFailed(curState, nil, err)
		return true, err
	}
	if halt || w.Halted {
		return true, nil
	}
	if w.Checker.CheckDeadlock && preNext == w.StatesGenerated {
		w.RecordOutDegree()
		return w.Checker.doNextSetErr(curState, nil, false, ECTLCDeadlockReached, ""), nil
	}
	if w.SetOfStates != nil && w.SetOfStates.Capacity() > w.SetOfStatesMultiplier*workerSetOfStatesInitialCapacity {
		w.SetOfStatesMultiplier++
	}
	w.RecordOutDegree()
	return false, nil
}

func (w *Worker) AddNextElement(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
	if w == nil {
		return nil, newTLCError(ECGeneral, "worker is nil")
	}
	if w.Checker == nil {
		return nil, newTLCError(ECGeneral, "worker has no model checker")
	}
	if w.Halted {
		return w, nil
	}
	if action != nil && action.CM.node != nil {
		action.CM.IncInvocations()
	}
	w.StatesGenerated++
	stop, queued, err := w.Checker.processSuccessor(curState, succState, action, w.SetOfStates)
	if stop || err != nil {
		w.Halted = true
	}
	if err != nil {
		return nil, err
	}
	if stop {
		return w, nil
	}
	if queued && succState != nil {
		if succState.Level() > w.MaxLevel {
			w.MaxLevel = succState.Level()
		}
		w.UnseenSuccessorStates++
	}
	return w, nil
}

func (w *Worker) AddUnsatisfiedNextState(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
	if w != nil && w.Checker != nil && w.Checker.AllStateWriter != nil && w.Checker.AllStateWriter.IsConstrained() {
		_ = w.Checker.AllStateWriter.WriteTransition(curState, succState, StateVisitNotInModel, action, pred)
	}
	return succState
}

func (w *Worker) IncrementStatesGenerated(count int64) {
	if w != nil {
		w.StatesGenerated += count
	}
}

func (w *Worker) Halt() bool {
	return w != nil && w.Halted
}

func (w *Worker) GetStatesGenerated() int64 {
	if w == nil {
		return 0
	}
	return w.StatesGenerated
}

const workerSetOfStatesInitialCapacity = 16
const workerOutDegreeBucketCount = 32

func (w *Worker) CreateSetOfStates() *SetOfStates {
	if w == nil {
		return NewSetOfStates(workerSetOfStatesInitialCapacity)
	}
	multiplier := w.SetOfStatesMultiplier
	if multiplier < 1 {
		multiplier = 1
		w.SetOfStatesMultiplier = multiplier
	}
	return NewSetOfStates(multiplier * workerSetOfStatesInitialCapacity)
}

func (w *Worker) HasStates() bool {
	return w != nil && w.SetOfStates != nil
}

func (w *Worker) GetStates() *SetOfStates {
	if w != nil && w.SetOfStates != nil {
		return w.SetOfStates
	}
	return NewSetOfStates(0)
}

func (w *Worker) RecordOutDegree() {
	if w == nil {
		return
	}
	if w.OutDegree == nil {
		w.OutDegree = NewFixedSizedBucketStatistics(fmt.Sprintf("TLCWorkerThread-%03d", w.ID), workerOutDegreeBucketCount)
	}
	w.OutDegree.AddSample(w.UnseenSuccessorStates)
	w.UnseenSuccessorStates = 0
}

func (w *Worker) GetLocalValue(index int) Value {
	if w == nil || index < 0 || index >= len(w.LocalValues) {
		return nil
	}
	return w.LocalValues[index]
}

func (w *Worker) SetLocalValue(index int, value Value) {
	if w == nil || index < 0 {
		return
	}
	for len(w.LocalValues) <= index {
		w.LocalValues = append(w.LocalValues, nil)
	}
	w.LocalValues[index] = value
}

func (w *Worker) GetNamedRegister(name *UniqueString) Value {
	if w == nil || w.NamedRegisters == nil {
		return nil
	}
	return w.NamedRegisters.Get(name)
}

func (w *Worker) SetNamedRegister(name *UniqueString, value Value) {
	if w == nil {
		return
	}
	if w.NamedRegisters == nil {
		w.NamedRegisters = NewInsMap[*UniqueString, Value]()
	}
	w.NamedRegisters.Set(name, value)
}
