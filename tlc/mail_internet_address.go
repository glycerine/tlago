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

// Go port of JavaMail 1.6.8 InternetAddress parsing and validation.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/InternetAddress.java

package tlc

import (
	"strings"
	"sync"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var mailAddressClass struct {
	sync.Once
	ignoreBogusGroupName, useCanonicalHostName, allowUTF8 bool
}

func initializeMailAddressProperties() {
	mailAddressClass.Do(func() {
		mailAddressClass.ignoreBogusGroupName = mailBooleanProperty("mail.mime.address.ignorebogusgroupname", true)
		mailAddressClass.useCanonicalHostName = mailBooleanProperty("mail.mime.address.usecanonicalhostname", true)
		mailAddressClass.allowUTF8 = mailBooleanProperty("mail.mime.allowutf8", false)
	})
}
func mailBooleanProperty(name string, fallback bool) bool {
	value, present := tlcLookupSystemProperty(name)
	if !present {
		return fallback
	}
	if fallback {
		return !mailAddressEqualsIgnoreCase(value, "false")
	}
	return mailAddressEqualsIgnoreCase(value, "true")
}

// ParseMailInternetAddresses is InternetAddress.parse, whose default is strict
// list parsing. Detailed mailbox validation is a separate operation; a simple
// local name is accepted here even in strict list mode.
func ParseMailInternetAddresses(value string, strict ...bool) ([]*MailInternetAddress, error) {
	enforce := true
	if len(strict) > 0 {
		enforce = strict[0]
	}
	return parseMailInternetAddresses(value, enforce, false)
}
func NewMailInternetAddress(value string, validate ...bool) (*MailInternetAddress, error) {
	addresses, err := ParseMailInternetAddresses(value, true)
	if err != nil {
		return nil, err
	}
	if len(addresses) != 1 {
		return nil, mailAddressError("Illegal address", value, -1)
	}
	address := addresses[0]
	if len(validate) > 0 && validate[0] {
		if err = address.Validate(); err != nil {
			return nil, err
		}
	}
	return address, nil
}
func ParseMailInternetAddressHeader(value string, strict bool) ([]*MailInternetAddress, error) {
	initializeMailAddressProperties()
	return parseMailInternetAddresses(UnfoldMailHeader(value), strict, true)
}
func (m *MailInternetAddress) IsGroup() bool {
	if m == nil {
		panic(NewNullPointerException())
	}
	return strings.HasSuffix(m.Address, ";") && strings.IndexByte(m.Address, ':') > 0
}
func (m *MailInternetAddress) GetGroup(strict bool) ([]*MailInternetAddress, error) {
	if m == nil {
		return nil, NewNullPointerException()
	}
	if !strings.HasSuffix(m.Address, ";") {
		return nil, nil
	}
	i := strings.IndexByte(m.Address, ':')
	if i < 0 {
		return nil, nil
	}
	return ParseMailInternetAddressHeader(m.Address[i+1:len(m.Address)-1], strict)
}
func (m *MailInternetAddress) Validate() error {
	if m == nil {
		return NewNullPointerException()
	}
	if m.IsGroup() {
		_, err := m.GetGroup(true)
		return err
	}
	return CheckMailInternetAddress(&m.Address, true, true)
}
func mailAddressError(message, ref string, position int) *MailAddressException {
	return NewMailAddressException(javaString(message), javaString(ref), position)
}

func parseMailInternetAddresses(value string, strict, header bool) (addresses []*MailInternetAddress, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			addresses = nil
			err = panicValueAsError(failure)
		}
	}()
	initializeMailAddressProperties()
	chars := mailAddressUTF16(value)
	length := len(chars)
	ignoreErrors := header && !strict
	inGroup, routeAddress, rfc822 := false, false, false
	start, end, startPersonal, endPersonal := -1, -1, -1, -1
	addresses = []*MailInternetAddress{}
	substring := func(a, b int) string { return mailAddressUTF16String(chars[a:b]) }
	addToken := func() error {
		addr := mailAddressTrim(substring(start, end))
		var personal *string
		if rfc822 && startPersonal >= 0 {
			p := mailAddressUnquote(mailAddressTrim(substring(startPersonal, endPersonal)))
			if mailAddressTrim(p) != "" {
				personal = javaString(p)
			}
		}
		if header && !strict && personal != nil && strings.Contains(*personal, "@") && !strings.ContainsAny(addr, "@!") {
			addr, *personal = *personal, addr
		}
		if rfc822 || strict || header {
			if !ignoreErrors {
				if e := CheckMailInternetAddress(&addr, routeAddress, false); e != nil {
					return e
				}
			}
			addresses = append(addresses, &MailInternetAddress{Address: addr, EncodedPersonal: personal})
		} else {
			for _, token := range strings.FieldsFunc(addr, func(r rune) bool { return strings.ContainsRune(" \t\n\r\f", r) }) {
				if e := CheckMailInternetAddress(&token, false, false); e != nil {
					return e
				}
				addresses = append(addresses, &MailInternetAddress{Address: token})
			}
		}
		return nil
	}
	for index := 0; index < length; index++ {
		c := chars[index]
		switch c {
		case '(':
			rfc822 = true
			if start >= 0 && end == -1 {
				end = index
			}
			pindex := index
			index++
			nesting := 1
			for index < length && nesting > 0 {
				c = chars[index]
				switch c {
				case '\\':
					index++
				case '(':
					nesting++
				case ')':
					nesting--
				}
				index++
			}
			if nesting > 0 {
				if !ignoreErrors {
					return nil, mailAddressError("Missing ')'", value, index)
				}
				index = pindex + 1
				continue
			}
			index--
			if startPersonal == -1 {
				startPersonal = pindex + 1
			}
			if endPersonal == -1 {
				endPersonal = index
			}
		case ')':
			if !ignoreErrors {
				return nil, mailAddressError("Missing '('", value, index)
			}
			if start == -1 {
				start = index
			}
		case '<':
			rfc822 = true
			if routeAddress {
				if !ignoreErrors {
					return nil, mailAddressError("Extra route-addr", value, index)
				}
				if start == -1 {
					routeAddress = false
					rfc822 = false
					start, end = -1, -1
					continue
				}
				if !inGroup {
					if end == -1 {
						end = index
					}
					a := &MailInternetAddress{Address: mailAddressTrim(substring(start, end))}
					if startPersonal >= 0 {
						a.EncodedPersonal = javaString(mailAddressUnquote(mailAddressTrim(substring(startPersonal, endPersonal))))
					}
					addresses = append(addresses, a)
					routeAddress = false
					rfc822 = false
					start, end = -1, -1
					startPersonal, endPersonal = -1, -1
				}
			}
			rindex := index
			inquote := false
			index++
			for index < length {
				c = chars[index]
				if c == '\\' {
					index++
				} else if c == '"' {
					inquote = !inquote
				} else if c == '>' && !inquote {
					break
				}
				index++
			}
			if inquote {
				if !ignoreErrors {
					return nil, mailAddressError("Missing '\"'", value, index)
				}
				for index = rindex + 1; index < length; index++ {
					c = chars[index]
					if c == '\\' {
						index++
					} else if c == '>' {
						break
					}
				}
			}
			if index >= length {
				if !ignoreErrors {
					return nil, mailAddressError("Missing '>'", value, index)
				}
				index = rindex + 1
				if start == -1 {
					start = rindex
				}
				continue
			}
			if !inGroup {
				if start >= 0 {
					startPersonal = start
					endPersonal = rindex
				}
				start = rindex + 1
			}
			routeAddress = true
			end = index
		case '>':
			if !ignoreErrors {
				return nil, mailAddressError("Missing '<'", value, index)
			}
			if start == -1 {
				start = index
			}
		case '"':
			qindex := index
			rfc822 = true
			if start == -1 {
				start = index
			}
			index++
			for index < length {
				c = chars[index]
				if c == '\\' {
					index++
				} else if c == '"' {
					break
				}
				index++
			}
			if index >= length {
				if !ignoreErrors {
					return nil, mailAddressError("Missing '\"'", value, index)
				}
				index = qindex + 1
			}
		case '[':
			lindex := index
			rfc822 = true
			if start == -1 {
				start = index
			}
			index++
			for index < length {
				c = chars[index]
				if c == '\\' {
					index++
				} else if c == ']' {
					break
				}
				index++
			}
			if index >= length {
				if !ignoreErrors {
					return nil, mailAddressError("Missing ']'", value, index)
				}
				index = lindex + 1
			}
		case ';':
			if start == -1 {
				routeAddress = false
				rfc822 = false
				start, end = -1, -1
				continue
			}
			if inGroup {
				inGroup = false
				if header && !strict && index+1 < length && chars[index+1] == '@' {
					continue
				}
				end = index + 1
				addresses = append(addresses, &MailInternetAddress{Address: mailAddressTrim(substring(start, end))})
				routeAddress = false
				rfc822 = false
				start, end = -1, -1
				startPersonal, endPersonal = -1, -1
				continue
			}
			if !ignoreErrors {
				return nil, mailAddressError("Illegal semicolon, not in group", value, index)
			}
			fallthrough
		case ',':
			if start == -1 {
				routeAddress = false
				rfc822 = false
				start, end = -1, -1
				continue
			}
			if inGroup {
				routeAddress = false
				continue
			}
			if end == -1 {
				end = index
			}
			if e := addToken(); e != nil {
				return nil, e
			}
			routeAddress = false
			rfc822 = false
			start, end = -1, -1
			startPersonal, endPersonal = -1, -1
		case ':':
			rfc822 = true
			if inGroup && !ignoreErrors {
				return nil, mailAddressError("Nested group", value, index)
			}
			if start == -1 {
				start = index
			}
			if header && !strict {
				if index+1 < length {
					specials := ")>[]:@\\,."
					next := chars[index+1]
					if strings.ContainsRune(specials, rune(next)) {
						if next != '@' {
							continue
						}
						for i := index + 2; i < length; i++ {
							next = chars[i]
							if next == ';' || strings.ContainsRune(specials, rune(next)) {
								break
							}
						}
						if next == ';' {
							continue
						}
					}
				}
				groupName := substring(start, index)
				if mailAddressClass.ignoreBogusGroupName && (mailAddressEqualsIgnoreCase(groupName, "mailto") || mailAddressEqualsIgnoreCase(groupName, "From") || mailAddressEqualsIgnoreCase(groupName, "To") || mailAddressEqualsIgnoreCase(groupName, "Cc") || mailAddressEqualsIgnoreCase(groupName, "Subject") || mailAddressEqualsIgnoreCase(groupName, "Re")) {
					start = -1
				} else {
					inGroup = true
				}
			} else {
				inGroup = true
			}
		case ' ', '\t', '\r', '\n':
		default:
			if start == -1 {
				start = index
			}
		}
	}
	if start >= 0 {
		if end == -1 {
			end = length
		}
		if e := addToken(); e != nil {
			return nil, e
		}
	}
	return addresses, nil
}

// CheckMailInternetAddress is the dependency's mailbox checker. route permits
// the old RFC822 source-route syntax; validate adds final-domain/quote checks.
func CheckMailInternetAddress(address *string, route, validate bool) error {
	initializeMailAddressProperties()
	if address == nil {
		return NewMailAddressException(javaString("Address is null"), nil, -1)
	}
	addr := *address
	chars := mailAddressUTF16(addr)
	length := len(chars)
	if length == 0 {
		return mailAddressError("Empty address", addr, -1)
	}
	fail := func(message string) error { return mailAddressError(message, addr, -1) }
	start := 0
	if route && chars[0] == '@' {
		for {
			i := -1
			for j := start; j < length; j++ {
				if chars[j] == ',' || chars[j] == ':' {
					i = j
					break
				}
			}
			if i < 0 {
				break
			}
			if chars[start] != '@' {
				return fail("Illegal route-addr")
			}
			start = i + 1
			if chars[i] == ':' {
				break
			}
		}
	}
	c, last := uint16(0xffff), uint16(0xffff)
	inquote := false
	i := start
	for ; i < length; i++ {
		last = c
		c = chars[i]
		if c == '\\' || last == '\\' {
			continue
		}
		if c == '"' {
			if inquote {
				if validate && i+1 < length && chars[i+1] != '@' {
					return fail("Quote not at end of local address")
				}
				inquote = false
			} else {
				if validate && i != 0 {
					return fail("Quote not at start of local address")
				}
				inquote = true
			}
			continue
		} else if c == '\r' {
			if i+1 < length && chars[i+1] != '\n' {
				return fail("Quoted local address contains CR without LF")
			}
		} else if c == '\n' {
			if i+1 < length && chars[i+1] != ' ' && chars[i+1] != '\t' {
				return fail("Quoted local address contains newline without whitespace")
			}
		}
		if inquote {
			continue
		}
		if c == '.' {
			if i == start {
				return fail("Local address starts with dot")
			}
			if last == '.' {
				return fail("Local address contains dot-dot")
			}
		}
		if c == '@' {
			if i == 0 {
				return fail("Missing local name")
			}
			if last == '.' {
				return fail("Local address ends with dot")
			}
			break
		}
		if c <= 0x20 || c == 0x7f {
			return fail("Local address contains control or whitespace")
		}
		if strings.ContainsRune("()<>,;:\\\"[]@", rune(c)) {
			return fail("Local address contains illegal character")
		}
	}
	if inquote {
		return fail("Unterminated quote")
	}
	if c != '@' {
		if validate {
			return fail("Missing final '@domain'")
		}
		return nil
	}
	start = i + 1
	if start >= length {
		return fail("Missing domain")
	}
	if chars[start] == '.' {
		return fail("Domain starts with dot")
	}
	inLiteral := false
	for i = start; i < length; i++ {
		c = chars[i]
		if c == '[' {
			if i != start {
				return fail("Domain literal not at start of domain")
			}
			inLiteral = true
		} else if c == ']' {
			if i != length-1 {
				return fail("Domain literal end not at end of domain")
			}
			inLiteral = false
		} else if c <= 0x20 || c == 0x7f {
			return fail("Domain contains control or whitespace")
		} else if !inLiteral {
			if !(unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '-' || c == '.') {
				return fail("Domain contains illegal character")
			}
			if c == '.' && last == '.' {
				return fail("Domain contains dot-dot")
			}
		}
		last = c
	}
	if last == '.' {
		return fail("Domain ends with dot")
	}
	return nil
}

// String.equalsIgnoreCase compares UTF-16 length and applies upper-then-lower
// character mappings. Unicode simple folding differs for dotted/dotless I.
func mailAddressEqualsIgnoreCase(left, right string) bool {
	a, b := mailAddressUTF16(left), mailAddressUTF16(right)
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		d := b[i]
		if c == d {
			continue
		}
		upperC, upperD := unicode.ToUpper(rune(c)), unicode.ToUpper(rune(d))
		if upperC != upperD && unicode.ToLower(upperC) != unicode.ToLower(upperD) {
			return false
		}
	}
	return true
}

func mailAddressTrim(value string) string {
	chars := mailAddressUTF16(value)
	start, end := 0, len(chars)
	for start < end && chars[start] <= 0x20 {
		start++
	}
	for end > start && chars[end-1] <= 0x20 {
		end--
	}
	return mailAddressUTF16String(chars[start:end])
}
func mailAddressUnquote(value string) string {
	chars := mailAddressUTF16(value)
	if len(chars) > 1 && chars[0] == '"' && chars[len(chars)-1] == '"' {
		chars = chars[1 : len(chars)-1]
		unquoted := make([]uint16, 0, len(chars))
		for i := 0; i < len(chars); i++ {
			c := chars[i]
			if c == '\\' && i < len(chars)-1 {
				i++
				c = chars[i]
			}
			unquoted = append(unquoted, c)
		}
		chars = unquoted
	}
	return mailAddressUTF16String(chars)
}
func mailAddressUTF16(value string) []uint16 {
	chars := make([]uint16, 0, len(value))
	for len(value) > 0 {
		if len(value) >= 3 && value[0] == 0xed && value[1] >= 0xa0 && value[1] <= 0xbf && value[2]&0xc0 == 0x80 {
			chars = append(chars, uint16(value[0]&15)<<12|uint16(value[1]&63)<<6|uint16(value[2]&63))
			value = value[3:]
			continue
		}
		r, n := utf8.DecodeRuneInString(value)
		value = value[n:]
		if r <= 0xffff {
			chars = append(chars, uint16(r))
		} else {
			high, low := utf16.EncodeRune(r)
			chars = append(chars, uint16(high), uint16(low))
		}
	}
	return chars
}
func mailAddressUTF16String(chars []uint16) string {
	var data []byte
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if c >= 0xd800 && c <= 0xdbff && i+1 < len(chars) && chars[i+1] >= 0xdc00 && chars[i+1] <= 0xdfff {
			data = utf8.AppendRune(data, utf16.DecodeRune(rune(c), rune(chars[i+1])))
			i++
		} else if c >= 0xd800 && c <= 0xdfff {
			data = append(data, byte(0xe0|c>>12), byte(0x80|(c>>6)&63), byte(0x80|c&63))
		} else {
			data = utf8.AppendRune(data, rune(c))
		}
	}
	return string(data)
}
