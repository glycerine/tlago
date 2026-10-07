package tlago

import (
	"reflect"
	"testing"
)

// Ported from TestBuiltInOperatorInitialization.testInitAndReInit, including
// its complete testCorrectAndComplete helper and original reinitialization.
func TestTestBuiltInOperatorInitialization_testInitAndReInit(t *testing.T) {
	check := func() {
		context := sanyGlobalInitialContext(false)
		for _, expected := range sanyBuiltinOperators {
			name := expected.name
			node := context.getSymbol(name)
			if node == nil {
				t.Fatalf("%s: missing symbol", name)
			}
			actual, ok := node.(*sanySemOpDefNode)
			if !ok {
				t.Fatalf("%s: node = %T, want OpDefNode", name, node)
			}
			if actual.semName() != name {
				t.Fatalf("%s: name = %s", name, actual.semName())
			}
			if actual.semKind() != sanyBuiltInKind {
				t.Fatalf("%s: kind = %d", name, actual.semKind())
			}
			if !sanyContextIsBuiltIn(node) {
				t.Fatalf("%s: isBuiltIn = false", name)
			}
			if actual.IsStandardModule() {
				t.Fatalf("%s: isStandardModule = true", name)
			}
			if sanySymbolIsParam(node) {
				t.Fatalf("%s: isParam = true", name)
			}
			if actual.semLocal() {
				t.Fatalf("%s: isLocal = true", name)
			}
			if actual.getBody() != nil {
				t.Fatalf("%s: body is non-null", name)
			}
			if actual.semArity() != expected.arity {
				t.Fatalf("%s: arity = %d, want %d", name, actual.semArity(), expected.arity)
			}
			if expected.arity == -1 {
				if actual.getParams() != nil {
					t.Fatalf("%s: params are non-null", name)
				}
			} else {
				params := actual.getParams()
				// Java reads getParams().length; a null array must fail even
				// when expected arity is zero (Go's len(nil) would pass).
				if params == nil {
					t.Fatalf("%s: cannot read length of null params", name)
				}
				if len(params) != expected.arity {
					t.Fatalf("%s: params length = %d, want %d", name, len(params), expected.arity)
				}
			}
			if *actual.level != expected.level {
				t.Fatalf("%s: level = %d, want %d", name, *actual.level, expected.level)
			}
			if !reflect.DeepEqual(actual.getArgMaxLevels(), expected.argMaxLevels) {
				t.Fatalf("%s: argMaxLevels = %#v, want %#v", name, actual.getArgMaxLevels(), expected.argMaxLevels)
			}
			if !reflect.DeepEqual(actual.getArgWeights(), expected.argWeights) {
				t.Fatalf("%s: argWeights = %#v, want %#v", name, actual.getArgWeights(), expected.argWeights)
			}
		}
		count := 0
		for _, symbol := range context.orderedSymbols() {
			found := false
			for _, expected := range sanyBuiltinOperators {
				if symbol.semName() == expected.name {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("unexpected builtin %s", symbol.semName())
			}
			count++
		}
		if count != len(sanyBuiltinOperators) {
			t.Fatalf("builtin count = %d, want %d", count, len(sanyBuiltinOperators))
		}
	}
	check()
	sanyGlobalInitialContext(true)
	check()
}
