package tlc

import "testing"

func TestOutputMessageControlMirrorsJavaRecorderSemantics(t *testing.T) {
	Globals.Lock()
	oldWarn := Globals.Warn
	oldSuppressed := Globals.SuppressedMessages
	oldAsErrors := Globals.MessagesAsErrors
	Globals.Warn = true
	Globals.SuppressedMessages = NewInsMap[int, bool]()
	Globals.MessagesAsErrors = NewInsMap[int, bool]()
	Globals.Unlock()
	ClearMessageRecorders()
	t.Cleanup(func() {
		ClearMessageRecorders()
		Globals.Lock()
		Globals.Warn = oldWarn
		Globals.SuppressedMessages = oldSuppressed
		Globals.MessagesAsErrors = oldAsErrors
		Globals.Unlock()
	})

	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)

	SuppressTLCMessage(ECTLCStarting)
	PrintMessage(ECTLCStarting)
	records := recorder.Records(ECTLCStarting)
	if len(records) != 1 {
		t.Fatalf("suppressed PrintMessage records = %d, want 1 like Java MP recorder", len(records))
	}
	if !records[0].Suppressed || records[0].Severity != SeverityNone {
		t.Fatalf("suppressed PrintMessage = suppressed %v severity %v, want true/none", records[0].Suppressed, records[0].Severity)
	}

	SuppressTLCMessage(ECTLCDeadlockReached)
	PrintError(ECTLCDeadlockReached)
	records = recorder.Records(ECTLCDeadlockReached)
	if len(records) != 1 {
		t.Fatalf("suppressed PrintError records = %d, want 1", len(records))
	}
	if records[0].Suppressed || records[0].Severity != SeverityError {
		t.Fatalf("PrintError suppression = suppressed %v severity %v, want false/error like Java printError", records[0].Suppressed, records[0].Severity)
	}

	Globals.Lock()
	Globals.Warn = false
	Globals.Unlock()
	PrintWarning(ECGeneral, "quiet warning")
	records = recorder.Records(ECGeneral)
	if len(records) != 1 {
		t.Fatalf("disabled warning records = %d, want 1 like Java MP recorder", len(records))
	}
	if !records[0].Suppressed || records[0].Severity != SeverityWarning {
		t.Fatalf("disabled warning = suppressed %v severity %v, want true/warning", records[0].Suppressed, records[0].Severity)
	}

	Globals.Lock()
	Globals.Warn = true
	Globals.Unlock()
	TreatTLCMessageAsError(ECGeneral)
	before := len(recorder.Records(ECGeneral))
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		PrintWarning(ECGeneral, "boom")
	}()
	if recovered == nil {
		t.Fatalf("PrintWarning with messages-as-errors did not panic")
	}
	evalErr, ok := recovered.(*EvalException)
	if !ok || evalErr.GetErrorCode() != ECGeneral {
		t.Fatalf("PrintWarning panic = %T %[1]v, want *EvalException code %d", recovered, ECGeneral)
	}
	after := len(recorder.Records(ECGeneral))
	if after != before {
		t.Fatalf("messages-as-errors recorded warning before abort = %d, want unchanged %d like Java Assert.fail path", after, before)
	}
}
