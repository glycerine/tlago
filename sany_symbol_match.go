// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

func sanyMatchDiagnostic(application *sanySemOpApplNode, code, format string, parameters ...any) Diagnostic {
	var position Position
	var syntax *SanySyntaxNode
	if application != nil {
		syntax, _ = application.GetTreeNode().(*SanySyntaxNode)
		if syntax == nil && application.GetTreeNode() != nil {
			location := application.Location
			position = Position{File: location.Source, Line: location.BeginLine, Column: location.BeginColumn, EndLine: location.EndLine, EndColumn: location.EndColumn}
		}
	}
	if syntax != nil {
		position = sanyNodePosition(syntax)
	}
	diagnostic := errorAt(position, code, format, parameters...)
	diagnostic.SANYMessage = diagnostic.Message
	diagnostic.SANYParameters = parameters
	if syntax != nil {
		diagnostic.SANYRange = syntax.Range
	}
	return diagnostic
}

func (n *sanySemOpDeclNode) match(application *sanySemOpApplNode) (bool, Diagnostics, error) {
	if application.operands == nil || n.arity != len(application.operands) {
		return false, Diagnostics{sanyMatchDiagnostic(application, "E4004", "Operator used with the wrong number of arguments.")}, nil
	}
	return true, nil, nil
}

func (n *sanySemOpDefNode) matchingOpArgOperand(argument sanySemanticGraphNode, index int) bool {
	arg, ok := argument.(*sanySemOpArgNode)
	if !ok || arg == nil || n.formalNodes[index].semArity() != arg.arity {
		return false
	}
	if operator, ok := arg.operator.(*sanySemOpDefNode); ok {
		for i := 0; i < operator.arity; i++ {
			if operator.formalNodes[i].semArity() > 0 {
				return false
			}
		}
	}
	return true
}

func (n *sanySemOpDefNode) match(application *sanySemOpApplNode) (bool, Diagnostics, error) {
	arguments := application.operands
	correct := true
	var diagnostics Diagnostics
	add := func(code, format string, parameters ...any) {
		diagnostics = appendSanyDiagnostics(diagnostics, sanyMatchDiagnostic(application, code, format, parameters...))
	}
	abort := func(format string, parameters ...any) (bool, Diagnostics, error) {
		add("E4003", format, parameters...)
		return false, diagnostics, newSanySemanticAbort(diagnostics[len(diagnostics)-1], diagnostics, nil)
	}
	if n.semKind() == sanyModuleInstanceKind {
		add("E4004", "Module instance identifier where operator should be.")
		return false, diagnostics, nil
	}
	if n.arity == -1 {
		if arguments == nil {
			return abort("Internal error: null args vector for operator '%s' that should take variable number of args.", n.name)
		}
		for i, argument := range arguments {
			if arg, ok := argument.(*sanySemOpArgNode); ok && arg != nil {
				add("E4004", "Illegal expression used as argument %d to operator '%s'.", i+1, n.name)
				correct = false
			}
		}
		return correct, diagnostics, nil
	}
	if arguments == nil || n.formalNodes == nil {
		return abort("Internal error: Null args or params vector for operator '%s'.", n.name)
	}
	if len(n.formalNodes) != len(arguments) {
		add("E4004", "Wrong number of arguments (%d) given to operator '%s', \nwhich requires %d arguments.", len(arguments), n.name, len(n.formalNodes))
		return false, diagnostics, nil
	}
	switch n.semKind() {
	case sanyBuiltInKind:
		for i, argument := range arguments {
			if arg, ok := argument.(*sanySemOpArgNode); ok && arg != nil {
				add("E4004", "Non-expression used as argument number %d to BuiltIn operator '%s'.", i+1, n.name)
				correct = false
			}
		}
	case sanyUserDefinedOpKind:
		for i, parameter := range n.formalNodes {
			arity := parameter.semArity()
			if arity == 0 {
				if arg, ok := arguments[i].(*sanySemOpArgNode); ok && arg != nil {
					add("E4004", "Operator used in argument number %d has incorrect number of arguments.", i+1)
					correct = false
				}
			} else if arity > 0 {
				if !n.matchingOpArgOperand(arguments[i], i) {
					add("E4271", "Argument number %d to operator '%s' \nshould be a %d-parameter operator.", i+1, n.name, arity)
					correct = false
				}
			} else {
				// Source logs this internal error without changing the match boolean.
				add("E4003", "Internal error: Operator '%s' indicates that it requires \na negative number of arguments.", n.name)
			}
		}
	default:
		diagnostic := sanyMatchDiagnostic(nil, "E4003", "Internal error: operator neither BuiltIn nor UserDefined \nin call to OpDefNode.match()")
		diagnostics = appendSanyDiagnostics(diagnostics, diagnostic)
		return false, diagnostics, newSanySemanticAbort(diagnostic, diagnostics, nil)
	}
	return correct, diagnostics, nil
}
