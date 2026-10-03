/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
// Port of EvalExceptionLivenessTest.testSpec and its coverage-disabled constructor.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaEvalExceptionLiveness(t *testing.T) {
	result := runJavaTLCModelTestWithCoverage(t, "DistBakery3aAuxMC", false, "-config", "DistBakery3aAuxMC.tla")
	if result.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit status=%d, want ERROR", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "950", "555", "89")
	requireJavaTLCRecordedParams(t, result, tlc.ECGeneral,
		"TLC threw an unexpected exception.\n"+
			"This was probably caused by an error in the spec or model.\n"+
			"See the User Output or TLC Console for clues to what happened.\n"+
			"The exception was a java.lang.RuntimeException\n"+
			": Attempted to check equality of the function <<-2>> with the value:\n"+
			"-2")
	expectedTrace := []string{
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 0 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> TRUE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"ncs\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ numBar = (-2 :> 0 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"ncs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"M1s\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 0 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> TRUE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"L1\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ numBar = (-2 :> 0 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"L1\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"M1s\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 0 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"M1s\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ numBar = (-2 :> 0 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"M1\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"M1s\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 0 @@ -1 :> 0))\n/\\ msgStop = (-2 :> TRUE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 2 @@ -1 :> 1)\n/\\ pc = (-2 :> \"M1s\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ numBar = (-2 :> 0 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"M1\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L2\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 0 @@ -1 :> 0))\n/\\ msgStop = (-2 :> TRUE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"enter\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ numBar = (-2 :> 0 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"M1\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L2\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 0 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"wait\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = ( -2 :> (-2 :> <<>> @@ -1 :> <<[num |-> 1, type |-> \"write\"]>>) @@\n  -1 :> (-2 :> <<>> @@ -1 :> <<>>) )\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"pcs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L2\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"wait\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = ( -2 :> (-2 :> <<>> @@ -1 :> <<>>) @@\n  -1 :> (-2 :> <<[type |-> \"ack\"]>> @@ -1 :> <<>>) )\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"pcs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L2\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"wait\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {-1} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"pcs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L2\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> TRUE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"L2\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {-1} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"pcs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L2\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> TRUE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 2 @@ -1 :> 1)\n/\\ pc = (-2 :> \"L2\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {-1} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"pcs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L3\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> TRUE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"L3\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {-1} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"pcs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"L3\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> TRUE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 2 @@ -1 :> 1)\n/\\ pc = (-2 :> \"L3\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {-1} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"pcs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"scs\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> TRUE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"cs\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {-1} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 1 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"cs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"scs\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> FALSE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"exit\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {-1} @@ -1 :> {})\n/\\ net = (-2 :> (-2 :> <<>> @@ -1 :> <<>>) @@ -1 :> (-2 :> <<>> @@ -1 :> <<>>))\n/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ numBar = (-2 :> 0 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"ncs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"scs\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
		"/\\ rnum = (-2 :> (-2 :> 0 @@ -1 :> 0) @@ -1 :> (-2 :> 1 @@ -1 :> 0))\n/\\ msgStop = (-2 :> FALSE @@ -1 :> FALSE)\n/\\ mustRdBar = (<<-2, -1>> :> FALSE @@ <<-1, -2>> :> FALSE)\n/\\ j = (-2 :> 1 @@ -1 :> 1)\n/\\ pc = (-2 :> \"scs\" @@ -1 :> \"ncs\" @@ 12 :> \"a\" @@ 21 :> \"a\")\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ net = ( -2 :> (-2 :> <<>> @@ -1 :> <<[num |-> 0, type |-> \"write\"]>>) @@\n  -1 :> (-2 :> <<>> @@ -1 :> <<>>) )\n/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ numBar = (-2 :> 0 @@ -1 :> 0)\n/\\ tempBar = (<<-2, -1>> :> 0 @@ <<-1, -2>> :> 0)\n/\\ pcBar = ( <<-2>> :> \"ncs\" @@\n  <<-1>> :> \"ncs\" @@\n  <<-2, -1>> :> \"scs\" @@\n  <<-1, -2>> :> \"M1s\" @@\n  <<-2, -1, \"wr0\">> :> \"wr0\" @@\n  <<-1, -2, \"wr0\">> :> \"wr0\" )",
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expectedTrace) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expectedTrace))
	}
	for i, record := range records {
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		info, ok := record.StateInfo.Info.(string)
		if !ok || info == tlc.InitialPredicate || strings.HasPrefix(info, "<Action") {
			t.Fatalf("trace action %d=%v, want a named action", i+1, record.StateInfo.Info)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
}
