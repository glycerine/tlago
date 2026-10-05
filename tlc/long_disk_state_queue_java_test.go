package tlc

import "testing"

// All nine original inherited StateQueueTest methods with DiskStateQueueTest's
// concrete disk-queue setup. Its 2,147,483,648-state growth method remains pending.
func TestJavaLongDiskStateQueue(t *testing.T) {
	for _, method := range javaStateQueueOriginalMethods {
		t.Run(method, func(t *testing.T) {
			oldWorkers := NumWorkers()
			SetNumWorkers(1)
			t.Cleanup(func() { SetNumWorkers(oldWorkers) })
			queue := NewDiskStateQueue(t.TempDir())
			// Native equivalent of the source fork's process teardown for its
			// background pool threads, before removing the temporary directory.
			t.Cleanup(queue.FinishAll)
			javaStateQueueOriginalMethod(t, queue, method)
		})
	}
}
