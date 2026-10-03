/*******************************************************************************
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
// Translation of the complete Debug05SimTest.testSpec after production fixes.
package tlago

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaDebug05DebuggerSim(t *testing.T) {
	const RM = "Debug05"
	// Isolate the source BASE_DIR and its relative trace export files.
	baseDir := t.TempDir()
	classpath, err := tlcApplicationClasspath(nil)
	if err != nil {
		t.Fatal(err)
	}
	resolver := tlc.NewSimpleFilenameToStream([]string{baseDir}, tlc.FilenameResolverOptions{Classpath: classpath})
	h := startJavaDebuggerModelWithOptions(t, "debug", RM, tlc.ExitStatusSuccess,
		javaDebuggerModelOptions{dumpTrace: false, resolver: resolver},
		"-config", "Debug05.tla", "-simulate", "num=1", "-depth", "25")
	defer h.close()
	expression := func(s string) *string { return &s }
	stackFrames := h.frames()

	ea := tlc.TLCEvaluateArguments{}
	ea.FrameID = &stackFrames[0].Base.ID
	ea.Context = "repl"

	// Evaluate expressions that require Java module overrides.

	// An expression involving a module without any dependencies.
	ea.Expression = expression("LET N == INSTANCE Naturals IN N!+(1, 2)")
	debuggerAssertEqual(t, "3", *h.debugger.Evaluate(ea).Result)

	// An expression involving a module with one dependency.
	ea.Expression = expression("LET S == INSTANCE Sequences IN S!Len(<<1,2,3>>)")
	debuggerAssertEqual(t, "3", *h.debugger.Evaluate(ea).Result)

	ea.Expression = expression("LET N == INSTANCE Integers IN N!+(1, 2)")
	debuggerAssertEqual(t, "3", *h.debugger.Evaluate(ea).Result)

	// An expression involving a module with multiple dependencies.
	ea.Expression = expression("LET T == INSTANCE TLC IN T!RandomElement({1})")
	debuggerAssertEqual(t, "1", *h.debugger.Evaluate(ea).Result)

	ea.Expression = expression("LET B == INSTANCE Bags IN B!SetToBag({\"a\", \"a\", \"b\"})")
	result := *h.debugger.Evaluate(ea).Result
	debuggerAssertTrue(t, result == "[b |-> 1, a |-> 1]" || result == "[a |-> 1, b |-> 1]")

	ea.Expression = expression("LET J == INSTANCE Json IN J!ToJson({1,2,3})")
	debuggerAssertEqual(t, "[1,2,3]", *h.debugger.Evaluate(ea).Result)

	// 88888888888888888888888888888888888888888888888888888 //

	h.setSpecBreakpoint()
	h.continueFrames()

	for i := 1; i < 20; i++ {
		stackFrames = h.continueFrames()
		stackFrame := stackFrames[0]
		ea.FrameID = &stackFrame.Base.ID

		traceFile := filepath.ToSlash(filepath.Join(baseDir, "Debug05SimTest.json"))
		ea.Expression = expression("LET J == INSTANCE _JsonTrace WITH _JsonTraceFile <- \"" + traceFile + "\", _JsonTraceInputFile <- \"" + traceFile + "\" IN J!_JsonTrace")
		debuggerAssertEqual(t, "TRUE", *h.debugger.Evaluate(ea).Result)

		// Validate the content of the generated file (by using the very functionality
		// we are testing).
		ea.Expression = expression("LET S == INSTANCE Sequences J == INSTANCE Json T == J!JsonDeserialize(\"" + traceFile + "\") IN S!Len(T.counterexample.state) = " + strconv.Itoa(i))
		debuggerAssertEqual(t, "TRUE", *h.debugger.Evaluate(ea).Result)

		// 88888888888888888888888888888888888888888888888888888 //

		traceFile = filepath.ToSlash(filepath.Join(baseDir, "Debug05SimTest.bin"))
		ea.Expression = expression("LET J == INSTANCE _TLCTrace WITH _TLCTraceFile <- \"" + traceFile + "\", _TLCTraceInputFile <- \"" + traceFile + "\" IN J!_TLCTrace")
		debuggerAssertEqual(t, "TRUE", *h.debugger.Evaluate(ea).Result)

		// 88888888888888888888888888888888888888888888888888888 //
	}

	h.unsetBreakpoints()
	h.continueFrames()
}
