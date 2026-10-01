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
