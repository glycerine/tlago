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

type Enumerable interface {
	Elements() ValueEnumeration
}

type ValueVec struct {
	data []Value
}

func NewValueVec(capacity int) *ValueVec {
	if capacity < 0 {
		panic(fmt.Errorf("%d", capacity))
	}
	return &ValueVec{data: make([]Value, 0, capacity)}
}

func NewValueVecFrom(values []Value) *ValueVec {
	return &ValueVec{data: values}
}

func (v *ValueVec) Add(val Value) {
	if len(v.data) == cap(v.data) {
		v.ensureCapacity(len(v.data) + 1)
	}
	v.data = append(v.data, val)
}

func (v *ValueVec) AddAt(val Value, index int) {
	oldLen := len(v.data)
	newLen := oldLen + 1
	if index < 0 || index >= cap(v.data) || newLen > cap(v.data) {
		panic("ValueVec index out of bounds")
	}
	if index >= len(v.data) {
		v.data = v.data[:index+1]
	}
	v.data[index] = val
	if len(v.data) != newLen {
		v.data = v.data[:newLen]
	}
}

func (v *ValueVec) AddSortedUnique(val Value) error {
	for i, elem := range v.data {
		cmp, err := elem.Compare(val)
		if err != nil {
			return err
		}
		if cmp == 0 {
			return nil
		}
		if cmp > 0 {
			v.insertAt(val, i)
			return nil
		}
	}
	v.Add(val)
	return nil
}

func (v *ValueVec) insertAt(val Value, index int) {
	if len(v.data) == cap(v.data) {
		v.ensureCapacity(len(v.data) + 1)
	}
	var zero Value
	v.data = append(v.data, zero)
	copy(v.data[index+1:], v.data[index:len(v.data)-1])
	v.data[index] = val
}

func (v *ValueVec) Len() int             { return len(v.data) }
func (v *ValueVec) Cap() int             { return cap(v.data) }
func (v *ValueVec) Empty() bool          { return len(v.data) == 0 }
func (v *ValueVec) At(i int) Value       { return v.data[i] }
func (v *ValueVec) Set(i int, val Value) { v.data[i] = val }
func (v *ValueVec) First() Value         { return v.data[0] }
func (v *ValueVec) Last() Value          { return v.data[len(v.data)-1] }

func (v *ValueVec) ToArray() []Value {
	out := make([]Value, len(v.data))
	copy(out, v.data)
	return out
}

func (v *ValueVec) Contains(val Value) bool {
	return v.IndexOf(val) != -1
}

func (v *ValueVec) IndexOf(val Value) int {
	for i, elem := range v.data {
		eq, err := val.Equal(elem)
		if err == nil && eq {
			return i
		}
	}
	return -1
}

func (v *ValueVec) Search(val Value, sorted bool) (bool, error) {
	if sorted {
		low, high := 0, len(v.data)
		for low < high {
			mid := (low + high) >> 1
			if val == nil {
				panic(NewNullPointerException())
			}
			cmp, err := val.Compare(v.data[mid])
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
	for i := 0; i < len(v.data); i++ {
		equal, err := v.data[i].Equal(val)
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
	if len(v.data) != 0 {
		newCount = 1
	}
	for i := 1; i < len(v.data); i++ {
		elem := v.data[i]
		cmp := 0
		idx := 0
		low, high := 0, newCount
		for low < high {
			idx = (low + high) >> 1
			var err error
			if elem == nil {
				panic(NewNullPointerException())
			}
			cmp, err = elem.Compare(v.data[idx])
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
				v.data[j] = v.data[j-1]
			}
			v.data[idx] = elem
			newCount++
		}
	}
	v.data = v.data[:newCount]
	return nil
}

func (v *ValueVec) String() string {
	parts := make([]string, len(v.data))
	for i, elem := range v.data {
		parts[i] = elem.String()
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (v *ValueVec) ensureCapacity(minCapacity int) {
	if cap(v.data) >= Globals.SetBound {
		panic(newTLCError(ECGeneral, "Attempted to construct a set with too many elements (>%d).", Globals.SetBound))
	}
	if cap(v.data) >= minCapacity {
		return
	}
	newCapacity := cap(v.data) + cap(v.data)
	if newCapacity < minCapacity {
		newCapacity = minCapacity
	}
	if newCapacity > Globals.SetBound {
		newCapacity = Globals.SetBound
	}
	out := make([]Value, len(v.data), newCapacity)
	copy(out, v.data)
	v.data = out
}

type TupleValue struct {
	BaseValue
	Elems []Value
}

var EmptyTuple = &TupleValue{Elems: []Value{}}

func NewTupleValue(elems []Value, cms ...CostModel) *TupleValue {
	return &TupleValue{BaseValue: newBaseValue(cms...), Elems: elems}
}

func (v *TupleValue) Kind() ValueKind    { return TupleValueKind }
func (v *TupleValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *TupleValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	o := asTupleValue(other)
	if o == nil {
		return asFcnRcdValue(v).Compare(other)
	}
	if v.Elems == nil || o.Elems == nil {
		panic(NewNullPointerException())
	}
	if len(v.Elems) != len(o.Elems) {
		return len(v.Elems) - len(o.Elems), nil
	}
	for i := range v.Elems {
		if v.Elems[i] == nil {
			panic(NewNullPointerException())
		}
		cmp, err := v.Elems[i].Compare(o.Elems[i])
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *TupleValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	o := asTupleValue(other)
	if o == nil {
		return asFcnRcdValue(v).Equal(other)
	}
	if v.Elems == nil || o.Elems == nil {
		panic(NewNullPointerException())
	}
	if len(v.Elems) != len(o.Elems) {
		return false, nil
	}
	for i := range v.Elems {
		if v.Elems[i] == nil {
			panic(NewNullPointerException())
		}
		eq, err := v.Elems[i].Equal(o.Elems[i])
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *TupleValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure("Attempted to check set membership in a tuple value.")
}

func (v *TupleValue) IsFinite() (bool, error) { return true, nil }
func (v *TupleValue) Size() (int, error) {
	if v.Elems == nil {
		return 0, NewNullPointerException()
	}
	return len(v.Elems), nil
}
func (v *TupleValue) Normalize() Value   { return v }
func (v *TupleValue) IsNormalized() bool { return true }

func (v *TupleValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	if v.Elems == nil {
		panic(NewNullPointerException())
	}
	for _, elem := range v.Elems {
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		elem.DeepNormalize()
	}
}

func (v *TupleValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	if v.Elems == nil {
		panic(NewNullPointerException())
	}
	for _, elem := range v.Elems {
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
	defer catchValueFailure(v, nil)
	if v.Elems == nil {
		panic(NewNullPointerException())
	}
	out := make([]Value, len(v.Elems))
	for i, elem := range v.Elems {
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		out[i] = elem.DeepCopy()
	}
	return &TupleValue{Elems: out}
}

func (v *TupleValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	if v.Elems == nil {
		panic(NewNullPointerException())
	}
	out := make([]Value, len(v.Elems))
	changed := false
	for i, elem := range v.Elems {
		if isNil(elem) {
			panic(NewNullPointerException())
		}
		out[i] = elem.Permute(perm)
		changed = changed || out[i] != elem
	}
	if changed {
		return &TupleValue{Elems: out}
	}
	return v
}

func (v *TupleValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	if v.Elems == nil {
		panic(NewNullPointerException())
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.Elems)))
	for i, elem := range v.Elems {
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
	defer catchValueFailure(v, &err)
	if arg == nil {
		panic(NewNullPointerException())
	}
	i, ok := arg.(*IntValue)
	if !ok {
		return nil, v.tupleAssertFailure("Attempted to access tuple at a non integral index: " + ValuesPPR(arg))
	}
	idx := int(i.Val)
	if idx > 0 && v.Elems == nil {
		panic(NewNullPointerException())
	}
	if idx <= 0 || idx > len(v.Elems) {
		return nil, v.tupleAssertFailure(fmt.Sprintf("Attempted to access index %d of tuple\n%s\nwhich is out of bounds.", idx, ValuesPPR(v)))
	}
	return v.Elems[idx-1], nil
}

// ApplyArgs ports TupleValue.apply(Value[], int). Java discards control when
// delegating to the single-argument overload, whose result is control-independent.
func (v *TupleValue) ApplyArgs(args []Value, control int) (resultValue Value, err error) {
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
	defer catchValueFailure(v, &err)
	if isNil(arg) {
		panic(NewNullPointerException())
	}
	i, ok := arg.(*IntValue)
	if !ok {
		return nil, v.tupleAssertFailure("Attempted to access tuple at a non integral index: " + ValuesPPR(arg))
	}
	idx := int(i.Val)
	if idx > 0 && v.Elems == nil {
		panic(NewNullPointerException())
	}
	if idx > 0 && idx <= len(v.Elems) {
		return v.Elems[idx-1], nil
	}
	return nil, nil
}

func (v *TupleValue) Domain() Value {
	defer catchValueFailure(v, nil)
	size, err := v.Size()
	if err != nil {
		panic(err)
	}
	return NewIntervalValue(1, int32(size))
}

func (v *TupleValue) ToFcnRcd() *FcnRcdValue {
	size, err := v.Size()
	if err != nil {
		panic(err)
	}
	v.CM.incValueSecondary(int64(size))
	return NewFcnRcdIntervalValue(NewIntervalValue(1, int32(size)), v.Elems, v.CM)
}

func (v *TupleValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
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
	if v.Elems == nil {
		panic(NewNullPointerException())
	}
	out := make([]Value, len(v.Elems))
	arcVal := ex.Current()
	arc, ok := arcVal.(*IntValue)
	if !ok {
		if arcVal == nil {
			panic(NewNullPointerException())
		}
		PrintWarning(ECTLCWrongTupleFieldName, ValuesPPR(arcVal))
		return ex.Value, nil
	}
	idx := int(arc.Val - 1)
	// Java allocates the replacement tuple before checking bounds and returns it
	// even when the integer field is out of range, leaving the slots unfilled.
	if 0 <= idx && idx < len(v.Elems) {
		copy(out, v.Elems)
		next := ex
		next.Index++
		if v.Elems[idx] == nil {
			panic(NewNullPointerException())
		}
		val, err := v.Elems[idx].TakeExcept(next)
		if err != nil {
			return nil, err
		}
		out[idx] = val
	}
	return &TupleValue{Elems: out}, nil
}

func (v *TupleValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	var cur Value = v
	for _, ex := range exs {
		if cur == nil {
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
	return ValueToString(v, "", true)
}

func (v *TupleValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	sb.WriteString("<<")
	if v.Elems == nil {
		panic(NewNullPointerException())
	}
	for i, elem := range v.Elems {
		if i > 0 {
			sb.WriteString(", ")
		}
		if elem == nil {
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

var EmptySet = &SetEnumValue{Elems: NewValueVec(0), IsNorm: true}

func NewSetEnumValue(values []Value, isNorm bool, cms ...CostModel) *SetEnumValue {
	return &SetEnumValue{BaseValue: newBaseValue(cms...), Elems: NewValueVecFrom(values), IsNorm: isNorm}
}

func NewSetEnumValueVec(values *ValueVec, isNorm bool, cms ...CostModel) *SetEnumValue {
	if values == nil {
		values = NewValueVec(0)
	}
	return &SetEnumValue{BaseValue: newBaseValue(cms...), Elems: values, IsNorm: isNorm}
}

func (v *SetEnumValue) Kind() ValueKind    { return SetEnumValueKind }
func (v *SetEnumValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetEnumValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	o, err := tryToSetEnumValue(other)
	if err != nil {
		return 0, err
	}
	if o == nil {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the set %s with the value:\n%s", ValuesPPR(v), ValuesPPR(other)))
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
		cmp, err := v.Elems.At(i).Compare(o.Elems.At(i))
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *SetEnumValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	o, err := tryToSetEnumValue(other)
	if err != nil {
		return false, err
	}
	if o == nil {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of the set %s with the value:\n%s", ValuesPPR(v), ValuesPPR(other)))
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
		eq, err := v.Elems.At(i).Equal(o.Elems.At(i))
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *SetEnumValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return v.Elems.Search(elem, v.IsNorm)
}

func (v *SetEnumValue) IsFinite() (bool, error) { return true, nil }

func (v *SetEnumValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if _, err := v.normalizeSet(); err != nil {
		return 0, err
	}
	return v.Elems.Len(), nil
}

func (v *SetEnumValue) IsNormalized() bool { return v.IsNorm }
func (v *SetEnumValue) Normalize() Value {
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
	defer catchValueFailure(v, nil)
	for i := 0; i < v.Elems.Len(); i++ {
		if v.Elems.At(i) == nil {
			panic(NewNullPointerException())
		}
		v.Elems.At(i).DeepNormalize()
	}
	if _, err := v.normalizeSet(); err != nil {
		panic(err)
	}
}

func (v *SetEnumValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	for i := 0; i < v.Elems.Len(); i++ {
		if !v.Elems.At(i).IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetEnumValue) DeepCopy() Value { return v }

func (v *SetEnumValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	out := make([]Value, v.Elems.Len())
	changed := false
	for i := range out {
		elem := v.Elems.At(i)
		out[i] = elem.Permute(perm)
		changed = changed || out[i] != elem
	}
	if changed {
		return NewSetEnumValue(out, false)
	}
	return v
}

func (v *SetEnumValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	if _, err := v.normalizeSet(); err != nil {
		panic(err)
	}
	fp = FP64ExtendByte(fp, byte(SetEnumValueKind))
	fp = FP64ExtendInt(fp, int32(v.Elems.Len()))
	for i := 0; i < v.Elems.Len(); i++ {
		fp = v.Elems.At(i).FingerPrint(fp)
	}
	return fp
}

func (v *SetEnumValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
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
	if _, err := v.normalizeSet(); err != nil {
		panic(err)
	}
	return NewTupleValue(v.Elems.ToArray())
}

func (v *SetEnumValue) Elements() ValueEnumeration {
	defer catchValueFailure(v, nil)
	if _, err := v.normalizeSet(); err != nil {
		return newErrorEnumeration(wrapValueFailure(v, err))
	}
	return &setEnumEnumeration{owner: v}
}

func (v *SetEnumValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetEnumValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
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
		appendValueString(v.Elems.At(i), sb, offset, swallow)
	}
	sb.WriteString("}")
	return sb
}

type IntervalValue struct {
	BaseValue
	Low  int32
	High int32
}

func NewIntervalValue(low, high int32) *IntervalValue {
	return &IntervalValue{Low: low, High: high}
}

func (v *IntervalValue) Kind() ValueKind    { return IntervalValueKind }
func (v *IntervalValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *IntervalValue) Compare(other Value) (resultInt int, err error) {
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
	return i.Val >= v.Low && i.Val <= v.High, nil
}

func (v *IntervalValue) IsFinite() (bool, error) { return true, nil }

func (v *IntervalValue) Size() (resultInt int, err error) {
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

func (v *IntervalValue) intervalAssertFailure(message string) *TLCError {
	if source := v.GetSource(); source != nil {
		return NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
	}
	failure := newTLCError(ECGeneral, "%s", message)
	failure.Runtime = true
	return failure
}

func (v *IntervalValue) Normalize() Value      { return v }
func (v *IntervalValue) DeepNormalize()        {}
func (v *IntervalValue) IsNormalized() bool    { return true }
func (v *IntervalValue) IsDefined() bool       { return true }
func (v *IntervalValue) DeepCopy() Value       { return v }
func (v *IntervalValue) Permute(*MVPerm) Value { return v }

func (v *IntervalValue) FingerPrint(fp uint64) uint64 {
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
	defer catchValueFailure(v, nil)
	enum := newIntervalValueEnumeration(v.Low, v.High)
	enum.cm = v.CM
	return enum
}

func (v *IntervalValue) String() string {
	return ValueToString(v, "", true)
}

func (v *IntervalValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
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
	e.index = e.low
	e.done = e.high < e.low
}

func (e *intervalValueEnumeration) NextElement() Value {
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

func (e *setEnumEnumeration) Reset() { e.index = 0 }

func (e *setEnumEnumeration) NextElement() Value {
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
	e.index = 0
}

func (e *sliceValueEnumeration) NextElement() Value {
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
