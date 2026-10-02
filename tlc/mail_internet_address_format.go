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

// Go port of JavaMail 1.6.8 InternetAddress personal-name and formatting methods.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/InternetAddress.java

package tlc

import "strings"

// NewMailInternetAddressPersonal assumes a valid raw address, as does Java's
// address/personal constructor. It must not call the parsing constructor.
func NewMailInternetAddressPersonal(address string, personal *string, charset ...*string) (*MailInternetAddress, error) {
	initializeMailAddressProperties()
	a := &MailInternetAddress{Address: address}
	if err := a.SetPersonal(personal, charset...); err != nil {
		return nil, err
	}
	return a, nil
}
func (m *MailInternetAddress) SetPersonal(value *string, charset ...*string) error {
	if m == nil {
		return NewNullPointerException()
	}
	initializeMailAddressProperties()
	m.Personal = copyJavaMessage(value)
	if value == nil {
		m.EncodedPersonal = nil
		return nil
	}
	var cs *string
	if len(charset) > 0 {
		cs = charset[0]
	}
	encoded, err := EncodeMailWord(*value, cs, nil)
	if err != nil {
		return err
	}
	m.EncodedPersonal = javaString(encoded)
	return nil
}
func (m *MailInternetAddress) GetPersonal() *string {
	if m == nil {
		panic(NewNullPointerException())
	}
	initializeMailAddressProperties()
	if m.Personal != nil {
		return copyJavaMessage(m.Personal)
	}
	if m.EncodedPersonal != nil {
		value, err := DecodeMailText(*m.EncodedPersonal)
		if err == nil {
			m.Personal = javaString(value)
			return copyJavaMessage(m.Personal)
		}
		if javaSystemFailureCode(err) != NoError {
			panic(err)
		}
		return copyJavaMessage(m.EncodedPersonal)
	}
	return nil
}
func (m *MailInternetAddress) String() string {
	if m == nil {
		panic(NewNullPointerException())
	}
	initializeMailAddressProperties()
	if m.EncodedPersonal == nil && m.Personal != nil {
		value, err := EncodeMailWord(*m.Personal, nil, nil)
		if err == nil {
			m.EncodedPersonal = javaString(value)
		} else if _, ok := err.(*UnsupportedEncodingException); !ok {
			panic(err)
		}
	}
	if m.EncodedPersonal != nil {
		return quoteMailPhrase(*m.EncodedPersonal) + " <" + m.Address + ">"
	}
	if m.IsGroup() || m.isSimple() {
		return m.Address
	}
	return "<" + m.Address + ">"
}
func (m *MailInternetAddress) ToUnicodeString() string {
	if m == nil {
		panic(NewNullPointerException())
	}
	if personal := m.GetPersonal(); personal != nil {
		return quoteMailPhrase(*personal) + " <" + m.Address + ">"
	}
	if m.IsGroup() || m.isSimple() {
		return m.Address
	}
	return "<" + m.Address + ">"
}
func (m *MailInternetAddress) isSimple() bool {
	return !strings.ContainsAny(m.Address, "()<>,;:\\\"[]")
}
func quoteMailPhrase(value string) string {
	chars := mailAddressUTF16(value)
	quote := false
	for _, c := range chars {
		if c == '"' || c == '\\' {
			out := []uint16{'"'}
			for _, cc := range chars {
				if cc == '"' || cc == '\\' {
					out = append(out, '\\')
				}
				out = append(out, cc)
			}
			out = append(out, '"')
			return mailAddressUTF16String(out)
		}
		if (c < 0x20 && c != '\r' && c != '\n' && c != '\t') || (c >= 0x7f && !mailAddressClass.allowUTF8) || strings.ContainsRune("()<>@,;:\\\"[].\x00", rune(c)) {
			quote = true
		}
	}
	if quote {
		return "\"" + value + "\""
	}
	return value
}

// FormatMailInternetAddresses mirrors InternetAddress.toString(Address[],used).
// A nil or empty list returns null, distinct from the empty string.
func FormatMailInternetAddresses(addresses []*MailInternetAddress, used ...int) *string {
	return formatMailInternetAddresses(addresses, false, used...)
}
func FormatMailInternetAddressesUnicode(addresses []*MailInternetAddress, used ...int) *string {
	return formatMailInternetAddresses(addresses, true, used...)
}
func formatMailInternetAddresses(addresses []*MailInternetAddress, unicode bool, used ...int) *string {
	initializeMailAddressProperties()
	if len(addresses) == 0 {
		return nil
	}
	columns := int32(0)
	if len(used) > 0 {
		columns = int32(used[0])
	}
	out := []uint16{}
	sawNonASCII := false
	for i, a := range addresses {
		if i != 0 {
			out = append(out, ',', ' ')
			columns += 2
		}
		var text string
		if unicode {
			text = a.ToUnicodeString()
			if checkMailASCII(text) != 1 {
				sawNonASCII = true
				data := []byte(mailPrintUTF8String(text))
				as := make([]uint16, len(data))
				for j, c := range data {
					as[j] = uint16(c)
				}
				text = mailAddressUTF16String(as)
			}
		} else {
			text = a.String()
		}
		folded := mailAddressUTF16(FoldMailHeader(0, text))
		first, last := -1, -1
		for j := 0; j+1 < len(folded); j++ {
			if folded[j] == '\r' && folded[j+1] == '\n' {
				if first < 0 {
					first = j
				}
				last = j
			}
		}
		length := len(folded)
		if first >= 0 {
			length = first
		}
		if columns+int32(length) > 76 {
			if len(out) > 0 && out[len(out)-1] == ' ' {
				out = out[:len(out)-1]
			}
			out = append(out, '\r', '\n', '\t')
			columns = 8
		}
		out = append(out, folded...)
		if last >= 0 {
			columns = int32(len(folded) - last - 2)
		} else {
			columns += int32(len(folded))
		}
	}
	result := mailAddressUTF16String(out)
	if sawNonASCII {
		// String.getBytes(ISO_8859_1) then String(bytes,UTF_8) restores the source
		// byte-counting trick, including charset replacement of isolated units.
		data, _ := encodeMailCharset(result, "ISO-8859-1")
		result = decodeMailUTF8(data)
	}
	return javaString(result)
}
