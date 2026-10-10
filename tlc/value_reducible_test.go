package tlc

import (
	"math"
	"testing"
)

func TestToolReducibleSetOperationsPreserveRawElements(t *testing.T) {
	// Capture this interning context's class-static operators before building
	// the case table; a preceding checker test may have reset the intern table.
	ensureBuiltInOPs()
	for _, tc := range []struct {
		name     string
		op       *UniqueString
		interval bool
		want     []int32
		norm     bool
	}{
		{"set difference", OpSetdiff, false, []int32{3, 3}, false},
		{"set intersection", OpCap, false, []int32{1}, false},
		{"set union", OpCup, false, []int32{3, 1, 3, 2}, false},
		{"interval difference", OpSetdiff, true, []int32{2, 3}, true},
		{"interval intersection", OpCap, true, []int32{1}, true},
		{"interval union", OpCup, true, []int32{1, 2, 3, 0}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			leftSet := NewSetEnumValue([]Value{NewIntValue(3), NewIntValue(1), NewIntValue(3)}, false)
			var left Value = leftSet
			if tc.interval {
				left = NewIntervalValue(1, 3)
			}
			right := NewSetEnumValue([]Value{NewIntValue(1)}, true)
			if tc.op == OpCup {
				right = NewSetEnumValue([]Value{NewIntValue(2), NewIntValue(3), NewIntValue(2)}, false)
				if tc.interval {
					right = NewSetEnumValue([]Value{NewIntValue(2), NewIntValue(0)}, false)
				}
			}
			// ValueNode captures display text, which normalizes sets. Supply the
			// values directly so evaluation observes the raw backing vectors.
			expr := NewBuiltinOpApplNode(tc.op, left, right)
			result, err := NewTool().Eval(expr)
			if err != nil {
				t.Fatal(err)
			}
			set, ok := result.(*SetEnumValue)
			if !ok {
				t.Fatalf("result type = %T, want SetEnumValue", result)
			}
			if set.IsNorm != tc.norm || set.Elems.Len() != len(tc.want) {
				t.Fatalf("result normalized=%v, raw length=%d; want %v, %d", set.IsNorm, set.Elems.Len(), tc.norm, len(tc.want))
			}
			for i, want := range tc.want {
				if got := set.Elems.At(i).(*IntValue).Val; got != want {
					t.Fatalf("raw element %d = %d, want %d", i, got, want)
				}
			}
			if leftSet.IsNorm || leftSet.Elems.Len() != 3 {
				t.Fatal("operation normalized or deduplicated the left backing vector")
			}
			if tc.op == OpCup && !right.IsNorm {
				t.Fatal("union did not normalize the right set through its enumeration")
			}
		})
	}
}

func TestToolIntersectionReducesEitherOperandWithoutEnumeratingTheOther(t *testing.T) {
	bound := NewFormalParamSymbolNode("element", 0)
	predicate := NewBuiltinOpApplNode(OpEq, NewOpApplNode(bound), NewValueNode(NewIntValue(1)))
	for _, reducibleFirst := range []bool{true, false} {
		tool := NewTool()
		calls := 0
		tool.EvalFunc = func(tl *Tool, node SemanticNode, args ...any) (Value, error) {
			if node == predicate {
				calls++
			}
			c, s0, s1, control, cm := parseEvalArgs(args...)
			return tl.EvalImpl(node, c, s0, s1, control, cm)
		}
		// Nat cannot be enumerated. Java intersects it via membership of the small operand.
		state := NewEmptyState()
		lazy := NewSetPredValue(bound, NatValue, predicate, tool, EmptyContext, state, state, EvalClear)
		var left, right Value = lazy, NewIntervalValue(1, 2)
		if reducibleFirst {
			left, right = right, left
		}
		result, err := tool.Eval(NewBuiltinOpApplNode(OpCap, NewValueNode(left), NewValueNode(right)))
		if err != nil {
			t.Fatal(err)
		}
		set, ok := result.(*SetEnumValue)
		if !ok || set.Elems.Len() != 1 || set.Elems.At(0).(*IntValue).Val != 1 || calls != 2 {
			t.Fatalf("intersection type=%T, predicate calls=%d; want eager {1} with two membership checks", result, calls)
		}
	}
}

func TestToolUnionCollapsesEmptyReducibleOperandsAndKeepsLazyOperandOrder(t *testing.T) {
	lazy := NewSubsetValue(NewIntervalValue(1, 2))
	for _, empty := range []Value{EmptySet, NewIntervalValue(2, 1)} {
		for _, left := range []bool{true, false} {
			var a, b Value = empty, lazy
			if !left {
				a, b = b, a
			}
			result, err := NewTool().Eval(NewBuiltinOpApplNode(OpCup, NewValueNode(a), NewValueNode(b)))
			if err != nil || result != lazy {
				t.Fatalf("empty union type=%T, error=%v; want original lazy operand", result, err)
			}
		}
	}
	interval := NewIntervalValue(1, 2)
	result, err := NewTool().Eval(NewBuiltinOpApplNode(OpCup, NewValueNode(lazy), NewValueNode(interval)))
	if err != nil {
		t.Fatal(err)
	}
	union, ok := result.(*SetCupValue)
	if !ok || union.Set1 != interval || union.Set2 != lazy {
		t.Fatalf("union = %T; want reducible operand first in the lazy union", result)
	}
}

func TestIntervalReducibleOperationsRejectSizeOverflow(t *testing.T) {
	interval := NewIntervalValue(math.MinInt32, math.MaxInt32)
	for _, op := range []*UniqueString{OpSetdiff, OpCap, OpCup} {
		if _, err := NewTool().Eval(NewBuiltinOpApplNode(op, NewValueNode(interval), NewValueNode(EmptySet))); err == nil {
			t.Fatalf("%s accepted an interval exceeding Java's 32-bit size", op)
		}
	}
}
