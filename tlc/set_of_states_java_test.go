/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/

package tlc

import (
	"fmt"
	"reflect"
	"testing"
)

// The original queue.DummyTLCState's state-set surface. Its fingerprint is a
// supplied long and its default Object equality/hash remain object identity.
type javaDummyTLCState struct {
	fp  uint64
	uid int64
}

func newJavaDummyTLCState(fp int64) *javaDummyTLCState {
	return &javaDummyTLCState{fp: uint64(fp), uid: TLCStateInitUID}
}
func (s *javaDummyTLCState) FingerPrint() uint64                { return s.fp }
func (s *javaDummyTLCState) FingerPrintWithTool(_ *Tool) uint64 { return s.FingerPrint() }
func (s *javaDummyTLCState) Equal(other TLCState) bool          { return s == other }
func (s *javaDummyTLCState) HashCode() int32                    { return int32(reflect.ValueOf(s).Pointer()) }
func (s *javaDummyTLCState) GetAction() *Action                 { return nil }
func (s *javaDummyTLCState) String() string                     { return fmt.Sprintf("Dummy#%d:%d", s.uid, int64(s.fp)) }

// Original EqualityDummyTLCState, including its full class/fp/id equality and
// 32-bit hashCode calculation. Embedded methods retain DummyTLCState behavior.
type javaEqualityDummyTLCState struct {
	*javaDummyTLCState
	id int32
}

func newJavaEqualityDummyTLCState(fp, id int32) *javaEqualityDummyTLCState {
	return &javaEqualityDummyTLCState{javaDummyTLCState: newJavaDummyTLCState(int64(fp)), id: id}
}
func (s *javaEqualityDummyTLCState) HashCode() int32 {
	result := int32(1)
	result = 31*result + s.id
	return int32(int64(31*result) + int64(s.FingerPrint()))
}
func (s *javaEqualityDummyTLCState) Equal(obj TLCState) bool {
	if s == obj {
		return true
	}
	if obj == nil {
		return false
	}
	other, ok := obj.(*javaEqualityDummyTLCState)
	if !ok {
		return false
	}
	if s.FingerPrint() != other.FingerPrint() {
		return false
	}
	if s.id != other.id {
		return false
	}
	return true
}

func TestJavaSetOfStatesSizeEmpty(t *testing.T) {
	s := NewSetOfStates(16)
	if s.Capacity() != 16 {
		t.Fatalf("capacity=%d, want16", s.Capacity())
	}
	if s.Size() != 0 {
		t.Fatalf("size=%d, want0", s.Size())
	}
}
func TestJavaSetOfStatesSize(t *testing.T) {
	s := NewSetOfStates(16)
	s.Put(newJavaDummyTLCState(1))
	if s.Capacity() != 16 {
		t.Fatalf("capacity=%d, want16", s.Capacity())
	}
	if s.Size() != 1 {
		t.Fatalf("size=%d, want1", s.Size())
	}
}
func TestJavaSetOfStatesGrow(t *testing.T) {
	s := NewSetOfStates(1)
	for i := 0; i < 32; i++ {
		s.Put(newJavaDummyTLCState(int64(i)))
	}
	if s.Capacity() <= 32 {
		t.Fatalf("capacity=%d, want>32", s.Capacity())
	}
	if s.Size() != 32 {
		t.Fatalf("size=%d, want32", s.Size())
	}
}
func TestJavaSetOfStatesIterate(t *testing.T) {
	s := NewSetOfStates(1)
	for i := 1; i <= 32; i++ {
		if s.Put(newJavaDummyTLCState(int64(i))) {
			t.Fatalf("insert%d reported duplicate", i)
		}
	}
	if s.Size() != 32 {
		t.Fatalf("size=%d, want32", s.Size())
	}
	var predecessor TLCState
	for i := 0; i < s.Size(); i++ {
		state := s.Next()
		if predecessor == state {
			t.Fatal("successor is same object as predecessor")
		}
		predecessor = state
	}
	s.ResetNext()
	var sum int64
	for i := 0; i < s.Size(); i++ {
		sum += int64(s.Next().FingerPrint())
	}
	if sum != (32/2)*(1+32) {
		t.Fatalf("fingerprint sum=%d, want528", sum)
	}
}
func TestJavaSetOfStatesDuplicates(t *testing.T) {
	s := NewSetOfStates(1)
	for i := 1; i <= 32; i++ {
		if s.Put(newJavaDummyTLCState(int64(i))) {
			t.Fatalf("insert%d reported duplicate", i)
		}
	}
	if s.Size() != 32 {
		t.Fatalf("size=%d, want32", s.Size())
	}
	states := NewTLCStateSet()
	for i := 0; i < s.Size(); i++ {
		states.Add(s.Next())
	}
	for _, state := range states.ToSlice() {
		if !s.Put(state) {
			t.Fatal("same object not reported duplicate")
		}
	}
	if states.Size() != 32 {
		t.Fatalf("HashSet size=%d, want32", states.Size())
	}
}
func TestJavaSetOfStatesDuplicatesButNotEqual(t *testing.T) {
	s := NewSetOfStates(1)
	id := int32(1)
	for i := int32(1); i <= 32; i++ {
		if s.Put(newJavaEqualityDummyTLCState(i, id)) {
			t.Fatalf("first insert%d duplicate", i)
		}
		id++
	}
	if s.Size() != 32 {
		t.Fatalf("size=%d, want32", s.Size())
	}
	for i := int32(1); i <= 32; i++ {
		if s.Put(newJavaEqualityDummyTLCState(i, id)) {
			t.Fatalf("same fp/different id%d duplicate", i)
		}
		id++
	}
	if s.Size() != 64 {
		t.Fatalf("size=%d, want64", s.Size())
	}
	id = 1
	for i := int32(1); i <= 32; i++ {
		if !s.Put(newJavaEqualityDummyTLCState(i, id)) {
			t.Fatalf("same fp/id%d not duplicate", i)
		}
		id++
	}
	if s.Size() != 64 {
		t.Fatalf("size=%d, want64", s.Size())
	}
}
