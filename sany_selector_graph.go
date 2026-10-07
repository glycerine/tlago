// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "strconv"

// Complete an already validated selector using its actual canonical graph.
// Arguments were generated in the caller's scope; selected bodies and formals
// retain their declaration identities. No selected AST body is regenerated.
func (g *sanyExpressionGeneration) retainCanonicalSubexpression(expr Expr, operatorArgument, symbolReferenceOnly, fact bool) Diagnostics {
	source, selected := sanyExprSource(expr), sanyExprSelection(expr)
	if source == nil || selected == nil || source.Selector == nil || symbolReferenceOnly {
		return nil
	}
	steps := source.Selector.Steps
	const (
		findingName = iota
		followingLabels
		findingOperand
	)
	mode, previous := findingName, -1
	inAPSuffices := false
	var current sanySemanticGraphNode
	var context *sanyContext
	var original sanySemSymbol
	var parameters []*sanyFormalParamNode
	var prefixes []sanySemanticGraphNode
	definitionArgs := make([]sanySemanticGraphNode, 0)
	var arguments []sanySemanticGraphNode
	argumentIndex := 0
	consume := func(step SanySelectorStep) ([]sanySemanticGraphNode, bool) {
		count := len(expressionChildren(step.Arguments))
		if operatorArgument {
			count = 0
		}
		if argumentIndex+count > len(selected.args) {
			return nil, false
		}
		args := make([]sanySemanticGraphNode, count)
		for i := range args {
			args[i] = sanyGeneratedExpressionNode(selected.args[argumentIndex+i])
			if args[i] == nil && sanyExpressionGenerationFailure(selected.args[argumentIndex+i]) != sanyGenerationNullExpression {
				return nil, false
			}
		}
		argumentIndex += count
		return args, true
	}
	parts := func(symbol sanySemSymbol) ([]*sanyFormalParamNode, sanySemanticGraphNode, bool) {
		switch node := symbol.(type) {
		case *sanySemOpDefNode:
			return node.formalNodes, node.body, true
		case *sanySemThmOrAssumpDefNode:
			return node.formalNodes, node.body, true
		}
		return nil, nil, false
	}
	label := func(node sanySemanticGraphNode, name string) *sanySemLabelNode {
		switch n := node.(type) {
		case *sanySemOpDefNode:
			return n.getLabel(name)
		case *sanySemThmOrAssumpDefNode:
			return n.getLabel(name)
		case *sanySemLabelNode:
			return n.getLabel(name)
		}
		return nil
	}
	unprefix := func(node sanySemanticGraphNode) sanySemanticGraphNode {
		for {
			switch n := node.(type) {
			case *sanySemSubstInNode:
				prefixes = append(prefixes, n)
				node = n.body
			case *sanySemAPSubstInNode:
				prefixes = append(prefixes, n)
				node = n.body
			default:
				return node
			}
		}
	}
	var diagnostics Diagnostics
	for index := 0; index < len(steps); index++ {
		step := steps[index]
		switch mode {
		case findingName:
			if definitionArgs == nil {
				definitionArgs = make([]sanySemanticGraphNode, 0)
			}
			name := ""
			var symbol sanySemSymbol
			for index < len(steps) && symbol == nil {
				step = steps[index]
				if step.Kind != SanySelectorName {
					return nil
				}
				if name != "" {
					name += "!"
				}
				name += sanyCanonicalOperatorImage(step.Name)
				args, ok := consume(step)
				if !ok {
					return nil
				}
				definitionArgs = append(definitionArgs, args...)
				if context == nil {
					symbol = g.formalSymbolTable().resolveSymbol(name)
				} else {
					symbol = context.getSymbol(name)
				}
				if symbol == nil {
					index++
				}
			}
			if symbol == nil {
				return nil
			}
			current = symbol.(sanySemanticGraphNode)
			if symbol.semKind() == sanyModuleInstanceKind {
				// Instance prefixes keep their accumulated arguments until the
				// full qualified operator name has been found.
				for index+1 < len(steps) && symbol.semKind() == sanyModuleInstanceKind {
					index++
					step = steps[index]
					if step.Kind != SanySelectorName {
						return nil
					}
					name += "!" + sanyCanonicalOperatorImage(step.Name)
					args, ok := consume(step)
					if !ok {
						return nil
					}
					definitionArgs = append(definitionArgs, args...)
					if context == nil {
						symbol = g.formalSymbolTable().resolveSymbol(name)
					} else {
						symbol = context.getSymbol(name)
					}
					if symbol == nil {
						return nil
					}
					current = symbol.(sanySemanticGraphNode)
				}
			}
			if original == nil {
				original = symbol
			}
			if index+1 < len(steps) {
				formals, body, ok := parts(symbol)
				if !ok {
					return nil
				}
				parameters = append(parameters, formals...)
				arguments = append(arguments, definitionArgs...)
				definitionArgs = nil
				body = unprefix(body)
				if steps[index+1].Kind == SanySelectorName {
					mode = followingLabels
				} else {
					mode = findingOperand
					current = body
				}
			}
			previous = findingName
		case followingLabels:
			node := label(current, step.Name)
			if node == nil {
				return nil
			}
			current = node
			args, ok := consume(step)
			if !ok {
				return nil
			}
			arguments = append(arguments, args...)
			parameters = append(parameters, node.formalNodes...)
			if index+1 < len(steps) && steps[index+1].Kind != SanySelectorName {
				mode = findingOperand
				current = node.body
			}
			previous = followingLabels
		case findingOperand:
			if step.Kind != SanySelectorColon {
				if current == nil {
					return nil
				}
				switch node := current.(type) {
				case *sanySemLabelNode:
					current = node.body
					index--
					continue
				case *sanySemLetInNode:
					if sanySelectorArgNumber(step, 1) != 1 {
						return nil
					}
					current = node.body
				case *sanySemOpApplNode:
					if node.operator == nil {
						return nil
					}
					name := node.operator.semName()
					boundCount := 0
					for _, group := range node.boundedBoundSymbols {
						boundCount += len(group)
					}
					switch {
					case node.operator.semKind() != sanyBuiltInKind:
						position := sanySelectorArgNumber(step, node.operator.semArity())
						if position < 1 || position > len(node.operands) {
							return nil
						}
						current = node.operands[position-1]
					case name == "$RcdConstructor" || name == "$SetOfRcds":
						position := sanySelectorArgNumber(step, len(node.operands))
						if position < 1 {
							return nil
						}
						current = node.operands[position-1].(*sanySemOpApplNode).operands[1]
					case name == "$Case":
						position := sanySelectorArgNumber(step, len(node.operands))
						if position < 1 || index+1 >= len(steps) {
							return nil
						}
						pair := node.operands[position-1].(*sanySemOpApplNode)
						index++
						position = sanySelectorArgNumber(steps[index], 2)
						if position < 1 {
							return nil
						}
						current = pair.operands[position-1]
					case name == "$Except":
						if sanySelectorArgNumber(step, len(node.operands)) != 1 {
							return nil
						}
						current = node.operands[0]
					case boundCount > 0 || len(node.unboundedBoundSymbols) > 0:
						if step.Kind == SanySelectorAt || step.Kind == SanySelectorNull {
							if boundCount > 0 {
								for _, group := range node.boundedBoundSymbols {
									parameters = append(parameters, group...)
								}
							} else {
								parameters = append(parameters, node.unboundedBoundSymbols...)
							}
							args, ok := consume(step)
							if !ok {
								return nil
							}
							arguments = append(arguments, args...)
							current = node.operands[0]
						} else {
							position := sanySelectorArgNumber(step, len(node.ranges))
							if position < 1 {
								return nil
							}
							current = node.ranges[position-1]
						}
					default:
						position := sanySelectorArgNumber(step, len(node.operands))
						if position < 1 {
							return nil
						}
						current = node.operands[position-1]
					}
				case *sanySemAssumeProveNode:
					if node.suffices && !inAPSuffices {
						if sanySelectorArgNumber(step, 1) != 1 {
							return nil
						}
						inAPSuffices = true
						break
					}
					inAPSuffices = false
					position := sanySelectorArgNumber(step, len(node.assumes)+1)
					if position < 1 {
						return nil
					}
					if position <= len(node.assumes) {
						current = node.assumes[position-1]
					} else {
						current = node.prove
					}
				case *sanySemOpArgNode:
					definition, ok := node.operator.(*sanySemOpDefNode)
					if !ok || definition.semName() != "LAMBDA" || (step.Kind != SanySelectorAt && step.Kind != SanySelectorNull) {
						return nil
					}
					parameters = append(parameters, definition.formalNodes...)
					args, ok := consume(step)
					if !ok {
						return nil
					}
					arguments = append(arguments, args...)
					current = definition.body
				default:
					return nil
				}
			}
			if current == nil {
				return nil
			}
			if index+1 < len(steps) && steps[index+1].Kind == SanySelectorName {
				for {
					n, ok := current.(*sanySemLabelNode)
					if !ok {
						break
					}
					current = n.body
				}
				let, ok := current.(*sanySemLetInNode)
				if !ok {
					return nil
				}
				context = let.context
				mode = findingName
			}
			previous = findingOperand
		}
	}
	if current == nil {
		return nil
	}
	if fact {
		switch current.(type) {
		case *sanySemAssumeProveNode, *sanySemNewSymbNode:
			source.semanticGraph = current
			return diagnostics
		}
	}
	if previous == findingName && len(parameters)+len(prefixes) > 0 {
		symbol, ok := current.(sanySemSymbol)
		if !ok {
			return nil
		}
		formals, _, ok := parts(symbol)
		if !ok {
			return nil
		}
		arguments = append(arguments, definitionArgs...)
		args := make([]sanySemanticGraphNode, len(formals))
		for i, param := range formals {
			syntax, _ := param.TreeNode.(*SanySyntaxNode)
			fresh := newSanyFormalParamNode(param.semName(), param.semArity(), param.pos, syntax, g.currentModule)
			parameters = append(parameters, fresh)
			if param.semArity() == 0 {
				node, ds, err := newSanySemOpApplNode(fresh, []sanySemanticGraphNode{}, source.Syntax)
				if err != nil {
					panic(err)
				}
				diagnostics = append(diagnostics, ds...)
				args[i] = node
			} else {
				args[i] = newSanySemOpArgNode(fresh, source.Syntax, g.currentModule.semanticNode)
			}
		}
		node, ds, err := newSanySemOpApplNode(symbol, args, source.Syntax)
		if err != nil {
			panic(err)
		}
		diagnostics = append(diagnostics, ds...)
		current = node
	}
	if symbol, ok := current.(sanySemSymbol); ok {
		if operatorArgument {
			source.semanticGraph = newSanySemOpArgNode(symbol, source.Syntax, g.currentModule.semanticNode)
			return diagnostics
		}
		node, ds, err := newSanySemOpApplNode(symbol, definitionArgs, source.Syntax)
		if err != nil {
			panic(err)
		}
		diagnostics = append(diagnostics, ds...)
		node.subExpressionOf = original
		source.semanticGraph = node
		return diagnostics
	}
	if operator, ok := current.(*sanySemOpArgNode); ok {
		if len(parameters)+len(prefixes) == 0 {
			source.semanticGraph = current
			return diagnostics
		}
		for i := 0; i < operator.arity; i++ {
			name := "NewParam" + strconv.Itoa(i)
			position := Position{File: "--TLA+ BUILTINS--"}
			syntax := &SanySyntaxNode{Image: name, FileName: position.File, Range: SanyRange{Begin: position, End: position}, Zero: []*SanySyntaxNode{}, One: []*SanySyntaxNode{}}
			parameters = append(parameters, newSanyFormalParamNode(name, 0, position, syntax, g.currentModule))
		}
		node, ds, err := newSanySemOpApplNode(operator.operator, definitionArgs, source.Syntax)
		if err != nil {
			panic(err)
		}
		diagnostics = append(diagnostics, ds...)
		current = node
	}
	for i := len(prefixes) - 1; i >= 0; i-- {
		switch prefix := prefixes[i].(type) {
		case *sanySemSubstInNode:
			node, ds := newSanySemSubstInNodeFromSubst(prefix, current)
			diagnostics = append(diagnostics, ds...)
			current = node
		case *sanySemAPSubstInNode:
			node, ds := newSanySemSubstInNodeFromAP(prefix, current)
			diagnostics = append(diagnostics, ds...)
			current = node
		}
	}
	var lambda *sanySemOpDefNode
	if len(parameters) > 0 {
		node, ds := newSanySemOpDefNode("LAMBDA", sanyUserDefinedOpKind, parameters, false, current, g.currentModule.semanticNode, nil, source.Syntax, true, nil)
		diagnostics = append(diagnostics, ds...)
		lambda = node
	}
	if operatorArgument {
		source.semanticGraph = newSanySemOpArgNode(lambda, source.Syntax, g.currentModule.semanticNode)
		return diagnostics
	}
	if len(parameters) != len(arguments) {
		return nil
	}
	if len(parameters) == 0 {
		node := newSanySemBuiltInOpApplNode("$Nop", []sanySemanticGraphNode{current}, source.Syntax)
		node.subExpressionOf = original
		source.semanticGraph = node
		return diagnostics
	}
	node, ds, err := newSanySemOpApplNode(lambda, arguments, source.Syntax)
	if err != nil {
		panic(err)
	}
	diagnostics = append(diagnostics, ds...)
	node.subExpressionOf = original
	source.semanticGraph = node
	return diagnostics
}
