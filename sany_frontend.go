package tlago

import (
	"fmt"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type sanyDiagnosticPhase int

const (
	sanyParsePhase sanyDiagnosticPhase = iota
	sanySemanticPhase
	sanyLintPhase
)

// runSanyFrontEnd preserves the SANY.parse exception boundaries. parseFailed
// includes a caught ParseException-equivalent even when parseErrors is empty.
// Legacy SANY returns OK for ordinary semantic errors; callers inspect them.
func runSanyFrontEnd(file string, opts LoadOptions, report func(Diagnostics, sanyDiagnosticPhase) Diagnostics) (spec *Spec, parseDiags, semanticDiags Diagnostics, parseFailed bool) {
	println := opts.ParsingProgress
	if println == nil {
		println = func(string) {}
	}
	defer func() {
		if failure := recover(); failure != nil {
			exception, ok := failure.(error)
			if !ok || tlc.IsJavaError(exception) {
				panic(failure)
			}
			// SANY.parse logs the unexpected exception, then chains it in its
			// checked FrontEndException. Errors and logging failures propagate.
			println(tlc.JavaThrowableString(exception))
			panic(tlc.NewFrontEndExceptionFromCause(exception))
		}
	}()
	loader := newSanyLoader(opts)
	loader.initialContext = sanyGlobalInitialContext(true)
	spec, parseDiags, parseFailed = runSanyFrontEndParse(file, loader, report)
	if parseFailed {
		return
	}
	var controlled Diagnostics
	semanticDiags = checkSpecWithModuleReport(spec, opts.ParsingProgress, func(accumulated Diagnostics) {
		controlled = accumulated
		if report != nil {
			controlled = report(accumulated, sanySemanticPhase)
		}
	})
	if !controlled.HasErrors() {
		lintDiags := lintSanySpec(spec, opts.ParsingProgress)
		semanticDiags = append(semanticDiags, lintDiags...)
		if report != nil {
			lintDiags = report(lintDiags, sanyLintPhase)
		}
		controlled = append(controlled, lintDiags...)
	}
	spec.Diags = append(append(Diagnostics(nil), parseDiags...), controlled...)
	return
}

// runSanyFrontEndParse is the actual parsing phase and its checked failure
// boundary. Keeping it separate also lets original parsing tests inspect their
// recorded parser output without running semantic analysis.
func runSanyFrontEndParse(file string, loader *sanyLoader, report func(Diagnostics, sanyDiagnosticPhase) Diagnostics) (spec *Spec, parseDiags Diagnostics, parseFailed bool) {
	println := loader.opts.ParsingProgress
	if println == nil {
		println = func(string) {}
	}
	defer func() {
		if failure := recover(); failure != nil {
			exception, ok := failure.(error)
			if !ok || tlc.IsJavaError(exception) {
				panic(failure)
			}
			// frontEndParse converts caught Exceptions to ParseException.
			// Preserve diagnostics accumulated before the interrupted load.
			parseFailed = true
			parseDiags = loader.diags
			spec = loader.snapshot(nil)
			println(fmt.Sprintf("\nFatal errors while parsing TLA+ spec in file %s\n", file))
			if abort, ok := exception.(*sanyParseAbort); ok {
				println(abort.Error())
			} else {
				println(tlc.JavaThrowableString(exception))
			}
			println(sanyErrorsString(parseDiags))
		}
	}()
	spec, parseDiags = loader.loadSpec(file)
	controlled := parseDiags
	if report != nil {
		controlled = report(parseDiags, sanyParsePhase)
	}
	parseFailed = controlled.HasErrors()
	return
}

// sanyErrorsString follows Errors.toString's grouping and trailing newlines.
// Mapping all parser diagnostics to original ErrorDetails remains separate work.
func sanyErrorsString(diags Diagnostics) string {
	var output strings.Builder
	for _, severity := range []Severity{SeverityError, SeverityWarning} {
		var matching Diagnostics
		for _, diagnostic := range diags {
			if diagnostic.Severity == severity {
				matching = append(matching, diagnostic)
			}
		}
		if len(matching) == 0 {
			continue
		}
		kind := "Errors"
		if severity == SeverityWarning {
			kind = "Warnings"
		}
		fmt.Fprintf(&output, "*** %s: %d\n\n", kind, len(matching))
		for _, diagnostic := range matching {
			output.WriteString(sanyJavaErrorDetails(diagnostic))
			output.WriteString("\n\n\n")
		}
	}
	return output.String()
}
