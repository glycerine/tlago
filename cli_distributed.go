package tlago

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/glycerine/tlago/tlc"
)

func isDistributedCLICommand(command string) bool {
	return command == "server" || command == "worker" || command == "fpserver" || command == "worker-fpserver"
}

// Each invocation owns one process role. FP64 and interning startup are
// process-wide, just as in the original distributed commands; run coordinator
// and workers in separate OS processes rather than sharing evaluator globals.
func runDistributedCLI(role string, args []string, stdout, stderr io.Writer) (status int) {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	if role != "server" && len(args) != 1 {
		fmt.Fprintf(stderr, "usage: tlago %s [-DPROPERTY=VALUE] COORDINATOR_HOST\n", role)
		return ExitToolFailure
	}
	restore := tlc.ToolIOSetSystemStreams(stdout, stderr)
	defer restore()
	bindHost := tlc.DistributedSystemProperty("tlago.distributed.bindHost", "")
	advertiseHost := tlc.DistributedSystemProperty("tlago.distributed.advertiseHost", "")
	if role == "server" {
		network := tlc.NewDistributedCoordinatorNetwork(bindHost, advertiseHost)
		defer closeDistributedCLINetwork(network, stderr, &status)
		process := tlc.NewDistributedServerProcess()
		var stopSignals func()
		defer func() {
			if stopSignals != nil {
				stopSignals()
			}
		}()
		env := tlc.DistributedServerEnvironment{CreateServer: func(app *tlc.TLCApp, count int) (*tlc.TLCServer, error) {
			var server *tlc.TLCServer
			var err error
			if count > 0 {
				server, err = tlc.NewDistributedFPSetTLCServer(app, count)
			} else {
				server, err = tlc.NewTLCServerFromApp(app)
			}
			if err == nil {
				server.ConfigurePublication(network.Publication())
			}
			return server, err
		}, InstallShutdownHook: func(hook func() error) error {
			stopSignals = installDistributedSignalHook(hook, stderr)
			return nil
		}}
		_, err = RunDistributedServer(process, args, env, tlc.RuntimeParameters{})
	} else {
		if advertiseHost == "" {
			advertiseHost, err = os.Hostname()
			if err != nil {
				fmt.Fprintln(stderr, err)
				return ExitToolFailure
			}
		}
		portText := tlc.DistributedSystemProperty("tlago.distributed.callbackPort", "0")
		port, parseErr := strconv.Atoi(portText)
		if parseErr != nil || port < 0 || port > 65535 {
			fmt.Fprintln(stderr, "tlago.distributed.callbackPort must be between 0 and 65535")
			return ExitToolFailure
		}
		network, openErr := tlc.NewDistributedWorkerNetwork(tlc.DistributedBindAddress(bindHost, port), advertiseHost)
		if openErr != nil {
			fmt.Fprintln(stderr, openErr)
			return ExitToolFailure
		}
		defer closeDistributedCLINetwork(network, stderr, &status)
		worker := func(args []string) error {
			process := tlc.NewDistributedWorkerProcess()
			_, failure := RunDistributedWorker(process, args, network.Environment(tlc.DistributedWorkerEnvironment{ToolOut: stdout, SystemErr: stderr}), tlc.RuntimeParameters{})
			if failure == nil && process.Group != nil {
				// Go exits when main returns. Retain the worker process until
				// TLC's exit latch completes, including the source grace period.
				failure = process.Runtime.AwaitTermination()
			}
			return failure
		}
		fp := func(args []string) error {
			tlc.RunDistributedFPServer(args, network.FPEnvironment(tlc.DistributedFPServerEnvironment{ToolOut: stdout, SystemOut: stdout, SystemErr: stderr}))
			return nil
		}
		switch role {
		case "worker":
			err = worker(args)
		case "fpserver":
			err = fp(args)
		case "worker-fpserver":
			var lifetime sync.WaitGroup
			err = tlc.RunDistributedWorkerAndFPServer(nil, args, tlc.DistributedWorkerAndFPServerEnvironment{
				FPMain: fp, WorkerMain: worker, SystemErr: stderr,
				StartThread: func(_ string, run func()) {
					lifetime.Add(1)
					go func() { defer lifetime.Done(); run() }()
				},
			})
			lifetime.Wait()
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitToolFailure
	}
	return ExitOK
}

// Native networking belongs to the command invocation. Its shutdown result
// must reach the caller even when the TLC command body completed successfully.
// Network owners already classify benign closure and retain real failure causes.
func closeDistributedCLINetwork(network io.Closer, stderr io.Writer, status *int) {
	if err := network.Close(); err != nil {
		fmt.Fprintln(stderr, err)
		*status = ExitToolFailure
	}
}

func printDistributedCLIHelp(w io.Writer, role string) {
	operand := "COORDINATOR_HOST"
	if role == "server" {
		operand = "[TLC FLAGS] SPEC"
	}
	fmt.Fprintf(w, "Usage: tlago %s [-DPROPERTY=VALUE] %s\n\n", role, operand)
	classes := map[string]string{"server": "tlc2.tool.distributed.TLCServer", "worker": "tlc2.tool.distributed.TLCWorker", "fpserver": "tlc2.tool.distributed.fp.DistributedFPSet", "worker-fpserver": "tlc2.tool.distributed.fp.TLCWorkerAndFPSet"}
	fmt.Fprintf(w, "Java counterpart: java -DPROPERTY=VALUE -cp tla2tools.jar %s.\n", classes[role])
	fmt.Fprintln(w, "These roles use native Go TCP RPC. Start the coordinator first; workers fetch the model and configuration from it.")
	fmt.Fprintln(w, "Properties use the original Java names and are applied before role initialization. Help: -help, --help, -h.")
	fmt.Fprintln(w, "\nProcess properties:")
	fmt.Fprintln(w, "  -Dtlc2.tool.distributed.TLCServer.port=N           Coordinator port (default 10997); set identically on all roles.")
	fmt.Fprintln(w, "  -Dtlago.distributed.bindHost=HOST                 Native listener interface; IPv6 may be bare or bracketed (default all interfaces).")
	fmt.Fprintln(w, "  -Dtlago.distributed.advertiseHost=HOST            Reachable callback/publication hostname (default local hostname).")
	fmt.Fprintln(w, "  -Dtlago.distributed.callbackPort=N                Worker/FP callback port (default 0: OS chooses); allow this port through firewalls.")
	fmt.Fprintln(w, "The tlago.distributed properties configure Go networking and have no Java flag equivalent.")
	if role == "server" {
		fmt.Fprintln(w, "  -Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=N  Wait for N separate fingerprint servers; default uses coordinator storage.")
		fmt.Fprintln(w, "  -Dtlc2.tool.distributed.TLCServer.report=N         Progress reporting interval in milliseconds (default 60000).")
		fmt.Fprintln(w, "\nCoordinator flags (same spellings and behavior as Java TLCApp.create):")
		fmt.Fprintln(w, "  -config FILE         Model configuration; default SPEC.cfg.")
		fmt.Fprintln(w, "  -deadlock            Disable deadlock checking.")
		fmt.Fprintln(w, "  -recover DIR         Recover model checking from the checkpoint directory.")
		fmt.Fprintln(w, "  -checkpoint MINUTES  Checkpoint interval; zero disables checkpoints.")
		fmt.Fprintln(w, "  -coverage MINUTES    Accepted with the original warning: distributed coverage is unsupported.")
		fmt.Fprintln(w, "  -tool                Emit tool-mode messages for editor integrations.")
		fmt.Fprintln(w, "  -terse               Disable value expansion in output.")
		fmt.Fprintln(w, "  -nowarning           Suppress warnings.")
		fmt.Fprintln(w, "  -maxSetSize N        Maximum permitted set size.")
		fmt.Fprintln(w, "  -fp INDEX            Fingerprint polynomial index.")
		fmt.Fprintln(w, "  -fpbits BITS         Fingerprint storage partition bits.")
		fmt.Fprintln(w, "  -fpmem FRACTION      Fraction of available memory for fingerprint storage.")
		fmt.Fprintln(w, "  -metadir DIR         Metadata directory for queues, traces and checkpoints.")
		fmt.Fprintln(w, "Model choices correspond to Toolbox constants, behavior, invariants and properties in the .tla/.cfg files.")
		fmt.Fprintln(w, "The distributed coordinator accepts its original subset of TLC flags; worker counts are configured on worker processes.")
	} else {
		fmt.Fprintln(w, "  -Dtlc2.tool.distributed.TLCWorker.threadCount=N    Worker threads (default available processors); applies to worker roles.")
		fmt.Fprintln(w, "COORDINATOR_HOST is the server hostname. Model options belong on the coordinator, not the worker/FP command.")
		fmt.Fprintln(w, "IPv6 coordinator hosts may be bare or bracketed (for example ::1 or [::1]); scoped hosts retain their %zone suffix.")
	}
	if role != "worker" {
		fmt.Fprintln(w, "\nFingerprint storage:")
		fmt.Fprintln(w, "  -Dtlc2.tool.fp.FPSet.impl=NAME                    Select the native Go fingerprint backend; default tlc2.tool.fp.MSBDiskFPSet.")
		fmt.Fprintln(w, "Supported disk choices are tlc2.tool.fp.MSBDiskFPSet, tlc2.tool.fp.LSBDiskFPSet and tlc2.tool.fp.OffHeapDiskFPSet.")
		fmt.Fprintln(w, "These original Java property values select Go implementations; they do not load Java classes or use RMI.")
		fmt.Fprintln(w, "Set this property on the process owning storage: server for coordinator storage, fpserver or worker-fpserver for remote storage.")
		fmt.Fprintln(w, "Standalone fingerprint servers use their own memory budget and two storage partitions; coordinator -fpmem and -fpbits do not configure them.")
		fmt.Fprintln(w, "Go memory settings: GOMEMLIMIT supplies the heap budget; TLAGO_MAX_DIRECT_MEMORY supplies the off-heap budget (for example 256m).")
		fmt.Fprintln(w, "Remote off-heap named checkpoints are unsupported, matching Java; MSB and LSB storage support them.")
	}
	fmt.Fprintln(w, "\nCoordinator storage example (run commands in separate terminals):")
	fmt.Fprintln(w, "  tlago server -config Spec.cfg Spec.tla")
	fmt.Fprintln(w, "  tlago worker -Dtlc2.tool.distributed.TLCWorker.threadCount=4 coordinator-host")
	fmt.Fprintln(w, "\nRemote storage example (one fingerprint server and one worker process):")
	fmt.Fprintln(w, "  tlago server -Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=1 -config Spec.cfg Spec.tla")
	fmt.Fprintln(w, "  tlago fpserver coordinator-host")
	fmt.Fprintln(w, "  tlago worker coordinator-host")
	fmt.Fprintln(w, "Replace the last two commands with tlago worker-fpserver coordinator-host to combine their roles.")
	fmt.Fprintln(w, "The coordinator waits for expectedFPSetCount registrations before starting; its default local-storage mode rejects remote fingerprint registrations.")
}
