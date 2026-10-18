package tlago

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"testing"
)

type distributedCLITestCloser struct {
	err   error
	calls int
}

func (c *distributedCLITestCloser) Close() error {
	c.calls++
	return c.err
}

// Java has no native Go network owner. Check the deferred CLI boundary after
// both successful and failed command bodies: shutdown must complete once, keep
// earlier diagnostics, and expose real cleanup failures to shell callers.
func TestDistributedCLINetworkShutdownResult(t *testing.T) {
	for _, commandFailure := range []bool{false, true} {
		for _, cleanupFailure := range []bool{false, true} {
			t.Run(fmt.Sprintf("command_failure=%t/cleanup_failure=%t", commandFailure, cleanupFailure), func(t *testing.T) {
				var stderr bytes.Buffer
				owner := &distributedCLITestCloser{}
				if cleanupFailure {
					owner.err = errors.Join(net.ErrClosed, errors.New("accepted connection release failed"))
				}
				status := func() (status int) {
					defer closeDistributedCLINetwork(owner, &stderr, &status)
					if commandFailure {
						fmt.Fprintln(&stderr, "command failed")
						return ExitToolFailure
					}
					return ExitOK
				}()
				expectedStatus := ExitOK
				if commandFailure || cleanupFailure {
					expectedStatus = ExitToolFailure
				}
				expectedDiagnostic := ""
				if commandFailure {
					expectedDiagnostic += "command failed\n"
				}
				if cleanupFailure {
					expectedDiagnostic += owner.err.Error() + "\n"
				}
				if owner.calls != 1 || status != expectedStatus || stderr.String() != expectedDiagnostic {
					t.Fatalf("close calls=%d, status=%d, stderr=%q; want one close, status=%d, stderr=%q", owner.calls, status, stderr.String(), expectedStatus, expectedDiagnostic)
				}
			})
		}
	}
}
