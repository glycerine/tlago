package tlago

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const testOperatorsPath = "test_vectors/java-sany/src/tla2sany/parser/Operators.java"

func TestSanyOperatorTableMatchesJavaReference(t *testing.T) {
	source := activeJavaOperatorSource(readTestFile(t, testOperatorsPath))
	operatorRE := regexp.MustCompile(`new\s+Operator\(\s*"((?:\\.|[^"\\])*)"\s*,\s*([0-9]+)\s*,\s*([0-9]+)\s*,\s*Operators\.(assocNone|assocLeft|assocRight)\s*,\s*Operators\.(nofix|prefix|postfix|infix|nfix)\s*\)`)
	matches := operatorRE.FindAllStringSubmatch(source, -1)
	if len(SanyCanonicalOperators) != len(matches) {
		t.Fatalf("canonical operator count = %d, Java count = %d", len(SanyCanonicalOperators), len(matches))
	}
	for _, match := range matches {
		symbol := unquoteTestJavaString(t, `"`+match[1]+`"`)
		op, ok := GetSanyOperator(symbol)
		if !ok {
			t.Fatalf("missing operator %q", symbol)
		}
		low, _ := strconv.Atoi(match[2])
		high, _ := strconv.Atoi(match[3])
		if op.Symbol != symbol || op.LowPrecedence != low || op.HighPrecedence != high {
			t.Fatalf("operator %q = %#v, want low/high %d/%d", symbol, op, low, high)
		}
	}
}

func activeJavaOperatorSource(source string) string {
	var b strings.Builder
	for _, line := range strings.Split(source, "\n") {
		if strings.Contains(line, "//new Operator") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func TestSanyOperatorSynonymsAndRelations(t *testing.T) {
	land, ok := GetSanyOperator("∧")
	if !ok {
		t.Fatal("unicode conjunction synonym not found")
	}
	if land.Symbol != "\\land" || land.Fixity != SanyOperatorInfix || !land.AssocLeft() {
		t.Fatalf("∧ resolved to %#v", land)
	}
	times, ok := GetSanyOperator("×")
	if !ok {
		t.Fatal("unicode times synonym not found")
	}
	if times.Symbol != "\\times" || !times.IsNfix() {
		t.Fatalf("× resolved to %#v", times)
	}
	box, ok := GetSanyOperator("□")
	if !ok {
		t.Fatal("box synonym not found")
	}
	if box.Symbol != "[]" || box.LowPrecedence != 40 || box.HighPrecedence != 150 || !box.IsPrefix() {
		t.Fatalf("□ resolved to %#v", box)
	}
	implies, _ := GetSanyOperator("=>")
	eq, _ := GetSanyOperator("=")
	if !SanyOperatorPrec(implies, eq) {
		t.Fatalf("=> should precede = in SANY precedence relation")
	}
	if !SanyOperatorSucc(eq, implies) {
		t.Fatalf("= should succeed => in SANY precedence relation")
	}
}

func TestSanyOperatorStackReductions(t *testing.T) {
	t.Run("reduces by Java precedence and left associativity", func(t *testing.T) {
		stack := NewSanyOperatorStack()
		stack.NewStack()
		pushExpr := func(name string) {
			stack.Push(NewSanyTokenNode(&SanyToken{Kind: SanyTokenIdentifier, Image: name}), nil)
		}
		pushOp := func(symbol string) {
			op, ok := GetSanyOperator(symbol)
			if !ok {
				t.Fatalf("operator %s not found", symbol)
			}
			stack.Push(NewSanyTokenNode(&SanyToken{Kind: SanyTokenKindByName["op_81"], Image: symbol}), &op)
			if err := stack.ReduceStack(); err != nil {
				t.Fatalf("reduce after %s: %v", symbol, err)
			}
		}

		pushExpr("A")
		pushOp("+")
		pushExpr("B")
		pushOp("*")
		pushExpr("C")
		pushOp("+")
		pushExpr("D")
		root, err := stack.FinalReduce()
		if err != nil {
			t.Fatal(err)
		}
		if root.Kind.JavaName() != "N_InfixExpr" {
			t.Fatalf("root = %s, want N_InfixExpr", root.Kind.JavaName())
		}
		if got := root.GetHeirs()[1].Image; got != "+" {
			t.Fatalf("root operator = %q, want +", got)
		}
		left := root.GetHeirs()[0]
		if left.Kind.JavaName() != "N_InfixExpr" || left.GetHeirs()[1].Image != "+" {
			t.Fatalf("left branch = %#v", left)
		}
		mul := left.GetHeirs()[2]
		if mul.Kind.JavaName() != "N_InfixExpr" || mul.GetHeirs()[1].Image != "*" {
			t.Fatalf("multiplication branch = %#v", mul)
		}
	})

	t.Run("rejects non-associative chains", func(t *testing.T) {
		stack := NewSanyOperatorStack()
		stack.NewStack()
		eq, _ := GetSanyOperator("=")
		stack.Push(NewSanyTokenNode(&SanyToken{Kind: SanyTokenIdentifier, Image: "A"}), nil)
		stack.Push(NewSanyTokenNode(&SanyToken{Kind: SanyTokenEquals, Image: "="}), &eq)
		if err := stack.ReduceStack(); err != nil {
			t.Fatal(err)
		}
		stack.Push(NewSanyTokenNode(&SanyToken{Kind: SanyTokenIdentifier, Image: "B"}), nil)
		stack.Push(NewSanyTokenNode(&SanyToken{Kind: SanyTokenEquals, Image: "="}), &eq)
		if err := stack.ReduceStack(); err == nil {
			t.Fatal("expected precedence conflict for A = B = C")
		}
	})
}

func unquoteTestJavaString(t *testing.T, s string) string {
	t.Helper()
	var b strings.Builder
	for i := 1; i < len(s)-1; i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case '"', '\'', '\\':
			b.WriteByte(s[i])
		case 'u':
			value, err := strconv.ParseInt(s[i+1:i+5], 16, 32)
			if err != nil {
				t.Fatal(err)
			}
			b.WriteRune(rune(value))
			i += 4
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
