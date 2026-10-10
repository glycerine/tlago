package tlc

import (
	"fmt"
	"math"
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
	if perm == nil {
		panic(NewNullPointerException())
	}
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
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the model value %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *ModelValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
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
		selfText := ValuesPPR(v)
		if other == nil {
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
	if v.Names == nil || rcd.Names == nil {
		panic(NewNullPointerException())
	}
	length := len(v.Names)
	if length != len(rcd.Names) {
		return len(v.Names) - len(rcd.Names), nil
	}
	for i := 0; i < length; i++ {
		if v.Names[i] == nil || rcd.Names[i] == nil {
			panic(NewNullPointerException())
		}
		if cmp := v.Names[i].Compare(rcd.Names[i]); cmp != 0 {
			return cmp, nil
		}
	}
	for i := 0; i < length; i++ {
		left := fcnParameterDomain(v.Values, i)
		right := fcnParameterDomain(rcd.Values, i)
		if left == nil {
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
	defer catchValueFailure(v, &err)
	rcd := asRecordValue(other)
	if rcd == nil {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		selfText := ValuesPPR(v)
		if other == nil {
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
	if v.Names == nil || rcd.Names == nil {
		panic(NewNullPointerException())
	}
	length := len(v.Names)
	if length != len(rcd.Names) {
		return false, nil
	}
	for i := 0; i < length; i++ {
		if v.Names[i] == nil || rcd.Names[i] == nil {
			panic(NewNullPointerException())
		}
		if !v.Names[i].Equal(rcd.Names[i]) {
			return false, nil
		}
	}
	for i := 0; i < length; i++ {
		left := fcnParameterDomain(v.Values, i)
		right := fcnParameterDomain(rcd.Values, i)
		if left == nil {
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
		if arc, ok := asStringValue(arcVal); ok {
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
	if v.Names == nil {
		panic(NewNullPointerException())
	}
	domain := make([]Value, len(v.Names))
	for i, name := range v.Names {
		domain[i] = NewStringValueFromUnique(name, v.CM)
	}
	v.CM.incValueSecondary(int64(len(domain)))
	return NewFcnRcdValue(domain, v.Values, v.IsNorm, v.CM)
}

func (v *RecordValue) ToTuple() *TupleValue {
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
	if v.Names == nil {
		return 0, NewNullPointerException()
	}
	return len(v.Names), nil
}

func (v *RecordValue) Apply(arg Value) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if arg == nil {
		return nil, NewNullPointerException()
	}
	if sv, ok := asStringValue(arg); ok {
		if v.Names == nil {
			panic(NewNullPointerException())
		}
		for i, name := range v.Names {
			if sv.Val == nil || name == nil {
				panic(NewNullPointerException())
			}
			if sv.Val.Equal(name) {
				return fcnParameterDomain(v.Values, i), nil
			}
		}
		name := "null"
		if sv.Val != nil {
			name = sv.Val.String()
		}
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to access nonexistent field '%s' of record\n%s", name, ValuesPPR(v)))
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
	if v.Names == nil {
		panic(NewNullPointerException())
	}
	for i, name := range v.Names {
		if sv.Val == nil || name == nil {
			panic(NewNullPointerException())
		}
		if sv.Val.Equal(name) {
			return fcnParameterDomain(v.Values, i), nil
		}
	}
	return nil, nil
}

func (v *RecordValue) DomainValue() Value {
	defer catchValueFailure(v, nil)
	if v.Names == nil {
		panic(NewNullPointerException())
	}
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
	if v.Names == nil {
		panic(NewNullPointerException())
	}
	for i := 1; i < len(v.Names); i++ {
		if v.Names[0] == nil || v.Names[i] == nil {
			panic(NewNullPointerException())
		}
		cmp := v.Names[0].Compare(v.Names[i])
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("Field name %s occurs multiple times in record.", v.Names[i]))
		}
		if cmp > 0 {
			v.Names[0], v.Names[i] = v.Names[i], v.Names[0]
			first := fcnParameterDomain(v.Values, 0)
			other := fcnParameterDomain(v.Values, i)
			v.Values[0], v.Values[i] = other, first
		}
	}
	for i := 2; i < len(v.Names); i++ {
		j := i
		st := v.Names[i]
		val := fcnParameterDomain(v.Values, i)
		cmp := -1
		for j > 0 {
			if st == nil || v.Names[j-1] == nil {
				panic(NewNullPointerException())
			}
			cmp = st.Compare(v.Names[j-1])
			if cmp >= 0 {
				break
			}
			v.Names[j] = v.Names[j-1]
			v.Values[j] = fcnParameterDomain(v.Values, j-1)
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
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.Values {
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
	defer catchValueFailure(v, nil)
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.Values {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *RecordValue) IsNormalized() bool { return v.IsNorm }

func (v *RecordValue) DeepCopy() Value {
	defer catchValueFailure(v, nil)
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, len(v.Values))
	for i, value := range v.Values {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		values[i] = value.DeepCopy()
	}
	if v.Names == nil {
		panic(NewNullPointerException())
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
	if v.Names == nil {
		panic(NewNullPointerException())
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.Names)))
	for i, name := range v.Names {
		if name == nil {
			panic(NewNullPointerException())
		}
		fp = FP64ExtendByte(fp, byte(StringValueKind))
		fp = FP64ExtendInt(fp, int32(name.Length()))
		fp = FP64ExtendString(fp, name.String())
		value := fcnParameterDomain(v.Values, i)
		if isNil(value) {
			panic(NewNullPointerException())
		}
		fp = value.FingerPrint(fp)
	}
	return fp
}

func (v *RecordValue) Permute(perm *MVPerm) Value {
	defer catchValueFailure(v, nil)
	if err := v.normalizeRecord(); err != nil {
		panic(err)
	}
	if v.Names == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, len(v.Names))
	changed := false
	for i := range values {
		value := fcnParameterDomain(v.Values, i)
		if isNil(value) {
			panic(NewNullPointerException())
		}
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
	if v.Names == nil {
		panic(NewNullPointerException())
	}
	sb.WriteString("[")
	for i, name := range v.Names {
		if i > 0 {
			sb.WriteString(", ")
		}
		text := "null"
		if name != nil {
			text = name.String()
		}
		sb.WriteString(text + recordArrow)
		value := fcnParameterDomain(v.Values, i)
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
}

var EmptyFcn = &FcnRcdValue{Domain: []Value{}, Values: []Value{}, IsNorm: true}

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
	EmptyFcn = &FcnRcdValue{Domain: []Value{}, Values: []Value{}, IsNorm: true}
}

func NewFcnRcdValue(domain []Value, values []Value, isNorm bool, cms ...CostModel) *FcnRcdValue {
	fcnRcdLinearSearchThreshold()
	return &FcnRcdValue{BaseValue: newBaseValue(cms...), Domain: domain, Values: values, IsNorm: isNorm}
}

func NewFcnRcdIntervalValue(intv *IntervalValue, values []Value, cms ...CostModel) *FcnRcdValue {
	fcnRcdLinearSearchThreshold()
	return &FcnRcdValue{BaseValue: newBaseValue(cms...), Intv: intv, Values: values, IsNorm: true}
}

func (v *FcnRcdValue) Kind() ValueKind {
	fcnRcdLinearSearchThreshold()
	return FcnRcdValueKind
}
func (v *FcnRcdValue) KindString() string {
	fcnRcdLinearSearchThreshold()
	return v.KindStringFor(v.Kind())
}

func (v *FcnRcdValue) Compare(other Value) (resultInt int, err error) {
	fcnRcdLinearSearchThreshold()
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
	if v.Values == nil || fcn.Values == nil {
		panic(NewNullPointerException())
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
	if v.Domain == nil {
		panic(NewNullPointerException())
	}
	if fcn.Intv != nil {
		for i := 0; i < len(v.Domain); i++ {
			dElem := fcnParameterDomain(v.Domain, i)
			iv, ok := dElem.(*IntValue)
			if !ok {
				if dElem == nil {
					panic(NewNullPointerException())
				}
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
		for i := 0; i < len(v.Domain); i++ {
			left := fcnParameterDomain(v.Values, i)
			right := fcnParameterDomain(fcn.Values, i)
			if left == nil {
				panic(NewNullPointerException())
			}
			cmp, err := left.Compare(right)
			if err != nil || cmp != 0 {
				return cmp, err
			}
		}
		return 0, nil
	}
	for i := 0; i < len(v.Domain); i++ {
		left := fcnParameterDomain(v.Domain, i)
		right := fcnParameterDomain(fcn.Domain, i)
		if left == nil {
			panic(NewNullPointerException())
		}
		cmp, err := left.Compare(right)
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	for i := 0; i < len(v.Domain); i++ {
		left := fcnParameterDomain(v.Values, i)
		right := fcnParameterDomain(fcn.Values, i)
		if left == nil {
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
		for i := 0; i < len(v.Values); i++ {
			left := fcnParameterDomain(v.Values, i)
			right := fcnParameterDomain(fcn.Values, i)
			if left == nil {
				panic(NewNullPointerException())
			}
			cmp, err := left.Compare(right)
			if err != nil || cmp != 0 {
				return cmp, err
			}
		}
		return 0, nil
	}
	if fcn.Domain == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < len(fcn.Domain); i++ {
		dElem := fcnParameterDomain(fcn.Domain, i)
		iv, ok := dElem.(*IntValue)
		if !ok {
			if dElem == nil {
				panic(NewNullPointerException())
			}
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
	for i := 0; i < len(fcn.Domain); i++ {
		left := fcnParameterDomain(v.Values, i)
		right := fcnParameterDomain(fcn.Values, i)
		if left == nil {
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
	fcnRcdLinearSearchThreshold()
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
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the function %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *FcnRcdValue) IsFinite() (bool, error) {
	fcnRcdLinearSearchThreshold()
	return true, nil
}

func (v *FcnRcdValue) Apply(arg Value) (resultValue Value, err error) {
	fcnRcdLinearSearchThreshold()
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
	fcnRcdLinearSearchThreshold()
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
			if v.Values == nil {
				panic(NewNullPointerException())
			}
			if offset < int64(len(v.Values)) {
				return v.Values[int(offset)], nil
			}
		}
		return nil, nil
	}
	if v.Domain == nil {
		panic(NewNullPointerException())
	}
	if v.IsNorm && len(v.Domain) >= fcnRcdLinearSearchThreshold() {
		// Arrays.binarySearch uses an inclusive upper bound and returns at the
		// first equal comparison. The visited values also determine which typed
		// model-value comparison reports an error.
		low, high := 0, len(v.Domain)-1
		for low <= high {
			mid := (low + high) >> 1
			if isNil(v.Domain[mid]) {
				panic(NewNullPointerException())
			}
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
				return fcnParameterDomain(v.Values, mid), nil
			}
		}
		return nil, nil
	}
	for i, value := range v.Domain {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		eq, err := value.Equal(arg)
		if err != nil {
			return nil, err
		}
		if eq {
			return fcnParameterDomain(v.Values, i), nil
		}
	}
	return nil, nil
}

func fcnRcdLinearSearchThreshold() int {
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
	fcnRcdLinearSearchThreshold()
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
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.Intv != nil {
		return v.Intv
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	if v.Domain == nil {
		panic(NewNullPointerException())
	}
	return NewSetEnumValue(v.Domain, true)
}

func (v *FcnRcdValue) DomainAsValues() []Value {
	fcnRcdLinearSearchThreshold()
	if v.Intv != nil {
		return v.Intv.AsValues()
	}
	return v.Domain
}

func (v *FcnRcdValue) Size() (resultInt int, err error) {
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, &err)
	if err := v.normalizeFcn(); err != nil {
		return 0, err
	}
	if v.Values == nil {
		return 0, NewNullPointerException()
	}
	return len(v.Values), nil
}

func (v *FcnRcdValue) NonNormalizedSize() int {
	fcnRcdLinearSearchThreshold()
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	return len(v.Values)
}

func (v *FcnRcdValue) ToTuple() *TupleValue {
	fcnRcdLinearSearchThreshold()
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
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	elems := make([]Value, len(v.Values))
	for i := range v.Values {
		iv, ok := fcnParameterDomain(v.Domain, i).(*IntValue)
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
	fcnRcdLinearSearchThreshold()
	if v.Domain == nil {
		return nil
	}
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	names := make([]*UniqueString, len(v.Domain))
	for i, d := range v.Domain {
		s, ok := asStringValue(d)
		if !ok {
			return nil
		}
		names[i] = s.Val
	}
	if CoverageEnabled() {
		if v.Values == nil {
			panic(NewNullPointerException())
		}
		v.CM.incValueSecondary(int64(len(v.Values)))
	}
	return NewRecordValue(names, v.Values, v.IsNorm, v.CM)
}

func (v *FcnRcdValue) Normalize() Value {
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
	if v.Domain == nil {
		panic(NewNullPointerException())
	}
	for i := 1; i < len(v.Domain); i++ {
		if isNil(v.Domain[0]) {
			panic(NewNullPointerException())
		}
		cmp, err := v.Domain[0].Compare(v.Domain[i])
		if err != nil {
			return err
		}
		if cmp == 0 {
			return v.runtimeFailure(fmt.Sprintf("The value\n%s\noccurs multiple times in the function domain.", v.Domain[i]))
		}
		if cmp > 0 {
			v.Domain[0], v.Domain[i] = v.Domain[i], v.Domain[0]
			first := fcnParameterDomain(v.Values, 0)
			other := fcnParameterDomain(v.Values, i)
			v.Values[0], v.Values[i] = other, first
		}
	}
	for i := 2; i < len(v.Domain); i++ {
		d := v.Domain[i]
		val := fcnParameterDomain(v.Values, i)
		j := i
		cmp := -1
		for j > 0 {
			if isNil(d) {
				panic(NewNullPointerException())
			}
			nextCmp, err := d.Compare(v.Domain[j-1])
			if err != nil {
				return err
			}
			cmp = nextCmp
			if cmp >= 0 {
				break
			}
			v.Domain[j] = v.Domain[j-1]
			v.Values[j] = fcnParameterDomain(v.Values, j-1)
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
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	for _, value := range v.Values {
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
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	defined := true
	if v.Intv == nil {
		for i := range v.Values {
			if defined {
				value := fcnParameterDomain(v.Domain, i)
				if isNil(value) {
					panic(NewNullPointerException())
				}
				defined = value.IsDefined()
			}
		}
	}
	for _, value := range v.Values {
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
	fcnRcdLinearSearchThreshold()
	return v.IsNorm
}

func (v *FcnRcdValue) DeepCopy() Value {
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, len(v.Values))
	for i, value := range v.Values {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		values[i] = value.DeepCopy()
	}
	if v.Intv == nil {
		if v.Domain == nil {
			panic(NewNullPointerException())
		}
		domain := make([]Value, len(v.Domain))
		copy(domain, v.Domain)
		return &FcnRcdValue{Domain: domain, Values: values}
	}
	return &FcnRcdValue{Intv: v.Intv, Values: values, IsNorm: v.IsNorm}
}

func (v *FcnRcdValue) FingerPrint(fp uint64) uint64 {
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if err := v.normalizeFcn(); err != nil {
		panic(err)
	}
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	fp = FP64ExtendByte(fp, byte(FcnRcdValueKind))
	fp = FP64ExtendInt(fp, int32(len(v.Values)))
	if v.Intv == nil {
		for i := range v.Values {
			domain := fcnParameterDomain(v.Domain, i)
			if isNil(domain) {
				panic(NewNullPointerException())
			}
			fp = domain.FingerPrint(fp)
			if isNil(v.Values[i]) {
				panic(NewNullPointerException())
			}
			fp = v.Values[i].FingerPrint(fp)
		}
		return fp
	}
	for i := range v.Values {
		fp = FP64ExtendByte(fp, byte(IntValueKind))
		index := int64(v.Intv.Low) + int64(i)
		if index > math.MaxInt32 {
			panic(NewArithmeticException("integer overflow"))
		}
		fp = FP64ExtendInt(fp, int32(index))
		if isNil(v.Values[i]) {
			panic(NewNullPointerException())
		}
		fp = v.Values[i].FingerPrint(fp)
	}
	return fp
}

func (v *FcnRcdValue) Permute(perm *MVPerm) Value {
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
	for i, value := range v.Values {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		values[i] = value.Permute(perm)
		vchanged = vchanged || values[i] != value
	}
	if v.Intv == nil {
		domain := make([]Value, size)
		dchanged := false
		for i := range domain {
			value := fcnParameterDomain(v.Domain, i)
			if isNil(value) {
				panic(NewNullPointerException())
			}
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
	fcnRcdLinearSearchThreshold()
	return ValueToString(v, "", true)
}

func (v *FcnRcdValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	fcnRcdLinearSearchThreshold()
	defer catchValueFailure(v, nil)
	if v.Values == nil {
		panic(NewNullPointerException())
	}
	length := len(v.Values)
	if length == 0 {
		sb.WriteString("<<>>")
	} else if v.isRecordLike() {
		sb.WriteString("[")
		for i := 0; i < length; i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			key, _ := asStringValue(fcnParameterDomain(v.Domain, i))
			sb.WriteString(key.Val.String() + recordArrow)
			value := fcnParameterDomain(v.Values, i)
			if value == nil {
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
			value := fcnParameterDomain(v.Values, i)
			if value == nil {
				panic(NewNullPointerException())
			}
			sb = appendValueString(value, sb, offset, swallow)
		}
		sb.WriteString(">>")
	} else {
		domain := v.DomainAsValues()
		sb.WriteString("(")
		for i := 0; i < length; i++ {
			if i > 0 {
				sb.WriteString(" @@ ")
			}
			element := fcnParameterDomain(domain, i)
			if element == nil {
				panic(NewNullPointerException())
			}
			sb = appendValueString(element, sb, offset, swallow)
			sb.WriteString(" :> ")
			// Formatting a key can normalize a set sharing the values array.
			// Read the current slot only after the key has been rendered.
			value := fcnParameterDomain(v.Values, i)
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
	if v.Domain == nil {
		panic(NewNullPointerException())
	}
	for _, dval := range v.Domain {
		sv, ok := asStringValue(dval)
		if !ok {
			return false
		}
		if sv.Val == nil {
			panic(NewNullPointerException())
		}
		if !isTLAName(sv.Val.String()) {
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
		size, err := v.Size()
		if err != nil {
			panic(err)
		}
		if size == 0 {
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
