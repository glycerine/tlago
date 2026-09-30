package tlc

import "strings"

type ListConsCell struct {
	value any
	next  *ListConsCell
}

type List struct {
	first *ListConsCell
	last  *ListConsCell
}

var EmptyList = &List{}

func NewList(values ...any) *List {
	if len(values) == 0 {
		return &List{}
	}
	l := &List{}
	for _, value := range values {
		l.Append1D(value)
	}
	return l
}

func NewListFromList(src *List) *List {
	if src == nil {
		return &List{}
	}
	return &List{first: src.first, last: src.last}
}

func (l *List) IsEmpty() bool {
	return l == nil || l.first == nil
}

func (l *List) Length() int {
	length := 0
	for cell := l.first; cell != nil; cell = cell.next {
		length++
	}
	return length
}

func (l *List) Push(value any) {
	cell := &ListConsCell{value: value, next: l.first}
	l.first = cell
	if l.last == nil {
		l.last = cell
	}
}

func (l *List) Pop() any {
	result := l.first.value
	l.first = l.first.next
	if l.first == nil {
		l.last = nil
	}
	return result
}

func (l *List) Cons(value any) *List {
	cell := &ListConsCell{value: value, next: l.first}
	return &List{first: cell, last: chooseListLast(l.last, cell)}
}

func (l *List) Car() any {
	return l.first.value
}

func (l *List) Cdr() *List {
	out := &List{first: l.first.next}
	if out.first != nil {
		out.last = l.last
	}
	return out
}

func (l *List) Member(value any) bool {
	for cell := l.first; cell != nil; cell = cell.next {
		if cell.value == value {
			return true
		}
	}
	return false
}

func (l *List) Append1(value any) *List {
	if l == nil || l.first == nil {
		return NewList(value)
	}
	out := NewList(l.first.value)
	for cell := l.first.next; cell != nil; cell = cell.next {
		newCell := &ListConsCell{value: cell.value}
		out.last.next = newCell
		out.last = newCell
	}
	newCell := &ListConsCell{value: value}
	out.last.next = newCell
	out.last = newCell
	return out
}

func (l *List) Append(other *List) *List {
	if l == nil || l.first == nil {
		return NewListFromList(other)
	}
	out := NewList(l.first.value)
	for cell := l.first.next; cell != nil; cell = cell.next {
		newCell := &ListConsCell{value: cell.value}
		out.last.next = newCell
		out.last = newCell
	}
	if other != nil {
		for cell := other.first; cell != nil; cell = cell.next {
			newCell := &ListConsCell{value: cell.value}
			out.last.next = newCell
			out.last = newCell
		}
	}
	return out
}

func (l *List) Append1D(value any) *List {
	if l.first == nil {
		l.first = &ListConsCell{value: value}
		l.last = l.first
	} else {
		l.last.next = &ListConsCell{value: value}
		l.last = l.last.next
	}
	return l
}

func (l *List) AppendD(other *List) *List {
	if other == nil || other.first == nil {
		return l
	}
	if l.first == nil {
		l.first = other.first
	} else {
		l.last.next = other.first
	}
	l.last = other.last
	return l
}

func (l *List) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for cell := l.first; cell != nil; cell = cell.next {
		b.WriteString(toContextString(cell.value))
	}
	b.WriteByte(']')
	return b.String()
}

func chooseListLast(last *ListConsCell, fallback *ListConsCell) *ListConsCell {
	if last == nil {
		return fallback
	}
	return last
}
