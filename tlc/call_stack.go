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
				t.CallStack.Freeze()
			}
		}
		t.CallStack.Pop()
	}
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
		if node == prev {
			continue
		}
		prev = node
		b.WriteString(strconv.Itoa(depth))
		b.WriteString(". ")
		b.WriteString(SemanticString(node))
		b.WriteString("\n")
		depth++
	}
	b.WriteString("\n")
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
	var nodes []string
	for ptr := e; ptr != nil; ptr = ptr.Next {
		if ptr.Node != nil {
			nodes = append(nodes, SemanticString(ptr.Node))
		} else if ptr.Value != nil {
			nodes = append(nodes, ptr.Value.String())
		}
	}
	var b strings.Builder
	last := ""
	label := 0
	for i := len(nodes) - 1; i >= 0; i-- {
		if nodes[i] == "" || nodes[i] == last {
			continue
		}
		b.WriteString(strconv.Itoa(label))
		b.WriteString(") ")
		b.WriteString(nodes[i])
		b.WriteString("\n")
		last = nodes[i]
		label++
	}
	return b.String()
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
