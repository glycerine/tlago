/*
 * Copyright (c) 2015, 2021, Oracle and/or its affiliates. All rights reserved.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * This code is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 2 only, as
 * published by the Free Software Foundation.  Oracle designates this
 * particular file as subject to the "Classpath" exception as provided
 * by Oracle in the LICENSE file that accompanied this code.
 *
 * This code is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * version 2 for more details (a copy is included in the LICENSE file that
 * accompanied this code).
 *
 * You should have received a copy of the GNU General Public License version
 * 2 along with this work; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin St, Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * Please contact Oracle, 500 Oracle Parkway, Redwood Shores, CA 94065 USA
 * or visit www.oracle.com if you need additional information or have any
 * questions.
 */

// OpenJDK DataFlavorUtil text-subtype cache and canonical charset-name metadata.
// Codec implementations and desktop/default-charset providers are independent.
package tlc

import (
	_ "embed"
	"strings"
	"sync"
)

// Name inventory generated from the standard OpenJDK 21.0.12.1 charset providers.
// Lookup does not assert that the corresponding Go conversion codecs are ported.
//
//go:embed resources/openjdk/CharsetNames
var mailFlavorCharsetNames string
var mailFlavorCharsetInventory = func() map[string]string {
	m := map[string]string{}
	for _, line := range strings.Split(mailFlavorCharsetNames, "\n") {
		if key, value, ok := strings.Cut(line, "\t"); ok {
			m[key] = value
		}
	}
	return m
}()

func mailFlavorCanonicalCharset(name *string) *string {
	if name == nil {
		return nil
	}
	if cb := DefaultMailActivationEnvironment.CanonicalCharset; cb != nil {
		return cb(name)
	}
	if !mailFlavorLegalCharsetName(*name) {
		return copyJavaMessage(name)
	}
	if canon, ok := mailFlavorCharsetInventory[strings.ToLower(*name)]; ok {
		return javaString(canon)
	}
	return copyJavaMessage(name)
}
func mailFlavorEncodingSupported(name *string) bool {
	if name == nil {
		return false
	}
	if cb := DefaultMailActivationEnvironment.EncodingSupported; cb != nil {
		return cb(name)
	}
	if !mailFlavorLegalCharsetName(*name) {
		return false
	}
	_, ok := mailFlavorCharsetInventory[strings.ToLower(*name)]
	return ok
}

var mailFlavorSubtypeCache = struct {
	sync.Mutex
	values map[string]bool
}{values: map[string]bool{
	"sgml": true, "xml": true, "html": true, "enriched": true, "richtext": true, "uri-list": true, "directory": true, "css": true, "calendar": true, "plain": true,
	"rtf": false, "tab-separated-values": false, "t140": false, "rfc822-headers": false, "parityfec": false,
}}

func mailFlavorSubtypeCharset(f *MailDataFlavor) bool {
	sub := f.GetSubType()
	if sub == nil {
		return false
	}
	mailFlavorSubtypeCache.Lock()
	support, ok := mailFlavorSubtypeCache.values[*sub]
	mailFlavorSubtypeCache.Unlock()
	if ok {
		return support
	}
	support = f.GetParameter(javaString("charset")) != nil
	// JDK synchronizes the get and put separately; this is not computeIfAbsent.
	mailFlavorSubtypeCache.Lock()
	mailFlavorSubtypeCache.values[*sub] = support
	mailFlavorSubtypeCache.Unlock()
	return support
}

func mailFlavorLegalCharsetName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i, c := range []byte(name) {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if i > 0 && (c == '-' || c == '+' || c == '.' || c == ':' || c == '_') {
			continue
		}
		return false
	}
	return true
}
func mailFlavorIsString(f *MailDataFlavor) bool {
	if f.GetRepresentationClass() != mailFlavorStringClass || f.mime == nil {
		return false
	}
	m, err := NewMailAWTMimeType(javaString("application/x-java-serialized-object"))
	if err != nil {
		panic(err)
	}
	return m.Match(f.mime)
}
func mailFlavorCharsetText(f *MailDataFlavor) bool {
	if mailFlavorIsString(f) {
		return true
	}
	if !mailNullableStringsEqual(f.GetPrimaryType(), javaString("text")) || !mailFlavorSubtypeCharset(f) {
		return false
	}
	class := f.GetRepresentationClass()
	if f.IsRepresentationClassReader() || class == mailFlavorStringClass || f.IsRepresentationClassCharBuffer() || class == mailFlavorCharArrayClass {
		return true
	}
	if !(f.IsRepresentationClassInputStream() || f.IsRepresentationClassByteBuffer() || class == mailFlavorByteArrayClass) {
		return false
	}
	charset := f.GetParameter(javaString("charset"))
	return charset == nil || mailFlavorEncodingSupported(charset)
}
func mailFlavorNoncharsetText(f *MailDataFlavor) bool {
	if !mailNullableStringsEqual(f.GetPrimaryType(), javaString("text")) || mailFlavorSubtypeCharset(f) {
		return false
	}
	return f.IsRepresentationClassInputStream() || f.IsRepresentationClassByteBuffer() || f.GetRepresentationClass() == mailFlavorByteArrayClass
}
func mailFlavorDefaultCharset() string {
	if cb := DefaultMailActivationEnvironment.DefaultCharset; cb != nil {
		return cb()
	}
	return "UTF-8"
}
