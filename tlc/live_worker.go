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
	prefix, cycle, err := w.traceFingerprintLasso(state, tidx, com)
	if err != nil {
		return false, err
	}
	w.Checker.ErrorPrefix = prefix
	w.Checker.ErrorCycle = cycle
	counterExample, printableTrace, rawTrace, closingInfo, loopOrdinal, cycleIndex, stuttering, err := w.buildCounterExample(prefix, cycle)
	if err != nil {
		return false, err
	}
	w.Checker.ErrorTrace = printableTrace
	w.Checker.ErrorRawTrace = rawTrace
	w.Checker.ErrorCounterEx = counterExample
	w.Checker.ErrorClosingInfo = closingInfo
	w.Checker.ErrorLoopOrdinal = loopOrdinal
	w.Checker.ErrorCycleIndex = cycleIndex
	w.Checker.ErrorStuttering = stuttering
	if w.Tool != nil && counterExample != nil {
		w.Tool.CheckPostConditionWithCounterExample(counterExample)
	}
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

func (w *LiveWorker) traceFingerprintLasso(state uint64, tidx int, nodeTbl *TableauNodePtrTable) (*LongVec, *LongVec, error) {
	w.createCache()
	defer w.destroyCache()

	prefix, err := w.getPath(state, tidx)
	if err != nil {
		return nil, nil, err
	}
	cycleStack := NewIntStack()
	curNode, err := w.dfsPostFix(state, tidx, nodeTbl, cycleStack)
	if err != nil {
		return nil, nil, err
	}
	postfix, err := w.bfsPostFix(state, tidx, nodeTbl, curNode)
	if err != nil {
		return nil, nil, err
	}
	for cycleStack.Size() > 0 {
		fp := cycleStack.PopLong()
		if postfix.IsEmpty() || postfix.LastElement() != fp {
			postfix.AddElement(fp)
		}
		_ = cycleStack.PopInt()
	}
	return prefix, postfix, nil
}

func (w *LiveWorker) buildCounterExample(prefix *LongVec, cycle *LongVec) (*CounterExample, []*TLCStateInfo, []*TLCStateInfo, *TLCStateInfo, int, int, bool, error) {
	if w.Tool == nil {
		return nil, nil, nil, nil, 0, 0, false, fmt.Errorf("cannot reconstruct liveness counterexample without a tool")
	}
	if prefix == nil || prefix.Size() == 0 {
		return nil, nil, nil, nil, 0, 0, false, fmt.Errorf("cannot reconstruct liveness counterexample without a prefix")
	}
	plen := prefix.Size()
	fp := uint64(prefix.ElementAt(plen - 1))
	sinfo, err := w.Tool.GetState(fp)
	if err != nil {
		return nil, nil, nil, nil, 0, 0, false, err
	}
	if sinfo == nil {
		return nil, nil, nil, nil, 0, 0, false, fmt.Errorf("failed to recover initial liveness state %d", fp)
	}
	states := make([]*TLCStateInfo, 0, plen)
	states = append(states, sinfo)

	for i := plen - 2; i >= 0; i-- {
		curFP := uint64(prefix.ElementAt(i))
		if curFP == fp {
			continue
		}
		sinfo, err = w.Tool.GetState(curFP, sinfo)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, false, err
		}
		if sinfo == nil {
			return nil, nil, nil, nil, 0, 0, false, fmt.Errorf("failed to recover liveness successor state %d", curFP)
		}
		states = append(states, sinfo)
		fp = curFP
	}
	if len(states) == 0 {
		return nil, nil, nil, nil, 0, 0, false, fmt.Errorf("cannot reconstruct empty liveness counterexample")
	}

	printableStates := make([]*TLCStateInfo, 0, len(states)+16)
	for i := 0; i < len(states)-1; i++ {
		alias, err := w.Tool.EvalAliasInfoPrefix(states[i], states[i+1].State, states[:i])
		if err != nil {
			return nil, nil, nil, nil, 0, 0, false, err
		}
		printableStates = append(printableStates, alias)
	}

	cycleState := states[len(states)-1]
	cycleIndex := len(states) - 1
	loopOrdinal := int(cycleState.StateNumber)
	sinfo = cycleState
	closingInfo := sinfo
	stuttering := true
	if cycle != nil && !cycle.IsEmpty() {
		cycle.Pack().RemoveLastIf(int64(cycleState.FingerPrint()))
		for i := cycle.Size() - 1; i >= 0; i-- {
			curFP := uint64(cycle.ElementAt(i))
			sucinfo, err := w.Tool.GetState(curFP, sinfo)
			if err != nil {
				return nil, nil, nil, nil, 0, 0, false, err
			}
			if sucinfo == nil {
				return nil, nil, nil, nil, 0, 0, false, fmt.Errorf("failed to recover liveness cycle state %d", curFP)
			}
			alias, err := w.Tool.EvalAliasInfoPrefix(sinfo, sucinfo.State, states)
			if err != nil {
				return nil, nil, nil, nil, 0, 0, false, err
			}
			printableStates = append(printableStates, alias)
			states = append(states, sucinfo)
			sinfo = sucinfo
		}
		alias, err := w.Tool.EvalAliasInfoPrefix(sinfo, cycleState.State, states)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, false, err
		}
		printableStates = append(printableStates, alias)
		if sinfo.FingerPrint() != cycleState.FingerPrint() {
			stuttering = false
			closing, err := w.Tool.GetState(cycleState.FingerPrint(), sinfo)
			if err != nil {
				return nil, nil, nil, nil, 0, 0, false, err
			}
			if closing != nil {
				closingInfo = closing
			}
		} else {
			closingInfo = sinfo
		}
	} else {
		alias, err := w.Tool.EvalAliasInfoPrefix(cycleState, cycleState.State, states)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, false, err
		}
		printableStates = append(printableStates, alias)
	}
	return NewCounterExample(states, closingInfo.Action(), loopOrdinal, true), printableStates, states, closingInfo, loopOrdinal, cycleIndex, stuttering, nil
}

func (w *LiveWorker) dfsPostFix(state uint64, tidx int, nodeTbl *TableauNodePtrTable, cycleStack *IntStack) (*GraphNode, error) {
	slen := len(w.Solution.CheckState)
	alen := len(w.Solution.CheckAction)
	aeStateRes := make([]bool, len(w.PEM.AEState))
	aeActionRes := make([]bool, len(w.PEM.AEAction))
	promiseRes := make([]bool, len(w.Solution.Promises))
	eaaction := w.PEM.EAAction
	cnt := len(aeStateRes) + len(aeActionRes) + len(promiseRes)

	nodes := nodeTbl.GetNodes(state)
	if nodes == nil {
		return nil, fmt.Errorf("liveness SCC missing start state %d", state)
	}
	tloc := nodeTbl.GetIdx(nodes, tidx)
	if tloc == -1 {
		return nil, fmt.Errorf("liveness SCC missing start tableau index %d for state %d", tidx, state)
	}
	ptr := TableauGetElem(nodes, tloc)
	TableauSetSeenAt(nodes, tloc)
	curNode, err := w.getNode(state, tidx, ptr)
	if err != nil {
		return nil, err
	}

	for cnt > 0 {
		cnt0 := cnt
	next:
		for {
			for i, idx := range w.PEM.AEState {
				if !aeStateRes[i] && curNode.GetCheckState(idx) {
					aeStateRes[i] = true
					cnt--
				}
			}
			if w.Solution.HasTableau() && curNode.TIndex >= 0 && curNode.TIndex < w.Solution.Tableau.Size() {
				par := curNode.GetTNode(w.Solution.Tableau).Par
				for i, promise := range w.Solution.Promises {
					if !promiseRes[i] && par.IsFulfilling(promise) {
						promiseRes[i] = true
						cnt--
					}
				}
			}
			if cnt <= 0 {
				break
			}

			var nextState1 uint64
			var nextState2 uint64
			nextTidx1 := 0
			nextTidx2 := 0
			tloc1 := -1
			tloc2 := -1
			var nodes1 []int32
			var nodes2 []int32
			hasUnvisitedSucc := false
			cnt1 := cnt
			for i := 0; i < curNode.SuccSize(); i++ {
				nextState := curNode.GetStateFP(i)
				nextTidx := curNode.GetTIndex(i)
				nodes = nodeTbl.GetNodes(nextState)
				tloc = -1
				if nodes != nil {
					tloc = nodeTbl.GetIdx(nodes, nextTidx)
					if tloc == -1 {
						continue
					}
					if !curNode.GetCheckActionAll(slen, alen, i, eaaction) {
						continue
					}
					nextState1 = nextState
					nextTidx1 = nextTidx
					tloc1 = tloc
					nodes1 = nodes
					for j, idx := range w.PEM.AEAction {
						if !aeActionRes[j] && curNode.GetCheckAction(slen, alen, i, idx) {
							aeActionRes[j] = true
							cnt--
						}
					}
				}
				if cnt < cnt1 {
					cycleStack.PushInt(int32(curNode.TIndex))
					cycleStack.PushLong(int64(curNode.StateFP))
					nextPtr := TableauGetPtr(TableauGetElem(nodes, tloc))
					curNode, err = w.getNode(nextState, nextTidx, nextPtr)
					if err != nil {
						return nil, err
					}
					nodeTbl.ResetElems()
					break next
				}
				if nodes != nil && tloc != -1 && !TableauIsSeenAt(nodes, tloc) {
					hasUnvisitedSucc = true
					nextState2 = nextState
					nextTidx2 = nextTidx
					tloc2 = tloc
					nodes2 = nodes
				}
			}

			if cnt < cnt0 {
				cycleStack.PushInt(int32(curNode.TIndex))
				cycleStack.PushLong(int64(curNode.StateFP))
				nextPtr := TableauGetPtr(TableauGetElem(nodes1, tloc1))
				curNode, err = w.getNode(nextState1, nextTidx1, nextPtr)
				if err != nil {
					return nil, err
				}
				nodeTbl.ResetElems()
				break
			}

			for !hasUnvisitedSucc {
				if cycleStack.Size() < 3 {
					return nil, fmt.Errorf("liveness DFS postfix could not find unvisited successor")
				}
				curState := uint64(cycleStack.PopLong())
				curTidx := int(cycleStack.PopInt())
				curPtr := TableauGetPtr(nodeTbl.Get(curState, curTidx))
				curNode, err = w.getNode(curState, curTidx, curPtr)
				if err != nil {
					return nil, err
				}
				for i := 0; i < curNode.SuccSize(); i++ {
					nextState2 = curNode.GetStateFP(i)
					nextTidx2 = curNode.GetTIndex(i)
					nodes2 = nodeTbl.GetNodes(nextState2)
					if nodes2 != nil {
						tloc2 = nodeTbl.GetIdx(nodes2, nextTidx2)
						if tloc2 != -1 && !TableauIsSeenAt(nodes2, tloc2) {
							hasUnvisitedSucc = true
							break
						}
					}
				}
			}

			cycleStack.PushInt(int32(curNode.TIndex))
			cycleStack.PushLong(int64(curNode.StateFP))
			nextPtr := TableauGetPtr(TableauGetElem(nodes2, tloc2))
			curNode, err = w.getNode(nextState2, nextTidx2, nextPtr)
			if err != nil {
				return nil, err
			}
			TableauSetSeenAt(nodes2, tloc2)
		}
	}
	nodeTbl.ResetElems()
	return curNode, nil
}

func (w *LiveWorker) bfsPostFix(state uint64, tidx int, nodeTbl *TableauNodePtrTable, curNode *GraphNode) (*LongVec, error) {
	slen := len(w.Solution.CheckState)
	alen := len(w.Solution.CheckAction)
	eaaction := w.PEM.EAAction
	postfix := NewLongVecWithCapacity(16)
	startState := curNode.StateFP
	startTidx := curNode.TIndex

	if startState == state && startTidx == tidx {
		return postfix, nil
	}

	type postfixEntry struct {
		state uint64
		ploc  int
	}
	queue := make([]postfixEntry, 0)
	curState := startState
	ploc := TableauNoParent
	curLoc := nodeTbl.GetNodesLoc(curState)
	nodes := nodeTbl.GetNodesByLoc(curLoc)
	if nodes == nil {
		return nil, fmt.Errorf("liveness BFS postfix missing start state %d", curState)
	}
	TableauSetSeen(nodes)

	for {
		tloc := TableauStartLoc(nodes)
		for tloc != TableauEndMarker {
			curTidx := TableauGetTidx(nodes, tloc)
			curPtr := TableauGetPtr(TableauGetElem(nodes, tloc))
			curNode, err := w.getNode(curState, curTidx, curPtr)
			if err != nil {
				return nil, err
			}
			for j := 0; j < curNode.SuccSize(); j++ {
				nextState := curNode.GetStateFP(j)
				nextTidx := curNode.GetTIndex(j)
				if curState == nextState && curTidx == nextTidx {
					continue
				}
				if !curNode.GetCheckActionAll(slen, alen, j, eaaction) {
					continue
				}
				if nextState == state && nextTidx == tidx {
					for curState != startState {
						postfix.AddElement(int64(curState))
						nodes = nodeTbl.GetNodesByLoc(ploc)
						if nodes == nil {
							return nil, fmt.Errorf("liveness BFS postfix missing parent at %d", ploc)
						}
						curState = TableauGetKey(nodes)
						ploc = TableauGetParent(nodes)
					}
					postfix.AddElement(int64(startState))
					return postfix, nil
				}

				nodes1 := nodeTbl.GetNodes(nextState)
				if nodes1 != nil && !TableauIsSeen(nodes1) {
					TableauSetSeen(nodes1)
					queue = append(queue, postfixEntry{state: nextState, ploc: curLoc})
				}
			}
			tloc = TableauNextLoc(nodes, tloc)
		}
		TableauSetParent(nodes, ploc)
		if len(queue) == 0 {
			return nil, fmt.Errorf("liveness BFS postfix could not close cycle")
		}
		entry := queue[0]
		queue = queue[1:]
		curState = entry.state
		ploc = entry.ploc
		curLoc = nodeTbl.GetNodesLoc(curState)
		nodes = nodeTbl.GetNodesByLoc(curLoc)
		if nodes == nil {
			return nil, fmt.Errorf("liveness BFS postfix missing queued state %d", curState)
		}
	}
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

func (w *LiveWorker) createCache() {
	if w.Checker.TableauDiskGraph != nil {
		w.Checker.TableauDiskGraph.CreateCache()
		return
	}
	if w.Checker.DiskGraph != nil {
		w.Checker.DiskGraph.CreateCache()
	}
}

func (w *LiveWorker) destroyCache() {
	if w.Checker.TableauDiskGraph != nil {
		w.Checker.TableauDiskGraph.DestroyCache()
		return
	}
	if w.Checker.DiskGraph != nil {
		w.Checker.DiskGraph.DestroyCache()
	}
}

func (w *LiveWorker) getPath(state uint64, tidx int) (*LongVec, error) {
	if w.Checker.TableauDiskGraph != nil {
		return w.Checker.TableauDiskGraph.GetPath(state, tidx)
	}
	if w.Checker.DiskGraph != nil {
		return w.Checker.DiskGraph.GetPath(state, tidx)
	}
	return nil, fmt.Errorf("liveness worker has no disk graph")
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
