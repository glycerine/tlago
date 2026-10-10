package tlc

import "strings"

type SetCupValue struct {
	BaseValue
	Set1        Value
	Set2        Value
	CupSet      *SetEnumValue
	CupSetDummy bool
}

func NewSetCupValue(set1, set2 Value, cms ...CostModel) *SetCupValue {
	return &SetCupValue{BaseValue: newBaseValue(cms...), Set1: set1, Set2: set2}
}

func (v *SetCupValue) Kind() ValueKind    { return SetCupValueKind }
func (v *SetCupValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetCupValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetCupValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetCupValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	ok, err := v.Set1.Member(elem)
	if err != nil || ok {
		return ok, err
	}
	return v.Set2.Member(elem)
}

func (v *SetCupValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	f1, err := v.Set1.IsFinite()
	if err != nil || !f1 {
		return f1, err
	}
	return v.Set2.IsFinite()
}

func (v *SetCupValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetCupValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	return v.CupSet != nil && !v.CupSetDummy && v.CupSet.IsNormalized()
}

func (v *SetCupValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.CupSet != nil && !v.CupSetDummy {
		v.CupSet.Normalize()
	}
	return v
}

func (v *SetCupValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	v.Set1.DeepNormalize()
	v.Set2.DeepNormalize()
	if v.CupSet == nil {
		v.CupSetDummy = true
	} else if !v.CupSetDummy {
		v.CupSet.DeepNormalize()
	}
}

func (v *SetCupValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	return v.Set1.IsDefined() && v.Set2.IsDefined()
}

func (v *SetCupValue) DeepCopy() Value { return v }

func (v *SetCupValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetCupValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetCupValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SetCupValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SetCupValue) ToSetEnum() (*SetEnumValue, error) {
	if v.CupSet != nil && !v.CupSetDummy {
		return v.CupSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumeration(enum, false, v.CM)
}

func (v *SetCupValue) convertAndCache() (*SetEnumValue, error) {
	if v.CupSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.CupSet = set
		v.CupSetDummy = false
	} else if v.CupSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.CupSet = set
		v.CupSetDummy = false
	}
	return v.CupSet, nil
}

func (v *SetCupValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.CupSet != nil && !v.CupSetDummy {
		return v.CupSet.Elements()
	}
	enum1, ok1 := asEnumerable(v.Set1)
	enum2, ok2 := asEnumerable(v.Set2)
	if !ok1 || !ok2 {
		return newErrorEnumeration(v.runtimeFailure("Attempted to enumerate S \\cup T when S:\n" + ValuesPPR(v.Set1) + "\nand T:\n" + ValuesPPR(v.Set2) + "\nare not both enumerable"))
	}
	return &setCupEnumeration{enum1: enum1.Elements(), enum2: enum2.Elements(), cm: v.CM}
}

func (v *SetCupValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetCupValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if Globals.Expand {
		if expanded, ok := tryExpandedSetString(sb, offset, swallow, v.ToSetEnum); ok {
			return expanded
		}
	}
	sb = appendValueString(v.Set1, sb, offset, swallow)
	sb.WriteString(" \\cup ")
	return appendValueString(v.Set2, sb, offset, swallow)
}

type setCupEnumeration struct {
	cm    CostModel
	enum1 ValueEnumeration
	enum2 ValueEnumeration
}

func (e *setCupEnumeration) Reset() {
	e.enum1.Reset()
	e.enum2.Reset()
}

func (e *setCupEnumeration) NextElement() Value {
	e.cm.incValueSecondary()
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
	Set1        Value
	Set2        Value
	CapSet      *SetEnumValue
	CapSetDummy bool
}

func NewSetCapValue(set1, set2 Value) *SetCapValue {
	return &SetCapValue{Set1: set1, Set2: set2}
}

func (v *SetCapValue) Kind() ValueKind    { return SetCapValueKind }
func (v *SetCapValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetCapValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetCapValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetCapValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	ok, err := v.Set1.Member(elem)
	if err != nil || !ok {
		return ok, err
	}
	return v.Set2.Member(elem)
}

func (v *SetCapValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	f1, err := v.Set1.IsFinite()
	if err != nil || f1 {
		return f1, err
	}
	f2, err := v.Set2.IsFinite()
	if err != nil {
		return false, err
	}
	if !f2 {
		return false, v.runtimeFailure("Attempted to check if the set " + ValuesPPR(v) + "is finite.")
	}
	return true, nil
}

func (v *SetCapValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetCapValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	if v.CapSet == nil || v.CapSetDummy {
		return v.Set1.IsNormalized() && v.Set2.IsNormalized()
	}
	return v.CapSet.IsNormalized()
}

func (v *SetCapValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.CapSet == nil || v.CapSetDummy {
		v.Set1.Normalize()
		v.Set2.Normalize()
	} else {
		v.CapSet.Normalize()
	}
	return v
}

func (v *SetCapValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	v.Set1.DeepNormalize()
	v.Set2.DeepNormalize()
	if v.CapSet == nil {
		v.CapSetDummy = true
	} else if !v.CapSetDummy {
		v.CapSet.DeepNormalize()
	}
}

func (v *SetCapValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	return v.Set1.IsDefined() && v.Set2.IsDefined()
}

func (v *SetCapValue) DeepCopy() Value { return v }

func (v *SetCapValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetCapValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetCapValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SetCapValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SetCapValue) ToSetEnum() (*SetEnumValue, error) {
	if v.CapSet != nil && !v.CapSetDummy {
		return v.CapSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumerationWithNormalization(enum, v.IsNormalized, v.CM)
}

func (v *SetCapValue) convertAndCache() (*SetEnumValue, error) {
	if v.CapSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.CapSet = set
		v.CapSetDummy = false
	} else if v.CapSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.CapSet = set
		v.CapSetDummy = false
	}
	return v.CapSet, nil
}

func (v *SetCapValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.CapSet != nil && !v.CapSetDummy {
		return v.CapSet.Elements()
	}
	if enum1, ok := asEnumerable(v.Set1); ok {
		return &setFilterEnumeration{enum: enum1.Elements(), predicate: v.Set2, includeWhenMember: true, cm: v.CM}
	}
	if enum2, ok := asEnumerable(v.Set2); ok {
		return &setFilterEnumeration{enum: enum2.Elements(), predicate: v.Set1, includeWhenMember: true, cm: v.CM}
	}
	return newErrorEnumeration(v.runtimeFailure("Attempted to enumerate S \\cap T when neither S:\n" + ValuesPPR(v.Set1) + "\nnor T:\n" + ValuesPPR(v.Set2) + "\nis enumerable"))
}

func (v *SetCapValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetCapValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if Globals.Expand {
		if expanded, ok := tryExpandedSetString(sb, offset, swallow, v.ToSetEnum); ok {
			return expanded
		}
	}
	sb = appendValueString(v.Set1, sb, offset, swallow)
	sb.WriteString(" \\cap ")
	return appendValueString(v.Set2, sb, offset, swallow)
}

type SetDiffValue struct {
	BaseValue
	Set1         Value
	Set2         Value
	DiffSet      *SetEnumValue
	DiffSetDummy bool
}

func NewSetDiffValue(set1, set2 Value) *SetDiffValue {
	return &SetDiffValue{Set1: set1, Set2: set2}
}

func (v *SetDiffValue) Kind() ValueKind    { return SetDiffValueKind }
func (v *SetDiffValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetDiffValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetDiffValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetDiffValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
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

func (v *SetDiffValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
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
		return false, v.runtimeFailure("Attempted to check if the set " + ValuesPPR(v) + "is finite.")
	}
	return false, nil
}

func (v *SetDiffValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetDiffValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	if v.DiffSet == nil || v.DiffSetDummy {
		return v.Set1.IsNormalized()
	}
	return v.DiffSet.IsNormalized()
}

func (v *SetDiffValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.DiffSet == nil || v.DiffSetDummy {
		v.Set1.Normalize()
		v.Set2.Normalize()
	} else {
		v.DiffSet.Normalize()
	}
	return v
}

func (v *SetDiffValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	v.Set1.DeepNormalize()
	v.Set2.DeepNormalize()
	if v.DiffSet == nil {
		v.DiffSetDummy = true
	} else if !v.DiffSetDummy {
		v.DiffSet.DeepNormalize()
	}
}

func (v *SetDiffValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	return v.Set1.IsDefined() && v.Set2.IsDefined()
}

func (v *SetDiffValue) DeepCopy() Value { return v }

func (v *SetDiffValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetDiffValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetDiffValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SetDiffValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SetDiffValue) ToSetEnum() (*SetEnumValue, error) {
	if v.DiffSet != nil && !v.DiffSetDummy {
		return v.DiffSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumerationWithNormalization(enum, v.Set1.IsNormalized, v.CM)
}

func (v *SetDiffValue) convertAndCache() (*SetEnumValue, error) {
	if v.DiffSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.DiffSet = set
		v.DiffSetDummy = false
	} else if v.DiffSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.DiffSet = set
		v.DiffSetDummy = false
	}
	return v.DiffSet, nil
}

func (v *SetDiffValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.DiffSet != nil && !v.DiffSetDummy {
		return v.DiffSet.Elements()
	}
	enum1, ok := asEnumerable(v.Set1)
	if !ok {
		return newErrorEnumeration(v.runtimeFailure("Attempted to enumerate S \\ T when S:\n" + ValuesPPR(v.Set1) + "\nis not enumerable."))
	}
	return &setFilterEnumeration{enum: enum1.Elements(), predicate: v.Set2, includeWhenMember: false, cm: v.CM}
}

func (v *SetDiffValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetDiffValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if Globals.Expand {
		if expanded, ok := tryExpandedSetString(sb, offset, swallow, v.ToSetEnum); ok {
			return expanded
		}
	}
	sb = appendValueString(v.Set1, sb, offset, swallow)
	sb.WriteString(" \\ ")
	return appendValueString(v.Set2, sb, offset, swallow)
}

type UnionValue struct {
	BaseValue
	Set          Value
	RealSet      *SetEnumValue
	RealSetDummy bool
}

func NewUnionValue(set Value, cms ...CostModel) *UnionValue {
	return &UnionValue{BaseValue: newBaseValue(cms...), Set: set}
}

func Union(set Value) (Value, error) {
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
			result := NewSetEnumValueVec(resElems, false, set.GetCostModel())
			for i := 0; i < setEnum.Elems.Len(); i++ {
				inner := setEnum.Elems.At(i).(*SetEnumValue)
				for j := 0; j < inner.Elems.Len(); j++ {
					elem := inner.Elems.At(j)
					member, err := result.Member(elem)
					if err != nil {
						return nil, err
					}
					if !member {
						resElems.Add(elem)
					}
				}
			}
			return result, nil
		}
	}
	return NewUnionValue(set, set.GetCostModel()), nil
}

func (v *UnionValue) Kind() ValueKind    { return UnionValueKind }
func (v *UnionValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *UnionValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *UnionValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *UnionValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	enum, ok := asEnumerable(v.Set)
	if !ok {
		if elem == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure("Attempted to check if:\n " + ValuesPPR(elem) + "\nis an element of the non-enumerable set:\n " + ValuesPPR(v))
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

func (v *UnionValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	enum, ok := asEnumerable(v.Set)
	if !ok {
		return false, v.runtimeFailure("Attempted to check if the nonenumerable set:\n" + ValuesPPR(v) + "\nis a finite set.")
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

func (v *UnionValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *UnionValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	return v.RealSet != nil && !v.RealSetDummy && v.RealSet.IsNormalized()
}

func (v *UnionValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.RealSet != nil && !v.RealSetDummy {
		v.RealSet.Normalize()
	}
	return v
}

func (v *UnionValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	v.Set.DeepNormalize()
	if v.RealSet == nil {
		v.RealSetDummy = true
	} else if !v.RealSetDummy {
		v.RealSet.DeepNormalize()
	}
}

func (v *UnionValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	return v.Set.IsDefined()
}
func (v *UnionValue) DeepCopy() Value { return v }

func (v *UnionValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *UnionValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *UnionValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set:\n" + ValuesPPR(v))
	}
	return ex.Value, nil
}

func (v *UnionValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set:\n " + ValuesPPR(v) + ".")
	}
	return v, nil
}

func (v *UnionValue) ToSetEnum() (*SetEnumValue, error) {
	if v.RealSet != nil && !v.RealSetDummy {
		return v.RealSet, nil
	}
	enum := v.Elements()
	return setEnumFromEnumeration(enum, false, v.CM)
}

func (v *UnionValue) convertAndCache() (*SetEnumValue, error) {
	if v.RealSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.RealSet = set
		v.RealSetDummy = false
	} else if v.RealSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.RealSet = set
		v.RealSetDummy = false
	}
	return v.RealSet, nil
}

func (v *UnionValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.RealSet != nil && !v.RealSetDummy {
		return v.RealSet.Elements()
	}
	enum, ok := asEnumerable(v.Set)
	if !ok {
		return newErrorEnumeration(v.runtimeFailure("Attempted to enumerate the nonenumerable set:\n" + ValuesPPR(v.Set)))
	}
	return newUnionEnumeration(enum.Elements(), v)
}

func (v *UnionValue) String() string {
	return ValueToString(v, "", true)
}

func (v *UnionValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if Globals.Expand {
		set, err := v.ToSetEnum()
		if err != nil {
			panic(err)
		}
		return appendValueString(set, sb, offset, swallow)
	}
	sb.WriteString("UNION(")
	sb = appendValueString(v.Set, sb, offset, swallow)
	sb.WriteString(")")
	return sb
}

type setFilterEnumeration struct {
	cm                CostModel
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
		e.cm.incValueSecondary()
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
	val := e.elemSetEnum.NextElement()
	if val == nil {
		if err := e.elemSetEnum.Err(); err != nil {
			e.err = err
			return nil
		}
		e.advanceElementSet()
		if e.err != nil || e.elemSet == nil {
			return nil
		}
		val = e.NextElement()
		if e.err != nil {
			return nil
		}
	}
	e.owner.CM.incValueSecondary()
	return val
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
		e.err = e.owner.runtimeFailure("Attempted to enumerate UNION(s), but some element of s is nonenumerable.")
		return
	}
	e.elemSetEnum = enum.Elements()
	e.err = e.elemSetEnum.Err()
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

func setEnumFromEnumeration(enum ValueEnumeration, isNorm bool, cms ...CostModel) (*SetEnumValue, error) {
	return setEnumFromEnumerationWithNormalization(enum, func() bool { return isNorm }, cms...)
}

func setEnumFromEnumerationWithNormalization(enum ValueEnumeration, isNorm func() bool, cms ...CostModel) (*SetEnumValue, error) {
	vals := NewValueVec(0)
	for {
		elem := enum.NextElement()
		if elem == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			if len(cms) > 0 {
				cms[0].incValueSecondary(int64(vals.Len()))
			}
			return NewSetEnumValueVec(vals, isNorm(), cms...), nil
		}
		vals.Add(elem)
	}
}

// setExceptFailure preserves Assert.fail(reason, getSource()) for set values.
func setExceptFailure(v Value) error {
	message := "Attempted to apply EXCEPT to the set " + ValuesPPR(v) + "."
	if source := valueSource(v); source != nil {
		return NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
	}
	return NewTLCRuntimeExceptionMessage(message)
}

func takeExceptOnSet(v Value, ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, setExceptFailure(v)
	}
	return ex.Value, nil
}

func takeExceptsOnSet(v Value, exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, setExceptFailure(v)
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
