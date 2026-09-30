package tlc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var possibleCountsKey = UniqueStringOf("s:_possible")
var counterExampleContextSymbol = NewSymbolNode("CounterExample")
var pickSuccessorMu sync.Mutex

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

func TLCExtPickSuccessor(tool *Tool, guard Value, curState *TLCStateMut, succState *TLCStateMut) (*BoolValue, error) {
	if checker := MainChecker(); checker != nil && checker.FPSet != nil && succState != nil {
		if checker.FPSet.Contains(succState.FingerPrint()) {
			return BoolTrue, nil
		}
	}
	boolGuard, ok := guard.(*BoolValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "PickSuccessor guard must be boolean, got %s", guard)
	}
	if boolGuard.Val {
		return BoolTrue, nil
	}
	if succState == nil || succState == EmptyState || !succState.AllAssigned() {
		return BoolTrue, nil
	}
	action, err := pickSuccessorAction(tool, curState, succState)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		level := 0
		if curState != nil {
			level = curState.Level()
		}
		fmt.Fprintf(os.Stdout, "Extend behavior of length %d with a %q step [%s]? (Yes/no/explored/states/diff):\n", level, actionName(action), action)
		_ = os.Stdout.Sync()
		nextLine, err := reader.ReadString('\n')
		if err != nil && nextLine == "" {
			return nil, err
		}
		nextLine = strings.TrimRight(nextLine, "\r\n")
		if strings.TrimSpace(nextLine) == "" || strings.HasPrefix(strings.ToLower(nextLine), "y") {
			return BoolTrue, nil
		}
		switch nextLine[0] {
		case 's':
			curText := ""
			if curState != nil {
				curText = strings.TrimSpace(curState.String())
			}
			fmt.Fprintf(os.Stdout, "%s\n~>\n%s\n", curText, strings.TrimSpace(succState.String()))
		case 'd':
			if curState != nil {
				fmt.Fprint(os.Stdout, succState.StringForVariables(curState))
			} else {
				fmt.Fprint(os.Stdout, succState.String())
			}
		case 'e':
			if checker := MainChecker(); checker != nil && checker.FPSet != nil {
				checker.FPSet.Put(succState.FingerPrint())
				return BoolTrue, nil
			}
			fmt.Fprintln(os.Stdout, "Marking a state explored is unsupported by the current TLC mode. Is TLC running in simulation mode?")
		case 'n':
			return BoolFalse, nil
		}
	}
}

func pickSuccessorAction(tool *Tool, curState *TLCStateMut, succState *TLCStateMut) (*Action, error) {
	if succState != nil && succState.HasAction() {
		return succState.GetAction(), nil
	}
	if tool != nil && curState != nil && succState != nil {
		restoreCurrentState := PushCurrentState(curState)
		defer restoreCurrentState()
		for _, action := range tool.GetActions() {
			nextStates, err := tool.GetNextStates(action, curState)
			if err != nil {
				return nil, err
			}
			if nextStates.Contains(succState) {
				return action, nil
			}
		}
	}
	return UnknownAction, nil
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

func TLCExtCounterExampleWithContext(ctxt *Context) Value {
	if ctxt != nil {
		if value, ok := ctxt.Lookup(counterExampleContextSymbol).(Value); ok {
			return value
		}
	}
	return TLCExtCounterExample()
}

func TLCExtTrace(state *TLCStateMut) (Value, error) {
	return TLCExtTraceWithTool(nil, state)
}

func TLCExtTraceWithTool(tool *Tool, state *TLCStateMut) (Value, error) {
	_ = tool
	if state == nil {
		return EmptyTuple, nil
	}
	if !state.AllAssigned() {
		unassigned := state.Unassigned()
		names := make([]string, 0, len(unassigned))
		for _, variable := range unassigned {
			if variable.Name != nil {
				names = append(names, variable.Name.String())
			}
		}
		plural := ""
		if len(names) > 1 {
			plural = "s"
		}
		return nil, newTLCError(ECGeneral, "In evaluating TLCExt!Trace, the state is not completely specified yet (variable%s %s undefined).", plural, strings.Join(names, ", "))
	}
	if simulator := CurrentSimulator(); simulator != nil {
		trace := simulator.GetTrace(state)
		values := make([]Value, 0, trace.Size())
		for i := 0; i < trace.Size(); i++ {
			values = append(values, NewRecordValueFromInsMap(trace.At(i).Values()))
		}
		return NewTupleValue(values), nil
	}
	if state.IsInitial() {
		return NewTupleValue([]Value{NewRecordValueFromInsMap(state.Values())}), nil
	}
	if state.UID == TLCStateInitUID {
		current, ok := CurrentState()
		if ok && current != nil {
			trace := make([]*TLCStateInfo, 0)
			if current.IsInitial() {
				trace = append(trace, NewTLCStateInfo(current), NewTLCStateInfo(state))
			} else if checker := MainChecker(); checker != nil {
				trace = append(trace, checker.traceInfoPrefix(current)...)
				trace = appendTraceStateIfMissing(trace, current)
				trace = appendTraceStateIfMissing(trace, state)
			}
			if len(trace) > 0 {
				return traceInfoTupleValue(trace), nil
			}
		}
	}
	if checker := MainChecker(); checker != nil {
		trace := appendTraceStateIfMissing(checker.traceInfoPrefix(state), state)
		return traceInfoTupleValue(trace), nil
	}
	return predecessorTraceTupleValue(state), nil
}

func predecessorTraceTupleValue(state *TLCStateMut) Value {
	reversed := make([]Value, 0)
	for cur := state; cur != nil; cur = cur.Predecessor() {
		reversed = append(reversed, NewRecordValueFromInsMap(cur.Values()))
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return NewTupleValue(reversed)
}

func traceInfoTupleValue(trace []*TLCStateInfo) Value {
	values := make([]Value, 0, len(trace))
	for _, info := range trace {
		if info != nil && info.State != nil {
			values = append(values, NewRecordValueFromInsMap(info.State.Values()))
		}
	}
	return NewTupleValue(values)
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
	values []tlcExtCacheEntry
}

type tlcExtCacheEntry struct {
	key   Value
	value Value
}

func NewTLCExtCache() *TLCExtCache {
	return &TLCExtCache{}
}

func (c *TLCExtCache) Eval(key Value, compute func() (Value, error)) (Value, error) {
	if c == nil {
		c = NewTLCExtCache()
	}
	c.mu.RLock()
	if value, err := c.lookup(key); value != nil || err != nil {
		c.mu.RUnlock()
		return value, err
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if value, err := c.lookup(key); value != nil || err != nil {
		return value, err
	}
	if compute == nil {
		return ValUndef, nil
	}
	value, err := compute()
	if err != nil {
		return nil, err
	}
	InitializeValue(value)
	c.values = append(c.values, tlcExtCacheEntry{key: key, value: value})
	return value, nil
}

func (c *TLCExtCache) lookup(key Value) (Value, error) {
	if c == nil {
		return nil, nil
	}
	for _, entry := range c.values {
		equal, err := valuesEqualForCache(entry.key, key)
		if err != nil || equal {
			return entry.value, err
		}
	}
	return nil, nil
}

func valuesEqualForCache(left Value, right Value) (bool, error) {
	if left == nil || right == nil {
		return left == right, nil
	}
	return left.Equal(right)
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
