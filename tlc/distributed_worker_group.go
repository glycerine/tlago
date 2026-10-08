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
// worker JVM. All registration runnables share the Tool, executor, and latch.
type DistributedWorkerGroup struct {
	Runtime   *DistributedWorkerRuntime
	server    *TLCServer
	runnables []*DistributedWorkerRunnable
	started   atomic.Bool
}

func DistributedWorkerThreadCount() int {
	if count, ok := distributedIntProperty("tlc2.tool.distributed.TLCWorker.threadCount"); ok {
		return count
	}
	return runtime.NumCPU()
}

func NewDistributedWorkerGroup(count int, server *TLCServer, tool *Tool, address ...DistributedWorkerAddress) *DistributedWorkerGroup {
	count = int(int32(count))
	if count < 0 {
		panic(NewIllegalArgumentException("count < 0"))
	}
	if server == nil {
		panic(NewNullPointerException())
	}
	r := NewDistributedWorkerRuntime(make([]*DistributedWorker, count)...)
	r.launchKeepAlive = true
	group := &DistributedWorkerGroup{Runtime: r, server: server, runnables: make([]*DistributedWorkerRunnable, count)}
	manager := server.GetFPSetManager()
	var app *TLCApp
	if tool != nil {
		app = NewTLCApp(tool, server.GetCheckDeadlock())
	}
	for i := range group.runnables {
		group.runnables[i] = &DistributedWorkerRunnable{threadID: i, server: server, app: app, manager: manager, runtime: r, done: make(chan struct{})}
		if len(address) > 0 {
			endpoint := address[0]
			group.runnables[i].address = &endpoint
		}
	}
	r.runnables = group.runnables
	return group
}

// Start starts each registration thread before scheduling the timer and
// printing readiness. Java does not wait for those threads to finish first.
func (g *DistributedWorkerGroup) Start() {
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
	fmt.Fprintf(os.Stdout, "TLC worker with %d threads ready at: %s\n", len(g.runnables), time.Now().Format("Mon Jan 02 15:04:05 MST 2006"))
}

// WaitForRegistrations observes the registration threads for local callers;
// it is distinct from Java's awaitTermination latch, which counts worker exits.
func (g *DistributedWorkerGroup) WaitForRegistrations() []error {
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
	workers := make([]*DistributedWorker, len(g.runnables))
	for i, runnable := range g.runnables {
		workers[i] = runnable.GetTLCWorker()
	}
	return workers
}

type DistributedWorkerRunnable struct {
	threadID     int
	server       *TLCServer
	app          *TLCApp
	manager      *DistributedFPSetManager
	runtime      *DistributedWorkerRuntime
	address      *DistributedWorkerAddress
	localHost    func() (string, error)
	register     func(*TLCServer, *DistributedWorker) error
	worker       atomic.Pointer[DistributedWorker]
	done         chan struct{}
	completeOnce sync.Once
	errorMu      sync.Mutex
	lastError    error
}

func (r *DistributedWorkerRunnable) ThreadName() string {
	return fmt.Sprintf("%s%03d", TLCWorkerThreadNamePrefix, r.threadID)
}

func (r *DistributedWorkerRunnable) GetTLCWorker() *DistributedWorker {
	if r == nil {
		panic(NewNullPointerException())
	}
	return r.worker.Load()
}

func (r *DistributedWorkerRunnable) Run() (err error) {
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
	r.worker.Store(worker)
	if r.register != nil {
		return r.register(r.server, worker)
	}
	return r.server.RegisterWorker(NewLocalWorkerEndpoint(worker))
}
