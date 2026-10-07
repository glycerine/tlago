// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"fmt"

	"github.com/glycerine/tlago/tlc"
)

// Generator.LSlabels and LSformalParams retain a separate label table and
// formal-group sequence for each definition or label body.
type sanyLabelScope struct {
	labels     map[string]*sanySemLabelNode
	parameters [][]*sanyFormalParamNode
}

func newSanySemLabelNode(syntax *SanySyntaxNode, name string, parameters []*sanyFormalParamNode, goal sanySemanticGraphNode, clause int, body sanySemanticGraphNode, assumeProve bool) *sanySemLabelNode {
	node := &sanySemLabelNode{sanySemanticNode: newSanySemanticNode(sanyLabelKind), name: name, formalNodes: parameters}
	if syntax == nil {
		node.TreeNode, node.Location = nil, tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	}
	if parameters == nil {
		panic(tlc.NewNullPointerException(""))
	}
	node.arity, node.goal, node.goalClause, node.body, node.isAssumeProve = len(parameters), goal, clause, body, assumeProve
	return node
}

func (node *sanySemLabelNode) setLabels(labels map[string]*sanySemLabelNode) { node.labels = labels }
func (node *sanySemLabelNode) getLabel(name string) *sanySemLabelNode        { return node.labels[name] }
func (node *sanySemLabelNode) addLabel(label *sanySemLabelNode) bool {
	if node.labels == nil {
		node.labels = make(map[string]*sanySemLabelNode)
	}
	if _, exists := node.labels[label.name]; exists {
		return false
	}
	node.labels[label.name] = label
	return true
}

func (g *sanyExpressionGeneration) pushLabelScope() func() map[string]*sanySemLabelNode {
	previousEnabled := g.labelsEnabled
	g.labelsEnabled = true
	scope := &sanyLabelScope{parameters: make([][]*sanyFormalParamNode, 0)}
	g.labelScopes = append(g.labelScopes, scope)
	return func() map[string]*sanySemLabelNode {
		if len(g.labelScopes) == 0 || g.labelScopes[len(g.labelScopes)-1] != scope {
			panic(tlc.NewWrongInvocationException("popLabelNodeSet called on empty stack."))
		}
		g.labelScopes = g.labelScopes[:len(g.labelScopes)-1]
		g.labelsEnabled = previousEnabled
		return scope.labels
	}
}

func (g *sanyExpressionGeneration) pushLabelFormals(parameters []*sanyFormalParamNode) func() {
	if !g.labelsEnabled || len(g.labelScopes) == 0 {
		return func() {}
	}
	scope := g.labelScopes[len(g.labelScopes)-1]
	scope.parameters = append(scope.parameters, parameters)
	return func() {
		if len(scope.parameters) == 0 {
			panic(tlc.NewWrongInvocationException("popFormalParams called on empty stack."))
		}
		scope.parameters = scope.parameters[:len(scope.parameters)-1]
	}
}

func (g *sanyExpressionGeneration) generateLabel(label *LabelExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	label.labelGenerated = true
	guard := func(code, message string) Diagnostics {
		if g.nodes != nil {
			label.semanticGraph = g.nodes.nullLabelNode
		}
		diagnostic := errorAt(label.Pos, code, "%s", message)
		diagnostic.SANYMessage = message
		if label.Syntax != nil {
			diagnostic.SANYRange = label.Syntax.Range
		}
		return Diagnostics{diagnostic}
	}
	if len(g.labelScopes) == 0 {
		return guard("E4333", "Label not in definition or proof step.")
	}
	if g.labelAPForbidden {
		return guard("E4334", "Label not allowed within scope of declaration in nested ASSUME/PROVE.")
	}
	if g.labelExceptDepth > 0 {
		return guard("E4335", "Labels inside EXCEPT clauses are not yet implemented.")
	}
	finishBodyLabels := g.pushLabelScope()
	diagnostics := g.checkExpr(label.Body, context, locals)
	bodyLabels := finishBodyLabels()
	g.resolveLabelFormals(label, context)
	body := sanyGeneratedExpressionNode(label.Body)
	if body == nil && sanyExpressionGenerationFailure(label.Body) == sanyGenerationNullOperator {
		g.retainNullOperatorOperand(label.Body, false)
		body = sanyGeneratedExpressionNode(label.Body)
	}
	// Canonical AP/goal nodes are handled separately; this path constructs
	// ordinary expression labels with the source's initial goal/clause values.
	node := newSanySemLabelNode(label.Syntax, label.Name, label.formalNodes, nil, 0, body, false)
	node.setLabels(bodyLabels)
	if body != nil || sanyExpressionGenerationFailure(label.Body) == sanyGenerationNullExpression {
		label.semanticGraph = node
	}
	scope := g.labelScopes[len(g.labelScopes)-1]
	check := labelCheckContext{allowed: true, formalGroups: scope.parameters}
	diagnostics = append(diagnostics, checkLabelParameters(label, check)...)
	if scope.labels == nil {
		scope.labels = make(map[string]*sanySemLabelNode)
	}
	if _, exists := scope.labels[label.Name]; exists {
		diagnostic := sanyDiagnosticParameters(errorAt(label.Pos, "E4336", "Duplicate label %s", label.Name), label.Name)
		diagnostic.SANYMessage = fmt.Sprintf("Duplicate label `%s'.", label.Name)
		if label.Syntax != nil {
			diagnostic.SANYRange = label.Syntax.Range
		}
		diagnostics = append(diagnostics, diagnostic)
	} else {
		scope.labels[label.Name] = node
	}
	return diagnostics
}
