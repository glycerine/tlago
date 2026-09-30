package tlc

import "strings"

var (
	counterExampleStates  = UniqueStringOf("state")
	counterExampleActions = UniqueStringOf("action")
	counterExampleConsole = UniqueStringOf("console")
	actionRecordName      = UniqueStringOf("name")
	actionRecordLocation  = UniqueStringOf("location")
	actionRecordContext   = UniqueStringOf("context")
	actionRecordParams    = UniqueStringOf("parameters")
)

type CounterExample struct {
	*RecordValue
}

func NewCounterExample(trace []*TLCStateInfo, action *Action, loopOrdinal int, isConsole bool) *CounterExample {
	if action == nil {
		action = UnknownAction
	}
	loopIdx := loopOrdinal - 1
	stateNodes := make([]Value, 0, len(trace))
	actionEdges := make([]Value, 0, len(trace))
	for i, info := range trace {
		stateNumber := int32(0)
		if info != nil {
			stateNumber = int32(info.GetStateNumber())
		}
		node := NewTupleValue([]Value{NewIntValue(stateNumber), stateInfoRecordValue(info)})
		if i > 0 {
			edgeAction := UnknownAction
			if info != nil {
				edgeAction = info.Action()
			}
			edge := NewTupleValue([]Value{stateNodes[len(stateNodes)-1], edgeAction.ToRecordValue(), node})
			actionEdges = append(actionEdges, edge)
		}
		stateNodes = append(stateNodes, node)
	}
	if loopOrdinal > 0 && loopIdx >= 0 && loopIdx < len(stateNodes) && len(stateNodes) != 0 {
		edge := NewTupleValue([]Value{stateNodes[len(stateNodes)-1], action.ToRecordValue(), stateNodes[loopIdx]})
		actionEdges = append(actionEdges, edge)
	}
	names := []*UniqueString{counterExampleActions, counterExampleStates}
	values := []Value{
		NewSetEnumValue(actionEdges, false),
		NewSetEnumValue(stateNodes, false),
	}
	if !isConsole {
		names = append(names, counterExampleConsole)
		values = append(values, BoolFalse)
	}
	return &CounterExample{RecordValue: NewRecordValue(names, values, false)}
}

func NewEmptyCounterExample() *CounterExample {
	return NewCounterExample(nil, UnknownAction, 0, true)
}

func NewCounterExampleFromTrace(trace []*TLCStateInfo) *CounterExample {
	return NewCounterExample(trace, UnknownAction, 0, true)
}

func NewCounterExampleFromInitialState(state *TLCStateMut) *CounterExample {
	return NewCounterExample([]*TLCStateInfo{NewTLCStateInfo(state)}, UnknownAction, 0, true)
}

func NewCounterExampleFromStateVec(vec *StateVec) *CounterExample {
	if vec == nil || vec.Size() == 0 {
		return NewEmptyCounterExample()
	}
	trace := make([]*TLCStateInfo, 0, vec.Size())
	for i := 0; i < vec.Size(); i++ {
		trace = append(trace, NewTLCStateInfo(vec.At(i)))
	}
	return NewCounterExampleFromTrace(trace)
}

func (c *CounterExample) ToTrace() Value {
	if c == nil || c.RecordValue == nil {
		return EmptyTuple
	}
	value := c.Select(NewStringValueFromUnique(counterExampleStates))
	set, ok := value.(*SetEnumValue)
	if !ok {
		return EmptyTuple
	}
	values := make([]Value, set.Elems.Len())
	for i := 0; i < set.Elems.Len(); i++ {
		tv, ok := set.Elems.At(i).(*TupleValue)
		if !ok || len(tv.Elems) < 2 {
			continue
		}
		ordinal, ok := tv.Elems[0].(*IntValue)
		if !ok {
			continue
		}
		idx := int(ordinal.Val) - 1
		if idx >= 0 && idx < len(values) {
			values[idx] = tv.Elems[1]
		}
	}
	return NewTupleValue(values)
}

func stateInfoRecordValue(info *TLCStateInfo) Value {
	if info == nil {
		return EmptyRecord
	}
	return info.ToRecordValue()
}

func (a *Action) GetLocation() string {
	if a == nil {
		return ""
	}
	name := "Action"
	if a.IsNamed() {
		name = a.GetName()
	}
	params := a.GetParameters()
	if params.Len() == 0 {
		return "<" + name + " " + a.GetDefinition() + ">"
	}
	parts := make([]string, 0, params.Len())
	for _, value := range params.All() {
		parts = append(parts, value.String())
	}
	return "<" + name + "(" + strings.Join(parts, ",") + ") " + a.GetDefinition() + ">"
}

func (a *Action) ToRecordValue() *RecordValue {
	return actionRecordValue(a, nil, nil)
}

func actionRecordValue(a *Action, extraName *UniqueString, extraValue Value) *RecordValue {
	if a == nil {
		a = UnknownAction
	}
	names := []*UniqueString{actionRecordName, actionRecordLocation}
	values := []Value{NewStringValue(a.GetName()), NewStringValue(a.GetLocation())}
	if extraName != nil {
		names = append(names, extraName)
		values = append(values, extraValue)
	}
	params := a.GetParameters()
	if params.Len() != 0 {
		paramNames := make([]Value, 0, params.Len())
		contextNames := make([]*UniqueString, 0, params.Len())
		contextValues := make([]Value, 0, params.Len())
		for name, value := range params.All() {
			paramNames = append(paramNames, NewStringValueFromUnique(name))
			contextNames = append(contextNames, name)
			contextValues = append(contextValues, value)
		}
		names = append(names, actionRecordContext, actionRecordParams)
		values = append(values, NewRecordValue(contextNames, contextValues, false), NewTupleValue(paramNames))
	}
	return NewRecordValue(names, values, false)
}
