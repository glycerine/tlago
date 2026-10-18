package tlc

import "testing"

// No original method directly covers missing disk enumerator owners.
func TestDiskTraceEnumeratorRequiresReader(t *testing.T) {
	for _, operation := range []string{"position", "fingerprint", "close"} {
		for _, scenario := range []string{"missing-reader", "missing-enumerator"} {
			t.Run(operation+"/"+scenario, func(t *testing.T) {
				trace := NewTLCTrace(t.TempDir(), "Spec")
				defer trace.Close()
				enumerator, err := trace.Elements()
				if err != nil {
					t.Fatal(err)
				}
				if err := enumerator.Close(); err != nil {
					t.Fatal(err)
				}
				enumerator.raf = nil
				if scenario == "missing-enumerator" {
					enumerator = nil
				}
				err = invokeDistributedServerOperation(func() error {
					switch operation {
					case "position":
						_, err := enumerator.NextPos()
						return err
					case "fingerprint":
						_, err := enumerator.NextFP()
						return err
					default:
						return enumerator.Close()
					}
				})
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing reader returned completion: %T/%v", err, err)
				}
			})
		}
	}
}

func TestDiskTraceEnumeratorResetPreservesOwnerOrdering(t *testing.T) {
	for _, scenario := range []string{"missing-enumerator", "missing-trace", "missing-writer", "missing-reader-current", "missing-reader-explicit"} {
		t.Run(scenario, func(t *testing.T) {
			trace := NewTLCTrace(t.TempDir(), "Spec")
			defer trace.Close()
			if err := trace.raf.WriteFull(make([]byte, 12)); err != nil {
				t.Fatal(err)
			}
			if err := trace.raf.Flush(); err != nil {
				t.Fatal(err)
			}
			enumerator, err := trace.Elements()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if enumerator != nil && enumerator.raf != nil {
					_ = enumerator.Close()
				}
			}()
			enumerator.length = 37
			position := int64(-1)
			switch scenario {
			case "missing-enumerator":
				_ = enumerator.Close()
				enumerator = nil
			case "missing-trace":
				enumerator.trace = nil
			case "missing-writer":
				if err := trace.raf.Close(); err != nil {
					t.Fatal(err)
				}
				trace.raf = nil
			default:
				if err := enumerator.Close(); err != nil {
					t.Fatal(err)
				}
				enumerator.raf = nil
				if scenario == "missing-reader-explicit" {
					position = 4
				}
			}
			var old *BufferedRandomAccessFile
			if enumerator != nil {
				old = enumerator.raf
			}
			err = invokeDistributedServerOperation(func() error { return enumerator.Reset(position) })
			if scenario == "missing-reader-explicit" {
				if err != nil || enumerator.length != 12 || enumerator.raf == nil || enumerator.raf.curr != 4 {
					t.Fatalf("explicit reset did not replace missing reader: %v", err)
				}
			} else {
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing reset owner failure: %T/%v", err, err)
				}
				if enumerator != nil {
					wantLength := int64(37)
					if scenario == "missing-reader-current" {
						wantLength = 12
					}
					if enumerator.length != wantLength || enumerator.raf != old {
						t.Fatal("reset failure lost source length/owner publication order")
					}
				}
			}
			if !trace.mu.TryLock() {
				t.Fatal("reset retained trace lock")
			}
			trace.mu.Unlock()
		})
	}
}
