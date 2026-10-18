package tlc

// These views retain SANY's statement and proof graph. SANY owns their level
// checking; TLC traverses their source edges without treating goals as bodies.
type AssumeProveNode struct {
	*SemanticNodeBase
	Assumes                             []SemanticNode
	Prove                               SemanticNode
	Goal                                *ThmOrAssumpDefNode
	InScopeOfDecl                       []bool
	InProof, Suffices, IsBoxAssumeProve bool
}

func (n *AssumeProveNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *AssumeProveNode) GetSuffices() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeAssumeProveFlags() (bool, bool, bool)
	}); ok {
		if suffices, _, present := provider.RuntimeAssumeProveFlags(); present {
			return suffices
		}
	}
	return n.Suffices
}

func (n *AssumeProveNode) GetIsBoxAssumeProve() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeAssumeProveFlags() (bool, bool, bool)
	}); ok {
		if _, boxed, present := provider.RuntimeAssumeProveFlags(); present {
			return boxed
		}
	}
	return n.IsBoxAssumeProve
}

func (n *AssumeProveNode) GetAssumes() []SemanticNode {
	n.require()
	return n.Assumes
}

func (n *AssumeProveNode) GetProve() SemanticNode {
	n.require()
	return n.Prove
}

func (n *AssumeProveNode) GetGoal() *ThmOrAssumpDefNode {
	n.require()
	return n.Goal
}

func (n *AssumeProveNode) GetChildren() []SemanticNode {
	n.require()
	if n.Assumes == nil {
		panic(NewNullPointerException())
	}
	children := append([]SemanticNode{}, n.Assumes...)
	return append(children, n.Prove)
}

type NewSymbNode struct {
	*SemanticNodeBase
	OpDecl *SymbolNode
	Set    SemanticNode
}

func (n *NewSymbNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *NewSymbNode) GetSet() SemanticNode {
	n.require()
	return n.Set
}

func (n *NewSymbNode) GetOpDeclNode() *SymbolNode {
	n.require()
	return n.OpDecl
}

func (n *NewSymbNode) GetChildren() []SemanticNode {
	n.require()
	if n.Set == nil {
		return nil
	}
	return []SemanticNode{n.Set}
}

type LeafProofNode struct {
	*SemanticNodeBase
	Facts         []SemanticNode
	Defs          []*SymbolNode
	Omitted, Only bool
}

func (n *LeafProofNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *LeafProofNode) GetFacts() []SemanticNode {
	n.require()
	return n.Facts
}

func (n *LeafProofNode) GetDefs() []*SymbolNode {
	n.require()
	return n.Defs
}

func (n *LeafProofNode) GetOmitted() bool {
	n.require()
	return n.Omitted
}

func (n *LeafProofNode) GetOnlyFlag() bool {
	n.require()
	return n.Only
}

func proofFactChildren(facts []SemanticNode) []SemanticNode {
	if len(facts) == 0 {
		return nil
	}
	return append([]SemanticNode(nil), facts...)
}

func (n *LeafProofNode) GetChildren() []SemanticNode {
	n.require()
	return proofFactChildren(n.Facts)
}

type NonLeafProofNode struct {
	*SemanticNodeBase
	Steps, Instances []SemanticNode
	Context          *SemanticContext
}

func (n *NonLeafProofNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *NonLeafProofNode) GetSteps() []SemanticNode {
	n.require()
	return n.Steps
}

func (n *NonLeafProofNode) GetContext() *SemanticContext {
	n.require()
	return n.Context
}

func (n *NonLeafProofNode) GetChildren() []SemanticNode {
	n.require()
	return proofFactChildren(n.Steps)
}

type DefStepNode struct {
	*SemanticNodeBase
	StepNumber *UniqueString
	Defs       []*OpDefNode
}

func (n *DefStepNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *DefStepNode) GetStepNumber() *UniqueString {
	n.require()
	return n.StepNumber
}

func (n *DefStepNode) GetDefs() []*OpDefNode {
	n.require()
	return n.Defs
}

func (n *DefStepNode) GetChildren() []SemanticNode {
	n.require()
	if n.Defs == nil {
		panic(NewNullPointerException())
	}
	children := make([]SemanticNode, len(n.Defs))
	for i, definition := range n.Defs {
		children[i] = definition
	}
	return children
}

type UseOrHideNode struct {
	*SemanticNodeBase
	Facts    []SemanticNode
	Defs     []*SymbolNode
	Only     bool
	StepName *UniqueString
}

func (n *UseOrHideNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *UseOrHideNode) SetStepName(name *UniqueString) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSetUseOrHideStepName(*UniqueString) bool
	}); ok {
		provider.RuntimeSetUseOrHideStepName(name)
	}
	n.StepName = name
}

func (n *UseOrHideNode) GetStepName() *UniqueString {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeUseOrHideStepName() (*UniqueString, bool)
	}); ok {
		if stepName, present := provider.RuntimeUseOrHideStepName(); present {
			return stepName
		}
	}
	return n.StepName
}

func (n *UseOrHideNode) GetChildren() []SemanticNode {
	n.require()
	return proofFactChildren(n.Facts)
}

type InstanceNode struct {
	*SemanticNodeBase
	Name, StepName *UniqueString
	Params         []*SymbolNode
	Module         *ModuleNode
	Substs         []Subst
	Local          bool
}

func (n *InstanceNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *InstanceNode) GetName() *UniqueString {
	n.require()
	return n.Name
}

func (n *InstanceNode) GetModule() *ModuleNode {
	n.require()
	return n.Module
}

func (n *InstanceNode) GetLocal() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeInstanceState() (bool, *UniqueString, bool)
	}); ok {
		if local, _, present := provider.RuntimeInstanceState(); present {
			return local
		}
	}
	return n.Local
}

func (n *InstanceNode) SetStepName(name *UniqueString) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSetInstanceStepName(*UniqueString) bool
	}); ok {
		provider.RuntimeSetInstanceStepName(name)
	}
	n.StepName = name
}

func (n *InstanceNode) GetStepName() *UniqueString {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeInstanceState() (bool, *UniqueString, bool)
	}); ok {
		if _, stepName, present := provider.RuntimeInstanceState(); present {
			return stepName
		}
	}
	return n.StepName
}

func (n *InstanceNode) IsLocal() bool {
	return n.GetLocal()
}

// GetLevel preserves InstanceNode's override: an INSTANCE only imports
// definitions and therefore reports constant level regardless of inherited
// level storage.
func (n *InstanceNode) GetLevel() int {
	n.require()
	return TLCLevelConstant
}

func (n *InstanceNode) GetChildren() []SemanticNode {
	n.require()
	if n.Substs == nil {
		panic(NewNullPointerException())
	}
	children := make([]SemanticNode, len(n.Substs))
	for i := range n.Substs {
		children[i] = n.Substs[i].GetExpr()
	}
	return children
}

// WalkGaph preserves the misspelled public method on Java InstanceNode. The
// correctly spelled walkGraph remains SemanticNode's leaf traversal because
// the source method does not override it.
func (n *InstanceNode) WalkGaph(table map[int32]any, visitor *SemanticExplorerVisitor) {
	n.require()
	uid := n.GetUid()
	if table == nil {
		panic(NewNullPointerException())
	}
	if !semanticExploreNull(table[uid]) {
		return
	}
	table[uid] = n
	visitor.pre(n)
	for i := 0; i < semanticGraphArrayLength(true, n.Params); i++ {
		WalkSemanticGraph(n.Params[i], table, visitor)
	}
	WalkSemanticGraph(n.Module, table, visitor)
	visitor.post(n)
}
