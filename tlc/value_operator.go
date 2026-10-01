package tlc

import (
	"strings"
)

type OpDefNode struct {
	SemanticNodeBase
	Symbol      *SymbolNode
	Name        *UniqueString
	Params      []*SymbolNode
	Body        SemanticNode
	InRecursive bool
}

func NewOpDefNode(name string, params []*SymbolNode, body SemanticNode) *OpDefNode {
	return NewOpDefNodeForSymbol(NewSymbolNode(name), params, body)
}

func NewOpDefNodeForSymbol(symbol *SymbolNode, params []*SymbolNode, body SemanticNode) *OpDefNode {
	outParams := make([]*SymbolNode, len(params))
	copy(outParams, params)
	if symbol == nil {
		symbol = NewSymbolNode("")
	}
	name := symbol.Name
	image := ""
	if name != nil {
		image = name.String()
	}
	return &OpDefNode{
		SemanticNodeBase: newSemanticNodeBase(SemanticUserDefinedOpKind, image),
		Symbol:           symbol,
		Name:             name,
		Params:           outParams,
		Body:             body,
	}
}

func (n *OpDefNode) Arity() int {
	if n == nil {
		return 0
	}
	return len(n.Params)
}

func (n *OpDefNode) SetInRecursive(value bool) {
	if n != nil {
		n.InRecursive = value
	}
}

func (n *OpDefNode) GetInRecursive() bool {
	return n != nil && n.InRecursive
}

func (n *OpDefNode) String() string {
	if n == nil || n.Name == nil {
		return "<anonymous>"
	}
	return n.Name.String()
}

func (n *OpDefNode) Signature() string {
	if n == nil || n.Name == nil {
		return "<anonymous>"
	}
	if len(n.Params) == 0 {
		return n.Name.String()
	}
	parts := make([]string, 0, len(n.Params))
	for _, param := range n.Params {
		if param == nil || param.Name == nil {
			continue
		}
		parts = append(parts, param.Name.String())
	}
	return n.Name.String() + "(" + strings.Join(parts, ", ") + ")"
}

type OperatorEvalFunc func(args []Value, control int) (Value, error)
type EvaluatingEvalFunc func(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error)
type CallableEvalFunc func(args []Value) (func() (any, error), error)

type operatorValueBase struct {
	BaseValue
	KindValue        ValueKind
	Label            string
	NormalizeMessage string
}

func (b *operatorValueBase) Kind() ValueKind { return b.KindValue }
func (b *operatorValueBase) KindString() string {
	return b.KindStringFor(b.KindValue)
}

func (b *operatorValueBase) Compare(other Value) (int, error) {
	return 0, b.unsupported("Attempted to compare operator %s with value:\n%s", b, other)
}

func (b *operatorValueBase) Equal(other Value) (bool, error) {
	return false, b.unsupported("Attempted to check equality of operator %s with value:\n%s", b, other)
}

func (b *operatorValueBase) Member(elem Value) (bool, error) {
	return false, b.unsupported("Attempted to check if the value:\n%s\nis an element of operator %s", elem, b)
}

func (b *operatorValueBase) IsFinite() (bool, error) {
	return false, b.unsupported("Attempted to check if the operator %s is a finite set.", b)
}

func (b *operatorValueBase) Size() (int, error) {
	return 0, b.unsupported("Attempted to compute the number of elements in the operator %s.", b)
}

func (b *operatorValueBase) Normalize() Value {
	panic(newTLCError(ECGeneral, "%s", b.normalizeMessage()))
}

func (b *operatorValueBase) DeepNormalize() {}

func (b *operatorValueBase) IsNormalized() bool {
	panic(newTLCError(ECGeneral, "%s", b.normalizeMessage()))
}

func (b *operatorValueBase) IsDefined() bool { return true }
func (b *operatorValueBase) DeepCopy() Value { return b }

func (b *operatorValueBase) FingerPrint(fp uint64) uint64 {
	return unsupportedValueFingerprint(b)
}

func (b *operatorValueBase) Permute(*MVPerm) Value {
	unsupportedValueFingerprint(b)
	return b
}

func (b *operatorValueBase) TakeExcept(ex ValueExcept) (Value, error) {
	return nil, b.unsupported("Attempted to appy EXCEPT construct to the operator %s.", b)
}

func (b *operatorValueBase) TakeExcepts(exs []ValueExcept) (Value, error) {
	return nil, b.unsupported("Attempted to apply EXCEPT construct to the operator %s.", b)
}

func (b *operatorValueBase) String() string { return b.Label }

func (b *operatorValueBase) normalizeMessage() string {
	if b.NormalizeMessage != "" {
		return b.NormalizeMessage
	}
	return "It is a TLC bug: Attempted to normalize an operator."
}

type OpLambdaValue struct {
	operatorValueBase
	OpDef   *OpDefNode
	Tool    *Tool
	Con     *Context
	State   *TLCStateMut
	PState  *TLCStateMut
	Control int
	CM      CostModel
}

func NewOpLambdaValue(opDef *OpDefNode, tool *Tool, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cms ...CostModel) *OpLambdaValue {
	if con == nil {
		con = EmptyContext
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	return &OpLambdaValue{
		operatorValueBase: operatorValueBase{KindValue: OpLambdaValueKind, Label: "<Operator " + opDef.String() + ">", NormalizeMessage: "Should not normalize an operator."},
		OpDef:             opDef,
		Tool:              tool,
		Con:               con,
		State:             state,
		PState:            pstate,
		Control:           control,
		CM:                cm,
	}
}

func NewOpLambdaValueFrom(other *OpLambdaValue, tool *Tool) *OpLambdaValue {
	if other == nil {
		return nil
	}
	if tool == nil {
		tool = other.Tool
	}
	return NewOpLambdaValue(other.OpDef, tool, other.Con, other.State, other.PState, other.Control)
}

func (v *OpLambdaValue) DeepCopy() Value { return v }
func (v *OpLambdaValue) Permute(perm *MVPerm) Value {
	return v.operatorValueBase.Permute(perm)
}

func (v *OpLambdaValue) Eval(args []Value, control int) (Value, error) {
	if v.OpDef == nil {
		return nil, newTLCError(ECGeneral, "attempted to apply a nil operator")
	}
	if v.OpDef.Arity() != len(args) {
		return nil, v.unsupported("applying the operator %s with wrong number of arguments", v)
	}
	ctx := v.Con
	if ctx == nil {
		ctx = EmptyContext
	}
	for i, param := range v.OpDef.Params {
		ctx = ctx.Cons(param, args[i])
	}
	if v.Tool == nil {
		return ValUndef, nil
	}
	if EvalIsEnabled(v.Control) {
		control = EvalSetEnabled(control)
	}
	return v.Tool.Eval(v.OpDef.Body, ctx, v.State, v.PState, control, v.CM)
}

type OpRcdValue struct {
	operatorValueBase
	Domain [][]Value
	Values []Value
}

func NewOpRcdValue() *OpRcdValue {
	return &OpRcdValue{
		operatorValueBase: operatorValueBase{KindValue: OpRcdValueKind, Label: "<Operator record>", NormalizeMessage: "Should not normalize an operator."},
	}
}

func NewOpRcdValueFrom(domain [][]Value, values []Value) *OpRcdValue {
	outDomain := make([][]Value, len(domain))
	for i, args := range domain {
		outDomain[i] = make([]Value, len(args))
		copy(outDomain[i], args)
	}
	outValues := make([]Value, len(values))
	copy(outValues, values)
	return &OpRcdValue{
		operatorValueBase: operatorValueBase{KindValue: OpRcdValueKind, Label: "<Operator record>", NormalizeMessage: "Should not normalize an operator."},
		Domain:            outDomain,
		Values:            outValues,
	}
}

func (v *OpRcdValue) AddLine(values []Value) {
	if len(values) < 2 {
		return
	}
	args := make([]Value, len(values)-2)
	copy(args, values[1:len(values)-1])
	v.Domain = append(v.Domain, args)
	v.Values = append(v.Values, values[len(values)-1])
}

func (v *OpRcdValue) Eval(args []Value, control int) (Value, error) {
	_ = control
	for i, vals := range v.Domain {
		if len(args) != len(vals) {
			return nil, v.unsupported("attempted to apply the operator %s\nwith wrong number of arguments", v)
		}
		matched := true
		for j := range vals {
			eq, err := vals[j].Equal(args[j])
			if err != nil {
				return nil, err
			}
			if !eq {
				matched = false
				break
			}
		}
		if matched {
			return v.Values[i], nil
		}
	}
	return nil, v.unsupported("attempted to apply operator:\n%s\nto arguments (%s), which is undefined", v, joinValueStrings(args, ", "))
}

func (v *OpRcdValue) IsDefined() bool {
	for _, value := range v.Values {
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *OpRcdValue) DeepNormalize() {
	for i := range v.Domain {
		for _, arg := range v.Domain[i] {
			arg.DeepNormalize()
		}
	}
	for _, value := range v.Values {
		value.DeepNormalize()
	}
}

func (v *OpRcdValue) DeepCopy() Value { return v }
func (v *OpRcdValue) Permute(perm *MVPerm) Value {
	return v.operatorValueBase.Permute(perm)
}

func (v *OpRcdValue) String() string {
	var b strings.Builder
	b.WriteString("{ ")
	for i, value := range v.Values {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("<")
		for _, arg := range v.Domain[i] {
			b.WriteString(arg.String())
			b.WriteString(", ")
		}
		b.WriteString(value.String())
		b.WriteString(">")
	}
	b.WriteString("}")
	return b.String()
}

type MethodValue struct {
	operatorValueBase
	Name     string
	MinLevel int
	EvalFunc OperatorEvalFunc
}

func NewMethodValue(name string, minLevel int, eval OperatorEvalFunc) *MethodValue {
	return &MethodValue{
		operatorValueBase: operatorValueBase{KindValue: MethodValueKind, Label: "<Java Method: " + name + ">", NormalizeMessage: "It is a TLC bug: Attempted to normalize an operator."},
		Name:              name,
		MinLevel:          minLevel,
		EvalFunc:          eval,
	}
}

func (v *MethodValue) Eval(args []Value, control int) (Value, error) {
	if v.EvalFunc == nil {
		return nil, v.unsupported("attempted to apply Java method %s without an implementation", v.Name)
	}
	return v.EvalFunc(args, control)
}

func (v *MethodValue) DeepCopy() Value { return v }
func (v *MethodValue) Permute(perm *MVPerm) Value {
	return v.operatorValueBase.Permute(perm)
}

type EvaluatingValue struct {
	operatorValueBase
	Name     string
	MinLevel int
	Priority int
	OpDef    *OpDefNode
	EvalFunc EvaluatingEvalFunc
}

func NewEvaluatingValue(name string, minLevel int, priority int, opDef *OpDefNode, eval EvaluatingEvalFunc) *EvaluatingValue {
	return &EvaluatingValue{
		operatorValueBase: operatorValueBase{KindValue: MethodValueKind, Label: "<Java Method: " + name + ">", NormalizeMessage: "It is a TLC bug: Attempted to normalize an operator."},
		Name:              name,
		MinLevel:          minLevel,
		Priority:          priority,
		OpDef:             opDef,
		EvalFunc:          eval,
	}
}

func (v *EvaluatingValue) Eval(args []Value, control int) (Value, error) {
	return nil, v.unsupported("it is a TLC bug: should use the unevaluated-argument eval method for %s", v)
}

func (v *EvaluatingValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	if v.EvalFunc != nil {
		value, err := v.EvalFunc(tool, args, con, state, pstate, control, cm)
		if err != nil || value != nil {
			return value, err
		}
	}
	if tool != nil && v.OpDef != nil {
		return tool.EvalPure(v.OpDef, args, con, state, pstate, control, cm)
	}
	return ValUndef, nil
}

func (v *EvaluatingValue) DeepCopy() Value { return v }
func (v *EvaluatingValue) Permute(perm *MVPerm) Value {
	return v.operatorValueBase.Permute(perm)
}

type PriorityEvaluatingValue struct {
	*EvaluatingValue
	Handles []*EvaluatingValue
}

func NewPriorityEvaluatingValue(primary *EvaluatingValue, secondary *EvaluatingValue) *PriorityEvaluatingValue {
	out := &PriorityEvaluatingValue{}
	out.Add(primary)
	out.Add(secondary)
	return out
}

func (v *PriorityEvaluatingValue) Add(ev *EvaluatingValue) {
	if ev == nil {
		return
	}
	if v.EvaluatingValue != nil && (v.OpDef != ev.OpDef || v.MinLevel != ev.MinLevel) {
		panic(newTLCError(ECGeneral, "priority evaluating values must refer to the same operator definition and level"))
	}
	v.Handles = append(v.Handles, ev)
	for i := 1; i < len(v.Handles); i++ {
		cur := v.Handles[i]
		j := i
		for j > 0 && v.Handles[j-1].Priority > cur.Priority {
			v.Handles[j] = v.Handles[j-1]
			j--
		}
		v.Handles[j] = cur
	}
	v.EvaluatingValue = v.Handles[0]
}

func (v *PriorityEvaluatingValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	for _, ev := range v.Handles {
		if ev.EvalFunc != nil {
			value, err := ev.EvalFunc(tool, args, con, state, pstate, control, cm)
			if err != nil || value != nil {
				return value, err
			}
		}
	}
	if tool != nil && v.OpDef != nil {
		return tool.Eval(v.OpDef.Body, con, state, pstate, control, cm)
	}
	return ValUndef, nil
}

type CallableValue struct {
	*EvaluatingValue
	CallableFunc CallableEvalFunc
}

func NewCallableValue(name string, minLevel int, opDef *OpDefNode, callable CallableEvalFunc) *CallableValue {
	return &CallableValue{
		EvaluatingValue: NewEvaluatingValue(name, minLevel, 100, opDef, nil),
		CallableFunc:    callable,
	}
}

func (v *CallableValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	argVals := make([]Value, len(args))
	for i, arg := range args {
		value, err := tool.Eval(arg, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		argVals[i] = value
	}
	if v.CallableFunc == nil {
		return nil, v.unsupported("attempted to apply callable Java method %s without an implementation", v.Name)
	}
	callable, err := v.CallableFunc(argVals)
	if err != nil {
		return nil, err
	}
	if pstate != nil {
		pstate.SetCallable(callable)
	}
	return BoolTrue, nil
}

func WithEvaluatingOpDef(value any, opDef *OpDefNode) any {
	if opDef == nil {
		return value
	}
	switch v := value.(type) {
	case *EvaluatingValue:
		return v.withOpDef(opDef)
	case *PriorityEvaluatingValue:
		out := &PriorityEvaluatingValue{}
		for _, handle := range v.Handles {
			out.Add(handle.withOpDef(opDef))
		}
		return out
	default:
		return value
	}
}

func (v *EvaluatingValue) withOpDef(opDef *OpDefNode) *EvaluatingValue {
	if v == nil || opDef == nil {
		return v
	}
	out := *v
	out.OpDef = opDef
	return &out
}

func EvalOperatorValue(op Value, args []Value, control int) (Value, error) {
	switch v := op.(type) {
	case *OpLambdaValue:
		return v.Eval(args, control)
	case *OpRcdValue:
		return v.Eval(args, control)
	case *MethodValue:
		return v.Eval(args, control)
	case *EvaluatingValue:
		return v.Eval(args, control)
	case *PriorityEvaluatingValue:
		return v.Eval(args, control)
	case *CallableValue:
		return v.Eval(args, control)
	default:
		return nil, newTLCError(ECGeneral, "attempted to apply non-operator value %s", op)
	}
}

func EvalOperatorValueWithTool(op Value, tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	switch v := op.(type) {
	case *EvaluatingValue:
		return v.EvalWithTool(tool, args, con, state, pstate, control, cm)
	case *PriorityEvaluatingValue:
		return v.EvalWithTool(tool, args, con, state, pstate, control, cm)
	case *CallableValue:
		return v.EvalWithTool(tool, args, con, state, pstate, control, cm)
	default:
		argVals := make([]Value, len(args))
		for i, arg := range args {
			value, err := tool.Eval(arg, con, state, pstate, control, cm)
			if err != nil {
				return nil, err
			}
			argVals[i] = value
		}
		return EvalOperatorValue(op, argVals, control)
	}
}
