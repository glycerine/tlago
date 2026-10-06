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

type diagnosticCodeInfo struct {
	Severity Severity
}

var diagnosticCodeInfos = map[string]diagnosticCodeInfo{
	"E4003": {Severity: SeverityError},
	"E1200": {Severity: SeverityError},
	"E1201": {Severity: SeverityError},
	"E1202": {Severity: SeverityError},
	"E1203": {Severity: SeverityError},
	"E1300": {Severity: SeverityError},
	"E1301": {Severity: SeverityError},
	"E1302": {Severity: SeverityError},
	"E1303": {Severity: SeverityError},
	"E1304": {Severity: SeverityError},
	"E1305": {Severity: SeverityError},
	"E1306": {Severity: SeverityError},
	"E1307": {Severity: SeverityError},
	"E1308": {Severity: SeverityError},
	"E1309": {Severity: SeverityError},
	"E1310": {Severity: SeverityError},
	"E1311": {Severity: SeverityError},
	"E1312": {Severity: SeverityError},
	"E1313": {Severity: SeverityError},
	"E1314": {Severity: SeverityError},
	"E1315": {Severity: SeverityError},
	"E1316": {Severity: SeverityError},
	"E1317": {Severity: SeverityError},
	"E1318": {Severity: SeverityError},
	"E1319": {Severity: SeverityError},
	"E1320": {Severity: SeverityError},
	"E1321": {Severity: SeverityError},
	"E1322": {Severity: SeverityError},
	"E1400": {Severity: SeverityError},
	"E1401": {Severity: SeverityError},
	"E1402": {Severity: SeverityError},
	"E1403": {Severity: SeverityError},
	"E1404": {Severity: SeverityError},
	"E1405": {Severity: SeverityError},
	"E1406": {Severity: SeverityError},
	"E1407": {Severity: SeverityError},
	"E1408": {Severity: SeverityError},
	"E1409": {Severity: SeverityError},
	"E1410": {Severity: SeverityError},
	"E1411": {Severity: SeverityError},
	"E1412": {Severity: SeverityError},
	"E1413": {Severity: SeverityError},
	"E1414": {Severity: SeverityError},
	"E1500": {Severity: SeverityError},
	"E1501": {Severity: SeverityError},
	"E1502": {Severity: SeverityError},
	"E1503": {Severity: SeverityError},
	"E1504": {Severity: SeverityError},
	"E1505": {Severity: SeverityError},
	"E1506": {Severity: SeverityError},
	"E1507": {Severity: SeverityError},
	"E1508": {Severity: SeverityError},
	"E1510": {Severity: SeverityError},
	"E1511": {Severity: SeverityError},
	"E1512": {Severity: SeverityError},
	"E1513": {Severity: SeverityError},
	"E1514": {Severity: SeverityError},
	"E1515": {Severity: SeverityError},
	"E1516": {Severity: SeverityError},
	"E4200": {Severity: SeverityError},
	"E4201": {Severity: SeverityError},
	"E4202": {Severity: SeverityError},
	"E4203": {Severity: SeverityError},
	"E4204": {Severity: SeverityError},
	"E4205": {Severity: SeverityError},
	"E4206": {Severity: SeverityError},
	"E4220": {Severity: SeverityError},
	"E4221": {Severity: SeverityError},
	"E4222": {Severity: SeverityError},
	"E4223": {Severity: SeverityError},
	"E4224": {Severity: SeverityError},
	"E4240": {Severity: SeverityError},
	"E4241": {Severity: SeverityError},
	"E4242": {Severity: SeverityError},
	"E4243": {Severity: SeverityError},
	"E4244": {Severity: SeverityError},
	"E4245": {Severity: SeverityError},
	"E4246": {Severity: SeverityError},
	"E4247": {Severity: SeverityError},
	"E4260": {Severity: SeverityError},
	"E4261": {Severity: SeverityError},
	"E4262": {Severity: SeverityError},
	"E4270": {Severity: SeverityError},
	"E4271": {Severity: SeverityError},
	"E4272": {Severity: SeverityError},
	"E4273": {Severity: SeverityError},
	"E4274": {Severity: SeverityError},
	"E4275": {Severity: SeverityError},
	"E4290": {Severity: SeverityError},
	"E4291": {Severity: SeverityError},
	"E4292": {Severity: SeverityError},
	"E4293": {Severity: SeverityError},
	"E4294": {Severity: SeverityError},
	"E4310": {Severity: SeverityError},
	"E4311": {Severity: SeverityError},
	"E4312": {Severity: SeverityError},
	"E4313": {Severity: SeverityError},
	"E4314": {Severity: SeverityError},
	"E4315": {Severity: SeverityError},
	"E4330": {Severity: SeverityError},
	"E4331": {Severity: SeverityError},
	"E4332": {Severity: SeverityError},
	"E4333": {Severity: SeverityError},
	"E4334": {Severity: SeverityError},
	"E4335": {Severity: SeverityError},
	"E4336": {Severity: SeverityError},
	"E4337": {Severity: SeverityError},
	"E4350": {Severity: SeverityError},
	"E4351": {Severity: SeverityError},
	"E4352": {Severity: SeverityError},
	"E4353": {Severity: SeverityError},
	"E4354": {Severity: SeverityError},
	"E4355": {Severity: SeverityError},
	"E4356": {Severity: SeverityError},
	"E4357": {Severity: SeverityError},
	"E4801": {Severity: SeverityError},
	"E4802": {Severity: SeverityError},
	"E4803": {Severity: SeverityError},
	"E6000": {Severity: SeverityError},
	"E6001": {Severity: SeverityError},
	"E6002": {Severity: SeverityError},
	"E6003": {Severity: SeverityError},
	"E6004": {Severity: SeverityError},
	"E6005": {Severity: SeverityError},
	"E6006": {Severity: SeverityError},
	"E6007": {Severity: SeverityError},
	"E6008": {Severity: SeverityError},
	"E6009": {Severity: SeverityError},
	"E7000": {Severity: SeverityError},
	"E7001": {Severity: SeverityError},
	"E7002": {Severity: SeverityError},
	"E7003": {Severity: SeverityError},
	"E7004": {Severity: SeverityError},
	"E7005": {Severity: SeverityError},
	"E7006": {Severity: SeverityError},
	"E7007": {Severity: SeverityError},
	"E7010": {Severity: SeverityError},
	"E7011": {Severity: SeverityError},
	"W4800": {Severity: SeverityWarning},
	"W4801": {Severity: SeverityWarning},
	"W4802": {Severity: SeverityWarning},
	"W4803": {Severity: SeverityWarning},
	"W4804": {Severity: SeverityWarning},
	"W4805": {Severity: SeverityWarning},
}

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
	Code             string
	Severity         Severity
	Pos              Position
	Message          string
	SANYRange        SanyRange
	SANYMessage      string
	SANYParameters   []any
	SANYParseMessage string
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

func lookupDiagnosticCode(code string) (string, diagnosticCodeInfo, bool) {
	normalized := normalizeDiagnosticCode(code)
	info, ok := diagnosticCodeInfos[normalized]
	return normalized, info, ok
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
