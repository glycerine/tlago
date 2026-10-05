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
	array      *LongArray
	indexer    *OffHeapIndexer
	numThreads int
	probeLimit int
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
	return &NonCheckpointableDiskFPSet{DiskFPSet: newHeapDiskFPSet(config, diskFPSetModeMSB, false)}
}

func NewOffHeapDiskFPSet(config *FPSetConfiguration) *OffHeapDiskFPSet {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	positions := config.GetMemoryInFingerprintCnt()
	if positions <= 0 {
		positions = diskFPSetDefaultMaxTblCnt
	}
	base := NewNonCheckpointableDiskFPSet(config)
	base.DiskFPSet.maxTblCnt = positions
	base.DiskFPSet.tbl = nil
	base.DiskFPSet.capacity = 0
	base.DiskFPSet.mask = 0
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
	s.fileCnt = 0
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

func (s *DiskFPSet) Size() uint64 {
	s.acquireTblWriteLock()
	defer s.releaseTblWriteLock()
	return uint64(s.tblCnt + s.fileCnt)
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
		insertions := s.tblCnt
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
	s.flusherChosen.Store(true)
	s.acquireTblWriteLock()
	defer func() {
		s.releaseTblWriteLock()
		s.flusherChosen.Store(false)
	}()
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
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return fmt.Errorf("DiskFPSet.commitChkpt: cannot delete %s", oldChkpt)
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
	defer s.releaseTblWriteLock()
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
	s.acquireTblWriteLock()
	defer s.releaseTblWriteLock()
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
	return float64(s.tblCnt) / float64(s.maxTblCnt)
}

func (s *DiskFPSet) checkValid(fp uint64) uint64 {
	return fp
}

func (s *DiskFPSet) needsDiskFlush() bool {
	s.mu.Lock()
	tblCnt := s.tblCnt
	s.mu.Unlock()
	return tblCnt >= s.maxTblCnt || s.forceFlush.Load()
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
		s.tblLoad++
		s.tblCnt++
		s.mu.Unlock()
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
	s.mu.Lock()
	s.tblCnt++
	s.mu.Unlock()
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
		hiEntry = s.fileCnt - 1
	}
	for loEntry < hiEntry {
		midEntry := s.calculateMidEntry(loVal, hiVal, dfp, loEntry, hiEntry)
		if midEntry < loEntry || midEntry >= hiEntry {
			return false, newTLCErrorCode(ECSystemIndexError)
		}
		v, err := s.readDiskFP(midEntry)
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
			return true, nil
		}
	}
	return false, nil
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
	s.mu.Lock()
	s.tblCnt = 0
	s.bucketsCap = 0
	s.tblLoad = 0
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
	cnt := int(int32(s.tblCnt))
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
		atomic.AddUint64(&s.diskWriteCnt, 1)
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
			return newTLCErrorCode(ECTLCFPValueAlreadyOnDisk, fmt.Sprint(oldValues[i]))
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
	if poolCnt <= 0 {
		poolCnt = diskFPSetBRAFPoolSize
	}
	if err := s.closeBRAFReaders(); err != nil {
		return err
	}
	return s.openBRAFReaders(readerCnt, poolCnt)
}

func (s *DiskFPSet) openDiskReader() (*BufferedRandomAccessFile, bool, error) {
	s.poolMu.Lock()
	defer s.poolMu.Unlock()
	id := CurrentThreadIDOr(len(s.braf))
	if id >= 0 && id < len(s.braf) && s.braf[id] != nil {
		return s.braf[id], false, nil
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
	s.poolMu.Lock()
	defer s.poolMu.Unlock()
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
