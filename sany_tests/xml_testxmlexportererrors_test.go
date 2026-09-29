package sany_tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterErrors.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestXMLExporterErrors_testHelpReturnsOk(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tlago.RunCLI([]string{"sany-xml", "-help"}, &stdout, &stderr)
	if code != tlago.ExitOK {
		t.Fatalf("sany-xml -help exit = %d, want %d; stderr=%s", code, tlago.ExitOK, stderr.String())
	}
}

func TestTestXMLExporterErrors_testNoArgs(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"sany-xml"}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("sany-xml with no args exit = %d, want %d; stderr=%s", code, tlago.ExitToolFailure, stderr.String())
	}
}

func TestTestXMLExporterErrors_testIncludeDirWithoutSpec(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"sany-xml", "-I", "SomeDir"}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("sany-xml include without spec exit = %d, want %d; stderr=%s", code, tlago.ExitToolFailure, stderr.String())
	}
}

func TestTestXMLExporterErrors_testCannotFindSpec(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"sany-xml", "ThisModuleDoesNotExist.tla"}, nil, &stderr)
	if code != tlago.ExitSyntaxFailure {
		t.Fatalf("sany-xml missing spec exit = %d, want %d; stderr=%s", code, tlago.ExitSyntaxFailure, stderr.String())
	}
}

func TestTestXMLExporterErrors_testSpecParseFailure(t *testing.T) {
	var stderr bytes.Buffer
	path := sanyTestVectorPath("tla2sany", "semantic", "error_corpus", "E4200_Test.tla")
	code := tlago.RunCLI([]string{"sany-xml", path}, nil, &stderr)
	if code == tlago.ExitOK {
		t.Fatalf("sany-xml parse/semantic failure exit = %d, want failure", code)
	}
}

func TestTestXMLExporterErrors_testNullCharacterInStringLiteral(t *testing.T) {
	assertXMLUnrepresentableCharacter(t, `op == "a`+"\u0000"+`b"`)
}

func TestTestXMLExporterErrors_testNullCharacterInComment(t *testing.T) {
	assertXMLUnrepresentableCharacter(t, "\\* comment a\u0000b\nop == 1")
}

func assertXMLUnrepresentableCharacter(t *testing.T, body string) {
	t.Helper()
	modulePath := writeXMLExporterErrorModule(t, body)
	_, diags := xmlForUnrepresentableStringFixture(modulePath)
	if !diags.HasErrors() {
		t.Fatalf("expected XML export to reject unrepresentable character in:\n%s", body)
	}
	if got, want := diags.Error(), "U+0000"; !strings.Contains(got, want) {
		t.Fatalf("XML diagnostics missing %q\n%s", want, got)
	}
}

func writeXMLExporterErrorModule(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "SanyTest.tla")
	source := "---- MODULE SanyTest ----\n" + body + "\n====\n"
	if err := os.WriteFile(modulePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return modulePath
}
