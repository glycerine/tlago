package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPoolReadersRejectFullBufferWithPendingFileBeforeMutation(t *testing.T) {
	for _, kind := range []string{"state", "byte"} {
		for _, method := range []string{"do_work", "get_cache"} {
			t.Run(kind+"/"+method, func(t *testing.T) {
				pending := filepath.Join(t.TempDir(), "pending")
				var call func()
				var unchanged func() bool
				if kind == "state" {
					cached := []*TLCStateMut{{UID: 1}}
					reader := NewStatePoolReader(1, pending)
					reader.buf, reader.isFull, reader.canRead = cached, true, true
					if method == "do_work" {
						call = func() { _, _ = reader.DoWork([]*TLCStateMut{{UID: 2}}, "next") }
					} else {
						call = func() { _, _ = reader.GetCache([]*TLCStateMut{{UID: 2}}, "next") }
					}
					unchanged = func() bool {
						return reader.poolFile == pending && reader.isFull && reader.canRead && reader.buf[0] == cached[0]
					}
				} else {
					cached := [][]byte{{1}}
					reader := NewByteArrayPoolReader(1, pending)
					reader.buf, reader.isFull, reader.canRead = cached, true, true
					if method == "do_work" {
						call = func() { _, _ = reader.DoWork([][]byte{{2}}, "next") }
					} else {
						call = func() { _, _ = reader.GetCache([][]byte{{2}}, "next") }
					}
					unchanged = func() bool {
						return reader.poolFile == pending && reader.isFull && reader.canRead && reader.buf[0][0] == 1
					}
				}

				defer func() {
					failure := recover()
					assertion, ok := failure.(*AssertionError)
					if !ok || assertion.GetMessage() == nil || *assertion.GetMessage() != "2137" {
						t.Fatalf("failure = %T: %v, want AssertionError(%d)", failure, failure, ECSystemFileNull)
					}
					if !unchanged() {
						t.Fatal("assertion mutated pool reader state")
					}
				}()
				call()
			})
		}
	}
}

// No enabled original method checks partial destination publication in the
// source StatePoolReader.doWork/getCache synchronous read branches.
func TestDistributedPoolReadRetainsPartialDestination(t *testing.T) {
	for _, branch := range []string{"pending", "direct", "cache"} {
		for _, field := range []string{"worker", "uid", "level", "complete"} {
			t.Run(branch+"/"+field, func(t *testing.T) {
				oldVariables, oldVarCount, oldEmpty := stateVariables, UniqueStringVariableCount(), EmptyState
				stateVariables = nil
				EmptyState = newEmptyStateOfKind(false)
				SetUniqueStringVariableCount(0)
				defer func() {
					stateVariables = oldVariables
					SetUniqueStringVariableCount(oldVarCount)
					EmptyState = oldEmpty
				}()
				name := filepath.Join(t.TempDir(), "0")
				file, err := os.Create(name)
				if err != nil {
					t.Fatal(err)
				}
				out := NewValueOutputStreamWithGlobalCompression(file)
				if err := (&TLCStateMut{WorkerID: 5, UID: 42, level: 3}).Write(out); err != nil {
					t.Fatal(err)
				}
				if field == "complete" {
					for _, state := range []*TLCStateMut{{WorkerID: 7, UID: 91, level: 4}, {WorkerID: 9, UID: 109, level: 5}} {
						if err := state.Write(out); err != nil {
							t.Fatal(err)
						}
					}
				} else {
					if field != "worker" {
						if err := out.WriteShortNat(7); err != nil {
							t.Fatal(err)
						}
					}
					if field == "level" {
						if err := out.WriteLongNat(91); err != nil {
							t.Fatal(err)
						}
					}
				}
				if err := out.Close(); err != nil {
					t.Fatal(err)
				}
				original := []*TLCStateMut{{UID: 98}, {UID: 97}, {UID: 96}}
				destination := append([]*TLCStateMut(nil), original...)
				reader := NewStatePoolReader(3, name)
				if branch == "direct" {
					reader.poolFile = ""
				} else if branch == "cache" {
					reader.canRead = true
				}
				priorFile, priorRead := reader.poolFile, reader.canRead
				var result []*TLCStateMut
				if branch == "cache" {
					result, err = reader.GetCache(destination, name+"next")
				} else if branch == "direct" {
					result, err = reader.DoWork(destination, name)
				} else {
					result, err = reader.DoWork(destination, name+"next")
				}
				if field == "complete" {
					if err != nil || len(result) != 3 || &result[0] != &destination[0] || destination[2].UID != 109 {
						t.Fatalf("complete pool read = %v", err)
					}
				} else {
					if !isJavaIOException(err) || result != nil {
						t.Fatalf("partial pool read = %v/%v", result, err)
					}
					if reader.poolFile != priorFile || reader.canRead != priorRead || reader.isFull {
						t.Fatal("failed read advanced reader work")
					}
					if destination[2] != original[2] {
						t.Fatal("failed read changed an unvisited slot")
					}
				}
				first, second := destination[0], destination[1]
				if first == original[0] || first.WorkerID != 5 || first.UID != 42 || first.level != 3 {
					t.Fatal("completed prefix state was lost")
				}
				worker, uid, level := int16(TLCStateInitWorkerID), int64(TLCStateInitUID), TLCStateInitLevel
				if field != "worker" {
					worker = 7
				}
				if field == "level" || field == "complete" {
					uid = 91
				}
				if field == "complete" {
					level = 4
				}
				if second == original[1] || second.WorkerID != worker || second.UID != uid || second.level != level {
					t.Fatalf("failed slot lost source field publication: old=%t worker=%d uid=%d level=%d", second == original[1], second.WorkerID, second.UID, second.level)
				}
			})
		}
	}
}
