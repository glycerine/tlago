package tlc

import (
	"testing"
	"time"
)

// Java has no direct registerWorker failure-order test. Exercise the remote
// boundary so an IOException cannot silently turn into successful registration.
func TestTLCServerRegistrationRemoteURIOrder(t *testing.T) {
	for _, failAt := range []int{1, 2, 0} {
		t.Run(fmtInt(failAt), func(t *testing.T) {
			oldWorkers := NumWorkers()
			t.Cleanup(func() { SetNumWorkers(oldWorkers) })
			queue := NewMemStateQueue()
			queue.FinishAll()
			server := &TLCServer{StateQueue: queue}
			failure := distributedTestRemoteFailure("URI lookup failed")
			worker := &registrationRemoteWorker{
				DistributedWorkerSmartProxy: NewDistributedWorkerSmartProxy(nil),
				server:                      server, failAt: failAt, failure: failure,
			}
			err := server.RegisterWorker(worker)
			threads := server.GetServerThreads()
			// Observe the already-started thread before restoring class globals.
			for _, thread := range threads {
				thread.Join()
			}
			wantCalls, wantThreads := 2, 1
			if failAt == 1 {
				wantCalls, wantThreads = 1, 0
			}
			if worker.calls != wantCalls || len(threads) != wantThreads {
				t.Fatalf("calls/threads = %d/%d, want %d/%d", worker.calls, len(threads), wantCalls, wantThreads)
			}
			if failAt != 0 && err != failure {
				t.Fatalf("error = %v, want original remote failure %v", err, failure)
			}
			if failAt == 0 && err != nil {
				t.Fatal(err)
			}
			if worker.registeredBeforeFirst || !worker.registeredBeforeSecond && wantCalls == 2 {
				t.Fatalf("registration at URI calls = %v/%v, want false/true", worker.registeredBeforeFirst, worker.registeredBeforeSecond)
			}
		})
	}
}

type registrationRemoteWorker struct {
	*DistributedWorkerSmartProxy
	server                 *TLCServer
	failAt, calls          int
	failure                error
	registeredBeforeFirst  bool
	registeredBeforeSecond bool
}

func (w *registrationRemoteWorker) GetURI() (string, error) {
	w.calls++
	registered := len(w.server.GetServerThreads()) != 0
	if w.calls == 1 {
		w.registeredBeforeFirst = registered
	} else {
		w.registeredBeforeSecond = registered
	}
	if w.calls == w.failAt {
		return "", w.failure
	}
	return "tcp://worker:10997/0", nil
}

func (w *registrationRemoteWorker) GetCacheRateRatio() (float64, error) {
	return 0, nil
}

// Source registerWorker resumes the queue before either URI call or thread
// creation. A missing queue must not turn into a partially started registration.
func TestTLCServerRegistrationRequiresQueueBeforeWorkerContact(t *testing.T) {
	server := &TLCServer{}
	failure := distributedTestRemoteFailure("worker must not be contacted")
	worker := &registrationRemoteWorker{server: server, failAt: 1, failure: failure}
	err := server.RegisterWorker(worker)
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("registration error %T/%v, want missing queue failure", err, err)
	}
	if worker.calls != 0 || len(server.GetServerThreads()) != 0 {
		t.Fatal("missing queue reached worker contact or thread registration")
	}
}

type registrationWakeQueue struct {
	*MemStateQueue
	calls   int
	failure error
}

func (q *registrationWakeQueue) ResumeAllStuck() {
	q.calls++
	if q.failure != nil {
		panic(q.failure)
	}
}

func TestTLCServerRegistrationWakePrecedesNullWorker(t *testing.T) {
	queue := &registrationWakeQueue{MemStateQueue: NewMemStateQueue()}
	server := &TLCServer{StateQueue: queue}
	err := server.RegisterWorker(nil)
	if _, ok := err.(*NullPointerException); !ok || queue.calls != 1 || len(server.GetServerThreads()) != 0 {
		t.Fatalf("error %T/%v, wake calls %d, threads %d", err, err, queue.calls, len(server.GetServerThreads()))
	}
}

func TestTLCServerRegistrationWakeFailureReleasesMonitor(t *testing.T) {
	failure := NewIllegalStateException("queue wake failed")
	queue := &registrationWakeQueue{MemStateQueue: NewMemStateQueue(), failure: failure}
	server := &TLCServer{StateQueue: queue}
	worker := &registrationRemoteWorker{server: server, failAt: 1, failure: distributedTestRemoteFailure("unexpected URI call")}
	func() {
		defer func() {
			if got := recover(); got != failure {
				t.Fatalf("wake failure %v, want original %v", got, failure)
			}
		}()
		_ = server.RegisterWorker(worker)
	}()
	if worker.calls != 0 || queue.calls != 1 || len(server.GetServerThreads()) != 0 {
		t.Fatal("wake failure contacted or registered worker")
	}
	done := make(chan struct{})
	go func() { server.monitor.Lock(); server.monitor.Unlock(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("registration failure retained coordinator monitor")
	}
}
