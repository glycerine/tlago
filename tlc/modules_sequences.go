package tlc

import (
	"math"
)

const maxSeqBound = math.MaxInt32

func Seq(rangeValue Value) Value {
	return NewUserValue(&sequencesObj{Range: rangeValue, SizeBound: maxSeqBound})
}

func BSeq(rangeValue Value, size int) Value {
	return NewUserValue(&sequencesObj{Range: rangeValue, SizeBound: size})
}

func Len(s Value) (*IntValue, error) {
	if sv, ok := asStringValue(s); ok {
		return NewIntValue(int32(sv.Length())), nil
	}
	seq := sequenceTuple(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Len", "sequence", ValuesPPR(s))
	}
	size, err := seq.Size()
	if err != nil {
		return nil, err
	}
	return NewIntValue(int32(size)), nil
}

func Head(s Value) (Value, error) {
	seq := sequenceTuple(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Head", "sequence", ValuesPPR(s))
	}
	size, err := seq.Size()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, newTLCErrorCode(ECTLCModuleApplyEmptySeq, "Head")
	}
	return seq.Elems[0], nil
}

func Tail(s Value) (Value, error) {
	if sv, ok := asStringValue(s); ok {
		if sv.Val == nil {
			panic(NewNullPointerException())
		}
		if sv.Val.String() == "" {
			return nil, newTLCErrorCode(ECTLCModuleApplyEmptySeq, "Tail")
		}
		return NewStringValue(utf16Substring(sv.Val.String(), 1, sv.Length())), nil
	}
	seq := sequenceTuple(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Tail", "sequence", ValuesPPR(s))
	}
	size, err := seq.Size()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, newTLCErrorCode(ECTLCModuleApplyEmptySeq, "Tail")
	}
	out := make([]Value, size-1)
	copy(out, seq.Elems[1:])
	return NewTupleValue(out), nil
}

func Cons(v Value, s Value) (Value, error) {
	seq := sequenceTuple(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "Cons(v, s)", "sequence", ValuesPPR(s))
	}
	size, err := seq.Size()
	if err != nil {
		return nil, err
	}
	out := make([]Value, size+1)
	out[0] = v
	copy(out[1:], seq.Elems)
	return NewTupleValue(out), nil
}

func Append(s Value, v Value) (Value, error) {
	seq := sequenceTuple(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "Append(s, v)", "sequence", ValuesPPR(s))
	}
	size, err := seq.Size()
	if err != nil {
		return nil, err
	}
	out := make([]Value, size+1)
	copy(out, seq.Elems)
	out[size] = v
	return NewTupleValue(out), nil
}

func Concat(s1, s2 Value) (Value, error) {
	if sv1, ok := asStringValue(s1); ok {
		sv2, ok := asStringValue(s2)
		if !ok {
			if s2 == nil {
				panic(NewNullPointerException())
			}
			return nil, newTLCErrorCode(ECTLCModuleEvaluating, "t \\o s", "string", ValuesPPR(s2))
		}
		if sv1.Val == nil || sv2.Val == nil {
			panic(NewNullPointerException())
		}
		return NewStringValue(javaStringConcat(sv1.Val.String(), sv2.Val.String())), nil
	}
	seq1 := sequenceTuple(s1)
	if seq1 == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "s \\o t", "sequence", ValuesPPR(s1))
	}
	seq2 := sequenceTuple(s2)
	if seq2 == nil {
		return nil, newTLCErrorCode(ECTLCModuleEvaluating, "t \\o s", "sequence", ValuesPPR(s2))
	}
	len1, err := seq1.Size()
	if err != nil {
		return nil, err
	}
	len2, err := seq2.Size()
	if err != nil {
		return nil, err
	}
	if len1 == 0 {
		return seq2, nil
	}
	if len2 == 0 {
		return seq1, nil
	}
	out := make([]Value, len1+len2)
	copy(out, seq1.Elems)
	copy(out[len1:], seq2.Elems)
	return NewTupleValue(out), nil
}

func SubSeq(s, m, n Value) (Value, error) {
	var (
		str      string
		seq      *TupleValue
		isString bool
	)
	if value, ok := asStringValue(s); ok {
		if value.Val == nil {
			panic(NewNullPointerException())
		}
		str = value.Val.String()
		isString = true
	} else {
		seq = sequenceTuple(s)
		if seq == nil {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SubSeq", "sequence", ValuesPPR(s))
		}
	}
	begValue, ok := m.(*IntValue)
	if !ok {
		if m == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SubSeq", "natural number", ValuesPPR(m))
	}
	endValue, ok := n.(*IntValue)
	if !ok {
		if n == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "SubSeq", "natural number", ValuesPPR(n))
	}
	beg := int(begValue.Val)
	end := int(endValue.Val)
	if isString {
		if beg > end {
			return NewStringValue(""), nil
		}
		length := len(javaStringUTF16(str))
		if beg < 1 || beg > length {
			return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SubSeq", "first", ValuesPPR(s), ValuesPPR(m))
		}
		if end < 1 || end > length {
			return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SubSeq", "first", ValuesPPR(s), ValuesPPR(n))
		}
		return NewStringValue(utf16Substring(str, beg-1, end)), nil
	}
	if beg > end {
		return EmptyTuple, nil
	}
	size, err := seq.Size()
	if err != nil {
		return nil, err
	}
	if beg < 1 || beg > size {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "second", "SubSeq", "first", ValuesPPR(s), ValuesPPR(m))
	}
	if end < 1 || end > size {
		return nil, newTLCErrorCode(ECTLCModuleArgumentNotInDomain, "third", "SubSeq", "first", ValuesPPR(s), ValuesPPR(n))
	}
	out := make([]Value, end-beg+1)
	copy(out, seq.Elems[beg-1:end])
	return NewTupleValue(out), nil
}

// Sequences invokes value.toTuple() directly, before checking conversion and
// reading size. Keep that null boundary separate from the general conversion.
func sequenceTuple(value Value) *TupleValue {
	if value == nil {
		panic(NewNullPointerException())
	}
	return asTupleValue(value)
}

func SelectInSeq(s Value, test Value) (Value, error) {
	seq := asTupleValue(s)
	if seq == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SelectInSeq", "sequence", ValuesPPR(s))
	}
	if !isFunctionValue(test) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SelectInSeq", "function", ValuesPPR(test))
	}
	for i, elem := range seq.Elems {
		value, err, _ := applyFunctionValue(test, []Value{elem}, EvalClear)
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
	if !isFunctionValue(test) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "Insert", "function", ValuesPPR(test))
	}
	values := make([]Value, len(seq.Elems)+1)
	idx := len(seq.Elems)
	for idx > 0 {
		right := seq.Elems[idx-1]
		value, err, _ := applyFunctionValue(test, []Value{v, right}, EvalClear)
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

func isFunctionValue(value Value) bool {
	switch value.(type) {
	case *TupleValue, *RecordValue, *CounterExample, *FcnRcdValue, *FcnLambdaValue:
		return true
	default:
		return false
	}
}

func applyFunctionValue(value Value, args []Value, control int) (Value, error, bool) {
	switch v := value.(type) {
	case *TupleValue:
		out, err := v.ApplyArgs(args, control)
		return out, err, true
	case *RecordValue:
		out, err := v.ApplyArgs(args, control)
		return out, err, true
	case *CounterExample:
		if v == nil || v.RecordValue == nil {
			out, err := EmptyRecord.ApplyArgs(args, control)
			return out, err, true
		}
		out, err := v.RecordValue.ApplyArgs(args, control)
		return out, err, true
	case *FcnRcdValue:
		out, err := v.Apply(functionApplyArg(args))
		return out, err, true
	case *FcnLambdaValue:
		out, err := v.ApplyArgs(args, control)
		return out, err, true
	default:
		return nil, nil, false
	}
}

func functionApplyArg(args []Value) Value {
	if len(args) == 1 {
		return args[0]
	}
	return NewTupleValue(args)
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
	encoded := javaStringUTF16(s)
	if begin < 0 {
		begin = 0
	}
	if end > len(encoded) {
		end = len(encoded)
	}
	if begin > end {
		begin = end
	}
	return javaStringFromUTF16(encoded[begin:end])
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
