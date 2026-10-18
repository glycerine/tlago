package tlc

import "sync"

type semanticLabelEntry struct {
	name  *UniqueString
	hash  uint32
	value *LabelNode
	next  *semanticLabelEntry
}

// SemanticLabelTable preserves java.util.Hashtable's default capacity,
// resizing and values-enumeration order for OpDefOrLabelNode label tables.
type SemanticLabelTable struct {
	mu        sync.RWMutex
	buckets   []*semanticLabelEntry
	entries   map[string]*semanticLabelEntry
	threshold int
	runtime   SemanticLabelTableRuntime
}

type SemanticLabelTableRuntime interface {
	RuntimeLabelGet(*UniqueString) (*LabelNode, bool)
	RuntimeLabelAdd(*LabelNode) (bool, bool)
	RuntimeLabelElements() ([]*LabelNode, bool)
}

func NewSemanticLabelTable() *SemanticLabelTable {
	return &SemanticLabelTable{
		buckets:   make([]*semanticLabelEntry, 11),
		entries:   make(map[string]*semanticLabelEntry),
		threshold: 8,
	}
}

// NewSemanticLabelTableFromBuckets installs a parser-owned Hashtable snapshot.
// Each inner slice is in head-to-tail chain order and remains independently
// mutable after the conversion.
func NewSemanticLabelTableFromBuckets(buckets [][]*LabelNode, threshold int) *SemanticLabelTable {
	table := &SemanticLabelTable{
		buckets:   make([]*semanticLabelEntry, len(buckets)),
		entries:   make(map[string]*semanticLabelEntry),
		threshold: threshold,
	}
	for i, labels := range buckets {
		var tail **semanticLabelEntry = &table.buckets[i]
		for _, label := range labels {
			name := label.GetName()
			entry := &semanticLabelEntry{name: name, hash: uint32(name.HashCode()), value: label}
			*tail = entry
			tail = &entry.next
			table.entries[name.String()] = entry
		}
	}
	return table
}

func (t *SemanticLabelTable) Get(name *UniqueString) *LabelNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	name.HashCode()
	if t.runtime != nil {
		if label, present := t.runtime.RuntimeLabelGet(name); present {
			return label
		}
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if entry := t.entries[name.String()]; entry != nil {
		return entry.value
	}
	return nil
}

func (t *SemanticLabelTable) Add(label *LabelNode) bool {
	if t == nil {
		panic(NewNullPointerException())
	}
	name := label.GetName()
	hash := uint32(name.HashCode())
	if t.runtime != nil {
		if added, present := t.runtime.RuntimeLabelAdd(label); present {
			return added
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.entries[name.String()] != nil {
		return false
	}
	if len(t.entries) >= t.threshold {
		t.rehash()
	}
	bucket := int(hash&0x7fffffff) % len(t.buckets)
	entry := &semanticLabelEntry{name: name, hash: hash, value: label, next: t.buckets[bucket]}
	t.buckets[bucket] = entry
	t.entries[name.String()] = entry
	return true
}

func (t *SemanticLabelTable) rehash() {
	old := t.buckets
	t.buckets = make([]*semanticLabelEntry, len(old)*2+1)
	t.threshold = 3 * len(t.buckets) / 4
	for i := len(old) - 1; i >= 0; i-- {
		for entry := old[i]; entry != nil; {
			next := entry.next
			bucket := int(entry.hash&0x7fffffff) % len(t.buckets)
			entry.next = t.buckets[bucket]
			t.buckets[bucket] = entry
			entry = next
		}
	}
}

func (t *SemanticLabelTable) Elements() []*LabelNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.runtime != nil {
		if labels, present := t.runtime.RuntimeLabelElements(); present {
			return labels
		}
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*LabelNode, 0, len(t.entries))
	for i := len(t.buckets) - 1; i >= 0; i-- {
		for entry := t.buckets[i]; entry != nil; entry = entry.next {
			result = append(result, entry.value)
		}
	}
	return result
}

func (t *SemanticLabelTable) BindRuntime(runtime SemanticLabelTableRuntime) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.runtime = runtime
}

func (t *SemanticLabelTable) CanonicalRuntime() any {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.runtime
}

func (t *SemanticLabelTable) SnapshotBuckets() ([][]*LabelNode, int) {
	if t == nil {
		panic(NewNullPointerException())
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	buckets := make([][]*LabelNode, len(t.buckets))
	for i, entry := range t.buckets {
		for ; entry != nil; entry = entry.next {
			buckets[i] = append(buckets[i], entry.value)
		}
	}
	return buckets, t.threshold
}

func semanticNodeLabels(base *SemanticNodeBase, fallback *SemanticLabelTable) *SemanticLabelTable {
	if base != nil {
		if provider, ok := base.CanonicalLevelData.(interface {
			RuntimeLabels() (*SemanticLabelTable, bool)
		}); ok {
			if labels, present := provider.RuntimeLabels(); present {
				return labels
			}
		}
	}
	return fallback
}

func setSemanticNodeLabels(base *SemanticNodeBase, labels *SemanticLabelTable) {
	if base != nil {
		if provider, ok := base.CanonicalLevelData.(interface {
			RuntimeSetLabels(*SemanticLabelTable) bool
		}); ok {
			provider.RuntimeSetLabels(labels)
		}
	}
}
