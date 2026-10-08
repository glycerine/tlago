package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Native publication integration has no direct upstream test. Use the existing
// initialization-error model to exercise actual modelCheck ordering and its
// deliberate early return before worker publication and normal unbind cleanup.
func TestNativeCoordinatorModelInitPublication(t *testing.T) {
	oldCheckpoint, oldMeta, oldPort := tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir, tlc.TLCServerPort()
	tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir = 0, t.TempDir()
	tlc.SetTLCServerPort(0)
	t.Cleanup(func() {
		tlc.Globals.CheckpointDurationMillis, tlc.Globals.MetaDir = oldCheckpoint, oldMeta
		tlc.SetTLCServerPort(oldPort)
	})
	directory, err := filepath.Abs("tlc/test_vectors/distributed/DoInitFunctor")
	if err != nil {
		t.Fatal(err)
	}
	oldDirectory := tlc.GetFilenameUserDirectory()
	tlc.SetFilenameUserDirectory(&directory)
	t.Cleanup(func() { tlc.SetFilenameUserDirectory(oldDirectory) })
	config := tlc.NewFPSetConfigurationWithRatioAndImplementation(1, "tlc2.tool.fp.MSBDiskFPSet")
	config.SetMemory(1 << 20)
	name := filepath.Join(directory, "DoInitFunctorEvalException")
	app, diags, err := LoadTLCAppWithMetadata(name, name, false, nil, config, nil, tlc.RuntimeParameters{})
	requireNoErrors(t, diags)
	if err != nil {
		t.Fatal(err)
	}
	server, err := tlc.NewTLCServerFromApp(app)
	if err != nil {
		t.Fatal(err)
	}
	network := tlc.NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	t.Cleanup(func() {
		if err := network.Close(); err != nil {
			t.Error(err)
		}
	})
	server.ConfigurePublication(network.Publication())
	recorder := &tlc.MemoryRecorder{}
	tlc.AddMessageRecorder(recorder)
	t.Cleanup(func() { tlc.RemoveMessageRecorder(recorder) })
	if _, err := server.ModelCheck(); err != nil {
		t.Fatal(err)
	}
	if !recorder.Recorded(tlc.ECTLCInitialState) || !recorder.Recorded(tlc.ECTLCFinished) || recorder.Recorded(tlc.ECGeneral) {
		t.Fatal("native publication changed initialization-error handling")
	}
	discovery := tlc.NewDistributedNetworkDiscovery()
	defer discovery.Close()
	coordinator, err := discovery.Lookup("tcp://" + network.Address + "/" + tlc.TLCServerName)
	if err != nil {
		t.Fatalf("early return removed coordinator binding: %v", err)
	}
	if done, err := coordinator.IsDone(); err != nil || !done {
		t.Fatalf("early init completion = %v/%v", done, err)
	}
	if _, err := discovery.Lookup("tcp://" + network.Address + "/" + tlc.TLCServerWorkerName); err == nil {
		t.Fatal("failed initialization published worker binding")
	}
	if err := network.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := coordinator.IsDone(); err == nil {
		t.Fatal("process owner failed to close retained binding transport")
	}
}
