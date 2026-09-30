package tlc

import (
	"fmt"
	"testing"
)

func TestKSubsetValueEnumeratesBeyondUint32Masks(t *testing.T) {
	cases := []struct {
		n    int32
		want int
	}{
		{32, 496},
		{33, 528},
		{63, 1953},
		{64, 2016},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("n=%d", tc.n), func(t *testing.T) {
			kSubset := NewKSubsetValue(2, NewIntervalValue(1, tc.n))
			size, err := kSubset.Size()
			if err != nil {
				t.Fatalf("Size returned error: %v", err)
			}
			if size != tc.want {
				t.Fatalf("Size = %d, want %d", size, tc.want)
			}
			set, err := kSubset.ToSetEnum()
			if err != nil {
				t.Fatalf("ToSetEnum returned error: %v", err)
			}
			setSize, err := set.Size()
			if err != nil {
				t.Fatalf("set Size returned error: %v", err)
			}
			if setSize != tc.want {
				t.Fatalf("enumerated size = %d, want %d", setSize, tc.want)
			}
			for i := 0; i < set.Elems.Len(); i++ {
				elemSize, err := set.Elems.At(i).Size()
				if err != nil {
					t.Fatalf("element %d Size returned error: %v", i, err)
				}
				if elemSize != 2 {
					t.Fatalf("element %d has size %d, want 2: %s", i, elemSize, set.Elems.At(i))
				}
			}
		})
	}
}

func TestKSubsetValueEnumerationNormalizesBaseSet(t *testing.T) {
	base := NewSetEnumValue([]Value{
		IntOne,
		NewIntValue(7),
		NewIntValue(42),
		NewIntValue(42),
		NewIntValue(23),
	}, false)
	kSubset := NewKSubsetValue(2, base)

	enum := kSubset.Elements()
	want := []Value{
		NewSetEnumValue([]Value{IntOne, NewIntValue(7)}, false),
		NewSetEnumValue([]Value{IntOne, NewIntValue(23)}, false),
		NewSetEnumValue([]Value{IntOne, NewIntValue(42)}, false),
		NewSetEnumValue([]Value{NewIntValue(7), NewIntValue(23)}, false),
		NewSetEnumValue([]Value{NewIntValue(7), NewIntValue(42)}, false),
		NewSetEnumValue([]Value{NewIntValue(23), NewIntValue(42)}, false),
	}
	for i, expected := range want {
		got := enum.NextElement()
		if got == nil {
			t.Fatalf("element %d = nil, want %s", i, expected)
		}
		if !mustValueEqual(got, expected) {
			t.Fatalf("element %d = %s, want %s", i, got, expected)
		}
	}
	if got := enum.NextElement(); got != nil {
		t.Fatalf("extra element = %s", got)
	}
	if err := enum.Err(); err != nil {
		t.Fatalf("enumeration error: %v", err)
	}
}

func TestKSubsetValueInvalidKDenotesEmptySet(t *testing.T) {
	base := NewIntervalValue(1, 3)
	for _, k := range []int{-1, 4} {
		kSubset := NewKSubsetValue(k, base)
		size, err := kSubset.Size()
		if err != nil {
			t.Fatalf("Size(k=%d) returned error: %v", k, err)
		}
		if size != 0 {
			t.Fatalf("Size(k=%d) = %d, want 0", k, size)
		}
		empty, err := IsEmptyValue(kSubset)
		if err != nil {
			t.Fatalf("IsEmptyValue(k=%d) returned error: %v", k, err)
		}
		if !empty {
			t.Fatalf("IsEmptyValue(k=%d) = false, want true", k)
		}
		eq, err := kSubset.Equal(EmptySet)
		if err != nil {
			t.Fatalf("Equal(k=%d) returned error: %v", k, err)
		}
		if !eq {
			t.Fatalf("invalid k=%d should equal empty set", k)
		}
		cmp, err := kSubset.Compare(EmptySet)
		if err != nil {
			t.Fatalf("Compare(k=%d) returned error: %v", k, err)
		}
		if cmp != 0 {
			t.Fatalf("Compare(k=%d) = %d, want 0", k, cmp)
		}
		if got := kSubset.Elements().NextElement(); got != nil {
			t.Fatalf("Elements(k=%d) first = %s, want nil", k, got)
		}
		set, err := kSubset.ToSetEnum()
		if err != nil {
			t.Fatalf("ToSetEnum(k=%d) returned error: %v", k, err)
		}
		eq, err = set.Equal(EmptySet)
		if err != nil {
			t.Fatalf("empty set Equal(k=%d) returned error: %v", k, err)
		}
		if !eq || kSubset.String() != "{}" {
			t.Fatalf("invalid k=%d set/string = %s/%q, want empty", k, set, kSubset.String())
		}
	}
}

func TestKSubsetValueZeroIsBaseIndependent(t *testing.T) {
	a := NewKSubsetValue(0, NewIntervalValue(1, 3))
	b := NewKSubsetValue(0, NewIntervalValue(1, 4))
	eq, err := a.Equal(b)
	if err != nil {
		t.Fatalf("Equal returned error: %v", err)
	}
	if !eq {
		t.Fatalf("k=0 subsets should be equal independent of finite base set")
	}
	set, err := a.ToSetEnum()
	if err != nil {
		t.Fatalf("ToSetEnum returned error: %v", err)
	}
	if size, err := set.Size(); err != nil || size != 1 {
		t.Fatalf("enumerated size = %d/%v, want 1/nil", size, err)
	}
	if elem := set.Elems.At(0); !mustValueEqual(elem, EmptySet) {
		t.Fatalf("k=0 element = %s, want empty set", elem)
	}
}

func TestKSubsetValueLargeStringDoesNotNeedIntSizedCardinality(t *testing.T) {
	got := NewKSubsetValue(32, NewIntervalValue(1, 64)).String()
	want := "{s \\in SUBSET (1..64) : Cardinality(s) = 32}"
	if got != want {
		t.Fatalf("String = %q, want %q", got, want)
	}
}
