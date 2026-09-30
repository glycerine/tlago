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
