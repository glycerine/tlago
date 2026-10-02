package tlc

import "strings"

const lazyValueOffProperty = "tlc2.value.impl.LazyValue.off"

type LazyValue struct {
	BaseValue
	Expr       SemanticNode
	Con        *Context
	Val        Value
	ToolID     int64
	State      *TLCStateMut
	PState     *TLCStateMut
	Control    int
	CacheCount int
}

func NewLazyValue(expr SemanticNode, con *Context, cacheable bool, cms ...CostModel) *LazyValue {
	if con == nil {
		con = EmptyContext
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	if CoverageEnabled() {
		cm = cm.Get(expr)
	}
	out := &LazyValue{BaseValue: newBaseValue(cm), Expr: expr, Con: con}
	if lazyValueOff() || !cacheable {
		out.Val = ValUndef
	}
	return out
}

func (v *LazyValue) Kind() ValueKind    { return LazyValueKind }
func (v *LazyValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *LazyValue) IsCacheable() bool {
	return v != nil && v.Val != ValUndef
}

func lazyValueOff() bool {
	if value, ok := tlcLookupSystemProperty(lazyValueOffProperty); ok {
		return javaBooleanProperty(value)
	}
	return false
}

func (v *LazyValue) GetCachedValue(tool *Tool, state *TLCStateMut, pstate *TLCStateMut, control int) Value {
	if v == nil || v.Val == nil || !v.IsCacheable() {
		return nil
	}
	if tool.GetID() != v.ToolID {
		return nil
	}
	if !IsStateSubset(state, v.State).IsDefinitely(true) {
		return nil
	}
	if !IsStateSubset(pstate, v.PState).IsDefinitely(true) {
		return nil
	}
	if !EvalSemanticallyEquivalent(control, v.Control).IsDefinitely(true) {
		return nil
	}
	return v.Val
}

func (v *LazyValue) GetValue(tool *Tool, state *TLCStateMut, pstate *TLCStateMut, control int) (Value, error) {
	if cached := v.GetCachedValue(tool, state, pstate, control); cached != nil {
		return cached, nil
	}
	ctx := v.Con
	if ctx == nil {
		ctx = EmptyContext
	}
	res, err := tool.Eval(v.Expr, ctx, state, pstate, control, v.CM)
	if err != nil {
		return nil, err
	}
	attachValueSourceIfMissing(res, v.Expr)
	if v.IsCacheable() {
		v.Val = res
		v.ToolID = tool.GetID()
		v.State = copyTLCStateForLambda(state)
		v.PState = copyTLCStateForLambda(pstate)
		v.Control = control
		v.CacheCount++
	}
	return res, nil
}

func (v *LazyValue) ready(message string) (Value, error) {
	if v.Val == nil || v.Val == ValUndef {
		return nil, v.unsupported("%s", message)
	}
	return v.Val, nil
}

func (v *LazyValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to compare lazy values.")
	if err != nil {
		return 0, err
	}
	return val.Compare(other)
}

func (v *LazyValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to check equality of lazy values.")
	if err != nil {
		return false, err
	}
	return val.Equal(other)
}

func (v *LazyValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to check set membership of lazy values.")
	if err != nil {
		return false, err
	}
	return val.Member(elem)
}

func (v *LazyValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to check if a lazy value is a finite set.")
	if err != nil {
		return false, err
	}
	return val.IsFinite()
}

func (v *LazyValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to compute size of lazy value.")
	if err != nil {
		return 0, err
	}
	return val.Size()
}

func (v *LazyValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to normalize lazy value.")
	if err != nil {
		panic(err)
	}
	val.Normalize()
	return v
}

func (v *LazyValue) DeepNormalize() {}

func (v *LazyValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to normalize lazy value.")
	if err != nil {
		panic(err)
	}
	return val.IsNormalized()
}

func (v *LazyValue) IsDefined() bool { return true }

func (v *LazyValue) DeepCopy() Value {
	defer catchValueFailure(v, nil)
	if v.Val == nil || v.Val == ValUndef {
		return v
	}
	return v.Val.DeepCopy()
}

func (v *LazyValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to fingerprint a lazy value.")
	if err != nil {
		panic(err)
	}
	return val.FingerPrint(fp)
}

func (v *LazyValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to apply permutation to lazy value.")
	if err != nil {
		panic(err)
	}
	return val.Permute(perm)
}

func (v *LazyValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to apply EXCEPT construct to lazy value.")
	if err != nil {
		return nil, err
	}
	return val.TakeExcept(ex)
}

func (v *LazyValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to apply EXCEPT construct to lazy value.")
	if err != nil {
		return nil, err
	}
	return val.TakeExcepts(exs)
}

func (v *LazyValue) Eval(tool *Tool, state *TLCStateMut, pstate *TLCStateMut) (Value, error) {
	value, err := tool.Eval(v.Expr, v.Con, state, pstate, EvalClear, v.CM)
	if err != nil {
		return nil, err
	}
	attachValueSourceIfMissing(value, v.Expr)
	return value, nil
}

func (v *LazyValue) String() string {
	defer catchValueFailure(v, nil)
	if v.Val == nil || v.Val == ValUndef {
		return "<LAZY " + toContextString(v.Expr) + ">"
	}
	return v.Val.String()
}

type LazySupplierValue struct {
	*LazyValue
	Supplier func() Value
}

func NewLazySupplierValue(expr SemanticNode, supplier func() Value) *LazySupplierValue {
	return &LazySupplierValue{
		LazyValue: NewLazyValue(expr, EmptyContext, false),
		Supplier:  supplier,
	}
}

func (v *LazySupplierValue) GetValue(tool *Tool, state *TLCStateMut, pstate *TLCStateMut, control int) (Value, error) {
	_ = tool
	_ = state
	_ = pstate
	_ = control
	if v.Supplier == nil {
		return ValUndef, nil
	}
	value := v.Supplier()
	attachValueSourceIfMissing(value, v.Expr)
	return value, nil
}

func asLazyValue(value any) *LazyValue {
	switch v := value.(type) {
	case *LazyValue:
		return v
	case *LazySupplierValue:
		if v == nil {
			return nil
		}
		return v.LazyValue
	default:
		return nil
	}
}

func lazyValueGetValue(value any, tool *Tool, state *TLCStateMut, pstate *TLCStateMut, control int) (Value, error, bool) {
	switch v := value.(type) {
	case *LazyValue:
		res, err := v.GetValue(tool, state, pstate, control)
		return res, err, true
	case *LazySupplierValue:
		if v == nil {
			return nil, nil, false
		}
		res, err := v.GetValue(tool, state, pstate, control)
		return res, err, true
	default:
		return nil, nil, false
	}
}

type sourceAssignableValue interface {
	GetSource() SemanticNode
	SetSource(SemanticNode)
}

func attachValueSourceIfMissing(value Value, source SemanticNode) {
	if value == nil || source == nil {
		return
	}
	assignable, ok := value.(sourceAssignableValue)
	if !ok || assignable.GetSource() != nil {
		return
	}
	assignable.SetSource(source)
}

type SetPredValue struct {
	BaseValue
	Vars      any
	InVal     Value
	Pred      SemanticNode
	Tool      *Tool
	Converted bool
	Con       *Context
	State     *TLCStateMut
	PState    *TLCStateMut
	Control   int
}

func NewSetPredValue(vars any, inVal Value, pred SemanticNode, tool *Tool, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cms ...CostModel) *SetPredValue {
	if con == nil {
		con = EmptyContext
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	return &SetPredValue{
		BaseValue: newBaseValue(cm),
		Vars:      vars,
		InVal:     inVal,
		Pred:      pred,
		Tool:      tool,
		Con:       con,
		State:     copyTLCStateForLambda(state),
		PState:    copyTLCStateForLambda(pstate),
		Control:   control,
	}
}

func NewSetPredValueFrom(other *SetPredValue, tool *Tool) *SetPredValue {
	if other == nil {
		return nil
	}
	if tool == nil {
		tool = other.Tool
	}
	return NewSetPredValue(other.Vars, other.InVal, other.Pred, tool, other.Con, other.State, other.PState, other.Control, other.CM)
}

func (v *SetPredValue) Kind() ValueKind    { return SetPredValueKind }
func (v *SetPredValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *SetPredValue) materialize() (*SetEnumValue, error) {
	set, err := v.ToSetEnum()
	if err != nil {
		return nil, err
	}
	v.InVal = set
	v.Converted = true
	return set, nil
}

func (v *SetPredValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.materialize()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetPredValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.materialize()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetPredValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if v.Converted {
		return v.InVal.Member(elem)
	}
	return v.memberUnconverted(elem)
}

func (v *SetPredValue) memberUnconverted(elem Value) (resultBool bool, err error) {
	// Java's inner catch only rewrites EvalException, including failures in
	// domain membership or tuple binding; FingerprintException stays intact.
	defer func() {
		if failure := recover(); failure != nil {
			if thrown, ok := failure.(error); ok && isSetPredEvalException(thrown) {
				err = v.membershipUndecidable(elem)
			} else {
				panic(failure)
			}
		}
		if isSetPredEvalException(err) {
			err = v.membershipUndecidable(elem)
		}
	}()
	in, err := v.InVal.Member(elem)
	if err != nil || !in {
		return false, err
	}
	ctx, err := v.bind(elem)
	if err != nil {
		return false, err
	}
	// Unlike enumeration, Java member() uses the overload without a cost model.
	res, err := v.Tool.Eval(v.Pred, ctx, v.State, v.PState, v.Control)
	if err != nil {
		return false, err
	}
	boolValue, ok := res.(*BoolValue)
	if !ok {
		return false, v.unsupported("The evaluation of predicate %s yielded non-Boolean value.", v.Pred)
	}
	return boolValue.Val, nil
}

func isSetPredEvalException(err error) bool {
	switch failure := err.(type) {
	case *EvalException:
		return failure != nil
	case *TLCError:
		// Native module errors use this existing Go EvalException carrier.
		return failure != nil && (failure.Params != nil || failure.Code != ECGeneral)
	default:
		return false
	}
}

func (v *SetPredValue) membershipUndecidable(elem Value) error {
	return v.unsupported("Cannot decide if element:\n%s\n is element of:\n%s\nand satisfies the predicate %s", ValuesPPR(elem), ValuesPPR(v.InVal), v.Pred)
}

func (v *SetPredValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	finite, err := v.InVal.IsFinite()
	if err != nil {
		return false, err
	}
	if !finite {
		return false, v.unsupported("Attempted to check if expression of form {x \\in S : p(x)} is a finite set, but cannot check if S:\n%s\nis finite.", ValuesPPR(v.InVal))
	}
	return true, nil
}

func (v *SetPredValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	set, err := v.materialize()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetPredValue) Normalize() Value {
	defer catchValueFailure(v, nil)
	v.InVal.Normalize()
	return v
}

func (v *SetPredValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	v.InVal.DeepNormalize()
}

func (v *SetPredValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	return v.InVal.IsNormalized()
}
func (v *SetPredValue) IsDefined() bool { return true }
func (v *SetPredValue) DeepCopy() Value { return v }

func (v *SetPredValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	set, err := v.materialize()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetPredValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	set, err := v.materialize()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetPredValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SetPredValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SetPredValue) ToSetEnum() (*SetEnumValue, error) {
	if v.Converted {
		set, ok := v.InVal.(*SetEnumValue)
		if !ok {
			return nil, v.unsupported("converted set predicate contains non-enumerated set %s", v.InVal)
		}
		return set, nil
	}
	values := NewValueVec(0)
	enum := v.Elements()
	for {
		elem := enum.NextElement()
		if elem == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			v.CM.incValueSecondary(int64(values.Len()))
			return NewSetEnumValueVec(values, v.IsNormalized(), v.CM), nil
		}
		values.Add(elem)
	}
}

func (v *SetPredValue) Elements() ValueEnumeration {
	defer catchValueFailure(v, nil)
	if v.Converted {
		set := v.InVal.(*SetEnumValue)
		enum := set.Elements()
		if err := enum.Err(); err != nil {
			return newErrorEnumeration(wrapValueFailure(v, err))
		}
		return enum
	}
	enum, ok := asEnumerable(v.InVal)
	if !ok {
		return newErrorEnumeration(wrapValueFailure(v, v.unsupported("Attempted to enumerate { x \\in S : p(x) } when S:\n%s\nis not enumerable", ValuesPPR(v.InVal))))
	}
	elements := enum.Elements()
	if err := elements.Err(); err != nil {
		return newErrorEnumeration(wrapValueFailure(v, err))
	}
	return &setPredEnumeration{set: v, enum: elements}
}

func (v *SetPredValue) bind(elem Value) (*Context, error) {
	ctx := v.Con
	if ctx == nil {
		ctx = EmptyContext
	}
	switch vars := v.Vars.(type) {
	case *SymbolNode:
		return ctx.Cons(vars, elem), nil
	case []*SymbolNode:
		tuple := asTupleValue(elem)
		if tuple == nil || len(tuple.Elems) != len(vars) {
			return nil, v.unsupported("Attempted to check if the value:\n%s\nis an element of a set of %d-tuples.", ValuesPPR(elem), len(vars))
		}
		for i, variable := range vars {
			ctx = ctx.Cons(variable, tuple.Elems[i])
		}
		return ctx, nil
	default:
		return nil, v.unsupported("unsupported set predicate variables %T", v.Vars)
	}
}

func (v *SetPredValue) String() string {
	defer catchValueFailure(v, nil)
	if Globals.Expand {
		if expanded, ok := v.expandedString(); ok {
			return expanded
		}
	}
	var names []string
	switch vars := v.Vars.(type) {
	case *SymbolNode:
		names = []string{vars.String()}
	case []*SymbolNode:
		names = make([]string, len(vars))
		for i, variable := range vars {
			names[i] = variable.String()
		}
	}
	return "{" + strings.Join(names, ", ") + " \\in " + v.InVal.String() + " : <expression " + toContextString(v.Pred) + "> }"
}

func (v *SetPredValue) expandedString() (value string, ok bool) {
	// Java's checked toString swallows Throwable only around expansion.
	defer func() {
		if recover() != nil {
			value, ok = "", false
		}
	}()
	set, err := v.ToSetEnum()
	if err != nil {
		return "", false
	}
	return set.String(), true
}

type setPredEnumeration struct {
	set  *SetPredValue
	enum ValueEnumeration
	err  error
}

func (e *setPredEnumeration) Reset() {
	if e.enum != nil {
		e.enum.Reset()
	}
	e.err = nil
}

func (e *setPredEnumeration) NextElement() Value {
	if e.err != nil {
		return nil
	}
	for {
		elem := e.enum.NextElement()
		if elem == nil {
			e.err = e.enum.Err()
			return nil
		}
		e.set.CM.incValueSecondary()
		ctx, err := e.set.bind(elem)
		if err != nil {
			e.err = err
			return nil
		}
		res, err := e.set.Tool.Eval(e.set.Pred, ctx, e.set.State, e.set.PState, e.set.Control, e.set.CM)
		if err != nil {
			e.err = err
			return nil
		}
		boolValue, ok := res.(*BoolValue)
		if !ok {
			e.err = e.set.unsupported("Evaluating predicate %s yielded non-Boolean value.", e.set.Pred)
			return nil
		}
		if boolValue.Val {
			return elem
		}
	}
}

func (e *setPredEnumeration) Err() error { return e.err }
