package tlago

import "testing"

func TestSanyTokenManagerBehaviors(t *testing.T) {
	t.Run("enters spec mode at the first module header", func(t *testing.T) {
		tokens, diags := SanyTokenize("M.tla", "toolbox prose\n---- MODULE M ----\n====")
		requireNoErrors(t, diags)
		requireSanyTokenNames(t, tokens, "_BM1", "IDENTIFIER", "SEPARATOR", "END_MODULE", "EOF")
		if tokens[0].Image != "---- MODULE" || tokens[1].Image != "M" {
			t.Fatalf("unexpected module tokens: %#v %#v", tokens[0], tokens[1])
		}
	})

	t.Run("keeps pragma tokens before the module", func(t *testing.T) {
		tokens, diags := SanyTokenize("M.tla", "---> 123 profiler\n---- MODULE M ----\n====")
		requireNoErrors(t, diags)
		requireSanyTokenNames(t, tokens, "BEGIN_PRAGMA", "NUMBER", "IDENTIFIER", "_BM2", "IDENTIFIER", "SEPARATOR", "END_MODULE", "EOF")
		if tokens[1].LexState != SanyLexPragma || tokens[3].LexState != SanyLexPragma {
			t.Fatalf("pragma lex states were not preserved: %s %s", tokens[1].LexState, tokens[3].LexState)
		}
	})

	t.Run("attaches line and nested block comments as special tokens", func(t *testing.T) {
		source := "---- MODULE M ----\n\\* one\n(* two (* deep *) done *)\nASSUME TRUE\n===="
		tokens, diags := SanyTokenize("M.tla", source)
		requireNoErrors(t, diags)
		assume := findSanyToken(t, tokens, "ASSUME")
		if assume.Special == nil || assume.Special.Next == nil {
			t.Fatalf("ASSUME token did not receive both special comments: %#v", assume)
		}
		if assume.Special.LexState != SanyLexInEOLComment || assume.Special.Next.LexState != SanyLexInComment {
			t.Fatalf("unexpected special lex states: %#v %#v", assume.Special, assume.Special.Next)
		}
	})

	t.Run("uses longest-match JavaCC literals before identifiers", func(t *testing.T) {
		source := `---- MODULE M ----
CONSTANTS WF_x WF_ Enabled ENABLED ∧ ⇒ \in ℕ
====`
		tokens, diags := SanyTokenize("M.tla", source)
		requireNoErrors(t, diags)
		requireSanyTokenNames(t, tokens,
			"_BM1", "IDENTIFIER", "SEPARATOR",
			"CONSTANT", "WF", "IDENTIFIER", "WF", "IDENTIFIER", "op_112", "AND", "op_infix_implies_uc", "IN", "IDENTIFIER",
			"END_MODULE", "EOF",
		)
	})

	t.Run("recognizes proof step lexemes without swallowing incomplete less-than expressions", func(t *testing.T) {
		source := "---- MODULE M ----\n<1>Foo <+>bar <2>Foo.. <3> <4\n===="
		tokens, diags := SanyTokenize("M.tla", source)
		requireNoErrors(t, diags)
		requireSanyTokenNames(t, tokens,
			"_BM1", "IDENTIFIER", "SEPARATOR",
			"ProofStepLexeme", "ProofImplicitStepLexeme", "ProofStepDotLexeme", "BareLevelLexeme", "op_85", "NUMBER_LITERAL",
			"END_MODULE", "EOF",
		)
	})

	t.Run("recognizes bare level QED proof steps", func(t *testing.T) {
		source := "---- MODULE M ----\n<1> QED\n===="
		tokens, diags := SanyTokenize("M.tla", source)
		requireNoErrors(t, diags)
		requireSanyTokenNames(t, tokens,
			"_BM1", "IDENTIFIER", "SEPARATOR",
			"BareLevelLexeme", "QED",
			"END_MODULE", "EOF",
		)
	})

	t.Run("expands tabs to Java SANY source columns", func(t *testing.T) {
		source := "---- MODULE M ----\nA ==\n\tTRUE\n        \tFALSE\n===="
		tokens, diags := SanyTokenize("M.tla", source)
		requireNoErrors(t, diags)
		truth := findSanyTokenImage(t, tokens, "TRUE")
		if truth.Begin.Column != 9 || truth.End.Column != 12 {
			t.Fatalf("TRUE columns = %d..%d, want 9..12", truth.Begin.Column, truth.End.Column)
		}
		falsity := findSanyTokenImage(t, tokens, "FALSE")
		if falsity.Begin.Column != 17 || falsity.End.Column != 21 {
			t.Fatalf("FALSE columns = %d..%d, want 17..21", falsity.Begin.Column, falsity.End.Column)
		}
	})

	t.Run("counts CRLF as one Java SANY source line", func(t *testing.T) {
		source := "---- MODULE M ----\r\nA == TRUE\r\n====\r\n"
		tokens, diags := SanyTokenize("M.tla", source)
		requireNoErrors(t, diags)

		defName := findSanyTokenImage(t, tokens, "A")
		if defName.Begin.Line != 2 || defName.Begin.Column != 1 {
			t.Fatalf("A position = %d:%d, want 2:1", defName.Begin.Line, defName.Begin.Column)
		}
		end := findSanyToken(t, tokens, "END_MODULE")
		if end.Begin.Line != 3 {
			t.Fatalf("END_MODULE line = %d, want 3", end.Begin.Line)
		}
	})
}

func requireSanyTokenNames(t *testing.T, tokens []*SanyToken, want ...string) {
	t.Helper()
	got := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		got = append(got, tok.Kind.JavaName())
	}
	if len(got) != len(want) {
		t.Fatalf("token names = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token names = %v, want %v", got, want)
		}
	}
}

func findSanyToken(t *testing.T, tokens []*SanyToken, name string) *SanyToken {
	t.Helper()
	for _, tok := range tokens {
		if tok.Kind.JavaName() == name {
			return tok
		}
	}
	t.Fatalf("token %s not found in %v", name, sanyTokenNames(tokens))
	return nil
}

func findSanyTokenImage(t *testing.T, tokens []*SanyToken, image string) *SanyToken {
	t.Helper()
	for _, tok := range tokens {
		if tok.Image == image {
			return tok
		}
	}
	t.Fatalf("token image %q not found in %v", image, sanyTokenImages(tokens))
	return nil
}

func sanyTokenNames(tokens []*SanyToken) []string {
	names := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		names = append(names, tok.Kind.JavaName())
	}
	return names
}

func sanyTokenImages(tokens []*SanyToken) []string {
	images := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		images = append(images, tok.Image)
	}
	return images
}
