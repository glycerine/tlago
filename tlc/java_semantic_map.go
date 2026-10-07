// Copyright (c) 1997, 2023, Oracle and/or its affiliates. All rights reserved.
// Derived from OpenJDK 21 java.util.HashMap and HashSet; GPL v2 with Classpath exception.
// See licenses/openjdk-LICENSE and licenses/openjdk-ADDITIONAL_LICENSE_INFO.
package tlc

// JavaSemanticMap exposes the shared HashMap backing implementation to SANY.
// Callers supply the source hashCode, equals and non-Comparable tree tie-break.
// Values may be nullable; copy construction preserves them without calling a
// derived constraint map's tightening put method.
type JavaSemanticMap[K comparable, V any] struct {
	entries  *javaHashMap[K, V]
	modCount int32
}

func NewJavaSemanticMap[K comparable, V any](hash func(K) int32, equal func(K, K) bool, tieBreak func(K, K) int) *JavaSemanticMap[K, V] {
	m := newJavaHashMap[K, V](hash, nil)
	m.equal, m.tieBreak = equal, tieBreak
	return &JavaSemanticMap[K, V]{entries: m}
}

func NewJavaSemanticMapCopy[K comparable, V any](source *JavaSemanticMap[K, V], hash func(K) int32, equal func(K, K) bool, tieBreak func(K, K) int) *JavaSemanticMap[K, V] {
	if source == nil {
		panic(NewNullPointerException())
	}
	result := NewJavaSemanticMap[K, V](hash, equal, tieBreak)
	if size := source.Len(); size > 0 {
		// OpenJDK 21 HashMap.putMapEntries: ceil(size / 0.75), tableSizeFor.
		target := (int64(size)*4 + 2) / 3
		capacity := int64(1)
		for capacity < target && capacity < 1<<30 {
			capacity <<= 1
		}
		result.entries.threshold = int(capacity)
		for key, value := range source.All() {
			result.Put(key, value)
		}
	}
	return result
}

// NewJavaSemanticSetCopy supplies HashSet(Collection)'s backing HashMap.
// OpenJDK 21 reserves max(collection.size(), 12) mappings even for an empty set.
func NewJavaSemanticSetCopy[K comparable](source *JavaSemanticMap[K, struct{}], hash func(K) int32, equal func(K, K) bool, tieBreak func(K, K) int) *JavaSemanticMap[K, struct{}] {
	if source == nil {
		panic(NewNullPointerException())
	}
	result := NewJavaSemanticMap[K, struct{}](hash, equal, tieBreak)
	target := (int64(max(source.Len(), 12))*4 + 2) / 3
	capacity := int64(1)
	for capacity < target && capacity < 1<<30 {
		capacity <<= 1
	}
	result.entries.threshold = int(capacity)
	for key, value := range source.All() {
		result.Put(key, value)
	}
	return result
}

func (m *JavaSemanticMap[K, V]) Len() int {
	if m == nil {
		panic(NewNullPointerException())
	}
	return m.entries.Len()
}
func (m *JavaSemanticMap[K, V]) Get(key K) (V, bool) {
	if m == nil {
		panic(NewNullPointerException())
	}
	return m.entries.Get2(key)
}
func (m *JavaSemanticMap[K, V]) Put(key K, value V) (V, bool) {
	if m == nil {
		panic(NewNullPointerException())
	}
	old, present := m.entries.Get2(key)
	m.entries.Set(key, value)
	if !present {
		m.modCount++
	}
	return old, present
}
func (m *JavaSemanticMap[K, V]) Remove(key K) bool {
	if m == nil {
		panic(NewNullPointerException())
	}
	if m.entries.Remove(key) {
		m.modCount++
		return true
	}
	return false
}

// HashMap.clear increments modCount even when already empty, retaining capacity.
func (m *JavaSemanticMap[K, V]) Clear() {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.modCount++
	m.entries.size = 0
	clear(m.entries.table)
}
func (m *JavaSemanticMap[K, V]) All() func(func(K, V) bool) {
	if m == nil {
		panic(NewNullPointerException())
	}
	return func(yield func(K, V) bool) {
		expected := m.modCount
		index := 0
		var next *javaHashNode[K, V]
		advance := func() {
			for next == nil && index < len(m.entries.table) {
				next = m.entries.table[index]
				index++
			}
		}
		advance()
		for next != nil {
			// HashMap iterator checks on next(), then saves the following node
			// before returning the current entry to its caller.
			if expected != m.modCount {
				panic(NewConcurrentModificationException())
			}
			current := next
			next = current.next
			advance()
			if !yield(current.key, current.value) {
				return
			}
		}
	}
}
