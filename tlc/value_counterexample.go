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
	locationBeginLine     = UniqueStringOf("beginLine")
	locationBeginColumn   = UniqueStringOf("beginColumn")
	locationEndLine       = UniqueStringOf("endLine")
	locationEndColumn     = UniqueStringOf("endColumn")
	locationModule        = UniqueStringOf("module")
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
	if loopOrdinal > 0 {
		if loopIdx < 0 || loopIdx >= len(stateNodes) {
			panic(newTLCError(ECGeneral, "counterexample loop ordinal %d outside trace length %d", loopOrdinal, len(stateNodes)))
		}
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
	value, err := c.Select(NewStringValueFromUnique(counterExampleStates))
	if err != nil {
		return EmptyTuple
	}
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
	return a.GetLocationNamed(name)
}

func (a *Action) GetLocationNamed(name string) string {
	if a == nil {
		return ""
	}
	if name == "" {
		name = "Action"
	}
	parts := a.locationParameters()
	if len(parts) == 0 {
		return "<" + name + " " + a.GetDefinition() + ">"
	}
	return "<" + name + "(" + strings.Join(parts, ",") + ") " + a.GetDefinition() + ">"
}

func (a *Action) locationParameters() []string {
	if a == nil || a.OpDef == nil || a.Con == nil {
		return nil
	}
	parts := make([]string, 0, len(a.OpDef.Params))
	for _, param := range a.OpDef.Params {
		if param == nil {
			continue
		}
		value := a.Con.Lookup(param)
		if value != nil {
			parts = append(parts, toContextString(value))
		}
	}
	return parts
}

func (a *Action) ToRecordValue() *RecordValue {
	return actionRecordValue(a, nil, nil)
}

func actionRecordValue(a *Action, extraName *UniqueString, extraValue Value) *RecordValue {
	if a == nil {
		a = UnknownAction
	}
	names := []*UniqueString{actionRecordName, actionRecordLocation}
	values := []Value{NewStringValue(a.GetName()), sourceLocationRecordValue(a.GetDefinitionLocation())}
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

func sourceLocationRecordValue(location SourceLocation) *RecordValue {
	return NewRecordValue(
		[]*UniqueString{locationBeginLine, locationBeginColumn, locationEndLine, locationEndColumn, locationModule},
		[]Value{
			NewIntValue(int32(location.BeginLine)),
			NewIntValue(int32(location.BeginColumn)),
			NewIntValue(int32(location.EndLine)),
			NewIntValue(int32(location.EndColumn)),
			NewStringValue(location.Source),
		},
		false,
	)
}
