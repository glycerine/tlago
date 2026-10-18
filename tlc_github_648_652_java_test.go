/*******************************************************************************
 * Copyright (c) 2022 Microsoft Research. All rights reserved.
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

// Original Github648Test.testSpec, including every original coverage value/cost.
func TestJavaGithub648(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github648", "Github648", true, false, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github648Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1332", "36", "0")
	assertJavaCoverage(t, r, `<Init line 34, col 1 to line 34, col 4 of module Github648>: 36:36
  line 35, col 8 to line 35, col 23 of module Github648: 1
  line 36, col 8 to line 36, col 23 of module Github648: 36
  |line 36, col 14 to line 36, col 23 of module Github648: 6
<Next line 38, col 1 to line 38, col 4 of module Github648>: 0:1296
  line 39, col 8 to line 39, col 24 of module Github648: 216
  |line 39, col 15 to line 39, col 24 of module Github648: 36
  line 40, col 8 to line 40, col 24 of module Github648: 1296
  |line 40, col 15 to line 40, col 24 of module Github648: 216
<Inv line 42, col 1 to line 42, col 3 of module Github648>
  line 43, col 5 to line 49, col 62 of module Github648: 36
  |line 43, col 8 to line 43, col 23 of module Github648: 36
  |line 44, col 8 to line 44, col 27 of module Github648: 36
  |line 45, col 8 to line 45, col 23 of module Github648: 36
  |line 46, col 8 to line 46, col 27 of module Github648: 36
  |line 47, col 8 to line 49, col 62 of module Github648: 36
  ||line 47, col 8 to line 49, col 54 of module Github648: 36
  |||line 47, col 16 to line 49, col 53 of module Github648: 1
  ||||line 48, col 9 to line 49, col 52 of module Github648: 1
  |||||line 48, col 17 to line 49, col 51 of module Github648: 1
  ||||||line 13, col 3 to line 13, col 34 of module Github648: 1
  |||||||line 13, col 9 to line 13, col 34 of module Github648: 1:131290
  ||||||||line 13, col 10 to line 13, col 20 of module Github648: 4:326526
  |||||||||line 13, col 11 to line 13, col 14 of module Github648: 4
  |||||||||line 13, col 19 to line 13, col 19 of module Github648: 4
  ||||||||line 13, col 30 to line 13, col 33 of module Github648: 1
  ||||||line 49, col 17 to line 49, col 47 of module Github648: 1
  |||||||line 49, col 25 to line 49, col 46 of module Github648: 1
  ||||||||line 13, col 3 to line 13, col 34 of module Github648: 1
  |||||||||line 13, col 9 to line 13, col 34 of module Github648: 1:69094
  ||||||||||line 13, col 10 to line 13, col 20 of module Github648: 4:186
  |||||||||||line 13, col 11 to line 13, col 14 of module Github648: 4
  |||||||||||line 13, col 19 to line 13, col 19 of module Github648: 4
  ||||||||||line 13, col 30 to line 13, col 33 of module Github648: 1
  ||||||||line 49, col 36 to line 49, col 42 of module Github648: 1:75`)
}

// Original Github648wNTest.testSpec, including every original coverage value/cost.
func TestJavaGithub648wN(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github648", "Github648", true, false, true, 10,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github648wNTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1332", "36", "0")
	assertJavaCoverage(t, r, `<Init line 34, col 1 to line 34, col 4 of module Github648>: 36:36
  line 35, col 8 to line 35, col 23 of module Github648: 1
  line 36, col 8 to line 36, col 23 of module Github648: 36
  |line 36, col 14 to line 36, col 23 of module Github648: 6
<Next line 38, col 1 to line 38, col 4 of module Github648>: 0:1296
  line 39, col 8 to line 39, col 24 of module Github648: 216
  |line 39, col 15 to line 39, col 24 of module Github648: 36
  line 40, col 8 to line 40, col 24 of module Github648: 1296
  |line 40, col 15 to line 40, col 24 of module Github648: 216
<Inv line 42, col 1 to line 42, col 3 of module Github648>
  line 43, col 5 to line 49, col 62 of module Github648: 36
  |line 43, col 8 to line 43, col 23 of module Github648: 36
  |line 44, col 8 to line 44, col 27 of module Github648: 36
  |line 45, col 8 to line 45, col 23 of module Github648: 36
  |line 46, col 8 to line 46, col 27 of module Github648: 36
  |line 47, col 8 to line 49, col 62 of module Github648: 36
  ||line 47, col 8 to line 49, col 54 of module Github648: 36
  |||line 47, col 16 to line 49, col 53 of module Github648: 1
  ||||line 48, col 9 to line 49, col 52 of module Github648: 1
  |||||line 48, col 17 to line 49, col 51 of module Github648: 10
  ||||||line 13, col 3 to line 13, col 34 of module Github648: 10
  |||||||line 13, col 9 to line 13, col 34 of module Github648: 10:131290
  ||||||||line 13, col 10 to line 13, col 20 of module Github648: 40:2674464
  |||||||||line 13, col 11 to line 13, col 14 of module Github648: 40
  |||||||||line 13, col 19 to line 13, col 19 of module Github648: 40
  ||||||||line 13, col 30 to line 13, col 33 of module Github648: 10
  ||||||line 49, col 17 to line 49, col 47 of module Github648: 10
  |||||||line 49, col 25 to line 49, col 46 of module Github648: 10
  ||||||||line 13, col 3 to line 13, col 34 of module Github648: 10
  |||||||||line 13, col 9 to line 13, col 34 of module Github648: 10:690148
  ||||||||||line 13, col 10 to line 13, col 20 of module Github648: 40:1473
  |||||||||||line 13, col 11 to line 13, col 14 of module Github648: 40
  |||||||||||line 13, col 19 to line 13, col 19 of module Github648: 40
  ||||||||||line 13, col 30 to line 13, col 33 of module Github648: 10
  ||||||||line 49, col 36 to line 49, col 42 of module Github648: 10:750`)
}

// Original Github652Test.testSpec with its embedded config and all defaults.
func TestJavaGithub652(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github652", "-config", "Github652.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github652Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
}
