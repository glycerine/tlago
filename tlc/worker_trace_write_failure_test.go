package tlc

import "testing"

// No original method directly covers partial writes in Worker.writeState.
func TestWorkerTracePartialWriteRetainsMutationOrder(t *testing.T) {
	for _, kind := range []string{"initial", "successor"} {
		for _, phase := range []string{"predecessor", "worker", "fingerprint"} {
			t.Run(kind+"/"+phase, func(t *testing.T) {
				worker := NewWorker(2)
				worker.SetTraceContext(t.TempDir(), "Spec")
				defer worker.CloseTrace()
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				offset := int64(BufferedRandomAccessFileBuffSz - 2)
				if phase == "worker" {
					offset = BufferedRandomAccessFileBuffSz - 5
				}
				if phase == "fingerprint" {
					offset = BufferedRandomAccessFileBuffSz - 9
				}
				if err := worker.traceRAF.WriteFull(make([]byte, offset)); err != nil {
					t.Fatal(err)
				}
				if err := worker.traceRAF.Flush(); err != nil {
					t.Fatal(err)
				}
				if err := worker.traceRAF.file.Close(); err != nil {
					t.Fatal(err)
				}
				worker.lastPtr = 1
				worker.SetLevel(3)
				worker.UnseenSuccessorStates = 17
				mirror := NewTLCTrace()
				worker.Checker = &ModelChecker{Trace: mirror}
				oldParent, oldAction := &TLCStateMut{level: 2}, &Action{}
				state := &TLCStateMut{UID: 314, WorkerID: 7, level: 5, pred: oldParent, action: oldAction}
				parent := &TLCStateMut{UID: 37, WorkerID: 2, level: 8}
				var err error
				if kind == "initial" {
					err = worker.WriteInitState(state, 0x0102030405060708)
				} else {
					err = worker.WriteNextState(parent, 0x0102030405060708, state, &Action{})
				}
				if !isJavaIOException(err) {
					t.Fatalf("write failure %T/%v", err, err)
				}
				if worker.lastPtr != offset {
					t.Fatalf("last record pointer %d, want %d", worker.lastPtr, offset)
				}
				wantLevel := 3
				if kind == "successor" {
					wantLevel = 9
				}
				if worker.GetMaxLevel() != wantLevel || worker.UnseenSuccessorStates != 17 {
					t.Fatalf("failed write level/count %d/%d", worker.GetMaxLevel(), worker.UnseenSuccessorStates)
				}
				if state.UID != 314 || state.WorkerID != 7 || state.Level() != 5 || state.TracePredecessor() != oldParent || state.GetAction() != oldAction || len(mirror.Records()) != 0 {
					t.Fatal("failed write mutated state or published mirror")
				}
				if cursor, err := worker.traceRAF.GetFilePointer(); err != nil || cursor != BufferedRandomAccessFileBuffSz {
					t.Fatalf("failed write cursor %d/%v", cursor, err)
				}
				if phase != "predecessor" {
					predecessor := byte(1)
					if kind == "successor" {
						predecessor = 37
					}
					if worker.traceRAF.buff[offset+3] != predecessor || worker.traceRAF.buff[offset+4] != 2 {
						t.Fatal("failure lost completed predecessor/worker bytes")
					}
				}
				if phase == "fingerprint" && (worker.traceRAF.buff[offset+5] != 1 || worker.traceRAF.buff[offset+8] != 4) {
					t.Fatal("failure lost partial fingerprint")
				}
			})
		}
	}
}

func TestWorkerTraceWritePreservesGeneratedActionAndMetadataPolicy(t *testing.T) {
	oldPolicy := statePreserveMetadata
	defer func() { statePreserveMetadata = oldPolicy }()
	for _, preserve := range []bool{false, true} {
		for _, mirrored := range []bool{false, true} {
			statePreserveMetadata = preserve
			worker := NewWorker(2)
			worker.SetTraceContext(t.TempDir(), "Spec")
			generatedAction := &Action{Name: "Generated"}
			state := &TLCStateMut{UID: 314, WorkerID: 7, level: 5, action: generatedAction}
			parent := &TLCStateMut{UID: 37, WorkerID: 1, level: 8}
			if mirrored {
				worker.Checker = &ModelChecker{Trace: NewTLCTrace()}
			}
			if err := worker.WriteNextState(parent, 99, state, &Action{Name: "WriterArgument"}); err != nil {
				t.Fatal(err)
			}
			if state.UID != 0 || state.WorkerID != 2 || state.Level() != 9 || state.GetAction() != generatedAction {
				t.Fatalf("writer changed generated metadata: preserve=%v, mirror=%v", preserve, mirrored)
			}
			if preserve && state.TracePredecessor() != parent || !preserve && state.TracePredecessor() != nil {
				t.Fatalf("writer/mirror changed predecessor policy: preserve=%v, mirror=%v", preserve, mirrored)
			}
			if worker.GetMaxLevel() != 9 || worker.UnseenSuccessorStates != 1 {
				t.Fatal("successful write lost level/count updates")
			}
			if err := worker.CloseTrace(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
