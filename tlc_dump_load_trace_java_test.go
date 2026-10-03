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
// Port of DumpLoadTraceTest's safety and bidirectional methods after
// implementing source native/body override lookup through INSTANCE imports.
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
	traceFile := filepath.Join(t.TempDir(), name+"."+format)
	dump := runJavaTLCModelTestWithArguments(t, name, name, func(meta, _ string) []string {
		args := []string{"-metadir", meta, "-workers", dumpWorkers, "-noGenerateSpecTE", "-fp", "4"}
		args = append(args, extraArgs...)
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

	load := runJavaTLCModelTestWithArguments(t, name, name, func(meta, _ string) []string {
		args := []string{"-metadir", meta, "-workers", loadWorkers, "-noGenerateSpecTE", "-fp", "4"}
		args = append(args, extraArgs...)
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
		if original != replayed {
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
