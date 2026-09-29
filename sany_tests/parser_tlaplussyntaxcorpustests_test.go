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
// Each test starts skipped until its Java assertions are ported and made green.
func TestTlaPlusSyntaxCorpusTests_testAll(t *testing.T) {
	for _, file := range sanySyntaxCorpusFiles(t) {
		for _, tc := range readSanySyntaxCorpusCases(t, file) {
			tc := tc
			t.Run(filepath.Base(file)+"/"+tc.Title, func(t *testing.T) {
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
	used := map[string]bool{}
	for _, file := range sanySyntaxCorpusFiles(t) {
		for _, tc := range readSanySyntaxCorpusCases(t, file) {
			if isExpectedSanySyntaxParseFailure(tc) {
				continue
			}
			root, diags := tlago.ParseSanySyntax("Test.tla", tc.Source)
			requireNoSANYDiagnostics(t, "parse "+tc.Title, diags)
			collectSanySyntaxNodeKinds(root, used)
		}
	}

	var unused []string
	for _, def := range tlago.SanySyntaxNodeKinds {
		if !strings.HasPrefix(def.Name, "N_") {
			continue
		}
		if isLegacyUnusedJavaSyntaxNodeKind(def.Name) {
			continue
		}
		if !used[def.Name] {
			unused = append(unused, def.Name)
		}
	}
	sort.Strings(unused)
	if len(unused) != 0 {
		t.Fatalf("unused TLA+ syntax node kinds: %v", unused)
	}
}

type sanySyntaxCorpusCase struct {
	Title       string
	Source      string
	ExpectError bool
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
	lines := strings.Split(string(data), "\n")
	var cases []sanySyntaxCorpusCase
	for i := 0; i < len(lines); {
		if !isSanySyntaxCorpusSeparator(lines[i], '=') {
			i++
			continue
		}
		i++
		var header []string
		for i < len(lines) && !isSanySyntaxCorpusSeparator(lines[i], '=') {
			header = append(header, lines[i])
			i++
		}
		if i >= len(lines) {
			break
		}
		i++
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
		start := i
		for i < len(lines) && !isSanySyntaxCorpusSeparator(lines[i], '-') {
			i++
		}
		source := strings.TrimRight(strings.Join(lines[start:i], "\n"), "\n")
		title, expectError := parseSanySyntaxCorpusHeader(header)
		if title != "" && source != "" {
			cases = append(cases, sanySyntaxCorpusCase{Title: title, Source: source, ExpectError: expectError})
		}
		for i < len(lines) && !isSanySyntaxCorpusSeparator(lines[i], '=') {
			i++
		}
	}
	return cases
}

func parseSanySyntaxCorpusHeader(lines []string) (string, bool) {
	var title []string
	expectError := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed == ":error" {
			expectError = true
			continue
		}
		title = append(title, trimmed)
	}
	if len(title) == 0 {
		return "", expectError
	}
	return title[0], expectError
}

func isSanySyntaxCorpusSeparator(line string, marker byte) bool {
	line = strings.TrimSpace(line)
	if !strings.HasSuffix(line, "|||") {
		return false
	}
	prefix := strings.TrimSuffix(line, "|||")
	if prefix == "" {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if prefix[i] != marker {
			return false
		}
	}
	return true
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
		"Invalid Use of LOCAL in Proof",
		"Label with Subexpression Prefix (GH tlaplus/tlaplus #885)",
		"Mistaken Set Filter Tuples Test":
		return true
	default:
		return false
	}
}

func isLegacyUnusedJavaSyntaxNodeKind(kind string) bool {
	switch kind {
	case "N_ActDecl",
		"N_AssumeDecl",
		"N_FunctionParam",
		"N_InnerProof",
		"N_Integer",
		"N_NonExprBody",
		"N_NumberedAssumeProve",
		"N_NumerableStep",
		"N_ParamDecl",
		"N_TempDecl":
		return true
	default:
		return false
	}
}

func collectSanySyntaxNodeKinds(node *tlago.SanySyntaxNode, used map[string]bool) {
	if node == nil {
		return
	}
	used[node.Kind.JavaName()] = true
	for _, child := range node.GetHeirs() {
		collectSanySyntaxNodeKinds(child, used)
	}
}
