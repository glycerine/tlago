package tlc

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const maxJavaInt = 1<<31 - 1

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
	ID        int
	Name      string
	Node      SemanticNode
	Context   *Context
	Tool      *Tool
	Exception error
	Value     Value
	Parent    *TLCStackFrame
	ContextID int
}

func NewTLCStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, exception error) *TLCStackFrame {
	frame := &TLCStackFrame{
		Parent:    parent,
		Node:      node,
		Context:   ctxt,
		Tool:      tool,
		Exception: exception,
		ContextID: debugVariableReference(nil),
	}
	frame.ID = semanticNodeDebugID(node)
	frame.Name = semanticNodeDebugName(node, exception)
	return frame
}

func NewTLCStackFrameNoException(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool) *TLCStackFrame {
	return NewTLCStackFrame(parent, node, ctxt, tool, nil)
}

func (f *TLCStackFrame) GetNode() SemanticNode {
	if f == nil {
		return nil
	}
	return f.Node
}

func (f *TLCStackFrame) GetContext() *Context {
	if f == nil {
		return nil
	}
	return f.Context
}

func (f *TLCStackFrame) GetTool() *Tool {
	if f == nil {
		return nil
	}
	return f.Tool
}

func (f *TLCStackFrame) HasException() bool {
	return f != nil && f.Exception != nil
}

func (f *TLCStackFrame) GetException() error {
	if f == nil {
		return nil
	}
	return f.Exception
}

func (f *TLCStackFrame) String() string {
	if f == nil {
		return "TLCStackFrame [node=<nil>]"
	}
	return fmt.Sprintf("TLCStackFrame [node=%v]", f.Node)
}

func (f *TLCStackFrame) SetValue(value Value) Value {
	if f != nil {
		f.Value = value
	}
	return value
}

func (f *TLCStackFrame) GetConstantsID() int {
	if f == nil {
		return 1
	}
	return f.ContextID + 1
}

func (f *TLCStackFrame) GetStackID() int {
	if f == nil {
		return 3
	}
	return f.ContextID + 3
}

func (f *TLCStackFrame) MatchesFrame(other *TLCStackFrame) bool {
	if f == nil || other == nil {
		return f == other
	}
	return semanticNodeLevel(f.Node) == semanticNodeLevel(other.Node) && semanticNodeSame(f.Node, other.Node)
}

func (f *TLCStackFrame) MatchesBreakpoint(bp *TLCSourceBreakpoint) bool {
	if f == nil || bp == nil {
		return false
	}
	loc, ok := semanticNodeSourceLocation(f.Node)
	if !ok {
		return false
	}
	return bp.MatchesLocation(loc)
}

func (f *TLCStackFrame) MatchesNode(expr SemanticNode) bool {
	return f != nil && semanticNodeSame(f.Node, expr)
}

func (f *TLCStackFrame) IsTarget(expr SemanticNode) bool {
	return f.MatchesNode(expr)
}

type ResetEvalException struct {
	Frame *TLCStackFrame
}

func NewResetEvalException(frame *TLCStackFrame) *ResetEvalException {
	return &ResetEvalException{Frame: frame}
}

func (e *ResetEvalException) Error() string {
	return "debug reset evaluation"
}

func (e *ResetEvalException) IsTarget(expr SemanticNode) bool {
	return e != nil && e.Frame != nil && e.Frame.IsTarget(expr)
}

type AbortEvalException struct{}

func (e *AbortEvalException) Error() string {
	return "debug abort evaluation"
}

var DebuggerNotEvaluatedValue Value = NewStringValue("?")

const (
	debugScopeState      = "State"
	debugScopeAction     = "Action"
	debugScopeInitials   = "Initials"
	debugScopeSuccessors = "Successors"
	debugScopeTrace      = "Trace"
)

type TLCStateStackFrame struct {
	TLCStackFrame
	State   *TLCStateMut
	StateID int
}

func NewTLCStateStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, state *TLCStateMut, exception error) *TLCStateStackFrame {
	frame := &TLCStateStackFrame{
		TLCStackFrame: *NewTLCStackFrame(parent, node, ctxt, tool, exception),
		State:         debugStateCopy(state),
		StateID:       debugVariableReference(nil),
	}
	return frame
}

func NewTLCStateStackFrameNoException(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, state *TLCStateMut) *TLCStateStackFrame {
	return NewTLCStateStackFrame(parent, node, ctxt, tool, state, nil)
}

func (f *TLCStateStackFrame) GetS() *TLCStateMut {
	if f == nil {
		return nil
	}
	return f.GetT()
}

func (f *TLCStateStackFrame) GetT() *TLCStateMut {
	if f == nil {
		return nil
	}
	return f.State
}

func (f *TLCStateStackFrame) AddT() bool {
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
	return debugStateAsVariable(state, f.ToRecordValue(), name, rnd)
}

type TLCActionStackFrame struct {
	TLCStateStackFrame
	Action *Action
}

func NewTLCActionStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, predecessor *TLCStateMut, action *Action, state *TLCStateMut, exception error) *TLCActionStackFrame {
	copied := debugStateCopy(state)
	if copied != nil && copied.Predecessor() == nil && predecessor != nil {
		copied.SetPredecessor(debugStateCopy(predecessor))
	}
	if copied != nil && action != nil {
		copied.SetAction(action)
	}
	return &TLCActionStackFrame{
		TLCStateStackFrame: *NewTLCStateStackFrame(parent, node, ctxt, tool, copied, exception),
		Action:             action,
	}
}

func NewTLCActionStackFrameNoException(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, predecessor *TLCStateMut, action *Action, state *TLCStateMut) *TLCActionStackFrame {
	return NewTLCActionStackFrame(parent, node, ctxt, tool, predecessor, action, state, nil)
}

func (f *TLCActionStackFrame) GetS() *TLCStateMut {
	if f == nil || f.State == nil {
		return nil
	}
	return f.State.Predecessor()
}

func (f *TLCActionStackFrame) GetT() *TLCStateMut {
	if f == nil {
		return nil
	}
	return f.State
}

func (f *TLCActionStackFrame) AddT() bool {
	return true
}

func (f *TLCActionStackFrame) ToRecordValue() *RecordValue {
	state := f.GetT()
	if state == nil {
		return EmptyRecord
	}
	return debugStateRecordValue(state, f.GetS())
}

type TLCInitStatesStackFrame struct {
	TLCStackFrame
	Functor      *StateFunctor
	IDToStateMap map[int]*TLCStateMut
	StateID      int
}

func NewTLCInitStatesStackFrame(parent *TLCStackFrame, pred SemanticNode, con *Context, tool *Tool, functor *StateFunctor) *TLCInitStatesStackFrame {
	return &TLCInitStatesStackFrame{
		TLCStackFrame: *NewTLCStackFrameNoException(parent, pred, con, tool),
		Functor:       functor,
		IDToStateMap:  make(map[int]*TLCStateMut),
		StateID:       debugVariableReference(nil),
	}
}

func (f *TLCInitStatesStackFrame) GetStates() *SetOfStates {
	if f == nil || f.Functor == nil {
		return NewSetOfStates(0)
	}
	return f.Functor.GetStates()
}

func (f *TLCInitStatesStackFrame) GetStateVariables(rnd *rand.Rand) []*DebugTLCVariable {
	if f == nil {
		return nil
	}
	states := sortedStatesByString(f.GetStates())
	width := len(strconv.Itoa(len(states)))
	out := make([]*DebugTLCVariable, 0, len(states))
	for i, state := range states {
		name := fmt.Sprintf("%d.%0*d: %s", state.Level(), width, i+1, debugActionLocation(state.GetAction()))
		variable := debugStateAsVariable(state, NewRecordValueFromInsMap(state.Values()), name, rnd)
		f.IDToStateMap[variable.VariablesReference] = state
		out = append(out, variable)
	}
	return out
}

func (f *TLCInitStatesStackFrame) SelectStateByReference(ref int) (bool, error) {
	if f == nil || f.Functor == nil {
		return false, nil
	}
	state := f.IDToStateMap[ref]
	if state == nil {
		return false, nil
	}
	_, err := f.Functor.SetElement(state)
	return err == nil, err
}

type TLCSyntheticStateStackFrame struct {
	TLCStateStackFrame
	Successor *TLCStateMut
}

func NewTLCSyntheticStateStackFrame(tool *Tool, info *TLCStateInfo, successor *TLCStateMut, width int) *TLCSyntheticStateStackFrame {
	action := UnknownAction
	if info != nil {
		action = info.Action()
	}
	frame := NewTLCStateStackFrame(nil, action.Pred, action.Con, tool, stateInfoState(info), nil)
	level := 0
	label := ""
	if info != nil {
		level = info.GetStateNumber()
		label = fmt.Sprint(info.Info)
	}
	frame.Name = fmt.Sprintf("%0*d: %s", width, level, label)
	return &TLCSyntheticStateStackFrame{
		TLCStateStackFrame: *frame,
		Successor:          debugStateCopy(successor),
	}
}

func (f *TLCSyntheticStateStackFrame) GetT() *TLCStateMut {
	if f == nil {
		return nil
	}
	return f.State
}

func (f *TLCSyntheticStateStackFrame) GetSuccessor() *TLCStateMut {
	if f == nil {
		return nil
	}
	return f.Successor
}

type TLCNextStatesStackFrame struct {
	TLCStateStackFrame
	Action       *Action
	Functor      *NextStateFunctor
	IDToStateMap map[int]*TLCStateMut
}

func NewTLCNextStatesStackFrame(parent *TLCStackFrame, node SemanticNode, ctxt *Context, tool *Tool, state *TLCStateMut, functor *NextStateFunctor, action *Action) *TLCNextStatesStackFrame {
	frame := NewTLCStateStackFrameNoException(parent, node, ctxt, tool, state)
	frame.Name = fmt.Sprint(node)
	return &TLCNextStatesStackFrame{
		TLCStateStackFrame: *frame,
		Action:             action,
		Functor:            functor,
		IDToStateMap:       make(map[int]*TLCStateMut),
	}
}

func (f *TLCNextStatesStackFrame) AddT() bool {
	return true
}

func (f *TLCNextStatesStackFrame) GetSuccessors() *SetOfStates {
	if f == nil || f.Functor == nil {
		return NewSetOfStates(0)
	}
	return f.Functor.GetStates()
}

func (f *TLCNextStatesStackFrame) HasScope() bool {
	return f != nil && f.GetSuccessors().Size() > 0
}

func (f *TLCNextStatesStackFrame) GetStateVariables(rnd *rand.Rand) []*DebugTLCVariable {
	if f == nil {
		return nil
	}
	return []*DebugTLCVariable{f.ToVariable(rnd)}
}

func (f *TLCNextStatesStackFrame) GetSuccessorVariables(rnd *rand.Rand) []*DebugTLCVariable {
	if f == nil {
		return nil
	}
	successors := sortedSuccessors(f.GetSuccessors())
	width := len(strconv.Itoa(len(successors)))
	out := make([]*DebugTLCVariable, 0, len(successors))
	for i, state := range successors {
		name := fmt.Sprintf("%d.%0*d: %s", state.Level(), width, i+1, debugActionLocation(state.GetAction()))
		variable := debugStateAsVariable(state, NewRecordValueFromInsMap(state.Values()), name, rnd)
		f.IDToStateMap[variable.VariablesReference] = state
		out = append(out, variable)
	}
	return out
}

func (f *TLCNextStatesStackFrame) SelectStateByReference(ref int) (bool, error) {
	if f == nil || f.Functor == nil {
		return false, nil
	}
	state := f.IDToStateMap[ref]
	if state == nil {
		return false, nil
	}
	_, err := f.Functor.SetElement(state)
	return err == nil, err
}

func (f *TLCNextStatesStackFrame) StepInSelect() (bool, error) {
	return f.selectSuccessorByDistance(true)
}

func (f *TLCNextStatesStackFrame) StepOverSelect() (bool, error) {
	return f.selectSuccessorByDistance(false)
}

func (f *TLCNextStatesStackFrame) StepOutSelect() (bool, error) {
	if f == nil || f.Functor == nil {
		return false, nil
	}
	predecessor := (*TLCStateMut)(nil)
	if state := f.GetS(); state != nil {
		predecessor = state.Predecessor()
	}
	if predecessor == nil {
		return f.Functor.Halt(), nil
	}
	_, err := f.Functor.SetElement(predecessor)
	return err == nil, err
}

func (f *TLCNextStatesStackFrame) selectSuccessorByDistance(minimum bool) (bool, error) {
	if f == nil || f.Functor == nil {
		return false, nil
	}
	states := f.GetSuccessors().ToSlice()
	if len(states) == 0 {
		return false, nil
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
	_, err := f.Functor.SetElement(selected)
	return err == nil, err
}

type TLCCapabilities struct {
	SupportsStepBack  bool
	SupportsGotoState bool
}

var (
	TLCCapabilitiesStepBack   = TLCCapabilities{SupportsStepBack: true}
	TLCCapabilitiesNoStepBack = TLCCapabilities{SupportsStepBack: false}
)

type GotoStateArgument struct {
	VariablesReference int
}

func (a *GotoStateArgument) SetVariablesReference(ref int) *GotoStateArgument {
	if a == nil {
		a = &GotoStateArgument{}
	}
	a.VariablesReference = ref
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
		BeginLine:   beginLine,
		BeginColumn: beginColumn,
		EndLine:     endLine,
		EndColumn:   endColumn,
	}
}

func (l SourceLocation) IsNull() bool {
	return l.Source == "" && l.BeginLine == 0 && l.BeginColumn == 0 && l.EndLine == 0 && l.EndColumn == 0
}

type DebugTLCVariable struct {
	Name                      string
	Type                      string
	Value                     string
	VariablesReference        int
	VSCodeVariableMenuContext string
	TLCValue                  Value
}

func NewDebugTLCVariable(name *UniqueString) *DebugTLCVariable {
	if name == nil {
		return &DebugTLCVariable{}
	}
	return NewDebugTLCVariableName(name.String())
}

func NewDebugTLCVariableName(name string) *DebugTLCVariable {
	return &DebugTLCVariable{Name: name}
}

func NewDebugTLCVariableValue(value Value, rnd *rand.Rand) *DebugTLCVariable {
	var name string
	if value != nil {
		name = value.String()
	}
	return debugValueToVariable(NewDebugTLCVariableName(name), value, rnd)
}

func (v *DebugTLCVariable) SetVscodeVariableMenuContext(context string) {
	if v != nil {
		v.VSCodeVariableMenuContext = context
	}
}

func (v *DebugTLCVariable) GetVscodeVariableMenuContext() string {
	if v == nil {
		return ""
	}
	return v.VSCodeVariableMenuContext
}

func (v *DebugTLCVariable) SetInstance(value Value) *DebugTLCVariable {
	if v == nil {
		v = &DebugTLCVariable{}
	}
	v.TLCValue = value
	return v
}

func (v *DebugTLCVariable) GetTLCValue() Value {
	if v == nil {
		return nil
	}
	return v.TLCValue
}

func (v *DebugTLCVariable) NewInstance(name string, value Value, rnd *rand.Rand) *DebugTLCVariable {
	variable := NewDebugTLCVariableName(name).SetInstance(value)
	if debugValueMayHaveNested(value) {
		variable.VariablesReference = debugVariableReference(rnd)
	}
	return debugValueToVariable(variable, value, rnd)
}

func (v *DebugTLCVariable) NewInstanceValue(value Value, rnd *rand.Rand) *DebugTLCVariable {
	name := ""
	if value != nil {
		name = value.String()
	}
	return v.NewInstance(name, value, rnd)
}

func (v *DebugTLCVariable) Nested(rnd *rand.Rand) []*DebugTLCVariable {
	if v == nil || v.TLCValue == nil {
		return nil
	}
	return debugValueNested(v.TLCValue, v, rnd)
}

func (v *DebugTLCVariable) Compare(other *DebugTLCVariable) int {
	if v == nil || other == nil {
		switch {
		case v == other:
			return 0
		case v == nil:
			return -1
		default:
			return 1
		}
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

func debugValueMayHaveNested(value Value) bool {
	switch value.(type) {
	case *FcnRcdValue, *RecordValue, *TupleValue:
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
	if s, ok := value.(*StringValue); ok {
		return s.UnquotedString()
	}
	return value.String()
}

func debugValueTypeString(value Value) string {
	if value == nil {
		return ""
	}
	typ := reflect.TypeOf(value)
	name := "Value"
	if typ != nil {
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		name = typ.Name()
	}
	return name + ": " + value.KindString()
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

func debugStateCopy(state *TLCStateMut) *TLCStateMut {
	if state == nil {
		return nil
	}
	return state.DeepCopy()
}

func debugStateAsVariable(state *TLCStateMut, record Value, name string, rnd *rand.Rand) *DebugTLCVariable {
	variable := NewDebugTLCVariableName(name).SetInstance(record)
	variable.SetVscodeVariableMenuContext("state")
	variable.Type = "TLCState"
	if state != nil && state.AllAssigned() {
		variable.Type = fmt.Sprintf("FP64: %d", int64(state.FingerPrint()))
	}
	if debugValueMayHaveNested(variable.TLCValue) {
		variable.VariablesReference = debugVariableReference(rnd)
	}
	variable.Value = debugValueString(variable.TLCValue)
	return variable
}

func stateInfoState(info *TLCStateInfo) *TLCStateMut {
	if info == nil {
		return nil
	}
	return info.State
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
	names := []*UniqueString{}
	values := []Value{}
	if predecessor != nil {
		for name, value := range predecessor.Values().All() {
			names = append(names, UniqueStringOf(name.String()))
			values = append(values, value)
		}
	}
	for name, value := range state.Values().All() {
		field := name
		if predecessor != nil {
			field = UniqueStringOf(name.String() + "'")
		}
		if value == nil {
			value = DebuggerNotEvaluatedValue
		}
		names = append(names, field)
		values = append(values, value)
	}
	return NewRecordValue(names, values, false)
}

func sortedStatesByString(set *SetOfStates) []*TLCStateMut {
	if set == nil {
		return nil
	}
	states := set.ToSlice()
	sort.Slice(states, func(i, j int) bool {
		return states[i].String() < states[j].String()
	})
	return states
}

func sortedSuccessors(set *SetOfStates) []*TLCStateMut {
	states := sortedStatesByString(set)
	sort.SliceStable(states, func(i, j int) bool {
		left := debugActionLocation(states[i].GetAction())
		right := debugActionLocation(states[j].GetAction())
		if left != right {
			return left < right
		}
		return states[i].String() < states[j].String()
	})
	return states
}

func debugHammingDistance(left *TLCStateMut, right *TLCStateMut) int {
	leftText := ""
	if left != nil {
		leftText = left.String()
	}
	rightText := ""
	if right != nil {
		rightText = right.String()
	}
	minLen := len(leftText)
	if len(rightText) < minLen {
		minLen = len(rightText)
	}
	distance := absInt(len(leftText) - len(rightText))
	for i := 0; i < minLen; i++ {
		if leftText[i] != rightText[i] {
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

func debugValueNested(value Value, prototype *DebugTLCVariable, rnd *rand.Rand) []*DebugTLCVariable {
	switch v := value.(type) {
	case *TupleValue:
		return debugTupleVariables(v, prototype, rnd)
	case *RecordValue:
		return debugRecordVariables(v, prototype, rnd)
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
	out := make([]*DebugTLCVariable, 0, len(value.Elems))
	width := len(strconv.Itoa(len(value.Elems)))
	format := "%0" + strconv.Itoa(width) + "d"
	for i, elem := range value.Elems {
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
	out := make([]*DebugTLCVariable, 0, len(value.Values))
	for i, name := range value.Names {
		field := ""
		if name != nil {
			field = name.String()
		}
		var elem Value
		if i < len(value.Values) {
			elem = value.Values[i]
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
	out := make([]*DebugTLCVariable, 0, len(value.Values))
	for i, elem := range value.Values {
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
	if value.Intv == nil {
		return append([]Value(nil), value.Domain...)
	}
	if value.Intv.High < value.Intv.Low {
		return nil
	}
	out := make([]Value, 0, int(value.Intv.High-value.Intv.Low+1))
	for i := value.Intv.Low; i <= value.Intv.High; i++ {
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
	for {
		elem := enum.NextElement()
		if elem == nil {
			break
		}
		out = append(out, debugValueToVariable(prototype.NewInstanceValue(elem, rnd), elem, rnd))
	}
	return out
}

func valueStringOrEmpty(value Value) string {
	if value == nil {
		return ""
	}
	return value.String()
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
}

func NewTLCSourceBreakpoint(condition string, op *OpDefNode, conditionErr error) *TLCSourceBreakpoint {
	return &TLCSourceBreakpoint{
		Line:               NullSourceLocation.BeginLine,
		Condition:          condition,
		Location:           NullSourceLocation,
		ConditionOp:        op,
		ConditionException: conditionErr,
	}
}

func NewTLCSourceBreakpointAt(module string, line int, column *int, condition string, logMessage string, hitCondition string, op *OpDefNode, conditionErr error) *TLCSourceBreakpoint {
	col := 1
	if column != nil {
		col = *column
	}
	return &TLCSourceBreakpoint{
		Line:               line,
		Column:             column,
		Condition:          condition,
		LogMessage:         logMessage,
		HitCondition:       hitCondition,
		Hits:               parseBreakpointHits(hitCondition),
		Location:           NewSourceLocation(module, line+1, col, line, col+1),
		ConditionOp:        op,
		ConditionException: conditionErr,
	}
}

func (b *TLCSourceBreakpoint) GetHits() int {
	if b == nil {
		return 0
	}
	return b.Hits
}

func (b *TLCSourceBreakpoint) IsInline() bool {
	return b != nil && b.GetColumnAsInt() == -1
}

func (b *TLCSourceBreakpoint) GetColumnAsInt() int {
	if b == nil || b.Column == nil {
		return -1
	}
	return *b.Column
}

func (b *TLCSourceBreakpoint) GetLocation() SourceLocation {
	if b == nil {
		return NullSourceLocation
	}
	return b.Location
}

func (b *TLCSourceBreakpoint) GetConditionException() error {
	if b == nil {
		return nil
	}
	return b.ConditionException
}

func (b *TLCSourceBreakpoint) MatchesExpression(tool *Tool, s *TLCStateMut, t *TLCStateMut, c *Context, fire bool) bool {
	if b == nil || b.ConditionOp == nil || tool == nil {
		return fire
	}
	ctxt := EmptyContext
	if c == nil {
		c = EmptyContext
	}
	for _, param := range b.ConditionOp.Params {
		value := c.LookupFunc(func(sym *SymbolNode) bool {
			return param != nil && sym != nil && param.Name == sym.Name
		})
		ctxt = ctxt.Cons(param, value)
	}
	eval, err := tool.NoDebug().Eval(b.ConditionOp.Body, ctxt, s, t, EvalClear)
	if err != nil {
		return fire
	}
	if boolValue, ok := eval.(*BoolValue); ok {
		fire = fire && boolValue.Val
	}
	return fire
}

func (b *TLCSourceBreakpoint) MatchesLocation(loc SourceLocation) bool {
	if b == nil {
		return false
	}
	if b.Location.IsNull() {
		return true
	}
	return b.Line == loc.BeginLine && b.GetColumnAsInt() <= loc.BeginColumn
}

func parseBreakpointHits(hitCondition string) int {
	if hitCondition == "" {
		return 0
	}
	hits, err := strconv.Atoi(hitCondition)
	if err != nil {
		return 0
	}
	return hits
}

type TLCDebugger struct {
	mu                sync.Mutex
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
	return f != nil && f.Base != nil && f.Base.MatchesBreakpoint(bp)
}

func NewTLCDebugger(tool *Tool) *TLCDebugger {
	return &TLCDebugger{
		Tool:        tool,
		Granularity: DebugGranularityFormula,
		Direction:   DebugStepContinue,
		Step:        DebugStepCommandContinue,
		Breakpoints: NewInsMap[string, []*TLCSourceBreakpoint](),
	}
}

func (d *TLCDebugger) SetTool(tool *Tool) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	} else {
		d.Tool = tool
	}
	return d
}

func (d *TLCDebugger) SetGranularity(granularity DebugGranularity) {
	if d != nil {
		d.Granularity = granularity
	}
}

func (d *TLCDebugger) GetGranularity() DebugGranularity {
	if d == nil {
		return DebugGranularityFormula
	}
	return d.Granularity
}

func (d *TLCDebugger) TopFrame() *TLCDebuggerFrame {
	if d == nil || len(d.Stack) == 0 {
		return nil
	}
	return d.Stack[len(d.Stack)-1]
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
	d.Stack = append(d.Stack, frame)
	return d
}

func (d *TLCDebugger) popDebuggerFrame() *TLCDebuggerFrame {
	if d == nil || len(d.Stack) == 0 {
		return nil
	}
	idx := len(d.Stack) - 1
	frame := d.Stack[idx]
	d.Stack[idx] = nil
	d.Stack = d.Stack[:idx]
	return frame
}

func (d *TLCDebugger) StackFrames() []*TLCStackFrame {
	if d == nil || len(d.Stack) == 0 {
		return nil
	}
	out := make([]*TLCStackFrame, 0, len(d.Stack))
	for i := len(d.Stack) - 1; i >= 0; i-- {
		if frame := d.Stack[i]; frame != nil && frame.Base != nil {
			out = append(out, frame.Base)
		}
	}
	return out
}

func (d *TLCDebugger) HaltExecution(frame *TLCStackFrame, level ...int) {
	if d == nil {
		return
	}
	d.ExecutionIsHalted = true
	if frame != nil {
		d.SourceFrame = frame
	}
}

func (d *TLCDebugger) PushFrame(tool *Tool, expr SemanticNode, c *Context) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStackFrameNoException(d.topBaseFrame(), expr, c, tool)
	d.pushDebuggerFrame(NewDebuggerBaseFrame(frame))
	d.HaltExecution(frame, len(d.Stack))
	return d
}

func (d *TLCDebugger) PushStateFrame(tool *Tool, expr SemanticNode, c *Context, state *TLCStateMut) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStateStackFrameNoException(d.topBaseFrame(), expr, c, tool, state)
	d.pushDebuggerFrame(NewDebuggerStateFrame(frame))
	d.HaltExecution(&frame.TLCStackFrame, len(d.Stack))
	return d
}

func (d *TLCDebugger) PushActionFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCActionStackFrameNoException(d.topBaseFrame(), expr, c, tool, predecessor, action, state)
	d.pushDebuggerFrame(NewDebuggerActionFrame(frame))
	d.HaltExecution(&frame.TLCStackFrame, len(d.Stack))
	return d
}

func (d *TLCDebugger) PopFrame(tool *Tool, expr SemanticNode, c *Context) *TLCDebugger {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	top := d.TopFrame()
	if top != nil && top.Base == d.SourceFrame {
		d.SourceFrame = nil
		d.Step = DebugStepCommandIn
	}
	_ = d.popDebuggerFrame()
	return d
}

func (d *TLCDebugger) PopValueFrame(tool *Tool, expr SemanticNode, c *Context, value Value) *TLCDebugger {
	d = d.PopFrame(tool, expr, c)
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if top := d.TopFrame(); top != nil && top.MatchesNode(expr) {
		top.Base.SetValue(value)
	}
	return d
}

func (d *TLCDebugger) PushNextStatesFrame(tool *Tool, functor *NextStateFunctor, state *TLCStateMut) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	next := UnknownAction
	if tool != nil && tool.GetNextStateSpec() != nil {
		next = tool.GetNextStateSpec()
	}
	frame := NewTLCNextStatesStackFrame(d.topBaseFrame(), next.Pred, next.Con, tool, state, functor, next)
	d.pushDebuggerFrame(NewDebuggerNextFrame(frame))
	return d
}

func (d *TLCDebugger) PopNextStatesFrame(tool *Tool, functor *NextStateFunctor, state *TLCStateMut) *TLCDebugger {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	top := d.TopFrame()
	if d.HaltSpec != nil && top != nil && top.MatchesBreakpoint(d.HaltSpec) {
		d.HaltExecution(top.Base)
	}
	_ = d.popDebuggerFrame()
	return d
}

func (d *TLCDebugger) PushInitStatesFrame(tool *Tool, functor *StateFunctor) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	initAction := UnknownAction
	if tool != nil {
		inits := tool.GetInitStateSpec()
		if len(inits) > 0 && inits[0] != nil {
			initAction = inits[0]
		}
	}
	frame := NewTLCInitStatesStackFrame(d.topBaseFrame(), initAction.Pred, initAction.Con, tool, functor)
	d.pushDebuggerFrame(NewDebuggerInitFrame(frame))
	return d
}

func (d *TLCDebugger) PopInitStatesFrame(tool *Tool, functor *StateFunctor) *TLCDebugger {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	top := d.TopFrame()
	if d.HaltSpec != nil && top != nil && top.Init != nil && top.Init.GetStates().Size() > 0 && top.MatchesBreakpoint(d.HaltSpec) {
		d.HaltExecution(top.Base)
	}
	_ = d.popDebuggerFrame()
	return d
}

func (d *TLCDebugger) pushFrameAndMaybeHalt(halt bool, frame *TLCDebuggerFrame) *TLCDebugger {
	if d == nil || frame == nil {
		return d
	}
	d.pushDebuggerFrame(frame)
	if halt && frame.Base != nil {
		d.HaltExecution(frame.Base)
	}
	return d
}

func (d *TLCDebugger) PushExceptionFrame(tool *Tool, expr SemanticNode, c *Context, err error) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStackFrame(d.topBaseFrame(), expr, c, tool, err)
	return d.pushFrameAndMaybeHalt(d.HaltExp, NewDebuggerBaseFrame(frame))
}

func (d *TLCDebugger) PushStateExceptionFrame(tool *Tool, expr SemanticNode, c *Context, state *TLCStateMut, err error) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStateStackFrame(d.topBaseFrame(), expr, c, tool, state, err)
	return d.pushFrameAndMaybeHalt(d.HaltExp, NewDebuggerStateFrame(frame))
}

func (d *TLCDebugger) PushActionExceptionFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut, err error) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCActionStackFrame(d.topBaseFrame(), expr, c, tool, predecessor, action, state, err)
	return d.pushFrameAndMaybeHalt(d.HaltExp, NewDebuggerActionFrame(frame))
}

func (d *TLCDebugger) PopExceptionFrame(tool *Tool, expr SemanticNode, c *Context, value Value, err error) *TLCDebugger {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	_ = d.popDebuggerFrame()
	return d
}

func (d *TLCDebugger) PushUnsatisfiedFrame(tool *Tool, expr SemanticNode, c *Context, state *TLCStateMut) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStateStackFrameNoException(d.topBaseFrame(), expr, c, tool, state)
	debuggerFrame := NewDebuggerStateFrame(frame)
	d.pushDebuggerFrame(debuggerFrame)
	if d.HaltUnsat != nil && debuggerFrame.MatchesBreakpoint(d.HaltUnsat) {
		d.HaltExecution(debuggerFrame.Base)
	}
	return d
}

func (d *TLCDebugger) PushUnsatisfiedActionFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCActionStackFrameNoException(d.topBaseFrame(), expr, c, tool, predecessor, action, state)
	debuggerFrame := NewDebuggerActionFrame(frame)
	d.pushDebuggerFrame(debuggerFrame)
	if d.HaltUnsat != nil && debuggerFrame.MatchesBreakpoint(d.HaltUnsat) {
		d.HaltExecution(debuggerFrame.Base)
	}
	return d
}

func (d *TLCDebugger) MarkInvariantViolatedFrame(tool *Tool, expr SemanticNode, c *Context, predecessor *TLCStateMut, action *Action, state *TLCStateMut, err error) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCActionStackFrame(d.topBaseFrame(), expr, c, tool, predecessor, action, state, err)
	return d.pushFrameAndMaybeHalt(d.HaltInv, NewDebuggerActionFrame(frame))
}

func (d *TLCDebugger) MarkAssumptionViolatedFrame(tool *Tool, expr SemanticNode, c *Context) *TLCDebugger {
	if d == nil {
		d = NewTLCDebugger(tool)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	frame := NewTLCStackFrameNoException(nil, expr, c, tool)
	return d.pushFrameAndMaybeHalt(d.HaltInv, NewDebuggerBaseFrame(frame))
}
