// Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

// processAssumption generates the expression before constructing a named
// definition, then adds a separate AssumeNode to the module's ordered vectors.
func (g *sanyExpressionGeneration) generateAssumptionExpression(assumption *NamedExpr, context map[string]Position) Diagnostics {
	if assumption.Syntax == nil || assumption.AssumeProve {
		return g.checkExpr(assumption.Expr, context, nil)
	}
	previousEnabled := g.labelsEnabled
	g.labelsEnabled = true
	defer func() { g.labelsEnabled = previousEnabled }()
	var finishLabels func() *sanyLabelTable
	if assumption.Name != "" {
		finishLabels = g.pushLabelScope()
	}
	diagnostics := g.checkExpr(assumption.Expr, context, nil)
	var labels *sanyLabelTable
	if finishLabels != nil {
		labels = finishLabels()
	}
	if sanyExpressionGenerationFailure(assumption.Expr) == sanyGenerationNullOperator {
		g.retainNullOperatorOperand(assumption.Expr, false)
	}
	source := sanyGenerationSource(assumption.Expr)
	if source == nil || source.semanticGraph == nil && sanyExpressionGenerationFailure(assumption.Expr) != sanyGenerationNullExpression {
		return diagnostics
	}
	module := g.currentModule.semanticNode
	if assumption.Name != "" {
		// An earlier native-only import has no canonical identity yet. Keep
		// this owner incomplete instead of replacing that unseen binding.
		if _, exists := context[assumption.Name]; exists && g.formalSymbolTable().resolveSymbol(assumption.Name) == nil {
			return diagnostics
		}
		definition, current := newSanySemCompletedThmOrAssumpDefNode(assumption.Name, false, source.semanticGraph, module, g.formalSymbolTable(), assumption.Syntax, nil, nil, nil)
		diagnostics = append(diagnostics, current...)
		definition.setLabels(labels)
		module.definitions = append(module.definitions, definition)
		assumption.definitionNode = definition
	}
	module.addAssumption(assumption.Syntax, source.semanticGraph, g.formalSymbolTable(), assumption.definitionNode)
	assumption.semanticNode = module.assumptionVec[len(module.assumptionVec)-1]
	return diagnostics
}
