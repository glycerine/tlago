package tlc

import (
	"encoding/gob"
	"testing"
)

func init() {
	gob.Register("")
	gob.Register(0)
}

func TestMemObjectStackPortedJavaLIFOBehaviors(t *testing.T) {
	stack := NewMemObjectStack(t.TempDir(), "stack")
	stack.Push("a")
	stack.Push("b")
	stack.SPush("c")

	if stack.Size() != 3 {
		t.Fatalf("size = %d, want 3", stack.Size())
	}
	if got := stack.Pop(); got != "c" {
		t.Fatalf("Pop = %v, want c", got)
	}

	stack.SPushAll([]any{"d", "e"})
	got := stack.SPopMany(3)
	requireAnySlice(t, got, []any{"e", "d", "b"})
	if got := stack.Pop(); got != "a" {
		t.Fatalf("final Pop = %v, want a", got)
	}
	if got := stack.Pop(); got != nil {
		t.Fatalf("empty Pop = %v, want nil", got)
	}
}

func TestMemObjectStackCheckpointRoundTrip(t *testing.T) {
	dir := t.TempDir()
	stack := NewMemObjectStack(dir, "stack")
	stack.Push("a")
	stack.Push("b")
	if err := stack.BeginChkpt(); err != nil {
		t.Fatalf("BeginChkpt: %v", err)
	}
	if err := stack.CommitChkpt(); err != nil {
		t.Fatalf("CommitChkpt: %v", err)
	}

	recovered := NewMemObjectStack(dir, "stack")
	if err := recovered.Recover(); err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if recovered.Size() != 2 {
		t.Fatalf("recovered size = %d, want 2", recovered.Size())
	}
	if got := recovered.Pop(); got != "b" {
		t.Fatalf("first recovered Pop = %v, want b", got)
	}
	if got := recovered.Pop(); got != "a" {
		t.Fatalf("second recovered Pop = %v, want a", got)
	}
}

func TestMemObjectQueuePortedJavaFIFOAndCheckpointBehaviors(t *testing.T) {
	dir := t.TempDir()
	queue := NewMemObjectQueue(dir)
	queue.Enqueue("a")
	queue.Enqueue("b")
	if got := queue.Dequeue(); got != "a" {
		t.Fatalf("Dequeue = %v, want a", got)
	}
	queue.Enqueue("c")

	if err := queue.BeginChkpt(); err != nil {
		t.Fatalf("BeginChkpt: %v", err)
	}
	if err := queue.CommitChkpt(); err != nil {
		t.Fatalf("CommitChkpt: %v", err)
	}

	recovered := NewMemObjectQueue(dir)
	if err := recovered.Recover(); err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if recovered.Size() != 2 {
		t.Fatalf("recovered size = %d, want 2", recovered.Size())
	}
	if got := recovered.Dequeue(); got != "b" {
		t.Fatalf("first recovered Dequeue = %v, want b", got)
	}
	if got := recovered.Dequeue(); got != "c" {
		t.Fatalf("second recovered Dequeue = %v, want c", got)
	}
	if got := recovered.Dequeue(); got != nil {
		t.Fatalf("empty Dequeue = %v, want nil", got)
	}
}

func TestDiskObjectStackBasicBufferLIFOBehavior(t *testing.T) {
	stack := NewDiskObjectStack(t.TempDir(), "stack")
	stack.Push(1)
	stack.Push(2)
	stack.SPush(3)

	if stack.Size() != 3 {
		t.Fatalf("size = %d, want 3", stack.Size())
	}
	if got := stack.SPop(); got != 3 {
		t.Fatalf("SPop = %v, want 3", got)
	}
	if got := stack.Pop(); got != 2 {
		t.Fatalf("Pop = %v, want 2", got)
	}
	if got := stack.Pop(); got != 1 {
		t.Fatalf("Pop = %v, want 1", got)
	}
}

func requireAnySlice(t *testing.T, got []any, want []any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("element %d = %v, want %v (slice %v)", i, got[i], want[i], got)
		}
	}
}
