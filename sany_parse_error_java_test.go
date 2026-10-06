/*******************************************************************************
 * Copyright (c) 2025 Linux Foundation. All rights reserved.
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
 ******************************************************************************/
package tlago

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ParseErrorTests.testAll retains the source parameter, generated module name,
// parsing-phase failure and assertion on the actual recorded parser output.
func TestParseErrorTests_testAll(t *testing.T) {
	for _, testCase := range []struct{ body, error string }{
		{"x = 0", "Was expecting \"==== or more Module body\"\nEncountered \"x\" at line"},
	} {
		file, err := os.CreateTemp(t.TempDir(), "SanyTest*.tla")
		if err != nil {
			t.Fatal(err)
		}
		path := file.Name()
		moduleName := strings.TrimSuffix(filepath.Base(path), ".tla")
		input := fmt.Sprintf("---- MODULE %s ----\n%s\n====", moduleName, testCase.body)
		if _, err := file.WriteString(input); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		var out strings.Builder
		loader := newSanyLoader(LoadOptions{
			LibraryPaths:    []string{filepath.Dir(path)},
			ParsingProgress: func(message string) { out.WriteString(message); out.WriteByte('\n') },
		})
		_, _, parseFailed := runSanyFrontEndParse(path, loader, nil)
		if !parseFailed {
			t.Fatal("Expected parse failure.")
		}
		if !strings.Contains(out.String(), testCase.error) {
			t.Fatalf("recorded parser output missing %q\n%s", testCase.error, out.String())
		}
	}
}
