package tlc

import (
	"path/filepath"
	"testing"
)

func TestStatePoolWriterThrowsRuntimeSerializationFailureBeforeMutation(t *testing.T) {
	oldGZIP := UseGZIP()
	SetUseGZIP(false)
	t.Cleanup(func() { SetUseGZIP(oldGZIP) })

	name := filepath.Join(t.TempDir(), "pool")
	pending := []*TLCStateMut{{functional: true}}
	writer := NewStatePoolWriter(1, nil)
	if _, err := writer.DoWork(pending, name); err != nil {
		t.Fatal(err)
	}

	defer func() {
		failure := recover()
		if _, ok := failure.(*WrongInvocationException); !ok {
			t.Fatalf("failure = %T: %v, want WrongInvocationException", failure, failure)
		}
		if writer.poolFile != name || len(writer.buf) != 1 || writer.buf[0] != pending[0] {
			t.Fatal("runtime serialization failure advanced writer work")
		}
	}()
	_, _ = writer.DoWork(make([]*TLCStateMut, 1), name+"next")
}

func TestBytePoolReaderDirectBranchThrowsNullFileFailureBeforeMutation(t *testing.T) {
	reader := NewByteArrayPoolReader(1, "")
	destination := [][]byte{{1}}

	defer func() {
		failure := recover()
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
		}
		if reader.poolFile != "" || reader.isFull || reader.canRead || reader.buf[0] != nil || destination[0][0] != 1 {
			t.Fatal("direct null-file failure mutated reader state")
		}
	}()
	_, _ = reader.DoWork(destination, "ignored")
}
