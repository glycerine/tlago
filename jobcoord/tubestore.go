package jobcoord

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/glycerine/rpc25519/tube"
)

var (
	ErrControlConflict   = errors.New("job control record version conflict")
	ErrControlUnresolved = errors.New("job control mutation outcome unresolved")
)

// TubeRecordStore provides linearizable reads and guarded writes to an existing
// Tube key. Payloads are opaque application records. The lifecycle layer must
// validate Czar authority and application transitions before preparing mutations.
// Bootstrap/atomic creation of a new job is a separate operation: this API never
// treats a zero version as permission to perform an unconditional Tube write.
type TubeRecordStore struct {
	node  *tube.TubeNode
	table tube.Key
	key   tube.Key
	read func(context.Context) (*tube.Ticket, error)
	newSession func(context.Context) (*tube.Session, error)
	compare func(context.Context, *tube.Session, int64, []byte) (*tube.Ticket, error)
	closeSession func(*tube.Session) error
}

type TubeRecord struct {
	Payload []byte
	Version int64
}

func NewTubeRecordStore(node *tube.TubeNode, table, key string) *TubeRecordStore {
	s := &TubeRecordStore{node: node, table: tube.Key(table), key: tube.Key(key)}
	if node != nil {
		s.read = func(ctx context.Context) (*tube.Ticket, error) { return node.Read(ctx, s.table, s.key, 0, nil) }
		s.newSession = func(ctx context.Context) (*tube.Session, error) {
			sess, _, err := node.CreateNewSession(ctx, "", "")
			return sess, err
		}
		s.compare = func(ctx context.Context, sess *tube.Session, version int64, payload []byte) (*tube.Ticket, error) {
			return node.CAS(ctx, s.table, s.key, nil, payload, 0, sess, "jobcoord/control-v1", 0, false, version, 0)
		}
		s.closeSession = func(sess *tube.Session) error { return sess.Close() }
	}
	return s
}

func (s *TubeRecordStore) Load(ctx context.Context) (TubeRecord, error) {
	if s.read == nil || s.table == "" || s.key == "" {
		return TubeRecord{}, ErrInvalidRequest
	}
	tkt, err := s.read(ctx)
	if err != nil {
		return TubeRecord{}, err
	}
	if tkt == nil || tkt.VersionRead <= 0 {
		return TubeRecord{}, fmt.Errorf("invalid Tube control read")
	}
	return TubeRecord{Payload: bytes.Clone(tkt.Val), Version: tkt.VersionRead}, nil
}

// PrepareCAS reserves one Tube session and one fixed serial for this logical
// mutation. Each retry uses that same pair, including after an ambiguous reply.
// Session.CAS cannot be called afresh for retries: it increments SessionSerial.
// Different logical mutations receive separate sessions, so acknowledging a
// later operation cannot retire an unresolved earlier operation's Tube outcome.
// The control plane caps outstanding mutations; this is not a bulk data API.
func (s *TubeRecordStore) PrepareCAS(ctx context.Context, expectedVersion int64, payload []byte) (*TubeControlMutation, error) {
	if s.newSession == nil || s.compare == nil || s.closeSession == nil || s.table == "" || s.key == "" || expectedVersion <= 0 {
		return nil, ErrInvalidRequest
	}
	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("missing Tube control session")
	}
	session.SessionSerial = 1
	value := bytes.Clone(payload)
	m := &TubeControlMutation{session: session, close: func() error { return s.closeSession(session) }}
	m.cas = func(ctx context.Context) (*tube.Ticket, error) {
		return s.compare(ctx, session, expectedVersion, value)
	}
	return m, nil
}

// TubeControlMutation owns a fixed CAS request and its unresolved obligation.
// Execute retries the same server-side operation, never another mutation. Its
// cached reply is execution-scoped; process failure requires job recovery. A Tube
// session expiration/transport error leaves the outcome unresolved, not a reason
// to create a fresh session and repeat the old write.
type TubeControlMutation struct {
	close     func() error
	mu        sync.Mutex
	session   *tube.Session
	cas       func(context.Context) (*tube.Ticket, error)
	submitted bool
	known     bool
	closed    bool
	swapped   bool
	outcome   error
}

func (m *TubeControlMutation) Execute(ctx context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return false, ErrRetired
	}
	if m.known {
		return m.swapped, m.outcome
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	m.submitted = true
	tkt, err := m.cas(ctx)
	if err == nil && tkt != nil && tkt.CASwapped {
		m.known, m.swapped = true, true
		return true, nil
	}
	// Tube uses this diagnostic prefix in doCAS, also recognized by RMember.
	// A version rejection is a known, repeatable outcome; connection/session errors
	// are not. Do not mistake some other rejected write for a CAS version conflict.
	if tkt != nil && !tkt.CASwapped && err != nil && strings.HasPrefix(err.Error(), "rejected write CAS on OldVersionCAS=") {
		m.known = true
		m.outcome = errors.Join(ErrControlConflict, err)
		return false, m.outcome
	}
	if err == nil {
		err = fmt.Errorf("Tube returned no definitive CAS outcome")
	}
	return false, errors.Join(ErrControlUnresolved, err)
}

// Close releases a known (or never-submitted) client's session resources. It
// refuses to discard an unresolved mutation. Job failure/recovery may abandon
// such an obligation only by fencing the entire execution.
func (m *TubeControlMutation) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	if m.submitted && !m.known {
		return ErrControlUnresolved
	}
	m.closed = true
	return m.close()
}
