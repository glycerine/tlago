package tlc

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// DistributedWorkerNetwork supplies native discovery, publication and callbacks
// to the existing worker command. The command still owns initialization order,
// evaluator/runtime lifetime, registration threads, keepalive and exit latch.
// Close drains networking and closes fingerprint storage published through
// FPEnvironment. Shut down the worker runtime separately. Storage published
// directly through Host remains caller-owned.
type DistributedWorkerNetwork struct {
	Host           *DistributedRPCServer
	Discovery      *DistributedNetworkDiscovery
	Address        string
	workerAddress  DistributedWorkerAddress
	sequence       atomic.Uint64
	fingerprintsMu sync.Mutex
	fingerprints   map[FPSet][]string
	ownedFPSets    map[FPSet]bool
	fpOwners       []FPSet
	done           chan error
	closeOnce      sync.Once
	closeError     error
}

func NewDistributedWorkerNetwork(listenAddress, advertisedAddress string) (*DistributedWorkerNetwork, error) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, err
	}
	if advertisedAddress == "" {
		advertisedAddress = listener.Addr().String()
	} else if net.ParseIP(advertisedAddress) != nil || !strings.Contains(advertisedAddress, ":") {
		// A host alone advertises the actual bound callback port, including
		// an ephemeral port chosen by the OS. Explicit host:port stays intact.
		advertisedAddress = net.JoinHostPort(advertisedAddress, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	}
	host, portText, err := net.SplitHostPort(advertisedAddress)
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port <= 0 || port > 65535 || host == "" || host == "0.0.0.0" || host == "::" {
		_ = listener.Close()
		return nil, fmt.Errorf("worker advertised address must be a reachable host and port: %q", advertisedAddress)
	}
	uriHost := host
	if strings.Contains(host, ":") {
		uriHost = "[" + host + "]"
	}
	n := &DistributedWorkerNetwork{Host: NewDistributedRPCServer(), Discovery: NewDistributedNetworkDiscovery(), Address: advertisedAddress, workerAddress: DistributedWorkerAddress{Hostname: uriHost, Port: port}, done: make(chan error, 1)}
	go func() { n.done <- n.Host.Serve(listener) }()
	return n, nil
}
func (n *DistributedWorkerNetwork) Environment(base DistributedWorkerEnvironment) DistributedWorkerEnvironment {
	base.Lookup = n.Discovery.Lookup
	base.LocalCanonicalHostName = func() (string, error) { return n.workerAddress.Hostname, nil }
	base.PublishWorker = func(worker *DistributedWorker) error {
		name := fmt.Sprintf("worker-%d-%d", worker.ID, n.sequence.Add(1))
		// Identity is complete before publication or the runnable's worker
		// pointer becomes visible to keepalive/shutdown.
		worker.uri = newDistributedWorkerEndpointURI(n.workerAddress, name)
		worker.networkReference = &DistributedEndpointReference{Address: n.Address, Object: name}
		return n.Host.RegisterWorker(name, NewLocalWorkerEndpoint(worker))
	}
	base.RegisterWorker = func(server DistributedServerEndpoint, worker *DistributedWorker) error {
		network, ok := server.(*NetworkServerEndpoint)
		if !ok || worker.networkReference == nil {
			return errors.New("native worker registration requires a discovered TCP coordinator and published worker")
		}
		return network.RegisterWorkerReference(*worker.networkReference)
	}
	return base
}
func (n *DistributedWorkerNetwork) Close() error {
	n.closeOnce.Do(func() {
		n.closeError = errors.Join(n.Host.CloseGracefully(), n.Discovery.Close())
		if err := <-n.done; err != nil && !errors.Is(err, net.ErrClosed) {
			n.closeError = errors.Join(n.closeError, err)
		}
		// Unpublication removes names, not native ownership. Close storage
		// only after accepted RPC calls and replies have drained. A failed
		// registration may have completed remotely, so it retains ownership
		// and publication until this same process-lifetime boundary.
		n.fingerprintsMu.Lock()
		owners := n.fpOwners
		n.fpOwners, n.ownedFPSets, n.fingerprints = nil, nil, nil
		n.fingerprintsMu.Unlock()
		for _, set := range owners {
			set.Close()
		}
	})
	return n.closeError
}
