package tlc

import "fmt"

type TLCError struct {
	Code   int
	Msg    string
	Params []string
	// Runtime distinguishes Java TLCRuntimeException from the legacy native
	// EvalException carriers represented by this same Go type.
	Runtime bool
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
	return &TLCError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func newTLCErrorCode(code int, params ...string) *TLCError {
	copied := append([]string(nil), params...)
	return &TLCError{Code: code, Msg: formatMessage(code, copied), Params: copied}
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
	Code   int
	Params []string
}

func NewConfigError(code int, params ...string) *ConfigError {
	copied := append([]string(nil), params...)
	return &ConfigError{Code: code, Params: copied}
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
	if failure, ok := err.(*TLCError); ok && failure != nil && !isValueEvalException(failure) {
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
	if failure, ok := err.(*TLCError); ok && failure != nil && !failure.Runtime && failure.Params == nil {
		switch failure.Code {
		case ECSystemStackOverflow, ECSystemOutOfMemory, ECSystemOutOfMemoryLiveness, ECTLCBug:
			return failure.Code
		}
	}
	return NoError
}
