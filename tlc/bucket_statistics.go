package tlc

import (
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type BucketSample struct {
	Amount int
	Count  int64
}

// BucketStatisticsView is the Go representation of Java's IBucketStatistics
// surface consumed by BucketStatistics.add.
type BucketStatisticsView interface {
	Observations() int64
	Samples() []BucketSample
}

type BucketStatistics struct {
	Title          string
	titleCaptured  bool
	titleOwner     string
	observations   int64
	counts         map[int]int64
	keys           []int
	managementBean *BucketStatisticsMXWrapper
}

func (s *BucketStatistics) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func NewBucketStatistics(title string) *BucketStatistics {
	return &BucketStatistics{
		Title: title, titleCaptured: true, titleOwner: title,
		counts: make(map[int]int64),
	}
}

func NewBucketStatisticsWithMX(title string, pkg string, name string) *BucketStatistics {
	stats := NewBucketStatistics(title)
	stats.managementBean = NewBucketStatisticsMXWrapper(stats, name, pkg)
	return stats
}

func (s *BucketStatistics) AddSample(amount int) {
	s.require()
	amount = int(int32(amount))
	if amount < 0 {
		panic(NewIllegalArgumentException("Negative amount invalid"))
	}
	s.AddSampleCount(amount, 1)
}

func (s *BucketStatistics) AddSampleCount(amount int, count int64) {
	s.require()
	amount = int(int32(amount))
	if amount < 0 {
		panic(NewIllegalArgumentException("Negative amount invalid"))
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
	s.require()
	return s.observations
}

func (s *BucketStatistics) GetObservations() int64 {
	return s.Observations()
}

func (s *BucketStatistics) Samples() []BucketSample {
	s.require()
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

func (s *BucketStatistics) Add(other BucketStatisticsView) {
	if s == nil || isNil(other) {
		panic(NewNullPointerException())
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
	return bucketMedian(s)
}

func (s *BucketStatistics) GetMedian() int {
	return s.Median()
}

func (s *BucketStatistics) Mean() float64 {
	return bucketMean(s)
}

func (s *BucketStatistics) GetMean() float64 {
	return s.Mean()
}

func (s *BucketStatistics) Min() int {
	return bucketMin(s)
}

func (s *BucketStatistics) GetMin() int {
	return s.Min()
}

func (s *BucketStatistics) Max() int {
	return bucketMax(s)
}

func (s *BucketStatistics) GetMax() int {
	return s.Max()
}

func (s *BucketStatistics) StdDev() float64 {
	return bucketStdDev(s, s.Mean)
}

func (s *BucketStatistics) GetStdDev() float64 {
	return s.StdDev()
}

func (s *BucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s, quantile)
}

func (s *BucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
}

func (s *BucketStatistics) String() string {
	s.require()
	return bucketString(s.ownedTitle(), s, s.Min, s.Max, s.Mean, s.Median, s.StdDev, s.Percentile)
}

func (s *BucketStatistics) ownedTitle() string {
	if s.titleCaptured {
		return s.titleOwner
	}
	return s.Title
}

type FixedSizedBucketStatistics struct {
	Title         string
	titleCaptured bool
	titleOwner    string
	observations  int64
	buckets       []int64
}

func (s *FixedSizedBucketStatistics) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func NewFixedSizedBucketStatistics(title string, maximum int) *FixedSizedBucketStatistics {
	maximum = int(int32(maximum))
	if maximum < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(maximum)))
	}
	return &FixedSizedBucketStatistics{
		Title: title, titleCaptured: true, titleOwner: title,
		buckets: make([]int64, maximum),
	}
}

func (s *FixedSizedBucketStatistics) AddSample(amount int) {
	s.require()
	amount = int(int32(amount))
	if amount < 0 {
		panic(NewIllegalArgumentException("Negative amount invalid"))
	}
	idx := amount
	if idx >= len(s.buckets) {
		idx = len(s.buckets) - 1
	}
	if idx < 0 {
		panic(NewArrayIndexOutOfBoundsException(idx, len(s.buckets)))
	}
	s.buckets[idx]++
	s.observations++
}

func (s *FixedSizedBucketStatistics) Observations() int64 {
	s.require()
	return s.observations
}

func (s *FixedSizedBucketStatistics) GetObservations() int64 {
	return s.Observations()
}

func (s *FixedSizedBucketStatistics) Samples() []BucketSample {
	s.require()
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
	return bucketMedian(s)
}

func (s *FixedSizedBucketStatistics) GetMedian() int {
	return s.Median()
}

func (s *FixedSizedBucketStatistics) Mean() float64 {
	return bucketMean(s)
}

func (s *FixedSizedBucketStatistics) GetMean() float64 {
	return s.Mean()
}

func (s *FixedSizedBucketStatistics) Min() int {
	return bucketMin(s)
}

func (s *FixedSizedBucketStatistics) GetMin() int {
	return s.Min()
}

func (s *FixedSizedBucketStatistics) Max() int {
	return bucketMax(s)
}

func (s *FixedSizedBucketStatistics) GetMax() int {
	return s.Max()
}

func (s *FixedSizedBucketStatistics) StdDev() float64 {
	return bucketStdDev(s, s.Mean)
}

func (s *FixedSizedBucketStatistics) GetStdDev() float64 {
	return s.StdDev()
}

func (s *FixedSizedBucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s, quantile)
}

func (s *FixedSizedBucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
}

func (s *FixedSizedBucketStatistics) String() string {
	s.require()
	return bucketString(s.ownedTitle(), s, s.Min, s.Max, s.Mean, s.Median, s.StdDev, s.Percentile)
}

func (s *FixedSizedBucketStatistics) ownedTitle() string {
	if s.titleCaptured {
		return s.titleOwner
	}
	return s.Title
}

func bucketMedian(stats BucketStatisticsView) int {
	observations := stats.Observations()
	if observations <= 0 {
		return -1
	}
	sum := int64(0)
	for _, sample := range stats.Samples() {
		sum += sample.Count
		if sum > observations/2 {
			return sample.Amount
		}
	}
	panic(NewRuntimeException("bug, shoud not get here"))
}

func bucketMean(stats BucketStatisticsView) float64 {
	sum := int64(0)
	for _, sample := range stats.Samples() {
		sum += sample.Count * int64(sample.Amount)
	}
	if stats.Observations() > 0 {
		return float64(sum) / float64(stats.Observations())
	}
	return -1
}

func bucketMin(stats BucketStatisticsView) int {
	if stats.Observations() <= 0 {
		return -1
	}
	samples := stats.Samples()
	if len(samples) == 0 {
		panic(NewNoSuchElementException())
	}
	return samples[0].Amount
}

func bucketMax(stats BucketStatisticsView) int {
	if stats.Observations() <= 0 {
		return -1
	}
	samples := stats.Samples()
	if len(samples) == 0 {
		panic(NewNoSuchElementException())
	}
	return samples[len(samples)-1].Amount
}

func bucketStdDev(stats BucketStatisticsView, meanValue func() float64) float64 {
	observations := stats.Observations()
	if observations <= 0 {
		return -1
	}
	mean := meanValue()
	sum := 0.0
	for _, sample := range stats.Samples() {
		diff := float64(sample.Amount) - mean
		sum += diff * diff * float64(sample.Count)
	}
	return math.Sqrt(sum / float64(observations))
}

func bucketPercentile(stats BucketStatisticsView, quantile float64) float64 {
	if math.IsNaN(quantile) {
		panic(NewIllegalArgumentException("NaN"))
	}
	observations := stats.Observations()
	if observations <= 0 {
		return -1
	}
	quantile = math.Min(1, quantile)
	quantile = math.Max(0, quantile)
	samples := stats.Samples()
	pos := javaDoubleToInt(float64(observations) * quantile)
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

func bucketString(title string, stats BucketStatisticsView, minValue func() int, maxValue func() int, meanValue func() float64, medianValue func() int, stddevValue func() float64, percentile func(float64) float64) string {
	var b strings.Builder
	b.WriteString("============================%n")
	b.WriteString("=")
	b.WriteString(title)
	b.WriteString("=%n")
	b.WriteString("============================%n")
	b.WriteString("Observations: " + bucketFormatNumber(fmt.Sprintf("%d", stats.Observations())) + javaFormatLineSeparator())
	b.WriteString("Min: " + bucketFormatNumber(fmt.Sprintf("%d", minValue())) + javaFormatLineSeparator())
	b.WriteString("Max: " + bucketFormatNumber(fmt.Sprintf("%d", maxValue())) + javaFormatLineSeparator())
	b.WriteString("Mean: " + bucketFormatFixed2(meanValue()) + javaFormatLineSeparator())
	b.WriteString("Median: " + bucketFormatNumber(fmt.Sprintf("%d", medianValue())) + javaFormatLineSeparator())
	b.WriteString("Standard deviation: " + bucketFormatFixed2(stddevValue()) + javaFormatLineSeparator())
	b.WriteString("75%: " + bucketFormatFixed2(percentile(0.75)) + javaFormatLineSeparator())
	b.WriteString("95%: " + bucketFormatFixed2(percentile(0.95)) + javaFormatLineSeparator())
	b.WriteString("98%: " + bucketFormatFixed2(percentile(0.98)) + javaFormatLineSeparator())
	b.WriteString("99%: " + bucketFormatFixed2(percentile(0.99)) + javaFormatLineSeparator())
	b.WriteString("99.9%: " + bucketFormatFixed2(percentile(0.999)) + javaFormatLineSeparator())
	b.WriteString("numEdges/occurrences (log scale)%n")
	b.WriteString("--------------------------------%n")
	for _, sample := range stats.Samples() {
		b.WriteString(bucketFormatNumber(fmt.Sprintf("%02d:%02d ", sample.Amount, sample.Count)))
		for j := 0; float64(j) < math.Log(float64(sample.Count)); j++ {
			b.WriteByte('#')
		}
		b.WriteString("%n")
	}
	b.WriteString("============================")
	return b.String()
}

type DummyBucketStatistics struct{}

func NewDummyBucketStatistics() *DummyBucketStatistics { return &DummyBucketStatistics{} }

func (s *DummyBucketStatistics) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *DummyBucketStatistics) AddSample(amount int) { s.require() }
func (s *DummyBucketStatistics) Observations() int64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) GetObservations() int64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) Samples() []BucketSample {
	s.require()
	return []BucketSample{}
}
func (s *DummyBucketStatistics) GetSamples() []BucketSample {
	return s.Samples()
}
func (s *DummyBucketStatistics) Median() int {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) GetMedian() int {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) Mean() float64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) GetMean() float64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) Min() int {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) GetMin() int {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) Max() int {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) GetMax() int {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) StdDev() float64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) GetStdDev() float64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) Percentile(float64) float64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) GetPercentile(float64) float64 {
	s.require()
	return 0
}
func (s *DummyBucketStatistics) AddSamples(BucketStatisticsView) {
	s.require()
}

type ConcurrentBucketStatistics struct {
	Title          string
	titleCaptured  bool
	titleOwner     string
	observations   atomic.Int64
	buckets        sync.Map // map[int]*atomic.Int64
	managementBean *BucketStatisticsMXWrapper
}

func (s *ConcurrentBucketStatistics) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func NewConcurrentBucketStatistics(title string) *ConcurrentBucketStatistics {
	return &ConcurrentBucketStatistics{Title: title, titleCaptured: true, titleOwner: title}
}

func NewConcurrentBucketStatisticsWithMX(title string, pkg string, name string) *ConcurrentBucketStatistics {
	stats := NewConcurrentBucketStatistics(title)
	stats.managementBean = NewBucketStatisticsMXWrapper(stats, name, pkg)
	return stats
}

func bucketStatisticsAddSample(stats any, amount int) {
	switch s := stats.(type) {
	case *BucketStatistics:
		s.AddSample(amount)
	case *FixedSizedBucketStatistics:
		s.AddSample(amount)
	case *ConcurrentBucketStatistics:
		s.AddSample(amount)
	case *FixedSizedConcurrentBucketStatistics:
		s.AddSample(amount)
	case *DummyBucketStatistics:
		s.AddSample(amount)
	}
}

func (s *ConcurrentBucketStatistics) AddSample(amount int) {
	s.require()
	amount = int(int32(amount))
	if amount < 0 {
		panic(NewIllegalArgumentException("Negative amount invalid"))
	}
	if bucket, ok := s.buckets.Load(amount); ok {
		bucket.(*atomic.Int64).Add(1)
	} else {
		// Java deliberately ignores putIfAbsent's return value. Two first
		// insertions can therefore both count as observations while only the
		// winning bucket retains its initial one.
		counter := &atomic.Int64{}
		counter.Store(1)
		s.buckets.LoadOrStore(amount, counter)
	}
	s.observations.Add(1)
}

func (s *ConcurrentBucketStatistics) Observations() int64 {
	s.require()
	return s.observations.Load()
}

func (s *ConcurrentBucketStatistics) GetObservations() int64 {
	return s.Observations()
}

func (s *ConcurrentBucketStatistics) Samples() []BucketSample {
	s.require()
	samples := make([]BucketSample, 0)
	s.buckets.Range(func(key, value any) bool {
		count := value.(*atomic.Int64).Load()
		if count > 0 {
			samples = append(samples, BucketSample{Amount: key.(int), Count: count})
		}
		return true
	})
	sort.Slice(samples, func(i, j int) bool { return samples[i].Amount < samples[j].Amount })
	return samples
}

func (s *ConcurrentBucketStatistics) GetSamples() []BucketSample {
	return s.Samples()
}

func (s *ConcurrentBucketStatistics) Median() int {
	return bucketMedian(s)
}

func (s *ConcurrentBucketStatistics) GetMedian() int { return s.Median() }
func (s *ConcurrentBucketStatistics) Mean() float64 {
	return bucketMean(s)
}
func (s *ConcurrentBucketStatistics) GetMean() float64 { return s.Mean() }
func (s *ConcurrentBucketStatistics) Min() int {
	return bucketMin(s)
}
func (s *ConcurrentBucketStatistics) GetMin() int { return s.Min() }
func (s *ConcurrentBucketStatistics) Max() int {
	return bucketMax(s)
}
func (s *ConcurrentBucketStatistics) GetMax() int { return s.Max() }
func (s *ConcurrentBucketStatistics) StdDev() float64 {
	return bucketStdDev(s, s.Mean)
}
func (s *ConcurrentBucketStatistics) GetStdDev() float64 { return s.StdDev() }
func (s *ConcurrentBucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s, quantile)
}
func (s *ConcurrentBucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
}
func (s *ConcurrentBucketStatistics) String() string {
	s.require()
	return bucketString(s.ownedTitle(), s, s.Min, s.Max, s.Mean, s.Median, s.StdDev, s.Percentile)
}

func (s *ConcurrentBucketStatistics) ownedTitle() string {
	if s.titleCaptured {
		return s.titleOwner
	}
	return s.Title
}

type FixedSizedConcurrentBucketStatistics struct {
	Title         string
	titleCaptured bool
	titleOwner    string
	observations  atomic.Int64
	buckets       []atomic.Int64
}

func (s *FixedSizedConcurrentBucketStatistics) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func NewFixedSizedConcurrentBucketStatistics(title string, maximum int) *FixedSizedConcurrentBucketStatistics {
	maximum = int(int32(maximum))
	if maximum < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(maximum)))
	}
	return &FixedSizedConcurrentBucketStatistics{
		Title: title, titleCaptured: true, titleOwner: title,
		buckets: make([]atomic.Int64, maximum),
	}
}

func (s *FixedSizedConcurrentBucketStatistics) AddSample(amount int) {
	s.require()
	amount = int(int32(amount))
	if amount < 0 {
		panic(NewIllegalArgumentException("Negative amount invalid"))
	}
	idx := amount
	if idx >= len(s.buckets) {
		idx = len(s.buckets) - 1
	}
	if idx < 0 {
		panic(NewArrayIndexOutOfBoundsException(idx, len(s.buckets)))
	}
	s.buckets[idx].Add(1)
	s.observations.Add(1)
}

func (s *FixedSizedConcurrentBucketStatistics) Observations() int64 {
	s.require()
	return s.observations.Load()
}

func (s *FixedSizedConcurrentBucketStatistics) GetObservations() int64 {
	return s.Observations()
}

func (s *FixedSizedConcurrentBucketStatistics) Samples() []BucketSample {
	s.require()
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
	return bucketMedian(s)
}
func (s *FixedSizedConcurrentBucketStatistics) GetMedian() int { return s.Median() }
func (s *FixedSizedConcurrentBucketStatistics) Mean() float64 {
	return bucketMean(s)
}
func (s *FixedSizedConcurrentBucketStatistics) GetMean() float64 { return s.Mean() }
func (s *FixedSizedConcurrentBucketStatistics) Min() int {
	return bucketMin(s)
}
func (s *FixedSizedConcurrentBucketStatistics) GetMin() int { return s.Min() }
func (s *FixedSizedConcurrentBucketStatistics) Max() int {
	return bucketMax(s)
}
func (s *FixedSizedConcurrentBucketStatistics) GetMax() int { return s.Max() }
func (s *FixedSizedConcurrentBucketStatistics) StdDev() float64 {
	return bucketStdDev(s, s.Mean)
}
func (s *FixedSizedConcurrentBucketStatistics) GetStdDev() float64 { return s.StdDev() }
func (s *FixedSizedConcurrentBucketStatistics) Percentile(quantile float64) float64 {
	return bucketPercentile(s, quantile)
}
func (s *FixedSizedConcurrentBucketStatistics) GetPercentile(quantile float64) float64 {
	return s.Percentile(quantile)
}
func (s *FixedSizedConcurrentBucketStatistics) String() string {
	s.require()
	return bucketString(s.ownedTitle(), s, s.Min, s.Max, s.Mean, s.Median, s.StdDev, s.Percentile)
}

func (s *FixedSizedConcurrentBucketStatistics) ownedTitle() string {
	if s.titleCaptured {
		return s.titleOwner
	}
	return s.Title
}

type CounterStatistic struct {
	enabled  bool
	count    atomic.Int64
	identity uint32
}

func NewCounterStatistic(enabled bool) *CounterStatistic {
	counter := &CounterStatistic{enabled: enabled}
	counter.identity = uint32(reflect.ValueOf(counter).Pointer())
	return counter
}

func NewCounterStatisticFunc(enabled func() bool) *CounterStatistic {
	if enabled == nil {
		panic(NewNullPointerException())
	}
	return NewCounterStatistic(enabled())
}

func (s *CounterStatistic) Increment() {
	s.require()
	if s.enabled {
		s.count.Add(1)
	}
}

func (s *CounterStatistic) Add(evalCount int64) {
	s.require()
	if s.enabled {
		s.count.Add(evalCount)
	}
}

func (s *CounterStatistic) GetCount() int64 {
	s.require()
	if !s.enabled {
		return 0
	}
	return s.count.Load()
}

func (s *CounterStatistic) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *CounterStatistic) String() string {
	s.require()
	if !s.enabled {
		return fmt.Sprintf("tlc2.util.statistics.CounterStatistic$NoopCounterStatistic@%x", s.identity)
	}
	return fmt.Sprint(s.count.Load())
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

func (c *CountDistinct) require() {
	if c == nil {
		panic(NewNullPointerException())
	}
}

func NewCountDistinctNoop() *CountDistinct {
	return &CountDistinct{mode: CountDistinctNoop}
}

func NewCountDistinctNaive() *CountDistinct {
	return &CountDistinct{mode: CountDistinctNaive, fpSet: make(map[uint64]struct{})}
}

func NewCountDistinctHyperLogLog(b int) *CountDistinct {
	m32 := int32(uint32(1) << (uint(b) & 31))
	if m32 < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(m32)))
	}
	m := int(m32)
	return &CountDistinct{mode: CountDistinctHyperLogLog, b: b, m: m, regs: make([]int, m)}
}

func NewCountDistinctSyncedHyperLogLog(b int) *CountDistinct {
	cd := NewCountDistinctHyperLogLog(b)
	cd.mode = CountDistinctSyncedHyperLogLog
	return cd
}

func (c *CountDistinct) AddValue(v Value) {
	c.require()
	if c.mode == CountDistinctNoop {
		return
	}
	if isNil(v) {
		panic(NewNullPointerException())
	}
	c.AddHash(v.FingerPrint(FP64New()))
}

func (c *CountDistinct) AddState(state *TLCStateMut) {
	c.require()
	if c.mode == CountDistinctNoop {
		return
	}
	if state == nil {
		panic(NewNullPointerException())
	}
	c.AddHash(state.FingerPrint())
}

func (c *CountDistinct) AddHash(hash uint64) {
	c.require()
	switch c.mode {
	case CountDistinctNaive:
		// Java's Naive implementation delegates to MemFPSet, whose put and
		// size methods share the set monitor.
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.fpSet == nil {
			c.fpSet = make(map[uint64]struct{})
		}
		c.fpSet[hash] = struct{}{}
	case CountDistinctHyperLogLog:
		c.addHyperLogLog(hash)
	case CountDistinctSyncedHyperLogLog:
		j, w := c.hyperLogLogRegister(hash)
		c.mu.Lock()
		defer c.mu.Unlock()
		c.updateHyperLogLogRegister(j, w)
	}
}

func (c *CountDistinct) Count() int64 {
	c.require()
	switch c.mode {
	case CountDistinctNaive:
		c.mu.Lock()
		defer c.mu.Unlock()
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
	c.require()
	j, w := c.hyperLogLogRegister(x)
	c.updateHyperLogLogRegister(j, w)
}

func (c *CountDistinct) hyperLogLogRegister(x uint64) (int, int) {
	j := int(int32(uint32(x >> (uint(64-c.b) & 63))))
	w := bits.LeadingZeros64((x<<(uint(c.b)&63))|(uint64(1)<<(uint(c.b-1)&63))) + 1
	return j, w
}

func (c *CountDistinct) updateHyperLogLogRegister(j int, w int) {
	if j < 0 || j >= len(c.regs) {
		panic(NewArrayIndexOutOfBoundsException(j, len(c.regs)))
	}
	if w > c.regs[j] {
		c.regs[j] = w
	}
}

func (c *CountDistinct) countHyperLogLog() int64 {
	c.require()
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
	c.require()
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

// AbstractBucketStatistics uses Formatter rather than Go's binary tie-to-even
// formatting. Reuse the worker formatter's decimal half-up rounding, without
// grouping, and retain Formatter's ASCII sign and special-value spellings.
func bucketFormatFixed2(value float64) string {
	if math.IsNaN(value) {
		return "NaN"
	}
	if math.IsInf(value, 1) {
		return "Infinity"
	}
	if math.IsInf(value, -1) {
		return "-Infinity"
	}
	mpNumberSymbolsOnce.Do(initializeMessageNumberSymbols)
	symbols := mpNumberSymbols
	symbols.group = ""
	text := distributedWorkerCacheRatioWithSymbols(math.Abs(value), symbols, distributedCacheDecimalSeparator(mpNumberSymbols, mpNumberLocaleKey))
	if math.Signbit(value) {
		return "-" + text
	}
	return text
}

func bucketFormatNumber(text string) string {
	mpNumberSymbolsOnce.Do(initializeMessageNumberSymbols)
	var out strings.Builder
	for _, char := range text {
		if char >= '0' && char <= '9' {
			char = mpNumberSymbols.zero + char - '0'
		}
		out.WriteRune(char)
	}
	return out.String()
}
