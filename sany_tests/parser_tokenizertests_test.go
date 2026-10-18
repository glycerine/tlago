package sany_tests

import (
	"reflect"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/TokenizerTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTokenizerTests_runTokenizerCase(t *testing.T) {
	for _, tc := range []sanyTokenizerCase{
		{
			name:     "Empty module",
			start:    tlago.SanyLexDefault,
			end:      tlago.SanyLexSpec,
			input:    wrapSANYTestModule(),
			expected: specTokenKinds(),
		},
		{
			name:     "Constant",
			start:    tlago.SanyLexDefault,
			end:      tlago.SanyLexSpec,
			input:    wrapSANYTestModule("CONSTANT x"),
			expected: specTokenKinds(tlago.SanyTokenConstant, tlago.SanyTokenIdentifier),
		},
		{
			name:     "Constants",
			start:    tlago.SanyLexDefault,
			end:      tlago.SanyLexSpec,
			input:    wrapSANYTestModule("CONSTANTS x, y"),
			expected: specTokenKinds(tlago.SanyTokenConstant, tlago.SanyTokenIdentifier, tlago.SanyTokenComma, tlago.SanyTokenIdentifier),
		},
		{
			name:     "Simple opdef",
			start:    tlago.SanyLexDefault,
			end:      tlago.SanyLexSpec,
			input:    wrapSANYTestModule("op == 0"),
			expected: specTokenKinds(tlago.SanyTokenIdentifier, tlago.SanyTokenDef, tlago.SanyTokenNumberLiteral),
		},
		{
			name:     "Pragma mode",
			start:    tlago.SanyLexDefault,
			end:      tlago.SanyLexPragma,
			input:    "--->",
			expected: []tlago.SanyTokenKind{tlago.SanyTokenBeginPragma, tlago.SanyTokenEOF},
		},
		{
			name:     "Pragma input",
			start:    tlago.SanyLexDefault,
			end:      tlago.SanyLexPragma,
			input:    "---> id1 1 id2 2 id3 3",
			expected: []tlago.SanyTokenKind{tlago.SanyTokenBeginPragma, tlago.SanyTokenIdentifier, tlago.SanyTokenNumber, tlago.SanyTokenIdentifier, tlago.SanyTokenNumber, tlago.SanyTokenIdentifier, tlago.SanyTokenNumber, tlago.SanyTokenEOF},
		},
		{
			name:     "Pragma/spec",
			start:    tlago.SanyLexDefault,
			end:      tlago.SanyLexSpec,
			input:    "---> 2 ---- MODULE",
			expected: []tlago.SanyTokenKind{tlago.SanyTokenBeginPragma, tlago.SanyTokenNumber, tlago.SanyTokenBm2, tlago.SanyTokenEOF},
		},
		{
			name:     "EOL comment",
			start:    tlago.SanyLexInEOLComment,
			end:      tlago.SanyLexSpec,
			input:    "\n",
			expected: []tlago.SanyTokenKind{tlago.SanyTokenEOF},
		},
		{
			name:     "Block comment",
			start:    tlago.SanyLexInComment,
			end:      tlago.SanyLexSpec,
			input:    "*)",
			expected: []tlago.SanyTokenKind{tlago.SanyTokenEOF},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tokenizer := tlago.NewSanyTokenManager("Test.tla", tc.input)
			if got := tokenizer.State(); got != tlago.SanyLexDefault {
				t.Fatalf("initial lexer state = %s, want %s", got, tlago.SanyLexDefault)
			}
			tokenizer.SwitchTo(tc.start)
			if got := tokenizer.State(); got != tc.start {
				t.Fatalf("start lexer state = %s, want %s", got, tc.start)
			}
			tokens, diags := tokenizer.LexAll()
			requireNoSANYDiagnostics(t, "tokenize", diags)
			if got := tokenizer.State(); got != tc.end {
				t.Fatalf("end lexer state = %s, want %s", got, tc.end)
			}
			if got := sanyTokenKinds(tokens); !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("token kinds = %#v, want %#v", got, tc.expected)
			}
		})
	}
}

type sanyTokenizerCase struct {
	name     string
	start    tlago.SanyLexState
	end      tlago.SanyLexState
	input    string
	expected []tlago.SanyTokenKind
}

func wrapSANYTestModule(definitions ...string) string {
	units := ""
	for i, def := range definitions {
		if i > 0 {
			units += "\n"
		}
		units += def
	}
	return "---- MODULE Test ----\n" + units + "\n===="
}

func specTokenKinds(kinds ...tlago.SanyTokenKind) []tlago.SanyTokenKind {
	out := []tlago.SanyTokenKind{tlago.SanyTokenBm1, tlago.SanyTokenIdentifier, tlago.SanyTokenSeparator}
	out = append(out, kinds...)
	out = append(out, tlago.SanyTokenEndModule, tlago.SanyTokenEOF)
	return out
}

func sanyTokenKinds(tokens []*tlago.SanyToken) []tlago.SanyTokenKind {
	out := make([]tlago.SanyTokenKind, 0, len(tokens))
	for _, tok := range tokens {
		out = append(out, tok.Kind)
	}
	return out
}
