package tlc

import (
	"errors"
	"fmt"
	"net/rpc"
	"net/url"
	"sort"
	"strings"
	"sync"
	"syscall"
)

// Lookup tests binding presence without calling any coordinator setting/status
// method. A listener can be reachable before either TLC binding is published.
type DistributedCoordinatorLookupReply struct{ Present bool }

// DistributedLocationError distinguishes invalid native coordinator locations
// from connection loss. Keepalive logs this category and continues, matching
// the source malformed-location catch without exposing a Java transport type.
type DistributedLocationError struct {
	Location string
	Cause    error
}

func (e *DistributedLocationError) Error() string { return e.Cause.Error() }
func (e *DistributedLocationError) Unwrap() error { return e.Cause }

func isDistributedMalformedLocation(err error) bool {
	switch failure := err.(type) {
	case *DistributedLocationError:
		return failure != nil
	case *MalformedURLException:
		return failure != nil
	default:
		return false
	}
}

func (service *distributedServerService) Lookup(name string, reply *DistributedCoordinatorLookupReply) error {
	service.server.mu.Lock()
	_, reply.Present = service.server.coordinators[name]
	service.server.mu.Unlock()
	return nil
}
func (s *DistributedRPCServer) UnregisterCoordinator(name string) {
	s.mu.Lock()
	delete(s.coordinators, name)
	s.mu.Unlock()
}

// DistributedNetworkDiscovery owns the TCP connections used by bootstrap and
// repeated keepalive lookups. Failed binding probes reuse a reachable listener;
// failed connections are discarded. Lookup never retries or sleeps itself.
type DistributedNetworkDiscovery struct {
	mu      sync.Mutex
	closed  bool
	clients map[string]*NetworkServerEndpoint
}

func NewDistributedNetworkDiscovery() *DistributedNetworkDiscovery {
	return &DistributedNetworkDiscovery{clients: make(map[string]*NetworkServerEndpoint)}
}
func (d *DistributedNetworkDiscovery) Lookup(location string) (DistributedServerEndpoint, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, &DistributedLocationError{Location: location, Cause: err}
	}
	if u.Scheme != "" && u.Scheme != "tcp" {
		return nil, &DistributedLocationError{Location: location, Cause: fmt.Errorf("unsupported coordinator scheme %q", u.Scheme)}
	}
	name := strings.TrimPrefix(u.Path, "/")
	if u.Host == "" || name == "" || strings.Contains(name, "/") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, &DistributedLocationError{Location: location, Cause: fmt.Errorf("coordinator location must identify a TCP address and binding: %q", location)}
	}
	key := u.Host + "/" + name
	d.mu.Lock()
	client, closed := d.clients[key], d.closed
	d.mu.Unlock()
	if closed {
		return nil, workerConnectionFailure(rpc.ErrShutdown)
	}
	if client == nil {
		candidate, err := DialServerEndpoint(u.Host, name)
		if err != nil {
			if operation, ok := err.(*DistributedOperationError); ok {
				operation.DiscoveryRetry = errors.Is(operation.Cause, syscall.ECONNREFUSED)
			}
			return nil, err
		}
		d.mu.Lock()
		if d.closed {
			d.mu.Unlock()
			_ = candidate.CloseConnection()
			return nil, workerConnectionFailure(rpc.ErrShutdown)
		}
		client = d.clients[key]
		if client == nil {
			client = candidate
			d.clients[key] = client
		}
		d.mu.Unlock()
		if client != candidate {
			_ = candidate.CloseConnection()
		}
	}
	var reply DistributedCoordinatorLookupReply
	if err := client.client.Call("Coordinator.Lookup", name, &reply); err != nil {
		d.mu.Lock()
		if d.clients[key] == client {
			delete(d.clients, key)
		}
		d.mu.Unlock()
		_ = client.CloseConnection()
		return nil, workerConnectionFailure(err)
	}
	if !reply.Present {
		return nil, &DistributedOperationError{Message: javaString("coordinator binding is not ready: " + name), Remote: true, IO: true, DiscoveryRetry: true, Reachable: true}
	}
	return client, nil
}
func (d *DistributedNetworkDiscovery) Close() error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	clients := d.clients
	d.clients = nil
	d.mu.Unlock()
	var failures []error
	keys := make([]string, 0, len(clients))
	for key := range clients {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		client := clients[key]
		if err := client.CloseConnection(); err != nil && !errors.Is(err, rpc.ErrShutdown) {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
