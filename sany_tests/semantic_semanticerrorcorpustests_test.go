package sany_tests

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/SemanticErrorCorpusTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestSemanticErrorCorpusTests_test(t *testing.T) {
	t.Skip("tla2sany wip")

	corpusDir := sanyTestVectorPath("tla2sany", "semantic", "error_corpus")
	files := sanyTLAFilesUnder(t, corpusDir, func(path string) bool {
		return semanticErrorCorpusFilenameRE.MatchString(filepath.Base(path))
	})
	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			expectedCode, expectError := expectedSemanticErrorCorpusCode(t, file)
			spec, parseDiags := tlago.LoadSanySpec(file, tlago.LoadOptions{LibraryPaths: []string{corpusDir}, PreferLibraryModules: true})
			diags := append(tlago.Diagnostics{}, parseDiags...)
			if !parseDiags.HasErrors() {
				diags = append(diags, tlago.CheckSpec(spec)...)
			}
			if got := diags.HasErrors(); got != expectError {
				t.Fatalf("failure state = %v, want %v for %s\n%s", got, expectError, expectedCode, diags.Error())
			}
			for _, diag := range diags {
				if diag.Code == expectedCode {
					return
				}
			}
			t.Fatalf("diagnostics missing expected code %s\n%s", expectedCode, diags.Error())
		})
	}
}

var semanticErrorCorpusFilenameRE = regexp.MustCompile(`^[WE]\d+(?:_\S+)?_Test\.tla$`)

func expectedSemanticErrorCorpusCode(t *testing.T, file string) (string, bool) {
	t.Helper()
	name := filepath.Base(file)
	if !semanticErrorCorpusFilenameRE.MatchString(name) {
		t.Fatalf("semantic error corpus filename %q does not encode an error code", name)
	}
	code := strings.SplitN(strings.TrimSuffix(name, "_Test.tla"), "_", 2)[0]
	return code, strings.HasPrefix(code, "E")
}
