package tlc

import "testing"

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
			failure := NewRemoteException(javaString("URI lookup failed"), nil)
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
	return "rmi://worker:10997/0", nil
}

func (w *registrationRemoteWorker) GetCacheRateRatio() (float64, error) {
	return 0, nil
}
