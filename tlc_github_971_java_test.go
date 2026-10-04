/*******************************************************************************
 * Copyright (c) 2024 Microsoft Research. All rights reserved.
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
package tlago

import (
	"math"
	"sync"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Java Value.equals throws on comparison failures; propagate Go's error too.
func javaGithub971Equal(left, right tlc.Value) bool {
	equal, err := left.Equal(right)
	if err != nil {
		panic(err)
	}
	return equal
}

// Original anonymous IStateQueue wrapper and CountDownLatch in beforeSetUp.
// Embedding delegates every other original queue method to MemStateQueue.
type javaGithub971Queue struct {
	tlc.StateQueue
	mu        sync.Mutex
	remaining int
	signal    chan struct{}
	blocks    func(*tlc.TLCStateMut) bool
}

func (q *javaGithub971Queue) SEnqueue(state *tlc.TLCStateMut) {
	q.StateQueue.SEnqueue(state)
	if q.blocks(state) {
		<-q.signal
	}
}

func (q *javaGithub971Queue) SDequeue() *tlc.TLCStateMut {
	q.mu.Lock()
	if q.remaining > 0 {
		q.remaining--
		if q.remaining == 0 {
			close(q.signal)
		}
	}
	q.mu.Unlock()
	return q.StateQueue.SDequeue()
}

func runJavaGithub971(t *testing.T, root, config, lncheck string, count int, blocks func(*tlc.TLCStateMut) bool) *tlc.Result {
	t.Helper()
	setJavaModelLivenessThreshold(t, math.MaxFloat64)
	var queue *javaGithub971Queue
	return runJavaTLCModelTestWithRunnerSetup(t, "Github971", root, func(meta, traceDirectory string) []string {
		queue = &javaGithub971Queue{StateQueue: tlc.NewMemStateQueue(meta), remaining: count, signal: make(chan struct{}), blocks: blocks}
		args := []string{"-metadir", meta, "-teSpecOutDir", traceDirectory, "-fp", "0", "-seed", "1", "-workers", "2", "-checkpoint", "0", "-deadlock", "-noGenerateSpecTE", "-config", config}
		if lncheck != "" {
			args = append(args, "-lncheck", lncheck)
		}
		return args
	}, func(runner *tlc.TLC) { runner.StateQueue = queue })
}

// Original Github971aTest.testSpec and beforeSetUp barrier.
func TestJavaGithub971a(t *testing.T) {
	r := runJavaGithub971(t, "Github971", "Github971a.cfg", "off", 3, func(state *tlc.TLCStateMut) bool {
		return javaGithub971Equal(tlc.NewIntValue(1), state.Lookup(tlc.UniqueStringOf("x")))
	})
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AtMostOnce")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{`x = 0`, `x = 1`, `x = 0`, `x = 1`}, false)
}

// Original Github971bTest.testSpec and beforeSetUp barrier.
func TestJavaGithub971b(t *testing.T) {
	r := runJavaGithub971(t, "Github971", "Github971b.cfg", "off", 4, func(state *tlc.TLCStateMut) bool {
		return javaGithub971Equal(tlc.NewIntValue(0), state.Lookup(tlc.UniqueStringOf("x")))
	})
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "7", "3", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AtMostOnce")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{`x = -1`, `x = 1`, `x = 0`, `x = 1`}, false)
}

// Original Github971cTest.testSpec and beforeSetUp barrier.
func TestJavaGithub971c(t *testing.T) {
	r := runJavaGithub971(t, "Github971", "Github971c.cfg", "off", 4, func(state *tlc.TLCStateMut) bool {
		return javaGithub971Equal(tlc.NewIntValue(0), state.Lookup(tlc.UniqueStringOf("x")))
	})
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "13", "4", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AtMostOnceC")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{`x = -1`, `x = 1`, `x = 0`}, false)
}

// Original Github971dTest.testSpec and beforeSetUp barrier.
func TestJavaGithub971d(t *testing.T) {
	r := runJavaGithub971(t, "Github971", "Github971d.cfg", "final", 3, func(state *tlc.TLCStateMut) bool {
		return javaGithub971Equal(tlc.NewIntValue(1), state.Lookup(tlc.UniqueStringOf("x")))
	})
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "LeadsTo")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{`x = 0`, `x = 1`, `x = 0`, `x = 1`, `x = 0`, `x = 1`, `x = 0`}, false)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "2")
}

// Original Github971eTest.testSpec and beforeSetUp barrier.
func TestJavaGithub971e(t *testing.T) {
	r := runJavaGithub971(t, "Github971e", "Github971d.cfg", "", 4, func(state *tlc.TLCStateMut) bool {
		return javaGithub971Equal(tlc.NewIntValue(0), state.Lookup(tlc.UniqueStringOf("x"))) && javaGithub971Equal(tlc.NewStringValue("b"), state.Lookup(tlc.UniqueStringOf("y")))
	})
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "17", "4", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "LeadsTo")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{`/\ x = 0
/\ y = "a"`, `/\ x = 1
/\ y = "a"`, `/\ x = 0
/\ y = "a"`, `/\ x = 1
/\ y = "b"`, `/\ x = 1
/\ y = "a"`, `/\ x = 1
/\ y = "b"`}, false)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "2")
}
