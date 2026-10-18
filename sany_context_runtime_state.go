// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// runtimeState preserves actual Pair membership and Hashtable topology. The
// caller maps symbols through shared runtime adapters; neither class filtering
// nor declaration replay is appropriate for this complete context transfer.
func (context *sanyContext) runtimeState(convert func(sanySemSymbol) tlc.SemanticNode) tlc.SemanticContextState {
	if context == nil {
		panic(tlc.NewNullPointerException())
	}
	state := tlc.SemanticContextState{
		Pairs:     make([]tlc.SemanticContextBinding, len(context.order)),
		Buckets:   make([][]int, len(context.contentBuckets)),
		Threshold: context.contentThreshold,
	}
	// SANY's current Go context allocates its initial Hashtable lazily. Java's
	// empty Context already owns the default eleven buckets and threshold eight.
	if context.contentBuckets == nil {
		state.Buckets = make([][]int, 11)
		state.Threshold = 8
	}
	indexes := make(map[*sanyContextEntry]int, len(context.order))
	for i, pair := range context.order {
		indexes[pair] = i
		state.Pairs[i] = tlc.SemanticContextBinding{
			Key:  tlc.SemanticContextKey{Name: tlc.UniqueStringOf(pair.key.name), Module: pair.key.module},
			Node: convert(pair.sym),
		}
	}
	for i, bucket := range context.contentBuckets {
		for entry := bucket; entry != nil; entry = entry.next {
			index, present := indexes[entry.entry]
			if !present {
				panic("SANY Hashtable entry has no retained Context Pair")
			}
			state.Buckets[i] = append(state.Buckets[i], index)
		}
	}
	return state
}
