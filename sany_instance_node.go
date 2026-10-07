// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"reflect"

	"github.com/glycerine/tlago/tlc"
)

// Subst is not a SemanticNode: constructing or mutating it does not allocate a
// semantic UID. It retains the original declaration, expression and syntax.
type sanySemSubst struct {
	op       *sanySemOpDeclNode
	expr     sanySemanticGraphNode
	exprSTN  *SanySyntaxNode
	implicit bool
}

func newSanySemSubst(op *sanySemOpDeclNode, expr sanySemanticGraphNode, syntax *SanySyntaxNode, implicit bool) *sanySemSubst {
	return &sanySemSubst{op: op, expr: expr, exprSTN: syntax, implicit: implicit}
}
func (subst *sanySemSubst) getOp() *sanySemOpDeclNode {
	if subst == nil {
		panic(tlc.NewNullPointerException(""))
	}
	return subst.op
}
func (subst *sanySemSubst) setOp(op *sanySemOpDeclNode) { subst.op = op }
func (subst *sanySemSubst) getExpr() sanySemanticGraphNode {
	if subst == nil {
		panic(tlc.NewNullPointerException(""))
	}
	return subst.expr
}
func (subst *sanySemSubst) setExpr(expr sanySemanticGraphNode, implicit bool) {
	subst.expr, subst.implicit = expr, implicit
}
func (subst *sanySemSubst) getExprSTN() *SanySyntaxNode       { return subst.exprSTN }
func (subst *sanySemSubst) setExprSTN(syntax *SanySyntaxNode) { subst.exprSTN = syntax }
func (subst *sanySemSubst) isImplicit() bool                  { return subst.implicit }
func sanySubstGetSub(param any, substitutions []*sanySemSubst) sanySemanticGraphNode {
	if symbol, ok := param.(sanySemSymbol); ok {
		param = sanyLevelSymbolReference(symbol)
	}
	if param != nil {
		value := reflect.ValueOf(param)
		switch value.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
			if value.IsNil() {
				param = nil
			}
		}
	}

	if substitutions == nil {
		panic(tlc.NewNullPointerException(""))
	}
	for _, subst := range substitutions {
		op := subst.getOp()
		if (op == nil && param == nil) || (op != nil && op == param) {
			return subst.getExpr()
		}
	}
	return nil
}

// InstanceNode normalizes null parameter/substitution arrays to distinct empty
// arrays; supplied arrays, their entries and the target module retain identity.
type sanySemInstanceNode struct {
	sanySemanticNode
	name     *tlc.UniqueString
	params   []*sanyFormalParamNode
	module   *sanySemModuleNode
	substs   []*sanySemSubst
	local    bool
	stepName *tlc.UniqueString
}

func newSanySemInstanceNode(name *tlc.UniqueString, local bool, params []*sanyFormalParamNode, module *sanySemModuleNode, substs []*sanySemSubst, syntax *SanySyntaxNode) *sanySemInstanceNode {
	node := &sanySemInstanceNode{sanySemanticNode: newSanySemanticNode(sanyInstanceKind), name: name, local: local, params: params, module: module, substs: substs}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	if node.params == nil {
		node.params = make([]*sanyFormalParamNode, 0)
	}
	if node.substs == nil {
		node.substs = make([]*sanySemSubst, 0)
	}
	return node
}
func (node *sanySemInstanceNode) getName() *tlc.UniqueString         { return node.name }
func (node *sanySemInstanceNode) getModule() *sanySemModuleNode      { return node.module }
func (node *sanySemInstanceNode) getLocal() bool                     { return node.local }
func (node *sanySemInstanceNode) isLocal() bool                      { return node.local }
func (node *sanySemInstanceNode) setStepName(name *tlc.UniqueString) { node.stepName = name }
func (node *sanySemInstanceNode) getStepName() *tlc.UniqueString     { return node.stepName }
func (node *sanySemInstanceNode) getLevel() int                      { return 0 }
func (node *sanySemInstanceNode) getChildren() []sanySemanticGraphNode {
	if node.substs == nil {
		panic(tlc.NewNullPointerException(""))
	}
	children := make([]sanySemanticGraphNode, len(node.substs))
	for i, subst := range node.substs {
		children[i] = subst.getExpr()
	}
	return children
}
