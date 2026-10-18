package tlc

import (
	"math/rand"
	"testing"
	"time"
)

func TestJavaDebugTLCVariable(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	assertEquals := func(t *testing.T, want, got int) {
		t.Helper()
		if want != got {
			t.Fatalf("nested variables=%d, want %d", got, want)
		}
	}
	t.Run("testFiniteEmptySetValue", func(t *testing.T) {
		assertEquals(t, 0, len(NewDebugTLCVariableName("4711").SetInstance(NewSetEnumValue([]Value{}, false)).GetNested(rnd)))
	})
	t.Run("testFiniteSetValue", func(t *testing.T) {
		outer := NewDebugTLCVariableName("4711").SetInstance(NewSetEnumValue([]Value{NewSetEnumValue([]Value{}, false)}, false)).GetNested(rnd)
		assertEquals(t, 1, len(outer))
		assertEquals(t, 0, len(outer[0].GetNested(rnd)))
	})
	t.Run("testFiniteNestedValue", func(t *testing.T) {
		vars := NewDebugTLCVariableName("4711").SetInstance(NewSetEnumValue([]Value{NewSetEnumValue([]Value{NewPairTupleValue(IntZero, IntOne)}, false)}, false)).GetNested(rnd)
		assertEquals(t, 1, len(vars))
		vars = vars[0].GetNested(rnd)
		assertEquals(t, 1, len(vars))
		vars = vars[0].GetNested(rnd)
		assertEquals(t, 2, len(vars))
		for _, variable := range vars {
			assertEquals(t, 0, len(variable.GetNested(rnd)))
		}
	})
	t.Run("testInfiniteValue", func(t *testing.T) {
		assertEquals(t, 0, len(NewDebugTLCVariableName("4711").SetInstance(STRING()).GetNested(rnd)))
	})
}
