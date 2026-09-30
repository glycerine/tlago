package tlc

import (
	"strconv"
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

func (r *ErrorTraceMessageRecorder) MCErrorTrace() (*MCError, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.errorTrace == nil || len(r.errorTrace.States) == 0 {
		return nil, false
	}
	return r.errorTrace, true
}

func (r *ErrorTraceMessageRecorder) Record(msg Message) {
	if r == nil {
		return
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
		if msg.StateNumber > 0 {
			info.StateNumber = int64(msg.StateNumber)
		}
		r.ensureTrace().AddState(NewMCStateFromTLCStateInfo(&info))
	case ECTLCStatePrint3:
		r.traceFinished = true
		if r.errorTrace == nil || len(r.errorTrace.States) == 0 {
			return
		}
		finalState := r.errorTrace.States[len(r.errorTrace.States)-1]
		r.errorTrace.AddState(NewMCStateMarker(finalState, true, false))
	case ECTLCBackToState:
		stateOrdinal := 0
		if msg.StateNumber > 0 {
			stateOrdinal = msg.StateNumber
		} else if len(msg.Params) > 0 {
			stateOrdinal, _ = strconv.Atoi(msg.Params[0])
		}
		if stateOrdinal <= 0 || r.errorTrace == nil || stateOrdinal > len(r.errorTrace.States) {
			return
		}
		r.traceFinished = true
		target := r.errorTrace.States[stateOrdinal-1]
		r.errorTrace.AddState(NewMCStateMarker(target, false, true))
	}
}

func (r *ErrorTraceMessageRecorder) ensureTrace() *MCError {
	if r.errorTrace == nil {
		r.errorTrace = NewMCError()
	}
	return r.errorTrace
}
