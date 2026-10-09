// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlc

// These views retain SANY's statement and proof graph. SANY owns their level
// checking; TLC traverses their source edges without treating goals as bodies.
type AssumeProveNode struct {
	*SemanticNodeBase
	Assumes                             []SemanticNode
	Prove, Goal                         SemanticNode
	InScopeOfDecl                       []bool
	InProof, Suffices, IsBoxAssumeProve bool
}

type NewSymbNode struct {
	*SemanticNodeBase
	OpDecl *SymbolNode
	Set    SemanticNode
}

type LeafProofNode struct {
	*SemanticNodeBase
	Facts         []SemanticNode
	Defs          []*SymbolNode
	Omitted, Only bool
}

type NonLeafProofNode struct {
	*SemanticNodeBase
	Steps, Instances []SemanticNode
	Context          *SemanticContext
}

type DefStepNode struct {
	*SemanticNodeBase
	StepNumber *UniqueString
	Defs       []*OpDefNode
}

type UseOrHideNode struct {
	*SemanticNodeBase
	Facts    []SemanticNode
	Defs     []*SymbolNode
	Only     bool
	StepName *UniqueString
}

type InstanceNode struct {
	*SemanticNodeBase
	Name, StepName *UniqueString
	Params         []*SymbolNode
	Module         *ModuleNode
	Substs         []Subst
	Local          bool
}
