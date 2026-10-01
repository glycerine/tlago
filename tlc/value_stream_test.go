package tlc

import (
	"bytes"
	"math"
	"testing"
)

func TestValueStreamsRoundTripPrimitiveEncodings(t *testing.T) {
	var buf bytes.Buffer
	out := NewValueOutputStream(&buf)
	if err := out.WriteShort(math.MaxInt16); err != nil {
		t.Fatalf("WriteShort max returned error: %v", err)
	}
	if err := out.WriteShort(math.MinInt16); err != nil {
		t.Fatalf("WriteShort min returned error: %v", err)
	}
	if err := out.WriteShort(0); err != nil {
		t.Fatalf("WriteShort zero returned error: %v", err)
	}
	if err := out.WriteInt(math.MaxInt32); err != nil {
		t.Fatalf("WriteInt max returned error: %v", err)
	}
	if err := out.WriteInt(math.MinInt32); err != nil {
		t.Fatalf("WriteInt min returned error: %v", err)
	}
	if err := out.WriteInt(0); err != nil {
		t.Fatalf("WriteInt zero returned error: %v", err)
	}

	in := NewValueInputStream(bytes.NewReader(buf.Bytes()))
	for i, want := range []int16{math.MaxInt16, math.MinInt16, 0} {
		got, err := in.ReadShort()
		if err != nil || got != want {
			t.Fatalf("ReadShort[%d] = %d/%v, want %d/nil", i, got, err, want)
		}
	}
	for i, want := range []int32{math.MaxInt32, math.MinInt32, 0} {
		got, err := in.ReadInt()
		if err != nil || got != want {
			t.Fatalf("ReadInt[%d] = %d/%v, want %d/nil", i, got, err, want)
		}
	}
}

func TestValueStreamsCompactNaturalEncodings(t *testing.T) {
	var shortBuf bytes.Buffer
	shortOut := NewValueOutputStream(&shortBuf)
	if err := shortOut.WriteShortNat(math.MaxInt16); err != nil {
		t.Fatalf("WriteShortNat max returned error: %v", err)
	}
	if err := shortOut.WriteShortNat(0); err != nil {
		t.Fatalf("WriteShortNat zero returned error: %v", err)
	}
	if got, want := shortBuf.Len(), 3; got != want {
		t.Fatalf("short nat encoded length = %d, want %d", got, want)
	}
	shortIn := NewValueInputStream(bytes.NewReader(shortBuf.Bytes()))
	if got, err := shortIn.ReadShortNat(); err != nil || got != math.MaxInt16 {
		t.Fatalf("ReadShortNat max = %d/%v, want %d/nil", got, err, math.MaxInt16)
	}
	if got, err := shortIn.ReadShortNat(); err != nil || got != 0 {
		t.Fatalf("ReadShortNat zero = %d/%v, want 0/nil", got, err)
	}

	var natBuf bytes.Buffer
	natOut := NewValueOutputStream(&natBuf)
	if err := natOut.WriteNat(math.MaxInt32); err != nil {
		t.Fatalf("WriteNat max returned error: %v", err)
	}
	if err := natOut.WriteNat(0); err != nil {
		t.Fatalf("WriteNat zero returned error: %v", err)
	}
	if got, want := natBuf.Len(), 6; got != want {
		t.Fatalf("nat encoded length = %d, want %d", got, want)
	}
	natIn := NewValueInputStream(bytes.NewReader(natBuf.Bytes()))
	if got, err := natIn.ReadNat(); err != nil || got != math.MaxInt32 {
		t.Fatalf("ReadNat max = %d/%v, want %d/nil", got, err, math.MaxInt32)
	}
	if got, err := natIn.ReadNat(); err != nil || got != 0 {
		t.Fatalf("ReadNat zero = %d/%v, want 0/nil", got, err)
	}
}

func TestValueStreamsGlobalCompressionRoundTrip(t *testing.T) {
	old := UseGZIP()
	SetUseGZIP(true)
	defer SetUseGZIP(old)

	var buf bytes.Buffer
	out := NewValueOutputStreamWithGlobalCompression(&buf)
	if err := out.WriteExternal(NewStringValue("compressed-value-stream")); err != nil {
		t.Fatalf("WriteExternal returned error: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if got := buf.Bytes(); len(got) < 2 || got[0] != 0x1f || got[1] != 0x8b {
		t.Fatalf("compressed stream missing gzip header: %x", got[:min(len(got), 2)])
	}

	in, err := NewValueInputStreamWithGlobalCompression(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("NewValueInputStreamWithGlobalCompression returned error: %v", err)
	}
	value, err := in.ReadExternal()
	if err != nil {
		t.Fatalf("ReadExternal returned error: %v", err)
	}
	if !mustValueEqual(value, NewStringValue("compressed-value-stream")) {
		t.Fatalf("roundtrip value = %v", value)
	}
}

func TestValueOutputStreamWithoutHandlesExpandsRepeatedObjects(t *testing.T) {
	repeated := NewTupleValue([]Value{NewIntValue(1)})

	var shared bytes.Buffer
	sharedOut := NewValueOutputStream(&shared)
	if err := sharedOut.Write(repeated); err != nil {
		t.Fatalf("shared first Write returned error: %v", err)
	}
	if err := sharedOut.Write(repeated); err != nil {
		t.Fatalf("shared second Write returned error: %v", err)
	}
	if !bytes.Contains(shared.Bytes(), []byte{byte(DummyValueKind)}) {
		t.Fatalf("ordinary stream did not emit dummy handle for repeated value: %x", shared.Bytes())
	}

	var expanded bytes.Buffer
	expandedOut := NewValueOutputStreamWithoutHandles(&expanded)
	if err := expandedOut.Write(repeated); err != nil {
		t.Fatalf("expanded first Write returned error: %v", err)
	}
	if err := expandedOut.Write(repeated); err != nil {
		t.Fatalf("expanded second Write returned error: %v", err)
	}
	if bytes.Contains(expanded.Bytes(), []byte{byte(DummyValueKind)}) {
		t.Fatalf("no-handle stream emitted dummy handle: %x", expanded.Bytes())
	}
	if expanded.Len() <= shared.Len() {
		t.Fatalf("expanded stream length = %d, want larger than handle-sharing length %d", expanded.Len(), shared.Len())
	}
}

func TestValueInputStreamBlindReadStringValue(t *testing.T) {
	const text = "Hippopotomonstrosesquippedaliophobia"
	var buf bytes.Buffer
	out := NewValueOutputStream(&buf)
	if err := out.WriteByte(byte(StringValueKind)); err != nil {
		t.Fatalf("WriteByte kind returned error: %v", err)
	}
	writeExternalUniqueString(t, out, text)

	value, err := NewValueInputStream(bytes.NewReader(buf.Bytes())).ReadExternal()
	if err != nil {
		t.Fatalf("ReadExternal returned error: %v", err)
	}
	str, ok := value.(*StringValue)
	if !ok {
		t.Fatalf("ReadExternal returned %T, want *StringValue", value)
	}
	if got := str.Val.String(); got != text {
		t.Fatalf("string value = %q, want %q", got, text)
	}
}

func TestValueInputStreamBlindReadRecordValue(t *testing.T) {
	const key = "Well, let's see, we have on the bags, Who's on first, What's on second, I Don't Know is on third"
	var buf bytes.Buffer
	out := NewValueOutputStream(&buf)
	if err := out.WriteByte(byte(RecordValueKind)); err != nil {
		t.Fatalf("WriteByte record kind returned error: %v", err)
	}
	if err := out.WriteInt(1); err != nil {
		t.Fatalf("WriteInt length returned error: %v", err)
	}
	if err := out.WriteByte(byte(StringValueKind)); err != nil {
		t.Fatalf("WriteByte key kind returned error: %v", err)
	}
	writeExternalUniqueString(t, out, key)
	if err := out.WriteByte(byte(IntValueKind)); err != nil {
		t.Fatalf("WriteByte int kind returned error: %v", err)
	}
	if err := out.WriteInt(42); err != nil {
		t.Fatalf("WriteInt value returned error: %v", err)
	}

	value, err := NewValueInputStream(bytes.NewReader(buf.Bytes())).ReadExternal()
	if err != nil {
		t.Fatalf("ReadExternal returned error: %v", err)
	}
	record, ok := value.(*RecordValue)
	if !ok {
		t.Fatalf("ReadExternal returned %T, want *RecordValue", value)
	}
	got, err := record.Select(NewStringValue(key))
	if err != nil {
		t.Fatalf("record.Select returned error: %v", err)
	}
	if !mustValueEqual(got, NewIntValue(42)) {
		t.Fatalf("record[%q] = %v, want 42", key, got)
	}
}

func TestValueStreamsWriteMaterializedReducibleSetValuesAsJavaSetEnums(t *testing.T) {
	one := NewIntValue(1)
	two := NewIntValue(2)
	s1 := NewSetEnumValue([]Value{one}, true)
	s2 := NewSetEnumValue([]Value{two}, true)

	values := []Value{
		NewSubsetValue(s1),
		NewSetCupValue(s1, s2),
		NewSetCapValue(s1, s1),
		NewSetDiffValue(s1, s2),
		NewUnionValue(NewSetEnumValue([]Value{s1, s2}, true)),
		NewSetOfTuplesValue([]Value{s1, s2}),
		NewSetOfFcnsValue(s1, s2),
		NewKSubsetValue(1, s2),
	}

	rcds, err := NewSetOfRcdsValue([]*UniqueString{UniqueStringOf("x")}, []Value{s1}, true)
	if err != nil {
		t.Fatalf("NewSetOfRcdsValue returned error: %v", err)
	}
	values = append(values, rcds)

	for _, value := range values {
		value.FingerPrint(0)
		expected, err := toSetEnumValue(value)
		if err != nil {
			t.Fatalf("%T toSetEnumValue returned error: %v", value, err)
		}
		expected.DeepNormalize()

		var buf bytes.Buffer
		out := NewValueOutputStream(&buf)
		if err := out.Write(value); err != nil {
			t.Fatalf("%T Write returned error: %v", value, err)
		}
		got, err := NewValueInputStream(bytes.NewReader(buf.Bytes())).Read()
		if err != nil {
			t.Fatalf("%T Read returned error: %v", value, err)
		}
		if !mustValueEqual(got, expected) {
			t.Fatalf("%T roundtrip = %v, want %v", value, got, expected)
		}
	}
}

func TestValueStreamsWriteMaterializedFcnLambdaAsJavaFcnRcd(t *testing.T) {
	fcn := NewFcnRcdValue([]Value{NewStringValue("k")}, []Value{NewIntValue(99)}, true)
	lambda := &FcnLambdaValue{FcnRcd: fcn}

	var buf bytes.Buffer
	out := NewValueOutputStream(&buf)
	if err := out.Write(lambda); err != nil {
		t.Fatalf("Write materialized FcnLambdaValue returned error: %v", err)
	}
	got, err := NewValueInputStream(bytes.NewReader(buf.Bytes())).Read()
	if err != nil {
		t.Fatalf("Read materialized FcnLambdaValue returned error: %v", err)
	}
	if !mustValueEqual(got, fcn) {
		t.Fatalf("materialized FcnLambdaValue roundtrip = %v, want %v", got, fcn)
	}
}

func TestValueStreamsWriteCounterExampleAsJavaRecord(t *testing.T) {
	counterExample := NewEmptyCounterExample()

	var buf bytes.Buffer
	out := NewValueOutputStream(&buf)
	if err := out.Write(counterExample); err != nil {
		t.Fatalf("Write CounterExample returned error: %v", err)
	}
	got, err := NewValueInputStream(bytes.NewReader(buf.Bytes())).Read()
	if err != nil {
		t.Fatalf("Read CounterExample returned error: %v", err)
	}
	if !mustValueEqual(got, counterExample.RecordValue) {
		t.Fatalf("CounterExample roundtrip = %v, want %v", got, counterExample.RecordValue)
	}
}

func writeExternalUniqueString(t *testing.T, out *ValueOutputStream, value string) {
	t.Helper()
	if err := out.WriteInt(-1); err != nil {
		t.Fatalf("WriteInt stale token returned error: %v", err)
	}
	if err := out.WriteInt(-1); err != nil {
		t.Fatalf("WriteInt stale loc returned error: %v", err)
	}
	if err := out.WriteInt(int32(len([]byte(value)))); err != nil {
		t.Fatalf("WriteInt string length returned error: %v", err)
	}
	if _, err := out.WriteRaw([]byte(value)); err != nil {
		t.Fatalf("WriteRaw string returned error: %v", err)
	}
}
