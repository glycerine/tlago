package tlc

import (
	"fmt"
	"strings"
	"sync"
)

const lazyValueOffProperty = "tlc2.value.impl.LazyValue.off"

var lazyValueStatics struct {
	sync.Once
	off bool
}

// LazyEvalOff represents LazyValue.LAZYEVAL_OFF. The source captures the
// property once during class initialization and emits its notice from the same
// static initializer.
func LazyEvalOff() bool {
	lazyValueStatics.Do(func() {
		if value, ok := tlcLookupSystemProperty(lazyValueOffProperty); ok {
			lazyValueStatics.off = javaBooleanProperty(value)
		}
		if lazyValueStatics.off {
			ToolIOPrintln("LazyValue is disabled.")
		}
	})
	return lazyValueStatics.off
}

// InitializeLazyValueStatics represents a fresh Java classloader. It requires
// an idle runtime, like the other value-static reset helpers.
func InitializeLazyValueStatics() {
	lazyValueStatics.Once = sync.Once{}
	lazyValueStatics.off = false
}

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

	exprConCaptured bool
	exprOwner       SemanticNode
	conOwner        *Context
}

func NewLazyValue(expr SemanticNode, con *Context, cacheable bool, cms ...CostModel) *LazyValue {
	if len(cms) > 1 {
		panic(NewIllegalArgumentException("NewLazyValue accepts at most one cost model"))
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	if valueCoverageEnabled() {
		cm = cm.Get(expr)
	}
	out := &LazyValue{
		BaseValue: newBaseValue(cm), Expr: expr, Con: con,
		exprConCaptured: true, exprOwner: expr, conOwner: con,
	}
	if LazyEvalOff() || !cacheable {
		out.Val = ValUndef
	}
	return out
}

func (v *LazyValue) ownedExpr() SemanticNode {
	if v.exprConCaptured {
		return v.exprOwner
	}
	return v.Expr
}

func (v *LazyValue) ownedContext() *Context {
	if v.exprConCaptured {
		return v.conOwner
	}
	return v.Con
}

func NewLazyValueWithCostModel(expr SemanticNode, con *Context, cm CostModel) *LazyValue {
	if valueCoverageEnabled() {
		cm = cm.Get(expr)
	}
	return NewLazyValue(expr, con, true, cm)
}

func (v *LazyValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return LazyValueKind
}
func (v *LazyValue) KindString() string { return ValueKindString(v) }
func (v *LazyValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *LazyValue) IsCacheable() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.Val != ValUndef
}

func (v *LazyValue) GetNumTimesThatANewValueWasCached() int {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.CacheCount
}

func (v *LazyValue) GetCachedValue(tool *Tool, state *TLCStateMut, pstate *TLCStateMut, control int) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	control = int(int32(control))
	if v.Val == nil || !v.IsCacheable() {
		return nil
	}
	if tool == nil {
		panic(NewNullPointerException())
	}
	if tool.GetID() != v.ToolID {
		return nil
	}
	if !TLCStateIsSubset(state, v.State).IsDefinitely(true) {
		return nil
	}
	if !TLCStateIsSubset(pstate, v.PState).IsDefinitely(true) {
		return nil
	}
	if !EvalSemanticallyEquivalent(control, v.Control).IsDefinitely(true) {
		return nil
	}
	return v.Val
}

func (v *LazyValue) GetValue(tool *Tool, state *TLCStateMut, pstate *TLCStateMut, control int) (Value, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	control = int(int32(control))
	if cached := v.GetCachedValue(tool, state, pstate, control); cached != nil {
		return cached, nil
	}
	if tool == nil {
		panic(NewNullPointerException())
	}
	res, err := tool.Eval(v.ownedExpr(), v.ownedContext(), state, pstate, control, v.CM)
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
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to compare lazy values.")
	if err != nil {
		return 0, err
	}
	return val.Compare(other)
}

func (v *LazyValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to check equality of lazy values.")
	if err != nil {
		return false, err
	}
	return val.Equal(other)
}

func (v *LazyValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to check set membership of lazy values.")
	if err != nil {
		return false, err
	}
	return val.Member(elem)
}

func (v *LazyValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to check if a lazy value is a finite set.")
	if err != nil {
		return false, err
	}
	return val.IsFinite()
}

func (v *LazyValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to compute size of lazy value.")
	if err != nil {
		return 0, err
	}
	return val.Size()
}

func (v *LazyValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to normalize lazy value.")
	if err != nil {
		panic(err)
	}
	val.Normalize()
	return v
}

func (v *LazyValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
}

func (v *LazyValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to normalize lazy value.")
	if err != nil {
		panic(err)
	}
	return val.IsNormalized()
}

func (v *LazyValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}

func (v *LazyValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.Val == nil || v.Val == ValUndef {
		return v
	}
	return v.Val.DeepCopy()
}

func (v *LazyValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to fingerprint a lazy value.")
	if err != nil {
		panic(err)
	}
	return val.FingerPrint(fp)
}

func (v *LazyValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	val, err := v.ready("Error(TLC): Attempted to apply permutation to lazy value.")
	if err != nil {
		panic(err)
	}
	return val.Permute(perm)
}

func (v *LazyValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to apply EXCEPT construct to lazy value.")
	if err != nil {
		return nil, err
	}
	return val.TakeExcept(ex)
}

func (v *LazyValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	val, err := v.ready("Error(TLC): Attempted to apply EXCEPT construct to lazy value.")
	if err != nil {
		return nil, err
	}
	return val.TakeExcepts(exs)
}

func (v *LazyValue) Eval(tool *Tool, states ...*TLCStateMut) (Value, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	state := EmptyState
	var pstate *TLCStateMut
	switch len(states) {
	case 0:
	case 1:
		state = states[0]
	case 2:
		state = states[0]
		pstate = states[1]
	default:
		panic(NewWrongInvocationException("LazyValue.Eval accepts at most two states."))
	}
	if tool == nil {
		panic(NewNullPointerException())
	}
	value, err := tool.Eval(v.ownedExpr(), v.ownedContext(), state, pstate, EvalClear, v.CM)
	if err != nil {
		return nil, err
	}
	attachValueSourceIfMissing(value, v.ownedExpr())
	return value, nil
}

func (v *LazyValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *LazyValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.Val == nil || v.Val == ValUndef {
		image := "null"
		if v.ownedExpr() != nil {
			image = semanticNodeJavaString(v.ownedExpr())
		}
		sb.WriteString("<LAZY " + image + ">")
		return sb
	}
	return appendValueString(v.Val, sb, offset, swallow)
}

type LazySupplierValue struct {
	*LazyValue
	Supplier func() Value

	supplierCaptured bool
	supplierOwner    func() Value
}

func (v *LazySupplierValue) ownedSupplier() func() Value {
	if v.supplierCaptured {
		return v.supplierOwner
	}
	return v.Supplier
}

func (v *LazySupplierValue) require() *LazyValue {
	if v == nil || v.LazyValue == nil {
		panic(NewNullPointerException())
	}
	return v.LazyValue
}

func (v *LazySupplierValue) Kind() ValueKind { return v.require().Kind() }
func (v *LazySupplierValue) KindString() string {
	return v.require().KindString()
}
func (v *LazySupplierValue) GetCostModel() CostModel {
	return v.require().GetCostModel()
}
func (v *LazySupplierValue) Normalize() Value {
	lazy := v.require()
	result := lazy.Normalize()
	if result == lazy {
		return v
	}
	return result
}
func (v *LazySupplierValue) DeepNormalize() { v.require().DeepNormalize() }
func (v *LazySupplierValue) IsNormalized() bool {
	return v.require().IsNormalized()
}
func (v *LazySupplierValue) IsDefined() bool { return v.require().IsDefined() }
func (v *LazySupplierValue) DeepCopy() Value {
	lazy := v.require()
	result := lazy.DeepCopy()
	if result == lazy {
		return v
	}
	return result
}
func (v *LazySupplierValue) Permute(perm *MVPerm) Value {
	return v.require().Permute(perm)
}
func (v *LazySupplierValue) Compare(other Value) (int, error) {
	return v.require().Compare(other)
}
func (v *LazySupplierValue) Equal(other Value) (bool, error) {
	return v.require().Equal(other)
}
func (v *LazySupplierValue) Member(elem Value) (bool, error) {
	return v.require().Member(elem)
}
func (v *LazySupplierValue) IsFinite() (bool, error) { return v.require().IsFinite() }
func (v *LazySupplierValue) Size() (int, error)      { return v.require().Size() }
func (v *LazySupplierValue) FingerPrint(fp uint64) uint64 {
	return v.require().FingerPrint(fp)
}
func (v *LazySupplierValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.require().TakeExcept(ex)
}
func (v *LazySupplierValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.require().TakeExcepts(exs)
}
func (v *LazySupplierValue) String() string { return v.require().String() }
func (v *LazySupplierValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	return v.require().ToString(sb, offset, swallow)
}

func NewLazySupplierValue(expr SemanticNode, supplier func() Value) *LazySupplierValue {
	return &LazySupplierValue{
		LazyValue:        NewLazyValueWithCostModel(expr, EmptyContext, DoNotRecordCostModel),
		Supplier:         supplier,
		supplierCaptured: true,
		supplierOwner:    supplier,
	}
}

func (v *LazySupplierValue) GetValue(tool *Tool, state *TLCStateMut, pstate *TLCStateMut, control int) (Value, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	_ = tool
	_ = state
	_ = pstate
	_ = control
	if v.ownedSupplier() == nil {
		panic(NewNullPointerException())
	}
	return v.ownedSupplier()(), nil
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

	ownersCaptured bool
	varsOwner      any
	predOwner      SemanticNode
	toolOwner      *Tool
	conOwner       *Context
	stateOwner     *TLCStateMut
	pstateOwner    *TLCStateMut
	controlOwner   int
}

func (v *SetPredValue) IsSubsetEq(other Value) (*BoolValue, error) {
	return defaultEnumerableSubsetEq(v, v, other)
}

func NewSetPredValue(vars any, inVal Value, pred SemanticNode, tool *Tool, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cms ...CostModel) *SetPredValue {
	if len(cms) > 1 {
		panic(NewIllegalArgumentException("NewSetPredValue accepts at most one cost model"))
	}
	if state == nil {
		panic(NewNullPointerException())
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	stateCopy := copyTLCStateForLambda(state)
	pstateCopy := copyTLCStateForLambda(pstate)
	control = int(int32(control))
	return &SetPredValue{
		BaseValue: newBaseValue(cm),
		Vars:      vars,
		InVal:     inVal,
		Pred:      pred,
		Tool:      tool,
		Con:       con,
		State:     stateCopy,
		PState:    pstateCopy,
		Control:   control,

		ownersCaptured: true,
		varsOwner:      vars,
		predOwner:      pred,
		toolOwner:      tool,
		conOwner:       con,
		stateOwner:     stateCopy,
		pstateOwner:    pstateCopy,
		controlOwner:   control,
	}
}

func (v *SetPredValue) ownedVars() any {
	if v.ownersCaptured {
		return v.varsOwner
	}
	return v.Vars
}
func (v *SetPredValue) ownedPred() SemanticNode {
	if v.ownersCaptured {
		return v.predOwner
	}
	return v.Pred
}
func (v *SetPredValue) ownedTool() *Tool {
	if v.ownersCaptured {
		return v.toolOwner
	}
	return v.Tool
}
func (v *SetPredValue) ownedContext() *Context {
	if v.ownersCaptured {
		return v.conOwner
	}
	return v.Con
}
func (v *SetPredValue) ownedState() *TLCStateMut {
	if v.ownersCaptured {
		return v.stateOwner
	}
	return v.State
}
func (v *SetPredValue) ownedPState() *TLCStateMut {
	if v.ownersCaptured {
		return v.pstateOwner
	}
	return v.PState
}
func (v *SetPredValue) ownedControl() int {
	if v.ownersCaptured {
		return v.controlOwner
	}
	return v.Control
}

func NewSetPredValueFrom(other *SetPredValue, tool *Tool) *SetPredValue {
	if other == nil {
		panic(NewNullPointerException())
	}
	return NewSetPredValue(other.ownedVars(), other.InVal, other.ownedPred(), tool, other.ownedContext(), other.ownedState(), other.ownedPState(), other.ownedControl(), other.CM)
}

func (v *SetPredValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return SetPredValueKind
}
func (v *SetPredValue) KindString() string { return ValueKindString(v) }
func (v *SetPredValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

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
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	set, err := v.materialize()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *SetPredValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	set, err := v.materialize()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *SetPredValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	tool := v.ownedTool()
	if tool == nil {
		panic(NewNullPointerException())
	}
	res, err := tool.Eval(v.ownedPred(), ctx, v.ownedState(), v.ownedPState(), v.ownedControl())
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
		return failure != nil && !failure.Runtime && javaSystemFailureCode(failure) == NoError && (failure.Params != nil || failure.ownedCode() != ECGeneral)
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
	pred := v.ownedPred()
	if pred == nil {
		return "null"
	}
	return toContextString(pred)
}

func (v *SetPredValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	set, err := v.materialize()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (v *SetPredValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
	v.InVal.Normalize()
	return v
}

func (v *SetPredValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
	v.InVal.DeepNormalize()
}

func (v *SetPredValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.InVal == nil {
		panic(NewNullPointerException())
	}
	return v.InVal.IsNormalized()
}
func (v *SetPredValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *SetPredValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *SetPredValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.materialize()
	if err != nil {
		panic(err)
	}
	return set.FingerPrint(fp)
}

func (v *SetPredValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	set, err := v.materialize()
	if err != nil {
		panic(err)
	}
	return set.Permute(perm)
}

func (v *SetPredValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return takeExceptOnSet(v, ex)
}

func (v *SetPredValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return takeExceptsOnSet(v, exs)
}

func (v *SetPredValue) ToSetEnum() (*SetEnumValue, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	values := NewValueVec(10)
	enum := v.Elements()
	for {
		elem := enum.NextElement()
		if err := enum.Err(); err != nil {
			return nil, err
		}
		if elem == nil {
			v.CM.incValueSecondary(int64(values.Len()))
			return NewSetEnumValueVec(values, v.IsNormalized(), v.CM), nil
		}
		values.Add(elem)
	}
}

func (v *SetPredValue) Elements() ValueEnumeration {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	ctx := v.ownedContext()
	if vars, ok := v.ownedVars().(*SymbolNode); ok && vars != nil && vars.Kind == SymbolFormalParam {
		if ctx == nil {
			panic(NewNullPointerException())
		}
		return ctx.Cons(vars, elem), nil
	}
	return v.bindTuple(elem, v.tupleVars(), ctx)
}

func (v *SetPredValue) tupleVars() []*SymbolNode {
	varsOwner := v.ownedVars()
	if varsOwner == nil {
		return nil
	}
	if vars, ok := varsOwner.(*SymbolNode); ok && vars == nil {
		return nil
	}
	if vars, ok := varsOwner.([]*SymbolNode); ok {
		return vars
	}
	panic(valueStreamClassCast(varsOwner, "[Ltla2sany.semantic.FormalParamNode;"))
}

func (v *SetPredValue) bindTuple(elem Value, vars []*SymbolNode, ctx *Context) (*Context, error) {
	if elem == nil {
		panic(NewNullPointerException())
	}
	tuple := asTupleValue(elem)
	if tuple != nil && tuple.GetElems() == nil {
		panic(NewNullPointerException())
	}
	if vars == nil {
		panic(NewNullPointerException())
	}
	if tuple == nil || len(tuple.GetElems()) != len(vars) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of a set of %d-tuples.", ValuesPPR(elem), len(vars)))
	}
	values := tuple.GetElems()
	for i := 0; i < len(vars); i++ {
		if ctx == nil {
			panic(NewNullPointerException())
		}
		ctx = ctx.Cons(vars[i], fcnParameterDomain(values, i))
	}
	return ctx, nil
}

func (v *SetPredValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *SetPredValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if Globals.Expand {
		if expanded, ok := tryExpandedSetString(sb, offset, swallow, v.ToSetEnum); ok {
			return expanded
		}
	}
	sb.WriteString("{")
	if vars, ok := v.ownedVars().(*SymbolNode); ok && vars != nil && vars.Kind == SymbolFormalParam {
		sb.WriteString(vars.String())
	} else {
		vars := v.tupleVars()
		if vars == nil {
			panic(NewNullPointerException())
		}
		for i, variable := range vars {
			if variable == nil {
				panic(NewNullPointerException())
			}
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(variable.String())
		}
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
	requireValueEnumerationReceiver(e)
	if e.enum != nil {
		resetEnumeration(e.enum)
	}
	e.err = nil
}

func (e *setPredEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
	if e.err != nil {
		return nil
	}
	for {
		elem := e.enum.NextElement()
		if err := e.enum.Err(); err != nil {
			e.err = err
			return nil
		}
		if elem == nil {
			return nil
		}
		e.set.CM.incValueSecondary()
		ctx, err := e.set.bind(elem)
		if err != nil {
			e.err = err
			return nil
		}
		tool := e.set.ownedTool()
		if tool == nil {
			panic(NewNullPointerException())
		}
		res, err := tool.Eval(e.set.ownedPred(), ctx, e.set.ownedState(), e.set.ownedPState(), e.set.ownedControl(), e.set.CM)
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
