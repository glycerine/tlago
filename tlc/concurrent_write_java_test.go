// Copyright (c) 2016 Markus Alexander Kuppe. All rights reserved.
package tlc

import (
	"sync"
	"testing"
)

// Original ConcurrentWriteTest methods test and test3. Every source loop,
// partition boundary and read assertion is retained. The remaining methods
// stay in the inventory for reconciliation of the original Java failures/offsets.
func TestJavaConcurrentWrite(t *testing.T) {
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

	t.Run("test", func(t *testing.T) {
		path := javaBRAFTemp(t, "ConcurrentWriteTest_test")
		writer := javaBRAFOpen(path, "rw")
		writer.setLength(4000 * 8)
		for i := int64(0); i < 1000; i++ {
			writer.writeLong(i)
		}
		for i := int64(1000); i < 2000; i++ {
			writer.writeLong(i)
		}
		for i := int64(2000); i < 3000; i++ {
			writer.writeLong(i)
		}
		for i := int64(3000); i < 4000; i++ {
			writer.writeLong(i)
		}
		writer.close()
		check(t, path, 4000)
	})

	t.Run("test3", func(t *testing.T) {
		path := javaBRAFTemp(t, "ConcurrentWriteTest_test3")
		const limit int64 = 400000000
		const writers int64 = 8
		const partition = limit / writers
		var completed sync.WaitGroup
		// invokeAll waits for every Callable, returning futures which the source
		// never inspects. Capture thrown exceptions at that same task boundary.
		var failures [writers]any
		t.Logf("writing all %d values through %d independent buffered handles", limit, writers)
		for id := int64(0); id < writers; id++ {
			completed.Add(1)
			go func(id int64) {
				defer completed.Done()
				defer func() { failures[id] = recover() }()
				writer := javaBRAFOpen(path, "rw")
				writer.setLength(limit * 8)
				writer.seek(id * partition * 8)
				for j := id * partition; j < (id+1)*partition; j++ {
					writer.writeLong(j)
				}
				writer.close()
			}(id)
		}
		completed.Wait()
		t.Logf("checking all %d original read assertions", limit)
		check(t, path, limit)
	})
}
