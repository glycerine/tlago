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
	throwableTrace
	ErrorCode      int
	hasMessage     bool
	Msg            string
	CounterExample *CounterExample

	ownersCaptured  bool
	errorCodeOwner  int
	hasMessageOwner bool
	msgOwner        string
}

func NewLiveException(errorCode int, msg ...string) *LiveException {
	if len(msg) > 1 {
		panic(NewIllegalArgumentException("NewLiveException accepts at most one message"))
	}
	if len(msg) == 0 {
		return NewLiveExceptionNullable(errorCode, nil)
	}
	return NewLiveExceptionNullable(errorCode, javaString(msg[0]))
}

func NewLiveExceptionNullable(errorCode int, msg *string) *LiveException {
	errorCode = int(int32(errorCode))
	ex := &LiveException{
		throwableTrace: captureThrowableTrace(), ErrorCode: errorCode, hasMessage: msg != nil,
		ownersCaptured: true, errorCodeOwner: errorCode, hasMessageOwner: msg != nil,
	}
	if msg != nil {
		ex.Msg = *msg
		ex.msgOwner = *msg
	}
	return ex
}

func NewLiveExceptionWithCounterExample(errorCode int, msg string, counterExample *CounterExample) *LiveException {
	return NewLiveExceptionNullableWithCounterExample(errorCode, javaString(msg), counterExample)
}

func NewLiveExceptionNullableWithCounterExample(errorCode int, msg *string, counterExample *CounterExample) *LiveException {
	ex := NewLiveExceptionNullable(errorCode, msg)
	ex.CounterExample = counterExample
	return ex
}

func (e *LiveException) Error() string {
	if e == nil {
		return ""
	}
	if message := e.GetMessage(); message != nil {
		return *message
	}
	return javaThrowableClassName(e)
}

func (e *LiveException) GetCause() error {
	e.require()
	return nil
}

type LiveCounterExampleException struct {
	*LiveException
	CounterExample *CounterExample

	ownersCaptured      bool
	liveExceptionOwner  *LiveException
	counterExampleOwner *CounterExample
}

func NewLiveCounterExampleException(errorCode int, msg string, counterExample *CounterExample) *LiveCounterExampleException {
	return NewLiveCounterExampleExceptionNullable(errorCode, javaString(msg), counterExample)
}

func NewLiveCounterExampleExceptionNullable(errorCode int, msg *string, counterExample *CounterExample) *LiveCounterExampleException {
	liveException := NewLiveExceptionNullable(errorCode, msg)
	return &LiveCounterExampleException{
		LiveException: liveException, CounterExample: counterExample,
		ownersCaptured: true, liveExceptionOwner: liveException, counterExampleOwner: counterExample,
	}
}

func (e *LiveCounterExampleException) Error() string {
	if e == nil {
		return ""
	}
	if message := e.GetMessage(); message != nil {
		return *message
	}
	return javaThrowableClassName(e)
}

func (e *LiveCounterExampleException) GetMessage() *string {
	e.require()
	return e.ownedLiveException().GetMessage()
}

func (e *LiveCounterExampleException) GetCause() error {
	e.require()
	return e.ownedLiveException().GetCause()
}

func (e *LiveCounterExampleException) require() {
	if e == nil || e.ownedLiveException() == nil {
		panic(NewNullPointerException())
	}
}

// Go embedding does not implement Java's instanceof inheritance. Match both
// concrete classes without traversing the exception's cause chain.
func javaLiveException(err error) *LiveException {
	switch failure := err.(type) {
	case *LiveCounterExampleException:
		if failure != nil {
			return failure.ownedLiveException()
		}
	case *LiveException:
		return failure
	}
	return nil
}

type LivenessStateWriter struct {
	Noop bool
	*StateWriter
	ownersCaptured   bool
	noOpOwner        bool
	stateWriterOwner *StateWriter
}

var _ IStateWriter = (*LivenessStateWriter)(nil)

func NewNoopLivenessStateWriter() *LivenessStateWriter {
	return &LivenessStateWriter{Noop: true, ownersCaptured: true, noOpOwner: true}
}

func NewDotLivenessStateWriter(stateWriter IStateWriter) (*LivenessStateWriter, error) {
	if isNil(stateWriter) {
		panic(NewNullPointerException())
	}
	fname := dotWriterJavaReplace(stateWriter.GetDumpFileName(), "_liveness.dot")
	writer, err := NewDotStateWriter(fname, DotStateWriterOptions{})
	if err != nil {
		return nil, err
	}
	return &LivenessStateWriter{StateWriter: writer, ownersCaptured: true, stateWriterOwner: writer}, nil
}

func (w *LivenessStateWriter) IsNoop() bool {
	w.require()
	writer := w.ownedStateWriter()
	return w.ownedNoOp() || writer == nil || writer.IsNoop()
}

func (w *LivenessStateWriter) IsDot() bool {
	w.require()
	if w.ownedNoOp() {
		return false
	}
	writer := w.ownedStateWriter()
	return writer != nil && writer.IsDot()
}

func (w *LivenessStateWriter) IsConstrained() bool {
	w.require()
	if w.ownedNoOp() {
		return false
	}
	writer := w.ownedStateWriter()
	return writer != nil && writer.IsConstrained()
}

func (w *LivenessStateWriter) GetDumpFileName() string {
	w.require()
	writer := w.ownedStateWriter()
	if w.ownedNoOp() || writer == nil {
		return ""
	}
	return writer.GetDumpFileName()
}

func (w *LivenessStateWriter) Snapshot() error {
	w.require()
	writer := w.ownedStateWriter()
	if w.ownedNoOp() || writer == nil {
		return nil
	}
	return writer.Snapshot()
}

func (w *LivenessStateWriter) Close() error {
	w.require()
	writer := w.ownedStateWriter()
	if w.ownedNoOp() || writer == nil {
		return nil
	}
	return writer.Close()
}

func (w *LivenessStateWriter) WriteInitState(state *TLCStateMut) error {
	w.require()
	writer := w.ownedStateWriter()
	if w.IsNoop() || writer == nil {
		return nil
	}
	return writer.WriteInitState(state)
}

func (w *LivenessStateWriter) WriteTransition(state *TLCStateMut, successor *TLCStateMut, status StateVisitStatus, action *Action, reason ...SemanticNode) error {
	w.require()
	if len(reason) > 1 {
		panic(NewIllegalArgumentException("WriteTransition accepts at most one predicate"))
	}
	return w.WriteTransitionVisual(state, successor, status, action, StateVisualizationDefault, reason...)
}

func (w *LivenessStateWriter) WriteTransitionVisual(state *TLCStateMut, successor *TLCStateMut, status StateVisitStatus, action *Action, visualization StateVisualization, reason ...SemanticNode) error {
	w.require()
	if len(reason) > 1 {
		panic(NewIllegalArgumentException("WriteTransitionVisual accepts at most one predicate"))
	}
	writer := w.ownedStateWriter()
	if w.IsNoop() || writer == nil {
		return nil
	}
	return writer.WriteTransitionVisual(state, successor, status, action, visualization, reason...)
}

func (w *LivenessStateWriter) WriteLivenessState(state *TLCStateMut) error {
	return w.WriteInitState(state)
}

func (w *LivenessStateWriter) WriteLivenessStateTransition(state *TLCStateMut, successor *TLCStateMut, actionChecks *BitVector, from int, length int, status StateVisitStatus, visualization StateVisualization) error {
	// The source inherits DotStateWriter's non-tableau overload, which does
	// not render action checks and maintains the ordinary state DOT ranks.
	return w.WriteTransitionVisual(state, successor, status, nil, visualization)
}

func (w *LivenessStateWriter) WriteLivenessInitState(state *TLCStateMut, tableauNode *TBGraphNode) error {
	w.require()
	if w.IsNoop() {
		return nil
	}
	if w.ownedStateWriter().writer == nil {
		panic(NewNullPointerException())
	}
	_, _ = w.ownedStateWriter().writer.WriteString("\"")
	_, _ = w.ownedStateWriter().writer.WriteString(fmtInt64(int64(state.FingerPrint())))
	_, _ = w.ownedStateWriter().writer.WriteString(".")
	_, _ = w.ownedStateWriter().writer.WriteString(fmtInt(tableauNode.GetIndex()))
	_, _ = w.ownedStateWriter().writer.WriteString("\" [style = filled] [label=\"")
	_, _ = w.ownedStateWriter().writer.WriteString(stateToDot(state, nil, false))
	_, _ = w.ownedStateWriter().writer.WriteString("\n#" + fmtInt64(int64(state.FingerPrint())) + "." + fmtInt(tableauNode.GetIndex()) + "#")
	_, _ = w.ownedStateWriter().writer.WriteString("\"]\n")
	return nil
}

func (w *LivenessStateWriter) WriteLivenessTransition(state *TLCStateMut, tableauNode *TBGraphNode, successor *TLCStateMut, successorTableauNode *TBGraphNode, actionChecks *BitVector, from int, length int, status StateVisitStatus) error {
	return w.WriteLivenessTransitionVisual(state, tableauNode, successor, successorTableauNode, actionChecks, from, length, status, StateVisualizationDefault)
}

func (w *LivenessStateWriter) WriteLivenessTransitionVisual(state *TLCStateMut, tableauNode *TBGraphNode, successor *TLCStateMut, successorTableauNode *TBGraphNode, actionChecks *BitVector, from int, length int, status StateVisitStatus, visualization StateVisualization) error {
	w.require()
	if w.IsNoop() {
		return nil
	}
	if w.ownedStateWriter().writer == nil {
		panic(NewNullPointerException())
	}
	successorFP := fmtInt64(int64(successor.FingerPrint()))
	_, _ = w.ownedStateWriter().writer.WriteString("\"")
	_, _ = w.ownedStateWriter().writer.WriteString(fmtInt64(int64(state.FingerPrint())))
	_, _ = w.ownedStateWriter().writer.WriteString(".")
	_, _ = w.ownedStateWriter().writer.WriteString(fmtInt(tableauNode.GetIndex()))
	_, _ = w.ownedStateWriter().writer.WriteString("\" -> \"")
	_, _ = w.ownedStateWriter().writer.WriteString(successorFP)
	_, _ = w.ownedStateWriter().writer.WriteString(".")
	_, _ = w.ownedStateWriter().writer.WriteString(fmtInt(successorTableauNode.GetIndex()))
	_, _ = w.ownedStateWriter().writer.WriteString("\"")
	switch visualization {
	case StateVisualizationStuttering:
		_, _ = w.ownedStateWriter().writer.WriteString(" [style=\"dashed\"]")
	case StateVisualizationDotted:
		_, _ = w.ownedStateWriter().writer.WriteString(" [style=\"dotted\"]")
	}
	if length > 0 {
		label := actionChecks.StringRangeChars(from, length, 't', 'f')
		_, _ = w.ownedStateWriter().writer.WriteString(" [label=\"" + dotEscape(label) + "\"]")
	}
	_, _ = w.ownedStateWriter().writer.WriteString(";\n")
	if !stateVisitStatusHas(status, StateVisitSeen) {
		_, _ = w.ownedStateWriter().writer.WriteString("\"" + successorFP + "." + fmtInt(successorTableauNode.GetIndex()) + "\" [label=\"")
		_, _ = w.ownedStateWriter().writer.WriteString(stateToDot(successor, nil, false))
		_, _ = w.ownedStateWriter().writer.WriteString("\n#" + fmtInt64(int64(successor.FingerPrint())) + "." + fmtInt(successorTableauNode.GetIndex()) + "#\"];")
		_, _ = w.ownedStateWriter().writer.WriteString("\n")
	}
	return nil
}

func (w *LivenessStateWriter) ownedNoOp() bool {
	if w.ownersCaptured {
		return w.noOpOwner
	}
	return w.Noop
}

func (w *LivenessStateWriter) ownedStateWriter() *StateWriter {
	if w.ownersCaptured {
		return w.stateWriterOwner
	}
	return w.StateWriter
}

func (w *LivenessStateWriter) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

type LiveChecker struct {
	Solution         *OrderOfSolution
	Soln             int
	Writer           *LivenessStateWriter
	tableau          bool
	ownersCaptured   bool
	solution         *OrderOfSolution
	writer           *LivenessStateWriter
	diskGraph        *DiskGraph
	tableauDiskGraph *TableauDiskGraph
	errOwner         error
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
	soln = int(int32(soln))
	if writer == nil {
		writer = NewNoopLivenessStateWriter()
	}
	checker := &LiveChecker{
		Solution:       solution,
		Soln:           soln,
		Writer:         writer,
		ownersCaptured: true,
		solution:       solution,
		writer:         writer,
	}
	if solution != nil {
		checker.tableau = solution.HasTableau()
	}
	if metadir != "" && solution != nil {
		if checker.tableau {
			if liveCheckDebugEnabled() {
				checker.TableauDiskGraph, checker.Err = NewDebugTableauDiskGraph(metadir, soln, outDegreeStats, solution)
			} else {
				checker.TableauDiskGraph, checker.Err = NewTableauDiskGraph(metadir, soln, outDegreeStats)
			}
			checker.tableauDiskGraph = checker.TableauDiskGraph
		} else {
			checker.DiskGraph, checker.Err = NewDiskGraph(metadir, soln, outDegreeStats)
			checker.diskGraph = checker.DiskGraph
		}
	}
	checker.errOwner = checker.Err
	return checker
}

func (c *LiveChecker) GetSolution() *OrderOfSolution {
	c.require()
	return c.ownedSolution()
}

func (c *LiveChecker) ownedSolution() *OrderOfSolution {
	if c.ownersCaptured {
		return c.solution
	}
	return c.Solution
}

func (c *LiveChecker) ownedWriter() *LivenessStateWriter {
	if c.ownersCaptured {
		return c.writer
	}
	return c.Writer
}

func (c *LiveChecker) ownedDiskGraph() *DiskGraph {
	if c.ownersCaptured {
		return c.diskGraph
	}
	return c.DiskGraph
}

func (c *LiveChecker) ownedTableauDiskGraph() *TableauDiskGraph {
	if c.ownersCaptured {
		return c.tableauDiskGraph
	}
	return c.TableauDiskGraph
}

func (c *LiveChecker) ownedErr() error {
	if c.ownersCaptured {
		return c.errOwner
	}
	return c.Err
}

func (c *LiveChecker) GetDiskGraph() any {
	c.require()
	if c.tableau {
		graph := c.ownedTableauDiskGraph()
		if graph == nil {
			return nil
		}
		return graph
	}
	graph := c.ownedDiskGraph()
	if graph == nil {
		return nil
	}
	return graph
}

func liveCheckDebugEnabled() bool {
	value, ok := tlcLookupSystemProperty(liveCheckDebugProperty)
	return ok && javaBooleanProperty(value)
}

func (c *LiveChecker) AddInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	c.require()
	oos := c.ownedSolution()
	if oos == nil {
		panic(NewNullPointerException())
	}
	if c.ownedErr() != nil {
		return c.ownedErr()
	}
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		tableau := oos.GetTableau()
		for i := 0; i < tableau.GetInitCnt(); i++ {
			tnode := tableau.GetNode(i)
			ok, err := tnode.IsConsistent(state, tool)
			if err != nil {
				return err
			}
			if ok {
				if dgraph == nil {
					panic(NewNullPointerException())
				}
				dgraph.AddInitNode(stateFP, tnode.GetIndex())
				dgraph.RecordNode(stateFP, tnode.GetIndex())
				if err := c.ownedWriter().WriteLivenessInitState(state, tnode); err != nil {
					return err
				}
			}
		}
		return nil
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	dgraph.AddInitNode(stateFP, -1)
	if err := c.ownedWriter().WriteLivenessState(state); err != nil {
		return err
	}
	return nil
}

func (c *LiveChecker) AddNextState(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
	c.require()
	if nextStates == nil {
		panic(NewNullPointerException())
	}
	oos := c.ownedSolution()
	if oos == nil {
		panic(NewNullPointerException())
	}
	if c.ownedErr() != nil {
		return c.ownedErr()
	}
	if c.tableau {
		return c.addNextStateTableau(tool, s0, fp0, nextStates, actionResults, checkStateRes)
	}
	oos.graphMu.Lock()
	defer oos.graphMu.Unlock()
	if c.ownedDiskGraph() == nil {
		panic(NewNullPointerException())
	}
	return c.addNextStateDisk(s0, fp0, nextStates, actionResults, checkStateRes)
}

func (c *LiveChecker) addNextStateDisk(s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	cnt := 0
	succCnt := nextStates.Size()
	alen := len(c.ownedSolution().GetCheckAction())
	node0, err := dgraph.GetNodeNoTableau(fp0)
	if err != nil {
		return err
	}
	originalSuccSize := node0.SuccSize()
	node0.SetCheckState(checkStateRes)
	nextStates.ResetNext()
	for sidx := 0; sidx < succCnt; sidx++ {
		successorState := nextStates.Next().(*TLCStateMut)
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
		if err := c.ownedWriter().WriteLivenessStateTransition(s0, successorState, actionResults, sidx*alen, alen, status, StateVisualizationDefault); err != nil {
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
	} else if MainChecker() != nil {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	return nil
}

func (c *LiveChecker) addNextStateTableau(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
	oos := c.ownedSolution()
	succCnt := nextStates.Size()
	tableau := oos.GetTableau()
	consistency := NewBitVector(tableau.Size() * succCnt)
	for _, tableauNode := range tableau.ownedNodes() {
		nextStates.ResetNext()
		for sidx := 0; sidx < succCnt; sidx++ {
			s1 := nextStates.Next().(*TLCStateMut)
			if s1 == nil {
				continue
			}
			ok, err := tableauNode.IsConsistent(s1, tool)
			if err != nil {
				return err
			}
			if ok {
				consistency.Set(tableauNode.GetIndex()*succCnt + sidx)
			}
		}
	}

	nextStates.ResetNext()

	var prefix *LongVec
	err := func() error {
		oos.graphMu.Lock()
		defer oos.graphMu.Unlock()
		if c.ownedTableauDiskGraph() == nil {
			panic(NewNullPointerException())
		}
		var err error
		prefix, err = c.addNextStateTableauDisk(tool, s0, fp0, nextStates, actionResults, checkStateRes, consistency)
		return err
	}()
	if err != nil {
		return err
	}
	// Java reconstructs the fingerprint prefix under the solution monitor,
	// then regenerates and prints the states after releasing it.
	if prefix != nil {
		return c.printSafetyLikeLivenessError(tool, prefix)
	}
	return nil
}

func (c *LiveChecker) addNextStateTableauDisk(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool, consistency *BitVector) (*LongVec, error) {
	dgraph := c.ownedTableauDiskGraph()
	oos := c.ownedSolution()
	if dgraph == nil || oos == nil || oos.GetTableau() == nil {
		panic(NewNullPointerException())
	}
	cnt := 0
	succCnt := nextStates.Size()

	loc0 := dgraph.SetDone(fp0)
	nodes := dgraph.GetNodesByLoc(loc0)
	if nodes == nil {
		return nil, nil
	}
	alen := len(oos.GetCheckAction())
	allocationHint := (len(nodes) / dgraph.GetElemLength()) * succCnt
	for nidx := 2; nidx < len(nodes); nidx += dgraph.GetElemLength() {
		tidx0 := int(nodes[nidx])
		tnode0 := oos.GetTableau().GetNode(tidx0)
		node0, err := dgraph.GetNode(fp0, tidx0)
		if err != nil {
			return nil, err
		}
		originalSuccSize := node0.SuccSize()
		node0.SetCheckState(checkStateRes)
		nextStates.ResetNext()
		for sidx := 0; sidx < succCnt; sidx++ {
			s1 := nextStates.Next().(*TLCStateMut)
			if s1 == nil {
				continue
			}
			successor := s1.FingerPrint()
			isDone := dgraph.IsDone(successor)
			for _, tnode1 := range tnode0.ownedNexts() {
				ptr1 := dgraph.GetPtr(successor, tnode1.GetIndex())
				if consistency.Get(tnode1.GetIndex()*succCnt+sidx) && (ptr1 == -1 || !node0.TransExists(successor, tnode1.GetIndex())) {
					node0.AddTransition(successor, tnode1.GetIndex(), len(checkStateRes), alen, actionResults, sidx*alen, allocationHint-cnt)
					if err := c.ownedWriter().WriteLivenessTransition(s0, tnode0, s1, tnode1, actionResults, sidx*alen, alen, StateVisitUnseen); err != nil {
						return nil, err
					}
					if ptr1 == -1 {
						dgraph.RecordNode(successor, tnode1.GetIndex())
						if isDone {
							if err := c.addNextStateTableauDone(tool, s1, successor, tnode1); err != nil {
								return nil, err
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
				return nil, err
			}
			c.Size = int64(dgraph.Size())
		} else if MainChecker() != nil {
			panic(NewTLCRuntimeException(ECGeneral))
		}
	}
	if c.ErrorGraphNode != nil {
		dgraph.CreateCache()
		prefix, err := dgraph.GetPath(c.ErrorGraphNode.ownedStateFP(), c.ErrorGraphNode.ownedTIndex())
		if err != nil {
			return nil, err
		}
		dgraph.DestroyCache()
		c.ErrorPrefix = prefix
		return prefix, nil
	}
	return nil, nil
}

func (c *LiveChecker) printSafetyLikeLivenessError(tool *Tool, prefix *LongVec) error {
	if c == nil {
		panic(NewNullPointerException())
	}
	mc := MainChecker()
	if mc == nil || mc.abstractChecker() == nil {
		panic(NewNullPointerException())
	}
	checker := mc.abstractChecker()
	checker.mu.Lock()
	defer checker.mu.Unlock()
	if checker.PrintedLivenessErrorStack {
		return nil
	}
	checker.PrintedLivenessErrorStack = true

	trace, err := c.reconstructSafetyLikeLivenessPrefix(tool, prefix)
	if err != nil {
		return err
	}
	if len(trace) == 0 {
		return nil
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

	last := trace[len(trace)-1].GetOriginalState()
	var pred *TLCStateMut
	if len(trace) > 1 {
		pred = trace[len(trace)-2].GetOriginalState()
	} else {
		pred = last
		last = nil
	}
	mc.Stop()
	mc.SetErrState(pred, last, false, ECTLCInvariantViolatedBehavior)
	tool.CheckPostConditionWithCounterExample(NewCounterExampleFromTrace(trace))
	c.ErrorGraphNode = nil
	return NewInvariantViolatedException()
}

func (c *LiveChecker) reconstructSafetyLikeLivenessPrefix(tool *Tool, prefix *LongVec) ([]*TLCStateInfo, error) {
	plen := prefix.Size()
	fp := uint64(prefix.ElementAt(plen - 1))
	info, err := tool.GetState(fp)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, NewEvalException(ECTLCFailedToRecoverInit)
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
			panic(NewNullPointerException())
		}
		alias, err := tool.EvalAliasInfoPair(trace[len(trace)-1], next.State)
		if err != nil {
			return nil, err
		}
		trace[len(trace)-1] = alias
		trace = append(trace, next)
		info = next
		fp = curFP
	}
	last := trace[len(trace)-1]
	alias, err := tool.EvalAliasInfoPair(last, last.State)
	if err != nil {
		return nil, err
	}
	trace[len(trace)-1] = alias
	return trace, nil
}

func (c *LiveChecker) addNextStateTableauDone(tool *Tool, state *TLCStateMut, fp uint64, tnode *TBGraphNode) error {
	dgraph := c.ownedTableauDiskGraph()
	oos := c.ownedSolution()
	if oos == nil {
		panic(NewNullPointerException())
	}
	checkStateRes, err := oos.CheckState(tool, state)
	if err != nil {
		return err
	}
	slen := len(checkStateRes)
	alen := len(oos.GetCheckAction())
	if tnode == nil || dgraph == nil {
		panic(NewNullPointerException())
	}
	node, err := dgraph.GetNode(fp, tnode.GetIndex())
	if err != nil {
		return err
	}
	numSucc := node.SuccSize()
	node.SetCheckState(checkStateRes)

	cnt := 0
	nextSize := len(tnode.ownedNexts())
	var selfLoopActionResults *BitVector
	if nextSize > 0 {
		selfLoopActionResults, err = oos.CheckActionBitVector(tool, state, state, NewBitVector(alen), 0)
		if err != nil {
			return err
		}
	}
	for _, tnode1 := range tnode.ownedNexts() {
		tidx1 := tnode1.GetIndex()
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
				for _, tnode1 := range tnode.ownedNexts() {
					tidx1 := tnode1.GetIndex()
					ptr1 := dgraph.GetPtr(fp1, tidx1)
					total := len(actions) * nextCnt * len(tnode.ownedNexts())
					ok, err := tnode1.IsConsistent(s1, tool)
					if err != nil {
						return err
					}
					if ok && (ptr1 == -1 || !node.TransExists(fp1, tidx1)) {
						node.AddTransition(fp1, tidx1, slen, alen, checkActionRes, 0, total-cnt)
						if err := c.ownedWriter().WriteLivenessTransitionVisual(state, tnode, s1, tnode1, checkActionRes, 0, alen, StateVisitSeen, StateVisualizationDotted); err != nil {
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

func (c *LiveChecker) CheckSccs(tool *Tool, finalCheck bool, liveCheck *LiveCheck, workerID int, numWorkers int) (bool, error) {
	c.require()
	if (c.tableau && c.ownedTableauDiskGraph() == nil) || (!c.tableau && c.ownedDiskGraph() == nil) {
		panic(NewNullPointerException())
	}
	c.createDiskGraphCache()
	oos := c.ownedSolution()
	if oos == nil {
		panic(NewNullPointerException())
	}
	for _, pem := range oos.GetPems() {
		if liveCheck != nil && liveCheck.hasLiveError() {
			break
		}
		worker := NewLiveWorker(tool, workerID, numWorkers, liveCheck, c, pem, finalCheck)
		found, err := worker.CheckSccs()
		if err != nil {
			return false, err
		}
		if found {
			c.printCounterExampleAndCheckPostCondition(tool)
			c.destroyDiskGraphCache()
			c.recordGraphSize()
			if err := c.assertDiskGraphInvariants(); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	c.destroyDiskGraphCache()
	c.recordGraphSize()
	if err := c.assertDiskGraphInvariants(); err != nil {
		return false, err
	}
	return false, nil
}

func (c *LiveChecker) Reset() {
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
	Tool             *Tool
	MetaDir          string
	Checkers         []*LiveChecker
	OutDegreeStats   any
	NoOp             bool
	noOpTool         *Tool
	noOpMetaDir      string
	noOpStats        any
	ownersCaptured   bool
	metaDirOwner     string
	checkersOwner    []*LiveChecker
	statsOwner       any
	subtypeCaptured  bool
	noOpOwner        bool
	addAndCheckOwner bool
	AddAndCheck      bool
	mu               sync.Mutex
	errMu            sync.Mutex
	errFoundByID     int
}

func NewNoOpLiveCheck(tool *Tool, metadir string) *LiveCheck {
	stats := NewDummyBucketStatistics()
	return &LiveCheck{
		Tool: tool, MetaDir: metadir, OutDegreeStats: stats, NoOp: true,
		noOpTool: tool, noOpMetaDir: metadir, noOpStats: stats,
		subtypeCaptured: true, noOpOwner: true, errFoundByID: -1,
	}
}

func NewLiveCheck(tool *Tool, solutions []*OrderOfSolution, metadir string) *LiveCheck {
	check, err := newLiveCheckWithStateWriter(tool, solutions, metadir, NewNoopStateWriter())
	if err != nil {
		panic(err)
	}
	return check
}

func NewLiveCheckWithStateWriter(tool *Tool, solutions []*OrderOfSolution, metadir string, stateWriter IStateWriter) (*LiveCheck, error) {
	if solutions == nil {
		panic(NewNullPointerException())
	}
	return newLiveCheckWithStateWriter(tool, solutions, metadir, stateWriter)
}

func newLiveCheckWithStateWriter(tool *Tool, solutions []*OrderOfSolution, metadir string, stateWriter IStateWriter) (*LiveCheck, error) {
	if solutions == nil {
		panic(NewNullPointerException())
	}
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
		if isNil(stateWriter) {
			panic(NewNullPointerException())
		}
		writer := NewNoopLivenessStateWriter()
		if !stateWriter.IsNoop() && stateWriter.IsDot() {
			dotWriter, err := NewDotLivenessStateWriter(stateWriter)
			if err != nil {
				return nil, err
			}
			writer = dotWriter
		}
		if solution == nil {
			panic(NewNullPointerException())
		}
		checker := NewLiveChecker(solution, i, writer, metadir, check.OutDegreeStats)
		if checker.Err != nil {
			return nil, checker.Err
		}
		check.Checkers = append(check.Checkers, checker)
	}
	check.ownersCaptured = true
	check.metaDirOwner = metadir
	check.checkersOwner = append([]*LiveChecker(nil), check.Checkers...)
	check.statsOwner = outDegreeStats
	check.subtypeCaptured = true
	return check, nil
}

func NewAddAndCheckLiveCheck(tool *Tool, solutions []*OrderOfSolution, metadir string) *LiveCheck {
	if tool == nil {
		panic(NewNullPointerException())
	}
	check, err := newAddAndCheckLiveCheck(tool.NoDebug(), solutions, metadir)
	if err != nil {
		panic(err)
	}
	return check
}

func newAddAndCheckLiveCheck(tool *Tool, solutions []*OrderOfSolution, metadir string) (*LiveCheck, error) {
	check, err := newLiveCheckWithStateWriter(tool, solutions, metadir, NewNoopStateWriter())
	if err != nil {
		return nil, err
	}
	check.AddAndCheck = true
	check.addAndCheckOwner = true
	PrintWarning(ECUnitTest, "!!!WARNING: TLC is running in inefficient unit testing mode!!!", "")
	return check, nil
}

func (lc *LiveCheck) GetNumChecker() int {
	lc.require()
	if lc.isNoOp() {
		return 0
	}
	return len(lc.ownedCheckers())
}

func (lc *LiveCheck) NumChecker() int {
	return lc.GetNumChecker()
}

func (lc *LiveCheck) ownedCheckers() []*LiveChecker {
	if lc.ownersCaptured {
		return lc.checkersOwner
	}
	return lc.Checkers
}

func (lc *LiveCheck) ownedMetaDir() string {
	if lc.ownersCaptured {
		return lc.metaDirOwner
	}
	return lc.MetaDir
}

func (lc *LiveCheck) ownedOutDegreeStats() any {
	if lc.ownersCaptured {
		return lc.statsOwner
	}
	return lc.OutDegreeStats
}

func (lc *LiveCheck) isNoOp() bool {
	if lc.subtypeCaptured {
		return lc.noOpOwner
	}
	return lc.NoOp
}

func (lc *LiveCheck) isAddAndCheck() bool {
	if lc.subtypeCaptured {
		return lc.addAndCheckOwner
	}
	return lc.AddAndCheck
}

func (lc *LiveCheck) GetMetaDir() string {
	lc.require()
	if lc.isNoOp() {
		return lc.noOpMetaDir
	}
	return lc.ownedMetaDir()
}

func (lc *LiveCheck) GetTool() *Tool {
	lc.require()
	if lc.isNoOp() {
		return lc.noOpTool
	}
	return lc.Tool
}

func (lc *LiveCheck) GetChecker(idx int) *LiveChecker {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	idx = int(int32(idx))
	if idx < 0 || idx >= len(lc.ownedCheckers()) {
		panic(NewArrayIndexOutOfBoundsException(idx, len(lc.ownedCheckers())))
	}
	return lc.ownedCheckers()[idx]
}

func (lc *LiveCheck) AddInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	if lc.isAddAndCheck() {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		if err := lc.addInitState(tool, state, stateFP); err != nil {
			return err
		}
		_, err := lc.check0(tool, false)
		return addAndCheckInitCheckResult(err)
	}
	return lc.addInitState(tool, state, stateFP)
}

func (lc *LiveCheck) addInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	for _, checker := range lc.ownedCheckers() {
		if err := checker.AddInitState(tool, state, stateFP); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) AddNextState(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates) error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	if lc.isAddAndCheck() {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		if err := lc.addNextState(tool, s0, fp0, nextStates); err != nil {
			return err
		}
		_, err := lc.check0(tool, false)
		return addAndCheckNextCheckResult(err)
	}
	return lc.addNextState(tool, s0, fp0, nextStates)
}

// AddAndCheckLiveCheck has intentionally asymmetric source exception handling:
// addInitState cannot throw and prints both checked check0 failures, while
// addNextState declares IOException and catches only interruption.
func addAndCheckInitCheckResult(err error) error {
	if err == nil {
		return nil
	}
	if _, interrupted := err.(*InterruptedException); interrupted || isJavaIOException(err) {
		_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
		return nil
	}
	return err
}

func addAndCheckNextCheckResult(err error) error {
	if err == nil {
		return nil
	}
	if _, interrupted := err.(*InterruptedException); interrupted {
		_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
		return nil
	}
	return err
}

func (lc *LiveCheck) addNextState(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates) error {
	for _, checker := range lc.ownedCheckers() {
		oos := checker.ownedSolution()
		alen := len(oos.GetCheckAction())
		actionResults := NewBitVector(alen * nextStates.Size())
		nextStates.ResetNext()
		for sidx := 0; sidx < nextStates.Size(); sidx++ {
			s1 := nextStates.Next().(*TLCStateMut)
			if s1 == nil {
				continue
			}
			if _, err := oos.CheckActionBitVector(tool, s0, s1, actionResults, alen*sidx); err != nil {
				return err
			}
		}
		nextStates.ResetNext()
		checkState, err := oos.CheckState(tool, s0)
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
	lc.require()
	if lc.isNoOp() {
		return false
	}
	threshold := 0.0
	Globals.Lock()
	threshold = Globals.LivenessThreshold
	Globals.Unlock()
	for _, checker := range lc.ownedCheckers() {
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
	lc.require()
	if lc.isNoOp() {
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
	lc.require()
	if lc.isNoOp() {
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

type liveCheckResult struct {
	found   bool
	checker *LiveChecker
	err     error
}

func collectLiveCheckResults(results <-chan liveCheckResult, count int) (*LiveChecker, error) {
	var foundChecker *LiveChecker
	var lastErr error
	for i := 0; i < count; i++ {
		result := <-results
		if result.found && foundChecker == nil {
			foundChecker = result.checker
		}
		if result.err != nil {
			lastErr = result.err
		}
	}
	return foundChecker, lastErr
}

func restoreLiveCheckNodePtrTables(checkers []*LiveChecker) error {
	for _, checker := range checkers {
		checker.require()
		if checker.tableau {
			dgraph := checker.ownedTableauDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			if err := dgraph.MakeNodePtrTbl(); err != nil {
				return err
			}
		} else {
			dgraph := checker.ownedDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			if err := dgraph.MakeNodePtrTbl(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (lc *LiveCheck) check0(tool *Tool, finalCheck bool) (int, error) {
	startMillis := time.Now().UnixMilli()
	sum := int64(0)
	for _, checker := range lc.ownedCheckers() {
		sum += checker.GraphSize()
	}
	branches := ""
	if len(lc.ownedCheckers()) != 1 {
		branches = fmt.Sprintf("%d branches of ", len(lc.ownedCheckers()))
	}
	space := "current"
	if finalCheck {
		space = "complete"
	}
	PrintMessage(ECTLCCheckingTemporalProps, space, fmtInt64(sum), branches)
	if len(lc.ownedCheckers()) == 0 {
		panic(NewIllegalArgumentException())
	}
	lc.resetLiveError()

	wNum := 1
	if !DoSequentialLiveness() {
		wNum = min(len(lc.ownedCheckers()), NumWorkers())
		if wNum < 1 {
			panic(NewIllegalArgumentException())
		}
	}
	queue := make(chan *LiveChecker, len(lc.ownedCheckers()))
	for _, checker := range lc.ownedCheckers() {
		queue <- checker
	}
	close(queue)
	results := make(chan liveCheckResult, wNum)
	var workers sync.WaitGroup
	workers.Add(wNum)
	for id := 0; id < wNum; id++ {
		go func(workerID int) {
			defer workers.Done()
			var result liveCheckResult
			defer func() {
				if recovered := recover(); recovered != nil {
					result.err = panicValueAsError(recovered)
				}
				results <- result
			}()
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
		}(id)
	}

	workers.Wait()
	foundChecker, lastErr := collectLiveCheckResults(results, wNum)
	if foundChecker != nil {
		PrintMessage(ECTLCCheckingTemporalPropsEnd, ConvertRuntimeToHumanReadable(time.Now().UnixMilli()-startMillis))
		// Java check0 reports a violation as a result code. AddAndCheckLiveCheck
		// deliberately ignores that code during insertion and keeps building the
		// graph; throwing here would turn that path into an unexpected failure.
		return ECTLCTemporalPropertyViolated, nil
	}
	if lastErr != nil {
		return printLivenessWorkerFailure(lastErr), lastErr
	}
	if !finalCheck {
		if err := restoreLiveCheckNodePtrTables(lc.ownedCheckers()); err != nil {
			return ECGeneral, err
		}
	}
	PrintMessage(ECTLCCheckingTemporalPropsEnd, ConvertRuntimeToHumanReadable(time.Now().UnixMilli()-startMillis))
	return NoError, nil
}

func printLivenessWorkerFailure(err error) int {
	if err == nil {
		PrintError(ECGeneral)
		return ECGeneral
	}
	// err is already the worker failure, corresponding to the single
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
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	if trace == nil {
		panic(NewNullPointerException())
	}
	states := trace()
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
	c.require()
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return int64(dgraph.Size())
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return int64(dgraph.Size())
}

func (c *LiveChecker) SizeAtLastCheck() int64 {
	c.require()
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.GetSizeAtLastCheck()
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.GetSizeAtLastCheck()
}

func (c *LiveChecker) recordGraphSize() {
	if c == nil {
		return
	}
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		dgraph.RecordSize()
		c.LastSize = c.GraphSize()
		return
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	dgraph.RecordSize()
	c.LastSize = c.GraphSize()
}

func (c *LiveChecker) assertDiskGraphInvariants() error {
	oos := c.ownedSolution()
	if oos == nil {
		panic(NewNullPointerException())
	}
	slen := len(oos.GetCheckState())
	alen := len(oos.GetCheckAction())
	var (
		ok  bool
		err error
	)
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		ok, err = dgraph.CheckInvariants(slen, alen)
	} else {
		dgraph := c.ownedDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		ok, err = dgraph.CheckInvariants(slen, alen)
	}
	if err != nil {
		return err
	}
	if !ok {
		panic(NewAssertionError())
	}
	return nil
}

func (c *LiveChecker) createDiskGraphCache() {
	if c == nil {
		return
	}
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		dgraph.CreateCache()
		return
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	dgraph.CreateCache()
}

func (c *LiveChecker) destroyDiskGraphCache() {
	if c == nil {
		return
	}
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		dgraph.DestroyCache()
		return
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	dgraph.DestroyCache()
}

func (c *LiveChecker) Close() error {
	c.require()
	// Java AbstractLiveChecker.close retains the graph when VETO_CLEANUP
	// is set, so callers can inspect it after model checking finishes.
	if !modelCheckerVetoCleanup() {
		if c.tableau {
			dgraph := c.ownedTableauDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			if err := dgraph.Close(); err != nil {
				return err
			}
		} else {
			dgraph := c.ownedDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			if err := dgraph.Close(); err != nil {
				return err
			}
		}
	}
	writer := c.ownedWriter()
	if writer == nil {
		panic(NewNullPointerException())
	}
	return writer.Close()
}

func (c *LiveChecker) require() {
	if c == nil {
		panic(NewNullPointerException())
	}
}

func (c *LiveChecker) BeginChkpt() error {
	c.require()
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.BeginChkpt()
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.BeginChkpt()
}

func (c *LiveChecker) CommitChkpt() error {
	c.require()
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.CommitChkpt()
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.CommitChkpt()
}

func (c *LiveChecker) FlushWritesToDiskFiles() error {
	c.require()
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.FlushWritesToDiskFiles()
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.FlushWritesToDiskFiles()
}

func (c *LiveChecker) Recover() error {
	c.require()
	if c.tableau {
		dgraph := c.ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		if err := dgraph.Recover(); err != nil {
			return err
		}
		c.Size = int64(dgraph.Size())
		return nil
	}
	dgraph := c.ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	if err := dgraph.Recover(); err != nil {
		return err
	}
	c.Size = int64(dgraph.Size())
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
		if solution := c.ownedSolution(); solution != nil && !solution.HasEmptyPEMAndBoxFreePromises() {
			LivenessPrintStutteringCounterExampleWarning(tool)
		}
	} else {
		PrintBackToState(c.ErrorClosingInfo, c.ErrorLoopOrdinal)
	}
}

func (c *LiveChecker) printCounterExampleAndCheckPostCondition(tool *Tool) {
	c.PrintCounterExample(tool)
	tool.CheckPostConditionWithCounterExample(c.ErrorCounterEx)
}

func (lc *LiveCheck) Close() error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	for _, checker := range lc.ownedCheckers() {
		if err := checker.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) BeginChkpt() error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	lc.mu.Lock()
	defer lc.mu.Unlock()
	for _, checker := range lc.ownedCheckers() {
		if err := checker.BeginChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) CommitChkpt() error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	for _, checker := range lc.ownedCheckers() {
		if err := checker.CommitChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) FlushWritesToDiskFiles() error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	for _, checker := range lc.ownedCheckers() {
		if err := checker.FlushWritesToDiskFiles(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) GetOutDegreeStatistics() any {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	return lc.ownedOutDegreeStats()
}

func (lc *LiveCheck) CalculateInDegreeDiskGraphs(stats any) (any, error) {
	lc.require()
	if lc.isNoOp() {
		return lc.noOpStats, nil
	}
	for _, checker := range lc.ownedCheckers() {
		checker.require()
		var err error
		if checker.tableau {
			dgraph := checker.ownedTableauDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			stats, err = dgraph.CalculateInDegreeDiskGraph(stats)
		} else {
			dgraph := checker.ownedDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			stats, err = dgraph.CalculateInDegreeDiskGraph(stats)
		}
		if err != nil {
			return stats, err
		}
	}
	return stats, nil
}

func (lc *LiveCheck) CalculateOutDegreeDiskGraphs(stats any) (any, error) {
	lc.require()
	if lc.isNoOp() {
		return lc.noOpStats, nil
	}
	for _, checker := range lc.ownedCheckers() {
		checker.require()
		var err error
		if checker.tableau {
			dgraph := checker.ownedTableauDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			stats, err = dgraph.CalculateOutDegreeDiskGraph(stats)
		} else {
			dgraph := checker.ownedDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			stats, err = dgraph.CalculateOutDegreeDiskGraph(stats)
		}
		if err != nil {
			return stats, err
		}
	}
	return stats, nil
}

func (lc *LiveCheck) Recover() error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	for _, checker := range lc.ownedCheckers() {
		PrintMessage(ECTLCAAAAAAA)
		if err := checker.Recover(); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LiveCheck) Reset() error {
	lc.require()
	if lc.isNoOp() {
		return nil
	}
	for _, checker := range lc.ownedCheckers() {
		checker.require()
		if checker.tableau {
			dgraph := checker.ownedTableauDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			if err := dgraph.Reset(); err != nil {
				return err
			}
		} else {
			dgraph := checker.ownedDiskGraph()
			if dgraph == nil {
				panic(NewNullPointerException())
			}
			if err := dgraph.Reset(); err != nil {
				return err
			}
		}
		checker.Reset()
	}
	return nil
}

func (lc *LiveCheck) require() {
	if lc == nil {
		panic(NewNullPointerException())
	}
}

func (e *LiveException) GetMessage() *string {
	e.require()
	if e.ownedHasMessage() || e.ownedMessage() != "" {
		return javaString(e.ownedMessage())
	}
	return nil
}

func (e *LiveException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func (e *LiveException) ownedErrorCode() int {
	if e.ownersCaptured {
		return e.errorCodeOwner
	}
	return e.ErrorCode
}

func (e *LiveException) ownedHasMessage() bool {
	if e.ownersCaptured {
		return e.hasMessageOwner
	}
	return e.hasMessage
}

func (e *LiveException) ownedMessage() string {
	if e.ownersCaptured {
		return e.msgOwner
	}
	return e.Msg
}

func (e *LiveCounterExampleException) ownedLiveException() *LiveException {
	if e.ownersCaptured {
		return e.liveExceptionOwner
	}
	return e.LiveException
}

func (e *LiveCounterExampleException) ownedCounterExample() *CounterExample {
	if e.ownersCaptured {
		return e.counterExampleOwner
	}
	return e.CounterExample
}
