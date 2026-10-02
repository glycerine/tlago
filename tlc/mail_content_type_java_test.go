/*
 * Copyright (c) 2014, 2018 Oracle and/or its affiliates. All rights reserved.
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

// Port of JavaMail 1.6.8 ContentTypeTest.testMatch after ContentType.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/ContentTypeTest.java
package tlc

import "testing"

func TestJavaMailContentTypeMatch(t *testing.T) {
	empty, _ := NewMailContentType()
	if empty.MatchString(javaString("text/plain")) {
		t.Fatal("empty matches text/plain")
	}
	plain, err := NewMailContentType(javaString("text/plain"))
	if err != nil {
		t.Fatal(err)
	}
	if !plain.MatchString(javaString("text/plain")) {
		t.Fatal("plain does not match text/plain")
	}
	if empty.Match(plain) {
		t.Fatal("empty matches plain")
	}
	if plain.Match(empty) {
		t.Fatal("plain matches empty")
	}
	if !plain.MatchString(javaString("text/*")) {
		t.Fatal("plain does not match text/*")
	}
	text, err := NewMailContentType(javaString("text/*"))
	if err != nil {
		t.Fatal(err)
	}
	if !text.Match(plain) {
		t.Fatal("text wildcard does not match plain")
	}
	if !plain.Match(text) {
		t.Fatal("plain does not match text wildcard")
	}
}
