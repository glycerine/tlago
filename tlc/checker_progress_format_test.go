package tlc

import (
	"reflect"
	"testing"
	"time"
)

func TestModelCheckerProgressUsesJavaNumberFormatting(t *testing.T) {
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() { RemoveMessageRecorder(recorder) })

	set := NewMemFPSet()
	queue := NewMemStateQueue()
	for fp := range uint64(1234) {
		set.Put(fp)
		queue.Enqueue(&TLCStateMut{UID: TLCStateInitUID, level: 1})
	}
	checker := &ModelChecker{
		AbstractChecker:       &AbstractChecker{},
		NumberOfInitialStates: 1234,
		FPSet:                 set,
		StateQueue:            queue,
		ConcurrentTrace:       &ConcurrentTLCTrace{TLCTrace: NewTLCTrace()},
	}
	checker.PrintProgressStats(time.Time{}, false)

	records := recorder.Records(ECTLCProgressStats)
	if len(records) != 1 {
		t.Fatalf("progress records = %d, want 1", len(records))
	}
	want := []string{
		"1",
		MessageNumberFormat(1234),
		MessageNumberFormat(1234),
		MessageNumberFormat(1234),
		MessageNumberFormat(checker.StatesPerMinute),
		MessageNumberFormat(checker.DistinctStatesPerMinute),
	}
	if !reflect.DeepEqual(records[0].Params, want) {
		t.Fatalf("progress params = %q, want %q", records[0].Params, want)
	}
}
