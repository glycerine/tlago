package tlc

import "strings"

type SetOfTuplesValue struct {
	BaseValue
	Sets     []Value
	TupleSet *SetEnumValue
}

func NewSetOfTuplesValue(sets []Value) *SetOfTuplesValue {
	out := make([]Value, len(sets))
	copy(out, sets)
	return &SetOfTuplesValue{Sets: out}
}

func (v *SetOfTuplesValue) Kind() ValueKind    { return SetOfTuplesValueKind }
func (v *SetOfTuplesValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetOfTuplesValue) Compare(other Value) (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfTuplesValue) Equal(other Value) (bool, error) {
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

func (v *SetOfTuplesValue) Member(elem Value) (bool, error) {
	tv, ok := elem.(*TupleValue)
	if !ok {
		if fcn, ok := elem.(*FcnRcdValue); ok {
			if fcn.Intv != nil {
				return false, nil
			}
			for _, d := range fcn.Domain {
				if _, ok := d.(*IntValue); !ok {
					return false, v.unsupported("attempted to check if non-tuple\n%s\nis in the set of tuples:\n%s", elem, v)
				}
			}
			return false, nil
		}
		if mv, ok := elem.(*ModelValue); ok {
			return mv.modelValueMember(v)
		}
		return false, v.unsupported("attempted to check if non-tuple\n%s\nis in the set of tuples:\n%s", elem, v)
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

func (v *SetOfTuplesValue) IsFinite() (bool, error) {
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

func (v *SetOfTuplesValue) Size() (int, error) {
	empty, err := IsEmptyValue(v)
	if err != nil || empty {
		return 0, err
	}
	return checkedProductSize(v.Sets, v.String())
}

func (v *SetOfTuplesValue) IsNormalized() bool {
	if v.TupleSet != nil {
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
	if v.TupleSet != nil {
		v.TupleSet.Normalize()
	} else {
		for _, set := range v.Sets {
			set.Normalize()
		}
	}
	return v
}

func (v *SetOfTuplesValue) DeepNormalize() {
	for _, set := range v.Sets {
		set.DeepNormalize()
	}
	if v.TupleSet != nil {
		v.TupleSet.DeepNormalize()
	}
}

func (v *SetOfTuplesValue) IsDefined() bool {
	for _, set := range v.Sets {
		if !set.IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetOfTuplesValue) DeepCopy() Value { return v }

func (v *SetOfTuplesValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *SetOfTuplesValue) Permute(perm ModelValuePermutation) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SetOfTuplesValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SetOfTuplesValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *SetOfTuplesValue) ToSetEnum() (*SetEnumValue, error) {
	if v.TupleSet != nil {
		return v.TupleSet, nil
	}
	return setEnumFromEnumeration(v.Elements(), v.IsNormalized())
}

func (v *SetOfTuplesValue) convertAndCache() (*SetEnumValue, error) {
	if v.TupleSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.TupleSet = set
	}
	return v.TupleSet, nil
}

func (v *SetOfTuplesValue) Elements() ValueEnumeration {
	if v.TupleSet != nil {
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
		return NewTupleValue(elems)
	}, func(i int, set Value) error {
		return v.unsupported("attempted to enumerate a set of the form s1 \\X s2 ... \\X sn, but cannot enumerate s%d:\n%s", i, set)
	})
}

func (v *SetOfTuplesValue) String() string {
	if shouldExpandProduct(v.Sets) {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
		}
	}
	parts := make([]string, len(v.Sets))
	for i, set := range v.Sets {
		parts[i] = set.String()
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, " \\X ") + ")"
}

type SetOfRcdsValue struct {
	BaseValue
	Names  []*UniqueString
	Values []Value
	RcdSet *SetEnumValue
}

func NewSetOfRcdsValue(names []*UniqueString, values []Value, isNorm bool) (*SetOfRcdsValue, error) {
	outNames := make([]*UniqueString, len(names))
	outValues := make([]Value, len(values))
	copy(outNames, names)
	copy(outValues, values)
	out := &SetOfRcdsValue{Names: outNames, Values: outValues}
	if !isNorm {
		if err := out.sortByNames(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (v *SetOfRcdsValue) Kind() ValueKind    { return SetOfRcdsValueKind }
func (v *SetOfRcdsValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetOfRcdsValue) Compare(other Value) (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfRcdsValue) Equal(other Value) (bool, error) {
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

func (v *SetOfRcdsValue) Member(elem Value) (bool, error) {
	rcd := asRecordValue(elem)
	if rcd == nil {
		if mv, ok := elem.(*ModelValue); ok {
			return mv.modelValueMember(v)
		}
		return false, v.unsupported("attempted to check if non-record\n%s\nis in the set of records:\n%s", elem, v)
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

func (v *SetOfRcdsValue) IsFinite() (bool, error) {
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

func (v *SetOfRcdsValue) Size() (int, error) {
	empty, err := IsEmptyValue(v)
	if err != nil || empty {
		return 0, err
	}
	return checkedProductSize(v.Values, v.String())
}

func (v *SetOfRcdsValue) IsNormalized() bool {
	if v.RcdSet != nil {
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
	if v.RcdSet != nil {
		v.RcdSet.Normalize()
	} else {
		for _, value := range v.Values {
			value.Normalize()
		}
	}
	return v
}

func (v *SetOfRcdsValue) DeepNormalize() {
	for _, value := range v.Values {
		value.DeepNormalize()
	}
	if v.RcdSet != nil {
		v.RcdSet.DeepNormalize()
	}
}

func (v *SetOfRcdsValue) IsDefined() bool {
	for _, value := range v.Values {
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetOfRcdsValue) DeepCopy() Value { return v }

func (v *SetOfRcdsValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *SetOfRcdsValue) Permute(perm ModelValuePermutation) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SetOfRcdsValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SetOfRcdsValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *SetOfRcdsValue) ToSetEnum() (*SetEnumValue, error) {
	if v.RcdSet != nil {
		return v.RcdSet, nil
	}
	return setEnumFromEnumeration(v.Elements(), v.IsNormalized())
}

func (v *SetOfRcdsValue) convertAndCache() (*SetEnumValue, error) {
	if v.RcdSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.RcdSet = set
	}
	return v.RcdSet, nil
}

func (v *SetOfRcdsValue) Elements() ValueEnumeration {
	if v.RcdSet != nil {
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
		return &RecordValue{Names: v.Names, Values: fields, IsNorm: true}
	}, func(i int, value Value) error {
		return v.unsupported("attempted to enumerate a set of the form [l1 : v1, ..., ln : vn], but cannot enumerate the value of the `%s' field:\n%s", v.Names[i], value)
	})
}

func (v *SetOfRcdsValue) String() string {
	if shouldExpandProduct(v.Values) {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
		}
	}
	var b strings.Builder
	b.WriteString("[")
	for i := range v.Names {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.Names[i].String())
		b.WriteString(": ")
		b.WriteString(v.Values[i].String())
	}
	b.WriteString("]")
	return b.String()
}

func (v *SetOfRcdsValue) sortByNames() error {
	for i := 1; i < len(v.Names); i++ {
		cmp := v.Names[0].Compare(v.Names[i])
		if cmp == 0 {
			return v.unsupported("field name %s occurs multiple times in set of records", v.Names[0])
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
			return v.unsupported("field name %s occurs multiple times in set of records", v.Names[i])
		}
		v.Names[j] = st
		v.Values[j] = val
	}
	return nil
}

type SetOfFcnsValue struct {
	BaseValue
	Domain Value
	Range  Value
	FcnSet *SetEnumValue
}

func NewSetOfFcnsValue(domain, rangeValue Value) *SetOfFcnsValue {
	return &SetOfFcnsValue{Domain: domain, Range: rangeValue}
}

func (v *SetOfFcnsValue) Kind() ValueKind    { return SetOfFcnsValueKind }
func (v *SetOfFcnsValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetOfFcnsValue) Compare(other Value) (int, error) {
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfFcnsValue) Equal(other Value) (bool, error) {
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

func (v *SetOfFcnsValue) Member(elem Value) (bool, error) {
	fcn := asFcnRcdValue(elem)
	if fcn == nil {
		if mv, ok := elem.(*ModelValue); ok {
			return mv.modelValueMember(v)
		}
		return false, v.unsupported("attempted to check if \n%s\nwhich is not a TLC function value, is in the set of functions:\n%s", elem, v)
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

func (v *SetOfFcnsValue) IsFinite() (bool, error) {
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

func (v *SetOfFcnsValue) Size() (int, error) {
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
			return 0, v.unsupported("overflow when computing the number of elements in:\n%s", v)
		}
	}
	return int(size), nil
}

func (v *SetOfFcnsValue) IsNormalized() bool {
	if v.FcnSet != nil {
		return v.FcnSet.IsNormalized()
	}
	return v.Domain.IsNormalized() && v.Range.IsNormalized()
}

func (v *SetOfFcnsValue) Normalize() Value {
	if v.FcnSet != nil {
		v.FcnSet.Normalize()
	} else {
		v.Domain.Normalize()
		v.Range.Normalize()
	}
	return v
}

func (v *SetOfFcnsValue) DeepNormalize() {
	v.Domain.DeepNormalize()
	v.Range.DeepNormalize()
	if v.FcnSet != nil {
		v.FcnSet.DeepNormalize()
	}
}

func (v *SetOfFcnsValue) IsDefined() bool {
	return v.Domain.IsDefined() && v.Range.IsDefined()
}

func (v *SetOfFcnsValue) DeepCopy() Value { return v }

func (v *SetOfFcnsValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *SetOfFcnsValue) Permute(perm ModelValuePermutation) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SetOfFcnsValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SetOfFcnsValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *SetOfFcnsValue) ToSetEnum() (*SetEnumValue, error) {
	if v.FcnSet != nil {
		return v.FcnSet, nil
	}
	return setEnumFromEnumeration(v.Elements(), v.IsNormalized())
}

func (v *SetOfFcnsValue) convertAndCache() (*SetEnumValue, error) {
	if v.FcnSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.FcnSet = set
	}
	return v.FcnSet, nil
}

func (v *SetOfFcnsValue) Elements() ValueEnumeration {
	if v.FcnSet != nil {
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
	domSet, err := toSetEnumValue(v.Domain)
	if err != nil {
		return newErrorEnumeration(v.unsupported("attempted to enumerate a set of the form [D -> R], but the domain D:\n%s\ncannot be enumerated", v.Domain))
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
		return newErrorEnumeration(v.unsupported("attempted to enumerate a set of the form [D -> R], but the range R:\n%s\ncannot be enumerated", v.Range))
	}
	sets := make([]Value, size)
	for i := range sets {
		sets[i] = v.Range
	}
	if size == 0 {
		return &singleValueEnumeration{value: NewFcnRcdIntervalValue(intv, []Value{})}
	}
	_ = rangeEnum
	return newProductEnumeration(sets, func(elems []Value) Value {
		return NewFcnRcdIntervalValue(intv, elems)
	}, func(i int, value Value) error {
		return v.unsupported("attempted to enumerate a set of the form [D -> R], but the range R:\n%s\ncannot be enumerated", v.Range)
	})
}

func (v *SetOfFcnsValue) domainElements(dom []Value) ValueEnumeration {
	rangeEnum, ok := asEnumerable(v.Range)
	if len(dom) > 0 && !ok {
		return newErrorEnumeration(v.unsupported("attempted to enumerate a set of the form [D -> R], but the range R:\n%s\ncannot be enumerated", v.Range))
	}
	if len(dom) == 0 {
		return &singleValueEnumeration{value: NewFcnRcdValue(dom, []Value{}, true)}
	}
	_ = rangeEnum
	sets := make([]Value, len(dom))
	for i := range sets {
		sets[i] = v.Range
	}
	return newProductEnumeration(sets, func(elems []Value) Value {
		return NewFcnRcdValue(dom, elems, true)
	}, func(i int, value Value) error {
		return v.unsupported("attempted to enumerate a set of the form [D -> R], but the range R:\n%s\ncannot be enumerated", v.Range)
	})
}

func (v *SetOfFcnsValue) String() string {
	if Globals.Expand {
		if size, err := v.Size(); err == nil && size < Globals.EnumBound {
			if set, err := v.ToSetEnum(); err == nil {
				return set.String()
			}
		}
	}
	return "[" + v.Domain.String() + " -> " + v.Range.String() + "]"
}

type SubsetValue struct {
	BaseValue
	Set  Value
	PSet *SetEnumValue
}

func NewSubsetValue(set Value) *SubsetValue {
	return &SubsetValue{Set: set}
}

func (v *SubsetValue) Kind() ValueKind    { return SubsetValueKind }
func (v *SubsetValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SubsetValue) Compare(other Value) (int, error) {
	if o, ok := other.(*SubsetValue); ok {
		return v.Set.Compare(o.Set)
	}
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SubsetValue) Equal(other Value) (bool, error) {
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
	enum, ok := asEnumerable(elem)
	if !ok {
		return false, v.unsupported("attempted to check if the non-enumerable value\n%s\nis element of\n%s", elem, v)
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
		ok, err := v.Set.Member(value)
		if err != nil || !ok {
			return ok, err
		}
	}
}

func (v *SubsetValue) IsFinite() (bool, error) {
	return v.Set.IsFinite()
}

func (v *SubsetValue) Size() (int, error) {
	size, err := v.Set.Size()
	if err != nil {
		return 0, err
	}
	if size >= 31 {
		return 0, v.unsupported("overflow when computing the number of elements in:\n%s", v)
	}
	return 1 << size, nil
}

func (v *SubsetValue) IsNormalized() bool {
	return v.PSet != nil && v.PSet.IsNormalized()
}

func (v *SubsetValue) Normalize() Value {
	if v.PSet == nil {
		v.Set.Normalize()
	} else {
		v.PSet.Normalize()
	}
	return v
}

func (v *SubsetValue) DeepNormalize() {
	v.Set.DeepNormalize()
	if v.PSet != nil {
		v.PSet.DeepNormalize()
	}
}

func (v *SubsetValue) IsDefined() bool {
	return v.Set.IsDefined()
}

func (v *SubsetValue) DeepCopy() Value { return v }

func (v *SubsetValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *SubsetValue) Permute(perm ModelValuePermutation) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SubsetValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SubsetValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *SubsetValue) ToSetEnum() (*SetEnumValue, error) {
	if v.PSet != nil {
		return v.PSet, nil
	}
	return setEnumFromEnumeration(v.Elements(), true)
}

func (v *SubsetValue) convertAndCache() (*SetEnumValue, error) {
	if v.PSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.PSet = set
	}
	return v.PSet, nil
}

func (v *SubsetValue) Elements() ValueEnumeration {
	if v.PSet != nil {
		return v.PSet.Elements()
	}
	set, err := toSetEnumValue(v.Set)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if _, err := set.normalizeSet(); err != nil {
		return newErrorEnumeration(err)
	}
	return &subsetEnumeration{elems: set.Elems}
}

func (v *SubsetValue) String() string {
	if Globals.Expand {
		if size, err := v.Set.Size(); err == nil && size < 7 {
			if set, err := v.ToSetEnum(); err == nil {
				return set.String()
			}
		}
	}
	return "SUBSET " + v.Set.String()
}

type singleValueEnumeration struct {
	value Value
	done  bool
}

func (e *singleValueEnumeration) Reset() {
	e.done = false
}

func (e *singleValueEnumeration) NextElement() Value {
	if e.done {
		return nil
	}
	e.done = true
	return e.value
}

func (e *singleValueEnumeration) Err() error { return nil }

type subsetEnumeration struct {
	elems      *ValueVec
	descriptor []bool
	done       bool
}

func (e *subsetEnumeration) Reset() {
	e.descriptor = make([]bool, e.elems.Len())
	e.done = false
}

func (e *subsetEnumeration) NextElement() Value {
	if e.descriptor == nil {
		e.Reset()
	}
	if e.done {
		return nil
	}
	vals := NewValueVec(0)
	size := e.elems.Len()
	if size == 0 {
		e.done = true
	} else {
		for i := 0; i < size; i++ {
			if e.descriptor[i] {
				vals.Add(e.elems.At(i))
			}
		}
		for i := 0; i < size; i++ {
			if e.descriptor[i] {
				e.descriptor[i] = false
				if i >= size-1 {
					e.done = true
					break
				}
			} else {
				e.descriptor[i] = true
				break
			}
		}
	}
	return NewSetEnumValueVec(vals, true)
}

func (e *subsetEnumeration) Err() error { return nil }
