package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

// No enabled original test covers interrupted trace records during distributed
// fingerprint reconstruction. The source enumerator propagates IOException;
// it must never replace a failed fingerprint read with fingerprint zero.
func TestDistributedTraceRecoveryPropagatesTruncatedRecord(t *testing.T) {
	for _, truncatedField := range []string{"predecessor", "fingerprint"} {
		t.Run(truncatedField, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "Spec.st")
			file, err := NewBufferedRandomAccessFile(path, "rw")
			if err != nil {
				t.Fatal(err)
			}
			if err := file.WriteLongNat(1); err != nil {
				t.Fatal(err)
			}
			if err := file.WriteLong(41); err != nil {
				t.Fatal(err)
			}
			if truncatedField == "fingerprint" {
				if err := file.WriteLongNat(1); err != nil {
					t.Fatal(err)
				}
			}
			if err := file.WriteByteValue(1); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			set := NewMultiFPSet(NewFPSetConfigurationWithRatioAndImplementation(1, "tlc2.tool.fp.MemFPSet"))
			err = set.RecoverTrace(trace)
			if !isJavaIOException(err) {
				t.Fatalf("truncated %s did not propagate IO failure: %v", truncatedField, err)
			}
			if set.Size() != 1 || !set.Contains(41) || set.Contains(0) {
				t.Fatal("failed trace record was inserted or valid preceding record lost")
			}
		})
	}
}

func TestDistributedTraceRecoveryPropagatesMissingTrace(t *testing.T) {
	directory := t.TempDir()
	trace := NewTLCTrace(directory, "Spec")
	defer trace.Close()
	if err := os.Remove(filepath.Join(directory, "Spec.st")); err != nil {
		t.Fatal(err)
	}
	set := NewMultiFPSet(NewFPSetConfigurationWithRatioAndImplementation(1, "tlc2.tool.fp.MemFPSet"))
	if err := set.RecoverTrace(trace); !isJavaIOException(err) {
		t.Fatalf("missing trace did not propagate IO failure: %v", err)
	}
	if set.Size() != 0 {
		t.Fatal("missing trace inserted fingerprints")
	}
}

func TestTraceEnumeratorResetAndZeroFingerprint(t *testing.T) {
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	write := func(fp int64) {
		if err := trace.raf.WriteLongNat(1); err != nil {
			t.Fatal(err)
		}
		if err := trace.raf.WriteLong(fp); err != nil {
			t.Fatal(err)
		}
		if err := trace.raf.Flush(); err != nil {
			t.Fatal(err)
		}
	}
	write(0)
	elements, err := trace.Elements()
	if err != nil {
		t.Fatal(err)
	}
	defer elements.Close()
	if fp, err := elements.NextFP(); err != nil || fp != 0 {
		t.Fatalf("valid zero fingerprint = %d/%v", fp, err)
	}
	if pos, err := elements.NextPos(); err != nil || pos != -1 {
		t.Fatalf("trace end = %d/%v", pos, err)
	}
	write(73)
	if pos, err := elements.NextPos(); err != nil || pos != -1 {
		t.Fatalf("snapshot length changed before reset: %d/%v", pos, err)
	}
	old := elements.raf
	if err := elements.Reset(-1); err != nil {
		t.Fatal(err)
	}
	if _, err := old.GetFilePointer(); !isJavaIOException(err) {
		t.Fatalf("replaced native reader was not released: %v", err)
	}
	if pos, err := elements.NextPos(); err != nil || pos == -1 {
		t.Fatalf("reset did not retain cursor/new owner length: %d/%v", pos, err)
	}
	if fp, err := elements.NextFP(); err != nil || fp != 73 {
		t.Fatalf("new trace fingerprint = %d/%v", fp, err)
	}
}

func TestTraceEnumeratorPropagatesCursorAndResetFailures(t *testing.T) {
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	elements, err := trace.Elements()
	if err != nil {
		t.Fatal(err)
	}
	if err := elements.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := elements.NextPos(); !isJavaIOException(err) {
		t.Fatalf("closed cursor failure = %v", err)
	}
	if _, err := elements.NextFP(); !isJavaIOException(err) {
		t.Fatalf("closed read failure = %v", err)
	}
	if err := elements.Reset(-1); !isJavaIOException(err) {
		t.Fatalf("closed reset cursor failure = %v", err)
	}
	elements, err = trace.Elements()
	if err != nil {
		t.Fatal(err)
	}
	defer elements.Close()
	old := elements.raf
	if err := os.Remove(trace.traceFileName()); err != nil {
		t.Fatal(err)
	}
	if err := elements.Reset(0); !isJavaIOException(err) {
		t.Fatalf("reset open failure = %v", err)
	}
	if elements.raf != old {
		t.Fatal("failed reset replaced the previous reader")
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	if err := elements.Reset(0); !isJavaIOException(err) {
		t.Fatalf("closed owner reset length failure = %v", err)
	}
	if _, err := trace.Elements(); !isJavaIOException(err) {
		t.Fatalf("closed owner enumerator creation failure = %v", err)
	}
}

func TestDiskTraceRecoveryPropagatesTruncatedRecord(t *testing.T) {
	for _, implementation := range []string{"tlc2.tool.fp.MSBDiskFPSet", "tlc2.tool.fp.OffHeapDiskFPSet"} {
		t.Run(implementation, func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "Spec.st"), []byte{1}, 0600); err != nil {
				t.Fatal(err)
			}
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
			config.SetMemory(1 << 20)
			config.SetFPBits(0)
			config.NoNesting = true
			set := NewFPSet(config).Init(1, directory, "fp")
			if _, nested := set.(*MultiFPSet); nested {
				t.Fatal("disk recovery fixture must exercise the direct implementation")
			}
			defer set.Close()
			if err := set.RecoverTrace(trace); !isJavaIOException(err) {
				t.Fatalf("truncated disk recovery lost IO failure: %v", err)
			}
			if set.Size() != 0 {
				t.Fatal("truncated disk trace inserted a fingerprint")
			}
		})
	}
}
