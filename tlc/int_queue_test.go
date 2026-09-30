package tlc

import "testing"

func TestIntQueueDequeuePastLastElementPanics(t *testing.T) {
	queue := NewIntQueue()
	queue.EnqueueInt(1)

	if got := queue.DequeueInt(); got != 1 {
		t.Fatalf("DequeueInt = %d, want 1", got)
	}
	expectPanic(t, func() { queue.DequeueInt() })
}

func TestIntQueueEnqueuesZerosAsRealValues(t *testing.T) {
	queue := NewIntQueue()
	queue.EnqueueInt(0)
	queue.EnqueueInt(0)
	queue.EnqueueInt(0)

	for i := 0; i < 3; i++ {
		if got := queue.DequeueInt(); got != 0 {
			t.Fatalf("zero dequeue %d = %d, want 0", i, got)
		}
	}
	expectPanic(t, func() { queue.DequeueInt() })
}

func TestIntQueueEnqueueLongStoresHighThenLowInts(t *testing.T) {
	queue := NewIntQueue()
	queue.EnqueueLong(0)

	if got := queue.DequeueInt(); got != 0 {
		t.Fatalf("high int = %d, want 0", got)
	}
	if got := queue.DequeueInt(); got != 0 {
		t.Fatalf("low int = %d, want 0", got)
	}
	expectPanic(t, func() { queue.DequeueInt() })
}

func TestIntQueueEnqueueDequeueLong(t *testing.T) {
	queue := NewIntQueue()
	queue.EnqueueLong(0x0102030405060708)

	if got := queue.DequeueLong(); got != 0x0102030405060708 {
		t.Fatalf("DequeueLong = %#x, want %#x", got, int64(0x0102030405060708))
	}
	expectPanic(t, func() { queue.DequeueLong() })
}

func TestIntQueueRingWrapPreservesFIFOOrder(t *testing.T) {
	queue := NewIntQueueWithCapacity(4)
	queue.EnqueueInt(0)
	queue.EnqueueInt(1)
	queue.EnqueueInt(2)
	queue.EnqueueInt(3)
	if queue.Size() != 4 {
		t.Fatalf("size = %d, want 4", queue.Size())
	}

	for i := int32(0); i < 4; i++ {
		if got := queue.DequeueInt(); got != i {
			t.Fatalf("initial dequeue = %d, want %d", got, i)
		}
	}
	if queue.Size() != 0 {
		t.Fatalf("size after draining = %d, want 0", queue.Size())
	}

	queue.EnqueueInt(4)
	queue.EnqueueInt(5)
	queue.EnqueueInt(6)
	queue.EnqueueInt(7)
	if queue.Size() != 4 {
		t.Fatalf("size after refill = %d, want 4", queue.Size())
	}
	if got := queue.DequeueInt(); got != 4 {
		t.Fatalf("head dequeue = %d, want 4", got)
	}

	queue.EnqueueInt(8)
	queue.EnqueueInt(9)
	for want := int32(5); want < 10; want++ {
		if got := queue.DequeueInt(); got != want {
			t.Fatalf("post-compact dequeue = %d, want %d", got, want)
		}
	}
	if queue.Size() != 0 {
		t.Fatalf("size after compact drain = %d, want 0", queue.Size())
	}
}
