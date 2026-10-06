package tlc

import (
	"fmt"
	"os"
)

// DebugPrintMessage ports util.DebugPrinter.print(String). Java's System.out
// is separate from ToolIO.out; native execution uses its actual goroutine ID.
func DebugPrintMessage(message string) {
	Globals.Lock()
	debug := Globals.Debug
	Globals.Unlock()
	if debug {
		_, _ = fmt.Fprintf(os.Stdout, "%d\t%s\n", currentGoroutineID(), message)
	}
}

// DebugPrintThrowable ports util.DebugPrinter.print(Throwable), including its
// distinct stderr routing and source prefix. Stack frames describe Go execution.
func DebugPrintThrowable(failure error) {
	Globals.Lock()
	debug := Globals.Debug
	Globals.Unlock()
	if debug {
		_, _ = fmt.Fprintf(os.Stderr, "%dthrown an exception\n", currentGoroutineID())
		if failure == nil {
			panic(NewNullPointerException())
		}
		_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(failure))
	}
}

// formatMPMessageBody enters MP's private formatting boundary, after any raw
// recorder notification and before template construction. Leaving is reported
// after the severity/tool envelope has been constructed by its caller.
func formatMPMessageBody(code int, params []string, nullable []*string, severity Severity) string {
	if nullable != nil {
		params = messageParameterStrings(nullable)
	}
	DebugPrintMessage(fmt.Sprintf("entering MP.getMessage() with error code %d and %d parameters", code, len(params)))
	for i, parameter := range params {
		if nullable != nil && nullable[i] == nil {
			parameter = "null"
		}
		DebugPrintMessage(fmt.Sprintf("param %d: '%s'", i, parameter))
	}
	if nullable != nil {
		return formatNullableMessage(code, nullable, severity)
	}
	return formatMessage(code, params, severity)
}

func debugMessagePrinterEnter(code int, severity Severity) {
	name := "printMessage"
	switch severity {
	case SeverityError:
		name = "printError"
	case SeverityWarning:
		name = "printWarning"
	case SeverityTLCBug:
		name = "printTLCBug"
	}
	DebugPrintMessage(fmt.Sprintf("entering %s(int, String[]) with errorCode %d", name, code))
}

func debugMessagePrinterLeave(severity Severity) {
	message := "leaving printError(int, String[]) with errorCode "
	switch severity {
	case SeverityError:
		message = "leaving printError(int, String[])"
	case SeverityWarning:
		message = "leaving printWarning(int, String[])"
	case SeverityTLCBug:
		message = "leaving printTLCBug(int, String[])"
	}
	DebugPrintMessage(message)
}
