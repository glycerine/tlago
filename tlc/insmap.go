package tlc

import (
	"fmt"
	"iter"
	"reflect"
	"sync/atomic"

	rb "github.com/glycerine/rbtree"
)

// InsMap is a key-value dictionary like the built
// in Go map, except that we iterate in insertion order
// when using All() to range.
//
// InsMap is not goroutine safe on its own.
// Users must supply their own synchronization if the
// InsMap is reachable from multiple goroutines.
type InsMap[K comparable, V any] struct {
	version int64

	tree *rb.Tree
	idx  map[K]rb.Iterator

	// ordercache caches the pointers in contiguous memory.
	ordercache   []*IKV[K, V]
	cacheversion int64

	// nextInsertSeq ensures we sort by insertion time.
	nextInsertSeq uint64
}

type IKV[K comparable, V any] struct {
	seq uint64 // insertion order sequence
	key K
	val V
}

// NewInsMap makes a new InsMap.
func NewInsMap[K comparable, V any]() *InsMap[K, V] {
	return &InsMap[K, V]{
		idx: make(map[K]rb.Iterator),
		tree: rb.NewTree(func(a, b rb.Item) int {
			aseq := a.(*IKV[K, V]).seq
			bseq := b.(*IKV[K, V]).seq
			if aseq < bseq {
				return -1
			}
			if aseq > bseq {
				return 1
			}
			return 0
		}),
	}
}

// Cached returns the raw internal IKV slice.
func (s *InsMap[K, V]) Cached() []*IKV[K, V] {
	n := s.tree.Len()
	nc := len(s.ordercache)
	vers := atomic.LoadInt64(&s.version)
	if nc == n && s.cacheversion == vers {
		return s.ordercache
	}
	s.ordercache = nil
	s.cacheversion = vers
	for it := s.tree.Min(); !it.Limit(); it = it.Next() {
		kv := it.Item().(*IKV[K, V])
		s.ordercache = append(s.ordercache, kv)
	}
	return s.ordercache
}

func (s *InsMap[K, V]) Len() int {
	return len(s.idx)
}

func (s *InsMap[K, V]) String() string {
	r := "InsMap{"
	for k, v := range s.All() {
		r += fmt.Sprintf("%v:%v, ", k, v)
	}
	return r + "}"
}

// Delkey deletes a key from the InsMap and returns the next iterator in the tree.
func (s *InsMap[K, V]) Delkey(key K) (found bool, next rb.Iterator) {
	if s.idx == nil || isNil(key) {
		return false, s.tree.Limit()
	}
	it, ok := s.idx[key]
	if !ok {
		return false, s.tree.Limit()
	}
	found = true
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0
	next = it.Next()
	s.tree.DeleteWithIterator(it)
	delete(s.idx, key)
	return
}

// DeleteWithIter deletes by iterator and returns the next iterator.
func (s *InsMap[K, V]) DeleteWithIter(it rb.Iterator) (found bool, next rb.Iterator) {
	if it.Limit() || s.idx == nil {
		return false, s.tree.Limit()
	}
	kv := it.Item().(*IKV[K, V])
	if _, ok := s.idx[kv.key]; !ok {
		return false, s.tree.Limit()
	}
	found = true
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0
	next = it.Next()
	s.tree.DeleteWithIterator(it)
	delete(s.idx, kv.key)
	return
}

func (s *InsMap[K, V]) DeleteAll() {
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0
	s.tree.DeleteAll()
	s.idx = make(map[K]rb.Iterator)
	s.nextInsertSeq = 0
}

func (s *InsMap[K, V]) Set(key K, val V) (newlyAdded bool) {
	if isNil(key) {
		return false
	}
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0

	if s.idx == nil {
		s.idx = make(map[K]rb.Iterator)
	}
	it, ok := s.idx[key]

	if !ok {
		newlyAdded = true
		seq := s.nextInsertSeq
		s.nextInsertSeq++
		item := &IKV[K, V]{seq: seq, key: key, val: val}
		_, it2 := s.tree.InsertGetIt(item)
		s.idx[key] = it2
		return
	}
	it.Item().(*IKV[K, V]).val = val
	return
}

// all iterates in insertion order.
func (s *InsMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if s == nil || s.tree == nil {
			return
		}

		vers := atomic.LoadInt64(&s.version)
		if len(s.ordercache) == s.tree.Len() && s.cacheversion == vers {
			for _, kv := range s.ordercache {
				lastSeq := kv.seq
				if !yield(kv.key, kv.val) {
					return
				}
				if atomic.LoadInt64(&s.version) != vers {
					s.resumeFrom(lastSeq+1, yield)
					return
				}
			}
			return
		}

		it := s.tree.Min()
		for !it.Limit() {
			kv := it.Item().(*IKV[K, V])
			lastSeq := kv.seq
			if !yield(kv.key, kv.val) {
				return
			}
			if atomic.LoadInt64(&s.version) != vers {
				vers = atomic.LoadInt64(&s.version)
				it = s.tree.FindGE(&IKV[K, V]{seq: lastSeq + 1})
			} else {
				it = it.Next()
			}
		}
	}
}

// resumeFrom handles the re-sync logic after a mutation.
func (s *InsMap[K, V]) resumeFrom(nextSeq uint64, yield func(K, V) bool) {
	it := s.tree.FindGE(&IKV[K, V]{seq: nextSeq})
	for !it.Limit() {
		kv := it.Item().(*IKV[K, V])
		lastSeq := kv.seq
		vers := atomic.LoadInt64(&s.version)
		if !yield(kv.key, kv.val) {
			return
		}
		if atomic.LoadInt64(&s.version) != vers {
			it = s.tree.FindGE(&IKV[K, V]{seq: lastSeq + 1})
		} else {
			it = it.Next()
		}
	}
}

func (s *InsMap[K, V]) AllIKV() iter.Seq2[K, *IKV[K, V]] {
	return func(yield func(K, *IKV[K, V]) bool) {
		it := s.tree.Min()
		for !it.Limit() {
			kv := it.Item().(*IKV[K, V])
			lastSeq := kv.seq
			vers := atomic.LoadInt64(&s.version)
			if !yield(kv.key, kv) {
				return
			}
			if atomic.LoadInt64(&s.version) != vers {
				it = s.tree.FindGE(&IKV[K, V]{seq: lastSeq + 1})
			} else {
				it = it.Next()
			}
		}
	}
}

func (s *InsMap[K, V]) Get2(key K) (val V, found bool) {
	if s.idx == nil || isNil(key) {
		return
	}
	if it, ok := s.idx[key]; ok {
		return it.Item().(*IKV[K, V]).val, true
	}
	return
}

func (s *InsMap[K, V]) Get(key K) (val V) {
	v, _ := s.Get2(key)
	return v
}

func (s *InsMap[K, V]) GetIKV(key K) (kv *IKV[K, V], found bool) {
	if s.idx == nil || isNil(key) {
		return
	}
	if it, ok := s.idx[key]; ok {
		return it.Item().(*IKV[K, V]), true
	}
	return
}

// IsNil uses reflect to to return true iff the face
// contains a nil pointer, map, array, slice, or channel.
func isNil(face interface{}) bool {
	if face == nil {
		return true
	}
	switch reflect.TypeOf(face).Kind() {
	case reflect.Ptr, reflect.Array, reflect.Map, reflect.Slice, reflect.Chan:
		return reflect.ValueOf(face).IsNil()
	}
	return false
}
