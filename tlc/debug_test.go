package tlc

import (
	"math/rand"
	"strings"
	"testing"
)

func TestDebugStateVariableTypeMatchesJava(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x", "y"})

	partial := NewEmptyState().Bind(UniqueStringOf("x"), IntOne)
	partialVar := debugStateAsVariable(partial, debugStateRecordValue(partial, nil), "state", rand.New(rand.NewSource(1)))
	if partialVar.Type != "State" {
		t.Fatalf("partial state type = %q, want State", partialVar.Type)
	}

	full := partial.Bind(UniqueStringOf("y"), NewIntValue(2))
	fullVar := debugStateAsVariable(full, debugStateRecordValue(full, nil), "state", rand.New(rand.NewSource(1)))
	if !strings.HasPrefix(fullVar.Type, "FP64: ") {
		t.Fatalf("assigned state type = %q, want FP64 prefix", fullVar.Type)
	}
}

func TestDebugIntervalVariableUsesSetEnumPresentationLikeJava(t *testing.T) {
	variable := debugValueToVariable(NewDebugTLCVariableName("r"), NewIntervalValue(1, 2), rand.New(rand.NewSource(1)))

	if got, want := variable.Type, "SetEnumValue: a set of the form {e1, ... ,eN}"; got != want {
		t.Fatalf("interval debug type = %q, want %q", got, want)
	}
	if got, want := variable.Value, "{1, 2}"; got != want {
		t.Fatalf("interval debug value = %q, want %q", got, want)
	}
	if _, ok := variable.TLCValue.(*SetEnumValue); !ok {
		t.Fatalf("interval debug backing value = %T, want *SetEnumValue", variable.TLCValue)
	}
	if variable.VariablesReference == 0 {
		t.Fatalf("interval debug variable has no nested-variable reference")
	}
}
