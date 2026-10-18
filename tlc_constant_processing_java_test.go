/*******************************************************************************
 * Copyright (c) 2019 Microsoft Research. All rights reserved.
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
// Port of Github361Test.testSpec after module constant processing.
package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaGithub361(t *testing.T) {
	// The Java override uses two workers so single-worker optimizations cannot
	// hide expensive initialization of the Partitions constant.
	result := runJavaTLCModelTestWithWorkers(t, "Github361", true, true, 2,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github361.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitGenerated1, "1")
}
