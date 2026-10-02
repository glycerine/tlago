package tlc

import (
	"fmt"
	"io"
	"os"
)

const DistributedFPServerThreadName = "tlc2.tool.distributed.fp.DistributedFPSet"
const DistributedWorkerThreadName = "tlc2.tool.distributed.TLCWorker"

// DistributedWorkerAndFPServerEnvironment supplies the two command bodies and
// Thread.start/default uncaught-handler boundaries. StartThread must start
// asynchronously. A failure to start escapes the launcher; a failure inside
// either command is delivered to UncaughtException on that command's thread.
type DistributedWorkerAndFPServerEnvironment struct {
	FPServer          DistributedFPServerEnvironment
	Worker            DistributedWorkerEnvironment
	FPMain            func([]string) error
	WorkerMain        func([]string) error
	StartThread       func(string, func())
	UncaughtException func(string, error)
	SystemErr         io.Writer
}

// RunDistributedWorkerAndFPServer ports TLCWorkerAndFPSet.main. Both commands
// receive the original argument array. Starting the FP thread first does not
// establish FP readiness, and this launcher neither waits nor cleans up after
// a failure to start the second thread. A nil process is constructed lazily on
// the worker thread, where that command initializes its startup properties.
func RunDistributedWorkerAndFPServer(process *DistributedWorkerProcess, args []string, env DistributedWorkerAndFPServerEnvironment) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	if env.FPMain == nil {
		env.FPMain = func(args []string) error { RunDistributedFPServer(args, env.FPServer); return nil }
	}
	if env.WorkerMain == nil {
		env.WorkerMain = func(args []string) error {
			if process == nil {
				process = NewDistributedWorkerProcess()
			}
			return process.Run(args, env.Worker)
		}
	}
	if env.StartThread == nil {
		env.StartThread = func(_ string, run func()) { go run() }
	}
	if env.UncaughtException == nil {
		output := env.SystemErr
		if output == nil {
			output = os.Stderr
		}
		env.UncaughtException = func(name string, failure error) {
			fmt.Fprintf(output, "Exception in thread \"%s\" %s", name, javaThrowableStackTrace(failure))
		}
	}
	start := func(name string, main func([]string) error) {
		env.StartThread(name, func() {
			if failure := invokeDistributedServerOperation(func() error { return main(args) }); failure != nil {
				env.UncaughtException(name, failure)
			}
		})
	}
	start(DistributedFPServerThreadName, env.FPMain)
	start(DistributedWorkerThreadName, env.WorkerMain)
	return nil
}
