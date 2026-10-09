package tlc

import (
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
	capacity := javaSimpleCacheCapacity(bits)
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmtInt(capacity)))
	}
	out := &SimpleCache{
		mask:  uint64(capacity - 1),
		cache: make([]uint64, capacity),
	}
	out.cacheHit.Store(1)
	out.cacheMiss.Store(1)
	return out
}

func javaSimpleCacheCapacity(size int) int {
	shift := uint(size) & 31
	return int(int32(1) << shift)
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
		panic(NewNullPointerException())
	}
	return float64(c.cacheHit.Load()) / float64(c.cacheMiss.Load())
}

func (c *SimpleCache) GetHitRatioAsString() string {
	mpNumberSymbolsOnce.Do(initializeMessageNumberSymbols)
	return compactCacheRatio(c.GetHitRatio(), mpNumberSymbols, cacheNumberSpecialSymbols)
}

func (c *SimpleCache) GetHitRate() int64 {
	if c == nil {
		return 0
	}
	return c.cacheHit.Load()
}

func groupDecimalIntegerPart(text string) string {
	parts := strings.SplitN(text, ".", 2)
	intPart := parts[0]
	sign := ""
	if strings.HasPrefix(intPart, "-") {
		sign = "-"
		intPart = intPart[1:]
	}
	if len(intPart) <= 3 {
		return text
	}
	var b strings.Builder
	b.WriteString(sign)
	head := len(intPart) % 3
	if head == 0 {
		head = 3
	}
	b.WriteString(intPart[:head])
	for i := head; i < len(intPart); i += 3 {
		b.WriteByte(',')
		b.WriteString(intPart[i : i+3])
	}
	if len(parts) == 2 {
		b.WriteByte('.')
		b.WriteString(parts[1])
	}
	return b.String()
}
