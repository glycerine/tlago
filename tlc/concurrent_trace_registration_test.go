package tlc

import "testing"

// No original method directly covers invalid trace-worker registration.
func TestConcurrentTraceRegistrationPreservesFixedOwnerList(t *testing.T) {
	for _, scenario := range []string{"missing-trace", "missing-worker", "negative", "outside", "empty", "replace"} {
		t.Run(scenario, func(t *testing.T) {
			first, last := NewWorker(0), NewWorker(1)
			trace := &ConcurrentTLCTrace{Workers: []*Worker{first, last}}
			worker := NewWorker(1)
			switch scenario {
			case "missing-trace":
				trace = nil
			case "missing-worker":
				worker = nil
			case "negative":
				worker.ID = -1
			case "outside":
				worker.ID = 2
			case "empty":
				trace.Workers = nil
				worker.ID = 0
			}
			var result *Worker
			err := invokeDistributedServerOperation(func() error {
				result = trace.AddWorker(worker)
				return nil
			})
			switch scenario {
			case "replace":
				if err != nil || result != worker || trace.Workers[0] != first || trace.Workers[1] != worker || len(trace.Workers) != 2 {
					t.Fatalf("valid registration failed: %v", err)
				}
				return
			case "missing-trace", "missing-worker":
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("required owner failure = %T/%v", err, err)
				}
			default:
				if _, ok := err.(*ArrayIndexOutOfBoundsException); !ok {
					t.Fatalf("invalid index failure = %T/%v", err, err)
				}
			}
			if result != nil {
				t.Fatal("invalid registration returned a worker")
			}
			if trace != nil {
				if scenario == "empty" {
					if len(trace.Workers) != 0 {
						t.Fatal("registration expanded an empty owner list")
					}
				} else if len(trace.Workers) != 2 || trace.Workers[0] != first || trace.Workers[1] != last {
					t.Fatal("invalid registration changed owner list")
				}
			}
		})
	}
}
