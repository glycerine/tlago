package tlc

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"
)

// DistributedWorkerRuntime represents the static fields shared by workers in
// one Java worker JVM. Separate runtimes allow several such processes to be
// represented in one Go program. Construct the group before registering it.
type DistributedWorkerRuntime struct {
	workers     []*DistributedWorker
	executor    DistributedExecutor
	latch       distributedWorkerLatch
	keepAliveMu sync.Mutex
	keepAlive   *distributedWorkerKeepAlive
}

func NewDistributedWorkerRuntime(workers ...*DistributedWorker) *DistributedWorkerRuntime {
	r := &DistributedWorkerRuntime{workers: append([]*DistributedWorker(nil), workers...)}
	r.latch.remaining = len(workers)
	r.latch.done = make(chan struct{})
	if len(workers) == 0 {
		close(r.latch.done)
	}
	for _, worker := range workers {
		if worker != nil {
			worker.Runtime = r
		}
	}
	return r
}

// StartKeepAlive supplies the local equivalent of main's server lookup. The
// transport's registry/lookup failures are intentionally left to the later
// transport port; activity, done checks, timer order and exits are represented.
func (r *DistributedWorkerRuntime) StartKeepAlive(server *TLCServer) {
	r.keepAliveMu.Lock()
	defer r.keepAliveMu.Unlock()
	if r.keepAlive != nil {
		return
	}
	timeout := 60
	if configured, ok := distributedIntProperty("tlc2.tool.distributed.TLCTimerTask.timeout"); ok {
		timeout = configured
	}
	task := &distributedWorkerKeepAlive{workers: append([]*DistributedWorker(nil), r.workers...), server: server, done: make(chan struct{}), timeout: int64(int32(timeout) * 1000)}
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
	r.keepAliveMu.Lock()
	workers := append([]*DistributedWorker(nil), r.workers...)
	r.keepAliveMu.Unlock()
	for _, worker := range workers {
		if worker == nil {
			continue
		}
		if err := worker.Exit(); err != nil {
			if _, missing := err.(*NoSuchObjectException); !missing {
				return err
			}
		}
	}
	r.keepAliveMu.Lock()
	r.workers = []*DistributedWorker{}
	r.keepAliveMu.Unlock()
	// Java neither recreates the executor nor resets the completion latch.
	return nil
}

func (r *DistributedWorkerRuntime) AwaitTermination() error {
	if r == nil || r.latch.done == nil {
		return NewNullPointerException()
	}
	<-r.latch.done
	// Wait for the master to disappear before a caller reconnects.
	time.Sleep(10 * time.Second)
	return nil
}

type distributedWorkerLatch struct {
	mu        sync.Mutex
	remaining int
	done      chan struct{}
}

func (l *distributedWorkerLatch) countDown() {
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
	mu       sync.Mutex
	shutdown bool
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

func (e *DistributedExecutor) IsShutdown() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.shutdown
}

type distributedWorkerKeepAlive struct {
	workers    []*DistributedWorker
	server     *TLCServer
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
			if !t.run() {
				return
			}
			timer.Reset(60 * time.Second)
		case <-t.done:
			return
		}
	}
}

func (t *distributedWorkerKeepAlive) run() bool {
	var latest int64
	for _, worker := range t.workers {
		if worker == nil {
			panic(NewNullPointerException())
		}
		if worker.IsComputing() {
			return true
		}
		latest = max(latest, worker.LastInvocation.Load())
	}
	if latest != 0 && time.Now().UnixMilli()-latest <= t.timeout {
		return true
	}
	if t.server == nil {
		panic(NewNullPointerException())
	}
	if !t.server.IsDone() {
		return true
	}
	PrintError(ECTLCDistributedServerFinished)
	for _, worker := range t.workers {
		if err := worker.Exit(); err != nil {
			if _, missing := err.(*NoSuchObjectException); !missing {
				// An uncaught runtime failure terminates Java's timer thread.
				fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
				return false
			}
		}
	}
	t.cancel()
	return false
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
	if uri, parseErr := url.Parse(w.URI); parseErr == nil && uri.Hostname() != "" {
		host = uri.Hostname()
	}
	if w.Cache == nil {
		return NewNullPointerException()
	}
	fmt.Fprintf(os.Stdout, "%s, work completed at: %s Computed: %d and a cache hit ratio of %s, Thank you!\n",
		host, time.Now().Format("Mon Jan 02 15:04:05 MST 2006"), w.OverallStatesComputed.Load(), w.Cache.GetHitRatioAsString())
	if w.Runtime == nil {
		return NewNullPointerException()
	}
	w.Runtime.executor.Shutdown()
	if err := w.Runtime.cancelKeepAlive(true); err != nil {
		return err
	}
	if !w.unexported.CompareAndSwap(false, true) {
		return NewNoSuchObjectException("object not exported")
	}
	w.Runtime.latch.countDown()
	return nil
}

// Endpoint lookup precedes dispatch, so this error is not a ServerException.
func (w *DistributedWorker) remoteEndpointError() error {
	if w == nil {
		return NewNullPointerException()
	}
	if w.unexported.Load() {
		return NewNoSuchObjectException("no such object in table")
	}
	return nil
}
