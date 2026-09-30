package tlc

import "fmt"

func PrintRuntimeErrorStateTraceState(currentState *TLCStateMut, num int) {
	stateString := ""
	if currentState != nil {
		stateString = currentState.String()
	}
	PrintState(ECTLCStatePrint1, []string{fmt.Sprintf("%d", num), stateString}, currentState, num)
}

func PrintStandaloneErrorState(currentState *TLCStateMut) {
	stateString := ""
	if currentState != nil {
		stateString = currentState.String()
	}
	PrintState(ECTLCStatePrint1, []string{"", stateString}, currentState, -1)
}

func PrintInvariantViolationStateTraceState(currentStateInfo *TLCStateInfo, args ...any) {
	var previousState *TLCStateMut
	num := 0
	isFinal := false
	if currentStateInfo != nil {
		num = int(currentStateInfo.StateNumber)
		if currentStateInfo.State != nil && !currentStateInfo.State.IsInitial() {
			previousState = currentStateInfo.State.Predecessor()
		}
	}
	if len(args) > 0 {
		if typed, ok := args[0].(*TLCStateMut); ok {
			previousState = typed
		}
	}
	if len(args) > 1 {
		if typed, ok := args[1].(int); ok {
			num = typed
		}
	}
	if len(args) > 2 {
		if typed, ok := args[2].(bool); ok {
			isFinal = typed
		}
	}
	printInvariantViolationStateTraceState(currentStateInfo, previousState, num, isFinal)
}

func printInvariantViolationStateTraceState(currentStateInfo *TLCStateInfo, previousState *TLCStateMut, num int, isFinal bool) {
	_ = isFinal
	stateString := ""
	fingerprint := "-1"
	infoString := ""
	if currentStateInfo != nil {
		infoString = fmt.Sprint(currentStateInfo.Info)
		state := currentStateInfo.State
		if state != nil {
			if previousState != nil && printDiffsOnly() {
				stateString = state.StringForVariables(previousState)
			} else {
				stateString = state.String()
			}
			if state.AllAssigned() {
				fingerprint = fmt.Sprintf("%d", currentStateInfo.FingerPrint())
			}
		}
	}
	PrintStateInfo(ECTLCStatePrint2, []string{
		fmt.Sprintf("%d", num),
		infoString,
		stateString,
		fingerprint,
	}, currentStateInfo, num)
}

func PrintStutteringState(num int) {
	PrintState(ECTLCStatePrint3, []string{fmt.Sprintf("%d", num+1)}, nil, num+1)
}

func PrintBackToState(currentStateInfo *TLCStateInfo, stateNum int) {
	infoString := ""
	if currentStateInfo != nil {
		infoString = fmt.Sprint(currentStateInfo.Info)
	}
	PrintMessage(ECTLCBackToState, fmt.Sprintf("%d", stateNum), infoString)
}

func printDiffsOnly() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.PrintDiffsOnly
}
