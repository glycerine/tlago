package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestInstanceNode.java.
// The source uses separate semantic-generation and level-check phases. This
// bridge currently combines them; keep that phase API work pending.
func TestTestInstanceNode_testOperatorArgumentMinimumLevelDiagnostic(t *testing.T) {
	module := "---- MODULE Test ----\n" +
		"---- MODULE Inner ----\n" +
		"CONSTANT F(_, _)\n" +
		"op == F([]TRUE, 0)\n" +
		"====\n" +
		"INSTANCE Inner WITH F <- =\n" +
		"====\n"
	_, diags := tlago.CheckSanySource("Test.tla", module)
	if !diags.HasErrors() {
		t.Fatal("expected INSTANCE substitution level diagnostic")
	}
	var errors, diagnostics tlago.Diagnostics
	for _, diagnostic := range diags {
		if diagnostic.Severity == tlago.SeverityError {
			errors = append(errors, diagnostic)
		}
		if diagnostic.Code == "E4246" {
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	if len(errors) != 1 {
		t.Fatalf("want one level error: %v", diags)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("want one INSTANCE_SUBSTITUTION_LEVEL_CONSTRAINT_NOT_MET: %v", diags)
	}
	parameters := diagnostics[0].SANYParameters
	if len(parameters) < 4 || parameters[1] != 1 || parameters[3] != 3 {
		t.Fatalf("The diagnostic should report a one-based argument position and its required level: %v", parameters)
	}
}
