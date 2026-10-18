package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestInstanceNode.java.
// Retains the original semantic-success and level-failure phase assertions.
func TestTestInstanceNode_testOperatorArgumentMinimumLevelDiagnostic(t *testing.T) {
	module := "---- MODULE Test ----\n" +
		"---- MODULE Inner ----\n" +
		"CONSTANT F(_, _)\n" +
		"op == F([]TRUE, 0)\n" +
		"====\n" +
		"INSTANCE Inner WITH F <- =\n" +
		"====\n"
	spec, parseDiags := tlago.ParseSanySpecSource("Test.tla", module, tlago.LoadOptions{})
	if parseDiags.HasErrors() {
		t.Fatal(parseDiags)
	}
	semanticLog := tlago.GenerateSanySpec(spec)
	if semanticLog.HasErrors() {
		t.Fatal(semanticLog)
	}
	levelOK, diags := tlago.CheckSanySpecLevels(spec)
	if levelOK {
		t.Fatal("level checking succeeded unexpectedly")
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
	if len(parameters) < 4 || parameters[1] != int32(1) || parameters[3] != int32(3) {
		t.Fatalf("The diagnostic should report a one-based argument position and its required level: %v", parameters)
	}
}
