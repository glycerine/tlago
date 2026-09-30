package tlc

import (
	"fmt"
	"strings"
)

type TBTriple struct {
	A *LiveExprNode
	B *LiveExprNode
	C *LiveExprNode
}

func NewTBTriple(a, b, c *LiveExprNode) TBTriple {
	return TBTriple{A: a, B: b, C: c}
}

type TBPar struct {
	Exprs []*LiveExprNode
}

func NewTBPar(capacity int) *TBPar {
	if capacity < 0 {
		capacity = 0
	}
	return &TBPar{Exprs: make([]*LiveExprNode, 0, capacity)}
}

func (p *TBPar) Size() int {
	if p == nil {
		return 0
	}
	return len(p.Exprs)
}

func (p *TBPar) ExprAt(i int) *LiveExprNode {
	return p.Exprs[i]
}

func (p *TBPar) AddElement(expr *LiveExprNode) {
	p.Exprs = append(p.Exprs, expr)
}

func (p *TBPar) RemoveLastElement() *LiveExprNode {
	if p == nil || len(p.Exprs) == 0 {
		return nil
	}
	last := p.Exprs[len(p.Exprs)-1]
	p.Exprs = p.Exprs[:len(p.Exprs)-1]
	return last
}

func (p *TBPar) Equals(other *TBPar) bool {
	return p.Contains(other) && other.Contains(p)
}

func (p *TBPar) Member(expr *LiveExprNode) bool {
	if p == nil {
		return false
	}
	for _, candidate := range p.Exprs {
		if expr.Equal(candidate) {
			return true
		}
	}
	return false
}

func (p *TBPar) Contains(other *TBPar) bool {
	if other == nil {
		return true
	}
	for _, expr := range other.Exprs {
		if !p.Member(expr) {
			return false
		}
	}
	return true
}

func (p *TBPar) Append(exprs ...*LiveExprNode) *TBPar {
	out := NewTBPar(p.Size() + len(exprs))
	if p != nil {
		out.Exprs = append(out.Exprs, p.Exprs...)
	}
	out.Exprs = append(out.Exprs, exprs...)
	return out
}

func (p *TBPar) Union(other *TBPar) *TBPar {
	out := NewTBPar(p.Size() + other.Size())
	for _, expr := range p.Exprs {
		if !other.Member(expr) {
			out.AddElement(expr)
		}
	}
	for _, expr := range other.Exprs {
		out.AddElement(expr)
	}
	return out
}

func (p *TBPar) ParticleClosure() *TBParVec {
	positiveClosure := p.PositiveClosure()
	alphas := positiveClosure.AlphaTriples()
	betas := positiveClosure.BetaTriples()
	return particleClosure(p, alphas, betas)
}

func particleClosure(terms *TBPar, alphas []TBTriple, betas []TBTriple) *TBParVec {
	if !terms.IsLocallyConsistent() {
		return NewTBParVec(0)
	}
	terms1 := terms
	for i := 0; i < terms1.Size(); i++ {
		ln := terms1.ExprAt(i)
		var kappa1, kappa2 *LiveExprNode
		switch ln.Kind {
		case LiveExprAll:
			kappa1 = ln.Body
			kappa2 = NewLNNext(ln)
		case LiveExprConj:
			if ln.Count() >= 2 {
				kappa1 = ln.GetBody(0)
				kappa2 = ln.GetBody(1)
			}
		}
		if kappa1 != nil {
			switch {
			case terms1.Member(kappa1):
				if !terms1.Member(kappa2) {
					terms1 = terms1.Append(kappa2)
				}
			case terms1.Member(kappa2):
				terms1 = terms1.Append(kappa1)
			default:
				terms1 = terms1.Append(kappa1, kappa2)
			}
		}
	}
	for {
		done := true
		for _, alpha := range alphas {
			if terms1.Member(alpha.B) && terms1.Member(alpha.C) && !terms1.Member(alpha.A) {
				terms1.AddElement(alpha.A)
				done = false
			}
		}
		if done {
			break
		}
	}
	if terms1.Size() > terms.Size() && !terms1.IsLocallyConsistent() {
		return NewTBParVec(0)
	}
	return particleClosureBeta(terms1, alphas, betas)
}

func particleClosureBeta(terms *TBPar, alphas []TBTriple, betas []TBTriple) *TBParVec {
	for _, ln := range terms.Exprs {
		var kappa1, kappa2 *LiveExprNode
		switch ln.Kind {
		case LiveExprEven:
			kappa1 = ln.Body
			kappa2 = NewLNNext(ln)
		case LiveExprDisj:
			if ln.Count() >= 2 {
				kappa1 = ln.GetBody(0)
				kappa2 = ln.GetBody(1)
			}
		}
		if kappa1 != nil && !terms.Member(kappa1) && !terms.Member(kappa2) {
			ps1 := particleClosure(terms.Append(kappa1), alphas, betas)
			ps2 := particleClosure(terms.Append(kappa2), alphas, betas)
			return ps1.Union(ps2)
		}
	}
	for _, beta := range betas {
		if (terms.Member(beta.B) || terms.Member(beta.C)) && !terms.Member(beta.A) {
			return particleClosure(terms.Append(beta.A), alphas, betas)
		}
	}
	out := NewTBParVec(1)
	out.AddElement(terms)
	return out
}

func (p *TBPar) AlphaTriples() []TBTriple {
	triples := make([]TBTriple, 0)
	for _, ln := range p.Exprs {
		switch ln.Kind {
		case LiveExprAll:
			triples = append(triples, NewTBTriple(ln, ln.Body, NewLNNext(ln)))
		case LiveExprConj:
			if ln.Count() >= 2 {
				triples = append(triples, NewTBTriple(ln, ln.GetBody(0), ln.GetBody(1)))
			}
		}
	}
	return triples
}

func (p *TBPar) BetaTriples() []TBTriple {
	triples := make([]TBTriple, 0)
	for _, ln := range p.Exprs {
		switch ln.Kind {
		case LiveExprEven:
			triples = append(triples, NewTBTriple(ln, ln.Body, NewLNNext(ln)))
		case LiveExprDisj:
			if ln.Count() >= 2 {
				triples = append(triples, NewTBTriple(ln, ln.GetBody(0), ln.GetBody(1)))
			}
		}
	}
	return triples
}

func (p *TBPar) IsLocallyConsistent() bool {
	pos := NewTBPar(p.Size())
	neg := NewTBPar(p.Size())
	for _, ln := range p.Exprs {
		if ln.Kind == LiveExprState || ln.Kind == LiveExprBool {
			pos.AddElement(ln)
		} else if ln.Kind == LiveExprNeg && ln.Body != nil && (ln.Body.Kind == LiveExprState || ln.Body.Kind == LiveExprBool) {
			neg.AddElement(ln.Body)
		}
	}
	for _, expr := range pos.Exprs {
		if neg.Member(expr) {
			return false
		}
	}
	return true
}

func (p *TBPar) PositiveClosure() *TBPar {
	tps := NewTBPar(p.Size() * 2)
	tps.Exprs = append(tps.Exprs, p.Exprs...)
	result := NewTBPar(p.Size() * 2)
	for tps.Size() > 0 {
		ln := tps.RemoveLastElement()
		switch ln.Kind {
		case LiveExprNeg:
			tps.AddElement(ln.Body)
		case LiveExprNext:
			result.AddElement(ln)
			tps.AddElement(ln.Body)
		case LiveExprEven, LiveExprAll:
			result.AddElement(ln)
			result.AddElement(NewLNNext(ln))
			tps.AddElement(ln.Body)
		case LiveExprConj, LiveExprDisj:
			for _, body := range ln.Bodies {
				tps.AddElement(body)
			}
			result.AddElement(ln)
		case LiveExprState, LiveExprBool:
			result.AddElement(ln)
		case LiveExprAction:
			panic(newTLCError(ECGeneral, "encountered action formula while building liveness tableau"))
		}
	}
	return result
}

func (p *TBPar) ImpliedSuccessors() *TBPar {
	successors := NewTBPar(p.Size())
	for _, ln := range p.Exprs {
		if ln.Kind == LiveExprNext {
			successors.AddElement(ln.Body)
		}
	}
	return successors
}

func (p *TBPar) IsFulfilling(promise *LiveExprNode) bool {
	return !p.Member(promise) || p.Member(promise.Body)
}

func (p *TBPar) String() string {
	parts := make([]string, len(p.Exprs))
	for i, expr := range p.Exprs {
		parts[i] = expr.String()
	}
	return "{" + strings.Join(parts, ",\n ") + "}"
}

func (p *TBPar) ToDotViz() string {
	parts := make([]string, len(p.Exprs))
	for i, expr := range p.Exprs {
		parts[i] = expr.ToDotViz()
	}
	return strings.ReplaceAll("{"+strings.Join(parts, ",\n")+"}", "\\", "\\\\")
}

type TBParVec struct {
	Pars []*TBPar
}

func NewTBParVec(capacity int) *TBParVec {
	if capacity < 0 {
		capacity = 0
	}
	return &TBParVec{Pars: make([]*TBPar, 0, capacity)}
}

func (v *TBParVec) Size() int { return len(v.Pars) }
func (v *TBParVec) ParAt(i int) *TBPar {
	return v.Pars[i]
}
func (v *TBParVec) AddElement(par *TBPar) {
	v.Pars = append(v.Pars, par)
}

func (v *TBParVec) Contains(par *TBPar) bool {
	for _, elem := range v.Pars {
		if par.Equals(elem) {
			return true
		}
	}
	return false
}

func (v *TBParVec) Union(other *TBParVec) *TBParVec {
	out := NewTBParVec(v.Size() + other.Size())
	for _, par := range v.Pars {
		if !other.Contains(par) {
			out.AddElement(par)
		}
	}
	for _, par := range other.Pars {
		out.AddElement(par)
	}
	return out
}

func (v *TBParVec) String() string {
	parts := make([]string, len(v.Pars))
	for i, par := range v.Pars {
		parts[i] = par.String()
	}
	return "{" + strings.Join(parts, ",\n ") + "}"
}

type TBGraph struct {
	TF      *LiveExprNode
	InitCnt int
	Nodes   []*TBGraphNode
}

func NewTBGraph(tf *LiveExprNode) *TBGraph {
	graph := &TBGraph{TF: tf}
	if tf == nil {
		return graph
	}
	initTerms := NewTBPar(1)
	initTerms.AddElement(tf)
	pars := initTerms.ParticleClosure()
	for _, par := range pars.Pars {
		graph.Nodes = append(graph.Nodes, NewTBGraphNode(par))
	}
	graph.InitCnt = len(graph.Nodes)
	for i := 0; i < len(graph.Nodes); i++ {
		src := graph.Nodes[i]
		imps := src.Par.ImpliedSuccessors()
		succs := imps.ParticleClosure()
		for _, par := range succs.Pars {
			dst := graph.findOrCreateNode(par)
			src.Nexts = append(src.Nexts, dst)
		}
	}
	for i, node := range graph.Nodes {
		node.Index = i
	}
	return graph
}

func (g *TBGraph) Size() int {
	if g == nil {
		return 0
	}
	return len(g.Nodes)
}

func (g *TBGraph) GetNode(i int) *TBGraphNode {
	return g.Nodes[i]
}

func (g *TBGraph) findOrCreateNode(par *TBPar) *TBGraphNode {
	for _, node := range g.Nodes {
		if par.Equals(node.Par) {
			return node
		}
	}
	node := NewTBGraphNode(par)
	g.Nodes = append(g.Nodes, node)
	return node
}

func (g *TBGraph) ToDotViz() string {
	var b strings.Builder
	b.WriteString("digraph TableauGraph {\n")
	b.WriteString("nodesep = 0.7\n")
	b.WriteString("rankdir=LR;\n")
	for _, node := range g.Nodes {
		b.WriteString(node.ToDotViz(node.Index < g.InitCnt))
	}
	b.WriteString("}")
	return b.String()
}

type TBGraphNode struct {
	Par        *TBPar
	Nexts      []*TBGraphNode
	Index      int
	StatePreds []*LiveExprNode
}

func NewTBGraphNode(par *TBPar) *TBGraphNode {
	node := &TBGraphNode{Par: par}
	for _, ln := range par.Exprs {
		if ln.GetLevel() <= LiveLevelState {
			node.StatePreds = append(node.StatePreds, ln)
		}
	}
	return node
}

func (n *TBGraphNode) NextSize() int { return len(n.Nexts) }
func (n *TBGraphNode) NextAt(i int) *TBGraphNode {
	return n.Nexts[i]
}

func (n *TBGraphNode) HasLink(target *TBGraphNode) bool {
	for _, next := range n.Nexts {
		if next == target {
			return true
		}
	}
	return false
}

func (n *TBGraphNode) IsConsistent(state *TLCStateMut, tool *Tool) (bool, error) {
	for _, pred := range n.StatePreds {
		ok, err := pred.Eval(tool, state, nil)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (n *TBGraphNode) IsAccepting() bool {
	return n.Par.Size() == 0 && len(n.Nexts) == 1 && n.Nexts[0] == n
}

func (n *TBGraphNode) ToDotViz(isInitNode bool) string {
	var b strings.Builder
	label := fmt.Sprintf("\"Id: %d\n%s\"", n.Index, n.Par.ToDotViz())
	b.WriteString(fmt.Sprintf("%d [label=%s]\n", n.Index, label))
	if isInitNode {
		b.WriteString("[style = filled]\n")
	}
	for _, successor := range n.Nexts {
		b.WriteString(fmt.Sprintf("%d -> %d\n", n.Index, successor.Index))
	}
	return b.String()
}

type PossibleErrorModel struct {
	EAAction []int
	AEState  []int
	AEAction []int
}

func NewPossibleErrorModel(aeAction []int, aeState []int, eaAction []int) *PossibleErrorModel {
	return &PossibleErrorModel{
		AEAction: append([]int(nil), aeAction...),
		AEState:  append([]int(nil), aeState...),
		EAAction: append([]int(nil), eaAction...),
	}
}

func (p *PossibleErrorModel) IsEmpty() bool {
	return p == nil || (len(p.EAAction) == 0 && len(p.AEState) == 0 && len(p.AEAction) == 0)
}

func (p *PossibleErrorModel) IsSatisfiedByStuttering(checkState []bool, checkAction *BitVector) bool {
	for _, idx := range p.AEState {
		if idx < 0 || idx >= len(checkState) || !checkState[idx] {
			return false
		}
	}
	for _, idx := range p.AEAction {
		if !checkAction.Get(idx) {
			return false
		}
	}
	for _, idx := range p.EAAction {
		if !checkAction.Get(idx) {
			return false
		}
	}
	return true
}

func (p *PossibleErrorModel) StringWithChecks(checkState []*LiveExprNode, checkAction []*LiveExprNode) string {
	var b strings.Builder
	p.writeString(&b, "", checkState, checkAction)
	return b.String()
}

func (p *PossibleErrorModel) writeString(b *strings.Builder, padding string, checkState []*LiveExprNode, checkAction []*LiveExprNode) {
	noPadding := true
	padding1 := padding + "       "
	write := func(prefix string, expr *LiveExprNode) {
		if noPadding {
			noPadding = false
		} else {
			b.WriteString(padding)
		}
		b.WriteString(prefix)
		expr.writeString(b, padding1)
		b.WriteByte('\n')
	}
	for _, idx := range p.EAAction {
		write("/\\ <>[]", checkAction[idx])
	}
	for _, idx := range p.AEState {
		write("/\\ []<>", checkState[idx])
	}
	for _, idx := range p.AEAction {
		write("/\\ []<>", checkAction[idx])
	}
}

type OrderOfSolution struct {
	Tableau              *TBGraph
	Promises             []*LiveExprNode
	ContainsBoxInPromise bool
	CheckState           []*LiveExprNode
	CheckAction          []*LiveExprNode
	PEMs                 []*PossibleErrorModel
}

func NewOrderOfSolution(tableau *TBGraph, promises []*LiveExprNode) *OrderOfSolution {
	out := &OrderOfSolution{Tableau: tableau, Promises: append([]*LiveExprNode(nil), promises...)}
	for _, promise := range promises {
		if containsBoxOperator(promise.Body) {
			out.ContainsBoxInPromise = true
			break
		}
	}
	return out
}

func (o *OrderOfSolution) HasTableau() bool { return o != nil && o.Tableau != nil }
func (o *OrderOfSolution) SetCheckState(values []*LiveExprNode) {
	o.CheckState = append([]*LiveExprNode(nil), values...)
}
func (o *OrderOfSolution) SetCheckAction(values []*LiveExprNode) {
	o.CheckAction = append([]*LiveExprNode(nil), values...)
}
func (o *OrderOfSolution) SetPEMs(values []*PossibleErrorModel) {
	o.PEMs = append([]*PossibleErrorModel(nil), values...)
}

func (o *OrderOfSolution) CheckStateValues(tool *Tool, state *TLCStateMut) ([]bool, error) {
	result := make([]bool, len(o.CheckState))
	for i, expr := range o.CheckState {
		ok, err := expr.Eval(tool, state, nil)
		if err != nil {
			return nil, err
		}
		result[i] = ok
	}
	return result, nil
}

func (o *OrderOfSolution) CheckActionValues(tool *Tool, state0 *TLCStateMut, state1 *TLCStateMut) ([]bool, error) {
	result := make([]bool, len(o.CheckAction))
	for i, expr := range o.CheckAction {
		ok, err := expr.Eval(tool, state0, state1)
		if err != nil {
			return nil, err
		}
		result[i] = ok
	}
	return result, nil
}

func (o *OrderOfSolution) CheckActionBitVector(tool *Tool, state0 *TLCStateMut, state1 *TLCStateMut, result *BitVector, offset int) (*BitVector, error) {
	if result == nil {
		result = NewBitVector(offset + len(o.CheckAction))
	}
	for i, expr := range o.CheckAction {
		ok, err := expr.Eval(tool, state0, state1)
		if err != nil {
			return nil, err
		}
		if ok {
			result.Set(offset + i)
		}
	}
	return result, nil
}

func (o *OrderOfSolution) HasEmptyPEM() bool {
	for _, pem := range o.PEMs {
		if pem.IsEmpty() {
			return true
		}
	}
	return false
}

func (o *OrderOfSolution) IsPEMSatisfiedByStuttering(checkState []bool, checkAction *BitVector) bool {
	for _, pem := range o.PEMs {
		if pem.IsSatisfiedByStuttering(checkState, checkAction) {
			return true
		}
	}
	return false
}

func (o *OrderOfSolution) HasEmptyPEMAndBoxFreePromises() bool {
	if o.ContainsBoxInPromise {
		return false
	}
	for _, pem := range o.PEMs {
		if !pem.IsEmpty() {
			return false
		}
	}
	return true
}

func (o *OrderOfSolution) String() string {
	if o == nil || len(o.PEMs) == 0 {
		return ""
	}
	var b strings.Builder
	padding := ""
	if o.HasTableau() {
		if len(o.PEMs) == 1 && o.PEMs[0].IsEmpty() {
			o.Tableau.TF.writeString(&b, "   ")
			return b.String()
		}
		b.WriteString("/\\ ")
		o.Tableau.TF.writeString(&b, "   ")
		b.WriteString("\n/\\ ")
		padding = "   "
	}
	if len(o.PEMs) == 1 {
		o.PEMs[0].writeString(&b, padding, o.CheckState, o.CheckAction)
		return b.String()
	}
	b.WriteString("\\/ ")
	padding1 := padding + "   "
	o.PEMs[0].writeString(&b, padding1, o.CheckState, o.CheckAction)
	for i := 1; i < len(o.PEMs); i++ {
		b.WriteString(padding)
		b.WriteString("\\/ ")
		o.PEMs[i].writeString(&b, padding1, o.CheckState, o.CheckAction)
	}
	return b.String()
}

func containsBoxOperator(node *LiveExprNode) bool {
	if node == nil {
		return false
	}
	if node.Kind == LiveExprAll {
		return true
	}
	if node.Body != nil && containsBoxOperator(node.Body) {
		return true
	}
	for _, body := range node.Bodies {
		if containsBoxOperator(body) {
			return true
		}
	}
	return false
}

type AbstractGraphNode struct {
	Checks *BitVector
}

func NewAbstractGraphNode(checks *BitVector) AbstractGraphNode {
	if checks == nil {
		checks = NewBitVector(0)
	}
	return AbstractGraphNode{Checks: checks}
}

func (n *AbstractGraphNode) GetCheckState(i int) bool {
	return n.Checks.Get(i)
}

func (n *AbstractGraphNode) GetCheckAction(slen int, alen int, nodeIdx int, i int) bool {
	return n.Checks.Get(slen + alen*nodeIdx + i)
}

func (n *AbstractGraphNode) GetCheckActionVector(slen int, alen int, nodeIdx int) *BitVector {
	out := NewBitVector(alen)
	for i := 0; i < alen; i++ {
		if n.GetCheckAction(slen, alen, nodeIdx, i) {
			out.Set(i)
		}
	}
	return out
}

func (n *AbstractGraphNode) GetCheckActionAll(slen int, alen int, nodeIdx int, indices []int) bool {
	for _, i := range indices {
		if !n.GetCheckAction(slen, alen, nodeIdx, i) {
			return false
		}
	}
	return true
}

func (n *AbstractGraphNode) SetCheckState(values []bool) {
	for i, value := range values {
		if value {
			n.Checks.Set(i)
		}
	}
}

type GraphNode struct {
	AbstractGraphNode
	StateFP uint64
	TIndex  int
	Nodes   []int
	offset  int
}

const graphNodeRecordSize = 3
const graphNodeNoFreeSlots = -1

func NewGraphNode(fp uint64, tindex int) *GraphNode {
	return &GraphNode{AbstractGraphNode: NewAbstractGraphNode(NewBitVector(0)), StateFP: fp, TIndex: tindex, offset: graphNodeNoFreeSlots}
}

func (n *GraphNode) SuccSize() int {
	if n.offset != graphNodeNoFreeSlots {
		return n.offset / graphNodeRecordSize
	}
	return len(n.Nodes) / graphNodeRecordSize
}

func (n *GraphNode) GetStateFP(i int) uint64 {
	high := uint64(uint32(n.Nodes[graphNodeRecordSize*i]))
	low := uint64(uint32(n.Nodes[graphNodeRecordSize*i+1]))
	return (high << 32) | low
}

func (n *GraphNode) GetTIndex(i int) int {
	return n.Nodes[graphNodeRecordSize*i+2]
}

func (n *GraphNode) AddTransition(fp uint64, tidx int, slen int, alen int, acts *BitVector, actsOffset int, allocationHint int) {
	if acts != nil {
		pos := slen + alen*n.SuccSize()
		for i := 0; i < alen; i++ {
			if acts.Get(actsOffset + i) {
				n.Checks.Set(pos + i)
			}
		}
	}
	if n.offset == graphNodeNoFreeSlots {
		n.allocate(max(allocationHint, 1))
	}
	n.Nodes[n.offset] = int(uint32(fp >> 32))
	n.Nodes[n.offset+1] = int(uint32(fp))
	n.Nodes[n.offset+2] = tidx
	n.offset += graphNodeRecordSize
	if n.offset == len(n.Nodes) {
		n.offset = graphNodeNoFreeSlots
	}
}

func (n *GraphNode) allocate(transitions int) {
	oldLen := len(n.Nodes)
	next := make([]int, oldLen+graphNodeRecordSize*transitions)
	copy(next, n.Nodes)
	n.Nodes = next
	n.offset = oldLen
}

func (n *GraphNode) Realign() int {
	if n.offset == graphNodeNoFreeSlots {
		return 0
	}
	result := (len(n.Nodes) - n.offset) / graphNodeRecordSize
	next := make([]int, n.offset)
	copy(next, n.Nodes)
	n.Nodes = next
	n.offset = graphNodeNoFreeSlots
	return result
}

func (n *GraphNode) TransExists(fp uint64, tidx int) bool {
	limit := len(n.Nodes)
	if n.offset != graphNodeNoFreeSlots {
		limit = n.offset
	}
	high := int(uint32(fp >> 32))
	low := int(uint32(fp))
	for i := 0; i < limit; i += graphNodeRecordSize {
		if n.Nodes[i] == high && n.Nodes[i+1] == low && n.Nodes[i+2] == tidx {
			return true
		}
	}
	return false
}

func (n *GraphNode) CheckInvariants(slen int, alen int) bool {
	seen := make(map[GraphTransition]struct{}, n.SuccSize())
	for _, transition := range n.GetTransitions(slen, alen) {
		if _, ok := seen[transition]; ok {
			return false
		}
		seen[transition] = struct{}{}
	}
	return len(seen) == n.SuccSize()
}

func (n *GraphNode) GetTransitions(slen int, alen int) []GraphTransition {
	transitions := make([]GraphTransition, 0, n.SuccSize())
	for i := 0; i < n.SuccSize(); i++ {
		transitions = append(transitions, NewGraphTransition(n.GetStateFP(i), n.GetTIndex(i), n.GetCheckActionVector(slen, alen, i)))
	}
	return transitions
}

func (n *GraphNode) GetTNode(tableau *TBGraph) *TBGraphNode {
	if tableau == nil {
		return nil
	}
	return tableau.GetNode(n.TIndex)
}

func (n *GraphNode) Write(out *ValueOutputStream) error {
	n.Realign()
	if err := out.WriteNat(int32(len(n.Nodes))); err != nil {
		return err
	}
	for _, value := range n.Nodes {
		if err := out.WriteInt(int32(value)); err != nil {
			return err
		}
	}
	return n.Checks.Write(out)
}

func (n *GraphNode) Read(in *ValueInputStream) error {
	count, err := in.ReadNat()
	if err != nil {
		return err
	}
	n.Nodes = make([]int, int(count))
	for i := range n.Nodes {
		value, err := in.ReadInt()
		if err != nil {
			return err
		}
		n.Nodes[i] = int(value)
	}
	n.Checks = NewBitVector(0)
	if err := n.Checks.Read(in); err != nil {
		return err
	}
	n.offset = graphNodeNoFreeSlots
	return nil
}

func (n *GraphNode) String() string {
	return strings.ReplaceAll(n.StringWithActionLength(0), "[] ", "")
}

func (n *GraphNode) StringWithActionLength(alen int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<%d,%d> --> ", n.StateFP, n.TIndex))
	for i := 0; i < n.SuccSize(); i++ {
		b.WriteByte('[')
		for j := 0; j < alen; j++ {
			if n.GetCheckAction(0, 2, i, j) {
				b.WriteByte('t')
			} else {
				b.WriteByte('f')
			}
		}
		b.WriteString(fmt.Sprintf("] <%d,%d>, ", n.GetStateFP(i), n.GetTIndex(i)))
	}
	out := b.String()
	if strings.HasSuffix(out, ", ") {
		return out[:len(out)-2]
	}
	return out
}

func (n *GraphNode) ToDotViz(isInitState bool, hasTableau bool, slen int, alen int, oos *OrderOfSolution, labels map[uint64]string) string {
	id := fmt.Sprint(n.StateFP)
	if hasTableau {
		id += fmt.Sprintf(".%d", n.TIndex)
	}
	labelPrefix := ""
	if labels != nil {
		labelPrefix = labels[n.StateFP]
	}
	fpLabel := fmt.Sprint(n.StateFP)
	if len(fpLabel) > 6 {
		fpLabel = fpLabel[:6]
	}
	label := labelPrefix + fpLabel
	if hasTableau {
		label += fmt.Sprintf(".%d", n.TIndex)
	}
	if slen > 0 {
		label += "\n"
		for i := 0; i < slen; i++ {
			if n.GetCheckState(i) {
				label += "t"
			} else {
				label += "f"
			}
		}
	}
	if oos != nil && len(oos.Promises) > 0 && hasTableau && oos.Tableau != nil {
		label += "\n"
		tnode := n.GetTNode(oos.Tableau)
		for _, promise := range oos.Promises {
			if tnode != nil && tnode.Par.IsFulfilling(promise) {
				label += "t"
			} else {
				label += "f"
			}
		}
	}
	var b strings.Builder
	if isInitState {
		b.WriteString(fmt.Sprintf("%q [style = filled][label = %q]\n", id, label))
	} else {
		b.WriteString(fmt.Sprintf("%q [label = %q]\n", id, label))
	}
	for i := 0; i < n.SuccSize(); i++ {
		stateFP := n.GetStateFP(i)
		tidx := n.GetTIndex(i)
		target := fmt.Sprint(stateFP)
		if hasTableau {
			target += fmt.Sprintf(".%d", tidx)
		}
		b.WriteString(fmt.Sprintf("%q -> %q [label=\"", id, target))
		for j := 0; j < alen; j++ {
			if n.GetCheckAction(slen, alen, i, j) {
				b.WriteByte('t')
			} else {
				b.WriteByte('f')
			}
		}
		b.WriteString("\"];\n")
	}
	return b.String()
}

type GraphTransition struct {
	FP         uint64
	TIndex     int
	checksHash int
	checksText string
}

func NewGraphTransition(fp uint64, tidx int, checks *BitVector) GraphTransition {
	text := ""
	hash := 0
	if checks != nil {
		text = checks.String()
		hash = checks.Hash()
	}
	return GraphTransition{FP: fp, TIndex: tidx, checksHash: hash, checksText: text}
}

type BEGraphNode struct {
	AbstractGraphNode
	StateFP uint64
	Nexts   []*BEGraphNode
	Number  int64
}

const beGraphVisitedMask = int64(-1 << 63)

func NewBEGraphNode(fp uint64) *BEGraphNode {
	return &BEGraphNode{AbstractGraphNode: NewAbstractGraphNode(NewBitVector(0)), StateFP: fp}
}

func (n *BEGraphNode) NextAt(i int) *BEGraphNode { return n.Nexts[i] }
func (n *BEGraphNode) NextSize() int             { return len(n.Nexts) }

func (n *BEGraphNode) ResetNumberField() int64 {
	old := n.Number
	n.Number = 0
	return old
}

func (n *BEGraphNode) GetNumber() int64 { return n.Number & 0x7fffffffffffffff }
func (n *BEGraphNode) IncNumber()       { n.Number++ }
func (n *BEGraphNode) SetNumber(num int64) {
	if n.Number < 0 {
		n.Number = num | beGraphVisitedMask
	} else {
		n.Number = num
	}
}

type BTGraphNode struct {
	*BEGraphNode
	TIndex int32
}

func NewBTGraphNode(fp uint64, index int) *BTGraphNode {
	return &BTGraphNode{BEGraphNode: NewBEGraphNode(fp), TIndex: int32(index)}
}

func (n *BTGraphNode) GetIndex() int {
	if n == nil {
		return 0
	}
	return int(uint32(n.TIndex) & 0x3fffffff)
}

func (n *BTGraphNode) SetIndex(index int) {
	if n != nil {
		n.TIndex = int32((uint32(n.TIndex) & 0xc0000000) | (uint32(index) & 0x3fffffff))
	}
}

func (n *BTGraphNode) IsDone() bool {
	return n != nil && n.TIndex < 0
}

func (n *BTGraphNode) SetDone() {
	if n != nil {
		n.TIndex = int32(uint32(n.TIndex) | 0x80000000)
	}
}

func NewDummyBTGraphNode(fp uint64) *BTGraphNode {
	return NewBTGraphNode(fp, 0x40000000)
}

func (n *BTGraphNode) IsDummy() bool {
	return n != nil && (uint32(n.TIndex)&0x40000000) != 0
}

func (n *BTGraphNode) GetTNode(tableau *TBGraph) *TBGraphNode {
	if n == nil || tableau == nil {
		return nil
	}
	return tableau.GetNode(n.GetIndex())
}

func (n *BTGraphNode) NodeInfo() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("<%d,%d>", n.StateFP, n.GetIndex())
}
func (n *BEGraphNode) GetVisited() bool { return n.Number < 0 }
func (n *BEGraphNode) FlipVisited()     { n.Number ^= beGraphVisitedMask }
func (n *BEGraphNode) AddTransition(target *BEGraphNode, slen int, alen int, acts []bool) {
	num := len(n.Nexts)
	if acts != nil {
		pos := slen + alen*num
		for i, value := range acts {
			if value {
				n.Checks.Set(pos + i)
			}
		}
	}
	n.Nexts = append(n.Nexts, target)
}

func (n *BEGraphNode) TransExists(target *BEGraphNode) bool {
	for _, next := range n.Nexts {
		if next == target || (next != nil && target != nil && next.StateFP == target.StateFP) {
			return true
		}
	}
	return false
}

func (n *BEGraphNode) NodeInfo() string {
	return fmt.Sprint(n.StateFP)
}

func (n *BEGraphNode) SetParent(parent *BEGraphNode) {
	if len(n.Nexts) == 0 {
		n.Nexts = make([]*BEGraphNode, 1)
	}
	n.Nexts[0] = parent
}

func (n *BEGraphNode) GetParent() *BEGraphNode {
	if len(n.Nexts) == 0 {
		return nil
	}
	return n.Nexts[0]
}

func (n *BEGraphNode) String() string {
	var b strings.Builder
	n.writeString(&b, !n.GetVisited())
	return b.String()
}

func (n *BEGraphNode) writeString(b *strings.Builder, unseen bool) {
	if n.GetVisited() != unseen {
		return
	}
	n.FlipVisited()
	b.WriteString(fmt.Sprintf("%d --> ", n.StateFP))
	if len(n.Nexts) != 0 && n.Nexts[0] != nil {
		b.WriteString(fmt.Sprint(n.Nexts[0].StateFP))
	}
	for i := 1; i < len(n.Nexts); i++ {
		if n.Nexts[i] == nil {
			continue
		}
		b.WriteString(", ")
		b.WriteString(fmt.Sprint(n.Nexts[i].StateFP))
	}
	b.WriteByte('\n')
	for _, next := range n.Nexts {
		if next != nil {
			next.writeString(b, unseen)
		}
	}
}
