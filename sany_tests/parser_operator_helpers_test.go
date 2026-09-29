package sany_tests

import "github.com/glycerine/tlago"

type sanyOperatorFixture struct {
	op          tlago.SanyOperatorInfo
	symbols     []string
	associative bool
}

func sanyOperatorFixtures() []sanyOperatorFixture {
	return []sanyOperatorFixture{
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"\\lnot", "~", "\\neg"}, 4, 4, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"ENABLED"}, 4, 15, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"UNCHANGED"}, 4, 15, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"[]"}, 4, 15, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"<>"}, 4, 15, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"SUBSET"}, 10, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"UNION"}, 10, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"DOMAIN"}, 10, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPrefix, []string{"-"}, 12, 12, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"=>"}, 1, 1, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"-+->"}, 2, 2, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\equiv", "<=>"}, 2, 2, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"~>"}, 2, 2, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\lor", "\\/"}, 3, 3, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\land", "/\\"}, 3, 3, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"/=", "#"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"-|"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"::="}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{":="}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"<"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"="}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"=|"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{">"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\approx"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\asymp"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\cong"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\doteq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\geq", ">="}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\gg"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\in"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\notin"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\leq", "<=", "=<"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\ll"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\prec"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\preceq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\propto"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\sim"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\simeq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\sqsubset"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\sqsubseteq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\sqsupset"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\sqsupseteq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\subset"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\subseteq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\succ"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\succeq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\supset"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\supseteq"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"|-"}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"|="}, 5, 5, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\cdot"}, 5, 14, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"@@"}, 6, 6, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{":>"}, 7, 7, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"<:"}, 7, 7, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\"}, 8, 8, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\intersect", "\\cap"}, 8, 8, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\union", "\\cup"}, 8, 8, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{".."}, 9, 9, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"..."}, 9, 9, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"!!"}, 9, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"##"}, 9, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"$"}, 9, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"$$"}, 9, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"??"}, 9, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\sqcap"}, 9, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\sqcup"}, 9, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\uplus"}, 9, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\wr"}, 9, 14, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\oplus", "(+)"}, 10, 10, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"+"}, 10, 10, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"++"}, 10, 10, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"%"}, 10, 11, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"%%"}, 10, 11, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"|"}, 10, 11, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"||"}, 10, 11, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\ominus", "(-)"}, 11, 11, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"-"}, 11, 11, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"--"}, 11, 11, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"&"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"&&"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\odot", "(.)"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\oslash", "(/)"}, 13, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\otimes", "(\\X)"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"*"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"**"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"/"}, 13, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"//"}, 13, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\bigcirc"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\bullet"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\div"}, 13, 13, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\o", "\\circ"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"\\star"}, 13, 13, true),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"^"}, 14, 14, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorInfix, []string{"^^"}, 14, 14, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPostfix, []string{"^+"}, 15, 15, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPostfix, []string{"^*"}, 15, 15, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPostfix, []string{"^#"}, 15, 15, false),
		sanyOperatorFixtureOf(tlago.SanyOperatorPostfix, []string{"'"}, 15, 15, false),
	}
}

func sanyOperatorFixtureOf(fixity tlago.SanyOperatorFixity, symbols []string, low, high int, associative bool) sanyOperatorFixture {
	assoc := tlago.SanyAssociativityNone
	if associative {
		assoc = tlago.SanyAssociativityLeft
	}
	return sanyOperatorFixture{
		op: tlago.SanyOperatorInfo{
			Symbol:         symbols[0],
			LowPrecedence:  low,
			HighPrecedence: high,
			Associativity:  assoc,
			Fixity:         fixity,
		},
		symbols:     symbols,
		associative: associative,
	}
}

func sanyOperatorLowerPrecThan(left, right sanyOperatorFixture) bool {
	return left.op.LowPrecedence < right.op.LowPrecedence &&
		left.op.HighPrecedence < right.op.HighPrecedence
}

func sanyOperatorConflictsWith(left, right sanyOperatorFixture) bool {
	if left.op.Fixity == right.op.Fixity && (left.op.IsPrefix() || left.op.IsPostfix()) {
		return false
	}
	if left.op.IsInfix() && right.op.IsPrefix() {
		return false
	}
	if left.op.IsPostfix() && right.op.IsInfix() {
		return false
	}
	if sanySameOperatorFixture(left, right) {
		return !left.associative
	}
	return (left.op.LowPrecedence <= right.op.LowPrecedence && right.op.LowPrecedence <= left.op.HighPrecedence) ||
		(left.op.LowPrecedence <= right.op.LowPrecedence && right.op.HighPrecedence <= left.op.HighPrecedence) ||
		(right.op.LowPrecedence <= left.op.LowPrecedence && left.op.HighPrecedence <= right.op.HighPrecedence) ||
		(left.op.LowPrecedence <= right.op.HighPrecedence && right.op.HighPrecedence <= left.op.HighPrecedence)
}

func sanySameOperatorFixture(left, right sanyOperatorFixture) bool {
	return left.op.Symbol == right.op.Symbol &&
		left.op.Fixity == right.op.Fixity &&
		left.op.LowPrecedence == right.op.LowPrecedence &&
		left.op.HighPrecedence == right.op.HighPrecedence
}

func parseSANYOperatorExpression(expr string) (*tlago.SanySyntaxNode, tlago.Diagnostics) {
	source := "---- MODULE Test ----\nASSUME " + expr + " \n===="
	return tlago.ParseSanySyntax("Test.tla", source)
}
