package tlc

import (
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

func (s *CallStack) Freeze(errors ...*FingerprintError) {
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

type StatefulRuntimeError struct {
	Message string
	Cause   error
	known   bool
}

func NewStatefulRuntimeError(message string, cause ...error) *StatefulRuntimeError {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}
	return &StatefulRuntimeError{Message: message, Cause: c}
}

func (e *StatefulRuntimeError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "stateful runtime error"
}

func (e *StatefulRuntimeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *StatefulRuntimeError) SetKnown() bool {
	if e == nil {
		return false
	}
	old := e.known
	e.known = true
	return old
}

func (e *StatefulRuntimeError) IsKnown() bool {
	return e != nil && e.known
}

type FingerprintError struct {
	Value Value
	Node  SemanticNode
	Next  *FingerprintError
	Cause error
}

func NewFingerprintErrorHead(value Value, cause error) *FingerprintError {
	if existing, ok := cause.(*FingerprintError); ok {
		return existing.Prepend(value)
	}
	if value == nil || cause == nil {
		return nil
	}
	return &FingerprintError{Value: value, Cause: cause}
}

func (e *FingerprintError) Prepend(value Value) *FingerprintError {
	if value == nil {
		return nil
	}
	return &FingerprintError{Value: value, Next: e}
}

func (e *FingerprintError) Error() string {
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

func (e *FingerprintError) Unwrap() error {
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

func (e *FingerprintError) RootCause() error {
	ptr := e
	for ptr != nil && ptr.Next != nil {
		ptr = ptr.Next
	}
	if ptr == nil {
		return nil
	}
	return ptr.Cause
}

func (e *FingerprintError) Trace() string {
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

func (e *FingerprintError) AsTrace() []SemanticNode {
	var out []SemanticNode
	for ptr := e; ptr != nil; ptr = ptr.Next {
		if ptr.Node != nil {
			out = append(out, ptr.Node)
		}
	}
	return out
}
