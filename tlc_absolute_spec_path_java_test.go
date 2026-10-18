/*******************************************************************************
 * Copyright (c) 2017 Microsoft Research. All rights reserved.
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
	"os"
	"path/filepath"
	"testing"
)

// Original AbsoluteSpecPathTest.test: the only CLI argument is the absolute
// source Test2 path. The workspace provides BASE_DIR's nonempty absolute root.
func TestJavaAbsoluteSpecPath(t *testing.T) {
	// Run the unchanged fixtures from an isolated absolute BASE_PATH so the
	// source default states directory is removed with this test's workspace.
	basePath := t.TempDir()
	for _, filename := range []string{"Test2.tla", "Test2.cfg"} {
		bytes, err := os.ReadFile(filepath.Join("tlc", "test_vectors", "models", "Test2", filename))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(basePath, filename), bytes, 0600); err != nil {
			t.Fatal(err)
		}
	}
	spec := filepath.Join(basePath, "Test2")
	r := runJavaTLCModelTestWithArguments(t, "Test2", spec, func(_, _ string) []string { return nil })
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "5")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "6", "5", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}
