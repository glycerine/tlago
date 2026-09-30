package tlc

import (
	"math"
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

func requirePanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
