package tlc

import (
	"fmt"
	"strings"
)

type OpDefNode struct {
	*SemanticNodeBase
	Symbol                    *SymbolNode
	Name                      *UniqueString
	Params                    []*SymbolNode
	Body                      SemanticNode
	DeclarationLocation       SourceLocation
	InRecursive               bool
	Local                     bool
	OriginallyDefinedInModule *ModuleNode
	SourceDefinition          *OpDefNode
	CompoundID                []*UniqueString
}

func (n *OpDefNode) IsLocal() bool   { return n.Local }
func (n *OpDefNode) HasSource() bool { return n.SourceDefinition != nil }
func (n *OpDefNode) GetOriginallyDefinedInModuleNode() *ModuleNode {
	return n.OriginallyDefinedInModule
}
func (n *OpDefNode) GetSource() *OpDefNode {
	if n.SourceDefinition != nil {
		return n.SourceDefinition
	}
	return n
}

func (n *OpDefNode) GetCompoundID() []*UniqueString {
	if n.CompoundID != nil {
		return n.CompoundID
	}
	return []*UniqueString{n.Name}
}

func (n *OpDefNode) GetLocalName() *UniqueString {
	if n.CompoundID != nil {
		return n.CompoundID[len(n.CompoundID)-1]
	}
	return n.Name
}

func (n *OpDefNode) HasPath() bool { return len(n.CompoundID) > 1 }

func (n *OpDefNode) GetPathName() *UniqueString {
	var names []string
	for _, part := range n.CompoundID[:max(0, len(n.CompoundID)-1)] {
		names = append(names, part.String())
	}
	return UniqueStringOf(strings.Join(names, "!"))
}

func NewOpDefNode(name string, params []*SymbolNode, body SemanticNode) *OpDefNode {
	return NewOpDefNodeForSymbol(NewSymbolNode(name), params, body)
}

func NewOpDefNodeForSymbol(symbol *SymbolNode, params []*SymbolNode, body SemanticNode) *OpDefNode {
	return NewOpDefNodeForSymbolWithBase(symbol, params, body, nil)
}

// A parser adapter borrows the already constructed SANY identity. Standalone
// runtime definitions allocate their own semantic base through the same path.
func NewOpDefNodeForSymbolWithBase(symbol *SymbolNode, params []*SymbolNode, body SemanticNode, base *SemanticNodeBase) *OpDefNode {
	outParams := make([]*SymbolNode, len(params))
	copy(outParams, params)
	if symbol == nil {
		symbol = NewSymbolNode("")
	}
	symbol.MarkUserDefinedOp()
	symbol.Arity = len(params)
	name := symbol.Name
	image := ""
	if name != nil {
		image = name.String()
	}
	if base == nil {
		owned := NewSemanticNodeBase(SemanticUserDefinedOpKind, image)
		base = &owned
	}
	definition := &OpDefNode{
		SemanticNodeBase: base,
		Symbol:           symbol,
		Name:             name,
		Params:           outParams,
		Body:             body,
	}
	symbol.Definition = definition
	symbol.SemanticBase = definition.SemanticNodeBase
	return definition
}

func (n *OpDefNode) Arity() int {
	if n == nil {
		return 0
	}
	if n.Kind() == SemanticBuiltInKind {
		return n.Symbol.Arity
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

func (n *OpDefNode) SetDeclarationLocation(location SourceLocation) {
	if n != nil {
		n.DeclarationLocation = location
	}
}

func (n *OpDefNode) GetDeclarationLocation() SourceLocation {
	if n == nil {
		return NullSourceLocation
	}
	return n.DeclarationLocation
}

func (n *OpDefNode) String() string {
	if n == nil || n.Name == nil {
		return "<anonymous>"
	}
	return n.Name.String()
}

func (n *OpDefNode) GetSignature() string {
	var out strings.Builder
	out.WriteString(n.Name.String())
	if n.Arity() > 0 && n.Kind() != SemanticBuiltInKind {
		out.WriteByte('(')
		for i, param := range n.Params {
			if param.GetTreeNode() != nil {
				out.WriteString(syntaxNodeHumanReadableImage(param))
				if i < len(n.Params)-1 {
					out.WriteString(", ")
				}
			}
		}
		out.WriteByte(')')
	}
	return out.String()
}

func (n *OpDefNode) GetComment() string {
	var out strings.Builder
	if tree, ok := n.GetTreeNode().(interface{ GetAttachedComments() []string }); ok {
		for _, comment := range tree.GetAttachedComments() {
			out.WriteString(comment)
			out.WriteByte('\n')
		}
	}
	return strings.TrimFunc(strings.ReplaceAll(out.String(), "\n$", ""), func(r rune) bool { return r <= 0x20 })
}

func (n *OpDefNode) GetHumanReadableImage() string {
	var out strings.Builder
	out.WriteString(n.GetComment())
	out.WriteByte('\n')
	tree, ok := n.GetTreeNode().(interface{ GetOneHumanReadableImages() []string })
	var images []string
	if ok {
		images = tree.GetOneHumanReadableImages()
	}
	if images != nil {
		for _, image := range images {
			out.WriteString(image)
			out.WriteByte(' ')
		}
	} else {
		out.WriteString(n.GetSourceLocation().String())
	}
	return strings.TrimFunc(out.String(), func(r rune) bool { return r <= 0x20 })
}

type OperatorEvalFunc func(args []Value, control int) (Value, error)
type EvaluatingEvalFunc func(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error)
type CallableEvalFunc func(args []Value) (func() (any, error), error)

type operatorValueBase struct {
	BaseValue
	owner            Value
	KindValue        ValueKind
	Label            string
	NormalizeMessage string
}

// Shared Go methods represent Java's concrete methods and inherited defaults.
// Keep their receiver identity, source, and dynamically dispatched printing.
func (b *operatorValueBase) receiver() Value {
	if b.owner != nil {
		return b.owner
	}
	return b
}

func (b *operatorValueBase) Kind() ValueKind { return b.KindValue }
func (b *operatorValueBase) KindString() string {
	return b.KindStringFor(b.KindValue)
}

func (b *operatorValueBase) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(b.receiver(), &err)
	if b.KindValue == MethodValueKind {
		return 0, b.unsupported("%s", b.methodComparisonMessage(other))
	}
	return 0, b.unsupported("Attempted to compare operator %s with value:\n%s", ValuesPPR(b.receiver()), ValuesPPRString(other.String()))
}

func (b *operatorValueBase) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(b.receiver(), &err)
	if b.KindValue == MethodValueKind {
		return false, b.unsupported("%s", b.methodComparisonMessage(other))
	}
	return false, b.unsupported("Attempted to check equality of operator %s with value:\n%s", ValuesPPR(b.receiver()), ValuesPPRString(other.String()))
}

func (b *operatorValueBase) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(b.receiver(), &err)
	if b.KindValue == MethodValueKind {
		if elem != nil {
			_ = elem.String()
		}
		return false, b.unsupported("%s\nis an element of operator %s", ValuesPPRString(elem.String()), b.receiver().String())
	}
	return false, b.unsupported("Attempted to check if the value:\n%s\nis an element of operator %s", ValuesPPRString(elem.String()), ValuesPPR(b.receiver()))
}

func (b *operatorValueBase) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return false, b.unsupported("Attempted to check if the operator %s is a finite set.", b.diagnosticString())
}

func (b *operatorValueBase) Size() (resultInt int, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return 0, b.unsupported("Attempted to compute the number of elements in the operator %s.", b.diagnosticString())
}

func (b *operatorValueBase) Normalize() Value {
	defer catchValueFailure(b.receiver(), nil)
	panic(fmt.Errorf("%s", b.normalizeMessage()))
}

func (b *operatorValueBase) DeepNormalize() {}

func (b *operatorValueBase) IsNormalized() bool {
	defer catchValueFailure(b.receiver(), nil)
	panic(fmt.Errorf("%s", b.normalizeMessage()))
}

func (b *operatorValueBase) IsDefined() bool { return true }
func (b *operatorValueBase) DeepCopy() Value { return b.receiver() }

func (b *operatorValueBase) FingerPrint(fp uint64) uint64 {
	return unsupportedValueFingerprint(b.receiver())
}

func (b *operatorValueBase) Permute(*MVPerm) Value {
	return unsupportedValuePermutation(b.receiver())
}

func (b *operatorValueBase) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return nil, b.unsupported("Attempted to appy EXCEPT construct to the operator %s.", b.diagnosticString())
}

func (b *operatorValueBase) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return nil, b.unsupported("Attempted to apply EXCEPT construct to the operator %s.", b.diagnosticString())
}

func (b *operatorValueBase) String() string {
	return ValueToString(b.receiver(), "", true)
}

func (b *operatorValueBase) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(b.receiver(), nil)
	sb.WriteString(b.Label)
	return sb
}

func (b *operatorValueBase) diagnosticString() string {
	if b.KindValue == MethodValueKind {
		return b.receiver().String()
	}
	return ValuesPPR(b.receiver())
}

func (b *operatorValueBase) methodComparisonMessage(other Value) string {
	// Preserve MethodValue/EvaluatingValue's Java precedence quirk: the entire
	// concatenation is tested for null, so only obj.toString() becomes the error.
	_ = b.receiver().String()
	if other != nil {
		_ = other.String()
	}
	return ValuesPPRString(other.String())
}

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
}

func NewOpLambdaValue(opDef *OpDefNode, tool *Tool, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cms ...CostModel) *OpLambdaValue {
	if con == nil {
		con = EmptyContext
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	out := &OpLambdaValue{
		operatorValueBase: operatorValueBase{BaseValue: newBaseValue(cm), KindValue: OpLambdaValueKind, Label: "<Operator " + opDef.String() + ">", NormalizeMessage: "Should not normalize an operator."},
		OpDef:             opDef,
		Tool:              tool,
		Con:               con,
		State:             state,
		PState:            pstate,
		Control:           control,
	}
	out.owner = out
	return out
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

func (v *OpLambdaValue) String() string {
	return ValueToString(v, "", true)
}

func (v *OpLambdaValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	sb.WriteString("<Operator " + v.OpDef.Name.String() + ">")
	return sb
}
func (v *OpLambdaValue) Permute(perm *MVPerm) Value {
	return v.operatorValueBase.Permute(perm)
}

func (v *OpLambdaValue) Eval(args []Value, control int) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if v.OpDef == nil {
		return nil, newTLCError(ECGeneral, "Attempted to apply a nil operator.")
	}
	if v.OpDef.Arity() != len(args) {
		return nil, v.unsupported("Applying the operator %s with wrong number of arguments.", ValuesPPR(v))
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
	return v.Tool.Eval(v.OpDef.Body, ctx, v.State, v.PState, control, DoNotRecordCostModel)
}

type OpRcdValue struct {
	operatorValueBase
	Domain [][]Value
	Values []Value
}

func NewOpRcdValue() *OpRcdValue {
	out := &OpRcdValue{
		operatorValueBase: operatorValueBase{KindValue: OpRcdValueKind, Label: "<Operator record>", NormalizeMessage: "Should not normalize an operator."},
	}
	out.owner = out
	return out
}

func NewOpRcdValueFrom(domain [][]Value, values []Value) *OpRcdValue {
	out := &OpRcdValue{
		operatorValueBase: operatorValueBase{KindValue: OpRcdValueKind, Label: "<Operator record>", NormalizeMessage: "Should not normalize an operator."},
		Domain:            domain,
		Values:            values,
	}
	out.owner = out
	return out
}

func (v *OpRcdValue) AddLine(values []Value) {
	defer catchValueFailure(v, nil)
	args := make([]Value, len(values)-2)
	copy(args, values[1:len(values)-1])
	v.Domain = append(v.Domain, args)
	v.Values = append(v.Values, values[len(values)-1])
}

func (v *OpRcdValue) Eval(args []Value, control int) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	_ = control
	for i, vals := range v.Domain {
		if len(args) != len(vals) {
			return nil, v.unsupported("Attempted to apply the operator %s\nwith wrong number of arguments.", ValuesPPR(v))
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
	return nil, v.unsupported("Attempted to apply operator:\n%s\nto arguments (%s), which is undefined.", ValuesPPR(v), joinValueStrings(args, ", "))
}

func (v *OpRcdValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	for _, value := range v.Values {
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *OpRcdValue) DeepCopy() Value { return v }
func (v *OpRcdValue) Permute(perm *MVPerm) Value {
	return v.operatorValueBase.Permute(perm)
}

func (v *OpRcdValue) String() string {
	return ValueToString(v, "", true)
}

func (v *OpRcdValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	sb.WriteString("{ ")
	for i, value := range v.Values {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("<")
		for _, arg := range v.Domain[i] {
			sb = appendValueString(arg, sb, offset, swallow)
			sb.WriteString(", ")
		}
		sb = appendValueString(value, sb, offset, swallow)
		sb.WriteString(">")
	}
	sb.WriteString("}")
	return sb
}

type MethodValue struct {
	operatorValueBase
	Name     string
	MinLevel int
	// ParameterCount retains the reflected Java method's arity for native
	// override registration. -1 denotes an unrepresented method signature.
	ParameterCount int
	EvalFunc       OperatorEvalFunc
}

func NewMethodValue(name string, minLevel int, eval OperatorEvalFunc) *MethodValue {
	out := &MethodValue{
		operatorValueBase: operatorValueBase{KindValue: MethodValueKind, Label: "<Java Method: " + name + ">", NormalizeMessage: "It is a TLC bug: Attempted to normalize an operator."},
		Name:              name,
		MinLevel:          minLevel,
		ParameterCount:    -1,
		EvalFunc:          eval,
	}
	out.owner = out
	return out
}

func (v *MethodValue) Eval(args []Value, control int) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	defer catchJavaMethodFailure(v.Name, &resultValue, &err, true)
	return v.EvalFunc(args, control)
}

// MethodValue preserves direct EvalExceptions; unevaluated and callable
// overrides use Assert.fail for every failure in their Java catch(Throwable).
// The defer belongs inside CallableValue's argument evaluation boundary.
func catchJavaMethodFailure(signature string, result *Value, err *error, preserveEval bool) {
	if failure := recover(); failure != nil {
		cause, ok := failure.(error)
		if !ok {
			cause = fmt.Errorf("%v", failure)
		}
		*result = nil
		*err = adaptJavaMethodFailure(signature, cause, preserveEval)
		return
	}
	if *err != nil {
		*result = nil
		*err = adaptJavaMethodFailure(signature, *err, preserveEval)
	}
}

func adaptJavaMethodFailure(signature string, cause error, preserveEval bool) error {
	if preserveEval {
		if isValueEvalException(cause) {
			return cause
		}
		if isJavaNullPointerException(cause) {
			return NewEvalExceptionNullable(ECTLCModuleValueJavaMethodOverride, javaString(signature), javaThrowableDetailMessage(cause))
		}
	}
	message := javaThrowableDetailMessage(cause)
	if preserveEval && message == nil {
		message = javaString(javaThrowableStackTrace(cause))
	}
	failure := newTLCErrorCodeNullable(ECTLCModuleValueJavaMethodOverride, javaString(signature), message)
	failure.Runtime = true
	return failure
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
	if GetOpCode(opDef.Name) != 0 {
		panic(javaMethodOverrideRuntimeError(name, "@Evaluation fallback to pure TLA+ definition only works for user-defined operators."))
	}
	return newEvaluatingValue(name, minLevel, priority, opDef, eval)
}

// CallableValue uses Java's protected constructor, which skips the pure-fallback guard.
func newEvaluatingValue(name string, minLevel int, priority int, opDef *OpDefNode, eval EvaluatingEvalFunc) *EvaluatingValue {
	out := &EvaluatingValue{
		operatorValueBase: operatorValueBase{KindValue: MethodValueKind, Label: "<Java Method: " + name + ">", NormalizeMessage: "It is a TLC bug: Attempted to normalize an operator."},
		Name:              name,
		MinLevel:          minLevel,
		Priority:          priority,
		OpDef:             opDef,
		EvalFunc:          eval,
	}
	out.owner = out
	return out
}

func (v *EvaluatingValue) Eval(args []Value, control int) (resultValue Value, err error) {
	defer catchValueFailure(v.receiver(), &err)
	return nil, fmt.Errorf("It is a TLC bug: Should use the other eval method.")
}

func (v *EvaluatingValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (resultValue Value, err error) {
	defer catchJavaMethodFailure(v.Name, &resultValue, &err, false)
	value, err := v.EvalFunc(tool, args, con, state, pstate, control, cm)
	if err != nil || value != nil {
		return value, err
	}
	return tool.EvalPure(v.OpDef, args, con, state, pstate, control, cm)
}

func (v *EvaluatingValue) DeepCopy() Value { return v.receiver() }
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
		panic(newTLCErrorCode(ECGeneral))
	}
	if v.EvaluatingValue == nil {
		// Java constructs a new wrapper from the primary method and adds this
		// wrapper itself as a handle. Sorting never changes its method metadata.
		primary := *ev
		primary.BaseValue = BaseValue{}
		v.EvaluatingValue = &primary
		v.owner = v
		ev = v.EvaluatingValue
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
}

func (v *PriorityEvaluatingValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (resultValue Value, err error) {
	defer catchJavaMethodFailure(v.Name, &resultValue, &err, false)
	for _, ev := range v.Handles {
		value, err := ev.EvalFunc(tool, args, con, state, pstate, control, cm)
		if err != nil || value != nil {
			return value, err
		}
	}
	return tool.Eval(v.OpDef.Body, con, state, pstate, control, cm)
}

type CallableValue struct {
	*EvaluatingValue
	CallableFunc CallableEvalFunc
}

func NewCallableValue(name string, minLevel int, opDef *OpDefNode, callable CallableEvalFunc) *CallableValue {
	out := &CallableValue{
		EvaluatingValue: newEvaluatingValue(name, minLevel, 100, opDef, nil),
		CallableFunc:    callable,
	}
	out.owner = out
	return out
}

func (v *CallableValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (resultValue Value, err error) {
	argVals := make([]Value, len(args))
	for i, arg := range args {
		value, err := tool.Eval(arg, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		argVals[i] = value
	}
	defer catchJavaMethodFailure(v.Name, &resultValue, &err, false)
	callable, err := v.CallableFunc(argVals)
	if err != nil {
		return nil, err
	}
	if pstate == nil {
		return nil, fmt.Errorf("null")
	}
	pstate.SetCallable(callable)
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
		out := &PriorityEvaluatingValue{EvaluatingValue: v.EvaluatingValue.withOpDef(opDef)}
		out.owner = out
		for _, handle := range v.Handles {
			if handle == v.EvaluatingValue {
				out.Add(out.EvaluatingValue)
			} else {
				out.Add(handle.withOpDef(opDef))
			}
		}
		return out
	case *CallableValue:
		return v.withOpDef(opDef)
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
	out.owner = &out
	return &out
}

func (v *CallableValue) withOpDef(opDef *OpDefNode) *CallableValue {
	if v == nil || opDef == nil {
		return v
	}
	out := *v
	out.EvaluatingValue = v.EvaluatingValue.withOpDef(opDef)
	out.owner = &out
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
		return nil, newTLCError(ECGeneral, "Attempted to apply non-operator value %s.", op)
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
