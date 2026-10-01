package tlago

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateREPLExpressionGeneratedSpecExtendsJavaModules(t *testing.T) {
	dir := t.TempDir()
	value, diags, err := EvaluateREPLExpression(`1 \in Nat`, REPLEvalOptions{
		TempDir:     dir,
		PrintOutput: io.Discard,
	})
	if err != nil {
		t.Fatalf("EvaluateREPLExpression returned error: %v", err)
	}
	if diags.HasErrors() {
		t.Fatalf("EvaluateREPLExpression diagnostics: %v", diags)
	}
	if value != "TRUE" {
		t.Fatalf("EvaluateREPLExpression value = %q, want TRUE", value)
	}

	spec, err := os.ReadFile(filepath.Join(dir, replSpecName+".tla"))
	if err != nil {
		t.Fatalf("reading generated REPL spec: %v", err)
	}
	if !strings.Contains(string(spec), "EXTENDS Naturals, Reals, Sequences, Bags, FiniteSets, TLC\n") {
		t.Fatalf("generated REPL spec does not mirror Java REPLSpecWriter EXTENDS line:\n%s", spec)
	}
}
