// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// Both source wrappers retain the same Subst objects, body and module identities.
// They remain distinct semantic kinds; APSubstIn may contain an ASSUME/PROVE.
type sanySemSubstitutionNode struct {
	sanySemanticNode
	substs              []*sanySemSubst
	body                sanySemanticGraphNode
	instantiatingModule *sanySemModuleNode
	instantiatedModule  *sanySemModuleNode
}
type sanySemSubstInNode struct{ sanySemSubstitutionNode }
type sanySemAPSubstInNode struct{ sanySemSubstitutionNode }

func newSanySemSubstitutionNode(kind sanySemKind, syntax *SanySyntaxNode, substs []*sanySemSubst, body sanySemanticGraphNode, instantiating, instantiated *sanySemModuleNode, checkBody bool) (sanySemSubstitutionNode, Diagnostics) {
	node := sanySemSubstitutionNode{sanySemanticNode: newSanySemanticNode(kind), substs: substs, body: body, instantiatingModule: instantiating, instantiatedModule: instantiated}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	if checkBody && body == nil {
		if syntax == nil {
			panic(tlc.NewNullPointerException(""))
		}
		if kind == sanyAPSubstInKind {
			return node, Diagnostics{sanyRegistrationDiagnostic(sanyNodePosition(syntax), "E4004", "Substitution error, probably due to error in module being instantiated.")}
		}
		return node, Diagnostics{sanyRegistrationDiagnostic(sanyNodePosition(syntax), "E4003", "Substitution error, probably due to error \nin module being instantiated.")}
	}
	return node, nil
}

// Copy constructors share the supplied template's array; they never copy its
// Subst entries or normalize a null array. Source template access precedes UID
// allocation, whereas a null-body diagnostic follows allocation.
func newSanySemSubstInNode(syntax *SanySyntaxNode, subst *sanySemSubstInNode, body sanySemanticGraphNode, instantiating, instantiated *sanySemModuleNode) (*sanySemSubstInNode, Diagnostics) {
	if subst == nil {
		panic(tlc.NewNullPointerException(""))
	}
	node, diagnostics := newSanySemSubstitutionNode(sanySubstInKind, syntax, subst.substs, body, instantiating, instantiated, true)
	return &sanySemSubstInNode{node}, diagnostics
}
func newSanySemSubstInNodeFromSubst(subst *sanySemSubstInNode, body sanySemanticGraphNode) (*sanySemSubstInNode, Diagnostics) {
	if subst == nil {
		panic(tlc.NewNullPointerException(""))
	}
	var syntax *SanySyntaxNode
	if subst.TreeNode != nil {
		syntax = subst.TreeNode.(*SanySyntaxNode)
	}
	return newSanySemSubstInNode(syntax, subst, body, subst.instantiatingModule, subst.instantiatedModule)
}
func newSanySemSubstInNodeFromAP(subst *sanySemAPSubstInNode, body sanySemanticGraphNode) (*sanySemSubstInNode, Diagnostics) {
	if subst == nil {
		panic(tlc.NewNullPointerException(""))
	}
	var syntax *SanySyntaxNode
	if subst.TreeNode != nil {
		syntax = subst.TreeNode.(*SanySyntaxNode)
	}
	node, diagnostics := newSanySemSubstitutionNode(sanySubstInKind, syntax, subst.substs, body, subst.instantiatingModule, subst.instantiatedModule, true)
	return &sanySemSubstInNode{node}, diagnostics
}
func newSanySemAPSubstInNode(syntax *SanySyntaxNode, subst *sanySemSubstInNode, body sanySemanticGraphNode, instantiating, instantiated *sanySemModuleNode) (*sanySemAPSubstInNode, Diagnostics) {
	if subst == nil {
		panic(tlc.NewNullPointerException(""))
	}
	node, diagnostics := newSanySemSubstitutionNode(sanyAPSubstInKind, syntax, subst.substs, body, instantiating, instantiated, true)
	return &sanySemAPSubstInNode{node}, diagnostics
}

// The special default constructor allocates the wrapper before expressions.
// Missing names create no slot; later WITH processing supplies them explicitly.
func newSanySemDefaultSubstitutionNode(kind sanySemKind, syntax *SanySyntaxNode, table *sanySymbolTable, declarations []*sanySemOpDeclNode, instantiating, instantiated *sanySemModuleNode) (sanySemSubstitutionNode, Diagnostics, error) {
	node, _ := newSanySemSubstitutionNode(kind, syntax, nil, nil, instantiating, instantiated, false)
	if declarations == nil {
		panic(tlc.NewNullPointerException(""))
	}
	var diagnostics Diagnostics
	substs := make([]*sanySemSubst, 0, len(declarations))
	for _, declaration := range declarations {
		if declaration == nil || table == nil {
			panic(tlc.NewNullPointerException(""))
		}
		symbol := table.resolveSymbol(declaration.semName())
		if symbol == nil {
			continue
		}
		var expr sanySemanticGraphNode
		if declaration.semKind() == sanyVariableDeclKind || (declaration.semKind() == sanyConstantDeclKind && declaration.semArity() == 0) {
			application, generated, err := newSanySemOpApplNode(symbol, make([]sanySemanticGraphNode, 0), syntax)
			diagnostics = append(diagnostics, generated...)
			if err != nil {
				return node, diagnostics, err
			}
			expr = application
		} else {
			expr = newSanySemOpArgNode(symbol, syntax, instantiating)
		}
		substs = append(substs, newSanySemSubst(declaration, expr, nil, true))
	}
	node.substs = substs
	return node, diagnostics, nil
}
func newSanySemDefaultSubstInNode(syntax *SanySyntaxNode, table *sanySymbolTable, declarations []*sanySemOpDeclNode, instantiating, instantiated *sanySemModuleNode) (*sanySemSubstInNode, Diagnostics, error) {
	node, diagnostics, err := newSanySemDefaultSubstitutionNode(sanySubstInKind, syntax, table, declarations, instantiating, instantiated)
	if err != nil {
		return nil, diagnostics, err
	}
	return &sanySemSubstInNode{node}, diagnostics, nil
}
func newSanySemDefaultAPSubstInNode(syntax *SanySyntaxNode, table *sanySymbolTable, declarations []*sanySemOpDeclNode, instantiating, instantiated *sanySemModuleNode) (*sanySemAPSubstInNode, Diagnostics, error) {
	node, diagnostics, err := newSanySemDefaultSubstitutionNode(sanyAPSubstInKind, syntax, table, declarations, instantiating, instantiated)
	if err != nil {
		return nil, diagnostics, err
	}
	return &sanySemAPSubstInNode{node}, diagnostics, nil
}
func (node *sanySemSubstitutionNode) getSubsts() []*sanySemSubst         { return node.substs }
func (node *sanySemSubstitutionNode) getBody() sanySemanticGraphNode     { return node.body }
func (node *sanySemSubstitutionNode) setBody(body sanySemanticGraphNode) { node.body = body }
func (node *sanySemSubstitutionNode) getInstantiatingModule() *sanySemModuleNode {
	return node.instantiatingModule
}
func (node *sanySemSubstitutionNode) getInstantiatedModule() *sanySemModuleNode {
	return node.instantiatedModule
}
func (node *sanySemSubstitutionNode) substitutionAt(index int) *sanySemSubst {
	if node.substs == nil {
		panic(tlc.NewNullPointerException(""))
	}
	if index < 0 || index >= len(node.substs) {
		panic(tlc.NewArrayIndexOutOfBoundsException(index, len(node.substs)))
	}
	return node.substs[index]
}
func (node *sanySemSubstitutionNode) getSubFor(index int) *sanySemOpDeclNode {
	return node.substitutionAt(index).getOp()
}
func (node *sanySemSubstitutionNode) getSubWith(index int) sanySemanticGraphNode {
	return node.substitutionAt(index).getExpr()
}
func (node *sanySemSubstitutionNode) getChildren() []sanySemanticGraphNode {
	if node.substs == nil {
		panic(tlc.NewNullPointerException(""))
	}
	children := make([]sanySemanticGraphNode, len(node.substs)+1)
	children[0] = node.body
	for i, subst := range node.substs {
		children[i+1] = subst.getExpr()
	}
	return children
}

// Replacing an implicit slot mutates its existing Subst, so all wrapper copies
// observe the change. Appending instead installs a new array only on this node.
func (node *sanySemSubstitutionNode) addExplicitSubstitute(context *sanyContext, name *tlc.UniqueString, syntax *SanySyntaxNode, expr sanySemanticGraphNode) Diagnostics {
	if node.substs == nil {
		panic(tlc.NewNullPointerException(""))
	}
	for _, subst := range node.substs {
		op := subst.getOp()
		if op == nil {
			panic(tlc.NewNullPointerException(""))
		}
		if tlc.UniqueStringOf(op.semName()) == name {
			if !subst.isImplicit() {
				if syntax == nil {
					panic(tlc.NewNullPointerException(""))
				}
				code := "E4241"
				if node.getKind() == sanyAPSubstInKind {
					code = "E4004"
				}
				return Diagnostics{sanyRegistrationDiagnostic(sanyNodePosition(syntax), code, "Multiple substitutions for symbol '%s' in substitution.", name.String())}
			}
			subst.setExpr(expr, false)
			subst.setExprSTN(syntax)
			return nil
		}
	}
	if context == nil || name == nil {
		panic(tlc.NewNullPointerException(""))
	}
	symbol := context.getSymbol(name.String())
	if declaration, ok := symbol.(*sanySemOpDeclNode); ok {
		substs := make([]*sanySemSubst, len(node.substs)+1)
		copy(substs, node.substs)
		substs[len(node.substs)] = newSanySemSubst(declaration, expr, syntax, false)
		node.substs = substs
	}
	return nil
}
func (node *sanySemSubstitutionNode) matchAll(declarations []*sanySemOpDeclNode) Diagnostics {
	if declarations == nil {
		panic(tlc.NewNullPointerException(""))
	}
	var diagnostics Diagnostics
	for _, declaration := range declarations {
		if declaration == nil || node.substs == nil {
			panic(tlc.NewNullPointerException(""))
		}
		found := false
		for _, subst := range node.substs {
			op := subst.getOp()
			if op == nil {
				panic(tlc.NewNullPointerException(""))
			}
			if op.semName() == declaration.semName() {
				found = true
				break
			}
		}
		if !found {
			if node.TreeNode == nil || declaration.TreeNode == nil || node.instantiatingModule == nil {
				panic(tlc.NewNullPointerException(""))
			}
			code := "E4240"
			if node.getKind() == sanyAPSubstInKind {
				code = "E4004"
			}
			diagnostics = append(diagnostics, sanyRegistrationDiagnostic(sanyNodePosition(node.TreeNode.(*SanySyntaxNode)), code, "Substitution missing for symbol %s declared at %s \nand instantiated in module %s.", declaration.semName(), sanySymbolLocation(declaration.semPosition()), node.instantiatingModule.semName()))
		}
	}
	return diagnostics
}
