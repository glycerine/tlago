// Portions Copyright (c) 2025, Oracle and/or its affiliates.
package tlc

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
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

	ECTLCPPParsingValue   = 2000
	ECTLCPPFormatingValue = 2001

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
	ECTLCAAAAAAA                                         = 2130
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
	ECTLCLiveImpliedDebug                                = 2263
	ECTLCLiveCannotHandleFormula                         = 2213
	ECTLCLiveWrongFormulaFormat                          = 2214
	ECTLCExpectedValue                                   = 2215
	ECTLCCounterExample                                  = 2264
	ECTLCStatePrint1                                     = 2216
	ECTLCStatePrint2                                     = 2217
	ECTLCStatePrint3                                     = 2218
	ECTLCSanyEnd                                         = 2219
	ECTLCSanyStart                                       = 2220
	ECTLCConfigNotBothSpecAndInit                        = 2227
	ECTLCConfigValueNotAssignedToConstantParam           = 2222
	ECTLCConfigRHSIDAppearedAfterLHSID                   = 2223
	ECTLCConfigWrongSubstitution                         = 2224
	ECTLCConfigWrongSubstitutionNumberOfArgs             = 2225
	ECTLCConfigIDDoesNotAppearInSpec                     = 2226
	ECTLCConfigIDRequiresNoArg                           = 2228
	ECTLCConfigSpecifiedNotDefined                       = 2229
	ECTLCConfigIDHasValue                                = 2230
	ECTLCConfigMissingInit                               = 2231
	ECTLCConfigMissingNext                               = 2232
	ECTLCConfigIDMustNotBeConstant                       = 2233
	ECTLCConfigOpNoArgs                                  = 2234
	ECTLCConfigOpNotInSpec                               = 2235
	ECTLCConfigOpIsEqual                                 = 2236
	ECTLCConfigSpecIsTrivial                             = 2237
	ECTLCCantHandleSubscript                             = 2238
	ECTLCCantHandleConjunct                              = 2239
	ECTLCCantHandleTooManyNextStateRels                  = 2240
	ECTLCConfigPropertyNotCorrectlyDefined               = 2241
	ECTLCConfigOpArityInconsistent                       = 2242
	ECTLCConfigNoStateType                               = 2243
	ECTLCCantHandleRealNumbers                           = 2244
	ECTLCNoModules                                       = 2245
	ECTLCExpectedExpression                              = 2246
	ECTLCExpectedExpressionInComputing                   = 2247
	ECTLCExpectedExpressionInComputing2                  = 2248
	ECTLCLiveEncounteredActions                          = 2249
	ECTLCLiveStatePredicateNonBool                       = 2250
	ECTLCLiveCannotEvalFormula                           = 2251
	ECTLCLiveEncounteredNonboolPredicate                 = 2252
	ECTLCLiveFormulaTautology                            = 2253
	ECSystemFingerprintOverflowError                     = 2254
	ECTLCLiveFormulaStateLevel                           = 2255
	ECTLCConfigNoSpecButProperty                         = 2257
	ECTLCLiveFormulaAndFairnessTautology                 = 2258
	ECTLCConfigNoFairnessButLiveProperty                 = 2259
	ECTLCEnabledWrongFormula                             = 2260
	ECTLCEncounteredFormulaInPredicate                   = 2261
	ECTLCVersion                                         = 2262
	ECTLCNoStatesSatisfyingInitAndConstraint             = 2256
	ECTLCModuleArgumentErrorAn                           = 2266
	ECTLCIntegerTooBig                                   = 2265
	ECTLCCheckingTemporalPropsEnd                        = 2267
	ECTLCStateGraphOutdegree                             = 2268
	ECTLCComputingInitProgress                           = 2269
	ECSystemErrorCleaningPool                            = 2270
	ECTLCModeMCDFS                                       = 2271
	ECTLCConfigPropertyActionLevel                       = 2272
	ECTLCConfigPropertyActionLevelSquareASubV            = 2273
	ECTLCConfigPropertyActionLevelAngleASubV             = 2274
	ECTLCFeatureUnsupportedLivenessSymmetry              = 2279
	ECTLCConfigUndefinedOrNoOperator                     = 2280
	ECTLCConfigSubstitutionNonConstant                   = 2281
	ECTLCTraceTooLong                                    = 2282
	ECTLCModuleOneArgumentError                          = 2283
	ECTLCFeatureLivenessConstraints                      = 2284
	ECTLCSymmetrySetTooSmall                             = 2300
	ECTLCSpecificationFeaturesTemporalQuantifier         = 2301
	ECTLCModuleValueJavaMethodOverrideMismatch           = 2400
	ECTLCEnvironmentJVMGC                                = 2401
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
	ECCheckFailedToCheck                                 = 3000
	ECCheckCouldNotReadTrace                             = 3001
	ECTLCParsingFailed                                   = 3002

	ECCheckParamExpectConfigFilename   = 3100
	ECCheckParamUsage                  = 3101
	ECCheckParamMissingTLAModule       = 3102
	ECCheckParamNeedToSpecifyConfigDir = 3103
	ECCheckParamWorkerNumberRequired   = 3104
	ECCheckParamWorkerNumberTooSmall   = 3105
	ECCheckParamWorkerNumberRequired2  = 3106
	ECCheckParamDepthRequired          = 3107
	ECCheckParamDepthRequired2         = 3108
	ECCheckParamTraceRequired          = 3109
	ECCheckParamCovreageRequired       = 3110
	ECCheckParamCovreageRequired2      = 3111
	ECCheckParamCovreageTooSmall       = 3112
	ECCheckParamUnrecognized           = 3113
	ECCheckParamTooManyInputFiles      = 3114

	ECSanyParserCheck1 = 4000
	ECSanyParserCheck2 = 4001
	ECSanyParserCheck3 = 4002

	ECCFGErrorReadingFile = 5001
	ECCFGGeneral          = 5002
	ECCFGMissingID        = 5003
	ECCFGTwiceKeyword     = 5004
	ECCFGExpectID         = 5005
	ECCFGExpectedSymbol   = 5006

	ECTLCDistributedServerRunning         = 7000
	ECTLCDistributedWorkerRegistered      = 7001
	ECTLCDistributedWorkerDeregistered    = 7002
	ECTLCDistributedWorkerStats           = 7003
	ECTLCDistributedServerNotRunning      = 7004
	ECTLCDistributedVMVersion             = 7005
	ECTLCDistributedWorkerLost            = 7006
	ECTLCDistributedExceedBlocksize       = 7007
	ECTLCDistributedServerFPSetWaiting    = 7008
	ECTLCDistributedServerFPSetRegistered = 7009
	ECTLCDistributedServerFinished        = 7010

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

const mpGeneralDebugProperty = "tlc2.output.MP.noDebug"

var mpGeneralDebug = initialBooleanProperty(mpGeneralDebugProperty)

// Message is the raw MP recorder event. Console text is formatted only after
// notification, so recorders must inspect parameters rather than printed text.
type Message struct {
	Code     int
	Severity Severity
	Params   []string
	// NullableParams preserves Java null entries for exception diagnostics.
	NullableParams []*string
	Suppressed     bool
	// FormattingOnly identifies MP.getMessage events, which reach recorders
	// without printing a diagnostic.
	FormattingOnly bool
	// Throwable preserves object-valued recorder arguments in exception printers.
	Throwable   error
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

// PrintErrorNullable preserves null parameters in recorder events. Java MP
// stops all placeholder substitution at the first null entry.
func PrintErrorNullable(code int, params ...*string) int {
	recordMessageParameters(code, SeverityError, messageParameterStrings(params), params)
	return code
}

func formatNullableMessage(code int, params []*string, messageClass ...Severity) string {
	class := SeverityNone
	if len(messageClass) > 0 {
		class = messageClass[0]
	}
	text := javaMessageTemplate(code, params, class)
	for i, param := range params {
		if param == nil {
			break
		}
		placeholder := fmt.Sprintf("%%%d%%", i+1)
		// Java substitutes sequentially, including markers introduced by an earlier
		// parameter. Stop at the first null; absent parameters retain their markers.
		for strings.Contains(text, placeholder) {
			text = strings.Replace(text, placeholder, *param, 1)
		}
	}
	return text
}

func PrintTLCBug(code int, params ...string) int {
	recordMessage(code, SeverityTLCBug, params...)
	return code
}

func PrintState(code int, params []string, state *TLCStateMut, stateNumber int) string {
	return PrintStateInfo(code, params, NewTLCStateInfoWithOrdinal(state, stateNumber), stateNumber)
}

func PrintStateInfo(code int, params []string, info *TLCStateInfo, stateNumber int) string {
	var state *TLCStateMut
	if info != nil {
		state = info.State
	}
	return recordStateMessage(code, params, state, info, stateNumber)
}

func recordMessage(code int, severity Severity, params ...string) {
	recordMessageParameters(code, severity, params, nil)
}

// GetMessage mirrors Java MP.getMessage, including its recorder notification.
func GetMessage(code int, params ...string) string {
	return getMessageParameters(code, params, nil, SeverityNone)
}

func GetMessageNullable(code int, params ...*string) string {
	return getMessageParameters(code, messageParameterStrings(params), params, SeverityNone)
}

// GetError mirrors MP.getError: notify the recorder and format an error
// without printing it or applying the warning/suppression policy.
func GetError(code int, params ...string) string {
	return getMessageParameters(code, params, nil, SeverityError)
}

func getMessageParameters(code int, params []string, nullableParams []*string, severity Severity) string {
	copied := copyMessageParameters(params)
	nullableCopied := copyNullableMessageParameters(nullableParams)
	// Java notifies the recorder before formatting, including when formatting
	// subsequently throws (for example, while substituting a null parameter).
	defaultRecorder.Record(Message{Code: code, Severity: severity,
		Params: copied, NullableParams: nullableCopied, FormattingOnly: true})
	return getRenderedMessage(code, copied, nullableCopied, severity)
}

// GetTLCBug ports MP.getTLCBug, which formats without notifying recorders.
func GetTLCBug(code int) string {
	return getRenderedMessage(code, nil, nil, SeverityTLCBug)
}

func getRenderedMessage(code int, params []string, nullable []*string, severity Severity) string {
	text := formatMPMessageBody(code, params, nullable, severity)
	Globals.Lock()
	tool := Globals.Tool
	Globals.Unlock()
	if tool {
		text = consoleMessageEnvelope(code, severity, text)
	} else {
		switch severity {
		case SeverityError:
			text = "Error: " + text
		case SeverityTLCBug:
			text = "TLC Bug: " + text
		}
	}
	DebugPrintMessage("Leaving getMessage()")
	return text
}

func recordMessageParameters(code int, severity Severity, params []string, nullableParams []*string) {
	// Source warning elevation happens before the printer's recorder event.
	if severity == SeverityWarning {
		_, asError, _ := messageControlFor(code)
		warningToError, _ := tlcLookupSystemProperty("tlc2.output.MP.warning2error")
		if javaBooleanProperty(warningToError) || asError {
			if nullableParams != nil {
				failure := newTLCErrorCodeNullable(code, nullableParams...)
				failure.Runtime = true
				panic(failure)
			}
			panic(NewTLCRuntimeException(code, params...))
		}
	}
	copied := copyMessageParameters(params)
	nullableCopied := copyNullableMessageParameters(nullableParams)
	suppressed, _, warn := messageControlFor(code)
	defaultRecorder.Record(Message{
		Code: code, Severity: severity, Params: copied, NullableParams: nullableCopied,
		Suppressed: severity != SeverityError && (suppressed || severity == SeverityWarning && !warn),
	})
	debugMessagePrinterEnter(code, severity)
	// Recorder callbacks precede Java's visibility checks and formatting. They
	// can change the controls or throw, and formatting itself can throw.
	suppressed, _, warn = messageControlFor(code)
	if severity == SeverityWarning {
		if !warn {
			debugMessagePrinterLeave(severity)
			return
		}
	} else if severity != SeverityError && suppressed {
		debugMessagePrinterLeave(severity)
		return
	}
	text := formatMPMessageBody(code, copied, nullableCopied, severity)
	// Enabled warnings enter the history even when individually suppressed.
	printConsoleMessage(code, severity, text, severity == SeverityError || !suppressed)
	debugMessagePrinterLeave(severity)
}

func recordStateMessage(code int, params []string, state *TLCStateMut, info *TLCStateInfo, stateNumber int) string {
	suppressed, _, _ := messageControlFor(code)
	copied := copyMessageParameters(params)
	defaultRecorder.Record(Message{
		Code: code, Severity: SeverityState, Params: copied, Suppressed: suppressed,
		State: state, StateInfo: info, StateNumber: stateNumber,
	})
	// Unlike ordinary messages, Java formats suppressed states and returns the
	// formatted message. Its recorder notification still precedes formatting.
	DebugPrintMessage("entering printState(String[])")
	text := formatMPMessageBody(code, copied, nil, SeverityState)
	suppressed, _, _ = messageControlFor(code)
	message := printConsoleMessage(code, SeverityState, text, !suppressed)
	DebugPrintMessage("leaving printState(String[])")
	return message
}

func formatMessage(code int, params []string, messageClass ...Severity) string {
	nullable := make([]*string, len(params))
	for i := range params {
		nullable[i] = &params[i]
	}
	return formatNullableMessage(code, nullable, messageClass...)
}

func messageNow() string {
	if javaBooleanProperty(tlcGetSystemProperty("tlc2.output.MP.noTimestamps", "")) {
		return "NOW"
	}
	return time.Now().Format("2006-01-02 15:04:05")
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
	case ECTLCLiveFormulaTautology:
		return ExitStatusFailureLivenessEval
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
		ECTLCConfigValueNotAssignedToConstantParam,
		ECTLCConfigRHSIDAppearedAfterLHSID,
		ECTLCConfigWrongSubstitution,
		ECTLCConfigWrongSubstitutionNumberOfArgs,
		ECTLCConfigUndefinedOrNoOperator,
		ECTLCConfigSubstitutionNonConstant,
		ECTLCConfigIDDoesNotAppearInSpec,
		ECTLCConfigNotBothSpecAndInit,
		ECTLCConfigIDRequiresNoArg,
		ECTLCConfigSpecifiedNotDefined,
		ECTLCConfigIDHasValue,
		ECTLCConfigMissingInit,
		ECTLCConfigMissingNext,
		ECTLCConfigIDMustNotBeConstant,
		ECTLCConfigOpNoArgs,
		ECTLCConfigOpNotInSpec,
		ECTLCConfigOpIsEqual,
		ECTLCConfigSpecIsTrivial,
		ECTLCCantHandleSubscript,
		ECTLCCantHandleConjunct,
		ECTLCCantHandleTooManyNextStateRels,
		ECTLCConfigPropertyNotCorrectlyDefined,
		ECTLCConfigPropertyActionLevel,
		ECTLCConfigPropertyActionLevelSquareASubV,
		ECTLCConfigPropertyActionLevelAngleASubV,
		ECTLCConfigOpArityInconsistent,
		ECTLCConfigNoStateType,
		ECTLCCantHandleRealNumbers,
		ECTLCNoModules:
		return ExitStatusErrorConfigParse
	case ECTLCParsingFailed2,
		ECTLCParsingFailed:
		return ExitStatusErrorSpecParse
	default:
		return ExitStatusError
	}
}

// MP retains an unsubstituted placeholder when its parameter is absent.
func configMessageParam(params []string, index int) string {
	if index >= len(params) {
		return fmt.Sprintf("%%%d%%", index+1)
	}
	return params[index]
}
