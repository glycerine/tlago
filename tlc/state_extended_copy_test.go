package tlc

import (
	"math"
	"strconv"
	"testing"
)

// No direct upstream method tests TLCStateMutExt.copyExt. The public shallow
// and deep copy paths both reapply the predecessor after allocating values.
func TestExtendedStateCopyReappliesPredecessor(t *testing.T) {
	oldPolicy := statePreserveMetadata
	statePreserveMetadata = true
	t.Cleanup(func() { statePreserveMetadata = oldPolicy })
	for _, deep := range []bool{false, true} {
		name := "shallow"
		if deep {
			name = "deep"
		}
		for _, parentLevel := range []int{-1, 3, 8, math.MaxInt32} {
			t.Run(name+"/"+strconv.Itoa(parentLevel), func(t *testing.T) {
				state := &TLCStateMut{level: 4, UID: 42, WorkerID: 7, action: &Action{Name: "Next"}}
				if parentLevel >= 0 {
					state.pred = &TLCStateMut{level: parentLevel}
				}
				var copied *TLCStateMut
				err := invokeDistributedServerOperation(func() error {
					if deep {
						copied = state.DeepCopy()
					} else {
						copied = state.Copy()
					}
					return nil
				})
				if parentLevel == math.MaxInt32 {
					if failure, ok := err.(*TLCError); !ok || failure.Code != ECTLCTraceTooLong || copied != nil {
						t.Fatalf("copy depth-limit failure = %T/%v, copy=%v", err, err, copied)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				wantLevel := TLCStateInitLevel
				if deep {
					wantLevel = state.Level()
				}
				if parentLevel >= 0 {
					wantLevel = parentLevel + 1
				}
				if copied.Level() != wantLevel || copied.TracePredecessor() != state.TracePredecessor() || copied.GetAction() != state.GetAction() {
					t.Fatalf("copy level/metadata = %d/%p/%p, want %d/%p/%p", copied.Level(), copied.TracePredecessor(), copied.GetAction(), wantLevel, state.TracePredecessor(), state.GetAction())
				}
				wantUID, wantWorker := int64(TLCStateInitUID), TLCStateInitWorkerID
				if deep {
					wantUID, wantWorker = state.UID, state.WorkerID
				}
				if copied.UID != wantUID || copied.WorkerID != wantWorker || state.Level() != 4 {
					t.Fatal("copy changed source level or worker/UID copy policy")
				}
			})
		}
	}
}
