package tlc

import (
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"runtime"
	"strconv"
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
	if len(messages) > 1 {
		panic(NewIllegalArgumentException("Java exception constructors accept at most one message"))
	}
	if len(messages) == 0 {
		return nil
	}
	return javaString(messages[0])
}

type EOFException struct{ javaExceptionBase }

type MalformedJsonException struct{ *IOException }

func NewMalformedJsonException(message string) *MalformedJsonException {
	return &MalformedJsonException{IOException: NewIOException(message)}
}

func (e *MalformedJsonException) Error() string { return javaThrowableMessage(e) }
func (e *MalformedJsonException) GetMessage() *string {
	if e == nil {
		panic(NewNullPointerException())
	}
	return e.IOException.GetMessage()
}
func (e *MalformedJsonException) GetCause() error {
	if e == nil {
		panic(NewNullPointerException())
	}
	return e.IOException.GetCause()
}

func NewEOFException(message ...string) *EOFException {
	return NewEOFExceptionNullable(optionalJavaMessage(message))
}

func NewEOFExceptionNullable(message *string) *EOFException {
	return &EOFException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func (e *EOFException) Error() string { return javaThrowableMessage(e) }

func (e *EOFException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}

func (e *EOFException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}

func (e *EOFException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

// Go recovery loops use errors.Is(io.EOF) for Java's catch(EOFException).
func (e *EOFException) Is(target error) bool { return target == io.EOF }

type IOException struct{ javaExceptionBase }

type FileNotFoundException struct{ *IOException }

type UnknownHostException struct{ *IOException }

type ZipException struct{ *IOException }

func NewZipException(message ...string) *ZipException {
	return NewZipExceptionNullable(optionalJavaMessage(message))
}

func NewZipExceptionNullable(message *string) *ZipException {
	return &ZipException{IOException: NewIOExceptionNullable(message)}
}

func (e *ZipException) Error() string { return javaThrowableMessage(e) }
func (e *ZipException) GetMessage() *string {
	e.require()
	return e.IOException.GetMessage()
}
func (e *ZipException) GetCause() error {
	e.require()
	return e.IOException.GetCause()
}
func (e *ZipException) require() {
	if e == nil || e.IOException == nil {
		panic(NewNullPointerException())
	}
}

type InvalidPathException struct {
	*IllegalArgumentException
	input  string
	reason string
	index  int
}

func NewInvalidPathException(input, reason string) *InvalidPathException {
	return NewInvalidPathExceptionAt(input, reason, -1)
}

func NewInvalidPathExceptionAt(input, reason string, index int) *InvalidPathException {
	return newInvalidPathException(javaString(input), javaString(reason), index)
}

func NewInvalidPathExceptionNullable(input, reason *string, index int) *InvalidPathException {
	return newInvalidPathException(input, reason, index)
}

func newInvalidPathException(input, reason *string, index int) *InvalidPathException {
	index = int(int32(index))
	if input == nil || reason == nil {
		panic(NewNullPointerException())
	}
	if index < -1 {
		panic(NewIllegalArgumentException())
	}
	return &InvalidPathException{
		IllegalArgumentException: NewIllegalArgumentException(*reason),
		input:                    *input,
		reason:                   *reason,
		index:                    index,
	}
}

func (e *InvalidPathException) Error() string { return javaThrowableMessage(e) }
func (e *InvalidPathException) GetInput() string {
	e.require()
	return e.input
}
func (e *InvalidPathException) GetReason() string {
	e.require()
	return e.reason
}
func (e *InvalidPathException) GetIndex() int {
	e.require()
	return e.index
}
func (e *InvalidPathException) GetMessage() *string {
	e.require()
	message := e.reason
	if e.index > -1 {
		message += fmt.Sprintf(" at index %d", e.index)
	}
	message += ": " + e.input
	return javaString(message)
}
func (e *InvalidPathException) GetCause() error {
	e.require()
	return e.IllegalArgumentException.GetCause()
}
func (e *InvalidPathException) require() {
	if e == nil || e.IllegalArgumentException == nil {
		panic(NewNullPointerException())
	}
}

func NewUnknownHostException(message ...string) *UnknownHostException {
	return NewUnknownHostExceptionNullable(optionalJavaMessage(message))
}

func NewUnknownHostExceptionNullable(message *string) *UnknownHostException {
	return &UnknownHostException{IOException: NewIOExceptionNullable(message)}
}

func (e *UnknownHostException) Error() string { return javaThrowableMessage(e) }
func (e *UnknownHostException) GetMessage() *string {
	e.require()
	return e.IOException.GetMessage()
}
func (e *UnknownHostException) GetCause() error {
	e.require()
	return e.IOException.GetCause()
}
func (e *UnknownHostException) require() {
	if e == nil || e.IOException == nil {
		panic(NewNullPointerException())
	}
}

func NewFileNotFoundException(message ...string) *FileNotFoundException {
	return NewFileNotFoundExceptionNullable(optionalJavaMessage(message))
}

func NewFileNotFoundExceptionNullable(message *string) *FileNotFoundException {
	return &FileNotFoundException{IOException: NewIOExceptionNullable(message)}
}

func (e *FileNotFoundException) Error() string { return javaThrowableMessage(e) }
func (e *FileNotFoundException) GetMessage() *string {
	e.require()
	return e.IOException.GetMessage()
}
func (e *FileNotFoundException) GetCause() error {
	e.require()
	return e.IOException.GetCause()
}
func (e *FileNotFoundException) require() {
	if e == nil || e.IOException == nil {
		panic(NewNullPointerException())
	}
}

// FrontEndException ports SANY's checked exception and chained constructors.
type FrontEndException struct{ javaExceptionBase }

func NewFrontEndException(message string) *FrontEndException {
	return NewFrontEndExceptionNullable(javaString(message))
}

func NewFrontEndExceptionNullable(message *string) *FrontEndException {
	return &FrontEndException{newJavaExceptionBase(message, nil)}
}

func NewFrontEndExceptionFromCause(cause error) *FrontEndException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return &FrontEndException{newJavaExceptionBase(message, cause)}
}

func (e *FrontEndException) Error() string { return javaThrowableMessage(e) }

func (e *FrontEndException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}

func (e *FrontEndException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}

func (e *FrontEndException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type RuntimeException struct{ javaExceptionBase }

type UncheckedIOException struct{ *RuntimeException }

type JsonParseException struct{ *RuntimeException }
type JsonSyntaxException struct{ *JsonParseException }

func NewUncheckedIOExceptionFromCause(cause error) *UncheckedIOException {
	return &UncheckedIOException{RuntimeException: NewRuntimeExceptionFromCause(cause)}
}

func (e *UncheckedIOException) Error() string { return javaThrowableMessage(e) }
func (e *UncheckedIOException) GetMessage() *string {
	if e == nil || e.RuntimeException == nil {
		panic(NewNullPointerException())
	}
	return e.RuntimeException.GetMessage()
}
func (e *UncheckedIOException) GetCause() error {
	if e == nil || e.RuntimeException == nil {
		panic(NewNullPointerException())
	}
	return e.RuntimeException.GetCause()
}

func NewJsonSyntaxExceptionFromCause(cause error) *JsonSyntaxException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return &JsonSyntaxException{JsonParseException: &JsonParseException{
		RuntimeException: NewRuntimeExceptionWithCause(message, cause),
	}}
}

func (e *JsonSyntaxException) Error() string { return javaThrowableMessage(e) }
func (e *JsonSyntaxException) GetMessage() *string {
	if e == nil {
		panic(NewNullPointerException())
	}
	return e.JsonParseException.GetMessage()
}
func (e *JsonSyntaxException) GetCause() error {
	if e == nil {
		panic(NewNullPointerException())
	}
	return e.JsonParseException.GetCause()
}

func NewRuntimeException(message ...string) *RuntimeException {
	return NewRuntimeExceptionNullable(optionalJavaMessage(message))
}

func NewRuntimeExceptionNullable(message *string) *RuntimeException {
	return NewRuntimeExceptionWithCause(message, nil)
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

func (e *RuntimeException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}

func (e *RuntimeException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}

func (e *RuntimeException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func NewIOException(message ...string) *IOException {
	return NewIOExceptionNullable(optionalJavaMessage(message))
}

func NewIOExceptionNullable(message *string) *IOException {
	return NewIOExceptionWithCause(message, nil)
}

func NewIOExceptionWithCause(message *string, cause error) *IOException {
	return &IOException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}

func NewIOExceptionFromCause(cause error) *IOException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewIOExceptionWithCause(message, cause)
}

func (e *IOException) Error() string { return javaThrowableMessage(e) }

func (e *IOException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}

func (e *IOException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}

func (e *IOException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func isJavaIOException(err error) bool {
	if operation, ok := err.(*DistributedOperationError); ok {
		return operation.IO
	}
	switch err.(type) {
	case *ZipException:
		return true
	case *IOException, *MalformedJsonException, *FPSetManagerException, *FileSystemException, *NoSuchFileException, *AccessDeniedException, *FileAlreadyExistsException, *DirectoryNotEmptyException, *CharacterCodingException, *MalformedInputException, *UnmappableCharacterException, *UnsupportedEncodingException, *FileNotFoundException, *UnknownHostException, *NetConnectException, *NetBindException, *NoRouteToHostException, *MalformedURLException, *EOFException, *os.PathError, *os.LinkError, *os.SyscallError:
		return true
	}
	return err == io.EOF || err == io.ErrUnexpectedEOF || err == io.ErrClosedPipe || err == io.ErrShortWrite
}

type WrongInvocationException struct{ javaExceptionBase }

type NoSuchElementException struct{ javaExceptionBase }

func NewNoSuchElementException(message ...string) *NoSuchElementException {
	return NewNoSuchElementExceptionWithCause(optionalJavaMessage(message), nil)
}
func NewNoSuchElementExceptionWithCause(message *string, cause error) *NoSuchElementException {
	return &NoSuchElementException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}
func NewNoSuchElementExceptionFromCause(cause error) *NoSuchElementException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewNoSuchElementExceptionWithCause(message, cause)
}
func (e *NoSuchElementException) Error() string { return javaThrowableMessage(e) }
func (e *NoSuchElementException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NoSuchElementException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NoSuchElementException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

// AssertionError is the Error thrown by Java language assertions.
type AssertionError struct{ javaExceptionBase }

func NewAssertionError(message ...string) *AssertionError {
	return NewAssertionErrorWithCause(optionalJavaMessage(message), nil)
}

func NewAssertionErrorWithCause(message *string, cause error) *AssertionError {
	return &AssertionError{newJavaExceptionBase(message, cause)}
}

func NewAssertionErrorObject(value any) *AssertionError {
	message := javaAssertionValueString(value)
	var cause error
	if failure, ok := value.(error); ok && !javaNilObject(failure) {
		cause = failure
	}
	return NewAssertionErrorWithCause(javaString(message), cause)
}

func NewAssertionErrorBoolean(value bool) *AssertionError {
	return NewAssertionErrorWithCause(javaString(strconv.FormatBool(value)), nil)
}

func NewAssertionErrorChar(value uint16) *AssertionError {
	return NewAssertionErrorWithCause(javaString(javaStringFromUTF16([]uint16{value})), nil)
}

func NewAssertionErrorInt(value int32) *AssertionError {
	return NewAssertionErrorWithCause(javaString(strconv.FormatInt(int64(value), 10)), nil)
}

func NewAssertionErrorLong(value int64) *AssertionError {
	return NewAssertionErrorWithCause(javaString(strconv.FormatInt(value, 10)), nil)
}

func NewAssertionErrorFloat(value float32) *AssertionError {
	return NewAssertionErrorWithCause(javaString(javaAssertionFloatingString(float64(value), 32)), nil)
}

func NewAssertionErrorDouble(value float64) *AssertionError {
	return NewAssertionErrorWithCause(javaString(javaAssertionFloatingString(value, 64)), nil)
}

func (e *AssertionError) Error() string { return javaThrowableMessage(e) }
func (e *AssertionError) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *AssertionError) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *AssertionError) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func javaNilObject(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func javaAssertionValueString(value any) string {
	if javaNilObject(value) {
		return "null"
	}
	switch value := value.(type) {
	case string:
		return value
	case bool:
		return strconv.FormatBool(value)
	case uint16:
		return javaStringFromUTF16([]uint16{value})
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int:
		return strconv.Itoa(value)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case float32:
		return javaAssertionFloatingString(float64(value), 32)
	case float64:
		return javaAssertionFloatingString(value, 64)
	case error:
		return javaThrowableString(value)
	case fmt.Stringer:
		return value.String()
	default:
		return fmt.Sprint(value)
	}
}

func javaAssertionFloatingString(value float64, bitSize int) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	case value == 0:
		if math.Signbit(value) {
			return "-0.0"
		}
		return "0.0"
	}

	raw := strconv.FormatFloat(value, 'e', -1, bitSize)
	sign := ""
	if raw[0] == '-' {
		sign, raw = "-", raw[1:]
	}
	parts := strings.Split(raw, "e")
	exponent, _ := strconv.Atoi(parts[1])
	digits := strings.ReplaceAll(parts[0], ".", "")
	if exponent >= -3 && exponent < 7 {
		point := exponent + 1
		switch {
		case point <= 0:
			return sign + "0." + strings.Repeat("0", -point) + digits
		case point >= len(digits):
			return sign + digits + strings.Repeat("0", point-len(digits)) + ".0"
		default:
			return sign + digits[:point] + "." + digits[point:]
		}
	}
	if len(digits) == 1 {
		rounded := strconv.FormatFloat(value, 'e', 1, bitSize)
		if rounded[0] == '-' {
			rounded = rounded[1:]
		}
		roundedParts := strings.Split(rounded, "e")
		digits = strings.ReplaceAll(roundedParts[0], ".", "")
		exponent, _ = strconv.Atoi(roundedParts[1])
	}
	mantissa := digits[:1] + ".0"
	if len(digits) > 1 {
		mantissa = digits[:1] + "." + digits[1:]
	}
	return sign + mantissa + "E" + strconv.Itoa(exponent)
}

type JavaError struct{ javaExceptionBase }

func NewJavaError(message ...string) *JavaError {
	return NewJavaErrorWithCause(optionalJavaMessage(message), nil)
}
func NewJavaErrorWithCause(message *string, cause error) *JavaError {
	return &JavaError{newJavaExceptionBase(message, cause)}
}
func NewJavaErrorFromCause(cause error) *JavaError {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewJavaErrorWithCause(message, cause)
}
func (e *JavaError) Error() string { return javaThrowableMessage(e) }
func (e *JavaError) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *JavaError) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *JavaError) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

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

func NewExceptionInInitializerErrorDefault() *ExceptionInInitializerError {
	return &ExceptionInInitializerError{newJavaExceptionBase(nil, nil)}
}

func NewExceptionInInitializerError(cause error) *ExceptionInInitializerError {
	return &ExceptionInInitializerError{newJavaExceptionBase(nil, cause)}
}
func NewExceptionInInitializerErrorMessage(message *string) *ExceptionInInitializerError {
	return &ExceptionInInitializerError{newJavaExceptionBase(message, nil)}
}
func (e *ExceptionInInitializerError) GetException() error {
	e.require()
	return e.Cause
}

func (e *ExceptionInInitializerError) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}

func (e *ExceptionInInitializerError) GetCause() error {
	e.require()
	return e.Cause
}

func (e *ExceptionInInitializerError) Error() string { return javaThrowableMessage(e) }

func (e *ExceptionInInitializerError) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type NoClassDefFoundError struct{ javaExceptionBase }

func NewNoClassDefFoundError(message ...string) *NoClassDefFoundError {
	return NewNoClassDefFoundErrorNullable(optionalJavaMessage(message))
}
func NewNoClassDefFoundErrorNullable(message *string) *NoClassDefFoundError {
	return &NoClassDefFoundError{newJavaExceptionBase(message, nil)}
}
func (e *NoClassDefFoundError) Error() string { return javaThrowableMessage(e) }
func (e *NoClassDefFoundError) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NoClassDefFoundError) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NoClassDefFoundError) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type ClassNotFoundException struct{ javaExceptionBase }

func NewClassNotFoundException(message ...string) *ClassNotFoundException {
	return NewClassNotFoundExceptionWithCause(optionalJavaMessage(message), nil)
}
func NewClassNotFoundExceptionWithCause(message *string, cause error) *ClassNotFoundException {
	return &ClassNotFoundException{newJavaExceptionBase(message, cause)}
}
func (e *ClassNotFoundException) Error() string { return javaThrowableMessage(e) }
func (e *ClassNotFoundException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *ClassNotFoundException) GetCause() error {
	e.require()
	return e.Cause
}
func (e *ClassNotFoundException) GetException() error {
	e.require()
	return e.Cause
}
func (e *ClassNotFoundException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type IllegalAccessException struct{ javaExceptionBase }

func NewIllegalAccessException(message ...string) *IllegalAccessException {
	return NewIllegalAccessExceptionNullable(optionalJavaMessage(message))
}
func NewIllegalAccessExceptionNullable(message *string) *IllegalAccessException {
	return &IllegalAccessException{newJavaExceptionBase(message, nil)}
}
func (e *IllegalAccessException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalAccessException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *IllegalAccessException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *IllegalAccessException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type InstantiationException struct{ javaExceptionBase }

type NoSuchMethodException struct{ javaExceptionBase }

func NewNoSuchMethodException(message ...string) *NoSuchMethodException {
	return NewNoSuchMethodExceptionNullable(optionalJavaMessage(message))
}
func NewNoSuchMethodExceptionNullable(message *string) *NoSuchMethodException {
	return &NoSuchMethodException{newJavaExceptionBase(message, nil)}
}
func (e *NoSuchMethodException) Error() string { return javaThrowableMessage(e) }
func (e *NoSuchMethodException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NoSuchMethodException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NoSuchMethodException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type InvocationTargetException struct{ javaExceptionBase }

func NewInvocationTargetException(target error, message ...string) *InvocationTargetException {
	return NewInvocationTargetExceptionNullable(target, optionalJavaMessage(message))
}
func NewInvocationTargetExceptionDefault() *InvocationTargetException {
	return NewInvocationTargetExceptionNullable(nil, nil)
}
func NewInvocationTargetExceptionNullable(target error, message *string) *InvocationTargetException {
	return &InvocationTargetException{newJavaExceptionBase(message, target)}
}
func (e *InvocationTargetException) Error() string { return javaThrowableMessage(e) }
func (e *InvocationTargetException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *InvocationTargetException) GetCause() error {
	e.require()
	return e.Cause
}
func (e *InvocationTargetException) GetTargetException() error {
	e.require()
	return e.Cause
}
func (e *InvocationTargetException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func NewInstantiationException(message ...string) *InstantiationException {
	return NewInstantiationExceptionNullable(optionalJavaMessage(message))
}
func NewInstantiationExceptionNullable(message *string) *InstantiationException {
	return &InstantiationException{newJavaExceptionBase(message, nil)}
}
func (e *InstantiationException) Error() string { return javaThrowableMessage(e) }
func (e *InstantiationException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *InstantiationException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *InstantiationException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type SecurityException struct{ javaExceptionBase }

func NewSecurityException(message ...string) *SecurityException {
	return NewSecurityExceptionWithCause(optionalJavaMessage(message), nil)
}
func NewSecurityExceptionWithCause(message *string, cause error) *SecurityException {
	return &SecurityException{newJavaExceptionBase(message, cause)}
}
func NewSecurityExceptionFromCause(cause error) *SecurityException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewSecurityExceptionWithCause(message, cause)
}
func (e *SecurityException) Error() string { return javaThrowableMessage(e) }
func (e *SecurityException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *SecurityException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *SecurityException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type ConcurrentModificationException struct{ javaExceptionBase }

func NewConcurrentModificationException(message ...string) *ConcurrentModificationException {
	return NewConcurrentModificationExceptionWithCause(optionalJavaMessage(message), nil)
}
func NewConcurrentModificationExceptionWithCause(message *string, cause error) *ConcurrentModificationException {
	return &ConcurrentModificationException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}
func NewConcurrentModificationExceptionFromCause(cause error) *ConcurrentModificationException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewConcurrentModificationExceptionWithCause(message, cause)
}
func (e *ConcurrentModificationException) Error() string { return javaThrowableMessage(e) }
func (e *ConcurrentModificationException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *ConcurrentModificationException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *ConcurrentModificationException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func NewWrongInvocationException(message string) *WrongInvocationException {
	return NewWrongInvocationExceptionNullable(javaString(message))
}

func NewWrongInvocationExceptionNullable(message *string) *WrongInvocationException {
	return &WrongInvocationException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func (e *WrongInvocationException) Error() string { return javaThrowableMessage(e) }
func (e *WrongInvocationException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *WrongInvocationException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *WrongInvocationException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type ArrayIndexOutOfBoundsException struct{ javaExceptionBase }

// Vector.checkBounds throws the no-argument Java constructor, with a null message.
func NewArrayIndexOutOfBoundsExceptionNoMessage() *ArrayIndexOutOfBoundsException {
	return NewArrayIndexOutOfBoundsExceptionNullable(nil)
}

func NewArrayIndexOutOfBoundsExceptionNullable(message *string) *ArrayIndexOutOfBoundsException {
	return &ArrayIndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func NewArrayIndexOutOfBoundsExceptionForIndex(index int) *ArrayIndexOutOfBoundsException {
	index = int(int32(index))
	message := fmt.Sprintf("Array index out of range: %d", index)
	return NewArrayIndexOutOfBoundsExceptionNullable(javaString(message))
}

// NewArrayIndexOutOfBoundsException carries the JVM-generated array-access
// message, which differs from the public one-index Java constructor.
func NewArrayIndexOutOfBoundsException(index, length int) *ArrayIndexOutOfBoundsException {
	index = int(int32(index))
	length = int(int32(length))
	message := fmt.Sprintf("Index %d out of bounds for length %d", index, length)
	return NewArrayIndexOutOfBoundsExceptionNullable(javaString(message))
}

func (e *ArrayIndexOutOfBoundsException) Error() string { return javaThrowableMessage(e) }
func (e *ArrayIndexOutOfBoundsException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *ArrayIndexOutOfBoundsException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *ArrayIndexOutOfBoundsException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type IndexOutOfBoundsException struct{ javaExceptionBase }

func NewIndexOutOfBoundsExceptionDefault() *IndexOutOfBoundsException {
	return NewIndexOutOfBoundsExceptionNullable(nil)
}

func NewIndexOutOfBoundsExceptionNullable(message *string) *IndexOutOfBoundsException {
	return &IndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func NewIndexOutOfBoundsExceptionForInt(index int) *IndexOutOfBoundsException {
	index = int(int32(index))
	return NewIndexOutOfBoundsExceptionNullable(javaString(fmt.Sprintf("Index out of range: %d", index)))
}

func NewIndexOutOfBoundsExceptionForLong(index int64) *IndexOutOfBoundsException {
	return NewIndexOutOfBoundsExceptionNullable(javaString(fmt.Sprintf("Index out of range: %d", index)))
}

// NewIndexOutOfBoundsException carries the JDK-generated indexed-access
// message, which differs from the public one-index constructors.
func NewIndexOutOfBoundsException(index, length int) *IndexOutOfBoundsException {
	index = int(int32(index))
	length = int(int32(length))
	return NewIndexOutOfBoundsExceptionNullable(javaString(fmt.Sprintf("Index %d out of bounds for length %d", index, length)))
}

func (e *IndexOutOfBoundsException) Error() string { return javaThrowableMessage(e) }
func (e *IndexOutOfBoundsException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *IndexOutOfBoundsException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *IndexOutOfBoundsException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type StringIndexOutOfBoundsException struct{ *IndexOutOfBoundsException }

func NewStringIndexOutOfBoundsExceptionDefault() *StringIndexOutOfBoundsException {
	return NewStringIndexOutOfBoundsExceptionNullable(nil)
}

func NewStringIndexOutOfBoundsExceptionNullable(message *string) *StringIndexOutOfBoundsException {
	return &StringIndexOutOfBoundsException{IndexOutOfBoundsException: NewIndexOutOfBoundsExceptionNullable(message)}
}

func NewStringIndexOutOfBoundsExceptionForIndex(index int) *StringIndexOutOfBoundsException {
	index = int(int32(index))
	message := fmt.Sprintf("String index out of range: %d", index)
	return NewStringIndexOutOfBoundsExceptionNullable(javaString(message))
}

// NewStringIndexOutOfBoundsException carries the JDK-generated string-access
// message, which differs from the public one-index Java constructor.
func NewStringIndexOutOfBoundsException(index, length int) *StringIndexOutOfBoundsException {
	index = int(int32(index))
	length = int(int32(length))
	message := fmt.Sprintf("Index %d out of bounds for length %d", index, length)
	return NewStringIndexOutOfBoundsExceptionNullable(javaString(message))
}

func (e *StringIndexOutOfBoundsException) Error() string { return javaThrowableMessage(e) }
func (e *StringIndexOutOfBoundsException) GetMessage() *string {
	e.require()
	return e.IndexOutOfBoundsException.GetMessage()
}
func (e *StringIndexOutOfBoundsException) GetCause() error {
	e.require()
	return e.IndexOutOfBoundsException.GetCause()
}
func (e *StringIndexOutOfBoundsException) require() {
	if e == nil || e.IndexOutOfBoundsException == nil {
		panic(NewNullPointerException())
	}
}

type NegativeArraySizeException struct{ javaExceptionBase }

func NewNegativeArraySizeException(message ...string) *NegativeArraySizeException {
	return NewNegativeArraySizeExceptionNullable(optionalJavaMessage(message))
}
func NewNegativeArraySizeExceptionNullable(message *string) *NegativeArraySizeException {
	return &NegativeArraySizeException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func (e *NegativeArraySizeException) Error() string { return javaThrowableMessage(e) }
func (e *NegativeArraySizeException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NegativeArraySizeException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NegativeArraySizeException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type ClassCastException struct{ javaExceptionBase }

func NewClassCastException(message ...string) *ClassCastException {
	return NewClassCastExceptionNullable(optionalJavaMessage(message))
}
func NewClassCastExceptionNullable(message *string) *ClassCastException {
	return &ClassCastException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func (e *ClassCastException) Error() string { return javaThrowableMessage(e) }
func (e *ClassCastException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *ClassCastException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *ClassCastException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type IllegalArgumentException struct{ javaExceptionBase }

func NewIllegalArgumentException(message ...string) *IllegalArgumentException {
	return NewIllegalArgumentExceptionWithCause(optionalJavaMessage(message), nil)
}

func NewIllegalArgumentExceptionWithCause(message *string, cause error) *IllegalArgumentException {
	return &IllegalArgumentException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}

func NewIllegalArgumentExceptionFromCause(cause error) *IllegalArgumentException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewIllegalArgumentExceptionWithCause(message, cause)
}

func (e *IllegalArgumentException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalArgumentException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *IllegalArgumentException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *IllegalArgumentException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type IllegalStateException struct{ javaExceptionBase }

func NewIllegalStateException(message ...string) *IllegalStateException {
	return NewIllegalStateExceptionWithCause(optionalJavaMessage(message), nil)
}
func NewIllegalStateExceptionWithCause(message *string, cause error) *IllegalStateException {
	return &IllegalStateException{newJavaExceptionBase(message, cause)}
}
func NewIllegalStateExceptionFromCause(cause error) *IllegalStateException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewIllegalStateExceptionWithCause(message, cause)
}
func (e *IllegalStateException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalStateException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *IllegalStateException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *IllegalStateException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type IllegalMonitorStateException struct{ javaExceptionBase }

func NewIllegalMonitorStateException(message ...string) *IllegalMonitorStateException {
	return NewIllegalMonitorStateExceptionNullable(optionalJavaMessage(message))
}
func NewIllegalMonitorStateExceptionNullable(message *string) *IllegalMonitorStateException {
	return &IllegalMonitorStateException{newJavaExceptionBase(message, nil)}
}
func (e *IllegalMonitorStateException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalMonitorStateException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *IllegalMonitorStateException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *IllegalMonitorStateException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type IllegalThreadStateException struct{ javaExceptionBase }

func NewIllegalThreadStateException(message ...string) *IllegalThreadStateException {
	return NewIllegalThreadStateExceptionNullable(optionalJavaMessage(message))
}
func NewIllegalThreadStateExceptionNullable(message *string) *IllegalThreadStateException {
	return &IllegalThreadStateException{newJavaExceptionBase(message, nil)}
}

func (e *IllegalThreadStateException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalThreadStateException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *IllegalThreadStateException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *IllegalThreadStateException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type URISyntaxException struct {
	javaExceptionBase
	input  string
	reason string
	index  int
}

func NewURISyntaxException(input, reason string, index int) *URISyntaxException {
	return newURISyntaxException(javaString(input), javaString(reason), index)
}

func NewURISyntaxExceptionNoIndex(input, reason string) *URISyntaxException {
	return NewURISyntaxException(input, reason, -1)
}

func NewURISyntaxExceptionNullable(input, reason *string, index int) *URISyntaxException {
	return newURISyntaxException(input, reason, index)
}

func newURISyntaxException(input, reason *string, index int) *URISyntaxException {
	index = int(int32(index))
	if input == nil || reason == nil {
		panic(NewNullPointerException())
	}
	if index < -1 {
		panic(NewIllegalArgumentException())
	}
	return &URISyntaxException{
		javaExceptionBase: newJavaExceptionBase(reason, nil),
		input:             *input,
		reason:            *reason,
		index:             index,
	}
}

func (e *URISyntaxException) GetInput() string {
	e.require()
	return e.input
}
func (e *URISyntaxException) GetReason() string {
	e.require()
	return e.reason
}
func (e *URISyntaxException) GetIndex() int {
	e.require()
	return e.index
}
func (e *URISyntaxException) GetMessage() *string {
	e.require()
	message := e.reason
	if e.index > -1 {
		message += fmt.Sprintf(" at index %d", e.index)
	}
	message += ": " + e.input
	return javaString(message)
}
func (e *URISyntaxException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *URISyntaxException) Error() string { return javaThrowableMessage(e) }
func (e *URISyntaxException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type ArithmeticException struct{ javaExceptionBase }

func NewArithmeticException(message ...string) *ArithmeticException {
	return NewArithmeticExceptionNullable(optionalJavaMessage(message))
}
func NewArithmeticExceptionNullable(message *string) *ArithmeticException {
	return &ArithmeticException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func (e *ArithmeticException) Error() string { return javaThrowableMessage(e) }
func (e *ArithmeticException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *ArithmeticException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *ArithmeticException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type UnsupportedOperationException struct{ javaExceptionBase }

func NewUnsupportedOperationException(message ...string) *UnsupportedOperationException {
	return NewUnsupportedOperationExceptionWithCause(optionalJavaMessage(message), nil)
}
func NewUnsupportedOperationExceptionWithCause(message *string, cause error) *UnsupportedOperationException {
	return &UnsupportedOperationException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}
func NewUnsupportedOperationExceptionFromCause(cause error) *UnsupportedOperationException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewUnsupportedOperationExceptionWithCause(message, cause)
}

func (e *UnsupportedOperationException) Error() string { return javaThrowableMessage(e) }
func (e *UnsupportedOperationException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *UnsupportedOperationException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *UnsupportedOperationException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type ExecutionException struct{ javaExceptionBase }

func NewExecutionException(cause error) *ExecutionException {
	var message *string
	if cause != nil {
		message = javaString(javaThrowableString(cause))
	}
	return NewExecutionExceptionWithMessage(message, cause)
}

func NewExecutionExceptionWithMessage(message *string, cause error) *ExecutionException {
	return &ExecutionException{javaExceptionBase: newJavaExceptionBase(message, cause)}
}

func (e *ExecutionException) Error() string { return javaThrowableMessage(e) }

func (e *ExecutionException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}

func (e *ExecutionException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}

func (e *ExecutionException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type StackOverflowError struct{ javaExceptionBase }

func NewStackOverflowError(message ...string) *StackOverflowError {
	return NewStackOverflowErrorNullable(optionalJavaMessage(message))
}
func NewStackOverflowErrorNullable(message *string) *StackOverflowError {
	return &StackOverflowError{newJavaExceptionBase(message, nil)}
}

func (e *StackOverflowError) Error() string { return javaThrowableMessage(e) }
func (e *StackOverflowError) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *StackOverflowError) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *StackOverflowError) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type OutOfMemoryError struct{ javaExceptionBase }

func NewOutOfMemoryError(message ...string) *OutOfMemoryError {
	return NewOutOfMemoryErrorNullable(optionalJavaMessage(message))
}
func NewOutOfMemoryErrorNullable(message *string) *OutOfMemoryError {
	return &OutOfMemoryError{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func (e *OutOfMemoryError) Error() string { return javaThrowableMessage(e) }
func (e *OutOfMemoryError) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *OutOfMemoryError) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *OutOfMemoryError) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type NullPointerException struct{ javaExceptionBase }

func NewNullPointerException(message ...string) *NullPointerException {
	return NewNullPointerExceptionNullable(optionalJavaMessage(message))
}
func NewNullPointerExceptionNullable(message *string) *NullPointerException {
	return &NullPointerException{javaExceptionBase: newJavaExceptionBase(message, nil)}
}

func (e *NullPointerException) Error() string { return javaThrowableMessage(e) }
func (e *NullPointerException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NullPointerException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NullPointerException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

// RejectedExecutionException has both message+cause and cause-only Java
// constructors; the latter sets its detail message to the cause's toString().
type RejectedExecutionException struct{ javaExceptionBase }

func NewRejectedExecutionExceptionDefault() *RejectedExecutionException {
	return NewRejectedExecutionException(nil, nil)
}

func NewRejectedExecutionExceptionMessage(message *string) *RejectedExecutionException {
	return NewRejectedExecutionException(message, nil)
}

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

func (e *RejectedExecutionException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}

func (e *RejectedExecutionException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}

func (e *RejectedExecutionException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

// Fingerprint registration rejection is a TLC application failure. Retain its
// checked-I/O classification without implementing Java's RMI exception hierarchy.
type FPSetManagerException struct{ *IOException }

func NewFPSetManagerException(message string) *FPSetManagerException {
	return NewFPSetManagerExceptionNullable(javaString(message))
}

func NewFPSetManagerExceptionNullable(message *string) *FPSetManagerException {
	return &FPSetManagerException{IOException: &IOException{javaExceptionBase: newJavaExceptionBase(message, nil)}}
}

func (e *FPSetManagerException) Error() string { return javaThrowableMessage(e) }

func (e *FPSetManagerException) GetMessage() *string {
	e.require()
	return e.IOException.GetMessage()
}

func (e *FPSetManagerException) GetCause() error {
	e.require()
	return e.IOException.GetCause()
}

func (e *FPSetManagerException) require() {
	if e == nil || e.IOException == nil {
		panic(NewNullPointerException())
	}
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

type NamingException struct{ javaExceptionBase }

func NewNamingException(message ...string) *NamingException {
	return NewNamingExceptionNullable(optionalJavaMessage(message))
}
func NewNamingExceptionNullable(message *string) *NamingException {
	return &NamingException{newJavaExceptionBase(message, nil)}
}
func (e *NamingException) Error() string { return javaThrowableMessage(e) }
func (e *NamingException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NamingException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NamingException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type UnsupportedEncodingException struct{ *IOException }

func NewUnsupportedEncodingException(message ...string) *UnsupportedEncodingException {
	return NewUnsupportedEncodingExceptionNullable(optionalJavaMessage(message))
}
func NewUnsupportedEncodingExceptionNullable(message *string) *UnsupportedEncodingException {
	return &UnsupportedEncodingException{NewIOExceptionNullable(message)}
}
func (e *UnsupportedEncodingException) Error() string { return javaThrowableMessage(e) }
func (e *UnsupportedEncodingException) GetMessage() *string {
	e.require()
	return e.IOException.GetMessage()
}
func (e *UnsupportedEncodingException) GetCause() error {
	e.require()
	return e.IOException.GetCause()
}
func (e *UnsupportedEncodingException) require() {
	if e == nil || e.IOException == nil {
		panic(NewNullPointerException())
	}
}

type NumberFormatException struct{ *IllegalArgumentException }

func NewNumberFormatException(value string) *NumberFormatException {
	return NewNumberFormatExceptionForInputString(value, 10)
}

func NewNumberFormatExceptionDefault() *NumberFormatException {
	return NewNumberFormatExceptionNullable(nil)
}

func NewNumberFormatExceptionNullable(message *string) *NumberFormatException {
	return &NumberFormatException{NewIllegalArgumentExceptionWithCause(message, nil)}
}

func NewNumberFormatExceptionForInputString(value string, radix int) *NumberFormatException {
	radix = int(int32(radix))
	message := "For input string: \"" + value + "\""
	if radix != 10 {
		message += " under radix " + strconv.Itoa(radix)
	}
	return NewNumberFormatExceptionNullable(javaString(message))
}
func (e *NumberFormatException) Error() string { return javaThrowableMessage(e) }
func (e *NumberFormatException) GetMessage() *string {
	e.require()
	return e.IllegalArgumentException.GetMessage()
}
func (e *NumberFormatException) GetCause() error {
	e.require()
	return e.IllegalArgumentException.GetCause()
}
func (e *NumberFormatException) require() {
	if e == nil || e.IllegalArgumentException == nil {
		panic(NewNullPointerException())
	}
}
