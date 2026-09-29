package sany_tests

import (
	"reflect"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/utilities/VectorTest.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestVectorTest_containsUsesIdentityComparison(t *testing.T) {
	vector := tlago.NewSanyVector[*sanyVectorBox[string]]()
	reference := &sanyVectorBox[string]{value: "value"}
	vector.AddElement(reference)

	if !vector.Contains(reference) {
		t.Fatal("vector should contain the same reference")
	}
	if vector.Contains(&sanyVectorBox[string]{value: "value"}) {
		t.Fatal("vector should not contain an equal value at a different reference")
	}
}

func TestVectorTest_insertElementAtRejectsAppendPosition(t *testing.T) {
	vector := tlago.NewSanyVector[string]()
	vector.AddElement("a")
	vector.AddElement("b")

	defer func() {
		if recover() == nil {
			t.Fatal("expected append-position insert to panic")
		}
	}()
	vector.InsertElementAt("c", 2)
}

func TestVectorTest_elementsReturnsSnapshot(t *testing.T) {
	vector := tlago.NewSanyVector[string]()
	vector.AddElement("a")
	vector.AddElement("b")

	enumeration := vector.Elements()
	vector.AddElement("c")

	if want := []string{"a", "b"}; !reflect.DeepEqual(enumeration, want) {
		t.Fatalf("enumeration snapshot = %#v, want %#v", enumeration, want)
	}
}

func TestVectorTest_appendNoRepeatsUsesIdentity(t *testing.T) {
	base := tlago.NewSanyVector[*sanyVectorBox[string]]()
	shared := &sanyVectorBox[string]{value: "dup"}
	equalButDistinct := &sanyVectorBox[string]{value: "dup"}
	base.AddElement(shared)

	incoming := tlago.NewSanyVector[*sanyVectorBox[string]]()
	incoming.AddElement(equalButDistinct)
	incoming.AddElement(shared)

	base.AppendNoRepeats(incoming)

	if got, want := base.Size(), 2; got != want {
		t.Fatalf("vector size = %d, want %d", got, want)
	}
	if got := base.ElementAt(0); got != shared {
		t.Fatalf("first element = %p, want shared %p", got, shared)
	}
	if got := base.ElementAt(1); got != equalButDistinct {
		t.Fatalf("second element = %p, want distinct duplicate %p", got, equalButDistinct)
	}
}

type sanyVectorBox[T any] struct {
	value T
}
