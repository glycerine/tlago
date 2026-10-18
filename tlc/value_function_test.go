package tlc

import (
	"math"
	"testing"
)

func javaFcnRcdInts(from, to, offset int) []Value {
	values := make([]Value, 0, to-from)
	for i := from; i < to; i++ {
		values = append(values, NewIntValue(int32(offset+i)))
	}
	return values
}

func TestFcnRcdValueDirectConversionSurface(t *testing.T) {
	fcn := NewFcnRcdValue([]Value{NewStringValue("a")}, []Value{IntOne}, false)
	if fcn.ToFcnRcd() != fcn {
		t.Fatal("ToFcnRcd did not preserve receiver identity")
	}
	record := fcn.ToRcd()
	if record == nil || len(record.Names) != 1 || record.Names[0] != UniqueStringOf("a") || record.Values[0] != IntOne {
		t.Fatalf("ToRcd = %#v, want singleton a |-> 1", record)
	}
	if compatibility := fcn.ToRecord(); compatibility == nil {
		t.Fatal("ToRecord compatibility adapter returned nil")
	}
}

func TestFcnRcdValueVecConstructorCopiesActiveDomain(t *testing.T) {
	domain := NewValueVec()
	domain.Add(IntOne)
	fcn := NewFcnRcdValueVec(domain, []Value{NewStringValue("one")}, true)
	domain.Set(0, NewIntValue(2))
	if len(fcn.Domain) != 1 || fcn.Domain[0] != IntOne {
		t.Fatalf("function domain after vector mutation = %v, want [1]", fcn.Domain)
	}
}

func TestJavaFcnRcdValue(t *testing.T) {
	oldSeed, oldPoly := RandomEnumerableSeed(), FP64IrredPoly()
	oldRandom := ResetRandomEnumerableValues()
	SetRandomEnumerableSeed(15041980)
	FP64Init()
	// Isolate the Java class's static model-value registry from other Go classes.
	modelValues.Lock()
	oldCount, oldTable, oldMVs := modelValues.count, modelValues.table, modelValues.mvs
	modelValues.Unlock()
	ModelValueInit()
	t.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRandom)
		FP64InitPoly(oldPoly)
		modelValues.Lock()
		modelValues.count, modelValues.table, modelValues.mvs = oldCount, oldTable, oldMVs
		modelValues.Unlock()
	})
	t.Run("testSelecEmpty", func(t *testing.T) {
		fcn := NewFcnRcdValue([]Value{}, []Value{}, false)
		if _, err := fcn.Select(IntNegOne); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("testSelecNormalizedEmpty", func(t *testing.T) {
		fcn := NewFcnRcdValue([]Value{}, []Value{}, false)
		fcn.Normalize()
		if _, err := fcn.Select(IntNegOne); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("testEmptyIntervalDomainToTuple", func(t *testing.T) {
		fcn := NewFcnRcdIntervalValue(NewIntervalValue(2, 1), []Value{})
		tuple := fcn.ToTuple()
		if tuple == nil {
			t.Fatal("toTuple returned nil")
		}
		if eq, err := EmptyTuple.Equal(tuple); err != nil || !eq {
			t.Fatalf("empty tuple equality=%v/%v", eq, err)
		}
		if NewFcnRcdIntervalValue(NewIntervalValue(2, 2), []Value{NewIntValue(42)}).ToTuple() != nil {
			t.Fatal("expected null tuple")
		}
	})
	for _, tc := range []struct {
		name      string
		normalize bool
	}{{"testSelect", false}, {"testSelectNormalized", true}} {
		t.Run(tc.name, func(t *testing.T) {
			for upper := -64; upper < 64; upper++ {
				for lower := -64; lower < upper; lower++ {
					dom, rng := javaFcnRcdInts(lower, upper, 0), javaFcnRcdInts(lower, upper, 1024)
					fcn := NewFcnRcdValue(dom, rng, false)
					if tc.normalize {
						fcn.Normalize()
					}
					for j := -128; j < 128; j++ {
						val, err := fcn.Select(NewIntValue(int32(j)))
						if err != nil {
							t.Fatal(err)
						}
						if j < int(dom[0].(*IntValue).Val) || j > int(dom[len(dom)-1].(*IntValue).Val) {
							if val != nil {
								t.Fatalf("select(%d) on [%d,%d)=%v, want null", j, lower, upper, val)
							}
						} else {
							iv, ok := val.(*IntValue)
							if !ok || iv == nil {
								t.Fatalf("select(%d)=%T, want non-null IntValue", j, val)
							}
							if eq, err := NewIntValue(int32(j + 1024)).Equal(iv); err != nil || !eq {
								t.Fatalf("select(%d)=%v, equality error=%v", j, iv, err)
							}
						}
					}
				}
			}
		})
	}
	for _, tc := range []struct{ name, letters, message, nonModelMessage string }{
		{"testSelectLinearSearchTypedMV", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "Attempted to check equality of the differently-typed model values A_A and B_c", "Attempted to check equality of typed model value A_A and non-model value\n-1"},
		{"testSelectBinarySearchTypedMV", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "Attempted to compare the differently-typed model values A_Z and B_c", "Attempted to compare the typed model value A_Z and non-model value\n-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dom := make([]Value, 0, len(tc.letters))
			for _, c := range tc.letters {
				mv, err := TLCExtTLCModelValue(NewStringValue("A_" + string(c)))
				if err != nil {
					t.Fatal(err)
				}
				dom = append(dom, mv)
			}
			fcn := NewFcnRcdValue(dom, javaFcnRcdInts(0, len(dom), 0), false)
			fcn.Normalize()
			for _, bad := range []struct {
				value   Value
				message string
			}{{MakeModelValue("B_c"), tc.message}, {IntNegOne, tc.nonModelMessage}} {
				_, err := fcn.Select(bad.value)
				failure := javaRuntimeException(err)
				if failure == nil {
					t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
				}
				if failure.Error() != bad.message {
					t.Fatalf("message=%q, want %q", failure.Error(), bad.message)
				}
			}
			for i := range dom {
				val, err := fcn.Select(dom[i])
				if err != nil {
					t.Fatal(err)
				}
				iv, ok := val.(*IntValue)
				if !ok || iv == nil {
					t.Fatalf("select=%T, want non-null IntValue", val)
				}
				if eq, err := NewIntValue(int32(i)).Equal(iv); err != nil || !eq {
					t.Fatalf("select(%d)=%v/%v", i, iv, err)
				}
			}
		})
	}
	t.Run("testMalformedExplicitFcnEqualsIntervalDoesNotWrap", func(t *testing.T) {
		zero := IntZero
		interval := NewFcnRcdIntervalValue(NewIntervalValue(math.MaxInt32, math.MaxInt32), []Value{zero, zero})
		explicit := NewFcnRcdValue([]Value{NewIntValue(math.MaxInt32), NewIntValue(math.MinInt32)}, []Value{zero, zero}, true)
		if eq, err := explicit.Equal(interval); err != nil || eq {
			t.Fatalf("equality=%v/%v, want false", eq, err)
		}
	})
	t.Run("testMalformedIntervalFcnSelectDoesNotWrap", func(t *testing.T) {
		fcn := NewFcnRcdIntervalValue(NewIntervalValue(math.MinInt32, math.MaxInt32), []Value{IntZero})
		if val, err := fcn.Select(NewIntValue(math.MaxInt32)); err != nil || val != nil {
			t.Fatalf("select=%v/%v, want null", val, err)
		}
	})
	t.Run("testMalformedIntervalFcnExceptDoesNotWrap", func(t *testing.T) {
		fcn := NewFcnRcdIntervalValue(NewIntervalValue(math.MinInt32, math.MaxInt32), []Value{IntZero})
		ex := &ValueExcept{Path: []Value{NewIntValue(math.MaxInt32)}, Value: IntOne}
		if val, err := fcn.TakeExcept(ex); err != nil || val != fcn {
			t.Fatalf("takeExcept=%p/%v, want original %p", val, err, fcn)
		}
	})
	t.Run("testEmptyIntervalFcnCompareToAgreesWithEquals", func(t *testing.T) {
		a := NewFcnRcdIntervalValue(NewIntervalValue(2, 1), []Value{})
		b := NewFcnRcdIntervalValue(NewIntervalValue(3, 2), []Value{})
		requireJavaEmptyFcnComparison(t, a, b)
	})
	t.Run("testEmptyIntervalFcnVsEmptyTupleCompareTo", func(t *testing.T) {
		a := NewFcnRcdIntervalValue(NewIntervalValue(2, 1), []Value{})
		b := asFcnRcdValue(NewTupleValue([]Value{}))
		requireJavaEmptyFcnComparison(t, a, b)
	})
	t.Run("testEmptyIntervalFcnsNormalize", func(t *testing.T) {
		a := NewFcnRcdIntervalValue(NewIntervalValue(2, 1), []Value{})
		b := NewFcnRcdIntervalValue(NewIntervalValue(3, 2), []Value{})
		set := NewSetEnumValue([]Value{a, b}, false)
		if size, err := set.Size(); err != nil || size != 1 {
			t.Fatalf("size=%d/%v, want 1", size, err)
		}
	})
}

func requireJavaEmptyFcnComparison(t *testing.T, a, b *FcnRcdValue) {
	t.Helper()
	if eq, err := a.Equal(b); err != nil || !eq {
		t.Fatalf("equality=%v/%v, want true", eq, err)
	}
	if cmp, err := a.Compare(b); err != nil || cmp != 0 {
		t.Fatalf("compare(a,b)=%d/%v, want 0", cmp, err)
	}
	if cmp, err := b.Compare(a); err != nil || cmp != 0 {
		t.Fatalf("compare(b,a)=%d/%v, want 0", cmp, err)
	}
}
