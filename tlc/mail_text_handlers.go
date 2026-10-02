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

// Bundled Geronimo AbstractTextHandler and its plain/HTML/XML subclasses.
// The actual bundled bytecode uses the default charset and never closes streams;
// the newer org.apache.geronimo.javamail handlers have different algorithms.
package tlc

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Foreign Go objects are mapped to Java null before virtual dispatch.
func mailHandlerNull(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	}
	return false
}
func mailHandlerObjectString(value any) *string {
	switch v := value.(type) {
	case string:
		return javaString(v)
	case *string:
		return copyJavaMessage(v)
	}
	if cb := DefaultMailActivationEnvironment.ObjectString; cb != nil {
		return cb(value)
	}
	switch v := value.(type) {
	case interface{ JavaToString() *string }:
		return v.JavaToString()
	case fmt.Stringer:
		return javaString(v.String())
	case bool:
		return javaString(strconv.FormatBool(v))
	case int:
		return javaString(strconv.FormatInt(int64(v), 10))
	case int8:
		return javaString(strconv.FormatInt(int64(v), 10))
	case int16:
		return javaString(strconv.FormatInt(int64(v), 10))
	case int32:
		return javaString(strconv.FormatInt(int64(v), 10))
	case int64:
		return javaString(strconv.FormatInt(v, 10))
	}
	panic(NewUnsupportedOperationException("Java Object.toString provider required"))
}

func NewMailAbstractTextHandler(flavor *MailDataFlavor) *MailDataContentHandler {
	h := &MailDataContentHandler{}
	h.TransferDataFlavorsFunc = func() []any {
		if flavor == nil {
			return []any{nil}
		}
		return []any{flavor}
	}
	h.TransferDataFunc = func(value any, source *MailDataSource) (any, error) {
		var other *MailDataFlavor
		if !mailHandlerNull(value) {
			c, ok := value.(interface{ AsMailDataFlavor() *MailDataFlavor })
			if !ok {
				return nil, NewClassCastException()
			}
			other = c.AsMailDataFlavor()
		}
		if flavor.Equals(other) {
			return h.GetContent(source)
		}
		return nil, nil
	}
	h.ContentFunc = func(source *MailDataSource) (any, error) {
		input, err := source.GetInputStream()
		if err != nil {
			return nil, err
		}
		var content *string
		if cb := DefaultMailActivationEnvironment.ReadText; cb != nil {
			content, err = cb(input)
		} else {
			content, err = mailReadDefaultText(input)
		}
		if err != nil {
			return nil, err
		}
		if content == nil {
			return nil, nil
		}
		return *content, nil
	}
	h.WriteToFunc = func(value any, _ *string, out io.Writer) error {
		if mailHandlerNull(value) {
			return nil
		}
		text := mailHandlerObjectString(value)
		var writer *MailTextWriter
		var err error
		if cb := DefaultMailActivationEnvironment.TextWriter; cb != nil {
			writer, err = cb(out)
		} else {
			writer, err = NewMailDefaultTextWriter(out)
		}
		if err != nil {
			return err
		}
		if err = writer.WriteString(text); err != nil {
			return err
		}
		return writer.Flush()
	}
	return h
}
func NewMailTextPlainHandler() *MailDataContentHandler {
	return NewMailAbstractTextHandler(NewMailActivationTypedDataFlavor(mailFlavorStringClass, javaString("text/plain"), javaString("Plain Text")))
}
func NewMailTextHTMLHandler() *MailDataContentHandler {
	return NewMailAbstractTextHandler(NewMailActivationTypedDataFlavor(mailFlavorStringClass, javaString("text/html"), javaString("HTML Text")))
}
func NewMailTextXMLHandler() *MailDataContentHandler {
	return NewMailAbstractTextHandler(NewMailActivationTypedDataFlavor(mailFlavorStringClass, javaString("text/xml"), javaString("XML Text")))
}

var mailAbstractTextHandlerClass = &MailActivationClass{Name: "org.apache.geronimo.activation.handlers.AbstractTextHandler", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailTextPlainHandlerClass = &MailActivationClass{Name: "org.apache.geronimo.activation.handlers.TextPlainHandler", Parents: []*MailActivationClass{mailAbstractTextHandlerClass}, NewInstance: func() (any, error) { return NewMailTextPlainHandler(), nil }}
var mailTextHTMLHandlerClass = &MailActivationClass{Name: "org.apache.geronimo.activation.handlers.TextHtmlHandler", Parents: []*MailActivationClass{mailAbstractTextHandlerClass}, NewInstance: func() (any, error) { return NewMailTextHTMLHandler(), nil }}
var mailTextXMLHandlerClass = &MailActivationClass{Name: "org.apache.geronimo.activation.handlers.TextXmlHandler", Parents: []*MailActivationClass{mailAbstractTextHandlerClass}, NewInstance: func() (any, error) { return NewMailTextXMLHandler(), nil }}

func mailNativeContentHandlerClass(name string) *MailActivationClass {
	switch name {
	case mailTextPlainHandlerClass.Name:
		return mailTextPlainHandlerClass
	case mailTextHTMLHandlerClass.Name:
		return mailTextHTMLHandlerClass
	case mailTextXMLHandlerClass.Name:
		return mailTextXMLHandlerClass
	case mailImageGIFHandlerClass.Name:
		return mailImageGIFHandlerClass
	case mailImageJPEGHandlerClass.Name:
		return mailImageJPEGHandlerClass
	}
	return nil
}
