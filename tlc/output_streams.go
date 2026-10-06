package tlc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

// DelayedPrintStream ports SpecProcessor's retained SANY output buffer.
type DelayedPrintStream struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	original io.Writer
}

func NewDelayedPrintStream(original io.Writer) *DelayedPrintStream {
	if original == nil {
		panic(NewNullPointerException("Null output stream"))
	}
	return &DelayedPrintStream{original: original}
}

func (s *DelayedPrintStream) Write(data []byte) (int, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.Write(data)
}

func (s *DelayedPrintStream) WriteByte(value byte) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.WriteByte(value)
}

func (s *DelayedPrintStream) Println(text string) {
	_, _ = fmt.Fprintln(s, text)
}

func (s *DelayedPrintStream) Release() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	defer func() {
		if failure := recover(); failure != nil {
			exception, ok := failure.(error)
			if !ok || isJavaError(exception) {
				panic(failure)
			}
			// Source catches Exception, prints to System.err regardless of debug,
			// and retains the buffer when an unchecked write/flush failure occurs.
			_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(exception))
		}
	}()
	// ByteArrayOutputStream synchronizes snapshot/reset separately. Original
	// writes may call back into this buffer; release does not hold its lock.
	s.mu.Lock()
	data := append([]byte(nil), s.buffer.Bytes()...)
	s.mu.Unlock()
	// original is a PrintStream in Java and swallows write/flush IOExceptions.
	_, _ = s.original.Write(data)
	if stream, ok := s.original.(interface{ Flush() error }); ok {
		_ = stream.Flush()
	}
	s.mu.Lock()
	s.buffer.Reset()
	s.mu.Unlock()
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
