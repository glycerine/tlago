package tlc

import "testing"

// Complete FormulaTest.testUnnamed and testNamed, in their original case order.
func TestFormulaLeftAndRightHandSideMatchJava(t *testing.T) {
	unnamed := NewFormula("TRUE")
	if unnamed.IsNamed() {
		t.Fatalf("TRUE should not be named")
	}
	if got := unnamed.GetRightHandSide(); got != "TRUE" {
		t.Fatalf("unnamed RHS = %q, want TRUE", got)
	}

	body := "LET clock[i \\in 1..(__trace_var_state)] ==\n" +
		"   IF i = 1\n" +
		"   THEN [ p \\in DOMAIN pc |-> 0 ]\n" +
		"   ELSE clock[i - 1]\n" +
		"IN clock[__trace_var_state]"
	unnamed = NewFormula(body)
	if unnamed.IsNamed() {
		t.Fatalf("LET body should not be named")
	}
	if got := unnamed.GetRightHandSide(); got != body {
		t.Fatalf("unnamed LET RHS = %q, want %q", got, body)
	}

	named := NewFormula("foo == TRUE")
	if got := named.GetLeftHandSide(); got != "foo" {
		t.Fatalf("named LHS = %q, want foo", got)
	}
	if got := named.GetRightHandSide(); got != "TRUE" {
		t.Fatalf("named RHS = %q, want TRUE", got)
	}

	named = NewFormula("foo == LET bar == TRUE IN bar")
	if got := named.GetLeftHandSide(); got != "foo" {
		t.Fatalf("named LET LHS = %q, want foo", got)
	}
	if got := named.GetRightHandSide(); got != "LET bar == TRUE IN bar" {
		t.Fatalf("named LET RHS = %q, want LET bar == TRUE IN bar", got)
	}

	named = NewFormula("bar == " + body)
	if got := named.GetLeftHandSide(); got != "bar" {
		t.Fatalf("named LET LHS = %q, want bar", got)
	}
	if got := named.GetRightHandSide(); got != body {
		t.Fatalf("named LET RHS = %q, want %q", got, body)
	}
}

func TestAssignmentPrettyPrintMatchesJava(t *testing.T) {
	a := NewAssignment("X", []string{}, "X")
	a.SetModelValue(true)
	if got := a.PrettyPrint(); got != "X" {
		t.Fatalf("simple model value pretty print = %q, want X", got)
	}

	b := NewAssignment("Y", []string{}, "{a1, b1}")
	b.SetModelValue(true)
	if got, want := b.PrettyPrint(), "Y"+assignmentSign+"{a1, b1}"; got != want {
		t.Fatalf("set model value pretty print = %q, want %q", got, want)
	}

	c := NewAssignment("Z", []string{}, "{s1, s2}")
	c.SetModelValue(true)
	c.SetSymmetric(true)
	if got, want := c.PrettyPrint(), "Z"+assignmentSign+"s{s1, s2}"; got != want {
		t.Fatalf("symmetric set model value pretty print = %q, want %q", got, want)
	}

	d := NewAssignment("W", []string{}, "1")
	if got, want := d.PrettyPrint(), "W"+assignmentSign+"1"; got != want {
		t.Fatalf("ordinary assignment pretty print = %q, want %q", got, want)
	}
}

func TestTypedSetParseMatchesJava(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		typ    string
		values []string
	}{
		{"plain", "a, b, c,     d,   dsfdf", "", []string{"a", "b", "c", "d", "dsfdf"}},
		{"punctuation", "1, 2, p, h!@#$%^&*()_, dsfdf", "", []string{"1", "2", "p", "h!@#$%^&*()_", "dsfdf"}},
		{"typed", "p_1,     p_2,    p_3, \n p_4, p_5", "p", []string{"1", "2", "3", "4", "5"}},
		{"mixedType", "p_1, i_2, p_3, p_4, p_5", "", []string{"p_1", "i_2", "p_3", "p_4", "p_5"}},
		{"emptyTypedValue", "p_, p_2, p_3, p_4, p_5", "", []string{"p_", "p_2", "p_3", "p_4", "p_5"}},
		{"empty", "", "", nil},
		{"commas", ", , , , ", "", nil},
		{"bracedCommas", "{, , , ,}", "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expected := NewTypedSet()
			expected.SetType(tc.typ)
			expected.SetValues(tc.values)
			if got := ParseTypedSet(tc.input); !got.Equals(expected) {
				t.Fatalf("ParseTypedSet(%q) = type %q values %v, want type %q values %v", tc.input, got.Type, got.Values, expected.Type, expected.Values)
			}
		})
	}
}

func TestAssignmentSetFormulaPanicsLikeJavaUnsupportedOperation(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatalf("Assignment.SetFormula did not panic")
		}
	}()
	NewAssignment("X", nil, "1").SetFormula("X == 2")
}
