package tlc

import (
	"strings"
	"sync/atomic"
	"testing"
)

type observedNilTraceFingerprintEndpoint struct {
	*LocalFingerprintEndpoint
	recoveries atomic.Int32
}

func (e *observedNilTraceFingerprintEndpoint) RecoverTrace(trace *TLCTrace) error {
	if trace != nil {
		panic("native remote recovery received a coordinator trace")
	}
	e.recoveries.Add(1)
	return e.LocalFingerprintEndpoint.RecoverTrace(trace)
}

// No original method directly covers a null trace over the remote boundary.
// Mem ignores that argument; direct disk and MultiFPSet dereference it instead.
func TestFingerprintRPCRecoveryWithoutTrace(t *testing.T) {
	for _, row := range []struct {
		name       string
		needsTrace bool
	}{
		{"mem", false}, {"lsb", true}, {"msb", true}, {"nested-mem", true},
	} {
		t.Run(row.name, func(t *testing.T) {
			directory := t.TempDir()
			newStore := func() FPSet {
				config := NewFPSetConfigurationWithRatio(1)
				config.SetMemory(1 << 20)
				config.SetFPBits(0)
				var set FPSet
				switch row.name {
				case "lsb":
					set = NewLSBDiskFPSet(config)
				case "msb":
					set = NewMSBDiskFPSet(config)
				case "nested-mem":
					config.Implementation = "tlc2.tool.fp.MemFPSet"
					config.SetFPBits(1)
					set = NewFPSet(config)
				default:
					set = NewMemFPSetWithConfig(config)
				}
				return set.Init(1, directory, "Spec")
			}
			set := newStore()
			fingerprints := []uint64{41, uint64(1)<<63 | 43}
			for _, fp := range fingerprints {
				set.Put(fp)
			}
			if err := set.BeginChkptFile("Spec"); err != nil {
				t.Fatal(err)
			}
			if err := set.CommitChkptFile("Spec"); err != nil {
				t.Fatal(err)
			}
			set.Close()
			set = newStore()
			t.Cleanup(set.Close)
			observed := &observedNilTraceFingerprintEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(set)}
			_, client := startFingerprintRPC(t, observed)
			manager := NewNonDistributedFPSetManager(set, "coordinator", nil)
			manager.entry(0).set = client
			err := manager.Recover("not-the-stores-checkpoint-name")
			if row.needsTrace {
				if !isDistributedNullFailure(err) || isJavaIOException(err) {
					t.Fatalf("remote missing-trace failure = %T/%v, want unchecked null failure", err, err)
				}
				if set.Size() != 0 || set.Contains(41) {
					t.Fatal("missing trace fell back to committed file recovery")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if set.Size() != uint64(len(fingerprints)) {
					t.Fatal("remote nil-trace memory recovery lost checkpoint entries")
				}
				for _, fp := range fingerprints {
					if !set.Contains(fp) {
						t.Fatalf("recovered checkpoint omitted %d", fp)
					}
				}
			}
			// A real trace remains process-owned and must never reach this
			// endpoint. Rejection must also leave the RPC connection usable.
			if err := client.RecoverTrace(&TLCTrace{}); err == nil || !strings.Contains(err.Error(), "trace recovery runs at the coordinator") {
				t.Fatalf("non-null remote trace was accepted: %v", err)
			}
			if observed.recoveries.Load() != 1 {
				t.Fatalf("accepted recoveries = %d, want one null argument", observed.recoveries.Load())
			}
			if count, err := client.Size(); err != nil || count != set.Size() {
				t.Fatalf("trace recovery stopped the remote connection: %d/%v", count, err)
			}
		})
	}
}
