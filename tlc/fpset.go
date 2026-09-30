package tlc

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
)

const (
	fpSetLongSize       = 8
	multiFPSetMaxFPBits = 30
	multiFPSetMinFPBits = 0
)

type FPSetConfiguration struct {
	FPBits         int
	MemoryInBytes  int64
	Ratio          float64
	Implementation string
}

func NewFPSetConfiguration() *FPSetConfiguration {
	return NewFPSetConfigurationWithRatio(0.25)
}

func NewFPSetConfigurationWithRatio(ratio float64) *FPSetConfiguration {
	return &FPSetConfiguration{
		FPBits:         1,
		MemoryInBytes:  -1,
		Ratio:          ratio,
		Implementation: "tlc2.tool.fp.MSBDiskFPSet",
	}
}

func (c *FPSetConfiguration) AllowsNesting() bool {
	return c.GetFPBits() > 0
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
	if c.MemoryInBytes > 0 {
		if c.Ratio > 0 {
			return int64(float64(c.MemoryInBytes) * c.Ratio)
		}
		return c.MemoryInBytes
	}
	return c.MemoryInBytes
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

func IsValidFPBits(fpBits int) bool {
	return fpBits >= multiFPSetMinFPBits && fpBits <= multiFPSetMaxFPBits
}

func isDiskFPSetImplementation(implementation string) bool {
	switch implementation {
	case "tlc2.tool.fp.DiskFPSet",
		"tlc2.tool.fp.HeapBasedDiskFPSet",
		"tlc2.tool.fp.LSBDiskFPSet",
		"tlc2.tool.fp.MSBDiskFPSet",
		"tlc2.tool.fp.OffHeapDiskFPSet":
		return true
	default:
		return false
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

func (s *MemFPSet) Init(numThreads int, metadir string, filename string) *MemFPSet {
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
		return fmt.Errorf("MemFPSet.CommitChkpt: cannot delete %s: %w", oldChkpt, err)
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return fmt.Errorf("MemFPSet.CommitChkpt: cannot rename %s to %s: %w", newChkpt, oldChkpt, err)
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

func (s *MemFPSet) RecoverTrace(trace *MemoryTrace) error {
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

func (s *NoopFPSet) Init(numThreads int, metadir string, filename string) *NoopFPSet {
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
func (s *NoopFPSet) RecoverTrace(trace *MemoryTrace) error   { return nil }
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

type MemFPSet1 struct{ *MemFPSet }

func NewMemFPSet1(config *FPSetConfiguration) *MemFPSet1 {
	return &MemFPSet1{MemFPSet: NewMemFPSetWithConfig(config)}
}

type MemFPSet2 struct{ *MemFPSet }

func NewMemFPSet2(config *FPSetConfiguration) *MemFPSet2 {
	return &MemFPSet2{MemFPSet: NewMemFPSetWithConfig(config)}
}

type MultiFPSet struct {
	Sets       []*MemFPSet
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
	sets := make([]*MemFPSet, count)
	childConfig := *config
	childConfig.FPBits = 0
	for i := range sets {
		sets[i] = NewMemFPSetWithConfig(&childConfig)
	}
	return &MultiFPSet{
		Sets:   sets,
		FPBits: bits,
		Shift:  uint(64 - bits),
		config: config,
	}
}

func (s *MultiFPSet) Init(numThreads int, metadir string, filename string) *MultiFPSet {
	s.metadir = metadir
	s.filename = filename
	for i, set := range s.Sets {
		set.Init(numThreads, metadir, fmt.Sprintf("%s_%d", filename, i))
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

func (s *MultiFPSet) fpSet(fp uint64) *MemFPSet {
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
	return s.BeginChkptFile(s.filename)
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
	return s.CommitChkptFile(s.filename)
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

func (s *MultiFPSet) RecoverTrace(trace *MemoryTrace) error {
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
		if err := set.Exit(false); err != nil {
			return err
		}
	}
	if cleanup && s.metadir != "" {
		return os.RemoveAll(s.metadir)
	}
	return nil
}

func (s *MultiFPSet) UnexportObject(force bool) {
	for _, set := range s.Sets {
		set.UnexportObject(force)
	}
}
