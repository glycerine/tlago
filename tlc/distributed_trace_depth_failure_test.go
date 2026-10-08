package tlc

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// No original method directly covers trace depth read failure ordering.
func TestDistributedTraceDepthFailureRetainsCursor(t *testing.T) {
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	for _, predecessor := range []int64{1, 0} {
		if err := trace.raf.WriteLongNat(predecessor); err != nil {
			t.Fatal(err)
		}
	}
	if err := trace.raf.Seek(4); err != nil {
		t.Fatal(err)
	}
	trace.lastPtr = 6 // The predecessor has only two bytes before EOF.
	trace.previousLevel = 7
	_, err := trace.getLevelFromDiskLocked(trace.lastPtr)
	if !isJavaIOException(err) {
		t.Fatalf("truncated predecessor failure = %T/%v", err, err)
	}
	if pointer, err := trace.raf.GetFilePointer(); err != nil || pointer != 8 {
		t.Fatalf("failed traversal rewound the cursor: %d/%v, want 8", pointer, err)
	}
	if trace.previousLevel != 7 {
		t.Fatal("failed traversal published a new reported level")
	}
}

func TestDistributedTraceDepthReportingFailure(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	for _, predecessor := range []int64{1, 0} {
		if err := trace.raf.WriteLongNat(predecessor); err != nil {
			t.Fatal(err)
		}
	}
	trace.lastPtr = 4
	if depth, err := trace.GetLevelForReportingWithError(); err != nil || depth != 2 {
		t.Fatalf("valid depth = %d/%v, want 2", depth, err)
	}
	trace.lastPtr = 0
	if depth, err := trace.GetLevelForReportingWithError(); err != nil || depth != 2 {
		t.Fatalf("reported depth decreased: %d/%v", depth, err)
	}
	if pointer, err := trace.raf.GetFilePointer(); err != nil || pointer != 8 {
		t.Fatalf("successful depth traversal did not restore cursor: %d/%v", pointer, err)
	}
	trace.lastPtr = 6
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	server := &TLCServer{Trace: trace, StateQueue: NewMemStateQueue(), FPSetManager: NewDistributedFPSetManager()}
	server.WorkerStatesGenerated.Store(20)
	generated, distinct := int64(7), uint64(8)
	if err := server.PrintProgressStats(time.Time{}, &generated, &distinct); !isJavaIOException(err) {
		t.Fatalf("periodic reporting swallowed trace failure: %T/%v", err, err)
	}
	if generated != 7 || distinct != 8 || recorder.Recorded(ECTLCProgressStats) || trace.previousLevel != 2 {
		t.Fatal("failed trace reporting published counters, output or depth")
	}
	if got := (&ModelChecker{Trace: trace}).GetProgress(); got != -1 {
		t.Fatalf("model checker I/O catch = %d, want -1", got)
	}
	output, err := os.CreateTemp(t.TempDir(), "depth-stderr")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stderr
	os.Stderr = output
	defer func() { os.Stderr = previous; _ = output.Close() }()
	wrapper := &TLCServerMXWrapper{Server: server}
	if got := wrapper.GetProgress(); got != -1 {
		t.Fatalf("server management I/O catch = %d, want -1", got)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil || !strings.Contains(string(data), "EOFException") {
		t.Fatalf("management failure diagnostic missing: %q/%v", data, err)
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := trace.GetLevelForReportingWithError(); !isJavaIOException(err) {
		t.Fatalf("closed trace reported a stale depth: %v", err)
	}
}

func TestDistributedFinalTraceDepthFailureProcess(t *testing.T) {
	if os.Getenv("TLAGO_FINAL_DEPTH_FAILURE") == "" {
		command := exec.Command(os.Args[0], "-test.run=^TestDistributedFinalTraceDepthFailureProcess$")
		command.Env = append(os.Environ(), "TLAGO_FINAL_DEPTH_FAILURE=true", tlcServerPropertyPrefix+".report=1")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("native final reporting process failed: %v\n%s", err, output)
		}
		return
	}
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	directory := t.TempDir()
	trace := NewTLCTrace(directory, "Spec")
	defer trace.Close()
	server := NewTLCServer("Spec", "Spec", directory, NewDistributedFPSetManager(), NewMemStateQueue(directory), trace)
	server.SetTool(NewTool())
	server.StatesPerMinute, server.DistinctStatesPerMinute = 17, 19
	lateCalls := 0
	server.ConfigurePublication(TLCServerPublication{
		LocalHostName: func() (string, error) { return "local", nil },
		CreateRegistry: func(int) (*TLCServerRegistry, error) {
			return &TLCServerRegistry{
				Rebind: func(name string, s *TLCServer) error {
					if name == TLCServerWorkerName {
						s.SetDone()
						return trace.Close()
					}
					return nil
				},
				Unbind: func(string) error { lateCalls++; return nil },
			}, nil
		},
		Flush:    func() { lateCalls++ },
		Unexport: func(*TLCServer, bool) (bool, error) { lateCalls++; return true, nil },
	})
	code, err := server.ModelCheck()
	if code != ECGeneral || !isJavaIOException(err) {
		t.Fatalf("final trace failure = %d/%T/%v", code, err, err)
	}
	if !server.executor.IsShutdown() || TLCServerFinalNumberOfDistinctStates() != 0 || server.StatesPerMinute != 17 || server.DistinctStatesPerMinute != 19 {
		t.Fatal("final trace failure changed preceding shutdown/counts or later rate reset")
	}
	if lateCalls != 0 || recorder.Recorded(ECTLCStats) || recorder.Recorded(ECTLCFinished) || recorder.Recorded(ECTLCSuccess) {
		t.Fatal("trace failure continued into success, summary or cleanup")
	}
	if _, err := os.Stat(directory); err != nil {
		t.Fatal("trace failure deleted metadata", err)
	}
}
