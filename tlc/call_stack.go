package tlc

import (
	"strconv"
	"strings"
)

type CallStack struct {
	stack  []SemanticNode
	index  int
	frozen bool
}

func NewCallStack() *CallStack {
	return &CallStack{stack: make([]SemanticNode, 64)}
}

func NewCallStackTool(other *Tool) *Tool {
	if other == nil {
		panic(NewNullPointerException())
	}
	// Java constructs a Tool from the shared Spec fields, not a DebugTool.
	// Start from its ordinary evaluator so replay cannot bypass this call stack.
	tool := *other.NoDebug()
	tool.CallStack = NewCallStack()
	return &tool
}

func (t *Tool) HasCallStack() bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.CallStack != nil && t.CallStack.Size() > 0
}

func (t *Tool) CallStackString() string {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.CallStack == nil {
		return NewCallStack().String()
	}
	return t.CallStack.String()
}

func (t *Tool) callStackEnter(expr SemanticNode) func(error) {
	if t == nil || t.CallStack == nil {
		return func(error) {}
	}
	t.CallStack.Push(expr)
	return func(err error) {
		if err != nil {
			switch failure := err.(type) {
			case *FingerprintException:
				t.CallStack.Freeze(failure)
			case *EvalException:
				t.CallStack.Freeze()
			case *TLCError:
				if javaSystemFailureCode(failure) == NoError {
					t.CallStack.Freeze()
				}
			}
		}
		t.CallStack.Pop()
	}
}

// SetSource mirrors Tool.setSource: FastTool and DebugTool do nothing, while
// CallStackTool replaces the source even if a value already has one.
func (t *Tool) SetSource(expr SemanticNode, value Value) Value {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.CallStack != nil {
		ValueSetSource(value, expr)
	}
	return value
}

func (t *Tool) callStackToolValue(value Value) Value {
	if isNil(value) {
		return nil
	}
	if t == nil || t.CallStack == nil || value == nil {
		return value
	}
	switch v := value.(type) {
	case *SetPredValue:
		return NewSetPredValueFrom(v, t)
	case *FcnLambdaValue:
		return NewFcnLambdaValueFrom(v, t)
	case *OpLambdaValue:
		return NewOpLambdaValueFrom(v, t)
	default:
		return value
	}
}

func (s *CallStack) Push(expr SemanticNode) {
	s.require()
	if s.index == len(s.stack) {
		resized := make([]SemanticNode, len(s.stack)*2)
		copy(resized, s.stack)
		s.stack = resized
	}
	index := s.index
	s.index = int(int32(s.index) + 1)
	if index < 0 || index >= len(s.stack) {
		panic(NewArrayIndexOutOfBoundsException(index, len(s.stack)))
	}
	s.stack[index] = expr
}

func (s *CallStack) Pop() {
	s.require()
	if !s.frozen {
		s.index = int(int32(s.index) - 1)
	}
}

func (s *CallStack) Freeze(errors ...*FingerprintException) {
	s.require()
	if len(errors) > 1 {
		panic(NewIllegalArgumentException("Freeze accepts at most one fingerprint exception"))
	}
	if s.frozen {
		return
	}
	s.frozen = true
	if len(errors) == 0 {
		return
	}
	if errors[0] == nil {
		panic(NewNullPointerException())
	}
	for _, node := range errors[0].AsTrace() {
		s.Push(node)
	}
}

func (s *CallStack) Size() int {
	s.require()
	return s.index
}

func (s *CallStack) String() string {
	s.require()
	if s.index <= 0 {
		return "    The error call stack is empty.\n"
	}
	var b strings.Builder
	var prev SemanticNode
	depth := 0
	for i := 0; i < s.index; i++ {
		if i < 0 || i >= len(s.stack) {
			panic(NewArrayIndexOutOfBoundsException(i, len(s.stack)))
		}
		node := s.stack[i]
		if semanticNodeSame(node, prev) {
			continue
		}
		prev = node
		b.WriteString(strconv.Itoa(depth))
		b.WriteString(". ")
		b.WriteString(callStackLocationString(node))
		b.WriteString("\n")
		depth++
	}
	b.WriteString("\n")
	return b.String()
}

func callStackLocationString(node SemanticNode) string {
	if node == nil {
		panic(NewNullPointerException())
	}
	if syntax, ok := node.(interface{ GetTreeNode() any }); ok && semanticExploreNull(syntax.GetTreeNode()) {
		panic(NewNullPointerException())
	}
	loc, ok := semanticNodeSourceLocation(node)
	if !ok {
		loc = NullSourceLocation
	}
	var b strings.Builder
	b.WriteString("Line ")
	b.WriteString(strconv.Itoa(loc.BeginLine))
	b.WriteString(", column ")
	b.WriteString(strconv.Itoa(loc.BeginColumn))
	b.WriteString(" to line ")
	b.WriteString(strconv.Itoa(loc.EndLine))
	b.WriteString(", column ")
	b.WriteString(strconv.Itoa(loc.EndColumn))
	b.WriteString(" in ")
	b.WriteString(loc.Source)
	return b.String()
}

func (s *CallStack) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

type FingerprintException struct {
	throwableTrace
	Value Value
	Next  *FingerprintException
	Cause error

	ownersCaptured bool
	valueOwner     Value
	nextOwner      *FingerprintException
}

func FingerprintExceptionGetNewHead(value Value, cause error) *FingerprintException {
	if existing, ok := cause.(*FingerprintException); ok && existing != nil {
		return existing.PrependNewHead(value)
	}
	return NewFingerprintException(value, cause)
}

func NewFingerprintExceptionHead(value Value, cause error) *FingerprintException {
	return FingerprintExceptionGetNewHead(value, cause)
}

func NewFingerprintException(value Value, cause error) *FingerprintException {
	if isNil(value) || isNil(cause) {
		return nil
	}
	return &FingerprintException{
		throwableTrace: captureThrowableTrace(), Value: value, Cause: cause,
		ownersCaptured: true, valueOwner: value,
	}
}

func (e *FingerprintException) PrependNewHead(value Value) *FingerprintException {
	if e == nil {
		panic(NewNullPointerException())
	}
	if isNil(value) {
		return nil
	}
	return &FingerprintException{
		throwableTrace: captureThrowableTrace(), Value: value, Next: e,
		ownersCaptured: true, valueOwner: value, nextOwner: e,
	}
}

func (e *FingerprintException) ownedValue() Value {
	if e.ownersCaptured {
		return e.valueOwner
	}
	return e.Value
}

func (e *FingerprintException) ownedNext() *FingerprintException {
	if e.ownersCaptured {
		return e.nextOwner
	}
	return e.Next
}

func (e *FingerprintException) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if next := e.ownedNext(); next != nil {
		return next.Error()
	}
	if value := e.ownedValue(); value != nil {
		return "fingerprint error at " + value.String()
	}
	return "fingerprint error"
}

func (e *FingerprintException) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.Cause != nil {
		return e.Cause
	}
	if next := e.ownedNext(); next != nil {
		return next
	}
	return nil
}

func (e *FingerprintException) GetRootCause() error {
	e.require()
	ptr := e
	for ptr.ownedNext() != nil {
		ptr = ptr.ownedNext()
	}
	return ptr.Cause
}

func (e *FingerprintException) GetTrace() string {
	e.require()
	return e.getTraceImpl(0, nil)
}

func (e *FingerprintException) getTraceImpl(traceIndexLabel int, last SemanticNode) string {
	if e == nil {
		return ""
	}
	node := e.traceNode()
	if node == nil || fingerprintSemanticNodeSame(node, last) {
		next := e.ownedNext()
		if next == nil {
			return ""
		}
		return next.getTraceImpl(traceIndexLabel, last)
	}
	description := strconv.Itoa(traceIndexLabel) + ") " + SemanticString(node) + "\n"
	next := e.ownedNext()
	if next == nil {
		return description
	}
	return next.getTraceImpl(traceIndexLabel+1, node) + description
}

func (e *FingerprintException) traceNode() SemanticNode {
	return valueSource(e.ownedValue())
}

func (e *FingerprintException) AsTrace() []SemanticNode {
	e.require()
	out := make([]SemanticNode, 0)
	for ptr := e; ptr != nil && ptr.ownedValue() != nil; ptr = ptr.ownedNext() {
		out = append(out, ptr.traceNode())
	}
	return out
}

func fingerprintSemanticNodeSame(left SemanticNode, right SemanticNode) bool {
	if left == nil || right == nil {
		return left == right
	}
	leftUID, leftOK := left.(interface{ GetUid() int32 })
	rightUID, rightOK := right.(interface{ GetUid() int32 })
	if leftOK && rightOK {
		return leftUID.GetUid() == rightUID.GetUid()
	}
	leftCompat, leftOK := left.(interface{ GetUID() int32 })
	rightCompat, rightOK := right.(interface{ GetUID() int32 })
	if leftOK && rightOK {
		return leftCompat.GetUID() == rightCompat.GetUID()
	}
	return semanticNodeSame(left, right)
}

func (e *FingerprintException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func valueSource(value Value) SemanticNode {
	if isNil(value) {
		return nil
	}
	return ValueGetSource(value)
}

// The Java constructor calls initCause without setting a detail message.
func (e *FingerprintException) GetMessage() *string {
	e.require()
	return nil
}
