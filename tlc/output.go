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

type Message struct {
	Code        int
	Severity    Severity
	Params      []string
	Text        string
	Suppressed  bool
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

func PrintTLCBug(code int, params ...string) int {
	recordMessage(code, SeverityTLCBug, params...)
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
	visible := !suppressed
	if severity == SeverityWarning {
		if asError {
			panic(NewEvalException(code, params...))
		}
		visible = warn && !suppressed
	} else if severity == SeverityError {
		visible = true
	}
	copied := append([]string(nil), params...)
	defaultRecorder.Record(Message{
		Code:       code,
		Severity:   severity,
		Params:     copied,
		Text:       formatMessage(code, copied),
		Suppressed: !visible,
	})
}

func recordStateMessage(code int, params []string, text string, state *TLCStateMut, info *TLCStateInfo, stateNumber int) {
	suppressed, _, _ := messageControlFor(code)
	copied := append([]string(nil), params...)
	defaultRecorder.Record(Message{
		Code:        code,
		Severity:    SeverityState,
		Params:      copied,
		Text:        text,
		Suppressed:  suppressed,
		State:       state,
		StateInfo:   info,
		StateNumber: stateNumber,
	})
}

func formatMessage(code int, params []string) string {
	switch code {
	case ECSystemStackOverflow:
		return "This was a Java StackOverflowError. It was probably the result\n" +
			"of an incorrect recursive function definition that caused TLC to enter\n" +
			"an infinite loop when trying to compute the function or its application\n" +
			"to an element in its putative domain."
	case ECSystemOutOfMemory:
		return "Java ran out of memory.  Running Java with a larger memory allocation\n" +
			"pool (heap) may fix this.  But it won't help if some state has an enormous\n" +
			"number of successor states, or if TLC must compute the value of a huge set."
	case ECSystemOutOfMemoryLiveness:
		return "Java ran out of memory during liveness checking.  Running Java with a larger memory\n" +
			"allocation pool (heap) may fix this.  But it won't help if paths in the liveness graph\n" +
			"have an enormous number of states."
	case ECSystemOutOfMemoryTooManyInit:
		return "Out Of Memory. There are probably too many initial states."
	case ECSystemFingerprintOverflowError:
		return "The requested fingerprint set size is too large, which could result in precision or overflow errors. To resolve this, either increase the fingerprint size to the next power of two (ensuring safe bit-shifting) or adjust the fpbits parameter to a higher value."
	case ECWrongCommandlineParamsTLC:
		if len(params) >= 1 {
			return fmt.Sprintf("%s\nUsage: java tlc2.TLC [-help] [-option] inputfile", params[0])
		}
	case ECWrongCommandlineParamsSimulator:
		if len(params) >= 1 {
			return fmt.Sprintf("%s\nUsage: java tlc2.Simulator [-help] [-option] inputfile", params[0])
		}
	case ECTLCVersion:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC2 %s", params[0])
		}
	case ECTLCPPFormatingValue:
		if len(params) >= 2 {
			return fmt.Sprintf("error while formating %s\n%s", params[0], params[1])
		}
	case ECTLCPPParsingValue:
		if len(params) >= 2 {
			return fmt.Sprintf("error while parsing %s\n%s", params[0], params[1])
		}
	case ECTLCStarting:
		return fmt.Sprintf("Starting... (%s)", messageNow())
	case ECTLCFinished:
		if len(params) >= 1 {
			return fmt.Sprintf("Finished in %s at (%s)", params[0], messageNow())
		}
	case ECTLCModeMC:
		return formatModeMCMessage(params, false)
	case ECTLCModeMCDFS:
		return formatModeMCMessage(params, true)
	case ECTLCModeSimu:
		return formatModeSimulationMessage(params)
	case ECTLCComputingInit:
		return "Computing initial states..."
	case ECTLCComputingInitProgress:
		if len(params) >= 1 {
			return fmt.Sprintf("Computed %s initial states...", params[0])
		}
	case ECTLCInitGenerated1:
		if len(params) >= 2 {
			return fmt.Sprintf("Finished computing initial states: %s distinct state%s generated at %s.", params[0], params[1], messageNow())
		}
	case ECTLCInitGenerated2:
		if len(params) >= 3 {
			return fmt.Sprintf("Finished computing initial states: %s state%s generated, with %s of them distinct at %s.", params[0], params[1], params[2], messageNow())
		}
	case ECTLCInitGenerated3:
		if len(params) >= 2 {
			return fmt.Sprintf("Finished computing initial states: %s states generated.\nBecause TLC recovers from a previous checkpoint, only %s of them require further exploration at %s.", params[0], params[1], messageNow())
		}
	case ECTLCInitGenerated4:
		if len(params) >= 2 {
			return fmt.Sprintf("Finished computing initial states: %s states generated, with %s of them distinct.", params[0], params[1])
		}
	case ECTLCSuccess:
		if len(params) == 1 {
			return fmt.Sprintf("Model checking completed. No error has been found.\n  Estimates of the probability that TLC did not check all reachable states\n  because two distinct states had the same fingerprint:\n  calculated (optimistic):  %s", params[0])
		}
		if len(params) >= 2 {
			return fmt.Sprintf("Model checking completed. No error has been found.\n  Estimates of the probability that TLC did not check all reachable states\n  because two distinct states had the same fingerprint:\n  calculated (optimistic):  %s\n  based on the actual fingerprints:  %s", params[0], params[1])
		}
	case ECTLCSearchDepth:
		if len(params) >= 1 {
			return fmt.Sprintf("The depth of the complete state graph search is %s.", params[0])
		}
	case ECTLCStateGraphOutdegree:
		if len(params) >= 4 {
			return fmt.Sprintf("The average outdegree of the complete state graph is %s (minimum is %s, the maximum %s and the 95th percentile is %s).", params[1], params[0], params[3], params[2])
		}
	case ECTLCCheckpointStart:
		if len(params) >= 1 {
			return fmt.Sprintf("Checkpointing of run %s", params[0])
		}
	case ECTLCCheckpointEnd:
		return fmt.Sprintf("Checkpointing completed at (%s)", messageNow())
	case ECTLCCheckpointRecoverStart:
		if len(params) >= 1 {
			return fmt.Sprintf("Starting recovery from checkpoint %s", params[0])
		}
	case ECTLCCheckpointRecoverEnd:
		if len(params) >= 2 {
			return fmt.Sprintf("Recovery completed. %s states examined. %s states on queue.", params[0], params[1])
		}
	case ECTLCCheckpointRecoverEndDFID:
		if len(params) >= 1 {
			return fmt.Sprintf("Recovery completed. %s states examined.", params[0])
		}
	case ECSystemCheckpointRecoveryCorrupt:
		if len(params) >= 1 {
			return "TLC encountered the following error while restarting from a checkpoint;\n the checkpoint file is probably corrupted.\n" + params[0]
		}
	case ECSystemErrorReadingStates:
		if len(params) >= 2 {
			return fmt.Sprintf("TLC encountered the following error reading the %s of unexplored states:\n%s", params[0], params[1])
		}
	case ECSystemErrorWritingStates:
		if len(params) >= 2 {
			return fmt.Sprintf("TLC encountered the following error writing the %s of unexplored states:\n%s", params[0], params[1])
		}
	case ECTLCStats:
		if len(params) >= 3 {
			return fmt.Sprintf("%s states generated, %s distinct states found, %s states left on queue.", params[0], params[1], params[2])
		}
	case ECTLCStatsDFID:
		if len(params) >= 2 {
			return fmt.Sprintf("%s states generated, %s distinct states found.", params[0], params[1])
		}
	case ECTLCStatsSimu:
		if len(params) >= 3 {
			return fmt.Sprintf("The number of states generated: %s\nSimulation using seed %s and aril %s", params[0], params[1], params[2])
		}
	case ECTLCProgressStats:
		if len(params) == 4 {
			return fmt.Sprintf("Progress(%s) at %s: %s states generated, %s distinct states found, %s states left on queue.", params[0], messageNow(), params[1], params[2], params[3])
		}
		if len(params) >= 6 {
			return fmt.Sprintf("Progress(%s) at %s: %s states generated (%s s/min), %s distinct states found (%s ds/min), %s states left on queue.", params[0], messageNow(), params[1], params[4], params[2], params[5], params[3])
		}
	case ECTLCProgressStartStatsDFID:
		if len(params) >= 3 {
			return fmt.Sprintf("Starting level %s: %s states generated, %s distinct states found.", params[0], params[1], params[2])
		}
	case ECTLCProgressStatsDFID:
		if len(params) >= 2 {
			return fmt.Sprintf("Progress: %s states generated, %s distinct states found.", params[0], params[1])
		}
	case ECTLCProgressSimu:
		if len(params) >= 5 {
			return fmt.Sprintf("Progress: %s states checked, %s traces generated (trace length: mean=%s, var(x)=%s, sd=%s)", params[0], params[1], params[2], params[3], params[4])
		}
	case ECTLCCoverageStart:
		return fmt.Sprintf("The coverage statistics at %s (see https://explain.tlapl.us/module-coverage-statistics for how to interpret the following statistics).", messageNow())
	case ECTLCCoverageValue:
		if len(params) >= 2 {
			return fmt.Sprintf("  %s: %s", params[0], params[1])
		}
	case ECTLCCoverageValueCost:
		if len(params) >= 3 {
			return fmt.Sprintf("  %s: %s:%s", params[0], params[1], params[2])
		}
	case ECTLCCoverageVar:
		if len(params) >= 3 {
			return fmt.Sprintf("<%s %s>: %s", params[0], params[1], params[2])
		}
	case ECTLCCoverageInit, ECTLCCoverageNext, ECTLCCoverageConstraint:
		if len(params) >= 3 {
			return fmt.Sprintf("%s: %s:%s", params[0], params[1], params[2])
		}
	case ECTLCCoverageProperty:
		if len(params) >= 1 {
			return params[0]
		}
	case ECTLCCoverageMismatch:
		if len(params) >= 2 {
			return fmt.Sprintf("CostModel lookup failed for expression <%s>. Reporting costs into <%s> instead (Safety and Liveness checking is unaffected. Please report a bug.)", params[0], params[1])
		}
	case ECTLCCoverageEnd:
		return "End of statistics."
	case ECTLCCoverageEndOverhead:
		return "End of statistics (please note that for performance reasons large models\nare best checked with coverage and cost statistics disabled)."
	case ECGeneral:
		return strings.Join(params, "")
	case ECTLCFeatureUnsupported:
		return strings.Join(params, "")
	case ECTLCFPCompleted:
		if len(params) >= 1 {
			return fmt.Sprintf("%s, work completed. Thank you!", params[0])
		}
	case ECTLCFPValueAlreadyOnDisk:
		if len(params) >= 1 {
			return fmt.Sprintf("DiskFPSet.mergeNewEntries: %s is already on disk.\n", params[0])
		}
	case ECTLCDistributedServerRunning:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC server at %s is ready (%s)", params[0], messageNow())
		}
	case ECTLCDistributedWorkerRegistered:
		if len(params) >= 1 {
			return fmt.Sprintf("Registration for worker at %s completed (%s)", params[0], messageNow())
		}
	case ECTLCDistributedWorkerDeregistered:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC worker %s disconnected (%s)", params[0], messageNow())
		}
	case ECTLCDistributedWorkerStats:
		if len(params) >= 4 {
			return fmt.Sprintf("Worker: %s Sent: %s Rcvd: %s CacheRatio: %s (%s)", params[0], params[1], params[2], params[3], messageNow())
		}
	case ECTLCDistributedServerNotRunning:
		if len(params) >= 1 {
			return fmt.Sprintf("TLCServer is gone due to %s, exiting worker... (%s)", params[0], messageNow())
		}
	case ECTLCDistributedServerFinished:
		return fmt.Sprintf("TLCServer has finished, exiting worker... (%s)", messageNow())
	case ECTLCDistributedVMVersion:
		return fmt.Sprintf("VM does not allow to get the UnicastRef port.\nWorker will be identified with port 0 in output (%s)", messageNow())
	case ECTLCDistributedWorkerLost:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC worker connection lost %s (%s)", params[0], messageNow())
		}
	case ECTLCDistributedExceedBlocksize:
		if len(params) >= 1 {
			return fmt.Sprintf("Trying to limit max block size (to recover from transport failure): %s (%s)", params[0], messageNow())
		}
	case ECTLCDistributedServerFPSetRegistered:
		if len(params) >= 2 {
			return fmt.Sprintf("%s out of %s FPSet server(s) registered (%s)", params[0], params[1], messageNow())
		}
	case ECTLCDistributedServerFPSetWaiting:
		if len(params) >= 1 {
			return fmt.Sprintf("Waiting for %s FPSet server(s) to register (%s)", params[0], messageNow())
		}
	case ECTLCConfigValueNotAssignedToConstantParam:
		if len(params) >= 1 {
			return fmt.Sprintf("The constant parameter %s is not assigned a value by the configuration file.", params[0])
		}
	case ECTLCConfigRHSIDAppearedAfterLHSID:
		if len(params) >= 1 {
			return fmt.Sprintf("In the configuration file, the identifier %s appears\non the right-hand side of a <- after already appearing on the\nleft-hand side of one.", params[0])
		}
	case ECTLCConfigWrongSubstitution:
		if len(params) >= 2 {
			return fmt.Sprintf("The configuration file substitutes for %s with the undefined identifier %s.", params[0], params[1])
		}
	case ECTLCConfigUndefinedOrNoOperator:
		if len(params) >= 2 {
			return fmt.Sprintf("In evaluation, the identifier %s is either undefined or not an operator.\n%s", params[0], params[1])
		}
	case ECTLCConfigSubstitutionNonConstant:
		if len(params) >= 3 {
			return fmt.Sprintf("The configuration file substitutes constant %s with non-constant %s%s", params[0], params[1], params[2])
		}
	case ECTLCConfigWrongSubstitutionNumberOfArgs:
		if len(params) >= 2 {
			return fmt.Sprintf("The configuration file substitutes for %s with %s of different number of arguments.", params[0], params[1])
		}
	case ECTLCConfigIDDoesNotAppearInSpec:
		if len(params) >= 1 {
			return fmt.Sprintf("In the configuration file, the identifier %s does not appear in the specification.", params[0])
		}
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
	case ECTLCSymmetrySetTooSmall:
		if len(params) >= 4 {
			return fmt.Sprintf("The set%s %s %s been defined to be a symmetry set but contain%s less than two elements.", params[0], params[1], params[2], params[3])
		}
	case ECTLCInitialState:
		if len(params) >= 2 {
			return fmt.Sprintf("%s\nWhile working on the initial state:\n%s", params[0], params[1])
		}
	case ECTLCNestedExpression:
		if len(params) >= 1 {
			return fmt.Sprintf("The error occurred when TLC was evaluating the nested\nexpressions at the following positions:\n%s", params[0])
		}
	case ECTLCAssumptionFalse:
		if len(params) >= 1 {
			return fmt.Sprintf("Assumption %s is false.", params[0])
		}
	case ECTLCPostconditionFalse:
		if len(params) >= 2 {
			return fmt.Sprintf("Postcondition %s at %s is false.", params[0], params[1])
		}
	case ECTLCPossibleUnwitnessed:
		if len(params) >= 2 {
			return fmt.Sprintf("The _POSSIBLE predicate %s at %s was never witnessed.\nOthers may be unwitnessed too; only witnessed predicates appear with a\nnon-zero count in the record printed above.\nSee https://explain.tlapl.us/possible-conditions for additional details.", params[0], params[1])
		}
	case ECTLCPostconditionEvaluationError:
		if len(params) >= 3 {
			return fmt.Sprintf("Evaluating postcondition %s at %s failed.\n%s", params[0], params[1], params[2])
		}
	case ECTLCSanyStart:
		return "Starting SANY..."
	case ECTLCSanyEnd:
		return "SANY finished."
	case ECTLCAssumptionEvaluationError:
		if len(params) >= 2 {
			return fmt.Sprintf("Evaluating assumption %s failed.\n%s", params[0], params[1])
		}
	case ECTLCStateNotCompletelySpecifiedInitial:
		if len(params) >= 1 {
			return fmt.Sprintf("State is not completely specified by the initial predicate:\n%s", params[0])
		}
	case ECTLCInvariantViolatedInitial:
		if len(params) >= 2 {
			return fmt.Sprintf("Invariant %s is violated by the initial state:\n%s", params[0], params[1])
		}
	case ECTLCPropertyViolatedInitial:
		if len(params) >= 2 {
			return fmt.Sprintf("Property %s is violated by the initial state:\n%s", params[0], params[1])
		}
	case ECTLCStateNotCompletelySpecifiedNext:
		if len(params) >= 3 {
			return fmt.Sprintf("Successor state is not completely specified by action %s of the next-state relation. The following variable%s not defined: %s.\n", params[0], params[1], params[2])
		}
		if len(params) >= 2 {
			return fmt.Sprintf("Successor state is not completely specified by the next-state action. The following variable%s not defined: %s.\n", params[0], params[1])
		}
		return "Successor state is not completely specified by the next-state action.\n"
	case ECTLCStateNotCompletelySpecifiedLive:
		if len(params) >= 1 {
			return fmt.Sprintf("The action formula A appearing in a WF_v(A) or SF_v(A) operator does not specify the primed value of the variable %s occurring in the state formula v.\n", params[0])
		}
	case ECTLCInvariantViolatedBehavior:
		if len(params) >= 1 {
			return fmt.Sprintf("Invariant %s is violated.", params[0])
		}
	case ECTLCInvariantEvaluationFailed:
		if len(params) >= 2 {
			return fmt.Sprintf("Evaluating invariant %s failed.\n%s", params[0], params[1])
		}
		if len(params) >= 1 {
			return fmt.Sprintf("Evaluating invariant %s failed.", params[0])
		}
	case ECTLCInvariantViolatedLevel:
		if len(params) >= 1 {
			msg := fmt.Sprintf("The invariant %s is not a state predicate (one with no primes or temporal operators).", params[0])
			if len(params) > 1 {
				msg += "\nNote that a bug can cause TLC to incorrectly report this error.\nIf you believe your TLA+ or PlusCal specification to be correct,\nplease check if this bug described in LevelNode.java starting at line 590ff affects you."
			}
			return msg
		}
	case ECTLCInvariantConstantLevel:
		if len(params) >= 2 {
			return fmt.Sprintf("The invariant %s is a constant-level formula (i.e., it contains no variables, primes, or temporal operators) and evaluates to %s. To assert constant-level formulas in your spec, use ASSUME ConstInv. If you optionally want to give the assumption a name, write ASSUME YourAssumption == ConstInv instead. See https://explain.tlapl.us/assumptions-and-invariants for additional details.", params[0], params[1])
		}
	case ECTLCActionPropertyViolatedBehavior:
		if len(params) >= 1 {
			return fmt.Sprintf("Action property %s is violated.", params[0])
		}
	case ECTLCActionPropertyEvaluationFailed:
		if len(params) >= 2 {
			return fmt.Sprintf("Evaluating action property %s failed.\n%s", params[0], params[1])
		}
		if len(params) >= 1 {
			return fmt.Sprintf("Evaluating action property %s failed.", params[0])
		}
	case ECTLCDeadlockReached:
		return "Deadlock reached."
	case ECTLCStatesAndNoNextAction:
		return "No next state actions defined to generate successor states from."
	case ECTLCFailedToRecoverNext:
		return "Failed to recover the next state from its fingerprint."
	case ECTLCFailedToRecoverInit:
		return "Failed to recover the initial state from its fingerprint."
	case ECTLCBug:
		if len(params) >= 1 {
			return fmt.Sprintf("This is probably a TLC bug(%s).", params[0])
		}
		return "This is probably a TLC bug(%1%)."
	case ECTLCFingerprintException:
		if len(params) >= 2 {
			return fmt.Sprintf("TLC was unable to fingerprint.\n\nFingerprint Stack Trace:\n%s\nReason:\n%s", params[0], params[1])
		}
		return "TLC was unable to fingerprint.\n\nFingerprint Stack Trace:\n\nReason:\n"
	case ECTLCNoStatesSatisfyingInit:
		return "There is no state satisfying the initial state predicate."
	case ECTLCNoStatesSatisfyingInitAndConstraint:
		return "There is no state satisfying the initial state predicate and the state-constraint(s)."
	case ECTLCBehaviorUpToThisPoint:
		return "The behavior up to this point is:"
	case ECTLCErrorState:
		return "The error state is:\n"
	case ECTLCReporterDied:
		return "Progress report thread died."
	case ECTLCAAAAAAA:
		return "AAAAAA"
	case ECTLCRegistryInitError:
		if len(params) >= 1 {
			return fmt.Sprintf("TLA+ Registry initialization error. The name %s is already in use.", params[0])
		}
	case ECCheckFailedToCheck:
		return "TLC failed in checking traces."
	case ECCheckParamUsage:
		return "Usage: java tlc2.tool.CheckImplFile [-option] inputfile"
	case ECCheckParamMissingTLAModule:
		return "Missing input TLA+ module."
	case ECCheckParamExpectConfigFilename:
		return "Expect a file name for -config option."
	case ECCheckParamNeedToSpecifyConfigDir:
		return "need to specify the metadata directory for recovery."
	case ECCheckParamWorkerNumberRequired:
		if len(params) >= 1 {
			return fmt.Sprintf("Worker number required. But encountered %s", params[0])
		}
	case ECCheckParamWorkerNumberTooSmall:
		return "At least one worker is required."
	case ECCheckParamWorkerNumberRequired2:
		return "Expect an integer for -workers option."
	case ECCheckParamDepthRequired:
		if len(params) >= 1 {
			return fmt.Sprintf("Depth must be an integer. But encountered %s", params[0])
		}
	case ECCheckParamDepthRequired2:
		return "Expect an integer for -depth option."
	case ECCheckParamTraceRequired:
		return "Expect a filename for -trace option."
	case ECCheckParamCovreageRequired:
		if len(params) >= 1 {
			return fmt.Sprintf("An integer for coverage report interval required. But encountered %s", params[0])
		}
	case ECCheckParamCovreageRequired2:
		return "Coverage report interval required."
	case ECCheckParamCovreageTooSmall:
		return "Expect a nonnegative integer for -coverage option."
	case ECCheckParamUnrecognized:
		if len(params) == 1 {
			return fmt.Sprintf("Unrecognized option: %s", params[0])
		}
		if len(params) >= 2 {
			return fmt.Sprintf("Unrecognized option in file %s: %s", params[1], params[0])
		}
	case ECCheckParamTooManyInputFiles:
		if len(params) >= 2 {
			return fmt.Sprintf("More than one input files: %s and %s", params[0], params[1])
		}
	case ECCheckCouldNotReadTrace:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC could not read in the trace. %s", params[0])
		}
	case ECTLCParsingFailed:
		return "Parsing or semantic analysis failed."
	case ECTLCParsingFailed2:
		if len(params) >= 1 {
			return fmt.Sprintf("Parsing or semantic analysis failed.%s", params[0])
		}
	case ECTLCStatePrint1:
		if len(params) >= 2 {
			return fmt.Sprintf("%s:\n%s", params[0], params[1])
		}
	case ECTLCParameterMustBePostfix:
		return "Parameter must be a postfix operator"
	case ECTLCCouldNotDetermineSubscript:
		return "TLC could not determine if the subscript of the next-state relation contains\nall state variables. Proceed with fingers crossed."
	case ECTLCSubscriptContainNoStateVar:
		if len(params) >= 1 {
			return fmt.Sprintf("The subscript of the next-state relation specified by the specification\ndoes not seem to contain the state variable %s", params[0])
		}
	case ECTLCWrongTupleFieldName:
		if len(params) >= 1 {
			return fmt.Sprintf("Tuple field name %s is not an integer.", params[0])
		}
	case ECTLCWrongRecordFieldName:
		if len(params) >= 1 {
			return fmt.Sprintf("Record field name %s is not a string.", params[0])
		}
	case ECTLCUnchangedVariableChanged:
		if len(params) >= 2 {
			return fmt.Sprintf("The variable %s was changed while it is specified as UNCHANGED at\n%s", params[0], params[1])
		}
	case ECTLCExceptAppliedToUnknownField:
		if len(params) >= 1 {
			return fmt.Sprintf("The EXCEPT was applied to non-existing fields of the value at\n%s", params[0])
		}
	case ECTLCConfigIDMustNotBeConstant:
		if len(params) >= 2 {
			return fmt.Sprintf("The %s %s cannot be a constant.", params[0], params[1])
		}
	case ECTLCConfigOpNoArgs:
		if len(params) >= 1 {
			return fmt.Sprintf("The operator %s cannot take any argument.", params[0])
		}
	case ECTLCConfigOpNotInSpec:
		if len(params) >= 1 {
			return fmt.Sprintf("The operator %s is not defined in the spec.", params[0])
		}
	case ECTLCConfigOpIsEqual:
		if len(params) >= 3 {
			return fmt.Sprintf("The operator %s, which equals %s,\ncannot be used as a %s", params[0], params[1], params[2])
		}
	case ECTLCConfigSpecIsTrivial:
		if len(params) >= 1 {
			return fmt.Sprintf("The spec is trivially false because %s is false.", params[0])
		}
	case ECSanyParserCheck1:
		return "TLA+ Parser sanity check."
	case ECSanyParserCheck2:
		return "TLA+ Parser check: Assertion error in epa()."
	case ECSanyParserCheck3:
		return "TLA+ Parser check: Assertion error in SBracketCases()."
	case ECTLCArgumentMismatch:
		if len(params) >= 1 {
			return fmt.Sprintf("Argument mismatch in operator application.%s", params[0])
		}
		return "Argument mismatch in operator application.%1"
	case ECTLCCantHandleSubscript:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC cannot handle subscript %s", params[0])
		}
	case ECTLCCantHandleConjunct:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC cannot handle this conjunct of the spec:\n%s", params[0])
		}
	case ECTLCCantHandleTooManyNextStateRels:
		return "The specification contains more than one conjunct of the form [][Next]_v,\nbut TLC can handle only specifications with one next-state relation."
	case ECTLCConfigPropertyActionLevelSquareASubV:
		if len(params) >= 2 {
			return fmt.Sprintf("The formula %s at %s is an action-level formula (i.e., it contains no temporal operators). Only temporal-level (or state-level) formulas are allowed under PROPERTY or PROPERTIES. To check that action %s holds in *every* step of the behavior, define a temporal property by applying the \"always\" temporal operator ([]) to the formula at %s (compare page 90 of Specifying Systems at https://lamport.azurewebsites.net/tla/book.html).", params[0], params[1], params[0], params[1])
		}
	case ECTLCConfigPropertyActionLevelAngleASubV:
		if len(params) >= 2 {
			return fmt.Sprintf("The formula %s at %s is an action-level formula (i.e., it contains no temporal operators). Only temporal-level (or state-level) formulas are allowed under PROPERTY or PROPERTIES. To check that infinitely many %s steps occur, define a temporal property by applying the \"always eventually\" temporal operator combination ([]<>) to the formula at %s (compare page 91 of Specifying Systems at https://lamport.azurewebsites.net/tla/book.html).", params[0], params[1], params[0], params[1])
		}
	case ECTLCConfigPropertyActionLevel:
		if len(params) >= 2 {
			return fmt.Sprintf("The formula %s at %s is an action-level formula (i.e., it contains no temporal operators). Only temporal-level (or state-level) formulas are allowed under PROPERTY or PROPERTIES. To check that action %s holds in *every* step of the behavior, define a temporal property: Let A be the formula at %s; rewrite A as [][A]_v, where v is a state function (typically the tuple of variables) (compare page 90 of Specifying Systems at https://lamport.azurewebsites.net/tla/book.html).", params[0], params[1], params[0], params[1])
		}
	case ECTLCConfigPropertyNotCorrectlyDefined:
		if len(params) >= 1 {
			return fmt.Sprintf("The property %s is not correctly defined.", params[0])
		}
	case ECTLCConfigOpArityInconsistent:
		if len(params) >= 1 {
			return fmt.Sprintf("The arity of the operator %s is inconsistent in the configuration file.", params[0])
		}
	case ECTLCConfigNoStateType:
		return "The configuration file did not specify types for state variables."
	case ECTLCCantHandleRealNumbers:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC can't handle real numbers.\n%s", params[0])
		}
	case ECTLCIntegerTooBig:
		if len(params) >= 1 {
			return fmt.Sprintf("TLC can't handle a number this big.\n%s", params[0])
		}
	case ECTLCNoModules:
		if len(params) >= 1 {
			return fmt.Sprintf("In the configuration file, the module name %s is not a module in the specification.", params[0])
		}
	case ECTLCSpecificationFeaturesTemporalQuantifier:
		return "TLC does not support temporal existential, nor universal, quantification over state variables."
	case ECTLCLiveFormulaStateLevel:
		if len(params) >= 1 {
			return fmt.Sprintf("The formula %s is a state-level formula, but it was used as a PROPERTY (or PROPERTIES), where a temporal formula is typically expected. State-level formulas used as PROPERTY or PROPERTIES are only checked in the initial state. To verify that the formula %s holds in all states of every behavior, use INVARIANT %s instead. Alternatively, applying the \"always\" temporal operator ([]) to the state-level formula %s changes it into a temporal formula, asserting that %s holds in all states of every behavior. See https://explain.tlapl.us/invariants-and-properties for additional details.", params[0], params[0], params[0], params[0], params[0])
		}
	case ECTLCLiveFormulaAndFairnessTautology:
		if len(params) >= 4 {
			return fmt.Sprintf("Implied-temporal checking is a tautology.\nThe temporal formula %s at %s equals\nthe fairness constraints of the behavior specification %s at %s.\nTherefore, checking whether %s => %s using TLC will never result in a violation of %s.", params[0], params[1], params[2], params[3], params[2], params[0], params[0])
		}
	case ECTLCValueAssertFailed:
		if len(params) >= 1 {
			return fmt.Sprintf("The first argument of Assert evaluated to FALSE; the second argument was:\n%s", params[0])
		}
	case ECTLCFPNotInSet:
		return "The fingerprint is not in set."
	case ECTLCModuleTLCGetUndefined:
		if len(params) >= 1 {
			return fmt.Sprintf("TLCGet(%s) was undefined.", params[0])
		}
	case ECTLCModuleOverflow:
		if len(params) >= 1 {
			return fmt.Sprintf("Overflow when computing %s", params[0])
		}
	case ECTLCModuleOneArgumentError:
		if len(params) >= 3 {
			return fmt.Sprintf("The argument of %s should be a %s, but instead it is:\n%s", params[0], params[1], params[2])
		}
	case ECTLCModuleValueJavaMethodOverride:
		if len(params) >= 2 {
			return fmt.Sprintf("Attempted to apply the operator overridden by the Java method\n%s,\nbut it produced the following error:\n%s", params[0], params[1])
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
	case ECTLCModuleApplyingToWrongValue:
		if len(params) >= 3 {
			return fmt.Sprintf("Applying %s to the following value,\nwhich is not %s:\n%s", params[0], params[1], params[2])
		}
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
	case ECTLCModuleBagUnion1:
		if len(params) >= 1 {
			return fmt.Sprintf("Attempted to apply BagUnion to the following set, whose\nelement is not a bag:\n%s", params[0])
		}
	case ECTLCModuleTransitiveClosure:
		if len(params) >= 1 {
			return fmt.Sprintf("Attempted to apply TransitiveClosure to a set containing\nthe following value:\n%s", params[0])
		}
	case ECTLCModuleComputingCardinality:
		if len(params) >= 1 {
			return fmt.Sprintf("Attempted to compute cardinality of the value\n%s", params[0])
		}
	case ECTLCModuleEvaluating:
		if len(params) >= 3 {
			return fmt.Sprintf("Evaluating an expression of the form %s when s is not a %s:\n%s", params[0], params[1], params[2])
		}
	case ECTLCLiveBEGraphFailedToConstruct:
		return "BEGraph.GetPath: Failed to construct a path."
	case ECTLCLiveImplied:
		if len(params) >= 1 {
			return fmt.Sprintf("Implied-temporal checking--satisfiability problem has %s branches.", params[0])
		}
	case ECTLCLiveImpliedDebug:
		if len(params) >= 5 {
			return fmt.Sprintf("Implied-temporal checking--branch of satisfiable problem has %s possible error model(s), %s promise(s), %s state check(s), %s action check(s), and a tableau with %s node(s).", params[0], params[1], params[2], params[3], params[4])
		}
	case ECTLCLiveCannotHandleFormula:
		if len(params) > 1 {
			return fmt.Sprintf("TLC cannot handle the temporal formula %s:\n%s", params[0], params[1])
		}
		if len(params) >= 1 {
			return fmt.Sprintf("TLC cannot handle the temporal formula %s", params[0])
		}
	case ECTLCLiveWrongFormulaFormat:
		return "Temporal formulas containing actions must be of forms <>[]A or []<>A."
	case ECTLCLiveEncounteredActions:
		return "TLC encountered actions when computing closure."
	case ECTLCLiveStatePredicateNonBool:
		return "A state predicate was evaluated to a non-boolean value."
	case ECTLCLiveCannotEvalFormula:
		if len(params) >= 1 {
			return fmt.Sprintf("Can not evaluate a temporal formula %sF.", params[0])
		}
		return "Can not evaluate a temporal formula %1%F."
	case ECTLCLiveEncounteredNonboolPredicate:
		return "Encountered an action predicate that's not a boolean."
	case ECTLCExpectedValue:
		if len(params) >= 2 {
			return fmt.Sprintf("TLC expected a %s value, but did not find one. %s", params[0], params[1])
		}
	case ECTLCExpectedExpression:
		if len(params) >= 2 {
			return fmt.Sprintf("TLC expected a %s expression, but did not find one.\n%s", params[0], params[1])
		}
	case ECTLCExpectedExpressionInComputing:
		if len(params) >= 4 {
			return fmt.Sprintf("In computing %s, TLC expected a %s expression,\nbut instead found %s.\n%s", params[0], params[1], params[2], params[3])
		}
	case ECTLCExpectedExpressionInComputing2:
		if len(params) >= 3 {
			return fmt.Sprintf("In computing %s, TLC expected a %s expression,\nbut didn't find one.\n%s", params[0], params[1], params[2])
		}
	case ECTLCCheckingTemporalProps:
		if len(params) >= 3 {
			return fmt.Sprintf("Checking %stemporal properties for the %s state space with %s total distinct states at (%s)", params[2], params[0], params[1], messageNow())
		}
	case ECTLCCheckingTemporalPropsEnd:
		if len(params) >= 1 {
			return fmt.Sprintf("Finished checking temporal properties in %s at %s", params[0], messageNow())
		}
	case ECTLCLiveFormulaTautology:
		return "Temporal formula is a tautology (its negation is unsatisfiable)."
	case ECTLCConfigNoSpecButProperty:
		return "The stuttering counterexample above may be caused by the absence of a behavior specification (SPECIFICATION). Only INIT and NEXT have been provided, so TLC permits infinite stuttering. To rule out such counterexamples, use SPECIFICATION Spec, with Spec asserting a suitable fairness constraint (compare Chapter 8, page 87ff of Specifying Systems at https://lamport.azurewebsites.net/tla/book.html)."
	case ECTLCConfigNoFairnessButLiveProperty:
		if len(params) >= 2 {
			return fmt.Sprintf("The stuttering counterexample above may be caused by the absence of a fairness constraint in the behavior specification %s defined at %s. To rule out such counterexamples, conjoin a suitable fairness constraint to %s (compare Chapter 8, page 87ff of Specifying Systems at https://lamport.azurewebsites.net/tla/book.html).", params[0], params[1], params[0])
		}
	case ECTLCEnabledWrongFormula:
		if len(params) >= 2 {
			return fmt.Sprintf("In computing ENABLED, TLC encountered a temporal formula (%s).\n%s", params[0], params[1])
		}
	case ECTLCEncounteredFormulaInPredicate:
		if len(params) >= 2 {
			return fmt.Sprintf("TLC encountered a temporal formula (%s) when evaluating a predicate or action.\n%s", params[0], params[1])
		}
	case ECTLCEnvironmentJVMGC:
		return "Please run the Java VM, which executes TLC with a throughput optimized garbage collector, by passing the \"-XX:+UseParallelGC\" property."
	case ECTLCModuleOverrideStdout:
		return strings.Join(params, "")
	}
	if len(params) == 0 {
		return fmt.Sprintf("%d", code)
	}
	return fmt.Sprintf("%d %s", code, strings.Join(params, " "))
}

func messageNow() string {
	return time.Now().String()
}

func messageParam(params []string, index int) string {
	if index < 0 || index >= len(params) {
		return ""
	}
	return params[index]
}

func formatModeMCMessage(params []string, dfs bool) string {
	mode := "breadth-first search"
	if dfs {
		mode = "depth-first search"
	}
	workerCount := messageParam(params, 0)
	workerPlural := messageParam(params, 1)
	cores := messageParam(params, 2)
	osName := messageParam(params, 3)
	osVersion := messageParam(params, 4)
	osArch := messageParam(params, 5)
	javaVendor := messageParam(params, 6)
	javaVersion := messageParam(params, 7)
	javaArch := messageParam(params, 8)
	heap := messageParam(params, 9)
	offheap := messageParam(params, 10)
	seed := messageParam(params, 11)
	fp := messageParam(params, 12)
	pid := messageParam(params, 13)
	extra1 := messageParam(params, 14)
	extra2 := messageParam(params, 15)
	msg := fmt.Sprintf("Running %s Model-Checking with fp %s and seed %s with %s worker%s on %s cores with %sMB heap and %sMB offheap memory",
		mode, fp, seed, workerCount, workerPlural, cores, heap, offheap)
	if pid != "" {
		msg += fmt.Sprintf(" [pid: %s]", pid)
	}
	if dfs {
		return fmt.Sprintf("%s (%s %s %s, %s %s %s).", msg, osName, osVersion, osArch, javaVendor, javaVersion, javaArch)
	}
	return fmt.Sprintf("%s (%s %s %s, %s %s %s, %s, %s).", msg, osName, osVersion, osArch, javaVendor, javaVersion, javaArch, extra1, extra2)
}

func formatModeSimulationMessage(params []string) string {
	seed := messageParam(params, 0)
	workerCount := messageParam(params, 1)
	workerPlural := messageParam(params, 2)
	cores := messageParam(params, 3)
	osName := messageParam(params, 4)
	osVersion := messageParam(params, 5)
	osArch := messageParam(params, 6)
	javaVendor := messageParam(params, 7)
	javaVersion := messageParam(params, 8)
	javaArch := messageParam(params, 9)
	heap := messageParam(params, 10)
	offheap := messageParam(params, 11)
	pid := messageParam(params, 12)
	mode := messageParam(params, 13)
	msg := fmt.Sprintf("Running %s Simulation with seed %s with %s worker%s on %s cores with %sMB heap and %sMB offheap memory",
		mode, seed, workerCount, workerPlural, cores, heap, offheap)
	if pid != "" {
		msg += fmt.Sprintf(" [pid: %s]", pid)
	}
	return fmt.Sprintf("%s (%s %s %s, %s %s %s).", msg, osName, osVersion, osArch, javaVendor, javaVersion, javaArch)
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
	case ECTLCLiveFormulaTautology,
		ECTLCLiveCannotHandleFormula,
		ECTLCLiveWrongFormulaFormat:
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
