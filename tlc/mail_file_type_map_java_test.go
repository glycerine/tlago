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

// Port of the bundled Activation implementation's upstream
// MimetypesFileTypeMapTest.testDefault/testCommentRemoval, after the feature.
// Source: https://github.com/apache/geronimo-specs/blob/trunk/geronimo-activation_1.1_spec/src/test/java/javax/activation/MimetypesFileTypeMapTest.java
package tlc

import "testing"

func TestJavaMailMimetypesFileTypeMap(t *testing.T) {
	t.Run("testDefault", func(t *testing.T) {
		m := NewMailMimetypesFileTypeMap()
		if v := m.GetContentType("x.foo"); v != "application/octet-stream" {
			t.Fatal(v)
		}
		if v := m.GetContentType("x.html"); v != "text/html" {
			t.Fatal(v)
		}
	})
	t.Run("testCommentRemoval", func(t *testing.T) {
		m := NewMailMimetypesFileTypeMap()
		for _, line := range []string{" text/foo foo #txt", "#text/foo bar", "text/foo #bar"} {
			m.AddMIMETypes(line)
			if v := m.GetContentType("x.foo"); v != "text/foo" {
				t.Fatal(v)
			}
		}
	})
}
