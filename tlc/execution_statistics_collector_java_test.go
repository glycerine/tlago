// Copyright (c) 2019 Microsoft Research. All rights reserved.
package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

type javaTestExecutionStatisticsCollector struct {
	*ExecutionStatisticsCollector
	parameters *InsMap[string, string]
	hostname   string
	submitted  bool
}

func newJavaTestExecutionStatisticsCollector(path, hostname string) *javaTestExecutionStatisticsCollector {
	c := &javaTestExecutionStatisticsCollector{ExecutionStatisticsCollector: NewExecutionStatisticsCollector(path, hostname)}
	c.SubmitOverride = func(hostname string, parameters *InsMap[string, string]) {
		c.submitted = true
		c.hostname = hostname
		c.parameters = parameters
	}
	return c
}
func javaExecutionStatisticsParameters(foo string) *InsMap[string, string] {
	parameters := NewInsMap[string, string]()
	if foo != "" {
		parameters.Set("foo", foo)
	}
	return parameters
}

// Whole original ExecutionStatisticsCollectorTest: all twelve methods and the
// original capture-only submit subclass. Each file/call/assertion is retained.
func TestJavaExecutionStatisticsCollector(t *testing.T) {
	const company = "localhost"
	const noCompany = "tlaplus-execution-stats-collection01.invalid."
	address := executionStatisticsOptIn(company)
	if address == nil {
		t.Fatal("source COMPANY_CANONICAL static initializer: localhost lookup failed")
	}
	companyCanonical := executionStatisticsCanonicalHost(address)
	for _, companyLevel := range []bool{true, false} {
		for _, kind := range []string{"NoFile", "Unreadable", "EmptyFile", "NoESCFile", "RandomIdFile", "UserDefinedIdFile"} {
			name := "test" + kind
			if !companyLevel {
				if kind == "Unreadable" {
					name = "testUnreadableFile"
				}
			} else {
				name = "testCompanyLevel" + kind
			}
			t.Run(name, func(t *testing.T) {
				hostname := noCompany
				expectedHost := "esc01.tlapl.us"
				if companyLevel {
					hostname = company
					expectedHost = companyCanonical
				}
				path := "/path/does/not/exist"
				if kind != "NoFile" {
					path = filepath.Join(t.TempDir(), "esc.txt")
					if err := os.WriteFile(path, nil, 0600); err != nil {
						t.Fatal(err)
					}
					if kind == "Unreadable" {
						info, err := os.Stat(path)
						if err != nil {
							t.Fatal(err)
						}
						if err := os.Chmod(path, info.Mode().Perm()&^0400); err != nil {
							t.Skip("source Assume: File.setReadable(false) succeeds")
						}
					}
					content := ""
					if kind == "NoESCFile" {
						content = "   " + ExecutionStatisticsNoStatistics + "   "
					}
					if kind == "RandomIdFile" {
						content = "   " + ExecutionStatisticsRandomIdentifier + "   "
					}
					if kind == "UserDefinedIdFile" {
						content = "   123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ   "
					}
					if content != "" {
						bytes, err := javaCharsetEncode(content, javaDefaultCharset(), false)
						if err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(path, bytes, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				c := newJavaTestExecutionStatisticsCollector(path, hostname)
				const userID = "123456789ABCDEFGHIJKLMNOPQRSTUVW"
				var identifierA, identifierB string
				if kind == "RandomIdFile" {
					id := c.GetIdentifier()
					if id == nil {
						t.Fatal("identifierA unexpectedly null")
					}
					identifierA = *id
					id = c.GetIdentifier()
					if id == nil {
						t.Fatal("identifierB unexpectedly null")
					}
					identifierB = *id
					if identifierA == identifierB {
						t.Fatal("random identifiers unexpectedly equal")
					}
				} else if kind == "UserDefinedIdFile" {
					id := c.GetIdentifier()
					if id == nil || *id != userID {
						t.Fatalf("identifier=%v, want %s", id, userID)
					}
				} else {
					if id := c.GetIdentifier(); id != nil {
						t.Fatalf("expected null identifier, got %s", *id)
					}
				}
				submitExpected := companyLevel && kind != "NoESCFile" || kind == "RandomIdFile" || kind == "UserDefinedIdFile"
				foo := ""
				if submitExpected {
					foo = "bar"
				}
				c.Collect0(javaExecutionStatisticsParameters(foo))
				if !submitExpected {
					if c.submitted {
						t.Fatal("submission unexpectedly occurred")
					}
					return
				}
				checkSubmission := func(wantFoo string) {
					t.Helper()
					if !c.submitted {
						t.Fatal("submission did not occur")
					}
					got := c.parameters.Get("foo")
					if got != wantFoo {
						t.Fatalf("foo=%q, want %q", got, wantFoo)
					}
					if c.hostname != expectedHost {
						t.Fatalf("hostname=%q, want %q", c.hostname, expectedHost)
					}
				}
				checkSubmission("bar")
				firstID := c.parameters.Get("id")
				if kind == "UserDefinedIdFile" {
					if firstID != userID {
						t.Fatalf("submitted id=%q, want %q", firstID, userID)
					}
				}
				if kind == "RandomIdFile" {
					if firstID == identifierA || firstID == identifierB {
						t.Fatal("submission reused an earlier random identifier")
					}
				}
				if companyLevel && (kind == "Unreadable" || kind == "EmptyFile" || kind == "RandomIdFile") {
					secondFoo := "blub"
					if kind == "RandomIdFile" {
						secondFoo = "bar"
					}
					c.Collect0(javaExecutionStatisticsParameters(secondFoo))
					checkSubmission(secondFoo)
					secondID := c.parameters.Get("id")
					if kind == "RandomIdFile" {
						if secondID == identifierA || secondID == identifierB {
							t.Fatal("second submission reused an earlier random identifier")
						}
					}
					if firstID == secondID {
						t.Fatal("submission identifiers unexpectedly equal")
					}
					if len(firstID) <= 20 || len(secondID) <= 20 {
						t.Fatalf("identifier lengths=%d,%d, want both >20", len(firstID), len(secondID))
					}
					if kind == "RandomIdFile" {
						if firstID[20:] == secondID[20:] {
							t.Fatal("random identifier suffixes unexpectedly equal")
						}
					} else if firstID[20:] != secondID[20:] {
						t.Fatalf("host identifier suffixes differ: %q, %q", firstID[20:], secondID[20:])
					}
				}
			})
		}
	}
}
