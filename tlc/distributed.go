package tlc

import (
	"sort"
	"sync/atomic"
	"time"
)

const (
	TLCServerName             = "TLCServer"
	TLCServerWorkerName       = TLCServerName + "WORKER"
	TLCWorkerThreadNamePrefix = "TLCWorkerThread-"
	TLCServerDefaultPort      = 10997
)

type NextStateResult struct {
	ComputationTime  int64
	StatesComputed   int64
	NextStates       []*StateVec
	NextFingerprints []*LongVec
}

func NewNextStateResult(nextStates []*StateVec, nextFingerprints []*LongVec, computationTime int64, statesComputed int64) *NextStateResult {
	return &NextStateResult{
		NextStates:       nextStates,
		NextFingerprints: nextFingerprints,
		ComputationTime:  computationTime,
		StatesComputed:   statesComputed,
	}
}

func (r *NextStateResult) GetStatesComputedDelta() int64 {
	if r == nil {
		return 0
	}
	return r.StatesComputed - int64(len(r.NextStates))
}

func (r *NextStateResult) GetComputationTime() int64 {
	if r == nil {
		return 0
	}
	return r.ComputationTime
}

func (r *NextStateResult) GetNextFingerprints() []*LongVec {
	if r == nil {
		return nil
	}
	return r.NextFingerprints
}

func (r *NextStateResult) GetNextStates() []*StateVec {
	if r == nil {
		return nil
	}
	return r.NextStates
}

type DistributedFPSetManager struct {
	Sets        []FPSet
	Shift       uint
	StatesSeen  atomic.Uint64
	Description string
}

func NewDistributedFPSetManager(sets ...FPSet) *DistributedFPSetManager {
	out := &DistributedFPSetManager{Sets: append([]FPSet(nil), sets...)}
	out.normalize()
	return out
}

func NewDistributedFPSetManagerFromFPSet(set FPSet) *DistributedFPSetManager {
	if multi, ok := set.(*MultiFPSet); ok && multi != nil {
		return &DistributedFPSetManager{
			Sets:  append([]FPSet(nil), multi.Sets...),
			Shift: multi.Shift,
		}
	}
	return NewDistributedFPSetManager(set)
}

func (m *DistributedFPSetManager) normalize() {
	if m == nil || len(m.Sets) <= 1 || m.Shift != 0 {
		return
	}
	bits := 0
	for count := len(m.Sets); count > 1; count >>= 1 {
		bits++
	}
	m.Shift = uint(64 - bits)
}

func (m *DistributedFPSetManager) NumOfServers() int {
	if m == nil || len(m.Sets) == 0 {
		return 1
	}
	return len(m.Sets)
}

func (m *DistributedFPSetManager) GetFPSetIndex(fp uint64) int {
	if m == nil || len(m.Sets) <= 1 {
		return 0
	}
	index := int(fp >> m.Shift)
	if index < 0 {
		return 0
	}
	if index >= len(m.Sets) {
		index %= len(m.Sets)
	}
	return index
}

func (m *DistributedFPSetManager) ContainsBlock(fingerprints []*LongVec) []*BitVector {
	out := make([]*BitVector, len(fingerprints))
	for i, fpv := range fingerprints {
		size := 0
		if fpv != nil {
			size = fpv.Size()
		}
		if m == nil || i >= len(m.Sets) || m.Sets[i] == nil {
			out[i] = newAllTrueBitVector(size)
			m.addStatesSeen(uint64(size))
			continue
		}
		out[i] = m.Sets[i].ContainsBlock(fpv)
	}
	return out
}

func (m *DistributedFPSetManager) PutBlock(fingerprints []*LongVec) []*BitVector {
	out := make([]*BitVector, len(fingerprints))
	for i, fpv := range fingerprints {
		size := 0
		if fpv != nil {
			size = fpv.Size()
		}
		if m == nil || i >= len(m.Sets) || m.Sets[i] == nil {
			out[i] = newAllTrueBitVector(size)
			m.addStatesSeen(uint64(size))
			continue
		}
		out[i] = m.Sets[i].PutBlock(fpv)
	}
	return out
}

func (m *DistributedFPSetManager) Size() uint64 {
	if m == nil {
		return 0
	}
	var size uint64
	for _, set := range m.Sets {
		if set != nil {
			size += set.Size()
		}
	}
	return size
}

func (m *DistributedFPSetManager) GetStatesSeen() uint64 {
	if m == nil {
		return 0
	}
	total := m.StatesSeen.Load()
	for _, set := range m.Sets {
		if set != nil {
			total += set.GetStatesSeen()
		}
	}
	return total
}

func (m *DistributedFPSetManager) Close(cleanup bool) error {
	if m == nil {
		return nil
	}
	for _, set := range m.Sets {
		if set != nil {
			if err := set.Exit(cleanup); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) Checkpoint(filename string) error {
	if m == nil {
		return nil
	}
	for _, set := range m.Sets {
		if set != nil {
			if err := set.BeginChkptFile(filename); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) CommitCheckpoint() error {
	if m == nil {
		return nil
	}
	for _, set := range m.Sets {
		if set != nil {
			if err := set.CommitChkpt(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) Recover(filename string) error {
	if m == nil {
		return nil
	}
	for _, set := range m.Sets {
		if set != nil {
			if err := set.RecoverFile(filename); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) addStatesSeen(delta uint64) {
	if m != nil && delta != 0 {
		m.StatesSeen.Add(delta)
	}
}

type DistributedWorker struct {
	ID                    int
	Tool                  *Tool
	FPSetManager          *DistributedFPSetManager
	Cache                 *SimpleCache
	URI                   string
	Computing             atomic.Bool
	LastInvocation        time.Time
	OverallStatesComputed int64
	CheckStateFunc        func(predecessor *TLCStateMut, successor *TLCStateMut) error
	IsInModelFunc         func(state *TLCStateMut) (bool, error)
	IsInActionsFunc       func(predecessor *TLCStateMut, successor *TLCStateMut) (bool, error)
}

func NewDistributedWorker(id int, tool *Tool, fpSetManager *DistributedFPSetManager) *DistributedWorker {
	if fpSetManager == nil {
		fpSetManager = NewDistributedFPSetManager()
	}
	return &DistributedWorker{
		ID:           id,
		Tool:         tool,
		FPSetManager: fpSetManager,
		Cache:        NewSimpleCache(),
	}
}

func (w *DistributedWorker) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	if w == nil {
		return nil, newTLCError(ECGeneral, "distributed worker is nil")
	}
	w.Computing.Store(true)
	w.LastInvocation = time.Now()
	var statesComputed int64
	var state1, state2 *TLCStateMut
	defer w.Computing.Store(false)

	holdersByFP := make(map[uint64]distributedStateHolder, len(states))
	orderedFPs := make([]uint64, 0, len(states))
	for _, state := range states {
		state1 = state
		nextStates, err := w.computeNextStates(state)
		if err != nil {
			return nil, NewWorkerException(err.Error(), err, state1, state2, true)
		}
		statesComputed += int64(nextStates.Size())
		for i := 0; i < nextStates.Size(); i++ {
			state2 = nextStates.At(i)
			fp := state2.FingerPrint()
			if w.Cache != nil && w.Cache.Hit(fp) {
				continue
			}
			if _, ok := holdersByFP[fp]; ok {
				continue
			}
			holdersByFP[fp] = distributedStateHolder{Fingerprint: fp, Successor: state2, Predecessor: state1}
			orderedFPs = append(orderedFPs, fp)
		}
	}
	w.OverallStatesComputed += statesComputed

	sort.Slice(orderedFPs, func(i, j int) bool { return orderedFPs[i] < orderedFPs[j] })
	serverCount := w.FPSetManager.NumOfServers()
	predecessors := make([]*StateVec, serverCount)
	successors := make([]*StateVec, serverCount)
	fingerprints := make([]*LongVec, serverCount)
	for i := 0; i < serverCount; i++ {
		predecessors[i] = NewStateVec(0)
		successors[i] = NewStateVec(0)
		fingerprints[i] = NewLongVec()
	}
	for _, fp := range orderedFPs {
		holder := holdersByFP[fp]
		fpIndex := w.FPSetManager.GetFPSetIndex(fp)
		predecessors[fpIndex].Add(holder.Predecessor)
		successors[fpIndex].Add(holder.Successor)
		fingerprints[fpIndex].AddElement(int64(fp))
	}

	visited := w.FPSetManager.ContainsBlock(fingerprints)
	newStates := make([]*StateVec, serverCount)
	newFingerprints := make([]*LongVec, serverCount)
	for i := 0; i < serverCount; i++ {
		newStates[i] = NewStateVec(0)
		newFingerprints[i] = NewLongVec()
	}
	for i := 0; i < serverCount; i++ {
		iter := NewBitVectorIter(visited[i])
		for {
			index := iter.Next()
			if index == -1 {
				break
			}
			state1 = predecessors[i].At(index)
			state2 = successors[i].At(index)
			if err := w.CheckState(state1, state2); err != nil {
				return nil, NewWorkerException(err.Error(), err, state1, state2, true)
			}
			inModel, err := w.IsInModel(state2)
			if err != nil {
				return nil, NewWorkerException(err.Error(), err, state1, state2, true)
			}
			inActions, err := w.IsInActions(state1, state2)
			if err != nil {
				return nil, NewWorkerException(err.Error(), err, state1, state2, true)
			}
			if inModel && inActions {
				state2.UID = state1.UID
				newStates[i].Add(state2)
				newFingerprints[i].AddElement(fingerprints[i].ElementAt(index))
			}
		}
	}

	elapsed := time.Since(w.LastInvocation).Milliseconds()
	return NewNextStateResult(newStates, newFingerprints, elapsed, statesComputed), nil
}

func (w *DistributedWorker) computeNextStates(state *TLCStateMut) (*StateVec, error) {
	out := NewStateVec(0)
	if w == nil || w.Tool == nil {
		return out, nil
	}
	functor := NewNextStateFunctor(func(curState *TLCStateMut, action *Action, succState *TLCStateMut) (any, error) {
		if succState != nil {
			succState.SetPredecessor(curState).SetAction(action)
			out.Add(succState)
		}
		return out, nil
	})
	_, err := w.Tool.GetNextStatesWithFunctor(functor, state)
	return out, err
}

func (w *DistributedWorker) CheckState(predecessor *TLCStateMut, successor *TLCStateMut) error {
	if w != nil && w.CheckStateFunc != nil {
		return w.CheckStateFunc(predecessor, successor)
	}
	if w == nil || w.Tool == nil || successor == nil || w.Tool.IsGoodState(successor) {
		return nil
	}
	return newTLCError(ECTLCStateNotCompletelySpecifiedNext, "%s", successor)
}

func (w *DistributedWorker) IsInModel(state *TLCStateMut) (bool, error) {
	if w != nil && w.IsInModelFunc != nil {
		return w.IsInModelFunc(state)
	}
	if w == nil || w.Tool == nil {
		return true, nil
	}
	return w.Tool.IsInModel(state)
}

func (w *DistributedWorker) IsInActions(predecessor *TLCStateMut, successor *TLCStateMut) (bool, error) {
	if w != nil && w.IsInActionsFunc != nil {
		return w.IsInActionsFunc(predecessor, successor)
	}
	if w == nil || w.Tool == nil {
		return true, nil
	}
	return w.Tool.IsInActions(predecessor, successor)
}

func (w *DistributedWorker) IsAlive() bool {
	return w != nil
}

func (w *DistributedWorker) GetCacheRateRatio() float64 {
	if w == nil || w.Cache == nil {
		return 0
	}
	return w.Cache.GetHitRatio()
}

func (w *DistributedWorker) IsComputing() bool {
	return w != nil && w.Computing.Load()
}

type distributedStateHolder struct {
	Fingerprint uint64
	Successor   *TLCStateMut
	Predecessor *TLCStateMut
}

func newAllTrueBitVector(size int) *BitVector {
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		bv.Set(i)
	}
	return bv
}
