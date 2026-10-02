/**
 *  Licensed to the Apache Software Foundation (ASF) under one or more
 *  contributor license agreements.  See the NOTICE file distributed with
 *  this work for additional information regarding copyright ownership.
 *  The ASF licenses this file to You under the Apache License, Version 2.0
 *  (the "License"); you may not use this file except in compliance with
 *  the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Existing Geronimo MailcapCommandMapTest methods translated after the registry
// parsing, discovery and lookup features. Disabled upstream assertions stay disabled.
// Source: geronimo-activation_1.1_spec/src/test/java/javax/activation/MailcapCommandMapTest.java.
package tlc

import "testing"

func TestJavaMailActivationMailcapCommandMap(t *testing.T) {
	assertInfo := func(t *testing.T, info *MailCommandInfo, name, class string) {
		t.Helper()
		if info == nil {
			t.Fatal("null command info")
		}
		javaActivationAssertString(t, info.GetCommandName(), name)
		javaActivationAssertString(t, info.GetCommandClass(), class)
	}
	// The original setUp creates a new map before every method, including the
	// parameterized method whose entire body is commented out upstream.
	run := func(name string, body func(*testing.T, *MailMailcapCommandMap)) {
		t.Run(name, func(t *testing.T) { body(t, NewMailMailcapCommandMap()) })
	}
	run("testAdd", func(t *testing.T, m *MailMailcapCommandMap) {
		m.AddMailcap(javaString("foo/bar ;; x-java-view=Foo; x-java-edit=Bar"))
		assertInfo(t, m.GetCommand(javaString("foo/bar"), javaString("view")), "view", "Foo")
		assertInfo(t, m.GetCommand(javaString("foo/bar"), javaString("edit")), "edit", "Bar")
	})
	run("testExplicitWildcard", func(t *testing.T, m *MailMailcapCommandMap) {
		m.AddMailcap(javaString("foo/bar ;; x-java-view=Bar"))
		m.AddMailcap(javaString("foo/* ;; x-java-view=Star"))
		assertInfo(t, m.GetCommand(javaString("foo/bar"), javaString("view")), "view", "Bar")
		assertInfo(t, m.GetCommand(javaString("foo/foo"), javaString("view")), "view", "Star")
		assertInfo(t, m.GetCommand(javaString("foo/*"), javaString("view")), "view", "Star")
		// The original reference-implementation comment disables the foo query.
	})
	run("testImplicitWildcard", func(t *testing.T, m *MailMailcapCommandMap) {
		m.AddMailcap(javaString("foo/bar ;; x-java-view=Bar"))
		m.AddMailcap(javaString("foo ;; x-java-view=Star"))
		assertInfo(t, m.GetCommand(javaString("foo/bar"), javaString("view")), "view", "Bar")
		assertInfo(t, m.GetCommand(javaString("foo/foo"), javaString("view")), "view", "Star")
		// The original RI comment disables the foo/* query.
	})
	run("testParameterizedMimeType", func(t *testing.T, m *MailMailcapCommandMap) {
		// The entire original method is commented out and has no assertions.
	})
	run("testGetNativeCommands", func(t *testing.T, m *MailMailcapCommandMap) {
		tmap := NewMailMailcapCommandMap()
		for _, row := range []string{
			"image/gif;;x-java-view=com.sun.activation.viewers.ImageViewer",
			"image/jpeg;;x-java-view=com.sun.activation.viewers.ImageViewer",
			"text/*;yada yada;x-java-view=com.sun.activation.viewers.TextViewer",
			"text/*;;x-java-edit=com.sun.activation.viewers.TextEditor",
			"text/*; ;x-java-edit=com.sun.activation.viewers.TextEditor",
			"text/plain;yada yada;x-java-view=com.sun.activation.viewers.TextViewer",
			"text/plain;bad a bing;x-java-view=com.sun.activation.viewers.TextViewer",
		} {
			tmap.AddMailcap(javaString(row))
		}
		commands := tmap.GetNativeCommands(javaString("text/*"))
		if len(commands) != 1 {
			t.Fatal(len(commands))
		}
		if commands[0] != "text/*;yada yada;x-java-view=com.sun.activation.viewers.TextViewer" {
			t.Fatal(commands[0])
		}
		commands = tmap.GetNativeCommands(javaString("text/plain"))
		if len(commands) != 2 {
			t.Fatal(len(commands))
		}
		for _, command := range commands {
			if command != "text/plain;yada yada;x-java-view=com.sun.activation.viewers.TextViewer" && command != "text/plain;bad a bing;x-java-view=com.sun.activation.viewers.TextViewer" {
				t.Fatal(command)
			}
		}
		if commands = tmap.GetNativeCommands(javaString("image/gif")); len(commands) != 0 {
			t.Fatal(len(commands))
		}
		if commands = tmap.GetNativeCommands(javaString("text/html")); len(commands) != 0 {
			t.Fatal(len(commands))
		}
	})
}
