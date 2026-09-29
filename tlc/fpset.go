package tlc

import "sync"

type MemFPSet struct {
	mu        sync.Mutex
	metadir   string
	filename  string
	table     [][]uint64
	count     uint64
	threshold uint64
	mask      uint64
}

const (
	memFPSetMaxLoad            = 20
	memFPSetLogInitialCapacity = 16
)

func NewMemFPSet() *MemFPSet {
	initialCapacity := 1 << memFPSetLogInitialCapacity
	return &MemFPSet{
		table:     make([][]uint64, initialCapacity),
		threshold: uint64(initialCapacity * memFPSetMaxLoad),
		mask:      uint64(initialCapacity - 1),
	}
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

	dis := ^uint64(0)
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
