/*******************************************************************************
 * Copyright (c) 2016 Microsoft Research. All rights reserved.
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
	"sort"
	"strings"
	"testing"
)

// Original UserModuleOverrideTest.testSpec.
func TestJavaUserModuleOverride(t *testing.T) {
	linkJavaNativeOverrideFixtures(t)
	r := runJavaTLCModelTest(t, "UserModuleOverride", "-dumpTrace", "json", filepath.Join(t.TempDir(), "UserModuleOverrideTest.json"))
	requireJavaUserModuleOverrideMismatches(t, r, "UserModuleOverride")
	requireJavaUserModuleOverrideResult(t, r)
}

// Original UserModuleOverrideFromJarTest.testSpec, retaining customBuild's jar.
func TestJavaUserModuleOverrideFromJar(t *testing.T) {
	linkJavaNativeOverrideFixtures(t)
	r := runJavaTLCModelTest(t, "UserModuleOverrideBase", "-dumpTrace", "json", filepath.Join(t.TempDir(), "UserModuleOverrideFromJarTest.json"))
	requireJavaUserModuleOverrideMismatches(t, r, "UserModuleOverrideFromJar")
	requireJavaUserModuleOverrideResult(t, r)
}

// Original UserModuleOverrideAnnotationTest.testSpec. Its initial recorder calls
// inspect availability but do not assert their results.
func TestJavaUserModuleOverrideAnnotation(t *testing.T) {
	linkJavaNativeOverrideFixtures(t)
	r := runJavaTLCModelTest(t, "UserModuleOverrideAnnotation", "-dumpTrace", "json", filepath.Join(t.TempDir(), "UserModuleOverrideAnnotationTest.json"))
	_ = len(javaTLCRecords(r, tlc.ECTLCModuleValueJavaMethodOverrideModuleMismatch)) != 0
	_ = len(javaTLCRecords(r, tlc.ECTLCModuleValueJavaMethodOverrideIdentifierMismatch)) != 0
	_ = len(javaTLCRecords(r, tlc.ECTLCModuleValueJavaMethodOverrideMismatch)) != 0
	requireJavaUserModuleOverrideResult(t, r)
}

func requireJavaUserModuleOverrideMismatches(t *testing.T, r *tlc.Result, class string) {
	t.Helper()
	records := javaTLCRecords(r, tlc.ECTLCModuleValueJavaMethodOverrideMismatch)
	if len(records) != 2 {
		t.Fatalf("mismatch count=%d, want 2", len(records))
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Params[0] < records[j].Params[0] })
	for i, name := range []string{"Get2", "Get3"} {
		params := records[i].Params
		if params[0] != name {
			t.Fatalf("mismatch name=%s, want %s", params[0], name)
		}
		if !strings.HasSuffix(params[1], class+".class") {
			t.Fatalf("resource=%s, want suffix %s.class", params[1], class)
		}
		args := ""
		if i == 0 {
			args = "tlc2.value.impl.Value"
		}
		want := "<Java Method: public static tlc2.value.impl.Value " + class + "." + name + "(" + args + ")>"
		if params[2] != want {
			t.Fatalf("method=%s, want %s", params[2], want)
		}
	}
}

func requireJavaUserModuleOverrideResult(t *testing.T, r *tlc.Result) {
	t.Helper()
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCUncovered(t, r)
}
