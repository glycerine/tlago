// Copyright (c) 2025, Oracle and/or its affiliates.
package tlc

//go:generate go run generate_mp.go ../../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/output/MP.java ../../tlaplus/tlatools/org.lamport.tlatools/src/tlc2/output/EC.java output_mp_generated.go

import (
	"fmt"
	"strings"
)

// javaMessageTemplate ports the twenty conditional MP.getMessage0 cases.
// The remaining 225 literal append sequences are generated from the source.
// Substitution happens afterwards, retaining missing markers and null boundaries.
func javaMessageTemplate(code int, params []*string, class Severity) string {
	if text, ok := javaStaticMessageTemplate(code); ok {
		return text
	}
	count := len(params)
	switch code {
	case ECSystemErrorReadingPool:
		if count == 2 {
			return "when reading pool file %2% (StatePoolReader.run):\n%1%"
		}
		return "when reading the disk (StatePoolReader.run):\n%1%"
	case ECSystemErrorCleaningPool:
		if class == SeverityError {
			return "Exception cleaning up an obsolete disk file.\n%1%"
		}
		if class == SeverityWarning {
			return "Failed to clean up an obsolete disk file. Please manually delete %1% if free disk space is low."
		}
		return ""
	case ECTLCStateNotCompletelySpecifiedNext:
		if count == 3 {
			return "Successor state is not completely specified by action %1% of the next-state relation. The following variable%2% not defined: %3%.\n"
		}
		if count == 2 {
			return "Successor state is not completely specified by the next-state action. The following variable%1% not defined: %2%.\n"
		}
		return "Successor state is not completely specified by the next-state action.\n"
	case ECTLCInvariantViolatedLevel:
		text := "The invariant %1% is not a state predicate (one with no primes or temporal operators)."
		if count > 1 {
			text += "\nNote that a bug can cause TLC to incorrectly report this error.\nIf you believe your TLA+ or PlusCal specification to be correct,\nplease check if this bug described in LevelNode.java starting at line 590ff affects you."
		}
		return text
	case ECTLCInvariantEvaluationFailed:
		if count == 1 {
			return "Evaluating invariant %1% failed."
		}
		if count == 2 {
			return "Evaluating invariant %1% failed.\n%2%"
		}
		return ""
	case ECTLCActionPropertyEvaluationFailed:
		if count == 1 {
			return "Evaluating action property %1% failed."
		}
		if count == 2 {
			return "Evaluating action property %1% failed.\n%2%"
		}
		return ""
	case ECTLCTemporalPropertyViolated:
		if count == 0 {
			return "Temporal properties were violated.\n"
		}
		if count == 1 {
			return "Temporal property %1% was violated.\n"
		}
		var text strings.Builder
		text.WriteString("Temporal properties ")
		for i := 1; i <= count; i++ {
			if i > 1 && i < count {
				text.WriteString(", ")
			} else if i == count {
				if count == 2 {
					text.WriteString(" and ")
				} else {
					text.WriteString(", and ")
				}
			}
			fmt.Fprintf(&text, "%%%d%%", i)
		}
		text.WriteString(" were violated.\n")
		return text.String()
	case ECTLCBackToState:
		if count != 1 && count != 2 {
			return ""
		}
		text := "Back to state %1%"
		if messageToolMode() {
			text = "%1%: Back to state"
		}
		if count == 2 {
			text += ": %2%"
		}
		return text + "\n"
	case ECTLCConfigIDRequiresNoArg:
		if count == 1 {
			return "TLC requires %1% not to take any argument."
		}
		if count == 2 {
			return "TLC requires %1% not to take any argument, but one was given: %2%"
		}
		return ""
	case ECTLCLiveCannotHandleFormula:
		if count > 1 {
			return "TLC cannot handle the temporal formula %1%:\n%2%"
		}
		return "TLC cannot handle the temporal formula %1%"
	case ECTLCModeMC, ECTLCModeMCDFS:
		text := "Running breadth-first search Model-Checking with fp %13% and seed %12% with %1% worker%2% on %3% cores with %10%MB heap and %11%MB offheap memory"
		if code == ECTLCModeMCDFS {
			text = "Running depth-first search Model-Checking with fp %13% and seed %12% with %1% worker%2% on %3% cores with %10%MB heap and %11%MB offheap memory"
		}
		if messageParameterPresent(params, 13) {
			text += " [pid: %14%]"
		} else {
			text += "%14%"
		}
		text += " (%4% %5% %6%, %7% %8% %9%"
		if code == ECTLCModeMC {
			text += ", %15%, %16%"
		}
		return text + ")."
	case ECTLCModeSimu:
		text := "Running %14% Simulation with seed %1% with %2% worker%3% on %4% cores with %11%MB heap and %12%MB offheap memory"
		if messageParameterPresent(params, 12) {
			text += " [pid: %13%]"
		} else {
			text += "%13%"
		}
		return text + " (%5% %6% %7%, %8% %9% %10%)."
	case ECTLCSuccess:
		text := "Model checking completed. No error has been found.\n  Estimates of the probability that TLC did not check all reachable states\n  because two distinct states had the same fingerprint:\n  calculated (optimistic):  %1%"
		if count != 1 {
			text += "\n  based on the actual fingerprints:  %2%"
		}
		return text
	case ECTLCProgressStats:
		if count == 4 {
			return "Progress(%1%) at " + messageNow() + ": %2% states generated, %3% distinct states found, %4% states left on queue."
		}
		if count == 6 {
			return "Progress(%1%) at " + messageNow() + ": %2% states generated (%5% s/min), %3% distinct states found (%6% ds/min), %4% states left on queue."
		}
		return ""
	case ECTLCProgressStatsDFID:
		if messageToolMode() {
			return "Progress(-1) at " + messageNow() + ": %1% states generated, %2% distinct states found, -1 states left on queue."
		}
		return "Progress: %1% states generated, %2% distinct states found."
	case ECTLCProgressSimu:
		if messageToolMode() {
			return "Progress(%2%) at " + messageNow() + ": %1% states generated, -1 distinct states found, -1 states left on queue."
		}
		return "Progress: %1% states checked, %2% traces generated (trace length: mean=%3%, var(x)=%4%, sd=%5%)"
	case ECCheckParamUnrecognized:
		if count == 1 {
			return "Unrecognized option: %1%"
		}
		if count == 2 {
			return "Unrecognized option in file %2%: %1%"
		}
		return ""
	case ECTLCStatePrint2:
		if mpGeneralDebug {
			return "%1%: %2%\n%3%fp: %4%\n"
		}
		return "%1%: %2%\n%3%"
	case ECGeneral:
		var text strings.Builder
		for i := 1; i <= count; i++ {
			fmt.Fprintf(&text, "%%%d%%", i)
		}
		return text.String()
	}
	return "Wrong invocation of TLC error printer. Error code not found."
}

func messageToolMode() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.Tool
}

// Source !"".equals(parameters[index]) treats null as present and throws on
// missing array elements. Preserve that check before any parameter replacement.
func messageParameterPresent(params []*string, index int) bool {
	if index >= len(params) {
		panic(NewArrayIndexOutOfBoundsException(index, len(params)))
	}
	return params[index] == nil || *params[index] != ""
}
