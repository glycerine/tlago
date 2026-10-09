package tlc

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// Upstream has no enabled direct checks of these ToolIO destinations. Preserve
// the existing diagnostic text and lifecycle while honoring configured streams.
func TestDistributedWorkerConsoleUsesToolIO(t *testing.T) {
	for _, mode := range []int{ToolIOSystem, ToolIOTool} {
		for _, event := range []string{"ready", "exit"} {
			t.Run(fmt.Sprintf("%s/mode=%d", event, mode), func(t *testing.T) {
				captureFailoverToolIO(t, mode)
				var output, errors bytes.Buffer
				if mode == ToolIOSystem {
					ToolIOSetSystemStreams(&output, &errors)
				}
				var pattern string
				if event == "ready" {
					group := &DistributedWorkerGroup{Runtime: NewDistributedWorkerRuntime(), server: NewLocalServerEndpoint(&TLCServer{})}
					t.Cleanup(func() { _ = group.Runtime.cancelKeepAlive(false) })
					group.Start()
					pattern = `^TLC worker with 0 threads ready at: .+$`
				} else {
					worker := NewDistributedWorker(0, nil, NewDistributedFPSetManager(), DistributedWorkerAddress{Hostname: "worker.example", Port: 1})
					worker.OverallStatesComputed.Store(17)
					worker.Cache.Hit(9)
					worker.Cache.Hit(9)
					worker.Runtime.StartKeepAlive(NewLocalServerEndpoint(&TLCServer{}))
					t.Cleanup(func() { _ = worker.Runtime.cancelKeepAlive(false) })
					if err := worker.Exit(); err != nil {
						t.Fatal(err)
					}
					if !worker.Runtime.executor.IsShutdown() || !worker.unexported.Load() {
						t.Fatal("worker output routing changed executor/unpublication lifetime")
					}
					select {
					case <-worker.Runtime.latch.Load().done:
					default:
						t.Fatal("worker exit did not release its completion latch")
					}
					pattern = `^worker\.example, work completed at: .+ Computed: 17 and a cache hit ratio of 1, Thank you!$`
				}
				messages := ToolIOGetAllMessages()
				if mode == ToolIOSystem {
					messages = []string{strings.TrimSuffix(output.String(), "\n")}
				}
				if len(messages) != 1 || !regexp.MustCompile(pattern).MatchString(messages[0]) || errors.Len() != 0 {
					t.Fatalf("worker console messages = %q, stderr %q", messages, errors.String())
				}
			})
		}
	}
}

func TestDistributedOptionConsoleUsesToolIO(t *testing.T) {
	for _, mode := range []int{ToolIOSystem, ToolIOTool} {
		for _, test := range []struct {
			name  string
			args  []string
			want  []string
			valid bool
		}{
			{"coverage", []string{"-coverage", "3", "Spec"}, []string{"Warning: coverage reporting not supported in distributed TLC, ignoring -coverage 3 parameter."}, true},
			{"absolute_fp_memory", []string{"-fpmem", "2", "Spec"}, []string{"Using -fpmem with an absolute byte value is deprecated. Configure the Go runtime memory budget (for example, GOMEMLIMIT) and use -fpmem to select the fraction for fingerprint storage."}, true},
			{"missing_config", []string{"-config"}, []string{"Error: configuration file required.", "Usage: tlago server [-option] inputfile"}, false},
		} {
			t.Run(fmt.Sprintf("%s/mode=%d", test.name, mode), func(t *testing.T) {
				captureFailoverToolIO(t, mode)
				var output, errors bytes.Buffer
				if mode == ToolIOSystem {
					ToolIOSetSystemStreams(&output, &errors)
				}
				options := ParseTLCAppOptions(test.args)
				if (options != nil) != test.valid {
					t.Fatal("diagnostic routing changed option acceptance")
				}
				if mode == ToolIOTool {
					if got := ToolIOGetAllMessages(); !reflect.DeepEqual(got, test.want) {
						t.Fatalf("option messages = %q, want %q", got, test.want)
					}
				} else if output.String() != strings.Join(test.want, "\n")+"\n" || errors.Len() != 0 {
					t.Fatalf("option console = %q, stderr %q", output.String(), errors.String())
				}
			})
		}
	}
}

func TestDistributedRoleUsageNamesNativeCommands(t *testing.T) {
	for _, role := range []string{"worker", "fpserver"} {
		for _, args := range [][]string{nil, {"first", "second"}} {
			t.Run(fmt.Sprintf("%s/args=%d", role, len(args)), func(t *testing.T) {
				var output, errors bytes.Buffer
				lookup := func(string) (DistributedServerEndpoint, error) {
					t.Fatal("invalid role arguments reached coordinator discovery")
					return nil, nil
				}
				banner := "TLC Worker "
				if role == "worker" {
					if err := NewDistributedWorkerProcess().Run(args, DistributedWorkerEnvironment{ToolOut: &output, SystemErr: &errors, Lookup: lookup}); err != nil {
						t.Fatal(err)
					}
				} else {
					banner = "TLC Distributed FP Server "
					RunDistributedFPServer(args, DistributedFPServerEnvironment{ToolOut: &output, SystemErr: &errors, Lookup: lookup})
				}
				want := banner + TLCVersion() + "\nError: Missing hostname of the TLC server to be contacted.\nUsage: tlago " + role + " host\n"
				if output.String() != want || errors.Len() != 0 {
					t.Fatalf("role usage = %q, stderr %q", output.String(), errors.String())
				}
			})
		}
	}
}
