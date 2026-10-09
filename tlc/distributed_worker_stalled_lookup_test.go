package tlc

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type stalledWorkerFingerprintLookup struct {
	*LocalFingerprintEndpoint
	entered, release, finished chan struct{}
	calls                      atomic.Int64
}

func (e *stalledWorkerFingerprintLookup) ContainsBlock(fps *LongVec) (*BitVector, error) {
	e.calls.Add(1)
	close(e.entered)
	defer close(e.finished)
	<-e.release
	return e.LocalFingerprintEndpoint.ContainsBlock(fps)
}

// Java synchronizes computation, but not isAlive/exit. Its executor shutdown
// allows accepted work to finish. No original direct test covers a stalled
// fingerprint RPC overlapping these calls; this is native transport coverage.
func TestWorkerRPCStalledFingerprintLookupAllowsControlCalls(t *testing.T) {
	for _, exit := range []bool{false, true} {
		name := "keepalive"
		if exit {
			name = "exit"
		}
		t.Run(name, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			storage := NewMemFPSet()
			lookup := &stalledWorkerFingerprintLookup{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(storage),
				entered: make(chan struct{}), release: make(chan struct{}), finished: make(chan struct{})}
			fpHost, fpClient := startFingerprintRPC(t, lookup)
			worker, generated := managerOwnershipWorker(t, 2, NewDistributedFPSetManager(fpClient))
			var statusCalls atomic.Int32
			if !exit {
				coordinatorHost, coordinator := startCoordinatorRPC(t, NewLocalServerEndpoint(&TLCServer{}))
				discovery := NewDistributedNetworkDiscovery()
				t.Cleanup(func() {
					if err := discovery.Close(); err != nil {
						t.Error(err)
					}
				})
				location := "tcp://" + coordinator.Address + "/main"
				view, err := discovery.Lookup(location)
				if err != nil {
					t.Fatal(err)
				}
				if done, err := view.IsDone(); err != nil || done {
					t.Fatalf("initial coordinator status = %v/%v", done, err)
				}
				if err := coordinatorHost.Close(); err != nil {
					t.Fatal(err)
				}
				status := NewTLCServerStatusLookup(discovery.Lookup)
				worker.Runtime.ConfigureKeepAliveLookup(location, func(url string) (bool, error) {
					statusCalls.Add(1)
					return status(url)
				}, nil)
			}
			worker.Runtime.StartKeepAlive(nil)
			t.Cleanup(func() { _ = worker.Runtime.cancelKeepAlive(false) })
			workerHost, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
			type answer struct {
				result *NextStateResult
				err    error
			}
			done := make(chan answer, 1)
			joined := false
			var control chan error
			controlJoined := false
			var release sync.Once
			t.Cleanup(func() {
				release.Do(func() { close(lookup.release) })
				if !joined {
					<-done
				}
				if control != nil && !controlJoined {
					<-control
				}
				if err := workerHost.CloseGracefully(); err != nil {
					t.Error(err)
				}
				if err := fpHost.CloseGracefully(); err != nil {
					t.Error(err)
				}
			})
			go func() {
				result, err := client.GetNextStates([]*TLCStateMut{{UID: 31, level: 4}})
				done <- answer{result, err}
			}()
			select {
			case <-lookup.entered:
			case <-time.After(5 * time.Second):
				t.Fatal("worker did not reach actual fingerprint lookup")
			}
			if !worker.Computing.Load() || *generated != 1 || worker.OverallStatesComputed.Load() != 2 || storage.Size() != 0 {
				t.Fatal("blocked lookup changed source computation/statistics boundary")
			}
			invocationStarted := worker.LastInvocation.Load()
			if invocationStarted == 0 {
				t.Fatal("actual computation did not record its invocation start")
			}
			if !exit {
				// The real GetNextStates call owns Computing here. Keepalive
				// must not consult the lost coordinator while lookup is blocked.
				if err := worker.Runtime.RunKeepAliveOnce(); err != nil {
					t.Fatal(err)
				}
				if statusCalls.Load() != 0 || worker.unexported.Load() || worker.Runtime.executor.IsShutdown() {
					t.Fatal("keepalive interrupted real computation after coordinator loss")
				}
			}
			control = make(chan error, 1)
			go func() {
				alive, err := client.IsAlive()
				if err == nil && !alive {
					err = NewIllegalStateException("computing worker is not alive")
				}
				if err == nil {
					_, err = client.GetCacheRateRatio()
				}
				if err == nil && exit {
					err = client.Exit()
				}
				control <- err
			}()
			select {
			case err := <-control:
				controlJoined = true
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("control calls waited for fingerprint lookup")
			}
			if !worker.Computing.Load() || worker.unexported.Load() != exit || worker.Runtime.executor.IsShutdown() != exit {
				t.Fatal("control calls cancelled accepted computation or changed exit state")
			}
			if exit {
				if _, err := client.IsAlive(); !isDistributedWorkerEndpointRemoved(err) {
					t.Fatalf("exited worker remained available: %v", err)
				}
			}
			release.Do(func() { close(lookup.release) })
			select {
			case reply := <-done:
				joined = true
				if reply.err != nil || reply.result == nil || reply.result.StatesComputed != 2 || len(reply.result.NextStates) != 1 || reply.result.NextStates[0].Size() != 2 {
					t.Fatalf("accepted lookup/computation did not finish: %v/%v", reply.result, reply.err)
				}
				for i := 0; i < 2; i++ {
					if reply.result.NextStates[0].At(i).UID != 31 {
						t.Fatal("successor lost predecessor trace identity")
					}
				}
			case <-time.After(5 * time.Second):
				t.Fatal("accepted computation did not finish after lookup release")
			}
			<-lookup.finished
			if worker.Computing.Load() || lookup.calls.Load() != 1 || storage.Size() != 0 {
				t.Fatal("accepted lookup was replayed, inserted fingerprints or retained computation flag")
			}
			if !exit {
				// Source records invocation start, not completion. A short call
				// remains recent, but completing a long call must not refresh it.
				if worker.LastInvocation.Load() != invocationStarted {
					t.Fatal("computation completion changed its invocation-start timestamp")
				}
				if err := worker.Runtime.RunKeepAliveOnce(); err != nil {
					t.Fatal(err)
				}
				if statusCalls.Load() != 0 || worker.unexported.Load() || worker.Runtime.executor.IsShutdown() {
					t.Fatal("keepalive ignored actual recent computation after coordinator loss")
				}
			}
		})
	}
}
