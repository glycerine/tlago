package tlc

import (
	"sync"
	"sync/atomic"
)

// Java initializes a class's static UniqueStrings on its first active use.
// Keep that ordering in each fresh TLC context: ModelConfig can create model
// values before loading the built-in operator and native-module classes.
type classInterning struct {
	mu    sync.Mutex
	table atomic.Pointer[InternTable]
}

func (c *classInterning) ensure(initialize func()) {
	if c.table.Load() == internTable {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.table.Load() != internTable {
		initialize()
		c.table.Store(internTable)
	}
}
