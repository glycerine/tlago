/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
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

// Translation of the complete EWD998TraceDebuggerTest.testSpec.
package tlago

import (
	"net/url"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaEWD998TraceDebugger(t *testing.T) {
	const FOLDER, MDL = "EWD998", "EWD998_TTrace"
	h := startJavaDebuggerModel(t, FOLDER, MDL, tlc.ExitStatusViolationLiveness, "-config", "EWD998_TTrace.tla", "-noGenerateSpecTE")
	defer h.close()
	h.replaceBreakpoints(MDL, 67)
	frames := h.continueFrames()
	path, err := filepath.Abs(MDL)
	if err != nil {
		t.Fatal(err)
	}
	uri := url.URL{Scheme: "tlaplus", Path: filepath.ToSlash(path), RawQuery: "_TETrace", Fragment: "66 22 66 29"}
	expression := "tlaplus://" + uri.String()[len("tlaplus:"):]
	id := frames[0].Base.ID
	variable := h.debugger.Evaluate(tlc.TLCEvaluateArguments{Context: "hover", Expression: &expression, FrameID: &id})
	debuggerAssertEqual(t, "TupleValue: "+tlc.EmptyTuple.KindString(), *variable.Type)
	debuggerAssertTrue(t, variable.VariablesReference != 0)
	debuggerAssertEqual(t, "<<[color |-> (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\" @@ 4 :> \"white\"), pending |-> (0 :> 0 @@ 1 :> 0 @@ 2 :> 0 @@ 3 :> 0 @@ 4 :> 0), active |-> (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> FALSE @@ 4 :> FALSE), counter |-> (0 :> 0 @@ 1 :> 0 @@ 2 :> 0 @@ 3 :> 0 @@ 4 :> 0), token |-> [color |-> \"black\", pos |-> 0, q |-> 0]], [color |-> (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\" @@ 4 :> \"white\"), pending |-> (0 :> 0 @@ 1 :> 0 @@ 2 :> 0 @@ 3 :> 0 @@ 4 :> 0), active |-> (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> FALSE @@ 4 :> FALSE), counter |-> (0 :> 0 @@ 1 :> 0 @@ 2 :> 0 @@ 3 :> 0 @@ 4 :> 0), token |-> [color |-> \"white\", pos |-> 4, q |-> 0]]>>", *variable.Result)
	h.unsetBreakpoints()
	h.continueFrames()
}
