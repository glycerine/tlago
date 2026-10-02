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

// Port of all JavaMail 1.6.8 MimeUtilityTest assertions, after their features.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/MimeUtilityTest.java

package tlc

import (
	"strings"
	"testing"
)

func TestJavaMailSurrogatePairs(t *testing.T) {
	sp := "a" + strings.Repeat("\U00010400", 16)
	en, err := EncodeMailText(sp, javaString("utf-8"), javaString("B"))
	if err != nil {
		t.Fatal(err)
	}
	dt, err := DecodeMailText(en)
	if err != nil {
		t.Fatal(err)
	}
	if dt != sp {
		t.Fatalf("got %q, want %q", dt, sp)
	}
	words := strings.Split(en, " ")
	dw, err := DecodeMailWord(words[0])
	if err != nil {
		t.Fatal(err)
	}
	units := mailAddressUTF16(dw)
	if units[len(units)-1] != 0xdc00 {
		t.Fatal("first word does not end with the second half of a pair")
	}
	dw, err = DecodeMailWord(words[1])
	if err != nil {
		t.Fatal(err)
	}
	units = mailAddressUTF16(dw)
	if units[0] != 0xd801 {
		t.Fatal("second word does not start with the first half of a pair")
	}
	value := ""
	for i := 0; i < 50; i++ {
		value += "\U000FE000"
		encoded, err := EncodeMailText(value, javaString("UTF-8"), javaString("B"))
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeMailText(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if decoded != value {
			t.Fatalf("got %q, want %q", decoded, value)
		}
	}
}
func TestJavaMailBadChineseCharsets(t *testing.T) {
	good := "=?gb18030?B?xbfUqqLjIChFVVIpttK7u4EwhDYgKENOWSk=?="
	expected, err := DecodeMailWord(good)
	if err != nil {
		t.Fatal(err)
	}
	for _, charset := range []string{"gb2312", "gbk", "ms936", "cp936"} {
		decoded, err := DecodeMailWord("=?" + charset + "?B?xbfUqqLjIChFVVIpttK7u4EwhDYgKENOWSk=?=")
		if err != nil {
			t.Fatal(err)
		}
		if decoded != expected {
			t.Fatalf("%s: got %q, want %q", charset, decoded, expected)
		}
	}
}
