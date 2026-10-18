package tlago

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original SemanticCorpusTests.test and helpers. Keep this translation in the
// root package so assertions inspect actual canonical nodes and references.
func TestSemanticCorpusTests_test(t *testing.T) {
	directory := filepath.Join("sany_tests", "test_vectors", "tla2sany", "semantic", "corpus")
	var files []string
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Ext(path) == ".tla" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			if filepath.Base(file) == "NegativeOpTest.tla" {
				t.Skip("https://github.com/tlaplus/tlaplus/issues/1130")
			}
			initial := sanyGlobalInitialContext(true)
			spec, d := LoadSanySpec(file, LoadOptions{LibraryPaths: []string{filepath.Dir(file)}})
			if d.HasErrors() {
				t.Fatal(d)
			}
			spec.initialContext = initial
			if d = CheckSpec(spec); d.HasErrors() {
				t.Fatal(d)
			}
			for _, diagnostic := range spec.SemanticDiags {
				if diagnostic.Severity == SeverityWarning {
					t.Fatal(diagnostic)
				}
			}
			semantics := spec.semanticModules.getModule("Semantics")
			refersTo := semantics.context.getSymbol("RefersTo").(*sanySemOpDefNode)
			isLevel := semantics.context.getSymbol("IsLevel").(*sanySemOpDefNode)
			levelSymbols := []sanySemSymbol{semantics.context.getSymbol("ConstantLevel"), semantics.context.getSymbol("VariableLevel"), semantics.context.getSymbol("ActionLevel"), semantics.context.getSymbol("TemporalLevel")}
			find := func(def *sanySemOpDefNode) []*sanySemOpApplNode {
				var assertions []*sanySemOpApplNode
				visitor := &sanyExplorerVisitor{postVisit: func(node any) {
					if app, ok := node.(*sanySemOpApplNode); ok && app.operator == def {
						assertions = append(assertions, app)
					}
				}}
				sanyWalkGraph(spec.semanticModules.root, make(map[int32]any), visitor)
				return assertions
			}
			for _, assertion := range find(refersTo) {
				symbol := assertion.operands[0].(*sanySemOpApplNode).operator
				if definition, ok := symbol.(*sanySemOpDefNode); ok {
					symbol = definition.getSource()
				}
				expected := assertion.operands[1].(*tlc.StringNode).GetRep().String()
				comments := symbol.semBase().GetPreComments()
				actual := ""
				if len(comments) > 0 {
					m := canonicalCorpusCommentPattern.FindStringSubmatch(comments[0])
					if m != nil {
						actual = m[1]
					}
				}
				if actual != expected {
					t.Fatalf("%s: expected ID %q, actual %q", assertion.Location, expected, actual)
				}
			}
			for _, assertion := range find(isLevel) {
				operator := assertion.operands[1].(*sanySemOpApplNode).operator
				expected := tlaLevel(-1)
				for i, symbol := range levelSymbols {
					if operator == symbol {
						expected = tlaLevel(i)
						break
					}
				}
				if expected < 0 {
					t.Fatalf("Not a level: %s", operator.semName())
				}
				actual := sanyRequireCanonicalLevelNode(assertion.operands[0]).getLevel()
				if actual != expected {
					t.Fatalf("%s: expected level %d, actual %d", assertion.Location, expected, actual)
				}
			}
		})
	}
}

var canonicalCorpusCommentPattern = regexp.MustCompile(`^\(\*\s*ID:\s*(\S+)\s*\*\)$`)
