// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import "github.com/glycerine/tlago/tlc"

type tlcBridgeInstanceKey struct {
	owner        *Module
	pos          Position
	name, module string
}

// Java's selected SubstIn copies retain the instance's existing Subst array
// and formal nodes. Module definitions and selections share their binding.
// Native selector wrappers still own separate substitution records until the
// canonical Subst adapter is implemented. Actual declaration formals are shared.
func (b *tlcBridge) instanceBinding(owner *Module, inst Instance) *tlcBridgeInstance {
	key := tlcBridgeInstanceKey{owner: owner, pos: inst.SourcePosition(), name: inst.Name, module: inst.Module}
	moduleInstance := false
	if owner != nil {
		for _, source := range owner.Instances {
			if source.SourcePosition() == key.pos && source.Name == key.name && source.Module == key.module {
				moduleInstance = true
				break
			}
		}
	}
	if !moduleInstance || positionIsZero(key.pos) {
		return &tlcBridgeInstance{owner: owner, inst: inst}
	}
	if b.instanceBindings == nil {
		b.instanceBindings = map[tlcBridgeInstanceKey]*tlcBridgeInstance{}
	}
	if binding := b.instanceBindings[key]; binding != nil {
		return binding
	}
	binding := &tlcBridgeInstance{owner: owner, inst: inst}
	b.instanceBindings[key] = binding
	return binding
}

// selectorNode is selectorToNode's final expression/operator construction.
// Source definitions and INSTANCE substitution nodes remain distinct; the
// selected body is compiled in its original module, with the lifted formals.
func (b *tlcBridge) selectorNode(expr Expr, selected *sanySelectorSelection) tlc.SemanticNode {
	var params []*tlc.SymbolNode
	var canonicalParameters []*sanyFormalParamNode
	switch node := sanyGeneratedExpressionNode(expr).(type) {
	case *sanySemOpApplNode:
		if definition, ok := node.operator.(*sanySemOpDefNode); ok && definition.semName() == "LAMBDA" {
			canonicalParameters = definition.formalNodes
		}
	case *sanySemOpArgNode:
		if definition, ok := node.operator.(*sanySemOpDefNode); ok {
			canonicalParameters = definition.formalNodes
		}
	}
	if len(canonicalParameters) == len(selected.params) {
		params = make([]*tlc.SymbolNode, len(canonicalParameters))
		for i, parameter := range canonicalParameters {
			params[i] = b.canonicalFormalParameter(parameter)
		}
	} else {
		params = b.boundParameters(selected.params, selected.definition.def.Syntax)
	}
	var bindings []*tlcBridgeInstance
	offset := 0
	for _, wrapper := range selected.definition.wrappers {
		if b.reusesInstanceSource(wrapper.inst, selected.definition.def) {
			continue
		}
		binding := b.instanceBinding(wrapper.owner, wrapper.inst)
		b.prepareInstanceBinding(binding)
		count := len(wrapper.inst.Params)
		if count > 0 {
			copy(params[offset:offset+count], binding.params)
		}
		offset += count
		bindings = append(bindings, binding)
	}
	previous := b.convertingModule
	b.convertingModule = selected.definition.module.Name
	restore := b.pushFormalParameters(params)
	bodyExpr := selected.body
	seen := map[*LetExpr]bool{}
	for i := len(selected.lets) - 1; i >= 0; i-- {
		original := selected.lets[i]
		if seen[original] {
			continue
		}
		seen[original] = true
		let := *original
		let.Body = bodyExpr
		// This native lexical wrapper has a different body from the actual
		// source LET. It must not claim that node's adapter identity.
		let.semanticGraph = nil
		bodyExpr = &let
	}
	body := b.convertExpr(bodyExpr)
	for i := len(bindings) - 1; i >= 0; i-- {
		binding := bindings[i]
		if len(binding.substs) > 0 {
			body = b.withPositionLocation(binding.inst.SourcePosition(), tlc.NewSubstInNode(body, binding.substs...))
		}
	}
	restore()
	b.convertingModule = previous
	if len(params) == 0 {
		return b.builtinNode(tlc.OpNop, body)
	}
	symbol := tlc.NewSymbolNode("LAMBDA")
	definition := tlc.NewOpDefNodeForSymbol(symbol, params, body)
	b.withPositionLocation(expr.Position(), definition)
	symbol.Data = definition
	if selected.operator {
		return tlc.NewOpArgNode(symbol)
	}
	args := make([]tlc.SemanticNode, len(selected.args))
	for i, arg := range selected.args {
		if selected.params[i].OperatorArity > 0 {
			if id, ok := arg.(*IdentExpr); ok && sanyExprSelection(arg) == nil {
				args[i] = b.withExprLocation(arg, tlc.NewOpArgNode(b.exprSymbol(id.Name)))
				continue
			}
		}
		args[i] = b.convertExpr(arg)
	}
	return tlc.NewOpApplNode(symbol, args...)
}
