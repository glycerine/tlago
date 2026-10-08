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
// includes caught parse or semantic checked failures, even when parseErrors is empty.
// Legacy SANY returns OK for ordinary semantic errors; callers inspect them.
func runSanyFrontEnd(file string, opts LoadOptions, report func(Diagnostics, sanyDiagnosticPhase) Diagnostics) (spec *Spec, parseDiags, semanticDiags Diagnostics, parseFailed bool) {
	return runSanyFrontEndWithSettings(file, opts, defaultSanyDriverSettings(), report)
}

// These switches control the generation, level-checking and linting phases.
// The legacy TLC adapter keeps its separate exit-status policy.
type sanyExitCode int

const (
	sanyExitOK              sanyExitCode = 0
	sanyExitSyntaxFailure   sanyExitCode = 2
	sanyExitSemanticFailure sanyExitCode = 4
	sanyExitError           sanyExitCode = -1
)

type sanyDriverSettings struct {
	strict, semantic, levels, lint bool
	messages                       diagnosticCLIOptions
}

func defaultSanyDriverSettings() sanyDriverSettings {
	return sanyDriverSettings{strict: true, semantic: true, levels: true, lint: true}
}

// parseSanyWithSettings follows SANY.parse's named exit codes. The surrounding
// tlago CLI maps SANY ERROR (-1) to its process-level tool failure status (1).
func parseSanyWithSettings(file string, opts LoadOptions, settings sanyDriverSettings, output func(Severity, string)) (*Spec, sanyExitCode) {
	status := sanyExitOK
	report := func(raw Diagnostics, phase sanyDiagnosticPhase) Diagnostics {
		visible := settings.messages.apply(raw)
		if visible.HasErrors() {
			status = sanyExitSemanticFailure
			if phase == sanyParsePhase {
				status = sanyExitSyntaxFailure
			}
		}
		reportSanyMessages(file, raw, phase, settings.messages, output)
		return visible
	}
	spec, _, _, failed := runSanyFrontEndWithSettings(file, opts, settings, report)
	if failed {
		return spec, sanyExitError
	}
	if !settings.strict {
		return spec, sanyExitOK
	}
	return spec, status
}

func reportSanyMessages(file string, raw Diagnostics, phase sanyDiagnosticPhase, controls diagnosticCLIOptions, output func(Severity, string)) {
	if output == nil {
		return
	}
	var warnings Diagnostics
	for _, d := range raw.Warnings() {
		if !diagnosticCodeSetContains(controls.suppressed, d.Code) {
			warnings = append(warnings, d)
		}
	}
	if len(warnings) > 0 && phase == sanySemanticPhase {
		output(SeverityWarning, fmt.Sprintf("*** Warnings: %d\n", len(warnings)))
	}
	if len(warnings) > 0 && phase == sanyParsePhase {
		output(SeverityWarning, fmt.Sprintf("Warnings (%d) during syntax parsing of %s:\n\n", len(warnings), file))
	}
	for _, elevated := range []bool{false, true} {
		for _, d := range warnings {
			if diagnosticCodeSetContains(controls.elevated, d.Code) != elevated {
				continue
			}
			severity, prefix := SeverityWarning, ""
			if elevated {
				severity, prefix = SeverityError, "Warning treated as error: "
			}
			output(severity, prefix+sanyJavaErrorDetails(d)+"\n\n\n")
		}
	}
	errors := raw.Errors()
	if len(errors) > 0 {
		prefix := ""
		if phase == sanySemanticPhase {
			prefix = "Semantic errors:\n\n"
		}
		output(SeverityError, prefix+sanyErrorsString(errors))
	}
}

func runSanyFrontEndWithSettings(file string, opts LoadOptions, settings sanyDriverSettings, report func(Diagnostics, sanyDiagnosticPhase) Diagnostics) (spec *Spec, parseDiags, semanticDiags Diagnostics, parseFailed bool) {
	println := opts.ParsingProgress
	if println == nil {
		println = func(string) {}
	}
	defer func() {
		if failure := recover(); failure != nil {
			if exception, ok := failure.(*sanySemanticException); ok {
				parseFailed = true // Legacy SANY.parse returns ERROR for this checked failure.
				semanticDiags = append(Diagnostics(nil), (*exception.GetSourceErrorLog())...)
				if spec != nil {
					spec.Diags = append(append(Diagnostics(nil), parseDiags...), semanticDiags...)
				}
				return
			}
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
	var controlledParse Diagnostics
	parseReport := func(raw Diagnostics, phase sanyDiagnosticPhase) Diagnostics {
		controlledParse = raw
		if report != nil {
			controlledParse = report(raw, phase)
		}
		return controlledParse
	}
	spec, parseDiags, parseFailed = runSanyFrontEndParse(file, loader, parseReport)
	if parseFailed || !settings.semantic {
		return
	}
	var controlled Diagnostics
	semanticDiags = runSanyFrontEndSemanticsWithLevels(file, spec, opts.ParsingProgress, settings.levels, func(accumulated Diagnostics) {
		controlled = accumulated
		if report != nil {
			controlled = report(accumulated, sanySemanticPhase)
		}
	})
	if settings.levels && settings.lint && !controlledParse.HasErrors() && !controlled.HasErrors() {
		lintDiags := lintSanySpec(spec, opts.ParsingProgress)
		semanticDiags = append(semanticDiags, lintDiags...)
		if report != nil {
			lintDiags = report(lintDiags, sanyLintPhase)
		}
		controlled = append(controlled, lintDiags...)
	}
	spec.Diags = append(append(Diagnostics(nil), controlledParse...), controlled...)
	return
}

// frontEndSemanticAnalysis catches only AbortException; other failures propagate.
func runSanyFrontEndSemantics(file string, spec *Spec, println func(string), report func(Diagnostics)) (diagnostics Diagnostics) {
	return runSanyFrontEndSemanticsWithLevels(file, spec, println, true, report)
}

func runSanyFrontEndSemanticsWithLevels(file string, spec *Spec, println func(string), levels bool, report func(Diagnostics)) (diagnostics Diagnostics) {
	if println == nil {
		println = func(string) {}
	}
	defer func() {
		if failure := recover(); failure != nil {
			abort, ok := failure.(*sanySemanticAbort)
			if !ok {
				panic(failure)
			}
			println(fmt.Sprintf("Fatal errors in semantic processing of TLA spec %s\nnull\nStack trace for exception:\n", file))
			println(strings.TrimSuffix(tlc.JavaThrowableStackTrace(abort), "\n"))
			if log := *abort.GetSourceErrorLog(); len(log) != 0 {
				println("Semantic errors detected before the unexpected exception:\n\n" + sanyErrorsString(log))
			}
			panic(newSanySemanticException(abort))
		}
	}()
	return generateSpecWithModuleReport(spec, println, report, levels)
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
	if report != nil {
		report(parseDiags, sanyParsePhase)
	}
	// Elevating a parsing warning changes the status but does not throw ParseException.
	parseFailed = parseDiags.HasErrors()
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
