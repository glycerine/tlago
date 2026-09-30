package tlc

import (
	"math"
	"reflect"
	"sort"
)

func (w *SimulationWorker) SetMode(mode SimulationWorkerMode) {
	if w == nil {
		return
	}
	w.Mode = mode
	if w.IsRLMode() {
		w.EnsureRLQ()
	}
}

func (w *SimulationWorker) IsRLMode() bool {
	return w != nil && (w.Mode == SimulationWorkerRL || w.Mode == SimulationWorkerRLAction)
}

func (w *SimulationWorker) NextStateFunctor() *NextStateFunctor {
	return &NextStateFunctor{
		StateFunctor: StateFunctor{
			AddElementFunc: func(state *TLCStateMut) (any, error) {
				w.NextStates.Clear()
				w.NextStates.Add(w.markWorkerState(state))
				return w, nil
			},
			SetElementFunc: func(state *TLCStateMut) (any, error) {
				w.NextStates.Clear()
				w.NextStates.Add(w.markWorkerState(state))
				return w, nil
			},
			HasStatesFunc: func() bool {
				return w.NextStates != nil && !w.NextStates.IsEmpty()
			},
			GetStatesFunc: func() *SetOfStates {
				return NewSetOfStatesFromVec(w.NextStates)
			},
		},
		AddNextElementFunc: func(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
			if workerErr := w.AddGeneratedSuccessor(curState, action, succState); workerErr != nil {
				return nil, workerErr
			}
			return w, nil
		},
		IncrementStatesGeneratedFn: func(count int64) {
			if w.Statistics != nil && w.Statistics.NumGenStates != nil {
				w.Statistics.NumGenStates.Add(count)
			}
		},
		HaltFunc: func() bool {
			return w.Halted.Load()
		},
	}
}

func (w *SimulationWorker) SimulateExplorationTrace() *SimulationWorkerError {
	if w.InitStates == nil {
		w.InitStates = NewStateVec(0)
	}
	initFunctor := &StateFunctor{
		AddElementFunc: func(state *TLCStateMut) (any, error) {
			inModel, err := w.Tool.IsInModel(state)
			if err != nil {
				return nil, err
			}
			if inModel {
				w.InitStates.Add(state)
			}
			return state, nil
		},
		SetElementFunc: func(state *TLCStateMut) (any, error) {
			w.InitStates.Clear()
			w.InitStates.Add(state)
			return state, nil
		},
		GetStatesFunc: func() *SetOfStates {
			return NewSetOfStatesFromVec(w.InitStates)
		},
	}
	if err := w.Tool.GetInitStates(initFunctor); err != nil {
		return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
	}
	w.CurState = w.claimInitialState(w.RandomState(w.InitStates))
	for traceIdx := 0; traceIdx < w.MaxTraceDepth; traceIdx++ {
		if w.Stopped.Load() {
			return nil
		}
		w.NextStates.Clear()
		if _, err := w.Tool.GetNextStatesWithFunctor(w.NextStateFunctor(), w.CurState); err != nil {
			if workerErr, ok := err.(*SimulationWorkerError); ok {
				return workerErr
			}
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
		}
		if w.Halted.Load() {
			w.Halted.Store(false)
			return nil
		}
		if w.NextStates.IsEmpty() {
			if w.CheckDeadlock {
				return &SimulationWorkerError{Code: ECTLCDeadlockReached, StateTrace: w.GetTrace(w.CurState)}
			}
			break
		}
		next := w.RandomState(w.NextStates)
		if _, err := next.ExecCallable(); err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(next), Err: err}
		}
		w.Statistics.CollectPostSuccessor(w.CurState, next.GetAction(), next)
		w.CurState = next
	}
	if w.Stopped.Load() {
		return nil
	}
	if w.LiveCheck != nil {
		if err := w.LiveCheck.CheckTrace(w.Tool.NoDebug(), func() *StateVec { return w.GetTrace(w.CurState) }); err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
		}
	}
	w.Statistics.CollectPostTrace(w.CurState, w.MaxTraceDepth)
	return w.PostTrace(w.CurState)
}

func (w *SimulationWorker) EnsureRLQ() {
	if w == nil {
		return
	}
	if w.RLQ == nil {
		w.RLQ = NewInsMap[*Action, *InsMap[int64, float64]]()
	}
	if w.Tool == nil {
		return
	}
	for _, action := range w.Tool.GetActions() {
		w.rlActionQ(action)
	}
}

func (w *SimulationWorker) rlActionQ(action *Action) *InsMap[int64, float64] {
	if action == nil {
		action = UnknownAction
	}
	if w.RLQ == nil {
		w.RLQ = NewInsMap[*Action, *InsMap[int64, float64]]()
	}
	q := w.RLQ.Get(action)
	if q == nil {
		q = NewInsMap[int64, float64]()
		w.RLQ.Set(action, q)
	}
	return q
}

func (w *SimulationWorker) RLHash(state *TLCStateMut) int64 {
	if state == nil {
		return 0
	}
	if w.Mode == SimulationWorkerRLAction {
		return actionIdentityHash(state.GetAction())
	}
	return int64(state.FingerPrint())
}

func actionIdentityHash(action *Action) int64 {
	if action == nil {
		return 0
	}
	return int64(reflect.ValueOf(action).Pointer())
}

func (w *SimulationWorker) GetRLReward(s *TLCStateMut, action *Action, t *TLCStateMut) (float64, error) {
	if w == nil || w.Tool == nil {
		return 0, newTLCError(ECGeneral, "simulation worker has no tool")
	}
	return w.Tool.EvalReward(s, t, w.RLReward)
}

func (w *SimulationWorker) GetMaxRLQ(fp int64) float64 {
	maxQ := -math.MaxFloat64
	if w == nil || w.RLQ == nil {
		return maxQ
	}
	for _, q := range w.RLQ.All() {
		if value, ok := q.Get2(fp); ok && value > maxQ {
			maxQ = value
		}
	}
	return maxQ
}

type rlActionProbability struct {
	probability float64
	index       int
}

func (w *SimulationWorker) GetRLNextActionIndex(actions []*Action, state *TLCStateMut) int {
	w.EnsureRLQ()
	fp := w.RLHash(state)
	for _, q := range w.RLQ.All() {
		if _, ok := q.Get2(fp); !ok {
			q.Set(fp, 0)
		}
	}
	weights := make([]rlActionProbability, 0, len(actions))
	denom := 0.0
	for i, action := range actions {
		value := w.rlActionQ(action).Get(fp)
		exp := math.Exp(value)
		denom += exp
		weights = append(weights, rlActionProbability{probability: exp, index: i})
	}
	for i := range weights {
		weights[i].probability = weights[i].probability / denom
	}
	sort.SliceStable(weights, func(i, j int) bool {
		return weights[i].probability > weights[j].probability
	})
	draw := w.Rand.NextDouble()
	cumulative := 0.0
	for _, weight := range weights {
		cumulative += weight.probability
		if cumulative >= draw {
			return weight.index
		}
	}
	return weights[len(weights)-1].index
}

func (w *SimulationWorker) UpdateRLDisabledAction(index int, actions []*Action, curState *TLCStateMut) {
	if index < 0 || index >= len(actions) {
		return
	}
	action := actions[index]
	if w.Mode == SimulationWorkerRLAction {
		prev := curState.Predecessor()
		reward, err := w.GetRLReward(prev, action, curState)
		if err == nil {
			w.rlActionQ(action).Set(w.RLHash(curState), w.RLAlpha*reward)
		}
	}
	if !w.RLEnabledOnly {
		w.rlActionQ(action).Set(w.RLHash(curState), -math.MaxFloat64)
	}
}

func (w *SimulationWorker) FilterEnabledRLActions(actions []*Action, curState *TLCStateMut) ([]*Action, *SimulationWorkerError) {
	enabled := make([]*Action, 0, len(actions))
	for _, action := range actions {
		nextStates, err := w.Tool.GetNextStates(action, curState)
		if err != nil {
			return nil, &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(curState), Err: err}
		}
		if nextStates != nil && !nextStates.IsEmpty() {
			enabled = append(enabled, action)
		}
	}
	return enabled, nil
}

func (w *SimulationWorker) PostRLTrace(finalState *TLCStateMut) *SimulationWorkerError {
	s := finalState
	if s == nil {
		return nil
	}
	for i := s.Level() - 1; i > 0; i-- {
		maxQ := w.GetMaxRLQ(w.RLHash(s))
		prev := s.Predecessor()
		if prev == nil {
			break
		}
		action := s.GetAction()
		fp := w.RLHash(prev)
		qi := w.rlActionQ(action).Get(fp)
		reward, err := w.GetRLReward(prev, action, s)
		if err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(s), Err: err}
		}
		q := ((1 - w.RLAlpha) * qi) + (w.RLAlpha * (reward + (w.RLGamma * maxQ)))
		w.rlActionQ(action).Set(fp, q)
		s = prev
	}
	return nil
}
