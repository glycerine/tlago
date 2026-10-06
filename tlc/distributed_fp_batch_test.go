package tlc

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/glycerine/tlago/jobcoord"
)

func TestFingerprintBatchLostReply(t *testing.T) {
	t.Run("given TLC fingerprints inserted before reply loss then retry preserves the publication bitmap", func(t *testing.T) {
		set := NewMemFPSet()
		r, err := NewFingerprintBatchResource(set, 8, 8192)
		if err != nil {
			t.Fatal(err)
		}
		fps := []uint64{7, math.MaxUint64, 7, 1 << 63, 0}
		first, err := r.Insert(context.Background(), 1, fps)
		if err != nil {
			t.Fatal(err)
		}
		if !first.Get(0) || !first.Get(1) || first.Get(2) || !first.Get(3) || !first.Get(4) {
			t.Fatalf("original bitmap = %s", first.String())
		}
		first.Reset(0)
		replay, err := r.Insert(context.Background(), 1, fps)
		if err != nil || !replay.Get(0) || !replay.Get(1) || replay.Get(2) || !replay.Get(3) || !replay.Get(4) {
			t.Fatalf("replay = %v, %v", replay, err)
		}
		if set.Size() != 4 {
			t.Fatalf("distinct fingerprints = %d", set.Size())
		}
		other, err := r.Insert(context.Background(), 2, fps)
		if err != nil || other.TrueCount() != 0 {
			t.Fatalf("different operation = %v, %v", other, err)
		}
		if err := r.RetireThrough(2); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Insert(context.Background(), 1, fps); !errors.Is(err, jobcoord.ErrRetired) {
			t.Fatalf("late replay = %v", err)
		}
	})
}

func TestFingerprintBatchCapacityBeforeMutation(t *testing.T) {
	t.Run("given insufficient reservation then reject before TLC fingerprints become visited", func(t *testing.T) {
		set := NewMemFPSet()
		r, err := NewFingerprintBatchResource(set, 1, 8)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Insert(context.Background(), 1, []uint64{42}); !errors.Is(err, jobcoord.ErrBackpressure) {
			t.Fatalf("capacity = %v", err)
		}
		if set.Size() != 0 {
			t.Fatal("fingerprint inserted without reply reservation")
		}
	})
}
