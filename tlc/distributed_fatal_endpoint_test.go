package tlc

import (
	"fmt"
	"testing"
)

type fatalFingerprintEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
	panics  bool
	calls   int
}

func (e *fatalFingerprintEndpoint) fail() error {
	e.calls++
	if e.panics {
		panic(e.failure)
	}
	return e.failure
}
func (e *fatalFingerprintEndpoint) Put(uint64) (bool, error)                   { return false, e.fail() }
func (e *fatalFingerprintEndpoint) Contains(uint64) (bool, error)              { return false, e.fail() }
func (e *fatalFingerprintEndpoint) PutBlock(*LongVec) (*BitVector, error)      { return nil, e.fail() }
func (e *fatalFingerprintEndpoint) ContainsBlock(*LongVec) (*BitVector, error) { return nil, e.fail() }
func (e *fatalFingerprintEndpoint) Size() (uint64, error)                      { return 0, e.fail() }
func (e *fatalFingerprintEndpoint) GetStatesSeen() (uint64, error)             { return 0, e.fail() }
func (e *fatalFingerprintEndpoint) BeginChkptFile(string) error                { return e.fail() }
func (e *fatalFingerprintEndpoint) RecoverFile(string) error                   { return e.fail() }
func (e *fatalFingerprintEndpoint) Exit(bool) error                            { return e.fail() }

// No enabled upstream test directly covers local Error versus catch(Exception)
// here. Returning a fatal error at the Go boundary must not turn it into failover.
func TestDistributedFingerprintLocalFatalFailuresEscape(t *testing.T) {
	for _, operation := range []struct {
		name string
		call func(*DistributedFPSetManager)
	}{
		{"put", func(m *DistributedFPSetManager) { m.Put(1) }},
		{"contains", func(m *DistributedFPSetManager) { m.Contains(1) }},
		{"put_block", func(m *DistributedFPSetManager) { m.PutBlock([]*LongVec{NewLongVecFrom([]int64{1})}) }},
		{"contains_block", func(m *DistributedFPSetManager) { m.ContainsBlock([]*LongVec{NewLongVecFrom([]int64{1})}) }},
		{"size", func(m *DistributedFPSetManager) { m.Size() }},
		{"states_seen", func(m *DistributedFPSetManager) { m.GetStatesSeen() }},
		{"checkpoint", func(m *DistributedFPSetManager) { _ = m.Checkpoint("Spec") }},
		{"recover", func(m *DistributedFPSetManager) { _ = m.Recover("Spec") }},
		{"close", func(m *DistributedFPSetManager) { _ = m.Close(false) }},
	} {
		for _, panics := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/panic=%v", operation.name, panics), func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				failure := NewAssertionError("fatal local storage")
				endpoint := &fatalFingerprintEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), failure: failure, panics: panics}
				manager := NewDistributedFPSetManager(endpoint)
				defer func() {
					if got := recover(); got != failure {
						t.Fatalf("fatal failure = %v, want original %v", got, failure)
					}
					if endpoint.calls != 1 || manager.managerIsBroken || !manager.entry(0).available || len(ToolIOGetAllMessages()) != 0 {
						t.Fatal("fatal failure was retried, marked unavailable or reported as failover")
					}
				}()
				operation.call(manager)
			})
		}
	}
}
