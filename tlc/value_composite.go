package tlc

import (
	"fmt"
	"math"
	"strings"
)

type ValueEnumeration interface {
	Reset()
	NextElement() Value
	Err() error
}

func requireValueEnumerationReceiver(enumeration any) {
	if isNil(enumeration) {
		panic(NewNullPointerException())
	}
}

// FunctionValue represents values implementing Java's FunctionValue interface.
type FunctionValue interface {
	ApplyWithControl(Value, int) (Value, error)
	ApplyArgs([]Value, int) (Value, error)
	GetDomain() (Value, error)
	Select(Value) (Value, error)
}

var (
	_ FunctionValue = (*TupleValue)(nil)
	_ FunctionValue = (*RecordValue)(nil)
	_ FunctionValue = (*CounterExample)(nil)
	_ FunctionValue = (*FcnRcdValue)(nil)
	_ FunctionValue = (*FcnLambdaValue)(nil)
)

// ValueEnumerationAll represents ValueEnumeration.all. Go interfaces cannot
// supply Java-style default methods, so the shared operation takes the
// enumeration explicitly.
func ValueEnumerationAll(enumeration ValueEnumeration) []Value {
	if isNil(enumeration) {
		panic(NewNullPointerException())
	}
	values := make([]Value, 0)
	for value := nextEnumerationElement(enumeration); value != nil; value = nextEnumerationElement(enumeration) {
		values = append(values, value)
	}
	return values
}

// ValueEnumerationForEach represents ValueEnumeration.forEach.
func ValueEnumerationForEach(enumeration ValueEnumeration, action func(Value)) {
	if isNil(enumeration) {
		panic(NewNullPointerException())
	}
	for value := nextEnumerationElement(enumeration); value != nil; value = nextEnumerationElement(enumeration) {
		if action == nil {
			panic(NewNullPointerException())
		}
		action(value)
	}
}

// ValueEnumerationAsSet represents ValueEnumeration.asSet.
func ValueEnumerationAsSet(enumeration ValueEnumeration) *SetEnumValue {
	values := NewValueVec()
	ValueEnumerationForEach(enumeration, values.AddElement)
	return NewSetEnumValueVec(values, false)
}

type EnumerableOrdering byte

const (
	EnumerableOrderingUndefined EnumerableOrdering = iota
	EnumerableOrderingNormalized
	EnumerableOrderingRandomized
)

type Enumerable interface {
	Value
	Elements() ValueEnumeration
	ElementsK(int) ValueEnumeration
	ElementsOrdered(EnumerableOrdering) ValueEnumeration
	GetRandomSubset(int) (*SetEnumValue, error)
	IsSubsetEq(Value) (*BoolValue, error)
}

// ValueVec keeps the Java array and active count independently. Direct slot
// writes can address unused storage, and failed stores can leave count > capacity.
type ValueVec struct {
	data  []Value
	count int32
}

func NewValueVec(capacities ...int) *ValueVec {
	if len(capacities) > 1 {
		panic(NewIllegalArgumentException("NewValueVec accepts at most one capacity"))
	}
	capacity := 10
	if len(capacities) != 0 {
		capacity = int(int32(capacities[0]))
	}
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(capacity)))
	}
	return &ValueVec{data: make([]Value, capacity)}
}

func NewValueVecFrom(values []Value) *ValueVec {
	if values == nil {
		panic(NewNullPointerException())
	}
	return &ValueVec{data: values[:len(values):len(values)], count: int32(len(values))}
}

func NewValueVecFromCollection(values []Value) *ValueVec {
	if values == nil {
		panic(NewNullPointerException())
	}
	vec := NewValueVec(len(values))
	for _, value := range values {
		vec.AddElement(value)
	}
	return vec
}

func (v *ValueVec) Add(val Value) {
	if v.Len() == v.Cap() {
		v.ensureCapacity(int(v.count + 1))
	}
	index := v.count
	v.count++ // Source post-increment runs before the array store's bounds check.
	v.Set(int(index), val)
}

func (v *ValueVec) AddAt(val Value, index int) {
	v.Set(index, val)
	v.count++
}

func (v *ValueVec) AddElementAt(val Value, index int) { v.AddAt(val, index) }

func (v *ValueVec) AddElement(val Value) { v.Add(val) }

func (v *ValueVec) AddSortedUnique(val Value) error {
	if v.Len() == v.Cap() {
		v.ensureCapacity(int(v.count + 1))
	}
	for i := 0; i < v.Len(); i++ {
		elem := v.At(i)
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		cmp, err := elem.Compare(val)
		if err != nil {
			return err
		}
		if cmp == 0 {
			return nil
		}
		if cmp > 0 {
			for j := v.Len() - 1; j >= i; j-- {
				v.Set(j+1, v.At(j))
			}
			v.Set(i, val)
			v.count++
			return nil
		}
	}
	index := v.count
	v.count++
	v.Set(int(index), val)
	return nil
}

func (v *ValueVec) AddElement1(val Value) {
	if err := v.AddSortedUnique(val); err != nil {
		panic(err)
	}
}

func (v *ValueVec) InsertAt(val Value, index int) {
	index = int(int32(index))
	if v.Len() == v.Cap() {
		v.ensureCapacity(int(v.count + 1))
	}
	valueVecArrayCopy(v.data, index, v.data, int(int32(index)+1), int(v.count-int32(index)))
	v.Set(index, val)
	v.count++
}

func (v *ValueVec) Len() int {
	if v == nil {
		panic(NewNullPointerException())
	}
	return int(v.count)
}
func (v *ValueVec) Cap() int {
	if v == nil || v.data == nil {
		panic(NewNullPointerException())
	}
	return len(v.data)
}
func (v *ValueVec) Capacity() int { return v.Cap() }
func (v *ValueVec) Empty() bool   { return v.Len() == 0 }
func (v *ValueVec) IsEmpty() bool { return v.Empty() }
func (v *ValueVec) At(index int) Value {
	if v == nil || v.data == nil {
		panic(NewNullPointerException())
	}
	index = int(int32(index))
	if index < 0 || index >= len(v.data) {
		panic(NewArrayIndexOutOfBoundsException(index, len(v.data)))
	}
	return v.data[index]
}
func (v *ValueVec) ElementAt(index int) Value { return v.At(index) }
func (v *ValueVec) Set(index int, val Value) {
	if v == nil || v.data == nil {
		panic(NewNullPointerException())
	}
	index = int(int32(index))
	if index < 0 || index >= len(v.data) {
		panic(NewArrayIndexOutOfBoundsException(index, len(v.data)))
	}
	v.data[index] = val
}
func (v *ValueVec) First() Value        { return v.At(0) }
func (v *ValueVec) FirstElement() Value { return v.First() }
func (v *ValueVec) Last() Value         { return v.At(int(int32(v.Len()) - 1)) }
func (v *ValueVec) LastElement() Value  { return v.Last() }

func (v *ValueVec) InsertElementAt(val Value, index int) { v.InsertAt(val, index) }

func (v *ValueVec) SetElementAt(val Value, index int) { v.Set(index, val) }

func (v *ValueVec) Size() int { return v.Len() }

func (v *ValueVec) ToArray() []Value {
	length := v.Len()
	if length < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(length)))
	}
	out := make([]Value, length)
	valueVecArrayCopy(v.data, 0, out, 0, length)
	return out
}

func (v *ValueVec) CopyInto(values []Value) {
	length := v.Len()
	valueVecArrayCopy(v.data, 0, values, 0, length)
}

func (v *ValueVec) Contains(val Value) bool { return v.IndexOf(val) != -1 }
func (v *ValueVec) IndexOf(val Value) int   { return v.IndexOfFrom(val, 0) }
func (v *ValueVec) IndexOfFrom(val Value, index int) int {
	for i := int32(index); int(i) < v.Len(); i++ {
		elem := v.At(int(i))
		if isNil(val) {
			panic(NewNullPointerException())
		}
		equal, err := val.Equal(elem)
		if err != nil {
			panic(err)
		}
		if equal {
			return int(i)
		}
	}
	return -1
}

func (v *ValueVec) Search(val Value, sorted bool) (bool, error) {
	if sorted {
		low, high := 0, v.Len()
		for low < high {
			mid := int(int32(low+high) >> 1)
			elem := v.At(mid)
			if isNil(val) {
				panic(NewNullPointerException())
			}
			cmp, err := val.Compare(elem)
			if err != nil {
				return false, err
			}
			if cmp == 0 {
				return true, nil
			}
			if cmp < 0 {
				high = mid
			} else {
				low = mid + 1
			}
		}
		return false, nil
	}
	for i := 0; i < v.Len(); i++ {
		elem := v.At(i)
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		equal, err := elem.Equal(val)
		if err != nil {
			return false, err
		}
		if equal {
			return true, nil
		}
	}
	return false, nil
}

func (v *ValueVec) Sort(noDup bool) error {
	newCount := 0
	if v.Len() != 0 {
		newCount = 1
	}
	for i := 1; i < v.Len(); i++ {
		elem := v.At(i)
		cmp, idx, low, high := 0, 0, 0, newCount
		for low < high {
			idx = int(int32(low+high) >> 1)
			other := v.At(idx)
			if isNil(elem) {
				panic(NewNullPointerException())
			}
			var err error
			cmp, err = elem.Compare(other)
			if err != nil {
				return err
			}
			if cmp == 0 {
				break
			}
			if cmp < 0 {
				high = idx
			} else {
				low = idx + 1
			}
		}
		if cmp != 0 || !noDup {
			if cmp >= 0 {
				idx++
			}
			for j := newCount; j > idx; j-- {
				v.Set(j, v.At(j-1))
			}
			v.Set(idx, elem)
			newCount++
		}
	}
	v.count = int32(newCount)
	return nil
}

func (v *ValueVec) String() string {
	var out strings.Builder
	out.WriteString("{")
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			out.WriteString(", ")
		}
		elem := v.At(i)
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		out.WriteString(elem.String())
	}
	out.WriteString("}")
	return out.String()
}

func (v *ValueVec) EnsureCapacity(minCapacity int) { v.ensureCapacity(minCapacity) }
func (v *ValueVec) ensureCapacity(minCapacity int) {
	capacity := v.Cap()
	minCapacity = int(int32(minCapacity))
	if capacity >= Globals.SetBound {
		panic(NewWrongInvocationException(fmt.Sprintf("Attempted to construct a set with too many elements (>%d).", Globals.SetBound)))
	}
	if capacity >= minCapacity {
		return
	}
	newCapacity := int(int32(capacity) + int32(capacity))
	if newCapacity < minCapacity {
		newCapacity = minCapacity
	}
	if newCapacity > Globals.SetBound {
		newCapacity = Globals.SetBound
	}
	if newCapacity < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(newCapacity)))
	}
	old := v.data
	v.data = make([]Value, newCapacity)
	valueVecArrayCopy(old, 0, v.data, 0, v.Len())
}

// System.arraycopy checks all bounds before moving any elements. Source vector
// insertion uses this overload, including its object-array diagnostics.
func valueVecArrayCopy(source []Value, sourceIndex int, destination []Value, destinationIndex int, length int) {
	if source == nil || destination == nil {
		panic(NewNullPointerException())
	}
	fail := func(message string) {
		panic(&ArrayIndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString("arraycopy: "+message), nil)})
	}
	if sourceIndex < 0 {
		fail(fmt.Sprintf("source index %d out of bounds for object array[%d]", sourceIndex, len(source)))
	}
	if destinationIndex < 0 {
		fail(fmt.Sprintf("destination index %d out of bounds for object array[%d]", destinationIndex, len(destination)))
	}
	if length < 0 {
		fail(fmt.Sprintf("length %d is negative", length))
	}
	if int64(sourceIndex)+int64(length) > int64(len(source)) {
		fail(fmt.Sprintf("last source index %d out of bounds for object array[%d]", int64(sourceIndex)+int64(length), len(source)))
	}
	if int64(destinationIndex)+int64(length) > int64(len(destination)) {
		fail(fmt.Sprintf("last destination index %d out of bounds for object array[%d]", int64(destinationIndex)+int64(length), len(destination)))
	}
	copy(destination[destinationIndex:destinationIndex+length], source[sourceIndex:sourceIndex+length])
}

type TupleValue struct {
	BaseValue
	Elems []Value

	elemsCaptured bool
	elemsOwner    []Value
}

var EmptyTuple = NewTupleValue([]Value{})

func NewTupleValue(elems []Value, cms ...CostModel) *TupleValue {
	return &TupleValue{BaseValue: newBaseValue(cms...), Elems: elems, elemsCaptured: true, elemsOwner: elems}
}

func (v *TupleValue) ownedElems() []Value {
	if v.elemsCaptured {
		return v.elemsOwner
	}
	return v.Elems
}

func (v *TupleValue) captureElems(elems []Value) {
	v.Elems, v.elemsOwner, v.elemsCaptured = elems, elems, true
}

func NewSingletonTupleValue(value Value) *TupleValue {
	return NewTupleValue([]Value{value})
}

func NewPairTupleValue(first, second Value) *TupleValue {
	return NewTupleValue([]Value{first, second})
}

func (v *TupleValue) GetElem(index int) Value {
	if v == nil || v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	index = int(int32(index))
	if index < 0 || index >= len(v.ownedElems()) {
		panic(NewArrayIndexOutOfBoundsException(index, len(v.ownedElems())))
	}
	return v.ownedElems()[index]
}

func (v *TupleValue) GetElems() []Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.ownedElems()
}

func (v *TupleValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return TupleValueKind
}
func (v *TupleValue) KindString() string { return ValueKindString(v) }
func (v *TupleValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *TupleValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	var o *TupleValue
	if !isNil(other) {
		o = asTupleValue(other)
	}
	if o == nil {
		return asFcnRcdValue(v).Compare(other)
	}
	if v.ownedElems() == nil || o.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	if len(v.ownedElems()) != len(o.ownedElems()) {
		return len(v.ownedElems()) - len(o.ownedElems()), nil
	}
	for i := range v.ownedElems() {
		if isNil(v.ownedElems()[i]) {
			panic(NewNullPointerException())
		}
		cmp, err := v.ownedElems()[i].Compare(o.ownedElems()[i])
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *TupleValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	var o *TupleValue
	if !isNil(other) {
		o = asTupleValue(other)
	}
	if o == nil {
		return asFcnRcdValue(v).Equal(other)
	}
	if v.ownedElems() == nil || o.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	if len(v.ownedElems()) != len(o.ownedElems()) {
		return false, nil
	}
	for i := range v.ownedElems() {
		if isNil(v.ownedElems()[i]) {
			panic(NewNullPointerException())
		}
		eq, err := v.ownedElems()[i].Equal(o.ownedElems()[i])
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *TupleValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure("Attempted to check set membership in a tuple value.")
}

func (v *TupleValue) IsFinite() (bool, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true, nil
}
func (v *TupleValue) Size() (int, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.ownedElems() == nil {
		return 0, NewNullPointerException()
	}
	return len(v.ownedElems()), nil
}
func (v *TupleValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *TupleValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}

func (v *TupleValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	for _, elem := range v.ownedElems() {
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		elem.DeepNormalize()
	}
}

func (v *TupleValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	for _, elem := range v.ownedElems() {
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		if !elem.IsDefined() {
			return false
		}
	}
	return true
}

func (v *TupleValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	out := make([]Value, len(v.ownedElems()))
	for i, elem := range v.ownedElems() {
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		out[i] = elem.DeepCopy()
	}
	return NewTupleValue(out)
}

func (v *TupleValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	out := make([]Value, len(v.ownedElems()))
	changed := false
	for i, elem := range v.ownedElems() {
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		out[i] = elem.Permute(perm)
		changed = changed || out[i] != elem
	}
	if changed {
		return NewTupleValue(out)
	}
	return v
}

func (v *TupleValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.ownedElems())))
	for i, elem := range v.ownedElems() {
		fp = FP64ExtendByte(fp, byte(IntValueKind))
		fp = FP64ExtendInt(fp, int32(i+1))
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		fp = elem.FingerPrint(fp)
	}
	return fp
}

func (v *TupleValue) Apply(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(arg) {
		panic(NewNullPointerException())
	}
	i, ok := arg.(*IntValue)
	if !ok {
		return nil, v.tupleAssertFailure("Attempted to access tuple at a non integral index: " + ValuesPPR(arg))
	}
	idx := int(i.ownedVal())
	if idx > 0 && v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	if idx <= 0 || idx > len(v.ownedElems()) {
		return nil, v.tupleAssertFailure(fmt.Sprintf("Attempted to access index %d of tuple\n%s\nwhich is out of bounds.", idx, ValuesPPR(v)))
	}
	return v.ownedElems()[idx-1], nil
}

func (v *TupleValue) ApplyWithControl(arg Value, control int) (Value, error) {
	return v.Apply(arg)
}

// ApplyArgs ports TupleValue.apply(Value[], int). Java discards control when
// delegating to the single-argument overload, whose result is control-independent.
func (v *TupleValue) ApplyArgs(args []Value, control int) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if args == nil {
		panic(NewNullPointerException())
	}
	if len(args) != 1 {
		return nil, v.tupleAssertFailure(fmt.Sprintf("Attempted to access tuple with %d arguments when it expects 1.", len(args)))
	}
	return v.Apply(args[0])
}

func (v *TupleValue) tupleAssertFailure(message string) *TLCError {
	if source := v.GetSource(); source != nil {
		return NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
	}
	failure := newTLCError(ECGeneral, "%s", message)
	failure.Runtime = true
	return failure
}

func (v *TupleValue) Select(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(arg) {
		panic(NewNullPointerException())
	}
	i, ok := arg.(*IntValue)
	if !ok {
		return nil, v.tupleAssertFailure("Attempted to access tuple at a non integral index: " + ValuesPPR(arg))
	}
	idx := int(i.ownedVal())
	if idx > 0 && v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	if idx > 0 && idx <= len(v.ownedElems()) {
		return v.ownedElems()[idx-1], nil
	}
	return nil, nil
}

func (v *TupleValue) Domain() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	size, err := v.Size()
	if err != nil {
		panic(err)
	}
	return NewIntervalValue(1, int32(size))
}

func (v *TupleValue) GetDomain() (Value, error) {
	return v.Domain(), nil
}

func (v *TupleValue) ToTuple() *TupleValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *TupleValue) ToRcd() *RecordValue {
	if v == nil || v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	if len(v.ownedElems()) == 0 {
		return EmptyRecord
	}
	return nil
}

func (v *TupleValue) ToFcnRcd() *FcnRcdValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	size, err := v.Size()
	if err != nil {
		panic(err)
	}
	v.CM.incValueSecondary(int64(size))
	return NewFcnRcdIntervalValue(NewIntervalValue(1, int32(size)), v.ownedElems(), v.CM)
}

func (v *TupleValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index >= len(ex.Path) {
		return ex.Value, nil
	}
	if v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	out := make([]Value, len(v.ownedElems()))
	arcVal := ex.Current()
	arc, ok := arcVal.(*IntValue)
	if !ok || arc == nil {
		if isNil(arcVal) {
			panic(NewNullPointerException())
		}
		PrintWarning(ECTLCWrongTupleFieldName, ValuesPPR(arcVal))
		return ex.Value, nil
	}
	idx := int(arc.ownedVal() - 1)
	// Java allocates the replacement tuple before checking bounds and returns it
	// even when the integer field is out of range, leaving the slots unfilled.
	if 0 <= idx && idx < len(v.ownedElems()) {
		copy(out, v.ownedElems())
		next := ex
		next.Index++
		if isNil(v.ownedElems()[idx]) {
			panic(NewNullPointerException())
		}
		val, err := v.ownedElems()[idx].TakeExcept(next)
		if err != nil {
			return nil, err
		}
		out[idx] = val
	}
	return NewTupleValue(out), nil
}

func (v *TupleValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	var cur Value = v
	for _, ex := range exs {
		if isNil(cur) {
			panic(NewNullPointerException())
		}
		next, err := cur.TakeExcept(ex)
		if err != nil {
			return nil, err
		}
		cur = next
	}
	return cur, nil
}

func (v *TupleValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *TupleValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	sb.WriteString("<<")
	if v.ownedElems() == nil {
		panic(NewNullPointerException())
	}
	for i, elem := range v.ownedElems() {
		if i > 0 {
			sb.WriteString(", ")
		}
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		sb = appendValueString(elem, sb, offset, swallow)
	}
	sb.WriteString(">>")
	return sb
}

type SetEnumValue struct {
	BaseValue
	Elems  *ValueVec
	IsNorm bool
}

func (v *SetEnumValue) IsSubsetEq(other Value) (*BoolValue, error) {
	return defaultEnumerableSubsetEq(v, v, other)
}

var (
	EmptySet  = &SetEnumValue{Elems: NewValueVec(0), IsNorm: true}
	DummyEnum = &SetEnumValue{IsNorm: true}
)

func NewEmptySetEnumValue(cms ...CostModel) *SetEnumValue {
	return &SetEnumValue{BaseValue: newBaseValue(cms...), Elems: NewValueVec(0), IsNorm: true}
}

func NewSingletonSetEnumValue(value Value) *SetEnumValue {
	return NewSetEnumValue([]Value{value}, true)
}

func NewSetEnumValue(values []Value, isNorm bool, cms ...CostModel) *SetEnumValue {
	return &SetEnumValue{BaseValue: newBaseValue(cms...), Elems: NewValueVecFrom(values), IsNorm: isNorm}
}

func NewSetEnumValueVec(values *ValueVec, isNorm bool, cms ...CostModel) *SetEnumValue {
	return &SetEnumValue{BaseValue: newBaseValue(cms...), Elems: values, IsNorm: isNorm}
}

func (v *SetEnumValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return SetEnumValueKind
}
func (v *SetEnumValue) KindString() string { return ValueKindString(v) }
func (v *SetEnumValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *SetEnumValue) IsSetOfAtoms() bool {
	if v == nil {
		panic(NewNullPointerException(`Cannot read field "elems" because "this" is null`))
	}
	if v.Elems == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.value.impl.ValueVec.size()" because "this.elems" is null`))
	}
	for i := 0; i < v.Elems.Len(); i++ {
		element := v.Elems.At(i)
		if nested, ok := element.(*SetEnumValue); ok && nested != nil {
			if !nested.IsSetOfAtoms() {
				return false
			}
			continue
		}
		if !ValueIsAtom(element) {
			return false
		}
	}
	return true
}

func (v *SetEnumValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	var o *SetEnumValue
	if !isNil(other) {
		o, err = tryToSetEnumValue(other)
	}
	if err != nil {
		return 0, err
	}
	if o == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueCompareTo(v)
		}
		setText := ValuesPPR(v)
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the set %s with the value:\n%s", setText, ValuesPPR(other)))
	}
	if _, err := v.normalizeSet(); err != nil {
		return 0, err
	}
	if _, err := o.normalizeSet(); err != nil {
		return 0, err
	}
	if v.Elems.Len() != o.Elems.Len() {
		return v.Elems.Len() - o.Elems.Len(), nil
	}
	for i := 0; i < v.Elems.Len(); i++ {
		left, right := v.Elems.At(i), o.Elems.At(i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		cmp, err := left.Compare(right)
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *SetEnumValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	var o *SetEnumValue
	if !isNil(other) {
		o, err = tryToSetEnumValue(other)
	}
	if err != nil {
		return false, err
	}
	if o == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueEquals(v)
		}
		setText := ValuesPPR(v)
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of the set %s with the value:\n%s", setText, ValuesPPR(other)))
	}
	if _, err := v.normalizeSet(); err != nil {
		return false, err
	}
	if _, err := o.normalizeSet(); err != nil {
		return false, err
	}
	if v.Elems.Len() != o.Elems.Len() {
		return false, nil
	}
	for i := 0; i < v.Elems.Len(); i++ {
		left, right := v.Elems.At(i), o.Elems.At(i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		eq, err := left.Equal(right)
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *SetEnumValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return v.Elems.Search(elem, v.IsNorm)
}

func (v *SetEnumValue) IsFinite() (bool, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true, nil
}

func (v *SetEnumValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if _, err := v.normalizeSet(); err != nil {
		return 0, err
	}
	return v.Elems.Len(), nil
}

func (v *SetEnumValue) RandomElement() (Value, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	size, err := v.Size()
	if err != nil {
		return nil, err
	}
	rng := RandomEnumerableGenerator()
	if rng == nil {
		panic(NewNullPointerException())
	}
	index := javaDoubleToInt(math.Floor(rng.NextDouble() * float64(size)))
	return v.Elems.At(int(index)), nil
}

func (v *SetEnumValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.IsNorm
}
func (v *SetEnumValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	normalized, err := v.normalizeSet()
	if err != nil {
		panic(err)
	}
	return normalized
}

func (v *SetEnumValue) normalizeSet() (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if !v.IsNorm {
		if err := v.Elems.Sort(true); err != nil {
			return nil, err
		}
		v.IsNorm = true
	}
	return v, nil
}

func (v *SetEnumValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	for i := 0; i < v.Elems.Len(); i++ {
		if isNil(v.Elems.At(i)) {
			panic(NewNullPointerException())
		}
		v.Elems.At(i).DeepNormalize()
	}
	if _, err := v.normalizeSet(); err != nil {
		panic(err)
	}
}

// ToSetEnum ports SetEnumValue.toSetEnum. An enumerated set is already the
// concrete representation requested by the conversion, so identity is exact.
func (v *SetEnumValue) ToSetEnum() *SetEnumValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *SetEnumValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	for i := 0; i < v.Elems.Len(); i++ {
		if isNil(v.Elems.At(i)) {
			panic(NewNullPointerException())
		}
		if !v.Elems.At(i).IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetEnumValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *SetEnumValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	out := make([]Value, v.Elems.Len())
	changed := false
	for i := range out {
		elem := v.Elems.At(i)
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		out[i] = elem.Permute(perm)
		changed = changed || out[i] != elem
	}
	if changed {
		return NewSetEnumValue(out, false)
	}
	return v
}

func (v *SetEnumValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if _, err := v.normalizeSet(); err != nil {
		panic(err)
	}
	fp = FP64ExtendByte(fp, byte(SetEnumValueKind))
	fp = FP64ExtendInt(fp, int32(v.Elems.Len()))
	for i := 0; i < v.Elems.Len(); i++ {
		if isNil(v.Elems.At(i)) {
			panic(NewNullPointerException())
		}
		fp = v.Elems.At(i).FingerPrint(fp)
	}
	return fp
}

func (v *SetEnumValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT to the set %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *SetEnumValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT to the set %s.", ValuesPPR(v)))
	}
	return v, nil
}

func (v *SetEnumValue) ToTupleValue() *TupleValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	if _, err := v.normalizeSet(); err != nil {
		panic(err)
	}
	return NewTupleValue(v.Elems.ToArray())
}

func (v *SetEnumValue) Elements() ValueEnumeration {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if _, err := v.normalizeSet(); err != nil {
		panic(err)
	}
	return &setEnumEnumeration{owner: v}
}

func (v *SetEnumValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *SetEnumValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if !v.IsNormalized() {
		if _, err := v.normalizeSet(); err != nil {
			panic(err)
		}
	}
	length := v.Elems.Len()
	sb.WriteString("{")
	for i := 0; i < length; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		element := v.Elems.At(i)
		if isNil(element) {
			panic(NewNullPointerException())
		}
		appendValueString(element, sb, offset, swallow)
	}
	sb.WriteString("}")
	return sb
}

type IntervalValue struct {
	BaseValue
	Low  int32
	High int32
}

func (v *IntervalValue) Mutates() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return false
}

func NewIntervalValue(low, high int32) *IntervalValue {
	return &IntervalValue{Low: low, High: high}
}

func (v *IntervalValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return IntervalValueKind
}
func (v *IntervalValue) KindString() string { return ValueKindString(v) }
func (v *IntervalValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *IntervalValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if o, ok := other.(*IntervalValue); ok {
		sz, err := v.Size()
		if err != nil {
			return 0, err
		}
		osz, err := o.Size()
		if err != nil {
			return 0, err
		}
		if sz != osz {
			return sz - osz, nil
		}
		if sz == 0 {
			return 0, nil
		}
		if v.Low < o.Low {
			return -1, nil
		}
		if v.Low > o.Low {
			return 1, nil
		}
		return 0, nil
	}
	return v.ToSetEnum().Compare(other)
}

func (v *IntervalValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if o, ok := other.(*IntervalValue); ok {
		sz, err := v.Size()
		if err != nil {
			return false, err
		}
		osz, err := o.Size()
		if err != nil {
			return false, err
		}
		if sz == 0 {
			return osz == 0, nil
		}
		return v.Low == o.Low && v.High == o.High, nil
	}
	return v.ToSetEnum().Equal(other)
}

func (v *IntervalValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	i, ok := elem.(*IntValue)
	if !ok {
		if v.Low <= v.High {
			if mv, ok := elem.(*ModelValue); ok && mv.Type == typedModelValueUntypedCodeUnit {
				return false, nil
			}
			if elem == nil {
				panic(NewNullPointerException())
			}
			return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis in the integer interval %s", ValuesPPR(elem), ValuesPPR(v)))
		}
		return false, nil
	}
	return i.ownedVal() >= v.Low && i.ownedVal() <= v.High, nil
}

func (v *IntervalValue) IsSubsetEq(other Value) (resultBool *BoolValue, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if interval, ok := other.(*IntervalValue); ok && interval != nil && interval.Low <= v.Low && interval.High >= v.High {
		return BoolTrue, nil
	}
	return defaultEnumerableSubsetEq(v, v, other)
}

func (v *IntervalValue) IsFinite() (bool, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true, nil
}

func (v *IntervalValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if v.High < v.Low {
		return 0, nil
	}
	size := int64(v.High) - int64(v.Low) + 1
	if size > math.MaxInt32 {
		return 0, v.intervalAssertFailure("Size of interval value exceeds the maximum representable size (32bits): " + ValuesPPR(v) + ".")
	}
	return int(size), nil
}

// ElementAt ports IntervalValue.elementAt, including its short-circuit bounds
// check: a negative index must not evaluate size() on an overflowing interval.
func (v *IntervalValue) ElementAt(index int) (Value, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	index = int(int32(index))
	if index >= 0 {
		size, err := v.Size()
		if err != nil {
			return nil, err
		}
		if index < size {
			return NewIntValue(v.Low + int32(index)), nil
		}
	}
	return nil, v.intervalAssertFailure("Attempted to retrieve out-of-bounds element from the interval value " + ValuesPPR(v) + ".")
}

func (v *IntervalValue) RandomElement() (Value, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	size, err := v.Size()
	if err != nil {
		return nil, err
	}
	rng := RandomEnumerableGenerator()
	if rng == nil {
		panic(NewNullPointerException())
	}
	index := javaDoubleToInt(math.Floor(rng.NextDouble() * float64(size)))
	return v.ElementAt(int(index))
}

func (v *IntervalValue) intervalAssertFailure(message string) *TLCError {
	if source := v.GetSource(); source != nil {
		return NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
	}
	failure := newTLCError(ECGeneral, "%s", message)
	failure.Runtime = true
	return failure
}

func (v *IntervalValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *IntervalValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
}
func (v *IntervalValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *IntervalValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *IntervalValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *IntervalValue) Permute(*MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *IntervalValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	sz, err := v.Size()
	if err != nil {
		panic(err)
	}
	fp = FP64ExtendByte(fp, byte(SetEnumValueKind))
	fp = FP64ExtendInt(fp, int32(sz))
	for i := 0; i < sz; i++ {
		fp = FP64ExtendByte(fp, byte(IntValueKind))
		fp = FP64ExtendInt(fp, v.Low+int32(i))
	}
	return fp
}

func (v *IntervalValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the interval value %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *IntervalValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the interval value %s.", ValuesPPR(v)))
	}
	return v, nil
}

func (v *IntervalValue) ToSetEnum() *SetEnumValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	sz, err := v.Size()
	if err != nil {
		panic(err)
	}
	values := make([]Value, sz)
	for i := 0; i < sz; i++ {
		values[i] = NewIntValue(v.Low + int32(i))
	}
	v.CM.incValueSecondary(int64(sz))
	return NewSetEnumValue(values, true, v.CM)
}

func (v *IntervalValue) AsValues() []Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	sz, err := v.Size()
	if err != nil {
		panic(err)
	}
	values := make([]Value, sz)
	for i := 0; i < sz; i++ {
		values[i] = NewIntValue(v.Low + int32(i))
	}
	return values
}

func (v *IntervalValue) Elements() ValueEnumeration {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	enum := newIntervalValueEnumeration(v.Low, v.High)
	enum.cm = v.CM
	return enum
}

func (v *IntervalValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *IntervalValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.Low <= v.High {
		sb.WriteString(fmt.Sprintf("%d..%d", v.Low, v.High))
	} else {
		sb.WriteString("{}")
	}
	return sb
}

type intervalValueEnumeration struct {
	cm    CostModel
	low   int32
	high  int32
	index int32
	done  bool
}

func newIntervalValueEnumeration(low int32, high int32) *intervalValueEnumeration {
	e := &intervalValueEnumeration{low: low, high: high}
	e.Reset()
	return e
}

func (e *intervalValueEnumeration) Reset() {
	requireValueEnumerationReceiver(e)
	e.index = e.low
	e.done = e.high < e.low
}

func (e *intervalValueEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
	if e.done {
		return nil
	}
	e.cm.incValueSecondary()
	current := e.index
	if current == e.high {
		e.done = true
	} else {
		e.index++
	}
	return NewIntValue(current)
}

func (e *intervalValueEnumeration) Err() error {
	return nil
}

// setEnumEnumeration retains the set owner: Java's Enumerator reads the
// current elems vector on every call, including after reset or replacement.
type setEnumEnumeration struct {
	owner *SetEnumValue
	index int
}

func (e *setEnumEnumeration) Reset() {
	requireValueEnumerationReceiver(e)
	e.index = 0
}

func (e *setEnumEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
	e.owner.CM.incValueSecondary()
	if e.owner.Elems == nil {
		panic(NewNullPointerException())
	}
	if e.index >= e.owner.Elems.Len() {
		return nil
	}
	value := e.owner.Elems.At(e.index)
	e.index++
	return value
}

func (e *setEnumEnumeration) Err() error { return nil }

type sliceValueEnumeration struct {
	cm     CostModel
	values []Value
	index  int
}

func (e *sliceValueEnumeration) Reset() {
	requireValueEnumerationReceiver(e)
	e.index = 0
}

func (e *sliceValueEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
	e.cm.incValueSecondary()
	if e.index >= len(e.values) {
		return nil
	}
	value := e.values[e.index]
	e.index++
	return value
}

func (e *sliceValueEnumeration) Err() error {
	return nil
}

func randomizedValueEnumeration(enumerable Enumerable) (ValueEnumeration, error) {
	value := enumerable.(Value)
	if subset, ok := value.(*KSubsetValue); ok {
		return newRandomKSubsetEnumeration(subset)
	}
	if _, ok := value.(*SubsetValue); ok {
		// SubsetValue's ordering overload materializes, unlike elements(k).
		set, err := toSetEnumValue(value)
		if err != nil {
			return nil, err
		}
		return randomizedValueEnumeration(set)
	}
	n, err := value.Size()
	if err != nil {
		return nil, err
	}
	return randomValueEnumeration(value, n)
}
