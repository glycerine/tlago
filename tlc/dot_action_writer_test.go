package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDotActionWriterPortedOutputShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "actions.dot")
	writer, err := NewDotActionWriter(path, "strict ")
	if err != nil {
		t.Fatalf("NewDotActionWriter: %v", err)
	}
	if err := writer.WriteAction(&Action{Name: "Init", IsInitPred: true}, 1); err != nil {
		t.Fatalf("WriteAction init: %v", err)
	}
	if err := writer.WriteAction(&Action{Name: "Next"}, 2); err != nil {
		t.Fatalf("WriteAction next: %v", err)
	}
	if err := writer.WriteEdge(1, 2); err != nil {
		t.Fatalf("WriteEdge unseen: %v", err)
	}
	if err := writer.WriteEdge(2, 1, 3.5); err != nil {
		t.Fatalf("WriteEdge weighted: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := string(raw)
	for _, want := range []string{
		"strict digraph ActionGraph {",
		"1 [shape=box,label=\"Init\",style = filled]",
		"2 [shape=box,label=\"Next\"]",
		"1 -> 2[color=\"green\",style=dotted];",
		"2 -> 1[penwidth=3.5];",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dot output missing %q:\n%s", want, text)
		}
	}
}
