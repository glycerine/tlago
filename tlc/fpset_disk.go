package tlc

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	diskFPSetMarkFlushed           = uint64(0x8000000000000000)
	diskFPSetFlushedMask           = uint64(0x7fffffffffffffff)
	diskFPSetJavaRefSize           = 4
	diskFPSetLogMaxLoad            = 4
	diskFPSetInitialBucketCap      = 1 << diskFPSetLogMaxLoad
	diskFPSetNumEntriesPerPage     = 8192 / fpSetLongSize
	diskFPSetBucketSizeIncrement   = 4
	diskFPSetLogDefaultMaxTblCnt   = 19
	diskFPSetDefaultMaxTblCnt      = 1 << diskFPSetLogDefaultMaxTblCnt
	diskFPSetModeLSB               = "lsb"
	diskFPSetModeMSB               = "msb"
	diskFPSetBRAFPoolSize          = 5
	DiskFPSetLogLockCntProperty    = "tlc2.tool.fp.DiskFPSet.logLockCnt"
	DiskFPSetMetadirPrefixProperty = "tlc2.tool.fp.DiskFPSet.metadirPrefix"
	DiskFPSetError2WarningProperty = "tlc2.tool.fp.DiskFPSet.error2warning"
)

type DiskFPSet struct {
	fpSetLifecycle
	mu     sync.Mutex
	poolMu sync.Mutex
	// A source array reference is read atomically. Publish the native slice
	// header and its entries together when readers are added or reopened.
	readers atomic.Pointer[[]*BufferedRandomAccessFile]

	config *FPSetConfiguration

	maxTblCnt   int64
	metadir     string
	fpFilename  string
	tmpFilename string
	filename    string
	// Counters use atomic accesses for source LongAdder reads and disk-count
	// publication; plain reads race with concurrent writers and reporters.
	fileCnt        int64
	tblCnt         int64
	tblLoad        int64
	bucketsCap     int64
	index          []uint64
	checkPointMark int
	growDiskMark   int
	forceFlush     atomic.Bool
	flusherChosen  atomic.Bool
	flushTime      int64

	tbl          [][]uint64
	mask         uint64
	rwLock       *Striped
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

type OffHeapDiskFPSet struct {
	*NonCheckpointableDiskFPSet
	array             *LongArray
	indexer           *OffHeapIndexer
	numThreads        int
	probeLimit        int
	concurrentFlusher *offHeapConcurrentFlusher
}

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
	if config == nil {
		config = NewFPSetConfiguration()
	}
	maxTblCnt := config.GetMemoryInFingerprintCnt()
	if maxTblCnt <= 0 {
		panic(NewIllegalArgumentException("Negative or zero upper storage limit"))
	}
	return &NonCheckpointableDiskFPSet{DiskFPSet: &DiskFPSet{
		config: config, maxTblCnt: maxTblCnt, mode: diskFPSetModeMSB,
	}}
}

func NewOffHeapDiskFPSet(config *FPSetConfiguration) *OffHeapDiskFPSet {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	base := NewNonCheckpointableDiskFPSet(config)
	positions := config.GetMemoryInFingerprintCnt()
	set := &OffHeapDiskFPSet{
		NonCheckpointableDiskFPSet: base,
		array:                      NewLongArray(positions),
		indexer:                    NewOffHeapIndexer(positions, config.GetFPBits()),
		probeLimit:                 offHeapDiskFPSetProbeLimit(),
	}
	offHeapGlobalSync.add(set)
	return set
}

func newHeapDiskFPSet(config *FPSetConfiguration, mode string, checkpoint bool) *DiskFPSet {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	lockCnt := diskFPSetLockCount()
	rwLock := StripedReadWriteLock(lockCnt)
	aux := (&DiskFPSet{mode: mode}).GetAuxiliaryStorageRequirement()
	maxMemCnt := int64(float64(config.GetMemoryInFingerprintCnt()) / aux)
	if maxMemCnt-diskFPSetLogMaxLoad <= 0 {
		maxMemCnt = diskFPSetDefaultMaxTblCnt
	}
	logMaxMemCnt := 63 - bitsLeadingZeros64(uint64(maxMemCnt))
	if logMaxMemCnt-diskFPSetLogMaxLoad < 0 {
		panic(NewTLCRuntimeExceptionMessage("Underflow when computing HeapBasedDiskFPSet"))
	}
	// Java shifts an int here, including its signed overflow and masked distance.
	cap := int32(1) << uint((logMaxMemCnt-diskFPSetLogMaxLoad)&31)
	capacity := int(cap)
	if cap < 0 {
		capacity = math.MaxInt32 - 8
	}
	maxTblCnt := int64(1) << uint(logMaxMemCnt)
	if maxTblCnt > config.GetMemoryInFingerprintCnt() {
		panic(NewTLCRuntimeExceptionMessage("Exceeded upper memory storage limit"))
	}
	if !(maxTblCnt > int64(capacity) && capacity > 0) {
		panic(NewTLCRuntimeExceptionMessage("negative maxTblCnt"))
	}
	set := &DiskFPSet{
		config:       config,
		maxTblCnt:    maxTblCnt,
		tbl:          make([][]uint64, capacity),
		mask:         uint64(capacity - 1),
		rwLock:       rwLock,
		capacity:     capacity,
		logMaxMemCnt: logMaxMemCnt,
		lockCnt:      lockCnt,
		lockMask:     lockCnt - 1,
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

func diskFPSetLockCount() int {
	if value, ok := tlcLookupSystemProperty(DiskFPSetLogLockCntProperty); ok {
		if logLockCnt, ok := javaIntProperty(value); ok && logLockCnt >= 0 {
			return 1 << uint(logLockCnt)
		}
	}
	if value := os.Getenv("TLAGO_DISK_FPSET_LOG_LOCK_CNT"); value != "" {
		if logLockCnt, ok := javaIntProperty(value); ok && logLockCnt >= 0 {
			return 1 << uint(logLockCnt)
		}
	}
	workers := NumWorkers()
	if workers < 1 {
		workers = 1
	}
	logLockCnt := 31 - bitsLeadingZeros32(uint32(workers)) + 8
	return 1 << uint(logLockCnt)
}

func diskFPSetMetadir(metadir string) string {
	if prefix, ok := tlcLookupSystemProperty(DiskFPSetMetadirPrefixProperty); ok {
		if filepath.IsAbs(metadir) {
			metadir = distributedJavaFileName(metadir)
		}
		folder := prefix + string(os.PathSeparator) + metadir
		filenameFileMkdirs(folder)
		return folder
	}
	return metadir
}

func diskFPSetError2Warning() bool {
	if value, ok := tlcLookupSystemProperty(DiskFPSetError2WarningProperty); ok {
		return javaBooleanProperty(value)
	}
	return false
}

func (s *DiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	if numThreads < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(numThreads)))
	}
	s.metadir = diskFPSetMetadir(metadir)
	s.filename = filename
	if s.metadir == "" {
		s.metadir = filepath.Join(os.TempDir(), "DiskFPSet")
	}
	base := s.metadir + string(os.PathSeparator) + filename
	if filename == "" {
		base = filepath.Join(s.metadir, "fpset")
	}
	s.tmpFilename = base + ".tmp"
	s.fpFilename = base + ".fp"
	if err := os.WriteFile(s.fpFilename, nil, 0o644); err != nil {
		panic(diskFPSetInitIOException(s.fpFilename, err))
	}
	if err := s.openBRAFReaders(numThreads, diskFPSetBRAFPoolSize); err != nil {
		panic(diskFPSetInitIOException(s.fpFilename, err))
	}
	atomic.StoreInt64(&s.fileCnt, 0)
	s.index = nil
	s.clearTable()
	return s
}

func diskFPSetInitIOException(filename string, failure error) error {
	message := javaThrowableDetailMessage(failure)
	if _, native := failure.(*os.PathError); native {
		message = javaThrowableDetailMessage(distributedFileOpenException(filename, failure))
	}
	return NewIOException(GetMessageNullable(ECSystemUnableToOpenFile, javaString(filename), message))
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

func (s *NonCheckpointableDiskFPSet) BeginChkptFile(fname string) error {
	printNonCheckpointableFPSetWarning("tlc2.tool.fp.NonCheckpointableDiskFPSet")
	return nil
}

func (s *NonCheckpointableDiskFPSet) CommitChkptFile(fname string) error {
	printNonCheckpointableFPSetWarning("tlc2.tool.fp.NonCheckpointableDiskFPSet")
	return nil
}

func (s *NonCheckpointableDiskFPSet) RecoverFile(fname string) error {
	printNonCheckpointableFPSetWarning("tlc2.tool.fp.NonCheckpointableDiskFPSet")
	return nil
}

func (s *OffHeapDiskFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.DiskFPSet.Init(numThreads, metadir, filename)
	s.numThreads = numThreads
	offHeapGlobalSync.add(s)
	if err := s.array.ZeroMemory(numThreads); err != nil {
		panic(err)
	}
	return s
}

func (s *OffHeapDiskFPSet) BeginChkptFile(fname string) error {
	printNonCheckpointableFPSetWarning("tlc2.tool.fp.OffHeapDiskFPSet")
	return nil
}

func (s *OffHeapDiskFPSet) CommitChkptFile(fname string) error {
	printNonCheckpointableFPSetWarning("tlc2.tool.fp.OffHeapDiskFPSet")
	return nil
}

func (s *OffHeapDiskFPSet) RecoverFile(fname string) error {
	printNonCheckpointableFPSetWarning("tlc2.tool.fp.OffHeapDiskFPSet")
	return nil
}

func printNonCheckpointableFPSetWarning(className string) {
	PrintWarning(ECGeneral, "Checkpointing is not implemented for "+className)
}

// Source size reads the concurrent table counter and disk count independently;
// acquiring all table locks both changed that behavior and missed offheap writers.
func (s *DiskFPSet) Size() uint64 {
	return uint64(atomic.LoadInt64(&s.tblCnt) + atomic.LoadInt64(&s.fileCnt))
}

func (s *DiskFPSet) Sizeof() uint64 {
	s.acquireTblWriteLock()
	defer s.releaseTblWriteLock()
	size := uint64(44)
	size += 16 + uint64(len(s.tbl))*diskFPSetJavaRefSize
	for _, bucket := range s.tbl {
		if bucket != nil {
			size += 16 + uint64(len(bucket))*fpSetLongSize
		}
	}
	size += uint64(len(s.index)) * diskFPSetJavaRefSize
	return size
}

func (s *DiskFPSet) Put(fp uint64) bool {
	fp0 := s.checkValid(fp) & diskFPSetFlushedMask
	lockIndex := s.getLockIndex(fp0)
	lock := s.rwLock.GetAt(lockIndex)

	lock.RLock()
	if s.memLookup(fp0) {
		lock.RUnlock()
		atomic.AddUint64(&s.memHitCnt, 1)
		return true
	}
	hit, err := s.diskLookup(fp0)
	if err != nil {
		lock.RUnlock()
		panic(err)
	}
	if hit {
		lock.RUnlock()
		atomic.AddUint64(&s.diskHitCnt, 1)
		return true
	}
	lock.RUnlock()

	lock.Lock()
	defer lock.Unlock()
	if s.memInsert(fp0) {
		atomic.AddUint64(&s.memHitCnt, 1)
		return true
	}
	if s.needsDiskFlush() && s.flusherChosen.CompareAndSwap(false, true) {
		s.mu.Lock()
		s.growDiskMark++
		insertions := atomic.LoadInt64(&s.tblCnt)
		s.mu.Unlock()
		start := time.Now()
		s.rwLock.AcquireAllLocksExcept(lockIndex)
		if err := s.flushTable(); err != nil {
			s.rwLock.ReleaseAllLocksExcept(lockIndex)
			s.flusherChosen.Store(false)
			panic(err)
		}
		s.rwLock.ReleaseAllLocksExcept(lockIndex)
		s.forceFlush.Store(false)
		s.flusherChosen.Store(false)
		_ = insertions
		s.mu.Lock()
		s.flushTime += int64(time.Since(start) / time.Millisecond)
		s.mu.Unlock()
	}
	return false
}

func (s *DiskFPSet) Contains(fp uint64) bool {
	fp0 := s.checkValid(fp) & diskFPSetFlushedMask
	lockIndex := s.getLockIndex(fp0)
	lock := s.rwLock.GetAt(lockIndex)
	lock.RLock()
	defer lock.RUnlock()
	if s.memLookup(fp0) {
		atomic.AddUint64(&s.memHitCnt, 1)
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

func (s *DiskFPSet) PutBlock(fpv *LongVec) *BitVector {
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

func (s *DiskFPSet) ContainsBlock(fpv *LongVec) *BitVector {
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

func (s *DiskFPSet) Close() {
	s.acquireTblWriteLock()
	defer s.releaseTblWriteLock()
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
	s.publishBRAFReaders()
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
		s.publishBRAFReaders()
	}
}

func (s *DiskFPSet) Exit(cleanup bool) error {
	fpSetBaseExit(s)
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
	// Source releases ownership only after a successful checkpoint copy.
	s.flusherChosen.Store(true)
	s.acquireTblWriteLock()
	if err := s.flushTable(); err != nil {
		return err
	}
	if err := copyFile(s.fpFilename, s.chkptName(fname, "tmp")); err != nil {
		return err
	}
	s.checkPointMark++
	s.releaseTblWriteLock()
	s.flusherChosen.Store(false)
	return nil
}

func (s *DiskFPSet) CommitChkptFile(fname string) error {
	if !s.checkpoint {
		return nil
	}
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException("DiskFPSet.commitChkpt: cannot delete " + oldChkpt)
	}
	return nil
}

func (s *DiskFPSet) Recover() error {
	return nil
}

func (s *DiskFPSet) RecoverFile(fname string) error {
	if !s.checkpoint {
		return nil
	}
	s.acquireTblWriteLock()
	defer s.releaseTblWriteLock()
	return s.recoverFileLocked(s.chkptName(fname, "chkpt"))
}

func (s *DiskFPSet) RecoverTrace(trace *TLCTrace) error {
	if trace == nil {
		return s.Recover()
	}
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

func (s *DiskFPSet) RecoverFP(fp uint64) error {
	fp0 := fp & diskFPSetFlushedMask
	s.acquireTblWriteLock()
	defer s.releaseTblWriteLock()
	if s.memInsert(fp0) {
		if !diskFPSetError2Warning() {
			return NewTLCRuntimeException(ECSystemCheckpointRecoveryCorrupt, "")
		}
		PrintWarning(ECSystemCheckpointRecoveryCorrupt, fmt.Sprintf("Encountered duplicate fingerprint value %d", fp0))
	}
	if s.needsDiskFlush() {
		return s.flushTable()
	}
	return nil
}

func (s *DiskFPSet) CheckFPs() uint64 {
	s.acquireTblWriteLock()
	err := s.flushTable()
	s.releaseTblWriteLock()
	if err != nil {
		panic(err)
	}
	in, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
	if err != nil {
		panic(err)
	}
	defer in.Close()
	length, err := in.Length()
	if err != nil {
		panic(err)
	}
	dis := uint64(math.MaxInt64)
	if length > 0 {
		x, err := in.ReadLong()
		if err != nil {
			panic(err)
		}
		for pos := int64(fpSetLongSize); pos < length; pos += fpSetLongSize {
			y, err := in.ReadLong()
			if err != nil {
				panic(err)
			}
			if difference := y - x; difference >= 0 {
				dis = minUint64(dis, uint64(difference))
			}
			x = y
		}
	}
	if err := in.Close(); err != nil {
		panic(err)
	}
	return dis
}

func (s *DiskFPSet) CheckInvariant(expectFPs ...uint64) bool {
	s.acquireTblWriteLock()
	defer s.releaseTblWriteLock()
	if err := s.flushTable(); err != nil {
		panic(err)
	}
	ok, err := s.checkFile()
	if err != nil {
		panic(err)
	}
	if !ok {
		return false
	}
	return len(expectFPs) == 0 || s.Size() == expectFPs[0]
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
func (s *DiskFPSet) GetTblLoad() int64        { return atomic.LoadInt64(&s.tblLoad) }
func (s *DiskFPSet) GetTblCnt() int64         { return atomic.LoadInt64(&s.tblCnt) }
func (s *DiskFPSet) GetMaxTblCnt() int64      { return s.maxTblCnt }
func (s *DiskFPSet) GetFileCnt() int64        { return atomic.LoadInt64(&s.fileCnt) }
func (s *DiskFPSet) GetDiskLookupCnt() uint64 { return atomic.LoadUint64(&s.diskLookupCnt) }
func (s *DiskFPSet) GetMemHitCnt() uint64     { return atomic.LoadUint64(&s.memHitCnt) }
func (s *DiskFPSet) GetDiskHitCnt() uint64    { return atomic.LoadUint64(&s.diskHitCnt) }
func (s *DiskFPSet) GetDiskWriteCnt() uint64  { return atomic.LoadUint64(&s.diskWriteCnt) }
func (s *DiskFPSet) GetDiskSeekCnt() uint64   { return atomic.LoadUint64(&s.diskSeekCnt) }
func (s *DiskFPSet) GetDiskSeekCache() uint64 { return atomic.LoadUint64(&s.diskSeekCache) }
func (s *DiskFPSet) GetGrowDiskMark() int     { return s.growDiskMark }
func (s *DiskFPSet) GetCheckPointMark() int   { return s.checkPointMark }
func (s *DiskFPSet) GetFlushTime() int64      { return s.flushTime }
func (s *DiskFPSet) ForceFlush()              { s.forceFlush.Store(true) }
func (s *DiskFPSet) GetLockCnt() int          { return s.lockCnt }
func (s *DiskFPSet) GetReaderWriterCnt() int  { return len(s.braf) + len(s.brafPool) }
func (s *DiskFPSet) GetLoadFactor() float64 {
	return float64(atomic.LoadInt64(&s.tblCnt)) / float64(s.maxTblCnt)
}

func (s *DiskFPSet) checkValid(fp uint64) uint64 {
	return fp
}

func (s *DiskFPSet) needsDiskFlush() bool {
	return atomic.LoadInt64(&s.tblCnt) >= s.maxTblCnt || s.forceFlush.Load()
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
		s.mu.Lock()
		s.bucketsCap += diskFPSetInitialBucketCap
		s.mu.Unlock()
		atomic.AddInt64(&s.tblLoad, 1)
		atomic.AddInt64(&s.tblCnt, 1)
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
			s.mu.Lock()
			s.bucketsCap += diskFPSetBucketSizeIncrement
			s.mu.Unlock()
		}
		bucket[j] = fp
	} else {
		if j != bucketLen {
			bucket[j] = bucket[reusable]
		}
		bucket[reusable] = fp
	}
	atomic.AddInt64(&s.tblCnt, 1)
	return false
}

func (s *DiskFPSet) getIndex(fp uint64) int {
	if s.mode == diskFPSetModeMSB {
		return int((uint64(uint32(fp>>32)) & s.mask) >> uint(s.moveBy))
	}
	return int(fp & s.mask)
}

func (s *DiskFPSet) getLockIndex(fp uint64) int {
	if s == nil || s.rwLock == nil || s.lockCnt <= 1 {
		return 0
	}
	if s.mode == diskFPSetModeMSB {
		return int((uint64(uint32(fp>>32)) & uint64(s.lockMask)) >> uint(s.moveBy))
	}
	return int(fp & uint64(s.lockMask))
}

func (s *DiskFPSet) acquireTblWriteLock() {
	if s == nil || s.rwLock == nil {
		return
	}
	s.rwLock.AcquireAllLocks()
}

func (s *DiskFPSet) releaseTblWriteLock() {
	if s == nil || s.rwLock == nil {
		return
	}
	s.rwLock.ReleaseAllLocks()
}

func (s *DiskFPSet) diskLookup(fp uint64) (bool, error) {
	if s.index == nil || len(s.index) == 0 {
		return false, nil
	}
	atomic.AddUint64(&s.diskLookupCnt, 1)
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
			return false, newTLCErrorCode(ECSystemIndexError)
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
	loEntry := int64(loPage) * diskFPSetNumEntriesPerPage
	hiEntry := int64(hiPage) * diskFPSetNumEntriesPerPage
	if loPage == indexLength-2 {
		hiEntry = atomic.LoadInt64(&s.fileCnt) - 1
	}
	if hiPage != loPage+1 {
		return false, newTLCErrorCode(ECSystemIndexError)
	}
	// Source diskLookupBinarySearch selects one reader for the entire search.
	raf, pooled, err := s.openDiskReader()
	if err != nil {
		return false, err
	}
	diskHit := false
	for loEntry < hiEntry {
		midEntry := s.calculateMidEntry(loVal, hiVal, dfp, loEntry, hiEntry)
		if midEntry < loEntry || midEntry >= hiEntry {
			return false, newTLCErrorCode(ECSystemIndexError)
		}
		v, err := s.readDiskFP(raf, midEntry)
		if err != nil {
			return false, err
		}
		if fp < v {
			hiEntry = midEntry
			hiVal = v
		} else if fp > v {
			loEntry = midEntry + 1
			loVal = v
		} else {
			diskHit = true
			break
		}
	}
	// Like Java, return a pooled reader after a successful search. An I/O
	// failure propagates before this step and causes TLC to exit.
	if pooled {
		if err := s.poolClose(raf); err != nil {
			return false, err
		}
	}
	return diskHit, nil
}

func (s *DiskFPSet) calculateMidEntry(loVal uint64, hiVal uint64, dfp float64, loEntry int64, hiEntry int64) int64 {
	dhi, dlo := float64(hiEntry), float64(loEntry)
	dhiVal, dloVal := float64(int64(hiVal)), float64(int64(loVal))
	// Adjacent large fingerprints can round to the same double. Java casts
	// the resulting NaN to zero, rather than Go's implementation-dependent int.
	midEntry := loEntry + javaDoubleToLong((dhi-dlo)*(dfp-dloVal)/(dhiVal-dloVal))
	if midEntry == hiEntry {
		midEntry--
	}
	return midEntry
}

func (s *DiskFPSet) readDiskFP(raf *BufferedRandomAccessFile, entry int64) (uint64, error) {
	seeked, err := raf.Seeek(entry * fpSetLongSize)
	if err != nil {
		return 0, err
	}
	if seeked {
		atomic.AddUint64(&s.diskSeekCnt, 1)
	} else {
		atomic.AddUint64(&s.diskSeekCache, 1)
	}
	value, err := raf.ReadLong()
	if err != nil {
		return 0, err
	}
	return uint64(value), nil
}

func (s *DiskFPSet) flushTable() error {
	if atomic.LoadInt64(&s.tblCnt) == 0 {
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
		// Catch the thrown I/O family, not a runtime exception's nested cause.
		if isJavaIOException(err) {
			return NewIOException("Error: merging entries into file " + s.fpFilename + "  " + javaThrowableString(err))
		}
		return err
	}
	s.mu.Lock()
	atomic.StoreInt64(&s.tblCnt, 0)
	s.bucketsCap = 0
	atomic.StoreInt64(&s.tblLoad, 0)
	s.lsbBuff = nil
	s.mu.Unlock()
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
	cnt := int(int32(atomic.LoadInt64(&s.tblCnt)))
	if cnt <= 0 {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	s.lsbBuff = make([]uint64, cnt)
	// Java allocates cnt slots, retaining unfilled zeros in the sorted buffer.
	idx := 0
	for _, bucket := range s.tbl {
		for k := 0; k < len(bucket) && int64(bucket[k]) > 0; k++ {
			s.lsbBuff[idx] = bucket[k]
			idx++
			bucket[k] |= diskFPSetMarkFlushed
		}
	}
	sort.Slice(s.lsbBuff, func(i, j int) bool { return s.lsbBuff[i] < s.lsbBuff[j] })
}

// mergeNewEntries follows DiskFPSet.Flusher's buffered, streaming merge.
// The caller holds the table write barrier throughout reader replacement.
func (s *DiskFPSet) mergeNewEntries() error {
	readerCnt, poolCnt := len(s.braf), len(s.brafPool)
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
	if err := out.SetLength((atomic.LoadInt64(&s.tblCnt) + atomic.LoadInt64(&s.fileCnt)) * fpSetLongSize); err != nil {
		return err
	}
	if err := s.mergeFingerprintStreams(s.braf[0], out); err != nil {
		return err
	}
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
	ok, err := s.checkFlushedFile()
	if err != nil {
		return err
	}
	if !ok {
		panic(NewAssertionError())
	}
	return nil
}

// DiskFPSet.Flusher checks the reopened file's count, endpoints and signed
// ordering. This differs from the public checkInvariant order-only scan.
func (s *DiskFPSet) checkFlushedFile() (bool, error) {
	raf := s.braf[0]
	length, err := raf.Length()
	if err != nil {
		return false, err
	}
	if length/fpSetLongSize != atomic.LoadInt64(&s.fileCnt) {
		return false, nil
	}
	ptr, err := raf.GetFilePointer()
	if err != nil {
		return false, err
	}
	predecessor := int64(math.MinInt64)
	if length > 0 {
		predecessor, err = raf.ReadLong()
		if err != nil {
			return false, err
		}
		if uint64(predecessor) != s.index[0] {
			return false, nil
		}
		for {
			position, err := raf.GetFilePointer()
			if err != nil {
				return false, err
			}
			if position >= length {
				break
			}
			value, err := raf.ReadLong()
			if err != nil {
				return false, err
			}
			if predecessor >= value {
				return false, nil
			}
			predecessor = value
		}
	}
	if err := raf.Seek(ptr); err != nil {
		return false, err
	}
	return uint64(predecessor) == s.index[len(s.index)-1], nil
}

// MSB walks the sorted table directly; LSB uses the source flusher's sorted
// buffer. Neither variant copies the existing fingerprint file into memory.
func (s *DiskFPSet) mergeFingerprintStreams(in, out *BufferedRandomAccessFile) error {
	buffLen := atomic.LoadInt64(&s.tblCnt)
	var itr *msbDiskIterator
	var maxVal uint64
	if s.mode == diskFPSetModeLSB {
		buffLen = int64(len(s.lsbBuff))
		maxVal = s.lsbBuff[buffLen-1]
	} else {
		itr = newMSBDiskIterator(s.tbl)
		var err error
		maxVal, err = itr.getLast()
		if err != nil {
			return err
		}
	}
	if len(s.index) > 0 && s.index[len(s.index)-1] > maxVal {
		maxVal = s.index[len(s.index)-1]
	}
	indexLen := s.calculateIndexLen(buffLen)
	s.index = make([]uint64, indexLen)
	s.index[indexLen-1] = maxVal
	currIndex, counter := 0, 0
	writeFP := func(fp uint64) error {
		if err := out.WriteLong(int64(fp)); err != nil {
			return err
		}
		atomic.AddUint64(&s.diskWriteCnt, 1)
		if counter == 0 {
			s.index[currIndex] = fp
			currIndex++
			counter = diskFPSetNumEntriesPerPage
		}
		counter--
		return nil
	}
	var value uint64
	eof := atomic.LoadInt64(&s.fileCnt) == 0
	readOld := func() error {
		v, err := in.ReadLong()
		var end *EOFException
		if errors.As(err, &end) {
			eof = true
			return nil
		}
		if err != nil {
			return err
		}
		value = uint64(v)
		return nil
	}
	if !eof {
		if err := readOld(); err != nil {
			return err
		}
	}
	if itr != nil {
		fp, err := itr.next()
		if err != nil {
			return err
		}
		eol := false
		for !eof || !eol {
			if (value < fp || eol) && !eof {
				if err := writeFP(value); err != nil {
					return err
				}
				if err := readOld(); err != nil {
					return err
				}
			} else {
				if value == fp {
					return NewTLCRuntimeException(ECTLCFPValueAlreadyOnDisk, fmt.Sprint(value))
				}
				if err := writeFP(fp); err != nil {
					return err
				}
				next, err := itr.next()
				var end *NoSuchElementException
				if errors.As(err, &end) {
					if itr.hasNext() || itr.reads() != buffLen {
						return NewTLCRuntimeException(ECGeneral)
					}
					eol = true
				} else if err != nil {
					return err
				} else {
					fp = next
				}
			}
		}
	} else {
		i := 0
		for !eof && i < len(s.lsbBuff) {
			if value < s.lsbBuff[i] {
				if err := writeFP(value); err != nil {
					return err
				}
				if err := readOld(); err != nil {
					return err
				}
			} else {
				if value == s.lsbBuff[i] {
					return NewTLCRuntimeException(ECTLCFPValueAlreadyOnDisk, fmt.Sprint(value))
				}
				if err := writeFP(s.lsbBuff[i]); err != nil {
					return err
				}
				i++
			}
		}
		if eof {
			for ; i < len(s.lsbBuff); i++ {
				if err := writeFP(s.lsbBuff[i]); err != nil {
					return err
				}
			}
		} else {
			for !eof {
				if err := writeFP(value); err != nil {
					return err
				}
				if err := readOld(); err != nil {
					return err
				}
			}
		}
	}
	if currIndex != indexLen-1 {
		return NewTLCRuntimeException(ECSystemIndexError)
	}
	atomic.AddInt64(&s.fileCnt, buffLen)
	return nil
}

func (s *DiskFPSet) calculateIndexLen(buffLen int64) int {
	indexLen := ((atomic.LoadInt64(&s.fileCnt) + buffLen - 1) / diskFPSetNumEntriesPerPage) + 2
	if indexLen <= 0 {
		indexLen = 2
	}
	return int(indexLen)
}

// recoverFileLocked ports DiskFPSet.recover(String), including write statistics
// and its index assertions. Recovery has exclusive access to the set.
func (s *DiskFPSet) recoverFileLocked(path string) error {
	checkpoint, err := NewBufferedRandomAccessFile(path, "r")
	if err != nil {
		return err
	}
	defer checkpoint.Close()
	current, err := NewBufferedRandomAccessFile(s.fpFilename, "rw")
	if err != nil {
		return err
	}
	defer current.Close()
	length, err := checkpoint.Length()
	if err != nil {
		return err
	}
	fileCnt := length / fpSetLongSize
	atomic.StoreInt64(&s.fileCnt, fileCnt)
	indexLen := int(int32((fileCnt-1)/diskFPSetNumEntriesPerPage) + 2)
	if indexLen < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(indexLen)))
	}
	s.index = make([]uint64, indexLen)
	currIndex, counter := 0, 0
	var fp int64
	predecessor := int64(math.MinInt64)
	for {
		next, err := checkpoint.ReadLong()
		var end *EOFException
		if errors.As(err, &end) {
			if currIndex != indexLen-1 {
				return NewTLCRuntimeException(ECSystemIndexError)
			}
			s.index[indexLen-1] = uint64(fp)
			break
		}
		if err != nil {
			return err
		}
		fp = next
		if err := current.WriteLong(fp); err != nil {
			return err
		}
		atomic.AddUint64(&s.diskWriteCnt, 1)
		if counter == 0 {
			s.index[currIndex] = uint64(fp)
			currIndex++
			counter = diskFPSetNumEntriesPerPage
		}
		counter--
		if predecessor >= fp {
			return NewTLCRuntimeException(ECSystemIndexError)
		}
		predecessor = fp
	}
	if err := checkpoint.Close(); err != nil {
		return err
	}
	if err := current.Close(); err != nil {
		return err
	}
	return s.reopenBRAFReaders()
}

func (s *DiskFPSet) openBRAFReaders(numReaders int, poolSize int) error {
	if numReaders < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(numReaders)))
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
	s.publishBRAFReaders()
	return nil
}

func (s *DiskFPSet) publishBRAFReaders() {
	readers := append([]*BufferedRandomAccessFile(nil), s.braf...)
	s.readers.Store(&readers)
}

func (s *DiskFPSet) closeBRAFReaders() error {
	s.readers.Store(nil)
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
	// Recovery holds the table write lock. Retain each source slot until its
	// replacement opens successfully, including partial mutation on failure.
	defer s.publishBRAFReaders()
	for _, readers := range [][]*BufferedRandomAccessFile{s.braf, s.brafPool} {
		for i, old := range readers {
			if err := old.Close(); err != nil {
				return err
			}
			next, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
			if err != nil {
				return err
			}
			readers[i] = next
		}
	}
	s.poolIndex = 0
	return nil
}

func (s *DiskFPSet) openDiskReader() (*BufferedRandomAccessFile, bool, error) {
	var readers []*BufferedRandomAccessFile
	if snapshot := s.readers.Load(); snapshot != nil {
		readers = *snapshot
	}
	id := CurrentThreadIDOr(len(readers))
	if id >= 0 && id < len(readers) && readers[id] != nil {
		return readers[id], false, nil
	}
	s.poolMu.Lock()
	defer s.poolMu.Unlock()
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

func (s *DiskFPSet) poolClose(raf *BufferedRandomAccessFile) error {
	if raf == nil {
		return nil
	}
	s.poolMu.Lock()
	defer s.poolMu.Unlock()
	if len(s.brafPool) > 0 && s.poolIndex > 0 {
		s.poolIndex--
		s.brafPool[s.poolIndex] = raf
		return nil
	}
	return raf.Close()
}

func (s *DiskFPSet) clearTable() {
	for i := range s.tbl {
		s.tbl[i] = nil
	}
	atomic.StoreInt64(&s.tblCnt, 0)
	atomic.StoreInt64(&s.tblLoad, 0)
	s.bucketsCap = 0
	s.lsbBuff = nil
}

// checkFile performs DiskFPSet.checkInvariant's signed, sequential order scan.
// The expected-size overload compares Size(), not the backing file length.
func (s *DiskFPSet) checkFile() (ok bool, err error) {
	in, err := NewBufferedRandomAccessFile(s.fpFilename, "r")
	if err != nil {
		return false, err
	}
	// Source finally propagates close failures, including after a false result.
	defer func() {
		if closeErr := in.Close(); closeErr != nil {
			err = closeErr
		}
	}()
	length, err := in.Length()
	if err != nil {
		return false, err
	}
	predecessor := int64(math.MinInt64)
	for pos := int64(0); pos < length; pos += fpSetLongSize {
		value, err := in.ReadLong()
		if err != nil {
			return false, err
		}
		if predecessor >= value {
			return false, nil
		}
		predecessor = value
	}
	return true, nil
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
		return 0, NewNoSuchElementException()
	}
	if i.havePrevious && i.previous >= result {
		return 0, NewTLCRuntimeException(ECGeneral)
	}
	i.previous = result
	i.havePrevious = true
	i.readElements++
	return result, nil
}

// getLast scans the unflushed positive entries from the last bucket backward,
// matching MSBDiskFPSet.TLCIterator.getLast independently of the read cursor.
func (i *msbDiskIterator) getLast() (uint64, error) {
	for first := len(i.buff) - 1; first >= 0; first-- {
		bucket := i.buff[first]
		for second := len(bucket) - 1; second >= 0; second-- {
			if int64(bucket[second]) > 0 {
				return bucket[second], nil
			}
		}
	}
	return 0, NewNoSuchElementException()
}

func (i *msbDiskIterator) reads() int64 {
	return i.readElements
}

// FileUtil.copyFile uses Files.copy(REPLACE_EXISTING), with no parent-directory
// creation. Keep the source open while replacing the destination, preserve the
// same-file no-op, and replace destination links rather than following them.
func copyFile(src string, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	defer func() {
		if closeErr := in.Close(); err == nil && closeErr != nil {
			err = bufferedRandomAccessFileIOError(closeErr)
		}
	}()
	info, err := in.Stat()
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	if target, statErr := os.Lstat(dst); statErr == nil {
		if os.SameFile(info, target) {
			return nil
		}
		if err := os.Remove(dst); err != nil {
			return bufferedRandomAccessFileIOError(err)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return bufferedRandomAccessFileIOError(statErr)
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return bufferedRandomAccessFileIOError(copyErr)
	}
	return bufferedRandomAccessFileIOError(closeErr)
}

// FileUtil.replaceFile delegates to Files.move(REPLACE_EXISTING). The native
// rename replaces an existing file without a separate delete that could lose
// the live file when the move fails.
func replaceFile(src string, dst string) error {
	return bufferedRandomAccessFileIOError(os.Rename(src, dst))
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

func bitsLeadingZeros32(x uint32) int {
	if x == 0 {
		return 32
	}
	n := 0
	for bit := uint32(1) << 31; bit != 0 && x&bit == 0; bit >>= 1 {
		n++
	}
	return n
}

// Original DiskFPSet auxiliary-storage requirement and LSB/MSB overrides.
func (s *DiskFPSet) GetAuxiliaryStorageRequirement() float64 {
	switch s.mode {
	case diskFPSetModeLSB:
		return 2.5
	case diskFPSetModeMSB:
		return 1.5
	default:
		return 1
	}
}
