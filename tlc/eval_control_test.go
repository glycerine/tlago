package tlc

import "testing"

func TestPartialBooleanInvalidValuesFailLikeJavaDefault(t *testing.T) {
	expectPanic(t, func() { _ = PartialBoolean(99).String() })
	expectPanic(t, func() { _ = PartialBoolean(99).IsDefinitely(true) })
}
