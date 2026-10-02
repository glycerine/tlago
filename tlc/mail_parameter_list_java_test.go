/*
 * Copyright (c) 2009, 2018 Oracle and/or its affiliates. All rights reserved.
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

// Ports the existing JavaMail 1.6.8 ParameterListTests, WindowsFileNames,
// AppleFileNames, DecodeParameters and ParametersNoStrict JUnit assertions.
// ParameterListTestSuite uses isolated class loaders for static flags. Each Go
// suite runs in a fresh process, preserving first-use property lifetimes.
// NonAsciiFileNames awaits the separate MimeBodyPart feature it exercises.
// Source: https://github.com/eclipse-ee4j/mail/tree/1.6.8/mail/src/test/java/javax/mail/internet
package tlc

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestJavaMailParameterListSuites(t *testing.T) {
	const modeKey = "TLAGO_JAVAMAIL_PARAMETER_SUITE"
	mode := os.Getenv(modeKey)
	if mode == "" {
		for _, suite := range []string{"Basic", "Windows", "Apple", "Decode", "NoStrict"} {
			t.Run(suite, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestJavaMailParameterListSuites$", "-test.v")
				for _, value := range os.Environ() {
					key := strings.SplitN(value, "=", 2)[0]
					if key == modeKey || strings.HasPrefix(key, "mail.mime.") {
						continue
					}
					cmd.Env = append(cmd.Env, value)
				}
				cmd.Env = append(cmd.Env, modeKey+"="+suite)
				switch suite {
				case "Windows":
					cmd.Env = append(cmd.Env, "mail.mime.windowsfilenames=true")
				case "Apple":
					cmd.Env = append(cmd.Env, "mail.mime.applefilenames=true")
				case "Decode":
					cmd.Env = append(cmd.Env, "mail.mime.decodeparameters=true")
				case "NoStrict":
					cmd.Env = append(cmd.Env, "mail.mime.parameters.strict=false")
				}
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("source suite %s: %v\n%s", suite, err, output)
				}
				t.Logf("%s", output)
			})
		}
		return
	}
	get := func(t *testing.T, header, name string) *string {
		t.Helper()
		p, err := NewMailParameterList(javaString(header))
		if err != nil {
			t.Fatal(err)
		}
		return p.Get(name)
	}
	equal := func(t *testing.T, got *string, want string) {
		t.Helper()
		if got == nil || *got != want {
			t.Fatalf("got %v, want %q", got, want)
		}
	}
	switch mode {
	case "Basic":
		t.Run("testBackslash", func(t *testing.T) { equal(t, get(t, `; filename="\a\b\c.txt"`, "filename"), "abc.txt") })
		p0, p1 := strings.Repeat("a", 56), strings.Repeat("b", 56)
		t.Run("testLongParse", func(t *testing.T) { equal(t, get(t, "; p*0="+p0+"; p*1="+p1, "p"), p0+p1) })
		t.Run("testLongSet", func(t *testing.T) {
			p, _ := NewMailParameterList()
			if err := p.Set("p", javaString(p0+p1)); err != nil {
				t.Fatal(err)
			}
			equal(t, p.Get("p"), p0+p1)
			s := p.String()
			if !strings.Contains(s, "p*0=") || !strings.Contains(s, "p*1=") {
				t.Fatal("long value not split into p*0 and p*1")
			}
		})
	case "Windows":
		equal(t, get(t, `; filename="\a\b\c.txt"`, "filename"), `\a\b\c.txt`)
	case "Apple":
		equal(t, get(t, "; filename=a b.txt", "filename"), "a b.txt")
	case "Decode", "NoStrict":
		resource := "paramdata"
		if mode == "NoStrict" {
			resource = "paramdatanostrict"
		}
		for i, c := range loadJavaMailHeaderCases(t, resource, true) {
			t.Run(fmt.Sprintf("%03d_%s", i, c.name), func(t *testing.T) {
				ct, err := NewMailContentType(javaString(c.value))
				if err != nil {
					if _, ok := err.(*MailParseException); !ok {
						t.Fatal(err)
					}
					if len(c.expected) != 1 || c.expected[0] != "Exception" {
						t.Fatalf("Expected exception: %v", err)
					}
					return
				}
				pl := ct.Parameters
				if pl.Size() != len(c.expected) {
					t.Fatalf("Number of parameters: got %d, want %d", pl.Size(), len(c.expected))
				}
				for i, name := range pl.GetNames() {
					if i < len(c.expected) {
						value := name + "=" + javaNullableString(pl.Get(name))
						if value != c.expected[i] {
							t.Errorf("Parameter value: got %q, want %q", value, c.expected[i])
						}
					}
				}
			})
		}
	default:
		t.Fatalf("unknown source suite %q", mode)
	}
}
