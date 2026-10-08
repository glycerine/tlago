package tlc

import (
	"fmt"
	"reflect"
)

// DistributedOperationError is an operation failure received from another Go
// process. Its traits drive TLC's catch/retry decisions without a Java remote
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
	DiscoveryRetry      bool
	Reachable           bool
	FingerprintRejected bool
}

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
	if failure, ok := err.(*DistributedOperationError); ok {
		return failure.ExitIgnorable
	}
	switch err.(type) {
	case *NoSuchObjectException, *ConnectException, *ServerException:
		return true
	}
	return false
}

func isDistributedRemoteFailure(err error) bool {
	if failure, ok := err.(*DistributedOperationError); ok {
		return failure.Remote
	}
	if failure, ok := err.(*DistributedEndpointError); ok {
		return failure.IO
	}
	return javaRemoteException(err) != nil
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
			FingerprintRejected: isDistributedFPRegistrationRejected(err),
		}
		message := javaThrowableDetailMessage(err)
		if operation, ok := err.(*DistributedOperationError); ok {
			node.DiscoveryRetry, node.Reachable = operation.DiscoveryRetry, operation.Reachable
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
			failures[i+1] = &DistributedOperationError{Message: message, Class: node.Class, Stack: node.Stack, Remote: node.Remote, Recoverable: node.Recoverable, IO: node.IO, Null: node.Null, ExitIgnorable: node.ExitIgnorable, DiscoveryRetry: node.DiscoveryRetry, Reachable: node.Reachable, FingerprintRejected: node.FingerprintRejected}
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
