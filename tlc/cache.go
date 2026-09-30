package tlc

import (
	"strconv"
	"strings"
	"sync/atomic"
)

type SimpleCache struct {
	cacheHit  atomic.Int64
	cacheMiss atomic.Int64
	mask      uint64
	cache     []uint64
}

func NewSimpleCache(size ...int) *SimpleCache {
	bits := 10
	if len(size) > 0 {
		bits = size[0]
	}
	if bits < 0 {
		bits = 0
	}
	capacity := 1 << bits
	out := &SimpleCache{
		mask:  uint64(capacity - 1),
		cache: make([]uint64, capacity),
	}
	out.cacheHit.Store(1)
	out.cacheMiss.Store(1)
	return out
}

func (c *SimpleCache) Hit(fp uint64) bool {
	if c == nil || len(c.cache) == 0 {
		return false
	}
	index := int(fp & c.mask)
	if c.cache[index] == fp {
		c.cacheHit.Add(1)
		return true
	}
	c.cacheMiss.Add(1)
	c.cache[index] = fp
	return false
}

func (c *SimpleCache) GetHitRatio() float64 {
	if c == nil {
		return 0
	}
	miss := c.cacheMiss.Load()
	if miss == 0 {
		return 0
	}
	return float64(c.cacheHit.Load()) / float64(miss)
}

func (c *SimpleCache) GetHitRatioAsString() string {
	text := strconv.FormatFloat(c.GetHitRatio(), 'f', 3, 64)
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "" {
		return "0"
	}
	return text
}

func (c *SimpleCache) GetHitRate() int64 {
	if c == nil {
		return 0
	}
	return c.cacheHit.Load()
}
