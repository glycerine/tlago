// Java test sources: tlc2/tool/distributed/fp/DynamicFPSetManagerTest.java,
// FPSetManagerTest.java, and FaultyFPSet.java (Microsoft, MIT license).
package tlc

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

func requireJavaIllegalArgument(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		failure := recover()
		if _, ok := failure.(*IllegalArgumentException); !ok {
			t.Fatalf("failure = %T (%v), want IllegalArgumentException", failure, failure)
		}
	}()
	call()
}

func javaDynamicFPManager(t *testing.T, count int) *DistributedFPSetManager {
	t.Helper()
	m := NewDynamicDistributedFPSetManager(count)
	for i := range count {
		if err := m.RegisterFPSet(NewLocalFingerprintEndpoint(NewMemFPSet()), fmt.Sprintf("localhost%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	return m
}

func TestJavaDynamicFPSetManagerConstructors(t *testing.T) {
	for _, invalid := range []int{0, -1} {
		t.Run(fmt.Sprintf("Invalid%d", invalid), func(t *testing.T) {
			requireJavaIllegalArgument(t, func() { NewDynamicDistributedFPSetManager(invalid) })
		})
	}
	for _, test := range []struct {
		count int
		mask  uint64
	}{{1, 1}, {10, 15}, {31, 31}, {32, 63}, {33, 63}, {math.MaxInt32, math.MaxInt32}} {
		t.Run(fmt.Sprint(test.count), func(t *testing.T) {
			if got := NewDynamicDistributedFPSetManager(test.count).GetMask(); got != test.mask {
				t.Fatalf("mask = %d, want %d", got, test.mask)
			}
		})
	}
}

func TestJavaDynamicFPSetManagerIndices(t *testing.T) {
	t.Run("SingleFPSet", func(t *testing.T) {
		m := javaDynamicFPManager(t, 1)
		for _, fp := range []uint64{math.MaxInt64, 1 << 63, 0, 1} {
			if got := m.GetFPSetIndex(fp); got != 0 {
				t.Fatalf("index(%d) = %d", fp, got)
			}
		}
	})
	t.Run("10FPSets", func(t *testing.T) {
		m := javaDynamicFPManager(t, 10)
		pairs := []struct {
			fp    uint64
			index int
		}{{math.MaxInt64, 5}, {1 << 63, 0}, {0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}, {10, 0}, {11, 1}, {12, 2}, {48, 0}, {49, 1}, {50, 2}, {51, 3}}
		for _, pair := range pairs {
			if got := m.GetFPSetIndex(pair.fp); got != pair.index {
				t.Fatalf("index(%d) = %d, want %d", pair.fp, got, pair.index)
			}
		}
	})
}

func TestJavaDynamicFPSetManagerReassign(t *testing.T) {
	for _, invalid := range []int{-1, 1} {
		t.Run(fmt.Sprintf("Invalid%d", invalid), func(t *testing.T) {
			m := javaDynamicFPManager(t, 1)
			requireJavaIllegalArgument(t, func() { m.Reassign(invalid) })
		})
	}
	t.Run("Terminate", func(t *testing.T) {
		if got := javaDynamicFPManager(t, 1).Reassign(0); got != -1 {
			t.Fatalf("reassign = %d", got)
		}
	})
	t.Run("Successors", func(t *testing.T) {
		m := javaDynamicFPManager(t, 10)
		for index := 1; index <= 9; index++ {
			if got, want := m.Reassign(index), (index+1)%10; got != want {
				t.Fatalf("reassign(%d) = %d, want %d", index, got, want)
			}
		}
		if got := m.Reassign(0); got != -1 {
			t.Fatalf("last reassign = %d", got)
		}
	})
}

// Embedding alone does not preserve Java's virtual dispatch from MemFPSet's
// block methods. The block overrides below call these scalar overrides too.
type javaFaultyFPSet struct {
	*MemFPSet
	mu                  sync.Mutex
	putInvocations      int
	containsInvocations int
}

func newJavaFaultyFPSet() *javaFaultyFPSet { return &javaFaultyFPSet{MemFPSet: NewMemFPSet()} }

func (s *javaFaultyFPSet) Put(fp uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.putInvocations
	s.putInvocations++
	if old > 0 {
		panic(NewStatefulRuntimeException("Test FPSet"))
	}
	return s.MemFPSet.Put(fp)
}

func (s *javaFaultyFPSet) Contains(fp uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.containsInvocations
	s.containsInvocations++
	if old > 0 {
		panic(NewStatefulRuntimeException("Test FPSet"))
	}
	return s.MemFPSet.Contains(fp)
}

func (s *javaFaultyFPSet) PutBlock(fps *LongVec) *BitVector {
	bv := NewBitVector(fps.Size())
	for i := 0; i < fps.Size(); i++ {
		if !s.Put(uint64(fps.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func (s *javaFaultyFPSet) ContainsBlock(fps *LongVec) *BitVector {
	s.MemFPSet.mu.Lock()
	s.MemFPSet.statesSeen += uint64(fps.Size())
	s.MemFPSet.mu.Unlock()
	bv := NewBitVector(fps.Size())
	for i := 0; i < fps.Size(); i++ {
		if !s.Contains(uint64(fps.ElementAt(i))) {
			bv.Set(i)
		}
	}
	return bv
}

func javaFailoverManager(t *testing.T, secondFaulty bool) *DistributedFPSetManager {
	t.Helper()
	m := NewDynamicDistributedFPSetManager(2)
	if err := m.RegisterFPSet(NewLocalFingerprintEndpoint(newJavaFaultyFPSet()), "TestFPSet1"); err != nil {
		t.Fatal(err)
	}
	var second FPSet = NewMemFPSet()
	if secondFaulty {
		second = newJavaFaultyFPSet()
	}
	if err := m.RegisterFPSet(NewLocalFingerprintEndpoint(second), "TestFPSet2"); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestJavaDynamicFPSetManagerFailoverPut(t *testing.T) {
	m := javaFailoverManager(t, false)
	if got := m.GetFPSetIndex(2); got != 0 {
		t.Fatalf("index = %d", got)
	}
	for i := 0; i < 2; i++ {
		if m.Put(2) {
			t.Fatal("put reported fingerprint already known")
		}
		if !m.Contains(2) {
			t.Fatal("contains did not find fingerprint")
		}
	}
}

func requireJavaBitCounts(t *testing.T, vectors []*BitVector, counts ...int) {
	t.Helper()
	if len(vectors) != len(counts) {
		t.Fatalf("vector count = %d, want %d", len(vectors), len(counts))
	}
	for i, want := range counts {
		if vectors[i] == nil {
			t.Fatalf("partition %d is null", i)
		}
		if got := vectors[i].TrueCount(); got != want {
			t.Fatalf("partition %d trueCnt = %d, want %d", i, got, want)
		}
	}
}

func TestJavaDynamicFPSetManagerFailoverBlocks(t *testing.T) {
	for _, test := range []struct {
		name                     string
		secondFaulty, concurrent bool
	}{{"PutBlock", false, false}, {"TerminationPutBlock", true, false}, {"TerminationPutBlockConcurrent", true, true}} {
		t.Run(test.name, func(t *testing.T) {
			m := javaFailoverManager(t, test.secondFaulty)
			fps := []*LongVec{NewLongVec(), NewLongVec()}
			for i := range fps {
				fps[i].AddElement(int64(i))
				if got := m.GetFPSetIndex(uint64(i)); got != i {
					t.Fatalf("index = %d, want %d", got, i)
				}
			}
			var executor []*DistributedExecutor
			if test.concurrent {
				es := NewDistributedExecutor()
				defer es.Shutdown()
				executor = append(executor, es)
			}
			requireJavaBitCounts(t, m.PutBlock(fps, executor...), 1, 1)
			requireJavaBitCounts(t, m.ContainsBlock(fps, executor...), 0, 0)
			if test.secondFaulty {
				// Java's BitVector(size,true) initializes the closed range
				// [0,wordCount], so a one-element fallback has two set bits.
				requireJavaBitCounts(t, m.PutBlock(fps, executor...), 2, 2)
				if alive := m.NumOfAliveServers(); alive != 0 {
					t.Fatalf("alive = %d, want 0", alive)
				}
				requireJavaBitCounts(t, m.ContainsBlock(fps, executor...), 2, 2)
			} else {
				requireJavaBitCounts(t, m.PutBlock(fps), 1, 0)
				if alive := m.NumOfAliveServers(); alive != 1 {
					t.Fatalf("alive = %d, want 1", alive)
				}
				requireJavaBitCounts(t, m.ContainsBlock(fps), 0, 0)
			}
		})
	}
}

func TestJavaDynamicFPSetManagerPutBlockConcurrentOrder(t *testing.T) {
	const count = 20
	m := NewDynamicDistributedFPSetManager(count)
	for i := range count {
		set := NewMemFPSet()
		if i == count-1 {
			set.Put(1)
		}
		if err := m.RegisterFPSet(NewLocalFingerprintEndpoint(set), fmt.Sprintf("TestFPSet%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	fps := make([]*LongVec, count)
	want := make([]int, count)
	for i := range count {
		fps[i] = NewLongVec()
		fps[i].AddElement(1)
		if i != count-1 {
			want[i] = 1
		}
	}
	executor := NewDistributedExecutor()
	defer executor.Shutdown()
	requireJavaBitCounts(t, m.ContainsBlock(fps, executor), want...)
}

func TestJavaFPSetManagerNestedFingerprintPartitions(t *testing.T) {
	for _, expected := range []int{2, 3, 4, 5, 8} {
		t.Run(fmt.Sprint(expected), func(t *testing.T) {
			config := NewFPSetConfiguration()
			config.SetFPBits(1)
			// Bound the Go equivalent of the Java test JVM's fingerprint
			// heap budget, retaining its production factory and nesting.
			config.SetMemory(64 * 1024 * 1024)
			m := NewDynamicDistributedFPSetManager(expected)
			defer m.Close(true)
			for i := 0; i < config.GetMultiFPSetCnt(); i++ {
				set := NewFPSet(config)
				set.Init(1, t.TempDir(), fmt.Sprintf("test%d", expected))
				if err := m.RegisterFPSet(NewLocalFingerprintEndpoint(set), fmt.Sprintf("localhost%d", i)); err != nil {
					t.Fatal(err)
				}
			}
			for _, pair := range []struct {
				fp    uint64
				index int
			}{{0, 0}, {1, 1}, {2, 0}, {3, 1}, {1<<63 | 2, 0}, {1<<63 | 1, 1}} {
				if got := m.GetFPSetIndex(pair.fp); got != pair.index {
					t.Fatalf("index(%d) = %d, want %d", pair.fp, got, pair.index)
				}
			}
			var fps []uint64
			for low := uint64(1); low <= 8; low++ {
				for high := uint64(0); high < 4; high++ {
					fps = append(fps, high<<62|low)
				}
			}
			for i, fp := range fps {
				for _, unseen := range fps[i:] {
					if m.Contains(unseen) {
						t.Fatalf("unseen fingerprint %064b found", unseen)
					}
				}
				if m.Put(fp) {
					t.Fatalf("put(%064b) already known", fp)
				}
				if !m.Contains(fp) {
					t.Fatalf("contains(%064b) = false", fp)
				}
			}
			if got := m.Size(); got != uint64(len(fps)) {
				t.Fatalf("size = %d, want %d", got, len(fps))
			}
			if !m.CheckInvariant() {
				t.Fatal("invariant failed")
			}
		})
	}
}
