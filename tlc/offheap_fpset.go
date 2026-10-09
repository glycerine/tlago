package tlc

import (
	"fmt"
	"math"
	"os"
	"reflect"
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
	mu             sync.Mutex
	cond           *sync.Cond
	sets           *InsMap[*OffHeapDiskFPSet, struct{}]
	flusherChosen  atomic.Bool
	parties        int
	waiting        int
	generation     uint64
	phaserIdentity uint32
}

var offHeapGlobalSync = newOffHeapSynchronizer()

// InitializeOffHeapDiskFPSetStatics represents a fresh Java classloader's
// singleton eviction barrier. Call only when the previous runtime's workers
// have joined; ordinary FPSet construction and close retain source lifetime.
func InitializeOffHeapDiskFPSetStatics() {
	offHeapGlobalSync = newOffHeapSynchronizer()
}

func newOffHeapSynchronizer() *offHeapSynchronizer {
	s := &offHeapSynchronizer{
		sets:    NewInsMap[*OffHeapDiskFPSet, struct{}](),
		parties: 1,
	}
	// Source Object.toString uses a runtime-specific 32-bit identity hash.
	s.phaserIdentity = uint32(reflect.ValueOf(s).Pointer())
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
		if numWorkers > 65535 {
			panic(NewIllegalStateException("Attempt to register more than 65535 parties for " + s.phaserStateString()))
		}
		// bulkRegister waits for onAdvance when the current phase has no
		// unarrived parties. A failed callback does not complete that phase.
		for s.waiting == s.parties {
			s.cond.Wait()
		}
		s.parties = numWorkers
	}
}

func (s *offHeapSynchronizer) phaserStateString() string {
	return fmt.Sprintf("tlc2.tool.fp.OffHeapDiskFPSet$OffHeapSynchronizer$1@%x[phase = %d parties = %d arrived = %d]", s.phaserIdentity, s.generation, s.parties, s.waiting)
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
	// Phaser has already recorded the last arrival before onAdvance runs.
	// If that callback throws, the phase remains with zero unarrived parties;
	// a further arrival is illegal, rather than retrying or releasing waiters.
	if s.waiting == s.parties {
		return NewIllegalStateException("Attempted arrival of unregistered party for " + s.phaserStateString())
	}
	generation := s.generation
	s.waiting++
	if s.waiting == s.parties {
		for set := range s.sets.All() {
			if err := set.evict(); err != nil {
				return err
			}
		}
		if !s.flusherChosen.CompareAndSwap(true, false) {
			panic(NewTLCRuntimeException(ECGeneral))
		}
		s.waiting = 0
		s.generation = (s.generation + 1) & math.MaxInt32
		s.cond.Broadcast()
		return nil
	}
	for generation == s.generation {
		s.cond.Wait()
	}
	return nil
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
		panic(NewNullPointerException())
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
		panic(NewNullPointerException())
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
	elements, err := trace.Elements()
	if err != nil {
		return err
	}
	defer elements.Close() // Release native ownership on failure without replacing the cause.
	for {
		pos, err := elements.NextPos()
		if err != nil {
			return err
		}
		if pos == -1 {
			break
		}
		fp, err := elements.NextFP()
		if err != nil {
			return err
		}
		if err := s.RecoverFP(fp); err != nil {
			return err
		}
	}
	return elements.Close()
}

// DiskFPSet.recoverFP has exclusive access during recovery. OffHeap's
// memInsert0 falls back to ordinary put only when all probes are exhausted;
// a full table otherwise flushes the currently selected flusher directly.
func (s *OffHeapDiskFPSet) RecoverFP(fp uint64) (err error) {
	// Put exposes Java checked I/O failures through the native panic boundary.
	// Restore the checked return for this source throws-IOException method.
	defer func() {
		if failure := recover(); failure != nil {
			if ioErr, ok := failure.(error); ok && isJavaIOException(ioErr) {
				err = ioErr
				return
			}
			panic(failure)
		}
	}()
	fp0 := fp & diskFPSetFlushedMask
	seen, inserted := s.memInsert0(fp0, 0)
	if !inserted {
		s.ForceFlush()
		seen = s.Put(fp0)
	}
	if seen {
		if !diskFPSetError2Warning() {
			return NewTLCRuntimeException(ECSystemCheckpointRecoveryCorrupt, "")
		}
		PrintWarning(ECSystemCheckpointRecoveryCorrupt, fmt.Sprintf("Encountered duplicate fingerprint value %d", fp0))
	}
	if s.needsDiskFlush() {
		return s.flushOffHeapTable()
	}
	return nil
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
		panic(NewAssertionError("Table violates invariants prior to eviction"))
	}
	LongArraysSortRange(s.array, 0, s.array.Size()-1+int64(s.probeLimit), s.offHeapLongComparator)
	if s.checkOffHeapSorted() != -1 {
		panic(NewAssertionError(fmt.Sprintf("Array %s not fully sorted at index %d and reprobe %d.", s.array.String(), s.checkOffHeapSorted(), s.probeLimit)))
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
	if !s.checkOffHeapInput() {
		panic(NewAssertionError("Table violates invariants prior to eviction: " + s.array.String()))
	}
	s.selectOffHeapConcurrentFlusher(s.numThreads)
	if err := s.flushOffHeapTable(); err != nil {
		if isJavaIOException(err) {
			return newOffHeapRuntimeException(err)
		}
		return err
	}
	s.flushTime += int64(time.Since(start) / time.Millisecond)
	return nil
}

// DiskFPSet.Flusher.flushTable dispatches through the existing flusher. Recovery
// and public invariant checks do not count a normal eviction.
func (s *OffHeapDiskFPSet) flushOffHeapTable() error {
	if atomic.LoadInt64(&s.tblCnt) == 0 {
		return nil
	}
	if s.concurrentFlusher != nil && s.concurrentFlusher.shutdown && s.concurrentFlusher.flushCompleted {
		// Authorized correction of Java's closed-flusher lifecycle bug also
		// covers direct invariant/recovery flushing, which bypasses selection.
		// Use the sequential path rather than submitting to a closed executor
		// with partition counts captured before the previous merge. A failed
		// flush retains its source failure state instead of enabling a retry.
		// See ../JAVA_BUG_FOUND.md.
		s.concurrentFlusher = nil
	}
	if s.concurrentFlusher != nil {
		s.concurrentFlusher.prepareTable()
	} else {
		s.prepareOffHeapTableLocked()
	}
	itr := newOffHeapIterator(s.array, atomic.LoadInt64(&s.tblCnt), 0, s.indexer, true)
	if err := s.mergeOffHeapIteratorWithFlusher(offHeapMergeIterator{atomic.LoadInt64(&s.tblCnt), itr.markNext, itr.hasNext}, s.concurrentFlusher); err != nil {
		return err
	}
	atomic.StoreInt64(&s.tblCnt, 0)
	s.bucketsCap = 0
	atomic.StoreInt64(&s.tblLoad, 0)
	if s.concurrentFlusher != nil {
		s.concurrentFlusher.flushCompleted = true
	}
	return nil
}

func (s *OffHeapDiskFPSet) mergeOffHeapIteratorWithFlusher(itr offHeapMergeIterator, flusher *offHeapConcurrentFlusher) (err error) {
	// DiskFPSet.Flusher.flushTable wraps checked I/O failures from the entire
	// merge/replacement lifecycle, while Assert runtime failures pass through.
	defer func() {
		if isJavaIOException(err) {
			err = NewIOException("Error: merging entries into file " + s.fpFilename + "  " + javaThrowableString(err))
		}
	}()
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
	newIndex := make([]uint64, s.calculateOffHeapIndexLen(itr.elements))
	s.index = newIndex
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
	if !checkOffHeapIndexOrder(newIndex) {
		panic(NewAssertionError("Broken disk index."))
	}
	ok, err := checkOffHeapIndexFile(newIndex, out, length/fpSetLongSize-1)
	if err != nil {
		return err
	}
	if !ok {
		panic(NewAssertionError("Misaligned disk index."))
	}
	atomic.AddInt64(&s.fileCnt, itr.elements)
	readerCnt, poolCnt := len(s.braf), len(s.brafPool)
	// Source closes dedicated readers in order; pooled readers were closed
	// before the merge. Propagate the first close failure before closing out.
	for _, reader := range s.braf {
		if err := reader.Close(); err != nil {
			return err
		}
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := replaceFile(s.tmpFilename, s.fpFilename); err != nil {
		return newTLCRuntimeExceptionWithCause(ECSystemUnableNotRenameFile, bufferedRandomAccessFileIOError(err))
	}
	if err := s.openBRAFReaders(readerCnt, poolCnt); err != nil {
		return err
	}
	ok, err = s.checkFlushedFile()
	if err != nil {
		return err
	}
	if !ok {
		panic(NewAssertionError())
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
