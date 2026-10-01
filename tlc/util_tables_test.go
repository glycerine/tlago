package tlc

import "testing"

func TestVectRemoveAllPreservesJavaRawElementCount(t *testing.T) {
	v := NewVectWithCapacity[string](4)
	v.AddElement("a")
	v.AddElement("b")
	v.AddElement("c")

	v.RemoveAll(1)
	if got := v.String(); got != "{a}" {
		t.Fatalf("after shrink String = %q, want {a}", got)
	}

	v.RemoveAll(3)
	if got := v.String(); got != "{a,b,c}" {
		t.Fatalf("after re-expanding within capacity String = %q, want stale Java backing elements", got)
	}
}
