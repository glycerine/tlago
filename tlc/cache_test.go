package tlc

import "testing"

func TestSimpleCachePortedJavaHitMissBehavior(t *testing.T) {
	cache := NewSimpleCache(1)
	if cache.GetHitRate() != 1 {
		t.Fatalf("initial hit rate = %d, want Java sentinel 1", cache.GetHitRate())
	}
	if cache.GetHitRatio() != 1 {
		t.Fatalf("initial hit ratio = %v, want 1", cache.GetHitRatio())
	}

	if !cache.Hit(0) {
		t.Fatalf("zero fingerprint should hit the zero-initialized Java cache")
	}
	if cache.Hit(1) {
		t.Fatalf("first nonzero fingerprint should miss")
	}
	if !cache.Hit(1) {
		t.Fatalf("repeated fingerprint should hit")
	}
	if cache.Hit(3) {
		t.Fatalf("fingerprint colliding into the same slot should replace and miss")
	}
	if !cache.Hit(3) {
		t.Fatalf("repeated replacement fingerprint should hit")
	}
	if cache.GetHitRate() != 4 {
		t.Fatalf("hit rate = %d, want 4", cache.GetHitRate())
	}
}

func TestSimpleCacheHitRatioStringIsCompact(t *testing.T) {
	cache := NewSimpleCache()
	if got := cache.GetHitRatioAsString(); got != "1" {
		t.Fatalf("ratio string = %q, want 1", got)
	}
	cache.Hit(42)
	if got := cache.GetHitRatioAsString(); got != "0.5" {
		t.Fatalf("ratio string after miss = %q, want 0.5", got)
	}
	cache.cacheHit.Store(1234568)
	cache.cacheMiss.Store(1000)
	if got := cache.GetHitRatioAsString(); got != "1,234.568" {
		t.Fatalf("large ratio string = %q, want 1,234.568", got)
	}
}
