/*
 * Copyright (c) 1997, 2018 Oracle and/or its affiliates. All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * This Source Code may also be made available under the following Secondary
 * Licenses when the conditions for such availability set forth in the
 * Eclipse Public License v. 2.0 are satisfied: GNU General Public License,
 * version 2 with the GNU Classpath Exception, which is available at
 * https://www.gnu.org/software/classpath/license.html.
 *
 * SPDX-License-Identifier: EPL-2.0 OR GPL-2.0 WITH Classpath-exception-2.0
 */

// Port of JavaMail 1.6.8 FoldTest.data/testFold.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/FoldTest.java

package tlc

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type javaMailFoldCase struct{ direction, original, expected string }

func TestJavaMailFold(t *testing.T) {
	for i, c := range loadJavaMailFoldCases(t) {
		t.Run(fmt.Sprintf("%03d_%s", i, c.direction), func(t *testing.T) {
			var actual, expected string
			switch c.direction {
			case "BOTH":
				actual = UnfoldMailHeader(FoldMailHeader(0, c.original))
				expected = c.original
			case "FOLD":
				actual = FoldMailHeader(0, c.original)
				expected = c.expected
			case "UNFOLD":
				actual = UnfoldMailHeader(c.original)
				expected = c.expected
			default:
				t.Fatalf("Unknown direction: %s", c.direction)
			}
			if actual != expected {
				t.Fatalf("%s: got %q, want %q", c.direction, actual, expected)
			}
		})
	}
}

func loadJavaMailFoldCases(t *testing.T) []javaMailFoldCase {
	t.Helper()
	data, err := os.ReadFile("test_vectors/javamail/folddata")
	if err != nil {
		t.Fatal(err)
	}
	rest := string(data)
	// BufferedReader.readLine recognizes CR, LF and CRLF. The string reader below
	// retains all of them inside the '$'-terminated strings, as the source does.
	readLine := func() string {
		end := strings.IndexAny(rest, "\r\n")
		if end < 0 {
			line := rest
			rest = ""
			return line
		}
		line := rest[:end]
		next := end + 1
		if rest[end] == '\r' && next < len(rest) && rest[next] == '\n' {
			next++
		}
		rest = rest[next:]
		return line
	}
	readString := func() string {
		end := strings.IndexByte(rest, '$')
		if end < 0 {
			t.Fatal("missing upstream string terminator")
		}
		value := rest[:end]
		rest = rest[end+1:]
		readLine()
		return value
	}
	cases := []javaMailFoldCase{}
	for len(rest) > 0 {
		direction := readLine()
		if direction == "" || strings.HasPrefix(direction, "#") {
			continue
		}
		c := javaMailFoldCase{direction: direction, original: readString()}
		if direction != "BOTH" {
			if readLine() != "EXPECT" {
				t.Fatal("TEST DATA FORMAT ERROR")
			}
			c.expected = readString()
		}
		cases = append(cases, c)
	}
	return cases
}
