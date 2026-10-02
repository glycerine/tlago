package tlc

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

// The concrete Java exception families below preserve the types inspected by
// TLC's catches and stream boundaries. They carry Java types without adding RPC.
type javaExceptionBase struct {
	throwableTrace
	Message *string
	Cause   error
}

func newJavaExceptionBase(message *string, cause error) javaExceptionBase {
	return javaExceptionBase{throwableTrace: captureThrowableTrace(), Message: copyJavaMessage(message), Cause: cause}
}

func copyJavaMessage(message *string) *string {
	if message == nil {
		return nil
	}
	return javaString(*message)
}

func (e *javaExceptionBase) GetMessage() *string { return copyJavaMessage(e.Message) }
func (e *javaExceptionBase) GetCause() error     { return e.Cause }
func (e *javaExceptionBase) Unwrap() error       { return e.Cause }

func optionalJavaMessage(messages []string) *string {
	if len(messages) == 0 {
		return nil
	}
	return javaString(messages[0])
}

type EOFException struct{ javaExceptionBase }

func NewEOFException(message ...string) *EOFException {
	return &EOFException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func (e *EOFException) Error() string { return javaThrowableMessage(e) }

// Go recovery loops use errors.Is(io.EOF) for Java's catch(EOFException).
func (e *EOFException) Is(target error) bool { return target == io.EOF }

type WrongInvocationException struct{ javaExceptionBase }

func NewWrongInvocationException(message string) *WrongInvocationException {
	return &WrongInvocationException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)}
}

func (e *WrongInvocationException) Error() string { return javaThrowableMessage(e) }

type ArrayIndexOutOfBoundsException struct{ javaExceptionBase }

func NewArrayIndexOutOfBoundsException(index, length int) *ArrayIndexOutOfBoundsException {
	message := fmt.Sprintf("Index %d out of bounds for length %d", index, length)
	return &ArrayIndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)}
}

func (e *ArrayIndexOutOfBoundsException) Error() string { return javaThrowableMessage(e) }

type NegativeArraySizeException struct{ javaExceptionBase }

func NewNegativeArraySizeException(message ...string) *NegativeArraySizeException {
	return &NegativeArraySizeException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func (e *NegativeArraySizeException) Error() string { return javaThrowableMessage(e) }

type ClassCastException struct{ javaExceptionBase }

func NewClassCastException(message ...string) *ClassCastException {
	return &ClassCastException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func (e *ClassCastException) Error() string { return javaThrowableMessage(e) }

type OutOfMemoryError struct{ javaExceptionBase }

func NewOutOfMemoryError(message ...string) *OutOfMemoryError {
	return &OutOfMemoryError{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func (e *OutOfMemoryError) Error() string { return javaThrowableMessage(e) }

type NullPointerException struct{ javaExceptionBase }

func NewNullPointerException(message ...string) *NullPointerException {
	return &NullPointerException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func (e *NullPointerException) Error() string { return javaThrowableMessage(e) }

// RejectedExecutionException has both message+cause and cause-only Java
// constructors; the latter sets its detail message to the cause's toString().
type RejectedExecutionException struct{ javaExceptionBase }

func NewRejectedExecutionException(message *string, cause error) *RejectedExecutionException {
	return &RejectedExecutionException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}

func NewRejectedExecutionExceptionWithCause(cause error) *RejectedExecutionException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewRejectedExecutionException(message, cause)
}

func (e *RejectedExecutionException) Error() string { return javaThrowableMessage(e) }

// RemoteException.getCause() uses the public detail field, and getMessage()
// appends detail.toString() even when the original message was null or empty.
type RemoteException struct {
	throwableTrace
	Message *string
	Detail  error
}

func NewRemoteException(message *string, detail error) *RemoteException {
	return &RemoteException{throwableTrace: captureThrowableTrace(), Message: copyJavaMessage(message), Detail: detail}
}

func (e *RemoteException) GetMessage() *string {
	if e == nil {
		return nil
	}
	if e.Detail == nil {
		return copyJavaMessage(e.Message)
	}
	message := "null"
	if e.Message != nil {
		message = *e.Message
	}
	return javaString(message + "; nested exception is: \n\t" + javaThrowableString(e.Detail))
}

func (e *RemoteException) GetCause() error {
	if e == nil {
		return nil
	}
	return e.Detail
}

func (e *RemoteException) Unwrap() error { return e.GetCause() }
func (e *RemoteException) Error() string { return javaThrowableMessage(e) }

type ServerException struct{ *RemoteException }

func NewServerException(message *string, cause error) *ServerException {
	return &ServerException{RemoteException: NewRemoteException(message, cause)}
}

func (e *ServerException) Error() string { return javaThrowableMessage(e) }

type NoSuchObjectException struct{ *RemoteException }

func NewNoSuchObjectException(message string) *NoSuchObjectException {
	return &NoSuchObjectException{RemoteException: NewRemoteException(javaString(message), nil)}
}

func (e *NoSuchObjectException) Error() string { return javaThrowableMessage(e) }

type ConnectException struct{ *RemoteException }

func NewConnectException(message string, cause error) *ConnectException {
	return &ConnectException{RemoteException: NewRemoteException(javaString(message), cause)}
}

func (e *ConnectException) Error() string { return javaThrowableMessage(e) }

func javaRemoteException(err error) *RemoteException {
	switch failure := err.(type) {
	case *RemoteException:
		return failure
	case *ServerException:
		if failure != nil {
			return failure.RemoteException
		}
	case *NoSuchObjectException:
		if failure != nil {
			return failure.RemoteException
		}
	case *ConnectException:
		if failure != nil {
			return failure.RemoteException
		}
	}
	return nil
}

func isJavaNullPointerException(err error) bool {
	if failure, ok := err.(*NullPointerException); ok {
		return failure != nil
	}
	// A Go nil dereference is the language's equivalent runtime failure.
	failure, ok := err.(runtime.Error)
	return ok && strings.Contains(failure.Error(), "invalid memory address or nil pointer dereference")
}

func isJavaOutOfMemoryError(err error) bool {
	code := javaSystemFailureCode(err)
	return code == ECSystemOutOfMemory || code == ECSystemOutOfMemoryLiveness
}
