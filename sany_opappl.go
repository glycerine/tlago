// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"strings"
)

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

// Unbounded OpApplNode constructor 4 retains the caller's formal array.
func newSanySemUnboundedOpApplNode(name string, operands []sanySemanticGraphNode, parameters []*sanyFormalParamNode, syntax *SanySyntaxNode) *sanySemOpApplNode {
	node := newSanySemBuiltInOpApplNode(name, operands, syntax)
	node.unboundedBoundSymbols = parameters
	return node
}

// Bounded constructor 5 retains all arrays, including nil ranges if supplied.
func newSanySemBoundedOpApplNode(name string, functionNames []*sanyFormalParamNode, operands []sanySemanticGraphNode, parameters [][]*sanyFormalParamNode, tuples []bool, ranges []sanySemanticGraphNode, syntax *SanySyntaxNode) *sanySemOpApplNode {
	node := newSanySemBuiltInOpApplNode(name, operands, syntax)
	node.unboundedBoundSymbols = functionNames
	node.boundedBoundSymbols = parameters
	node.tupleOrs = tuples
	node.ranges = ranges
	return node
}

// The general constructor invokes the concrete symbol's match after field
// initialization. A false result is deliberately ignored; a thrown error is
// returned with its diagnostic and no completed application result.
func newSanySemOpApplNode(operator sanySemSymbol, operands []sanySemanticGraphNode, syntax *SanySyntaxNode) (*sanySemOpApplNode, Diagnostics, error) {
	node := &sanySemOpApplNode{sanySemanticNode: newSanySemanticNode(sanyOpApplKind), operator: operator, operands: operands, ranges: make([]sanySemanticGraphNode, 0)}
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	}
	var diagnostics Diagnostics
	var err error
	switch symbol := operator.(type) {
	case *sanyFormalParamNode:
		symbol.match(node)
	case *sanySemOpDeclNode:
		_, diagnostics, err = symbol.match(node)
	case *sanySemOpDefNode:
		_, diagnostics, err = symbol.match(node)
	default:
		panic("unported semantic symbol match")
	}
	if err != nil {
		return nil, diagnostics, err
	}
	return node, diagnostics, nil
}

// OpArgNode retains the resolved symbol and its name/arity without matching.
func newSanySemOpArgNode(operator sanySemSymbol, syntax *SanySyntaxNode, module *sanySemModuleNode) *sanySemOpArgNode {
	node := &sanySemOpArgNode{sanySemanticNode: newSanySemanticNode(sanyOpArgKind), operator: operator, module: module}
	// Source's primary constructor dereferences op despite its historical null comment.
	node.name, node.arity = operator.semName(), operator.semArity()
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	}
	return node
}

func retainSanySymbolReference(expr Expr, operator sanySemSymbol, asOperator bool, module *Module) Diagnostics {
	source := sanyGenerationSource(expr)
	if source == nil || operator == nil {
		return nil
	}
	if asOperator {
		var semanticModule *sanySemModuleNode
		if module != nil {
			semanticModule = module.semanticNode
		}
		source.semanticGraph = newSanySemOpArgNode(operator, source.Syntax, semanticModule)
		return nil
	}
	if operator.semArity() != 0 {
		return nil
	}
	node, diagnostics, err := newSanySemOpApplNode(operator, make([]sanySemanticGraphNode, 0), source.Syntax)
	if err != nil {
		panic(err)
	}
	source.semanticGraph = node
	return diagnostics
}

// Parsed names sharing one domain belong to one bound group. Distinct domain
// syntax expressions keep their own groups, even when their text is equal.
func retainSanyBoundApplication(expr Expr, name string, bounds []BoundVar, parameters []*sanyFormalParamNode, body Expr) {
	source := sanyGenerationSource(expr)
	operand := sanyGeneratedExpressionNode(body)
	if source == nil || operand == nil || len(bounds) != len(parameters) {
		return
	}
	groups := make([][]*sanyFormalParamNode, 0)
	tuples := make([]bool, 0)
	ranges := make([]sanySemanticGraphNode, 0)
	for i := 0; i < len(bounds); {
		domain := bounds[i].Set
		rangeNode := sanyGeneratedExpressionNode(domain)
		if rangeNode == nil {
			return
		}
		end := i + 1
		for end < len(bounds) && bounds[end].Set == domain {
			end++
		}
		group := make([]*sanyFormalParamNode, end-i)
		copy(group, parameters[i:end])
		groups = append(groups, group)
		tuples = append(tuples, bounds[i].TupleBound)
		ranges = append(ranges, rangeNode)
		i = end
	}
	source.semanticGraph = newSanySemBoundedOpApplNode(name, nil, []sanySemanticGraphNode{operand}, groups, tuples, ranges, source.Syntax)
}

// Generator resolves the raw GenID, including prefix '-' becoming '-.', before
// generating operands. The native AST's normalized spelling cannot replace it.
func (g *sanyExpressionGeneration) applicationOperator(op string, syntax *SanySyntaxNode, context map[string]Position) sanySemSymbol {
	name := op
	if syntax != nil {
		heirs := syntax.GetHeirs()
		index := -1
		switch syntax.Kind.JavaName() {
		case "N_PrefixExpr":
			index = 0
		case "N_InfixExpr", "N_PostfixExpr":
			index = 1
		}
		if index >= 0 && index < len(heirs) {
			if raw := sanyOperatorGenIDName(heirs[index]); raw != "" {
				name = raw
				if syntax.Kind.JavaName() == "N_PrefixExpr" && (name == "-" || strings.HasSuffix(name, "!-")) {
					name += "."
				}
			}
		}
	}
	name = ResolveSanyOperatorSynonym(name)
	if symbol := g.formalSymbolTable().resolveSymbol(name); symbol != nil {
		return symbol
	}
	if symbol, ok := g.lookupSymbol(name, context); ok {
		if symbol.opDefNode != nil {
			return symbol.opDefNode
		}
		if symbol.formalNode != nil {
			return symbol.formalNode
		}
		if symbol.declarationNode != nil {
			return symbol.declarationNode
		}
	}
	return nil
}

func retainSanyMatchedApplication(expr Expr, operator sanySemSymbol, children []Expr) Diagnostics {
	source := sanyGenerationSource(expr)
	if source == nil || operator == nil {
		return nil
	}
	operands := make([]sanySemanticGraphNode, len(children))
	for i, child := range children {
		operands[i] = sanyGeneratedExpressionNode(child)
		if operands[i] == nil {
			return nil
		}
	}
	node, diagnostics, err := newSanySemOpApplNode(operator, operands, source.Syntax)
	if err != nil {
		panic(err)
	}
	source.semanticGraph = node
	return diagnostics
}

// A single parser junction/product node is flattened into native wrappers.
// Only wrappers lacking their own syntax belong to this source application;
// explicitly nested expressions retain separate applications and identities.
func sanySourceNaryOperands(root *BinaryExpr) []Expr {
	var operands []Expr
	var collect func(Expr)
	collect = func(expr Expr) {
		if nested, ok := expr.(*BinaryExpr); ok && nested.Syntax == nil && nested.Op == root.Op && ((root.JunctionList && nested.JunctionList) || (root.Syntax != nil && root.Syntax.Kind.JavaName() == "N_Times" && nested.SanyNary)) {
			collect(nested.Left)
			collect(nested.Right)
			return
		}
		operands = append(operands, expr)
	}
	collect(root.Left)
	collect(root.Right)
	return operands
}

// CASE allocates each pair after its two expressions, before the next arm.
// OTHER has an actual null condition; it is not a string or placeholder node.
func sanyGeneratedCasePair(condition, value Expr, syntax *SanySyntaxNode, other bool) sanySemanticGraphNode {
	conditionNode, valueNode := sanyGeneratedExpressionNode(condition), sanyGeneratedExpressionNode(value)
	if syntax == nil || valueNode == nil || (!other && conditionNode == nil) {
		return nil
	}
	return newSanySemBuiltInOpApplNode("$Pair", []sanySemanticGraphNode{conditionNode, valueNode}, syntax)
}

func retainSanyGeneratedOperands(expr Expr, name string, operands []sanySemanticGraphNode) {
	source := sanyGenerationSource(expr)
	if source == nil {
		return
	}
	for _, operand := range operands {
		if operand == nil {
			return
		}
	}
	source.semanticGraph = newSanySemBuiltInOpApplNode(name, operands, source.Syntax)
}
