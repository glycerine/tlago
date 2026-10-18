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

type liveCheck1SCCNumbers struct {
	firstNum       int64
	secondNum      int64
	startSecondNum int64
	thirdNum       int64
	startThirdNum  int64
	numFirstCom    int64
	numSecondCom   int64
}

func initialLiveCheck1SCCNumbers() liveCheck1SCCNumbers {
	return liveCheck1SCCNumbers{
		firstNum:       1,
		secondNum:      liveCheck1MaxFirst + 1,
		startSecondNum: liveCheck1MaxFirst + 1,
		thirdNum:       liveCheck1MaxSecond + 1,
		startThirdNum:  liveCheck1MaxSecond + 1,
		numFirstCom:    liveCheck1MaxFirst + 1,
		numSecondCom:   liveCheck1MaxSecond + 1,
	}
}

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

	ownersCaptured  bool
	errorFoundOwner *atomic.Bool
	toolOwner       *Tool
	metaDirOwner    string
	actionsOwner    []*Action
	solutionsOwner  []*OrderOfSolution
	bgraphsOwner    []*BEGraph
	stateTraceOwner *StateVec
	currentOOSOwner *OrderOfSolution
	currentPEMOwner *PossibleErrorModel
	comStackOwner   *MemObjectStack
	initNodeOwner   *BEGraphNode
	sccNumbersOwner liveCheck1SCCNumbers
	mu              sync.Mutex
}

func NewLiveCheck1(tool *Tool) (*LiveCheck1, error) {
	return NewLiveCheck1WithError(tool, &atomic.Bool{}, false)
}

func NewLiveCheck1WithError(tool *Tool, errorFound *atomic.Bool, silent bool) (*LiveCheck1, error) {
	solutions, err := ProcessLivenessSilent(tool, silent)
	if err != nil {
		return nil, err
	}
	numbers := initialLiveCheck1SCCNumbers()
	return &LiveCheck1{
		ErrorFound:      errorFound,
		MyTool:          tool,
		Solutions:       solutions,
		BGraphs:         []*BEGraph{},
		ownersCaptured:  true,
		errorFoundOwner: errorFound,
		toolOwner:       tool,
		solutionsOwner:  append([]*OrderOfSolution(nil), solutions...),
		bgraphsOwner:    []*BEGraph{},
		FirstNum:        numbers.firstNum,
		SecondNum:       numbers.secondNum,
		StartSecondNum:  numbers.startSecondNum,
		ThirdNum:        numbers.thirdNum,
		StartThirdNum:   numbers.startThirdNum,
		NumFirstCom:     numbers.numFirstCom,
		NumSecondCom:    numbers.numSecondCom,
		sccNumbersOwner: numbers,
	}, nil
}

func (lc *LiveCheck1) Init(tool *Tool, actions []*Action, metadir string) error {
	lc.require()
	solutions, err := ProcessLiveness(tool)
	if err != nil {
		return err
	}
	lc.MyTool = tool
	lc.MetaDir = metadir
	if lc.ownersCaptured {
		lc.toolOwner = tool
		lc.metaDirOwner = metadir
	}
	lc.Actions = actions
	lc.Solutions = solutions
	lc.BGraphs = make([]*BEGraph, len(solutions))
	if lc.ownersCaptured {
		lc.actionsOwner = actions
		lc.solutionsOwner = append([]*OrderOfSolution(nil), solutions...)
		lc.bgraphsOwner = lc.BGraphs
	}
	for i, solution := range solutions {
		if solution == nil {
			panic(NewNullPointerException())
		}
		lc.BGraphs[i] = NewBEGraph(metadir, solution.HasTableau())
	}
	if lc.ownersCaptured {
		lc.bgraphsOwner = append([]*BEGraph(nil), lc.BGraphs...)
	}
	return nil
}

func (lc *LiveCheck1) Reset() error {
	lc.require()
	for i := range lc.ownedBGraphs() {
		lc.bgraphAt(i).ResetNumberField()
	}
	return nil
}

func (lc *LiveCheck1) initSccParams(solution *OrderOfSolution) {
	lc.setCurrentOOS(solution)
	lc.setComStack(NewMemObjectStack(lc.ownedMetaDir(), "comstack"))
	lc.setSCCNumbers(initialLiveCheck1SCCNumbers())
}

func (lc *LiveCheck1) ConstructBEGraph(tool *Tool, solution *OrderOfSolution) ([]*BEGraphNode, error) {
	lc.require()
	if solution == nil {
		panic(NewNullPointerException())
	}
	initNodes := make([]*BEGraphNode, 0, 1)
	slen := len(solution.GetCheckState())
	alen := len(solution.GetCheckAction())
	stateTrace := lc.ownedStateTrace()
	srcState := stateTrace.At(0)
	srcFP := srcState.FingerPrint()
	checkStateRes, err := solution.CheckState(tool, srcState)
	if err != nil {
		return nil, err
	}
	checkActionRes, err := solution.CheckAction(tool, srcState, srcState)
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
		for i := 1; i < stateTrace.Size(); i++ {
			destState := stateTrace.At(i)
			destFP := destState.FingerPrint()
			destNode, ok := allNodes.Get(int64(destFP))
			if !ok || destNode == nil {
				destNode = NewBEGraphNode(destFP)
				checks, err := solution.CheckState(tool, srcState)
				if err != nil {
					return nil, err
				}
				destNode.SetCheckState(checks)
				selfActions, err := solution.CheckAction(tool, destState, destState)
				if err != nil {
					return nil, err
				}
				destNode.AddTransition(destNode, slen, alen, selfActions)
				actions, err := solution.CheckAction(tool, srcState, destState)
				if err != nil {
					return nil, err
				}
				srcNode.AddTransition(destNode, slen, alen, actions)
				allNodes.Put(int64(destFP), destNode)
			} else if !srcNode.TransExists(destNode) {
				actions, err := solution.CheckAction(tool, srcState, destState)
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
	tableau := solution.GetTableau()
	for i := 0; i < tableau.GetInitCnt(); i++ {
		tnode := tableau.GetNode(i)
		ok, err := tnode.IsConsistent(srcState, lc.ownedTool())
		if err != nil {
			return nil, err
		}
		if ok {
			destNode := NewBTGraphNode(srcFP, tnode.GetIndex()).BEGraphNode
			destNode.SetCheckState(checkStateRes)
			initNodes = append(initNodes, destNode)
			srcNodes = append(srcNodes, destNode)
			allNodes.Put(int64(FP64ExtendInt(srcFP, int32(tnode.GetIndex()))), destNode)
		}
	}
	for _, srcNode := range srcNodes {
		tnode := srcNode.GetTNode(tableau)
		if tnode == nil {
			panic(NewNullPointerException())
		}
		for _, tnode1 := range tnode.ownedNexts() {
			destFP := FP64ExtendInt(srcFP, int32(tnode1.GetIndex()))
			if destNode, ok := allNodes.Get(int64(destFP)); ok && destNode != nil {
				srcNode.AddTransition(destNode, slen, alen, checkActionRes)
			}
		}
	}
	for i := 1; i < stateTrace.Size(); i++ {
		destNodes := make([]*BEGraphNode, 0)
		destState := stateTrace.At(i)
		destStateFP := destState.FingerPrint()
		checkStateRes, err = solution.CheckState(lc.ownedTool(), destState)
		if err != nil {
			return nil, err
		}
		checkActionRes, err = solution.CheckAction(tool, srcState, destState)
		if err != nil {
			return nil, err
		}
		for _, srcNode := range srcNodes {
			tnode := srcNode.GetTNode(tableau)
			if tnode == nil {
				panic(NewNullPointerException())
			}
			for _, tnode1 := range tnode.ownedNexts() {
				destFP := FP64ExtendInt(destStateFP, int32(tnode1.GetIndex()))
				destNode, ok := allNodes.Get(int64(destFP))
				if !ok || destNode == nil {
					consistent, err := tnode1.IsConsistent(destState, lc.ownedTool())
					if err != nil {
						return nil, err
					}
					if consistent {
						destNode = NewBTGraphNode(destStateFP, tnode1.GetIndex()).BEGraphNode
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
		checkActionRes, err = solution.CheckAction(tool, destState, destState)
		if err != nil {
			return nil, err
		}
		for _, srcNode := range destNodes {
			tnode := srcNode.GetTNode(tableau)
			if tnode == nil {
				panic(NewNullPointerException())
			}
			for _, tnode1 := range tnode.ownedNexts() {
				destFP := FP64ExtendInt(destStateFP, int32(tnode1.GetIndex()))
				destNode, ok := allNodes.Get(int64(destFP))
				if !ok || destNode == nil {
					consistent, err := tnode1.IsConsistent(destState, lc.ownedTool())
					if err != nil {
						return nil, err
					}
					if consistent {
						destNode = NewBTGraphNode(destStateFP, tnode1.GetIndex()).BEGraphNode
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
	lc.require()
	for soln, solution := range lc.ownedSolutions() {
		bgraph := lc.bgraphAt(soln)
		if solution == nil {
			panic(NewNullPointerException())
		}
		slen := len(solution.GetCheckState())
		alen := len(solution.GetCheckAction())
		checkStateRes, err := solution.CheckState(tool, state)
		if err != nil {
			return err
		}
		checkActionRes, err := solution.CheckAction(tool, state, state)
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
			tableau := solution.GetTableau()
			for i := 0; i < tableau.GetInitCnt(); i++ {
				tnode := tableau.GetNode(i)
				ok, err := tnode.IsConsistent(state, lc.ownedTool())
				if err != nil {
					return err
				}
				if ok {
					destNode := NewBTGraphNode(stateFP, tnode.GetIndex())
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
	lc.require()
	if nextStates == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < nextStates.Size(); i++ {
		s2 := nextStates.Next().(*TLCStateMut)
		if err := lc.AddNextStatePair(tool, s0, fp0, s2, s2.FingerPrint()); err != nil {
			return err
		}
	}
	nextStates.ResetNext()
	return nil
}

func (lc *LiveCheck1) AddNextStatePair(tool *Tool, s1 *TLCStateMut, fp1 uint64, s2 *TLCStateMut, fp2 uint64) error {
	lc.require()
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.addNextStatePair(tool, s1, fp1, s2, fp2)
}

func (lc *LiveCheck1) addNextStatePair(tool *Tool, s1 *TLCStateMut, fp1 uint64, s2 *TLCStateMut, fp2 uint64) error {
	for soln, solution := range lc.ownedSolutions() {
		bgraph := lc.bgraphAt(soln)
		if solution == nil {
			panic(NewNullPointerException())
		}
		slen := len(solution.GetCheckState())
		alen := len(solution.GetCheckAction())
		if !solution.HasTableau() {
			node1 := bgraph.AllNodes.GetBENode(fp1)
			if node1 == nil {
				continue
			}
			node2 := bgraph.AllNodes.GetBENode(fp2)
			if node2 == nil {
				node2 = NewBEGraphNode(fp2)
				checks, err := solution.CheckState(tool, s2)
				if err != nil {
					return err
				}
				node2.SetCheckState(checks)
				actions, err := solution.CheckAction(tool, s1, s2)
				if err != nil {
					return err
				}
				node1.AddTransition(node2, slen, alen, actions)
				selfActions, err := solution.CheckAction(tool, s2, s2)
				if err != nil {
					return err
				}
				node2.AddTransition(node2, slen, alen, selfActions)
				bgraph.AllNodes.PutBENode(node2)
			} else if !node1.TransExists(node2) {
				actions, err := solution.CheckAction(tool, s1, s2)
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
		checkActionRes, err := solution.CheckAction(tool, s1, s2)
		if err != nil {
			return err
		}
		var checkActionRes1 []bool
		for _, srcBTNode := range srcNodes {
			srcNode := srcBTNode.BEGraphNode
			tnode := solution.GetTableau().GetNode(srcBTNode.GetIndex())
			for _, tnode1 := range tnode.ownedNexts() {
				destBTNode := bgraph.AllNodes.GetBTNode(fp2, tnode1.GetIndex())
				if destBTNode == nil {
					consistent, err := tnode1.IsConsistent(s2, lc.ownedTool())
					if err != nil {
						return err
					}
					if consistent {
						destBTNode = NewBTGraphNode(fp2, tnode1.GetIndex())
						if checkStateRes == nil {
							checkStateRes, err = solution.CheckState(tool, s2)
							if err != nil {
								return err
							}
						}
						destBTNode.SetCheckState(checkStateRes)
						srcNode.AddTransition(destBTNode.BEGraphNode, slen, alen, checkActionRes)
						idx := bgraph.AllNodes.PutBTNode(destBTNode)
						if checkActionRes1 == nil {
							checkActionRes1, err = solution.CheckAction(tool, s2, s2)
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
	slen := len(solution.GetCheckState())
	alen := len(solution.GetCheckAction())
	tnode := node.GetTNode(solution.GetTableau())
	if tnode == nil {
		panic(NewNullPointerException())
	}
	for _, tnode1 := range tnode.ownedNexts() {
		destBTNode := bgraph.AllNodes.GetBTNode(fp, tnode1.GetIndex())
		if destBTNode == nil {
			consistent, err := tnode1.IsConsistent(state, lc.ownedTool())
			if err != nil {
				return err
			}
			if consistent {
				destBTNode = NewBTGraphNode(fp, tnode1.GetIndex())
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
	tnode := node.GetTNode(solution.GetTableau())
	if tnode == nil {
		panic(NewNullPointerException())
	}
	slen := len(solution.GetCheckState())
	alen := len(solution.GetCheckAction())
	var checkStateRes []bool
	var checkActionRes []bool
	for _, action := range lc.ownedActions() {
		nextStates, err := lc.ownedTool().GetNextStates(action, state)
		if err != nil {
			return err
		}
		for j := 0; j < nextStates.Size(); j++ {
			s1 := nextStates.At(j)
			fp1 := s1.FingerPrint()
			var checkActionRes1 []bool
			for _, tnode1 := range tnode.ownedNexts() {
				destBTNode := bgraph.AllNodes.GetBTNode(fp1, tnode1.GetIndex())
				if destBTNode == nil {
					consistent, err := tnode1.IsConsistent(s1, lc.ownedTool())
					if err != nil {
						return err
					}
					if consistent {
						destBTNode = NewBTGraphNode(fp1, tnode1.GetIndex())
						if checkStateRes == nil {
							checkStateRes, err = solution.CheckState(tool, s1)
							if err != nil {
								return err
							}
						}
						if checkActionRes == nil {
							checkActionRes, err = solution.CheckAction(tool, state, s1)
							if err != nil {
								return err
							}
						}
						destBTNode.SetCheckState(checkStateRes)
						node.AddTransition(destBTNode.BEGraphNode, slen, alen, checkActionRes)
						if checkActionRes1 == nil {
							checkActionRes1, err = solution.CheckAction(tool, s1, s1)
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
						checkActionRes, err = solution.CheckAction(tool, state, s1)
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
	lc.require()
	lc.mu.Lock()
	defer lc.mu.Unlock()
	for soln := range lc.ownedSolutions() {
		graph := lc.bgraphAt(soln)
		if graph.AllNodes == nil {
			panic(NewNullPointerException())
		}
		graph.AllNodes.SetDone(fp)
	}
}

func (lc *LiveCheck1) DoLiveCheck() bool {
	lc.require()
	return true
}

func (lc *LiveCheck1) Check(tool *Tool, forceCheck bool) (int, error) {
	_ = forceCheck
	lc.require()
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if len(lc.ownedSolutions()) == 0 {
		return NoError, nil
	}
	for soln, solution := range lc.ownedSolutions() {
		lc.initSccParams(solution)
		bgraph := lc.bgraphAt(soln)
		for i := 0; i < bgraph.InitSize(); i++ {
			lc.setInitNode(bgraph.GetInitNode(i))
			if lc.ownedInitNode().GetNumber() == 0 {
				if _, err := lc.checkSccs(lc.ownedInitNode()); err != nil {
					return ECTLCTemporalPropertyViolated, err
				}
			}
		}
	}
	return NoError, nil
}

func (lc *LiveCheck1) CheckTrace(tool *Tool, trace func() *StateVec) error {
	lc.require()
	if trace == nil {
		panic(NewNullPointerException())
	}
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.StateTrace = trace()
	if lc.ownersCaptured {
		lc.stateTraceOwner = lc.StateTrace
	}
	for _, solution := range lc.ownedSolutions() {
		initNodes, err := lc.ConstructBEGraph(tool, solution)
		if err != nil {
			return err
		}
		lc.initSccParams(solution)
		for _, initNode := range initNodes {
			lc.setInitNode(initNode)
			if initNode.GetNumber() == 0 {
				if _, err := lc.checkSccs(initNode); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (lc *LiveCheck1) printErrorTrace(node *BEGraphNode) (*CounterExample, error) {
	cycleStack := NewMemObjectStack(lc.ownedMetaDir(), "cyclestack")
	slen := len(lc.ownedCurrentOOS().GetCheckState())
	alen := len(lc.ownedCurrentOOS().GetCheckAction())
	lowNum := lc.ownedSCCNumbers().thirdNum - 1
	aeStateRes := make([]bool, len(lc.ownedCurrentPEM().ownedAEState()))
	aeActionRes := make([]bool, len(lc.ownedCurrentPEM().ownedAEAction()))
	promiseRes := make([]bool, len(lc.ownedCurrentOOS().GetPromises()))
	cnt := len(aeStateRes) + len(aeActionRes) + len(promiseRes)
	curNode := node
	for cnt > 0 {
		curNode.SetNumber(lc.ownedSCCNumbers().thirdNum)
		cnt0 := cnt
	next:
		for {
			for i, idx := range lc.ownedCurrentPEM().ownedAEState() {
				if !aeStateRes[i] && curNode.GetCheckState(idx) {
					aeStateRes[i] = true
					cnt--
				}
			}
			for i, promise := range lc.ownedCurrentOOS().GetPromises() {
				tnode := curNode.GetTNode(lc.ownedCurrentOOS().GetTableau())
				if tnode == nil {
					panic(NewNullPointerException())
				}
				if !promiseRes[i] && tnode.GetPar().IsFulfilling(promise) {
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
				if lowNum <= num && num <= lc.ownedSCCNumbers().thirdNum {
					nextNode1 = node1
					for j, idx := range lc.ownedCurrentPEM().ownedAEAction() {
						if !aeActionRes[j] && curNode.GetCheckAction(slen, alen, i, idx) {
							aeActionRes[j] = true
							cnt--
						}
					}
				}
				if cnt < cnt1 {
					cycleStack.Push(curNode)
					curNode = node1
					lc.addThirdNum(1)
					break next
				}
				if lowNum <= num && num < lc.ownedSCCNumbers().thirdNum {
					nextNode2 = node1
				}
			}
			if cnt < cnt0 {
				cycleStack.Push(curNode)
				curNode = nextNode1
				lc.addThirdNum(1)
				break
			}
			for nextNode2 == nil {
				popped, _ := cycleStack.Pop().(*BEGraphNode)
				if popped == nil {
					panic(NewNullPointerException())
				}
				curNode = popped
				for i := 0; i < curNode.NextSize(); i++ {
					node1 := curNode.NextAt(i)
					num := node1.GetNumber()
					if lowNum <= num && num < lc.ownedSCCNumbers().thirdNum {
						cycleStack.Push(curNode)
						nextNode2 = node1
						break
					}
				}
			}
			cycleStack.Push(curNode)
			curNode = nextNode2
			curNode.SetNumber(lc.ownedSCCNumbers().thirdNum)
		}
	}

	curNode.SetNumber(lc.addThirdNum(1))
done:
	for curNode != node {
		found := false
		for i := 0; i < curNode.NextSize(); i++ {
			nextNode := curNode.NextAt(i)
			num := nextNode.GetNumber()
			if lowNum <= num && num < lc.ownedSCCNumbers().thirdNum {
				cycleStack.Push(curNode)
				if nextNode == node {
					break done
				}
				found = true
				curNode = nextNode
				curNode.SetNumber(lc.ownedSCCNumbers().thirdNum)
				break
			}
		}
		if !found {
			popped, _ := cycleStack.Pop().(*BEGraphNode)
			if popped == nil {
				panic(NewNullPointerException())
			}
			curNode = popped
		}
	}
	if cycleStack.Size() == 0 {
		cycleStack.Push(curNode)
	}

	stateNum := 0
	prefix, err := BEGraphGetPath(lc.ownedInitNode(), node)
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
			sinfo, err = lc.recoverTraceState(fp1, sinfo.ownedState(), stateNum)
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
		alias, err := lc.ownedTool().EvalAliasInfoPair(states[i], cycleState)
		if err != nil {
			return nil, err
		}
		trace = append(trace, alias)
		cycleState = states[i].ownedState()
	}

	lastState := cycleState
	cyclePos := stateNum
	fps := make([]uint64, cycleStack.Size())
	for idx := len(fps) - 1; idx >= 0; idx-- {
		popped, _ := cycleStack.Pop().(*BEGraphNode)
		if popped == nil {
			panic(NewNullPointerException())
		}
		fps[idx] = popped.StateFP
	}
	sinfo = states[stateNum-1]
	for i := 1; i < len(fps); i++ {
		if fps[i] != fps[i-1] {
			sinfo, err = lc.recoverTraceState(fps[i], sinfo.ownedState(), stateNum)
			if err != nil {
				return nil, err
			}
			stateNum++
			sinfo, err = lc.ownedTool().EvalAliasInfoPair(sinfo, lastState)
			if err != nil {
				return nil, err
			}
			trace = append(trace, sinfo)
			lastState = sinfo.ownedState()
		}
	}

	violated := LivenessFindViolatedPropertiesFromTrace(lc.ownedTool(), trace, cyclePos-1)
	PrintError(ECTLCTemporalPropertyViolated, violated...)
	PrintError(ECTLCCounterExample)
	for i, info := range trace {
		var prev *TLCStateMut
		if i > 0 && trace[i-1] != nil {
			prev = trace[i-1].ownedState()
		}
		PrintInvariantViolationStateTraceState(info, prev, i+1)
	}

	if lastState != nil && node.StateFP == lastState.FingerPrint() {
		PrintStutteringState(stateNum)
		return NewCounterExampleFromTraceLoop(trace, stateNum), nil
	}
	sinfo, err = lc.recoverTraceState(cycleState.FingerPrint(), sinfo, stateNum)
	if err != nil {
		return nil, err
	}
	PrintBackToState(sinfo, cyclePos)
	return NewCounterExampleFromTraceActionLoop(trace, sinfo.GetAction(), cyclePos), nil
}

func (lc *LiveCheck1) recoverTraceState(fp uint64, predecessor any, stateNum int) (*TLCStateInfo, error) {
	stateTrace := lc.ownedStateTrace()
	if stateTrace != nil && !stateTrace.IsEmpty() {
		if stateNum < 0 || stateNum >= stateTrace.Size() {
			return nil, newTLCError(ECTLCFailedToRecoverNext, "state %d is outside the simulation trace", stateNum)
		}
		state := stateTrace.At(stateNum)
		if state == nil || state.FingerPrint() != fp {
			return nil, newTLCError(ECTLCFailedToRecoverNext, "trace state %d does not match fingerprint %d", stateNum, fp)
		}
		return NewTLCStateInfo(state), nil
	}
	if predecessor == nil {
		return lc.ownedTool().GetState(fp)
	}
	return lc.ownedTool().GetState(fp, predecessor)
}

func (lc *LiveCheck1) checkSubcomponent(node *BEGraphNode) error {
	slen := len(lc.ownedCurrentOOS().GetCheckState())
	alen := len(lc.ownedCurrentOOS().GetCheckAction())
	aeStateRes := make([]bool, len(lc.ownedCurrentPEM().ownedAEState()))
	aeActionRes := make([]bool, len(lc.ownedCurrentPEM().ownedAEAction()))
	promiseRes := make([]bool, len(lc.ownedCurrentOOS().GetPromises()))
	stack := NewMemObjectStack(lc.ownedMetaDir(), "subcomstack")

	node.IncNumber()
	stack.Push(node)
	for stack.Size() != 0 {
		curNode, _ := stack.Pop().(*BEGraphNode)
		if curNode == nil {
			panic(NewNullPointerException())
		}
		for i, idx := range lc.ownedCurrentPEM().ownedAEState() {
			if !aeStateRes[i] {
				aeStateRes[i] = curNode.GetCheckState(idx)
			}
		}
		for i := 0; i < curNode.NextSize(); i++ {
			node1 := curNode.NextAt(i)
			num := node1.GetNumber()
			if num >= lc.ownedSCCNumbers().thirdNum {
				for j, idx := range lc.ownedCurrentPEM().ownedAEAction() {
					if !aeActionRes[j] {
						aeActionRes[j] = curNode.GetCheckAction(slen, alen, i, idx)
					}
				}
			}
			if num == lc.ownedSCCNumbers().thirdNum {
				node1.IncNumber()
				stack.Push(node1)
			}
		}
		for i, promise := range lc.ownedCurrentOOS().GetPromises() {
			tnode := curNode.GetTNode(lc.ownedCurrentOOS().GetTableau())
			if tnode == nil {
				panic(NewNullPointerException())
			}
			if tnode.GetPar().IsFulfilling(promise) {
				promiseRes[i] = true
			}
		}
	}
	lc.addThirdNum(2)

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
	if lc.claimError() {
		counterExample, err := lc.printErrorTrace(node)
		if err != nil {
			PrintError(ECGeneral, javaGeneralErrorMessage("printing an error trace", err))
			return NewLiveException(ECTLCTemporalPropertyViolated, "LiveCheck: Found error trace.")
		}
		return NewLiveCounterExampleException(ECTLCTemporalPropertyViolated, "LiveCheck: Found error trace.", counterExample)
	}
	return nil
}

func (lc *LiveCheck1) claimError() bool {
	lc.require()
	errorFound := lc.ErrorFound
	if lc.ownersCaptured {
		errorFound = lc.errorFoundOwner
	}
	if errorFound == nil {
		panic(NewNullPointerException())
	}
	return errorFound.CompareAndSwap(false, true)
}

func (lc *LiveCheck1) checkSccs(node *BEGraphNode) (int64, error) {
	lowlink := lc.takeFirstNum()
	node.SetNumber(lowlink)
	lc.ownedComStack().Push(node)
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
	for _, pem := range lc.ownedCurrentOOS().GetPems() {
		lc.setCurrentPEM(pem)
		lc.setSearchStarts()
		for j := len(nodes) - 1; j >= 0; j-- {
			node1 := nodes[j]
			if node1.GetNumber() < lc.ownedSCCNumbers().startThirdNum {
				if _, err := lc.checkSccs1(node1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (lc *LiveCheck1) extractComponent(node *BEGraphNode) []*BEGraphNode {
	node1, _ := lc.ownedComStack().Pop().(*BEGraphNode)
	if node == node1 && !node.TransExists(node) {
		node.SetNumber(liveCheck1MaxFirst)
		return nil
	}
	nodes := make([]*BEGraphNode, 0)
	numFirstCom := lc.startFirstComponent()
	node1.SetNumber(numFirstCom)
	nodes = append(nodes, node1)
	for node != node1 {
		node1, _ = lc.ownedComStack().Pop().(*BEGraphNode)
		node1.SetNumber(numFirstCom)
		nodes = append(nodes, node1)
	}
	return nodes
}

func (lc *LiveCheck1) checkSccs1(node *BEGraphNode) (int64, error) {
	lowlink := lc.takeSecondNum()
	node.SetNumber(lowlink)
	lc.ownedComStack().Push(node)
	for i := 0; i < node.NextSize(); i++ {
		destNode := node.NextAt(i)
		destNum := destNode.GetNumber()
		numbers := lc.ownedSCCNumbers()
		if numbers.numFirstCom <= destNum && node.GetCheckActionAll(len(lc.ownedCurrentOOS().GetCheckState()), len(lc.ownedCurrentOOS().GetCheckAction()), i, lc.ownedCurrentPEM().ownedEAAction()) {
			if destNum < numbers.startSecondNum || (numbers.numSecondCom <= destNum && destNum < numbers.startThirdNum) {
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
	node1, _ := lc.ownedComStack().Pop().(*BEGraphNode)
	if node == node1 && !lc.canStutter(node) {
		node.SetNumber(lc.takeThirdNum())
		return false
	}
	thirdNum := lc.ownedSCCNumbers().thirdNum
	node1.SetNumber(thirdNum)
	for node != node1 {
		node1, _ = lc.ownedComStack().Pop().(*BEGraphNode)
		node1.SetNumber(thirdNum)
	}
	return true
}

func (lc *LiveCheck1) canStutter(node *BEGraphNode) bool {
	slen := len(lc.ownedCurrentOOS().GetCheckState())
	alen := len(lc.ownedCurrentOOS().GetCheckAction())
	for i := 0; i < node.NextSize(); i++ {
		node1 := node.NextAt(i)
		if beGraphNodesEqual(node, node1) {
			nodeNum := node.GetNumber()
			return lc.ownedSCCNumbers().numFirstCom <= nodeNum && node.GetCheckActionAll(slen, alen, i, lc.ownedCurrentPEM().ownedEAAction())
		}
	}
	return false
}

func (lc *LiveCheck1) FinalCheck(tool *Tool) (int, error) {
	return lc.Check(tool, true)
}

func (lc *LiveCheck1) ownedTool() *Tool {
	if lc.ownersCaptured {
		return lc.toolOwner
	}
	return lc.MyTool
}

func (lc *LiveCheck1) ownedMetaDir() string {
	if lc.ownersCaptured {
		return lc.metaDirOwner
	}
	return lc.MetaDir
}

func (lc *LiveCheck1) ownedActions() []*Action {
	if lc.ownersCaptured {
		return lc.actionsOwner
	}
	return lc.Actions
}

func (lc *LiveCheck1) ownedSolutions() []*OrderOfSolution {
	if lc.ownersCaptured {
		return lc.solutionsOwner
	}
	return lc.Solutions
}

func (lc *LiveCheck1) ownedBGraphs() []*BEGraph {
	if lc.ownersCaptured {
		return lc.bgraphsOwner
	}
	return lc.BGraphs
}

func (lc *LiveCheck1) ownedStateTrace() *StateVec {
	if lc.ownersCaptured {
		return lc.stateTraceOwner
	}
	return lc.StateTrace
}

func (lc *LiveCheck1) setCurrentOOS(solution *OrderOfSolution) {
	lc.CurrentOOS = solution
	if lc.ownersCaptured {
		lc.currentOOSOwner = solution
	}
}

func (lc *LiveCheck1) ownedCurrentOOS() *OrderOfSolution {
	if lc.ownersCaptured {
		return lc.currentOOSOwner
	}
	return lc.CurrentOOS
}

func (lc *LiveCheck1) setCurrentPEM(pem *PossibleErrorModel) {
	lc.CurrentPEM = pem
	if lc.ownersCaptured {
		lc.currentPEMOwner = pem
	}
}

func (lc *LiveCheck1) ownedCurrentPEM() *PossibleErrorModel {
	if lc.ownersCaptured {
		return lc.currentPEMOwner
	}
	return lc.CurrentPEM
}

func (lc *LiveCheck1) setComStack(stack *MemObjectStack) {
	lc.ComStack = stack
	if lc.ownersCaptured {
		lc.comStackOwner = stack
	}
}

func (lc *LiveCheck1) ownedComStack() *MemObjectStack {
	if lc.ownersCaptured {
		return lc.comStackOwner
	}
	return lc.ComStack
}

func (lc *LiveCheck1) setInitNode(node *BEGraphNode) {
	lc.InitNode = node
	if lc.ownersCaptured {
		lc.initNodeOwner = node
	}
}

func (lc *LiveCheck1) ownedInitNode() *BEGraphNode {
	if lc.ownersCaptured {
		return lc.initNodeOwner
	}
	return lc.InitNode
}

func (lc *LiveCheck1) ownedSCCNumbers() liveCheck1SCCNumbers {
	if lc.ownersCaptured {
		return lc.sccNumbersOwner
	}
	return liveCheck1SCCNumbers{
		firstNum:       lc.FirstNum,
		secondNum:      lc.SecondNum,
		startSecondNum: lc.StartSecondNum,
		thirdNum:       lc.ThirdNum,
		startThirdNum:  lc.StartThirdNum,
		numFirstCom:    lc.NumFirstCom,
		numSecondCom:   lc.NumSecondCom,
	}
}

func (lc *LiveCheck1) setSCCNumbers(numbers liveCheck1SCCNumbers) {
	lc.FirstNum = numbers.firstNum
	lc.SecondNum = numbers.secondNum
	lc.StartSecondNum = numbers.startSecondNum
	lc.ThirdNum = numbers.thirdNum
	lc.StartThirdNum = numbers.startThirdNum
	lc.NumFirstCom = numbers.numFirstCom
	lc.NumSecondCom = numbers.numSecondCom
	if lc.ownersCaptured {
		lc.sccNumbersOwner = numbers
	}
}

func (lc *LiveCheck1) takeFirstNum() int64 {
	numbers := lc.ownedSCCNumbers()
	value := numbers.firstNum
	numbers.firstNum++
	lc.setSCCNumbers(numbers)
	return value
}

func (lc *LiveCheck1) takeSecondNum() int64 {
	numbers := lc.ownedSCCNumbers()
	value := numbers.secondNum
	numbers.secondNum++
	lc.setSCCNumbers(numbers)
	return value
}

func (lc *LiveCheck1) takeThirdNum() int64 {
	numbers := lc.ownedSCCNumbers()
	value := numbers.thirdNum
	numbers.thirdNum++
	lc.setSCCNumbers(numbers)
	return value
}

func (lc *LiveCheck1) addThirdNum(delta int64) int64 {
	numbers := lc.ownedSCCNumbers()
	numbers.thirdNum += delta
	lc.setSCCNumbers(numbers)
	return numbers.thirdNum
}

func (lc *LiveCheck1) setSearchStarts() {
	numbers := lc.ownedSCCNumbers()
	numbers.startSecondNum = numbers.secondNum
	numbers.startThirdNum = numbers.thirdNum
	lc.setSCCNumbers(numbers)
}

func (lc *LiveCheck1) startFirstComponent() int64 {
	numbers := lc.ownedSCCNumbers()
	numbers.numFirstCom = numbers.secondNum
	numbers.secondNum++
	numbers.numSecondCom = numbers.thirdNum
	lc.setSCCNumbers(numbers)
	return numbers.numFirstCom
}

func (lc *LiveCheck1) GetMetaDir() string {
	lc.require()
	return lc.ownedMetaDir()
}

func (lc *LiveCheck1) GetTool() *Tool {
	lc.require()
	return lc.ownedTool()
}

func (lc *LiveCheck1) GetOutDegreeStatistics() any {
	lc.require()
	return NewDummyBucketStatistics()
}

func (lc *LiveCheck1) GetChecker(idx int) *LiveChecker {
	lc.require()
	return nil
}

func (lc *LiveCheck1) GetNumChecker() int {
	lc.require()
	return 0
}

func (lc *LiveCheck1) NumChecker() int {
	return lc.GetNumChecker()
}

func (lc *LiveCheck1) Close() error                  { lc.require(); return nil }
func (lc *LiveCheck1) BeginChkpt() error             { lc.require(); return nil }
func (lc *LiveCheck1) CommitChkpt() error            { lc.require(); return nil }
func (lc *LiveCheck1) Recover() error                { lc.require(); return nil }
func (lc *LiveCheck1) FlushWritesToDiskFiles() error { lc.require(); return nil }

func (lc *LiveCheck1) CalculateInDegreeDiskGraphs(stats any) (any, error) {
	lc.require()
	return NewDummyBucketStatistics(), nil
}

func (lc *LiveCheck1) CalculateOutDegreeDiskGraphs(stats any) (any, error) {
	lc.require()
	return NewDummyBucketStatistics(), nil
}

func (lc *LiveCheck1) require() {
	if lc == nil {
		panic(NewNullPointerException())
	}
}

func (lc *LiveCheck1) bgraphAt(index int) *BEGraph {
	lc.require()
	if index < 0 || index >= len(lc.ownedBGraphs()) {
		panic(NewArrayIndexOutOfBoundsException(index, len(lc.ownedBGraphs())))
	}
	graph := lc.ownedBGraphs()[index]
	if graph == nil {
		panic(NewNullPointerException())
	}
	return graph
}

func (lc *LiveCheck1) String() string {
	if lc == nil {
		return ""
	}
	return fmt.Sprintf("LiveCheck1[%d]", len(lc.ownedSolutions()))
}
