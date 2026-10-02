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

// Bundled Geronimo image handler algorithms, reconstructed from its bytecode.
// ImageIO lookup, image methods and raster/graphics operations have distinct
// native carriers; the handler never sets reader input or disposes graphics.
package tlc

import "io"

type MailUnsupportedDataTypeException struct{ javaExceptionBase }

func NewMailUnsupportedDataTypeException(message ...string) *MailUnsupportedDataTypeException {
	return &MailUnsupportedDataTypeException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *MailUnsupportedDataTypeException) Error() string { return javaThrowableMessage(e) }

func NewMailAbstractImageHandler(flavor *MailDataFlavor) *MailDataContentHandler {
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
		readers, err := GetMailImageReadersByMIMEType(source.GetContentType())
		if err != nil {
			return nil, err
		}
		more, err := readers.HasNext()
		if err != nil {
			return nil, err
		}
		if !more {
			return nil, NewMailUnsupportedDataTypeException("Unknown image type " + javaNullableString(source.GetContentType()))
		}
		reader, err := readers.Next()
		if err != nil {
			return nil, err
		}
		image, err := reader.Read(0)
		if err != nil {
			return nil, err
		}
		if image == nil {
			return nil, nil
		}
		return image, nil
	}
	h.WriteToFunc = func(value any, mime *string, out io.Writer) error {
		writers, err := GetMailImageWritersByMIMEType(mime)
		if err != nil {
			return err
		}
		more, err := writers.HasNext()
		if err != nil {
			return err
		}
		if !more {
			return NewMailUnsupportedDataTypeException("Unknown image type " + javaNullableString(mime))
		}
		writer, err := writers.Next()
		if err != nil {
			return err
		}
		if err = writer.SetOutput(out); err != nil {
			return err
		}
		if rendered, ok := mailAsRenderedImage(value); ok {
			err = writer.WriteRenderedImage(rendered)
		} else if buffered, ok := mailAsBufferedImage(value); ok {
			var image *MailIIOImage
			image, err = NewMailRasterIIOImage(buffered.GetRaster())
			if err == nil {
				err = writer.WriteIIOImage(image)
			}
		} else if image, ok := mailAsImage(value); ok {
			width := image.GetWidth(nil)
			height := image.GetHeight(nil)
			buffer, createErr := mailCreateBufferedImage(width, height, 2)
			if createErr != nil {
				return createErr
			}
			graphics := buffer.CreateGraphics()
			_, err = graphics.DrawImage(image, 0, 0, nil, nil)
			if err != nil {
				return err
			}
			var iio *MailIIOImage
			iio, err = NewMailRasterIIOImage(buffer.GetRaster())
			if err == nil {
				err = writer.WriteIIOImage(iio)
			}
		} else {
			if mailHandlerNull(value) {
				return NewNullPointerException()
			}
			return NewMailUnsupportedDataTypeException("Unknown image type " + mailImageObjectClassName(value))
		}
		if err != nil {
			return err
		}
		if mailHandlerNull(out) {
			return NewNullPointerException()
		}
		if f, ok := out.(interface{ Flush() error }); ok {
			return f.Flush()
		}
		return nil
	}
	return h
}
func NewMailImageGIFHandler() *MailDataContentHandler {
	return NewMailAbstractImageHandler(NewMailActivationTypedDataFlavor(mailFlavorImageClass, javaString("image/gif"), javaString("GIF Image")))
}
func NewMailImageJPEGHandler() *MailDataContentHandler {
	return NewMailAbstractImageHandler(NewMailActivationTypedDataFlavor(mailFlavorImageClass, javaString("image/jpeg"), javaString("JPEG Image")))
}

var mailAbstractImageHandlerClass = &MailActivationClass{Name: "org.apache.geronimo.activation.handlers.AbstractImageHandler", Parents: []*MailActivationClass{mailFlavorObjectClass}}
var mailImageGIFHandlerClass = &MailActivationClass{Name: "org.apache.geronimo.activation.handlers.ImageGifHandler", Parents: []*MailActivationClass{mailAbstractImageHandlerClass}}
var mailImageJPEGHandlerClass = &MailActivationClass{Name: "org.apache.geronimo.activation.handlers.ImageJpegHandler", Parents: []*MailActivationClass{mailAbstractImageHandlerClass}}

func init() {
	mailImageGIFHandlerClass.NewInstance = func() (any, error) { return NewMailImageGIFHandler(), nil }
	mailImageJPEGHandlerClass.NewInstance = func() (any, error) { return NewMailImageJPEGHandler(), nil }
}
