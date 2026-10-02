package tlc

import (
	"errors"
	"strconv"
	"strings"
)

type CallStack struct {
	stack  []SemanticNode
	frozen bool
}

func NewCallStack() *CallStack {
	return &CallStack{stack: make([]SemanticNode, 0, 64)}
}

func NewCallStackTool(other *Tool) *Tool {
	if other == nil {
		other = NewTool()
	}
	tool := *other
	tool.CallStack = NewCallStack()
	return &tool
}

func (t *Tool) HasCallStack() bool {
	return t != nil && t.CallStack != nil && t.CallStack.Size() > 0
}

func (t *Tool) CallStackString() string {
	if t == nil || t.CallStack == nil {
		return NewCallStack().String()
	}
	return t.CallStack.String()
}

func (t *Tool) callStackEnter(expr SemanticNode) func(error) {
	if t == nil || t.CallStack == nil || expr == nil {
		return func(error) {}
	}
	t.CallStack.Push(expr)
	return func(err error) {
		if err != nil {
			var fpErr *FingerprintException
			if errors.As(err, &fpErr) {
				t.CallStack.Freeze(fpErr)
			} else {
				var runtimeErr *TLCError
				var evalErr *EvalException
				if errors.As(err, &runtimeErr) || errors.As(err, &evalErr) {
					t.CallStack.Freeze()
				}
			}
		}
		t.CallStack.Pop()
	}
}

// setValueSource mirrors Tool.setSource: FastTool and DebugTool do nothing,
// while CallStackTool replaces the source even if a value already has one.
func (t *Tool) setValueSource(expr SemanticNode, value Value) Value {
	if t != nil && t.CallStack != nil && value != nil {
		if assignable, ok := value.(sourceAssignableValue); ok {
			assignable.SetSource(expr)
		}
	}
	return value
}

func (t *Tool) callStackToolValue(value Value) Value {
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
	if s == nil {
		return
	}
	s.stack = append(s.stack, expr)
}

func (s *CallStack) Pop() {
	if s == nil || s.frozen || len(s.stack) == 0 {
		return
	}
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *CallStack) Freeze(errors ...*FingerprintException) {
	if s == nil || s.frozen {
		return
	}
	s.frozen = true
	if len(errors) == 0 || errors[0] == nil {
		return
	}
	for _, node := range errors[0].AsTrace() {
		s.Push(node)
	}
}

func (s *CallStack) Size() int {
	if s == nil {
		return 0
	}
	return len(s.stack)
}

func (s *CallStack) String() string {
	if s == nil || len(s.stack) == 0 {
		return "    The error call stack is empty.\n"
	}
	var b strings.Builder
	var prev SemanticNode
	depth := 0
	for _, node := range s.stack {
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

type FingerprintException struct {
	Value Value
	Node  SemanticNode
	Next  *FingerprintException
	Cause error
}

func NewFingerprintExceptionHead(value Value, cause error) *FingerprintException {
	if existing, ok := cause.(*FingerprintException); ok {
		return existing.PrependNewHead(value)
	}
	return NewFingerprintException(value, cause)
}

func NewFingerprintException(value Value, cause error) *FingerprintException {
	if value == nil || cause == nil {
		return nil
	}
	return &FingerprintException{Value: value, Node: valueSource(value), Cause: cause}
}

func (e *FingerprintException) PrependNewHead(value Value) *FingerprintException {
	if value == nil {
		return nil
	}
	return &FingerprintException{Value: value, Node: valueSource(value), Next: e}
}

func (e *FingerprintException) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if e.Value != nil {
		return "fingerprint error at " + e.Value.String()
	}
	if e.Next != nil {
		return e.Next.Error()
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
	if e.Next != nil {
		return e.Next
	}
	return nil
}

func (e *FingerprintException) GetRootCause() error {
	ptr := e
	for ptr != nil && ptr.Next != nil {
		ptr = ptr.Next
	}
	if ptr == nil {
		return nil
	}
	return ptr.Cause
}

func (e *FingerprintException) GetTrace() string {
	if e == nil {
		return ""
	}
	return e.getTraceImpl(0, nil)
}

func (e *FingerprintException) getTraceImpl(traceIndexLabel int, last SemanticNode) string {
	if e == nil {
		return ""
	}
	node := e.Node
	if node == nil || semanticNodeSame(node, last) {
		if e.Next == nil {
			return ""
		}
		return e.Next.getTraceImpl(traceIndexLabel, last)
	}
	description := strconv.Itoa(traceIndexLabel) + ") " + SemanticString(node) + "\n"
	if e.Next == nil {
		return description
	}
	return e.Next.getTraceImpl(traceIndexLabel+1, node) + description
}

func (e *FingerprintException) AsTrace() []SemanticNode {
	var out []SemanticNode
	for ptr := e; ptr != nil; ptr = ptr.Next {
		if ptr.Node != nil {
			out = append(out, ptr.Node)
		}
	}
	return out
}

type sourcedValue interface {
	GetSource() SemanticNode
}

func valueSource(value Value) SemanticNode {
	if value == nil {
		return nil
	}
	if sourced, ok := value.(sourcedValue); ok {
		return sourced.GetSource()
	}
	return nil
}
