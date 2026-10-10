// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Derived from the util.Set implementation in the pinned TLC source.
package tlc

// permutationSet ports the util.Set used by MVPerms.permutationSubgroup.
// Bucket chains retain source equality order, including after a rehash.
type permutationSet struct {
	buckets   []*permutationSetEntry
	count     int
	threshold int
}

type permutationSetEntry struct {
	hash int32
	key  *MVPerm
	next *permutationSetEntry
}

func newPermutationSet(capacity int) *permutationSet {
	return &permutationSet{
		buckets:   make([]*permutationSetEntry, capacity),
		threshold: permutationSetThreshold(capacity),
	}
}

func permutationSetThreshold(capacity int) int {
	return int(javaDoubleToInt(float64(float32(capacity) * 0.75)))
}

func (s *permutationSet) put(key *MVPerm) *MVPerm {
	hash := key.HashCode()
	index := int(uint32(hash)&0x7fffffff) % len(s.buckets)
	for entry := s.buckets[index]; entry != nil; entry = entry.next {
		if entry.hash == hash && entry.key.Equal(key) {
			old := entry.key
			entry.key = key
			return old
		}
	}
	if s.count >= s.threshold {
		s.rehash()
		return s.put(key)
	}
	s.buckets[index] = &permutationSetEntry{hash: hash, key: key, next: s.buckets[index]}
	s.count++
	return nil
}

func (s *permutationSet) rehash() {
	old := s.buckets
	capacity := valueStreamArrayLength(int32(len(old))*2 + 1)
	s.buckets = make([]*permutationSetEntry, capacity)
	s.threshold = permutationSetThreshold(capacity)
	for i := len(old) - 1; i >= 0; i-- {
		for entry := old[i]; entry != nil; {
			next := entry.next
			index := int(uint32(entry.hash)&0x7fffffff) % capacity
			entry.next = s.buckets[index]
			s.buckets[index] = entry
			entry = next
		}
	}
}
