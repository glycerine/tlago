package tlc

// ProcessConstantsDynamicExtendee is separate from constant operator
// pre-evaluation: it binds predefined operators and literal tool objects on
// the semantic graph, using the snapshot taken before ordinary definitions.
func (p *SpecProcessor) ProcessConstantsDynamicExtendee(module *ModuleNode) {
	p.require()
	p.ProcessConstants(module, p.PreConstantSnap.Snapshot())
}

func (p *SpecProcessor) ProcessConstants(node SemanticNode, definitions *Defns) {
	if isNil(node) {
		panic(NewNullPointerException())
	}
	process := func(child SemanticNode) { p.ProcessConstants(child, definitions) }
	switch n := node.(type) {
	case *ModuleNode:
		for _, op := range n.GetOpDefs() {
			if replacement, ok := op.GetToolObjectAt(p.ToolID).(*OpDefNode); ok && replacement != nil {
				p.ProcessedDefs.Set(replacement, struct{}{})
				process(replacement.GetBody())
			}
			process(op.GetBody())
		}
		for _, inner := range n.GetInnerModules() {
			process(inner)
		}
		for _, assume := range n.GetAssumptions() {
			process(assume)
		}
		for _, theorem := range n.GetTheorems() {
			process(theorem)
		}
	case *OpApplNode:
		if n.Operator == nil {
			panic(NewNullPointerException())
		}
		if value := definitions.Get(n.Operator.Name); !isNil(value) {
			if n.Operator.SemanticBase != nil {
				n.Operator.SetToolObjectAt(p.ToolID, value)
			} else {
				n.Operator.Data = value
			}
		} else {
			if n.Args == nil {
				panic(NewNullPointerException())
			}
			for _, argument := range n.Args {
				if !isNil(argument) {
					process(argument)
				}
			}
			if n.BdedQuantBounds == nil {
				panic(NewNullPointerException())
			}
			for _, bound := range n.BdedQuantBounds {
				process(bound)
			}
		}
	case *LetInNode:
		if n.Lets == nil {
			panic(NewNullPointerException())
		}
		for _, op := range n.Lets {
			if op == nil {
				panic(NewNullPointerException())
			}
			process(op.GetBody())
		}
		process(n.Body)
	case *SubstInNode:
		if n.Substs == nil {
			panic(NewNullPointerException())
		}
		for _, substitution := range n.Substs {
			process(substitution.GetExpr())
		}
		process(n.GetBody())
	case *APSubstInNode:
		if n.Substs == nil {
			panic(NewNullPointerException())
		}
		for _, substitution := range n.Substs {
			process(substitution.GetExpr())
		}
		process(n.GetBody())
	case *NumeralNode:
		value := NewIntValue(n.Value.ownedVal())
		if n.BigValue != nil {
			panic(NewTLCRuntimeException(ECTLCIntegerTooBig, SemanticString(n)))
		}
		n.SetToolObjectAt(p.ToolID, value)
	case *DecimalNode:
		panic(NewTLCRuntimeException(ECTLCCantHandleRealNumbers, SemanticString(n)))
	case *StringNode:
		n.SetToolObjectAt(p.ToolID, NewStringValueFromUnique(n.GetRep()))
	case *AssumeNode:
		process(n.Assume)
	case *TheoremNode:
		process(n.GetTheorem())
	case *OpArgNode:
		if n.Op == nil {
			panic(NewNullPointerException())
		}
		if op := n.Op.Definition; op != nil && n.Op.IsUserDefinedOp() {
			if _, processed := p.ProcessedDefs.Get2(op); !processed {
				p.ProcessedDefs.Set(op, struct{}{})
				process(op.GetBody())
			}
		}
	case *LabelNode:
		process(n.Body)
	}
}
