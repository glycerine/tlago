/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

// Bundled Geronimo Activation MimeType and MimeTypeParameterList. These are
// distinct from JavaMail ContentType/ParameterList, including token validation,
// case-sensitive base types, parser failure state and default-locale name case.
// Source: apache/geronimo-specs geronimo-activation_1.1_spec javax.activation.
package tlc

import (
	"fmt"
	"strings"
)

type MailMimeTypeParseException struct{ javaExceptionBase }

func NewMailMimeTypeParseException(message ...string) *MailMimeTypeParseException {
	return &MailMimeTypeParseException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *MailMimeTypeParseException) Error() string { return javaThrowableMessage(e) }

// Character.isWhitespace(char), excluding Java's three nonbreaking spaces.
func mailActivationWhitespace(c uint16) bool {
	return c >= 9 && c <= 13 || c >= 0x1c && c <= 0x20 || c == 0x1680 || c >= 0x2000 && c <= 0x2006 || c >= 0x2008 && c <= 0x200a || c == 0x2028 || c == 0x2029 || c == 0x205f || c == 0x3000
}
func mailActivationISOControl(c uint16) bool { return c <= 0x1f || c >= 0x7f && c <= 0x9f }
func mailActivationSpecial(c uint16) bool {
	return mailActivationWhitespace(c) || mailActivationISOControl(c) || strings.ContainsRune("()<>@,;:\\\"/[]?=", rune(c))
}

// Full JVM Locale.getDefault lifetime/custom locale providers use the callback.
// The native English lowercaser keeps isolated UTF-16 units losslessly.
func mailActivationLower(value string) string {
	if f := DefaultMailActivationEnvironment.Lowercase; f != nil {
		return f(value)
	}
	return mailJDKEnglishLower(value)
}

type MailActivationParameterList struct {
	params *mailParameterHashTable[*string]
}

func NewMailActivationParameterList(value ...*string) (*MailActivationParameterList, error) {
	p := &MailActivationParameterList{params: newMailParameterHashTable[*string]()}
	if len(value) > 0 {
		if e := p.Parse(value[0]); e != nil {
			return nil, e
		}
	}
	return p, nil
}
func (p *MailActivationParameterList) Size() int     { return p.params.items.Len() }
func (p *MailActivationParameterList) IsEmpty() bool { return p.Size() == 0 }
func (p *MailActivationParameterList) Get(name *string) *string {
	if name == nil {
		panic(NewNullPointerException())
	}
	return copyJavaMessage(p.params.items.Get(mailActivationLower(*name)))
}
func (p *MailActivationParameterList) Set(name, value *string) {
	if name == nil {
		panic(NewNullPointerException())
	}
	p.params.set(mailActivationLower(*name), copyJavaMessage(value))
}
func (p *MailActivationParameterList) Remove(name *string) {
	if name == nil {
		panic(NewNullPointerException())
	}
	p.params.remove(mailActivationLower(*name))
}
func (p *MailActivationParameterList) GetNames() []string { return p.params.names() }

type MailActivationParameterNames struct {
	owner   *MailActivationParameterList
	names   []string
	index   int
	version int32
}

func (p *MailActivationParameterList) NamesEnumeration() *MailActivationParameterNames {
	return &MailActivationParameterNames{owner: p, names: p.GetNames(), version: p.params.version}
}
func (e *MailActivationParameterNames) HasMoreElements() bool { return e.index < len(e.names) }
func (e *MailActivationParameterNames) NextElement() (string, error) {
	if e.version != e.owner.params.version {
		return "", NewConcurrentModificationException()
	}
	if e.index == len(e.names) {
		return "", NewNoSuchElementException()
	}
	v := e.names[e.index]
	e.index++
	return v, nil
}
func (p *MailActivationParameterList) String() string {
	var out strings.Builder
	for _, name := range p.params.names() {
		out.WriteString("; ")
		out.WriteString(name)
		out.WriteByte('=')
		value := p.params.items.Get(name)
		if value == nil {
			panic(NewNullPointerException())
		}
		units := mailAddressUTF16(*value)
		quote := false
		for _, c := range units {
			if mailActivationSpecial(c) {
				quote = true
				break
			}
		}
		if quote {
			escaped := []uint16{'"'}
			for _, c := range units {
				if c == '\\' || c == '"' {
					escaped = append(escaped, '\\')
				}
				escaped = append(escaped, c)
			}
			escaped = append(escaped, '"')
			out.WriteString(mailAddressUTF16String(escaped))
		} else {
			out.WriteString(*value)
		}
	}
	return out.String()
}

type mailActivationRFC2045Parser struct {
	text  string
	units []uint16
	index int
}

func (p *mailActivationRFC2045Parser) expected(what string) error {
	return NewMailMimeTypeParseException(fmt.Sprintf("Expected %s at %d in %s", what, p.index-1, p.text))
}
func (p *mailActivationRFC2045Parser) first(what string) (uint16, error) {
	for {
		if p.index == len(p.units) {
			return 0, p.expected(what)
		}
		c := p.units[p.index]
		p.index++
		if !mailActivationWhitespace(c) {
			return c, nil
		}
	}
}
func (p *mailActivationRFC2045Parser) more() (bool, error) {
	for {
		if p.index == len(p.units) {
			return false, nil
		}
		c := p.units[p.index]
		p.index++
		if mailActivationWhitespace(c) {
			continue
		}
		if c != ';' {
			return false, p.expected("\";\"")
		}
		return true, nil
	}
}
func (p *mailActivationRFC2045Parser) token() string {
	start := p.index - 1
	for p.index != len(p.units) && !mailActivationSpecial(p.units[p.index]) {
		p.index++
	}
	return mailAddressUTF16String(p.units[start:p.index])
}
func (p *mailActivationRFC2045Parser) value() (string, error) {
	c, e := p.first("value")
	if e != nil {
		return "", e
	}
	if c != '"' {
		return p.token(), nil
	}
	out := []uint16{}
	for {
		if p.index == len(p.units) {
			return "", p.expected("closing quote")
		}
		c = p.units[p.index]
		p.index++
		if c == '"' {
			return mailAddressUTF16String(out), nil
		}
		if c == '\\' {
			if p.index == len(p.units) {
				return "", p.expected("escaped char")
			}
			c = p.units[p.index]
			p.index++
		}
		out = append(out, c)
	}
}

// Protected Java parse appends to the existing map and keeps earlier parameters
// when a later parameter fails. readExternal uses the same retained map.
func (p *MailActivationParameterList) Parse(value *string) error {
	if value == nil {
		return NewMailMimeTypeParseException("parameterList is null")
	}
	parser := mailActivationRFC2045Parser{text: *value, units: mailAddressUTF16(*value)}
	for {
		more, e := parser.more()
		if e != nil {
			return e
		}
		if !more {
			return nil
		}
		_, e = parser.first("attribute")
		if e != nil {
			return e
		}
		attribute := parser.token()
		c, e := parser.first("\"=\"")
		if e != nil {
			return e
		}
		if c != '=' {
			return parser.expected("\"=\"")
		}
		v, e := parser.value()
		if e != nil {
			return e
		}
		p.params.set(mailActivationLower(attribute), javaString(v))
	}
}

type MailActivationMimeType struct {
	primaryType, subType string
	parameters           *MailActivationParameterList
}

func NewMailActivationMimeType(value ...*string) (*MailActivationMimeType, error) {
	return mailInvoke(func() (*MailActivationMimeType, error) {
		p, _ := NewMailActivationParameterList()
		m := &MailActivationMimeType{primaryType: "application", subType: "*", parameters: p}
		if len(value) == 1 {
			if e := m.parse(value[0]); e != nil {
				return nil, e
			}
		} else if len(value) > 1 {
			if e := m.SetPrimaryType(value[0]); e != nil {
				return nil, e
			}
			if e := m.SetSubType(value[1]); e != nil {
				return nil, e
			}
		}
		return m, nil
	})
}
func mailActivationParseToken(value *string) (string, error) {
	if value == nil {
		return "", NewNullPointerException()
	}
	token := mailAddressTrim(*value)
	for _, c := range mailAddressUTF16(token) {
		if mailActivationSpecial(c) {
			return "", NewMailMimeTypeParseException("Special '" + mailAddressUTF16String([]uint16{c}) + "' not allowed in token")
		}
	}
	return token, nil
}
func (m *MailActivationMimeType) SetPrimaryType(value *string) error {
	v, e := mailActivationParseToken(value)
	if e == nil {
		m.primaryType = v
	}
	return e
}
func (m *MailActivationMimeType) SetSubType(value *string) error {
	v, e := mailActivationParseToken(value)
	if e == nil {
		m.subType = v
	}
	return e
}
func (m *MailActivationMimeType) GetPrimaryType() string { return m.primaryType }
func (m *MailActivationMimeType) GetSubType() string     { return m.subType }
func (m *MailActivationMimeType) GetBaseType() string {
	return m.GetPrimaryType() + "/" + m.GetSubType()
}
func (m *MailActivationMimeType) GetParameters() *MailActivationParameterList { return m.parameters }
func (m *MailActivationMimeType) GetParameter(name *string) *string           { return m.parameters.Get(name) }
func (m *MailActivationMimeType) SetParameter(name, value *string)            { m.parameters.Set(name, value) }
func (m *MailActivationMimeType) RemoveParameter(name *string)                { m.parameters.Remove(name) }
func (m *MailActivationMimeType) String() string                              { return m.GetBaseType() + m.parameters.String() }
func (m *MailActivationMimeType) Match(other *MailActivationMimeType) bool {
	if other == nil {
		panic(NewNullPointerException())
	}
	return m.primaryType == other.primaryType && (m.subType == "*" || other.subType == "*" || m.subType == other.subType)
}
func (m *MailActivationMimeType) MatchString(value *string) (bool, error) {
	other, e := NewMailActivationMimeType(value)
	if e != nil {
		return false, e
	}
	return m.Match(other), nil
}
func (m *MailActivationMimeType) parse(value *string) error {
	if value == nil {
		return NewNullPointerException()
	}
	units := mailAddressUTF16(*value)
	slash := -1
	for i, c := range units {
		if c == '/' {
			slash = i
			break
		}
	}
	if slash < 0 {
		return NewMailMimeTypeParseException("Expected '/'")
	}
	if e := m.SetPrimaryType(javaString(mailAddressUTF16String(units[:slash]))); e != nil {
		return e
	}
	semi := -1
	for i := slash + 1; i < len(units); i++ {
		if units[i] == ';' {
			semi = i
			break
		}
	}
	if semi < 0 {
		return m.SetSubType(javaString(mailAddressUTF16String(units[slash+1:])))
	}
	if e := m.SetSubType(javaString(mailAddressUTF16String(units[slash+1 : semi]))); e != nil {
		return e
	}
	return m.parameters.Parse(javaString(mailAddressUTF16String(units[semi:])))
}

// Native ObjectInput/ObjectOutput boundaries used by the Externalizable methods.
// Full JVM object serialization is a separate provider; these methods preserve
// writeUTF/flush ordering and parse-to-IOException conversion without closing.
type MailActivationObjectInput struct{ ReadUTF func() (*string, error) }
type MailActivationObjectOutput struct {
	WriteUTF func(string) error
	Flush    func() error
}

func (m *MailActivationMimeType) WriteExternal(out *MailActivationObjectOutput) error {
	_, e := mailInvoke(func() (bool, error) {
		if out == nil || out.WriteUTF == nil {
			return false, NewNullPointerException()
		}
		if e := out.WriteUTF(m.String()); e != nil {
			return false, e
		}
		if out.Flush == nil {
			return false, NewNullPointerException()
		}
		return true, out.Flush()
	})
	return e
}
func (m *MailActivationMimeType) ReadExternal(in *MailActivationObjectInput) error {
	_, e := mailInvoke(func() (bool, error) {
		if in == nil || in.ReadUTF == nil {
			return false, NewNullPointerException()
		}
		v, e := in.ReadUTF()
		if e != nil {
			return false, e
		}
		return true, m.parse(v)
	})
	if p, ok := e.(*MailMimeTypeParseException); ok {
		failure := NewIOException()
		failure.Message = copyJavaMessage(p.GetMessage())
		return failure
	}
	return e
}
