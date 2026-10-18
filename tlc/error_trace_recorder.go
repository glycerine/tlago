package tlc

import (
	"sync"
)

type ErrorTraceMessageRecorder struct {
	mu            sync.Mutex
	errorTrace    *MCError
	traceFinished bool
}

func NewErrorTraceMessageRecorder() *ErrorTraceMessageRecorder {
	return &ErrorTraceMessageRecorder{}
}

// GetMCErrorTrace is ErrorTraceMessagePrinterRecorder.getMCErrorTrace. The bool
// represents Java Optional presence.
func (r *ErrorTraceMessageRecorder) GetMCErrorTrace() (*MCError, bool) {
	if r == nil {
		panic(NewNullPointerException())
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.errorTrace == nil || len(r.errorTrace.GetStates()) == 0 {
		return nil, false
	}
	return r.errorTrace, true
}

func (r *ErrorTraceMessageRecorder) MCErrorTrace() (*MCError, bool) {
	return r.GetMCErrorTrace()
}

func (r *ErrorTraceMessageRecorder) Record(msg Message) {
	if r == nil {
		panic(NewNullPointerException())
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.traceFinished {
		return
	}
	switch msg.Code {
	case ECTLCStatePrint2:
		if msg.StateInfo == nil {
			return
		}
		info := *msg.StateInfo
		info.StateNumber = int64(int32(msg.StateNumber))
		r.ensureTrace().AddState(NewMCStateFromTLCStateInfo(&info))
	case ECTLCStatePrint3:
		r.traceFinished = true
		if r.errorTrace == nil || len(r.errorTrace.GetStates()) == 0 {
			return
		}
		states := r.errorTrace.GetStates()
		finalState := states[len(states)-1]
		r.errorTrace.AddState(NewMCStateMarker(finalState, true, false))
	case ECTLCBackToState:
		stateOrdinal := 0
		hasStateOrdinal := false
		if msg.StateInfo != nil {
			stateOrdinal = int(int32(msg.StateNumber))
			hasStateOrdinal = true
		} else if len(msg.Params) >= 2 {
			if parsed, ok := javaParseDecimalInt(msg.Params[0]); ok {
				stateOrdinal = int(parsed)
				hasStateOrdinal = true
			}
		}
		if !hasStateOrdinal {
			return
		}
		r.traceFinished = true
		if r.errorTrace == nil || stateOrdinal <= 0 || stateOrdinal > len(r.errorTrace.GetStates()) {
			return
		}
		target := r.errorTrace.GetStates()[stateOrdinal-1]
		r.errorTrace.AddState(NewMCStateMarker(target, false, true))
	}
}

func (r *ErrorTraceMessageRecorder) ensureTrace() *MCError {
	if r.errorTrace == nil {
		r.errorTrace = NewMCError()
	}
	return r.errorTrace
}
