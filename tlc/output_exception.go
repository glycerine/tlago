// Portions Copyright (c) 2025, Oracle and/or its affiliates.
package tlc

import (
	"fmt"
	"strings"
)

// PrintErrorThrowable ports MP.printError(int, Throwable). General errors use
// ECGeneralMsg; other codes substitute the nullable detail and print a stack
// only with TLCGlobals.debug enabled.
func PrintErrorThrowable(code int, cause error) int {
	if cause == nil {
		panic(NewNullPointerException())
	}
	if code == ECGeneral {
		PrintErrorWithCause(code, "", cause)
	} else {
		PrintErrorNullable(code, javaThrowableDetailMessage(cause))
		printDebugThrowable(cause)
	}
	return code
}

// PrintErrorWithCause ports the String-cause overload. GENERAL has its own
// stack policy in ECGeneralMsg, independent of the global debug switch.
func PrintErrorWithCause(code int, cause string, failure error) {
	if code == ECGeneral {
		PrintError(code, javaGeneralErrorMessage(cause, failure))
	} else {
		PrintError(code, cause)
		printDebugThrowable(failure)
	}
}

// PrintErrorParametersWithThrowable ports the String[]-cause overload, which
// retains the ordinary parameter event even for GENERAL.
func PrintErrorParametersWithThrowable(code int, params []*string, failure error) {
	PrintErrorNullable(code, params...)
	printDebugThrowable(failure)
}

func printDebugThrowable(failure error) {
	Globals.Lock()
	debug := Globals.Debug
	Globals.Unlock()
	if debug {
		DebugPrintMessage("printing stacktrace in printError(int, Throwable, boolean)")
		printThrowable(failure)
	}
}

// PrintTLCRuntimeException ports MP.printTLCRuntimeException. A coded exception
// with parameters is recorded as the exception object, not a string array. Its
// suppressed path skips formatting; the legacy no-parameter path prints an
// ordinary throwable error and uses that overload's debug policy.
func PrintTLCRuntimeException(failure *TLCError) int {
	if failure == nil {
		panic(NewNullPointerException())
	}
	if failure.Params == nil && failure.NullableParams == nil {
		return PrintErrorThrowable(failure.Code, failure)
	}
	suppressed, _, _ := messageControlFor(failure.Code)
	defaultRecorder.Record(Message{Code: failure.Code, Severity: SeverityError,
		Throwable: failure, Suppressed: suppressed})
	DebugPrintMessage(fmt.Sprintf("entering printTLCRuntimeException(TLCRuntimeException) with errorCode %d", failure.Code))
	suppressed, _, _ = messageControlFor(failure.Code)
	if !suppressed {
		text := formatMPMessageBody(failure.Code, failure.Params, failure.NullableParams, SeverityError)
		printConsoleMessage(failure.Code, SeverityError, text, true)
	}
	DebugPrintMessage("leaving printTLCRuntimeException(TLCRuntimeException) with errorCode ")
	return failure.Code
}

// PrintWarningThrowable ports MP.printWarning(int, String, Throwable). Enabled
// warnings print the throwable stack even when the message is suppressed or
// already in history, independently of TLCGlobals.debug.
func PrintWarningThrowable(code int, parameter string, failure error) {
	_, asError, _ := messageControlFor(code)
	warningToError, _ := tlcLookupSystemProperty("tlc2.output.MP.warning2error")
	if javaBooleanProperty(warningToError) || asError {
		panic(newTLCRuntimeExceptionWithCause(code, failure))
	}
	suppressed, _, warn := messageControlFor(code)
	params := []string{parameter}
	defaultRecorder.Record(Message{Code: code, Severity: SeverityWarning,
		Params: params, Throwable: failure, Suppressed: suppressed || !warn})
	DebugPrintMessage(fmt.Sprintf("entering printWarning(int, String, Exception) with errorCode %d", code))
	_, _, warn = messageControlFor(code)
	if warn {
		text := formatMPMessageBody(code, params, nil, SeverityWarning)
		printConsoleMessage(code, SeverityWarning, text, true)
		DebugPrintMessage("printing stacktrace in printError(int, Throwable, boolean)")
		printThrowable(failure)
	}
	DebugPrintMessage("leaving printWarning(int, String[])")
}

func printThrowable(failure error) {
	if failure == nil {
		panic(NewNullPointerException())
	}
	// Throwable.printStackTrace uses println for each line. A single print
	// leaves ToolIO's buffered stream holding an unfinished message instead.
	for _, line := range strings.Split(strings.TrimSuffix(javaThrowableStackTrace(failure), "\n"), "\n") {
		ToolIOPrintln(line)
	}
}
