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

// Existing Geronimo AbstractHandler and TextPlain/Html/XmlTest methods,
// translated after the bundled text handler implementation. The current upstream
// package is org.apache.geronimo.javamail.handlers; these unchanged assertions
// also pass against the bundled org.apache.geronimo.activation.handlers classes.
package tlc

import (
	"bytes"
	"io"
	"testing"
)

func TestJavaMailGeronimoTextHandlers(t *testing.T) {
	for _, tc := range []struct {
		name, mime, human string
		constructor       func() *MailDataContentHandler
	}{
		{"TextPlainTest", "text/plain", "Plain Text", NewMailTextPlainHandler},
		{"TextHtmlTest", "text/html", "HTML Text", NewMailTextHTMLHandler},
		{"TextXmlTest", "text/xml", "XML Text", NewMailTextXMLHandler},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("testDataFlavor", func(t *testing.T) {
				dch := tc.constructor()
				flavors := dch.GetTransferDataFlavors()
				if len(flavors) != 1 {
					t.Fatal(len(flavors))
				}
				flavor := flavors[0].(*MailDataFlavor)
				if flavor.GetRepresentationClass() != mailFlavorStringClass {
					t.Fatal("representation class")
				}
				javaActivationAssertString(t, flavor.GetMIMEType(), tc.mime)
				javaActivationAssertString(t, flavor.GetHumanPresentableName(), tc.human)
			})
			t.Run("testGetContent", func(t *testing.T) {
				dch := tc.constructor()
				data := []byte("Hello World")
				ds := &MailDataSource{
					InputFunc:       func() (*MailInputStream, error) { return NewMailInputStream(bytes.NewReader(data), nil), nil },
					OutputFunc:      func() (io.WriteCloser, error) { return nil, NewUnsupportedOperationException() },
					ContentTypeFunc: func() *string { return javaString(tc.mime) },
					NameFunc:        func() *string { panic(NewUnsupportedOperationException()) },
				}
				content, err := dch.GetContent(ds)
				if err != nil {
					t.Fatal(err)
				}
				if value, ok := content.(string); !ok || value != "Hello World" {
					t.Fatalf("got %#v, want Hello World", content)
				}
			})
			t.Run("testWriteTo", func(t *testing.T) {
				dch := tc.constructor()
				var out bytes.Buffer
				if err := dch.WriteTo("Hello World", javaString(tc.mime), &out); err != nil {
					t.Fatal(err)
				}
				if out.String() != "Hello World" {
					t.Fatal(out.String())
				}
			})
		})
	}
}
