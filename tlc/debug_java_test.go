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
// Translation of all 17 TLCDebuggerTest methods after worker suspension.
package tlc

import "testing"

func TestJavaTLCDebugger(t *testing.T) {
	integer := func(value int) *int { return &value }
	type request struct {
		start, levels *int
		ids           []int
	}
	ascending := func(size int) []int {
		ids := make([]int, size)
		for i := range ids {
			ids[i] = i
		}
		return ids
	}
	descending := func(size int) []int {
		ids := ascending(size)
		for i := range ids {
			ids[i] = size - 1 - i
		}
		return ids
	}
	cases := []struct {
		name     string
		frames   int
		halted   bool
		requests []request
	}{
		{"testStackFrameWhileRunning", 20, false, []request{{nil, nil, nil}}},
		{"testStackFramePaginationEmpty", 0, true, []request{{nil, nil, nil}}},
		{"testStackFramePaginationEmpty2", 0, true, []request{{nil, integer(20), nil}}},
		{"testStackFramePaginationEmpty3", 0, true, []request{{integer(20), nil, nil}}},
		{"testStackFramePaginationEmpty4", 0, true, []request{{integer(20), integer(20), nil}}},
		{"testStackFramePaginationEmpty5", 0, true, []request{{integer(-20), nil, nil}}},
		{"testStackFramePaginationEmpty6", 0, true, []request{{nil, integer(-20), nil}}},
		{"testStackFramePaginationEmpty7", 0, true, []request{{integer(-20), integer(-20), nil}}},
		{"testStackFramePaginationLevelNull", 1, true, []request{{nil, nil, []int{0}}}},
		{"testStackFramePaginationLevel0", 1, true, []request{{nil, integer(0), []int{0}}, {nil, nil, []int{0}}}},
		{"testStackFramePaginationStartFrame0", 1, true, []request{{integer(0), nil, []int{0}}, {nil, nil, []int{0}}}},
		{"testStackFramePaginationStartFrame1", 1, true, []request{{integer(1), nil, nil}}},
		{"testStackFramePaginationStartFrameNegative", 1, true, []request{{integer(-1), nil, nil}}},
		{"testStackFramePagination", 20, true, []request{{nil, nil, descending(20)}}},
		{"testStackFramePaginationsStartFrame1", 20, true, []request{{integer(1), nil, descending(19)}}},
		{"testStackFramePaginationsStartFrameSubList", 20, true, []request{{integer(5), integer(5), []int{14, 13, 12, 11, 10}}}},
		{"testStackFramePaginationsStartFrameOutOfRange", 20, true, []request{{integer(20), nil, nil}}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			for _, request := range test.requests {
				debugger := NewTLCDebugger(nil)
				debugger.ExecutionIsHalted = test.halted
				// Java's deque stores newest first; Go's stack stores newest last.
				for id := 0; id < test.frames; id++ {
					debugger.Stack = append(debugger.Stack, NewDebuggerBaseFrame(&TLCStackFrame{ID: id}))
				}
				frames := debugger.StackTrace(TLCStackTraceArguments{StartFrame: request.start, Levels: request.levels}).StackFrames
				if len(frames) != len(request.ids) {
					t.Fatalf("frames=%d, want %d", len(frames), len(request.ids))
				}
				// Only the three multi-frame Java methods assert frame IDs; the
				// one-frame methods assert size alone.
				if test.frames == 20 && test.halted && request.ids != nil {
					for i, id := range request.ids {
						if frames[i].ID != id {
							t.Fatalf("frame %d id=%d, want %d", i, frames[i].ID, id)
						}
					}
				}
			}
		})
	}
}
