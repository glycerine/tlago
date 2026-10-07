// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"reflect"
)

func sanyGraphNodePresent(node sanySemanticGraphNode) bool {
	return node != nil && !reflect.ValueOf(node).IsNil()
}

func (n *sanySemTheoremNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.theoremLevelChecked >= iter {
		return true
	}
	n.theoremLevelChecked = iter
	sub := []sanySemanticGraphNode{n.theoremExprOrAssumeProve}
	if sanyGraphNodePresent(n.proof) {
		sub = append(sub, n.proof)
	}
	if n.def != nil {
		sub[0] = n.def
	}
	result := n.levelCheckGraphSubnodes(iter, sub, errors)
	if !sanyGraphNodePresent(n.theoremExprOrAssumeProve) {
		return result
	}
	var application *sanySemOpApplNode
	var operator sanySemSymbol
	if value, ok := n.theoremExprOrAssumeProve.(*sanySemOpApplNode); ok {
		application = value
		operator = value.operator
	}
	rawLevel := func() tlaLevel {
		return *sanyRequireCanonicalLevelNode(n.theoremExprOrAssumeProve).getLevelData().level
	}
	if sanyLevelSymbolReference(operator) != nil && operator.semName() == "$Pick" && application.ranges != nil && rawLevel() == temporalLevel {
		for i := 0; i < len(application.ranges); i++ {
			bound := sanyRequireCanonicalLevelNode(application.ranges[i])
			if bound.getLevel() != constantLevel {
				sanyAddFixedLevelMessage(errors, sanyGraphTreeNode(application.ranges[i]), "E4354", "Non-constant bound of temporal PICK.")
			}
		}
	}
	if rawLevel() == temporalLevel {
		sanyLevelCheckTemporal(n.proof, errors)
	}
	return result
}

func sanyLevelCheckTemporal(proof sanySemanticGraphNode, errors *Diagnostics) {
	if !sanyGraphNodePresent(proof) || proof.Kind() != tlc.SemanticKind(sanyNonLeafProofKind) {
		return
	}
	node, ok := proof.(*sanySemNonLeafProofNode)
	if !ok {
		panic(tlc.NewClassCastException("not a NonLeafProofNode"))
	}
	for i := 0; ; i++ {
		steps := node.getSteps()
		if steps == nil {
			panic(tlc.NewNullPointerException())
		}
		if i >= len(steps) {
			break
		}
		step := steps[i]
		if !sanyGraphNodePresent(step) {
			panic(tlc.NewNullPointerException())
		}
		var theorem *sanySemTheoremNode
		var application *sanySemOpApplNode
		if step.Kind() == tlc.SemanticKind(sanyTheoremKind) {
			theorem, ok = step.(*sanySemTheoremNode)
			if !ok {
				panic(tlc.NewClassCastException("not a TheoremNode"))
			}
			application, _ = theorem.theoremExprOrAssumeProve.(*sanySemOpApplNode)
		}
		if application != nil {
			operator := sanyLevelSymbolReference(application.operator)
			if operator == nil {
				panic(tlc.NewNullPointerException())
			}
			name := operator.semName()
			if (name == "$Take" || name == "$Witness" || name == "$Have") && application.getLevel() != constantLevel {
				sanyAddFixedLevelMessage(errors, application.TreeNode, "E4352", "Non-constant TAKE, WITNESS, or HAVE for temporal goal.")
			} else if name == "$Pfcase" {
				if application.getLevel() != constantLevel {
					sanyAddFixedLevelMessage(errors, application.TreeNode, "E4353", "Non-constant CASE for temporal goal.")
				}
				sanyLevelCheckTemporal(theorem.getProof(), errors)
			} else if name == "$Qed" {
				sanyLevelCheckTemporal(theorem.getProof(), errors)
			}
		}
	}
}

func sanyGraphTreeNode(node sanySemanticGraphNode) any {
	if !sanyGraphNodePresent(node) {
		panic(tlc.NewNullPointerException())
	}
	if value, ok := node.(interface{ GetTreeNode() any }); ok {
		return value.GetTreeNode()
	}
	panic(tlc.NewUnsupportedOperationException("semantic syntax ownership is not integrated"))
}

// Level diagnostics with fixed source formats retain parameters for source
// duplicate equality, even when two parameter lists render identical text.
func sanyAddFixedLevelMessage(errors *Diagnostics, tree any, code, message string, parameters ...any) {
	syntax, ok := tree.(*SanySyntaxNode)
	if tree == nil || ok && syntax == nil {
		panic(tlc.NewNullPointerException())
	}
	if !ok {
		panic(tlc.NewClassCastException("level-check syntax"))
	}
	location := syntax.Range
	if errors == nil {
		panic(tlc.NewNullPointerException())
	}
	diagnostic := errorAt(location.Begin, code, "%s", message)
	diagnostic.SANYRange = location
	diagnostic.SANYMessage = message
	diagnostic.SANYParameters = append([]any(nil), parameters...)
	for _, existing := range *errors {
		if existing.Code == code && existing.SANYMessage == message && reflect.DeepEqual(existing.SANYParameters, diagnostic.SANYParameters) && existing.SANYRange.Begin.File == location.Begin.File && existing.SANYRange.Begin.Line == location.Begin.Line && existing.SANYRange.Begin.Column == location.Begin.Column && existing.SANYRange.End.Line == location.End.Line && existing.SANYRange.End.Column == location.End.Column {
			return
		}
	}
	*errors = append(*errors, diagnostic)
}
