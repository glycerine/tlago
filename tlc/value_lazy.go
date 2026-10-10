package tlc

import (
	"fmt"
	"strings"
)

const lazyValueOffProperty = "tlc2.value.impl.LazyValue.off"

type LazyValue struct {
	BaseValue
	Expr       SemanticNode
	Con        *Context
	Val        Value
	ToolID     int32
	State      *TLCStateMut
	PState     *TLCStateMut
	Control    int
	CacheCount int
}

func NewLazyValue(expr SemanticNode, con *Context, cacheable bool, cms ...CostModel) *LazyValue {
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
	if tool == nil {
		panic(NewNullPointerException())
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
	if tool == nil {
		panic(NewNullPointerException())
	}
	res, err := tool.Eval(v.Expr, v.Con, state, pstate, control, v.CM)
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

func (v *LazyValue) ready(message string) (Value, error) {
	if v.Val == nil || v.Val == ValUndef {
		return nil, v.runtimeFailure(message)
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

func (v *LazyValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to apply EXCEPT construct to lazy value.")
	if err != nil {
		return nil, err
	}
	return val.TakeExcept(ex)
}

func (v *LazyValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
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
	return ValueToString(v, "", true)
}

func (v *LazyValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if v.Val == nil || v.Val == ValUndef {
		sb.WriteString("<LAZY " + semanticNodeJavaString(v.Expr) + ">")
		return sb
	}
	return appendValueString(v.Val, sb, offset, swallow)
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
		panic(NewNullPointerException())
	}
	return v.Supplier(), nil
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
	if set == nil {
		// Each caller stores the null conversion before invoking its next method.
		v.InVal = nil
		v.Converted = true
		panic(NewNullPointerException())
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
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
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
			if thrown, ok := failure.(error); ok && isValueEvalException(thrown) {
				err = v.membershipUndecidable(elem)
			} else {
				panic(failure)
			}
		}
		if isValueEvalException(err) {
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
	if v.Tool == nil {
		panic(NewNullPointerException())
	}
	res, err := v.Tool.Eval(v.Pred, ctx, v.State, v.PState, v.Control)
	if err != nil {
		return false, err
	}
	boolValue, ok := res.(*BoolValue)
	if !ok {
		return false, v.runtimeFailure("The evaluation of predicate " + v.predicateImage() + " yielded non-Boolean value.")
	}
	return boolValue.Val, nil
}

func isValueEvalException(err error) bool {
	switch failure := err.(type) {
	case *EvalException:
		return failure != nil
	case *TLCError:
		// Native module errors use this existing Go EvalException carrier.
		return failure != nil && !failure.Runtime && javaSystemFailureCode(failure) == NoError && (failure.Params != nil || failure.Code != ECGeneral)
	default:
		return false
	}
}

func (v *SetPredValue) membershipUndecidable(elem Value) error {
	if elem == nil {
		panic(NewNullPointerException())
	}
	elementText := ValuesPPR(elem)
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
	return v.runtimeFailure("Cannot decide if element:\n" + elementText + "\n is element of:\n" + ValuesPPR(v.InVal) + "\nand satisfies the predicate " + v.predicateImage())
}

func (v *SetPredValue) predicateImage() string {
	if v.Pred == nil {
		return "null"
	}
	return toContextString(v.Pred)
}

func (v *SetPredValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
	finite, err := v.InVal.IsFinite()
	if err != nil {
		return false, err
	}
	if !finite {
		return false, v.runtimeFailure("Attempted to check if expression of form {x \\in S : p(x)} is a finite set, but cannot check if S:\n" + ValuesPPR(v.InVal) + "\nis finite.")
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
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
	v.InVal.Normalize()
	return v
}

func (v *SetPredValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
	v.InVal.DeepNormalize()
}

func (v *SetPredValue) IsNormalized() bool {
	defer catchValueFailure(v, nil)
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
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

func (v *SetPredValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SetPredValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SetPredValue) ToSetEnum() (*SetEnumValue, error) {
	if v.Converted {
		if v.InVal == nil {
			return nil, nil
		}
		set, ok := v.InVal.(*SetEnumValue)
		if !ok {
			panic(valueStreamClassCast(v.InVal, "tlc2.value.impl.SetEnumValue"))
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
		if v.InVal == nil {
			panic(NewNullPointerException())
		}
		set, ok := v.InVal.(*SetEnumValue)
		if !ok {
			panic(valueStreamClassCast(v.InVal, "tlc2.value.impl.SetEnumValue"))
		}
		enum := set.Elements()
		if err := enum.Err(); err != nil {
			return newErrorEnumeration(wrapValueFailure(v, err))
		}
		return enum
	}
	enum, ok := asEnumerable(v.InVal)
	if !ok {
		if v.InVal == nil {
			panic(NewNullPointerException())
		}
		return newErrorEnumeration(wrapValueFailure(v, v.runtimeFailure("Attempted to enumerate { x \\in S : p(x) } when S:\n"+ValuesPPR(v.InVal)+"\nis not enumerable")))
	}
	elements := enum.Elements()
	if err := elements.Err(); err != nil {
		return newErrorEnumeration(wrapValueFailure(v, err))
	}
	return &setPredEnumeration{set: v, enum: elements}
}

func (v *SetPredValue) bind(elem Value) (*Context, error) {
	ctx := v.Con
	switch vars := v.Vars.(type) {
	case *SymbolNode:
		if vars == nil {
			return v.bindTuple(elem, nil, ctx)
		}
		if ctx == nil {
			panic(NewNullPointerException())
		}
		return ctx.Cons(vars, elem), nil
	case []*SymbolNode:
		return v.bindTuple(elem, vars, ctx)
	case nil:
		return v.bindTuple(elem, nil, ctx)
	default:
		return nil, v.unsupported("unsupported set predicate variables %T", v.Vars)
	}
}

func (v *SetPredValue) bindTuple(elem Value, vars []*SymbolNode, ctx *Context) (*Context, error) {
	if elem == nil {
		panic(NewNullPointerException())
	}
	tuple := asTupleValue(elem)
	if tuple != nil && tuple.Elems == nil {
		panic(NewNullPointerException())
	}
	if vars == nil {
		panic(NewNullPointerException())
	}
	if tuple == nil || len(tuple.Elems) != len(vars) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of a set of %d-tuples.", ValuesPPR(elem), len(vars)))
	}
	values := tuple.Elems
	for i := 0; i < len(vars); i++ {
		if ctx == nil {
			panic(NewNullPointerException())
		}
		ctx = ctx.Cons(vars[i], fcnParameterDomain(values, i))
	}
	return ctx, nil
}

func (v *SetPredValue) String() string {
	return ValueToString(v, "", true)
}

func (v *SetPredValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if Globals.Expand {
		if expanded, ok := tryExpandedSetString(sb, offset, swallow, v.ToSetEnum); ok {
			return expanded
		}
	}
	sb.WriteString("{")
	switch vars := v.Vars.(type) {
	case *SymbolNode:
		sb.WriteString(vars.String())
	case []*SymbolNode:
		for i, variable := range vars {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(variable.String())
		}
	default:
		if v.Vars == nil {
			panic(NewNullPointerException())
		}
		// Java casts every non-scalar vars object to FormalParamNode[].
		_ = v.Vars.([]*SymbolNode)
	}
	// Java concatenates inVal here, entering its public checked string path.
	inText := "null"
	if v.InVal != nil {
		inText = v.InVal.String()
	}
	sb.WriteString(" \\in " + inText + " : <expression ")
	sb.WriteString(v.predicateImage() + "> }")
	return sb
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
		if e.set.Tool == nil {
			panic(NewNullPointerException())
		}
		res, err := e.set.Tool.Eval(e.set.Pred, ctx, e.set.State, e.set.PState, e.set.Control, e.set.CM)
		if err != nil {
			e.err = err
			return nil
		}
		boolValue, ok := res.(*BoolValue)
		if !ok {
			e.err = e.set.runtimeFailure("Evaluating predicate " + e.set.predicateImage() + " yielded non-Boolean value.")
			return nil
		}
		if boolValue.Val {
			return elem
		}
	}
}

func (e *setPredEnumeration) Err() error { return e.err }
