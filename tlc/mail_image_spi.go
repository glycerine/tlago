/*
 * Copyright (c) 2000, 2021, Oracle and/or its affiliates. All rights reserved.
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

// OpenJDK IIOServiceProvider/ImageReaderWriterSpi/reader/writer/stream/transcoder
// base algorithms. Concrete codecs, streams and metadata schemas are separate
// implementations; module/class/reflection operations are explicit VM boundaries.
package tlc

import "strings"

// Pointer identity preserves the public mutable STANDARD_INPUT_TYPE and
// STANDARD_OUTPUT_TYPE arrays. Constructors recognize those exact arrays even
// when their elements have been changed, and substitute the real stream class.
type MailImageClassArray struct{ Elements []*MailActivationClass }

var mailImageInputStreamClass = &MailActivationClass{Name: "javax.imageio.stream.ImageInputStream"}
var mailImageOutputStreamClass = &MailActivationClass{Name: "javax.imageio.stream.ImageOutputStream", Parents: []*MailActivationClass{mailImageInputStreamClass}}
var mailImageMetadataFormatClass = &MailActivationClass{Name: "javax.imageio.metadata.IIOMetadataFormat"}
var mailImageStandardInputTypes = &MailImageClassArray{Elements: []*MailActivationClass{mailImageInputStreamClass}}
var mailImageStandardOutputTypes = &MailImageClassArray{Elements: []*MailActivationClass{mailImageOutputStreamClass}}

func MailImageStandardInputTypes() *MailImageClassArray  { return mailImageStandardInputTypes }
func MailImageStandardOutputTypes() *MailImageClassArray { return mailImageStandardOutputTypes }

type MailImageSPIMetadata struct {
	VendorName, Version                                                 *string
	Names, Suffixes, MIMETypes                                          []*string
	PluginClassName                                                     *string
	SupportsStandardStreamMetadataFormat                                bool
	NativeStreamMetadataFormatName, NativeStreamMetadataFormatClassName *string
	ExtraStreamMetadataFormatNames, ExtraStreamMetadataFormatClassNames []*string
	SupportsStandardImageMetadataFormat                                 bool
	NativeImageMetadataFormatName, NativeImageMetadataFormatClassName   *string
	ExtraImageMetadataFormatNames, ExtraImageMetadataFormatClassNames   []*string
	InputTypes, OutputTypes                                             *MailImageClassArray
	WriterSPINames, ReaderSPINames                                      []*string
	InputClass, OutputClass                                             *MailActivationClass
}

// Optional virtual overrides preserve subclass dispatch independently of the
// protected fields read by base method bodies.
type MailImageSPIOverrides struct {
	VendorName, Version, PluginClassName, NativeStreamMetadataFormatName, NativeImageMetadataFormatName                                func() (*string, error)
	FormatNames, FileSuffixes, ExtraStreamMetadataFormatNames, ExtraImageMetadataFormatNames, ImageWriterSPINames, ImageReaderSPINames func() ([]*string, error)
	InputTypes, OutputTypes                                                                                                            func() (*MailImageClassArray, error)
	StandardStreamMetadataSupported, StandardImageMetadataSupported, FormatLossless, CanUseCacheFile, NeedsCacheFile                   func() (bool, error)
	StreamMetadataFormat, ImageMetadataFormat                                                                                          func(*string) (*MailIIOMetadataFormat, error)
	Description                                                                                                                        func(any) (*string, error)
	CreateReader                                                                                                                       func(any) (*MailImageReader, error)
	CreateWriter                                                                                                                       func(any) (*MailImageWriter, error)
	CanDecodeInput                                                                                                                     func(any) (bool, error)
	CanEncodeImage                                                                                                                     func(*MailImageTypeSpecifier) (bool, error)
	CreateInputStream                                                                                                                  func(any, bool, *TLAFile) (*MailImageInputStream, error)
	CreateOutputStream                                                                                                                 func(any, bool, *TLAFile) (*MailImageOutputStream, error)
	ReaderServiceProviderName, WriterServiceProviderName                                                                               func() (*string, error)
	CreateTranscoder                                                                                                                   func() (*MailImageTranscoder, error)
	InputClass, OutputClass                                                                                                            func() (*MailActivationClass, error)
	OwnReader                                                                                                                          func(*MailImageReader) (bool, error)
	OwnWriter                                                                                                                          func(*MailImageWriter) (bool, error)
}
type MailImageInputStream struct{}

func (s *MailImageInputStream) AsMailImageInputStream() *MailImageInputStream { return s }

type MailImageTranscoder struct{ Class *MailActivationClass }
type MailImageTypeSpecifier struct{ Class *MailActivationClass }
type MailIIOMetadataFormat struct{ Class *MailActivationClass }

func (f *MailIIOMetadataFormat) AsMailIIOMetadataFormat() *MailIIOMetadataFormat { return f }

// Module identity, class loaders, exported packages and reflective getInstance
// invocation belong to the VM adapter. A method's Invoke operation returns the
// actual Method.invoke throwable (including InvocationTargetException), not a
// raw target-body exception.
type MailImageSPIModule struct {
	Name       *string
	IsExported func(string, *MailImageSPIModule) (bool, error)
}
type MailImageSPIReflectMethod struct{ Invoke func(any, []any) (any, error) }
type MailImageSPIMetadataEnvironment struct {
	ClassModule                               func(*MailActivationClass) (*MailImageSPIModule, error)
	ClassLoader                               func(*MailActivationClass) (*MailActivationClassLoader, error)
	ForName                                   func(*string, bool, *MailActivationClassLoader) (*MailActivationClass, error)
	GetMethod                                 func(*MailActivationClass, string, []*MailActivationClass) (*MailImageSPIReflectMethod, error)
	StandardFormatInstance                    func() (*MailIIOMetadataFormat, error)
	DoPrivileged                              func(func() error) error
	CreateImageTypeSpecifierFromRenderedImage func(*MailRenderedImage) (*MailImageTypeSpecifier, error)
}

var DefaultMailImageSPIMetadataEnvironment MailImageSPIMetadataEnvironment

func mailImageCloneStrings(v []*string) []*string {
	if v == nil {
		return nil
	}
	return append([]*string{}, v...)
}
func mailImageNonemptyStrings(v []*string) []*string {
	if len(v) == 0 {
		return nil
	}
	return mailImageCloneStrings(v)
}
func mailImageCloneClasses(v *MailImageClassArray) *MailImageClassArray {
	if v == nil {
		panic(NewNullPointerException())
	}
	return &MailImageClassArray{Elements: append([]*MailActivationClass{}, v.Elements...)}
}
func mailImageProviderMetadata(vendor, version *string) (MailImageSPIMetadata, error) {
	if vendor == nil {
		return MailImageSPIMetadata{}, NewIllegalArgumentException("vendorName == null!")
	}
	if version == nil {
		return MailImageSPIMetadata{}, NewIllegalArgumentException("version == null!")
	}
	return MailImageSPIMetadata{VendorName: vendor, Version: version}, nil
}
func NewMailIIOServiceProvider(class *MailActivationClass, vendor, version *string) (*MailImageSPI, error) {
	m, err := mailImageProviderMetadata(vendor, version)
	if err != nil {
		return nil, err
	}
	return &MailImageSPI{Class: class, Metadata: m}, nil
}
func NewMailBlankImageSPI(class *MailActivationClass) *MailImageSPI {
	return &MailImageSPI{Class: class}
}
func NewMailImageReaderWriterSPI(class *MailActivationClass, args MailImageSPIMetadata) (*MailImageSPI, error) {
	m, err := mailImageProviderMetadata(args.VendorName, args.Version)
	if err != nil {
		return nil, err
	}
	if args.Names == nil {
		return nil, NewIllegalArgumentException("names == null!")
	}
	if len(args.Names) == 0 {
		return nil, NewIllegalArgumentException("names.length == 0!")
	}
	if args.PluginClassName == nil {
		return nil, NewIllegalArgumentException("pluginClassName == null!")
	}
	m.Names = mailImageCloneStrings(args.Names)
	m.Suffixes = mailImageNonemptyStrings(args.Suffixes)
	m.MIMETypes = mailImageNonemptyStrings(args.MIMETypes)
	m.PluginClassName = args.PluginClassName
	m.SupportsStandardStreamMetadataFormat = args.SupportsStandardStreamMetadataFormat
	m.NativeStreamMetadataFormatName = args.NativeStreamMetadataFormatName
	m.NativeStreamMetadataFormatClassName = args.NativeStreamMetadataFormatClassName
	m.ExtraStreamMetadataFormatNames = mailImageNonemptyStrings(args.ExtraStreamMetadataFormatNames)
	m.ExtraStreamMetadataFormatClassNames = mailImageNonemptyStrings(args.ExtraStreamMetadataFormatClassNames)
	m.SupportsStandardImageMetadataFormat = args.SupportsStandardImageMetadataFormat
	m.NativeImageMetadataFormatName = args.NativeImageMetadataFormatName
	m.NativeImageMetadataFormatClassName = args.NativeImageMetadataFormatClassName
	m.ExtraImageMetadataFormatNames = mailImageNonemptyStrings(args.ExtraImageMetadataFormatNames)
	m.ExtraImageMetadataFormatClassNames = mailImageNonemptyStrings(args.ExtraImageMetadataFormatClassNames)
	return &MailImageSPI{Class: class, Metadata: m}, nil
}
func NewMailImageReaderSPI(class *MailActivationClass, args MailImageSPIMetadata) (*MailImageSPI, error) {
	p, err := NewMailImageReaderWriterSPI(class, args)
	if err != nil {
		return nil, err
	}
	if args.InputTypes == nil {
		return nil, NewIllegalArgumentException("inputTypes == null!")
	}
	if len(args.InputTypes.Elements) == 0 {
		return nil, NewIllegalArgumentException("inputTypes.length == 0!")
	}
	if args.InputTypes == mailImageStandardInputTypes {
		p.Metadata.InputTypes = &MailImageClassArray{Elements: []*MailActivationClass{mailImageInputStreamClass}}
	} else {
		p.Metadata.InputTypes = mailImageCloneClasses(args.InputTypes)
	}
	p.Metadata.WriterSPINames = mailImageNonemptyStrings(args.WriterSPINames)
	return p, nil
}
func NewMailImageWriterSPI(class *MailActivationClass, args MailImageSPIMetadata) (*MailImageSPI, error) {
	p, err := NewMailImageReaderWriterSPI(class, args)
	if err != nil {
		return nil, err
	}
	if args.OutputTypes == nil {
		return nil, NewIllegalArgumentException("outputTypes == null!")
	}
	if len(args.OutputTypes.Elements) == 0 {
		return nil, NewIllegalArgumentException("outputTypes.length == 0!")
	}
	if args.OutputTypes == mailImageStandardOutputTypes {
		p.Metadata.OutputTypes = &MailImageClassArray{Elements: []*MailActivationClass{mailImageOutputStreamClass}}
	} else {
		p.Metadata.OutputTypes = mailImageCloneClasses(args.OutputTypes)
	}
	p.Metadata.ReaderSPINames = mailImageNonemptyStrings(args.ReaderSPINames)
	return p, nil
}
func NewMailImageInputStreamSPI(class *MailActivationClass, vendor, version *string, inputClass *MailActivationClass) (*MailImageSPI, error) {
	p, err := NewMailIIOServiceProvider(class, vendor, version)
	if err != nil {
		return nil, err
	}
	p.Metadata.InputClass = inputClass
	return p, nil
}
func NewMailImageOutputStreamSPI(class *MailActivationClass, vendor, version *string, outputClass *MailActivationClass) (*MailImageSPI, error) {
	p, err := NewMailIIOServiceProvider(class, vendor, version)
	if err != nil {
		return nil, err
	}
	p.Metadata.OutputClass = outputClass
	return p, nil
}
func NewMailImageTranscoderSPI(class *MailActivationClass, vendor, version *string) (*MailImageSPI, error) {
	return NewMailIIOServiceProvider(class, vendor, version)
}

func (p *MailImageSPI) GetVendorName() (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.VendorName; f != nil {
			return f()
		}
		return p.Metadata.VendorName, nil
	})
}

func (p *MailImageSPI) GetVersion() (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.Version; f != nil {
			return f()
		}
		return p.Metadata.Version, nil
	})
}

func (p *MailImageSPI) GetPluginClassName() (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.PluginClassName; f != nil {
			return f()
		}
		return p.Metadata.PluginClassName, nil
	})
}

func (p *MailImageSPI) GetNativeStreamMetadataFormatName() (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.NativeStreamMetadataFormatName; f != nil {
			return f()
		}
		return p.Metadata.NativeStreamMetadataFormatName, nil
	})
}

func (p *MailImageSPI) GetNativeImageMetadataFormatName() (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.NativeImageMetadataFormatName; f != nil {
			return f()
		}
		return p.Metadata.NativeImageMetadataFormatName, nil
	})
}

func (p *MailImageSPI) GetFormatNames() ([]*string, error) {
	return mailInvoke(func() ([]*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.FormatNames; f != nil {
			return f()
		}
		if p.Metadata.Names == nil {
			return nil, NewNullPointerException()
		}
		return mailImageCloneStrings(p.Metadata.Names), nil
	})
}

func (p *MailImageSPI) GetFileSuffixes() ([]*string, error) {
	return mailInvoke(func() ([]*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.FileSuffixes; f != nil {
			return f()
		}
		return mailImageCloneStrings(p.Metadata.Suffixes), nil
	})
}

func (p *MailImageSPI) GetExtraStreamMetadataFormatNames() ([]*string, error) {
	return mailInvoke(func() ([]*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.ExtraStreamMetadataFormatNames; f != nil {
			return f()
		}
		return mailImageCloneStrings(p.Metadata.ExtraStreamMetadataFormatNames), nil
	})
}

func (p *MailImageSPI) GetExtraImageMetadataFormatNames() ([]*string, error) {
	return mailInvoke(func() ([]*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.ExtraImageMetadataFormatNames; f != nil {
			return f()
		}
		return mailImageCloneStrings(p.Metadata.ExtraImageMetadataFormatNames), nil
	})
}

func (p *MailImageSPI) GetImageWriterSPINames() ([]*string, error) {
	return mailInvoke(func() ([]*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.ImageWriterSPINames; f != nil {
			return f()
		}
		return mailImageCloneStrings(p.Metadata.WriterSPINames), nil
	})
}

func (p *MailImageSPI) GetImageReaderSPINames() ([]*string, error) {
	return mailInvoke(func() ([]*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.ImageReaderSPINames; f != nil {
			return f()
		}
		return mailImageCloneStrings(p.Metadata.ReaderSPINames), nil
	})
}

func (p *MailImageSPI) GetMIMETypes() ([]*string, error) {
	return mailInvoke(func() ([]*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if p.MIMETypesFunc != nil {
			return p.MIMETypesFunc()
		}
		return mailImageCloneStrings(p.Metadata.MIMETypes), nil
	})
}
func (p *MailImageSPI) GetDescription(locale any) (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.Description; f != nil {
			return f(locale)
		}
		return nil, NewUnsupportedOperationException("Image SPI description provider required")
	})
}

func (p *MailImageSPI) GetInputTypes() (*MailImageClassArray, error) {
	return mailInvoke(func() (*MailImageClassArray, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.InputTypes; f != nil {
			return f()
		}
		return mailImageCloneClasses(p.Metadata.InputTypes), nil
	})
}

func (p *MailImageSPI) GetOutputTypes() (*MailImageClassArray, error) {
	return mailInvoke(func() (*MailImageClassArray, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.OutputTypes; f != nil {
			return f()
		}
		return mailImageCloneClasses(p.Metadata.OutputTypes), nil
	})
}

func (p *MailImageSPI) GetInputClass() (*MailActivationClass, error) {
	return mailInvoke(func() (*MailActivationClass, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.InputClass; f != nil {
			return f()
		}
		return p.Metadata.InputClass, nil
	})
}

func (p *MailImageSPI) GetOutputClass() (*MailActivationClass, error) {
	return mailInvoke(func() (*MailActivationClass, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.OutputClass; f != nil {
			return f()
		}
		return p.Metadata.OutputClass, nil
	})
}

func (p *MailImageSPI) IsStandardStreamMetadataFormatSupported() (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.StandardStreamMetadataSupported; f != nil {
			return f()
		}
		return p.Metadata.SupportsStandardStreamMetadataFormat, nil
	})
}

func (p *MailImageSPI) IsStandardImageMetadataFormatSupported() (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.StandardImageMetadataSupported; f != nil {
			return f()
		}
		return p.Metadata.SupportsStandardImageMetadataFormat, nil
	})
}

func (p *MailImageSPI) IsFormatLossless() (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.FormatLossless; f != nil {
			return f()
		}
		return true, nil
	})
}

func (p *MailImageSPI) CanUseCacheFile() (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.CanUseCacheFile; f != nil {
			return f()
		}
		return false, nil
	})
}

func (p *MailImageSPI) NeedsCacheFile() (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.NeedsCacheFile; f != nil {
			return f()
		}
		return false, nil
	})
}

func (p *MailImageSPI) CreateReaderInstance() (*MailImageReader, error) {
	return mailInvoke(func() (*MailImageReader, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if p.CreateReaderFunc != nil {
			return p.CreateReaderFunc()
		}
		return p.CreateReaderInstanceWithExtension(nil)
	})
}
func (p *MailImageSPI) CreateReaderInstanceWithExtension(extension any) (*MailImageReader, error) {
	return mailInvoke(func() (*MailImageReader, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.CreateReader; f != nil {
			return f(extension)
		}
		return nil, NewUnsupportedOperationException("Image SPI reader factory required")
	})
}
func (p *MailImageSPI) CreateWriterInstance() (*MailImageWriter, error) {
	return mailInvoke(func() (*MailImageWriter, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if p.CreateWriterFunc != nil {
			return p.CreateWriterFunc()
		}
		return p.CreateWriterInstanceWithExtension(nil)
	})
}
func (p *MailImageSPI) CreateWriterInstanceWithExtension(extension any) (*MailImageWriter, error) {
	return mailInvoke(func() (*MailImageWriter, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.CreateWriter; f != nil {
			return f(extension)
		}
		return nil, NewUnsupportedOperationException("Image SPI writer factory required")
	})
}
func (p *MailImageSPI) CanDecodeInput(source any) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.CanDecodeInput; f != nil {
			return f(source)
		}
		return false, NewUnsupportedOperationException("Image SPI input probe provider required")
	})
}
func (p *MailImageSPI) CanEncodeImageType(t *MailImageTypeSpecifier) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.CanEncodeImage; f != nil {
			return f(t)
		}
		return false, NewUnsupportedOperationException("Image SPI encode probe provider required")
	})
}
func (p *MailImageSPI) CanEncodeRenderedImage(image *MailRenderedImage) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		t, err := mailImageTypeSpecifierFromRenderedImage(image)
		if err != nil {
			return false, err
		}
		return p.CanEncodeImageType(t)
	})
}

// The static AWT factory checks null before selecting a BufferedImage cached
// type or constructing a specifier from its color/sample models. The non-null
// raster/cache work remains a required AWT provider/port.
func mailImageTypeSpecifierFromRenderedImage(image *MailRenderedImage) (*MailImageTypeSpecifier, error) {
	if image == nil {
		return nil, NewIllegalArgumentException("image == null!")
	}
	f := DefaultMailImageSPIMetadataEnvironment.CreateImageTypeSpecifierFromRenderedImage
	if f == nil {
		return nil, NewUnsupportedOperationException("AWT ImageTypeSpecifier.createFromRenderedImage provider required")
	}
	return f(image)
}
func (p *MailImageSPI) IsOwnReader(reader *MailImageReader) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.OwnReader; f != nil {
			return f(reader)
		}
		if reader == nil {
			return false, NewIllegalArgumentException("reader == null!")
		}
		if reader.Class == nil {
			return false, NewUnsupportedOperationException("ImageReader leaf Class provider required")
		}
		name := reader.Class.GetName()
		return p.Metadata.PluginClassName != nil && name == *p.Metadata.PluginClassName, nil
	})
}
func (p *MailImageSPI) IsOwnWriter(writer *MailImageWriter) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if p == nil {
			return false, NewNullPointerException()
		}
		if f := p.Overrides.OwnWriter; f != nil {
			return f(writer)
		}
		if writer == nil {
			return false, NewIllegalArgumentException("writer == null!")
		}
		if writer.Class == nil {
			return false, NewUnsupportedOperationException("ImageWriter leaf Class provider required")
		}
		name := writer.Class.GetName()
		return p.Metadata.PluginClassName != nil && name == *p.Metadata.PluginClassName, nil
	})
}
func (p *MailImageSPI) CreateInputStreamInstance(input any) (*MailImageInputStream, error) {
	return p.CreateInputStreamInstanceWithCache(input, true, nil)
}
func (p *MailImageSPI) CreateInputStreamInstanceWithCache(input any, useCache bool, cacheDir *TLAFile) (*MailImageInputStream, error) {
	return mailInvoke(func() (*MailImageInputStream, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.CreateInputStream; f != nil {
			return f(input, useCache, cacheDir)
		}
		return nil, NewUnsupportedOperationException("Image SPI input-stream factory required")
	})
}
func (p *MailImageSPI) CreateOutputStreamInstance(output any) (*MailImageOutputStream, error) {
	return p.CreateOutputStreamInstanceWithCache(output, true, nil)
}
func (p *MailImageSPI) CreateOutputStreamInstanceWithCache(output any, useCache bool, cacheDir *TLAFile) (*MailImageOutputStream, error) {
	return mailInvoke(func() (*MailImageOutputStream, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.CreateOutputStream; f != nil {
			return f(output, useCache, cacheDir)
		}
		return nil, NewUnsupportedOperationException("Image SPI output-stream factory required")
	})
}
func (p *MailImageSPI) GetReaderServiceProviderName() (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.ReaderServiceProviderName; f != nil {
			return f()
		}
		return nil, NewUnsupportedOperationException("Image transcoder reader-SPI name provider required")
	})
}
func (p *MailImageSPI) GetWriterServiceProviderName() (*string, error) {
	return mailInvoke(func() (*string, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.WriterServiceProviderName; f != nil {
			return f()
		}
		return nil, NewUnsupportedOperationException("Image transcoder writer-SPI name provider required")
	})
}
func (p *MailImageSPI) CreateTranscoderInstance() (*MailImageTranscoder, error) {
	return mailInvoke(func() (*MailImageTranscoder, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.CreateTranscoder; f != nil {
			return f()
		}
		return nil, NewUnsupportedOperationException("Image transcoder factory provider required")
	})
}

func (p *MailImageSPI) GetStreamMetadataFormat(name *string) (*MailIIOMetadataFormat, error) {
	return mailInvoke(func() (*MailIIOMetadataFormat, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.StreamMetadataFormat; f != nil {
			return f(name)
		}
		m := p.Metadata
		return p.metadataFormat(name, m.SupportsStandardStreamMetadataFormat, m.NativeStreamMetadataFormatName, m.NativeStreamMetadataFormatClassName, m.ExtraStreamMetadataFormatNames, m.ExtraStreamMetadataFormatClassNames)
	})
}
func (p *MailImageSPI) GetImageMetadataFormat(name *string) (*MailIIOMetadataFormat, error) {
	return mailInvoke(func() (*MailIIOMetadataFormat, error) {
		if p == nil {
			return nil, NewNullPointerException()
		}
		if f := p.Overrides.ImageMetadataFormat; f != nil {
			return f(name)
		}
		m := p.Metadata
		return p.metadataFormat(name, m.SupportsStandardImageMetadataFormat, m.NativeImageMetadataFormatName, m.NativeImageMetadataFormatClassName, m.ExtraImageMetadataFormatNames, m.ExtraImageMetadataFormatClassNames)
	})
}
func (p *MailImageSPI) metadataFormat(name *string, standard bool, nativeName, nativeClass *string, extraNames, extraClasses []*string) (*MailIIOMetadataFormat, error) {
	if name == nil {
		return nil, NewIllegalArgumentException("formatName == null!")
	}
	env := DefaultMailImageSPIMetadataEnvironment
	if standard && *name == "javax_imageio_1.0" {
		if env.StandardFormatInstance == nil {
			return nil, NewUnsupportedOperationException("JVM standard IIOMetadataFormat schema/instance provider required")
		}
		return env.StandardFormatInstance()
	}
	var className *string
	if nativeName != nil && *name == *nativeName {
		className = nativeClass
	} else if extraNames != nil {
		for i, extra := range extraNames {
			if extra != nil && *name == *extra {
				if extraClasses == nil {
					return nil, NewNullPointerException()
				}
				if i >= len(extraClasses) {
					return nil, NewArrayIndexOutOfBoundsException(i, len(extraClasses))
				}
				className = extraClasses[i]
				break
			}
		}
	}
	if className == nil {
		return nil, NewIllegalArgumentException("Unsupported format name")
	}
	result, err := mailInvoke(func() (*MailIIOMetadataFormat, error) {
		var class *MailActivationClass
		action := func() error { var err error; class, err = p.metadataFormatClass(*className); return err }
		if env.DoPrivileged != nil {
			if err := env.DoPrivileged(action); err != nil {
				return nil, err
			}
		} else {
			if mailImageSecurityManagerPresent() {
				return nil, NewUnsupportedOperationException("JVM secured metadata-format doPrivileged provider required")
			}
			if err := action(); err != nil {
				return nil, err
			}
		}
		if class == nil {
			return nil, NewNullPointerException()
		}
		if env.GetMethod == nil {
			return nil, NewUnsupportedOperationException("JVM metadata-format getMethod/invoke provider required")
		}
		method, err := env.GetMethod(class, "getInstance", []*MailActivationClass{})
		if err != nil {
			return nil, err
		}
		if method == nil {
			return nil, NewNullPointerException()
		}
		if method.Invoke == nil {
			return nil, NewUnsupportedOperationException("JVM Method.invoke provider required")
		}
		value, err := method.Invoke(nil, []any{})
		if err != nil {
			return nil, err
		}
		if mailHandlerNull(value) {
			return nil, nil
		}
		if f, ok := value.(interface{ AsMailIIOMetadataFormat() *MailIIOMetadataFormat }); ok {
			format := f.AsMailIIOMetadataFormat()
			if format != nil {
				return format, nil
			}
		}
		return nil, NewClassCastException()
	})
	if err != nil && !isJavaError(err) {
		failure := NewIllegalStateException("Can't obtain format")
		failure.Cause = err
		return nil, failure
	}
	return result, err
}
func (p *MailImageSPI) metadataFormatClass(name string) (*MailActivationClass, error) {
	env := DefaultMailImageSPIMetadataEnvironment
	if env.ClassModule == nil {
		return nil, NewUnsupportedOperationException("JVM Image SPI module metadata provider required")
	}
	sourceModule, err := env.ClassModule(mailImageReaderWriterSPIClass)
	if err != nil {
		return nil, err
	}
	class := mailImageProviderClass(p)
	targetModule, err := env.ClassModule(class)
	if err != nil {
		return nil, err
	}
	loaded, err := mailInvoke(func() (*MailActivationClass, error) {
		if env.ClassLoader == nil {
			return nil, NewUnsupportedOperationException("JVM Image SPI defining-loader provider required")
		}
		loader, err := env.ClassLoader(class)
		if err != nil {
			return nil, err
		}
		if env.ForName != nil {
			return env.ForName(javaString(name), false, loader)
		}
		if loader == nil {
			return nil, NewUnsupportedOperationException("JVM bootstrap metadata-format Class.forName provider required")
		}
		return loader.LoadClass(javaString(name))
	})
	if err != nil {
		if _, ok := err.(*ClassNotFoundException); ok {
			loaded = nil
		} else {
			return nil, err
		}
	}
	if loaded != nil && !mailImageMetadataFormatClass.IsAssignableFrom(loaded) {
		return nil, nil
	}
	if sourceModule == nil {
		return nil, NewNullPointerException()
	}
	if sourceModule == targetModule || loaded == nil {
		return loaded, nil
	}
	if targetModule == nil {
		return nil, NewNullPointerException()
	}
	if targetModule.Name != nil {
		index := strings.LastIndex(name, ".")
		pkg := ""
		if index > 0 {
			pkg = name[:index]
		}
		if targetModule.IsExported == nil {
			return nil, NewUnsupportedOperationException("JVM metadata-format module export provider required")
		}
		exported, err := targetModule.IsExported(pkg, sourceModule)
		if err != nil {
			return nil, err
		}
		if !exported {
			return nil, NewIllegalStateException("Class " + name + " in named module must be exported to java.desktop module.")
		}
	}
	return loaded, nil
}

func (p *MailImageSPI) OnRegistration(registry *MailImageServiceRegistry, category *MailActivationClass) error {
	return mailImageRegistryCall(func() error {
		if p == nil {
			return NewNullPointerException()
		}
		if p.OnRegistrationFunc != nil {
			return p.OnRegistrationFunc(registry, category)
		}
		return nil
	})
}
func (p *MailImageSPI) OnDeregistration(registry *MailImageServiceRegistry, category *MailActivationClass) error {
	return mailImageRegistryCall(func() error {
		if p == nil {
			return NewNullPointerException()
		}
		if p.OnDeregistrationFunc != nil {
			return p.OnDeregistrationFunc(registry, category)
		}
		return nil
	})
}
