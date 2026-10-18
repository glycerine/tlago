package tlago

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"syscall"
	"testing"
)

func TestDistributedSignalRunsWorkerHookBeforeExit(t *testing.T) {
	for _, received := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		signals := make(chan os.Signal, 1)
		signals <- received
		var order []string
		var stderr bytes.Buffer
		distributedSignalShutdown(signals, make(chan struct{}), func() error {
			order = append(order, "workers")
			return errors.New("worker shutdown failure")
		}, func(status int) {
			if status != 128+int(received.(syscall.Signal)) {
				t.Errorf("exit status %d for signal %v", status, received)
			}
			order = append(order, "exit")
		}, &stderr)
		if !reflect.DeepEqual(order, []string{"workers", "exit"}) || stderr.String() != "worker shutdown failure\n" {
			t.Fatalf("shutdown order %v, stderr %q", order, stderr.String())
		}
	}
}

func TestDistributedSignalHookStopsOnNormalReturn(t *testing.T) {
	stopped := make(chan struct{})
	close(stopped)
	distributedSignalShutdown(make(chan os.Signal), stopped,
		func() error { t.Fatal("normal return ran signal hook"); return nil },
		func(int) { t.Fatal("normal return exited process") }, &bytes.Buffer{})
	// Also exercise real signal registration and disposal without sending a
	// process signal or invoking os.Exit inside the test process.
	stop := installDistributedSignalHook(func() error { t.Fatal("unexpected hook"); return nil }, &bytes.Buffer{})
	stop()
}
