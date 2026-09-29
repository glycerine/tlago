package sany_tests

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago"
)

func sanyTestVectorPath(parts ...string) string {
	all := append([]string{"test_vectors"}, parts...)
	return filepath.Join(all...)
}

func requireNoSANYDiagnostics(t *testing.T, phase string, diags tlago.Diagnostics) {
	t.Helper()
	if diags.HasErrors() {
		t.Fatalf("%s diagnostics:\n%s", phase, diags.Error())
	}
}

func findSANYNodeByJavaKind(root *tlago.SanySyntaxNode, kind string) *tlago.SanySyntaxNode {
	if root == nil {
		return nil
	}
	if root.Kind.JavaName() == kind {
		return root
	}
	for _, child := range root.GetHeirs() {
		if found := findSANYNodeByJavaKind(child, kind); found != nil {
			return found
		}
	}
	return nil
}

func requireSANYNodeKind(t *testing.T, node *tlago.SanySyntaxNode, kind string) {
	t.Helper()
	if node == nil {
		t.Fatalf("missing SANY node kind %s", kind)
	}
	if got := node.Kind.JavaName(); got != kind {
		t.Fatalf("SANY node kind = %s, want %s", got, kind)
	}
}
