package tlc

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
