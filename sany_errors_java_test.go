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
	return SanyErrorCodeFromStandardValue(value)
}
func originalExpectedDetail(code int, loc tlc.SourceLocation, message string) *SanyErrorDetails {
	return &SanyErrorDetails{code: originalErrorCode(code), location: &loc, format: message}
}
func originalExpectedErrorStrings(details []*SanyErrorDetails) []string {
	result := make([]string, len(details))
	for i, d := range details {
		result[i] = d.location.String() + "\n\n" + d.format
	}
	return result
}
func requireOriginalErrorsAssertions(t *testing.T, log *SanyErrors, warnings, errors []*SanyErrorDetails, checkMessages bool) {
	t.Helper()
	expectedWarnings, expectedErrors := originalExpectedErrorStrings(warnings), originalExpectedErrorStrings(errors)
	if !reflect.DeepEqual(log.GetWarnings(), expectedWarnings) {
		t.Fatalf("warnings = %#v, want %#v", log.GetWarnings(), expectedWarnings)
	}
	if !reflect.DeepEqual(log.GetErrors(), expectedErrors) {
		t.Fatalf("errors = %#v, want %#v", log.GetErrors(), expectedErrors)
	}
	if log.IsFailure() != (len(errors) > 0) || log.IsSuccess() != (len(errors) == 0) {
		t.Fatal("incorrect failure/success classification")
	}
	if log.GetNumMessages() != len(warnings)+len(errors) || log.GetNumErrors() != len(errors) {
		t.Fatalf("message/error counts = %d/%d", log.GetNumMessages(), log.GetNumErrors())
	}
	if !reflect.DeepEqual(log.GetWarningDetails(), warnings) {
		t.Fatalf("warning details = %#v, want %#v", log.GetWarningDetails(), warnings)
	}
	if !reflect.DeepEqual(log.GetErrorDetails(), errors) {
		t.Fatalf("error details = %#v, want %#v", log.GetErrorDetails(), errors)
	}
	if checkMessages {
		want := append(append([]*SanyErrorDetails{}, warnings...), errors...)
		if !reflect.DeepEqual(log.GetMessages(), want) {
			t.Fatalf("messages = %#v, want %#v", log.GetMessages(), want)
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
	log := NewSanyErrors()
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	message1, message2, message3 := "This is a test warning message", "This is another test warning message", "This is yet another test warning message"
	log.AddMessage(originalErrorCode(4800), &loc1, message1)
	log.AddMessage(originalErrorCode(4801), &loc2, message2)
	log.AddMessage(originalErrorCode(4800), nil, message3)
	expected := []*SanyErrorDetails{originalExpectedDetail(4800, loc1, message1), originalExpectedDetail(4801, loc2, message2), originalExpectedDetail(4800, tlc.NullSourceLocation, message3)}
	requireOriginalErrorsAssertions(t, log, expected, []*SanyErrorDetails{}, true)
}
func TestTestErrors_testErrorMessages(t *testing.T) {
	log := NewSanyErrors()
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	message1, message2, message3 := "This is a test error message", "This is another test error message", "This is yet another test error message"
	log.AddMessage(originalErrorCode(4003), &loc1, message1)
	log.AddMessage(originalErrorCode(4003), &loc2, message2)
	log.AddMessage(originalErrorCode(4003), nil, message3)
	expected := []*SanyErrorDetails{originalExpectedDetail(4003, loc1, message1), originalExpectedDetail(4003, loc2, message2), originalExpectedDetail(4003, tlc.NullSourceLocation, message3)}
	requireOriginalErrorsAssertions(t, log, []*SanyErrorDetails{}, expected, true)
}
func TestTestErrors_testMixedMessageLevels(t *testing.T) {
	log := NewSanyErrors()
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	warning, message := "This is a test warning message", "This is a test error message"
	log.AddMessage(originalErrorCode(4800), &loc1, warning)
	log.AddMessage(originalErrorCode(4003), &loc2, message)
	requireOriginalErrorsAssertions(t, log, []*SanyErrorDetails{originalExpectedDetail(4800, loc1, warning)}, []*SanyErrorDetails{originalExpectedDetail(4003, loc2, message)}, false)
}
func TestTestErrors_testDuplicateErrorsIgnored(t *testing.T) {
	log := NewSanyErrors()
	next := originalErrorsLocationGenerator()
	loc1, loc2 := next(), next()
	warning, message := "This is a test warning message", "This is a test error message"
	log.AddMessage(originalErrorCode(4800), &loc1, warning)
	log.AddMessage(originalErrorCode(4800), &loc1, warning)
	log.AddMessage(originalErrorCode(4800), &loc1, warning)
	log.AddMessage(originalErrorCode(4003), &loc2, message)
	log.AddMessage(originalErrorCode(4003), &loc2, message)
	log.AddMessage(originalErrorCode(4003), &loc2, message)
	requireOriginalErrorsAssertions(t, log, []*SanyErrorDetails{originalExpectedDetail(4800, loc1, warning)}, []*SanyErrorDetails{originalExpectedDetail(4003, loc2, message)}, false)
}
