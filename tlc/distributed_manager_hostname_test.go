package tlc

import (
	"net"
	"strings"
	"testing"
	"time"
)

// Java has no direct manager hostname test. Use the real shared native host
// cache and a real failover instead of substituting the warning formatter.
func TestDistributedManagerFailoverUsesSharedLocalHost(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	cached := &distributedLocalHost{address: net.IPv4(127, 0, 0, 1), hostName: "cached-manager-host", expires: time.Now().Add(time.Minute)}
	previous := distributedCachedLocalHost.Swap(cached)
	t.Cleanup(func() { distributedCachedLocalHost.Store(previous) })
	failed := &fingerprintGraphFailureEndpoint{
		LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()),
		failure:                  NewIOException("host diagnostic failover"),
	}
	storage := NewMemFPSet()
	manager := NewDynamicDistributedFPSetManager(2)
	if err := manager.RegisterFPSet(failed, "failed-fp-host"); err != nil {
		t.Fatal(err)
	}
	if err := manager.RegisterFPSet(NewLocalFingerprintEndpoint(storage), "healthy-fp-host"); err != nil {
		t.Fatal(err)
	}
	if manager.GetHostName() != cached.hostName {
		t.Fatal("manager bypassed the shared native local-host cache")
	}
	if manager.Put(2) || !storage.Contains(2) || manager.NumOfAliveServers() != 1 {
		t.Fatal("hostname reporting changed fingerprint reassignment or insertion")
	}
	output := strings.Join(ToolIOGetAllMessages(), "\n")
	if !strings.Contains(output, "Warning: Failed to connect from cached-manager-host to the fp server at failed-fp-host.\nhost diagnostic failover") || strings.Contains(output, "no fp server available") {
		t.Fatalf("failover lost the shared local hostname or original cause: %q", output)
	}
	local := NewNonDistributedFPSetManager(nil, "captured-storage-host", nil)
	snapshot := manager.snapshotForWorker()
	distributedCachedLocalHost.Store(&distributedLocalHost{address: net.IPv4(127, 0, 0, 1), hostName: "replacement-local-host", expires: time.Now().Add(time.Minute)})
	if manager.GetHostName() != "replacement-local-host" || snapshot.GetHostName() != "replacement-local-host" || local.GetHostName() != "captured-storage-host" {
		t.Fatal("manager froze the caller hostname or replaced local storage's captured hostname")
	}
}
