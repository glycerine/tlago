package tlc

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type ObjLongTable[K comparable] struct {
	elems *InsMap[K, int64]
}

func NewObjLongTable[K comparable](size int) *ObjLongTable[K] {
	_ = size
	return &ObjLongTable[K]{elems: NewInsMap[K, int64]()}
}

func (t *ObjLongTable[K]) Size() int {
	if t == nil || t.elems == nil {
		return 0
	}
	return t.elems.Len()
}

func (t *ObjLongTable[K]) Put(key K, elem int64) int {
	t.ensure()
	if _, ok := t.elems.Get2(key); !ok {
		t.elems.Set(key, elem)
		return t.elems.Len() - 1
	}
	index := t.indexOf(key)
	t.elems.Set(key, elem)
	return index
}

func (t *ObjLongTable[K]) Add(key K, elem int64) int {
	t.ensure()
	if cur, ok := t.elems.Get2(key); ok {
		index := t.indexOf(key)
		t.elems.Set(key, cur+elem)
		return index
	}
	t.elems.Set(key, elem)
	return t.elems.Len() - 1
}

func (t *ObjLongTable[K]) Get(key K) int64 {
	if t == nil || t.elems == nil {
		return 0
	}
	return t.elems.Get(key)
}

func (t *ObjLongTable[K]) MergeInto(other *ObjLongTable[K]) *ObjLongTable[K] {
	if other == nil || other.elems == nil {
		return t
	}
	t.ensure()
	for key, value := range other.elems.All() {
		t.Add(key, value)
	}
	return t
}

func (t *ObjLongTable[K]) ToArray() []K {
	if t == nil || t.elems == nil {
		return nil
	}
	out := make([]K, 0, t.elems.Len())
	for key := range t.elems.All() {
		out = append(out, key)
	}
	return out
}

func (t *ObjLongTable[K]) Keys() *ObjLongTableEnumerator[K] {
	return &ObjLongTableEnumerator[K]{keys: t.ToArray()}
}

func (t *ObjLongTable[K]) ensure() {
	if t.elems == nil {
		t.elems = NewInsMap[K, int64]()
	}
}

func (t *ObjLongTable[K]) indexOf(key K) int {
	i := 0
	for existing := range t.elems.All() {
		if existing == key {
			return i
		}
		i++
	}
	return -1
}

type ObjLongTableEnumerator[K comparable] struct {
	keys  []K
	index int
}

func (e *ObjLongTableEnumerator[K]) NextElement() (K, bool) {
	var zero K
	if e == nil || e.index >= len(e.keys) {
		return zero, false
	}
	value := e.keys[e.index]
	e.index++
	return value, true
}

type SemanticNodeLongTable struct {
	elems *InsMap[semanticNodeKey, SemanticNodeLongEntry]
}

type SemanticNodeLongEntry struct {
	Node  SemanticNode
	Value int64
}

func NewSemanticNodeLongTable(size int) *SemanticNodeLongTable {
	_ = size
	return &SemanticNodeLongTable{elems: NewInsMap[semanticNodeKey, SemanticNodeLongEntry]()}
}

func (t *SemanticNodeLongTable) Size() int {
	if t == nil || t.elems == nil {
		return 0
	}
	return t.elems.Len()
}

func (t *SemanticNodeLongTable) Put(node SemanticNode, elem int64) int {
	t.ensure()
	key := newSemanticNodeKey(node)
	if _, ok := t.elems.Get2(key); !ok {
		t.elems.Set(key, SemanticNodeLongEntry{Node: node, Value: elem})
		return t.elems.Len() - 1
	}
	index := t.indexOf(key)
	t.elems.Set(key, SemanticNodeLongEntry{Node: node, Value: elem})
	return index
}

func (t *SemanticNodeLongTable) Add(node SemanticNode, elem int64) int {
	t.ensure()
	key := newSemanticNodeKey(node)
	if cur, ok := t.elems.Get2(key); ok {
		index := t.indexOf(key)
		cur.Value += elem
		t.elems.Set(key, cur)
		return index
	}
	t.elems.Set(key, SemanticNodeLongEntry{Node: node, Value: elem})
	return t.elems.Len() - 1
}

func (t *SemanticNodeLongTable) Get(node SemanticNode) int64 {
	if t == nil || t.elems == nil {
		return 0
	}
	return t.elems.Get(newSemanticNodeKey(node)).Value
}

func (t *SemanticNodeLongTable) MergeInto(other *SemanticNodeLongTable) *SemanticNodeLongTable {
	if other == nil || other.elems == nil {
		return t
	}
	t.ensure()
	for _, entry := range other.elems.All() {
		t.Add(entry.Node, entry.Value)
	}
	return t
}

func (t *SemanticNodeLongTable) ToArray() []SemanticNode {
	if t == nil || t.elems == nil {
		return nil
	}
	out := make([]SemanticNode, 0, t.elems.Len())
	for _, entry := range t.elems.All() {
		out = append(out, entry.Node)
	}
	return out
}

func (t *SemanticNodeLongTable) Keys() *SemanticNodeLongTableEnumerator {
	return &SemanticNodeLongTableEnumerator{keys: t.ToArray()}
}

func (t *SemanticNodeLongTable) ensure() {
	if t.elems == nil {
		t.elems = NewInsMap[semanticNodeKey, SemanticNodeLongEntry]()
	}
}

func (t *SemanticNodeLongTable) indexOf(key semanticNodeKey) int {
	i := 0
	for existing := range t.elems.All() {
		if existing == key {
			return i
		}
		i++
	}
	return -1
}

type SemanticNodeLongTableEnumerator struct {
	keys  []SemanticNode
	index int
}

func (e *SemanticNodeLongTableEnumerator) NextElement() SemanticNode {
	if e == nil || e.index >= len(e.keys) {
		return nil
	}
	value := e.keys[e.index]
	e.index++
	return value
}

type LongObjTable[V any] struct {
	elems *InsMap[int64, V]
}

func NewLongObjTable[V any](size int) *LongObjTable[V] {
	_ = size
	return &LongObjTable[V]{elems: NewInsMap[int64, V]()}
}

func (t *LongObjTable[V]) Size() int {
	if t == nil || t.elems == nil {
		return 0
	}
	return t.elems.Len()
}

func (t *LongObjTable[V]) Put(key int64, elem V) int {
	t.ensure()
	if _, ok := t.elems.Get2(key); !ok {
		t.elems.Set(key, elem)
		return t.elems.Len() - 1
	}
	index := t.indexOf(key)
	t.elems.Set(key, elem)
	return index
}

func (t *LongObjTable[V]) Get(key int64) (V, bool) {
	var zero V
	if t == nil || t.elems == nil {
		return zero, false
	}
	return t.elems.Get2(key)
}

func (t *LongObjTable[V]) ensure() {
	if t.elems == nil {
		t.elems = NewInsMap[int64, V]()
	}
}

func (t *LongObjTable[V]) indexOf(key int64) int {
	i := 0
	for existing := range t.elems.All() {
		if existing == key {
			return i
		}
		i++
	}
	return -1
}

type Vect[E any] struct {
	data []E
}

func NewVect[E any]() *Vect[E] {
	return NewVectWithCapacity[E](10)
}

func NewVectWithCapacity[E any](capacity int) *Vect[E] {
	if capacity < 0 {
		capacity = 0
	}
	return &Vect[E]{data: make([]E, 0, capacity)}
}

func NewVectFrom[E any](values []E) *Vect[E] {
	out := make([]E, len(values))
	copy(out, values)
	return &Vect[E]{data: out}
}

func (v *Vect[E]) AddElement(elem E) {
	v.data = append(v.data, elem)
}

func (v *Vect[E]) Concat(other *Vect[E]) *Vect[E] {
	out := NewVectWithCapacity[E](v.Size() + other.Size())
	for _, elem := range v.data {
		out.AddElement(elem)
	}
	if other != nil {
		for _, elem := range other.data {
			out.AddElement(elem)
		}
	}
	return out
}

func (v *Vect[E]) Capacity() int {
	return cap(v.data)
}

func (v *Vect[E]) Contains(elem E) bool {
	return v.IndexOf(elem) != -1
}

func (v *Vect[E]) CopyInto(array []E) {
	copy(array, v.data)
}

func (v *Vect[E]) ElementAt(index int) E {
	return v.data[index]
}

func (v *Vect[E]) Elements() *VectEnumerator[E] {
	return &VectEnumerator[E]{values: v.ToSlice()}
}

func (v *Vect[E]) EnsureCapacity(minCapacity int) {
	if cap(v.data) >= minCapacity {
		return
	}
	next := cap(v.data) * 2
	if next < minCapacity {
		next = minCapacity
	}
	if next < 1 {
		next = 1
	}
	out := make([]E, len(v.data), next)
	copy(out, v.data)
	v.data = out
}

func (v *Vect[E]) FirstElement() E {
	return v.data[0]
}

func (v *Vect[E]) IndexOf(elem E) int {
	return v.IndexOfFrom(elem, 0)
}

func (v *Vect[E]) IndexOfFrom(elem E, index int) int {
	for pos := index; pos < len(v.data); pos++ {
		if reflect.DeepEqual(elem, v.data[pos]) {
			return pos
		}
	}
	return -1
}

func (v *Vect[E]) InsertElementAt(elem E, index int) {
	if index < 0 || index > len(v.data) {
		panic("Vect index out of bounds")
	}
	var zero E
	v.data = append(v.data, zero)
	copy(v.data[index+1:], v.data[index:])
	v.data[index] = elem
}

func (v *Vect[E]) IsEmpty() bool {
	return len(v.data) == 0
}

func (v *Vect[E]) LastElement() E {
	return v.data[len(v.data)-1]
}

func (v *Vect[E]) RemoveLastElement() {
	if len(v.data) == 0 {
		panic("Vect is empty")
	}
	var zero E
	v.data[len(v.data)-1] = zero
	v.data = v.data[:len(v.data)-1]
}

func (v *Vect[E]) SetElementAt(elem E, index int) {
	v.data[index] = elem
}

func (v *Vect[E]) RemoveElementAt(index int) {
	if index < 0 || index >= len(v.data) {
		panic("Vect index out of bounds")
	}
	copy(v.data[index:], v.data[index+1:])
	var zero E
	v.data[len(v.data)-1] = zero
	v.data = v.data[:len(v.data)-1]
}

func (v *Vect[E]) RemoveAll(cnt int) {
	if cnt < 0 {
		cnt = 0
	}
	if cnt > len(v.data) {
		cnt = len(v.data)
	}
	var zero E
	for i := cnt; i < len(v.data); i++ {
		v.data[i] = zero
	}
	v.data = v.data[:cnt]
}

func (v *Vect[E]) Pop() E {
	elem := v.LastElement()
	v.RemoveLastElement()
	return elem
}

func (v *Vect[E]) Push(elem E) {
	v.AddElement(elem)
}

func (v *Vect[E]) Size() int {
	return len(v.data)
}

func (v *Vect[E]) ToSlice() []E {
	out := make([]E, len(v.data))
	copy(out, v.data)
	return out
}

func (v *Vect[E]) Equal(other *Vect[E]) bool {
	if v == nil || other == nil {
		return v == other
	}
	if len(v.data) != len(other.data) {
		return false
	}
	for i := range v.data {
		if !reflect.DeepEqual(v.data[i], other.data[i]) {
			return false
		}
	}
	return true
}

func (v *Vect[E]) String() string {
	parts := make([]string, len(v.data))
	for i, elem := range v.data {
		parts[i] = fmt.Sprint(elem)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

type VectEnumerator[E any] struct {
	values []E
	index  int
}

func (e *VectEnumerator[E]) HasMoreElements() bool {
	return e != nil && e.index < len(e.values)
}

func (e *VectEnumerator[E]) NextElement() E {
	if e == nil || e.index >= len(e.values) {
		panic("Vect enumerator exhausted")
	}
	value := e.values[e.index]
	e.index++
	return value
}

type SetOfLong struct {
	count   int
	length  int
	thresh  int
	table   []int64
	hasZero bool
}

func NewSetOfLong(size int) *SetOfLong {
	if size <= 0 {
		size = 1
	}
	return &SetOfLong{
		length: size,
		thresh: size / 2,
		table:  make([]int64, size),
	}
}

func (s *SetOfLong) Put(key int64) bool {
	if s.count >= s.thresh {
		s.grow()
	}
	if key == 0 {
		if s.hasZero {
			return true
		}
		s.hasZero = true
		s.count++
		return false
	}
	loc := int(uint32(key)&0x7fffffff) % s.length
	for {
		elem := s.table[loc]
		if elem == key {
			return true
		}
		if elem == 0 {
			s.table[loc] = key
			s.count++
			return false
		}
		loc = (loc + 1) % s.length
	}
}

func (s *SetOfLong) Contains(key int64) bool {
	if key == 0 {
		return s.hasZero
	}
	loc := int(uint32(key)&0x7fffffff) % s.length
	for {
		elem := s.table[loc]
		if elem == key {
			return true
		}
		if elem == 0 {
			return false
		}
		loc = (loc + 1) % s.length
	}
}

func (s *SetOfLong) Size() int {
	if s == nil {
		return 0
	}
	return s.count
}

func (s *SetOfLong) Sizeof() int64 {
	if s == nil {
		return 0
	}
	return 20 + int64(8*s.length)
}

func (s *SetOfLong) CheckFPs() int64 {
	if s == nil {
		return int64(^uint64(0) >> 1)
	}
	values := make([]int64, 0, s.count)
	if s.hasZero {
		values = append(values, 0)
	}
	for _, x := range s.table {
		if x != 0 {
			values = append(values, x)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	dis := int64(^uint64(0) >> 1)
	for i := 1; i < len(values); i++ {
		d := values[i] - values[i-1]
		if d < dis {
			dis = d
		}
	}
	return dis
}

func (s *SetOfLong) ToSlice() []int64 {
	values := make([]int64, 0, s.count)
	if s.hasZero {
		values = append(values, 0)
	}
	for _, x := range s.table {
		if x != 0 {
			values = append(values, x)
		}
	}
	return values
}

func (s *SetOfLong) BeginChkpt(out *ValueOutputStream) error {
	if s == nil {
		if err := out.WriteInt(0); err != nil {
			return err
		}
		if err := out.WriteInt(0); err != nil {
			return err
		}
		if err := out.WriteInt(0); err != nil {
			return err
		}
		return out.WriteBool(false)
	}
	if err := out.WriteInt(int32(s.count)); err != nil {
		return err
	}
	if err := out.WriteInt(int32(s.length)); err != nil {
		return err
	}
	if err := out.WriteInt(int32(s.thresh)); err != nil {
		return err
	}
	if err := out.WriteBool(s.hasZero); err != nil {
		return err
	}
	for _, key := range s.table {
		if key == 0 {
			continue
		}
		if err := out.WriteLong(key); err != nil {
			return err
		}
	}
	return nil
}

func (s *SetOfLong) Recover(in *ValueInputStream) error {
	count, err := in.ReadInt()
	if err != nil {
		return err
	}
	length, err := in.ReadInt()
	if err != nil {
		return err
	}
	thresh, err := in.ReadInt()
	if err != nil {
		return err
	}
	hasZero, err := in.ReadBool()
	if err != nil {
		return err
	}
	if length <= 0 {
		length = 1
	}
	s.count = 0
	s.length = int(length)
	s.thresh = int(thresh)
	s.table = make([]int64, s.length)
	s.hasZero = false
	if hasZero {
		s.hasZero = true
		s.count = 1
	}
	num := int(count)
	if hasZero {
		num--
	}
	for i := 0; i < num; i++ {
		key, err := in.ReadLong()
		if err != nil {
			return err
		}
		s.putWithoutGrow(key)
	}
	return nil
}

func (s *SetOfLong) grow() {
	old := s.table
	oldHasZero := s.hasZero
	s.count = 0
	s.length = 2*s.length + 1
	s.thresh = s.length / 2
	s.table = make([]int64, s.length)
	s.hasZero = false
	if oldHasZero {
		s.Put(0)
	}
	for _, key := range old {
		if key != 0 {
			s.Put(key)
		}
	}
}

func (s *SetOfLong) putWithoutGrow(key int64) {
	if key == 0 {
		if !s.hasZero {
			s.hasZero = true
			s.count++
		}
		return
	}
	loc := int(uint32(key)&0x7fffffff) % s.length
	for {
		elem := s.table[loc]
		if elem == key {
			return
		}
		if elem == 0 {
			s.table[loc] = key
			s.count++
			return
		}
		loc = (loc + 1) % s.length
	}
}
