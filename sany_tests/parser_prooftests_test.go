package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/ProofTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestProofTests_test(t *testing.T) {
	t.Skip("tla2sany wip")

	for _, tc := range []struct {
		proof    string
		expected sanyProofAST
	}{
		{
			proof:    "PROOF <1>a A <1>b QED",
			expected: sanyProof(1, sanyProofStep("a"), sanyProofStep("b")),
		},
		{
			proof:    "<1>a A <1>b QED",
			expected: sanyProof(1, sanyProofStep("a"), sanyProofStep("b")),
		},
		{
			proof: "<1>a. A PROOF <+> A <*> QED <1>b.. QED",
			expected: sanyProof(1,
				sanyProofStep("a", sanyProof(2,
					sanyProofStep(""),
					sanyProofStep(""),
				)),
				sanyProofStep("b"),
			),
		},
		{
			proof: "<*> A <+> A <*> QED <+> QED <*> QED",
			expected: sanyProof(0,
				sanyProofStep("", sanyProof(1,
					sanyProofStep(""),
					sanyProofStep("", sanyProof(2,
						sanyProofStep(""),
					)),
				)),
				sanyProofStep(""),
			),
		},
	} {
		t.Run(tc.proof, func(t *testing.T) {
			source := "---- MODULE Test ----\nTHEOREM T == TRUE\n" + tc.proof + "\n===="
			root, diags := tlago.ParseSanySyntax("Test.tla", source)
			requireNoSANYDiagnostics(t, "parse", diags)
			actual := findSANYNodeByJavaKind(root, "N_Proof")
			matchSANYProofAST(t, tc.expected, actual)
		})
	}
}

type sanyProofKind int

const (
	sanyProofKindProof sanyProofKind = iota
	sanyProofKindStep
)

type sanyProofAST struct {
	kind     sanyProofKind
	name     string
	level    int
	children []sanyProofAST
}

func sanyProof(level int, children ...sanyProofAST) sanyProofAST {
	return sanyProofAST{kind: sanyProofKindProof, level: level, children: children}
}

func sanyProofStep(name string, children ...sanyProofAST) sanyProofAST {
	return sanyProofAST{kind: sanyProofKindStep, name: name, level: -1, children: children}
}

func matchSANYProofAST(t *testing.T, expected sanyProofAST, actual *tlago.SanySyntaxNode) {
	t.Helper()
	if actual == nil {
		t.Fatalf("missing proof AST node for %#v", expected)
	}
	if got := actual.GetProofLevel(); got != expected.level {
		t.Fatalf("%s proof level = %d, want %d", actual.Kind.JavaName(), got, expected.level)
	}
	actualChildren := actual.GetHeirs()
	switch expected.kind {
	case sanyProofKindProof:
		requireSANYNodeKind(t, actual, "N_Proof")
		i := 0
		if len(actualChildren) > 0 && actualChildren[0].Kind == tlago.SanyNodeKind(tlago.SanyTokenProof) {
			i = 1
		}
		for _, expectedChild := range expected.children {
			if i >= len(actualChildren) {
				t.Fatalf("proof has %d children, want child %#v", len(actualChildren), expectedChild)
			}
			matchSANYProofAST(t, expectedChild, actualChildren[i])
			i++
		}
		if i != len(actualChildren) {
			t.Fatalf("proof has %d children, consumed %d", len(actualChildren), i)
		}
	case sanyProofKindStep:
		requireSANYNodeKind(t, actual, "N_ProofStep")
		if got := sanyProofStepName(actual); got != expected.name {
			t.Fatalf("proof step name = %q, want %q", got, expected.name)
		}
		actualChild := actualChildren[len(actualChildren)-1]
		if actualChild.Kind.JavaName() == "N_Proof" {
			if len(expected.children) != 1 {
				t.Fatalf("proof step has nested proof, want %d expected nested proofs", len(expected.children))
			}
			matchSANYProofAST(t, expected.children[0], actualChild)
		}
	}
}

func sanyProofStepName(node *tlago.SanySyntaxNode) string {
	heirs := node.GetHeirs()
	if len(heirs) == 0 {
		return ""
	}
	image := heirs[0].Image
	if idx := strings.IndexByte(image, '>'); idx >= 0 {
		image = image[idx+1:]
	}
	if idx := strings.IndexByte(image, '.'); idx >= 0 {
		image = image[:idx]
	}
	return image
}
