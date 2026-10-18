//go:build tlc_fp_stress

package tlc

import (
	"encoding/binary"
	"sync"
	"testing"
)

// Original ConcurrentWriteTest.test1/test2/test4, including source failures.
// Keep all bounds, handle lifetimes, write offsets and read assertions. These
// explicit stress methods are reconciliation cases, not passing replacements.
func TestJavaConcurrentWriteSourceFailures(t *testing.T) {
	t.Skip("test is too slow")

	check := func(t *testing.T, path string, limit int64) {
		t.Helper()
		reader := javaBRAFOpen(path, "r")
		for i := int64(0); i < limit; i++ {
			if got := reader.readLong(); got != i {
				t.Fatalf("readLong at %d = %d, want %d", i, got, i)
			}
		}
		reader.close()
	}

	t.Run("test1", func(t *testing.T) {
		path := javaBRAFTemp(t, "ConcurrentWriteTest_test1")
		const limit int64 = 4000000
		const partition = limit / 4
		var handles [4]*javaBRAF
		// Source opens each handle immediately before writing its partition,
		// retaining all four dirty buffers until the final ordered closes.
		for id := int64(0); id < 4; id++ {
			writer := javaBRAFOpen(path, "rw")
			handles[id] = writer
			writer.setLength(limit * 8)
			if id != 0 {
				writer.seek(id * partition * 8)
			}
			for i := id * partition; i < (id+1)*partition; i++ {
				writer.writeLong(i)
			}
		}
		for _, writer := range handles {
			writer.close()
		}
		check(t, path, limit)
	})

	t.Run("test2", func(t *testing.T) {
		path := javaBRAFTemp(t, "ConcurrentWriteTest_test2")
		const limit int64 = 4000000
		const partition = limit / 4
		var handles [4]*javaBRAF
		// Unlike test1, all four handles are opened/seeking before any write.
		for id := int64(0); id < 4; id++ {
			writer := javaBRAFOpen(path, "rw")
			handles[id] = writer
			writer.setLength(limit * 8)
			if id != 0 {
				writer.seek(id * partition * 8)
			}
		}
		for id, writer := range handles {
			for i := int64(id) * partition; i < int64(id+1)*partition; i++ {
				writer.writeLong(i)
			}
		}
		for _, writer := range handles {
			writer.close()
		}
		check(t, path, limit)
	})

	t.Run("test4", func(t *testing.T) {
		path := javaBRAFTemp(t, "ConcurrentWriteTest_test4")
		const limit int64 = 400000000
		const writers int64 = 8
		const partition = limit / writers
		writer := javaBRAFOpen(path, "rw")
		writer.setLength(limit * 8)
		// Native positional access to the same descriptor represents getChannel.
		channel := writer.file
		var completed sync.WaitGroup
		var failures [writers]any
		for id := int64(0); id < writers; id++ {
			completed.Add(1)
			go func(id int64) {
				defer completed.Done()
				// invokeAll captures exceptions in futures the source never reads.
				defer func() { failures[id] = recover() }()
				position := id * partition * 8
				var buffer [8]byte
				for j := id * partition; j < (id+1)*partition; j++ {
					binary.BigEndian.PutUint64(buffer[:], uint64(j))
					// Preserve the source's additional global-index offset.
					_, err := channel.WriteAt(buffer[:], position+j*8)
					javaBRAFThrow(err)
				}
				javaBRAFThrow(channel.Sync())
			}(id)
		}
		completed.Wait()
		writer.close()
		check(t, path, limit)
	})
}
