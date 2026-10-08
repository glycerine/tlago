package tlc

import (
	"fmt"
	"reflect"
)

// DistributedOperationError carries a local endpoint, transport or received Go
// operation failure. Its traits drive TLC's catch/retry decisions without a Java remote
// object or exception protocol. Diagnostic text and causes belong to the sender.
type DistributedOperationError struct {
	Message             *string
	Class               string
	Stack               string
	Cause               error
	Suppressed          []error
	Remote              bool
	Recoverable         bool
	IO                  bool
	Null                bool
	ExitIgnorable       bool
	WorkerUnavailable   bool
	EndpointRemoved     bool
	BindingMissing      bool
	DiscoveryRetry      bool
	Reachable           bool
	FingerprintRejected bool
	pcs                 []uintptr
}

// Capture Go frames at the originating endpoint, not when a received payload
// is decoded. The existing stack field carries the sender's formatted frames.
func newDistributedOperationError(failure DistributedOperationError) *DistributedOperationError {
	failure.pcs = captureThrowableTrace().pcs
	if failure.Class == "" {
		failure.Class = "tlc.DistributedOperationError"
	}
	return &failure
}

func (e *DistributedOperationError) throwablePCs() []uintptr { return e.pcs }

func (e *DistributedOperationError) Error() string {
	if e.Message != nil {
		return *e.Message
	}
	return e.Class
}
func (e *DistributedOperationError) GetMessage() *string { return copyJavaMessage(e.Message) }
func (e *DistributedOperationError) Unwrap() error       { return e.Cause }
func (e *DistributedOperationError) GetSuppressed() []error {
	return append([]error{}, e.Suppressed...)
}
func (e *DistributedOperationError) diagnosticClassName() string  { return e.Class }
func (e *DistributedOperationError) remoteThrowableStack() string { return e.Stack }

// Fingerprint calls preserve their Go transport cause without worker retry or
// availability categories. Exit may mark its own lost reply as ignorable after
// this boundary returns; other operations retain these traits. No call is replayed.
func fingerprintConnectionFailure(cause error) *DistributedOperationError {
	return newDistributedOperationError(DistributedOperationError{
		Message: javaString(cause.Error()), Class: fmt.Sprintf("%T", cause),
		Cause: cause, Remote: true, IO: true,
	})
}

// Worker resource failures carry coordinator decisions directly. Exhausting
// worker memory can be retried with a smaller batch; rejected execution cannot.
// The original cause remains available without a fabricated transport envelope.
func workerComputationFailure(message string, cause error, recoverable bool) *DistributedOperationError {
	return newDistributedOperationError(DistributedOperationError{
		Message: javaString(message), Class: "tlc.DistributedWorkerFailure",
		Cause: cause, Remote: true, IO: true, Recoverable: recoverable,
		ExitIgnorable: true,
	})
}

func workerEndpointRemovedFailure(message string) *DistributedOperationError {
	return newDistributedOperationError(DistributedOperationError{
		Message: javaString(message), Class: "tlc.WorkerEndpointRemoved",
		Remote: true, IO: true, ExitIgnorable: true, WorkerUnavailable: true,
		EndpointRemoved: true,
	})
}

// Worker cleanup tolerates prior removal, but must not suppress a different
// connection or operation failure merely because it also means unavailable.
func isDistributedWorkerEndpointRemoved(err error) bool {
	failure, ok := err.(*DistributedOperationError)
	return ok && failure != nil && failure.EndpointRemoved && failure.WorkerUnavailable
}

func coordinatorBindingMissingFailure(name string) *DistributedOperationError {
	return newDistributedOperationError(DistributedOperationError{
		Message: javaString("coordinator binding is not ready: " + name),
		Class:   "tlc.CoordinatorBindingMissing", BindingMissing: true,
		DiscoveryRetry: true, Reachable: true,
	})
}

func isDistributedCoordinatorBindingMissing(err error) bool {
	failure, ok := err.(*DistributedOperationError)
	return ok && failure != nil && failure.BindingMissing
}

func coordinatorEndpointRemovedFailure() *DistributedOperationError {
	return newDistributedOperationError(DistributedOperationError{
		Message: javaString("coordinator endpoint already removed"),
		Class:   "tlc.CoordinatorEndpointRemoved", Remote: true, IO: true,
		EndpointRemoved: true,
	})
}

func coordinatorPublicationFailure(message string) *DistributedOperationError {
	return newDistributedOperationError(DistributedOperationError{
		Message: javaString(message), Class: "tlc.CoordinatorPublicationFailed",
		Remote: true, IO: true,
	})
}

// Failure references are one-based and can contain shared or cyclic causes.
// Worker failure states share one state/value graph across the entire failure.
type DistributedFailurePayload struct {
	Root   int
	Nodes  []DistributedFailureNode
	States *DistributedStatePayload
}
type DistributedFailureNode struct {
	Message             string
	MessageNil          bool
	Class               string
	Stack               string
	Cause               int
	Suppressed          []int
	Worker              bool
	State1              int
	State2              int
	KeepCallStack       bool
	Remote              bool
	Recoverable         bool
	IO                  bool
	Null                bool
	ExitIgnorable       bool
	WorkerUnavailable   bool
	EndpointRemoved     bool
	BindingMissing      bool
	DiscoveryRetry      bool
	Reachable           bool
	FingerprintRejected bool
}

func isDistributedFPRegistrationRejected(err error) bool {
	if failure, ok := err.(*DistributedOperationError); ok {
		return failure.FingerprintRejected
	}
	failure, ok := err.(*FPSetManagerException)
	return ok && failure != nil
}

func isIgnorableDistributedWorkerExit(err error) bool {
	failure, ok := err.(*DistributedOperationError)
	return ok && failure != nil && failure.ExitIgnorable
}

// Shutdown silently ignores unavailable workers, but reports other I/O
// failures, including a remote server failure ignored by normal completion.
func isDistributedWorkerUnavailable(err error) bool {
	failure, ok := err.(*DistributedOperationError)
	return ok && failure != nil && failure.WorkerUnavailable
}

func isDistributedRemoteFailure(err error) bool {
	failure, ok := err.(*DistributedOperationError)
	return ok && failure != nil && failure.Remote
}
func isDistributedNullFailure(err error) bool {
	if failure, ok := err.(*DistributedOperationError); ok {
		return failure.Null
	}
	return isJavaNullPointerException(err)
}

// Fatal endpoint failures cross the remote invocation's I/O boundary. Apply
// this to returned errors and panics alike; preserve the original cause graph.
func encodeDistributedRPCFailure(failure error) (*DistributedFailurePayload, error) {
	if isJavaError(failure) {
		failure = &DistributedOperationError{Message: javaString(failure.Error()), Class: javaThrowableClassName(failure), Stack: javaThrowableStackTrace(failure), Cause: failure, Remote: true, IO: true}
	}
	return EncodeDistributedFailure(failure)
}

func EncodeDistributedFailure(failure error) (*DistributedFailurePayload, error) {
	payload := &DistributedFailurePayload{}
	ids := make(map[error]int)
	var states []*TLCStateMut
	var visit func(error) int
	visit = func(err error) int {
		if err == nil {
			return 0
		}
		comparable := reflect.TypeOf(err).Comparable()
		if comparable && ids[err] != 0 {
			return ids[err]
		}
		id := len(payload.Nodes) + 1
		if comparable {
			ids[err] = id
		}
		payload.Nodes = append(payload.Nodes, DistributedFailureNode{})
		node := DistributedFailureNode{
			Class: javaThrowableClassName(err), Stack: javaThrowableStackTrace(err),
			Remote: isDistributedRemoteFailure(err), Recoverable: isRecoverableDistributedError(err),
			IO: isJavaIOException(err), Null: isDistributedNullFailure(err),
			ExitIgnorable:       isIgnorableDistributedWorkerExit(err),
			WorkerUnavailable:   isDistributedWorkerUnavailable(err),
			FingerprintRejected: isDistributedFPRegistrationRejected(err),
		}
		message := javaThrowableDetailMessage(err)
		if operation, ok := err.(*DistributedOperationError); ok {
			node.DiscoveryRetry, node.Reachable = operation.DiscoveryRetry, operation.Reachable
			node.EndpointRemoved = operation.EndpointRemoved
			node.BindingMissing = operation.BindingMissing
		}
		node.MessageNil = message == nil
		if message != nil {
			node.Message = *message
		}
		if worker, ok := err.(*WorkerException); ok {
			node.Worker, node.KeepCallStack = true, worker.KeepCallStack
			node.State1 = len(states) + 1
			node.State2 = len(states) + 2
			states = append(states, worker.State1, worker.State2)
		}
		node.Cause = visit(javaThrowableCause(err))
		if suppressed, ok := err.(interface{ GetSuppressed() []error }); ok {
			for _, cause := range suppressed.GetSuppressed() {
				node.Suppressed = append(node.Suppressed, visit(cause))
			}
		}
		payload.Nodes[id-1] = node
		return id
	}
	payload.Root = visit(failure)
	var err error
	payload.States, err = EncodeDistributedStates(states)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func DecodeDistributedFailure(payload *DistributedFailurePayload) (error, error) {
	if payload == nil {
		return nil, fmt.Errorf("missing distributed failure payload")
	}
	if payload.Root < 0 || payload.Root > len(payload.Nodes) {
		return nil, fmt.Errorf("invalid failure root %d", payload.Root)
	}
	if payload.Root == 0 && len(payload.Nodes) != 0 {
		return nil, fmt.Errorf("null failure contains nodes")
	}
	states, err := DecodeDistributedStates(payload.States)
	if err != nil {
		return nil, err
	}
	failures := make([]error, len(payload.Nodes)+1)
	for i, node := range payload.Nodes {
		var message *string
		if !node.MessageNil {
			message = javaString(node.Message)
		}
		if node.Worker {
			worker := NewWorkerException(node.Message)
			worker.messageNull = node.MessageNil
			worker.KeepCallStack = node.KeepCallStack
			worker.throwableTrace.remoteStack = node.Stack
			failures[i+1] = worker
		} else {
			failures[i+1] = &DistributedOperationError{Message: message, Class: node.Class, Stack: node.Stack, Remote: node.Remote, Recoverable: node.Recoverable, IO: node.IO, Null: node.Null, ExitIgnorable: node.ExitIgnorable, WorkerUnavailable: node.WorkerUnavailable, EndpointRemoved: node.EndpointRemoved, BindingMissing: node.BindingMissing, DiscoveryRetry: node.DiscoveryRetry, Reachable: node.Reachable, FingerprintRejected: node.FingerprintRejected}
		}
	}
	for i, node := range payload.Nodes {
		if node.Cause < 0 || node.Cause >= len(failures) {
			return nil, fmt.Errorf("invalid failure cause %d", node.Cause)
		}
		var suppressed []error
		for _, id := range node.Suppressed {
			if id <= 0 || id >= len(failures) {
				return nil, fmt.Errorf("invalid suppressed failure %d", id)
			}
			suppressed = append(suppressed, failures[id])
		}
		if worker, ok := failures[i+1].(*WorkerException); ok {
			if node.State1 <= 0 || node.State1 > len(states) || node.State2 <= 0 || node.State2 > len(states) {
				return nil, fmt.Errorf("invalid worker failure state references")
			}
			worker.State1, worker.State2 = states[node.State1-1], states[node.State2-1]
			worker.Cause = failures[node.Cause]
			for _, cause := range suppressed {
				worker.addSuppressedError(cause)
			}
		} else {
			if node.State1 != 0 || node.State2 != 0 {
				return nil, fmt.Errorf("non-worker failure contains state references")
			}
			failure := failures[i+1].(*DistributedOperationError)
			failure.Cause, failure.Suppressed = failures[node.Cause], suppressed
		}
	}
	return failures[payload.Root], nil
}
