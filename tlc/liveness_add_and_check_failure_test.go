package tlc

import "testing"

// No upstream test directly exercises AddAndCheckLiveCheck's asymmetric
// check0 catch clauses. Keep this focused on the source exception policy.
func TestAddAndCheckLiveCheckCheckFailurePolicyMatchesJava(t *testing.T) {
	ioFailure := NewIOException("io")
	if err := addAndCheckInitCheckResult(ioFailure); err != nil {
		t.Fatalf("addInitState check failure = %v, want swallowed", err)
	}
	interrupted := NewInterruptedException("stop")
	if err := addAndCheckNextCheckResult(interrupted); err != nil {
		t.Fatalf("addNextState interruption = %v, want swallowed", err)
	}
	if err := addAndCheckNextCheckResult(ioFailure); err != ioFailure {
		t.Fatalf("addNextState I/O failure = %v, want original %v", err, ioFailure)
	}
	runtimeFailure := NewRuntimeException("runtime")
	if err := addAndCheckInitCheckResult(runtimeFailure); err != runtimeFailure {
		t.Fatalf("addInitState runtime failure = %v, want original %v", err, runtimeFailure)
	}
	if err := addAndCheckNextCheckResult(runtimeFailure); err != runtimeFailure {
		t.Fatalf("addNextState runtime failure = %v, want original %v", err, runtimeFailure)
	}
}

func TestAddAndCheckLiveCheckConstructorUsesNoDebugToolLikeJava(t *testing.T) {
	plain := NewTool()
	debug := NewTool()
	calls := 0
	debug.NoDebugFunc = func(received *Tool) *Tool {
		calls++
		if received != debug {
			t.Fatalf("NoDebug receiver = %p, want %p", received, debug)
		}
		return plain
	}
	check := NewAddAndCheckLiveCheck(debug, make([]*OrderOfSolution, 0), t.TempDir())
	if calls != 1 || check.Tool != plain || !check.AddAndCheck {
		t.Fatalf("constructor result calls/tool/addAndCheck = %d/%p/%v, want 1/%p/true", calls, check.Tool, check.AddAndCheck, plain)
	}
}

func TestAddAndCheckLiveCheckConstructorRejectsNullToolLikeJava(t *testing.T) {
	defer func() {
		failure := recover()
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("null tool failure = %T: %v, want NullPointerException", failure, failure)
		}
	}()
	_ = NewAddAndCheckLiveCheck(nil, nil, t.TempDir())
}
