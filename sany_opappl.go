// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// OpApplNode's builtin constructor deliberately skips operator.match. Its
// operator comes from the global context, not the enclosing module's table.
func newSanySemBuiltInOpApplNode(name string, operands []sanySemanticGraphNode, syntax *SanySyntaxNode) *sanySemOpApplNode {
	node := &sanySemOpApplNode{
		sanySemanticNode: newSanySemanticNode(sanyOpApplKind),
		operands:         operands, ranges: make([]sanySemanticGraphNode, 0),
	}
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	}
	node.operator = sanyGlobalInitialContext(false).getSymbol(name)
	return node
}

func sanyGenerationSource(expr Expr) *SanyExprSource {
	if source, ok := expr.(interface{ generationSource() *SanyExprSource }); ok {
		return source.generationSource()
	}
	return nil
}

func sanyGeneratedExpressionNode(expr Expr) sanySemanticGraphNode {
	if source := sanyGenerationSource(expr); source != nil {
		return source.semanticGraph
	}
	return nil
}

// Retain a builtin application only when every child has its actual generated
// graph. Unported child kinds remain visibly incomplete, rather than being
// represented by a placeholder or a regenerated evaluator node.
func retainSanyBuiltInApplication(expr Expr, name string, children []Expr) {
	source := sanyGenerationSource(expr)
	if source == nil {
		return
	}
	operands := make([]sanySemanticGraphNode, len(children))
	for i, child := range children {
		operands[i] = sanyGeneratedExpressionNode(child)
		if operands[i] == nil {
			return
		}
	}
	source.semanticGraph = newSanySemBuiltInOpApplNode(name, operands, source.Syntax)
}

// Multiple function arguments are represented by a $Tuple at the application's
// syntax node, including zero arguments, as in Generator's N_FcnAppl branch.
func retainSanyFunctionApplication(expr *FunctionAppExpr) {
	function := sanyGeneratedExpressionNode(expr.Function)
	if function == nil {
		return
	}
	args := make([]sanySemanticGraphNode, len(expr.Args))
	for i, argument := range expr.Args {
		args[i] = sanyGeneratedExpressionNode(argument)
		if args[i] == nil {
			return
		}
	}
	var argument sanySemanticGraphNode
	if len(args) == 1 {
		argument = args[0]
	} else {
		argument = newSanySemBuiltInOpApplNode("$Tuple", args, expr.Syntax)
	}
	expr.semanticGraph = newSanySemBuiltInOpApplNode("$FcnApply", []sanySemanticGraphNode{function, argument}, expr.Syntax)
}
