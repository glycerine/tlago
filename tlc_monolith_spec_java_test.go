/*******************************************************************************
 * Copyright (c) 2020 Microsoft Research. All rights reserved.
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
	"bytes"
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Original MonolithSpecTest.testSpec and beforeSetUp's actual ToolIO streams.
func TestJavaMonolithSpec(t *testing.T) {
	var stdout, stderr bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&stdout, &stderr)
	defer restore()
	tlc.ToolIOReset()
	r := runJavaTLCModelTest(t, "MonolithSpec", "-config", "MonolithSpec.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "MonolithSpecTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "214", "54", "0")
	if strings.Contains(stderr.String(), "File does not exist") {
		t.Fatalf("unexpected missing-file warning: %s", stderr.String())
	}
	// Source test-model path denotes the model-fixture directory; retain its
	// exact suffix using this port's test_vectors/models/MonolithSpec directory.
	sep := regexp.QuoteMeta(string(filepath.Separator))
	for _, name := range []string{"EWD840", "Mod4711", "Mod4712"} {
		pattern := "Parsing file .*" + sep + name + `.tla \(.*` + sep + "models" + sep + "MonolithSpec" + sep + `MonolithSpec.tla\)`
		if !regexp.MustCompile(pattern).MatchString(stdout.String()) {
			t.Fatalf("parsing provenance missing for %s: %s", name, stdout.String())
		}
	}
	requireJavaTLCUncovered(t, r)
}
