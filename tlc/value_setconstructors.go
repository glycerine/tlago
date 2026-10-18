package tlc

import (
	"fmt"
	"strings"
)

type SetOfTuplesValue struct {
	BaseValue
	Sets     []Value
	TupleSet *SetEnumValue

	setsCaptured bool
	setsOwner    []Value
}

func (v *SetOfTuplesValue) IsSubsetEq(other Value) (*BoolValue, error) {
	return defaultEnumerableSubsetEq(v, v, other)
}

func NewSetOfTuplesValue(sets []Value, cms ...CostModel) *SetOfTuplesValue {
	if sets == nil {
		panic(NewNullPointerException())
	}
	if len(sets) == 0 {
		panic(NewAssertionError())
	}
	return &SetOfTuplesValue{BaseValue: newBaseValue(cms...), Sets: sets, setsCaptured: true, setsOwner: sets}
}

func (v *SetOfTuplesValue) ownedSets() []Value {
	if v.setsCaptured {
		return v.setsOwner
	}
	return v.Sets
}

func (v *SetOfTuplesValue) captureSets(sets []Value) {
	v.Sets = sets
	v.setsCaptured = true
	v.setsOwner = sets
}

// NewSingletonSetOfTuplesValue represents SetOfTuplesValue(Value).
func NewSingletonSetOfTuplesValue(set Value) *SetOfTuplesValue {
	return NewSetOfTuplesValue([]Value{set})
}

// NewPairSetOfTuplesValue represents SetOfTuplesValue(Value, Value).
func NewPairSetOfTuplesValue(first, second Value) *SetOfTuplesValue {
	return NewSetOfTuplesValue([]Value{first, second})
}

func (v *SetOfTuplesValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return SetOfTuplesValueKind
}
func (v *SetOfTuplesValue) KindString() string { return ValueKindString(v) }
func (v *SetOfTuplesValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *SetOfTuplesValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfTuplesValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
		if len(v.ownedSets()) != len(o.ownedSets()) {
			return false, nil
		}
		for i := range v.ownedSets() {
			eq, err := v.ownedSets()[i].Equal(o.ownedSets()[i])
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
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	tv := asTupleValue(elem)
	if tv == nil {
		if fcn := asFcnRcdValue(elem); fcn != nil {
			if fcn.ownedIntv() != nil {
				return false, nil
			}
			for _, d := range fcn.ownedDomain() {
				if _, ok := d.(*IntValue); !ok {
					return false, v.runtimeFailure("Attempted to check if non-tuple\n" + ValuesPPR(elem) + "\nis in the set of tuples:\n" + ValuesPPR(v))
				}
			}
			return false, nil
		}
		if mv, ok := elem.(*ModelValue); ok {
			return mv.ModelValueMember(v)
		}
		return false, v.runtimeFailure("Attempted to check if non-tuple\n" + ValuesPPR(elem) + "\nis in the set of tuples:\n" + ValuesPPR(v))
	}
	if tv.GetElems() == nil || v.ownedSets() == nil {
		panic(NewNullPointerException())
	}
	if len(tv.GetElems()) != len(v.ownedSets()) {
		return false, nil
	}
	for i := range v.ownedSets() {
		if v.ownedSets()[i] == nil {
			panic(NewNullPointerException())
		}
		ok, err := v.ownedSets()[i].Member(tv.GetElems()[i])
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (v *SetOfTuplesValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	allFinite := true
	if v.ownedSets() == nil {
		panic(NewNullPointerException())
	}
	for _, set := range v.ownedSets() {
		if isNil(set) {
			panic(NewNullPointerException())
		}
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
	return checkedProductSize(v.ownedSets(), func() error {
		return v.runtimeFailure("Overflow when computing the number of elements in " + ValuesPPR(v))
	})
}

func (v *SetOfTuplesValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.TupleSet != nil && v.TupleSet != DummyEnum {
		return v.TupleSet.IsNormalized()
	}
	if v.ownedSets() == nil {
		panic(NewNullPointerException())
	}
	for _, set := range v.ownedSets() {
		if set == nil {
			panic(NewNullPointerException())
		}
		if !set.IsNormalized() {
			return false
		}
	}
	return true
}

func (v *SetOfTuplesValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.TupleSet != nil && v.TupleSet != DummyEnum {
		v.TupleSet.Normalize()
	} else {
		if v.ownedSets() == nil {
			panic(NewNullPointerException())
		}
		for _, set := range v.ownedSets() {
			if set == nil {
				panic(NewNullPointerException())
			}
			set.Normalize()
		}
	}
	return v
}

func (v *SetOfTuplesValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedSets() == nil {
		panic(NewNullPointerException())
	}
	for _, set := range v.ownedSets() {
		if set == nil {
			panic(NewNullPointerException())
		}
		set.DeepNormalize()
	}
	if v.TupleSet == nil {
		v.TupleSet = DummyEnum
	} else if v.TupleSet != DummyEnum {
		v.TupleSet.DeepNormalize()
	}
}

func (v *SetOfTuplesValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedSets() == nil {
		panic(NewNullPointerException())
	}
	for _, set := range v.ownedSets() {
		if set == nil {
			panic(NewNullPointerException())
		}
		if !set.IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetOfTuplesValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *SetOfTuplesValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetOfTuplesValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetOfTuplesValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the set of tuples:\n" + ValuesPPR(v))
	}
	return ex.Value, nil
}

func (v *SetOfTuplesValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the set of tuples:\n" + ValuesPPR(v))
	}
	return v, nil
}

func (v *SetOfTuplesValue) ToSetEnum() (*SetEnumValue, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.TupleSet != nil && v.TupleSet != DummyEnum {
		return v.TupleSet, nil
	}
	return setEnumFromEnumerationWithNormalization(v.Elements(), v.IsNormalized, v.CM)
}

func (v *SetOfTuplesValue) convertAndCache() (*SetEnumValue, error) {
	if v.TupleSet == DummyEnum {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.TupleSet = set
	} else if v.TupleSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.TupleSet = set
	}
	return v.TupleSet, nil
}

func (v *SetOfTuplesValue) Elements() (enumeration ValueEnumeration) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	defer raiseInitialEnumerationFailure(&enumeration)
	if v.TupleSet != nil && v.TupleSet != DummyEnum {
		return v.TupleSet.Elements()
	}
	empty, err := IsEmptyValue(v)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if empty {
		return EmptySet.Elements()
	}
	return newProductEnumeration(v.ownedSets(), func(elems []Value) Value {
		return NewTupleValue(elems, v.CM)
	}, func(i int, set Value) error {
		return v.runtimeFailure(fmt.Sprintf("Attempted to enumerate a set of the form s1 \\X s2 ... \\X sn,\nbut can't enumerate s%d:\n%s", i, ValuesPPR(set)))
	}, v.CM)
}

func (v *SetOfTuplesValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *SetOfTuplesValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if set := tryExpandedSet(swallow, func() (*SetEnumValue, error) {
		if shouldExpandProduct(v.ownedSets()) {
			return v.ToSetEnum()
		}
		return nil, nil
	}); set != nil {
		return appendValueString(set, sb, offset, swallow)
	}
	if v.ownedSets() == nil {
		panic(NewNullPointerException())
	}
	if len(v.ownedSets()) > 0 {
		sb.WriteString("(")
	}
	for i, set := range v.ownedSets() {
		if i > 0 {
			sb.WriteString(" \\X ")
		}
		appendValueString(set, sb, offset, swallow)
	}
	if len(v.ownedSets()) > 0 {
		sb.WriteString(")")
	}
	return sb
}

type SetOfRcdsValue struct {
	BaseValue
	Names  []*UniqueString
	Values []Value
	RcdSet *SetEnumValue

	arraysCaptured bool
	namesOwner     []*UniqueString
	valuesOwner    []Value
}

func (v *SetOfRcdsValue) IsSubsetEq(other Value) (*BoolValue, error) {
	return defaultEnumerableSubsetEq(v, v, other)
}

func NewSetOfRcdsValue(names []*UniqueString, values []Value, isNorm bool, cms ...CostModel) (*SetOfRcdsValue, error) {
	if names == nil || values == nil {
		panic(NewNullPointerException())
	}
	if len(names) != len(values) || len(values) == 0 {
		panic(NewAssertionError())
	}
	out := &SetOfRcdsValue{
		BaseValue: newBaseValue(cms...), Names: names, Values: values,
		arraysCaptured: true, namesOwner: names, valuesOwner: values,
	}
	if !isNorm {
		if err := out.sortByNames(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (v *SetOfRcdsValue) ownedNames() []*UniqueString {
	if v.arraysCaptured {
		return v.namesOwner
	}
	return v.Names
}

func (v *SetOfRcdsValue) ownedValues() []Value {
	if v.arraysCaptured {
		return v.valuesOwner
	}
	return v.Values
}

func (v *SetOfRcdsValue) captureArrays(names []*UniqueString, values []Value) {
	v.Names, v.Values = names, values
	v.arraysCaptured, v.namesOwner, v.valuesOwner = true, names, values
}

func (v *SetOfRcdsValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return SetOfRcdsValueKind
}
func (v *SetOfRcdsValue) KindString() string { return ValueKindString(v) }
func (v *SetOfRcdsValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *SetOfRcdsValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfRcdsValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
		if len(v.ownedNames()) != len(o.ownedNames()) {
			return false, nil
		}
		for i := range v.ownedNames() {
			if !v.ownedNames()[i].Equal(o.ownedNames()[i]) {
				return false, nil
			}
			eq, err := v.ownedValues()[i].Equal(o.ownedValues()[i])
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
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	rcd := asRecordValue(elem)
	if rcd == nil {
		if mv, ok := elem.(*ModelValue); ok {
			return mv.ModelValueMember(v)
		}
		return false, v.runtimeFailure("Attempted to check if non-record\n" + elem.String() + "\nis in the set of records:\n" + ValuesPPR(v))
	}
	if err := rcd.normalizeRecord(); err != nil {
		return false, err
	}
	if v.ownedNames() == nil || rcd.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	if len(v.ownedNames()) != len(rcd.ownedNames()) {
		return false, nil
	}
	for i := range v.ownedNames() {
		if v.ownedNames()[i] == nil || rcd.ownedNames()[i] == nil {
			panic(NewNullPointerException())
		}
		if !v.ownedNames()[i].Equal(rcd.ownedNames()[i]) {
			return false, nil
		}
		field := v.fieldValue(i)
		// Java evaluates the argument array read before checking the receiver.
		arg := fcnParameterDomain(rcd.ownedValues(), i)
		if field == nil {
			panic(NewNullPointerException())
		}
		ok, err := field.Member(arg)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (v *SetOfRcdsValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	allFinite := true
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.ownedValues() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
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
	return checkedProductSize(v.ownedValues(), func() error {
		return NewTLCRuntimeException(ECTLCModuleOverflow, "the number of elements in:\n"+ValuesPPR(v))
	})
}

func (v *SetOfRcdsValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.RcdSet != nil && v.RcdSet != DummyEnum {
		return v.RcdSet.IsNormalized()
	}
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	for i := range v.ownedNames() {
		value := v.fieldValue(i)
		if value == nil {
			panic(NewNullPointerException())
		}
		if !value.IsNormalized() {
			return false
		}
	}
	return true
}

func (v *SetOfRcdsValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.RcdSet != nil && v.RcdSet != DummyEnum {
		v.RcdSet.Normalize()
	} else {
		if v.ownedNames() == nil {
			panic(NewNullPointerException())
		}
		for i := range v.ownedNames() {
			value := v.fieldValue(i)
			if value == nil {
				panic(NewNullPointerException())
			}
			value.Normalize()
		}
	}
	return v
}

func (v *SetOfRcdsValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.ownedValues() {
		if value == nil {
			panic(NewNullPointerException())
		}
		value.DeepNormalize()
	}
	if v.RcdSet == nil {
		v.RcdSet = DummyEnum
	} else if v.RcdSet != DummyEnum {
		v.RcdSet.DeepNormalize()
	}
}

func (v *SetOfRcdsValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.ownedValues() {
		if value == nil {
			panic(NewNullPointerException())
		}
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *SetOfRcdsValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *SetOfRcdsValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetOfRcdsValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetOfRcdsValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set of records:\n" + ValuesPPR(v))
	}
	return ex.Value, nil
}

func (v *SetOfRcdsValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set of records:\n" + ValuesPPR(v))
	}
	return v, nil
}

func (v *SetOfRcdsValue) ToSetEnum() (*SetEnumValue, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.RcdSet != nil && v.RcdSet != DummyEnum {
		return v.RcdSet, nil
	}
	return setEnumFromEnumerationWithNormalization(v.Elements(), v.IsNormalized, v.CM)
}

func (v *SetOfRcdsValue) convertAndCache() (*SetEnumValue, error) {
	if v.RcdSet == DummyEnum {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.RcdSet = set
	} else if v.RcdSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.RcdSet = set
	}
	return v.RcdSet, nil
}

func (v *SetOfRcdsValue) Elements() (enumeration ValueEnumeration) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	defer raiseInitialEnumerationFailure(&enumeration)
	if v.RcdSet != nil && v.RcdSet != DummyEnum {
		return v.RcdSet.Elements()
	}
	empty, err := IsEmptyValue(v)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if empty {
		return EmptySet.Elements()
	}
	return newProductEnumeration(v.ownedValues(), func(fields []Value) Value {
		return NewRecordValue(v.ownedNames(), fields, true, v.CM)
	}, func(i int, value Value) error {
		return v.runtimeFailure("Attempted to enumerate a set of the form [l1 : v1, ..., ln : vn],\nbut can't enumerate the value of the `" + v.ownedNames()[i].String() + "' field:\n" + ValuesPPR(value))
	}, v.CM)
}

func (v *SetOfRcdsValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *SetOfRcdsValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if set := tryExpandedSet(swallow, func() (*SetEnumValue, error) {
		if shouldExpandProduct(v.ownedValues()) {
			return v.ToSetEnum()
		}
		return nil, nil
	}); set != nil {
		return appendValueString(set, sb, offset, swallow)
	}
	sb.WriteString("[")
	for i, name := range v.ownedNames() {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name.String() + ": ")
		appendValueString(v.ownedValues()[i], sb, offset, swallow)
	}
	sb.WriteString("]")
	return sb
}

func (v *SetOfRcdsValue) sortByNames() error {
	names, values := v.ownedNames(), v.ownedValues()
	for i := 1; i < len(names); i++ {
		cmp := names[0].Compare(names[i])
		if cmp == 0 {
			return v.runtimeFailure("Field name " + names[0].String() + " occurs multiple times in set of records.")
		}
		if cmp > 0 {
			names[0], names[i] = names[i], names[0]
			values[0], values[i] = values[i], values[0]
		}
	}
	for i := 2; i < len(names); i++ {
		j := i
		st := names[i]
		val := values[i]
		cmp := -1
		for j > 0 {
			cmp = st.Compare(names[j-1])
			if cmp >= 0 {
				break
			}
			names[j] = names[j-1]
			values[j] = values[j-1]
			j--
		}
		if cmp == 0 {
			return v.runtimeFailure("Field name " + names[i].String() + " occurs multiple times in set of records.")
		}
		names[j] = st
		values[j] = val
	}
	return nil
}

type SetOfFcnsValue struct {
	BaseValue
	Domain Value
	Range  Value
	FcnSet *SetEnumValue

	operandsCaptured bool
	domainOwner      Value
	rangeOwner       Value
}

func (v *SetOfFcnsValue) IsSubsetEq(other Value) (*BoolValue, error) {
	return defaultEnumerableSubsetEq(v, v, other)
}

func NewSetOfFcnsValue(domain, rangeValue Value, cms ...CostModel) *SetOfFcnsValue {
	return &SetOfFcnsValue{
		BaseValue: newBaseValue(cms...), Domain: domain, Range: rangeValue,
		operandsCaptured: true, domainOwner: domain, rangeOwner: rangeValue,
	}
}

func (v *SetOfFcnsValue) ownedDomain() Value {
	if v.operandsCaptured {
		return v.domainOwner
	}
	return v.Domain
}

func (v *SetOfFcnsValue) ownedRange() Value {
	if v.operandsCaptured {
		return v.rangeOwner
	}
	return v.Range
}

func (v *SetOfFcnsValue) captureOperands(domain, rangeValue Value) {
	v.Domain, v.Range = domain, rangeValue
	v.operandsCaptured, v.domainOwner, v.rangeOwner = true, domain, rangeValue
}

func (v *SetOfFcnsValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return SetOfFcnsValueKind
}
func (v *SetOfFcnsValue) KindString() string { return ValueKindString(v) }
func (v *SetOfFcnsValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *SetOfFcnsValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetOfFcnsValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if o, ok := other.(*SetOfFcnsValue); ok {
		unit1, err := IsEmptyValue(v.ownedDomain())
		if err != nil {
			return false, err
		}
		unit2, err := IsEmptyValue(o.ownedDomain())
		if err != nil {
			return false, err
		}
		if unit1 || unit2 {
			return unit1 == unit2, nil
		}
		empty1, err := IsEmptyValue(v.ownedRange())
		if err != nil {
			return false, err
		}
		empty2, err := IsEmptyValue(o.ownedRange())
		if err != nil {
			return false, err
		}
		if empty1 || empty2 {
			return empty1 == empty2, nil
		}
		domainEqual, err := v.ownedDomain().Equal(o.ownedDomain())
		if err != nil || !domainEqual {
			return domainEqual, err
		}
		return v.ownedRange().Equal(o.ownedRange())
	}
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetOfFcnsValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	fcn := asFcnRcdValue(elem)
	if fcn == nil {
		if mv, ok := elem.(*ModelValue); ok {
			return mv.ModelValueMember(v)
		}
		return false, v.runtimeFailure("Attempted to check if \n" + elem.String() + "\nwhich is not a TLC function value, is in the set of functions:\n" + ValuesPPR(v))
	}
	if fcn.ownedIntv() == nil {
		if err := fcn.normalizeFcn(); err != nil {
			return false, err
		}
		fdom := NewSetEnumValue(fcn.ownedDomain(), true)
		domainEqual, err := v.ownedDomain().Equal(fdom)
		if err != nil || !domainEqual {
			return domainEqual, err
		}
		for _, value := range fcn.ownedValues() {
			ok, err := v.ownedRange().Member(value)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	}
	domainEqual, err := fcn.ownedIntv().Equal(v.ownedDomain())
	if err != nil || !domainEqual {
		return domainEqual, err
	}
	for _, value := range fcn.ownedValues() {
		ok, err := v.ownedRange().Member(value)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (v *SetOfFcnsValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(v.ownedDomain()) {
		panic(NewNullPointerException())
	}
	finiteDomain, err := v.ownedDomain().IsFinite()
	if err != nil {
		return false, err
	}
	if finiteDomain {
		empty, err := IsEmptyValue(v.ownedDomain())
		if err != nil {
			return false, err
		}
		if empty {
			return true, nil
		}
	}
	if isNil(v.ownedRange()) {
		panic(NewNullPointerException())
	}
	finiteRange, err := v.ownedRange().IsFinite()
	if err != nil {
		return false, err
	}
	if finiteRange {
		empty, err := IsEmptyValue(v.ownedRange())
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
		size, err := v.ownedRange().Size()
		if err != nil {
			return false, err
		}
		return size == 1, nil
	}
	return false, nil
}

func (v *SetOfFcnsValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	domainEmpty, err := IsEmptyValue(v.ownedDomain())
	if err != nil {
		return 0, err
	}
	if domainEmpty {
		return 1, nil
	}
	rangeEmpty, err := IsEmptyValue(v.ownedRange())
	if err != nil {
		return 0, err
	}
	if rangeEmpty {
		return 0, nil
	}
	rangeSize, err := v.ownedRange().Size()
	if err != nil {
		return 0, err
	}
	if rangeSize == 1 {
		return 1, nil
	}
	domainSize, err := v.ownedDomain().Size()
	if err != nil {
		return 0, err
	}
	// The source asks for the range size again after the domain size.
	rangeSize, err = v.ownedRange().Size()
	if err != nil {
		return 0, err
	}
	size := int64(1)
	for i := 0; i < domainSize; i++ {
		size *= int64(rangeSize)
		if size < -2147483648 || size > 2147483647 {
			return 0, v.runtimeFailure("Overflow when computing the number of elements in:\n" + ValuesPPR(v))
		}
	}
	return int(size), nil
}

func (v *SetOfFcnsValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.FcnSet != nil && v.FcnSet != DummyEnum {
		return v.FcnSet.IsNormalized()
	}
	return v.ownedDomain().IsNormalized() && v.ownedRange().IsNormalized()
}

func (v *SetOfFcnsValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.FcnSet != nil && v.FcnSet != DummyEnum {
		v.FcnSet.Normalize()
	} else {
		if v.ownedDomain() == nil {
			panic(NewNullPointerException())
		}
		v.ownedDomain().Normalize()
		if v.ownedRange() == nil {
			panic(NewNullPointerException())
		}
		v.ownedRange().Normalize()
	}
	return v
}

func (v *SetOfFcnsValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	v.ownedDomain().DeepNormalize()
	v.ownedRange().DeepNormalize()
	if v.FcnSet == nil {
		v.FcnSet = DummyEnum
	} else if v.FcnSet != DummyEnum {
		v.FcnSet.DeepNormalize()
	}
}

func (v *SetOfFcnsValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	return v.ownedDomain().IsDefined() && v.ownedRange().IsDefined()
}

func (v *SetOfFcnsValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *SetOfFcnsValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetOfFcnsValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetOfFcnsValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set of functions:\n" + ValuesPPR(v))
	}
	return ex.Value, nil
}

func (v *SetOfFcnsValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the set of functions:\n" + ValuesPPR(v))
	}
	return v, nil
}

func (v *SetOfFcnsValue) ToSetEnum() (*SetEnumValue, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.FcnSet != nil && v.FcnSet != DummyEnum {
		return v.FcnSet, nil
	}
	return setEnumFromEnumerationWithNormalization(v.Elements(), v.IsNormalized, v.CM)
}

func (v *SetOfFcnsValue) convertAndCache() (*SetEnumValue, error) {
	if v.FcnSet == DummyEnum {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.FcnSet = set
	} else if v.FcnSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.FcnSet = set
	}
	return v.FcnSet, nil
}

func (v *SetOfFcnsValue) Elements() (enumeration ValueEnumeration) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	defer raiseInitialEnumerationFailure(&enumeration)
	if v.FcnSet != nil && v.FcnSet != DummyEnum {
		return v.FcnSet.Elements()
	}
	empty, err := IsEmptyValue(v)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if empty {
		return EmptySet.Elements()
	}
	if intv, ok := v.ownedDomain().(*IntervalValue); ok {
		return v.intervalDomainElements(intv)
	}
	domSet, err := tryToSetEnumValue(v.ownedDomain())
	if err != nil {
		return newErrorEnumeration(err)
	}
	if domSet == nil {
		return newErrorEnumeration(v.runtimeFailure("Attempted to enumerate a set of the form [D -> R],but the domain D:\n" + ValuesPPR(v.ownedDomain()) + "\ncannot be enumerated."))
	}
	if _, err := domSet.normalizeSet(); err != nil {
		return newErrorEnumeration(err)
	}
	dom := domSet.Elems.ToArray()
	return v.domainElements(dom, domSet.Elems)
}

func (v *SetOfFcnsValue) intervalDomainElements(intv *IntervalValue) ValueEnumeration {
	size, err := intv.Size()
	if err != nil {
		return newErrorEnumeration(err)
	}
	rangeEnum, ok := asEnumerable(v.ownedRange())
	if size > 0 && !ok {
		return newErrorEnumeration(v.runtimeFailure("Attempted to enumerate a set of the form [D -> R],but the range R:\n" + ValuesPPR(v.ownedRange()) + "\ncannot be enumerated."))
	}
	sets := make([]Value, size)
	for i := range sets {
		sets[i] = v.ownedRange()
	}
	if size == 0 {
		return &singleValueEnumeration{value: NewFcnRcdIntervalValue(intv, []Value{}, v.CM), cm: v.CM, secondary: 1}
	}
	_ = rangeEnum
	return newProductEnumeration(sets, func(elems []Value) Value {
		return NewFcnRcdIntervalValue(intv, elems, v.CM)
	}, func(i int, value Value) error {
		return v.runtimeFailure("Attempted to enumerate a set of the form [D -> R],but the range R:\n" + ValuesPPR(v.ownedRange()) + "\ncannot be enumerated.")
	}, v.CM)
}

func (v *SetOfFcnsValue) domainElements(dom []Value, domainVectors ...*ValueVec) ValueEnumeration {
	if len(domainVectors) > 1 {
		panic(NewIllegalArgumentException("domainElements accepts at most one domain vector"))
	}
	rangeEnum, ok := asEnumerable(v.ownedRange())
	if len(dom) > 0 && !ok {
		return newErrorEnumeration(v.runtimeFailure("Attempted to enumerate a set of the form [D -> R],but the range R:\n" + ValuesPPR(v.ownedRange()) + "\ncannot be enumerated."))
	}
	if len(dom) == 0 {
		return &singleValueEnumeration{value: NewFcnRcdValue(dom, []Value{}, true, v.CM), cm: v.CM, secondary: 1}
	}
	_ = rangeEnum
	sets := make([]Value, len(dom))
	for i := range sets {
		sets[i] = v.ownedRange()
	}
	return newProductEnumeration(sets, func(elems []Value) Value {
		if len(domainVectors) != 0 {
			return NewFcnRcdValueVec(domainVectors[0], elems, true, v.CM)
		}
		return NewFcnRcdValue(dom, elems, true, v.CM)
	}, func(i int, value Value) error {
		return v.runtimeFailure("Attempted to enumerate a set of the form [D -> R],but the range R:\n" + ValuesPPR(v.ownedRange()) + "\ncannot be enumerated.")
	}, v.CM)
}

func (v *SetOfFcnsValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *SetOfFcnsValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if set := tryExpandedSet(swallow, func() (*SetEnumValue, error) {
		if shouldExpandFcnSet(v.ownedDomain(), v.ownedRange()) {
			return v.ToSetEnum()
		}
		return nil, nil
	}); set != nil {
		return appendValueString(set, sb, offset, swallow)
	}
	sb.WriteString("[")
	appendValueString(v.ownedDomain(), sb, offset, swallow)
	sb.WriteString(" -> ")
	appendValueString(v.ownedRange(), sb, offset, swallow)
	sb.WriteString("]")
	return sb
}

type SubsetValue struct {
	BaseValue
	Set  Value
	PSet *SetEnumValue
}

func (v *SubsetValue) IsSubsetEq(other Value) (resultBool *BoolValue, err error) {
	defer catchValueFailure(v, &err)
	if subset, ok := other.(*SubsetValue); ok {
		if enum, ok := asEnumerable(v.Set); ok {
			return enum.IsSubsetEq(subset.Set)
		}
	}
	return defaultEnumerableSubsetEq(v, v, other)
}

func NewSubsetValue(set Value, cms ...CostModel) *SubsetValue {
	return &SubsetValue{BaseValue: newBaseValue(cms...), Set: set}
}

func (v *SubsetValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return SubsetValueKind
}
func (v *SubsetValue) KindString() string { return ValueKindString(v) }
func (v *SubsetValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *SubsetValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	if v == nil {
		panic(NewNullPointerException())
	}
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
	if v == nil {
		panic(NewNullPointerException())
	}
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
		if err := e.Err(); err != nil {
			return false, err
		}
		if value == nil {
			return true, nil
		}
		ok, err := set.Member(value)
		if err != nil || !ok {
			return ok, err
		}
	}
}

func (v *SubsetValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(v.Set) {
		panic(NewNullPointerException())
	}
	return v.Set.IsFinite()
}

func (v *SubsetValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(v.Set) {
		panic(NewNullPointerException())
	}
	size, err := v.Set.Size()
	if err != nil {
		return 0, err
	}
	if size >= 31 {
		return 0, NewTLCRuntimeException(ECTLCModuleOverflow, "the number of elements in:\n"+ValuesPPR(v))
	}
	return int(int32(1) << (uint32(size) & 31)), nil
}

func (v *SubsetValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	return v.PSet != nil && v.PSet != DummyEnum && v.PSet.IsNormalized()
}

func (v *SubsetValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.PSet != nil && v.PSet != DummyEnum {
		v.PSet.Normalize()
	} else {
		v.Set.Normalize()
	}
	return v
}

func (v *SubsetValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	v.Set.DeepNormalize()
	if v.PSet == nil {
		v.PSet = DummyEnum
	} else if v.PSet != DummyEnum {
		v.PSet.DeepNormalize()
	}
}

func (v *SubsetValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	return v.Set.IsDefined()
}

func (v *SubsetValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *SubsetValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SubsetValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.convertAndCache()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SubsetValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SubsetValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SubsetValue) ToSetEnum() (*SetEnumValue, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.PSet != nil && v.PSet != DummyEnum {
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
	if v.PSet == DummyEnum {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.PSet = set
	} else if v.PSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.PSet = set
	}
	return v.PSet, nil
}

func (v *SubsetValue) Elements() (enumeration ValueEnumeration) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	defer raiseInitialEnumerationFailure(&enumeration)
	if v.PSet != nil && v.PSet != DummyEnum {
		return v.PSet.Elements()
	}
	return v.ElementsNormalized()
}

// ElementsNormalized ports the direct SubsetValue.elementsNormalized entry
// point, independently of the cached powerset used by Elements.
func (v *SubsetValue) ElementsNormalized() (enumeration ValueEnumeration) {
	defer raiseInitialEnumerationFailure(&enumeration)
	if v == nil || v.Set == nil {
		panic(NewNullPointerException())
	}
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
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *SubsetValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	requireValueEnumerationReceiver(e)
	e.done = false
}

func (e *singleValueEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
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
	requireValueEnumerationReceiver(e)
	e.done = false
	e.err = nil
	e.k = 0
	e.cur = newKSubsetEnumeration(e.elems, e.k, e.cm)
}

func (e *subsetEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
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

// Record-set normalization walks names and reads the corresponding value slot.
func (v *SetOfRcdsValue) fieldValue(index int) Value {
	values := v.ownedValues()
	if values == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(values) {
		panic(NewArrayIndexOutOfBoundsException(index, len(values)))
	}
	return values[index]
}
