package tlc

import (
	"bytes"
	"errors"
	"testing"
)

func TestDelayedPrintStreamBuffersUntilRelease(t *testing.T) {
	var original flushBuffer
	stream := NewDelayedPrintStream(&original)
	if _, err := stream.Write([]byte("one")); err != nil {
		t.Fatalf("Write one: %v", err)
	}
	if _, err := stream.Write([]byte(" two")); err != nil {
		t.Fatalf("Write two: %v", err)
	}
	if got := original.String(); got != "" {
		t.Fatalf("original before release = %q, want empty", got)
	}
	if err := stream.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if got := original.String(); got != "one two" {
		t.Fatalf("original after release = %q, want one two", got)
	}
	if original.flushes != 1 {
		t.Fatalf("flushes = %d, want 1", original.flushes)
	}
	if err := stream.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}
	if got := original.String(); got != "one two" {
		t.Fatalf("second release wrote buffered data again: %q", got)
	}
}

func TestTeeOutputStreamWritesFlushesAndClosesBothBranches(t *testing.T) {
	var left flushCloseBuffer
	var right flushCloseBuffer
	tee := NewTeeOutputStream(&left, &right)
	if n, err := tee.Write([]byte("abc")); err != nil || n != 3 {
		t.Fatalf("Write = %d, %v; want 3, nil", n, err)
	}
	if left.String() != "abc" || right.String() != "abc" {
		t.Fatalf("tee write left=%q right=%q, want both abc", left.String(), right.String())
	}
	if err := tee.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if left.flushes != 1 || right.flushes != 1 {
		t.Fatalf("flushes left=%d right=%d, want 1/1", left.flushes, right.flushes)
	}
	if err := tee.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !left.closed || !right.closed {
		t.Fatalf("closed left=%v right=%v, want true/true", left.closed, right.closed)
	}
}

func TestTeeOutputStreamCloseReturnsBranchErrorLikeJavaFinally(t *testing.T) {
	mainErr := errors.New("main")
	branchErr := errors.New("branch")
	left := &flushCloseBuffer{closeErr: mainErr}
	right := &flushCloseBuffer{closeErr: branchErr}
	if err := NewTeeOutputStream(left, right).Close(); !errors.Is(err, branchErr) {
		t.Fatalf("Close error = %v, want branch error", err)
	}
	if !left.closed || !right.closed {
		t.Fatalf("closed left=%v right=%v, want true/true", left.closed, right.closed)
	}
}

type flushBuffer struct {
	bytes.Buffer
	flushes int
}

func (b *flushBuffer) Flush() error {
	b.flushes++
	return nil
}

type flushCloseBuffer struct {
	bytes.Buffer
	flushes  int
	closed   bool
	closeErr error
}

func (b *flushCloseBuffer) Flush() error {
	b.flushes++
	return nil
}

func (b *flushCloseBuffer) Close() error {
	b.closed = true
	return b.closeErr
}
