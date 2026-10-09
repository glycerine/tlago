package tlc

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
)

// DistributedCoordinatorNetwork connects TLCServer's existing publication
// lifecycle to named Go TCP objects. The listener opens at CreateRegistry,
// after recovery and hostname resolution, rather than during construction.
// The caller owns Close, including early initialization-failure returns where
// the source deliberately skips normal unbind/unexport cleanup.
type DistributedCoordinatorNetwork struct {
	Host           *DistributedRPCServer
	mu             sync.Mutex
	bindHost       string
	advertisedHost string
	Address        string
	registry       *TLCServerRegistry
	done           chan error
	closed         bool
	closeOnce      sync.Once
	closeError     error
}

func NewDistributedCoordinatorNetwork(bindHost, advertisedHost string) *DistributedCoordinatorNetwork {
	return &DistributedCoordinatorNetwork{Host: NewDistributedRPCServer(), bindHost: bindHost, advertisedHost: advertisedHost}
}
func (n *DistributedCoordinatorNetwork) hostname() (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.advertisedHost == "" {
		host, err := distributedLocalHostName()
		if err != nil {
			return "", err
		}
		n.advertisedHost = host
	}
	return n.advertisedHost, nil
}
func (n *DistributedCoordinatorNetwork) createRegistry(port int) (*TLCServerRegistry, error) {
	host, err := n.hostname()
	if err != nil {
		return nil, err
	}
	if host == "0.0.0.0" || host == "::" {
		return nil, errors.New("coordinator advertised host must be reachable, not a wildcard")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed {
		return nil, net.ErrClosed
	}
	if n.registry != nil {
		return nil, coordinatorPublicationFailure("coordinator listener is already created")
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(n.bindHost, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port
	n.Address = net.JoinHostPort(host, strconv.Itoa(actualPort))
	n.registry = &TLCServerRegistry{
		Rebind: func(name string, server *TLCServer) error {
			if server == nil {
				return NewNullPointerException()
			}
			endpoint := NewLocalServerEndpoint(server)
			n.Host.mu.Lock()
			defer n.Host.mu.Unlock()
			if n.Host.closed {
				return net.ErrClosed
			}
			n.Host.bindCoordinatorLocked(name, endpoint, n.Address)
			return nil
		},
		Unbind: func(name string) error {
			n.Host.mu.Lock()
			defer n.Host.mu.Unlock()
			if n.Host.closed {
				return net.ErrClosed
			}
			if _, found := n.Host.coordinators[name]; !found {
				return coordinatorBindingMissingFailure(name)
			}
			delete(n.Host.coordinators, name)
			return nil
		},
		Lookup: func(name string) (*TLCServer, error) {
			n.Host.mu.Lock()
			defer n.Host.mu.Unlock()
			if n.Host.closed {
				return nil, workerConnectionFailure(net.ErrClosed)
			}
			binding, found := n.Host.coordinators[name]
			if !found {
				return nil, coordinatorBindingMissingFailure(name)
			}
			endpoint, ok := binding.endpoint.(*LocalServerEndpoint)
			if !ok {
				return nil, fmt.Errorf("coordinator binding %q is not owned by this process", name)
			}
			return endpoint.Server, nil
		},
	}
	n.done = make(chan error, 1)
	go func() { n.done <- n.Host.Serve(listener) }()
	return n.registry, nil
}
func (n *DistributedCoordinatorNetwork) Publication() TLCServerPublication {
	return TLCServerPublication{
		LocalHostName: n.hostname, CreateRegistry: n.createRegistry,
		GetRegistry: func(int) (*TLCServerRegistry, error) {
			n.mu.Lock()
			defer n.mu.Unlock()
			if n.closed || n.registry == nil {
				return nil, workerConnectionFailure(net.ErrClosed)
			}
			return n.registry, nil
		},
		Unexport: func(server *TLCServer, force bool) (bool, error) {
			if server == nil {
				return false, NewNullPointerException()
			}
			if !server.unexported.CompareAndSwap(false, true) {
				return false, coordinatorEndpointRemovedFailure()
			}
			// Calls that already captured an endpoint may finish. Unpublishing
			// does not close the listener or cancel another object's calls.
			n.Host.mu.Lock()
			defer n.Host.mu.Unlock()
			for name, binding := range n.Host.coordinators {
				if endpoint, ok := binding.endpoint.(*LocalServerEndpoint); ok && endpoint.Server == server {
					delete(n.Host.coordinators, name)
				}
			}
			for object, binding := range n.Host.coordinatorObjects {
				if endpoint, ok := binding.endpoint.(*LocalServerEndpoint); ok && endpoint.Server == server {
					delete(n.Host.coordinatorObjects, object)
				}
			}
			return true, nil
		},
	}
}
func (n *DistributedCoordinatorNetwork) Close() error {
	n.closeOnce.Do(func() {
		n.mu.Lock()
		n.closed = true
		done := n.done
		n.mu.Unlock()
		n.closeError = n.Host.Close()
		if done != nil {
			if serveErr := <-done; !distributedCloseIsBenign(serveErr) {
				n.closeError = errors.Join(n.closeError, serveErr)
			}
		}
	})
	return n.closeError
}
