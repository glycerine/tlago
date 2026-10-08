// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlc

// ProcessConstantsDynamicExtendee is separate from constant operator
// pre-evaluation: it binds predefined operators and literal tool objects on
// the semantic graph, using the snapshot taken before ordinary definitions.
func (p *SpecProcessor) ProcessConstantsDynamicExtendee(module *ModuleNode) {
	p.ProcessConstants(module, p.PreConstantSnap.Snapshot())
}

func (p *SpecProcessor) ProcessConstants(node SemanticNode, definitions *Defns) {
	if node == nil {
		return
	}
	process := func(child SemanticNode) { p.ProcessConstants(child, definitions) }
	switch n := node.(type) {
	case *ModuleNode:
		for _, op := range n.GetOpDefs() {
			if replacement, ok := op.GetToolObjectAt(p.ToolID).(*OpDefNode); ok {
				p.ProcessedDefs.Set(replacement, struct{}{})
				process(replacement.Body)
			}
			process(op.Body)
		}
		for _, inner := range n.GetInnerModules() {
			process(inner)
		}
		for _, top := range n.TopLevel {
			process(top)
		}
	case *OpApplNode:
		if value := definitions.Get(n.Operator.Name); value != nil {
			if n.Operator.SemanticBase != nil {
				n.Operator.SetToolObjectAt(p.ToolID, value)
			} else {
				n.Operator.Data = value
			}
		} else {
			for _, argument := range n.Args {
				process(argument)
			}
			for _, bound := range n.BdedQuantBounds {
				process(bound)
			}
		}
	case *LetInNode:
		for _, op := range n.Lets {
			process(op.Body)
		}
		process(n.Body)
	case *SubstInNode:
		for _, substitution := range n.Substs {
			process(substitution.Expr)
		}
		process(n.Body)
	case *APSubstInNode:
		for _, substitution := range n.Substs {
			process(substitution.Expr)
		}
		process(n.Body)
	case *NumeralNode:
		if n.BigValue != nil {
			panic(NewTLCRuntimeException(ECTLCIntegerTooBig, SemanticString(n)))
		}
		n.SetToolObjectAt(p.ToolID, n.Value)
	case *DecimalNode:
		panic(NewTLCRuntimeException(ECTLCCantHandleRealNumbers, SemanticString(n)))
	case *StringNode:
		n.SetToolObjectAt(p.ToolID, n.Value)
	case *AssumeNode:
		process(n.Assume)
	case *OpArgNode:
		if op := n.Op.Definition; op != nil && n.Op.IsUserDefinedOp() {
			if _, processed := p.ProcessedDefs.Get2(op); !processed {
				p.ProcessedDefs.Set(op, struct{}{})
				process(op.Body)
			}
		}
	case *LabelNode:
		process(n.Body)
	}
}
