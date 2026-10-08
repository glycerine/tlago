package tlc

import (
	"bytes"
	"encoding/gob"
	"math"
	"reflect"
	"testing"
)

// Upstream has no direct network object-graph test. These short unit checks
// cover the native Go payload, not the separate on-disk value stream.
func distributedPayloadRoundTrip(t *testing.T, states []*TLCStateMut) []*TLCStateMut {
	t.Helper()
	payload, err := EncodeDistributedStates(states)
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(payload); err != nil {
		t.Fatal(err)
	}
	var wire DistributedStatePayload
	if err := gob.NewDecoder(&buffer).Decode(&wire); err != nil {
		t.Fatal(err)
	}
	got, err := DecodeDistributedStates(&wire)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestDistributedStatePayloadSharingAndDepth(t *testing.T) {
	name := &UniqueString{s: "A_λ\x00", tok: 731, loc: 19}
	text := &StringValue{Val: name}
	tuple := NewTupleValue([]Value{NewIntValue(math.MinInt32), NewIntValue(math.MaxInt32), text})
	record := NewRecordValue([]*UniqueString{name}, []Value{tuple}, false)
	model := &ModelValue{Val: name, Index: 7, Type: 'A', Data: tuple}
	first := &TLCStateMut{WorkerID: math.MaxInt16, UID: -1, level: math.MaxInt32, values: []Value{BoolTrue, tuple, tuple, record, model}}
	second := &TLCStateMut{WorkerID: 12, UID: math.MaxInt64, level: 40000, values: []Value{tuple, nil, ValUndef}}
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{first, first, nil, second})
	if len(got) != 4 || got[0] != got[1] || got[0] == first || got[2] != nil || got[3] == second {
		t.Fatal("state identity or null state was lost")
	}
	if got[0].UID != -1 || got[0].WorkerID != math.MaxInt16 || got[0].Level() != math.MaxInt32 || got[3].UID != math.MaxInt64 || got[3].Level() != 40000 {
		t.Fatal("state metadata inherited the disk depth/identifier limits")
	}
	copied := got[0].values[1].(*TupleValue)
	if copied == tuple || got[0].values[2] != copied || got[3].values[0] != copied {
		t.Fatal("cross-state value sharing or ownership was lost")
	}
	copiedRecord := got[0].values[3].(*RecordValue)
	copiedModel := got[0].values[4].(*ModelValue)
	copiedName := copied.Elems[2].(*StringValue).Val
	if copiedName == name || copiedRecord.Names[0] != copiedName || copiedModel.Val != copiedName || copiedName.s != name.s || copiedName.tok != 731 || copiedName.loc != 19 {
		t.Fatal("interned string metadata or sharing was lost")
	}
	if copiedRecord.IsNorm || copiedModel.Index != 7 || copiedModel.Type != 'A' || copiedModel.Data != copied {
		t.Fatal("record/model fields were lost")
	}
	if got[3].values[1] != nil {
		t.Fatal("incomplete state value was filled in")
	}
	if original, copiedFP := first.FingerPrintWithTool(nil), got[0].FingerPrintWithTool(nil); original != copiedFP {
		t.Fatalf("fingerprint = %x, want %x", copiedFP, original)
	}
	copiedName.loc = 80
	if name.loc != 19 {
		t.Fatal("receiver metadata aliases the sender")
	}
}

func TestDistributedStatePayloadRepresentations(t *testing.T) {
	small := NewSetEnumValue([]Value{NewIntValue(2), NewIntValue(1), NewIntValue(1)}, false)
	name := &UniqueString{s: "field", tok: 3, loc: -1}
	recordSet, err := NewSetOfRcdsValue([]*UniqueString{name}, []Value{small}, true)
	if err != nil {
		t.Fatal(err)
	}
	subset := NewSubsetValue(small)
	subset.PSetDummy = true
	values := []Value{
		small, NewIntervalValue(-3, 6), NewTupleValue(nil), NewTupleValue([]Value{}),
		NewFcnRcdValue([]Value{NewStringValue("key")}, []Value{small}, false),
		NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{BoolTrue, BoolFalse}),
		NewSetOfTuplesValue([]Value{small, small}), recordSet,
		NewSetOfFcnsValue(small, NewIntervalValue(0, 1)), subset, NewKSubsetValue(1, small),
		NewSetCupValue(small, small), NewSetCapValue(small, small), NewSetDiffValue(small, small),
		NewUnionValue(NewSetEnumValue([]Value{small}, true)),
		NatValue, IntValueSet, StringSetValue, AnySetValue,
		NewUserValue(&sequencesObj{Range: small, SizeBound: 7}),
	}
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: values}})[0].values
	for i, original := range values {
		if reflect.TypeOf(original) != reflect.TypeOf(got[i]) || original == got[i] {
			t.Fatalf("value %d representation/ownership = %T/%T", i, original, got[i])
		}
	}
	set := got[0].(*SetEnumValue)
	if set.IsNorm || set.Elems.Len() != 3 || set.Elems.At(0).(*IntValue).Val != 2 {
		t.Fatal("transport normalized or deduplicated a set")
	}
	if got[2].(*TupleValue).Elems != nil || got[3].(*TupleValue).Elems == nil {
		t.Fatal("null and empty value arrays were collapsed")
	}
	if got[6].(*SetOfTuplesValue).Sets[0] != set || got[6].(*SetOfTuplesValue).Sets[1] != set {
		t.Fatal("symbolic constructor children lost sharing")
	}
	if !got[9].(*SubsetValue).PSetDummy || got[9].(*SubsetValue).PSet != nil {
		t.Fatal("dummy set cache was forced or discarded")
	}
	if got[19].(*UserValue).UserObj.(*sequencesObj).SizeBound != 7 {
		t.Fatal("sequence bound changed")
	}
	if got[19].(*UserValue).UserObj.(*sequencesObj).Range != set {
		t.Fatal("sequence range lost sharing")
	}
}

func TestDistributedStatePayloadMaterialization(t *testing.T) {
	bound := NewSymbolNode("payloadParameter")
	tool := NewTool()
	lambda := NewFcnLambdaValue(NewSingleFcnParam(bound, NewIntervalValue(1, 2)), NewValueNode(NewIntValue(42)), tool, EmptyContext, EmptyState, nil, EvalClear)
	predicate := NewSetPredValue(bound, NewIntervalValue(1, 2), NewValueNode(BoolTrue), tool, EmptyContext, EmptyState, nil, EvalClear)
	lazy := &LazyValue{Val: NewIntValue(8)}
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{lambda, predicate, lazy}}})[0].values
	decoded := got[0].(*FcnLambdaValue)
	if lambda.FcnRcd == nil || decoded.FcnRcd == nil || decoded.Body != nil || decoded.Tool != nil {
		t.Fatal("function was not materialized or evaluator machinery was transported")
	}
	if value, err := decoded.Apply(NewIntValue(2)); err != nil || value.(*IntValue).Val != 42 {
		t.Fatalf("decoded function application = %v/%v", value, err)
	}
	pred := got[1].(*SetPredValue)
	if !predicate.Converted || !pred.Converted || pred.Tool != nil || pred.Pred != nil {
		t.Fatal("predicate was not converted or evaluator machinery was transported")
	}
	if member, err := pred.Member(NewIntValue(1)); err != nil || !member {
		t.Fatalf("decoded predicate membership = %v/%v", member, err)
	}
	if got[2].(*LazyValue).Val.(*IntValue).Val != 8 {
		t.Fatal("cached lazy value changed")
	}
	for _, unevaluated := range []Value{&LazyValue{}, &LazyValue{Val: ValUndef}} {
		if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{unevaluated}}}); err == nil {
			t.Fatal("unevaluated lazy value accepted")
		}
	}
}

func TestDistributedStatePayloadNullsCyclesAndValidation(t *testing.T) {
	if got := distributedPayloadRoundTrip(t, nil); got != nil {
		t.Fatal("null state array became empty")
	}
	if got := distributedPayloadRoundTrip(t, []*TLCStateMut{}); got == nil || len(got) != 0 {
		t.Fatal("empty state array became null")
	}
	cycle := NewTupleValue(make([]Value, 1))
	cycle.Elems[0] = cycle
	copied := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{cycle}}})[0].values[0].(*TupleValue)
	if copied == cycle || copied.Elems[0] != copied {
		t.Fatal("object-graph cycle lost identity")
	}
	for _, payload := range []*DistributedStatePayload{
		nil,
		{Roots: []int{1}},
		{Nil: true, Roots: []int{0}},
		{States: []DistributedStateNode{{Level: -1}}},
		{States: []DistributedStateNode{{Values: []int{1}}}},
		{Values: []DistributedValueNode{{Kind: "unknown"}}},
		{Values: []DistributedValueNode{{Kind: "int", Integer: math.MaxInt64}}},
		{Values: []DistributedValueNode{{Kind: "lambda", References: []int{0, 0}}}},
		{Values: []DistributedValueNode{{Kind: "string", String: 1}}},
		{Values: []DistributedValueNode{{Kind: "cup", References: []int{0, 0}, Cache: 1}}},
	} {
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid payload accepted: %#v", payload)
		}
	}
	if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, action: &Action{}}}); err == nil {
		t.Fatal("extended state metadata silently discarded")
	}
}

func TestDistributedStatePayloadModelData(t *testing.T) {
	for _, data := range []any{nil, "payload", true, int(4), int32(-5), int64(math.MaxInt64), float64(0.5), []byte{}, []byte(nil), []byte{0, 255}} {
		model := &ModelValue{Val: &UniqueString{s: "Model", tok: 3, loc: -1}, Index: 2, Data: data}
		copied := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{model}}})[0].values[0].(*ModelValue)
		if !reflect.DeepEqual(data, copied.Data) {
			t.Fatalf("model data = %#v, want %#v", copied.Data, data)
		}
	}
	model := &ModelValue{Data: struct{ Name string }{"opaque"}}
	if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{model}}}); err == nil {
		t.Fatal("opaque model data was silently discarded")
	}
}
