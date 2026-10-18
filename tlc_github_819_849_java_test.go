/*******************************************************************************
 * Copyright (c) 2023 Microsoft Research. All rights reserved.
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
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github819Test.testSpec and constructor overrides.
func TestJavaGithub819(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github819", "-noGenerateSpecTE", "-config", "Github819.tla")
	if r.ExitStatus != tlc.ExitStatusFailureLivenessEval {
		t.Fatalf("exit=%d, want FAILURE_LIVENESS_EVAL; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_TAUTOLOGY absent")
	}
}

// Original Github849Test.testSpec and default trace/coverage/debugger settings.
func TestJavaGithub849(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github849", "-config", "Github849.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github849Test.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCModuleValueJavaMethodOverride)) != 0 {
		t.Fatal("TLC_MODULE_VALUE_JAVA_METHOD_OVERRIDE present")
	}
}
