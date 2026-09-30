package tlc

import (
	"fmt"
)

type LiveException struct {
	Code int
	Err  error
}

func (e *LiveException) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("liveness error %d", e.Code)
}

type LiveCounterExampleException struct {
	Trace []*TLCStateInfo
}

func (e *LiveCounterExampleException) Error() string {
	return "temporal property violated"
}

type LivenessStateWriter struct {
	Noop bool
}

func NewNoopLivenessStateWriter() *LivenessStateWriter {
	return &LivenessStateWriter{Noop: true}
}

type LiveChecker struct {
	Solution *OrderOfSolution
	Soln     int
	Writer   *LivenessStateWriter
	Graph    *InsMap[string, *GraphNode]
	Initial  []*GraphNode
	Size     int64
	LastSize int64
}

func NewLiveChecker(solution *OrderOfSolution, soln int, writer *LivenessStateWriter) *LiveChecker {
	if writer == nil {
		writer = NewNoopLivenessStateWriter()
	}
	return &LiveChecker{
		Solution: solution,
		Soln:     soln,
		Writer:   writer,
		Graph:    NewInsMap[string, *GraphNode](),
	}
}

func (c *LiveChecker) AddInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	if c == nil || c.Solution == nil {
		return nil
	}
	if c.Solution.HasTableau() {
		for i := 0; i < c.Solution.Tableau.InitCnt; i++ {
			tnode := c.Solution.Tableau.GetNode(i)
			ok, err := tnode.IsConsistent(state, tool)
			if err != nil {
				return err
			}
			if ok {
				node, err := c.ensureGraphNode(tool, state, stateFP, tnode.Index)
				if err != nil {
					return err
				}
				c.Initial = append(c.Initial, node)
			}
		}
		return nil
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
	if c.Solution.HasTableau() {
		return c.addNextStateTableau(tool, s0, fp0, nextStates, actionResults, checkStateRes)
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

func (c *LiveChecker) addNextStateTableau(tool *Tool, s0 *TLCStateMut, fp0 uint64, nextStates *SetOfStates, actionResults *BitVector, checkStateRes []bool) error {
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

func (c *LiveChecker) CheckSccs(tool *Tool, finalCheck bool) (bool, error) {
	_ = tool
	_ = finalCheck
	if c == nil || c.Solution == nil || c.Graph == nil {
		return false, nil
	}
	for _, node := range c.Graph.All() {
		node.Realign()
	}
	for _, pem := range c.Solution.PEMs {
		search := newLiveSCCSearch(c, pem)
		if search.check() {
			c.LastSize = c.Size
			return true, nil
		}
	}
	c.LastSize = c.Size
	return false, nil
}

func (c *LiveChecker) Reset() {
	c.Graph = NewInsMap[string, *GraphNode]()
	c.Initial = nil
	c.Size = 0
	c.LastSize = 0
}

func graphNodeKey(fp uint64, tidx int) string {
	return fmt.Sprintf("%d:%d", fp, tidx)
}

type liveSCCSearch struct {
	checker *LiveChecker
	pem     *PossibleErrorModel
	slen    int
	alen    int
	next    int
	index   map[string]int
	lowlink map[string]int
	onStack map[string]bool
	stack   []*GraphNode
	found   bool
}

func newLiveSCCSearch(checker *LiveChecker, pem *PossibleErrorModel) *liveSCCSearch {
	return &liveSCCSearch{
		checker: checker,
		pem:     pem,
		slen:    len(checker.Solution.CheckState),
		alen:    len(checker.Solution.CheckAction),
		index:   make(map[string]int),
		lowlink: make(map[string]int),
		onStack: make(map[string]bool),
	}
}

func (s *liveSCCSearch) check() bool {
	for _, init := range s.checker.Initial {
		if init == nil {
			continue
		}
		key := graphNodeKey(init.StateFP, init.TIndex)
		if _, ok := s.index[key]; !ok {
			s.strongConnect(init)
			if s.found {
				return true
			}
		}
	}
	return false
}

func (s *liveSCCSearch) strongConnect(node *GraphNode) {
	if node == nil || s.found {
		return
	}
	key := graphNodeKey(node.StateFP, node.TIndex)
	s.index[key] = s.next
	s.lowlink[key] = s.next
	s.next++
	s.stack = append(s.stack, node)
	s.onStack[key] = true

	for _, edge := range s.successors(node) {
		succ := edge.node
		succKey := graphNodeKey(succ.StateFP, succ.TIndex)
		if _, ok := s.index[succKey]; !ok {
			s.strongConnect(succ)
			if s.found {
				return
			}
			if s.lowlink[succKey] < s.lowlink[key] {
				s.lowlink[key] = s.lowlink[succKey]
			}
		} else if s.onStack[succKey] && s.index[succKey] < s.lowlink[key] {
			s.lowlink[key] = s.index[succKey]
		}
	}

	if s.lowlink[key] != s.index[key] {
		return
	}
	component := make([]*GraphNode, 0)
	for len(s.stack) > 0 {
		last := s.stack[len(s.stack)-1]
		s.stack = s.stack[:len(s.stack)-1]
		lastKey := graphNodeKey(last.StateFP, last.TIndex)
		s.onStack[lastKey] = false
		component = append(component, last)
		if lastKey == key {
			break
		}
	}
	if s.componentViolates(component) {
		s.found = true
	}
}

type liveGraphEdge struct {
	node  *GraphNode
	index int
}

func (s *liveSCCSearch) successors(node *GraphNode) []liveGraphEdge {
	out := make([]liveGraphEdge, 0, node.SuccSize())
	for i := 0; i < node.SuccSize(); i++ {
		if !s.edgeSatisfiesEA(node, i) {
			continue
		}
		fp := node.GetStateFP(i)
		tidx := node.GetTIndex(i)
		succ := s.checker.Graph.Get(graphNodeKey(fp, tidx))
		if succ != nil {
			out = append(out, liveGraphEdge{node: succ, index: i})
		}
	}
	return out
}

func (s *liveSCCSearch) edgeSatisfiesEA(node *GraphNode, edge int) bool {
	if s.pem == nil {
		return true
	}
	for _, idx := range s.pem.EAAction {
		if !node.GetCheckAction(s.slen, s.alen, edge, idx) {
			return false
		}
	}
	return true
}

func (s *liveSCCSearch) componentViolates(component []*GraphNode) bool {
	if len(component) == 0 {
		return false
	}
	componentSet := make(map[string]struct{}, len(component))
	for _, node := range component {
		componentSet[graphNodeKey(node.StateFP, node.TIndex)] = struct{}{}
	}
	if len(component) == 1 && !s.hasComponentSelfLoop(component[0]) {
		return false
	}

	aeslen := 0
	aealen := 0
	if s.pem != nil {
		aeslen = len(s.pem.AEState)
		aealen = len(s.pem.AEAction)
	}
	aeStateRes := make([]bool, aeslen)
	aeActionRes := make([]bool, aealen)
	promiseRes := make([]bool, len(s.checker.Solution.Promises))

	for _, node := range component {
		for i := 0; i < aeslen; i++ {
			if !aeStateRes[i] {
				aeStateRes[i] = node.GetCheckState(s.pem.AEState[i])
			}
		}

		if aealen > 0 {
			for edge := 0; edge < node.SuccSize(); edge++ {
				if !s.edgeSatisfiesEA(node, edge) {
					continue
				}
				fp := node.GetStateFP(edge)
				tidx := node.GetTIndex(edge)
				if _, ok := componentSet[graphNodeKey(fp, tidx)]; !ok {
					continue
				}
				for i := 0; i < aealen; i++ {
					if !aeActionRes[i] {
						aeActionRes[i] = node.GetCheckAction(s.slen, s.alen, edge, s.pem.AEAction[i])
					}
				}
			}
		}

		if s.checker.Solution.HasTableau() && node.TIndex >= 0 && node.TIndex < s.checker.Solution.Tableau.Size() {
			par := s.checker.Solution.Tableau.GetNode(node.TIndex).Par
			for i, promise := range s.checker.Solution.Promises {
				if !promiseRes[i] && par.IsFulfilling(promise) {
					promiseRes[i] = true
				}
			}
		}
	}

	for _, ok := range aeStateRes {
		if !ok {
			return false
		}
	}
	for _, ok := range aeActionRes {
		if !ok {
			return false
		}
	}
	for _, ok := range promiseRes {
		if !ok {
			return false
		}
	}
	return true
}

func (s *liveSCCSearch) hasComponentSelfLoop(node *GraphNode) bool {
	if node == nil {
		return false
	}
	for i := 0; i < node.SuccSize(); i++ {
		if node.GetStateFP(i) == node.StateFP && node.GetTIndex(i) == node.TIndex && s.edgeSatisfiesEA(node, i) {
			return true
		}
	}
	return false
}

type LiveCheck struct {
	Tool     *Tool
	MetaDir  string
	Checkers []*LiveChecker
	NoOp     bool
}

func NewNoOpLiveCheck(tool *Tool, metadir string) *LiveCheck {
	return &LiveCheck{Tool: tool, MetaDir: metadir, NoOp: true}
}

func NewLiveCheck(tool *Tool, solutions []*OrderOfSolution, metadir string) *LiveCheck {
	check := &LiveCheck{Tool: tool, MetaDir: metadir}
	for i, solution := range solutions {
		check.Checkers = append(check.Checkers, NewLiveChecker(solution, i, NewNoopLivenessStateWriter()))
	}
	return check
}

func NewAddAndCheckLiveCheck(tool *Tool, solutions []*OrderOfSolution, metadir string) *LiveCheck {
	return NewLiveCheck(tool, solutions, metadir)
}

func (lc *LiveCheck) AddInitState(tool *Tool, state *TLCStateMut, stateFP uint64) error {
	if lc == nil || lc.NoOp {
		return nil
	}
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
		if checker.LastSize == 0 && checker.Size > 0 {
			return true
		}
		if checker.LastSize > 0 {
			delta := float64(checker.Size-checker.LastSize) / float64(checker.LastSize)
			if delta > threshold {
				return true
			}
		}
	}
	return false
}

func (lc *LiveCheck) Check(tool *Tool, forceCheck bool) (int, error) {
	if lc == nil || lc.NoOp {
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
	return lc.check0(tool, true)
}

func (lc *LiveCheck) check0(tool *Tool, finalCheck bool) (int, error) {
	for _, checker := range lc.Checkers {
		found, err := checker.CheckSccs(tool, finalCheck)
		if err != nil {
			return ECGeneral, err
		}
		if found {
			return ECGeneral, nil
		}
	}
	return NoError, nil
}

func (lc *LiveCheck) CheckTrace(tool *Tool, trace func() *StateVec) error {
	if lc == nil || lc.NoOp || trace == nil {
		return nil
	}
	states := trace()
	if states == nil || states.Size() == 0 {
		return nil
	}
	for i := 0; i < states.Size(); i++ {
		state := states.At(i)
		if i == 0 {
			if err := lc.AddInitState(tool, state, state.FingerPrint()); err != nil {
				return err
			}
			continue
		}
		set := NewSetOfStates(1)
		set.Put(state)
		if err := lc.AddNextState(tool, states.At(i-1), states.At(i-1).FingerPrint(), set); err != nil {
			return err
		}
	}
	_, err := lc.FinalCheck(tool)
	return err
}

func (lc *LiveCheck) Close() error                  { return nil }
func (lc *LiveCheck) BeginChkpt() error             { return nil }
func (lc *LiveCheck) CommitChkpt() error            { return nil }
func (lc *LiveCheck) FlushWritesToDiskFiles() error { return nil }
func (lc *LiveCheck) Recover() error                { return nil }
func (lc *LiveCheck) Reset() error {
	if lc != nil {
		for _, checker := range lc.Checkers {
			checker.Reset()
		}
	}
	return nil
}
