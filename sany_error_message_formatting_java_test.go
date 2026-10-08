// Copyright (c) 2026 NVIDIA Corp. All rights reserved.
// SPDX-License-Identifier: MIT
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"strings"
	"testing"
)

func TestTestErrorMessageFormatting_testPercentSignInMessageTextIsNotAFormatSpecifier(t *testing.T) {
	log := &sanyErrors{}
	message := "Couldn't resolve infix operator symbol `%'."
	log.addMessage(originalErrorCode(4004), &tlc.NullSourceLocation, message)
	if got := log.getErrorDetails()[0].getMessage(); got != message {
		t.Fatalf("message = %q, want %q", got, message)
	}
}
func TestTestErrorMessageFormatting_testPercentSignInMessageParameterIsNotAFormatSpecifier(t *testing.T) {
	log := &sanyErrors{}
	log.addMessage(originalErrorCode(4004), &tlc.NullSourceLocation, "Couldn't resolve infix operator symbol `%s'.", "%%")
	if got, want := log.getErrorDetails()[0].getMessage(), "Couldn't resolve infix operator symbol `%%'."; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}
func processOriginalFormattingModule(t *testing.T, body string) Diagnostics {
	t.Helper()
	module := "---- MODULE Test ----\n" + body + "\n====\n"
	spec, parse := ParseSanySpecSource("Test.tla", module, LoadOptions{})
	requireNoErrors(t, parse)
	// AbortException is unexpected in the original helper and must propagate.
	return GenerateSanySpec(spec)
}
func requireOriginalFormattingErrorContains(t *testing.T, log Diagnostics, expected string) {
	t.Helper()
	found := false
	var rendered strings.Builder
	for _, diagnostic := range log.Errors() {
		message := sanyJavaErrorDetails(diagnostic)
		rendered.WriteString(message)
		rendered.WriteByte('\n')
		found = found || strings.Contains(message, expected)
	}
	if !found {
		t.Fatalf("No error mentions %s; got:\n%s", expected, rendered.String())
	}
}
func TestTestErrorMessageFormatting_testUnresolvedPercentOperator(t *testing.T) {
	requireOriginalFormattingErrorContains(t, processOriginalFormattingModule(t, "op == x % y"), "%")
}
func TestTestErrorMessageFormatting_testUnresolvedNonfixPercentOperator(t *testing.T) {
	requireOriginalFormattingErrorContains(t, processOriginalFormattingModule(t, "op == A!B!%(x, y)"), "%")
}
func TestTestErrorMessageFormatting_testUnresolvedDoublePercentOperatorIsNotRenamed(t *testing.T) {
	requireOriginalFormattingErrorContains(t, processOriginalFormattingModule(t, "op == x %% y"), "%%")
}
