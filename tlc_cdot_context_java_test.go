package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original CdotWithContextATest.testSpec; coverage and debugger disabled by source.
func TestJavaCdotWithContextA(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "CdotWithContextA", false, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "CdotWithContextATest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "3", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
}

// Original CdotWithContextBTest.testSpec; coverage and debugger disabled by source.
func TestJavaCdotWithContextB(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "CdotWithContextB", false, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "CdotWithContextBTest.json"), "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "3", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
}

// Original CdotWithContextCTest.testSpec; coverage and debugger disabled by source.
func TestJavaCdotWithContextC(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "CdotWithContextC", false, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "CdotWithContextCTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "3", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
}

// Original CdotWithContextDTest.testSpec; coverage and debugger disabled by source.
func TestJavaCdotWithContextD(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "CdotWithContextD", false, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "CdotWithContextDTest.json"), "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
}

// Original ChainedCdotsTest.testSpec; coverage and debugger disabled by source.
func TestJavaChainedCdots(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "ChainedCdots", false, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ChainedCdotsTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR recorded")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 is empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 9 {
		t.Fatalf("register42=%v, want IntValue9", values[0])
	}
}
