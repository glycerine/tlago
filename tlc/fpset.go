package tlc

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
)

const (
	fpSetLongSize       = 8
	multiFPSetMaxFPBits = 30
	multiFPSetMinFPBits = 0
	FPSetImplProperty   = "tlc2.tool.fp.FPSet.impl"
)

const (
	tlcRuntimeMinFPMemSize        = int64(20 * (1 << 19))
	tlcRuntimeDefaultNonHeapBytes = int64(64 * 1024 * 1024)
)

type FPSetConfiguration struct {
	FPBits         int
	MemoryInBytes  int64
	Ratio          float64
	Implementation string
	NoNesting      bool
	MemoryDivisor  int64
}

func NewFPSetConfiguration() *FPSetConfiguration {
	return NewFPSetConfigurationWithRatio(0.25)
}

func NewFPSetConfigurationWithRatio(ratio float64) *FPSetConfiguration {
	return NewFPSetConfigurationWithRatioAndImplementation(ratio, fpSetImplementationFromEnv())
}

func NewFPSetConfigurationWithRatioAndImplementation(ratio float64, implementation string) *FPSetConfiguration {
	if implementation == "" {
		implementation = GetFPSetImplementationDefault()
	}
	return &FPSetConfiguration{
		FPBits:         1,
		MemoryInBytes:  -1,
		Ratio:          ratio,
		Implementation: implementation,
	}
}

func (c *FPSetConfiguration) AllowsNesting() bool {
	return !c.NoNesting && c.GetFPBits() > 0
}

func (c *FPSetConfiguration) GetFPBits() int {
	if c == nil {
		return 1
	}
	if c.FPBits == 0 && isDiskFPSetImplementation(c.Implementation) {
		c.FPBits = 1
	}
	return c.FPBits
}

func (c *FPSetConfiguration) SetFPBits(fpBits int) {
	if !IsValidFPBits(fpBits) {
		panic("illegal number of FPSets")
	}
	c.FPBits = fpBits
}

func (c *FPSetConfiguration) GetMemoryInBytes() int64 {
	if c == nil {
		return 0
	}
	divisor := c.MemoryDivisor
	if divisor <= 0 {
		divisor = 1
	}
	var memory int64
	if FPSetAllocatesOnHeap(c.Implementation) {
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
	memory := c.GetMemoryInBytes()
	if memory <= 0 {
		return memory
	}
	return int64(math.Floor(float64(memory) / fpSetLongSize))
}

func (c *FPSetConfiguration) GetMultiFPSetCnt() int {
	return 1 << c.GetFPBits()
}

func (c *FPSetConfiguration) SetRatio(ratio float64) {
	if ratio < 0 || ratio > 1 {
		panic("FPSet ratio out of range")
	}
	c.Ratio = ratio
}

func (c *FPSetConfiguration) GetRatio() float64 {
	if c == nil {
		return 0
	}
	return c.Ratio
}

func (c *FPSetConfiguration) SetMemory(memory int64) {
	if memory < 0 {
		panic("FPSet memory cannot be negative")
	}
	c.MemoryInBytes = memory
}

func (c *FPSetConfiguration) GetImplementation() string {
	if c == nil {
		return ""
	}
	return c.Implementation
}

func (c *FPSetConfiguration) SetImplementation(implementation string) {
	if implementation == "" {
		implementation = GetFPSetImplementationDefault()
	}
	c.Implementation = implementation
}

func NewMultiFPSetConfiguration(config *FPSetConfiguration) *FPSetConfiguration {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	child := *config
	child.NoNesting = true
	child.MemoryDivisor = int64(config.GetMultiFPSetCnt())
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
	if fpMemSize < 0 {
		return 0
	}
	return int64(fpMemSize)
}

func tlcRuntimeMaxHeapMemoryBytes() int64 {
	limit := debug.SetMemoryLimit(-1)
	if limit > 0 && limit < math.MaxInt64/4 {
		return limit
	}
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	if stats.Sys > 0 {
		return int64(stats.Sys)
	}
	return tlcRuntimeMinFPMemSize * 4
}

func tlcRuntimeNonHeapPhysicalMemory() int64 {
	return tlcRuntimeDefaultNonHeapBytes
}

func fpSetImplementationFromEnv() string {
	if value, ok := tlcLookupSystemProperty(FPSetImplProperty); ok {
		return value
	}
	if value, ok := os.LookupEnv("TLAGO_FPSET_IMPL"); ok {
		return value
	}
	return GetFPSetImplementationDefault()
}

func GetFPSetImplementations() []string {
	return []string{
		"tlc2.tool.fp.MSBDiskFPSet",
		"tlc2.tool.fp.LSBDiskFPSet",
		"tlc2.tool.fp.OffHeapDiskFPSet",
	}
}

func GetFPSetImplementationDefault() string {
	return "tlc2.tool.fp.MSBDiskFPSet"
}

func FPSetAllocatesOnHeap(implementation string) bool {
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

func GetFPSetVMArguments(implementation string, memoryMiB int64) string {
	if FPSetAllocatesOnHeap(implementation) {
		return fmt.Sprintf("-Xmx%dm", memoryMiB)
	}
	return fmt.Sprintf("-XX:MaxDirectMemorySize=%dm", memoryMiB)
}

func IsValidFPBits(fpBits int) bool {
	return fpBits >= multiFPSetMinFPBits && fpBits <= multiFPSetMaxFPBits
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
	Recover() error
	RecoverFile(fname string) error
	RecoverTrace(trace *TLCTrace) error
	RecoverFP(fp uint64) error
	UnexportObject(force bool)
}

func NewFPSet(config *FPSetConfiguration) FPSet {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	if config.AllowsNesting() {
		return NewMultiFPSet(config)
	}
	switch config.GetImplementation() {
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
	case "tlc2.tool.fp.MSBDiskFPSet", "":
		return NewMSBDiskFPSet(config)
	case "tlc2.tool.fp.NonCheckpointableDiskFPSet":
		return NewNonCheckpointableDiskFPSet(config)
	case "tlc2.tool.fp.OffHeapDiskFPSet":
		return NewOffHeapDiskFPSet(config)
	default:
		PrintWarning(ECGeneral, "unsuccessfully trying to load custom FPSet class: "+config.GetImplementation())
		return nil
	}
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
	mu         sync.Mutex
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
	if config == nil {
		config = NewFPSetConfiguration()
	}
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

func (s *MemFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.metadir = metadir
	s.filename = filename
	return s
}

func (s *MemFPSet) Size() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

func (s *MemFPSet) Sizeof() uint64 {
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
	if fpv == nil {
		return NewBitVector(0)
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
	if fpv == nil {
		return NewBitVector(0)
	}
	size := fpv.Size()
	s.mu.Lock()
	s.statesSeen += uint64(size)
	s.mu.Unlock()

	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet) GetStatesSeen() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statesSeen
}

func (s *MemFPSet) GetConfiguration() *FPSetConfiguration {
	if s == nil || s.config == nil {
		return NewFPSetConfiguration()
	}
	return s.config
}

func (s *MemFPSet) Close() {}

func (s *MemFPSet) AddThread() error {
	return nil
}

func (s *MemFPSet) IncWorkers(num int) {}

func (s *MemFPSet) Exit(cleanup bool) error {
	if cleanup && s.metadir != "" {
		return os.RemoveAll(s.metadir)
	}
	return nil
}

func (s *MemFPSet) CheckInvariant(expectFPs ...uint64) bool {
	return len(expectFPs) == 0 || s.Size() == expectFPs[0]
}

func (s *MemFPSet) Contains(fp uint64) bool {
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
		var left, right []uint64
		for _, fp := range bucket {
			if fp&oneBitMask == 0 {
				left = append(left, fp)
			} else {
				right = append(right, fp)
			}
		}
		newTable[i] = left
		newTable[i+oldCapacity] = right
	}

	s.threshold *= 2
	s.table = newTable
	s.mask = uint64(len(newTable) - 1)
}

func (s *MemFPSet) CheckFPs() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	dis := uint64(1<<63 - 1)
	for i, bucket := range s.table {
		for j, x := range bucket {
			for k := j + 1; k < len(bucket); k++ {
				dis = minUint64(dis, absDiffUint64(x, bucket[k]))
			}
			for _, otherBucket := range s.table[i+1:] {
				for _, y := range otherBucket {
					dis = minUint64(dis, absDiffUint64(x, y))
				}
			}
		}
	}
	return dis
}

func (s *MemFPSet) BeginChkpt() error {
	return s.BeginChkptFile(s.filename)
}

func (s *MemFPSet) BeginChkptFile(fname string) error {
	path := s.chkptName(fname, "tmp")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, bucket := range s.table {
		for _, fp := range bucket {
			if err := out.WriteLong(int64(fp)); err != nil {
				_ = out.Close()
				return err
			}
		}
	}
	return out.Close()
}

func (s *MemFPSet) CommitChkpt() error {
	return s.CommitChkptFile(s.filename)
}

func (s *MemFPSet) CommitChkptFile(fname string) error {
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("MemFPSet.commitChkpt: cannot delete %s", oldChkpt)
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return fmt.Errorf("MemFPSet.commitChkpt: cannot delete %s", oldChkpt)
	}
	return nil
}

func (s *MemFPSet) Recover() error {
	return s.RecoverFile(s.filename)
}

func (s *MemFPSet) RecoverFile(fname string) error {
	file, err := os.Open(s.chkptName(fname, "chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()

	for {
		fp, err := in.ReadLong()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := s.RecoverFP(uint64(fp)); err != nil {
			return err
		}
	}
}

func (s *MemFPSet) RecoverTrace(trace *TLCTrace) error {
	return s.Recover()
}

func (s *MemFPSet) RecoverFP(fp uint64) error {
	if s.Put(fp) {
		return fmt.Errorf("fingerprint %d already in set during recovery", fp)
	}
	return nil
}

func (s *MemFPSet) UnexportObject(force bool) {}

func (s *MemFPSet) chkptName(fname string, ext string) string {
	if fname == "" {
		fname = "fpset"
	}
	return filepath.Join(s.metadir, fname+".fp."+ext)
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func absDiffUint64(a, b uint64) uint64 {
	if a >= b {
		return a - b
	}
	return b - a
}

type NoopFPSet struct {
	statesSeen uint64
	config     *FPSetConfiguration
}

func NewNoopFPSet(config *FPSetConfiguration) *NoopFPSet {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	return &NoopFPSet{config: config}
}

func (s *NoopFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	return s
}

func (s *NoopFPSet) Size() uint64                            { return 0 }
func (s *NoopFPSet) Sizeof() uint64                          { return 0 }
func (s *NoopFPSet) Put(fp uint64) bool                      { return false }
func (s *NoopFPSet) Contains(fp uint64) bool                 { return false }
func (s *NoopFPSet) CheckFPs() uint64                        { return 0 }
func (s *NoopFPSet) BeginChkpt() error                       { return nil }
func (s *NoopFPSet) BeginChkptFile(fname string) error       { return nil }
func (s *NoopFPSet) CommitChkpt() error                      { return nil }
func (s *NoopFPSet) CommitChkptFile(fname string) error      { return nil }
func (s *NoopFPSet) Recover() error                          { return nil }
func (s *NoopFPSet) RecoverFile(fname string) error          { return nil }
func (s *NoopFPSet) RecoverTrace(trace *TLCTrace) error      { return nil }
func (s *NoopFPSet) RecoverFP(fp uint64) error               { return nil }
func (s *NoopFPSet) Close()                                  {}
func (s *NoopFPSet) AddThread() error                        { return nil }
func (s *NoopFPSet) IncWorkers(num int)                      {}
func (s *NoopFPSet) Exit(cleanup bool) error                 { return nil }
func (s *NoopFPSet) CheckInvariant(expectFPs ...uint64) bool { return true }
func (s *NoopFPSet) UnexportObject(force bool)               {}

func (s *NoopFPSet) PutBlock(fpv *LongVec) *BitVector {
	if fpv == nil {
		return NewBitVector(0)
	}
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		bv.Set(i)
	}
	return bv
}

func (s *NoopFPSet) ContainsBlock(fpv *LongVec) *BitVector {
	if fpv == nil {
		return NewBitVector(0)
	}
	s.statesSeen += uint64(fpv.Size())
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		bv.Set(i)
	}
	return bv
}

func (s *NoopFPSet) GetStatesSeen() uint64 {
	return s.statesSeen
}

func (s *NoopFPSet) GetConfiguration() *FPSetConfiguration {
	if s == nil || s.config == nil {
		return NewFPSetConfiguration()
	}
	return s.config
}

type MemFPSet1 struct {
	mu         sync.Mutex
	metadir    string
	filename   string
	set        *SetOfLong
	statesSeen uint64
	config     *FPSetConfiguration
}

func NewMemFPSet1(config *FPSetConfiguration) *MemFPSet1 {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	return &MemFPSet1{
		set:    NewSetOfLong(10001),
		config: config,
	}
}

func (s *MemFPSet1) Init(numThreads int, metadir string, filename string) FPSet {
	s.metadir = metadir
	s.filename = filename
	return s
}

func (s *MemFPSet1) Size() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint64(s.set.Size())
}

func (s *MemFPSet1) Sizeof() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint64(8 + s.set.Sizeof())
}

func (s *MemFPSet1) Put(fp uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.Put(int64(fp))
}

func (s *MemFPSet1) Contains(fp uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.Contains(int64(fp))
}

func (s *MemFPSet1) PutBlock(fpv *LongVec) *BitVector {
	if fpv == nil {
		return NewBitVector(0)
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
	if fpv == nil {
		return NewBitVector(0)
	}
	size := fpv.Size()
	s.mu.Lock()
	s.statesSeen += uint64(size)
	s.mu.Unlock()
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet1) GetStatesSeen() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statesSeen
}

func (s *MemFPSet1) GetConfiguration() *FPSetConfiguration {
	if s == nil || s.config == nil {
		return NewFPSetConfiguration()
	}
	return s.config
}

func (s *MemFPSet1) Close() {}

func (s *MemFPSet1) AddThread() error {
	return nil
}

func (s *MemFPSet1) IncWorkers(num int) {}

func (s *MemFPSet1) Exit(cleanup bool) error {
	if cleanup && s.metadir != "" {
		return os.RemoveAll(s.metadir)
	}
	return nil
}

func (s *MemFPSet1) CheckInvariant(expectFPs ...uint64) bool {
	return len(expectFPs) == 0 || s.Size() == expectFPs[0]
}

func (s *MemFPSet1) CheckFPs() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint64(s.set.CheckFPs())
}

func (s *MemFPSet1) BeginChkpt() error {
	return s.BeginChkptFile(s.filename)
}

func (s *MemFPSet1) BeginChkptFile(fname string) error {
	path := s.chkptName(fname, "tmp")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.set.BeginChkpt(out); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func (s *MemFPSet1) CommitChkpt() error {
	return s.CommitChkptFile(s.filename)
}

func (s *MemFPSet1) CommitChkptFile(fname string) error {
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("MemFPSet1.commitChkpt: cannot delete %s", oldChkpt)
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return fmt.Errorf("MemFPSet1.commitChkpt: cannot delete %s", oldChkpt)
	}
	return nil
}

func (s *MemFPSet1) Recover() error {
	return s.RecoverFile(s.filename)
}

func (s *MemFPSet1) RecoverFile(fname string) error {
	file, err := os.Open(s.chkptName(fname, "chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.Recover(in)
}

func (s *MemFPSet1) RecoverTrace(trace *TLCTrace) error {
	return s.Recover()
}

func (s *MemFPSet1) RecoverFP(fp uint64) error {
	if s.Put(fp) {
		return fmt.Errorf("fingerprint %d already in set during recovery", fp)
	}
	return nil
}

func (s *MemFPSet1) UnexportObject(force bool) {}

func (s *MemFPSet1) chkptName(fname string, ext string) string {
	if fname == "" {
		fname = "fpset"
	}
	return filepath.Join(s.metadir, fname+".fp."+ext)
}

const memFPSet2LogSpineSize = 24

type MemFPSet2 struct {
	mu         sync.Mutex
	metadir    string
	filename   string
	table      [][]byte
	count      uint64
	mask       uint64
	statesSeen uint64
	config     *FPSetConfiguration
}

func NewMemFPSet2(config *FPSetConfiguration) *MemFPSet2 {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	spineSize := 1 << memFPSet2LogSpineSize
	return &MemFPSet2{
		table:  make([][]byte, spineSize),
		mask:   uint64(spineSize - 1),
		config: config,
	}
}

func (s *MemFPSet2) Init(numThreads int, metadir string, filename string) FPSet {
	s.metadir = metadir
	s.filename = filepath.Join(metadir, filename)
	return s
}

func (s *MemFPSet2) Size() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

func (s *MemFPSet2) Sizeof() uint64 {
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
	if fpv == nil {
		return NewBitVector(0)
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
	if fpv == nil {
		return NewBitVector(0)
	}
	size := fpv.Size()
	s.mu.Lock()
	s.statesSeen += uint64(size)
	s.mu.Unlock()
	bv := NewBitVector(size)
	for i := 0; i < size; i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MemFPSet2) GetStatesSeen() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statesSeen
}

func (s *MemFPSet2) GetConfiguration() *FPSetConfiguration {
	if s == nil || s.config == nil {
		return NewFPSetConfiguration()
	}
	return s.config
}

func (s *MemFPSet2) Close() {}

func (s *MemFPSet2) AddThread() error {
	return nil
}

func (s *MemFPSet2) IncWorkers(num int) {}

func (s *MemFPSet2) Exit(cleanup bool) error {
	if cleanup && s.metadir != "" {
		return os.RemoveAll(s.metadir)
	}
	return nil
}

func (s *MemFPSet2) CheckInvariant(expectFPs ...uint64) bool {
	return len(expectFPs) == 0 || s.Size() == expectFPs[0]
}

func (s *MemFPSet2) CheckFPs() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	dis := uint64(1<<63 - 1)
	for i, bucket := range s.table {
		low := uint64(i) & 0xffffff
		j := 0
		for j < len(bucket) {
			fp := memFPSet2Fingerprint(low, bucket[j], bucket[j+1], bucket[j+2], bucket[j+3], bucket[j+4])
			j += 5
			for k := j; k < len(bucket); k += 5 {
				fp1 := memFPSet2Fingerprint(low, bucket[k], bucket[k+1], bucket[k+2], bucket[k+3], bucket[k+4])
				dis = minUint64(dis, absDiffUint64(fp, fp1))
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
					dis = minUint64(dis, absDiffUint64(fp, fp1))
				}
			}
		}
	}
	return dis
}

func (s *MemFPSet2) BeginChkpt() error {
	return s.BeginChkptFile(s.filename)
}

func (s *MemFPSet2) BeginChkptFile(fname string) error {
	path := s.chkptName(fname, "tmp")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, bucket := range s.table {
		low := uint64(i) & 0xffffff
		for j := 0; j < len(bucket); j += 5 {
			fp := memFPSet2Fingerprint(low, bucket[j], bucket[j+1], bucket[j+2], bucket[j+3], bucket[j+4])
			if err := out.WriteLong(int64(fp)); err != nil {
				_ = out.Close()
				return err
			}
		}
	}
	return out.Close()
}

func (s *MemFPSet2) CommitChkpt() error {
	return s.CommitChkptFile(s.filename)
}

func (s *MemFPSet2) CommitChkptFile(fname string) error {
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("MemFPSet2.commitChkpt: cannot delete %s", oldChkpt)
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return fmt.Errorf("MemFPSet2.commitChkpt: cannot delete %s", oldChkpt)
	}
	return nil
}

func (s *MemFPSet2) Recover() error {
	return s.RecoverFile(s.filename)
}

func (s *MemFPSet2) RecoverFile(fname string) error {
	file, err := os.Open(s.chkptName(fname, "chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()
	for {
		fp, err := in.ReadLong()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := s.RecoverFP(uint64(fp)); err != nil {
			return err
		}
	}
}

func (s *MemFPSet2) RecoverTrace(trace *TLCTrace) error {
	return s.Recover()
}

func (s *MemFPSet2) RecoverFP(fp uint64) error {
	if s.Put(fp) {
		return fmt.Errorf("fingerprint %d already in set during recovery", fp)
	}
	return nil
}

func (s *MemFPSet2) UnexportObject(force bool) {}

func (s *MemFPSet2) chkptName(fname string, ext string) string {
	if fname == "" {
		fname = "fpset"
	}
	return filepath.Join(s.metadir, fname+".fp."+ext)
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
	Sets       []FPSet
	FPBits     int
	Shift      uint
	metadir    string
	filename   string
	statesSeen uint64
	config     *FPSetConfiguration
}

func NewMultiFPSet(config *FPSetConfiguration) *MultiFPSet {
	if config == nil {
		config = NewFPSetConfiguration()
	}
	bits := config.GetFPBits()
	if bits <= 0 || bits > multiFPSetMaxFPBits {
		panic("Illegal number of FPSets found.")
	}
	count := 1 << bits
	sets := make([]FPSet, count)
	childConfig := NewMultiFPSetConfiguration(config)
	for i := range sets {
		sets[i] = NewFPSet(childConfig)
	}
	return &MultiFPSet{
		Sets:   sets,
		FPBits: bits,
		Shift:  uint(64 - bits),
		config: config,
	}
}

func (s *MultiFPSet) Init(numThreads int, metadir string, filename string) FPSet {
	s.metadir = metadir
	s.filename = filename
	for i, set := range s.Sets {
		s.Sets[i] = set.Init(numThreads, metadir, fmt.Sprintf("%s_%d", filename, i))
	}
	return s
}

func (s *MultiFPSet) Size() uint64 {
	var total uint64
	for _, set := range s.Sets {
		total += set.Size()
	}
	return total
}

func (s *MultiFPSet) Sizeof() uint64 {
	var total uint64
	for _, set := range s.Sets {
		total += set.Sizeof()
	}
	return total
}

func (s *MultiFPSet) fpSet(fp uint64) FPSet {
	idx := int(fp >> s.Shift)
	return s.Sets[idx]
}

func (s *MultiFPSet) Put(fp uint64) bool {
	return s.fpSet(fp).Put(fp)
}

func (s *MultiFPSet) Contains(fp uint64) bool {
	return s.fpSet(fp).Contains(fp)
}

func (s *MultiFPSet) PutBlock(fpv *LongVec) *BitVector {
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

func (s *MultiFPSet) ContainsBlock(fpv *LongVec) *BitVector {
	if fpv == nil {
		return NewBitVector(0)
	}
	s.statesSeen += uint64(fpv.Size())
	bv := NewBitVector(fpv.Size())
	for i := 0; i < fpv.Size(); i++ {
		if !s.Contains(uint64(fpv.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *MultiFPSet) GetStatesSeen() uint64 {
	total := s.statesSeen
	for _, set := range s.Sets {
		total += set.GetStatesSeen()
	}
	return total
}

func (s *MultiFPSet) GetConfiguration() *FPSetConfiguration {
	if s == nil || s.config == nil {
		return NewFPSetConfiguration()
	}
	return s.config
}

func (s *MultiFPSet) CheckFPs() uint64 {
	dis := uint64(1<<63 - 1)
	for _, set := range s.Sets {
		dis = minUint64(dis, set.CheckFPs())
	}
	return dis
}

func (s *MultiFPSet) CheckInvariant(expectFPs ...uint64) bool {
	for _, set := range s.Sets {
		if !set.CheckInvariant() {
			return false
		}
	}
	return len(expectFPs) == 0 || s.Size() == expectFPs[0]
}

func (s *MultiFPSet) BeginChkpt() error {
	for _, set := range s.Sets {
		if err := set.BeginChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) BeginChkptFile(fname string) error {
	for i, set := range s.Sets {
		if err := set.BeginChkptFile(fmt.Sprintf("%s_%d", fname, i)); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) CommitChkpt() error {
	for _, set := range s.Sets {
		if err := set.CommitChkpt(); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) CommitChkptFile(fname string) error {
	for i, set := range s.Sets {
		if err := set.CommitChkptFile(fmt.Sprintf("%s_%d", fname, i)); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) Recover() error {
	return s.RecoverFile(s.filename)
}

func (s *MultiFPSet) RecoverFile(fname string) error {
	for i, set := range s.Sets {
		if err := set.RecoverFile(fmt.Sprintf("%s_%d", fname, i)); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) RecoverTrace(trace *TLCTrace) error {
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

func (s *MultiFPSet) RecoverFP(fp uint64) error {
	if s.Put(fp) {
		return fmt.Errorf("fingerprint %d already in set during recovery", fp)
	}
	return nil
}

func (s *MultiFPSet) Close() {
	for _, set := range s.Sets {
		set.Close()
	}
}

func (s *MultiFPSet) AddThread() error {
	for _, set := range s.Sets {
		if err := set.AddThread(); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) IncWorkers(num int) {
	for _, set := range s.Sets {
		set.IncWorkers(num)
	}
}

func (s *MultiFPSet) Exit(cleanup bool) error {
	for _, set := range s.Sets {
		if err := set.Exit(cleanup); err != nil {
			return err
		}
	}
	return nil
}

func (s *MultiFPSet) UnexportObject(force bool) {
	for _, set := range s.Sets {
		set.UnexportObject(force)
	}
}
