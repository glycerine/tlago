package tlc

import (
	"reflect"
	"testing"
)

func malformedNativeCoordinatorLocations() []string {
	return []string{
		"tcp://host:1/%xx", "rmi://host:1/main", "tcp://host:1",
		"tcp://host:1/main/extra", "tcp://user@host:1/main",
		"tcp://host:1/main?query=yes", "tcp://host:1/main#fragment",
	}
}

// No enabled Java test covers native location validation. TLCTimerTask.run
// catches malformed locations separately from coordinator connection failures.
func TestNativeDiscoveryMalformedLocationCategory(t *testing.T) {
	discovery := NewDistributedNetworkDiscovery()
	t.Cleanup(func() { _ = discovery.Close() })
	for _, location := range malformedNativeCoordinatorLocations() {
		endpoint, err := discovery.Lookup(location)
		if failure, ok := err.(*DistributedLocationError); !ok || failure == nil || failure.Location != location || failure.Cause == nil || endpoint != nil {
			t.Errorf("location %q returned %T/%v, want malformed-location category", location, err, err)
		}
	}
	if len(discovery.clients) != 0 {
		t.Fatal("malformed locations opened discovery connections")
	}
}

func TestNativeWorkerKeepAliveMalformedLocation(t *testing.T) {
	for _, location := range malformedNativeCoordinatorLocations() {
		t.Run(location, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			discovery := NewDistributedNetworkDiscovery()
			t.Cleanup(func() { _ = discovery.Close() })
			worker := NewDistributedWorker(0, nil, NewDistributedFPSetManager(), DistributedWorkerAddress{Hostname: "worker", Port: 1234})
			runtime := worker.Runtime
			var messages []string
			var failures []error
			runtime.ConfigureKeepAliveLookup(location, NewTLCServerStatusLookup(discovery.Lookup), func(message string, failure error) {
				messages = append(messages, message)
				failures = append(failures, failure)
			})
			runtime.StartKeepAlive(nil)
			t.Cleanup(func() { _ = worker.Exit() })
			for range 2 {
				if err := runtime.RunKeepAliveOnce(); err != nil {
					t.Fatalf("malformed location escaped keepalive catch: %T/%v", err, err)
				}
			}
			if !reflect.DeepEqual(messages, []string{"Failed to exit worker", "Failed to exit worker"}) {
				t.Fatalf("finest log messages = %q", messages)
			}
			for _, failure := range failures {
				if malformed, ok := failure.(*DistributedLocationError); !ok || malformed == nil || malformed.Location != location || malformed.Cause == nil {
					t.Fatalf("logger lost malformed-location category: %T", failure)
				}
			}
			if worker.unexported.Load() || runtime.executor.IsShutdown() || len(ToolIOGetAllMessages()) != 0 {
				t.Fatal("malformed location exited or diagnosed worker as a lost coordinator")
			}
			select {
			case <-runtime.keepAlive.done:
				t.Fatal("malformed location cancelled keepalive")
			default:
			}
			select {
			case <-runtime.latch.Load().done:
				t.Fatal("malformed location released worker completion latch")
			default:
			}
		})
	}
}
