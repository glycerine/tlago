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
