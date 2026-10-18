package tlc

import "testing"

func TestTLCExtCacheIsScopedByToolIDLikeJava(t *testing.T) {
	expr := NewValueNode(NewIntValue(0))
	key := NewStringValue("same-closure")
	toolA := &Tool{ID: 101}
	toolB := &Tool{ID: 202}
	computesA := 0
	computesB := 0

	valueA, err := tlcExtCacheForTool(toolA, expr).Eval(key, func() (Value, error) {
		computesA++
		return NewIntValue(1), nil
	})
	if err != nil {
		t.Fatalf("tool A first cache eval returned error: %v", err)
	}
	valueB, err := tlcExtCacheForTool(toolB, expr).Eval(key, func() (Value, error) {
		computesB++
		return NewIntValue(2), nil
	})
	if err != nil {
		t.Fatalf("tool B first cache eval returned error: %v", err)
	}
	valueA2, err := tlcExtCacheForTool(toolA, expr).Eval(key, func() (Value, error) {
		computesA++
		return NewIntValue(3), nil
	})
	if err != nil {
		t.Fatalf("tool A second cache eval returned error: %v", err)
	}

	if valueA.(*IntValue).Val != 1 || valueB.(*IntValue).Val != 2 || valueA2.(*IntValue).Val != 1 {
		t.Fatalf("cached values = %v, %v, %v; want 1, 2, cached 1", valueA, valueB, valueA2)
	}
	if computesA != 1 || computesB != 1 {
		t.Fatalf("compute counts = %d/%d, want one compute per tool", computesA, computesB)
	}
}
