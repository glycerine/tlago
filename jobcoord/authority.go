package jobcoord

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrStaleAuthority   = errors.New("wrong coordinator authority or execution")
	ErrAuthorityExpired = errors.New("coordinator operating lease expired")
	ErrAuthorityFenced  = errors.New("execution admission fenced")
	ErrAuthorityBusy    = errors.New("previous execution is not fenced and drained")
)

// Execution is a job's activation identity. CoordinatorEpoch is supplied by
// Tube RMember's CzarLeaseEpoch; RunGeneration is committed in Tube JobControl.
type Execution struct {
	JobID            [16]byte
	CoordinatorEpoch int64
	RunGeneration    uint64
}

type AuthorityLease struct {
	Execution
	Owner [16]byte
	Until time.Time
}

// AuthorityGate implements local side-effect admission and transition draining.
// Only the trusted Tube/RMember control adapter may Activate it. It is not a
// consensus engine and must never install authority from an ordinary message.
// On machine loss, the replacement uses isolated storage; draining this process
// cannot stand in for fencing an unreachable process's independent storage.
type AuthorityGate struct {
	mu          sync.Mutex
	lease       AuthorityLease
	initialized bool
	enabled     bool
	active      int
	drift       time.Duration
	now         func() time.Time
}

func NewAuthorityGate(drift time.Duration, now func() time.Time) *AuthorityGate {
	if drift < 0 {
		panic("negative clock drift bound")
	}
	if now == nil {
		now = time.Now
	}
	return &AuthorityGate{drift: drift, now: now}
}

// Activate installs validated control authority. Renewals extend the same lease
// without invalidating its active permits. Changing execution requires Fence and
// release of all prior permits. A fenced execution cannot be reopened by renewal.
func (g *AuthorityGate) Activate(next AuthorityLease) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if next.JobID == ([16]byte{}) || next.Owner == ([16]byte{}) || next.CoordinatorEpoch <= 0 || next.RunGeneration == 0 {
		return ErrInvalidRequest
	}
	if g.initialized {
		old := g.lease
		if old.JobID != next.JobID || next.CoordinatorEpoch < old.CoordinatorEpoch || (next.CoordinatorEpoch == old.CoordinatorEpoch && next.RunGeneration < old.RunGeneration) {
			return ErrStaleAuthority
		}
		if next.Execution == old.Execution {
			if next.Owner != old.Owner {
				return ErrStaleAuthority
			}
			if !g.enabled {
				return ErrAuthorityFenced
			}
			if next.Until.Before(old.Until) {
				return ErrStaleAuthority
			}
		} else if g.enabled || g.active != 0 {
			return ErrAuthorityBusy
		}
	}
	if !g.now().Before(next.Until.Add(-g.drift)) {
		return ErrAuthorityExpired
	}
	g.lease = next
	g.initialized = true
	g.enabled = true
	return nil
}

// Fence immediately stops admission. Previously admitted operations must release
// their permits before a successor execution can reuse this gate/storage.
func (g *AuthorityGate) Fence() { g.mu.Lock(); g.enabled = false; g.mu.Unlock() }

func (g *AuthorityGate) validate(exec Execution, owner [16]byte) error {
	if !g.initialized || exec != g.lease.Execution || owner != g.lease.Owner {
		return ErrStaleAuthority
	}
	if !g.enabled {
		return ErrAuthorityFenced
	}
	if !g.now().Before(g.lease.Until.Add(-g.drift)) {
		return ErrAuthorityExpired
	}
	return nil
}

func (g *AuthorityGate) Admit(exec Execution, owner [16]byte) (*AuthorityPermit, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.validate(exec, owner); err != nil {
		return nil, err
	}
	g.active++
	return &AuthorityPermit{gate: g, execution: exec, owner: owner}, nil
}

// AuthorityPermit holds the drain obligation for one admitted operation. Check
// after pauses and before beginning side effects. An already-started bounded
// mutation must finish into old storage even if authority is lost mid-operation;
// its permit blocks successor activation until Release.
type AuthorityPermit struct {
	gate      *AuthorityGate
	execution Execution
	owner     [16]byte
	released  atomic.Bool
}

func (p *AuthorityPermit) Check() error {
	p.gate.mu.Lock()
	defer p.gate.mu.Unlock()
	if p.released.Load() {
		return ErrAuthorityFenced
	}
	return p.gate.validate(p.execution, p.owner)
}

func (p *AuthorityPermit) Release() {
	if !p.released.CompareAndSwap(false, true) {
		return
	}
	p.gate.mu.Lock()
	p.gate.active--
	p.gate.mu.Unlock()
}
