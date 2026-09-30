package tlc

import (
	"fmt"
	"math"
	"math/bits"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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
	s.AddSampleCount(amount, 1)
}

func (s *BucketStatistics) AddSampleCount(amount int, count int64) {
	if amount < 0 {
		panic("Negative amount invalid")
	}
	if count <= 0 {
		return
	}
	if _, ok := s.counts[amount]; !ok {
		idx := sort.SearchInts(s.keys, amount)
		s.keys = append(s.keys, 0)
		copy(s.keys[idx+1:], s.keys[idx:])
		s.keys[idx] = amount
	}
	s.counts[amount] += count
	s.observations += count
}

func (s *BucketStatistics) Observations() int64 {
	if s == nil {
		return 0
	}
	return s.observations
}

func (s *BucketStatistics) GetObservations() int64 {
	return s.Observations()
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

func (s *BucketStatistics) GetSamples() []BucketSample {
	return s.Samples()
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

func (s *BucketStatistics) GetMedian() int {
	return s.Median()
}

func (s *BucketStatistics) Mean() float64 {
	return bucketMean(s.Observations(), s.Samples())
}

func (s *BucketStatistics) GetMean() float64 {
	return s.Mean()
}

func (s *BucketStatistics) Min() int {
	return bucketMin(s.Observations(), s.Samples())
}

func (s *BucketStatistics) GetMin() int {
	return s.Min()
}

func (s *BucketStatistics) Max() int {
	return bucketMax(s.Observations(), s.Samples())
}

func (s *BucketStatistics) GetMax() int {
	return s.Max()
}

func (s *BucketStatistics) StdDev() float64 {
	return bucketStdDev(s.Observations(), s.Samples())
}

func (s *BucketStatistics) GetStdDev() float64 {
	return s.StdDev()
}

func (s *BucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s.Observations(), s.Samples(), quantile)
}

func (s *BucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
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

func (s *FixedSizedBucketStatistics) GetObservations() int64 {
	return s.Observations()
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

func (s *FixedSizedBucketStatistics) GetSamples() []BucketSample {
	return s.Samples()
}

func (s *FixedSizedBucketStatistics) Median() int {
	return bucketMedian(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) GetMedian() int {
	return s.Median()
}

func (s *FixedSizedBucketStatistics) Mean() float64 {
	return bucketMean(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) GetMean() float64 {
	return s.Mean()
}

func (s *FixedSizedBucketStatistics) Min() int {
	return bucketMin(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) GetMin() int {
	return s.Min()
}

func (s *FixedSizedBucketStatistics) Max() int {
	return bucketMax(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) GetMax() int {
	return s.Max()
}

func (s *FixedSizedBucketStatistics) StdDev() float64 {
	return bucketStdDev(s.Observations(), s.Samples())
}

func (s *FixedSizedBucketStatistics) GetStdDev() float64 {
	return s.StdDev()
}

func (s *FixedSizedBucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s.Observations(), s.Samples(), quantile)
}

func (s *FixedSizedBucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
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
	b.WriteString("============================%n")
	b.WriteString("=")
	b.WriteString(title)
	b.WriteString("=%n")
	b.WriteString("============================%n")
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
	b.WriteString("numEdges/occurrences (log scale)%n")
	b.WriteString("--------------------------------%n")
	for _, sample := range samples {
		b.WriteString(fmt.Sprintf("%02d:%02d ", sample.Amount, sample.Count))
		for j := 0; j < int(math.Log(float64(sample.Count))); j++ {
			b.WriteByte('#')
		}
		b.WriteString("%n")
	}
	b.WriteString("============================")
	return b.String()
}

type DummyBucketStatistics struct{}

func NewDummyBucketStatistics() *DummyBucketStatistics { return &DummyBucketStatistics{} }

func (s *DummyBucketStatistics) AddSample(amount int)       {}
func (s *DummyBucketStatistics) Observations() int64        { return 0 }
func (s *DummyBucketStatistics) GetObservations() int64     { return 0 }
func (s *DummyBucketStatistics) Samples() []BucketSample    { return nil }
func (s *DummyBucketStatistics) GetSamples() []BucketSample { return nil }
func (s *DummyBucketStatistics) Median() int                { return 0 }
func (s *DummyBucketStatistics) GetMedian() int             { return 0 }
func (s *DummyBucketStatistics) Mean() float64              { return 0 }
func (s *DummyBucketStatistics) GetMean() float64           { return 0 }
func (s *DummyBucketStatistics) Min() int                   { return 0 }
func (s *DummyBucketStatistics) GetMin() int                { return 0 }
func (s *DummyBucketStatistics) Max() int                   { return 0 }
func (s *DummyBucketStatistics) GetMax() int                { return 0 }
func (s *DummyBucketStatistics) StdDev() float64            { return 0 }
func (s *DummyBucketStatistics) GetStdDev() float64         { return 0 }
func (s *DummyBucketStatistics) Percentile(float64) float64 { return 0 }
func (s *DummyBucketStatistics) GetPercentile(float64) float64 {
	return 0
}
func (s *DummyBucketStatistics) AddSamples(any) {}

type ConcurrentBucketStatistics struct {
	mu    sync.Mutex
	stats *BucketStatistics
}

func NewConcurrentBucketStatistics(title string) *ConcurrentBucketStatistics {
	return &ConcurrentBucketStatistics{stats: NewBucketStatistics(title)}
}

func (s *ConcurrentBucketStatistics) AddSample(amount int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.AddSample(amount)
}

func (s *ConcurrentBucketStatistics) Observations() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats.Observations()
}

func (s *ConcurrentBucketStatistics) GetObservations() int64 {
	return s.Observations()
}

func (s *ConcurrentBucketStatistics) Samples() []BucketSample {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats.Samples()
}

func (s *ConcurrentBucketStatistics) GetSamples() []BucketSample {
	return s.Samples()
}

func (s *ConcurrentBucketStatistics) Median() int {
	return bucketMedian(s.Observations(), s.Samples())
}

func (s *ConcurrentBucketStatistics) GetMedian() int { return s.Median() }
func (s *ConcurrentBucketStatistics) Mean() float64 {
	return bucketMean(s.Observations(), s.Samples())
}
func (s *ConcurrentBucketStatistics) GetMean() float64 { return s.Mean() }
func (s *ConcurrentBucketStatistics) Min() int {
	return bucketMin(s.Observations(), s.Samples())
}
func (s *ConcurrentBucketStatistics) GetMin() int { return s.Min() }
func (s *ConcurrentBucketStatistics) Max() int {
	return bucketMax(s.Observations(), s.Samples())
}
func (s *ConcurrentBucketStatistics) GetMax() int { return s.Max() }
func (s *ConcurrentBucketStatistics) StdDev() float64 {
	return bucketStdDev(s.Observations(), s.Samples())
}
func (s *ConcurrentBucketStatistics) GetStdDev() float64 { return s.StdDev() }
func (s *ConcurrentBucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s.Observations(), s.Samples(), quantile)
}
func (s *ConcurrentBucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
}
func (s *ConcurrentBucketStatistics) String() string {
	if s == nil {
		return ""
	}
	return bucketString(s.stats.Title, s.Observations(), s.Samples(), s.Min(), s.Max(), s.Mean(), s.Median(), s.StdDev(), s.Percentile)
}

type FixedSizedConcurrentBucketStatistics struct {
	Title        string
	observations atomic.Int64
	buckets      []atomic.Int64
}

func NewFixedSizedConcurrentBucketStatistics(title string, maximum int) *FixedSizedConcurrentBucketStatistics {
	if maximum <= 0 {
		panic("FixedSizedConcurrentBucketStatistics requires a positive maximum")
	}
	return &FixedSizedConcurrentBucketStatistics{Title: title, buckets: make([]atomic.Int64, maximum)}
}

func (s *FixedSizedConcurrentBucketStatistics) AddSample(amount int) {
	if amount < 0 {
		panic("Negative amount invalid")
	}
	idx := amount
	if idx >= len(s.buckets) {
		idx = len(s.buckets) - 1
	}
	s.buckets[idx].Add(1)
	s.observations.Add(1)
}

func (s *FixedSizedConcurrentBucketStatistics) Observations() int64 {
	if s == nil {
		return 0
	}
	return s.observations.Load()
}

func (s *FixedSizedConcurrentBucketStatistics) GetObservations() int64 {
	return s.Observations()
}

func (s *FixedSizedConcurrentBucketStatistics) Samples() []BucketSample {
	if s == nil {
		return nil
	}
	samples := make([]BucketSample, 0)
	for amount := range s.buckets {
		if count := s.buckets[amount].Load(); count > 0 {
			samples = append(samples, BucketSample{Amount: amount, Count: count})
		}
	}
	return samples
}

func (s *FixedSizedConcurrentBucketStatistics) GetSamples() []BucketSample {
	return s.Samples()
}
func (s *FixedSizedConcurrentBucketStatistics) Median() int {
	return bucketMedian(s.Observations(), s.Samples())
}
func (s *FixedSizedConcurrentBucketStatistics) GetMedian() int { return s.Median() }
func (s *FixedSizedConcurrentBucketStatistics) Mean() float64 {
	return bucketMean(s.Observations(), s.Samples())
}
func (s *FixedSizedConcurrentBucketStatistics) GetMean() float64 { return s.Mean() }
func (s *FixedSizedConcurrentBucketStatistics) Min() int {
	return bucketMin(s.Observations(), s.Samples())
}
func (s *FixedSizedConcurrentBucketStatistics) GetMin() int { return s.Min() }
func (s *FixedSizedConcurrentBucketStatistics) Max() int {
	return bucketMax(s.Observations(), s.Samples())
}
func (s *FixedSizedConcurrentBucketStatistics) GetMax() int { return s.Max() }
func (s *FixedSizedConcurrentBucketStatistics) StdDev() float64 {
	return bucketStdDev(s.Observations(), s.Samples())
}
func (s *FixedSizedConcurrentBucketStatistics) GetStdDev() float64 { return s.StdDev() }
func (s *FixedSizedConcurrentBucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s.Observations(), s.Samples(), quantile)
}
func (s *FixedSizedConcurrentBucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
}
func (s *FixedSizedConcurrentBucketStatistics) String() string {
	if s == nil {
		return ""
	}
	return bucketString(s.Title, s.Observations(), s.Samples(), s.Min(), s.Max(), s.Mean(), s.Median(), s.StdDev(), s.Percentile)
}

type CounterStatistic struct {
	enabled bool
	count   atomic.Int64
}

func NewCounterStatistic(enabled bool) *CounterStatistic {
	return &CounterStatistic{enabled: enabled}
}

func NewCounterStatisticFunc(enabled func() bool) *CounterStatistic {
	if enabled == nil {
		return NewCounterStatistic(false)
	}
	return NewCounterStatistic(enabled())
}

func (s *CounterStatistic) Increment() {
	if s != nil && s.enabled {
		s.count.Add(1)
	}
}

func (s *CounterStatistic) Add(evalCount int64) {
	if s != nil && s.enabled {
		s.count.Add(evalCount)
	}
}

func (s *CounterStatistic) GetCount() int64 {
	if s == nil || !s.enabled {
		return 0
	}
	return s.count.Load()
}

func (s *CounterStatistic) String() string {
	return fmt.Sprint(s.GetCount())
}

type CountDistinctMode int

const (
	CountDistinctNoop CountDistinctMode = iota
	CountDistinctNaive
	CountDistinctHyperLogLog
	CountDistinctSyncedHyperLogLog
)

type CountDistinct struct {
	mode  CountDistinctMode
	fpSet map[uint64]struct{}
	b     int
	m     int
	regs  []int
	mu    sync.Mutex
}

func NewCountDistinctNoop() *CountDistinct {
	return &CountDistinct{mode: CountDistinctNoop}
}

func NewCountDistinctNaive() *CountDistinct {
	return &CountDistinct{mode: CountDistinctNaive, fpSet: make(map[uint64]struct{})}
}

func NewCountDistinctHyperLogLog(b int) *CountDistinct {
	if b <= 0 {
		b = 1
	}
	if b > 30 {
		panic("HyperLogLog register bits must be <= 30")
	}
	m := 1 << b
	return &CountDistinct{mode: CountDistinctHyperLogLog, b: b, m: m, regs: make([]int, m)}
}

func NewCountDistinctSyncedHyperLogLog(b int) *CountDistinct {
	cd := NewCountDistinctHyperLogLog(b)
	cd.mode = CountDistinctSyncedHyperLogLog
	return cd
}

func (c *CountDistinct) AddValue(v Value) {
	if v == nil {
		return
	}
	c.AddHash(v.FingerPrint(FP64New()))
}

func (c *CountDistinct) AddState(state *TLCStateMut) {
	if state == nil {
		return
	}
	c.AddHash(state.FingerPrint())
}

func (c *CountDistinct) AddHash(hash uint64) {
	if c == nil {
		return
	}
	switch c.mode {
	case CountDistinctNaive:
		if c.fpSet == nil {
			c.fpSet = make(map[uint64]struct{})
		}
		c.fpSet[hash] = struct{}{}
	case CountDistinctHyperLogLog:
		c.addHyperLogLog(hash)
	case CountDistinctSyncedHyperLogLog:
		c.mu.Lock()
		c.addHyperLogLog(hash)
		c.mu.Unlock()
	}
}

func (c *CountDistinct) Count() int64 {
	if c == nil {
		return -1
	}
	switch c.mode {
	case CountDistinctNaive:
		return int64(len(c.fpSet))
	case CountDistinctHyperLogLog:
		return c.countHyperLogLog()
	case CountDistinctSyncedHyperLogLog:
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.countHyperLogLog()
	default:
		return -1
	}
}

func (c *CountDistinct) addHyperLogLog(x uint64) {
	if c.m == 0 || c.b <= 0 {
		return
	}
	j := int(x >> (64 - c.b))
	w := bits.LeadingZeros64((x<<c.b)|(uint64(1)<<(c.b-1))) + 1
	if w > c.regs[j] {
		c.regs[j] = w
	}
}

func (c *CountDistinct) countHyperLogLog() int64 {
	if c.m == 0 {
		return 0
	}
	alpha := c.hyperLogLogAlpha()
	z := 0.0
	for _, register := range c.regs {
		z += 1.0 / float64(javaIntOneLeftShift(register))
	}
	estimate := alpha * float64(c.m) * float64(c.m) / z
	if estimate <= 2.5*float64(c.m) {
		zeros := 0
		for _, register := range c.regs {
			if register == 0 {
				zeros++
			}
		}
		if zeros != 0 {
			estimate = float64(c.m) * math.Log(float64(c.m)/float64(zeros))
		}
	} else if estimate > (1.0/30.0)*math.Pow(2, 64) {
		estimate = -math.Pow(2, 64) * math.Log(1-estimate/math.Pow(2, 64))
	}
	return int64(javaDoubleToInt(estimate))
}

func javaIntOneLeftShift(shift int) int32 {
	return int32(uint32(1) << (uint(shift) & 31))
}

func javaDoubleToInt(value float64) int32 {
	if math.IsNaN(value) {
		return 0
	}
	if value <= math.MinInt32 {
		return math.MinInt32
	}
	if value >= math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

func (c *CountDistinct) hyperLogLogAlpha() float64 {
	switch c.m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default:
		return 0.7213 / (1 + 1.079/float64(c.m))
	}
}
