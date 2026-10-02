/**
 *  Licensed to the Apache Software Foundation (ASF) under one or more
 *  contributor license agreements.  See the NOTICE file distributed with
 *  this work for additional information regarding copyright ownership.
 *  The ASF licenses this file to You under the Apache License, Version 2.0
 *  (the "License"); you may not use this file except in compliance with
 *  the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Existing Geronimo MimeTypeTest and MimeTypeParameterListTest methods,
// translated after the corresponding Activation features were implemented.
// Source: apache/geronimo-specs geronimo-activation_1.1_spec/src/test/java/javax/activation.
package tlc

import "testing"

func javaActivationAssertString(t *testing.T, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}
func javaActivationMime(t *testing.T, args ...*string) *MailActivationMimeType {
	t.Helper()
	m, e := NewMailActivationMimeType(args...)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func javaActivationParameters(t *testing.T, args ...*string) *MailActivationParameterList {
	t.Helper()
	p, e := NewMailActivationParameterList(args...)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestJavaMailActivationMimeType(t *testing.T) {
	t.Run("testDefaultConstructor", func(t *testing.T) {
		m := javaActivationMime(t)
		if m.GetBaseType() != "application/*" || m.GetPrimaryType() != "application" || m.GetSubType() != "*" {
			t.Fatal(m.String())
		}
		if !m.Match(javaActivationMime(t)) || !m.Match(javaActivationMime(t, javaString("application/*"))) {
			t.Fatal("match")
		}
		if m.GetParameter(javaString("foo")) != nil {
			t.Fatal("foo")
		}
		if m.GetParameters().Size() != 0 || !m.GetParameters().IsEmpty() {
			t.Fatal("parameters")
		}
	})
	for _, name := range []string{"testMimeTypeConstructor", "testTypeConstructor"} {
		t.Run(name, func(t *testing.T) {
			args := []*string{javaString("text/plain")}
			if name == "testTypeConstructor" {
				args = []*string{javaString("text"), javaString("plain")}
			}
			m := javaActivationMime(t, args...)
			if m.GetBaseType() != "text/plain" || m.GetPrimaryType() != "text" || m.GetSubType() != "plain" || m.String() != "text/plain" {
				t.Fatal(m.String())
			}
		})
	}
	for _, tc := range []struct{ name, input, value, output string }{
		{"testConstructorWithParams", "text/plain; charset=\"iso-8859-1\"", "iso-8859-1", "text/plain; charset=iso-8859-1"},
		{"testConstructorWithQuotableParams", "text/plain; charset=\"iso(8859)\"", "iso(8859)", "text/plain; charset=\"iso(8859)\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := javaActivationMime(t, javaString(tc.input))
			if m.GetBaseType() != "text/plain" || m.GetPrimaryType() != "text" || m.GetSubType() != "plain" {
				t.Fatal(m.GetBaseType())
			}
			p := m.GetParameters()
			if p.Size() != 1 {
				t.Fatal(p.Size())
			}
			javaActivationAssertString(t, p.Get(javaString("charset")), tc.value)
			if m.String() != tc.output {
				t.Fatal(m.String())
			}
		})
	}
	t.Run("testWriteExternal", func(t *testing.T) {
		m := javaActivationMime(t, javaString("text/plain; charset=iso8859-1"))
		out := &MailActivationObjectOutput{WriteUTF: func(s string) error {
			if s != "text/plain; charset=iso8859-1" {
				t.Fatal(s)
			}
			return nil
		}, Flush: func() error { return nil }}
		if e := m.WriteExternal(out); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("testReadExternal", func(t *testing.T) {
		m := javaActivationMime(t)
		in := &MailActivationObjectInput{ReadUTF: func() (*string, error) { return javaString("text/plain; charset=iso-8859-1"), nil }}
		if e := m.ReadExternal(in); e != nil {
			t.Fatal(e)
		}
		if m.GetBaseType() != "text/plain" || m.GetPrimaryType() != "text" || m.GetSubType() != "plain" {
			t.Fatal(m.GetBaseType())
		}
		p := m.GetParameters()
		if p.Size() != 1 {
			t.Fatal(p.Size())
		}
		javaActivationAssertString(t, p.Get(javaString("charset")), "iso-8859-1")
		if m.String() != "text/plain; charset=iso-8859-1" {
			t.Fatal(m.String())
		}
	})
}
func TestJavaMailActivationMimeTypeParameterList(t *testing.T) {
	t.Run("testEmptyParameterList", func(t *testing.T) {
		p := javaActivationParameters(t)
		if p.Size() != 0 || !p.IsEmpty() {
			t.Fatal(p.Size())
		}
	})
	t.Run("testSimpleParameterList", func(t *testing.T) {
		p := javaActivationParameters(t, javaString(";name=value"))
		if p.Size() != 1 || p.IsEmpty() {
			t.Fatal(p.Size())
		}
		e := p.NamesEnumeration()
		if !e.HasMoreElements() {
			t.Fatal("empty")
		}
		name, err := e.NextElement()
		if err != nil {
			t.Fatal(err)
		}
		if name != "name" {
			t.Fatal(name)
		}
		if e.HasMoreElements() {
			t.Fatal("more")
		}
		javaActivationAssertString(t, p.Get(javaString("name")), "value")
	})
	t.Run("testQuotedValue", func(t *testing.T) {
		p := javaActivationParameters(t, javaString(";name=\"val()ue\""))
		if p.Size() != 1 {
			t.Fatal(p.Size())
		}
		javaActivationAssertString(t, p.Get(javaString("name")), "val()ue")
	})
	t.Run("testWhiteSpacesParameterList", func(t *testing.T) {
		p := javaActivationParameters(t, javaString("; name= value"))
		if p.Size() != 1 {
			t.Fatal(p.Size())
		}
		n, e := p.NamesEnumeration().NextElement()
		if e != nil {
			t.Fatal(e)
		}
		if n != "name" {
			t.Fatal(n)
		}
		javaActivationAssertString(t, p.Get(javaString("name")), "value")
	})
	t.Run("testLongParameterList", func(t *testing.T) {
		p := javaActivationParameters(t, javaString(";name1=value1; name2 = value2; name3=value3;name4  = value4"))
		if p.Size() != 4 {
			t.Fatal(p.Size())
		}
		for _, kv := range []struct{ name, value string }{{"name1", "value1"}, {"name2", "value2"}, {"name3", "value3"}, {"name4", "value4"}} {
			javaActivationAssertString(t, p.Get(javaString(kv.name)), kv.value)
		}
	})
	t.Run("testCaseInsensitivity", func(t *testing.T) {
		p := javaActivationParameters(t, javaString(";name1=value; NAME2=VALUE; NaMe3=VaLuE"))
		if p.Size() != 3 {
			t.Fatal(p.Size())
		}
		for _, kv := range []struct{ name, value string }{{"name1", "value"}, {"name2", "VALUE"}, {"name3", "VaLuE"}, {"NAME1", "value"}, {"NaMe1", "value"}} {
			javaActivationAssertString(t, p.Get(javaString(kv.name)), kv.value)
		}
		p.Remove(javaString("NAME1"))
		if p.Get(javaString("name1")) != nil {
			t.Fatal("name1")
		}
		p.Remove(javaString("name3"))
		if p.String() != "; name2=VALUE" {
			t.Fatal(p.String())
		}
	})
	for _, tc := range []struct{ name, input string }{
		{"testNoValueParameterList", "; name="}, {"testMissingValueParameterList", "; name=;name2=value"}, {"testNoNameParameterList", "; = value"}, {"testUnterminatedQuotedString", "; = \"value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, e := NewMailActivationParameterList(javaString(tc.input))
			if e == nil {
				if tc.name != "testNoNameParameterList" {
					t.Fatal("Expected MimeTypeParseException")
				}
				return
			}
			if _, ok := e.(*MailMimeTypeParseException); !ok {
				t.Fatal(e)
			}
		})
	}
	t.Run("testSpecialInAttribute", func(t *testing.T) {
		specials := "()<>@,;:\\\"/[]?= \t"
		for _, c := range specials {
			_, e := NewMailActivationParameterList(javaString(";na" + string(c) + "me=value"))
			if e == nil {
				t.Fatalf("Expected MimeTypeParseException for special: %c", c)
			}
			if _, ok := e.(*MailMimeTypeParseException); !ok {
				t.Fatal(e)
			}
		}
	})
}
