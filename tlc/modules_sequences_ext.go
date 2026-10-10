package tlc

import (
	"strings"
	"unicode/utf16"
)

func SequencesExtSetToSeq(value Value) (Value, error) {
	set, err := toSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	set.Normalize()
	return NewTupleValue(set.Elems.ToArray()), nil
}

func SequencesExtSetToSeqs(value Value) (Value, error) {
	set, err := toSetEnumValue(value)
	if err != nil {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "SetToSeqs", "a finite set", ValuesPPR(value))
	}
	set.Normalize()
	elems := set.Elems
	length := elems.Len()
	if length == 0 {
		return NewSetEnumValue([]Value{emptyFcnValue()}, true), nil
	}

	factorial := 1
	domain := make([]Value, length)
	idxArray := make([]int, length)
	inUse := make([]bool, length)
	for i := 0; i < length; i++ {
		domain[i] = elems.At(i)
		idxArray[i] = i
		inUse[i] = true
		factorial *= i + 1
	}

	fcns := NewValueVec(factorial)
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
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Contains", "sequence", ValuesPPR(seq))
	}
	for _, value := range tuple.Elems {
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

func SequencesExtLongestCommonPrefix(value Value) (Value, error) {
	set, err := toSetEnumValue(value)
	if err != nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "non-empty set", ValuesPPR(value))
	}
	set.Normalize()
	if set.Elems.Len() == 0 {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "non-empty set", ValuesPPR(value))
	}

	first := set.Elems.At(0)
	if str, ok := first.(*StringValue); ok {
		prefix := str.RawString()
		prefixChars := utf16.Encode([]rune(prefix))
		upper := len(prefixChars)
		for i := 1; i < set.Elems.Len(); i++ {
			other, ok := set.Elems.At(i).(*StringValue)
			if !ok {
				return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "sequence", ValuesPPR(value))
			}
			otherChars := utf16.Encode([]rune(other.RawString()))
			for idx := 0; idx < upper; idx++ {
				if idx >= len(otherChars) || prefixChars[idx] != otherChars[idx] {
					upper = idx
					if upper == 0 {
						return NewStringValue(""), nil
					}
					break
				}
			}
		}
		return NewStringValue(utf16Substring(prefix, 0, upper)), nil
	}

	prefixTuple := asTupleValue(first)
	if prefixTuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "sequence", ValuesPPR(value))
	}
	prefix := prefixTuple.Elems
	upper := len(prefix)
	for i := 1; i < set.Elems.Len(); i++ {
		otherTuple := asTupleValue(set.Elems.At(i))
		if otherTuple == nil {
			return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "LongestCommonPrefix", "sequence", ValuesPPR(value))
		}
		other := otherTuple.Elems
		for idx := 0; idx < upper; idx++ {
			if idx >= len(other) {
				upper = idx
				break
			}
			eq, err := prefix[idx].Equal(other[idx])
			if err != nil {
				return nil, err
			}
			if !eq {
				upper = idx
				if upper == 0 {
					return EmptyTuple, nil
				}
				break
			}
		}
	}
	if upper == 0 {
		return EmptyTuple, nil
	}
	return NewTupleValue(prefix[:upper]), nil
}

func SequencesExtFoldSeq(op Value, base Value, seq Value) (Value, error) {
	return FunctionsFoldFunction(op, base, seq)
}

func SequencesExtFoldLeft(op Value, base Value, seq Value) (Value, error) {
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldLeft", "sequence", ValuesPPR(seq))
	}
	args := []Value{base, nil}
	for _, elem := range tuple.Elems {
		args[1] = elem
		value, err := EvalOperatorValue(op, args, EvalClear)
		if err != nil {
			return nil, err
		}
		args[0] = value
	}
	return args[0], nil
}

func SequencesExtFoldRight(op Value, seq Value, base Value) (Value, error) {
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "FoldRight", "sequence", ValuesPPR(seq))
	}
	args := []Value{nil, base}
	for i := len(tuple.Elems) - 1; i >= 0; i-- {
		args[0] = tuple.Elems[i]
		value, err := EvalOperatorValue(op, args, EvalClear)
		if err != nil {
			return nil, err
		}
		args[1] = value
	}
	return args[1], nil
}

func SequencesExtFoldLeftDomain(op Value, base Value, seq Value) (Value, error) {
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldLeftDomain", "sequence", ValuesPPR(seq))
	}
	args := []Value{base, nil}
	for i := range tuple.Elems {
		args[1] = NewIntValue(int32(i + 1))
		value, err := EvalOperatorValue(op, args, EvalClear)
		if err != nil {
			return nil, err
		}
		args[0] = value
	}
	return args[0], nil
}

func SequencesExtFoldRightDomain(op Value, seq Value, base Value) (Value, error) {
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "FoldRightDomain", "sequence", ValuesPPR(seq))
	}
	args := []Value{nil, base}
	for i := len(tuple.Elems) - 1; i >= 0; i-- {
		args[0] = NewIntValue(int32(i + 1))
		value, err := EvalOperatorValue(op, args, EvalClear)
		if err != nil {
			return nil, err
		}
		args[1] = value
	}
	return args[1], nil
}

func SequencesExtReplaceFirstSubSeq(replacement Value, subseq Value, target Value) (Value, error) {
	if r, ok := replacement.(*StringValue); ok {
		s, sok := subseq.(*StringValue)
		t, tok := target.(*StringValue)
		if sok && tok {
			if s.RawString() == "" {
				return NewStringValue(r.RawString() + t.RawString()), nil
			}
			return NewStringValue(stringsReplaceOnce(t.RawString(), s.RawString(), r.RawString())), nil
		}
	}
	rTuple := asTupleValue(replacement)
	sTuple := asTupleValue(subseq)
	tTuple := asTupleValue(target)
	if rTuple == nil || sTuple == nil || tTuple == nil {
		return target, nil
	}
	idx, err := sequencesExtIndexFirstSubSeq(sTuple.Elems, tTuple.Elems)
	if err != nil || idx < 0 {
		return target, err
	}
	return NewTupleValue(sequencesExtReplaceAt(idx, rTuple.Elems, sTuple.Elems, tTuple.Elems)), nil
}

func SequencesExtReplaceAllSubSeqs(replacement Value, subseq Value, target Value) (Value, error) {
	if r, ok := replacement.(*StringValue); ok {
		s, sok := subseq.(*StringValue)
		t, tok := target.(*StringValue)
		if sok && tok {
			if s.RawString() == "" {
				var b strings.Builder
				b.WriteString(r.RawString())
				for i, ch := range t.RawString() {
					if i != 0 {
						b.WriteString(r.RawString())
					}
					b.WriteRune(ch)
				}
				return NewStringValue(b.String()), nil
			}
			return NewStringValue(strings.ReplaceAll(t.RawString(), s.RawString(), r.RawString())), nil
		}
	}
	rTuple := asTupleValue(replacement)
	sTuple := asTupleValue(subseq)
	tTuple := asTupleValue(target)
	if rTuple == nil || sTuple == nil || tTuple == nil {
		return target, nil
	}
	if sequencesExtTupleEqual(sTuple.Elems, tTuple.Elems) {
		return replacement, nil
	}
	if sequencesExtTupleEqual(rTuple.Elems, sTuple.Elems) {
		return target, nil
	}
	if len(sTuple.Elems) == 0 {
		out := make([]Value, 0, len(rTuple.Elems)*len(tTuple.Elems)+len(tTuple.Elems))
		for _, elem := range tTuple.Elems {
			out = append(out, rTuple.Elems...)
			out = append(out, elem)
		}
		if len(tTuple.Elems) == 0 {
			out = append(out, rTuple.Elems...)
		}
		return NewTupleValue(out), nil
	}
	out := make([]Value, 0, len(tTuple.Elems))
	for i := 0; i < len(tTuple.Elems); {
		matches, err := sequencesExtSubSeqAt(tTuple.Elems, sTuple.Elems, i)
		if err != nil {
			return nil, err
		}
		if matches {
			out = append(out, rTuple.Elems...)
			i += len(sTuple.Elems)
		} else {
			out = append(out, tTuple.Elems[i])
			i++
		}
	}
	return NewTupleValue(out), nil
}

func SequencesExtIsPrefix(left Value, right Value) (Value, error) {
	if s1, ok := left.(*StringValue); ok {
		s2, ok := right.(*StringValue)
		if ok {
			return NewBoolValue(strings.HasPrefix(s2.RawString(), s1.RawString())), nil
		}
	}
	s := asTupleValue(left)
	if s == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "IsPrefix", "sequence", ValuesPPR(left))
	}
	t := asTupleValue(right)
	if t == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "IsPrefix", "sequence", ValuesPPR(right))
	}
	if len(s.Elems) > len(t.Elems) {
		return BoolFalse, nil
	}
	for i, elem := range s.Elems {
		eq, err := elem.Equal(t.Elems[i])
		if err != nil {
			return nil, err
		}
		if !eq {
			return BoolFalse, nil
		}
	}
	return BoolTrue, nil
}

func SequencesExtSelectInSeq(seq Value, test Value) (Value, error) {
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectInSeq", "sequence", ValuesPPR(seq))
	}
	for i, elem := range tuple.Elems {
		value, err := EvalOperatorValue(test, []Value{elem}, EvalClear)
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
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectInSubSeq", "sequence", ValuesPPR(seq))
	}
	fromInt, ok := from.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectInSubSeq", "natural", ValuesPPR(from))
	}
	toInt, ok := to.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SelectInSubSeq", "natural", ValuesPPR(to))
	}
	start := int(fromInt.Val)
	end := int(toInt.Val)
	if start > end {
		return IntZero, nil
	}
	if start < 1 || start > len(tuple.Elems) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SelectInSubSeq", "first", ValuesPPR(seq), ValuesPPR(from))
	}
	if end < 1 || end > len(tuple.Elems) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SelectInSubSeq", "first", ValuesPPR(seq), ValuesPPR(to))
	}
	for i := start; i <= end; i++ {
		value, err := EvalOperatorValue(test, []Value{tuple.Elems[i-1]}, EvalClear)
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
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectLastInSeq", "sequence", ValuesPPR(seq))
	}
	for i := len(tuple.Elems) - 1; i >= 0; i-- {
		value, err := EvalOperatorValue(test, []Value{tuple.Elems[i]}, EvalClear)
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
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectLastInSubSeq", "sequence", ValuesPPR(seq))
	}
	fromInt, ok := from.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectLastInSubSeq", "natural", ValuesPPR(from))
	}
	toInt, ok := to.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SelectLastInSubSeq", "natural", ValuesPPR(to))
	}
	start := int(fromInt.Val)
	end := int(toInt.Val)
	if start > end {
		return IntZero, nil
	}
	if start < 1 || start > len(tuple.Elems) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SelectLastInSubSeq", "first", ValuesPPR(seq), ValuesPPR(from))
	}
	if end < 1 || end > len(tuple.Elems) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SelectLastInSubSeq", "first", ValuesPPR(seq), ValuesPPR(to))
	}
	for i := end; i >= start; i-- {
		value, err := EvalOperatorValue(test, []Value{tuple.Elems[i-1]}, EvalClear)
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
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "RemoveFirst", "sequence", ValuesPPR(seq))
	}
	out := make([]Value, 0, len(tuple.Elems))
	found := false
	for _, value := range tuple.Elems {
		if !found {
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
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "RemoveFirstMatch", "sequence", ValuesPPR(seq))
	}
	out := make([]Value, 0, len(tuple.Elems))
	found := false
	for _, elem := range tuple.Elems {
		if !found {
			value, err := EvalOperatorValue(test, []Value{elem}, EvalClear)
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
		out = append(out, elem)
	}
	return NewTupleValue(out), nil
}

func SequencesExtSuffixes(seq Value) (Value, error) {
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Suffixes", "sequence", ValuesPPR(seq))
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
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "AllSubSeqs", "sequence", ValuesPPR(seq))
	}
	n := len(tuple.Elems)
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

func stringsReplaceOnce(s string, old string, new string) string {
	idx := strings.Index(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

func sequencesExtIndexFirstSubSeq(subseq []Value, target []Value) (int, error) {
	if len(subseq) == 0 {
		return 0, nil
	}
	for i := 0; i+len(subseq) <= len(target); i++ {
		ok, err := sequencesExtSubSeqAt(target, subseq, i)
		if err != nil {
			return -1, err
		}
		if ok {
			return i, nil
		}
	}
	return -1, nil
}

func sequencesExtSubSeqAt(target []Value, subseq []Value, start int) (bool, error) {
	if start+len(subseq) > len(target) {
		return false, nil
	}
	for i, elem := range subseq {
		eq, err := elem.Equal(target[start+i])
		if err != nil {
			return false, err
		}
		if !eq {
			return false, nil
		}
	}
	return true, nil
}

func sequencesExtReplaceAt(index int, replacement []Value, subseq []Value, target []Value) []Value {
	out := make([]Value, 0, len(target)-len(subseq)+len(replacement))
	out = append(out, target[:index]...)
	out = append(out, replacement...)
	out = append(out, target[index+len(subseq):]...)
	return out
}

func sequencesExtTupleEqual(left []Value, right []Value) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		eq, err := left[i].Equal(right[i])
		if err != nil || !eq {
			return false
		}
	}
	return true
}
