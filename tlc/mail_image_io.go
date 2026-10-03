/*
 * Copyright (c) 2000, 2022, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 1999, 2014, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 2000, 2004, Oracle and/or its affiliates. All rights reserved.
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
// OpenJDK ImageIO MIME filtering/factory iterators and the AWT operations used
// by Geronimo image handlers. Full IIORegistry discovery/lifetime, desktop rasterization, initialized
// readers/writers and codec implementations remain separate required ports.
package tlc

import (
	"sync/atomic"
)

type MailImageIOEnvironment struct {
	ProviderSPIs     func(readers, ordered bool) (*MailImageSPIIterator, error)
	Deregister       func(*MailImageSPI, bool) (bool, error)
	NewBufferedImage func(int32, int32, int32) (*MailBufferedImage, error)
	ObjectClassName  func(any) string
	// The opaque object is a native DigraphNode. VM identity hashes must remain
	// stable for its lifetime; this hook is configured before registry use.
	GraphNodeIdentityHash func(any) int32
}

var DefaultMailImageIOEnvironment MailImageIOEnvironment

type MailImageSPI struct {
	// Leaf class and virtual object/comparable operations are registry metadata.
	// ComparableClass denotes a verified Comparable<Self> declaration.
	Class                *MailActivationClass
	HashCodeFunc         func() int32
	IdentityHashCodeFunc func() int32
	EqualsFunc           func(any) bool
	ComparableClass      *MailActivationClass
	CompareToFunc        func(any) int32
	identityHash         atomic.Int32
	MIMETypesFunc        func() ([]*string, error)
	CreateReaderFunc     func() (*MailImageReader, error)
	CreateWriterFunc     func() (*MailImageWriter, error)
	OnRegistrationFunc   func(*MailImageServiceRegistry, *MailActivationClass) error
	OnDeregistrationFunc func(*MailImageServiceRegistry, *MailActivationClass) error
}
type MailImageSPIIterator struct {
	HasNextFunc func() (bool, error)
	NextFunc    func() (*MailImageSPI, error)
	RemoveFunc  func() error
}

func (i *MailImageSPIIterator) HasNext() (bool, error) {
	return mailInvoke(func() (bool, error) {
		if i == nil {
			return false, NewNullPointerException()
		}
		return i.HasNextFunc()
	})
}
func (i *MailImageSPIIterator) Next() (*MailImageSPI, error) {
	return mailInvoke(func() (*MailImageSPI, error) {
		if i == nil {
			return nil, NewNullPointerException()
		}
		return i.NextFunc()
	})
}
func (i *MailImageSPIIterator) Remove() error {
	return mailImageRegistryCall(func() error {
		if i == nil {
			return NewNullPointerException()
		}
		if i.RemoveFunc == nil {
			return NewUnsupportedOperationException()
		}
		return i.RemoveFunc()
	})
}
func mailImageSPIList(providers []*MailImageSPI) *MailImageSPIIterator {
	index := 0
	return &MailImageSPIIterator{HasNextFunc: func() (bool, error) { return index < len(providers), nil }, NextFunc: func() (*MailImageSPI, error) {
		if index >= len(providers) {
			return nil, NewNoSuchElementException()
		}
		p := providers[index]
		index++
		return p, nil
	}}
}

type MailImageReader struct {
	ReadFunc func(int32) (*MailBufferedImage, error)
}

func (r *MailImageReader) Read(index int32) (*MailBufferedImage, error) {
	return mailInvoke(func() (*MailBufferedImage, error) {
		if r == nil {
			return nil, NewNullPointerException()
		}
		if r.ReadFunc == nil {
			return nil, NewUnsupportedOperationException("ImageReader read provider required")
		}
		return r.ReadFunc(index)
	})
}

type MailImageWriter struct {
	SetOutputFunc     func(any) error
	WriteRenderedFunc func(*MailRenderedImage) error
	WriteIIOFunc      func(*MailIIOImage) error
	output            any
}

func (w *MailImageWriter) SetOutput(out any) error {
	_, err := mailInvoke(func() (bool, error) {
		if w == nil {
			return false, NewNullPointerException()
		}
		if w.SetOutputFunc != nil {
			return false, w.SetOutputFunc(out)
		}
		return false, NewUnsupportedOperationException("ImageWriter setOutput provider required")
	})
	return err
}
func (w *MailImageWriter) WriteRenderedImage(image *MailRenderedImage) error {
	_, err := mailInvoke(func() (bool, error) {
		if w == nil {
			return false, NewNullPointerException()
		}
		if w.WriteRenderedFunc != nil {
			return false, w.WriteRenderedFunc(image)
		}
		if image == nil {
			return false, NewIllegalArgumentException("image == null!")
		}
		if w.WriteIIOFunc == nil {
			return false, NewUnsupportedOperationException("ImageWriter write provider required")
		}
		return false, w.WriteIIOFunc(&MailIIOImage{Rendered: image})
	})
	return err
}
func (w *MailImageWriter) WriteIIOImage(image *MailIIOImage) error {
	_, err := mailInvoke(func() (bool, error) {
		if w == nil {
			return false, NewNullPointerException()
		}
		if w.WriteIIOFunc == nil {
			return false, NewUnsupportedOperationException("ImageWriter write provider required")
		}
		return false, w.WriteIIOFunc(image)
	})
	return err
}

type mailImageFilteredIterator struct {
	*mailImageRegistryFilter
	emptyCollection bool
}

func mailEmptyImageIterator() *mailImageFilteredIterator {
	return &mailImageFilteredIterator{mailImageRegistryFilter: &mailImageRegistryFilter{source: mailImageSPIList(nil)}, emptyCollection: true}
}
func mailImageMIMEFilter(mime string) func(*MailImageSPI) (bool, error) {
	return func(p *MailImageSPI) (bool, error) {
		// Method.invoke wraps target throwables in InvocationTargetException. The
		// ContainsFilter catches Exception, including that wrapper around an Error.
		found, err := mailInvoke(func() (bool, error) {
			if p == nil {
				return false, NewNullPointerException()
			}
			if p.MIMETypesFunc == nil {
				return false, NewUnsupportedOperationException("Image SPI MIME metadata provider required")
			}
			names, err := p.MIMETypesFunc()
			if err != nil {
				return false, err
			}
			for _, name := range names {
				if name != nil && mailAddressEqualsIgnoreCase(mime, *name) {
					return true, nil
				}
			}
			return false, nil
		})
		return err == nil && found, nil
	}
}

type MailImageReaderIterator struct {
	filtered    *mailImageFilteredIterator
	environment MailImageIOEnvironment
}
type MailImageWriterIterator struct {
	filtered    *mailImageFilteredIterator
	environment MailImageIOEnvironment
}

func mailImageIterator(mime *string, readers bool) (*mailImageFilteredIterator, MailImageIOEnvironment, error) {
	e := DefaultMailImageIOEnvironment
	if mime == nil {
		return nil, e, NewIllegalArgumentException("MIMEType == null!")
	}
	var source *MailImageSPIIterator
	var err error
	if e.ProviderSPIs != nil {
		source, err = mailInvoke(func() (*MailImageSPIIterator, error) { return e.ProviderSPIs(readers, true) })
	} else {
		source = mailStandardImageSPIs(readers)
	}
	if err != nil {
		if _, ok := err.(*IllegalArgumentException); ok {
			return mailEmptyImageIterator(), e, nil
		} else {
			return nil, e, err
		}
	}
	f := &mailImageFilteredIterator{mailImageRegistryFilter: &mailImageRegistryFilter{source: source, filter: mailImageMIMEFilter(*mime)}}
	if err := f.advance(); err != nil {
		if _, ok := err.(*IllegalArgumentException); ok {
			f = mailEmptyImageIterator()
		} else {
			return nil, e, err
		}
	}
	return f, e, nil
}
func GetMailImageReadersByMIMEType(mime *string) (*MailImageReaderIterator, error) {
	return mailInvoke(func() (*MailImageReaderIterator, error) {
		f, e, err := mailImageIterator(mime, true)
		if err != nil {
			return nil, err
		}
		return &MailImageReaderIterator{f, e}, nil
	})
}
func GetMailImageWritersByMIMEType(mime *string) (*MailImageWriterIterator, error) {
	return mailInvoke(func() (*MailImageWriterIterator, error) {
		f, e, err := mailImageIterator(mime, false)
		if err != nil {
			return nil, err
		}
		return &MailImageWriterIterator{f, e}, nil
	})
}
func (i *MailImageReaderIterator) HasNext() (bool, error) {
	if i == nil {
		return false, NewNullPointerException()
	}
	return i.filtered.next != nil, nil
}
func (i *MailImageWriterIterator) HasNext() (bool, error) {
	if i == nil {
		return false, NewNullPointerException()
	}
	return i.filtered.next != nil, nil
}
func (i *MailImageReaderIterator) Remove() error {
	if i == nil {
		return NewNullPointerException()
	}
	// A category error returns Collections.emptyIterator(), whose remove()
	// throws IllegalStateException instead of the ImageReaderIterator error.
	if i.filtered.emptyCollection {
		return NewIllegalStateException()
	}
	return NewUnsupportedOperationException()
}
func (i *MailImageWriterIterator) Remove() error {
	if i == nil {
		return NewNullPointerException()
	}
	if i.filtered.emptyCollection {
		return NewIllegalStateException()
	}
	return NewUnsupportedOperationException()
}
func mailImageDeregister(e MailImageIOEnvironment, p *MailImageSPI, readers bool) error {
	if p == nil {
		return NewIllegalArgumentException("provider == null!")
	}
	if e.Deregister != nil {
		_, err := e.Deregister(p, readers)
		return err
	}
	return mailStandardDeregisterImageSPI(p, readers)
}
func (i *MailImageReaderIterator) Next() (*MailImageReader, error) {
	return mailInvoke(func() (*MailImageReader, error) {
		if i == nil {
			return nil, NewNullPointerException()
		}
		var p *MailImageSPI
		r, err := mailInvoke(func() (*MailImageReader, error) {
			var err error
			p, err = i.filtered.nextProvider()
			if err != nil {
				return nil, err
			}
			if p == nil {
				return nil, NewNullPointerException()
			}
			if p.CreateReaderFunc == nil {
				return nil, NewUnsupportedOperationException("Image SPI reader factory required")
			}
			return p.CreateReaderFunc()
		})
		if isJavaIOException(err) {
			if e := mailImageDeregister(i.environment, p, true); e != nil {
				return nil, e
			}
			return nil, nil
		}
		return r, err
	})
}
func (i *MailImageWriterIterator) Next() (*MailImageWriter, error) {
	return mailInvoke(func() (*MailImageWriter, error) {
		if i == nil {
			return nil, NewNullPointerException()
		}
		var p *MailImageSPI
		w, err := mailInvoke(func() (*MailImageWriter, error) {
			var err error
			p, err = i.filtered.nextProvider()
			if err != nil {
				return nil, err
			}
			if p == nil {
				return nil, NewNullPointerException()
			}
			if p.CreateWriterFunc == nil {
				return nil, NewUnsupportedOperationException("Image SPI writer factory required")
			}
			return p.CreateWriterFunc()
		})
		if isJavaIOException(err) {
			if e := mailImageDeregister(i.environment, p, false); e != nil {
				return nil, e
			}
			return nil, nil
		}
		return w, err
	})
}

// Standard providers use the same ServiceRegistry class map, callbacks and
// ordering graph as public registries. IIORegistry AppContext/ServiceLoader
// discovery and the remaining six standard stream providers are still required.
var mailStandardImageRegistry *MailImageServiceRegistry
var mailStandardImageRegistryInitialization mailActivationClassInitialization

func mailInitStandardImageRegistry() {
	registry, err := NewMailImageServiceRegistry(mailImageClassList([]*MailActivationClass{
		mailImageReaderSPIClass, mailImageWriterSPIClass, mailImageTranscoderSPIClass,
		mailImageInputStreamSPIClass, mailImageOutputStreamSPIClass,
	}))
	if err != nil {
		panic(err)
	}
	for index, kind := range []struct{ mime, input, output string }{
		{"image/gif", "Input not set!", "output == null!"},
		{"image/bmp", "Input has not been set.", "Output has not been set."},
		{"image/vnd.wap.wbmp", "Input has not been set.", "Output has not been set."},
		{"image/tiff", "Input not set!", "output == null!"},
		{"image/png", "Input source not set!", "output == null!"},
		{"image/jpeg", "Input not set", "Output has not been set!"},
	} {
		names := []*string{javaString(kind.mime)}
		if kind.mime == "image/png" {
			names = append(names, javaString("image/x-png"))
		}
		metadata := func() ([]*string, error) { return append([]*string{}, names...), nil }
		reader := &MailImageSPI{Class: mailStandardImageReaderClasses[index], MIMETypesFunc: metadata, CreateReaderFunc: func() (*MailImageReader, error) {
			return &MailImageReader{ReadFunc: func(index int32) (*MailBufferedImage, error) {
				if index != 0 {
					return nil, NewUnsupportedOperationException("Full ImageReader index validation provider required")
				}
				return nil, NewIllegalStateException(kind.input)
			}}, nil
		}}
		writer := &MailImageSPI{Class: mailStandardImageWriterClasses[index], MIMETypesFunc: metadata, CreateWriterFunc: func() (*MailImageWriter, error) {
			w := &MailImageWriter{}
			w.SetOutputFunc = func(value any) error {
				if !mailHandlerNull(value) {
					c, ok := value.(interface{ AsMailImageOutputStream() *MailImageOutputStream })
					if !ok || c.AsMailImageOutputStream() == nil {
						if kind.mime == "image/tiff" {
							return NewIllegalArgumentException("output not an ImageOutputStream!")
						}
						return NewIllegalArgumentException("Illegal output type!")
					}
					return NewUnsupportedOperationException("JVM " + kind.mime + " writer stream initialization provider required")
				}
				w.output = value
				return nil
			}
			w.WriteIIOFunc = func(image *MailIIOImage) error {
				if mailHandlerNull(w.output) {
					return NewIllegalStateException(kind.output)
				}
				return NewUnsupportedOperationException("JVM " + kind.mime + " writer codec provider required")
			}
			return w, nil
		}}
		if err := registry.RegisterServiceProvider(reader); err != nil {
			panic(err)
		}
		if err := registry.RegisterServiceProvider(writer); err != nil {
			panic(err)
		}
	}
	mailStandardImageRegistry = registry
}
func mailStandardImageSPIs(readers bool) *MailImageSPIIterator {
	mailStandardImageRegistryInitialization.initialize("javax.imageio.ImageIO", nil, mailInitStandardImageRegistry)
	category := mailImageWriterSPIClass
	if readers {
		category = mailImageReaderSPIClass
	}
	iter, err := mailStandardImageRegistry.GetServiceProviders(category, true)
	if err != nil {
		panic(err)
	}
	return iter
}
func mailStandardDeregisterImageSPI(p *MailImageSPI, readers bool) error {
	category := mailImageWriterSPIClass
	if readers {
		category = mailImageReaderSPIClass
	}
	_, err := mailStandardImageRegistry.DeregisterServiceProviderInCategory(p, category)
	return err
}

// Stream and AWT carriers retain JVM type distinctions; Go image.Image or
// io.Writer alone does not imply these Java interfaces.
type MailImageOutputStream struct{}

func (s *MailImageOutputStream) AsMailImageOutputStream() *MailImageOutputStream { return s }

type MailRenderedImage struct{ ClassName string }

func (i *MailRenderedImage) AsMailRenderedImage() *MailRenderedImage { return i }

type MailImage struct {
	WidthFunc, HeightFunc func(any) int32
	ClassName             string
}

func (i *MailImage) AsMailImage() *MailImage { return i }
func (i *MailImage) GetWidth(observer any) int32 {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.WidthFunc == nil {
		panic(NewUnsupportedOperationException("AWT Image width provider required"))
	}
	return i.WidthFunc(observer)
}
func (i *MailImage) GetHeight(observer any) int32 {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.HeightFunc == nil {
		panic(NewUnsupportedOperationException("AWT Image height provider required"))
	}
	return i.HeightFunc(observer)
}

type MailRaster struct{}
type MailBufferedImage struct {
	Image        MailImage
	Rendered     MailRenderedImage
	RasterFunc   func() *MailRaster
	GraphicsFunc func() *MailGraphics2D
}

func (i *MailBufferedImage) AsMailBufferedImage() *MailBufferedImage { return i }
func (i *MailBufferedImage) AsMailRenderedImage() *MailRenderedImage {
	if i == nil {
		return nil
	}
	return &i.Rendered
}
func (i *MailBufferedImage) AsMailImage() *MailImage {
	if i == nil {
		return nil
	}
	return &i.Image
}
func (i *MailBufferedImage) GetRaster() *MailRaster {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.RasterFunc == nil {
		panic(NewUnsupportedOperationException("AWT BufferedImage raster provider required"))
	}
	return i.RasterFunc()
}
func (i *MailBufferedImage) CreateGraphics() *MailGraphics2D {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.GraphicsFunc == nil {
		panic(NewUnsupportedOperationException("AWT BufferedImage graphics provider required"))
	}
	return i.GraphicsFunc()
}

type MailGraphics2D struct {
	DrawImageFunc func(*MailImage, int32, int32, any, any) (bool, error)
}

func (g *MailGraphics2D) DrawImage(i *MailImage, x, y int32, background, observer any) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if g == nil {
			return false, NewNullPointerException()
		}
		if g.DrawImageFunc == nil {
			return false, NewUnsupportedOperationException("AWT Graphics2D.drawImage provider required")
		}
		return g.DrawImageFunc(i, x, y, background, observer)
	})
}

type MailIIOImage struct {
	Rendered *MailRenderedImage
	Raster   *MailRaster
}

func NewMailRasterIIOImage(raster *MailRaster) (*MailIIOImage, error) {
	if raster == nil {
		return nil, NewIllegalArgumentException("raster == null!")
	}
	return &MailIIOImage{Raster: raster}, nil
}
func mailAsRenderedImage(value any) (*MailRenderedImage, bool) {
	if mailHandlerNull(value) {
		return nil, false
	}
	i, ok := value.(interface{ AsMailRenderedImage() *MailRenderedImage })
	if !ok {
		return nil, false
	}
	return i.AsMailRenderedImage(), true
}
func mailAsBufferedImage(value any) (*MailBufferedImage, bool) {
	if mailHandlerNull(value) {
		return nil, false
	}
	i, ok := value.(interface{ AsMailBufferedImage() *MailBufferedImage })
	if !ok {
		return nil, false
	}
	return i.AsMailBufferedImage(), true
}
func mailAsImage(value any) (*MailImage, bool) {
	if mailHandlerNull(value) {
		return nil, false
	}
	i, ok := value.(interface{ AsMailImage() *MailImage })
	if !ok {
		return nil, false
	}
	return i.AsMailImage(), true
}
func mailCreateBufferedImage(width, height, kind int32) (*MailBufferedImage, error) {
	if cb := DefaultMailImageIOEnvironment.NewBufferedImage; cb != nil {
		return cb(width, height, kind)
	}
	return nil, NewUnsupportedOperationException("AWT BufferedImage constructor provider required")
}
func mailImageObjectClassName(value any) string {
	if cb := DefaultMailImageIOEnvironment.ObjectClassName; cb != nil {
		return cb(value)
	}
	switch value.(type) {
	case string, *string:
		return "java.lang.String"
	case bool:
		return "java.lang.Boolean"
	case int, int32:
		return "java.lang.Integer"
	case int64:
		return "java.lang.Long"
	case int16:
		return "java.lang.Short"
	case int8:
		return "java.lang.Byte"
	}
	if c, ok := value.(interface{ MailJavaClassName() string }); ok {
		return c.MailJavaClassName()
	}
	panic(NewUnsupportedOperationException("JVM image object Class name provider required"))
}
