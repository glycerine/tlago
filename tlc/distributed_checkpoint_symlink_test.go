package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

// The source's exists/delete/rename sequence follows a symlink for existence.
// A dangling link must survive when promotion fails, whereas a live link is
// deleted before promotion. Neither operation may change the link's target.
func TestDistributedCheckpointCommitSymlinkOrdering(t *testing.T) {
	for _, owner := range []string{"trace", "worker", "memory_queue", "disk_queue", "intern", "memory_fp", "memory_fp1", "memory_fp2"} {
		for _, live := range []bool{false, true} {
			for _, promote := range []bool{false, true} {
				phase := "dangling"
				if live {
					phase = "live"
				}
				if promote {
					phase += "_success"
				} else {
					phase += "_missing_temporary"
				}
				t.Run(owner+"/"+phase, func(t *testing.T) {
					directory := t.TempDir()
					var commit func() error
					var oldPath, newPath, message string
					switch owner {
					case "trace":
						trace := NewTLCTrace(directory, "Spec")
						t.Cleanup(func() { _ = trace.Close() })
						commit = trace.CommitChkpt
						oldPath, newPath = trace.chkptName("chkpt"), trace.chkptName("tmp")
						message = "Trace.commitChkpt: cannot delete "
					case "worker":
						worker := NewWorker(0)
						worker.SetTraceContext(directory, "Spec")
						commit = worker.CommitChkpt
						oldPath, newPath = worker.traceFileBase+".chkpt", worker.traceFileBase+".tmp"
						message = "Trace.commitChkpt: cannot delete "
					case "memory_queue":
						commit = NewMemStateQueue(directory).CommitChkpt
						oldPath, newPath = filepath.Join(directory, "queue.chkpt"), filepath.Join(directory, "queue.tmp")
						message = "MemStateQueue.commitChkpt: cannot delete "
					case "disk_queue":
						commit = (&DiskStateQueue{diskdir: directory}).CommitChkpt
						oldPath, newPath = filepath.Join(directory, "queue.chkpt"), filepath.Join(directory, "queue.tmp")
						message = "DiskStateQueue.commitChkpt: cannot delete "
					case "intern":
						table := NewInternTable(16)
						commit = func() error { return table.CommitChkpt(directory) }
						oldPath, newPath = filepath.Join(directory, "vars.chkpt"), filepath.Join(directory, "vars.tmp")
						message = "InternTable.commitChkpt: cannot delete "
					case "memory_fp", "memory_fp1", "memory_fp2":
						var set FPSet
						switch owner {
						case "memory_fp":
							set = &MemFPSet{}
							message = "MemFPSet.commitChkpt: cannot delete "
						case "memory_fp1":
							set = &MemFPSet1{}
							message = "MemFPSet1.commitChkpt: cannot delete "
						case "memory_fp2":
							set = &MemFPSet2{}
							message = "MemFPSet2.commitChkpt: cannot delete "
						}
						set.Init(1, directory, "Spec")
						commit = func() error { return set.CommitChkptFile("Spec") }
						oldPath, newPath = filepath.Join(directory, "Spec.fp.chkpt"), filepath.Join(directory, "Spec.fp.tmp")
					}
					target := filepath.Join(directory, "target")
					if live {
						if err := os.WriteFile(target, []byte("target retained"), 0600); err != nil {
							t.Fatal(err)
						}
					}
					if err := os.Symlink(target, oldPath); err != nil {
						t.Fatal(err)
					}
					if promote {
						if err := os.WriteFile(newPath, []byte("new checkpoint"), 0600); err != nil {
							t.Fatal(err)
						}
					}
					err := commit()
					if promote {
						if err != nil {
							t.Fatal(err)
						}
						data, err := os.ReadFile(oldPath)
						info, statErr := os.Lstat(oldPath)
						if err != nil || string(data) != "new checkpoint" || statErr != nil || !info.Mode().IsRegular() {
							t.Fatalf("checkpoint not promoted to regular file: %q/%v/%v", data, err, statErr)
						}
						if _, err := os.Lstat(newPath); !os.IsNotExist(err) {
							t.Fatalf("promoted temporary remains: %v", err)
						}
					} else {
						if !isJavaIOException(err) || err.Error() != message+oldPath {
							t.Fatalf("commit failure = %T/%v", err, err)
						}
						if live {
							if _, err := os.Lstat(oldPath); !os.IsNotExist(err) {
								t.Fatalf("live old link not deleted before failed promotion: %v", err)
							}
						} else if link, err := os.Readlink(oldPath); err != nil || link != target {
							t.Fatalf("dangling old link lost before failed promotion: %q/%v", link, err)
						}
					}
					if live {
						if data, err := os.ReadFile(target); err != nil || string(data) != "target retained" {
							t.Fatalf("commit changed symlink target: %q/%v", data, err)
						}
					} else if _, err := os.Lstat(target); !os.IsNotExist(err) {
						t.Fatalf("commit created dangling link target: %v", err)
					}
				})
			}
		}
	}
}
