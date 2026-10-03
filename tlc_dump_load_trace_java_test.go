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
// Port of DumpLoadTraceTest methods after implementing source native/body
// overrides and external trace deserialization. Two timing-sensitive enabled
// EWD840 binary methods remain pending; only original @Ignore methods skip.
package tlago

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func runJavaDumpLoadTrace(t *testing.T, name, format string, status int, extraArgs ...string) {
	t.Helper()
	runJavaDumpLoadTraceWithWorkers(t, name, format, status, "1", "1", extraArgs...)
}

func runJavaDumpLoadTraceWithWorkers(t *testing.T, name, format string, status int, dumpWorkers, loadWorkers string, extraArgs ...string) {
	t.Helper()
	runJavaDumpLoadTraceWithSpecs(t, name, name, name, format, status, dumpWorkers, loadWorkers, extraArgs, extraArgs)
}

func runJavaDumpLoadTraceWithSpecs(t *testing.T, fixture, dumpSpec, loadSpec, format string, status int, dumpWorkers, loadWorkers string, dumpExtraArgs, loadExtraArgs []string) {
	t.Helper()
	traceFile := filepath.Join(t.TempDir(), dumpSpec+"."+format)
	dump := runJavaTLCModelTestWithArguments(t, fixture, dumpSpec, func(meta, _ string) []string {
		args := []string{"-metadir", meta, "-workers", dumpWorkers, "-noGenerateSpecTE", "-fp", "4"}
		args = append(args, dumpExtraArgs...)
		return append(args, "-dumpTrace", format, traceFile)
	})
	file, err := os.Stat(traceFile)
	if err != nil {
		t.Fatalf("trace file missing after dump: %v", err)
	}
	if file.Size() == 0 {
		t.Fatal("trace file empty after dump")
	}
	if len(javaTLCRecords(dump, tlc.ECTLCFinished)) == 0 {
		t.Fatal("dump did not record TLC_FINISHED")
	}

	load := runJavaTLCModelTestWithArguments(t, fixture, loadSpec, func(meta, _ string) []string {
		args := []string{"-metadir", meta, "-workers", loadWorkers, "-noGenerateSpecTE", "-fp", "4"}
		args = append(args, loadExtraArgs...)
		return append(args, "-loadTrace", format, traceFile)
	})
	if len(javaTLCRecords(load, tlc.ECTLCFinished)) == 0 {
		t.Fatal("load did not record TLC_FINISHED")
	}
	if dump.ExitStatus != status {
		t.Fatalf("dump exit status=%d, want %d", dump.ExitStatus, status)
	}
	if load.ExitStatus != dump.ExitStatus {
		t.Fatalf("load exit status=%d, want dump status %d", load.ExitStatus, dump.ExitStatus)
	}
	violationCode := tlc.ECTLCInvariantViolatedBehavior
	if status == tlc.ExitStatusViolationLiveness {
		violationCode = tlc.ECTLCTemporalPropertyViolated
	}
	dumpViolation := len(javaTLCRecords(dump, violationCode)) != 0
	loadViolation := len(javaTLCRecords(load, violationCode)) != 0
	if dumpViolation != loadViolation {
		t.Fatal("dump/load violation records differ")
	}

	dumpTrace := javaTLCRecords(dump, tlc.ECTLCStatePrint2)
	loadTrace := javaTLCRecords(load, tlc.ECTLCStatePrint2)
	if len(dumpTrace) == 0 || len(loadTrace) == 0 {
		t.Fatal("dump/load error trace missing")
	}
	if dumpWorkers != loadWorkers {
		if len(loadTrace) > len(dumpTrace) {
			t.Fatalf("prefix length %d exceeds original trace length %d", len(loadTrace), len(dumpTrace))
		}
	} else if len(dumpTrace) != len(loadTrace) {
		t.Fatalf("trace lengths differ: dump=%d load=%d", len(dumpTrace), len(loadTrace))
	}
	for i, loaded := range loadTrace {
		state := dumpTrace[i]
		if state.StateNumber != loaded.StateNumber {
			t.Fatalf("state numbers differ at position %d", i)
		}
		if state.StateInfo == nil || loaded.StateInfo == nil {
			t.Fatalf("state info missing at position %d", i)
		}
		original := strings.TrimSpace(state.StateInfo.String())
		replayed := strings.TrimSpace(loaded.StateInfo.String())
		if dumpSpec != loadSpec {
			assertJavaTraceStatesEqualOnIntersection(t, original, replayed, i)
		} else if original != replayed {
			t.Fatalf("state strings differ at position %d: dump=%q load=%q", i, original, replayed)
		}
	}
}

func TestJavaSafetyDumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTrace(t, "DieHard", "json", tlc.ExitStatusViolationSafety)
}
func TestJavaSafetyDumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTrace(t, "DieHard", "tlc", tlc.ExitStatusViolationSafety)
}

func TestJavaLivenessBidirectionalDumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTrace(t, "BidirectionalTransitions", "json", tlc.ExitStatusViolationLiveness, "-config", "BidirectionalTransitions1Bx.cfg")
}

func TestJavaLivenessBidirectionalDumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTrace(t, "BidirectionalTransitions", "tlc", tlc.ExitStatusViolationLiveness, "-config", "BidirectionalTransitions1Bx.cfg")
}

func TestJavaSafetyDumpLoadTraceJSONAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithWorkers(t, "DieHard", "json", tlc.ExitStatusViolationSafety, "auto", "1")
}

func TestJavaSafetyDumpLoadTraceTLCAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithWorkers(t, "DieHard", "tlc", tlc.ExitStatusViolationSafety, "auto", "1")
}

func TestJavaLivenessBidirectionalDumpLoadTraceJSONAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithWorkers(t, "BidirectionalTransitions", "json", tlc.ExitStatusViolationLiveness, "auto", "1", "-config", "BidirectionalTransitions1Bx.cfg")
}

func TestJavaLivenessBidirectionalDumpLoadTraceTLCAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithWorkers(t, "BidirectionalTransitions", "tlc", tlc.ExitStatusViolationLiveness, "auto", "1", "-config", "BidirectionalTransitions1Bx.cfg")
}

// Port of DumpLoadTraceTest.parseStateString and assertStatesEqualOnIntersection.
func parseJavaTraceStateString(state string) *tlc.InsMap[string, string] {
	vars := tlc.NewInsMap[string, string]()
	for _, line := range strings.Split(state, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "/\\") {
			line = strings.TrimSpace(line[2:])
		}
		if eq := strings.Index(line, "="); eq > 0 {
			vars.Set(strings.TrimSpace(line[:eq]), strings.TrimSpace(line[eq+1:]))
		}
	}
	return vars
}

func assertJavaTraceStatesEqualOnIntersection(t *testing.T, original, replayed string, index int) {
	t.Helper()
	vars1 := parseJavaTraceStateString(original)
	vars2 := parseJavaTraceStateString(replayed)
	common := 0
	for name, value := range vars1.All() {
		other, found := vars2.Get2(name)
		if !found {
			continue
		}
		common++
		if value != other {
			t.Fatalf("state %d variable %q differs: dump=%q load=%q", index, name, value, other)
		}
	}
	if common == 0 {
		t.Fatalf("state %d has no common variables: dump=%v load=%v", index, vars1, vars2)
	}
}

func TestJavaSafetyDieHardAliasSubDumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "json", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "DieHardAliasSub.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSubDumpLoadTraceJSONAutoWorkers(t *testing.T) {
	t.Skip("Upstream @Ignore: Multi-worker BFS + ALIAS variable subsetting: non-shortest counterexample breaks level-indexed trace constraint")
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "json", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "DieHardAliasSub.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSubDumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "tlc", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "DieHardAliasSub.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSubDumpLoadTraceTLCAutoWorkers(t *testing.T) {
	t.Skip("Upstream @Ignore: Multi-worker BFS + ALIAS variable subsetting: non-shortest counterexample breaks level-indexed trace constraint")
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "tlc", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "DieHardAliasSub.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSub2DumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "json", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "DieHardAliasSub2.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "json", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "DieHardAliasSub2.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSub2DumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "tlc", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "DieHardAliasSub2.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSub2DumpLoadTraceTLCAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "tlc", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "DieHardAliasSub2.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSupDumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "json", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "DieHardAliasSup.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSupDumpLoadTraceJSONAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "json", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "DieHardAliasSup.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSupDumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "tlc", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "DieHardAliasSup.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyDieHardAliasSupDumpLoadTraceTLCAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "DieHardAlias", "DieHardAlias", "DieHard", "tlc", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "DieHardAliasSup.cfg"}, []string{"-config", "DieHard.cfg"})
}

func TestJavaSafetyTESpecEqAliasDumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "TESpecTest", "TESpecTest", "TESpecTest", "json", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "TESpecEqAliasSafetyTest.cfg"}, []string{"-config", "TESpecEqAliasSafetyTest.cfg"})
}

func TestJavaSafetyTESpecEqAliasDumpLoadTraceJSONAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "TESpecTest", "TESpecTest", "TESpecTest", "json", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "TESpecEqAliasSafetyTest.cfg"}, []string{"-config", "TESpecEqAliasSafetyTest.cfg"})
}

func TestJavaSafetyTESpecEqAliasDumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "TESpecTest", "TESpecTest", "TESpecTest", "tlc", tlc.ExitStatusViolationSafety, "1", "1", []string{"-config", "TESpecEqAliasSafetyTest.cfg"}, []string{"-config", "TESpecEqAliasSafetyTest.cfg"})
}

func TestJavaSafetyTESpecEqAliasDumpLoadTraceTLCAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "TESpecTest", "TESpecTest", "TESpecTest", "tlc", tlc.ExitStatusViolationSafety, "auto", "1", []string{"-config", "TESpecEqAliasSafetyTest.cfg"}, []string{"-config", "TESpecEqAliasSafetyTest.cfg"})
}

func TestJavaLivenessExample1DumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "Example1", "Example1", "Example1", "json", tlc.ExitStatusViolationLiveness, "1", "1", []string{"-config", "Example1.cfg"}, []string{"-config", "Example1.cfg"})
}

func TestJavaLivenessExample1DumpLoadTraceJSONAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "Example1", "Example1", "Example1", "json", tlc.ExitStatusViolationLiveness, "auto", "1", []string{"-config", "Example1.cfg"}, []string{"-config", "Example1.cfg"})
}

func TestJavaLivenessExample1DumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "Example1", "Example1", "Example1", "tlc", tlc.ExitStatusViolationLiveness, "1", "1", []string{"-config", "Example1.cfg"}, []string{"-config", "Example1.cfg"})
}

func TestJavaLivenessExample1DumpLoadTraceTLCAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "Example1", "Example1", "Example1", "tlc", tlc.ExitStatusViolationLiveness, "auto", "1", []string{"-config", "Example1.cfg"}, []string{"-config", "Example1.cfg"})
}

func TestJavaLivenessMCDumpLoadTraceJSON(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "CodePlexBug08", "MC", "MC", "json", tlc.ExitStatusViolationLiveness, "1", "1", []string{"-deadlock"}, []string{"-deadlock"})
}

func TestJavaLivenessMCDumpLoadTraceJSONAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "CodePlexBug08", "MC", "MC", "json", tlc.ExitStatusViolationLiveness, "auto", "1", []string{"-deadlock"}, []string{"-deadlock"})
}

func TestJavaLivenessMCDumpLoadTraceTLC(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "CodePlexBug08", "MC", "MC", "tlc", tlc.ExitStatusViolationLiveness, "1", "1", []string{"-deadlock"}, []string{"-deadlock"})
}

func TestJavaLivenessMCDumpLoadTraceTLCAutoWorkers(t *testing.T) {
	runJavaDumpLoadTraceWithSpecs(t, "CodePlexBug08", "MC", "MC", "tlc", tlc.ExitStatusViolationLiveness, "auto", "1", []string{"-deadlock"}, []string{"-deadlock"})
}

func TestJavaLivenessEWD840MC3DumpLoadTraceJSON(t *testing.T) {
	t.Skip("Upstream @Ignore: Disabled because JSON trace serialization is garbled because of limited types.")
	runJavaDumpLoadTraceWithSpecs(t, "CodePlexBug08", "EWD840MC3", "EWD840MC3", "json", tlc.ExitStatusViolationLiveness, "1", "1", []string{"-deadlock"}, []string{"-deadlock"})
}
