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
