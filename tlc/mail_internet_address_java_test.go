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

// Port of JavaMail 1.6.8 InternetAddressTest.data/testAddress JUnit assertions.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/InternetAddressTest.java

package tlc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

type javaMailAddressCase struct {
	name, value    string
	expected       []string
	strict, header bool
}

// The original JUnit test asserts parsed address count/values and expected
// AddressException. Its logging of personal names makes no assertions, and its
// standalone toString/reparse checks explicitly return early in JUnit mode.
func TestJavaMailInternetAddress(t *testing.T) {
	for i, c := range loadJavaMailAddressCases(t) {
		t.Run(fmt.Sprintf("%03d_%s", i, c.name), func(t *testing.T) {
			var addresses []*MailInternetAddress
			var err error
			if c.header {
				addresses, err = ParseMailInternetAddressHeader(c.value, c.strict)
			} else {
				addresses, err = ParseMailInternetAddresses(c.value, c.strict)
			}
			if err != nil {
				if _, ok := err.(*MailAddressException); !ok {
					t.Fatalf("For %s got non-address failure: %v", c.value, err)
				}
				if c.expected != nil && (len(c.expected) != 1 || c.expected[0] != "Exception") {
					t.Fatalf("For %s expected %d addresses, got Exception: %v", c.value, len(c.expected), err)
				}
				return
			}
			if c.expected != nil && len(addresses) != len(c.expected) {
				t.Fatalf("For %s number of addresses: got %d, want %d", c.value, len(addresses), len(c.expected))
			}
			for j, address := range addresses {
				if c.expected != nil && j < len(c.expected) && address.Address != c.expected[j] {
					t.Errorf("For %s address[%d]: got %q, want %q", c.value, j, address.Address, c.expected[j])
				}
			}
		})
	}
}

// Keep the source mail-format reader and continuation rules, including the
// blank-line rule that skips input until a mailbox "From " separator.
func loadJavaMailAddressCases(t *testing.T) []javaMailAddressCase {
	t.Helper()
	file, err := os.Open("test_vectors/javamail/addrlist")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewScanner(file)
	lines := []string{}
	for reader.Scan() {
		lines = append(lines, reader.Text())
	}
	if err = reader.Err(); err != nil {
		t.Fatal(err)
	}
	lineIndex := 0
	read := func() (string, bool) {
		if lineIndex == len(lines) {
			return "", false
		}
		line := lines[lineIndex]
		lineIndex++
		return line, true
	}
	readExpected := func() string {
		line, ok := read()
		if !ok {
			t.Fatal("missing upstream expected line")
		}
		if !strings.HasSuffix(line, "\\") {
			return line
		}
		if strings.HasSuffix(line, "\\\\") {
			return line[:len(line)-1]
		}
		var b strings.Builder
		b.WriteString(line[:len(line)-1])
		b.WriteByte('\n')
		for {
			line, ok = read()
			if !ok {
				t.Fatal("missing upstream continuation")
			}
			if !strings.HasSuffix(line, "\\") {
				b.WriteString(line)
				break
			}
			if strings.HasSuffix(line, "\\\\") {
				b.WriteString(line[:len(line)-1])
				break
			}
			b.WriteString(line[:len(line)-1])
			b.WriteByte('\n')
		}
		return b.String()
	}
	header := ""
	strict, parseHeader := false, false
	cases := []javaMailAddressCase{}
	for {
		next, ok := read()
		if ok && len(next) > 0 && (next[0] == ' ' || next[0] == '\t') {
			header += "\n" + next
			continue
		}
		switch {
		case strings.HasPrefix(header, "Strict: "):
			strict = strings.EqualFold(mailAddressTrim(header[strings.IndexByte(header, ':')+1:]), "true")
		case strings.HasPrefix(header, "Header: "):
			parseHeader = strings.EqualFold(mailAddressTrim(header[strings.IndexByte(header, ':')+1:]), "true")
		case strings.HasPrefix(header, "From: ") || strings.HasPrefix(header, "To: ") || strings.HasPrefix(header, "Cc: "):
			var expected []string
			if ok && strings.HasPrefix(next, "Expect: ") {
				count, e := strconv.Atoi(next[8:])
				if e == nil {
					expected = make([]string, count)
					for i := range expected {
						expected[i] = mailAddressTrim(readExpected())
					}
				} else if strings.HasPrefix(next[8:], "Exception") {
					expected = []string{"Exception"}
				}
			}
			colon := strings.IndexByte(header, ':')
			cases = append(cases, javaMailAddressCase{header[:colon], header[colon+2:], expected, strict, parseHeader})
		}
		if !ok {
			return cases
		}
		if next == "" {
			for {
				next, ok = read()
				if !ok {
					return cases
				}
				if strings.HasPrefix(next, "From ") {
					break
				}
			}
		}
		header = next
	}
}
