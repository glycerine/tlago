/*
 * Copyright (c) 2015, 2016, Oracle and/or its affiliates. All rights reserved.
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

// Translated after ServiceRegistry/SubRegistry implementation from OpenJDK
// ServiceRegistryRestriction (8068749) and ServiceRegistrySyncTest (8022640).
package tlc

import (
	"testing"
	"time"
)

func TestMailImageServiceRegistryRestriction(t *testing.T) {
	loader := &MailActivationClassLoader{}
	dummy := &MailActivationClass{Name: "ServiceRegistryRestriction$DummyTestSpi", Parents: []*MailActivationClass{mailFlavorObjectClass}}
	for _, tc := range []struct {
		class *MailActivationClass
		fail  bool
	}{
		{mailImageInputStreamSPIClass, false}, {mailImageOutputStreamSPIClass, false},
		{mailImageReaderSPIClass, false}, {mailImageTranscoderSPIClass, false},
		{mailImageWriterSPIClass, false}, {dummy, true},
	} {
		for _, operation := range []struct {
			name string
			run  func(*MailActivationClass) error
		}{
			{"constructor", func(c *MailActivationClass) error {
				_, err := NewMailImageServiceRegistry(mailImageClassList([]*MailActivationClass{c}))
				return err
			}},
			{"lookup", func(c *MailActivationClass) error { _, err := LookupMailImageProviders(c); return err }},
			{"lookupCL", func(c *MailActivationClass) error {
				_, err := LookupMailImageProvidersWithLoader(c, loader)
				return err
			}},
		} {
			t.Run(tc.class.GetName()+"/"+operation.name, func(t *testing.T) {
				err := operation.run(tc.class)
				if _, ok := err.(*IllegalArgumentException); ok && tc.fail {
					return
				}
				if err != nil {
					t.Fatalf("unexpected exception: %v", err)
				}
				if tc.fail {
					t.Fatal("operation succeeded unexpectedly")
				}
			})
		}
	}
}

func TestMailImageServiceRegistrySync(t *testing.T) {
	reg, err := NewMailImageServiceRegistry(mailImageClassList([]*MailActivationClass{mailImageInputStreamSPIClass}))
	if err != nil {
		t.Fatal(err)
	}
	// MyService's default SPI constructor has no registration side effects.
	// Its abstract description/stream factories return null and are not invoked.
	class := &MailActivationClass{Name: "MyService", Parents: []*MailActivationClass{mailImageInputStreamSPIClass}}
	done := make(chan error, 2)
	go func() {
		services := make([]*MailImageSPI, 20)
		for i := range services {
			services[i] = &MailImageSPI{Class: class}
		}
		end := time.Now().Add(5000 * time.Millisecond)
		for time.Now().Before(end) {
			for _, s := range services {
				if err := reg.RegisterServiceProvider(s); err != nil {
					done <- err
					return
				}
			}
			for _, s := range services {
				if err := reg.DeregisterServiceProvider(s); err != nil {
					done <- err
					return
				}
			}
		}
		done <- nil
	}()
	go func() {
		end := time.Now().Add(5000 * time.Millisecond)
		for time.Now().Before(end) {
			if _, err := reg.GetServiceProviders(mailImageInputStreamSPIClass, true); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	// Join both native goroutines as the JVM waits for both non-daemon threads.
	for range 2 {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}
}
