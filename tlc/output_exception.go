// Portions Copyright (c) 2025, Oracle and/or its affiliates.
package tlc

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
	suppressed, _, _ = messageControlFor(failure.Code)
	if !suppressed {
		var text string
		if failure.NullableParams != nil {
			text = formatNullableMessage(failure.Code, failure.NullableParams, SeverityError)
		} else {
			text = formatMessage(failure.Code, failure.Params, SeverityError)
		}
		printConsoleMessage(failure.Code, SeverityError, text, true)
	}
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
	_, _, warn = messageControlFor(code)
	if warn {
		text := formatMessage(code, params, SeverityWarning)
		printConsoleMessage(code, SeverityWarning, text, true)
		printThrowable(failure)
	}
}

func printThrowable(failure error) {
	if failure == nil {
		panic(NewNullPointerException())
	}
	ToolIOPrint(javaThrowableStackTrace(failure))
}
