package tlc

import "testing"

func TestSetEnumElementsNormalizeLikeJavaEnumerator(t *testing.T) {
	set := NewSetEnumValue([]Value{NewIntValue(3), IntOne, NewIntValue(3)}, false)

	enum := set.Elements()
	got := collectEnumerationStrings(t, enum)

	if len(got) != 2 || got[0] != "1" || got[1] != "3" {
		t.Fatalf("Elements = %v, want normalized duplicate-free order [1 3]", got)
	}
	if !set.IsNormalized() {
		t.Fatalf("Elements should normalize the set like Java SetEnumValue.Enumerator")
	}
}

func TestSetEnumCompareAndEqualConvertOtherValueToSetEnumLikeJava(t *testing.T) {
	set := NewSetEnumValue([]Value{IntOne, NewIntValue(2)}, false)
	interval := NewIntervalValue(1, 2)

	cmp, err := set.Compare(interval)
	if err != nil {
		t.Fatalf("Compare interval: %v", err)
	}
	if cmp != 0 {
		t.Fatalf("Compare interval = %d, want 0", cmp)
	}

	eq, err := set.Equal(interval)
	if err != nil {
		t.Fatalf("Equal interval: %v", err)
	}
	if !eq {
		t.Fatalf("Equal interval = false, want true")
	}
}

func TestSetEnumValueJavaIsSetOfAtoms(t *testing.T) {
	ModelValueInit()
	atoms := NewSetEnumValue([]Value{IntOne, BoolTrue, NewStringValue("s"), MakeModelValue("m")}, true)
	cm := NewCostModel(nil)
	if got := NewEmptySetEnumValue(cm); got.Elems == nil || got.Elems.Len() != 0 || !got.IsNorm || got.GetCostModel().node != cm.node {
		t.Fatal("empty CostModel constructor did not retain its empty normalized vector and cost model")
	}
	if got := NewSingletonSetEnumValue(IntOne); got.Elems.Len() != 1 || got.Elems.At(0) != IntOne || !got.IsNorm {
		t.Fatal("singleton constructor did not retain its normalized element")
	}
	for _, tc := range []struct {
		name string
		set  *SetEnumValue
		want bool
	}{
		{"empty", NewEmptySetEnumValue(), true},
		{"atoms", atoms, true},
		{"nested", NewSingletonSetEnumValue(NewSingletonSetEnumValue(atoms)), true},
		{"interval", NewSetEnumValue([]Value{NewIntervalValue(1, 2)}, true), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.set.IsSetOfAtoms(); got != tc.want {
				t.Fatalf("isSetOfAtoms = %t, want %t", got, tc.want)
			}
		})
	}

	for _, tc := range []struct {
		name string
		set  *SetEnumValue
		want string
	}{
		{"null_element", NewSetEnumValue([]Value{nil}, true), `Cannot invoke "tlc2.value.impl.Value.isAtom()" because "v" is null`},
		{"null_vector", NewSetEnumValueVec(nil, true), `Cannot invoke "tlc2.value.impl.ValueVec.size()" because "this.elems" is null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				failure, ok := recover().(*NullPointerException)
				if !ok || failure.Error() != tc.want {
					t.Fatalf("panic = %T %v, want NullPointerException %q", failure, failure, tc.want)
				}
			}()
			_ = tc.set.IsSetOfAtoms()
		})
	}
}

func TestSetEnumValueToSetEnumPreservesIdentity(t *testing.T) {
	set := NewSetEnumValue([]Value{NewIntValue(2), IntOne}, false)
	if got := set.ToSetEnum(); got != set {
		t.Fatalf("ToSetEnum() = %p, want receiver %p", got, set)
	}
	if set.IsNormalized() {
		t.Fatal("ToSetEnum normalized its receiver")
	}
	converted, err := tryToSetEnumValue(set)
	if err != nil {
		t.Fatalf("tryToSetEnumValue: %v", err)
	}
	if converted != set {
		t.Fatalf("tryToSetEnumValue = %p, want receiver %p", converted, set)
	}
}

func collectEnumerationStrings(t *testing.T, enum ValueEnumeration) []string {
	t.Helper()
	var out []string
	for value := enum.NextElement(); value != nil; value = enum.NextElement() {
		out = append(out, value.String())
	}
	if err := enum.Err(); err != nil {
		t.Fatalf("enumeration error: %v", err)
	}
	return out
}
