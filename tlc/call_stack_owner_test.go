package tlc

import "testing"

// Source CallStackTool copies an existing Tool's shared Spec; null cannot select
// a new evaluator. There is no direct original null-constructor method.
func TestCallStackToolRequiresSourceTool(t *testing.T) {
	err := invokeDistributedServerOperation(func() error {
		NewCallStackTool(nil)
		return nil
	})
	if !isDistributedNullFailure(err) {
		t.Fatalf("missing replay owner selected a replacement tool: %T/%v", err, err)
	}
}
