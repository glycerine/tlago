package tlc

import (
	"math"
	"sort"
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
		capacity = 0
	}
	return &ValueVec{data: make([]Value, 0, capacity)}
}

func NewValueVecFrom(values []Value) *ValueVec {
	out := make([]Value, len(values))
	copy(out, values)
	return &ValueVec{data: out}
}

func (v *ValueVec) Add(val Value) {
	v.data = append(v.data, val)
}

func (v *ValueVec) AddAt(val Value, index int) {
	if index == len(v.data) {
		v.data = append(v.data, val)
		return
	}
	v.data = append(v.data[:index+1], v.data[index:]...)
	v.data[index] = val
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
			v.AddAt(val, i)
			return nil
		}
	}
	v.Add(val)
	return nil
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

func (v *ValueVec) Search(val Value, sorted bool) bool {
	if sorted {
		low, high := 0, len(v.data)
		for low < high {
			mid := (low + high) >> 1
			cmp, err := val.Compare(v.data[mid])
			if err != nil {
				return false
			}
			if cmp == 0 {
				return true
			}
			if cmp < 0 {
				high = mid
			} else {
				low = mid + 1
			}
		}
		return false
	}
	return v.Contains(val)
}

func (v *ValueVec) Sort(noDup bool) error {
	var sortErr error
	sort.SliceStable(v.data, func(i, j int) bool {
		if sortErr != nil {
			return false
		}
		cmp, err := v.data[i].Compare(v.data[j])
		if err != nil {
			sortErr = err
			return false
		}
		return cmp < 0
	})
	if sortErr != nil {
		return sortErr
	}
	if noDup && len(v.data) > 1 {
		out := v.data[:1]
		for _, elem := range v.data[1:] {
			eq, err := elem.Equal(out[len(out)-1])
			if err != nil {
				return err
			}
			if !eq {
				out = append(out, elem)
			}
		}
		v.data = out
	}
	return nil
}

func (v *ValueVec) String() string {
	parts := make([]string, len(v.data))
	for i, elem := range v.data {
		parts[i] = elem.String()
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

type TupleValue struct {
	BaseValue
	Elems []Value
}

var EmptyTuple = &TupleValue{Elems: []Value{}}

func NewTupleValue(elems []Value) *TupleValue {
	out := make([]Value, len(elems))
	copy(out, elems)
	return &TupleValue{Elems: out}
}

func (v *TupleValue) Kind() ValueKind    { return TupleValueKind }
func (v *TupleValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *TupleValue) Compare(other Value) (int, error) {
	o, ok := other.(*TupleValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		return 0, v.unsupported("attempted to compare tuple %s with non-tuple %s", v, other)
	}
	if len(v.Elems) != len(o.Elems) {
		return len(v.Elems) - len(o.Elems), nil
	}
	for i := range v.Elems {
		cmp, err := v.Elems[i].Compare(o.Elems[i])
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *TupleValue) Equal(other Value) (bool, error) {
	o, ok := other.(*TupleValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		return false, v.unsupported("attempted to compare equality of tuple %s with non-tuple %s", v, other)
	}
	if len(v.Elems) != len(o.Elems) {
		return false, nil
	}
	for i := range v.Elems {
		eq, err := v.Elems[i].Equal(o.Elems[i])
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *TupleValue) Member(elem Value) (bool, error) {
	return false, v.unsupported("attempted to check set membership in a tuple value")
}

func (v *TupleValue) IsFinite() (bool, error) { return true, nil }
func (v *TupleValue) Size() (int, error)      { return len(v.Elems), nil }
func (v *TupleValue) Normalize() Value        { return v }
func (v *TupleValue) IsNormalized() bool      { return true }

func (v *TupleValue) DeepNormalize() {
	for _, elem := range v.Elems {
		elem.DeepNormalize()
	}
}

func (v *TupleValue) IsDefined() bool {
	for _, elem := range v.Elems {
		if !elem.IsDefined() {
			return false
		}
	}
	return true
}

func (v *TupleValue) DeepCopy() Value {
	out := make([]Value, len(v.Elems))
	for i, elem := range v.Elems {
		out[i] = elem.DeepCopy()
	}
	return &TupleValue{Elems: out}
}

func (v *TupleValue) Permute(perm *MVPerm) Value {
	out := make([]Value, len(v.Elems))
	changed := false
	for i, elem := range v.Elems {
		out[i] = elem.Permute(perm)
		changed = changed || out[i] != elem
	}
	if changed {
		return &TupleValue{Elems: out}
	}
	return v
}

func (v *TupleValue) FingerPrint(fp uint64) uint64 {
	fp = FP64ExtendInt(fp, int32(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.Elems)))
	for i, elem := range v.Elems {
		fp = FP64ExtendInt(fp, int32(IntValueKind))
		fp = FP64ExtendInt(fp, int32(i+1))
		fp = elem.FingerPrint(fp)
	}
	return fp
}

func (v *TupleValue) Apply(arg Value) (Value, error) {
	i, ok := arg.(*IntValue)
	if !ok {
		return nil, v.unsupported("attempted to access tuple at a non integral index: %s", arg)
	}
	idx := int(i.Val)
	if idx <= 0 || idx > len(v.Elems) {
		return nil, v.unsupported("attempted to access index %d of tuple %s which is out of bounds", idx, v)
	}
	return v.Elems[idx-1], nil
}

func (v *TupleValue) Select(arg Value) Value {
	i, ok := arg.(*IntValue)
	if !ok {
		return nil
	}
	idx := int(i.Val)
	if idx > 0 && idx <= len(v.Elems) {
		return v.Elems[idx-1]
	}
	return nil
}

func (v *TupleValue) Domain() Value {
	return NewIntervalValue(1, int32(len(v.Elems)))
}

func (v *TupleValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index >= len(ex.Path) {
		return ex.Value, nil
	}
	out := make([]Value, len(v.Elems))
	arc, ok := ex.Path[ex.Index].(*IntValue)
	if !ok {
		PrintWarning(ECTLCWrongTupleFieldName, valueString(ex.Path[ex.Index]))
		return ex.Value, nil
	}
	idx := int(arc.Val) - 1
	// Java allocates the replacement tuple before checking bounds and returns it
	// even when the integer field is out of range, leaving the slots unfilled.
	if 0 <= idx && idx < len(v.Elems) {
		copy(out, v.Elems)
		next := ex
		next.Index++
		val, err := v.Elems[idx].TakeExcept(next)
		if err != nil {
			return nil, err
		}
		out[idx] = val
	}
	return &TupleValue{Elems: out}, nil
}

func (v *TupleValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	var cur Value = v
	for _, ex := range exs {
		next, err := cur.TakeExcept(ex)
		if err != nil {
			return nil, err
		}
		cur = next
	}
	return cur, nil
}

func (v *TupleValue) String() string {
	parts := make([]string, len(v.Elems))
	for i, elem := range v.Elems {
		parts[i] = elem.String()
	}
	return "<<" + strings.Join(parts, ", ") + ">>"
}

type SetEnumValue struct {
	BaseValue
	Elems  *ValueVec
	IsNorm bool
}

var EmptySet = &SetEnumValue{Elems: NewValueVec(0), IsNorm: true}

func NewSetEnumValue(values []Value, isNorm bool) *SetEnumValue {
	return &SetEnumValue{Elems: NewValueVecFrom(values), IsNorm: isNorm}
}

func NewSetEnumValueVec(values *ValueVec, isNorm bool) *SetEnumValue {
	if values == nil {
		values = NewValueVec(0)
	}
	return &SetEnumValue{Elems: values, IsNorm: isNorm}
}

func (v *SetEnumValue) Kind() ValueKind    { return SetEnumValueKind }
func (v *SetEnumValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetEnumValue) Compare(other Value) (int, error) {
	o, ok := other.(*SetEnumValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		return 0, v.unsupported("attempted to compare the set %s with the value %s", v, other)
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

func (v *SetEnumValue) Equal(other Value) (bool, error) {
	o, ok := other.(*SetEnumValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		return false, v.unsupported("attempted to check equality of the set %s with the value %s", v, other)
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

func (v *SetEnumValue) Member(elem Value) (bool, error) {
	return v.Elems.Search(elem, v.IsNorm), nil
}

func (v *SetEnumValue) IsFinite() (bool, error) { return true, nil }

func (v *SetEnumValue) Size() (int, error) {
	if _, err := v.normalizeSet(); err != nil {
		return 0, err
	}
	return v.Elems.Len(), nil
}

func (v *SetEnumValue) IsNormalized() bool { return v.IsNorm }
func (v *SetEnumValue) Normalize() Value   { normalized, _ := v.normalizeSet(); return normalized }

func (v *SetEnumValue) normalizeSet() (Value, error) {
	if !v.IsNorm {
		if err := v.Elems.Sort(true); err != nil {
			return nil, err
		}
		v.IsNorm = true
	}
	return v, nil
}

func (v *SetEnumValue) DeepNormalize() {
	for i := 0; i < v.Elems.Len(); i++ {
		v.Elems.At(i).DeepNormalize()
	}
	_, _ = v.normalizeSet()
}

func (v *SetEnumValue) IsDefined() bool {
	for i := 0; i < v.Elems.Len(); i++ {
		if !v.Elems.At(i).IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetEnumValue) DeepCopy() Value { return v }

func (v *SetEnumValue) Permute(perm *MVPerm) Value {
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
	_, _ = v.normalizeSet()
	fp = FP64ExtendInt(fp, int32(SetEnumValueKind))
	fp = FP64ExtendInt(fp, int32(v.Elems.Len()))
	for i := 0; i < v.Elems.Len(); i++ {
		fp = v.Elems.At(i).FingerPrint(fp)
	}
	return fp
}

func (v *SetEnumValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, v.unsupported("attempted to apply EXCEPT to the set %s", v)
	}
	return ex.Value, nil
}

func (v *SetEnumValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, v.unsupported("attempted to apply EXCEPT to the set %s", v)
	}
	return v, nil
}

func (v *SetEnumValue) ToTupleValue() *TupleValue {
	_, _ = v.normalizeSet()
	return NewTupleValue(v.Elems.ToArray())
}

func (v *SetEnumValue) Elements() ValueEnumeration {
	return &sliceValueEnumeration{values: v.Elems.ToArray()}
}

func (v *SetEnumValue) String() string {
	_, _ = v.normalizeSet()
	return v.Elems.String()
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

func (v *IntervalValue) Compare(other Value) (int, error) {
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
	if mv, ok := other.(*ModelValue); ok {
		return mv.modelValueCompareTo(v)
	}
	return v.ToSetEnum().Compare(other)
}

func (v *IntervalValue) Equal(other Value) (bool, error) {
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
	if mv, ok := other.(*ModelValue); ok {
		return mv.modelValueEquals(v)
	}
	return v.ToSetEnum().Equal(other)
}

func (v *IntervalValue) Member(elem Value) (bool, error) {
	i, ok := elem.(*IntValue)
	if !ok {
		sz, err := v.Size()
		if err != nil {
			return false, err
		}
		if sz > 0 {
			return false, v.unsupported("attempted to check if %s is in the integer interval %s", elem, v)
		}
		return false, nil
	}
	return i.Val >= v.Low && i.Val <= v.High, nil
}

func (v *IntervalValue) IsFinite() (bool, error) { return true, nil }

func (v *IntervalValue) Size() (int, error) {
	if v.High < v.Low {
		return 0, nil
	}
	size := int64(v.High) - int64(v.Low) + 1
	if size > math.MaxInt32 {
		return 0, newTLCError(ECGeneral, "Size of interval value exceeds the maximum representable size (32bits)")
	}
	return int(size), nil
}

func (v *IntervalValue) Normalize() Value      { return v }
func (v *IntervalValue) DeepNormalize()        {}
func (v *IntervalValue) IsNormalized() bool    { return true }
func (v *IntervalValue) IsDefined() bool       { return true }
func (v *IntervalValue) DeepCopy() Value       { return v }
func (v *IntervalValue) Permute(*MVPerm) Value { return v }

func (v *IntervalValue) FingerPrint(fp uint64) uint64 {
	sz, err := v.Size()
	if err != nil {
		panic(err)
	}
	fp = FP64ExtendInt(fp, int32(SetEnumValueKind))
	fp = FP64ExtendInt(fp, int32(sz))
	for i := 0; i < sz; i++ {
		fp = FP64ExtendInt(fp, int32(IntValueKind))
		fp = FP64ExtendInt(fp, v.Low+int32(i))
	}
	return fp
}

func (v *IntervalValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, v.unsupported("attempted to apply EXCEPT construct to the interval value %s", v)
	}
	return ex.Value, nil
}

func (v *IntervalValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, v.unsupported("attempted to apply EXCEPT construct to the interval value %s", v)
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
	return NewSetEnumValue(values, true)
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
	sz, err := v.Size()
	if err != nil {
		panic(err)
	}
	values := make([]Value, sz)
	for i := 0; i < sz; i++ {
		values[i] = NewIntValue(v.Low + int32(i))
	}
	return &sliceValueEnumeration{values: values}
}

func (v *IntervalValue) String() string {
	if v.Low <= v.High {
		return NewIntValue(v.Low).String() + ".." + NewIntValue(v.High).String()
	}
	return "{}"
}

type sliceValueEnumeration struct {
	values []Value
	index  int
}

func (e *sliceValueEnumeration) Reset() {
	e.index = 0
}

func (e *sliceValueEnumeration) NextElement() Value {
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
	enum := enumerable.Elements()
	values := make([]Value, 0)
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		values = append(values, elem)
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	randomized := make([]Value, len(values))
	for i, idx := range randomSubsetIndices(len(values), len(values)) {
		randomized[i] = values[idx]
	}
	return &sliceValueEnumeration{values: randomized}, nil
}
