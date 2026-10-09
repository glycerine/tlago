package tlc

import "testing"

// MultiFPSet inherits FPSet.addThread's no-op, but overrides incWorkers to
// visit its children. No original test directly checks this distinction.
func TestDistributedNestedThreadRegistration(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "tcp"}[remote], func(t *testing.T) {
			first := NewLSBDiskFPSet(NewFPSetConfiguration())
			second := NewMSBDiskFPSet(NewFPSetConfiguration())
			first.Init(1, t.TempDir(), "first")
			t.Cleanup(first.Close)
			second.Init(1, t.TempDir(), "second")
			t.Cleanup(second.Close)
			set := &MultiFPSet{Sets: []FPSet{first, second}, Shift: 63}
			var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(set)
			if remote {
				_, endpoint = startFingerprintRPC(t, endpoint)
			}
			if err := endpoint.AddThread(); err != nil {
				t.Fatal(err)
			}
			if len(first.braf) != 1 || len(second.braf) != 1 {
				t.Fatalf("nested addThread opened child readers: %d/%d, want 1/1", len(first.braf), len(second.braf))
			}
			set.IncWorkers(2)
			if len(first.braf) != 3 || len(second.braf) != 3 {
				t.Fatalf("nested incWorkers lost child allocation: %d/%d, want 3/3", len(first.braf), len(second.braf))
			}
		})
	}
}
