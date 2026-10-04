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
package tlago

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// TTraceModelCheckerTestCase rechecks the actual generated artifact, with a
// resolver that includes the original spec path. No source-model rerun replaces
// this second phase. The original first tests also retain all their assertions.
func runJavaTTraceRecheck(t *testing.T, name, generated string, randomFPSeed, dump bool) *tlc.Result {
	t.Helper()
	return runJavaTTraceRecheckWithDeadlock(t, name, generated, randomFPSeed, dump, false)
}

func runJavaTTraceRecheckWithDeadlock(t *testing.T, name, generated string, randomFPSeed, dump, checkDeadlock bool) *tlc.Result {
	t.Helper()
	return runJavaTTraceRecheckWithWorkers(t, name, generated, randomFPSeed, dump, checkDeadlock, 1)
}

func runJavaTTraceRecheckWithWorkers(t *testing.T, name, generated string, randomFPSeed, dump, checkDeadlock bool, workers int) *tlc.Result {
	t.Helper()
	if info, err := os.Stat(generated); err != nil || !info.Mode().IsRegular() {
		t.Skip("No TE spec was generated, please run test with original spec")
	}
	original, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", name))
	if err != nil {
		t.Fatal(err)
	}
	classpath, err := tlcApplicationClasspath(nil)
	if err != nil {
		t.Fatal(err)
	}
	generatedDirectory := filepath.Dir(generated)
	generatedName := filepath.Base(generated)
	resolver := tlc.NewSimpleFilenameToStream([]string{original}, tlc.FilenameResolverOptions{Classpath: classpath, UserDirectory: &generatedDirectory})
	setJavaModelLivenessThreshold(t, math.MaxFloat64)
	return runJavaTLCModelTestWithArguments(t, name, generatedName, func(meta, traceDirectory string) []string {
		args := []string{"-metadir", meta, "-noGenerateSpecTE", "-workers", strconv.Itoa(workers), "-checkpoint", "0"}
		if workers == 1 {
			args = append(args, "-debugger", "nosuspend,port=4712,nohalt")
		}
		if !checkDeadlock {
			args = append(args, "-deadlock")
		}
		if !randomFPSeed {
			args = append(args, "-fp", "0", "-seed", "1")
		}
		if dump {
			args = append(args, "-dump", "dot", filepath.Join(meta, filepath.Base(generated)+".dot"))
		}
		return append(args, "-dumpTrace", "json", filepath.Join(t.TempDir(), "TTrace.json"), "-config", generatedName)
	}, resolver)
}

// Original Github461Test_TTraceTest.testSpec and inherited settings.
func TestJavaGithub461TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "Github461TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaGithub461(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "Github461", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2", "x = 3", "x = 4"}, true)
	requireJavaTLCUncovered(t, r)
}

// Original Github597Test_TTraceTest.testSpec, including random fp/seed and no DOT.
func TestJavaGithub597TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "Github597TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaGithub597(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "Github597", generated, true, false)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	for _, code := range []int{tlc.ECTLCTemporalPropertyViolated, tlc.ECTLCCounterExample, tlc.ECTLCStatePrint2, tlc.ECTLCBackToState} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic%d absent", code)
		}
	}
}
