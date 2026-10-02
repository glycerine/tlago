/*
 * Copyright (c) 1997, 2021, Oracle and/or its affiliates. All rights reserved.
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

// OpenJDK AWT MimeType. Unlike Geronimo Activation, this grammar rejects empty
// tokens and lowercases ASCII base types, and clones parameter lists on access.
package tlc

type MailAWTMimeType struct {
	primary, sub *string
	parameters   *MailAWTMimeParameters
}

func NewMailAWTMimeType(values ...*string) (*MailAWTMimeType, error) {
	return mailInvoke(func() (*MailAWTMimeType, error) {
		m := &MailAWTMimeType{}
		if len(values) == 1 {
			if err := m.Parse(values[0]); err != nil {
				return nil, err
			}
		}
		if len(values) == 2 {
			p, _ := NewMailAWTMimeParameters()
			return NewMailAWTMimeTypeWithParameters(values[0], values[1], p)
		}
		return m, nil
	})
}
func NewMailAWTMimeTypeWithParameters(primary, sub *string, params *MailAWTMimeParameters) (*MailAWTMimeType, error) {
	return mailInvoke(func() (*MailAWTMimeType, error) {
		if !mailAWTValidToken(primary) {
			return nil, NewMailAWTMimeTypeParseException("Primary type is invalid.")
		}
		p := mailCharsetLower(*primary)
		if !mailAWTValidToken(sub) {
			return nil, NewMailAWTMimeTypeParseException("Sub type is invalid.")
		}
		s := mailCharsetLower(*sub)
		if params == nil {
			return nil, NewNullPointerException()
		}
		return &MailAWTMimeType{primary: javaString(p), sub: javaString(s), parameters: params.Clone()}, nil
	})
}
func mailAWTValidToken(token *string) bool {
	if token == nil {
		panic(NewNullPointerException())
	}
	units := mailAddressUTF16(*token)
	if len(units) == 0 {
		return false
	}
	for _, c := range units {
		if !mailAWTToken(c) {
			return false
		}
	}
	return true
}
func (m *MailAWTMimeType) Parse(value *string) error {
	if m == nil {
		panic(NewNullPointerException())
	}

	_, err := mailInvoke(func() (bool, error) {
		if value == nil {
			return false, NewNullPointerException()
		}
		s := mailAddressUTF16(*value)
		slash, semi := -1, -1
		for i, c := range s {
			if c == '/' && slash < 0 {
				slash = i
			}
			if c == ';' && semi < 0 {
				semi = i
			}
		}
		if slash < 0 || semi >= 0 && slash > semi {
			return false, NewMailAWTMimeTypeParseException("Unable to find a sub type.")
		}
		end := len(s)
		if semi >= 0 {
			end = semi
		}
		m.primary = javaString(mailCharsetLower(mailAddressTrim(mailAddressUTF16String(s[:slash]))))
		m.sub = javaString(mailCharsetLower(mailAddressTrim(mailAddressUTF16String(s[slash+1 : end]))))
		var p *MailAWTMimeParameters
		var err error
		if semi >= 0 {
			p, err = NewMailAWTMimeParameters(javaString(mailAddressUTF16String(s[semi:])))
		} else {
			p, err = NewMailAWTMimeParameters()
		}
		if err != nil {
			return false, err
		}
		m.parameters = p
		if !mailAWTValidToken(m.primary) {
			return false, NewMailAWTMimeTypeParseException("Primary type is invalid.")
		}
		if !mailAWTValidToken(m.sub) {
			return false, NewMailAWTMimeTypeParseException("Sub type is invalid.")
		}
		return false, nil
	})
	return err
}
func (m *MailAWTMimeType) GetPrimaryType() *string {
	if m == nil {
		panic(NewNullPointerException())
	}
	return copyJavaMessage(m.primary)
}
func (m *MailAWTMimeType) GetSubType() *string {
	if m == nil {
		panic(NewNullPointerException())
	}
	return copyJavaMessage(m.sub)
}
func (m *MailAWTMimeType) GetParameters() *MailAWTMimeParameters {
	if m == nil {
		panic(NewNullPointerException())
	}

	if m.parameters == nil {
		panic(NewNullPointerException())
	}
	return m.parameters.Clone()
}
func (m *MailAWTMimeType) GetParameter(name *string) *string {
	if m == nil {
		panic(NewNullPointerException())
	}

	if m.parameters == nil {
		panic(NewNullPointerException())
	}
	return m.parameters.Get(name)
}
func (m *MailAWTMimeType) SetParameter(name, value *string) {
	if m == nil {
		panic(NewNullPointerException())
	}

	if m.parameters == nil {
		panic(NewNullPointerException())
	}
	m.parameters.Set(name, value)
}
func (m *MailAWTMimeType) RemoveParameter(name *string) {
	if m == nil {
		panic(NewNullPointerException())
	}

	if m.parameters == nil {
		panic(NewNullPointerException())
	}
	m.parameters.Remove(name)
}
func (m *MailAWTMimeType) GetBaseType() string {
	if m == nil {
		panic(NewNullPointerException())
	}

	return javaNullableString(m.primary) + "/" + javaNullableString(m.sub)
}
func (m *MailAWTMimeType) String() string {
	if m == nil {
		panic(NewNullPointerException())
	}

	if m.parameters == nil {
		panic(NewNullPointerException())
	}
	return m.GetBaseType() + m.parameters.String()
}
func (m *MailAWTMimeType) Match(other *MailAWTMimeType) bool {
	if m == nil {
		panic(NewNullPointerException())
	}

	if other == nil {
		return false
	}
	if m.primary == nil {
		panic(NewNullPointerException())
	}
	if !mailNullableStringsEqual(m.primary, other.primary) {
		return false
	}
	if m.sub == nil {
		panic(NewNullPointerException())
	}
	if *m.sub == "*" {
		return true
	}
	if other.sub == nil {
		panic(NewNullPointerException())
	}
	return *other.sub == "*" || *m.sub == *other.sub
}
func (m *MailAWTMimeType) MatchString(value *string) (bool, error) {
	if m == nil {
		panic(NewNullPointerException())
	}

	if value == nil {
		return false, nil
	}
	other, err := NewMailAWTMimeType(value)
	if err != nil {
		return false, err
	}
	return m.Match(other), nil
}
func (m *MailAWTMimeType) HashCode() int32 {
	if m == nil {
		panic(NewNullPointerException())
	}

	if m.primary == nil || m.sub == nil || m.parameters == nil {
		panic(NewNullPointerException())
	}
	return mailAWTStringHash(*m.primary) + mailAWTStringHash(*m.sub) + m.parameters.HashCode()
}
func (m *MailAWTMimeType) Equals(other *MailAWTMimeType) bool {
	if m == nil {
		panic(NewNullPointerException())
	}

	if other == nil {
		return false
	}
	if m.primary == nil {
		panic(NewNullPointerException())
	}
	if !mailNullableStringsEqual(m.primary, other.primary) {
		return false
	}
	if m.sub == nil {
		panic(NewNullPointerException())
	}
	if !mailNullableStringsEqual(m.sub, other.sub) {
		return false
	}
	if m.parameters == nil {
		panic(NewNullPointerException())
	}
	return m.parameters.Equals(other.parameters)
}
func (m *MailAWTMimeType) Clone() *MailAWTMimeType {
	if m == nil {
		panic(NewNullPointerException())
	}

	if m.parameters == nil {
		panic(NewNullPointerException())
	}
	return &MailAWTMimeType{primary: m.primary, sub: m.sub, parameters: m.parameters.Clone()}
}
func mailNullableStringsEqual(a, b *string) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

// Externalization's long-string/default-charset path and object streams are the
// next AWT serialization dependencies; no JavaMail UTF serialization is substituted.
