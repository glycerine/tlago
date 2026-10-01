package tlc

import "strings"

type SetCupValue struct {
	BaseValue
	Set1   Value
	Set2   Value
	CupSet *SetEnumValue
}

func NewSetCupValue(set1, set2 Value) *SetCupValue {
	return &SetCupValue{Set1: set1, Set2: set2}
}

func (v *SetCupValue) Kind() ValueKind    { return SetCupValueKind }
func (v *SetCupValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetCupValue) Compare(other Value) (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetCupValue) Equal(other Value) (bool, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetCupValue) Member(elem Value) (bool, error) {
	ok, err := v.Set1.Member(elem)
	if err != nil || ok {
		return ok, err
	}
	return v.Set2.Member(elem)
}

func (v *SetCupValue) IsFinite() (bool, error) {
	f1, err := v.Set1.IsFinite()
	if err != nil || !f1 {
		return f1, err
	}
	return v.Set2.IsFinite()
}

func (v *SetCupValue) Size() (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetCupValue) IsNormalized() bool {
	return v.CupSet != nil && v.CupSet.IsNormalized()
}

func (v *SetCupValue) Normalize() Value {
	if v.CupSet != nil {
		v.CupSet.Normalize()
	}
	return v
}

func (v *SetCupValue) DeepNormalize() {
	v.Set1.DeepNormalize()
	v.Set2.DeepNormalize()
	if v.CupSet != nil {
		v.CupSet.DeepNormalize()
	}
}

func (v *SetCupValue) IsDefined() bool {
	return v.Set1.IsDefined() && v.Set2.IsDefined()
}

func (v *SetCupValue) DeepCopy() Value { return v }

func (v *SetCupValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *SetCupValue) Permute(perm *MVPerm) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SetCupValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SetCupValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *SetCupValue) ToSetEnum() (*SetEnumValue, error) {
	if v.CupSet != nil {
		return v.CupSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumeration(enum, false)
}

func (v *SetCupValue) convertAndCache() (*SetEnumValue, error) {
	if v.CupSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.CupSet = set
	}
	return v.CupSet, nil
}

func (v *SetCupValue) Elements() ValueEnumeration {
	if v.CupSet != nil {
		return v.CupSet.Elements()
	}
	enum1, ok1 := asEnumerable(v.Set1)
	enum2, ok2 := asEnumerable(v.Set2)
	if !ok1 || !ok2 {
		return newErrorEnumeration(v.unsupported("Attempted to enumerate S \\cup T when S:\n%s\nand T:\n%s\nare not both enumerable", v.Set1, v.Set2))
	}
	return &setCupEnumeration{enum1: enum1.Elements(), enum2: enum2.Elements()}
}

func (v *SetCupValue) String() string {
	if Globals.Expand {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
		}
	}
	return v.Set1.String() + " \\cup " + v.Set2.String()
}

type setCupEnumeration struct {
	enum1 ValueEnumeration
	enum2 ValueEnumeration
}

func (e *setCupEnumeration) Reset() {
	e.enum1.Reset()
	e.enum2.Reset()
}

func (e *setCupEnumeration) NextElement() Value {
	elem := e.enum1.NextElement()
	if elem != nil {
		return elem
	}
	return e.enum2.NextElement()
}

func (e *setCupEnumeration) Err() error {
	if err := e.enum1.Err(); err != nil {
		return err
	}
	return e.enum2.Err()
}

type SetCapValue struct {
	BaseValue
	Set1   Value
	Set2   Value
	CapSet *SetEnumValue
}

func NewSetCapValue(set1, set2 Value) *SetCapValue {
	return &SetCapValue{Set1: set1, Set2: set2}
}

func (v *SetCapValue) Kind() ValueKind    { return SetCapValueKind }
func (v *SetCapValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetCapValue) Compare(other Value) (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetCapValue) Equal(other Value) (bool, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetCapValue) Member(elem Value) (bool, error) {
	ok, err := v.Set1.Member(elem)
	if err != nil || !ok {
		return ok, err
	}
	return v.Set2.Member(elem)
}

func (v *SetCapValue) IsFinite() (bool, error) {
	f1, err := v.Set1.IsFinite()
	if err != nil {
		return false, err
	}
	f2, err := v.Set2.IsFinite()
	if err != nil {
		return false, err
	}
	if !f1 && !f2 {
		return false, v.unsupported("Attempted to check if the set %sis finite.", v)
	}
	return true, nil
}

func (v *SetCapValue) Size() (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetCapValue) IsNormalized() bool {
	if v.CapSet == nil {
		return v.Set1.IsNormalized() && v.Set2.IsNormalized()
	}
	return v.CapSet.IsNormalized()
}

func (v *SetCapValue) Normalize() Value {
	if v.CapSet == nil {
		v.Set1.Normalize()
		v.Set2.Normalize()
	} else {
		v.CapSet.Normalize()
	}
	return v
}

func (v *SetCapValue) DeepNormalize() {
	v.Set1.DeepNormalize()
	v.Set2.DeepNormalize()
	if v.CapSet != nil {
		v.CapSet.DeepNormalize()
	}
}

func (v *SetCapValue) IsDefined() bool {
	return v.Set1.IsDefined() && v.Set2.IsDefined()
}

func (v *SetCapValue) DeepCopy() Value { return v }

func (v *SetCapValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *SetCapValue) Permute(perm *MVPerm) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SetCapValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SetCapValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *SetCapValue) ToSetEnum() (*SetEnumValue, error) {
	if v.CapSet != nil {
		return v.CapSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumeration(enum, v.IsNormalized())
}

func (v *SetCapValue) convertAndCache() (*SetEnumValue, error) {
	if v.CapSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.CapSet = set
	}
	return v.CapSet, nil
}

func (v *SetCapValue) Elements() ValueEnumeration {
	if v.CapSet != nil {
		return v.CapSet.Elements()
	}
	if enum1, ok := asEnumerable(v.Set1); ok {
		return &setFilterEnumeration{enum: enum1.Elements(), predicate: v.Set2, includeWhenMember: true}
	}
	if enum2, ok := asEnumerable(v.Set2); ok {
		return &setFilterEnumeration{enum: enum2.Elements(), predicate: v.Set1, includeWhenMember: true}
	}
	return newErrorEnumeration(v.unsupported("Attempted to enumerate S \\cap T when neither S:\n%s\nnor T:\n%s\nis enumerable", v.Set1, v.Set2))
}

func (v *SetCapValue) String() string {
	if Globals.Expand {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
		}
	}
	return v.Set1.String() + " \\cap " + v.Set2.String()
}

type SetDiffValue struct {
	BaseValue
	Set1    Value
	Set2    Value
	DiffSet *SetEnumValue
}

func NewSetDiffValue(set1, set2 Value) *SetDiffValue {
	return &SetDiffValue{Set1: set1, Set2: set2}
}

func (v *SetDiffValue) Kind() ValueKind    { return SetDiffValueKind }
func (v *SetDiffValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetDiffValue) Compare(other Value) (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetDiffValue) Equal(other Value) (bool, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetDiffValue) Member(elem Value) (bool, error) {
	ok, err := v.Set1.Member(elem)
	if err != nil || !ok {
		return ok, err
	}
	inSecond, err := v.Set2.Member(elem)
	if err != nil {
		return false, err
	}
	return !inSecond, nil
}

func (v *SetDiffValue) IsFinite() (bool, error) {
	f1, err := v.Set1.IsFinite()
	if err != nil {
		return false, err
	}
	if f1 {
		return true, nil
	}
	f2, err := v.Set2.IsFinite()
	if err != nil {
		return false, err
	}
	if !f2 {
		return false, v.unsupported("Attempted to check if the set %sis finite.", v)
	}
	return false, nil
}

func (v *SetDiffValue) Size() (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetDiffValue) IsNormalized() bool {
	if v.DiffSet == nil {
		return v.Set1.IsNormalized()
	}
	return v.DiffSet.IsNormalized()
}

func (v *SetDiffValue) Normalize() Value {
	if v.DiffSet == nil {
		v.Set1.Normalize()
		v.Set2.Normalize()
	} else {
		v.DiffSet.Normalize()
	}
	return v
}

func (v *SetDiffValue) DeepNormalize() {
	v.Set1.DeepNormalize()
	v.Set2.DeepNormalize()
	if v.DiffSet != nil {
		v.DiffSet.DeepNormalize()
	}
}

func (v *SetDiffValue) IsDefined() bool {
	return v.Set1.IsDefined() && v.Set2.IsDefined()
}

func (v *SetDiffValue) DeepCopy() Value { return v }

func (v *SetDiffValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *SetDiffValue) Permute(perm *MVPerm) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SetDiffValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SetDiffValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *SetDiffValue) ToSetEnum() (*SetEnumValue, error) {
	if v.DiffSet != nil {
		return v.DiffSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumeration(enum, v.Set1.IsNormalized())
}

func (v *SetDiffValue) convertAndCache() (*SetEnumValue, error) {
	if v.DiffSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.DiffSet = set
	}
	return v.DiffSet, nil
}

func (v *SetDiffValue) Elements() ValueEnumeration {
	if v.DiffSet != nil {
		return v.DiffSet.Elements()
	}
	enum1, ok := asEnumerable(v.Set1)
	if !ok {
		return newErrorEnumeration(v.unsupported("Attempted to enumerate S \\ T when S:\n%s\nis not enumerable.", v.Set1))
	}
	return &setFilterEnumeration{enum: enum1.Elements(), predicate: v.Set2, includeWhenMember: false}
}

func (v *SetDiffValue) String() string {
	if Globals.Expand {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
		}
	}
	return v.Set1.String() + " \\ " + v.Set2.String()
}

type UnionValue struct {
	BaseValue
	Set     Value
	RealSet *SetEnumValue
}

func NewUnionValue(set Value) *UnionValue {
	return &UnionValue{Set: set}
}

func Union(set Value) Value {
	setEnum, ok := set.(*SetEnumValue)
	if ok {
		canCombine := true
		for i := 0; i < setEnum.Elems.Len(); i++ {
			if _, ok := setEnum.Elems.At(i).(*SetEnumValue); !ok {
				canCombine = false
				break
			}
		}
		if canCombine {
			resElems := NewValueVec(0)
			result := NewSetEnumValueVec(resElems, false)
			for i := 0; i < setEnum.Elems.Len(); i++ {
				inner := setEnum.Elems.At(i).(*SetEnumValue)
				for j := 0; j < inner.Elems.Len(); j++ {
					elem := inner.Elems.At(j)
					member, err := result.Member(elem)
					if err == nil && !member {
						resElems.Add(elem)
					}
				}
			}
			return result
		}
	}
	return NewUnionValue(set)
}

func (v *UnionValue) Kind() ValueKind    { return UnionValueKind }
func (v *UnionValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *UnionValue) Compare(other Value) (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *UnionValue) Equal(other Value) (bool, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *UnionValue) Member(elem Value) (bool, error) {
	enum, ok := asEnumerable(v.Set)
	if !ok {
		return false, v.unsupported("Attempted to check if:\n %s\nis an element of the non-enumerable set\n%s", elem, v)
	}
	e := enum.Elements()
	for {
		val := e.NextElement()
		if val == nil {
			return false, e.Err()
		}
		ok, err := val.Member(elem)
		if err != nil || ok {
			return ok, err
		}
	}
}

func (v *UnionValue) IsFinite() (bool, error) {
	enum, ok := asEnumerable(v.Set)
	if !ok {
		return false, v.unsupported("Attempted to check if the nonenumerable set:\n%s\nis a finite set.", v)
	}
	e := enum.Elements()
	for {
		val := e.NextElement()
		if val == nil {
			return true, e.Err()
		}
		finite, err := val.IsFinite()
		if err != nil || !finite {
			return finite, err
		}
	}
}

func (v *UnionValue) Size() (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *UnionValue) IsNormalized() bool {
	return v.RealSet != nil && v.RealSet.IsNormalized()
}

func (v *UnionValue) Normalize() Value {
	if v.RealSet != nil {
		v.RealSet.Normalize()
	}
	return v
}

func (v *UnionValue) DeepNormalize() {
	v.Set.DeepNormalize()
	if v.RealSet != nil {
		v.RealSet.DeepNormalize()
	}
}

func (v *UnionValue) IsDefined() bool { return v.Set.IsDefined() }
func (v *UnionValue) DeepCopy() Value { return v }

func (v *UnionValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *UnionValue) Permute(perm *MVPerm) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *UnionValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, newTLCError(ECGeneral, "Attempted to apply EXCEPT to the set:\n%s", v)
	}
	return ex.Value, nil
}

func (v *UnionValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, newTLCError(ECGeneral, "Attempted to apply EXCEPT to the set:\n %s.", v)
	}
	return v, nil
}

func (v *UnionValue) ToSetEnum() (*SetEnumValue, error) {
	if v.RealSet != nil {
		return v.RealSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumeration(enum, false)
}

func (v *UnionValue) convertAndCache() (*SetEnumValue, error) {
	if v.RealSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.RealSet = set
	}
	return v.RealSet, nil
}

func (v *UnionValue) Elements() ValueEnumeration {
	if v.RealSet != nil {
		return v.RealSet.Elements()
	}
	enum, ok := asEnumerable(v.Set)
	if !ok {
		return newErrorEnumeration(v.unsupported("Attempted to enumerate the nonenumerable set:\n%s", v.Set))
	}
	return newUnionEnumeration(enum.Elements(), v)
}

func (v *UnionValue) String() string {
	if Globals.Expand {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
		}
	}
	return "UNION(" + v.Set.String() + ")"
}

type setFilterEnumeration struct {
	enum              ValueEnumeration
	predicate         Value
	includeWhenMember bool
	err               error
}

func (e *setFilterEnumeration) Reset() {
	e.err = nil
	e.enum.Reset()
}

func (e *setFilterEnumeration) NextElement() Value {
	if e.err != nil {
		return nil
	}
	for {
		elem := e.enum.NextElement()
		if elem == nil {
			return nil
		}
		member, err := e.predicate.Member(elem)
		if err != nil {
			e.err = err
			return nil
		}
		if member == e.includeWhenMember {
			return elem
		}
	}
}

func (e *setFilterEnumeration) Err() error {
	if e.err != nil {
		return e.err
	}
	return e.enum.Err()
}

type unionEnumeration struct {
	enum        ValueEnumeration
	elemSet     Value
	elemSetEnum ValueEnumeration
	owner       *UnionValue
	err         error
}

func newUnionEnumeration(enum ValueEnumeration, owner *UnionValue) *unionEnumeration {
	out := &unionEnumeration{enum: enum, owner: owner}
	out.advanceElementSet()
	return out
}

func (e *unionEnumeration) Reset() {
	e.err = nil
	e.enum.Reset()
	e.advanceElementSet()
}

func (e *unionEnumeration) NextElement() Value {
	if e.err != nil || e.elemSet == nil {
		return nil
	}
	for {
		val := e.elemSetEnum.NextElement()
		if val != nil {
			return val
		}
		if err := e.elemSetEnum.Err(); err != nil {
			e.err = err
			return nil
		}
		e.advanceElementSet()
		if e.err != nil || e.elemSet == nil {
			return nil
		}
	}
}

func (e *unionEnumeration) Err() error {
	if e.err != nil {
		return e.err
	}
	return e.enum.Err()
}

func (e *unionEnumeration) advanceElementSet() {
	e.elemSet = e.enum.NextElement()
	if e.elemSet == nil {
		return
	}
	enum, ok := asEnumerable(e.elemSet)
	if !ok {
		e.err = e.owner.unsupported("Attempted to enumerate UNION(s), but some element of s is nonenumerable.")
		return
	}
	e.elemSetEnum = enum.Elements()
}

type errorEnumeration struct {
	err error
}

func newErrorEnumeration(err error) *errorEnumeration {
	return &errorEnumeration{err: err}
}

func (e *errorEnumeration) Reset()             {}
func (e *errorEnumeration) NextElement() Value { return nil }
func (e *errorEnumeration) Err() error         { return e.err }

func asEnumerable(value Value) (Enumerable, bool) {
	enum, ok := value.(Enumerable)
	return enum, ok
}

func setEnumFromEnumeration(enum ValueEnumeration, isNorm bool) (*SetEnumValue, error) {
	vals := NewValueVec(0)
	for {
		elem := enum.NextElement()
		if elem == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			return NewSetEnumValueVec(vals, isNorm), nil
		}
		vals.Add(elem)
	}
}

func takeExceptOnSet(v Value, ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, newTLCError(ECGeneral, "Attempted to apply EXCEPT to the set %s.", v)
	}
	return ex.Value, nil
}

func takeExceptsOnSet(v Value, exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, newTLCError(ECGeneral, "Attempted to apply EXCEPT to the set %s.", v)
	}
	return v, nil
}

func joinValueStrings(values []Value, sep string) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = value.String()
	}
	return strings.Join(parts, sep)
}
