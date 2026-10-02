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

// Port of JavaMail 1.6.8 InternetAddressFoldTest.data/testFold.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/InternetAddressFoldTest.java

package tlc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestJavaMailInternetAddressFold(t *testing.T) {
	data, err := os.ReadFile("test_vectors/javamail/addrfolddata")
	if err != nil {
		t.Fatal(err)
	}
	rest := string(data)
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
	index := 0
	for len(rest) > 0 {
		line := readLine()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, "FOLD") {
			t.Fatal("TEST DATA FORMAT ERROR, MISSING FOLD")
		}
		count, err := strconv.Atoi(line[5:])
		if err != nil {
			t.Fatal(err)
		}
		addresses := make([]*MailInternetAddress, count)
		for i := range addresses {
			addresses[i], err = NewMailInternetAddress(readString())
			if err != nil {
				t.Fatal(err)
			}
		}
		if readLine() != "EXPECT" {
			t.Fatal("TEST DATA FORMAT ERROR, MISSING EXPECT")
		}
		expected := readString()
		t.Run(fmt.Sprintf("%03d", index), func(t *testing.T) {
			actual := FormatMailInternetAddresses(addresses, 0)
			if actual == nil || *actual != expected {
				t.Fatalf("Fold: got %v, want %q", actual, expected)
			}
		})
		index++
	}
}
