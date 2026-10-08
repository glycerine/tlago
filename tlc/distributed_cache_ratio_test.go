package tlc

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// No original SimpleCache test exists. Reference counters and ratio bits come
// from the source hit / (double) miss expression on OpenJDK 21.0.12.1.
func TestDistributedCacheRatioCounterReference(t *testing.T) {
	file, err := os.Open("test_vectors/distributed/simple_cache_ratios.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	rows := 0
	cache := NewSimpleCache()
	worker := &DistributedWorker{Cache: cache}
	for scanner.Scan() {
		row := strings.Split(scanner.Text(), "\t")
		if len(row) != 4 {
			t.Fatal("invalid cache counter reference row")
		}
		hits, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		misses, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		bits, err := strconv.ParseUint(row[2], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		cache.cacheHit.Store(hits)
		cache.cacheMiss.Store(misses)
		for _, ratio := range []float64{cache.GetHitRatio(), worker.GetCacheRateRatio()} {
			// Arithmetic NaN sign/payload is hardware-dependent. Finite values,
			// infinities and signed zero retain the exact reference bits.
			if math.IsNaN(math.Float64frombits(bits)) {
				if !math.IsNaN(ratio) {
					t.Errorf("cache counters %d/%d = %v, want NaN", hits, misses, ratio)
				}
			} else if math.Float64bits(ratio) != bits {
				t.Errorf("cache counters %d/%d ratio bits = %x, want %x", hits, misses, math.Float64bits(ratio), bits)
			}
		}
		if cache.GetHitRate() != hits {
			t.Fatal("ratio observation mutated counters")
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 271 {
		t.Fatalf("cache counter rows = %d, want 271", rows)
	}
}

func TestDistributedCacheRatioMissingOwner(t *testing.T) {
	for _, call := range []func(){
		func() { (*SimpleCache)(nil).GetHitRatio() },
		func() { (*DistributedWorker)(nil).GetCacheRateRatio() },
		func() { (&DistributedWorker{}).GetCacheRateRatio() },
	} {
		func() {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("missing cache owner = %T/%v", failure, failure)
				}
			}()
			call()
		}()
	}
}

func TestWorkerRPCCacheCounterExtremes(t *testing.T) {
	cache := NewSimpleCache()
	_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(&DistributedWorker{Cache: cache}))
	for _, test := range []struct {
		hits, misses int64
		want         float64
	}{
		{1, 0, math.Inf(1)}, {-1, 0, math.Inf(-1)}, {0, 0, math.NaN()},
		{0, -1, math.Copysign(0, -1)}, {math.MinInt64, -1, 9223372036854775808.0},
		{math.MaxInt64, 1, 9223372036854775808.0},
	} {
		cache.cacheHit.Store(test.hits)
		cache.cacheMiss.Store(test.misses)
		ratio, err := client.GetCacheRateRatio()
		if err != nil {
			t.Fatal(err)
		}
		if math.IsNaN(test.want) {
			if !math.IsNaN(ratio) {
				t.Fatal("native RPC replaced NaN")
			}
		} else if math.Float64bits(ratio) != math.Float64bits(test.want) {
			t.Fatalf("native RPC counter ratio = %v, want %v", ratio, test.want)
		}
		if math.Float64bits(ratio) != math.Float64bits(cache.GetHitRatio()) {
			t.Fatal("native RPC changed sender floating-point bits")
		}
	}
}
