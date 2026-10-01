package tlc

import "fmt"

type TLCError struct {
	Code int
	Msg  string
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
	return &TLCError{Code: code, Msg: formatMessage(code, params)}
}

func javaMethodOverrideError(signature string, message string) *TLCError {
	return newTLCErrorCode(ECTLCModuleValueJavaMethodOverride, signature, message)
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
