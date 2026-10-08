package tlc

import (
	"fmt"
	"reflect"
	"testing"
)

// MultiFPSet.recover(TLCTrace) calls each selected child's recoverFP. No enabled
// original method tests that dispatch through an actual trace enumerator.
func TestDistributedNestedTraceRecoveryRetainsChildSemantics(t *testing.T) {
	for _, implementation := range []string{"tlc2.tool.fp.MSBDiskFPSet", "tlc2.tool.fp.OffHeapDiskFPSet"} {
		for _, warning := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/warning=%v", implementation, warning), func(t *testing.T) {
				t.Setenv(DiskFPSetError2WarningProperty, fmt.Sprint(warning))
				captureFailoverToolIO(t, ToolIOTool)
				directory := t.TempDir()
				trace := NewTLCTrace(directory, "Spec")
				defer trace.Close()
				for _, fp := range []uint64{41, uint64(1)<<63 | 43, 41, 97} {
					if err := trace.raf.WriteLongNat(1); err != nil {
						t.Fatal(err)
					}
					if err := trace.raf.WriteLong(int64(fp)); err != nil {
						t.Fatal(err)
					}
				}
				if err := trace.raf.Flush(); err != nil {
					t.Fatal(err)
				}
				config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
				config.SetMemory(1 << 20)
				config.SetFPBits(1)
				set := NewMultiFPSet(config)
				set.Init(1, directory, "fp")
				defer set.Close()
				recorder := &MemoryRecorder{}
				AddMessageRecorder(recorder)
				defer RemoveMessageRecorder(recorder)
				err := set.RecoverTrace(trace)
				if warning {
					if err != nil {
						t.Fatal(err)
					}
					if !set.Contains(97) || set.Size() != 3 || set.Sets[0].Size() != 2 || set.Sets[1].Size() != 1 {
						t.Fatal("warning recovery did not continue through selected children")
					}
					warnings := recorder.Records(ECSystemCheckpointRecoveryCorrupt)
					if len(warnings) != 1 || !reflect.DeepEqual(warnings[0].Params, []string{"Encountered duplicate fingerprint value 41"}) {
						t.Fatal("child duplicate warning was not preserved")
					}
				} else {
					failure, ok := err.(*TLCError)
					if !ok || failure.Code != ECSystemCheckpointRecoveryCorrupt || !failure.Runtime || !reflect.DeepEqual(failure.Params, []string{""}) {
						t.Fatalf("child recovery failure = %#v", err)
					}
					if set.Contains(97) || set.Size() != 2 || set.Sets[0].Size() != 1 || set.Sets[1].Size() != 1 {
						t.Fatal("failed recovery lost earlier inserts or continued after duplicate")
					}
					messages := recorder.Records(ECSystemCheckpointRecoveryCorrupt)
					if len(messages) != 1 || !messages[0].FormattingOnly || !reflect.DeepEqual(messages[0].Params, []string{""}) || len(ToolIOGetAllMessages()) != 0 {
						t.Fatal("child runtime failure did not retain formatting-only diagnostics")
					}
				}
				if !set.Contains(41) || !set.Contains(uint64(1)<<63|43) {
					t.Fatal("recovery changed partition routing")
				}
			})
		}
	}
}

type nestedRecoveryChild struct {
	*MemFPSet
	failure error
	seen    []uint64
}

func (s *nestedRecoveryChild) RecoverFP(fp uint64) error {
	s.seen = append(s.seen, fp)
	if s.failure != nil {
		return s.failure
	}
	return s.MemFPSet.RecoverFP(fp)
}

func TestDistributedNestedRecoveryChildIOFailurePrecedesPublication(t *testing.T) {
	queue := javaLongDiskStateQueueSetup(t)
	queue.Enqueue(&TLCStateMut{UID: 0, level: 1})
	if err := queue.BeginChkpt(); err != nil {
		t.Fatal(err)
	}
	if err := queue.CommitChkpt(); err != nil {
		t.Fatal(err)
	}
	trace := NewTLCTrace(queue.diskdir, "Spec")
	defer trace.Close()
	for _, fp := range []uint64{41, uint64(1)<<63 | 43, 97} {
		if err := trace.raf.WriteLongNat(1); err != nil {
			t.Fatal(err)
		}
		if err := trace.raf.WriteLong(int64(fp)); err != nil {
			t.Fatal(err)
		}
	}
	if err := trace.BeginChkpt(); err != nil {
		t.Fatal(err)
	}
	if err := trace.CommitChkpt(); err != nil {
		t.Fatal(err)
	}
	failure := NewIOException("child recovery disk failure")
	first := &nestedRecoveryChild{MemFPSet: NewMemFPSet()}
	second := &nestedRecoveryChild{MemFPSet: NewMemFPSet(), failure: failure}
	set := &MultiFPSet{Sets: []FPSet{first, second}, FPBits: 1, Shift: 63}
	manager := NewNonDistributedFPSetManager(set, "local", trace)
	server := NewTLCServer("Spec", "Spec", queue.diskdir, manager, queue, trace)
	server.SetTool(NewTool())
	server.app.fromCheckpoint = &queue.diskdir
	publicationCalls := 0
	server.ConfigurePublication(TLCServerPublication{LocalHostName: func() (string, error) { publicationCalls++; return "unexpected", nil }})
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	code, err := server.ModelCheck()
	if code != ECGeneral || err != failure {
		t.Fatalf("child IO cause = %d/%v", code, err)
	}
	if !reflect.DeepEqual(first.seen, []uint64{41}) || !reflect.DeepEqual(second.seen, []uint64{uint64(1)<<63 | 43}) || set.Size() != 1 || !set.Contains(41) || set.Contains(97) {
		t.Fatal("child failure changed partition dispatch or partial recovery")
	}
	if queue.Size() != 1 || publicationCalls != 0 || server.IsDone() || !recorder.Recorded(ECTLCCheckpointRecoverStart) || recorder.Recorded(ECTLCCheckpointRecoverEnd) || recorder.Recorded(ECTLCComputingInit) || recorder.Recorded(ECGeneral) {
		t.Fatal("child failure changed trace/queue recovery or escaped its startup boundary")
	}
}
