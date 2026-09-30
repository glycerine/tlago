package tlc

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type BucketSample struct {
	Amount int
	Count  int64
}

type BucketStatistics struct {
	Title        string
	observations int64
	counts       map[int]int64
	keys         []int
}

func NewBucketStatistics(title string) *BucketStatistics {
	return &BucketStatistics{
		Title:  title,
		counts: make(map[int]int64),
	}
}

func (s *BucketStatistics) AddSample(amount int) {
	if amount < 0 {
		panic("Negative amount invalid")
	}
	if _, ok := s.counts[amount]; !ok {
		idx := sort.SearchInts(s.keys, amount)
		s.keys = append(s.keys, 0)
		copy(s.keys[idx+1:], s.keys[idx:])
		s.keys[idx] = amount
	}
	s.counts[amount]++
	s.observations++
}

func (s *BucketStatistics) Observations() int64 {
	if s == nil {
		return 0
	}
	return s.observations
}

func (s *BucketStatistics) Samples() []BucketSample {
	if s == nil {
		return nil
	}
	samples := make([]BucketSample, 0, len(s.keys))
	for _, key := range s.keys {
		if count := s.counts[key]; count > 0 {
			samples = append(samples, BucketSample{Amount: key, Count: count})
		}
	}
	return samples
}

func (s *BucketStatistics) Add(other *BucketStatistics) {
	if s == nil || other == nil {
		return
	}
	s.observations += other.Observations()
	for _, sample := range other.Samples() {
		if _, ok := s.counts[sample.Amount]; !ok {
			idx := sort.SearchInts(s.keys, sample.Amount)
			s.keys = append(s.keys, 0)
			copy(s.keys[idx+1:], s.keys[idx:])
			s.keys[idx] = sample.Amount
		}
		s.counts[sample.Amount] += sample.Count
	}
}

func (s *BucketStatistics) Median() int {
	return bucketMedian(s.Observations(), s.Samples())
}

func (s *BucketStatistics) Mean() float64 {
	return bucketMean(s.Observations(), s.Samples())
}

func (s *BucketStatistics) Min() int {
	return bucketMin(s.Observations(), s.Samples())
}

func (s *BucketStatistics) Max() int {
	return bucketMax(s.Observations(), s.Samples())
}

func (s *BucketStatistics) StdDev() float64 {
	return bucketStdDev(s.Observations(), s.Samples())
}

func (s *BucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s.Observations(), s.Samples(), quantile)
}

func (s *BucketStatistics) String() string {
	if s == nil {
		return ""
	}
	return bucketString(s.Title, s.Observations(), s.Samples(), s.Min(), s.Max(), s.Mean(), s.Median(), s.StdDev(), s.Percentile)
}

type FixedSizedBucketStatistics struct {
	Title        string
	observations int64
	buckets      []int64
}

func NewFixedSizedBucketStatistics(title string, maximum int) *FixedSizedBucketStatistics {
	if maximum <= 0 {
		panic("FixedSizedBucketStatistics requires a positive maximum")
	}
	return &FixedSizedBucketStatistics{
		Title:   title,
		buckets: make([]int64, maximum),
	}
}

func (s *FixedSizedBucketStatistics) AddSample(amount int) {
	if amount < 0 {
		panic("Negative amount invalid")
	}
	idx := amount
	if idx >= len(s.buckets) {
		idx = len(s.buckets) - 1
	}
	s.buckets[idx]++
	s.observations++
}

func (s *FixedSizedBucketStatistics) Observations() int64 {
	if s == nil {
		return 0
	}
	return s.observations
}

func (s *FixedSizedBucketStatistics) Samples() []BucketSample {
	if s == nil {
		return nil
	}
	samples := make([]BucketSample, 0)
	for amount, count := range s.buckets {
		if count > 0 {
			samples = append(samples, BucketSample{Amount: amount, Count: count})
		}
	}
	return samples
}

func (s *FixedSizedBucketStatistics) Median() int {
	return bucketMedian(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) Mean() float64 {
	return bucketMean(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) Min() int {
	return bucketMin(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) Max() int {
	return bucketMax(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) StdDev() float64 {
	return bucketStdDev(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s.Observations(), s.Samples(), quantile)
}

func (s *FixedSizedBucketStatistics) String() string {
	if s == nil {
		return ""
	}
	return bucketString(s.Title, s.Observations(), s.Samples(), s.Min(), s.Max(), s.Mean(), s.Median(), s.StdDev(), s.Percentile)
}

func bucketMedian(observations int64, samples []BucketSample) int {
	if observations <= 0 {
		return -1
	}
	sum := int64(0)
	for _, sample := range samples {
		sum += sample.Count
		if sum > observations/2 {
			return sample.Amount
		}
	}
	panic("bug, shoud not get here")
}

func bucketMean(observations int64, samples []BucketSample) float64 {
	sum := int64(0)
	for _, sample := range samples {
		sum += sample.Count * int64(sample.Amount)
	}
	if observations > 0 {
		return float64(sum) / float64(observations)
	}
	return -1
}

func bucketMin(observations int64, samples []BucketSample) int {
	if observations <= 0 {
		return -1
	}
	return samples[0].Amount
}

func bucketMax(observations int64, samples []BucketSample) int {
	if observations <= 0 {
		return -1
	}
	return samples[len(samples)-1].Amount
}

func bucketStdDev(observations int64, samples []BucketSample) float64 {
	if observations <= 0 {
		return -1
	}
	mean := bucketMean(observations, samples)
	sum := 0.0
	for _, sample := range samples {
		diff := float64(sample.Amount) - mean
		sum += diff * diff * float64(sample.Count)
	}
	return math.Sqrt(sum / float64(observations))
}

func bucketPercentile(observations int64, samples []BucketSample, quantile float64) float64 {
	if math.IsNaN(quantile) {
		panic("NaN")
	}
	if observations <= 0 {
		return -1
	}
	quantile = math.Min(1, quantile)
	quantile = math.Max(0, quantile)
	pos := int(float64(observations) * quantile)
	if int64(pos) > observations {
		return float64(len(samples))
	}
	if pos < 0 {
		return 0
	}
	count := int64(0)
	for _, sample := range samples {
		count += sample.Count
		if count > int64(pos) {
			return float64(sample.Amount)
		}
	}
	return quantile
}

func bucketString(title string, observations int64, samples []BucketSample, min int, max int, mean float64, median int, stddev float64, percentile func(float64) float64) string {
	var b strings.Builder
	b.WriteString("============================\n")
	b.WriteString("=")
	b.WriteString(title)
	b.WriteString("=\n")
	b.WriteString("============================\n")
	b.WriteString(fmt.Sprintf("Observations: %d\n", observations))
	b.WriteString(fmt.Sprintf("Min: %d\n", min))
	b.WriteString(fmt.Sprintf("Max: %d\n", max))
	b.WriteString(fmt.Sprintf("Mean: %.2f\n", mean))
	b.WriteString(fmt.Sprintf("Median: %d\n", median))
	b.WriteString(fmt.Sprintf("Standard deviation: %.2f\n", stddev))
	b.WriteString(fmt.Sprintf("75%%: %.2f\n", percentile(0.75)))
	b.WriteString(fmt.Sprintf("95%%: %.2f\n", percentile(0.95)))
	b.WriteString(fmt.Sprintf("98%%: %.2f\n", percentile(0.98)))
	b.WriteString(fmt.Sprintf("99%%: %.2f\n", percentile(0.99)))
	b.WriteString(fmt.Sprintf("99.9%%: %.2f\n", percentile(0.999)))
	b.WriteString("numEdges/occurrences (log scale)\n")
	b.WriteString("--------------------------------\n")
	for _, sample := range samples {
		b.WriteString(fmt.Sprintf("%02d:%02d ", sample.Amount, sample.Count))
		for j := 0; j < int(math.Log(float64(sample.Count))); j++ {
			b.WriteByte('#')
		}
		b.WriteByte('\n')
	}
	b.WriteString("============================")
	return b.String()
}
