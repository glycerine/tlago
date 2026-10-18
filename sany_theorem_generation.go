// Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

// Generate a theorem's statement and complete its named definition before the
// proof. The returned completion runs after proof generation, as processTheorem
// does, to close the AP context and then allocate the theorem owner.
func (g *sanyExpressionGeneration) generateTheoremStatement(theorem *NamedExpr, context map[string]Position) (Diagnostics, func()) {
	theorem.semanticNode, theorem.definitionNode = nil, nil
	var goal *sanySemThmOrAssumpDefNode
	var finishLabels func() *sanyLabelTable
	previousEnabled := g.labelsEnabled
	g.labelsEnabled = true
	if theorem.Name != "" {
		finishLabels = g.pushLabelScope()
		goal = newSanySemThmOrAssumpDefNode(theorem.Name, theorem.Syntax)
	}
	previousGoal := g.currentGoal
	previousSymbols := g.symbols
	var closeAPContext func()
	var body sanySemanticGraphNode
	var diagnostics Diagnostics
	if theorem.AssumeProveBody != nil {
		closeAPContext = g.pushFormalContext(0)
		g.currentGoal = goal
		previousOwned := g.outerAPContextOwned
		g.outerAPContextOwned = true
		diagnostics = g.checkAssumeProveBody(theorem.AssumeProveBody, context, nil, false)
		g.outerAPContextOwned = previousOwned
		g.currentGoal = previousGoal
		if theorem.AssumeProveBody.semanticNode != nil {
			body = theorem.AssumeProveBody.semanticNode
		}
	} else {
		diagnostics = g.checkExpr(theorem.Expr, context, nil)
		if sanyExpressionGenerationFailure(theorem.Expr) == sanyGenerationNullOperator {
			g.retainNullOperatorOperand(theorem.Expr, false)
		}
		body = sanyGeneratedExpressionNode(theorem.Expr)
	}
	complete := body != nil || theorem.AssumeProveBody == nil && sanyExpressionGenerationFailure(theorem.Expr) == sanyGenerationNullExpression
	if goal != nil && complete {
		table := g.formalSymbolTable()
		var apContext *sanyContext
		if closeAPContext != nil {
			apContext = table.topContext()
			table.popContext()
		}
		_, earlier := context[theorem.Name]
		if !earlier || table.resolveSymbol(theorem.Name) != nil {
			diagnostics = append(diagnostics, goal.construct(true, body, g.currentModule.semanticNode, table, nil)...)
			goal.setLabels(finishLabels())
			finishLabels = nil
			g.currentModule.semanticNode.definitions = append(g.currentModule.semanticNode.definitions, goal)
			theorem.definitionNode = goal
		} else {
			complete = false
		}
		if apContext != nil {
			table.pushContext(apContext)
		}
	}
	if finishLabels != nil {
		finishLabels()
	}
	g.labelsEnabled = previousEnabled
	return diagnostics, func() {
		// BY and non-leaf proofs need their actual graph constructors and allocation
		// order before the theorem owner can be retained.
		proof, proofComplete := g.completedTheoremProof(theorem.Syntax)
		if closeAPContext != nil {
			closeAPContext()
			g.symbols = previousSymbols
			if theorem.AssumeProveBody.semanticNode != nil {
				theorem.AssumeProveBody.semanticNode.inProof = false
			}
		}
		if complete && proofComplete {
			module := g.currentModule.semanticNode
			module.addTheorem(theorem.Syntax, body, proof, theorem.definitionNode)
			theorem.semanticNode = module.theoremVec[len(module.theoremVec)-1]
		}
	}
}

func (g *sanyExpressionGeneration) completedTheoremProof(syntax *SanySyntaxNode) (sanySemanticGraphNode, bool) {
	if syntax == nil {
		return nil, true
	}
	heirs := syntax.GetHeirs()
	if len(heirs) == 0 {
		return nil, true
	}
	proof := heirs[len(heirs)-1]
	if proof.Kind.JavaName() != "N_Proof" && proof.Kind.JavaName() != "N_TerminalProof" {
		return nil, true
	}
	if proof.Kind.JavaName() != "N_TerminalProof" {
		if node := g.structuredProofGraphs[proof]; node != nil {
			return node, true
		}
		return nil, false
	}
	tokens := proof.GetHeirs()
	index := 0
	if len(tokens) > 0 && tokens[0].Image == "PROOF" {
		index++
	}
	if index < len(tokens) && tokens[index].Image == "BY" {
		if node := g.leafProofGraphs[proof]; node != nil {
			return node, true
		}
		return nil, false
	}
	if index >= len(tokens) || tokens[index].Image != "OMITTED" && tokens[index].Image != "OBVIOUS" {
		return nil, false
	}
	return newSanySemLeafProofNode(proof, make([]sanySemanticGraphNode, 0), make([]sanySemSymbol, 0), tokens[index].Image == "OMITTED", false), true
}
