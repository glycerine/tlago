package tlc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	flusherChosen   bool
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.flusherChosen {
		s.flusherChosen = true
	}
}

func (s *offHeapSynchronizer) awaitIfPending() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.flusherChosen {
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
		s.flusherChosen = false
		s.cond.Broadcast()
		return firstErr
	}
	for generation == s.generation && s.flusherChosen {
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

func (s *OffHeapDiskFPSet) Put(fp uint64) bool {
	fp0 := fp & diskFPSetFlushedMask
	for {
		if err := offHeapGlobalSync.awaitIfPending(); err != nil {
			panic(err)
		}
		s.mu.Lock()
		start := 0
		if s.index != nil {
			if found := s.memLookup0(fp0); found == offHeapFound {
				atomic.AddUint64(&s.memHitCnt, 1)
				s.mu.Unlock()
				return true
			} else {
				start = found
			}
			hit, err := s.diskLookup(fp0)
			if err != nil {
				s.mu.Unlock()
				panic(err)
			}
			if hit {
				atomic.AddUint64(&s.diskHitCnt, 1)
				s.mu.Unlock()
				return true
			}
		}
		seen, inserted := s.memInsert0(fp0, start)
		if seen {
			s.mu.Unlock()
			return true
		}
		if inserted {
			s.mu.Unlock()
			return false
		}
		s.mu.Unlock()
		offHeapGlobalSync.evict()
	}
}

func (s *OffHeapDiskFPSet) Contains(fp uint64) bool {
	if err := offHeapGlobalSync.awaitIfPending(); err != nil {
		panic(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
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
	if s.tblCnt <= 0 {
		return uint64(1<<63 - 1)
	}
	s.prepareOffHeapTableLocked()
	numThreads := NumWorkers()
	if numThreads < 1 {
		numThreads = 1
	}
	partitionLen := s.array.Size() / int64(numThreads)
	distance := int64(1<<63 - 1)
	for id := 0; id < numThreads; id++ {
		isLast := id == numThreads-1
		start := int64(id) * partitionLen
		end := start + partitionLen
		if isLast {
			end = s.array.Size() - 1
		}
		end++
		canWrap := !isLast || id == 0
		itr := newOffHeapIterator(s.array, s.tblCnt, start, s.indexer, canWrap)
		x, ok := itr.next()
		if !ok {
			continue
		}
		for {
			y, ok := itr.nextUntil(end)
			if !ok {
				break
			}
			d := y - x
			if d < distance {
				distance = d
			}
			x = y
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
func (s *OffHeapDiskFPSet) GetTblLoad() int64         { return s.tblCnt }
func (s *OffHeapDiskFPSet) GetOverallCapacity() int64 { return s.array.Size() }
func (s *OffHeapDiskFPSet) GetBucketCapacity() int64  { return int64(s.probeLimit) }

func (s *OffHeapDiskFPSet) ForceFlush() {
	offHeapGlobalSync.evict()
}

func (s *OffHeapDiskFPSet) needsDiskFlush() bool {
	return s.tblCnt >= s.maxTblCnt || s.forceFlush.Load()
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
				s.tblCnt++
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
	LongArraysSortRange(s.array, 0, s.array.Size()-1+int64(s.probeLimit), s.offHeapLongComparator)
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
	if s.tblCnt == 0 {
		s.forceFlush.Store(false)
		return nil
	}
	values := s.unflushedValuesLocked()
	if len(values) == 0 {
		s.tblCnt = 0
		s.forceFlush.Store(false)
		return nil
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	if err := s.mergeOffHeapValues(values); err != nil {
		return err
	}
	for _, fp := range values {
		s.markFlushed(fp)
	}
	s.tblCnt = 0
	s.tblLoad = 0
	s.forceFlush.Store(false)
	return nil
}

func (s *OffHeapDiskFPSet) unflushedValuesLocked() []uint64 {
	values := make([]uint64, 0, s.tblCnt)
	for pos := int64(0); pos < s.array.Size(); pos++ {
		value := s.array.Get(pos)
		if value > 0 {
			values = append(values, uint64(value))
		}
	}
	return values
}

func (s *OffHeapDiskFPSet) markFlushed(fp uint64) {
	for i := 0; i <= s.probeLimit; i++ {
		position := s.indexer.GetIdxProbe(fp, i)
		value := s.array.Get(position)
		if uint64(value) == fp {
			s.array.Set(position, int64(fp|diskFPSetMarkFlushed))
			return
		}
	}
}

func (s *OffHeapDiskFPSet) mergeOffHeapValues(newValues []uint64) error {
	oldValues, err := readFingerprintFile(s.fpFilename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	indexLen := s.calculateOffHeapIndexLen(int64(len(newValues)))
	newIndex := make([]uint64, indexLen)
	if err := os.MkdirAll(filepath.Dir(s.tmpFilename), 0o755); err != nil {
		return err
	}
	tmp, err := os.Create(s.tmpFilename)
	if err != nil {
		return err
	}
	currIndex := 0
	written := int64(0)
	var last uint64
	writeFP := func(fp uint64) error {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], fp)
		if _, err := tmp.Write(buf[:]); err != nil {
			return err
		}
		atomic.AddUint64(&s.diskWriteCnt, 1)
		if written%diskFPSetNumEntriesPerPage == 0 && currIndex < len(newIndex) {
			newIndex[currIndex] = fp
			currIndex++
		}
		last = fp
		written++
		return nil
	}
	i, j := 0, 0
	for i < len(oldValues) && j < len(newValues) {
		switch {
		case oldValues[i] < newValues[j]:
			if err := writeFP(oldValues[i]); err != nil {
				_ = tmp.Close()
				return err
			}
			i++
		case oldValues[i] > newValues[j]:
			if err := writeFP(newValues[j]); err != nil {
				_ = tmp.Close()
				return err
			}
			j++
		default:
			PrintWarning(ECTLCFPValueAlreadyOnDisk, fmt.Sprint(oldValues[i]))
			if err := writeFP(oldValues[i]); err != nil {
				_ = tmp.Close()
				return err
			}
			i++
			j++
		}
	}
	for ; i < len(oldValues); i++ {
		if err := writeFP(oldValues[i]); err != nil {
			_ = tmp.Close()
			return err
		}
	}
	for ; j < len(newValues); j++ {
		if err := writeFP(newValues[j]); err != nil {
			_ = tmp.Close()
			return err
		}
	}
	if written == 0 {
		last = 0
	}
	if currIndex < len(newIndex) {
		newIndex[currIndex] = last
		currIndex++
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	readerCnt := len(s.braf)
	poolCnt := len(s.brafPool)
	if err := s.closeBRAFReaders(); err != nil {
		return err
	}
	if currIndex != indexLen {
		_ = s.openBRAFReaders(readerCnt, poolCnt)
		return fmt.Errorf("OffHeapDiskFPSet index mismatch: got %d want %d", currIndex, indexLen)
	}
	if err := replaceFile(s.tmpFilename, s.fpFilename); err != nil {
		_ = s.openBRAFReaders(readerCnt, poolCnt)
		return err
	}
	if err := s.openBRAFReaders(readerCnt, poolCnt); err != nil {
		return err
	}
	s.index = newIndex
	s.fileCnt += int64(len(newValues))
	return nil
}

func (s *OffHeapDiskFPSet) calculateOffHeapIndexLen(buffLen int64) int {
	indexLen := s.calculateIndexLen(buffLen)
	if (buffLen+s.fileCnt-1)%diskFPSetNumEntriesPerPage == 0 {
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
	return i.next0(1<<63 - 1)
}

func (i *offHeapIterator) nextUntil(maxPos int64) (int64, bool) {
	if i.pos >= maxPos {
		return 0, false
	}
	return i.next0(maxPos)
}

func (i *offHeapIterator) next0(maxPos int64) (int64, bool) {
	if i == nil || i.array == nil || i.array.Size() == 0 {
		return 0, false
	}
	for i.hasNext() && i.pos < maxPos {
		position := i.pos % i.array.Size()
		elem := i.array.Get(position)
		if elem <= 0 {
			i.pos++
			continue
		}
		baseIdx := i.indexer.GetIdx(uint64(elem))
		if baseIdx > i.pos {
			i.pos++
			continue
		}
		i.pos++
		i.elementsRead++
		return elem, true
	}
	return 0, false
}

func (i *offHeapIterator) hasNext() bool {
	return i != nil && i.elementsRead < i.elements
}
