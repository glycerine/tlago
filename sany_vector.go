package tlago

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type SanyVector[T comparable] struct {
	items []T
}

func NewSanyVector[T comparable](initialSizes ...int) *SanyVector[T] {
	if len(initialSizes) > 1 {
		panic(tlc.NewIllegalArgumentException("NewSanyVector accepts at most one initial size"))
	}
	initialSize := 10
	if len(initialSizes) == 1 {
		initialSize = initialSizes[0]
	}
	if initialSize < 0 {
		panic(tlc.NewIllegalArgumentException(fmt.Sprintf("Illegal Capacity: %d", initialSize)))
	}
	return &SanyVector[T]{items: make([]T, 0, initialSize)}
}

func (v *SanyVector[T]) require() {
	if v == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (v *SanyVector[T]) AddElement(item T) {
	v.require()
	v.items = append(v.items, item)
}

func (v *SanyVector[T]) Contains(item T) bool {
	v.require()
	for _, existing := range v.items {
		if existing == item {
			return true
		}
	}
	return false
}

func (v *SanyVector[T]) InsertElementAt(item T, index int) {
	v.require()
	if index < 0 || index >= len(v.items) {
		panic(tlc.NewArrayIndexOutOfBoundsExceptionNoMessage())
	}
	var zero T
	v.items = append(v.items, zero)
	copy(v.items[index+1:], v.items[index:])
	v.items[index] = item
}

func (v *SanyVector[T]) Elements() []T {
	v.require()
	out := make([]T, len(v.items))
	copy(out, v.items)
	return out
}

func (v *SanyVector[T]) AppendNoRepeats(incoming *SanyVector[T]) {
	v.require()
	incoming.require()
	for _, item := range incoming.items {
		if !v.Contains(item) {
			v.AddElement(item)
		}
	}
}

func (v *SanyVector[T]) Size() int {
	v.require()
	return len(v.items)
}

func (v *SanyVector[T]) ElementAt(index int) T {
	v.require()
	v.checkBounds(index)
	return v.items[index]
}

func (v *SanyVector[T]) FirstElement() T {
	v.require()
	return v.ElementAt(0)
}

func (v *SanyVector[T]) LastElement() T {
	v.require()
	return v.ElementAt(len(v.items) - 1)
}

func (v *SanyVector[T]) RemoveAllElements() {
	v.require()
	v.items = v.items[:0]
}

func (v *SanyVector[T]) RemoveElementAt(index int) {
	v.require()
	v.checkBounds(index)
	copy(v.items[index:], v.items[index+1:])
	var zero T
	v.items[len(v.items)-1] = zero
	v.items = v.items[:len(v.items)-1]
}

func (v *SanyVector[T]) SetElementAt(item T, index int) {
	v.require()
	v.checkBounds(index)
	v.items[index] = item
}

func (v *SanyVector[T]) Append(incoming *SanyVector[T]) {
	v.require()
	incoming.require()
	v.items = append(v.items, incoming.items...)
}

func (v *SanyVector[T]) String() string {
	v.require()
	var result strings.Builder
	result.WriteString("[ ")
	for index, item := range v.items {
		if index != 0 {
			result.WriteString(", ")
		}
		value := reflect.ValueOf(item)
		if !value.IsValid() || (value.Kind() == reflect.Chan || value.Kind() == reflect.Func || value.Kind() == reflect.Interface || value.Kind() == reflect.Map || value.Kind() == reflect.Pointer || value.Kind() == reflect.Slice) && value.IsNil() {
			panic(tlc.NewNullPointerException())
		}
		result.WriteString(fmt.Sprint(item))
	}
	result.WriteString(" ]")
	return result.String()
}

func (v *SanyVector[T]) checkBounds(index int) {
	if index < 0 || index >= len(v.items) {
		panic(tlc.NewArrayIndexOutOfBoundsExceptionNoMessage())
	}
}
