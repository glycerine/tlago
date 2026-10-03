package tlc

type ConfigFileException struct {
	throwableTrace
	detailMessage *string
	ErrorCode     int
	Parameters    []string
	Cause         error
}

func NewConfigFileException(errorCode int, parameters []string, cause ...error) *ConfigFileException {
	ex := &ConfigFileException{throwableTrace: captureThrowableTrace(), ErrorCode: errorCode, Parameters: copyMessageParameters(parameters), detailMessage: javaString(GetMessage(errorCode, parameters...))}
	if len(cause) > 0 {
		ex.Cause = cause[0]
	}
	return ex
}

func (e *ConfigFileException) Error() string {
	if e == nil {
		return ""
	}
	return *e.GetMessage()
}

func (e *ConfigFileException) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type EvalException struct {
	*StatefulRuntimeException
	ErrorCode          int
	Parameters         []string
	NullableParameters []*string
}

func NewEvalException(errorCode int, parameters ...string) *EvalException {
	var copied []string
	if parameters != nil {
		copied = copyMessageParameters(parameters)
	}
	return &EvalException{
		StatefulRuntimeException: NewStatefulRuntimeException(GetMessage(errorCode, copied...)),
		ErrorCode:                errorCode,
		Parameters:               copied,
	}
}

func NewEvalExceptionNullable(errorCode int, parameters ...*string) *EvalException {
	copied := copyNullableMessageParameters(parameters)
	return &EvalException{
		StatefulRuntimeException: NewStatefulRuntimeException(GetMessageNullable(errorCode, copied...)),
		ErrorCode:                errorCode,
		Parameters:               messageParameterStrings(copied),
		NullableParameters:       copied,
	}
}

func (e *EvalException) GetErrorCode() int {
	if e == nil {
		return NoError
	}
	return e.ErrorCode
}

func (e *EvalException) GetParameters() []string {
	if e == nil {
		return nil
	}
	return copyMessageParameters(e.Parameters)
}

func (e *EvalException) HasParameters() bool {
	return e != nil && e.Parameters != nil
}

type WorkerException struct {
	throwableTrace
	Msg           string
	Cause         error
	State1        *TLCStateMut
	State2        *TLCStateMut
	KeepCallStack bool
	messageNull   bool
}

func NewWorkerException(msg string, states ...any) *WorkerException {
	ex := &WorkerException{throwableTrace: captureThrowableTrace(), Msg: msg}
	if len(states) > 0 {
		if s, ok := states[0].(*TLCStateMut); ok {
			ex.State1 = s
		}
	}
	if len(states) > 1 {
		if s, ok := states[1].(*TLCStateMut); ok {
			ex.State2 = s
		}
	}
	if len(states) > 2 {
		if keep, ok := states[2].(bool); ok {
			ex.KeepCallStack = keep
		}
	}
	return ex
}

func NewWorkerExceptionWithCause(msg string, cause error, state1 *TLCStateMut, state2 *TLCStateMut, keep bool) *WorkerException {
	return &WorkerException{throwableTrace: captureThrowableTrace(), Msg: msg, Cause: cause, State1: state1, State2: state2, KeepCallStack: keep}
}

func (e *WorkerException) Error() string {
	if e == nil {
		return ""
	}
	if message := e.GetMessage(); message != nil {
		return *message
	}
	return javaThrowableClassName(e)
}

func (e *WorkerException) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type StatefulRuntimeException struct {
	throwableTrace
	Msg      string
	Cause    error
	known    bool
	hasMsg   bool
	hasCause bool
}

// INextStateFunctor.InvariantViolatedException is stateful; the similarly
// named DoInitFunctor exception is a separate, plain RuntimeException.
type InvariantViolatedException struct{ *StatefulRuntimeException }

func NewInvariantViolatedException() *InvariantViolatedException {
	return &InvariantViolatedException{NewStatefulRuntimeException("Invariant violated")}
}

func isInvariantViolatedException(err error) bool {
	_, ok := err.(*InvariantViolatedException)
	return ok
}

type DoInitInvariantViolatedException struct{ throwableTrace }

func NewDoInitInvariantViolatedException() *DoInitInvariantViolatedException {
	return &DoInitInvariantViolatedException{captureThrowableTrace()}
}

func (e *DoInitInvariantViolatedException) Error() string       { return javaThrowableClassName(e) }
func (e *DoInitInvariantViolatedException) GetMessage() *string { return nil }

func isDoInitInvariantViolatedException(err error) bool {
	_, ok := err.(*DoInitInvariantViolatedException)
	return ok
}

func NewStatefulRuntimeException(args ...any) *StatefulRuntimeException {
	ex := &StatefulRuntimeException{throwableTrace: captureThrowableTrace()}
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			ex.Msg = v
			ex.hasMsg = true
		case error:
			ex.Cause = v
			ex.hasCause = true
		}
	}
	return ex
}

func (e *StatefulRuntimeException) Error() string {
	if e == nil {
		return ""
	}
	if message := e.GetMessage(); message != nil {
		return *message
	}
	return javaThrowableClassName(e)
}

func (e *StatefulRuntimeException) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *StatefulRuntimeException) SetKnown() bool {
	if e == nil {
		return false
	}
	old := e.known
	e.known = true
	return old
}

func (e *StatefulRuntimeException) IsKnown() bool {
	return e != nil && e.known
}

func (e *ConfigFileException) GetMessage() *string {
	if e == nil {
		return nil
	}
	if e.detailMessage != nil {
		return javaString(*e.detailMessage)
	}
	return javaString(formatMessage(e.ErrorCode, e.Parameters))
}

func (e *WorkerException) GetMessage() *string {
	if e == nil || e.messageNull {
		return nil
	}
	return javaString(e.Msg)
}

func newWorkerExceptionFromThrowable(cause error, state1 *TLCStateMut, state2 *TLCStateMut, keep bool) *WorkerException {
	message := javaThrowableDetailMessage(cause)
	ex := NewWorkerExceptionWithCause("", cause, state1, state2, keep)
	if message == nil {
		ex.messageNull = true
	} else {
		ex.Msg = *message
	}
	return ex
}

func (e *StatefulRuntimeException) GetMessage() *string {
	if e == nil {
		return nil
	}
	if e.hasMsg || e.Msg != "" {
		return javaString(e.Msg)
	}
	return nil
}

func (e *EvalException) GetNullableParameters() []*string {
	if e == nil {
		return nil
	}
	return copyNullableMessageParameters(e.NullableParameters)
}
