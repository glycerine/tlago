package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestParseLivenessFromParsedSANYSpec(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "LivenessParser.tla")
	writeFile(t, root, `---- MODULE LivenessParser ----
VARIABLE x
Init == x = 0
Next == x' = x
Spec == Init /\ [][Next]_x /\ WF_x(Next)
Prop == []<>(x = 0)
====`)

	spec, diags := LoadSanySpec(root, LoadOptions{})
	requireNoErrors(t, diags)
	requireNoErrors(t, CheckSpec(spec))

	cfg, err := tlc.ParseModelConfigSource(filepath.Join(dir, "LivenessParser.cfg"), `SPECIFICATION Spec
PROPERTY Prop
`)
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}
	tool, toolDiags := BuildTLCTool(spec, cfg, tlc.RuntimeParameters{})
	requireNoErrors(t, toolDiags)

	if got := len(tool.GetTemporals()); got != 1 {
		t.Fatalf("temporals = %d, want one WF conjunct from SPECIFICATION", got)
	}
	if got := len(tool.GetImpliedTemporals()); got != 1 {
		t.Fatalf("implied temporals = %d, want one PROPERTY", got)
	}

	live, err := tlc.ParseLiveness(tool)
	if err != nil {
		t.Fatalf("ParseLiveness returned error: %v", err)
	}
	if live == nil || live.Kind != tlc.LiveExprConj || live.Count() != 2 {
		t.Fatalf("ParseLiveness = %#v, want fairness /\\ negated property", live)
	}
	fair := live.GetBody(0)
	if fair == nil || fair.Kind != tlc.LiveExprAll || fair.Body == nil || fair.Body.Kind != tlc.LiveExprEven {
		t.Fatalf("fairness conjunct = %#v, want WF expansion headed by []<>", fair)
	}
	property := live.GetBody(1)
	if property == nil || property.Kind != tlc.LiveExprNeg {
		t.Fatalf("property conjunct = %#v, want negated configured property", property)
	}
	body := property.Body
	if body == nil || body.Kind != tlc.LiveExprAll || body.Body == nil || body.Body.Kind != tlc.LiveExprEven {
		t.Fatalf("property body = %#v, want parsed []<> property", body)
	}
}
