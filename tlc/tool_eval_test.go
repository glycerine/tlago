package tlc

import "testing"

func TestToolExceptWarnsForMissingFieldsWithoutEvaluatingReplacement(t *testing.T) {
	ClearMessageRecorders()
	t.Cleanup(ClearMessageRecorders)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	tool := NewTool()
	record := NewRecordValue([]*UniqueString{UniqueStringOf("present")}, []Value{NewIntValue(1)}, true)
	base := NewValueNode(record)
	base.Image = "r"
	path := NewBuiltinOpApplNode(OpTup, NewStringNode("missing"))
	// An undefined replacement would fail if evaluated. Java warns and skips it.
	rhs := NewOpApplNode(NewSymbolNode("UndefinedReplacement"))
	pair := NewBuiltinOpApplNode(OpPair, path, rhs)
	expr := NewBuiltinOpApplNode(OpExc, base, pair, pair)
	result, err := tool.Eval(expr)
	if err != nil || result != record {
		t.Fatalf("EXCEPT = %v, %v; want original record", result, err)
	}
	warnings := recorder.Records(ECTLCExceptAppliedToUnknownField)
	if len(warnings) != 2 {
		t.Fatalf("missing-field warnings = %d, want one for each update", len(warnings))
	}
	for _, warning := range warnings {
		if warning.Severity != SeverityWarning || len(warning.Params) != 1 || warning.Params[0] != "r" {
			t.Fatalf("warning = %#v, want Java's warning with the original base expression", warning)
		}
	}
}

func TestToolExceptKeepsSequentialAtValuesAndPairPathCoverage(t *testing.T) {
	Globals.Lock()
	oldCoverage := Globals.CoverageInterval
	Globals.CoverageInterval = 0
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.CoverageInterval = oldCoverage
		Globals.Unlock()
	})
	tool := NewTool()
	base := NewTupleValue([]Value{NewIntValue(7)})
	args := []SemanticNode{NewValueNode(base)}
	indexes := make([]*OpApplNode, 2)
	paths := make([]*OpApplNode, 2)
	replacements := make([]*OpApplNode, 2)
	pairs := make([]*OpApplNode, 2)
	index := tool.DefineName("Index", NewIntValue(1))
	for i := range pairs {
		indexes[i] = NewOpApplNode(index)
		paths[i] = NewBuiltinOpApplNode(OpTup, indexes[i])
		replacements[i] = NewBuiltinOpApplNode(OpPlus, NewAtNode(), NewValueNode(NewIntValue(1)))
		pairs[i] = NewBuiltinOpApplNode(OpPair, paths[i], replacements[i])
		args = append(args, pairs[i])
	}
	expr := NewBuiltinOpApplNode(OpExc, args...)
	cm := NewCostModel(expr)
	indexCosts := make([]CostModel, 2)
	replacementCosts := make([]CostModel, 2)
	for i, pair := range pairs {
		pairCM := cm.AddChild(pair)
		indexCosts[i] = pairCM.AddChild(paths[i]).AddChild(indexes[i])
		replacementCosts[i] = pairCM.AddChild(replacements[i])
	}
	result, err := tool.Eval(expr, EmptyContext, EmptyState, EmptyState, EvalClear, cm)
	if err != nil {
		t.Fatal(err)
	}
	want := NewTupleValue([]Value{NewIntValue(9)})
	if equal, err := result.Equal(want); err != nil || !equal {
		t.Fatalf("EXCEPT = %v, want %v; equality error: %v", result, want, err)
	}
	if base.Elems[0].(*IntValue).Val != 7 {
		t.Fatal("EXCEPT mutated the original tuple")
	}
	for i := range pairs {
		if got := indexCosts[i].GetPrimary(); got != 1 {
			t.Errorf("update %d path expression count = %d, want 1", i, got)
		}
		if got := replacementCosts[i].GetPrimary(); got != 1 {
			t.Errorf("update %d replacement expression count = %d, want 1", i, got)
		}
	}
	if got := cm.GetPrimary(); got != 1 {
		t.Errorf("EXCEPT expression count = %d, want 1", got)
	}
}

func TestToolRecordConstructorsUseFieldPairCoverage(t *testing.T) {
	Globals.Lock()
	oldCoverage := Globals.CoverageInterval
	Globals.CoverageInterval = 0
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.CoverageInterval = oldCoverage
		Globals.Unlock()
	})
	for _, recordSet := range []bool{false, true} {
		tool := NewTool()
		field := NewBuiltinOpApplNode(OpPlus, NewValueNode(NewIntValue(1)), NewValueNode(NewIntValue(2)))
		var fieldExpr SemanticNode = field
		op := OpRC
		if recordSet {
			op = OpSOR
			fieldExpr = NewBuiltinOpApplNode(OpSE, field)
		}
		pair := NewBuiltinOpApplNode(OpPair, NewStringNode("a"), fieldExpr)
		expr := NewBuiltinOpApplNode(op, pair)
		cm := NewCostModel(expr)
		fieldCM := cm.AddChild(pair).AddChild(fieldExpr)
		if recordSet {
			fieldCM = fieldCM.AddChild(field)
		}
		result, err := tool.Eval(expr, EmptyContext, EmptyState, EmptyState, EvalClear, cm)
		if err != nil {
			t.Fatal(err)
		}
		want := NewRecordValue([]*UniqueString{UniqueStringOf("a")}, []Value{NewIntValue(3)}, true)
		if recordSet {
			if member, err := result.Member(want); err != nil || !member {
				t.Fatalf("record-set result = %v, want member %v; error: %v", result, want, err)
			}
		} else if equal, err := result.Equal(want); err != nil || !equal {
			t.Fatalf("record result = %v, want %v; error: %v", result, want, err)
		}
		if got := fieldCM.GetPrimary(); got != 1 {
			t.Errorf("recordSet=%v: field expression count = %d, want 1", recordSet, got)
		}
		if got := cm.GetPrimary(); got != 1 {
			t.Errorf("recordSet=%v: constructor count = %d, want 1", recordSet, got)
		}
	}
}
