package tlc

import "fmt"

type TLAPlusExecutor struct {
	mu    reentrantReadWriteLock
	Tool  *Tool
	Defns *Defns
	State *TLCStateMut

	toolCaptured bool
	toolOwner    *Tool
}

type TLAPlusMapping struct {
	Action       *Action
	Params       *InsMap[*SymbolNode, Value]
	SymbolToNode *InsMap[string, *SymbolNode]
	nullParam    Value
	hasNullParam bool

	ownersCaptured    bool
	actionOwner       *Action
	paramsOwner       *InsMap[*SymbolNode, Value]
	symbolToNodeOwner *InsMap[string, *SymbolNode]
}

func NewTLAPlusExecutor(tool *Tool, defns ...*Defns) (*TLAPlusExecutor, error) {
	if len(defns) > 1 {
		panic(NewIllegalArgumentException("NewTLAPlusExecutor accepts at most one definition table"))
	}
	if tool == nil {
		return nil, newTLCError(ECGeneral, "TLAPlusExecutor requires a tool")
	}
	tool.SetMode(ModeExecutor)
	SetTLCStateTool(tool)
	vec := NewStateVec(0)
	err := tool.GetInitStates(NewStateFunctor(func(state *TLCStateMut) (any, error) {
		vec.Add(state)
		return vec, nil
	}))
	if err != nil {
		return nil, err
	}
	exec := &TLAPlusExecutor{Tool: tool, State: vec.First(), toolCaptured: true, toolOwner: tool}
	if len(defns) > 0 {
		exec.Defns = defns[0]
	}
	return exec, nil
}

func (e *TLAPlusExecutor) ownedTool() *Tool {
	if e.toolCaptured {
		return e.toolOwner
	}
	return e.Tool
}

func NewTLAPlusMapping(action *Action, params ...string) *TLAPlusMapping {
	values := NewInsMap[*SymbolNode, Value]()
	symbols := NewInsMap[string, *SymbolNode]()
	mapping := &TLAPlusMapping{
		Action:            action,
		Params:            values,
		SymbolToNode:      symbols,
		ownersCaptured:    true,
		actionOwner:       action,
		paramsOwner:       values,
		symbolToNodeOwner: symbols,
	}
	if action == nil && len(params) > 0 {
		panic(NewNullPointerException(`Cannot read field "con" because "this.action" is null`))
	}
	if action != nil && action.ownedContext() == nil && len(params) > 0 {
		panic(NewNullPointerException(`Cannot invoke "tlc2.util.Context.lookupName(java.util.function.Function)" because "this.action.con" is null`))
	}
	for _, param := range params {
		var node *SymbolNode
		if action != nil {
			node = action.ownedContext().LookupName(func(sym *SymbolNode) bool {
				if sym == nil {
					panic(NewNullPointerException(`Cannot invoke "tla2sany.semantic.SymbolNode.getName()" because "s" is null`))
				}
				if sym.Name == nil {
					panic(NewNullPointerException(`Cannot invoke "util.UniqueString.equals(String)" because the return value of "tla2sany.semantic.SymbolNode.getName()" is null`))
				}
				return sym.Name.String() == param
			})
		}
		mapping.ownedSymbolToNode().Set(param, node)
	}
	return mapping
}

func (m *TLAPlusMapping) ownedAction() *Action {
	if m.ownersCaptured {
		return m.actionOwner
	}
	return m.Action
}

func (m *TLAPlusMapping) ownedParams() *InsMap[*SymbolNode, Value] {
	if m.ownersCaptured {
		return m.paramsOwner
	}
	return m.Params
}

func (m *TLAPlusMapping) ownedSymbolToNode() *InsMap[string, *SymbolNode] {
	if m.ownersCaptured {
		return m.symbolToNodeOwner
	}
	return m.SymbolToNode
}

func (m *TLAPlusMapping) Set(param string, value Value) *TLAPlusMapping {
	if m == nil {
		panic(NewNullPointerException())
	}
	node := m.ownedSymbolToNode().Get(param)
	if node == nil {
		m.nullParam = value
		m.hasNullParam = true
	} else {
		m.ownedParams().Set(node, value)
	}
	return m
}

func (m *TLAPlusMapping) Context() *Context {
	if m == nil {
		panic(NewNullPointerException())
	}
	action := m.ownedAction()
	if action == nil {
		panic(NewNullPointerException(`Cannot read field "con" because "this.action" is null`))
	}
	ctx := action.ownedContext()
	if m.hasNullParam {
		if ctx == nil {
			panic(NewNullPointerException(`Cannot invoke "tlc2.util.Context.cons(tla2sany.semantic.SymbolNode, Object)" because "ctx" is null`))
		}
		ctx = ctx.Cons(nil, m.nullParam)
	}
	for node, value := range m.ownedParams().All() {
		ctx = ctx.Cons(node, value)
	}
	return ctx
}

func (e *TLAPlusExecutor) Map(actionName string, processNode string, value Value, params ...string) (*TLAPlusMapping, error) {
	if e == nil {
		return nil, NewNullPointerException()
	}
	tool := e.ownedTool()
	if tool == nil {
		return nil, newTLCError(ECGeneral, "TLAPlusExecutor has no tool")
	}
	for _, action := range tool.GetActions() {
		if action == nil {
			return nil, NewNullPointerException(`Cannot invoke "tlc2.tool.Action.getName()" because "action" is null`)
		}
		name := action.GetNameUnique()
		if name == nil {
			return nil, NewNullPointerException()
		}
		if name.String() != actionName {
			continue
		}
		con := action.ownedContext()
		if con == nil {
			return nil, NewNullPointerException(`Cannot invoke "tlc2.util.Context.lookup(java.util.function.Function)" because "action.con" is null`)
		}
		var lookupFailure error
		lookup := con.LookupName(func(sym *SymbolNode) bool {
			if lookupFailure != nil {
				return false
			}
			if sym == nil {
				lookupFailure = NewNullPointerException(`Cannot invoke "tla2sany.semantic.SymbolNode.getName()" because "s" is null`)
				return false
			}
			if sym.Name == nil {
				lookupFailure = NewNullPointerException(`Cannot invoke "util.UniqueString.equals(String)" because the return value of "tla2sany.semantic.SymbolNode.getName()" is null`)
				return false
			}
			return sym.Name.String() == processNode
		})
		if lookupFailure != nil {
			return nil, lookupFailure
		}
		if lookup == nil {
			continue
		}
		bound, ok := con.Lookup(lookup).(Value)
		if !ok {
			continue
		}
		if value == nil {
			return nil, NewNullPointerException(`Cannot invoke "Object.toString()" because "obj" is null`)
		}
		eq, err := bound.Equal(value)
		if err != nil {
			return nil, err
		}
		if eq {
			return NewTLAPlusMapping(action, params...), nil
		}
	}
	return nil, NewRuntimeException("failed to create mapping")
}

func (e *TLAPlusExecutor) Step(mapping *TLAPlusMapping) (any, error) {
	if e == nil {
		return nil, NewNullPointerException()
	}
	tool := e.ownedTool()
	if tool == nil {
		return nil, newTLCError(ECGeneral, "TLAPlusExecutor step requires a tool and mapping")
	}
	if mapping == nil {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.tool.TLAPlusExecutor$Mapping.getContext()" because "m" is null`)
	}
	action := mapping.ownedAction()
	if action == nil {
		return nil, NewNullPointerException(`Cannot read field "con" because "this.action" is null`)
	}
	ctx := mapping.Context()
	e.mu.Lock()
	defer e.mu.Unlock()
	for {
		nextStates, err := tool.getNextStatesWithContext(action, ctx, e.State)
		if err != nil {
			return nil, err
		}
		if nextStates == nil || nextStates.IsEmpty() {
			return nil, NewAssertionError()
		}
		e.State = nextStates.First()
		for _, inv := range tool.GetInvariants() {
			ok, err := tool.IsValidState(inv, e.State)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, NewInvariantViolatedException()
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

func (e *TLAPlusExecutor) GetLock() *reentrantReadWriteLock {
	if e == nil {
		panic(NewNullPointerException())
	}
	return &e.mu
}

func (e *TLAPlusExecutor) Lock() *reentrantReadWriteLock { return e.GetLock() }

func (e *TLAPlusExecutor) GetConstant(name string) Value {
	if e == nil {
		panic(NewNullPointerException())
	}
	if e.Defns != nil {
		candidate := e.Defns.Get(name)
		if value, ok := candidate.(Value); ok {
			return value
		}
		if candidate != nil {
			panic(valueStreamClassCast(candidate, "tlc2.value.impl.Value"))
		}
	}
	tool := e.ownedTool()
	if tool != nil && tool.DefnsByName != nil {
		candidate := tool.DefnsByName[UniqueStringOf(name)]
		if value, ok := candidate.(Value); ok {
			return value
		}
		if candidate != nil {
			panic(valueStreamClassCast(candidate, "tlc2.value.impl.Value"))
		}
	}
	if tool != nil {
		candidate := tool.Lookup(NewSymbolNode(name), EmptyContext, EmptyState, false)
		if value, ok := candidate.(Value); ok {
			return value
		}
		if candidate != nil {
			panic(valueStreamClassCast(candidate, "tlc2.value.impl.Value"))
		}
	}
	return nil
}

func (m *TLAPlusMapping) String() string {
	if m == nil || m.ownedAction() == nil {
		return "<mapping>"
	}
	return fmt.Sprintf("%s%s", m.ownedAction().GetName(), m.ownedParams())
}
