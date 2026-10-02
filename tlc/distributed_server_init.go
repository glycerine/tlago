package tlc

import "strings"

// This is the modelCheck init catch(Throwable), separate from the callback's
// catch(Exception). It also protects call-stack replay.
func (s *TLCServer) runDistributedInit() (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	_, err = s.DoInit()
	return err
}

func (s *TLCServer) reportDistributedInitFailure(app *TLCApp, failure error) int {
	s.LastError = failure
	message := javaThrowableString(failure)
	if detail := javaThrowableDetailMessage(failure); detail != nil {
		message = *detail
	}
	code := NoError
	switch err := failure.(type) {
	case *EvalException:
		code = err.ErrorCode
	case *TLCError:
		code = err.Code
	}
	if isValueEvalException(failure) && code == ECTLCModuleTLCGetUndefined && (strings.Contains(message, "TLCSet") || strings.Contains(message, "TLCGet")) || javaRuntimeException(failure) != nil && code == ECTLCModuleValueJavaMethodOverride {
		PrintError(ECTLCFeatureUnsupported, "TLCSet & TLCGet operators not supported by distributed TLC.")
		return ECTLCFeatureUnsupported
	}
	result := ECGeneral
	if !s.HasNoErrors() {
		PrintError(ECTLCInitialState, message, s.ErrState.String())
		result = ECTLCInitialState
	} else {
		PrintError(ECGeneral, message)
	}
	app.SetCallStack()
	// TraceApp is the application in Java; retain its current tool when the
	// concrete Go trace delegates state reconstruction after this replacement.
	s.Tool = app.Tool
	if s.Trace != nil {
		s.Trace.SetTool(app.Tool)
	}
	if err := s.runDistributedInit(); err != nil {
		PrintError(ECTLCNestedExpression, app.PrintCallStack())
	}
	return result
}
