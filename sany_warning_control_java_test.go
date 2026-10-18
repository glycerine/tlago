package tlago

import (
	"path/filepath"
	"strings"
	"testing"
)

// WarningControlTest.java's settings tests run SANY.parse and observe only
// WARNING/ERROR output, matching RecordedSanyOutput(LogLevel.WARNING).
func originalWarningControlParse(settings *SanySettings) (SanyExitCode, map[Severity][]string) {
	output := map[Severity][]string{}
	_, code := ParseSanyWithSettings(filepath.Join("sany_tests", "test_vectors", "tla2sany", "semantic", "error_corpus", "W4802_Pre_Test.tla"), LoadOptions{}, settings, func(level Severity, message string) {
		output[level] = append(output[level], message)
	})
	return code, output
}

func TestWarningControlTest_testWarningAppearsWithDefaultSettings(t *testing.T) {
	code, output := originalWarningControlParse(DefaultSanySettings())
	if code != SanyExitOK {
		t.Fatalf("SANY.parse = %d, want OK", code)
	}
	if !strings.Contains(strings.Join(output[SeverityWarning], ""), "Foo") {
		t.Fatalf("WARNING output lacks Foo: %v", output)
	}
}

func TestWarningControlTest_testSuppressMessagesViaSettings(t *testing.T) {
	code4802 := SanyErrorCodeFromStandardValue(4802)
	settings := NewSanySettingsWithCodes(true, true, true, true, []SanyErrorCode{code4802}, nil)
	code, output := originalWarningControlParse(settings)
	if code != SanyExitOK {
		t.Fatalf("SANY.parse = %d, want OK", code)
	}
	if len(output[SeverityWarning]) != 0 {
		t.Fatalf("suppressed WARNING output: %v", output[SeverityWarning])
	}
}

func TestWarningControlTest_testMessagesAsErrorsViaSettings(t *testing.T) {
	code4802 := SanyErrorCodeFromStandardValue(4802)
	settings := NewSanySettingsWithCodes(true, true, true, true, nil, []SanyErrorCode{code4802})
	code, output := originalWarningControlParse(settings)
	if code != SanyExitSemanticFailure {
		t.Fatalf("SANY.parse = %d, want SEMANTIC_ANALYSIS_OR_LEVEL_CHECKING_FAILURE", code)
	}
	if !strings.Contains(strings.Join(output[SeverityError], ""), "Warning treated as error") {
		t.Fatalf("ERROR output lacks elevation message: %v", output)
	}
}
