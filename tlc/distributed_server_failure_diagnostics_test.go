package tlc

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// No enabled upstream test directly covers main's system-failure reporting.
// Keep the full process catch/finally order while asserting the source overload.
func TestDistributedServerSystemFailureDiagnostics(t *testing.T) {
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	for _, family := range []string{"stack", "memory"} {
		for _, panics := range []bool{false, true} {
			for _, debug := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/panic=%v/debug=%v", family, panics, debug), func(t *testing.T) {
					captureFailoverToolIO(t, ToolIOTool)
					Globals.Lock()
					oldDebug := Globals.Debug
					Globals.Debug = debug
					Globals.Unlock()
					t.Cleanup(func() { Globals.Lock(); Globals.Debug = oldDebug; Globals.Unlock() })
					recorder := &MemoryRecorder{}
					AddMessageRecorder(recorder)
					t.Cleanup(func() { RemoveMessageRecorder(recorder) })
					var failure error = NewStackOverflowError("system failure detail")
					code := ECSystemStackOverflow
					if family == "memory" {
						failure = NewOutOfMemoryError("system failure detail")
						code = ECSystemOutOfMemory
					}
					var calls []string
					server := &TLCServer{}
					process := NewDistributedServerProcess()
					env := DistributedServerEnvironment{
						LoadProperties: func() { calls = append(calls, "properties") },
						CreateApp:      func([]string) (*TLCApp, error) { calls = append(calls, "app"); return &TLCApp{}, nil },
						CreateServer:   func(*TLCApp, int) (*TLCServer, error) { calls = append(calls, "server"); return server, nil },
						CreateMBean: func(*TLCServer) (*TLCStandardMBean, error) {
							calls = append(calls, "bean")
							return NewNullTLCStandardMBean(), nil
						},
						InstallShutdownHook: func(func() error) error { calls = append(calls, "hook"); return nil },
						ModelCheck: func(*TLCServer) error {
							calls = append(calls, "model")
							if panics {
								panic(failure)
							}
							return failure
						},
						GC: func() { calls = append(calls, "gc") },
						Close: func(got *TLCServer, cleanup bool) error {
							calls = append(calls, "close")
							if got != server || cleanup {
								t.Fatal("failure close ownership changed")
							}
							return nil
						},
						ShutdownNow: func(got *TLCServer) error {
							calls = append(calls, "shutdown")
							if got != server {
								t.Fatal("executor ownership changed")
							}
							got.executor.ShutdownNow()
							return nil
						},
						Unregister: func(*TLCStandardMBean) (bool, error) { calls = append(calls, "unregister"); return true, nil },
					}
					if err := process.Run([]string{"Spec"}, env); err != nil {
						t.Fatalf("reported system error unexpectedly escaped: %v", err)
					}
					want := []string{"properties", "app", "server", "bean", "hook", "model", "gc", "close", "shutdown", "unregister"}
					if !reflect.DeepEqual(calls, want) || !server.executor.IsShutdown() || len(process.ShutdownHooks) != 1 {
						t.Fatalf("process lifecycle = %v, want %v", calls, want)
					}
					records := recorder.Records(code)
					if len(records) != 1 || !reflect.DeepEqual(records[0].Params, []string{"system failure detail"}) || recorder.Recorded(ECGeneral) {
						t.Fatalf("system failure event changed: %v", recorder.Messages)
					}
					toolIO.Lock()
					pending := toolIO.nextMessage
					toolIO.Unlock()
					if pending != "" {
						t.Fatal("stack printer left an unfinished ToolIO message")
					}
					output := strings.Join(ToolIOGetAllMessages(), "\n")
					if strings.Contains(output, "distributed_server_failure_diagnostics_test.go") != debug {
						t.Fatalf("source debug-stack policy lost: %q", output)
					}
				})
			}
		}
	}
}
