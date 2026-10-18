package tlc

import (
	"strings"
	"testing"
	"time"
)

func TestSpecWriterJavaIdentifierCounterStartsLikeJava(t *testing.T) {
	specWriterCounter.Store(1)
	t.Cleanup(func() { specWriterCounter.Store(1) })
	id := SpecWriterValidIdentifier(tlaSchemeInvariant)
	if !strings.HasPrefix(id, tlaSchemeInvariant+"_") {
		t.Fatalf("identifier %q does not have invariant scheme prefix", id)
	}
	if !strings.HasSuffix(id, "2000") {
		t.Fatalf("first identifier suffix = %q, want Java incrementAndGet suffix 2000", id)
	}
}

func TestSpecWriterJavaDateStringPadsSingleDigitDayLikeDateToString(t *testing.T) {
	when := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.FixedZone("CST", -6*60*60))
	if got, want := specWriterJavaDateString(when), "Fri Jan  2 03:04:05 CST 2026"; got != want {
		t.Fatalf("specWriterJavaDateString() = %q, want %q", got, want)
	}
}

func TestSpecWriterUtilityIDMatcherMirrorsJavaSchemes(t *testing.T) {
	if !SpecWriterIDMatcher.MatchString("inv_12345678901232000") {
		t.Fatalf("ID matcher rejected Java-shaped invariant id")
	}
	if SpecWriterIDMatcher.MatchString("view_12345678901232000") {
		t.Fatalf("ID matcher accepted view id; Java utility matcher excludes view scheme")
	}
	if SpecWriterIDMatcher.MatchString("inv_123") {
		t.Fatalf("ID matcher accepted too-short id")
	}
}

func TestSpecWriterFalseInitNextUtilitiesMirrorJavaText(t *testing.T) {
	specWriterCounter.Store(1)
	t.Cleanup(func() { specWriterCounter.Store(1) })

	init := SpecWriterCreateFalseInit("pc")
	if len(init) != 1 || len(init[0]) != 2 {
		t.Fatalf("false init shape = %#v", init)
	}
	if !strings.HasPrefix(init[0][0], "init_") || !strings.HasSuffix(init[0][0], "2000") {
		t.Fatalf("false init id = %q", init[0][0])
	}
	if got, want := init[0][1], init[0][0]+" ==\nFALSE/\\pc = 0"; got != want {
		t.Fatalf("false init content = %q, want %q", got, want)
	}

	next := SpecWriterCreateFalseNext("pc")
	if len(next) != 1 || len(next[0]) != 2 {
		t.Fatalf("false next shape = %#v", next)
	}
	if !strings.HasPrefix(next[0][0], "next_") || !strings.HasSuffix(next[0][0], "3000") {
		t.Fatalf("false next id = %q", next[0][0])
	}
	if got, want := next[0][1], next[0][0]+" ==\nFALSE/\\pc' = pc"; got != want {
		t.Fatalf("false next content = %q, want %q", got, want)
	}
}
