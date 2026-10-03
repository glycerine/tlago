package tlago

import (
	"slices"
	"sort"
	"unicode/utf16"

	"github.com/glycerine/tlago/tlc"
)

// SANY Context keeps insertion links for EXTENDS and operator definitions, but
// ModuleNode obtains declarations by reversing Hashtable.elements(). Keep both
// orders: sorting declaration names changes state printing and variable slots.
type tlcBridgeContextEntry struct {
	name      string
	kind      DeclarationKind
	module    *Module
	position  Position
	local     bool
	moduleKey bool
	initial   bool
}

func tlcBridgeContextEntries(spec *Spec, mod *Module, visiting map[*Module]bool) []tlcBridgeContextEntry {
	if mod == nil || visiting[mod] {
		return nil
	}
	visiting[mod] = true
	defer delete(visiting, mod)
	var entries []tlcBridgeContextEntry
	type key struct {
		name   string
		module bool
	}
	indices := map[key]int{}
	put := func(entry tlcBridgeContextEntry, replace bool) {
		k := key{entry.name, entry.moduleKey}
		if i, found := indices[k]; found {
			if replace {
				entries[i] = entry
			}
			return
		}
		indices[k] = len(entries)
		entries = append(entries, entry)
	}
	// External SymbolTable starts with Context.getGlobalContext().duplicate().
	// Builtins occupy buckets even though declaration filtering omits them.
	if moduleNameForSourcePosition(mod.Pos) == mod.Name {
		for _, builtin := range sanyBuiltinOperators {
			put(tlcBridgeContextEntry{name: builtin.name, initial: true}, false)
		}
	}
	for _, ext := range mod.Extends {
		if slices.Contains(mod.ImplicitExtends, ext) {
			continue
		}
		for _, entry := range tlcBridgeContextEntries(spec, spec.Modules[ext], visiting) {
			if !entry.local {
				entry.initial = false
				put(entry, false)
			}
		}
	}
	type item struct {
		position Position
		entries  []tlcBridgeContextEntry
		instance *Instance
	}
	var items []item
	for _, declaration := range mod.Declarations {
		var declared []tlcBridgeContextEntry
		for _, name := range declaration.Names {
			position := declaration.NamePositions[name]
			if position.Line == 0 {
				position = declaration.Pos
			}
			declared = append(declared, tlcBridgeContextEntry{name: name, kind: declaration.Kind, module: mod, position: position})
		}
		items = append(items, item{position: declaration.Pos, entries: declared})
	}
	for _, declaration := range mod.Recursives {
		var declared []tlcBridgeContextEntry
		for _, name := range declaration.Names {
			declared = append(declared, tlcBridgeContextEntry{name: name, kind: OperatorDecl, module: mod})
		}
		items = append(items, item{position: declaration.Pos, entries: declared})
	}
	for _, definition := range mod.Definitions {
		kind := OperatorDecl
		if definition.TheoremLike {
			kind = ""
		}
		items = append(items, item{position: definition.SourcePosition(), entries: []tlcBridgeContextEntry{{name: definition.Name, kind: kind, module: mod, local: definition.Local}}})
	}
	for _, assumption := range mod.Assumptions {
		if assumption.Name != "" {
			items = append(items, item{position: assumption.SourcePosition(), entries: []tlcBridgeContextEntry{{name: assumption.Name, module: mod}}})
		}
	}
	for _, inner := range mod.Nested {
		items = append(items, item{position: inner.Pos, entries: []tlcBridgeContextEntry{{name: inner.Name, module: inner, moduleKey: true}}})
	}
	for i := range mod.Instances {
		instance := &mod.Instances[i]
		items = append(items, item{position: instance.SourcePosition(), instance: instance})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].position.Line != items[j].position.Line {
			return items[i].position.Line < items[j].position.Line
		}
		return items[i].position.Column < items[j].position.Column
	})
	for _, item := range items {
		if instance := item.instance; instance != nil {
			for _, entry := range tlcBridgeContextEntries(spec, spec.Modules[instance.Module], visiting) {
				if entry.local || entry.kind != OperatorDecl {
					continue
				}
				if instance.Name != "" {
					entry.name = instance.Name + "!" + entry.name
				}
				entry.local = instance.Local
				put(entry, true)
			}
			if instance.Name != "" {
				put(tlcBridgeContextEntry{name: instance.Name, kind: OperatorDecl, module: mod, local: instance.Local}, true)
			}
		}
		for _, entry := range item.entries {
			put(entry, true)
		}
	}
	return entries
}

func tlcBridgeDeclarationOrder(entries []tlcBridgeContextEntry, kind DeclarationKind) []tlcBridgeContextEntry {
	buckets := make([][]tlcBridgeContextEntry, 11)
	count, threshold := 0, 8
	put := func(table [][]tlcBridgeContextEntry, entry tlcBridgeContextEntry) {
		var hash uint32
		for _, unit := range utf16.Encode([]rune(entry.name)) {
			hash = 31*hash + uint32(unit)
		}
		i := int(hash&0x7fffffff) % len(table)
		table[i] = append([]tlcBridgeContextEntry{entry}, table[i]...)
	}
	insert := func(entry tlcBridgeContextEntry) {
		if count >= threshold {
			old := buckets
			buckets = make([][]tlcBridgeContextEntry, 2*len(old)+1)
			threshold = 3 * len(buckets) / 4
			for i := len(old) - 1; i >= 0; i-- {
				for _, previous := range old[i] {
					put(buckets, previous)
				}
			}
		}
		put(buckets, entry)
		count++
	}
	// Context.duplicate inserts the global lastPair list from newest to oldest
	// while retaining the original pair links for subsequent EXTENDS merges.
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].initial {
			insert(entries[i])
		}
	}
	for _, entry := range entries {
		if !entry.initial {
			insert(entry)
		}
	}
	var declarations []tlcBridgeContextEntry
	for i := len(buckets) - 1; i >= 0; i-- {
		for _, entry := range buckets[i] {
			if entry.kind == kind {
				declarations = append(declarations, entry)
			}
		}
	}
	for i, j := 0, len(declarations)-1; i < j; i, j = i+1, j-1 {
		declarations[i], declarations[j] = declarations[j], declarations[i]
	}
	return declarations
}

func (b *tlcBridge) variableDeclarations() []*tlc.SymbolNode {
	entries := tlcBridgeDeclarationOrder(tlcBridgeContextEntries(b.spec, b.spec.Root, map[*Module]bool{}), VariableDecl)
	nodes := make([]*tlc.SymbolNode, len(entries))
	for i, entry := range entries {
		nodes[i] = b.symbol(entry.name)
		nodes[i].MarkVariableDecl()
		position := entry.position
		position.File = entry.module.Name
		nodes[i].Location = b.sourceLocationForPosition(position)
	}
	return nodes
}
