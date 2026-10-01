package tlc

import (
	"strconv"
	"strings"
)

type SetOfStates struct {
	states        []*TLCStateMut
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
		length = 0
	}
	return &SetOfStates{
		states: make([]*TLCStateMut, length),
		length: length,
		thresh: length / 2,
	}
}

func NewSetOfStatesFromVec(vec *StateVec) *SetOfStates {
	if vec == nil {
		return NewSetOfStates(0)
	}
	set := NewSetOfStates(vec.Size())
	for i := 0; i < vec.Size(); i++ {
		set.Put(vec.At(i))
	}
	return set
}

func (s *SetOfStates) Clear() {
	s.count = 0
	s.iteratorIndex = 0
	s.states = make([]*TLCStateMut, s.length)
}

func (s *SetOfStates) Put(state *TLCStateMut) bool {
	return s.PutFP(state.FingerPrint(), state)
}

func (s *SetOfStates) PutFP(fingerprint uint64, state *TLCStateMut) bool {
	if s.count >= s.thresh {
		s.grow()
	}
	return s.put0(fingerprint, state)
}

func (s *SetOfStates) put0(fingerprint uint64, state *TLCStateMut) bool {
	loc := int(fingerprint&0x7fffffff) % s.length
	for {
		ent := s.states[loc]
		if ent == nil {
			s.states[loc] = state
			s.count++
			return false
		}
		if state.Equal(ent) {
			return true
		}
		loc = (loc + 1) % s.length
	}
}

func (s *SetOfStates) grow() {
	old := s.states
	s.count = 0
	s.length = 2*s.length + 1
	if s.length < 1 {
		s.length = 1
	}
	s.thresh = s.length / 2
	s.states = make([]*TLCStateMut, s.length)
	s.iteratorIndex = 0
	for _, state := range old {
		if state != nil {
			s.PutFP(state.FingerPrint(), state)
		}
	}
}

func (s *SetOfStates) Capacity() int {
	return s.length
}

func (s *SetOfStates) Size() int {
	return s.count
}

func (s *SetOfStates) Next() *TLCStateMut {
	for {
		state := s.states[s.iteratorIndex]
		s.iteratorIndex++
		if state != nil {
			return state
		}
	}
}

func (s *SetOfStates) ResetNext() {
	s.iteratorIndex = 0
}

func (s *SetOfStates) ToSlice() []*TLCStateMut {
	out := make([]*TLCStateMut, 0, s.count)
	for _, state := range s.states {
		if state != nil {
			out = append(out, state)
		}
	}
	return out
}

func (s *SetOfStates) GetSubSet(action *Action) []*TLCStateMut {
	subset := make([]*TLCStateMut, 0)
	for i := 0; i < s.Size(); i++ {
		next := s.Next()
		if next != nil && action == next.GetAction() {
			subset = append(subset, next)
		}
	}
	s.ResetNext()
	return subset
}

func (s *SetOfStates) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for _, state := range s.states {
		if state == nil {
			continue
		}
		b.WriteString("<<")
		b.WriteString(strconv.FormatUint(state.FingerPrint(), 10))
		b.WriteByte(',')
		b.WriteString(strings.TrimSuffix(state.String(), "\n"))
		b.WriteString(">>,\n")
	}
	b.WriteByte('}')
	return b.String()
}
