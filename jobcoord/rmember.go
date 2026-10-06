package jobcoord

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/glycerine/rpc25519/tube"
)

// RMemberAuthority is the coordinator identity observed through Tube's trusted
// membership path. It is not an Active job record: activation still requires a
// committed RunGeneration, recovery, and resource fencing.
type RMemberAuthority struct {
	Epoch   int64
	Version int64
	Name    string
	PeerID  string
	URL     string
	Until   time.Time
}

// RMemberMonitor consumes Tube/RMember's typed membership messages. Call Observe
// only for authenticated Czar Ping replies or the local RMember upcall channel,
// never for arbitrary batch messages or the stale value under Tube's czar key.
// It copies the fields it owns and applies consensus epoch/version ordering.
type RMemberMonitor struct {
	mu         sync.Mutex
	current    RMemberAuthority
	drift      time.Duration
	now        func() time.Time
	stopped    bool
	running    bool
	refreshErr error
	localName  string
	localUntil time.Time
}

func NewRMemberMonitor(drift time.Duration, now func() time.Time) *RMemberMonitor {
	if drift < 0 {
		panic("negative clock drift bound")
	}
	if now == nil {
		now = time.Now
	}
	return &RMemberMonitor{drift: drift, now: now}
}

func (m *RMemberMonitor) Observe(reply *tube.PingReply) error {
	if reply == nil || reply.Vers == nil || reply.Members == nil || reply.Members.CzarDet == nil || reply.Members.CzarDet.Det == nil {
		return ErrInvalidRequest
	}
	v, detail := reply.Vers, reply.Members.CzarDet.Det
	next := RMemberAuthority{Epoch: v.CzarLeaseEpoch, Version: v.WithinCzarVersion, Name: reply.Members.CzarName, PeerID: detail.PeerID, URL: detail.URL, Until: v.CzarLeaseUntilTm}
	if next.Epoch <= 0 || next.Version < 0 || next.Name == "" || detail.Name != next.Name || next.PeerID == "" || next.URL == "" {
		return ErrInvalidRequest
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return ErrAuthorityFenced
	}
	old := m.current
	if next.Epoch < old.Epoch {
		return ErrStaleAuthority
	}
	if next.Epoch == old.Epoch {
		if next.Version < old.Version || next.PeerID != old.PeerID || next.Name != old.Name || next.URL != old.URL || next.Until.Before(old.Until) {
			return ErrStaleAuthority
		}
	}
	if !m.now().Before(next.Until.Add(-m.drift)) {
		return ErrAuthorityExpired
	}
	m.current = next
	if m.localName != "" {
		m.localUntil = time.Time{}
		if m.localName == next.Name {
			m.localUntil = next.Until
		} else if reply.Members.PeerNames != nil {
			if detail, ok := reply.Members.PeerNames.Get2(m.localName); ok && detail != nil && detail.Det != nil && detail.Det.Name == m.localName {
				m.localUntil = detail.RMemberLeaseUntilTm
			}
		}
	}
	m.refreshErr = nil
	return nil
}

func (m *RMemberMonitor) Current() (RMemberAuthority, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return RMemberAuthority{}, ErrAuthorityFenced
	}
	if m.current.Epoch == 0 {
		return RMemberAuthority{}, ErrStaleAuthority
	}
	if !m.now().Before(m.current.Until.Add(-m.drift)) {
		return RMemberAuthority{}, ErrAuthorityExpired
	}
	view := m.current
	if m.localName != "" {
		if m.localUntil.IsZero() {
			return RMemberAuthority{}, ErrAuthorityFenced
		}
		if !m.now().Before(m.localUntil.Add(-m.drift)) {
			return RMemberAuthority{}, ErrAuthorityExpired
		}
		if m.localUntil.Before(view.Until) {
			view.Until = m.localUntil
		}
	}
	return view, nil
}

// RefreshError reports the most recent failed authoritative refresh. A failure
// does not extend or instantly invalidate an existing, still-valid lease.
func (m *RMemberMonitor) RefreshError() error { m.mu.Lock(); defer m.mu.Unlock(); return m.refreshErr }

// Run drains an already-started RMember's finite event channels. Ready means the
// channels are usable, not that coordinator authority or job activation exists.
// A time-only lease event requests an authenticated, epoch-tagged refresh. Its
// timestamp cannot change coordinator identity or extend the cached authority.
// Refreshes are serial and coalesced, run independently of the event loop, and
// must honor context cancellation. The caller owns starting/stopping RMember.
// This monitor is single-use; returning fences its observed authority.
func (m *RMemberMonitor) Run(ctx context.Context, member *tube.RMember, refresh func(context.Context) (*tube.PingReply, error)) error {
	if member == nil || member.Ready == nil || refresh == nil {
		return ErrInvalidRequest
	}
	m.mu.Lock()
	if m.running || m.stopped {
		m.mu.Unlock()
		return ErrInvalidRequest
	}
	m.running = true
	m.mu.Unlock()
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	defer func() {
		cancel()
		m.mu.Lock()
		m.stopped = true
		m.mu.Unlock()
		wg.Wait()
	}()
	select {
	case <-member.Ready.Chan:
	case <-ctx.Done():
		return ctx.Err()
	}
	if member.UpcallMembershipChangeCh == nil || member.OperatingLeaseRenewCh == nil {
		return ErrInvalidRequest
	}
	m.mu.Lock()
	m.localName = member.Name
	m.localUntil = time.Time{}
	m.mu.Unlock()
	wake := make(chan struct{}, 1)
	requestRefresh := func() {
		select {
		case wake <- struct{}{}:
		default:
		}
	}
	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-wake:
			}
			reply, err := refresh(ctx)
			if err == nil {
				err = m.Observe(reply)
			}
			if err != nil {
				m.mu.Lock()
				m.refreshErr = err
				m.mu.Unlock()
			}
		}
	})
	requestRefresh()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case reply, ok := <-member.UpcallMembershipChangeCh:
			if !ok {
				return ErrAuthorityFenced
			}
			if err := m.Observe(reply); err != nil && !errors.Is(err, ErrStaleAuthority) && !errors.Is(err, ErrAuthorityExpired) {
				return err
			}
		case _, ok := <-member.OperatingLeaseRenewCh:
			if !ok {
				return ErrAuthorityFenced
			}
			requestRefresh()
		}
	}
}
