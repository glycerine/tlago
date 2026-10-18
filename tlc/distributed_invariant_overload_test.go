package tlc

import "testing"

type invariantOverloadChild struct {
	*NoopFPSet
	checks int
}

func (s *invariantOverloadChild) CheckInvariant(...uint64) bool {
	s.checks++
	return false
}

func (s *invariantOverloadChild) Size() uint64 {
	panic("inherited expected-count invariant must not read child size")
}

// The source base class overload returns true without calling the no-argument
// overload. There is no direct original test of memory/nested overload dispatch.
func TestDistributedInvariantOverloadDispatch(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "tcp"}[remote], func(t *testing.T) {
			for _, fixture := range []struct {
				name string
				set  FPSet
			}{
				{"memory", NewMemFPSet()},
				{"memory1", NewMemFPSet1(NewFPSetConfiguration())},
				{"memory2", NewMemFPSet2(NewFPSetConfiguration())},
			} {
				t.Run(fixture.name, func(t *testing.T) {
					fixture.set.Put(17)
					var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(fixture.set)
					if remote {
						_, endpoint = startFingerprintRPC(t, endpoint)
					}
					if valid, err := endpoint.CheckInvariant(99); err != nil || !valid {
						t.Fatalf("base expected-count overload = %v/%v, want true", valid, err)
					}
				})
			}
			for _, fixture := range []struct {
				name string
				make func(*FPSetConfiguration) *DiskFPSet
			}{
				{"lsb", func(c *FPSetConfiguration) *DiskFPSet { return NewLSBDiskFPSet(c).DiskFPSet }},
				{"msb", func(c *FPSetConfiguration) *DiskFPSet { return NewMSBDiskFPSet(c).DiskFPSet }},
			} {
				t.Run(fixture.name, func(t *testing.T) {
					config := NewFPSetConfiguration()
					config.SetMemory(1 << 20)
					set := fixture.make(config)
					set.Init(1, t.TempDir(), fixture.name)
					defer set.Close()
					set.Put(17)
					var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(set)
					if remote {
						_, endpoint = startFingerprintRPC(t, endpoint)
					}
					for _, expected := range []uint64{99, 1} {
						if valid, err := endpoint.CheckInvariant(expected); err != nil || valid != (expected == 1) {
							t.Fatalf("disk expected-count overload (%d) = %v/%v", expected, valid, err)
						}
					}
				})
			}
			t.Run("nested", func(t *testing.T) {
				child := &invariantOverloadChild{NoopFPSet: NewNoopFPSet(nil)}
				set := &MultiFPSet{Sets: []FPSet{child}}
				var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(set)
				if remote {
					_, endpoint = startFingerprintRPC(t, endpoint)
				}
				if valid, err := endpoint.CheckInvariant(99); err != nil || !valid {
					t.Fatalf("inherited nested overload = %v/%v, want true", valid, err)
				}
				if child.checks != 0 {
					t.Fatalf("expected-count overload checked children %d times", child.checks)
				}
				if valid, err := endpoint.CheckInvariant(); err != nil || valid || child.checks != 1 {
					t.Fatalf("no-argument nested overload = %v/%v, child checks %d", valid, err, child.checks)
				}
			})
		})
	}
}
