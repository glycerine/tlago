// Copyright (c) 2015 Microsoft Research. All rights reserved.
package tlc

import (
	"reflect"
	"testing"
)

// Whole original util.StringHelperTest, all 20 active JUnit 3 methods. The
// source's commented TODO tests remain inactive and uncredited.
func TestJavaStringHelper(t *testing.T) {
	t.Run("testGetWords", func(t *testing.T) {
		for _, text := range []string{"Abc def ghi.", "Abc  def    ghi.      "} {
			got := StringHelperGetWords(&text)
			want := []string{"Abc", "def", "ghi."}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("words=%q, want %q", got, want)
			}
		}
	})
	t.Run("testGetWordsLeadingSpace", func(t *testing.T) {
		got := StringHelperGetWords(javaString("     Abc def ghi."))
		want := []string{"Abc", "def", "ghi."}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("words=%q, want %q", got, want)
		}
	})
	for _, input := range []struct {
		name   string
		copies int32
		want   string
	}{
		{"testCopyString0", 0, ""}, {"testCopyStringPos1", 1, "abc"}, {"testCopyStringPos5", 5, "abcabcabcabcabc"},
		{"testCopyStringNeg1", -1, ""}, {"testCopyStringNeg5", -5, ""},
	} {
		t.Run(input.name, func(t *testing.T) {
			got := StringHelperCopyString(javaString("abc"), input.copies)
			if got != input.want {
				t.Fatalf("copy=%q, want %q", got, input.want)
			}
		})
	}
	t.Run("testOnlySpaces", func(t *testing.T) {
		for _, input := range []struct {
			text string
			want bool
		}{
			{"        ", true}, {" ", true}, {"  a  ", false}, {" a", false}, {"a ", false}, {"a", false},
		} {
			if got := StringHelperOnlySpaces(&input.text); got != input.want {
				t.Fatalf("onlySpaces(%q)=%v, want %v", input.text, got, input.want)
			}
		}
	})
	for _, input := range []struct {
		name  string
		fn    func(*string) string
		pairs [][2]string
	}{
		{"testTrimFront", StringHelperTrimFront, [][2]string{{"    ", ""}, {" a", "a"}, {" a ", "a "}, {" aaa", "aaa"}, {"  aa", "aa"}, {"   a ", "a "}, {"  aa  a ", "aa  a "}}},
		{"testTrimFrontWhitespaces", StringHelperTrimFront, [][2]string{{"Abc def ghi.", "Abc def ghi."}}},
		{"testTrimEnd", StringHelperTrimEnd, [][2]string{{"    ", ""}, {"a ", "a"}, {" a ", " a"}, {"aaa   ", "aaa"}, {"aa  ", "aa"}, {" a  ", " a"}, {" a aa ", " a aa"}}},
		{"testTrimEndWhitespaces", StringHelperTrimEnd, [][2]string{{"Abc def ghi.", "Abc def ghi."}}},
	} {
		t.Run(input.name, func(t *testing.T) {
			for _, pair := range input.pairs {
				if got := input.fn(&pair[0]); got != pair[1] {
					t.Fatalf("trim(%q)=%q, want %q", pair[0], got, pair[1])
				}
			}
		})
	}
	t.Run("testLeadingSpace", func(t *testing.T) {
		for _, input := range []struct {
			text string
			want int
		}{{" ", 1}, {"  ", 2}, {" a ", 1}, {"  a  a  ", 2}, {"a", 0}} {
			if got := StringHelperLeadingSpaces(&input.text); got != input.want {
				t.Fatalf("leadingSpaces(%q)=%d, want %d", input.text, got, input.want)
			}
		}
	})
	t.Run("testIsIdentifier", func(t *testing.T) {
		for _, input := range []struct {
			text string
			want bool
		}{
			{"123", false}, {" 123", false}, {"123 ", false}, {" 123 ", false}, {"1A1", true}, {"_1", true}, {"1_", true}, {"_", true},
		} {
			if got := StringHelperIsIdentifier(&input.text); got != input.want {
				t.Fatalf("isIdentifier(%q)=%v, want %v", input.text, got, input.want)
			}
		}
	})
	for _, input := range []struct {
		name string
		fn   func(*string)
	}{
		{"testGetWordsNull", func(s *string) { StringHelperGetWords(s) }},
		{"testOnlySpacesNull", func(s *string) { StringHelperOnlySpaces(s) }},
		{"testTrimFrontNull", func(s *string) { StringHelperTrimFront(s) }},
		{"testTrimEndNull", func(s *string) { StringHelperTrimEnd(s) }},
		{"testLeadingSpacesNull", func(s *string) { StringHelperLeadingSpaces(s) }},
		{"testIsIdentifierNull", func(s *string) { StringHelperIsIdentifier(s) }},
	} {
		t.Run(input.name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("null should not be accepted; got %T: %v", failure, failure)
				}
			}()
			input.fn(nil)
		})
	}
}
