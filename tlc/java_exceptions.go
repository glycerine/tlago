package tlc

import (
	"fmt"
	"io"
	"os"
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

type IOException struct{ javaExceptionBase }

type FileNotFoundException struct{ *IOException }

type UnknownHostException struct{ *IOException }

type InvalidPathException struct{ *IllegalArgumentException }

func NewInvalidPathException(path, reason string) *InvalidPathException {
	return &InvalidPathException{IllegalArgumentException: NewIllegalArgumentException(reason + ": " + path)}
}

func (e *InvalidPathException) Error() string { return javaThrowableMessage(e) }

func NewUnknownHostException(message string) *UnknownHostException {
	return &UnknownHostException{IOException: NewIOException(message)}
}

func (e *UnknownHostException) Error() string { return javaThrowableMessage(e) }

func NewFileNotFoundException(message string) *FileNotFoundException {
	return &FileNotFoundException{IOException: NewIOException(message)}
}

func (e *FileNotFoundException) Error() string { return javaThrowableMessage(e) }

type RuntimeException struct{ javaExceptionBase }

func NewRuntimeException(message ...string) *RuntimeException {
	return &RuntimeException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func NewRuntimeExceptionWithCause(message *string, cause error) *RuntimeException {
	return &RuntimeException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}

func NewRuntimeExceptionFromCause(cause error) *RuntimeException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewRuntimeExceptionWithCause(message, cause)
}

func (e *RuntimeException) Error() string { return javaThrowableMessage(e) }

func NewIOException(message ...string) *IOException {
	return &IOException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func (e *IOException) Error() string { return javaThrowableMessage(e) }

func isJavaIOException(err error) bool {
	if javaRemoteException(err) != nil {
		return true
	}
	switch err.(type) {
	case *IOException, *UnsupportedEncodingException, *FileNotFoundException, *UnknownHostException, *NetConnectException, *NetBindException, *NoRouteToHostException, *MalformedURLException, *EOFException, *os.PathError, *os.LinkError, *os.SyscallError:
		return true
	}
	return err == io.EOF || err == io.ErrUnexpectedEOF || err == io.ErrClosedPipe || err == io.ErrShortWrite
}

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

type IndexOutOfBoundsException struct{ javaExceptionBase }

func NewIndexOutOfBoundsException(index, length int) *IndexOutOfBoundsException {
	return &IndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString(fmt.Sprintf("Index %d out of bounds for length %d", index, length)), nil)}
}

func (e *IndexOutOfBoundsException) Error() string { return javaThrowableMessage(e) }

type StringIndexOutOfBoundsException struct{ *IndexOutOfBoundsException }

func NewStringIndexOutOfBoundsException(index, length int) *StringIndexOutOfBoundsException {
	return &StringIndexOutOfBoundsException{IndexOutOfBoundsException: NewIndexOutOfBoundsException(index, length)}
}

func (e *StringIndexOutOfBoundsException) Error() string { return javaThrowableMessage(e) }

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

type IllegalArgumentException struct{ javaExceptionBase }

func NewIllegalArgumentException(message ...string) *IllegalArgumentException {
	return &IllegalArgumentException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func NewIllegalArgumentExceptionWithCause(message *string, cause error) *IllegalArgumentException {
	return &IllegalArgumentException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}

func (e *IllegalArgumentException) Error() string { return javaThrowableMessage(e) }

type URISyntaxException struct {
	javaExceptionBase
	input  string
	reason string
	index  int
}

func NewURISyntaxException(input, reason string, index int) *URISyntaxException {
	if index < -1 {
		panic(NewIllegalArgumentException())
	}
	message := reason
	if index >= 0 {
		message += fmt.Sprintf(" at index %d", index)
	}
	message += ": " + input
	return &URISyntaxException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil), input: input, reason: reason, index: index}
}

func (e *URISyntaxException) GetInput() string  { return e.input }
func (e *URISyntaxException) GetReason() string { return e.reason }
func (e *URISyntaxException) GetIndex() int     { return e.index }
func (e *URISyntaxException) Error() string     { return javaThrowableMessage(e) }

type ArithmeticException struct{ javaExceptionBase }

func NewArithmeticException(message string) *ArithmeticException {
	return &ArithmeticException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)}
}

func (e *ArithmeticException) Error() string { return javaThrowableMessage(e) }

type UnsupportedOperationException struct{ javaExceptionBase }

func NewUnsupportedOperationException(message string) *UnsupportedOperationException {
	return &UnsupportedOperationException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)}
}

func (e *UnsupportedOperationException) Error() string { return javaThrowableMessage(e) }

type ExecutionException struct{ javaExceptionBase }

func NewExecutionException(cause error) *ExecutionException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return &ExecutionException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}

func (e *ExecutionException) Error() string { return javaThrowableMessage(e) }

type StackOverflowError struct{ javaExceptionBase }

func NewStackOverflowError(message ...string) *StackOverflowError {
	return &StackOverflowError{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}

func (e *StackOverflowError) Error() string { return javaThrowableMessage(e) }

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

type UnmarshalException struct{ *RemoteException }

func NewUnmarshalException(message string, cause error) *UnmarshalException {
	return &UnmarshalException{RemoteException: NewRemoteException(javaString(message), cause)}
}

func (e *UnmarshalException) Error() string { return javaThrowableMessage(e) }

type FPSetManagerException struct{ *RemoteException }

func NewFPSetManagerException(message string) *FPSetManagerException {
	return &FPSetManagerException{RemoteException: NewRemoteException(javaString(message), nil)}
}

func (e *FPSetManagerException) Error() string { return javaThrowableMessage(e) }

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
	case *ConnectIOException:
		if failure != nil {
			return failure.RemoteException
		}
	case *RMIUnknownHostException:
		if failure != nil {
			return failure.RemoteException
		}
	case *FPSetManagerException:
		if failure != nil {
			return failure.RemoteException
		}
	case *UnmarshalException:
		if failure != nil {
			return failure.RemoteException
		}
	case *ExportException:
		if failure != nil {
			return failure.RemoteException
		}
	case *AccessException:
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
