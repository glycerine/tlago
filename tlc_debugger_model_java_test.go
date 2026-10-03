/*******************************************************************************
 * Copyright (c) 2020 Microsoft Research. All rights reserved.
 * Copyright (c) 2023, Oracle and/or its affiliates.
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
// Source: test/tlc2/debug/TLCDebuggerTestCase.java. This harness translates the
// existing model tests' synchronization and inherited frame assertions.
package tlago

import (
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/tlago/tlc"
)

type javaDebuggerModel struct {
	t          *testing.T
	debugger   *tlc.TLCDebugger
	done       chan struct{}
	result     *tlc.Result
	prior      *tlc.TLCDebugger
	exitStatus int
}

func startJavaDebuggerModel(t *testing.T, folder, model string, exitStatus int, extraArgs ...string) *javaDebuggerModel {
	t.Helper()
	setJavaModelLivenessThreshold(t, math.MaxFloat64)
	h := &javaDebuggerModel{t: t, debugger: tlc.NewTLCDebugger(nil), done: make(chan struct{}), prior: tlc.TLCDebuggerFactoryOverride, exitStatus: exitStatus}
	tlc.TLCDebuggerFactoryOverride = h.debugger
	go func() {
		defer close(h.done)
		h.result = runJavaTLCModelTestWithArguments(t, folder, model, func(meta, traceDirectory string) []string {
			return append([]string{"-metadir", meta, "-deadlock", "-fp", "0", "-seed", "1", "-workers", "1", "-checkpoint", "0", "-dumpTrace", "json", filepath.Join(traceDirectory, "tlc2.debug."+folder+"DebuggerTest.json"), "-debugger", "-noGenerateSpecTE"}, extraArgs...)
		})
	}()
	defer func() {
		if t.Failed() {
			h.close()
		}
	}()
	h.awaitStop()
	return h
}

func (h *javaDebuggerModel) awaitStop() {
	h.t.Helper()
	select {
	case <-h.debugger.StoppedEvents():
	case <-h.done:
		h.t.Fatalf("TLC terminated before the expected debugger stop: %v", h.result)
	case <-time.After(30 * time.Second):
		h.t.Fatal("timed out awaiting debugger stop")
	}
}

func (h *javaDebuggerModel) frames() []*tlc.TLCDebuggerFrame {
	h.t.Helper()
	response := h.debugger.StackTrace(tlc.TLCStackTraceArguments{})
	frames := make([]*tlc.TLCDebuggerFrame, len(response.StackFrames))
	for i, base := range response.StackFrames {
		for _, f := range h.debugger.Stack {
			if f.Base == base {
				frames[i] = f
				break
			}
		}
		if frames[i] == nil {
			h.t.Fatal("stack frame has no concrete debugger frame")
		}
	}
	return frames
}

func (h *javaDebuggerModel) stepIn(count ...int) []*tlc.TLCDebuggerFrame {
	n := 1
	if len(count) > 0 {
		n = count[0]
	}
	for i := 0; i < n; i++ {
		h.debugger.StepInCommand()
		h.awaitStop()
	}
	return h.frames()
}
func (h *javaDebuggerModel) stepOut() []*tlc.TLCDebuggerFrame {
	h.debugger.StepOutCommand()
	h.awaitStop()
	return h.frames()
}
func (h *javaDebuggerModel) next() []*tlc.TLCDebuggerFrame {
	h.debugger.StepOverCommand()
	h.awaitStop()
	return h.frames()
}
func (h *javaDebuggerModel) continueFrames(count ...int) []*tlc.TLCDebuggerFrame {
	n := 1
	if len(count) > 0 {
		n = count[0]
	}
	for i := 0; i < n; i++ {
		h.debugger.ContinueCommand()
		select {
		case <-h.debugger.StoppedEvents():
		case <-h.done:
			return nil
		case <-time.After(30 * time.Second):
			h.t.Fatal("timed out awaiting debugger continue")
		}
	}
	return h.frames()
}
func (h *javaDebuggerModel) close() {
	defer func() { tlc.TLCDebuggerFactoryOverride = h.prior }()
	// ModelCheckerTestCase.actualExitStatus starts at -1. Error debugger tests
	// assert that sentinel while TLC is still paused, before cleanup resumes it.
	if !h.t.Failed() && h.exitStatus == -1 {
		select {
		case <-h.done:
			h.t.Errorf("debugger model completed before the source's pending-exit assertion: %v", h.result)
		default:
		}
	}
	h.debugger.DisconnectCommand()
	select {
	case <-h.done:
	case <-time.After(30 * time.Second):
		h.t.Fatal("TLC did not terminate after debugger disconnect")
	}
	if !h.t.Failed() && h.exitStatus != -1 {
		if h.result == nil || h.result.ExitStatus != h.exitStatus {
			h.t.Fatalf("debugger model result=%v, want exit status %d", h.result, h.exitStatus)
		}
	}
}
func (h *javaDebuggerModel) unsetBreakpoints() {
	for module := range h.debugger.Breakpoints.All() {
		h.debugger.SetBreakpoints(module, nil)
	}
	h.debugger.SetExceptionBreakpoints([]tlc.TLCExceptionBreakpointFilterOption{{FilterID: tlc.TLCInvariantBreakpointsFilter}, {FilterID: tlc.TLCExceptionBreakpointsFilter}})
}
func (h *javaDebuggerModel) replaceBreakpoints(module string, line int) {
	h.unsetBreakpoints()
	h.setBreakpoints(debuggerBreakpoint(module, line))
}
func (h *javaDebuggerModel) setSpecBreakpoint() {
	h.debugger.SetExceptionBreakpoints([]tlc.TLCExceptionBreakpointFilterOption{{FilterID: tlc.TLCInvariantBreakpointsFilter}, {FilterID: tlc.TLCExceptionBreakpointsFilter}, {FilterID: tlc.TLCSpecBreakpointsFilter}})
}

type javaDebuggerBreakpoints struct {
	module   string
	requests []tlc.TLCSourceBreakpointRequest
}

func debuggerBreakpoint(module string, line int, options ...any) javaDebuggerBreakpoints {
	col := 0
	hit := -1
	condition := ""
	for i, option := range options {
		switch v := option.(type) {
		case int:
			if i == 0 {
				col = v
			} else {
				hit = v
			}
		case string:
			condition = v
		}
	}
	request := tlc.TLCSourceBreakpointRequest{Line: line, Column: &col, Condition: condition}
	if hit >= 0 {
		request.HitCondition = strconv.Itoa(hit)
	}
	return javaDebuggerBreakpoints{module, []tlc.TLCSourceBreakpointRequest{request}}
}
func (h *javaDebuggerModel) setBreakpoints(bp javaDebuggerBreakpoints) {
	h.debugger.SetBreakpoints(bp.module, bp.requests)
}
func (h *javaDebuggerModel) sourcePath(f *tlc.TLCDebuggerFrame) string {
	syntax := f.Base.Node.(interface{ GetTreeNode() any }).GetTreeNode().(*SanySyntaxNode)
	return syntax.Range.Begin.File
}
func debuggerAssertTrue(t *testing.T, yes bool) {
	t.Helper()
	if !yes {
		t.Fatal("Java assertion failed")
	}
}
func debuggerAssertEqual(t *testing.T, want, got any) {
	t.Helper()
	if expected, ok := want.(tlc.Value); ok {
		actual, ok := got.(tlc.Value)
		if !ok {
			t.Fatalf("value=%T, want TLC Value", got)
		}
		equal, err := expected.Equal(actual)
		if err != nil || !equal {
			t.Fatalf("value=%v, want %v (error=%v)", got, want, err)
		}
	} else if !reflect.DeepEqual(want, got) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
func debuggerScope(t *testing.T, scopes []tlc.TLCScope, name string) tlc.TLCScope {
	t.Helper()
	for _, scope := range scopes {
		if scope.Name == name {
			return scope
		}
	}
	t.Fatalf("missing %s scope", name)
	return tlc.TLCScope{}
}

// Preserve the inherited assertion forms used by Echo. Source's Context
// comparator checks binding values, deliberately ignoring their symbol names.
func assertDebuggerContext(t *testing.T, expected, actual *tlc.Context) {
	t.Helper()
	for expected != nil && expected != tlc.EmptyContext {
		if actual == nil || actual == tlc.EmptyContext {
			t.Fatal("context ended before expected bindings")
		}
		debuggerAssertEqual(t, expected.Value(), actual.Value())
		expected = expected.Next()
		actual = actual.Next()
	}
	debuggerAssertTrue(t, actual == nil || actual == tlc.EmptyContext)
}
func debuggerFrameArguments(t *testing.T, f *tlc.TLCDebuggerFrame, args []any) []any {
	t.Helper()
	if f == nil || f.Base == nil {
		t.Fatal("nil stack frame")
	}
	location := f.Base.Node.(interface{ GetSourceLocation() tlc.SourceLocation }).GetSourceLocation()
	begin := args[0].(int)
	_, short := args[2].(string)
	end := 0
	module := ""
	remaining := []any{}
	if short {
		end = args[1].(int)
		module = args[2].(string)
		remaining = args[3:]
	} else {
		end = args[2].(int)
		module = args[4].(string)
		remaining = args[5:]
		debuggerAssertEqual(t, args[1].(int), location.BeginColumn)
		debuggerAssertEqual(t, args[3].(int)+1, location.EndColumn+1)
	}
	debuggerAssertEqual(t, begin, location.BeginLine)
	debuggerAssertEqual(t, end, location.EndLine)
	debuggerAssertEqual(t, module, location.Source)
	debuggerAssertTrue(t, f.Base.Tool != nil)
	var expected *tlc.Context = tlc.EmptyContext
	if len(remaining) > 0 {
		if c, ok := remaining[0].(*tlc.Context); ok {
			expected = c
		} else if _, ok := remaining[0].(map[string]string); ok || remaining[0] == nil {
			expected = nil
		} else if _, ok := remaining[0].(javaDebuggerVariableSet); ok {
			expected = nil
		}
	}
	found := false
	for _, scope := range f.GetScopes() {
		if scope.Name == "Context" {
			found = true
		}
	}
	if expected == tlc.EmptyContext {
		debuggerAssertTrue(t, !found)
	} else {
		debuggerAssertTrue(t, found)
		variables := f.GetVariables(f.Base.ContextID, nil)
		if expected != nil {
			assertDebuggerContext(t, expected, f.Base.Context)
			debuggerAssertEqual(t, f.Base.Context.ToMap().Len(), len(variables))
		}
	}
	return remaining
}
func assertTLCFrame(t *testing.T, f *tlc.TLCDebuggerFrame, args ...any) {
	t.Helper()
	debuggerFrameArguments(t, f, args)
	debuggerAssertTrue(t, f.State == nil && f.Action == nil)
}
func assertTLCStateFrame(t *testing.T, f *tlc.TLCDebuggerFrame, args ...any) {
	t.Helper()
	remaining := debuggerFrameArguments(t, f, args)
	debuggerAssertTrue(t, f.State != nil && f.Action == nil)
	if len(remaining) > 0 {
		if expected, ok := remaining[0].(map[string]string); ok {
			assertDebuggerContextVariables(t, f, expected)
			return
		}
	}
	debuggerAssertState(t, f, remaining)
}

// Source's Map overload compares the displayed bindings and checks that
// GetVariables leaves each captured LazyValue cache count unchanged.
func assertDebuggerContextVariables(t *testing.T, f *tlc.TLCDebuggerFrame, expected map[string]string) {
	t.Helper()
	var lazies []*tlc.LazyValue
	var counts []int
	for context := f.Base.Context; context != nil; context = context.Next() {
		if lazy, ok := context.Value().(*tlc.LazyValue); ok {
			lazies = append(lazies, lazy)
			counts = append(counts, lazy.CacheCount)
		}
	}
	variables := f.GetVariables(f.Base.ContextID, nil)
	debuggerAssertEqual(t, len(expected), len(variables))
	for _, variable := range variables {
		value, exists := expected[variable.Name]
		debuggerAssertTrue(t, exists)
		debuggerAssertEqual(t, value, variable.Value)
	}
	for i, lazy := range lazies {
		debuggerAssertEqual(t, counts[i], lazy.CacheCount)
	}
}
func assertTLCActionFrame(t *testing.T, f *tlc.TLCDebuggerFrame, args ...any) {
	t.Helper()
	remaining := debuggerFrameArguments(t, f, args)
	debuggerAssertTrue(t, f.Action != nil)
	if len(remaining) > 0 {
		if expected, ok := remaining[0].(javaDebuggerVariableSet); ok {
			assertDebuggerContextVariableSet(t, f, expected)
			return
		}
	}
	debuggerScope(t, f.GetScopes(), "Action")
	predecessor := f.Action.GetS()
	debuggerAssertTrue(t, predecessor != nil && predecessor.AllAssigned())
	debuggerAssertState(t, f, remaining)
	debuggerAssertTrue(t, f.Action.State.GetAction() != nil && f.Action.State.Predecessor() != nil)
	stuttering := false
	if predecessor.AllAssigned() && f.Action.GetT().AllAssigned() {
		stuttering = predecessor.Equal(f.Action.GetT())
	}
	want := predecessor.Level() + 1
	if stuttering {
		want = predecessor.Level()
	}
	debuggerAssertEqual(t, want, f.Action.State.Level())
}

// Source TLCDebuggerTestCase's Set<Variable> overload checks displayed name,
// value and type, plus cache counts, without invoking the Context overload's
// state/trace assertions. Its unassigned parameter is intentionally unused.
type javaDebuggerVariableSet []*tlc.DebugTLCVariable

func assertDebuggerContextVariableSet(t *testing.T, f *tlc.TLCDebuggerFrame, expected javaDebuggerVariableSet) {
	t.Helper()
	var lazies []*tlc.LazyValue
	var counts []int
	for context := f.Base.Context; context != nil; context = context.Next() {
		if lazy, ok := context.Value().(*tlc.LazyValue); ok {
			lazies = append(lazies, lazy)
			counts = append(counts, lazy.CacheCount)
		}
	}
	variables := f.GetVariables(f.Base.ContextID, nil)
	debuggerAssertEqual(t, f.Base.Context.ToMap().Len(), len(variables))
	for _, variable := range variables {
		found := false
		for _, e := range expected {
			if variable.Name == e.Name && variable.Value == e.Value && variable.Type == e.Type {
				found = true
				break
			}
		}
		debuggerAssertTrue(t, found)
	}
	for i, lazy := range lazies {
		debuggerAssertEqual(t, counts[i], lazy.CacheCount)
	}
}
func debuggerAssertState(t *testing.T, f *tlc.TLCDebuggerFrame, remaining []any) {
	t.Helper()
	debuggerAssertTrue(t, f.State.State != nil)
	offset := 0
	if len(remaining) > 0 {
		if remaining[0] == nil {
			offset = 1
		} else if c, ok := remaining[0].(*tlc.Context); ok {
			offset = 1
			if c != nil {
				assertDebuggerContext(t, c, f.Base.Context)
			}
			if c == tlc.EmptyContext {
				debuggerAssertEqual(t, 0, f.Base.NestedVariables.Len())
			}
		}
	} else {
		debuggerAssertEqual(t, 0, f.Base.NestedVariables.Len())
	}
	wantUnassigned := map[string]bool{}
	for _, variable := range remaining[offset:] {
		if variables, ok := variable.([]tlc.StateVariable); ok {
			for _, v := range variables {
				wantUnassigned[v.Name.String()] = true
			}
		} else {
			wantUnassigned[variable.(tlc.StateVariable).Name.String()] = true
		}
	}
	gotUnassigned := map[string]bool{}
	for _, variable := range f.State.State.Unassigned() {
		gotUnassigned[variable.Name.String()] = true
	}
	debuggerAssertEqual(t, wantUnassigned, gotUnassigned)
	// These two source helpers restore nested references after their queries.
	old := tlc.NewInsMap[int, *tlc.DebugTLCVariable]()
	for ref, variable := range f.Base.NestedVariables.All() {
		old.Set(ref, variable)
	}
	func() {
		defer func() { f.Base.NestedVariables = old }()
		var variables []*tlc.DebugTLCVariable
		if f.Next != nil {
			variables = f.Next.GetStateVariables(nil)
		} else {
			variables = f.GetVariables(f.State.StateID, nil)
		}
		debuggerAssertEqual(t, 1, len(variables))
		expected := tlc.NewRecordValueFromState(f.State.State, tlc.DebuggerNotEvaluatedValue)
		if f.Action != nil {
			expected = tlc.NewRecordValueFromStates(f.Action.GetS(), f.Action.GetT(), tlc.DebuggerNotEvaluatedValue)
		}
		debuggerAssertEqual(t, expected, variables[0].TLCValue)
	}()
	old = tlc.NewInsMap[int, *tlc.DebugTLCVariable]()
	for ref, variable := range f.Base.NestedVariables.All() {
		old.Set(ref, variable)
	}
	func() {
		defer func() { f.Base.NestedVariables = old }()
		trace := f.GetVariables(f.State.StateID+1, nil)
		debuggerAssertTrue(t, len(trace) > 0)
		debuggerAssertTrue(t, strings.HasPrefix(strings.TrimLeft(trace[len(trace)-1].Name, "0"), "1: "))
		for i := len(trace) - 1; i > 0; i-- {
			ordinal := len(trace) - i
			debuggerAssertTrue(t, strings.HasPrefix(strings.TrimLeft(trace[i].Name, "0"), fmt.Sprintf("%d:", ordinal)))
			record, ok := trace[i].TLCValue.(*tlc.RecordValue)
			debuggerAssertTrue(t, ok)
			for _, value := range record.Values {
				_, pending := value.(*tlc.DebuggerValue)
				debuggerAssertTrue(t, !pending)
			}
		}
	}()
}
func assertTLCSyntheticStateStackFrame(t *testing.T, f *tlc.TLCDebuggerFrame, args ...any) {
	t.Helper()
	assertTLCStateFrame(t, f, args[:len(args)-1]...)
	debuggerAssertTrue(t, f.Synthetic != nil)
	debuggerAssertEqual(t, args[len(args)-1].(int), f.State.State.Level())
	debuggerAssertTrue(t, f.State.State.AllAssigned())
}
func assertTLCNextStatesFrame(t *testing.T, f *tlc.TLCDebuggerFrame, args ...any) {
	t.Helper()
	assertTLCStateFrame(t, f, args[:len(args)-1]...)
	debuggerAssertTrue(t, f.Next != nil)
	expected := args[len(args)-1].(int)
	debuggerAssertEqual(t, expected, f.Next.GetSuccessors().Size())
	if expected == 0 {
		for _, scope := range f.GetScopes() {
			debuggerAssertTrue(t, scope.Name != "Successors")
		}
		return
	}
	scope := debuggerScope(t, f.GetScopes(), "Successors")
	variables := f.GetVariables(scope.VariablesReference, nil)
	debuggerAssertEqual(t, expected, len(variables))
	for _, variable := range variables {
		found := false
		for _, state := range f.Next.GetSuccessors().ToSlice() {
			equal, err := tlc.NewRecordValueFromState(state).Equal(variable.TLCValue)
			if err == nil && equal {
				found = true
				break
			}
		}
		debuggerAssertTrue(t, found)
	}
}

func assertTLCInitStatesFrame(t *testing.T, f *tlc.TLCDebuggerFrame, args ...any) {
	t.Helper()
	assertTLCFrame(t, f, args[:len(args)-1]...)
	debuggerAssertTrue(t, f.Init != nil)
	// Source TODO: expectedSuccessors is unused; no initial-state assertions.
}
