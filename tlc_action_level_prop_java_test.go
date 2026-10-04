/*******************************************************************************
 * Copyright (c) 2025 NVIDIA Corp. All rights reserved.
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

// Original ActionLevelPropATest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropA(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropA.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropATest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevel)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevel absent")
	}
}

// Original ActionLevelPropBTest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropB(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropB.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropBTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevel)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevel absent")
	}
}

// Original ActionLevelPropCTest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropC(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropC.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropCTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevel)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevel absent")
	}
}

// Original ActionLevelPropDTest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropD(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropD.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropDTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevelSquareASubV)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevelSquareASubV absent")
	}
}

// Original ActionLevelPropETest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropE(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropE.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropETest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevelAngleASubV)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevelAngleASubV absent")
	}
}
