package tlc

import (
	"math"
	"reflect"
	"sort"
	"sync"
)

type rlSimulationWorkerProperties struct {
	alpha       float64
	gamma       float64
	reward      float64
	enabledOnly bool
}

var rlSimulationWorkerStatics struct {
	sync.Once
	properties rlSimulationWorkerProperties
}

func getRLSimulationWorkerProperties() rlSimulationWorkerProperties {
	rlSimulationWorkerStatics.Do(func() {
		rlSimulationWorkerStatics.properties = rlSimulationWorkerProperties{
			alpha:       simulatorPropertyFloat("tlc2.tool.Simulator.rl.alpha", "TLAGO_SIMULATOR_RL_ALPHA", 0.3),
			gamma:       simulatorPropertyFloat("tlc2.tool.Simulator.rl.gamma", "TLAGO_SIMULATOR_RL_GAMMA", 0.7),
			reward:      simulatorPropertyFloat("tlc2.tool.Simulator.rl.reward", "TLAGO_SIMULATOR_RL_REWARD", -10),
			enabledOnly: simulatorPropertyBool("tlc2.tool.Simulator.rl.enabledOnly", "TLAGO_SIMULATOR_RL_ENABLED_ONLY"),
		}
	})
	return rlSimulationWorkerStatics.properties
}

// InitializeRLSimulationWorkerStatics represents a fresh Java classloader for
// RLSimulationWorker's property-backed static finals. It requires idle workers.
func InitializeRLSimulationWorkerStatics() {
	rlSimulationWorkerStatics.Once = sync.Once{}
	rlSimulationWorkerStatics.properties = rlSimulationWorkerProperties{}
}

func (w *SimulationWorker) SetMode(mode SimulationWorkerMode) {
	w.require()
	w.Mode, w.modeOwner, w.modeCaptured = mode, mode, true
	if w.IsRLMode() {
		if !w.rlPropertiesCaptured {
			properties := getRLSimulationWorkerProperties()
			w.RLAlpha = properties.alpha
			w.RLGamma = properties.gamma
			w.RLReward = properties.reward
			w.RLEnabledOnly = properties.enabledOnly
			w.rlPropertiesOwner = properties
			w.rlPropertiesCaptured = true
		}
		w.EnsureRLQ()
	}
}

func (w *SimulationWorker) IsRLMode() bool {
	w.require()
	return w.ownedMode() == SimulationWorkerRL || w.ownedMode() == SimulationWorkerRLAction
}

func (w *SimulationWorker) ownedMode() SimulationWorkerMode {
	if w.modeCaptured {
		return w.modeOwner
	}
	return w.Mode
}

func (w *SimulationWorker) NextStateFunctor() *NextStateFunctor {
	w.require()
	return &NextStateFunctor{
		StateFunctor: StateFunctor{
			SetElementFunc: func(state *TLCStateMut) (any, error) {
				return w.SetElement(state), nil
			},
			HasStatesFunc: func() bool {
				return w.HasStates()
			},
			GetStatesFunc: func() *SetOfStates {
				return w.GetStates()
			},
		},
		AddNextElementFunc: func(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
			return w.AddElement(curState, action, succState)
		},
		IncrementStatesGeneratedFn: func(count int64) {
			if w.ownedStatistics() != nil && w.ownedStatistics().ownedNumGenStates() != nil {
				w.ownedStatistics().ownedNumGenStates().Add(count)
			}
		},
		HaltFunc: func() bool {
			return w.Halt()
		},
		ShouldHaltFunc: func() bool {
			return w.Halted.Load()
		},
	}
}

func (w *SimulationWorker) SimulateExplorationTrace() *SimulationWorkerError {
	w.require()
	if w.ownedTool() == nil {
		panic(NewNullPointerException())
	}
	initFunctor := &StateFunctor{
		AddElementFunc: func(state *TLCStateMut) (any, error) {
			inModel, err := w.ownedTool().IsInModel(state)
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
	if err := w.ownedTool().GetInitStates(initFunctor); err != nil {
		panic(err)
	}
	w.CurState = w.RandomState(w.InitStates)
	SetCurrentState(w.CurState)
	for traceIdx := 0; traceIdx < w.ownedMaxTraceDepth(); traceIdx++ {
		w.CheckForInterrupt()
		w.ownedNextStates().Clear()
		var err error
		func() {
			restoreCurrentState := PushCurrentState(w.CurState)
			defer restoreCurrentState()
			_, err = w.ownedTool().GetNextStatesWithFunctor(w.NextStateFunctor(), w.CurState)
		}()
		if err != nil {
			if workerErr, ok := err.(*SimulationWorkerError); ok {
				return workerErr
			}
			panic(err)
		}
		if w.Halted.Load() {
			w.Halted.Store(false)
			return nil
		}
		if w.ownedNextStates().IsEmpty() {
			if w.ownedCheckDeadlock() {
				return w.newError(ECTLCDeadlockReached, nil, w.GetTrace(w.CurState))
			}
			break
		}
		next := w.RandomState(w.ownedNextStates())
		if _, err := next.ExecCallable(); err != nil {
			panic(err)
		}
		w.ownedStatistics().CollectPostSuccessor(w.CurState, next.GetAction(), next)
		w.CurState = next
		SetCurrentState(w.CurState)
	}
	w.CheckForInterrupt()
	if workerErr := w.CheckLivenessTrace(); workerErr != nil {
		return workerErr
	}
	w.ownedStatistics().CollectPostTrace(w.CurState)
	w.PostTrace(w.CurState)
	return nil
}

func (w *SimulationWorker) EnsureRLQ() {
	w.require()
	if !w.rlQCaptured {
		if w.RLQ == nil {
			w.RLQ = NewInsMap[*Action, *InsMap[int64, float64]]()
		}
		w.rlQOwner = w.RLQ
		w.rlQCaptured = true
	}
	if w.ownedTool() == nil {
		return
	}
	for _, action := range w.ownedTool().GetActions() {
		w.rlActionQ(action)
	}
}

func (w *SimulationWorker) ownedRLQ() *InsMap[*Action, *InsMap[int64, float64]] {
	if w.rlQCaptured {
		return w.rlQOwner
	}
	return w.RLQ
}

func (w *SimulationWorker) ownedRLProperties() rlSimulationWorkerProperties {
	if w.rlPropertiesCaptured {
		return w.rlPropertiesOwner
	}
	return rlSimulationWorkerProperties{
		alpha: w.RLAlpha, gamma: w.RLGamma, reward: w.RLReward, enabledOnly: w.RLEnabledOnly,
	}
}

func (w *SimulationWorker) rlActionQ(action *Action) *InsMap[int64, float64] {
	w.require()
	if action == nil {
		action = UnknownAction
	}
	qMap := w.ownedRLQ()
	if qMap == nil {
		w.EnsureRLQ()
		qMap = w.ownedRLQ()
	}
	q := qMap.Get(action)
	if q == nil {
		q = NewInsMap[int64, float64]()
		qMap.Set(action, q)
	}
	return q
}

func (w *SimulationWorker) RLHash(state *TLCStateMut) int64 {
	w.require()
	if state == nil {
		panic(NewNullPointerException())
	}
	if w.ownedMode() == SimulationWorkerRLAction {
		return actionIdentityHash(state.GetAction())
	}
	return int64(state.FingerPrint())
}

func actionIdentityHash(action *Action) int64 {
	if action == nil {
		panic(NewNullPointerException())
	}
	return int64(reflect.ValueOf(action).Pointer())
}

func (w *SimulationWorker) GetRLReward(s *TLCStateMut, action *Action, t *TLCStateMut) (float64, error) {
	w.require()
	if w.ownedTool() == nil {
		return 0, newTLCError(ECGeneral, "simulation worker has no tool")
	}
	return w.ownedTool().EvalReward(s, t, w.ownedRLProperties().reward)
}

func (w *SimulationWorker) GetMaxRLQ(fp int64) float64 {
	w.require()
	maxQ := -math.MaxFloat64
	qMap := w.ownedRLQ()
	if qMap == nil {
		return maxQ
	}
	for _, q := range qMap.All() {
		if value, ok := q.Get2(fp); ok {
			maxQ = max(maxQ, value)
		}
	}
	return maxQ
}

type rlActionProbability struct {
	probability float64
	index       int
}

func (w *SimulationWorker) GetRLNextActionIndex(actions []*Action, state *TLCStateMut) int {
	w.require()
	qMap := w.ownedRLQ()
	fp := w.RLHash(state)
	for _, q := range qMap.All() {
		if _, ok := q.Get2(fp); !ok {
			q.Set(fp, 0)
		}
	}
	if actions == nil {
		panic(NewNullPointerException())
	}
	weights := make([]rlActionProbability, 0, len(actions))
	denom := 0.0
	for i, action := range actions {
		q := qMap.Get(action)
		if q == nil {
			panic(NewNullPointerException())
		}
		value, ok := q.Get2(fp)
		if !ok {
			panic(NewNullPointerException())
		}
		exp := math.Exp(value)
		denom += exp
		weights = append(weights, rlActionProbability{probability: exp, index: i})
	}
	for i := range weights {
		weights[i].probability = weights[i].probability / denom
	}
	draw := w.ownedRand().NextDouble()
	sort.SliceStable(weights, func(i, j int) bool {
		a, b := weights[i].probability, weights[j].probability
		// Pair.compareTo uses descending Double.compare, with NaNs first.
		if math.IsNaN(a) {
			return !math.IsNaN(b)
		}
		if math.IsNaN(b) {
			return false
		}
		if a == 0 && b == 0 {
			return !math.Signbit(a) && math.Signbit(b)
		}
		return a > b
	})
	cumulative := 0.0
	for _, weight := range weights {
		cumulative += weight.probability
		if cumulative >= draw {
			return weight.index
		}
	}
	if len(weights) == 0 {
		panic(NewIndexOutOfBoundsException(-1, 0))
	}
	return weights[len(weights)-1].index
}

func (w *SimulationWorker) UpdateRLDisabledAction(index int, actions []*Action, curState *TLCStateMut) {
	w.require()
	index = int(int32(index))
	actionAt := func() *Action {
		if actions == nil {
			panic(NewNullPointerException())
		}
		if index < 0 || index >= len(actions) {
			panic(NewArrayIndexOutOfBoundsException(index, len(actions)))
		}
		return actions[index]
	}
	if w.ownedMode() == SimulationWorkerRLAction {
		if curState == nil {
			panic(NewNullPointerException())
		}
		prev := curState.Predecessor()
		action := actionAt()
		q := w.ownedRLQ().Get(action)
		fp := w.RLHash(curState)
		reward, err := w.GetRLReward(prev, action, curState)
		if err != nil {
			panic(err)
		}
		if q == nil {
			panic(NewNullPointerException())
		}
		q.Set(fp, w.ownedRLProperties().alpha*reward)
	}
	if !w.ownedRLProperties().enabledOnly {
		q := w.ownedRLQ().Get(actionAt())
		fp := w.RLHash(curState)
		if q == nil {
			panic(NewNullPointerException())
		}
		q.Set(fp, -math.MaxFloat64)
	}
}

func (w *SimulationWorker) FilterEnabledRLActions(actions []*Action, curState *TLCStateMut) ([]*Action, *SimulationWorkerError) {
	w.require()
	if actions == nil {
		panic(NewNullPointerException())
	}
	enabled := make([]*Action, 0, len(actions))
	for _, action := range actions {
		var nextStates *StateVec
		var err error
		func() {
			restoreCurrentState := PushCurrentState(curState)
			defer restoreCurrentState()
			nextStates, err = w.ownedTool().GetNextStates(action, curState)
		}()
		if err != nil {
			if workerErr, ok := err.(*SimulationWorkerError); ok {
				return nil, workerErr
			}
			panic(err)
		}
		if nextStates == nil {
			panic(NewNullPointerException())
		}
		if !nextStates.IsEmpty() {
			enabled = append(enabled, action)
		}
	}
	return enabled, nil
}

func (w *SimulationWorker) PostRLTrace(finalState *TLCStateMut) bool {
	w.require()
	if finalState == nil {
		panic(NewNullPointerException())
	}
	s := finalState
	for i := s.Level() - 1; i > 0; i-- {
		maxQ := w.GetMaxRLQ(w.RLHash(s))
		prev := s.Predecessor()
		fp := w.RLHash(prev)
		action := s.GetAction()
		actionQ := w.ownedRLQ().Get(action)
		if actionQ == nil {
			panic(NewNullPointerException())
		}
		qi, ok := actionQ.Get2(fp)
		if !ok {
			panic(NewNullPointerException())
		}
		reward, err := w.GetRLReward(prev, action, s)
		if err != nil {
			panic(err)
		}
		properties := w.ownedRLProperties()
		q := ((1 - properties.alpha) * qi) + (properties.alpha * (reward + (properties.gamma * maxQ)))
		actionQ.Set(fp, q)
		s = prev
	}
	return true
}
