package tlc

import "testing"

func TestCounterExampleInvalidPositiveLoopOrdinalFailsLikeJava(t *testing.T) {
	initTLCCheckerTest(t)
	trace := []*TLCStateInfo{NewTLCStateInfo(checkerTestState(1))}
	expectPanic(t, func() {
		_ = NewCounterExample(trace, UnknownAction, 2, true)
	})
}
