// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// Context.content enumerates Hashtable values, separately from the linked Pair
// history used for definitions. Replacing a binding does not move its bucket.
type sanyContextBucketEntry struct {
	entry *sanyContextEntry
	hash  uint32
	next  *sanyContextBucketEntry
}

func (context *sanyContext) putContentEntry(entry *sanyContextEntry) {
	if context.contentBuckets == nil {
		context.contentBuckets = make([]*sanyContextBucketEntry, 11)
		context.contentThreshold = 8
	}
	hash := sanyLabelNameHash(entry.key.name)
	bucket := int(hash&0x7fffffff) % len(context.contentBuckets)
	for current := context.contentBuckets[bucket]; current != nil; current = current.next {
		if current.entry.key == entry.key {
			current.entry = entry
			return
		}
	}
	if len(context.table) >= context.contentThreshold {
		old := context.contentBuckets
		context.contentBuckets = make([]*sanyContextBucketEntry, len(old)*2+1)
		context.contentThreshold = 3 * len(context.contentBuckets) / 4
		for i := len(old) - 1; i >= 0; i-- {
			for current := old[i]; current != nil; {
				next := current.next
				bucket := int(current.hash&0x7fffffff) % len(context.contentBuckets)
				current.next = context.contentBuckets[bucket]
				context.contentBuckets[bucket] = current
				current = next
			}
		}
		bucket = int(hash&0x7fffffff) % len(context.contentBuckets)
	}
	context.contentBuckets[bucket] = &sanyContextBucketEntry{entry: entry, hash: hash, next: context.contentBuckets[bucket]}
}
func (context *sanyContext) contentSymbols() []sanySemSymbol {
	if context == nil {
		panic(tlc.NewNullPointerException(""))
	}
	symbols := make([]sanySemSymbol, 0, len(context.table))
	for i := len(context.contentBuckets) - 1; i >= 0; i-- {
		for entry := context.contentBuckets[i]; entry != nil; entry = entry.next {
			symbols = append(symbols, entry.entry.sym)
		}
	}
	return symbols
}
