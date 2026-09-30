package tlc

import "testing"

func TestIntStackPeakMatchesJavaCellPositions(t *testing.T) {
	stack := NewIntStack()

	stack.PushLong(4711)
	stack.PushLong(2323)
	stack.PushInt(1)
	stack.PushLong(77)

	if got := stack.PeekLongAt(0); got != 4711 {
		t.Fatalf("PeekLongAt(0) = %d, want 4711", got)
	}
	if got := stack.PeekLongAt(2); got != 2323 {
		t.Fatalf("PeekLongAt(2) = %d, want 2323", got)
	}
	if got := stack.PeekIntAt(4); got != 1 {
		t.Fatalf("PeekIntAt(4) = %d, want 1", got)
	}
	if got := stack.PeekLongAt(5); got != 77 {
		t.Fatalf("PeekLongAt(5) = %d, want 77", got)
	}
}

func TestIntStackPushPopLongUsesJavaLowThenHighCellOrder(t *testing.T) {
	stack := NewIntStack()
	stack.PushLong(0x0102030405060708)

	if got := stack.PeekIntAt(0); got != 0x05060708 {
		t.Fatalf("low cell = %#x, want %#x", got, int32(0x05060708))
	}
	if got := stack.PeekIntAt(1); got != 0x01020304 {
		t.Fatalf("high cell = %#x, want %#x", got, int32(0x01020304))
	}
	if got := stack.PopLong(); got != 0x0102030405060708 {
		t.Fatalf("PopLong = %#x, want %#x", got, int64(0x0102030405060708))
	}
	expectPanic(t, func() { stack.PopInt() })
}

func TestIntStackResetClearsElements(t *testing.T) {
	stack := NewIntStack()
	stack.PushInt(1)
	stack.PushInt(2)
	stack.Reset()

	if stack.Size() != 0 {
		t.Fatalf("size after Reset = %d, want 0", stack.Size())
	}
	expectPanic(t, func() { stack.PopInt() })
}

func TestSynchronousDiskIntStackPushIntNoWrite(t *testing.T) {
	const size = 8
	stack := NewSynchronousDiskIntStack(t.TempDir(), "SynchronousDiskIntStackTest", size)
	for i := 0; i < size; i++ {
		stack.PushInt(int32(i))
	}
	for i := size - 1; i >= 0; i-- {
		if got := stack.PopInt(); got != int32(i) {
			t.Fatalf("PopInt = %d, want %d", got, i)
		}
	}
}

func TestSynchronousDiskIntStackPushIntWrite(t *testing.T) {
	const size = 8
	stack := NewSynchronousDiskIntStack(t.TempDir(), "SynchronousDiskIntStackTest", size)
	for i := 0; i < size*3; i++ {
		stack.PushInt(int32(i))
	}
	for i := size*3 - 1; i >= 0; i-- {
		if got := stack.PopInt(); got != int32(i) {
			t.Fatalf("PopInt = %d, want %d", got, i)
		}
	}
	if got := stack.Size(); got != 0 {
		t.Fatalf("Size = %d, want 0", got)
	}
}
