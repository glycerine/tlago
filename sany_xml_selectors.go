// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import "fmt"

// Use the semantic selector's ordered formal parameters and selected source
// body for XML export and level checking, as for native TLC construction.
func (x *sanyXMLExporter) lambdaForResolvedSelector(e *CallExpr, selected *sanySelectorSelection, ctx sanyXMLExprContext) (*sanyXMLSymbol, Diagnostics) {
	if len(selected.params) == 0 {
		return nil, nil
	}
	key := fmt.Sprintf("selectorlambda:%s:%s:%d:%d:%d:%d", ctx.module.Name, selected.name, e.Pos.Line, e.Pos.Column, e.Pos.EndLine, e.Pos.EndColumn)
	if sym := x.lambdas[key]; sym != nil {
		return sym, nil
	}
	defSym := x.definitionSymbol(selected.name, ctx)
	source := selected.definition.module
	lambdaCtx := sanyXMLExprContext{
		module:          source,
		scope:           x.scopeForModule(source, map[string]bool{}),
		formals:         copySanyXMLSymbolMap(ctx.formals),
		defs:            map[string]*sanyXMLSymbol{},
		proofDefs:       ctx.proofDefs,
		suppressLetDefs: ctx.suppressLetDefs,
	}
	params := make([]*sanyXMLSymbol, len(selected.params))
	for i, bound := range selected.params {
		var param *sanyXMLSymbol
		if defSym != nil && i < len(defSym.Params) && defSym.Params[i].Name == bound.Name {
			param = defSym.Params[i]
		} else {
			param = x.newBoundFormal("expr", bound.Name, bound.Pos)
			param.Arity = bound.OperatorArity
		}
		param.LevelKnown = param.LevelKnown || exprReferencesName(selected.body, bound.Name, nil)
		x.emitFormalEntry(param)
		lambdaCtx.formals[bound.Name] = param
		params[i] = param
	}
	var lets []*LetExpr
	seen := map[*LetExpr]bool{}
	for _, let := range selected.lets {
		if !seen[let] {
			seen[let] = true
			lets = append(lets, let)
		}
	}
	var diags Diagnostics
	lambdaCtx, diags = x.applySelectedLetContexts(lambdaCtx, lets)
	if diags.HasErrors() {
		return nil, diags
	}
	sym := x.newSymbol("UserDefinedOpKind", key, "LAMBDA", len(params), constantLevel, e.Pos)
	sym.Params = params
	x.lambdas[key] = sym
	diags = x.emitLambdaEntry(sym, selected.body, lambdaCtx, nil, defSym, ctx.module)
	if diags.HasErrors() {
		return nil, diags
	}
	paramNames := make([]string, len(params))
	for i, param := range params {
		paramNames[i] = param.Name
	}
	x.setOperatorLevelData(sym, &Definition{Params: paramNames}, x.exprLevelData(selected.body, lambdaCtx, nil))
	return sym, nil
}
