package tlc

import "testing"

// Complete original DiskStateQueueTest: nine inherited StateQueueTest methods
// and its full 2,147,483,648-state growth method.
func TestJavaLongDiskStateQueue(t *testing.T) {
	for _, method := range javaStateQueueOriginalMethods {
		t.Run(method, func(t *testing.T) {
			queue := javaLongDiskStateQueueSetup(t)
			javaStateQueueOriginalMethod(t, queue, method)
		})
	}
	t.Run("testGrowBeyondIntMaxValue", func(t *testing.T) {
		queue := javaLongDiskStateQueueSetup(t)
		// DummyTLCState's default constructor sets uid=0, fp=0, Empty=this.
		// Its inherited write encodes only the source three-field header;
		// this method never dequeues or fingerprints it.
		state := &TLCStateMut{WorkerID: TLCStateInitWorkerID, UID: 0, level: TLCStateInitLevel}
		EmptyState = state
		const j = int64(1 << 31) // Integer.MAX_VALUE + 1L
		for i := int64(0); i < j; i++ {
			queue.SEnqueue(state)
		}
		if size := queue.Size(); size != j {
			t.Fatalf("size=%d, want %d", size, j)
		}
	})
}

func javaLongDiskStateQueueSetup(t *testing.T) *DiskStateQueue {
	t.Helper()
	oldWorkers, oldVariables, oldVarCount, oldEmpty := NumWorkers(), stateVariables, UniqueStringVariableCount(), EmptyState
	t.Cleanup(func() {
		SetNumWorkers(oldWorkers)
		stateVariables = oldVariables
		SetUniqueStringVariableCount(oldVarCount)
		EmptyState = oldEmpty
	})
	SetNumWorkers(1)
	// The source class's fresh JVM has no spec variables. DummyTLCState.read
	// consumes only the base header, including in the background pool reader.
	stateVariables = nil
	SetUniqueStringVariableCount(0)
	EmptyState = nil
	queue := NewDiskStateQueue(t.TempDir())
	// Stop native pool threads before removing files or restoring class statics.
	t.Cleanup(queue.FinishAll)
	return queue
}
