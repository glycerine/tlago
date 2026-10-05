// Copyright (c) 2016 Microsoft Research. All rights reserved.
package tlc

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

type OffHeapRuntimeException struct{ *RuntimeException }

func newOffHeapRuntimeException(cause error) *OffHeapRuntimeException {
	return &OffHeapRuntimeException{NewRuntimeExceptionFromCause(cause)}
}

type offHeapFlusherBarrier struct {
	mu               sync.Mutex
	cond             *sync.Cond
	parties, waiting int
	generation       uint64
}

func newOffHeapFlusherBarrier(parties int) *offHeapFlusherBarrier {
	b := &offHeapFlusherBarrier{parties: parties}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *offHeapFlusherBarrier) await() {
	b.mu.Lock()
	defer b.mu.Unlock()
	generation := b.generation
	b.waiting++
	if b.waiting == b.parties {
		b.waiting = 0
		b.generation++
		b.cond.Broadcast()
		return
	}
	for b.generation == generation {
		b.cond.Wait()
	}
}

type offHeapFlusherResult struct {
	table, disk, outOffset, inOffset int64
}

type offHeapConcurrentFlusher struct {
	set                   *OffHeapDiskFPSet
	numThreads            int
	r, insertions, length int64
	offsets               []offHeapFlusherResult
	shutdown              bool
}

func (s *OffHeapDiskFPSet) selectOffHeapConcurrentFlusher(numThreads int) *offHeapConcurrentFlusher {
	length := math.Floor(float64(s.array.Size()) / float64(numThreads))
	if s.array.Size() >= 8192 && length > float64(2*s.probeLimit) {
		if numThreads <= 0 {
			panic(NewIllegalArgumentException())
		}
		s.concurrentFlusher = &offHeapConcurrentFlusher{set: s, numThreads: numThreads,
			r: int64(s.probeLimit), insertions: atomic.LoadInt64(&s.tblCnt), length: javaDoubleToLong(length)}
	}
	return s.concurrentFlusher
}

// Callable failures become Future failures. prepareTable inspects every future;
// mergeNewEntries deliberately leaves those futures uninspected, as Java does.
func (f *offHeapConcurrentFlusher) invokeAll(task func(int)) []error {
	if f.shutdown {
		panic(NewRejectedExecutionException(nil, nil))
	}
	failures := make([]error, f.numThreads)
	var finished sync.WaitGroup
	for id := 0; id < f.numThreads; id++ {
		finished.Add(1)
		go func(id int) {
			defer finished.Done()
			defer func() {
				if failure := recover(); failure != nil {
					if err, ok := failure.(error); ok {
						failures[id] = err
					} else {
						failures[id] = NewRuntimeException(fmt.Sprint(failure))
					}
				}
			}()
			task(id)
		}(id)
	}
	finished.Wait()
	return failures
}

func (s *OffHeapDiskFPSet) offHeapTableOffset(start, limit int64) int64 {
	occupied := int64(0)
	for pos := start; pos < limit; pos++ {
		fp := s.array.Get(pos % s.array.Size())
		if fp <= 0 {
			continue
		}
		idx := s.indexer.GetIdx(uint64(fp))
		if idx > pos || idx+int64(s.probeLimit) < pos {
			continue
		}
		occupied++
	}
	return occupied
}

func (s *OffHeapDiskFPSet) offHeapNextLower(idx int64) uint64 {
	fp := s.array.Get(idx)
	for fp <= 0 || s.indexer.GetIdx(uint64(fp)) > idx {
		idx--
		fp = s.array.Get(idx)
	}
	return uint64(fp)
}

func (f *offHeapConcurrentFlusher) prepareTable() {
	s := f.set
	phase := newOffHeapFlusherBarrier(f.numThreads)
	f.offsets = make([]offHeapFlusherResult, f.numThreads)
	failures := f.invokeAll(func(id int) {
		isFirst, isLast := id == 0, id == f.numThreads-1
		start := int64(id) * f.length
		end := start + f.length
		if isLast {
			end = s.array.Size() - 1
		}
		left := start + 1
		if isFirst {
			left = 0
		}
		LongArraysSortRange(s.array, left, end, s.offHeapLongComparator)
		if s.checkOffHeapSortedRange(left, end) != -1 {
			panic(NewAssertionError())
		}
		phase.await()
		LongArraysSortRange(s.array, end-f.r+1, end+f.r+1, s.offHeapLongComparator)
		phase.await()
		limit := end
		if isLast {
			limit = s.array.Size() + f.r
		}
		occupied := s.offHeapTableOffset(start, limit)
		if occupied > limit-start {
			panic(NewAssertionError())
		}
		result := offHeapFlusherResult{table: occupied}
		if s.index != nil {
			offset := func(pos int64) int64 {
				value, err := s.getDiskOffset(id, s.offHeapNextLower(pos))
				if err != nil {
					panic(err)
				}
				return value
			}
			switch {
			case isFirst && isLast:
				result.disk = atomic.LoadInt64(&s.fileCnt)
			case isFirst:
				result.disk = offset(end)
			case isLast:
				result.disk = atomic.LoadInt64(&s.fileCnt) - offset(start)
			default:
				result.disk = offset(end) - offset(start)
			}
		}
		f.offsets[id] = result
	})
	for _, failure := range failures {
		if failure != nil {
			panic(newOffHeapRuntimeException(NewExecutionException(failure)))
		}
	}
	if s.checkOffHeapSorted() != -1 {
		panic(NewAssertionError())
	}
}

func offHeapCheckRAFs(files []*BufferedRandomAccessFile) bool {
	for id := 0; id < len(files)-1; id++ {
		end, err := files[id].GetFilePointer()
		if err != nil {
			panic(err)
		}
		if end != files[id+1].GetMark() {
			return false
		}
	}
	return true
}

func (f *offHeapConcurrentFlusher) mergeNewEntries(out *BufferedRandomAccessFile) error {
	s := f.set
	var table, disk int64
	for _, result := range f.offsets {
		table += result.table
		disk += result.disk
	}
	if table != f.insertions || disk != atomic.LoadInt64(&s.fileCnt) {
		panic(NewAssertionError())
	}
	for id := 1; id < f.numThreads; id++ {
		prev := f.offsets[id-1]
		f.offsets[id].inOffset = prev.inOffset + prev.disk
		f.offsets[id].outOffset = prev.outOffset + prev.disk + prev.table
	}
	outLength, err := out.Length()
	if err != nil {
		return err
	}
	writers := make([]*BufferedRandomAccessFile, f.numThreads)
	iterators := make([]*offHeapIterator, f.numThreads)
	reads := make([]int64, f.numThreads)
	for id := 0; id < f.numThreads; id++ {
		writers[id], err = NewBufferedRandomAccessFile(s.tmpFilename, "rw")
		if err != nil {
			return err
		}
		if err = writers[id].SetLength(outLength); err != nil {
			return err
		}
		result := f.offsets[id]
		if err = writers[id].SeekAndMark(result.outOffset * fpSetLongSize); err != nil {
			return err
		}
		iterators[id] = newOffHeapIterator(s.array, result.table, int64(id)*f.length, s.indexer, id == 0)
		length, e := s.braf[id].Length()
		if e != nil {
			return e
		}
		if (result.inOffset+result.disk)*fpSetLongSize > length {
			panic(NewAssertionError())
		}
		if err = s.braf[id].SeekAndMark(result.inOffset * fpSetLongSize); err != nil {
			return err
		}
		reads[id] = result.disk
		if id == f.numThreads-1 {
			reads[id] = atomic.LoadInt64(&s.fileCnt) - result.inOffset
		}
	}
	f.invokeAll(func(id int) {
		result := f.offsets[id]
		itr := iterators[id]
		if err := s.mergeOffHeapEntries(s.braf[id], writers[id], offHeapMergeIterator{result.table, itr.markNext, itr.hasNext}, reads[id]); err != nil {
			panic(err)
		}
		position, err := writers[id].GetFilePointer()
		if err != nil {
			panic(err)
		}
		if position != (result.outOffset+result.table+result.disk)*fpSetLongSize {
			panic(NewAssertionError())
		}
	})
	f.shutdown = true
	if !offHeapCheckRAFs(writers) {
		panic(NewAssertionError())
	}
	for _, writer := range writers {
		if err = writer.Close(); err != nil {
			return err
		}
	}
	if err = out.InvalidateBufferedData(); err != nil {
		return err
	}
	if !offHeapCheckRAFs(s.braf) {
		panic(NewAssertionError())
	}
	for pos := int64(0); pos < s.array.Size(); pos++ {
		if s.array.Get(pos) > 0 {
			panic(NewAssertionError())
		}
	}
	return nil
}
