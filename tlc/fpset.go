package tlc

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	fpSetLongSize       = 8
	multiFPSetMaxFPBits = 30
	multiFPSetMinFPBits = 0
	FPSetImplProperty   = "tlc2.tool.fp.FPSet.impl"
)

// Java's FPSet.statesSeen is a plain long updated with a separate
// read/modify/write. Atomic snapshots avoid a Go data race without turning the
// source operation into an atomic addition, so concurrent updates may still be
// lost in the same way.
func fpSetRecordStatesSeen(statesSeen *uint64, count int) {
	atomic.StoreUint64(statesSeen, atomic.LoadUint64(statesSeen)+uint64(count))
}

func fpSetAddPlainLong(counter *int64, delta int64) {
	atomic.StoreInt64(counter, atomic.LoadInt64(counter)+delta)
}

func fpSetIncrementPlainInt(counter *int32) {
	atomic.StoreInt32(counter, atomic.LoadInt32(counter)+1)
}

const (
	tlcRuntimeMinFPMemSize        = int64(20 * (1 << 19))
	tlcRuntimeDefaultNonHeapBytes = int64(64 * 1024 * 1024)
)

type FPSetConfiguration struct {
	// GetMemoryInBytesOverride preserves subclass dispatch of Java's virtual
	// getMemoryInBytes, including calls from getMemoryInFingerprintCnt.
	GetMemoryInBytesOverride func() int64
	FPBits                   int
	MemoryInBytes            int64
	Ratio                    float64
	Implementation           string
	NoNesting                bool
	MemoryDivisor            int64
}

func NewFPSetConfiguration() *FPSetConfiguration {
	return NewFPSetConfigurationWithRatio(0.25)
}

func NewFPSetConfigurationWithRatio(ratio float64) *FPSetConfiguration {
	return NewFPSetConfigurationWithRatioAndImplementation(ratio, fpSetImplementationFromEnv())
}

func NewFPSetConfigurationWithRatioAndImplementation(ratio float64, implementation string) *FPSetConfiguration {
	return &FPSetConfiguration{
		FPBits:         1,
		MemoryInBytes:  -1,
		Ratio:          ratio,
		Implementation: implementation,
	}
}

func (c *FPSetConfiguration) require() *FPSetConfiguration {
	if c == nil {
		panic(NewNullPointerException())
	}
	return c
}

func (c *FPSetConfiguration) AllowsNesting() bool {
	c.require()
	return !c.NoNesting && c.GetFpBits() > 0
}

func (c *FPSetConfiguration) GetFpBits() int {
	c.require()
	if c.FPBits == 0 && isDiskFPSetImplementation(c.Implementation) {
		c.FPBits = 1
	}
	return c.FPBits
}

func (c *FPSetConfiguration) GetFPBits() int {
	return c.GetFpBits()
}

func (c *FPSetConfiguration) SetFpBits(fpBits int) {
	c.require()
	fpBits = int(int32(fpBits))
	if !IsValid(fpBits) {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	c.FPBits = fpBits
}

func (c *FPSetConfiguration) SetFPBits(fpBits int) {
	c.SetFpBits(fpBits)
}

func (c *FPSetConfiguration) GetMemoryInBytes() int64 {
	c.require()
	if c.GetMemoryInBytesOverride != nil {
		return c.GetMemoryInBytesOverride()
	}
	divisor := c.MemoryDivisor
	if divisor <= 0 {
		divisor = 1
	}
	var memory int64
	if AllocatesOnHeap(c.Implementation) {
		requested := c.Ratio
		if c.MemoryInBytes > 0 {
			requested = float64(c.MemoryInBytes) * c.Ratio
		}
		memory = tlcRuntimeFPMemSize(requested)
	} else {
		memory = tlcRuntimeNonHeapPhysicalMemory()
	}
	if divisor > 1 {
		return memory / divisor
	}
	return memory
}

func (c *FPSetConfiguration) GetMemoryInFingerprintCnt() int64 {
	c.require()
	return c.GetMemoryInBytes() / fpSetLongSize
}

func (c *FPSetConfiguration) GetMultiFPSetCnt() int {
	c.require()
	return 1 << c.GetFpBits()
}

func (c *FPSetConfiguration) SetRatio(ratio float64) {
	c.require()
	if !(ratio >= 0 && ratio <= 1) {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	c.Ratio = ratio
}

func (c *FPSetConfiguration) GetRatio() float64 {
	c.require()
	return c.Ratio
}

func (c *FPSetConfiguration) SetMemory(memory int64) {
	c.require()
	if memory < 0 {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	c.MemoryInBytes = memory
}

func (c *FPSetConfiguration) GetImplementation() string {
	c.require()
	return c.Implementation
}

func (c *FPSetConfiguration) SetImplementation(implementation string) {
	c.require()
	c.Implementation = implementation
}

func NewMultiFPSetConfiguration(config *FPSetConfiguration) *FPSetConfiguration {
	config.require()
	child := *config
	child.NoNesting = true
	child.MemoryDivisor = int64(config.GetMultiFPSetCnt())
	if child.GetMemoryInFingerprintCnt() <= 0 {
		panic(NewIllegalArgumentException("Given fpSetConfig results in zero or negative fp count."))
	}
	return &child
}

func tlcRuntimeFPMemSize(fpMemSize float64) int64 {
	maxMemory := tlcRuntimeMaxHeapMemoryBytes()
	if fpMemSize == -1 {
		fpMemSize = float64(maxMemory >> 2)
	}
	if 0 <= fpMemSize && fpMemSize <= 1 {
		fpMemSize = float64(maxMemory) * fpMemSize
	}
	if fpMemSize < float64(tlcRuntimeMinFPMemSize) {
		fpMemSize = float64(tlcRuntimeMinFPMemSize)
	}
	if fpMemSize >= float64(maxMemory) {
		fpMemSize = float64(maxMemory - (maxMemory >> 2))
	}
	return javaDoubleToLong(fpMemSize)
}

var tlcDefaultHeapBudget = sync.OnceValue(func() int64 {
	memory, err := tlcPhysicalMemoryBytes()
	if err != nil {
		panic(NewIllegalStateException("Cannot determine TLC memory budget: " + err.Error() + "; configure GOMEMLIMIT"))
	}
	// Match the source VM's default maximum-heap fraction. The native Go limit,
	// when configured, takes precedence; committed/reserved memory is not a limit.
	budget := memory / 4
	// Apply the default through the native runtime as well: returning a budget
	// while leaving the GC unlimited would not model Runtime.maxMemory.
	// Preserve a limit configured before initialization.
	if configured := debug.SetMemoryLimit(-1); configured != math.MaxInt64 {
		return configured
	}
	debug.SetMemoryLimit(budget)
	return budget
})

func tlcRuntimeMaxHeapMemoryBytes() int64 {
	limit := debug.SetMemoryLimit(-1)
	if limit != math.MaxInt64 {
		return limit
	}
	return tlcDefaultHeapBudget()
}

// TLAGO_MAX_DIRECT_MEMORY supplies the native equivalent of the Java VM's
// -XX:MaxDirectMemorySize argument. TLCRuntime accepts bytes or k/m/g suffixes;
// absent an explicit limit, preserve its 64 MiB default.
func tlcRuntimeNonHeapPhysicalMemory() int64 {
	value, ok := os.LookupEnv("TLAGO_MAX_DIRECT_MEMORY")
	if !ok {
		return tlcRuntimeDefaultNonHeapBytes
	}
	value = strings.ToLower(value)
	shift := uint(0)
	if len(value) > 0 {
		switch value[len(value)-1] {
		case 'k':
			shift = 10
		case 'm':
			shift = 20
		case 'g':
			shift = 30
		}
		if shift != 0 {
			value = value[:len(value)-1]
		}
	}
	memory, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(NewNumberFormatException(value))
	}
	return memory << shift
}

func fpSetImplementationFromEnv() string {
	if value, ok := tlcLookupSystemProperty(FPSetImplProperty); ok {
		return value
	}
	if value, ok := os.LookupEnv("TLAGO_FPSET_IMPL"); ok {
		return value
	}
	return GetImplementationDefault()
}

func GetImplementations() []string {
	return []string{
		"tlc2.tool.fp.MSBDiskFPSet",
		"tlc2.tool.fp.LSBDiskFPSet",
		"tlc2.tool.fp.OffHeapDiskFPSet",
	}
}

func GetFPSetImplementations() []string {
	return GetImplementations()
}

func GetImplementationDefault() string {
	return "tlc2.tool.fp.MSBDiskFPSet"
}

func GetFPSetImplementationDefault() string {
	return GetImplementationDefault()
}

func AllocatesOnHeap(implementation string) bool {
	switch implementation {
	case "tlc2.tool.fp.OffHeapDiskFPSet":
		return false
	case "tlc2.tool.fp.DiskFPSet",
		"tlc2.tool.fp.HeapBasedDiskFPSet",
		"tlc2.tool.fp.LSBDiskFPSet",
		"tlc2.tool.fp.MSBDiskFPSet",
		"tlc2.tool.fp.NonCheckpointableDiskFPSet",
		"tlc2.tool.fp.MemFPSet",
		"tlc2.tool.fp.MemFPSet1",
		"tlc2.tool.fp.MemFPSet2",
		"tlc2.tool.fp.NoopFPSet":
		return true
	default:
		return false
	}
}

func FPSetAllocatesOnHeap(implementation string) bool {
	return AllocatesOnHeap(implementation)
}

func GetVMArguments(implementation string, memoryMiB int64) string {
	if AllocatesOnHeap(implementation) {
		return fmt.Sprintf("-Xmx%dm", memoryMiB)
	}
	return fmt.Sprintf("-XX:MaxDirectMemorySize=%dm", memoryMiB)
}

func GetFPSetVMArguments(implementation string, memoryMiB int64) string {
	return GetVMArguments(implementation, memoryMiB)
}

func IsValid(fpBits int) bool {
	fpBits = int(int32(fpBits))
	return fpBits >= multiFPSetMinFPBits && fpBits <= multiFPSetMaxFPBits
}

func IsValidFPBits(fpBits int) bool {
	return IsValid(fpBits)
}

func isDiskFPSetImplementation(implementation string) bool {
	switch implementation {
	case "tlc2.tool.fp.DiskFPSet",
		"tlc2.tool.fp.HeapBasedDiskFPSet",
		"tlc2.tool.fp.LSBDiskFPSet",
		"tlc2.tool.fp.MSBDiskFPSet",
		"tlc2.tool.fp.OffHeapDiskFPSet",
		"tlc2.tool.fp.NonCheckpointableDiskFPSet":
		return true
	default:
		return false
	}
}

// FPSetTrace preserves the virtual trace enumeration used by FPSet.recover:
// disk-backed sets enumerate either the standalone or concurrent trace, while
// memory-backed sets recover their own snapshots without reading the trace.
type FPSetTrace interface {
	Elements() (*TLCTraceEnumerator, error)
}

type FPSet interface {
	Init(numThreads int, metadir string, filename string) FPSet
	Size() uint64
	Sizeof() uint64
	Put(fp uint64) bool
	Contains(fp uint64) bool
	PutBlock(fpv *LongVec) *BitVector
	ContainsBlock(fpv *LongVec) *BitVector
	GetStatesSeen() uint64
	GetConfiguration() *FPSetConfiguration
	Close()
	AddThread() error
	IncWorkers(num int)
	Exit(cleanup bool) error
	CheckInvariant(expectFPs ...uint64) bool
	CheckFPs() uint64
	BeginChkpt() error
	BeginChkptFile(fname string) error
	CommitChkpt() error
	CommitChkptFile(fname string) error
	Recover(trace FPSetTrace) error
	RecoverFile(fname string) error
	RecoverTrace(trace FPSetTrace) error
	RecoverFP(fp uint64) error
	UnexportObject(force bool)
}

func requireInvariantExpectedCount(expected []uint64) {
	if len(expected) > 1 {
		panic(NewIllegalArgumentException("CheckInvariant accepts at most one expected fingerprint count"))
	}
}

type fpSetFutureResult struct {
	set FPSet
	err error
}

// FPSetFuture represents the Future<FPSet> returned by
// FPSetFactory.getFPSetInitialized. Closing done publishes one immutable result
// so repeated Get calls observe the same set or ExecutionException.
type FPSetFuture struct {
	done   chan struct{}
	result fpSetFutureResult
}

func GetFPSetInitialized(config *FPSetConfiguration, metadir string, mainFile string) *FPSetFuture {
	future := &FPSetFuture{done: make(chan struct{})}
	go func() {
		defer close(future.done)
		defer func() {
			if failure := recover(); failure != nil {
				future.result.err = panicValueAsError(failure)
			}
		}()
		set := GetFPSet(config)
		if isNil(set) {
			panic(NewNullPointerException())
		}
		future.result.set = set.Init(NumWorkers(), metadir, mainFile)
	}()
	return future
}

func (f *FPSetFuture) Get() (FPSet, error) {
	if f == nil {
		panic(NewNullPointerException())
	}
	<-f.done
	if f.result.err != nil {
		return nil, NewExecutionException(f.result.err)
	}
	return f.result.set, nil
}

func (f *FPSetFuture) IsDone() bool {
	if f == nil {
		panic(NewNullPointerException())
	}
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

func GetFPSet(configs ...*FPSetConfiguration) FPSet {
	if len(configs) > 1 {
		panic(NewIllegalArgumentException("GetFPSet accepts at most one configuration"))
	}
	var config *FPSetConfiguration
	if len(configs) == 0 {
		config = NewFPSetConfiguration()
	} else {
		config = configs[0]
		if config == nil {
			panic(NewNullPointerException())
		}
	}
	if config.AllowsNesting() {
		return NewMultiFPSet(config)
	}
	implementation := config.GetImplementation()
	if !fpSetSupportsArchitecture(implementation) {
		PrintWarning(ECTLCFeatureUnsupported, fmt.Sprintf(
			"Selected fingerprint set (set of visited states) %s does not support current architecture %s. "+
				"Reverting to default fingerprint set. "+
				"Off-heap memory allocated via -XX:MaxDirectMemorySize flag cannot be used by default "+
				"fingerprint set and is therefore wasted.",
			implementation, fpSetArchitecture()))
		return NewMSBDiskFPSet(config)
	}
	if set := loadFPSetImplementation(implementation, config); set != nil {
		return set
	}
	PrintWarning(ECGeneral, "unsuccessfully trying to load custom FPSet class: "+implementation)
	return nil
}

func NewFPSet(config *FPSetConfiguration) FPSet {
	if config == nil {
		return GetFPSet()
	}
	return GetFPSet(config)
}

func fpSetSupportsArchitecture(implementation string) bool {
	switch implementation {
	case "tlc2.tool.fp.FPSet",
		"tlc2.tool.fp.DiskFPSet",
		"tlc2.tool.fp.HeapBasedDiskFPSet",
		"tlc2.tool.fp.LSBDiskFPSet",
		"tlc2.tool.fp.MSBDiskFPSet",
		"tlc2.tool.fp.OffHeapDiskFPSet",
		"tlc2.tool.fp.NonCheckpointableDiskFPSet",
		"tlc2.tool.fp.MemFPSet",
		"tlc2.tool.fp.MemFPSet1",
		"tlc2.tool.fp.MemFPSet2",
		"tlc2.tool.fp.MultiFPSet",
		"tlc2.tool.fp.NoopFPSet":
		return implementation != "tlc2.tool.fp.OffHeapDiskFPSet" || OffHeapDiskFPSetIsSupported()
	default:
		return false
	}
}

func fpSetArchitecture() string {
	if strconv.IntSize == 32 {
		return "BIT_32"
	}
	return "BIT_64"
}

func loadFPSetImplementation(implementation string, config *FPSetConfiguration) FPSet {
	switch implementation {
	case "tlc2.tool.fp.MemFPSet":
		return NewMemFPSetWithConfig(config)
	case "tlc2.tool.fp.MemFPSet1":
		return NewMemFPSet1(config)
	case "tlc2.tool.fp.MemFPSet2":
		return NewMemFPSet2(config)
	case "tlc2.tool.fp.NoopFPSet":
		return NewNoopFPSet(config)
	case "tlc2.tool.fp.LSBDiskFPSet":
		return NewLSBDiskFPSet(config)
	case "tlc2.tool.fp.MSBDiskFPSet":
		return NewMSBDiskFPSet(config)
	case "tlc2.tool.fp.OffHeapDiskFPSet":
		return NewOffHeapDiskFPSet(config)
	case "tlc2.tool.fp.MultiFPSet":
		return NewMultiFPSet(config)
	}
	return nil
}

func fpSetInitialized(set FPSet) bool {
	switch s := set.(type) {
	case nil:
		return false
	case *MemFPSet:
		return s.metadir != "" && s.filename != ""
	case *MemFPSet1:
		return s.metadir != "" && s.filename != ""
	case *MemFPSet2:
		return s.metadir != "" && s.filename != ""
	case *DiskFPSet:
		return s.metadir != "" && s.filename != ""
	case *HeapBasedDiskFPSet:
		return s.DiskFPSet != nil && s.DiskFPSet.metadir != "" && s.DiskFPSet.filename != ""
	case *LSBDiskFPSet:
		return s.DiskFPSet != nil && s.DiskFPSet.metadir != "" && s.DiskFPSet.filename != ""
	case *MSBDiskFPSet:
		return s.DiskFPSet != nil && s.DiskFPSet.metadir != "" && s.DiskFPSet.filename != ""
	case *NonCheckpointableDiskFPSet:
		return s.DiskFPSet != nil && s.DiskFPSet.metadir != "" && s.DiskFPSet.filename != ""
	case *OffHeapDiskFPSet:
		return s.DiskFPSet != nil && s.DiskFPSet.metadir != "" && s.DiskFPSet.filename != ""
	case *MultiFPSet:
		return s.metadir != "" && s.filename != ""
	case *NoopFPSet:
		return true
	default:
		return true
	}
}

type MemFPSet struct {
	fpSetLifecycle
	mu         distributedServerMonitor
	metadir    string
	filename   string
	table      [][]uint64
	count      uint64
	threshold  uint64
	mask       uint64
	statesSeen uint64
	config     *FPSetConfiguration
}

const (
	memFPSetMaxLoad            = 20
	memFPSetLogInitialCapacity = 16
)

func NewMemFPSet() *MemFPSet {
	return NewMemFPSetWithConfig(NewFPSetConfiguration())
}

func NewMemFPSetWithConfig(config *FPSetConfiguration) *MemFPSet {
	initialCapacity := 1 << memFPSetLogInitialCapacity
	return &MemFPSet{
		table:     make([][]uint64, initialCapacity),
		threshold: uint64(initialCapacity * memFPSetMaxLoad),
		mask:      uint64(initialCapacity - 1),
		config:    config,
	}
}

func NewMemFPSetUnchecked() *MemFPSet {
	return NewMemFPSet()
}

func (s *MemFPSet) requireMemFPSet() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *MemFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.requireMemFPSet()
	s.metadir = metadir
	s.filename = filename
	return s
}

func (s *MemFPSet) Size() uint64 {
	s.requireMemFPSet()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

func (s *MemFPSet) Sizeof() uint64 {
	s.requireMemFPSet()
	s.mu.Lock()
	defer s.mu.Unlock()

	size := uint64(28)
	size += 16 + uint64(len(s.table))*8
	for _, bucket := range s.table {
		if bucket != nil {
			size += 16 + uint64(len(bucket))*8
		}
	}
	return size
}

func (s *MemFPSet) Put(fp uint64) bool {
	s.requireMemFPSet()
	s.mu.Lock()
	defer s.mu.Unlock()

	index := fp & s.mask
	bucket := s.table[index]
	for _, existing := range bucket {
		if existing == fp {
			return true
		}
	}

	if s.count >= s.threshold {
		s.rehash()
		index = fp & s.mask
		bucket = s.table[index]
	}

	s.table[index] = append(bucket, fp)
	s.count++
	return false
}

func (s *MemFPSet) PutBlock(fpv *LongVec) *BitVector {
	s.requireMemFPSet()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	size := fpv.Size()
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Put(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet) ContainsBlock(fpv *LongVec) *BitVector {
	s.requireMemFPSet()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	size := fpv.Size()
	fpSetRecordStatesSeen(&s.statesSeen, size)

	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet) GetStatesSeen() uint64 {
	s.requireMemFPSet()
	return atomic.LoadUint64(&s.statesSeen)
}

func (s *MemFPSet) GetConfiguration() *FPSetConfiguration {
	s.requireMemFPSet()
	return s.config
}

func (s *MemFPSet) Close() { s.requireMemFPSet() }

func (s *MemFPSet) AddThread() error {
	s.requireMemFPSet()
	return nil
}

func (s *MemFPSet) IncWorkers(num int) { s.requireMemFPSet() }

func (s *MemFPSet) Exit(cleanup bool) error {
	s.requireMemFPSet()
	fpSetBaseExit(s)
	return completeFingerprintExit(s.metadir, cleanup)
}

func (s *MemFPSet) CheckInvariant(expectFPs ...uint64) bool {
	s.requireMemFPSet()
	requireInvariantExpectedCount(expectFPs)
	return true
}

func (s *MemFPSet) Contains(fp uint64) bool {
	s.requireMemFPSet()
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket := s.table[fp&s.mask]
	for _, existing := range bucket {
		if existing == fp {
			return true
		}
	}
	return false
}

func (s *MemFPSet) rehash() {
	old := s.table
	oldCapacity := len(old)
	newTable := make([][]uint64, oldCapacity*2)
	oneBitMask := uint64(oldCapacity)

	for i, bucket := range old {
		if len(bucket) == 0 {
			continue
		}
		cnt0, cnt1 := 0, 0
		for _, fp := range bucket {
			if fp&oneBitMask == 0 {
				cnt0++
			} else {
				cnt1++
			}
		}
		if cnt0 == 0 {
			newTable[i+oldCapacity] = bucket
		} else if cnt1 == 0 {
			newTable[i] = bucket
		} else {
			left := make([]uint64, cnt0)
			right := make([]uint64, cnt1)
			for _, fp := range bucket {
				if fp&oneBitMask == 0 {
					cnt0--
					left[cnt0] = fp
				} else {
					cnt1--
					right[cnt1] = fp
				}
			}
			newTable[i] = left
			newTable[i+oldCapacity] = right
		}
	}

	s.threshold *= 2
	s.table = newTable
	s.mask = uint64(len(newTable) - 1)
}

func (s *MemFPSet) CheckFPs() uint64 {
	s.requireMemFPSet()
	dis := int64(1<<63 - 1)
	for i, bucket := range s.table {
		for j, x := range bucket {
			for k := j + 1; k < len(bucket); k++ {
				dis = javaLongMin(dis, javaLongAbs(javaLongSub(int64(x), int64(bucket[k]))))
			}
			for _, otherBucket := range s.table[i+1:] {
				for _, y := range otherBucket {
					dis1 := javaLongDistanceIfNonnegative(int64(x), int64(y))
					if dis1 >= 0 {
						dis = javaLongMin(dis, dis1)
					}
				}
			}
		}
	}
	return uint64(dis)
}

func (s *MemFPSet) BeginChkpt() error {
	s.requireMemFPSet()
	return s.BeginChkptFile(s.filename)
}

func (s *MemFPSet) BeginChkptFile(fname string) (err error) {
	s.requireMemFPSet()
	path := s.chkptName(fname, "tmp")
	file, err := openFileOutputStream(path)
	if err != nil {
		return err
	}
	// Preserve source buffering, but close the owned file directly on failure:
	// closing the buffer would flush again and replay a failed write.
	out := NewBufferedDataOutputStream(file)
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = bufferedRandomAccessFileIOError(closeErr)
		}
	}()

	for _, bucket := range s.table {
		for _, fp := range bucket {
			if err := out.WriteLong(int64(fp)); err != nil {
				return err
			}
		}
	}
	return out.Flush()
}

func (s *MemFPSet) CommitChkpt() error {
	s.requireMemFPSet()
	return s.CommitChkptFile(s.filename)
}

func (s *MemFPSet) CommitChkptFile(fname string) error {
	s.requireMemFPSet()
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if _, err := os.Stat(oldChkpt); err == nil {
		if err := os.Remove(oldChkpt); err != nil {
			return NewIOException(fmt.Sprintf("MemFPSet.commitChkpt: cannot delete %s", oldChkpt))
		}
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException(fmt.Sprintf("MemFPSet.commitChkpt: cannot delete %s", oldChkpt))
	}
	return nil
}

func (s *MemFPSet) RecoverFile(fname string) (err error) {
	s.requireMemFPSet()
	file, err := openFileInputStream(s.chkptName(fname, "chkpt"))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	// The source refills eagerly, including inside the read of the last
	// fingerprint in each buffer, before that fingerprint is inserted.
	in, err := NewBufferedDataInputStream(file)
	if err != nil {
		return err
	}
	for !in.AtEOF() {
		fp, err := in.ReadLong()
		if errors.Is(err, io.EOF) {
			return NewTLCRuntimeException(ECSystemDiskIOErrorForFile, "checkpoints")
		}
		if err != nil {
			return err
		}
		if err := s.RecoverFP(uint64(fp)); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemFPSet) Recover(trace FPSetTrace) error {
	s.requireMemFPSet()
	return s.RecoverFile(s.filename)
}

func (s *MemFPSet) RecoverTrace(trace FPSetTrace) error { return s.Recover(trace) }

func (s *MemFPSet) RecoverFP(fp uint64) error {
	s.requireMemFPSet()
	if s.Put(fp) {
		return NewTLCRuntimeException(ECTLCFPNotInSet)
	}
	return nil
}

func (s *MemFPSet) UnexportObject(force bool) { s.requireMemFPSet() }

func (s *MemFPSet) chkptName(fname string, ext string) string {
	return s.metadir + string(os.PathSeparator) + fname + ".fp." + ext
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func javaLongSub(a int64, b int64) int64 {
	return int64(uint64(a) - uint64(b))
}

func javaLongAbs(v int64) int64 {
	if v >= 0 {
		return v
	}
	return int64(uint64(0) - uint64(v))
}

func javaLongMin(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func javaLongDistanceIfNonnegative(a int64, b int64) int64 {
	if a > b {
		return javaLongSub(a, b)
	}
	return javaLongSub(b, a)
}

func javaLongMinBits(a uint64, b uint64) uint64 {
	if int64(a) < int64(b) {
		return a
	}
	return b
}

type NoopFPSet struct {
	fpSetLifecycle
	statesSeen atomic.Uint64
	config     *FPSetConfiguration
}

func NewNoopFPSet(config *FPSetConfiguration) *NoopFPSet {
	return &NoopFPSet{config: config}
}

func (s *NoopFPSet) requireNoopFPSet() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *NoopFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.requireNoopFPSet()
	return s
}

func (s *NoopFPSet) Size() uint64       { s.requireNoopFPSet(); return 0 }
func (s *NoopFPSet) Sizeof() uint64     { s.requireNoopFPSet(); return 0 }
func (s *NoopFPSet) Put(fp uint64) bool { s.requireNoopFPSet(); return false }
func (s *NoopFPSet) Contains(fp uint64) bool {
	s.requireNoopFPSet()
	return false
}
func (s *NoopFPSet) CheckFPs() uint64 { s.requireNoopFPSet(); return 0 }
func (s *NoopFPSet) BeginChkpt() error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) BeginChkptFile(fname string) error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) CommitChkpt() error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) CommitChkptFile(fname string) error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) RecoverFile(fname string) error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) Recover(trace FPSetTrace) error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) RecoverTrace(trace FPSetTrace) error { return s.Recover(trace) }
func (s *NoopFPSet) RecoverFP(fp uint64) error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) Close() { s.requireNoopFPSet() }
func (s *NoopFPSet) AddThread() error {
	s.requireNoopFPSet()
	return nil
}
func (s *NoopFPSet) IncWorkers(num int) { s.requireNoopFPSet() }
func (s *NoopFPSet) Exit(cleanup bool) error {
	s.requireNoopFPSet()
	fpSetBaseExit(s)
	return nil
}
func (s *NoopFPSet) CheckInvariant(expectFPs ...uint64) bool {
	s.requireNoopFPSet()
	requireInvariantExpectedCount(expectFPs)
	return true
}
func (s *NoopFPSet) UnexportObject(force bool) { s.requireNoopFPSet() }

func (s *NoopFPSet) PutBlock(fpv *LongVec) *BitVector {
	s.requireNoopFPSet()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		bv.Set(i)
	}
	return bv
}

func (s *NoopFPSet) ContainsBlock(fpv *LongVec) *BitVector {
	s.requireNoopFPSet()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	// Keep FPSet's separate load/modify/store (including lost updates),
	// with atomic snapshots instead of an unsynchronized Go memory access.
	s.statesSeen.Store(s.statesSeen.Load() + uint64(fpv.Size()))
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		bv.Set(i)
	}
	return bv
}

func (s *NoopFPSet) GetStatesSeen() uint64 {
	s.requireNoopFPSet()
	return s.statesSeen.Load()
}

func (s *NoopFPSet) GetConfiguration() *FPSetConfiguration {
	s.requireNoopFPSet()
	return s.config
}

type MemFPSet1 struct {
	fpSetLifecycle
	mu         distributedServerMonitor
	metadir    string
	filename   string
	set        *SetOfLong
	statesSeen uint64
	config     *FPSetConfiguration
}

func NewMemFPSet1(config *FPSetConfiguration) *MemFPSet1 {
	return &MemFPSet1{
		set:    NewSetOfLong(10001),
		config: config,
	}
}

func (s *MemFPSet1) requireMemFPSet1() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *MemFPSet1) Init(numThreads int, metadir string, filename string) FPSet {
	s.requireMemFPSet1()
	s.metadir = metadir
	s.filename = filename
	return s
}

func (s *MemFPSet1) Size() uint64 {
	s.requireMemFPSet1()
	return uint64(s.set.Size())
}

func (s *MemFPSet1) Sizeof() uint64 {
	s.requireMemFPSet1()
	return uint64(8 + s.set.Sizeof())
}

func (s *MemFPSet1) Put(fp uint64) bool {
	s.requireMemFPSet1()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.Put(int64(fp))
}

func (s *MemFPSet1) Contains(fp uint64) bool {
	s.requireMemFPSet1()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.Contains(int64(fp))
}

func (s *MemFPSet1) PutBlock(fpv *LongVec) *BitVector {
	s.requireMemFPSet1()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	size := fpv.Size()
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Put(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet1) ContainsBlock(fpv *LongVec) *BitVector {
	s.requireMemFPSet1()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	size := fpv.Size()
	fpSetRecordStatesSeen(&s.statesSeen, size)
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet1) GetStatesSeen() uint64 {
	s.requireMemFPSet1()
	return atomic.LoadUint64(&s.statesSeen)
}

func (s *MemFPSet1) GetConfiguration() *FPSetConfiguration {
	s.requireMemFPSet1()
	return s.config
}

func (s *MemFPSet1) Close() { s.requireMemFPSet1() }

func (s *MemFPSet1) AddThread() error {
	s.requireMemFPSet1()
	return nil
}

func (s *MemFPSet1) IncWorkers(num int) { s.requireMemFPSet1() }

func (s *MemFPSet1) Exit(cleanup bool) error {
	s.requireMemFPSet1()
	fpSetBaseExit(s)
	return completeFingerprintExit(s.metadir, cleanup)
}

func (s *MemFPSet1) CheckInvariant(expectFPs ...uint64) bool {
	s.requireMemFPSet1()
	requireInvariantExpectedCount(expectFPs)
	return true
}

func (s *MemFPSet1) CheckFPs() uint64 {
	s.requireMemFPSet1()
	return uint64(s.set.CheckFPs())
}

func (s *MemFPSet1) BeginChkpt() error {
	s.requireMemFPSet1()
	return s.BeginChkptFile(s.filename)
}

func (s *MemFPSet1) BeginChkptFile(fname string) (err error) {
	s.requireMemFPSet1()
	path := s.chkptName(fname, "tmp")
	file, err := openFileOutputStream(path)
	if err != nil {
		return err
	}
	// FileUtil.newDFOS writes primitives directly, without buffering ahead
	// of a failed field. Cleanup must close the file without replaying writes.
	out := NewValueOutputStreamWithoutHandles(file)
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	return s.set.BeginChkpt(out)
}

func (s *MemFPSet1) CommitChkpt() error {
	s.requireMemFPSet1()
	return s.CommitChkptFile(s.filename)
}

func (s *MemFPSet1) CommitChkptFile(fname string) error {
	s.requireMemFPSet1()
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if _, err := os.Stat(oldChkpt); err == nil {
		if err := os.Remove(oldChkpt); err != nil {
			return NewIOException(fmt.Sprintf("MemFPSet1.commitChkpt: cannot delete %s", oldChkpt))
		}
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException(fmt.Sprintf("MemFPSet1.commitChkpt: cannot delete %s", oldChkpt))
	}
	return nil
}

func (s *MemFPSet1) RecoverFile(fname string) (err error) {
	s.requireMemFPSet1()
	file, err := openFileInputStream(s.chkptName(fname, "chkpt"))
	if err != nil {
		return err
	}
	// FileUtil.newDFIS wraps the file directly in DataInputStream. Read each
	// primitive on demand so I/O cannot move ahead of SetOfLong's assignments.
	in := NewValueInputStreamWithoutHandles(file)
	defer func() {
		if closeErr := in.Close(); err == nil {
			err = closeErr
		}
	}()
	return s.set.Recover(in)
}

func (s *MemFPSet1) Recover(trace FPSetTrace) error {
	s.requireMemFPSet1()
	return s.RecoverFile(s.filename)
}

func (s *MemFPSet1) RecoverTrace(trace FPSetTrace) error { return s.Recover(trace) }

func (s *MemFPSet1) RecoverFP(fp uint64) error {
	s.requireMemFPSet1()
	if s.set.Put(int64(fp)) {
		return NewTLCRuntimeException(ECTLCFPNotInSet)
	}
	return nil
}

func (s *MemFPSet1) UnexportObject(force bool) { s.requireMemFPSet1() }

func (s *MemFPSet1) chkptName(fname string, ext string) string {
	return s.metadir + string(os.PathSeparator) + fname + ".fp." + ext
}

const memFPSet2LogSpineSize = 24

type MemFPSet2 struct {
	fpSetLifecycle
	mu         distributedServerMonitor
	metadir    string
	filename   string
	table      [][]byte
	count      uint64
	mask       uint64
	statesSeen uint64
	config     *FPSetConfiguration
}

func NewMemFPSet2(config *FPSetConfiguration) *MemFPSet2 {
	spineSize := 1 << memFPSet2LogSpineSize
	return &MemFPSet2{
		table:  make([][]byte, spineSize),
		mask:   uint64(spineSize - 1),
		config: config,
	}
}

func (s *MemFPSet2) requireMemFPSet2() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *MemFPSet2) Init(numThreads int, metadir string, filename string) FPSet {
	s.requireMemFPSet2()
	s.metadir = metadir
	s.filename = metadir + string(os.PathSeparator) + filename
	return s
}

func (s *MemFPSet2) Size() uint64 {
	s.requireMemFPSet2()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

func (s *MemFPSet2) Sizeof() uint64 {
	s.requireMemFPSet2()
	s.mu.Lock()
	defer s.mu.Unlock()
	size := uint64(28)
	size += 16 + uint64(len(s.table))*8
	for _, bucket := range s.table {
		if bucket != nil {
			size += 16 + uint64(len(bucket))
		}
	}
	return size
}

func (s *MemFPSet2) Put(fp uint64) bool {
	s.requireMemFPSet2()
	s.mu.Lock()
	defer s.mu.Unlock()
	index := int(fp & s.mask)
	bucket := s.table[index]
	b1, b2, b3, b4, b5 := memFPSet2HighBytes(fp)
	for i := 0; i < len(bucket); i += 5 {
		if bucket[i] == b1 && bucket[i+1] == b2 && bucket[i+2] == b3 && bucket[i+3] == b4 && bucket[i+4] == b5 {
			return true
		}
	}
	s.table[index] = append(bucket, b1, b2, b3, b4, b5)
	s.count++
	return false
}

func (s *MemFPSet2) Contains(fp uint64) bool {
	s.requireMemFPSet2()
	s.mu.Lock()
	defer s.mu.Unlock()
	index := int(fp & s.mask)
	bucket := s.table[index]
	b1, b2, b3, b4, b5 := memFPSet2HighBytes(fp)
	for i := 0; i < len(bucket); i += 5 {
		if bucket[i] == b1 && bucket[i+1] == b2 && bucket[i+2] == b3 && bucket[i+3] == b4 && bucket[i+4] == b5 {
			return true
		}
	}
	return false
}

func (s *MemFPSet2) PutBlock(fpv *LongVec) *BitVector {
	s.requireMemFPSet2()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	size := fpv.Size()
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Put(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet2) ContainsBlock(fpv *LongVec) *BitVector {
	s.requireMemFPSet2()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	size := fpv.Size()
	fpSetRecordStatesSeen(&s.statesSeen, size)
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet2) GetStatesSeen() uint64 {
	s.requireMemFPSet2()
	return atomic.LoadUint64(&s.statesSeen)
}

func (s *MemFPSet2) GetConfiguration() *FPSetConfiguration {
	s.requireMemFPSet2()
	return s.config
}

func (s *MemFPSet2) Close() { s.requireMemFPSet2() }

func (s *MemFPSet2) AddThread() error {
	s.requireMemFPSet2()
	return nil
}

func (s *MemFPSet2) IncWorkers(num int) { s.requireMemFPSet2() }

func (s *MemFPSet2) Exit(cleanup bool) error {
	s.requireMemFPSet2()
	fpSetBaseExit(s)
	return completeFingerprintExit(s.metadir, cleanup)
}

func (s *MemFPSet2) CheckInvariant(expectFPs ...uint64) bool {
	s.requireMemFPSet2()
	requireInvariantExpectedCount(expectFPs)
	return true
}

func (s *MemFPSet2) CheckFPs() uint64 {
	s.requireMemFPSet2()
	dis := int64(1<<63 - 1)
	for i, bucket := range s.table {
		low := uint64(i) & 0xffffff
		j := 0
		for j < len(bucket) {
			fp := memFPSet2Fingerprint(low, bucket[j], bucket[j+1], bucket[j+2], bucket[j+3], bucket[j+4])
			j += 5
			for k := j; k < len(bucket); k += 5 {
				fp1 := memFPSet2Fingerprint(low, bucket[k], bucket[k+1], bucket[k+2], bucket[k+3], bucket[k+4])
				dis1 := javaLongDistanceIfNonnegative(int64(fp), int64(fp1))
				if dis1 >= 0 {
					dis = javaLongMin(dis, dis1)
				}
			}
			for k := i + 1; k < len(s.table); k++ {
				bucket1 := s.table[k]
				if bucket1 == nil {
					continue
				}
				low1 := uint64(k) & 0xffffff
				// Java MemFPSet2.checkFPs accidentally iterates the current
				// bucket here, not bucket1. Preserve that diagnostic quirk.
				for k1 := 0; k1 < len(bucket); k1 += 5 {
					fp1 := memFPSet2Fingerprint(low1, bucket[k1], bucket[k1+1], bucket[k1+2], bucket[k1+3], bucket[k1+4])
					dis1 := javaLongDistanceIfNonnegative(int64(fp), int64(fp1))
					if dis1 >= 0 {
						dis = javaLongMin(dis, dis1)
					}
				}
			}
		}
	}
	return uint64(dis)
}

func (s *MemFPSet2) BeginChkpt() error {
	s.requireMemFPSet2()
	return s.BeginChkptFile(s.filename)
}

func (s *MemFPSet2) BeginChkptFile(fname string) (err error) {
	s.requireMemFPSet2()
	path := s.chkptName(fname, "tmp")
	file, err := openFileOutputStream(path)
	if err != nil {
		return err
	}
	// Cleanup must not flush a buffer whose write has already failed.
	out := NewBufferedDataOutputStream(file)
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = bufferedRandomAccessFileIOError(closeErr)
		}
	}()
	for i, bucket := range s.table {
		low := uint64(i) & 0xffffff
		for j := 0; j < len(bucket); j += 5 {
			fp := memFPSet2Fingerprint(low, bucket[j], bucket[j+1], bucket[j+2], bucket[j+3], bucket[j+4])
			if err := out.WriteLong(int64(fp)); err != nil {
				return err
			}
		}
	}
	return out.Flush()
}

func (s *MemFPSet2) CommitChkpt() error {
	s.requireMemFPSet2()
	return s.CommitChkptFile(s.filename)
}

func (s *MemFPSet2) CommitChkptFile(fname string) error {
	s.requireMemFPSet2()
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if _, err := os.Stat(oldChkpt); err == nil {
		if err := os.Remove(oldChkpt); err != nil {
			return NewIOException(fmt.Sprintf("MemFPSet2.commitChkpt: cannot delete %s", oldChkpt))
		}
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return NewIOException(fmt.Sprintf("MemFPSet2.commitChkpt: cannot delete %s", oldChkpt))
	}
	return nil
}

func (s *MemFPSet2) RecoverFile(fname string) (err error) {
	s.requireMemFPSet2()
	file, err := openFileInputStream(s.chkptName(fname, "chkpt"))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	in, err := NewBufferedDataInputStream(file)
	if err != nil {
		return err
	}
	for !in.AtEOF() {
		fp, err := in.ReadLong()
		if errors.Is(err, io.EOF) {
			return NewIOException("MemFPSet2.recover: failed.")
		}
		if err != nil {
			return err
		}
		if err := s.RecoverFP(uint64(fp)); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemFPSet2) Recover(trace FPSetTrace) error {
	s.requireMemFPSet2()
	return s.RecoverFile(s.filename)
}

func (s *MemFPSet2) RecoverTrace(trace FPSetTrace) error { return s.Recover(trace) }

func (s *MemFPSet2) RecoverFP(fp uint64) error {
	s.requireMemFPSet2()
	if s.Put(fp) {
		return NewTLCRuntimeException(ECTLCFPNotInSet)
	}
	return nil
}

func (s *MemFPSet2) UnexportObject(force bool) { s.requireMemFPSet2() }

func (s *MemFPSet2) chkptName(fname string, ext string) string {
	return s.metadir + string(os.PathSeparator) + fname + ".fp." + ext
}

func memFPSet2HighBytes(fp uint64) (byte, byte, byte, byte, byte) {
	return byte((fp >> 24) & 0xff),
		byte((fp >> 32) & 0xff),
		byte((fp >> 40) & 0xff),
		byte((fp >> 48) & 0xff),
		byte((fp >> 56) & 0xff)
}

func memFPSet2Fingerprint(low uint64, b1 byte, b2 byte, b3 byte, b4 byte, b5 byte) uint64 {
	return (uint64(b5) << 56) | (uint64(b4) << 48) | (uint64(b3) << 40) | (uint64(b2) << 32) | (uint64(b1) << 24) | low
}

type MultiFPSet struct {
	fpSetLifecycle
	Sets         []FPSet
	FPBits       int
	Shift        uint
	metadir      string
	filename     string
	statesSeen   atomic.Uint64
	config       *FPSetConfiguration
	setsCaptured bool
	setsOwner    []FPSet
}

func NewMultiFPSet(config *FPSetConfiguration) *MultiFPSet {
	bits := config.GetFpBits()
	if bits <= 0 || bits > multiFPSetMaxFPBits {
		failure := newTLCError(ECGeneral, "Illegal number of FPSets found.")
		failure.Runtime = true
		panic(failure)
	}
	childConfig := NewMultiFPSetConfiguration(config)
	count := 1 << bits
	// Java's hard heap limit throws a catchable OutOfMemoryError while building
	// this object graph. Go's GC limit is soft and a fatal allocator failure
	// cannot be recovered. Reject a provably impossible graph before allocation;
	// this is a lower bound from concrete native sizes, not a child-count cap.
	perChild := uint64(reflect.TypeOf((*FPSet)(nil)).Elem().Size())
	implementation := childConfig.GetImplementation()
	switch implementation {
	case "tlc2.tool.fp.FPSet", "tlc2.tool.fp.DiskFPSet", "tlc2.tool.fp.HeapBasedDiskFPSet", "tlc2.tool.fp.NonCheckpointableDiskFPSet":
		// These known, non-instantiable classes return nil from the factory;
		// their temporary configurations are not retained by the parent list.
	default:
		perChild += uint64(reflect.TypeOf(FPSetConfiguration{}).Size())
	}
	if !fpSetSupportsArchitecture(implementation) || implementation == "tlc2.tool.fp.MSBDiskFPSet" || implementation == "tlc2.tool.fp.LSBDiskFPSet" {
		perChild += uint64(reflect.TypeOf(DiskFPSet{}).Size())
	}
	budget := tlcRuntimeMaxHeapMemoryBytes()
	if budget < 0 || uint64(count) > uint64(budget)/perChild {
		panic(NewOutOfMemoryError("Java heap space"))
	}
	set := &MultiFPSet{
		FPBits: bits,
		Shift:  uint(64 - bits),
		config: config,
	}
	set.Sets = set.GetNestedFPSets(config)
	set.setsOwner = set.Sets
	set.setsCaptured = true
	return set
}

func (s *MultiFPSet) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *MultiFPSet) ownedSets() []FPSet {
	if s.setsCaptured {
		return s.setsOwner
	}
	return s.Sets
}

func requireMultiFPSetChild(set FPSet) FPSet {
	if isNil(set) {
		panic(NewNullPointerException())
	}
	return set
}

// GetNestedFPSets is MultiFPSet.getNestedFPSets.
func (s *MultiFPSet) GetNestedFPSets(config *FPSetConfiguration) []FPSet {
	s.require()
	sets := make([]FPSet, config.GetMultiFPSetCnt())
	for i := range sets {
		// Java constructs a new child configuration for every nested set.
		sets[i] = NewFPSet(NewMultiFPSetConfiguration(config))
	}
	return sets
}

func (s *MultiFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.require()
	numThreads = int(int32(numThreads))
	s.metadir = metadir
	s.filename = filename
	// IntStream.parallel initializes distinct children independently. The
	// callback ignores init's return value and wraps its checked IOException.
	var pending sync.WaitGroup
	var failureMu sync.Mutex
	var failure error
	sets := s.ownedSets()
	initialize := func(i int) {
		defer pending.Done()
		err := func() (err error) {
			defer func() {
				if failure := recover(); failure != nil {
					err = panicValueAsError(failure)
				}
			}()
			requireMultiFPSetChild(sets[i]).Init(numThreads, metadir, fmt.Sprintf("%s_%d", filename, i))
			return nil
		}()
		if isJavaIOException(err) {
			err = NewRuntimeExceptionFromCause(err)
		}
		if err != nil {
			failureMu.Lock()
			if failure == nil {
				failure = err
			}
			failureMu.Unlock()
		}
	}
	pending.Add(len(sets))
	for i := 0; i < len(sets)-1; i++ {
		go initialize(i)
	}
	if len(sets) != 0 {
		initialize(len(sets) - 1)
	}
	pending.Wait()
	if failure != nil {
		panic(failure)
	}
	return s
}

func (s *MultiFPSet) Size() uint64 {
	s.require()
	var total uint64
	for _, value := range parallelNestedFPSetCalls(s.ownedSets(), "size", false, func(set FPSet) uint64 { return set.Size() }) {
		total += value
	}
	return total
}

func (s *MultiFPSet) Sizeof() uint64 {
	s.require()
	var total uint64
	for _, set := range s.ownedSets() {
		total += requireMultiFPSetChild(set).Sizeof()
	}
	return total
}

// GetFPSet is MultiFPSet.getFPSet.
func (s *MultiFPSet) GetFPSet(fp uint64) FPSet {
	s.require()
	idx := int(fp >> s.Shift)
	return s.ownedSets()[idx]
}

func (s *MultiFPSet) Put(fp uint64) bool {
	return requireMultiFPSetChild(s.GetFPSet(fp)).Put(fp)
}

func (s *MultiFPSet) Contains(fp uint64) bool {
	return requireMultiFPSetChild(s.GetFPSet(fp)).Contains(fp)
}

// GetFPSets returns the source method's fresh array with the same child objects.
func (s *MultiFPSet) GetFPSets() []FPSet {
	s.require()
	owned := s.ownedSets()
	sets := make([]FPSet, len(owned))
	copy(sets, owned)
	return sets
}

func (s *MultiFPSet) PutBlock(fpv *LongVec) *BitVector {
	s.require()
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

func (s *MultiFPSet) ContainsBlock(fpv *LongVec) *BitVector {
	s.require()
	if fpv == nil {
		panic(NewNullPointerException())
	}
	// Java's inherited FPSet counter is a separate read/modify/write.
	s.statesSeen.Store(s.statesSeen.Load() + uint64(fpv.Size()))
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MultiFPSet) GetStatesSeen() uint64 {
	s.require()
	return s.statesSeen.Load()
}

func (s *MultiFPSet) GetConfiguration() *FPSetConfiguration {
	s.require()
	return s.config
}

func (s *MultiFPSet) CheckFPs() uint64 {
	s.require()
	dis := uint64(1<<63 - 1)
	for _, value := range parallelNestedFPSetCalls(s.ownedSets(), "check fingerprints", true, func(set FPSet) uint64 { return set.CheckFPs() }) {
		dis = javaLongMinBits(dis, value)
	}
	return dis
}

func (s *MultiFPSet) CheckInvariant(expectFPs ...uint64) bool {
	s.require()
	requireInvariantExpectedCount(expectFPs)
	// Only the no-argument overload is overridden by MultiFPSet in the source.
	// The inherited expected-count overload returns true without child checks.
	if len(expectFPs) != 0 {
		return true
	}
	var stopped atomic.Bool
	for _, valid := range parallelNestedFPSetCalls(s.ownedSets(), "check invariant", true, func(set FPSet) bool {
		// Source allMatch may skip work not started when a false result is
		// already known. Work already in progress still belongs to this call.
		if stopped.Load() {
			return true
		}
		valid := set.CheckInvariant()
		if !valid {
			stopped.Store(true)
		}
		return valid
	}) {
		if !valid {
			return false
		}
	}
	return true
}

// Source child size/check calls use parallel streams. Only check lambdas wrap
// child IOException as an operation failure. Join started storage work before
// returning or propagating a failure, retaining causes with native wrapping.
func parallelNestedFPSetCalls[T any](sets []FPSet, operation string, wrapIO bool, call func(FPSet) T) []T {
	values := make([]T, len(sets))
	failures := make([]any, len(sets))
	var pending sync.WaitGroup
	pending.Add(len(sets))
	for i, set := range sets {
		go func(i int, set FPSet) {
			defer pending.Done()
			defer func() {
				if failure := recover(); failure != nil {
					if err, ok := failure.(error); wrapIO && ok && isJavaIOException(err) {
						failures[i] = fmt.Errorf("%s partition %d: %w", operation, i, err)
					} else {
						failures[i] = failure
					}
				}
			}()
			values[i] = call(requireMultiFPSetChild(set))
		}(i, set)
	}
	pending.Wait()
	for _, failure := range failures {
		if failure != nil {
			panic(failure)
		}
	}
	return values
}

func (s *MultiFPSet) BeginChkpt() error {
	s.require()
	for _, set := range s.ownedSets() {
		if err := requireMultiFPSetChild(set).BeginChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) BeginChkptFile(fname string) error {
	s.require()
	return s.namedCheckpoint("begin checkpoint", fname, func(set FPSet, name string) error { return set.BeginChkptFile(name) })
}

func (s *MultiFPSet) CommitChkpt() error {
	s.require()
	for _, set := range s.ownedSets() {
		if err := requireMultiFPSetChild(set).CommitChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) CommitChkptFile(fname string) error {
	s.require()
	return s.namedCheckpoint("commit checkpoint", fname, func(set FPSet, name string) error { return set.CommitChkptFile(name) })
}

func (s *MultiFPSet) RecoverFile(fname string) error {
	s.require()
	return s.namedCheckpoint("recover checkpoint", fname, func(set FPSet, name string) error { return set.RecoverFile(name) })
}

// Named checkpoint operations run independently across children, matching the
// source's parallel traversal. Join native goroutines before returning or
// propagating a panic so no storage operation outlives its caller. A child IO
// failure is an operation failure here, rather than a remote-server outage that
// the fingerprint manager may report and ignore. Keep its cause with Go's %w.
func (s *MultiFPSet) namedCheckpoint(operation, fname string, call func(FPSet, string) error) error {
	type result struct {
		err        error
		panicValue any
	}
	sets := s.ownedSets()
	results := make([]result, len(sets))
	var pending sync.WaitGroup
	pending.Add(len(sets))
	for i, set := range sets {
		go func(i int, set FPSet) {
			defer pending.Done()
			defer func() {
				if value := recover(); value != nil {
					if err, ok := value.(error); ok && isJavaIOException(err) {
						results[i].err = fmt.Errorf("%s %s_%d: %w", operation, fname, i, err)
					} else {
						results[i].panicValue = value
					}
				}
			}()
			err := call(requireMultiFPSetChild(set), fmt.Sprintf("%s_%d", fname, i))
			if isJavaIOException(err) {
				err = fmt.Errorf("%s %s_%d: %w", operation, fname, i, err)
			}
			results[i].err = err
		}(i, set)
	}
	pending.Wait()
	for _, result := range results {
		if result.panicValue != nil {
			panic(result.panicValue)
		}
		if result.err != nil {
			return result.err
		}
	}
	return nil
}

func (s *MultiFPSet) Recover(trace FPSetTrace) error {
	s.require()
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
		if err := requireMultiFPSetChild(s.GetFPSet(fp)).RecoverFP(fp); err != nil {
			return err
		}
	}
	return elements.Close()
}

func (s *MultiFPSet) RecoverTrace(trace FPSetTrace) error { return s.Recover(trace) }

func (s *MultiFPSet) RecoverFP(fp uint64) error {
	s.require()
	if s.Put(fp) {
		return NewTLCRuntimeException(ECTLCFPNotInSet)
	}
	return nil
}

func (s *MultiFPSet) Close() {
	s.require()
	for _, set := range s.ownedSets() {
		requireMultiFPSetChild(set).Close()
	}
}

func (s *MultiFPSet) AddThread() error {
	s.require()
	// MultiFPSet inherits the base no-op; only IncWorkers visits children.
	return nil
}

func (s *MultiFPSet) IncWorkers(num int) {
	s.require()
	num = int(int32(num))
	for _, set := range s.ownedSets() {
		requireMultiFPSetChild(set).IncWorkers(num)
	}
}

func (s *MultiFPSet) Exit(cleanup bool) error {
	s.require()
	fpSetBaseExit(s)
	for _, set := range s.ownedSets() {
		if err := requireMultiFPSetChild(set).Exit(cleanup); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) UnexportObject(force bool) {
	s.require()
	for _, set := range s.ownedSets() {
		requireMultiFPSetChild(set).UnexportObject(force)
	}
}

func (s *MemFPSet) fpSetMonitor() *distributedServerMonitor { return &s.mu }

func (s *MemFPSet1) fpSetMonitor() *distributedServerMonitor { return &s.mu }

func (s *MemFPSet2) fpSetMonitor() *distributedServerMonitor { return &s.mu }
