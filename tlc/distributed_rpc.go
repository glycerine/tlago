package tlc

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
)

// DistributedEndpointError carries a remote operation failure. IO distinguishes
// storage/connection failures from an unchecked operation failure, preserving
// the existing TLC manager's different catch behavior for local FP managers.
type DistributedEndpointError struct {
	Message string
	IO      bool
}

func (e *DistributedEndpointError) Error() string { return e.Message }

// DistributedRPCServer hosts named TLC objects using Go's native RPC protocol.
// Calls are independent and concurrent. It performs no implicit retries: a
// failed insertion may already have changed storage, so TLC owns retry policy.
type DistributedRPCServer struct {
	rpc          *rpc.Server
	mu           sync.Mutex
	closed       bool
	listeners    map[net.Listener]struct{}
	connections  map[net.Conn]struct{}
	fingerprints map[string]DistributedFingerprintEndpoint
	workers      map[string]DistributedWorkerEndpoint
}

func NewDistributedRPCServer() *DistributedRPCServer {
	s := &DistributedRPCServer{
		rpc: rpc.NewServer(), listeners: make(map[net.Listener]struct{}),
		connections:  make(map[net.Conn]struct{}),
		fingerprints: make(map[string]DistributedFingerprintEndpoint),
		workers:      make(map[string]DistributedWorkerEndpoint),
	}
	if err := s.rpc.RegisterName("Fingerprint", &distributedFingerprintService{server: s}); err != nil {
		panic(err)
	}
	if err := s.rpc.RegisterName("Worker", &distributedWorkerService{server: s}); err != nil {
		panic(err)
	}
	return s
}

func (s *DistributedRPCServer) RegisterFingerprint(name string, endpoint DistributedFingerprintEndpoint) error {
	if name == "" || endpoint == nil {
		return errors.New("fingerprint name and endpoint are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return net.ErrClosed
	}
	if _, present := s.fingerprints[name]; present {
		return fmt.Errorf("fingerprint endpoint %q is already registered", name)
	}
	s.fingerprints[name] = endpoint
	return nil
}

// Serve owns accepted connections, but does not own fingerprint storage.
// Close shuts down listeners/connections; it does not call FPSet.Exit or retry
// an interrupted call. A caller can host several independent listeners.
func (s *DistributedRPCServer) Serve(listener net.Listener) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = listener.Close()
		return net.ErrClosed
	}
	s.listeners[listener] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.listeners, listener)
		s.mu.Unlock()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return nil
			}
			return err
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			_ = conn.Close()
			return nil
		}
		s.connections[conn] = struct{}{}
		s.mu.Unlock()
		go func() {
			defer func() {
				_ = conn.Close()
				s.mu.Lock()
				delete(s.connections, conn)
				s.mu.Unlock()
			}()
			s.rpc.ServeConn(conn)
		}()
	}
}

func (s *DistributedRPCServer) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	listeners := make([]net.Listener, 0, len(s.listeners))
	for listener := range s.listeners {
		listeners = append(listeners, listener)
	}
	connections := make([]net.Conn, 0, len(s.connections))
	for conn := range s.connections {
		connections = append(connections, conn)
	}
	s.mu.Unlock()
	var failures []error
	for _, listener := range listeners {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			failures = append(failures, err)
		}
	}
	for _, conn := range connections {
		if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

// Exported request/reply fields are the native Go wire format, not Java object
// serialization. Fingerprints retain all 64 bits and bit vectors retain their
// backing words, including an empty vector distinct from a null response.
type DistributedFingerprintRequest struct {
	Object        string
	Operation     string
	Fingerprint   uint64
	Fingerprints  []int64
	VectorPresent bool
	Expected      []uint64
	Filename      string
	Cleanup       bool
}

type DistributedFingerprintReply struct {
	Bool          bool
	Count         uint64
	Words         []uint64
	VectorPresent bool
	Failure       *DistributedEndpointError
}

type distributedFingerprintService struct{ server *DistributedRPCServer }

func (service *distributedFingerprintService) Call(request DistributedFingerprintRequest, reply *DistributedFingerprintReply) (err error) {
	// An endpoint panic must not take down the RPC process. Remote fatal
	// failures are endpoint I/O failures to callers, as in the original remote
	// invocation boundary. Ordinary runtime failures retain their category.
	defer func() {
		if failure := recover(); failure != nil {
			cause := panicValueAsError(failure)
			reply.Failure = &DistributedEndpointError{Message: cause.Error(), IO: isJavaError(cause) || isJavaIOException(cause)}
		}
	}()
	service.server.mu.Lock()
	endpoint := service.server.fingerprints[request.Object]
	service.server.mu.Unlock()
	if endpoint == nil {
		reply.Failure = &DistributedEndpointError{Message: "fingerprint endpoint is not available: " + request.Object, IO: true}
		return nil
	}
	var vector *LongVec
	if request.VectorPresent {
		vector = NewLongVecFrom(request.Fingerprints)
	}
	var bits *BitVector
	var failure error
	switch request.Operation {
	case "put":
		reply.Bool, failure = endpoint.Put(request.Fingerprint)
	case "contains":
		reply.Bool, failure = endpoint.Contains(request.Fingerprint)
	case "putBlock":
		bits, failure = endpoint.PutBlock(vector)
	case "containsBlock":
		bits, failure = endpoint.ContainsBlock(vector)
	case "size":
		reply.Count, failure = endpoint.Size()
	case "statesSeen":
		reply.Count, failure = endpoint.GetStatesSeen()
	case "checkFPs":
		reply.Count, failure = endpoint.CheckFPs()
	case "checkInvariant":
		reply.Bool, failure = endpoint.CheckInvariant(request.Expected...)
	case "addThread":
		failure = endpoint.AddThread()
	case "close":
		failure = endpoint.Close()
	case "exit":
		failure = endpoint.Exit(request.Cleanup)
	case "begin":
		failure = endpoint.BeginChkpt()
	case "beginFile":
		failure = endpoint.BeginChkptFile(request.Filename)
	case "commit":
		failure = endpoint.CommitChkpt()
	case "commitFile":
		failure = endpoint.CommitChkptFile(request.Filename)
	case "recoverFile":
		failure = endpoint.RecoverFile(request.Filename)
	default:
		failure = fmt.Errorf("unknown fingerprint operation %q", request.Operation)
	}
	if failure != nil {
		reply.Failure = &DistributedEndpointError{Message: failure.Error(), IO: isJavaIOException(failure)}
	}
	if bits != nil {
		reply.VectorPresent = true
		reply.Words = append([]uint64(nil), bits.word...)
	}
	return nil
}

type NetworkFingerprintEndpoint struct {
	client  *rpc.Client
	Address string
	Object  string
}

func DialFingerprintEndpoint(address, object string) (*NetworkFingerprintEndpoint, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, &DistributedEndpointError{Message: err.Error(), IO: true}
	}
	return &NetworkFingerprintEndpoint{client: client, Address: address, Object: object}, nil
}

// CloseConnection closes only this client's connection, independently of the
// remote storage Close/Exit lifecycle. No call is retried after a disconnect.
func (e *NetworkFingerprintEndpoint) CloseConnection() error { return e.client.Close() }

func (e *NetworkFingerprintEndpoint) call(request DistributedFingerprintRequest) (DistributedFingerprintReply, error) {
	request.Object = e.Object
	var reply DistributedFingerprintReply
	if err := e.client.Call("Fingerprint.Call", request, &reply); err != nil {
		return reply, &DistributedEndpointError{Message: err.Error(), IO: true}
	}
	if reply.Failure != nil {
		return reply, reply.Failure
	}
	return reply, nil
}

func (e *NetworkFingerprintEndpoint) Put(fp uint64) (bool, error) {
	r, err := e.call(DistributedFingerprintRequest{Operation: "put", Fingerprint: fp})
	return r.Bool, err
}
func (e *NetworkFingerprintEndpoint) Contains(fp uint64) (bool, error) {
	r, err := e.call(DistributedFingerprintRequest{Operation: "contains", Fingerprint: fp})
	return r.Bool, err
}
func (e *NetworkFingerprintEndpoint) block(operation string, fps *LongVec) (*BitVector, error) {
	request := DistributedFingerprintRequest{Operation: operation, VectorPresent: fps != nil}
	if fps != nil {
		request.Fingerprints = append([]int64(nil), fps.data...)
	}
	r, err := e.call(request)
	if err != nil || !r.VectorPresent {
		return nil, err
	}
	return &BitVector{word: r.Words}, nil
}
func (e *NetworkFingerprintEndpoint) PutBlock(fps *LongVec) (*BitVector, error) {
	return e.block("putBlock", fps)
}
func (e *NetworkFingerprintEndpoint) ContainsBlock(fps *LongVec) (*BitVector, error) {
	return e.block("containsBlock", fps)
}
func (e *NetworkFingerprintEndpoint) count(operation string) (uint64, error) {
	r, err := e.call(DistributedFingerprintRequest{Operation: operation})
	return r.Count, err
}
func (e *NetworkFingerprintEndpoint) Size() (uint64, error)          { return e.count("size") }
func (e *NetworkFingerprintEndpoint) GetStatesSeen() (uint64, error) { return e.count("statesSeen") }
func (e *NetworkFingerprintEndpoint) CheckFPs() (uint64, error)      { return e.count("checkFPs") }
func (e *NetworkFingerprintEndpoint) CheckInvariant(expected ...uint64) (bool, error) {
	r, err := e.call(DistributedFingerprintRequest{Operation: "checkInvariant", Expected: expected})
	return r.Bool, err
}
func (e *NetworkFingerprintEndpoint) operation(operation, filename string, cleanup bool) error {
	_, err := e.call(DistributedFingerprintRequest{Operation: operation, Filename: filename, Cleanup: cleanup})
	return err
}
func (e *NetworkFingerprintEndpoint) AddThread() error { return e.operation("addThread", "", false) }
func (e *NetworkFingerprintEndpoint) Close() error     { return e.operation("close", "", false) }
func (e *NetworkFingerprintEndpoint) Exit(cleanup bool) error {
	return e.operation("exit", "", cleanup)
}
func (e *NetworkFingerprintEndpoint) BeginChkpt() error { return e.operation("begin", "", false) }
func (e *NetworkFingerprintEndpoint) BeginChkptFile(name string) error {
	return e.operation("beginFile", name, false)
}
func (e *NetworkFingerprintEndpoint) CommitChkpt() error { return e.operation("commit", "", false) }
func (e *NetworkFingerprintEndpoint) CommitChkptFile(name string) error {
	return e.operation("commitFile", name, false)
}
func (e *NetworkFingerprintEndpoint) RecoverFile(name string) error {
	return e.operation("recoverFile", name, false)
}
func (e *NetworkFingerprintEndpoint) RecoverTrace(*TLCTrace) error {
	return errors.New("trace recovery runs at the coordinator; remote recovery uses a checkpoint file")
}

var _ DistributedFingerprintEndpoint = (*NetworkFingerprintEndpoint)(nil)
