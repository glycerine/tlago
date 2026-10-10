package tlc

import (
	"fmt"
	"io"
	"time"
)

var (
	StringSetValue      Value = NewUserValue(stringsObj{})
	TLCOutput           io.Writer
	TLCOutputToUserFile bool
)

func STRING() Value {
	return StringSetValue
}

func IsFiniteSet(value Value) (*BoolValue, error) {
	finite, err := value.IsFinite()
	if err != nil {
		return nil, err
	}
	return NewBoolValue(finite), nil
}

func Cardinality(value Value) (*IntValue, error) {
	if _, ok := asEnumerable(value); !ok {
		return nil, newTLCErrorCode(ECTLCModuleComputingCardinality, ValuesPPR(value))
	}
	size, err := value.Size()
	if err != nil {
		return nil, err
	}
	return NewIntValue(int32(size)), nil
}

func TLCPrint(v1, v2 Value) Value {
	v1c := v1.DeepCopy()
	v2c := v2.DeepCopy()
	v1c.DeepNormalize()
	v2c.DeepNormalize()
	text := ValuesPPRString(ValueToStringUnchecked(v1c)) + "  " + ValuesPPRString(ValueToStringUnchecked(v2c))
	if TLCOutput == nil {
		ToolIOPrintln(text)
		return v2
	}
	if _, err := fmt.Fprintln(TLCOutput, text); err != nil {
		PrintError(ECGeneral, javaGeneralErrorMessage("", err))
	}
	return v2
}

func TLCPrintT(v Value) Value {
	vc := v.DeepCopy()
	vc.DeepNormalize()
	if TLCOutput == nil {
		ToolIOPrintln(ValuesPPRString(ValueToStringUnchecked(vc)))
		return BoolTrue
	}
	_, err := fmt.Fprint(TLCOutput, ValuesPPRString(ValueToStringUnchecked(vc, "\n")))

	if err != nil {
		PrintError(ECGeneral, javaGeneralErrorMessage("", err))
	}
	return BoolTrue
}

func TLCToString(v Value) Value {
	return NewStringValue(ValueToStringUnchecked(v))
}

func TLCAssert(condition Value, message Value) (Value, error) {
	if b, ok := condition.(*BoolValue); ok && b.Val {
		return condition, nil
	}
	return nil, newTLCErrorCode(ECTLCValueAssertFailed, ValuesPPR(message))
}

func JavaTime() Value {
	seconds := time.Now().Unix()
	return NewIntValue(int32(seconds) & 0x7fffffff)
}

func MakeFcn(domain Value, elem Value) Value {
	return NewFcnRcdValue([]Value{domain}, []Value{elem}, true)
}

func CombineFcn(f1, f2 Value) (Value, error) {
	if f1 == nil {
		panic(NewNullPointerException())
	}
	fcn1 := asFcnRcdValue(f1)
	if f2 == nil {
		panic(NewNullPointerException())
	}
	// Java converts both operands before checking either conversion result.
	// Materializing the second lambda can evaluate its body or throw.
	fcn2 := asFcnRcdValue(f2)
	if fcn1 == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "@@", "function", ValuesPPR(f1))
	}
	if fcn2 == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "@@", "function", ValuesPPR(f2))
	}
	dom := NewValueVec(0)
	vals := NewValueVec(0)
	appendFunctionPairs(dom, vals, fcn1)
	len1 := dom.Len()
	dom2 := fcn2.DomainAsValues()
	for i, val := range dom2 {
		found := false
		for j := 0; j < len1; j++ {
			eq, err := val.Equal(dom.At(j))
			if err != nil {
				return nil, err
			}
			if eq {
				found = true
				break
			}
		}
		if !found {
			dom.Add(val)
			vals.Add(fcn2.Values[i])
		}
	}
	return NewFcnRcdValue(dom.ToArray(), vals.ToArray(), false), nil
}

func Permutations(value Value) (*SetEnumValue, error) {
	set, err := tryToSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "Permutations", "a finite set", ValuesPPR(value))
	}
	if _, err := set.normalizeSet(); err != nil {
		return nil, err
	}

	length := set.Elems.Len()
	if length == 0 {
		return NewSetEnumValue([]Value{EmptyFcn}, true), nil
	}

	factorial := 1
	domain := make([]Value, length)
	idxArray := make([]int, length)
	inUse := make([]bool, length)
	for i := 0; i < length; i++ {
		domain[i] = set.Elems.At(i)
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
		fcns.Add(NewFcnRcdValue(domain, vals, true))

		i := length - 1
		done := false
		for ; i >= 0; i-- {
			found := false
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
				done = true
				break
			}
			inUse[idxArray[i]] = false
		}
		if done {
			break
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
	return NewSetEnumValueVec(fcns, false), nil
}

func PermutationSubgroup(value Value) ([]*MVPerm, error) {
	enumerable, ok := asEnumerable(value)
	if !ok {
		return nil, newTLCError(ECGeneral, "symmetry operator must specify an enumerable set of functions")
	}
	if _, err := value.Size(); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	perms := make([]*MVPerm, 0)
	enum := enumerable.Elements()
	for {
		elem := enum.NextElement()
		if elem == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			break
		}
		fcn := asFcnRcdValue(elem)
		if fcn == nil {
			return nil, newTLCError(ECGeneral, "symmetry operator must specify a set of functions")
		}
		perm := NewMVPerm()
		domain := fcn.DomainAsValues()
		for i, dval := range domain {
			dmv, ok := dval.(*ModelValue)
			if !ok {
				return nil, newTLCError(ECGeneral, "symmetry function must have model values as domain and range")
			}
			rmv, ok := fcn.Values[i].(*ModelValue)
			if !ok {
				return nil, newTLCError(ECGeneral, "symmetry function must have model values as domain and range")
			}
			perm.Put(dmv, rmv)
		}
		if perm.Size() > 0 {
			key := perm.key()
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				perms = append(perms, perm)
			}
		}
	}

	generatorCount := len(perms)
	start := 0
	for {
		sizeBefore := len(perms)
		for i := 0; i < generatorCount; i++ {
			for j := start; j < sizeBefore; j++ {
				perm := perms[i].Compose(perms[j])
				if perm.Size() == 0 {
					continue
				}
				key := perm.key()
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				perms = append(perms, perm)
			}
		}
		if sizeBefore == len(perms) {
			break
		}
		start = sizeBefore
	}
	return perms, nil
}

func appendFunctionPairs(dom *ValueVec, vals *ValueVec, fcn *FcnRcdValue) {
	domain := fcn.DomainAsValues()
	for i, value := range domain {
		dom.Add(value)
		vals.Add(fcn.Values[i])
	}
}

type stringsObj struct{}

func (stringsObj) Compare(val Value) (int, error) {
	if uv, ok := val.(*UserValue); ok {
		if _, ok := uv.UserObj.(stringsObj); ok {
			return 0, nil
		}
	}
	if _, ok := val.(*ModelValue); ok {
		return 1, nil
	}
	return 0, newTLCErrorCode(ECTLCModuleCompareValue, "STRING", ValuesPPR(val))
}

func (stringsObj) Member(val Value) (bool, error) {
	if _, ok := val.(*StringValue); ok {
		return true, nil
	}
	if mv, ok := val.(*ModelValue); ok {
		return mv.modelValueMember(StringSetValue)
	}
	return false, newTLCErrorCode(ECTLCModuleCheckMemberOf, ValuesPPR(val), "STRING")
}

func (stringsObj) IsFinite() (bool, error) { return false, nil }
func (stringsObj) IsEmpty() (bool, error)  { return false, nil }
func (stringsObj) String() string          { return "STRING" }
