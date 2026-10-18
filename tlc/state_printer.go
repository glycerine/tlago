package tlc

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const statePrinterOverwriteProperty = "tlc2.output.StatePrinter.overwrite"

var statePrinterStatics struct {
	once     sync.Once
	interval int
}

// InitializeStatePrinterStatics represents a fresh Java classloader. Ordinary
// printer calls retain the captured interval. Reset only with an idle runtime.
func InitializeStatePrinterStatics() {
	statePrinterStatics.once = sync.Once{}
	statePrinterStatics.interval = 0
}

func PrintRuntimeErrorStateTraceState(currentState *TLCStateMut, num int) {
	statePrinterOverwriteInterval()
	if currentState == nil {
		panic(NewNullPointerException())
	}
	num = int(int32(num))
	stateString := currentState.String()
	PrintState(ECTLCStatePrint1, []string{fmt.Sprintf("%d", num), stateString}, currentState, num)
}

func PrintStandaloneErrorState(currentState *TLCStateMut) {
	statePrinterOverwriteInterval()
	if currentState == nil {
		panic(NewNullPointerException())
	}
	stateString := currentState.String()
	PrintState(ECTLCStatePrint1, []string{"", stateString}, currentState, -1)
}

func PrintInvariantViolationStateTraceState(currentStateInfo *TLCStateInfo, args ...any) {
	if len(args) != 0 && len(args) != 2 && len(args) != 3 {
		panic(NewIllegalArgumentException("PrintInvariantViolationStateTraceState accepts no extra arguments, a previous state and index, or those values plus a final-state flag"))
	}
	var previousState *TLCStateMut
	num := 0
	isFinal := false
	if len(args) != 0 {
		if args[0] != nil {
			var ok bool
			previousState, ok = args[0].(*TLCStateMut)
			if !ok {
				panic(NewIllegalArgumentException("PrintInvariantViolationStateTraceState requires a mutable previous state"))
			}
		}
		var ok bool
		num, ok = args[1].(int)
		if !ok {
			panic(NewIllegalArgumentException("PrintInvariantViolationStateTraceState requires an integer state index"))
		}
		if len(args) == 3 {
			isFinal, ok = args[2].(bool)
			if !ok {
				panic(NewIllegalArgumentException("PrintInvariantViolationStateTraceState requires a boolean final-state flag"))
			}
		}
		num = int(int32(num))
	}
	statePrinterOverwriteInterval()
	if currentStateInfo == nil || currentStateInfo.State == nil {
		panic(NewNullPointerException())
	}
	if len(args) == 0 {
		num = int(int32(currentStateInfo.StateNumber))
		if !currentStateInfo.State.IsInitial() {
			previousState = currentStateInfo.State.Predecessor()
		}
	}
	printInvariantViolationStateTraceState(currentStateInfo, previousState, num, isFinal)
}

func printInvariantViolationStateTraceState(currentStateInfo *TLCStateInfo, previousState *TLCStateMut, num int, isFinal bool) {
	state := currentStateInfo.State
	var stateString string
	if previousState != nil && printDiffsOnly() {
		stateString = state.StringForVariables(previousState)
	} else {
		stateString = state.String()
	}
	fingerprint := "-1"
	if state.AllAssigned() {
		fingerprint = fmt.Sprintf("%d", currentStateInfo.FingerPrint())
	}
	infoString := statePrinterInfoString(currentStateInfo.Info)
	message := PrintStateInfo(ECTLCStatePrint2, []string{
		fmt.Sprintf("%d", num),
		infoString,
		stateString,
		fingerprint,
	}, currentStateInfo, num)
	if interval := statePrinterOverwriteInterval(); interval > 0 {
		time.Sleep(time.Duration(interval) * time.Millisecond)
		if !isFinal {
			lines := strings.Count(message, "\n") + 1
			toolIOPrintfString(strings.Repeat("\033[F\033[K", lines))
		}
	}
}

func PrintStutteringState(num int) {
	statePrinterOverwriteInterval()
	ordinal := int(int32(num) + 1)
	PrintState(ECTLCStatePrint3, []string{fmt.Sprintf("%d", ordinal)}, nil, ordinal)
}

func PrintBackToState(currentStateInfo *TLCStateInfo, stateNum int) {
	statePrinterOverwriteInterval()
	if currentStateInfo == nil {
		panic(NewNullPointerException())
	}
	stateNum = int(int32(stateNum))
	infoString := statePrinterInfoString(currentStateInfo.Info)
	if toolModeEnabled() {
		PrintState(ECTLCBackToState, []string{fmt.Sprintf("%d", stateNum), infoString}, nil, stateNum)
		return
	}
	PrintMessage(ECTLCBackToState, fmt.Sprintf("%d", stateNum), infoString)
}

func statePrinterInfoString(info any) string {
	if isNil(info) {
		panic(NewNullPointerException())
	}
	if text, ok := info.(fmt.Stringer); ok {
		// fmt.Sprint recovers String panics, whereas Java toString propagates.
		return text.String()
	}
	return fmt.Sprint(info)
}

func printDiffsOnly() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.PrintDiffsOnly
}

func toolModeEnabled() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return Globals.Tool
}

func statePrinterOverwriteInterval() int {
	statePrinterStatics.once.Do(func() {
		statePrinterStatics.interval = -1
		if value, ok := tlcLookupSystemProperty(statePrinterOverwriteProperty); ok {
			if parsed, ok := javaDecodeIntProperty(value); ok {
				statePrinterStatics.interval = parsed
			}
		}
	})
	return statePrinterStatics.interval
}
