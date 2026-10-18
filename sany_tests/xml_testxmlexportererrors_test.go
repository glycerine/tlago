package sany_tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from TestXMLExporterErrors.java, retaining library and CLI contracts.
func TestTestXMLExporterErrors_testHelpReturnsOk(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := tlago.XMLModuleToXML([]string{"-help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if code := tlago.RunCLI([]string{"sany-xml", "-help"}, &stdout, &stderr); code != int(tlago.XMLExporterOK) {
		t.Fatalf("help = %d, want OK", code)
	}
}
func TestTestXMLExporterErrors_testNoArgs(t *testing.T) {
	requireXMLExporterFailure(t, nil, tlago.XMLArgsParsingFailure, "", false)
}
func TestTestXMLExporterErrors_testIncludeDirWithoutSpec(t *testing.T) {
	requireXMLExporterFailure(t, []string{"-I", "SomeDir"}, tlago.XMLArgsParsingFailure, "", false)
}
func TestTestXMLExporterErrors_testCannotFindSpec(t *testing.T) {
	requireXMLExporterFailure(t, []string{"ThisModuleDoesNotExist.tla"}, tlago.XMLSpecParsingFailure, "", false)
}
func TestTestXMLExporterErrors_testSpecParseFailure(t *testing.T) {
	files := sanyTLAFilesUnder(t, sanyTestVectorPath("tla2sany", "semantic", "error_corpus"), func(path string) bool { return semanticErrorCorpusFilenameRE.MatchString(filepath.Base(path)) })
	requireXMLExporterFailure(t, []string{files[0]}, tlago.XMLSpecParsingFailure, "", false)
}
func TestTestXMLExporterErrors_testNullCharacterInStringLiteral(t *testing.T) {
	assertXMLUnrepresentableCharacter(t, `op == "a`+"\u0000"+`b"`)
}
func TestTestXMLExporterErrors_testNullCharacterInComment(t *testing.T) {
	assertXMLUnrepresentableCharacter(t, "\\* comment a\u0000b\nop == 1")
}
func assertXMLUnrepresentableCharacter(t *testing.T, body string) {
	t.Helper()
	requireXMLExporterFailure(t, []string{writeXMLExporterErrorModule(t, body)}, tlago.XMLUnrepresentableCharacter, "U+0000", true)
}
func requireXMLExporterFailure(t *testing.T, args []string, expected tlago.XMLExporterExitCode, message string, requireNonBug bool) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := tlago.XMLModuleToXML(args, &stdout, &stderr)
	failure, ok := err.(*tlago.XMLExportingException)
	if !ok {
		t.Fatalf("moduleToXML(%v) = %T (%v), want XMLExportingException", args, err, err)
	}
	if failure.Code != expected {
		t.Fatalf("moduleToXML(%v) code = %d, want %d: %v; stderr=%s", args, failure.Code, expected, failure, stderr.String())
	}
	if message != "" && !strings.Contains(failure.Message, message) {
		t.Fatalf("exception message %q lacks %q", failure.Message, message)
	}
	if requireNonBug && failure.Code.IsBug() {
		t.Fatalf("%d incorrectly classified as bug", failure.Code)
	}
	stdout.Reset()
	stderr.Reset()
	if code := tlago.RunCLI(append([]string{"sany-xml"}, args...), &stdout, &stderr); code != int(expected) {
		t.Fatalf("run(%v) = %d, want %d: %s", args, code, expected, stderr.String())
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
