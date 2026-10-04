/*******************************************************************************
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
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
 ******************************************************************************/
package tlc

import (
	"bytes"
	"testing"
)

// Full original WarningControlTest. The Go error return represents Java's
// boolean handleParameters result; errors are not swallowed in runtime tests.
func TestJavaWarningControl(t *testing.T) {
	t.Chdir(t.TempDir()) // Isolate the source's relative MC metadata directories.
	Globals.Lock()
	oldWarn, oldTool, oldMeta := Globals.Warn, Globals.Tool, Globals.MetaDir
	oldTLC, oldElevated := Globals.SuppressedMessages, Globals.MessagesAsErrors
	oldSANY, oldSANYElevated := Globals.SANYSuppressedMessages, Globals.SANYMessagesAsErrors
	Globals.Warn, Globals.Tool = true, false
	Globals.SuppressedMessages, Globals.MessagesAsErrors = NewInsMap[int, bool](), NewInsMap[int, bool]()
	Globals.SANYSuppressedMessages, Globals.SANYMessagesAsErrors = NewInsMap[int, bool](), NewInsMap[int, bool]()
	Globals.Unlock()
	mpConsole.Lock()
	oldHistory := mpConsole.warningHistory
	mpConsole.warningHistory = make(map[string]bool)
	mpConsole.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.Warn, Globals.Tool, Globals.MetaDir = oldWarn, oldTool, oldMeta
		Globals.SuppressedMessages, Globals.MessagesAsErrors = oldTLC, oldElevated
		Globals.SANYSuppressedMessages, Globals.SANYMessagesAsErrors = oldSANY, oldSANYElevated
		Globals.Unlock()
		mpConsole.Lock()
		mpConsole.warningHistory = oldHistory
		mpConsole.Unlock()
	})
	run := func(name string, body func(*testing.T)) {
		t.Run(name, func(t *testing.T) {
			// Original @After runs even when an assertion/expected exception fails.
			t.Cleanup(func() { ResetMessageControl(); Globals.Lock(); Globals.Warn = true; Globals.Unlock() })
			body(t)
		})
	}
	run("testSuppressMessagesSanyCode", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-suppressMessages", "4802", "MC"}) == nil)
		if !GetSANYSuppressedCodes().Contains(4802) {
			t.Fatal("assertTrue: GetSANYSuppressedCodes().Contains(4802)")
		}
	})
	run("testSuppressMessagesSanyErrorCodeFails", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-suppressMessages", "4200", "MC"}) == nil)
		if !GetSANYSuppressedCodes().IsEmpty() {
			t.Fatal("assertTrue: GetSANYSuppressedCodes().IsEmpty()")
		}
	})
	run("testSuppressMessagesTlcCode", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-suppressMessages", "2156", "MC"}) == nil)
		if !GetTLCSuppressedCodes().Contains(2156) {
			t.Fatal("assertTrue: GetTLCSuppressedCodes().Contains(2156)")
		}
	})
	run("testSuppressMessagesMultipleCodes", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-suppressMessages", "2156,4802", "MC"}) == nil)
		if !GetTLCSuppressedCodes().Contains(2156) {
			t.Fatal("assertTrue: GetTLCSuppressedCodes().Contains(2156)")
		}
		if !GetSANYSuppressedCodes().Contains(4802) {
			t.Fatal("assertTrue: GetSANYSuppressedCodes().Contains(4802)")
		}
	})
	run("testSuppressMessagesUnknownCodeFails", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-suppressMessages", "9999999", "MC"}) == nil)
		if !GetTLCSuppressedCodes().IsEmpty() {
			t.Fatal("assertTrue: GetTLCSuppressedCodes().IsEmpty()")
		}
		if !GetSANYSuppressedCodes().IsEmpty() {
			t.Fatal("assertTrue: GetSANYSuppressedCodes().IsEmpty()")
		}
	})
	run("testSuppressMessagesMissingArgFails", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-suppressMessages", "MC"}) == nil)
	})
	run("testMessagesAsErrorsSanyWarningCode", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-messagesAsErrors", "4802", "MC"}) == nil)
		if !GetSANYMessagesAsErrorCodes().Contains(4802) {
			t.Fatal("assertTrue: GetSANYMessagesAsErrorCodes().Contains(4802)")
		}
	})
	run("testMessagesAsErrorsTlcCode", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-messagesAsErrors", "2156", "MC"}) == nil)
		if !GetTLCMessagesAsErrorCodes().Contains(2156) {
			t.Fatal("assertTrue: GetTLCMessagesAsErrorCodes().Contains(2156)")
		}
	})
	run("testMessagesAsErrorsUnknownCodeFails", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-messagesAsErrors", "9999999", "MC"}) == nil)
	})
	run("testMessagesAsErrorsMissingArgFails", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-messagesAsErrors", "MC"}) == nil)
	})
	run("testNowarningConflictWithSuppressMessages", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-nowarning", "-suppressMessages", "4802", "MC"}) == nil)
	})
	run("testNowarningConflictWithmessagesAsErrors", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-nowarning", "-messagesAsErrors", "4802", "MC"}) == nil)
	})
	run("testSameTlcCode", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-suppressMessages", "2156", "-messagesAsErrors", "2156", "MC"}) == nil)
	})
	run("testSameSanyCode", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-suppressMessages", "4802", "-messagesAsErrors", "4802", "MC"}) == nil)
	})
	run("testRuntimeSuppressedWarningProducesNoOutput", func(t *testing.T) {
		var captured bytes.Buffer
		restore := ToolIOSetSystemStreams(&captured, &captured)
		defer restore()
		SuppressTLCMessage(ECTLCFeatureUnsupported)
		PrintWarning(ECTLCFeatureUnsupported, "suppression-test")
		if captured.String() != "" {
			t.Fatal("Suppressed warning should produce no output to ToolIO.out")
		}
	})
	run("testRuntimeUnsuppressedWarningProducesOutput", func(t *testing.T) {
		var captured bytes.Buffer
		restore := ToolIOSetSystemStreams(&captured, &captured)
		defer restore()
		PrintWarning(ECTLCFeatureUnsupported, "unsuppressed-test")
		if captured.String() == "" {
			t.Fatal("Unsuppressed warning should produce output to ToolIO.out")
		}
	})
	run("testRuntimeWarningAsErrorThrowsTLCRuntimeException", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected TLCRuntimeException")
			}
			switch e := r.(type) {
			case *TLCError:
				if !e.Runtime {
					panic(r)
				}
			default:
				panic(r)
			}
		}()
		TreatTLCMessageAsError(ECTLCFeatureUnsupported)
		PrintWarning(ECTLCFeatureUnsupported, "elevation-test")
	})
	run("testRuntimeWarningWithoutElevationDoesNotThrow", func(t *testing.T) {
		PrintWarning(ECTLCFeatureUnsupported, "no-elevation-test")
	})
}
