// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// AtNode retains the actual mutable EXCEPT and replacement pair, not copies.
type sanySemAtNode struct {
	sanySemanticNode
	exceptRef          *sanySemOpApplNode
	exceptComponentRef *sanySemOpApplNode
}

func newSanySemAtNode(except, component *sanySemOpApplNode) *sanySemAtNode {
	node := &sanySemAtNode{sanySemanticNode: newSanySemanticNode(sanyAtNodeKind), exceptRef: except, exceptComponentRef: component}
	node.TreeNode, node.Location = component.TreeNode, component.Location
	return node
}
func (n *sanySemAtNode) getExceptRef() *sanySemOpApplNode          { return n.exceptRef }
func (n *sanySemAtNode) getExceptComponentRef() *sanySemOpApplNode { return n.exceptComponentRef }
func (n *sanySemAtNode) getAtBase() sanySemanticGraphNode          { return n.exceptRef.operands[0] }
func (n *sanySemAtNode) getAtModifier() *sanySemOpApplNode {
	return n.exceptComponentRef.operands[0].(*sanySemOpApplNode)
}

// Generator.processExcept constructs EXCEPT before its specs and each Pair
// before its RHS. Constructors alias the operand arrays filled in afterward.
func (g *sanyExpressionGeneration) generateExcept(expr *ExceptExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	diagnostics := g.checkExpr(expr.Base, context, locals)
	operands := make([]sanySemanticGraphNode, len(expr.Specs)+1)
	operands[0] = sanyGeneratedExpressionNode(expr.Base)
	complete := operands[0] != nil || sanyExpressionGenerationFailure(expr.Base) == sanyGenerationNullExpression
	except := newSanySemBuiltInOpApplNode("$Except", operands, expr.Syntax)
	var children []*SanySyntaxNode
	if expr.Syntax != nil {
		children = expr.Syntax.GetHeirs()
	}
	for i, spec := range expr.Specs {
		var syntax *SanySyntaxNode
		if 3+2*i < len(children) {
			syntax = children[3+2*i]
		}
		var parts []*SanySyntaxNode
		if syntax != nil {
			parts = syntax.GetHeirs()
		}
		components := make([]sanySemanticGraphNode, len(spec.Components))
		for j, component := range spec.Components {
			var componentSyntax *SanySyntaxNode
			if 1+j < len(parts) {
				componentSyntax = parts[1+j]
			}
			if component.Field != "" {
				field := tlc.NewStringNode(component.Field)
				if componentSyntax != nil {
					heirs := componentSyntax.GetHeirs()
					if len(heirs) > 1 {
						bridge := tlcBridge{}
						bridge.withSyntaxNode(heirs[1], field)
						bridge.withPositionLocation(sanyNodePosition(heirs[1]), field)
					}
				}
				components[j] = field
			} else {
				indices := make([]sanySemanticGraphNode, len(component.Indices))
				for k, index := range component.Indices {
					diagnostics = append(diagnostics, g.checkExpr(index, context, locals)...)
					indices[k] = sanyGeneratedExpressionNode(index)
					complete = complete && (indices[k] != nil || sanyExpressionGenerationFailure(index) == sanyGenerationNullExpression)
				}
				if len(indices) == 1 {
					components[j] = indices[0]
				} else {
					// Preserve the source's tuple syntax selection, including its indexing.
					var tupleSyntax *SanySyntaxNode
					if parts != nil {
						tupleSyntax = parts[2*j+1]
					}
					components[j] = newSanySemBuiltInOpApplNode("$Tuple", indices, tupleSyntax)
				}
			}
		}
		pairOperands := []sanySemanticGraphNode{newSanySemBuiltInOpApplNode("$Seq", components, syntax), nil}
		pair := newSanySemBuiltInOpApplNode("$Pair", pairOperands, syntax)
		g.excepts = append(g.excepts, except)
		g.exceptSpecs = append(g.exceptSpecs, pair)
		replacementLocals := copyBoolMap(locals)
		replacementLocals["@"] = true
		diagnostics = append(diagnostics, g.checkExpr(spec.Value, context, replacementLocals)...)
		pairOperands[1] = sanyGeneratedExpressionNode(spec.Value)
		complete = complete && (pairOperands[1] != nil || sanyExpressionGenerationFailure(spec.Value) == sanyGenerationNullExpression)
		g.exceptSpecs = g.exceptSpecs[:len(g.exceptSpecs)-1]
		g.excepts = g.excepts[:len(g.excepts)-1]
		operands[i+1] = pair
	}
	if complete {
		expr.semanticGraph = except
	}
	return diagnostics
}
