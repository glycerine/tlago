package tlc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"reflect"
	"sort"
	"sync"
)

// Connections acquired for endpoint callbacks belong to their hosting process.
// Adding after shutdown closes the connection immediately, including a dial
// that overlapped shutdown. No RPC or dial is performed while holding this lock.
type distributedConnections struct {
	mu      sync.Mutex
	closed  bool
	clients []io.Closer
}

type distributedConnectionCloser func() error

func (close distributedConnectionCloser) Close() error { return close() }

func (c *distributedConnections) add(client io.Closer) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		_ = client.Close()
		return net.ErrClosed
	}
	c.clients = append(c.clients, client)
	return nil
}
func (c *distributedConnections) close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	clients := c.clients
	c.clients = nil
	c.mu.Unlock()
	var failures []error
	for _, client := range clients {
		if err := client.Close(); !distributedCloseIsBenign(err) {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

// An already-closed component is harmless only when it is the entire failure.
// errors.Is alone would also match one branch of a joined error and discard
// unrelated callback cleanup failures alongside it. Keep those original errors
// intact so their context and errors.Is/errors.As causes survive shutdown.
func distributedCloseIsBenign(err error) bool {
	if err == nil {
		return true
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, cause := range causes {
			if !distributedCloseIsBenign(cause) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		if cause := wrapped.Unwrap(); cause != nil {
			return distributedCloseIsBenign(cause)
		}
	}
	return err == rpc.ErrShutdown || errors.Is(err, net.ErrClosed)
}

type distributedCoordinatorBinding struct {
	endpoint DistributedServerEndpoint
	address  string
}

// address is the externally reachable address of this host's TCP listener.
func (s *DistributedRPCServer) RegisterCoordinator(name string, endpoint DistributedServerEndpoint, address string) error {
	if name == "" || endpoint == nil || address == "" {
		return errors.New("coordinator name, endpoint and advertised address are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return net.ErrClosed
	}
	if _, exists := s.coordinators[name]; exists {
		return fmt.Errorf("coordinator %q is already registered", name)
	}
	s.coordinators[name] = distributedCoordinatorBinding{endpoint: endpoint, address: address}
	return nil
}
func (s *DistributedRPCServer) fingerprintReference(endpoint DistributedFingerprintEndpoint, address string) (DistributedEndpointReference, error) {
	if network, ok := endpoint.(*NetworkFingerprintEndpoint); ok {
		return DistributedEndpointReference{network.Address, network.Object}, nil
	}
	if endpoint == nil {
		return DistributedEndpointReference{}, errors.New("nil fingerprint endpoint")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return DistributedEndpointReference{}, net.ErrClosed
	}
	// Reuse a publication for the same endpoint or owning local storage.
	names := make([]string, 0, len(s.fingerprints))
	for name := range s.fingerprints {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		existing := s.fingerprints[name]
		equal := reflect.TypeOf(endpoint).Comparable() && endpoint == existing
		if a, ok := endpoint.(*LocalFingerprintEndpoint); ok {
			if b, ok := existing.(*LocalFingerprintEndpoint); ok && reflect.TypeOf(a.Set).Comparable() {
				equal = a.Set == b.Set
			}
		}
		if equal {
			return DistributedEndpointReference{address, name}, nil
		}
	}
	name := ""
	for {
		s.fingerprintSequence++
		name = s.generatedEndpointName("coordinator-fp", s.fingerprintSequence)
		if _, exists := s.fingerprints[name]; !exists {
			break
		}
	}
	s.fingerprints[name] = endpoint
	return DistributedEndpointReference{address, name}, nil
}

type DistributedServerRequest struct {
	Object, Operation, Text, Hostname string
	Endpoint                          DistributedEndpointReference
}
type DistributedServerReply struct {
	Bool       bool
	Polynomial uint64
	Text       string
	Bytes      []byte
	BytesNil   bool
	String     *DistributedStringNode
	Manager    *DistributedManagerPayload
	Failure    *DistributedFailurePayload
}
type distributedServerService struct{ server *DistributedRPCServer }

func (service *distributedServerService) Call(request DistributedServerRequest, reply *DistributedServerReply) (callErr error) {
	var failure error
	defer func() {
		if caught := recover(); caught != nil {
			failure = panicValueAsError(caught)
		}
		if failure != nil {
			var err error
			reply.Failure, err = encodeDistributedRPCFailure(failure)
			if err != nil {
				callErr = fmt.Errorf("encode coordinator failure: %w", err)
			}
		}
	}()
	service.server.mu.Lock()
	binding := service.server.coordinators[request.Object]
	service.server.mu.Unlock()
	endpoint := binding.endpoint
	if endpoint == nil {
		failure = newDistributedOperationError(DistributedOperationError{Message: javaString("coordinator is not available: " + request.Object), Remote: true, IO: true})
		return nil
	}
	switch request.Operation {
	case "deadlock":
		reply.Bool, failure = endpoint.GetCheckDeadlock()
	case "preprocess":
		reply.Bool, failure = endpoint.GetPreprocess()
	case "polynomial":
		reply.Polynomial, failure = endpoint.GetIrredPolyForFP()
	case "done":
		reply.Bool, failure = endpoint.IsDone()
	case "spec":
		reply.Text, failure = endpoint.GetSpecFileName()
	case "config":
		reply.Text, failure = endpoint.GetConfigFileName()
	case "file":
		reply.Bytes, failure = endpoint.GetFile(request.Text)
		reply.BytesNil = reply.Bytes == nil
	case "intern":
		var value *UniqueString
		value, failure = endpoint.Intern(request.Text)
		if value != nil {
			reply.String = &DistributedStringNode{Text: value.s, Token: value.tok, Location: value.loc, Unregistered: value.unregistered}
		}
	case "manager":
		var manager *DistributedFPSetManager
		manager, failure = endpoint.GetFPSetManager()
		if failure == nil {
			reply.Manager, failure = encodeDistributedManager(manager, func(fp DistributedFingerprintEndpoint) (DistributedEndpointReference, error) {
				return service.server.fingerprintReference(fp, binding.address)
			})
		}
	case "registerWorker":
		if request.Endpoint.Address == "" || request.Endpoint.Object == "" {
			failure = workerConnectionFailure(errors.New("incomplete worker endpoint reference"))
			return nil
		}
		worker := &NetworkWorkerEndpoint{Address: request.Endpoint.Address, Object: request.Endpoint.Object}
		failure = service.server.outbound.add(distributedConnectionCloser(worker.CloseConnection))
		if failure == nil {
			failure = endpoint.RegisterWorker(worker)
		}
	case "registerFP":
		if request.Endpoint.Address == "" || request.Endpoint.Object == "" {
			failure = workerConnectionFailure(errors.New("incomplete fingerprint endpoint reference"))
			return nil
		}
		// Registration stores a reference and checks capacity without an
		// aliveness call. The manager contacts it only during an operation.
		fp := &NetworkFingerprintEndpoint{Address: request.Endpoint.Address, Object: request.Endpoint.Object}
		failure = service.server.outbound.add(distributedConnectionCloser(fp.CloseConnection))
		if failure == nil {
			failure = endpoint.RegisterFPSet(fp, request.Hostname)
		}
	default:
		failure = fmt.Errorf("unknown coordinator operation %q", request.Operation)
	}
	return nil
}

type NetworkServerEndpoint struct {
	client          *rpc.Client
	Address, Object string
	children        distributedConnections
}

func DialServerEndpoint(address, object string) (*NetworkServerEndpoint, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, workerConnectionFailure(err)
	}
	return &NetworkServerEndpoint{client: client, Address: address, Object: object}, nil
}
func (e *NetworkServerEndpoint) CloseConnection() error {
	return errors.Join(e.client.Close(), e.children.close())
}
func (e *NetworkServerEndpoint) call(request DistributedServerRequest) (DistributedServerReply, error) {
	request.Object = e.Object
	var reply DistributedServerReply
	if err := e.client.Call("Coordinator.Call", request, &reply); err != nil {
		return reply, workerConnectionFailure(err)
	}
	if reply.Failure != nil {
		failure, err := DecodeDistributedFailure(reply.Failure)
		if err != nil {
			return reply, workerConnectionFailure(err)
		}
		return reply, failure
	}
	return reply, nil
}
func (e *NetworkServerEndpoint) GetCheckDeadlock() (bool, error) {
	r, err := e.call(DistributedServerRequest{Operation: "deadlock"})
	return r.Bool, err
}
func (e *NetworkServerEndpoint) GetPreprocess() (bool, error) {
	r, err := e.call(DistributedServerRequest{Operation: "preprocess"})
	return r.Bool, err
}
func (e *NetworkServerEndpoint) GetIrredPolyForFP() (uint64, error) {
	r, err := e.call(DistributedServerRequest{Operation: "polynomial"})
	return r.Polynomial, err
}
func (e *NetworkServerEndpoint) IsDone() (bool, error) {
	r, err := e.call(DistributedServerRequest{Operation: "done"})
	return r.Bool, err
}
func (e *NetworkServerEndpoint) GetSpecFileName() (string, error) {
	r, err := e.call(DistributedServerRequest{Operation: "spec"})
	return r.Text, err
}
func (e *NetworkServerEndpoint) GetConfigFileName() (string, error) {
	r, err := e.call(DistributedServerRequest{Operation: "config"})
	return r.Text, err
}
func (e *NetworkServerEndpoint) GetFile(name string) ([]byte, error) {
	r, err := e.call(DistributedServerRequest{Operation: "file", Text: name})
	if err != nil || r.BytesNil {
		return nil, err
	}
	if r.Bytes == nil {
		return []byte{}, nil
	}
	return r.Bytes, nil
}
func (e *NetworkServerEndpoint) Intern(text string) (*UniqueString, error) {
	r, err := e.call(DistributedServerRequest{Operation: "intern", Text: text})
	if err != nil || r.String == nil {
		return nil, err
	}
	return &UniqueString{s: r.String.Text, tok: r.String.Token, loc: r.String.Location, unregistered: r.String.Unregistered}, nil
}
func (e *NetworkServerEndpoint) GetFPSetManager() (*DistributedFPSetManager, error) {
	r, err := e.call(DistributedServerRequest{Operation: "manager"})
	if err != nil {
		return nil, err
	}
	// Receiving references must not require every store to be alive. TLC's
	// manager owns availability/failover when an operation contacts a store.
	// Commit ownership only after the entire reference graph decodes.
	var acquired []*NetworkFingerprintEndpoint
	manager, err := decodeDistributedManager(r.Manager, func(ref DistributedEndpointReference) (DistributedFingerprintEndpoint, error) {
		fp := &NetworkFingerprintEndpoint{Address: ref.Address, Object: ref.Object}
		acquired = append(acquired, fp)
		return fp, nil
	})
	if err != nil {
		for _, fp := range acquired {
			_ = fp.CloseConnection()
		}
		return nil, err
	}
	for _, fp := range acquired {
		if err := e.children.add(distributedConnectionCloser(fp.CloseConnection)); err != nil {
			for _, other := range acquired {
				_ = other.CloseConnection()
			}
			return nil, err
		}
	}
	return manager, nil
}
func (e *NetworkServerEndpoint) RegisterWorker(worker DistributedWorkerEndpoint) error {
	network, ok := worker.(*NetworkWorkerEndpoint)
	if !ok {
		return fmt.Errorf("worker registration requires a published TCP endpoint, got %T", worker)
	}
	return e.RegisterWorkerReference(DistributedEndpointReference{network.Address, network.Object})
}

func (e *NetworkServerEndpoint) RegisterWorkerReference(reference DistributedEndpointReference) error {
	_, err := e.call(DistributedServerRequest{Operation: "registerWorker", Endpoint: reference})
	return err
}
func (e *NetworkServerEndpoint) RegisterFPSet(fp DistributedFingerprintEndpoint, hostname string) error {
	network, ok := fp.(*NetworkFingerprintEndpoint)
	if !ok {
		return fmt.Errorf("fingerprint registration requires a published TCP endpoint, got %T", fp)
	}
	return e.RegisterFPSetReference(DistributedEndpointReference{network.Address, network.Object}, hostname)
}

func (e *NetworkServerEndpoint) RegisterFPSetReference(reference DistributedEndpointReference, hostname string) error {
	_, err := e.call(DistributedServerRequest{Operation: "registerFP", Hostname: hostname, Endpoint: reference})
	return err
}

var _ DistributedServerEndpoint = (*NetworkServerEndpoint)(nil)
