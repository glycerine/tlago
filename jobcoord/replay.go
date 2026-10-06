// Package jobcoord implements reusable distributed job lifecycle primitives.
// These local primitives do not by themselves provide network or crash recovery.
package jobcoord

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
)

var (
	ErrIdentityConflict = errors.New("operation identity reused with different contents")
	ErrRetired          = errors.New("operation retired")
	ErrNotCompleted     = errors.New("retirement includes unfinished operations")
	ErrBackpressure     = errors.New("replay capacity exhausted")
	ErrPoisoned         = errors.New("mutation outcome uncertain: execution must recover")
	ErrInvalidRequest   = errors.New("invalid replay request")
)

type replayRecord struct {
	request  []byte
	maxReply int
	apply    func([]byte) ([]byte, error)
	done     chan struct{}
	reply    []byte
	err      error
}

// ReplayStream is one ordered mutation stream in one externally fenced execution.
// Bind it to a job/execution/resource/stream identity before admitting requests.
// It retains original replies until the coordinator acknowledges publication.
// Discarding it on reconnect would violate the protocol; after process failure,
// restore a coordinated snapshot in a new execution rather than replay here.
//
// A fixed per-record byte reservation prevents out-of-order requests from using
// all capacity needed to admit a missing earlier sequence. Payload plus declared
// maximum reply must fit that slot. Record overhead is bounded by MaxRecords.
type ReplayStream struct {
	mu         sync.Mutex
	records    map[uint64]*replayRecord
	next       uint64
	retired    uint64
	maxRecords int
	slotBytes  int
	running    bool
	poisoned   error
}

func NewReplayStream(maxRecords, maxBytes int) (*ReplayStream, error) {
	if maxRecords <= 0 || maxBytes < maxRecords {
		return nil, fmt.Errorf("%w: positive record and byte capacity required", ErrInvalidRequest)
	}
	return &ReplayStream{records: make(map[uint64]*replayRecord), next: 1, maxRecords: maxRecords, slotBytes: maxBytes / maxRecords}, nil
}

// Execute admits a sequence and waits for its immutable original reply. Sequence
// numbers start at one; duplicates use the same logical request bytes. apply is
// invoked once, serially with other mutations. It receives a private copy.
// The first admitted caller owns the callback; duplicates never replace it.
// Cancellation stops waiting, not an already accepted mutation. Callbacks must
// finish independently of that caller and return an outcome or explicit error.
func (s *ReplayStream) Execute(ctx context.Context, seq uint64, request []byte, maxReply int, apply func([]byte) ([]byte, error)) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if seq == 0 || seq == math.MaxUint64 || maxReply < 0 || apply == nil {
		return nil, ErrInvalidRequest
	}
	s.mu.Lock()
	if s.poisoned != nil {
		err := s.poisoned
		s.mu.Unlock()
		return nil, err
	}
	if seq <= s.retired {
		s.mu.Unlock()
		return nil, ErrRetired
	}
	r, exists := s.records[seq]
	if exists {
		if !bytes.Equal(r.request, request) {
			s.mu.Unlock()
			return nil, ErrIdentityConflict
		}
	} else {
		if seq < s.next {
			s.mu.Unlock()
			return nil, ErrRetired
		}
		if seq-s.next >= uint64(s.maxRecords) || len(s.records) >= s.maxRecords || maxReply > s.slotBytes || len(request) > s.slotBytes-maxReply {
			s.mu.Unlock()
			return nil, ErrBackpressure
		}
		r = &replayRecord{request: bytes.Clone(request), maxReply: maxReply, apply: apply, done: make(chan struct{})}
		s.records[seq] = r
	}
	if !s.running && s.records[s.next] != nil {
		s.running = true
		go s.run()
	}
	s.mu.Unlock()
	select {
	case <-r.done:
		return bytes.Clone(r.reply), r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *ReplayStream) run() {
	for {
		s.mu.Lock()
		r := s.records[s.next]
		if r == nil || s.poisoned != nil {
			s.running = false
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
		reply, err := invokeMutation(r.apply, bytes.Clone(r.request))
		if err == nil && len(reply) > r.maxReply {
			err = fmt.Errorf("reply exceeds reserved capacity: %d > %d", len(reply), r.maxReply)
		}
		s.mu.Lock()
		if err != nil {
			s.poisoned = errors.Join(ErrPoisoned, err)
			for _, pending := range s.records {
				select {
				case <-pending.done:
				default:
					pending.err = s.poisoned
					pending.apply = nil
					close(pending.done)
				}
			}
			s.running = false
			s.mu.Unlock()
			return
		}
		r.reply = bytes.Clone(reply)
		r.apply = nil
		s.next++
		close(r.done)
		s.mu.Unlock()
	}
}

func invokeMutation(apply func([]byte) ([]byte, error), request []byte) (reply []byte, err error) {
	defer func() {
		if p := recover(); p != nil {
			reply = nil
			err = fmt.Errorf("mutation panic: %v", p)
		}
	}()
	return apply(request)
}

// RetireThrough is the resource side of an AppliedThrough acknowledgment. The
// caller must authenticate/fence that acknowledgment and ensure publication is
// complete. This stream can check completion of mutations, not application
// publication. Repeated or older acknowledgments are harmless.
func (s *ReplayStream) RetireThrough(seq uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.poisoned != nil {
		return s.poisoned
	}
	if seq <= s.retired {
		return nil
	}
	if seq >= s.next {
		return ErrNotCompleted
	}
	for n := s.retired + 1; n <= seq; n++ {
		delete(s.records, n)
	}
	s.retired = seq
	return nil
}

type ReplaySnapshot struct {
	Records          int
	CompletedThrough uint64
	RetiredThrough   uint64
	Poisoned         bool
}

func (s *ReplayStream) Snapshot() ReplaySnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ReplaySnapshot{Records: len(s.records), CompletedThrough: s.next - 1, RetiredThrough: s.retired, Poisoned: s.poisoned != nil}
}
