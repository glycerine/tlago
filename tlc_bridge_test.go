package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestPossibleStandardModuleKeepsOnlyCountsNative(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "PossibleBridge.tla")
	writeFile(t, root, `---- MODULE PossibleBridge ----
EXTENDS _Possible
P == INSTANCE _Possible
VARIABLE x
Init == x = 0
Next == x' = x
====`)

	spec, diags := LoadSanySpec(root, LoadOptions{})
	requireNoErrors(t, diags)
	requireNoErrors(t, CheckSpec(spec))

	cfg := tlc.NewModelConfig("PossibleBridge")
	tool, toolDiags := BuildTLCTool(spec, cfg, tlc.RuntimeParameters{})
	requireNoErrors(t, toolDiags)

	if _, ok := tool.DefnsByName[tlc.UniqueStringOf("_Counts")].(*tlc.MethodValue); !ok {
		t.Fatalf("_Counts = %T, want native MethodValue", tool.DefnsByName[tlc.UniqueStringOf("_Counts")])
	}
	for _, name := range []string{"_Track", "_CheckName", "_PrintCounts", "P!_Track", "P!_CheckName", "P!_PrintCounts"} {
		if _, ok := tool.DefnsByName[tlc.UniqueStringOf(name)].(*tlc.OpDefNode); !ok {
			t.Fatalf("%s = %T, want standard TLA OpDefNode", name, tool.DefnsByName[tlc.UniqueStringOf(name)])
		}
	}
}
