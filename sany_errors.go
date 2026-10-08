// Copyright (c) 2024 Linux Foundation. All rights reserved.
// SPDX-License-Identifier: MIT
package tlago

import (
	"fmt"
	"slices"

	"github.com/glycerine/tlago/tlc"
)

// String-parameter logging retains the format separately from its arguments.
// SANY call sites with no interpolation must never scan percent signs in text.
type sanyErrorDetails struct {
	code       SanyErrorCode
	location   tlc.SourceLocation
	format     string
	parameters []string
}

func (d *sanyErrorDetails) getMessage() string {
	if len(d.parameters) == 0 {
		return d.format
	}
	message, err := tlc.JavaFormatStrings(d.format, d.parameters...)
	if err != nil {
		panic(err)
	}
	return message
}

func (d *sanyErrorDetails) String() string { return d.location.String() + "\n\n" + d.getMessage() }

type sanyErrors struct{ diagnostics Diagnostics }

func (e *sanyErrors) addMessage(code SanyErrorCode, location *tlc.SourceLocation, format string, parameters ...string) *sanySemanticAbort {
	loc := tlc.NullSourceLocation
	if location != nil {
		loc = *location
	}
	details := &sanyErrorDetails{code: code, location: loc, format: format, parameters: parameters}
	prefix := "E"
	if code.Severity == SeverityWarning {
		prefix = "W"
	}
	diagnostic := Diagnostic{Code: fmt.Sprintf("%s%d", prefix, code.Value), Severity: code.Severity,
		Pos:     Position{File: loc.Source, Line: loc.BeginLine, Column: loc.BeginColumn, EndLine: loc.EndLine, EndColumn: loc.EndColumn},
		Message: format, sanyDetails: details}
	e.diagnostics = appendSanyDiagnostics(e.diagnostics, diagnostic)
	return newSanySemanticAbort(diagnostic, nil, &e.diagnostics)
}

func (e *sanyErrors) getMessages() []*sanyErrorDetails {
	result := make([]*sanyErrorDetails, 0, len(e.diagnostics))
	for _, diagnostic := range e.diagnostics {
		result = append(result, diagnostic.sanyDetails)
	}
	return result
}

func (e *sanyErrors) getMessagesOfLevel(level Severity) []*sanyErrorDetails {
	result := make([]*sanyErrorDetails, 0)
	for _, diagnostic := range e.diagnostics {
		if diagnostic.Severity == level {
			result = append(result, diagnostic.sanyDetails)
		}
	}
	return result
}

func (e *sanyErrors) getErrorDetails() []*sanyErrorDetails {
	return e.getMessagesOfLevel(SeverityError)
}
func (e *sanyErrors) getWarningDetails() []*sanyErrorDetails {
	return e.getMessagesOfLevel(SeverityWarning)
}
func (e *sanyErrors) getErrors() []string   { return sanyRenderDetails(e.getErrorDetails()) }
func (e *sanyErrors) getWarnings() []string { return sanyRenderDetails(e.getWarningDetails()) }
func (e *sanyErrors) isSuccess() bool       { return len(e.getErrorDetails()) == 0 }
func (e *sanyErrors) isFailure() bool       { return !e.isSuccess() }
func (e *sanyErrors) getNumErrors() int     { return len(e.getErrorDetails()) }
func (e *sanyErrors) getNumMessages() int   { return len(e.diagnostics) }
func (e *sanyErrors) String() string        { return sanyErrorsString(e.diagnostics) }

func sanyRenderDetails(details []*sanyErrorDetails) []string {
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
		return x.code == y.code && x.location == y.location && x.format == y.format &&
			slices.Equal(x.parameters, y.parameters)
	}
	return a.Code == b.Code && sanyJavaErrorDetails(a) == sanyJavaErrorDetails(b)
}
