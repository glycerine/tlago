package tlc

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	liveCheck1MaxFirst  int64 = 0x2000000000000000
	liveCheck1MaxSecond int64 = 0x5000000000000000
)

type LiveCheck1 struct {
	ErrorFound *atomic.Bool
	MyTool     *Tool
	MetaDir    string
	Actions    []*Action
	Solutions  []*OrderOfSolution
	BGraphs    []*BEGraph
	StateTrace *StateVec

	CurrentOOS *OrderOfSolution
	CurrentPEM *PossibleErrorModel
	ComStack   *MemObjectStack

	FirstNum       int64
	SecondNum      int64
	StartSecondNum int64
	ThirdNum       int64
	StartThirdNum  int64
	NumFirstCom    int64
	NumSecondCom   int64
	InitNode       *BEGraphNode

	mu sync.Mutex
}

func NewLiveCheck1(tool *Tool) (*LiveCheck1, error) {
	return NewLiveCheck1WithError(tool, nil, false)
}

func NewLiveCheck1WithError(tool *Tool, errorFound *atomic.Bool, silent bool) (*LiveCheck1, error) {
	if errorFound == nil {
		errorFound = &atomic.Bool{}
	}
	solutions, err := ProcessLivenessSilent(tool, silent)
	if err != nil {
		return nil, err
	}
	return &LiveCheck1{
		ErrorFound: errorFound,
		MyTool:     tool,
		Solutions:  solutions,
		BGraphs:    []*BEGraph{},
	}, nil
}

func (lc *LiveCheck1) Init(tool *Tool, actions []*Action, metadir string) error {
	if lc == nil {
		return nil
	}
	solutions, err := ProcessLiveness(tool)
	if err != nil {
		return err
	}
	lc.MyTool = tool
	lc.MetaDir = metadir
	lc.Actions = actions
	lc.Solutions = solutions
	lc.BGraphs = make([]*BEGraph, len(solutions))
	for i, solution := range solutions {
		lc.BGraphs[i] = NewBEGraph(metadir, solution.HasTableau())
	}
	return nil
}

func (lc *LiveCheck1) Reset() error {
	if lc == nil {
		return nil
	}
	for _, graph := range lc.BGraphs {
		if graph != nil {
			graph.ResetNumberField()
		}
	}
	return nil
}

func (lc *LiveCheck1) initSccParams(solution *OrderOfSolution) {
	lc.CurrentOOS = solution
	lc.ComStack = NewMemObjectStack(lc.MetaDir, "comstack")
	lc.FirstNum = 1
	lc.SecondNum = liveCheck1MaxFirst + 1
	lc.ThirdNum = liveCheck1MaxSecond + 1
	lc.StartSecondNum = lc.SecondNum
	lc.StartThirdNum = lc.ThirdNum
	lc.NumFirstCom = lc.SecondNum
	lc.NumSecondCom = lc.ThirdNum
}

func (lc *LiveCheck1) ConstructBEGraph(tool *Tool, solution *OrderOfSolution) ([]*BEGraphNode, error) {
	if lc == nil || solution == nil || lc.StateTrace == nil || lc.StateTrace.Size() == 0 {
		return nil, nil
	}
	initNodes := make([]*BEGraphNode, 0, 1)
	slen := len(solution.CheckState)
	alen := len(solution.CheckAction)
	srcState := lc.StateTrace.At(0)
	srcFP := srcState.FingerPrint()
	checkStateRes, err := solution.CheckStateValues(tool, srcState)
	if err != nil {
		return nil, err
	}
	checkActionRes, err := solution.CheckActionValues(tool, srcState, srcState)
	if err != nil {
		return nil, err
	}
	if !solution.HasTableau() {
		allNodes := NewLongObjTable[*BEGraphNode](127)
		srcNode := NewBEGraphNode(srcFP)
		srcNode.SetCheckState(checkStateRes)
		srcNode.AddTransition(srcNode, slen, alen, checkActionRes)
		allNodes.Put(int64(srcFP), srcNode)
		initNodes = append(initNodes, srcNode)
		for i := 1; i < lc.StateTrace.Size(); i++ {
			destState := lc.StateTrace.At(i)
			destFP := destState.FingerPrint()
			destNode, ok := allNodes.Get(int64(destFP))
			if !ok || destNode == nil {
				destNode = NewBEGraphNode(destFP)
				checks, err := solution.CheckStateValues(tool, srcState)
				if err != nil {
					return nil, err
				}
				destNode.SetCheckState(checks)
				selfActions, err := solution.CheckActionValues(tool, destState, destState)
				if err != nil {
					return nil, err
				}
				destNode.AddTransition(destNode, slen, alen, selfActions)
				actions, err := solution.CheckActionValues(tool, srcState, destState)
				if err != nil {
					return nil, err
				}
				srcNode.AddTransition(destNode, slen, alen, actions)
				allNodes.Put(int64(destFP), destNode)
			} else if !srcNode.TransExists(destNode) {
				actions, err := solution.CheckActionValues(tool, srcState, destState)
				if err != nil {
					return nil, err
				}
				srcNode.AddTransition(destNode, slen, alen, actions)
			}
			srcNode = destNode
			srcState = destState
		}
		return initNodes, nil
	}

	allNodes := NewLongObjTable[*BEGraphNode](255)
	srcNodes := make([]*BEGraphNode, 0)
	for i := 0; i < solution.Tableau.InitCnt; i++ {
		tnode := solution.Tableau.GetNode(i)
		ok, err := tnode.IsConsistent(srcState, lc.MyTool)
		if err != nil {
			return nil, err
		}
		if ok {
			destNode := NewBTGraphNode(srcFP, tnode.Index).BEGraphNode
			destNode.SetCheckState(checkStateRes)
			initNodes = append(initNodes, destNode)
			srcNodes = append(srcNodes, destNode)
			allNodes.Put(int64(FP64ExtendInt(srcFP, int32(tnode.Index))), destNode)
		}
	}
	for _, srcNode := range srcNodes {
		tnode := srcNode.GetTNode(solution.Tableau)
		if tnode == nil {
			continue
		}
		for _, tnode1 := range tnode.Nexts {
			destFP := FP64ExtendInt(srcFP, int32(tnode1.Index))
			if destNode, ok := allNodes.Get(int64(destFP)); ok && destNode != nil {
				srcNode.AddTransition(destNode, slen, alen, checkActionRes)
			}
		}
	}
	for i := 1; i < lc.StateTrace.Size(); i++ {
		destNodes := make([]*BEGraphNode, 0)
		destState := lc.StateTrace.At(i)
		destStateFP := destState.FingerPrint()
		checkStateRes, err = solution.CheckStateValues(lc.MyTool, destState)
		if err != nil {
			return nil, err
		}
		checkActionRes, err = solution.CheckActionValues(tool, srcState, destState)
		if err != nil {
			return nil, err
		}
		for _, srcNode := range srcNodes {
			tnode := srcNode.GetTNode(solution.Tableau)
			if tnode == nil {
				continue
			}
			for _, tnode1 := range tnode.Nexts {
				destFP := FP64ExtendInt(destStateFP, int32(tnode1.Index))
				destNode, ok := allNodes.Get(int64(destFP))
				if !ok || destNode == nil {
					consistent, err := tnode1.IsConsistent(destState, lc.MyTool)
					if err != nil {
						return nil, err
					}
					if consistent {
						destNode = NewBTGraphNode(destStateFP, tnode1.Index).BEGraphNode
						destNode.SetCheckState(checkStateRes)
						srcNode.AddTransition(destNode, slen, alen, checkActionRes)
						destNodes = append(destNodes, destNode)
						allNodes.Put(int64(destFP), destNode)
					}
				} else if !srcNode.TransExists(destNode) {
					srcNode.AddTransition(destNode, slen, alen, checkActionRes)
				}
			}
		}
		checkActionRes, err = solution.CheckActionValues(tool, destState, destState)
		if err != nil {
			return nil, err
		}
		for _, srcNode := range destNodes {
			tnode := srcNode.GetTNode(solution.Tableau)
			if tnode == nil {
				continue
			}
			for _, tnode1 := range tnode.Nexts {
				destFP := FP64ExtendInt(destStateFP, int32(tnode1.Index))
				destNode, ok := allNodes.Get(int64(destFP))
				if !ok || destNode == nil {
					consistent, err := tnode1.IsConsistent(destState, lc.MyTool)
					if err != nil {
						return nil, err
					}
					if consistent {
						destNode = NewBTGraphNode(destStateFP, tnode1.Index).BEGraphNode
						destNode.SetCheckState(checkStateRes)
						srcNode.AddTransition(destNode, slen, alen, checkActionRes)
						destNodes = append(destNodes, destNode)
						allNodes.Put(int64(destFP), destNode)
					}
				} else if !srcNode.TransExists(destNode) {
					srcNode.AddTransition(destNode, slen, alen, checkActionRes)
				}
			}
		}
		srcNodes = destNodes
		srcState = destState
	}
	return initNodes, nil
}

func (lc *LiveCheck1) AddInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	if lc == nil {
		return nil
	}
	for soln, solution := range lc.Solutions {
		if soln >= len(lc.BGraphs) {
			continue
		}
		bgraph := lc.BGraphs[soln]
		slen := len(solution.CheckState)
		alen := len(solution.CheckAction)
		checkStateRes, err := solution.CheckStateValues(tool, state)
		if err != nil {
			return err
		}
		checkActionRes, err := solution.CheckActionValues(tool, state, state)
		if err != nil {
			return err
		}
		if !solution.HasTableau() {
			node := NewBEGraphNode(stateFP)
			node.SetCheckState(checkStateRes)
			bgraph.AddInitNode(node)
			node.AddTransition(node, slen, alen, checkActionRes)
			bgraph.AllNodes.PutBENode(node)
		} else {
			for i := 0; i < solution.Tableau.InitCnt; i++ {
				tnode := solution.Tableau.GetNode(i)
				ok, err := tnode.IsConsistent(state, lc.MyTool)
				if err != nil {
					return err
				}
				if ok {
					destNode := NewBTGraphNode(stateFP, tnode.Index)
					destNode.SetCheckState(checkStateRes)
					bgraph.AddInitNode(destNode.BEGraphNode)
					bgraph.AllNodes.PutBTNode(destNode)
					if err := lc.addNodesForStut(state, stateFP, destNode.BEGraphNode, checkStateRes, checkActionRes, solution, bgraph); err != nil {
						return err
					}
				}
			}
		}
		bgraph.AllNodes.SetDone(stateFP)
	}
	return nil
}

func (lc *LiveCheck1) AddNextState(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates) error {
	if lc == nil || nextStates == nil {
		return nil
	}
	lc.mu.Lock()
	defer lc.mu.Unlock()
	for i := 0; i < nextStates.Size(); i++ {
		s2 := nextStates.Next()
		if s2 == nil {
			continue
		}
		if err := lc.addNextStatePair(tool, s0, fp0, s2, s2.FingerPrint()); err != nil {
			return err
		}
	}
	nextStates.ResetNext()
	return nil
}

func (lc *LiveCheck1) addNextStatePair(tool *Tool, s1 *TLCStateMut, fp1 uint64, s2 *TLCStateMut, fp2 uint64) error {
	for soln, solution := range lc.Solutions {
		if soln >= len(lc.BGraphs) {
			continue
		}
		bgraph := lc.BGraphs[soln]
		slen := len(solution.CheckState)
		alen := len(solution.CheckAction)
		if !solution.HasTableau() {
			node1 := bgraph.AllNodes.GetBENode(fp1)
			if node1 == nil {
				continue
			}
			node2 := bgraph.AllNodes.GetBENode(fp2)
			if node2 == nil {
				node2 = NewBEGraphNode(fp2)
				checks, err := solution.CheckStateValues(tool, s2)
				if err != nil {
					return err
				}
				node2.SetCheckState(checks)
				actions, err := solution.CheckActionValues(tool, s1, s2)
				if err != nil {
					return err
				}
				node1.AddTransition(node2, slen, alen, actions)
				selfActions, err := solution.CheckActionValues(tool, s2, s2)
				if err != nil {
					return err
				}
				node2.AddTransition(node2, slen, alen, selfActions)
				bgraph.AllNodes.PutBENode(node2)
			} else if !node1.TransExists(node2) {
				actions, err := solution.CheckActionValues(tool, s1, s2)
				if err != nil {
					return err
				}
				node1.AddTransition(node2, slen, alen, actions)
			}
			continue
		}

		srcNodes := bgraph.AllNodes.GetBTNodes(fp1)
		if srcNodes == nil {
			continue
		}
		var checkStateRes []bool
		checkActionRes, err := solution.CheckActionValues(tool, s1, s2)
		if err != nil {
			return err
		}
		var checkActionRes1 []bool
		for _, srcBTNode := range srcNodes {
			srcNode := srcBTNode.BEGraphNode
			tnode := solution.Tableau.GetNode(srcBTNode.GetIndex())
			for _, tnode1 := range tnode.Nexts {
				destBTNode := bgraph.AllNodes.GetBTNode(fp2, tnode1.Index)
				if destBTNode == nil {
					consistent, err := tnode1.IsConsistent(s2, lc.MyTool)
					if err != nil {
						return err
					}
					if consistent {
						destBTNode = NewBTGraphNode(fp2, tnode1.Index)
						if checkStateRes == nil {
							checkStateRes, err = solution.CheckStateValues(tool, s2)
							if err != nil {
								return err
							}
						}
						destBTNode.SetCheckState(checkStateRes)
						srcNode.AddTransition(destBTNode.BEGraphNode, slen, alen, checkActionRes)
						idx := bgraph.AllNodes.PutBTNode(destBTNode)
						if checkActionRes1 == nil {
							checkActionRes1, err = solution.CheckActionValues(tool, s2, s2)
							if err != nil {
								return err
							}
						}
						if err := lc.addNodesForStut(s2, fp2, destBTNode.BEGraphNode, checkStateRes, checkActionRes1, solution, bgraph); err != nil {
							return err
						}
						if bgraph.AllNodes.IsDone(idx) {
							if err := lc.addNextStateDone(tool, s2, fp2, destBTNode.BEGraphNode, solution, bgraph); err != nil {
								return err
							}
						}
					}
				} else if !srcNode.TransExists(destBTNode.BEGraphNode) {
					srcNode.AddTransition(destBTNode.BEGraphNode, slen, alen, checkActionRes)
				}
			}
		}
	}
	return nil
}

func (lc *LiveCheck1) addNodesForStut(state *TLCStateMut, fp uint64, node *BEGraphNode, checkState []bool, checkAction []bool, solution *OrderOfSolution, bgraph *BEGraph) error {
	slen := len(solution.CheckState)
	alen := len(solution.CheckAction)
	tnode := node.GetTNode(solution.Tableau)
	if tnode == nil {
		return nil
	}
	for _, tnode1 := range tnode.Nexts {
		destBTNode := bgraph.AllNodes.GetBTNode(fp, tnode1.Index)
		if destBTNode == nil {
			consistent, err := tnode1.IsConsistent(state, lc.MyTool)
			if err != nil {
				return err
			}
			if consistent {
				destBTNode = NewBTGraphNode(fp, tnode1.Index)
				destBTNode.SetCheckState(checkState)
				node.AddTransition(destBTNode.BEGraphNode, slen, alen, checkAction)
				bgraph.AllNodes.PutBTNode(destBTNode)
				if err := lc.addNodesForStut(state, fp, destBTNode.BEGraphNode, checkState, checkAction, solution, bgraph); err != nil {
					return err
				}
			}
		} else {
			node.AddTransition(destBTNode.BEGraphNode, slen, alen, checkAction)
		}
	}
	return nil
}

func (lc *LiveCheck1) addNextStateDone(tool *Tool, state *TLCStateMut, fp uint64, node *BEGraphNode, solution *OrderOfSolution, bgraph *BEGraph) error {
	tnode := node.GetTNode(solution.Tableau)
	if tnode == nil {
		return nil
	}
	slen := len(solution.CheckState)
	alen := len(solution.CheckAction)
	var checkStateRes []bool
	var checkActionRes []bool
	for _, action := range lc.Actions {
		nextStates, err := lc.MyTool.GetNextStates(action, state)
		if err != nil {
			return err
		}
		for j := 0; j < nextStates.Size(); j++ {
			s1 := nextStates.At(j)
			fp1 := s1.FingerPrint()
			var checkActionRes1 []bool
			for _, tnode1 := range tnode.Nexts {
				destBTNode := bgraph.AllNodes.GetBTNode(fp1, tnode1.Index)
				if destBTNode == nil {
					consistent, err := tnode1.IsConsistent(s1, lc.MyTool)
					if err != nil {
						return err
					}
					if consistent {
						destBTNode = NewBTGraphNode(fp1, tnode1.Index)
						if checkStateRes == nil {
							checkStateRes, err = solution.CheckStateValues(tool, s1)
							if err != nil {
								return err
							}
						}
						if checkActionRes == nil {
							checkActionRes, err = solution.CheckActionValues(tool, state, s1)
							if err != nil {
								return err
							}
						}
						destBTNode.SetCheckState(checkStateRes)
						node.AddTransition(destBTNode.BEGraphNode, slen, alen, checkActionRes)
						if checkActionRes1 == nil {
							checkActionRes1, err = solution.CheckActionValues(tool, s1, s1)
							if err != nil {
								return err
							}
						}
						if err := lc.addNodesForStut(s1, fp1, destBTNode.BEGraphNode, checkStateRes, checkActionRes1, solution, bgraph); err != nil {
							return err
						}
						idx := bgraph.AllNodes.PutBTNode(destBTNode)
						if bgraph.AllNodes.IsDone(idx) {
							if err := lc.addNextStateDone(tool, s1, fp1, destBTNode.BEGraphNode, solution, bgraph); err != nil {
								return err
							}
						}
					}
				} else if !node.TransExists(destBTNode.BEGraphNode) {
					if checkActionRes == nil {
						checkActionRes, err = solution.CheckActionValues(tool, state, s1)
						if err != nil {
							return err
						}
					}
					node.AddTransition(destBTNode.BEGraphNode, slen, alen, checkActionRes)
				}
			}
		}
	}
	return nil
}

func (lc *LiveCheck1) SetDone(fp uint64) {
	if lc == nil {
		return
	}
	lc.mu.Lock()
	defer lc.mu.Unlock()
	for _, graph := range lc.BGraphs {
		if graph != nil && graph.AllNodes != nil {
			graph.AllNodes.SetDone(fp)
		}
	}
}

func (lc *LiveCheck1) DoLiveCheck() bool {
	return true
}

func (lc *LiveCheck1) Check(tool *Tool, forceCheck bool) (int, error) {
	_ = forceCheck
	if lc == nil {
		return NoError, nil
	}
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if len(lc.Solutions) == 0 {
		return NoError, nil
	}
	for soln, solution := range lc.Solutions {
		if soln >= len(lc.BGraphs) {
			continue
		}
		lc.initSccParams(solution)
		bgraph := lc.BGraphs[soln]
		for i := 0; i < bgraph.InitSize(); i++ {
			lc.InitNode = bgraph.GetInitNode(i)
			if lc.InitNode != nil && lc.InitNode.GetNumber() == 0 {
				if _, err := lc.checkSccs(lc.InitNode); err != nil {
					return ECTLCTemporalPropertyViolated, err
				}
			}
		}
	}
	return NoError, nil
}

func (lc *LiveCheck1) CheckTrace(tool *Tool, trace func() *StateVec) error {
	if lc == nil || trace == nil {
		return nil
	}
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.StateTrace = trace()
	if lc.StateTrace == nil || lc.StateTrace.IsEmpty() {
		return nil
	}
	for _, solution := range lc.Solutions {
		initNodes, err := lc.ConstructBEGraph(tool, solution)
		if err != nil {
			return err
		}
		lc.initSccParams(solution)
		for _, initNode := range initNodes {
			lc.InitNode = initNode
			if initNode != nil && initNode.GetNumber() == 0 {
				if _, err := lc.checkSccs(initNode); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (lc *LiveCheck1) printErrorTrace(node *BEGraphNode) (*CounterExample, error) {
	cycleStack := NewMemObjectStack(lc.MetaDir, "cyclestack")
	slen := len(lc.CurrentOOS.CheckState)
	alen := len(lc.CurrentOOS.CheckAction)
	lowNum := lc.ThirdNum - 1
	aeStateRes := make([]bool, len(lc.CurrentPEM.AEState))
	aeActionRes := make([]bool, len(lc.CurrentPEM.AEAction))
	promiseRes := make([]bool, len(lc.CurrentOOS.Promises))
	cnt := len(aeStateRes) + len(aeActionRes) + len(promiseRes)
	curNode := node
	for cnt > 0 {
		curNode.SetNumber(lc.ThirdNum)
		cnt0 := cnt
	next:
		for {
			for i, idx := range lc.CurrentPEM.AEState {
				if !aeStateRes[i] && curNode.GetCheckState(idx) {
					aeStateRes[i] = true
					cnt--
				}
			}
			for i, promise := range lc.CurrentOOS.Promises {
				tnode := curNode.GetTNode(lc.CurrentOOS.Tableau)
				if !promiseRes[i] && tnode != nil && tnode.Par.IsFulfilling(promise) {
					promiseRes[i] = true
					cnt--
				}
			}
			if cnt <= 0 {
				break
			}
			var nextNode1 *BEGraphNode
			var nextNode2 *BEGraphNode
			cnt1 := cnt
			for i := 0; i < curNode.NextSize(); i++ {
				node1 := curNode.NextAt(i)
				num := node1.GetNumber()
				if lowNum <= num && num <= lc.ThirdNum {
					nextNode1 = node1
					for j, idx := range lc.CurrentPEM.AEAction {
						if !aeActionRes[j] && curNode.GetCheckAction(slen, alen, i, idx) {
							aeActionRes[j] = true
							cnt--
						}
					}
				}
				if cnt < cnt1 {
					cycleStack.Push(curNode)
					curNode = node1
					lc.ThirdNum++
					break next
				}
				if lowNum <= num && num < lc.ThirdNum {
					nextNode2 = node1
				}
			}
			if cnt < cnt0 {
				cycleStack.Push(curNode)
				curNode = nextNode1
				lc.ThirdNum++
				break
			}
			for nextNode2 == nil {
				popped, _ := cycleStack.Pop().(*BEGraphNode)
				if popped == nil {
					return nil, newTLCError(ECTLCLiveBEGraphFailedToConstruct, "failed to backtrack liveness cycle")
				}
				curNode = popped
				for i := 0; i < curNode.NextSize(); i++ {
					node1 := curNode.NextAt(i)
					num := node1.GetNumber()
					if lowNum <= num && num < lc.ThirdNum {
						cycleStack.Push(curNode)
						nextNode2 = node1
						break
					}
				}
			}
			cycleStack.Push(curNode)
			curNode = nextNode2
			curNode.SetNumber(lc.ThirdNum)
		}
	}

	curNode.SetNumber(lc.ThirdNum + 1)
	lc.ThirdNum++
done:
	for curNode != node {
		found := false
		for i := 0; i < curNode.NextSize(); i++ {
			nextNode := curNode.NextAt(i)
			num := nextNode.GetNumber()
			if lowNum <= num && num < lc.ThirdNum {
				cycleStack.Push(curNode)
				if nextNode == node {
					break done
				}
				found = true
				curNode = nextNode
				curNode.SetNumber(lc.ThirdNum)
				break
			}
		}
		if !found {
			popped, _ := cycleStack.Pop().(*BEGraphNode)
			if popped == nil {
				return nil, newTLCError(ECTLCLiveBEGraphFailedToConstruct, "failed to close liveness cycle")
			}
			curNode = popped
		}
	}
	if cycleStack.Size() == 0 {
		cycleStack.Push(curNode)
	}

	stateNum := 0
	prefix, err := BEGraphGetPath(lc.InitNode, node)
	if err != nil {
		return nil, err
	}
	states := make([]*TLCStateInfo, len(prefix))
	trace := make([]*TLCStateInfo, 0, len(states))

	fp := prefix[0].StateFP
	sinfo, err := lc.recoverTraceState(fp, nil, stateNum)
	if err != nil {
		return nil, err
	}
	states[stateNum] = sinfo
	stateNum++

	for i := 1; i < len(states); i++ {
		fp1 := prefix[i].StateFP
		if fp1 != fp {
			sinfo, err = lc.recoverTraceState(fp1, sinfo.State, stateNum)
			if err != nil {
				return nil, err
			}
			states[stateNum] = sinfo
			stateNum++
		}
		fp = fp1
	}

	var cycleState *TLCStateMut
	for i := 0; i < stateNum; i++ {
		alias, err := lc.MyTool.EvalAliasInfoPair(states[i], cycleState)
		if err != nil {
			return nil, err
		}
		trace = append(trace, alias)
		cycleState = states[i].State
	}

	lastState := cycleState
	cyclePos := stateNum
	fps := make([]uint64, cycleStack.Size())
	for idx := len(fps) - 1; idx >= 0; idx-- {
		popped, _ := cycleStack.Pop().(*BEGraphNode)
		if popped != nil {
			fps[idx] = popped.StateFP
		}
	}
	sinfo = states[stateNum-1]
	for i := 1; i < len(fps); i++ {
		if fps[i] != fps[i-1] {
			sinfo, err = lc.recoverTraceState(fps[i], sinfo.State, stateNum)
			if err != nil {
				return nil, err
			}
			stateNum++
			sinfo, err = lc.MyTool.EvalAliasInfoPair(sinfo, lastState)
			if err != nil {
				return nil, err
			}
			trace = append(trace, sinfo)
			lastState = sinfo.State
		}
	}

	violated := LivenessFindViolatedPropertiesFromTrace(lc.MyTool, trace, cyclePos-1)
	PrintError(ECTLCTemporalPropertyViolated, violated...)
	PrintError(ECTLCCounterExample)
	for i, info := range trace {
		var prev *TLCStateMut
		if i > 0 && trace[i-1] != nil {
			prev = trace[i-1].State
		}
		PrintInvariantViolationStateTraceState(info, prev, i+1)
	}

	if lastState != nil && node.StateFP == lastState.FingerPrint() {
		PrintStutteringState(stateNum)
		return NewCounterExample(trace, UnknownAction, stateNum, true), nil
	}
	sinfo, err = lc.recoverTraceState(cycleState.FingerPrint(), sinfo, stateNum)
	if err != nil {
		return nil, err
	}
	PrintBackToState(sinfo, cyclePos)
	return NewCounterExample(trace, sinfo.Action(), cyclePos, true), nil
}

func (lc *LiveCheck1) recoverTraceState(fp uint64, predecessor any, stateNum int) (*TLCStateInfo, error) {
	if lc.StateTrace != nil && !lc.StateTrace.IsEmpty() {
		if stateNum < 0 || stateNum >= lc.StateTrace.Size() {
			return nil, newTLCError(ECTLCFailedToRecoverNext, "state %d is outside the simulation trace", stateNum)
		}
		state := lc.StateTrace.At(stateNum)
		if state == nil || state.FingerPrint() != fp {
			return nil, newTLCError(ECTLCFailedToRecoverNext, "trace state %d does not match fingerprint %d", stateNum, fp)
		}
		return NewTLCStateInfo(state), nil
	}
	if predecessor == nil {
		return lc.MyTool.GetState(fp)
	}
	return lc.MyTool.GetState(fp, predecessor)
}

func (lc *LiveCheck1) checkSubcomponent(node *BEGraphNode) error {
	slen := len(lc.CurrentOOS.CheckState)
	alen := len(lc.CurrentOOS.CheckAction)
	aeStateRes := make([]bool, len(lc.CurrentPEM.AEState))
	aeActionRes := make([]bool, len(lc.CurrentPEM.AEAction))
	promiseRes := make([]bool, len(lc.CurrentOOS.Promises))
	stack := NewMemObjectStack(lc.MetaDir, "subcomstack")

	node.IncNumber()
	stack.Push(node)
	for stack.Size() != 0 {
		curNode, _ := stack.Pop().(*BEGraphNode)
		if curNode == nil {
			continue
		}
		for i, idx := range lc.CurrentPEM.AEState {
			if !aeStateRes[i] {
				aeStateRes[i] = curNode.GetCheckState(idx)
			}
		}
		for i := 0; i < curNode.NextSize(); i++ {
			node1 := curNode.NextAt(i)
			num := node1.GetNumber()
			if num >= lc.ThirdNum {
				for j, idx := range lc.CurrentPEM.AEAction {
					if !aeActionRes[j] {
						aeActionRes[j] = curNode.GetCheckAction(slen, alen, i, idx)
					}
				}
			}
			if num == lc.ThirdNum {
				node1.IncNumber()
				stack.Push(node1)
			}
		}
		for i, promise := range lc.CurrentOOS.Promises {
			tnode := curNode.GetTNode(lc.CurrentOOS.Tableau)
			if tnode != nil && tnode.Par.IsFulfilling(promise) {
				promiseRes[i] = true
			}
		}
	}
	lc.ThirdNum += 2

	for _, ok := range aeStateRes {
		if !ok {
			return nil
		}
	}
	for _, ok := range aeActionRes {
		if !ok {
			return nil
		}
	}
	for _, ok := range promiseRes {
		if !ok {
			return nil
		}
	}
	if lc.ErrorFound == nil || lc.ErrorFound.CompareAndSwap(false, true) {
		counterExample, err := lc.printErrorTrace(node)
		if err != nil {
			PrintError(ECGeneral, javaGeneralErrorMessage("printing an error trace", err))
			return NewLiveException(ECTLCTemporalPropertyViolated, "LiveCheck: Found error trace.")
		}
		return NewLiveCounterExampleException(ECTLCTemporalPropertyViolated, "LiveCheck: Found error trace.", counterExample)
	}
	return nil
}

func (lc *LiveCheck1) checkSccs(node *BEGraphNode) (int64, error) {
	lowlink := lc.FirstNum
	lc.FirstNum++
	node.SetNumber(lowlink)
	lc.ComStack.Push(node)
	for i := 0; i < node.NextSize(); i++ {
		destNode := node.NextAt(i)
		destNum := destNode.GetNumber()
		if destNum == 0 {
			var err error
			destNum, err = lc.checkSccs(destNode)
			if err != nil {
				return lowlink, err
			}
		}
		if destNum < lowlink {
			lowlink = destNum
		}
	}
	if lowlink == node.GetNumber() {
		if err := lc.checkComponent(node); err != nil {
			return lowlink, err
		}
	}
	return lowlink, nil
}

func (lc *LiveCheck1) checkComponent(node *BEGraphNode) error {
	nodes := lc.extractComponent(node)
	if nodes == nil {
		return nil
	}
	for _, pem := range lc.CurrentOOS.PEMs {
		lc.CurrentPEM = pem
		lc.StartSecondNum = lc.SecondNum
		lc.StartThirdNum = lc.ThirdNum
		for j := len(nodes) - 1; j >= 0; j-- {
			node1 := nodes[j]
			if node1.GetNumber() < lc.StartThirdNum {
				if _, err := lc.checkSccs1(node1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (lc *LiveCheck1) extractComponent(node *BEGraphNode) []*BEGraphNode {
	node1, _ := lc.ComStack.Pop().(*BEGraphNode)
	if node == node1 && !node.TransExists(node) {
		node.SetNumber(liveCheck1MaxFirst)
		return nil
	}
	nodes := make([]*BEGraphNode, 0)
	lc.NumFirstCom = lc.SecondNum
	lc.SecondNum++
	lc.NumSecondCom = lc.ThirdNum
	node1.SetNumber(lc.NumFirstCom)
	nodes = append(nodes, node1)
	for node != node1 {
		node1, _ = lc.ComStack.Pop().(*BEGraphNode)
		node1.SetNumber(lc.NumFirstCom)
		nodes = append(nodes, node1)
	}
	return nodes
}

func (lc *LiveCheck1) checkSccs1(node *BEGraphNode) (int64, error) {
	lowlink := lc.SecondNum
	lc.SecondNum++
	node.SetNumber(lowlink)
	lc.ComStack.Push(node)
	for i := 0; i < node.NextSize(); i++ {
		destNode := node.NextAt(i)
		destNum := destNode.GetNumber()
		if lc.NumFirstCom <= destNum && node.GetCheckActionAll(len(lc.CurrentOOS.CheckState), len(lc.CurrentOOS.CheckAction), i, lc.CurrentPEM.EAAction) {
			if destNum < lc.StartSecondNum || (lc.NumSecondCom <= destNum && destNum < lc.StartThirdNum) {
				var err error
				destNum, err = lc.checkSccs1(destNode)
				if err != nil {
					return lowlink, err
				}
			}
			if destNum < lowlink {
				lowlink = destNum
			}
		}
	}
	if lowlink == node.GetNumber() {
		if lc.extractComponent1(node) {
			if err := lc.checkSubcomponent(node); err != nil {
				return lowlink, err
			}
		}
	}
	return lowlink, nil
}

func (lc *LiveCheck1) extractComponent1(node *BEGraphNode) bool {
	node1, _ := lc.ComStack.Pop().(*BEGraphNode)
	if node == node1 && !lc.canStutter(node) {
		node.SetNumber(lc.ThirdNum)
		lc.ThirdNum++
		return false
	}
	node1.SetNumber(lc.ThirdNum)
	for node != node1 {
		node1, _ = lc.ComStack.Pop().(*BEGraphNode)
		node1.SetNumber(lc.ThirdNum)
	}
	return true
}

func (lc *LiveCheck1) canStutter(node *BEGraphNode) bool {
	slen := len(lc.CurrentOOS.CheckState)
	alen := len(lc.CurrentOOS.CheckAction)
	for i := 0; i < node.NextSize(); i++ {
		node1 := node.NextAt(i)
		if beGraphNodesEqual(node, node1) {
			nodeNum := node.GetNumber()
			return lc.NumFirstCom <= nodeNum && node.GetCheckActionAll(slen, alen, i, lc.CurrentPEM.EAAction)
		}
	}
	return false
}

func (lc *LiveCheck1) FinalCheck(tool *Tool) (int, error) {
	return lc.Check(tool, true)
}

func (lc *LiveCheck1) GetMetaDir() string {
	if lc == nil {
		return ""
	}
	return lc.MetaDir
}

func (lc *LiveCheck1) GetTool() *Tool {
	if lc == nil {
		return nil
	}
	return lc.MyTool
}

func (lc *LiveCheck1) GetOutDegreeStatistics() *BucketStatistics {
	return NewBucketStatistics("Histogram vertex out-degree")
}

func (lc *LiveCheck1) NumChecker() int { return 0 }

func (lc *LiveCheck1) Close() error                  { return nil }
func (lc *LiveCheck1) BeginChkpt() error             { return nil }
func (lc *LiveCheck1) CommitChkpt() error            { return nil }
func (lc *LiveCheck1) Recover() error                { return nil }
func (lc *LiveCheck1) FlushWritesToDiskFiles() error { return nil }

func (lc *LiveCheck1) CalculateInDegreeDiskGraphs(stats *BucketStatistics) (*BucketStatistics, error) {
	if stats == nil {
		stats = NewBucketStatistics("Histogram vertex in-degree")
	}
	return stats, nil
}

func (lc *LiveCheck1) CalculateOutDegreeDiskGraphs(stats *BucketStatistics) (*BucketStatistics, error) {
	if stats == nil {
		stats = NewBucketStatistics("Histogram vertex out-degree")
	}
	return stats, nil
}

func (lc *LiveCheck1) String() string {
	if lc == nil {
		return ""
	}
	return fmt.Sprintf("LiveCheck1[%d]", len(lc.Solutions))
}
