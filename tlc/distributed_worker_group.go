package tlc

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// DistributedWorkerGroup is the post-lookup portion of TLCWorker.main in one
// worker process. All registration runnables share the Tool, executor, and latch.
type DistributedWorkerGroup struct {
	Runtime   *DistributedWorkerRuntime
	server    DistributedServerEndpoint
	runnables []*DistributedWorkerRunnable
	started   atomic.Bool
}

func DistributedWorkerThreadCount() int {
	if count, ok := distributedIntProperty("tlc2.tool.distributed.TLCWorker.threadCount"); ok {
		return count
	}
	return runtime.NumCPU()
}

func NewDistributedWorkerGroup(count int, server DistributedServerEndpoint, tool *Tool, address ...DistributedWorkerAddress) (*DistributedWorkerGroup, error) {
	if len(address) > 1 {
		panic(NewIllegalArgumentException("NewDistributedWorkerGroup accepts at most one address"))
	}
	count = int(int32(count))
	if count < 0 {
		panic(NewIllegalArgumentException("count < 0"))
	}
	if isNil(server) {
		panic(NewNullPointerException())
	}
	r := NewDistributedWorkerRuntime(make([]*DistributedWorker, count)...)
	r.launchKeepAlive = true
	group := &DistributedWorkerGroup{Runtime: r, server: server, runnables: make([]*DistributedWorkerRunnable, count)}
	manager, err := server.GetFPSetManager()
	if err != nil {
		return nil, err
	}
	var app *TLCApp
	if tool != nil {
		deadlock, err := server.GetCheckDeadlock()
		if err != nil {
			return nil, err
		}
		app = NewTLCApp(tool, deadlock)
	}
	for i := range group.runnables {
		group.runnables[i] = newDistributedWorkerRunnable(i, server, manager, app, r)
		if len(address) > 0 {
			endpoint := address[0]
			group.runnables[i].address = &endpoint
		}
	}
	r.runnables = group.runnables
	return group, nil
}

// Start starts each registration thread before scheduling the timer and
// printing readiness. Java does not wait for those threads to finish first.
func (g *DistributedWorkerGroup) Start() {
	g.require()
	if !g.started.CompareAndSwap(false, true) {
		return
	}
	for _, runnable := range g.runnables {
		go func(runnable *DistributedWorkerRunnable) {
			if err := runnable.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Exception in thread \"%s\" %s", runnable.ThreadName(), javaThrowableStackTrace(err))
			}
		}(runnable)
	}
	g.Runtime.StartKeepAlive(g.server)
	ToolIOPrintln(fmt.Sprintf("TLC worker with %d threads ready at: %s", len(g.runnables), time.Now().Format("Mon Jan 02 15:04:05 MST 2006")))
}

// WaitForRegistrations observes the registration threads for local callers;
// it is distinct from Java's awaitTermination latch, which counts worker exits.
func (g *DistributedWorkerGroup) WaitForRegistrations() []error {
	g.require()
	errors := make([]error, len(g.runnables))
	for i, runnable := range g.runnables {
		<-runnable.done
		runnable.errorMu.Lock()
		errors[i] = runnable.lastError
		runnable.errorMu.Unlock()
	}
	return errors
}

func (g *DistributedWorkerGroup) Workers() []*DistributedWorker {
	g.require()
	workers := make([]*DistributedWorker, len(g.runnables))
	for i, runnable := range g.runnables {
		workers[i] = runnable.GetTLCWorker()
	}
	return workers
}

func (g *DistributedWorkerGroup) require() {
	if g == nil {
		panic(NewNullPointerException())
	}
}

type DistributedWorkerRunnable struct {
	threadID     int
	server       DistributedServerEndpoint
	app          *TLCApp
	manager      *DistributedFPSetManager
	runtime      *DistributedWorkerRuntime
	address      *DistributedWorkerAddress
	localHost    func() (string, error)
	register     func(DistributedServerEndpoint, *DistributedWorker) error
	publish      func(*DistributedWorker) error
	worker       atomic.Pointer[DistributedWorker]
	done         chan struct{}
	completeOnce sync.Once
	errorMu      sync.Mutex
	lastError    error
}

// NewDistributedWorkerRunnable exposes TLCWorker.TLCWorkerRunnable's public
// constructor. A directly constructed runnable owns the one-worker process
// runtime that Java supplies through TLCWorker's static fields.
func NewDistributedWorkerRunnable(threadID int, server DistributedServerEndpoint, manager *DistributedFPSetManager, app *TLCApp) *DistributedWorkerRunnable {
	runtime := NewDistributedWorkerRuntime(nil)
	runnable := newDistributedWorkerRunnable(threadID, server, manager, app, runtime)
	runtime.runnables = []*DistributedWorkerRunnable{runnable}
	return runnable
}

func newDistributedWorkerRunnable(threadID int, server DistributedServerEndpoint, manager *DistributedFPSetManager, app *TLCApp, runtime *DistributedWorkerRuntime) *DistributedWorkerRunnable {
	return &DistributedWorkerRunnable{
		threadID: int(int32(threadID)),
		server:   server,
		app:      app,
		manager:  manager,
		runtime:  runtime,
		done:     make(chan struct{}),
	}
}

func (r *DistributedWorkerRunnable) ThreadName() string {
	r.require()
	return fmt.Sprintf("%s%03d", TLCWorkerThreadNamePrefix, r.threadID)
}

func (r *DistributedWorkerRunnable) GetTLCWorker() *DistributedWorker {
	r.require()
	return r.worker.Load()
}

func (r *DistributedWorkerRunnable) Run() (err error) {
	r.require()
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
		if err != nil && isJavaIOException(err) {
			err = NewRuntimeExceptionFromCause(err)
		}
		r.errorMu.Lock()
		r.lastError = err
		r.errorMu.Unlock()
		r.completeOnce.Do(func() { close(r.done) })
	}()
	endpoint := DistributedWorkerAddress{}
	if r.address != nil {
		endpoint = *r.address
	} else {
		localHost := r.localHost
		if localHost == nil {
			localHost = distributedCanonicalLocalHost
		}
		host, err := localHost()
		if err != nil {
			return err
		}
		endpoint.Hostname = host
	}
	worker := NewDistributedWorker(r.threadID, nil, r.manager, endpoint)
	worker.App = r.app
	worker.Runtime = r.runtime
	if r.publish != nil {
		if err := r.publish(worker); err != nil {
			return err
		}
	}
	r.worker.Store(worker)
	if r.register != nil {
		return r.register(r.server, worker)
	}
	if isNil(r.server) {
		return NewNullPointerException()
	}
	return r.server.RegisterWorker(NewLocalWorkerEndpoint(worker))
}

func (r *DistributedWorkerRunnable) require() {
	if r == nil {
		panic(NewNullPointerException())
	}
}
