package tlc

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

const livenessTableauExportPathProperty = "tlc2.tool.liveness.Liveness.tableauExportPath"

func livenessFPString(fp uint64) string {
	return strconv.FormatInt(int64(fp), 10)
}

type TBTriple struct {
	A              *LiveExprNode
	B              *LiveExprNode
	C              *LiveExprNode
	ownersCaptured bool
	aOwner         *LiveExprNode
	bOwner         *LiveExprNode
	cOwner         *LiveExprNode
}

func NewTBTriple(a, b, c *LiveExprNode) TBTriple {
	return TBTriple{A: a, B: b, C: c, ownersCaptured: true, aOwner: a, bOwner: b, cOwner: c}
}

func (t *TBTriple) GetA() *LiveExprNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownersCaptured {
		return t.aOwner
	}
	return t.A
}

func (t *TBTriple) GetB() *LiveExprNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownersCaptured {
		return t.bOwner
	}
	return t.B
}

func (t *TBTriple) GetC() *LiveExprNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownersCaptured {
		return t.cOwner
	}
	return t.C
}

func (t *TBTriple) IsAlpha() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	a := t.GetA()
	return a != nil && (a.ownedKind() == LiveExprConj || a.ownedKind() == LiveExprAll)
}

func (t *TBTriple) IsBeta() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	a := t.GetA()
	return a != nil && (a.ownedKind() == LiveExprDisj || a.ownedKind() == LiveExprEven)
}

func (t *TBTriple) String() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	return "<" + tbTripleExprString(t.GetA()) + ",\n " + tbTripleExprString(t.GetB()) + ",\n " + tbTripleExprString(t.GetC()) + ">"
}

func tbTripleExprString(expr *LiveExprNode) string {
	if expr == nil {
		return "null"
	}
	return expr.String()
}

type TBPar struct {
	Exprs          []*LiveExprNode
	ownersCaptured bool
	exprsOwner     []*LiveExprNode
}

func NewTBPar(capacity int) *TBPar {
	capacity = int(int32(capacity))
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmtInt(capacity)))
	}
	exprs := make([]*LiveExprNode, 0, capacity)
	return &TBPar{Exprs: exprs, ownersCaptured: true, exprsOwner: exprs}
}

func (p *TBPar) Size() int {
	if p == nil {
		panic(NewNullPointerException())
	}
	return len(p.ownedExprs())
}

func (p *TBPar) ExprAt(i int) *LiveExprNode {
	if p == nil {
		panic(NewNullPointerException())
	}
	i = int(int32(i))
	exprs := p.ownedExprs()
	if i < 0 || i >= cap(exprs) {
		panic(NewArrayIndexOutOfBoundsException(i, cap(exprs)))
	}
	return exprs[:cap(exprs)][i]
}

func (p *TBPar) AddElement(expr *LiveExprNode) {
	if p == nil {
		panic(NewNullPointerException())
	}
	exprs := p.ownedExprs()
	if len(exprs) == cap(exprs) {
		newCapacity := cap(exprs) * 2
		if newCapacity < len(exprs)+1 {
			newCapacity = len(exprs) + 1
		}
		grown := make([]*LiveExprNode, len(exprs), newCapacity)
		copy(grown, exprs)
		exprs = grown
	}
	p.setExprs(append(exprs, expr))
}

func (p *TBPar) RemoveLastElement() *LiveExprNode {
	if p == nil {
		panic(NewNullPointerException())
	}
	exprs := p.ownedExprs()
	if len(exprs) == 0 {
		panic(NewNoSuchElementException())
	}
	last := exprs[len(exprs)-1]
	exprs[:cap(exprs)][len(exprs)-1] = nil
	p.setExprs(exprs[:len(exprs)-1])
	return last
}

func (p *TBPar) Equals(other *TBPar) bool {
	if p == nil || other == nil {
		panic(NewNullPointerException())
	}
	return p.Contains(other) && other.Contains(p)
}

func (p *TBPar) Member(expr *LiveExprNode) bool {
	if p == nil {
		panic(NewNullPointerException())
	}
	for _, candidate := range p.ownedExprs() {
		if expr == nil {
			panic(NewNullPointerException())
		}
		if expr.Equal(candidate) {
			return true
		}
	}
	return false
}

func (p *TBPar) Contains(other *TBPar) bool {
	if p == nil || other == nil {
		panic(NewNullPointerException())
	}
	for _, expr := range other.ownedExprs() {
		if !p.Member(expr) {
			return false
		}
	}
	return true
}

func (p *TBPar) Append(exprs ...*LiveExprNode) *TBPar {
	if p == nil {
		panic(NewNullPointerException())
	}
	out := NewTBPar(p.Size() + len(exprs))
	out.setExprs(append(out.ownedExprs(), p.ownedExprs()...))
	out.setExprs(append(out.ownedExprs(), exprs...))
	return out
}

func (p *TBPar) Union(other *TBPar) *TBPar {
	if p == nil || other == nil {
		panic(NewNullPointerException())
	}
	out := NewTBPar(p.Size() + other.Size())
	for _, expr := range p.ownedExprs() {
		if !other.Member(expr) {
			out.AddElement(expr)
		}
	}
	for _, expr := range other.ownedExprs() {
		out.AddElement(expr)
	}
	return out
}

func (p *TBPar) ParticleClosure() *TBParVec {
	p.require()
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
		if ln == nil {
			continue
		}
		var kappa1, kappa2 *LiveExprNode
		switch ln.ownedKind() {
		case LiveExprAll:
			kappa1 = ln.GetBody()
			kappa2 = NewLNNext(ln)
		case LiveExprConj:
			kappa1 = ln.GetBody(0)
			kappa2 = ln.GetBody(1)
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
			if terms1.Member(alpha.GetB()) && terms1.Member(alpha.GetC()) && !terms1.Member(alpha.GetA()) {
				terms1.AddElement(alpha.GetA())
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
	for _, ln := range terms.ownedExprs() {
		if ln == nil {
			continue
		}
		var kappa1, kappa2 *LiveExprNode
		switch ln.ownedKind() {
		case LiveExprEven:
			kappa1 = ln.GetBody()
			kappa2 = NewLNNext(ln)
		case LiveExprDisj:
			kappa1 = ln.GetBody(0)
			kappa2 = ln.GetBody(1)
		}
		if kappa1 != nil && !terms.Member(kappa1) && !terms.Member(kappa2) {
			ps1 := particleClosure(terms.Append(kappa1), alphas, betas)
			ps2 := particleClosure(terms.Append(kappa2), alphas, betas)
			return ps1.Union(ps2)
		}
	}
	for _, beta := range betas {
		if (terms.Member(beta.GetB()) || terms.Member(beta.GetC())) && !terms.Member(beta.GetA()) {
			return particleClosure(terms.Append(beta.GetA()), alphas, betas)
		}
	}
	out := NewTBParVec(1)
	out.AddElement(terms)
	return out
}

func (p *TBPar) AlphaTriples() []TBTriple {
	p.require()
	triples := make([]TBTriple, 0)
	for _, ln := range p.ownedExprs() {
		if ln == nil {
			continue
		}
		switch ln.ownedKind() {
		case LiveExprAll:
			triples = append(triples, NewTBTriple(ln, ln.GetBody(), NewLNNext(ln)))
		case LiveExprConj:
			triples = append(triples, NewTBTriple(ln, ln.GetBody(0), ln.GetBody(1)))
		}
	}
	return triples
}

func (p *TBPar) BetaTriples() []TBTriple {
	p.require()
	triples := make([]TBTriple, 0)
	for _, ln := range p.ownedExprs() {
		if ln == nil {
			continue
		}
		switch ln.ownedKind() {
		case LiveExprEven:
			triples = append(triples, NewTBTriple(ln, ln.GetBody(), NewLNNext(ln)))
		case LiveExprDisj:
			triples = append(triples, NewTBTriple(ln, ln.GetBody(0), ln.GetBody(1)))
		}
	}
	return triples
}

func (p *TBPar) IsLocallyConsistent() bool {
	p.require()
	pos := NewTBPar(p.Size())
	neg := NewTBPar(p.Size())
	for _, ln := range p.ownedExprs() {
		if ln == nil {
			continue
		}
		if ln.ownedKind() == LiveExprState || ln.ownedKind() == LiveExprBool {
			pos.AddElement(ln)
		} else if ln.ownedKind() == LiveExprNeg && ln.GetBody() != nil && (ln.GetBody().ownedKind() == LiveExprState || ln.GetBody().ownedKind() == LiveExprBool) {
			neg.AddElement(ln.GetBody())
		}
	}
	for _, expr := range pos.ownedExprs() {
		if neg.Member(expr) {
			return false
		}
	}
	return true
}

func (p *TBPar) PositiveClosure() *TBPar {
	p.require()
	tps := NewTBPar(p.Size() * 2)
	tps.setExprs(append(tps.ownedExprs(), p.ownedExprs()...))
	result := NewTBPar(p.Size() * 2)
	for tps.Size() > 0 {
		ln := tps.RemoveLastElement()
		if ln == nil {
			panic(NewTLCRuntimeException(ECTLCLiveEncounteredActions))
		}
		switch ln.ownedKind() {
		case LiveExprNeg:
			tps.AddElement(ln.GetBody())
		case LiveExprNext:
			result.AddElement(ln)
			tps.AddElement(ln.GetBody())
		case LiveExprEven, LiveExprAll:
			result.AddElement(ln)
			result.AddElement(NewLNNext(ln))
			tps.AddElement(ln.GetBody())
		case LiveExprConj, LiveExprDisj:
			for _, body := range ln.ownedBodies() {
				tps.AddElement(body)
			}
			result.AddElement(ln)
		case LiveExprState, LiveExprBool:
			result.AddElement(ln)
		case LiveExprAction:
			panic(NewTLCRuntimeException(ECTLCLiveEncounteredActions))
		default:
			panic(NewTLCRuntimeException(ECTLCLiveEncounteredActions))
		}
	}
	return result
}

func (p *TBPar) ImpliedSuccessors() *TBPar {
	p.require()
	successors := NewTBPar(p.Size())
	for _, ln := range p.ownedExprs() {
		if ln == nil {
			continue
		}
		if ln.ownedKind() == LiveExprNext {
			successors.AddElement(ln.GetBody())
		}
	}
	return successors
}

func (p *TBPar) IsFulfilling(promise *LiveExprNode) bool {
	p.require()
	return !p.Member(promise) || p.Member(promise.GetBody())
}

func (p *TBPar) String() string {
	p.require()
	exprs := p.ownedExprs()
	parts := make([]string, len(exprs))
	for i, expr := range exprs {
		parts[i] = expr.String()
	}
	return "{" + strings.Join(parts, ",\n ") + "}"
}

func (p *TBPar) ToDotViz() string {
	p.require()
	exprs := p.ownedExprs()
	parts := make([]string, len(exprs))
	for i, expr := range exprs {
		parts[i] = expr.ToDotViz()
	}
	return strings.ReplaceAll("{"+strings.Join(parts, ",\n")+"}", "\\", "\\\\")
}

func (p *TBPar) require() {
	if p == nil {
		panic(NewNullPointerException())
	}
}

func (p *TBPar) ownedExprs() []*LiveExprNode {
	if p.ownersCaptured {
		return p.exprsOwner
	}
	return p.Exprs
}

func (p *TBPar) setExprs(exprs []*LiveExprNode) {
	p.Exprs = exprs
	if p.ownersCaptured {
		p.exprsOwner = exprs
	}
}

type TBParVec struct {
	Pars           []*TBPar
	ownersCaptured bool
	parsOwner      []*TBPar
}

func NewTBParVec(capacity int) *TBParVec {
	capacity = int(int32(capacity))
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmtInt(capacity)))
	}
	pars := make([]*TBPar, 0, capacity)
	return &TBParVec{Pars: pars, ownersCaptured: true, parsOwner: pars}
}

func (v *TBParVec) Size() int {
	if v == nil {
		panic(NewNullPointerException())
	}
	return len(v.ownedPars())
}
func (v *TBParVec) ParAt(i int) *TBPar {
	if v == nil {
		panic(NewNullPointerException())
	}
	i = int(int32(i))
	pars := v.ownedPars()
	if i < 0 || i >= cap(pars) {
		panic(NewArrayIndexOutOfBoundsException(i, cap(pars)))
	}
	return pars[:cap(pars)][i]
}
func (v *TBParVec) AddElement(par *TBPar) {
	if v == nil {
		panic(NewNullPointerException())
	}
	pars := v.ownedPars()
	if len(pars) == cap(pars) {
		newCapacity := cap(pars) * 2
		if newCapacity < len(pars)+1 {
			newCapacity = len(pars) + 1
		}
		grown := make([]*TBPar, len(pars), newCapacity)
		copy(grown, pars)
		pars = grown
	}
	v.setPars(append(pars, par))
}

func (v *TBParVec) Contains(par *TBPar) bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	for _, elem := range v.ownedPars() {
		if par.Equals(elem) {
			return true
		}
	}
	return false
}

func (v *TBParVec) Union(other *TBParVec) *TBParVec {
	if v == nil || other == nil {
		panic(NewNullPointerException())
	}
	out := NewTBParVec(v.Size() + other.Size())
	for _, par := range v.ownedPars() {
		if !other.Contains(par) {
			out.AddElement(par)
		}
	}
	for _, par := range other.ownedPars() {
		out.AddElement(par)
	}
	return out
}

func (v *TBParVec) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	pars := v.ownedPars()
	parts := make([]string, len(pars))
	for i, par := range pars {
		parts[i] = par.String()
	}
	return "{" + strings.Join(parts, ",\n ") + "}"
}

func (v *TBParVec) ownedPars() []*TBPar {
	if v.ownersCaptured {
		return v.parsOwner
	}
	return v.Pars
}

func (v *TBParVec) setPars(pars []*TBPar) {
	v.Pars = pars
	if v.ownersCaptured {
		v.parsOwner = pars
	}
}

type TBGraph struct {
	TF             *LiveExprNode
	InitCnt        int
	Nodes          []*TBGraphNode
	ownersCaptured bool
	tfOwner        *LiveExprNode
	initCntOwner   int
	nodesOwner     []*TBGraphNode
}

func NewTBGraph(tf *LiveExprNode) *TBGraph {
	graph := NewEmptyTBGraph()
	graph.setTF(tf)
	initTerms := NewTBPar(1)
	initTerms.AddElement(tf)
	pars := initTerms.ParticleClosure()
	for _, par := range pars.ownedPars() {
		graph.addNode(NewTBGraphNode(par))
	}
	graph.SetInitCnt(graph.Size())
	for i := 0; i < graph.Size(); i++ {
		src := graph.GetNode(i)
		imps := src.GetPar().ImpliedSuccessors()
		succs := imps.ParticleClosure()
		for _, par := range succs.ownedPars() {
			dst := graph.findOrCreateNode(par)
			src.addNext(dst)
		}
	}
	for i, node := range graph.ownedNodes() {
		node.SetIndex(i)
	}
	return graph
}

func NewEmptyTBGraph() *TBGraph {
	nodes := make([]*TBGraphNode, 0, 10)
	return &TBGraph{Nodes: nodes, ownersCaptured: true, nodesOwner: nodes}
}

func exportTableauGraphIfRequested(graph *TBGraph) {
	if graph == nil {
		return
	}
	if path, ok := tlcLookupSystemProperty(livenessTableauExportPathProperty); ok {
		data, err := javaCharsetEncode(graph.ToDotViz(), javaDefaultCharset(), false)
		if err != nil {
			panic(err)
		}
		filePath, err := ioUtilsTXTPath(path)
		if err != nil {
			panic(err)
		}
		options, err := ioUtilsFileOptionsFromEnums(nil)
		if err != nil {
			panic(err)
		}
		if err := ioUtilsTXTWriteFile(filePath, data, options); err != nil {
			_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
		}
	}
}

func (g *TBGraph) Size() int {
	g.requireTBGraph()
	return len(g.ownedNodes())
}

func (g *TBGraph) requireTBGraph() {
	if g == nil {
		panic(NewNullPointerException())
	}
}

func (g *TBGraph) setTF(tf *LiveExprNode) {
	g.TF = tf
	if g.ownersCaptured {
		g.tfOwner = tf
	}
}

func (g *TBGraph) ownedTF() *LiveExprNode {
	if g.ownersCaptured {
		return g.tfOwner
	}
	return g.TF
}

func (g *TBGraph) ownedNodes() []*TBGraphNode {
	if g.ownersCaptured {
		return g.nodesOwner
	}
	return g.Nodes
}

func (g *TBGraph) setNodes(nodes []*TBGraphNode) {
	g.Nodes = nodes
	if g.ownersCaptured {
		g.nodesOwner = nodes
	}
}

func (g *TBGraph) GetNode(i int) *TBGraphNode {
	g.requireTBGraph()
	i = int(int32(i))
	nodes := g.ownedNodes()
	if i < 0 || i >= cap(nodes) {
		panic(NewArrayIndexOutOfBoundsException(i, cap(nodes)))
	}
	return nodes[:cap(nodes)][i]
}

func (g *TBGraph) SetInitCnt(n int) {
	g.requireTBGraph()
	n = int(int32(n))
	g.InitCnt = n
	if g.ownersCaptured {
		g.initCntOwner = n
	}
}

func (g *TBGraph) GetInitCnt() int {
	g.requireTBGraph()
	if g.ownersCaptured {
		return g.initCntOwner
	}
	return g.InitCnt
}

func (g *TBGraph) findOrCreateNode(par *TBPar) *TBGraphNode {
	for _, node := range g.ownedNodes() {
		if par.Equals(node.GetPar()) {
			return node
		}
	}
	node := NewTBGraphNode(par)
	g.addNode(node)
	return node
}

func (g *TBGraph) addNode(node *TBGraphNode) {
	nodes := g.ownedNodes()
	if len(nodes) == cap(nodes) {
		newCapacity := cap(nodes) * 2
		if newCapacity < len(nodes)+1 {
			newCapacity = len(nodes) + 1
		}
		grown := make([]*TBGraphNode, len(nodes), newCapacity)
		copy(grown, nodes)
		nodes = grown
	}
	g.setNodes(append(nodes, node))
}

func (g *TBGraph) AddElement(node *TBGraphNode) {
	g.requireTBGraph()
	g.addNode(node)
}

func (g *TBGraph) ToDotViz() string {
	g.requireTBGraph()
	var b strings.Builder
	b.WriteString("digraph TableauGraph {\n")
	b.WriteString("nodesep = 0.7\n")
	b.WriteString("rankdir=LR;\n")
	for _, node := range g.ownedNodes() {
		b.WriteString(node.ToDotViz(node.GetIndex() < g.GetInitCnt()))
	}
	b.WriteString("}")
	return b.String()
}

func (g *TBGraph) String() string {
	g.requireTBGraph()
	var b strings.Builder
	for i, node := range g.ownedNodes() {
		fmt.Fprintf(&b, "Node %d.\n", i)
		b.WriteString(node.GetPar().String())
		b.WriteString(" --> ")
		for _, successor := range node.ownedNexts() {
			fmt.Fprintf(&b, "%d ", successor.GetIndex())
		}
		b.WriteByte('\n')
	}
	return b.String()
}

type TBGraphNode struct {
	Par             *TBPar
	Nexts           []*TBGraphNode
	Index           int
	StatePreds      []*LiveExprNode
	ownersCaptured  bool
	parOwner        *TBPar
	nextsOwner      []*TBGraphNode
	indexOwner      int
	statePredsOwner []*LiveExprNode
}

func NewTBGraphNode(par *TBPar) *TBGraphNode {
	if par == nil {
		panic(NewNullPointerException())
	}
	nexts := make([]*TBGraphNode, 0, 10)
	statePreds := make([]*LiveExprNode, 0, par.Size())
	for _, ln := range par.ownedExprs() {
		if ln == nil {
			panic(NewNullPointerException())
		}
		if ln.GetLevel() <= LiveLevelState {
			statePreds = append(statePreds, ln)
		}
	}
	return &TBGraphNode{
		Par: par, Nexts: nexts, StatePreds: statePreds,
		ownersCaptured: true, parOwner: par, nextsOwner: nexts,
		statePredsOwner: append([]*LiveExprNode(nil), statePreds...),
	}
}

func (n *TBGraphNode) requireTBGraphNode() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *TBGraphNode) ownedNexts() []*TBGraphNode {
	if n.ownersCaptured {
		return n.nextsOwner
	}
	return n.Nexts
}

func (n *TBGraphNode) setNexts(nexts []*TBGraphNode) {
	n.Nexts = nexts
	if n.ownersCaptured {
		n.nextsOwner = nexts
	}
}

func (n *TBGraphNode) ownedStatePreds() []*LiveExprNode {
	if n.ownersCaptured {
		return n.statePredsOwner
	}
	return n.StatePreds
}

func (n *TBGraphNode) GetIndex() int {
	n.requireTBGraphNode()
	if n.ownersCaptured {
		return n.indexOwner
	}
	return n.Index
}

func (n *TBGraphNode) SetIndex(i int) {
	n.requireTBGraphNode()
	i = int(int32(i))
	if i < 0 {
		panic(NewAssertionError())
	}
	n.Index = i
	if n.ownersCaptured {
		n.indexOwner = i
	}
}

func (n *TBGraphNode) GetPar() *TBPar {
	n.requireTBGraphNode()
	if n.ownersCaptured {
		return n.parOwner
	}
	return n.Par
}

func (n *TBGraphNode) NextSize() int {
	n.requireTBGraphNode()
	return len(n.ownedNexts())
}

func (n *TBGraphNode) NextAt(i int) *TBGraphNode {
	n.requireTBGraphNode()
	i = int(int32(i))
	nexts := n.ownedNexts()
	if i < 0 || i >= cap(nexts) {
		panic(NewArrayIndexOutOfBoundsException(i, cap(nexts)))
	}
	return nexts[:cap(nexts)][i]
}

func (n *TBGraphNode) addNext(next *TBGraphNode) {
	n.requireTBGraphNode()
	nexts := n.ownedNexts()
	if len(nexts) == cap(nexts) {
		newCapacity := cap(nexts) * 2
		if newCapacity < len(nexts)+1 {
			newCapacity = len(nexts) + 1
		}
		grown := make([]*TBGraphNode, len(nexts), newCapacity)
		copy(grown, nexts)
		nexts = grown
	}
	n.setNexts(append(nexts, next))
}

func (n *TBGraphNode) HasLink(target *TBGraphNode) bool {
	n.requireTBGraphNode()
	for _, next := range n.ownedNexts() {
		if next == target {
			return true
		}
	}
	return false
}

func (n *TBGraphNode) IsConsistent(state *TLCStateMut, tool *Tool) (bool, error) {
	n.requireTBGraphNode()
	for _, pred := range n.ownedStatePreds() {
		ok, err := pred.Eval(tool, state, nil)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (n *TBGraphNode) IsAccepting() bool {
	n.requireTBGraphNode()
	nexts := n.ownedNexts()
	return n.GetPar().Size() == 0 && len(nexts) == 1 && nexts[0] == n
}

func (n *TBGraphNode) String() string {
	n.requireTBGraphNode()
	var b strings.Builder
	n.writeString(&b, make(map[int]struct{}))
	return b.String()
}

func (n *TBGraphNode) writeString(b *strings.Builder, visited map[int]struct{}) {
	n.requireTBGraphNode()
	index := n.GetIndex()
	if _, ok := visited[index]; ok {
		return
	}
	visited[index] = struct{}{}
	b.WriteString(n.GetPar().String())
	for _, successor := range n.ownedNexts() {
		fmt.Fprintf(b, "%d ", successor.GetIndex())
	}
	b.WriteByte('\n')
	for _, successor := range n.ownedNexts() {
		successor.writeString(b, visited)
	}
}

func (n *TBGraphNode) ToDotViz(isInitNode bool) string {
	n.requireTBGraphNode()
	var b strings.Builder
	index := n.GetIndex()
	label := fmt.Sprintf("\"Id: %d\n%s\"", index, n.GetPar().ToDotViz())
	b.WriteString(fmt.Sprintf("%d [label=%s]\n", index, label))
	if isInitNode {
		b.WriteString("[style = filled]\n")
	}
	for _, successor := range n.ownedNexts() {
		b.WriteString(fmt.Sprintf("%d -> %d\n", index, successor.GetIndex()))
	}
	return b.String()
}

type PossibleErrorModel struct {
	EAAction       []int
	AEState        []int
	AEAction       []int
	ownersCaptured bool
	eaActionOwner  []int
	aeStateOwner   []int
	aeActionOwner  []int
}

func NewPossibleErrorModel(aeAction []int, aeState []int, eaAction []int) *PossibleErrorModel {
	return &PossibleErrorModel{
		AEAction: aeAction, AEState: aeState, EAAction: eaAction,
		ownersCaptured: true, aeActionOwner: aeAction,
		aeStateOwner: aeState, eaActionOwner: eaAction,
	}
}

func (p *PossibleErrorModel) ownedEAAction() []int {
	if p.ownersCaptured {
		return p.eaActionOwner
	}
	return p.EAAction
}

func (p *PossibleErrorModel) ownedAEState() []int {
	if p.ownersCaptured {
		return p.aeStateOwner
	}
	return p.AEState
}

func (p *PossibleErrorModel) ownedAEAction() []int {
	if p.ownersCaptured {
		return p.aeActionOwner
	}
	return p.AEAction
}

func (p *PossibleErrorModel) IsEmpty() bool {
	if p == nil || p.ownedEAAction() == nil {
		panic(NewNullPointerException())
	}
	if len(p.ownedEAAction()) != 0 {
		return false
	}
	if p.ownedAEState() == nil {
		panic(NewNullPointerException())
	}
	if len(p.ownedAEState()) != 0 {
		return false
	}
	if p.ownedAEAction() == nil {
		panic(NewNullPointerException())
	}
	return len(p.ownedAEAction()) == 0
}

func (p *PossibleErrorModel) IsSatisfiedByStuttering(checkState []bool, checkAction *BitVector) bool {
	if p == nil || p.ownedAEState() == nil {
		panic(NewNullPointerException())
	}
	for _, idx := range p.ownedAEState() {
		if checkState == nil {
			panic(NewNullPointerException())
		}
		if idx < 0 || idx >= len(checkState) {
			panic(NewArrayIndexOutOfBoundsException(idx, len(checkState)))
		}
		if !checkState[idx] {
			return false
		}
	}
	if p.ownedAEAction() == nil {
		panic(NewNullPointerException())
	}
	for _, idx := range p.ownedAEAction() {
		if checkAction == nil {
			panic(NewNullPointerException())
		}
		if !checkAction.Get(idx) {
			return false
		}
	}
	if p.ownedEAAction() == nil {
		panic(NewNullPointerException())
	}
	for _, idx := range p.ownedEAAction() {
		if checkAction == nil {
			panic(NewNullPointerException())
		}
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
	if p == nil || p.ownedEAAction() == nil {
		panic(NewNullPointerException())
	}
	noPadding := true
	padding1 := padding + "       "
	write := func(prefix string, expr *LiveExprNode) {
		if expr == nil {
			panic(NewNullPointerException())
		}
		if noPadding {
			noPadding = false
		} else {
			b.WriteString(padding)
		}
		b.WriteString(prefix)
		expr.writeString(b, padding1)
		b.WriteByte('\n')
	}
	at := func(values []*LiveExprNode, idx int) *LiveExprNode {
		if values == nil {
			panic(NewNullPointerException())
		}
		if idx < 0 || idx >= len(values) {
			panic(NewArrayIndexOutOfBoundsException(idx, len(values)))
		}
		return values[idx]
	}
	for _, idx := range p.ownedEAAction() {
		write("/\\ <>[]", at(checkAction, idx))
	}
	if p.ownedAEState() == nil {
		panic(NewNullPointerException())
	}
	for _, idx := range p.ownedAEState() {
		write("/\\ []<>", at(checkState, idx))
	}
	if p.ownedAEAction() == nil {
		panic(NewNullPointerException())
	}
	for _, idx := range p.ownedAEAction() {
		write("/\\ []<>", at(checkAction, idx))
	}
}

type OrderOfSolution struct {
	// Java LiveCheck synchronizes graph mutations on the OrderOfSolution.
	graphMu sync.Mutex

	Tableau              *TBGraph
	Promises             []*LiveExprNode
	ContainsBoxInPromise bool
	checkStateExprs      []*LiveExprNode
	checkActionExprs     []*LiveExprNode
	PEMs                 []*PossibleErrorModel

	ownersCaptured            bool
	tableauOwner              *TBGraph
	promisesOwner             []*LiveExprNode
	containsBoxInPromiseOwner bool
	checkStateOwner           []*LiveExprNode
	checkActionOwner          []*LiveExprNode
	pemsOwner                 []*PossibleErrorModel
}

func NewOrderOfSolution(tableau *TBGraph, promises []*LiveExprNode) *OrderOfSolution {
	if promises == nil {
		panic(NewNullPointerException())
	}
	out := &OrderOfSolution{
		Tableau: tableau, Promises: promises,
		ownersCaptured: true, tableauOwner: tableau, promisesOwner: promises,
	}
	for _, promise := range promises {
		if promise == nil {
			panic(NewNullPointerException())
		}
		if promise.ownedKind() != LiveExprEven {
			panic(valueStreamClassCast(promise, "tlc2.tool.liveness.LNEven"))
		}
		if containsBoxOperator(promise.GetBody()) {
			out.ContainsBoxInPromise = true
			out.containsBoxInPromiseOwner = true
			break
		}
	}
	return out
}

func (o *OrderOfSolution) requireOrderOfSolution() {
	if o == nil {
		panic(NewNullPointerException())
	}
}

func (o *OrderOfSolution) HasTableau() bool {
	o.requireOrderOfSolution()
	return o.ownedTableau() != nil
}

func (o *OrderOfSolution) GetTableau() *TBGraph {
	o.requireOrderOfSolution()
	return o.ownedTableau()
}

func (o *OrderOfSolution) GetPromises() []*LiveExprNode {
	o.requireOrderOfSolution()
	return o.ownedPromises()
}

func (o *OrderOfSolution) GetCheckState() []*LiveExprNode {
	o.requireOrderOfSolution()
	return o.ownedCheckState()
}

func (o *OrderOfSolution) GetCheckAction() []*LiveExprNode {
	o.requireOrderOfSolution()
	return o.ownedCheckAction()
}

func (o *OrderOfSolution) GetPems() []*PossibleErrorModel {
	o.requireOrderOfSolution()
	return o.ownedPEMs()
}

func (o *OrderOfSolution) GetPEMs() []*PossibleErrorModel {
	return o.GetPems()
}

func (o *OrderOfSolution) PrintPromises(out io.Writer) {
	o.requireOrderOfSolution()
	promises := o.ownedPromises()
	if promises == nil {
		panic(NewNullPointerException())
	}
	for _, promise := range promises {
		if promise == nil {
			panic(NewNullPointerException())
		}
		line := promise.String()
		if isNil(out) {
			panic(NewNullPointerException())
		}
		_, _ = fmt.Fprintln(out, line)
	}
}

func (o *OrderOfSolution) SetCheckState(values []*LiveExprNode) {
	o.requireOrderOfSolution()
	o.checkStateExprs = values
	if o.ownersCaptured {
		o.checkStateOwner = values
	}
}
func (o *OrderOfSolution) SetCheckAction(values []*LiveExprNode) {
	o.requireOrderOfSolution()
	o.checkActionExprs = values
	if o.ownersCaptured {
		o.checkActionOwner = values
	}
}
func (o *OrderOfSolution) SetPEMs(values []*PossibleErrorModel) {
	o.requireOrderOfSolution()
	o.PEMs = values
	if o.ownersCaptured {
		o.pemsOwner = values
	}
}

func (o *OrderOfSolution) CheckState(tool *Tool, state *TLCStateMut) ([]bool, error) {
	o.requireOrderOfSolution()
	checkState := o.ownedCheckState()
	if checkState == nil {
		panic(NewNullPointerException())
	}
	result := make([]bool, len(checkState))
	for i, expr := range checkState {
		if expr == nil {
			panic(NewNullPointerException())
		}
		ok, err := expr.Eval(tool, state, nil)
		if err != nil {
			return nil, err
		}
		result[i] = ok
	}
	return result, nil
}

func (o *OrderOfSolution) CheckStateValues(tool *Tool, state *TLCStateMut) ([]bool, error) {
	return o.CheckState(tool, state)
}

func (o *OrderOfSolution) CheckAction(tool *Tool, state0 *TLCStateMut, state1 *TLCStateMut) ([]bool, error) {
	o.requireOrderOfSolution()
	checkAction := o.ownedCheckAction()
	if checkAction == nil {
		panic(NewNullPointerException())
	}
	result := make([]bool, len(checkAction))
	for i, expr := range checkAction {
		if expr == nil {
			panic(NewNullPointerException())
		}
		ok, err := expr.Eval(tool, state0, state1)
		if err != nil {
			return nil, err
		}
		result[i] = ok
	}
	return result, nil
}

func (o *OrderOfSolution) CheckActionValues(tool *Tool, state0 *TLCStateMut, state1 *TLCStateMut) ([]bool, error) {
	return o.CheckAction(tool, state0, state1)
}

func (o *OrderOfSolution) CheckActionBitVector(tool *Tool, state0 *TLCStateMut, state1 *TLCStateMut, result *BitVector, offset int) (*BitVector, error) {
	o.requireOrderOfSolution()
	checkAction := o.ownedCheckAction()
	if checkAction == nil {
		panic(NewNullPointerException())
	}
	for i, expr := range checkAction {
		if expr == nil {
			panic(NewNullPointerException())
		}
		ok, err := expr.Eval(tool, state0, state1)
		if err != nil {
			return nil, err
		}
		if ok {
			if result == nil {
				panic(NewNullPointerException())
			}
			result.Set(offset + i)
		}
	}
	return result, nil
}

func (o *OrderOfSolution) HasEmptyPEM() bool {
	o.requireOrderOfSolution()
	pems := o.ownedPEMs()
	if pems == nil {
		panic(NewNullPointerException())
	}
	for _, pem := range pems {
		if pem.IsEmpty() {
			return true
		}
	}
	return false
}

func (o *OrderOfSolution) IsPEMSatisfiedByStuttering(checkState []bool, checkAction *BitVector) bool {
	o.requireOrderOfSolution()
	pems := o.ownedPEMs()
	if pems == nil {
		panic(NewNullPointerException())
	}
	for _, pem := range pems {
		if pem.IsSatisfiedByStuttering(checkState, checkAction) {
			return true
		}
	}
	return false
}

func (o *OrderOfSolution) HasEmptyPEMAndBoxFreePromises() bool {
	o.requireOrderOfSolution()
	if o.ownedContainsBoxInPromise() {
		return false
	}
	pems := o.ownedPEMs()
	if pems == nil {
		panic(NewNullPointerException())
	}
	for _, pem := range pems {
		if !pem.IsEmpty() {
			return false
		}
	}
	return true
}

func (o *OrderOfSolution) String() string {
	o.requireOrderOfSolution()
	pems := o.ownedPEMs()
	if pems == nil {
		panic(NewNullPointerException())
	}
	if len(pems) == 0 {
		return ""
	}
	tableau := o.ownedTableau()
	checkState := o.ownedCheckState()
	checkAction := o.ownedCheckAction()
	var b strings.Builder
	padding := ""
	if o.HasTableau() {
		if len(pems) == 1 && pems[0].IsEmpty() {
			tableau.ownedTF().writeString(&b, "   ")
			return b.String()
		}
		b.WriteString("/\\ ")
		tableau.ownedTF().writeString(&b, "   ")
		b.WriteString("\n/\\ ")
		padding = "   "
	}
	if len(pems) == 1 {
		pems[0].writeString(&b, padding, checkState, checkAction)
		return b.String()
	}
	b.WriteString("\\/ ")
	padding1 := padding + "   "
	pems[0].writeString(&b, padding1, checkState, checkAction)
	for i := 1; i < len(pems); i++ {
		b.WriteString(padding)
		b.WriteString("\\/ ")
		pems[i].writeString(&b, padding1, checkState, checkAction)
	}
	return b.String()
}

func (o *OrderOfSolution) ownedTableau() *TBGraph {
	if o.ownersCaptured {
		return o.tableauOwner
	}
	return o.Tableau
}

func (o *OrderOfSolution) ownedPromises() []*LiveExprNode {
	if o.ownersCaptured {
		return o.promisesOwner
	}
	return o.Promises
}

func (o *OrderOfSolution) ownedContainsBoxInPromise() bool {
	if o.ownersCaptured {
		return o.containsBoxInPromiseOwner
	}
	return o.ContainsBoxInPromise
}

func (o *OrderOfSolution) ownedCheckState() []*LiveExprNode {
	if o.ownersCaptured {
		return o.checkStateOwner
	}
	return o.checkStateExprs
}

func (o *OrderOfSolution) ownedCheckAction() []*LiveExprNode {
	if o.ownersCaptured {
		return o.checkActionOwner
	}
	return o.checkActionExprs
}

func (o *OrderOfSolution) ownedPEMs() []*PossibleErrorModel {
	if o.ownersCaptured {
		return o.pemsOwner
	}
	return o.PEMs
}

func containsBoxOperator(node *LiveExprNode) bool {
	if node == nil {
		return false
	}
	switch node.ownedKind() {
	case LiveExprAll:
		return true
	case LiveExprEven, LiveExprNeg, LiveExprNext:
		return containsBoxOperator(node.GetBody())
	case LiveExprConj, LiveExprDisj:
		for _, body := range node.ownedBodies() {
			if containsBoxOperator(body) {
				return true
			}
		}
	}
	return false
}

type AbstractGraphNode struct {
	Checks *BitVector

	ownersCaptured bool
	checksOwner    *BitVector
}

func NewAbstractGraphNode(checks *BitVector) AbstractGraphNode {
	return AbstractGraphNode{Checks: checks, ownersCaptured: true, checksOwner: checks}
}

func (n *AbstractGraphNode) GetCheckState(i int) bool {
	if n == nil || n.ownedChecks() == nil {
		panic(NewNullPointerException())
	}
	return n.ownedChecks().Get(graphInt(i))
}

func (n *AbstractGraphNode) GetCheckAction(slen int, alen int, nodeIdx int, i int) bool {
	if n == nil || n.ownedChecks() == nil {
		panic(NewNullPointerException())
	}
	pos := graphIntAdd(slen, graphIntMul(alen, nodeIdx))
	return n.ownedChecks().Get(graphIntAdd(pos, i))
}

func (n *AbstractGraphNode) GetCheckActionVector(slen int, alen int, nodeIdx int) *BitVector {
	if n == nil {
		panic(NewNullPointerException())
	}
	javaAlen := graphInt(alen)
	out := NewBitVector(javaAlen)
	for i := 0; i < javaAlen; i++ {
		if n.GetCheckAction(slen, javaAlen, nodeIdx, i) {
			out.Set(i)
		}
	}
	return out
}

func graphInt(value int) int {
	return int(int32(value))
}

func graphIntAdd(left, right int) int {
	return int(int32(left) + int32(right))
}

func graphIntMul(left, right int) int {
	return int(int32(left) * int32(right))
}

func (n *AbstractGraphNode) GetCheckActionAll(slen int, alen int, nodeIdx int, indices []int) bool {
	if n == nil || indices == nil {
		panic(NewNullPointerException())
	}
	for _, i := range indices {
		if !n.GetCheckAction(slen, alen, nodeIdx, i) {
			return false
		}
	}
	return true
}

func (n *AbstractGraphNode) SetCheckState(values []bool) {
	if n == nil || values == nil {
		panic(NewNullPointerException())
	}
	for i, value := range values {
		if value {
			if n.ownedChecks() == nil {
				panic(NewNullPointerException())
			}
			n.ownedChecks().Set(i)
		}
	}
}

func (n *AbstractGraphNode) ownedChecks() *BitVector {
	if n.ownersCaptured {
		return n.checksOwner
	}
	return n.Checks
}

func (n *AbstractGraphNode) setChecks(checks *BitVector) {
	n.Checks = checks
	if n.ownersCaptured {
		n.checksOwner = checks
	}
}

type GraphNode struct {
	AbstractGraphNode
	StateFP uint64
	TIndex  int
	Nodes   []int
	offset  int

	ownersCaptured bool
	stateFPOwner   uint64
	tIndexOwner    int
	nodesOwner     []int
}

const graphNodeRecordSize = 3
const graphNodeNoFreeSlots = -1

func NewGraphNode(fp uint64, tindex int) *GraphNode {
	tindex = graphInt(tindex)
	nodes := make([]int, 0)
	return &GraphNode{
		AbstractGraphNode: NewAbstractGraphNode(NewBitVector(0)),
		StateFP:           fp, TIndex: tindex, Nodes: nodes, offset: graphNodeNoFreeSlots,
		ownersCaptured: true, stateFPOwner: fp, tIndexOwner: tindex, nodesOwner: nodes,
	}
}

func (n *GraphNode) requireGraphNode() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *GraphNode) HashCode() int32 {
	if n == nil {
		panic(NewNullPointerException())
	}
	result := int32(1)
	stateFP := n.ownedStateFP()
	result = 31*result + int32(stateFP^(stateFP>>32))
	result = 31*result + int32(n.ownedTIndex())
	return result
}

func (n *GraphNode) Equals(other any) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	node, ok := other.(*GraphNode)
	return ok && node != nil && n.ownedStateFP() == node.ownedStateFP() && n.ownedTIndex() == node.ownedTIndex()
}

func (n *GraphNode) SuccSize() int {
	n.requireGraphNode()
	if n.offset != graphNodeNoFreeSlots {
		return n.offset / graphNodeRecordSize
	}
	return len(n.ownedNodes()) / graphNodeRecordSize
}

func (n *GraphNode) GetStateFP(i int) uint64 {
	n.requireGraphNode()
	loc := int(int32(graphNodeRecordSize * int32(i)))
	nodes := n.ownedNodes()
	high := uint64(uint32(graphNodeArrayGet(nodes, loc)))
	low := uint64(uint32(graphNodeArrayGet(nodes, int(int32(int32(loc)+1)))))
	return (high << 32) | low
}

func (n *GraphNode) GetTIndex(i int) int {
	n.requireGraphNode()
	loc := int(int32(graphNodeRecordSize * int32(i)))
	return graphNodeArrayGet(n.ownedNodes(), int(int32(int32(loc)+2)))
}

func (n *GraphNode) AddTransition(fp uint64, tidx int, slen int, alen int, acts *BitVector, actsOffset int, allocationHint int) {
	n.requireGraphNode()
	tidx = graphInt(tidx)
	slen = graphInt(slen)
	alen = graphInt(alen)
	actsOffset = graphInt(actsOffset)
	allocationHint = graphInt(allocationHint)
	if acts != nil {
		pos := graphIntAdd(slen, graphIntMul(alen, n.SuccSize()))
		for i := 0; i < alen; i++ {
			if acts.Get(graphIntAdd(actsOffset, i)) {
				n.ownedChecks().Set(graphIntAdd(pos, i))
			}
		}
	}
	if n.offset == graphNodeNoFreeSlots {
		n.allocate(max(allocationHint, 1))
	}
	nodes := n.ownedNodes()
	graphNodeArraySet(nodes, n.offset, int(int32(fp>>32)))
	graphNodeArraySet(nodes, int(int32(int32(n.offset)+1)), int(int32(fp)))
	graphNodeArraySet(nodes, int(int32(int32(n.offset)+2)), tidx)
	n.offset = int(int32(int32(n.offset) + graphNodeRecordSize))
	if n.offset == len(nodes) {
		n.offset = graphNodeNoFreeSlots
	}
}

func (n *GraphNode) allocate(transitions int) {
	n.requireGraphNode()
	nodes := n.ownedNodes()
	oldLen := len(nodes)
	newLen := int(int32(int32(oldLen) + graphNodeRecordSize*int32(transitions)))
	if newLen < 0 {
		panic(NewNegativeArraySizeException(fmtInt(newLen)))
	}
	next := make([]int, newLen)
	if newLen < oldLen {
		panic(NewArrayIndexOutOfBoundsException(oldLen, newLen))
	}
	copy(next, nodes)
	n.setNodes(next)
	n.offset = oldLen
}

func (n *GraphNode) Realign() int {
	n.requireGraphNode()
	if n.offset == graphNodeNoFreeSlots {
		return 0
	}
	nodes := n.ownedNodes()
	result := (len(nodes) - n.offset) / graphNodeRecordSize
	next := make([]int, n.offset)
	copy(next, nodes)
	n.setNodes(next)
	n.offset = graphNodeNoFreeSlots
	return result
}

func (n *GraphNode) TransExists(fp uint64, tidx int) bool {
	n.requireGraphNode()
	tidx = graphInt(tidx)
	nodes := n.ownedNodes()
	limit := len(nodes)
	if n.offset != graphNodeNoFreeSlots {
		limit = n.offset
	}
	high := int(int32(fp >> 32))
	low := int(int32(fp))
	for i := 0; i < limit; i += graphNodeRecordSize {
		if nodes[i] == high && nodes[i+1] == low && nodes[i+2] == tidx {
			return true
		}
	}
	return false
}

func (n *GraphNode) CheckInvariants(slen int, alen int) bool {
	n.requireGraphNode()
	return len(n.GetTransitions(slen, alen)) == n.SuccSize()
}

func (n *GraphNode) GetTransition() map[GraphTransition]struct{} {
	n.requireGraphNode()
	return n.GetTransitions(0, 0)
}

func (n *GraphNode) GetTransitions(slen int, alen int) map[GraphTransition]struct{} {
	n.requireGraphNode()
	transitions := make(map[GraphTransition]struct{}, n.SuccSize())
	for i := 0; i < n.SuccSize(); i++ {
		transition := NewGraphTransition(n.GetStateFP(i), n.GetTIndex(i), n.GetCheckActionVector(slen, alen, i))
		transitions[transition] = struct{}{}
	}
	return transitions
}

func (n *GraphNode) GetTNode(tableau *TBGraph) *TBGraphNode {
	if n == nil || tableau == nil {
		panic(NewNullPointerException())
	}
	return tableau.GetNode(n.ownedTIndex())
}

func (n *GraphNode) Write(out dataOutput) error {
	n.requireGraphNode()
	if n.offset != graphNodeNoFreeSlots {
		panic(NewAssertionError())
	}
	if isNil(out) {
		panic(NewNullPointerException())
	}
	nodes := n.ownedNodes()
	if err := out.WriteNat(int32(len(nodes))); err != nil {
		return err
	}
	for _, value := range nodes {
		if err := out.WriteInt(int32(value)); err != nil {
			return err
		}
	}
	return n.ownedChecks().Write(out)
}

func (n *GraphNode) Read(in dataInput) error {
	n.requireGraphNode()
	if isNil(in) {
		panic(NewNullPointerException())
	}
	count, err := in.ReadNat()
	if err != nil {
		return err
	}
	nodes := make([]int, int(count))
	for i := range nodes {
		value, err := in.ReadInt()
		if err != nil {
			return err
		}
		nodes[i] = int(value)
	}
	n.setNodes(nodes)
	n.setChecks(NewBitVector(0))
	if err := n.ownedChecks().Read(in); err != nil {
		return err
	}
	if n.offset != graphNodeNoFreeSlots {
		panic(NewAssertionError())
	}
	return nil
}

func (n *GraphNode) String() string {
	n.requireGraphNode()
	return strings.ReplaceAll(n.StringWithActionLength(0), "[] ", "")
}

func (n *GraphNode) StringWithActionLength(alen int) string {
	n.requireGraphNode()
	alen = graphInt(alen)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<%s,%d> --> ", livenessFPString(n.ownedStateFP()), n.ownedTIndex()))
	for i := 0; i < n.SuccSize(); i++ {
		b.WriteByte('[')
		for j := 0; j < alen; j++ {
			if n.GetCheckAction(0, 2, i, j) {
				b.WriteByte('t')
			} else {
				b.WriteByte('f')
			}
		}
		b.WriteString(fmt.Sprintf("] <%s,%d>, ", livenessFPString(n.GetStateFP(i)), n.GetTIndex(i)))
	}
	out := b.String()
	// Java unconditionally chops the presumed trailing comma and space.  For
	// an empty successor list this instead removes the final "> " from the
	// arrow, yielding the source's peculiar "--" suffix.
	return out[:len(out)-2]
}

func (n *GraphNode) ToDotViz(isInitState bool, hasTableau bool, slen int, alen int, oos *OrderOfSolution, labels map[uint64]string) string {
	return n.ToDotVizFiltered(isInitState, hasTableau, slen, alen, nil, oos, labels)
}

func (n *GraphNode) ToDotVizWithFilter(isInitState bool, hasTableau bool, slen int, alen int, filter *TableauNodePtrTable, oos *OrderOfSolution) string {
	// Java's corresponding overload accepts the filter but delegates without
	// it. LiveWorker calls this overload, so retain every outgoing edge.
	return n.ToDotVizFiltered(isInitState, hasTableau, slen, alen, nil, oos, map[uint64]string{})
}

func (n *GraphNode) ToDotVizFiltered(isInitState bool, hasTableau bool, slen int, alen int, filter *TableauNodePtrTable, oos *OrderOfSolution, labels map[uint64]string) string {
	n.requireGraphNode()
	slen = graphInt(slen)
	alen = graphInt(alen)
	if labels == nil {
		panic(NewNullPointerException())
	}
	stateFP := n.ownedStateFP()
	id := livenessFPString(stateFP)
	if hasTableau {
		id += fmt.Sprintf(".%d", n.ownedTIndex())
	}
	labelPrefix := labels[stateFP]
	fpLabel := livenessFPString(stateFP)
	if len(fpLabel) < 6 {
		panic(NewStringIndexOutOfBoundsException(6, len(fpLabel)))
	}
	fpLabel = fpLabel[:6]
	label := labelPrefix + fpLabel
	if hasTableau {
		label += fmt.Sprintf(".%d", n.ownedTIndex())
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
	promises := oos.GetPromises()
	if len(promises) > 0 {
		label += "\n"
		for _, promise := range promises {
			par := n.GetTNode(oos.GetTableau()).GetPar()
			if par.IsFulfilling(promise) {
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
		if filter != nil && filter.Get(stateFP, tidx) == -1 {
			continue
		}
		target := livenessFPString(stateFP)
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

func (n *GraphNode) ownedStateFP() uint64 {
	if n.ownersCaptured {
		return n.stateFPOwner
	}
	return n.StateFP
}

func (n *GraphNode) ownedTIndex() int {
	if n.ownersCaptured {
		return n.tIndexOwner
	}
	return n.TIndex
}

func (n *GraphNode) ownedNodes() []int {
	if n.ownersCaptured {
		return n.nodesOwner
	}
	return n.Nodes
}

func (n *GraphNode) setNodes(nodes []int) {
	n.Nodes = nodes
	if n.ownersCaptured {
		n.nodesOwner = nodes
	}
}

func graphNodeArrayGet(nodes []int, index int) int {
	if nodes == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(nodes) {
		panic(NewArrayIndexOutOfBoundsException(index, len(nodes)))
	}
	return nodes[index]
}

func graphNodeArraySet(nodes []int, index int, value int) {
	if nodes == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(nodes) {
		panic(NewArrayIndexOutOfBoundsException(index, len(nodes)))
	}
	nodes[index] = value
}

type GraphTransition struct {
	FP            uint64
	TIndex        int
	checksPresent bool
	checksHash    int
	checksText    string
}

func NewGraphTransition(fp uint64, tidx int, checks *BitVector) GraphTransition {
	tidx = int(int32(tidx))
	text := ""
	hash := 0
	if checks != nil {
		text = checks.String()
		hash = checks.Hash()
	}
	return GraphTransition{FP: fp, TIndex: tidx, checksPresent: checks != nil, checksHash: hash, checksText: text}
}

func (t GraphTransition) HashCode() int32 {
	result := int32(1)
	if t.checksPresent {
		result = 31*result + int32(t.checksHash)
	} else {
		result *= 31
	}
	result = 31*result + int32(t.FP^(t.FP>>32))
	result = 31*result + int32(t.TIndex)
	return result
}

func (t GraphTransition) Equals(other any) bool {
	var transition GraphTransition
	switch value := other.(type) {
	case GraphTransition:
		transition = value
	case *GraphTransition:
		if value == nil {
			return false
		}
		transition = *value
	default:
		return false
	}
	return t.FP == transition.FP && t.TIndex == transition.TIndex && t.checksPresent == transition.checksPresent && t.checksText == transition.checksText
}

func (t GraphTransition) GetChecks() *BitVector {
	if !t.checksPresent {
		return nil
	}
	checks := NewBitVector(0)
	if len(t.checksText) < 2 || t.checksText == "[]" {
		return checks
	}
	bits := t.checksText[1 : len(t.checksText)-1]
	for i := range bits {
		if bits[len(bits)-1-i] == '1' {
			checks.Set(i)
		}
	}
	return checks
}

func (t GraphTransition) GetFP() uint64 { return t.FP }

func (t GraphTransition) GetTidx() int { return t.TIndex }

type BEGraphNode struct {
	AbstractGraphNode
	StateFP    uint64
	Nexts      []*BEGraphNode
	Number     int64
	Tableau    bool
	TableauIdx int32

	beOwnersCaptured bool
	nextsOwner       []*BEGraphNode
	numberOwner      int64
	tableauOwner     bool
	tableauIdxOwner  int32
}

const beGraphVisitedMask = int64(-1 << 63)

func NewBEGraphNode(fp uint64) *BEGraphNode {
	nexts := make([]*BEGraphNode, 0)
	return &BEGraphNode{
		AbstractGraphNode: NewAbstractGraphNode(NewBitVector(0)), StateFP: fp, Nexts: nexts,
		beOwnersCaptured: true, nextsOwner: nexts,
	}
}

func (n *BEGraphNode) HashCode() int32 {
	if n == nil {
		panic(NewNullPointerException())
	}
	result := int32(31) + int32(n.StateFP^(n.StateFP>>32))
	if n.ownedTableau() {
		result = 31*result + int32(n.GetIndex())
	}
	return result
}

func (n *BEGraphNode) Equals(other any) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	var node *BEGraphNode
	switch value := other.(type) {
	case *BEGraphNode:
		node = value
	case *BTGraphNode:
		if value != nil {
			node = value.BEGraphNode
		}
	default:
		return false
	}
	return beGraphNodesEqual(n, node)
}

func (n *BEGraphNode) requireBEGraphNode() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *BEGraphNode) NextAt(i int) *BEGraphNode {
	n.requireBEGraphNode()
	i = int(int32(i))
	nexts := n.ownedNexts()
	if i < 0 || i >= len(nexts) {
		panic(NewArrayIndexOutOfBoundsException(i, len(nexts)))
	}
	return nexts[i]
}

func (n *BEGraphNode) NextSize() int {
	n.requireBEGraphNode()
	return len(n.ownedNexts())
}

func (n *BEGraphNode) ResetNumberField() int64 {
	n.requireBEGraphNode()
	old := n.ownedNumber()
	n.setNumberWord(0)
	return old
}

func (n *BEGraphNode) GetNumber() int64 {
	n.requireBEGraphNode()
	return n.ownedNumber() & 0x7fffffffffffffff
}

func (n *BEGraphNode) IncNumber() {
	n.requireBEGraphNode()
	n.setNumberWord(n.ownedNumber() + 1)
}

func (n *BEGraphNode) SetNumber(num int64) {
	n.requireBEGraphNode()
	if n.ownedNumber() < 0 {
		n.setNumberWord(num | beGraphVisitedMask)
	} else {
		n.setNumberWord(num)
	}
}

type BTGraphNode struct {
	*BEGraphNode
	TIndex int32

	btOwnersCaptured bool
	tIndexOwner      int32
}

func NewBTGraphNode(fp uint64, index int) *BTGraphNode {
	node := NewBEGraphNode(fp)
	node.setTableauIdentity(true, int32(index))
	return newBTGraphNodeView(node)
}

func newBTGraphNodeView(node *BEGraphNode) *BTGraphNode {
	index := node.ownedTableauIdx()
	return &BTGraphNode{BEGraphNode: node, TIndex: index, btOwnersCaptured: true, tIndexOwner: index}
}

func (n *BTGraphNode) HashCode() int32 {
	if n == nil || n.BEGraphNode == nil {
		panic(NewNullPointerException())
	}
	result := int32(31) + int32(n.StateFP^(n.StateFP>>32))
	return 31*result + int32(n.GetIndex())
}

func (n *BTGraphNode) Equals(other any) bool {
	if n == nil || n.BEGraphNode == nil {
		panic(NewNullPointerException())
	}
	var node *BEGraphNode
	switch value := other.(type) {
	case *BTGraphNode:
		if value != nil {
			node = value.BEGraphNode
		}
	case *BEGraphNode:
		node = value
	default:
		return false
	}
	return node != nil && node.ownedTableau() && n.StateFP == node.StateFP && n.GetIndex() == node.GetIndex()
}

func (n *BEGraphNode) IsTableauNode() bool {
	return n != nil && n.ownedTableau()
}

func (n *BEGraphNode) GetIndex() int {
	if n == nil {
		return 0
	}
	return int(uint32(n.ownedTableauIdx()) & 0x3fffffff)
}

func (n *BEGraphNode) GetTNode(tableau *TBGraph) *TBGraphNode {
	n.requireBEGraphNode()
	if !n.ownedTableau() {
		panic(NewWrongInvocationException("TLC bug: should never call BEGraphNode.getTNode()."))
	}
	return tableau.GetNode(n.GetIndex())
}

func (n *BTGraphNode) GetIndex() int {
	if n == nil {
		panic(NewNullPointerException())
	}
	return int(uint32(n.ownedTIndex()) & 0x3fffffff)
}

func (n *BTGraphNode) SetIndex(index int) {
	if n == nil {
		panic(NewNullPointerException())
	}
	n.setTIndex(int32((uint32(n.ownedTIndex()) & 0xc0000000) | uint32(index)))
}

func (n *BTGraphNode) IsDone() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	return n.ownedTIndex() < 0
}

func (n *BTGraphNode) SetDone() {
	if n == nil {
		panic(NewNullPointerException())
	}
	n.setTIndex(int32(uint32(n.ownedTIndex()) | 0x80000000))
}

func NewDummyBTGraphNode(fp uint64) *BTGraphNode {
	return NewBTGraphNode(fp, 0x40000000)
}

func (n *BTGraphNode) IsDummy() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	return (uint32(n.ownedTIndex()) & 0x40000000) != 0
}

func (n *BTGraphNode) GetTNode(tableau *TBGraph) *TBGraphNode {
	if n == nil || tableau == nil {
		panic(NewNullPointerException())
	}
	return tableau.GetNode(n.GetIndex())
}

func (n *BTGraphNode) NodeInfo() string {
	if n == nil || n.BEGraphNode == nil {
		panic(NewNullPointerException())
	}
	return fmt.Sprintf("<%s,%d>", livenessFPString(n.StateFP), n.GetIndex())
}

type BEGraph struct {
	InitNodes []*BEGraphNode
	Metadir   string
	AllNodes  *NodeTable
}

func NewBEGraph(metadir string, isBT bool) *BEGraph {
	return &BEGraph{
		InitNodes: make([]*BEGraphNode, 0, 10),
		Metadir:   metadir,
		AllNodes:  NewNodeTable(127, isBT),
	}
}

func (g *BEGraph) ResetNumberField() {
	if g == nil {
		panic(NewNullPointerException())
	}
	stack := NewMemObjectStack(g.Metadir, "resetstack")
	for _, node := range g.InitNodes {
		if node.ResetNumberField() != 0 {
			stack.Push(node)
		}
	}
	for stack.Size() != 0 {
		node, _ := stack.Pop().(*BEGraphNode)
		for _, next := range node.ownedNexts() {
			if next.ResetNumberField() != 0 {
				stack.Push(next)
			}
		}
	}
}

func (g *BEGraph) GetInitNode(i int) *BEGraphNode {
	if g == nil {
		panic(NewNullPointerException())
	}
	i = int(int32(i))
	if i < 0 || i >= cap(g.InitNodes) {
		panic(NewArrayIndexOutOfBoundsException(i, cap(g.InitNodes)))
	}
	return g.InitNodes[:cap(g.InitNodes)][i]
}

func (g *BEGraph) AddInitNode(node *BEGraphNode) {
	if g == nil {
		panic(NewNullPointerException())
	}
	if len(g.InitNodes) == cap(g.InitNodes) {
		newCapacity := cap(g.InitNodes) * 2
		if newCapacity < len(g.InitNodes)+1 {
			newCapacity = len(g.InitNodes) + 1
		}
		nodes := make([]*BEGraphNode, len(g.InitNodes), newCapacity)
		copy(nodes, g.InitNodes)
		g.InitNodes = nodes
	}
	g.InitNodes = append(g.InitNodes, node)
}

func (g *BEGraph) InitSize() int {
	if g == nil {
		panic(NewNullPointerException())
	}
	return len(g.InitNodes)
}

func BEGraphGetPath(start *BEGraphNode, end *BEGraphNode) ([]*BEGraphNode, error) {
	if start.Equals(end) {
		start.SetParent(nil)
	} else {
		unseen := start.GetVisited()
		queue := NewMemObjectQueue("")
		start.FlipVisited()
		queue.Enqueue(beGraphNodeAndParent{node: start})
		found := false
		for !found {
			item, _ := queue.Dequeue().(beGraphNodeAndParent)
			if item.node == nil {
				return nil, newTLCErrorCode(ECTLCLiveBEGraphFailedToConstruct)
			}
			curNode := item.node
			for _, nextNode := range curNode.ownedNexts() {
				if nextNode.GetVisited() != unseen {
					continue
				}
				if nextNode.Equals(end) {
					end.SetParent(curNode)
					found = true
					break
				}
				nextNode.FlipVisited()
				queue.Enqueue(beGraphNodeAndParent{node: nextNode, parent: curNode})
			}
			curNode.SetParent(item.parent)
		}
	}
	var rev []*BEGraphNode
	for cur := end; cur != nil; cur = cur.GetParent() {
		rev = append(rev, cur)
	}
	path := make([]*BEGraphNode, len(rev))
	for i := range rev {
		path[i] = rev[len(rev)-i-1]
	}
	return path, nil
}

func (g *BEGraph) String() string {
	if g == nil {
		panic(NewNullPointerException())
	}
	if len(g.InitNodes) == 0 {
		return ""
	}
	var b strings.Builder
	unseen := g.InitNodes[0].GetVisited()
	for _, node := range g.InitNodes {
		node.writeString(&b, unseen)
	}
	return b.String()
}

type beGraphNodeAndParent struct {
	node   *BEGraphNode
	parent *BEGraphNode
}

func (n *BEGraphNode) GetVisited() bool {
	n.requireBEGraphNode()
	return n.ownedNumber() < 0
}

func (n *BEGraphNode) FlipVisited() {
	n.requireBEGraphNode()
	n.setNumberWord(n.ownedNumber() ^ beGraphVisitedMask)
}

func (n *BEGraphNode) AddTransition(target *BEGraphNode, slen int, alen int, acts []bool) {
	n.requireBEGraphNode()
	nexts := n.ownedNexts()
	num := len(nexts)
	if acts != nil {
		pos := graphIntAdd(slen, graphIntMul(alen, num))
		for i, value := range acts {
			if value {
				n.ownedChecks().Set(graphIntAdd(pos, i))
			}
		}
	}
	n.setNexts(append(nexts, target))
}

func (n *BEGraphNode) TransExists(target *BEGraphNode) bool {
	n.requireBEGraphNode()
	for _, next := range n.ownedNexts() {
		if target == nil {
			panic(NewNullPointerException())
		}
		if target.Equals(next) {
			return true
		}
	}
	return false
}

func beGraphNodesEqual(left *BEGraphNode, right *BEGraphNode) bool {
	if left == right {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	if left.StateFP != right.StateFP {
		return false
	}
	if left.ownedTableau() != right.ownedTableau() {
		return false
	}
	if left.ownedTableau() {
		return left.GetIndex() == right.GetIndex()
	}
	return true
}

func (n *BEGraphNode) NodeInfo() string {
	n.requireBEGraphNode()
	if n.ownedTableau() {
		return fmt.Sprintf("<%s,%d>", livenessFPString(n.StateFP), n.GetIndex())
	}
	return livenessFPString(n.StateFP)
}

func (n *BEGraphNode) SetParent(parent *BEGraphNode) {
	n.requireBEGraphNode()
	nexts := n.ownedNexts()
	if len(nexts) == 0 {
		nexts = make([]*BEGraphNode, 1)
	}
	nexts[0] = parent
	n.setNexts(nexts)
}

func (n *BEGraphNode) GetParent() *BEGraphNode {
	n.requireBEGraphNode()
	nexts := n.ownedNexts()
	if len(nexts) == 0 {
		panic(NewArrayIndexOutOfBoundsException(0, 0))
	}
	return nexts[0]
}

func (n *BEGraphNode) String() string {
	n.requireBEGraphNode()
	var b strings.Builder
	n.writeString(&b, n.GetVisited())
	return b.String()
}

func (n *BEGraphNode) writeString(b *strings.Builder, unseen bool) {
	n.requireBEGraphNode()
	if n.GetVisited() != unseen {
		return
	}
	n.FlipVisited()
	if n.ownedTableau() {
		b.WriteString(fmt.Sprintf("(%s,%d) --> ", livenessFPString(n.StateFP), n.GetIndex()))
	} else {
		b.WriteString(fmt.Sprintf("%s --> ", livenessFPString(n.StateFP)))
	}
	nexts := n.ownedNexts()
	if len(nexts) != 0 {
		if n.ownedTableau() {
			nexts[0].writeTableauNodeInfo(b)
		} else {
			nexts[0].writeNodeInfo(b)
		}
	}
	for i := 1; i < len(nexts); i++ {
		b.WriteString(", ")
		if n.ownedTableau() {
			nexts[i].writeTableauNodeInfo(b)
		} else {
			nexts[i].writeNodeInfo(b)
		}
	}
	b.WriteByte('\n')
	for _, next := range nexts {
		next.writeString(b, unseen)
	}
}

func (n *BEGraphNode) ownedNexts() []*BEGraphNode {
	if n.beOwnersCaptured {
		return n.nextsOwner
	}
	return n.Nexts
}

func (n *BEGraphNode) setNexts(nexts []*BEGraphNode) {
	n.Nexts = nexts
	if n.beOwnersCaptured {
		n.nextsOwner = nexts
	}
}

func (n *BEGraphNode) ownedNumber() int64 {
	if n.beOwnersCaptured {
		return n.numberOwner
	}
	return n.Number
}

func (n *BEGraphNode) setNumberWord(number int64) {
	n.Number = number
	if n.beOwnersCaptured {
		n.numberOwner = number
	}
}

func (n *BEGraphNode) ownedTableau() bool {
	if n.beOwnersCaptured {
		return n.tableauOwner
	}
	return n.Tableau
}

func (n *BEGraphNode) ownedTableauIdx() int32 {
	if n.beOwnersCaptured {
		return n.tableauIdxOwner
	}
	return n.TableauIdx
}

func (n *BEGraphNode) setTableauIdentity(tableau bool, index int32) {
	n.Tableau = tableau
	n.TableauIdx = index
	if n.beOwnersCaptured {
		n.tableauOwner = tableau
		n.tableauIdxOwner = index
	}
}

func (n *BTGraphNode) ownedTIndex() int32 {
	if n.btOwnersCaptured {
		return n.tIndexOwner
	}
	return n.TIndex
}

func (n *BTGraphNode) setTIndex(index int32) {
	n.TIndex = index
	if n.btOwnersCaptured {
		n.tIndexOwner = index
	}
	if n.BEGraphNode != nil {
		n.BEGraphNode.setTableauIdentity(true, index)
	}
}

func (n *BEGraphNode) writeNodeInfo(b *strings.Builder) {
	n.requireBEGraphNode()
	if n.ownedTableau() {
		b.WriteString(fmt.Sprintf("(%s,%d)", livenessFPString(n.StateFP), n.GetIndex()))
	} else {
		b.WriteString(livenessFPString(n.StateFP))
	}
}

func (n *BEGraphNode) writeTableauNodeInfo(b *strings.Builder) {
	n.requireBEGraphNode()
	if !n.ownedTableau() {
		panic(valueStreamClassCast(n, "tlc2.tool.liveness.BTGraphNode"))
	}
	b.WriteString(fmt.Sprintf("(%s,%d)", livenessFPString(n.StateFP), n.GetIndex()))
}
