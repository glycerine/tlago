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
