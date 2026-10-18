package tlago

import (
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// SanySemanticAbort adapts AbortException and retains the triggering details
// and shared Errors log. Its default Exception message and cause are null;
// toString renders details.
type SanySemanticAbort struct {
	tlc.JavaExceptionBase
	details           *SanyErrorDetails
	sourceErrors      *SanyErrors
	diagnostic        *Diagnostic
	sourceDiagnostics *Diagnostics
}

// NewSanySemanticAbort constructs an abort with Java's nullable detail and log.
func NewSanySemanticAbort(details *SanyErrorDetails, sourceErrorLog *SanyErrors) *SanySemanticAbort {
	return &SanySemanticAbort{JavaExceptionBase: tlc.NewJavaExceptionBase(nil, nil), details: details, sourceErrors: sourceErrorLog}
}

func newSanySemanticAbort(diagnostic Diagnostic, diagnostics Diagnostics, sourceErrorLog *Diagnostics) *SanySemanticAbort {
	if sourceErrorLog == nil && diagnostics != nil {
		retained := diagnostics
		sourceErrorLog = &retained
	}
	var errors *SanyErrors
	if sourceErrorLog != nil {
		errors = &SanyErrors{diagnostics: *sourceErrorLog}
	}
	abort := NewSanySemanticAbort(sanyErrorDetailsForDiagnostic(diagnostic), errors)
	abort.diagnostic = &diagnostic
	abort.sourceDiagnostics = sourceErrorLog
	return abort
}

func sanyErrorDetailsForDiagnostic(diagnostic Diagnostic) *SanyErrorDetails {
	if diagnostic.sanyDetails != nil {
		return diagnostic.sanyDetails
	}
	code, ok := sanyErrorCodeLookupStandardValue(1000)
	if value, err := strconv.Atoi(strings.TrimLeft(diagnostic.Code, "EW")); err == nil {
		if resolved, found := sanyErrorCodeLookupStandardValue(value); found {
			code, ok = resolved, true
		}
	}
	if !ok {
		code = SanyErrorCode{Name: "GENERAL", Value: 1000, Severity: diagnostic.Severity, ParameterCount: SanyErrorCodeVariadicParameters}
	}
	location := tlc.NewSourceLocation(diagnostic.Pos.File, diagnostic.Pos.Line, diagnostic.Pos.Column, diagnostic.Pos.EndLine, diagnostic.Pos.EndColumn)
	message := diagnostic.SANYMessage
	if message == "" {
		message = diagnostic.Message
	}
	parameters := append([]any{}, diagnostic.SANYParameters...)
	return &SanyErrorDetails{code: code, location: &location, format: message, parameters: parameters}
}

func (e *SanySemanticAbort) Error() string {
	if e == nil || e.details == nil {
		panic(tlc.NewNullPointerException())
	}
	if e.diagnostic != nil {
		return sanyJavaErrorDetails(*e.diagnostic)
	}
	return e.details.String()
}
func (e *SanySemanticAbort) JavaThrowableString() string { return e.Error() }
func (e *SanySemanticAbort) GetDetails() *SanyErrorDetails {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return e.details
}
func (e *SanySemanticAbort) GetSourceErrorLog() *SanyErrors {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return e.sourceErrors
}
func (e *SanySemanticAbort) diagnosticLog() *Diagnostics {
	if e.sourceDiagnostics != nil {
		return e.sourceDiagnostics
	}
	if e.sourceErrors != nil {
		return &e.sourceErrors.diagnostics
	}
	return nil
}

// SanySemanticException adapts the semantic driver's checked exception, which
// chains an AbortException and is caught by SANY.parse.
type SanySemanticException struct {
	tlc.JavaExceptionBase
	abort *SanySemanticAbort
}

// NewSanySemanticException constructs the driver exception around an abort.
func NewSanySemanticException(abort *SanySemanticAbort) *SanySemanticException {
	var cause error
	if abort != nil {
		cause = abort
	}
	var message *string
	if cause != nil {
		text := tlc.JavaThrowableString(cause)
		message = &text
	}
	return &SanySemanticException{JavaExceptionBase: tlc.NewJavaExceptionBase(message, cause), abort: abort}
}
func (e *SanySemanticException) Error() string {
	name := "tla2sany.drivers.SemanticException"
	if message := e.GetMessage(); message != nil {
		return name + ": " + *message
	}
	return name
}
func (e *SanySemanticException) JavaThrowableString() string { return e.Error() }
func (e *SanySemanticException) GetDetails() *SanyErrorDetails {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return e.abort.GetDetails()
}
func (e *SanySemanticException) GetSourceErrorLog() *SanyErrors {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return e.abort.GetSourceErrorLog()
}

func (e *SanySemanticException) diagnosticLog() *Diagnostics {
	if e == nil || e.abort == nil {
		return nil
	}
	return e.abort.diagnosticLog()
}
