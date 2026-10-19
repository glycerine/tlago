//go:build tlc_fp_stress

package tlc

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

const javaConcurrentFPProperty = "tlc2.tool.fp.MultiThreadedFPSetTest"

// These source static finals capture the original properties once.
var javaConcurrentFPWorkers = func() int {
	if n, ok := initialIntProperty(javaConcurrentFPProperty + ".numThreads"); ok {
		return n
	}
	return runtime.NumCPU()
}()
var javaConcurrentFPInsertions = func() int64 {
	if n, ok := javaConcurrentFPDecodeLong(os.Getenv(javaConcurrentFPProperty + ".insertions")); ok {
		return n
	}
	return 2147483649
}()

// Long.getLong delegates to Long.decode, including # hexadecimal and octal.
func javaConcurrentFPDecodeLong(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	negative := value[0] == '-'
	if negative || value[0] == '+' {
		value = value[1:]
	}
	radix := uint64(10)
	if len(value) > 1 && value[0] == '0' && (value[1] == 'x' || value[1] == 'X') {
		radix, value = 16, value[2:]
	} else if strings.HasPrefix(value, "#") {
		radix, value = 16, value[1:]
	} else if len(value) > 1 && value[0] == '0' {
		radix, value = 8, value[1:]
	}
	if value == "" {
		return 0, false
	}
	limit := uint64(math.MaxInt64)
	if negative {
		limit++
	}
	var result uint64
	for _, char := range value {
		digit := javaIntegerDigit(char)
		if digit < 0 || uint64(digit) >= radix || result > (limit-uint64(digit))/radix {
			return 0, false
		}
		result = result*radix + uint64(digit)
	}
	if negative {
		return -int64(result), true
	}
	return int64(result), true
}

type javaConcurrentFPGenerator struct {
	id               int
	threads          int
	set              FPSet
	rng              *JavaRandom
	total            int64
	quota            int64
	puts             int64
	collisions       int64
	partitionBuckets int64
	increment        int64
	fingerprint      int64
}

func newJavaConcurrentFPGenerator(id, threads int, set FPSet, total int64, kind string) *javaConcurrentFPGenerator {
	g := &javaConcurrentFPGenerator{id: id, threads: threads, set: set, rng: NewJavaRandom(15041980 + int64(id)), total: total, quota: total / int64(threads)}
	if kind == "PartitionedFingerPrintGenerator" {
		buckets := set.GetConfiguration().GetMemoryInFingerprintCnt()
		g.partitionBuckets = buckets / int64(threads)
		g.increment = javaDoubleToLong(math.Ceil(float64(math.MaxInt64-1) / float64(buckets)))
		g.fingerprint = g.increment * (g.partitionBuckets * int64(id))
	}
	return g
}

func (g *javaConcurrentFPGenerator) run(t *testing.T, kind string) {
	t.Helper()
	if kind == "PartitionedFingerPrintGenerator" {
		insertions := int64(0)
		for {
			old := insertions
			insertions++
			if old >= g.partitionBuckets {
				break
			}
			if g.fingerprint != 0 {
				if g.set.Put(uint64(g.fingerprint)) {
					t.Fatal("Linear fill-up should not cause a collision")
				}
				g.puts++
			}
			g.fingerprint += g.increment
		}
		return
	}
	const batch = 1024
	predecessor := int64(0)
	predecessors := make([]int64, batch)
	vector := NewLongVecWithCapacity(batch)
	initialized := false
	// Preserve short-circuit size reads and the original quota/size OR condition.
	for g.puts+g.collisions < g.quota || int64(g.set.Size()) < g.total {
		switch kind {
		case "FingerPrintGenerator":
			if predecessor != 0 && !g.set.Contains(uint64(predecessor)) {
				t.Fatal("predecessor missing")
			}
			predecessor = g.rng.NextLong()
			if !g.set.Put(uint64(predecessor)) {
				g.puts++
			} else {
				g.collisions++
			}
		case "BatchedFingerPrintGenerator":
			if initialized {
				for _, fp := range predecessors {
					if !g.set.Contains(uint64(fp)) {
						t.Fatal("batch predecessor missing")
					}
				}
			}
			for i := range predecessors {
				predecessors[i] = g.rng.NextLong()
			}
			initialized = true
			sort.Slice(predecessors, func(i, j int) bool { return predecessors[i] < predecessors[j] })
			for _, fp := range predecessors {
				if !g.set.Put(uint64(fp)) {
					g.puts++
				} else {
					g.collisions++
				}
			}
		case "LongVecFingerPrintGenerator":
			if initialized {
				bits := g.set.ContainsBlock(vector)
				if got := batch - bits.TrueCount(); got != batch {
					t.Fatalf("batch=%d, want %d", got, batch)
				}
			}
			raw := vector.data[:cap(vector.data)]
			for i := 0; i < batch; i++ {
				raw[i] = g.rng.NextLong()
				// Original private TestLongVec.setElement increments its count twice.
				vector.data = raw[:(len(vector.data)+1)%len(raw)+1]
			}
			initialized = true
			sort.Slice(raw, func(i, j int) bool { return raw[i] < raw[j] })
			bits := g.set.PutBlock(vector)
			g.puts += int64(bits.TrueCount())
			g.collisions += int64(batch - bits.TrueCount())
		}
	}
}

func javaConcurrentFPReport(t *testing.T, set FPSet, start time.Time, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	previous := uint64(0)
	next := start.Add(time.Millisecond)
	for {
		timer := time.NewTimer(time.Until(next))
		select {
		case <-stop:
			timer.Stop()
			return
		case <-timer.C:
			current := set.Size()
			stats := set.(interface{ GetLoadFactor() float64 })
			t.Logf("%d s (epoch); %s insertions/min; %g load factor", time.Now().UnixMilli(), javaLongFPSetInteger(int64(current-previous)), stats.GetLoadFactor())
			previous = current
			next = next.Add(time.Minute)
		}
	}
}

// Complete MultiThreadedFPSetTest.doTest with its source generator factories.
// Failed producer assertions do not count down the latch in the original.
func javaConcurrentFPDoTest(t *testing.T, implementation, kind string) FPSet {
	t.Helper()
	if excluded, _ := tlcLookupSystemProperty(javaConcurrentFPProperty + ".excludes"); strings.Contains(excluded, "_"+kind) {
		t.Skip("source Assume: excluded generator " + kind)
	}
	threads, total := javaConcurrentFPWorkers, javaConcurrentFPInsertions
	t.Logf("Insertions: %s (approx: %s GiB)", javaLongFPSetInteger(total), javaLongFPSetInteger(total*fpSetLongSize>>20))
	t.Logf("Thread count: %d; Running test: %s", threads, kind)
	previousWorkers := NumWorkers()
	SetNumWorkers(threads)
	t.Cleanup(func() { SetNumWorkers(previousWorkers) })
	config := NewFPSetConfiguration()
	var set FPSet
	switch implementation {
	case "lsb":
		config.SetRatio(1.0)
		set = NewLSBDiskFPSet(config)
	case "msb":
		config.SetRatio(1.0)
		set = NewMSBDiskFPSet(config)
	case "offheap":
		set = NewOffHeapDiskFPSet(NewFPSetConfigurationWithRatioAndImplementation(1.0, "tlc2.tool.fp.OffHeapDiskFPSet"))
	}
	set.Init(threads, t.TempDir(), "FPSetTestTest")
	t.Cleanup(set.Close)
	stats := set.(interface {
		GetMaxTblCnt() int64
		GetLockCnt() int
		GetTblCapacity() int64
		GetLoadFactor() float64
	})
	t.Logf("Maximum FPSet table count is: %s (approx: %s GiB)", javaLongFPSetInteger(stats.GetMaxTblCnt()), javaLongFPSetInteger(stats.GetMaxTblCnt()*fpSetLongSize>>20))
	t.Logf("FPSet lock count is: %d; FPSet bucket count is: %d", stats.GetLockCnt(), stats.GetTblCapacity())
	set.IncWorkers(threads)
	var latch, arrived sync.WaitGroup
	latch.Add(threads)
	arrived.Add(threads)
	startBarrier := make(chan struct{})
	generators := make([]*javaConcurrentFPGenerator, threads)
	for i := range generators {
		g := newJavaConcurrentFPGenerator(i, threads, set, total, kind)
		generators[i] = g
		go func() {
			restore := PushCurrentWorkerID(g.id)
			defer restore()
			// Java prints an uncaught worker error's stack, then ends that thread
			// without latch.countDown, even when its detail message is empty.
			defer func() {
				if failure := recover(); failure != nil {
					t.Errorf("Producer#%d: %s", g.id, javaThrowableStackTrace(panicValueAsError(failure)))
				}
			}()
			arrived.Done()
			<-startBarrier
			g.run(t, kind)
			latch.Done()
		}()
	}
	arrived.Wait()
	start := time.Now()
	stopReporter, reporterDone := make(chan struct{}), make(chan struct{})
	go javaConcurrentFPReport(t, set, start, stopReporter, reporterDone)
	close(startBarrier)
	latch.Wait()
	end := time.Now()
	close(stopReporter)
	<-reporterDone
	var overallPuts, overallCollisions int64
	for _, g := range generators {
		ratio := "none"
		if g.collisions != 0 {
			ratio = fmt.Sprint(float64(g.puts / g.collisions))
		}
		t.Logf("Producer: %d, puts: %d, puts/collisions: %s", g.id, g.puts, ratio)
		overallPuts += g.puts
		overallCollisions += g.collisions
	}
	t.Logf("Total puts: %d, total collisions: %d, total load factor: %g, duration: %d ms.", overallPuts, overallCollisions, stats.GetLoadFactor(), end.Sub(start).Milliseconds())
	duration := max(end.Sub(start).Milliseconds(), int64(1))
	t.Logf("%d s; %s insertions/min; %g load factor", time.Now().UnixMilli(), javaLongFPSetInteger(int64(set.Size())/duration*60000), stats.GetLoadFactor())
	if got := set.Size(); int64(got) != overallPuts {
		t.Fatalf("size=%d, want overallPuts=%d", got, overallPuts)
	}
	if got := int64(set.Size()); got < total {
		t.Fatalf("size=%d, want >= %d", got, total)
	}
	if got := int64(set.Size()); got > total+int64(threads) {
		t.Fatalf("size=%d, want <= %d", got, total+int64(threads))
	}
	if !set.CheckInvariant() {
		t.Fatal("checkInvariant failed")
	}
	return set
}

func javaConcurrentFPClass(t *testing.T, implementation string) {
	for _, method := range []struct{ name, kind string }{
		{"testMaxFPSetSizeRndBatched", "BatchedFingerPrintGenerator"},
		{"testMaxFPSetSizeRndBlock", "LongVecFingerPrintGenerator"},
		{"testMaxFPSetSizeRnd", "FingerPrintGenerator"},
		{"testMaxFPSetSizePartitioned", "PartitionedFingerPrintGenerator"},
	} {
		t.Run(method.name, func(t *testing.T) {
			set := javaConcurrentFPDoTest(t, implementation, method.kind)
			if implementation == "offheap" && method.kind == "PartitionedFingerPrintGenerator" {
				if got := set.(*OffHeapDiskFPSet).GetBucketCapacity(); got != 0 {
					t.Fatalf("bucket capacity=%d, want 0", got)
				}
			}
		})
	}
}
func TestJavaConcurrentLSBDiskFPSet(t *testing.T) {
	skipSlowTest(t)
	javaConcurrentFPClass(t, "lsb")
}
func TestJavaConcurrentMSBDiskFPSet(t *testing.T) {
	skipSlowTest(t)
	javaConcurrentFPClass(t, "msb")
}
func TestJavaConcurrentOffHeapDiskFPSet(t *testing.T) {
	skipSlowTest(t)
	javaConcurrentFPClass(t, "offheap")
}
