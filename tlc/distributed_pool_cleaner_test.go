package tlc

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// No enabled original method covers failed obsolete-pool deletion diagnostics.
func TestDistributedPoolCleanerReportsFailedDeletesAndContinues(t *testing.T) {
	for _, backend := range []string{"states", "bytes"} {
		for _, linked := range []bool{false, true} {
			t.Run(backend+"/"+map[bool]string{false: "direct", true: "symlink"}[linked], func(t *testing.T) {
				base := t.TempDir()
				target := filepath.Join(base, "target")
				if err := os.MkdirAll(filepath.Join(target, "child"), 0700); err != nil {
					t.Fatal(err)
				}
				directory := target
				if linked {
					link := filepath.Join(base, "link")
					if err := os.Symlink(filepath.Join(target, "child"), link); err != nil {
						t.Fatal(err)
					}
					directory = link + string(os.PathSeparator) + ".."
				}
				if err := os.Mkdir(filepath.Join(target, "1"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, "1", "retained"), []byte("blocked"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, "2"), []byte("obsolete"), 0600); err != nil {
					t.Fatal(err)
				}
				var cleaner interface {
					Start()
					DeleteUpTo(int)
					FinishAndWait()
				}
				var committed func() bool
				if backend == "states" {
					queue := &DiskStateQueue{diskdir: directory}
					cleaner = NewStatePoolCleaner(queue)
					committed = func() bool {
						queue.mu.Lock()
						defer queue.mu.Unlock()
						return queue.lastLoPool == 3
					}
				} else {
					queue := &DiskByteArrayQueue{diskdir: directory}
					cleaner = NewByteArrayPoolCleaner(queue)
					committed = func() bool {
						queue.mu.Lock()
						defer queue.mu.Unlock()
						return queue.lastLoPool == 3
					}
				}
				messages := make(chan Message, 4)
				recorder := RecorderFunc(func(message Message) {
					if message.Code == ECSystemErrorCleaningPool {
						messages <- message
					}
				})
				AddMessageRecorder(recorder)
				defer RemoveMessageRecorder(recorder)
				cleaner.Start()
				defer cleaner.FinishAndWait()
				cleaner.DeleteUpTo(3)
				deadline := time.NewTimer(time.Second)
				defer deadline.Stop()
				for _, pool := range []string{"0", "1"} {
					select {
					case message := <-messages:
						if message.Severity != SeverityWarning || len(message.Params) != 1 || message.Params[0] != filepath.Join(target, pool) {
							t.Fatalf("failed deletion warning lost canonical path/order: %+v", message)
						}
					case <-deadline.C:
						t.Fatal("cleaner did not warn for every failed deletion")
					}
				}
				// Observe the cleaner's committed range under its native owner lock.
				for {
					if committed() {
						break
					}
					select {
					case <-deadline.C:
						t.Fatal("failed deletions stopped cleanup progress")
					case <-time.After(time.Millisecond):
					}
				}
				if _, err := os.Stat(filepath.Join(target, "2")); !os.IsNotExist(err) {
					t.Fatal("failed deletions prevented a later successful deletion")
				}
				if data, err := os.ReadFile(filepath.Join(target, "1", "retained")); err != nil || string(data) != "blocked" {
					t.Fatal("failed deletion changed retained contents")
				}
			})
		}
	}
}
