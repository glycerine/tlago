/*
 * Copyright (c) 2000, 2010, Oracle and/or its affiliates. All rights reserved.
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

// OpenJDK's six concrete standard stream SPIs. The backing ImageIO stream
// constructors remain explicit operations until their native implementations
// are ported; missing operations fail before the source catch(Exception).
package tlc

import "io"

// An explicit Java OutputStream carrier; arbitrary Go writers do not acquire
// Java class identity through structural io.Writer compatibility.
type MailOutputStream struct{ Writer io.Writer }

func (s *MailOutputStream) AsMailOutputStream() *MailOutputStream { return s }
func (s *MailInputStream) AsMailInputStream() *MailInputStream    { return s }
func (f *TLAFile) AsTLAFile() *TLAFile                            { return f }
func (f *RandomAccessFile) AsRandomAccessFile() *RandomAccessFile { return f }

var mailImageFileClass = &MailActivationClass{Name: "java.io.File", Parents: []*MailActivationClass{mailFlavorObjectClass, mailFlavorSerializableClass}}
var mailImageRAFClass = &MailActivationClass{Name: "java.io.RandomAccessFile", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailImageJavaOutputStreamClass = &MailActivationClass{Name: "java.io.OutputStream", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailStandardImageStreamClasses = mailImageStreamClasses()

func mailImageStreamClasses() []*MailActivationClass {
	names := []string{"FileImageInputStreamSpi", "InputStreamImageInputStreamSpi", "RAFImageInputStreamSpi", "FileImageOutputStreamSpi", "OutputStreamImageOutputStreamSpi", "RAFImageOutputStreamSpi"}
	classes := make([]*MailActivationClass, len(names))
	for i, name := range names {
		parent := mailImageInputStreamSPIClass
		if i >= 3 {
			parent = mailImageOutputStreamSPIClass
		}
		index := i
		class := &MailActivationClass{Name: "com.sun.imageio.spi." + name, Parents: []*MailActivationClass{parent}}
		class.NewInstance = func() (any, error) { return mailNewImageStreamSPI(index, class), nil }
		classes[i] = class
	}
	return classes
}

// These operations construct the actual FileImage/MemoryCacheImage/FileCacheImage
// stream subclasses, including their resources, disposer and error semantics.
type MailImageStreamConstructorEnvironment struct {
	FileInput         func(*TLAFile) (*MailImageInputStream, error)
	RAFInput          func(*RandomAccessFile) (*MailImageInputStream, error)
	FileCacheInput    func(*MailInputStream, *TLAFile) (*MailImageInputStream, error)
	MemoryCacheInput  func(*MailInputStream) (*MailImageInputStream, error)
	FileOutput        func(*TLAFile) (*MailImageOutputStream, error)
	RAFOutput         func(*RandomAccessFile) (*MailImageOutputStream, error)
	FileCacheOutput   func(*MailOutputStream, *TLAFile) (*MailImageOutputStream, error)
	MemoryCacheOutput func(*MailOutputStream) (*MailImageOutputStream, error)
}

var DefaultMailImageStreamConstructorEnvironment MailImageStreamConstructorEnvironment

func NewMailFileImageInputStreamSPI() *MailImageSPI          { return mailNewStandardImageStreamSPI(0) }
func NewMailInputStreamImageInputStreamSPI() *MailImageSPI   { return mailNewStandardImageStreamSPI(1) }
func NewMailRAFImageInputStreamSPI() *MailImageSPI           { return mailNewStandardImageStreamSPI(2) }
func NewMailFileImageOutputStreamSPI() *MailImageSPI         { return mailNewStandardImageStreamSPI(3) }
func NewMailOutputStreamImageOutputStreamSPI() *MailImageSPI { return mailNewStandardImageStreamSPI(4) }
func NewMailRAFImageOutputStreamSPI() *MailImageSPI          { return mailNewStandardImageStreamSPI(5) }
func mailNewStandardImageStreamSPI(index int) *MailImageSPI {
	return mailNewImageStreamSPI(index, mailStandardImageStreamClasses[index])
}
func mailNewImageStreamSPI(index int, class *MailActivationClass) *MailImageSPI {
	inputClass := mailImageFileClass
	if index%3 == 1 {
		inputClass = mailFlavorInputStreamClass
		if index >= 3 {
			inputClass = mailImageJavaOutputStreamClass
		}
	}
	if index%3 == 2 {
		inputClass = mailImageRAFClass
	}
	var p *MailImageSPI
	var err error
	if index < 3 {
		p, err = NewMailImageInputStreamSPI(class, javaString("Oracle Corporation"), javaString("1.0"), inputClass)
	} else {
		p, err = NewMailImageOutputStreamSPI(class, javaString("Oracle Corporation"), javaString("1.0"), inputClass)
	}
	if err != nil {
		panic(err)
	}
	descriptions := []string{
		"Service provider that instantiates a FileImageInputStream from a File",
		"Service provider that instantiates a FileCacheImageInputStream or MemoryCacheImageInputStream from an InputStream",
		"Service provider that instantiates a FileImageInputStream from a RandomAccessFile",
		"Service provider that instantiates a FileImageOutputStream from a File",
		"Service provider that instantiates an OutputStreamImageOutputStream from an OutputStream",
		"Service provider that instantiates a FileImageOutputStream from a RandomAccessFile",
	}
	p.Overrides.Description = func(any) (*string, error) { return javaString(descriptions[index]), nil }
	if index%3 == 1 {
		p.Overrides.CanUseCacheFile = func() (bool, error) { return true, nil }
		p.Overrides.NeedsCacheFile = func() (bool, error) { return false, nil }
	}
	if index < 3 {
		p.Overrides.CreateInputStream = func(input any, cache bool, dir *TLAFile) (*MailImageInputStream, error) {
			return mailStandardImageInput(index, input, cache, dir)
		}
	} else {
		p.Overrides.CreateOutputStream = func(output any, cache bool, dir *TLAFile) (*MailImageOutputStream, error) {
			return mailStandardImageOutput(index-3, output, cache, dir)
		}
	}
	return p
}
func mailImageFile(value any) *TLAFile {
	if mailHandlerNull(value) {
		return nil
	}
	if f, ok := value.(interface{ AsTLAFile() *TLAFile }); ok {
		return f.AsTLAFile()
	}
	return nil
}
func mailImageRAF(value any) *RandomAccessFile {
	if mailHandlerNull(value) {
		return nil
	}
	if f, ok := value.(interface{ AsRandomAccessFile() *RandomAccessFile }); ok {
		return f.AsRandomAccessFile()
	}
	return nil
}
func mailStandardImageInput(kind int, input any, cache bool, dir *TLAFile) (*MailImageInputStream, error) {
	env := DefaultMailImageStreamConstructorEnvironment
	var construct func() (*MailImageInputStream, error)
	switch kind {
	case 0:
		f := mailImageFile(input)
		if f == nil {
			return nil, NewIllegalArgumentException()
		}
		if env.FileInput != nil {
			construct = func() (*MailImageInputStream, error) { return env.FileInput(f) }
		}
	case 1:
		var stream *MailInputStream
		if !mailHandlerNull(input) {
			if v, ok := input.(interface{ AsMailInputStream() *MailInputStream }); ok {
				stream = v.AsMailInputStream()
			}
		}
		if stream == nil {
			return nil, NewIllegalArgumentException()
		}
		if cache {
			if env.FileCacheInput != nil {
				construct = func() (*MailImageInputStream, error) { return env.FileCacheInput(stream, dir) }
			}
		} else if env.MemoryCacheInput != nil {
			construct = func() (*MailImageInputStream, error) { return env.MemoryCacheInput(stream) }
		}
	case 2:
		f := mailImageRAF(input)
		if f == nil {
			return nil, NewIllegalArgumentException("input not a RandomAccessFile!")
		}
		if env.RAFInput != nil {
			construct = func() (*MailImageInputStream, error) { return env.RAFInput(f) }
		}
	}
	if construct == nil {
		return nil, NewUnsupportedOperationException("ImageIO backing input-stream constructor required")
	}
	result, err := mailInvoke(construct)
	if kind != 1 && err != nil && !isJavaError(err) {
		return nil, nil
	}
	return result, err
}
func mailStandardImageOutput(kind int, output any, cache bool, dir *TLAFile) (*MailImageOutputStream, error) {
	env := DefaultMailImageStreamConstructorEnvironment
	var construct func() (*MailImageOutputStream, error)
	switch kind {
	case 0:
		f := mailImageFile(output)
		if f == nil {
			return nil, NewIllegalArgumentException()
		}
		if env.FileOutput != nil {
			construct = func() (*MailImageOutputStream, error) { return env.FileOutput(f) }
		}
	case 1:
		var stream *MailOutputStream
		if !mailHandlerNull(output) {
			if v, ok := output.(interface{ AsMailOutputStream() *MailOutputStream }); ok {
				stream = v.AsMailOutputStream()
			}
		}
		if stream == nil {
			return nil, NewIllegalArgumentException()
		}
		if cache {
			if env.FileCacheOutput != nil {
				construct = func() (*MailImageOutputStream, error) { return env.FileCacheOutput(stream, dir) }
			}
		} else if env.MemoryCacheOutput != nil {
			construct = func() (*MailImageOutputStream, error) { return env.MemoryCacheOutput(stream) }
		}
	case 2:
		f := mailImageRAF(output)
		if f == nil {
			return nil, NewIllegalArgumentException("input not a RandomAccessFile!")
		}
		if env.RAFOutput != nil {
			construct = func() (*MailImageOutputStream, error) { return env.RAFOutput(f) }
		}
	}
	if construct == nil {
		return nil, NewUnsupportedOperationException("ImageIO backing output-stream constructor required")
	}
	result, err := mailInvoke(construct)
	if kind != 1 && err != nil && !isJavaError(err) {
		if printErr := mailImagePrintStackTrace(err); printErr != nil {
			return nil, printErr
		}
		return nil, nil
	}
	return result, err
}
