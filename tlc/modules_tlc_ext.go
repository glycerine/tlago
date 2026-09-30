package tlc

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var possibleCountsKey = UniqueStringOf("s:_possible")

func TLCExtAssertError(expected *StringValue, eval func() (Value, error)) (*BoolValue, error) {
	if expected == nil {
		return nil, newTLCError(ECGeneral, "AssertError expected a string error")
	}
	if eval == nil {
		return BoolFalse, nil
	}
	_, err := eval()
	if err == nil {
		return BoolFalse, nil
	}
	if err.Error() == expected.UnquotedString() {
		return BoolTrue, nil
	}
	return BoolFalse, nil
}

func TLCExtPickSuccessor(guard Value, curState *TLCStateMut, succState *TLCStateMut) (*BoolValue, error) {
	boolGuard, ok := guard.(*BoolValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "PickSuccessor guard must be boolean, got %s", guard)
	}
	if boolGuard.Val {
		return BoolTrue, nil
	}
	if succState == nil || !succState.AllAssigned() {
		return BoolTrue, nil
	}
	_ = curState
	return BoolTrue, nil
}

func TLCExtToTrace(value Value) (Value, error) {
	counterExample, ok := value.(*CounterExample)
	if !ok {
		return nil, newTLCError(ECGeneral, "ToTrace expected a CounterExample, got %s", value)
	}
	return counterExample.ToTrace(), nil
}

func TLCExtCounterExample() Value {
	return NewEmptyCounterExample()
}

func TLCExtTrace(state *TLCStateMut) (Value, error) {
	if state == nil {
		return EmptyTuple, nil
	}
	if !state.AllAssigned() {
		return nil, newTLCError(ECGeneral, "In evaluating TLCExt!Trace, the state is not completely specified yet")
	}
	reversed := make([]Value, 0)
	for cur := state; cur != nil; cur = cur.Predecessor() {
		reversed = append(reversed, NewRecordValueFromInsMap(cur.Values()))
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return NewTupleValue(reversed), nil
}

func TLCExtTLCDefer(states []*TLCStateMut, callable func() (any, error)) Value {
	for _, state := range states {
		if state != nil {
			state.SetCallable(callable)
		}
	}
	return BoolTrue
}

func TLCExtTLCNoOp(value Value) Value {
	return value
}

func TLCExtTLCModelValue(value Value) (Value, error) {
	str, ok := value.(*StringValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "ModelValue expected a string, got %s", value)
	}
	return AddModelValue(str.UnquotedString()), nil
}

func TLCExtTLCFP(value Value) *IntValue {
	value.DeepNormalize()
	return NewIntValue(FP64Hash(value.FingerPrint(FP64New())))
}

func TLCExtTLCEvalDefinition(tool *Tool, name Value, args ...any) (Value, error) {
	if tool == nil {
		return nil, newTLCError(ECGeneral, "TLCEvalDefinition has no tool")
	}
	str, ok := name.(*StringValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "TLCEvalDefinition expected a string definition name, got %s", name)
	}
	value := tool.DefnsByName[str.Val]
	opDef, ok := value.(*OpDefNode)
	if !ok || opDef == nil {
		return nil, newTLCError(ECGeneral, "TLCEvalDefinition could not find zero-arity definition %q", str.UnquotedString())
	}
	if opDef.Arity() != 0 {
		return nil, newTLCError(ECGeneral, "TLCEvalDefinition expected a zero-arity definition, got %s", opDef)
	}
	con, s0, s1, control, cm := parseEvalArgs(args...)
	return tool.Eval(opDef.Body, con, s0, s1, control, cm)
}

type TLCExtCache struct {
	mu     sync.RWMutex
	values *InsMap[string, Value]
}

func NewTLCExtCache() *TLCExtCache {
	return &TLCExtCache{values: NewInsMap[string, Value]()}
}

func (c *TLCExtCache) Eval(key Value, compute func() (Value, error)) (Value, error) {
	if c == nil {
		c = NewTLCExtCache()
	}
	cacheKey := key.String()
	c.mu.RLock()
	if value := c.values.Get(cacheKey); value != nil {
		c.mu.RUnlock()
		return value, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if value := c.values.Get(cacheKey); value != nil {
		return value, nil
	}
	if compute == nil {
		return ValUndef, nil
	}
	value, err := compute()
	if err != nil {
		return nil, err
	}
	InitializeValue(value)
	c.values.Set(cacheKey, value)
	return value, nil
}

func PossibleCounts() Value {
	values := []Value(nil)
	if checker := MainChecker(); checker != nil {
		values = append(values, checker.GetAllNamedValues(possibleCountsKey)...)
	}
	if simulator := CurrentSimulator(); simulator != nil {
		values = append(values, simulator.GetAllNamedValues(possibleCountsKey)...)
	}
	domain := NewValueVec(0)
	counts := NewValueVec(0)
	for _, value := range values {
		fcn := asFcnRcdValue(value)
		if fcn == nil {
			continue
		}
		fcn.Normalize()
		fcnDomain := fcn.DomainAsValues()
		for i, dval := range fcnDomain {
			count, ok := fcn.Values[i].(*IntValue)
			if !ok {
				continue
			}
			idx, err := findEqualValue(domain, dval)
			if err != nil {
				panic(err)
			}
			if idx < 0 {
				domain.Add(dval)
				counts.Add(count)
			} else {
				prev := counts.At(idx).(*IntValue)
				counts.Set(idx, NewIntValue(prev.Val+count.Val))
			}
		}
	}
	return NewFcnRcdValue(domain.ToArray(), counts.ToArray(), false)
}

func findEqualValue(values *ValueVec, target Value) (int, error) {
	for i := 0; i < values.Len(); i++ {
		eq, err := values.At(i).Equal(target)
		if err != nil || eq {
			return i, err
		}
	}
	return -1, nil
}

func TLCTraceSerialize(value Value, absolutePath *StringValue) (Value, error) {
	if absolutePath == nil {
		return nil, newTLCError(ECGeneral, "_TLCTraceSerialize expected a string path")
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath.UnquotedString()), 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}
	file, err := os.Create(absolutePath.UnquotedString())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	value.FingerPrint(0)
	out := NewValueOutputStream(file)
	if err := out.WriteExternal(value); err != nil {
		return nil, err
	}
	if err := out.Close(); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func TLCTraceDeserialize(absolutePath *StringValue) (Value, error) {
	if absolutePath == nil {
		return nil, newTLCError(ECGeneral, "_TLCTraceDeserialize expected a string path")
	}
	file, err := os.Open(absolutePath.UnquotedString())
	if errors.Is(err, os.ErrNotExist) {
		return EmptyRecord, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return NewValueInputStream(file).ReadExternal()
}

func TLCTraceState(state *TLCStateMut) Value {
	if state == nil {
		return EmptyRecord
	}
	return NewRecordValueFromInsMap(state.Values())
}

func JsonTraceState(state *TLCStateMut) Value {
	return TLCTraceState(state)
}
