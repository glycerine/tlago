/*******************************************************************************
 * Copyright (c) 2024 Microsoft Research. All rights reserved.
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

// Original Github858Test.testSpec and all simulation/dump settings.
func TestJavaGithub858(t *testing.T) {
	// Java's tool-jar manifest supplies the CommunityModules dependencies
	// imported by the original _TLCActionTrace runtime postcondition.
	modules, err := filepath.Abs(filepath.Join("test_vectors", "java-sany", "CommunityModules.jar"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLASSPATH", modules)
	output := t.TempDir()
	r := runJavaTLCModelTestWithRoot(t, "EWD998", "EWD998", false, true, false, 1,
		"-dumpTrace", "json", filepath.Join(output, "Github858Test.json"),
		"-config", "Github858.cfg", "-simulate", "-dumptrace", "tlcaction", filepath.Join(output, "states", "Github858Test.tla"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(records) < 100 {
		t.Fatalf("trace length=%d, want at least 100", len(records))
	}
	record := records[99]
	if record.StateNumber != 100 {
		t.Fatalf("state ordinal=%d, want 100", record.StateNumber)
	}
	const expectedState = `/\ active = (0 :> TRUE @@ 1 :> FALSE @@ 2 :> FALSE)
/\ pending = (0 :> 0 @@ 1 :> 0 @@ 2 :> 0)
/\ color = (0 :> "white" @@ 1 :> "white" @@ 2 :> "white")
/\ counter = (0 :> -1 @@ 1 :> 0 @@ 2 :> 1)
/\ token = [pos |-> 1, q |-> 1, color |-> "black"]
/\ TokenInits = 4
/\ TokenPasses = 7
/\ Activations = (0 :> 6 @@ 1 :> 7 @@ 2 :> 9)
/\ Activations2 = (0 :> 6 @@ 1 :> 7 @@ 2 :> 9)
/\ Actives = { (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE),
  (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE),
  (0 :> FALSE @@ 1 :> TRUE @@ 2 :> TRUE),
  (0 :> TRUE @@ 1 :> FALSE @@ 2 :> FALSE),
  (0 :> TRUE @@ 1 :> FALSE @@ 2 :> TRUE),
  (0 :> TRUE @@ 1 :> TRUE @@ 2 :> FALSE),
  (0 :> TRUE @@ 1 :> TRUE @@ 2 :> TRUE) }
/\ SumQ = 129
/\ SumQP = 130`
	if actual := replJavaTrim(record.StateInfo.String()); actual != expectedState {
		t.Fatalf("state100=%q, want %q", actual, expectedState)
	}
	const expectedAction = `<PassToken(2) line 57, col 3 to line 64, col 43 of module EWD998>`
	if actual, ok := record.StateInfo.Info.(string); !ok || actual != expectedAction {
		t.Fatalf("action100=%v, want %s", record.StateInfo.Info, expectedAction)
	}
}
