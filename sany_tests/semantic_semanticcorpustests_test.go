package sany_tests

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/SemanticCorpusTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestSemanticCorpusTests_test(t *testing.T) {
	t.Skip("tla2sany wip")

	corpusDir := sanyTestVectorPath("tla2sany", "semantic", "corpus")
	files := sanyTLAFilesUnder(t, corpusDir, func(path string) bool {
		name := filepath.Base(path)
		return name != "Semantics.tla" && name != "NegativeOpTest.tla"
	})
	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			spec, parseDiags := tlago.LoadSanySpec(file, tlago.LoadOptions{LibraryPaths: []string{corpusDir}, PreferLibraryModules: true})
			requireNoSANYDiagnostics(t, "parse", parseDiags)
			semDiags := tlago.CheckSpec(spec)
			if warnings := semDiags.Warnings(); len(warnings) != 0 {
				t.Fatalf("semantic warnings:\n%s", warnings.Error())
			}
			requireNoSANYDiagnostics(t, "semantic", semDiags)
			requireSemanticCorpusAssertions(t, spec)
		})
	}
}

func requireSemanticCorpusAssertions(t *testing.T, spec *tlago.Spec) {
	t.Helper()
	refersTo := findSemanticCorpusCalls(spec.Root, "RefersTo")
	isLevel := findSemanticCorpusCalls(spec.Root, "IsLevel")
	if len(refersTo) == 0 && len(isLevel) == 0 {
		t.Fatalf("%s contains no RefersTo or IsLevel assertions", spec.Root.SourcePath)
	}
	ids := semanticCorpusCommentIDs(spec)
	for _, call := range refersTo {
		if len(call.Args) != 2 {
			t.Fatalf("RefersTo at %+v has %d args, want 2", call.Pos, len(call.Args))
		}
		name, ok := semanticCorpusReferencedName(call.Args[0])
		if !ok {
			t.Fatalf("RefersTo first argument at %+v is not an operator reference: %#v", call.Pos, call.Args[0])
		}
		expected, ok := semanticCorpusStringLiteral(call.Args[1])
		if !ok {
			t.Fatalf("RefersTo second argument at %+v is not a string literal: %#v", call.Pos, call.Args[1])
		}
		actual := ids[name]
		if actual == "" && strings.Contains(name, "!") {
			actual = ids[name[strings.LastIndex(name, "!")+1:]]
		}
		if actual != expected {
			t.Fatalf("RefersTo(%s, %q) resolved to comment ID %q", name, expected, actual)
		}
	}
	leveler := newSemanticCorpusLeveler(spec)
	for _, call := range isLevel {
		if len(call.Args) != 2 {
			t.Fatalf("IsLevel at %+v has %d args, want 2", call.Pos, len(call.Args))
		}
		expected, ok := semanticCorpusExpectedLevel(call.Args[1])
		if !ok {
			t.Fatalf("IsLevel second argument at %+v is not a level constant: %#v", call.Pos, call.Args[1])
		}
		if actual := leveler.level(call.Args[0]); actual != expected {
			t.Fatalf("IsLevel at %+v = %s, want %s", call.Pos, actual, expected)
		}
	}
}

func findSemanticCorpusCalls(mod *tlago.Module, name string) []*tlago.CallExpr {
	if mod == nil {
		return nil
	}
	var calls []*tlago.CallExpr
	for _, def := range mod.Definitions {
		walkSemanticCorpusExpr(def.Expr, func(expr tlago.Expr) {
			call, ok := expr.(*tlago.CallExpr)
			if !ok {
				return
			}
			if callee, ok := semanticCorpusReferencedName(call.Callee); ok && callee == name {
				calls = append(calls, call)
			}
		})
	}
	for _, assumption := range mod.Assumptions {
		walkSemanticCorpusExpr(assumption.Expr, func(expr tlago.Expr) {
			call, ok := expr.(*tlago.CallExpr)
			if !ok {
				return
			}
			if callee, ok := semanticCorpusReferencedName(call.Callee); ok && callee == name {
				calls = append(calls, call)
			}
		})
	}
	for _, theorem := range mod.Theorems {
		walkSemanticCorpusExpr(theorem.Expr, func(expr tlago.Expr) {
			call, ok := expr.(*tlago.CallExpr)
			if !ok {
				return
			}
			if callee, ok := semanticCorpusReferencedName(call.Callee); ok && callee == name {
				calls = append(calls, call)
			}
		})
	}
	for _, nested := range mod.Nested {
		calls = append(calls, findSemanticCorpusCalls(nested, name)...)
	}
	return calls
}

func semanticCorpusCommentIDs(spec *tlago.Spec) map[string]string {
	ids := map[string]string{}
	for _, mod := range spec.Modules {
		for _, def := range mod.Definitions {
			id := semanticCorpusCommentID(def.PreComments)
			if id == "" {
				continue
			}
			ids[def.Name] = id
			if mod.Name != "" {
				ids[mod.Name+"!"+def.Name] = id
			}
		}
	}
	return ids
}

var semanticCorpusCommentIDRE = regexp.MustCompile(`^\(\*\s*ID:\s*(\S+)\s*\*\)$`)

func semanticCorpusCommentID(comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	match := semanticCorpusCommentIDRE.FindStringSubmatch(strings.TrimSpace(comments[0]))
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func semanticCorpusStringLiteral(expr tlago.Expr) (string, bool) {
	lit, ok := expr.(*tlago.LiteralExpr)
	if !ok || lit.Kind != "string" {
		return "", false
	}
	return strings.Trim(lit.Value, `"`), true
}

func semanticCorpusReferencedName(expr tlago.Expr) (string, bool) {
	switch e := expr.(type) {
	case *tlago.IdentExpr:
		return e.Name, true
	case *tlago.CallExpr:
		return semanticCorpusReferencedName(e.Callee)
	default:
		return "", false
	}
}

type semanticCorpusLevel string

const (
	semanticCorpusConstantLevel semanticCorpusLevel = "ConstantLevel"
	semanticCorpusVariableLevel semanticCorpusLevel = "VariableLevel"
	semanticCorpusActionLevel   semanticCorpusLevel = "ActionLevel"
	semanticCorpusTemporalLevel semanticCorpusLevel = "TemporalLevel"
)

func semanticCorpusExpectedLevel(expr tlago.Expr) (semanticCorpusLevel, bool) {
	name, ok := semanticCorpusReferencedName(expr)
	if !ok {
		return "", false
	}
	switch name {
	case "ConstantLevel":
		return semanticCorpusConstantLevel, true
	case "VariableLevel":
		return semanticCorpusVariableLevel, true
	case "ActionLevel":
		return semanticCorpusActionLevel, true
	case "TemporalLevel":
		return semanticCorpusTemporalLevel, true
	default:
		return "", false
	}
}

type semanticCorpusLeveler struct {
	decls      map[string]tlago.DeclarationKind
	defs       map[string]tlago.Definition
	cache      map[string]semanticCorpusLevel
	inProgress map[string]bool
}

func newSemanticCorpusLeveler(spec *tlago.Spec) *semanticCorpusLeveler {
	leveler := &semanticCorpusLeveler{
		decls:      map[string]tlago.DeclarationKind{},
		defs:       map[string]tlago.Definition{},
		cache:      map[string]semanticCorpusLevel{},
		inProgress: map[string]bool{},
	}
	for _, mod := range spec.Modules {
		for _, decl := range mod.Declarations {
			for _, name := range decl.Names {
				leveler.decls[name] = decl.Kind
				if mod.Name != "" {
					leveler.decls[mod.Name+"!"+name] = decl.Kind
				}
			}
		}
		for _, def := range mod.Definitions {
			leveler.defs[def.Name] = def
			if mod.Name != "" {
				leveler.defs[mod.Name+"!"+def.Name] = def
			}
		}
	}
	return leveler
}

func (l *semanticCorpusLeveler) level(expr tlago.Expr) semanticCorpusLevel {
	switch e := expr.(type) {
	case nil:
		return semanticCorpusConstantLevel
	case *tlago.IdentExpr:
		if def, ok := l.defs[e.Name]; ok {
			return l.definitionLevel(e.Name, def)
		}
		switch l.decls[e.Name] {
		case tlago.VariableDecl:
			return semanticCorpusVariableLevel
		default:
			return semanticCorpusConstantLevel
		}
	case *tlago.LiteralExpr:
		return semanticCorpusConstantLevel
	case *tlago.UnaryExpr:
		switch e.Op {
		case "'", "UNCHANGED":
			return maxSemanticCorpusLevel(semanticCorpusActionLevel, l.level(e.Expr))
		case "[]", "<>":
			return semanticCorpusTemporalLevel
		default:
			return l.level(e.Expr)
		}
	case *tlago.BinaryExpr:
		level := maxSemanticCorpusLevel(l.level(e.Left), l.level(e.Right))
		if e.Op == "~>" || e.Op == "-+->" {
			return maxSemanticCorpusLevel(semanticCorpusTemporalLevel, level)
		}
		return level
	case *tlago.CallExpr:
		level := l.level(e.Callee)
		for _, arg := range e.Args {
			level = maxSemanticCorpusLevel(level, l.level(arg))
		}
		if name, ok := semanticCorpusReferencedName(e.Callee); ok {
			switch name {
			case "ENABLED":
				level = maxSemanticCorpusLevel(level, semanticCorpusVariableLevel)
			case "UNCHANGED":
				level = maxSemanticCorpusLevel(level, semanticCorpusActionLevel)
			case "[]", "<>", "WF_", "SF_", "~>", "-+->":
				level = maxSemanticCorpusLevel(level, semanticCorpusTemporalLevel)
			}
		}
		return level
	case *tlago.IfExpr:
		return maxSemanticCorpusLevel(l.level(e.Cond), l.level(e.Then), l.level(e.Else))
	case *tlago.LetExpr:
		level := l.level(e.Body)
		for _, def := range e.Definitions {
			level = maxSemanticCorpusLevel(level, l.level(def.Expr))
		}
		return level
	case *tlago.QuantifierExpr:
		return maxSemanticCorpusLevel(l.level(e.Set), l.level(e.Body))
	case *tlago.CaseExpr:
		level := semanticCorpusConstantLevel
		for _, arm := range e.Arms {
			level = maxSemanticCorpusLevel(level, l.level(arm.Test), l.level(arm.Value))
		}
		return maxSemanticCorpusLevel(level, l.level(e.Other))
	case *tlago.ChooseExpr:
		return maxSemanticCorpusLevel(l.level(e.Set), l.level(e.Body))
	case *tlago.TupleExpr:
		return l.exprsLevel(e.Elems)
	case *tlago.SetExpr:
		return l.exprsLevel(e.Elems)
	case *tlago.RecordExpr:
		level := semanticCorpusConstantLevel
		for _, field := range e.Fields {
			level = maxSemanticCorpusLevel(level, l.level(field.Value))
		}
		return level
	case *tlago.RecordComponentExpr:
		return l.level(e.Record)
	case *tlago.RecordSetExpr:
		level := semanticCorpusConstantLevel
		for _, field := range e.Fields {
			level = maxSemanticCorpusLevel(level, l.level(field.Set))
		}
		return level
	case *tlago.FunctionExpr:
		level := l.level(e.Body)
		for _, bound := range e.Bounds {
			level = maxSemanticCorpusLevel(level, l.level(bound.Set))
		}
		return level
	case *tlago.FunctionAppExpr:
		return maxSemanticCorpusLevel(l.level(e.Function), l.exprsLevel(e.Args))
	case *tlago.ExceptExpr:
		level := l.level(e.Base)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				level = maxSemanticCorpusLevel(level, l.exprsLevel(component.Indices))
			}
			level = maxSemanticCorpusLevel(level, l.level(spec.Value))
		}
		return level
	case *tlago.LabelExpr:
		return l.level(e.Body)
	case *tlago.ActionExpr:
		return maxSemanticCorpusLevel(semanticCorpusActionLevel, l.level(e.Action), l.level(e.Subscript))
	case *tlago.FairnessExpr:
		return semanticCorpusTemporalLevel
	case *tlago.FunctionSetExpr:
		return maxSemanticCorpusLevel(l.level(e.Domain), l.level(e.Range))
	case *tlago.SetComprehensionExpr:
		level := l.level(e.Element)
		for _, bound := range e.Bounds {
			level = maxSemanticCorpusLevel(level, l.level(bound.Set))
		}
		return maxSemanticCorpusLevel(level, l.level(e.Predicate))
	default:
		return semanticCorpusConstantLevel
	}
}

func (l *semanticCorpusLeveler) definitionLevel(name string, def tlago.Definition) semanticCorpusLevel {
	if cached, ok := l.cache[name]; ok {
		return cached
	}
	if l.inProgress[name] {
		return semanticCorpusConstantLevel
	}
	l.inProgress[name] = true
	level := l.level(def.Expr)
	l.inProgress[name] = false
	l.cache[name] = level
	return level
}

func (l *semanticCorpusLeveler) exprsLevel(exprs []tlago.Expr) semanticCorpusLevel {
	level := semanticCorpusConstantLevel
	for _, expr := range exprs {
		level = maxSemanticCorpusLevel(level, l.level(expr))
	}
	return level
}

func maxSemanticCorpusLevel(levels ...semanticCorpusLevel) semanticCorpusLevel {
	max := semanticCorpusConstantLevel
	for _, level := range levels {
		if semanticCorpusLevelRank(level) > semanticCorpusLevelRank(max) {
			max = level
		}
	}
	return max
}

func semanticCorpusLevelRank(level semanticCorpusLevel) int {
	switch level {
	case semanticCorpusTemporalLevel:
		return 3
	case semanticCorpusActionLevel:
		return 2
	case semanticCorpusVariableLevel:
		return 1
	default:
		return 0
	}
}

func walkSemanticCorpusExpr(expr tlago.Expr, visit func(tlago.Expr)) {
	if expr == nil {
		return
	}
	visit(expr)
	switch e := expr.(type) {
	case *tlago.UnaryExpr:
		walkSemanticCorpusExpr(e.Expr, visit)
	case *tlago.BinaryExpr:
		walkSemanticCorpusExpr(e.Left, visit)
		walkSemanticCorpusExpr(e.Right, visit)
	case *tlago.CallExpr:
		walkSemanticCorpusExpr(e.Callee, visit)
		for _, arg := range e.Args {
			walkSemanticCorpusExpr(arg, visit)
		}
	case *tlago.IfExpr:
		walkSemanticCorpusExpr(e.Cond, visit)
		walkSemanticCorpusExpr(e.Then, visit)
		walkSemanticCorpusExpr(e.Else, visit)
	case *tlago.LetExpr:
		for _, def := range e.Definitions {
			walkSemanticCorpusExpr(def.Expr, visit)
		}
		walkSemanticCorpusExpr(e.Body, visit)
	case *tlago.QuantifierExpr:
		walkSemanticCorpusExpr(e.Set, visit)
		walkSemanticCorpusExpr(e.Body, visit)
	case *tlago.CaseExpr:
		for _, arm := range e.Arms {
			walkSemanticCorpusExpr(arm.Test, visit)
			walkSemanticCorpusExpr(arm.Value, visit)
		}
		walkSemanticCorpusExpr(e.Other, visit)
	case *tlago.ChooseExpr:
		walkSemanticCorpusExpr(e.Set, visit)
		walkSemanticCorpusExpr(e.Body, visit)
	case *tlago.TupleExpr:
		for _, elem := range e.Elems {
			walkSemanticCorpusExpr(elem, visit)
		}
	case *tlago.SetExpr:
		for _, elem := range e.Elems {
			walkSemanticCorpusExpr(elem, visit)
		}
	case *tlago.RecordExpr:
		for _, field := range e.Fields {
			walkSemanticCorpusExpr(field.Value, visit)
		}
	case *tlago.RecordComponentExpr:
		walkSemanticCorpusExpr(e.Record, visit)
	case *tlago.RecordSetExpr:
		for _, field := range e.Fields {
			walkSemanticCorpusExpr(field.Set, visit)
		}
	case *tlago.FunctionExpr:
		for _, bound := range e.Bounds {
			walkSemanticCorpusExpr(bound.Set, visit)
		}
		walkSemanticCorpusExpr(e.Body, visit)
	case *tlago.FunctionAppExpr:
		walkSemanticCorpusExpr(e.Function, visit)
		for _, arg := range e.Args {
			walkSemanticCorpusExpr(arg, visit)
		}
	case *tlago.ExceptExpr:
		walkSemanticCorpusExpr(e.Base, visit)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					walkSemanticCorpusExpr(index, visit)
				}
			}
			walkSemanticCorpusExpr(spec.Value, visit)
		}
	case *tlago.LabelExpr:
		walkSemanticCorpusExpr(e.Body, visit)
	case *tlago.ActionExpr:
		walkSemanticCorpusExpr(e.Action, visit)
		walkSemanticCorpusExpr(e.Subscript, visit)
	case *tlago.FairnessExpr:
		walkSemanticCorpusExpr(e.Subscript, visit)
		walkSemanticCorpusExpr(e.Action, visit)
	case *tlago.FunctionSetExpr:
		walkSemanticCorpusExpr(e.Domain, visit)
		walkSemanticCorpusExpr(e.Range, visit)
	case *tlago.SetComprehensionExpr:
		walkSemanticCorpusExpr(e.Element, visit)
		for _, bound := range e.Bounds {
			walkSemanticCorpusExpr(bound.Set, visit)
		}
		walkSemanticCorpusExpr(e.Predicate, visit)
	}
}

func sanyTLAFilesUnder(t *testing.T, root string, accept func(string) bool) []string {
	t.Helper()
	var files []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".tla" {
			return nil
		}
		if accept == nil || accept(path) {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		t.Fatalf("no TLA+ files under %s", root)
	}
	return files
}
