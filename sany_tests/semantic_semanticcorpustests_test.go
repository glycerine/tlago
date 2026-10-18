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

// Supplementary native AST facade checks. The faithful original class lives in
// root sany_semantic_corpus_java_test.go and inspects canonical graph identities.
func TestSemanticCorpusASTFacadeBehaviors(t *testing.T) {
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
			requireNoSANYDiagnostics(t, "semantic", semDiags)
			requireSemanticCorpusAssertions(t, spec)
		})
	}
}

func requireSemanticCorpusAssertions(t *testing.T, spec *tlago.Spec) {
	t.Helper()
	refersTo := findSemanticCorpusCalls(spec.Root, "RefersTo")
	leveler := newSemanticCorpusLeveler(spec)
	isLevel := findSemanticCorpusLevelAssertions(spec.Root, leveler)
	if len(refersTo) == 0 && len(isLevel) == 0 {
		return
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
		if actual == "" && semanticCorpusIDMatchesName(expected, name) && ids[expected] == expected {
			actual = expected
		}
		if actual != expected {
			t.Fatalf("RefersTo(%s, %q) resolved to comment ID %q", name, expected, actual)
		}
	}
	for _, assertion := range isLevel {
		call := assertion.call
		if len(call.Args) != 2 {
			t.Fatalf("IsLevel at %+v has %d args, want 2", call.Pos, len(call.Args))
		}
		expected, ok := semanticCorpusExpectedLevel(call.Args[1])
		if !ok {
			t.Fatalf("IsLevel second argument at %+v is not a level constant: %#v", call.Pos, call.Args[1])
		}
		if actual := assertion.leveler.level(call.Args[0]); actual != expected {
			t.Fatalf("IsLevel at %+v = %s, want %s", call.Pos, actual, expected)
		}
	}
}

type semanticCorpusLevelAssertion struct {
	call    *tlago.CallExpr
	leveler *semanticCorpusLeveler
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

func findSemanticCorpusLevelAssertions(mod *tlago.Module, leveler *semanticCorpusLeveler) []semanticCorpusLevelAssertion {
	if mod == nil {
		return nil
	}
	var assertions []semanticCorpusLevelAssertion
	visit := func(expr tlago.Expr, scoped *semanticCorpusLeveler) {
		call, ok := expr.(*tlago.CallExpr)
		if !ok {
			return
		}
		if callee, ok := semanticCorpusReferencedName(call.Callee); ok && callee == "IsLevel" {
			assertions = append(assertions, semanticCorpusLevelAssertion{call: call, leveler: scoped})
		}
	}
	for _, def := range mod.Definitions {
		walkSemanticCorpusExprWithLeveler(def.Expr, leveler, visit)
	}
	for _, assumption := range mod.Assumptions {
		walkSemanticCorpusExprWithLeveler(assumption.Expr, leveler, visit)
	}
	for _, theorem := range mod.Theorems {
		walkSemanticCorpusExprWithLeveler(theorem.Expr, leveler, visit)
	}
	for _, nested := range mod.Nested {
		assertions = append(assertions, findSemanticCorpusLevelAssertions(nested, leveler)...)
	}
	return assertions
}

func semanticCorpusCommentIDs(spec *tlago.Spec) map[string]string {
	ids := map[string]string{}
	add := func(modName, name, id string) {
		if name == "" || id == "" {
			return
		}
		ids[name] = id
		ids[id] = id
		if modName != "" {
			ids[modName+"!"+name] = id
		}
	}
	for _, mod := range spec.Modules {
		for _, decl := range mod.Declarations {
			for _, name := range decl.Names {
				add(mod.Name, name, semanticCorpusCommentID(decl.NamePreComments[name]))
			}
		}
		for _, def := range mod.Definitions {
			add(mod.Name, def.Name, semanticCorpusCommentID(def.PreComments))
		}
		for _, assumption := range mod.Assumptions {
			add(mod.Name, assumption.Name, semanticCorpusCommentID(assumption.PreComments))
		}
		for _, theorem := range mod.Theorems {
			add(mod.Name, theorem.Name, semanticCorpusCommentID(theorem.PreComments))
		}
		for _, id := range semanticCorpusSourceCommentIDs(mod.Source) {
			ids[id] = id
		}
	}
	return ids
}

func semanticCorpusIDMatchesName(id, name string) bool {
	if id == name {
		return true
	}
	if strings.HasSuffix(id, "!"+name) {
		return true
	}
	return false
}

var semanticCorpusCommentIDRE = regexp.MustCompile(`^\(\*\s*ID:\s*(\S+)\s*\*\)$`)
var semanticCorpusAnyCommentIDRE = regexp.MustCompile(`\(\*\s*ID:\s*(\S+)\s*\*\)`)

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

func semanticCorpusSourceCommentIDs(source string) []string {
	matches := semanticCorpusAnyCommentIDRE.FindAllStringSubmatch(source, -1)
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) == 2 {
			ids = append(ids, match[1])
		}
	}
	return ids
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
	case *tlago.UnaryExpr:
		return e.Op, true
	case *tlago.BinaryExpr:
		return e.Op, true
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
	modules    map[string]*tlago.Module
	instances  map[string]tlago.Instance
	locals     map[string]semanticCorpusLevel
	cache      map[string]semanticCorpusLevel
	inProgress map[string]bool
}

func newSemanticCorpusLeveler(spec *tlago.Spec) *semanticCorpusLeveler {
	leveler := &semanticCorpusLeveler{
		decls:      map[string]tlago.DeclarationKind{},
		defs:       map[string]tlago.Definition{},
		modules:    spec.Modules,
		instances:  map[string]tlago.Instance{},
		locals:     map[string]semanticCorpusLevel{},
		cache:      map[string]semanticCorpusLevel{},
		inProgress: map[string]bool{},
	}
	for _, mod := range spec.Modules {
		for _, inst := range mod.Instances {
			if qualifier := semanticCorpusInstanceQualifier(inst); qualifier != "" {
				leveler.instances[qualifier] = inst
			}
		}
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

func (l *semanticCorpusLeveler) child() *semanticCorpusLeveler {
	child := &semanticCorpusLeveler{
		decls:      l.decls,
		defs:       map[string]tlago.Definition{},
		modules:    l.modules,
		instances:  l.instances,
		locals:     map[string]semanticCorpusLevel{},
		cache:      map[string]semanticCorpusLevel{},
		inProgress: map[string]bool{},
	}
	for name, def := range l.defs {
		child.defs[name] = def
	}
	for name, level := range l.locals {
		child.locals[name] = level
	}
	return child
}

func (l *semanticCorpusLeveler) level(expr tlago.Expr) semanticCorpusLevel {
	switch e := expr.(type) {
	case nil:
		return semanticCorpusConstantLevel
	case *tlago.IdentExpr:
		if level, ok := l.locals[e.Name]; ok {
			return level
		}
		if level, ok := l.instanceDefinitionLevel(e.Name); ok {
			return level
		}
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
		opLevel := l.operatorLevel(e.Op)
		switch e.Op {
		case "'", "UNCHANGED":
			return maxSemanticCorpusLevel(semanticCorpusActionLevel, opLevel, l.level(e.Expr))
		case "[]", "<>":
			return semanticCorpusTemporalLevel
		default:
			return maxSemanticCorpusLevel(opLevel, l.level(e.Expr))
		}
	case *tlago.BinaryExpr:
		level := maxSemanticCorpusLevel(l.operatorLevel(e.Op), l.level(e.Left), l.level(e.Right))
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
		child := l.child()
		for _, def := range e.Definitions {
			child.defs[def.Name] = def
		}
		level := child.level(e.Body)
		for _, def := range e.Definitions {
			level = maxSemanticCorpusLevel(level, child.level(def.Expr))
		}
		return level
	case *tlago.QuantifierExpr:
		child := l.child()
		child.locals[e.Var] = semanticCorpusBoundVarLevel(e)
		level := maxSemanticCorpusLevel(l.level(e.Set), child.level(e.Body))
		switch e.Kind {
		case "\\AA", "\\EE", "TEMPORAL_FORALL", "TEMPORAL_EXISTS":
			return maxSemanticCorpusLevel(semanticCorpusTemporalLevel, level)
		default:
			return level
		}
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

func semanticCorpusInstanceQualifier(inst tlago.Instance) string {
	if inst.Name != "" {
		return inst.Name
	}
	return inst.Module
}

func (l *semanticCorpusLeveler) instanceDefinitionLevel(name string) (semanticCorpusLevel, bool) {
	bang := strings.LastIndex(name, "!")
	if bang <= 0 || bang+1 >= len(name) {
		return "", false
	}
	qualifier, member := name[:bang], name[bang+1:]
	inst, ok := l.instances[qualifier]
	if !ok {
		return "", false
	}
	target := l.modules[inst.Module]
	if target == nil {
		return "", false
	}
	var targetDef *tlago.Definition
	for i := range target.Definitions {
		if target.Definitions[i].Name == member {
			targetDef = &target.Definitions[i]
			break
		}
	}
	if targetDef == nil {
		return "", false
	}
	child := l.child()
	for _, subst := range inst.SubstitutionList {
		if subst.Name == "" || subst.Expr == nil {
			continue
		}
		child.defs[subst.Name] = tlago.Definition{Name: subst.Name, Expr: subst.Expr, Pos: subst.Pos}
	}
	return child.definitionLevel(name, *targetDef), true
}

func (l *semanticCorpusLeveler) operatorLevel(name string) semanticCorpusLevel {
	if level, ok := l.locals[name]; ok {
		return level
	}
	if name == "-." {
		if level, ok := l.locals["-"]; ok {
			return level
		}
	}
	if def, ok := l.defs[name]; ok {
		return l.definitionLevel(name, def)
	}
	return semanticCorpusConstantLevel
}

func semanticCorpusBoundVarLevel(e *tlago.QuantifierExpr) semanticCorpusLevel {
	if e != nil && e.LevelKnown {
		return semanticCorpusLevelFromTLAInt(e.Level)
	}
	return semanticCorpusConstantLevel
}

func semanticCorpusBoundLevel(bound tlago.BoundVar) semanticCorpusLevel {
	if bound.LevelKnown {
		return semanticCorpusLevelFromTLAInt(bound.Level)
	}
	return semanticCorpusConstantLevel
}

func semanticCorpusLevelFromTLAInt(level int) semanticCorpusLevel {
	switch level {
	case 1:
		return semanticCorpusVariableLevel
	case 2:
		return semanticCorpusActionLevel
	case 3:
		return semanticCorpusTemporalLevel
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

func walkSemanticCorpusExprWithLeveler(expr tlago.Expr, leveler *semanticCorpusLeveler, visit func(tlago.Expr, *semanticCorpusLeveler)) {
	if expr == nil {
		return
	}
	visit(expr, leveler)
	switch e := expr.(type) {
	case *tlago.UnaryExpr:
		walkSemanticCorpusExprWithLeveler(e.Expr, leveler, visit)
	case *tlago.BinaryExpr:
		walkSemanticCorpusExprWithLeveler(e.Left, leveler, visit)
		walkSemanticCorpusExprWithLeveler(e.Right, leveler, visit)
	case *tlago.CallExpr:
		walkSemanticCorpusExprWithLeveler(e.Callee, leveler, visit)
		for _, arg := range e.Args {
			walkSemanticCorpusExprWithLeveler(arg, leveler, visit)
		}
	case *tlago.IfExpr:
		walkSemanticCorpusExprWithLeveler(e.Cond, leveler, visit)
		walkSemanticCorpusExprWithLeveler(e.Then, leveler, visit)
		walkSemanticCorpusExprWithLeveler(e.Else, leveler, visit)
	case *tlago.LetExpr:
		child := leveler.child()
		for _, def := range e.Definitions {
			child.defs[def.Name] = def
			walkSemanticCorpusExprWithLeveler(def.Expr, child, visit)
		}
		walkSemanticCorpusExprWithLeveler(e.Body, child, visit)
	case *tlago.QuantifierExpr:
		walkSemanticCorpusExprWithLeveler(e.Set, leveler, visit)
		child := leveler.child()
		child.locals[e.Var] = semanticCorpusBoundVarLevel(e)
		walkSemanticCorpusExprWithLeveler(e.Body, child, visit)
	case *tlago.CaseExpr:
		for _, arm := range e.Arms {
			walkSemanticCorpusExprWithLeveler(arm.Test, leveler, visit)
			walkSemanticCorpusExprWithLeveler(arm.Value, leveler, visit)
		}
		walkSemanticCorpusExprWithLeveler(e.Other, leveler, visit)
	case *tlago.ChooseExpr:
		walkSemanticCorpusExprWithLeveler(e.Set, leveler, visit)
		child := leveler.child()
		child.locals[e.Var] = semanticCorpusConstantLevel
		walkSemanticCorpusExprWithLeveler(e.Body, child, visit)
	case *tlago.TupleExpr:
		for _, elem := range e.Elems {
			walkSemanticCorpusExprWithLeveler(elem, leveler, visit)
		}
	case *tlago.SetExpr:
		for _, elem := range e.Elems {
			walkSemanticCorpusExprWithLeveler(elem, leveler, visit)
		}
	case *tlago.RecordExpr:
		for _, field := range e.Fields {
			walkSemanticCorpusExprWithLeveler(field.Value, leveler, visit)
		}
	case *tlago.RecordComponentExpr:
		walkSemanticCorpusExprWithLeveler(e.Record, leveler, visit)
	case *tlago.RecordSetExpr:
		for _, field := range e.Fields {
			walkSemanticCorpusExprWithLeveler(field.Set, leveler, visit)
		}
	case *tlago.FunctionExpr:
		child := leveler.child()
		for _, bound := range e.Bounds {
			walkSemanticCorpusExprWithLeveler(bound.Set, leveler, visit)
			child.locals[bound.Name] = semanticCorpusBoundLevel(bound)
		}
		walkSemanticCorpusExprWithLeveler(e.Body, child, visit)
	case *tlago.FunctionAppExpr:
		walkSemanticCorpusExprWithLeveler(e.Function, leveler, visit)
		for _, arg := range e.Args {
			walkSemanticCorpusExprWithLeveler(arg, leveler, visit)
		}
	case *tlago.ExceptExpr:
		walkSemanticCorpusExprWithLeveler(e.Base, leveler, visit)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					walkSemanticCorpusExprWithLeveler(index, leveler, visit)
				}
			}
			walkSemanticCorpusExprWithLeveler(spec.Value, leveler, visit)
		}
	case *tlago.LabelExpr:
		walkSemanticCorpusExprWithLeveler(e.Body, leveler, visit)
	case *tlago.ActionExpr:
		walkSemanticCorpusExprWithLeveler(e.Action, leveler, visit)
		walkSemanticCorpusExprWithLeveler(e.Subscript, leveler, visit)
	case *tlago.FairnessExpr:
		walkSemanticCorpusExprWithLeveler(e.Subscript, leveler, visit)
		walkSemanticCorpusExprWithLeveler(e.Action, leveler, visit)
	case *tlago.FunctionSetExpr:
		walkSemanticCorpusExprWithLeveler(e.Domain, leveler, visit)
		walkSemanticCorpusExprWithLeveler(e.Range, leveler, visit)
	case *tlago.SetComprehensionExpr:
		child := leveler.child()
		for _, bound := range e.Bounds {
			walkSemanticCorpusExprWithLeveler(bound.Set, leveler, visit)
			child.locals[bound.Name] = semanticCorpusBoundLevel(bound)
		}
		walkSemanticCorpusExprWithLeveler(e.Element, child, visit)
		walkSemanticCorpusExprWithLeveler(e.Predicate, child, visit)
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
