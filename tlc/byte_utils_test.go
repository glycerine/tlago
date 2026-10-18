package tlc

import (
	"bytes"
	"io"
	"math"
	"math/big"
	"testing"
)

func TestByteUtilsIntEncodingMatchesJavaBigEndian(t *testing.T) {
	cases := []int32{
		0,
		1,
		-1,
		math.MaxInt32,
		math.MinInt32,
		0x01020304,
	}
	for _, value := range cases {
		encoded := IntToByteArray(value)
		if len(encoded) != 4 {
			t.Fatalf("IntToByteArray(%d) length = %d, want 4", value, len(encoded))
		}
		if decoded := ByteArrayToInt(encoded); decoded != value {
			t.Fatalf("ByteArrayToInt(IntToByteArray(%d)) = %d", value, decoded)
		}
	}
	if got := IntToByteArray(0x01020304); !bytes.Equal(got, []byte{1, 2, 3, 4}) {
		t.Fatalf("IntToByteArray byte order = %v, want [1 2 3 4]", got)
	}
	if got := ByteArrayToInt([]byte{0xff, 0xff, 0xff, 0xff}); got != -1 {
		t.Fatalf("ByteArrayToInt(-1 bytes) = %d, want -1", got)
	}
}

func TestByteUtilsLongEncodingMatchesJavaBigEndian(t *testing.T) {
	cases := []int64{
		0,
		1,
		-1,
		math.MaxInt64,
		math.MinInt64,
		0x0102030405060708,
	}
	for _, value := range cases {
		encoded := LongToByteArray(value)
		if len(encoded) != 8 {
			t.Fatalf("LongToByteArray(%d) length = %d, want 8", value, len(encoded))
		}
		if decoded := ByteArrayToLong(encoded); decoded != value {
			t.Fatalf("ByteArrayToLong(LongToByteArray(%d)) = %d", value, decoded)
		}
	}
	if got := LongToByteArray(0x0102030405060708); !bytes.Equal(got, []byte{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatalf("LongToByteArray byte order = %v, want [1 2 3 4 5 6 7 8]", got)
	}
	if got := ByteArrayToLong([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}); got != -1 {
		t.Fatalf("ByteArrayToLong(-1 bytes) = %d, want -1", got)
	}
}

func TestByteUtilsWriteAndReadPrimitiveValues(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteInt(&buf, -1234567); err != nil {
		t.Fatalf("WriteInt returned error: %v", err)
	}
	if err := WriteLong(&buf, -987654321012345678); err != nil {
		t.Fatalf("WriteLong returned error: %v", err)
	}
	gotInt, err := ReadInt(&buf)
	if err != nil {
		t.Fatalf("ReadInt returned error: %v", err)
	}
	if gotInt != -1234567 {
		t.Fatalf("ReadInt = %d, want -1234567", gotInt)
	}
	gotLong, err := ReadLong(&buf)
	if err != nil {
		t.Fatalf("ReadLong returned error: %v", err)
	}
	if gotLong != -987654321012345678 {
		t.Fatalf("ReadLong = %d, want -987654321012345678", gotLong)
	}
}

func TestByteUtilsByteArrayPaddingMatchesJavaSignExtension(t *testing.T) {
	positive, err := ByteArrayToByteArray([]byte{0x7f}, 4)
	if err != nil {
		t.Fatalf("ByteArrayToByteArray(positive) returned error: %v", err)
	}
	if !bytes.Equal(positive, []byte{0, 0, 0, 0x7f}) {
		t.Fatalf("positive padding = %v, want [0 0 0 127]", positive)
	}
	negative, err := ByteArrayToByteArray([]byte{0x80}, 4)
	if err != nil {
		t.Fatalf("ByteArrayToByteArray(negative) returned error: %v", err)
	}
	if !bytes.Equal(negative, []byte{0xff, 0xff, 0xff, 0x80}) {
		t.Fatalf("negative padding = %v, want [255 255 255 128]", negative)
	}
	if _, err := ByteArrayToByteArray([]byte{1, 2, 3, 4, 5}, 4); err == nil {
		t.Fatalf("ByteArrayToByteArray too-large input returned nil error")
	}
}

func TestByteUtilsBigIntEncodingUsesJavaTwosComplementBytes(t *testing.T) {
	cases := []struct {
		value string
		bytes []byte
	}{
		{"0", []byte{0x00}},
		{"1", []byte{0x01}},
		{"127", []byte{0x7f}},
		{"128", []byte{0x00, 0x80}},
		{"255", []byte{0x00, 0xff}},
		{"256", []byte{0x01, 0x00}},
		{"-1", []byte{0xff}},
		{"-2", []byte{0xfe}},
		{"-128", []byte{0x80}},
		{"-129", []byte{0xff, 0x7f}},
		{"-256", []byte{0xff, 0x00}},
		{"-257", []byte{0xfe, 0xff}},
	}
	for _, tc := range cases {
		value, ok := new(big.Int).SetString(tc.value, 10)
		if !ok {
			t.Fatalf("bad test integer %q", tc.value)
		}
		if got := BigIntToJavaBytes(value); !bytes.Equal(got, tc.bytes) {
			t.Fatalf("BigIntToJavaBytes(%s) = %v, want %v", tc.value, got, tc.bytes)
		}
		roundTrip, err := JavaBytesToBigInt(tc.bytes)
		if err != nil {
			t.Fatalf("JavaBytesToBigInt(%v) returned error: %v", tc.bytes, err)
		}
		if roundTrip.Cmp(value) != 0 {
			t.Fatalf("JavaBytesToBigInt(%v) = %s, want %s", tc.bytes, roundTrip, value)
		}
	}
}

func TestByteUtilsBigIntReadWriteRoundTripWithoutWrapper(t *testing.T) {
	values := []*big.Int{
		big.NewInt(0),
		big.NewInt(128),
		big.NewInt(-129),
		new(big.Int).Lsh(big.NewInt(1), 90),
		new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 90)),
	}
	var buf bytes.Buffer
	if err := WriteSizeArrayOfSizeBigInts(&buf, values, 0, len(values)-1); err != nil {
		t.Fatalf("WriteSizeArrayOfSizeBigInts returned error: %v", err)
	}
	got, err := ReadSizeArrayOfSizeBigInts(&buf)
	if err != nil {
		t.Fatalf("ReadSizeArrayOfSizeBigInts returned error: %v", err)
	}
	if len(got) != len(values) {
		t.Fatalf("round trip length = %d, want %d", len(got), len(values))
	}
	for i := range values {
		if got[i].Cmp(values[i]) != 0 {
			t.Fatalf("round trip[%d] = %s, want %s", i, got[i], values[i])
		}
	}

	padded, err := BigIntToByteArray(big.NewInt(-1), 4)
	if err != nil {
		t.Fatalf("BigIntToByteArray(-1, 4) returned error: %v", err)
	}
	if !bytes.Equal(padded, []byte{0xff, 0xff, 0xff, 0xff}) {
		t.Fatalf("BigIntToByteArray(-1, 4) = %v, want [255 255 255 255]", padded)
	}
}

func TestByteUtilsReadIntoConsumesShortReadsLikeJava(t *testing.T) {
	reader := &shortReader{chunks: [][]byte{{1}, {2, 3}, {4, 5}}}
	buf := make([]byte, 4)
	n, err := ReadInto(reader, buf, 0, len(buf))
	if err != nil {
		t.Fatalf("ReadInto returned error: %v", err)
	}
	if n != 4 || !bytes.Equal(buf, []byte{1, 2, 3, 4}) {
		t.Fatalf("ReadInto read %d/%v, want 4/[1 2 3 4]", n, buf)
	}
}

func TestByteUtilsReadPrimitiveErrorsMatchJavaCases(t *testing.T) {
	if _, err := ReadInt(bytes.NewReader(nil)); err == nil || err.Error() != "readInt: the input stream is empty." {
		t.Fatalf("ReadInt(empty) error = %v", err)
	}
	if _, err := ReadInt(bytes.NewReader([]byte{1, 2, 3})); err == nil || err.Error() != "readInt: not enought bytes." {
		t.Fatalf("ReadInt(short) error = %v", err)
	}
	if _, err := ReadLong(bytes.NewReader(nil)); err == nil || err.Error() != "readLong: the imput stream is empty." {
		t.Fatalf("ReadLong(empty) error = %v", err)
	}
	if _, err := ReadLong(bytes.NewReader([]byte{1, 2, 3})); err == nil || err.Error() != "readLong: not enought bytes." {
		t.Fatalf("ReadLong(short) error = %v", err)
	}
}

func TestByteUtilsSizeBigIntErrorsAndHexMatchJava(t *testing.T) {
	var truncated bytes.Buffer
	if err := WriteInt(&truncated, 2); err != nil {
		t.Fatalf("WriteInt: %v", err)
	}
	truncated.WriteByte(0xff)
	if _, err := ReadSizeBigInt(bytes.NewReader(truncated.Bytes())); err == nil || err.Error() != "readSizeBigInt: not enough bytes." {
		t.Fatalf("ReadSizeBigInt(short) error = %v", err)
	}
	if got := PrintHex([]byte{0x00, 0x01, 0x0f, 0x10, 0x80, 0xff}); got != "01f1080ff" {
		t.Fatalf("PrintHex = %q, want 01f1080ff", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatalf("ReadSizeByteArray(negative length) did not panic like Java")
		}
	}()
	_, _ = ReadSizeByteArray(bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff}))
}

func TestByteUtilsSizeByteArrayRoundTripAndAppend(t *testing.T) {
	var src bytes.Buffer
	first := []byte{1, 2, 3}
	second := []byte{4, 5}
	if err := WriteSizeByteArray(&src, first); err != nil {
		t.Fatalf("WriteSizeByteArray(first) returned error: %v", err)
	}
	if err := WriteSizeByteArray(&src, second); err != nil {
		t.Fatalf("WriteSizeByteArray(second) returned error: %v", err)
	}

	gotFirst, err := ReadSizeByteArray(&src)
	if err != nil {
		t.Fatalf("ReadSizeByteArray(first) returned error: %v", err)
	}
	if !bytes.Equal(gotFirst, first) {
		t.Fatalf("first bytes = %v, want %v", gotFirst, first)
	}
	gotSecond, err := ReadSizeByteArray(&src)
	if err != nil {
		t.Fatalf("ReadSizeByteArray(second) returned error: %v", err)
	}
	if !bytes.Equal(gotSecond, second) {
		t.Fatalf("second bytes = %v, want %v", gotSecond, second)
	}

	var counted bytes.Buffer
	if err := WriteInt(&counted, 2); err != nil {
		t.Fatalf("WriteInt(count) returned error: %v", err)
	}
	if err := WriteSizeByteArray(&counted, first); err != nil {
		t.Fatalf("WriteSizeByteArray(first counted) returned error: %v", err)
	}
	if err := WriteSizeByteArray(&counted, second); err != nil {
		t.Fatalf("WriteSizeByteArray(second counted) returned error: %v", err)
	}
	var appended bytes.Buffer
	if err := AppendSizeByteArray(&counted, &appended); err != nil {
		t.Fatalf("AppendSizeByteArray returned error: %v", err)
	}
	if !bytes.Equal(appended.Bytes(), append(IntToByteArray(2), appendSizeByteArrayBytes(first, second)...)) {
		t.Fatalf("appended bytes = %v", appended.Bytes())
	}
}

func appendSizeByteArrayBytes(values ...[]byte) []byte {
	var buf bytes.Buffer
	for _, value := range values {
		_ = WriteSizeByteArray(&buf, value)
	}
	return buf.Bytes()
}

type shortReader struct {
	chunks [][]byte
}

func (r *shortReader) Read(p []byte) (int, error) {
	if len(r.chunks) == 0 {
		return 0, io.EOF
	}
	chunk := r.chunks[0]
	r.chunks = r.chunks[1:]
	n := copy(p, chunk)
	if n < len(chunk) {
		r.chunks = append([][]byte{chunk[n:]}, r.chunks...)
	}
	return n, nil
}
