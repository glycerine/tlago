package tlc

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync/atomic"
)

type SimulationWorkerError struct {
	Code       int
	Params     []string
	StateTrace *StateVec
	Err        error
}

func (e *SimulationWorkerError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("simulation worker error %d", e.Code)
}

func (e *SimulationWorkerError) HasTrace() bool {
	return e != nil && e.StateTrace != nil && e.StateTrace.Size() > 0
}

type SimulationWorkerResult struct {
	WorkerID int
	Error    *SimulationWorkerError
}

func SimulationWorkerOK(workerID int) SimulationWorkerResult {
	return SimulationWorkerResult{WorkerID: workerID}
}

func SimulationWorkerFailed(workerID int, err *SimulationWorkerError) SimulationWorkerResult {
	return SimulationWorkerResult{WorkerID: workerID, Error: err}
}

func (r SimulationWorkerResult) IsError() bool {
	return r.Error != nil
}

type SimulationWorkerStatistics struct {
	TraceActions    string
	NumGenStates    *atomic.Int64
	NumGenTraces    *atomic.Int64
	WelfordM2Mean   *atomic.Int64
	ActionStats     [][]int64
	NextRetries     int64
	DistinctStates  int64
	DistinctValues  *InsMap[*UniqueString, int64]
	Extended        bool
	workerActionIDs *InsMap[*UniqueString, int]
}

func NewSimulationWorkerStatistics(tool *Tool, traceActions string, states *atomic.Int64, traces *atomic.Int64, m2Mean *atomic.Int64) *SimulationWorkerStatistics {
	if states == nil {
		states = &atomic.Int64{}
	}
	if traces == nil {
		traces = &atomic.Int64{}
	}
	if m2Mean == nil {
		m2Mean = &atomic.Int64{}
	}
	size := 1
	if traceActions != "" && tool != nil {
		size = len(tool.GetActions())
		if size < 1 {
			size = 1
		}
	}
	stats := &SimulationWorkerStatistics{
		TraceActions:    traceActions,
		NumGenStates:    states,
		NumGenTraces:    traces,
		WelfordM2Mean:   m2Mean,
		ActionStats:     make([][]int64, size),
		DistinctValues:  NewInsMap[*UniqueString, int64](),
		workerActionIDs: NewInsMap[*UniqueString, int](),
	}
	for i := range stats.ActionStats {
		stats.ActionStats[i] = make([]int64, size)
	}
	if tool != nil {
		for id, action := range tool.GetActions() {
			if action != nil {
				stats.workerActionIDs.Set(UniqueStringOf(action.GetName()), id)
			}
		}
	}
	return stats
}

func (s *SimulationWorkerStatistics) CollectPreSuccessor(state *TLCStateMut, action *Action, next *TLCStateMut) {
	if s == nil {
		return
	}
	s.NumGenStates.Add(1)
	if s.Extended && next != nil {
		s.DistinctStates++
		for _, variable := range StateVariables() {
			if variable.Name == nil || next.Lookup(variable.Name) == nil {
				continue
			}
			cur := s.DistinctValues.Get(variable.Name)
			s.DistinctValues.Set(variable.Name, cur+1)
		}
	}
}

func (s *SimulationWorkerStatistics) CollectPostSuccessor(state *TLCStateMut, action *Action, next *TLCStateMut) {
	if s == nil || s.TraceActions == "" || state == nil || next == nil {
		return
	}
	from := actionIDFromStateAction(s, state.GetAction())
	to := actionIDFromStateAction(s, next.GetAction())
	if from >= 0 && from < len(s.ActionStats) && to >= 0 && to < len(s.ActionStats[from]) {
		s.ActionStats[from][to]++
	}
	_ = action
}

func (s *SimulationWorkerStatistics) CollectPreTrace() int64 {
	if s == nil {
		return 0
	}
	return s.NumGenTraces.Add(1)
}

func (s *SimulationWorkerStatistics) CollectNextRetries() {
	if s != nil {
		s.NextRetries++
	}
}

func (s *SimulationWorkerStatistics) CollectPostTrace(state *TLCStateMut, maxTraceDepth int) {
	if s == nil || state == nil {
		return
	}
	traceLen := int64(state.Level())
	if maxTraceDepth >= 0 && traceLen > int64(maxTraceDepth) {
		traceLen = int64(maxTraceDepth)
	}
	for {
		old := s.WelfordM2Mean.Load()
		mean := int64(int32(old & 0xffffffff))
		m2 := old >> 32
		count := s.NumGenTraces.Load()
		if count < 1 {
			count = 1
		}
		delta := traceLen - mean
		nextMean := mean + delta/count
		nextM2 := m2 + delta*(traceLen-nextMean)
		next := (nextM2 << 32) | (nextMean & 0xffffffff)
		if s.WelfordM2Mean.CompareAndSwap(old, next) {
			return
		}
	}
}

func (s *SimulationWorkerStatistics) GetNextRetries() Value {
	if s == nil || !s.Extended {
		return NewIntValue(-1)
	}
	return intValueFromInt64(s.NextRetries)
}

type SimulationWorker struct {
	ID            int
	Tool          *Tool
	Rand          *rand.Rand
	CurState      *TLCStateMut
	InitStates    *StateVec
	ResultQueue   chan SimulationWorkerResult
	TraceCnt      int64
	GlobalTrace   int64
	MaxTraceNum   int64
	MaxTraceDepth int
	CheckDeadlock bool
	Debug         bool
	TraceFile     string
	LiveCheck     *LiveCheck
	Statistics    *SimulationWorkerStatistics
	Stopped       atomic.Bool
	NextStates    *StateVec
}

func NewSimulationWorker(id int, tool *Tool, results chan SimulationWorkerResult, seed int64, maxTraceDepth int, maxTraceNum int64, traceActions string, checkDeadlock bool, debug bool, traceFile string, liveCheck *LiveCheck, states *atomic.Int64, traces *atomic.Int64, m2Mean *atomic.Int64) *SimulationWorker {
	if results == nil {
		results = make(chan SimulationWorkerResult, 1)
	}
	if maxTraceNum <= 0 {
		maxTraceNum = math.MaxInt64
	}
	return &SimulationWorker{
		ID:            id,
		Tool:          tool,
		Rand:          rand.New(rand.NewSource(seed)),
		ResultQueue:   results,
		MaxTraceNum:   maxTraceNum,
		MaxTraceDepth: maxTraceDepth,
		CheckDeadlock: checkDeadlock,
		Debug:         debug,
		TraceFile:     traceFile,
		LiveCheck:     liveCheck,
		Statistics:    NewSimulationWorkerStatistics(tool, traceActions, states, traces, m2Mean),
		NextStates:    NewStateVec(1),
	}
}

func (w *SimulationWorker) Stop() {
	if w != nil {
		w.Stopped.Store(true)
	}
}

func (w *SimulationWorker) Start(initStates *StateVec) {
	if w == nil {
		return
	}
	w.InitStates = initStates
	go w.Run()
}

func (w *SimulationWorker) Run() {
	for w != nil && !w.Stopped.Load() {
		if !w.SimulateAndReport() {
			return
		}
	}
}

func (w *SimulationWorker) SimulateAndReport() bool {
	if w == nil {
		return false
	}
	w.GlobalTrace = w.Statistics.CollectPreTrace()
	err := w.SimulateRandomTrace()
	w.TraceCnt++
	if err != nil {
		w.ResultQueue <- SimulationWorkerFailed(w.ID, err)
	}
	if w.TraceCnt >= w.MaxTraceNum || w.Stopped.Load() {
		w.ResultQueue <- SimulationWorkerOK(w.ID)
		return false
	}
	return true
}

func (w *SimulationWorker) RandomState(states *StateVec) *TLCStateMut {
	if states == nil || states.Size() == 0 {
		return nil
	}
	return states.At(w.Rand.Intn(states.Size()))
}

func (w *SimulationWorker) GetNextActionIndex(actions []*Action, curState *TLCStateMut) int {
	if len(actions) == 0 {
		return -1
	}
	return w.Rand.Intn(len(actions))
}

func (w *SimulationWorker) GetNextActionAltIndex(index int, p int, actions []*Action, curState *TLCStateMut) int {
	if w != nil && w.Statistics != nil {
		w.Statistics.CollectNextRetries()
	}
	if len(actions) == 0 {
		return -1
	}
	return (index + p) % len(actions)
}

func (w *SimulationWorker) SimulateRandomTrace() *SimulationWorkerError {
	if w == nil || w.Tool == nil {
		return &SimulationWorkerError{Code: ECGeneral, Err: newTLCError(ECGeneral, "simulation worker has no tool")}
	}
	w.CurState = w.RandomState(w.InitStates)
	if w.CurState != nil {
		w.CurState = w.CurState.DeepCopy()
	}
	actions := w.Tool.GetActions()
	for traceIdx := 0; traceIdx < w.MaxTraceDepth; traceIdx++ {
		if w.Stopped.Load() {
			return nil
		}
		w.NextStates.Clear()
		if len(actions) == 0 {
			if w.CheckDeadlock {
				return &SimulationWorkerError{Code: ECTLCDeadlockReached, StateTrace: w.GetTrace(w.CurState)}
			}
			break
		}
		index := w.GetNextActionIndex(actions, w.CurState)
		step := 1
		for i := 0; i < len(actions); i++ {
			action := actions[index]
			nextStates, err := w.Tool.GetNextStates(action, w.CurState)
			if err != nil {
				return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
			}
			if workerErr := w.AddGeneratedSuccessors(action, nextStates); workerErr != nil {
				return workerErr
			}
			if !w.NextStates.IsEmpty() {
				break
			}
			index = w.GetNextActionAltIndex(index, step, actions, w.CurState)
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
		if err := w.LiveCheck.CheckTrace(w.Tool, func() *StateVec { return w.GetTrace(w.CurState) }); err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
		}
	}
	w.Statistics.CollectPostTrace(w.CurState, w.MaxTraceDepth)
	if w.TraceFile != "" {
		if err := w.WriteTraceFile(); err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(w.CurState), Err: err}
		}
	}
	return nil
}

func (w *SimulationWorker) AddGeneratedSuccessors(action *Action, nextStates *StateVec) *SimulationWorkerError {
	if nextStates == nil {
		return nil
	}
	for i := 0; i < nextStates.Size(); i++ {
		succ := nextStates.At(i)
		if succ == nil {
			continue
		}
		if action != nil && action.CM.node != nil {
			action.CM.IncInvocations()
		}
		succ.SetPredecessor(w.CurState).SetAction(action)
		if !w.Tool.IsGoodState(succ) {
			return &SimulationWorkerError{Code: ECTLCStateNotCompletelySpecifiedNext, StateTrace: w.GetTrace(succ)}
		}
		w.Statistics.CollectPreSuccessor(w.CurState, action, succ)
		if workerErr := w.CheckInvariants(succ); workerErr != nil {
			return workerErr
		}
		if workerErr := w.CheckImpliedActions(succ); workerErr != nil {
			return workerErr
		}
		inModel, err := w.Tool.IsInModel(succ)
		if err != nil {
			return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(succ), Err: err}
		}
		if inModel {
			inActions, err := w.Tool.IsInActions(w.CurState, succ)
			if err != nil {
				return &SimulationWorkerError{Code: ECGeneral, StateTrace: w.GetTrace(succ), Err: err}
			}
			inModel = inActions
		}
		if inModel {
			if action != nil && action.CM.node != nil {
				action.CM.IncSecondary()
			}
			w.NextStates.Add(succ)
		}
	}
	return nil
}

func (w *SimulationWorker) CheckInvariants(state *TLCStateMut) *SimulationWorkerError {
	names := w.Tool.GetInvNames()
	for i, invariant := range w.Tool.GetInvariants() {
		valid, err := w.Tool.IsValidState(invariant, state)
		if err != nil {
			return &SimulationWorkerError{Code: ECTLCInvariantEvaluationFailed, Params: []string{nameAt(names, i), err.Error()}, StateTrace: w.GetTrace(state), Err: err}
		}
		if !valid {
			return &SimulationWorkerError{Code: ECTLCInvariantViolatedBehavior, Params: []string{nameAt(names, i)}, StateTrace: w.GetTrace(state)}
		}
	}
	return nil
}

func (w *SimulationWorker) CheckImpliedActions(state *TLCStateMut) *SimulationWorkerError {
	names := w.Tool.GetImpliedActNames()
	for i, action := range w.Tool.GetImpliedActions() {
		valid, err := w.Tool.IsValidTransition(action, w.CurState, state)
		if err != nil {
			return &SimulationWorkerError{Code: ECTLCActionPropertyEvaluationFailed, Params: []string{nameAt(names, i), err.Error()}, StateTrace: w.GetTrace(state), Err: err}
		}
		if !valid {
			return &SimulationWorkerError{Code: ECTLCActionPropertyViolatedBehavior, Params: []string{nameAt(names, i)}, StateTrace: w.GetTrace(state)}
		}
	}
	return nil
}

func (w *SimulationWorker) GetTrace(state *TLCStateMut) *StateVec {
	stack := make([]*TLCStateMut, 0)
	for cur := state; cur != nil; cur = cur.Predecessor() {
		stack = append(stack, cur)
	}
	trace := NewStateVec(len(stack))
	for i := len(stack) - 1; i >= 0; i-- {
		trace.Add(stack[i])
	}
	return trace
}

func (w *SimulationWorker) WriteTraceFile() error {
	trace := w.GetTrace(w.CurState)
	name := fmt.Sprintf("%s_%d_%d", w.TraceFile, w.ID, w.TraceCnt)
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := fmt.Fprintf(file, "---------------- MODULE %s -----------------\n", name); err != nil {
		return err
	}
	for i := 0; i < trace.Size(); i++ {
		state := trace.At(i)
		if state.GetAction() != nil {
			if _, err := fmt.Fprintf(file, "\\* %s\n", state.GetAction().GetLocation()); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(file, "STATE_%d == \n%s\n\n", i+1, state.String()); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(file, "=================================================")
	return err
}

func actionIDFromStateAction(stats *SimulationWorkerStatistics, action *Action) int {
	if action == nil {
		return -1
	}
	if action.ID >= 0 {
		return action.ID
	}
	if stats == nil || stats.workerActionIDs == nil {
		return -1
	}
	id, ok := stats.workerActionIDs.Get2(UniqueStringOf(action.GetName()))
	if !ok {
		return -1
	}
	return id
}
