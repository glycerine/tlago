// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlc

// SemanticContextState transfers a checked parser context into a runtime view.
// Pairs retain oldest-to-newest Pair history, including replaced bindings.
// Each bucket contains indexes into Pairs in Hashtable chain order. These are
// separate structures in Java: replaying history would rehash the context and
// can change enumeration order. Node adapters must preserve source identity.
type SemanticContextState struct {
	Pairs     []SemanticContextBinding
	Buckets   [][]int
	Threshold int
}

type SemanticContextBinding struct {
	Key  SemanticContextKey
	Node SemanticNode
}

// ImportState populates an already published context shell without replaying
// AddSymbolToContext. The descriptor arrays are copied into runtime structures;
// semantic nodes and the shared Pair references retain their supplied identity.
// It preserves ModuleTable, which belongs to the runtime module graph.
func (c *SemanticContext) ImportState(state SemanticContextState) {
	if c == nil {
		panic(NewNullPointerException())
	}
	c.lastPair = nil
	pairs := make([]*semanticContextPair, len(state.Pairs))
	for i, binding := range state.Pairs {
		pair := &semanticContextPair{link: c.lastPair, key: binding.Key, node: binding.Node}
		pairs[i] = pair
		c.lastPair = pair
	}
	c.buckets = make([]*semanticContextEntry, len(state.Buckets))
	c.count = 0
	c.threshold = state.Threshold
	for i, chain := range state.Buckets {
		for j := len(chain) - 1; j >= 0; j-- {
			pair := pairs[chain[j]]
			c.buckets[i] = &semanticContextEntry{key: pair.key, pair: pair, next: c.buckets[i]}
			c.count++
		}
	}
}
