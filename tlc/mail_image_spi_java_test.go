/*
 * Copyright (c) 2001, 2017, Oracle and/or its affiliates. All rights reserved.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * This code is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 2 only, as
 * published by the Free Software Foundation.
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

// Translated after the SPI constructor algorithms from OpenJDK SpiTest
// testImageReaderSpiConstructor and testImageWriterSpiConstructor. Its standard
// registry/metadata-format enumeration remains with IIORegistry/schema work.
package tlc

import "testing"

func TestMailImageSPIJavaReaderConstructor(t *testing.T) { mailImageSPIJavaConstructor(t, true) }
func TestMailImageSPIJavaWriterConstructor(t *testing.T) { mailImageSPIJavaConstructor(t, false) }
func mailImageSPIJavaConstructor(t *testing.T, reader bool) {
	t.Helper()
	var args MailImageSPIMetadata
	class := &MailActivationClass{Name: "SpiTest$2", Parents: []*MailActivationClass{mailImageWriterSPIClass}}
	if reader {
		class = &MailActivationClass{Name: "SpiTest$1", Parents: []*MailActivationClass{mailImageReaderSPIClass}}
	}
	construct := func() (*MailImageSPI, error) {
		var p *MailImageSPI
		var err error
		if reader {
			p, err = NewMailImageReaderSPI(class, args)
		} else {
			p, err = NewMailImageWriterSPI(class, args)
		}
		if err != nil {
			return nil, err
		}
		p.Overrides.Description = func(any) (*string, error) { return nil, nil }
		if reader {
			p.Overrides.CanDecodeInput = func(any) (bool, error) { return false, nil }
			p.Overrides.CreateReader = func(any) (*MailImageReader, error) { return nil, nil }
		} else {
			p.Overrides.CanEncodeImage = func(*MailImageTypeSpecifier) (bool, error) { return false, nil }
			p.Overrides.CreateWriter = func(any) (*MailImageWriter, error) { return nil, nil }
		}
		return p, nil
	}
	check := func(shouldFail bool) {
		t.Helper()
		_, err := construct()
		if _, ok := err.(*IllegalArgumentException); ok {
			if shouldFail {
				return
			}
			t.Fatal("SPI constructor threw an IAE")
		}
		if err != nil {
			t.Fatal(err)
		}
		if shouldFail {
			t.Fatal("SPI constructor did not throw an IAE")
		}
	}
	check(true)
	args.VendorName = javaString("My Vendor")
	check(true)
	args.Version = javaString("My Version")
	check(true)
	args.Names = []*string{}
	check(true)
	args.Names = []*string{javaString("My Format Name")}
	check(true)
	if reader {
		args.PluginClassName = javaString("com.mycompany.Reader")
	} else {
		args.PluginClassName = javaString("com.mycompany.Writer")
	}
	check(true)
	if reader {
		args.InputTypes = &MailImageClassArray{Elements: []*MailActivationClass{}}
	} else {
		args.OutputTypes = &MailImageClassArray{Elements: []*MailActivationClass{}}
	}
	check(true)
	if reader {
		args.InputTypes = &MailImageClassArray{Elements: []*MailActivationClass{mailFlavorObjectClass}}
	} else {
		args.OutputTypes = &MailImageClassArray{Elements: []*MailActivationClass{mailFlavorObjectClass}}
	}
	check(false)
	args.Suffixes = []*string{}
	args.MIMETypes = []*string{}
	if reader {
		args.WriterSPINames = []*string{}
	} else {
		args.ReaderSPINames = []*string{}
	}
	args.ExtraStreamMetadataFormatNames = []*string{}
	args.ExtraImageMetadataFormatNames = []*string{}
	if !reader {
		args.ExtraStreamMetadataFormatClassNames = []*string{}
		args.ExtraImageMetadataFormatClassNames = []*string{}
	}
	p, err := construct()
	if err != nil {
		t.Fatal(err)
	}
	for _, get := range []struct {
		name string
		fn   func() ([]*string, error)
	}{{"suffixes", p.GetFileSuffixes}, {"MIMETypes", p.GetMIMETypes}, {"extraStreamMetadataFormatNames", p.GetExtraStreamMetadataFormatNames}, {"extraImageMetadataFormatNames", p.GetExtraImageMetadataFormatNames}} {
		values, err := get.fn()
		if err != nil {
			t.Fatal(err)
		}
		if values != nil {
			t.Fatalf("failed to normalize %s", get.name)
		}
	}
	if reader {
		values, err := p.GetImageWriterSPINames()
		if err != nil {
			t.Fatal(err)
		}
		if values != nil {
			t.Fatal("failed to normalize writerSpiNames")
		}
	} else {
		values, err := p.GetImageReaderSPINames()
		if err != nil {
			t.Fatal(err)
		}
		if values != nil {
			t.Fatal("failed to normalize readerSpiNames")
		}
	}
}
