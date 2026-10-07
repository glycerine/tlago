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
	labels     *sanyLabelTable
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

func (node *sanySemLabelNode) setLabels(labels *sanyLabelTable)       { node.labels = labels }
func (node *sanySemLabelNode) getLabel(name string) *sanySemLabelNode { return node.labels.get(name) }
func (node *sanySemLabelNode) addLabel(label *sanySemLabelNode) bool {
	if node.labels == nil {
		node.labels = newSanyLabelTable()
	}
	return node.labels.add(label)
}
func (node *sanySemLabelNode) getLabels() []*sanySemLabelNode { return node.labels.elements() }
func (node *sanySemLabelNode) getName() string                { return node.name }
func (node *sanySemLabelNode) getArity() int                  { return node.arity }
func (node *sanySemLabelNode) getBody() sanySemanticGraphNode { return node.body }
func (node *sanySemLabelNode) getGoal() sanySemanticGraphNode { return node.goal }
func (node *sanySemLabelNode) getChildren() []sanySemanticGraphNode {
	return []sanySemanticGraphNode{node.body}
}

func (node *sanySemOpDefNode) setLabels(labels *sanyLabelTable)       { node.labels = labels }
func (node *sanySemOpDefNode) getLabel(name string) *sanySemLabelNode { return node.labels.get(name) }
func (node *sanySemOpDefNode) addLabel(label *sanySemLabelNode) bool {
	if node.labels == nil {
		node.labels = newSanyLabelTable()
	}
	return node.labels.add(label)
}
func (node *sanySemOpDefNode) getLabels() []*sanySemLabelNode { return node.labels.elements() }
func (node *sanySemOpDefNode) getLabelsHT() *sanyLabelTable   { return node.labels }

func (g *sanyExpressionGeneration) pushLabelScope() func() *sanyLabelTable {
	previousEnabled := g.labelsEnabled
	g.labelsEnabled = true
	scope := &sanyLabelScope{parameters: make([][]*sanyFormalParamNode, 0)}
	g.labelScopes = append(g.labelScopes, scope)
	return func() *sanyLabelTable {
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
	if len(g.excepts) > 0 && len(g.exceptSpecs) > 0 {
		return guard("E4335", "Labels inside EXCEPT clauses are not yet implemented.")
	}
	finishBodyLabels := g.pushLabelScope()
	var diagnostics Diagnostics
	var body sanySemanticGraphNode
	if label.AssumeProveBody != nil {
		diagnostics = g.checkAssumeProveBody(label.AssumeProveBody, context, locals, false)
		if label.AssumeProveBody.semanticNode != nil {
			body = label.AssumeProveBody.semanticNode
		}
	} else {
		diagnostics = g.checkExpr(label.Body, context, locals)
		body = sanyGeneratedExpressionNode(label.Body)
	}
	bodyLabels := finishBodyLabels()
	g.resolveLabelFormals(label, context)
	if body == nil && sanyExpressionGenerationFailure(label.Body) == sanyGenerationNullOperator {
		g.retainNullOperatorOperand(label.Body, false)
		body = sanyGeneratedExpressionNode(label.Body)
	}
	// Ordinary expression labels retain the generator's current goal and clause.
	var goal sanySemanticGraphNode
	if g.currentGoal != nil {
		goal = g.currentGoal
	}
	node := newSanySemLabelNode(label.Syntax, label.Name, label.formalNodes, goal, g.currentGoalClause, body, label.AssumeProveBody != nil)
	node.setLabels(bodyLabels)
	if body != nil || sanyExpressionGenerationFailure(label.Body) == sanyGenerationNullExpression {
		label.semanticGraph = node
	}
	scope := g.labelScopes[len(g.labelScopes)-1]
	check := labelCheckContext{allowed: true, formalGroups: scope.parameters}
	diagnostics = append(diagnostics, checkLabelParameters(label, check)...)
	if scope.labels == nil {
		scope.labels = newSanyLabelTable()
	}
	if !scope.labels.add(node) {
		diagnostic := sanyDiagnosticParameters(errorAt(label.Pos, "E4336", "Duplicate label %s", label.Name), label.Name)
		diagnostic.SANYMessage = fmt.Sprintf("Duplicate label `%s'.", label.Name)
		if label.Syntax != nil {
			diagnostic.SANYRange = label.Syntax.Range
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	return diagnostics
}
