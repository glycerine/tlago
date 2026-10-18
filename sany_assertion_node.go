// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

type sanySemAssumeNode struct {
	sanySemanticNode
	assumeLevelChecked int32 // Java AssumeNode shadows LevelNode.levelChecked.
	module             *sanySemModuleNode
	assumeExpr         sanySemanticGraphNode
	def                *sanySemThmOrAssumpDefNode
	isAxiom            bool
}

type sanySemTheoremNode struct {
	sanySemanticNode
	theoremLevelChecked      int32 // Java TheoremNode shadows LevelNode.levelChecked.
	module                   *sanySemModuleNode
	theoremExprOrAssumeProve sanySemanticGraphNode
	def                      *sanySemThmOrAssumpDefNode
	suffices                 bool
	proof                    sanySemanticGraphNode
}

func sanyAssertionSyntax(node *sanySemanticNode, syntax *SanySyntaxNode) {
	if syntax == nil {
		node.TreeNode, node.Location = nil, tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	}
}

func newSanySemAssumeNode(syntax *SanySyntaxNode, expr sanySemanticGraphNode, module *sanySemModuleNode, definition *sanySemThmOrAssumpDefNode) *sanySemAssumeNode {
	node := &sanySemAssumeNode{sanySemanticNode: newSanySemanticNode(sanyAssumeKind), module: module, assumeExpr: expr, def: definition}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	if syntax == nil {
		panic(tlc.NewNullPointerException(""))
	}
	heirs := syntax.GetHeirs()
	if len(heirs) == 0 {
		panic(tlc.NewArrayIndexOutOfBoundsException(0, 0))
	}
	if heirs[0] == nil {
		panic(tlc.NewNullPointerException(""))
	}
	node.isAxiom = heirs[0].Image == "AXIOM"
	if definition != nil {
		definition.thmOrAssump = node
	}
	return node
}
func (node *sanySemAssumeNode) getIsAxiom() bool                   { return node.isAxiom }
func (node *sanySemAssumeNode) getAssume() sanySemanticGraphNode   { return node.assumeExpr }
func (node *sanySemAssumeNode) getDef() *sanySemThmOrAssumpDefNode { return node.def }
func (node *sanySemAssumeNode) getChildren() []sanySemanticGraphNode {
	return []sanySemanticGraphNode{node.assumeExpr}
}

func newSanySemTheoremNode(syntax *SanySyntaxNode, theorem sanySemanticGraphNode, module *sanySemModuleNode, proof sanySemanticGraphNode, definition *sanySemThmOrAssumpDefNode) *sanySemTheoremNode {
	node := &sanySemTheoremNode{sanySemanticNode: newSanySemanticNode(sanyTheoremKind), module: module, theoremExprOrAssumeProve: theorem, proof: proof, def: definition}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	if definition != nil {
		definition.thmOrAssump = node
		if definition.getBody() != theorem {
			panic(tlc.NewAssertionError())
		}
	}
	return node
}
func (node *sanySemTheoremNode) getTheorem() sanySemanticGraphNode {
	return node.theoremExprOrAssumeProve
}
func (node *sanySemTheoremNode) getDef() *sanySemThmOrAssumpDefNode { return node.def }
func (node *sanySemTheoremNode) isSuffices() bool                   { return node.suffices }
func (node *sanySemTheoremNode) getProof() sanySemanticGraphNode    { return node.proof }

// The bool distinguishes a null Java UniqueString from an empty name.
func (node *sanySemTheoremNode) getName() (string, bool) {
	if node.def == nil {
		return "", false
	}
	return node.def.semName(), true
}
func (node *sanySemTheoremNode) getChildren() []sanySemanticGraphNode {
	if node.proof == nil {
		return []sanySemanticGraphNode{node.theoremExprOrAssumeProve}
	}
	return []sanySemanticGraphNode{node.theoremExprOrAssumeProve, node.proof}
}

func (module *sanySemModuleNode) addAssumption(syntax *SanySyntaxNode, body sanySemanticGraphNode, table *sanySymbolTable, definition *sanySemThmOrAssumpDefNode) {
	node := newSanySemAssumeNode(syntax, body, module, definition)
	module.assumptionVec = append(module.assumptionVec, node)
	module.topLevelVec = append(module.topLevelVec, node)
}
func (module *sanySemModuleNode) addTheorem(syntax *SanySyntaxNode, body sanySemanticGraphNode, proof sanySemanticGraphNode, definition *sanySemThmOrAssumpDefNode) {
	node := newSanySemTheoremNode(syntax, body, module, proof, definition)
	module.theoremVec = append(module.theoremVec, node)
	module.topLevelVec = append(module.topLevelVec, node)
}
func (module *sanySemModuleNode) addTopLevel(node sanySemanticGraphNode) {
	module.topLevelVec = append(module.topLevelVec, node)
}
func (module *sanySemModuleNode) getAssumptions() []*sanySemAssumeNode {
	if module.assumptions == nil {
		module.assumptions = make([]*sanySemAssumeNode, len(module.assumptionVec))
		copy(module.assumptions, module.assumptionVec)
	}
	return module.assumptions
}
func (module *sanySemModuleNode) getTheorems() []*sanySemTheoremNode {
	if module.theorems == nil {
		module.theorems = make([]*sanySemTheoremNode, len(module.theoremVec))
		copy(module.theorems, module.theoremVec)
	}
	return module.theorems
}
func (module *sanySemModuleNode) getTopLevel() []sanySemanticGraphNode {
	if module.topLevel == nil {
		module.topLevel = make([]sanySemanticGraphNode, len(module.topLevelVec))
		copy(module.topLevel, module.topLevelVec)
	}
	return module.topLevel
}
