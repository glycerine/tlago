/*
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Finn Hackett - initial test, related to TLC string deserialization support
 ******************************************************************************/
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// StringDeserializeTLCTest.test, including its ModelCheckerTestCase setup and
// successful exit assertion. The original .vos bytes are copied without using
// the Go Value API, so reading cannot benefit from pre-interning the data.
func TestJavaStringDeserialize(t *testing.T) {
	path, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "StringDeserialize", "StringDeserialize.vos"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("StringDeserializeTLCTest.vosFileName", path)
	// The original Java classpath supplies IOUtils from CommunityModules.
	modules, err := filepath.Abs(filepath.Join("test_vectors", "java-sany", "CommunityModules.jar"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLASSPATH", modules)
	result := runJavaTLCModelTest(t, "StringDeserialize")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status %d, want success; messages: %+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}
