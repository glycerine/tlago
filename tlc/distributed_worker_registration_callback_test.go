package tlc

import (
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

type distributedRegistrationWakeQueue struct {
	StateQueue
	wakes atomic.Int32
}

func TestWorkerRPCLazyCallbackConnectionOwnership(t *testing.T) {
	host, eager := startWorkerRPC(t, &rpcTestWorker{})
	unused := &NetworkWorkerEndpoint{Address: eager.Address, Object: eager.Object}
	if err := unused.CloseConnection(); err != nil {
		t.Fatal(err)
	}
	if _, err := unused.IsAlive(); !isDistributedWorkerUnavailable(err) || unused.client != nil {
		t.Fatalf("closed unused callback reference dialed: %v", err)
	}
	lazy := &NetworkWorkerEndpoint{Address: eager.Address, Object: eager.Object}
	t.Cleanup(func() { _ = lazy.CloseConnection() })
	start := make(chan struct{})
	var calls sync.WaitGroup
	for range 8 {
		calls.Add(1)
		go func() {
			defer calls.Done()
			<-start
			if uri, err := lazy.GetURI(); err != nil || uri != "tcp://worker:1234/primary" {
				t.Errorf("concurrent first callback = %q/%v", uri, err)
			}
		}()
	}
	close(start)
	calls.Wait()
	client := lazy.client
	if client == nil {
		t.Fatal("successful callback did not retain its connection")
	}
	if err := host.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := lazy.IsAlive(); !isDistributedWorkerUnavailable(err) || lazy.client != client {
		t.Fatalf("established failed callback connection was replaced: %v", err)
	}
	if err := lazy.CloseConnection(); err != nil {
		t.Fatal(err)
	}
	if err := lazy.CloseConnection(); err != nil {
		t.Fatal(err)
	}
	if _, err := lazy.GetURI(); !isDistributedWorkerUnavailable(err) {
		t.Fatalf("closed callback reference remained usable: %v", err)
	}
}

func (q *distributedRegistrationWakeQueue) ResumeAllStuck() { q.wakes.Add(1) }

// TLCServer.registerWorker wakes stuck queue consumers before invoking getURI.
// A native transport dial must not fail before that source-owned side effect.
func TestCoordinatorRPCWorkerRegistrationWakesQueueBeforeCallbackFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	queue := &distributedRegistrationWakeQueue{}
	server := &TLCServer{StateQueue: queue}
	_, coordinator := startCoordinatorRPC(t, NewLocalServerEndpoint(server))
	err = coordinator.RegisterWorkerReference(DistributedEndpointReference{Address: address, Object: "worker"})
	if !isDistributedRemoteFailure(err) || !isJavaIOException(err) {
		t.Fatalf("unreachable callback lost its remote I/O failure: %v", err)
	}
	if queue.wakes.Load() != 1 || server.GetWorkerCount() != 0 {
		t.Fatalf("callback failure changed registration order: wakes %d, workers %d", queue.wakes.Load(), server.GetWorkerCount())
	}
}
