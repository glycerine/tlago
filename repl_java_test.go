package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original tlc2.REPLTest.testProcessInput, in its original call order.
func TestREPLJavaProcessInput(t *testing.T) {
	oldMode := tlc.ToolIOGetMode()
	oldOutput, oldUserFile := tlc.TLCOutput, tlc.TLCOutputToUserFile
	t.Cleanup(func() { tlc.ToolIOSetMode(oldMode); tlc.TLCOutput, tlc.TLCOutputToUserFile = oldOutput, oldUserFile })
	repl := NewREPL(t.TempDir())
	for _, tc := range []struct{ expression, result string }{
		{"2+2", "4"}, {"4-2", "2"}, {`10 \div 2`, "5"},
		{`{1,2} \X {3,4}`, "{<<1, 3>>, <<1, 4>>, <<2, 3>>, <<2, 4>>}"},
		{`{1,2} \cup {3,4}`, "{1, 2, 3, 4}"}, {`{1,2} \cap {2,3}`, "{2}"},
		{"Append(<<1,2>>, 3)", "<<1, 2, 3>>"}, {"Append(3, <<1,2>>)", ""},
		{"Tail(<<1,2,3>>)", "<<2, 3>>"}, {"Head(<<1,2,3>>)", "1"},
		{`<<1,2>> \o <<3>>`, "<<1, 2, 3>>"}, {"invalid", ""}, {"123abc", ""},
	} {
		if got := repl.ProcessInput(tc.expression); got != tc.result {
			t.Fatalf("processInput(%q) = %q, want %q", tc.expression, got, tc.result)
		}
	}
}
