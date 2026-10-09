package tlago

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// GDB kills only the owned coordinator inferior at an inspected caller PC.
// This opt-in check needs ptrace permission and an optimized, symbol-bearing
// binary built with go test -c; an ordinary transient go test binary can omit DWARF.
func TestNativeDistributedRemoteCheckpointTraceToInternInterruption(t *testing.T) {
	if os.Getenv("TLAGO_DISTRIBUTED_GDB_BOUNDARY") != "1" {
		t.Skip("set TLAGO_DISTRIBUTED_GDB_BOUNDARY=1 to run the external GDB boundary check")
	}
	checkNativeDistributedRemoteCheckpointRestart(t, "lsb", 1, nativeRemoteCheckpointTraceToIntern, func(command *exec.Cmd) {
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			// GDB may put its inferior in a separate process group. Find the
			// startup PID and verify that it is still this debugger's child.
			output := command.Stdout.(*nativeDistributedTestLog).String()
			match := regexp.MustCompile(`(?m)^\s*\*\s+1\s+process (\d+)\s`).FindStringSubmatch(output)
			if len(match) == 2 {
				pid, _ := strconv.Atoi(match[1])
				status, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
				if err == nil {
					for _, line := range strings.Split(string(status), "\n") {
						fields := strings.Fields(line)
						if len(fields) == 2 && fields[0] == "PPid:" && fields[1] == strconv.Itoa(command.Process.Pid) {
							_ = syscall.Kill(pid, syscall.SIGKILL)
						}
					}
				}
			}
			return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
		command.WaitDelay = time.Second
	})
}
