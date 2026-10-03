/*******************************************************************************
 * Copyright (c) 2018 Microsoft Research. All rights reserved.
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
// Port of the four OpApplNodeWrapperTest reporting methods.
package tlc

import (
	"strconv"
	"strings"
	"testing"
)

func javaCoverageRecorder(t *testing.T) *MemoryRecorder {
	t.Helper()
	previous := Globals.CoverageInterval
	Globals.CoverageInterval = 1
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() {
		RemoveMessageRecorder(recorder)
		Globals.CoverageInterval = previous
	})
	return recorder
}

func javaCoverageNode(count int64) CostModel {
	node := NewOpApplNode(&SymbolNode{Name: UniqueStringOf(strconv.FormatInt(count, 10))})
	// DummyOpApplNode(SymbolNode) uses SyntaxTreeNode.nullSTN in Java.
	node.SetSourceLocation(NewSourceLocation("--TLA+ BUILTINS--", 0, 0, 0, 0))
	return NewCostModel(node).IncInvocations(count)
}

func requireJavaCoverageContains(t *testing.T, recorder *MemoryRecorder, want string) {
	t.Helper()
	var output strings.Builder
	for _, message := range recorder.Messages {
		output.WriteString(message.Text)
		output.WriteByte('\n')
	}
	if !strings.Contains(output.String(), want) {
		t.Fatalf("coverage output=%q, want substring %q", output.String(), want)
	}
}

func TestJavaReportCoverage01(t *testing.T) {
	recorder := javaCoverageRecorder(t)
	root := NewCostModel(nil)
	root.Report()
	if len(recorder.Messages) != 0 {
		t.Fatalf("expected empty output, got %v", recorder.Messages)
	}
	root.AddChildModel(NewCostModel(nil))
}

func TestJavaReportCoverage02(t *testing.T) {
	recorder := javaCoverageRecorder(t)
	root := NewCostModel(nil).IncInvocations(42)
	root.AddChildModel(javaCoverageNode(23))
	root.AddChildModel(javaCoverageNode(24))
	root.AddChildModel(javaCoverageNode(0)) // Not reported.
	root.Report()
	requireJavaCoverageContains(t, recorder, "  Unknown location: 42\n"+
		"  |In module --TLA+ BUILTINS--: 23\n"+
		"  |In module --TLA+ BUILTINS--: 24")
}

func TestJavaReportCoverage03(t *testing.T) {
	recorder := javaCoverageRecorder(t)
	root := NewCostModel(nil).IncInvocations(42)
	childA := javaCoverageNode(23)
	childA.AddChildModel(javaCoverageNode(546))
	root.AddChildModel(childA)
	childB := javaCoverageNode(24)
	root.AddChildModel(childB)
	childB.AddChildModel(javaCoverageNode(0)) // Not reported because zero.
	childC := javaCoverageNode(0)
	root.AddChildModel(childC)
	childC.AddChildModel(javaCoverageNode(17)) // Reported despite C being zero.
	root.Report()
	requireJavaCoverageContains(t, recorder, "  Unknown location: 42\n"+
		"  |In module --TLA+ BUILTINS--: 23\n"+
		"  ||In module --TLA+ BUILTINS--: 546\n"+
		"  |In module --TLA+ BUILTINS--: 24\n"+
		"  |In module --TLA+ BUILTINS--: 17")
}

func TestJavaReportCoverage04(t *testing.T) {
	recorder := javaCoverageRecorder(t)
	root := NewCostModel(nil).IncInvocations(1)
	childA := javaCoverageNode(1)
	root.AddChildModel(childA)
	childA.AddChildModel(javaCoverageNode(131072))
	cChildA := javaCoverageNode(131072)
	childA.AddChildModel(cChildA)
	cChildA.AddChildModel(javaCoverageNode(1))
	childB := javaCoverageNode(1)
	root.AddChildModel(childB)
	root.Report()
	requireJavaCoverageContains(t, recorder, "  Unknown location: 1\n"+
		"  |In module --TLA+ BUILTINS--: 1\n"+
		"  ||In module --TLA+ BUILTINS--: 131072\n"+
		"  ||In module --TLA+ BUILTINS--: 131072\n"+
		"  |||In module --TLA+ BUILTINS--: 1\n"+
		"  |In module --TLA+ BUILTINS--: 1")
}
