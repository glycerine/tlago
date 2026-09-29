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
