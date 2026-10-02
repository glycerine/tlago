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
// Go port of JavaMail 1.6.8 ByteArrayDataSource. Source:
// https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/util/ByteArrayDataSource.java
package tlc

import (
	"io"
	"os"
	"sync"
)

// MailDataSource is a concrete native DataSource. Callback fields preserve
// custom/subclass providers; byte/file constructors install their own storage.
type MailDataSource struct {
	NameValue, TypeValue                    *string
	NameFunc, ContentTypeFunc, EncodingFunc func() *string
	InputFunc                               func() (*MailInputStream, error)
	OutputFunc                              func() (io.WriteCloser, error)
	mu                                      sync.Mutex
	data                                    []byte
	length                                  int
	file                                    *TLAFile
	isFile                                  bool
	fileTypeMap                             *MailMimetypesFileTypeMap
}

func NewMailByteArrayDataSource(data []byte, contentType *string) *MailDataSource {
	return &MailDataSource{NameValue: javaString(""), TypeValue: copyJavaMessage(contentType), data: data, length: -1}
}
func NewMailStringDataSource(data, contentType *string) (*MailDataSource, error) {
	return mailInvoke(func() (*MailDataSource, error) {
		var charset *string
		ct, e := NewMailContentType(contentType)
		if e != nil {
			if _, ok := e.(*MailParseException); !ok {
				return nil, e
			}
		} else {
			charset = ct.GetParameter("charset")
		}
		cs := DefaultMailMIMECodec.defaultJavaCharset
		name := ""
		if charset == nil {
			name = cs()
		} else {
			name = DefaultMailMIMECodec.JavaCharset(*charset)
		}
		if data == nil {
			return nil, NewNullPointerException()
		}
		encoder := DefaultMailMIMECodec.EncodeCharset
		if encoder == nil {
			encoder = encodeMailCharset
		}
		b, e := encoder(*data, name)
		if e != nil {
			return nil, e
		}
		return NewMailByteArrayDataSource(b, contentType), nil
	})
}
func NewMailReaderDataSource(stream *MailInputStream, contentType *string) (*MailDataSource, error) {
	return mailInvoke(func() (*MailDataSource, error) {
		data := make([]byte, 32)
		count := 0
		buffer := make([]byte, 8192)
		for {
			n, e := stream.ReadJava(buffer, 0, 8192)
			if e != nil {
				return nil, e
			}
			if n <= 0 {
				break
			}
			if n > len(buffer) {
				return nil, NewIndexOutOfBoundsException(n, len(buffer))
			}
			need := count + n
			if need > len(data) {
				capacity := len(data) * 2
				if need > capacity {
					capacity = need
				}
				larger := make([]byte, capacity)
				copy(larger, data[:count])
				data = larger
			}
			copy(data[count:need], buffer[:n])
			count = need
		}
		if len(data)-count > 256*1024 {
			data = append([]byte{}, data[:count]...)
		}
		return &MailDataSource{NameValue: javaString(""), TypeValue: copyJavaMessage(contentType), data: data, length: count}, nil
	})
}
func (ds *MailDataSource) GetName() *string {
	if ds == nil {
		panic(NewNullPointerException())
	}
	if ds.NameFunc != nil {
		return ds.NameFunc()
	}
	if ds.isFile {
		if ds.file == nil {
			panic(NewNullPointerException())
		}
		return javaString(ds.file.GetName())
	}
	return copyJavaMessage(ds.NameValue)
}
func (ds *MailDataSource) SetName(name *string) {
	if ds == nil {
		panic(NewNullPointerException())
	}
	ds.NameValue = copyJavaMessage(name)
}
func (ds *MailDataSource) GetContentType() *string {
	if ds == nil {
		panic(NewNullPointerException())
	}
	if ds.ContentTypeFunc != nil {
		return ds.ContentTypeFunc()
	}
	if ds.isFile {
		m := ds.fileTypeMap
		if m == nil {
			m = GetDefaultMailFileTypeMap()
		}
		return javaString(m.GetFileContentType(ds.file))
	}
	return copyJavaMessage(ds.TypeValue)
}
func (ds *MailDataSource) GetInputStream() (*MailInputStream, error) {
	if ds == nil {
		return nil, NewNullPointerException()
	}
	if ds.InputFunc != nil {
		return ds.InputFunc()
	}
	if ds.isFile {
		if ds.file == nil {
			return nil, NewNullPointerException()
		}
		return mailOpenFileInputStream(ds.file.GetPath())
	}
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if ds.data == nil {
		return nil, NewIOException("no data")
	}
	if ds.length < 0 {
		ds.length = len(ds.data)
	}
	return NewMailSharedByteArrayInputStream(ds.data, 0, int32(ds.length)).InputStream(), nil
}
func (ds *MailDataSource) GetOutputStream() (io.WriteCloser, error) {
	if ds == nil {
		return nil, NewNullPointerException()
	}
	if ds.OutputFunc != nil {
		return ds.OutputFunc()
	}
	if ds.isFile {
		if ds.file == nil {
			return nil, NewNullPointerException()
		}
		f, e := os.Create(ds.file.GetPath())
		if e != nil {
			return nil, distributedFileOpenException(ds.file.GetPath(), e)
		}
		return &mailFileStream{filenameOSStream: filenameOSStream{File: f}}, nil
	}
	return nil, NewIOException("cannot do this")
}
