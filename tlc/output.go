package tlc

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	NoError = 0

	ECUnknown  = -1
	ECUnitTest = -123456

	ECGeneral                      = 1000
	ECSystemOutOfMemory            = 1001
	ECSystemOutOfMemoryTooManyInit = 1002
	ECSystemOutOfMemoryLiveness    = 1003
	ECSystemStackOverflow          = 1005

	ECWrongCommandlineParamsSimulator = 1101
	ECWrongCommandlineParamsTLC       = 1102

	ECTLCMetadirExists                                   = 2100
	ECTLCMetadirCanNotBeCreated                          = 2101
	ECTLCInitialState                                    = 2102
	ECTLCNestedExpression                                = 2103
	ECTLCAssumptionFalse                                 = 2104
	ECTLCAssumptionEvaluationError                       = 2105
	ECTLCStateNotCompletelySpecifiedInitial              = 2106
	ECTLCInvariantViolatedInitial                        = 2107
	ECTLCPropertyViolatedInitial                         = 2108
	ECTLCStateNotCompletelySpecifiedNext                 = 2109
	ECTLCInvariantViolatedBehavior                       = 2110
	ECTLCInvariantEvaluationFailed                       = 2111
	ECTLCActionPropertyViolatedBehavior                  = 2112
	ECTLCActionPropertyEvaluationFailed                  = 2113
	ECTLCDeadlockReached                                 = 2114
	ECTLCStatesAndNoNextAction                           = 2115
	ECTLCTemporalPropertyViolated                        = 2116
	ECTLCFailedToRecoverNext                             = 2117
	ECTLCNoStatesSatisfyingInit                          = 2118
	ECTLCStringModuleNotFound                            = 2119
	ECTLCErrorState                                      = 2120
	ECTLCBehaviorUpToThisPoint                           = 2121
	ECTLCBackToState                                     = 2122
	ECTLCFailedToRecoverInit                             = 2123
	ECTLCReporterDied                                    = 2124
	ECSystemErrorReadingPool                             = 2125
	ECSystemCheckpointRecoveryCorrupt                    = 2126
	ECSystemErrorWritingPool                             = 2127
	ECTLCBug                                             = 2128
	ECSystemDiskgraphAccess                              = 2129
	ECTLCRegistryInitError                               = 2131
	ECTLCValueAssertFailed                               = 2132
	ECTLCFPNotInSet                                      = 2133
	ECSystemIndexError                                   = 2134
	ECSystemStreamEmpty                                  = 2135
	ECTLCParameterMustBePostfix                          = 2136
	ECSystemFileNull                                     = 2137
	ECSystemInterrupted                                  = 2138
	ECTLCCouldNotDetermineSubscript                      = 2139
	ECTLCSubscriptContainNoStateVar                      = 2140
	ECTLCWrongTupleFieldName                             = 2141
	ECTLCWrongRecordFieldName                            = 2142
	ECTLCUnchangedVariableChanged                        = 2143
	ECTLCExceptAppliedToUnknownField                     = 2144
	ECTLCModuleTLCGetUndefined                           = 2145
	ECTLCInvariantViolatedLevel                          = 2146
	ECTLCFingerprintException                            = 2147
	ECTLCStateNotCompletelySpecifiedLive                 = 2148
	ECTLCInvariantConstantLevel                          = 2149
	ECTLCModuleValueJavaMethodOverride                   = 2154
	ECTLCModuleCompareValue                              = 2155
	ECTLCFeatureUnsupported                              = 2156
	ECTLCModuleTransitiveClosure                         = 2157
	ECTLCModuleCheckMemberOf                             = 2158
	ECTLCLiveBEGraphFailedToConstruct                    = 2159
	ECSystemUnableNotRenameFile                          = 2160
	ECSystemDiskIOErrorForFile                           = 2161
	ECSystemMetadirExists                                = 2162
	ECSystemMetadirCreationError                         = 2163
	ECTLCChooseArgumentsWrong                            = 2164
	ECTLCChooseUpperBound                                = 2165
	ECTLCFPValueAlreadyOnDisk                            = 2166
	ECSystemUnableToOpenFile                             = 2167
	ECTLCModuleValueJavaMethodOverrideLoaded             = 2168
	ECTLCModuleArgumentError                             = 2169
	ECTLCArgumentMismatch                                = 2170
	ECTLCParsingFailed2                                  = 2171
	ECTLCTooManyPossibleStates                           = 2172
	ECTLCErrorReplacingModules                           = 2173
	ECSystemErrorReadingStates                           = 2174
	ECSystemErrorWritingStates                           = 2175
	ECTLCModuleApplyingToWrongValue                      = 2176
	ECTLCModuleBagUnion1                                 = 2177
	ECTLCModuleOverflow                                  = 2178
	ECTLCModuleDivisionByZero                            = 2179
	ECTLCModuleNullPowerNull                             = 2180
	ECTLCModuleComputingCardinality                      = 2181
	ECTLCModuleEvaluating                                = 2182
	ECTLCModuleArgumentNotInDomain                       = 2183
	ECTLCModuleApplyEmptySeq                             = 2184
	ECTLCStarting                                        = 2185
	ECTLCFinished                                        = 2186
	ECTLCModeMC                                          = 2187
	ECTLCModeSimu                                        = 2188
	ECTLCComputingInit                                   = 2189
	ECTLCInitGenerated1                                  = 2190
	ECTLCInitGenerated2                                  = 2191
	ECTLCCheckingTemporalProps                           = 2192
	ECTLCSuccess                                         = 2193
	ECTLCSearchDepth                                     = 2194
	ECTLCCheckpointStart                                 = 2195
	ECTLCCheckpointEnd                                   = 2196
	ECTLCCheckpointRecoverStart                          = 2197
	ECTLCCheckpointRecoverEnd                            = 2198
	ECTLCStats                                           = 2199
	ECTLCProgressStats                                   = 2200
	ECTLCCoverageStart                                   = 2201
	ECTLCCoverageEnd                                     = 2202
	ECTLCCoverageValue                                   = 2221
	ECTLCCheckpointRecoverEndDFID                        = 2203
	ECTLCStatsDFID                                       = 2204
	ECTLCProgressStartStatsDFID                          = 2205
	ECTLCProgressStatsDFID                               = 2206
	ECTLCInitGenerated3                                  = 2207
	ECTLCInitGenerated4                                  = 2208
	ECTLCProgressSimu                                    = 2209
	ECTLCStatsSimu                                       = 2210
	ECTLCFPCompleted                                     = 2211
	ECTLCLiveImplied                                     = 2212
	ECTLCCounterExample                                  = 2264
	ECTLCStatePrint1                                     = 2216
	ECTLCStatePrint2                                     = 2217
	ECTLCStatePrint3                                     = 2218
	ECTLCSanyEnd                                         = 2219
	ECTLCConfigNotBothSpecAndInit                        = 2227
	ECTLCConfigIDRequiresNoArg                           = 2228
	ECTLCConfigSpecifiedNotDefined                       = 2229
	ECTLCConfigIDHasValue                                = 2230
	ECTLCConfigMissingInit                               = 2231
	ECTLCConfigMissingNext                               = 2232
	ECTLCConfigIDMustNotBeConstant                       = 2233
	ECTLCNoStatesSatisfyingInitAndConstraint             = 2256
	ECTLCModuleArgumentErrorAn                           = 2266
	ECTLCCheckingTemporalPropsEnd                        = 2267
	ECTLCStateGraphOutdegree                             = 2268
	ECTLCComputingInitProgress                           = 2269
	ECSystemErrorCleaningPool                            = 2270
	ECTLCModeMCDFS                                       = 2271
	ECTLCFeatureUnsupportedLivenessSymmetry              = 2279
	ECTLCTraceTooLong                                    = 2282
	ECTLCModuleOneArgumentError                          = 2283
	ECTLCFeatureLivenessConstraints                      = 2284
	ECTLCSymmetrySetTooSmall                             = 2300
	ECTLCSpecificationFeaturesTemporalQuantifier         = 2301
	ECTLCModuleValueJavaMethodOverrideMismatch           = 2400
	ECTLCModuleValueJavaMethodOverrideModuleMismatch     = 2402
	ECTLCModuleValueJavaMethodOverrideIdentifierMismatch = 2403
	ECTLCPostconditionFalse                              = 2404
	ECTLCPostconditionEvaluationError                    = 2405
	ECTLCTESpecGenerationComplete                        = 2501
	ECTLCTESpecGenerationError                           = 2502
	ECTLCCoverageNext                                    = 2772
	ECTLCCoverageInit                                    = 2773
	ECTLCCoverageProperty                                = 2774
	ECTLCCoverageValueCost                               = 2775
	ECTLCCoverageMismatch                                = 2776
	ECTLCCoverageEndOverhead                             = 2777
	ECTLCCoverageConstraint                              = 2778
	ECTLCCoverageVar                                     = 2779
	ECTLCPossibleUnwitnessed                             = 2780
	ECTLCParsingFailed                                   = 3002

	ECCFGErrorReadingFile = 5001
	ECCFGGeneral          = 5002
	ECCFGMissingID        = 5003
	ECCFGTwiceKeyword     = 5004
	ECCFGExpectID         = 5005
	ECCFGExpectedSymbol   = 5006

	ECTLCModuleOverrideStdout = 20000
)

type Severity int

const (
	SeverityNone Severity = iota
	SeverityError
	SeverityTLCBug
	SeverityWarning
	SeverityState
)

type Message struct {
	Code        int
	Severity    Severity
	Params      []string
	Text        string
	State       *TLCStateMut
	StateInfo   *TLCStateInfo
	StateNumber int
}

type MessageRecorder interface {
	Record(Message)
}

type RecorderFunc func(Message)

func (f RecorderFunc) Record(msg Message) {
	f(msg)
}

type BroadcastRecorder struct {
	mu        sync.Mutex
	recorders []MessageRecorder
}

func (b *BroadcastRecorder) Add(rec MessageRecorder) {
	if rec == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, existing := range b.recorders {
		if sameMessageRecorder(existing, rec) {
			return
		}
	}
	b.recorders = append(b.recorders, rec)
}

func (b *BroadcastRecorder) Remove(rec MessageRecorder) {
	if rec == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, existing := range b.recorders {
		if sameMessageRecorder(existing, rec) {
			copy(b.recorders[i:], b.recorders[i+1:])
			b.recorders[len(b.recorders)-1] = nil
			b.recorders = b.recorders[:len(b.recorders)-1]
			return
		}
	}
}

func sameMessageRecorder(a MessageRecorder, b MessageRecorder) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	if !av.IsValid() || !bv.IsValid() || av.Type() != bv.Type() {
		return false
	}
	if av.Comparable() {
		return av.Interface() == bv.Interface()
	}
	if av.Kind() == reflect.Func {
		return av.Pointer() == bv.Pointer()
	}
	return false
}

func (b *BroadcastRecorder) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.recorders = nil
}

func (b *BroadcastRecorder) Record(msg Message) {
	b.mu.Lock()
	recorders := append([]MessageRecorder(nil), b.recorders...)
	b.mu.Unlock()

	for _, rec := range recorders {
		rec.Record(msg)
	}
}

type MemoryRecorder struct {
	mu       sync.Mutex
	Messages []Message
}

func (r *MemoryRecorder) Record(msg Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Messages = append(r.Messages, msg)
}

func (r *MemoryRecorder) Recorded(code int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, msg := range r.Messages {
		if msg.Code == code {
			return true
		}
	}
	return false
}

func (r *MemoryRecorder) Records(code int) []Message {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Message
	for _, msg := range r.Messages {
		if msg.Code == code {
			out = append(out, msg)
		}
	}
	return out
}

var defaultRecorder BroadcastRecorder

func AddMessageRecorder(rec MessageRecorder) {
	defaultRecorder.Add(rec)
}

func RemoveMessageRecorder(rec MessageRecorder) {
	defaultRecorder.Remove(rec)
}

func ClearMessageRecorders() {
	defaultRecorder.Clear()
}

func PrintMessage(code int, params ...string) int {
	recordMessage(code, SeverityNone, params...)
	return code
}

func PrintWarning(code int, params ...string) int {
	recordMessage(code, SeverityWarning, params...)
	return code
}

func PrintError(code int, params ...string) int {
	recordMessage(code, SeverityError, params...)
	return code
}

func PrintState(code int, params []string, state *TLCStateMut, stateNumber int) string {
	text := formatMessage(code, params)
	recordStateMessage(code, params, text, state, nil, stateNumber)
	return text
}

func PrintStateInfo(code int, params []string, info *TLCStateInfo, stateNumber int) string {
	text := formatMessage(code, params)
	var state *TLCStateMut
	if info != nil {
		state = info.State
	}
	recordStateMessage(code, params, text, state, info, stateNumber)
	return text
}

func recordMessage(code int, severity Severity, params ...string) {
	suppressed, asError, warn := messageControlFor(code)
	if severity == SeverityWarning {
		if !warn || suppressed {
			return
		}
	} else if suppressed {
		return
	}
	if asError && severity != SeverityError {
		severity = SeverityError
	}
	copied := append([]string(nil), params...)
	defaultRecorder.Record(Message{
		Code:     code,
		Severity: severity,
		Params:   copied,
		Text:     formatMessage(code, copied),
	})
}

func recordStateMessage(code int, params []string, text string, state *TLCStateMut, info *TLCStateInfo, stateNumber int) {
	copied := append([]string(nil), params...)
	defaultRecorder.Record(Message{
		Code:        code,
		Severity:    SeverityState,
		Params:      copied,
		Text:        text,
		State:       state,
		StateInfo:   info,
		StateNumber: stateNumber,
	})
}

func formatMessage(code int, params []string) string {
	switch code {
	case ECTLCConfigNotBothSpecAndInit:
		return "The configuration file cannot specify both INIT/NEXT and SPECIFICATION fields."
	case ECTLCConfigIDRequiresNoArg:
		if len(params) == 1 {
			return fmt.Sprintf("TLC requires %s not to take any argument.", params[0])
		}
		if len(params) >= 2 {
			return fmt.Sprintf("TLC requires %s not to take any argument, but one was given: %s", params[0], params[1])
		}
	case ECTLCConfigSpecifiedNotDefined:
		if len(params) >= 2 {
			return fmt.Sprintf("The %s %s specified in the configuration file\nis not defined in the specification.", params[0], params[1])
		}
	case ECTLCConfigIDHasValue:
		if len(params) >= 3 {
			return fmt.Sprintf("The %s of %s is equal to %s", params[0], params[1], params[2])
		}
	case ECTLCConfigMissingInit:
		return "The configuration file did not specify the initial state predicate."
	case ECTLCConfigMissingNext:
		return "The configuration file did not specify the next state predicate."
	case ECTLCConfigIDMustNotBeConstant:
		if len(params) >= 2 {
			return fmt.Sprintf("The %s %s cannot be a constant.", params[0], params[1])
		}
	case ECTLCModuleOverflow:
		if len(params) >= 1 {
			return fmt.Sprintf("Overflow when computing %s", params[0])
		}
	case ECTLCModuleOneArgumentError:
		if len(params) >= 3 {
			return fmt.Sprintf("The argument of %s should be a %s, but instead it is:\n%s", params[0], params[1], params[2])
		}
	case ECTLCModuleArgumentError:
		if len(params) >= 4 {
			return fmt.Sprintf("The %s argument of %s should be a %s, but instead it is:\n%s", params[0], params[1], params[2], params[3])
		}
	case ECTLCModuleArgumentErrorAn:
		if len(params) >= 4 {
			return fmt.Sprintf("The %s argument of %s should be an %s, but instead it is:\n%s", params[0], params[1], params[2], params[3])
		}
	case ECTLCModuleArgumentNotInDomain:
		if len(params) >= 5 {
			return fmt.Sprintf("The %s argument of %s must be in the domain of its %s argument:\n%s\n, but instead it is\n%s", params[0], params[1], params[2], params[3], params[4])
		}
	case ECTLCModuleDivisionByZero:
		return "The second argument of \\div is 0."
	case ECTLCModuleNullPowerNull:
		return "0^0 is undefined."
	case ECTLCModuleApplyEmptySeq:
		if len(params) >= 1 {
			return fmt.Sprintf("Attempted to apply %s to the empty sequence.", params[0])
		}
	case ECTLCModuleCompareValue:
		if len(params) >= 2 {
			return fmt.Sprintf("Attempted to compare %s with the value\n%s", params[0], params[1])
		}
	case ECTLCModuleCheckMemberOf:
		if len(params) >= 2 {
			return fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of %s.", params[0], params[1])
		}
	case ECTLCModuleComputingCardinality:
		if len(params) >= 1 {
			return fmt.Sprintf("Attempted to compute cardinality of the value\n%s", params[0])
		}
	case ECTLCModuleEvaluating:
		if len(params) >= 3 {
			return fmt.Sprintf("Evaluating an expression of the form %s when s is not a %s:\n%s", params[0], params[1], params[2])
		}
	}
	if len(params) == 0 {
		return fmt.Sprintf("%d", code)
	}
	return fmt.Sprintf("%d %s", code, strings.Join(params, " "))
}

const (
	ExitStatusError                = 255
	ExitStatusSuccess              = 0
	ExitStatusViolationAssumption  = 10
	ExitStatusViolationDeadlock    = 11
	ExitStatusViolationSafety      = 12
	ExitStatusViolationLiveness    = 13
	ExitStatusViolationAssert      = 14
	ExitStatusFailureSpecEval      = 75
	ExitStatusFailureSafetyEval    = 76
	ExitStatusFailureLivenessEval  = 77
	ExitStatusErrorSpecParse       = 150
	ExitStatusErrorConfigParse     = 151
	ExitStatusErrorStateSpaceLarge = 152
	ExitStatusErrorSystem          = 153
)

func ExitStatusForErrorCode(code int) int {
	switch code {
	case NoError:
		return ExitStatusSuccess
	case ECTLCStateNotCompletelySpecifiedNext,
		ECTLCStatesAndNoNextAction,
		ECTLCNestedExpression,
		ECTLCFingerprintException:
		return ExitStatusFailureSpecEval
	case ECTLCInvariantEvaluationFailed,
		ECTLCInvariantViolatedLevel:
		return ExitStatusFailureSafetyEval
	case ECTLCInvariantViolatedInitial,
		ECTLCInvariantViolatedBehavior:
		return ExitStatusViolationSafety
	case ECTLCActionPropertyViolatedBehavior,
		ECTLCActionPropertyEvaluationFailed,
		ECTLCTemporalPropertyViolated,
		ECTLCPropertyViolatedInitial:
		return ExitStatusViolationLiveness
	case ECTLCDeadlockReached:
		return ExitStatusViolationDeadlock
	case ECTLCAssumptionFalse,
		ECTLCAssumptionEvaluationError,
		ECTLCPostconditionFalse,
		ECTLCPostconditionEvaluationError,
		ECTLCPossibleUnwitnessed:
		return ExitStatusViolationAssumption
	case ECTLCValueAssertFailed:
		return ExitStatusViolationAssert
	case ECCFGErrorReadingFile,
		ECCFGGeneral,
		ECCFGMissingID,
		ECCFGTwiceKeyword,
		ECCFGExpectID,
		ECCFGExpectedSymbol,
		ECTLCConfigNotBothSpecAndInit,
		ECTLCConfigIDRequiresNoArg,
		ECTLCConfigSpecifiedNotDefined,
		ECTLCConfigIDHasValue,
		ECTLCConfigMissingInit,
		ECTLCConfigMissingNext,
		ECTLCConfigIDMustNotBeConstant:
		return ExitStatusErrorConfigParse
	case ECTLCParsingFailed2,
		ECTLCParsingFailed:
		return ExitStatusErrorSpecParse
	default:
		return ExitStatusError
	}
}
