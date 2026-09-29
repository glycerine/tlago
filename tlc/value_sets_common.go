package tlc

import "math"

func IsEmptyValue(value Value) (bool, error) {
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
	case *SetOfTuplesValue:
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
			}
		}
		return false, nil
	case *SetOfRcdsValue:
		for _, set := range v.Values {
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
	case *UserValue:
		if obj, ok := v.UserObj.(UserObjWithIsEmpty); ok {
			return obj.IsEmpty()
		}
		return false, v.unsupported("should not call isEmpty() on value %s", v)
	default:
		return false, newTLCError(ECGeneral, "should not call isEmpty() on value %s", value)
	}
}

func toSetEnumValue(value Value) (*SetEnumValue, error) {
	switch v := value.(type) {
	case *SetEnumValue:
		return v, nil
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
	default:
		return nil, newTLCError(ECGeneral, "value %s cannot be converted to an enumerated set", value)
	}
}

func checkedProductSize(values []Value, rendered string) (int, error) {
	size := int64(1)
	for _, value := range values {
		next, err := value.Size()
		if err != nil {
			return 0, err
		}
		size *= int64(next)
		if size < math.MinInt32 || size > math.MaxInt32 {
			return 0, newTLCError(ECGeneral, "overflow when computing the number of elements in %s", rendered)
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
			return false
		}
		if next == 0 {
			return true
		}
		if size <= math.MaxInt32 {
			size *= int64(next)
		}
	}
	return size < int64(Globals.EnumBound)
}

type productEnumeration struct {
	enums        []ValueEnumeration
	currentElems []Value
	done         bool
	err          error
	makeValue    func([]Value) Value
}

func newProductEnumeration(sets []Value, makeValue func([]Value) Value, errf func(int, Value) error) ValueEnumeration {
	out := &productEnumeration{
		enums:        make([]ValueEnumeration, len(sets)),
		currentElems: make([]Value, len(sets)),
		makeValue:    makeValue,
	}
	for i, set := range sets {
		enum, ok := asEnumerable(set)
		if !ok {
			out.err = errf(i, set)
			out.done = true
			return out
		}
		out.enums[i] = enum.Elements()
		out.currentElems[i] = out.enums[i].NextElement()
		if err := out.enums[i].Err(); err != nil {
			out.err = err
			out.done = true
			return out
		}
		if out.currentElems[i] == nil {
			out.done = true
			return out
		}
	}
	return out
}

func (e *productEnumeration) Reset() {
	if e.err != nil {
		return
	}
	for i := range e.enums {
		e.enums[i].Reset()
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			return
		}
		if e.currentElems[i] == nil {
			e.done = true
			return
		}
	}
	e.done = false
}

func (e *productEnumeration) NextElement() Value {
	if e.done || e.err != nil {
		return nil
	}
	elems := make([]Value, len(e.currentElems))
	copy(elems, e.currentElems)
	for i := len(e.currentElems) - 1; i >= 0; i-- {
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			break
		}
		if e.currentElems[i] != nil {
			break
		}
		if i == 0 {
			e.done = true
			break
		}
		e.enums[i].Reset()
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			break
		}
	}
	return e.makeValue(elems)
}

func (e *productEnumeration) Err() error {
	return e.err
}
