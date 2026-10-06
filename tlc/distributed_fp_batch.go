package tlc

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/glycerine/tlago/jobcoord"
)

// FingerprintBatchResource binds TLC's conditional insertion semantics to one
// reusable replay stream. Create it for one shard in one fenced execution and
// preserve it across reconnects. The transport must validate authority before
// calling Insert or RetireThrough; this adapter does not elect a coordinator.
// After process loss recover the FPSet and recreate the adapter in a new execution.
type FingerprintBatchResource struct {
	set       FPSet
	stream    *jobcoord.ReplayStream
	slotBytes int
}

func NewFingerprintBatchResource(set FPSet, maxRecords, maxBytes int) (*FingerprintBatchResource, error) {
	if set == nil {
		return nil, fmt.Errorf("%w: missing fingerprint set", jobcoord.ErrInvalidRequest)
	}
	stream, err := jobcoord.NewReplayStream(maxRecords, maxBytes)
	if err != nil {
		return nil, err
	}
	return &FingerprintBatchResource{set: set, stream: stream, slotBytes: maxBytes / maxRecords}, nil
}

// Insert returns the original bitmap of fingerprints newly inserted by seq,
// including on retry. A different sequence competing for the same fingerprint
// receives the ordinary FPSet already-present answer. Vector order and all signed
// Java long bit patterns are preserved; no sorting or duplicate removal is added.
func (r *FingerprintBatchResource) Insert(ctx context.Context, seq uint64, fps []uint64) (*BitVector, error) {
	count := len(fps)
	replyBytes := count / 8
	if count%8 != 0 {
		replyBytes++
	}
	if replyBytes > r.slotBytes || count > (r.slotBytes-replyBytes)/8 {
		return nil, jobcoord.ErrBackpressure
	}
	request := make([]byte, count*8)
	for i, fp := range fps {
		binary.BigEndian.PutUint64(request[i*8:], fp)
	}
	reply, err := r.stream.Execute(ctx, seq, request, replyBytes, func(payload []byte) ([]byte, error) {
		vector := NewLongVecWithCapacity(len(payload) / 8)
		for i := 0; i < len(payload); i += 8 {
			vector.AddElement(int64(binary.BigEndian.Uint64(payload[i:])))
		}
		inserted := r.set.PutBlock(vector)
		if inserted == nil {
			return nil, fmt.Errorf("fingerprint insertion returned no bitmap")
		}
		bitmap := make([]byte, replyBytes)
		for i := 0; i < count; i++ {
			if inserted.Get(i) {
				bitmap[i/8] |= 1 << uint(i%8)
			}
		}
		return bitmap, nil
	})
	if err != nil {
		return nil, err
	}
	if len(reply) != replyBytes {
		return nil, fmt.Errorf("invalid retained fingerprint bitmap length")
	}
	result := NewBitVector(count)
	for i := 0; i < count; i++ {
		if reply[i/8]&(1<<uint(i%8)) != 0 {
			result.Set(i)
		}
	}
	return result, nil
}

// RetireThrough requires a fenced coordinator acknowledgment that all states
// associated with these insertion outcomes have been published to trace/queue.
func (r *FingerprintBatchResource) RetireThrough(seq uint64) error {
	return r.stream.RetireThrough(seq)
}
