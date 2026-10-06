package jobcoord

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/glycerine/idem"
	"github.com/glycerine/rpc25519/tube"
)

func memberReply(epoch, version int64, name, peer string, until time.Time) *tube.PingReply {
	return &tube.PingReply{Members: &tube.ReliableMembershipList{CzarName: name, CzarDet: &tube.PeerDetailPlus{Det: &tube.PeerDetail{Name: name, PeerID: peer, URL: "tcp://127.0.0.1:1234"}}}, Vers: &tube.RMVersionTuple{CzarLeaseEpoch: epoch, WithinCzarVersion: version, CzarLeaseUntilTm: until}}
}

func TestRMemberAuthorityUsesConsensusEpoch(t *testing.T) {
	t.Run("given an authoritative Tube upcall then use the consensus epoch instead of the stale serialized version", func(t *testing.T) {
		now := time.Unix(100, 0)
		monitor := NewRMemberMonitor(500*time.Millisecond, func() time.Time { return now })
		reply := memberReply(7, 3, "candidate-a", "peer-a", now.Add(time.Second))
		reply.Members.CzarVersProbablyStale = &tube.RMVersionTuple{CzarLeaseEpoch: 99}
		if err := monitor.Observe(reply); err != nil {
			t.Fatal(err)
		}
		reply.Vers.CzarLeaseEpoch = 98
		reply.Members.CzarDet.Det.PeerID = "mutated"
		view, err := monitor.Current()
		if err != nil || view.Epoch != 7 || view.Version != 3 || view.PeerID != "peer-a" {
			t.Fatalf("authority = %+v, %v", view, err)
		}
		now = now.Add(500 * time.Millisecond)
		if _, err := monitor.Current(); !errors.Is(err, ErrAuthorityExpired) {
			t.Fatalf("expired view = %v", err)
		}
	})
}

func TestRMemberAuthorityRejectsDelayedViews(t *testing.T) {
	t.Run("given a replacement coordinator then a delayed old view cannot restore old authority", func(t *testing.T) {
		now := time.Unix(100, 0)
		m := NewRMemberMonitor(0, func() time.Time { return now })
		if err := m.Observe(memberReply(7, 4, "a", "peer-a", now.Add(time.Hour))); err != nil {
			t.Fatal(err)
		}
		if err := m.Observe(memberReply(8, 0, "b", "peer-b", now.Add(time.Hour))); err != nil {
			t.Fatal(err)
		}
		if err := m.Observe(memberReply(7, 999, "a", "peer-a", now.Add(2*time.Hour))); !errors.Is(err, ErrStaleAuthority) {
			t.Fatalf("delayed view = %v", err)
		}
		if err := m.Observe(memberReply(8, 1, "a", "peer-a", now.Add(time.Hour))); !errors.Is(err, ErrStaleAuthority) {
			t.Fatalf("owner change within epoch = %v", err)
		}
		view, err := m.Current()
		if err != nil || view.PeerID != "peer-b" {
			t.Fatalf("replacement = %+v, %v", view, err)
		}
	})
}

func TestRMemberLoopRefreshesUntaggedRenewals(t *testing.T) {
	t.Run("given Tube's time-only renewal event then refresh tagged authority rather than trusting that event as a new epoch", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			member := &tube.RMember{Ready: idem.NewIdemCloseChan(), UpcallMembershipChangeCh: make(chan *tube.PingReply, 10), OperatingLeaseRenewCh: make(chan time.Time, 10)}
			member.Ready.Close()
			now := time.Now()
			m := NewRMemberMonitor(0, time.Now)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			entered, release := make(chan struct{}), make(chan struct{})
			var enteredOnce sync.Once
			refresh := func(ctx context.Context) (*tube.PingReply, error) {
				enteredOnce.Do(func() { close(entered) })
				select {
				case <-release:
					return memberReply(9, 0, "a", "peer-a", now.Add(time.Hour)), nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			done := make(chan error, 1)
			go func() { done <- m.Run(ctx, member, refresh) }()
			member.OperatingLeaseRenewCh <- now.Add(99 * time.Hour)
			<-entered
			if _, err := m.Current(); !errors.Is(err, ErrStaleAuthority) {
				t.Fatalf("time-only authority = %v", err)
			}
			// Refresh is blocked, but the finite membership channel must still drain.
			member.UpcallMembershipChangeCh <- memberReply(8, 0, "a", "peer-a", now.Add(time.Hour))
			synctest.Wait()
			view, err := m.Current()
			if err != nil || view.Epoch != 8 {
				t.Fatalf("upcall during refresh = %+v, %v", view, err)
			}
			close(release)
			synctest.Wait()
			view, err = m.Current()
			if err != nil || view.Epoch != 9 {
				t.Fatalf("refreshed authority = %+v, %v", view, err)
			}
			cancel()
			if err := <-done; !errors.Is(err, context.Canceled) {
				t.Fatalf("shutdown = %v", err)
			}
		})
	})
}

func TestRMemberFailedRefreshDoesNotExtendLease(t *testing.T) {
	t.Run("given quorum or transport loss then preserve only the existing unexpired authority", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			m := NewRMemberMonitor(0, time.Now)
			if err := m.Observe(memberReply(1, 0, "a", "peer-a", time.Now().Add(time.Second))); err != nil {
				t.Fatal(err)
			}
			member := &tube.RMember{Ready: idem.NewIdemCloseChan(), UpcallMembershipChangeCh: make(chan *tube.PingReply, 10), OperatingLeaseRenewCh: make(chan time.Time, 10)}
			member.Ready.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			failure := errors.New("quorum unavailable")
			done := make(chan error, 1)
			go func() {
				done <- m.Run(ctx, member, func(context.Context) (*tube.PingReply, error) { return nil, failure })
			}()
			synctest.Wait()
			if !errors.Is(m.RefreshError(), failure) {
				t.Fatalf("refresh status = %v", m.RefreshError())
			}
			if _, err := m.Current(); err != nil {
				t.Fatalf("still-valid lease = %v", err)
			}
			time.Sleep(time.Second)
			if _, err := m.Current(); !errors.Is(err, ErrAuthorityExpired) {
				t.Fatalf("expired authority after failed refresh = %v", err)
			}
			cancel()
			if err := <-done; !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			if _, err := m.Current(); !errors.Is(err, ErrAuthorityFenced) {
				t.Fatalf("stopped monitor = %v", err)
			}
		})
	})
}

func TestRMemberOwnLeaseConstrainsAuthority(t *testing.T) {
	t.Run("given a candidate's shorter operating lease then stop local admission even while the Czar lease remains valid", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			m := NewRMemberMonitor(0, time.Now)
			member := &tube.RMember{Name: "local", Ready: idem.NewIdemCloseChan(), UpcallMembershipChangeCh: make(chan *tube.PingReply, 10), OperatingLeaseRenewCh: make(chan time.Time, 10)}
			member.Ready.Close()
			reply := memberReply(1, 0, "czar", "peer-czar", time.Now().Add(time.Hour))
			reply.Members.PeerNames = tube.NewOmap[string, *tube.PeerDetailPlus]()
			reply.Members.PeerNames.Set("local", &tube.PeerDetailPlus{Det: &tube.PeerDetail{Name: "local", PeerID: "peer-local"}, RMemberLeaseUntilTm: time.Now().Add(time.Second)})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				done <- m.Run(ctx, member, func(context.Context) (*tube.PingReply, error) { return reply, nil })
			}()
			synctest.Wait()
			if _, err := m.Current(); err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Second)
			if _, err := m.Current(); !errors.Is(err, ErrAuthorityExpired) {
				t.Fatalf("expired local operating lease = %v", err)
			}
			cancel()
			<-done
		})
	})
}

func TestRMemberMonitorHasOneConsumer(t *testing.T) {
	t.Run("given an active event consumer then reject a second consumer without stealing membership messages", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			m := NewRMemberMonitor(0, time.Now)
			member := &tube.RMember{Ready: idem.NewIdemCloseChan(), UpcallMembershipChangeCh: make(chan *tube.PingReply, 10), OperatingLeaseRenewCh: make(chan time.Time, 10)}
			member.Ready.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			refresh := func(ctx context.Context) (*tube.PingReply, error) { <-ctx.Done(); return nil, ctx.Err() }
			done := make(chan error, 1)
			go func() { done <- m.Run(ctx, member, refresh) }()
			synctest.Wait()
			other, stopOther := context.WithCancel(context.Background())
			stopOther()
			if err := m.Run(other, member, refresh); !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("second consumer = %v", err)
			}
			cancel()
			<-done
		})
	})
}
