package tlc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	diskFPSetMarkFlushed          = uint64(0x8000000000000000)
	diskFPSetFlushedMask          = uint64(0x7fffffffffffffff)
	diskFPSetLogMaxLoad           = 4
	diskFPSetInitialBucketCap     = 1 << diskFPSetLogMaxLoad
	diskFPSetNumEntriesPerPage    = 8192 / fpSetLongSize
	diskFPSetBucketSizeIncrement  = 4
	diskFPSetLogDefaultMaxTblCnt  = 19
	diskFPSetDefaultMaxTblCnt     = 1 << diskFPSetLogDefaultMaxTblCnt
	diskFPSetModeLSB              = "lsb"
	diskFPSetModeMSB              = "msb"
	diskFPSetDefaultWorkerReaders = 1
	diskFPSetBRAFPoolSize         = 5
)

type DiskFPSet struct {
	mu sync.Mutex

	config *FPSetConfiguration

	maxTblCnt      int64
	metadir        string
	fpFilename     string
	tmpFilename    string
	filename       string
	fileCnt        int64
	tblCnt         int64
	tblLoad        int64
	bucketsCap     int64
	index          []uint64
	checkPointMark int
	growDiskMark   int
	forceFlush     bool
	flushTime      int64

	tbl          [][]uint64
	mask         uint64
	capacity     int
	logMaxMemCnt int
	lockCnt      int
	lockMask     int
	moveBy       int
	mode         string
	checkpoint   bool

	memHitCnt     uint64
	diskHitCnt    uint64
	diskLookupCnt uint64
	diskWriteCnt  uint64
	diskSeekCnt   uint64
	diskSeekCache uint64
	statesSeen    uint64

	lsbBuff []uint64

	braf      []*BufferedRandomAccessFile
	brafPool  []*BufferedRandomAccessFile
	poolIndex int
}

type HeapBasedDiskFPSet struct{ *DiskFPSet }

type LSBDiskFPSet struct{ *HeapBasedDiskFPSet }

type MSBDiskFPSet struct{ *HeapBasedDiskFPSet }

type NonCheckpointableDiskFPSet struct{ *DiskFPSet }

type OffHeapDiskFPSet struct{ *NonCheckpointableDiskFPSet }

func NewDiskFPSet(config *FPSetConfiguration) *DiskFPSet {
	return newHeapDiskFPSet(config, diskFPSetModeMSB, true)
}

func NewHeapBasedDiskFPSet(config *FPSetConfiguration) *HeapBasedDiskFPSet {
	return &HeapBasedDiskFPSet{DiskFPSet: newHeapDiskFPSet(config, diskFPSetModeLSB, true)}
}

func NewLSBDiskFPSet(config *FPSetConfiguration) *LSBDiskFPSet {
	return &LSBDiskFPSet{HeapBasedDiskFPSet: &HeapBasedDiskFPSet{DiskFPSet: newHeapDiskFPSet(config, diskFPSetModeLSB, true)}}
}

func NewMSBDiskFPSet(config *FPSetConfiguration) *MSBDiskFPSet {
	return &MSBDiskFPSet{HeapBasedDiskFPSet: &HeapBasedDiskFPSet{DiskFPSet: newHeapDiskFPSet(config, diskFPSetModeMSB, true)}}
}

func NewNonCheckpointableDiskFPSet(config *FPSetConfiguration) *NonCheckpointableDiskFPSet {
	return &NonCheckpointableDiskFPSet{DiskFPSet: newHeapDiskFPSet(config, diskFPSetModeMSB, false)}
}

func NewOffHeapDiskFPSet(config *FPSetConfiguration) *OffHeapDiskFPSet {
	return &OffHeapDiskFPSet{NonCheckpointableDiskFPSet: NewNonCheckpointableDiskFPSet(config)}
}

func newHeapDiskFPSet(config *FPSetConfiguration, mode string, checkpoint bool) *DiskFPSet {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	aux := 1.0
	switch mode {
	case diskFPSetModeMSB:
		aux = 1.5
	case diskFPSetModeLSB:
		aux = 2.5
	}
	maxMemCnt := int64(float64(config.GetMemoryInFingerprintCnt()) / aux)
	if maxMemCnt-diskFPSetLogMaxLoad <= 0 {
		maxMemCnt = diskFPSetDefaultMaxTblCnt
	}
	logMaxMemCnt := 63 - bitsLeadingZeros64(uint64(maxMemCnt))
	if logMaxMemCnt-diskFPSetLogMaxLoad < 0 {
		logMaxMemCnt = diskFPSetLogMaxLoad
	}
	capacity64 := int64(1) << uint(logMaxMemCnt-diskFPSetLogMaxLoad)
	if capacity64 <= 0 || capacity64 > int64(math.MaxInt32-8) {
		capacity64 = int64(math.MaxInt32 - 8)
	}
	capacity := int(capacity64)
	set := &DiskFPSet{
		config:       config,
		maxTblCnt:    int64(1) << uint(logMaxMemCnt),
		tbl:          make([][]uint64, capacity),
		mask:         uint64(capacity - 1),
		capacity:     capacity,
		logMaxMemCnt: logMaxMemCnt,
		lockCnt:      1,
		lockMask:     0,
		mode:         mode,
		checkpoint:   checkpoint,
	}
	if mode == diskFPSetModeMSB {
		fpBits := config.GetFPBits()
		if fpBits <= 0 {
			fpBits = 1
		}
		set.moveBy = (32 - fpBits) - (logMaxMemCnt - diskFPSetLogMaxLoad)
		if set.moveBy < 0 {
			set.moveBy = 0
		}
		set.mask = uint64(capacity-1) << uint(set.moveBy)
	}
	return set
}

func (s *DiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	if numThreads <= 0 {
		numThreads = diskFPSetDefaultWorkerReaders
	}
	s.metadir = metadir
	s.filename = filename
	if s.metadir == "" {
		s.metadir = filepath.Join(os.TempDir(), "DiskFPSet")
	}
	_ = os.MkdirAll(s.metadir, 0o755)
	base := filepath.Join(s.metadir, filename)
	if filename == "" {
		base = filepath.Join(s.metadir, "fpset")
	}
	s.tmpFilename = base + ".tmp"
	s.fpFilename = base + ".fp"
	_ = os.WriteFile(s.fpFilename, nil, 0o644)
	if err := s.openBRAFReaders(numThreads, diskFPSetBRAFPoolSize); err != nil {
		panic(err)
	}
	s.fileCnt = 0
	s.index = nil
	s.clearTable()
	return s
}

func (s *HeapBasedDiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.DiskFPSet.Init(numThreads, metadir, filename)
	return s
}

func (s *LSBDiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.DiskFPSet.Init(numThreads, metadir, filename)
	return s
}

func (s *MSBDiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.DiskFPSet.Init(numThreads, metadir, filename)
	return s
}

func (s *NonCheckpointableDiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.DiskFPSet.Init(numThreads, metadir, filename)
	return s
}

func (s *OffHeapDiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.DiskFPSet.Init(numThreads, metadir, filename)
	return s
}

func (s *DiskFPSet) Size() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint64(s.tblCnt + s.fileCnt)
}

func (s *DiskFPSet) Sizeof() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	size := uint64(44)
	size += 16 + uint64(len(s.tbl))*8
	for _, bucket := range s.tbl {
		if bucket != nil {
			size += 16 + uint64(len(bucket))*fpSetLongSize
		}
	}
	size += uint64(len(s.index)) * 8
	return size
}

func (s *DiskFPSet) Put(fp uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	fp0 := s.checkValid(fp) & diskFPSetFlushedMask
	if s.memLookup(fp0) {
		s.memHitCnt++
		return true
	}
	if hit, _ := s.diskLookup(fp0); hit {
		s.diskHitCnt++
		return true
	}
	if s.memInsert(fp0) {
		s.memHitCnt++
		return true
	}
	if s.needsDiskFlush() {
		s.growDiskMark++
		start := time.Now()
		insertions := s.tblCnt
		if err := s.flushTable(); err != nil {
			panic(err)
		}
		s.forceFlush = false
		_ = insertions
		s.flushTime += int64(time.Since(start) / time.Millisecond)
	}
	return false
}

func (s *DiskFPSet) Contains(fp uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	fp0 := s.checkValid(fp) & diskFPSetFlushedMask
	if s.memLookup(fp0) {
		s.memHitCnt++
		return true
	}
	hit, _ := s.diskLookup(fp0)
	if hit {
		s.diskHitCnt++
	}
	return hit
}

func (s *DiskFPSet) PutBlock(fpv *LongVec) *BitVector {
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

func (s *DiskFPSet) ContainsBlock(fpv *LongVec) *BitVector {
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

func (s *DiskFPSet) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.closeBRAFReaders()
}

func (s *DiskFPSet) AddThread() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raf, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
	if err != nil {
		return err
	}
	s.braf = append(s.braf, raf)
	return nil
}

func (s *DiskFPSet) IncWorkers(num int) {
	if num <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := 0; i < num; i++ {
		raf, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
		if err != nil {
			panic(err)
		}
		s.braf = append(s.braf, raf)
	}
}

func (s *DiskFPSet) Exit(cleanup bool) error {
	s.Close()
	if cleanup && s.metadir != "" {
		return os.RemoveAll(s.metadir)
	}
	return nil
}

func (s *DiskFPSet) BeginChkpt() error {
	return nil
}

func (s *DiskFPSet) CommitChkpt() error {
	return nil
}

func (s *DiskFPSet) BeginChkptFile(fname string) error {
	if !s.checkpoint {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.flushTable(); err != nil {
		return err
	}
	if err := copyFile(s.fpFilename, s.chkptName(fname, "tmp")); err != nil {
		return err
	}
	s.checkPointMark++
	return nil
}

func (s *DiskFPSet) CommitChkptFile(fname string) error {
	if !s.checkpoint {
		return nil
	}
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(newChkpt, oldChkpt)
}

func (s *DiskFPSet) Recover() error {
	return nil
}

func (s *DiskFPSet) RecoverFile(fname string) error {
	if !s.checkpoint {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recoverFileLocked(s.chkptName(fname, "chkpt"))
}

func (s *DiskFPSet) RecoverTrace(trace *TLCTrace) error {
	if trace == nil {
		return s.Recover()
	}
	for _, record := range trace.Records() {
		if err := s.RecoverFP(record.FP); err != nil {
			return err
		}
	}
	return nil
}

func (s *DiskFPSet) RecoverFP(fp uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fp0 := fp & diskFPSetFlushedMask
	if s.memInsert(fp0) {
		return fmt.Errorf("duplicate fingerprint %d during DiskFPSet recovery", fp0)
	}
	if s.needsDiskFlush() {
		return s.flushTable()
	}
	return nil
}

func (s *DiskFPSet) CheckFPs() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.flushTable(); err != nil {
		return 0
	}
	values, err := readFingerprintFile(s.fpFilename)
	if err != nil {
		return 0
	}
	dis := uint64(1<<63 - 1)
	for i := 1; i < len(values); i++ {
		if values[i] >= values[i-1] {
			dis = minUint64(dis, values[i]-values[i-1])
		}
	}
	return dis
}

func (s *DiskFPSet) CheckInvariant(expectFPs ...uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.flushTable(); err != nil {
		return false
	}
	ok, count := s.checkFile()
	if !ok {
		return false
	}
	return len(expectFPs) == 0 || uint64(count) == expectFPs[0]
}

func (s *DiskFPSet) UnexportObject(force bool) {}

func (s *DiskFPSet) GetStatesSeen() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statesSeen
}

func (s *DiskFPSet) GetConfiguration() *FPSetConfiguration {
	if s == nil || s.config == nil {
		return NewFPSetConfiguration()
	}
	return s.config
}

func (s *DiskFPSet) GetBucketCapacity() int64 { return s.bucketsCap }
func (s *DiskFPSet) GetTblCapacity() int64    { return int64(len(s.tbl)) }
func (s *DiskFPSet) GetIndexCapacity() int64  { return int64(len(s.index)) }
func (s *DiskFPSet) GetOverallCapacity() int64 {
	return s.GetBucketCapacity() + s.GetTblCapacity() + s.GetIndexCapacity()
}
func (s *DiskFPSet) GetTblLoad() int64        { return s.tblLoad }
func (s *DiskFPSet) GetTblCnt() int64         { return s.tblCnt }
func (s *DiskFPSet) GetMaxTblCnt() int64      { return s.maxTblCnt }
func (s *DiskFPSet) GetFileCnt() int64        { return s.fileCnt }
func (s *DiskFPSet) GetDiskLookupCnt() uint64 { return s.diskLookupCnt }
func (s *DiskFPSet) GetMemHitCnt() uint64     { return s.memHitCnt }
func (s *DiskFPSet) GetDiskHitCnt() uint64    { return s.diskHitCnt }
func (s *DiskFPSet) GetDiskWriteCnt() uint64  { return s.diskWriteCnt }
func (s *DiskFPSet) GetDiskSeekCnt() uint64   { return s.diskSeekCnt }
func (s *DiskFPSet) GetDiskSeekCache() uint64 { return s.diskSeekCache }
func (s *DiskFPSet) GetGrowDiskMark() int     { return s.growDiskMark }
func (s *DiskFPSet) GetCheckPointMark() int   { return s.checkPointMark }
func (s *DiskFPSet) GetFlushTime() int64      { return s.flushTime }
func (s *DiskFPSet) ForceFlush()              { s.forceFlush = true }
func (s *DiskFPSet) GetLockCnt() int          { return s.lockCnt }
func (s *DiskFPSet) GetReaderWriterCnt() int  { return len(s.braf) + len(s.brafPool) }
func (s *DiskFPSet) GetLoadFactor() float64 {
	if s.maxTblCnt == 0 {
		return 0
	}
	return float64(s.tblCnt) / float64(s.maxTblCnt)
}

func (s *DiskFPSet) checkValid(fp uint64) uint64 {
	return fp
}

func (s *DiskFPSet) needsDiskFlush() bool {
	return s.tblCnt >= s.maxTblCnt || s.forceFlush
}

func (s *DiskFPSet) memLookup(fp uint64) bool {
	if len(s.tbl) == 0 {
		return false
	}
	bucket := s.tbl[s.getIndex(fp)]
	for i := 0; i < len(bucket) && bucket[i] != 0; i++ {
		if fp == (bucket[i] & diskFPSetFlushedMask) {
			return true
		}
	}
	return false
}

func (s *DiskFPSet) memInsert(fp uint64) bool {
	idx := s.getIndex(fp)
	bucket := s.tbl[idx]
	if bucket == nil {
		bucket = make([]uint64, diskFPSetInitialBucketCap)
		bucket[0] = fp
		s.tbl[idx] = bucket
		s.bucketsCap += diskFPSetInitialBucketCap
		s.tblLoad++
		s.tblCnt++
		return false
	}
	bucketLen := len(bucket)
	reusable := -1
	j := 0
	for ; j < bucketLen && bucket[j] != 0; j++ {
		fp1 := bucket[j]
		if fp == (fp1 & diskFPSetFlushedMask) {
			return true
		}
		if reusable == -1 && int64(fp1) < 0 {
			reusable = j
		}
	}
	if reusable == -1 {
		if j == bucketLen {
			old := bucket
			bucket = make([]uint64, bucketLen+diskFPSetBucketSizeIncrement)
			copy(bucket, old)
			s.tbl[idx] = bucket
			s.bucketsCap += diskFPSetBucketSizeIncrement
		}
		bucket[j] = fp
	} else {
		if j != bucketLen {
			bucket[j] = bucket[reusable]
		}
		bucket[reusable] = fp
	}
	s.tblCnt++
	return false
}

func (s *DiskFPSet) getIndex(fp uint64) int {
	if s.mode == diskFPSetModeMSB {
		return int((uint64(uint32(fp>>32)) & s.mask) >> uint(s.moveBy))
	}
	return int(fp & s.mask)
}

func (s *DiskFPSet) diskLookup(fp uint64) (bool, error) {
	if s.index == nil || len(s.index) == 0 {
		return false, nil
	}
	s.diskLookupCnt++
	indexLength := len(s.index)
	loPage, hiPage := 0, indexLength-1
	loVal, hiVal := s.index[loPage], s.index[hiPage]
	if fp < loVal || fp > hiVal {
		return false, nil
	}
	if fp == hiVal {
		return true, nil
	}
	dfp := float64(fp)
	for loPage < hiPage-1 {
		midPage := (loPage + 1) + int((float64(hiPage-loPage-1))*(dfp-float64(loVal))/(float64(hiVal)-float64(loVal)))
		if midPage == hiPage {
			midPage--
		}
		if midPage <= loPage || midPage >= hiPage {
			midPage = loPage + ((hiPage - loPage) / 2)
		}
		v := s.index[midPage]
		if fp < v {
			hiPage = midPage
			hiVal = v
		} else if fp > v {
			loPage = midPage
			loVal = v
		} else {
			return true, nil
		}
	}
	loEntry := int64(loPage * diskFPSetNumEntriesPerPage)
	hiEntry := int64(hiPage * diskFPSetNumEntriesPerPage)
	if loPage == indexLength-2 {
		hiEntry = s.fileCnt - 1
	}
	for loEntry <= hiEntry {
		midEntry := s.calculateMidEntry(loVal, hiVal, dfp, loEntry, hiEntry)
		if midEntry < loEntry || midEntry > hiEntry {
			midEntry = loEntry + ((hiEntry - loEntry) / 2)
		}
		v, err := s.readDiskFP(midEntry)
		if err != nil {
			return false, err
		}
		if fp < v {
			hiEntry = midEntry - 1
			hiVal = v
		} else if fp > v {
			loEntry = midEntry + 1
			loVal = v
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (s *DiskFPSet) calculateMidEntry(loVal uint64, hiVal uint64, dfp float64, loEntry int64, hiEntry int64) int64 {
	if hiVal == loVal {
		return loEntry
	}
	midEntry := loEntry + int64((float64(hiEntry-loEntry))*(dfp-float64(loVal))/(float64(hiVal)-float64(loVal)))
	if midEntry == hiEntry && hiEntry > loEntry {
		midEntry--
	}
	return midEntry
}

func (s *DiskFPSet) readDiskFP(entry int64) (uint64, error) {
	raf, pooled, err := s.openDiskReader()
	if err != nil {
		return 0, err
	}
	if pooled {
		defer s.poolClose(raf)
	}
	seeked, err := raf.Seeek(entry * fpSetLongSize)
	if err != nil {
		return 0, err
	}
	if seeked {
		s.diskSeekCnt++
	} else {
		s.diskSeekCache++
	}
	value, err := raf.ReadLong()
	if err != nil {
		return 0, err
	}
	return uint64(value), nil
}

func (s *DiskFPSet) flushTable() error {
	if s.tblCnt == 0 {
		return nil
	}
	switch s.mode {
	case diskFPSetModeMSB:
		s.prepareMSBTable()
	case diskFPSetModeLSB:
		s.prepareLSBTable()
	default:
		s.prepareMSBTable()
	}
	if err := s.mergeNewEntries(); err != nil {
		return err
	}
	s.tblCnt = 0
	s.bucketsCap = 0
	s.tblLoad = 0
	s.lsbBuff = nil
	return nil
}

func (s *DiskFPSet) prepareMSBTable() {
	for _, bucket := range s.tbl {
		if bucket == nil {
			continue
		}
		k := 0
		for ; k < len(bucket) && int64(bucket[k]) > 0; k++ {
		}
		sort.Slice(bucket[:k], func(i, j int) bool { return bucket[i] < bucket[j] })
	}
}

func (s *DiskFPSet) prepareLSBTable() {
	cnt := int(s.tblCnt)
	s.lsbBuff = make([]uint64, 0, cnt)
	for _, bucket := range s.tbl {
		for k := 0; k < len(bucket) && int64(bucket[k]) > 0; k++ {
			s.lsbBuff = append(s.lsbBuff, bucket[k])
			bucket[k] |= diskFPSetMarkFlushed
		}
	}
	sort.Slice(s.lsbBuff, func(i, j int) bool { return s.lsbBuff[i] < s.lsbBuff[j] })
}

func (s *DiskFPSet) mergeNewEntries() error {
	oldValues, err := readFingerprintFile(s.fpFilename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var newValues []uint64
	if s.mode == diskFPSetModeLSB {
		newValues = s.lsbBuff
	} else {
		itr := newMSBDiskIterator(s.tbl)
		newValues = make([]uint64, 0, s.tblCnt)
		for itr.hasNext() {
			next, err := itr.next()
			if err != nil {
				return err
			}
			newValues = append(newValues, next)
		}
	}
	if len(newValues) == 0 {
		return nil
	}
	total := int64(len(oldValues) + len(newValues))
	indexLen := s.calculateIndexLen(int64(len(newValues)))
	newIndex := make([]uint64, indexLen)
	maxVal := newValues[len(newValues)-1]
	if len(s.index) > 0 && s.index[len(s.index)-1] > maxVal {
		maxVal = s.index[len(s.index)-1]
	}
	newIndex[indexLen-1] = maxVal

	if err := os.MkdirAll(filepath.Dir(s.tmpFilename), 0o755); err != nil {
		return err
	}
	tmp, err := os.Create(s.tmpFilename)
	if err != nil {
		return err
	}
	currIndex := 0
	counter := 0
	writeFP := func(fp uint64) error {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], fp)
		if _, err := tmp.Write(buf[:]); err != nil {
			return err
		}
		s.diskWriteCnt++
		if counter == 0 {
			if currIndex >= len(newIndex)-1 {
				_ = tmp.Close()
				return fmt.Errorf("DiskFPSet index overflow")
			}
			newIndex[currIndex] = fp
			currIndex++
			counter = diskFPSetNumEntriesPerPage
		}
		counter--
		return nil
	}
	i, j := 0, 0
	for i < len(oldValues) && j < len(newValues) {
		if oldValues[i] < newValues[j] {
			if err := writeFP(oldValues[i]); err != nil {
				_ = tmp.Close()
				return err
			}
			i++
		} else if oldValues[i] > newValues[j] {
			if err := writeFP(newValues[j]); err != nil {
				_ = tmp.Close()
				return err
			}
			j++
		} else {
			_ = tmp.Close()
			return fmt.Errorf("fingerprint %d already on disk", oldValues[i])
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
	if err := tmp.Close(); err != nil {
		return err
	}
	readerCnt := len(s.braf)
	poolCnt := len(s.brafPool)
	if err := s.closeBRAFReaders(); err != nil {
		return err
	}
	if currIndex != indexLen-1 {
		_ = s.openBRAFReaders(readerCnt, poolCnt)
		return fmt.Errorf("DiskFPSet index mismatch: got %d want %d", currIndex, indexLen-1)
	}
	if err := replaceFile(s.tmpFilename, s.fpFilename); err != nil {
		_ = s.openBRAFReaders(readerCnt, poolCnt)
		return err
	}
	if err := s.openBRAFReaders(readerCnt, poolCnt); err != nil {
		return err
	}
	s.index = newIndex
	s.fileCnt = total
	return nil
}

func (s *DiskFPSet) calculateIndexLen(buffLen int64) int {
	indexLen := ((s.fileCnt + buffLen - 1) / diskFPSetNumEntriesPerPage) + 2
	if indexLen <= 0 {
		indexLen = 2
	}
	return int(indexLen)
}

func (s *DiskFPSet) recoverFileLocked(path string) error {
	values, err := readFingerprintFile(path)
	if err != nil {
		return err
	}
	s.clearTable()
	if err := os.MkdirAll(filepath.Dir(s.fpFilename), 0o755); err != nil {
		return err
	}
	out, err := os.Create(s.fpFilename)
	if err != nil {
		return err
	}
	var predecessor uint64
	for i, fp := range values {
		if i > 0 && predecessor >= fp {
			_ = out.Close()
			return fmt.Errorf("checkpoint fingerprints out of order")
		}
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], fp)
		if _, err := out.Write(buf[:]); err != nil {
			_ = out.Close()
			return err
		}
		predecessor = fp
	}
	if err := out.Close(); err != nil {
		return err
	}
	s.fileCnt = int64(len(values))
	s.rebuildIndex(values)
	if err := s.reopenBRAFReaders(); err != nil {
		return err
	}
	return nil
}

func (s *DiskFPSet) openBRAFReaders(numReaders int, poolSize int) error {
	if numReaders <= 0 {
		numReaders = diskFPSetDefaultWorkerReaders
	}
	if poolSize <= 0 {
		poolSize = diskFPSetBRAFPoolSize
	}
	s.braf = make([]*BufferedRandomAccessFile, numReaders)
	for i := range s.braf {
		raf, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
		if err != nil {
			_ = s.closeBRAFReaders()
			return err
		}
		s.braf[i] = raf
	}
	s.brafPool = make([]*BufferedRandomAccessFile, poolSize)
	for i := range s.brafPool {
		raf, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
		if err != nil {
			_ = s.closeBRAFReaders()
			return err
		}
		s.brafPool[i] = raf
	}
	s.poolIndex = 0
	return nil
}

func (s *DiskFPSet) closeBRAFReaders() error {
	var firstErr error
	for i, raf := range s.braf {
		if err := raf.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.braf[i] = nil
	}
	for i, raf := range s.brafPool {
		if err := raf.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.brafPool[i] = nil
	}
	s.braf = nil
	s.brafPool = nil
	s.poolIndex = 0
	return firstErr
}

func (s *DiskFPSet) reopenBRAFReaders() error {
	readerCnt := len(s.braf)
	poolCnt := len(s.brafPool)
	if readerCnt <= 0 {
		readerCnt = diskFPSetDefaultWorkerReaders
	}
	if poolCnt <= 0 {
		poolCnt = diskFPSetBRAFPoolSize
	}
	if err := s.closeBRAFReaders(); err != nil {
		return err
	}
	return s.openBRAFReaders(readerCnt, poolCnt)
}

func (s *DiskFPSet) openDiskReader() (*BufferedRandomAccessFile, bool, error) {
	if len(s.braf) > 0 && s.braf[0] != nil {
		return s.braf[0], false, nil
	}
	if len(s.brafPool) > 0 && s.poolIndex < len(s.brafPool) {
		raf := s.brafPool[s.poolIndex]
		s.poolIndex++
		if raf != nil {
			return raf, true, nil
		}
	}
	raf, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
	return raf, true, err
}

func (s *DiskFPSet) poolClose(raf *BufferedRandomAccessFile) {
	if raf == nil {
		return
	}
	if len(s.brafPool) > 0 && s.poolIndex > 0 {
		s.poolIndex--
		s.brafPool[s.poolIndex] = raf
		return
	}
	_ = raf.Close()
}

func (s *DiskFPSet) rebuildIndex(values []uint64) {
	if len(values) == 0 {
		s.index = nil
		return
	}
	indexLen := ((int64(len(values)) - 1) / diskFPSetNumEntriesPerPage) + 2
	index := make([]uint64, indexLen)
	curr := 0
	counter := 0
	for _, fp := range values {
		if counter == 0 {
			index[curr] = fp
			curr++
			counter = diskFPSetNumEntriesPerPage
		}
		counter--
	}
	index[len(index)-1] = values[len(values)-1]
	s.index = index
}

func (s *DiskFPSet) clearTable() {
	for i := range s.tbl {
		s.tbl[i] = nil
	}
	s.tblCnt = 0
	s.tblLoad = 0
	s.bucketsCap = 0
	s.lsbBuff = nil
}

func (s *DiskFPSet) checkFile() (bool, int64) {
	values, err := readFingerprintFile(s.fpFilename)
	if err != nil {
		return false, 0
	}
	for i := 1; i < len(values); i++ {
		if values[i-1] >= values[i] {
			return false, int64(len(values))
		}
	}
	if len(values) > 0 && len(s.index) > 0 {
		if values[0] != s.index[0] || values[len(values)-1] != s.index[len(s.index)-1] {
			return false, int64(len(values))
		}
	}
	return true, int64(len(values))
}

func (s *DiskFPSet) chkptName(fname string, ext string) string {
	if fname == "" {
		fname = s.filename
	}
	if fname == "" {
		fname = "fpset"
	}
	return filepath.Join(s.metadir, fname+".fp."+ext)
}

type msbDiskIterator struct {
	buff         [][]uint64
	firstIdx     int
	secondIdx    int
	previous     uint64
	havePrevious bool
	readElements int64
}

func newMSBDiskIterator(buff [][]uint64) *msbDiskIterator {
	return &msbDiskIterator{buff: buff}
}

func (i *msbDiskIterator) hasNext() bool {
	if i.firstIdx >= len(i.buff) {
		return false
	}
	bucket := i.buff[i.firstIdx]
	if bucket != nil && i.secondIdx < len(bucket) && int64(bucket[i.secondIdx]) > 0 {
		return true
	}
	for idx := i.firstIdx + 1; idx < len(i.buff); idx++ {
		if i.buff[idx] != nil && len(i.buff[idx]) > 0 && int64(i.buff[idx][0]) > 0 {
			return true
		}
	}
	return false
}

func (i *msbDiskIterator) next() (uint64, error) {
	var result uint64
	found := false
	if i.firstIdx < len(i.buff) {
		bucket := i.buff[i.firstIdx]
		if bucket != nil && i.secondIdx < len(bucket) && int64(bucket[i.secondIdx]) > 0 {
			result = bucket[i.secondIdx]
			bucket[i.secondIdx] |= diskFPSetMarkFlushed
			i.secondIdx++
			found = true
		} else {
			for idx := i.firstIdx + 1; idx < len(i.buff); idx++ {
				if i.buff[idx] != nil && len(i.buff[idx]) > 0 && int64(i.buff[idx][0]) > 0 {
					i.firstIdx = idx
					i.secondIdx = 0
					result = i.buff[i.firstIdx][i.secondIdx]
					i.buff[i.firstIdx][i.secondIdx] |= diskFPSetMarkFlushed
					i.secondIdx++
					found = true
					break
				}
			}
		}
	}
	if !found {
		return 0, io.EOF
	}
	if i.havePrevious && i.previous >= result {
		return 0, fmt.Errorf("MSBDiskFPSet iterator is not strictly increasing")
	}
	i.previous = result
	i.havePrevious = true
	i.readElements++
	return result, nil
}

func readFingerprintFile(path string) ([]uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size()%fpSetLongSize != 0 {
		return nil, fmt.Errorf("fingerprint file %s has invalid length %d", path, info.Size())
	}
	values := make([]uint64, 0, info.Size()/fpSetLongSize)
	var buf [8]byte
	for {
		_, err := io.ReadFull(file, buf[:])
		if errors.Is(err, io.EOF) {
			return values, nil
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		values = append(values, binary.BigEndian.Uint64(buf[:]))
	}
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func replaceFile(src string, dst string) error {
	if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(src, dst)
}

func bitsLeadingZeros64(x uint64) int {
	if x == 0 {
		return 64
	}
	n := 0
	for bit := uint64(1) << 63; bit != 0 && x&bit == 0; bit >>= 1 {
		n++
	}
	return n
}
