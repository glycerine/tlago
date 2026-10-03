/*******************************************************************************
 * Copyright (c) 2017 Microsoft Research. All rights reserved.
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
// Ports of EmptySubsetEqTest.testSpec and SubsetEqTest.testSpec after the subset/interval value feature.
// Run the original model through the production Go SANY bridge and TLC runner.
package tlago

import (
	"context"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func runJavaTLCModelTest(t *testing.T, name string, extraArgs ...string) *tlc.Result {
	t.Helper()
	return runJavaTLCModelTestWithCoverage(t, name, true, extraArgs...)
}

func runJavaTLCModelTestWithCoverage(t *testing.T, name string, coverage bool, extraArgs ...string) *tlc.Result {
	t.Helper()
	return runJavaTLCModelTestWithSettings(t, name, coverage, true, extraArgs...)
}

func runJavaTLCModelTestWithSettings(t *testing.T, name string, coverage, dump bool, extraArgs ...string) *tlc.Result {
	t.Helper()
	return runJavaTLCModelTestWithWorkers(t, name, coverage, dump, 1, extraArgs...)
}

func runJavaTLCModelTestWithWorkers(t *testing.T, name string, coverage, dump bool, workers int, extraArgs ...string) *tlc.Result {
	t.Helper()
	return runJavaTLCModelTestWithDebugger(t, name, coverage, dump, true, workers, extraArgs...)
}

func runJavaTLCModelTestWithDebugger(t *testing.T, name string, coverage, dump, debugger bool, workers int, extraArgs ...string) *tlc.Result {
	t.Helper()
	oldCoverage, oldCheckpoint, oldMeta := tlc.Globals.CoverageInterval, tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir
	oldWorkers, oldMain, oldSimulator := tlc.Globals.NumWorkers, tlc.Globals.MainChecker, tlc.Globals.Simulator
	oldTool, oldDFID, oldStart := tlc.Globals.Tool, tlc.Globals.DFIDMax, tlc.Globals.StartTime
	oldContinuation := tlc.Globals.Continuation
	oldPoly := tlc.FP64IrredPoly()
	oldUserDir := tlc.GetFilenameUserDirectory()
	// The upstream runner isolates TLC statics with a per-test classloader.
	tlc.SetMainChecker(nil)
	tlc.SetSimulator(nil)
	tlc.Globals.Continuation = false
	t.Cleanup(func() {
		tlc.Globals.CoverageInterval, tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir = oldCoverage, oldCheckpoint, oldMeta
		tlc.Globals.NumWorkers, tlc.Globals.MainChecker, tlc.Globals.Simulator = oldWorkers, oldMain, oldSimulator
		tlc.Globals.Tool, tlc.Globals.DFIDMax, tlc.Globals.StartTime = oldTool, oldDFID, oldStart
		tlc.Globals.Continuation = oldContinuation
		tlc.FP64InitPoly(oldPoly)
		tlc.SetFilenameUserDirectory(oldUserDir)
		tlc.SetTLCStateTool(nil)
	})
	directory, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", name))
	if err != nil {
		t.Fatal(err)
	}
	tlc.SetFilenameUserDirectory(&directory)
	meta := t.TempDir()
	traceDirectory := t.TempDir()
	args := []string{"-metadir", meta, "-deadlock", "-generateSpecTE", "-teSpecOutDir", traceDirectory, "-fp", "0", "-seed", "1", "-workers", strconv.Itoa(workers), "-checkpoint", "0"}
	if workers == 1 && debugger {
		args = append(args, "-debugger", "nosuspend,port=4712,nohalt")
	}
	tlc.Globals.CoverageInterval = -1
	if coverage {
		args = append(args, "-coverage", "1")
	}
	if dump {
		args = append(args, "-dump", "dot", filepath.Join(meta, name+".dot"))
	}
	args = append(args, extraArgs...)
	args = append(args, name)
	opts, err := tlc.ParseTLCOptions(args)
	if err != nil {
		t.Fatal(err)
	}
	opts.FPSetConfiguration = tlc.NewFPSetConfigurationWithRatioAndImplementation(1, "tlc2.tool.fp.MSBDiskFPSet")
	opts.FPSetConfiguration.SetMemory(1 << 20)
	// ModelCheckerTestCase installs its recorder before tool construction, so
	// retain configuration diagnostics as well as messages from TLC.process.
	recorder := &tlc.MemoryRecorder{}
	tlc.AddMessageRecorder(recorder)
	defer tlc.RemoveMessageRecorder(recorder)
	opts.LoadTool = func() (*tlc.Tool, error) {
		tool, diags, err := loadTLCAppTool(opts.SpecFile, opts.ConfigFile, nil, opts.RuntimeParams)
		requireNoErrors(t, diags)
		return tool, err
	}
	result, _ := tlc.NewTLC(opts).Process(context.Background())
	if result == nil {
		t.Fatal("TLC result is nil")
	}
	result.Messages = append([]tlc.Message(nil), recorder.Messages...)
	return result
}

func javaTLCRecords(result *tlc.Result, code int) []tlc.Message {
	var matches []tlc.Message
	for _, m := range result.Messages {
		if m.Code == code {
			matches = append(matches, m)
		}
	}
	return matches
}

func TestJavaEmptySubsetEq(t *testing.T) {
	result := runJavaTLCModelTest(t, "EmptySubsetEq")
	if result.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit status=%d, want %d", result.ExitStatus, tlc.ExitStatusFailureSpecEval)
	}
	records := func(code int) []tlc.Message { return javaTLCRecords(result, code) }
	if len(records(tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	matched := false
	for _, m := range records(tlc.ECTLCStats) {
		if reflect.DeepEqual(m.Params, []string{"1", "1", "0"}) {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_STATS=%v, want 1/1/0", records(tlc.ECTLCStats))
	}
	if len(records(tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("A TE spec was generated, but it should not be")
	}
	want := "TLC threw an unexpected exception.\nThis was probably caused by an error in the spec or model.\nSee the User Output or TLC Console for clues to what happened.\nThe exception was a java.lang.RuntimeException\n: Attempted to check if the value:\n{}\nis in the integer interval 1..4"
	matched = false
	for _, m := range records(tlc.ECGeneral) {
		if reflect.DeepEqual(m.Params, []string{want}) {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("GENERAL=%v, want %q", records(tlc.ECGeneral), want)
	}
	states := records(tlc.ECTLCStatePrint2)
	if len(states) != 1 {
		t.Fatalf("trace has %d states, want 1", len(states))
	}
	if states[0].StateInfo == nil || strings.TrimSpace(states[0].StateInfo.String()) != "b = TRUE" || states[0].StateNumber != 1 {
		t.Fatalf("trace state=%+v, want state 1: b = TRUE", states[0])
	}
	info, ok := states[0].StateInfo.Info.(string)
	if !ok || info == "<Initial predicate>" || strings.HasPrefix(info, "<Action") {
		t.Fatalf("extended trace action=%v", states[0].StateInfo.Info)
	}
	uncovered := map[string]bool{}
	for _, m := range records(tlc.ECTLCCoverageValue) {
		if len(m.Params) >= 2 && m.Params[1] == "0" {
			uncovered[strings.TrimSpace(strings.ReplaceAll(m.Params[0], "|", ""))] = true
		}
	}
	wantUncovered := map[string]bool{"line 8, col 9 to line 8, col 48 of module EmptySubsetEq": true}
	if !reflect.DeepEqual(uncovered, wantUncovered) {
		t.Fatalf("uncovered=%v, want %v; coverage=%+v cost=%+v", uncovered, wantUncovered, records(tlc.ECTLCCoverageValue), records(tlc.ECTLCCoverageValueCost))
	}
}

// Port of SubsetEqTest.testSpec: the power-set reduction must complete the
// original large-interval model without enumerating either power set.
func TestJavaSubsetEq(t *testing.T) {
	result := runJavaTLCModelTest(t, "SubsetEq")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d code=%d, want success; messages=%+v", result.ExitStatus, result.ErrorCode, result.Messages)
	}
	records := func(code int) []tlc.Message { return javaTLCRecords(result, code) }
	if len(records(tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if got := records(tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
	if len(records(tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("A TE spec was generated, but it should not be")
	}
	matched := false
	for _, m := range records(tlc.ECTLCStats) {
		if reflect.DeepEqual(m.Params, []string{"2", "1", "0"}) {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_STATS=%v, want 2/1/0", records(tlc.ECTLCStats))
	}
	if got := records(tlc.ECTLCStatePrint2); len(got) != 0 {
		t.Fatalf("unexpected error trace: %v", got)
	}
	for _, m := range records(tlc.ECTLCCoverageValue) {
		if len(m.Params) >= 2 && strings.TrimSpace(m.Params[1]) == "0" {
			t.Fatalf("unexpected uncovered line: %v", m.Params)
		}
	}
}
