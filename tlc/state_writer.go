package tlc

type StateVisitStatus int

const (
	StateVisitUnseen StateVisitStatus = iota
	StateVisitSeen
	StateVisitNotInModel
)

type StateWriter struct {
	Constrained         bool
	WriteInitStateFunc  func(*TLCStateMut) error
	WriteTransitionFunc func(*TLCStateMut, *TLCStateMut, StateVisitStatus, *Action, ...SemanticNode) error
	CloseFunc           func() error
}

func NewNoopStateWriter() *StateWriter {
	return &StateWriter{}
}

func (w *StateWriter) WriteInitState(state *TLCStateMut) error {
	if w != nil && w.WriteInitStateFunc != nil {
		return w.WriteInitStateFunc(state)
	}
	return nil
}

func (w *StateWriter) WriteTransition(curState *TLCStateMut, succState *TLCStateMut, status StateVisitStatus, action *Action, reason ...SemanticNode) error {
	if w != nil && w.WriteTransitionFunc != nil {
		return w.WriteTransitionFunc(curState, succState, status, action, reason...)
	}
	return nil
}

func (w *StateWriter) IsConstrained() bool {
	return w != nil && w.Constrained
}

func (w *StateWriter) Close() error {
	if w != nil && w.CloseFunc != nil {
		return w.CloseFunc()
	}
	return nil
}
