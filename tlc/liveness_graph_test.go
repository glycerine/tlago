package tlc

import "testing"

func TestBEGraphGetPathUsesJavaFailureCode(t *testing.T) {
	start := NewBEGraphNode(1)
	end := NewBEGraphNode(2)
	_, err := BEGraphGetPath(start, end)
	if err == nil {
		t.Fatalf("BEGraphGetPath returned nil error for disconnected nodes")
	}
	tlcErr, ok := err.(*TLCError)
	if !ok {
		t.Fatalf("BEGraphGetPath error = %T, want TLCError", err)
	}
	if tlcErr.Code != ECTLCLiveBEGraphFailedToConstruct {
		t.Fatalf("BEGraphGetPath error code = %d, want %d", tlcErr.Code, ECTLCLiveBEGraphFailedToConstruct)
	}
	if got, want := tlcErr.Error(), "BEGraph.GetPath: Failed to construct a path."; got != want {
		t.Fatalf("BEGraphGetPath error = %q, want Java message %q", got, want)
	}
}
