package tlc

import "strings"

var (
	counterExampleStates  *UniqueString
	counterExampleActions *UniqueString
	counterExampleConsole *UniqueString
	actionRecordName      *UniqueString
	actionRecordLocation  *UniqueString
	actionRecordContext   *UniqueString
	actionRecordParams    *UniqueString
	locationBeginLine     *UniqueString
	locationBeginColumn   *UniqueString
	locationEndLine       *UniqueString
	locationEndColumn     *UniqueString
	locationModule        *UniqueString
)

// Refresh Java's class-static names when a fresh worker interning context is
// installed. Their tokens must come from the source rather than the discarded
// Go package-initialization context.
var counterExampleClassInterning classInterning

func initCounterExampleUniqueStrings() {
	counterExampleStates = UniqueStringOf("state")
	counterExampleActions = UniqueStringOf("action")
	counterExampleConsole = UniqueStringOf("console")
	actionRecordName = UniqueStringOf("name")
	actionRecordLocation = UniqueStringOf("location")
	actionRecordContext = UniqueStringOf("context")
	actionRecordParams = UniqueStringOf("parameters")
	locationBeginLine = UniqueStringOf("beginLine")
	locationBeginColumn = UniqueStringOf("beginColumn")
	locationEndLine = UniqueStringOf("endLine")
	locationEndColumn = UniqueStringOf("endColumn")
	locationModule = UniqueStringOf("module")
}

type CounterExample struct {
	*RecordValue
}

func (c *CounterExample) require() *RecordValue {
	if c == nil || c.RecordValue == nil {
		panic(NewNullPointerException())
	}
	return c.RecordValue
}

func (c *CounterExample) Kind() ValueKind    { return c.require().Kind() }
func (c *CounterExample) KindString() string { return ValueKindString(c) }
func (c *CounterExample) GetCostModel() CostModel {
	return c.require().GetCostModel()
}
func (c *CounterExample) DeepNormalize()     { c.require().DeepNormalize() }
func (c *CounterExample) IsNormalized() bool { return c.require().IsNormalized() }
func (c *CounterExample) IsDefined() bool    { return c.require().IsDefined() }
func (c *CounterExample) DeepCopy() Value    { return c.require().DeepCopy() }
func (c *CounterExample) Normalize() Value {
	c.require().Normalize()
	return c
}
func (c *CounterExample) Permute(perm *MVPerm) Value {
	record := c.require()
	result := record.Permute(perm)
	if result == record {
		return c
	}
	return result
}
func (c *CounterExample) Compare(other Value) (int, error) {
	return c.require().Compare(other)
}
func (c *CounterExample) Equal(other Value) (bool, error) {
	return c.require().Equal(other)
}
func (c *CounterExample) Member(elem Value) (bool, error) {
	return c.require().Member(elem)
}
func (c *CounterExample) IsFinite() (bool, error) { return c.require().IsFinite() }
func (c *CounterExample) Size() (int, error)      { return c.require().Size() }
func (c *CounterExample) FingerPrint(fp uint64) uint64 {
	return c.require().FingerPrint(fp)
}
func (c *CounterExample) TakeExcept(ex *ValueExcept) (Value, error) {
	record := c.require()
	result, err := record.TakeExcept(ex)
	if result == record {
		return c, err
	}
	return result, err
}
func (c *CounterExample) TakeExcepts(exs []*ValueExcept) (Value, error) {
	record := c.require()
	result, err := record.TakeExcepts(exs)
	if result == record {
		return c, err
	}
	return result, err
}
func (c *CounterExample) String() string { return c.require().String() }
func (c *CounterExample) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	return c.require().ToString(sb, offset, swallow)
}
func (c *CounterExample) Apply(arg Value) (Value, error) {
	return c.require().Apply(arg)
}
func (c *CounterExample) ApplyWithControl(arg Value, control int) (Value, error) {
	return c.require().ApplyWithControl(arg, control)
}
func (c *CounterExample) ApplyArgs(args []Value, control int) (Value, error) {
	return c.require().ApplyArgs(args, control)
}
func (c *CounterExample) Select(arg Value) (Value, error) {
	return c.require().Select(arg)
}
func (c *CounterExample) GetDomain() (Value, error) {
	return c.require().GetDomain()
}

func NewCounterExample(trace []*TLCStateInfo, action *Action, loopOrdinal int, isConsole bool) *CounterExample {
	ensureCounterExampleUniqueStrings()
	loopOrdinal = int(int32(loopOrdinal))
	loopIdx := int(int32(loopOrdinal - 1))
	stateNodes := make([]Value, 0, len(trace))
	actionEdges := make([]Value, 0, len(trace))
	for i, info := range trace {
		if info == nil {
			panic(NewNullPointerException())
		}
		stateNumber := int32(info.GetStateNumber())
		node := NewTupleValue([]Value{NewIntValue(stateNumber), stateInfoRecordValue(info)})
		if i > 0 {
			edgeAction := info.GetAction()
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
	ensureCounterExampleUniqueStrings()
	return NewCounterExample(nil, UnknownAction, 0, true)
}

func NewCounterExampleFromTrace(trace []*TLCStateInfo) *CounterExample {
	return NewCounterExample(trace, UnknownAction, 0, true)
}

func NewCounterExampleFromTraceConsole(trace []*TLCStateInfo, isConsole bool) *CounterExample {
	return NewCounterExample(trace, UnknownAction, 0, isConsole)
}

func NewCounterExampleFromTraceLoop(trace []*TLCStateInfo, loopOrdinal int) *CounterExample {
	return NewCounterExample(trace, UnknownAction, loopOrdinal, true)
}

func NewCounterExampleFromTraceActionLoop(trace []*TLCStateInfo, action *Action, loopOrdinal int) *CounterExample {
	return NewCounterExample(trace, action, loopOrdinal, true)
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
	ensureCounterExampleUniqueStrings()
	if c == nil || c.RecordValue == nil {
		panic(NewNullPointerException())
	}
	value, err := c.Select(NewStringValueFromUnique(counterExampleStates))
	if err != nil {
		panic(err)
	}
	if value == nil {
		panic(NewNullPointerException())
	}
	set, ok := value.(*SetEnumValue)
	if !ok {
		panic(valueStreamClassCast(value, "tlc2.value.impl.SetEnumValue"))
	}
	values := make([]Value, set.Elems.Len())
	for i := 0; i < set.Elems.Len(); i++ {
		elem := set.Elems.At(i)
		if elem == nil {
			panic(NewNullPointerException())
		}
		tv, ok := elem.(*TupleValue)
		if !ok {
			panic(valueStreamClassCast(elem, "tlc2.value.impl.TupleValue"))
		}
		ordinalValue := tv.GetElem(0)
		if ordinalValue == nil {
			panic(NewNullPointerException())
		}
		ordinal, ok := ordinalValue.(*IntValue)
		if !ok {
			panic(valueStreamClassCast(ordinalValue, "tlc2.value.impl.IntValue"))
		}
		idx := int(ordinal.ownedVal()) - 1
		if idx < 0 || idx >= len(values) {
			panic(NewArrayIndexOutOfBoundsException(idx, len(values)))
		}
		values[idx] = tv.GetElem(1)
	}
	return NewTupleValue(values)
}

func stateInfoRecordValue(info *TLCStateInfo) Value {
	ensureCounterExampleUniqueStrings()
	if info == nil {
		return EmptyRecord
	}
	return info.ToRecordValue()
}

func (a *Action) GetLocation(actionNames ...string) string {
	a.require()
	if len(actionNames) > 1 {
		panic(NewIllegalArgumentException("GetLocation accepts at most one action name"))
	}
	name := "Action"
	if len(actionNames) > 0 {
		name = actionNames[0]
	} else if a.IsNamed() {
		name = a.GetName()
	}
	parts := a.locationParameters()
	if len(parts) == 0 {
		return "<" + name + " " + a.GetDefinition().String() + ">"
	}
	return "<" + name + "(" + strings.Join(parts, ",") + ") " + a.GetDefinition().String() + ">"
}

// GetLocationNamed retains the older native spelling.
func (a *Action) GetLocationNamed(name string) string {
	return a.GetLocation(name)
}

func (a *Action) locationParameters() []string {
	a.require()
	if a.OpDef == nil {
		return nil
	}
	params := a.OpDef.GetParams()
	parts := make([]string, 0, len(params))
	for _, param := range params {
		con := a.ownedContext()
		if con == nil {
			panic(NewNullPointerException())
		}
		value := con.Lookup(param)
		if !isNil(value) {
			parts = append(parts, toContextString(value))
		}
	}
	return parts
}

func (a *Action) ToRecordValue() *RecordValue {
	return NewRecordValueFromAction(a)
}

// NewRecordValueFromAction represents RecordValue(Action).
func NewRecordValueFromAction(action *Action) *RecordValue {
	return newRecordValueFromAction(action, nil, nil)
}

// NewRecordValueFromActionField represents RecordValue(Action, UniqueString, Value).
func NewRecordValueFromActionField(action *Action, name *UniqueString, value Value) *RecordValue {
	return newRecordValueFromAction(action, name, value)
}

func newRecordValueFromAction(a *Action, extraName *UniqueString, extraValue Value) *RecordValue {
	ensureCounterExampleUniqueStrings()
	if a == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.Action.getParameters()" because "action" is null`))
	}
	names := []*UniqueString{actionRecordName, actionRecordLocation}
	values := []Value{NewStringValueFromUnique(a.GetNameUnique()), NewRecordValueFromLocation(a.GetDefinition())}
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

// NewRecordValueFromLocation represents RecordValue(Location).
func NewRecordValueFromLocation(location SourceLocation) *RecordValue {
	ensureCounterExampleUniqueStrings()
	source := location.Source
	if location == NullSourceLocation {
		// Java's Location.nullLoc carries this source name even though its
		// human-readable form is "Unknown location".
		source = "--unknown--"
	}
	return NewRecordValue(
		[]*UniqueString{locationBeginLine, locationBeginColumn, locationEndLine, locationEndColumn, locationModule},
		[]Value{
			NewIntValue(int32(location.BeginLine)),
			NewIntValue(int32(location.BeginColumn)),
			NewIntValue(int32(location.EndLine)),
			NewIntValue(int32(location.EndColumn)),
			NewStringValue(source),
		},
		false,
	)
}

// NewRecordValueFromOpDef represents RecordValue(OpDefNode).
func NewRecordValueFromOpDef(op *OpDefNode) *RecordValue {
	ensureCounterExampleUniqueStrings()
	if op == nil {
		panic(NewNullPointerException(`Cannot invoke "tla2sany.semantic.OpDefNode.getName()" because "odn" is null`))
	}
	name := ""
	if op.Name != nil {
		name = op.Name.String()
	}
	return NewRecordValue(
		[]*UniqueString{actionRecordName, actionRecordLocation},
		[]Value{NewStringValue(name), NewRecordValueFromLocation(semanticNodeLocation(op))},
		false,
	)
}

// NewRecordValueFromOpDecl represents RecordValue(OpDeclNode).
func NewRecordValueFromOpDecl(decl *SymbolNode) *RecordValue {
	return NewRecordValueFromOpDeclField(decl, nil, nil)
}

// NewRecordValueFromOpDeclField represents RecordValue(OpDeclNode, UniqueString, Value).
func NewRecordValueFromOpDeclField(decl *SymbolNode, field *UniqueString, value Value) *RecordValue {
	ensureCounterExampleUniqueStrings()
	if decl == nil {
		panic(NewNullPointerException(`Cannot invoke "tla2sany.semantic.OpDeclNode.getName()" because "odn" is null`))
	}
	names := []*UniqueString{actionRecordName, actionRecordLocation}
	values := []Value{NewStringValueFromUnique(decl.Name), NewRecordValueFromLocation(decl.GetSourceLocation())}
	if field != nil {
		names = append(names, field)
		values = append(values, value)
	}
	return NewRecordValue(names, values, false)
}

func ensureCounterExampleUniqueStrings() {
	counterExampleClassInterning.ensure(initCounterExampleUniqueStrings)
}
