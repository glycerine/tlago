package tlc

import "testing"

type javaRandomSetOfSubsetsBenchmarkCase struct {
	name  string
	size  int32
	picks int
	exact bool
}

var javaRandomSetOfSubsetsBenchmarkCases = []javaRandomSetOfSubsetsBenchmarkCase{
	{name: "randomSetOfSubsetExact024k008", size: 24, picks: 1<<8 - 1, exact: true},
	{name: "randomSetOfSubsetExact024k012", size: 24, picks: 1<<12 - 1, exact: true},
	{name: "randomSetOfSubsetExact024k016", size: 24, picks: 1<<16 - 1, exact: true},
	{name: "randomSetOfSubsetExact031k008", size: 31, picks: 1<<8 - 1, exact: true},
	{name: "randomSetOfSubsetExact031k012", size: 31, picks: 1<<12 - 1, exact: true},
	{name: "randomSetOfSubsetExact031k016", size: 31, picks: 1<<16 - 1, exact: true},
	{name: "randomSetOfSubsetExact65k008", size: 65, picks: 1<<8 - 1, exact: true},
	{name: "randomSetOfSubsetExact065k012", size: 65, picks: 1<<12 - 1, exact: true},
	{name: "randomSetOfSubsetExact065k016", size: 65, picks: 1<<16 - 1, exact: true},
	{name: "randomSetOfSubsetExact150k008", size: 150, picks: 1<<8 - 1, exact: true},
	{name: "randomSetOfSubsetExact150k012", size: 150, picks: 1<<12 - 1, exact: true},
	{name: "randomSetOfSubsetExact150k016", size: 150, picks: 1<<16 - 1, exact: true},
	{name: "randomSetOfSubsetExact268k008", size: 268, picks: 1<<8 - 1, exact: true},
	{name: "randomSetOfSubsetExact268k012", size: 268, picks: 1<<12 - 1, exact: true},
	{name: "randomSetOfSubsetExact268k016", size: 268, picks: 1<<16 - 1, exact: true},
	{name: "randomSetOfSubset024k008", size: 24, picks: 1<<8 - 1},
	{name: "randomSetOfSubset024k012", size: 24, picks: 1<<12 - 1},
	{name: "randomSetOfSubset024k016", size: 24, picks: 1<<16 - 1},
	{name: "randomSetOfSubset031k008", size: 31, picks: 1<<8 - 1},
	{name: "randomSetOfSubset031k012", size: 31, picks: 1<<12 - 1},
	{name: "randomSetOfSubset031k016", size: 31, picks: 1<<16 - 1},
	{name: "randomSetOfSubset065k008", size: 65, picks: 1<<8 - 1},
	{name: "randomSetOfSubset065k012", size: 65, picks: 1<<12 - 1},
	{name: "randomSetOfSubset065k016", size: 65, picks: 1<<16 - 1},
	{name: "randomSetOfSubset150k008", size: 150, picks: 1<<8 - 1},
	{name: "randomSetOfSubset150k012", size: 150, picks: 1<<12 - 1},
	{name: "randomSetOfSubset150k016", size: 150, picks: 1<<16 - 1},
	{name: "randomSetOfSubset268k008", size: 268, picks: 1<<8 - 1},
	{name: "randomSetOfSubset268k012", size: 268, picks: 1<<12 - 1},
	{name: "randomSetOfSubset268k016", size: 268, picks: 1<<16 - 1},
}

type javaRandomIntervalBenchmarkCase struct {
	name string
	high int32
	kind byte
}

var javaRandomIntervalBenchmarkCases = []javaRandomIntervalBenchmarkCase{
	{name: "randomInterval016setBound", high: 2 << 16, kind: 'b'},
	{name: "randomInterval020setBound", high: 2 << 20, kind: 'b'},
	{name: "randomInterval024setBound", high: 2 << 24, kind: 'b'},
	{name: "randomInterval028setBound", high: 2 << 28, kind: 'b'},
	{name: "randomInterval031setBound", high: 1<<31 - 1, kind: 'b'},
	{name: "randomIntervalt016all", high: 2 << 16, kind: 'a'},
	{name: "randomInterval020all", high: 2 << 20, kind: 'a'},
	{name: "randomInterval024all", high: 2 << 24, kind: 'a'},
	{name: "randomIntervalt016half", high: 2 << 16, kind: 'h'},
	{name: "randomInterval020half", high: 2 << 20, kind: 'h'},
	{name: "randomInterval024half", high: 2 << 24, kind: 'h'},
}

type javaRandomSetEnumBenchmarkCase struct {
	name string
	size int
}

var javaRandomSetEnumBenchmarkCases = []javaRandomSetEnumBenchmarkCase{
	{name: "randomSubset016", size: 16},
	{name: "randomSubset032", size: 32},
	{name: "randomSubset064", size: 64},
	{name: "randomSubset128", size: 128},
	{name: "randomSubset256", size: 256},
	{name: "randomSubset512", size: 512},
	{name: "randomSubset1024", size: 1024},
	{name: "randomSubset2048", size: 2048},
	{name: "randomSubset4096", size: 4096},
	{name: "randomSubset8192", size: 8192},
	{name: "randomSubset16384", size: 16384},
	{name: "randomSubset32768", size: 32768},
	{name: "randomSubsetTLCBound", size: 1000000},
}

type javaRandomFunctionSetBenchmarkCase struct {
	name                  string
	domainSize, rangeSize int32
	picks                 int
}

var javaRandomFunctionSetBenchmarkCases = []javaRandomFunctionSetBenchmarkCase{
	{name: "randomSubsetFcns008x008p16", domainSize: 8, rangeSize: 8, picks: 1<<16 - 1},
	{name: "randomSubsetFcns008x008p08", domainSize: 8, rangeSize: 8, picks: 1<<8 - 1},
	{name: "randomSubsetFcns008x008p12", domainSize: 8, rangeSize: 8, picks: 1<<12 - 1},
	// These two p16 methods deliberately retain the Java source's surprising targets.
	{name: "randomSubsetFcns011x011p16", domainSize: 8, rangeSize: 8, picks: 1<<16 - 1},
	{name: "randomSubsetFcns011x011p08", domainSize: 11, rangeSize: 11, picks: 1<<8 - 1},
	{name: "randomSubsetFcns011x011p12", domainSize: 11, rangeSize: 11, picks: 1<<12 - 1},
	{name: "randomSubsetFcns016x008p16", domainSize: 11, rangeSize: 11, picks: 1<<16 - 1},
	{name: "randomSubsetFcns016x008p08", domainSize: 16, rangeSize: 8, picks: 1<<8 - 1},
	{name: "randomSubsetFcns016x008p12", domainSize: 16, rangeSize: 8, picks: 1<<12 - 1},
	{name: "randomSubsetFcns016x016p16", domainSize: 16, rangeSize: 16, picks: 1<<16 - 1},
	{name: "randomSubsetFcns016x016p08", domainSize: 16, rangeSize: 16, picks: 1<<8 - 1},
	{name: "randomSubsetFcns016x016p12", domainSize: 16, rangeSize: 16, picks: 1<<12 - 1},
	{name: "randomSubsetFcns032x016p16", domainSize: 32, rangeSize: 16, picks: 1<<16 - 1},
	{name: "randomSubsetFcns032x016p08", domainSize: 32, rangeSize: 16, picks: 1<<8 - 1},
	{name: "randomSubsetFcns032x016p12", domainSize: 32, rangeSize: 16, picks: 1<<12 - 1},
	{name: "randomSubsetFcns032x032p16", domainSize: 32, rangeSize: 32, picks: 1<<16 - 1},
	{name: "randomSubsetFcns032x032p08", domainSize: 32, rangeSize: 32, picks: 1<<8 - 1},
	{name: "randomSubsetFcns032x032p12", domainSize: 32, rangeSize: 32, picks: 1<<12 - 1},
	{name: "randomSubsetFcns048x048p16", domainSize: 48, rangeSize: 48, picks: 1<<16 - 1},
	{name: "randomSubsetFcns048x048p08", domainSize: 48, rangeSize: 48, picks: 1<<8 - 1},
	{name: "randomSubsetFcns048x048p12", domainSize: 48, rangeSize: 48, picks: 1<<12 - 1},
}

var javaRandomizationBenchmarkResult *SetEnumValue

func javaRandomizationBenchmarkValues(size int) *SetEnumValue {
	values := make([]Value, size)
	for i := range values {
		values[i] = NewIntValue(int32(i + 1))
	}
	return NewSetEnumValue(values, true)
}

func BenchmarkJavaRandomization(b *testing.B) {
	oldSeed, oldRNG, oldPoly := RandomEnumerableSeed(), ResetRandomEnumerableValues(), FP64IrredPoly()
	oldSetBound := Globals.SetBound
	b.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRNG)
		FP64InitPoly(oldPoly)
		Globals.SetBound = oldSetBound
	})
	SetRandomEnumerableSeed(15041980)
	ResetRandomEnumerableValues()
	FP64Init()
	Globals.SetBound = 1000000

	subsets := make(map[int32]*SubsetValue, 5)
	for _, size := range []int32{24, 31, 65, 150, 268} {
		subsets[size] = NewSubsetValue(NewIntervalValue(1, size))
	}
	intervals := make(map[int32]*IntervalValue, 5)
	for _, high := range []int32{2 << 16, 2 << 20, 2 << 24, 2 << 28, 1<<31 - 1} {
		intervals[high] = NewIntervalValue(0, high)
	}
	sets := make(map[int]*SetEnumValue, len(javaRandomSetEnumBenchmarkCases))
	for _, benchmark := range javaRandomSetEnumBenchmarkCases {
		sets[benchmark.size] = javaRandomizationBenchmarkValues(benchmark.size)
	}
	functionSets := make(map[[2]int32]*SetOfFcnsValue, 7)
	for _, pair := range [][2]int32{{8, 8}, {11, 11}, {16, 8}, {16, 16}, {32, 16}, {32, 32}, {48, 48}} {
		functionSets[pair] = NewSetOfFcnsValue(javaRandomizationBenchmarkValues(int(pair[0])), javaRandomizationBenchmarkValues(int(pair[1])))
	}

	reset := func() {
		SetRandomEnumerableSeed(15041980)
		ResetRandomEnumerableValues()
		FP64Init()
	}
	for _, benchmark := range javaRandomSetOfSubsetsBenchmarkCases {
		benchmark := benchmark
		b.Run(benchmark.name, func(b *testing.B) {
			reset()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result *SetEnumValue
				var err error
				if benchmark.exact {
					result, err = subsets[benchmark.size].GetRandomSetOfSubsetsUpTo(benchmark.picks, 10)
				} else {
					result, err = subsets[benchmark.size].GetRandomSetOfSubsets(benchmark.picks, 0.1)
				}
				if err != nil {
					b.Fatal(err)
				}
				javaRandomizationBenchmarkResult = result
			}
		})
	}
	for _, benchmark := range javaRandomIntervalBenchmarkCases {
		benchmark := benchmark
		b.Run(benchmark.name, func(b *testing.B) {
			reset()
			interval := intervals[benchmark.high]
			size, err := interval.Size()
			if err != nil {
				b.Fatal(err)
			}
			picks := Globals.SetBound
			if benchmark.kind == 'a' {
				picks = size - 1
			} else if benchmark.kind == 'h' {
				picks = size / 2
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := interval.GetRandomSubset(picks)
				if err != nil {
					b.Fatal(err)
				}
				javaRandomizationBenchmarkResult = result
			}
		})
	}
	for _, benchmark := range javaRandomSetEnumBenchmarkCases {
		benchmark := benchmark
		b.Run(benchmark.name, func(b *testing.B) {
			reset()
			set := sets[benchmark.size]
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := set.GetRandomSubset(benchmark.size - 1)
				if err != nil {
					b.Fatal(err)
				}
				javaRandomizationBenchmarkResult = result
			}
		})
	}
	for _, benchmark := range javaRandomFunctionSetBenchmarkCases {
		benchmark := benchmark
		b.Run(benchmark.name, func(b *testing.B) {
			reset()
			set := functionSets[[2]int32{benchmark.domainSize, benchmark.rangeSize}]
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := set.GetRandomSubset(benchmark.picks)
				if err != nil {
					b.Fatal(err)
				}
				javaRandomizationBenchmarkResult = result
			}
		})
	}
}
