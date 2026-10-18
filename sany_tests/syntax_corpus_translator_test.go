package sany_tests

// Faithful mechanical port of TlaPlusParserOutputTranslator.java's recursive
// translation switches. Constants below retain upstream numeric kind values.
import (
	"fmt"
	"github.com/glycerine/tlago"
	"regexp"
)

func syntaxCorpusEnumName(name string) string {
	for _, kind := range syntaxCorpusKinds {
		if kind.name == name {
			return kind.enum
		}
	}
	return name
}
func syntaxCorpusMatches(image, pattern string) bool {
	return regexp.MustCompile("^(?:" + pattern + ")$").MatchString(image)
}
func (p *syntaxCorpusReparser) translate(expected ...any) *syntaxCorpusAST {
	if len(expected) == 1 {
		if label, ok := expected[0].(string); ok {
			if p.isAtEnd() {
				panic(&syntaxCorpusDSLError{"EOF; expected " + label, p.current})
			}
			return syntaxCorpusTranslate(p.advance())
		}
	}
	kinds := make([]tlago.SanyNodeKind, len(expected))
	for i, kind := range expected {
		switch kind := kind.(type) {
		case int:
			kinds[i] = tlago.SanyNodeKind(kind)
		case tlago.SanyNodeKind:
			kinds[i] = kind
		default:
			panic("invalid translated source kind")
		}
	}
	return syntaxCorpusTranslate(p.consume(kinds...))
}
func (p *syntaxCorpusReparser) flatTranslate(parent *syntaxCorpusAST, kinds ...tlago.SanyNodeKind) {
	syntaxCorpusFlatTranslate(parent, p.consume(kinds...))
}
func syntaxCorpusCommaSeparatedNodes(p *syntaxCorpusReparser, label string) []*syntaxCorpusAST {
	var nodes []*syntaxCorpusAST
	for {
		nodes = append(nodes, p.translate(label))
		if !p.match(88) {
			break
		}
	}
	return nodes
}
func syntaxCorpusUnstructuredBound(p *syntaxCorpusReparser) *syntaxCorpusAST {
	bound := newSyntaxCorpusAST("quantifier_bound")
	if p.check(102) {
		bound.addChild(syntaxCorpusTupleOfIdentifiers(p))
	} else if p.check(289) {
		bound.addChildren(syntaxCorpusCommaSeparatedIDs(p))
	} else {
		panic(&syntaxCorpusDSLError{fmt.Sprintf("Failed to parse quantifier bound %d", p.peek().Kind), p.current})
	}
	p.consume(427)
	bound.addChild(newSyntaxCorpusAST("set_in"))
	bound.addField("set", p.translate("expression"))
	return bound
}
func syntaxCorpusBoundListOrIdentifierList(p *syntaxCorpusReparser) []*syntaxCorpusAST {
	var children []*syntaxCorpusAST
	lookahead := p.lookahead()
	failed := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(*syntaxCorpusDSLError); !ok {
					panic(r)
				}
				failed = true
			}
		}()
		for {
			children = append(children, syntaxCorpusUnstructuredBound(lookahead))
			if !lookahead.match(88) {
				break
			}
		}
		p.merge(lookahead)
	}()
	if failed {
		children = append(children, syntaxCorpusCommaSeparatedIDs(p)...)
	}
	return children
}
func syntaxCorpusUseBodyDefs(p *syntaxCorpusReparser, parent *syntaxCorpusAST) {
	for {
		if p.match(55) {
			ref := newSyntaxCorpusAST("module_ref")
			ref.addChild(p.translate(289))
			parent.addChild(ref)
		} else {
			parent.addChild(p.translate("operator or expression"))
		}
		if !p.match(88) {
			break
		}
	}
}
func syntaxCorpusUseBody(p *syntaxCorpusReparser) *syntaxCorpusAST {
	body := newSyntaxCorpusAST("use_body")
	if !p.check(66) {
		expr := newSyntaxCorpusAST("use_body_expr")
		syntaxCorpusUseBodyDefs(p, expr)
		body.addChild(expr)
	}
	if p.match(66) {
		defs := newSyntaxCorpusAST("use_body_def")
		syntaxCorpusUseBodyDefs(p, defs)
		body.addChild(defs)
	}
	return body
}

func syntaxCorpusFlatTranslate(parent *syntaxCorpusAST, node *tlago.SanySyntaxNode) {
	parser := newSyntaxCorpusReparser(node.GetHeirs())
	switch node.Kind {
	case 333:
		parser.consume(35, 3)
		parent.addChild(newSyntaxCorpusAST("header_line"))
		parser.consume(289)
		parent.addField("name", newSyntaxCorpusAST("identifier"))
		parser.consume(36)
		parent.addChild(newSyntaxCorpusAST("header_line"))
		break
	case 350:
		if parser.match(48) {
			extensions := newSyntaxCorpusAST("extends")
			for {
				extensions.addChild(parser.translate(289))
				if !(parser.match(88)) {
					break
				}
			}
			parent.addChild(extensions)
		}
		break
	case 334:
		for !parser.isAtEnd() {
			parent.addChild(parser.translate("unit"))
		}
		break
	case 366:
		parent.addField("name", parser.translate(289).alias("identifier_ref", "identifier"))
		if parser.match(94) {
			for {
				parent.addField("parameter", parser.translate(363, 398, 370, 394))
				if !(parser.match(88)) {
					break
				}
			}
			parser.consume(95)
		}
		break
	case 400:
		op := newSyntaxCorpusAST("prefix_op_symbol")
		op.addChild(syntaxCorpusPrefixOpFromString(parser.advance().Image))
		parent.addField("name", op)
		parser.consume(289)
		parent.addField("parameter", newSyntaxCorpusAST("identifier"))
		break
	case 372:
		op := newSyntaxCorpusAST("infix_op_symbol")
		parser.consume(289)
		parent.addField("parameter", newSyntaxCorpusAST("identifier"))
		op.addChild(syntaxCorpusInfixOpFromString(parser.advance().Image))
		parent.addField("name", op)
		parser.consume(289)
		parent.addField("parameter", newSyntaxCorpusAST("identifier"))
		break
	case 396:
		op := newSyntaxCorpusAST("postfix_op_symbol")
		parser.consume(289)
		parent.addField("parameter", newSyntaxCorpusAST("identifier"))
		op.addChild(syntaxCorpusPostfixOpFromString(parser.advance().Image))
		parent.addField("name", op)
		break
	case 388:
		parser.consume(94)
		for {
			parent.addField("parameter", parser.translate("expression"))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(95)
		break
	case 379:
		for {
			parent.addField("definitions", parser.translate("unit definition"))
			if !(!parser.isAtEnd()) {
				break
			}
		}
		fallthrough
	case 381:
		if !parser.isAtEnd() {
			parser.consume(427)
			parent.addChild(newSyntaxCorpusAST("set_in"))
			parent.addField("set", parser.translate("expression"))
		}
		break
	case 355:
		parser.consume(289)
		parent.addChild(newSyntaxCorpusAST("identifier"))
		parent.addChild(parser.translate(108))
		parent.addChild(parser.translate("expression"))
		break
	case 354:
		parser.consume(289)
		parent.addChild(newSyntaxCorpusAST("identifier"))
		parser.consume(89)
		parent.addChild(parser.translate("expression"))
		break
	case 406:
		qedStep := newSyntaxCorpusAST("qed_step")
		qedStep.addChild(parser.translate(290, 291, 292, 293).alias("proof_step_ref", "proof_step_id"))
		parser.consume(407)
		if !parser.isAtEnd() {
			qedStep.addChild(parser.translate("proof"))
		}
		parent.addChild(qedStep)
		break
	default:
		panic(&syntaxCorpusDSLError{fmt.Sprintf("Unhandled conversion from kind %d image %s", node.Kind, node.Image), 0})
	}
	syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
}
func syntaxCorpusTranslate(node *tlago.SanySyntaxNode) *syntaxCorpusAST {
	parser := newSyntaxCorpusReparser(node.GetHeirs())
	switch node.Kind {
	case 382:
		module := newSyntaxCorpusAST("module")
		parser.flatTranslate(module, 333)
		parser.flatTranslate(module, 350)
		parser.flatTranslate(module, 334)
		module.addChild(parser.translate(345))
		return module
	case 345:
		parser.consume(37)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("double_line")
	case 36:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("single_line")
	case 392:
		constants := newSyntaxCorpusAST("constant_declaration")
		parser.consume(342)
		for {
			constants.addChild(parser.translate(363, 398, 370, 394))
			if !(parser.match(88)) {
				break
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return constants
	case 426:
		variables := newSyntaxCorpusAST("variable_declaration")
		parser.consume(85)
		variables.addChildren(syntaxCorpusCommaSeparatedIDs(parser))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return variables
	case 375:
		if parser.match(54) {
			localDefn := newSyntaxCorpusAST("local_definition")
			localDefn.addChild(parser.translate(376))
			syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
			return localDefn
		} else {
			instance := parser.translate(376)
			syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
			return instance
		}
	case 376:
		instance := newSyntaxCorpusAST("instance")
		parser.consume(51)
		instance.addChild(parser.translate(289))
		if parser.match(87) {
			for {
				instance.addChild(parser.translate(420))
				if !(parser.match(88)) {
					break
				}
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return instance
	case 383:
		moduleDefinition := newSyntaxCorpusAST("module_definition")
		parent := moduleDefinition
		if parser.match(54) {
			localDefn := newSyntaxCorpusAST("local_definition")
			localDefn.addChild(moduleDefinition)
			parent = localDefn
		}
		parser.flatTranslate(moduleDefinition, 366)
		moduleDefinition.addChild(parser.translate(93))
		moduleDefinition.addField("definition", parser.translate(376))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return parent
	case 420:
		substitution := newSyntaxCorpusAST("substitution")
		substitution.addChild(parser.translate(289, 384, 373, 397))
		substitution.addChild(parser.translate(107))
		substitution.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return substitution
	case 107:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("gets")
	case 389:
		operatorDefinition := newSyntaxCorpusAST("operator_definition")
		parent := operatorDefinition
		if parser.match(54) {
			localDefn := newSyntaxCorpusAST("local_definition")
			localDefn.addChild(operatorDefinition)
			parent = localDefn
		}
		parser.flatTranslate(operatorDefinition, 366, 400, 372, 396)
		operatorDefinition.addChild(parser.translate(93))
		operatorDefinition.addField("definition", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return parent
	case 363:
		parser.consume(289)
		if parser.isAtEnd() {
			return newSyntaxCorpusAST("identifier")
		}
		op := newSyntaxCorpusAST("operator_declaration")
		op.addField("name", newSyntaxCorpusAST("identifier"))
		parser.consume(94)
		for {
			op.addField("parameter", parser.translate(92))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(95)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op
	case 356:
		fn := newSyntaxCorpusAST("function_definition")
		parent := fn
		if parser.match(54) {
			localDefn := newSyntaxCorpusAST("local_definition")
			localDefn.addChild(fn)
			parent = localDefn
		}
		parser.consume(289)
		fn.addField("name", newSyntaxCorpusAST("identifier"))
		parser.consume(97)
		for {
			fn.addChild(parser.translate(408))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(99)
		fn.addChild(parser.translate(93))
		fn.addField("definition", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return parent
	case 431:
		recursiveDeclaration := newSyntaxCorpusAST("recursive_declaration")
		parser.consume(79)
		recursiveDeclaration.addChildren(syntaxCorpusCommaSeparatedNodes(parser, "operator declaration"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return recursiveDeclaration
	case 92:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("placeholder")
	case 93:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("def_eq")
	case 341:
		conjList := newSyntaxCorpusAST("conj_list")
		for {
			conjList.addChild(parser.translate(340))
			if !(!parser.isAtEnd()) {
				break
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return conjList
	case 340:
		conjItem := newSyntaxCorpusAST("conj_item")
		conjItem.addChild(parser.translate(129))
		conjItem.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return conjItem
	case 129:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("bullet_conj")
	case 344:
		disjList := newSyntaxCorpusAST("disj_list")
		for {
			disjList.addChild(parser.translate(343))
			if !(!parser.isAtEnd()) {
				break
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return disjList
	case 343:
		disjItem := newSyntaxCorpusAST("disj_item")
		disjItem.addChild(parser.translate(132))
		disjItem.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return disjItem
	case 132:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("bullet_disj")
	case 358:
		prefix := parser.consume(367)
		if parser.match(433) {
			subexpr := newSyntaxCorpusAST("subexpression")
			subexpr.addChild(syntaxCorpusTranslate(prefix))
			subexpr.addChild(newSyntaxCorpusAST("subexpr_tree_nav").addChild(syntaxCorpusTranslate(parser.previous())))
			syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
			return subexpr
		} else {
			parser.consume(289, 384, 373, 397, 290, 291)
		}
		op := syntaxCorpusTranslate(parser.previous())
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		if 0 == len(prefix.GetHeirs()) {
			return op
		}
		return newSyntaxCorpusAST("prefixed_op").addField("prefix", syntaxCorpusTranslate(prefix)).addField("op", op)
	case 367:
		subexpr := newSyntaxCorpusAST("subexpr_prefix")
		for {
			subexpr.addChild(parser.translate(368))
			if !(!parser.isAtEnd()) {
				break
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return subexpr
	case 368:
		var component *syntaxCorpusAST
		if parser.match(289) {
			component = newSyntaxCorpusAST("subexpr_component")
			if parser.match(388) {
				op := newSyntaxCorpusAST("bound_op")
				op.addField("name", newSyntaxCorpusAST("identifier_ref"))
				syntaxCorpusFlatTranslate(op, parser.previous())
				component.addChild(op)
			} else {
				component.addChild(newSyntaxCorpusAST("identifier_ref"))
			}
		} else if parser.match(388) {
			component = newSyntaxCorpusAST("subexpr_tree_nav")
			args := newSyntaxCorpusAST("operator_args")
			argParser := newSyntaxCorpusReparser(parser.previous().GetHeirs())
			argParser.consume(94)
			for {
				args.addChild(argParser.translate("op or expression"))
				if !(argParser.match(88)) {
					break
				}
			}
			argParser.consume(95)
			syntaxCorpusAssert(argParser.isAtEnd(), "source assertTrue")
			component.addChild(args)
		} else if parser.match(433) {
			component = newSyntaxCorpusAST("subexpr_tree_nav")
			component.addChild(syntaxCorpusTranslate(parser.previous()))
		} else {
			component = parser.translate(290, 291)
		}
		parser.consume(105)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return component
	case 433:
		if parser.match(385) {
			syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
			return newSyntaxCorpusAST("child_id")
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		image := node.Image
		switch image {
		case "<<":
			return newSyntaxCorpusAST("langle_bracket")
		case ">>":
			return newSyntaxCorpusAST("rangle_bracket")
		case ":":
			return newSyntaxCorpusAST("colon")
		case "@":
			return newSyntaxCorpusAST("address")
		default:
			panic(&syntaxCorpusDSLError{fmt.Sprintf("Unknown subexpression tree nav symbol %s", image), parser.current})
		}
	case 289:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return syntaxCorpusIdentifier(node)
	case 385:
		image := parser.consume(109).Image
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		if syntaxCorpusMatches(image, "\\d+") {
			return newSyntaxCorpusAST("nat_number")
		} else if syntaxCorpusMatches(image, "\\\\[bB][0|1]+") {
			return newSyntaxCorpusAST("binary_number").addChild(newSyntaxCorpusAST("format")).addChild(newSyntaxCorpusAST("value"))
		} else if syntaxCorpusMatches(image, "\\\\[oO][0-7]+") {
			return newSyntaxCorpusAST("octal_number").addChild(newSyntaxCorpusAST("format")).addChild(newSyntaxCorpusAST("value"))
		} else if syntaxCorpusMatches(image, "\\\\[hH][0-9a-fA-F]+") {
			return newSyntaxCorpusAST("hex_number").addChild(newSyntaxCorpusAST("format")).addChild(newSyntaxCorpusAST("value"))
		} else {
			panic(&syntaxCorpusDSLError{fmt.Sprintf("Invalid number literal format %s", image), 0})
		}
	case 364:
		parser.consume(109)
		parser.consume(91)
		parser.consume(109)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("real_number")
	case 418:
		string := newSyntaxCorpusAST("string")
		codepoints := []rune(node.Image)
		for i := 0; i < len(codepoints); i++ {
			c := codepoints[i]
			if 0 == i || len(codepoints)-1 == i {
				continue
			}
			if c == '\\' || c == '\'' || c == '"' || c == '\n' || c == '\r' || c == '\t' || c == '\f' {
				string.addChild(newSyntaxCorpusAST("escape_char"))
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return string
	case 423:
		tuple := newSyntaxCorpusAST("tuple_literal")
		tuple.addChild(parser.translate(102))
		if !parser.check(104) {
			tuple.addChildren(syntaxCorpusCommaSeparatedNodes(parser, "expression"))
		}
		tuple.addChild(parser.translate(104))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return tuple
	case 102:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("langle_bracket")
	case 104:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("rangle_bracket")
	case 103:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("rangle_bracket_sub")
	case 411:
		setLiteral := newSyntaxCorpusAST("finite_set_literal")
		parser.consume(100)
		if !parser.check(101) {
			setLiteral.addChildren(syntaxCorpusCommaSeparatedNodes(parser, "expression"))
		}
		parser.consume(101)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return setLiteral
	case 419:
		setFilter := newSyntaxCorpusAST("set_filter")
		parser.consume(100)
		bound := newSyntaxCorpusAST("quantifier_bound")
		if parser.match(289) {
			bound.addField("intro", newSyntaxCorpusAST("identifier"))
		} else {
			bound.addField("intro", parser.translate(365))
		}
		parser.consume(427)
		bound.addChild(newSyntaxCorpusAST("set_in"))
		bound.addField("set", parser.translate("expression"))
		setFilter.addField("generator", bound)
		parser.consume(89)
		setFilter.addField("filter", parser.translate("expression"))
		parser.consume(101)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return setFilter
	case 413:
		setMap := newSyntaxCorpusAST("set_map")
		parser.consume(100)
		setMap.addField("map", parser.translate("expression"))
		parser.consume(89)
		for {
			setMap.addField("generator", parser.translate(408))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(101)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return setMap
	case 335:
		quant := newSyntaxCorpusAST("bounded_quantification")
		quant.addField("quantifier", parser.translate(47, 49))
		for {
			quant.addField("bound", parser.translate(408))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(89)
		quant.addField("expression", parser.translate("expression"))
		return quant
	case 425:
		quant := newSyntaxCorpusAST("unbounded_quantification")
		quant.addField("quantifier", parser.translate(47, 49, 61, 60))
		for {
			parser.consume(289)
			quant.addField("intro", newSyntaxCorpusAST("identifier"))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(89)
		quant.addField("expression", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return quant
	case 424:
		choose := newSyntaxCorpusAST("choose")
		parser.consume(43)
		if parser.match(289) {
			choose.addField("intro", newSyntaxCorpusAST("identifier"))
		} else {
			choose.addField("intro", parser.translate(365))
		}
		parser.flatTranslate(choose, 381)
		parser.consume(89)
		choose.addField("expression", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return choose
	case 47:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("exists")
	case 49:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("forall")
	case 61:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("temporal_forall")
	case 60:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("temporal_exists")
	case 408:
		quantBound := newSyntaxCorpusAST("quantifier_bound")
		if parser.check(289) {
			for {
				parser.consume(289)
				quantBound.addField("intro", newSyntaxCorpusAST("identifier"))
				if !(parser.match(88)) {
					break
				}
			}
		} else {
			quantBound.addField("intro", parser.translate(365))
		}
		parser.consume(427)
		quantBound.addChild(newSyntaxCorpusAST("set_in"))
		quantBound.addField("set", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return quantBound
	case 365:
		tuple := syntaxCorpusTupleOfIdentifiers(parser)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return tuple
	case 353:
		function := newSyntaxCorpusAST("function_literal")
		parser.consume(97)
		for {
			function.addChild(parser.translate(408))
			if !(parser.match(88)) {
				break
			}
		}
		function.addChild(parser.translate(108))
		function.addChild(parser.translate("expression"))
		parser.consume(99)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return function
	case 108:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("all_map_to")
	case 414:
		setOfFunctions := newSyntaxCorpusAST("set_of_functions")
		parser.consume(97)
		setOfFunctions.addChild(parser.translate("expression"))
		parser.consume(106)
		setOfFunctions.addChild(newSyntaxCorpusAST("maps_to"))
		setOfFunctions.addChild(parser.translate("expression"))
		parser.consume(99)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return setOfFunctions
	case 346:
		except := newSyntaxCorpusAST("except")
		parser.consume(97)
		except.addField("expr_to_update", parser.translate("expression"))
		parser.consume(46)
		for {
			except.addChild(parser.translate(348))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(99)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return except
	case 348:
		update := newSyntaxCorpusAST("except_update")
		parser.consume(105)
		updateSpec := newSyntaxCorpusAST("except_update_specifier")
		for {
			updateSpec.addChild(parser.translate(347))
			if !(!parser.match(428)) {
				break
			}
		}
		update.addField("update_specifier", updateSpec)
		update.addField("new_val", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return update
	case 347:
		if parser.match(97) {
			update := newSyntaxCorpusAST("except_update_fn_appl")
			update.addChildren(syntaxCorpusCommaSeparatedNodes(parser, "expression"))
			parser.consume(99)
			syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
			return update
		} else {
			update := newSyntaxCorpusAST("except_update_record_field")
			parser.consume(91)
			update.addChild(parser.translate(289))
			syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
			return update
		}
	case 409:
		record := newSyntaxCorpusAST("record_literal")
		parser.consume(97)
		for {
			parser.flatTranslate(record, 355)
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(99)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return record
	case 415:
		record := newSyntaxCorpusAST("set_of_records")
		parser.consume(97)
		for {
			parser.flatTranslate(record, 354)
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(99)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return record
	case 410:
		record := newSyntaxCorpusAST("record_value")
		record.addChild(parser.translate("expression"))
		parser.consume(91)
		record.addChild(parser.translate(289))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return record
	case 369:
		ite := newSyntaxCorpusAST("if_then_else")
		parser.consume(50)
		ite.addField("if", parser.translate("expression"))
		parser.consume(62)
		ite.addField("then", parser.translate("expression"))
		parser.consume(45)
		ite.addField("else", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return ite
	case 336:
		cases := newSyntaxCorpusAST("case")
		parser.consume(42)
		cases.addChild(parser.translate(337))
		for parser.match(121) {
			cases.addChild(newSyntaxCorpusAST("case_box"))
			cases.addChild(parser.translate(337, 390))
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return cases
	case 337:
		caseArm := newSyntaxCorpusAST("case_arm")
		caseArm.addChild(parser.translate("expression"))
		parser.consume(106)
		caseArm.addChild(newSyntaxCorpusAST("case_arrow"))
		caseArm.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return caseArm
	case 390:
		otherArm := newSyntaxCorpusAST("other_arm")
		parser.consume(57)
		parser.consume(106)
		otherArm.addChild(newSyntaxCorpusAST("case_arrow"))
		otherArm.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return otherArm
	case 329:
		var actionExpr *syntaxCorpusAST
		if parser.match(97) {
			actionExpr = newSyntaxCorpusAST("step_expr_or_stutter")
			actionExpr.addChild(parser.translate("expression"))
			parser.consume(98)
			actionExpr.addChild(parser.translate("subscript expression"))
		} else {
			actionExpr = newSyntaxCorpusAST("step_expr_no_stutter")
			actionExpr.addChild(parser.translate(102))
			actionExpr.addChild(parser.translate("expression"))
			actionExpr.addChild(parser.translate(103))
			actionExpr.addChild(parser.translate("subscript expression"))
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return actionExpr
	case 351:
		fairness := newSyntaxCorpusAST("fairness")
		parser.consume(86, 59)
		fairness.addChild(parser.translate("subscript expression"))
		parser.consume(94)
		fairness.addChild(parser.translate("expression"))
		parser.consume(95)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return fairness
	case 380:
		letIn := newSyntaxCorpusAST("let_in")
		parser.consume(52)
		parser.flatTranslate(letIn, 379)
		parser.consume(53)
		letIn.addField("expression", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return letIn
	case 393:
		paren := newSyntaxCorpusAST("parentheses")
		parser.consume(94)
		paren.addChild(parser.translate("expression"))
		parser.consume(95)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return paren
	case 349:
		exprs := []*syntaxCorpusAST{}
		for {
			exprs = append(exprs, parser.translate("expression"))
			if !(parser.match(359)) {
				break
			}
		}
		lhs := exprs[0]
		for i := 1; i < len(exprs); i++ {
			op := newSyntaxCorpusAST("bound_infix_op")
			op.addField("lhs", lhs)
			op.addField("symbol", newSyntaxCorpusAST("times"))
			op.addField("rhs", exprs[i])
			lhs = op
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return lhs
	case 352:
		functionEvaluation := newSyntaxCorpusAST("function_evaluation")
		functionEvaluation.addChild(parser.translate("expression"))
		parser.consume(97)
		functionEvaluation.addChildren(syntaxCorpusCommaSeparatedNodes(parser, "expression"))
		parser.consume(99)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return functionEvaluation
	case 387:
		id := parser.consume(358)
		idParser := newSyntaxCorpusReparser(id.GetHeirs())
		prefix := idParser.consume(367)
		nameOrSymbol := idParser.translate(289, 384, 373, 397)
		syntaxCorpusAssert(idParser.isAtEnd(), "source assertTrue")
		var op *syntaxCorpusAST
		switch nameOrSymbol.kind {
		case "identifier_ref":
			op = newSyntaxCorpusAST("bound_op")
			op.addField("name", nameOrSymbol)
			break
		case "prefix_op_symbol", "infix_op_symbol", "postfix_op_symbol":
			op = newSyntaxCorpusAST("bound_nonfix_op")
			op.addField("symbol", nameOrSymbol)
			break
		default:
			panic(&syntaxCorpusDSLError{fmt.Sprintf("Unhandled op case %s", syntaxCorpusEnumName(nameOrSymbol.kind)), parser.current})
		}
		parser.flatTranslate(op, 388)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		if len(prefix.GetHeirs()) > 0 {
			prefixedOp := newSyntaxCorpusAST("prefixed_op")
			prefixedOp.addField("prefix", syntaxCorpusTranslate(prefix))
			prefixedOp.addField("op", op)
			return prefixedOp
		} else {
			return op
		}
	case 398:
		op := newSyntaxCorpusAST("operator_declaration")
		op.addField("name", parser.translate(384))
		op.addChild(parser.translate(92))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op
	case 399:
		boundPrefixOp := newSyntaxCorpusAST("bound_prefix_op")
		if parser.match(359) {
			boundPrefixOp.addField("symbol", newSyntaxCorpusAST("negative"))
		} else {
			boundPrefixOp.addField("symbol", parser.translate(362))
		}
		boundPrefixOp.addField("rhs", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return boundPrefixOp
	case 362:
		prefix := parser.consume(367)
		syntaxCorpusAssert(0 == len(prefix.GetHeirs()), "source assertEquals")
		op := parser.consume(401)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return syntaxCorpusPrefixOpFromString(op.Image)
	case 360:
		prefix := parser.consume(367)
		syntaxCorpusAssert(0 == len(prefix.GetHeirs()), "source assertEquals")
		op := parser.translate(384)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op
	case 384:
		op := newSyntaxCorpusAST("prefix_op_symbol")
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op.addChild(syntaxCorpusPrefixOpFromString(node.Image))
	case 370:
		op := newSyntaxCorpusAST("operator_declaration")
		op.addChild(parser.translate(92))
		op.addField("name", parser.translate(373))
		op.addChild(parser.translate(92))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op
	case 371:
		boundInfixOp := newSyntaxCorpusAST("bound_infix_op")
		boundInfixOp.addField("lhs", parser.translate("expression"))
		genOp := parser.consume(359)
		genOpParser := newSyntaxCorpusReparser(genOp.GetHeirs())
		prefix := genOpParser.consume(367)
		syntaxCorpusAssert(0 == len(prefix.GetHeirs()), "source assertEquals")
		op := genOpParser.consume(373, 427)
		syntaxCorpusAssert(genOpParser.isAtEnd(), "source assertTrue")
		boundInfixOp.addField("symbol", syntaxCorpusInfixOpFromString(op.Image))
		boundInfixOp.addField("rhs", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return boundInfixOp
	case 359:
		prefix := parser.consume(367)
		syntaxCorpusAssert(0 == len(prefix.GetHeirs()), "source assertEquals")
		op := parser.translate(373, 427)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op
	case 373:
		op := newSyntaxCorpusAST("infix_op_symbol")
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op.addChild(syntaxCorpusInfixOpFromString(node.Image))
	case 394:
		op := newSyntaxCorpusAST("operator_declaration")
		op.addChild(parser.translate(92))
		op.addField("name", parser.translate(397))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op
	case 395:
		boundPostfixOp := newSyntaxCorpusAST("bound_postfix_op")
		boundPostfixOp.addField("lhs", parser.translate("expression"))
		genOp := parser.consume(361)
		genOpParser := newSyntaxCorpusReparser(genOp.GetHeirs())
		prefix := genOpParser.consume(367)
		syntaxCorpusAssert(0 == len(prefix.GetHeirs()), "source assertEquals")
		op := genOpParser.consume(397)
		syntaxCorpusAssert(genOpParser.isAtEnd(), "source assertTrue")
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		boundPostfixOp.addField("symbol", syntaxCorpusPostfixOpFromString(op.Image))
		return boundPostfixOp
	case 361:
		prefix := parser.consume(367)
		syntaxCorpusAssert(0 == len(prefix.GetHeirs()), "source assertEquals")
		op := parser.translate(397)
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op
	case 397:
		op := newSyntaxCorpusAST("postfix_op_symbol")
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return op.addChild(syntaxCorpusPostfixOpFromString(node.Image))
	case 427:
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newSyntaxCorpusAST("in")
	case 332:
		assumption := newSyntaxCorpusAST("assumption")
		parser.consume(39, 41)
		if parser.match(289) {
			assumption.addField("name", newSyntaxCorpusAST("identifier"))
			parser.consume(93)
			assumption.addChild(newSyntaxCorpusAST("def_eq"))
		}
		assumption.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return assumption
	case 430:
		lambda := newSyntaxCorpusAST("lambda")
		parser.consume(73)
		for {
			parser.consume(289)
			lambda.addChild(newSyntaxCorpusAST("identifier"))
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(89)
		lambda.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return lambda
	case 422:
		theorem := newSyntaxCorpusAST("theorem")
		parser.consume(67, 58)
		if parser.match(289) {
			theorem.addField("name", newSyntaxCorpusAST("identifier"))
			parser.consume(93)
			theorem.addChild(newSyntaxCorpusAST("def_eq"))
		}
		theorem.addField("statement", parser.translate("expression or assume/prove"))
		if !parser.isAtEnd() {
			theorem.addField("proof", parser.translate("proof"))
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return theorem
	case 435:
		proof := newSyntaxCorpusAST("terminal_proof")
		parser.match(75)
		if parser.match(71, 72) {
			return proof
		}
		parser.consume(63)
		parser.match(64)
		proof.addChild(syntaxCorpusUseBody(parser))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 402:
		proof := newSyntaxCorpusAST("non_terminal_proof")
		parser.match(75)
		for parser.match(406) {
			if parser.isAtEnd() {
				syntaxCorpusFlatTranslate(proof, parser.previous())
			} else {
				proof.addChild(syntaxCorpusTranslate(parser.previous()))
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 406:
		proofStep := newSyntaxCorpusAST("proof_step")
		proofStep.addChild(parser.translate(290, 291, 292, 293).alias("proof_step_ref", "proof_step_id"))
		proofStepStatement := parser.translate("proof step statement")
		proofStep.addChild(proofStepStatement)
		if !parser.isAtEnd() {
			proofStepStatement.addChild(parser.translate("proof"))
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proofStep
	case 290, 291, 292, 293:
		proofStepId := newSyntaxCorpusAST("proof_step_ref")
		proofStepId.addChild(newSyntaxCorpusAST("level"))
		proofStepId.addChild(newSyntaxCorpusAST("name"))
		return proofStepId
	case 438:
		proof := newSyntaxCorpusAST("definition_proof_step")
		parser.match(65)
		for {
			proof.addChild(parser.translate("expression"))
			if !(!parser.isAtEnd()) {
				break
			}
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 439:
		proof := newSyntaxCorpusAST("have_proof_step")
		parser.consume(70)
		proof.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 441:
		proof := newSyntaxCorpusAST("witness_proof_step")
		parser.consume(83)
		proof.addChildren(syntaxCorpusCommaSeparatedNodes(parser, "definition or expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 440:
		proof := newSyntaxCorpusAST("take_proof_step")
		parser.consume(74)
		proof.addChildren(syntaxCorpusBoundListOrIdentifierList(parser))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 444:
		proof := newSyntaxCorpusAST("suffices_proof_step")
		parser.match(84)
		proof.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 443:
		proof := newSyntaxCorpusAST("case_proof_step")
		parser.consume(42)
		proof.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 442:
		proof := newSyntaxCorpusAST("pick_proof_step")
		parser.consume(82)
		proof.addChildren(syntaxCorpusBoundListOrIdentifierList(parser))
		parser.consume(89)
		proof.addChild(parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 436:
		proof := newSyntaxCorpusAST("use_or_hide")
		parser.consume(68, 69)
		proof.addChild(syntaxCorpusUseBody(parser))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return proof
	case 331:
		assumeProve := newSyntaxCorpusAST("assume_prove")
		parser.consume(39)
		for {
			if parser.match(432) {
				inner := newSyntaxCorpusAST("inner_assume_prove")
				innerParser := newSyntaxCorpusReparser(parser.previous().GetHeirs())
				innerParser.consume(358)
				inner.addChild(newSyntaxCorpusAST("identifier"))
				inner.addChild(innerParser.translate(90))
				inner.addChild(innerParser.translate(331))
				syntaxCorpusAssert(innerParser.isAtEnd(), "source assertTrue")
				assumeProve.addField("assumption", inner)
			} else {
				assumeProve.addField("assumption", parser.translate("expression or new statement"))
			}
			if !(parser.match(88)) {
				break
			}
		}
		parser.consume(76)
		assumeProve.addField("conclusion", parser.translate("expression"))
		return assumeProve
	case 429:
		newStatement := newSyntaxCorpusAST("new")
		statementLevels := []tlago.SanyNodeKind{44, 85, 80, 38, 81}
		if parser.match(56) {
			if parser.match(statementLevels...) {
				newStatement.addChild(newSyntaxCorpusAST("statement_level"))
			}
		} else {
			parser.consume(statementLevels...)
			newStatement.addChild(newSyntaxCorpusAST("statement_level"))
		}
		newStatement.addChild(parser.translate(363))
		if parser.match(147) {
			newStatement.addChild(newSyntaxCorpusAST("set_in"))
			newStatement.addChild(parser.translate("expression"))
		}
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return newStatement
	case 432:
		label := newSyntaxCorpusAST("label")
		label.addField("name", newSyntaxCorpusAST("identifier"))
		labelName := parser.consume(358, 387)
		if labelName.IsKind(387) {
			nameParser := newSyntaxCorpusReparser(labelName.GetHeirs())
			nameParser.consume(358)
			nameParser.flatTranslate(label, 388)
		}
		label.addChild(parser.translate(90))
		label.addField("expression", parser.translate("expression"))
		syntaxCorpusAssert(parser.isAtEnd(), "source assertTrue")
		return label
	case 90:
		return newSyntaxCorpusAST("label_as")
	default:
		panic(&syntaxCorpusDSLError{fmt.Sprintf("Unhandled conversion from kind %d image %s", node.Kind, node.Image), 0})
	}
}
func syntaxCorpusToAST(root *tlago.SanySyntaxNode) *syntaxCorpusAST {
	return newSyntaxCorpusAST("source_file").addChild(syntaxCorpusTranslate(root))
}
