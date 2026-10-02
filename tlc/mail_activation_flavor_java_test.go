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

// Geronimo ActivationDataFlavorTest, translated after the feature's Go port.
// Source: apache/geronimo-specs geronimo-activation_1.1_spec/src/test/java/javax/activation.
package tlc

import "testing"

func TestJavaMailActivationDataFlavor(t *testing.T) {
	t.Run("testMimeTypeConstructorWithoutClass", func(t *testing.T) {
		adf := NewMailActivationDataFlavor(javaString("application/*"), nil)
		javaActivationAssertString(t, adf.GetMIMEType(), "application/*")
		if adf.GetRepresentationClass() != mailFlavorInputStreamClass {
			t.Fatal("representation class")
		}
	})
	t.Run("testMimeTypeConstructorWithClass", func(t *testing.T) {
		mime := "application/x-java-serialized-object; class=java.lang.Object"
		adf := NewMailActivationDataFlavor(javaString(mime), nil)
		javaActivationAssertString(t, adf.GetMIMEType(), mime)
		if adf.GetRepresentationClass() != mailFlavorInputStreamClass {
			t.Fatal("representation class")
		}
	})
	t.Run("testHumanName", func(t *testing.T) {
		adf := NewMailActivationDataFlavor(javaString("text/html"), javaString("Human Name"))
		javaActivationAssertString(t, adf.GetHumanPresentableName(), "Human Name")
		adf.SetHumanPresentableName(javaString("Name 2"))
		javaActivationAssertString(t, adf.GetHumanPresentableName(), "Name 2")
		adf = NewMailActivationDataFlavor(javaString("text/html"), nil)
		if adf.GetHumanPresentableName() != nil {
			t.Fatal("human name")
		}
	})
	t.Run("testEquals", func(t *testing.T) {
		adf1 := NewMailActivationDataFlavor(javaString("text/plain"), javaString("text/plain"))
		adf2, err := NewMailMIMEDataFlavor(javaString("text/plain"), javaString("text/plain"))
		if err != nil {
			t.Fatal(err)
		}
		if !adf1.Equals(adf2) {
			t.Fatal("equals")
		}
	})
}
