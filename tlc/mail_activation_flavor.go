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

// Bundled Geronimo ActivationDataFlavor. Its MIME/class/name fields shadow
// AWT DataFlavor's fields; inherited methods must retain the uninitialized base.
package tlc

type mailActivationFlavorFields struct {
	representation *MailActivationClass
	mime, human    *string
}

func NewMailActivationDataFlavor(mime, human *string) *MailDataFlavor {
	return NewMailActivationTypedDataFlavor(mailFlavorInputStreamClass, mime, human)
}
func NewMailActivationClassDataFlavor(class *MailActivationClass, human *string) *MailDataFlavor {
	return NewMailActivationTypedDataFlavor(class, javaString("application/x-java-serialized-object"), human)
}
func NewMailActivationTypedDataFlavor(class *MailActivationClass, mime, human *string) *MailDataFlavor {
	return &MailDataFlavor{activation: &mailActivationFlavorFields{representation: class, mime: copyJavaMessage(mime), human: copyJavaMessage(human)}}
}
func (f *MailDataFlavor) activationEquals(other *MailDataFlavor) bool {
	// The argument getter is evaluated before parsing this object's MIME type.
	return f.IsMIMETypeEqual(other.GetMIMEType()) && f.activation.representation == other.GetRepresentationClass()
}
func (f *MailDataFlavor) activationMIMEEqual(value *string) bool {
	left, err := NewMailActivationMimeType(f.activation.mime)
	if err != nil {
		if _, ok := err.(*MailMimeTypeParseException); ok {
			return false
		}
		panic(err)
	}
	right, err := NewMailActivationMimeType(value)
	if err != nil {
		if _, ok := err.(*MailMimeTypeParseException); ok {
			return false
		}
		panic(err)
	}
	return left.Match(right)
}
func (f *MailDataFlavor) NormalizeMIMETypeParameter(name, value *string) *string {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.activation == nil {
		return copyJavaMessage(value)
	}
	return javaString(javaNullableString(name) + "=" + javaNullableString(value))
}
func (f *MailDataFlavor) NormalizeMIMEType(mime *string) *string {
	if f == nil {
		panic(NewNullPointerException())
	}
	return copyJavaMessage(mime)
}
