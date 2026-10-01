package tlc

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf16"
)

const (
	assignmentSign        = " <- "
	assignmentIsMV        = " [ model value ] "
	assignmentSymmetry    = " <symmetrical> "
	tlaAnd                = "/\\"
	tlaEq                 = "="
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

var formulaNamePattern = regexp.MustCompile(`(?s)^\s*(\w+)\s*==(.*)$`)
var typedSetValidTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
var typedSetNumberOnlyPattern = regexp.MustCompile(`^[0-9]*$`)

type Formula struct {
	Text string
}

func NewFormula(text string) *Formula {
	return &Formula{Text: text}
}

func DeserializeFormulaList(serialized []string) []*Formula {
	out := make([]*Formula, 0, len(serialized))
	for _, entry := range serialized {
		if strings.HasPrefix(entry, "1") {
			out = append(out, NewFormula(entry[1:]))
		}
	}
	return out
}

func (f *Formula) GetFormula() string {
	if f == nil {
		return ""
	}
	return f.Text
}

func (f *Formula) SetFormula(text string) {
	if f != nil {
		f.Text = text
	}
}

func (f *Formula) String() string {
	return f.GetFormula()
}

func (f *Formula) IsNamed() bool {
	return f.GetLeftHandSide() != f.GetFormula()
}

func (f *Formula) GetLeftHandSide() string {
	text := f.GetFormula()
	match := formulaNamePattern.FindStringSubmatch(text)
	if len(match) == 3 {
		return strings.TrimSpace(match[1])
	}
	return text
}

func (f *Formula) GetRightHandSide() string {
	text := f.GetFormula()
	match := formulaNamePattern.FindStringSubmatch(text)
	if len(match) == 3 {
		return strings.TrimSpace(match[2])
	}
	return text
}

type Assignment struct {
	Formula
	Label            string
	Params           []string
	ModelValue       bool
	Symmetry         bool
	setOfModelValues *TypedSet
}

func NewAssignment(label string, params []string, right string) *Assignment {
	a := &Assignment{Formula: Formula{Text: right}, Label: label}
	a.SetParams(params)
	if label != "" && label == right {
		a.SetModelValue(true)
	}
	return a
}

func (a *Assignment) GetFormula(tab ...string) string {
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

func (a *Assignment) GetLeft() string {
	if a == nil {
		return ""
	}
	return a.GetParametrizedLabel(a.Label)
}

func (a *Assignment) GetParametrizedLabel(id string) string {
	if len(a.Params) == 0 {
		return id
	}
	return id + "(" + strings.Join(a.Params, ", ") + ")"
}

func (a *Assignment) GetLabel() string {
	if a == nil {
		return ""
	}
	return a.Label
}

func (a *Assignment) GetLocalLabel() string {
	label := a.GetLabel()
	idx := strings.LastIndex(label, "!")
	if idx < 0 {
		return label
	}
	return label[idx+1:]
}

func (a *Assignment) GetRight() string {
	if a == nil {
		return ""
	}
	return a.Formula.GetFormula()
}

func (a *Assignment) SetRight(right string) {
	if a == nil {
		return
	}
	a.Formula.SetFormula(right)
	a.setOfModelValues = nil
}

func (a *Assignment) GetFormattedRight() string {
	right := a.GetRight()
	if idx := strings.Index(right, "\n"); idx >= 0 {
		return right[:idx+1] + " ..."
	}
	return right
}

func (a *Assignment) SetParams(params []string) {
	if a == nil {
		return
	}
	a.Params = a.Params[:0]
	for _, param := range params {
		a.Params = append(a.Params, strings.TrimSpace(param))
	}
}

func (a *Assignment) IsModelValue() bool {
	return a != nil && a.ModelValue
}

func (a *Assignment) IsSimpleModelValue() bool {
	return a.IsModelValue() && !a.IsSetOfModelValues()
}

func (a *Assignment) IsSetOfModelValues() bool {
	return a != nil && a.ModelValue && a.GetLabel() != a.GetRight()
}

func (a *Assignment) GetSetOfModelValues() *TypedSet {
	if a == nil || !a.IsSetOfModelValues() {
		return nil
	}
	if a.setOfModelValues == nil {
		a.setOfModelValues = ParseTypedSet(a.GetRight())
	}
	return a.setOfModelValues
}

func (a *Assignment) IsSymmetricalSet() bool {
	return a != nil && a.Symmetry
}

func (a *Assignment) SetSymmetric(sym bool) {
	if a == nil {
		return
	}
	if sym && !a.ModelValue {
		panic("current assignment is not a set of model values")
	}
	a.Symmetry = sym
}

func (a *Assignment) SetModelValue(modelValue bool) {
	if a == nil {
		return
	}
	if modelValue && len(a.Params) != 0 {
		panic("operators can not be instantiated with model values")
	}
	a.ModelValue = modelValue
}

func AssignmentArrayOfEmptyStrings(number int) []string {
	out := make([]string, number)
	for i := range out {
		out[i] = ""
	}
	return out
}

func (a *Assignment) EqualSignature(other *Assignment) bool {
	if a == other {
		return true
	}
	if a == nil || other == nil {
		return false
	}
	return a.Label == other.Label && len(a.Params) == len(other.Params)
}

func (a *Assignment) PrettyPrint(delim ...string) string {
	if a == nil {
		return ""
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

type TypedSet struct {
	Values []string
	Type   string
}

func ParseTypedSet(set string) *TypedSet {
	result := &TypedSet{}
	set = strings.TrimSpace(set)
	if set == "" {
		return result
	}
	if strings.HasPrefix(set, "{") && strings.HasSuffix(set, "}") {
		set = strings.TrimSpace(set[1 : len(set)-1])
	}
	if set == "" {
		return result
	}
	parts := regexp.MustCompile(`[ \t\r\n]*,[ \t\r\n]*`).Split(set, -1)
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
		result.SetValues(regexp.MustCompile(`[ \t\r\n]*,[ \t\r\n]*`).Split(set, -1))
		return result
	}
	result.Type = typ
	result.SetValues(parts)
	return result
}

func (s *TypedSet) HasType() bool {
	return s != nil && s.Type != ""
}

func (s *TypedSet) GetType() string {
	if s == nil {
		return ""
	}
	return s.Type
}

func TypeOfTypedSetID(id string) string {
	if len(id) < 2 || id[1:2] != "_" {
		return ""
	}
	return id[:1]
}

func (s *TypedSet) SetType(typ string) {
	if s != nil {
		s.Type = typ
	}
}

func (s *TypedSet) UnsetType() {
	if s != nil {
		s.Type = ""
	}
}

func (s *TypedSet) Contains(value string) bool {
	if s == nil || value == "" {
		return false
	}
	for _, existing := range s.Values {
		if existing == value {
			return true
		}
	}
	return false
}

func (s *TypedSet) GetValues() []string {
	if s == nil {
		return nil
	}
	out := make([]string, len(s.Values))
	copy(out, s.Values)
	return out
}

func (s *TypedSet) ValuesAsList() []string {
	if s == nil {
		return nil
	}
	out := make([]string, len(s.Values))
	if !s.HasType() {
		copy(out, s.Values)
		return out
	}
	for i, value := range s.Values {
		out[i] = s.Type + "_" + value
	}
	return out
}

func (s *TypedSet) ValueCount() int {
	if s == nil {
		return 0
	}
	return len(s.Values)
}

func (s *TypedSet) Value(index int) string {
	if s == nil || index < 0 || index >= len(s.Values) {
		return ""
	}
	if s.HasType() {
		return s.Type + "_" + s.Values[index]
	}
	return s.Values[index]
}

func (s *TypedSet) SetValues(values []string) {
	if s == nil {
		return
	}
	s.Values = append([]string(nil), values...)
}

func (s *TypedSet) Equals(other *TypedSet) bool {
	if s == other {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	if s.Type != other.Type || len(s.Values) != len(other.Values) {
		return false
	}
	for i := range s.Values {
		if s.Values[i] != other.Values[i] {
			return false
		}
	}
	return true
}

func (s *TypedSet) HashCode() int32 {
	if s == nil {
		return 0
	}
	result := int32(1)
	result = 31*result + javaStringHashCode(s.Type)
	result = 31*result + javaStringArrayHashCode(s.Values)
	return result
}

func javaStringArrayHashCode(values []string) int32 {
	if values == nil {
		return 0
	}
	result := int32(1)
	for _, value := range values {
		result = 31*result + javaStringHashCode(value)
	}
	return result
}

func javaStringHashCode(value string) int32 {
	hash := int32(0)
	for _, unit := range utf16.Encode([]rune(value)) {
		hash = 31*hash + int32(unit)
	}
	return hash
}

func (s *TypedSet) String() string {
	if s == nil {
		return "{}"
	}
	return "{" + s.StringWithoutBraces() + "}"
}

func (s *TypedSet) StringWithoutBraces() string {
	if s == nil {
		return ""
	}
	return strings.Join(s.ValuesAsList(), ", ")
}

func (s *TypedSet) HasANumberOnlyValue() bool {
	if s == nil {
		return false
	}
	if s.HasType() {
		return !s.HasValidType()
	}
	for _, value := range s.Values {
		if typedSetNumberOnlyPattern.MatchString(value) {
			return true
		}
	}
	return false
}

func (s *TypedSet) HasValidType() bool {
	if s == nil || !s.HasType() {
		return true
	}
	return typedSetValidTypePattern.MatchString(s.Type)
}

type MCVariable struct {
	Name               string
	ValueAsString      string
	TLCValue           Value
	TraceExpression    string
	traceExpressionSet bool
}

func NewMCVariable(name string, value any) *MCVariable {
	v := &MCVariable{Name: name}
	switch typed := value.(type) {
	case Value:
		v.TLCValue = typed
		if typed != nil {
			v.ValueAsString = typed.String()
		}
	case string:
		v.ValueAsString = typed
	case nil:
		v.ValueAsString = ""
	default:
		v.ValueAsString = fmt.Sprint(typed)
	}
	return v
}

func (v *MCVariable) SingleLineDisplayName() string {
	if v == nil {
		return ""
	}
	name := v.Name
	if v.IsTraceExplorerExpression() {
		name = v.TraceExpression
	}
	name = strings.ReplaceAll(name, "\n", "")
	return strings.ReplaceAll(name, "\r", "")
}

func (v *MCVariable) ValueAsStringReIndentedAs(indent string) string {
	if v == nil {
		return ""
	}
	lines := regexp.MustCompile(`\r\n|\r|\n`).Split(v.ValueAsString, -1)
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

func (v *MCVariable) IsTraceExplorerExpression() bool {
	return v != nil && v.traceExpressionSet
}

func (v *MCVariable) SetTraceExpression(expr string) {
	if v != nil {
		v.TraceExpression = expr
		v.traceExpressionSet = true
	}
}

type MCState struct {
	Variables   []*MCVariable
	Name        string
	Label       string
	Location    string
	Stuttering  bool
	BackToState bool
	StateNumber int
	Record      Value
}

func NewMCState(vars []*MCVariable, name string, label string, location string, stuttering bool, backToState bool, ordinal int) *MCState {
	return &MCState{Variables: vars, Name: name, Label: label, Location: location, Stuttering: stuttering, BackToState: backToState, StateNumber: ordinal}
}

func NewMCStateMarker(other *MCState, stuttering bool, backToState bool) *MCState {
	if other == nil {
		return &MCState{Stuttering: stuttering, BackToState: backToState}
	}
	return &MCState{
		Variables:   append([]*MCVariable(nil), other.Variables...),
		Name:        other.Name,
		Label:       other.Label,
		Location:    other.Location,
		Stuttering:  stuttering,
		BackToState: backToState,
		StateNumber: other.StateNumber,
		Record:      other.Record,
	}
}

func NewMCStateFromTLCStateInfo(info *TLCStateInfo) *MCState {
	if info == nil || info.State == nil {
		return &MCState{}
	}
	state := &MCState{StateNumber: int(info.StateNumber)}
	original := info.OriginalState()
	if original == nil {
		original = info.State
	}
	values := original.Values()
	var vars []*MCVariable
	for name, value := range values.All() {
		vars = append(vars, NewMCVariable(name.String(), value))
	}
	state.Variables = vars
	state.Record = NewRecordValueFromInsMap(values)
	return state
}

func ParseMCState(input string) *MCState {
	idx := strings.Index(input, ":")
	if idx < 0 {
		return NewMCState(parseMCVariables(input), "", "", "", false, false, 0)
	}
	stateNumber := 0
	_, _ = fmt.Sscanf(strings.TrimSpace(input[:idx]), "%d", &stateNumber)
	lineEnd := strings.Index(input[idx+1:], "\n")
	if lineEnd < 0 {
		lineEnd = len(input)
	} else {
		lineEnd += idx + 1
	}
	label := input[idx+1 : lineEnd]
	stuttering := strings.HasPrefix(strings.TrimSpace(label), tlaStuttering)
	backToState := strings.HasPrefix(strings.TrimSpace(label), tlaBackToState)
	var vars []*MCVariable
	name := ""
	location := ""
	if !stuttering && !backToState {
		if lineEnd+1 <= len(input) {
			vars = parseMCVariables(input[lineEnd+1:])
		}
		trimmed := strings.TrimSpace(label)
		if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
			trimmed = strings.TrimSuffix(strings.TrimPrefix(trimmed, "<"), ">")
			if lineIdx := strings.Index(trimmed, "line "); lineIdx >= 0 {
				name = strings.TrimSpace(trimmed[:lineIdx])
				location = strings.TrimSpace(trimmed[lineIdx:])
			} else {
				name = trimmed
			}
		}
	}
	return NewMCState(vars, name, label, location, stuttering, backToState, stateNumber)
}

func (s *MCState) AsRecord(includeHeader bool) string {
	if s == nil {
		return "[]"
	}
	var b strings.Builder
	b.WriteString(tlaLeftSquareBracket)
	b.WriteString(tlaCR)
	if includeHeader {
		b.WriteString(" action")
		b.WriteString(tlaRecordArrow)
		b.WriteString(tlaLeftSquareBracket)
		b.WriteString(tlaCR)
		b.WriteString("   position")
		b.WriteString(tlaRecordArrow)
		b.WriteString(fmt.Sprint(s.StateNumber))
		b.WriteString(tlaComma)
		b.WriteString(tlaCR)
		b.WriteString("   name")
		b.WriteString(tlaRecordArrow)
		b.WriteString(fmt.Sprintf("%q", s.Name))
		b.WriteString(tlaComma)
		b.WriteString(tlaCR)
		b.WriteString("   location")
		b.WriteString(tlaRecordArrow)
		b.WriteString(fmt.Sprintf("%q", s.Location))
		b.WriteString(tlaCR)
		b.WriteString(" ]")
		if len(s.Variables) != 0 {
			b.WriteString(tlaComma)
			b.WriteString(tlaCR)
		}
	}
	for i, variable := range s.Variables {
		if variable.IsTraceExplorerExpression() {
			b.WriteString(variable.SingleLineDisplayName())
		} else {
			b.WriteString(variable.Name)
		}
		b.WriteString(tlaRecordArrow)
		b.WriteString(variable.ValueAsString)
		if i < len(s.Variables)-1 {
			b.WriteString(tlaComma)
			b.WriteString(tlaCR)
		}
	}
	b.WriteString(tlaCR)
	b.WriteString(tlaRightSquareBracket)
	return b.String()
}

func (s *MCState) AsSimpleRecord() string {
	if s == nil {
		return "[]"
	}
	parts := make([]string, 0, len(s.Variables))
	for _, variable := range s.Variables {
		parts = append(parts, variable.Name+tlaRecordArrow+variable.ValueAsString)
	}
	return tlaLeftSquareBracket + strings.Join(parts, tlaComma) + tlaRightSquareBracket
}

func (s *MCState) ConjunctiveDescription(includeTraceExpressions bool, indent string, ansiMarkup ...bool) string {
	if s == nil {
		return ""
	}
	useANSI := len(ansiMarkup) > 0 && ansiMarkup[0]
	var b strings.Builder
	for _, variable := range s.Variables {
		if variable.IsTraceExplorerExpression() && !includeTraceExpressions {
			continue
		}
		b.WriteString(indent)
		b.WriteString(tlaAnd)
		b.WriteString(" ")
		if variable.IsTraceExplorerExpression() {
			if useANSI {
				b.WriteString("\033[1m")
			}
			b.WriteString(variable.SingleLineDisplayName())
			if useANSI {
				b.WriteString("\033[0m")
			}
		} else {
			b.WriteString(variable.Name)
		}
		b.WriteString(" = ")
		b.WriteString(variable.ValueAsString)
		b.WriteString("\n")
	}
	return b.String()
}

func parseMCVariables(input string) []*MCVariable {
	lines := strings.Split(input, "\n")
	var out []*MCVariable
	var current []string
	flush := func() {
		if len(current) == 2 {
			out = append(out, NewMCVariable(strings.TrimSpace(current[0]), strings.TrimSpace(current[1])))
		}
	}
	for _, line := range lines {
		if idx := strings.Index(line, tlaAnd); idx >= 0 {
			flush()
			current = strings.SplitN(line[idx+len(tlaAnd):], tlaEq, 2)
			continue
		}
		if current != nil {
			current[1] += "\n" + line
		} else {
			current = strings.SplitN(line, tlaEq, 2)
		}
	}
	flush()
	return out
}

type MCError struct {
	Message string
	Cause   *MCError
	States  []*MCState
}

func NewMCError(message ...string) *MCError {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return &MCError{Message: msg}
}

func NewMCErrorWithCause(cause *MCError, message string) *MCError {
	return &MCError{Cause: cause, Message: message}
}

func (e *MCError) AddState(state *MCState) {
	if e != nil {
		e.States = append(e.States, state)
	}
}

func (e *MCError) UpdateStatesForTraceExpression(exprs map[string]string) {
	if e == nil {
		return
	}
	for _, state := range e.States {
		for _, variable := range state.Variables {
			if expr, ok := exprs[variable.Name]; ok {
				variable.SetTraceExpression(expr)
			}
		}
	}
}

func (e *MCError) ToSequenceOfRecords(includeHeaders bool) string {
	if e == nil {
		return tlaBeginTuple + tlaEndTuple
	}
	var b strings.Builder
	b.WriteString(tlaBeginTuple)
	b.WriteString(tlaCR)
	for i, state := range e.States {
		if state.BackToState || state.Stuttering {
			continue
		}
		if len(state.Variables) == 0 && !includeHeaders {
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
	return e != nil && len(e.States) > 0 && e.States[len(e.States)-1].BackToState
}

func (e *MCError) IsLassoWithDuplicates() bool {
	if !e.IsLasso() {
		return false
	}
	values := NewValueVec(0)
	for i := 0; i < len(e.States)-1; i++ {
		if e.States[i] == nil || e.States[i].Record == nil {
			return e.isLassoWithDuplicateSimpleRecords()
		}
		values.Add(e.States[i].Record)
	}
	size, err := NewSetEnumValueVec(values, false).Size()
	if err != nil {
		return e.isLassoWithDuplicateSimpleRecords()
	}
	return len(e.States) != size
}

func (e *MCError) isLassoWithDuplicateSimpleRecords() bool {
	seen := make(map[string]struct{}, len(e.States))
	for i := 0; i < len(e.States)-1; i++ {
		if e.States[i] == nil {
			continue
		}
		record := e.States[i].AsSimpleRecord()
		if _, ok := seen[record]; ok {
			return true
		}
		seen[record] = struct{}{}
	}
	return len(e.States) != len(seen)
}
