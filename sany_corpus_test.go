package tlago

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanyJavaSemanticCorpusBehaviors(t *testing.T) {
	corpusDir := filepath.Join("test_vectors", "java-sany", "test", "tla2sany", "semantic", "corpus")
	files, err := filepath.Glob(filepath.Join(corpusDir, "*.tla"))
	if err != nil {
		t.Fatalf("glob semantic corpus: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("semantic corpus not found under %s", corpusDir)
	}
	for _, file := range files {
		if filepath.Base(file) == "Semantics.tla" {
			continue
		}
		t.Run(filepath.Base(file), func(t *testing.T) {
			_, diags := LoadSanySpec(file, LoadOptions{})
			requireNoErrors(t, diags)
		})
	}
}

func TestSanyJavaSyntaxCorpusAcceptedBehaviors(t *testing.T) {
	corpusDir := filepath.Join("test_vectors", "java-sany", "test", "tla2sany", "corpus")
	files, err := filepath.Glob(filepath.Join(corpusDir, "*.txt"))
	if err != nil {
		t.Fatalf("glob syntax corpus: %v", err)
	}
	unicodeFiles, err := filepath.Glob(filepath.Join(corpusDir, "unicode", "*.txt"))
	if err != nil {
		t.Fatalf("glob unicode syntax corpus: %v", err)
	}
	files = append(files, unicodeFiles...)
	if len(files) == 0 {
		t.Fatalf("syntax corpus not found under %s", corpusDir)
	}
	for _, file := range files {
		cases := readSyntaxCorpusCases(t, file)
		for _, tc := range cases {
			if isExpectedSyntaxCorpusFailure(tc.Title) {
				continue
			}
			t.Run(filepath.Base(file)+"/"+tc.Title, func(t *testing.T) {
				_, diags := ParseSanySyntax("Test.tla", tc.Source)
				requireNoErrors(t, diags)
			})
		}
	}
}

type syntaxCorpusCase struct {
	Title  string
	Source string
}

func readSyntaxCorpusCases(t *testing.T, file string) []syntaxCorpusCase {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read syntax corpus %s: %v", file, err)
	}
	lines := strings.Split(string(data), "\n")
	var cases []syntaxCorpusCase
	for i := 0; i < len(lines); i++ {
		if !isSyntaxCorpusSeparator(lines[i], '=') {
			continue
		}
		if i+2 >= len(lines) {
			break
		}
		title := strings.TrimSpace(lines[i+1])
		i += 2
		if !isSyntaxCorpusSeparator(lines[i], '=') {
			continue
		}
		i++
		if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
		start := i
		for i < len(lines) && !isSyntaxCorpusSeparator(lines[i], '-') {
			i++
		}
		source := strings.TrimRight(strings.Join(lines[start:i], "\n"), "\n")
		if title != "" && source != "" {
			cases = append(cases, syntaxCorpusCase{Title: title, Source: source})
		}
	}
	return cases
}

func isSyntaxCorpusSeparator(line string, marker byte) bool {
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
	if strings.Contains(strings.ToLower(title), " error") {
		return true
	}
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
		"Empty Tuple Quantification (GH tlaplus/tlaplus #888)",
		"Negative Prefix Op on RHS of Infix (GH tlaplus/tlaplus #893)",
		"Mistaken Set Filter Tuples Test":
		return true
	default:
		return false
	}
}
