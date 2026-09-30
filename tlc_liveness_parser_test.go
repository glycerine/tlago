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

func TestParseLivenessFromParsedSANYSpecWithTemporalOperators(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "LivenessOperators.tla")
	writeFile(t, root, `---- MODULE LivenessOperators ----
VARIABLE x
Init == x = 0
Next == x' = x
Spec == Init /\ [][Next]_x /\ SF_x(Next)
Prop ==
  /\ (x = 0) ~> (x = 1)
  /\ []<>(x = 0)
  /\ <>[](x = 1)
====`)

	spec, diags := LoadSanySpec(root, LoadOptions{})
	requireNoErrors(t, diags)
	requireNoErrors(t, CheckSpec(spec))

	cfg, err := tlc.ParseModelConfigSource(filepath.Join(dir, "LivenessOperators.cfg"), `SPECIFICATION Spec
PROPERTY Prop
`)
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}
	tool, toolDiags := BuildTLCTool(spec, cfg, tlc.RuntimeParameters{})
	requireNoErrors(t, toolDiags)
	if got := len(tool.GetTemporals()); got != 1 {
		t.Fatalf("temporals = %d, want one SF conjunct from SPECIFICATION", got)
	}
	if got := len(tool.GetImpliedTemporals()); got != 3 {
		t.Fatalf("implied temporals = %d, want PROPERTY conjunction split into three checks", got)
	}

	live, err := tlc.ParseLiveness(tool)
	if err != nil {
		t.Fatalf("ParseLiveness returned error: %v", err)
	}
	if live == nil || live.Kind != tlc.LiveExprConj || live.Count() != 2 {
		t.Fatalf("ParseLiveness = %#v, want SF fairness /\\ negated property", live)
	}

	fair := live.GetBody(0)
	if fair == nil || fair.Kind != tlc.LiveExprDisj || fair.Count() != 2 {
		t.Fatalf("fairness conjunct = %#v, want SF expansion disjunction", fair)
	}
	eventuallyAlwaysEnabled := fair.GetBody(0)
	if eventuallyAlwaysEnabled == nil || eventuallyAlwaysEnabled.Kind != tlc.LiveExprEven ||
		eventuallyAlwaysEnabled.Body == nil || eventuallyAlwaysEnabled.Body.Kind != tlc.LiveExprAll {
		t.Fatalf("first SF disjunct = %#v, want <>[] enabled branch", eventuallyAlwaysEnabled)
	}
	alwaysEventuallyAction := fair.GetBody(1)
	if alwaysEventuallyAction == nil || alwaysEventuallyAction.Kind != tlc.LiveExprAll ||
		alwaysEventuallyAction.Body == nil || alwaysEventuallyAction.Body.Kind != tlc.LiveExprEven {
		t.Fatalf("second SF disjunct = %#v, want []<> action branch", alwaysEventuallyAction)
	}

	property := live.GetBody(1)
	if property == nil || property.Kind != tlc.LiveExprDisj || property.Count() != 3 {
		t.Fatalf("property conjunct = %#v, want disjunction of negated configured properties", property)
	}
	for i := 0; i < property.Count(); i++ {
		if property.GetBody(i) == nil || property.GetBody(i).Kind != tlc.LiveExprNeg {
			t.Fatalf("property disjunct %d = %#v, want negated split PROPERTY check", i, property.GetBody(i))
		}
	}
	leadsto := property.GetBody(0).Body
	if leadsto == nil || leadsto.Kind != tlc.LiveExprAll ||
		leadsto.Body == nil || leadsto.Body.Kind != tlc.LiveExprDisj ||
		leadsto.Body.Count() != 2 || leadsto.Body.GetBody(0).Kind != tlc.LiveExprNeg ||
		leadsto.Body.GetBody(1).Kind != tlc.LiveExprEven {
		t.Fatalf("leadsto body = %#v, want [](~left \\/ <>right)", leadsto)
	}
	alwaysEventually := property.GetBody(1).Body
	if alwaysEventually == nil || alwaysEventually.Kind != tlc.LiveExprAll ||
		alwaysEventually.Body == nil || alwaysEventually.Body.Kind != tlc.LiveExprEven {
		t.Fatalf("second property conjunct = %#v, want []<>", alwaysEventually)
	}
	eventuallyAlways := property.GetBody(2).Body
	if eventuallyAlways == nil || eventuallyAlways.Kind != tlc.LiveExprEven ||
		eventuallyAlways.Body == nil || eventuallyAlways.Body.Kind != tlc.LiveExprAll {
		t.Fatalf("third property conjunct = %#v, want <>[]", eventuallyAlways)
	}
}
