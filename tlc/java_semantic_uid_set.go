package tlc

import "reflect"

// JavaSemanticUIDSet preserves HashSet membership and iteration for nodes of
// one concrete semantic class/kind whose equality uses UID and whose hashCode includes kind and UID.
// UID hashes are injective, so equal hashes cannot require an identity tie-break.
type JavaSemanticUIDSet struct {
	entries *javaHashMap[int32, struct{}]
}

func NewJavaSemanticUIDSet(kind int32) *JavaSemanticUIDSet {
	return &JavaSemanticUIDSet{entries: newJavaHashMap[int32, struct{}](func(uid int32) int32 { return 31*(31+kind) + uid }, nil)}
}

func (s *JavaSemanticUIDSet) Add(uid int32) bool {
	if _, exists := s.entries.Get2(uid); exists {
		return false
	}
	s.entries.Set(uid, struct{}{})
	return true
}
func (s *JavaSemanticUIDSet) Remove(uid int32) bool { return s.entries.Remove(uid) }
func (s *JavaSemanticUIDSet) Len() int              { return s.entries.Len() }
func (s *JavaSemanticUIDSet) All() func(func(int32) bool) {
	return func(yield func(int32) bool) {
		for uid := range s.entries.All() {
			if !yield(uid) {
				return
			}
		}
	}
}

// JavaSemanticSet uses the shared Java HashSet backing map for semantic keys
// whose hashCode is supplied by their source node classes.
type JavaSemanticSet[K comparable] struct {
	entries    *javaHashMap[K, struct{}]
	identities map[K]*byte
}

func NewJavaSemanticSet[K comparable](hashCode func(K) int32) *JavaSemanticSet[K] {
	set := &JavaSemanticSet[K]{entries: newJavaHashMap[K, struct{}](hashCode, nil), identities: map[K]*byte{}}
	// ArgLevelParam is not Comparable. HashMap's same-class tree tie-break
	// uses object identity, as the other native Java collection ports do.
	set.entries.tieBreak = func(a, b K) int {
		ah := uint32(reflect.ValueOf(set.identities[a]).Pointer()) & 0x7fffffff
		bh := uint32(reflect.ValueOf(set.identities[b]).Pointer()) & 0x7fffffff
		if ah <= bh {
			return -1
		}
		return 1
	}
	return set
}

func (s *JavaSemanticSet[K]) Add(key K) bool {
	if _, exists := s.entries.Get2(key); exists {
		return false
	}
	s.identities[key] = new(byte)
	s.entries.Set(key, struct{}{})
	return true
}

func (s *JavaSemanticSet[K]) All() func(func(K) bool) {
	return func(yield func(K) bool) {
		for key := range s.entries.All() {
			if !yield(key) {
				return
			}
		}
	}
}
