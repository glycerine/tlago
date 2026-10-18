// Copyright (c) 2026 NVIDIA Corp. All rights reserved.
package tlc

import (
	"bytes"
	"testing"
)

func javaBufferedDataInput(t *testing.T, data []byte) *BufferedDataInputStream {
	t.Helper()
	stream, err := NewBufferedDataInputStream(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if stream.in != nil {
			if err := stream.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	return stream
}

// Whole original util.BufferedDataInputStreamTest, all seven original methods.
func TestJavaBufferedDataInputStream(t *testing.T) {
	for _, input := range []struct {
		name, data string
		n          int
	}{
		{"testReadStringThrowsOnShortStreamAtHalfBoundary", "ABCDEFGHIJ", 15},
		{"testReadStringThrowsOnShortStreamBelowHalfBoundary", "AB", 10},
	} {
		t.Run(input.name, func(t *testing.T) {
			stream := javaBufferedDataInput(t, []byte(input.data))
			_, err := stream.ReadString(input.n)
			if _, ok := err.(*EOFException); !ok {
				t.Fatalf("readString(%d) should throw EOFException, got %T: %v", input.n, err, err)
			}
			if err := stream.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("testReadStringExactLength", func(t *testing.T) {
		stream := javaBufferedDataInput(t, []byte("ABCDEFGHIJ"))
		result, err := stream.ReadString(10)
		if err != nil {
			t.Fatal(err)
		}
		if result != "ABCDEFGHIJ" {
			t.Fatalf("result=%q", result)
		}
		if err := stream.Close(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("testReadStringAcrossBufferRefill", func(t *testing.T) {
		data := make([]byte, 8292)
		for i := range data {
			data[i] = byte('A' + i%26)
		}
		stream := javaBufferedDataInput(t, data)
		if err := stream.ReadFully(make([]byte, 8190), 0, 8190); err != nil {
			t.Fatal(err)
		}
		result, err := stream.ReadString(10)
		if err != nil {
			t.Fatal(err)
		}
		if result != "ABCDEFGHIJ" {
			t.Fatalf("result=%q", result)
		}
		if err := stream.Close(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("testConstructorRejectsStreamReturningZero", func(t *testing.T) {
		defer func() {
			failure := recover()
			exception, ok := failure.(*TLCError)
			if !ok || !exception.Runtime {
				t.Fatalf("Constructor should throw TLCRuntimeException, got %T: %v", failure, failure)
			}
			if exception.Code != ECSystemStreamEmpty {
				t.Fatalf("errorCode=%d, want SYSTEM_STREAM_EMPTY", exception.Code)
			}
		}()
		_, err := NewBufferedDataInputStream(javaBufferedDataZeroStream{})
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("Constructor should throw when underlying stream returns 0 from read")
	})
	t.Run("testEmptyStream", func(t *testing.T) {
		stream := javaBufferedDataInput(t, []byte{})
		if !stream.AtEOF() {
			t.Fatal("atEOF() should be true")
		}
		_, err := stream.ReadByte()
		if _, ok := err.(*EOFException); !ok {
			t.Fatalf("readByte() should throw EOFException, got %T: %v", err, err)
		}
		_, err = stream.ReadString(1)
		if _, ok := err.(*EOFException); !ok {
			t.Fatalf("readString(1) should throw EOFException, got %T: %v", err, err)
		}
		if err := stream.Close(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("testWriteReadStringRoundTrip", func(t *testing.T) {
		original := "Hello, TLA+!"
		var buffer bytes.Buffer
		output := NewBufferedDataOutputStream(&buffer)
		if err := output.WriteInt(int32(len(javaStringUTF16(original)))); err != nil {
			t.Fatal(err)
		}
		if err := output.WriteString(original); err != nil {
			t.Fatal(err)
		}
		if err := output.WriteInt(42); err != nil {
			t.Fatal(err)
		}
		if err := output.Close(); err != nil {
			t.Fatal(err)
		}
		input := javaBufferedDataInput(t, buffer.Bytes())
		length, err := input.ReadInt()
		if err != nil {
			t.Fatal(err)
		}
		result, err := input.ReadString(int(length))
		if err != nil {
			t.Fatal(err)
		}
		if result != original {
			t.Fatalf("result=%q, want %q", result, original)
		}
		sentinel, err := input.ReadInt()
		if err != nil {
			t.Fatal(err)
		}
		if sentinel != 42 {
			t.Fatalf("sentinel=%d, want 42", sentinel)
		}
		if err := input.Close(); err != nil {
			t.Fatal(err)
		}
	})
}

type javaBufferedDataZeroStream struct{}

func (javaBufferedDataZeroStream) Read(p []byte) (int, error)  { return 0, nil }
func (javaBufferedDataZeroStream) ReadByteValue() (int, error) { return 0, nil }
