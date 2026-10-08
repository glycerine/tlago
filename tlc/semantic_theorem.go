// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
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

func (n *TheoremNode) GetTheorem() SemanticNode    { return n.Theorem }
func (n *TheoremNode) GetDef() *ThmOrAssumpDefNode { return n.Def }
func (n *TheoremNode) GetProof() SemanticNode      { return n.Proof }
func (n *TheoremNode) IsSuffices() bool            { return n.Suffices }
func (n *TheoremNode) GetName() *UniqueString {
	if n.Def == nil {
		return nil
	}
	return n.Def.Name
}
