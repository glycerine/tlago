/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
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
// Port of the complete JsonTest.test after the JSON writer feature.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestJavaJson(t *testing.T) {
	const childVariable = "TLAGO_JAVA_JSON_TEST_CHILD"
	if os.Getenv(childVariable) != "true" {
		fixture, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "JsonTests"))
		if err != nil {
			t.Fatal(err)
		}
		directory := t.TempDir()
		if err := os.CopyFS(filepath.Join(directory, "tlc", "test_vectors", "models", "JsonTests"), os.DirFS(fixture)); err != nil {
			t.Fatal(err)
		}
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		// Source test classloader isolation and a private cwd for target/json output.
		command := exec.Command(executable, "-test.run=^TestJavaJson$", "-test.v")
		command.Dir = directory
		command.Env = append(os.Environ(), childVariable+"=true")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("original JsonTest: %v\n%s", err, output)
		}
		return
	}
	result := runJavaTLCModelTest(t, "JsonTests", "-config", "JsonTests.tla")
	// ModelCheckerTestCase.tearDown checks its default successful exit too.
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "0")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL unexpectedly recorded: %v", records)
	}
}
