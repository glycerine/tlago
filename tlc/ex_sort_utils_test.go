package tlc

import (
	"bytes"
	"testing"
)

func TestExSortUtilsBigIntBehaviors(t *testing.T) {
	values := []*BigInt{
		NewBigIntString("-3"),
		NewBigIntString("0"),
		NewBigIntString("987654321987654321"),
	}

	var sized bytes.Buffer
	if err := WriteSizeArrayOfExternalSortableBigInts(&sized, values, 0, len(values)-1); err != nil {
		t.Fatalf("WriteSizeArrayOfExternalSortableBigInts: %v", err)
	}
	got, err := ReadSizeArrayOfExternalSortableBigInts(bytes.NewReader(sized.Bytes()))
	if err != nil {
		t.Fatalf("ReadSizeArrayOfExternalSortableBigInts: %v", err)
	}
	assertBigIntsEqual(t, got, values)

	var appendedSized bytes.Buffer
	if err := AppendSizeExternalSortableBigIntArraySizeArray(bytes.NewReader(sized.Bytes()), &appendedSized); err != nil {
		t.Fatalf("AppendSizeExternalSortableBigIntArraySizeArray: %v", err)
	}
	got, err = ReadSizeArrayOfExternalSortableBigInts(bytes.NewReader(appendedSized.Bytes()))
	if err != nil {
		t.Fatalf("ReadSizeArrayOfExternalSortableBigInts(appended): %v", err)
	}
	assertBigIntsEqual(t, got, values)

	var appendedBare bytes.Buffer
	if err := AppendSizeExternalSortableBigIntArrayArray(bytes.NewReader(sized.Bytes()), &appendedBare); err != nil {
		t.Fatalf("AppendSizeExternalSortableBigIntArrayArray: %v", err)
	}
	got = ReadArrayOfExternalSortableBigInts(bytes.NewReader(appendedBare.Bytes()))
	assertBigIntsEqual(t, got, values)
}

func TestExSortUtilsBigIntReadErrorsMatchJavaMessages(t *testing.T) {
	if _, err := ReadSizeArrayOfExternalSortableBigInts(bytes.NewReader(nil)); err == nil || err.Error() != "Can't read an array of ExternalSortables from the input stream; it's empty." {
		t.Fatalf("empty stream error = %v", err)
	}

	var truncated bytes.Buffer
	if err := WriteInt(&truncated, 1); err != nil {
		t.Fatalf("WriteInt: %v", err)
	}
	if _, err := ReadSizeArrayOfExternalSortableBigInts(bytes.NewReader(truncated.Bytes())); err == nil || err.Error() != "Can't read an array of ExternalSortables from the input stream; not enough bytes, but not empty." {
		t.Fatalf("truncated stream error = %v", err)
	}
}

func TestExSortUtilsWritesJavaLengthBeforeRangeUse(t *testing.T) {
	var out bytes.Buffer
	if err := WriteSizeArrayOfExternalSortableBigInts(&out, nil, 3, 1); err != nil {
		t.Fatalf("WriteSizeArrayOfExternalSortableBigInts negative range: %v", err)
	}
	got, err := ReadInt(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("ReadInt: %v", err)
	}
	if got != -1 {
		t.Fatalf("written length = %d, want Java finish-start+1 = -1", got)
	}
	if out.Len() != 4 {
		t.Fatalf("bytes written = %d, want only Java int length", out.Len())
	}
}

func assertBigIntsEqual(t *testing.T, got []*BigInt, want []*BigInt) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if !got[i].Equal(want[i]) {
			t.Fatalf("value[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}
