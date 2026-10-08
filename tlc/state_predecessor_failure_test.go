package tlc

import "testing"

// The source has no direct test for a missing predecessor at this boundary.
func TestStateMissingPredecessorPreservesSourceMutationOrder(t *testing.T) {
	oldPolicy := statePreserveMetadata
	defer func() { statePreserveMetadata = oldPolicy }()
	for _, preserve := range []bool{false, true} {
		for _, method := range []string{"mutable", "polymorphic", "polymorphic-typed-nil"} {
			name := "ordinary/" + method
			if preserve {
				name = "extended/" + method
			}
			t.Run(name, func(t *testing.T) {
				statePreserveMetadata = preserve
				prior := &TLCStateMut{level: 4}
				action := &Action{Name: "Generated"}
				state := &TLCStateMut{UID: 73, WorkerID: 2, level: 5, pred: prior, action: action}
				err := invokeDistributedServerOperation(func() error {
					switch method {
					case "mutable":
						state.SetPredecessor(nil)
					case "polymorphic":
						state.SetTracePredecessor(nil)
					case "polymorphic-typed-nil":
						var missing *TLCStateMut
						state.SetTracePredecessor(missing)
					}
					return nil
				})
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing predecessor error = %T/%v", err, err)
				}
				if preserve && state.TracePredecessor() != nil || !preserve && state.TracePredecessor() != prior {
					t.Fatal("failure lost source predecessor assignment order")
				}
				if state.level != 5 || state.UID != 73 || state.WorkerID != 2 || state.action != action {
					t.Fatal("failure changed level or unrelated metadata")
				}
			})
		}
	}
}
