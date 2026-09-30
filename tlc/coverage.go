package tlc

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type CoverageRelation int

const (
	CoverageRelationUnknown CoverageRelation = iota
	CoverageRelationInit
	CoverageRelationNext
	CoverageRelationProp
	CoverageRelationConstraint
)

type coveragePair struct {
	Primary   int64
	Secondary int64
}

func (p coveragePair) isZero() bool {
	return p.Primary == 0 && p.Secondary == 0
}

func (p coveragePair) isNonZero() bool {
	return p.Primary > 0 || p.Secondary > 0
}

func newActionCostModel(action *Action, relation CoverageRelation) CostModel {
	if action == nil {
		return DoNotRecordCostModel
	}
	root := newCostModelNode(action.Pred, nil)
	root.Action = action
	root.Relation = relation
	root.Substs = NewInsMap[semanticNodeKey, *CostModelNode]()
	return CostModel{node: root}
}

func (m CostModel) PutSubst(subst Subst, child CostModel) CostModel {
	root := m.GetRoot()
	if root.node == nil || child.node == nil {
		return m
	}
	if root.node.Substs == nil {
		root.node.Substs = NewInsMap[semanticNodeKey, *CostModelNode]()
	}
	root.node.Substs.Set(newSemanticNodeKey(subst.Expr), child.node)
	return m
}

func (m CostModel) MarkUnchanged() CostModel {
	if m.node != nil {
		m.node.Unchanged = true
		m.node.Primed = true
	}
	return m
}

func (m CostModel) report() CostModel {
	if m.node == nil {
		return m
	}
	if m.node.Action != nil {
		m.reportAction()
		return m
	}
	m.reportNode(0, true)
	return m
}

func (m CostModel) reportAction() {
	action := m.node.Action
	location := coverageActionLocation(action)
	switch m.node.Relation {
	case CoverageRelationProp:
		PrintMessage(ECTLCCoverageProperty, location)
	case CoverageRelationInit:
		PrintMessage(ECTLCCoverageInit, location, fmt.Sprint(m.GetPrimary()), fmt.Sprint(m.GetPrimary()+m.GetSecondary()))
	case CoverageRelationConstraint:
		PrintMessage(ECTLCCoverageConstraint, location, fmt.Sprint(m.GetSecondary()), fmt.Sprint(m.GetPrimary()+m.GetSecondary()))
	default:
		PrintMessage(ECTLCCoverageNext, location, fmt.Sprint(m.GetSecondary()), fmt.Sprint(m.GetPrimary()))
	}
	for _, child := range m.node.Children.All() {
		CostModel{node: child}.reportNode(0, true)
	}
}

func (m CostModel) reportNode(level int, fresh bool) {
	if m.node == nil {
		return
	}
	if m.node.Unchanged {
		m.reportUnchangedNode(level, fresh)
		return
	}
	collected := m.collectChildren(fresh)
	if len(collected) == 0 {
		if m.evalCount(fresh) == 0 && !m.IsPrimed() {
			return
		}
		m.printSelf(level)
		return
	}
	node := coveragePair{Primary: m.evalCount(fresh), Secondary: m.secondaryCount(fresh)}
	if len(collected) == 1 {
		consistent := collected[0]
		if consistent.Primary < node.Primary || consistent.Secondary < node.Secondary {
			m.printSelf(level)
			m.printChildren(level + 1)
			return
		}
		if !m.IsPrimed() && node.isZero() && consistent.isNonZero() {
			m.printSelfCounts(level, consistent.Primary, consistent.Secondary)
			return
		}
		if node.isZero() && consistent.isZero() {
			if m.IsPrimed() {
				m.printSelf(level)
			}
			m.printChildren(level)
			return
		}
		if node == consistent {
			m.printSelf(level)
			return
		}
	}
	if node.isNonZero() || m.IsPrimed() {
		m.printSelf(level)
		level++
	}
	m.printChildren(level)
}

func (m CostModel) reportUnchangedNode(level int, fresh bool) {
	collected := make([]coveragePair, 0)
	for _, pair := range m.collectChildren(fresh) {
		if !pair.isZero() {
			collected = append(collected, pair)
		}
	}
	if len(collected) == 0 {
		m.printSelf(level)
		return
	}
	count := m.evalCount(fresh)
	if collected[0].Primary > count {
		count = collected[0].Primary
	}
	m.printSelfCounts(level, count, 0)
}

func (m CostModel) collectChildren(fresh bool) []coveragePair {
	out := make([]coveragePair, 0)
	if m.node == nil || m.node.Children == nil {
		return out
	}
	seen := make(map[coveragePair]bool)
	for _, child := range m.node.Children.All() {
		for _, pair := range (CostModel{node: child}).collectAndFreezeEvalCounts(fresh) {
			if !seen[pair] {
				seen[pair] = true
				out = append(out, pair)
			}
		}
	}
	return out
}

func (m CostModel) collectAndFreezeEvalCounts(fresh bool) []coveragePair {
	if m.node == nil {
		return nil
	}
	if fresh {
		m.node.SnapshotPrimary = m.GetPrimary()
		m.node.SnapshotSecondary = m.GetSecondary()
		m.node.ChildCounts = m.node.ChildCounts[:0]
		if m.node.SnapshotPrimary > 0 || m.node.SnapshotSecondary > 0 || m.IsPrimed() {
			m.node.ChildCounts = append(m.node.ChildCounts, coveragePair{m.node.SnapshotPrimary, m.node.SnapshotSecondary})
		}
		for _, pair := range m.collectChildren(true) {
			m.node.ChildCounts = appendCoveragePair(m.node.ChildCounts, pair)
		}
	}
	return append([]coveragePair(nil), m.node.ChildCounts...)
}

func appendCoveragePair(pairs []coveragePair, pair coveragePair) []coveragePair {
	for _, existing := range pairs {
		if existing == pair {
			return pairs
		}
	}
	return append(pairs, pair)
}

func (m CostModel) evalCount(fresh bool) int64 {
	if m.node == nil {
		return -1
	}
	if fresh {
		return m.node.Primary
	}
	return m.node.SnapshotPrimary
}

func (m CostModel) secondaryCount(fresh bool) int64 {
	if m.node == nil {
		return -1
	}
	if fresh {
		return m.node.Secondary
	}
	return m.node.SnapshotSecondary
}

func (m CostModel) printChildren(level int) {
	if m.node == nil || m.node.Children == nil {
		return
	}
	for _, child := range m.node.Children.All() {
		CostModel{node: child}.reportNode(level, false)
	}
}

func (m CostModel) printSelf(level int) {
	m.printSelfCounts(level, m.GetPrimary(), m.GetSecondary())
}

func (m CostModel) printSelfCounts(level int, count int64, cost int64) {
	location := coverageNodeLocation(m.node)
	indent := strings.Repeat(string(CoverageIndent), max(level, 0))
	if cost > 0 {
		PrintMessage(ECTLCCoverageValueCost, indent+location, fmt.Sprint(count), fmt.Sprint(cost))
		return
	}
	PrintMessage(ECTLCCoverageValue, indent+location, fmt.Sprint(count))
}

func coverageActionLocation(action *Action) string {
	if action == nil {
		return "<unknown>"
	}
	return action.GetLocation()
}

func coverageNodeLocation(node *CostModelNode) string {
	if node == nil {
		return "<unknown>"
	}
	if node.Action != nil {
		return coverageActionLocation(node.Action)
	}
	return SemanticString(node.Expr)
}

type coverageCreator struct {
	tool      *Tool
	primed    map[semanticNodeKey]bool
	stack     []CostModel
	root      CostModel
	visiting  map[semanticNodeKey]bool
	recursive map[*OpDefNode]CostModel
}

func CreateCoverageCostModels(tool *Tool) {
	if tool == nil {
		return
	}
	InitializeStateVariableCoverageCounters()
	creator := newCoverageCreator(tool)
	init := tool.GetInitStateSpec()
	for _, action := range init {
		if action != nil {
			action.CM = creator.createForAction(action, CoverageRelationInit)
		}
	}
	sharedNext := NewInsMap[semanticNodeKey, CostModel]()
	for _, action := range tool.GetActions() {
		if action == nil {
			continue
		}
		key := newSemanticNodeKey(action.Pred)
		if cm, ok := sharedNext.Get2(key); ok {
			action.CM = cm
			continue
		}
		action.CM = creator.createForAction(action, CoverageRelationNext)
		sharedNext.Set(key, action.CM)
	}
	for _, invariant := range tool.GetInvariants() {
		if invariant != nil && !invariant.IsInternal() {
			invariant.CM = creator.createForAction(invariant, CoverageRelationProp)
		}
	}
	for _, impliedInit := range tool.GetImpliedInits() {
		if impliedInit != nil {
			impliedInit.CM = creator.createForAction(impliedInit, CoverageRelationProp)
		}
	}
	for _, impliedAction := range tool.GetImpliedActions() {
		if impliedAction != nil {
			impliedAction.CM = creator.createForAction(impliedAction, CoverageRelationProp)
		}
	}
	for _, constraint := range tool.GetActionConstraints() {
		creator.assignConstraintCostModel(constraint)
	}
	for _, constraint := range tool.GetModelConstraints() {
		creator.assignConstraintCostModel(constraint)
	}
}

func ReportCoverage(tool *Tool, startTime time.Time) {
	reportCoverage(tool)
	if !startTime.IsZero() && time.Since(startTime) > 5*time.Minute {
		PrintMessage(ECTLCCoverageEndOverhead)
		return
	}
	PrintMessage(ECTLCCoverageEnd)
}

func reportCoverage(tool *Tool) {
	if tool == nil {
		return
	}
	PrintMessage(ECTLCCoverageStart)
	for _, variable := range StateVariables() {
		if variable.CountDistinct == nil {
			continue
		}
		count := variable.CountDistinct.Count()
		if count >= 0 {
			PrintMessage(ECTLCCoverageVar, variable.Name.String(), variable.Name.String(), fmt.Sprint(count))
		}
	}
	for _, action := range tool.GetInitStateSpec() {
		if action != nil {
			action.CM.Report()
		}
	}
	actions := tool.GetActions()
	sort.SliceStable(actions, func(i, j int) bool {
		return SemanticString(actions[i].Pred) < SemanticString(actions[j].Pred)
	})
	reported := NewInsMap[*CostModelNode, bool]()
	for _, action := range actions {
		if action == nil || action.CM.node == nil {
			continue
		}
		if reported.Get(action.CM.node) {
			continue
		}
		action.CM.Report()
		reported.Set(action.CM.node, true)
	}
	for _, invariant := range tool.GetInvariants() {
		if invariant != nil && !invariant.IsInternal() {
			invariant.CM.Report()
		}
	}
	for _, impliedInit := range tool.GetImpliedInits() {
		if impliedInit != nil {
			impliedInit.CM.Report()
		}
	}
	for _, impliedAction := range tool.GetImpliedActions() {
		if impliedAction != nil {
			impliedAction.CM.Report()
		}
	}
	reportConstraintCoverage(tool.GetActionConstraints())
	reportConstraintCoverage(tool.GetModelConstraints())
}

func reportConstraintCoverage(nodes []SemanticNode) {
	for _, node := range nodes {
		if action, ok := SemanticToolObject(node).(*Action); ok && action != nil {
			action.CM.Report()
		}
	}
}

func newCoverageCreator(tool *Tool) *coverageCreator {
	creator := &coverageCreator{
		tool:      tool,
		primed:    make(map[semanticNodeKey]bool),
		visiting:  make(map[semanticNodeKey]bool),
		recursive: make(map[*OpDefNode]CostModel),
	}
	for enum := tool.GetPrimedLocs().Keys(); ; {
		node := enum.NextElement()
		if node == nil {
			break
		}
		creator.primed[newSemanticNodeKey(node)] = true
	}
	return creator
}

func (c *coverageCreator) createForAction(action *Action, relation CoverageRelation) CostModel {
	c.stack = c.stack[:0]
	c.visiting = make(map[semanticNodeKey]bool)
	c.recursive = make(map[*OpDefNode]CostModel)
	c.root = newActionCostModel(action, relation)
	c.stack = append(c.stack, c.root)
	c.walk(action.Pred)
	if len(c.stack) != 1 {
		c.stack = c.stack[:1]
	}
	return c.root
}

func (c *coverageCreator) assignConstraintCostModel(expr SemanticNode) {
	if expr == nil {
		return
	}
	action, _ := SemanticToolObject(expr).(*Action)
	if action == nil {
		action = NewAction(expr, EmptyContext, SemanticString(expr))
		setSemanticToolObject(expr, action)
	}
	action.CM = c.createForAction(action, CoverageRelationConstraint)
}

func (c *coverageCreator) walk(node SemanticNode) {
	if node == nil {
		return
	}
	key := newSemanticNodeKey(node)
	if c.visiting[key] {
		return
	}
	c.visiting[key] = true
	defer delete(c.visiting, key)

	switch n := node.(type) {
	case *OpApplNode:
		c.preOpAppl(n)
		for _, child := range n.Args {
			c.walk(child)
		}
		c.walkQuantifierBounds(n)
		c.postOpAppl(n)
	case *LetInNode:
		for _, let := range n.Lets {
			c.walkOpDef(let)
		}
		c.walk(n.Body)
	case *SubstInNode:
		for _, subst := range n.Substs {
			c.walk(subst.Expr)
			if c.root.node != nil {
				substCM := c.root.Get(subst.Expr)
				if substCM.node != c.root.node {
					c.root.PutSubst(subst, substCM)
				}
			}
		}
		c.walk(n.Body)
	case *APSubstInNode:
		for _, subst := range n.Substs {
			c.walk(subst.Expr)
			if c.root.node != nil {
				substCM := c.root.Get(subst.Expr)
				if substCM.node != c.root.node {
					c.root.PutSubst(subst, substCM)
				}
			}
		}
		c.walk(n.Body)
	case *LabelNode:
		c.walk(n.Body)
	case *ThmOrAssumpDefNode:
		c.walk(n.Body)
	}
}

func (c *coverageCreator) walkQuantifierBounds(n *OpApplNode) {
	if n == nil {
		return
	}
	for _, bound := range n.BdedQuantBounds {
		c.walk(bound)
	}
}

func (c *coverageCreator) walkOpDef(def *OpDefNode) {
	if def == nil {
		return
	}
	if prior, ok := c.recursive[def]; ok {
		c.peek().SetRecursive(prior)
		return
	}
	c.recursive[def] = c.peek()
	c.walk(def.Body)
	delete(c.recursive, def)
}

func (c *coverageCreator) preOpAppl(node *OpApplNode) {
	if c.isStandardModuleNode(node) {
		return
	}
	parent := c.peek()
	cm := parent.AddChild(node)
	if c.primed[newSemanticNodeKey(node)] {
		cm.SetPrimed()
	}
	if node.Operator != nil && node.Operator.Name != nil && GetOpCode(node.Operator.Name) == OpcodeUnchanged {
		cm.MarkUnchanged()
	}
	c.stack = append(c.stack, cm)
	if def, ok := c.lookupOpDef(node); ok {
		c.walkOpDef(def)
	}
}

func (c *coverageCreator) postOpAppl(node *OpApplNode) {
	if c.isStandardModuleNode(node) {
		return
	}
	if len(c.stack) > 1 {
		c.stack = c.stack[:len(c.stack)-1]
	}
}

func (c *coverageCreator) lookupOpDef(node *OpApplNode) (*OpDefNode, bool) {
	if c == nil || c.tool == nil || node == nil || node.Operator == nil {
		return nil, false
	}
	val := c.tool.Lookup(node.Operator, EmptyContext, EmptyState, false)
	def, ok := val.(*OpDefNode)
	return def, ok && def != nil
}

func (c *coverageCreator) isStandardModuleNode(node *OpApplNode) bool {
	if node == nil || node.Operator == nil || node.Operator.Name == nil {
		return false
	}
	return GetOpCode(node.Operator.Name) != 0
}

func (c *coverageCreator) peek() CostModel {
	if len(c.stack) == 0 {
		return c.root
	}
	return c.stack[len(c.stack)-1]
}

func setSemanticToolObject(node SemanticNode, value any) {
	if setter, ok := node.(interface{ SetToolObject(any) }); ok {
		setter.SetToolObject(value)
	}
}
