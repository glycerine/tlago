package tlc

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// DistributedWorkerRuntime owns the shared resources of one worker process.
// Separate runtimes allow several worker groups in one Go program. Construct
// the group before registering it.
type DistributedWorkerRuntime struct {
	launchKeepAlive bool
	runnables       []*DistributedWorkerRunnable
	workers         []*DistributedWorker
	executor        DistributedExecutor
	latch           atomic.Pointer[distributedWorkerLatch]
	keepAliveMu     sync.Mutex
	keepAlive       *distributedWorkerKeepAlive
	keepAliveURL    string
	statusLookup    TLCServerStatusLookup
	keepAliveLog    func(string, error)
}

func NewDistributedWorkerRuntime(workers ...*DistributedWorker) *DistributedWorkerRuntime {
	r := &DistributedWorkerRuntime{workers: append([]*DistributedWorker(nil), workers...)}
	r.latch.Store(newDistributedWorkerLatch(len(workers)))
	for _, worker := range workers {
		if worker != nil {
			worker.Runtime = r
		}
	}
	return r
}

// StartKeepAlive captures the configured coordinator discovery/status boundary. Local
// groups retain a direct-server adapter; discovery groups configure relookup.
func (r *DistributedWorkerRuntime) StartKeepAlive(server DistributedServerEndpoint) {
	r.keepAliveMu.Lock()
	defer r.keepAliveMu.Unlock()
	if r.keepAlive != nil {
		return
	}
	r.startKeepAliveLocked(server)
}

// main replaces the static timer without canceling a preceding invocation's
// timer. Worker.exit still cancels the currently published static timer.
func (r *DistributedWorkerRuntime) replaceKeepAlive(server DistributedServerEndpoint) {
	r.keepAliveMu.Lock()
	defer r.keepAliveMu.Unlock()
	r.startKeepAliveLocked(server)
}

func (r *DistributedWorkerRuntime) startKeepAliveLocked(server DistributedServerEndpoint) {
	lookup := r.statusLookup
	if lookup == nil {
		lookup = func(string) (bool, error) {
			if server == nil {
				return false, NewNullPointerException()
			}
			return server.IsDone()
		}
	}
	// The Java constructor retains its supplied runnable array. Shutdown's
	// replacement of the runtime array leaves this task holding the old one.
	task := &distributedWorkerKeepAlive{workers: r.workers, runnables: r.runnables, lookup: lookup, serverURL: r.keepAliveURL, logFinest: r.keepAliveLog, done: make(chan struct{}), timeout: int64(DistributedWorkerKeepAliveTimeoutMillis())}
	r.keepAlive = task
	go task.runTimer()
}

func (r *DistributedWorkerRuntime) cancelKeepAlive(required bool) error {
	r.keepAliveMu.Lock()
	timer := r.keepAlive
	r.keepAliveMu.Unlock()
	if timer == nil {
		if required {
			return NewNullPointerException()
		}
		return nil
	}
	timer.cancel()
	return nil
}

func (r *DistributedWorkerRuntime) Shutdown() error {
	if r == nil {
		return NewNullPointerException()
	}
	_ = r.cancelKeepAlive(false)
	for index := 0; ; index++ {
		worker, present := r.shutdownWorkerAt(index)
		if !present {
			break
		}
		if worker == nil {
			continue
		}
		if err := worker.Exit(); err != nil {
			if !isDistributedWorkerEndpointRemoved(err) {
				return err
			}
		}
	}
	r.keepAliveMu.Lock()
	r.workers = []*DistributedWorker{}
	r.runnables = []*DistributedWorkerRunnable{}
	r.keepAliveMu.Unlock()
	// Java neither recreates the executor nor resets the completion latch.
	return nil
}

// Shutdown resolves each worker only when its turn is reached. Startup can
// publish a later runnable's worker while an earlier exit is in progress.
func (r *DistributedWorkerRuntime) shutdownWorkerAt(index int) (*DistributedWorker, bool) {
	r.keepAliveMu.Lock()
	defer r.keepAliveMu.Unlock()
	if r.runnables != nil {
		if index >= len(r.runnables) {
			return nil, false
		}
		return r.runnables[index].GetTLCWorker(), true
	}
	if index >= len(r.workers) {
		return nil, false
	}
	return r.workers[index], true
}

func (r *DistributedWorkerRuntime) AwaitTermination() error {
	return r.AwaitTerminationWithBoundary(nil, nil)
}

// DistributedWorkerWait is CountDownLatch.await's interruptible native boundary.
// The channel closes only after the captured latch reaches zero.
type DistributedWorkerWait func(<-chan struct{}) error

func (r *DistributedWorkerRuntime) AwaitTerminationWithBoundary(wait DistributedWorkerWait, sleep DistributedLookupSleep) error {
	if r == nil {
		return NewNullPointerException()
	}
	latch := r.latch.Load()
	if latch == nil {
		return NewNullPointerException()
	}
	if wait == nil {
		wait = func(done <-chan struct{}) error { <-done; return nil }
	}
	if err := wait(latch.done); err != nil {
		return err
	}
	if sleep == nil {
		sleep = func(duration time.Duration) error { time.Sleep(duration); return nil }
	}
	// Wait for the master to disappear before a caller reconnects.
	return sleep(10 * time.Second)
}

type distributedWorkerLatch struct {
	mu        sync.Mutex
	remaining int
	done      chan struct{}
}

func newDistributedWorkerLatch(count int) *distributedWorkerLatch {
	if count < 0 {
		panic(NewIllegalArgumentException("count < 0"))
	}
	latch := &distributedWorkerLatch{remaining: count, done: make(chan struct{})}
	if count == 0 {
		close(latch.done)
	}
	return latch
}

func (l *distributedWorkerLatch) countDown() {
	if l == nil {
		panic(NewNullPointerException())
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.done == nil {
		panic(NewNullPointerException())
	}
	if l.remaining > 0 {
		l.remaining--
		if l.remaining == 0 {
			close(l.done)
		}
	}
}

// Shutdown rejects new submissions, but allows already accepted tasks to
// finish. Exit does not wait for these tasks or an active worker invocation.
type DistributedExecutor struct {
	mu            sync.Mutex
	shutdown      bool
	interruptions chan struct{}
}

func NewDistributedExecutor() *DistributedExecutor { return &DistributedExecutor{} }

func (e *DistributedExecutor) submit(task func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.shutdown {
		panic(NewRejectedExecutionException(javaString("executor has been shut down"), nil))
	}
	go task()
}

func (e *DistributedExecutor) Shutdown() {
	e.mu.Lock()
	e.shutdown = true
	e.mu.Unlock()
}

// Interruptions is the native boundary for the interrupt sent to active
// cached-pool tasks by shutdownNow. Task/RPC adapters must observe this signal;
// ordinary Java computation can ignore interruption, and is not forcibly killed.
func (e *DistributedExecutor) Interruptions() <-chan struct{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.interruptions == nil {
		e.interruptions = make(chan struct{})
	}
	return e.interruptions
}

func (e *DistributedExecutor) ShutdownNow() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.shutdown = true
	if e.interruptions == nil {
		e.interruptions = make(chan struct{})
	}
	select {
	case <-e.interruptions:
	default:
		close(e.interruptions)
	}
}

func (e *DistributedExecutor) IsShutdown() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.shutdown
}

type distributedWorkerKeepAlive struct {
	runnables  []*DistributedWorkerRunnable
	workers    []*DistributedWorker
	lookup     TLCServerStatusLookup
	serverURL  string
	logFinest  func(string, error)
	done       chan struct{}
	cancelOnce sync.Once
	timeout    int64
}

func (t *distributedWorkerKeepAlive) cancel() { t.cancelOnce.Do(func() { close(t.done) }) }

func (t *distributedWorkerKeepAlive) runTimer() {
	// Java's uncaught timer exception terminates that thread, not the process.
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprint(os.Stderr, javaThrowableStackTrace(panicValueAsError(failure)))
		}
	}()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			started := time.Now()
			if err := t.run(); err != nil {
				panic(err)
			}
			select {
			case <-t.done:
				return
			default:
			}
			// Timer.schedule uses the preceding actual execution start for
			// fixed-delay rescheduling; a long task can make the next run due.
			timer.Reset(time.Until(started.Add(60 * time.Second)))
		case <-t.done:
			return
		}
	}
}

func (t *distributedWorkerKeepAlive) run() (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	var latest int64
	count := len(t.workers)
	if t.runnables != nil {
		count = len(t.runnables)
	}
	for i := 0; i < count; i++ {
		worker := t.workerAt(i)
		if worker == nil {
			panic(NewNullPointerException())
		}
		if worker.IsComputing() {
			return nil
		}
		latest = max(latest, worker.LastInvocation.Load())
	}
	if latest != 0 && time.Now().UnixMilli()-latest <= t.timeout {
		return nil
	}
	done, err := invokeTLCServerStatusLookup(t.lookup, t.serverURL)
	if err == nil && done {
		// This invocation is inside the source try: an exception from
		// exitWorker can itself reach that try's typed catches.
		err = t.exitWorker(nil, count)
	}
	if err == nil {
		return nil
	}
	if isDistributedMalformedLocation(err) {
		t.logFailure(err)
		return nil
	}
	if isDistributedRemoteFailure(err) {
		return t.exitWorker(err, count)
	}
	if isDistributedCoordinatorBindingMissing(err) {
		return t.exitWorker(err, count)
	}
	return err
}

func (t *distributedWorkerKeepAlive) exitWorker(failure error, count int) error {
	if failure == nil {
		PrintError(ECTLCDistributedServerFinished)
	} else {
		PrintErrorThrowable(ECTLCDistributedServerNotRunning, failure)
	}
	for i := 0; i < count; i++ {
		worker := t.workerAt(i)
		if err := worker.Exit(); err != nil {
			if isDistributedWorkerEndpointRemoved(err) {
				t.logFailure(err)
			} else {
				return err
			}
		}
	}
	t.cancel()
	return nil
}

func (t *distributedWorkerKeepAlive) logFailure(err error) {
	if t.logFinest != nil {
		t.logFinest("Failed to exit worker", err)
	}
}

func (t *distributedWorkerKeepAlive) workerAt(index int) *DistributedWorker {
	if t.runnables != nil {
		return t.runnables[index].GetTLCWorker()
	}
	return t.workers[index]
}

func (w *DistributedWorker) Exit() (err error) {
	if w == nil {
		return NewNullPointerException()
	}
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	host := "null"
	if w.uri == nil {
		return NewNullPointerException()
	}
	if w.uri.host != nil {
		host = *w.uri.host
	}
	if w.Cache == nil {
		return NewNullPointerException()
	}
	ToolIOPrintln(fmt.Sprintf("%s, work completed at: %s Computed: %d and a cache hit ratio of %s, Thank you!",
		host, time.Now().Format("Mon Jan 02 15:04:05 MST 2006"), w.OverallStatesComputed.Load(), w.Cache.GetHitRatioAsString()))
	if w.Runtime == nil {
		return NewNullPointerException()
	}
	w.Runtime.executor.Shutdown()
	if err := w.Runtime.cancelKeepAlive(true); err != nil {
		return err
	}
	if !w.unexported.CompareAndSwap(false, true) {
		return workerEndpointRemovedFailure("worker endpoint already removed")
	}
	w.Runtime.latch.Load().countDown()
	return nil
}

// Endpoint lookup precedes dispatch; removed workers cannot run another call.
func (w *DistributedWorker) endpointError() error {
	if w == nil {
		return NewNullPointerException()
	}
	if w.unexported.Load() {
		return workerEndpointRemovedFailure("worker endpoint is removed")
	}
	return nil
}
