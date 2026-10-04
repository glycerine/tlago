package tlc

import (
	"strconv"
	"strings"
)

type SetOfStates struct {
	states        []TLCState
	count         int
	length        int
	thresh        int
	iteratorIndex int
}

func NewSetOfStates(size ...int) *SetOfStates {
	length := 16
	if len(size) > 0 {
		length = size[0]
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(length)))
	}
	return &SetOfStates{
		states: make([]TLCState, length),
		length: length,
		thresh: length / 2,
	}
}

func NewSetOfStatesFromVec(vec *StateVec, tools ...*Tool) *SetOfStates {
	if vec == nil {
		panic(NewNullPointerException())
	}
	set := NewSetOfStates(vec.Size())
	for i := 0; i < vec.Size(); i++ {
		set.Put(vec.At(i), tools...)
	}
	return set
}

func (s *SetOfStates) Clear() {
	s.count = 0
	s.states = make([]TLCState, s.length)
}

func (s *SetOfStates) Put(state TLCState, tools ...*Tool) bool {
	if len(tools) != 0 {
		return s.PutFP(state.FingerPrintWithTool(tools[0]), state, tools[0])
	}
	return s.PutFP(state.FingerPrint(), state)
}

func (s *SetOfStates) PutFP(fingerprint uint64, state TLCState, tools ...*Tool) bool {
	if s.count >= s.thresh {
		s.grow(tools...)
	}
	return s.put0(fingerprint, state)
}

func (s *SetOfStates) put0(fingerprint uint64, state TLCState) bool {
	loc := int(fingerprint&0x7fffffff) % s.length
	for {
		ent := s.states[loc]
		if ent == nil {
			s.states[loc] = state
			s.count++
			return false
		}
		if setOfStatesEqual(state, ent) {
			return true
		}
		loc = (loc + 1) % s.length
	}
}

func (s *SetOfStates) grow(tools ...*Tool) {
	old := s.states
	s.count = 0
	s.length = int(int32(2*int32(s.length) + 1))
	if s.length < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(s.length)))
	}
	s.thresh = s.length / 2
	s.states = make([]TLCState, s.length)
	for _, state := range old {
		if state != nil {
			s.Put(state, tools...)
		}
	}
}

func (s *SetOfStates) Capacity() int {
	return s.length
}

func (s *SetOfStates) Size() int {
	return s.count
}

func (s *SetOfStates) Next() TLCState {
	for {
		index := s.iteratorIndex
		s.iteratorIndex++
		if index < 0 || index >= len(s.states) {
			panic(NewArrayIndexOutOfBoundsException(index, len(s.states)))
		}
		state := s.states[index]
		if state != nil {
			return state
		}
	}
}

func (s *SetOfStates) ResetNext() {
	s.iteratorIndex = 0
}

// ToSlice is the existing concrete evaluator adapter. General collection
// callers use Next or ToSet; a nonmutable state cannot enter this adapter.
func (s *SetOfStates) ToSlice() []*TLCStateMut {
	out := make([]*TLCStateMut, 0, s.count)
	for _, state := range s.states {
		if state != nil {
			out = append(out, state.(*TLCStateMut))
		}
	}
	return out
}

// GetSubSet follows Java's action-identity filtering and HashSet semantics.
func (s *SetOfStates) GetSubSet(action *Action) []TLCState {
	subset := NewTLCStateSet()
	for i := 0; i < s.Size(); i++ {
		next := s.Next()
		if action == next.GetAction() {
			subset.Add(next)
		}
	}
	s.ResetNext()
	return subset.ToSlice()
}

func (s *SetOfStates) ToSet() *TLCStateSet {
	result := NewTLCStateSet()
	for i := 0; i < s.Size(); i++ {
		result.Add(s.Next())
	}
	s.ResetNext()
	return result
}

func (s *SetOfStates) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for _, state := range s.states {
		if state == nil {
			continue
		}
		b.WriteString("<<")
		b.WriteString(strconv.FormatInt(int64(state.FingerPrint()), 10))
		b.WriteByte(',')
		units := javaStringUTF16(state.String())
		if len(units) == 0 {
			panic(NewStringIndexOutOfBoundsException(-1, 0))
		}
		b.WriteString(javaStringFromUTF16(units[:len(units)-1]))
		b.WriteString(">>,\n")
	}
	b.WriteByte('}')
	return b.String()
}

// Java's equals catch permits incomparable value types only at the state-set
// boundary. Other exception types propagate; the source assertion constrains
// the caught runtime exception's message.
func setOfStatesEqual(state, entry TLCState) (equal bool) {
	defer func() {
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); ok {
				if runtime := javaRuntimeException(err); runtime != nil {
					message := runtime.GetMessage()
					if message == nil || !(strings.HasPrefix(*message, "Attempted to check equality of") || strings.HasPrefix(*message, "Attempted to compare equality of")) {
						panic(newTLCError(ECTLCBug, ""))
					}
					equal = false
					return
				}
			}
			panic(failure)
		}
	}()
	return state.Equal(entry)
}

// TLCStateSet retains Java HashSet's per-object hashCode and equals contract,
// including TLCStateMut's source identity hashCode despite structural equals.
type TLCStateSet struct {
	buckets map[int32][]TLCState
	size    int
}

func NewTLCStateSet() *TLCStateSet { return &TLCStateSet{buckets: make(map[int32][]TLCState)} }
func (s *TLCStateSet) Size() int   { return s.size }
func (s *TLCStateSet) Add(state TLCState) bool {
	hash := state.HashCode()
	for _, entry := range s.buckets[hash] {
		if state == entry || state.Equal(entry) {
			return false
		}
	}
	s.buckets[hash] = append(s.buckets[hash], state)
	s.size++
	return true
}
func (s *TLCStateSet) ToSlice() []TLCState {
	states := make([]TLCState, 0, s.size)
	for _, bucket := range s.buckets {
		states = append(states, bucket...)
	}
	return states
}
