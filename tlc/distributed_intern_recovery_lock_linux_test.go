//go:build linux

package tlc

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
)

// There is no original direct InternTable recovery test. A FIFO handshake
// proves recovery has opened its input while its header read is still blocked,
// without a timing sleep or production hook. Native recovery must already
// own the interning lock at this point; Java uses distinct instance/class locks.
func TestDistributedInternRecoveryLocksBeforeHeaderRead(t *testing.T) {
	for _, truncated := range []bool{false, true} {
		name := "complete"
		if truncated {
			name = "truncated"
		}
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			path := uniqueStringChkptName(directory, "chkpt")
			if err := syscall.Mkfifo(path, 0600); err != nil {
				t.Fatal(err)
			}
			table := NewInternTable(16)
			table.tokenCnt = 11
			done := make(chan error, 1)
			go func() { done <- table.Recover(directory) }()
			// Opening the writer completes only after the recovery reader has
			// opened the FIFO. With no bytes written, its header cannot finish.
			writer, err := os.OpenFile(path, os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			joined := false
			defer func() {
				_ = writer.Close()
				if !joined {
					<-done
				}
			}()
			if table.mu.TryLock() {
				table.mu.Unlock()
				t.Fatal("recovery read its header without owning the interning lock")
			}
			payload := make([]byte, 4)
			binary.BigEndian.PutUint32(payload, 7)
			if truncated {
				payload = payload[:2]
			}
			if _, err := writer.Write(payload); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			err = <-done
			joined = true
			wantNextToken := 8
			if truncated {
				if !errors.Is(err, io.EOF) {
					t.Fatalf("truncated header: %v", err)
				}
				wantNextToken = 12
			} else if err != nil {
				t.Fatal(err)
			}
			if !table.mu.TryLock() {
				t.Fatal("recovery retained the interning lock after return")
			}
			table.mu.Unlock()
			if got := table.Put("after-recovery").Token(); got != wantNextToken {
				t.Fatalf("next token = %d, want %d", got, wantNextToken)
			}
		})
	}
}
