package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateWriterNullStateRetainsJavaHeaderAndNumber(t *testing.T) {
	writer := newTestPlainStateWriter(t)
	for attempt, want := range []string{"State 1:\n", "State 1:\nState 1:\n"} {
		var failure any
		func() {
			defer func() { failure = recover() }()
			_ = writer.WriteInitState(nil)
		}()
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("attempt %d failure = %T: %v, want NullPointerException", attempt+1, failure, failure)
		}
		if got := readPlainStateWriterBytes(t, writer); got != want {
			t.Fatalf("attempt %d bytes = %q, want %q", attempt+1, got, want)
		}
	}
}

func TestStateWriterTransitionDereferencesNullSuccessorOnlyWhenUnseen(t *testing.T) {
	writer := newTestPlainStateWriter(t)
	for _, status := range []StateVisitStatus{StateVisitSeen, StateVisitSeen | StateVisitNotInModel} {
		if err := writer.WriteTransitionVisual(nil, nil, status, nil, StateVisualizationDefault); err != nil {
			t.Fatalf("seen status %d null successor = %v, want nil", status, err)
		}
	}
	if got := readPlainStateWriterBytes(t, writer); got != "" {
		t.Fatalf("seen null successor bytes = %q, want empty", got)
	}

	var failure any
	func() {
		defer func() { failure = recover() }()
		_ = writer.WriteTransitionVisual(nil, nil, StateVisitUnseen, nil, StateVisualizationDefault)
	}()
	if _, ok := failure.(*NullPointerException); !ok {
		t.Fatalf("unseen null successor failure = %T: %v, want NullPointerException", failure, failure)
	}
	if got := readPlainStateWriterBytes(t, writer); got != "State 1:\n" {
		t.Fatalf("unseen null successor bytes = %q, want state header", got)
	}
}

func TestNoopStateWriterStillIgnoresNullStates(t *testing.T) {
	writer := NewNoopStateWriter()
	if err := writer.WriteInitState(nil); err != nil {
		t.Fatalf("init = %v, want nil", err)
	}
	if err := writer.WriteTransitionVisual(nil, nil, StateVisitUnseen, nil, StateVisualizationDefault); err != nil {
		t.Fatalf("transition = %v, want nil", err)
	}
}

func TestStateWriterNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var writer *StateWriter
	for name, call := range map[string]func(){
		"write_init":       func() { _ = writer.WriteInitState(nil) },
		"write_transition": func() { _ = writer.WriteTransition(nil, nil, StateVisitSeen, nil) },
		"write_visual":     func() { _ = writer.WriteTransitionVisual(nil, nil, StateVisitSeen, nil, StateVisualizationDefault) },
		"is_constrained":   func() { _ = writer.IsConstrained() },
		"is_noop":          func() { _ = writer.IsNoop() },
		"is_dot":           func() { _ = writer.IsDot() },
		"dump_filename":    func() { _ = writer.GetDumpFileName() },
		"snapshot":         func() { _ = writer.Snapshot() },
		"close":            func() { _ = writer.Close() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}
}

func newTestPlainStateWriter(t *testing.T) *StateWriter {
	t.Helper()
	writer, err := NewStateDumpWriter(filepath.Join(t.TempDir(), "states.out"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = writer.Close() })
	return writer
}

func readPlainStateWriterBytes(t *testing.T, writer *StateWriter) string {
	t.Helper()
	writer.writer.Flush()
	data, err := os.ReadFile(writer.GetDumpFileName())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
