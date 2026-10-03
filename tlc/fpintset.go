package tlc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	fpIntSetSBits       int32 = 2
	fpIntSetSBitsMask   int32 = 0x3
	fpIntSetDoneMask    int32 = 0x1
	fpIntSetLeveledMask int32 = 0x2

	FPIntStatusNew  int32 = 0
	FPIntStatusDone int32 = 1

	memFPIntSetMaxLoad            = 20
	memFPIntSetLogInitialCapacity = 16
)

var (
	fpIntSetMu      sync.Mutex
	fpIntSetLevel   int32 = 1
	fpIntSetLeveled int32
)

// InitializeFPIntSetStatics provides the initial static field values of a
// newly loaded Java FPIntSet class. A checker does not reset these fields;
// isolated runtimes, such as the upstream per-test classloader, do.
func InitializeFPIntSetStatics() {
	fpIntSetMu.Lock()
	fpIntSetLevel = 1
	fpIntSetLeveled = 0
	fpIntSetMu.Unlock()
}

func FPIntSetIncLevel() {
	fpIntSetMu.Lock()
	fpIntSetLevel++
	fpIntSetLeveled = 2 - fpIntSetLeveled
	fpIntSetMu.Unlock()
}

func FPIntSetIsCompleted(status int32) bool {
	fpIntSetMu.Lock()
	leveled := fpIntSetLeveled
	fpIntSetMu.Unlock()
	return (status&fpIntSetLeveledMask) == leveled || (status&fpIntSetDoneMask) == FPIntStatusDone
}

func FPIntSetIsDone(status int32) bool {
	return (status & fpIntSetDoneMask) == FPIntStatusDone
}

func FPIntSetLevelOf(status int32) int32 {
	return int32(uint32(status) >> fpIntSetSBits)
}

func FPIntSetIsLeaf(status int32) bool {
	fpIntSetMu.Lock()
	level := fpIntSetLevel
	fpIntSetMu.Unlock()
	return status == FPIntStatusNew || int32(uint32(status)>>fpIntSetSBits) == level
}

func fpIntSetNewStatus(status int32) int32 {
	fpIntSetMu.Lock()
	level := fpIntSetLevel
	leveled := fpIntSetLeveled
	fpIntSetMu.Unlock()
	return (level << fpIntSetSBits) | leveled | status
}

func fpIntSetCurrentLeveled() int32 {
	fpIntSetMu.Lock()
	leveled := fpIntSetLeveled
	fpIntSetMu.Unlock()
	return leveled
}

type MemFPIntSet struct {
	mu        sync.Mutex
	metadir   string
	filename  string
	table     [][]int32
	count     uint64
	threshold uint64
	mask      uint64
}

type MultiFPIntSet struct {
	Sets   []*MemFPIntSet
	FPBits uint
}

func NewMemFPIntSet() *MemFPIntSet {
	return NewMemFPIntSetWithCapacity(memFPIntSetLogInitialCapacity, memFPIntSetMaxLoad)
}

func NewMemFPIntSetWithCapacity(logInitialCapacity int, maxLoad int) *MemFPIntSet {
	if logInitialCapacity < 0 {
		logInitialCapacity = 0
	}
	if maxLoad <= 0 {
		maxLoad = memFPIntSetMaxLoad
	}
	initialCapacity := 1 << logInitialCapacity
	return &MemFPIntSet{
		table:     make([][]int32, initialCapacity),
		threshold: uint64(initialCapacity * maxLoad),
		mask:      uint64(initialCapacity - 1),
	}
}

func NewMultiFPIntSet(bits int) *MultiFPIntSet {
	if bits < 0 {
		bits = 0
	}
	count := 1 << bits
	sets := make([]*MemFPIntSet, count)
	for i := range sets {
		sets[i] = NewMemFPIntSet()
	}
	return &MultiFPIntSet{Sets: sets, FPBits: uint(64 - bits)}
}

func (s *MemFPIntSet) Init(numThreads int, metadir string, filename string) *MemFPIntSet {
	_ = numThreads
	s.metadir = metadir
	s.filename = filename
	return s
}

func (s *MultiFPIntSet) Init(numThreads int, metadir string, filename string) *MultiFPIntSet {
	if s == nil {
		return nil
	}
	for i, set := range s.Sets {
		if set != nil {
			set.Init(numThreads, metadir, fmt.Sprintf("%s_%d", filename, i))
		}
	}
	return s
}

func (s *MemFPIntSet) Size() uint64 {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

func (s *MultiFPIntSet) Size() uint64 {
	if s == nil {
		return 0
	}
	var total uint64
	for _, set := range s.Sets {
		if set != nil {
			total += set.Size()
		}
	}
	return total
}

func (s *MemFPIntSet) Sizeof() uint64 {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	size := uint64(28)
	size += 16 + uint64(len(s.table))*8
	for _, bucket := range s.table {
		if bucket != nil {
			size += 16 + uint64(len(bucket))*4
		}
	}
	return size
}

func (s *MultiFPIntSet) Sizeof() uint64 {
	if s == nil {
		return 0
	}
	var total uint64
	for _, set := range s.Sets {
		if set != nil {
			total += set.Sizeof()
		}
	}
	return total
}

func (s *MemFPIntSet) SetLeveled(fp uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := fp & s.mask
	bucket := s.table[index]
	hi, lo := splitFingerprint(fp)
	for i := 0; i < len(bucket); i += 3 {
		if bucket[i] == hi && bucket[i+1] == lo {
			bucket[i+2] = (bucket[i+2] &^ fpIntSetLeveledMask) | fpIntSetCurrentLeveled()
			return
		}
	}
	panic("MemFPIntSet.SetLeveled: fingerprint must already be in the set")
}

func (s *MultiFPIntSet) SetLeveled(fp uint64) {
	if set := s.fpSet(fp); set != nil {
		set.SetLeveled(fp)
	}
}

func (s *MemFPIntSet) SetStatus(fp uint64, status int32) int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := fp & s.mask
	bucket := s.table[index]
	hi, lo := splitFingerprint(fp)
	for i := 0; i < len(bucket); i += 3 {
		if bucket[i] == hi && bucket[i+1] == lo {
			oldStatus := bucket[i+2]
			bucket[i+2] = oldStatus | status
			return oldStatus
		}
	}
	if s.count >= s.threshold {
		s.rehash()
		index = fp & s.mask
		bucket = s.table[index]
	}
	s.table[index] = append(bucket, hi, lo, fpIntSetNewStatus(status))
	s.count++
	return FPIntStatusNew
}

func (s *MultiFPIntSet) SetStatus(fp uint64, status int32) int32 {
	if set := s.fpSet(fp); set != nil {
		return set.SetStatus(fp, status)
	}
	return FPIntStatusNew
}

func (s *MemFPIntSet) GetStatus(fp uint64) int32 {
	if s == nil {
		return FPIntStatusNew
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	index := fp & s.mask
	bucket := s.table[index]
	hi, lo := splitFingerprint(fp)
	for i := 0; i < len(bucket); i += 3 {
		if bucket[i] == hi && bucket[i+1] == lo {
			return bucket[i+2]
		}
	}
	return FPIntStatusNew
}

func (s *MultiFPIntSet) GetStatus(fp uint64) int32 {
	if set := s.fpSet(fp); set != nil {
		return set.GetStatus(fp)
	}
	return FPIntStatusNew
}

func (s *MemFPIntSet) AllLeveled() bool {
	if s == nil {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	leveled := fpIntSetCurrentLeveled()
	for _, bucket := range s.table {
		for i := 0; i < len(bucket); i += 3 {
			if bucket[i+2]&fpIntSetLeveledMask != leveled {
				return false
			}
		}
	}
	return true
}

func (s *MultiFPIntSet) AllLeveled() bool {
	if s == nil {
		return true
	}
	for _, set := range s.Sets {
		if set != nil && !set.AllLeveled() {
			return false
		}
	}
	return true
}

func (s *MemFPIntSet) Close() {}

func (s *MultiFPIntSet) Close() {
	if s == nil {
		return
	}
	for _, set := range s.Sets {
		if set != nil {
			set.Close()
		}
	}
}

func (s *MemFPIntSet) AddThread() error {
	return nil
}

func (s *MultiFPIntSet) AddThread() error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.AddThread(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MemFPIntSet) Exit(cleanup bool) error {
	if cleanup && s.metadir != "" {
		return os.RemoveAll(s.metadir)
	}
	return nil
}

func (s *MultiFPIntSet) Exit(cleanup bool) error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.Exit(cleanup); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MemFPIntSet) CheckFPs() uint64 {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	dis := uint64(1<<63 - 1)
	for i, bucket := range s.table {
		for j := 0; j < len(bucket); j += 3 {
			x := joinFingerprintForMemFPIntSetCheckFPs(bucket[j], bucket[j+1])
			for k := j + 3; k < len(bucket); k += 3 {
				y := joinFingerprintForMemFPIntSetCheckFPs(bucket[k], bucket[k+1])
				dis = minUint64(dis, javaSignedAbsDiffAsUint64(x, y))
			}
			for _, otherBucket := range s.table[i+1:] {
				for k := 0; k < len(otherBucket); k += 3 {
					y := joinFingerprintForMemFPIntSetCheckFPs(otherBucket[k], otherBucket[k+1])
					dis1 := javaSignedDiffAsInt64(x, y)
					if dis1 >= 0 {
						dis = minUint64(dis, uint64(dis1))
					}
				}
			}
		}
	}
	return dis
}

func (s *MultiFPIntSet) CheckFPs() uint64 {
	if s == nil {
		return 0
	}
	var maxDistance uint64
	for _, set := range s.Sets {
		if set != nil {
			if distance := set.CheckFPs(); distance > maxDistance {
				maxDistance = distance
			}
		}
	}
	return maxDistance
}

func (s *MemFPIntSet) BeginChkpt() error {
	return s.BeginChkptFile(s.filename)
}

func (s *MultiFPIntSet) BeginChkpt() error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.BeginChkpt(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MemFPIntSet) BeginChkptFile(fname string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.beginChkptFileLocked(fname)
}

func (s *MemFPIntSet) beginChkptLocked() error {
	return s.beginChkptFileLocked(s.filename)
}

func (s *MemFPIntSet) beginChkptFileLocked(fname string) error {
	path := s.chkptName(fname, "tmp")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	for _, bucket := range s.table {
		for _, word := range bucket {
			if err := out.WriteInt(word); err != nil {
				_ = out.Close()
				return err
			}
		}
	}
	return out.Close()
}

func (s *MultiFPIntSet) BeginChkptFile(fname string) error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.BeginChkptFile(fname); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MemFPIntSet) CommitChkpt() error {
	return s.CommitChkptFile(s.filename)
}

func (s *MultiFPIntSet) CommitChkpt() error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.CommitChkpt(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MemFPIntSet) CommitChkptFile(fname string) error {
	oldChkpt := s.chkptName(fname, "chkpt")
	newChkpt := s.chkptName(fname, "tmp")
	if err := os.Remove(oldChkpt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("MemFPIntSet.commitChkpt: cannot delete %s", oldChkpt)
	}
	if err := os.Rename(newChkpt, oldChkpt); err != nil {
		return fmt.Errorf("MemFPIntSet.commitChkpt: cannot delete %s", oldChkpt)
	}
	return nil
}

func (s *MultiFPIntSet) CommitChkptFile(fname string) error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.CommitChkptFile(fname); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MemFPIntSet) Recover() error {
	return s.RecoverFile(s.filename)
}

func (s *MultiFPIntSet) Recover() error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.Recover(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MemFPIntSet) RecoverFile(fname string) error {
	file, err := os.Open(s.chkptName(fname, "chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		hi, err := in.ReadInt()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		lo, err := in.ReadInt()
		if err != nil {
			return err
		}
		status, err := in.ReadInt()
		if err != nil {
			return err
		}
		if s.count >= s.threshold {
			s.rehash()
		}
		fp := joinFingerprint(hi, lo)
		index := fp & s.mask
		s.table[index] = append(s.table[index], hi, lo, status)
		s.count++
	}
}

func (s *MultiFPIntSet) RecoverFile(fname string) error {
	if s == nil {
		return nil
	}
	for _, set := range s.Sets {
		if set != nil {
			if err := set.RecoverFile(fname); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MultiFPIntSet) fpSet(fp uint64) *MemFPIntSet {
	if s == nil || len(s.Sets) == 0 {
		return nil
	}
	idx := int(fp >> s.FPBits)
	if idx >= len(s.Sets) {
		idx %= len(s.Sets)
	}
	return s.Sets[idx]
}

func (s *MemFPIntSet) rehash() {
	oldTable := s.table
	oldCapacity := len(oldTable)
	newTable := make([][]int32, oldCapacity*2)
	oneBitMask := int32(oldCapacity)
	for i, bucket := range oldTable {
		if bucket == nil {
			continue
		}
		cnt0, cnt1 := 0, 0
		for j := 0; j < len(bucket); j += 3 {
			if bucket[j+1]&oneBitMask == 0 {
				cnt0 += 3
			} else {
				cnt1 += 3
			}
		}
		if cnt0 == 0 {
			newTable[i+oldCapacity] = bucket
			continue
		}
		if cnt1 == 0 {
			newTable[i] = bucket
			continue
		}
		list0 := make([]int32, cnt0)
		list1 := make([]int32, cnt1)
		for j := 0; j < len(bucket); j += 3 {
			if bucket[j+1]&oneBitMask == 0 {
				list0[cnt0-3] = bucket[j]
				list0[cnt0-2] = bucket[j+1]
				list0[cnt0-1] = bucket[j+2]
				cnt0 -= 3
			} else {
				list1[cnt1-3] = bucket[j]
				list1[cnt1-2] = bucket[j+1]
				list1[cnt1-1] = bucket[j+2]
				cnt1 -= 3
			}
		}
		newTable[i] = list0
		newTable[i+oldCapacity] = list1
	}
	s.table = newTable
	s.threshold *= 2
	s.mask = uint64(len(newTable) - 1)
}

func (s *MemFPIntSet) chkptName(fname string, ext string) string {
	if fname == "" {
		fname = "fpset"
	}
	return filepath.Join(s.metadir, fname+".fp."+ext)
}

func splitFingerprint(fp uint64) (int32, int32) {
	return int32(fp >> 32), int32(fp & 0xffffffff)
}

func joinFingerprint(hi int32, lo int32) uint64 {
	return (uint64(uint32(hi)) << 32) | uint64(uint32(lo))
}

func joinFingerprintForMemFPIntSetCheckFPs(hi int32, lo int32) int64 {
	return (int64(hi) << 32) | int64(lo)
}

func javaSignedDiffAsInt64(x int64, y int64) int64 {
	if x > y {
		return x - y
	}
	return y - x
}

func javaSignedAbsDiffAsUint64(x int64, y int64) uint64 {
	dis := javaSignedDiffAsInt64(x, y)
	if dis < 0 {
		dis = -dis
	}
	return uint64(dis)
}
