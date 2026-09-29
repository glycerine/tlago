package sany_tests

import (
	"sort"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/st/LocationTest.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestLocationTest_testContains(t *testing.T) {
	t.Skip("tla2sany wip")

	if !sanyLocation(0, 0, 10, 10).Includes(sanyLocation(0, 0, 10, 10)) {
		t.Fatal("location should include itself")
	}
	if sanyLocation(1, 0, 10, 9).Includes(sanyLocation(0, 0, 10, 10)) {
		t.Fatal("inner location should not include its outer location")
	}

	if !sanyLocation(0, 0, 10, 10).Includes(sanyLocation(1, 0, 10, 9)) {
		t.Fatal("outer location should include inner location")
	}
	if sanyLocation(1, 0, 10, 9).Includes(sanyLocation(0, 0, 10, 10)) {
		t.Fatal("inner location should not include outer location")
	}

	parsedLocations := tlago.ParseSANYLocations("line 781, col 31 to line 784, col 68 of module OpenAddressing\n" +
		"line 783, col 38 to line 784, col 68 of module OpenAddressing\n" +
		"line 784, col 41 to line 784, col 68 of module OpenAddressing\n" +
		"line 786, col 30 to line 786, col 69 of module OpenAddressing\n" +
		"line 793, col 29 to line 793, col 63 of module OpenAddressing")

	if !parsedLocations[0].Includes(parsedLocations[0]) {
		t.Fatal("parsed location should include itself")
	}
	if !parsedLocations[0].Includes(parsedLocations[1]) {
		t.Fatal("outer parsed location should include first inner parsed location")
	}
	if parsedLocations[1].Includes(parsedLocations[0]) {
		t.Fatal("first inner parsed location should not include outer parsed location")
	}
	if !parsedLocations[0].Includes(parsedLocations[2]) {
		t.Fatal("outer parsed location should include second inner parsed location")
	}
	if parsedLocations[2].Includes(parsedLocations[0]) {
		t.Fatal("second inner parsed location should not include outer parsed location")
	}
	if parsedLocations[0].Includes(parsedLocations[3]) {
		t.Fatal("disjoint later location should not be included")
	}
	if parsedLocations[3].Includes(parsedLocations[0]) {
		t.Fatal("disjoint later location should not include outer parsed location")
	}
	if parsedLocations[4].Includes(parsedLocations[0]) {
		t.Fatal("separate parsed location should not include outer parsed location")
	}
}

func TestLocationTest_testContains2(t *testing.T) {
	t.Skip("tla2sany wip")

	outer, ok := tlago.ParseSANYLocation("line 109, col 1 to line 120, col 19 of module EWD998Chan")
	if !ok {
		t.Fatal("failed to parse outer location")
	}
	inner, ok := tlago.ParseSANYLocation("line 119, col 24 to line 119, col 28 of module EWD998Chan")
	if !ok {
		t.Fatal("failed to parse inner location")
	}
	if inner.Includes(outer) {
		t.Fatal("inner location should not include outer location")
	}
	if !outer.Includes(inner) {
		t.Fatal("outer location should include inner location")
	}

	outer, ok = tlago.ParseSANYLocation("line 6, col 1 to line 6, col 16 of module Debug02")
	if !ok {
		t.Fatal("failed to parse single-line outer location")
	}
	inner, ok = tlago.ParseSANYLocation("line 6, col 9 to line 6, col 9 of module Debug02")
	if !ok {
		t.Fatal("failed to parse single-line inner location")
	}
	if inner.Includes(outer) {
		t.Fatal("single-line inner location should not include outer location")
	}
	if !outer.Includes(inner) {
		t.Fatal("single-line outer location should include inner location")
	}
}

func TestLocationTest_testComparator(t *testing.T) {
	t.Skip("tla2sany wip")

	parsedLocations := tlago.ParseSANYLocations("line 15, col 9 to line 15, col 9 of module CostMetrics\n" +
		"line 15, col 9 to line 15, col 17 of module CostMetrics\n" +
		"line 8, col 11 to line 8, col 11 of module CostMetrics\n" +
		"line 8, col 13 to line 8, col 13 of module CostMetrics\n" +
		"line 8, col 9 to line 8, col 15 of module CostMetrics\n" +
		"line 14, col 15 to line 14, col 17 of module CostMetrics\n" +
		"line 15, col 15 to line 15, col 17 of module CostMetrics\n" +
		"line 16, col 34 to line 16, col 52 of module CostMetrics\n" +
		"line 8, col 9 to line 8, col 15 of module CostMetrics\n" +
		"line 16, col 42 to line 16, col 51 of module CostMetrics\n" +
		"line 16, col 42 to line 16, col 50 of module CostMetrics\n" +
		"line 16, col 46 to line 16, col 46 of module CostMetrics\n" +
		"line 16, col 46 to line 16, col 50 of module CostMetrics\n" +
		"line 16, col 46 to line 16, col 50 of module CostMetrics\n" +
		"line 23, col 6 to line 25, col 18 of module CostMetrics\n" +
		"line 18, col 9 to line 18, col 9 of module CostMetrics")
	if got, want := len(parsedLocations), 16; got != want {
		t.Fatalf("parsed locations = %d, want %d", got, want)
	}

	sort.Slice(parsedLocations, func(i, j int) bool {
		return parsedLocations[i].Compare(parsedLocations[j]) < 0
	})
	locations := make([]tlago.Position, 0, len(parsedLocations))
	for _, loc := range parsedLocations {
		if len(locations) == 0 || locations[len(locations)-1].Compare(loc) != 0 {
			locations = append(locations, loc)
		}
	}
	if got, want := len(locations), 14; got != want {
		t.Fatalf("unique locations = %d, want %d", got, want)
	}

	for i := 1; i < len(locations); i++ {
		prev, next := locations[i-1], locations[i]
		if prev.Compare(next) >= 0 {
			t.Fatalf("locations out of order at %d: %#v >= %#v", i, prev, next)
		}
	}
}

func sanyLocation(beginLine, beginColumn, endLine, endColumn int) tlago.Position {
	return tlago.Position{
		Line:      beginLine,
		Column:    beginColumn,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}
