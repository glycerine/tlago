package tlc

import "testing"

func TestRuntimeParametersStringConstantsIncludeViewConstantsLikeJava(t *testing.T) {
	params := RuntimeParameters{
		Constraints: []RuntimeConstraint{{
			Module:       "C",
			Operator:     "Constraint",
			ConstantName: "CFile",
			FileName:     "constraint.trace",
		}},
		ActionConstraints: []RuntimeConstraint{{
			Module:       "AC",
			Operator:     "ActionConstraint",
			ConstantName: "ACFile",
			FileName:     "action.trace",
		}},
		PostConditions: []RuntimePostCondition{{
			Module:       "PC",
			Operator:     "Post",
			ConstantName: "PCFile",
			FileName:     "post.trace",
		}},
		View: &RuntimeView{
			Module:       "V",
			Operator:     "View",
			ConstantName: "ViewFile",
			FileName:     "view.trace",
		},
	}

	got := params.StringConstants()
	want := []RuntimeStringConstant{
		{Name: "CFile", Value: "constraint.trace"},
		{Name: "ACFile", Value: "action.trace"},
		{Name: "PCFile", Value: "post.trace"},
		{Name: "ViewFile", Value: "view.trace"},
	}
	if len(got) != len(want) {
		t.Fatalf("StringConstants length = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("StringConstants[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
