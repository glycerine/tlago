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
func TestTlaPlusSyntaxCorpusTests_testAll(t *testing.T) {
	for _, file := range sanySyntaxCorpusFiles(t) {
		for _, tc := range readSanySyntaxCorpusCases(t, file) {
			t.Run(filepath.Base(file)+"/"+tc.Title, func(t *testing.T) {
				if tc.Skip {
					t.Skip("source corpus SKIP attribute")
				}
				defer func() {
					if r := recover(); r != nil {
						if failure, ok := r.(*syntaxCorpusAssertionFailure); ok {
							t.Fatalf("translator assertion: %s", failure.message)
						}
						panic(r)
					}
				}()
				actual, err := syntaxCorpusParseTarget(tc.Source)
				if isExpectedSyntaxCorpusFailure(tc.Title) {
					t.Log("Expecting failure.")
					// Java catches checked translation ParseException, but not AssertionError.
					if err != nil {
						return
					}
					if tc.ExpectError == (actual == nil) {
						t.Fatalf("%s: expected known failure to invert parser acceptance", tc.Title)
					}
					return
				}
				if err != nil {
					t.Fatalf("translation: %v", err)
				}
				if tc.ExpectError {
					t.Log("Expecting parser rejection.")
					if actual != nil {
						t.Fatalf("%s: expected parser rejection; actual %s", tc.Title, actual)
					}
				} else {
					if actual == nil {
						t.Fatalf("%s: expected parser acceptance", tc.Title)
					}
					t.Logf("Expect: %s", tc.ExpectedAST)
					t.Logf("Actual: %s", actual)
					tc.ExpectedAST.testEquality(t, actual)
				}
			})
		}
	}
}

// Java's target returns null only for parser/lexer rejection; the translator's
// checked ParseException remains distinct from its JUnit assertion failures.
func syntaxCorpusParseTarget(input string) (actual *syntaxCorpusAST, err *syntaxCorpusDSLError) {
	defer func() {
		if r := recover(); r != nil {
			if failure, ok := r.(*syntaxCorpusDSLError); ok {
				actual = nil
				err = failure
			} else {
				panic(r)
			}
		}
	}()
	root, diags := tlago.ParseSanySyntax("Test.tla", input)
	if diags.HasErrors() {
		return nil, nil
	}
	return syntaxCorpusToAST(root), nil
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
	switch strings.ToLower(title) {
	case "cartesian product as parameter",
		"invalid use of local in let/in",
		"invalid use of local in proof",
		"step expression requiring lookahead",
		"string with comment start",
		"nonfix minus (gh tlaplus/tlaplus #gh884)",
		"nonfix submodule excl (gh tlaplus/tlaplus #gh884)",
		"nonfix double exclamation operator (gh tstla #gh97, gh tlaplus/tlaplus #884)",
		"label with subexpression prefix (gh tlaplus/tlaplus #885)",
		"empty tuple quantification (gh tlaplus/tlaplus #888)",
		"negative prefix op on rhs of infix (gh tlaplus/tlaplus #893)",
		"mistaken set filter tuples test":
		return true
	default:
		return false
	}
}
