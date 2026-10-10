package tlc

func SequencesExtSetToSeq(value Value) (Value, error) {
	if value == nil {
		panic(NewNullPointerException())
	}
	set, err := tryToSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	if set == nil {
		panic(NewNullPointerException())
	}
	set.Normalize()
	return NewTupleValue(set.Elems.ToArray()), nil
}

func SequencesExtSetToSeqs(value Value) (Value, error) {
	if value == nil {
		panic(NewNullPointerException())
	}
	set, err := tryToSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "SetToSeqs", "a finite set", ValuesPPR(value))
	}
	set.Normalize()
	elems := set.Elems
	length := elems.Len()
	if length == 0 {
		return NewSetEnumValue([]Value{emptyFcnValue()}, true), nil
	}

	factorial := int32(1)
	domain := make([]Value, length)
	idxArray := make([]int, length)
	inUse := make([]bool, length)
	for i := 0; i < length; i++ {
		domain[i] = elems.At(i)
		idxArray[i] = i
		inUse[i] = true
		factorial *= int32(i + 1)
	}

	fcns := NewValueVec(int(factorial))
	for {
		vals := make([]Value, length)
		for i := 0; i < length; i++ {
			vals[i] = domain[idxArray[i]]
		}
		fcns.Add(NewTupleValue(vals))

		i := length - 1
		found := false
		for ; i >= 0; i-- {
			for j := idxArray[i] + 1; j < length; j++ {
				if !inUse[j] {
					inUse[j] = true
					inUse[idxArray[i]] = false
					idxArray[i] = j
					found = true
					break
				}
			}
			if found {
				break
			}
			if i == 0 {
				return NewSetEnumValueVec(fcns, false), nil
			}
			inUse[idxArray[i]] = false
		}
		for j := i + 1; j < length; j++ {
			for k := 0; k < length; k++ {
				if !inUse[k] {
					inUse[k] = true
					idxArray[j] = k
					break
				}
			}
		}
	}
}

func SequencesExtContains(seq Value, elem Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Contains", "sequence", ValuesPPR(seq))
	}
	if tuple.Elems == nil {
		panic(NewNullPointerException())
	}
	for _, value := range tuple.Elems {
		if value == nil {
			panic(NewNullPointerException())
		}
		eq, err := value.Equal(elem)
		if err != nil {
			return nil, err
		}
		if eq {
			return BoolTrue, nil
		}
	}
	return BoolFalse, nil
}

func SequencesExtLongestCommonPrefix(value Value) (result Value, err error) {
	if value == nil {
		panic(NewNullPointerException())
	}
	set, err := tryToSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "non-empty set", ValuesPPR(value))
	}
	set.Normalize()
	elems := set.Elems
	if elems == nil {
		panic(NewNullPointerException())
	}
	if elems.Empty() {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "non-empty set", ValuesPPR(value))
	}
	// Java catches only direct cast/null failures in the prefix traversal. A
	// child failure already wrapped with source information passes through.
	defer func() {
		if failure := recover(); failure != nil {
			switch failure.(type) {
			case *NullPointerException, *ClassCastException:
				result = nil
				err = newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "sequence", ValuesPPR(value))
			default:
				panic(failure)
			}
		} else {
			switch err.(type) {
			case *NullPointerException, *ClassCastException:
				result = nil
				err = newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "sequence", ValuesPPR(value))
			}
		}
	}()
	first := elems.At(0)
	if str, ok := asStringValue(first); ok {
		if str.Val == nil {
			panic(NewNullPointerException())
		}
		prefix := javaStringUTF16(str.Val.String())
		upper := len(prefix)
		for i := 1; i < elems.Len(); i++ {
			other, ok := asStringValue(elems.At(i))
			if !ok {
				panic(NewClassCastException("Cannot cast value to tlc2.value.impl.StringValue"))
			}
			if other.Val == nil {
				panic(NewNullPointerException())
			}
			chars := javaStringUTF16(other.Val.String())
			for idx := 0; idx < upper; idx++ {
				if idx >= len(chars) {
					panic(NewStringIndexOutOfBoundsException(idx, len(chars)))
				}
				if prefix[idx] != chars[idx] {
					upper = idx
					if upper == 0 {
						return NewStringValue(""), nil
					}
				}
			}
		}
		return NewStringValue(javaStringFromUTF16(prefix[:upper])), nil
	}
	tuple := sequenceTuple(first)
	if tuple == nil || tuple.Elems == nil {
		panic(NewNullPointerException())
	}
	prefix := tuple.Elems
	upper := len(prefix)
	for i := 1; i < elems.Len(); i++ {
		otherTuple := sequenceTuple(elems.At(i))
		if otherTuple == nil {
			panic(NewNullPointerException())
		}
		other := otherTuple.Elems
		for idx := 0; idx < upper; idx++ {
			if other == nil {
				panic(NewNullPointerException())
			}
			if idx >= len(other) {
				panic(NewArrayIndexOutOfBoundsException(idx, len(other)))
			}
			if prefix[idx] == nil {
				panic(NewNullPointerException())
			}
			equal, err := prefix[idx].Equal(other[idx])
			if err != nil {
				return nil, err
			}
			if !equal {
				upper = idx
				if upper == 0 {
					return EmptyTuple, nil
				}
			}
		}
	}
	if upper == 0 {
		return EmptyTuple, nil
	}
	copied := make([]Value, upper)
	copy(copied, prefix[:upper])
	return NewTupleValue(copied), nil
}

func SequencesExtFoldSeq(op Value, base Value, seq Value) (Value, error) {
	return FunctionsFoldFunction(op, base, seq)
}

func SequencesExtFoldLeft(op Value, base Value, seq Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldLeft", "sequence", ValuesPPR(seq))
	}
	length, err := tuple.Size()
	if err != nil {
		return nil, err
	}
	args := []Value{base, nil}
	elems := tuple.Elems
	for i := 0; i < length; i++ {
		args[1] = elems[i]
		value, err := sequenceOperatorEval(op, args)
		if err != nil {
			return nil, err
		}
		args[0] = value
	}
	return args[0], nil
}

func SequencesExtFoldRight(op Value, seq Value, base Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "FoldRight", "sequence", ValuesPPR(seq))
	}
	length, err := tuple.Size()
	if err != nil {
		return nil, err
	}
	args := []Value{nil, base}
	elems := tuple.Elems
	for i := length - 1; i >= 0; i-- {
		args[0] = elems[i]
		value, err := sequenceOperatorEval(op, args)
		if err != nil {
			return nil, err
		}
		args[1] = value
	}
	return args[1], nil
}

func SequencesExtFoldLeftDomain(op Value, base Value, seq Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldLeftDomain", "sequence", ValuesPPR(seq))
	}
	args := []Value{base, nil}
	for i := 0; i < sequenceSize(tuple); i++ {
		args[1] = NewIntValue(int32(i + 1))
		value, err := sequenceOperatorEval(op, args)
		if err != nil {
			return nil, err
		}
		args[0] = value
	}
	return args[0], nil
}

func SequencesExtFoldRightDomain(op Value, seq Value, base Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "FoldRightDomain", "sequence", ValuesPPR(seq))
	}
	length, err := tuple.Size()
	if err != nil {
		return nil, err
	}
	args := []Value{nil, base}
	for i := length - 1; i >= 0; i-- {
		args[0] = NewIntValue(int32(i + 1))
		value, err := sequenceOperatorEval(op, args)
		if err != nil {
			return nil, err
		}
		args[1] = value
	}
	return args[1], nil
}

func SequencesExtReplaceFirstSubSeq(replacement Value, subseq Value, target Value) (Value, error) {
	r, s, t, ok := sequencesExtReplacementStrings(replacement, subseq, target)
	if !ok {
		return nil, nil
	}
	index := sequencesExtStringIndex(t, s, 0)
	if index < 0 {
		return NewStringValue(javaStringFromUTF16(t)), nil
	}
	out := append([]uint16{}, t[:index]...)
	out = append(out, r...)
	out = append(out, t[index+len(s):]...)
	return NewStringValue(javaStringFromUTF16(out)), nil
}

func SequencesExtReplaceAllSubSeqs(replacement Value, subseq Value, target Value) (Value, error) {
	r, s, t, ok := sequencesExtReplacementStrings(replacement, subseq, target)
	if !ok {
		return nil, nil
	}
	out := make([]uint16, 0, len(t))
	if len(s) == 0 {
		out = append(out, r...)
		for i, unit := range t {
			if i != 0 {
				out = append(out, r...)
			}
			out = append(out, unit)
		}
	} else {
		start := 0
		for {
			index := sequencesExtStringIndex(t, s, start)
			if index < 0 {
				out = append(out, t[start:]...)
				break
			}
			out = append(out, t[start:index]...)
			out = append(out, r...)
			start = index + len(s)
		}
	}
	return NewStringValue(javaStringFromUTF16(out)), nil
}

// The evaluating override handles strings; a nil result selects the parsed
// TLA+ definition for other values. Java dereferences target, pattern, then
// replacement after evaluating all three arguments.
func sequencesExtReplacementStrings(replacement, subseq, target Value) (r, s, t []uint16, ok bool) {
	rv, rok := asStringValue(replacement)
	sv, sok := asStringValue(subseq)
	tv, tok := asStringValue(target)
	if !rok || !sok || !tok {
		return nil, nil, nil, false
	}
	if tv.Val == nil {
		panic(NewNullPointerException())
	}
	t = javaStringUTF16(tv.Val.String())
	if sv.Val == nil {
		panic(NewNullPointerException())
	}
	s = javaStringUTF16(sv.Val.String())
	if rv.Val == nil {
		panic(NewNullPointerException())
	}
	r = javaStringUTF16(rv.Val.String())
	return r, s, t, true
}

func sequencesExtStringIndex(target, pattern []uint16, start int) int {
	for i := start; i <= len(target)-len(pattern); i++ {
		matches := true
		for j, unit := range pattern {
			if target[i+j] != unit {
				matches = false
				break
			}
		}
		if matches {
			return i
		}
	}
	return -1
}

func SequencesExtIsPrefix(left Value, right Value) (Value, error) {
	s1, leftString := asStringValue(left)
	s2, rightString := asStringValue(right)
	if leftString && rightString {
		// Java obtains the target string before the candidate prefix string.
		if s2.Val == nil {
			panic(NewNullPointerException())
		}
		target := javaStringUTF16(s2.Val.String())
		if s1.Val == nil {
			panic(NewNullPointerException())
		}
		prefix := javaStringUTF16(s1.Val.String())
		if len(prefix) > len(target) {
			return BoolFalse, nil
		}
		for i, unit := range prefix {
			if unit != target[i] {
				return BoolFalse, nil
			}
		}
		return BoolTrue, nil
	}
	s := sequenceTuple(left)
	t := sequenceTuple(right)
	if sequenceSize(s) <= sequenceSize(t) {
		for i := 0; i < sequenceSize(s); i++ {
			elem := s.Elems[i]
			other := t.Elems[i]
			if elem == nil {
				panic(NewNullPointerException())
			}
			equal, err := elem.Equal(other)
			if err != nil {
				return nil, err
			}
			if !equal {
				return BoolFalse, nil
			}
		}
		return BoolTrue, nil
	}
	return BoolFalse, nil
}

func SequencesExtSelectInSeq(seq Value, test Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectInSeq", "sequence", ValuesPPR(seq))
	}
	length, err := tuple.Size()
	if err != nil {
		return nil, err
	}
	args := make([]Value, 1)
	for i := 0; i < length; i++ {
		args[0] = tuple.Elems[i]
		value, err := sequenceOperatorEval(test, args)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SelectInSeq", "boolean-valued operator", ValuesPPR(test))
		}
		if boolValue.Val {
			return NewIntValue(int32(i + 1)), nil
		}
	}
	return IntZero, nil
}

func SequencesExtSelectInSubSeq(seq Value, from Value, to Value, test Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectInSubSeq", "sequence", ValuesPPR(seq))
	}
	fromInt, ok := from.(*IntValue)
	if !ok {
		if from == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectInSubSeq", "natural", ValuesPPR(from))
	}
	toInt, ok := to.(*IntValue)
	if !ok {
		if to == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SelectInSubSeq", "natural", ValuesPPR(to))
	}
	start := int(fromInt.Val)
	end := int(toInt.Val)
	if start > end {
		return IntZero, nil
	}
	if start < 1 || start > sequenceSize(tuple) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SelectInSubSeq", "first", ValuesPPR(seq), ValuesPPR(from))
	}
	if end < 1 || end > sequenceSize(tuple) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SelectInSubSeq", "first", ValuesPPR(seq), ValuesPPR(to))
	}
	args := make([]Value, 1)
	for i := start; i <= end; i++ {
		args[0] = tuple.Elems[i-1]
		value, err := sequenceOperatorEval(test, args)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "fourth", "SelectInSubSeq", "boolean-valued operator", ValuesPPR(test))
		}
		if boolValue.Val {
			return NewIntValue(int32(i)), nil
		}
	}
	return IntZero, nil
}

func SequencesExtSelectLastInSeq(seq Value, test Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectLastInSeq", "sequence", ValuesPPR(seq))
	}
	length, err := tuple.Size()
	if err != nil {
		return nil, err
	}
	args := make([]Value, 1)
	for i := length - 1; i >= 0; i-- {
		args[0] = tuple.Elems[i]
		value, err := sequenceOperatorEval(test, args)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SelectLastInSeq", "boolean-valued function", ValuesPPR(test))
		}
		if boolValue.Val {
			return NewIntValue(int32(i + 1)), nil
		}
	}
	return IntZero, nil
}

func SequencesExtSelectLastInSubSeq(seq Value, from Value, to Value, test Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectLastInSubSeq", "sequence", ValuesPPR(seq))
	}
	fromInt, ok := from.(*IntValue)
	if !ok {
		if from == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectLastInSubSeq", "natural", ValuesPPR(from))
	}
	toInt, ok := to.(*IntValue)
	if !ok {
		if to == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SelectLastInSubSeq", "natural", ValuesPPR(to))
	}
	start := int(fromInt.Val)
	end := int(toInt.Val)
	if start > end {
		return IntZero, nil
	}
	if start < 1 || start > sequenceSize(tuple) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SelectLastInSubSeq", "first", ValuesPPR(seq), ValuesPPR(from))
	}
	if end < 1 || end > sequenceSize(tuple) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SelectLastInSubSeq", "first", ValuesPPR(seq), ValuesPPR(to))
	}
	args := make([]Value, 1)
	for i := end; i >= start; i-- {
		args[0] = tuple.Elems[i-1]
		value, err := sequenceOperatorEval(test, args)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "fourth", "SelectLastInSubSeq", "boolean-valued function", ValuesPPR(test))
		}
		if boolValue.Val {
			return NewIntValue(int32(i)), nil
		}
	}
	return IntZero, nil
}

func SequencesExtRemoveFirst(seq Value, elem Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "RemoveFirst", "sequence", ValuesPPR(seq))
	}
	if tuple.Elems == nil {
		panic(NewNullPointerException())
	}
	out := make([]Value, 0, len(tuple.Elems))
	found := false
	for _, value := range tuple.Elems {
		if !found {
			if value == nil {
				panic(NewNullPointerException())
			}
			eq, err := value.Equal(elem)
			if err != nil {
				return nil, err
			}
			if eq {
				found = true
				continue
			}
		}
		out = append(out, value)
	}
	return NewTupleValue(out), nil
}

func SequencesExtRemoveFirstMatch(seq Value, test Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "RemoveFirstMatch", "sequence", ValuesPPR(seq))
	}
	length, err := tuple.Size()
	if err != nil {
		return nil, err
	}
	out := make([]Value, 0, length)
	args := make([]Value, 1)
	found := false
	for i := 0; i < sequenceSize(tuple); i++ {
		if !found {
			args[0] = tuple.Elems[i]
			value, err := sequenceOperatorEval(test, args)
			if err != nil {
				return nil, err
			}
			boolValue, ok := value.(*BoolValue)
			if !ok {
				return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "RemoveFirstMatch", "boolean-valued function", ValuesPPR(test))
			}
			if boolValue.Val {
				found = true
				continue
			}
		}
		out = append(out, tuple.Elems[i])
	}
	return NewTupleValue(out), nil
}

func SequencesExtSuffixes(seq Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Suffixes", "sequence", ValuesPPR(seq))
	}
	if tuple.Elems == nil {
		panic(NewNullPointerException())
	}
	vals := make([]Value, len(tuple.Elems)+1)
	vals[0] = EmptyTuple
	for i := len(tuple.Elems) - 1; i >= 0; i-- {
		suffix := make([]Value, len(tuple.Elems)-i)
		copy(suffix, tuple.Elems[i:])
		vals[len(tuple.Elems)-i] = NewTupleValue(suffix)
	}
	return NewSetEnumValue(vals, true), nil
}

func SequencesExtAllSubSeqs(seq Value) (Value, error) {
	tuple := sequenceTuple(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "AllSubSeqs", "sequence", ValuesPPR(seq))
	}
	if tuple.Elems == nil {
		panic(NewNullPointerException())
	}
	n := len(tuple.Elems)
	// The source casts Math.pow(2, n) to int before allocating its array.
	// At n >= 31 this saturates to MAX_INT, which exceeds the source array limit.
	if n >= 31 {
		panic(NewOutOfMemoryError("Requested array size exceeds VM limit"))
	}
	total := 1 << n
	vals := make([]Value, total)
	for mask := 0; mask < total; mask++ {
		sub := make([]Value, 0, n)
		for j := 0; j < n; j++ {
			if (mask & (1 << j)) != 0 {
				sub = append(sub, tuple.Elems[j])
			}
		}
		vals[mask] = NewTupleValue(sub)
	}
	return NewSetEnumValue(vals, false), nil
}

// These helpers retain direct Java dereferences without adding a module catch.
func sequenceOperatorEval(operator Value, args []Value) (Value, error) {
	if operator == nil {
		panic(NewNullPointerException())
	}
	return EvalOperatorValue(operator, args, EvalClear)
}

func sequenceSize(sequence *TupleValue) int {
	if sequence == nil {
		panic(NewNullPointerException())
	}
	size, err := sequence.Size()
	if err != nil {
		panic(err)
	}
	return size
}
