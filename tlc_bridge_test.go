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

func TestCombinatoricsStandardModuleUsesNativeOverrides(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "CombinatoricsBridge.tla")
	writeFile(t, root, `---- MODULE CombinatoricsBridge ----
EXTENDS Combinatorics
C == choose(6, 2) + factorial[3]
====`)

	spec, diags := LoadSanySpec(root, LoadOptions{LibraryPaths: []string{
		filepath.Join("test_vectors", "CommunityModules", "modules"),
	}})
	requireNoErrors(t, diags)
	requireNoErrors(t, CheckSpec(spec))

	tool, toolDiags := BuildTLCTool(spec, tlc.NewModelConfig("CombinatoricsBridge"), tlc.RuntimeParameters{})
	requireNoErrors(t, toolDiags)

	for _, name := range []string{"choose", "factorial"} {
		if _, ok := tool.DefnsByName[tlc.UniqueStringOf(name)].(*tlc.MethodValue); !ok {
			t.Fatalf("%s = %T, want native MethodValue", name, tool.DefnsByName[tlc.UniqueStringOf(name)])
		}
	}
}

func TestBitwiseStandardModuleKeepsLocalHelpersUnexported(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "BitwiseBridge.tla")
	writeFile(t, root, `---- MODULE BitwiseBridge ----
EXTENDS Bitwise
B == INSTANCE Bitwise
C == B!shiftR(8, 1) + B!Not(5) + (6 & 3)
====`)

	spec, diags := LoadSanySpec(root, LoadOptions{LibraryPaths: []string{
		filepath.Join("test_vectors", "CommunityModules", "modules"),
	}})
	requireNoErrors(t, diags)
	requireNoErrors(t, CheckSpec(spec))

	tool, toolDiags := BuildTLCTool(spec, tlc.NewModelConfig("BitwiseBridge"), tlc.RuntimeParameters{})
	requireNoErrors(t, toolDiags)

	for _, name := range []string{"And", "Or", "Xor", "Not", "shiftR", "B!Not", "B!shiftR"} {
		if _, ok := tool.DefnsByName[tlc.UniqueStringOf(name)].(*tlc.MethodValue); !ok {
			t.Fatalf("%s = %T, want native MethodValue", name, tool.DefnsByName[tlc.UniqueStringOf(name)])
		}
	}
	if _, ok := tool.DefnsByName[tlc.UniqueStringOf("B!&")].(*tlc.OpDefNode); !ok {
		t.Fatalf("B!& = %T, want exported TLA OpDefNode", tool.DefnsByName[tlc.UniqueStringOf("B!&")])
	}
	if got := tool.DefnsByName[tlc.UniqueStringOf("B!And")]; got != nil {
		t.Fatalf("B!And = %T, want no alias for LOCAL helper", got)
	}
}

func TestRuntimePostConditionCanTargetQualifiedLocalStandardDefinition(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "TracePostBridge.tla")
	writeFile(t, root, `---- MODULE TracePostBridge ----
CONSTANT K
VARIABLE x
Init == x = 0
Next == x' = x
CfgPost == x = 0
====`)
	cfgPath := filepath.Join(dir, "TracePostBridge.cfg")
	writeFile(t, cfgPath, `INIT Init
NEXT Next
CONSTANT K = "config"
POSTCONDITION CfgPost
`)

	params := tlc.RuntimeParameters{
		PostConditions: []tlc.RuntimePostCondition{{
			Module:       "_TLCTrace",
			Operator:     "_TLCTraceSilent",
			ConstantName: "K",
			FileName:     "runtime",
		}},
	}
	spec, diags := LoadSanySpec(root, LoadOptions{ExtraModules: params.ExtendeeModules()})
	requireNoErrors(t, diags)
	requireNoErrors(t, CheckSpec(spec))

	cfg, err := tlc.ParseModelConfigFile(cfgPath)
	if err != nil {
		t.Fatalf("ParseModelConfigFile: %v", err)
	}
	tool, toolDiags := BuildTLCTool(spec, cfg, params)
	requireNoErrors(t, toolDiags)

	if _, ok := tool.DefnsByName[tlc.UniqueStringOf("_TLCTrace!_TLCTraceSilent")].(*tlc.OpDefNode); !ok {
		t.Fatalf("_TLCTrace!_TLCTraceSilent = %T, want qualified LOCAL OpDefNode", tool.DefnsByName[tlc.UniqueStringOf("_TLCTrace!_TLCTraceSilent")])
	}
	if got := tool.DefnsByName[tlc.UniqueStringOf("_TLCTraceSilent")]; got != nil {
		t.Fatalf("_TLCTraceSilent = %T, want LOCAL helper not exported unqualified", got)
	}
	postConditions := tool.GetPostConditionSpecs()
	if len(postConditions) != 2 {
		t.Fatalf("postconditions = %d, want runtime plus config postcondition", len(postConditions))
	}
	if got := postConditions[0].GetName(); got != "_TLCTraceSilent" {
		t.Fatalf("postconditions[0] = %q, want runtime postcondition first like Java", got)
	}
	if got := postConditions[1].GetName(); got != "CfgPost" {
		t.Fatalf("postconditions[1] = %q, want config postcondition second like Java", got)
	}
	k, ok := tool.DefnsByName[tlc.UniqueStringOf("K")].(*tlc.StringValue)
	if !ok || k.RawString() != "config" {
		t.Fatalf("K = %#v, want config constant to override runtime constant like Java", tool.DefnsByName[tlc.UniqueStringOf("K")])
	}
}
