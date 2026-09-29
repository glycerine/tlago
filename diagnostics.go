package tlago

import (
	"fmt"
	"strings"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

const (
	ExitOK              = 0
	ExitSyntaxFailure   = 2
	ExitSemanticFailure = 4
	ExitToolFailure     = 1
)

type Position struct {
	File      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

func (p Position) String() string {
	if p.File == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}

func (p Position) SourceEnd() Position {
	if p.EndLine > 0 && p.EndColumn > 0 {
		return Position{File: p.File, Line: p.EndLine, Column: p.EndColumn}
	}
	return Position{File: p.File, Line: p.Line, Column: p.Column}
}

type Diagnostic struct {
	Code     string
	Severity Severity
	Pos      Position
	Message  string
}

func (d Diagnostic) String() string {
	code := d.Code
	if code == "" {
		code = "TLAGO"
	}
	return fmt.Sprintf("%s: %s %s: %s", d.Pos, d.Severity, code, d.Message)
}

type Diagnostics []Diagnostic

func (ds Diagnostics) HasErrors() bool {
	for _, d := range ds {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

func (ds Diagnostics) IsSuccess() bool {
	return !ds.HasErrors()
}

func (ds Diagnostics) Warnings() Diagnostics {
	return ds.withSeverity(SeverityWarning)
}

func (ds Diagnostics) Errors() Diagnostics {
	return ds.withSeverity(SeverityError)
}

func (ds Diagnostics) Deduplicated() Diagnostics {
	if len(ds) == 0 {
		return ds
	}
	seen := map[string]bool{}
	out := make(Diagnostics, 0, len(ds))
	for _, d := range ds {
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%s", d.Code, d.Severity, d.Pos.String(), d.Message)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, d)
	}
	return out
}

func (ds Diagnostics) ContainsMessage(text string) bool {
	for _, d := range ds {
		if strings.Contains(d.Message, text) || strings.Contains(d.String(), text) {
			return true
		}
	}
	return false
}

func (ds Diagnostics) Error() string {
	if len(ds) == 0 {
		return ""
	}
	var b strings.Builder
	for i, d := range ds {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(d.String())
	}
	return b.String()
}

func (ds Diagnostics) withSeverity(severity Severity) Diagnostics {
	out := make(Diagnostics, 0, len(ds))
	for _, d := range ds {
		if d.Severity == severity {
			out = append(out, d)
		}
	}
	return out
}

type DiagnosticOptions struct {
	SuppressedCodes map[string]bool
	ElevatedCodes   map[string]bool
}

func (opts DiagnosticOptions) Apply(diags Diagnostics) Diagnostics {
	if len(diags) == 0 {
		return diags
	}
	out := make(Diagnostics, 0, len(diags))
	for _, diag := range diags {
		code := normalizeDiagnosticCode(diag.Code)
		if diagnosticCodeSetContains(opts.SuppressedCodes, code) {
			continue
		}
		if diag.Severity == SeverityWarning && diagnosticCodeSetContains(opts.ElevatedCodes, code) {
			diag.Severity = SeverityError
			diag.Message = "Warning treated as error: " + diag.Message
		}
		out = append(out, diag)
	}
	return out
}

func diagnosticCodeSetContains(set map[string]bool, code string) bool {
	if len(set) == 0 {
		return false
	}
	normalized := normalizeDiagnosticCode(code)
	if set[normalized] {
		return true
	}
	for configured, enabled := range set {
		if enabled && normalizeDiagnosticCode(configured) == normalized {
			return true
		}
	}
	return false
}

func errorAt(pos Position, code, format string, args ...any) Diagnostic {
	return Diagnostic{
		Code:     code,
		Severity: SeverityError,
		Pos:      pos,
		Message:  fmt.Sprintf(format, args...),
	}
}

func warningAt(pos Position, code, format string, args ...any) Diagnostic {
	return Diagnostic{
		Code:     code,
		Severity: SeverityWarning,
		Pos:      pos,
		Message:  fmt.Sprintf(format, args...),
	}
}
