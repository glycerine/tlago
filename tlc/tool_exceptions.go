package tlc

type ConfigFileException struct {
	throwableTrace
	detailMessage *string
	ErrorCode     int
	Parameters    []string
	Cause         error
}

func NewConfigFileException(errorCode int, parameters []string, cause ...error) *ConfigFileException {
	if len(cause) > 1 {
		panic(NewIllegalArgumentException("NewConfigFileException accepts at most one cause"))
	}
	errorCode = int(int32(errorCode))
	ex := &ConfigFileException{throwableTrace: captureThrowableTrace(), ErrorCode: errorCode, Parameters: copyMessageParameters(parameters), detailMessage: javaString(GetMessage(errorCode, parameters...))}
	if len(cause) > 0 && !isNil(cause[0]) {
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

func (e *ConfigFileException) GetCause() error {
	e.require()
	return e.Cause
}

func (e *ConfigFileException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type EvalException struct {
	*StatefulRuntimeException
	ErrorCode          int
	Parameters         []string
	NullableParameters []*string

	ownersCaptured          bool
	errorCodeOwner          int
	parametersOwner         []string
	nullableParametersOwner []*string
}

func NewEvalException(errorCode int, parameters ...string) *EvalException {
	errorCode = int(int32(errorCode))
	return &EvalException{
		StatefulRuntimeException: NewStatefulRuntimeException(GetMessage(errorCode, parameters...)),
		ErrorCode:                errorCode,
		Parameters:               parameters,
		ownersCaptured:           true,
		errorCodeOwner:           errorCode,
		parametersOwner:          parameters,
	}
}

func NewEvalExceptionNullable(errorCode int, parameters ...*string) *EvalException {
	errorCode = int(int32(errorCode))
	strings := messageParameterStrings(parameters)
	return &EvalException{
		StatefulRuntimeException: NewStatefulRuntimeException(GetMessageNullable(errorCode, parameters...)),
		ErrorCode:                errorCode,
		Parameters:               strings,
		NullableParameters:       parameters,
		ownersCaptured:           true,
		errorCodeOwner:           errorCode,
		parametersOwner:          strings,
		nullableParametersOwner:  parameters,
	}
}

func (e *EvalException) GetErrorCode() int {
	e.require()
	if e.ownersCaptured {
		return e.errorCodeOwner
	}
	return e.ErrorCode
}

func (e *EvalException) GetParameters() []string {
	e.require()
	if e.ownersCaptured {
		return e.parametersOwner
	}
	return e.Parameters
}

func (e *EvalException) HasParameters() bool {
	e.require()
	return e.GetParameters() != nil
}

func (e *EvalException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
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
	if len(states) != 0 && len(states) != 3 {
		panic(NewIllegalArgumentException("NewWorkerException accepts either no state arguments or two states and a keep-call-stack flag"))
	}
	ex := newWorkerException(javaString(msg), nil)
	if len(states) == 0 {
		return ex
	}
	if states[0] != nil {
		state, ok := states[0].(*TLCStateMut)
		if !ok {
			panic(NewIllegalArgumentException("NewWorkerException state1 must be a TLC state"))
		}
		ex.State1 = state
	}
	if states[1] != nil {
		state, ok := states[1].(*TLCStateMut)
		if !ok {
			panic(NewIllegalArgumentException("NewWorkerException state2 must be a TLC state"))
		}
		ex.State2 = state
	}
	keep, ok := states[2].(bool)
	if !ok {
		panic(NewIllegalArgumentException("NewWorkerException keep-call-stack argument must be boolean"))
	}
	ex.KeepCallStack = keep
	return ex
}

// NewWorkerExceptionNullable preserves WorkerException(String, TLCState,
// TLCState, boolean), whose String argument may be null.
func NewWorkerExceptionNullable(msg *string, state1 *TLCStateMut, state2 *TLCStateMut, keep bool) *WorkerException {
	ex := newWorkerException(msg, nil)
	ex.State1, ex.State2, ex.KeepCallStack = state1, state2, keep
	return ex
}

func NewWorkerExceptionWithCause(msg string, cause error, state1 *TLCStateMut, state2 *TLCStateMut, keep bool) *WorkerException {
	return NewWorkerExceptionNullableWithCause(javaString(msg), cause, state1, state2, keep)
}

// NewWorkerExceptionNullableWithCause preserves WorkerException(String,
// Throwable, TLCState, TLCState, boolean), including a null detail message.
func NewWorkerExceptionNullableWithCause(msg *string, cause error, state1 *TLCStateMut, state2 *TLCStateMut, keep bool) *WorkerException {
	ex := newWorkerException(msg, cause)
	ex.State1, ex.State2, ex.KeepCallStack = state1, state2, keep
	return ex
}

func newWorkerException(msg *string, cause error) *WorkerException {
	ex := &WorkerException{throwableTrace: captureThrowableTrace(), messageNull: msg == nil}
	if !isNil(cause) {
		ex.Cause = cause
	}
	if msg != nil {
		ex.Msg = *msg
	}
	return ex
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

func (e *WorkerException) GetCause() error {
	e.require()
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

func (e *InvariantViolatedException) GetMessage() *string {
	e.require()
	return e.StatefulRuntimeException.GetMessage()
}

func (e *InvariantViolatedException) GetCause() error {
	e.require()
	return e.StatefulRuntimeException.GetCause()
}

func (e *InvariantViolatedException) SetKnown() bool {
	e.require()
	return e.StatefulRuntimeException.SetKnown()
}

func (e *InvariantViolatedException) IsKnown() bool {
	e.require()
	return e.StatefulRuntimeException.IsKnown()
}

func (e *InvariantViolatedException) require() {
	if e == nil || e.StatefulRuntimeException == nil {
		panic(NewNullPointerException())
	}
}

func IsInvariantViolatedException(err error) bool {
	switch failure := err.(type) {
	case *InvariantViolatedException:
		return failure != nil
	case *SimulationWorkerError:
		return failure != nil
	default:
		return false
	}
}

type DoInitInvariantViolatedException struct{ throwableTrace }

func NewDoInitInvariantViolatedException() *DoInitInvariantViolatedException {
	return &DoInitInvariantViolatedException{captureThrowableTrace()}
}

func (e *DoInitInvariantViolatedException) Error() string { return javaThrowableClassName(e) }

func (e *DoInitInvariantViolatedException) GetMessage() *string {
	e.require()
	return nil
}

func (e *DoInitInvariantViolatedException) GetCause() error {
	e.require()
	return nil
}

func (e *DoInitInvariantViolatedException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func isDoInitInvariantViolatedException(err error) bool {
	failure, ok := err.(*DoInitInvariantViolatedException)
	return ok && failure != nil
}

func NewStatefulRuntimeException(args ...any) *StatefulRuntimeException {
	if len(args) > 2 {
		panic(NewIllegalArgumentException("NewStatefulRuntimeException accepts at most a message and cause"))
	}
	var message *string
	var cause error
	hasCause := false
	if len(args) >= 1 {
		value, ok := args[0].(string)
		if !ok {
			panic(NewIllegalArgumentException("NewStatefulRuntimeException first argument must be a message"))
		}
		message = javaString(value)
	}
	if len(args) == 2 {
		if !isNil(args[1]) {
			value, ok := args[1].(error)
			if !ok {
				panic(NewIllegalArgumentException("NewStatefulRuntimeException second argument must be a cause"))
			}
			cause = value
		}
		hasCause = true
	}
	ex := NewStatefulRuntimeExceptionNullable(message, cause)
	ex.hasCause = hasCause
	return ex
}

// NewStatefulRuntimeExceptionNullable preserves the nullable String accepted
// by StatefulRuntimeException(String, Throwable).
func NewStatefulRuntimeExceptionNullable(message *string, cause error) *StatefulRuntimeException {
	ex := &StatefulRuntimeException{throwableTrace: captureThrowableTrace(), Cause: cause, hasMsg: message != nil, hasCause: true}
	if message != nil {
		ex.Msg = *message
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

func (e *StatefulRuntimeException) GetCause() error {
	e.require()
	return e.Cause
}

func (e *StatefulRuntimeException) SetKnown() bool {
	e.require()
	old := e.known
	e.known = true
	return old
}

func (e *StatefulRuntimeException) IsKnown() bool {
	e.require()
	return e.known
}

func (e *ConfigFileException) GetMessage() *string {
	e.require()
	if e.detailMessage != nil {
		return javaString(*e.detailMessage)
	}
	return javaString(formatMessage(e.ErrorCode, e.Parameters))
}

func (e *WorkerException) GetMessage() *string {
	e.require()
	if e.messageNull {
		return nil
	}
	return javaString(e.Msg)
}

func (e *WorkerException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func newWorkerExceptionFromThrowable(cause error, state1 *TLCStateMut, state2 *TLCStateMut, keep bool) *WorkerException {
	message := javaThrowableDetailMessage(cause)
	return NewWorkerExceptionNullableWithCause(message, cause, state1, state2, keep)
}

func (e *StatefulRuntimeException) GetMessage() *string {
	e.require()
	if e.hasMsg || e.Msg != "" {
		return javaString(e.Msg)
	}
	return nil
}

func (e *StatefulRuntimeException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

func (e *EvalException) GetNullableParameters() []*string {
	e.require()
	if e.ownersCaptured {
		return e.nullableParametersOwner
	}
	return e.NullableParameters
}
