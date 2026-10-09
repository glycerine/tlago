package tlc

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/rpc"
	"sync"
)

func (s *DistributedRPCServer) RegisterWorker(name string, endpoint DistributedWorkerEndpoint) error {
	if name == "" || endpoint == nil {
		return errors.New("worker name and endpoint are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return net.ErrClosed
	}
	if _, present := s.workers[name]; present {
		return fmt.Errorf("worker endpoint %q is already registered", name)
	}
	s.workers[name] = endpoint
	return nil
}

// UnregisterWorker removes a published endpoint without closing its connections
// or cancelling an in-flight computation. Worker runtime owns computation exit.
func (s *DistributedRPCServer) UnregisterWorker(name string) {
	s.mu.Lock()
	delete(s.workers, name)
	s.mu.Unlock()
}

type DistributedWorkerRequest struct {
	Object    string
	Operation string
	States    *DistributedStatePayload
}
type DistributedWorkerReply struct {
	Result        *DistributedResultPayload
	Alive         bool
	URI           string
	CacheRateBits uint64
	Failure       *DistributedFailurePayload
}

type distributedWorkerService struct{ server *DistributedRPCServer }

func (service *distributedWorkerService) Call(request DistributedWorkerRequest, reply *DistributedWorkerReply) (callErr error) {
	var failure error
	defer func() {
		if caught := recover(); caught != nil {
			failure = panicValueAsError(caught)
		}
		if failure != nil {
			var err error
			reply.Failure, err = encodeDistributedRPCFailure(failure)
			if err != nil {
				callErr = fmt.Errorf("encode worker failure: %w", err)
			}
		}
	}()
	service.server.mu.Lock()
	endpoint := service.server.workers[request.Object]
	service.server.mu.Unlock()
	if endpoint == nil {
		failure = workerEndpointRemovedFailure("worker endpoint is not available: " + request.Object)
		return nil
	}
	switch request.Operation {
	case "next":
		var states []*TLCStateMut
		states, failure = DecodeDistributedStates(request.States)
		if failure != nil {
			failure = workerCodecFailure(failure)
			return nil
		}
		var result *NextStateResult
		result, failure = endpoint.GetNextStates(states)
		if failure != nil {
			return nil
		}
		reply.Result, failure = EncodeDistributedResult(result)
		if failure != nil {
			failure = workerCodecFailure(failure)
		}
	case "alive":
		reply.Alive, failure = endpoint.IsAlive()
	case "uri":
		reply.URI, failure = endpoint.GetURI()
	case "cache":
		var ratio float64
		ratio, failure = endpoint.GetCacheRateRatio()
		reply.CacheRateBits = math.Float64bits(ratio)
	case "exit":
		failure = endpoint.Exit()
		if failure == nil {
			service.server.UnregisterWorker(request.Object)
		}
	default:
		failure = fmt.Errorf("unknown worker operation %q", request.Operation)
	}
	return nil
}

// NetworkWorkerEndpoint calls a named Go worker over TCP. It never retries a
// call itself; the coordinator owns retry/worker-loss decisions and its queue.
type NetworkWorkerEndpoint struct {
	connectionMu        sync.Mutex
	connectionClosed    bool
	connectionCloseOnce sync.Once
	connectionCloseErr  error
	client              *rpc.Client
	Address             string
	Object              string
}

func DialWorkerEndpoint(address, object string) (*NetworkWorkerEndpoint, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, workerConnectionFailure(err)
	}
	return &NetworkWorkerEndpoint{client: client, Address: address, Object: object}, nil
}
func (e *NetworkWorkerEndpoint) CloseConnection() error {
	e.connectionCloseOnce.Do(func() {
		e.connectionMu.Lock()
		e.connectionClosed = true
		client := e.client
		e.connectionMu.Unlock()
		if client != nil {
			if err := client.Close(); !distributedCloseIsBenign(err) {
				e.connectionCloseErr = err
			}
		}
	})
	return e.connectionCloseErr
}

// Registration receives an object reference. Dial only when TLC invokes the
// callback, after its queue wake and other preceding source-owned work. A late
// dial cannot outlive owner closure; an established failed call is never replayed.
func (e *NetworkWorkerEndpoint) clientForCall() (*rpc.Client, error) {
	e.connectionMu.Lock()
	client, closed := e.client, e.connectionClosed
	e.connectionMu.Unlock()
	if closed {
		return nil, rpc.ErrShutdown
	}
	if client != nil {
		return client, nil
	}
	candidate, err := rpc.Dial("tcp", e.Address)
	if err != nil {
		return nil, err
	}
	e.connectionMu.Lock()
	if e.connectionClosed {
		e.connectionMu.Unlock()
		_ = candidate.Close()
		return nil, rpc.ErrShutdown
	}
	client = e.client
	if client == nil {
		e.client = candidate
		client = candidate
	}
	e.connectionMu.Unlock()
	if client != candidate {
		_ = candidate.Close()
	}
	return client, nil
}
func workerConnectionFailure(err error) error {
	var network *net.OpError
	deadWorker := errors.As(err, &network) || errors.Is(err, net.ErrClosed) || err == rpc.ErrShutdown || err == io.EOF || err == io.ErrUnexpectedEOF
	return newDistributedOperationError(DistributedOperationError{
		Message: javaString(err.Error()), Class: fmt.Sprintf("%T", err), Cause: err,
		Remote: true, IO: true,
		ExitIgnorable:     deadWorker,
		WorkerUnavailable: deadWorker,
		// The original coordinator treats a reply ending abruptly without an
		// EOF detail message as a candidate for a smaller computation block.
		// Connection refusal/closure and RPC application errors are not this.
		Recoverable: err == io.EOF || err == io.ErrUnexpectedEOF,
	})
}

// Native graph validation/representation errors belong to the remote I/O
// boundary. Source exceptions raised while materializing a value retain their
// application category; serialization does not turn those into connection loss.
func workerCodecFailure(err error) error {
	var source interface{ GetMessage() *string }
	if errors.As(err, &source) && !isJavaIOException(err) {
		return err
	}
	return workerConnectionFailure(err)
}
func (e *NetworkWorkerEndpoint) call(request DistributedWorkerRequest) (DistributedWorkerReply, error) {
	request.Object = e.Object
	var reply DistributedWorkerReply
	client, err := e.clientForCall()
	if err != nil {
		return reply, workerConnectionFailure(err)
	}
	if err := client.Call("Worker.Call", request, &reply); err != nil {
		closeFailedDistributedClient(client, err)
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
func (e *NetworkWorkerEndpoint) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	payload, err := EncodeDistributedStates(states)
	if err != nil {
		return nil, workerCodecFailure(err)
	}
	reply, err := e.call(DistributedWorkerRequest{Operation: "next", States: payload})
	if err != nil {
		return nil, err
	}
	result, err := DecodeDistributedResult(reply.Result)
	if err != nil {
		return nil, workerConnectionFailure(err)
	}
	return result, nil
}
func (e *NetworkWorkerEndpoint) IsAlive() (bool, error) {
	reply, err := e.call(DistributedWorkerRequest{Operation: "alive"})
	return reply.Alive, err
}
func (e *NetworkWorkerEndpoint) GetURI() (string, error) {
	reply, err := e.call(DistributedWorkerRequest{Operation: "uri"})
	return reply.URI, err
}
func (e *NetworkWorkerEndpoint) GetCacheRateRatio() (float64, error) {
	reply, err := e.call(DistributedWorkerRequest{Operation: "cache"})
	return math.Float64frombits(reply.CacheRateBits), err
}
func (e *NetworkWorkerEndpoint) Exit() error {
	_, err := e.call(DistributedWorkerRequest{Operation: "exit"})
	return err
}

var _ DistributedWorkerEndpoint = (*NetworkWorkerEndpoint)(nil)
