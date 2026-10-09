package tlc

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// ByteArrayQueue dereferences each input state during conversion, before
// publishing any raw entries. No original method covers a null state input.
func TestDistributedByteQueueRejectsNilStateBeforePublication(t *testing.T) {
	initTLCCheckerTest(t)
	for _, variables := range [][]string{nil, {"x"}} {
		name := "no_variables"
		if len(variables) != 0 {
			name = "one_variable"
		}
		t.Run(name, func(t *testing.T) {
			SetStateVariables(variables)
			for _, operation := range []string{"encode", "enqueue", "synchronized_enqueue", "array"} {
				t.Run(operation, func(t *testing.T) {
					state := NewEmptyState()
					state.UID = 9
					if len(variables) != 0 {
						state.Bind(UniqueStringOf("x"), NewIntValue(9))
					}
					raw := mustStateToBytes(state)
					q := &DiskByteArrayQueue{len: 1, enqIndex: 1, deqIndex: 4,
						enqBuf: [][]byte{raw, nil, nil, nil}, deqBuf: make([][]byte, 4)}
					q.cond = sync.NewCond(&q.mu)
					var failure any
					func() {
						defer func() { failure = recover() }()
						switch operation {
						case "encode":
							mustStateToBytes(nil)
						case "enqueue":
							q.Enqueue(nil)
						case "synchronized_enqueue":
							q.SEnqueue(nil)
						case "array":
							q.SEnqueueAll([]*TLCStateMut{state, nil})
						}
					}()
					err, ok := failure.(error)
					if !ok || !strings.Contains(err.Error(), "nil TLC state") {
						t.Fatalf("null state was not rejected as null: %T %v", failure, failure)
					}
					if q.Size() != 1 || q.enqIndex != 1 || q.deqIndex != 4 || !bytes.Equal(q.enqBuf[0], raw) {
						t.Fatal("failed conversion published work or changed the existing state")
					}
					for _, entry := range q.enqBuf[1:] {
						if entry != nil {
							t.Fatal("failed array conversion published its valid prefix")
						}
					}
					q.SEnqueue(state)
					if q.Size() != 2 || q.enqIndex != 2 || !bytes.Equal(q.enqBuf[1], raw) {
						t.Fatal("queue did not accept ordinary work after conversion failure")
					}
				})
			}
		})
	}
}
