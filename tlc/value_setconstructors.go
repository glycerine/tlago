package tlc

import (
	"fmt"
	"strings"
)

type SetOfTuplesValue struct {
	BaseValue
	Sets          []Value
	TupleSet      *SetEnumValue
	TupleSetDummy bool
}

func NewSetOfTuplesValue(sets []Value, cms ...CostModel) *SetOfTuplesValue {
	return &SetOfTuplesValue{BaseValue: newBaseValue(cms...), Sets: sets}
}

func (v *SetOfTuplesValue) Kind() ValueKind    { return SetOfTuplesValueKind }
func (v *SetOfTuplesValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetOfTuplesValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfTuplesValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if o, ok := other.(*SetOfTuplesValue); ok {
		empty, err := IsEmptyValue(v)
		if err != nil || empty {
			if err != nil {
				return false, err
			}
			return IsEmptyValue(o)
		}
		otherEmpty, err := IsEmptyValue(o)
		if err != nil || otherEmpty {
			return false, err
		}
		if len(v.Sets) != len(o.Sets) {
			return false, nil
		}
		for i := range v.Sets {
			eq, err := v.Sets[i].Equal(o.Sets[i])
			if err != nil || !eq {
				return eq, err
			}
		}
		return true, nil
	}
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetOfTuplesValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	tv := asTupleValue(elem)
	if tv == nil {
		if fcn := asFcnRcdValue(elem); fcn != nil {
			if fcn.Intv != nil {
				return false, nil
			}
			for _, d := range fcn.Domain {
				if _, ok := d.(*IntValue); !ok {
					return false, v.runtimeFailure("Attempted to check if non-tuple\n" + ValuesPPR(elem) + "\nis in the set of tuples:\n" + ValuesPPR(v))
				}
			}
			return false, nil
		}
		if mv, ok := elem.(*ModelValue); ok {
			return mv.modelValueMember(v)
		}
		return false, v.runtimeFailure("Attempted to check if non-tuple\n" + ValuesPPR(elem) + "\nis in the set of tuples:\n" + ValuesPPR(v))
	}
	if len(tv.Elems) != len(v.Sets) {
		return false, nil
	}
	for i := range v.Sets {
		ok, err := v.Sets[i].Member(tv.Elems[i])
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (v *SetOfTuplesValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	allFinite := true
	for _, set := range v.Sets {
		finite, err := set.IsFinite()
		if err != nil {
			return false, err
		}
		if finite {
			empty, err := IsEmptyValue(set)
			if err != nil {
				return false, err
			}
			if empty {
				return true, nil
			}
		} else {
			allFinite = false
		}
	}
	return allFinite, nil
}

func (v *SetOfTuplesValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	empty, err := IsEmptyValue(v)
	if err != nil || empty {
		return 0, err
	}
	return checkedProductSize(v.Sets, func() error {
		return v.runtimeFailure("Overflow when computing the number of elements in " + ValuesPPR(v))
	})
}

func (v *SetOfTuplesValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	if v.TupleSet != nil && !v.TupleSetDummy {
		return v.TupleSet.IsNormalized()
	}
	for _, set := range v.Sets {
		if !set.IsNormalized() {
			return false
		}
	}
	return true
}

func (v *SetOfTuplesValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.TupleSet != nil && !v.TupleSetDummy {
		v.TupleSet.Normalize()
	} else {
		for _, set := range v.Sets {
			set.Normalize()
		}
	}
	return v
}

func (v *SetOfTuplesValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	for _, set := range v.Sets {
		set.DeepNormalize()
	}
	if v.TupleSet == nil {
		v.TupleSetDummy = true
	} else if !v.TupleSetDummy {
		v.TupleSet.DeepNormalize()
	}
}

func (v *SetOfTuplesValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	for _, set := range v.Sets {
		if !set.IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetOfTuplesValue) DeepCopy() Value { return v }

func (v *SetOfTuplesValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetOfTuplesValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetOfTuplesValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the set of tuples:\n" + ValuesPPR(v))
	}
	return ex.Value, nil
}

func (v *SetOfTuplesValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the set of tuples:\n" + ValuesPPR(v))
	}
	return v, nil
}

func (v *SetOfTuplesValue) ToSetEnum() (*SetEnumValue, error) {
	if v.TupleSet != nil && !v.TupleSetDummy {
		return v.TupleSet, nil
	}
	return setEnumFromEnumerationWithNormalization(v.Elements(), v.IsNormalized, v.CM)
}

func (v *SetOfTuplesValue) convertAndCache() (*SetEnumValue, error) {
	if v.TupleSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.TupleSet = set
		v.TupleSetDummy = false
	} else if v.TupleSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.TupleSet = set
		v.TupleSetDummy = false
	}
	return v.TupleSet, nil
}

func (v *SetOfTuplesValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.TupleSet != nil && !v.TupleSetDummy {
		return v.TupleSet.Elements()
	}
	empty, err := IsEmptyValue(v)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if empty {
		return EmptySet.Elements()
	}
	return newProductEnumeration(v.Sets, func(elems []Value) Value {
		return NewTupleValue(elems, v.CM)
	}, func(i int, set Value) error {
		return v.runtimeFailure(fmt.Sprintf("Attempted to enumerate a set of the form s1 \\X s2 ... \\X sn,\nbut can't enumerate s%d:\n%s", i, ValuesPPR(set)))
	}, v.CM)
}

func (v *SetOfTuplesValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetOfTuplesValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if set := tryExpandedSet(swallow, func() (*SetEnumValue, error) {
		if shouldExpandProduct(v.Sets) {
			return v.ToSetEnum()
		}
		return nil, nil
	}); set != nil {
		return appendValueString(set, sb, offset, swallow)
	}
	if len(v.Sets) > 0 {
		sb.WriteString("(")
	}
	for i, set := range v.Sets {
		if i > 0 {
			sb.WriteString(" \\X ")
		}
		appendValueString(set, sb, offset, swallow)
	}
	if len(v.Sets) > 0 {
		sb.WriteString(")")
	}
	return sb
}

type SetOfRcdsValue struct {
	BaseValue
	Names       []*UniqueString
	Values      []Value
	RcdSet      *SetEnumValue
	RcdSetDummy bool
}

func NewSetOfRcdsValue(names []*UniqueString, values []Value, isNorm bool, cms ...CostModel) (*SetOfRcdsValue, error) {
	out := &SetOfRcdsValue{BaseValue: newBaseValue(cms...), Names: names, Values: values}
	if !isNorm {
		if err := out.sortByNames(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (v *SetOfRcdsValue) Kind() ValueKind    { return SetOfRcdsValueKind }
func (v *SetOfRcdsValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetOfRcdsValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfRcdsValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if o, ok := other.(*SetOfRcdsValue); ok {
		empty, err := IsEmptyValue(v)
		if err != nil || empty {
			if err != nil {
				return false, err
			}
			return IsEmptyValue(o)
		}
		otherEmpty, err := IsEmptyValue(o)
		if err != nil || otherEmpty {
			return false, err
		}
		if len(v.Names) != len(o.Names) {
			return false, nil
		}
		for i := range v.Names {
			if !v.Names[i].Equal(o.Names[i]) {
				return false, nil
			}
			eq, err := v.Values[i].Equal(o.Values[i])
			if err != nil || !eq {
				return eq, err
			}
		}
		return true, nil
	}
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetOfRcdsValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	rcd := asRecordValue(elem)
	if rcd == nil {
		if mv, ok := elem.(*ModelValue); ok {
			return mv.modelValueMember(v)
		}
		return false, v.runtimeFailure("Attempted to check if non-record\n" + elem.String() + "\nis in the set of records:\n" + ValuesPPR(v))
	}
	if err := rcd.normalizeRecord(); err != nil {
		return false, err
	}
	if len(v.Names) != len(rcd.Names) {
		return false, nil
	}
	for i := range v.Names {
		if !v.Names[i].Equal(rcd.Names[i]) {
			return false, nil
		}
		ok, err := v.Values[i].Member(rcd.Values[i])
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (v *SetOfRcdsValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	allFinite := true
	for _, value := range v.Values {
		finite, err := value.IsFinite()
		if err != nil {
			return false, err
		}
		if finite {
			empty, err := IsEmptyValue(value)
			if err != nil {
				return false, err
			}
			if empty {
				return true, nil
			}
		} else {
			allFinite = false
		}
	}
	return allFinite, nil
}

func (v *SetOfRcdsValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	empty, err := IsEmptyValue(v)
	if err != nil || empty {
		return 0, err
	}
	return checkedProductSize(v.Values, func() error {
		return NewTLCRuntimeException(ECTLCModuleOverflow, "the number of elements in:\n"+ValuesPPR(v))
	})
}

func (v *SetOfRcdsValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	if v.RcdSet != nil && !v.RcdSetDummy {
		return v.RcdSet.IsNormalized()
	}
	for _, value := range v.Values {
		if !value.IsNormalized() {
			return false
		}
	}
	return true
}

func (v *SetOfRcdsValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.RcdSet != nil && !v.RcdSetDummy {
		v.RcdSet.Normalize()
	} else {
		for _, value := range v.Values {
			value.Normalize()
		}
	}
	return v
}

func (v *SetOfRcdsValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	for _, value := range v.Values {
		value.DeepNormalize()
	}
	if v.RcdSet == nil {
		v.RcdSetDummy = true
	} else if !v.RcdSetDummy {
		v.RcdSet.DeepNormalize()
	}
}

func (v *SetOfRcdsValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	for _, value := range v.Values {
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetOfRcdsValue) DeepCopy() Value { return v }

func (v *SetOfRcdsValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetOfRcdsValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetOfRcdsValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set of records:\n" + ValuesPPR(v))
	}
	return ex.Value, nil
}

func (v *SetOfRcdsValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set of records:\n" + ValuesPPR(v))
	}
	return v, nil
}

func (v *SetOfRcdsValue) ToSetEnum() (*SetEnumValue, error) {
	if v.RcdSet != nil && !v.RcdSetDummy {
		return v.RcdSet, nil
	}
	return setEnumFromEnumerationWithNormalization(v.Elements(), v.IsNormalized, v.CM)
}

func (v *SetOfRcdsValue) convertAndCache() (*SetEnumValue, error) {
	if v.RcdSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.RcdSet = set
		v.RcdSetDummy = false
	} else if v.RcdSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.RcdSet = set
		v.RcdSetDummy = false
	}
	return v.RcdSet, nil
}

func (v *SetOfRcdsValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.RcdSet != nil && !v.RcdSetDummy {
		return v.RcdSet.Elements()
	}
	empty, err := IsEmptyValue(v)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if empty {
		return EmptySet.Elements()
	}
	return newProductEnumeration(v.Values, func(fields []Value) Value {
		return NewRecordValue(v.Names, fields, true, v.CM)
	}, func(i int, value Value) error {
		return v.runtimeFailure("Attempted to enumerate a set of the form [l1 : v1, ..., ln : vn],\nbut can't enumerate the value of the `" + v.Names[i].String() + "' field:\n" + ValuesPPR(value))
	}, v.CM)
}

func (v *SetOfRcdsValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetOfRcdsValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if set := tryExpandedSet(swallow, func() (*SetEnumValue, error) {
		if shouldExpandProduct(v.Values) {
			return v.ToSetEnum()
		}
		return nil, nil
	}); set != nil {
		return appendValueString(set, sb, offset, swallow)
	}
	sb.WriteString("[")
	for i, name := range v.Names {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name.String() + ": ")
		appendValueString(v.Values[i], sb, offset, swallow)
	}
	sb.WriteString("]")
	return sb
}

func (v *SetOfRcdsValue) sortByNames() error {
	for i := 1; i < len(v.Names); i++ {
		cmp := v.Names[0].Compare(v.Names[i])
		if cmp == 0 {
			return v.runtimeFailure("Field name " + v.Names[0].String() + " occurs multiple times in set of records.")
		}
		if cmp > 0 {
			v.Names[0], v.Names[i] = v.Names[i], v.Names[0]
			v.Values[0], v.Values[i] = v.Values[i], v.Values[0]
		}
	}
	for i := 2; i < len(v.Names); i++ {
		j := i
		st := v.Names[i]
		val := v.Values[i]
		cmp := -1
		for j > 0 {
			cmp = st.Compare(v.Names[j-1])
			if cmp >= 0 {
				break
			}
			v.Names[j] = v.Names[j-1]
			v.Values[j] = v.Values[j-1]
			j--
		}
		if cmp == 0 {
			return v.runtimeFailure("Field name " + v.Names[i].String() + " occurs multiple times in set of records.")
		}
		v.Names[j] = st
		v.Values[j] = val
	}
	return nil
}

type SetOfFcnsValue struct {
	BaseValue
	Domain      Value
	Range       Value
	FcnSet      *SetEnumValue
	FcnSetDummy bool
}

func NewSetOfFcnsValue(domain, rangeValue Value, cms ...CostModel) *SetOfFcnsValue {
	return &SetOfFcnsValue{BaseValue: newBaseValue(cms...), Domain: domain, Range: rangeValue}
}

func (v *SetOfFcnsValue) Kind() ValueKind    { return SetOfFcnsValueKind }
func (v *SetOfFcnsValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetOfFcnsValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfFcnsValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if o, ok := other.(*SetOfFcnsValue); ok {
		unit1, err := IsEmptyValue(v.Domain)
		if err != nil {
			return false, err
		}
		unit2, err := IsEmptyValue(o.Domain)
		if err != nil {
			return false, err
		}
		if unit1 || unit2 {
			return unit1 == unit2, nil
		}
		empty1, err := IsEmptyValue(v.Range)
		if err != nil {
			return false, err
		}
		empty2, err := IsEmptyValue(o.Range)
		if err != nil {
			return false, err
		}
		if empty1 || empty2 {
			return empty1 == empty2, nil
		}
		domainEqual, err := v.Domain.Equal(o.Domain)
		if err != nil || !domainEqual {
			return domainEqual, err
		}
		return v.Range.Equal(o.Range)
	}
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetOfFcnsValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	fcn := asFcnRcdValue(elem)
	if fcn == nil {
		if mv, ok := elem.(*ModelValue); ok {
			return mv.modelValueMember(v)
		}
		return false, v.unsupported("Attempted to check if \n%s\nwhich is not a TLC function value, is in the set of functions:\n%s", elem, ValuesPPR(v))
	}
	if fcn.Intv == nil {
		if err := fcn.normalizeFcn(); err != nil {
			return false, err
		}
		fdom := NewSetEnumValue(fcn.Domain, true)
		domainEqual, err := v.Domain.Equal(fdom)
		if err != nil || !domainEqual {
			return domainEqual, err
		}
		for _, value := range fcn.Values {
			ok, err := v.Range.Member(value)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	}
	domainEqual, err := fcn.Intv.Equal(v.Domain)
	if err != nil || !domainEqual {
		return domainEqual, err
	}
	for _, value := range fcn.Values {
		ok, err := v.Range.Member(value)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (v *SetOfFcnsValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	finiteDomain, err := v.Domain.IsFinite()
	if err != nil {
		return false, err
	}
	if finiteDomain {
		empty, err := IsEmptyValue(v.Domain)
		if err != nil {
			return false, err
		}
		if empty {
			return true, nil
		}
	}
	finiteRange, err := v.Range.IsFinite()
	if err != nil {
		return false, err
	}
	if finiteRange {
		empty, err := IsEmptyValue(v.Range)
		if err != nil {
			return false, err
		}
		if empty {
			return true, nil
		}
	}
	if finiteDomain && finiteRange {
		return true, nil
	}
	if finiteRange {
		size, err := v.Range.Size()
		if err != nil {
			return false, err
		}
		return size == 1, nil
	}
	return false, nil
}

func (v *SetOfFcnsValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	domainEmpty, err := IsEmptyValue(v.Domain)
	if err != nil {
		return 0, err
	}
	if domainEmpty {
		return 1, nil
	}
	rangeEmpty, err := IsEmptyValue(v.Range)
	if err != nil {
		return 0, err
	}
	if rangeEmpty {
		return 0, nil
	}
	rangeSize, err := v.Range.Size()
	if err != nil {
		return 0, err
	}
	if rangeSize == 1 {
		return 1, nil
	}
	domainSize, err := v.Domain.Size()
	if err != nil {
		return 0, err
	}
	size := int64(1)
	for i := 0; i < domainSize; i++ {
		size *= int64(rangeSize)
		if size < -2147483648 || size > 2147483647 {
			return 0, v.unsupported("Overflow when computing the number of elements in:\n%s", ValuesPPR(v))
		}
	}
	return int(size), nil
}

func (v *SetOfFcnsValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	if v.FcnSet != nil && !v.FcnSetDummy {
		return v.FcnSet.IsNormalized()
	}
	return v.Domain.IsNormalized() && v.Range.IsNormalized()
}

func (v *SetOfFcnsValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.FcnSet != nil && !v.FcnSetDummy {
		v.FcnSet.Normalize()
	} else {
		v.Domain.Normalize()
		v.Range.Normalize()
	}
	return v
}

func (v *SetOfFcnsValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	v.Domain.DeepNormalize()
	v.Range.DeepNormalize()
	if v.FcnSet == nil {
		v.FcnSetDummy = true
	} else if !v.FcnSetDummy {
		v.FcnSet.DeepNormalize()
	}
}

func (v *SetOfFcnsValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	return v.Domain.IsDefined() && v.Range.IsDefined()
}

func (v *SetOfFcnsValue) DeepCopy() Value { return v }

func (v *SetOfFcnsValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetOfFcnsValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetOfFcnsValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex.Index < len(ex.Path) {
		return nil, newTLCError(ECGeneral, "Attempted to apply EXCEPT to the set of functions:\n%s", ValuesPPR(v))
	}
	return ex.Value, nil
}

func (v *SetOfFcnsValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if len(exs) != 0 {
		return nil, newTLCError(ECGeneral, "Attempted to apply EXCEPT to the set of functions:\n%s", ValuesPPR(v))
	}
	return v, nil
}

func (v *SetOfFcnsValue) ToSetEnum() (*SetEnumValue, error) {
	if v.FcnSet != nil && !v.FcnSetDummy {
		return v.FcnSet, nil
	}
	return setEnumFromEnumerationWithNormalization(v.Elements(), v.IsNormalized, v.CM)
}

func (v *SetOfFcnsValue) convertAndCache() (*SetEnumValue, error) {
	if v.FcnSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.FcnSet = set
		v.FcnSetDummy = false
	} else if v.FcnSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.FcnSet = set
		v.FcnSetDummy = false
	}
	return v.FcnSet, nil
}

func (v *SetOfFcnsValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.FcnSet != nil && !v.FcnSetDummy {
		return v.FcnSet.Elements()
	}
	empty, err := IsEmptyValue(v)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if empty {
		return EmptySet.Elements()
	}
	if intv, ok := v.Domain.(*IntervalValue); ok {
		return v.intervalDomainElements(intv)
	}
	domSet, err := tryToSetEnumValue(v.Domain)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if domSet == nil {
		return newErrorEnumeration(v.unsupported("Attempted to enumerate a set of the form [D -> R],but the domain D:\n%s\ncannot be enumerated.", ValuesPPR(v.Domain)))
	}
	if _, err := domSet.normalizeSet(); err != nil {
		return newErrorEnumeration(err)
	}
	dom := domSet.Elems.ToArray()
	return v.domainElements(dom)
}

func (v *SetOfFcnsValue) intervalDomainElements(intv *IntervalValue) ValueEnumeration {
	size, err := intv.Size()
	if err != nil {
		return newErrorEnumeration(err)
	}
	rangeEnum, ok := asEnumerable(v.Range)
	if size > 0 && !ok {
		return newErrorEnumeration(v.unsupported("Attempted to enumerate a set of the form [D -> R],but the range R:\n%s\ncannot be enumerated.", ValuesPPR(v.Range)))
	}
	sets := make([]Value, size)
	for i := range sets {
		sets[i] = v.Range
	}
	if size == 0 {
		return &singleValueEnumeration{value: NewFcnRcdIntervalValue(intv, []Value{}, v.CM), cm: v.CM, secondary: 1}
	}
	_ = rangeEnum
	return newProductEnumeration(sets, func(elems []Value) Value {
		return NewFcnRcdIntervalValue(intv, elems, v.CM)
	}, func(i int, value Value) error {
		return v.unsupported("Attempted to enumerate a set of the form [D -> R],but the range R:\n%s\ncannot be enumerated.", ValuesPPR(v.Range))
	}, v.CM)
}

func (v *SetOfFcnsValue) domainElements(dom []Value) ValueEnumeration {
	rangeEnum, ok := asEnumerable(v.Range)
	if len(dom) > 0 && !ok {
		return newErrorEnumeration(v.unsupported("Attempted to enumerate a set of the form [D -> R],but the range R:\n%s\ncannot be enumerated.", ValuesPPR(v.Range)))
	}
	if len(dom) == 0 {
		return &singleValueEnumeration{value: NewFcnRcdValue(dom, []Value{}, true, v.CM), cm: v.CM, secondary: 1}
	}
	_ = rangeEnum
	sets := make([]Value, len(dom))
	for i := range sets {
		sets[i] = v.Range
	}
	return newProductEnumeration(sets, func(elems []Value) Value {
		return NewFcnRcdValue(dom, elems, true, v.CM)
	}, func(i int, value Value) error {
		return v.unsupported("Attempted to enumerate a set of the form [D -> R],but the range R:\n%s\ncannot be enumerated.", ValuesPPR(v.Range))
	}, v.CM)
}

func (v *SetOfFcnsValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetOfFcnsValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if set := tryExpandedSet(swallow, func() (*SetEnumValue, error) {
		if shouldExpandFcnSet(v.Domain, v.Range) {
			return v.ToSetEnum()
		}
		return nil, nil
	}); set != nil {
		return appendValueString(set, sb, offset, swallow)
	}
	sb.WriteString("[")
	appendValueString(v.Domain, sb, offset, swallow)
	sb.WriteString(" -> ")
	appendValueString(v.Range, sb, offset, swallow)
	sb.WriteString("]")
	return sb
}

type SubsetValue struct {
	BaseValue
	Set       Value
	PSet      *SetEnumValue
	PSetDummy bool
}

func NewSubsetValue(set Value, cms ...CostModel) *SubsetValue {
	return &SubsetValue{BaseValue: newBaseValue(cms...), Set: set}
}

func (v *SubsetValue) Kind() ValueKind    { return SubsetValueKind }
func (v *SubsetValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SubsetValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if kSubset, ok := other.(*KSubsetValue); ok {
		cmp, err := kSubset.Compare(v)
		if err != nil {
			return 0, err
		}
		switch {
		case cmp < 0:
			return 1, nil
		case cmp > 0:
			return -1, nil
		default:
			return 0, nil
		}
	}
	if o, ok := other.(*SubsetValue); ok {
		return v.Set.Compare(o.Set)
	}
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SubsetValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if kSubset, ok := other.(*KSubsetValue); ok {
		return kSubset.Equal(v)
	}
	if o, ok := other.(*SubsetValue); ok {
		return v.Set.Equal(o.Set)
	}
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SubsetValue) Member(elem Value) (bool, error) {
	return subsetValueMember(v, v.Set, elem)
}

// This is SubsetValue.member's catch boundary, also used by k-subsets after
// their uncaught cardinality checks, preserving the actual receiver's source.
func subsetValueMember(owner Value, set Value, elem Value) (resultBool bool, err error) {
	defer catchValueFailure(owner, &err)
	enum, ok := asEnumerable(elem)
	if !ok {
		if elem == nil {
			panic(NewNullPointerException())
		}
		message := "Attempted to check if the non-enumerable value\n" + ValuesPPR(elem) + "\nis element of\n" + ValuesPPR(owner)
		if source := valueSource(owner); source != nil {
			return false, NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
		}
		return false, NewTLCRuntimeExceptionMessage(message)
	}
	e := enum.Elements()
	for {
		value := e.NextElement()
		if value == nil {
			if err := e.Err(); err != nil {
				return false, err
			}
			return true, nil
		}
		ok, err := set.Member(value)
		if err != nil || !ok {
			return ok, err
		}
	}
}

func (v *SubsetValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return v.Set.IsFinite()
}

func (v *SubsetValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	size, err := v.Set.Size()
	if err != nil {
		return 0, err
	}
	if size >= 31 {
		return 0, NewTLCRuntimeException(ECTLCModuleOverflow, "the number of elements in:\n"+ValuesPPR(v))
	}
	return 1 << size, nil
}

func (v *SubsetValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	return v.PSet != nil && !v.PSetDummy && v.PSet.IsNormalized()
}

func (v *SubsetValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	if v.PSet != nil && !v.PSetDummy {
		v.PSet.Normalize()
	} else {
		v.Set.Normalize()
	}
	return v
}

func (v *SubsetValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	v.Set.DeepNormalize()
	if v.PSet == nil {
		v.PSetDummy = true
	} else if !v.PSetDummy {
		v.PSet.DeepNormalize()
	}
}

func (v *SubsetValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	return v.Set.IsDefined()
}

func (v *SubsetValue) DeepCopy() Value { return v }

func (v *SubsetValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SubsetValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SubsetValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SubsetValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SubsetValue) ToSetEnum() (*SetEnumValue, error) {
	if v.PSet != nil && !v.PSetDummy {
		return v.PSet, nil
	}
	size, err := v.Size()
	if err != nil {
		return nil, err
	}
	values := NewValueVec(size)
	enum := v.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		values.Add(elem)
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	v.CM.incValueSecondary(int64(values.Len()))
	return NewSetEnumValueVec(values, true, v.CM), nil
}

func (v *SubsetValue) convertAndCache() (*SetEnumValue, error) {
	if v.PSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.PSet = set
		v.PSetDummy = false
	} else if v.PSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.PSet = set
		v.PSetDummy = false
	}
	return v.PSet, nil
}

func (v *SubsetValue) Elements() (enumeration ValueEnumeration) {
	defer catchValueFailure(v, nil)
	defer wrapInitialEnumerationFailure(v, &enumeration)
	if v.PSet != nil && !v.PSetDummy {
		return v.PSet.Elements()
	}
	return v.ElementsNormalized()
}

// ElementsNormalized ports the direct SubsetValue.elementsNormalized entry
// point, independently of the cached powerset used by Elements.
func (v *SubsetValue) ElementsNormalized() ValueEnumeration {
	size, err := v.Set.Size()
	if err != nil {
		return newErrorEnumeration(err)
	}
	if size == 0 {
		return &emptySubsetEnumeration{owner: v}
	}
	set, err := tryToSetEnumValue(v.Set)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if set == nil {
		message := "Attempted to compute the value of an expression of form\nSUBSET S, but S is a non-enumerable value:\n" + ValuesPPR(v.Set)
		if source := v.GetSource(); source != nil {
			return newErrorEnumeration(NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext))
		}
		failure := newTLCError(ECGeneral, "%s", message)
		failure.Runtime = true
		return newErrorEnumeration(failure)
	}
	if _, err := set.normalizeSet(); err != nil {
		return newErrorEnumeration(err)
	}
	return &subsetEnumeration{elems: set.Elems, cm: v.CM}
}

func (v *SubsetValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SubsetValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if Globals.Expand && tryValueSizeBelow(v.Set, 7, swallow) {
		set, err := v.ToSetEnum()
		if err != nil {
			panic(err)
		}
		return appendValueString(set, sb, offset, swallow)
	}
	sb.WriteString("SUBSET ")
	if _, ok := v.Set.(*IntervalValue); ok {
		sb.WriteString("(")
		sb = appendValueString(v.Set, sb, offset, swallow)
		sb.WriteString(")")
	} else {
		sb = appendValueString(v.Set, sb, offset, swallow)
	}
	return sb
}

type singleValueEnumeration struct {
	value     Value
	done      bool
	cm        CostModel
	secondary int64
}

func (e *singleValueEnumeration) Reset() {
	e.done = false
}

func (e *singleValueEnumeration) NextElement() Value {
	if e.done {
		return nil
	}
	e.done = true
	e.cm.incValueSecondary(e.secondary)
	return e.value
}

func (e *singleValueEnumeration) Err() error { return nil }

type subsetEnumeration struct {
	cm    CostModel
	elems *ValueVec
	k     int
	cur   *kSubsetEnumeration
	done  bool
	err   error
}

func (e *subsetEnumeration) Reset() {
	e.done = false
	e.err = nil
	e.k = 0
	e.cur = newKSubsetEnumeration(e.elems, e.k, e.cm)
}

func (e *subsetEnumeration) NextElement() Value {
	if e.cur == nil {
		e.Reset()
	}
	for !e.done && e.err == nil {
		if value := e.cur.NextElement(); value != nil {
			return value
		}
		if err := e.cur.Err(); err != nil {
			e.err = err
			return nil
		}
		e.k++
		if e.k > e.elems.Len() {
			e.done = true
			return nil
		}
		e.cur = newKSubsetEnumeration(e.elems, e.k, e.cm)
	}
	return nil
}

func (e *subsetEnumeration) Err() error { return e.err }
