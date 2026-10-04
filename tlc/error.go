package tlc

import "fmt"

type TLCError struct {
	throwableTrace
	Code           int
	Msg            string
	Params         []string
	NullableParams []*string
	// Runtime distinguishes Java TLCRuntimeException from the legacy native
	// EvalException carriers represented by this same Go type.
	Runtime bool
	// Java's detailed runtime failure retains the failed semantic expression
	// and its evaluation context on the same TLCRuntimeException carrier.
	Expr  SemanticNode
	Ctxt  *Context
	known bool
}

func NewTLCDetailedRuntimeException(code int, message string, expr SemanticNode, ctxt *Context) *TLCError {
	failure := newTLCError(code, "%s", message)
	failure.Runtime, failure.Expr, failure.Ctxt = true, expr, ctxt
	return failure
}

func (e *TLCError) Error() string {
	if e == nil {
		return ""
	}
	if e.Msg == "" {
		return fmt.Sprintf("TLC error %d", e.Code)
	}
	return e.Msg
}

func newTLCError(code int, format string, args ...any) *TLCError {
	return &TLCError{throwableTrace: captureThrowableTrace(), Code: code, Msg: fmt.Sprintf(format, args...)}
}

func newTLCErrorCode(code int, params ...string) *TLCError {
	copied := copyMessageParameters(params)
	return &TLCError{throwableTrace: captureThrowableTrace(), Code: code, Msg: GetMessage(code, copied...), Params: copied}
}

// NewTLCRuntimeException constructs the coded failure used by Java Assert.
func NewTLCRuntimeException(code int, params ...string) *TLCError {
	failure := newTLCErrorCode(code, params...)
	failure.Runtime = true
	return failure
}

func javaMethodOverrideError(signature string, message string) *TLCError {
	return newTLCErrorCode(ECTLCModuleValueJavaMethodOverride, signature, message)
}

func javaMethodOverrideRuntimeError(signature string, message string) *TLCError {
	err := javaMethodOverrideError(signature, message)
	err.Runtime = true
	return err
}

type ConfigError struct {
	throwableTrace
	Code   int
	Params []string
}

func NewConfigError(code int, params ...string) *ConfigError {
	copied := copyMessageParameters(params)
	return &ConfigError{throwableTrace: captureThrowableTrace(), Code: code, Params: copied}
}

func (e *ConfigError) Error() string {
	if e == nil {
		return ""
	}
	return formatMessage(e.Code, e.Params)
}

// javaRuntimeException tests the thrown exception itself. TLCError also carries
// legacy EvalExceptions, which must not enter Java's TLCRuntimeException catch.
func javaRuntimeException(err error) *TLCError {
	if failure, ok := err.(*TLCError); ok && failure != nil && javaSystemFailureCode(failure) == NoError && !isValueEvalException(failure) {
		return failure
	}
	return nil
}

func javaRuntimeFailureMessage(err *TLCError) (int, []string) {
	if err.Params != nil {
		return err.Code, err.Params
	}
	if err.Code == ECGeneral {
		return err.Code, generalErrorParams("", err)
	}
	return err.Code, []string{err.Error()}
}

// Go's existing system-failure carriers represent Java Error subclasses.
// Numeric codes on actual EvalExceptions or TLCRuntimeExceptions do not give
// those exceptions the type of a StackOverflowError or OutOfMemoryError.
func javaSystemFailureCode(err error) int {
	if failure, ok := err.(*StackOverflowError); ok && failure != nil {
		return ECSystemStackOverflow
	}
	if failure, ok := err.(*OutOfMemoryError); ok && failure != nil {
		return ECSystemOutOfMemory
	}
	if failure, ok := err.(*TLCError); ok && failure != nil && !failure.Runtime && failure.Params == nil {
		switch failure.Code {
		case ECSystemStackOverflow, ECSystemOutOfMemory, ECSystemOutOfMemoryLiveness, ECTLCBug:
			return failure.Code
		}
	}
	return NoError
}

func (e *TLCError) SetKnown() bool {
	if e == nil {
		return false
	}
	old := e.known
	e.known = true
	return old
}

func (e *TLCError) IsKnown() bool {
	return e != nil && e.known
}

func isJavaEvalOrRuntimeException(err error) bool {
	switch failure := err.(type) {
	case *EvalException:
		return failure != nil
	case *TLCError:
		return failure != nil && javaSystemFailureCode(err) == NoError
	default:
		return false
	}
}

// ClassifyEvaluationFailure preserves the Java EvalException and
// TLCRuntimeException catches at native API boundaries. Java Error subclasses
// and unrelated exceptions return neither classification.
func ClassifyEvaluationFailure(err error) (eval bool, runtime *TLCError) {
	if !isJavaEvalOrRuntimeException(err) {
		return false, nil
	}
	if _, ok := err.(*EvalException); ok {
		return true, nil
	}
	failure := err.(*TLCError)
	if isValueEvalException(failure) {
		return true, nil
	}
	return false, failure
}

func (e *TLCError) GetMessage() *string {
	if e == nil {
		return nil
	}
	if javaSystemFailureCode(e) != NoError && e.Msg == "" {
		return nil
	}
	return javaString(e.Msg)
}

func (e *ConfigError) GetMessage() *string {
	if e == nil {
		return nil
	}
	return javaString(formatMessage(e.Code, e.Params))
}

func newTLCErrorCodeNullable(code int, params ...*string) *TLCError {
	copied := copyNullableMessageParameters(params)
	return &TLCError{
		throwableTrace: captureThrowableTrace(),
		Code:           code,
		Msg:            GetMessageNullable(code, copied...),
		Params:         messageParameterStrings(copied),
		NullableParams: copied,
	}
}

func printJavaRuntimeException(err *TLCError) int {
	if err.NullableParams != nil {
		return PrintErrorNullable(err.Code, err.NullableParams...)
	}
	if err.Params != nil {
		return PrintError(err.Code, err.Params...)
	}
	if err.Code == ECGeneral {
		return PrintError(err.Code, generalErrorParams("", err)...)
	}
	return PrintErrorNullable(err.Code, err.GetMessage())
}
