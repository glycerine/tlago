package tlc

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const javaLivenessPackageName = "tlc2.tool.liveness"
const liveCheckDebugProperty = "tlc2.tool.liveness.LiveCheck.debug"

type LiveException struct {
	ErrorCode      int
	Msg            string
	CounterExample *CounterExample
	Err            error
}

func NewLiveException(errorCode int, msg ...string) *LiveException {
	ex := &LiveException{ErrorCode: errorCode}
	if len(msg) > 0 {
		ex.Msg = msg[0]
	}
	return ex
}

func NewLiveExceptionWithCounterExample(errorCode int, msg string, counterExample *CounterExample) *LiveException {
	return &LiveException{ErrorCode: errorCode, Msg: msg, CounterExample: counterExample}
}

func (e *LiveException) Error() string {
	if e == nil {
		return ""
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("liveness error %d", e.ErrorCode)
}

func (e *LiveException) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type LiveCounterExampleException struct {
	*LiveException
	CounterExample *CounterExample
}

func NewLiveCounterExampleException(errorCode int, msg string, counterExample *CounterExample) *LiveCounterExampleException {
	return &LiveCounterExampleException{
		LiveException:  NewLiveException(errorCode, msg),
		CounterExample: counterExample,
	}
}

func (e *LiveCounterExampleException) Error() string {
	if e != nil && e.LiveException != nil && e.LiveException.Msg != "" {
		return e.LiveException.Msg
	}
	return "temporal property violated"
}

// Go embedding does not implement Java's instanceof inheritance. Match both
// concrete classes without traversing the exception's cause chain.
func javaLiveException(err error) *LiveException {
	switch failure := err.(type) {
	case *LiveCounterExampleException:
		if failure != nil {
			return failure.LiveException
		}
	case *LiveException:
		return failure
	}
	return nil
}

type LivenessStateWriter struct {
	Noop bool
	*StateWriter
}

func NewNoopLivenessStateWriter() *LivenessStateWriter {
	return &LivenessStateWriter{Noop: true}
}

func NewDotLivenessStateWriter(stateWriter *StateWriter) (*LivenessStateWriter, error) {
	fname := "DotStateWriter_liveness.dot"
	if stateWriter != nil && stateWriter.GetDumpFileName() != "" {
		fname = dotWriterJavaReplace(stateWriter.GetDumpFileName(), "_liveness.dot")
	}
	writer, err := NewDotStateWriter(fname, DotStateWriterOptions{})
	if err != nil {
		return nil, err
	}
	return &LivenessStateWriter{StateWriter: writer}, nil
}

func (w *LivenessStateWriter) IsNoop() bool {
	return w == nil || w.Noop || w.StateWriter == nil || w.StateWriter.IsNoop()
}

func (w *LivenessStateWriter) IsDot() bool {
	return w != nil && w.StateWriter != nil && w.StateWriter.IsDot()
}

func (w *LivenessStateWriter) Close() error {
	if w == nil || w.StateWriter == nil {
		return nil
	}
	return w.StateWriter.Close()
}

func (w *LivenessStateWriter) WriteLivenessState(state *TLCStateMut) error {
	if w == nil || w.IsNoop() || w.StateWriter == nil {
		return nil
	}
	return w.StateWriter.WriteInitState(state)
}

func (w *LivenessStateWriter) WriteLivenessStateTransition(state *TLCStateMut, successor *TLCStateMut, actionChecks *BitVector, from int, length int, status StateVisitStatus, visualization StateVisualization) error {
	if w == nil || w.IsNoop() || w.writer == nil || state == nil || successor == nil {
		return nil
	}
	if !w.stuttering && visualization == StateVisualizationStuttering {
		return nil
	}
	sourceFP := state.FingerPrint()
	successorFP := successor.FingerPrint()
	if w.strict != nil {
		key := sourceFP ^ successorFP
		if _, ok := w.strict[key]; ok {
			return nil
		}
		w.strict[key] = struct{}{}
	}
	if _, err := fmt.Fprintf(w.writer, "%d -> %d", sourceFP, successorFP); err != nil {
		return err
	}
	if visualization == StateVisualizationStuttering {
		_, err := w.writer.WriteString(" [style=\"dashed\",color=\"lightgray\"];\n")
		return err
	}
	if length > 0 && actionChecks != nil {
		if _, err := fmt.Fprintf(w.writer, " [label=\"%s\"]", dotEscape(actionChecks.StringRangeChars(from, length, 't', 'f'))); err != nil {
			return err
		}
	}
	if _, err := w.writer.WriteString(";\n"); err != nil {
		return err
	}
	if status != StateVisitSeen {
		_, err := fmt.Fprintf(w.writer, "%d [label=\"%s\",tooltip=\"%s\"];\n",
			successorFP,
			stateToDot(successor.EvalStateLevelAlias(), state.EvalStateLevelAlias(), printDiffsOnly()),
			stateToDot(successor, nil, false),
		)
		return err
	}
	return nil
}

func (w *LivenessStateWriter) WriteLivenessInitState(state *TLCStateMut, tableauNode *TBGraphNode) error {
	if w == nil || w.IsNoop() || w.writer == nil || state == nil || tableauNode == nil {
		return nil
	}
	fp := state.FingerPrint()
	_, err := fmt.Fprintf(w.writer, "\"%d.%d\" [style = filled] [label=\"%s\\n#%d.%d#\"]\n",
		fp,
		tableauNode.Index,
		stateToDot(state, nil, false),
		fp,
		tableauNode.Index,
	)
	return err
}

func (w *LivenessStateWriter) WriteLivenessTransition(state *TLCStateMut, tableauNode *TBGraphNode, successor *TLCStateMut, successorTableauNode *TBGraphNode, actionChecks *BitVector, from int, length int, status StateVisitStatus) error {
	return w.WriteLivenessTransitionVisual(state, tableauNode, successor, successorTableauNode, actionChecks, from, length, status, StateVisualizationDefault)
}

func (w *LivenessStateWriter) WriteLivenessTransitionVisual(state *TLCStateMut, tableauNode *TBGraphNode, successor *TLCStateMut, successorTableauNode *TBGraphNode, actionChecks *BitVector, from int, length int, status StateVisitStatus, visualization StateVisualization) error {
	if w == nil || w.IsNoop() || w.writer == nil || state == nil || tableauNode == nil || successor == nil || successorTableauNode == nil {
		return nil
	}
	sourceFP := state.FingerPrint()
	successorFP := successor.FingerPrint()
	if _, err := fmt.Fprintf(w.writer, "\"%d.%d\" -> \"%d.%d\"",
		sourceFP,
		tableauNode.Index,
		successorFP,
		successorTableauNode.Index,
	); err != nil {
		return err
	}
	switch visualization {
	case StateVisualizationStuttering:
		if _, err := w.writer.WriteString(" [style=\"dashed\"]"); err != nil {
			return err
		}
	case StateVisualizationDotted:
		if _, err := w.writer.WriteString(" [style=\"dotted\"]"); err != nil {
			return err
		}
	}
	if length > 0 && actionChecks != nil {
		label := actionChecks.StringRangeChars(from, length, 't', 'f')
		if _, err := fmt.Fprintf(w.writer, " [label=\"%s\"]", dotEscape(label)); err != nil {
			return err
		}
	}
	if _, err := w.writer.WriteString(";\n"); err != nil {
		return err
	}
	if status != StateVisitSeen {
		_, err := fmt.Fprintf(w.writer, "\"%d.%d\" [label=\"%s\\n#%d.%d#\"];\n",
			successorFP,
			successorTableauNode.Index,
			stateToDot(successor, nil, false),
			successorFP,
			successorTableauNode.Index,
		)
		return err
	}
	return nil
}

type LiveChecker struct {
	Solution         *OrderOfSolution
	Soln             int
	Writer           *LivenessStateWriter
	Graph            *InsMap[string, *GraphNode]
	Initial          []*GraphNode
	Size             int64
	LastSize         int64
	DiskGraph        *DiskGraph
	TableauDiskGraph *TableauDiskGraph
	ErrorGraphNode   *GraphNode
	ErrorPrefix      *LongVec
	ErrorCycle       *LongVec
	ErrorTrace       []*TLCStateInfo
	ErrorRawTrace    []*TLCStateInfo
	ErrorCounterEx   *CounterExample
	ErrorLoopOrdinal int
	ErrorCycleIndex  int
	ErrorClosingInfo *TLCStateInfo
	ErrorStuttering  bool
	ErrorPrinted     bool
	Err              error
}

func NewLiveChecker(solution *OrderOfSolution, soln int, writer *LivenessStateWriter, metadir string, outDegreeStats any) *LiveChecker {
	if writer == nil {
		writer = NewNoopLivenessStateWriter()
	}
	checker := &LiveChecker{
		Solution: solution,
		Soln:     soln,
		Writer:   writer,
		Graph:    NewInsMap[string, *GraphNode](),
	}
	if metadir != "" && solution != nil {
		if solution.HasTableau() {
			if liveCheckDebugEnabled() {
				checker.TableauDiskGraph, checker.Err = NewDebugTableauDiskGraph(metadir, soln, outDegreeStats, solution)
			} else {
				checker.TableauDiskGraph, checker.Err = NewTableauDiskGraph(metadir, soln, outDegreeStats)
			}
		} else {
			checker.DiskGraph, checker.Err = NewDiskGraph(metadir, soln, outDegreeStats)
		}
	}
	return checker
}

func liveCheckDebugEnabled() bool {
	value, ok := tlcLookupSystemProperty(liveCheckDebugProperty)
	return ok && javaBooleanProperty(value)
}

func (c *LiveChecker) AddInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	if c == nil || c.Solution == nil {
		return nil
	}
	if c.Err != nil {
		return c.Err
	}
	if c.Solution.HasTableau() {
		for i := 0; i < c.Solution.Tableau.InitCnt; i++ {
			tnode := c.Solution.Tableau.GetNode(i)
			ok, err := tnode.IsConsistent(state, tool)
			if err != nil {
				return err
			}
			if ok {
				if c.TableauDiskGraph != nil {
					c.TableauDiskGraph.AddInitNode(stateFP, tnode.Index)
					c.TableauDiskGraph.RecordNode(stateFP, tnode.Index)
				}
				if err := c.Writer.WriteLivenessInitState(state, tnode); err != nil {
					return err
				}
				node, err := c.ensureGraphNode(tool, state, stateFP, tnode.Index)
				if err != nil {
					return err
				}
				c.Initial = append(c.Initial, node)
			}
		}
		return nil
	}
	if c.DiskGraph != nil {
		c.DiskGraph.AddInitNode(stateFP, -1)
	}
	if err := c.Writer.WriteLivenessState(state); err != nil {
		return err
	}
	node, err := c.ensureGraphNode(tool, state, stateFP, -1)
	if err != nil {
		return err
	}
	c.Initial = append(c.Initial, node)
	return nil
}

func (c *LiveChecker) AddNextState(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
	if c == nil || c.Solution == nil || nextStates == nil {
		return nil
	}
	if c.Err != nil {
		return c.Err
	}
	if c.Solution.HasTableau() {
		return c.addNextStateTableau(tool, s0, fp0, nextStates, actionResults, checkStateRes)
	}
	if c.DiskGraph != nil {
		if err := c.addNextStateDisk(s0, fp0, nextStates, actionResults, checkStateRes); err != nil {
			return err
		}
	}
	source, err := c.ensureGraphNodeWithStateChecks(fp0, -1, checkStateRes)
	if err != nil {
		return err
	}
	alen := len(c.Solution.CheckAction)
	nextStates.ResetNext()
	for idx := 0; idx < nextStates.Size(); idx++ {
		s1 := nextStates.Next()
		if s1 == nil {
			continue
		}
		target, err := c.ensureGraphNode(tool, s1, s1.FingerPrint(), -1)
		if err != nil {
			return err
		}
		_ = target
		source.AddTransition(s1.FingerPrint(), -1, len(c.Solution.CheckState), alen, actionResults, alen*idx, nextStates.Size()-idx)
	}
	nextStates.ResetNext()
	return nil
}

func (c *LiveChecker) addNextStateDisk(s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
	dgraph := c.DiskGraph
	if dgraph == nil {
		return nil
	}
	cnt := 0
	succCnt := nextStates.Size()
	alen := len(c.Solution.CheckAction)
	node0, err := dgraph.GetNodeNoTableau(fp0)
	if err != nil {
		return err
	}
	originalSuccSize := node0.SuccSize()
	node0.SetCheckState(checkStateRes)
	nextStates.ResetNext()
	for sidx := 0; sidx < succCnt; sidx++ {
		successorState := nextStates.Next()
		if successorState == nil {
			continue
		}
		successor := successorState.FingerPrint()
		ptr1 := dgraph.GetPtr(successor, -1)
		if ptr1 == -1 || !node0.TransExists(successor, -1) {
			node0.AddTransition(successor, -1, len(checkStateRes), alen, actionResults, sidx*alen, succCnt-cnt)
		}
		status := StateVisitSeen
		if ptr1 == -1 {
			status = StateVisitUnseen
		}
		if err := c.Writer.WriteLivenessStateTransition(s0, successorState, actionResults, sidx*alen, alen, status, StateVisualizationDefault); err != nil {
			return err
		}
		cnt++
	}
	nextStates.ResetNext()
	if (originalSuccSize == 0 && originalSuccSize == node0.SuccSize()) || originalSuccSize < node0.SuccSize() {
		node0.Realign()
		if _, err := dgraph.AddNode(node0); err != nil {
			return err
		}
		c.Size = int64(dgraph.Size())
	}
	return nil
}

func (c *LiveChecker) addNextStateTableau(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
	if c.TableauDiskGraph != nil {
		if err := c.addNextStateTableauDisk(tool, s0, fp0, nextStates, actionResults, checkStateRes); err != nil {
			return err
		}
	}
	alen := len(c.Solution.CheckAction)
	nextStates.ResetNext()
	for idx := 0; idx < nextStates.Size(); idx++ {
		s1 := nextStates.Next()
		if s1 == nil {
			continue
		}
		for _, srcTNode := range c.Solution.Tableau.Nodes {
			sourceKey := graphNodeKey(fp0, srcTNode.Index)
			source := c.Graph.Get(sourceKey)
			if source == nil {
				continue
			}
			for _, dstTNode := range srcTNode.Nexts {
				ok, err := dstTNode.IsConsistent(s1, tool)
				if err != nil {
					return err
				}
				if !ok {
					continue
				}
				target, err := c.ensureGraphNode(tool, s1, s1.FingerPrint(), dstTNode.Index)
				if err != nil {
					return err
				}
				_ = target
				source.AddTransition(s1.FingerPrint(), dstTNode.Index, len(c.Solution.CheckState), alen, actionResults, alen*idx, nextStates.Size()-idx)
			}
		}
	}
	nextStates.ResetNext()
	_ = checkStateRes
	return nil
}

func (c *LiveChecker) addNextStateTableauDisk(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
	dgraph := c.TableauDiskGraph
	oos := c.Solution
	if dgraph == nil || oos == nil || oos.Tableau == nil {
		return nil
	}
	cnt := 0
	succCnt := nextStates.Size()
	tableau := oos.Tableau
	consistency := NewBitVector(tableau.Size() * succCnt)
	for _, tableauNode := range tableau.Nodes {
		nextStates.ResetNext()
		for sidx := 0; sidx < succCnt; sidx++ {
			s1 := nextStates.Next()
			if s1 == nil {
				continue
			}
			ok, err := tableauNode.IsConsistent(s1, tool)
			if err != nil {
				return err
			}
			if ok {
				consistency.Set(tableauNode.Index*succCnt + sidx)
			}
		}
	}

	loc0 := dgraph.SetDone(fp0)
	nodes := dgraph.GetNodesByLoc(loc0)
	if nodes == nil {
		return nil
	}
	alen := len(oos.CheckAction)
	allocationHint := (len(nodes) / dgraph.GetElemLength()) * succCnt
	for nidx := 2; nidx < len(nodes); nidx += dgraph.GetElemLength() {
		tidx0 := int(nodes[nidx])
		tnode0 := oos.Tableau.GetNode(tidx0)
		node0, err := dgraph.GetNode(fp0, tidx0)
		if err != nil {
			return err
		}
		originalSuccSize := node0.SuccSize()
		node0.SetCheckState(checkStateRes)
		nextStates.ResetNext()
		for sidx := 0; sidx < succCnt; sidx++ {
			s1 := nextStates.Next()
			if s1 == nil {
				continue
			}
			successor := s1.FingerPrint()
			isDone := dgraph.IsDone(successor)
			for _, tnode1 := range tnode0.Nexts {
				ptr1 := dgraph.GetPtr(successor, tnode1.Index)
				if consistency.Get(tnode1.Index*succCnt+sidx) && (ptr1 == -1 || !node0.TransExists(successor, tnode1.Index)) {
					node0.AddTransition(successor, tnode1.Index, len(checkStateRes), alen, actionResults, sidx*alen, allocationHint-cnt)
					if err := c.Writer.WriteLivenessTransition(s0, tnode0, s1, tnode1, actionResults, sidx*alen, alen, StateVisitUnseen); err != nil {
						return err
					}
					if ptr1 == -1 {
						dgraph.RecordNode(successor, tnode1.Index)
						if isDone {
							if err := c.addNextStateTableauDone(tool, s1, successor, tnode1); err != nil {
								return err
							}
						}
					}
				}
				cnt++
			}
		}
		nextStates.ResetNext()
		if (originalSuccSize == 0 && originalSuccSize == node0.SuccSize()) || originalSuccSize < node0.SuccSize() {
			node0.Realign()
			if _, err := dgraph.AddNode(node0); err != nil {
				return err
			}
			c.Size = int64(dgraph.Size())
		}
	}
	if c.ErrorGraphNode != nil {
		dgraph.CreateCache()
		prefix, err := dgraph.GetPath(c.ErrorGraphNode.StateFP, c.ErrorGraphNode.TIndex)
		dgraph.DestroyCache()
		if err != nil {
			return err
		}
		c.ErrorPrefix = prefix
		c.ErrorGraphNode = nil
		return c.printSafetyLikeLivenessError(tool, prefix)
	}
	return nil
}

func (c *LiveChecker) printSafetyLikeLivenessError(tool *Tool, prefix *LongVec) error {
	if c == nil {
		return nil
	}
	trace, err := c.reconstructSafetyLikeLivenessPrefix(tool, prefix)
	if err != nil {
		return err
	}
	if len(trace) == 0 {
		return nil
	}
	if mc := MainChecker(); mc != nil {
		mc.mu.Lock()
		if mc.PrintedLivenessErrorStack {
			mc.mu.Unlock()
			return errInvariantViolated
		}
		mc.PrintedLivenessErrorStack = true
		mc.mu.Unlock()
	}

	names := LivenessFindViolatedPropertiesFromTrace(tool, trace, len(trace)-1)
	PrintError(ECTLCTemporalPropertyViolated, names...)
	PrintError(ECTLCCounterExample)
	for _, info := range trace {
		PrintInvariantViolationStateTraceState(info)
	}

	c.ErrorTrace = trace
	c.ErrorLoopOrdinal = len(trace)
	c.ErrorStuttering = true
	c.ErrorClosingInfo = trace[len(trace)-1]
	c.ErrorCounterEx = NewCounterExampleFromTrace(trace)
	c.ErrorPrinted = true

	if mc := MainChecker(); mc != nil {
		last := trace[len(trace)-1].OriginalState()
		var pred *TLCStateMut
		if len(trace) > 1 {
			pred = trace[len(trace)-2].OriginalState()
		} else {
			pred = last
			last = nil
		}
		mc.SetErrState(pred, last, false, ECTLCInvariantViolatedBehavior)
		mc.Stop()
		if tool != nil {
			tool.CheckPostConditionWithCounterExample(NewCounterExampleFromTrace(trace))
		}
	}
	return errInvariantViolated
}

func (c *LiveChecker) reconstructSafetyLikeLivenessPrefix(tool *Tool, prefix *LongVec) ([]*TLCStateInfo, error) {
	if tool == nil {
		return nil, newTLCError(ECTLCFailedToRecoverInit, "cannot recover liveness trace without a tool")
	}
	if prefix == nil || prefix.Size() == 0 {
		return nil, newTLCError(ECTLCFailedToRecoverInit, "cannot recover empty liveness trace")
	}
	plen := prefix.Size()
	fp := uint64(prefix.ElementAt(plen - 1))
	info, err := tool.GetState(fp)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, newTLCError(ECTLCFailedToRecoverInit, "initial state fingerprint %d could not be regenerated", fp)
	}
	trace := []*TLCStateInfo{info}
	for i := plen - 2; i >= 0; i-- {
		curFP := uint64(prefix.ElementAt(i))
		if curFP == fp {
			continue
		}
		next, err := tool.GetState(curFP, info)
		if err != nil {
			return nil, err
		}
		if next == nil {
			return nil, newTLCError(ECTLCFailedToRecoverNext, "successor fingerprint %d could not be regenerated", curFP)
		}
		if alias, err := tool.EvalAliasInfoPair(trace[len(trace)-1], next.State); err != nil {
			return nil, err
		} else if alias != nil {
			trace[len(trace)-1] = alias
		}
		trace = append(trace, next)
		info = next
		fp = curFP
	}
	last := trace[len(trace)-1]
	if alias, err := tool.EvalAliasInfoPair(last, last.State); err != nil {
		return nil, err
	} else if alias != nil {
		trace[len(trace)-1] = alias
	}
	return trace, nil
}

func (c *LiveChecker) addNextStateTableauDone(tool *Tool, state *TLCStateMut, fp uint64, tnode *TBGraphNode) error {
	dgraph := c.TableauDiskGraph
	oos := c.Solution
	if dgraph == nil || oos == nil || tnode == nil {
		return nil
	}
	checkStateRes, err := oos.CheckStateValues(tool, state)
	if err != nil {
		return err
	}
	slen := len(checkStateRes)
	alen := len(oos.CheckAction)
	node, err := dgraph.GetNode(fp, tnode.Index)
	if err != nil {
		return err
	}
	numSucc := node.SuccSize()
	node.SetCheckState(checkStateRes)

	cnt := 0
	nextSize := len(tnode.Nexts)
	var selfLoopActionResults *BitVector
	if nextSize > 0 {
		selfLoopActionResults, err = oos.CheckActionBitVector(tool, state, state, NewBitVector(alen), 0)
		if err != nil {
			return err
		}
	}
	for _, tnode1 := range tnode.Nexts {
		tidx1 := tnode1.Index
		ptr1 := dgraph.GetPtr(fp, tidx1)
		ok, err := tnode1.IsConsistent(state, tool)
		if err != nil {
			return err
		}
		if ok {
			if tnode1.IsAccepting() && c.ErrorGraphNode == nil && oos.HasEmptyPEM() {
				c.ErrorGraphNode = node
				return nil
			}
			if ptr1 == -1 || !node.TransExists(fp, tidx1) {
				node.AddTransition(fp, tidx1, slen, alen, selfLoopActionResults, 0, nextSize-cnt)
				if ptr1 == -1 {
					dgraph.RecordNode(fp, tidx1)
					if err := c.addNextStateTableauDone(tool, state, fp, tnode1); err != nil {
						return err
					}
				}
			}
		}
		cnt++
	}

	cnt = 0
	actions := tool.GetActions()
	restoreCurrentState := PushCurrentState(state)
	defer restoreCurrentState()
	for _, action := range actions {
		nextStates, err := tool.GetNextStates(action, state)
		if err != nil {
			return err
		}
		nextCnt := nextStates.Size()
		for j := 0; j < nextCnt; j++ {
			s1 := nextStates.At(j)
			inModel, err := tool.IsInModel(s1)
			if err != nil {
				return err
			}
			inActions := false
			if inModel {
				inActions, err = tool.IsInActions(state, s1)
				if err != nil {
					return err
				}
			}
			if inModel && inActions {
				fp1 := s1.FingerPrint()
				checkActionRes, err := oos.CheckActionBitVector(tool, state, s1, NewBitVector(alen), 0)
				if err != nil {
					return err
				}
				isDone := dgraph.IsDone(fp1)
				for _, tnode1 := range tnode.Nexts {
					tidx1 := tnode1.Index
					ptr1 := dgraph.GetPtr(fp1, tidx1)
					total := len(actions) * nextCnt * len(tnode.Nexts)
					ok, err := tnode1.IsConsistent(s1, tool)
					if err != nil {
						return err
					}
					if ok && (ptr1 == -1 || !node.TransExists(fp1, tidx1)) {
						node.AddTransition(fp1, tidx1, slen, alen, checkActionRes, 0, total-cnt)
						if err := c.Writer.WriteLivenessTransitionVisual(state, tnode, s1, tnode1, checkActionRes, 0, alen, StateVisitSeen, StateVisualizationDotted); err != nil {
							return err
						}
						if ptr1 == -1 {
							dgraph.RecordNode(fp1, tidx1)
							if isDone {
								if err := c.addNextStateTableauDone(tool, s1, fp1, tnode1); err != nil {
									return err
								}
							}
						}
					}
					cnt++
				}
			} else {
				cnt++
			}
		}
	}
	if numSucc < node.SuccSize() {
		node.Realign()
		if _, err := dgraph.AddNode(node); err != nil {
			return err
		}
		c.Size = int64(dgraph.Size())
	}
	return nil
}

func (c *LiveChecker) ensureGraphNode(tool *Tool, state *TLCStateMut, fp uint64, tidx int) (*GraphNode, error) {
	checks, err := c.Solution.CheckStateValues(tool, state)
	if err != nil {
		return nil, err
	}
	return c.ensureGraphNodeWithStateChecks(fp, tidx, checks)
}

func (c *LiveChecker) ensureGraphNodeWithStateChecks(fp uint64, tidx int, checks []bool) (*GraphNode, error) {
	key := graphNodeKey(fp, tidx)
	if node := c.Graph.Get(key); node != nil {
		node.SetCheckState(checks)
		return node, nil
	}
	node := NewGraphNode(fp, tidx)
	node.SetCheckState(checks)
	c.Graph.Set(key, node)
	c.Size++
	return node, nil
}

func (c *LiveChecker) CheckSccs(tool *Tool, finalCheck bool, liveCheck *LiveCheck, workerID int, numWorkers int) (bool, error) {
	_ = tool
	_ = finalCheck
	if c == nil || c.Solution == nil || c.Graph == nil {
		return false, nil
	}
	if c.DiskGraph != nil || c.TableauDiskGraph != nil {
		c.createDiskGraphCache()
		defer c.destroyDiskGraphCache()
		for _, pem := range c.Solution.PEMs {
			if liveCheck != nil && liveCheck.hasLiveError() {
				break
			}
			worker := NewLiveWorker(tool, workerID, numWorkers, liveCheck, c, pem, finalCheck)
			found, err := worker.CheckSccs()
			if err != nil {
				return false, err
			}
			if found {
				c.PrintCounterExample(tool)
				c.recordGraphSize()
				return true, nil
			}
		}
		c.recordGraphSize()
		return false, nil
	}
	return false, fmt.Errorf("liveness checker missing disk graph")
}

func (c *LiveChecker) Reset() {
	c.Graph = NewInsMap[string, *GraphNode]()
	c.Initial = nil
	c.Size = 0
	c.LastSize = 0
	c.ErrorGraphNode = nil
	c.ErrorPrefix = nil
	c.ErrorCycle = nil
	c.ErrorTrace = nil
	c.ErrorRawTrace = nil
	c.ErrorCounterEx = nil
	c.ErrorLoopOrdinal = 0
	c.ErrorCycleIndex = 0
	c.ErrorClosingInfo = nil
	c.ErrorStuttering = false
	c.ErrorPrinted = false
}

func graphNodeKey(fp uint64, tidx int) string {
	return fmt.Sprintf("%d:%d", fp, tidx)
}

type LiveCheck struct {
	Tool           *Tool
	MetaDir        string
	Checkers       []*LiveChecker
	OutDegreeStats any
	NoOp           bool
	AddAndCheck    bool
	mu             sync.Mutex
	errMu          sync.Mutex
	errFoundByID   int
}

func NewNoOpLiveCheck(tool *Tool, metadir string) *LiveCheck {
	return &LiveCheck{Tool: tool, MetaDir: metadir, OutDegreeStats: NewDummyBucketStatistics(), NoOp: true, errFoundByID: -1}
}

func NewLiveCheck(tool *Tool, solutions []*OrderOfSolution, metadir string) *LiveCheck {
	check, _ := NewLiveCheckWithStateWriter(tool, solutions, metadir, nil)
	return check
}

func NewLiveCheckWithStateWriter(tool *Tool, solutions []*OrderOfSolution, metadir string, stateWriter *StateWriter) (*LiveCheck, error) {
	if metadir == "" {
		tmp, err := os.MkdirTemp("", "tlago-livecheck-")
		if err != nil {
			return nil, err
		}
		metadir = tmp
	}
	var outDegreeStats any = NewDummyBucketStatistics()
	if livenessStatsEnabled() {
		outDegreeStats = NewConcurrentBucketStatisticsWithMX("Histogram vertex out-degree", javaLivenessPackageName, "DiskGraphsOutDegree")
	}
	check := &LiveCheck{Tool: tool, MetaDir: metadir, OutDegreeStats: outDegreeStats, errFoundByID: -1}
	for i, solution := range solutions {
		writer := NewNoopLivenessStateWriter()
		if stateWriter != nil && !stateWriter.IsNoop() && stateWriter.IsDot() {
			dotWriter, err := NewDotLivenessStateWriter(stateWriter)
			if err != nil {
				return nil, err
			}
			writer = dotWriter
		}
		check.Checkers = append(check.Checkers, NewLiveChecker(solution, i, writer, metadir, check.OutDegreeStats))
	}
	return check, nil
}

func NewAddAndCheckLiveCheck(tool *Tool, solutions []*OrderOfSolution, metadir string) *LiveCheck {
	check := NewLiveCheck(tool, solutions, metadir)
	if check != nil {
		check.AddAndCheck = true
	}
	PrintWarning(ECUnitTest, "!!!WARNING: TLC is running in inefficient unit testing mode!!!", "")
	return check
}

func (lc *LiveCheck) NumChecker() int {
	if lc == nil {
		return 0
	}
	return len(lc.Checkers)
}

func (lc *LiveCheck) AddInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	if lc == nil || lc.NoOp {
		return nil
	}
	if lc.AddAndCheck {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		if err := lc.addInitState(tool, state, stateFP); err != nil {
			return err
		}
		_, err := lc.check0(tool, false)
		return err
	}
	return lc.addInitState(tool, state, stateFP)
}

func (lc *LiveCheck) addInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	for _, checker := range lc.Checkers {
		if err := checker.AddInitState(tool, state, stateFP); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) AddNextState(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates) error {
	if lc == nil || lc.NoOp {
		return nil
	}
	if lc.AddAndCheck {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		if err := lc.addNextState(tool, s0, fp0, nextStates); err != nil {
			return err
		}
		_, err := lc.check0(tool, false)
		return err
	}
	return lc.addNextState(tool, s0, fp0, nextStates)
}

func (lc *LiveCheck) addNextState(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates) error {
	for _, checker := range lc.Checkers {
		oos := checker.Solution
		alen := len(oos.CheckAction)
		actionResults := NewBitVector(alen * nextStates.Size())
		nextStates.ResetNext()
		for sidx := 0; sidx < nextStates.Size(); sidx++ {
			s1 := nextStates.Next()
			if s1 == nil {
				continue
			}
			if _, err := oos.CheckActionBitVector(tool, s0, s1, actionResults, alen*sidx); err != nil {
				return err
			}
		}
		nextStates.ResetNext()
		checkState, err := oos.CheckStateValues(tool, s0)
		if err != nil {
			return err
		}
		if err := checker.AddNextState(tool, s0, fp0, nextStates, actionResults, checkState); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) DoLiveCheck() bool {
	if lc == nil || lc.NoOp {
		return false
	}
	threshold := 0.0
	Globals.Lock()
	threshold = Globals.LivenessThreshold
	Globals.Unlock()
	for _, checker := range lc.Checkers {
		size := checker.GraphSize()
		sizeAtLastCheck := checker.SizeAtLastCheck()
		delta := float64(size-sizeAtLastCheck) / float64(sizeAtLastCheck)
		if delta > threshold {
			return true
		}
	}
	return false
}

func (lc *LiveCheck) Check(tool *Tool, forceCheck bool) (int, error) {
	if lc == nil || lc.NoOp {
		return NoError, nil
	}
	if !forceCheck && !DoLiveness() {
		return NoError, nil
	}
	if !forceCheck && !lc.DoLiveCheck() {
		return NoError, nil
	}
	return lc.check0(tool, false)
}

func (lc *LiveCheck) FinalCheck(tool *Tool) (int, error) {
	if lc == nil || lc.NoOp {
		return NoError, nil
	}
	if livenessFinalCheckOff() {
		return NoError, nil
	}
	return lc.check0(tool, true)
}

func livenessFinalCheckOff() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.LNCheck == "off"
}

func (lc *LiveCheck) check0(tool *Tool, finalCheck bool) (int, error) {
	start := time.Now()
	sum := int64(0)
	for _, checker := range lc.Checkers {
		sum += checker.GraphSize()
	}
	branches := ""
	if len(lc.Checkers) != 1 {
		branches = fmt.Sprintf("%d branches of ", len(lc.Checkers))
	}
	space := "current"
	if finalCheck {
		space = "complete"
	}
	PrintMessage(ECTLCCheckingTemporalProps, space, fmtInt64(sum), branches)
	lc.resetLiveError()

	type liveCheckResult struct {
		found   bool
		checker *LiveChecker
		err     error
	}
	wNum := 1
	if !DoSequentialLiveness() {
		wNum = min(len(lc.Checkers), NumWorkers())
		if wNum < 1 {
			wNum = 1
		}
	}
	queue := make(chan *LiveChecker, len(lc.Checkers))
	for _, checker := range lc.Checkers {
		queue <- checker
	}
	close(queue)
	results := make(chan liveCheckResult, wNum)
	for id := 0; id < wNum; id++ {
		go func(workerID int) {
			var result liveCheckResult
			for checker := range queue {
				if lc.hasLiveError() {
					break
				}
				found, err := checker.CheckSccs(tool, finalCheck, lc, workerID, wNum)
				if err != nil {
					result.err = err
					break
				}
				if found {
					result.found = true
					result.checker = checker
					break
				}
			}
			results <- result
		}(id)
	}

	var firstErr error
	var foundChecker *LiveChecker
	for i := 0; i < wNum; i++ {
		result := <-results
		if result.found && foundChecker == nil {
			foundChecker = result.checker
		}
		if result.err != nil && firstErr == nil {
			firstErr = result.err
		}
	}
	if foundChecker != nil {
		PrintMessage(ECTLCCheckingTemporalPropsEnd, humanReadableTLCRuntime(time.Since(start)))
		if foundChecker.ErrorCounterEx != nil {
			return ECTLCTemporalPropertyViolated, NewLiveCounterExampleException(ECTLCTemporalPropertyViolated, "temporal property violated", foundChecker.ErrorCounterEx)
		}
		return ECTLCTemporalPropertyViolated, NewLiveException(ECTLCTemporalPropertyViolated, "temporal property violated")
	}
	if firstErr != nil {
		return printLivenessWorkerFailure(firstErr), firstErr
	}
	if !finalCheck {
		for _, checker := range lc.Checkers {
			if checker == nil {
				continue
			}
			if checker.TableauDiskGraph != nil {
				if err := checker.TableauDiskGraph.MakeNodePtrTbl(); err != nil {
					return ECGeneral, err
				}
			} else if checker.DiskGraph != nil {
				if err := checker.DiskGraph.MakeNodePtrTbl(); err != nil {
					return ECGeneral, err
				}
			}
		}
	}
	PrintMessage(ECTLCCheckingTemporalPropsEnd, humanReadableTLCRuntime(time.Since(start)))
	return NoError, nil
}

func printLivenessWorkerFailure(err error) int {
	if err == nil {
		PrintError(ECGeneral)
		return ECGeneral
	}
	// firstErr is already the worker failure, corresponding to the single
	// ExecutionException.getCause() in Java LiveCheck.check0.
	switch javaSystemFailureCode(err) {
	case ECSystemOutOfMemory, ECSystemOutOfMemoryLiveness:
		PrintError(ECSystemOutOfMemoryLiveness)
		return ECSystemOutOfMemoryLiveness
	case ECSystemStackOverflow:
		PrintError(ECSystemStackOverflow)
		return ECSystemStackOverflow
	}
	PrintError(ECGeneral, generalErrorParams("", err)...)
	return ECGeneral
}

func (lc *LiveCheck) resetLiveError() {
	lc.errMu.Lock()
	lc.errFoundByID = -1
	lc.errMu.Unlock()
}

func (lc *LiveCheck) hasLiveError() bool {
	if lc == nil {
		return false
	}
	lc.errMu.Lock()
	defer lc.errMu.Unlock()
	return lc.errFoundByID != -1
}

func (lc *LiveCheck) claimLiveError(workerID int) bool {
	if lc == nil {
		return true
	}
	lc.errMu.Lock()
	defer lc.errMu.Unlock()
	if lc.errFoundByID == -1 {
		lc.errFoundByID = workerID
		return true
	}
	return lc.errFoundByID == workerID
}

func (lc *LiveCheck) CheckTrace(tool *Tool, trace func() *StateVec) error {
	if lc == nil || lc.NoOp || trace == nil {
		return nil
	}
	states := trace()
	if states == nil || states.Size() == 0 {
		return nil
	}
	first := states.At(0)
	if err := lc.AddInitState(tool, first, first.FingerPrint()); err != nil {
		return err
	}
	successors := NewSetOfStates(states.Size() * 2)
	for i := 0; i < states.Size()-1; i++ {
		successors.Clear()
		state := states.At(i)
		successors.Put(state)
		successors.Put(states.At(i + 1))
		if err := lc.AddNextState(tool, state, state.FingerPrint(), successors); err != nil {
			return err
		}
	}
	last := states.At(states.Size() - 1)
	if err := lc.AddNextState(tool, last, last.FingerPrint(), NewSetOfStates(0)); err != nil {
		return err
	}
	result, err := lc.check0(tool, true)
	if err != nil {
		return err
	}
	if result != NoError {
		return NewLiveException(result)
	}
	return lc.Reset()
}

func (c *LiveChecker) GraphSize() int64 {
	if c == nil {
		return 0
	}
	if c.TableauDiskGraph != nil {
		return int64(c.TableauDiskGraph.Size())
	}
	if c.DiskGraph != nil {
		return int64(c.DiskGraph.Size())
	}
	return c.Size
}

func (c *LiveChecker) SizeAtLastCheck() int64 {
	if c == nil {
		return 0
	}
	if c.TableauDiskGraph != nil {
		return c.TableauDiskGraph.GetSizeAtLastCheck()
	}
	if c.DiskGraph != nil {
		return c.DiskGraph.GetSizeAtLastCheck()
	}
	return c.LastSize
}

func (c *LiveChecker) recordGraphSize() {
	if c == nil {
		return
	}
	if c.TableauDiskGraph != nil {
		c.TableauDiskGraph.RecordSize()
		c.LastSize = c.GraphSize()
		return
	}
	if c.DiskGraph != nil {
		c.DiskGraph.RecordSize()
		c.LastSize = c.GraphSize()
		return
	}
	c.LastSize = c.GraphSize()
}

func (c *LiveChecker) createDiskGraphCache() {
	if c == nil {
		return
	}
	if c.TableauDiskGraph != nil {
		c.TableauDiskGraph.CreateCache()
		return
	}
	if c.DiskGraph != nil {
		c.DiskGraph.CreateCache()
	}
}

func (c *LiveChecker) destroyDiskGraphCache() {
	if c == nil {
		return
	}
	if c.TableauDiskGraph != nil {
		c.TableauDiskGraph.DestroyCache()
		return
	}
	if c.DiskGraph != nil {
		c.DiskGraph.DestroyCache()
	}
}

func (c *LiveChecker) Close() error {
	if c == nil {
		return nil
	}
	var err error
	if c.TableauDiskGraph != nil {
		err = c.TableauDiskGraph.Close()
	} else if c.DiskGraph != nil {
		err = c.DiskGraph.Close()
	}
	if c.Writer != nil {
		if closeErr := c.Writer.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}

func (c *LiveChecker) BeginChkpt() error {
	if c == nil {
		return nil
	}
	if c.TableauDiskGraph != nil {
		return c.TableauDiskGraph.BeginChkpt()
	}
	if c.DiskGraph != nil {
		return c.DiskGraph.BeginChkpt()
	}
	return nil
}

func (c *LiveChecker) CommitChkpt() error {
	if c == nil {
		return nil
	}
	if c.TableauDiskGraph != nil {
		return c.TableauDiskGraph.CommitChkpt()
	}
	if c.DiskGraph != nil {
		return c.DiskGraph.CommitChkpt()
	}
	return nil
}

func (c *LiveChecker) FlushWritesToDiskFiles() error {
	if c == nil {
		return nil
	}
	if c.TableauDiskGraph != nil {
		return c.TableauDiskGraph.FlushWritesToDiskFiles()
	}
	if c.DiskGraph != nil {
		return c.DiskGraph.FlushWritesToDiskFiles()
	}
	return nil
}

func (c *LiveChecker) Recover() error {
	if c == nil {
		return nil
	}
	if c.TableauDiskGraph != nil {
		if err := c.TableauDiskGraph.Recover(); err != nil {
			return err
		}
		c.Size = int64(c.TableauDiskGraph.Size())
		return nil
	}
	if c.DiskGraph != nil {
		if err := c.DiskGraph.Recover(); err != nil {
			return err
		}
		c.Size = int64(c.DiskGraph.Size())
	}
	return nil
}

func (c *LiveChecker) PrintCounterExample(tool *Tool) {
	if c == nil || c.ErrorPrinted || len(c.ErrorTrace) == 0 {
		return
	}
	c.ErrorPrinted = true
	rawTrace := c.ErrorRawTrace
	if len(rawTrace) == 0 {
		rawTrace = c.ErrorTrace
	}
	names := LivenessFindViolatedPropertiesFromTrace(tool, rawTrace, c.ErrorCycleIndex)
	PrintError(ECTLCTemporalPropertyViolated, names...)
	PrintError(ECTLCCounterExample)
	for _, info := range c.ErrorTrace {
		PrintInvariantViolationStateTraceState(info)
	}
	if c.ErrorStuttering {
		PrintStutteringState(c.ErrorLoopOrdinal)
		if c.Solution != nil && !c.Solution.HasEmptyPEMAndBoxFreePromises() {
			LivenessPrintStutteringCounterExampleWarning(tool)
		}
	} else {
		PrintBackToState(c.ErrorClosingInfo, c.ErrorLoopOrdinal)
	}
}

func (lc *LiveCheck) Close() error {
	if lc == nil || lc.NoOp {
		return nil
	}
	var err error
	for _, checker := range lc.Checkers {
		if e := checker.Close(); err == nil {
			err = e
		}
	}
	return err
}

func (lc *LiveCheck) BeginChkpt() error {
	if lc == nil || lc.NoOp {
		return nil
	}
	for _, checker := range lc.Checkers {
		if err := checker.BeginChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) CommitChkpt() error {
	if lc == nil || lc.NoOp {
		return nil
	}
	for _, checker := range lc.Checkers {
		if err := checker.CommitChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) FlushWritesToDiskFiles() error {
	if lc == nil || lc.NoOp {
		return nil
	}
	for _, checker := range lc.Checkers {
		if err := checker.FlushWritesToDiskFiles(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) GetOutDegreeStatistics() any {
	if lc == nil || lc.NoOp {
		return nil
	}
	if lc.OutDegreeStats == nil {
		return NewDummyBucketStatistics()
	}
	return lc.OutDegreeStats
}

func (lc *LiveCheck) CalculateInDegreeDiskGraphs(stats any) (any, error) {
	if stats == nil {
		stats = NewBucketStatistics("Histogram vertex in-degree")
	}
	if lc == nil {
		return stats, nil
	}
	if lc.NoOp {
		return lc.OutDegreeStats, nil
	}
	for _, checker := range lc.Checkers {
		var err error
		if checker.TableauDiskGraph != nil {
			stats, err = checker.TableauDiskGraph.CalculateInDegreeDiskGraph(stats)
		} else if checker.DiskGraph != nil {
			stats, err = checker.DiskGraph.CalculateInDegreeDiskGraph(stats)
		}
		if err != nil {
			return stats, err
		}
	}
	return stats, nil
}

func (lc *LiveCheck) CalculateOutDegreeDiskGraphs(stats any) (any, error) {
	if stats == nil {
		stats = NewBucketStatistics("Histogram vertex out-degree")
	}
	if lc == nil {
		return stats, nil
	}
	if lc.NoOp {
		return lc.OutDegreeStats, nil
	}
	for _, checker := range lc.Checkers {
		var err error
		if checker.TableauDiskGraph != nil {
			stats, err = checker.TableauDiskGraph.CalculateOutDegreeDiskGraph(stats)
		} else if checker.DiskGraph != nil {
			stats, err = checker.DiskGraph.CalculateOutDegreeDiskGraph(stats)
		}
		if err != nil {
			return stats, err
		}
	}
	return stats, nil
}

func (lc *LiveCheck) Recover() error {
	if lc == nil || lc.NoOp {
		return nil
	}
	for _, checker := range lc.Checkers {
		PrintMessage(ECTLCAAAAAAA)
		if err := checker.Recover(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) Reset() error {
	if lc != nil {
		for _, checker := range lc.Checkers {
			if checker.TableauDiskGraph != nil {
				if err := checker.TableauDiskGraph.Reset(); err != nil {
					return err
				}
			}
			if checker.DiskGraph != nil {
				if err := checker.DiskGraph.Reset(); err != nil {
					return err
				}
			}
			checker.Reset()
		}
	}
	return nil
}
