package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPoolSynchronousNullArraysThrowBeforeMutation(t *testing.T) {
	oldGZIP := UseGZIP()
	SetUseGZIP(false)
	t.Cleanup(func() { SetUseGZIP(oldGZIP) })
	directory := t.TempDir()
	stateFile := filepath.Join(directory, "state")
	byteFile := filepath.Join(directory, "byte")
	if err := os.WriteFile(stateFile, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(byteFile, []byte{0, 0, 0, 0}, 0600); err != nil {
		t.Fatal(err)
	}

	type operation struct {
		call      func()
		unchanged func() bool
	}
	stateReader := NewStatePoolReader(1, stateFile)
	byteReader := NewByteArrayPoolReader(1, byteFile)
	stateWriter := NewStatePoolWriter(1, nil)
	byteWriter := NewByteArrayPoolWriter(1, nil)
	if _, err := stateWriter.DoWork(nil, stateFile); err != nil {
		t.Fatal(err)
	}
	if _, err := byteWriter.DoWork(nil, byteFile); err != nil {
		t.Fatal(err)
	}
	operations := map[string]operation{
		"state_reader": {
			call: func() { _, _ = stateReader.DoWork(nil, stateFile+"next") },
			unchanged: func() bool {
				return stateReader.poolFile == stateFile && !stateReader.isFull && !stateReader.canRead
			},
		},
		"byte_reader": {
			call: func() { _, _ = byteReader.DoWork(nil, byteFile+"next") },
			unchanged: func() bool {
				return byteReader.poolFile == byteFile && !byteReader.isFull && !byteReader.canRead
			},
		},
		"state_writer": {
			call: func() { _, _ = stateWriter.DoWork(make([]*TLCStateMut, 1), stateFile+"next") },
			unchanged: func() bool {
				return stateWriter.poolFile == stateFile && stateWriter.buf == nil
			},
		},
		"byte_writer": {
			call: func() { _, _ = byteWriter.DoWork(make([][]byte, 1), byteFile+"next") },
			unchanged: func() bool {
				return byteWriter.poolFile == byteFile && byteWriter.buf == nil
			},
		},
	}

	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
				if !operation.unchanged() {
					t.Fatal("null-array failure advanced pool work")
				}
			}()
			operation.call()
		})
	}
}
