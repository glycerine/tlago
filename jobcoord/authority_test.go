package jobcoord

import (
	"errors"
	"testing"
	"time"
)

func TestAuthorityLeaseExpiry(t *testing.T) {
	t.Run("given valid authority then admission stops before lease expiry by the drift bound", func(t *testing.T) {
		now := time.Unix(100, 0)
		gate := NewAuthorityGate(500*time.Millisecond, func() time.Time { return now })
		lease := AuthorityLease{Execution: Execution{JobID: [16]byte{1}, CoordinatorEpoch: 2, RunGeneration: 1}, Owner: [16]byte{3}, Until: now.Add(time.Second)}
		if err := gate.Activate(lease); err != nil {
			t.Fatal(err)
		}
		permit, err := gate.Admit(lease.Execution, lease.Owner)
		if err != nil {
			t.Fatal(err)
		}
		now = now.Add(500 * time.Millisecond)
		if err := permit.Check(); !errors.Is(err, ErrAuthorityExpired) {
			t.Fatalf("paused operation = %v", err)
		}
		if _, err := gate.Admit(lease.Execution, lease.Owner); !errors.Is(err, ErrAuthorityExpired) {
			t.Fatalf("expired admission = %v", err)
		}
		permit.Release()
	})
}

func TestAuthorityTransitionDrain(t *testing.T) {
	t.Run("given old admitted work then a replacement cannot activate until fenced and drained", func(t *testing.T) {
		now := time.Unix(100, 0)
		gate := NewAuthorityGate(0, func() time.Time { return now })
		old := AuthorityLease{Execution: Execution{JobID: [16]byte{1}, CoordinatorEpoch: 2, RunGeneration: 1}, Owner: [16]byte{3}, Until: now.Add(time.Hour)}
		if err := gate.Activate(old); err != nil {
			t.Fatal(err)
		}
		permit, err := gate.Admit(old.Execution, old.Owner)
		if err != nil {
			t.Fatal(err)
		}
		next := old
		next.CoordinatorEpoch = 3
		next.Owner = [16]byte{4}
		if err := gate.Activate(next); !errors.Is(err, ErrAuthorityBusy) {
			t.Fatalf("unfenced transition = %v", err)
		}
		gate.Fence()
		if _, err := gate.Admit(old.Execution, old.Owner); !errors.Is(err, ErrAuthorityFenced) {
			t.Fatalf("old admission = %v", err)
		}
		if err := gate.Activate(next); !errors.Is(err, ErrAuthorityBusy) {
			t.Fatalf("undrained transition = %v", err)
		}
		permit.Release()
		permit.Release()
		if err := gate.Activate(next); err != nil {
			t.Fatal(err)
		}
		if _, err := gate.Admit(old.Execution, old.Owner); !errors.Is(err, ErrStaleAuthority) {
			t.Fatalf("old sender = %v", err)
		}
		fresh, err := gate.Admit(next.Execution, next.Owner)
		if err != nil {
			t.Fatal(err)
		}
		fresh.Release()
		if err := gate.Activate(old); !errors.Is(err, ErrStaleAuthority) {
			t.Fatalf("rollback authority = %v", err)
		}
	})
}

func TestAuthorityRenewalPreservesExecution(t *testing.T) {
	t.Run("given lease renewal then outstanding work remains under the same authority", func(t *testing.T) {
		now := time.Unix(100, 0)
		gate := NewAuthorityGate(0, func() time.Time { return now })
		lease := AuthorityLease{Execution: Execution{JobID: [16]byte{1}, CoordinatorEpoch: 2, RunGeneration: 1}, Owner: [16]byte{3}, Until: now.Add(time.Second)}
		if err := gate.Activate(lease); err != nil {
			t.Fatal(err)
		}
		p, err := gate.Admit(lease.Execution, lease.Owner)
		if err != nil {
			t.Fatal(err)
		}
		renewed := lease
		renewed.Until = now.Add(time.Hour)
		if err := gate.Activate(renewed); err != nil {
			t.Fatal(err)
		}
		now = now.Add(time.Second)
		if err := p.Check(); err != nil {
			t.Fatal(err)
		}
		p.Release()
	})
}
