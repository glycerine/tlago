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

// No enabled original method tests the fatal StatePoolReader/Writer run catch.
// Child processes verify actual process exit without a production test hook.
func TestDistributedPoolBackgroundFailureExit(t *testing.T) {
	if phase := os.Getenv("TLAGO_POOL_FAILURE_EXIT"); phase != "" {
		initTLCCheckerTest(t)
		AddMessageRecorder(RecorderFunc(func(message Message) {
			fmt.Fprintf(os.Stdout, "POOL_EVENT %d %q\n", message.Code, message.Params)
		}))
		defer fmt.Fprintln(os.Stdout, "POOL_DEFER_RAN")
		directory := os.Getenv("TLAGO_POOL_FAILURE_DIRECTORY")
		name := filepath.Join(directory, "missing", "0")
		if phase == "reader_truncated" || phase == "writer_nil" {
			name = filepath.Join(directory, "0")
		}
		if strings.HasPrefix(phase, "reader_") {
			if phase == "reader_truncated" {
				if err := os.WriteFile(name, []byte{0}, 0600); err != nil {
					t.Fatal(err)
				}
			}
			reader := NewStatePoolReader(1, name)
			reader.Start()
			reader.Wakeup()
			<-reader.done
		} else {
			writer := NewStatePoolWriter(1, nil)
			writer.Start()
			if _, err := writer.DoWork([]*TLCStateMut{nil}, name); err != nil {
				t.Fatal(err)
			}
			<-writer.done
		}
		fmt.Fprintln(os.Stdout, "POOL_FAILURE_RETURNED")
		return
	}
	for _, phase := range []string{"reader_missing", "reader_truncated", "writer_missing", "writer_nil"} {
		t.Run(phase, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDistributedPoolBackgroundFailureExit$", "-test.v")
			command.Env = append(os.Environ(), "TLAGO_POOL_FAILURE_EXIT="+phase, "TLAGO_POOL_FAILURE_DIRECTORY="+t.TempDir())
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 {
				t.Fatalf("pool failure should exit 1: %v\n%s", err, output)
			}
			code := ECSystemErrorWritingPool
			if strings.HasPrefix(phase, "reader_") {
				code = ECSystemErrorReadingPool
			}
			text := string(output)
			if strings.Count(text, "POOL_EVENT ") != 1 || !strings.Contains(text, fmt.Sprintf("POOL_EVENT %d ", code)) {
				t.Fatalf("missing or extra source pool diagnostic:\n%s", output)
			}
			if strings.HasPrefix(phase, "reader_") && !strings.Contains(text, `"0"]`) {
				t.Fatalf("reader diagnostic lost pool basename:\n%s", output)
			}
			if strings.Contains(text, "POOL_DEFER_RAN") || strings.Contains(text, "POOL_FAILURE_RETURNED") {
				t.Fatalf("fatal failure returned or ran deferred cleanup:\n%s", output)
			}
		})
	}
}

func TestDistributedPoolSynchronousFailureReturnsToCaller(t *testing.T) {
	name := filepath.Join(t.TempDir(), "missing", "0")
	reader := NewStatePoolReader(1, name)
	if _, err := reader.DoWork(make([]*TLCStateMut, 1), name); !isJavaIOException(err) {
		t.Fatalf("synchronous reader error = %v", err)
	}
	if reader.poolFile != name || reader.isFull {
		t.Fatal("failed synchronous read advanced pending work")
	}
	writer := NewStatePoolWriter(1, nil)
	pending := []*TLCStateMut{nil}
	if _, err := writer.DoWork(pending, name); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.DoWork(make([]*TLCStateMut, 1), name+"next"); !isJavaIOException(err) {
		t.Fatalf("synchronous writer error = %v", err)
	}
	if writer.poolFile != name || &writer.buf[0] != &pending[0] {
		t.Fatal("failed synchronous write replaced pending work")
	}
}
