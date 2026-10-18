package tlc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTLCTraceIOFailuresMatchJava(t *testing.T) {
	directory := t.TempDir()
	validPath := filepath.Join(directory, "trace.bin")
	missingPath := filepath.Join(directory, "missing", "trace.bin")

	t.Run("serialize null path precedes value", func(t *testing.T) {
		_, err := TLCTraceSerialize(nil, nil)
		failure, ok := err.(*NullPointerException)
		want := `Cannot read field "val" because "absolutePath" is null`
		if !ok || failure.Error() != want {
			t.Fatalf("failure = %T/%v, want NullPointerException %q", err, err, want)
		}
	})

	t.Run("serialize null value follows open", func(t *testing.T) {
		_, err := TLCTraceSerialize(nil, NewStringValue(validPath))
		failure, ok := err.(*NullPointerException)
		want := `Cannot invoke "tlc2.value.IValue.fingerPrint(long)" because "value" is null`
		if !ok || failure.Error() != want {
			t.Fatalf("failure = %T/%v, want NullPointerException %q", err, err, want)
		}
		raw, readErr := os.ReadFile(validPath)
		wantBytes := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		if readErr != nil || !bytes.Equal(raw, wantBytes) {
			t.Fatalf("opened trace = %x error %v, want Java empty gzip %x", raw, readErr, wantBytes)
		}
	})

	openFailures := []struct {
		name string
		path string
		want string
	}{
		{name: "invalid", path: "bad\x00path", want: "Invalid file path"},
		{name: "missing parent", path: missingPath, want: missingPath + " (No such file or directory)"},
		{name: "directory", path: directory, want: directory + " (Is a directory)"},
	}
	for _, tc := range openFailures {
		t.Run("serialize "+tc.name, func(t *testing.T) {
			_, err := TLCTraceSerialize(NewIntValue(1), NewStringValue(tc.path))
			failure, ok := err.(*FileNotFoundException)
			if !ok || failure.Error() != tc.want {
				t.Fatalf("failure = %T/%v, want FileNotFoundException %q", err, err, tc.want)
			}
		})
	}

	t.Run("deserialize null path", func(t *testing.T) {
		_, err := TLCTraceDeserialize(nil)
		failure, ok := err.(*NullPointerException)
		want := `Cannot read field "val" because "absolutePath" is null`
		if !ok || failure.Error() != want {
			t.Fatalf("failure = %T/%v, want NullPointerException %q", err, err, want)
		}
	})

	for name, path := range map[string]string{"invalid": "bad\x00path", "missing": missingPath} {
		t.Run("deserialize "+name, func(t *testing.T) {
			value, err := TLCTraceDeserialize(NewStringValue(path))
			if err != nil || value != EmptyRecord {
				t.Fatalf("result = %T/%v error %v, want EmptyRecord", value, value, err)
			}
		})
	}

	t.Run("deserialize directory", func(t *testing.T) {
		_, err := TLCTraceDeserialize(NewStringValue(directory))
		failure, ok := err.(*FileNotFoundException)
		want := directory + " (Is a directory)"
		if !ok || failure.Error() != want {
			t.Fatalf("failure = %T/%v, want FileNotFoundException %q", err, err, want)
		}
	})
}
