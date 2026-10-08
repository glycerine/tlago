package tlc

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// DistributedWorkerEnvironment supplies TLCWorker.main's process and remote
// boundaries. LoadApp is the resolver-taking TLCApp constructor, after main
// has installed FP64 and the worker's interning source. StartThread must start
// asynchronously, like Thread.start; failures in that call reach main's catch,
// while failures inside the supplied function are uncaught thread failures.
type DistributedWorkerEnvironment struct {
	Lookup                 TLCServerLookup
	Sleep                  DistributedLookupSleep
	ToolOut                io.Writer
	SystemErr              io.Writer
	AvailableProcessors    func() int
	LoadApp                func(DistributedServerEndpoint, *RMIFilenameToStreamResolver) (*TLCApp, error)
	StartThread            func(string, func())
	LocalCanonicalHostName func() (string, error)
	RegisterWorker         func(DistributedServerEndpoint, *DistributedWorker) error
	PublishWorker          func(*DistributedWorker) error
	ReadyDate              func() string
}

// DistributedWorkerProcess owns TLCWorker's static resolver, current runnable
// array, timer, executor and volatile completion latch. Repeated main calls
// retain the executor/resolver, replace the latch before discovery, and replace
// the timer only after registration threads start. It does not isolate FP64 or
// UniqueString class globals from a server running in the same Go process.
// Invoke its setup/main/shutdown methods serially, as process lifecycle calls.
type DistributedWorkerProcess struct {
	Runtime  *DistributedWorkerRuntime
	Resolver *RMIFilenameToStreamResolver
	Group    *DistributedWorkerGroup
	intern   *InternTable
}

func NewDistributedWorkerProcess() *DistributedWorkerProcess {
	initializeDistributedWorkerProperties()
	return &DistributedWorkerProcess{Runtime: &DistributedWorkerRuntime{launchKeepAlive: true}}
}

// Run ports TLCWorker.main. Failures within the startup try are reported and
// swallowed. An error return represents an exception outside that catch, such
// as a negative CountDownLatch count; it does not invent a command exit status.
func (p *DistributedWorkerProcess) Run(args []string, env DistributedWorkerEnvironment) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	if p == nil || p.Runtime == nil {
		return NewNullPointerException()
	}
	if env.ToolOut == nil {
		env.ToolOut = os.Stdout
	}
	if env.SystemErr == nil {
		env.SystemErr = os.Stderr
	}
	if env.AvailableProcessors == nil {
		env.AvailableProcessors = runtime.NumCPU
	}
	if env.ReadyDate == nil {
		env.ReadyDate = func() string { return time.Now().Format("Mon Jan 02 15:04:05 MST 2006") }
	}
	fmt.Fprintln(env.ToolOut, "TLC Worker "+TLCVersion())
	if len(args) != 1 {
		fmt.Fprintln(env.ToolOut, "Error: Missing hostname of the TLC server to be contacted.")
		fmt.Fprintln(env.ToolOut, "Usage: java tlc2.tool.distributed.TLCWorker host")
		return nil
	}
	count := int(int32(env.AvailableProcessors()))
	if configured, ok := distributedIntProperty("tlc2.tool.distributed.TLCWorker.threadCount"); ok {
		count = configured
	}
	// Java assigns the volatile static field only after construction succeeds.
	p.Runtime.latch.Store(newDistributedWorkerLatch(count))
	if failure := p.start(args[0], count, env); failure != nil {
		PrintError(ECGeneral, javaGeneralErrorMessage("", failure))
		fmt.Fprintln(env.ToolOut, "Error: Failed to start worker  for server "+args[0]+".\n"+javaNullableString(javaThrowableDetailMessage(failure)))
	}
	// Invalid args and negative counts returned before this source flush.
	if stream, ok := env.ToolOut.(interface{ Flush() error }); ok {
		_ = stream.Flush()
	}
	return nil
}

func (p *DistributedWorkerProcess) start(serverName string, count int, env DistributedWorkerEnvironment) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	server, url, err := DiscoverTLCWorkerServer(serverName, env.Lookup, env.Sleep, env.ToolOut)
	if err != nil {
		return err
	}
	poly, err := server.GetIrredPolyForFP()
	if err != nil {
		return err
	}
	FP64InitPoly(poly)
	// Later invocations retain this worker process's existing strings and
	// replace only the coordinator supplying new string identities.
	if p.intern == nil {
		p.intern = NewInternTable(1024)
	}
	p.intern.SetSource(server)
	internTable = p.intern
	initBuiltInOPs()
	initCounterExampleUniqueStrings()
	if p.Resolver == nil {
		p.Resolver = NewRMIFilenameToStreamResolver()
	}
	p.Resolver.SetTLCServer(server)
	if env.LoadApp == nil {
		return NewNullPointerException()
	}
	app, err := env.LoadApp(server, p.Resolver)
	if err != nil {
		return err
	}
	manager, err := server.GetFPSetManager()
	if err != nil {
		return err
	}
	group := &DistributedWorkerGroup{Runtime: p.Runtime, server: server, runnables: make([]*DistributedWorkerRunnable, count)}
	p.Group = group
	p.Runtime.keepAliveMu.Lock()
	p.Runtime.workers = make([]*DistributedWorker, count)
	p.Runtime.runnables = group.runnables
	p.Runtime.keepAliveMu.Unlock()
	group.started.Store(true)
	for i := range group.runnables {
		runnable := &DistributedWorkerRunnable{threadID: i, server: server, app: app, manager: manager, runtime: p.Runtime, done: make(chan struct{}), localHost: env.LocalCanonicalHostName, register: env.RegisterWorker, publish: env.PublishWorker}
		p.Runtime.keepAliveMu.Lock()
		group.runnables[i] = runnable
		p.Runtime.keepAliveMu.Unlock()
		run := func() {
			if failure := runnable.Run(); failure != nil {
				fmt.Fprintf(env.SystemErr, "Exception in thread \"%s\" %s", runnable.ThreadName(), javaThrowableStackTrace(failure))
			}
		}
		if env.StartThread != nil {
			env.StartThread(runnable.ThreadName(), run)
		} else {
			go run()
		}
	}
	p.Runtime.ConfigureKeepAliveLookup(url, NewTLCServerStatusLookup(env.Lookup), nil)
	p.Runtime.replaceKeepAlive(server)
	fmt.Fprintf(env.ToolOut, "TLC worker with %d threads ready at: %s\n", count, env.ReadyDate())
	return nil
}

func (p *DistributedWorkerProcess) SetFilenameToStreamResolver(resolver *RMIFilenameToStreamResolver) {
	p.Resolver = resolver
}

// Shutdown only clears the resolver/runnable fields after all direct exits
// succeed (or throw the source ignored NoSuchObjectException).
func (p *DistributedWorkerProcess) Shutdown() (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	if p == nil || p.Runtime == nil {
		return NewNullPointerException()
	}
	if err := p.Runtime.Shutdown(); err != nil {
		return err
	}
	p.Resolver = nil
	p.Group = nil
	return nil
}
