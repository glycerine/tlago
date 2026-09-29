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
