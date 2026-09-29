package sany_tests

import (
	"sort"

	"github.com/glycerine/tlago"
)

type sanyOperatorFixture struct {
	op          tlago.SanyOperatorInfo
	symbols     []string
	associative bool
}

func sanyOperatorFixtures() []sanyOperatorFixture {
	symbolsByCanonical := map[string][]string{}
	for _, op := range tlago.SanyCanonicalOperators {
		if !sanyOperatorTestFixity(op) {
			continue
		}
		symbolsByCanonical[op.Symbol] = []string{op.Symbol}
	}
	for synonym, canonical := range tlago.SanyOperatorSynonymCanonical {
		if _, ok := symbolsByCanonical[canonical]; ok {
			symbolsByCanonical[canonical] = append(symbolsByCanonical[canonical], synonym)
		}
	}
	out := make([]sanyOperatorFixture, 0, len(symbolsByCanonical))
	for _, op := range tlago.SanyCanonicalOperators {
		symbols, ok := symbolsByCanonical[op.Symbol]
		if !ok {
			continue
		}
		sort.Strings(symbols[1:])
		out = append(out, sanyOperatorFixture{
			op:          op,
			symbols:     symbols,
			associative: op.AssocLeft(),
		})
	}
	return out
}

func sanyOperatorTestFixity(op tlago.SanyOperatorInfo) bool {
	return op.Fixity == tlago.SanyOperatorPrefix ||
		op.Fixity == tlago.SanyOperatorInfix ||
		op.Fixity == tlago.SanyOperatorPostfix
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
	if left.op.Symbol == right.op.Symbol {
		return !left.associative
	}
	return (left.op.LowPrecedence <= right.op.LowPrecedence && right.op.LowPrecedence <= left.op.HighPrecedence) ||
		(left.op.LowPrecedence <= right.op.LowPrecedence && right.op.HighPrecedence <= left.op.HighPrecedence) ||
		(right.op.LowPrecedence <= left.op.LowPrecedence && left.op.HighPrecedence <= right.op.HighPrecedence) ||
		(left.op.LowPrecedence <= right.op.HighPrecedence && right.op.HighPrecedence <= left.op.HighPrecedence)
}

func parseSANYOperatorExpression(expr string) (*tlago.SanySyntaxNode, tlago.Diagnostics) {
	source := "---- MODULE Test ----\nASSUME " + expr + " \n===="
	return tlago.ParseSanySyntax("Test.tla", source)
}
