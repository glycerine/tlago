/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Port of DistributedDoInitFunctorEvalExceptionTest.testSpec, after the server
// init callback and modelCheck catch/replay path have been implemented. Its
// TLCServerTestCase harness uses a disk FPSet whose exit does not stop the JVM;
// Go's native FPSet exit already leaves the host process running.
func TestDistributedDoInitFunctorEvalException(t *testing.T) {
	oldCheckpoint := tlc.Globals.CheckpointDurationMillis
	oldMeta := tlc.Globals.MetaDir
	tlc.Globals.CheckpointDurationMillis = 0
	tlc.Globals.MetaDir = t.TempDir()
	t.Cleanup(func() { tlc.Globals.CheckpointDurationMillis = oldCheckpoint; tlc.Globals.MetaDir = oldMeta })
	name := "DoInitFunctorEvalException"
	directory, err := filepath.Abs("tlc/test_vectors/distributed/DoInitFunctor")
	if err != nil {
		t.Fatal(err)
	}
	oldUserDirectory := tlc.GetFilenameUserDirectory()
	tlc.SetFilenameUserDirectory(&directory)
	t.Cleanup(func() { tlc.SetFilenameUserDirectory(oldUserDirectory) })
	config := tlc.NewFPSetConfigurationWithRatioAndImplementation(1, "tlc2.tool.fp.MSBDiskFPSet")
	config.SetMemory(1 << 20) // Bound the native harness allocation.
	app, diags, err := LoadTLCAppWithMetadata(filepath.Join(directory, name), filepath.Join(directory, name), false, nil, config, nil, tlc.RuntimeParameters{})
	requireNoErrors(t, diags)
	if err != nil {
		t.Fatal(err)
	}
	server, err := tlc.NewTLCServerFromApp(app)
	if err != nil {
		t.Fatal(err)
	}
	// The Java Ant harness forks per test. Give this translated test its own
	// naming namespace; early-done modelCheck deliberately retains its registry.
	server.ConfigurePublication(tlc.NewTLCRegistryNamespace().Publication())
	recorder := &tlc.MemoryRecorder{}
	tlc.AddMessageRecorder(recorder)
	t.Cleanup(func() { tlc.RemoveMessageRecorder(recorder) })
	if _, err := server.ModelCheck(); err != nil {
		t.Fatal(err)
	}
	if !recorder.Recorded(tlc.ECTLCFinished) {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if !recorder.Recorded(tlc.ECTLCStats) {
		t.Fatal("TLC_STATS not recorded")
	}
	if recorder.Recorded(tlc.ECGeneral) {
		t.Fatal("GENERAL recorded")
	}
	want := []string{"TLC expected a boolean value, but did not find one. line 15, col 15 to line 15, col 18 of module DoInitFunctorEvalException", "x = 1\n"}
	for _, message := range recorder.Records(tlc.ECTLCInitialState) {
		if reflect.DeepEqual(message.Params, want) {
			return
		}
	}
	t.Fatalf("TLC_INITIAL_STATE = %v, want parameters %q", recorder.Records(tlc.ECTLCInitialState), want)
}

// Original TLCSetTest.testSpec and TLCServerTestCase setup. Native FPSet.Exit
// leaves the Go process alive, matching the Java dummy's suppressed System.exit.
func TestJavaDistributedTLCSet(t *testing.T) {
	oldCheckpoint, oldMeta := tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir
	oldMain, oldSim, oldWorkers := tlc.Globals.MainChecker, tlc.Globals.Simulator, tlc.Globals.NumWorkers
	tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir = 0, t.TempDir()
	tlc.SetMainChecker(nil)
	tlc.SetSimulator(nil)
	tlc.Globals.NumWorkers = 1
	t.Cleanup(func() {
		tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir = oldCheckpoint, oldMeta
		tlc.SetMainChecker(oldMain)
		tlc.SetSimulator(oldSim)
		tlc.Globals.NumWorkers = oldWorkers
	})
	directory, err := filepath.Abs("tlc/test_vectors/models/TLCSet")
	if err != nil {
		t.Fatal(err)
	}
	oldDirectory := tlc.GetFilenameUserDirectory()
	tlc.SetFilenameUserDirectory(&directory)
	t.Cleanup(func() { tlc.SetFilenameUserDirectory(oldDirectory) })
	config := tlc.NewFPSetConfigurationWithRatioAndImplementation(tlc.NewFPSetConfiguration().GetRatio(), "tlc2.tool.fp.MSBDiskFPSet")
	app, diags, err := LoadTLCAppWithMetadata(filepath.Join(directory, "TLCSet"), "TLCSet", false, nil, config, nil, tlc.RuntimeParameters{})
	requireNoErrors(t, diags)
	if err != nil {
		t.Fatal(err)
	}
	server, err := tlc.NewTLCServerFromApp(app)
	if err != nil {
		t.Fatal(err)
	}
	server.ConfigurePublication(tlc.NewTLCRegistryNamespace().Publication())
	recorder := &tlc.MemoryRecorder{}
	tlc.AddMessageRecorder(recorder)
	defer tlc.RemoveMessageRecorder(recorder)
	if _, err := server.ModelCheck(); err != nil {
		t.Fatal(err)
	}
	for _, code := range []int{tlc.ECTLCComputingInit, tlc.ECTLCFeatureUnsupported} {
		if !recorder.Recorded(code) {
			t.Fatalf("event %d absent", code)
		}
	}
	if recorder.Recorded(tlc.ECGeneral) {
		t.Fatal("GENERAL recorded")
	}
	// Java harness explicitly assigns actualExitStatus=0 after server.modelCheck.
}
