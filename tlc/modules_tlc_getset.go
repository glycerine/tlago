package tlc

import (
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

const TLCNamedRegisterPrefix = "s:"

var (
	tlcSetExit  = UniqueStringOf("exit")
	tlcSetPause = UniqueStringOf("pause")

	tlcGetConfig     = UniqueStringOf("config")
	tlcGetSpec       = UniqueStringOf("spec")
	tlcGetCoverage   = UniqueStringOf("coverage")
	tlcGetAction     = UniqueStringOf("action")
	tlcGetInstall    = UniqueStringOf("install")
	tlcGetID         = UniqueStringOf("id")
	tlcGetBehavior   = UniqueStringOf("behavior")
	tlcGetAll        = UniqueStringOf("all")
	tlcGetAllNamed   = UniqueStringOf("all:named")
	tlcGetMode       = UniqueStringOf("mode")
	tlcGetDeadlock   = UniqueStringOf("deadlock")
	tlcGetSeed       = UniqueStringOf("seed")
	tlcGetWorker     = UniqueStringOf("worker")
	tlcGetTraces     = UniqueStringOf("traces")
	tlcGetDepth      = UniqueStringOf("depth")
	tlcGetAril       = UniqueStringOf("aril")
	tlcGetSched      = UniqueStringOf("sched")
	tlcGetLevel      = UniqueStringOf("level")
	tlcGetStats      = UniqueStringOf("stats")
	tlcGetDuration   = UniqueStringOf("duration")
	tlcGetGenerated  = UniqueStringOf("generated")
	tlcGetDiameter   = UniqueStringOf("diameter")
	tlcGetDistinct   = UniqueStringOf("distinct")
	tlcGetInitial    = UniqueStringOf("initial")
	tlcGetQueue      = UniqueStringOf("queue")
	tlcGetRetries    = UniqueStringOf("retries")
	tlcGetLevelMean  = UniqueStringOf("levelmean")
	tlcGetLevelVar   = UniqueStringOf("levelvariance")
	tlcGetCount      = UniqueStringOf("count")
	tlcSpecInits     = UniqueStringOf("inits")
	tlcSpecActions   = UniqueStringOf("actions")
	tlcSpecTemporals = UniqueStringOf("temporals")
	tlcSpecInvs      = UniqueStringOf("invariants")
	tlcSpecImplInits = UniqueStringOf("impliedinits")
	tlcSpecImplActs  = UniqueStringOf("impliedactions")
	tlcSpecImplTemps = UniqueStringOf("impliedtemporals")
	tlcSpecVars      = UniqueStringOf("variables")
	tlcSpecActCons   = UniqueStringOf("actionconstraints")
	tlcSpecCons      = UniqueStringOf("constraints")
)

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
			if value := checker.GetValue(0, int(idx.Val)); value != nil {
				return value, nil
			}
		}
		if simulator := CurrentSimulator(); simulator != nil {
			if value := simulator.GetLocalValue(int(idx.Val)); value != nil {
				return value, nil
			}
		}
		return nil, newTLCError(ECGeneral, "TLCGet(%d) is undefined", idx.Val)
	case *StringValue:
		return tlcGetStringValue(tool, idx, s0, control)
	}
	return nil, newTLCError(ECGeneral, "first argument of TLCGet must be a nonnegative integer or string, got %s", vidx)
}

func tlcGetStringValue(tool *Tool, vidx *StringValue, s0 *TLCStateMut, control int) (Value, error) {
	key := vidx.Val
	keyString := key.String()
	checker := MainChecker()
	simulator := CurrentSimulator()

	switch key {
	case tlcGetDiameter:
		if checker != nil {
			return intValueFromInt64(checker.GetProgress()), nil
		}
		if simulator != nil {
			return intValueFromInt64(simulator.TracesGenerated), nil
		}
	case tlcGetGenerated:
		if checker != nil {
			return intValueFromInt64(checker.GetStatesGenerated()), nil
		}
		if simulator != nil {
			return intValueFromInt64(simulator.StatesGenerated), nil
		}
	case tlcGetDistinct:
		if checker != nil {
			return intValueFromUint64(checker.GetDistinctStatesGenerated()), nil
		}
	case tlcGetQueue:
		if checker != nil {
			return intValueFromInt64(checker.GetStateQueueSize()), nil
		}
	case tlcGetDuration:
		return intValueFromDurationSince(TLCStartTime()), nil
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
				if value := checker.GetNamedValue(0, key); value != nil {
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
	return nil, newTLCError(ECGeneral, "TLCGet(%q) is undefined", keyString)
}

func TLCSet(vidx Value, val Value) (Value, error) {
	switch idx := vidx.(type) {
	case *IntValue:
		if idx.Val < 0 {
			break
		}
		if checker := MainChecker(); checker != nil {
			checker.SetAllValues(int(idx.Val), val)
		} else if simulator := CurrentSimulator(); simulator != nil {
			simulator.SetAllValues(int(idx.Val), val)
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
			return BoolTrue, nil
		default:
			if strings.HasPrefix(keyString, TLCNamedRegisterPrefix) {
				if checker := MainChecker(); checker != nil {
					checker.SetAllNamedValues(key, val)
				} else if simulator := CurrentSimulator(); simulator != nil {
					simulator.SetAllNamedValues(key, val)
				}
				return BoolTrue, nil
			}
			if strings.HasPrefix(keyString, "-D") {
				tlcSetSystemProperty(keyString[2:], val.String())
				return BoolTrue, nil
			}
		}
	}
	return nil, newTLCError(ECGeneral, "first argument of TLCSet must be a nonnegative integer or supported string, got %s", vidx)
}

func TLCGetOrDefault(vidx Value, defVal Value) Value {
	value, err := TLCGet(vidx)
	if err != nil || value == nil {
		return defVal
	}
	return value
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
		actionSetValue(tool.GetInitStateSpec()),
		actionSetValue(tool.GetActions()),
		actionSetValue(tool.GetTemporals()),
		actionSetValue(filterInternalActions(tool.GetInvariants())),
		actionSetValue(tool.GetImpliedInits()),
		actionSetValue(tool.GetImpliedTemporals()),
		stateVariablesSetValue(),
		semanticNodeSetValue(tool.GetActionConstraints()),
		semanticNodeSetValue(tool.GetModelConstraints()),
		actionSetValue(tool.GetImpliedActions()),
	}
	return NewRecordValue(names, values, false)
}

func actionSetValue(actions []*Action) Value {
	if len(actions) == 0 {
		return EmptySet
	}
	values := make([]Value, 0, len(actions))
	for _, action := range actions {
		if action == nil {
			action = UnknownAction
		}
		values = append(values, action.ToRecordValue())
	}
	return NewSetEnumValue(values, false)
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
		values = append(values, NewRecordValue(
			[]*UniqueString{actionRecordName},
			[]Value{NewStringValueFromUnique(variable.Name)},
			false,
		))
	}
	return NewSetEnumValue(values, false)
}

func semanticNodeSetValue(nodes []SemanticNode) Value {
	if len(nodes) == 0 {
		return EmptySet
	}
	values := make([]Value, 0, len(nodes))
	for _, node := range nodes {
		values = append(values, NewRecordValue(
			[]*UniqueString{actionRecordName},
			[]Value{NewStringValue(fmt.Sprint(node))},
			false,
		))
	}
	return NewSetEnumValue(values, false)
}

func tlcGetSystemProperty(name string, fallback string) string {
	tlcSystemProperties.Lock()
	defer tlcSystemProperties.Unlock()
	if value, ok := tlcSystemProperties.values[name]; ok {
		return value
	}
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return fallback
}

func tlcSetSystemProperty(name string, value string) {
	tlcSystemProperties.Lock()
	defer tlcSystemProperties.Unlock()
	tlcSystemProperties.values[name] = value
}

func intValueFromDurationSince(start time.Time) *IntValue {
	if start.IsZero() {
		return IntZero
	}
	return intValueFromInt64(int64(time.Since(start).Seconds()))
}

func intValueFromUint64(value uint64) *IntValue {
	if value > uint64(math.MaxInt32) {
		return NewIntValue(math.MaxInt32)
	}
	return NewIntValue(int32(value))
}

func intValueFromInt64(value int64) *IntValue {
	if value > math.MaxInt32 {
		return NewIntValue(math.MaxInt32)
	}
	if value < math.MinInt32 {
		return NewIntValue(math.MinInt32)
	}
	return NewIntValue(int32(value))
}
