package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLivenessStateWriterNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var writer *LivenessStateWriter
	for name, call := range map[string]func(){
		"is_noop":          func() { _ = writer.IsNoop() },
		"is_dot":           func() { _ = writer.IsDot() },
		"is_constrained":   func() { _ = writer.IsConstrained() },
		"dump_filename":    func() { _ = writer.GetDumpFileName() },
		"snapshot":         func() { _ = writer.Snapshot() },
		"close":            func() { _ = writer.Close() },
		"interface_state":  func() { _ = writer.WriteInitState(nil) },
		"interface_edge":   func() { _ = writer.WriteTransition(nil, nil, StateVisitSeen, nil) },
		"interface_visual": func() { _ = writer.WriteTransitionVisual(nil, nil, StateVisitSeen, nil, StateVisualizationDefault) },
		"state":            func() { _ = writer.WriteLivenessState(nil) },
		"state_transition": func() {
			_ = writer.WriteLivenessStateTransition(nil, nil, nil, 0, 0, StateVisitUnseen, StateVisualizationDefault)
		},
		"init_state": func() { _ = writer.WriteLivenessInitState(nil, nil) },
		"transition": func() { _ = writer.WriteLivenessTransition(nil, nil, nil, nil, nil, 0, 0, StateVisitUnseen) },
		"transition_visual": func() {
			_ = writer.WriteLivenessTransitionVisual(nil, nil, nil, nil, nil, 0, 0, StateVisitUnseen, StateVisualizationDefault)
		},
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

func TestDotLivenessStateWriterRejectsNullWriterLikeJava(t *testing.T) {
	for name, input := range map[string]IStateWriter{
		"null_interface": nil,
		"typed_null":     (*StateWriter)(nil),
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			_, _ = NewDotLivenessStateWriter(input)
		})
	}
}

func TestNoopLivenessStateWriterRetainsNullArgumentNoOps(t *testing.T) {
	writer := NewNoopLivenessStateWriter()
	if !writer.IsNoop() || writer.IsDot() || writer.IsConstrained() || writer.GetDumpFileName() != "" {
		t.Fatalf("no-op identity = noop %t, dot %t, constrained %t, file %q", writer.IsNoop(), writer.IsDot(), writer.IsConstrained(), writer.GetDumpFileName())
	}
	for name, call := range map[string]func() error{
		"close":           writer.Close,
		"snapshot":        writer.Snapshot,
		"interface_state": func() error { return writer.WriteInitState(nil) },
		"interface_edge":  func() error { return writer.WriteTransition(nil, nil, StateVisitUnseen, nil) },
		"interface_visual": func() error {
			return writer.WriteTransitionVisual(nil, nil, StateVisitUnseen, nil, StateVisualizationDefault)
		},
		"state": func() error { return writer.WriteLivenessState(nil) },
		"state_transition": func() error {
			return writer.WriteLivenessStateTransition(nil, nil, nil, 0, 0, StateVisitUnseen, StateVisualizationDefault)
		},
		"init_state": func() error { return writer.WriteLivenessInitState(nil, nil) },
		"transition": func() error { return writer.WriteLivenessTransition(nil, nil, nil, nil, nil, 0, 0, StateVisitUnseen) },
	} {
		if err := call(); err != nil {
			t.Fatalf("%s = %v, want nil", name, err)
		}
	}
}

func TestLivenessStateWriterDelegatesStateWriterInterface(t *testing.T) {
	var initCalls, transitionCalls int
	base := &StateWriter{
		Constrained: true,
		fname:       "states.dot",
		WriteInitStateFunc: func(*TLCStateMut) error {
			initCalls++
			return nil
		},
		WriteTransitionFunc: func(*TLCStateMut, *TLCStateMut, StateVisitStatus, *Action, ...SemanticNode) error {
			transitionCalls++
			return nil
		},
	}
	writer := &LivenessStateWriter{StateWriter: base}
	if writer.IsNoop() || writer.IsDot() || !writer.IsConstrained() || writer.GetDumpFileName() != "states.dot" {
		t.Fatalf("delegated identity = noop %t, dot %t, constrained %t, file %q", writer.IsNoop(), writer.IsDot(), writer.IsConstrained(), writer.GetDumpFileName())
	}
	if err := writer.WriteInitState(nil); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteTransition(nil, nil, StateVisitSeen, nil); err != nil {
		t.Fatal(err)
	}
	if initCalls != 1 || transitionCalls != 1 {
		t.Fatalf("delegated calls = init %d, transition %d, want 1/1", initCalls, transitionCalls)
	}
}

func TestDotLivenessStateWriterNullArgumentsFailInJavaWriteOrder(t *testing.T) {
	state := &TLCStateMut{values: []Value{NewIntValue(1)}}
	successor := &TLCStateMut{values: []Value{NewIntValue(2)}}
	sourceNode := &TBGraphNode{Index: 3}
	successorNode := &TBGraphNode{Index: 4}
	stateFP := fmtInt64(int64(state.FingerPrint()))
	successorFP := fmtInt64(int64(successor.FingerPrint()))

	tests := map[string]struct {
		call       func(*LivenessStateWriter) error
		wantSuffix string
	}{
		"init/null_state": {
			call:       func(writer *LivenessStateWriter) error { return writer.WriteLivenessInitState(nil, sourceNode) },
			wantSuffix: "\"",
		},
		"init/null_tableau": {
			call:       func(writer *LivenessStateWriter) error { return writer.WriteLivenessInitState(state, nil) },
			wantSuffix: "\"" + stateFP + ".",
		},
		"transition/null_successor": {
			call: func(writer *LivenessStateWriter) error {
				return writer.WriteLivenessTransitionVisual(state, sourceNode, nil, successorNode, nil, 0, 0, StateVisitSeen, StateVisualizationDefault)
			},
			wantSuffix: "",
		},
		"transition/null_state": {
			call: func(writer *LivenessStateWriter) error {
				return writer.WriteLivenessTransitionVisual(nil, sourceNode, successor, successorNode, nil, 0, 0, StateVisitSeen, StateVisualizationDefault)
			},
			wantSuffix: "\"",
		},
		"transition/null_source_tableau": {
			call: func(writer *LivenessStateWriter) error {
				return writer.WriteLivenessTransitionVisual(state, nil, successor, successorNode, nil, 0, 0, StateVisitSeen, StateVisualizationDefault)
			},
			wantSuffix: "\"" + stateFP + ".",
		},
		"transition/null_successor_tableau": {
			call: func(writer *LivenessStateWriter) error {
				return writer.WriteLivenessTransitionVisual(state, sourceNode, successor, nil, nil, 0, 0, StateVisitSeen, StateVisualizationDefault)
			},
			wantSuffix: "\"" + stateFP + ".3\" -> \"" + successorFP + ".",
		},
		"transition/null_action_checks": {
			call: func(writer *LivenessStateWriter) error {
				return writer.WriteLivenessTransitionVisual(state, sourceNode, successor, successorNode, nil, 0, 1, StateVisitSeen, StateVisualizationDefault)
			},
			wantSuffix: "\"" + stateFP + ".3\" -> \"" + successorFP + ".4\"",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			writer := newTestDotLivenessStateWriter(t)
			before := readLivenessWriterBytes(t, writer)
			var failure any
			func() {
				defer func() { failure = recover() }()
				_ = test.call(writer)
			}()
			if _, ok := failure.(*NullPointerException); !ok {
				t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
			}
			after := readLivenessWriterBytes(t, writer)
			if !strings.HasPrefix(after, before) || strings.TrimPrefix(after, before) != test.wantSuffix {
				t.Fatalf("bytes after failure = %q, want prefix %q plus %q", after, before, test.wantSuffix)
			}
		})
	}
}

func TestDotLivenessStateWriterOrdinaryNullStatesFollowJavaOrder(t *testing.T) {
	state := &TLCStateMut{values: []Value{NewIntValue(1)}}
	successor := &TLCStateMut{values: []Value{NewIntValue(2)}}
	for name, call := range map[string]func(*LivenessStateWriter) error{
		"init/null_state": func(writer *LivenessStateWriter) error {
			return writer.WriteLivenessState(nil)
		},
		"transition/null_successor": func(writer *LivenessStateWriter) error {
			return writer.WriteLivenessStateTransition(state, nil, nil, 0, 0, StateVisitSeen, StateVisualizationDefault)
		},
		"transition/null_state": func(writer *LivenessStateWriter) error {
			return writer.WriteLivenessStateTransition(nil, successor, nil, 0, 0, StateVisitSeen, StateVisualizationDefault)
		},
	} {
		t.Run(name, func(t *testing.T) {
			writer := newTestDotLivenessStateWriter(t)
			before := readLivenessWriterBytes(t, writer)
			var failure any
			func() {
				defer func() { failure = recover() }()
				_ = call(writer)
			}()
			if _, ok := failure.(*NullPointerException); !ok {
				t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
			}
			if after := readLivenessWriterBytes(t, writer); after != before {
				t.Fatalf("bytes after failure = %q, want unchanged %q", after, before)
			}
		})
	}

	t.Run("suppressed_stuttering_precedes_state_dereference", func(t *testing.T) {
		writer := newTestDotLivenessStateWriter(t)
		before := readLivenessWriterBytes(t, writer)
		if err := writer.WriteLivenessStateTransition(nil, nil, nil, 0, 0, StateVisitSeen, StateVisualizationStuttering); err != nil {
			t.Fatalf("suppressed stuttering = %v, want nil", err)
		}
		if after := readLivenessWriterBytes(t, writer); after != before {
			t.Fatalf("bytes after suppressed stuttering = %q, want unchanged %q", after, before)
		}
	})
}

func TestDotLivenessStateWriterTreatsCombinedFlagsAsSeen(t *testing.T) {
	state := &TLCStateMut{values: []Value{NewIntValue(1)}}
	successor := &TLCStateMut{values: []Value{NewIntValue(2)}}
	stateFP := fmtInt64(int64(state.FingerPrint()))
	successorFP := fmtInt64(int64(successor.FingerPrint()))
	status := StateVisitSeen | StateVisitNotInModel

	t.Run("ordinary", func(t *testing.T) {
		writer := newTestDotLivenessStateWriter(t)
		before := readLivenessWriterBytes(t, writer)
		if err := writer.WriteLivenessStateTransition(state, successor, nil, 0, 0, status, StateVisualizationDefault); err != nil {
			t.Fatal(err)
		}
		after := readLivenessWriterBytes(t, writer)
		want := fmtInt64(int64(state.FingerPrint())) + " -> " + fmtInt64(int64(successor.FingerPrint())) + ";\n"
		if !strings.HasPrefix(after, before) || strings.TrimPrefix(after, before) != want {
			t.Fatalf("combined ordinary flags bytes = %q, want prefix %q plus %q", after, before, want)
		}
	})

	t.Run("tableau", func(t *testing.T) {
		writer := newTestDotLivenessStateWriter(t)
		before := readLivenessWriterBytes(t, writer)
		if err := writer.WriteLivenessTransition(state, &TBGraphNode{Index: 3}, successor, &TBGraphNode{Index: 4}, nil, 0, 0, status); err != nil {
			t.Fatal(err)
		}
		after := readLivenessWriterBytes(t, writer)
		want := "\"" + stateFP + ".3\" -> \"" + successorFP + ".4\";\n"
		if !strings.HasPrefix(after, before) || strings.TrimPrefix(after, before) != want {
			t.Fatalf("combined tableau flags bytes = %q, want prefix %q plus %q", after, before, want)
		}
	})
}

func newTestDotLivenessStateWriter(t *testing.T) *LivenessStateWriter {
	t.Helper()
	base, err := NewDotStateWriter(filepath.Join(t.TempDir(), "states.dot"), DotStateWriterOptions{})
	if err != nil {
		t.Fatal(err)
	}
	writer, err := NewDotLivenessStateWriter(base)
	if err != nil {
		_ = base.Close()
		t.Fatal(err)
	}
	if err := base.Close(); err != nil {
		_ = writer.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = writer.Close() })
	return writer
}

func readLivenessWriterBytes(t *testing.T, writer *LivenessStateWriter) string {
	t.Helper()
	writer.writer.Flush()
	data, err := os.ReadFile(writer.GetDumpFileName())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
