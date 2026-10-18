package tlc

import (
	"testing"
	"time"
)

func TestTLCRevisionDateUsesJavaMillisecondFormat(t *testing.T) {
	date := time.Date(2026, 10, 1, 12, 34, 56, 789_123_456, time.FixedZone("offset", -5*60*60))
	got := javaRevisionDate(date)
	const want = "2026-10-01T17:34:56.789Z"
	if got != want {
		t.Fatalf("javaRevisionDate = %q, want %q", got, want)
	}
}
