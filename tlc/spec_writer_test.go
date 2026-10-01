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
