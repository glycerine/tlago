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
// Translation of the complete EWD998ChanDebuggerTest.testSpec.
package tlago

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaEWD998ChanDebugger(t *testing.T) {
	const UTILS, FOLDER, RM, MDL = "Utils", "EWD998", "EWD998Chan", "EWD998Chan"
	h := startJavaDebuggerModel(t, FOLDER, MDL)
	defer h.close()
	// These adapters retain the source's nullable String and protocol requests.
	str := func(s string) *string { return &s }
	replace := func(module string, line int, condition ...string) []tlc.TLCBreakpoint {
		h.unsetBreakpoints()
		bp := debuggerBreakpoint(module, line)
		if len(condition) > 0 {
			bp.requests[0].Condition = condition[0]
		}
		return h.debugger.SetBreakpoints(bp.module, bp.requests)
	}
	hover := func(module, symbol string, bl, bc, el, ec int) tlc.TLCEvaluateResponse {
		path, err := filepath.Abs(module)
		if err != nil {
			t.Fatal(err)
		}
		uri := url.URL{Scheme: "tlaplus", Path: filepath.ToSlash(path), RawQuery: symbol, Fragment: fmt.Sprintf("%d %d %d %d", bl, bc, el, ec)}
		expression := "tlaplus://" + strings.TrimPrefix(uri.String(), "tlaplus:")
		frameID := h.frames()[0].Base.ID
		return h.debugger.Evaluate(tlc.TLCEvaluateArguments{Context: "hover", Expression: &expression, FrameID: &frameID})
	}
	setType := "SetEnumValue: " + tlc.EmptySet.KindString()
	intType := "IntValue: " + tlc.IntZero.KindString()
	boolType := "BoolValue: " + tlc.BoolFalse.KindString()
	tupleType := "TupleValue: " + tlc.EmptyTuple.KindString()
	createVariable := func(name, value, typ string) *tlc.DebugTLCVariable {
		return &tlc.DebugTLCVariable{Name: name, Value: value, Type: typ}
	}

	stackFrames := h.frames()
	// ASSUME in EWD998Chan.
	debuggerAssertEqual(t, 1, len(stackFrames))
	assertTLCFrame(t, stackFrames[0], 11, 11, RM)
	debuggerAssertTrue(t, strings.HasSuffix(strings.ReplaceAll(h.sourcePath(stackFrames[0]), "\\", "/"), "test_vectors/models/EWD998/EWD998Chan.tla"))
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 2, len(stackFrames))
	assertTLCFrame(t, stackFrames[1], 11, 11, RM)
	assertTLCFrame(t, stackFrames[0], 11, 11, RM)
	stackFrame := stackFrames[1]
	constants := stackFrame.Base.GetConstants(nil)
	debuggerAssertEqual(t, 3, len(constants))

	// Watch expressions preserve null and the source's escaped error display.
	debuggerAssertEqual(t, tlc.TLCEvaluateResponse{}, stackFrame.Evaluate(nil))
	debuggerAssertEqual(t, "Syntax error while parsing breakpoint expression \\\"Does not exist\\\"", *stackFrame.Evaluate(str("Does not exist")).Result)
	debuggerAssertEqual(t, "In evaluation, the identifier counter is either undefined or not an operator.\\nline 43, col 6 to line 43, col 12 of module EWD998Chan", *stackFrame.Evaluate(str("Init")).Result)
	debuggerAssertEqual(t, "In evaluation, the identifier inbox is either undefined or not an operator.\\nline 53, col 22 to line 53, col 26 of module EWD998Chan", *stackFrame.Evaluate(str("InitiateProbe")).Result)
	debuggerAssertEqual(t, "In evaluation, the identifier inbox is either undefined or not an operator.\\nline 53, col 22 to line 53, col 26 of module EWD998Chan", *stackFrame.Evaluate(str("System")).Result)

	// High-level spec constants, ordered lexicographically.
	consts := stackFrame.GetVariables(constants[0].VariablesReference, nil)
	debuggerAssertEqual(t, 3, len(consts))
	debuggerAssertEqual(t, "Color", consts[0].Name)
	debuggerAssertEqual(t, setType, consts[0].Type)
	debuggerAssertEqual(t, "{\"white\", \"black\"}", consts[0].Value)
	debuggerAssertEqual(t, 2, consts[0].TLCValue.(*tlc.SetEnumValue).Elems.Len())
	debuggerAssertEqual(t, 2, len(stackFrame.GetVariables(consts[0].VariablesReference, nil)))
	debuggerAssertEqual(t, "\"black\"", stackFrame.GetVariables(consts[0].VariablesReference, nil)[0].Value)
	debuggerAssertEqual(t, "\"white\"", stackFrame.GetVariables(consts[0].VariablesReference, nil)[1].Value)
	debuggerAssertEqual(t, "Nodes", consts[1].Name)
	debuggerAssertEqual(t, setType, consts[1].Type)
	debuggerAssertEqual(t, "{0, 1, 2}", consts[1].Value)
	debuggerAssertEqual(t, 3, consts[1].TLCValue.(*tlc.SetEnumValue).Elems.Len())
	debuggerAssertEqual(t, "Token", consts[2].Name)
	debuggerAssertEqual(t, "SetOfRcdsValue: a set of the form [d1 : S1, ... , dN : SN]", consts[2].Type)
	debuggerAssertEqual(t, "[color: {\"white\", \"black\"}, q: Int, pos: 0..2]", consts[2].Value)
	debuggerAssertEqual(t, 3, len(consts[2].TLCValue.(*tlc.SetOfRcdsValue).Names))
	// Source TODO: no expansion if a field has an infinite domain.
	consts = stackFrame.GetVariables(consts[2].VariablesReference, nil)
	debuggerAssertEqual(t, 0, len(consts))

	// Low-level spec constants.
	consts = stackFrame.GetVariables(constants[1].VariablesReference, nil)
	debuggerAssertEqual(t, 6, len(consts))
	debuggerAssertEqual(t, "BasicMsg", consts[0].Name)
	debuggerAssertEqual(t, "SetOfRcdsValue: a set of the form [d1 : S1, ... , dN : SN]", consts[0].Type)
	debuggerAssertEqual(t, "{[type |-> \"pl\"]}", consts[0].Value)
	debuggerAssertEqual(t, "Color", consts[1].Name)
	debuggerAssertEqual(t, setType, consts[1].Type)
	debuggerAssertEqual(t, "{\"white\", \"black\"}", consts[1].Value)
	debuggerAssertEqual(t, "Message", consts[2].Name)
	debuggerAssertEqual(t, "SetCupValue: a set of the form S \\cup T", consts[2].Type)
	debuggerAssertEqual(t, "[color: {\"white\", \"black\"}, type: {\"tok\"}, q: Int] \\cup {[type |-> \"pl\"]}", consts[2].Value)
	debuggerAssertEqual(t, "N", consts[3].Name)
	debuggerAssertEqual(t, intType, consts[3].Type)
	debuggerAssertEqual(t, "3", consts[3].Value)
	debuggerAssertEqual(t, "Nodes", consts[4].Name)
	debuggerAssertEqual(t, setType, consts[4].Type)
	debuggerAssertEqual(t, "{0, 1, 2}", consts[4].Value)
	debuggerAssertEqual(t, "TokenMsg", consts[5].Name)
	debuggerAssertEqual(t, "SetOfRcdsValue: a set of the form [d1 : S1, ... , dN : SN]", consts[5].Type)
	debuggerAssertEqual(t, "[color: {\"white\", \"black\"}, type: {\"tok\"}, q: Int]", consts[5].Value)

	replace(RM, 12)
	stackFrames = h.continueFrames()
	ea := tlc.TLCEvaluateArguments{Context: "repl", Expression: str("x+1"), FrameID: &stackFrames[0].Base.ID}
	debuggerAssertEqual(t, "3", *h.debugger.Evaluate(ea).Result)
	replace(UTILS, 13)
	stackFrames = h.continueFrames()
	ea.FrameID = &stackFrames[0].Base.ID
	ea.Expression = str("i")
	debuggerAssertEqual(t, "2", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("op(6,7)")
	debuggerAssertEqual(t, "13", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("reduced")
	debuggerAssertEqual(t, "(0 :> 0 @@ 1 :> 1 @@ 2 :> 3)", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("reduced[i-1]")
	debuggerAssertEqual(t, "1", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("fun")
	debuggerAssertEqual(t, "(0 :> 0 @@ 1 :> 1 @@ 2 :> 2)", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("fun[i]")
	debuggerAssertEqual(t, "2", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("op(reduced[i - 1], fun[i])")
	debuggerAssertEqual(t, "3", *h.debugger.Evaluate(ea).Result)
	stackFrames = h.stepIn()
	ea.Expression = str("<<b, a>>")
	ea.FrameID = &stackFrames[0].Base.ID
	debuggerAssertEqual(t, "<<2, 1>>", *h.debugger.Evaluate(ea).Result)

	// Retain the complete original set of valid breakpoint lines.
	lines := map[int]bool{}
	for _, line := range []int{11, 13, 14, 15, 24, 27, 28, 29, 30, 31, 36, 37, 39, 40, 41, 45, 47, 48, 49, 50, 51, 53, 57, 58, 59, 60, 61, 62, 64, 66, 67, 73, 75, 77, 80, 83, 84, 86, 88, 90, 91, 94, 95, 96, 98, 103, 105, 113, 114, 123, 124, 125, 126, 127, 128, 133, 140, 141, 144, 150, 153, 154, 155, 156, 158, 160, 162, 168, 171} {
		lines[line] = true
	}
	for line := 1; line <= 171; line++ {
		if lines[line] {
			debuggerAssertTrue(t, replace(FOLDER, line)[0].Verified)
		}
	}
	for line := 1; line <= 171; line++ {
		if !lines[line] {
			debuggerAssertTrue(t, !replace(FOLDER, line)[0].Verified)
		}
	}
	debuggerAssertTrue(t, replace(RM, 102)[0].Verified)
	debuggerAssertTrue(t, replace(RM, 102, "")[0].Verified)
	breakpoint := replace(RM, 102, "garbled input not parsable")[0]
	debuggerAssertTrue(t, !breakpoint.Verified)
	debuggerAssertEqual(t, "Syntax error while parsing breakpoint expression \"garbled input not parsable\"", breakpoint.Message)
	breakpoint = replace(RM, 102, "LET T == INSTANCE DoesNotExist IN T!YOLO")[0]
	debuggerAssertTrue(t, !breakpoint.Verified)
	debuggerAssertEqual(t, "line 103, col 0 to line 102, col 1 of module EWD998Chan\n\nSemantic error while parsing breakpoint expression \"LET T == INSTANCE DoesNotExist IN T!YOLO\"", breakpoint.Message)
	debuggerAssertTrue(t, !replace(RM, 103)[0].Verified)
	debuggerAssertTrue(t, replace(RM, 104)[0].Verified)
	debuggerAssertTrue(t, !replace(RM, 105)[0].Verified)
	debuggerAssertTrue(t, !replace(RM, 105)[0].Verified)
	debuggerAssertTrue(t, replace(RM, 107)[0].Verified)

	vars := tlc.StateVariables()
	replace(RM, 49)
	stackFrames = h.continueFrames()
	assertTLCStateFrame(t, stackFrames[0], 49, 49, RM, vars[1])
	// Refinement mapping LazyValues must not make the debugger debug itself.
	replace(UTILS, 13)
	stackFrames = h.continueFrames()
	i := 19
	debuggerAssertEqual(t, i, len(stackFrames))
	i--
	pop := func() *tlc.TLCDebuggerFrame { i--; return stackFrames[i] }
	assertTLCFrame(t, pop(), 43, 49, RM)
	assertTLCStateFrame(t, pop(), 43, 49, RM, vars)
	assertTLCStateFrame(t, pop(), 43, 43, RM, vars)
	assertTLCStateFrame(t, pop(), 44, 46, RM, vars[0], vars[1], vars[3])
	assertTLCStateFrame(t, pop(), 48, 48, RM, vars[0], vars[1])
	assertTLCStateFrame(t, pop(), 49, 49, RM, vars[1])
	assertTLCStateFrame(t, pop(), 49, 49, RM)
	assertTLCStateFrame(t, pop(), 187, 187, RM)
	assertTLCStateFrame(t, pop(), 166, 181, RM)
	allVariables := map[string]string{}
	allVariables["pending"] = "(0 :> 0 @@ 1 :> 0 @@ 2 :> 0)"
	allVariables["token"] = "[pos |-> 0, q |-> 0, color |-> \"black\"]"
	allVariables["counter"] = "(0 :> 0 @@ 1 :> 0 @@ 2 :> 0)"
	allVariables["N"] = "3"
	allVariables["active"] = "(0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE)"
	allVariables["color"] = "(0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\")"
	assertTLCStateFrame(t, pop(), 150, 162, FOLDER, allVariables)
	assertTLCStateFrame(t, pop(), 150, 150, FOLDER, allVariables)
	assertTLCStateFrame(t, pop(), 150, 150, FOLDER, allVariables)
	assertTLCStateFrame(t, pop(), 150, 150, FOLDER, allVariables)
	assertTLCStateFrame(t, pop(), 133, 133, FOLDER, allVariables)
	allVariables["op"] = "sum(a,b) == a+b"
	allVariables["fun"] = "(0 :> 0 @@ 1 :> 0 @@ 2 :> 0)"
	allVariables["from"] = "0"
	allVariables["to"] = "2"
	allVariables["base"] = "0"
	assertTLCStateFrame(t, pop(), 11, 14, UTILS, allVariables)
	allVariables["reduced"] = "(0 :> 0 @@ 1 :> 0 @@ 2 :> 0)"
	assertTLCStateFrame(t, pop(), 14, 14, UTILS, allVariables)
	allVariables["i"] = "2"
	assertTLCStateFrame(t, pop(), 12, 13, UTILS, allVariables)
	assertTLCStateFrame(t, pop(), 13, 13, UTILS, allVariables)

	replace(RM, 63)
	stackFrames = h.continueFrames()
	assertTLCActionFrame(t, stackFrames[0], 63, 67, RM, tlc.EmptyContext.Cons(nil, tlc.IntOne), vars[0], vars[1], vars[2], vars[3])
	ea = tlc.TLCEvaluateArguments{Context: "repl", Expression: str("j"), FrameID: &stackFrames[0].Base.ID}
	debuggerAssertEqual(t, "1", *h.debugger.Evaluate(ea).Result)

	// Stack variables in InitiateProbe.
	replace(RM, 71)
	stackFrames = h.continueFrames()
	stackVariables := stackFrames[5].Base.GetVariables(stackFrames[5].Base.GetStackID(), nil)
	debuggerAssertEqual(t, 1, len(stackVariables))
	debuggerAssertEqual(t, "TRUE", stackVariables[0].Value)
	debuggerAssertEqual(t, "inbox[0][j].type=\"tok\"", stackVariables[0].Name)
	debuggerAssertEqual(t, boolType, stackVariables[0].Type)
	stackVariables = stackFrames[3].Base.GetVariables(stackFrames[3].Base.GetStackID(), nil)
	debuggerAssertEqual(t, 2, len(stackVariables))
	debuggerAssertEqual(t, "TRUE", stackVariables[1].Value)
	debuggerAssertEqual(t, "inbox[0][j].type=\"tok\"", stackVariables[1].Name)
	debuggerAssertEqual(t, boolType, stackVariables[1].Type)
	debuggerAssertEqual(t, "TRUE", stackVariables[0].Value)
	debuggerAssertEqual(t, "inbox[0][j].color=\"black\"", stackVariables[0].Name)
	debuggerAssertEqual(t, boolType, stackVariables[0].Type)
	// Stack variables in PassToken.
	replace(RM, 85)
	stackFrames = h.continueFrames()
	stackVariables = stackFrames[6].Base.GetVariables(stackFrames[6].Base.GetStackID(), nil)
	debuggerAssertEqual(t, 1, len(stackVariables))
	debuggerAssertEqual(t, "TRUE", stackVariables[0].Value)
	debuggerAssertEqual(t, "~active[i]", stackVariables[0].Name)
	debuggerAssertEqual(t, boolType, stackVariables[0].Type)
	stackVariables = stackFrames[3].Base.GetVariables(stackFrames[3].Base.GetStackID(), nil)
	debuggerAssertEqual(t, 2, len(stackVariables))
	debuggerAssertEqual(t, "TRUE", stackVariables[1].Value)
	debuggerAssertEqual(t, "~active[i]", stackVariables[1].Name)
	debuggerAssertEqual(t, boolType, stackVariables[1].Type)
	debuggerAssertEqual(t, "TRUE", stackVariables[0].Value)
	debuggerAssertEqual(t, "inbox[i][j].type=\"tok\"", stackVariables[0].Name)
	debuggerAssertEqual(t, boolType, stackVariables[0].Type)

	// Step through the complete original mildly complex expression.
	replace(RM, 119)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 10, len(stackFrames))
	context := tlc.EmptyContext.Cons(nil, tlc.IntOne).Cons(nil, tlc.IntOne).Cons(nil, tlc.IntOne)
	assertTLCActionFrame(t, stackFrames[0], 119, 119, RM, context, vars[3])
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 11, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 119, 119, RM, context, vars[3])
	stackFrames = h.stepIn(3)
	debuggerAssertEqual(t, 12, len(stackFrames))
	variables := javaDebuggerVariableSet{
		createVariable("i", "1", intType),
		createVariable("j", "1", intType),
		createVariable("@", "<<[type |-> \"pl\"]>>", tupleType),
	}
	assertTLCActionFrame(t, stackFrames[0], 119, 44, 119, 57, RM, variables, vars[3])
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 13, len(stackFrames))
	variables = append(variables, createVariable("s", "<<[type |-> \"pl\"]>>", tupleType))
	assertTLCActionFrame(t, stackFrames[0], 29, 29, UTILS, variables, vars[3])
	stackFrames = h.stepIn(13)
	debuggerAssertEqual(t, 11, len(stackFrames))
	variables = javaDebuggerVariableSet{createVariable("i", "1", intType), createVariable("j", "1", intType)}
	assertTLCActionFrame(t, stackFrames[0], 120, 6, 120, 19, RM, variables)

	replace(RM, 29)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 12, len(stackFrames))
	assertTLCStateFrame(t, stackFrames[0], 29, 3, 37, 25, RM, tlc.EmptyContext)
	replace(FOLDER, 150)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 12, len(stackFrames))
	assertTLCStateFrame(t, stackFrames[0], 150, 3, 162, 34, FOLDER, (*tlc.Context)(nil))

	// Hover/location resolution, retaining every source response assertion.
	replace(RM, 120)
	h.continueFrames()
	v := hover(RM, "inbox", 118, 14, 118, 18)
	debuggerAssertEqual(t, "FcnRcdValue: a function  of the form (d1 :> e1 @@ ... @@ dN :> eN)", *v.Type)
	debuggerAssertTrue(t, v.VariablesReference != 0)
	debuggerAssertEqual(t, "(0 :> <<[color |-> \"black\", type |-> \"tok\", q |-> 0]>> @@ 1 :> <<[type |-> \"pl\"]>> @@ 2 :> <<>>)", *v.Result)
	v = hover(RM, "inbox", 119, 24, 119, 28)
	debuggerAssertEqual(t, "FcnRcdValue: a function  of the form (d1 :> e1 @@ ... @@ dN :> eN)", *v.Type)
	debuggerAssertTrue(t, v.VariablesReference != 0)
	debuggerAssertEqual(t, "(0 :> <<[color |-> \"black\", type |-> \"tok\", q |-> 0]>> @@ 1 :> <<[type |-> \"pl\"]>> @@ 2 :> <<>>)", *v.Result)
	v = hover(RM, "inbox", 119, 14, 119, 19)
	debuggerAssertEqual(t, "FcnRcdValue: a function  of the form (d1 :> e1 @@ ... @@ dN :> eN)", *v.Type)
	debuggerAssertTrue(t, v.VariablesReference != 0)
	debuggerAssertEqual(t, "(0 :> <<[color |-> \"black\", type |-> \"tok\", q |-> 0]>> @@ 1 :> <<>> @@ 2 :> <<>>)", *v.Result)
	v = hover(RM, "i", 109, 9, 109, 9)
	debuggerAssertEqual(t, intType, *v.Type)
	debuggerAssertEqual(t, 0, v.VariablesReference)
	debuggerAssertEqual(t, "1", *v.Result)
	v = hover(RM, "i", 111, 35, 111, 35)
	debuggerAssertEqual(t, intType, *v.Type)
	debuggerAssertEqual(t, 0, v.VariablesReference)
	debuggerAssertEqual(t, "1", *v.Result)
	v = hover(RM, "i", 115, 34, 115, 34)
	debuggerAssertEqual(t, intType, *v.Type)
	debuggerAssertEqual(t, 0, v.VariablesReference)
	debuggerAssertEqual(t, "1", *v.Result)
	v = hover(RM, "j", 117, 9, 117, 9)
	debuggerAssertEqual(t, "FormalParamNode", *v.Type)
	debuggerAssertEqual(t, "line 117, col 9 to line 117, col 9 of module EWD998Chan", *v.Result)
	v = hover(RM, "j", 118, 23, 118, 23)
	debuggerAssertEqual(t, "OpApplNode", *v.Type)
	debuggerAssertEqual(t, "line 118, col 23 to line 118, col 23 of module EWD998Chan", *v.Result)
	v = hover(RM, "j", 119, 56, 119, 56)
	debuggerAssertEqual(t, "OpApplNode", *v.Type)
	debuggerAssertEqual(t, "line 119, col 56 to line 119, col 56 of module EWD998Chan", *v.Result)
	replace(RM, 119)
	h.continueFrames()
	v = hover(RM, "type", 118, 26, 118, 29)
	debuggerAssertEqual(t, "StringValue: a string", *v.Type)
	debuggerAssertEqual(t, "pl", *v.Result)
	replace(RM, 81)
	h.continueFrames()
	v = hover(RM, "type", 77, 26, 77, 29)
	debuggerAssertEqual(t, "StringValue: a string", *v.Type)
	debuggerAssertEqual(t, "tok", *v.Result)
	v = hover(RM, "type", 81, 72, 81, 72)
	debuggerAssertEqual(t, "IntValue: an integer", *v.Type)
	debuggerAssertEqual(t, "0", *v.Result)
	replace(RM, 85)
	h.continueFrames()
	v = hover(RM, "inbox", 80, 28, 80, 32)
	debuggerAssertEqual(t, "FcnRcdValue: a function  of the form (d1 :> e1 @@ ... @@ dN :> eN)", *v.Type)
	debuggerAssertTrue(t, v.VariablesReference != 0)
	debuggerAssertEqual(t, "(0 :> <<>> @@ 1 :> <<>> @@ 2 :> <<[color |-> \"white\", type |-> \"tok\", q |-> 0]>>)", *v.Result)
	v = hover(RM, "inbox", 80, 18, 80, 23)
	debuggerAssertEqual(t, tlc.DebuggerNotEvaluatedValue.String(), *v.Result)
	replace(RM, 179)
	h.continueFrames()
	v = hover(RM, "active", 166, 43, 166, 48)
	debuggerAssertEqual(t, "FcnRcdValue: a function  of the form (d1 :> e1 @@ ... @@ dN :> eN)", *v.Type)
	debuggerAssertTrue(t, v.VariablesReference != 0)
	debuggerAssertEqual(t, "(0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE)", *v.Result)
	v = hover(RM, "active", 166, 33, 166, 40)
	debuggerAssertEqual(t, "line 166, col 33 to line 166, col 40 of module EWD998Chan", *v.Result)

	h.unsetBreakpoints()
	sba := debuggerBreakpoint(RM, 107)
	sba.requests[0].HitCondition = "2"
	h.setBreakpoints(sba)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 8, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 107, 6, 107, 32, RM, (*tlc.Context)(nil), vars[0], vars[1])
	h.unsetBreakpoints()
	sba = debuggerBreakpoint(RM, 148)
	sba.requests[0].HitCondition = "3"
	h.setBreakpoints(sba)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 10, len(stackFrames))
	assertTLCStateFrame(t, stackFrames[0], 148, 3, 153, 45, RM, tlc.EmptyContext)
	h.unsetBreakpoints()
	h.setSpecBreakpoint()
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 3, len(stackFrames))
	assertTLCNextStatesFrame(t, stackFrames[0], 134, 20, 134, 23, RM, tlc.EmptyContext, 3)
	stackFrame = stackFrames[0]
	debuggerAssertEqual(t, tlc.TLCEvaluateResponse{}, stackFrame.Evaluate(nil))
	debuggerAssertEqual(t, "Syntax error while parsing breakpoint expression \\\"Does not exist\\\"", *stackFrame.Evaluate(str("Does not exist")).Result)
	debuggerAssertEqual(t, "FALSE", *stackFrame.Evaluate(str("Init")).Result)
	debuggerAssertEqual(t, "FALSE", *stackFrame.Evaluate(str("InitiateProbe")).Result)
	debuggerAssertEqual(t, "FALSE", *stackFrame.Evaluate(str("System")).Result)
	debuggerAssertEqual(t, "FALSE", *stackFrame.Evaluate(str("Environment")).Result)
	debuggerAssertEqual(t, "FALSE", *stackFrame.Evaluate(str("Next")).Result)
	debuggerAssertEqual(t, "FALSE", *stackFrame.Evaluate(str("Spec")).Result)
	debuggerAssertEqual(t, "TRUE", *stackFrame.Evaluate(str("StateConstraint")).Result)
	debuggerAssertEqual(t, "2", *stackFrame.Evaluate(str("tpos")).Result)
	debuggerAssertEqual(t, "TRUE", *stackFrame.Evaluate(str("Stop")).Result)
	debuggerAssertEqual(t, "TRUE", *stackFrame.Evaluate(str("ActionConstraint")).Result)
	er := stackFrame.Evaluate(str("EnabledAlias"))
	// Java's assertNotNull(er): Go's response is a concrete value.
	debuggerAssertEqual(t, "[InitiateProbe |-> FALSE, PassToken |-> TRUE, SendMsg |-> TRUE, RecvMsg |-> FALSE, Deactivate |-> TRUE]", *er.Result)
	debuggerAssertTrue(t, er.VariablesReference != 0)
	frob := stackFrame.GetVariables(er.VariablesReference, nil)
	debuggerAssertEqual(t, 5, len(frob))
	debuggerAssertEqual(t, "Deactivate", frob[0].Name)
	debuggerAssertEqual(t, "TRUE", frob[0].Value)
	debuggerAssertEqual(t, "InitiateProbe", frob[1].Name)
	debuggerAssertEqual(t, "FALSE", frob[1].Value)
	debuggerAssertEqual(t, "PassToken", frob[2].Name)
	debuggerAssertEqual(t, "TRUE", frob[2].Value)
	debuggerAssertEqual(t, "RecvMsg", frob[3].Name)
	debuggerAssertEqual(t, "FALSE", frob[3].Value)
	debuggerAssertEqual(t, "SendMsg", frob[4].Name)
	debuggerAssertEqual(t, "TRUE", frob[4].Value)
	debuggerAssertEqual(t, "<<FALSE, FALSE>>", *stackFrame.Evaluate(str("[i \\in Nodes \\ {0} |-> PassToken(i)]")).Result)
	debuggerAssertEqual(t, "<<FALSE, TRUE>>", *stackFrame.Evaluate(str("[i \\in Nodes \\ {0} |-> ENABLED PassToken(i)]")).Result)

	replace(RM, 142)
	stackFrames = h.continueFrames()
	ea = tlc.TLCEvaluateArguments{Context: "repl", Expression: str("ibx"), FrameID: &stackFrames[0].Base.ID}
	debuggerAssertEqual(t, "<<>>", *h.debugger.Evaluate(ea).Result)
	replace(RM, 84)
	stackFrames = h.continueFrames()
	ea = tlc.TLCEvaluateArguments{Context: "repl", Expression: str("@"), FrameID: &stackFrames[0].Base.ID}
	debuggerAssertEqual(t, "line 84, col 77 to line 84, col 85 of module EWD998Chan\\n\\nSemantic error while parsing breakpoint expression \\\"@\\\"", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("tpos")
	ea.FrameID = &stackFrames[0].Base.ID
	debuggerAssertEqual(t, "2", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("tkn.color")
	ea.FrameID = &stackFrames[0].Base.ID
	debuggerAssertEqual(t, "white", *h.debugger.Evaluate(ea).Result)
	ea.Expression = str("ccounter(0)+1")
	ea.FrameID = &stackFrames[0].Base.ID
	debuggerAssertEqual(t, "1", *h.debugger.Evaluate(ea).Result)

	// POSTCONDITION, then run the spec to completion.
	h.unsetBreakpoints()
	sba = debuggerBreakpoint(MDL, 212)
	h.setBreakpoints(sba)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 1, len(stackFrames))
	assertTLCFrame(t, stackFrames[0], 212, 9, 212, 68, MDL, tlc.EmptyContext)
	h.unsetBreakpoints()
	h.continueFrames()
}
