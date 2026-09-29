package tlago

import (
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	testParserConstantsPath = "test_vectors/java-sany/src/tla2sany/parser/TLAplusParserConstants.java"
	testSyntaxConstantsPath = "test_vectors/java-sany/src/tla2sany/st/SyntaxTreeConstants.java"
	testJavaCCGrammarPath   = "test_vectors/java-sany/javacc/tla+.jj"
)

func TestSanyGeneratedTablesMatchJavaReferences(t *testing.T) {
	parserConstants := readTestFile(t, testParserConstantsPath)
	javaTokens := parseTestJavaConstantsBefore(t, parserConstants, "DEFAULT")
	if len(SanyTokenKinds) != len(javaTokens) {
		t.Fatalf("generated token count = %d, Java token count = %d", len(SanyTokenKinds), len(javaTokens))
	}
	for name, value := range javaTokens {
		kind, ok := SanyTokenKindByName[name]
		if !ok {
			t.Fatalf("missing generated token constant for Java %s", name)
		}
		if int(kind) != value {
			t.Fatalf("token %s = %d, want Java value %d", name, kind, value)
		}
	}
	if got, want := len(SanyTokenImages), javaTokens["UnnumberedStepLexeme"]+1; got != want {
		t.Fatalf("token image count = %d, want %d", got, want)
	}

	syntaxConstants := readTestFile(t, testSyntaxConstantsPath)
	javaNodes := parseTestJavaConstants(t, syntaxConstants)
	if len(SanySyntaxNodeKinds) != len(javaNodes) {
		t.Fatalf("generated syntax node count = %d, Java syntax node count = %d", len(SanySyntaxNodeKinds), len(javaNodes))
	}
	for name, value := range javaNodes {
		kind, ok := SanySyntaxNodeKindByName[name]
		if !ok {
			t.Fatalf("missing generated syntax node constant for Java %s", name)
		}
		if int(kind) != value {
			t.Fatalf("syntax node %s = %d, want Java value %d", name, kind, value)
		}
	}
}

func TestSanyReferenceGrammarArchitectureIsTracked(t *testing.T) {
	grammar := readTestFile(t, testJavaCCGrammarPath)
	for _, marker := range []string{
		"PARSER_BEGIN(TLAplusParser)",
		"TOKEN_MGR_DECLS",
		"OperatorStack.newStack()",
		"JunctionListContext",
		"SyntaxTreeNode\nExpression()",
	} {
		if !strings.Contains(grammar, marker) {
			t.Fatalf("JavaCC grammar is missing expected marker %q", marker)
		}
	}
	productionRE := regexp.MustCompile(`(?m)^(?:SyntaxTreeNode|void|Token|boolean|int)\s*\n?[A-Za-z][A-Za-z0-9_]*\s*\(`)
	if got := len(productionRE.FindAllString(grammar, -1)); got < 80 {
		t.Fatalf("grammar production count = %d, want at least 80", got)
	}
	if len(SanyGrammarProductions) < 80 {
		t.Fatalf("generated grammar production inventory = %d, want at least 80", len(SanyGrammarProductions))
	}
	for _, name := range []string{"CompilationUnit", "Module", "Body", "Expression", "ExtendableExpr", "PrimitiveExp", "Proof"} {
		if _, ok := SanyGrammarProductionByName[name]; !ok {
			t.Fatalf("generated grammar production inventory missing %s", name)
		}
	}

	wantStates := []string{"DEFAULT", "PRAGMA", "SPEC", "IN_COMMENT", "EMBEDDED", "IN_EOL_COMMENT"}
	if strings.Join(SanyLexStateNames, ",") != strings.Join(wantStates, ",") {
		t.Fatalf("lex states = %v, want %v", SanyLexStateNames, wantStates)
	}
}

func TestSanyParserProductionCoverageIsTracked(t *testing.T) {
	parserSource := readTestFile(t, "sany_parser.go")
	methodRE := regexp.MustCompile(`(?m)^func \(p \*SanyParser\) ([A-Za-z][A-Za-z0-9_]*)\(`)
	methods := map[string]bool{}
	for _, match := range methodRE.FindAllStringSubmatch(parserSource, -1) {
		methods[match[1]] = true
	}

	validStatus := map[SanyGrammarProductionCoverageStatus]bool{
		SanyProductionFolded:   true,
		SanyProductionToken:    true,
		SanyProductionDeferred: true,
	}
	productions := map[string]bool{}
	var missing []string
	var deferred []string
	for _, production := range SanyGrammarProductions {
		productions[production.Name] = true
		if methods[production.Name] {
			continue
		}
		coverage, ok := SanyGrammarProductionCoverageByName[production.Name]
		if !ok {
			missing = append(missing, production.Name)
			continue
		}
		if !validStatus[coverage.Status] {
			t.Fatalf("production %s has invalid coverage status %q", production.Name, coverage.Status)
		}
		if strings.TrimSpace(coverage.Implementation) == "" || strings.TrimSpace(coverage.Note) == "" {
			t.Fatalf("production %s coverage must name implementation and note", production.Name)
		}
		if coverage.Status == SanyProductionDeferred {
			deferred = append(deferred, production.Name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("grammar productions without parser method or coverage note: %s", strings.Join(missing, ", "))
	}
	if len(deferred) > 0 {
		sort.Strings(deferred)
		t.Fatalf("grammar productions still deferred: %s", strings.Join(deferred, ", "))
	}

	for name := range SanyGrammarProductionCoverageByName {
		if !productions[name] {
			t.Fatalf("coverage note names unknown JavaCC production %s", name)
		}
	}
}

func TestSanyScaffoldBehaviors(t *testing.T) {
	t.Run("tokens keep lexical state location and special-token chain", func(t *testing.T) {
		comment := &SanyToken{
			Kind:     SanyTokenKindByName["END_MODULE"],
			Image:    "\\* comment",
			Begin:    Position{File: "Spec.tla", Line: 3, Column: 1},
			End:      Position{File: "Spec.tla", Line: 3, Column: 11},
			LexState: SanyLexInEOLComment,
		}
		tok := &SanyToken{
			Kind:     SanyTokenKindByName["ASSUME"],
			Image:    "ASSUME",
			Begin:    Position{File: "Spec.tla", Line: 4, Column: 1},
			End:      Position{File: "Spec.tla", Line: 4, Column: 6},
			LexState: SanyLexSpec,
			Special:  comment,
		}
		if tok.Kind.JavaName() != "ASSUME" || tok.Range().Begin.Line != 4 || tok.Special.LexState != SanyLexInEOLComment {
			t.Fatalf("unexpected token scaffold: %#v", tok)
		}
	})

	t.Run("junction context follows SANY column rules", func(t *testing.T) {
		var ctx SanyJunctionListContext
		if !ctx.IsAboveCurrent(1) {
			t.Fatal("empty junction context should accept any column")
		}
		if err := ctx.StartNewJunctionList(3, SanyTokenAND); err != nil {
			t.Fatal(err)
		}
		if !ctx.IsNewBullet(3, SanyTokenAND) {
			t.Fatal("aligned conjunction bullet was not recognized")
		}
		if ctx.IsNewBullet(3, SanyTokenOR) {
			t.Fatal("disjunction bullet should not continue a conjunction list")
		}
		if ctx.IsAboveCurrent(2) {
			t.Fatal("column left of current junction should terminate expression parsing")
		}
	})

	t.Run("operator stack preserves SANY stack-of-stacks shape", func(t *testing.T) {
		stack := NewSanyOperatorStack()
		stack.NewStack()
		stack.Push(NewSanyNode(SanySyntaxNodeKindByName["N_IdentifierTuple"]), nil)
		if stack.PreInEmptyTop() {
			t.Fatal("expression on top should not look like prefix/infix operator")
		}
		stack.Push(NewSanyNode(SanySyntaxNodeKindByName["N_GenInfixOp"]), &SanyOperatorInfo{Symbol: "=", Fixity: SanyOperatorInfix})
		if !stack.PreInEmptyTop() {
			t.Fatal("infix operator on top should permit junction/open-expression lookahead")
		}
		if stack.CurrentSize() != 2 {
			t.Fatalf("stack size = %d, want 2", stack.CurrentSize())
		}
		if err := stack.PopStack(); err != nil {
			t.Fatal(err)
		}
	})
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func parseTestJavaConstants(t *testing.T, text string) map[string]int {
	t.Helper()
	return parseTestJavaConstantsBefore(t, text, "")
}

func parseTestJavaConstantsBefore(t *testing.T, text, marker string) map[string]int {
	t.Helper()
	if marker != "" {
		idx := strings.Index(text, "int "+marker+" =")
		if idx < 0 {
			t.Fatalf("marker %q not found", marker)
		}
		text = text[:idx]
	}
	re := regexp.MustCompile(`(?m)^\s*int\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*([0-9]+)\s*;`)
	consts := make(map[string]int)
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		value, err := strconv.Atoi(match[2])
		if err != nil {
			t.Fatal(err)
		}
		consts[match[1]] = value
	}
	return consts
}
