package tlc

import "fmt"

// TLCApp is the runtime portion of Java's distributed application. The root
// loader builds its Tool through the production parser before this snapshot.
type TLCApp struct {
	Tool           *Tool
	Actions        []*Action
	Invariants     []*Action
	ImpliedInits   []*Action
	ImpliedActions []*Action
	checkDeadlock  bool
}

func NewTLCApp(tool *Tool, deadlock bool) *TLCApp {
	if tool == nil {
		panic(NewNullPointerException())
	}
	app := &TLCApp{Tool: tool, checkDeadlock: deadlock}
	app.ImpliedInits = tool.GetImpliedInits()
	app.Invariants = tool.GetInvariants()
	app.ImpliedActions = tool.GetImpliedActions()
	app.Actions = tool.GetActions()
	return app
}

func (a *TLCApp) GetCheckDeadlock() bool {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a.checkDeadlock
}
func (a *TLCApp) GetPreprocess() bool {
	if a == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (a *TLCApp) requireTool() *Tool {
	if a == nil || a.Tool == nil {
		panic(NewNullPointerException())
	}
	return a.Tool
}
func (a *TLCApp) GetInitStates(functor *StateFunctor) error {
	return a.requireTool().GetInitStates(functor)
}

func (a *TLCApp) GetNextStates(state *TLCStateMut) (*StateVec, error) {
	if a == nil {
		panic(NewNullPointerException())
	}
	out := NewStateVec(10)
	for _, action := range a.Actions {
		next, err := a.requireTool().GetNextStates(action, state)
		if err != nil {
			return nil, err
		}
		if next == nil {
			panic(NewNullPointerException())
		}
		for i := 0; i < next.Size(); i++ {
			if successor := next.At(i); successor != nil {
				successor.attachTraceMetadata(state, action)
			}
		}
		// Java addElements chooses the larger vector as the receiver. This
		// changes order when a later action generates more states.
		out = out.AddElements(next)
	}
	if out.Size() == 0 && a.checkDeadlock {
		return nil, NewWorkerException("Error: deadlock reached.", state, nil, false)
	}
	for i := 0; i < out.Size(); i++ {
		successor := out.At(i)
		if !a.requireTool().IsGoodState(successor) {
			return nil, NewWorkerException("Error: Successor state is not completely specified by the next-state action.", state, successor, false)
		}
	}
	return out, nil
}

func (a *TLCApp) CheckState(predecessor, successor *TLCStateMut) error {
	if a == nil {
		panic(NewNullPointerException())
	}
	for i, invariant := range a.Invariants {
		valid, err := a.requireTool().IsValidState(invariant, successor)
		if err != nil {
			return err
		}
		if !valid {
			return NewWorkerException(fmt.Sprintf("Error: Invariant %s is violated.", distributedActionName(a.Tool.GetInvNames(), i)), predecessor, successor, false)
		}
	}
	if predecessor == nil {
		for i, implied := range a.ImpliedInits {
			valid, err := a.requireTool().IsValidState(implied, successor)
			if err != nil {
				return err
			}
			if !valid {
				return NewWorkerException(fmt.Sprintf("Error: Implied-init %s is violated.", distributedActionName(a.Tool.GetImpliedInitNames(), i)), predecessor, successor, false)
			}
		}
	} else {
		for i, implied := range a.ImpliedActions {
			valid, err := a.requireTool().IsValidTransition(implied, predecessor, successor)
			if err != nil {
				return err
			}
			if !valid {
				return NewWorkerException(fmt.Sprintf("Error: Implied-action %s is violated.", distributedActionName(a.Tool.GetImpliedActNames(), i)), predecessor, successor, false)
			}
		}
	}
	return nil
}

func distributedActionName(names []string, index int) string {
	if index < 0 || index >= len(names) {
		panic(NewArrayIndexOutOfBoundsException(index, len(names)))
	}
	return names[index]
}

func (a *TLCApp) IsInModel(state *TLCStateMut) (bool, error) { return a.requireTool().IsInModel(state) }
func (a *TLCApp) IsInActions(predecessor, successor *TLCStateMut) (bool, error) {
	return a.requireTool().IsInActions(predecessor, successor)
}
func (a *TLCApp) GetState(fp uint64, previous ...any) (*TLCStateInfo, error) {
	return a.requireTool().GetState(fp, previous...)
}
func (a *TLCApp) GetStateForTransition(successor, predecessor *TLCStateMut) (*TLCStateInfo, error) {
	return a.requireTool().GetStateForTransition(successor, predecessor)
}
func (a *TLCApp) EvalAlias(current *TLCStateInfo, successor *TLCStateMut, prefix func() []*TLCStateInfo) (*TLCStateInfo, error) {
	return a.requireTool().EvalAliasInfo(current, successor, prefix)
}
func (a *TLCApp) SetCallStack() { a.Tool = NewCallStackTool(a.requireTool()) }
func (a *TLCApp) PrintCallStack() string {
	if a.requireTool().CallStack != nil {
		return a.Tool.CallStackString()
	}
	return fmt.Sprintf("%T@%p", a.Tool, a.Tool)
}
