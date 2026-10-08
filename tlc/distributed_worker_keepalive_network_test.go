package tlc

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type keepaliveFailureCoordinator struct{ *LocalServerEndpoint }

func (s *keepaliveFailureCoordinator) IsDone() (bool, error) {
	return false, NewRemoteException(javaString("coordinator status failed"), nil)
}

// No enabled Java test directly covers this timer boundary. Invoke its public
// run once against real native discovery/status calls rather than changing the
// production timer's schedule or activity timeout.
func TestNativeWorkerKeepAliveCoordinatorLifecycle(t *testing.T) {
	for _, event := range []string{"finished", "unbound", "disconnected", "status_failure"} {
		for _, debug := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/debug=%v", event, debug), func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				recorder := &MemoryRecorder{}
				AddMessageRecorder(recorder)
				t.Cleanup(func() { RemoveMessageRecorder(recorder) })
				Globals.Lock()
				oldDebug := Globals.Debug
				Globals.Debug = debug
				Globals.Unlock()
				t.Cleanup(func() { Globals.Lock(); Globals.Debug = oldDebug; Globals.Unlock() })
				server := &TLCServer{InternTable: NewInternTable(16)}
				var coordinator DistributedServerEndpoint = NewLocalServerEndpoint(server)
				if event == "status_failure" {
					coordinator = &keepaliveFailureCoordinator{coordinator.(*LocalServerEndpoint)}
				}
				host, client := startCoordinatorRPC(t, coordinator)
				network, err := NewDistributedWorkerNetwork("127.0.0.1:0", "")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = network.Close() })
				worker := NewDistributedWorker(0, nil, NewDistributedFPSetManager(), network.workerAddress)
				env := network.Environment(DistributedWorkerEnvironment{})
				if err := env.PublishWorker(worker); err != nil {
					t.Fatal(err)
				}
				callback, err := DialWorkerEndpoint(network.Address, worker.networkReference.Object)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = callback.CloseConnection() })
				runtime := worker.Runtime
				runtime.ConfigureKeepAliveLookup("tcp://"+client.Address+"/main", NewTLCServerStatusLookup(network.Discovery.Lookup), nil)
				runtime.StartKeepAlive(nil)
				t.Cleanup(func() { _ = runtime.cancelKeepAlive(false) })
				switch event {
				case "finished":
					server.SetDone()
				case "unbound":
					host.UnregisterCoordinator("main")
				case "disconnected":
					if err := host.Close(); err != nil {
						t.Fatal(err)
					}
				}
				// Computing and recent activity suppress discovery, even if the master
				// has vanished. Only the idle invocation is allowed to shut down workers.
				worker.Computing.Store(true)
				if err := runtime.RunKeepAliveOnce(); err != nil {
					t.Fatal(err)
				}
				worker.Computing.Store(false)
				worker.LastInvocation.Store(time.Now().UnixMilli())
				if err := runtime.RunKeepAliveOnce(); err != nil {
					t.Fatal(err)
				}
				if worker.unexported.Load() || runtime.executor.IsShutdown() || len(ToolIOGetAllMessages()) != 0 {
					t.Fatal("active/recent worker was shut down or diagnosed")
				}
				worker.LastInvocation.Store(0)
				if err := runtime.RunKeepAliveOnce(); err != nil {
					t.Fatal(err)
				}
				if !worker.unexported.Load() || !runtime.executor.IsShutdown() {
					t.Fatal("idle worker did not exit after coordinator completion/loss")
				}
				select {
				case <-runtime.latch.Load().done:
				default:
					t.Fatal("worker completion latch was not released")
				}
				select {
				case <-runtime.keepAlive.done:
				default:
					t.Fatal("keepalive timer was not cancelled")
				}
				if _, err := callback.IsAlive(); !isDistributedRemoteFailure(err) {
					t.Fatalf("exited worker callback stayed available: %v", err)
				}
				messages := strings.Join(ToolIOGetAllMessages(), "\n")
				code := ECTLCDistributedServerNotRunning
				if event == "finished" {
					code = ECTLCDistributedServerFinished
				}
				records := recorder.Records(code)
				if len(records) != 1 {
					t.Fatalf("keepalive event %d was not recorded exactly once: %v", code, recorder.Messages)
				}
				if event == "status_failure" && (len(records[0].Params) != 1 || records[0].Params[0] != "coordinator status failed" || !strings.Contains(messages, "TLCServer is gone due to coordinator status failed, exiting worker...")) {
					t.Fatalf("coordinator status diagnostic absent: %q", messages)
				}
				if event == "status_failure" && strings.Contains(messages, "distributed_worker_keepalive_network_test.go") != debug {
					t.Fatalf("sender stack debug policy changed: %q", messages)
				}
			})
		}
	}
}
