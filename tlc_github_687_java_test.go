/*******************************************************************************
 * Copyright (c) 2024 Microsoft. All rights reserved.
 *
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
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original Github687Test.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687", "Github687", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCSuccess} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "13", "9", "0")
	requireJavaTLCUncovered(t, r)
}

// Original Github687fifoTest.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687fifo(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687fifo", "Github687fifo", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687fifo.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCSuccess} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "169", "49", "0")
	requireJavaTLCUncovered(t, r)
}

// Original Github687specATest.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687specA(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687spec", "Github687spec", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687specA.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCSuccess} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "21", "7", "0")
	requireJavaTLCUncovered(t, r)
}

// Original Github687specBTest.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687specB(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687spec", "Github687spec", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687specB.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCSuccess} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
}

// Original Github687specCTest.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687specC(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687spec", "Github687spec", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687specC.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCSuccess} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
}

// Original Github687specDTest.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687specD(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687spec", "Github687spec", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687specD.cfg")
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ErrorConfigParse; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCConfigSpecIsTrivial} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
}

// Original Github687specInitNextTest.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687specInitNext(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687spec", "Github687spec", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687specInitNext.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCSuccess} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
}

// Original Github687specPrimeTest.testSpec, noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub687specPrime(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github687spec", "Github687spec", true, true, true, 1,
		"-noGenerateSpecTE", "-config", "Github687specPrime.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
}
