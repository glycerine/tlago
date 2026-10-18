package tlc

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

const (
	assignmentSign        = " <- "
	assignmentIsMV        = " [ model value ] "
	assignmentSymmetry    = " <symmetrical> "
	tlaAnd                = "/\\"
	tlaRecordArrow        = " |-> "
	tlaCR                 = "\n"
	tlaComma              = ","
	tlaBeginTuple         = "<<"
	tlaEndTuple           = ">>"
	tlaLeftSquareBracket  = "["
	tlaRightSquareBracket = "]"
	tlaBackToState        = "Back to state"
	tlaStuttering         = "Stuttering"
)

var formulaNamePattern = regexp.MustCompile(`(?s)^[ \t\n\v\f\r]*(\w+)[ \t\n\v\f\r]*==(.*)$`)
var typedSetValidTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
var typedSetNumberOnlyPattern = regexp.MustCompile(`^[0-9]*$`)

// Java's default regex \s includes vertical tab; Go's \s does not.
var typedSetSplitPattern = regexp.MustCompile(`[ \t\n\v\f\r]*,[ \t\n\v\f\r]*`)

type Formula struct {
	Text *string
}

func NewFormula(text string) *Formula {
	return NewFormulaNullable(javaString(text))
}

func NewFormulaNullable(text *string) *Formula {
	return &Formula{Text: text}
}

func DeserializeFormulaList(serialized []string) []*Formula {
	if serialized == nil {
		panic(NewNullPointerException())
	}
	out := make([]*Formula, 0, len(serialized))
	for _, entry := range serialized {
		if entry == "" {
			panic(NewStringIndexOutOfBoundsException(0, 0))
		}
		if entry[:1] == "1" {
			out = append(out, NewFormula(entry[1:]))
		}
	}
	return out
}

func (f *Formula) GetFormula() string {
	return javaNullableString(f.GetFormulaNullable())
}

func (f *Formula) GetFormulaNullable() *string {
	f.require()
	return f.Text
}

func (f *Formula) SetFormula(text string) {
	f.SetFormulaNullable(javaString(text))
}

func (f *Formula) SetFormulaNullable(text *string) {
	f.require()
	f.Text = text
}

func (f *Formula) String() string {
	return f.GetFormula()
}

func (f *Formula) IsNamed() bool {
	return f.GetLeftHandSide() != f.GetFormula()
}

func (f *Formula) GetLeftHandSide() string {
	text := f.GetFormulaNullable()
	if text == nil {
		panic(NewNullPointerException())
	}
	match := formulaNamePattern.FindStringSubmatch(*text)
	if len(match) == 3 {
		return strings.TrimFunc(match[1], func(r rune) bool { return r <= ' ' })
	}
	return *text
}

func (f *Formula) GetRightHandSide() string {
	text := f.GetFormulaNullable()
	if text == nil {
		panic(NewNullPointerException())
	}
	match := formulaNamePattern.FindStringSubmatch(*text)
	if len(match) == 3 {
		return strings.TrimFunc(match[2], func(r rune) bool { return r <= ' ' })
	}
	return *text
}

func (f *Formula) require() {
	if f == nil {
		panic(NewNullPointerException())
	}
}

type Assignment struct {
	mu sync.Mutex
	Formula
	Label            *string
	Params           []string
	nullableParams   []*string
	ModelValue       bool
	Symmetry         bool
	setOfModelValues *TypedSet
}

func NewAssignment(label string, params []string, right string) *Assignment {
	return NewAssignmentNullable(javaString(label), params, javaString(right))
}

func NewAssignmentNullable(label *string, params []string, right *string) *Assignment {
	a := &Assignment{Formula: Formula{Text: right}, Label: label}
	a.SetParams(params)
	if label != nil && right != nil && *label == *right {
		a.SetModelValue(true)
	}
	return a
}

func NewAssignmentWithNullableParams(label *string, params []*string, right *string) *Assignment {
	a := &Assignment{Formula: Formula{Text: right}, Label: label}
	a.SetParamsNullable(params)
	if label != nil && right != nil && *label == *right {
		a.SetModelValue(true)
	}
	return a
}

func (a *Assignment) GetFormula(tab ...string) string {
	a.require()
	if len(tab) > 1 {
		panic(NewIllegalArgumentException("GetFormula accepts at most one tab string"))
	}
	delim := ""
	if len(tab) > 0 {
		delim = tab[0]
	}
	var b strings.Builder
	b.WriteString(a.GetLeft())
	b.WriteString(delim)
	b.WriteString(assignmentSign)
	if a.ModelValue {
		b.WriteString(assignmentIsMV)
		if a.IsSetOfModelValues() {
			if a.IsSymmetricalSet() {
				b.WriteString(assignmentSymmetry)
			}
			b.WriteString(a.GetFormattedRight())
		}
	} else {
		b.WriteString(a.GetFormattedRight())
	}
	return b.String()
}

func (a *Assignment) GetFormulaNullableTab(tab *string) string {
	a.require()
	return a.GetFormula(javaNullableString(tab))
}

func (a *Assignment) GetFormulaNullable() *string {
	a.require()
	return javaString(a.GetFormula())
}

func (a *Assignment) GetLeft() string {
	a.require()
	return a.GetParametrizedLabelNullable(a.Label)
}

func (a *Assignment) GetParametrizedLabel(id string) string {
	return a.GetParametrizedLabelNullable(javaString(id))
}

func (a *Assignment) GetParametrizedLabelNullable(id *string) string {
	a.require()
	params := a.parameterStrings()
	if len(params) == 0 {
		return javaNullableString(id)
	}
	return javaNullableString(id) + "(" + strings.Join(params, ", ") + ")"
}

func (a *Assignment) GetLabel() string {
	return javaNullableString(a.GetLabelNullable())
}

func (a *Assignment) GetLabelNullable() *string {
	a.require()
	return a.Label
}

func (a *Assignment) GetParams() []string {
	a.require()
	if a.nullableParams != nil {
		return a.parameterStrings()
	}
	return a.Params
}

func (a *Assignment) GetParamsNullable() []*string {
	a.require()
	if a.nullableParams != nil {
		return a.nullableParams
	}
	out := make([]*string, len(a.Params))
	for i := range a.Params {
		out[i] = javaString(a.Params[i])
	}
	return out
}

func (a *Assignment) GetLocalLabel() string {
	label := a.GetLabelNullable()
	if label == nil {
		panic(NewNullPointerException())
	}
	idx := strings.LastIndex(*label, "!")
	if idx < 0 {
		return *label
	}
	return (*label)[idx+1:]
}

func (a *Assignment) GetRight() string {
	return javaNullableString(a.GetRightNullable())
}

func (a *Assignment) GetRightNullable() *string {
	a.require()
	return a.Formula.GetFormulaNullable()
}

func (a *Assignment) SetRight(right string) {
	a.SetRightNullable(javaString(right))
}

func (a *Assignment) SetRightNullable(right *string) {
	a.require()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Formula.SetFormulaNullable(right)
	a.setOfModelValues = nil
}

func (a *Assignment) SetFormula(string) {
	a.require()
	panic(NewUnsupportedOperationException("Not implemented yet"))
}

func (a *Assignment) SetFormulaNullable(*string) {
	a.require()
	panic(NewUnsupportedOperationException("Not implemented yet"))
}

func (a *Assignment) String() string {
	a.require()
	return a.GetFormula()
}

func (a *Assignment) GetFormattedRight() string {
	return javaNullableString(a.GetFormattedRightNullable())
}

func (a *Assignment) GetFormattedRightNullable() *string {
	a.require()
	right := a.GetRightNullable()
	if right == nil {
		return nil
	}
	if idx := strings.Index(*right, "\n"); idx >= 0 {
		return javaString((*right)[:idx+1] + " ...")
	}
	return right
}

func (a *Assignment) IsNamed() bool {
	a.require()
	return a.GetLeftHandSide() != a.GetFormula()
}

func (a *Assignment) GetLeftHandSide() string {
	a.require()
	text := a.GetRightNullable()
	if text == nil {
		panic(NewNullPointerException())
	}
	match := formulaNamePattern.FindStringSubmatch(*text)
	if len(match) == 3 {
		return strings.TrimFunc(match[1], func(r rune) bool { return r <= ' ' })
	}
	return a.GetFormula()
}

func (a *Assignment) GetRightHandSide() string {
	a.require()
	text := a.GetRightNullable()
	if text == nil {
		panic(NewNullPointerException())
	}
	match := formulaNamePattern.FindStringSubmatch(*text)
	if len(match) == 3 {
		return strings.TrimFunc(match[2], func(r rune) bool { return r <= ' ' })
	}
	return a.GetFormula()
}

func (a *Assignment) SetParams(params []string) {
	a.require()
	a.nullableParams = nil
	if params != nil {
		a.Params = params
		for i := range a.Params {
			a.Params[i] = strings.TrimFunc(a.Params[i], func(r rune) bool { return r <= ' ' })
		}
	} else {
		a.Params = []string{}
	}
}

func (a *Assignment) SetParamsNullable(params []*string) {
	a.require()
	a.Params = nil
	if params == nil {
		a.nullableParams = make([]*string, 0)
		return
	}
	a.nullableParams = params
	for i, param := range a.nullableParams {
		if param == nil {
			panic(NewNullPointerException())
		}
		trimmed := strings.TrimFunc(*param, func(r rune) bool { return r <= ' ' })
		if trimmed != *param {
			a.nullableParams[i] = javaString(trimmed)
		}
	}
}

func (a *Assignment) parameterStrings() []string {
	if a.nullableParams == nil {
		return a.Params
	}
	out := make([]string, len(a.nullableParams))
	for i, param := range a.nullableParams {
		out[i] = javaNullableString(param)
	}
	return out
}

func (a *Assignment) parameterCount() int {
	if a.nullableParams != nil {
		return len(a.nullableParams)
	}
	return len(a.Params)
}

func (a *Assignment) IsModelValue() bool {
	a.require()
	return a.ModelValue
}

func (a *Assignment) IsSimpleModelValue() bool {
	return a.IsModelValue() && !a.IsSetOfModelValues()
}

func (a *Assignment) IsSetOfModelValues() bool {
	a.require()
	if !a.ModelValue {
		return false
	}
	label := a.GetLabelNullable()
	if label == nil {
		panic(NewNullPointerException())
	}
	right := a.GetRightNullable()
	return right == nil || *label != *right
}

func (a *Assignment) GetSetOfModelValues() *TypedSet {
	a.require()
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.IsSetOfModelValues() {
		return nil
	}
	if a.setOfModelValues == nil {
		a.setOfModelValues = ParseTypedSetNullable(a.GetRightNullable())
	}
	return a.setOfModelValues
}

func (a *Assignment) IsSymmetricalSet() bool {
	a.require()
	return a.Symmetry
}

func (a *Assignment) SetSymmetric(sym bool) {
	a.require()
	if sym && !a.ModelValue {
		panic(NewIllegalArgumentException("Current assignment is not a set of model values"))
	}
	a.Symmetry = sym
}

func (a *Assignment) SetModelValue(modelValue bool) {
	a.require()
	if modelValue && a.parameterCount() != 0 {
		panic(NewIllegalArgumentException("Operators can not be instantiated with model values"))
	}
	a.ModelValue = modelValue
}

func AssignmentArrayOfEmptyStrings(number int) []string {
	number = int(int32(number))
	if number < 0 {
		panic(NewNegativeArraySizeException(fmtInt(number)))
	}
	out := make([]string, number)
	for i := range out {
		out[i] = ""
	}
	return out
}

func (a *Assignment) EqualSignature(other *Assignment) bool {
	a.require()
	if a == other {
		return true
	}
	if other == nil {
		return false
	}
	if (a.Label == nil) != (other.Label == nil) {
		return false
	}
	return (a.Label == nil || *a.Label == *other.Label) && a.parameterCount() == other.parameterCount()
}

func (a *Assignment) PrettyPrint(delim ...string) string {
	a.require()
	if len(delim) > 1 {
		panic(NewIllegalArgumentException("PrettyPrint accepts at most one delimiter"))
	}
	sep := ""
	if len(delim) > 0 {
		sep = delim[0]
	}
	if !a.IsModelValue() {
		return a.GetFormula(sep)
	}
	if !a.IsSetOfModelValues() {
		return a.GetLeft()
	}
	var b strings.Builder
	b.WriteString(a.GetLeft())
	b.WriteString(sep)
	b.WriteString(assignmentSign)
	if a.IsSymmetricalSet() {
		b.WriteString("s")
	}
	b.WriteString(a.GetFormattedRight())
	return b.String()
}

func (a *Assignment) PrettyPrintNullable(delim *string) string {
	a.require()
	return a.PrettyPrint(javaNullableString(delim))
}

func (a *Assignment) require() {
	if a == nil {
		panic(NewNullPointerException())
	}
}

type TypedSet struct {
	Values  []*string
	Type    string
	typeSet bool
}

func NewTypedSet() *TypedSet {
	return &TypedSet{Values: []*string{}}
}

func ParseTypedSet(set string) *TypedSet {
	return ParseTypedSetNullable(&set)
}

// ParseTypedSetNullable preserves TypedSet.parseSet's nullable String input.
func ParseTypedSetNullable(input *string) *TypedSet {
	result := NewTypedSet()
	if input == nil {
		return result
	}
	// Java String.trim removes only code units at or below U+0020.
	trim := func(r rune) bool { return r <= ' ' }
	set := strings.TrimFunc(*input, trim)
	if set == "" {
		return result
	}
	if strings.HasPrefix(set, "{") && strings.HasSuffix(set, "}") {
		set = strings.TrimFunc(set[1:len(set)-1], trim)
	}
	if set == "" {
		return result
	}
	parts := typedSetSplit(set)
	if len(parts) == 0 {
		return result
	}
	typeSep := strings.Index(parts[0], "_")
	if typeSep <= 0 {
		result.SetValues(parts)
		return result
	}
	typ := parts[0][:typeSep]
	parts[0] = parts[0][typeSep+1:]
	violated := parts[0] == ""
	for i := 1; i < len(parts); i++ {
		prefix := typ + "_"
		if strings.HasPrefix(parts[i], prefix) {
			parts[i] = parts[i][typeSep+1:]
			if parts[i] == "" {
				violated = true
			}
		} else {
			violated = true
		}
		if violated {
			break
		}
	}
	if violated {
		result.SetValues(typedSetSplit(set))
		return result
	}
	result.SetType(typ)
	result.SetValues(parts)
	return result
}

func typedSetSplit(set string) []string {
	parts := typedSetSplitPattern.Split(set, -1)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func (s *TypedSet) HasType() bool {
	s.require()
	return s.typeSet || s.Type != ""
}

func (s *TypedSet) GetType() string {
	typ := s.GetTypeNullable()
	if typ == nil {
		return ""
	}
	return *typ
}

// GetTypeNullable preserves TypedSet.getType's nullable result.
func (s *TypedSet) GetTypeNullable() *string {
	s.require()
	if !s.HasType() {
		return nil
	}
	return javaString(s.Type)
}

func TypedSetGetTypeOfID(id *string) *string {
	if id == nil {
		return nil
	}
	units := javaStringUTF16(*id)
	if len(units) < 2 || units[1] != '_' {
		return nil
	}
	result := javaStringFromUTF16(units[:1])
	return &result
}

func TypeOfTypedSetID(id string) string {
	result := TypedSetGetTypeOfID(&id)
	if result == nil {
		return ""
	}
	return *result
}

func (s *TypedSet) SetType(typ string) {
	s.SetTypeNullable(javaString(typ))
}

// SetTypeNullable preserves TypedSet.setType's nullable argument.
func (s *TypedSet) SetTypeNullable(typ *string) {
	s.require()
	if typ == nil {
		s.Type = ""
		s.typeSet = false
		return
	}
	s.Type = *typ
	s.typeSet = true
}

func (s *TypedSet) UnsetType() {
	s.SetTypeNullable(nil)
}

func (s *TypedSet) Contains(value string) bool {
	return s.ContainsNullable(javaString(value))
}

// ContainsNullable preserves TypedSet.contains's nullable argument.
func (s *TypedSet) ContainsNullable(value *string) bool {
	s.require()
	if value == nil {
		return false
	}
	for _, existing := range s.Values {
		if existing != nil && *existing == *value {
			return true
		}
	}
	return false
}

func (s *TypedSet) GetValues() []string {
	values := s.GetValuesNullable()
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = javaNullableString(value)
	}
	return out
}

func (s *TypedSet) GetValuesNullable() []*string {
	s.require()
	return s.Values
}

func (s *TypedSet) GetValuesAsList() []string {
	values := s.GetValuesAsListNullable()
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = javaNullableString(value)
	}
	return out
}

func (s *TypedSet) GetValuesAsListNullable() []*string {
	s.require()
	if !s.HasType() {
		return s.Values
	}
	out := make([]*string, len(s.Values))
	for i, value := range s.Values {
		out[i] = javaString(s.Type + "_" + javaNullableString(value))
	}
	return out
}

func (s *TypedSet) ValuesAsList() []string {
	return s.GetValuesAsList()
}

func (s *TypedSet) GetValueCount() int {
	s.require()
	return len(s.Values)
}

func (s *TypedSet) ValueCount() int {
	return s.GetValueCount()
}

func (s *TypedSet) GetValue(index int) string {
	value := s.GetValueNullable(index)
	if value == nil {
		return ""
	}
	return *value
}

// GetValueNullable preserves TypedSet.getValue's nullable out-of-range result.
func (s *TypedSet) GetValueNullable(index int) *string {
	s.require()
	index = int(int32(index))
	if index >= len(s.Values) {
		return nil
	}
	if index < 0 {
		panic(NewArrayIndexOutOfBoundsException(index, len(s.Values)))
	}
	if s.HasType() {
		return javaString(s.Type + "_" + javaNullableString(s.Values[index]))
	}
	return javaString(javaNullableString(s.Values[index]))
}

func (s *TypedSet) Value(index int) string {
	return s.GetValue(index)
}

func (s *TypedSet) SetValues(values []string) {
	nullable := make([]*string, len(values))
	for i := range values {
		nullable[i] = javaString(values[i])
	}
	s.SetValuesNullable(nullable)
}

func (s *TypedSet) SetValuesNullable(values []*string) {
	s.require()
	if values == nil {
		s.Values = []*string{}
		return
	}
	s.Values = values
}

func (s *TypedSet) Equals(other *TypedSet) bool {
	s.require()
	if s == other {
		return true
	}
	if other == nil {
		return false
	}
	if s.HasType() != other.HasType() || s.Type != other.Type || len(s.Values) != len(other.Values) {
		return false
	}
	for i := range s.Values {
		if (s.Values[i] == nil) != (other.Values[i] == nil) || s.Values[i] != nil && *s.Values[i] != *other.Values[i] {
			return false
		}
	}
	return true
}

func (s *TypedSet) HashCode() int32 {
	s.require()
	result := int32(1)
	result = 31*result + javaStringHashCode(s.Type)
	result = 31*result + javaStringArrayHashCode(s.Values)
	return result
}

func javaStringArrayHashCode(values []*string) int32 {
	result := int32(1)
	for _, value := range values {
		if value != nil {
			result = 31*result + javaStringHashCode(*value)
		} else {
			result *= 31
		}
	}
	return result
}

func javaStringHashCode(value string) int32 {
	return javaFormatStringHash(value)
}

func (s *TypedSet) String() string {
	s.require()
	return "{" + s.ToStringWithoutBraces() + "}"
}

func (s *TypedSet) ToStringWithoutBraces() string {
	s.require()
	return strings.Join(s.GetValuesAsList(), ", ")
}

func (s *TypedSet) StringWithoutBraces() string {
	return s.ToStringWithoutBraces()
}

func (s *TypedSet) HasANumberOnlyValue() bool {
	s.require()
	if s.HasType() {
		return !s.HasValidType()
	}
	for _, value := range s.Values {
		if value == nil {
			panic(NewNullPointerException())
		}
		if typedSetNumberOnlyPattern.MatchString(*value) {
			return true
		}
	}
	return false
}

func (s *TypedSet) HasValidType() bool {
	s.require()
	if !s.HasType() {
		return true
	}
	return typedSetValidTypePattern.MatchString(s.Type)
}

func (s *TypedSet) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

type MCVariable struct {
	Name            string
	ValueAsString   string
	TLCValue        Value
	TraceExpression *string

	ownersCaptured bool
	nameOwner      string
	valueOwner     string
	tlcValueOwner  Value
}

func NewMCVariable(name string, value any) *MCVariable {
	switch typed := value.(type) {
	case Value:
		return NewMCVariableFromValueNullable(javaString(name), typed)
	case string:
		return NewMCVariableFromStringNullable(javaString(name), javaString(typed))
	case nil:
		return NewMCVariableFromValueNullable(javaString(name), nil)
	default:
		return NewMCVariableFromStringNullable(javaString(name), javaString(fmt.Sprint(typed)))
	}

}

func NewMCVariableFromStringNullable(name *string, value *string) *MCVariable {
	if name == nil || value == nil {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	return &MCVariable{
		Name: *name, ValueAsString: *value,
		ownersCaptured: true, nameOwner: *name, valueOwner: *value,
	}
}

func NewMCVariableFromValueNullable(name *string, value Value) *MCVariable {
	if name == nil {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	v := &MCVariable{Name: *name, TLCValue: value, ownersCaptured: true, nameOwner: *name, tlcValueOwner: value}
	if !isNil(value) {
		v.ValueAsString = value.String()
		v.valueOwner = v.ValueAsString
	}
	return v
}

func (v *MCVariable) ownedName() string {
	if v.ownersCaptured {
		return v.nameOwner
	}
	return v.Name
}

func (v *MCVariable) ownedValueAsString() string {
	if v.ownersCaptured {
		return v.valueOwner
	}
	return v.ValueAsString
}

func (v *MCVariable) ownedTLCValue() Value {
	if v.ownersCaptured {
		return v.tlcValueOwner
	}
	return v.TLCValue
}

func (v *MCVariable) GetName() string {
	v.require()
	return v.ownedName()
}

func (v *MCVariable) GetTLCValue() Value {
	v.require()
	return v.ownedTLCValue()
}

func (v *MCVariable) GetSingleLineDisplayName() string {
	v.require()
	name := v.ownedName()
	if v.IsTraceExplorerExpression() {
		name = *v.TraceExpression
	}
	name = strings.ReplaceAll(name, "\n", "")
	return strings.ReplaceAll(name, "\r", "")
}

func (v *MCVariable) SingleLineDisplayName() string {
	return v.GetSingleLineDisplayName()
}

func (v *MCVariable) GetValueAsStringReIndentedAs(indent string) string {
	return v.GetValueAsStringReIndentedAsNullable(javaString(indent))
}

func (v *MCVariable) GetValueAsStringReIndentedAsNullable(indent *string) string {
	v.require()
	lines := javaSplitLines(v.ownedValueAsString())
	for i, line := range lines {
		lines[i] = javaNullableString(indent) + line
	}
	return strings.Join(lines, "\n")
}

func (v *MCVariable) ValueAsStringReIndentedAs(indent string) string {
	return v.GetValueAsStringReIndentedAs(indent)
}

func javaSplitLines(text string) []string {
	lines := regexp.MustCompile(`\r\n|\r|\n`).Split(text, -1)
	if text == "" {
		return lines
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func (v *MCVariable) GetValueAsString() string {
	v.require()
	return v.ownedValueAsString()
}

func (v *MCVariable) IsTraceExplorerExpression() bool {
	v.require()
	return v.TraceExpression != nil
}

func (v *MCVariable) SetTraceExpression(expr *string) {
	v.require()
	v.TraceExpression = expr
}

func (v *MCVariable) GetTraceExpression() *string {
	v.require()
	return v.TraceExpression
}

func (v *MCVariable) require() {
	if v == nil {
		panic(NewNullPointerException())
	}
}

type MCState struct {
	Variables   []*MCVariable
	Name        *string
	Label       *string
	Location    *SourceLocation
	Stuttering  bool
	BackToState bool
	StateNumber int
	Record      Value

	ownersCaptured   bool
	variablesOwner   []*MCVariable
	nameOwner        *string
	labelOwner       *string
	locationOwner    *SourceLocation
	stutteringOwner  bool
	backOwner        bool
	stateNumberOwner int
}

func NewMCState(vars []*MCVariable, name string, label string, location *SourceLocation, stuttering bool, backToState bool, ordinal int) *MCState {
	return NewMCStateNullable(vars, javaString(name), javaString(label), location, stuttering, backToState, ordinal)
}

func NewMCStateNullable(vars []*MCVariable, name *string, label *string, location *SourceLocation, stuttering bool, backToState bool, ordinal int) *MCState {
	return newMCState(vars, name, label, location, stuttering, backToState, ordinal)
}

func newMCState(vars []*MCVariable, name *string, label *string, location *SourceLocation, stuttering bool, backToState bool, ordinal int) *MCState {
	ordinal = int(int32(ordinal))
	return &MCState{
		Variables: vars, Name: name, Label: label, Location: location,
		Stuttering: stuttering, BackToState: backToState, StateNumber: ordinal,
		ownersCaptured: true, variablesOwner: vars, nameOwner: name, labelOwner: label,
		locationOwner: location, stutteringOwner: stuttering, backOwner: backToState, stateNumberOwner: ordinal,
	}
}

func NewMCStateMarker(other *MCState, stuttering bool, backToState bool) *MCState {
	if other == nil {
		panic(NewNullPointerException())
	}
	return newMCState(other.GetVariables(), other.GetNameNullable(), other.GetLabelNullable(), other.GetLocation(), stuttering, backToState, other.GetStateNumber())
}

func NewMCStateFromTLCStateInfo(info *TLCStateInfo) *MCState {
	if info == nil {
		panic(NewNullPointerException())
	}
	original := info.GetOriginalState()
	values := original.GetVals()
	var vars []*MCVariable
	for name, value := range values.All() {
		if name == nil {
			panic(NewNullPointerException())
		}
		vars = append(vars, NewMCVariable(name.String(), value))
	}
	state := newMCState(vars, javaString(""), javaString(""), nil, false, false, int(int32(info.StateNumber)))
	state.Record = NewRecordValueFromState(info.GetOriginalState())
	return state
}

func (s *MCState) ownedVariables() []*MCVariable {
	if s.ownersCaptured {
		return s.variablesOwner
	}
	return s.Variables
}

func (s *MCState) ownedName() *string {
	if s.ownersCaptured {
		return s.nameOwner
	}
	return s.Name
}

func (s *MCState) ownedLabel() *string {
	if s.ownersCaptured {
		return s.labelOwner
	}
	return s.Label
}

func (s *MCState) ownedLocation() *SourceLocation {
	if s.ownersCaptured {
		return s.locationOwner
	}
	return s.Location
}

func (s *MCState) ownedStuttering() bool {
	if s.ownersCaptured {
		return s.stutteringOwner
	}
	return s.Stuttering
}

func (s *MCState) ownedBackToState() bool {
	if s.ownersCaptured {
		return s.backOwner
	}
	return s.BackToState
}

func (s *MCState) ownedStateNumber() int {
	if s.ownersCaptured {
		return s.stateNumberOwner
	}
	return s.StateNumber
}

func ParseMCState(input string) *MCState {
	idx := strings.Index(input, ":")
	if idx < 0 {
		panic(NewStringIndexOutOfBoundsException(idx, len(input)))
	}
	stateNumberText := strings.TrimFunc(input[:idx], func(r rune) bool { return r <= ' ' })
	parsedStateNumber, ok := javaParseDecimalInt(stateNumberText)
	if !ok {
		panic(NewNumberFormatException(stateNumberText))
	}
	stateNumber := int(parsedStateNumber)
	lineEnd := strings.Index(input[idx+1:], "\n")
	if lineEnd < 0 {
		lineEnd = len(input)
	} else {
		lineEnd += idx + 1
	}
	label := input[idx+1 : lineEnd]
	stuttering := strings.HasPrefix(label, " "+tlaStuttering)
	backToState := strings.HasPrefix(label, " "+tlaBackToState)
	var vars []*MCVariable
	var name *string
	var location *SourceLocation
	if !stuttering && !backToState {
		if lineEnd >= len(input) {
			panic(NewStringIndexOutOfBoundsException(lineEnd+1, len(input)))
		}
		vars = parseMCVariables(input[lineEnd+1:])
		if len(label) > 2 {
			sublabel := label[2 : len(label)-1]
			if lineIdx := strings.Index(sublabel, "line "); lineIdx >= 0 {
				if lineIdx == 0 {
					panic(NewStringIndexOutOfBoundsException(-1, len(sublabel)))
				}
				name = javaString(sublabel[:lineIdx-1])
				parsed := parseMCStateLocation(sublabel[lineIdx:])
				location = &parsed
			} else {
				name = javaString(sublabel)
			}
		}
	}
	return NewMCStateNullable(vars, name, javaString(label), location, stuttering, backToState, stateNumber)
}

func (s *MCState) GetVariables() []*MCVariable {
	s.require()
	return s.ownedVariables()
}

func (s *MCState) GetLabel() string {
	return javaNullableString(s.GetLabelNullable())
}

func (s *MCState) GetLabelNullable() *string {
	s.require()
	return s.ownedLabel()
}

func (s *MCState) GetName() string {
	return javaNullableString(s.GetNameNullable())
}

func (s *MCState) GetNameNullable() *string {
	s.require()
	return s.ownedName()
}

func (s *MCState) IsStuttering() bool {
	s.require()
	return s.ownedStuttering()
}

func (s *MCState) IsBackToState() bool {
	s.require()
	return s.ownedBackToState()
}

func (s *MCState) GetStateNumber() int {
	s.require()
	return s.ownedStateNumber()
}

func (s *MCState) GetLocation() *SourceLocation {
	s.require()
	return s.ownedLocation()
}

func (s *MCState) GetRecord() Value {
	s.require()
	return s.Record
}

func (s *MCState) AsRecord(includeHeader bool) string {
	s.require()
	variables := s.ownedVariables()
	if variables == nil {
		panic(NewNullPointerException())
	}
	var b strings.Builder
	b.WriteString(tlaLeftSquareBracket)
	b.WriteString(tlaCR)
	if includeHeader {
		b.WriteString(" _TEAction")
		b.WriteString(tlaRecordArrow)
		b.WriteString(tlaLeftSquareBracket)
		b.WriteString(tlaCR)
		b.WriteString("   position")
		b.WriteString(tlaRecordArrow)
		b.WriteString(fmt.Sprint(s.ownedStateNumber()))
		b.WriteString(tlaComma)
		b.WriteString(tlaCR)
		b.WriteString("   name")
		b.WriteString(tlaRecordArrow)
		b.WriteString("\"")
		b.WriteString(javaNullableString(s.ownedName()))
		b.WriteString("\"")
		b.WriteString(tlaComma)
		b.WriteString(tlaCR)
		b.WriteString("   location")
		b.WriteString(tlaRecordArrow)
		b.WriteString("\"")
		location := s.ownedLocation()
		if location == nil {
			b.WriteString("null")
		} else {
			b.WriteString(location.String())
		}
		b.WriteString("\"")
		b.WriteString(tlaCR)
		b.WriteString(" ]")
		if len(variables) != 0 {
			b.WriteString(tlaComma)
			b.WriteString(tlaCR)
		}
	}
	for i, variable := range variables {
		if variable.IsTraceExplorerExpression() {
			b.WriteString(variable.GetSingleLineDisplayName())
		} else {
			b.WriteString(variable.GetName())
		}
		b.WriteString(tlaRecordArrow)
		b.WriteString(variable.GetValueAsString())
		if i < len(variables)-1 {
			b.WriteString(tlaComma)
			b.WriteString(tlaCR)
		}
	}
	b.WriteString(tlaCR)
	b.WriteString(tlaRightSquareBracket)
	return b.String()
}

func (s *MCState) AsSimpleRecord() string {
	s.require()
	variables := s.ownedVariables()
	if variables == nil {
		panic(NewNullPointerException())
	}
	parts := make([]string, 0, len(variables))
	for _, variable := range variables {
		parts = append(parts, variable.GetName()+tlaRecordArrow+variable.GetValueAsString())
	}
	return tlaLeftSquareBracket + strings.Join(parts, tlaComma) + tlaRightSquareBracket
}

func (s *MCState) GetConjunctiveDescription(includeTraceExpressions bool, indent string, ansiMarkup ...bool) string {
	return s.GetConjunctiveDescriptionNullable(includeTraceExpressions, javaString(indent), ansiMarkup...)
}

func (s *MCState) GetConjunctiveDescriptionNullable(includeTraceExpressions bool, indent *string, ansiMarkup ...bool) string {
	s.require()
	if len(ansiMarkup) > 1 {
		panic(NewIllegalArgumentException("GetConjunctiveDescription accepts at most one ANSI flag"))
	}
	useANSI := len(ansiMarkup) > 0 && ansiMarkup[0]
	variables := s.ownedVariables()
	if variables == nil {
		panic(NewNullPointerException())
	}
	var b strings.Builder
	for _, variable := range variables {
		if variable.IsTraceExplorerExpression() && !includeTraceExpressions {
			continue
		}
		if indent != nil {
			b.WriteString(*indent)
		}
		b.WriteString(tlaAnd)
		b.WriteString(" ")
		if variable.IsTraceExplorerExpression() {
			if useANSI {
				b.WriteString("\033[1m")
			}
			b.WriteString(variable.GetSingleLineDisplayName())
			if useANSI {
				b.WriteString("\033[0m")
			}
		} else {
			b.WriteString(variable.GetName())
		}
		b.WriteString(" = ")
		b.WriteString(variable.GetValueAsString())
		b.WriteString("\n")
	}
	return b.String()
}

func (s *MCState) ConjunctiveDescription(includeTraceExpressions bool, indent string, ansiMarkup ...bool) string {
	return s.GetConjunctiveDescription(includeTraceExpressions, indent, ansiMarkup...)
}

var mcStateLocationPattern = regexp.MustCompile(`^line ([0-9]+), col ([0-9]+) to line ([0-9]+), col ([0-9]+) of module ([A-Za-z_0-9]+)$`)

func parseMCStateLocation(value string) SourceLocation {
	match := mcStateLocationPattern.FindStringSubmatch(value)
	if match == nil {
		return NullSourceLocation
	}
	coordinates := [4]int{}
	for i := range coordinates {
		parsed, ok := javaParseDecimalInt(match[i+1])
		if !ok {
			return NullSourceLocation
		}
		coordinates[i] = int(parsed)
	}
	return NewSourceLocation(match[5], coordinates[0], coordinates[1], coordinates[2], coordinates[3])
}

func (s *MCState) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func parseMCVariables(input string) []*MCVariable {
	lines := javaSplitLiteral(input, "\n")
	var out []*MCVariable
	var current []string
	flush := func() {
		if current == nil {
			return
		}
		if len(current) < 2 {
			panic(NewArrayIndexOutOfBoundsException(1, len(current)))
		}
		out = append(out, NewMCVariable(current[0], current[1]))
	}
	for _, line := range lines {
		if idx := strings.Index(line, tlaAnd); idx >= 0 {
			flush()
			start := idx + len(tlaAnd) + 1
			if start > len(line) {
				panic(NewStringIndexOutOfBoundsException(start, len(line)))
			}
			current = javaSplitLiteral(line[start:], tlaEqSp)
			continue
		}
		if current != nil {
			if len(current) < 2 {
				panic(NewArrayIndexOutOfBoundsException(1, len(current)))
			}
			current[1] += "\n" + line
		} else {
			current = javaSplitLiteral(line, tlaEqSp)
		}
	}
	flush()
	return out
}

func javaSplitLiteral(text, separator string) []string {
	parts := strings.Split(text, separator)
	if text == "" {
		return parts
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

type MCError struct {
	Message *string
	Cause   *MCError
	States  []*MCState

	ownersCaptured bool
	messageOwner   *string
	causeOwner     *MCError
	statesOwner    []*MCState
}

func NewMCError(message ...string) *MCError {
	if len(message) > 1 {
		panic(NewIllegalArgumentException("NewMCError accepts at most one message"))
	}
	var msg *string
	if len(message) > 0 {
		msg = javaString(message[0])
	}
	return newMCError(nil, msg)
}

func NewMCErrorWithCause(cause *MCError, message *string) *MCError {
	return newMCError(cause, message)
}

func newMCError(cause *MCError, message *string) *MCError {
	states := make([]*MCState, 0)
	return &MCError{
		Cause: cause, Message: message, States: states,
		ownersCaptured: true, causeOwner: cause, messageOwner: message, statesOwner: states,
	}
}

func (e *MCError) ownedMessage() *string {
	if e.ownersCaptured {
		return e.messageOwner
	}
	return e.Message
}

func (e *MCError) ownedCause() *MCError {
	if e.ownersCaptured {
		return e.causeOwner
	}
	return e.Cause
}

func (e *MCError) ownedStates() []*MCState {
	if e.ownersCaptured {
		return e.statesOwner
	}
	return e.States
}

func (e *MCError) GetStates() []*MCState {
	e.require()
	return e.ownedStates()
}

func (e *MCError) GetMessage() *string {
	e.require()
	return e.ownedMessage()
}

func (e *MCError) GetCause() *MCError {
	e.require()
	return e.ownedCause()
}

func (e *MCError) AddState(state *MCState) {
	e.require()
	if e.ownersCaptured {
		e.statesOwner = append(e.statesOwner, state)
		e.States = e.statesOwner
		return
	}
	e.States = append(e.States, state)
}

func (e *MCError) UpdateStatesForTraceExpression(exprs map[string]string) {
	if exprs == nil {
		e.UpdateStatesForTraceExpressionNullable(nil)
		return
	}
	nullable := make(map[string]*string, len(exprs))
	for name, expression := range exprs {
		nullable[name] = javaString(expression)
	}
	e.UpdateStatesForTraceExpressionNullable(nullable)
}

func (e *MCError) UpdateStatesForTraceExpressionNullable(exprs map[string]*string) {
	e.require()
	for _, state := range e.ownedStates() {
		variables := state.GetVariables()
		if variables == nil {
			panic(NewNullPointerException())
		}
		for _, variable := range variables {
			name := variable.GetName()
			if exprs == nil {
				panic(NewNullPointerException())
			}
			if expression := exprs[name]; expression != nil {
				variable.SetTraceExpression(expression)
			}
		}
	}
}

func (e *MCError) ToSequenceOfRecords(includeHeaders bool) string {
	e.require()
	var b strings.Builder
	b.WriteString(tlaBeginTuple)
	b.WriteString(tlaCR)
	states := e.ownedStates()
	for i, state := range states {
		if state.IsBackToState() || state.IsStuttering() {
			continue
		}
		variables := state.GetVariables()
		if variables == nil {
			panic(NewNullPointerException())
		}
		if len(variables) == 0 && !includeHeaders {
			break
		}
		if i > 0 {
			b.WriteString(tlaComma)
			b.WriteString(tlaCR)
		}
		b.WriteString(state.AsRecord(includeHeaders))
	}
	b.WriteString(tlaCR)
	b.WriteString(tlaEndTuple)
	return b.String()
}

func (e *MCError) IsLasso() bool {
	e.require()
	states := e.ownedStates()
	return len(states) > 0 && states[len(states)-1].IsBackToState()
}

func (e *MCError) IsLassoWithDuplicates() bool {
	if !e.IsLasso() {
		return false
	}
	values := NewValueVec(0)
	states := e.ownedStates()
	for i := 0; i < len(states)-1; i++ {
		record := states[i].GetRecord()
		if record == nil {
			panic(NewNullPointerException())
		}
		values.Add(record)
	}
	size, err := NewSetEnumValueVec(values, false).Size()
	if err != nil {
		panic(err)
	}
	return len(states) != size
}

func (e *MCError) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}
