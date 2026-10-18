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
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original Github1244Test.testSpec and source runner overrides.
func TestJavaGithub1244(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1244", "Github1244", true, true, true, 1, "-config", "Github1244.tla", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github1244bTest.testSpec and source runner overrides.
func TestJavaGithub1244B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1244", "Github1244", true, true, true, 1, "-config", "Github1244b.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github1244cTest.testSpec and source runner overrides.
func TestJavaGithub1244C(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1244", "Github1244", true, true, true, 1, "-config", "Github1244c.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}
