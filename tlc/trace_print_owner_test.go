package tlc

import "testing"

// No original method directly covers missing trace-print entry arguments.
func TestTracePrintingMissingOwnerPreservesHeaderBoundary(t *testing.T) {
	for _, operation := range []string{"ordinary", "concurrent", "shared"} {
		for _, scenario := range []string{"missing-state", "missing-trace", "both"} {
			t.Run(operation+"/"+scenario, func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				recorder := &MemoryRecorder{}
				AddMessageRecorder(recorder)
				defer RemoveMessageRecorder(recorder)
				trace := NewTLCTrace()
				concurrent := &ConcurrentTLCTrace{TLCTrace: trace}
				state := &TLCStateMut{UID: 314, WorkerID: 7, level: TLCStateInitLevel}
				if scenario != "missing-trace" {
					state = nil
				}
				if scenario != "missing-state" {
					trace, concurrent = nil, nil
				}
				err := invokeDistributedServerOperation(func() error {
					switch operation {
					case "ordinary":
						trace.PrintTrace(state, nil)
					case "concurrent":
						concurrent.PrintTrace(state, nil)
					case "shared":
						trace.printTraceWithPrefix(state, nil, nil)
					}
					return nil
				})
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing print owner failure: %T/%v", err, err)
				}
				if recorder.Recorded(ECTLCBehaviorUpToThisPoint) || recorder.Recorded(ECTLCStatePrint2) || recorder.Recorded(ECTLCStatePrint1) {
					t.Fatal("missing owner changed source header/state output order")
				}
			})
		}
	}
}
