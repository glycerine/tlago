package tlc

import (
	"strconv"
	"sync"
)

// InternSource obtains a coordinator token for an absent worker string.
// Cached strings never call the source again.
type InternSource interface {
	Intern(string) (*UniqueString, error)
}

// InternTable preserves Java's linear probing, growth, and slot iteration.
// Its native mutex serializes token allocation within a worker context.
type InternTable struct {
	mu           sync.Mutex
	dataMu       sync.RWMutex
	table        []*UniqueString
	count        int
	thresh       int
	tokenCnt     int32
	varCount     int
	internSource InternSource
}

func NewInternTable(size int) *InternTable {
	size = int(int32(size))
	if size < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(size)))
	}
	return &InternTable{table: make([]*UniqueString, size), thresh: size / 2}
}

func (t *InternTable) grow() {
	old := t.table
	length := int(int32(2*len(old) + 1))
	if length < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(length)))
	}
	t.count = 0
	t.table = make([]*UniqueString, length)
	t.thresh = length / 2
	for _, value := range old {
		if value != nil {
			t.putValue(value)
		}
	}
}

func (t *InternTable) putValue(value *UniqueString) {
	if t.count >= t.thresh {
		t.grow()
	}
	loc := int(uint32(javaStringHashCode(value.s))&0x7fffffff) % len(t.table)
	for t.table[loc] != nil {
		loc = (loc + 1) % len(t.table)
	}
	t.table[loc] = value
	t.count++
}

func (t *InternTable) create(str string) (value *UniqueString) {
	if t.internSource == nil {
		t.dataMu.Lock()
		defer t.dataMu.Unlock()
		t.tokenCnt++
		return &UniqueString{s: str, tok: int(t.tokenCnt), loc: -1}
	}
	defer func() {
		if failure := recover(); failure != nil {
			err := panicValueAsError(failure)
			if javaSystemFailureCode(err) != NoError {
				panic(failure) // Java catches Exception, not Error.
			}
			failure := newTLCError(ECGeneral, "Failed to intern %s.", str)
			failure.Runtime = true
			panic(failure)
		}
	}()
	value, err := t.internSource.Intern(str)
	if err != nil {
		panic(err)
	}
	return value // A null response is stored and increments count in Java too.
}

func (t *InternTable) Put(str string) *UniqueString {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dataMu.Lock()
	dataLocked := true
	defer func() {
		if dataLocked {
			t.dataMu.Unlock()
		}
	}()
	if t.count >= t.thresh {
		t.grow()
	}
	loc := int(uint32(javaStringHashCode(str))&0x7fffffff) % len(t.table)
	for {
		ent := t.table[loc]
		if ent == nil {
			t.dataMu.Unlock()
			dataLocked = false
			value := t.create(str)
			t.dataMu.Lock()
			dataLocked = true
			t.table[loc] = value
			t.count++
			return value
		}
		if ent.s == str {
			return ent
		}
		loc = (loc + 1) % len(t.table)
	}
}

func (t *InternTable) SetSource(source InternSource) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.internSource = source
}

func (t *InternTable) Get(id int) *UniqueString {
	for i := 0; ; i++ {
		value, present := t.slotAt(i)
		if !present {
			break
		}
		if value != nil && value.tok == id {
			return value
		}
	}
	return nil
}

func (t *InternTable) Find(str string) *UniqueString {
	for i := 0; ; i++ {
		value, present := t.slotAt(i)
		if !present {
			break
		}
		if value != nil && value.s == str {
			return value
		}
	}
	return nil
}

func (t *InternTable) ToMap() *InsMap[string, *UniqueString] {
	values := NewInsMap[string, *UniqueString]()
	for i := 0; ; i++ {
		value, present := t.slotAt(i)
		if !present {
			break
		}
		if value != nil {
			values.Set(value.s, value)
		}
	}
	return values
}

// Java's readers scan the live table without acquiring the put monitor.
// Protect individual Go array accesses without turning these scans into a
// frozen snapshot or blocking them across a remote create call.
func (t *InternTable) slotAt(index int) (*UniqueString, bool) {
	t.dataMu.RLock()
	defer t.dataMu.RUnlock()
	if index >= len(t.table) {
		return nil, false
	}
	return t.table[index], true
}

func SetUniqueStringSource(source InternSource) { internTable.SetSource(source) }

// UniqueStringInitializeWithSource starts a fresh worker interning context.
// Java class-static names are initialized on first use in the new context.
// Call this during bootstrap, before constructing the worker's Tool or values.
func UniqueStringInitializeWithSource(source InternSource) {
	internTable = NewInternTable(1024)
	internTable.SetSource(source)
	resetBuiltInOPNames()
}

func (s *TLCServer) Intern(str string) (*UniqueString, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s.serverInternTable().Put(str), nil
}

func (s *TLCServer) serverInternTable() *InternTable {
	s.internMu.Lock()
	defer s.internMu.Unlock()
	if s.InternTable == nil {
		s.InternTable = internTable
	}
	return s.InternTable
}
