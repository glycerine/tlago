package tlc

import "testing"

// No direct original test exercises a negative TLCTrace.getTrace location.
func TestDiskTraceNegativeLocationPropagatesIOFailure(t *testing.T) {
	for _, included := range []bool{false, true} {
		name := "predecessor"
		if included {
			name = "included"
		}
		t.Run(name, func(t *testing.T) {
			trace := NewTLCTrace(t.TempDir(), "Spec")
			t.Cleanup(func() { _ = trace.Close() })
			var recovered []*TLCStateInfo
			err := invokeDistributedServerOperation(func() error {
				recovered = trace.GetTraceAt(-1, included)
				return nil
			})
			if !isJavaIOException(err) || recovered != nil {
				t.Fatalf("negative location = %v, error=%T/%v; want I/O failure", recovered, err, err)
			}
			position, pointerErr := trace.raf.GetFilePointer()
			if pointerErr != nil || position != -1 {
				t.Fatalf("failed seek pointer = %d/%v, want source mutation -1", position, pointerErr)
			}
		})
	}
}

func TestDistributedNegativeTraceLocationRetainsFailureCatch(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() { RemoveMessageRecorder(recorder) })
	trace := NewTLCTrace(t.TempDir(), "Spec")
	t.Cleanup(func() { _ = trace.Close() })
	queue := NewMemStateQueue()
	server := &TLCServer{Trace: trace, StateQueue: queue}
	state := &TLCStateMut{UID: -1, level: TLCStateInitLevel}
	failure := NewWorkerException("worker evaluation failed", state, nil, false)
	thread := &TLCServerThread{Server: server}
	thread.handleRunError(failure, queue)
	if !server.IsDone() || server.LastError != failure || server.ErrState != state || !queue.finish.Load() {
		t.Fatal("trace I/O failure replaced worker error or skipped coordinator completion")
	}
	if len(recorder.Records(ECGeneral)) != 1 || recorder.Recorded(ECTLCBehaviorUpToThisPoint) || recorder.Recorded(ECTLCStatePrint2) || recorder.Recorded(ECTLCFailedToRecoverInit) {
		t.Fatal("negative location entered empty-prefix trace printing or reconstruction failure")
	}
}
