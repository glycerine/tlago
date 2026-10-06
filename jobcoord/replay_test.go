package jobcoord

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
)

func stream(t *testing.T, records, bytes int) *ReplayStream {
	t.Helper()
	s, err := NewReplayStream(records, bytes)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestReplayLostInsertionReply(t *testing.T) {
	t.Run("given a lost reply when retried then preserve the original new-key answer", func(t *testing.T) {
		s := stream(t, 8, 1024)
		seen := false
		insert := func([]byte) ([]byte, error) {
			if seen {
				return []byte{0}, nil
			}
			seen = true
			return []byte{1}, nil
		}
		first, err := s.Execute(context.Background(), 1, []byte("fingerprint"), 1, insert)
		if err != nil {
			t.Fatal(err)
		}
		first[0] = 99
		replay, err := s.Execute(context.Background(), 1, []byte("fingerprint"), 1, insert)
		if err != nil || len(replay) != 1 || replay[0] != 1 {
			t.Fatalf("replay = %v, %v", replay, err)
		}
		other, err := s.Execute(context.Background(), 2, []byte("fingerprint"), 1, insert)
		if err != nil || len(other) != 1 || other[0] != 0 {
			t.Fatalf("distinct operation = %v, %v", other, err)
		}
	})
}

func TestReplayConcurrentDuplicates(t *testing.T) {
	t.Run("given an executing operation when duplicates arrive then mutate once", func(t *testing.T) {
		s := stream(t, 8, 1024)
		entered, release := make(chan struct{}), make(chan struct{})
		var calls atomic.Int32
		apply := func([]byte) ([]byte, error) { calls.Add(1); close(entered); <-release; return []byte("winner"), nil }
		var wg sync.WaitGroup
		errs := make(chan error, 32)
		for range 32 {
			wg.Go(func() {
				reply, err := s.Execute(context.Background(), 1, []byte("same"), 6, apply)
				if err == nil && string(reply) != "winner" {
					err = errors.New("wrong reply")
				}
				errs <- err
			})
		}
		<-entered
		close(release)
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatal(err)
			}
		}
		if calls.Load() != 1 {
			t.Fatalf("mutations = %d", calls.Load())
		}
	})
}

func TestReplayIdentityConflict(t *testing.T) {
	t.Run("given a retained operation when its contents change then reject identity reuse", func(t *testing.T) {
		s := stream(t, 8, 1024)
		if _, err := s.Execute(context.Background(), 1, []byte("a"), 1, func([]byte) ([]byte, error) { return []byte{1}, nil }); err != nil {
			t.Fatal(err)
		}
		_, err := s.Execute(context.Background(), 1, []byte("b"), 1, func([]byte) ([]byte, error) { t.Error("conflict executed"); return nil, nil })
		if !errors.Is(err, ErrIdentityConflict) {
			t.Fatalf("conflict = %v", err)
		}
	})
}

func TestReplayRetirement(t *testing.T) {
	t.Run("given applied operations when retired then late requests never execute afresh", func(t *testing.T) {
		s := stream(t, 2, 10)
		apply := func([]byte) ([]byte, error) { return []byte{1}, nil }
		if err := s.RetireThrough(1); !errors.Is(err, ErrNotCompleted) {
			t.Fatalf("premature retirement = %v", err)
		}
		if _, err := s.Execute(context.Background(), 1, []byte{1}, 1, apply); err != nil {
			t.Fatal(err)
		}
		if err := s.RetireThrough(1); err != nil {
			t.Fatal(err)
		}
		if err := s.RetireThrough(1); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Execute(context.Background(), 1, []byte{1}, 1, apply); !errors.Is(err, ErrRetired) {
			t.Fatalf("late request = %v", err)
		}
		if _, err := s.Execute(context.Background(), 2, []byte{2}, 1, apply); err != nil {
			t.Fatal(err)
		}
	})
}

func TestReplayBackpressure(t *testing.T) {
	t.Run("given full capacity when another operation arrives then reject before mutation", func(t *testing.T) {
		s := stream(t, 1, 4)
		apply := func([]byte) ([]byte, error) { return []byte{1}, nil }
		if _, err := s.Execute(context.Background(), 1, []byte{1, 2}, 2, apply); err != nil {
			t.Fatal(err)
		}
		called := false
		if _, err := s.Execute(context.Background(), 2, []byte{3}, 1, func([]byte) ([]byte, error) { called = true; return nil, nil }); !errors.Is(err, ErrBackpressure) {
			t.Fatalf("budget = %v", err)
		}
		if called {
			t.Fatal("mutated without reservation")
		}
		if err := s.RetireThrough(1); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Execute(context.Background(), 2, []byte{3}, 1, apply); err != nil {
			t.Fatal(err)
		}
	})
}

func TestReplayOutOfOrder(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := stream(t, 8, 1024)
		var order []byte
		second := make(chan error, 1)
		go func() {
			_, err := s.Execute(context.Background(), 2, []byte{2}, 1, func(p []byte) ([]byte, error) { order = append(order, p[0]); return p, nil })
			second <- err
		}()
		synctest.Wait()
		if got := s.Snapshot(); got.Records != 1 || got.CompletedThrough != 0 {
			t.Fatalf("buffered later request = %+v", got)
		}
		if _, err := s.Execute(context.Background(), 1, []byte{1}, 1, func(p []byte) ([]byte, error) { order = append(order, p[0]); return p, nil }); err != nil {
			t.Fatal(err)
		}
		if err := <-second; err != nil {
			t.Fatal(err)
		}
		if string(order) != string([]byte{1, 2}) {
			t.Fatalf("execution order = %v", order)
		}
	})
}

func TestReplayCanceledWait(t *testing.T) {
	t.Run("given cancellation after admission then retain and finish the obligation", func(t *testing.T) {
		s := stream(t, 8, 1024)
		entered, release := make(chan struct{}), make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		returned := make(chan error, 1)
		go func() {
			_, err := s.Execute(ctx, 1, []byte("a"), 1, func([]byte) ([]byte, error) { close(entered); <-release; return []byte{1}, nil })
			returned <- err
		}()
		<-entered
		cancel()
		if err := <-returned; !errors.Is(err, context.Canceled) {
			t.Fatalf("wait = %v", err)
		}
		close(release)
		replay, err := s.Execute(context.Background(), 1, []byte("a"), 1, func([]byte) ([]byte, error) { t.Error("reexecuted"); return nil, nil })
		if err != nil || len(replay) != 1 || replay[0] != 1 {
			t.Fatalf("recovered reply = %v, %v", replay, err)
		}
	})
}

func TestReplayPartialMutationFailure(t *testing.T) {
	t.Run("given uncertain mutation failure then poison this execution instead of retrying", func(t *testing.T) {
		s := stream(t, 8, 1024)
		failure := errors.New("partial storage write")
		_, err := s.Execute(context.Background(), 1, []byte("a"), 1, func([]byte) ([]byte, error) { return nil, failure })
		if !errors.Is(err, ErrPoisoned) || !errors.Is(err, failure) {
			t.Fatalf("failure = %v", err)
		}
		_, err = s.Execute(context.Background(), 2, []byte("b"), 1, func([]byte) ([]byte, error) { t.Error("poisoned mutation executed"); return nil, nil })
		if !errors.Is(err, ErrPoisoned) {
			t.Fatalf("new operation = %v", err)
		}
	})
}

func TestReplayReplyBudgetViolation(t *testing.T) {
	t.Run("given a mutation exceeding its reserved reply size then require recovery", func(t *testing.T) {
		s := stream(t, 2, 32)
		_, err := s.Execute(context.Background(), 1, []byte{1}, 1, func([]byte) ([]byte, error) { return []byte{1, 2}, nil })
		if !errors.Is(err, ErrPoisoned) {
			t.Fatalf("oversized outcome = %v", err)
		}
		if !s.Snapshot().Poisoned {
			t.Fatal("uncertain operation left usable stream")
		}
	})
}

func TestReplayMutationPanic(t *testing.T) {
	t.Run("given a panicking adapter then unblock all buffered callers with recovery required", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			s := stream(t, 2, 32)
			later := make(chan error, 1)
			go func() {
				_, err := s.Execute(context.Background(), 2, []byte{2}, 1, func([]byte) ([]byte, error) { t.Error("later mutation ran"); return nil, nil })
				later <- err
			}()
			synctest.Wait()
			_, err := s.Execute(context.Background(), 1, []byte{1}, 1, func([]byte) ([]byte, error) { panic("storage failure") })
			if !errors.Is(err, ErrPoisoned) {
				t.Fatalf("panic = %v", err)
			}
			if err := <-later; !errors.Is(err, ErrPoisoned) {
				t.Fatalf("buffered caller = %v", err)
			}
		})
	})
}

func TestReplayGapCapacity(t *testing.T) {
	t.Run("given a later request using its byte slot then the missing predecessor still fits", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			s := stream(t, 2, 8)
			later := make(chan error, 1)
			apply := func(p []byte) ([]byte, error) { return []byte{p[0], p[1]}, nil }
			go func() { _, err := s.Execute(context.Background(), 2, []byte{2, 2}, 2, apply); later <- err }()
			synctest.Wait()
			if _, err := s.Execute(context.Background(), 1, []byte{1, 1}, 2, apply); err != nil {
				t.Fatal(err)
			}
			if err := <-later; err != nil {
				t.Fatal(err)
			}
		})
	})
}
