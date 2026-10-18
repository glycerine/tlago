/*******************************************************************************
 * Copyright (c) 2019 Microsoft Research. All rights reserved.
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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// The original beforeSetUp counter supplies these four rows to its four identical
// test methods. Keep the complete matrix and its assertExitStatus override.
func runJavaGithub432(t *testing.T, humans, others string, expectedWarnings []string) {
	t.Helper()
	template, err := os.ReadFile(filepath.Join("tlc", "test_vectors", "models", "Github432", "Github432.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	config := strings.Replace(string(template), "%1%", humans, 1)
	config = strings.Replace(config, "%2%", others, 1)
	configPath := filepath.Join(t.TempDir(), "Github432.cfg")
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	r := runJavaTLCModelTest(t, "Github432", "-config", configPath,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github432Test.json"))
	records := javaTLCRecords(r, tlc.ECTLCSymmetrySetTooSmall)
	if expectedWarnings == nil {
		if len(records) != 0 {
			t.Fatalf("TLC_SYMMETRY_SET_TOO_SMALL present: %v", records)
		}
		return
	}
	if len(records) == 0 {
		t.Fatal("TLC_SYMMETRY_SET_TOO_SMALL absent")
	}
	if records[0].Params == nil {
		t.Fatal("Reported parameters should not be null.")
	}
	if len(records[0].Params) != len(expectedWarnings) {
		t.Fatalf("warning parameters=%v, want %v", records[0].Params, expectedWarnings)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSymmetrySetTooSmall, expectedWarnings...)
}

func TestJavaGithub432A(t *testing.T) {
	runJavaGithub432(t, "Alice", "Cat, Dog", []string{"", "Humans", "has", "s"})
}
func TestJavaGithub432B(t *testing.T) {
	runJavaGithub432(t, "", "Emu", []string{"s", "Humans, and Others", "have", ""})
}
func TestJavaGithub432C(t *testing.T) {
	runJavaGithub432(t, "Frank, Glenda", "", []string{"", "Others", "has", "s"})
}
func TestJavaGithub432D(t *testing.T) {
	runJavaGithub432(t, "Hauser, Ignatio", "Jackal, Kangaroo", nil)
}
