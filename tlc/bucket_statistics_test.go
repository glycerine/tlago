package tlc

import (
	"math"
	"strings"
	"testing"
)

func TestBucketStatisticsJavaMetrics(t *testing.T) {
	stats := NewBucketStatistics("BucketStatisticsTest")
	if stats.Mean() != -1 || stats.Median() != -1 || stats.Min() != -1 || stats.Max() != -1 || stats.StdDev() != -1 {
		t.Fatalf("empty statistics should report -1 metrics")
	}

	stats.AddSample(0)
	if stats.Mean() != 0 || stats.Median() != 0 {
		t.Fatalf("single zero metrics mean=%v median=%v, want 0/0", stats.Mean(), stats.Median())
	}
	stats.AddSample(1)
	stats.AddSample(2)
	if stats.Mean() != 1 || stats.Median() != 1 {
		t.Fatalf("0,1,2 metrics mean=%v median=%v, want 1/1", stats.Mean(), stats.Median())
	}
	stats.AddSample(2)
	stats.AddSample(2)
	if stats.Mean() != 1.4 || stats.Median() != 2 {
		t.Fatalf("0,1,2,2,2 metrics mean=%v median=%v, want 1.4/2", stats.Mean(), stats.Median())
	}
	if stats.Min() != 0 || stats.Max() != 2 || stats.Observations() != 5 {
		t.Fatalf("min/max/obs = %d/%d/%d, want 0/2/5", stats.Min(), stats.Max(), stats.Observations())
	}
	if got := stats.String(); !strings.HasPrefix(got, "============================%n=BucketStatisticsTest=%n") {
		t.Fatalf("String() header = %q, want Java literal %%n separators", got[:min(len(got), 64)])
	}
}

func TestBucketStatisticsStdDevPercentileAndSamples(t *testing.T) {
	stats := NewBucketStatistics("BucketStatisticsTest")
	for _, sample := range []int{0, 0, 0, 1, 1, 2, 2, 2, 2, 3} {
		stats.AddSample(sample)
	}
	if rounded := math.Round(stats.StdDev()*10000) / 10000; rounded != 1.005 {
		t.Fatalf("stddev = %.4f, want 1.0050", rounded)
	}
	if stats.Percentile(0.5) != 2 || stats.Percentile(0.75) != 2 || stats.Percentile(0.999) != 3 {
		t.Fatalf("percentiles p50/p75/p999 = %v/%v/%v, want 2/2/3",
			stats.Percentile(0.5), stats.Percentile(0.75), stats.Percentile(0.999))
	}
	samples := stats.Samples()
	want := []BucketSample{{0, 3}, {1, 2}, {2, 4}, {3, 1}}
	if len(samples) != len(want) {
		t.Fatalf("samples len = %d, want %d", len(samples), len(want))
	}
	for i := range want {
		if samples[i] != want[i] {
			t.Fatalf("samples[%d] = %#v, want %#v", i, samples[i], want[i])
		}
	}
}

func TestBucketStatisticsRejectsNegativeAndNaN(t *testing.T) {
	stats := NewBucketStatistics("BucketStatisticsTest")
	requirePanic(t, func() { stats.AddSample(-1) })
	requirePanic(t, func() { stats.Percentile(math.NaN()) })
}

func TestFixedSizedBucketStatisticsUsesOverflowBucket(t *testing.T) {
	stats := NewFixedSizedBucketStatistics("FixedSizedBucketStatisticsTest", 8)
	stats.AddSample(16)
	stats.AddSample(16)
	stats.AddSample(16)
	if stats.Max() != 7 || stats.Median() != 7 || stats.Observations() != 3 {
		t.Fatalf("fixed overflow max/median/obs = %d/%d/%d, want 7/7/3",
			stats.Max(), stats.Median(), stats.Observations())
	}
}

func TestFixedSizedBucketStatisticsAllowsZeroSizeUntilSampleLikeJava(t *testing.T) {
	stats := NewFixedSizedBucketStatistics("FixedSizedBucketStatisticsTest", 0)
	if stats.Observations() != 0 || stats.Min() != -1 || stats.Max() != -1 {
		t.Fatalf("zero-sized stats obs/min/max = %d/%d/%d, want Java empty metrics",
			stats.Observations(), stats.Min(), stats.Max())
	}
	requirePanic(t, func() { stats.AddSample(0) })

	concurrent := NewFixedSizedConcurrentBucketStatistics("FixedSizedConcurrentBucketStatisticsTest", 0)
	if concurrent.Observations() != 0 || concurrent.Min() != -1 || concurrent.Max() != -1 {
		t.Fatalf("zero-sized concurrent stats obs/min/max = %d/%d/%d, want Java empty metrics",
			concurrent.Observations(), concurrent.Min(), concurrent.Max())
	}
	requirePanic(t, func() { concurrent.AddSample(0) })
}

func TestPortedStatisticUtilityVariants(t *testing.T) {
	dummy := NewDummyBucketStatistics()
	dummy.AddSample(100)
	if dummy.Observations() != 0 || dummy.Mean() != 0 || len(dummy.Samples()) != 0 {
		t.Fatalf("dummy stats obs/mean/samples = %d/%v/%d, want 0/0/0",
			dummy.Observations(), dummy.Mean(), len(dummy.Samples()))
	}

	concurrent := NewConcurrentBucketStatistics("ConcurrentBucketStatisticsTest")
	concurrent.AddSample(3)
	concurrent.AddSample(3)
	concurrent.AddSample(5)
	if concurrent.Observations() != 3 || concurrent.Median() != 3 || concurrent.Max() != 5 {
		t.Fatalf("concurrent stats obs/median/max = %d/%d/%d, want 3/3/5",
			concurrent.Observations(), concurrent.Median(), concurrent.Max())
	}

	fixed := NewFixedSizedConcurrentBucketStatistics("FixedSizedConcurrentBucketStatisticsTest", 4)
	fixed.AddSample(9)
	fixed.AddSample(0)
	if fixed.Max() != 3 || fixed.Min() != 0 || fixed.Observations() != 2 {
		t.Fatalf("fixed concurrent min/max/obs = %d/%d/%d, want 0/3/2",
			fixed.Min(), fixed.Max(), fixed.Observations())
	}
}

func TestCounterAndCountDistinctUtilities(t *testing.T) {
	noopCounter := NewCounterStatistic(false)
	noopCounter.Increment()
	noopCounter.Add(10)
	if noopCounter.GetCount() != 0 {
		t.Fatalf("noop counter = %d, want 0", noopCounter.GetCount())
	}

	counter := NewCounterStatistic(true)
	counter.Increment()
	counter.Add(10)
	if counter.GetCount() != 11 {
		t.Fatalf("counter = %d, want 11", counter.GetCount())
	}

	noopDistinct := NewCountDistinctNoop()
	noopDistinct.AddHash(1)
	if noopDistinct.Count() != -1 {
		t.Fatalf("noop distinct count = %d, want -1", noopDistinct.Count())
	}

	naive := NewCountDistinctNaive()
	naive.AddHash(1)
	naive.AddHash(1)
	naive.AddHash(2)
	if naive.Count() != 2 {
		t.Fatalf("naive distinct count = %d, want 2", naive.Count())
	}

	hll := NewCountDistinctHyperLogLog(4)
	hll.AddHash(0x1000000000000000)
	hll.AddHash(0x2000000000000000)
	if hll.Count() <= 0 {
		t.Fatalf("hyperloglog count = %d, want positive estimate", hll.Count())
	}
}

func requirePanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
