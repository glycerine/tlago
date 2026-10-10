package tlc

import (
	"strings"
	"testing"
)

// Entire original tlc2.output.MPTest. ToolIO captures actual production println
// calls, rather than reconstructing console output from recorder metadata.
func TestJavaMP(t *testing.T) {
	run := func(name string, body func(*testing.T)) {
		t.Run(name, func(t *testing.T) {
			toolIO.Lock()
			oldMode, oldOut, oldErr := toolIO.mode, toolIO.out, toolIO.err
			oldOutHigh, oldErrHigh := toolIO.outHigh, toolIO.errHigh
			oldSystemOutHigh, oldSystemErrHigh := toolIOSystemOutHigh, toolIOSystemErrHigh
			oldMessages, oldNext := toolIO.messages, toolIO.nextMessage
			oldCaptureOut, oldCaptureErr := toolIO.captureOut, toolIO.captureErr
			toolIO.Unlock()
			Globals.Lock()
			oldTool := Globals.Tool
			Globals.Tool = false // Original standalone JUnit JVM default.
			Globals.Unlock()
			t.Cleanup(func() {
				toolIO.Lock()
				toolIO.mode, toolIO.out, toolIO.err = oldMode, oldOut, oldErr
				toolIO.outHigh, toolIO.errHigh = oldOutHigh, oldErrHigh
				toolIOSystemOutHigh, toolIOSystemErrHigh = oldSystemOutHigh, oldSystemErrHigh
				toolIO.captureOut, toolIO.captureErr = oldCaptureOut, oldCaptureErr
				toolIO.messages, toolIO.nextMessage = oldMessages, oldNext
				toolIO.Unlock()
				Globals.Lock()
				Globals.Tool = oldTool
				Globals.Unlock()
			})
			// Original per-method @Before.
			ToolIOSetMode(ToolIOTool)
			ToolIOReset()
			body(t)
		})
	}
	run("testPrintErrorInt", func(t *testing.T) {
		PrintError(ECUnitTest)
		allMessages := ToolIOGetAllMessages()
		javaMCEquals(t, 1, len(allMessages))
		javaMCEquals(t, "Error: [%1%][%2%]", allMessages[0])
	})
	run("testPrintErrorIntString", func(t *testing.T) {
		parameter := "EXPECTED"
		PrintError(ECUnitTest, parameter)
		allMessages := ToolIOGetAllMessages()
		javaMCEquals(t, 1, len(allMessages))
		javaMCEquals(t, "Error: ["+parameter+"][%2%]", allMessages[0])
	})
	run("testPrintErrorIntStringArray", func(t *testing.T) {
		parameters := []string{"EXPECTED", "EXPECTED2", "TOO MANY"}
		PrintError(ECUnitTest, parameters...)
		allMessages := ToolIOGetAllMessages()
		javaMCEquals(t, 1, len(allMessages))
		javaMCEquals(t, "Error: ["+parameters[0]+"]["+parameters[1]+"]", allMessages[0])
	})
	run("testPrintProgressStats", func(t *testing.T) {
		parameters := []string{
			"this.trace.getLevelForReporting()",
			MessageNumberFormat(3000000),
			MessageNumberFormat(5000),
			MessageNumberFormat(1222333444),
			MessageNumberFormat(10000),
			MessageNumberFormat(1234),
		}
		PrintMessage(ECTLCProgressStats, parameters...)
		allMessages := ToolIOGetAllMessages()
		javaMCEquals(t, 1, len(allMessages))
		if !strings.Contains(allMessages[0], "3,000,000 states generated (10,000 s/min), 5,000 distinct states found (1,234 ds/min), 1,222,333,444 states left on queue.") &&
			!strings.Contains(allMessages[0], "3.000.000 states generated (10.000 s/min), 5.000 distinct states found (1.234 ds/min), 1.222.333.444 states left on queue.") {
			t.Fatal(allMessages[0])
		}
	})
}
