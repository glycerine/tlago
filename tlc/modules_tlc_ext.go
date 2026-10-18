package tlc

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var possibleCountsKey *UniqueString
var tlcExtActionField *UniqueString

// Refresh class-static extension names with the current worker interner.
var tlcExtClassInterning classInterning

func initTLCExtUniqueStrings() {
	possibleCountsKey = UniqueStringOf("s:_possible")
	tlcExtActionField = UniqueStringOf("_action")
}

// The synchronized static methods and Trace reconstruction share TLCExt.class.
// TLCCache retains its separate source read/write lock.
var tlcExtClassMonitor distributedServerMonitor
var tlcExtCacheLock reentrantReadWriteLock

func TLCExtAssertError(expected *StringValue, eval func() (Value, error)) (*BoolValue, error) {
	ensureTLCExtConsole()
	tlcExtClassMonitor.Lock()
	defer tlcExtClassMonitor.Unlock()
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
	if !isJavaEvalOrRuntimeException(err) {
		return nil, err
	}
	message := javaThrowableDetailMessage(err)
	if message != nil && *message == expected.RawString() {
		return BoolTrue, nil
	}
	return BoolFalse, nil
}

// tlcExtPickSuccessorSeen is the pre-evaluation phase of the source override.
// An I/O failure prints its stack and accepts the state, skipping the guard.
func tlcExtPickSuccessorSeen(succState *TLCStateMut) (seen bool) {
	defer func() {
		if failure := recover(); failure != nil {
			err := panicValueAsError(failure)
			if !isJavaIOException(err) {
				panic(failure)
			}
			_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
			seen = true
		}
	}()
	checker := MainChecker()
	if checker == nil {
		return false
	}
	// Java captures the contains receiver before evaluating its argument,
	// then dereferences it only after fingerprinting has completed.
	set := requireBFSChecker(checker).FPSet
	if succState == nil {
		panic(NewNullPointerException())
	}
	fp := succState.FingerPrint()
	if set == nil {
		panic(NewNullPointerException())
	}
	return set.Contains(fp)
}

func TLCExtPickSuccessor(tool *Tool, guard Value, curState *TLCStateMut, succState *TLCStateMut) (*BoolValue, error) {
	ensureTLCExtConsole()
	tlcExtClassMonitor.Lock()
	defer tlcExtClassMonitor.Unlock()
	if tlcExtPickSuccessorSeen(succState) {
		return BoolTrue, nil
	}
	return tlcExtPickSuccessorGuard(tool, guard, curState, succState)
}

// The caller owns the class monitor and has already checked fingerprint history.
func tlcExtPickSuccessorGuard(tool *Tool, guard Value, curState *TLCStateMut, succState *TLCStateMut) (*BoolValue, error) {
	boolGuard, ok := guard.(*BoolValue)
	if !ok {
		return nil, NewTLCRuntimeExceptionMessage(fmt.Sprintf("PickSuccessor guard must be boolean, got %s", guard))
	}
	if boolGuard.Val {
		return BoolTrue, nil
	}
	if succState == nil {
		panic(NewNullPointerException())
	}
	if succState == EmptyState || !succState.AllAssigned() {
		return BoolTrue, nil
	}
	action, err := pickSuccessorAction(tool, curState, succState)
	if err != nil {
		return nil, err
	}
	reader := ensureTLCExtConsole()
	for {
		if curState == nil || action == nil {
			panic(NewNullPointerException())
		}
		level := curState.Level()
		PrintMessage(ECTLCModuleOverrideStdout, "Extend behavior of length "+strconv.Itoa(level)+" with a \""+actionName(action)+"\" step ["+action.String()+"]? (Yes/no/explored/states/diff):")
		nextLine := reader.nextLine()
		if tlcExtConsoleTrim(nextLine) == "" || strings.HasPrefix(strings.ToLower(nextLine), "y") {
			return BoolTrue, nil
		}
		switch nextLine[0] {
		case 's':
			curText := ""
			if curState != nil {
				curText = tlcExtConsoleTrim(curState.String())
			}
			PrintMessage(ECTLCModuleOverrideStdout, curText+"\n~>\n"+tlcExtConsoleTrim(succState.String()))
		case 'd':
			if curState != nil {
				PrintMessage(ECTLCModuleOverrideStdout, succState.StringForVariables(curState))
			} else {
				PrintMessage(ECTLCModuleOverrideStdout, succState.String())
			}
		case 'e':
			if checker := MainChecker(); checker != nil {
				tlcExtPickSuccessorExplore(requireBFSChecker(checker), succState)
				return BoolTrue, nil
			}
			PrintMessage(ECTLCModuleOverrideStdout, "Marking a state explored is unsupported by the current TLC mode. Is TLC running in simulation mode?")
		case 'n':
			return BoolFalse, nil
		}
	}
}

func tlcExtPickSuccessorExplore(checker *ModelChecker, succState *TLCStateMut) {
	defer func() {
		if failure := recover(); failure != nil {
			err := panicValueAsError(failure)
			if !isJavaIOException(err) {
				panic(failure)
			}
			_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
		}
	}()
	set := requireBFSChecker(checker).FPSet
	fp := succState.FingerPrint()
	if set == nil {
		panic(NewNullPointerException())
	}
	set.Put(fp)
}

func pickSuccessorAction(tool *Tool, curState *TLCStateMut, succState *TLCStateMut) (*Action, error) {
	if succState != nil && succState.retainsExtendedMetadata() {
		return succState.GetAction(), nil
	}
	if tool == nil {
		panic(NewNullPointerException())
	}
	for _, action := range tool.GetActions() {
		nextStates, err := tool.GetNextStates(action, curState)
		if err != nil {
			return nil, err
		}
		if nextStates.Contains(succState) {
			return action, nil
		}
	}
	return nil, nil
}

func TLCExtToTrace(value Value) (Value, error) {
	ensureTLCExtConsole()
	counterExample, ok := value.(*CounterExample)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "ToTrace", "CounterExample", ValuesPPR(value))
	}
	return counterExample.ToTrace(), nil
}

func TLCExtCounterExample() Value {
	ensureTLCExtConsole()
	return NewEmptyCounterExample()
}

func TLCExtCounterExampleWithTool(tool *Tool, ctxt *Context) Value {
	ensureTLCExtConsole()
	def := tool.GetCounterExampleDef()
	if def != nil && def.Symbol != nil && ctxt != nil {
		if value, ok := ctxt.Lookup(def.Symbol).(Value); ok {
			return value
		}
	}
	return TLCExtCounterExample()
}

func TLCExtTrace(state *TLCStateMut) (Value, error) {
	ensureTLCExtConsole()
	return TLCExtTraceWithTool(nil, state)
}

func TLCExtTraceWithTool(tool *Tool, state *TLCStateMut) (Value, error) {
	ensureTLCExtConsole()
	_ = tool
	if state == nil {
		panic(NewNullPointerException())
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
		message := fmt.Sprintf("In evaluating TLCExt!Trace, the state is not completely specified yet (variable%s %s undefined).", plural, strings.Join(names, ", "))
		panic(NewTLCRuntimeException(ECGeneral, message))
	}
	if simulator := CurrentSimulator(); simulator != nil {
		trace := simulator.GetTrace(state)
		values := make([]Value, 0, trace.Size())
		for i := 0; i < trace.Size(); i++ {
			traceState := trace.At(i)
			values = append(values, stateActionRecordValue(traceState, traceState.GetAction()))
		}
		return NewTupleValue(values), nil
	}
	if state.IsInitial() {
		return NewTupleValue([]Value{NewRecordValueFromState(state)}), nil
	}
	tlcExtClassMonitor.Lock()
	defer tlcExtClassMonitor.Unlock()
	if state.UID == TLCStateInitUID {
		current, _ := CurrentState()
		if current == nil {
			panic(NewNullPointerException())
		}
		trace := make([]*TLCStateInfo, 0)
		if current.IsInitial() {
			trace = append(trace, NewTLCStateInfo(current), NewTLCStateInfo(state))
		} else {
			trace = append(trace, MainChecker().traceInfoPrefix(current)...)
			trace = append(trace, NewTLCStateInfo(current), NewTLCStateInfo(state))
			// Source restores only after successful prefix reconstruction;
			// nested Tool.getState calls replace the current-state slot.
			SetCurrentState(current)
		}
		return traceInfoTupleValue(trace), nil
	}
	return traceInfoTupleValue(MainChecker().traceInfoPrefix(state), state), nil
}

func stateActionRecordValue(state *TLCStateMut, action *Action) *RecordValue {
	return NewRecordValueFromStateAndAction(state, action)
}

func traceInfoTupleValue(trace []*TLCStateInfo, suffix ...*TLCStateMut) Value {
	values := make([]Value, 0, len(trace)+len(suffix))
	for _, info := range trace {
		if info == nil {
			panic(NewNullPointerException())
		}
		values = append(values, NewRecordValueFromState(info.ownedState()))
	}
	for _, state := range suffix {
		values = append(values, NewRecordValueFromState(state))
	}
	return NewTupleValue(values)
}

func TLCExtTLCDefer(states []*TLCStateMut, callable func() (any, error)) (Value, error) {
	ensureTLCExtConsole()
	for _, state := range states {
		if state == nil {
			// The source catches the setup failure and calls Assert.fail. The
			// underlying null-pointer exception has no native detail message.
			failure := newTLCErrorCodeNullable(ECTLCModuleValueJavaMethodOverride, javaString("TLCDefer"), nil)
			failure.Runtime = true
			return nil, failure
		}
		state.SetCallable(callable)
	}
	return BoolTrue, nil
}

func TLCExtTLCNoOp(value Value) Value {
	ensureTLCExtConsole()
	return value
}

func TLCExtTLCModelValue(value Value) (Value, error) {
	ensureTLCExtConsole()
	tlcExtClassMonitor.Lock()
	defer tlcExtClassMonitor.Unlock()
	str, ok := asStringValue(value)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "ModelValue", "string", ValuesPPR(value))
	}
	return AddModelValue(str.RawString()), nil
}

func TLCExtTLCFP(value Value) *IntValue {
	ensureTLCExtConsole()
	tlcExtClassMonitor.Lock()
	defer tlcExtClassMonitor.Unlock()
	value.DeepNormalize()
	return NewIntValue(FP64Hash(value.FingerPrint(FP64New())))
}

func TLCExtTLCEvalDefinition(tool *Tool, name Value, args ...any) (Value, error) {
	ensureTLCExtConsole()
	str, ok := asStringValue(name)
	if !ok {
		if name == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "TLCEvalDefinition", "string", ValuesPPR(name))
	}
	if tool == nil || tool.ownedSpecProcessor() == nil || tool.ownedSpecProcessor().ModuleTbl == nil {
		panic(NewNullPointerException())
	}
	module := tool.ownedSpecProcessor().ModuleTbl.GetRootModule()
	if module == nil {
		panic(NewAssertionError())
	}
	if str == nil || str.ownedVal() == nil {
		panic(NewNullPointerException())
	}
	// Configured runtime bindings can replace entries in DefnsByName. Java
	// resolves the original semantic definition in the root module instead.
	opDef := module.GetOpDef(str.ownedVal())
	if opDef == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "TLCEvalDefinition", "name of a definition reachable from the root module", ValuesPPR(name))
	}
	if opDef.Arity() != 0 {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "TLCEvalDefinition", "a zero-arity definition", opDef.GetSignature())
	}
	con, s0, s1, control, cm := parseEvalArgs(args...)
	return tool.Eval(opDef.GetBody(), con, s0, s1, control, cm)
}

// TLCCache stores this HashMap on the expression node. Synchronization belongs
// to the Java override's class-wide reentrant lock, including nested calls.
type TLCExtCache struct {
	values *javaHashMap[Value, Value]
}

func NewTLCExtCache() *TLCExtCache {
	return &TLCExtCache{values: newTLCValueHashMap[Value]()}
}

func newTLCValueHashMap[V any]() *javaHashMap[Value, V] {
	return configureTLCValueHashMap(newJavaHashMap[Value, V](ValueJavaHashCode, nil))
}

func newTLCValueHashMapWithCapacity[V any](initialCapacity int) *javaHashMap[Value, V] {
	return configureTLCValueHashMap(newJavaHashMapWithCapacity[Value, V](initialCapacity, ValueJavaHashCode, nil))
}

func configureTLCValueHashMap[V any](values *javaHashMap[Value, V]) *javaHashMap[Value, V] {
	values.equal = func(key, stored Value) bool {
		if key == nil {
			return false
		}
		equal, err := key.Equal(stored)
		if err != nil {
			panic(err)
		}
		return equal
	}
	values.tieBreak = func(a, b Value) int {
		if a != nil && b != nil {
			if order := dotCompareString(reflect.TypeOf(a).Elem().Name(), reflect.TypeOf(b).Elem().Name()); order != 0 {
				return order
			}
		}
		var ah, bh uint32
		if a != nil {
			ah = uint32(reflect.ValueOf(a).Pointer()) & 0x7fffffff
		}
		if b != nil {
			bh = uint32(reflect.ValueOf(b).Pointer()) & 0x7fffffff
		}
		if ah <= bh {
			return -1
		}
		return 1
	}
	return values
}

func semanticTLCExtCache(tool *Tool, expr SemanticNode) *TLCExtCache {
	object := SemanticToolObjectForTool(tool, expr)
	if isNil(object) {
		return nil
	}
	cache, ok := object.(*TLCExtCache)
	if !ok {
		panic(valueStreamClassCast(object, "java.util.HashMap"))
	}
	return cache
}

func tlcExtCacheForTool(tool *Tool, expr SemanticNode) *TLCExtCache {
	tlcExtCacheLock.Lock()
	defer tlcExtCacheLock.Unlock()
	cache := semanticTLCExtCache(tool, expr)
	if cache == nil {
		cache = NewTLCExtCache()
		SetSemanticToolObjectForTool(tool, expr, cache)
	}
	return cache
}

func (c *TLCExtCache) Eval(key Value, compute func() (Value, error)) (Value, error) {
	if c == nil {
		c = NewTLCExtCache()
	}
	return evalTLCExtCache(c, nil, nil, key, compute)
}

func evalTLCExtCache(cache *TLCExtCache, tool *Tool, expr SemanticNode, key Value, compute func() (Value, error)) (Value, error) {
	tlcExtCacheLock.RLock()
	readHeld, writeHeld := true, false
	defer func() {
		if readHeld {
			tlcExtCacheLock.RUnlock()
		}
		if writeHeld {
			tlcExtCacheLock.Unlock()
		}
	}()
	if cache == nil {
		cache = semanticTLCExtCache(tool, expr)
	}
	if cache != nil {
		if value, _ := cache.values.Get2(key); value != nil {
			return value, nil
		}
	}
	tlcExtCacheLock.RUnlock()
	readHeld = false
	tlcExtCacheLock.Lock()
	writeHeld = true
	if cache == nil {
		cache = semanticTLCExtCache(tool, expr)
	}
	if cache == nil {
		cache = NewTLCExtCache()
		SetSemanticToolObjectForTool(tool, expr, cache)
	}
	if value, _ := cache.values.Get2(key); value != nil {
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
	cache.values.Set(key, value)
	return value, nil
}

func PossibleCounts() Value {
	ensureTLCExtUniqueStrings()
	values := []Value(nil)
	if checker := MainChecker(); checker != nil {
		values = append(values, checker.GetAllNamedValues(possibleCountsKey)...)
	} else if simulator := CurrentSimulator(); simulator != nil {
		values = append(values, simulator.GetAllNamedValues(possibleCountsKey)...)
	}
	totals := newTLCValueHashMap[int32]()
	for _, value := range values {
		fcn, ok := value.(*FcnRcdValue)
		if !ok {
			continue
		}
		if fcn == nil {
			panic(NewNullPointerException())
		}
		fcn.Normalize()
		if fcn.ownedIntv() != nil {
			// The source iterates rec.domain directly, which is null for an
			// interval-backed function even after normalization.
			panic(NewNullPointerException())
		}
		for i, dval := range fcn.ownedDomain() {
			count, ok := fcn.ownedValues()[i].(*IntValue)
			if !ok {
				if isNil(fcn.ownedValues()[i]) {
					panic(NewNullPointerException())
				}
				panic(valueStreamClassCast(fcn.ownedValues()[i], "tlc2.value.impl.IntValue"))
			}
			if count == nil {
				panic(NewNullPointerException())
			}
			totals.mergeNonNull(dval, count.ownedVal(), func(a, b int32) int32 { return a + b })
		}
	}
	domain := make([]Value, 0, totals.Len())
	counts := make([]Value, 0, totals.Len())
	for key, total := range totals.All() {
		domain = append(domain, key)
		counts = append(counts, NewIntValue(total))
	}
	return NewFcnRcdValue(domain, counts, false)
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

func TLCTraceSerialize(value Value, absolutePath *StringValue) (result Value, err error) {
	if absolutePath == nil {
		return nil, NewNullPointerException(`Cannot read field "val" because "absolutePath" is null`)
	}
	if absolutePath.ownedVal() == nil {
		return nil, NewNullPointerException(`Cannot invoke "util.UniqueString.toString()" because "absolutePath.val" is null`)
	}
	out, err := NewValueOutputStreamFile(absolutePath.RawString(), true)
	if err != nil {
		return nil, err
	}
	defer func() {
		failure := recover()
		closeErr := out.Close()
		if closeErr != nil {
			if failure != nil {
				panic(closeErr)
			}
			result, err = nil, closeErr
			return
		}
		if failure != nil {
			panic(failure)
		}
	}()

	if isNil(value) {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.IValue.fingerPrint(long)" because "value" is null`)
	}
	value.FingerPrint(0)
	if err := ValueWrite(value, out); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func TLCTraceDeserialize(absolutePath *StringValue) (result Value, err error) {
	if absolutePath == nil {
		return nil, NewNullPointerException(`Cannot read field "val" because "absolutePath" is null`)
	}
	if absolutePath.ownedVal() == nil {
		return nil, NewNullPointerException(`Cannot invoke "util.UniqueString.toString()" because "absolutePath.val" is null`)
	}
	name := filenameNormalizeFile(absolutePath.RawString())
	if strings.IndexByte(name, 0) >= 0 {
		return EmptyRecord, nil
	}
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		return EmptyRecord, nil
	} else if err != nil {
		return nil, err
	}
	in, err := NewValueInputStreamFile(name, true)
	if errors.Is(err, os.ErrNotExist) {
		return EmptyRecord, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		failure := recover()
		closeErr := in.Close()
		if closeErr != nil {
			if failure != nil {
				panic(closeErr)
			}
			result, err = nil, closeErr
			return
		}
		if failure != nil {
			panic(failure)
		}
	}()
	// The compatibility overload delegates to ReadExternal: the trace may come
	// from a different process with a different UniqueString order.
	return in.ReadWithTable(internTable.ToMap())
}

func TLCTraceState(state *TLCStateMut) Value {
	return NewRecordValueFromState(state)
}

func JsonTraceState(state *TLCStateMut) Value {
	return TLCTraceState(state)
}

func ensureTLCExtUniqueStrings() { tlcExtClassInterning.ensure(initTLCExtUniqueStrings) }
