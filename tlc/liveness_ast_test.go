package tlc

import "testing"

func TestASTToLiveRejectsUnhandledTemporalFormula(t *testing.T) {
	expr := NewOpApplNode(&SymbolNode{Name: OpCase})
	expr.SetLevel(TLCLevelTemporal)

	live, err := ASTToLive(nil, expr, EmptyContext)
	if err == nil {
		t.Fatalf("ASTToLive returned (%#v, nil), want cannot-handle error", live)
	}
	tlcErr, ok := err.(*TLCError)
	if !ok {
		t.Fatalf("ASTToLive error = %T %v, want *TLCError", err, err)
	}
	if tlcErr.Code != ECTLCLiveCannotHandleFormula {
		t.Fatalf("ASTToLive error code = %d, want %d", tlcErr.Code, ECTLCLiveCannotHandleFormula)
	}
}

func TestASTToLiveExpandsWeakFairnessLikeJava(t *testing.T) {
	subscript := NewNumeralNode(0)
	body := NewOpApplNode(NewSymbolNode("A"))
	expr := NewOpApplNode(&SymbolNode{Name: OpWF}, subscript, body)

	live, err := ASTToLive(nil, expr, EmptyContext)
	if err != nil {
		t.Fatalf("ASTToLive(WF) error = %v", err)
	}
	if live == nil || live.Kind != LiveExprAll {
		t.Fatalf("ASTToLive(WF) = %#v, want []<> disjunction", live)
	}
	eventually := live.Body
	if eventually == nil || eventually.Kind != LiveExprEven {
		t.Fatalf("WF body = %#v, want <>", eventually)
	}
	disj := eventually.Body
	if disj == nil || disj.Kind != LiveExprDisj || disj.Count() != 2 {
		t.Fatalf("WF eventual body = %#v, want disjunction", disj)
	}
	if neg := disj.GetBody(0); neg == nil || neg.Kind != LiveExprNeg || neg.Body == nil || neg.Body.Kind != LiveExprState {
		t.Fatalf("WF first disjunct = %#v, want -ENABLED state predicate", neg)
	}
	if action := disj.GetBody(1); action == nil || action.Kind != LiveExprAction {
		t.Fatalf("WF second disjunct = %#v, want action", action)
	}
}

func TestASTToLiveExpandsStrongFairnessLikeJava(t *testing.T) {
	subscript := NewNumeralNode(0)
	body := NewOpApplNode(NewSymbolNode("A"))
	expr := NewOpApplNode(&SymbolNode{Name: OpSF}, subscript, body)

	live, err := ASTToLive(nil, expr, EmptyContext)
	if err != nil {
		t.Fatalf("ASTToLive(SF) error = %v", err)
	}
	if live == nil || live.Kind != LiveExprDisj || live.Count() != 2 {
		t.Fatalf("ASTToLive(SF) = %#v, want disjunction", live)
	}
	left := live.GetBody(0)
	if left == nil || left.Kind != LiveExprEven || left.Body == nil || left.Body.Kind != LiveExprAll {
		t.Fatalf("SF left disjunct = %#v, want <>[]", left)
	}
	neg := left.Body.Body
	if neg == nil || neg.Kind != LiveExprNeg || neg.Body == nil || neg.Body.Kind != LiveExprState {
		t.Fatalf("SF left body = %#v, want -ENABLED state predicate", neg)
	}
	right := live.GetBody(1)
	if right == nil || right.Kind != LiveExprAll || right.Body == nil || right.Body.Kind != LiveExprEven {
		t.Fatalf("SF right disjunct = %#v, want []<>", right)
	}
	if action := right.Body.Body; action == nil || action.Kind != LiveExprAction {
		t.Fatalf("SF right body = %#v, want action", action)
	}
}

func TestASTToLiveExpandsBoundedExistsLikeJava(t *testing.T) {
	body := NewOpApplNode(&SymbolNode{Name: OpBox}, NewOpApplNode(NewSymbolNode("P")))
	expr := NewOpApplNode(&SymbolNode{Name: OpBE}, body)
	expr.BdedQuantBounds = []SemanticNode{NewValueNode(NewSetEnumValue([]Value{IntOne, NewIntValue(2)}, true))}
	expr.BdedQuantSymbolLists = [][]*SymbolNode{{NewSymbolNode("x")}}
	expr.BdedQuantATuple = []bool{false}

	live, err := ASTToLive(&Tool{}, expr, EmptyContext)
	if err != nil {
		t.Fatalf("ASTToLive(\\E) error = %v", err)
	}
	if live == nil || live.Kind != LiveExprDisj || live.Count() != 2 {
		t.Fatalf("ASTToLive(\\E) = %#v, want two-branch disjunction", live)
	}
	for i := 0; i < live.Count(); i++ {
		if branch := live.GetBody(i); branch == nil || branch.Kind != LiveExprAll {
			t.Fatalf("branch %d = %#v, want temporal body", i, branch)
		}
	}
}
