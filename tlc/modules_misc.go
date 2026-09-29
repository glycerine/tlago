package tlc

import (
	"fmt"
	"io"
	"os"
	"time"
)

var (
	StringSetValue Value     = NewUserValue(stringsObj{})
	TLCOutput      io.Writer = os.Stdout
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
		return nil, newTLCError(ECGeneral, "attempted to compute Cardinality of non-enumerable value %s", value)
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
	fmt.Fprintf(TLCOutput, "%s  %s\n", v1c, v2c)
	return v2
}

func TLCPrintT(v Value) Value {
	vc := v.DeepCopy()
	vc.DeepNormalize()
	fmt.Fprintln(TLCOutput, vc)
	return BoolTrue
}

func TLCToString(v Value) Value {
	return NewStringValue(v.String())
}

func TLCAssert(condition Value, message Value) (Value, error) {
	if b, ok := condition.(*BoolValue); ok && b.Val {
		return condition, nil
	}
	return nil, newTLCError(ECGeneral, "assertion failed: %s", message)
}

func JavaTime() Value {
	seconds := time.Now().Unix()
	return NewIntValue(int32(seconds) & 0x7fffffff)
}

func MakeFcn(domain Value, elem Value) Value {
	return NewFcnRcdValue([]Value{domain}, []Value{elem}, true)
}

func CombineFcn(f1, f2 Value) (Value, error) {
	fcn1 := asFcnRcdValue(f1)
	if fcn1 == nil {
		return nil, newTLCError(ECGeneral, "first argument of @@ must be a function, got %s", f1)
	}
	fcn2 := asFcnRcdValue(f2)
	if fcn2 == nil {
		return nil, newTLCError(ECGeneral, "second argument of @@ must be a function, got %s", f2)
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
	return 0, newTLCError(ECGeneral, "attempted to compare STRING with %s", val)
}

func (stringsObj) Member(val Value) (bool, error) {
	if _, ok := val.(*StringValue); ok {
		return true, nil
	}
	if mv, ok := val.(*ModelValue); ok {
		return mv.modelValueMember(StringSetValue)
	}
	return false, newTLCError(ECGeneral, "attempted to check if %s is in STRING", val)
}

func (stringsObj) IsFinite() (bool, error) { return false, nil }
func (stringsObj) IsEmpty() (bool, error)  { return false, nil }
func (stringsObj) String() string          { return "STRING" }
