package tlc

// TheoremNode is the TLC view of a retained SANY statement. Its semantic base
// belongs to the actual parser node; expression adapters remain shared with
// the named definition. SANY itself performs proof and level checking.
type TheoremNode struct {
	*SemanticNodeBase
	Theorem  SemanticNode
	Module   *ModuleNode
	Def      *ThmOrAssumpDefNode
	Proof    SemanticNode
	Suffices bool
}

func (n *TheoremNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *TheoremNode) GetTheorem() SemanticNode {
	n.require()
	return n.Theorem
}
func (n *TheoremNode) GetDef() *ThmOrAssumpDefNode {
	n.require()
	return n.Def
}
func (n *TheoremNode) GetProof() SemanticNode {
	n.require()
	return n.Proof
}
func (n *TheoremNode) IsSuffices() bool {
	n.require()
	return n.Suffices
}
func (n *TheoremNode) GetName() *UniqueString {
	n.require()
	if n.Def == nil {
		return nil
	}
	return n.Def.Name
}

func (n *TheoremNode) GetChildren() []SemanticNode {
	n.require()
	if n.Proof == nil {
		return []SemanticNode{n.Theorem}
	}
	return []SemanticNode{n.Theorem, n.Proof}
}
