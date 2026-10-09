package tlc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// No original queue method exercises the raw reader/writer failure catches.
// Run their actual loop in a child goroutine; failure must exit the process,
// rather than returning to this caller or leaving consumers waiting forever.
func TestDistributedBytePoolBackgroundFailureExit(t *testing.T) {
	if phase := os.Getenv("TLAGO_BYTE_POOL_FAILURE_EXIT"); phase != "" {
		AddMessageRecorder(RecorderFunc(func(message Message) {
			fmt.Fprintf(os.Stdout, "RAW_POOL_EVENT %d %q\n", message.Code, message.Params)
		}))
		defer fmt.Fprintln(os.Stdout, "RAW_POOL_DEFER_RAN")
		name := filepath.Join(os.Getenv("TLAGO_BYTE_POOL_FAILURE_DIRECTORY"), "0")
		if strings.HasSuffix(phase, "missing") {
			name = filepath.Join(filepath.Dir(name), "missing", "0")
		}
		done := make(chan struct{})
		if strings.HasPrefix(phase, "reader_") {
			if phase == "reader_truncated" || phase == "reader_negative" {
				data := []byte{0}
				if phase == "reader_negative" {
					data = []byte{255, 255, 255, 255}
				}
				if err := os.WriteFile(name, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			reader := NewByteArrayPoolReader(1, name)
			reader.Wakeup()
			go func() { reader.run(); close(done) }()
		} else {
			writer := NewByteArrayPoolWriter(1, nil)
			if _, err := writer.DoWork([][]byte{nil}, name); err != nil {
				t.Fatal(err)
			}
			go func() { writer.run(); close(done) }()
		}
		<-done
		fmt.Fprintln(os.Stdout, "RAW_POOL_FAILURE_RETURNED")
		return
	}
	for _, phase := range []string{"reader_missing", "reader_truncated", "reader_negative", "writer_missing", "writer_nil"} {
		t.Run(phase, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDistributedBytePoolBackgroundFailureExit$", "-test.v")
			command.Env = append(os.Environ(), "TLAGO_BYTE_POOL_FAILURE_EXIT="+phase, "TLAGO_BYTE_POOL_FAILURE_DIRECTORY="+t.TempDir())
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 {
				t.Fatalf("raw pool failure should exit 1: %v\n%s", err, output)
			}
			code := ECSystemErrorWritingPool
			if strings.HasPrefix(phase, "reader_") {
				code = ECSystemErrorReadingPool
			}
			text := string(output)
			if strings.Count(text, "RAW_POOL_EVENT ") != 1 || !strings.Contains(text, fmt.Sprintf("RAW_POOL_EVENT %d ", code)) {
				t.Fatalf("missing or extra raw pool diagnostic:\n%s", output)
			}
			if strings.HasPrefix(phase, "reader_") && !strings.Contains(text, `"0"]`) {
				t.Fatalf("reader diagnostic lost pool basename:\n%s", output)
			}
			if strings.Contains(text, "RAW_POOL_DEFER_RAN") || strings.Contains(text, "RAW_POOL_FAILURE_RETURNED") {
				t.Fatalf("fatal raw pool failure returned or ran deferred cleanup:\n%s", output)
			}
		})
	}
}

func TestDistributedByteQueuePoolFailureUsesCodedRuntimeError(t *testing.T) {
	for _, operation := range []string{"write", "read", "peek", "write_nil", "read_negative"} {
		t.Run(operation, func(t *testing.T) {
			name := filepath.Join(t.TempDir(), "missing", "0")
			if operation == "write_nil" || operation == "read_negative" {
				name = filepath.Join(t.TempDir(), "0")
			}
			if operation == "read_negative" {
				if err := os.WriteFile(name, []byte{255, 255, 255, 255}, 0600); err != nil {
					t.Fatal(err)
				}
			}
			state := []byte{41}
			q := &DiskByteArrayQueue{diskdir: filepath.Dir(name), len: 1, loPool: 1, hiPool: 3,
				loFile: name, deqIndex: 1, deqBuf: [][]byte{state}, enqIndex: 1, enqBuf: [][]byte{state}}
			q.reader = NewByteArrayPoolReader(1, name)
			q.writer = NewByteArrayPoolWriter(1, nil)
			if strings.HasPrefix(operation, "write") {
				pending := state
				if operation == "write_nil" {
					pending = nil
				}
				if _, err := q.writer.DoWork([][]byte{pending}, name); err != nil {
					t.Fatal(err)
				}
			}
			var failure any
			func() {
				defer func() { failure = recover() }()
				switch operation {
				case "write", "write_nil":
					q.enqueueRaw([]byte{97})
				case "read", "read_negative":
					q.dequeueRaw()
				case "peek":
					q.peekRaw()
				}
			}()
			code := ECSystemErrorReadingStates
			if strings.HasPrefix(operation, "write") {
				code = ECSystemErrorWritingStates
			}
			message := name
			if operation == "write_nil" {
				message = javaThrowableString(NewNullPointerException())
			} else if operation == "read_negative" {
				message = "-1"
			}
			coded, ok := failure.(*TLCError)
			if !ok || !coded.Runtime || coded.Code != code || len(coded.Params) != 2 || coded.Params[0] != "queue" || !strings.Contains(coded.Params[1], message) || coded.Cause != nil {
				t.Fatalf("raw queue %s lost source assertion contract: %T %v", operation, failure, failure)
			}
			if q.len != 1 || q.enqIndex != 1 || q.deqIndex != 1 || q.hiPool != 3 || q.loPool != 1 || &q.enqBuf[0][0] != &state[0] || &q.deqBuf[0][0] != &state[0] {
				t.Fatal("failed pool operation advanced queue or replaced its existing bytes")
			}
		})
	}
}

func TestDistributedByteQueueFailedCheckpointDoesNotFlushOrPublish(t *testing.T) {
	q := byteQueuePathFixture(t.TempDir())
	q.loPool, q.newLastLoPool, q.enqIndex = 5, 2, 1
	// A source null slot fails before the buffered header is flushed and before
	// newLastLoPool is assigned. Native ownership still closes the raw file.
	defer func() {
		if _, ok := recover().(*NullPointerException); !ok {
			t.Fatal("checkpoint null entry lost source runtime failure")
		}
		data, err := os.ReadFile(q.queuePath("queue.tmp"))
		if err != nil || len(data) != 0 || q.newLastLoPool != 2 {
			t.Fatalf("failed checkpoint flushed/published incomplete data: %v/%v, marker %d", data, err, q.newLastLoPool)
		}
	}()
	_ = q.BeginChkpt()
}
