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

// Go port of JavaMail 1.6.8 ParameterList.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/ParameterList.java

package tlc

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var mailParameterClass struct {
	sync.Once
	encode, decode, decodeStrict, apple, windows, strict, splitLong bool
}

func initializeMailParameterProperties() {
	mailParameterClass.Do(func() {
		mailParameterClass.encode = mailBooleanProperty("mail.mime.encodeparameters", true)
		mailParameterClass.decode = mailBooleanProperty("mail.mime.decodeparameters", true)
		mailParameterClass.decodeStrict = mailBooleanProperty("mail.mime.decodeparameters.strict", false)
		mailParameterClass.apple = mailBooleanProperty("mail.mime.applefilenames", false)
		mailParameterClass.windows = mailBooleanProperty("mail.mime.windowsfilenames", false)
		mailParameterClass.strict = mailBooleanProperty("mail.mime.parameters.strict", true)
		mailParameterClass.splitLong = mailBooleanProperty("mail.mime.splitlongparameters", true)
	})
}

type mailEncodedParameter struct{ value, charset, encoded *string }
type mailMultiParameter struct {
	value *string
	parts []mailParameterValue
}
type mailParameterValue struct {
	plain   *string
	encoded *mailEncodedParameter
	multi   *mailMultiParameter
	literal bool
}

func (v mailParameterValue) value() *string {
	if v.multi != nil {
		return v.multi.value
	}
	if v.encoded != nil {
		return v.encoded.value
	}
	return v.plain
}

// Values use InsMap; the auxiliary map/set traversal mirrors OpenJDK buckets
// and collision trees, including their retained capacity after clear.
type mailParameterHashTable[V any] struct {
	items   *InsMap[string, V]
	order   mailParameterHashOrder
	version int
}

func newMailParameterHashTable[V any]() *mailParameterHashTable[V] {
	return &mailParameterHashTable[V]{items: NewInsMap[string, V]()}
}
func parameterHash(value string) uint32 {
	var h uint32
	for _, c := range mailAddressUTF16(value) {
		h = 31*h + uint32(c)
	}
	return h ^ (h >> 16)
}
func (m *mailParameterHashTable[V]) set(name string, value V) {
	if m.items.Set(name, value) {
		m.version++
		m.order.add(name)
	}
}
func (m *mailParameterHashTable[V]) remove(name string) {
	if found, _ := m.items.Delkey(name); found {
		m.version++
		m.order.remove(name)
	}
}
func (m *mailParameterHashTable[V]) clear()          { m.items.DeleteAll(); m.version++; m.order.clear() }
func (m *mailParameterHashTable[V]) names() []string { return m.order.names() }

type MailParameterList struct {
	list       *InsMap[string, mailParameterValue]
	segments   *mailParameterHashTable[mailParameterValue]
	multiNames *mailParameterHashTable[bool]
	lastName   *string
	version    int
}

func NewMailParameterList(value ...*string) (p *MailParameterList, err error) {
	initializeMailParameterProperties()
	p = &MailParameterList{list: NewInsMap[string, mailParameterValue]()}
	if mailParameterClass.decode {
		p.segments = newMailParameterHashTable[mailParameterValue]()
		p.multiNames = newMailParameterHashTable[bool]()
	}
	if len(value) == 0 {
		return p, nil
	}
	if err = p.parse(value[0]); err != nil {
		return nil, err
	}
	return p, nil
}
func (p *MailParameterList) parse(value *string) error {
	h := NewMailHeaderTokenizer(value, MailHeaderMIME)
	bad := func(expected string, tk *MailHeaderToken) error {
		return NewMailParseException("In parameter list <" + javaNullableString(value) + ">, expected " + expected + ", got \"" + javaNullableString(tk.Value) + "\"")
	}
	for {
		tk, err := h.Next()
		if err != nil {
			return err
		}
		if tk.Type == MailHeaderEOF {
			break
		}
		if uint16(tk.Type) == ';' {
			tk, err = h.Next()
			if err != nil {
				return err
			}
			if tk.Type == MailHeaderEOF {
				break
			}
			if tk.Type != MailHeaderAtom {
				return bad("parameter name", tk)
			}
			name := mailCharsetLower(*tk.Value)
			tk, err = h.Next()
			if err != nil {
				return err
			}
			if uint16(tk.Type) != '=' {
				return bad("'='", tk)
			}
			if mailParameterClass.windows && (name == "name" || name == "filename") {
				tk, err = h.NextWithEscapes(';', true)
			} else if mailParameterClass.strict {
				tk, err = h.Next()
			} else {
				tk, err = h.Next(';')
			}
			if err != nil {
				return err
			}
			if tk.Type != MailHeaderAtom && tk.Type != MailHeaderQuotedString {
				return bad("parameter value", tk)
			}
			p.lastName = javaString(name)
			if mailParameterClass.decode {
				if err = p.putEncodedName(name, tk.Value); err != nil {
					return err
				}
			} else {
				p.setParameter(name, mailParameterValue{plain: tk.Value})
			}
		} else if tk.Type == MailHeaderAtom && p.lastName != nil && ((mailParameterClass.apple && (*p.lastName == "name" || *p.lastName == "filename")) || !mailParameterClass.strict) {
			previous, _ := p.list.Get2(*p.lastName)
			if previous.encoded != nil || previous.multi != nil || previous.literal {
				return NewClassCastException("class javax.mail.internet.ParameterList$Value cannot be cast to class java.lang.String")
			}
			p.setParameter(*p.lastName, mailParameterValue{plain: javaString(javaNullableString(previous.plain) + " " + *tk.Value)})
		} else {
			return bad("';'", tk)
		}
	}
	if mailParameterClass.decode {
		return p.combine(false)
	}
	return nil
}
func (p *MailParameterList) putEncodedName(name string, value *string) error {
	star := strings.IndexByte(name, '*')
	if star < 0 {
		p.setParameter(name, mailParameterValue{plain: copyJavaMessage(value)})
		return nil
	}
	if star == len(name)-1 {
		v, err := extractMailParameterCharset(value)
		if err != nil {
			return err
		}
		decoded, err := decodeMailParameterBytes(v.value, v.charset)
		if err != nil {
			if _, ok := err.(*UnsupportedEncodingException); !ok {
				return err
			}
			if mailParameterClass.decodeStrict {
				return NewMailParseException(javaThrowableString(err))
			}
		} else {
			v.value = javaString(decoded)
		}
		p.setParameter(name[:star], mailParameterValue{encoded: v})
		return nil
	}
	root := name[:star]
	p.multiNames.set(root, true)
	p.setParameter(root, mailParameterValue{plain: javaString("")})
	v := mailParameterValue{plain: copyJavaMessage(value)}
	if strings.HasSuffix(name, "*") {
		if strings.HasSuffix(name, "*0*") {
			encoded, err := extractMailParameterCharset(value)
			if err != nil {
				return err
			}
			v = mailParameterValue{encoded: encoded}
		} else {
			v = mailParameterValue{encoded: &mailEncodedParameter{value: copyJavaMessage(value), encoded: copyJavaMessage(value)}}
		}
		name = name[:len(name)-1]
	}
	p.segments.set(name, v)
	return nil
}
func (p *MailParameterList) combine(keep bool) (err error) {
	success := false
	defer func() {
		if !keep && !success {
			return
		}
		for _, name := range p.segments.names() {
			v, _ := p.segments.items.Get2(name)
			if v.encoded != nil {
				decoded, e := decodeMailParameterBytes(v.encoded.value, v.encoded.charset)
				if e != nil {
					if _, ok := e.(*UnsupportedEncodingException); !ok {
						err = e
						return
					}
					if mailParameterClass.decodeStrict {
						err = NewMailParseException(javaThrowableString(e))
						return
					}
				} else {
					v.encoded.value = javaString(decoded)
				}
			}
		}
		for _, name := range p.segments.names() {
			v, _ := p.segments.items.Get2(name)
			p.setParameter(name, v)
		}
		p.multiNames.clear()
		p.segments.clear()
	}()
	version := p.multiNames.version
	for _, name := range p.multiNames.names() {
		if version != p.multiNames.version {
			return NewConcurrentModificationException()
		}
		mv := &mailMultiParameter{}
		var charset *string
		bytes := []byte{}
		segment := 0
		for ; ; segment++ {
			key := name + "*" + strconv.Itoa(segment)
			v, ok := p.segments.items.Get2(key)
			if !ok {
				break
			}
			mv.parts = append(mv.parts, v)
			if v.encoded != nil {
				if segment == 0 {
					charset = v.encoded.charset
				} else if charset == nil {
					p.multiNames.remove(name)
					break
				}
				data, e := mailParameterPercentBytes(v.encoded.value)
				if e != nil {
					return e
				}
				bytes = append(bytes, data...)
			} else {
				if v.plain == nil {
					return NewNullPointerException()
				}
				for _, c := range mailAddressUTF16(*v.plain) {
					bytes = append(bytes, byte(c))
				}
			}
			p.segments.remove(key)
		}
		if segment == 0 {
			p.removeParameter(name)
		} else {
			cs := ""
			if charset != nil {
				cs = DefaultMailMIMECodec.JavaCharset(*charset)
			}
			if charset == nil || cs == "" {
				cs = DefaultMailMIMECodec.defaultJavaCharset()
			}
			decoded, e := mailParameterDecodeCharset(bytes, cs)
			if e != nil {
				if _, ok := e.(*UnsupportedEncodingException); !ok {
					return e
				}
				if mailParameterClass.decodeStrict {
					return NewMailParseException(javaThrowableString(e))
				}
				decoded, _ = decodeMailCharset(bytes, "iso-8859-1")
			}
			mv.value = javaString(decoded)
			p.setParameter(name, mailParameterValue{multi: mv})
		}
	}
	success = true
	return nil
}
func (p *MailParameterList) CombineSegments() error {
	if !mailParameterClass.decode || p.multiNames.items.Len() == 0 {
		return nil
	}
	err := p.combine(true)
	if _, ok := err.(*MailParseException); ok {
		return nil
	}
	return err
}
func (p *MailParameterList) setParameter(name string, value mailParameterValue) {
	if p.list.Set(name, value) {
		p.version++
	}
}
func (p *MailParameterList) removeParameter(name string) {
	if found, _ := p.list.Delkey(name); found {
		p.version++
	}
}

// NamesEnumeration ports ParamEnum over LinkedHashMap's fail-fast iterator.
// GetNames is the Go convenience snapshot for immediate enumeration consumers.
type MailParameterNames struct {
	owner          *MailParameterList
	names          []string
	index, version int
}

func (p *MailParameterList) NamesEnumeration() *MailParameterNames {
	return &MailParameterNames{owner: p, names: p.GetNames(), version: p.version}
}
func (e *MailParameterNames) HasMoreElements() bool { return e.index < len(e.names) }
func (e *MailParameterNames) NextElement() (string, error) {
	if e.owner.version != e.version {
		return "", NewConcurrentModificationException()
	}
	if e.index == len(e.names) {
		return "", NewNoSuchElementException()
	}
	name := e.names[e.index]
	e.index++
	return name, nil
}
func (p *MailParameterList) Size() int { return p.list.Len() }
func (p *MailParameterList) Get(name string) *string {
	v, ok := p.list.Get2(mailCharsetLower(mailAddressTrim(name)))
	if !ok {
		return nil
	}
	return copyJavaMessage(v.value())
}
func (p *MailParameterList) Set(name string, value *string, charset ...*string) error {
	name = mailCharsetLower(mailAddressTrim(name))
	if len(charset) > 0 && mailParameterClass.encode {
		v, err := encodeMailParameterValue(value, charset[0])
		if err != nil {
			return err
		}
		if v != nil {
			p.setParameter(name, mailParameterValue{encoded: v})
			return nil
		}
	}
	if mailParameterClass.decode {
		err := p.putEncodedName(name, value)
		if err != nil {
			if _, ok := err.(*MailParseException); !ok {
				return err
			}
			p.setParameter(name, mailParameterValue{plain: copyJavaMessage(value)})
		}
	} else {
		p.setParameter(name, mailParameterValue{plain: copyJavaMessage(value)})
	}
	return nil
}
func (p *MailParameterList) SetLiteral(name string, value *string) {
	p.setParameter(name, mailParameterValue{plain: copyJavaMessage(value), literal: true})
}
func (p *MailParameterList) Remove(name string) {
	p.removeParameter(mailCharsetLower(mailAddressTrim(name)))
}
func (p *MailParameterList) GetNames() []string {
	names := []string{}
	for name := range p.list.All() {
		names = append(names, name)
	}
	return names
}
func (p *MailParameterList) String() string { return p.ToString(0) }
func (p *MailParameterList) ToString(used int) string {
	b := mailParameterStringBuffer{used: int32(used)}
	for name, v := range p.list.All() {
		switch {
		case v.multi != nil:
			for i, part := range v.multi.parts {
				key := name + "*" + strconv.Itoa(i)
				value := part.plain
				if part.encoded != nil {
					key += "*"
					value = part.encoded.encoded
				}
				b.add(key, QuoteMailWord(value, MailHeaderMIME))
			}
		case v.literal:
			b.add(name, QuoteMailWord(v.plain, MailHeaderMIME))
		case v.encoded != nil:
			b.add(name+"*", QuoteMailWord(v.encoded.encoded, MailHeaderMIME))
		default:
			if v.plain == nil {
				panic(NewNullPointerException())
			}
			chars := mailAddressUTF16(*v.plain)
			if len(chars) > 60 && mailParameterClass.splitLong && mailParameterClass.encode {
				segment := 0
				for len(chars) > 60 {
					part := mailAddressUTF16String(chars[:60])
					b.add(name+"*"+strconv.Itoa(segment), QuoteMailWord(&part, MailHeaderMIME))
					chars = chars[60:]
					segment++
				}
				if len(chars) > 0 {
					part := mailAddressUTF16String(chars)
					b.add(name+"*"+strconv.Itoa(segment), QuoteMailWord(&part, MailHeaderMIME))
				}
			} else {
				b.add(name, QuoteMailWord(v.plain, MailHeaderMIME))
			}
		}
	}
	return b.text.String()
}

type mailParameterStringBuffer struct {
	used int32
	text strings.Builder
}

func (b *mailParameterStringBuffer) add(name, value string) {
	b.text.WriteString("; ")
	b.used += 2
	n := int32(len(mailAddressUTF16(name)))
	v := int32(len(mailAddressUTF16(value)))
	if b.used+n+v+1 > 76 {
		b.text.WriteString("\r\n\t")
		b.used = 8
	}
	b.text.WriteString(name)
	b.text.WriteByte('=')
	b.used += n + 1
	if b.used+v > 76 {
		s := FoldMailHeader(int(b.used), value)
		b.text.WriteString(s)
		units := mailAddressUTF16(s)
		last := -1
		for i, c := range units {
			if c == '\n' {
				last = i
			}
		}
		if last >= 0 {
			b.used += int32(len(units) - last - 1)
		} else {
			b.used += int32(len(units))
		}
	} else {
		b.text.WriteString(value)
		b.used += v
	}
}
func extractMailParameterCharset(value *string) (*mailEncodedParameter, error) {
	if value == nil {
		return nil, NewNullPointerException()
	}
	v := &mailEncodedParameter{value: copyJavaMessage(value), encoded: copyJavaMessage(value)}
	i := strings.IndexByte(*value, '\'')
	if i < 0 {
		if mailParameterClass.decodeStrict {
			return nil, NewMailParseException("Missing charset in encoded value: " + *value)
		}
		return v, nil
	}
	li := strings.IndexByte((*value)[i+1:], '\'')
	if li < 0 {
		if mailParameterClass.decodeStrict {
			return nil, NewMailParseException("Missing language in encoded value: " + *value)
		}
		return v, nil
	}
	li += i + 1
	v.charset = javaString((*value)[:i])
	v.value = javaString((*value)[li+1:])
	return v, nil
}
func encodeMailParameterValue(value, charset *string) (*mailEncodedParameter, error) {
	if value == nil {
		return nil, NewNullPointerException()
	}
	if checkMailASCII(*value) == 1 {
		return nil, nil
	}
	if charset == nil {
		return nil, NewNullPointerException()
	}
	encoder := DefaultMailMIMECodec.EncodeCharset
	if encoder == nil {
		encoder = encodeMailCharset
	}
	data, err := encoder(*value, DefaultMailMIMECodec.JavaCharset(*charset))
	if err != nil {
		if _, ok := err.(*UnsupportedEncodingException); ok {
			return nil, nil
		}
		return nil, err
	}
	var b strings.Builder
	b.WriteString(*charset)
	b.WriteString("''")
	for _, c := range data {
		if c <= ' ' || c >= 0x7f || c == '*' || c == '\'' || c == '%' || strings.ContainsRune(MailHeaderMIME, rune(c)) {
			fmt.Fprintf(&b, "%%%02X", c)
		} else {
			b.WriteByte(c)
		}
	}
	return &mailEncodedParameter{value: copyJavaMessage(value), charset: copyJavaMessage(charset), encoded: javaString(b.String())}, nil
}
func mailParameterDecodeCharset(data []byte, charset string) (string, error) {
	initializeMailMIMEProperties()
	decoder := DefaultMailMIMECodec.DecodeCharset
	if decoder == nil {
		decoder = decodeMailCharset
	}
	return decoder(data, charset)
}
func decodeMailParameterBytes(value, charset *string) (string, error) {
	data, err := mailParameterPercentBytes(value)
	if err != nil {
		return "", err
	}
	var cs string
	if charset != nil {
		cs = DefaultMailMIMECodec.JavaCharset(*charset)
	}
	if charset == nil || cs == "" {
		cs = DefaultMailMIMECodec.defaultJavaCharset()
	}
	return mailParameterDecodeCharset(data, cs)
}
func mailParameterPercentBytes(value *string) ([]byte, error) {
	if value == nil {
		return nil, NewNullPointerException()
	}
	chars := mailAddressUTF16(*value)
	out := make([]byte, 0, len(chars))
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if c == '%' {
			if i+3 > len(chars) {
				if mailParameterClass.decodeStrict {
					return nil, NewMailParseException(fmt.Sprintf("java.lang.StringIndexOutOfBoundsException: Range [%d, %d) out of bounds for length %d", i+1, i+3, len(chars)))
				}
			} else {
				hex := chars[i+1 : i+3]
				number, ok := mailParameterHex(hex)
				if ok {
					c = uint16(number)
					i += 2
				} else if mailParameterClass.decodeStrict {
					return nil, NewMailParseException("java.lang.NumberFormatException: For input string: \"" + mailAddressUTF16String(hex) + "\" under radix 16")
				}
			}
		}
		out = append(out, byte(c))
	}
	return out, nil
}
func mailParameterHex(chars []uint16) (int, bool) {
	negative := false
	start := 0
	if chars[0] == '-' || chars[0] == '+' {
		negative = chars[0] == '-'
		start = 1
	}
	number := 0
	for _, c := range chars[start:] {
		r := rune(c)
		digit := -1
		switch {
		case r >= '0' && r <= '9':
			digit = int(r - '0')
		case r >= 'a' && r <= 'f':
			digit = int(r-'a') + 10
		case r >= 'A' && r <= 'F':
			digit = int(r-'A') + 10
		case r >= 0xff10 && r <= 0xff19:
			digit = int(r - 0xff10)
		case r >= 0xff21 && r <= 0xff26:
			digit = int(r-0xff21) + 10
		case r >= 0xff41 && r <= 0xff46:
			digit = int(r-0xff41) + 10
		case unicode.IsDigit(r):
			for _, table := range []*unicode.RangeTable{unicode.Digit} {
				for _, rr := range table.R16 {
					if uint16(r) >= rr.Lo && uint16(r) <= rr.Hi && (uint16(r)-rr.Lo)%rr.Stride == 0 {
						digit = int((uint16(r)-rr.Lo)/rr.Stride) % 10
						break
					}
				}
			}
		}
		if digit < 0 || digit >= 16 {
			return 0, false
		}
		number = number*16 + digit
	}
	if negative {
		number = -number
	}
	return number, true
}
