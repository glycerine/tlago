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

// Go port of JavaMail 1.6.8 HeaderTokenizer.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/HeaderTokenizer.java

package tlc

import "strings"

const (
	MailHeaderAtom         = -1
	MailHeaderQuotedString = -2
	MailHeaderComment      = -3
	MailHeaderEOF          = -4
	MailHeaderRFC822       = "()<>@,;:\\\"\t .[]"
	MailHeaderMIME         = "()<>@,;:\\\"\t []/?="
)

type MailHeaderToken struct {
	Type  int
	Value *string
}

var mailHeaderEOFToken = &MailHeaderToken{Type: MailHeaderEOF}

type MailHeaderTokenizer struct {
	value               []uint16
	delimiters          string
	skipComments        bool
	current, next, peek int
}

func NewMailHeaderTokenizer(header *string, delimiters ...string) *MailHeaderTokenizer {
	delimiter := MailHeaderRFC822
	if len(delimiters) > 0 {
		delimiter = delimiters[0]
	}
	return NewMailHeaderTokenizerWithComments(header, delimiter, true)
}
func NewMailHeaderTokenizerWithComments(header *string, delimiters string, skipComments bool) *MailHeaderTokenizer {
	value := ""
	if header != nil {
		value = *header
	}
	return &MailHeaderTokenizer{value: mailAddressUTF16(value), delimiters: delimiters, skipComments: skipComments}
}
func (h *MailHeaderTokenizer) Next(endOfAtom ...uint16) (*MailHeaderToken, error) {
	end := uint16(0)
	if len(endOfAtom) > 0 {
		end = endOfAtom[0]
	}
	return h.NextWithEscapes(end, false)
}
func (h *MailHeaderTokenizer) NextWithEscapes(endOfAtom uint16, keepEscapes bool) (token *MailHeaderToken, err error) {
	defer func() {
		if x := recover(); x != nil {
			token = nil
			err = panicValueAsError(x)
		}
	}()
	h.current = h.next
	token, err = h.getNext(endOfAtom, keepEscapes)
	if err == nil {
		h.next = h.current
		h.peek = h.current
	}
	return
}
func (h *MailHeaderTokenizer) Peek() (token *MailHeaderToken, err error) {
	defer func() {
		if x := recover(); x != nil {
			token = nil
			err = panicValueAsError(x)
		}
	}()
	h.current = h.peek
	token, err = h.getNext(0, false)
	if err == nil {
		h.peek = h.current
	}
	return
}
func (h *MailHeaderTokenizer) GetRemainder() *string {
	if h.next >= len(h.value) {
		return nil
	}
	return javaString(mailAddressUTF16String(h.value[h.next:]))
}
func (h *MailHeaderTokenizer) getNext(end uint16, keep bool) (*MailHeaderToken, error) {
	if h.current >= len(h.value) || h.skipWhiteSpace() == MailHeaderEOF {
		return mailHeaderEOFToken, nil
	}
	c := h.value[h.current]
	filter := false
	for c == '(' {
		h.current++
		start := h.current
		nesting := 1
		for nesting > 0 && h.current < len(h.value) {
			c = h.value[h.current]
			switch c {
			case '\\':
				h.current++
				filter = true
			case '\r':
				filter = true
			case '(':
				nesting++
			case ')':
				nesting--
			}
			h.current++
		}
		if nesting != 0 {
			return nil, NewMailParseException("Unbalanced comments")
		}
		if !h.skipComments {
			s := h.substring(start, h.current-1)
			if filter {
				s = h.filterToken(start, h.current-1, keep)
			}
			return &MailHeaderToken{MailHeaderComment, javaString(s)}, nil
		}
		if h.skipWhiteSpace() == MailHeaderEOF {
			return mailHeaderEOFToken, nil
		}
		c = h.value[h.current]
	}
	if c == '"' {
		h.current++
		return h.collectString('"', keep)
	}
	if c < 0x20 || c >= 0x7f || strings.ContainsRune(h.delimiters, rune(c)) {
		if end > 0 && c != end {
			return h.collectString(end, keep)
		}
		h.current++
		return &MailHeaderToken{int(c), javaString(mailAddressUTF16String([]uint16{c}))}, nil
	}
	start := h.current
	for h.current < len(h.value) {
		c = h.value[h.current]
		if c < 0x20 || c >= 0x7f || c == '(' || c == ' ' || c == '"' || strings.ContainsRune(h.delimiters, rune(c)) {
			if end > 0 && c != end {
				h.current = start
				return h.collectString(end, keep)
			}
			break
		}
		h.current++
	}
	return &MailHeaderToken{MailHeaderAtom, javaString(h.substring(start, h.current))}, nil
}
func (h *MailHeaderTokenizer) collectString(end uint16, keep bool) (*MailHeaderToken, error) {
	start := h.current
	filter := false
	for h.current < len(h.value) {
		c := h.value[h.current]
		if c == '\\' {
			h.current++
			filter = true
		} else if c == '\r' {
			filter = true
		} else if c == end {
			h.current++
			s := h.substring(start, h.current-1)
			if filter {
				s = h.filterToken(start, h.current-1, keep)
			}
			if c != '"' {
				s = mailHeaderTrimWhiteSpace(s)
				h.current--
			}
			return &MailHeaderToken{MailHeaderQuotedString, javaString(s)}, nil
		}
		h.current++
	}
	if end == '"' {
		return nil, NewMailParseException("Unbalanced quoted string")
	}
	var s string
	if filter {
		s = h.filterToken(start, h.current, keep)
	} else {
		s = h.substring(start, h.current)
	}
	return &MailHeaderToken{MailHeaderQuotedString, javaString(mailHeaderTrimWhiteSpace(s))}, nil
}
func (h *MailHeaderTokenizer) skipWhiteSpace() int {
	for h.current < len(h.value) {
		c := h.value[h.current]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			return h.current
		}
		h.current++
	}
	return MailHeaderEOF
}
func mailHeaderTrimWhiteSpace(value string) string {
	chars := mailAddressUTF16(value)
	i := len(chars) - 1
	for i >= 0 && (chars[i] == ' ' || chars[i] == '\t' || chars[i] == '\r' || chars[i] == '\n') {
		i--
	}
	// Preserve the source's single-character loss for a lenient token.
	if i <= 0 {
		return ""
	}
	return mailAddressUTF16String(chars[:i+1])
}
func (h *MailHeaderTokenizer) substring(start, end int) string {
	if start < 0 || end > len(h.value) || start > end {
		panic(NewStringIndexOutOfBoundsException(end, len(h.value)))
	}
	return mailAddressUTF16String(h.value[start:end])
}
func (h *MailHeaderTokenizer) filterToken(start, end int, keep bool) string {
	out := []uint16{}
	escaped, cr := false, false
	for i := start; i < end; i++ {
		if i < 0 || i >= len(h.value) {
			panic(NewStringIndexOutOfBoundsException(i, len(h.value)))
		}
		c := h.value[i]
		if c == '\n' && cr {
			cr = false
			continue
		}
		cr = false
		if !escaped {
			if c == '\\' {
				escaped = true
			} else if c == '\r' {
				cr = true
			} else {
				out = append(out, c)
			}
		} else {
			if keep {
				out = append(out, '\\')
			}
			out = append(out, c)
			escaped = false
		}
	}
	return mailAddressUTF16String(out)
}
