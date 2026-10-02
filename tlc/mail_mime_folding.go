/*
 * Copyright (c) 1997, 2019 Oracle and/or its affiliates. All rights reserved.
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

// Go port of JavaMail 1.6.8 MimeUtility.fold/unfold and its class property snapshot.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/MimeUtility.java

package tlc

import (
	"strings"
	"sync"
)

var mailMIMEClass struct {
	sync.Once
	decodeStrict, encodeEOLStrict, ignoreUnknownEncoding, allowUTF8, foldEncodedWords, foldText bool
}

func initializeMailMIMEProperties() {
	mailMIMEClass.Do(func() {
		mailMIMEClass.decodeStrict = mailBooleanProperty("mail.mime.decodetext.strict", true)
		mailMIMEClass.encodeEOLStrict = mailBooleanProperty("mail.mime.encodeeol.strict", false)
		mailMIMEClass.ignoreUnknownEncoding = mailBooleanProperty("mail.mime.ignoreunknownencoding", false)
		mailMIMEClass.allowUTF8 = mailBooleanProperty("mail.mime.allowutf8", false)
		mailMIMEClass.foldEncodedWords = mailBooleanProperty("mail.mime.foldencodedwords", false)
		mailMIMEClass.foldText = mailBooleanProperty("mail.mime.foldtext", true)
	})
}

// FoldMailHeader folds text at whitespace runs using Java's UTF-16 column count.
// As in MimeUtility.fold, the first line already occupies used columns.
func FoldMailHeader(used int, value string) string {
	initializeMailMIMEProperties()
	if !mailMIMEClass.foldText {
		return value
	}
	s := mailAddressUTF16(value)
	end := len(s) - 1
	for end >= 0 && (s[end] == ' ' || s[end] == '\t' || s[end] == '\r' || s[end] == '\n') {
		end--
	}
	s = s[:end+1]
	columns := int32(used)
	if columns+int32(len(s)) <= 76 {
		return makeMailHeaderSafe(mailAddressUTF16String(s))
	}
	out := make([]uint16, 0, len(s)+4)
	var lastc uint16
	for columns+int32(len(s)) > 76 {
		lastspace := -1
		for i, c := range s {
			if lastspace != -1 && columns+int32(i) > 76 {
				break
			}
			if (c == ' ' || c == '\t') && lastc != ' ' && lastc != '\t' {
				lastspace = i
			}
			lastc = c
		}
		if lastspace == -1 {
			out = append(out, s...)
			s = nil
			columns = 0
			break
		}
		out = append(out, s[:lastspace]...)
		out = append(out, '\r', '\n')
		lastc = s[lastspace]
		out = append(out, lastc)
		s = s[lastspace+1:]
		columns = 1
	}
	out = append(out, s...)
	return makeMailHeaderSafe(mailAddressUTF16String(out))
}

// MimeUtility.makesafe uses BufferedReader lines: CR, LF and CRLF only. It
// drops lines blank under String.trim and makes subsequent lines continuations.
func makeMailHeaderSafe(value string) string {
	if !strings.ContainsAny(value, "\r\n") {
		return value
	}
	var b strings.Builder
	for len(value) > 0 {
		end := strings.IndexAny(value, "\r\n")
		var line string
		if end < 0 {
			line, value = value, ""
		} else {
			line = value[:end]
			next := end + 1
			if value[end] == '\r' && next < len(value) && value[next] == '\n' {
				next++
			}
			value = value[next:]
		}
		if mailAddressTrim(line) == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\r\n")
			if line[0] != ' ' && line[0] != '\t' {
				b.WriteByte(' ')
			}
		}
		b.WriteString(line)
	}
	return b.String()
}
func UnfoldMailHeader(value string) string {
	initializeMailMIMEProperties()
	if !mailMIMEClass.foldText {
		return value
	}
	var b strings.Builder
	for {
		start := strings.IndexAny(value, "\r\n")
		if start < 0 {
			break
		}
		i := start + 1
		if i < len(value) && value[start] == '\r' && value[i] == '\n' {
			i++
		}
		if start > 0 && value[start-1] == '\\' {
			b.WriteString(value[:start-1])
			b.WriteString(value[start:i])
		} else if i >= len(value) || value[i] == ' ' || value[i] == '\t' {
			b.WriteString(value[:start])
		} else {
			b.WriteString(value[:i])
		}
		value = value[i:]
	}
	b.WriteString(value)
	return b.String()
}
