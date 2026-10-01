package tlc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	offHeapDefaultProbeLimit  = 1024
	offHeapProbeLimitProperty = "tlc2.tool.fp.OffHeapDiskFPSet.probeLimit"
	offHeapFound              = -1
)

func offHeapDiskFPSetProbeLimit() int {
	if value, ok := tlcLookupSystemProperty(offHeapProbeLimitProperty); ok {
		if parsed, ok := javaIntProperty(value); ok {
			return parsed
		}
	}
	return offHeapDefaultProbeLimit
}

func (s *OffHeapDiskFPSet) Put(fp uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	fp0 := fp & diskFPSetFlushedMask
	for {
		start := 0
		if s.index != nil {
			if found := s.memLookup0(fp0); found == offHeapFound {
				s.memHitCnt++
				return true
			} else {
				start = found
			}
			hit, err := s.diskLookup(fp0)
			if err != nil {
				panic(err)
			}
			if hit {
				s.diskHitCnt++
				return true
			}
		}
		seen, inserted := s.memInsert0(fp0, start)
		if seen {
			return true
		}
		if inserted {
			if s.needsDiskFlush() {
				if err := s.evictLocked(); err != nil {
					panic(err)
				}
			}
			return false
		}
		if err := s.evictLocked(); err != nil {
			panic(err)
		}
	}
}

func (s *OffHeapDiskFPSet) Contains(fp uint64) bool {
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
		s.diskHitCnt++
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
	if s.Put(fp) {
		return fmt.Errorf("fingerprint %d already in set during recovery", fp)
	}
	return nil
}

func (s *OffHeapDiskFPSet) Sizeof() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint64(44) + uint64(s.maxTblCnt*fpSetLongSize) + uint64(len(s.index))*8
}

func (s *OffHeapDiskFPSet) GetTblCapacity() int64     { return s.maxTblCnt }
func (s *OffHeapDiskFPSet) GetTblLoad() int64         { return s.tblCnt }
func (s *OffHeapDiskFPSet) GetOverallCapacity() int64 { return s.array.Size() }
func (s *OffHeapDiskFPSet) GetBucketCapacity() int64  { return int64(s.probeLimit) }

func (s *OffHeapDiskFPSet) ForceFlush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.forceFlush = true
}

func (s *OffHeapDiskFPSet) needsDiskFlush() bool {
	return s.tblCnt >= s.maxTblCnt || s.forceFlush
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

func (s *OffHeapDiskFPSet) evictLocked() error {
	if s.tblCnt == 0 {
		s.forceFlush = false
		return nil
	}
	s.growDiskMark++
	values := s.unflushedValuesLocked()
	if len(values) == 0 {
		s.tblCnt = 0
		s.forceFlush = false
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
	s.forceFlush = false
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
		s.diskWriteCnt++
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
