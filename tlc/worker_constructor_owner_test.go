package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

// No original method directly covers constructor owner/context failures.
func TestWorkerConstructionUsesPassedToolAndCheckerDirectory(t *testing.T) {
	metadata, otherDirectory := t.TempDir(), t.TempDir()
	checkerTool, passedTool := NewTool(), NewTool()
	checkerTool.RootName, passedTool.RootName = "Checker", "Passed"
	trace := NewTLCTrace(otherDirectory, "Other")
	defer trace.Close()
	checker := &ModelChecker{AbstractChecker: &AbstractChecker{Metadir: metadata, Tool: checkerTool}, Trace: trace, ConcurrentTrace: &ConcurrentTLCTrace{Workers: make([]*Worker, 1)}}
	worker := NewModelCheckingWorker(0, checker, passedTool)
	defer worker.CloseTrace()
	want := filepath.Join(metadata, "Passed-0")
	if worker.traceFileBase != want || worker.traceRAF == nil || worker.Tool != passedTool || checker.Workers[0] != worker || checker.ConcurrentTrace.Workers[0] != worker {
		t.Fatalf("constructor borrowed another owner's context: %q, want %q", worker.traceFileBase, want)
	}
	if _, err := os.Stat(want + tlcTraceExt); err != nil {
		t.Fatal(err)
	}
	if checker.Tool != checkerTool || trace.diskdir != otherDirectory || trace.rootName != "Other" {
		t.Fatal("constructor changed supplied owners")
	}
}

func TestWorkerConstructionFailurePrecedesCheckerRegistration(t *testing.T) {
	for _, scenario := range []string{"missing-checker", "missing-tool", "missing-trace", "outside-trace", "open-failure"} {
		t.Run(scenario, func(t *testing.T) {
			directory := t.TempDir()
			tool := NewTool()
			prior := NewWorker(0)
			checker := &ModelChecker{AbstractChecker: &AbstractChecker{Metadir: directory, Tool: tool, Workers: []*Worker{prior}}, ConcurrentTrace: &ConcurrentTLCTrace{Workers: []*Worker{prior}}}
			original := checker
			id := 0
			switch scenario {
			case "missing-checker":
				checker = nil
			case "missing-tool":
				tool = nil
			case "missing-trace":
				checker.ConcurrentTrace = nil
			case "outside-trace":
				id = 2
			case "open-failure":
				checker.Metadir = filepath.Join(directory, "missing")
			}
			err := invokeDistributedServerOperation(func() error {
				worker := NewModelCheckingWorker(id, checker, tool)
				_ = worker.CloseTrace()
				return nil
			})
			switch scenario {
			case "outside-trace":
				if _, ok := err.(*ArrayIndexOutOfBoundsException); !ok {
					t.Fatalf("invalid registration: %T/%v", err, err)
				}
			case "open-failure":
				if !isJavaIOException(err) {
					t.Fatalf("open failure: %T/%v", err, err)
				}
			default:
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing constructor owner: %T/%v", err, err)
				}
			}
			if len(original.Workers) != 1 || original.Workers[0] != prior || original.ConcurrentTrace != nil && (len(original.ConcurrentTrace.Workers) != 1 || original.ConcurrentTrace.Workers[0] != prior) {
				t.Fatal("failed construction published worker registration")
			}
			base := workerTraceFileBase(original.Metadir, "Spec", id)
			_, fileErr := os.Stat(base + tlcTraceExt)
			if scenario == "missing-trace" || scenario == "outside-trace" {
				if fileErr != nil {
					t.Fatal("registration failed before source trace creation")
				}
			} else if !os.IsNotExist(fileErr) {
				t.Fatal("earlier constructor failure created a trace file")
			}
		})
	}
}
