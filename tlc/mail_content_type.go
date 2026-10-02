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

// Go port of JavaMail 1.6.8 ContentType.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/ContentType.java

package tlc

import "strings"

type MailContentType struct {
	PrimaryType, SubType *string
	Parameters           *MailParameterList
}

func NewMailContentType(value ...*string) (*MailContentType, error) {
	c := &MailContentType{}
	if len(value) == 0 {
		return c, nil
	}
	s := value[0]
	h := NewMailHeaderTokenizer(s, MailHeaderMIME)
	bad := func(expected string, token *MailHeaderToken) error {
		return NewMailParseException("In Content-Type string <" + javaNullableString(s) + ">, expected " + expected + ", got " + javaNullableString(token.Value))
	}
	token, err := h.Next()
	if err != nil {
		return nil, err
	}
	if token.Type != MailHeaderAtom {
		return nil, bad("MIME type", token)
	}
	c.PrimaryType = token.Value
	token, err = h.Next()
	if err != nil {
		return nil, err
	}
	if uint16(token.Type) != '/' {
		return nil, bad("'/'", token)
	}
	token, err = h.Next()
	if err != nil {
		return nil, err
	}
	if token.Type != MailHeaderAtom {
		return nil, bad("MIME subtype", token)
	}
	c.SubType = token.Value
	if rest := h.GetRemainder(); rest != nil {
		c.Parameters, err = NewMailParameterList(rest)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
func NewMailContentTypeParts(primary, subtype *string, parameters *MailParameterList) *MailContentType {
	return &MailContentType{copyJavaMessage(primary), copyJavaMessage(subtype), parameters}
}
func (c *MailContentType) GetBaseType() string {
	if c.PrimaryType == nil || c.SubType == nil {
		return ""
	}
	return *c.PrimaryType + "/" + *c.SubType
}
func (c *MailContentType) GetParameter(name string) *string {
	if c.Parameters == nil {
		return nil
	}
	return c.Parameters.Get(name)
}
func (c *MailContentType) SetParameter(name string, value *string) error {
	if c.Parameters == nil {
		c.Parameters, _ = NewMailParameterList()
	}
	return c.Parameters.Set(name, value)
}
func (c *MailContentType) String() string {
	if c.PrimaryType == nil || c.SubType == nil {
		return ""
	}
	base := c.GetBaseType()
	if c.Parameters != nil {
		base += c.Parameters.ToString(len(mailAddressUTF16(base)) + 14)
	}
	return base
}
func (c *MailContentType) Match(other *MailContentType) bool {
	if other == nil {
		panic(NewNullPointerException())
	}
	same := func(a, b *string) bool {
		return a == nil && b == nil || a != nil && b != nil && mailAddressEqualsIgnoreCase(*a, *b)
	}
	if !same(c.PrimaryType, other.PrimaryType) {
		return false
	}
	if c.SubType != nil && strings.HasPrefix(*c.SubType, "*") || other.SubType != nil && strings.HasPrefix(*other.SubType, "*") {
		return true
	}
	return same(c.SubType, other.SubType)
}
func (c *MailContentType) MatchString(value *string) bool {
	other, err := NewMailContentType(value)
	if err != nil {
		if _, ok := err.(*MailParseException); ok {
			return false
		}
		panic(err)
	}
	return c.Match(other)
}
