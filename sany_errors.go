package tlago

import (
	"fmt"
	"reflect"

	"github.com/glycerine/tlago/tlc"
)

// SanyErrorDetails adapts Errors.ErrorDetails. Object-parameter logging retains
// the format separately from its arguments, so text without interpolation does
// not scan percent signs.
type SanyErrorDetails struct {
	code       SanyErrorCode
	location   *tlc.SourceLocation
	format     string
	parameters []any
}

func (d *SanyErrorDetails) GetCode() SanyErrorCode {
	if d == nil {
		panic(tlc.NewNullPointerException())
	}
	return d.code
}

func (d *SanyErrorDetails) GetLocation() *tlc.SourceLocation {
	if d == nil {
		panic(tlc.NewNullPointerException())
	}
	return d.location
}

func (d *SanyErrorDetails) GetMessage() string {
	if d == nil {
		panic(tlc.NewNullPointerException())
	}
	if len(d.parameters) == 0 {
		return d.format
	}
	message, err := tlc.JavaFormatObjects(d.format, d.parameters...)
	if err != nil {
		panic(err)
	}
	return message
}

func (d *SanyErrorDetails) GetParameters() []any {
	if d == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, parameter := range d.parameters {
		if sanyErrorParameterIsNull(parameter) {
			panic(tlc.NewNullPointerException())
		}
	}
	return append([]any{}, d.parameters...)
}

func (d *SanyErrorDetails) String() string {
	if d == nil {
		panic(tlc.NewNullPointerException())
	}
	if d.location == nil {
		panic(tlc.NewNullPointerException())
	}
	return d.location.String() + "\n\n" + d.GetMessage()
}

func (d *SanyErrorDetails) Equals(other any) bool {
	if d == nil {
		panic(tlc.NewNullPointerException())
	}
	that, ok := other.(*SanyErrorDetails)
	if !ok || that == nil || d.code != that.code {
		return false
	}
	if d.location == nil || that.location == nil {
		panic(tlc.NewNullPointerException())
	}
	return *d.location == *that.location && d.format == that.format && sanyErrorParametersEqual(d.parameters, that.parameters)
}

// SanyErrors adapts SANY's ordered, duplicate-suppressing diagnostic log.
type SanyErrors struct{ diagnostics Diagnostics }

// NewSanyErrors constructs an empty SANY diagnostic log.
func NewSanyErrors() *SanyErrors { return &SanyErrors{} }

func (e *SanyErrors) AddMessage(code SanyErrorCode, location *tlc.SourceLocation, format string, parameters ...any) *SanySemanticAbort {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	loc := tlc.NullSourceLocation
	retainedLocation := &tlc.NullSourceLocation
	if location != nil {
		loc = *location
		retainedLocation = location
	}
	details := &SanyErrorDetails{code: code, location: retainedLocation, format: format, parameters: parameters}
	prefix := "E"
	if code.Severity == SeverityWarning {
		prefix = "W"
	}
	diagnostic := Diagnostic{Code: fmt.Sprintf("%s%d", prefix, code.Value), Severity: code.Severity,
		Pos:     Position{File: loc.Source, Line: loc.BeginLine, Column: loc.BeginColumn, EndLine: loc.EndLine, EndColumn: loc.EndColumn},
		Message: format, sanyDetails: details}
	e.diagnostics = appendSanyDiagnostics(e.diagnostics, diagnostic)
	abort := NewSanySemanticAbort(details, e)
	abort.diagnostic = &diagnostic
	abort.sourceDiagnostics = &e.diagnostics
	return abort
}

func (e *SanyErrors) GetMessages() []*SanyErrorDetails {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	result := make([]*SanyErrorDetails, 0, len(e.diagnostics))
	for _, diagnostic := range e.diagnostics {
		result = append(result, diagnostic.sanyDetails)
	}
	return result
}

func (e *SanyErrors) GetMessagesOfLevel(level SanyErrorLevel) []*SanyErrorDetails {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	result := make([]*SanyErrorDetails, 0)
	for _, diagnostic := range e.diagnostics {
		if diagnostic.sanyDetails != nil && diagnostic.sanyDetails.code.GetSeverityLevel() == level {
			result = append(result, diagnostic.sanyDetails)
		}
	}
	return result
}

func (e *SanyErrors) GetErrorDetails() []*SanyErrorDetails {
	return e.GetMessagesOfLevel(SanyErrorError)
}
func (e *SanyErrors) GetWarningDetails() []*SanyErrorDetails {
	return e.GetMessagesOfLevel(SanyErrorWarning)
}
func (e *SanyErrors) GetErrors() []string   { return sanyRenderDetails(e.GetErrorDetails()) }
func (e *SanyErrors) GetWarnings() []string { return sanyRenderDetails(e.GetWarningDetails()) }
func (e *SanyErrors) IsSuccess() bool       { return len(e.GetErrorDetails()) == 0 }
func (e *SanyErrors) IsFailure() bool       { return !e.IsSuccess() }
func (e *SanyErrors) GetNumErrors() int     { return len(e.GetErrorDetails()) }
func (e *SanyErrors) GetNumMessages() int {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return len(e.diagnostics)
}
func (e *SanyErrors) String() string {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyErrorsString(e.diagnostics)
}

func sanyRenderDetails(details []*SanyErrorDetails) []string {
	result := make([]string, len(details))
	for i, detail := range details {
		result[i] = detail.String()
	}
	return result
}

func sameSanyLoggedDiagnostic(a, b Diagnostic) bool {
	if a.sanyDetails != nil || b.sanyDetails != nil {
		if a.sanyDetails == nil || b.sanyDetails == nil {
			return false
		}
		x, y := a.sanyDetails, b.sanyDetails
		return x.code == y.code && *x.location == *y.location && x.format == y.format &&
			sanyErrorParametersEqual(x.parameters, y.parameters)
	}
	return a.Code == b.Code && sanyJavaErrorDetails(a) == sanyJavaErrorDetails(b)
}

func sanyErrorParametersEqual(left, right []any) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !sanyErrorParameterEqual(left[i], right[i]) {
			return false
		}
	}
	return true
}

func sanyErrorParameterEqual(left, right any) bool {
	if sanyErrorParameterIsNull(left) || sanyErrorParameterIsNull(right) {
		return sanyErrorParameterIsNull(left) && sanyErrorParameterIsNull(right)
	}
	if equal, ok := left.(interface{ Equals(any) bool }); ok {
		return equal.Equals(right)
	}
	leftValue, rightValue := reflect.ValueOf(left), reflect.ValueOf(right)
	if leftValue.Type() != rightValue.Type() {
		return false
	}
	if leftValue.Comparable() {
		return leftValue.Interface() == rightValue.Interface()
	}
	return reflect.DeepEqual(left, right)
}

func sanyErrorParameterIsNull(parameter any) bool {
	if parameter == nil {
		return true
	}
	value := reflect.ValueOf(parameter)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
