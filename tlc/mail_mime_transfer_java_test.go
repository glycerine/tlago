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

// Ports the remaining three JavaMail 1.6.8 MimeUtilityTest methods, after
// DataSource/DataHandler encoding selection and their dependencies.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/test/java/javax/mail/internet/MimeUtilityTest.java
package tlc

import (
	"os"
	"testing"
)

func TestJavaMailNonASCIIEncoding(t *testing.T) {
	data := []byte{0xfe, 0x5b, 0xdc, 0x5f, 0x92, 0x30, 0x88, 0x30, 0x8d, 0x30, 0x57, 0x30, 0x4f, 0x30, 0x4a, 0x30, 0x6d, 0x30, 0x4c, 0x30, 0x44, 0x30, 0x57, 0x30, 0x7e, 0x30, 0x59, 0x30, 0x0d, 0x00, 0x0a, 0x00}
	ds := NewMailByteArrayDataSource(data, javaString("text/plain; charset=utf-16be"))
	en, err := GetMailDataSourceEncoding(ds)
	if err != nil {
		t.Fatal(err)
	}
	if en != "base64" {
		t.Fatalf("non-ASCII encoding: got %q, want base64", en)
	}
}
func javaMailAssertEncoding(t *testing.T, source *MailDataSource) {
	t.Helper()
	valid := func(v string) bool { return v == "7bit" || v == "8bit" || v == "quoted-printable" || v == "base64" }
	en, err := GetMailDataSourceEncoding(source)
	if err != nil {
		t.Fatal(err)
	}
	if !valid(en) {
		t.Fatalf("getEncoding(DataSource): %q", en)
	}
	handler, err := NewMailDataHandler(source)
	if err != nil {
		t.Fatal(err)
	}
	en, err = GetMailDataHandlerEncoding(handler)
	if err != nil {
		t.Fatal(err)
	}
	if !valid(en) {
		t.Fatalf("getEncoding(DataHandler): %q", en)
	}
}
func TestJavaMailGetEncodingMissingFile(t *testing.T) {
	path := "javax.mail.internet.MimeUtilityTest"
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s exists", path)
	}
	ds := NewMailFilePathDataSource(path)
	name := ds.GetName()
	if name == nil || *name != NewTLAFile(path, false, nil).GetName() {
		t.Fatalf("getName: got %v", name)
	}
	javaMailAssertEncoding(t, ds)
}
func TestJavaMailGetEncodingBadContent(t *testing.T) {
	content := "bad-content-type"
	typ, err := NewMailContentType(javaString(content))
	if err == nil {
		t.Fatal(typ.String())
	}
	if _, ok := err.(*MailParseException); !ok || typ != nil {
		t.Fatal(err)
	}
	ds, err := NewMailStringDataSource(javaString(""), javaString(content))
	if err != nil {
		t.Fatal(err)
	}
	ds.SetName(nil)
	javaMailAssertEncoding(t, ds)
	ds.SetName(javaString(""))
	javaMailAssertEncoding(t, ds)
	ds.SetName(javaString("javax.mail.internet.MimeUtilityTest"))
	javaMailAssertEncoding(t, ds)
}
