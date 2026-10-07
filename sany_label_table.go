// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"unicode/utf16"

	"github.com/glycerine/tlago/tlc"
)

// Labels use java.util.Hashtable's default buckets and elements enumeration.
// The table is shared by reference when attached to a definition or label.
type sanyLabelEntry struct {
	name  string
	hash  uint32
	value *sanySemLabelNode
	next  *sanyLabelEntry
}
type sanyLabelTable struct {
	buckets   []*sanyLabelEntry
	entries   map[string]*sanyLabelEntry
	threshold int
}

func newSanyLabelTable() *sanyLabelTable {
	return &sanyLabelTable{buckets: make([]*sanyLabelEntry, 11), entries: make(map[string]*sanyLabelEntry), threshold: 8}
}
func (t *sanyLabelTable) get(name string) *sanySemLabelNode {
	if t == nil {
		return nil
	}
	if e := t.entries[name]; e != nil {
		return e.value
	}
	return nil
}
func sanyLabelNameHash(name string) uint32 {
	var hash uint32
	for _, unit := range utf16.Encode([]rune(name)) {
		hash = 31*hash + uint32(unit)
	}
	return hash
}
func (t *sanyLabelTable) put(name string, value *sanySemLabelNode) *sanySemLabelNode {
	if value == nil {
		panic(tlc.NewNullPointerException(""))
	}
	if e := t.entries[name]; e != nil {
		previous := e.value
		e.value = value
		return previous
	}
	if len(t.entries) >= t.threshold {
		old := t.buckets
		t.buckets = make([]*sanyLabelEntry, len(old)*2+1)
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
	hash := sanyLabelNameHash(name)
	bucket := int(hash&0x7fffffff) % len(t.buckets)
	entry := &sanyLabelEntry{name: name, hash: hash, value: value, next: t.buckets[bucket]}
	t.buckets[bucket] = entry
	t.entries[name] = entry
	return nil
}
func (t *sanyLabelTable) add(label *sanySemLabelNode) bool {
	if label == nil {
		panic(tlc.NewNullPointerException(""))
	}
	if t.entries[label.name] != nil {
		return false
	}
	t.put(label.name, label)
	return true
}
func (t *sanyLabelTable) elements() []*sanySemLabelNode {
	result := make([]*sanySemLabelNode, 0)
	if t != nil {
		for i := len(t.buckets) - 1; i >= 0; i-- {
			for e := t.buckets[i]; e != nil; e = e.next {
				result = append(result, e.value)
			}
		}
	}
	return result
}
