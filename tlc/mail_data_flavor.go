/*
 * Copyright (c) 1996, 2021, Oracle and/or its affiliates. All rights reserved.
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

// OpenJDK DataFlavor metadata, MIME constructors and equality paths required by
// Activation. Desktop text selection/readers and object serialization remain
// separate required features; arbitrary JVM classes use identity providers.
package tlc

import (
	"fmt"
	"sync/atomic"
)

var mailFlavorIdentityHash atomic.Int32
var mailFlavorObjectClass = &MailActivationClass{Name: "java.lang.Object"}
var mailFlavorSerializableClass = &MailActivationClass{Name: "java.io.Serializable"}
var mailFlavorInputStreamClass = &MailActivationClass{Name: "java.io.InputStream", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailFlavorStringClass = &MailActivationClass{Name: "java.lang.String", Parents: []*MailActivationClass{mailFlavorObjectClass, mailFlavorSerializableClass}}
var mailFlavorReaderClass = &MailActivationClass{Name: "java.io.Reader", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailFlavorCharBufferClass = &MailActivationClass{Name: "java.nio.CharBuffer", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailFlavorByteBufferClass = &MailActivationClass{Name: "java.nio.ByteBuffer", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailFlavorCharArrayClass = &MailActivationClass{Name: "[C", Parents: []*MailActivationClass{mailFlavorObjectClass, mailFlavorSerializableClass}}
var mailFlavorByteArrayClass = &MailActivationClass{Name: "[B", Parents: []*MailActivationClass{mailFlavorObjectClass, mailFlavorSerializableClass}}
var mailFlavorRemoteClass = &MailActivationClass{Name: "java.rmi.Remote"}
var mailFlavorListClass = &MailActivationClass{Name: "java.util.List"}

// This native inventory supplies only the classes required by flavor metadata.
// Equal names in different provider loaders retain distinct pointer identities.
func MailBuiltinFlavorClass(name string) *MailActivationClass {
	switch name {
	case "java.lang.Object":
		return mailFlavorObjectClass
	case "java.lang.String":
		return mailFlavorStringClass
	case "java.io.InputStream":
		return mailFlavorInputStreamClass
	case "java.io.Reader":
		return mailFlavorReaderClass
	case "java.nio.CharBuffer":
		return mailFlavorCharBufferClass
	case "java.nio.ByteBuffer":
		return mailFlavorByteBufferClass
	case "[C":
		return mailFlavorCharArrayClass
	case "[B":
		return mailFlavorByteArrayClass
	case "java.io.Serializable":
		return mailFlavorSerializableClass
	case "java.rmi.Remote":
		return mailFlavorRemoteClass
	case "java.util.List":
		return mailFlavorListClass
	}
	if class := mailNativeContentHandlerClass(name); class != nil {
		return class
	}
	return nil
}
func mailFlavorLoadClass(name *string) (*MailActivationClass, error) {
	if f := DefaultMailActivationEnvironment.FlavorClass; f != nil {
		return f(name)
	}
	if name == nil {
		return nil, NewNullPointerException()
	}
	if class := MailBuiltinFlavorClass(*name); class != nil {
		return class, nil
	}
	return nil, NewUnsupportedOperationException("JVM DataFlavor class-loading provider required for " + *name)
}
func (c *MailActivationClass) GetName() string {
	if c == nil {
		panic(NewNullPointerException())
	}
	if c.Name == "" {
		panic(NewUnsupportedOperationException("JVM Class name provider required"))
	}
	return c.Name
}
func (c *MailActivationClass) HashCode() int32 {
	if c == nil {
		panic(NewNullPointerException())
	}
	if c.HashCodeFunc != nil {
		return c.HashCodeFunc()
	}
	if h := c.identityHash.Load(); h != 0 {
		return h
	}
	h := mailFlavorIdentityHash.Add(1) & 0x7fffffff
	if h == 0 {
		h = 1
	}
	c.identityHash.CompareAndSwap(0, h)
	return c.identityHash.Load()
}
func (c *MailActivationClass) IsAssignableFrom(other *MailActivationClass) bool {
	if c == nil || other == nil {
		panic(NewNullPointerException())
	}
	if c.IsAssignableFromFunc != nil {
		return c.IsAssignableFromFunc(other)
	}
	if c == other {
		return true
	}
	if c == mailFlavorObjectClass && !other.Primitive {
		return true
	}
	for _, p := range other.Parents {
		if c.IsAssignableFrom(p) {
			return true
		}
	}
	return false
}

type MailDataFlavor struct {
	mime                    *MailAWTMimeType
	representation          *MailActivationClass
	human                   *string
	activation              *mailActivationFlavorFields
	MIMETypeFunc            func() *string
	RepresentationClassFunc func() *MailActivationClass
	PrimaryTypeFunc         func() *string
	SubTypeFunc             func() *string
	ParameterFunc           func(*string) *string
	EqualsFunc              func(*MailDataFlavor) bool
	ClassName               string
}

// No-argument construction retains null base fields, as required for externalization.
func NewMailDataFlavor() *MailDataFlavor { return &MailDataFlavor{} }
func NewMailMIMEDataFlavor(mime, human *string) (*MailDataFlavor, error) {
	return mailInvoke(func() (*MailDataFlavor, error) {
		if mime == nil {
			return nil, NewNullPointerException("mimeType")
		}
		m, err := NewMailAWTMimeType(mime)
		if err != nil {
			if _, ok := err.(*MailAWTMimeTypeParseException); ok {
				return nil, NewIllegalArgumentException("failed to parse:" + *mime)
			}
			return nil, err
		}
		f := &MailDataFlavor{mime: m}
		name := f.GetParameter(javaString("class"))
		if name == nil {
			if m.GetBaseType() == "application/x-java-serialized-object" {
				return nil, NewIllegalArgumentException("no representation class specified for:" + *mime)
			}
			f.representation = mailFlavorInputStreamClass
		} else {
			f.representation, err = mailFlavorLoadClass(name)
			if err != nil {
				if _, ok := err.(interface{ mailClassNotFoundException() }); ok {
					return nil, NewIllegalArgumentException("can't find specified class: " + javaNullableString(javaThrowableDetailMessage(err)))
				}
				return nil, err
			}
		}
		m.SetParameter(javaString("class"), javaString(f.representation.GetName()))
		if human == nil {
			human = m.GetParameter(javaString("humanPresentableName"))
			if human == nil {
				human = javaString(m.GetBaseType())
			}
		}
		f.human = copyJavaMessage(human)
		m.RemoveParameter(javaString("humanPresentableName"))
		return f, nil
	})
}
func NewMailClassDataFlavor(class *MailActivationClass, human *string) (*MailDataFlavor, error) {
	return mailInvoke(func() (*MailDataFlavor, error) {
		if class == nil {
			return nil, NewNullPointerException("representationClass")
		}
		params, _ := NewMailAWTMimeParameters()
		params.Set(javaString("class"), javaString(class.GetName()))
		if human == nil {
			human = params.Get(javaString("humanPresentableName"))
			if human == nil {
				human = javaString("application/x-java-serialized-object")
			}
		}
		mime, err := NewMailAWTMimeTypeWithParameters(javaString("application"), javaString("x-java-serialized-object"), params)
		if err != nil {
			return nil, NewIllegalArgumentException("MimeType Parse Exception: " + javaNullableString(javaThrowableDetailMessage(err)))
		}
		mime.RemoveParameter(javaString("humanPresentableName"))
		return &MailDataFlavor{mime: mime, representation: class, human: copyJavaMessage(human)}, nil
	})
}
func (f *MailDataFlavor) GetMIMEType() *string {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.MIMETypeFunc != nil {
		return f.MIMETypeFunc()
	}
	if f.activation != nil {
		return copyJavaMessage(f.activation.mime)
	}
	if f.mime == nil {
		return nil
	}
	return javaString(f.mime.String())
}
func (f *MailDataFlavor) GetRepresentationClass() *MailActivationClass {
	if f == nil {
		panic(NewNullPointerException())
	}
	if f.RepresentationClassFunc != nil {
		return f.RepresentationClassFunc()
	}
	if f.activation != nil {
		return f.activation.representation
	}
	return f.representation
}
func (f *MailDataFlavor) GetHumanPresentableName() *string {
	if f == nil {
		panic(NewNullPointerException())
	}

	if f.activation != nil {
		return copyJavaMessage(f.activation.human)
	}
	return copyJavaMessage(f.human)
}
func (f *MailDataFlavor) SetHumanPresentableName(name *string) {
	if f == nil {
		panic(NewNullPointerException())
	}

	if f.activation != nil {
		f.activation.human = copyJavaMessage(name)
	} else {
		f.human = copyJavaMessage(name)
	}
}
func (f *MailDataFlavor) GetPrimaryType() *string {
	if f == nil {
		panic(NewNullPointerException())
	}

	if f.PrimaryTypeFunc != nil {
		return f.PrimaryTypeFunc()
	}
	if f.mime == nil {
		return nil
	}
	return f.mime.GetPrimaryType()
}
func (f *MailDataFlavor) GetSubType() *string {
	if f == nil {
		panic(NewNullPointerException())
	}

	if f.SubTypeFunc != nil {
		return f.SubTypeFunc()
	}
	if f.mime == nil {
		return nil
	}
	return f.mime.GetSubType()
}
func (f *MailDataFlavor) GetParameter(name *string) *string {
	if f == nil {
		panic(NewNullPointerException())
	}

	if f.ParameterFunc != nil {
		return f.ParameterFunc(name)
	}
	if name == nil {
		panic(NewNullPointerException())
	}
	// This inherited method reads the base human field, not the virtual getter.
	if *name == "humanPresentableName" {
		return copyJavaMessage(f.human)
	}
	if f.mime == nil {
		return nil
	}
	return f.mime.GetParameter(name)
}
func (f *MailDataFlavor) Equals(other *MailDataFlavor) bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	if f.EqualsFunc != nil {
		return f.EqualsFunc(other)
	}
	if f.activation != nil {
		return f.activationEquals(other)
	}
	if other == nil {
		return false
	}
	if f == other {
		return true
	}
	if f.GetRepresentationClass() != other.GetRepresentationClass() {
		return false
	}
	if f.mime == nil {
		return other.mime == nil
	}
	if !f.mime.Match(other.mime) {
		return false
	}
	if mailNullableStringsEqual(f.GetPrimaryType(), javaString("text")) {
		if mailFlavorSubtypeCharset(f) && f.representation != nil && !f.standardTextRepresentation() {
			if !mailNullableStringsEqual(mailFlavorCanonicalCharset(f.GetParameter(javaString("charset"))), mailFlavorCanonicalCharset(other.GetParameter(javaString("charset")))) {
				return false
			}
		}
		if mailNullableStringsEqual(f.GetSubType(), javaString("html")) && !mailNullableStringsEqual(f.GetParameter(javaString("document")), other.GetParameter(javaString("document"))) {
			return false
		}
	}
	return true
}
func (f *MailDataFlavor) AsMailDataFlavor() *MailDataFlavor { return f }
func (f *MailDataFlavor) EqualsObject(value any) bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	other, ok := value.(interface{ AsMailDataFlavor() *MailDataFlavor })
	if !ok {
		return false
	}
	flavor := other.AsMailDataFlavor()
	return flavor != nil && f.Equals(flavor)
}
func (f *MailDataFlavor) EqualsString(value *string) bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return value != nil && f.mime != nil && f.IsMIMETypeEqual(value)
}
func (f *MailDataFlavor) Match(other *MailDataFlavor) bool {
	if f == nil {
		panic(NewNullPointerException())
	}
	return f.Equals(other)
}
func (f *MailDataFlavor) IsMIMETypeEqual(value *string) bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	if f.activation != nil {
		return f.activationMIMEEqual(value)
	}
	if value == nil {
		panic(NewNullPointerException("mimeType"))
	}
	if f.mime == nil {
		return false
	}
	other, err := NewMailAWTMimeType(value)
	if err != nil {
		if _, ok := err.(*MailAWTMimeTypeParseException); ok {
			return false
		}
		panic(err)
	}
	return f.mime.Match(other)
}
func (f *MailDataFlavor) IsMIMETypeEqualFlavor(other *MailDataFlavor) bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	if other == nil {
		panic(NewNullPointerException())
	}
	if f.mime == nil {
		return other.mime == nil
	}
	return f.mime.Match(other.mime)
}
func (f *MailDataFlavor) IsMIMETypeSerializedObject() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return f.IsMIMETypeEqual(javaString("application/x-java-serialized-object"))
}
func (f *MailDataFlavor) GetDefaultRepresentationClass() *MailActivationClass {
	if f == nil {
		panic(NewNullPointerException())
	}

	return mailFlavorInputStreamClass
}
func (f *MailDataFlavor) GetDefaultRepresentationClassAsString() string {
	if f == nil {
		panic(NewNullPointerException())
	}

	return f.GetDefaultRepresentationClass().GetName()
}
func (f *MailDataFlavor) IsRepresentationClassInputStream() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return mailFlavorInputStreamClass.IsAssignableFrom(f.representation)
}
func (f *MailDataFlavor) IsRepresentationClassReader() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return mailFlavorReaderClass.IsAssignableFrom(f.representation)
}
func (f *MailDataFlavor) IsRepresentationClassCharBuffer() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return mailFlavorCharBufferClass.IsAssignableFrom(f.representation)
}
func (f *MailDataFlavor) IsRepresentationClassByteBuffer() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return mailFlavorByteBufferClass.IsAssignableFrom(f.representation)
}
func (f *MailDataFlavor) IsRepresentationClassSerializable() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return mailFlavorSerializableClass.IsAssignableFrom(f.representation)
}
func (f *MailDataFlavor) IsRepresentationClassRemote() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return mailFlavorRemoteClass.IsAssignableFrom(f.representation)
}
func (f *MailDataFlavor) IsFlavorSerializedObjectType() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return f.IsRepresentationClassSerializable() && f.IsMIMETypeSerializedObject()
}
func (f *MailDataFlavor) IsFlavorRemoteObjectType() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return f.IsRepresentationClassRemote() && f.IsRepresentationClassSerializable() && f.IsMIMETypeEqual(javaString("application/x-java-remote-object"))
}
func (f *MailDataFlavor) IsFlavorJavaFileListType() bool {
	if f == nil {
		panic(NewNullPointerException())
	}

	return f.mime != nil && f.representation != nil && mailFlavorListClass.IsAssignableFrom(f.representation) && f.mime.Match(mailFlavorFileListMIME())
}
func mailFlavorFileListMIME() *MailAWTMimeType {
	m, err := NewMailAWTMimeType(javaString("application/x-java-file-list"))
	if err != nil {
		panic(err)
	}
	return m
}
func (f *MailDataFlavor) standardTextRepresentation() bool {
	return f.IsRepresentationClassReader() || f.representation == mailFlavorStringClass || f.IsRepresentationClassCharBuffer() || f.representation == mailFlavorCharArrayClass
}
func (f *MailDataFlavor) HashCode() int32 {
	if f == nil {
		panic(NewNullPointerException())
	}

	var total int32
	if f.representation != nil {
		total += f.representation.HashCode()
	}
	if f.mime != nil {
		primary := f.mime.GetPrimaryType()
		if primary != nil {
			total += mailAWTStringHash(*primary)
		}
		if mailNullableStringsEqual(primary, javaString("text")) {
			if mailFlavorSubtypeCharset(f) && f.representation != nil && !f.standardTextRepresentation() {
				charset := mailFlavorCanonicalCharset(f.GetParameter(javaString("charset")))
				if charset != nil {
					total += mailAWTStringHash(*charset)
				}
			}
			if mailNullableStringsEqual(f.GetSubType(), javaString("html")) {
				doc := f.GetParameter(javaString("document"))
				if doc != nil {
					total += mailAWTStringHash(*doc)
				}
			}
		}
	}
	return total
}
func (f *MailDataFlavor) Clone() *MailDataFlavor {
	if f == nil {
		panic(NewNullPointerException())
	}

	clone := *f
	if f.mime != nil {
		clone.mime = f.mime.Clone()
	}
	if f.activation != nil {
		x := *f.activation
		clone.activation = &x
	}
	return &clone
}
func (f *MailDataFlavor) String() string {
	if f == nil {
		panic(NewNullPointerException())
	}

	class := "java.awt.datatransfer.DataFlavor"
	if f.activation != nil {
		class = "javax.activation.ActivationDataFlavor"
	}
	mime, repr := "null", "null"
	if f.mime != nil {
		mime = f.mime.GetBaseType()
	}
	if f.representation != nil {
		repr = f.representation.GetName()
	}
	params := fmt.Sprintf("mimetype=%s;representationclass=%s", mime, repr)
	if mailFlavorCharsetText(f) && (f.IsRepresentationClassInputStream() || f.IsRepresentationClassByteBuffer() || f.representation == mailFlavorByteArrayClass) {
		charset := f.GetParameter(javaString("charset"))
		if charset == nil {
			charset = javaString(mailFlavorDefaultCharset())
		}
		params += ";charset=" + *charset
	}
	if f.ClassName != "" {
		class = f.ClassName
	}
	return class + "[" + params + "]"
}

func (f *MailDataFlavor) IsFlavorTextType() bool {
	if f == nil {
		panic(NewNullPointerException())
	}
	return mailFlavorCharsetText(f) || mailFlavorNoncharsetText(f)
}
