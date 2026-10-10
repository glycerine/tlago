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

const costModelCreatorImpliedProperty = "tlc2.tool.coverage.CostModelCreator.implied"

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
	root.Substs = NewInsMap[uint64, *CostModelNode]()
	return CostModel{node: root}
}

func (m CostModel) PutSubst(subst Subst, child CostModel) CostModel {
	root := m.GetRoot()
	if root.node == nil || child.node == nil {
		return m
	}
	if root.node.Substs == nil {
		root.node.Substs = NewInsMap[uint64, *CostModelNode]()
	}
	root.node.Substs.Set(substIdentity(subst), child.node)
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
		// Java compares consistentChildren.secondary to itself here. Preserve that
		// typo because it affects when coverage subtrees collapse in reports.
		if consistent.Primary < node.Primary || consistent.Secondary < consistent.Secondary {
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
				level++
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
	count := m.GetPrimary()
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
		return m.GetPrimary()
	}
	return m.node.SnapshotPrimary
}

func (m CostModel) secondaryCount(fresh bool) int64 {
	if m.node == nil {
		return -1
	}
	if fresh {
		return m.GetSecondary()
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
	if !action.IsDeclared() || action.OpDef == nil {
		return action.String()
	}
	declaration := action.GetDeclarationLocation()
	definition := semanticNodeLocation(action.OpDef.Body)
	actual := semanticNodeLocation(action.Pred)
	if definition == actual {
		return fmt.Sprintf("<%s %s>", action.GetName(), declaration.String())
	}
	return fmt.Sprintf("<%s %s (%d %d %d %d)>", action.GetName(), declaration.String(), actual.BeginLine, actual.BeginColumn, actual.EndLine, actual.EndColumn)
}

func coverageNodeLocation(node *CostModelNode) string {
	if node == nil {
		return "<unknown>"
	}
	if node.Action != nil {
		return coverageActionLocation(node.Action)
	}
	location, _ := semanticNodeSourceLocation(node.Expr)
	return location.String()
}

type coverageCreator struct {
	tool         *Tool
	primed       map[SourceLocation]bool
	stack        []CostModel
	root         CostModel
	ctx          *Context
	visiting     map[semanticNodeKey]bool
	opDefNodes   map[*OpDefNode]bool
	substs       map[semanticNodeKey]Subst
	node2Wrapper map[semanticNodeKey][]CostModel
	letIns       map[semanticNodeKey]SemanticNode
}

func CreateCoverageCostModels(tool *Tool) {
	if tool == nil {
		return
	}
	creator := newCoverageCreator(tool)
	init := tool.GetInitStateSpec()
	for i := 0; i < init.Size(); i++ {
		action := init.ElementAt(i)
		if action == nil && tool.SpecProcessor != nil {
			panic(NewNullPointerException())
		}
		if action != nil {
			action.CM = creator.createForAction(action, CoverageRelationInit)
		}
	}
	sharedNext := NewInsMap[semanticNodeKey, CostModel]()
	for _, action := range tool.requireActionArray(tool.GetActions()) {
		if action == nil {
			if tool.SpecProcessor != nil {
				panic(NewNullPointerException())
			}
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
	for _, invariant := range tool.requireActionArray(tool.GetInvariants()) {
		if invariant == nil && tool.SpecProcessor != nil {
			panic(NewNullPointerException())
		}
		if invariant != nil && !invariant.IsInternal() {
			invariant.CM = creator.createForAction(invariant, CoverageRelationProp)
		}
	}
	for _, constraint := range tool.requireConstraintArray(tool.GetActionConstraints()) {
		creator.assignConstraintCostModel(constraint)
	}
	for _, constraint := range tool.requireConstraintArray(tool.GetModelConstraints()) {
		creator.assignConstraintCostModel(constraint)
	}
	if coverageImpliedEnabled() {
		for _, impliedInit := range tool.requireActionArray(tool.GetImpliedInits()) {
			if impliedInit == nil && tool.SpecProcessor != nil {
				panic(NewNullPointerException())
			}
			if impliedInit != nil {
				impliedInit.CM = creator.createForAction(impliedInit, CoverageRelationProp)
			}
		}
		for _, impliedAction := range tool.requireActionArray(tool.GetImpliedActions()) {
			if impliedAction == nil && tool.SpecProcessor != nil {
				panic(NewNullPointerException())
			}
			if impliedAction != nil {
				impliedAction.CM = creator.createForAction(impliedAction, CoverageRelationProp)
			}
		}
	}
	if processor := tool.GetSpecProcessor(); processor != nil {
		nodes := processor.GetVariablesNodes()
		if nodes == nil {
			panic(NewNullPointerException())
		}
		for _, node := range nodes {
			if node == nil {
				panic(NewNullPointerException())
			}
			node.SetCountDistinct(NewCountDistinctSyncedHyperLogLog(10))
		}
	} else {
		InitializeStateVariableCoverageCounters()
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
	if processor := tool.GetSpecProcessor(); processor != nil {
		nodes := processor.GetVariablesNodes()
		if nodes == nil {
			panic(NewNullPointerException())
		}
		// Capture the declaration array, then read each current slot and counter
		// in order. Recorder callbacks can change later entries or counters.
		for _, node := range nodes {
			if node == nil {
				panic(NewNullPointerException())
			}
			counter := node.GetCountDistinct()
			if counter == nil {
				panic(NewNullPointerException())
			}
			count := counter.Count()
			if count >= 0 {
				if node.Name == nil {
					panic(NewNullPointerException())
				}
				PrintMessage(ECTLCCoverageVar, node.Name.String(), node.GetSourceLocation().String(), fmt.Sprint(count))
			}
		}
	} else {
		for _, variable := range stateVariablesForTool(tool) {
			if variable.CountDistinct == nil {
				if variable.declaration != nil {
					panic(NewNullPointerException())
				}
				continue
			}
			count := variable.CountDistinct.Count()
			if count >= 0 {
				PrintMessage(ECTLCCoverageVar, variable.Name.String(), variable.GetSourceLocation().String(), fmt.Sprint(count))
			}
		}
	}
	init := tool.GetInitStateSpec()
	for i := 0; i < init.Size(); i++ {
		action := init.ElementAt(i)
		if action == nil && tool.SpecProcessor != nil {
			panic(NewNullPointerException())
		}
		if action != nil {
			action.CM.Report()
		}
	}
	// Java reports through a separate TreeSet. Sorting the checker's action
	// array would change successor order after a periodic coverage report.
	currentActions := tool.requireActionArray(tool.GetActions())
	if tool.SpecProcessor != nil {
		for _, action := range currentActions {
			if action == nil {
				panic(NewNullPointerException())
			}
		}
	}
	actions := append([]*Action(nil), currentActions...)
	if tool.SpecProcessor != nil {
		// TreeSet keeps the first inserted action at each predicate location,
		// even if a later action there has a different cost model.
		seen := make(map[SourceLocation]bool)
		unique := make([]*Action, 0, len(actions))
		for _, action := range actions {
			if action.Pred == nil {
				panic(NewNullPointerException())
			}
			location := action.GetDefinitionLocation()
			if !seen[location] {
				seen[location] = true
				unique = append(unique, action)
			}
		}
		actions = unique
		sort.SliceStable(actions, func(i, j int) bool {
			return sourceLocationLess(actions[i].GetDefinitionLocation(), actions[j].GetDefinitionLocation())
		})
	} else {
		sort.SliceStable(actions, func(i, j int) bool {
			return coverageActionLess(actions[i], actions[j])
		})
	}
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
	for _, invariant := range tool.requireActionArray(tool.GetInvariants()) {
		if invariant == nil && tool.SpecProcessor != nil {
			panic(NewNullPointerException())
		}
		if invariant != nil && !invariant.IsInternal() {
			invariant.CM.Report()
		}
	}
	reportConstraintCoverage(tool, tool.requireConstraintArray(tool.GetActionConstraints()))
	reportConstraintCoverage(tool, tool.requireConstraintArray(tool.GetModelConstraints()))
	if coverageImpliedEnabled() {
		for _, impliedInit := range tool.requireActionArray(tool.GetImpliedInits()) {
			if impliedInit == nil && tool.SpecProcessor != nil {
				panic(NewNullPointerException())
			}
			if impliedInit != nil {
				impliedInit.CM.Report()
			}
		}
		for _, impliedAction := range tool.requireActionArray(tool.GetImpliedActions()) {
			if impliedAction == nil && tool.SpecProcessor != nil {
				panic(NewNullPointerException())
			}
			if impliedAction != nil {
				impliedAction.CM.Report()
			}
		}
	}
}

func coverageActionLess(left *Action, right *Action) bool {
	if left == nil || right == nil {
		return right != nil
	}
	lloc := left.GetDefinitionLocation()
	rloc := right.GetDefinitionLocation()
	if lloc != rloc {
		return sourceLocationLess(lloc, rloc)
	}
	lpred := ""
	if left != nil {
		lpred = SemanticString(left.Pred)
	}
	rpred := ""
	if right != nil {
		rpred = SemanticString(right.Pred)
	}
	if lpred != rpred {
		return lpred < rpred
	}
	return left.GetName() < right.GetName()
}

func sourceLocationLess(left SourceLocation, right SourceLocation) bool {
	if left.Source != right.Source {
		// Location.compareTo compares module UniqueString tokens, not text.
		return UniqueStringOf(left.Source).Compare(UniqueStringOf(right.Source)) < 0
	}
	if left.BeginLine != right.BeginLine {
		return left.BeginLine < right.BeginLine
	}
	if left.BeginColumn != right.BeginColumn {
		return left.BeginColumn < right.BeginColumn
	}
	if left.EndLine != right.EndLine {
		return left.EndLine < right.EndLine
	}
	return left.EndColumn < right.EndColumn
}

func coverageImpliedEnabled() bool {
	if value, ok := tlcLookupSystemProperty(costModelCreatorImpliedProperty); ok {
		return javaBooleanProperty(value)
	}
	return true
}

func reportConstraintCoverage(tool *Tool, nodes []SemanticNode) {
	for _, node := range nodes {
		if tool.SpecProcessor != nil && node == nil {
			panic(NewNullPointerException())
		}
		object := SemanticToolObjectForTool(tool, node)
		action, ok := object.(*Action)
		if tool.SpecProcessor != nil {
			if object != nil && !ok {
				panic(NewClassCastException())
			}
			if action == nil {
				panic(NewNullPointerException())
			}
		}
		if action != nil {
			action.CM.Report()
		}
	}
}

func newCoverageCreator(tool *Tool) *coverageCreator {
	creator := &coverageCreator{
		tool:         tool,
		primed:       make(map[SourceLocation]bool),
		visiting:     make(map[semanticNodeKey]bool),
		opDefNodes:   make(map[*OpDefNode]bool),
		substs:       make(map[semanticNodeKey]Subst),
		node2Wrapper: make(map[semanticNodeKey][]CostModel),
		letIns:       make(map[semanticNodeKey]SemanticNode),
	}
	for enum := tool.GetPrimedLocs().Keys(); ; {
		node := enum.NextElement()
		if node == nil {
			break
		}
		// Java's OpApplNodeWrapper equality compares source locations, so
		// separately instantiated applications at the same location match.
		location, _ := semanticNodeSourceLocation(node)
		creator.primed[location] = true
	}
	return creator
}

func (c *coverageCreator) createForAction(action *Action, relation CoverageRelation) CostModel {
	c.stack = c.stack[:0]
	c.ctx = EmptyContext
	c.visiting = make(map[semanticNodeKey]bool)
	c.opDefNodes = make(map[*OpDefNode]bool)
	c.substs = make(map[semanticNodeKey]Subst)
	c.node2Wrapper = make(map[semanticNodeKey][]CostModel)
	c.letIns = make(map[semanticNodeKey]SemanticNode)
	c.root = newActionCostModel(action, relation)
	c.stack = append(c.stack, c.root)
	c.walk(action.Pred)
	if len(c.stack) != 1 {
		c.stack = c.stack[:1]
	}
	return c.root
}

func (c *coverageCreator) assignConstraintCostModel(expr SemanticNode) {
	if c.tool.SpecProcessor != nil {
		if expr == nil {
			panic(NewNullPointerException())
		}
		existing := SemanticToolObjectForTool(c.tool, expr)
		opDef, ok := existing.(*OpDefNode)
		if existing != nil && !ok {
			panic(NewClassCastException())
		}
		action := NewActionFromOpDef(expr, EmptyContext, opDef, false, false)
		action.CM = c.createForAction(action, CoverageRelationConstraint)
		SetSemanticToolObjectForTool(c.tool, expr, action)
		return
	}
	if expr == nil {
		return
	}
	existing := SemanticToolObjectForTool(c.tool, expr)
	action, _ := existing.(*Action)
	if action == nil {
		if opDef, ok := existing.(*OpDefNode); ok && opDef != nil {
			action = NewActionFromOpDef(expr, EmptyContext, opDef, false, false)
		} else {
			action = NewAction(expr, EmptyContext, SemanticString(expr))
		}
		SetSemanticToolObjectForTool(c.tool, expr, action)
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
			if let != nil && let.Body != nil {
				c.letIns[newSemanticNodeKey(let.Body)] = n.Body
			}
		}
		if n.Context != nil {
			n.Context.WalkGraphNodes(c.walk)
		} else {
			for _, let := range coverageLetDefinitions(n.Lets) {
				c.walkOpDef(let)
			}
		}
		c.walk(n.Body)
	case *SubstInNode:
		for _, subst := range n.Substs {
			if subst.Expr != nil {
				c.substs[newSemanticNodeKey(subst.Expr)] = subst
			}
		}
		for _, subst := range n.Substs {
			c.walk(subst.Expr)
		}
		c.walk(n.Body)
	case *APSubstInNode:
		for _, subst := range n.Substs {
			if subst.Expr != nil {
				c.substs[newSemanticNodeKey(subst.Expr)] = subst
			}
		}
		for _, subst := range n.Substs {
			c.walk(subst.Expr)
		}
		c.walk(n.Body)
	case *OpDefNode:
		c.preOpDef(n)
		c.walk(n.Body)
		c.walk(n.StepNode)
		c.postOpDef(n)
	case *OpArgNode:
		// SANY's OpArgNode.walkGraph visits its operator, including the
		// otherwise unattached definition of a LAMBDA argument.
		if def, ok := c.tool.Lookup(n.Op, EmptyContext, EmptyState, false).(*OpDefNode); ok {
			c.walkOpDef(def)
		}
	case *LabelNode:
		c.walk(n.Body)
		for _, param := range n.Params {
			c.walk(param)
		}
	case *ThmOrAssumpDefNode:
		c.walk(n.Body)
	case *AssumeProveNode:
		for _, assume := range n.Assumes {
			c.walk(assume)
		}
		c.walk(n.Prove)
	case *NewSymbNode:
		c.walk(n.Set)
	case *AssumeNode:
		c.walk(n.Assume)
	case *TheoremNode:
		c.walk(n.Theorem)
		c.walk(n.Proof)
	case *LeafProofNode:
		for _, fact := range n.Facts {
			c.walk(fact)
		}
	case *NonLeafProofNode:
		for _, step := range n.Steps {
			c.walk(step)
		}
		if n.Context != nil {
			n.Context.WalkGraphNodes(c.walk)
		}
	case *DefStepNode:
		for _, def := range n.Defs {
			c.walk(def)
		}
	case *UseOrHideNode:
		for _, fact := range n.Facts {
			c.walk(fact)
		}
	}
}

// LetInNode.walkGraph visits SANY Context's default Java Hashtable, rather
// than its declaration-ordered getLets array. Preserve its bucket/chain order
// and rehashing while keeping the evaluator's LET declaration order intact.
func coverageLetDefinitions(lets []*OpDefNode) []*OpDefNode {
	buckets := make([][]*OpDefNode, 11)
	count, threshold := 0, 8
	index := func(def *OpDefNode, capacity int) int {
		return int(uint32(javaStringHashCode(def.Name.String()))&0x7fffffff) % capacity
	}
	put := func(table [][]*OpDefNode, def *OpDefNode) {
		i := index(def, len(table))
		table[i] = append([]*OpDefNode{def}, table[i]...)
	}
	for _, def := range lets {
		if def == nil || def.Name == nil {
			continue
		}
		if count >= threshold {
			old := buckets
			buckets = make([][]*OpDefNode, 2*len(old)+1)
			threshold = 3 * len(buckets) / 4
			for i := len(old) - 1; i >= 0; i-- {
				for _, entry := range old[i] {
					put(buckets, entry)
				}
			}
		}
		put(buckets, def)
		count++
	}
	out := make([]*OpDefNode, 0, count)
	for i := len(buckets) - 1; i >= 0; i-- {
		out = append(out, buckets[i]...)
	}
	return out
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
	if def.GetInRecursive() && c.opDefNodes[def] {
		return
	}
	c.preOpDef(def)
	c.walk(def.Body)
	c.postOpDef(def)
}

func (c *coverageCreator) preOpAppl(node *OpApplNode) {
	if c.isStandardModuleNode(node) {
		return
	}
	parent := c.peek()
	cm := parent.AddChild(node)
	location, _ := semanticNodeSourceLocation(node)
	if c.primed[location] {
		cm.SetPrimed()
	}
	if node.Operator != nil && node.Operator.Name != nil && GetOpCode(node.Operator.Name) == OpcodeUnchanged {
		cm.MarkUnchanged()
	}
	c.attachLetAlias(node, cm)
	if def, ok := c.lookupToolOpDef(node); ok && !sameSymbol(def.Symbol, node.Operator) {
		cm.AddChildModel(c.createSubstitutionChild(def.Body))
	}
	if def, ok := c.lookupToolOpDef(node); ok && def.GetInRecursive() {
		if prior := c.findRecursiveWrapper(def); prior.node != nil {
			cm.SetRecursive(prior)
		}
	}
	if def, ok := c.lookupToolOpDef(node); ok && sameSymbol(def.Symbol, node.Operator) && GetOpCode(node.Operator.Name) == 0 && GetOpCode(def.Name) == 0 && c.argsContainOpArgNodes(node) && !def.IsStandardModule() {
		c.ctx = c.coverageOpContext(def, node.Args, c.ctx)
	}
	if def, ok := c.lookupContextOpDef(node); ok {
		if body, ok := def.Body.(*OpApplNode); ok {
			key := newSemanticNodeKey(body)
			if !costModelSliceContains(c.node2Wrapper[key], cm) {
				c.node2Wrapper[key] = append(c.node2Wrapper[key], cm)
			}
		}
	}
	if wrappers := c.node2Wrapper[newSemanticNodeKey(node)]; len(wrappers) > 0 {
		for _, wrapper := range wrappers {
			if wrapper.node != nil && wrapper.node != cm.node {
				wrapper.AddChildModel(cm)
			}
		}
	}
	if subst, ok := c.substs[newSemanticNodeKey(node)]; ok {
		c.root.PutSubst(subst, cm)
	}
	c.stack = append(c.stack, cm)
	if def, ok := c.lookupToolOpDef(node); ok && sameSymbol(def.Symbol, node.Operator) {
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

func (c *coverageCreator) preOpDef(def *OpDefNode) {
	if def != nil {
		c.opDefNodes[def] = true
	}
}

func (c *coverageCreator) postOpDef(def *OpDefNode) {
	if def != nil {
		delete(c.opDefNodes, def)
	}
}

func (c *coverageCreator) attachLetAlias(node *OpApplNode, cm CostModel) {
	in, ok := c.letIns[newSemanticNodeKey(node)]
	if !ok {
		return
	}
	for _, candidate := range c.stack {
		if candidate.node != nil && sameSemanticNode(candidate.node.Expr, in) {
			candidate.AddLet(node, cm)
		}
	}
}

func (c *coverageCreator) createSubstitutionChild(body SemanticNode) CostModel {
	bodyAppl, ok := body.(*OpApplNode)
	if !ok || bodyAppl == nil {
		return DoNotRecordCostModel
	}
	sub := &coverageCreator{
		tool:         c.tool,
		primed:       make(map[SourceLocation]bool),
		root:         c.root,
		ctx:          EmptyContext,
		visiting:     make(map[semanticNodeKey]bool),
		opDefNodes:   make(map[*OpDefNode]bool),
		substs:       make(map[semanticNodeKey]Subst),
		node2Wrapper: make(map[semanticNodeKey][]CostModel),
		letIns:       make(map[semanticNodeKey]SemanticNode),
	}
	sentinel := NewCostModel(nil)
	sub.stack = append(sub.stack, sentinel)
	sub.walk(bodyAppl)
	child := sentinel.GetChild()
	if child.node == sentinel.node {
		return DoNotRecordCostModel
	}
	return child
}

func (c *coverageCreator) findRecursiveWrapper(def *OpDefNode) CostModel {
	for i := len(c.stack) - 1; i >= 0; i-- {
		candidate := c.stack[i]
		if candidate.node == nil {
			continue
		}
		appl, ok := candidate.node.Expr.(*OpApplNode)
		if !ok || appl == nil {
			continue
		}
		if sameSymbol(appl.Operator, def.Symbol) {
			return candidate
		}
		if lookedUp, ok := c.lookupToolOpDef(appl); ok && lookedUp == def {
			return candidate
		}
	}
	return DoNotRecordCostModel
}

func (c *coverageCreator) argsContainOpArgNodes(node *OpApplNode) bool {
	if node == nil {
		return false
	}
	for _, arg := range node.Args {
		if _, ok := arg.(*OpArgNode); ok {
			return true
		}
	}
	return false
}

func (c *coverageCreator) coverageOpContext(def *OpDefNode, args []SemanticNode, base *Context) *Context {
	if def == nil {
		return base
	}
	if base == nil {
		base = EmptyContext
	}
	c1 := base
	limit := len(def.Params)
	if len(args) < limit {
		limit = len(args)
	}
	for i := 0; i < limit; i++ {
		c1 = c1.Cons(def.Params[i], c.coverageVal(args[i], base))
	}
	return c1
}

func (c *coverageCreator) coverageVal(expr SemanticNode, con *Context) any {
	if opArg, ok := expr.(*OpArgNode); ok {
		if c != nil && c.tool != nil {
			if val := c.tool.Lookup(opArg.Op, con, EmptyState, false); val != nil {
				return val
			}
		}
		return opArg.Op
	}
	return NewLazyValue(expr, con, false, DoNotRecordCostModel)
}

func (c *coverageCreator) lookupToolOpDef(node *OpApplNode) (*OpDefNode, bool) {
	if c == nil || c.tool == nil || node == nil || node.Operator == nil {
		return nil, false
	}
	val := c.tool.Lookup(node.Operator, EmptyContext, EmptyState, false)
	def, ok := val.(*OpDefNode)
	return def, ok && def != nil
}

func (c *coverageCreator) lookupContextOpDef(node *OpApplNode) (*OpDefNode, bool) {
	if c == nil || node == nil || node.Operator == nil || c.ctx == nil {
		return nil, false
	}
	def, ok := c.ctx.Lookup(node.Operator).(*OpDefNode)
	return def, ok && def != nil
}

func (c *coverageCreator) isStandardModuleNode(node *OpApplNode) bool {
	return node != nil && node.IsStandardModule()
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

func sameSymbol(left *SymbolNode, right *SymbolNode) bool {
	if left == right {
		return true
	}
	if left == nil || right == nil || left.Name == nil || right.Name == nil {
		return false
	}
	return left.Name == right.Name || left.Name.String() == right.Name.String()
}

func costModelSliceContains(models []CostModel, target CostModel) bool {
	for _, model := range models {
		if model.node == target.node {
			return true
		}
	}
	return false
}
