package tlc

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const maxJavaInt = 1<<31 - 1

const (
	TLCExceptionBreakpointsFilter   = "ExceptionBreakpointsFilter"
	TLCUnsatisfiedBreakpointsFilter = "UnsatisfiedBreakpointsFilter"
	TLCSpecBreakpointsFilter        = "SpecBreakpointsFilter"
	TLCInvariantBreakpointsFilter   = "InvariantBreakpointsFilter"
	TLCMultiWorkerWarningFilter     = "MultiWorkerWarningFilter"
)

type DebugStepDirection int

const (
	DebugStepContinue DebugStepDirection = iota
	DebugStepIn
	DebugStepOut
	DebugStepOver
)

func (d DebugStepDirection) String() string {
	switch d {
	case DebugStepContinue:
		return "Continue"
	case DebugStepIn:
		return "In"
	case DebugStepOut:
		return "Out"
	case DebugStepOver:
		return "Over"
	default:
		return "Unknown"
	}
}

type DebugGranularity int

const (
	DebugGranularityState DebugGranularity = iota
	DebugGranularityFormula
)

func (g DebugGranularity) String() string {
	switch g {
	case DebugGranularityState:
		return "State"
	case DebugGranularityFormula:
		return "Formula"
	default:
		return "Unknown"
	}
}

type DebugStep int

const (
	DebugStepCommandIn DebugStep = iota
	DebugStepCommandOut
	DebugStepCommandOver
	DebugStepCommandContinue
	DebugStepCommandReset
	DebugStepCommandResetStart
)

func (s DebugStep) String() string {
	switch s {
	case DebugStepCommandIn:
		return "In"
	case DebugStepCommandOut:
		return "Out"
	case DebugStepCommandOver:
		return "Over"
	case DebugStepCommandContinue:
		return "Continue"
	case DebugStepCommandReset:
		return "Reset"
	case DebugStepCommandResetStart:
		return "Reset_Start"
	default:
		return "Unknown"
	}
}

type TLCStackFrame struct {
	ID               int
	Name             string
	Node             SemanticNode
	Context          *Context
	Tool             *Tool
	Exception        error
	Value            Value
	PresentationHint string
	Parent           *TLCStackFrame
	ContextID        int
	NestedVariables  *InsMap[int, *DebugTLCVariable]
	NestedConstants  *InsMap[int, []*DebugTLCVariable]

	finalsCaptured       bool
	nodeOwner            SemanticNode
	contextOwner         *Context
	toolOwner            *Tool
	exceptionOwner       error
	contextIDOwner       int
	nestedVariablesOwner *InsMap[int, *DebugTLCVariable]
	nestedConstantsOwner *InsMap[int, []*DebugTLCVariable]
}

func (f *TLCStackFrame) require() *TLCStackFrame {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f
}

func NewTLCStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, exception error) *TLCStackFrame {
	if tool == nil {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	if isNil(exception) {
		exception = nil
	}
	frameNode, frameContext := node, ctxt
	if detailed, ok := exception.(*TLCError); ok && detailed.detailed {
		frameNode, frameContext = detailed.ownedExpr(), detailed.ownedContext()
	}
	if isNil(frameNode) || frameContext == nil {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	frame := &TLCStackFrame{
		Parent:    parent,
		Node:      frameNode,
		Context:   frameContext,
		Tool:      tool,
		Exception: exception,
		ContextID: debugVariableReference(nil),
	}
	frame.NestedVariables = NewInsMap[int, *DebugTLCVariable]()
	frame.NestedConstants = NewInsMap[int, []*DebugTLCVariable]()
	// One semantic node can occur repeatedly on a recursive evaluation stack.
	frame.ID = int(int32(semanticNodeDebugID(node)) ^ int32(debugVariableReference(nil)))
	frame.Name = semanticNodeDebugName(node, exception)
	if exception != nil {
		frame.PresentationHint = "subtle"
	}
	frame.captureFinals()
	return frame
}

func NewTLCStackFrameNoException(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool) *TLCStackFrame {
	return NewTLCStackFrame(parent, node, ctxt, tool, nil)
}

func (f *TLCStackFrame) GetNode() SemanticNode {
	f.require()
	if f.finalsCaptured {
		return f.nodeOwner
	}
	return f.Node
}

func (f *TLCStackFrame) GetContext() *Context {
	f.require()
	if f.finalsCaptured {
		return f.contextOwner
	}
	return f.Context
}

func (f *TLCStackFrame) GetTool() *Tool {
	f.require()
	if f.finalsCaptured {
		return f.toolOwner
	}
	return f.Tool
}

func (f *TLCStackFrame) HasException() bool {
	return !isNil(f.GetException())
}

func (f *TLCStackFrame) GetException() error {
	f.require()
	exception := f.Exception
	if f.finalsCaptured {
		exception = f.exceptionOwner
	}
	if isNil(exception) {
		return nil
	}
	return exception
}

func (f *TLCStackFrame) ownedContextID() int {
	if f.finalsCaptured {
		return f.contextIDOwner
	}
	return f.ContextID
}

func (f *TLCStackFrame) ownedNestedVariables() *InsMap[int, *DebugTLCVariable] {
	if f.finalsCaptured {
		return f.nestedVariablesOwner
	}
	return f.NestedVariables
}

func (f *TLCStackFrame) ownedNestedConstants() *InsMap[int, []*DebugTLCVariable] {
	if f.finalsCaptured {
		return f.nestedConstantsOwner
	}
	return f.NestedConstants
}

func (f *TLCStackFrame) captureFinals() {
	f.nodeOwner = f.Node
	f.contextOwner = f.Context
	f.toolOwner = f.Tool
	f.exceptionOwner = f.Exception
	f.contextIDOwner = f.ContextID
	f.nestedVariablesOwner = f.NestedVariables
	f.nestedConstantsOwner = f.NestedConstants
	f.finalsCaptured = true
}

func (f *TLCStackFrame) GetStoppedEventArgument() TLCStoppedEvent {
	f.require()
	event := TLCStoppedEvent{ThreadID: 0, Reason: ""}
	exception := f.GetException()
	if isNil(exception) {
		return event
	}
	event.Reason = "exception"
	message := javaThrowableDetailMessage(exception)
	if message == nil {
		panic(NewNullPointerException())
	}
	lines := strings.Split(*message, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "@!@!@") {
			lines[i] = ""
		}
	}
	event.Text = strings.Join(lines, "\n")
	return event
}

func (f *TLCStackFrame) String() string {
	f.require()
	return fmt.Sprintf("TLCStackFrame [node=%v]", f.GetNode())
}

func (f *TLCStackFrame) SetValue(value Value) Value {
	f.require()
	f.Value = value
	return value
}

func (f *TLCStackFrame) GetConstantsID() int {
	f.require()
	return int(int32(f.ownedContextID()) + 1)
}

func (f *TLCStackFrame) GetStackID() int {
	f.require()
	return int(int32(f.ownedContextID()) + 3)
}

func (f *TLCStackFrame) GetScopes() []TLCScope {
	f.require()
	scopes := []TLCScope{}
	context := f.GetContext()
	if context != nil && !context.IsEmpty() {
		scopes = append(scopes, TLCScope{Name: debugScopeContext, VariablesReference: f.ownedContextID()})
	}
	tool := f.GetTool()
	if tool != nil && tool.ownedSpecProcessor() != nil && tool.ownedSpecProcessor().GetConstantDefns().Len() > 0 {
		scopes = append(scopes, TLCScope{Name: debugScopeConstants, VariablesReference: f.GetConstantsID(), PresentationHint: debugScopeHintRegisters})
	}
	if f.HasStackVariables() {
		scopes = append(scopes, TLCScope{Name: debugScopeStack, VariablesReference: f.GetStackID()})
	}
	return scopes
}

func (f *TLCStackFrame) GetVariables(ref int, rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	ref = int(int32(ref))
	return f.getVariablesWithStates(ref, rnd, EmptyState, nil, true)
}

func (f *TLCStackFrame) getVariablesWithStates(ref int, rnd *rand.Rand, s, t *TLCStateMut, catchNull bool) []*DebugTLCVariable {
	if f == nil {
		return nil
	}
	if tool := f.GetTool(); tool != nil {
		value, _ := tool.withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) {
			return f.getVariables(ref, rnd, s, t, catchNull), nil
		})
		return value.([]*DebugTLCVariable)
	}
	return f.getVariables(ref, rnd, s, t, catchNull)
}

func (f *TLCStackFrame) getVariables(ref int, rnd *rand.Rand, s, t *TLCStateMut, catchNull bool) []*DebugTLCVariable {
	if rnd == nil {
		rnd = rand.New(rand.NewSource(1))
	}
	var variables []*DebugTLCVariable
	if nestedVariables := f.ownedNestedVariables(); nestedVariables != nil {
		if variable, ok := nestedVariables.Get2(ref); ok && variable != nil {
			for _, nested := range variable.GetNested(rnd) {
				f.rememberNestedVariable(nested)
				variables = append(variables, nested)
			}
		}
	}
	if nestedConstants := f.ownedNestedConstants(); nestedConstants != nil {
		if constants, ok := nestedConstants.Get2(ref); ok {
			for _, constant := range constants {
				f.rememberNestedVariable(constant)
				variables = append(variables, constant)
				for _, nested := range constant.GetNested(rnd) {
					f.rememberNestedVariable(nested)
				}
			}
		}
	}
	switch ref {
	case f.ownedContextID():
		variables = append(variables, f.contextVariables(rnd, s, t, catchNull)...)
	case f.GetConstantsID():
		variables = append(variables, f.constantVariables(rnd)...)
	case f.GetStackID():
		return f.stackVariables(rnd)
	}
	return sortedDistinctDebugVariables(variables)
}

func (f *TLCStackFrame) GetConstants(rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	return f.GetVariables(f.GetConstantsID(), rnd)
}

func (f *TLCStackFrame) GetExceptionAsVariable() []*DebugTLCVariable {
	if f == nil || isNil(f.GetException()) {
		panic(NewNullPointerException())
	}
	node := f.GetNode()
	exception := f.GetException()
	name := semanticNodeJavaString(node)
	if node, ok := node.(interface{ GetHumanReadableImage() string }); ok {
		name = node.GetHumanReadableImage()
	}
	message := javaThrowableDetailMessage(exception)
	value := ""
	if message != nil {
		value = *message
	}
	typeName := javaThrowableClassName(exception)
	typeName = typeName[strings.LastIndexAny(typeName, ".$")+1:]
	return []*DebugTLCVariable{{
		Name: name, Value: value, ValueNull: message == nil, Type: typeName,
	}}
}

func (f *TLCStackFrame) HasStackVariables() bool {
	f.require()
	for cur := f; cur != nil; cur = cur.Parent {
		if !isNil(cur.Value) {
			return true
		}
	}
	return false
}

func (f *TLCStackFrame) contextVariables(rnd *rand.Rand, s, t *TLCStateMut, catchNull bool) []*DebugTLCVariable {
	var variables []*DebugTLCVariable
	for c := f.GetContext(); c != nil && c != EmptyContext; c = c.Next() {
		name := c.Name()
		value := c.Value()
		if isNil(value) {
			value = nil
		}
		if name == nil && value == nil {
			continue
		}
		if workerValue, ok := value.(*WorkerValue); ok {
			value = workerValue.ValueForWorker(CurrentThreadIDOr(0))
		}
		if lazy := asLazyValue(value); lazy != nil {
			// Eval deliberately bypasses GetValue: displaying a lazy binding must
			// not update the evaluator's cached value or cache count.
			evaluated, err := debugUnlazy(lazy, f.GetTool(), s, t, catchNull)
			if err != nil {
				variables = append(variables, &DebugTLCVariable{Name: symbolNodeDebugName(name), Value: fmt.Sprint(value), Type: err.Error()})
				continue
			}
			value = evaluated
		}
		switch v := value.(type) {
		case Value:
			variables = append(variables, f.debugVariableForValue(v, symbolNodeDebugName(name), rnd))
		case interface{ GetHumanReadableImage() string }:
			variables = append(variables, &DebugTLCVariable{Name: name.GetSignature(), Value: v.GetHumanReadableImage()})
		case error:
			variables = append(variables, &DebugTLCVariable{Name: symbolNodeDebugName(name), Value: fmt.Sprint(c.Value()), Type: v.Error()})
		default:
			fmt.Fprintf(os.Stderr, "This is interesting!!! What's this??? %v\n", value)
		}
	}
	return variables
}

func (f *TLCStackFrame) constantVariables(rnd *rand.Rand) []*DebugTLCVariable {
	if f == nil {
		return nil
	}
	tool := f.GetTool()
	if tool == nil || tool.ownedSpecProcessor() == nil {
		return nil
	}
	constants := tool.ownedSpecProcessor().GetConstantDefns()
	var out []*DebugTLCVariable
	for module, definitions := range constants.All() {
		if constants.Len() == 1 {
			for node, value := range definitions.All() {
				out = append(out, f.debugVariableForValue(MuxWorkerValue(value, CurrentThreadIDOr(0)), constantDefinitionName(node), rnd))
			}
			continue
		}
		variable := &DebugTLCVariable{Name: module.Name.String(), Value: module.Name.String(), VariablesReference: debugVariableReference(rnd)}
		for node := range definitions.All() {
			if def, ok := node.(*OpDefNode); ok {
				variable.Value = def.GetPathName().String()
				break
			}
		}
		out = append(out, variable)
		var nested []*DebugTLCVariable
		for node, value := range definitions.All() {
			name := constantDefinitionName(node)
			if def, ok := node.(*OpDefNode); ok {
				name = def.GetLocalName().String()
			}
			nested = append(nested, f.debugVariableForValue(MuxWorkerValue(value, CurrentThreadIDOr(0)), name, rnd))
		}
		nestedConstants := f.ownedNestedConstants()
		if nestedConstants == nil {
			nestedConstants = NewInsMap[int, []*DebugTLCVariable]()
			f.NestedConstants = nestedConstants
		}
		nestedConstants.Set(variable.VariablesReference, nested)
	}
	return out
}

func constantDefinitionName(node SemanticNode) string {
	switch node := node.(type) {
	case *OpDefNode:
		return node.Name.String()
	case *SymbolNode:
		return node.GetName().String()
	default:
		return SemanticString(node)
	}
}

func (f *TLCStackFrame) stackVariables(rnd *rand.Rand) []*DebugTLCVariable {
	var variables []*DebugTLCVariable
	for cur := f; cur != nil; cur = cur.Parent {
		if isNil(cur.Value) {
			continue
		}
		variable := cur.debugVariableForValue(cur.Value, syntaxNodeHumanReadableImage(cur.GetNode()), rnd)
		duplicate := false
		for _, prior := range variables {
			// Variable.equals excludes the transient TLC value, but includes
			// its nested reference as well as its display fields.
			if prior.Name == variable.Name && prior.Value == variable.Value && prior.Type == variable.Type && prior.VariablesReference == variable.VariablesReference && prior.VSCodeVariableMenuContext == variable.VSCodeVariableMenuContext {
				duplicate = true
				break
			}
		}
		if !duplicate {
			variables = append(variables, variable)
		}
	}
	return variables
}

func (f *TLCStackFrame) debugVariableForValue(value Value, name string, rnd *rand.Rand) *DebugTLCVariable {
	variable := ValueToTLCVariable(value, NewDebugTLCVariableName(name).SetInstance(value), rnd)
	f.rememberNestedVariable(variable)
	return variable
}

func (f *TLCStackFrame) rememberNestedVariable(variable *DebugTLCVariable) {
	if f == nil || variable == nil {
		return
	}
	nestedVariables := f.ownedNestedVariables()
	if nestedVariables == nil {
		nestedVariables = NewInsMap[int, *DebugTLCVariable]()
		f.NestedVariables = nestedVariables
	}
	nestedVariables.Set(variable.VariablesReference, variable)
}

func (f *TLCStackFrame) MatchesFrame(other *TLCStackFrame) bool {
	f.require()
	if other == nil {
		panic(NewNullPointerException())
	}
	left, leftOK := semanticNodeSyntax(f.GetNode())
	right, rightOK := semanticNodeSyntax(other.GetNode())
	return leftOK && rightOK && left.GetLevel() == right.GetLevel() && left.GetOperatorDefinition() == right.GetOperatorDefinition()
}

func semanticNodeSyntax(node SemanticNode) (interface {
	GetLevel() int
	GetOperatorDefinition() any
}, bool) {
	if node, ok := node.(interface{ GetTreeNode() any }); ok {
		syntax, ok := node.GetTreeNode().(interface {
			GetLevel() int
			GetOperatorDefinition() any
		})
		return syntax, ok
	}
	return nil, false
}

func (f *TLCStackFrame) MatchesBreakpoint(bp *TLCSourceBreakpoint) bool {
	f.require()
	if bp == nil {
		panic(NewNullPointerException())
	}
	loc, ok := semanticNodeSourceLocation(f.GetNode())
	if !ok {
		return false
	}
	return bp.MatchesLocation(loc)
}

func (f *TLCStackFrame) MatchesNode(expr SemanticNode) bool {
	return semanticNodeSame(f.GetNode(), expr)
}

func (f *TLCStackFrame) MatchesException(expr SemanticNode, exception error) bool {
	f.require()
	if semanticNodeSame(f.GetNode(), expr) {
		return true
	}
	detailed, ok := exception.(*TLCError)
	return ok && detailed != nil && detailed.detailed && semanticNodeSame(detailed.ownedExpr(), f.GetNode())
}

func (f *TLCStackFrame) IsTarget(expr SemanticNode) bool {
	return f.MatchesNode(expr)
}

type ResetEvalException struct {
	*RuntimeException
	Frame *TLCStackFrame

	frameCaptured bool
	frameOwner    *TLCStackFrame
}

func NewResetEvalException(frame *TLCStackFrame) *ResetEvalException {
	if frame == nil {
		panic(NewAssertionError())
	}
	return &ResetEvalException{RuntimeException: NewRuntimeException(), Frame: frame, frameCaptured: true, frameOwner: frame}
}

func (e *ResetEvalException) ownedFrame() *TLCStackFrame {
	if e.frameCaptured {
		return e.frameOwner
	}
	return e.Frame
}

func (e *ResetEvalException) Error() string { return javaThrowableMessage(e) }

func (e *ResetEvalException) GetMessage() *string {
	e.require()
	return e.RuntimeException.GetMessage()
}

func (e *ResetEvalException) GetCause() error {
	e.require()
	return e.RuntimeException.GetCause()
}

func (e *ResetEvalException) IsTarget(expr SemanticNode) bool {
	e.require()
	frame := e.ownedFrame()
	if frame == nil {
		panic(NewNullPointerException())
	}
	return frame.IsTarget(expr)
}

func (e *ResetEvalException) require() {
	if e == nil || e.RuntimeException == nil {
		panic(NewNullPointerException())
	}
}

type AbortEvalException struct{ *RuntimeException }

func NewAbortEvalException() *AbortEvalException {
	return &AbortEvalException{NewRuntimeException()}
}

func (e *AbortEvalException) Error() string { return javaThrowableMessage(e) }

func (e *AbortEvalException) GetMessage() *string {
	e.require()
	return e.RuntimeException.GetMessage()
}

func (e *AbortEvalException) GetCause() error {
	e.require()
	return e.RuntimeException.GetCause()
}

func (e *AbortEvalException) require() {
	if e == nil || e.RuntimeException == nil {
		panic(NewNullPointerException())
	}
}

type DebuggerValue struct{ *StringValue }

const debuggerPendingType = "Evaluation pending... (value has not been determined yet)"

func (v *DebuggerValue) require() *StringValue {
	if v == nil || v.StringValue == nil {
		panic(NewNullPointerException())
	}
	return v.StringValue
}

func (v *DebuggerValue) GetKindString() string { return v.GetTypeString() }
func (v *DebuggerValue) GetCostModel() CostModel {
	return v.require().GetCostModel()
}
func (v *DebuggerValue) GetTypeString() string {
	v.require()
	return debuggerPendingType
}
func (v *DebuggerValue) Kind() ValueKind    { return v.require().Kind() }
func (v *DebuggerValue) KindString() string { return v.GetKindString() }
func (v *DebuggerValue) Normalize() Value   { return v.require().Normalize() }
func (v *DebuggerValue) DeepNormalize()     { v.require().DeepNormalize() }
func (v *DebuggerValue) IsNormalized() bool { return v.require().IsNormalized() }
func (v *DebuggerValue) IsDefined() bool    { return v.require().IsDefined() }
func (v *DebuggerValue) DeepCopy() Value    { return v.require().DeepCopy() }
func (v *DebuggerValue) Permute(perm *MVPerm) Value {
	return v.require().Permute(perm)
}
func (v *DebuggerValue) Compare(other Value) (int, error) {
	return v.require().Compare(other)
}
func (v *DebuggerValue) Equal(other Value) (bool, error) {
	return v.require().Equal(other)
}
func (v *DebuggerValue) Member(elem Value) (bool, error) {
	return v.require().Member(elem)
}
func (v *DebuggerValue) IsFinite() (bool, error) { return v.require().IsFinite() }
func (v *DebuggerValue) Size() (int, error)      { return v.require().Size() }
func (v *DebuggerValue) FingerPrint(fp uint64) uint64 {
	return v.require().FingerPrint(fp)
}
func (v *DebuggerValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.require().TakeExcept(ex)
}
func (v *DebuggerValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.require().TakeExcepts(exs)
}
func (v *DebuggerValue) String() string {
	v.require()
	return "?"
}
func (v *DebuggerValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	v.require()
	sb.WriteString("?")
	return sb
}
func newDebuggerValue() *DebuggerValue {
	v := &DebuggerValue{NewStringValue("?")}
	v.StringValue.owner = v
	return v
}

var DebuggerNotEvaluatedValue Value = newDebuggerValue()

const (
	debugScopeException  = "Exception"
	debugScopeConstants  = "Constants"
	debugScopeContext    = "Context"
	debugScopeStack      = "Stack"
	debugScopeState      = "State"
	debugScopeAction     = "Action"
	debugScopeInitials   = "Initials"
	debugScopeSuccessors = "Successors"
	debugScopeTrace      = "Trace"

	debugScopeHintRegisters = "registers"
)

type TLCStateStackFrame struct {
	TLCStackFrame
	State   *TLCStateMut
	StateID int

	stateFinalsCaptured bool
	stateOwner          *TLCStateMut
	stateIDOwner        int
}

func (f *TLCStateStackFrame) require() *TLCStateStackFrame {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f
}

func NewTLCStateStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, state *TLCStateMut, exception error) *TLCStateStackFrame {
	base := NewTLCStackFrame(parent, node, ctxt, tool, exception)
	copied := state.DeepCopy()
	if !copied.retainsExtendedMetadata() {
		panic(NewAssertionError())
	}
	frame := &TLCStateStackFrame{
		TLCStackFrame: *base,
		State:         copied,
		StateID:       debugVariableReference(nil),
	}
	frame.captureStateFinals()
	return frame
}

func NewTLCStateStackFrameNoException(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, state *TLCStateMut) *TLCStateStackFrame {
	return NewTLCStateStackFrame(parent, node, ctxt, tool, state, nil)
}

func (f *TLCStateStackFrame) GetS() *TLCStateMut {
	f.require()
	return f.GetT()
}

func (f *TLCStateStackFrame) GetT() *TLCStateMut {
	f.require()
	return f.ownedState()
}

func (f *TLCStateStackFrame) ownedState() *TLCStateMut {
	if f.stateFinalsCaptured {
		return f.stateOwner
	}
	return f.State
}

func (f *TLCStateStackFrame) ownedStateID() int {
	if f.stateFinalsCaptured {
		return f.stateIDOwner
	}
	return f.StateID
}

func (f *TLCStateStackFrame) captureStateFinals() {
	f.stateOwner = f.State
	f.stateIDOwner = f.StateID
	f.stateFinalsCaptured = true
}

func (f *TLCStateStackFrame) AddT() bool {
	f.require()
	return false
}

func (f *TLCStateStackFrame) ToRecordValue() *RecordValue {
	state := f.GetT()
	if state == nil {
		return EmptyRecord
	}
	return debugStateRecordValue(state, nil)
}

func (f *TLCStateStackFrame) ToVariable(rnd *rand.Rand) *DebugTLCVariable {
	state := f.GetT()
	name := "State"
	if state != nil {
		name = fmt.Sprintf("%d: %s", state.Level(), debugActionLocation(state.GetAction()))
	}
	return f.getStateAsVariable(f.ToRecordValue(), name, rnd)
}

func (f *TLCStateStackFrame) GetVariables(ref int, rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	ref = int(int32(ref))
	if ref == f.ownedStateID() {
		return f.debugVariables(func() []*DebugTLCVariable { return []*DebugTLCVariable{f.ToVariable(rnd)} })
	}
	if ref == int(int32(f.ownedStateID())+1) {
		return f.TraceVariables(rnd)
	}
	return f.getVariablesWithStates(ref, rnd, f.GetS(), nil, false)
}

func (f *TLCStateStackFrame) GetScopes() []TLCScope {
	f.require()
	scopes := append([]TLCScope{}, f.TLCStackFrame.GetScopes()...)
	if f.HasScope() {
		scopes = append(scopes, TLCScope{Name: f.ScopeName(), VariablesReference: f.ownedStateID()})
	}
	scopes = append(scopes, TLCScope{Name: debugScopeTrace, VariablesReference: int(int32(f.ownedStateID()) + 1)})
	return scopes
}

func (f *TLCStateStackFrame) MatchesBreakpoint(bp *TLCSourceBreakpoint) bool {
	f.require()
	return f.matchesStateBreakpoint(bp, f.GetS(), f.GetT())
}

func (f *TLCStateStackFrame) matchesStateBreakpoint(bp *TLCSourceBreakpoint, s, t *TLCStateMut) bool {
	if !f.TLCStackFrame.MatchesBreakpoint(bp) {
		return false
	}
	fire := true
	if bp.GetHits() > 0 {
		fire = t != nil && t.Level() >= bp.GetHits()
	}
	return bp.MatchesExpression(f.GetTool(), s, t, f.GetContext(), fire)
}

func (f *TLCActionStackFrame) MatchesBreakpoint(bp *TLCSourceBreakpoint) bool {
	f.require()
	return f.matchesStateBreakpoint(bp, f.GetS(), f.GetT())
}

func (f *TLCStateStackFrame) HasScope() bool {
	f.require()
	return true
}

func (f *TLCStateStackFrame) ScopeName() string {
	f.require()
	return debugScopeState
}

func (f *TLCStackFrame) debugVariables(fn func() []*DebugTLCVariable) []*DebugTLCVariable {
	value, _ := f.GetTool().withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) { return fn(), nil })
	return value.([]*DebugTLCVariable)
}

func (f *TLCStackFrame) getStateAsVariable(record Value, name string, rnd *rand.Rand) *DebugTLCVariable {
	variable := f.debugVariableForValue(record, name, rnd)
	variable.Type = "State"
	variable.SetVscodeVariableMenuContext("state")
	return variable
}

func (f *TLCStateStackFrame) getStateAsVariable(record Value, name string, rnd *rand.Rand) *DebugTLCVariable {
	variable := f.TLCStackFrame.getStateAsVariable(record, name, rnd)
	variable.Type = debugStateVariableType(f.GetT())
	return variable
}

func (f *TLCStateStackFrame) TraceVariables(rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	return f.debugVariables(func() []*DebugTLCVariable { return f.traceVariables(rnd, f.AddT(), false, nil) })
}

// getT and the state's variable renderer are inherited by action frames. Only
// next-state frames add the current state to the trace before its predecessors.
func (f *TLCStateStackFrame) traceVariables(rnd *rand.Rand, addT, padded bool, idToState map[int]*TLCStateMut) []*DebugTLCVariable {
	t := f.GetT()
	if t == nil {
		return nil
	}
	render := func(state *TLCStateMut, name string) *DebugTLCVariable {
		return f.getStateAsVariable(debugStateRecordValue(state, nil), name, rnd)
	}
	if t.IsInitial() && !padded {
		if t.Predecessor() != nil {
			panic(NewAssertionError())
		}
		name := InitialPredicate
		if t.HasAction() {
			name = debugActionLocation(t.GetAction())
		}
		return []*DebugTLCVariable{render(t, "1: "+name)}
	}
	var trace []*DebugTLCVariable
	if addT {
		trace = append(trace, render(t, fmt.Sprintf("%d: %s", t.Level(), debugActionLocation(t.GetAction()))))
	}
	var prefix []*TLCStateInfo
	if simulator := CurrentSimulator(); simulator != nil {
		states := debuggerVariableSimulationTrace(simulator.GetUncompressedTrace(t))
		for i := 0; i < states.Size(); i++ {
			state := states.At(i)
			if state.AllAssigned() {
				prefix = append(prefix, NewTLCStateInfo(state))
			}
		}
	} else {
		last := t
		for state := t.Predecessor(); state != nil; state = state.Predecessor() {
			if state.IsInitial() {
				name := InitialPredicate
				if state.HasAction() {
					name = debugActionLocation(state.GetAction())
				}
				return append(trace, render(state, "1: "+name))
			}
			trace = append(trace, render(state, fmt.Sprintf("%d: %s", state.Level(), debugActionLocation(state.GetAction()))))
			last = state
		}
		checker := MainChecker()
		if checker == nil {
			panic(NewNullPointerException())
		}
		var ok bool
		prefix, ok = debuggerTraceInfo(func() []*TLCStateInfo { return checker.GetTraceInfo(last) })
		if !ok {
			return nil
		}
	}
	width := len(strconv.Itoa(len(prefix)))
	for i := len(prefix) - 1; i >= 0; i-- {
		info := prefix[i]
		state := info.ownedState()
		name := fmt.Sprintf("%d: %v", state.Level(), info.Info)
		if padded {
			name = fmt.Sprintf("%0*d: %v", width, state.Level(), info.Info)
		}
		variable := render(state, name)
		if idToState != nil {
			idToState[variable.VariablesReference] = state
		}
		trace = append(trace, variable)
	}
	return trace
}

func (f *TLCStateStackFrame) GetTraceAsStackFrames() []*TLCSyntheticStateStackFrame {
	f.require()
	value, _ := f.GetTool().withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) {
		return f.getTraceAsStackFrames(f.GetS(), f.AddT()), nil
	})
	return value.([]*TLCSyntheticStateStackFrame)
}

func (f *TLCStateStackFrame) getTraceAsStackFrames(state *TLCStateMut, addT bool) []*TLCSyntheticStateStackFrame {
	if simulator := CurrentSimulator(); simulator != nil {
		trace := debuggerFrameSimulationTrace(simulator.GetUncompressedTrace(state))
		if trace.Size() == 0 {
			return nil
		}
		successor := state
		width := len(strconv.Itoa(trace.Size()))
		frames := make([]*TLCSyntheticStateStackFrame, 0, trace.Size())
		for i := trace.Size() - 1; i >= 0; i-- {
			state := trace.At(i)
			frames = append(frames, NewTLCSyntheticStateStackFrame(f.GetTool(), NewTLCStateInfo(state), successor, width))
			successor = state
		}
		return frames
	}

	if state == nil {
		return nil
	}
	if state.IsInitial() {
		if state.AllAssigned() {
			return []*TLCSyntheticStateStackFrame{
				NewTLCSyntheticStateStackFrame(f.GetTool(), NewTLCStateInfo(state), state, 1),
			}
		}
		return nil
	}

	suffix := make([]*TLCStateInfo, 0)
	if addT {
		suffix = append(suffix, NewTLCStateInfo(state))
	}
	last := state
	for s := state.Predecessor(); s != nil; s = s.Predecessor() {
		if s.IsInitial() {
			suffix = append(suffix, NewTLCStateInfo(s))
			return syntheticTraceFramesFromNewestFirst(f.GetTool(), suffix, state)
		}
		suffix = append(suffix, NewTLCStateInfo(s))
		last = s
	}

	fullTrace := make([]*TLCStateInfo, 0, len(suffix))
	checker := MainChecker()
	if checker == nil {
		panic(NewNullPointerException())
	}
	prefix, ok := debuggerTraceInfo(func() []*TLCStateInfo { return checker.GetTraceInfo(last) })
	if !ok {
		return nil
	}
	fullTrace = append(fullTrace, prefix...)
	for i := len(suffix) - 1; i >= 0; i-- {
		fullTrace = append(fullTrace, suffix[i])
	}
	return syntheticTraceFramesFromChronological(f.GetTool(), fullTrace, state)
}

func debuggerVariableSimulationTrace(trace *StateVec) *StateVec {
	if trace == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.StateVec.stream()" because the return value of "tlc2.tool.Simulator.getUncompressedTrace(tlc2.tool.TLCState)" is null`))
	}
	return trace
}

func debuggerFrameSimulationTrace(trace *StateVec) *StateVec {
	if trace == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.StateVec.size()" because "<local2>" is null`))
	}
	return trace
}

// Java debugger trace reconstruction treats checked I/O failures as an absent
// trace while allowing unchecked failures to escape.
func debuggerTraceInfo(get func() []*TLCStateInfo) (trace []*TLCStateInfo, ok bool) {
	defer func() {
		if failure := recover(); failure != nil {
			if err, isError := failure.(error); isError && isJavaIOException(err) {
				trace = nil
				ok = false
				return
			}
			panic(failure)
		}
	}()
	return get(), true
}

func (f *TLCActionStackFrame) GetTraceAsStackFrames() []*TLCSyntheticStateStackFrame {
	f.require()
	value, _ := f.GetTool().withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) {
		return f.getTraceAsStackFrames(f.GetS(), f.AddT()), nil
	})
	return value.([]*TLCSyntheticStateStackFrame)
}

func (f *TLCNextStatesStackFrame) GetTraceAsStackFrames() []*TLCSyntheticStateStackFrame {
	f.require()
	value, _ := f.GetTool().withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) {
		return f.getTraceAsStackFrames(f.GetS(), f.AddT()), nil
	})
	return value.([]*TLCSyntheticStateStackFrame)
}

func (f *TLCDebuggerFrame) GetTraceAsStackFrames() []*TLCSyntheticStateStackFrame {
	switch {
	case f.Action != nil:
		return f.Action.GetTraceAsStackFrames()
	case f.Next != nil:
		return f.Next.GetTraceAsStackFrames()
	case f.State != nil:
		return f.State.GetTraceAsStackFrames()
	default:
		return nil
	}
}

func syntheticTraceFramesFromNewestFirst(tool *Tool, trace []*TLCStateInfo, successor *TLCStateMut) []*TLCSyntheticStateStackFrame {
	if len(trace) == 0 {
		return nil
	}
	width := len(strconv.Itoa(len(trace)))
	frames := make([]*TLCSyntheticStateStackFrame, 0, len(trace))
	for _, info := range trace {
		frames = append(frames, NewTLCSyntheticStateStackFrame(tool, info, successor, width))
		successor = stateInfoState(info)
	}
	return frames
}

func syntheticTraceFramesFromChronological(tool *Tool, trace []*TLCStateInfo, successor *TLCStateMut) []*TLCSyntheticStateStackFrame {
	if len(trace) == 0 {
		return nil
	}
	width := len(strconv.Itoa(len(trace)))
	frames := make([]*TLCSyntheticStateStackFrame, 0, len(trace))
	for i := len(trace) - 1; i >= 0; i-- {
		info := trace[i]
		frames = append(frames, NewTLCSyntheticStateStackFrame(tool, info, successor, width))
		successor = stateInfoState(info)
	}
	return frames
}

type TLCActionStackFrame struct {
	TLCStateStackFrame
	Action *Action
}

func (f *TLCActionStackFrame) require() *TLCActionStackFrame {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f
}

func NewTLCActionStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, predecessor *TLCStateMut, action *Action, state *TLCStateMut, exception error) *TLCActionStackFrame {
	frame := &TLCActionStackFrame{
		TLCStateStackFrame: *NewTLCStateStackFrame(parent, node, ctxt, tool, state, exception),
		Action:             action,
	}
	if predecessor == nil || action == nil || state == nil {
		panic(NewAssertionError())
	}
	if predecessor.Level() != state.Level() && state.Predecessor() == nil && !state.IsInitial() {
		panic(NewAssertionError())
	}
	return frame
}

func NewTLCActionStackFrameNoException(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, predecessor *TLCStateMut, action *Action, state *TLCStateMut) *TLCActionStackFrame {
	return NewTLCActionStackFrame(parent, node, ctxt, tool, predecessor, action, state, nil)
}

func (f *TLCActionStackFrame) GetS() *TLCStateMut {
	f.require()
	return f.ownedState().Predecessor()
}

func (f *TLCActionStackFrame) GetT() *TLCStateMut {
	f.require()
	return f.ownedState()
}

func (f *TLCActionStackFrame) AddT() bool {
	f.require()
	return false
}

func (f *TLCActionStackFrame) ToRecordValue() *RecordValue {
	state := f.GetT()
	if state == nil {
		return EmptyRecord
	}
	return debugStateRecordValue(state, f.GetS())
}

func (f *TLCActionStackFrame) ToVariable(rnd *rand.Rand) *DebugTLCVariable {
	state := f.GetT()
	name := "Action"
	if state != nil {
		name = fmt.Sprintf("%d: %s", state.Level(), debugActionLocation(state.GetAction()))
	}
	return f.getStateAsVariable(f.ToRecordValue(), name, rnd)
}

func (f *TLCActionStackFrame) GetVariables(ref int, rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	ref = int(int32(ref))
	if ref == f.ownedStateID() {
		return f.debugVariables(func() []*DebugTLCVariable { return []*DebugTLCVariable{f.ToVariable(rnd)} })
	}
	if ref == int(int32(f.ownedStateID())+1) {
		return f.TraceVariables(rnd)
	}
	return f.getVariablesWithStates(ref, rnd, f.GetS(), f.GetT(), false)
}

func (f *TLCActionStackFrame) GetScopes() []TLCScope {
	f.require()
	scopes := append([]TLCScope{}, f.TLCStackFrame.GetScopes()...)
	if f.HasScope() {
		scopes = append(scopes, TLCScope{Name: f.ScopeName(), VariablesReference: f.ownedStateID()})
	}
	scopes = append(scopes, TLCScope{Name: debugScopeTrace, VariablesReference: int(int32(f.ownedStateID()) + 1)})
	return scopes
}

func (f *TLCActionStackFrame) TraceVariables(rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	return f.TLCStateStackFrame.TraceVariables(rnd)
}

func (f *TLCActionStackFrame) ScopeName() string {
	f.require()
	return debugScopeAction
}

type TLCInitStatesStackFrame struct {
	TLCStackFrame
	Functor      *StateFunctor
	IDToStateMap map[int]*TLCStateMut
	StateID      int

	initFinalsCaptured bool
	functorOwner       *StateFunctor
	idToStateMapOwner  map[int]*TLCStateMut
	initStateIDOwner   int
}

func (f *TLCInitStatesStackFrame) require() *TLCInitStatesStackFrame {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f
}

func NewTLCInitStatesStackFrame(parent *TLCStackFrame, pred SemanticNode, con *Context, tool *Tool, functor *StateFunctor) *TLCInitStatesStackFrame {
	frame := &TLCInitStatesStackFrame{
		TLCStackFrame: *NewTLCStackFrameNoException(parent, pred, con, tool),
		Functor:       functor,
		IDToStateMap:  make(map[int]*TLCStateMut),
		StateID:       debugVariableReference(nil),
	}
	frame.captureInitFinals()
	return frame
}

func (f *TLCInitStatesStackFrame) ownedFunctor() *StateFunctor {
	if f.initFinalsCaptured {
		return f.functorOwner
	}
	return f.Functor
}

func (f *TLCInitStatesStackFrame) ownedIDToStateMap() map[int]*TLCStateMut {
	if f.initFinalsCaptured {
		return f.idToStateMapOwner
	}
	return f.IDToStateMap
}

func (f *TLCInitStatesStackFrame) ownedInitStateID() int {
	if f.initFinalsCaptured {
		return f.initStateIDOwner
	}
	return f.StateID
}

func (f *TLCInitStatesStackFrame) captureInitFinals() {
	f.functorOwner = f.Functor
	f.idToStateMapOwner = f.IDToStateMap
	f.initStateIDOwner = f.StateID
	f.initFinalsCaptured = true
}

func (f *TLCInitStatesStackFrame) GetStates() *SetOfStates {
	f.require()
	functor := f.ownedFunctor()
	if functor == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.IStateFunctor.getStates()" because "this.functor" is null`))
	}
	states := functor.GetStates()
	if states == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.util.SetOfStates.toSet()" because the return value of "tlc2.tool.IStateFunctor.getStates()" is null`))
	}
	// Java returns functor.getStates().toSet(). Keep the concrete carrier used
	// by Go callers while preserving conversion and failure effects.
	_ = states.ToSet()
	return states
}

func (f *TLCInitStatesStackFrame) GetStateVariables(rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	return f.debugVariables(func() []*DebugTLCVariable {
		states := sortedStatesByString(f.GetStates())
		width := len(strconv.Itoa(len(states)))
		out := make([]*DebugTLCVariable, 0, len(states))
		for i, state := range states {
			name := fmt.Sprintf("%d.%0*d: %s", state.Level(), width, i+1, debugActionLocation(state.GetAction()))
			variable := f.getStateAsVariable(NewRecordValueFromState(state), name, rnd)
			f.ownedIDToStateMap()[variable.VariablesReference] = state
			out = append(out, variable)
		}
		return out
	})
}

func (f *TLCInitStatesStackFrame) GetVariables(ref int, rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	ref = int(int32(ref))
	if ref == f.ownedInitStateID() {
		return f.GetStateVariables(rnd)
	}
	return f.TLCStackFrame.GetVariables(ref, rnd)
}

func (f *TLCInitStatesStackFrame) GetScopes() []TLCScope {
	f.require()
	scopes := append([]TLCScope{}, f.TLCStackFrame.GetScopes()...)
	scopes = append(scopes, TLCScope{Name: debugScopeInitials, VariablesReference: f.ownedInitStateID()})
	return scopes
}

func (f *TLCInitStatesStackFrame) MatchesBreakpoint(bp *TLCSourceBreakpoint) bool {
	f.require()
	if !f.TLCStackFrame.MatchesBreakpoint(bp) {
		return false
	}
	for _, state := range f.GetStates().ToSlice() {
		if bp.MatchesExpression(f.GetTool(), state, EmptyState, f.GetContext(), true) {
			return true
		}
	}
	return false
}

func (f *TLCInitStatesStackFrame) SelectStateByReference(ref int) (bool, error) {
	f.require()
	ref = int(int32(ref))
	state := f.ownedIDToStateMap()[ref]
	if state == nil {
		return false, nil
	}
	functor := f.ownedFunctor()
	if functor == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.IStateFunctor.setElement(tlc2.tool.TLCState)" because "this.functor" is null`))
	}
	_, err := functor.SetElement(state)
	return err == nil, err
}

type TLCSyntheticStateStackFrame struct {
	TLCStateStackFrame
	Successor *TLCStateMut

	successorCaptured bool
	successorOwner    *TLCStateMut
}

func (f *TLCSyntheticStateStackFrame) require() *TLCSyntheticStateStackFrame {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f
}

func NewTLCSyntheticStateStackFrame(tool *Tool, info *TLCStateInfo, successor *TLCStateMut, width int) *TLCSyntheticStateStackFrame {
	if info == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCStateInfo.getAction()" because "ti" is null`))
	}
	width = int(int32(width))
	action := info.GetAction()
	frame := NewTLCStateStackFrame(nil, action.ownedPred(), action.ownedContext(), tool, info.ownedState(), nil)
	level := info.ownedState().Level()
	formattedLevel, err := JavaFormatObjects("%0"+strconv.Itoa(width)+"d", level)
	if err != nil {
		panic(err)
	}
	if isNil(info.Info) {
		panic(NewNullPointerException(`Cannot invoke "Object.toString()" because "ti.info" is null`))
	}
	label := fmt.Sprint(info.Info)
	frame.Name = formattedLevel + ": " + label
	return &TLCSyntheticStateStackFrame{
		TLCStateStackFrame: *frame,
		Successor:          successor,
		successorCaptured:  true,
		successorOwner:     successor,
	}
}

func (f *TLCSyntheticStateStackFrame) ownedSuccessor() *TLCStateMut {
	if f.successorCaptured {
		return f.successorOwner
	}
	return f.Successor
}

func (f *TLCSyntheticStateStackFrame) GetT() *TLCStateMut {
	f.require()
	return f.ownedState()
}

func (f *TLCSyntheticStateStackFrame) GetSuccessor() *TLCStateMut {
	f.require()
	return f.ownedSuccessor()
}

type TLCNextStatesStackFrame struct {
	TLCStateStackFrame
	Action       *Action
	Functor      *NextStateFunctor
	IDToStateMap map[int]*TLCStateMut

	nextFinalsCaptured bool
	nextActionOwner    *Action
	nextFunctorOwner   *NextStateFunctor
	nextStateMapOwner  map[int]*TLCStateMut
}

func (f *TLCNextStatesStackFrame) require() *TLCNextStatesStackFrame {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f
}

func NewTLCNextStatesStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, state *TLCStateMut, functor *NextStateFunctor, action *Action) *TLCNextStatesStackFrame {
	frame := NewTLCStateStackFrameNoException(parent, node, ctxt, tool, state)
	frame.Name = fmt.Sprint(node)
	next := &TLCNextStatesStackFrame{
		TLCStateStackFrame: *frame,
		Action:             action,
		Functor:            functor,
		IDToStateMap:       make(map[int]*TLCStateMut),
	}
	next.captureNextFinals()
	return next
}

func (f *TLCNextStatesStackFrame) ownedNextAction() *Action {
	if f.nextFinalsCaptured {
		return f.nextActionOwner
	}
	return f.Action
}

func (f *TLCNextStatesStackFrame) ownedNextFunctor() *NextStateFunctor {
	if f.nextFinalsCaptured {
		return f.nextFunctorOwner
	}
	return f.Functor
}

func (f *TLCNextStatesStackFrame) ownedNextStateMap() map[int]*TLCStateMut {
	if f.nextFinalsCaptured {
		return f.nextStateMapOwner
	}
	return f.IDToStateMap
}

func (f *TLCNextStatesStackFrame) captureNextFinals() {
	f.nextActionOwner = f.Action
	f.nextFunctorOwner = f.Functor
	f.nextStateMapOwner = f.IDToStateMap
	f.nextFinalsCaptured = true
}

func (f *TLCNextStatesStackFrame) AddT() bool {
	f.require()
	return true
}

func (f *TLCNextStatesStackFrame) GetSuccessors() *SetOfStates {
	f.require()
	functor := f.ownedNextFunctor()
	if functor == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.INextStateFunctor.getStates()" because "this.fun" is null`))
	}
	states := functor.GetStates()
	if states == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.util.SetOfStates.toSet()" because the return value of "tlc2.tool.INextStateFunctor.getStates()" is null`))
	}
	// Java returns fun.getStates().toSet().  Keep the concrete Go return type
	// used by debugger callers, but perform that conversion for its hashing,
	// equality, iterator-reset, and failure effects.
	_ = states.ToSet()
	return states
}

func (f *TLCNextStatesStackFrame) HasScope() bool {
	return f.require().GetSuccessors().Size() > 0
}

func (f *TLCNextStatesStackFrame) ScopeName() string {
	f.require()
	return debugScopeSuccessors
}

func (f *TLCNextStatesStackFrame) GetStateVariables(rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	return []*DebugTLCVariable{f.ToVariable(rnd)}
}

func (f *TLCNextStatesStackFrame) GetSuccessorVariables(rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	return f.debugVariables(func() []*DebugTLCVariable {
		successors := sortedSuccessors(f.GetSuccessors())
		width := len(strconv.Itoa(len(successors)))
		out := make([]*DebugTLCVariable, 0, len(successors))
		for i, state := range successors {
			name := fmt.Sprintf("%d.%0*d: %s", state.Level(), width, i+1, debugActionLocation(state.GetAction()))
			variable := f.getStateAsVariable(NewRecordValueFromState(state), name, rnd)
			f.ownedNextStateMap()[variable.VariablesReference] = state
			out = append(out, variable)
		}
		return out
	})
}

func (f *TLCNextStatesStackFrame) GetVariables(ref int, rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	ref = int(int32(ref))
	if ref == f.ownedStateID() {
		return f.GetSuccessorVariables(rnd)
	}
	if ref == int(int32(f.ownedStateID())+1) {
		return f.TraceVariables(rnd)
	}
	return f.getVariablesWithStates(ref, rnd, f.GetS(), nil, false)
}

func (f *TLCNextStatesStackFrame) GetScopes() []TLCScope {
	f.require()
	scopes := append([]TLCScope{}, f.TLCStackFrame.GetScopes()...)
	if f.HasScope() {
		scopes = append(scopes, TLCScope{Name: f.ScopeName(), VariablesReference: f.ownedStateID()})
	}
	scopes = append(scopes, TLCScope{Name: debugScopeTrace, VariablesReference: int(int32(f.ownedStateID()) + 1)})
	return scopes
}

func (f *TLCNextStatesStackFrame) TraceVariables(rnd *rand.Rand) []*DebugTLCVariable {
	f.require()
	return f.debugVariables(func() []*DebugTLCVariable {
		simulation := CurrentSimulator() != nil
		var idToState map[int]*TLCStateMut
		if simulation {
			idToState = f.ownedNextStateMap()
		}
		return f.traceVariables(rnd, !simulation, simulation, idToState)
	})
}

func (f *TLCNextStatesStackFrame) MatchesBreakpoint(bp *TLCSourceBreakpoint) bool {
	f.require()
	if !f.TLCStackFrame.MatchesBreakpoint(bp) {
		return false
	}
	states := f.breakpointStates()
	if states.Size() == 0 {
		return bp.MatchesExpression(f.GetTool(), f.GetS(), EmptyState, f.GetContext(), true)
	}
	// The nonempty branch uses states.toSet().stream() in Java.
	_ = states.ToSet()
	successors := states.ToSlice()
	for _, successor := range successors {
		if bp.MatchesExpression(f.GetTool(), f.GetS(), successor, f.GetContext(), true) {
			return true
		}
	}
	return false
}

func (f *TLCNextStatesStackFrame) breakpointStates() *SetOfStates {
	functor := f.ownedNextFunctor()
	if functor == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.INextStateFunctor.getStates()" because "this.fun" is null`))
	}
	states := functor.GetStates()
	if states == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.util.SetOfStates.size()" because "<local3>" is null`))
	}
	return states
}

func (f *TLCNextStatesStackFrame) SelectStateByReference(ref int) (bool, error) {
	f.require()
	ref = int(int32(ref))
	state := f.ownedNextStateMap()[ref]
	if state == nil {
		return false, nil
	}
	functor := f.ownedNextFunctor()
	if functor == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.INextStateFunctor.setElement(tlc2.tool.TLCState)" because "this.fun" is null`))
	}
	_, err := functor.SetElement(state)
	return err == nil, err
}

func (f *TLCNextStatesStackFrame) StepInSelect() (bool, error) {
	f.require()
	return f.selectSuccessorByDistance(true)
}

func (f *TLCNextStatesStackFrame) StepOverSelect() (bool, error) {
	f.require()
	return f.selectSuccessorByDistance(false)
}

func (f *TLCNextStatesStackFrame) StepOutSelect() (bool, error) {
	f.require()
	predecessor := f.GetS().Predecessor()
	functor := f.ownedNextFunctor()
	if functor == nil {
		if predecessor == nil {
			panic(NewNullPointerException(`Cannot invoke "tlc2.tool.INextStateFunctor.halt()" because "this.fun" is null`))
		}
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.INextStateFunctor.setElement(tlc2.tool.TLCState)" because "this.fun" is null`))
	}
	if predecessor == nil {
		return functor.Halt(), nil
	}
	_, err := functor.SetElement(predecessor)
	return err == nil, err
}

func (f *TLCNextStatesStackFrame) selectSuccessorByDistance(minimum bool) (bool, error) {
	f.require()
	states := f.GetSuccessors().ToSlice()
	if len(states) == 0 {
		panic(NewNoSuchElementException("No value present"))
	}
	current := f.GetS()
	selected := states[0]
	selectedDistance := debugHammingDistance(current, selected)
	for _, state := range states[1:] {
		distance := debugHammingDistance(current, state)
		if minimum && distance < selectedDistance || !minimum && distance > selectedDistance {
			selected = state
			selectedDistance = distance
		}
	}
	functor := f.ownedNextFunctor()
	if functor == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.INextStateFunctor.setElement(tlc2.tool.TLCState)" because "this.fun" is null`))
	}
	_, err := functor.SetElement(selected)
	return err == nil, err
}

type TLCCapabilities struct {
	SupportsStepBack                  bool
	SupportsGotoState                 *bool
	SupportsEvaluateForHovers         bool
	SupportsTerminateRequest          bool
	SupportsExceptionInfoRequest      bool
	ExceptionBreakpointFilters        []TLCExceptionBreakpointFilter
	SupportsExceptionFilterOptions    bool
	SupportsHitConditionalBreakpoints bool
	SupportsConditionalBreakpoints    bool
	SupportsLogPoints                 bool
	SupportsValueFormattingOptions    bool
	SupportsSteppingGranularity       bool
	SupportsGotoTargetsRequest        bool
	SupportsDataBreakpoints           bool
	SupportsFunctionBreakpoints       bool
	SupportsInstructionBreakpoints    bool
	SupportsDisassembleRequest        bool
	SupportsClipboardContext          bool
}

func (c *TLCCapabilities) require() *TLCCapabilities {
	if c == nil {
		panic(NewNullPointerException())
	}
	return c
}

func (c *TLCCapabilities) GetSupportsGotoState() *bool {
	return c.require().SupportsGotoState
}

func (c *TLCCapabilities) SetSupportsGotoState(supported *bool) {
	c.require().SupportsGotoState = supported
}

func debuggerBoolean(value bool) *bool {
	return &value
}

var (
	TLCCapabilitiesStepBack   = TLCCapabilities{SupportsStepBack: true}
	TLCCapabilitiesNoStepBack = TLCCapabilities{SupportsStepBack: false}
)

type TLCStackTraceArguments struct {
	StartFrame *int
	Levels     *int
}

type TLCStackTraceResponse struct {
	StackFrames []*TLCStackFrame
	TotalFrames int
}

type TLCThread struct {
	ID   int
	Name string
}

type TLCThreadsResponse struct {
	Threads []TLCThread
}

type TLCSetVariableResponse struct{}

type GotoStateArgument struct {
	VariablesReference *int32
}

func NewGotoStateArgument() *GotoStateArgument { return &GotoStateArgument{} }

func (a *GotoStateArgument) require() *GotoStateArgument {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a
}

func (a *GotoStateArgument) GetVariablesReference() *int32 {
	return a.require().VariablesReference
}

func (a *GotoStateArgument) SetVariablesReference(ref any) *GotoStateArgument {
	a.require()
	switch value := ref.(type) {
	case nil:
		a.VariablesReference = nil
	case *int32:
		a.VariablesReference = value
	case int32:
		a.VariablesReference = &value
	case int:
		narrowed := int32(value)
		a.VariablesReference = &narrowed
	default:
		panic(NewIllegalArgumentException("variablesReference must be a boxed Java int"))
	}
	return a
}

type SourceLocation struct {
	Source      string
	BeginLine   int
	BeginColumn int
	EndLine     int
	EndColumn   int
}

var NullSourceLocation = SourceLocation{}

func NewSourceLocation(source string, beginLine int, beginColumn int, endLine int, endColumn int) SourceLocation {
	return SourceLocation{
		Source:      source,
		BeginLine:   int(int32(beginLine)),
		BeginColumn: int(int32(beginColumn)),
		EndLine:     int(int32(endLine)),
		EndColumn:   int(int32(endColumn)),
	}
}

func (l SourceLocation) IsNull() bool {
	return l.Source == "" && l.BeginLine == 0 && l.BeginColumn == 0 && l.EndLine == 0 && l.EndColumn == 0
}

func (l SourceLocation) String() string {
	if l.IsNull() {
		return "Unknown location"
	}
	if l.Source != "--unknown--" && l.BeginLine == 0 && l.BeginColumn == 0 && l.EndLine == 0 && l.EndColumn == 0 {
		return "In module " + l.Source
	}
	return fmt.Sprintf("line %d, col %d to line %d, col %d of module %s", l.BeginLine, l.BeginColumn, l.EndLine, l.EndColumn, l.Source)
}

func (l SourceLocation) Includes(other SourceLocation) bool {
	if l.Source != other.Source {
		return false
	}
	if l.BeginLine > other.BeginLine {
		return false
	}
	if l.BeginLine == other.BeginLine && l.BeginColumn > other.BeginColumn {
		return false
	}
	if l.EndLine < other.EndLine {
		return false
	}
	if l.EndLine == other.EndLine && l.EndColumn < other.EndColumn {
		return false
	}
	return true
}

type TLCExceptionBreakpointFilter struct {
	Filter               string
	Label                string
	Description          string
	ConditionDescription string
	Default              bool
	SupportsCondition    bool
}

type TLCExceptionBreakpointFilterOption struct {
	FilterID  string
	Condition string
}

type TLCSourceBreakpointRequest struct {
	Line         int
	Column       *int
	Condition    string
	LogMessage   string
	HitCondition string
}

type TLCBreakpoint struct {
	ID       int
	Source   string
	Module   string
	Line     int
	Column   *int
	Verified bool
	Message  string
}

type TLCScope struct {
	Name               string
	VariablesReference int
	PresentationHint   string
}

type DebugTLCVariable struct {
	Name                      string
	Type                      string
	Value                     string
	ValueNull                 bool // Java Variable.getValue may return null for a throwable message.
	VariablesReference        int
	VSCodeVariableMenuContext string
	TLCValue                  Value
}

func (v *DebugTLCVariable) GetValue() *string {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.ValueNull {
		return nil
	}
	return javaString(v.Value)
}

func NewDebugTLCVariable(name *UniqueString) *DebugTLCVariable {
	if name == nil {
		panic(NewNullPointerException())
	}
	return NewDebugTLCVariableName(name.String())
}

func NewDebugTLCVariableName(name string) *DebugTLCVariable {
	return (&DebugTLCVariable{}).SetName(name)
}

func NewDebugTLCVariableValue(value Value, rnd *rand.Rand) *DebugTLCVariable {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	return ValueToTLCVariable(value, NewDebugTLCVariableName(value.String()), rnd)
}

func (v *DebugTLCVariable) SetName(name string) *DebugTLCVariable {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.Name = name
	return v
}

func (v *DebugTLCVariable) SetValue(value string) {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.Value = value
	v.ValueNull = false
}

func (v *DebugTLCVariable) SetType(valueType string) {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.Type = valueType
}

func (v *DebugTLCVariable) SetVariablesReference(reference int) {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.VariablesReference = int(int32(reference))
}

func (v *DebugTLCVariable) SetVscodeVariableMenuContext(context string) {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.VSCodeVariableMenuContext = context
}

func (v *DebugTLCVariable) GetVscodeVariableMenuContext() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.VSCodeVariableMenuContext
}

func (v *DebugTLCVariable) SetInstance(value Value) *DebugTLCVariable {
	if v == nil {
		panic(NewNullPointerException())
	}
	if isNil(value) {
		value = nil
	}
	v.TLCValue = value
	return v
}

func (v *DebugTLCVariable) GetTLCValue() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.TLCValue
}

func (v *DebugTLCVariable) NewInstance(name string, value Value, rnd *rand.Rand) *DebugTLCVariable {
	if v == nil || isNil(value) {
		panic(NewNullPointerException())
	}
	variable := NewDebugTLCVariableName(name).SetInstance(value)
	if debugValueMayHaveNested(value) {
		variable.VariablesReference = debugVariableReference(rnd)
	}
	return ValueToTLCVariable(value, variable, rnd)
}

func (v *DebugTLCVariable) NewInstanceValue(value Value, rnd *rand.Rand) *DebugTLCVariable {
	if v == nil || isNil(value) {
		panic(NewNullPointerException())
	}
	return v.NewInstance(value.String(), value, rnd)
}

func (v *DebugTLCVariable) Nested(rnd *rand.Rand) []*DebugTLCVariable {
	if v == nil || isNil(v.TLCValue) {
		panic(NewNullPointerException())
	}
	return ValueGetTLCVariables(v.TLCValue, v, rnd)
}

func (v *DebugTLCVariable) GetNested(rnd *rand.Rand) []*DebugTLCVariable {
	return v.Nested(rnd)
}

func (v *DebugTLCVariable) Compare(other *DebugTLCVariable) int {
	if v == nil || other == nil {
		panic(NewNullPointerException())
	}
	if cmp := strings.Compare(v.Name, other.Name); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(v.Type, other.Type); cmp != 0 {
		return cmp
	}
	return strings.Compare(v.Value, other.Value)
}

func debugValueToVariable(variable *DebugTLCVariable, value Value, rnd *rand.Rand) *DebugTLCVariable {
	if variable == nil {
		variable = &DebugTLCVariable{}
	}
	if interval, ok := value.(*IntervalValue); ok {
		return ValueToTLCVariable(interval.ToSetEnum(), variable, rnd)
	}
	variable.SetInstance(value)
	if value == nil {
		return variable
	}
	variable.Type = debugValueTypeString(value)
	variable.Value = debugValueString(value)
	if debugValueMayHaveNested(value) && debugValueIsFinite(value) {
		variable.VariablesReference = debugVariableReference(rnd)
	}
	return variable
}

// ValueToTLCVariable represents Value.toTLCVariable and concrete overrides.
func ValueToTLCVariable(value Value, variable *DebugTLCVariable, rnd *rand.Rand) *DebugTLCVariable {
	if isNil(value) || variable == nil {
		panic(NewNullPointerException())
	}
	return debugValueToVariable(variable, value, rnd)
}

func debugValueMayHaveNested(value Value) bool {
	switch value.(type) {
	case *FcnRcdValue, *RecordValue, *CounterExample, *TupleValue:
		return true
	default:
		_, ok := value.(Enumerable)
		return ok
	}
}

func debugValueIsFinite(value Value) bool {
	if value == nil {
		return false
	}
	finite, err := value.IsFinite()
	return err == nil && finite
}

func debugValueString(value Value) string {
	if value == nil {
		return ""
	}
	return ValueToUnquotedString(value)
}

func debugValueTypeString(value Value) string {
	if value == nil {
		return ""
	}
	if debuggerValue, ok := value.(*DebuggerValue); ok {
		return debuggerValue.GetTypeString()
	}
	return ValueGetTypeString(value)
}

func debugVariableReference(rnd *rand.Rand) int {
	if rnd == nil {
		return rand.Intn(maxJavaInt-1) + 1
	}
	return rnd.Intn(maxJavaInt-1) + 1
}

func semanticNodeDebugID(node SemanticNode) int {
	if node == nil {
		return debugVariableReference(nil)
	}
	if node, ok := node.(interface{ GetUid() int32 }); ok {
		return int(node.GetUid())
	}
	if node, ok := node.(interface{ GetUID() int32 }); ok {
		return int(node.GetUID())
	}
	return int(FP64Hash(FP64NewString(fmt.Sprint(node))))
}

func semanticNodeDebugName(node SemanticNode, exception error) string {
	name := fmt.Sprint(node)
	if name == "" || name == "<nil>" {
		name = "<semantic node>"
	}
	if exception != nil {
		return "(Exception) " + name
	}
	return name
}

func symbolNodeDebugName(node *SymbolNode) string {
	if node == nil || node.Name == nil {
		return ""
	}
	return node.GetName().String()
}

func semanticNodeLevel(node SemanticNode) int {
	if getter, ok := node.(interface{ GetLevel() int }); ok {
		return getter.GetLevel()
	}
	return SemanticLevel(node)
}

func semanticNodeSame(left SemanticNode, right SemanticNode) bool {
	if left == nil || right == nil {
		return left == right
	}
	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)
	if leftValue.IsValid() && rightValue.IsValid() && leftValue.Type() == rightValue.Type() && leftValue.Type().Comparable() {
		return leftValue.Interface() == rightValue.Interface()
	}
	return fmt.Sprint(left) == fmt.Sprint(right)
}

func semanticNodeSourceLocation(node SemanticNode) (SourceLocation, bool) {
	if located, ok := node.(interface{ SourceLocation() SourceLocation }); ok {
		return located.SourceLocation(), true
	}
	if located, ok := node.(interface{ GetSourceLocation() SourceLocation }); ok {
		return located.GetSourceLocation(), true
	}
	return NullSourceLocation, false
}

func semanticWalk(node SemanticNode, visit func(SemanticNode) bool) {
	seen := make(map[uintptr]struct{})
	var walk func(SemanticNode)
	walk = func(current SemanticNode) {
		if current == nil {
			return
		}
		if key := semanticNodePointerKey(current); key != 0 {
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}
		}
		if visit != nil && !visit(current) {
			return
		}
		switch n := current.(type) {
		case *OpDefNode:
			walk(n.GetBody())
		case *LabelNode:
			walk(n.GetBody())
		case *OpApplNode:
			for _, arg := range n.Args {
				walk(arg)
			}
			for _, bound := range n.BdedQuantBounds {
				walk(bound)
			}
		case *LetInNode:
			for _, def := range n.Lets {
				walk(def)
			}
			for _, binding := range n.Bindings {
				if child, ok := binding.Value.(SemanticNode); ok {
					walk(child)
				}
			}
			walk(n.Body)
		case *SubstInNode:
			for _, subst := range n.Substs {
				walk(subst.GetExpr())
			}
			walk(n.GetBody())
		case *APSubstInNode:
			for _, subst := range n.Substs {
				walk(subst.GetExpr())
			}
			walk(n.GetBody())
		case *PossibleTrackNode:
			walk(n.Pred)
		case *ThmOrAssumpDefNode:
			walk(n.GetBody())
		case *LiveExprNode:
			walk(n.Body)
			for _, body := range n.Bodies {
				walk(body)
			}
			walk(n.Pred)
		}
	}
	walk(node)
}

func semanticNodePointerKey(node SemanticNode) uintptr {
	value := reflect.ValueOf(node)
	if !value.IsValid() {
		return 0
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		if value.IsNil() {
			return 0
		}
		return value.Pointer()
	default:
		return 0
	}
}

func debugStateCopy(state *TLCStateMut) *TLCStateMut {
	if state == nil {
		return nil
	}
	return state.DeepCopy()
}

func debugStateAsVariable(state *TLCStateMut, record Value, name string, rnd *rand.Rand) *DebugTLCVariable {
	variable := NewDebugTLCVariableName(name).SetInstance(record)
	variable.SetVscodeVariableMenuContext("state")
	variable.Type = debugStateVariableType(state)
	if debugValueMayHaveNested(variable.TLCValue) {
		variable.VariablesReference = debugVariableReference(rnd)
	}
	variable.Value = debugValueString(variable.TLCValue)
	return variable
}

func debugStateVariableType(state *TLCStateMut) (typ string) {
	typ = "State"
	defer func() {
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); ok && debugValueFailure(err) {
				return
			}
			panic(failure)
		}
	}()
	if state != nil && state.AllAssigned() {
		typ = fmt.Sprintf("FP64: %d", int64(state.FingerPrint()))
	}
	return typ
}

func debugValueFailure(err error) bool {
	if _, ok := err.(*FingerprintException); ok {
		return true
	}
	return isJavaEvalOrRuntimeException(err)
}

func debugUnlazy(lazy *LazyValue, tool *Tool, s, t *TLCStateMut, catchNull bool) (value Value, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			if caught, ok := failure.(error); ok && (debugValueFailure(caught) || (catchNull && isJavaNullPointerException(caught))) {
				err = caught
				return
			}
			panic(failure)
		}
	}()
	return lazy.Eval(tool, s, t)
}

func stateInfoState(info *TLCStateInfo) *TLCStateMut {
	if info == nil {
		return nil
	}
	return info.ownedState()
}

func debugActionLocation(action *Action) string {
	if action == nil {
		return "<???>"
	}
	location := action.GetLocation()
	if location == "" {
		return "<???>"
	}
	return location
}

func debugStateRecordValue(state *TLCStateMut, predecessor *TLCStateMut) *RecordValue {
	if state == nil {
		return EmptyRecord
	}
	if predecessor != nil {
		return NewRecordValueFromStates(predecessor, state, DebuggerNotEvaluatedValue)
	}
	return NewRecordValueFromState(state, DebuggerNotEvaluatedValue)
}

func sortedStatesByString(set *SetOfStates) []*TLCStateMut {
	if set == nil {
		return nil
	}
	states := set.ToSlice()
	sort.Slice(states, func(i, j int) bool {
		return javaStringCompareTo(states[i].String(), states[j].String()) < 0
	})
	return states[:compactSortedStates(states, func(left, right *TLCStateMut) bool {
		return javaStringCompareTo(left.String(), right.String()) == 0
	})]
}

func sortedSuccessors(set *SetOfStates) []*TLCStateMut {
	if set == nil {
		return nil
	}
	states := set.ToSlice()
	sort.Slice(states, func(i, j int) bool {
		left := debugActionLocation(states[i].GetAction())
		right := debugActionLocation(states[j].GetAction())
		if cmp := javaStringCompareTo(left, right); cmp != 0 {
			return cmp < 0
		}
		return javaStringCompareTo(states[i].String(), states[j].String()) < 0
	})
	return states[:compactSortedStates(states, func(left, right *TLCStateMut) bool {
		return javaStringCompareTo(debugActionLocation(left.GetAction()), debugActionLocation(right.GetAction())) == 0 &&
			javaStringCompareTo(left.String(), right.String()) == 0
	})]
}

func compactSortedStates(states []*TLCStateMut, equal func(*TLCStateMut, *TLCStateMut) bool) int {
	if len(states) == 0 {
		return 0
	}
	write := 1
	for read := 1; read < len(states); read++ {
		if !equal(states[write-1], states[read]) {
			states[write] = states[read]
			write++
		}
	}
	return write
}

func debugHammingDistance(left *TLCStateMut, right *TLCStateMut) int {
	if left == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCState.toString()" because "s1" is null`))
	}
	leftText := left.String()
	if right == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCState.toString()" because "s2" is null`))
	}
	rightText := right.String()
	leftUnits := javaStringUTF16(leftText)
	rightUnits := javaStringUTF16(rightText)
	minLen := len(leftUnits)
	if len(rightUnits) < minLen {
		minLen = len(rightUnits)
	}
	distance := absInt(len(leftUnits) - len(rightUnits))
	for i := 0; i < minLen; i++ {
		if leftUnits[i] != rightUnits[i] {
			distance++
		}
	}
	return distance
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// ValueGetTLCVariables represents Value.getTLCVariables and concrete overrides.
func ValueGetTLCVariables(value Value, prototype *DebugTLCVariable, rnd *rand.Rand) []*DebugTLCVariable {
	if isNil(value) || prototype == nil {
		panic(NewNullPointerException())
	}
	switch v := value.(type) {
	case *TupleValue:
		return debugTupleVariables(v, prototype, rnd)
	case *RecordValue:
		return debugRecordVariables(v, prototype, rnd)
	case *CounterExample:
		return debugRecordVariables(asRecordValue(v), prototype, rnd)
	case *FcnRcdValue:
		return debugFunctionVariables(v, prototype, rnd)
	case *SetEnumValue:
		return debugSetVariables(v, prototype, rnd)
	default:
		return debugEnumerableVariables(v, prototype, rnd)
	}
}

func debugTupleVariables(value *TupleValue, prototype *DebugTLCVariable, rnd *rand.Rand) []*DebugTLCVariable {
	if value == nil {
		return nil
	}
	out := make([]*DebugTLCVariable, 0, len(value.GetElems()))
	width := len(strconv.Itoa(len(value.GetElems())))
	format := "%0" + strconv.Itoa(width) + "d"
	for i, elem := range value.GetElems() {
		nested := prototype.NewInstance(fmt.Sprintf(format, i+1), elem, rnd)
		nested.Value = valueStringOrEmpty(elem)
		nested.Type = debugValueTypeString(elem)
		out = append(out, nested)
	}
	return out
}

func debugRecordVariables(value *RecordValue, prototype *DebugTLCVariable, rnd *rand.Rand) []*DebugTLCVariable {
	if value == nil {
		return nil
	}
	out := make([]*DebugTLCVariable, 0, len(value.ownedValues()))
	for i, name := range value.ownedNames() {
		field := ""
		if name != nil {
			field = name.String()
		}
		var elem Value
		if i < len(value.ownedValues()) {
			elem = value.ownedValues()[i]
		}
		nested := prototype.NewInstance(field, elem, rnd)
		nested.Value = valueStringOrEmpty(elem)
		nested.Type = debugValueTypeString(elem)
		out = append(out, nested)
	}
	return out
}

func debugFunctionVariables(value *FcnRcdValue, prototype *DebugTLCVariable, rnd *rand.Rand) []*DebugTLCVariable {
	if value == nil {
		return nil
	}
	domain := debugFunctionDomainValues(value)
	out := make([]*DebugTLCVariable, 0, len(value.ownedValues()))
	for i, elem := range value.ownedValues() {
		name := strconv.Itoa(i + 1)
		if i < len(domain) && domain[i] != nil {
			name = domain[i].String()
		}
		nested := prototype.NewInstance(name, elem, rnd)
		nested.Value = valueStringOrEmpty(elem)
		nested.Type = debugValueTypeString(elem)
		out = append(out, nested)
	}
	return out
}

func debugFunctionDomainValues(value *FcnRcdValue) []Value {
	if value == nil {
		return nil
	}
	if value.ownedIntv() == nil {
		return append([]Value(nil), value.ownedDomain()...)
	}
	if value.ownedIntv().High < value.ownedIntv().Low {
		return nil
	}
	out := make([]Value, 0, int(value.ownedIntv().High-value.ownedIntv().Low+1))
	for i := value.ownedIntv().Low; i <= value.ownedIntv().High; i++ {
		out = append(out, NewIntValue(i))
	}
	return out
}

func debugSetVariables(value *SetEnumValue, prototype *DebugTLCVariable, rnd *rand.Rand) []*DebugTLCVariable {
	if value == nil {
		return nil
	}
	out := make([]*DebugTLCVariable, 0, value.Elems.Len())
	enum := value.Elements()
	for {
		elem := enum.NextElement()
		if elem == nil {
			break
		}
		name := elem.String()
		nested := prototype.NewInstance(name, elem, rnd)
		nested.Name = name
		nested.Value = valueStringOrEmpty(elem)
		out = append(out, nested)
	}
	return out
}

func debugEnumerableVariables(value Value, prototype *DebugTLCVariable, rnd *rand.Rand) []*DebugTLCVariable {
	enumValue, ok := value.(Enumerable)
	if !ok || !debugValueIsFinite(value) {
		return nil
	}
	enum := enumValue.Elements()
	var out []*DebugTLCVariable
	ValueEnumerationForEach(enum, func(elem Value) {
		out = append(out, ValueToTLCVariable(elem, prototype.NewInstanceValue(elem, rnd), rnd))
	})
	return out
}

func valueStringOrEmpty(value Value) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func sortedDistinctDebugVariables(variables []*DebugTLCVariable) []*DebugTLCVariable {
	if len(variables) == 0 {
		return nil
	}
	sort.SliceStable(variables, func(i, j int) bool {
		return variables[i].Compare(variables[j]) < 0
	})
	out := variables[:0]
	for _, variable := range variables {
		if variable == nil {
			continue
		}
		if len(out) == 0 || out[len(out)-1].Compare(variable) != 0 {
			out = append(out, variable)
		}
	}
	return out
}

type TLCSourceBreakpoint struct {
	Line               int
	Column             *int
	Condition          string
	LogMessage         string
	HitCondition       string
	Hits               int
	Location           SourceLocation
	ConditionOp        *OpDefNode
	ConditionException error

	finalsCaptured bool
	hitsOwner      int
	locationOwner  SourceLocation
}

func (b *TLCSourceBreakpoint) require() *TLCSourceBreakpoint {
	if b == nil {
		panic(NewNullPointerException())
	}
	return b
}

func NewTLCSourceBreakpoint(condition string, op *OpDefNode, conditionErr error) *TLCSourceBreakpoint {
	if isNil(conditionErr) {
		conditionErr = nil
	}
	breakpoint := &TLCSourceBreakpoint{
		Line:               NullSourceLocation.BeginLine,
		Condition:          condition,
		Location:           NullSourceLocation,
		ConditionOp:        op,
		ConditionException: conditionErr,
	}
	breakpoint.captureFinals()
	return breakpoint
}

func NewTLCSourceBreakpointAt(module string, line int, column *int, condition string, logMessage string, hitCondition string, op *OpDefNode, conditionErr error) *TLCSourceBreakpoint {
	if isNil(conditionErr) {
		conditionErr = nil
	}
	line = int(int32(line))
	col := 1
	var storedColumn *int
	if column != nil {
		col = int(int32(*column))
		storedColumn = &col
	}
	breakpoint := &TLCSourceBreakpoint{
		Line:               line,
		Column:             storedColumn,
		Condition:          condition,
		LogMessage:         logMessage,
		HitCondition:       hitCondition,
		Hits:               parseBreakpointHits(hitCondition),
		Location:           NewSourceLocation(module, int(int32(line)+1), col, line, int(int32(col)+1)),
		ConditionOp:        op,
		ConditionException: conditionErr,
	}
	breakpoint.captureFinals()
	return breakpoint
}

func (b *TLCSourceBreakpoint) GetHits() int {
	b.require()
	if b.finalsCaptured {
		return b.hitsOwner
	}
	return b.Hits
}

func (b *TLCSourceBreakpoint) IsInline() bool {
	b.require()
	return b.GetColumnAsInt() == -1
}

func (b *TLCSourceBreakpoint) GetColumnAsInt() int {
	b.require()
	if b.Column == nil {
		return -1
	}
	return *b.Column
}

func (b *TLCSourceBreakpoint) GetLocation() SourceLocation {
	b.require()
	if b.finalsCaptured {
		return b.locationOwner
	}
	return b.Location
}

func (b *TLCSourceBreakpoint) GetConditionException() error {
	failure := b.require().ConditionException
	if isNil(failure) {
		return nil
	}
	return failure
}

func (b *TLCSourceBreakpoint) MatchesExpression(tool *Tool, s *TLCStateMut, t *TLCStateMut, c *Context, fire bool) bool {
	b.require()
	if b.ConditionOp == nil {
		return fire
	}
	ctxt := EmptyContext
	params := b.ConditionOp.GetParams()
	if len(params) > 0 && c == nil {
		panic(NewNullPointerException())
	}
	for _, param := range params {
		value := c.LookupFunc(func(sym *SymbolNode) bool {
			if sym == nil || sym.Name == nil {
				panic(NewNullPointerException())
			}
			if param == nil {
				panic(NewNullPointerException())
			}
			return sym.Name == param.Name
		})
		ctxt = ctxt.Cons(param, value)
	}
	eval, err := tool.NoDebug().Eval(b.ConditionOp.GetBody(), ctxt, s, t, EvalClear)
	if err != nil {
		if debugValueFailure(err) {
			return fire
		}
		panic(err)
	}
	if boolValue, ok := eval.(*BoolValue); ok {
		fire = fire && boolValue.Val
	}
	return fire
}

func (b *TLCSourceBreakpoint) MatchesLocation(loc SourceLocation) bool {
	b.require()
	if b.GetLocation().IsNull() {
		return true
	}
	return b.Line == loc.BeginLine && b.GetColumnAsInt() <= loc.BeginColumn
}

func (b *TLCSourceBreakpoint) captureFinals() {
	b.hitsOwner = b.Hits
	b.locationOwner = b.Location
	b.finalsCaptured = true
}

func parseBreakpointHits(hitCondition string) int {
	if hitCondition == "" {
		return 0
	}
	hits, ok := javaParseDecimalInt(hitCondition)
	if !ok {
		return 0
	}
	return int(hits)
}

func debugModuleNameFromSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}
	if idx := strings.LastIndexAny(source, `/\`); idx >= 0 {
		source = source[idx+1:]
	}
	return strings.TrimSuffix(source, ".tla")
}

func (d *TLCDebugger) newConditionalBreakpointLocked(condition string) *TLCSourceBreakpoint {
	if JavaStringIsBlank(condition) {
		condition = "TRUE"
	}
	op, err := d.breakpointConditionOpLocked(condition, NullSourceLocation)
	return NewTLCSourceBreakpoint(condition, op, err)
}

func (d *TLCDebugger) breakpointConditionOpLocked(condition string, location SourceLocation) (*OpDefNode, error) {
	if JavaStringIsBlank(condition) {
		return nil, nil
	}
	if d != nil && d.Tool != nil {
		if !location.IsNull() {
			return d.Tool.ParseDebuggerExpression(location, condition, d.Tool.GetModule(location.Source))
		}
		return d.Tool.ParseDebuggerExpression(location, condition)
	}
	return nil, fmt.Errorf("debug breakpoint expression parsing is not configured: %s", condition)
}

// Factory override corresponds to TLCDebugger.Factory.OVERRIDE. Install it
// before starting TLC, as the original debugger harness does.
var TLCDebuggerFactoryOverride *TLCDebugger

func TLCDebuggerFactoryGetInstance(_ int, suspend bool, halt bool) *TLCDebugger {
	if TLCDebuggerFactoryOverride != nil {
		return TLCDebuggerFactoryOverride
	}
	debugger := NewTLCDebugger(nil)
	if !suspend {
		debugger.Step = DebugStepCommandContinue
	}
	debugger.HaltExp, debugger.HaltInv = halt, halt
	return debugger
}

type TLCStoppedEvent struct {
	ThreadID int
	Reason   string
	Text     string
}

type TLCDebugger struct {
	mu                sync.Mutex
	resume            *sync.Cond
	stoppedEvents     chan TLCStoppedEvent
	capabilitiesEvent func(TLCCapabilities)
	Tool              *Tool
	Granularity       DebugGranularity
	Direction         DebugStepDirection
	Step              DebugStep
	Stack             []*TLCDebuggerFrame
	SourceFrame       *TLCStackFrame
	Breakpoints       *InsMap[string, []*TLCSourceBreakpoint]
	HaltSpec          *TLCSourceBreakpoint
	HaltUnsat         *TLCSourceBreakpoint
	HaltExp           bool
	HaltInv           bool
	ExecutionIsHalted bool
	Paused            bool

	breakpointsCaptured bool
	breakpointsOwner    *InsMap[string, []*TLCSourceBreakpoint]
	stackCaptured       bool
	stackOwner          []*TLCDebuggerFrame
}

func (d *TLCDebugger) require() *TLCDebugger {
	if d == nil {
		panic(NewNullPointerException())
	}
	return d
}

type TLCDebuggerFrame struct {
	Base      *TLCStackFrame
	State     *TLCStateStackFrame
	Action    *TLCActionStackFrame
	Init      *TLCInitStatesStackFrame
	Next      *TLCNextStatesStackFrame
	Synthetic *TLCSyntheticStateStackFrame
}

func NewDebuggerBaseFrame(frame *TLCStackFrame) *TLCDebuggerFrame {
	return &TLCDebuggerFrame{Base: frame}
}

func NewDebuggerStateFrame(frame *TLCStateStackFrame) *TLCDebuggerFrame {
	out := &TLCDebuggerFrame{State: frame}
	if frame != nil {
		out.Base = &frame.TLCStackFrame
	}
	return out
}

func NewDebuggerActionFrame(frame *TLCActionStackFrame) *TLCDebuggerFrame {
	out := &TLCDebuggerFrame{Action: frame}
	if frame != nil {
		out.State = &frame.TLCStateStackFrame
		out.Base = &frame.TLCStackFrame
	}
	return out
}

func NewDebuggerInitFrame(frame *TLCInitStatesStackFrame) *TLCDebuggerFrame {
	out := &TLCDebuggerFrame{Init: frame}
	if frame != nil {
		out.Base = &frame.TLCStackFrame
	}
	return out
}

func NewDebuggerNextFrame(frame *TLCNextStatesStackFrame) *TLCDebuggerFrame {
	out := &TLCDebuggerFrame{Next: frame}
	if frame != nil {
		out.State = &frame.TLCStateStackFrame
		out.Base = &frame.TLCStackFrame
	}
	return out
}

func NewDebuggerSyntheticFrame(frame *TLCSyntheticStateStackFrame) *TLCDebuggerFrame {
	out := &TLCDebuggerFrame{Synthetic: frame}
	if frame != nil {
		out.State = &frame.TLCStateStackFrame
		out.Base = &frame.TLCStackFrame
	}
	return out
}

func (f *TLCDebuggerFrame) MatchesNode(node SemanticNode) bool {
	return f != nil && f.Base != nil && f.Base.MatchesNode(node)
}

func (f *TLCDebuggerFrame) MatchesBreakpoint(bp *TLCSourceBreakpoint) bool {
	if f == nil {
		return false
	}
	if f.Init != nil {
		return f.Init.MatchesBreakpoint(bp)
	}
	if f.Next != nil {
		return f.Next.MatchesBreakpoint(bp)
	}
	if f.Action != nil {
		return f.Action.MatchesBreakpoint(bp)
	}
	if f.Synthetic != nil {
		return f.Synthetic.matchesStateBreakpoint(bp, f.Synthetic.GetS(), f.Synthetic.GetT())
	}
	if f.State != nil {
		return f.State.MatchesBreakpoint(bp)
	}
	return f.Base != nil && f.Base.MatchesBreakpoint(bp)
}

func (f *TLCDebuggerFrame) GetScopes() []TLCScope {
	if f == nil {
		return nil
	}
	switch {
	case f.Next != nil:
		return f.Next.GetScopes()
	case f.Init != nil:
		return f.Init.GetScopes()
	case f.Action != nil:
		return f.Action.GetScopes()
	case f.Synthetic != nil:
		return f.Synthetic.GetScopes()
	case f.State != nil:
		return f.State.GetScopes()
	case f.Base != nil:
		return f.Base.GetScopes()
	default:
		return nil
	}
}

func (f *TLCDebuggerFrame) GetVariables(ref int, rnd *rand.Rand) []*DebugTLCVariable {
	if f == nil {
		return nil
	}
	ref = int(int32(ref))
	switch {
	case f.Next != nil:
		return f.Next.GetVariables(ref, rnd)
	case f.Init != nil:
		return f.Init.GetVariables(ref, rnd)
	case f.Action != nil:
		return f.Action.GetVariables(ref, rnd)
	case f.Synthetic != nil:
		return f.Synthetic.GetVariables(ref, rnd)
	case f.State != nil:
		return f.State.GetVariables(ref, rnd)
	case f.Base != nil:
		return f.Base.GetVariables(ref, rnd)
	default:
		return nil
	}
}

func (f *TLCDebuggerFrame) stateTraceFrame() *TLCStateStackFrame {
	if f == nil {
		return nil
	}
	switch {
	case f.Next != nil:
		return &f.Next.TLCStateStackFrame
	case f.Action != nil:
		return &f.Action.TLCStateStackFrame
	case f.Synthetic != nil:
		return &f.Synthetic.TLCStateStackFrame
	case f.State != nil:
		return f.State
	default:
		return nil
	}
}

func (f *TLCDebuggerFrame) Handle(debugger *TLCDebugger) bool {
	if f == nil {
		return false
	}
	if f.Next != nil {
		return true
	}
	if f.Init != nil && debugger != nil {
		return debugger.stackLen() == 1
	}
	return false
}

func NewTLCDebugger(tool *Tool) *TLCDebugger {
	d := &TLCDebugger{
		Tool:          tool,
		Granularity:   DebugGranularityFormula,
		Direction:     DebugStepContinue,
		Step:          DebugStepCommandIn,
		Stack:         make([]*TLCDebuggerFrame, 0, 64),
		Breakpoints:   NewInsMap[string, []*TLCSourceBreakpoint](),
		HaltExp:       true,
		HaltInv:       true,
		stoppedEvents: make(chan TLCStoppedEvent, 1),
	}
	d.breakpointsOwner = d.Breakpoints
	d.breakpointsCaptured = true
	d.stackOwner = d.Stack
	d.stackCaptured = true
	d.resume = sync.NewCond(&d.mu)
	return d
}

func (d *TLCDebugger) ownedBreakpoints() *InsMap[string, []*TLCSourceBreakpoint] {
	if d.breakpointsCaptured {
		return d.breakpointsOwner
	}
	return d.Breakpoints
}

func debuggerStacksShareStorage(left, right []*TLCDebuggerFrame) bool {
	if cap(left) == 0 || cap(right) == 0 {
		return false
	}
	return &left[:1][0] == &right[:1][0]
}

func (d *TLCDebugger) ownedStack() []*TLCDebuggerFrame {
	if !d.stackCaptured {
		return d.Stack
	}
	// Package-local translated tests mutate the final Java list through the
	// exported slice mirror. Accept a changed length while the backing store is
	// still the retained list, but ignore replacement by an unrelated slice.
	if debuggerStacksShareStorage(d.stackOwner, d.Stack) {
		d.stackOwner = d.Stack
	}
	return d.stackOwner
}

func (d *TLCDebugger) setOwnedStack(stack []*TLCDebuggerFrame) {
	if d.stackCaptured {
		d.stackOwner = stack
	}
	d.Stack = stack
}

func (d *TLCDebugger) stackLen() int {
	if d == nil {
		return 0
	}
	return len(d.ownedStack())
}

func (d *TLCDebugger) StoppedEvents() <-chan TLCStoppedEvent {
	return d.require().stoppedEvents
}

// SetCapabilitiesEvent installs the native DAP transport boundary used by
// state-selection frames to publish their temporary step-back capability.
func (d *TLCDebugger) SetCapabilitiesEvent(send func(TLCCapabilities)) {
	d.require()
	d.mu.Lock()
	d.capabilitiesEvent = send
	d.mu.Unlock()
}

func (d *TLCDebugger) sendCapabilitiesLocked(capabilities TLCCapabilities) {
	if d.capabilitiesEvent != nil {
		d.capabilitiesEvent(capabilities)
	}
}

func (t *Tool) AttachDebugger(port int, suspend bool, halt bool) *Tool {
	if t == nil {
		return nil
	}
	t.DebugPort = port
	t.DebugSuspend = suspend
	t.DebugHalt = halt
	debugger := t.ownedDebugger()
	if debugger == nil {
		t.Debugger = TLCDebuggerFactoryGetInstance(port, suspend, halt).SetTool(t)
		debugger = t.Debugger
	} else {
		debugger.SetTool(t)
	}
	if t.GetMode() != ModeSimulation {
		t.SetMode(ModeDebugger)
	}
	return t.installDebugTool()
}

func (d *TLCDebugger) SetTool(tool *Tool) *TLCDebugger {
	d.require()
	d.Tool = tool
	return d
}

func (d *TLCDebugger) SetGranularity(granularity DebugGranularity) {
	d.require().Granularity = granularity
}

func (d *TLCDebugger) GetGranularity() DebugGranularity {
	return d.require().Granularity
}

func (d *TLCDebugger) InitializeCapabilities() TLCCapabilities {
	d.require()
	filters := d.ExceptionBreakpointFilters()
	return TLCCapabilities{
		SupportsGotoState:                 debuggerBoolean(CurrentSimulator() != nil),
		SupportsEvaluateForHovers:         true,
		SupportsTerminateRequest:          true,
		SupportsExceptionInfoRequest:      false,
		ExceptionBreakpointFilters:        append([]TLCExceptionBreakpointFilter(nil), filters...),
		SupportsExceptionFilterOptions:    true,
		SupportsHitConditionalBreakpoints: true,
		SupportsConditionalBreakpoints:    true,
		SupportsLogPoints:                 false,
		SupportsStepBack:                  true,
		SupportsValueFormattingOptions:    false,
		SupportsSteppingGranularity:       false,
		SupportsGotoTargetsRequest:        false,
		SupportsDataBreakpoints:           false,
		SupportsFunctionBreakpoints:       false,
		SupportsInstructionBreakpoints:    false,
		SupportsDisassembleRequest:        false,
		SupportsClipboardContext:          true,
	}
}

func (d *TLCDebugger) ExceptionBreakpointFilters() []TLCExceptionBreakpointFilter {
	d.require()
	return d.ExceptionBreakpointFiltersForWorkers(NumWorkers())
}

func (d *TLCDebugger) ExceptionBreakpointFiltersForWorkers(workers int) []TLCExceptionBreakpointFilter {
	d.require()
	haltExp := false
	haltInv := false
	haltSpec := false
	haltUnsat := false
	d.mu.Lock()
	haltExp = d.HaltExp
	haltInv = d.HaltInv
	haltSpec = d.HaltSpec != nil
	haltUnsat = d.HaltUnsat != nil
	d.mu.Unlock()
	filters := []TLCExceptionBreakpointFilter{
		{
			Filter:      TLCExceptionBreakpointsFilter,
			Label:       "Halt (break) on exceptions",
			Description: "TLC will halt when it encounters a silly expression",
			Default:     haltExp,
		},
		{
			Filter:               TLCUnsatisfiedBreakpointsFilter,
			Label:                "Halt (break) on unsatisfied",
			Description:          "TLC will halt when a successor state does not satisfy the next-state relation.",
			ConditionDescription: "A constant, state, or action level formula",
			Default:              haltUnsat,
			SupportsCondition:    true,
		},
		{
			Filter:               TLCSpecBreakpointsFilter,
			Label:                "Halt (break) after Init and Next",
			Description:          "TLC will halt after initial- and next-states have been generated.",
			ConditionDescription: "Init: constant or state formula. Next: constant, state, or action-level formula.",
			Default:              haltSpec,
			SupportsCondition:    true,
		},
		{
			Filter:      TLCInvariantBreakpointsFilter,
			Label:       "Halt (break) on violations",
			Description: "TLC will halt when an invariant is violated.",
			Default:     haltInv,
		},
	}
	if workers > 1 {
		warning := TLCExceptionBreakpointFilter{
			Filter:            TLCMultiWorkerWarningFilter,
			Label:             "MULTIPLE WORKER WARNING: Breakpoints only fire for one worker (use -workers 1)",
			Description:       "The debugger only attaches to one worker thread. With multiple workers, the breakpoints in this list will only fire if they occur in the attached worker. Breakpoint triggers in other workers will be missed. To consistently catch all breakpoint triggers, restart TLC with -workers 1.",
			SupportsCondition: false,
		}
		filters = append([]TLCExceptionBreakpointFilter{warning}, filters...)
	}
	return filters
}

func (d *TLCDebugger) SetExceptionBreakpointFilters(filterIDs []string) *TLCDebugger {
	d.require()
	options := make([]TLCExceptionBreakpointFilterOption, 0, len(filterIDs))
	for _, id := range filterIDs {
		options = append(options, TLCExceptionBreakpointFilterOption{FilterID: id})
	}
	return d.SetExceptionBreakpoints(options)
}

func (d *TLCDebugger) SetExceptionBreakpoints(options []TLCExceptionBreakpointFilterOption) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	d.HaltExp = false
	d.HaltInv = false
	d.HaltSpec = nil
	d.HaltUnsat = nil
	for _, option := range options {
		switch option.FilterID {
		case TLCExceptionBreakpointsFilter:
			d.HaltExp = true
		case TLCInvariantBreakpointsFilter:
			d.HaltInv = true
		case TLCSpecBreakpointsFilter:
			d.HaltSpec = d.newConditionalBreakpointLocked(option.Condition)
		case TLCUnsatisfiedBreakpointsFilter:
			d.HaltUnsat = d.newConditionalBreakpointLocked(option.Condition)
		}
	}
	return d
}

func (d *TLCDebugger) Threads() TLCThreadsResponse {
	d.require()
	return TLCThreadsResponse{Threads: []TLCThread{{ID: 0, Name: "worker"}}}
}

func (d *TLCDebugger) SetVariable() TLCSetVariableResponse {
	d.require()
	return TLCSetVariableResponse{}
}

func (d *TLCDebugger) ConfigurationDone() { d.require() }

func (d *TLCDebugger) SetBreakpoints(source string, requested []TLCSourceBreakpointRequest) []TLCBreakpoint {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	breakpointMap := d.ownedBreakpoints()
	if breakpointMap == nil {
		breakpointMap = NewInsMap[string, []*TLCSourceBreakpoint]()
		d.Breakpoints = breakpointMap
	}
	module := debugModuleNameFromSource(source)
	if len(requested) == 0 {
		if _, ok := breakpointMap.Get2(module); ok {
			breakpointMap.Set(module, nil)
		}
		return nil
	}
	breakpoints := make([]*TLCSourceBreakpoint, 0, len(requested))
	results := make([]TLCBreakpoint, 0, len(requested))
	for i, req := range requested {
		line := int(int32(req.Line))
		col := 1
		var column *int
		if req.Column != nil {
			col = int(int32(*req.Column))
			column = &col
		}
		location := NewSourceLocation(module, int(int32(line)+1), col, line, int(int32(col)+1))
		op, conditionErr := d.breakpointConditionOpLocked(req.Condition, location)
		breakpoint := NewTLCSourceBreakpointAt(module, line, column, req.Condition, req.LogMessage, req.HitCondition, op, conditionErr)
		breakpoints = append(breakpoints, breakpoint)
		moduleKnown, verified := d.verifyBreakpointLocationLocked(module, breakpoint)
		result := TLCBreakpoint{
			ID:       i,
			Source:   source,
			Module:   module,
			Line:     breakpoint.Line,
			Column:   breakpoint.Column,
			Verified: verified,
		}
		if d.nextBreakpointRejectsHitConditionLocked(breakpoint) {
			result.Verified = false
			result.Message = "A Next breakpoint does not support a hit condition."
		}
		if moduleKnown && conditionErr != nil {
			result.Verified = false
			result.Message = conditionErr.Error()
		}
		results = append(results, result)
	}
	breakpointMap.Set(module, breakpoints)
	return results
}

func (d *TLCDebugger) verifyBreakpointLocationLocked(module string, breakpoint *TLCSourceBreakpoint) (bool, bool) {
	if breakpoint == nil {
		return false, true
	}
	if d == nil || d.Tool == nil {
		return false, true
	}
	moduleNode := d.Tool.GetModule(module)
	if moduleNode == nil {
		return false, true
	}
	target := breakpoint.GetLocation()
	verified := false
	// SemanticNode.walkChildren visits the root without preemption. Each child
	// is preempted once a match is found or its location excludes the request.
	var visit func(SemanticNode)
	visit = func(node SemanticNode) {
		location, _ := semanticNodeSourceLocation(node)
		if location.BeginLine == breakpoint.Line && location.EndLine == breakpoint.Line {
			verified = true
		}
		for _, child := range SemanticChildren(node) {
			location, ok := semanticNodeSourceLocation(child)
			if !verified && ok && location.Includes(target) {
				visit(child)
			}
		}
	}
	visit(moduleNode)
	return true, verified
}

func (d *TLCDebugger) nextBreakpointRejectsHitConditionLocked(breakpoint *TLCSourceBreakpoint) bool {
	if d == nil || d.Tool == nil || d.Tool.GetSpecProcessor() == nil || breakpoint == nil || breakpoint.GetHits() <= 0 {
		return false
	}
	nextPred := d.Tool.GetSpecProcessor().GetNextPred()
	if nextPred == nil {
		return false
	}
	return nextPred.GetDefinition().Includes(breakpoint.GetLocation())
}

func (d *TLCDebugger) BreakpointsForSource(source string) []*TLCSourceBreakpoint {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	breakpointMap := d.ownedBreakpoints()
	if breakpointMap == nil {
		return nil
	}
	module := debugModuleNameFromSource(source)
	breakpoints, ok := breakpointMap.Get2(module)
	if !ok || len(breakpoints) == 0 {
		return nil
	}
	return append([]*TLCSourceBreakpoint(nil), breakpoints...)
}

func (d *TLCDebugger) MatchesBreakpointFrame(frame *TLCStackFrame) bool {
	d.require()
	if frame == nil {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.matchesBreakpointFrameLocked(frame)
}

func (d *TLCDebugger) DisconnectCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if breakpointMap := d.ownedBreakpoints(); breakpointMap != nil {
		breakpointMap.DeleteAll()
	}
	d.SourceFrame = nil
	d.Step = DebugStepCommandContinue
	d.HaltExp = false
	d.HaltInv = false
	d.HaltSpec = nil
	d.HaltUnsat = nil
	d.resumeExecutionLocked()
	d.Paused = false
	d.clearSyntheticTraceFramesLocked()
	return d
}

func (d *TLCDebugger) TerminateCommand() *TLCDebugger {
	d.require()
	if checker := MainChecker(); checker != nil {
		checker.Stop()
	}
	if simulator := CurrentSimulator(); simulator != nil {
		simulator.Stop()
	}
	return d.DisconnectCommand()
}

func (d *TLCDebugger) TopFrame() *TLCDebuggerFrame {
	d.require()
	stack := d.ownedStack()
	if len(stack) == 0 {
		return nil
	}
	return stack[len(stack)-1]
}

func (d *TLCDebugger) topBaseFrame() *TLCStackFrame {
	if top := d.TopFrame(); top != nil {
		return top.Base
	}
	return nil
}

func (d *TLCDebugger) pushDebuggerFrame(frame *TLCDebuggerFrame) *TLCDebugger {
	if d == nil || frame == nil {
		return d
	}
	d.setOwnedStack(append(d.ownedStack(), frame))
	return d
}

func (d *TLCDebugger) popDebuggerFrame() *TLCDebuggerFrame {
	if d == nil {
		return nil
	}
	stack := d.ownedStack()
	if len(stack) == 0 {
		return nil
	}
	idx := len(stack) - 1
	frame := stack[idx]
	stack[idx] = nil
	d.setOwnedStack(stack[:idx])
	return frame
}

func (d *TLCDebugger) stackFramesLocked() []*TLCStackFrame {
	if d == nil {
		return nil
	}
	stack := d.ownedStack()
	if len(stack) == 0 {
		return nil
	}
	out := make([]*TLCStackFrame, 0, len(stack))
	for i := len(stack) - 1; i >= 0; i-- {
		if frame := stack[i]; frame != nil && frame.Base != nil {
			out = append(out, frame.Base)
		}
	}
	return out
}

func (d *TLCDebugger) ensureSyntheticTraceFramesLocked() {
	if d == nil {
		return
	}
	stack := d.ownedStack()
	if len(stack) == 0 {
		return
	}
	if first := stack[0]; first != nil && first.Synthetic != nil {
		return
	}
	top := d.TopFrame()
	stateFrame := top.stateTraceFrame()
	if stateFrame == nil {
		return
	}
	traceFrames := top.GetTraceAsStackFrames()
	if len(traceFrames) == 0 {
		return
	}
	prefix := make([]*TLCDebuggerFrame, 0, len(traceFrames)+len(stack))
	for i := len(traceFrames) - 1; i >= 0; i-- {
		prefix = append(prefix, NewDebuggerSyntheticFrame(traceFrames[i]))
	}
	d.setOwnedStack(append(prefix, stack...))
}

func (d *TLCDebugger) clearSyntheticTraceFramesLocked() {
	if d == nil {
		return
	}
	stack := d.ownedStack()
	for len(stack) > 0 && stack[0] != nil && stack[0].Synthetic != nil {
		stack[0] = nil
		stack = stack[1:]
	}
	d.setOwnedStack(stack)
}

func (d *TLCDebugger) frameMatchesBreakpointLocked(frame *TLCStackFrame, bp *TLCSourceBreakpoint) bool {
	for _, candidate := range d.ownedStack() {
		if candidate != nil && candidate.Base == frame {
			return candidate.MatchesBreakpoint(bp)
		}
	}
	return frame.MatchesBreakpoint(bp)
}

func (d *TLCDebugger) matchesBreakpointFrameLocked(frame *TLCStackFrame) bool {
	if d == nil || frame == nil {
		return false
	}
	breakpointMap := d.ownedBreakpoints()
	if breakpointMap == nil {
		return false
	}
	loc, ok := semanticNodeSourceLocation(frame.GetNode())
	if !ok {
		return false
	}
	module := debugModuleNameFromSource(loc.Source)
	breakpoints, ok := breakpointMap.Get2(module)
	if !ok || len(breakpoints) == 0 {
		return false
	}
	for _, breakpoint := range breakpoints {
		if breakpoint == nil || !d.frameMatchesBreakpointLocked(frame, breakpoint) {
			continue
		}
		matchedParent := false
		for parent := frame.Parent; parent != nil; parent = parent.Parent {
			if d.frameMatchesBreakpointLocked(parent, breakpoint) {
				matchedParent = true
				break
			}
		}
		if !matchedParent {
			return true
		}
	}
	return false
}

func (d *TLCDebugger) StackFrames() []*TLCStackFrame {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stackFramesLocked()
}

func (d *TLCDebugger) StackTrace(args TLCStackTraceArguments) TLCStackTraceResponse {
	d.require()
	response := TLCStackTraceResponse{StackFrames: []*TLCStackFrame{}}
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.ExecutionIsHalted {
		return response
	}
	d.ensureSyntheticTraceFramesLocked()
	frames := d.stackFramesLocked()
	if len(frames) == 0 {
		return response
	}
	from := 0
	if args.StartFrame != nil {
		requested := int(int32(*args.StartFrame))
		if requested < 0 || requested > len(frames)-1 {
			return response
		}
		from = requested
	}
	to := len(frames)
	if args.Levels != nil {
		requested := int(int32(*args.Levels))
		end := int(int32(from) + int32(requested))
		if requested > 0 && end < to {
			to = end
		}
	}
	if from > to {
		panic(NewIllegalArgumentException(fmt.Sprintf("fromIndex(%d) > toIndex(%d)", from, to)))
	}
	response.StackFrames = append(response.StackFrames, frames[from:to]...)
	response.TotalFrames = len(frames)
	return response
}

func (d *TLCDebugger) Scopes(frameID int) []TLCScope {
	d.require()
	frameID = int(int32(frameID))
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, frame := range d.ownedStack() {
		if frame != nil && frame.Base != nil && frame.Base.ID == frameID {
			return frame.GetScopes()
		}
	}
	return nil
}

func (d *TLCDebugger) Variables(ref int) []*DebugTLCVariable {
	d.require()
	ref = int(int32(ref))
	d.mu.Lock()
	defer d.mu.Unlock()
	rnd := rand.New(rand.NewSource(int64(max(ref, 1))))
	var out []*DebugTLCVariable
	for _, frame := range d.ownedStack() {
		out = append(out, frame.GetVariables(ref, rnd)...)
	}
	return sortedDistinctDebugVariables(out)
}

func (d *TLCDebugger) HaltExecution(frame *TLCStackFrame, level ...int) {
	d.require()
	if len(level) > 1 {
		panic(NewIllegalArgumentException("HaltExecution accepts at most one level"))
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(level) == 1 {
		d.maybeHaltExecutionLocked(frame, level...)
		return
	}
	d.haltExecutionLocked(frame)
}

func (d *TLCDebugger) haltExecutionLocked(frame *TLCStackFrame) {
	top := d.TopFrame()
	stateSelection := top != nil && top.Base == frame && (top.Init != nil || top.Next != nil)
	if stateSelection {
		d.sendCapabilitiesLocked(TLCCapabilitiesNoStepBack)
		d.Granularity = DebugGranularityState
	}
	event := frame.GetStoppedEventArgument()
	d.ExecutionIsHalted = true
	d.stoppedEvents <- event
	for d.ExecutionIsHalted {
		d.resume.Wait()
	}
	d.clearSyntheticTraceFramesLocked()
	if d.Step == DebugStepCommandReset {
		d.Step = DebugStepCommandIn
		if frame != nil && frame.Parent != nil {
			panic(NewResetEvalException(frame.Parent))
		}
	} else if d.Step == DebugStepCommandResetStart {
		d.Step = DebugStepCommandIn
		stack := d.ownedStack()
		if len(stack) > 0 {
			panic(NewResetEvalException(stack[0].Base))
		}
	}
	if stateSelection {
		d.sendCapabilitiesLocked(TLCCapabilitiesStepBack)
		d.Granularity = DebugGranularityFormula
	}
}

func (d *TLCDebugger) MaybeHaltExecution(frame *TLCStackFrame, level ...int) {
	d.require()
	if len(level) > 1 {
		panic(NewIllegalArgumentException("MaybeHaltExecution accepts at most one level"))
	}
	if frame == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maybeHaltExecutionLocked(frame, level...)
}

func (d *TLCDebugger) maybeHaltExecutionLocked(frame *TLCStackFrame, level ...int) {
	if debugStepMatches(d.Step, d.SourceFrame, frame) || d.matchesBreakpointFrameLocked(frame) {
		d.haltExecutionLocked(frame)
	}
}

func (d *TLCDebugger) resumeExecutionLocked() {
	d.ExecutionIsHalted = false
	d.resume.Signal()
}

func debugStepMatches(step DebugStep, sourceFrame *TLCStackFrame, currentFrame *TLCStackFrame) bool {
	if step == DebugStepCommandIn {
		return true
	}
	if sourceFrame != nil && (step == DebugStepCommandOver || step == DebugStepCommandOut) {
		return sourceFrame.MatchesFrame(currentFrame)
	}
	return false
}

func (d *TLCDebugger) ContinueCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.Handle(d) {
		d.Granularity = DebugGranularityFormula
		d.resumeExecutionLocked()
		d.Paused = false
		d.clearSyntheticTraceFramesLocked()
		return d
	}
	d.SourceFrame = nil
	d.Step = DebugStepCommandContinue
	d.resumeExecutionLocked()
	d.Paused = false
	d.clearSyntheticTraceFramesLocked()
	return d
}

func (d *TLCDebugger) StepOverCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.Handle(d) {
		if top.Next != nil {
			if _, err := top.Next.StepOverSelect(); err != nil {
				panic(err)
			}
		}
		d.Granularity = DebugGranularityFormula
		d.resumeExecutionLocked()
		d.Paused = false
		d.clearSyntheticTraceFramesLocked()
		return d
	}
	d.SourceFrame = d.topBaseFrame()
	d.Step = DebugStepCommandOver
	d.resumeExecutionLocked()
	d.Paused = false
	d.clearSyntheticTraceFramesLocked()
	return d
}

func (d *TLCDebugger) StepInCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.Handle(d) {
		if top.Next != nil {
			if _, err := top.Next.StepInSelect(); err != nil {
				panic(err)
			}
		}
		d.Granularity = DebugGranularityFormula
		d.resumeExecutionLocked()
		d.Paused = false
		d.clearSyntheticTraceFramesLocked()
		return d
	}
	d.Step = DebugStepCommandIn
	d.resumeExecutionLocked()
	d.Paused = false
	d.clearSyntheticTraceFramesLocked()
	return d
}

func (d *TLCDebugger) StepOutCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil {
		if top.Handle(d) {
			if top.Next != nil {
				if _, err := top.Next.StepOutSelect(); err != nil {
					panic(err)
				}
			}
			d.Granularity = DebugGranularityFormula
			d.resumeExecutionLocked()
			d.Paused = false
			d.clearSyntheticTraceFramesLocked()
			return d
		}
		if top.Base != nil {
			d.SourceFrame = top.Base.Parent
		}
		d.Step = DebugStepCommandOut
	}
	d.resumeExecutionLocked()
	d.Paused = false
	d.clearSyntheticTraceFramesLocked()
	return d
}

func (d *TLCDebugger) PauseCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Paused = true
	return d
}

func (d *TLCDebugger) StepBackCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.Handle(d) {
		d.Granularity = DebugGranularityFormula
		d.resumeExecutionLocked()
		d.Paused = false
		d.clearSyntheticTraceFramesLocked()
		return d
	}
	d.Step = DebugStepCommandReset
	d.resumeExecutionLocked()
	d.Paused = false
	d.clearSyntheticTraceFramesLocked()
	return d
}

func (d *TLCDebugger) ReverseContinueCommand() *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.Handle(d) {
		d.Granularity = DebugGranularityFormula
		d.resumeExecutionLocked()
		d.Paused = false
		d.clearSyntheticTraceFramesLocked()
		return d
	}
	d.Step = DebugStepCommandResetStart
	d.resumeExecutionLocked()
	d.Paused = false
	d.clearSyntheticTraceFramesLocked()
	return d
}

func (d *TLCDebugger) GotoStateCommand(ref int) *TLCDebugger {
	d.require()
	ref = int(int32(ref))
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.Handle(d) {
		if top.Next != nil {
			if _, err := top.Next.SelectStateByReference(ref); err != nil {
				panic(err)
			}
		} else if top.Init != nil {
			if _, err := top.Init.SelectStateByReference(ref); err != nil {
				panic(err)
			}
		}
		d.Granularity = DebugGranularityFormula
		d.resumeExecutionLocked()
		d.Paused = false
		d.clearSyntheticTraceFramesLocked()
	}
	return d
}

func (d *TLCDebugger) PushFrame(tool *Tool, expr SemanticNode, c *Context) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStackFrameNoException(d.topBaseFrame(), expr, c, tool)
	d.pushDebuggerFrame(NewDebuggerBaseFrame(frame))
	d.maybeHaltExecutionLocked(frame, d.stackLen())
	return d
}

func (d *TLCDebugger) PushGeneratedStateFrame(state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	source := d.topBaseFrame()
	frame := NewTLCStateStackFrameNoException(source, source.GetNode(), source.GetContext(), source.GetTool(), state)
	d.pushDebuggerFrame(NewDebuggerStateFrame(frame))
	d.maybeHaltExecutionLocked(&frame.TLCStackFrame, d.stackLen())
	return d
}

func (d *TLCDebugger) PopGeneratedStateFrame(state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	source := d.topBaseFrame()
	return d.popFrameLocked(source.GetTool(), source.GetNode(), source.GetContext())
}

func (d *TLCDebugger) PushStateFrame(tool *Tool, expr SemanticNode, c *Context, state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStateStackFrameNoException(d.topBaseFrame(), expr, c, tool, state)
	d.pushDebuggerFrame(NewDebuggerStateFrame(frame))
	d.maybeHaltExecutionLocked(&frame.TLCStackFrame, d.stackLen())
	return d
}

func (d *TLCDebugger) PushActionFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCActionStackFrameNoException(d.topBaseFrame(), expr, c, tool, predecessor, action, state)
	d.pushDebuggerFrame(NewDebuggerActionFrame(frame))
	d.maybeHaltExecutionLocked(&frame.TLCStackFrame, d.stackLen())
	return d
}

func (d *TLCDebugger) PopFrame(tool *Tool, expr SemanticNode, c *Context) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.popFrameLocked(tool, expr, c)
}

func (d *TLCDebugger) popFrameLocked(tool *Tool, expr SemanticNode, c *Context) *TLCDebugger {
	top := d.TopFrame()
	if top != nil && top.Base == d.SourceFrame {
		d.SourceFrame = nil
		d.Step = DebugStepCommandIn
		top.Base.Name = "Exit: " + top.Base.Name
		top.Base.PresentationHint = "subtle"
		d.haltExecutionLocked(top.Base)
		top.Base.Name = "Exit: " + strings.ReplaceAll(top.Base.Name, "Exit: ", "")
		top.Base.PresentationHint = "normal"
	}
	popped := d.popDebuggerFrame()
	if popped == nil {
		panic(NewNoSuchElementException())
	}
	if popped.Base == nil || !popped.Base.MatchesNode(expr) {
		panic(NewAssertionError())
	}
	return d
}

func (d *TLCDebugger) PopValueFrame(tool *Tool, expr SemanticNode, c *Context, value Value) *TLCDebugger {
	d.require()
	d = d.PopFrame(tool, expr, c)
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.MatchesNode(expr) {
		top.Base.SetValue(value)
	}
	return d
}

func (d *TLCDebugger) PushNextStatesFrame(tool *Tool, functor *NextStateFunctor, state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	next := UnknownAction
	if tool != nil && tool.GetNextStateSpec() != nil {
		next = tool.GetNextStateSpec()
	}
	frame := NewTLCNextStatesStackFrame(d.topBaseFrame(), next.ownedPred(), next.ownedContext(), tool, state, functor, next)
	d.pushDebuggerFrame(NewDebuggerNextFrame(frame))
	return d
}

func (d *TLCDebugger) PopNextStatesFrame(tool *Tool, functor *NextStateFunctor, state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	top := d.TopFrame()
	if top == nil {
		panic(NewNoSuchElementException())
	}
	if top.Next == nil {
		panic(NewAssertionError())
	}
	if d.HaltSpec != nil && top != nil && top.MatchesBreakpoint(d.HaltSpec) {
		d.haltExecutionLocked(top.Base)
	}
	_ = d.popDebuggerFrame()
	return d
}

func (d *TLCDebugger) PushInitStatesFrame(tool *Tool, functor *StateFunctor) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	initAction := UnknownAction
	if tool != nil {
		inits := tool.GetInitStateSpec()
		if inits.Size() > 0 && inits.ElementAt(0) != nil {
			initAction = inits.ElementAt(0)
		}
	}
	frame := NewTLCInitStatesStackFrame(d.topBaseFrame(), initAction.ownedPred(), initAction.ownedContext(), tool, functor)
	d.pushDebuggerFrame(NewDebuggerInitFrame(frame))
	return d
}

func (d *TLCDebugger) PopInitStatesFrame(tool *Tool, functor *StateFunctor) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	top := d.TopFrame()
	if top == nil {
		panic(NewNoSuchElementException())
	}
	if top.Init == nil {
		panic(NewAssertionError())
	}
	if d.HaltSpec != nil && top != nil && top.Init != nil && top.Init.GetStates().Size() > 0 && top.MatchesBreakpoint(d.HaltSpec) {
		d.haltExecutionLocked(top.Base)
	}
	_ = d.popDebuggerFrame()
	return d
}

func (d *TLCDebugger) pushFrameAndMaybeHalt(halt bool, frame *TLCDebuggerFrame) *TLCDebugger {
	d.require()
	if frame == nil {
		return d
	}
	d.pushDebuggerFrame(frame)
	if halt && frame.Base != nil {
		d.haltExecutionLocked(frame.Base)
	}
	return d
}

func (d *TLCDebugger) PushExceptionFrame(tool *Tool, expr SemanticNode, c *Context, err error) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if !debugExceptionNotYetHandled(err) {
		return d
	}
	frame := NewTLCStackFrame(d.topBaseFrame(), expr, c, tool, err)
	return d.pushFrameAndMaybeHalt(d.HaltExp, NewDebuggerBaseFrame(frame))
}

func (d *TLCDebugger) PushStateExceptionFrame(tool *Tool, expr SemanticNode, c *Context, state *TLCStateMut, err error) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if !debugExceptionNotYetHandled(err) {
		return d
	}
	frame := NewTLCStateStackFrame(d.topBaseFrame(), expr, c, tool, state, err)
	return d.pushFrameAndMaybeHalt(d.HaltExp, NewDebuggerStateFrame(frame))
}

func (d *TLCDebugger) PushActionExceptionFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut, err error) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if !debugExceptionNotYetHandled(err) {
		return d
	}
	frame := NewTLCActionStackFrame(d.topBaseFrame(), expr, c, tool, predecessor, action, state, err)
	return d.pushFrameAndMaybeHalt(d.HaltExp, NewDebuggerActionFrame(frame))
}

func (d *TLCDebugger) PopExceptionFrame(tool *Tool, expr SemanticNode, c *Context, value Value, err error) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	popped := d.popDebuggerFrame()
	if popped == nil {
		panic(NewNoSuchElementException())
	}
	if popped.Base == nil || !popped.Base.MatchesException(expr, err) {
		panic(NewAssertionError())
	}
	return d
}

func (d *TLCDebugger) PushUnsatisfiedFrame(tool *Tool, expr SemanticNode, c *Context, state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStateStackFrameNoException(d.topBaseFrame(), expr, c, tool, state)
	debuggerFrame := NewDebuggerStateFrame(frame)
	d.pushDebuggerFrame(debuggerFrame)
	if d.HaltUnsat != nil && debuggerFrame.MatchesBreakpoint(d.HaltUnsat) {
		d.haltExecutionLocked(debuggerFrame.Base)
	}
	return d
}

func (d *TLCDebugger) PushUnsatisfiedActionFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCActionStackFrameNoException(d.topBaseFrame(), expr, c, tool, predecessor, action, state)
	debuggerFrame := NewDebuggerActionFrame(frame)
	d.pushDebuggerFrame(debuggerFrame)
	if d.HaltUnsat != nil && debuggerFrame.MatchesBreakpoint(d.HaltUnsat) {
		d.haltExecutionLocked(debuggerFrame.Base)
	}
	return d
}

func (d *TLCDebugger) MarkInvariantViolatedFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut, err error) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	if !debugExceptionNotYetHandled(err) {
		return d
	}
	frame := NewTLCActionStackFrame(d.topBaseFrame(), expr, c, tool, predecessor, action, state, err)
	return d.pushFrameAndMaybeHalt(d.HaltInv, NewDebuggerActionFrame(frame))
}

func (d *TLCDebugger) MarkAssumptionViolatedFrame(tool *Tool, expr SemanticNode, c *Context) *TLCDebugger {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStackFrameNoException(nil, expr, c, tool)
	return d.pushFrameAndMaybeHalt(d.HaltInv, NewDebuggerBaseFrame(frame))
}
