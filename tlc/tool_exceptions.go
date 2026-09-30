package tlc

type WorkerException struct {
	Msg           string
	Cause         error
	State1        *TLCStateMut
	State2        *TLCStateMut
	KeepCallStack bool
}

func NewWorkerException(msg string, states ...any) *WorkerException {
	ex := &WorkerException{Msg: msg}
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
	return &WorkerException{Msg: msg, Cause: cause, State1: state1, State2: state2, KeepCallStack: keep}
}

func (e *WorkerException) Error() string {
	if e == nil {
		return ""
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "worker exception"
}

func (e *WorkerException) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type StatefulRuntimeException struct {
	Msg      string
	Cause    error
	known    bool
	hasMsg   bool
	hasCause bool
}

func NewStatefulRuntimeException(args ...any) *StatefulRuntimeException {
	ex := &StatefulRuntimeException{}
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
	if e.hasMsg {
		return e.Msg
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "stateful runtime exception"
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
