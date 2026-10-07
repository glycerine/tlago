// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

// Function's source stack changes the first matching active specification when
// a GeneralId expression resolves to its name, even without a bracketed call.
type sanyFunctionGeneration struct {
	name        string
	application *sanySemOpApplNode
}

func (g *sanyExpressionGeneration) checkFunctionRecursion(name string) bool {
	for i := len(g.functions) - 1; i >= 0; i-- {
		if g.functions[i].name == name {
			g.functions[i].application.operator = sanyGlobalInitialContext(false).getSymbol("$RecursiveFcnSpec")
			return true
		}
	}
	return false
}

// processFunction constructs its specification and resolves/registers the
// definition before generating the body. Explicit RECURSIVE completion and
// recursive-section fields remain unported and do not receive substitutes.
func (g *sanyExpressionGeneration) prepareNamedFunctionDefinition(definition *Definition) Diagnostics {
	function, ok := definition.Expr.(*FunctionExpr)
	if !ok || g.bindings[definition.Name] != nil || (g.module != nil && g.module.sum != 0) {
		return nil
	}
	groups := make([][]*sanyFormalParamNode, 0)
	tuples := make([]bool, 0)
	ranges := make([]sanySemanticGraphNode, 0)
	for i := 0; i < len(function.Bounds); {
		bound := function.Bounds[i]
		domain := sanyGeneratedExpressionNode(bound.Set)
		if domain == nil && sanyExpressionGenerationFailure(bound.Set) == sanyGenerationNullOperator {
			g.retainNullOperatorOperand(bound.Set, false)
			domain = sanyGeneratedExpressionNode(bound.Set)
		}
		if domain == nil && sanyExpressionGenerationFailure(bound.Set) != sanyGenerationNullExpression {
			return nil
		}
		end := i + 1
		for end < len(function.Bounds) && function.Bounds[end].Set == bound.Set {
			end++
		}
		groups = append(groups, function.formalNodes[i:end])
		tuples = append(tuples, bound.TupleBound)
		ranges = append(ranges, domain)
		i = end
	}
	application := newSanySemBoundedOpApplNode("$NonRecursiveFcnSpec", []*sanyFormalParamNode{function.functionSymbol}, make([]sanySemanticGraphNode, 0), groups, tuples, ranges, definition.Syntax)
	function.functionApplication = application
	var node *sanySemOpDefNode
	var diagnostics Diagnostics
	if function.constructorSymbolExists {
		node = function.constructorSymbol.opDefNode
		if node == nil {
			node = function.constructorSymbol.builtinNode
		}
	} else {
		var module *sanySemModuleNode
		if g.currentModule != nil {
			module = g.currentModule.semanticNode
		}
		node, diagnostics = newSanySemOpDefNode(definition.Name, sanyUserDefinedOpKind, make([]*sanyFormalParamNode, 0), definition.Local, application, module, g.formalSymbolTable(), definition.Syntax, true, nil)
		node.letInLevel = g.level
	}
	definition.semanticNode = node
	if node != nil && g.currentModule != nil && g.currentModule.semanticNode != nil {
		g.currentModule.semanticNode.definitions = append(g.currentModule.semanticNode.definitions, node)
	}
	return diagnostics
}

func (g *sanyExpressionGeneration) finishNamedFunction(function *FunctionExpr) {
	application := function.functionApplication
	if application == nil {
		return
	}
	body := sanyGeneratedExpressionNode(function.Body)
	if body == nil && sanyExpressionGenerationFailure(function.Body) == sanyGenerationNullOperator {
		g.retainNullOperatorOperand(function.Body, false)
		body = sanyGeneratedExpressionNode(function.Body)
	}
	if body != nil || sanyExpressionGenerationFailure(function.Body) == sanyGenerationNullExpression {
		application.operands = []sanySemanticGraphNode{body}
		function.semanticGraph = application
	}
	if application.operator.semName() == "$NonRecursiveFcnSpec" {
		application.unboundedBoundSymbols = nil
	}
}

func (g *sanyExpressionGeneration) checkRejectedNamedFunctionBody(definition Definition, context map[string]Position, locals map[string]bool) Diagnostics {
	function, ok := definition.Expr.(*FunctionExpr)
	if !ok {
		return nil
	}
	if function.functionApplication != nil {
		g.functions = append(g.functions, sanyFunctionGeneration{definition.Name, function.functionApplication})
		defer func() { g.functions = g.functions[:len(g.functions)-1] }()
	}
	diagnostics := g.checkExpr(function.Body, context, locals)
	g.finishNamedFunction(function)
	return diagnostics
}
