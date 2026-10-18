//go:build tlc_verify

// Copyright (c) 2026 NVIDIA Corp. All rights reserved.
package tlc

import "sync"

// Original test-verify-model PoolFiles, restricted to the IValue stream
// operations actually used by DummyTLCState's base header. One model instance
// represents the original fresh JVM's single-name static pool registry.
type javaQueuePoolFiles struct {
	mu                      sync.Mutex
	hasName                 bool
	name                    string
	values                  []int64
	committed, writing      bool
	readers, completedReads int
}

func javaQueuePoolViolation(message string) error {
	// The source Violation extends Error so worker catches for Exception cannot
	// swallow it. Retain that Error boundary without adding a production type.
	return NewJavaError("PoolFiles.Violation: " + message)
}

func (p *javaQueuePoolFiles) pool(name string) error {
	if p.hasName {
		if p.name != name {
			return javaQueuePoolViolation("The JPF pool-file model supports at most 1 files.")
		}
	} else {
		p.hasName = true
		p.name = name
	}
	return nil
}

func (p *javaQueuePoolFiles) openOutput(name string) (statePoolOutputStream, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.pool(name); err != nil {
		return nil, err
	}
	if p.writing {
		return nil, javaQueuePoolViolation("Pool file " + name + " is already open for writing.")
	}
	if p.readers > 0 {
		return nil, javaQueuePoolViolation("Pool file " + name + " is open for reading, so writing it now is a read-write conflict.")
	}
	p.writing = true
	return &javaQueuePoolOutput{pool: p, name: name, values: make([]int64, 0, 8)}, nil
}

func (p *javaQueuePoolFiles) openInput(name string) (statePoolInputStream, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.pool(name); err != nil {
		return nil, err
	}
	if p.writing {
		return nil, javaQueuePoolViolation("Pool file " + name + " is open for writing, so reading it now is a read-write conflict.")
	}
	if !p.committed {
		return nil, javaQueuePoolViolation("Pool file " + name + " has never been written.")
	}
	p.readers++
	return &javaQueuePoolInput{pool: p, name: name, values: append([]int64(nil), p.values...)}, nil
}

type javaQueuePoolOutput struct {
	pool   *javaQueuePoolFiles
	name   string
	values []int64
	closed bool
}

func (s *javaQueuePoolOutput) record(value int64) error {
	if s.closed {
		return NewIOException("Pool file " + s.name + " is closed.")
	}
	s.values = append(s.values, value)
	return nil
}
func (s *javaQueuePoolOutput) WriteShortNat(value int16) error { return s.record(int64(value)) }
func (s *javaQueuePoolOutput) WriteLongNat(value int64) error  { return s.record(value) }
func (s *javaQueuePoolOutput) Write(Value) error {
	return NewUnsupportedOperationException("Value de-duplication is not modeled; a state on the queue of a JPF harness has no values.")
}
func (s *javaQueuePoolOutput) Close() error {
	if s.closed {
		return NewIOException("Pool file " + s.name + " is already closed.")
	}
	s.closed = true
	s.pool.mu.Lock()
	defer s.pool.mu.Unlock()
	if err := s.pool.pool(s.name); err != nil {
		return err
	}
	s.pool.values = append([]int64(nil), s.values...)
	s.pool.committed = true
	s.pool.writing = false
	return nil
}

type javaQueuePoolInput struct {
	pool   *javaQueuePoolFiles
	name   string
	values []int64
	index  int
	closed bool
}

func (s *javaQueuePoolInput) next() (int64, error) {
	if s.closed {
		return 0, NewIOException("Pool file " + s.name + " is closed.")
	}
	if s.index == len(s.values) {
		return 0, javaQueuePoolViolation("Pool file " + s.name + " has no values left for this read.")
	}
	value := s.values[s.index]
	s.index++
	return value, nil
}
func (s *javaQueuePoolInput) ReadShortNat() (int16, error) {
	value, err := s.next()
	return int16(value), err
}
func (s *javaQueuePoolInput) ReadLongNat() (int64, error) { return s.next() }
func (s *javaQueuePoolInput) Read() (Value, error) {
	return nil, NewUnsupportedOperationException("A state on the queue of a JPF harness has no values to read.")
}
func (s *javaQueuePoolInput) Close() error {
	if s.closed {
		return NewIOException("Pool file " + s.name + " is already closed.")
	}
	s.closed = true
	s.pool.mu.Lock()
	defer s.pool.mu.Unlock()
	if err := s.pool.pool(s.name); err != nil {
		return err
	}
	s.pool.readers--
	if s.index != len(s.values) {
		return javaQueuePoolViolation("The buffer that wrote the pool and the buffer reading it are not the same size.")
	}
	s.pool.completedReads++
	return nil
}
