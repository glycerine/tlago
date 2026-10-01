package tlc

import (
	"math"
	"unicode/utf16"
)

const maxSeqBound = math.MaxInt32

func Seq(rangeValue Value) Value {
	return NewUserValue(&sequencesObj{Range: rangeValue, SizeBound: maxSeqBound})
}

func BSeq(rangeValue Value, size int) Value {
	return NewUserValue(&sequencesObj{Range: rangeValue, SizeBound: size})
}

func Len(s Value) (*IntValue, error) {
	if sv, ok := s.(*StringValue); ok {
		return NewIntValue(int32(sv.Length())), nil
	}
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Len", "sequence", ValuesPPR(s))
	}
	return NewIntValue(int32(len(seq.Elems))), nil
}

func Head(s Value) (Value, error) {
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Head", "sequence", ValuesPPR(s))
	}
	if len(seq.Elems) == 0 {
		return nil, newTLCErrorCode(ECTLCModuleApplyEmptySeq, "Head")
	}
	return seq.Elems[0], nil
}

func Tail(s Value) (Value, error) {
	if sv, ok := s.(*StringValue); ok {
		if sv.Val.String() == "" {
			return nil, newTLCErrorCode(ECTLCModuleApplyEmptySeq, "Tail")
		}
		return NewStringValue(utf16Substring(sv.Val.String(), 1, sv.Length())), nil
	}
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Tail", "sequence", ValuesPPR(s))
	}
	if len(seq.Elems) == 0 {
		return nil, newTLCErrorCode(ECTLCModuleApplyEmptySeq, "Tail")
	}
	out := make([]Value, len(seq.Elems)-1)
	copy(out, seq.Elems[1:])
	return NewTupleValue(out), nil
}

func Cons(v Value, s Value) (Value, error) {
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "Cons(v, s)", "sequence", ValuesPPR(s))
	}
	out := make([]Value, len(seq.Elems)+1)
	out[0] = v
	copy(out[1:], seq.Elems)
	return NewTupleValue(out), nil
}

func Append(s Value, v Value) (Value, error) {
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "Append(s, v)", "sequence", ValuesPPR(s))
	}
	out := make([]Value, len(seq.Elems)+1)
	copy(out, seq.Elems)
	out[len(seq.Elems)] = v
	return NewTupleValue(out), nil
}

func Concat(s1, s2 Value) (Value, error) {
	if sv1, ok := s1.(*StringValue); ok {
		sv2, ok := s2.(*StringValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleEvaluating, "t \\o s", "string", ValuesPPR(s2))
		}
		return NewStringValue(sv1.Val.String() + sv2.Val.String()), nil
	}
	seq1 := asTupleValue(s1)
	if seq1 == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "s \\o t", "sequence", ValuesPPR(s1))
	}
	seq2 := asTupleValue(s2)
	if seq2 == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "t \\o s", "sequence", ValuesPPR(s2))
	}
	if len(seq1.Elems) == 0 {
		return seq2, nil
	}
	if len(seq2.Elems) == 0 {
		return seq1, nil
	}
	out := make([]Value, len(seq1.Elems)+len(seq2.Elems))
	copy(out, seq1.Elems)
	copy(out[len(seq1.Elems):], seq2.Elems)
	return NewTupleValue(out), nil
}

func SubSeq(s, m, n Value) (Value, error) {
	var (
		sv       *StringValue
		seq      *TupleValue
		isString bool
	)
	if value, ok := s.(*StringValue); ok {
		sv = value
		isString = true
	} else {
		seq = asTupleValue(s)
		if seq == nil {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SubSeq", "sequence", ValuesPPR(s))
		}
	}
	begValue, ok := m.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SubSeq", "natural number", ValuesPPR(m))
	}
	endValue, ok := n.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SubSeq", "natural number", ValuesPPR(n))
	}
	beg := int(begValue.Val)
	end := int(endValue.Val)
	if isString {
		if beg > end {
			return NewStringValue(""), nil
		}
		length := sv.Length()
		if beg < 1 || beg > length {
			return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SubSeq", "first", ValuesPPR(s), ValuesPPR(m))
		}
		if end < 1 || end > length {
			return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SubSeq", "first", ValuesPPR(s), ValuesPPR(n))
		}
		return NewStringValue(utf16Substring(sv.Val.String(), beg-1, end)), nil
	}
	if beg > end {
		return EmptyTuple, nil
	}
	if beg < 1 || beg > len(seq.Elems) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SubSeq", "first", ValuesPPR(s), ValuesPPR(m))
	}
	if end < 1 || end > len(seq.Elems) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SubSeq", "first", ValuesPPR(s), ValuesPPR(n))
	}
	out := make([]Value, end-beg+1)
	copy(out, seq.Elems[beg-1:end])
	return NewTupleValue(out), nil
}

func SelectInSeq(s Value, test Value) (Value, error) {
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectInSeq", "sequence", ValuesPPR(s))
	}
	if !isOperatorValue(test) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectInSeq", "function", ValuesPPR(test))
	}
	for i, elem := range seq.Elems {
		value, err := EvalOperatorValue(test, []Value{elem}, EvalClear)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectInSeq", "boolean-valued function", ValuesPPR(test))
		}
		if boolValue.Val {
			return NewIntValue(int32(i + 1)), nil
		}
	}
	return IntZero, nil
}

func SelectSeq(s Value, test Value) (Value, error) {
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectSeq", "sequence", ValuesPPR(s))
	}
	if len(seq.Elems) == 0 {
		return EmptyTuple, nil
	}
	if !isOperatorValue(test) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectSeq", "operator", ValuesPPR(test))
	}
	out := NewValueVec(0)
	for _, elem := range seq.Elems {
		value, err := EvalOperatorValue(test, []Value{elem}, EvalClear)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectSeq", "boolean-valued operator", ValuesPPR(test))
		}
		if boolValue.Val {
			out.Add(elem)
		}
	}
	return NewTupleValue(out.ToArray()), nil
}

func Insert(s Value, v Value, test Value) (Value, error) {
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "Insert", "sequence", ValuesPPR(s))
	}
	if !isOperatorValue(test) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "Insert", "function", ValuesPPR(test))
	}
	values := make([]Value, len(seq.Elems)+1)
	idx := len(seq.Elems)
	for idx > 0 {
		right := seq.Elems[idx-1]
		value, err := EvalOperatorValue(test, []Value{v, right}, EvalClear)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "Insert", "boolean-valued operator", ValuesPPR(test))
		}
		cmp, err := v.Compare(right)
		if err != nil {
			return nil, err
		}
		if boolValue.Val && cmp < 0 {
			values[idx] = right
			idx--
		} else {
			values[idx] = v
			break
		}
	}
	if idx == 0 {
		values[0] = v
	} else {
		for i := idx - 1; i >= 0; i-- {
			values[i] = seq.Elems[i]
		}
	}
	return NewTupleValue(values), nil
}

func asTupleValue(value Value) *TupleValue {
	switch v := value.(type) {
	case *TupleValue:
		return v
	case *FcnRcdValue:
		return v.ToTuple()
	case *FcnLambdaValue:
		return v.ToTuple()
	case *RecordValue:
		return v.ToTuple()
	case *CounterExample:
		if v == nil || v.RecordValue == nil {
			return EmptyRecord.ToTuple()
		}
		return v.RecordValue.ToTuple()
	default:
		return nil
	}
}

func utf16Substring(s string, begin, end int) string {
	encoded := utf16.Encode([]rune(s))
	if begin < 0 {
		begin = 0
	}
	if end > len(encoded) {
		end = len(encoded)
	}
	if begin > end {
		begin = end
	}
	return string(utf16.Decode(encoded[begin:end]))
}

type sequencesObj struct {
	Range     Value
	SizeBound int
}

func (s *sequencesObj) Compare(value Value) (int, error) {
	uv, ok := value.(*UserValue)
	if !ok {
		if _, ok := value.(*ModelValue); ok {
			return 1, nil
		}
		return 0, newTLCErrorCode(ECTLCModuleCompareValue, ValuesPPRString(s.String()), ValuesPPR(value))
	}
	other, ok := uv.UserObj.(*sequencesObj)
	if !ok {
		return 0, newTLCErrorCode(ECTLCModuleCompareValue, ValuesPPRString(s.String()), ValuesPPR(value))
	}
	if s.SizeBound != other.SizeBound {
		return s.SizeBound - other.SizeBound, nil
	}
	return s.Range.Compare(other.Range)
}

func (s *sequencesObj) Member(value Value) (bool, error) {
	seq := asTupleValue(value)
	if seq == nil {
		if mv, ok := value.(*ModelValue); ok {
			return mv.modelValueMember(NewUserValue(s))
		}
		return false, newTLCErrorCode(ECTLCModuleCheckMemberOf, ValuesPPR(value), ValuesPPRString(s.String()))
	}
	if len(seq.Elems) > s.SizeBound {
		return false, nil
	}
	for _, elem := range seq.Elems {
		ok, err := s.Range.Member(elem)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (s *sequencesObj) IsFinite() (bool, error) {
	if s.SizeBound != maxSeqBound {
		return true, nil
	}
	finite, err := s.Range.IsFinite()
	if err != nil || !finite {
		return false, err
	}
	empty, err := IsEmptyValue(s.Range)
	if err != nil {
		return false, err
	}
	return empty, nil
}

func (s *sequencesObj) IsEmpty() (bool, error) {
	return s.SizeBound < 0, nil
}

func (s *sequencesObj) nonEnumerableErrorMsg(expr SemanticNode) string {
	rendered := ValuesPPRString(s.String())
	return "TLC encountered the non-enumerable quantifier bound\n" +
		rendered + "\n" + SemanticString(expr) + "\n" +
		"In TLA+, Seq(S) represents the set of all finite sequences whose elements come from the set S. Even when S\n" +
		"is a finite set, the number of possible sequences in Seq(S) is unbounded because sequences can have any\n" +
		"finite length (e.g., length 0, 1, 2, and so on). As a result, TLC cannot evaluate expressions that\n" +
		"universally (\\A) or existentially (\\E) quantify over " + rendered + ", because this would require checking an\n" +
		"infinite number of cases. Note that for a finite set of sequences s, TLC handles s \\subseteq Seq(S).\n" +
		"See https://explain.tlapl.us/seq-unenumerable for additional details."
}

func (s *sequencesObj) String() string {
	if s.SizeBound == maxSeqBound {
		return "Seq(" + s.Range.String() + ")"
	}
	return "BSeq(" + s.Range.String() + ", " + NewIntValue(int32(s.SizeBound)).String() + ")"
}
