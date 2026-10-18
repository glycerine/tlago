package tlc

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"unicode"
)

const (
	recordArrow                         = " |-> "
	defaultFcnRcdLinearSearchThreshold  = 32
	fcnRcdLinearSearchThresholdProperty = "tlc2.value.impl.FcnRcdValue.threshold"
	typedModelValueSeparatorRune        = '_'
	typedModelValueUntypedCodeUnit      = rune(0)
)

type ModelValue struct {
	BaseValue
	Val   *UniqueString
	Index int
	Type  rune
	Data  any
}

func (v *ModelValue) Mutates() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return false
}

func (v *ModelValue) IsAtom() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}

var modelValues = struct {
	sync.Mutex
	count int
	table *InsMap[string, *ModelValue]
	mvs   []*ModelValue
}{
	table: NewInsMap[string, *ModelValue](),
}

func ModelValueInit() {
	modelValues.Lock()
	defer modelValues.Unlock()
	modelValues.count = 0
	modelValues.table = NewInsMap[string, *ModelValue]()
	modelValues.mvs = nil
}

func MakeModelValue(name string) *ModelValue {
	modelValues.Lock()
	defer modelValues.Unlock()
	if mv, ok := modelValues.table.Get2(name); ok {
		return mv
	}
	mv := newModelValueLocked(name)
	modelValues.table.Set(name, mv)
	return mv
}

func AddModelValue(name string) *ModelValue {
	modelValues.Lock()
	defer modelValues.Unlock()
	if mv, ok := modelValues.table.Get2(name); ok {
		return mv
	}
	mv := newModelValueLocked(name)
	modelValues.table.Set(name, mv)
	setModelValuesLocked()
	return mv
}

func SetModelValues() {
	modelValues.Lock()
	defer modelValues.Unlock()
	setModelValuesLocked()
}

func ModelValues() []*ModelValue {
	modelValues.Lock()
	defer modelValues.Unlock()
	out := make([]*ModelValue, len(modelValues.mvs))
	copy(out, modelValues.mvs)
	return out
}

func ModelValueAtIndex(index int) *ModelValue {
	modelValues.Lock()
	defer modelValues.Unlock()
	if len(modelValues.mvs) != modelValues.count {
		setModelValuesLocked()
	}
	if index < 0 || index >= len(modelValues.mvs) {
		return nil
	}
	return modelValues.mvs[index]
}

// Stream reads index the already-initialized Java ModelValue.mvs table. They
// must not create that table or translate null/bounds failures into I/O errors.
func modelValueFromStream(index int) *ModelValue {
	modelValues.Lock()
	defer modelValues.Unlock()
	if modelValues.mvs == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(modelValues.mvs) {
		panic(NewArrayIndexOutOfBoundsException(index, len(modelValues.mvs)))
	}
	return modelValues.mvs[index]
}

type MVPerm struct {
	elems []*ModelValue
	count int
}

func NewMVPerm() *MVPerm {
	modelValues.Lock()
	defer modelValues.Unlock()
	if modelValues.mvs == nil {
		panic(NewNullPointerException())
	}
	return &MVPerm{elems: make([]*ModelValue, len(modelValues.mvs))}
}

func (p *MVPerm) Get(value Value) Value {
	if p == nil {
		panic(NewNullPointerException())
	}
	mv, ok := value.(*ModelValue)
	if !ok && value != nil {
		panic(valueStreamClassCast(value, "tlc2.value.impl.ModelValue"))
	}
	if mv == nil {
		panic(NewNullPointerException())
	}
	res := p.elementAt(mv.Index)
	if res == nil {
		return nil
	}
	return res
}

func (p *MVPerm) Put(dval, rval *ModelValue) {
	if p == nil {
		panic(NewNullPointerException())
	}
	if !permutationModelValueEqual(dval, rval) && p.elementAt(dval.Index) == nil {
		p.elems[dval.Index] = rval
		p.count = int(int32(p.count) + 1)
	}
}

func (p *MVPerm) putIndex(index int, elem *ModelValue) {
	if p.elementAt(index) == nil && elem != nil {
		p.elems[index] = elem
		p.count = int(int32(p.count) + 1)
	}
}

func (p *MVPerm) Size() int {
	if p == nil {
		panic(NewNullPointerException())
	}
	return p.count
}

func (p *MVPerm) Compose(perm *MVPerm) *MVPerm {
	if p == nil {
		panic(NewNullPointerException())
	}
	res := NewMVPerm()
	for i, mv := range p.elems {
		if mv == nil {
			res.putIndex(i, perm.elementAt(i))
			continue
		}
		mv1 := perm.elementAt(mv.Index)
		if mv1 == nil {
			res.putIndex(i, mv)
		} else if !permutationModelValueEqual(currentPermutationModelValue(i), mv1) {
			res.putIndex(i, mv1)
		}
	}
	return res
}

func (p *MVPerm) Equal(other *MVPerm) bool {
	if p == nil {
		panic(NewNullPointerException())
	}
	if other == nil {
		return false
	}
	for i, mv := range p.elems {
		otherValue := other.elementAt(i)
		if mv == nil {
			if otherValue != nil {
				return false
			}
		} else if !permutationModelValueEqual(mv, otherValue) {
			return false
		}
	}
	return true
}

func (p *MVPerm) AllModelValues() []*ModelValue {
	if p == nil {
		panic(NewNullPointerException())
	}
	values := make([]*ModelValue, 0)
	for _, mv := range p.elems {
		if mv != nil {
			values = append(values, mv)
		}
	}
	return values
}

func (p *MVPerm) GetAllModelValues() []*ModelValue {
	return p.AllModelValues()
}

func (p *MVPerm) String() string {
	if p == nil {
		panic(NewNullPointerException())
	}
	var b strings.Builder
	b.WriteByte('[')
	wrote := false
	for i, mv := range p.elems {
		if mv == nil {
			continue
		}
		if wrote {
			b.WriteString(", ")
		}
		b.WriteString(currentPermutationModelValue(i).String())
		b.WriteString(" -> ")
		b.WriteString(mv.String())
		wrote = true
	}
	b.WriteByte(']')
	return b.String()
}

func (p *MVPerm) HashCode() int32 {
	if p == nil {
		panic(NewNullPointerException())
	}
	var hash int32
	for _, mv := range p.elems {
		if mv == nil {
			continue
		}
		if mv.Val == nil {
			panic(NewNullPointerException())
		}
		var valueHash int32
		for _, unit := range javaStringUTF16(mv.Val.String()) {
			valueHash = 31*valueHash + int32(unit)
		}
		hash = 31*hash + valueHash
	}
	return hash
}

func (p *MVPerm) elementAt(index int) *ModelValue {
	if p == nil || p.elems == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(p.elems) {
		panic(NewArrayIndexOutOfBoundsException(index, len(p.elems)))
	}
	return p.elems[index]
}

func currentPermutationModelValue(index int) *ModelValue {
	modelValues.Lock()
	defer modelValues.Unlock()
	if modelValues.mvs == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(modelValues.mvs) {
		panic(NewArrayIndexOutOfBoundsException(index, len(modelValues.mvs)))
	}
	return modelValues.mvs[index]
}

func permutationModelValueEqual(left, right *ModelValue) bool {
	if left == nil {
		panic(NewNullPointerException())
	}
	var value Value
	if right != nil {
		value = right
	}
	equal, err := left.Equal(value)
	if err != nil {
		panic(err)
	}
	return equal
}

func newModelValueLocked(name string) *ModelValue {
	typ := typedModelValueUntypedCodeUnit
	units := javaStringUTF16(name)
	if len(units) > 2 && units[1] == uint16(typedModelValueSeparatorRune) {
		typ = rune(units[0])
	}
	mv := &ModelValue{
		Val:   UniqueStringOf(name),
		Index: modelValues.count,
		Type:  typ,
	}
	modelValues.count++
	return mv
}

func setModelValuesLocked() {
	modelValues.mvs = make([]*ModelValue, modelValues.table.Len())
	for _, mv := range modelValues.table.All() {
		modelValues.mvs[mv.Index] = mv
	}
}

func (v *ModelValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ModelValueKind
}
func (v *ModelValue) KindString() string { return ValueKindString(v) }
func (v *ModelValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *ModelValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if o, ok := other.(*ModelValue); ok && o == nil {
		other = nil
	}
	if v.Type == typedModelValueUntypedCodeUnit {
		if o, ok := other.(*ModelValue); ok {
			if v.Val == nil || o.Val == nil {
				panic(NewNullPointerException())
			}
			return v.Val.Compare(o.Val), nil
		}
		return -1, nil
	}
	if o, ok := other.(*ModelValue); ok {
		if o.Type == v.Type || o.Type == typedModelValueUntypedCodeUnit {
			if v.Val == nil || o.Val == nil {
				panic(NewNullPointerException())
			}
			return v.Val.Compare(o.Val), nil
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the differently-typed model values %s and %s", ValuesPPR(v), ValuesPPR(o)))
	}
	if isNil(other) {
		panic(NewNullPointerException())
	}
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the typed model value %s and non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
}

func (v *ModelValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if o, ok := other.(*ModelValue); ok && o == nil {
		other = nil
	}
	if v.Type == typedModelValueUntypedCodeUnit {
		o, ok := other.(*ModelValue)
		if !ok || o == nil {
			return false, nil
		}
		if v.Val == nil || o.Val == nil {
			panic(NewNullPointerException())
		}
		return v.Val.Equal(o.Val), nil
	}
	if o, ok := other.(*ModelValue); ok && o != nil {
		if o.Type == v.Type || o.Type == typedModelValueUntypedCodeUnit {
			return o.Val == v.Val, nil
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of the differently-typed model values %s and %s", ValuesPPR(v), ValuesPPR(o)))
	}
	if isNil(other) {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of typed model value %s and non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
}

func (v *ModelValue) ModelValueCompareTo(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if v.Type != typedModelValueUntypedCodeUnit {
		if other == nil {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the typed model value %s and the non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return 1, nil
}

func (v *ModelValue) ModelValueEquals(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if v.Type != typedModelValueUntypedCodeUnit {
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of the typed model value %s and the non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return false, nil
}

func (v *ModelValue) ModelValueMember(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if v.Type != typedModelValueUntypedCodeUnit {
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the typed model value %s is an element of\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return false, nil
}

func (v *ModelValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if o, ok := elem.(*ModelValue); ok && o == nil {
		elem = nil
	}
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the model value %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *ModelValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the model value %s is a finite set.", ValuesPPR(v)))
}

func (v *ModelValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the model value %s.", ValuesPPR(v)))
}

func (v *ModelValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *ModelValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
}
func (v *ModelValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *ModelValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *ModelValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *ModelValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if perm == nil {
		panic(NewNullPointerException())
	}
	if res := perm.Get(v); res != nil {
		return res
	}
	return v
}

func (v *ModelValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	fp = FP64ExtendByte(fp, byte(ModelValueKind))
	if v.Val == nil {
		panic(NewNullPointerException())
	}
	return v.Val.FingerPrint(fp)
}

func (v *ModelValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the model value %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *ModelValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the model value %s.", ValuesPPR(v)))
	}
	return v, nil
}

func (v *ModelValue) HasData() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.Data != nil
}
func (v *ModelValue) GetData() any {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.Data
}

func (v *ModelValue) SetData(obj any) any {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.Data = obj
	return obj
}

func (v *ModelValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *ModelValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.Val == nil {
		sb.WriteString("null")
	} else {
		sb.WriteString(v.Val.String())
	}
	return sb
}

type RecordValue struct {
	BaseValue
	Names  []*UniqueString
	Values []Value
	IsNorm bool

	arraysCaptured bool
	namesOwner     []*UniqueString
	valuesOwner    []Value
}

var EmptyRecord = NewRecordValue([]*UniqueString{}, []Value{}, true)

func NewRecordValue(names []*UniqueString, values []Value, isNorm bool, cms ...CostModel) *RecordValue {
	return &RecordValue{BaseValue: newBaseValue(cms...), Names: names, Values: values, IsNorm: isNorm,
		arraysCaptured: true, namesOwner: names, valuesOwner: values}
}

func (v *RecordValue) ownedNames() []*UniqueString {
	if v.arraysCaptured {
		return v.namesOwner
	}
	return v.Names
}

func (v *RecordValue) ownedValues() []Value {
	if v.arraysCaptured {
		return v.valuesOwner
	}
	return v.Values
}

func (v *RecordValue) captureArrays(names []*UniqueString, values []Value) {
	v.Names, v.Values = names, values
	v.arraysCaptured, v.namesOwner, v.valuesOwner = true, names, values
}

func NewSingletonRecordValue(name *UniqueString, value Value, normalized ...bool) *RecordValue {
	if len(normalized) > 1 {
		panic(NewIllegalArgumentException("NewSingletonRecordValue accepts at most one normalization flag"))
	}
	isNorm := false
	if len(normalized) != 0 {
		isNorm = normalized[0]
	}
	return NewRecordValue([]*UniqueString{name}, []Value{value}, isNorm)
}

// NewExtendedRecordValue represents RecordValue(RecordValue, UniqueString, Value).
func NewExtendedRecordValue(existing *RecordValue, name *UniqueString, value Value) *RecordValue {
	if existing == nil {
		panic(NewNullPointerException(`Cannot read field "names" because "existing" is null`))
	}
	names := make([]*UniqueString, len(existing.ownedNames())+1)
	copy(names, existing.ownedNames())
	names[len(existing.ownedNames())] = name
	values := make([]Value, len(existing.ownedValues())+1)
	copy(values, existing.ownedValues())
	values[len(existing.ownedValues())] = value
	return NewRecordValue(names, values, false)
}

// NewRecordValueFromState ports RecordValue(TLCState[, Value]).
func NewRecordValueFromState(state *TLCStateMut, defaultValue ...Value) *RecordValue {
	if len(defaultValue) > 1 {
		panic(NewIllegalArgumentException("NewRecordValueFromState accepts at most one default value"))
	}
	if state == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCState.getVars()" because "state" is null`))
	}
	return newRecordValueFromStateLookup(state, defaultValue...)
}

// NewRecordValueFromStateInfo ports RecordValue(TLCStateInfo). The
// TLCStateInfo.ToRecordValue method retains the separate PrintTLCState path
// which returns the printable state's owned record by identity.
func NewRecordValueFromStateInfo(info *TLCStateInfo) *RecordValue {
	if info == nil {
		panic(NewNullPointerException(`Cannot read field "state" because "info" is null`))
	}
	return NewRecordValueFromState(info.ownedState())
}

// Collections retain polymorphic state lookup, as RecordValue(TLCState) does.
func newRecordValueFromStateLookup(state interface{ Lookup(*UniqueString) Value }, defaultValue ...Value) *RecordValue {
	if isNil(state) {
		panic(NewNullPointerException())
	}
	variables, declarations := stateVariables, stateVariableDeclarations
	count := len(variables)
	if declarations != nil {
		count = len(declarations)
	}
	names := make([]*UniqueString, count)
	values := make([]Value, count)
	for i := range names {
		names[i] = recordStateVariableName(variables, declarations, i)
		values[i] = state.Lookup(names[i])
	}
	if len(defaultValue) > 0 {
		for i, value := range values {
			if isNil(value) {
				values[i] = defaultValue[0]
			}
		}
	}
	return NewRecordValue(names, values, false)
}

// NewRecordValueFromStateAndAction ports RecordValue(TLCState, Action).
func NewRecordValueFromStateAndAction(state *TLCStateMut, action *Action) *RecordValue {
	if state == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCState.getVars()" because "state" is null`))
	}
	variables, declarations := stateVariables, stateVariableDeclarations
	count := stateVariableCount(variables, declarations)
	names := make([]*UniqueString, count+1)
	values := make([]Value, count+1)
	ensureTLCExtUniqueStrings()
	names[0] = tlcExtActionField
	if action == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.Action.getParameters()" because "action" is null`))
	}
	values[0] = action.ToRecordValue()
	for i := 0; i < count; i++ {
		names[i+1] = recordStateVariableName(variables, declarations, i)
		values[i+1] = state.Lookup(names[i+1])
	}
	return NewRecordValue(names, values, false)
}

// Read the retained declaration array at each source getName call. Native
// callers that initialize state variables by name have no declaration array.
func recordStateVariableName(variables []StateVariable, declarations []*SymbolNode, index int) *UniqueString {
	if declarations != nil {
		if declarations[index] == nil {
			panic(NewNullPointerException(`Cannot invoke "tla2sany.semantic.OpDeclNode.getName()" because "vars[i]" is null`))
		}
		return declarations[index].Name
	}
	return variables[index].Name
}

// NewRecordValueFromStates ports the debugger's paired state constructor,
// including the trailing space on unprimed keys and interleaved field order.
func NewRecordValueFromStates(state, successor *TLCStateMut, defaultValue Value) *RecordValue {
	if state == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCState.getVars()" because "state" is null`))
	}
	if successor == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCState.getVars()" because "successor" is null`))
	}
	variables, declarations := stateVariables, stateVariableDeclarations
	count := len(variables)
	if declarations != nil {
		count = len(declarations)
	}
	names := make([]*UniqueString, count*2)
	values := make([]Value, len(names))
	for i := 0; i < count; i++ {
		j := i * 2
		variable := recordStateVariableName(variables, declarations, i)
		name := "null"
		if variable != nil {
			name = variable.String()
		}
		names[j] = UniqueStringOf(name + " ")
		names[j+1] = UniqueStringOf(name + "'")
		values[j] = state.Lookup(variable)
		if isNil(values[j]) {
			values[j] = defaultValue
		}
		values[j+1] = successor.Lookup(variable)
		if isNil(values[j+1]) {
			values[j+1] = defaultValue
		}
	}
	return NewRecordValue(names, values, false)
}

func NewRecordValueFromInsMap(values *InsMap[*UniqueString, Value]) *RecordValue {
	if values == nil {
		panic(NewNullPointerException(`Cannot invoke "java.util.Map.entrySet()" because "m" is null`))
	}
	names := make([]*UniqueString, 0, values.Len())
	vals := make([]Value, 0, values.Len())
	for name, value := range values.All() {
		names = append(names, name)
		vals = append(vals, value)
	}
	return NewRecordValue(names, vals, false)
}

func (v *RecordValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return RecordValueKind
}
func (v *RecordValue) KindString() string { return ValueKindString(v) }
func (v *RecordValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *RecordValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	var rcd *RecordValue
	if !isNil(other) {
		rcd = asRecordValue(other)
	}
	if rcd == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueCompareTo(v)
		}
		selfText := ValuesPPR(v)
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare record:\n%s\nwith non-record\n%s", selfText, ValuesPPR(other)))
	}
	if err := v.normalizeRecord(); err != nil {
		return 0, err
	}
	if err := rcd.normalizeRecord(); err != nil {
		return 0, err
	}
	if v.ownedNames() == nil || rcd.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	length := len(v.ownedNames())
	if length != len(rcd.ownedNames()) {
		return len(v.ownedNames()) - len(rcd.ownedNames()), nil
	}
	for i := 0; i < length; i++ {
		if v.ownedNames()[i] == nil || rcd.ownedNames()[i] == nil {
			panic(NewNullPointerException())
		}
		if cmp := v.ownedNames()[i].Compare(rcd.ownedNames()[i]); cmp != 0 {
			return cmp, nil
		}
	}
	for i := 0; i < length; i++ {
		left := fcnParameterDomain(v.ownedValues(), i)
		right := fcnParameterDomain(rcd.ownedValues(), i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		cmp, err := left.Compare(right)
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *RecordValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	var rcd *RecordValue
	if !isNil(other) {
		rcd = asRecordValue(other)
	}
	if rcd == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueEquals(v)
		}
		selfText := ValuesPPR(v)
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of record:\n%s\nwith non-record\n%s", selfText, ValuesPPR(other)))
	}
	if err := v.normalizeRecord(); err != nil {
		return false, err
	}
	if err := rcd.normalizeRecord(); err != nil {
		return false, err
	}
	if v.ownedNames() == nil || rcd.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	length := len(v.ownedNames())
	if length != len(rcd.ownedNames()) {
		return false, nil
	}
	for i := 0; i < length; i++ {
		if v.ownedNames()[i] == nil || rcd.ownedNames()[i] == nil {
			panic(NewNullPointerException())
		}
		if !v.ownedNames()[i].Equal(rcd.ownedNames()[i]) {
			return false, nil
		}
	}
	for i := 0; i < length; i++ {
		left := fcnParameterDomain(v.ownedValues(), i)
		right := fcnParameterDomain(rcd.ownedValues(), i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		eq, err := left.Equal(right)
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *RecordValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(elem) {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if element:\n%s\nis in the record:\n%s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *RecordValue) IsFinite() (bool, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true, nil
}

func (v *RecordValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		if v.ownedNames() == nil {
			panic(NewNullPointerException())
		}
		rlen := len(v.ownedNames())
		newValues := make([]Value, rlen)
		arcVal := ex.Current()
		var arc *StringValue
		var ok bool
		if !isNil(arcVal) {
			arc, ok = asStringValue(arcVal)
		}
		if ok {
			for i := 0; i < rlen; i++ {
				if v.ownedNames()[i] == nil || arc.ownedVal() == nil {
					panic(NewNullPointerException())
				}
				if v.ownedNames()[i].Equal(arc.ownedVal()) {
					ex.Index++
					value := fcnParameterDomain(v.ownedValues(), i)
					if isNil(value) {
						panic(NewNullPointerException())
					}
					taken, err := value.TakeExcept(ex)
					if err != nil {
						return nil, err
					}
					newValues[i] = taken
				} else {
					newValues[i] = fcnParameterDomain(v.ownedValues(), i)
				}
			}
			newNames := v.ownedNames()
			if !v.IsNorm {
				newNames = make([]*UniqueString, len(v.ownedNames()))
				copy(newNames, v.ownedNames())
			}
			return NewRecordValue(newNames, newValues, v.IsNorm), nil
		}
		if isNil(arcVal) {
			panic(NewNullPointerException())
		}
		PrintWarning(ECTLCWrongRecordFieldName, ValuesPPR(arcVal))
	}
	return ex.Value, nil
}

func (v *RecordValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	var cur Value = v
	for _, ex := range exs {
		if isNil(cur) {
			panic(NewNullPointerException())
		}
		next, err := cur.TakeExcept(ex)
		if err != nil {
			return nil, err
		}
		cur = next
	}
	return cur, nil
}

func (v *RecordValue) ToRcd() *RecordValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *RecordValue) ToFcnRcd() *FcnRcdValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	domain := make([]Value, len(v.ownedNames()))
	for i, name := range v.ownedNames() {
		domain[i] = NewStringValueFromUnique(name, v.CM)
	}
	v.CM.incValueSecondary(int64(len(domain)))
	return NewFcnRcdValue(domain, v.ownedValues(), v.IsNorm, v.CM)
}

func (v *RecordValue) ToTuple() *TupleValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	size, err := v.Size()
	if err != nil {
		panic(err)
	}
	if size == 0 {
		return EmptyTuple
	}
	return nil
}

func (v *RecordValue) ToState() *TLCStateMut {
	if v == nil || EmptyState == nil {
		panic(NewNullPointerException())
	}
	state := EmptyState.CreateEmpty()
	variables, declarations := stateVariables, stateVariableDeclarations
	count := stateVariableCount(variables, declarations)
	for i := 0; i < count; i++ {
		name := stateVariableNameAt(variables, declarations, i)
		if v.ownedNames() == nil {
			panic(NewNullPointerException())
		}
		for j, field := range v.ownedNames() {
			if name == nil || field == nil {
				panic(NewNullPointerException())
			}
			if name.Equal(field) {
				state.Bind(name, v.recordValueAt(j))
			}
		}
	}
	return NewPrintTLCState(v, state)
}

// NewPrintTLCState exposes RecordValue.PrintTLCState's public constructor. The
// folded Go state retains both source owners and delegates its state behavior to
// state; neither nullable constructor argument is dereferenced during creation.
func NewPrintTLCState(record *RecordValue, state *TLCStateMut) *TLCStateMut {
	printStateFormatName()
	wrapper := newEmptyStateOfKind(false)
	if state != nil {
		// Keep a value-slice alias for native evaluator paths that consume the
		// concrete folded state directly.
		wrapper.values = state.values
	}
	wrapper.printRecord, wrapper.printState = record, state
	return wrapper
}

var printTLCStateStatics struct {
	sync.Mutex
	initialized bool
	formatName  *UniqueString
}

func printStateFormatName() *UniqueString {
	printTLCStateStatics.Lock()
	defer printTLCStateStatics.Unlock()
	if !printTLCStateStatics.initialized {
		printTLCStateStatics.formatName = UniqueStringOf("_format")
		printTLCStateStatics.initialized = true
	}
	return printTLCStateStatics.formatName
}

// InitializePrintTLCStateStatics starts an isolated runtime, like the original
// test runner's fresh classloader. It requires an idle runtime; an intern-table
// reset alone must retain the already captured format key.
func InitializePrintTLCStateStatics() {
	printTLCStateStatics.Lock()
	defer printTLCStateStatics.Unlock()
	printTLCStateStatics.initialized = false
	printTLCStateStatics.formatName = nil
}

func (v *RecordValue) StateString() string {
	formatName := printStateFormatName()
	if v == nil || v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	count := len(v.ownedNames())
	format := ""
	formatIndex := -1
	for i, name := range v.ownedNames() {
		if name == formatName {
			formatIndex = i
			break
		}
	}
	if formatIndex >= 0 {
		value := v.recordValueAt(formatIndex)
		if isNil(value) {
			panic(NewNullPointerException())
		}
		var sv *StringValue
		switch value := value.(type) {
		case *StringValue:
			sv = value
		case *DebuggerValue:
			sv = value.StringValue
		default:
			panic(valueStreamClassCast(value, "tlc2.value.impl.StringValue"))
		}
		if sv.ownedVal() == nil {
			panic(NewNullPointerException())
		}
		format = sv.ownedVal().String()
	} else if count == 1 {
		format = "%s = %s\n"
	} else {
		format = "/\\ %s = %s\n"
	}
	var out strings.Builder
	for i := 0; i < count; i++ {
		if i == formatIndex {
			continue
		}
		if v.ownedNames() == nil {
			panic(NewNullPointerException())
		}
		if i >= len(v.ownedNames()) {
			panic(NewArrayIndexOutOfBoundsException(i, len(v.ownedNames())))
		}
		name := v.ownedNames()[i]
		if name == nil {
			panic(NewNullPointerException())
		}
		key := name.String()
		value := ValuesPPR(v.recordValueAt(i))
		text, err := JavaFormatStrings(format, key, value)
		if err != nil {
			panic(err)
		}
		out.WriteString(text)
	}
	return javaStringFromUTF16(javaStringUTF16(out.String()))
}

func (v *RecordValue) recordValueAt(index int) Value {
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(v.ownedValues()) {
		panic(NewArrayIndexOutOfBoundsException(index, len(v.ownedValues())))
	}
	return v.ownedValues()[index]
}

func (v *RecordValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if v.ownedNames() == nil {
		return 0, NewNullPointerException()
	}
	return len(v.ownedNames()), nil
}

func (v *RecordValue) Apply(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(arg) {
		return nil, NewNullPointerException()
	}
	if sv, ok := asStringValue(arg); ok {
		if v.ownedNames() == nil {
			panic(NewNullPointerException())
		}
		for i, name := range v.ownedNames() {
			if sv.ownedVal() == nil || name == nil {
				panic(NewNullPointerException())
			}
			if sv.ownedVal().Equal(name) {
				return fcnParameterDomain(v.ownedValues(), i), nil
			}
		}
		name := "null"
		if sv.ownedVal() != nil {
			name = sv.ownedVal().String()
		}
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to access nonexistent field '%s' of record\n%s", name, ValuesPPR(v)))
	}
	return nil, v.runtimeFailure(fmt.Sprintf("Attempted to access record by a non-string argument: %s", ValuesPPR(arg)))
}

func (v *RecordValue) ApplyWithControl(arg Value, control int) (Value, error) {
	return v.Apply(arg)
}

// ApplyArgs retains RecordValue's array overload and its additional source
// catch boundary around single-argument application.
func (v *RecordValue) ApplyArgs(args []Value, control int) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if args == nil {
		panic(NewNullPointerException())
	}
	if len(args) != 1 {
		return nil, v.runtimeFailure("Attempted to apply record to more than one arguments.")
	}
	return v.Apply(args[0])
}

func (v *RecordValue) Select(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(arg) {
		// Java formats arg.toString() before Assert.fail for a non-string.
		panic(NewNullPointerException())
	}
	sv, ok := asStringValue(arg)
	if !ok {
		message := "Attempted to access record by a non-string argument: " + ValuesPPR(arg)
		if source := v.GetSource(); source != nil {
			return nil, NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
		}
		failure := newTLCError(ECGeneral, "%s", message)
		failure.Runtime = true
		return nil, failure
	}
	if sv == nil {
		panic(NewNullPointerException())
	}
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	for i, name := range v.ownedNames() {
		if sv.ownedVal() == nil || name == nil {
			panic(NewNullPointerException())
		}
		if sv.ownedVal().Equal(name) {
			return fcnParameterDomain(v.ownedValues(), i), nil
		}
	}
	return nil, nil
}

func (v *RecordValue) DomainValue() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, len(v.ownedNames()))
	for i, name := range v.ownedNames() {
		values[i] = NewStringValueFromUnique(name)
	}
	return NewSetEnumValue(values, v.IsNormalized())
}

func (v *RecordValue) GetDomain() (Value, error) {
	return v.DomainValue(), nil
}

func (v *RecordValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	return v
}

func (v *RecordValue) normalizeRecord() (err error) {
	defer catchValueFailure(v, &err)
	if v.IsNorm {
		return nil
	}
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	for i := 1; i < len(v.ownedNames()); i++ {
		if v.ownedNames()[0] == nil || v.ownedNames()[i] == nil {
			panic(NewNullPointerException())
		}
		cmp := v.ownedNames()[0].Compare(v.ownedNames()[i])
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("Field name %s occurs multiple times in record.", v.ownedNames()[i]))
		}
		if cmp > 0 {
			v.ownedNames()[0], v.ownedNames()[i] = v.ownedNames()[i], v.ownedNames()[0]
			first := fcnParameterDomain(v.ownedValues(), 0)
			other := fcnParameterDomain(v.ownedValues(), i)
			v.ownedValues()[0], v.ownedValues()[i] = other, first
		}
	}
	for i := 2; i < len(v.ownedNames()); i++ {
		j := i
		st := v.ownedNames()[i]
		val := fcnParameterDomain(v.ownedValues(), i)
		cmp := -1
		for j > 0 {
			if st == nil || v.ownedNames()[j-1] == nil {
				panic(NewNullPointerException())
			}
			cmp = st.Compare(v.ownedNames()[j-1])
			if cmp >= 0 {
				break
			}
			v.ownedNames()[j] = v.ownedNames()[j-1]
			v.ownedValues()[j] = fcnParameterDomain(v.ownedValues(), j-1)
			j--
		}
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("Field name %s occurs multiple times in record.", v.ownedNames()[i]))
		}
		v.ownedNames()[j] = st
		v.ownedValues()[j] = val
	}
	v.IsNorm = true
	return nil
}

func (v *RecordValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.ownedValues() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		value.DeepNormalize()
	}
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
}

func (v *RecordValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.ownedValues() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *RecordValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.IsNorm
}

func (v *RecordValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, len(v.ownedValues()))
	for i, value := range v.ownedValues() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		values[i] = value.DeepCopy()
	}
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	names := make([]*UniqueString, len(v.ownedNames()))
	copy(names, v.ownedNames())
	return NewRecordValue(names, values, v.IsNorm)
}

func (v *RecordValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.ownedNames())))
	for i, name := range v.ownedNames() {
		if name == nil {
			panic(NewNullPointerException())
		}
		fp = FP64ExtendByte(fp, byte(StringValueKind))
		fp = FP64ExtendInt(fp, int32(name.Length()))
		fp = FP64ExtendString(fp, name.String())
		value := fcnParameterDomain(v.ownedValues(), i)
		if isNil(value) {
			panic(NewNullPointerException())
		}
		fp = value.FingerPrint(fp)
	}
	return fp
}

func (v *RecordValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, len(v.ownedNames()))
	changed := false
	for i := range values {
		value := fcnParameterDomain(v.ownedValues(), i)
		if isNil(value) {
			panic(NewNullPointerException())
		}
		values[i] = value.Permute(perm)
		changed = changed || values[i] != value
	}
	if changed {
		return NewRecordValue(v.ownedNames(), values, true)
	}
	return v
}

func (v *RecordValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *RecordValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	if v.ownedNames() == nil {
		panic(NewNullPointerException())
	}
	sb.WriteString("[")
	for i, name := range v.ownedNames() {
		if i > 0 {
			sb.WriteString(", ")
		}
		text := "null"
		if name != nil {
			text = name.String()
		}
		sb.WriteString(text + recordArrow)
		value := fcnParameterDomain(v.ownedValues(), i)
		if isNil(value) {
			panic(NewNullPointerException())
		}
		sb = appendValueString(value, sb, offset, swallow)
	}
	sb.WriteString("]")
	return sb
}

type FcnRcdValue struct {
	BaseValue
	Domain []Value
	Intv   *IntervalValue
	Values []Value
	IsNorm bool

	carriersCaptured bool
	domainOwner      []Value
	intvOwner        *IntervalValue
	valuesOwner      []Value
}

var EmptyFcn = newFcnRcdValue(BaseValue{}, []Value{}, nil, []Value{}, true)

var fcnRcdStatics struct {
	once      sync.Once
	threshold int
}

// Go loads packages before TLC configures its runtime. Defer source class
// initialization until construction or use, including the empty singleton.
func emptyFcnValue() *FcnRcdValue {
	fcnRcdLinearSearchThreshold()
	return EmptyFcn
}

// InitializeFcnRcdValueStatics represents a fresh Java classloader. Ordinary
// checker runs retain the captured property and the empty function singleton.
// Like the other classloader resets, this requires an idle runtime.
func InitializeFcnRcdValueStatics() {
	fcnRcdStatics.once = sync.Once{}
	fcnRcdStatics.threshold = 0
	EmptyFcn = newFcnRcdValue(BaseValue{}, []Value{}, nil, []Value{}, true)
}

func NewFcnRcdValue(domain []Value, values []Value, isNorm bool, cms ...CostModel) *FcnRcdValue {
	fcnRcdLinearSearchThreshold()
	return newFcnRcdValue(newBaseValue(cms...), domain, nil, values, isNorm)
}

func NewFcnRcdValueVec(domain *ValueVec, values []Value, isNorm bool, cms ...CostModel) *FcnRcdValue {
	return NewFcnRcdValue(domain.ToArray(), values, isNorm, cms...)
}

func NewFcnRcdIntervalValue(intv *IntervalValue, values []Value, cms ...CostModel) *FcnRcdValue {
	fcnRcdLinearSearchThreshold()
	return newFcnRcdValue(newBaseValue(cms...), nil, intv, values, true)
}

func newFcnRcdValue(base BaseValue, domain []Value, intv *IntervalValue, values []Value, isNorm bool) *FcnRcdValue {
	return &FcnRcdValue{
		BaseValue: base, Domain: domain, Intv: intv, Values: values, IsNorm: isNorm,
		carriersCaptured: true, domainOwner: domain, intvOwner: intv, valuesOwner: values,
	}
}

func (v *FcnRcdValue) ownedDomain() []Value {
	if v.carriersCaptured {
		return v.domainOwner
	}
	return v.Domain
}

func (v *FcnRcdValue) ownedIntv() *IntervalValue {
	if v.carriersCaptured {
		return v.intvOwner
	}
	return v.Intv
}

func (v *FcnRcdValue) ownedValues() []Value {
	if v.carriersCaptured {
		return v.valuesOwner
	}
	return v.Values
}

func (v *FcnRcdValue) captureCarriers(domain []Value, intv *IntervalValue, values []Value) {
	v.Domain, v.Intv, v.Values = domain, intv, values
	v.carriersCaptured, v.domainOwner, v.intvOwner, v.valuesOwner = true, domain, intv, values
}

// NewFcnRcdValueFromValues ports FcnRcdValue(List<Value>).
func NewFcnRcdValueFromValues(values []Value) *FcnRcdValue {
	if values == nil {
		panic(NewNullPointerException())
	}
	copied := make([]Value, len(values))
	copy(copied, values)
	return NewFcnRcdIntervalValue(NewIntervalValue(1, int32(len(copied))), copied)
}

// FcnRcdValueMap is the Map<Value, Value> surface consumed by the matching
// FcnRcdValue constructors. Iteration must yield paired keys and values in the
// map's own entry order.
type FcnRcdValueMap interface {
	Len() int
	All() func(func(Value, Value) bool)
}

// NewFcnRcdValueFromMap ports FcnRcdValue(Map<Value, Value>[, boolean]).
func NewFcnRcdValueFromMap(pairs FcnRcdValueMap, normalized ...bool) *FcnRcdValue {
	if len(normalized) > 1 {
		panic(NewIllegalArgumentException("NewFcnRcdValueFromMap accepts at most one normalization flag"))
	}
	if isNil(pairs) {
		panic(NewNullPointerException())
	}
	domain := make([]Value, 0, pairs.Len())
	values := make([]Value, 0, pairs.Len())
	for key, value := range pairs.All() {
		domain = append(domain, key)
		values = append(values, value)
	}
	isNorm := false
	if len(normalized) != 0 {
		isNorm = normalized[0]
	}
	return NewFcnRcdValue(domain, values, isNorm)
}

func (v *FcnRcdValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	return FcnRcdValueKind
}
func (v *FcnRcdValue) KindString() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	return ValueKindString(v)
}
func (v *FcnRcdValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *FcnRcdValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	var fcn *FcnRcdValue
	if !isNil(other) {
		fcn = asFcnRcdValue(other)
	}
	if fcn == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueCompareTo(v)
		}
		selfText := ValuesPPR(v)
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the function %s with the value:\n%s", selfText, ValuesPPR(other)))
	}
	if err := v.normalizeFcn(); err != nil {
		return 0, err
	}
	if err := fcn.normalizeFcn(); err != nil {
		return 0, err
	}
	if v.ownedValues() == nil || fcn.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	if len(v.ownedValues()) != len(fcn.ownedValues()) {
		return len(v.ownedValues()) - len(fcn.ownedValues()), nil
	}
	if v.ownedIntv() != nil {
		return v.compareToInterval(fcn)
	}
	return v.compareOtherInterval(fcn)
}

func (v *FcnRcdValue) compareOtherInterval(fcn *FcnRcdValue) (int, error) {
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	if fcn.ownedIntv() != nil {
		for i := 0; i < len(v.ownedDomain()); i++ {
			dElem := fcnParameterDomain(v.ownedDomain(), i)
			iv, ok := dElem.(*IntValue)
			if !ok || iv == nil {
				if isNil(dElem) {
					panic(NewNullPointerException())
				}
				return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare integer with non-integer\n%s.", ValuesPPR(dElem)))
			}
			intervalElement := int64(fcn.ownedIntv().Low) + int64(i)
			domainElement := int64(iv.ownedVal())
			if domainElement < intervalElement {
				return -1, nil
			}
			if domainElement > intervalElement {
				return 1, nil
			}
		}
		for i := 0; i < len(v.ownedDomain()); i++ {
			left := fcnParameterDomain(v.ownedValues(), i)
			right := fcnParameterDomain(fcn.ownedValues(), i)
			if isNil(left) {
				panic(NewNullPointerException())
			}
			cmp, err := left.Compare(right)
			if err != nil || cmp != 0 {
				return cmp, err
			}
		}
		return 0, nil
	}
	for i := 0; i < len(v.ownedDomain()); i++ {
		left := fcnParameterDomain(v.ownedDomain(), i)
		right := fcnParameterDomain(fcn.ownedDomain(), i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		cmp, err := left.Compare(right)
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	for i := 0; i < len(v.ownedDomain()); i++ {
		left := fcnParameterDomain(v.ownedValues(), i)
		right := fcnParameterDomain(fcn.ownedValues(), i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		cmp, err := left.Compare(right)
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *FcnRcdValue) compareToInterval(fcn *FcnRcdValue) (int, error) {
	if len(v.ownedValues()) == 0 {
		return 0, nil
	}
	if fcn.ownedIntv() != nil {
		if v.ownedIntv().Low < fcn.ownedIntv().Low {
			return -1, nil
		}
		if v.ownedIntv().Low > fcn.ownedIntv().Low {
			return 1, nil
		}
		for i := 0; i < len(v.ownedValues()); i++ {
			left := fcnParameterDomain(v.ownedValues(), i)
			right := fcnParameterDomain(fcn.ownedValues(), i)
			if isNil(left) {
				panic(NewNullPointerException())
			}
			cmp, err := left.Compare(right)
			if err != nil || cmp != 0 {
				return cmp, err
			}
		}
		return 0, nil
	}
	if fcn.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < len(fcn.ownedDomain()); i++ {
		dElem := fcnParameterDomain(fcn.ownedDomain(), i)
		iv, ok := dElem.(*IntValue)
		if !ok || iv == nil {
			if isNil(dElem) {
				panic(NewNullPointerException())
			}
			return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare integer with non-integer:\n%s.", ValuesPPR(dElem)))
		}
		intervalElement := int64(v.ownedIntv().Low) + int64(i)
		domainElement := int64(iv.ownedVal())
		if intervalElement < domainElement {
			return -1, nil
		}
		if intervalElement > domainElement {
			return 1, nil
		}
	}
	for i := 0; i < len(fcn.ownedDomain()); i++ {
		left := fcnParameterDomain(v.ownedValues(), i)
		right := fcnParameterDomain(fcn.ownedValues(), i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		cmp, err := left.Compare(right)
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *FcnRcdValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	var fcn *FcnRcdValue
	if !isNil(other) {
		fcn = asFcnRcdValue(other)
	}
	if fcn == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueEquals(v)
		}
		selfText := ValuesPPR(v)
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of the function %s with the value:\n%s", selfText, ValuesPPR(other)))
	}
	if err := v.normalizeFcn(); err != nil {
		return false, err
	}
	if err := fcn.normalizeFcn(); err != nil {
		return false, err
	}
	if v.ownedIntv() != nil {
		if fcn.ownedIntv() != nil {
			eq, err := v.ownedIntv().Equal(fcn.ownedIntv())
			if err != nil || !eq {
				return eq, err
			}
			if v.ownedValues() == nil {
				panic(NewNullPointerException())
			}
			for i := range v.ownedValues() {
				left := fcnParameterDomain(v.ownedValues(), i)
				right := fcnParameterDomain(fcn.ownedValues(), i)
				if isNil(left) {
					panic(NewNullPointerException())
				}
				eq, err := left.Equal(right)
				if err != nil || !eq {
					return eq, err
				}
			}
			return true, nil
		}
		if fcn.ownedDomain() == nil {
			panic(NewNullPointerException())
		}
		if len(fcn.ownedDomain()) != mustIntervalSize(v.ownedIntv()) {
			return false, nil
		}
		for i, dElem := range fcn.ownedDomain() {
			iv, ok := dElem.(*IntValue)
			if !ok || iv == nil {
				if isNil(dElem) {
					panic(NewNullPointerException())
				}
				return false, v.runtimeFailure(fmt.Sprintf("Attempted to compare an integer with non-integer:\n%s.", ValuesPPR(dElem)))
			}
			if int64(iv.ownedVal()) != int64(v.ownedIntv().Low)+int64(i) {
				return false, nil
			}
		}
		if fcn.ownedValues() == nil {
			panic(NewNullPointerException())
		}
		for i := range fcn.ownedValues() {
			left := fcnParameterDomain(v.ownedValues(), i)
			right := fcnParameterDomain(fcn.ownedValues(), i)
			if isNil(left) {
				panic(NewNullPointerException())
			}
			eq, err := left.Equal(right)
			if err != nil || !eq {
				return eq, err
			}
		}
		return true, nil
	}
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	if fcn.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	if len(v.ownedValues()) != len(fcn.ownedValues()) {
		return false, nil
	}
	if fcn.ownedIntv() != nil {
		if v.ownedDomain() == nil {
			panic(NewNullPointerException())
		}
		for i, dElem := range v.ownedDomain() {
			iv, ok := dElem.(*IntValue)
			if !ok || iv == nil {
				if isNil(dElem) {
					panic(NewNullPointerException())
				}
				return false, v.runtimeFailure(fmt.Sprintf("Attempted to compare an integer with non-integer:\n%s.", ValuesPPR(dElem)))
			}
			if int64(iv.ownedVal()) != int64(fcn.ownedIntv().Low)+int64(i) {
				return false, nil
			}
		}
		if v.ownedValues() == nil {
			panic(NewNullPointerException())
		}
		for i := range v.ownedValues() {
			left := fcnParameterDomain(v.ownedValues(), i)
			right := fcnParameterDomain(fcn.ownedValues(), i)
			if isNil(left) {
				panic(NewNullPointerException())
			}
			eq, err := left.Equal(right)
			if err != nil || !eq {
				return eq, err
			}
		}
		return true, nil
	}
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	for i := range v.ownedDomain() {
		left := fcnParameterDomain(v.ownedDomain(), i)
		right := fcnParameterDomain(fcn.ownedDomain(), i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		eq, err := left.Equal(right)
		if err != nil || !eq {
			return eq, err
		}
	}
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	for i := range v.ownedValues() {
		left := fcnParameterDomain(v.ownedValues(), i)
		right := fcnParameterDomain(fcn.ownedValues(), i)
		if isNil(left) {
			panic(NewNullPointerException())
		}
		eq, err := left.Equal(right)
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *FcnRcdValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	if isNil(elem) {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the function %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *FcnRcdValue) IsFinite() (bool, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	return true, nil
}

func (v *FcnRcdValue) Apply(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	result, err := v.Select(arg)
	if err != nil {
		return nil, err
	}
	if isNil(result) {
		selfText := ValuesPPR(v)
		if isNil(arg) {
			panic(NewNullPointerException())
		}
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply function:\n%s\nto argument %s, which is not in the domain of the function.", selfText, ValuesPPR(arg)))
	}
	return result, nil
}

func (v *FcnRcdValue) ApplyWithControl(arg Value, control int) (Value, error) {
	return v.Apply(arg)
}

func (v *FcnRcdValue) ApplyArgs(args []Value, control int) (Value, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if args == nil {
		panic(NewNullPointerException())
	}
	if len(args) == 1 {
		return v.ApplyWithControl(args[0], control)
	}
	return v.ApplyWithControl(NewTupleValue(args), control)
}

func (v *FcnRcdValue) Select(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	if v.ownedIntv() != nil {
		iv, ok := arg.(*IntValue)
		if !ok || iv == nil {
			if isNil(arg) {
				panic(NewNullPointerException())
			}
			return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply function with integer domain to the non-integer argument %s", ValuesPPR(arg)))
		}
		if iv.ownedVal() >= v.ownedIntv().Low && iv.ownedVal() <= v.ownedIntv().High {
			offset := int64(iv.ownedVal()) - int64(v.ownedIntv().Low)
			if v.ownedValues() == nil {
				panic(NewNullPointerException())
			}
			if offset < int64(len(v.ownedValues())) {
				return v.ownedValues()[int(offset)], nil
			}
		}
		return nil, nil
	}
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	return v.SelectBinarySearch(arg)
}

// SelectLinearSearch ports FcnRcdValue.selectLinearSearch.
func (v *FcnRcdValue) SelectLinearSearch(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	for i, value := range v.ownedDomain() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		eq, err := value.Equal(arg)
		if err != nil {
			return nil, err
		}
		if eq {
			return fcnParameterDomain(v.ownedValues(), i), nil
		}
	}
	return nil, nil
}

// SelectBinarySearch ports FcnRcdValue.selectBinarySearch.
func (v *FcnRcdValue) SelectBinarySearch(arg Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	if v.IsNorm && len(v.ownedDomain()) >= fcnRcdLinearSearchThreshold() {
		// Arrays.binarySearch uses an inclusive upper bound and returns at the
		// first equal comparison. The visited values also determine which typed
		// model-value comparison reports an error.
		low, high := 0, len(v.ownedDomain())-1
		for low <= high {
			mid := (low + high) >> 1
			if isNil(v.ownedDomain()[mid]) {
				panic(NewNullPointerException())
			}
			cmp, err := v.ownedDomain()[mid].Compare(arg)
			if err != nil {
				return nil, err
			}
			if cmp < 0 {
				low = mid + 1
			} else if cmp > 0 {
				high = mid - 1
			} else {
				eq, err := v.ownedDomain()[mid].Equal(arg)
				if err != nil || !eq {
					return nil, err
				}
				return fcnParameterDomain(v.ownedValues(), mid), nil
			}
		}
		return nil, nil
	}
	return v.SelectLinearSearch(arg)
}

func fcnRcdLinearSearchThreshold() int {
	// Initializing FcnRcdValue initializes its Value superclass first in Java.
	_ = valueCoverageEnabled()
	fcnRcdStatics.once.Do(func() {
		fcnRcdStatics.threshold = defaultFcnRcdLinearSearchThreshold
		if value, ok := tlcLookupSystemProperty(fcnRcdLinearSearchThresholdProperty); ok {
			if parsed, ok := javaDecodeIntProperty(value); ok {
				fcnRcdStatics.threshold = parsed
			}
		}
		if fcnRcdStatics.threshold != defaultFcnRcdLinearSearchThreshold {
			ToolIOPrintln(fmt.Sprintf("FcnRcdValue#threshold is: %d", fcnRcdStatics.threshold))
		}
	})
	return fcnRcdStatics.threshold
}

func (v *FcnRcdValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index >= len(ex.Path) {
		return ex.Value, nil
	}
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	newValues := make([]Value, len(v.ownedValues()))
	copy(newValues, v.ownedValues())
	arg := ex.Current()
	if v.ownedIntv() != nil {
		if iv, ok := arg.(*IntValue); ok && iv != nil {
			if iv.ownedVal() >= v.ownedIntv().Low && iv.ownedVal() <= v.ownedIntv().High {
				offset := int64(iv.ownedVal()) - int64(v.ownedIntv().Low)
				if offset >= int64(len(newValues)) {
					return v, nil
				}
				vidx := int(offset)
				next := ex
				next.Index++
				if isNil(v.ownedValues()[vidx]) {
					panic(NewNullPointerException())
				}
				taken, err := v.ownedValues()[vidx].TakeExcept(next)
				if err != nil {
					return nil, err
				}
				newValues[vidx] = taken
			}
			return newFcnRcdValue(BaseValue{}, nil, v.ownedIntv(), newValues, true), nil
		}
		return v, nil
	}
	for i := range v.ownedValues() {
		domain := fcnParameterDomain(v.ownedDomain(), i)
		if isNil(arg) {
			panic(NewNullPointerException())
		}
		eq, err := arg.Equal(domain)
		if err != nil {
			return nil, err
		}
		if eq {
			next := ex
			next.Index++
			if isNil(newValues[i]) {
				panic(NewNullPointerException())
			}
			taken, err := newValues[i].TakeExcept(next)
			if err != nil {
				return nil, err
			}
			newValues[i] = taken
			newDomain := v.ownedDomain()
			if !v.IsNorm {
				// Java copies the captured value count even when the domain length differs.
				newDomain = make([]Value, len(newValues))
				for j := range newDomain {
					newDomain[j] = fcnParameterDomain(v.ownedDomain(), j)
				}
			}
			return newFcnRcdValue(BaseValue{}, newDomain, nil, newValues, v.IsNorm), nil
		}
	}
	return v, nil
}

func (v *FcnRcdValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	var cur Value = v
	for _, ex := range exs {
		if isNil(cur) {
			panic(NewNullPointerException())
		}
		next, err := cur.TakeExcept(ex)
		if err != nil {
			return nil, err
		}
		cur = next
	}
	return cur, nil
}

func (v *FcnRcdValue) DomainValue() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.ownedIntv() != nil {
		return v.ownedIntv()
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	return NewSetEnumValue(v.ownedDomain(), true)
}

func (v *FcnRcdValue) GetDomain() (Value, error) {
	return v.DomainValue(), nil
}

func (v *FcnRcdValue) GetDomainAsValues() []Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	if v.ownedIntv() != nil {
		return v.ownedIntv().AsValues()
	}
	return v.ownedDomain()
}

func (v *FcnRcdValue) DomainAsValues() []Value {
	return v.GetDomainAsValues()
}

func (v *FcnRcdValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	if err := v.normalizeFcn(); err != nil {
		return 0, err
	}
	if v.ownedValues() == nil {
		return 0, NewNullPointerException()
	}
	return len(v.ownedValues()), nil
}

func (v *FcnRcdValue) NonNormalizedSize() int {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	return len(v.ownedValues())
}

func (v *FcnRcdValue) ToTuple() *TupleValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	if v.ownedIntv() != nil {
		if v.ownedIntv().Low != 1 {
			size, err := v.ownedIntv().Size()
			if err != nil {
				panic(err)
			}
			if size != 0 {
				return nil
			}
		}
		return NewTupleValue(v.ownedValues())
	}
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	elems := make([]Value, len(v.ownedValues()))
	for i := range v.ownedValues() {
		iv, ok := fcnParameterDomain(v.ownedDomain(), i).(*IntValue)
		if !ok || iv == nil {
			return nil
		}
		idx := int(iv.ownedVal())
		if idx <= 0 || idx > len(v.ownedValues()) || !isNil(elems[idx-1]) {
			return nil
		}
		elems[idx-1] = v.ownedValues()[i]
	}
	v.CM.incValueSecondary(int64(len(elems)))
	return NewTupleValue(elems, v.CM)
}

func (v *FcnRcdValue) ToRcd() *RecordValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	if v.ownedDomain() == nil {
		return nil
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	names := make([]*UniqueString, len(v.ownedDomain()))
	for i, d := range v.ownedDomain() {
		if isNil(d) {
			return nil
		}
		s, ok := asStringValue(d)
		if !ok {
			return nil
		}
		names[i] = s.ownedVal()
	}
	if valueCoverageEnabled() {
		if v.ownedValues() == nil {
			panic(NewNullPointerException())
		}
		v.CM.incValueSecondary(int64(len(v.ownedValues())))
	}
	return NewRecordValue(names, v.ownedValues(), v.IsNorm, v.CM)
}

func (v *FcnRcdValue) ToState() *TLCStateMut {
	if v == nil {
		panic(NewNullPointerException())
	}
	record := v.ToRcd()
	if record == nil {
		return nil
	}
	return record.ToState()
}

func (v *FcnRcdValue) ToRecord() *RecordValue {
	return v.ToRcd()
}

func (v *FcnRcdValue) ToFcnRcd() *FcnRcdValue {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *FcnRcdValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	return v
}

func (v *FcnRcdValue) normalizeFcn() (err error) {
	defer catchValueFailure(v, &err)
	if v.IsNorm {
		return nil
	}
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	for i := 1; i < len(v.ownedDomain()); i++ {
		if isNil(v.ownedDomain()[0]) {
			panic(NewNullPointerException())
		}
		cmp, err := v.ownedDomain()[0].Compare(v.ownedDomain()[i])
		if err != nil {
			return err
		}
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("The value\n%s\noccurs multiple times in the function domain.", v.ownedDomain()[i]))
		}
		if cmp > 0 {
			v.ownedDomain()[0], v.ownedDomain()[i] = v.ownedDomain()[i], v.ownedDomain()[0]
			first := fcnParameterDomain(v.ownedValues(), 0)
			other := fcnParameterDomain(v.ownedValues(), i)
			v.ownedValues()[0], v.ownedValues()[i] = other, first
		}
	}
	for i := 2; i < len(v.ownedDomain()); i++ {
		d := v.ownedDomain()[i]
		val := fcnParameterDomain(v.ownedValues(), i)
		j := i
		cmp := -1
		for j > 0 {
			if isNil(d) {
				panic(NewNullPointerException())
			}
			nextCmp, err := d.Compare(v.ownedDomain()[j-1])
			if err != nil {
				return err
			}
			cmp = nextCmp
			if cmp >= 0 {
				break
			}
			v.ownedDomain()[j] = v.ownedDomain()[j-1]
			v.ownedValues()[j] = fcnParameterDomain(v.ownedValues(), j-1)
			j--
		}
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("The value\n%s\noccurs multiple times in the function domain.", v.ownedDomain()[i]))
		}
		v.ownedDomain()[j] = d
		v.ownedValues()[j] = val
	}
	v.IsNorm = true
	return nil
}

func (v *FcnRcdValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.ownedValues() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		value.DeepNormalize()
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
}

func (v *FcnRcdValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	defined := true
	if v.ownedIntv() == nil {
		for i := range v.ownedValues() {
			if defined {
				value := fcnParameterDomain(v.ownedDomain(), i)
				if isNil(value) {
					panic(NewNullPointerException())
				}
				defined = value.IsDefined()
			}
		}
	}
	for _, value := range v.ownedValues() {
		if defined {
			if isNil(value) {
				panic(NewNullPointerException())
			}
			defined = value.IsDefined()
		}
	}
	return defined
}

func (v *FcnRcdValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	return v.IsNorm
}

func (v *FcnRcdValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, len(v.ownedValues()))
	for i, value := range v.ownedValues() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		values[i] = value.DeepCopy()
	}
	if v.ownedIntv() == nil {
		if v.ownedDomain() == nil {
			panic(NewNullPointerException())
		}
		domain := make([]Value, len(v.ownedDomain()))
		copy(domain, v.ownedDomain())
		return newFcnRcdValue(BaseValue{}, domain, nil, values, false)
	}
	return newFcnRcdValue(BaseValue{}, nil, v.ownedIntv(), values, v.IsNorm)
}

func (v *FcnRcdValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.ownedValues())))
	if v.ownedIntv() == nil {
		for i := range v.ownedValues() {
			domain := fcnParameterDomain(v.ownedDomain(), i)
			if isNil(domain) {
				panic(NewNullPointerException())
			}
			fp = domain.FingerPrint(fp)
			if isNil(v.ownedValues()[i]) {
				panic(NewNullPointerException())
			}
			fp = v.ownedValues()[i].FingerPrint(fp)
		}
		return fp
	}
	for i := range v.ownedValues() {
		fp = FP64ExtendByte(fp, byte(IntValueKind))
		index := int64(v.ownedIntv().Low) + int64(i)
		if index > math.MaxInt32 {
			panic(NewArithmeticException("integer overflow"))
		}
		fp = FP64ExtendInt(fp, int32(index))
		if isNil(v.ownedValues()[i]) {
			panic(NewNullPointerException())
		}
		fp = v.ownedValues()[i].FingerPrint(fp)
	}
	return fp
}

func (v *FcnRcdValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	size, err := v.Size()
	if err != nil {
		panic(err)
	}
	values := make([]Value, size)
	vchanged := false
	for i, value := range v.ownedValues() {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		values[i] = value.Permute(perm)
		vchanged = vchanged || values[i] != value
	}
	if v.ownedIntv() == nil {
		domain := make([]Value, size)
		dchanged := false
		for i := range domain {
			value := fcnParameterDomain(v.ownedDomain(), i)
			if isNil(value) {
				panic(NewNullPointerException())
			}
			domain[i] = value.Permute(perm)
			dchanged = dchanged || domain[i] != value
		}
		if dchanged {
			return newFcnRcdValue(BaseValue{}, domain, nil, values, false)
		}
		if vchanged {
			return newFcnRcdValue(BaseValue{}, v.ownedDomain(), nil, values, true)
		}
		return v
	}
	if vchanged {
		return newFcnRcdValue(BaseValue{}, nil, v.ownedIntv(), values, true)
	}
	return v
}

func (v *FcnRcdValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	return ValueToString(v, "", true)
}

func (v *FcnRcdValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	length := len(v.ownedValues())
	if length == 0 {
		sb.WriteString("<<>>")
	} else if v.isRecordLike() {
		sb.WriteString("[")
		for i := 0; i < length; i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			key, _ := asStringValue(fcnParameterDomain(v.ownedDomain(), i))
			sb.WriteString(key.ownedVal().String() + recordArrow)
			value := fcnParameterDomain(v.ownedValues(), i)
			if isNil(value) {
				panic(NewNullPointerException())
			}
			sb = appendValueString(value, sb, offset, swallow)
		}
		sb.WriteString("]")
	} else if v.isTupleLike() {
		sb.WriteString("<<")
		for i := 0; i < length; i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			value := fcnParameterDomain(v.ownedValues(), i)
			if isNil(value) {
				panic(NewNullPointerException())
			}
			sb = appendValueString(value, sb, offset, swallow)
		}
		sb.WriteString(">>")
	} else {
		domain := v.GetDomainAsValues()
		sb.WriteString("(")
		for i := 0; i < length; i++ {
			if i > 0 {
				sb.WriteString(" @@ ")
			}
			element := fcnParameterDomain(domain, i)
			if isNil(element) {
				panic(NewNullPointerException())
			}
			sb = appendValueString(element, sb, offset, swallow)
			sb.WriteString(" :> ")
			// Formatting a key can normalize a set sharing the values array.
			// Read the current slot only after the key has been rendered.
			value := fcnParameterDomain(v.ownedValues(), i)
			if isNil(value) {
				panic(NewNullPointerException())
			}
			sb = appendValueString(value, sb, offset, swallow)
		}
		sb.WriteString(")")
	}
	return sb
}

func (v *FcnRcdValue) isRecordLike() bool {
	if v.ownedIntv() != nil {
		return false
	}
	if v.ownedDomain() == nil {
		panic(NewNullPointerException())
	}
	for _, dval := range v.ownedDomain() {
		if isNil(dval) {
			return false
		}
		sv, ok := asStringValue(dval)
		if !ok {
			return false
		}
		if sv.ownedVal() == nil {
			panic(NewNullPointerException())
		}
		if !isTLAName(sv.ownedVal().String()) {
			return false
		}
	}
	return true
}

func (v *FcnRcdValue) isTupleLike() bool {
	if v.ownedIntv() != nil {
		return v.ownedIntv().Low == 1 || mustIntervalSize(v.ownedIntv()) == 0
	}
	for _, dval := range v.ownedDomain() {
		if number, ok := dval.(*IntValue); !ok || number == nil {
			return false
		}
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	for i, dval := range v.ownedDomain() {
		if dval.(*IntValue).ownedVal() != int32(i+1) {
			return false
		}
	}
	return true
}

func isTLAName(name string) bool {
	hasLetter := false
	chars := javaStringUTF16(name)
	for _, unit := range chars {
		ch := rune(unit)
		if ch == '_' {
			continue
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return false
		}
		hasLetter = hasLetter || unicode.IsLetter(ch)
	}
	return hasLetter && (len(chars) < 4 || (!strings.HasPrefix(name, "WF_") && !strings.HasPrefix(name, "SF_")))
}

func asRecordValue(value Value) *RecordValue {
	switch v := value.(type) {
	case *RecordValue:
		return v.ToRcd()
	case *CounterExample:
		if v == nil || v.RecordValue == nil {
			return EmptyRecord
		}
		return v.RecordValue
	case *FcnRcdValue:
		return v.ToRcd()
	case *FcnLambdaValue:
		return v.ToRcd()
	case *TupleValue:
		return v.ToRcd()
	default:
		return nil
	}
}

// asConcreteRecordValue matches Java's instanceof RecordValue check. It
// includes CounterExample, which is a RecordValue subclass, but does not
// perform the broader toRcd conversion accepted by asRecordValue.
func asConcreteRecordValue(value Value) (*RecordValue, bool) {
	if isNil(value) {
		return nil, false
	}
	switch v := value.(type) {
	case *RecordValue:
		return v, true
	case *CounterExample:
		return v.RecordValue, true
	default:
		return nil, false
	}
}

func asFcnRcdValue(value Value) *FcnRcdValue {
	switch v := value.(type) {
	case *FcnRcdValue:
		return v.ToFcnRcd()
	case *FcnLambdaValue:
		return v.ToFcnRcd()
	case *TupleValue:
		return v.ToFcnRcd()
	case *RecordValue:
		return v.ToFcnRcd()
	case *CounterExample:
		if v == nil || v.RecordValue == nil {
			return EmptyRecord.ToFcnRcd()
		}
		return v.RecordValue.ToFcnRcd()
	default:
		return nil
	}
}

func mustIntervalSize(intv *IntervalValue) int {
	size, err := intv.Size()
	if err != nil {
		panic(err)
	}
	return size
}
