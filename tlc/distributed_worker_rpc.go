package tlc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/rpc"
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
	Result    *DistributedResultPayload
	Alive     bool
	URI       string
	CacheRate float64
	Failure   *DistributedFailurePayload
}

type distributedWorkerService struct{ server *DistributedRPCServer }

func (service *distributedWorkerService) Call(request DistributedWorkerRequest, reply *DistributedWorkerReply) (callErr error) {
	var failure error
	defer func() {
		if caught := recover(); caught != nil {
			failure = panicValueAsError(caught)
			if isJavaError(failure) {
				failure = &DistributedOperationError{Message: javaString(failure.Error()), Class: javaThrowableClassName(failure), Stack: javaThrowableStackTrace(failure), Cause: failure, Remote: true, IO: true}
			}
		}
		if failure != nil {
			var err error
			reply.Failure, err = EncodeDistributedFailure(failure)
			if err != nil {
				callErr = fmt.Errorf("encode worker failure: %w", err)
			}
		}
	}()
	service.server.mu.Lock()
	endpoint := service.server.workers[request.Object]
	service.server.mu.Unlock()
	if endpoint == nil {
		failure = &DistributedOperationError{Message: javaString("worker endpoint is not available: " + request.Object), Remote: true, IO: true, ExitIgnorable: true}
		return nil
	}
	switch request.Operation {
	case "next":
		var states []*TLCStateMut
		states, failure = DecodeDistributedStates(request.States)
		if failure != nil {
			return nil
		}
		var result *NextStateResult
		result, failure = endpoint.GetNextStates(states)
		if failure != nil {
			return nil
		}
		reply.Result, failure = EncodeDistributedResult(result)
	case "alive":
		reply.Alive, failure = endpoint.IsAlive()
	case "uri":
		reply.URI, failure = endpoint.GetURI()
	case "cache":
		reply.CacheRate, failure = endpoint.GetCacheRateRatio()
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
	client  *rpc.Client
	Address string
	Object  string
}

func DialWorkerEndpoint(address, object string) (*NetworkWorkerEndpoint, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, workerConnectionFailure(err)
	}
	return &NetworkWorkerEndpoint{client: client, Address: address, Object: object}, nil
}
func (e *NetworkWorkerEndpoint) CloseConnection() error { return e.client.Close() }
func workerConnectionFailure(err error) error {
	var network *net.OpError
	deadWorker := errors.As(err, &network) || errors.Is(err, net.ErrClosed) || err == rpc.ErrShutdown || err == io.EOF || err == io.ErrUnexpectedEOF
	return &DistributedOperationError{
		Message: javaString(err.Error()), Class: fmt.Sprintf("%T", err), Cause: err,
		Remote: true, IO: true,
		ExitIgnorable: deadWorker,
		// The original coordinator treats a reply ending abruptly without an
		// EOF detail message as a candidate for a smaller computation block.
		// Connection refusal/closure and RPC application errors are not this.
		Recoverable: err == io.EOF || err == io.ErrUnexpectedEOF,
	}
}
func (e *NetworkWorkerEndpoint) call(request DistributedWorkerRequest) (DistributedWorkerReply, error) {
	request.Object = e.Object
	var reply DistributedWorkerReply
	if err := e.client.Call("Worker.Call", request, &reply); err != nil {
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
		return nil, err
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
	return reply.CacheRate, err
}
func (e *NetworkWorkerEndpoint) Exit() error {
	_, err := e.call(DistributedWorkerRequest{Operation: "exit"})
	return err
}

var _ DistributedWorkerEndpoint = (*NetworkWorkerEndpoint)(nil)
