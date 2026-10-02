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

// Port of JavaMail 1.6.8 InternetAddressExtraTest validation assertions.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/InternetAddressExtraTest.java

package tlc

import "testing"

// These source tests use the raw-address/personal constructor, which does not
// parse the address. For its ASCII personal name, the equivalent raw fields
// bypass parsing here too; RFC2047 personal-name encoding remains a follow-up.
func javaMailExtraAddress(address string) *MailInternetAddress {
	return &MailInternetAddress{Address: address, Personal: javaString("test"), EncodedPersonal: javaString("test")}
}

func TestJavaMailNewlineInDomainLiteral(t *testing.T) {
	err := javaMailExtraAddress("test@[\r\nfoo]").Validate()
	if _, ok := err.(*MailAddressException); !ok {
		t.Fatalf("expected AddressException, got %v", err)
	}
}

func TestJavaMailNewlineInLocal(t *testing.T) {
	err := javaMailExtraAddress("\"test\r\nfoo\"@example.com").Validate()
	if _, ok := err.(*MailAddressException); !ok {
		t.Fatalf("expected AddressException, got %v", err)
	}
}

func TestJavaMailNewlineInLocalWithWhitespace(t *testing.T) {
	if err := javaMailExtraAddress("\"test\r\n foo\"@example.com").Validate(); err != nil {
		t.Fatal(err)
	}
}
