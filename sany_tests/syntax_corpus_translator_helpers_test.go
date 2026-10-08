package sany_tests

// Ported from test/tla2sany/parser/TlaPlusParserOutputTranslator.java.
import (
	"fmt"
	"github.com/glycerine/tlago"
	"strings"
)

type syntaxCorpusAssertionFailure struct{ message string }

func (e *syntaxCorpusAssertionFailure) Error() string { return e.message }
func syntaxCorpusAssert(condition bool, message string) {
	if !condition {
		panic(&syntaxCorpusAssertionFailure{message})
	}
}

type syntaxCorpusReparser struct {
	nodes   []*tlago.SanySyntaxNode
	current int
}

func newSyntaxCorpusReparser(nodes []*tlago.SanySyntaxNode, start ...int) *syntaxCorpusReparser {
	current := 0
	if len(start) != 0 {
		current = start[0]
	}
	return &syntaxCorpusReparser{nodes, current}
}
func (p *syntaxCorpusReparser) lookahead() *syntaxCorpusReparser {
	return newSyntaxCorpusReparser(p.nodes, p.current)
}
func (p *syntaxCorpusReparser) merge(other *syntaxCorpusReparser) { p.current = other.current }
func (p *syntaxCorpusReparser) isAtEnd() bool                     { return p.current == len(p.nodes) }
func (p *syntaxCorpusReparser) previous() *tlago.SanySyntaxNode   { return p.nodes[p.current-1] }
func (p *syntaxCorpusReparser) advance() *tlago.SanySyntaxNode {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}
func (p *syntaxCorpusReparser) peek() *tlago.SanySyntaxNode { return p.nodes[p.current] }
func (p *syntaxCorpusReparser) check(kind tlago.SanyNodeKind) bool {
	return !p.isAtEnd() && p.peek().IsKind(kind)
}
func (p *syntaxCorpusReparser) match(kinds ...tlago.SanyNodeKind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}
	return false
}
func syntaxCorpusKindToName(kind tlago.SanyNodeKind) string {
	name := ""
	if int(kind) < len(tlago.SanyTokenImages) {
		name = tlago.SanyTokenKind(kind).JavaImage()
	} else {
		name = tlago.NewSanyNode(kind).Image
	}
	return fmt.Sprintf("[%d] %s", kind, name)
}
func (p *syntaxCorpusReparser) consume(kinds ...tlago.SanyNodeKind) *tlago.SanySyntaxNode {
	for _, kind := range kinds {
		if p.check(kind) {
			return p.advance()
		}
	}
	names := make([]string, len(kinds))
	for i, kind := range kinds {
		names[i] = syntaxCorpusKindToName(kind)
	}
	expected := strings.Join(names, ", ")
	if p.isAtEnd() {
		panic(&syntaxCorpusDSLError{"EOF; expected " + expected, p.current})
	}
	panic(&syntaxCorpusDSLError{fmt.Sprintf("Expected %s; actual %s", expected, syntaxCorpusKindToName(p.peek().Kind)), p.current})
}
func syntaxCorpusCommaSeparatedIDs(parser *syntaxCorpusReparser) []*syntaxCorpusAST {
	var ids []*syntaxCorpusAST
	for {
		parser.consume(tlago.SanyNodeKind(tlago.SanyTokenIdentifier))
		ids = append(ids, newSyntaxCorpusAST("identifier"))
		if !parser.match(tlago.SanyNodeKind(tlago.SanyTokenComma)) {
			break
		}
	}
	return ids
}
func syntaxCorpusTupleOfIdentifiers(parser *syntaxCorpusReparser) *syntaxCorpusAST {
	tuple := newSyntaxCorpusAST("tuple_of_identifiers")
	parser.consume(tlago.SanyNodeKind(tlago.SanyTokenLab))
	tuple.addChild(newSyntaxCorpusAST("langle_bracket"))
	tuple.addChildren(syntaxCorpusCommaSeparatedIDs(parser))
	parser.consume(tlago.SanyNodeKind(tlago.SanyTokenRab))
	tuple.addChild(newSyntaxCorpusAST("rangle_bracket"))
	return tuple
}
func syntaxCorpusIdentifier(input *tlago.SanySyntaxNode) *syntaxCorpusAST {
	syntaxCorpusAssert(input.Kind == tlago.SanyNodeKind(tlago.SanyTokenIdentifier), fmt.Sprintf("expected:<%d> but was:<%d>", tlago.SanyTokenIdentifier, input.Kind))
	switch input.Image {
	case "TRUE":
		return newSyntaxCorpusAST("boolean")
	case "FALSE":
		return newSyntaxCorpusAST("boolean")
	case "BOOLEAN":
		return newSyntaxCorpusAST("boolean_set")
	case "STRING":
		return newSyntaxCorpusAST("string_set")
	case "Nat":
		return newSyntaxCorpusAST("nat_number_set")
	case "ℕ":
		return newSyntaxCorpusAST("nat_number_set")
	case "Int":
		return newSyntaxCorpusAST("int_number_set")
	case "ℤ":
		return newSyntaxCorpusAST("int_number_set")
	case "Real":
		return newSyntaxCorpusAST("real_number_set")
	case "ℝ":
		return newSyntaxCorpusAST("real_number_set")
	case "@":
		return newSyntaxCorpusAST("prev_func_val")
	default:
		return newSyntaxCorpusAST("identifier_ref")
	}
}
func syntaxCorpusPrefixOpFromString(op string) *syntaxCorpusAST {
	canonical := tlago.ResolveSanyOperatorSynonym(op)
	switch canonical {
	case "\\lnot":
		return newSyntaxCorpusAST("lnot")
	case "UNION":
		return newSyntaxCorpusAST("union")
	case "SUBSET":
		return newSyntaxCorpusAST("powerset")
	case "DOMAIN":
		return newSyntaxCorpusAST("domain")
	case "-.":
		return newSyntaxCorpusAST("negative")
	case "ENABLED":
		return newSyntaxCorpusAST("enabled")
	case "UNCHANGED":
		return newSyntaxCorpusAST("unchanged")
	case "[]":
		return newSyntaxCorpusAST("always")
	case "<>":
		return newSyntaxCorpusAST("eventually")
	default:
		panic(&syntaxCorpusDSLError{fmt.Sprintf("Operator translation not defined: %s", op), 0})
	}
}
func syntaxCorpusInfixOpFromString(op string) *syntaxCorpusAST {
	if op == "<=>" || op == "⇔" {
		return newSyntaxCorpusAST("iff")
	}
	canonical := tlago.ResolveSanyOperatorSynonym(op)
	switch canonical {
	case "&":
		return newSyntaxCorpusAST("amp")
	case "&&":
		return newSyntaxCorpusAST("ampamp")
	case "\\approx":
		return newSyntaxCorpusAST("approx")
	case ":=":
		return newSyntaxCorpusAST("assign")
	case "\\asymp":
		return newSyntaxCorpusAST("asymp")
	case "\\bigcirc":
		return newSyntaxCorpusAST("bigcirc")
	case "::=":
		return newSyntaxCorpusAST("bnf_rule")
	case "\\bullet":
		return newSyntaxCorpusAST("bullet")
	case "\\intersect":
		return newSyntaxCorpusAST("cap")
	case "\\cdot":
		return newSyntaxCorpusAST("cdot")
	case "\\o":
		return newSyntaxCorpusAST("circ")
	case "@@":
		return newSyntaxCorpusAST("compose")
	case "\\cong":
		return newSyntaxCorpusAST("cong")
	case "\\union":
		return newSyntaxCorpusAST("cup")
	case "\\div":
		return newSyntaxCorpusAST("div")
	case "$":
		return newSyntaxCorpusAST("dol")
	case "$$":
		return newSyntaxCorpusAST("doldol")
	case "\\doteq":
		return newSyntaxCorpusAST("doteq")
	case "..":
		return newSyntaxCorpusAST("dots_2")
	case "...":
		return newSyntaxCorpusAST("dots_3")
	case "=":
		return newSyntaxCorpusAST("eq")
	case "\\equiv":
		return newSyntaxCorpusAST("equiv")
	case "!!":
		return newSyntaxCorpusAST("excl")
	case "\\geq":
		return newSyntaxCorpusAST("geq")
	case "\\gg":
		return newSyntaxCorpusAST("gg")
	case ">":
		return newSyntaxCorpusAST("gt")
	case "##":
		return newSyntaxCorpusAST("hashhash")
	case "=>":
		return newSyntaxCorpusAST("implies")
	case "\\in":
		return newSyntaxCorpusAST("in")
	case "\\land":
		return newSyntaxCorpusAST("land")
	case "=|":
		return newSyntaxCorpusAST("ld_ttile")
	case "~>":
		return newSyntaxCorpusAST("leads_to")
	case "\\leq":
		return newSyntaxCorpusAST("leq")
	case "\\ll":
		return newSyntaxCorpusAST("ll")
	case "\\lor":
		return newSyntaxCorpusAST("lor")
	case "-|":
		return newSyntaxCorpusAST("ls_ttile")
	case "<":
		return newSyntaxCorpusAST("lt")
	case "<:":
		return newSyntaxCorpusAST("map_from")
	case ":>":
		return newSyntaxCorpusAST("map_to")
	case "-":
		return newSyntaxCorpusAST("minus")
	case "--":
		return newSyntaxCorpusAST("minusminus")
	case "%":
		return newSyntaxCorpusAST("mod")
	case "%%":
		return newSyntaxCorpusAST("modmod")
	case "*":
		return newSyntaxCorpusAST("mul")
	case "**":
		return newSyntaxCorpusAST("mulmul")
	case "/=":
		return newSyntaxCorpusAST("neq")
	case "#":
		return newSyntaxCorpusAST("neq")
	case "\\notin":
		return newSyntaxCorpusAST("notin")
	case "\\odot":
		return newSyntaxCorpusAST("odot")
	case "\\ominus":
		return newSyntaxCorpusAST("ominus")
	case "\\oplus":
		return newSyntaxCorpusAST("oplus")
	case "\\oslash":
		return newSyntaxCorpusAST("oslash")
	case "\\otimes":
		return newSyntaxCorpusAST("otimes")
	case "+":
		return newSyntaxCorpusAST("plus")
	case "-+->":
		return newSyntaxCorpusAST("plus_arrow")
	case "++":
		return newSyntaxCorpusAST("plusplus")
	case "^":
		return newSyntaxCorpusAST("pow")
	case "^^":
		return newSyntaxCorpusAST("powpow")
	case "\\prec":
		return newSyntaxCorpusAST("prec")
	case "\\preceq":
		return newSyntaxCorpusAST("preceq")
	case "\\propto":
		return newSyntaxCorpusAST("propto")
	case "??":
		return newSyntaxCorpusAST("qq")
	case "|=":
		return newSyntaxCorpusAST("rd_ttile")
	case "|-":
		return newSyntaxCorpusAST("rs_ttile")
	case "\\":
		return newSyntaxCorpusAST("setminus")
	case "\\sim":
		return newSyntaxCorpusAST("sim")
	case "\\simeq":
		return newSyntaxCorpusAST("simeq")
	case "/":
		return newSyntaxCorpusAST("slash")
	case "//":
		return newSyntaxCorpusAST("slashslash")
	case "\\sqcap":
		return newSyntaxCorpusAST("sqcap")
	case "\\sqcup":
		return newSyntaxCorpusAST("sqcup")
	case "\\sqsubset":
		return newSyntaxCorpusAST("sqsubset")
	case "\\sqsubseteq":
		return newSyntaxCorpusAST("sqsubseteq")
	case "\\sqsupset":
		return newSyntaxCorpusAST("sqsupset")
	case "\\sqsupseteq":
		return newSyntaxCorpusAST("sqsupseteq")
	case "\\star":
		return newSyntaxCorpusAST("star")
	case "\\subset":
		return newSyntaxCorpusAST("subset")
	case "\\subseteq":
		return newSyntaxCorpusAST("subseteq")
	case "\\succ":
		return newSyntaxCorpusAST("succ")
	case "\\succeq":
		return newSyntaxCorpusAST("succeq")
	case "\\supset":
		return newSyntaxCorpusAST("supset")
	case "\\supseteq":
		return newSyntaxCorpusAST("supseteq")
	case "\\times":
		return newSyntaxCorpusAST("times")
	case "\\uplus":
		return newSyntaxCorpusAST("uplus")
	case "|":
		return newSyntaxCorpusAST("vert")
	case "||":
		return newSyntaxCorpusAST("vertvert")
	case "\\wr":
		return newSyntaxCorpusAST("wr")
	default:
		panic(&syntaxCorpusDSLError{fmt.Sprintf("Operator translation not defined: %s", op), 0})
	}
}
func syntaxCorpusPostfixOpFromString(op string) *syntaxCorpusAST {
	canonical := tlago.ResolveSanyOperatorSynonym(op)
	switch canonical {
	case "^+":
		return newSyntaxCorpusAST("sup_plus")
	case "^*":
		return newSyntaxCorpusAST("asterisk")
	case "^#":
		return newSyntaxCorpusAST("sup_hash")
	case "'":
		return newSyntaxCorpusAST("prime")
	default:
		panic(&syntaxCorpusDSLError{fmt.Sprintf("Operator translation not defined: %s", op), 0})
	}
}
