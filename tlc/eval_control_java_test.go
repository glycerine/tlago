package tlc

import "testing"

// Original tlc2.tool.EvalControlTest.test, preserving all twelve assertions.
func TestJavaEvalControl(t *testing.T) {
	control := EvalClear
	if EvalIsEnabled(control) {
		t.Fatal("Clear is enabled")
	}
	if EvalIsKeepLazy(control) {
		t.Fatal("Clear keeps lazy")
	}
	if EvalIsPrimed(control) {
		t.Fatal("Clear is primed")
	}

	control = EvalSetEnabled(control)
	if !EvalIsEnabled(control) {
		t.Fatal("setEnabled did not enable")
	}
	if EvalIsKeepLazy(control) {
		t.Fatal("setEnabled keeps lazy")
	}
	if EvalIsPrimed(control) {
		t.Fatal("setEnabled is primed")
	}

	control = EvalSetKeepLazy(control)
	if !EvalIsEnabled(control) {
		t.Fatal("setKeepLazy lost enabled")
	}
	if !EvalIsKeepLazy(control) {
		t.Fatal("setKeepLazy did not keep lazy")
	}
	if EvalIsPrimed(control) {
		t.Fatal("setKeepLazy is primed")
	}

	control = EvalSetPrimed(control)
	if !EvalIsEnabled(control) {
		t.Fatal("setPrimed lost enabled")
	}
	if !EvalIsKeepLazy(control) {
		t.Fatal("setPrimed lost keep lazy")
	}
	if !EvalIsPrimed(control) {
		t.Fatal("setPrimed did not prime")
	}
}

// Original tlc2.tool.EvalControlTest.testIfEnabled.
func TestJavaEvalControlIfEnabled(t *testing.T) {
	control := EvalClear
	if EvalIsPrimed(EvalSetPrimedIfEnabled(control)) {
		t.Fatal("setPrimedIfEnabled primed Clear")
	}
	control = EvalSetEnabled(control)
	if !EvalIsEnabled(control) {
		t.Fatal("setEnabled did not enable")
	}
	if !EvalIsPrimed(EvalSetPrimedIfEnabled(control)) {
		t.Fatal("setPrimedIfEnabled did not prime enabled control")
	}
}
