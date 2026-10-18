package tlc

import (
	"strings"
	"sync"
)

type tlaRegistryEntry struct {
	hash  int32
	key   string
	value string
	next  *tlaRegistryEntry
}

type tlaRegistryTable struct {
	mu        sync.Mutex
	table     []*tlaRegistryEntry
	count     int
	threshold int
}

func newTLARegistryTable() *tlaRegistryTable {
	return &tlaRegistryTable{table: make([]*tlaRegistryEntry, 11), threshold: 8}
}

func tlaRegistryIndex(hash int32, capacity int) int {
	return int(uint32(hash)&0x7fffffff) % capacity
}

func (t *tlaRegistryTable) rehash() {
	old := t.table
	next := make([]*tlaRegistryEntry, len(old)*2+1)
	t.threshold = int(float64(len(next)) * 0.75)
	t.table = next
	for i := len(old) - 1; i >= 0; i-- {
		for entry := old[i]; entry != nil; {
			moved := entry
			entry = entry.next
			index := tlaRegistryIndex(moved.hash, len(next))
			moved.next = next[index]
			next[index] = moved
		}
	}
}

func (t *tlaRegistryTable) get(key string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	hash := javaStringHashCode(key)
	for entry := t.table[tlaRegistryIndex(hash, len(t.table))]; entry != nil; entry = entry.next {
		if entry.hash == hash && entry.key == key {
			return entry.value, true
		}
	}
	return "", false
}

func (t *tlaRegistryTable) put(key string, value string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	hash := javaStringHashCode(key)
	index := tlaRegistryIndex(hash, len(t.table))
	for entry := t.table[index]; entry != nil; entry = entry.next {
		if entry.hash == hash && entry.key == key {
			previous := entry.value
			entry.value = value
			return previous, true
		}
	}
	if t.count >= t.threshold {
		t.rehash()
		index = tlaRegistryIndex(hash, len(t.table))
	}
	t.table[index] = &tlaRegistryEntry{hash: hash, key: key, value: value, next: t.table[index]}
	t.count++
	return "", false
}

func (t *tlaRegistryTable) allNames() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var out strings.Builder
	out.WriteByte('{')
	first := true
	for i := len(t.table) - 1; i >= 0; i-- {
		for entry := t.table[i]; entry != nil; entry = entry.next {
			if !first {
				out.WriteString(", ")
			}
			first = false
			out.WriteString(entry.key)
		}
	}
	out.WriteByte('}')
	return out.String()
}

var tlaRegistry = newTLARegistryTable()

func init() {
	RegisterStandardTLAOperators()
}

func RegisterStandardTLAOperators() {
	TLARegistryPut("Plus", "+")
	TLARegistryPut("Minus", "-")
	TLARegistryPut("Times", "*")
	TLARegistryPut("LT", "<")
	TLARegistryPut("LE", "\\leq")
	TLARegistryPut("GT", ">")
	TLARegistryPut("GEQ", "\\geq")
	TLARegistryPut("DotDot", "..")
	TLARegistryPut("Neg", "-.")
	TLARegistryPut("Divide", "\\div")
	TLARegistryPut("Mod", "%")
	TLARegistryPut("Expt", "^")
	TLARegistryPut("Concat", "\\o")
	TLARegistryPut("MakeFcn", ":>")
	TLARegistryPut("CombineFcn", "@@")
	TLARegistryPut("BagCup", "\\oplus")
	TLARegistryPut("BagDiff", "\\ominus")
	TLARegistryPut("SqSubseteq", "\\sqsubseteq")
}

func TLARegistryGet(name string) (string, bool) {
	return tlaRegistry.get(name)
}

func TLARegistryPut(tlaName string, goName string) (previous string, replaced bool) {
	return tlaRegistry.put(tlaName, goName)
}

func TLARegistryMapName(name string) string {
	if mapped, ok := TLARegistryGet(name); ok {
		return mapped
	}
	return name
}

func TLARegistryAllNames() string {
	return tlaRegistry.allNames()
}
