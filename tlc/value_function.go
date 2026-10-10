package tlc

import (
	"fmt"
	"strconv"
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
	elems  []*ModelValue
	domain []*ModelValue
	count  int
}

func NewMVPerm() *MVPerm {
	modelValues.Lock()
	defer modelValues.Unlock()
	if len(modelValues.mvs) != modelValues.count {
		setModelValuesLocked()
	}
	domain := make([]*ModelValue, len(modelValues.mvs))
	copy(domain, modelValues.mvs)
	return &MVPerm{
		elems:  make([]*ModelValue, len(domain)),
		domain: domain,
	}
}

func (p *MVPerm) Get(value Value) Value {
	mv := value.(*ModelValue)
	res := p.elems[mv.Index]
	if res == nil {
		return nil
	}
	return res
}

func (p *MVPerm) Put(dval, rval *ModelValue) {
	eq, err := dval.Equal(rval)
	if err != nil {
		panic(err)
	}
	if !eq && p.elems[dval.Index] == nil {
		p.elems[dval.Index] = rval
		p.count++
	}
}

func (p *MVPerm) putIndex(index int, elem *ModelValue) {
	if p.elems[index] == nil && elem != nil {
		p.elems[index] = elem
		p.count++
	}
}

func (p *MVPerm) Size() int {
	return p.count
}

func (p *MVPerm) Compose(perm *MVPerm) *MVPerm {
	res := p.emptyLike()
	for i, mv := range p.elems {
		if mv == nil {
			res.putIndex(i, perm.elems[i])
			continue
		}
		mv1 := perm.elems[mv.Index]
		if mv1 == nil {
			res.putIndex(i, mv)
		} else if !p.domain[i].sameModelValue(mv1) {
			res.putIndex(i, mv1)
		}
	}
	return res
}

func (p *MVPerm) Equal(other *MVPerm) bool {
	if p == nil || other == nil {
		return p == other
	}
	if len(p.elems) != len(other.elems) {
		return false
	}
	for i, mv := range p.elems {
		if !mv.sameModelValue(other.elems[i]) {
			return false
		}
	}
	return true
}

func (p *MVPerm) AllModelValues() []*ModelValue {
	values := make([]*ModelValue, 0, p.count)
	for _, mv := range p.elems {
		if mv != nil {
			values = append(values, mv)
		}
	}
	return values
}

func (p *MVPerm) String() string {
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
		b.WriteString(p.domain[i].String())
		b.WriteString(" -> ")
		b.WriteString(mv.String())
		wrote = true
	}
	b.WriteByte(']')
	return b.String()
}

func (p *MVPerm) key() string {
	var b strings.Builder
	for i, mv := range p.elems {
		if mv == nil {
			continue
		}
		b.WriteString(strconv.Itoa(i))
		b.WriteByte('>')
		b.WriteString(strconv.Itoa(mv.Index))
		b.WriteByte(';')
	}
	return b.String()
}

func (p *MVPerm) emptyLike() *MVPerm {
	domain := make([]*ModelValue, len(p.domain))
	copy(domain, p.domain)
	return &MVPerm{
		elems:  make([]*ModelValue, len(p.elems)),
		domain: domain,
	}
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

func (v *ModelValue) Kind() ValueKind    { return ModelValueKind }
func (v *ModelValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *ModelValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if v.Type == typedModelValueUntypedCodeUnit {
		if o, ok := other.(*ModelValue); ok {
			return v.Val.Compare(o.Val), nil
		}
		return -1, nil
	}
	if o, ok := other.(*ModelValue); ok {
		if o.Type == v.Type || o.Type == typedModelValueUntypedCodeUnit {
			return v.Val.Compare(o.Val), nil
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the differently-typed model values %s and %s", ValuesPPR(v), ValuesPPR(o)))
	}
	if other == nil {
		panic(NewNullPointerException())
	}
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the typed model value %s and non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
}

func (v *ModelValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if v.Type == typedModelValueUntypedCodeUnit {
		o, ok := other.(*ModelValue)
		return ok && v.Val.Equal(o.Val), nil
	}
	if o, ok := other.(*ModelValue); ok {
		if o.Type == v.Type || o.Type == typedModelValueUntypedCodeUnit {
			return o.Val == v.Val || o.Val.Equal(v.Val), nil
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of the differently-typed model values %s and %s", ValuesPPR(v), ValuesPPR(o)))
	}
	if other == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of typed model value %s and non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
}

func (v *ModelValue) modelValueCompareTo(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if v.Type != typedModelValueUntypedCodeUnit {
		if other == nil {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare the typed model value %s and the non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return 1, nil
}

func (v *ModelValue) modelValueEquals(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if v.Type != typedModelValueUntypedCodeUnit {
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of the typed model value %s and the non-model value\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return false, nil
}

func (v *ModelValue) modelValueMember(other Value) (resultBool bool, err error) {
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
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the model value %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *ModelValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the model value %s is a finite set.", ValuesPPR(v)))
}

func (v *ModelValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the model value %s.", ValuesPPR(v)))
}

func (v *ModelValue) Normalize() Value   { return v }
func (v *ModelValue) DeepNormalize()     {}
func (v *ModelValue) IsNormalized() bool { return true }
func (v *ModelValue) IsDefined() bool    { return true }
func (v *ModelValue) DeepCopy() Value    { return v }

func (v *ModelValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	if res := perm.Get(v); res != nil {
		return res
	}
	return v
}

func (v *ModelValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	return v.Val.FingerPrint(FP64ExtendByte(fp, byte(ModelValueKind)))
}

func (v *ModelValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the model value %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *ModelValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the model value %s.", ValuesPPR(v)))
	}
	return v, nil
}

func (v *ModelValue) HasData() bool { return v.Data != nil }
func (v *ModelValue) GetData() any  { return v.Data }

func (v *ModelValue) SetData(obj any) any {
	v.Data = obj
	return obj
}

func (v *ModelValue) String() string {
	return ValueToString(v, "", true)
}

func (v *ModelValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	sb.WriteString(v.Val.String())
	return sb
}

func (v *ModelValue) sameModelValue(other *ModelValue) bool {
	if v == nil || other == nil {
		return v == other
	}
	eq, err := v.Equal(other)
	if err != nil {
		panic(err)
	}
	return eq
}

type RecordValue struct {
	BaseValue
	Names  []*UniqueString
	Values []Value
	IsNorm bool
}

var EmptyRecord = &RecordValue{Names: []*UniqueString{}, Values: []Value{}, IsNorm: true}

func NewRecordValue(names []*UniqueString, values []Value, isNorm bool, cms ...CostModel) *RecordValue {
	return &RecordValue{BaseValue: newBaseValue(cms...), Names: names, Values: values, IsNorm: isNorm}
}

// NewRecordValueFromState ports RecordValue(TLCState[, Value]).
func NewRecordValueFromState(state *TLCStateMut, defaultValue ...Value) *RecordValue {
	names := make([]*UniqueString, len(stateVariables))
	values := make([]Value, len(stateVariables))
	for i, variable := range stateVariables {
		names[i] = variable.Name
		values[i] = state.Lookup(variable.Name)
		if values[i] == nil && len(defaultValue) > 0 {
			values[i] = defaultValue[0]
		}
	}
	return NewRecordValue(names, values, false)
}

// NewRecordValueFromStates ports the debugger's paired state constructor,
// including the trailing space on unprimed keys and interleaved field order.
func NewRecordValueFromStates(state, successor *TLCStateMut, defaultValue Value) *RecordValue {
	names := make([]*UniqueString, len(stateVariables)*2)
	values := make([]Value, len(names))
	for i, variable := range stateVariables {
		j := i * 2
		names[j] = UniqueStringOf(variable.Name.String() + " ")
		names[j+1] = UniqueStringOf(variable.Name.String() + "'")
		values[j] = state.Lookup(variable.Name)
		values[j+1] = successor.Lookup(variable.Name)
		if values[j] == nil {
			values[j] = defaultValue
		}
		if values[j+1] == nil {
			values[j+1] = defaultValue
		}
	}
	return NewRecordValue(names, values, false)
}

func NewRecordValueFromInsMap(values *InsMap[*UniqueString, Value]) *RecordValue {
	if values == nil {
		return EmptyRecord
	}
	names := make([]*UniqueString, 0, values.Len())
	vals := make([]Value, 0, values.Len())
	for name, value := range values.All() {
		names = append(names, name)
		vals = append(vals, value)
	}
	return &RecordValue{Names: names, Values: vals}
}

func (v *RecordValue) Kind() ValueKind    { return RecordValueKind }
func (v *RecordValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *RecordValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	rcd := asRecordValue(other)
	if rcd == nil {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare record:\n%s\nwith non-record\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	if err := v.normalizeRecord(); err != nil {
		return 0, err
	}
	if err := rcd.normalizeRecord(); err != nil {
		return 0, err
	}
	if len(v.Names) != len(rcd.Names) {
		return len(v.Names) - len(rcd.Names), nil
	}
	for i := range v.Names {
		if cmp := v.Names[i].Compare(rcd.Names[i]); cmp != 0 {
			return cmp, nil
		}
	}
	for i := range v.Values {
		cmp, err := v.Values[i].Compare(rcd.Values[i])
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *RecordValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	rcd := asRecordValue(other)
	if rcd == nil {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of record:\n%s\nwith non-record\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	if err := v.normalizeRecord(); err != nil {
		return false, err
	}
	if err := rcd.normalizeRecord(); err != nil {
		return false, err
	}
	if len(v.Names) != len(rcd.Names) {
		return false, nil
	}
	for i := range v.Names {
		if !v.Names[i].Equal(rcd.Names[i]) {
			return false, nil
		}
	}
	for i := range v.Values {
		eq, err := v.Values[i].Equal(rcd.Values[i])
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *RecordValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if element:\n%s\nis in the record:\n%s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *RecordValue) IsFinite() (bool, error) { return true, nil }

func (v *RecordValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		if v.Names == nil {
			panic(NewNullPointerException())
		}
		rlen := len(v.Names)
		newValues := make([]Value, rlen)
		arcVal := ex.Current()
		if arc, ok := arcVal.(*StringValue); ok {
			for i := 0; i < rlen; i++ {
				if v.Names[i] == nil {
					panic(NewNullPointerException())
				}
				if v.Names[i].Equal(arc.Val) {
					ex.Index++
					value := fcnParameterDomain(v.Values, i)
					if value == nil {
						panic(NewNullPointerException())
					}
					taken, err := value.TakeExcept(ex)
					if err != nil {
						return nil, err
					}
					newValues[i] = taken
				} else {
					newValues[i] = fcnParameterDomain(v.Values, i)
				}
			}
			newNames := v.Names
			if !v.IsNorm {
				newNames = make([]*UniqueString, len(v.Names))
				copy(newNames, v.Names)
			}
			return &RecordValue{Names: newNames, Values: newValues, IsNorm: v.IsNorm}, nil
		}
		if arcVal == nil {
			panic(NewNullPointerException())
		}
		PrintWarning(ECTLCWrongRecordFieldName, ValuesPPR(arcVal))
	}
	return ex.Value, nil
}

func (v *RecordValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	var cur Value = v
	for _, ex := range exs {
		if cur == nil {
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

func (v *RecordValue) ToFcnRcd() *FcnRcdValue {
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	domain := make([]Value, len(v.Names))
	for i, name := range v.Names {
		domain[i] = NewStringValueFromUnique(name, v.CM)
	}
	v.CM.incValueSecondary(int64(len(domain)))
	return NewFcnRcdValue(domain, v.Values, v.IsNorm, v.CM)
}

func (v *RecordValue) ToTuple() *TupleValue {
	if len(v.Names) == 0 {
		return EmptyTuple
	}
	return nil
}

func (v *RecordValue) ToState() *TLCStateMut {
	if v == nil {
		return nil
	}
	state := NewEmptyState()
	for _, variable := range StateVariables() {
		for i, name := range v.Names {
			if name.Equal(variable.Name) {
				state.Bind(variable.Name, v.Values[i])
			}
		}
	}
	// PrintTLCState has its own base metadata and delegates state operations
	// to this separate owner. Keep a value-slice alias for native evaluators.
	wrapper := NewEmptyState()
	wrapper.values = state.values
	wrapper.printRecord, wrapper.printState = v, state
	return wrapper
}

func (v *RecordValue) StateString() string {
	if v == nil {
		return ""
	}
	format := ""
	formatIndex := -1
	for i, name := range v.Names {
		if name == UniqueStringOf("_format") {
			if isNil(v.Values[i]) {
				panic(NewNullPointerException("Cannot read field \"val\" because \"this.rcd.values[idx]\" is null"))
			}
			var sv *StringValue
			switch value := v.Values[i].(type) {
			case *StringValue:
				sv = value
			case *DebuggerValue:
				sv = value.StringValue
			default:
				panic(valueStreamClassCast(v.Values[i], "tlc2.value.impl.StringValue"))
			}
			if sv.Val == nil {
				panic(NewNullPointerException("Cannot invoke \"util.UniqueString.toString()\" because \"this.rcd.values[idx].val\" is null"))
			}
			format = sv.Val.String()
			formatIndex = i
			break
		}
	}
	if formatIndex < 0 {
		if len(v.Names) == 1 {
			format = "%s = %s\n"
		} else {
			format = "/\\ %s = %s\n"
		}
	}
	var b strings.Builder
	for i, name := range v.Names {
		if i == formatIndex {
			continue
		}
		text, err := JavaFormatStrings(format, name.String(), ValuesPPR(v.Values[i]))
		if err != nil {
			panic(err)
		}
		b.WriteString(text)
	}
	return javaStringFromUTF16(javaStringUTF16(b.String()))
}

func (v *RecordValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	return len(v.Names), nil
}

func (v *RecordValue) Apply(arg Value) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if arg == nil {
		return nil, NewNullPointerException()
	}
	if debugger, ok := arg.(*DebuggerValue); ok {
		arg = debugger.StringValue
	}
	if sv, ok := arg.(*StringValue); ok {
		for i, name := range v.Names {
			if sv.Val.Equal(name) {
				return v.Values[i], nil
			}
		}
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to access nonexistent field '%s' of record\n%s", sv.Val, ValuesPPR(v)))
	}
	return nil, v.runtimeFailure(fmt.Sprintf("Attempted to access record by a non-string argument: %s", ValuesPPR(arg)))
}

// ApplyArgs retains RecordValue's array overload and its additional source
// catch boundary around single-argument application.
func (v *RecordValue) ApplyArgs(args []Value, control int) (resultValue Value, err error) {
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
	defer catchValueFailure(v, &err)
	if arg == nil {
		// Java formats arg.toString() before Assert.fail for a non-string.
		panic(NewNullPointerException())
	}
	sv, ok := arg.(*StringValue)
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
	for i, name := range v.Names {
		if sv.Val.Equal(name) {
			return v.Values[i], nil
		}
	}
	return nil, nil
}

func (v *RecordValue) DomainValue() Value {
	defer catchValueFailure(v, nil)
	values := make([]Value, len(v.Names))
	for i, name := range v.Names {
		values[i] = NewStringValueFromUnique(name)
	}
	return NewSetEnumValue(values, v.IsNormalized())
}

func (v *RecordValue) Normalize() Value {
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
	for i := 1; i < len(v.Names); i++ {
		cmp := v.Names[0].Compare(v.Names[i])
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("Field name %s occurs multiple times in record.", v.Names[i]))
		}
		if cmp > 0 {
			v.Names[0], v.Names[i] = v.Names[i], v.Names[0]
			v.Values[0], v.Values[i] = v.Values[i], v.Values[0]
		}
	}
	for i := 2; i < len(v.Names); i++ {
		j := i
		st := v.Names[i]
		val := v.Values[i]
		cmp := -1
		for j > 0 {
			cmp = st.Compare(v.Names[j-1])
			if cmp >= 0 {
				break
			}
			v.Names[j] = v.Names[j-1]
			v.Values[j] = v.Values[j-1]
			j--
		}
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("Field name %s occurs multiple times in record.", v.Names[i]))
		}
		v.Names[j] = st
		v.Values[j] = val
	}
	v.IsNorm = true
	return nil
}

func (v *RecordValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	for _, value := range v.Values {
		value.DeepNormalize()
	}
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
}

func (v *RecordValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	for _, value := range v.Values {
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *RecordValue) IsNormalized() bool { return v.IsNorm }

func (v *RecordValue) DeepCopy() Value {
	defer catchValueFailure(v, nil)
	values := make([]Value, len(v.Values))
	for i, value := range v.Values {
		values[i] = value.DeepCopy()
	}
	names := make([]*UniqueString, len(v.Names))
	copy(names, v.Names)
	return &RecordValue{Names: names, Values: values, IsNorm: v.IsNorm}
}

func (v *RecordValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.Names)))
	for i, name := range v.Names {
		fp = FP64ExtendByte(fp, byte(StringValueKind))
		fp = FP64ExtendInt(fp, int32(name.Length()))
		fp = FP64ExtendString(fp, name.String())
		fp = v.Values[i].FingerPrint(fp)
	}
	return fp
}

func (v *RecordValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	values := make([]Value, len(v.Values))
	changed := false
	for i, value := range v.Values {
		values[i] = value.Permute(perm)
		changed = changed || values[i] != value
	}
	if changed {
		return &RecordValue{Names: v.Names, Values: values, IsNorm: true}
	}
	return v
}

func (v *RecordValue) String() string {
	return ValueToString(v, "", true)
}

func (v *RecordValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	sb.WriteString("[")
	for i, name := range v.Names {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name.String() + recordArrow)
		sb = appendValueString(v.Values[i], sb, offset, swallow)
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
}

var EmptyFcn = &FcnRcdValue{Domain: []Value{}, Values: []Value{}, IsNorm: true}

func NewFcnRcdValue(domain []Value, values []Value, isNorm bool, cms ...CostModel) *FcnRcdValue {
	return &FcnRcdValue{BaseValue: newBaseValue(cms...), Domain: domain, Values: values, IsNorm: isNorm}
}

func NewFcnRcdIntervalValue(intv *IntervalValue, values []Value, cms ...CostModel) *FcnRcdValue {
	return &FcnRcdValue{BaseValue: newBaseValue(cms...), Intv: intv, Values: values, IsNorm: true}
}

func (v *FcnRcdValue) Kind() ValueKind    { return FcnRcdValueKind }
func (v *FcnRcdValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *FcnRcdValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	fcn := asFcnRcdValue(other)
	if fcn == nil {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		selfText := ValuesPPR(v)
		if other == nil {
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
	if len(v.Values) != len(fcn.Values) {
		return len(v.Values) - len(fcn.Values), nil
	}
	if v.Intv != nil {
		return v.compareToInterval(fcn)
	}
	return v.compareOtherInterval(fcn)
}

func (v *FcnRcdValue) compareOtherInterval(fcn *FcnRcdValue) (int, error) {
	if fcn.Intv != nil {
		for i, dElem := range v.Domain {
			iv, ok := dElem.(*IntValue)
			if !ok {
				return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare integer with non-integer\n%s.", ValuesPPR(dElem)))
			}
			intervalElement := int64(fcn.Intv.Low) + int64(i)
			domainElement := int64(iv.Val)
			if domainElement < intervalElement {
				return -1, nil
			}
			if domainElement > intervalElement {
				return 1, nil
			}
		}
		for i := range v.Domain {
			cmp, err := v.Values[i].Compare(fcn.Values[i])
			if err != nil || cmp != 0 {
				return cmp, err
			}
		}
		return 0, nil
	}
	for i := range v.Domain {
		cmp, err := v.Domain[i].Compare(fcn.Domain[i])
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	for i := range v.Domain {
		cmp, err := v.Values[i].Compare(fcn.Values[i])
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *FcnRcdValue) compareToInterval(fcn *FcnRcdValue) (int, error) {
	if len(v.Values) == 0 {
		return 0, nil
	}
	if fcn.Intv != nil {
		if v.Intv.Low < fcn.Intv.Low {
			return -1, nil
		}
		if v.Intv.Low > fcn.Intv.Low {
			return 1, nil
		}
		for i := range v.Values {
			cmp, err := v.Values[i].Compare(fcn.Values[i])
			if err != nil || cmp != 0 {
				return cmp, err
			}
		}
		return 0, nil
	}
	for i, dElem := range fcn.Domain {
		iv, ok := dElem.(*IntValue)
		if !ok {
			return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare integer with non-integer:\n%s.", ValuesPPR(dElem)))
		}
		intervalElement := int64(v.Intv.Low) + int64(i)
		domainElement := int64(iv.Val)
		if intervalElement < domainElement {
			return -1, nil
		}
		if intervalElement > domainElement {
			return 1, nil
		}
	}
	for i := range fcn.Domain {
		cmp, err := v.Values[i].Compare(fcn.Values[i])
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}

func (v *FcnRcdValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	fcn := asFcnRcdValue(other)
	if fcn == nil {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		selfText := ValuesPPR(v)
		if other == nil {
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
	if v.Intv != nil {
		if fcn.Intv != nil {
			eq, err := v.Intv.Equal(fcn.Intv)
			if err != nil || !eq {
				return eq, err
			}
			for i := range v.Values {
				eq, err := v.Values[i].Equal(fcn.Values[i])
				if err != nil || !eq {
					return eq, err
				}
			}
			return true, nil
		}
		if len(fcn.Domain) != mustIntervalSize(v.Intv) {
			return false, nil
		}
		for i, dElem := range fcn.Domain {
			iv, ok := dElem.(*IntValue)
			if !ok {
				return false, v.runtimeFailure(fmt.Sprintf("Attempted to compare an integer with non-integer:\n%s.", ValuesPPR(dElem)))
			}
			if int64(iv.Val) != int64(v.Intv.Low)+int64(i) {
				return false, nil
			}
		}
		for i := range fcn.Values {
			eq, err := v.Values[i].Equal(fcn.Values[i])
			if err != nil || !eq {
				return eq, err
			}
		}
		return true, nil
	}
	if len(v.Values) != len(fcn.Values) {
		return false, nil
	}
	if fcn.Intv != nil {
		for i, dElem := range v.Domain {
			iv, ok := dElem.(*IntValue)
			if !ok {
				return false, v.runtimeFailure(fmt.Sprintf("Attempted to compare an integer with non-integer:\n%s.", ValuesPPR(dElem)))
			}
			if int64(iv.Val) != int64(fcn.Intv.Low)+int64(i) {
				return false, nil
			}
		}
		for i := range v.Values {
			eq, err := v.Values[i].Equal(fcn.Values[i])
			if err != nil || !eq {
				return eq, err
			}
		}
		return true, nil
	}
	for i := range v.Domain {
		eq, err := v.Domain[i].Equal(fcn.Domain[i])
		if err != nil || !eq {
			return eq, err
		}
	}
	for i := range v.Values {
		eq, err := v.Values[i].Equal(fcn.Values[i])
		if err != nil || !eq {
			return eq, err
		}
	}
	return true, nil
}

func (v *FcnRcdValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the function %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *FcnRcdValue) IsFinite() (bool, error) { return true, nil }

func (v *FcnRcdValue) Apply(arg Value) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	result, err := v.Select(arg)
	if err != nil {
		return nil, err
	}
	if result == nil {
		selfText := ValuesPPR(v)
		if arg == nil {
			panic(NewNullPointerException())
		}
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply function:\n%s\nto argument %s, which is not in the domain of the function.", selfText, ValuesPPR(arg)))
	}
	return result, nil
}

func (v *FcnRcdValue) Select(arg Value) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if v.Intv != nil {
		iv, ok := arg.(*IntValue)
		if !ok {
			if arg == nil {
				panic(NewNullPointerException())
			}
			return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply function with integer domain to the non-integer argument %s", ValuesPPR(arg)))
		}
		if iv.Val >= v.Intv.Low && iv.Val <= v.Intv.High {
			offset := int64(iv.Val) - int64(v.Intv.Low)
			if offset < int64(len(v.Values)) {
				return v.Values[int(offset)], nil
			}
		}
		return nil, nil
	}
	if v.IsNorm && len(v.Domain) >= fcnRcdLinearSearchThreshold() {
		// Arrays.binarySearch uses an inclusive upper bound and returns at the
		// first equal comparison. The visited values also determine which typed
		// model-value comparison reports an error.
		low, high := 0, len(v.Domain)-1
		for low <= high {
			mid := (low + high) >> 1
			cmp, err := v.Domain[mid].Compare(arg)
			if err != nil {
				return nil, err
			}
			if cmp < 0 {
				low = mid + 1
			} else if cmp > 0 {
				high = mid - 1
			} else {
				eq, err := v.Domain[mid].Equal(arg)
				if err != nil || !eq {
					return nil, err
				}
				return v.Values[mid], nil
			}
		}
		return nil, nil
	}
	for i, value := range v.Domain {
		eq, err := value.Equal(arg)
		if err != nil {
			return nil, err
		}
		if eq {
			return v.Values[i], nil
		}
	}
	return nil, nil
}

func fcnRcdLinearSearchThreshold() int {
	if value, ok := tlcLookupSystemProperty(fcnRcdLinearSearchThresholdProperty); ok {
		if parsed, ok := javaIntProperty(value); ok {
			return parsed
		}
	}
	return defaultFcnRcdLinearSearchThreshold
}

func (v *FcnRcdValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
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
	newValues := make([]Value, len(v.Values))
	copy(newValues, v.Values)
	arg := ex.Current()
	if v.Intv != nil {
		if iv, ok := arg.(*IntValue); ok {
			if iv.Val >= v.Intv.Low && iv.Val <= v.Intv.High {
				offset := int64(iv.Val) - int64(v.Intv.Low)
				if offset >= int64(len(newValues)) {
					return v, nil
				}
				vidx := int(offset)
				next := ex
				next.Index++
				taken, err := v.Values[vidx].TakeExcept(next)
				if err != nil {
					return nil, err
				}
				newValues[vidx] = taken
			}
			return &FcnRcdValue{Intv: v.Intv, Values: newValues, IsNorm: true}, nil
		}
		return v, nil
	}
	for i := range v.Values {
		domain := fcnParameterDomain(v.Domain, i)
		if arg == nil {
			panic(NewNullPointerException())
		}
		eq, err := arg.Equal(domain)
		if err != nil {
			return nil, err
		}
		if eq {
			next := ex
			next.Index++
			taken, err := newValues[i].TakeExcept(next)
			if err != nil {
				return nil, err
			}
			newValues[i] = taken
			newDomain := v.Domain
			if !v.IsNorm {
				newDomain = make([]Value, len(v.Domain))
				copy(newDomain, v.Domain)
			}
			return &FcnRcdValue{Domain: newDomain, Values: newValues, IsNorm: v.IsNorm}, nil
		}
	}
	return v, nil
}

func (v *FcnRcdValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	var cur Value = v
	for _, ex := range exs {
		if cur == nil {
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
	defer catchValueFailure(v, nil)
	if v.Intv != nil {
		return v.Intv
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	return NewSetEnumValue(v.Domain, true)
}

func (v *FcnRcdValue) DomainAsValues() []Value {
	if v.Intv != nil {
		return v.Intv.AsValues()
	}
	return v.Domain
}

func (v *FcnRcdValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if err := v.normalizeFcn(); err != nil {
		return 0, err
	}
	return len(v.Values), nil
}

func (v *FcnRcdValue) NonNormalizedSize() int { return len(v.Values) }

func (v *FcnRcdValue) ToTuple() *TupleValue {
	if v.Intv != nil {
		if v.Intv.Low != 1 {
			size, err := v.Intv.Size()
			if err != nil {
				panic(err)
			}
			if size != 0 {
				return nil
			}
		}
		return NewTupleValue(v.Values)
	}
	elems := make([]Value, len(v.Values))
	for i := range v.Values {
		iv, ok := v.Domain[i].(*IntValue)
		if !ok {
			return nil
		}
		idx := int(iv.Val)
		if idx <= 0 || idx > len(v.Values) || elems[idx-1] != nil {
			return nil
		}
		elems[idx-1] = v.Values[i]
	}
	v.CM.incValueSecondary(int64(len(elems)))
	return NewTupleValue(elems, v.CM)
}

func (v *FcnRcdValue) ToRecord() *RecordValue {
	if v.Domain == nil {
		return nil
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	names := make([]*UniqueString, len(v.Domain))
	for i, d := range v.Domain {
		s, ok := d.(*StringValue)
		if !ok {
			return nil
		}
		names[i] = s.Val
	}
	v.CM.incValueSecondary(int64(len(v.Values)))
	return NewRecordValue(names, v.Values, v.IsNorm, v.CM)
}

func (v *FcnRcdValue) Normalize() Value {
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
	for i := 1; i < len(v.Domain); i++ {
		cmp, err := v.Domain[0].Compare(v.Domain[i])
		if err != nil {
			return err
		}
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("The value\n%s\noccurs multiple times in the function domain.", v.Domain[i]))
		}
		if cmp > 0 {
			v.Domain[0], v.Domain[i] = v.Domain[i], v.Domain[0]
			v.Values[0], v.Values[i] = v.Values[i], v.Values[0]
		}
	}
	for i := 2; i < len(v.Domain); i++ {
		d := v.Domain[i]
		val := v.Values[i]
		j := i
		cmp := -1
		for j > 0 {
			nextCmp, err := d.Compare(v.Domain[j-1])
			if err != nil {
				return err
			}
			cmp = nextCmp
			if cmp >= 0 {
				break
			}
			v.Domain[j] = v.Domain[j-1]
			v.Values[j] = v.Values[j-1]
			j--
		}
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("The value\n%s\noccurs multiple times in the function domain.", v.Domain[i]))
		}
		v.Domain[j] = d
		v.Values[j] = val
	}
	v.IsNorm = true
	return nil
}

func (v *FcnRcdValue) DeepNormalize() {
	defer catchValueFailure(v, nil)
	for _, value := range v.Values {
		value.DeepNormalize()
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
}

func (v *FcnRcdValue) IsDefined() bool {
	defer catchValueFailure(v, nil)
	if v.Intv == nil {
		for _, value := range v.Domain {
			if !value.IsDefined() {
				return false
			}
		}
	}
	for _, value := range v.Values {
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *FcnRcdValue) IsNormalized() bool { return v.IsNorm }

func (v *FcnRcdValue) DeepCopy() Value {
	defer catchValueFailure(v, nil)
	values := make([]Value, len(v.Values))
	for i, value := range v.Values {
		values[i] = value.DeepCopy()
	}
	if v.Intv == nil {
		domain := make([]Value, len(v.Domain))
		copy(domain, v.Domain)
		return &FcnRcdValue{Domain: domain, Values: values}
	}
	return &FcnRcdValue{Intv: v.Intv, Values: values, IsNorm: v.IsNorm}
}

func (v *FcnRcdValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.Values)))
	if v.Intv == nil {
		for i := range v.Values {
			fp = v.Domain[i].FingerPrint(fp)
			fp = v.Values[i].FingerPrint(fp)
		}
		return fp
	}
	for i := range v.Values {
		fp = FP64ExtendByte(fp, byte(IntValueKind))
		fp = FP64ExtendInt(fp, v.Intv.Low+int32(i))
		fp = v.Values[i].FingerPrint(fp)
	}
	return fp
}

func (v *FcnRcdValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	values := make([]Value, len(v.Values))
	vchanged := false
	for i, value := range v.Values {
		values[i] = value.Permute(perm)
		vchanged = vchanged || values[i] != value
	}
	if v.Intv == nil {
		domain := make([]Value, len(v.Domain))
		dchanged := false
		for i, value := range v.Domain {
			domain[i] = value.Permute(perm)
			dchanged = dchanged || domain[i] != value
		}
		if dchanged {
			return &FcnRcdValue{Domain: domain, Values: values}
		}
		if vchanged {
			return &FcnRcdValue{Domain: v.Domain, Values: values, IsNorm: true}
		}
		return v
	}
	if vchanged {
		return &FcnRcdValue{Intv: v.Intv, Values: values, IsNorm: true}
	}
	return v
}

func (v *FcnRcdValue) String() string {
	return ValueToString(v, "", true)
}

func (v *FcnRcdValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if len(v.Values) == 0 {
		sb.WriteString("<<>>")
	} else if v.isRecordLike() {
		sb.WriteString("[")
		for i, value := range v.Values {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(v.Domain[i].(*StringValue).Val.String() + recordArrow)
			if value == nil {
				panic(NewNullPointerException())
			}
			sb = appendValueString(value, sb, offset, swallow)
		}
		sb.WriteString("]")
	} else if v.isTupleLike() {
		sb.WriteString("<<")
		for i, value := range v.Values {
			if i > 0 {
				sb.WriteString(", ")
			}
			if value == nil {
				panic(NewNullPointerException())
			}
			sb = appendValueString(value, sb, offset, swallow)
		}
		sb.WriteString(">>")
	} else {
		domain := v.DomainAsValues()
		sb.WriteString("(")
		for i, value := range v.Values {
			if i > 0 {
				sb.WriteString(" @@ ")
			}
			if domain[i] == nil {
				panic(NewNullPointerException())
			}
			sb = appendValueString(domain[i], sb, offset, swallow)
			sb.WriteString(" :> ")
			if value == nil {
				panic(NewNullPointerException())
			}
			sb = appendValueString(value, sb, offset, swallow)
		}
		sb.WriteString(")")
	}
	return sb
}

func (v *FcnRcdValue) isRecordLike() bool {
	if v.Intv != nil {
		return false
	}
	for _, dval := range v.Domain {
		sv, ok := dval.(*StringValue)
		if !ok || !isTLAName(sv.Val.String()) {
			return false
		}
	}
	return true
}

func (v *FcnRcdValue) isTupleLike() bool {
	if v.Intv != nil {
		return v.Intv.Low == 1 || mustIntervalSize(v.Intv) == 0
	}
	for _, dval := range v.Domain {
		if _, ok := dval.(*IntValue); !ok {
			return false
		}
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	for i, dval := range v.Domain {
		if dval.(*IntValue).Val != int32(i+1) {
			return false
		}
	}
	return true
}

func isTLAName(name string) bool {
	hasLetter := false
	for _, ch := range name {
		if ch == '_' {
			continue
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return false
		}
		hasLetter = hasLetter || unicode.IsLetter(ch)
	}
	return hasLetter && (len(name) < 4 || (!strings.HasPrefix(name, "WF_") && !strings.HasPrefix(name, "SF_")))
}

func asRecordValue(value Value) *RecordValue {
	switch v := value.(type) {
	case *RecordValue:
		return v
	case *CounterExample:
		if v == nil || v.RecordValue == nil {
			return EmptyRecord
		}
		return v.RecordValue
	case *FcnRcdValue:
		return v.ToRecord()
	case *FcnLambdaValue:
		return v.ToRecord()
	case *TupleValue:
		if len(v.Elems) == 0 {
			return EmptyRecord
		}
		return nil
	default:
		return nil
	}
}

func asFcnRcdValue(value Value) *FcnRcdValue {
	switch v := value.(type) {
	case *FcnRcdValue:
		return v
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
