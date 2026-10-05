// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import "strconv"

// SanySelector is Generator.Selector: arguments stay as syntax until the
// selected operator is known, so each argument can be generated with its
// expected expression or operator arity.
type SanySelector struct {
	Syntax *SanySyntaxNode
	Steps  []SanySelectorStep
}

// SanySelectorStep corresponds to one entry of Generator.Selector's parallel
// ops, opNames, opsSTN and args arrays. Positive Kind values are operand indices.
type SanySelectorStep struct {
	Kind      int
	Name      string
	Syntax    *SanySyntaxNode
	Arguments *SanySyntaxNode
}

const (
	SanySelectorName  = -1
	SanySelectorNull  = -2
	SanySelectorLast  = -3
	SanySelectorFirst = -4
	SanySelectorColon = -5
	SanySelectorAt    = -6
)

// sanySelectorFromSyntax mirrors genIdToSelector and N_OpApplication. It does
// not infer selectors from the flattened identifier image: argument-only
// selectors have no image component, and the original nodes retain positions.
func sanySelectorFromSyntax(node *SanySyntaxNode) *SanySelector {
	if node == nil {
		return nil
	}
	root := node
	var finalArguments *SanySyntaxNode
	if node.Kind.JavaName() == "N_OpApplication" {
		heirs := node.GetHeirs()
		if len(heirs) != 2 || heirs[1].Kind.JavaName() != "N_OpArgs" {
			return nil
		}
		node, finalArguments = heirs[0], heirs[1]
	}
	if node.Kind.JavaName() != "N_GeneralId" {
		return nil
	}
	heirs := node.GetHeirs()
	if len(heirs) != 2 {
		return nil
	}
	selector := &SanySelector{Syntax: root}
	for _, prefix := range heirs[0].GetHeirs() {
		parts := prefix.GetHeirs()
		if len(parts) == 0 {
			return nil
		}
		step, ok := sanySelectorStep(parts[0])
		if !ok {
			return nil
		}
		if step.Kind == SanySelectorNull {
			step.Arguments = parts[0]
		} else if step.Kind == SanySelectorName && len(parts) == 3 {
			if parts[1].Kind.JavaName() != "N_OpArgs" {
				return nil
			}
			step.Arguments = parts[1]
		}
		selector.Steps = append(selector.Steps, step)
	}
	last, ok := sanySelectorStep(heirs[1])
	if !ok {
		return nil
	}
	if last.Kind == SanySelectorNull {
		last.Arguments = heirs[1]
	}
	if finalArguments != nil {
		last.Arguments = finalArguments
	}
	selector.Steps = append(selector.Steps, last)
	return selector
}

func sanySelectorStep(node *SanySyntaxNode) (SanySelectorStep, bool) {
	step := SanySelectorStep{Syntax: node, Name: sanySelectorName(node)}
	switch node.Kind.JavaName() {
	case "N_OpArgs":
		step.Kind = SanySelectorNull
	case "N_StructOp":
		switch step.Name {
		case ">>":
			step.Kind = SanySelectorLast
		case "<<":
			step.Kind = SanySelectorFirst
		case ":":
			step.Kind = SanySelectorColon
		case "@":
			step.Kind = SanySelectorAt
		default:
			index, err := strconv.Atoi(step.Name)
			if err != nil || index < 0 {
				return step, false
			}
			step.Kind = index
		}
	case "IDENTIFIER", "N_InfixOp", "N_NonExpPrefixOp", "N_PostfixOp", "N_PrefixOp", "ProofStepLexeme", "ProofImplicitStepLexeme":
		step.Kind = SanySelectorName
	default:
		return step, false
	}
	return step, true
}

// sanySelectorOperand implements the operand part of FindingSubExpr. These
// operands are SANY's semantic operands, which differ from the central AST's
// traversal children (a call's callee, for example, is not an operand).
func sanySelectorOperand(expr Expr, step SanySelectorStep) (Expr, bool) {
	for {
		label, ok := expr.(*LabelExpr)
		if !ok {
			break
		}
		expr = label.Body
	}
	operands := sanySelectorOperands(expr)
	index := step.Kind
	switch index {
	case SanySelectorFirst:
		index = 1
	case SanySelectorLast:
		index = len(operands)
	}
	if index < 1 || index > len(operands) {
		return nil, false
	}
	// Generator forbids selecting EXCEPT replacements: they could contain
	// a dangling @. The base expression remains a legal operand.
	if _, except := expr.(*ExceptExpr); except && index > 1 {
		return nil, false
	}
	return operands[index-1], operands[index-1] != nil
}

func sanySelectorOperands(expr Expr) []Expr {
	switch e := expr.(type) {
	case *CallExpr:
		return e.Args
	case *RecordComponentExpr:
		return []Expr{e.Record, &LiteralExpr{Kind: "string", Value: e.Field, Pos: e.FieldPos}}
	case *FunctionAppExpr:
		var argument Expr
		if len(e.Args) == 1 {
			argument = e.Args[0]
		} else {
			argument = &TupleExpr{Elems: e.Args, Pos: e.Pos}
		}
		return []Expr{e.Function, argument}
	case *CaseExpr:
		// $Case's operands are $Pair applications, not an interleaving of
		// guards and values. A second selector chooses one of the pair's two
		// operands; OTHER's nil guard cannot be selected.
		pairs := make([]Expr, 0, len(e.Arms)+1)
		for _, arm := range e.Arms {
			pairs = append(pairs, &TupleExpr{Elems: []Expr{arm.Test, arm.Value}, Pos: arm.Pos})
		}
		if e.Other != nil {
			pairs = append(pairs, &TupleExpr{Elems: []Expr{nil, e.Other}, Pos: e.OtherPos})
		}
		return pairs
	case *FairnessExpr:
		return []Expr{e.Subscript, e.Action}
	case *QuantifierExpr:
		var domains []Expr
		var previous Expr
		for q := e; q != nil; {
			if q.Set != nil && q.Set != previous {
				domains = append(domains, q.Set)
			}
			previous = q.Set
			next, ok := q.Body.(*QuantifierExpr)
			if !ok || e.Syntax == nil || next.Syntax != e.Syntax || next.Kind != e.Kind {
				break
			}
			q = next
		}
		return domains
	case *ChooseExpr:
		if e.Set != nil {
			return []Expr{e.Set}
		}
		return nil
	case *FunctionExpr:
		return sanySelectorBoundDomains(e.Bounds)
	case *SetComprehensionExpr:
		return sanySelectorBoundDomains(e.Bounds)
	case *ExceptExpr:
		operands := []Expr{e.Base}
		for _, spec := range e.Specs {
			operands = append(operands, spec.Value)
		}
		return operands
	case *BinaryExpr:
		if e.SanyNary && (e.Op == "\\X" || e.Op == "\\times") {
			var operands []Expr
			var collect func(Expr)
			collect = func(expr Expr) {
				if nested, ok := expr.(*BinaryExpr); ok && nested.SanyNary && nested.Op == e.Op && nested.Pos == e.Pos {
					collect(nested.Left)
					collect(nested.Right)
				} else {
					operands = append(operands, expr)
				}
			}
			collect(e)
			return operands
		}
		return sanySubexpressionChildren(e)
	default:
		return sanySubexpressionChildren(expr)
	}
}

func sanySelectorBoundDomains(bounds []BoundVar) []Expr {
	var domains []Expr
	var previous Expr
	for _, bound := range bounds {
		if bound.Set != nil && bound.Set != previous {
			domains = append(domains, bound.Set)
		}
		previous = bound.Set
	}
	return domains
}
