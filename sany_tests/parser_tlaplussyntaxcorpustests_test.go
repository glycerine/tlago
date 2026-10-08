package sany_tests

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/TlaPlusSyntaxCorpusTests.java.
// testAll remains a supplementary status check until the original AST translator
// and equality assertions are ported. The DSL-kind usage assertion is translated.
func TestTlaPlusSyntaxCorpusTests_testAll(t *testing.T) {
	for _, file := range sanySyntaxCorpusFiles(t) {
		for _, tc := range readSanySyntaxCorpusCases(t, file) {
			tc := tc
			t.Run(filepath.Base(file)+"/"+tc.Title, func(t *testing.T) {
				if tc.Skip {
					t.Skip("source corpus SKIP attribute")
				}
				_, diags := tlago.ParseSanySyntax("Test.tla", tc.Source)
				if isExpectedSanySyntaxParseFailure(tc) {
					if !diags.HasErrors() {
						t.Fatalf("syntax corpus case %q parsed successfully, want parse failure", tc.Title)
					}
					return
				}
				requireNoSANYDiagnostics(t, "parse", diags)
			})
		}
	}
}

func TestTlaPlusSyntaxCorpusTests_testAllTlaPlusNodesUsed(t *testing.T) {
	// Source getTests loads all expected ASTs before parameterized methods run.
	type parameter struct {
		file string
		test sanySyntaxCorpusCase
	}
	var parameters []parameter
	for _, file := range sanySyntaxCorpusFiles(t) {
		for _, tc := range readSanySyntaxCorpusCases(t, file) {
			parameters = append(parameters, parameter{file, tc})
		}
	}
	for _, parameter := range parameters {
		t.Run(filepath.Base(parameter.file)+"/"+parameter.test.Title, func(t *testing.T) {
			var unused []string
			for _, kind := range syntaxCorpusKinds {
				if strings.HasPrefix(kind.name, "pcal") || kind.enum == "FAIR" {
					continue
				}
				if syntaxCorpusUnused[kind.name] {
					unused = append(unused, kind.enum)
				}
			}
			t.Logf("Total unused node kinds: %d", len(unused))
			t.Log(unused)
			if len(unused) != 0 {
				t.Fatalf("expected zero unused TLA+ AST DSL kinds; actual %d", len(unused))
			}
		})
	}
}

type sanySyntaxCorpusCase struct {
	Title       string
	Source      string
	ExpectError bool
	Skip        bool
	ExpectedAST *syntaxCorpusAST
}

func sanySyntaxCorpusFiles(t *testing.T) []string {
	t.Helper()
	root := sanyTestVectorPath("tla2sany", "corpus")
	var files []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".txt" {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		t.Fatalf("walk syntax corpus: %v", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		t.Fatalf("syntax corpus not found under %s", root)
	}
	return files
}

func readSanySyntaxCorpusCases(t *testing.T, file string) []sanySyntaxCorpusCase {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read syntax corpus %s: %v", file, err)
	}
	cases, err := parseSyntaxCorpusFile(file, string(data))
	if err != nil {
		t.Fatalf("parse syntax corpus %s: %v", file, err)
	}
	return cases
}

func isExpectedSyntaxCorpusFailure(title string) bool {
	switch title {
	case "Cartesian Product as Parameter",
		"Invalid Use of LOCAL in LET/IN",
		"Invalid Use of LOCAL in Proof",
		"Step Expression Requiring Lookahead",
		"String with comment start",
		"Nonfix Minus (GH tlaplus/tlaplus #GH884)",
		"Nonfix Submodule Excl (GH tlaplus/tlaplus #GH884)",
		"Nonfix Double Exclamation Operator (GH TSTLA #GH97, GH tlaplus/tlaplus #884)",
		"Label with Subexpression Prefix (GH tlaplus/tlaplus #885)",
		"Empty Tuple Quantification (GH tlaplus/tlaplus #GH888)",
		"Negative Prefix Op on RHS of Infix (GH tlaplus/tlaplus #893)",
		"Mistaken Set Filter Tuples Test":
		return true
	default:
		return false
	}
}

func isExpectedSanySyntaxParseFailure(tc sanySyntaxCorpusCase) bool {
	if isSanySyntaxCorpusJavaParseSuccess(tc.Title) {
		return false
	}
	return tc.ExpectError || isExpectedSyntaxCorpusFailure(tc.Title)
}

func isSanySyntaxCorpusJavaParseSuccess(title string) bool {
	switch title {
	case "Empty Tuple Quantification (GH tlaplus/tlaplus #888)",
		// Source expectFailures inverts this :error fixture: Java accepts
		// LOCAL in LET despite issue 616, so the original expects success.
		"Invalid Use of LOCAL in LET/IN",
		"Invalid Use of LOCAL in Proof",
		"Label with Subexpression Prefix (GH tlaplus/tlaplus #885)",
		"Mistaken Set Filter Tuples Test":
		return true
	default:
		return false
	}
}
