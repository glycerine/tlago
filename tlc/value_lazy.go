package tlc

import "strings"

type LazyValue struct {
	BaseValue
	Expr       SemanticNode
	Con        *Context
	CM         CostModel
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
	out := &LazyValue{Expr: expr, Con: con, CM: cm}
	if !cacheable {
		out.Val = ValUndef
	}
	return out
}

func (v *LazyValue) Kind() ValueKind    { return LazyValueKind }
func (v *LazyValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *LazyValue) IsCacheable() bool {
	return v != nil && v.Val != ValUndef
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

func (v *LazyValue) ready() (Value, error) {
	if v.Val == nil || v.Val == ValUndef {
		return nil, v.unsupported("error(TLC): attempted to use an unreduced lazy value")
	}
	return v.Val, nil
}

func (v *LazyValue) Compare(other Value) (int, error) {
	val, err := v.ready()
	if err != nil {
		return 0, err
	}
	return val.Compare(other)
}

func (v *LazyValue) Equal(other Value) (bool, error) {
	val, err := v.ready()
	if err != nil {
		return false, err
	}
	return val.Equal(other)
}

func (v *LazyValue) Member(elem Value) (bool, error) {
	val, err := v.ready()
	if err != nil {
		return false, err
	}
	return val.Member(elem)
}

func (v *LazyValue) IsFinite() (bool, error) {
	val, err := v.ready()
	if err != nil {
		return false, err
	}
	return val.IsFinite()
}

func (v *LazyValue) Size() (int, error) {
	val, err := v.ready()
	if err != nil {
		return 0, err
	}
	return val.Size()
}

func (v *LazyValue) Normalize() Value {
	val, err := v.ready()
	if err != nil {
		panic(err)
	}
	val.Normalize()
	return v
}

func (v *LazyValue) DeepNormalize() {
	if v.Val != nil && v.Val != ValUndef {
		v.Val.DeepNormalize()
	}
}

func (v *LazyValue) IsNormalized() bool {
	val, err := v.ready()
	if err != nil {
		panic(err)
	}
	return val.IsNormalized()
}

func (v *LazyValue) IsDefined() bool { return true }

func (v *LazyValue) DeepCopy() Value {
	if v.Val == nil || v.Val == ValUndef {
		return v
	}
	return v.Val.DeepCopy()
}

func (v *LazyValue) FingerPrint(fp uint64) uint64 {
	val, err := v.ready()
	if err != nil {
		panic(err)
	}
	return val.FingerPrint(fp)
}

func (v *LazyValue) Permute(perm *MVPerm) Value {
	val, err := v.ready()
	if err != nil {
		return v
	}
	return val.Permute(perm)
}

func (v *LazyValue) TakeExcept(ex ValueExcept) (Value, error) {
	val, err := v.ready()
	if err != nil {
		return nil, err
	}
	return val.TakeExcept(ex)
}

func (v *LazyValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	val, err := v.ready()
	if err != nil {
		return nil, err
	}
	return val.TakeExcepts(exs)
}

func (v *LazyValue) Eval(tool *Tool, state *TLCStateMut, pstate *TLCStateMut) (Value, error) {
	value, err := tool.Eval(v.Expr, v.Con, state, pstate, EvalClear)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (v *LazyValue) String() string {
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
	return v.Supplier(), nil
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

func NewSetPredValue(vars any, inVal Value, pred SemanticNode, tool *Tool, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int) *SetPredValue {
	if con == nil {
		con = EmptyContext
	}
	return &SetPredValue{
		Vars:    vars,
		InVal:   inVal,
		Pred:    pred,
		Tool:    tool,
		Con:     con,
		State:   copyTLCStateForLambda(state),
		PState:  copyTLCStateForLambda(pstate),
		Control: control,
	}
}

func NewSetPredValueFrom(other *SetPredValue, tool *Tool) *SetPredValue {
	if other == nil {
		return nil
	}
	if tool == nil {
		tool = other.Tool
	}
	return NewSetPredValue(other.Vars, other.InVal, other.Pred, tool, other.Con, other.State, other.PState, other.Control)
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

func (v *SetPredValue) Compare(other Value) (int, error) {
	set, err := v.materialize()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetPredValue) Equal(other Value) (bool, error) {
	set, err := v.materialize()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetPredValue) Member(elem Value) (bool, error) {
	if v.Converted {
		return v.InVal.Member(elem)
	}
	in, err := v.InVal.Member(elem)
	if err != nil || !in {
		return false, err
	}
	ctx, err := v.bind(elem)
	if err != nil {
		return false, err
	}
	res, err := v.Tool.Eval(v.Pred, ctx, v.State, v.PState, v.Control)
	if err != nil {
		return false, v.unsupported("cannot decide if element:\n%s\nis element of:\n%s\nand satisfies the predicate %s", elem, v.InVal, v.Pred)
	}
	boolValue, ok := res.(*BoolValue)
	if !ok {
		return false, v.unsupported("the evaluation of predicate %s yielded non-Boolean value", v.Pred)
	}
	return boolValue.Val, nil
}

func (v *SetPredValue) IsFinite() (bool, error) {
	finite, err := v.InVal.IsFinite()
	if err != nil {
		return false, err
	}
	if !finite {
		return false, v.unsupported("attempted to check if expression of form {x \\in S : p(x)} is a finite set, but cannot check if S:\n%s\nis finite", v.InVal)
	}
	return true, nil
}

func (v *SetPredValue) Size() (int, error) {
	set, err := v.materialize()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetPredValue) Normalize() Value {
	v.InVal.Normalize()
	return v
}

func (v *SetPredValue) DeepNormalize() {
	v.InVal.DeepNormalize()
}

func (v *SetPredValue) IsNormalized() bool { return v.InVal.IsNormalized() }
func (v *SetPredValue) IsDefined() bool    { return true }
func (v *SetPredValue) DeepCopy() Value    { return v }

func (v *SetPredValue) FingerPrint(fp uint64) uint64 {
	set, err := v.materialize()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetPredValue) Permute(perm *MVPerm) Value {
	set, err := v.materialize()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *SetPredValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *SetPredValue) TakeExcepts(exs []ValueExcept) (Value, error) {
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
			return NewSetEnumValueVec(values, v.IsNormalized()), nil
		}
		values.Add(elem)
	}
}

func (v *SetPredValue) Elements() ValueEnumeration {
	if v.Converted {
		if set, ok := v.InVal.(*SetEnumValue); ok {
			return set.Elements()
		}
	}
	enum, ok := asEnumerable(v.InVal)
	if !ok {
		return newErrorEnumeration(v.unsupported("attempted to enumerate { x \\in S : p(x) } when S:\n%s\nis not enumerable", v.InVal))
	}
	return &setPredEnumeration{set: v, enum: enum.Elements()}
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
			return nil, v.unsupported("attempted to check if the value:\n%s\nis an element of a set of %d-tuples", elem, len(vars))
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
	if Globals.Expand {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
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
		ctx, err := e.set.bind(elem)
		if err != nil {
			e.err = err
			return nil
		}
		res, err := e.set.Tool.Eval(e.set.Pred, ctx, e.set.State, e.set.PState, e.set.Control)
		if err != nil {
			e.err = err
			return nil
		}
		boolValue, ok := res.(*BoolValue)
		if !ok {
			e.err = e.set.unsupported("evaluating predicate %s yielded non-Boolean value", e.set.Pred)
			return nil
		}
		if boolValue.Val {
			return elem
		}
	}
}

func (e *setPredEnumeration) Err() error { return e.err }
