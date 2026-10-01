package tlc

import (
	"fmt"
	"sync"
)

type TLAPlusExecutor struct {
	mu    sync.Mutex
	Tool  *Tool
	Defns *Defns
	State *TLCStateMut
}

type TLAPlusMapping struct {
	Action       *Action
	Params       *InsMap[*SymbolNode, Value]
	SymbolToNode *InsMap[string, *SymbolNode]
}

func NewTLAPlusExecutor(tool *Tool, defns ...*Defns) (*TLAPlusExecutor, error) {
	if tool == nil {
		return nil, newTLCError(ECGeneral, "TLAPlusExecutor requires a tool")
	}
	tool.SetMode(ModeExecutor)
	SetTLCStateTool(tool)
	vec := NewStateVec(0)
	err := tool.GetInitStates(NewStateFunctor(func(state *TLCStateMut) (any, error) {
		vec.Add(state)
		return nil, nil
	}))
	if err != nil {
		return nil, err
	}
	if vec.Size() == 0 {
		return nil, newTLCError(ECGeneral, "TLAPlusExecutor could not construct an initial state")
	}
	exec := &TLAPlusExecutor{Tool: tool, State: vec.First()}
	if len(defns) > 0 {
		exec.Defns = defns[0]
	}
	return exec, nil
}

func NewTLAPlusMapping(action *Action, params ...string) *TLAPlusMapping {
	mapping := &TLAPlusMapping{
		Action:       action,
		Params:       NewInsMap[*SymbolNode, Value](),
		SymbolToNode: NewInsMap[string, *SymbolNode](),
	}
	for _, param := range params {
		var node *SymbolNode
		if action != nil && action.Con != nil {
			node = action.Con.LookupName(func(sym *SymbolNode) bool {
				return sym != nil && sym.Name != nil && sym.Name.String() == param
			})
		}
		mapping.SymbolToNode.Set(param, node)
	}
	return mapping
}

func (m *TLAPlusMapping) Set(param string, value Value) *TLAPlusMapping {
	if m == nil {
		return nil
	}
	node := m.SymbolToNode.Get(param)
	if node != nil {
		m.Params.Set(node, value)
	}
	return m
}

func (m *TLAPlusMapping) Context() *Context {
	if m == nil || m.Action == nil {
		return EmptyContext
	}
	ctx := m.Action.Con
	if ctx == nil {
		ctx = EmptyContext
	}
	for node, value := range m.Params.All() {
		ctx = ctx.Cons(node, value)
	}
	return ctx
}

func (e *TLAPlusExecutor) Map(actionName string, processNode string, value Value, params ...string) (*TLAPlusMapping, error) {
	if e == nil || e.Tool == nil {
		return nil, newTLCError(ECGeneral, "TLAPlusExecutor has no tool")
	}
	for _, action := range e.Tool.GetActions() {
		if action == nil || action.GetName() != actionName {
			continue
		}
		lookup := action.Con.LookupName(func(sym *SymbolNode) bool {
			return sym != nil && sym.Name != nil && sym.Name.String() == processNode
		})
		if lookup == nil {
			continue
		}
		bound, ok := action.Con.Lookup(lookup).(Value)
		if !ok {
			continue
		}
		eq, err := bound.Equal(value)
		if err != nil {
			return nil, err
		}
		if eq {
			return NewTLAPlusMapping(action, params...), nil
		}
	}
	return nil, newTLCError(ECGeneral, "failed to create mapping")
}

func (e *TLAPlusExecutor) Step(mapping *TLAPlusMapping) (any, error) {
	if e == nil || e.Tool == nil || mapping == nil || mapping.Action == nil {
		return nil, newTLCError(ECGeneral, "TLAPlusExecutor step requires a tool and mapping")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for {
		action := *mapping.Action
		action.Con = mapping.Context()
		nextStates, err := e.Tool.GetNextStates(&action, e.State)
		if err != nil {
			return nil, err
		}
		if nextStates == nil || nextStates.IsEmpty() {
			return nil, newTLCError(ECGeneral, "TLAPlusExecutor action %s produced no next states", action.GetName())
		}
		e.State = nextStates.First()
		for _, inv := range e.Tool.GetInvariants() {
			ok, err := e.Tool.IsValidState(inv, e.State)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errInvariantViolated
			}
		}
		result, err := e.State.ExecCallable()
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}
}

func (e *TLAPlusExecutor) Lock() *sync.Mutex {
	if e == nil {
		return nil
	}
	return &e.mu
}

func (e *TLAPlusExecutor) GetConstant(name string) Value {
	if e == nil {
		return nil
	}
	if e.Defns != nil {
		if value, ok := e.Defns.Get(name).(Value); ok {
			return value
		}
	}
	if e.Tool != nil && e.Tool.DefnsByName != nil {
		if value, ok := e.Tool.DefnsByName[UniqueStringOf(name)].(Value); ok {
			return value
		}
	}
	if e.Tool != nil {
		if value, ok := e.Tool.Lookup(NewSymbolNode(name), EmptyContext, EmptyState, false).(Value); ok {
			return value
		}
	}
	return nil
}

func (m *TLAPlusMapping) String() string {
	if m == nil || m.Action == nil {
		return "<mapping>"
	}
	return fmt.Sprintf("%s%s", m.Action.GetName(), m.Params)
}
