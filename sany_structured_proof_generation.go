// Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// Walk syntax in source order while the existing projected statement generator
// supplies each body. Exit events finish subproofs before their theorem owners.
type sanyProofGraphEventKind uint8

const (
	sanyProofEnter sanyProofGraphEventKind = iota
	sanyProofStepBegin
	sanyProofStepEnd
	sanyProofExit
)

type sanyProofGraphEvent struct {
	kind   sanyProofGraphEventKind
	syntax *SanySyntaxNode
}
type sanyProofGraphFrame struct {
	syntax   *SanySyntaxNode
	context  *sanyContext
	steps    []sanySemanticGraphNode
	complete bool
}
type sanyProofGraphStep struct {
	syntax, bodySyntax          *SanySyntaxNode
	goal                        *sanySemThmOrAssumpDefNode
	name                        *tlc.UniqueString
	finishLabels                func() *sanyLabelTable
	body, node                  sanySemanticGraphNode
	complete, theorem, suffices bool
	previousUnsupported         bool
}
type sanyProofGraphGeneration struct {
	g       *sanyExpressionGeneration
	events  []sanyProofGraphEvent
	index   int
	frames  []*sanyProofGraphFrame
	steps   map[*SanySyntaxNode]*sanyProofGraphStep
	current *sanyProofGraphStep
}

func newSanyProofGraphGeneration(g *sanyExpressionGeneration, syntax *SanySyntaxNode) *sanyProofGraphGeneration {
	root := sanyAttachedProofSyntax(syntax)
	if root == nil || root.Kind.JavaName() != "N_Proof" {
		return nil
	}
	generation := &sanyProofGraphGeneration{g: g, steps: make(map[*SanySyntaxNode]*sanyProofGraphStep)}
	var walk func(*SanySyntaxNode)
	walk = func(proof *SanySyntaxNode) {
		generation.events = append(generation.events, sanyProofGraphEvent{sanyProofEnter, proof})
		for _, step := range proof.GetHeirs() {
			if step.Kind.JavaName() != "N_ProofStep" {
				continue
			}
			generation.events = append(generation.events, sanyProofGraphEvent{sanyProofStepBegin, step})
			if nested := sanyAttachedProofSyntax(step); nested != nil && nested.Kind.JavaName() == "N_Proof" {
				walk(nested)
			}
			generation.events = append(generation.events, sanyProofGraphEvent{sanyProofStepEnd, step})
		}
		generation.events = append(generation.events, sanyProofGraphEvent{sanyProofExit, proof})
	}
	walk(root)
	return generation
}
func sanyAttachedProofSyntax(syntax *SanySyntaxNode) *SanySyntaxNode {
	if syntax == nil {
		return nil
	}
	if syntax.Kind.JavaName() == "N_Proof" || syntax.Kind.JavaName() == "N_TerminalProof" {
		return syntax
	}
	heirs := syntax.GetHeirs()
	if len(heirs) == 0 {
		return nil
	}
	last := heirs[len(heirs)-1]
	if last.Kind.JavaName() == "N_Proof" || last.Kind.JavaName() == "N_TerminalProof" {
		return last
	}
	return nil
}
func (generation *sanyProofGraphGeneration) beforeStep(syntax *SanySyntaxNode) {
	if generation == nil {
		return
	}
	for generation.index < len(generation.events) {
		event := generation.events[generation.index]
		generation.index++
		generation.advance(event)
		if event.kind == sanyProofStepBegin {
			if event.syntax != syntax {
				panic("proof projection is out of syntax order")
			}
			return
		}
	}
	panic("proof projection has extra steps")
}
func (generation *sanyProofGraphGeneration) finish() {
	if generation == nil {
		return
	}
	for generation.index < len(generation.events) {
		event := generation.events[generation.index]
		generation.index++
		if event.kind == sanyProofStepBegin {
			panic("proof projection omitted a syntax step")
		}
		generation.advance(event)
	}
}
func (generation *sanyProofGraphGeneration) advance(event sanyProofGraphEvent) {
	g := generation.g
	switch event.kind {
	case sanyProofEnter:
		context := newSanyContext()
		g.formalSymbolTable().pushContext(context)
		generation.frames = append(generation.frames, &sanyProofGraphFrame{syntax: event.syntax, context: context, steps: make([]sanySemanticGraphNode, 0), complete: true})
	case sanyProofStepBegin:
		heirs := event.syntax.GetHeirs()
		step := &sanyProofGraphStep{syntax: event.syntax, bodySyntax: heirs[1], complete: true, previousUnsupported: g.labelGoalUnsupported}
		number := heirs[0]
		if number.Token != nil {
			switch number.Token.Kind {
			case SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme, SanyTokenProofstepdotlexeme:
				name := number.Image
				if dot := strings.IndexByte(name, '.'); dot >= 0 {
					name = name[:dot]
				}
				step.name = tlc.UniqueStringOf(name)
			}
		}
		if step.name != nil {
			step.finishLabels = g.pushLabelScope()
			step.goal = newSanySemThmOrAssumpDefNode(step.name.String(), number)
		}
		g.labelGoalUnsupported = false
		generation.steps[event.syntax] = step
		generation.current = step
	case sanyProofStepEnd:
		step := generation.steps[event.syntax]
		if step.theorem && step.complete {
			proof, complete := g.completedTheoremProof(step.syntax)
			if complete {
				node := newSanySemTheoremNode(step.bodySyntax, step.body, g.currentModule.semanticNode, proof, step.goal)
				sanyAssertionSyntax(&node.sanySemanticNode, step.syntax)
				node.suffices = step.suffices
				step.node = node
			} else {
				step.complete = false
			}
		}
		if step.finishLabels != nil {
			step.finishLabels()
			step.finishLabels = nil
		}
		frame := generation.frames[len(generation.frames)-1]
		if step.node == nil || !step.complete {
			frame.complete = false
		} else {
			frame.steps = append(frame.steps, step.node)
		}
	case sanyProofExit:
		frame := generation.frames[len(generation.frames)-1]
		g.formalSymbolTable().popContext()
		generation.frames = generation.frames[:len(generation.frames)-1]
		if frame.complete {
			node := newSanySemNonLeafProofNode(frame.syntax, frame.steps, make([]sanySemanticGraphNode, 0), frame.context)
			if g.structuredProofGraphs == nil {
				g.structuredProofGraphs = make(map[*SanySyntaxNode]*sanySemNonLeafProofNode)
			}
			g.structuredProofGraphs[frame.syntax] = node
		}
	}
}
func (generation *sanyProofGraphGeneration) statement(step *ProofStep, useHide *sanySemUseOrHideNode, context map[string]Position) Diagnostics {
	if generation == nil {
		return nil
	}
	g := generation.g
	graph := generation.current
	var diagnostics Diagnostics
	switch step.Kind {
	case "DEFINE":
		definitions := make([]*sanySemOpDefNode, 0)
		for _, unit := range sanyProofDefinitionUnits(step) {
			if unit.definition == nil {
				graph.complete = false
				continue
			}
			node := unit.definition.semanticNode
			if node == nil {
				graph.complete = false
				continue
			}
			definitions = append(definitions, node)
			if !unit.definition.FunctionDef {
				g.currentModule.semanticNode.definitions = append(g.currentModule.semanticNode.definitions, node)
			}
		}
		if graph.complete {
			graph.node = newSanySemDefStepNode(graph.bodySyntax, graph.name, definitions)
		}
	case "USE", "HIDE":
		if useHide == nil {
			graph.complete = false
		} else {
			sanyAssertionSyntax(&useHide.sanySemanticNode, graph.syntax)
			useHide.setStepName(graph.name)
			graph.node = useHide
		}
	case "INSTANCE":
		graph.complete = false
	default:
		graph.theorem = true
		graph.suffices = step.Suffices
		if step.AssumeProveBody != nil || step.Kind == "PICK" || step.Kind == "TAKE" {
			graph.complete = false
			break
		}
		operands := make([]sanySemanticGraphNode, 0)
		op := ""
		switch step.Kind {
		case "ASSERT":
			graph.body = sanyGeneratedExpressionNode(step.Expr)
			if graph.body == nil && sanyExpressionGenerationFailure(step.Expr) != sanyGenerationNullExpression {
				graph.complete = false
			}
			if step.Suffices {
				op = "$Suffices"
				operands = append(operands, graph.body)
			}
		case "HAVE", "CASE":
			op = "$Have"
			if step.Kind == "CASE" {
				op = "$Pfcase"
			}
			body := sanyGeneratedExpressionNode(step.Expr)
			if body == nil && sanyExpressionGenerationFailure(step.Expr) != sanyGenerationNullExpression {
				graph.complete = false
			}
			operands = append(operands, body)
		case "WITNESS":
			op = "$Witness"
			for _, expression := range step.Exprs {
				body := sanyGeneratedExpressionNode(expression)
				if body == nil && sanyExpressionGenerationFailure(expression) != sanyGenerationNullExpression {
					graph.complete = false
				}
				operands = append(operands, body)
			}
		case "QED":
			op = "$Qed"
		default:
			graph.complete = false
		}
		if op != "" && graph.complete {
			node, diags, err := newSanySemOpApplNode(g.applicationOperator(op, nil, context), operands, graph.bodySyntax)
			diagnostics = append(diagnostics, diags...)
			if err != nil {
				panic(err)
			}
			graph.body = node
		}
		if graph.goal != nil && graph.complete {
			diagnostics = append(diagnostics, graph.goal.construct(true, graph.body, g.currentModule.semanticNode, g.formalSymbolTable(), nil)...)
			graph.goal.setLabels(graph.finishLabels())
			graph.finishLabels = nil
			if graph.suffices {
				graph.goal.setSuffices()
			}
		}
	}
	if !graph.theorem && graph.name != nil && graph.complete {
		node, diags := newSanySemNumberedProofStepNode(graph.name.String(), graph.node, g.currentModule.semanticNode, g.formalSymbolTable(), graph.syntax)
		diagnostics = append(diagnostics, diags...)
		node.setLabels(graph.finishLabels())
		graph.finishLabels = nil
	}
	if graph.finishLabels != nil {
		graph.finishLabels()
		graph.finishLabels = nil
	}
	g.labelGoalUnsupported = graph.previousUnsupported
	return diagnostics
}

// An @-step generates its right operand first, then wraps the existing previous
// RHS in $Nop. Ordinary binary generation would both lose sharing and reorder UIDs.
func (g *sanyExpressionGeneration) proofInfixExpression(expr Expr, context map[string]Position) (Diagnostics, bool) {
	infix, ok := expr.(*BinaryExpr)
	if !ok {
		return nil, false
	}
	left, ok := infix.Left.(*IdentExpr)
	if !ok || left.proofAtTarget == nil {
		return nil, false
	}
	previous := sanyGeneratedExpressionNode(left.proofAtTarget)
	operator := g.applicationOperator(infix.Op, infix.Syntax, context)
	if previous == nil || operator == nil {
		return nil, false
	}
	diagnostics := g.proofExpression(infix.Right, context, nil)
	right := sanyGeneratedExpressionNode(infix.Right)
	if right == nil && sanyExpressionGenerationFailure(infix.Right) != sanyGenerationNullExpression {
		return diagnostics, true
	}
	nop, diags, err := newSanySemOpApplNode(g.applicationOperator("$Nop", nil, context), []sanySemanticGraphNode{previous}, left.Syntax)
	diagnostics = append(diagnostics, diags...)
	if err != nil {
		panic(err)
	}
	left.semanticGraph = nop
	arity := 0
	left.generationArity = &arity
	node, diags, err := newSanySemOpApplNode(operator, []sanySemanticGraphNode{nop, right}, infix.Syntax)
	diagnostics = append(diagnostics, diags...)
	if err != nil {
		panic(err)
	}
	infix.semanticGraph = node
	return diagnostics, true
}

// A parameter-free named label selection wraps the existing LabelNode itself.
// Qualified INSTANCE, operand and parameterized selectors require further porting.
func (g *sanyExpressionGeneration) retainCanonicalLabelSelection(expr Expr) Diagnostics {
	source := sanyExprSource(expr)
	if source == nil || source.Selector == nil || len(source.Selector.Steps) != 2 {
		return nil
	}
	steps := source.Selector.Steps
	for _, step := range steps {
		if step.Kind != SanySelectorName || step.Arguments != nil {
			return nil
		}
	}
	symbol := g.formalSymbolTable().resolveSymbol(steps[0].Name)
	if symbol == nil || symbol.semArity() != 0 {
		return nil
	}
	var label *sanySemLabelNode
	switch definition := symbol.(type) {
	case *sanySemOpDefNode:
		label = definition.getLabel(steps[1].Name)
	case *sanySemThmOrAssumpDefNode:
		label = definition.getLabel(steps[1].Name)
	}
	if label == nil || label.arity != 0 {
		return nil
	}
	node, diagnostics, err := newSanySemOpApplNode(g.applicationOperator("$Nop", nil, nil), []sanySemanticGraphNode{label}, source.Syntax)
	if err != nil {
		panic(err)
	}
	source.semanticGraph = node
	return diagnostics
}
