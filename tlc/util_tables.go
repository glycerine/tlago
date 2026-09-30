package tlc

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type ObjLongTable[K comparable] struct {
	count  int
	length int
	thresh int
	keys   []K
	elems  []int64
	used   []bool
}

func NewObjLongTable[K comparable](size int) *ObjLongTable[K] {
	if size < 0 {
		size = 0
	}
	return &ObjLongTable[K]{
		length: size,
		thresh: size / 2,
		keys:   make([]K, size),
		elems:  make([]int64, size),
		used:   make([]bool, size),
	}
}

func (t *ObjLongTable[K]) Size() int {
	if t == nil {
		return 0
	}
	return t.count
}

func (t *ObjLongTable[K]) Put(key K, elem int64) int {
	if t.count >= t.thresh {
		t.grow()
	}
	loc := t.location(key)
	for {
		if !t.used[loc] {
			t.keys[loc] = key
			t.elems[loc] = elem
			t.used[loc] = true
			t.count++
			return loc
		}
		if t.keys[loc] == key {
			t.elems[loc] = elem
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *ObjLongTable[K]) Add(key K, elem int64) int {
	if t.count >= t.thresh {
		t.grow()
	}
	loc := t.location(key)
	for {
		if !t.used[loc] {
			t.keys[loc] = key
			t.elems[loc] = elem
			t.used[loc] = true
			t.count++
			return loc
		}
		if t.keys[loc] == key {
			t.elems[loc] += elem
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *ObjLongTable[K]) Get(key K) int64 {
	if t == nil || t.length == 0 {
		return 0
	}
	loc := t.location(key)
	for {
		if !t.used[loc] {
			return 0
		}
		if t.keys[loc] == key {
			return t.elems[loc]
		}
		loc = (loc + 1) % t.length
	}
}

func (t *ObjLongTable[K]) MergeInto(other *ObjLongTable[K]) *ObjLongTable[K] {
	if other == nil {
		return t
	}
	for i := 0; i < other.length; i++ {
		if other.used[i] {
			t.Add(other.keys[i], other.elems[i])
		}
	}
	return t
}

func (t *ObjLongTable[K]) ToArray() []K {
	if t == nil {
		return nil
	}
	out := make([]K, 0, t.count)
	for i := 0; i < t.length; i++ {
		if t.used[i] {
			out = append(out, t.keys[i])
		}
	}
	return out
}

func (t *ObjLongTable[K]) Keys() *ObjLongTableEnumerator[K] {
	return &ObjLongTableEnumerator[K]{keys: t.ToArray()}
}

func (t *ObjLongTable[K]) grow() {
	oldKeys := t.keys
	oldElems := t.elems
	oldUsed := t.used
	t.count = 0
	t.length = 2*t.length + 1
	t.thresh = t.length / 2
	t.keys = make([]K, t.length)
	t.elems = make([]int64, t.length)
	t.used = make([]bool, t.length)
	for i := 0; i < len(oldKeys); i++ {
		if oldUsed[i] {
			t.Put(oldKeys[i], oldElems[i])
		}
	}
}

func (t *ObjLongTable[K]) location(key K) int {
	hash := objLongKeyHashCode(key)
	return int(uint32(hash)&0x7fffffff) % t.length
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

func objLongKeyHashCode[K comparable](key K) int32 {
	if h, ok := any(key).(interface{ HashCode() int32 }); ok {
		return h.HashCode()
	}
	if h, ok := any(key).(interface{ JavaHashCode() int32 }); ok {
		return h.JavaHashCode()
	}
	switch value := any(key).(type) {
	case string:
		return javaStringHashCode(value)
	case *UniqueString:
		return javaStringHashCode(value.String())
	case int:
		return int32(value)
	case int8:
		return int32(value)
	case int16:
		return int32(value)
	case int32:
		return value
	case int64:
		return int32(value)
	case uint:
		return int32(value)
	case uint8:
		return int32(value)
	case uint16:
		return int32(value)
	case uint32:
		return int32(value)
	case uint64:
		return int32(value)
	case uintptr:
		return int32(value)
	case bool:
		if value {
			return 1231
		}
		return 1237
	}
	rv := reflect.ValueOf(key)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Chan, reflect.UnsafePointer:
		return int32(rv.Pointer())
	}
	return javaStringHashCode(fmt.Sprintf("%#v", key))
}

type SemanticNodeLongTable struct {
	count  int
	length int
	thresh int
	keys   []semanticNodeKey
	elems  []SemanticNodeLongEntry
	used   []bool
}

type SemanticNodeLongEntry struct {
	Node  SemanticNode
	Value int64
}

func NewSemanticNodeLongTable(size int) *SemanticNodeLongTable {
	if size <= 0 {
		size = 1
	}
	return &SemanticNodeLongTable{
		length: size,
		thresh: size / 2,
		keys:   make([]semanticNodeKey, size),
		elems:  make([]SemanticNodeLongEntry, size),
		used:   make([]bool, size),
	}
}

func (t *SemanticNodeLongTable) Size() int {
	if t == nil {
		return 0
	}
	return t.count
}

func (t *SemanticNodeLongTable) Put(node SemanticNode, elem int64) int {
	t.ensure()
	if t.count >= t.thresh {
		t.grow()
	}
	key := newSemanticNodeKey(node)
	loc := t.location(key)
	for {
		if !t.used[loc] {
			t.keys[loc] = key
			t.elems[loc] = SemanticNodeLongEntry{Node: node, Value: elem}
			t.used[loc] = true
			t.count++
			return loc
		}
		if t.keys[loc] == key {
			t.elems[loc] = SemanticNodeLongEntry{Node: node, Value: elem}
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *SemanticNodeLongTable) Add(node SemanticNode, elem int64) int {
	t.ensure()
	if t.count >= t.thresh {
		t.grow()
	}
	key := newSemanticNodeKey(node)
	loc := t.location(key)
	for {
		if !t.used[loc] {
			t.keys[loc] = key
			t.elems[loc] = SemanticNodeLongEntry{Node: node, Value: elem}
			t.used[loc] = true
			t.count++
			return loc
		}
		if t.keys[loc] == key {
			t.elems[loc].Value += elem
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *SemanticNodeLongTable) Get(node SemanticNode) int64 {
	if t == nil || t.length == 0 {
		return 0
	}
	key := newSemanticNodeKey(node)
	loc := t.location(key)
	for {
		if !t.used[loc] {
			return 0
		}
		if t.keys[loc] == key {
			return t.elems[loc].Value
		}
		loc = (loc + 1) % t.length
	}
}

func (t *SemanticNodeLongTable) MergeInto(other *SemanticNodeLongTable) *SemanticNodeLongTable {
	if other == nil {
		return t
	}
	t.ensure()
	for i := 0; i < other.length; i++ {
		if other.used[i] {
			entry := other.elems[i]
			t.Add(entry.Node, entry.Value)
		}
	}
	return t
}

func (t *SemanticNodeLongTable) ToArray() []SemanticNode {
	if t == nil {
		return nil
	}
	out := make([]SemanticNode, 0, t.count)
	for i := 0; i < t.length; i++ {
		if t.used[i] {
			out = append(out, t.elems[i].Node)
		}
	}
	return out
}

func (t *SemanticNodeLongTable) Keys() *SemanticNodeLongTableEnumerator {
	return &SemanticNodeLongTableEnumerator{keys: t.ToArray()}
}

func (t *SemanticNodeLongTable) ensure() {
	if t.length == 0 {
		t.length = 1
		t.keys = make([]semanticNodeKey, 1)
		t.elems = make([]SemanticNodeLongEntry, 1)
		t.used = make([]bool, 1)
	}
	t.thresh = t.length / 2
}

func (t *SemanticNodeLongTable) grow() {
	oldKeys := t.keys
	oldElems := t.elems
	oldUsed := t.used
	t.count = 0
	t.length = 2*t.length + 1
	t.thresh = t.length / 2
	t.keys = make([]semanticNodeKey, t.length)
	t.elems = make([]SemanticNodeLongEntry, t.length)
	t.used = make([]bool, t.length)
	for i := 0; i < len(oldKeys); i++ {
		if oldUsed[i] {
			t.Put(oldElems[i].Node, oldElems[i].Value)
		}
	}
}

func (t *SemanticNodeLongTable) location(key semanticNodeKey) int {
	hash := semanticNodeKeyHashCode(key)
	return int(uint32(hash)&0x7fffffff) % t.length
}

func semanticNodeKeyHashCode(key semanticNodeKey) int32 {
	result := int32(1)
	result = 31*result + javaStringHashCode(key.typ)
	if key.ptr != 0 {
		result = 31*result + int32(uintptr(key.ptr))
	} else {
		result = 31*result + javaStringHashCode(key.image)
	}
	return result
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
	count  int
	length int
	thresh int
	keys   []int64
	elems  []V
	used   []bool
}

func NewLongObjTable[V any](size int) *LongObjTable[V] {
	if size < 0 {
		size = 0
	}
	return &LongObjTable[V]{
		length: size,
		thresh: size / 2,
		keys:   make([]int64, size),
		elems:  make([]V, size),
		used:   make([]bool, size),
	}
}

func (t *LongObjTable[V]) Size() int {
	if t == nil {
		return 0
	}
	return t.count
}

func (t *LongObjTable[V]) Put(key int64, elem V) int {
	if t.count >= t.thresh {
		t.grow()
	}
	loc := t.location(key)
	for {
		if !t.used[loc] {
			t.keys[loc] = key
			t.elems[loc] = elem
			t.used[loc] = true
			t.count++
			return loc
		}
		if t.keys[loc] == key {
			t.elems[loc] = elem
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *LongObjTable[V]) Get(key int64) (V, bool) {
	var zero V
	if t == nil || t.length == 0 {
		return zero, false
	}
	loc := t.location(key)
	for {
		if !t.used[loc] {
			return zero, false
		}
		if t.keys[loc] == key {
			return t.elems[loc], true
		}
		loc = (loc + 1) % t.length
	}
}

func (t *LongObjTable[V]) grow() {
	oldKeys := t.keys
	oldElems := t.elems
	oldUsed := t.used
	t.count = 0
	t.length = 2*t.length + 1
	t.thresh = t.length / 2
	t.keys = make([]int64, t.length)
	t.elems = make([]V, t.length)
	t.used = make([]bool, t.length)
	for i := 0; i < len(oldKeys); i++ {
		if oldUsed[i] {
			t.Put(oldKeys[i], oldElems[i])
		}
	}
}

func (t *LongObjTable[V]) location(key int64) int {
	return int(uint32(key)&0x7fffffff) % t.length
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
	out := NewVect[E]()
	for _, elem := range v.data {
		out.AddElement(elem)
	}
	for _, elem := range other.data {
		out.AddElement(elem)
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
	return &VectEnumerator[E]{vect: v}
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
	vect  *Vect[E]
	index int
}

func (e *VectEnumerator[E]) HasMoreElements() bool {
	return e != nil && e.vect != nil && e.index < len(e.vect.data)
}

func (e *VectEnumerator[E]) NextElement() E {
	if e == nil || e.vect == nil || e.index >= len(e.vect.data) {
		panic("Vect enumerator exhausted")
	}
	value := e.vect.data[e.index]
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
	cnt := 0
	for i := 0; i < s.length; i++ {
		x := s.table[i]
		if x != 0 {
			s.table[cnt] = x
			cnt++
		}
	}
	sort.Slice(s.table[:cnt], func(i, j int) bool { return s.table[i] < s.table[j] })
	dis := int64(^uint64(0) >> 1)
	var x int64
	i := 0
	if !s.hasZero && cnt > 0 {
		x = s.table[0]
		i = 1
	}
	for ; i < cnt; i++ {
		d := s.table[i] - x
		if d < dis {
			dis = d
		}
		x = s.table[i]
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
