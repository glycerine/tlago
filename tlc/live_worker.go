package tlc

import "fmt"

const liveWorkerSCCMarker = int64(-42)

type LiveWorker struct {
	Tool       *Tool
	ID         int
	NumWorkers int
	Check      *LiveCheck
	Checker    *LiveChecker
	FinalCheck bool
	Solution   *OrderOfSolution
	PEM        *PossibleErrorModel
}

type liveWorkerNodeEntry struct {
	state uint64
	tidx  int
	loc   int64
}

func NewLiveWorker(tool *Tool, id int, numWorkers int, liveCheck *LiveCheck, checker *LiveChecker, pem *PossibleErrorModel, finalCheck bool) *LiveWorker {
	worker := &LiveWorker{
		Tool:       tool,
		ID:         id,
		NumWorkers: numWorkers,
		Check:      liveCheck,
		Checker:    checker,
		FinalCheck: finalCheck,
		PEM:        pem,
	}
	if checker != nil {
		worker.Solution = checker.Solution
	}
	return worker
}

func (w *LiveWorker) CheckSccs() (bool, error) {
	if w == nil || w.Checker == nil || w.Solution == nil || w.PEM == nil {
		return false, nil
	}
	if err := w.makeNodePtrTable(); err != nil {
		return false, err
	}

	initNodes := w.initNodes()
	numOfInits := initNodes.Size()
	nodeQueue := make([]liveWorkerNodeEntry, 0, max(numOfInits/2, 1))
	for j := 0; j < numOfInits; j += 2 {
		state := uint64(initNodes.ElementAt(j))
		tidx := int(initNodes.ElementAt(j + 1))
		ptr := w.getLink(state, tidx)
		if ptr >= 0 {
			nodeQueue = append(nodeQueue, liveWorkerNodeEntry{state: state, tidx: tidx, loc: ptr})
		} else if w.FinalCheck && (ptr == TableauNodePtrTableUndone || ptr == TableauNodePtrTableDone) {
			return false, fmt.Errorf("final liveness check found malformed initial node link %d", ptr)
		}
	}

	eaaction := w.PEM.EAAction
	slen := len(w.Solution.CheckState)
	alen := len(w.Solution.CheckAction)
	dfsStack := NewIntStack()
	comStack := NewIntStack()

	for head := 0; head < len(nodeQueue); head++ {
		entry := nodeQueue[head]
		dfsStack.Reset()
		dfsStack.PushLong(int64(entry.state))
		dfsStack.PushInt(int32(entry.tidx))
		dfsStack.PushLong(entry.loc)
		dfsStack.PushLong(DiskGraphMaxPtr)
		newLink := DiskGraphMaxPtr

		for dfsStack.Size() >= 7 {
			lowLink := dfsStack.PopLong()
			curLoc := dfsStack.PopLong()
			curTidx := int(dfsStack.PopInt())
			curState := uint64(dfsStack.PopLong())

			if curLoc == liveWorkerSCCMarker {
				curLink := w.getLink(curState, curTidx)
				if curLink == lowLink {
					ok, err := w.checkComponent(curState, curTidx, comStack)
					if err != nil || !ok {
						return !ok, err
					}
				}
				plowLink := dfsStack.PopLong()
				dfsStack.PushLong(minInt64(plowLink, lowLink))
				continue
			}

			link := w.putLink(curState, curTidx, newLink)
			if link == -1 {
				dfsStack.PushLong(lowLink)
				dfsStack.PushLong(int64(curState))
				dfsStack.PushInt(int32(curTidx))
				dfsStack.PushLong(liveWorkerSCCMarker)

				comStack.PushLong(curLoc)
				comStack.PushInt(int32(curTidx))
				comStack.PushLong(int64(curState))

				gnode, err := w.getNode(curState, curTidx, curLoc)
				if err != nil {
					return false, err
				}
				nextLowLink := newLink
				newLink++
				for i := 0; i < gnode.SuccSize(); i++ {
					nextState := gnode.GetStateFP(i)
					nextTidx := gnode.GetTIndex(i)
					nextLink := w.getLink(nextState, nextTidx)
					if nextLink >= 0 {
						if gnode.GetCheckActionAll(slen, alen, i, eaaction) {
							if IsDiskGraphFilePointer(nextLink) {
								dfsStack.PushLong(int64(nextState))
								dfsStack.PushInt(int32(nextTidx))
								dfsStack.PushLong(nextLink)
							} else {
								nextLowLink = minInt64(nextLowLink, nextLink)
							}
						} else if IsDiskGraphFilePointer(nextLink) {
							nodeQueue = append(nodeQueue, liveWorkerNodeEntry{state: nextState, tidx: nextTidx, loc: nextLink})
						}
					} else if w.FinalCheck && nextLink == TableauNodePtrTableUndone {
						return false, fmt.Errorf("final liveness check found undone successor link %d", nextLink)
					}
				}
				dfsStack.PushLong(nextLowLink)
			} else {
				dfsStack.PushLong(minInt64(lowLink, link))
			}
		}
	}
	if comStack.Size() != 0 {
		return false, fmt.Errorf("liveness component stack not empty after SCC search")
	}
	return false, nil
}

func (w *LiveWorker) checkComponent(state uint64, tidx int, comStack *IntStack) (bool, error) {
	if comStack == nil || comStack.Size() < 5 || comStack.Size()%5 != 0 {
		return false, fmt.Errorf("malformed liveness component stack")
	}
	comStackSize := comStack.Size()
	state1 := uint64(comStack.PopLong())
	tidx1 := int(comStack.PopInt())
	loc1 := comStack.PopLong()

	if state1 == state && tidx1 == tidx {
		stuttering, err := w.isStuttering(state1, tidx1, loc1)
		if err != nil {
			return false, err
		}
		if !stuttering {
			w.setMaxLink(state, tidx)
			return true, nil
		}
	}

	com := NewTableauNodePtrTable(128)
	for {
		com.PutElem(state1, tidx1, loc1)
		w.setMaxLink(state1, tidx1)
		if state == state1 && tidx == tidx1 {
			break
		}
		if comStack.Size() < 5 {
			return false, fmt.Errorf("liveness component stack ended before component root")
		}
		state1 = uint64(comStack.PopLong())
		tidx1 = int(comStack.PopInt())
		loc1 = comStack.PopLong()
	}
	if com.Size() > comStackSize/5 {
		return false, fmt.Errorf("liveness component table larger than source stack")
	}

	slen := len(w.Solution.CheckState)
	alen := len(w.Solution.CheckAction)
	aeslen := len(w.PEM.AEState)
	aealen := len(w.PEM.AEAction)
	plen := len(w.Solution.Promises)
	aeStateRes := make([]bool, aeslen)
	aeActionRes := make([]bool, aealen)
	promiseRes := make([]bool, plen)
	eaaction := w.PEM.EAAction

	for ci := 0; ci < com.GetSize(); ci++ {
		nodes := com.GetNodesByLoc(ci)
		if nodes == nil {
			continue
		}
		state1 = TableauGetKey(nodes)
		for nidx := 2; nidx < len(nodes); nidx += com.GetElemLength() {
			tidx1 = TableauGetTidx(nodes, nidx)
			loc1 = TableauGetElem(nodes, nidx)
			curNode, err := w.getNode(state1, tidx1, loc1)
			if err != nil {
				return false, err
			}
			for i, idx := range w.PEM.AEState {
				if !aeStateRes[i] {
					aeStateRes[i] = curNode.GetCheckState(idx)
				}
			}
			if aealen > 0 {
				for i := 0; i < curNode.SuccSize(); i++ {
					nextState := curNode.GetStateFP(i)
					nextTidx := curNode.GetTIndex(i)
					if com.GetLoc(nextState, nextTidx) == -1 {
						continue
					}
					if !curNode.GetCheckActionAll(slen, alen, i, eaaction) {
						continue
					}
					for j, idx := range w.PEM.AEAction {
						if !aeActionRes[j] {
							aeActionRes[j] = curNode.GetCheckAction(slen, alen, i, idx)
						}
					}
				}
			}
			if w.Solution.HasTableau() && curNode.TIndex >= 0 && curNode.TIndex < w.Solution.Tableau.Size() {
				par := curNode.GetTNode(w.Solution.Tableau).Par
				for i, promise := range w.Solution.Promises {
					if !promiseRes[i] && par.IsFulfilling(promise) {
						promiseRes[i] = true
					}
				}
			}
		}
	}

	for _, ok := range aeStateRes {
		if !ok {
			return true, nil
		}
	}
	for _, ok := range aeActionRes {
		if !ok {
			return true, nil
		}
	}
	for _, ok := range promiseRes {
		if !ok {
			return true, nil
		}
	}
	w.Checker.ErrorGraphNode = NewGraphNode(state, tidx)
	return false, nil
}

func (w *LiveWorker) isStuttering(state uint64, tidx int, loc int64) (bool, error) {
	gnode, err := w.getNode(state, tidx, loc)
	if err != nil {
		return false, err
	}
	slen := len(w.Solution.CheckState)
	alen := len(w.Solution.CheckAction)
	for i := 0; i < gnode.SuccSize(); i++ {
		if gnode.GetStateFP(i) == state && gnode.GetTIndex(i) == tidx {
			return gnode.GetCheckActionAll(slen, alen, i, w.PEM.EAAction), nil
		}
	}
	return false, nil
}

func (w *LiveWorker) makeNodePtrTable() error {
	if w.Checker.TableauDiskGraph != nil {
		return w.Checker.TableauDiskGraph.MakeNodePtrTbl()
	}
	if w.Checker.DiskGraph != nil {
		return w.Checker.DiskGraph.MakeNodePtrTbl()
	}
	return nil
}

func (w *LiveWorker) initNodes() *LongVec {
	if w.Checker.TableauDiskGraph != nil {
		return w.Checker.TableauDiskGraph.GetInitNodes()
	}
	if w.Checker.DiskGraph != nil {
		return w.Checker.DiskGraph.GetInitNodes()
	}
	return NewLongVecWithCapacity(0)
}

func (w *LiveWorker) getLink(state uint64, tidx int) int64 {
	if w.Checker.TableauDiskGraph != nil {
		return w.Checker.TableauDiskGraph.GetLink(state, tidx)
	}
	if w.Checker.DiskGraph != nil {
		return w.Checker.DiskGraph.GetLink(state, tidx)
	}
	return -1
}

func (w *LiveWorker) putLink(state uint64, tidx int, link int64) int64 {
	if w.Checker.TableauDiskGraph != nil {
		return w.Checker.TableauDiskGraph.PutLink(state, tidx, link)
	}
	if w.Checker.DiskGraph != nil {
		return w.Checker.DiskGraph.PutLink(state, tidx, link)
	}
	return -1
}

func (w *LiveWorker) setMaxLink(state uint64, tidx int) {
	if w.Checker.TableauDiskGraph != nil {
		w.Checker.TableauDiskGraph.SetMaxLink(state, tidx)
		return
	}
	if w.Checker.DiskGraph != nil {
		w.Checker.DiskGraph.SetMaxLink(state, tidx)
	}
}

func (w *LiveWorker) getNode(state uint64, tidx int, loc int64) (*GraphNode, error) {
	if w.Checker.TableauDiskGraph != nil {
		return w.Checker.TableauDiskGraph.diskGraphNodeAt(state, tidx, loc)
	}
	if w.Checker.DiskGraph != nil {
		return w.Checker.DiskGraph.diskGraphNodeAt(state, tidx, loc)
	}
	return nil, fmt.Errorf("liveness worker has no disk graph")
}

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
