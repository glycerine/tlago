package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original TestContext.testDifferentSymbolClassesDiagnostic, using the actual
// context merge and retained ErrorDetails parameters rather than a spec proxy.
func TestJavaSanyTestContextDifferentSymbolClassesDiagnostic(t *testing.T) {
	name := "symbol"
	existingDefinition := newSanySemNullOpDefNode(name)
	incomingDeclaration := newSanySemOpDeclNode(name, sanyConstantDeclKind, constantLevel, 0, nil, tlc.NullSemanticNodeInstance.GetTreeNode())
	context := newSanyContext()
	context.addSymbol(existingDefinition)
	incoming := newSanyContext()
	incoming.addSymbol(incomingDeclaration)
	success, log := context.mergeExtendContext(incoming)
	if success {
		t.Fatal("mergeExtendContext returned true, want false")
	}
	var errors Diagnostics
	for _, diagnostic := range log {
		if diagnostic.Severity == SeverityError {
			errors = append(errors, diagnostic)
		}
	}
	if len(errors) != 1 {
		t.Fatalf("errors=%d, want 1: %v", len(errors), log)
	}
	diagnostic := errors[0]
	if diagnostic.Code != "E4224" {
		t.Fatalf("code=%s, want EXTENDED_MODULES_SYMBOL_UNIFICATION_CONFLICT", diagnostic.Code)
	}
	parameters := diagnostic.SANYParameters
	if len(parameters) < 3 {
		t.Fatalf("parameters=%v, want at least 3", parameters)
	}
	if parameters[0] != "declaration" || parameters[2] != "definition" {
		t.Fatalf("incoming/existing symbol kinds=%v/%v, want declaration/definition", parameters[0], parameters[2])
	}
}
