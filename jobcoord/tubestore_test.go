package jobcoord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glycerine/rpc25519/tube"
)

func liveTube(t *testing.T) (*tube.TubeCluster, int) {
	t.Helper()
	cfg := tube.NewTubeConfigTest(3, t.Name(), false)
	t.Log("starting three bootstrapped Tube voters over real loopback circuits")
	c, name, leader, _ := tube.SetupTestClusterWithCustomConfig(cfg, t, 3, 0, 901)
	t.Cleanup(c.Close)
	t.Logf("leader %s committed its initial entry", name)
	return c, leader
}

func TestTubeRecordLostCASReply(t *testing.T) {
	t.Run("given a committed CAS with a lost reply then retry returns the original outcome after later writes", func(t *testing.T) {
		c, leader := liveTube(t)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		node := c.Nodes[(leader+1)%3]
		if _, err := node.Write(ctx, "jobs", "job-a", []byte("preparing"), 0, nil, "", 0, false); err != nil {
			t.Fatal(err)
		}
		store := NewTubeRecordStore(node, "jobs", "job-a")
		initial, err := store.Load(ctx)
		if err != nil {
			t.Fatal(err)
		}
		mutation, err := store.PrepareCAS(ctx, initial.Version, []byte("active"))
		if err != nil {
			t.Fatal(err)
		}
		defer mutation.Close()
		original := mutation.cas
		dropOnce := true
		mutation.cas = func(ctx context.Context) (*tube.Ticket, error) {
			reply, err := original(ctx)
			if err == nil && dropOnce {
				dropOnce = false
				return nil, context.DeadlineExceeded
			}
			return reply, err
		}
		if _, err := mutation.Execute(ctx); !errors.Is(err, ErrControlUnresolved) {
			t.Fatalf("lost reply = %v", err)
		}
		active, err := store.Load(ctx)
		if err != nil || string(active.Payload) != "active" {
			t.Fatalf("committed value = %+v, %v", active, err)
		}
		later, err := store.PrepareCAS(ctx, active.Version, []byte("completed"))
		if err != nil {
			t.Fatal(err)
		}
		defer later.Close()
		if swapped, err := later.Execute(ctx); err != nil || !swapped {
			t.Fatalf("later CAS = %v, %v", swapped, err)
		}
		if swapped, err := mutation.Execute(ctx); err != nil || !swapped {
			t.Fatalf("original outcome replay = %v, %v", swapped, err)
		}
		final, err := store.Load(ctx)
		if err != nil || string(final.Payload) != "completed" {
			t.Fatalf("replay overwrote newer record = %+v, %v", final, err)
		}
	})
}

func TestTubeRecordMajorityAfterLeaderLoss(t *testing.T) {
	t.Run("given a three-voter record then losing the leader preserves the record and rejects stale versions", func(t *testing.T) {
		c, leader := liveTube(t)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		node := c.Nodes[(leader+1)%3]
		if _, err := node.Write(ctx, "jobs", "job-b", []byte("preparing"), 0, nil, "", 0, false); err != nil {
			t.Fatal(err)
		}
		store := NewTubeRecordStore(node, "jobs", "job-b")
		initial, err := store.Load(ctx)
		if err != nil {
			t.Fatal(err)
		}
		mutation, err := store.PrepareCAS(ctx, initial.Version, []byte("active"))
		if err != nil {
			t.Fatal(err)
		}
		defer mutation.Close()
		if swapped, err := mutation.Execute(ctx); err != nil || !swapped {
			t.Fatalf("activation = %v, %v", swapped, err)
		}
		t.Log("stopping the active Raft leader; awaiting a majority-committed replacement")
		c.Nodes[leader].Close()
		for {
			select {
			case name := <-c.LeaderNoop0committedCh:
				if c.Name2num[name] == leader {
					continue
				}
				t.Logf("replacement leader %s committed its initial entry", name)
				goto elected
			case <-ctx.Done():
				t.Fatal("surviving majority did not elect and commit a replacement")
			}
		}
	elected:
		current, err := store.Load(ctx)
		if err != nil || string(current.Payload) != "active" {
			t.Fatalf("surviving value = %+v, %v", current, err)
		}
		stale, err := store.PrepareCAS(ctx, initial.Version, []byte("stale writer"))
		if err != nil {
			t.Fatal(err)
		}
		defer stale.Close()
		if swapped, err := stale.Execute(ctx); swapped || !errors.Is(err, ErrControlConflict) {
			t.Fatalf("stale writer = %v, %v", swapped, err)
		}
		if _, err := store.PrepareCAS(ctx, 0, []byte("unguarded")); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("unguarded CAS = %v", err)
		}
	})
}
