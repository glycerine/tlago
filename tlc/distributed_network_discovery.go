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
type DistributedCoordinatorLookupReply struct {
	Present bool
	Object  string
}

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
	binding, present := service.server.coordinators[name]
	reply.Present, reply.Object = present, binding.object
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
	closeOnce sync.Once
	closeErr  error
	mu        sync.Mutex
	closed    bool
	clients   map[string]*NetworkServerEndpoint
	views     map[distributedCoordinatorViewKey]*NetworkServerEndpoint
}

type distributedCoordinatorViewKey struct {
	client *NetworkServerEndpoint
	object string
}

func NewDistributedNetworkDiscovery() *DistributedNetworkDiscovery {
	return &DistributedNetworkDiscovery{clients: make(map[string]*NetworkServerEndpoint), views: make(map[distributedCoordinatorViewKey]*NetworkServerEndpoint)}
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
		for viewKey := range d.views {
			if viewKey.client == client {
				delete(d.views, viewKey)
			}
		}
		d.mu.Unlock()
		_ = client.CloseConnection()
		return nil, workerConnectionFailure(err)
	}
	if !reply.Present {
		return nil, coordinatorBindingMissingFailure(name)
	}
	if reply.Object == "" {
		return nil, workerConnectionFailure(errors.New("coordinator lookup returned no endpoint identity"))
	}
	// Each lookup captures the currently bound object. Only the connection and
	// its owned fingerprint callbacks are shared with the discovery cache.
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil, workerConnectionFailure(rpc.ErrShutdown)
	}
	viewKey := distributedCoordinatorViewKey{client: client, object: reply.Object}
	view := d.views[viewKey]
	if view == nil {
		view = &NetworkServerEndpoint{client: client.client, Address: client.Address, Object: reply.Object, children: client.children}
		d.views[viewKey] = view
	}
	return view, nil
}
func (d *DistributedNetworkDiscovery) Close() error {
	d.closeOnce.Do(func() {
		d.mu.Lock()
		d.closed = true
		clients := d.clients
		d.clients = nil
		d.views = nil
		d.mu.Unlock()
		var failures []error
		keys := make([]string, 0, len(clients))
		for key := range clients {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			client := clients[key]
			if err := client.CloseConnection(); !distributedCloseIsBenign(err) {
				failures = append(failures, err)
			}
		}
		d.closeErr = errors.Join(failures...)
	})
	return d.closeErr
}
