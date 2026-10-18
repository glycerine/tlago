/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// DistributedTLCTestCase.setUp unconditionally assumes false before starting
// any roles. Preserve that source assumption in ordinary original-test runs.
// Opt-in diagnostics call the assertion bodies below directly, keeping their
// deliberate assumption bypass separate from the faithful Java test contract.
func distributedOriginalModelSetup(t *testing.T) {
	t.Helper()
	t.Skip("DistributedTLCTestCase broken with OffHeapDiskFPSet.")
}

func TestJavaDieHardDistributed(t *testing.T) {
	distributedOriginalModelSetup(t)
	checkJavaDieHardDistributed(t)
}

func checkJavaDieHardDistributed(t *testing.T) {
	t.Helper()
	output := runNativeDistributedModel(t, "DieHard", false, true)
	requireDistributedOriginalEvent(t, output, tlc.ECTLCFinished)
	if len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireDistributedOriginalEvent(t, output, tlc.ECTLCBehaviorUpToThisPoint)
	requireDistributedOriginalEvent(t, output, tlc.ECTLCStatePrint2)
	want := []string{
		"/\\ action = \"nondet\"\n/\\ smallBucket = 0\n/\\ bigBucket = 0\n/\\ water_to_pour = 0",
		"/\\ action = \"fill big\"\n/\\ smallBucket = 0\n/\\ bigBucket = 5\n/\\ water_to_pour = 0",
		"/\\ action = \"pour big to small\"\n/\\ smallBucket = 3\n/\\ bigBucket = 2\n/\\ water_to_pour = 3",
		"/\\ action = \"empty small\"\n/\\ smallBucket = 0\n/\\ bigBucket = 2\n/\\ water_to_pour = 3",
		"/\\ action = \"pour big to small\"\n/\\ smallBucket = 2\n/\\ bigBucket = 0\n/\\ water_to_pour = 2",
		"/\\ action = \"fill big\"\n/\\ smallBucket = 2\n/\\ bigBucket = 5\n/\\ water_to_pour = 2",
		"/\\ action = \"pour big to small\"\n/\\ smallBucket = 3\n/\\ bigBucket = 4\n/\\ water_to_pour = 1",
	}
	states := nativeDistributedMessages(output, tlc.ECTLCStatePrint2)
	if len(states) != len(want) {
		t.Fatalf("trace length %d, want %d", len(states), len(want))
	}
	for i, state := range states {
		header, body, ok := strings.Cut(state, "\n")
		if !ok || !strings.HasPrefix(header, fmt.Sprintf("%d: ", i+1)) || strings.TrimSpace(body) != want[i] {
			t.Fatalf("trace state %d = %q, want %q", i+1, state, want[i])
		}
		info := strings.TrimPrefix(header, fmt.Sprintf("%d: ", i+1))
		if i == 0 {
			if info != "<Initial predicate>" {
				t.Fatalf("initial state action = %q", info)
			}
		} else if info == "<Initial predicate>" || strings.HasPrefix(info, "<Action") {
			t.Fatalf("successor state action = %q", info)
		}
	}
}

func TestJavaEWD840Distributed(t *testing.T) {
	distributedOriginalModelSetup(t)
	checkJavaEWD840Distributed(t)
}

func checkJavaEWD840Distributed(t *testing.T) {
	t.Helper()
	output := runNativeDistributedModel(t, "EWD840", false, true)
	requireDistributedOriginalEvent(t, output, tlc.ECTLCFinished)
	requireDistributedOriginalStats(t, output, 1, "114942")
	requireDistributedOriginalStats(t, output, 2, "0")
	if len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}

func TestJavaEWD840DistributedWithFPSet(t *testing.T) {
	distributedOriginalModelSetup(t)
	checkJavaEWD840DistributedWithFPSet(t)
}

func checkJavaEWD840DistributedWithFPSet(t *testing.T) {
	t.Helper()
	output := runNativeDistributedModel(t, "EWD840", true, true)
	requireDistributedOriginalEvent(t, output, tlc.ECTLCFinished)
	requireDistributedOriginalStats(t, output, 1, "114942")
	requireDistributedOriginalStats(t, output, 2, "0")
	if len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}

func TestJavaTSnapShotDistributed(t *testing.T) {
	distributedOriginalModelSetup(t)
	checkJavaTSnapShotDistributed(t)
}

func checkJavaTSnapShotDistributed(t *testing.T) {
	t.Helper()
	output := runNativeDistributedModel(t, "TSnapShot", false, true)
	requireDistributedOriginalEvent(t, output, tlc.ECTLCFinished)
	requireDistributedOriginalStats(t, output, 2, "0")
	if len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireDistributedOriginalEvent(t, output, tlc.ECTLCBehaviorUpToThisPoint)
}

func requireDistributedOriginalEvent(t *testing.T, output string, code int) {
	t.Helper()
	if len(nativeDistributedMessages(output, code)) == 0 {
		t.Fatalf("original required event %d absent", code)
	}
}

func requireDistributedOriginalStats(t *testing.T, output string, index int, expected string) {
	t.Helper()
	pattern := regexp.MustCompile(`^(\d+) states generated, (\d+) distinct states found, (\d+) states left on queue\.$`)
	for _, message := range nativeDistributedMessages(output, tlc.ECTLCStats) {
		if fields := pattern.FindStringSubmatch(message); len(fields) == 4 && fields[index+1] == expected {
			return
		}
	}
	t.Fatalf("original STATS parameter %d = %s absent", index, expected)
}
