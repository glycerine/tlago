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

// Port of the JavaMail 1.6.8 HeaderTokenizerTest JUnit assertions.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/HeaderTokenizerTest.java
package tlc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestJavaMailHeaderTokenizer(t *testing.T) {
	for i, c := range loadJavaMailHeaderCases(t, "tokenlist", false) {
		t.Run(fmt.Sprintf("%03d_%s", i, c.name), func(t *testing.T) {
			h := NewMailHeaderTokenizer(javaString(c.value))
			tokens := []*MailHeaderToken{}
			for {
				token, err := h.Next()
				if err != nil {
					if _, ok := err.(*MailParseException); !ok {
						t.Fatal(err)
					}
					if len(c.expected) != 1 || c.expected[0] != "Exception" {
						t.Fatalf("Expected exception: %v", err)
					}
					return
				}
				if token.Type == MailHeaderEOF {
					break
				}
				tokens = append(tokens, token)
			}
			if len(tokens) != len(c.expected) {
				t.Fatalf("Number of tokens: got %d, want %d", len(tokens), len(c.expected))
			}
			for i, token := range tokens {
				split := strings.IndexByte(c.expected[i], '\t')
				if split < 0 {
					t.Fatalf("invalid upstream token %q", c.expected[i])
				}
				kind, value := c.expected[i][:split], c.expected[i][split+1:]
				typ := 0
				switch kind {
				case "ATOM":
					typ = MailHeaderAtom
				case "QUOTEDSTRING":
					typ = MailHeaderQuotedString
				case "COMMENT":
					typ = MailHeaderComment
				case "EOF":
					typ = MailHeaderEOF
				default:
					typ = int(mailAddressUTF16(value)[0])
				}
				if token.Type != typ {
					t.Errorf("Token type: got %d, want %d", token.Type, typ)
				}
				if token.Value == nil || *token.Value != value {
					t.Errorf("Token value: got %v, want %q", token.Value, value)
				}
			}
		})
	}
}

type javaMailHeaderCase struct {
	name, value string
	expected    []string
}

// Source mailbox reader shared by HeaderTokenizerTest and ParameterListDecode.
// Parameter expected lines are left-trimmed and decode literal Unicode escapes;
// tokenizer expected lines use String.trim on both ends.
func loadJavaMailHeaderCases(t *testing.T, resource string, parameters bool) []javaMailHeaderCase {
	t.Helper()
	file, err := os.Open("test_vectors/javamail/" + resource)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	pos := 0
	read := func() (string, bool) {
		if pos == len(lines) {
			return "", false
		}
		line := lines[pos]
		pos++
		return line, true
	}
	cases := []javaMailHeaderCase{}
	header := ""
	for {
		line, ok := read()
		if ok && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			header += "\n" + line
			continue
		}
		match := strings.HasPrefix(header, "From: ") || strings.HasPrefix(header, "To: ") || strings.HasPrefix(header, "Cc: ")
		if parameters {
			match = len(header) >= 14 && mailAddressEqualsIgnoreCase(header[:14], "Content-Type: ")
		}
		if match {
			var expected []string
			if ok && strings.HasPrefix(line, "Expect: ") {
				count, e := strconv.Atoi(line[8:])
				if e == nil {
					expected = make([]string, count)
					for i := range expected {
						v, present := read()
						if !present {
							t.Fatal("missing upstream expected line")
						}
						if parameters {
							v = strings.TrimLeftFunc(v, func(r rune) bool { return r <= ' ' })
							chars := mailAddressUTF16(v)
							out := []uint16{}
							for i := 0; i < len(chars); i++ {
								c := chars[i]
								if c == '\\' && i+1 < len(chars) && chars[i+1] == 'u' {
									if i+6 > len(chars) {
										t.Fatal("incomplete upstream Unicode escape")
									}
									n, e := strconv.ParseUint(mailAddressUTF16String(chars[i+2:i+6]), 16, 16)
									if e != nil {
										t.Fatal(e)
									}
									c = uint16(n)
									i += 5
								}
								out = append(out, c)
							}
							v = mailAddressUTF16String(out)
						} else {
							v = mailAddressTrim(v)
						}
						expected[i] = v
					}
				} else if strings.HasPrefix(line[8:], "Exception") {
					expected = []string{"Exception"}
				}
			}
			colon := strings.IndexByte(header, ':')
			if colon >= 0 && colon+2 <= len(header) {
				cases = append(cases, javaMailHeaderCase{header[:colon], header[colon+2:], expected})
			}
		}
		if !ok {
			return cases
		}
		if line == "" {
			for {
				line, ok = read()
				if !ok {
					return cases
				}
				if strings.HasPrefix(line, "From ") {
					break
				}
			}
		}
		header = line
	}
}
