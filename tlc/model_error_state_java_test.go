package tlc

import (
	"fmt"
	"strings"
	"testing"
)

// Original tlc2.model.Utils helpers. The Go MCState API stores location text.
// Every nonempty source fixture is already in canonical Location.toString form.
func javaMCBuildState(ordinal int, name, location string, assignments ...string) *MCState {
	variables := make([]*MCVariable, len(assignments))
	for i, assignment := range assignments {
		split := strings.Split(assignment, "=")
		variables[i] = NewMCVariable(javaMCTrim(split[0]), javaMCTrim(split[1]))
	}
	label := "<" + javaMCTrim(fmt.Sprintf("%s %s", name, location)) + ">"
	// Location.parseLocation("") returns Location.nullLoc in the Java helper.
	parsedLocation := location
	if parsedLocation == "" {
		parsedLocation = "Unknown location"
	}
	return NewMCState(variables, name, label, parsedLocation, false, false, ordinal)
}

func javaMCTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
}

func javaMCOutputFormat(state *MCState) []string {
	lines := []string{fmt.Sprintf("%d: %s", state.GetStateNumber(), state.GetLabel())}
	for _, variable := range state.GetVariables() {
		lines = append(lines, fmt.Sprintf("/\\ %s = %s", variable.Name, variable.ValueAsString))
	}
	return lines
}

func javaMCEquals[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Fatalf("assertEquals: expected %v, got %v", expected, actual)
	}
}

func javaMCCompareStates(t *testing.T, expected, actual *MCState) {
	t.Helper()
	javaMCEquals(t, expected.GetName(), actual.GetName())
	// Source keeps the untrimmed label for backward compatibility.
	javaMCEquals(t, " "+expected.GetLabel(), actual.GetLabel())
	javaMCEquals(t, expected.IsStuttering(), actual.IsStuttering())
	javaMCEquals(t, expected.IsBackToState(), actual.IsBackToState())
	javaMCEquals(t, expected.GetStateNumber(), actual.GetStateNumber())
	expectedVars, actualVars := expected.GetVariables(), actual.GetVariables()
	javaMCEquals(t, len(expectedVars), len(actualVars))
	for i := range expectedVars {
		javaMCEquals(t, expectedVars[i].Name, actualVars[i].Name)
		javaMCEquals(t, expectedVars[i].ValueAsString, actualVars[i].ValueAsString)
	}
}

func javaMCParseRoundTrip(t *testing.T, ordinal int, name, location string, assignments ...string) {
	t.Helper()
	expected := javaMCBuildState(ordinal, name, location, assignments...)
	inputLines := javaMCOutputFormat(expected)
	input := strings.Join(inputLines, tlaCR)
	actual := ParseMCState(input)
	javaMCCompareStates(t, expected, actual)
}

// Entire original tlc2.model.MCStateTest, including all six round trips.
func TestJavaMCState(t *testing.T) {
	t.Run("testParseRoundTrips", func(t *testing.T) {
		javaMCParseRoundTrip(t, 1, "Initial predicate", "", "x = 8", "y = 7")
		javaMCParseRoundTrip(t, 2, "YIncr", "line 8, col 10 to line 10, col 26 of module Bla", "x = 8", "y = 15")
		javaMCParseRoundTrip(t, 1, "Initial predicate", "", "x = 1", "y = FALSE")
		javaMCParseRoundTrip(t, 2, "Next", "line 7, col 9 to line 11, col 23 of module Alias", "x = 2", "y = TRUE")
		javaMCParseRoundTrip(t, 3, "Next", "line 7, col 9 to line 11, col 23 of module Alias", "x = 3", "y = FALSE")
		javaMCParseRoundTrip(t, 4, "Next", "line 7, col 9 to line 11, col 23 of module Alias", "x = 4", "y = TRUE")
	})
	t.Run("testSimpleRecordPrinter", func(t *testing.T) {
		input := javaMCBuildState(1, "Initial predicate", "", "x = 8", "y = 7")
		actual := input.AsSimpleRecord()
		expectedTokens := []string{tlaLeftSquareBracket, "x", javaMCTrim(tlaRecordArrow), "8", tlaComma, "y", javaMCTrim(tlaRecordArrow), "7", tlaRightSquareBracket}
		remaining := javaMCTrim(actual)
		for _, expectedToken := range expectedTokens {
			if strings.HasPrefix(remaining, expectedToken) {
				remaining = javaMCTrim(remaining[len(expectedToken):])
			} else {
				t.Fatalf("Required token [%s]; received [%s]", expectedToken, remaining)
			}
		}
	})
}

// Entire original tlc2.model.MCErrorTest, retaining the nested variable loop.
func TestJavaMCError(t *testing.T) {
	t.Run("testGetErrorMessage", func(t *testing.T) {
		expected := "this is an error message"
		actual := NewMCError(expected)
		javaMCEquals(t, expected, actual.GetMessage())
	})
	t.Run("testUpdateStatesForTraceExpressions", func(t *testing.T) {
		error := NewMCError()
		states := []*MCState{
			javaMCBuildState(1, "init", "", "x = 1"),
			javaMCBuildState(2, "next", "", "x = 2"),
			javaMCBuildState(3, "next", "", "x = 3"),
			javaMCBuildState(5, "next", "", "x = 4"),
			javaMCBuildState(6, "next", "", "x = 5"),
		}
		for _, state := range states {
			error.AddState(state)
		}
		variableExpressionMap := make(map[string]string)
		variableExpressionMap["x"] = "y"
		error.UpdateStatesForTraceExpression(variableExpressionMap)
		actualStates := error.GetStates()
		javaMCEquals(t, len(states), len(actualStates))
		for _, actualState := range actualStates {
			actualVariables := actualState.GetVariables()
			javaMCEquals(t, 1, len(actualVariables))
			for _, actualVariable := range actualVariables {
				javaMCEquals(t, "x", actualVariable.Name)
				javaMCEquals(t, "y", actualVariable.SingleLineDisplayName())
			}
		}
	})
}
