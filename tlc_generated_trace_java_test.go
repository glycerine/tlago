/*******************************************************************************
 * Copyright (c) 2020 Microsoft Research. All rights reserved.
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
// Ports of the five concrete TraceExpressionSpec tests and their inherited
// generated-tool assertions.
package tlago

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func runJavaGeneratedTraceSpec(t *testing.T, root, config, mode string, status int, deadlock bool, threshold float64) *tlc.Tool {
	t.Helper()
	setJavaModelLivenessThreshold(t, threshold)
	var generatedDirectory string
	result := runJavaTLCModelTestWithArguments(t, "TESpecTest", root, func(meta, traceDirectory string) []string {
		generatedDirectory = meta
		args := []string{"-metadir", meta}
		if !deadlock {
			args = append(args, "-deadlock")
		}
		return append(args, "-debugger", "nosuspend,port=4712,nohalt",
			"-generateSpecTE", "-teSpecOutDir", meta, "-fp", "0", "-seed", "1",
			"-workers", "1", "-checkpoint", "0", "-noGenerateSpecTEBin", "-config", config, mode)
	})
	if result.ExitStatus != status {
		t.Fatalf("exit status=%d, want %d; messages=%+v", result.ExitStatus, status, result.Messages)
	}
	module := tlc.DeriveTESpecModuleName(root, tlc.Globals.StartTime)
	classpath, err := tlcApplicationClasspath(nil)
	if err != nil {
		t.Fatal(err)
	}
	// The source resolver searches the original user directory and the metadir
	// containing the generated monolithic module. Retain the same interner and
	// runtime for generation and FastTool construction, as the Java test does.
	resolver := tlc.NewSimpleFilenameToStream([]string{generatedDirectory}, tlc.FilenameResolverOptions{Classpath: classpath})
	tool, diags, err := loadTLCAppTool(module, module+".tla", resolver, tlc.RuntimeParameters{})
	requireNoErrors(t, diags)
	if err != nil {
		t.Fatal(err)
	}
	return tool
}

func javaGeneratedInitialStates(t *testing.T, tool *tlc.Tool) *tlc.StateVec {
	t.Helper()
	states := tlc.NewStateVec(1)
	if err := tool.GetInitStates(tlc.NewStateFunctor(func(state *tlc.TLCStateMut) (any, error) {
		states.AddElement(state)
		return nil, nil
	})); err != nil {
		t.Fatal(err)
	}
	return states
}

func requireJavaGeneratedTraceState(t *testing.T, tool *tlc.Tool, states *tlc.StateVec, invariant *tlc.Action, valid bool) *tlc.TLCStateMut {
	t.Helper()
	if states.Size() != 1 {
		t.Fatalf("states=%d, want 1", states.Size())
	}
	state := states.First()
	if !tool.IsGoodState(state) {
		t.Fatal("trace state is not good")
	}
	inModel, err := tool.IsInModel(state)
	if err != nil || !inModel {
		t.Fatalf("in model=%t: %v", inModel, err)
	}
	if invariant != nil {
		actual, err := tool.IsValidState(invariant, state)
		if err != nil || actual != valid {
			t.Fatalf("invariant=%t, want %t: %v", actual, valid, err)
		}
	}
	return state
}

func requireJavaGeneratedValue(t *testing.T, state *tlc.TLCStateMut, name string, expected tlc.Value) {
	t.Helper()
	actual := state.Lookup(tlc.UniqueStringOf(name))
	if actual == nil {
		t.Fatalf("%s is unassigned, want %v", name, expected)
	}
	equal, err := expected.Equal(actual)
	if err != nil || !equal {
		t.Fatalf("%s=%v, want %v", name, actual, expected)
	}
}

func testJavaTraceExpressionSpecSafety(t *testing.T, mode string) {
	t.Helper()
	tool := runJavaGeneratedTraceSpec(t, "TESpecTest", "TESpecSafetyTest.cfg", mode, tlc.ExitStatusViolationSafety, false, math.MaxFloat64)
	processor := tool.GetSpecProcessor()
	actions := tool.GetActions()
	if len(actions) != 1 {
		t.Fatalf("actions=%d, want 1", len(actions))
	}
	invariants := processor.GetInvariants()
	if len(invariants) != 1 {
		t.Fatalf("invariants=%d, want 1", len(invariants))
	}
	if processor.GetInitPred().Size() != 1 || processor.GetNextPred() == nil {
		t.Fatal("expected one init predicate and a next-state relation")
	}
	states := javaGeneratedInitialStates(t, tool)
	for i := 0; i < 4; i++ {
		state := requireJavaGeneratedTraceState(t, tool, states, invariants[0], i != 3)
		requireJavaGeneratedValue(t, state, "x", tlc.NewIntValue(int32(i)))
		y := tlc.BoolFalse
		if i%2 != 0 {
			y = tlc.BoolTrue
		}
		requireJavaGeneratedValue(t, state, "y", y)
		if i < 3 {
			var err error
			states, err = tool.GetNextStates(actions[0], state)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if tool.GetModelConfig().GetAlias() == "" {
		t.Fatal("generated config has no ALIAS")
	}
	if tool.GetModelConfig().GetCheckDeadlock() {
		t.Fatal("generated safety config checks deadlocks")
	}
}

func TestJavaTraceExpressionSpecSafetyBFS(t *testing.T) {
	testJavaTraceExpressionSpecSafety(t, "-modelcheck")
}

func TestJavaTraceExpressionSpecSafetySim(t *testing.T) {
	testJavaTraceExpressionSpecSafety(t, "-simulate")
}

// TraceExpressionSpecRuntimeTest.doTest replays the original evaluation-error
// trace, whose final record has val2 instead of val.
func TestJavaTraceExpressionSpecRuntime(t *testing.T) {
	tool := runJavaGeneratedTraceSpec(t, "TESpecRuntimeErrorTest", "TESpecRuntimeErrorTest.cfg", "-modelcheck", tlc.ExitStatusFailureSpecEval, false, math.MaxFloat64)
	processor := tool.GetSpecProcessor()
	actions := tool.GetActions()
	if len(actions) != 1 {
		t.Fatalf("actions=%d, want 1", len(actions))
	}
	invariants := processor.GetInvariants()
	if len(invariants) != 1 {
		t.Fatalf("invariants=%d, want 1", len(invariants))
	}
	if processor.GetInitPred().Size() != 1 || processor.GetNextPred() == nil {
		t.Fatal("expected one init predicate and a next-state relation")
	}
	states := javaGeneratedInitialStates(t, tool)
	for i := 0; i < 5; i++ {
		state := requireJavaGeneratedTraceState(t, tool, states, invariants[0], i != 4)
		field := "val"
		if i == 4 {
			field = "val2"
		}
		expected := tlc.NewRecordValue([]*tlc.UniqueString{tlc.UniqueStringOf(field)}, []tlc.Value{tlc.NewIntValue(int32(i))}, false)
		requireJavaGeneratedValue(t, state, "x", expected)
		if i < 4 {
			var err error
			states, err = tool.GetNextStates(actions[0], state)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if tool.GetModelConfig().GetAlias() == "" {
		t.Fatal("generated config has no ALIAS")
	}
	if tool.GetModelConfig().GetCheckDeadlock() {
		t.Fatal("generated runtime-error config checks deadlocks")
	}
}

func requireJavaGeneratedModules(t *testing.T, tool *tlc.Tool, root string) {
	t.Helper()
	table := tool.GetSpecProcessor().GetModuleTbl()
	if table == nil {
		t.Fatal("generated tool has no module table")
	}
	for _, name := range []string{root, root + "_TEExpression", root + "_TETrace"} {
		if table.GetModuleNode(tlc.UniqueStringOf(name)) == nil {
			t.Fatalf("module table has no %s", name)
		}
	}
}

func TestJavaTraceExpressionSpecDeadlock(t *testing.T) {
	tool := runJavaGeneratedTraceSpec(t, "TESpecDeadlockTest", "TESpecDeadlockTest", "-modelcheck", tlc.ExitStatusViolationDeadlock, true, math.MaxFloat64)
	processor := tool.GetSpecProcessor()
	actions := tool.GetActions()
	if len(actions) != 1 {
		t.Fatalf("actions=%d, want 1", len(actions))
	}
	invariants := processor.GetInvariants()
	if len(invariants) != 1 {
		t.Fatalf("invariants=%d, want 1", len(invariants))
	}
	if processor.GetInitPred().Size() != 1 || processor.GetNextPred() == nil {
		t.Fatal("expected one init predicate and a next-state relation")
	}
	states := javaGeneratedInitialStates(t, tool)
	for i := 0; i < 4; i++ {
		state := requireJavaGeneratedTraceState(t, tool, states, invariants[0], true)
		requireJavaGeneratedValue(t, state, "x", tlc.NewIntValue(int32(i)))
		requireJavaGeneratedValue(t, state, "y", tlc.NewBoolValue(i%2 != 0))
		if i < 3 {
			var err error
			states, err = tool.GetNextStates(actions[0], state)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if tool.GetModelConfig().GetAlias() == "" {
		t.Fatal("generated config has no ALIAS")
	}
	// Java leaves the generated-config deadlock assertion commented out (TODO).
	requireJavaGeneratedModules(t, tool, "TESpecDeadlockTest")
}

func TestJavaTraceExpressionSpecLasso(t *testing.T) {
	// The source override preserves the runtime's default partial-check setting.
	tool := runJavaGeneratedTraceSpec(t, "TESpecTest", "TESpecLassoTest.cfg", "-modelcheck", tlc.ExitStatusViolationLiveness, false, tlc.Globals.LivenessThreshold)
	processor := tool.GetSpecProcessor()
	actions := tool.GetActions()
	if len(actions) != 1 {
		t.Fatalf("actions=%d, want 1", len(actions))
	}
	if len(processor.GetInvariants()) != 0 {
		t.Fatalf("invariants=%d, want 0", len(processor.GetInvariants()))
	}
	if len(processor.GetImpliedTemporals()) != 1 {
		t.Fatalf("properties=%d, want 1", len(processor.GetImpliedTemporals()))
	}
	if processor.GetInitPred().Size() != 1 || processor.GetNextPred() == nil {
		t.Fatal("expected one init predicate and a next-state relation")
	}
	if tool.GetModelConfig().GetAlias() == "" || tool.GetModelConfig().GetCheckDeadlock() {
		t.Fatal("expected ALIAS and disabled deadlock checking")
	}
	requireJavaGeneratedModules(t, tool, "TESpecTest")
	checker, err := tlc.NewLiveCheck1(tool)
	if err != nil {
		t.Fatal(err)
	}
	// Java uses a relative states directory in its test working directory.
	if err := checker.Init(tool, actions, filepath.Join(t.TempDir(), "states")); err != nil {
		t.Fatal(err)
	}
	states := javaGeneratedInitialStates(t, tool)
	var previous *tlc.TLCStateMut
	for i := 0; i < 3; i++ {
		state := requireJavaGeneratedTraceState(t, tool, states, nil, true)
		requireJavaGeneratedValue(t, state, "x", tlc.NewIntValue(int32(i%2)))
		requireJavaGeneratedValue(t, state, "y", tlc.NewBoolValue(i%2 != 0))
		if i == 0 {
			err = checker.AddInitState(tool, state, state.FingerPrint())
		} else {
			next := tlc.NewSetOfStates(1)
			next.Put(state)
			err = checker.AddNextState(tool, previous, previous.FingerPrint(), next)
		}
		if err != nil {
			t.Fatal(err)
		}
		previous = state
		if i < 2 {
			states, err = tool.GetNextStates(actions[0], state)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	_, err = checker.FinalCheck(tool)
	// Java catches LiveException, including its LiveCounterExampleException
	// subclass. Do not accept another error or a successful final check.
	switch err.(type) {
	case *tlc.LiveException, *tlc.LiveCounterExampleException:
		return
	default:
		t.Fatalf("finalCheck error=%T %v, want LiveException", err, err)
	}
}
