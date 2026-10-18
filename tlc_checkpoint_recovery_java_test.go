/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
 * Copyright (c) 2023, Oracle and/or its affiliates.
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
	"archive/zip"
	"github.com/glycerine/tlago/tlc"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Complete original CodePlexBug08EWD840FL2FromCheckpointTest.testSpec and
// setup: restore the source Java-produced archive, with three workers.
func TestJavaCodePlexBug08EWD840FL2FromCheckpoint(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	directory := t.TempDir()
	z, err := zip.OpenReader("tlc/test_vectors/models/CodePlexBug08/checkpoint.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	for _, entry := range z.File {
		target := filepath.Join(directory, entry.Name)
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		in, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		out, err := os.Create(target)
		if err != nil {
			in.Close()
			t.Fatal(err)
		}
		_, copyErr := io.Copy(out, in)
		inErr, outErr := in.Close(), out.Close()
		for _, err := range []error{copyErr, inErr, outErr} {
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "EWD840MC2", true, true, true, 3,
		"-gzip", "-recover", filepath.Join(directory, "checkpoint"),
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL2FromCheckpointTest.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCCheckpointRecoverStart, tlc.ECTLCFinished, tlc.ECTLCTemporalPropertyViolated, tlc.ECTLCCounterExample, tlc.ECTLCStatePrint2} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("event %d absent", code)
		}
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCTESpecGenerationComplete} {
		if len(javaTLCRecords(r, code)) != 0 {
			t.Fatalf("unexpected event %d", code)
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCCheckpointRecoverEnd, "1510", "39")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2334", "1566", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "FalseLiveness2")
	requireJavaNodeAndPtrSizes(t, 54038132, 831296)
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ tpos = 0\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 3\n/\\ active = (0 :> TRUE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 3\n/\\ active = (0 :> TRUE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"black\" @@ 3 :> \"white\")",
		"/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
		"/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")",
	}, false)
	// Original assertBackToState(1) checks the loop ordinal.
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
}
