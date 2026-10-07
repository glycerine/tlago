// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// AbortException retains the triggering details and the shared Errors log.
// Its default Exception message and cause are null; toString renders details.
type sanySemanticAbort struct {
	tlc.JavaExceptionBase
	diagnostic     Diagnostic
	diagnostics    Diagnostics
	sourceErrorLog *Diagnostics
}

func newSanySemanticAbort(diagnostic Diagnostic, diagnostics Diagnostics, sourceErrorLog *Diagnostics) *sanySemanticAbort {
	return &sanySemanticAbort{JavaExceptionBase: tlc.NewJavaExceptionBase(nil, nil), diagnostic: diagnostic, diagnostics: diagnostics, sourceErrorLog: sourceErrorLog}
}

func (e *sanySemanticAbort) Error() string               { return sanyJavaErrorDetails(e.diagnostic) }
func (e *sanySemanticAbort) JavaThrowableString() string { return e.Error() }
func (e *sanySemanticAbort) GetDetails() *Diagnostic {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return &e.diagnostic
}
func (e *sanySemanticAbort) GetSourceErrorLog() *Diagnostics {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	if e.sourceErrorLog != nil {
		return e.sourceErrorLog
	}
	return &e.diagnostics
}

// The semantic driver chains an AbortException; SANY.parse catches this family.
type sanySemanticException struct {
	tlc.JavaExceptionBase
	abort *sanySemanticAbort
}

func newSanySemanticException(abort *sanySemanticAbort) *sanySemanticException {
	var cause error
	if abort != nil {
		cause = abort
	}
	var message *string
	if cause != nil {
		text := tlc.JavaThrowableString(cause)
		message = &text
	}
	return &sanySemanticException{JavaExceptionBase: tlc.NewJavaExceptionBase(message, cause), abort: abort}
}
func (e *sanySemanticException) Error() string {
	name := "tla2sany.drivers.SemanticException"
	if message := e.GetMessage(); message != nil {
		return name + ": " + *message
	}
	return name
}
func (e *sanySemanticException) JavaThrowableString() string { return e.Error() }
func (e *sanySemanticException) GetDetails() *Diagnostic {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return e.abort.GetDetails()
}
func (e *sanySemanticException) GetSourceErrorLog() *Diagnostics {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
	return e.abort.GetSourceErrorLog()
}
