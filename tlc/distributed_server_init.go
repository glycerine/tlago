package tlc

import "strings"

// This is the modelCheck init catch(Throwable), separate from the callback's
// catch(Exception). The initial announcement and state count are inside that
// catch in Java; call-stack replay invokes only doInit.
func (s *TLCServer) runDistributedInit(announce bool) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	if announce {
		PrintMessage(ECTLCComputingInit)
	}
	_, err = s.DoInit()
	if err != nil {
		return err
	}
	if announce {
		PrintMessage(ECTLCInitGenerated1, fmtInt64(s.requireStateQueue().Size()), "(s)")
	}
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
		code = err.ownedCode()
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
	if s.ownedTrace() != nil {
		s.ownedTrace().SetTool(app.Tool)
	}
	if err := s.runDistributedInit(false); err != nil {
		PrintError(ECTLCNestedExpression, app.PrintCallStack())
	}
	return result
}
