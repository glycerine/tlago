package tlc

import (
	"math"
	"strings"
)

func IsEmptyValue(value Value) (resultBool bool, err error) {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(value, &err)
	switch v := value.(type) {
	case *SetEnumValue:
		return v.Elems.Len() == 0, nil
	case *IntervalValue:
		size, err := v.Size()
		return size == 0, err
	case *SetCupValue:
		enum := v.Elements()
		elem := enum.NextElement()
		if err := enum.Err(); err != nil {
			return false, err
		}
		return elem == nil, nil
	case *SetCapValue:
		enum := v.Elements()
		elem := enum.NextElement()
		if err := enum.Err(); err != nil {
			return false, err
		}
		return elem == nil, nil
	case *SetDiffValue:
		enum := v.Elements()
		elem := enum.NextElement()
		if err := enum.Err(); err != nil {
			return false, err
		}
		return elem == nil, nil
	case *UnionValue:
		enum := v.Elements()
		elem := enum.NextElement()
		if err := enum.Err(); err != nil {
			return false, err
		}
		return elem == nil, nil
	case *SetPredValue:
		enum := v.Elements()
		elem := enum.NextElement()
		if err := enum.Err(); err != nil {
			return false, err
		}
		return elem == nil, nil
	case *SetOfTuplesValue:
		if v.Sets == nil {
			panic(NewNullPointerException())
		}
		for _, set := range v.Sets {
			empty, err := IsEmptyValue(set)
			if err != nil {
				return false, err
			}
			if empty {
				return true, nil
			}
		}
		return false, nil
	case *SetOfRcdsValue:
		if v.Values == nil {
			panic(NewNullPointerException())
		}
		for _, set := range v.Values {
			empty, err := IsEmptyValue(set)
			if err != nil {
				return false, err
			}
			if empty {
				return true, nil
			}
		}
		return false, nil
	case *SetOfFcnsValue:
		domainEmpty, err := IsEmptyValue(v.Domain)
		if err != nil {
			return false, err
		}
		if domainEmpty {
			return false, nil
		}
		rangeEmpty, err := IsEmptyValue(v.Range)
		if err != nil {
			return false, err
		}
		return !domainEmpty && rangeEmpty, nil
	case *SubsetValue:
		return false, nil
	case *KSubsetValue:
		if v.K < 0 {
			return true, nil
		}
		finite, err := v.Set.IsFinite()
		if err != nil {
			return false, err
		}
		if finite {
			size, err := v.Set.Size()
			if err != nil {
				return false, err
			}
			return size < v.K, nil
		}
		return false, nil
	case *UserValue:
		if obj, ok := v.UserObj.(UserObjWithIsEmpty); ok {
			return obj.IsEmpty()
		}
		return false, NewTLCRuntimeExceptionMessage("Shouldn't call isEmpty() on value " + v.UserObj.String())
	default:
		message := "Shouldn't call isEmpty() on value " + ValuesPPR(value)
		if source := valueSource(value); source != nil {
			return false, NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
		}
		return false, NewTLCRuntimeExceptionMessage(message)
	}
}

func toSetEnumValue(value Value) (*SetEnumValue, error) {
	set, err := tryToSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, newTLCError(ECGeneral, "value %s cannot be converted to an enumerated set", value)
	}
	return set, nil
}

func tryToSetEnumValue(value Value) (*SetEnumValue, error) {
	switch v := value.(type) {
	case *SetEnumValue:
		return v, nil
	case *SetPredValue:
		return v.ToSetEnum()
	case *IntervalValue:
		return v.ToSetEnum(), nil
	case *SetCupValue:
		return v.ToSetEnum()
	case *SetCapValue:
		return v.ToSetEnum()
	case *SetDiffValue:
		return v.ToSetEnum()
	case *UnionValue:
		return v.ToSetEnum()
	case *SetOfTuplesValue:
		return v.ToSetEnum()
	case *SetOfRcdsValue:
		return v.ToSetEnum()
	case *SetOfFcnsValue:
		return v.ToSetEnum()
	case *SubsetValue:
		return v.ToSetEnum()
	case *KSubsetValue:
		return v.ToSetEnum()
	default:
		return nil, nil
	}
}

func checkedProductSize(values []Value, overflow func() error) (int, error) {
	size := int64(1)
	for _, value := range values {
		next, err := value.Size()
		if err != nil {
			return 0, err
		}
		size *= int64(next)
		if size < math.MinInt32 || size > math.MaxInt32 {
			return 0, overflow()
		}
	}
	return int(size), nil
}

func shouldExpandProduct(values []Value) bool {
	if !Globals.Expand {
		return false
	}
	size := int64(1)
	for _, value := range values {
		next, err := value.Size()
		if err != nil {
			panic(err)
		}
		if next == 0 {
			return 0 < Globals.EnumBound
		}
		if size <= math.MaxInt32 {
			size *= int64(next)
		}
	}
	return size < int64(Globals.EnumBound)
}

// Java's product printers catch failures during the expansion decision and
// materialization. Printing the resulting set remains outside that catch.
func tryExpandedSet(swallow bool, build func() (*SetEnumValue, error)) (set *SetEnumValue) {
	tryValueString(swallow, func() {
		converted, err := build()
		if err != nil {
			panic(err)
		}
		set = converted
	})
	return set
}

// Cup, cap, difference, and predicate sets include expanded printing in their
// catch. A failed printer leaves its partial text in the same caller buffer.
func tryExpandedSetString(sb *strings.Builder, offset int, swallow bool, build func() (*SetEnumValue, error)) (*strings.Builder, bool) {
	var expanded *strings.Builder
	ok := tryValueString(swallow, func() {
		set, err := build()
		if err != nil {
			panic(err)
		}
		expanded = appendValueString(set, sb, offset, swallow)
	})
	return expanded, ok
}

func tryValueSizeBelow(value Value, bound int, swallow bool) (below bool) {
	tryValueString(swallow, func() {
		size, err := value.Size()
		if err != nil {
			panic(err)
		}
		below = size < bound
	})
	return below
}

func shouldExpandFcnSet(domain, rangeValue Value) bool {
	if !Globals.Expand {
		return false
	}
	empty, err := IsEmptyValue(domain)
	if err != nil {
		panic(err)
	}
	size := int64(1)
	if !empty {
		domainSize, err := domain.Size()
		if err != nil {
			panic(err)
		}
		rangeSize, err := rangeValue.Size()
		if err != nil {
			panic(err)
		}
		for i := 0; i < domainSize && size <= math.MaxInt32; i++ {
			size *= int64(rangeSize)
		}
	}
	return size < int64(Globals.EnumBound)
}

// Only failures already present when elements() returns belong to its Java
// catch boundary. Raise them before returning, so the enclosing catch adds its
// source frame once. Later nextElement/reset failures stay outside this catch.
func raiseInitialEnumerationFailure(enumeration *ValueEnumeration) {
	if *enumeration != nil {
		if err := (*enumeration).Err(); err != nil {
			panic(err)
		}
	}
}

type productEnumeration struct {
	cm           CostModel
	enums        []ValueEnumeration
	currentElems []Value
	done         bool
	err          error
	makeValue    func([]Value) Value
}

func newProductEnumeration(sets []Value, makeValue func([]Value) Value, errf func(int, Value) error, cms ...CostModel) ValueEnumeration {
	out := &productEnumeration{
		enums:        make([]ValueEnumeration, len(sets)),
		currentElems: make([]Value, len(sets)),
		makeValue:    makeValue,
	}
	if len(cms) > 0 {
		out.cm = cms[0]
	}
	for i, set := range sets {
		enum, ok := asEnumerable(set)
		if !ok {
			out.err = errf(i, set)
			out.done = true
			return out
		}
		out.enums[i] = enum.Elements()
		out.currentElems[i] = nextEnumerationElement(out.enums[i])
		if out.currentElems[i] == nil {
			out.enums = nil
			out.done = true
			return out
		}
	}
	return out
}

func (e *productEnumeration) Reset() {
	if e.enums == nil {
		return
	}
	for i := range e.enums {
		resetEnumeration(e.enums[i])
		e.currentElems[i] = nextEnumerationElement(e.enums[i])
	}
	e.done = false
}

func (e *productEnumeration) NextElement() Value {
	if e.done {
		return nil
	}
	elems := make([]Value, len(e.currentElems))
	e.cm.incValueSecondary(int64(len(elems)))
	copy(elems, e.currentElems)
	for i := len(e.currentElems) - 1; i >= 0; i-- {
		e.currentElems[i] = nextEnumerationElement(e.enums[i])
		if e.currentElems[i] != nil {
			break
		}
		if i == 0 {
			e.done = true
			break
		}
		resetEnumeration(e.enums[i])
		e.currentElems[i] = nextEnumerationElement(e.enums[i])
	}
	return e.makeValue(elems)
}

func (e *productEnumeration) Err() error {
	return e.err
}

// Raise a failed source invocation before the caller assigns its result. The
// iterator may advance internally before throwing; its caller retains the
// previous array slot or child reference.
func nextEnumerationElement(enumeration ValueEnumeration) Value {
	if enumeration == nil {
		panic(NewNullPointerException())
	}
	value := enumeration.NextElement()
	if err := enumeration.Err(); err != nil {
		panic(err)
	}
	return value
}

func resetEnumeration(enumeration ValueEnumeration) {
	if enumeration == nil {
		panic(NewNullPointerException())
	}
	enumeration.Reset()
	if err := enumeration.Err(); err != nil {
		panic(err)
	}
}
