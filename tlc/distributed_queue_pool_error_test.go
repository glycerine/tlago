package tlc

import (
	"path/filepath"
	"strings"
	"testing"
)

// No enabled original method tests the queue's synchronous pool-error catches.
func TestDistributedDiskQueuePoolFailureUsesCodedRuntimeError(t *testing.T) {
	for _, operation := range []string{"write", "read", "peek"} {
		t.Run(operation, func(t *testing.T) {
			name := filepath.Join(t.TempDir(), "missing", "0")
			state := &TLCStateMut{UID: 41}
			queue := &DiskStateQueue{diskdir: filepath.Dir(name), len: 1, loPool: 1, hiPool: 3, loFile: name, deqIndex: 1, deqBuf: []*TLCStateMut{state}, enqIndex: 1, enqBuf: []*TLCStateMut{state}}
			queue.reader = NewStatePoolReader(1, name)
			queue.writer = NewStatePoolWriter(1, nil)
			if operation == "write" {
				if _, err := queue.writer.DoWork([]*TLCStateMut{state}, name); err != nil {
					t.Fatal(err)
				}
			}
			var failure any
			func() {
				defer func() { failure = recover() }()
				switch operation {
				case "write":
					queue.Enqueue(&TLCStateMut{UID: 97})
				case "read":
					queue.Dequeue()
				case "peek":
					queue.peekInner()
				}
			}()
			code := ECSystemErrorReadingStates
			if operation == "write" {
				code = ECSystemErrorWritingStates
			}
			coded, ok := failure.(*TLCError)
			if !ok || !coded.Runtime || coded.Code != code || len(coded.Params) != 2 || coded.Params[0] != "queue" || !strings.Contains(coded.Params[1], name) || coded.Cause != nil {
				t.Fatalf("queue %s failure lost source assertion contract: %T %v", operation, failure, failure)
			}
			if queue.len != 1 || queue.enqIndex != 1 || queue.deqIndex != 1 || queue.hiPool != 3 || queue.loPool != 1 || queue.enqBuf[0] != state || queue.deqBuf[0] != state {
				t.Fatal("failed pool operation advanced queue or replaced existing states")
			}
		})
	}
}
