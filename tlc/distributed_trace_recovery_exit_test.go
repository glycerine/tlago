package tlc

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// No original method exercises these fatal reconstruction branches directly.
// Child processes test the real exit, including bypassing handler recovery and
// deferred cleanup, without replacing the production exit with a test hook.
func TestDistributedTraceRecoveryExit(t *testing.T) {
	if branch := os.Getenv("TLAGO_TRACE_RECOVERY_EXIT"); branch != "" {
		initTLCCheckerTest(t)
		captureFailoverToolIO(t, ToolIOTool)
		AddMessageRecorder(RecorderFunc(func(m Message) {
			fmt.Fprintf(os.Stdout, "TRACE_EVENT %d %q %d\n", m.Code, m.Params, m.StateNumber)
		}))
		defer fmt.Fprintln(os.Stdout, "TRACE_DEFER_RAN")
		defer func() {
			if v := recover(); v != nil {
				fmt.Fprintln(os.Stdout, "TRACE_PANIC", v)
			}
		}()
		current, successor := checkerTestState(2), checkerTestState(3)
		current.level, successor.level = 2, 3
		tool := &Tool{GetStateFunc: func(*Tool, uint64, ...any) (*TLCStateInfo, error) { return nil, nil }}
		var prefix []*TLCStateInfo
		switch branch {
		case "3":
			successor = nil
		case "4":
			successor = nil
			prefix = []*TLCStateInfo{NewTLCStateInfo(checkerTestState(1))}
		case "5":
			tool.GetStateFunc = func(*Tool, uint64, ...any) (*TLCStateInfo, error) { return NewTLCStateInfo(current), nil }
		default:
			t.Fatal("unknown branch", branch)
		}
		trace := NewTLCTrace()
		trace.Tool = tool
		trace.printTraceWithPrefix(current, successor, prefix)
		fmt.Fprintln(os.Stdout, "TRACE_RETURNED")
		return
	}
	for _, branch := range []string{"3", "4", "5"} {
		t.Run(branch, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestDistributedTraceRecoveryExit$", "-test.v")
			command.Env = append(os.Environ(), "TLAGO_TRACE_RECOVERY_EXIT="+branch)
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 {
				t.Fatalf("want exit 1, got %v\n%s", err, output)
			}
			text := string(output)
			var events []string
			for _, line := range strings.Split(text, "\n") {
				if strings.HasPrefix(line, "TRACE_EVENT ") {
					events = append(events, line)
				}
			}
			want := []string{fmt.Sprintf("TRACE_EVENT %d [] 0", ECTLCBehaviorUpToThisPoint)}
			if branch != "3" {
				want = append(want, fmt.Sprintf("TRACE_EVENT %d ", ECTLCStatePrint2))
			}
			want = append(want, fmt.Sprintf("TRACE_EVENT %d [] 0", ECTLCFailedToRecoverInit), fmt.Sprintf("TRACE_EVENT %d [%q] 0", ECTLCBug, branch))
			if branch != "3" {
				want = append(want, fmt.Sprintf("TRACE_EVENT %d ", ECTLCStatePrint1))
			}
			if len(events) != len(want) {
				t.Fatalf("event count: got %v, want %v\n%s", events, want, output)
			}
			for i := range want {
				if !strings.HasPrefix(events[i], want[i]) {
					t.Fatalf("event %d: got %q, want prefix %q\n%s", i, events[i], want[i], output)
				}
			}
			if branch != "3" && !strings.HasSuffix(events[len(events)-1], " -1") {
				t.Fatalf("missing standalone state: %v", events)
			}
			for _, forbidden := range []string{"TRACE_DEFER_RAN", "TRACE_PANIC", "TRACE_RETURNED"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("fatal recovery ran %s\n%s", forbidden, output)
				}
			}
		})
	}
}
