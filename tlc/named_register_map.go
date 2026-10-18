package tlc

import (
	"iter"
	"sync"
)

// NamedRegisterMap owns an IdThread's named registers. Its returned aliases share
// storage and synchronization, including when callers bypass worker accessors.
type NamedRegisterMap struct {
	mu     sync.RWMutex
	values map[*UniqueString]Value
}

func NewNamedRegisterMap() *NamedRegisterMap {
	return &NamedRegisterMap{values: make(map[*UniqueString]Value)}
}

func (m *NamedRegisterMap) Get(key *UniqueString) Value {
	value, _ := m.Get2(key)
	return value
}

func (m *NamedRegisterMap) Get2(key *UniqueString) (Value, bool) {
	if m == nil || key == nil {
		panic(NewNullPointerException())
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, found := m.values[key]
	return value, found
}

// Put is Map.put: it returns the previous value, or nil for an absent key.
func (m *NamedRegisterMap) Put(key *UniqueString, value Value) Value {
	if m == nil || key == nil || isNil(value) {
		panic(NewNullPointerException())
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	previous := m.values[key]
	if m.values == nil {
		m.values = make(map[*UniqueString]Value)
	}
	m.values[key] = value
	return previous
}

// Set retains the native dictionary spelling and reports a newly inserted key.
func (m *NamedRegisterMap) Set(key *UniqueString, value Value) bool {
	return m.Put(key, value) == nil
}

func (m *NamedRegisterMap) PutIfAbsent(key *UniqueString, value Value) Value {
	if m == nil || key == nil || isNil(value) {
		panic(NewNullPointerException())
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if previous, found := m.values[key]; found {
		return previous
	}
	if m.values == nil {
		m.values = make(map[*UniqueString]Value)
	}
	m.values[key] = value
	return nil
}

func (m *NamedRegisterMap) ContainsKey(key *UniqueString) bool {
	_, found := m.Get2(key)
	return found
}

func (m *NamedRegisterMap) Remove(key *UniqueString) Value {
	if m == nil || key == nil {
		panic(NewNullPointerException())
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	previous := m.values[key]
	delete(m.values, key)
	return previous
}

func (m *NamedRegisterMap) Clear() {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.values)
}

func (m *NamedRegisterMap) Len() int {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.values)
}

// All captures entries when iteration begins and releases the lock before
// yielding. Callbacks can access the same live map without deadlocking. Order
// and observation of concurrent mutations are unspecified.
func (m *NamedRegisterMap) All() iter.Seq2[*UniqueString, Value] {
	return func(yield func(*UniqueString, Value) bool) {
		if m == nil {
			panic(NewNullPointerException())
		}
		m.mu.RLock()
		snapshot := make(map[*UniqueString]Value, len(m.values))
		for key, value := range m.values {
			snapshot[key] = value
		}
		m.mu.RUnlock()
		for key, value := range snapshot {
			if !yield(key, value) {
				return
			}
		}
	}
}
