package tlc

import (
	"bytes"
	"io"
	"sync"
)

type DelayedPrintStream struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	original io.Writer
}

func NewDelayedPrintStream(original io.Writer) *DelayedPrintStream {
	return &DelayedPrintStream{original: original}
}

func (s *DelayedPrintStream) Write(p []byte) (int, error) {
	if s == nil {
		return 0, io.ErrClosedPipe
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.Write(p)
}

func (s *DelayedPrintStream) Release() error {
	if s == nil {
		return io.ErrClosedPipe
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.original == nil {
		return io.ErrClosedPipe
	}
	if _, err := s.original.Write(s.buffer.Bytes()); err != nil {
		return err
	}
	if flusher, ok := s.original.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return err
		}
	}
	s.buffer.Reset()
	return nil
}

type TeeOutputStream struct {
	mu     sync.Mutex
	out    io.Writer
	branch io.Writer
}

func NewTeeOutputStream(out io.Writer, branch io.Writer) *TeeOutputStream {
	return &TeeOutputStream{out: out, branch: branch}
}

func (s *TeeOutputStream) Write(p []byte) (int, error) {
	if s == nil {
		return 0, io.ErrClosedPipe
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.out == nil || s.branch == nil {
		return 0, io.ErrClosedPipe
	}
	n, err := s.out.Write(p)
	if err != nil {
		return n, err
	}
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	n, err = s.branch.Write(p)
	if err != nil {
		return n, err
	}
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	return len(p), nil
}

func (s *TeeOutputStream) Flush() error {
	if s == nil {
		return io.ErrClosedPipe
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if flusher, ok := s.out.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return err
		}
	}
	if flusher, ok := s.branch.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

func (s *TeeOutputStream) Close() error {
	if s == nil {
		return io.ErrClosedPipe
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var outErr error
	if closer, ok := s.out.(io.Closer); ok {
		outErr = closer.Close()
	}
	if closer, ok := s.branch.(io.Closer); ok {
		if branchErr := closer.Close(); branchErr != nil {
			return branchErr
		}
	}
	return outErr
}
