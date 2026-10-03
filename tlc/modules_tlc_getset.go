package tlc

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const TLCNamedRegisterPrefix = "s:"

var (
	tlcSetExit  = UniqueStringOf("exit")
	tlcSetPause = UniqueStringOf("pause")

	tlcGetConfig         = UniqueStringOf("config")
	tlcGetSpec           = UniqueStringOf("spec")
	tlcGetCoverage       = UniqueStringOf("coverage")
	tlcGetAction         = UniqueStringOf("action")
	tlcGetInstall        = UniqueStringOf("install")
	tlcGetID             = UniqueStringOf("id")
	tlcGetBehavior       = UniqueStringOf("behavior")
	tlcGetAll            = UniqueStringOf("all")
	tlcGetAllNamed       = UniqueStringOf("all:named")
	tlcGetMode           = UniqueStringOf("mode")
	tlcGetDeadlock       = UniqueStringOf("deadlock")
	tlcGetSeed           = UniqueStringOf("seed")
	tlcGetFingerprint    = UniqueStringOf("fingerprint")
	tlcGetWorker         = UniqueStringOf("worker")
	tlcGetTraces         = UniqueStringOf("traces")
	tlcGetDepth          = UniqueStringOf("depth")
	tlcGetAril           = UniqueStringOf("aril")
	tlcGetSched          = UniqueStringOf("sched")
	tlcGetRevision       = UniqueStringOf("revision")
	tlcGetLevel          = UniqueStringOf("level")
	tlcGetStats          = UniqueStringOf("stats")
	tlcGetDuration       = UniqueStringOf("duration")
	tlcGetGenerated      = UniqueStringOf("generated")
	tlcGetDiameter       = UniqueStringOf("diameter")
	tlcGetDistinct       = UniqueStringOf("distinct")
	tlcGetInitial        = UniqueStringOf("initial")
	tlcGetQueue          = UniqueStringOf("queue")
	tlcGetRetries        = UniqueStringOf("retries")
	tlcGetDistinctValues = UniqueStringOf("distinctvalues")
	tlcGetLevelMean      = UniqueStringOf("levelmean")
	tlcGetLevelVar       = UniqueStringOf("levelvariance")
	tlcGetCount          = UniqueStringOf("count")
	tlcRevTimestamp      = UniqueStringOf("timestamp")
	tlcRevDate           = UniqueStringOf("date")
	tlcRevTag            = UniqueStringOf("tag")
	tlcRevCalver         = UniqueStringOf("calver")
	tlcSpecInits         = UniqueStringOf("inits")
	tlcSpecActions       = UniqueStringOf("actions")
	tlcSpecTemporals     = UniqueStringOf("temporals")
	tlcSpecInvs          = UniqueStringOf("invariants")
	tlcSpecImplInits     = UniqueStringOf("impliedinits")
	tlcSpecImplActs      = UniqueStringOf("impliedactions")
	tlcSpecImplTemps     = UniqueStringOf("impliedtemporals")
	tlcSpecVars          = UniqueStringOf("variables")
	tlcSpecActCons       = UniqueStringOf("actionconstraints")
	tlcSpecCons          = UniqueStringOf("constraints")
)

// Refresh class-static register keys when installing a fresh interning context.
// Java initializes TLCGetSet within that worker or test classloader.
func initTLCGetSetUniqueStrings() {
	tlcSetExit = UniqueStringOf("exit")
	tlcSetPause = UniqueStringOf("pause")
	tlcGetConfig = UniqueStringOf("config")
	tlcGetSpec = UniqueStringOf("spec")
	tlcGetCoverage = UniqueStringOf("coverage")
	tlcGetAction = UniqueStringOf("action")
	tlcGetInstall = UniqueStringOf("install")
	tlcGetID = UniqueStringOf("id")
	tlcGetBehavior = UniqueStringOf("behavior")
	tlcGetAll = UniqueStringOf("all")
	tlcGetAllNamed = UniqueStringOf("all:named")
	tlcGetMode = UniqueStringOf("mode")
	tlcGetDeadlock = UniqueStringOf("deadlock")
	tlcGetSeed = UniqueStringOf("seed")
	tlcGetFingerprint = UniqueStringOf("fingerprint")
	tlcGetWorker = UniqueStringOf("worker")
	tlcGetTraces = UniqueStringOf("traces")
	tlcGetDepth = UniqueStringOf("depth")
	tlcGetAril = UniqueStringOf("aril")
	tlcGetSched = UniqueStringOf("sched")
	tlcGetRevision = UniqueStringOf("revision")
	tlcGetLevel = UniqueStringOf("level")
	tlcGetStats = UniqueStringOf("stats")
	tlcGetDuration = UniqueStringOf("duration")
	tlcGetGenerated = UniqueStringOf("generated")
	tlcGetDiameter = UniqueStringOf("diameter")
	tlcGetDistinct = UniqueStringOf("distinct")
	tlcGetInitial = UniqueStringOf("initial")
	tlcGetQueue = UniqueStringOf("queue")
	tlcGetRetries = UniqueStringOf("retries")
	tlcGetDistinctValues = UniqueStringOf("distinctvalues")
	tlcGetLevelMean = UniqueStringOf("levelmean")
	tlcGetLevelVar = UniqueStringOf("levelvariance")
	tlcGetCount = UniqueStringOf("count")
	tlcRevTimestamp = UniqueStringOf("timestamp")
	tlcRevDate = UniqueStringOf("date")
	tlcRevTag = UniqueStringOf("tag")
	tlcRevCalver = UniqueStringOf("calver")
	tlcSpecInits = UniqueStringOf("inits")
	tlcSpecActions = UniqueStringOf("actions")
	tlcSpecTemporals = UniqueStringOf("temporals")
	tlcSpecInvs = UniqueStringOf("invariants")
	tlcSpecImplInits = UniqueStringOf("impliedinits")
	tlcSpecImplActs = UniqueStringOf("impliedactions")
	tlcSpecImplTemps = UniqueStringOf("impliedtemporals")
	tlcSpecVars = UniqueStringOf("variables")
	tlcSpecActCons = UniqueStringOf("actionconstraints")
	tlcSpecCons = UniqueStringOf("constraints")
}

var tlcSystemProperties = struct {
	sync.Mutex
	values map[string]string
}{values: make(map[string]string)}

func TLCGet(vidx Value) (Value, error) {
	return TLCGetValue(nil, vidx, nil, nil, EvalClear)
}

func TLCGetValue(tool *Tool, vidx Value, s0 *TLCStateMut, s1 *TLCStateMut, control int) (Value, error) {
	_ = s1
	switch idx := vidx.(type) {
	case *IntValue:
		if idx.Val < 0 {
			break
		}
		if checker := MainChecker(); checker != nil {
			workerID := 0
			if id, ok := CurrentWorkerID(); ok {
				workerID = id
			}
			if value := checker.GetValue(workerID, int(idx.Val)); value != nil {
				return value, nil
			}
		}
		if simulator := CurrentSimulator(); simulator != nil {
			if value := simulator.GetLocalValue(int(idx.Val)); value != nil {
				return value, nil
			}
		}
		return nil, newTLCErrorCode(ECTLCModuleTLCGetUndefined, fmt.Sprintf("%d", idx.Val))
	case *StringValue:
		return tlcGetStringValue(tool, idx, s0, control)
	}
	return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "TLCGet", "nonnegative integer", ValuesPPR(vidx))
}

func tlcGetStringValue(tool *Tool, vidx *StringValue, s0 *TLCStateMut, control int) (Value, error) {
	key := vidx.Val
	keyString := key.String()
	checker := MainChecker()
	simulator := CurrentSimulator()

	switch key {
	case tlcGetDiameter:
		if checker != nil {
			return exactIntValueFromInt64(checker.GetProgress())
		}
		if simulator != nil {
			if worker := simulator.currentWorker(); worker != nil {
				return saturatedIntValueFromInt64(worker.GetTraceCnt()), nil
			}
			return IntZero, nil
		}
	case tlcGetGenerated:
		if checker != nil {
			return exactIntValueFromInt64(checker.GetStatesGenerated())
		}
	case tlcGetDistinct:
		if checker != nil {
			return exactIntValueFromUint64(checker.GetDistinctStatesGenerated())
		}
	case tlcGetQueue:
		if checker != nil {
			return exactIntValueFromInt64(checker.GetStateQueueSize())
		}
	case tlcGetDuration:
		return javaDurationIntValueSince(TLCStartTime()), nil
	case tlcGetStats:
		if checker != nil {
			return checker.GetStatistics(), nil
		}
		if simulator != nil {
			return simulator.GetStatistics(s0), nil
		}
	case tlcGetConfig:
		if checker != nil {
			return checker.GetConfig(), nil
		}
		if simulator != nil {
			return simulator.GetConfig(), nil
		}
	case tlcGetRevision:
		return tlcRevisionRecord(), nil
	case tlcGetSpec:
		return tlcSpecRecord(tool), nil
	case tlcGetLevel:
		if EvalIsConst(control) || EvalIsInit(control) || s0 == nil {
			return NewIntValue(int32(TLCStateInitLevel - 1)), nil
		}
		return NewIntValue(int32(s0.Level())), nil
	case tlcGetAction:
		if s0 == nil || s0.GetAction() == nil {
			return UnknownAction.ToRecordValue(), nil
		}
		return s0.GetAction().ToRecordValue(), nil
	case tlcGetAll:
		if checker != nil {
			return checker.GetAllValues(), nil
		}
		if simulator != nil {
			return simulator.GetAllValues(), nil
		}
	case tlcGetAllNamed:
		if checker != nil {
			return checker.GetAllNamedRegisterValues(), nil
		}
		if simulator != nil {
			return simulator.GetAllNamedRegisterValues(), nil
		}
	default:
		if strings.HasPrefix(keyString, TLCNamedRegisterPrefix) {
			if checker != nil {
				workerID := 0
				if id, ok := CurrentWorkerID(); ok {
					workerID = id
				}
				if value := checker.GetNamedValue(workerID, key); value != nil {
					return value, nil
				}
			}
			if simulator != nil {
				if value := simulator.GetLocalNamedValue(key); value != nil {
					return value, nil
				}
			}
		}
		if strings.HasPrefix(keyString, "-D") {
			return NewStringValue(tlcGetSystemProperty(keyString[2:], keyString)), nil
		}
	}
	return nil, newTLCErrorCode(ECTLCModuleTLCGetUndefined, keyString)
}

func TLCSet(vidx Value, val Value) (Value, error) {
	switch idx := vidx.(type) {
	case *IntValue:
		if idx.Val < 0 {
			break
		}
		if checker := MainChecker(); checker != nil {
			if workerID, ok := CurrentWorkerID(); ok {
				checker.SetValue(workerID, int(idx.Val), val)
			} else {
				checker.SetAllValues(int(idx.Val), val)
			}
		} else if simulator := CurrentSimulator(); simulator != nil {
			if worker := simulator.currentWorker(); worker != nil {
				worker.SetLocalValue(int(idx.Val), val)
			} else {
				simulator.SetAllValues(int(idx.Val), val)
			}
		} else {
			return nil, newTLCError(ECGeneral, "TLCSet cannot set integer register %d without a checker or simulator", idx.Val)
		}
		return BoolTrue, nil
	case *StringValue:
		key := idx.Val
		keyString := key.String()
		switch key {
		case tlcSetExit:
			if val == BoolTrue {
				if checker := MainChecker(); checker != nil {
					checker.Stop()
				}
				if simulator := CurrentSimulator(); simulator != nil {
					simulator.Stop()
				}
			}
			return BoolTrue, nil
		case tlcSetPause:
			if val == BoolTrue {
				if checker := MainChecker(); checker != nil {
					if err := tlcPauseModelChecker(checker); err != nil {
						return nil, err
					}
				}
			}
			return BoolTrue, nil
		default:
			if strings.HasPrefix(keyString, TLCNamedRegisterPrefix) {
				if checker := MainChecker(); checker != nil {
					if workerID, ok := CurrentWorkerID(); ok {
						checker.SetNamedValue(workerID, key, val)
					} else {
						checker.SetAllNamedValues(key, val)
					}
				} else if simulator := CurrentSimulator(); simulator != nil {
					if worker := simulator.currentWorker(); worker != nil {
						worker.SetNamedRegister(key, val)
					} else {
						simulator.SetAllNamedValues(key, val)
					}
				}
				return BoolTrue, nil
			}
			if strings.HasPrefix(keyString, "-D") {
				tlcSetSystemProperty(keyString[2:], val.String())
				return BoolTrue, nil
			}
		}
	}
	return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "TLCSet", "nonnegative integer", ValuesPPR(vidx))
}

func tlcPauseModelChecker(checker *ModelChecker) error {
	return withStateQueueMonitor(checker.StateQueue, func() error {
		fmt.Fprintln(os.Stdout, "Press enter to resume model checking.")
		_ = os.Stdout.Sync()
		var buf [1]byte
		if _, err := os.Stdin.Read(buf[:]); err != nil {
			return newTLCError(ECGeneral, "%s", err.Error())
		}
		return nil
	})
}

func withStateQueueMonitor(queue StateQueue, fn func() error) error {
	switch q := queue.(type) {
	case *MemStateQueue:
		q.mu.Lock()
		defer q.mu.Unlock()
	case *DiskStateQueue:
		q.mu.Lock()
		defer q.mu.Unlock()
	case *DiskByteArrayQueue:
		q.mu.Lock()
		defer q.mu.Unlock()
	case *StateDeque:
		q.mu.Lock()
		defer q.mu.Unlock()
	}
	return fn()
}

func TLCGetOrDefault(vidx Value, defVal Value) (Value, error) {
	switch idx := vidx.(type) {
	case *IntValue:
		workerID := 0
		if checker := MainChecker(); checker != nil {
			if id, ok := CurrentWorkerID(); ok {
				workerID = id
			}
			if value := checker.GetValue(workerID, int(idx.Val)); value != nil {
				return value, nil
			}
			return defVal, nil
		}
		if simulator := CurrentSimulator(); simulator != nil {
			if value := simulator.GetLocalValue(int(idx.Val)); value != nil {
				return value, nil
			}
			return defVal, nil
		}
	case *StringValue:
		key := idx.Val
		if key != nil && strings.HasPrefix(key.String(), TLCNamedRegisterPrefix) {
			workerID := 0
			if checker := MainChecker(); checker != nil {
				if id, ok := CurrentWorkerID(); ok {
					workerID = id
				}
				if value := checker.GetNamedValue(workerID, key); value != nil {
					return value, nil
				}
				return defVal, nil
			}
			if simulator := CurrentSimulator(); simulator != nil {
				if value := simulator.GetLocalNamedValue(key); value != nil {
					return value, nil
				}
				return defVal, nil
			}
		}
	}
	return defVal, nil
}

func tlcSpecRecord(tool *Tool) Value {
	if tool == nil {
		tool = NewTool()
	}
	names := []*UniqueString{
		tlcSpecInits,
		tlcSpecActions,
		tlcSpecTemporals,
		tlcSpecInvs,
		tlcSpecImplInits,
		tlcSpecImplTemps,
		tlcSpecVars,
		tlcSpecActCons,
		tlcSpecCons,
		tlcSpecImplActs,
	}
	values := []Value{
		initActionSetValue(tool.GetInitStateSpec()),
		nextActionSetValue(tool.GetActions()),
		propertyActionSetValue(tool.GetTemporals()),
		propertyActionSetValue(filterInternalActions(tool.GetInvariants())),
		propertyActionSetValue(tool.GetImpliedInits()),
		propertyActionSetValue(tool.GetImpliedTemporals()),
		stateVariablesSetValue(),
		semanticNodeSetValue(tool, tool.GetActionConstraints()),
		semanticNodeSetValue(tool, tool.GetModelConstraints()),
		propertyActionSetValue(tool.GetImpliedActions()),
	}
	return NewRecordValue(names, values, false)
}

func tlcRevisionRecord() Value {
	buildDate := TLCBuildDate().UTC()
	names := []*UniqueString{
		tlcGetCount,
		tlcRevTimestamp,
		tlcRevDate,
		tlcRevTag,
		tlcRevCalver,
	}
	values := []Value{
		NewIntValue(int32(TLCScmCommits())),
		NewIntValue(int32(buildDate.Unix())),
		NewStringValue(javaRevisionDate(buildDate)),
		NewStringValue(TLCRevisionOrDev()),
		NewStringValue(TLCVersionNumber()),
	}
	return NewRecordValue(names, values, false)
}

func javaRevisionDate(buildDate time.Time) string {
	buildDate = buildDate.UTC()
	millis := buildDate.Nanosecond() / int(time.Millisecond)
	return buildDate.Format("2006-01-02T15:04:05.") + strconv.Itoa(millis) + "Z"
}

func initActionSetValue(actions []*Action) Value {
	return actionSetValue(actions, initActionRecordValue)
}

func nextActionSetValue(actions []*Action) Value {
	return actionSetValue(actions, nextActionRecordValue)
}

func propertyActionSetValue(actions []*Action) Value {
	return actionSetValue(actions, propertyActionRecordValue)
}

func actionSetValue(actions []*Action, convert func(*Action) Value) Value {
	if len(actions) == 0 {
		return EmptySet
	}
	values := make([]Value, 0, len(actions))
	for _, action := range actions {
		if action == nil {
			action = UnknownAction
		}
		values = append(values, convert(action))
	}
	return NewSetEnumValue(values, false)
}

func nextActionRecordValue(action *Action) Value {
	if action != nil && action.CM.HasValues() {
		coverage := NewRecordValue(
			[]*UniqueString{tlcGetGenerated, tlcGetDistinct},
			[]Value{intValueFromInt64(action.CM.GetPrimary()), intValueFromInt64(action.CM.GetSecondary())},
			false,
		)
		return actionRecordValueWithCoverage(action, coverage)
	}
	return action.ToRecordValue()
}

func initActionRecordValue(action *Action) Value {
	if action != nil && action.CM.HasValues() {
		coverage := NewRecordValue(
			[]*UniqueString{tlcGetGenerated, tlcGetDistinct},
			[]Value{intValueFromInt64(action.CM.GetPrimary() + action.CM.GetSecondary()), intValueFromInt64(action.CM.GetPrimary())},
			false,
		)
		return actionRecordValueWithCoverage(action, coverage)
	}
	return action.ToRecordValue()
}

func propertyActionRecordValue(action *Action) Value {
	if action != nil && action.CM.HasValues() {
		child := action.CM.GetChild()
		coverage := NewRecordValue(
			[]*UniqueString{tlcGetCount},
			[]Value{intValueFromInt64(child.GetPrimary())},
			false,
		)
		return actionRecordValueWithCoverage(action, coverage)
	}
	return action.ToRecordValue()
}

func actionRecordValueWithCoverage(action *Action, coverage Value) Value {
	if action == nil {
		action = UnknownAction
	}
	return actionRecordValue(action, tlcGetCoverage, coverage)
}

func filterInternalActions(actions []*Action) []*Action {
	if len(actions) == 0 {
		return nil
	}
	out := make([]*Action, 0, len(actions))
	for _, action := range actions {
		if action == nil || !action.IsInternal() {
			out = append(out, action)
		}
	}
	return out
}

func stateVariablesSetValue() Value {
	vars := StateVariables()
	if len(vars) == 0 {
		return EmptySet
	}
	values := make([]Value, 0, len(vars))
	for _, variable := range vars {
		names := []*UniqueString{actionRecordName, actionRecordLocation}
		fields := []Value{NewStringValueFromUnique(variable.Name), sourceLocationRecordValue(variable.Location)}
		if variable.CountDistinct != nil {
			coverage := NewRecordValue(
				[]*UniqueString{tlcGetDistinct},
				[]Value{intValueFromInt64(variable.CountDistinct.Count())},
				false,
			)
			names = append(names, tlcGetCoverage)
			fields = append(fields, coverage)
		}
		values = append(values, NewRecordValue(names, fields, false))
	}
	return NewSetEnumValue(values, false)
}

func semanticNodeSetValue(tool *Tool, nodes []SemanticNode) Value {
	if len(nodes) == 0 {
		return EmptySet
	}
	values := make([]Value, 0, len(nodes))
	for _, node := range nodes {
		values = append(values, constraintRecordValue(tool, node))
	}
	return NewSetEnumValue(values, false)
}

func constraintRecordValue(tool *Tool, node SemanticNode) Value {
	switch value := SemanticToolObjectForTool(tool, node).(type) {
	case *OpDefNode:
		return opDefRecordValue(value)
	case *Action:
		return value.ToRecordValue()
	}
	return NewRecordValue(
		[]*UniqueString{actionRecordName},
		[]Value{NewStringValue(fmt.Sprint(node))},
		false,
	)
}

func opDefRecordValue(op *OpDefNode) Value {
	if op == nil {
		return EmptyRecord
	}
	name := ""
	if op.Name != nil {
		name = op.Name.String()
	}
	return NewRecordValue(
		[]*UniqueString{actionRecordName, actionRecordLocation},
		[]Value{NewStringValue(name), sourceLocationRecordValue(semanticNodeLocation(op))},
		false,
	)
}

func semanticNodeLocation(node SemanticNode) SourceLocation {
	if loc, ok := semanticNodeSourceLocation(node); ok {
		return loc
	}
	return NullSourceLocation
}

func tlcGetSystemProperty(name string, fallback string) string {
	if value, ok := tlcLookupSystemProperty(name); ok {
		return value
	}
	return fallback
}

func tlcLookupSystemProperty(name string) (string, bool) {
	tlcSystemProperties.Lock()
	defer tlcSystemProperties.Unlock()
	if value, ok := tlcSystemProperties.values[name]; ok {
		return value, true
	}
	if value, ok := os.LookupEnv(name); ok {
		return value, true
	}
	return "", false
}

func tlcSetSystemProperty(name string, value string) {
	tlcSystemProperties.Lock()
	defer tlcSystemProperties.Unlock()
	tlcSystemProperties.values[name] = value
}

func tlcSetStartupSystemProperty(name string, value string) {
	tlcSetSystemProperty(name, value)
	tlcApplyStartupSystemProperty(name, value)
}

func tlcApplyStartupSystemProperty(name string, value string) {
	if name == "file.encoding" {
		javaSetStartupFileEncoding(value)
	}
	if name == mpGeneralDebugProperty {
		mpGeneralDebug = javaBooleanProperty(value)
	}
	if name == actionCompositionProperty {
		Globals.Cdot = javaBooleanProperty(value)
	}
	if name == toolProbabilisticProperty {
		Globals.Probabilistic = javaBooleanProperty(value)
	}
	if name == tlcGlobalsCoverageProperty {
		if parsed, ok := javaIntProperty(value); ok {
			Globals.CoverageFlags = parsed
		}
	}
	if name == tlcProgressIntervalProperty {
		if seconds, ok := javaIntProperty(value); ok {
			if seconds < 0 {
				seconds = -seconds
			}
			if seconds < 1 {
				seconds = 1
			}
			Globals.ProgressIntervalMillis = int64(seconds) * 1000
		}
	}
	if name == tlcGlobalsCheckpointProperty {
		if millis, ok := javaIntProperty(value); ok {
			Globals.CheckpointDurationMillis = int64(millis)
		}
	}
}

func actionCompositionEnabled() bool {
	if value, ok := tlcLookupSystemProperty(actionCompositionProperty); ok {
		return javaBooleanProperty(value)
	}
	return Globals.Cdot
}

func javaBooleanProperty(value string) bool {
	return strings.EqualFold(value, "true")
}

func javaIntProperty(value string) (int, bool) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	return parsed, err == nil
}

func toolProbabilisticEnabled() bool {
	return ToolIsProbabilistic()
}

func intValueFromDurationSince(start time.Time) *IntValue {
	if start.IsZero() {
		return IntZero
	}
	return intValueFromInt64(int64(time.Since(start).Seconds()))
}

func intValueFromUint64(value uint64) *IntValue {
	if value > uint64(math.MaxInt32) {
		return IntNegOne
	}
	return NewIntValue(int32(value))
}

func intValueFromInt64(value int64) *IntValue {
	if int64(int32(value)) != value {
		return IntNegOne
	}
	return NewIntValue(int32(value))
}

func javaDurationIntValueSince(start time.Time) *IntValue {
	if start.IsZero() {
		return IntZero
	}
	return NewIntValue(int32(int64(time.Since(start).Seconds())))
}

func exactIntValueFromUint64(value uint64) (*IntValue, error) {
	if value > uint64(math.MaxInt32) {
		return nil, newTLCError(ECTLCModuleOverflow, "%d", value)
	}
	return NewIntValue(int32(value)), nil
}

func exactIntValueFromInt64(value int64) (*IntValue, error) {
	if value > math.MaxInt32 || value < math.MinInt32 {
		return nil, newTLCError(ECTLCModuleOverflow, "%d", value)
	}
	return NewIntValue(int32(value)), nil
}

func saturatedIntValueFromInt64(value int64) *IntValue {
	if value > math.MaxInt32 {
		return NewIntValue(math.MaxInt32)
	}
	if value < math.MinInt32 {
		return NewIntValue(math.MinInt32)
	}
	return NewIntValue(int32(value))
}
