package tlc

import (
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	offHeapDefaultProbeLimit  = 1024
	offHeapProbeLimitProperty = "tlc2.tool.fp.OffHeapDiskFPSet.probeLimit"
	offHeapFound              = -1
)

type offHeapSynchronizer struct {
	mu              sync.Mutex
	cond            *sync.Cond
	sets            *InsMap[*OffHeapDiskFPSet, struct{}]
	flusherChosen   atomic.Bool
	parties         int
	waiting         int
	generation      uint64
	lastEvictionErr error
}

var offHeapGlobalSync = newOffHeapSynchronizer()

func newOffHeapSynchronizer() *offHeapSynchronizer {
	s := &offHeapSynchronizer{
		sets:    NewInsMap[*OffHeapDiskFPSet, struct{}](),
		parties: 1,
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *offHeapSynchronizer) add(set *OffHeapDiskFPSet) {
	if set == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sets.Set(set, struct{}{})
}

func (s *offHeapSynchronizer) remove(set *OffHeapDiskFPSet) {
	if set == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sets.Delkey(set)
}

func (s *offHeapSynchronizer) incWorkers(numWorkers int) {
	if numWorkers <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.parties < numWorkers {
		s.parties = numWorkers
	}
}

func (s *offHeapSynchronizer) evict() {
	s.flusherChosen.CompareAndSwap(false, true)
}

func (s *offHeapSynchronizer) awaitIfPending() error {
	// Java checks its AtomicBoolean before entering the phaser. Ordinary
	// lookups and CAS insertions do not acquire the global barrier mutex.
	if !s.flusherChosen.Load() {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.flusherChosen.Load() {
		return nil
	}
	generation := s.generation
	s.waiting++
	if s.waiting >= s.parties {
		var firstErr error
		for set := range s.sets.All() {
			if err := set.evict(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		s.lastEvictionErr = firstErr
		s.waiting = 0
		s.generation++
		if !s.flusherChosen.CompareAndSwap(true, false) {
			panic(NewTLCRuntimeException(ECGeneral))
		}
		s.cond.Broadcast()
		return firstErr
	}
	for generation == s.generation && s.flusherChosen.Load() {
		s.cond.Wait()
	}
	return s.lastEvictionErr
}

func offHeapDiskFPSetProbeLimit() int {
	if value, ok := tlcLookupSystemProperty(offHeapProbeLimitProperty); ok {
		if parsed, ok := javaIntProperty(value); ok {
			return parsed
		}
	}
	return offHeapDefaultProbeLimit
}

// Source puts and membership checks use atomic array words and CAS, with the
// shared eviction barrier providing exclusive access only during a flush.
func (s *OffHeapDiskFPSet) Put(fp uint64) bool {
	fp0 := fp & diskFPSetFlushedMask
	for {
		if err := offHeapGlobalSync.awaitIfPending(); err != nil {
			panic(err)
		}
		start := 0
		if s.index != nil {
			if found := s.memLookup0(fp0); found == offHeapFound {
				atomic.AddUint64(&s.memHitCnt, 1)
				return true
			} else {
				start = found
			}
			hit, err := s.diskLookup(fp0)
			if err != nil {
				panic(err)
			}
			if hit {
				atomic.AddUint64(&s.diskHitCnt, 1)
				return true
			}
		}
		seen, inserted := s.memInsert0(fp0, start)
		if seen {
			return true
		}
		if inserted {
			return false
		}
		offHeapGlobalSync.evict()
	}
}

func (s *OffHeapDiskFPSet) Contains(fp uint64) bool {
	if err := offHeapGlobalSync.awaitIfPending(); err != nil {
		panic(err)
	}
	fp0 := fp & diskFPSetFlushedMask
	if s.memLookup(fp0) {
		return true
	}
	hit, err := s.diskLookup(fp0)
	if err != nil {
		panic(err)
	}
	if hit {
		atomic.AddUint64(&s.diskHitCnt, 1)
	}
	return hit
}

func (s *OffHeapDiskFPSet) PutBlock(fpv *LongVec) *BitVector {
	if fpv == nil {
		return NewBitVector(0)
	}
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		if !s.Put(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *OffHeapDiskFPSet) ContainsBlock(fpv *LongVec) *BitVector {
	if fpv == nil {
		return NewBitVector(0)
	}
	s.mu.Lock()
	s.statesSeen += uint64(fpv.Size())
	s.mu.Unlock()
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *OffHeapDiskFPSet) RecoverTrace(trace *TLCTrace) error {
	if trace == nil {
		return s.Recover()
	}
	elements := trace.Elements()
	defer elements.Close()
	for pos := elements.NextPos(); pos != -1; pos = elements.NextPos() {
		fp := elements.NextFP()
		if err := s.RecoverFP(fp); err != nil {
			return err
		}
	}
	return nil
}

func (s *OffHeapDiskFPSet) RecoverFP(fp uint64) error {
	fp0 := fp & diskFPSetFlushedMask
	for {
		if err := offHeapGlobalSync.awaitIfPending(); err != nil {
			return err
		}
		s.mu.Lock()
		seen, inserted := s.memInsert0(fp0, 0)
		if seen {
			s.mu.Unlock()
			if diskFPSetError2Warning() {
				PrintWarning(ECSystemCheckpointRecoveryCorrupt, fmt.Sprintf("Encountered duplicate fingerprint value %d", fp0))
				return nil
			}
			return newTLCErrorCode(ECSystemCheckpointRecoveryCorrupt, "")
		}
		if inserted {
			if s.needsDiskFlush() {
				err := s.evictLocked()
				s.mu.Unlock()
				return err
			}
			s.mu.Unlock()
			return nil
		}
		s.mu.Unlock()
		offHeapGlobalSync.evict()
	}
}

func (s *OffHeapDiskFPSet) CheckFPs() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if atomic.LoadInt64(&s.tblCnt) <= 0 {
		return uint64(1<<63 - 1)
	}
	numThreads := NumWorkers()
	if numThreads < 1 {
		numThreads = 1
	}
	if flusher := s.selectOffHeapConcurrentFlusher(numThreads); flusher != nil {
		flusher.prepareTable()
	} else {
		s.prepareOffHeapTableLocked()
	}
	partitionLen := int64(math.Floor(float64(s.array.Size()) / float64(numThreads)))
	distances := make([]int64, numThreads)
	for id := range distances {
		distances[id] = 1<<63 - 1
	}
	// The source creates a separate executor for the closest-pair scan; it does
	// not shut down the selected flusher's executor after preparing the table.
	scan := &offHeapConcurrentFlusher{numThreads: numThreads}
	failures := scan.invokeAll(func(id int) {
		isLast := id == numThreads-1
		start := int64(id) * partitionLen
		end := start + partitionLen
		if isLast {
			end = s.array.Size() - 1
		}
		end++
		defer func() {
			if failure := recover(); failure != nil {
				if _, ok := failure.(*NoSuchElementException); !ok {
					panic(failure)
				}
			}
		}()
		itr := newOffHeapIterator(s.array, atomic.LoadInt64(&s.tblCnt), start, s.indexer, !isLast || id == 0)
		x, ok := itr.next()
		if !ok {
			return
		}
		for {
			y, ok := itr.nextUntil(end)
			if !ok {
				break
			}
			d := y - x
			if (d > 0) == (d < 0 && itr.pos > end && isLast) {
				panic(NewAssertionError())
			}
			if d < distances[id] {
				distances[id] = d
			}
			x = y
		}
	})
	scan.shutdown = true
	distance := int64(1<<63 - 1)
	for id, result := range distances {
		if failures[id] != nil {
			panic(newOffHeapRuntimeException(NewExecutionException(failures[id])))
		}
		if result < distance {
			distance = result
		}
	}
	return uint64(distance)
}

func (s *OffHeapDiskFPSet) Sizeof() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint64(44) + uint64(s.maxTblCnt*fpSetLongSize) + uint64(len(s.index))*4
}

func (s *OffHeapDiskFPSet) GetTblCapacity() int64     { return s.maxTblCnt }
func (s *OffHeapDiskFPSet) GetTblLoad() int64         { return atomic.LoadInt64(&s.tblCnt) }
func (s *OffHeapDiskFPSet) GetOverallCapacity() int64 { return s.array.Size() }
func (s *OffHeapDiskFPSet) GetBucketCapacity() int64  { return int64(s.probeLimit) }

func (s *OffHeapDiskFPSet) ForceFlush() {
	offHeapGlobalSync.evict()
}

func (s *OffHeapDiskFPSet) needsDiskFlush() bool {
	return atomic.LoadInt64(&s.tblCnt) >= s.maxTblCnt || s.forceFlush.Load()
}

func (s *OffHeapDiskFPSet) memLookup(fp0 uint64) bool {
	return s.memLookup0(fp0) == offHeapFound
}

func (s *OffHeapDiskFPSet) memLookup0(fp0 uint64) int {
	free := s.probeLimit
	for i := 0; i <= s.probeLimit; i++ {
		position := s.indexer.GetIdxProbe(fp0, i)
		value := s.array.Get(position)
		if fp0 == (uint64(value) & diskFPSetFlushedMask) {
			return offHeapFound
		}
		if value == 0 {
			if i < free {
				return i
			}
			return free
		}
		if value < 0 && free == s.probeLimit {
			free = i
		}
	}
	return free
}

func (s *OffHeapDiskFPSet) memInsert0(fp0 uint64, start int) (seen bool, inserted bool) {
	for i := start; i < s.probeLimit; i++ {
		position := s.indexer.GetIdxProbe(fp0, i)
		expected := s.array.Get(position)
		if expected == 0 || (expected < 0 && fp0 != (uint64(expected)&diskFPSetFlushedMask)) {
			if s.array.TrySet(position, expected, int64(fp0)) {
				atomic.AddInt64(&s.tblCnt, 1)
				return false, true
			}
			i--
			continue
		}
		if (uint64(expected) & diskFPSetFlushedMask) == fp0 {
			return true, true
		}
	}
	return false, false
}

func (s *OffHeapDiskFPSet) prepareOffHeapTableLocked() {
	if s == nil || s.array == nil || s.array.Size() == 0 {
		return
	}
	if !s.checkOffHeapInput() {
		panic(NewAssertionError())
	}
	LongArraysSortRange(s.array, 0, s.array.Size()-1+int64(s.probeLimit), s.offHeapLongComparator)
	if s.checkOffHeapSorted() != -1 {
		panic(NewAssertionError())
	}
}

func (s *OffHeapDiskFPSet) offHeapLongComparator(fpA int64, posA int64, fpB int64, posB int64) int {
	if fpA <= 0 || fpB <= 0 {
		return 0
	}
	wrappedA := s.indexer.GetIdx(uint64(fpA)) > posA
	wrappedB := s.indexer.GetIdx(uint64(fpB)) > posB
	if wrappedA == wrappedB && posA > posB {
		if fpA < fpB {
			return -1
		}
		return 1
	}
	if wrappedA != wrappedB {
		if posA < posB && fpA < fpB {
			return -1
		}
		if posA > posB && fpA > fpB {
			return -1
		}
	}
	return 0
}

func (s *OffHeapDiskFPSet) IncWorkers(num int) {
	if num != s.numThreads {
		panic(NewAssertionError())
	}
	offHeapGlobalSync.incWorkers(num)
}

func (s *OffHeapDiskFPSet) Close() {
	offHeapGlobalSync.remove(s)
	if s != nil && s.DiskFPSet != nil {
		s.DiskFPSet.Close()
	}
}

func (s *OffHeapDiskFPSet) evict() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.evictLocked()
}

func (s *OffHeapDiskFPSet) evictLocked() error {
	s.growDiskMark++
	start := time.Now()
	defer func() {
		s.flushTime += int64(time.Since(start) / time.Millisecond)
	}()
	if atomic.LoadInt64(&s.tblCnt) == 0 {
		s.forceFlush.Store(false)
		return nil
	}
	if !s.checkOffHeapInput() {
		panic(NewAssertionError())
	}
	flusher := s.selectOffHeapConcurrentFlusher(s.numThreads)
	if flusher != nil {
		flusher.prepareTable()
	} else {
		s.prepareOffHeapTableLocked()
	}
	itr := newOffHeapIterator(s.array, atomic.LoadInt64(&s.tblCnt), 0, s.indexer, true)
	if err := s.mergeOffHeapIteratorWithFlusher(offHeapMergeIterator{atomic.LoadInt64(&s.tblCnt), itr.markNext, itr.hasNext}, flusher); err != nil {
		if isJavaIOException(err) {
			return newOffHeapRuntimeException(err)
		}
		return err
	}
	ok, err := s.checkOffHeapIndex()
	if err != nil {
		return err
	}
	if !ok {
		panic(NewAssertionError())
	}
	atomic.StoreInt64(&s.tblCnt, 0)
	atomic.StoreInt64(&s.tblLoad, 0)
	s.forceFlush.Store(false)
	return nil
}

// Retain the native slice entry point while sharing the source stream merge.
func (s *OffHeapDiskFPSet) mergeOffHeapValues(newValues []uint64) error {
	position := 0
	return s.mergeOffHeapIterator(offHeapMergeIterator{
		elements: int64(len(newValues)),
		markNext: func() (int64, bool) {
			if position >= len(newValues) {
				panic(NewNoSuchElementException())
			}
			v := newValues[position]
			position++
			return int64(v), true
		},
		hasNext: func() bool { return position < len(newValues) },
	})
}

func (s *OffHeapDiskFPSet) mergeOffHeapIterator(itr offHeapMergeIterator) error {
	return s.mergeOffHeapIteratorWithFlusher(itr, nil)
}

func (s *OffHeapDiskFPSet) mergeOffHeapIteratorWithFlusher(itr offHeapMergeIterator, flusher *offHeapConcurrentFlusher) (err error) {
	// DiskFPSet.Flusher.flushTable wraps checked I/O failures from the entire
	// merge/replacement lifecycle, while Assert runtime failures pass through.
	defer func() {
		if isJavaIOException(err) {
			err = NewIOException("Error: merging entries into file " + s.fpFilename + "  " + javaThrowableString(err))
		}
	}()
	newIndex := make([]uint64, s.calculateOffHeapIndexLen(itr.elements))
	for _, reader := range s.braf {
		if err := reader.Seek(0); err != nil {
			return err
		}
	}
	for _, reader := range s.brafPool {
		if err := reader.Close(); err != nil {
			return err
		}
	}
	_ = os.Remove(s.tmpFilename)
	out, err := NewBufferedRandomAccessFile(s.tmpFilename, "rw")
	if err != nil {
		return err
	}
	defer out.Close()
	outLength := (itr.elements + atomic.LoadInt64(&s.fileCnt)) * fpSetLongSize
	if err := out.SetLength(outLength); err != nil {
		return err
	}
	in := s.braf[0]
	if err := in.Seek(0); err != nil {
		return err
	}
	length, err := in.Length()
	if err != nil {
		return err
	}
	if flusher != nil {
		err = flusher.mergeNewEntries(out)
	} else {
		err = s.mergeOffHeapEntries(in, out, itr, length/fpSetLongSize)
	}
	if err != nil {
		return err
	}
	length, err = out.Length()
	if err != nil {
		return err
	}
	if err := s.writeIndex(newIndex, out, length/fpSetLongSize-1); err != nil {
		return err
	}
	s.index = newIndex
	atomic.AddInt64(&s.fileCnt, itr.elements)
	if err := out.Close(); err != nil {
		return err
	}
	readerCnt, poolCnt := len(s.braf), len(s.brafPool)
	if err := s.closeBRAFReaders(); err != nil {
		return err
	}
	if err := replaceFile(s.tmpFilename, s.fpFilename); err != nil {
		return newTLCRuntimeExceptionWithCause(ECSystemUnableNotRenameFile, bufferedRandomAccessFileIOError(err))
	}
	if err := s.openBRAFReaders(readerCnt, poolCnt); err != nil {
		return err
	}
	return nil
}

func (s *OffHeapDiskFPSet) calculateOffHeapIndexLen(buffLen int64) int {
	indexLen := s.calculateIndexLen(buffLen)
	if (buffLen+atomic.LoadInt64(&s.fileCnt)-1)%diskFPSetNumEntriesPerPage == 0 {
		indexLen--
	}
	return indexLen
}

type offHeapIterator struct {
	elements     int64
	array        *LongArray
	indexer      *OffHeapIndexer
	canWrap      bool
	pos          int64
	elementsRead int64
}

func newOffHeapIterator(array *LongArray, elements int64, start int64, indexer *OffHeapIndexer, canWrap bool) *offHeapIterator {
	return &offHeapIterator{array: array, elements: elements, pos: start, indexer: indexer, canWrap: canWrap}
}

func (i *offHeapIterator) next() (int64, bool) {
	return i.next0(false, 1<<63-1)
}

func (i *offHeapIterator) nextUntil(maxPos int64) (int64, bool) {
	if i.pos >= maxPos {
		return 0, false
	}
	return i.next0(false, maxPos)
}

// Original Iterator.markNext sets MARK_FLUSHED on the selected array entry.
func (i *offHeapIterator) markNext() (int64, bool) {
	return i.next0(true, 1<<63-1)
}

func (i *offHeapIterator) next0(mark bool, maxPos int64) (int64, bool) {
	if i == nil || i.array == nil {
		panic(NewNullPointerException())
	}
	if i.array.Size() == 0 {
		panic(NewArithmeticException("/ by zero"))
	}
	// Preserve Java's do/while: the first position is examined even when
	// hasNext is false, and continues evaluate the condition at the loop end.
	for {
		position := i.pos % i.array.Size()
		elem := i.array.Get(position)
		if elem > 0 {
			baseIdx := i.indexer.GetIdx(uint64(elem))
			if baseIdx > i.pos {
				if !i.canWrap {
					panic(NewAssertionError())
				}
			} else {
				i.pos++
				if mark {
					i.array.Set(position, int64(uint64(elem)|diskFPSetMarkFlushed))
				}
				i.elementsRead++
				return elem, true
			}
		}
		i.pos++
		if !i.hasNext() || i.pos >= maxPos {
			break
		}
	}
	if i.pos >= maxPos {
		return 0, false
	}
	panic(NewNoSuchElementException())
}

func (i *offHeapIterator) hasNext() bool {
	if i == nil {
		panic(NewNullPointerException())
	}
	return i.elementsRead < i.elements
}
