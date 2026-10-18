//go:build tlc_verify

package tlc

import (
	"bytes"
	"fmt"
	"testing"
)

// Original BufferedDataInputStreamJPFTest.testReadStringMustThrowOnShortStream.
// Its only choices are two bounded integers; enumerate their entire Cartesian
// product directly. There are no threads or scheduling choices in this method.
func TestJavaBufferedDataInputStreamVerify(t *testing.T) {
	for streamLen := 1; streamLen <= 15; streamLen++ {
		for requestLen := 1; requestLen <= 20; requestLen++ {
			t.Run(fmt.Sprintf("stream%d/request%d", streamLen, requestLen), func(t *testing.T) {
				data := make([]byte, streamLen)
				for i := range data {
					data[i] = byte('A' + i%26)
				}
				stream, err := NewBufferedDataInputStream(bytes.NewReader(data))
				if err != nil {
					t.Fatalf("Unexpected IOException: %v", err)
				}
				defer func() {
					if err := stream.Close(); err != nil {
						t.Errorf("Unexpected IOException: %v", err)
					}
				}()
				result, err := stream.ReadString(requestLen)
				if _, eof := err.(*EOFException); eof {
					if requestLen <= streamLen {
						t.Fatalf("readString(%d) threw EOFException but stream had %d bytes (enough)", requestLen, streamLen)
					}
					return
				}
				if err != nil {
					t.Fatalf("Unexpected IOException: %v", err)
				}
				if requestLen > streamLen {
					t.Fatalf("readString(%d) returned %q instead of throwing EOFException (stream had only %d bytes)", requestLen, result, streamLen)
				}
				for i := 0; i < requestLen; i++ {
					want := byte('A' + i%26)
					if i >= len(result) || result[i] != want {
						t.Fatalf("readString(%d) returned wrong char at index %d: expected %q, result %q", requestLen, i, want, result)
					}
				}
			})
		}
	}
}
