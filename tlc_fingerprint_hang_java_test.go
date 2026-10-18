// Copyright (c) 2023, Oracle and/or its affiliates.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original FingerprintExceptionHangTest.testSpec.
func TestJavaFingerprintExceptionHang(t *testing.T) {
	r := runJavaTLCModelTest(t, "FingerprintExceptionHang", "-dumpTrace", "json", filepath.Join(t.TempDir(), "FingerprintExceptionHangTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCFingerprintException,
		"0. Line 16, column 5 to line 16, column 46 in FingerprintExceptionHang\n1. Line 16, column 20 to line 16, column 46 in FingerprintExceptionHang\n2. Line 16, column 30 to line 16, column 45 in FingerprintExceptionHang\n\n",
		"Attempted to compare integer 1 with non-integer:\n{}")
}
