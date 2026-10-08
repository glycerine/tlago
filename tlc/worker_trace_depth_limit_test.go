package tlc

import (
	"math"
	"testing"
)

// No original method directly covers Worker.writeState's int overflow and
// extended predecessor mutation before a trace-depth failure.
func TestWorkerTraceDepthLimitMutationOrder(t *testing.T) {
	initTLCCheckerTest(t)
	oldPolicy := statePreserveMetadata
	defer func() { statePreserveMetadata = oldPolicy }()
	for _, preserve := range []bool{false, true} {
		for _, phase := range []string{"normal", "largest-success", "metadata-failure", "cursor-failure"} {
			t.Run(func() string {
				if preserve {
					return "extended/" + phase
				}
				return "plain/" + phase
			}(), func(t *testing.T) {
				statePreserveMetadata = preserve
				worker := NewWorker(2)
				worker.SetTraceContext(t.TempDir(), "Spec")
				defer worker.CloseTrace()
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				worker.SetLevel(3)
				worker.lastPtr = 1
				worker.UnseenSuccessorStates = 17
				mirror := NewTLCTrace()
				worker.Checker = &ModelChecker{Trace: mirror}
				oldParent, generatedAction := checkerTestState(0), &Action{Name: "Generated"}
				state := checkerTestState(2)
				state.UID, state.WorkerID, state.level, state.pred, state.action = 314, 7, 5, oldParent, generatedAction
				parent := checkerTestState(1)
				parent.UID, parent.WorkerID, parent.level = 37, 1, 7
				if phase == "largest-success" {
					parent.level = math.MaxInt32 - 1
				}
				if phase == "metadata-failure" || phase == "cursor-failure" {
					parent.level = math.MaxInt32
				}
				if phase == "cursor-failure" {
					if err := worker.traceRAF.Close(); err != nil {
						t.Fatal(err)
					}
				}
				err := invokeDistributedServerOperation(func() error { return worker.WriteNextState(parent, 99, state, &Action{Name: "WriterArgument"}) })
				switch phase {
				case "cursor-failure":
					if !isJavaIOException(err) || worker.GetMaxLevel() != 3 || worker.lastPtr != 1 || state.UID != 314 || state.WorkerID != 7 || state.TracePredecessor() != oldParent {
						t.Fatalf("cursor failure mutation order: %v, max %d, ptr %d", err, worker.GetMaxLevel(), worker.lastPtr)
					}
				case "metadata-failure":
					failure, ok := err.(*TLCError)
					if !ok || failure.Code != ECTLCTraceTooLong {
						t.Fatalf("depth failure: %T/%v", err, err)
					}
					if worker.GetMaxLevel() != 3 || worker.lastPtr != 0 || state.UID != 0 || state.WorkerID != 2 || state.Level() != 5 {
						t.Fatalf("depth failure metadata/max: max %d, ptr %d, state %d/%d/%d", worker.GetMaxLevel(), worker.lastPtr, state.UID, state.WorkerID, state.Level())
					}
					wantParent := oldParent
					if preserve {
						wantParent = parent
					}
					if state.TracePredecessor() != wantParent {
						t.Fatal("extended predecessor assignment did not precede base depth failure")
					}
					record, readErr := worker.ReadStateRecord(0)
					if readErr != nil || record.Ptr != 37 || record.Worker != 1 || record.FP != 99 {
						t.Fatalf("depth failure lost completed record: %v/%v", record, readErr)
					}
				default:
					want := parent.Level() + 1
					if err != nil || worker.GetMaxLevel() != want || state.Level() != want || state.UID != 0 || state.WorkerID != 2 || worker.UnseenSuccessorStates != 18 || len(mirror.Records()) != 1 {
						t.Fatalf("successful depth %d: %v, max %d", parent.Level(), err, worker.GetMaxLevel())
					}
				}
				if state.GetAction() != generatedAction {
					t.Fatal("depth handling changed generated action")
				}
				if phase == "cursor-failure" || phase == "metadata-failure" {
					if worker.UnseenSuccessorStates != 17 || len(mirror.Records()) != 0 {
						t.Fatal("failed write published unseen count or mirror")
					}
				}
				if !worker.mu.TryLock() {
					t.Fatal("depth failure retained worker monitor")
				}
				worker.mu.Unlock()
			})
		}
	}
}
