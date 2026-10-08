// Copyright (c) 2024 Linux Foundation. All rights reserved.
// SPDX-License-Identifier: MIT
package tlago

import (
	"fmt"
	"github.com/glycerine/tlago/tlc"
	"reflect"
	"strings"
	"testing"
)

// TestErrors.java's changing location seed is local to each Go test, retaining
// the zero-coordinate case and formula without depending on test execution order.
func originalErrorsLocationGenerator() func() tlc.SourceLocation {
	seed := 0
	return func() tlc.SourceLocation {
		loc := tlc.NewSourceLocation(fmt.Sprintf("Test%d.tla", seed), seed*3, seed*5, seed*7, seed*11)
		seed++
		return loc
	}
}
func originalErrorCode(value int) SanyErrorCode {
	code, ok := SanyErrorCodeFromStandardValue(value)
	if !ok {
		panic("missing source error code")
	}
	return code
}
func originalExpectedDetail(code int, loc tlc.SourceLocation, message string) *sanyErrorDetails {
	return &sanyErrorDetails{code: originalErrorCode(code), location: loc, format: message}
}
func originalExpectedErrorStrings(details []*sanyErrorDetails) []string {
	result := make([]string, len(details))
	for i, d := range details {
		result[i] = d.location.String() + "\n\n" + d.format
	}
	return result
}
func requireOriginalErrorsAssertions(t *testing.T, log *sanyErrors, warnings, errors []*sanyErrorDetails, checkMessages bool) {
	t.Helper()
	expectedWarnings, expectedErrors := originalExpectedErrorStrings(warnings), originalExpectedErrorStrings(errors)
	if !reflect.DeepEqual(log.getWarnings(), expectedWarnings) {
		t.Fatalf("warnings = %#v, want %#v", log.getWarnings(), expectedWarnings)
	}
	if !reflect.DeepEqual(log.getErrors(), expectedErrors) {
		t.Fatalf("errors = %#v, want %#v", log.getErrors(), expectedErrors)
	}
	if log.isFailure() != (len(errors) > 0) || log.isSuccess() != (len(errors) == 0) {
		t.Fatal("incorrect failure/success classification")
	}
	if log.getNumMessages() != len(warnings)+len(errors) || log.getNumErrors() != len(errors) {
		t.Fatalf("message/error counts = %d/%d", log.getNumMessages(), log.getNumErrors())
	}
	if !reflect.DeepEqual(log.getWarningDetails(), warnings) {
		t.Fatalf("warning details = %#v, want %#v", log.getWarningDetails(), warnings)
	}
	if !reflect.DeepEqual(log.getErrorDetails(), errors) {
		t.Fatalf("error details = %#v, want %#v", log.getErrorDetails(), errors)
	}
	if checkMessages {
		want := append(append([]*sanyErrorDetails{}, warnings...), errors...)
		if !reflect.DeepEqual(log.getMessages(), want) {
			t.Fatalf("messages = %#v, want %#v", log.getMessages(), want)
		}
	}
	summary := log.String()
	for _, expected := range append(expectedWarnings, expectedErrors...) {
		if !strings.Contains(summary, expected) {
			t.Fatalf("summary missing complete detail %q: %s", expected, summary)
		}
	}
}
func TestTestErrors_testWarningMessages(t *testing.T) {
	log := &sanyErrors{}
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	message1, message2, message3 := "This is a test warning message", "This is another test warning message", "This is yet another test warning message"
	log.addMessage(originalErrorCode(4800), &loc1, message1)
	log.addMessage(originalErrorCode(4801), &loc2, message2)
	log.addMessage(originalErrorCode(4800), nil, message3)
	expected := []*sanyErrorDetails{originalExpectedDetail(4800, loc1, message1), originalExpectedDetail(4801, loc2, message2), originalExpectedDetail(4800, tlc.NullSourceLocation, message3)}
	requireOriginalErrorsAssertions(t, log, expected, []*sanyErrorDetails{}, true)
}
func TestTestErrors_testErrorMessages(t *testing.T) {
	log := &sanyErrors{}
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	message1, message2, message3 := "This is a test error message", "This is another test error message", "This is yet another test error message"
	log.addMessage(originalErrorCode(4003), &loc1, message1)
	log.addMessage(originalErrorCode(4003), &loc2, message2)
	log.addMessage(originalErrorCode(4003), nil, message3)
	expected := []*sanyErrorDetails{originalExpectedDetail(4003, loc1, message1), originalExpectedDetail(4003, loc2, message2), originalExpectedDetail(4003, tlc.NullSourceLocation, message3)}
	requireOriginalErrorsAssertions(t, log, []*sanyErrorDetails{}, expected, true)
}
func TestTestErrors_testMixedMessageLevels(t *testing.T) {
	log := &sanyErrors{}
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	warning, message := "This is a test warning message", "This is a test error message"
	log.addMessage(originalErrorCode(4800), &loc1, warning)
	log.addMessage(originalErrorCode(4003), &loc2, message)
	requireOriginalErrorsAssertions(t, log, []*sanyErrorDetails{originalExpectedDetail(4800, loc1, warning)}, []*sanyErrorDetails{originalExpectedDetail(4003, loc2, message)}, false)
}
func TestTestErrors_testDuplicateErrorsIgnored(t *testing.T) {
	log := &sanyErrors{}
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	warning, message := "This is a test warning message", "This is a test error message"
	log.addMessage(originalErrorCode(4800), &loc1, warning)
	log.addMessage(originalErrorCode(4800), &loc1, warning)
	log.addMessage(originalErrorCode(4800), &loc1, warning)
	log.addMessage(originalErrorCode(4003), &loc2, message)
	log.addMessage(originalErrorCode(4003), &loc2, message)
	log.addMessage(originalErrorCode(4003), &loc2, message)
	requireOriginalErrorsAssertions(t, log, []*sanyErrorDetails{originalExpectedDetail(4800, loc1, warning)}, []*sanyErrorDetails{originalExpectedDetail(4003, loc2, message)}, false)
}
