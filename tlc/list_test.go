package tlc

import "testing"

func TestListPortedJavaCoreBehaviors(t *testing.T) {
	a := &struct{ name string }{"a"}
	b := &struct{ name string }{"b"}
	c := &struct{ name string }{"c"}

	l := NewList(a)
	if l.IsEmpty() || l.Length() != 1 || l.Car() != a {
		t.Fatalf("single-element list shape mismatch")
	}

	l.Push(b)
	if l.Car() != b || l.Length() != 2 {
		t.Fatalf("Push should destructively add at the front")
	}
	if got := l.Pop(); got != b {
		t.Fatalf("Pop = %v, want front element", got)
	}
	if l.Car() != a || l.Length() != 1 {
		t.Fatalf("Pop should destructively remove the front")
	}

	consed := l.Cons(b)
	if consed.Car() != b || l.Car() != a {
		t.Fatalf("Cons should return a new front without changing source")
	}
	if consed.Cdr().Car() != a {
		t.Fatalf("Cdr should expose the tail of the consed list")
	}

	appended := l.Append1(c)
	if appended.Length() != 2 || appended.Pop() != a || appended.Pop() != c {
		t.Fatalf("Append1 should return a new list with value at tail")
	}
	if l.Length() != 1 || l.Car() != a {
		t.Fatalf("Append1 should not mutate source list")
	}
}

func TestListAppendAndMemberUseJavaReferenceSemantics(t *testing.T) {
	a := &struct{ name string }{"a"}
	b := &struct{ name string }{"b"}
	c := &struct{ name string }{"c"}
	sameShapeAsA := &struct{ name string }{"a"}

	left := NewList(a, b)
	right := NewList(c)
	joined := left.Append(right)
	if joined.Length() != 3 {
		t.Fatalf("joined length = %d, want 3", joined.Length())
	}
	if !joined.Member(a) || joined.Member(sameShapeAsA) {
		t.Fatalf("Member should use object identity-like equality")
	}

	left.AppendD(right)
	if left.Length() != 3 || left.Pop() != a || left.Pop() != b || left.Pop() != c {
		t.Fatalf("AppendD should destructively append the second list")
	}
}
