package tlc

import "testing"

// No original method directly exercises missing disk-read owners.
func TestDistributedTraceReadsRequireExistingOwner(t *testing.T) {
	for _, operation := range []string{"previous", "fingerprint", "depth", "public-depth", "reporting", "elements"} {
		for _, scenario := range []string{"missing", "missing-prior-error", "closed", "healthy"} {
			t.Run(operation+"/"+scenario, func(t *testing.T) {
				trace := NewTLCTrace(t.TempDir(), "Spec")
				defer trace.Close()
				if err := trace.raf.WriteLongNat(1); err != nil {
					t.Fatal(err)
				}
				if err := trace.raf.WriteLong(73); err != nil {
					t.Fatal(err)
				}
				if err := trace.raf.Flush(); err != nil {
					t.Fatal(err)
				}
				owner := trace.raf
				if scenario == "missing" || scenario == "missing-prior-error" {
					if err := owner.Close(); err != nil {
						t.Fatal(err)
					}
					trace.raf, owner = nil, nil
				} else if scenario == "closed" {
					if err := trace.Close(); err != nil {
						t.Fatal(err)
					}
				}
				if scenario == "missing-prior-error" || scenario == "healthy" {
					trace.traceErr = NewIOException("earlier creation failure")
				}
				trace.lastPtr, trace.level, trace.previousLevel = 0, 7, 3
				var result int64
				err := invokeDistributedServerOperation(func() error {
					switch operation {
					case "previous":
						var err error
						result, err = trace.getPrevFromDiskLocked(0)
						return err
					case "fingerprint":
						fp, err := trace.getFPFromDiskLocked(0)
						result = int64(fp)
						return err
					case "depth":
						depth, err := trace.getLevelFromDiskLocked(0)
						result = int64(depth)
						return err
					case "public-depth":
						result = int64(trace.GetLevel(0))
					case "reporting":
						depth, err := trace.GetLevelForReportingWithError()
						result = int64(depth)
						return err
					case "elements":
						enumerator, err := trace.Elements()
						if err != nil {
							return err
						}
						defer enumerator.Close()
						fp, err := enumerator.NextFP()
						result = int64(fp)
						return err
					}
					return nil
				})
				if trace.raf != owner || trace.lastPtr != 0 || trace.previousLevel != 3 {
					t.Fatal("read changed owner or reporting metadata")
				}
				if scenario == "healthy" {
					want := int64(1)
					if operation == "fingerprint" || operation == "elements" {
						want = 73
					} else if operation == "reporting" {
						want = 3
					}
					if err != nil || result != want {
						t.Fatalf("read = %d/%v, want %d", result, err, want)
					}
				} else if scenario == "closed" {
					if !isJavaIOException(err) {
						t.Fatalf("closed owner failure: %T/%v", err, err)
					}
				} else if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing owner failure: %T/%v", err, err)
				}
				if !trace.mu.TryLock() {
					t.Fatal("read retained trace lock")
				}
				trace.mu.Unlock()
			})
		}
	}
}

func TestDistributedTracePublicDepthPropagatesReadFailure(t *testing.T) {
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	if err := trace.raf.WriteFull([]byte{0, 0}); err != nil {
		t.Fatal(err)
	}
	if err := trace.raf.Seek(17); err != nil {
		t.Fatal(err)
	}
	err := invokeDistributedServerOperation(func() error { trace.GetLevel(0); return nil })
	if !isJavaIOException(err) || trace.raf.curr != 2 {
		t.Fatalf("depth failure swallowed or cursor restored: %v, cursor %d", err, trace.raf.curr)
	}
}
