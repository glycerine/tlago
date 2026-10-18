package tlago

import (
	"reflect"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Ported from TestBuiltInOperatorInitialization.testInitAndReInit, including
// its complete testCorrectAndComplete helper and original reinitialization.
func TestTestBuiltInOperatorInitialization_testInitAndReInit(t *testing.T) {
	check := func() {
		context := SanyContextGetGlobalContext()
		for _, expected := range SanyBuiltinOperatorProperties {
			name := expected.Name.String()
			key := SanyContextKey{Name: tlc.UniqueStringOf(name)}
			if !context.OccurSymbol(key) {
				t.Fatalf("%s: occurSymbol = false", name)
			}
			node := context.GetSymbol(key)
			if node == nil {
				t.Fatalf("%s: missing symbol", name)
			}
			actual, ok := node.(*tlc.OpDefNode)
			if !ok {
				t.Fatalf("%s: node = %T, want OpDefNode", name, node)
			}
			if actual.Name != expected.Name {
				t.Fatalf("%s: name = %s", name, actual.Name)
			}
			if actual.Kind() != tlc.SemanticBuiltInKind {
				t.Fatalf("%s: kind = %d", name, actual.Kind())
			}
			if !actual.IsBuiltIn() {
				t.Fatalf("%s: isBuiltIn = false", name)
			}
			if actual.IsStandardModule() {
				t.Fatalf("%s: isStandardModule = true", name)
			}
			if actual.Symbol.IsParam() {
				t.Fatalf("%s: isParam = true", name)
			}
			if actual.IsLocal() {
				t.Fatalf("%s: isLocal = true", name)
			}
			if actual.GetBody() != nil {
				t.Fatalf("%s: body is non-null", name)
			}
			if actual.Arity() != expected.Arity {
				t.Fatalf("%s: arity = %d, want %d", name, actual.Arity(), expected.Arity)
			}
			if expected.Arity == -1 {
				if actual.GetParams() != nil {
					t.Fatalf("%s: params are non-null", name)
				}
			} else {
				params := actual.GetParams()
				// Java reads getParams().length; a null array must fail even
				// when expected arity is zero (Go's len(nil) would pass).
				if params == nil {
					t.Fatalf("%s: cannot read length of null params", name)
				}
				if len(params) != expected.Arity {
					t.Fatalf("%s: params length = %d, want %d", name, len(params), expected.Arity)
				}
			}
			if actual.LevelValue != expected.OpLevel {
				t.Fatalf("%s: level = %d, want %d", name, actual.LevelValue, expected.OpLevel)
			}
			if !reflect.DeepEqual(actual.GetArgMaxLevels(), expected.ArgMaxLevels) {
				t.Fatalf("%s: argMaxLevels = %#v, want %#v", name, actual.GetArgMaxLevels(), expected.ArgMaxLevels)
			}
			if !reflect.DeepEqual(actual.GetArgWeights(), expected.ArgWeights) {
				t.Fatalf("%s: argWeights = %#v, want %#v", name, actual.GetArgWeights(), expected.ArgWeights)
			}
		}
		count := 0
		enumeration := context.GetContextSymbolEnumeration()
		for enumeration.HasMoreElements() {
			symbol := enumeration.NextElement().(*tlc.OpDefNode)
			found := false
			for _, expected := range SanyBuiltinOperatorProperties {
				if symbol.Name == expected.Name {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("unexpected builtin %s", symbol.Name)
			}
			count++
		}
		if count != len(SanyBuiltinOperatorProperties) {
			t.Fatalf("builtin count = %d, want %d", count, len(SanyBuiltinOperatorProperties))
		}
	}
	check()
	SanyContextReInit()
	check()
}
