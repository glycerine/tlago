package tlc

// DistributedFingerprintEndpoint supplies the fingerprint-server operations
// used by TLC. Storage configuration and allocation stay at the owning server;
// callers receive fingerprint answers and failures through this Go boundary.
type DistributedFingerprintEndpoint interface {
	AddThread() error
	Put(uint64) (bool, error)
	Contains(uint64) (bool, error)
	PutBlock(*LongVec) (*BitVector, error)
	ContainsBlock(*LongVec) (*BitVector, error)
	Size() (uint64, error)
	GetStatesSeen() (uint64, error)
	CheckFPs() (uint64, error)
	CheckInvariant(...uint64) (bool, error)
	Close() error
	Exit(bool) error
	BeginChkpt() error
	BeginChkptFile(string) error
	CommitChkpt() error
	CommitChkptFile(string) error
	RecoverFile(string) error
	RecoverTrace(*TLCTrace) error
}

type LocalFingerprintEndpoint struct {
	Set FPSet
}

func NewLocalFingerprintEndpoint(set FPSet) *LocalFingerprintEndpoint {
	if isNil(set) {
		panic(NewNullPointerException())
	}
	return &LocalFingerprintEndpoint{Set: set}
}

func (e *LocalFingerprintEndpoint) requireSet() FPSet {
	if e == nil || isNil(e.Set) {
		panic(NewNullPointerException())
	}
	return e.Set
}

func (e *LocalFingerprintEndpoint) AddThread() error { return e.requireSet().AddThread() }

func (e *LocalFingerprintEndpoint) Put(fp uint64) (bool, error) {
	return e.requireSet().Put(fp), nil
}

func (e *LocalFingerprintEndpoint) Contains(fp uint64) (bool, error) {
	return e.requireSet().Contains(fp), nil
}

func (e *LocalFingerprintEndpoint) PutBlock(fps *LongVec) (*BitVector, error) {
	return e.requireSet().PutBlock(fps), nil
}

func (e *LocalFingerprintEndpoint) ContainsBlock(fps *LongVec) (*BitVector, error) {
	return e.requireSet().ContainsBlock(fps), nil
}

func (e *LocalFingerprintEndpoint) Size() (uint64, error) { return e.requireSet().Size(), nil }

func (e *LocalFingerprintEndpoint) GetStatesSeen() (uint64, error) {
	return e.requireSet().GetStatesSeen(), nil
}

func (e *LocalFingerprintEndpoint) CheckFPs() (uint64, error) {
	return e.requireSet().CheckFPs(), nil
}

func (e *LocalFingerprintEndpoint) CheckInvariant(expected ...uint64) (bool, error) {
	requireInvariantExpectedCount(expected)
	return e.requireSet().CheckInvariant(expected...), nil
}

func (e *LocalFingerprintEndpoint) Close() error { e.requireSet().Close(); return nil }

func (e *LocalFingerprintEndpoint) Exit(cleanup bool) error { return e.requireSet().Exit(cleanup) }

func (e *LocalFingerprintEndpoint) BeginChkpt() error { return e.requireSet().BeginChkpt() }

func (e *LocalFingerprintEndpoint) BeginChkptFile(name string) error {
	return e.requireSet().BeginChkptFile(name)
}

func (e *LocalFingerprintEndpoint) CommitChkpt() error { return e.requireSet().CommitChkpt() }

func (e *LocalFingerprintEndpoint) CommitChkptFile(name string) error {
	return e.requireSet().CommitChkptFile(name)
}

func (e *LocalFingerprintEndpoint) RecoverFile(name string) error {
	return e.requireSet().RecoverFile(name)
}

// Trace recovery is a coordinator-local operation. The trace is never included
// in the worker's fingerprint-manager snapshot.
func (e *LocalFingerprintEndpoint) RecoverTrace(trace *TLCTrace) error {
	return e.requireSet().RecoverTrace(trace)
}

// Preserve the existing manager's catch behavior for local storage panics while
// also accepting errors returned by network endpoints. Fatal storage failures
// still escape; adding an endpoint must not turn them into ordinary failover.
func invokeFingerprintEndpoint[T any](call func() (T, error)) (value T, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
			if isJavaError(err) {
				panic(failure)
			}
		}
	}()
	value, err = call()
	if isJavaError(err) {
		panic(err)
	}
	return value, err
}

var _ DistributedFingerprintEndpoint = (*LocalFingerprintEndpoint)(nil)
