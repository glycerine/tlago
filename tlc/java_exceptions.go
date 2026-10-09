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

// JavaExceptionBase carries Exception's nullable message, cause and captured
// port frames for parser exception classes defined outside package tlc.
type JavaExceptionBase = javaExceptionBase

func NewJavaExceptionBase(message *string, cause error) JavaExceptionBase {
	return newJavaExceptionBase(message, cause)
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

// FrontEndException ports SANY's checked exception and chained constructors.
type FrontEndException struct{ javaExceptionBase }

func NewFrontEndException(message string) *FrontEndException {
	return &FrontEndException{newJavaExceptionBase(javaString(message), nil)}
}

func NewFrontEndExceptionFromCause(cause error) *FrontEndException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return &FrontEndException{newJavaExceptionBase(message, cause)}
}

func (e *FrontEndException) Error() string { return javaThrowableMessage(e) }

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
	if operation, ok := err.(*DistributedOperationError); ok {
		return operation.IO
	}
	switch err.(type) {
	case *IOException, *FPSetManagerException, *FileSystemException, *NoSuchFileException, *AccessDeniedException, *FileAlreadyExistsException, *CharacterCodingException, *MalformedInputException, *UnmappableCharacterException, *UnsupportedEncodingException, *FileNotFoundException, *UnknownHostException, *NetConnectException, *NetBindException, *NoRouteToHostException, *MalformedURLException, *EOFException, *os.PathError, *os.LinkError, *os.SyscallError:
		return true
	}
	return err == io.EOF || err == io.ErrUnexpectedEOF || err == io.ErrClosedPipe || err == io.ErrShortWrite
}

type WrongInvocationException struct{ javaExceptionBase }

type NoSuchElementException struct{ javaExceptionBase }

func NewNoSuchElementException() *NoSuchElementException {
	return &NoSuchElementException{javaExceptionBase: newJavaExceptionBase(nil, nil)}
}
func (e *NoSuchElementException) Error() string { return javaThrowableMessage(e) }

// AssertionError is the Error thrown by Java language assertions.
type AssertionError struct{ javaExceptionBase }

func NewAssertionError(message ...string) *AssertionError {
	return &AssertionError{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *AssertionError) Error() string { return javaThrowableMessage(e) }

type JavaError struct{ javaExceptionBase }

func NewJavaError(message ...string) *JavaError {
	return &JavaError{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *JavaError) Error() string { return javaThrowableMessage(e) }

// Java catch(Exception) excludes Error even when it has no TLC system code.
func isJavaError(err error) bool {
	_, generic := err.(*JavaError)
	switch err.(type) {
	case *AssertionError, *ExceptionInInitializerError, *NoClassDefFoundError:
		return true
	}
	return generic || javaSystemFailureCode(err) != NoError
}

// IsJavaError distinguishes Error from the checked/runtime Exception families
// for parser integration outside package tlc.
func IsJavaError(err error) bool { return isJavaError(err) }

// Java class initialization and reflection inspect these exact
// checked/Error families; a throwable's cause does not change its catch type.
type ExceptionInInitializerError struct{ javaExceptionBase }

func NewExceptionInInitializerError(cause error) *ExceptionInInitializerError {
	return &ExceptionInInitializerError{newJavaExceptionBase(nil, cause)}
}
func NewExceptionInInitializerErrorMessage(message *string) *ExceptionInInitializerError {
	return &ExceptionInInitializerError{newJavaExceptionBase(message, nil)}
}
func (e *ExceptionInInitializerError) GetException() error { return e.Cause }
func (e *ExceptionInInitializerError) Error() string       { return javaThrowableMessage(e) }

type NoClassDefFoundError struct{ javaExceptionBase }

func NewNoClassDefFoundError(message ...string) *NoClassDefFoundError {
	return &NoClassDefFoundError{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NoClassDefFoundError) Error() string { return javaThrowableMessage(e) }

type ClassNotFoundException struct{ javaExceptionBase }

func NewClassNotFoundException(message ...string) *ClassNotFoundException {
	return &ClassNotFoundException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *ClassNotFoundException) Error() string { return javaThrowableMessage(e) }

type IllegalAccessException struct{ javaExceptionBase }

func NewIllegalAccessException(message ...string) *IllegalAccessException {
	return &IllegalAccessException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *IllegalAccessException) Error() string { return javaThrowableMessage(e) }

type InstantiationException struct{ javaExceptionBase }

type NoSuchMethodException struct{ javaExceptionBase }

func NewNoSuchMethodException(message ...string) *NoSuchMethodException {
	return &NoSuchMethodException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NoSuchMethodException) Error() string { return javaThrowableMessage(e) }

type InvocationTargetException struct{ javaExceptionBase }

func NewInvocationTargetException(target error, message ...string) *InvocationTargetException {
	return &InvocationTargetException{newJavaExceptionBase(optionalJavaMessage(message), target)}
}
func (e *InvocationTargetException) Error() string             { return javaThrowableMessage(e) }
func (e *InvocationTargetException) GetTargetException() error { return e.Cause }

func NewInstantiationException(message ...string) *InstantiationException {
	return &InstantiationException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *InstantiationException) Error() string { return javaThrowableMessage(e) }

type SecurityException struct{ javaExceptionBase }

func NewSecurityException(message ...string) *SecurityException {
	return &SecurityException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *SecurityException) Error() string { return javaThrowableMessage(e) }

type ConcurrentModificationException struct{ javaExceptionBase }

func NewConcurrentModificationException() *ConcurrentModificationException {
	return &ConcurrentModificationException{javaExceptionBase: newJavaExceptionBase(nil, nil)}
}
func (e *ConcurrentModificationException) Error() string { return javaThrowableMessage(e) }

func NewWrongInvocationException(message string) *WrongInvocationException {
	return &WrongInvocationException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)}
}

func (e *WrongInvocationException) Error() string { return javaThrowableMessage(e) }

type ArrayIndexOutOfBoundsException struct{ javaExceptionBase }

// Vector.checkBounds throws the no-argument Java constructor, with a null message.
func NewArrayIndexOutOfBoundsExceptionNoMessage() *ArrayIndexOutOfBoundsException {
	return &ArrayIndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(nil, nil)}
}

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

type IllegalStateException struct{ javaExceptionBase }

func NewIllegalStateException(message ...string) *IllegalStateException {
	return &IllegalStateException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *IllegalStateException) Error() string { return javaThrowableMessage(e) }

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

func NewUnsupportedOperationException(message ...string) *UnsupportedOperationException {
	return &UnsupportedOperationException{javaExceptionBase: newJavaExceptionBase(optionalJavaMessage(message), nil)}
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

// Fingerprint registration rejection is a TLC application failure. Retain its
// checked-I/O classification without implementing Java's RMI exception hierarchy.
type FPSetManagerException struct{ *IOException }

func NewFPSetManagerException(message string) *FPSetManagerException {
	return &FPSetManagerException{IOException: NewIOException(message)}
}

func (e *FPSetManagerException) Error() string { return javaThrowableMessage(e) }

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

type NamingException struct{ javaExceptionBase }

func NewNamingException(message ...string) *NamingException {
	return &NamingException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NamingException) Error() string { return javaThrowableMessage(e) }

type UnsupportedEncodingException struct{ *IOException }

func NewUnsupportedEncodingException(message string) *UnsupportedEncodingException {
	return &UnsupportedEncodingException{NewIOException(message)}
}
func (e *UnsupportedEncodingException) Error() string { return javaThrowableMessage(e) }

type NumberFormatException struct{ *IllegalArgumentException }

func NewNumberFormatException(value string) *NumberFormatException {
	return &NumberFormatException{NewIllegalArgumentException("For input string: \"" + value + "\"")}
}
func (e *NumberFormatException) Error() string { return javaThrowableMessage(e) }
