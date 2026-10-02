/*
 * Copyright (c) 1997, 2017, Oracle and/or its affiliates. All rights reserved.
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

// OpenJDK AWT MIME parameter grammar, kept separate from JavaMail/Activation.
package tlc

import (
	"fmt"
	"strings"
)

type MailAWTMimeTypeParseException struct{ javaExceptionBase }

func NewMailAWTMimeTypeParseException(message ...string) *MailAWTMimeTypeParseException {
	return &MailAWTMimeTypeParseException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *MailAWTMimeTypeParseException) Error() string {
	if e == nil {
		panic(NewNullPointerException())
	}
	return javaThrowableMessage(e)
}

type MailAWTMimeParameters struct{ parameters *mailAWTHashtable }

func NewMailAWTMimeParameters(value ...*string) (*MailAWTMimeParameters, error) {
	return mailInvoke(func() (*MailAWTMimeParameters, error) {
		p := &MailAWTMimeParameters{parameters: newMailAWTHashtable()}
		if len(value) > 0 {
			if err := p.Parse(value[0]); err != nil {
				return nil, err
			}
		}
		return p, nil
	})
}
func mailAWTToken(c uint16) bool {
	return c > 32 && c < 127 && !strings.ContainsRune("()<>@,;:\\\"/[]?=", rune(c))
}
func mailAWTChar(s []uint16, index int) uint16 {
	if index < 0 || index >= len(s) {
		panic(NewStringIndexOutOfBoundsException(index, len(s)))
	}
	return s[index]
}
func mailAWTSkipWhitespace(s []uint16, index int) int {
	if index < len(s) {
		c := mailAWTChar(s, index)
		for index < len(s) && mailActivationWhitespace(c) {
			index++
			c = mailAWTChar(s, index)
		}
	}
	return index
}
func (p *MailAWTMimeParameters) Parse(value *string) error {
	if p == nil {
		panic(NewNullPointerException())
	}

	_, err := mailInvoke(func() (bool, error) {
		if value == nil {
			return false, NewNullPointerException()
		}
		s := mailAddressUTF16(*value)
		length := len(s)
		if length == 0 {
			return false, nil
		}
		index := mailAWTSkipWhitespace(s, 0)
		if index >= length {
			return false, nil
		}
		current := s[index]
		for index < length && current == ';' {
			index++
			index = mailAWTSkipWhitespace(s, index)
			if index >= length {
				return false, NewMailAWTMimeTypeParseException("Couldn't find parameter name")
			}
			start := index
			current = s[index]
			for index < length && mailAWTToken(current) {
				index++
				current = mailAWTChar(s, index)
			}
			name := mailActivationLower(mailAddressUTF16String(s[start:index]))
			index = mailAWTSkipWhitespace(s, index)
			if index >= length || s[index] != '=' {
				return false, NewMailAWTMimeTypeParseException("Couldn't find the '=' that separates a parameter name from its value.")
			}
			index++
			index = mailAWTSkipWhitespace(s, index)
			if index >= length {
				return false, NewMailAWTMimeTypeParseException("Couldn't find a value for parameter named " + name)
			}
			current = s[index]
			var text string
			if current == '"' {
				index++
				start = index
				if index >= length {
					return false, NewMailAWTMimeTypeParseException("Encountered unterminated quoted parameter value.")
				}
				found := false
				for index < length && !found {
					current = s[index]
					if current == '\\' {
						index += 2
					} else if current == '"' {
						found = true
					} else {
						index++
					}
				}
				if current != '"' {
					return false, NewMailAWTMimeTypeParseException("Encountered unterminated quoted parameter value.")
				}
				text = mailAWTUnquote(s[start:index])
				index++
			} else if mailAWTToken(current) {
				start = index
				found := false
				for index < length && !found {
					current = s[index]
					if mailAWTToken(current) {
						index++
					} else {
						found = true
					}
				}
				text = mailAddressUTF16String(s[start:index])
			} else {
				return false, NewMailAWTMimeTypeParseException(fmt.Sprintf("Unexpected character encountered at index %d", index))
			}
			p.parameters.set(name, javaString(text))
			index = mailAWTSkipWhitespace(s, index)
			if index < length {
				current = s[index]
			}
		}
		if index < length {
			return false, NewMailAWTMimeTypeParseException("More characters encountered in input than expected.")
		}
		return false, nil
	})
	return err
}
func mailAWTUnquote(s []uint16) string {
	chars := []uint16{}
	escaped := false
	for _, c := range s {
		if !escaped && c != '\\' {
			chars = append(chars, c)
		} else if escaped {
			chars = append(chars, c)
			escaped = false
		} else {
			escaped = true
		}
	}
	return mailAddressUTF16String(chars)
}
func mailAWTQuote(value *string) string {
	if value == nil {
		panic(NewNullPointerException())
	}
	units := mailAddressUTF16(*value)
	quote := false
	for _, c := range units {
		if !mailAWTToken(c) {
			quote = true
			break
		}
	}
	if !quote {
		return *value
	}
	chars := []uint16{'"'}
	for _, c := range units {
		if c == '\\' || c == '"' {
			chars = append(chars, '\\')
		}
		chars = append(chars, c)
	}
	return mailAddressUTF16String(append(chars, '"'))
}
func mailAWTParameterName(name *string) string {
	if name == nil {
		panic(NewNullPointerException())
	}
	return mailActivationLower(mailAddressTrim(*name))
}
func (p *MailAWTMimeParameters) Size() int {
	if p == nil {
		panic(NewNullPointerException())
	}
	return p.parameters.length()
}
func (p *MailAWTMimeParameters) IsEmpty() bool {
	if p == nil {
		panic(NewNullPointerException())
	}
	return p.Size() == 0
}
func (p *MailAWTMimeParameters) Get(name *string) *string {
	if p == nil {
		panic(NewNullPointerException())
	}

	return p.parameters.get(mailAWTParameterName(name))
}
func (p *MailAWTMimeParameters) Set(name, value *string) {
	if p == nil {
		panic(NewNullPointerException())
	}

	p.parameters.set(mailAWTParameterName(name), value)
}
func (p *MailAWTMimeParameters) Remove(name *string) {
	if p == nil {
		panic(NewNullPointerException())
	}
	p.parameters.remove(mailAWTParameterName(name))
}
func (p *MailAWTMimeParameters) GetNames() *MailAWTParameterNames {
	if p == nil {
		panic(NewNullPointerException())
	}
	return p.parameters.names()
}
func (p *MailAWTMimeParameters) Clone() *MailAWTMimeParameters {
	if p == nil {
		panic(NewNullPointerException())
	}

	return &MailAWTMimeParameters{parameters: p.parameters.clone()}
}
func (p *MailAWTMimeParameters) String() string {
	if p == nil {
		panic(NewNullPointerException())
	}

	var b strings.Builder
	names := p.GetNames()
	for names.HasMoreElements() {
		name := names.NextElement()
		b.WriteString("; ")
		b.WriteString(name)
		b.WriteByte('=')
		b.WriteString(mailAWTQuote(p.parameters.get(name)))
	}
	return b.String()
}
func (p *MailAWTMimeParameters) HashCode() int32 {
	if p == nil {
		panic(NewNullPointerException())
	}

	code := int32(2147483647 / 45)
	names := p.GetNames()
	for names.HasMoreElements() {
		name := names.NextElement()
		code += mailAWTStringHash(name)
		value := p.Get(javaString(name))
		if value == nil {
			panic(NewNullPointerException())
		}
		code += mailAWTStringHash(*value)
	}
	return code
}
func (p *MailAWTMimeParameters) Equals(other *MailAWTMimeParameters) bool {
	if p == nil {
		panic(NewNullPointerException())
	}

	if other == nil || p.Size() != other.Size() {
		return false
	}
	names := p.GetNames()
	for names.HasMoreElements() {
		name, a := names.nextEntry()
		b := other.parameters.get(name)
		if !mailNullableStringsEqual(a, b) {
			return false
		}
	}
	return true
}
