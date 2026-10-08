package tlago

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/glycerine/tlago/tlc"
)

// Install only when TLCServer.main reaches its shutdown-hook registration.
// The hook uses the existing coordinator lookup guard and remote worker exits;
// signal handling supplies native process termination, not a JVM runtime.
func installDistributedSignalHook(hook func() error, stderr io.Writer) func() {
	signals := make(chan os.Signal, 1)
	stopped := make(chan struct{})
	done := make(chan struct{})
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer close(done)
		distributedSignalShutdown(signals, stopped, hook, func(status int) {
			tlc.CleanupDistributedFiles()
			os.Exit(status)
		}, stderr)
	}()
	return func() {
		signal.Stop(signals)
		close(stopped)
		<-done
	}
}

func distributedSignalShutdown(signals <-chan os.Signal, stopped <-chan struct{}, hook func() error, exit func(int), stderr io.Writer) {
	select {
	case <-stopped:
		return
	case received := <-signals:
		if err := hook(); err != nil {
			fmt.Fprintln(stderr, err)
		}
		status := 1
		if number, ok := received.(syscall.Signal); ok {
			status = 128 + int(number)
		}
		exit(status)
	}
}
