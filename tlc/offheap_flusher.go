// Copyright (c) 2016 Microsoft Research. All rights reserved.
package tlc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync/atomic"
)

// RandomAccessFile's virtual read operations are also used by source subclasses.
type offHeapIndexReader interface {
	io.Reader
	Seek(int64) error
}

func (s *OffHeapDiskFPSet) writeIndex(index []uint64, raf offHeapIndexReader, length int64) error {
	for i := range index {
		pos := min(int64(i)*diskFPSetNumEntriesPerPage, length)
		if err := raf.Seek(pos * fpSetLongSize); err != nil {
			return err
		}
		var data [8]byte
		if _, err := io.ReadFull(raf, data[:]); err != nil {
			return err
		}
		index[i] = binary.BigEndian.Uint64(data[:])
	}
	return nil
}

// The number of fingerprints stored on disk smaller than fp.
func (s *OffHeapDiskFPSet) getDiskOffset(id int, fp uint64) (int64, error) {
	if s.index == nil {
		return 0, nil
	}
	indexLength := len(s.index)
	loPage, hiPage := 0, indexLength-1
	loVal, hiVal := s.index[loPage], s.index[hiPage]
	if fp <= loVal {
		return 0, nil
	}
	raf := s.braf[id]
	if fp >= hiVal {
		length, err := raf.Length()
		return length / fpSetLongSize, err
	}
	dfp := float64(int64(fp))
	for loPage < hiPage-1 {
		dhi, dlo := float64(hiPage), float64(loPage)
		dhiVal, dloVal := float64(int64(hiVal)), float64(int64(loVal))
		midPage := loPage + 1 + int(javaDoubleToInt((dhi-dlo-1)*(dfp-dloVal)/(dhiVal-dloVal)))
		if midPage == hiPage {
			midPage--
		}
		v := s.index[midPage]
		if fp < v {
			hiPage, hiVal = midPage, v
		} else if fp > v {
			loPage, loVal = midPage, v
		} else {
			return int64(midPage) * diskFPSetNumEntriesPerPage, nil
		}
	}
	if hiPage != loPage+1 {
		return 0, NewTLCRuntimeException(ECSystemIndexError)
	}
	if !(s.index[loPage] < fp && fp < s.index[hiPage]) {
		panic(NewAssertionError())
	}
	midEntry := int64(-1)
	loEntry := int64(loPage) * diskFPSetNumEntriesPerPage
	hiEntry := int64(hiPage) * diskFPSetNumEntriesPerPage
	if loPage == indexLength-2 {
		hiEntry = s.fileCnt - 1
	}
	for loEntry < hiEntry {
		midEntry = s.calculateMidEntry(loVal, hiVal, dfp, loEntry, hiEntry)
		if err := raf.Seek(midEntry * fpSetLongSize); err != nil {
			return 0, err
		}
		v, err := raf.ReadLong()
		if err != nil {
			return 0, err
		}
		if fp < uint64(v) {
			hiEntry, hiVal = midEntry, uint64(v)
		} else if fp > uint64(v) {
			loEntry, loVal = midEntry+1, uint64(v)
			midEntry = loEntry
		} else {
			break
		}
	}
	higher, err := s.isHigher(midEntry, fp, raf)
	if err != nil {
		return 0, err
	}
	if !higher {
		panic(NewAssertionError())
	}
	return midEntry, nil
}

func (s *OffHeapDiskFPSet) isHigher(midEntry int64, fp uint64, raf *BufferedRandomAccessFile) (bool, error) {
	if err := raf.Seek((midEntry - 1) * fpSetLongSize); err != nil {
		return false, err
	}
	low, err := raf.ReadLong()
	if err != nil {
		return false, err
	}
	high, err := raf.ReadLong()
	if err != nil {
		return false, err
	}
	return low < int64(fp) && int64(fp) < high, nil
}

// Preserve Iterator's virtual markNext/hasNext methods, including subclasses
// whose hasNext deliberately differs from the fixed elements count.
type offHeapMergeIterator struct {
	elements int64
	markNext func() (int64, bool)
	hasNext  func() bool
}

func (s *OffHeapDiskFPSet) mergeOffHeapEntries(inRAF *BufferedRandomAccessFile, outRAF io.Writer, itr offHeapMergeIterator, diskReads int64) error {
	value := int64(0)
	if diskReads > 0 {
		var err error
		value, err = inRAF.ReadLong()
		if err != nil {
			return err
		}
	} else if s.fileCnt != 0 {
		panic(NewAssertionError())
	}
	tableReads := itr.elements
	fp, _ := itr.markNext()
	write := func(value int64) error {
		var data [8]byte
		binary.BigEndian.PutUint64(data[:], uint64(value))
		if _, err := outRAF.Write(data[:]); err != nil {
			return err
		}
		atomic.AddUint64(&s.diskWriteCnt, 1)
		return nil
	}
	nextFP := func() {
		next, _ := itr.markNext()
		if next <= fp {
			panic(NewAssertionError())
		}
		fp = next
	}
	nextValue := func() error {
		next, err := inRAF.ReadLong()
		if err != nil {
			return err
		}
		if value >= next {
			panic(NewAssertionError())
		}
		value = next
		return nil
	}
	for {
		if value == fp {
			PrintWarning(ECTLCFPValueAlreadyOnDisk, fmt.Sprint(value))
			tableReads--
			diskReads--
			if err := write(fp); err != nil {
				return err
			}
			if tableReads > 0 {
				nextFP()
			}
			if diskReads > 0 {
				if err := nextValue(); err != nil {
					return err
				}
			}
		}
		if fp <= 0 {
			panic(NewAssertionError())
		}
		if tableReads > 0 && (fp < value || diskReads == 0) {
			if err := write(fp); err != nil {
				return err
			}
			tableReads--
			if tableReads > 0 {
				nextFP()
			}
		}
		if diskReads > 0 && (value < fp || tableReads == 0) {
			if err := write(value); err != nil {
				return err
			}
			diskReads--
			if diskReads > 0 {
				if err := nextValue(); err != nil {
					return err
				}
			}
		}
		if diskReads <= 0 && tableReads <= 0 {
			break
		}
	}
	if diskReads != 0 || tableReads != 0 {
		return NewTLCRuntimeException(ECGeneral)
	}
	if itr.hasNext() {
		panic(NewAssertionError())
	}
	return nil
}

func (s *OffHeapDiskFPSet) CheckInvariant(expectFPs ...uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Dispatch through the OffHeap flusher, rather than the heap-only base method.
	if s.tblCnt > 0 {
		if s.concurrentFlusher != nil {
			s.concurrentFlusher.prepareTable()
		} else {
			s.prepareOffHeapTableLocked()
		}
		itr := newOffHeapIterator(s.array, s.tblCnt, 0, s.indexer, true)
		if err := s.mergeOffHeapIteratorWithFlusher(offHeapMergeIterator{s.tblCnt, itr.markNext, itr.hasNext}, s.concurrentFlusher); err != nil {
			return false
		}
		ok, err := s.checkOffHeapIndex()
		if err != nil {
			return false
		}
		if !ok {
			panic(NewAssertionError())
		}
		s.tblCnt, s.tblLoad = 0, 0
		s.forceFlush.Store(false)
	}
	ok, _ := s.checkFile()
	return ok && (len(expectFPs) == 0 || uint64(s.fileCnt+s.tblCnt) == expectFPs[0])
}

func (s *OffHeapDiskFPSet) checkOffHeapInput() bool {
	size, reprobe := s.array.Size(), int64(s.probeLimit)
	for pos := int64(0); pos <= size+reprobe-1; pos++ {
		value := s.array.Get(pos % size)
		if value == 0 {
			continue
		}
		if value < 0 {
			value = int64(uint64(value) & diskFPSetFlushedMask)
		}
		idx := s.indexer.GetIdx(uint64(value))
		if pos < reprobe && idx > size-1-pos-reprobe {
			continue
		}
		if pos > size-1 && idx+reprobe < pos {
			continue
		}
		if idx <= pos && pos <= idx+reprobe {
			continue
		}
		fmt.Fprintf(os.Stderr, "%d with idx %d at pos %d (reprobe: %d).\n", value, idx, pos, reprobe)
		return false
	}
	return true
}

func (s *OffHeapDiskFPSet) checkOffHeapSorted() int64 {
	return s.checkOffHeapSortedRange(0, s.array.Size()-1+int64(s.probeLimit))
}

func (s *OffHeapDiskFPSet) checkOffHeapSortedRange(start, end int64) int64 {
	size := s.array.Size()
	reprobe := int64(s.probeLimit)
	if reprobe >= size {
		reprobe = size - 1
	}
	previous := int64(0)
	for pos := start; pos <= end; pos++ {
		value := s.array.Get(pos % size)
		if value <= 0 {
			continue
		}
		idx := s.indexer.GetIdx(uint64(value))
		if idx > pos || idx+reprobe < pos {
			continue
		}
		if previous == 0 {
			previous = value
			continue
		}
		if previous >= value {
			fmt.Fprintf(os.Stderr, "%d >= %d at pos %d.\n", previous, value, pos)
			return pos
		}
		previous = value
	}
	return -1
}

func (s *OffHeapDiskFPSet) checkOffHeapIndex() (bool, error) {
	for i := 1; i < len(s.index); i++ {
		if s.index[i-1] >= s.index[i] {
			return false, nil
		}
	}
	raf := s.braf[0]
	length, err := raf.Length()
	if err != nil {
		return false, err
	}
	length = length/fpSetLongSize - 1
	for i, fp := range s.index {
		pos := min(int64(i)*diskFPSetNumEntriesPerPage, length)
		if err := raf.Seek(pos * fpSetLongSize); err != nil {
			return false, err
		}
		value, err := raf.ReadLong()
		if err != nil {
			return false, err
		}
		if uint64(value) != fp {
			return false, nil
		}
	}
	return true, nil
}
