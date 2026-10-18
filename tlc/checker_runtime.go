package tlc

// Checker is the shared runtime boundary of Java's AbstractChecker.
// Both concrete algorithms retain their own storage and worker objects.
type Checker interface {
	abstractChecker() *AbstractChecker
	GetValue(int, int) Value
	SetValue(int, int, Value)
	SetAllValues(int, Value)
	GetAllValue(int) []Value
	GetAllValues() Value
	GetNamedValue(int, *UniqueString) Value
	SetNamedValue(int, *UniqueString, Value)
	SetAllNamedValues(*UniqueString, Value)
	GetAllNamedValues(*UniqueString) []Value
	GetAllNamedRegisterValues() Value
	GetProgress() int64
	GetStatesGenerated() int64
	GetInitialStatesGenerated() int64
	GetStateQueueSize() int64
	distinctStateCount() int64
	GetStatistics() Value
	GetConfig() Value
	GetTraceInfo(*TLCStateMut) []*TLCStateInfo
	traceInfoPrefix(*TLCStateMut) []*TLCStateInfo
	SetErrState(*TLCStateMut, *TLCStateMut, bool, int) bool
	Stop()
}

func (c *AbstractChecker) abstractChecker() *AbstractChecker { return c }
func (mc *ModelChecker) distinctStateCount() int64           { return int64(mc.GetDistinctStatesGenerated()) }
func runtimeCheckerStatistics(c Checker) Value {
	ensureTLCGetSetUniqueStrings()
	workerID := int32(0)
	if id, ok := CurrentWorkerID(); ok {
		workerID = int32(id)
	}
	names := []*UniqueString{tlcGetQueue, tlcGetDistinct, tlcGetInitial, tlcGetGenerated, tlcGetDiameter, tlcGetDuration, tlcGetWorker}
	values := []Value{intValueFromInt64(c.GetStateQueueSize()), intValueFromInt64(c.distinctStateCount()), intValueFromInt64(c.GetInitialStatesGenerated()), intValueFromInt64(c.GetStatesGenerated()), intValueFromInt64(c.GetProgress()), intValueFromDurationSince(c.abstractChecker().ownedStartTime()), NewIntValue(workerID)}
	return NewRecordValue(names, values, false)
}

// asModelChecker matches Java's instanceof ModelChecker check. CheckImpl and
// CheckImplFile are subclasses represented by embedding in Go.
func asModelChecker(c Checker) (*ModelChecker, bool) {
	switch checker := c.(type) {
	case *ModelChecker:
		return checker, checker != nil
	case *CheckImpl:
		if checker == nil {
			return nil, false
		}
		return checker.ModelChecker, checker.ModelChecker != nil
	case *CheckImplFile:
		if checker == nil || checker.CheckImpl == nil {
			return nil, false
		}
		return checker.ModelChecker, checker.ModelChecker != nil
	default:
		return nil, false
	}
}

func requireBFSChecker(c Checker) *ModelChecker {
	if c == nil {
		return nil
	}
	if bfs, ok := asModelChecker(c); ok {
		return bfs
	}
	panic(NewClassCastException("class tlc2.tool.DFIDModelChecker cannot be cast to class tlc2.tool.ModelChecker (tlc2.tool.DFIDModelChecker and tlc2.tool.ModelChecker are in unnamed module of loader 'app')"))
}
