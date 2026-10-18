package tlc

import "testing"

type javaSubsetValueBenchmarkCase struct {
	name        string
	size        int32
	picks       int
	probability float64
	maxLength   int
	exact       bool
}

var javaSubsetValueBenchmarkCases = []javaSubsetValueBenchmarkCase{
	{name: "probabilisticN035k80d01", size: 35, picks: 80000, probability: 0.1},
	{name: "probabilisticN060k80d01", size: 60, picks: 80000, probability: 0.1},
	{name: "probabilisticN100k80d01", size: 100, picks: 80000, probability: 0.1},
	{name: "probabilisticN200k80d01", size: 200, picks: 80000, probability: 0.1},
	{name: "probabilisticN300k80d01", size: 300, picks: 80000, probability: 0.1},
	{name: "probabilisticN400k80d01", size: 400, picks: 80000, probability: 0.1},
	{name: "probabilisticN035k16d01", size: 35, picks: 160000, probability: 0.1},
	{name: "probabilisticN060k16d01", size: 60, picks: 160000, probability: 0.1},
	{name: "probabilisticN100k16d01", size: 100, picks: 160000, probability: 0.1},
	{name: "probabilisticN200k16d01", size: 200, picks: 160000, probability: 0.1},
	{name: "probabilisticN300k16d01", size: 300, picks: 160000, probability: 0.1},
	{name: "probabilisticN400k16d01", size: 400, picks: 160000, probability: 0.1},
	{name: "probabilisticN035k80d02", size: 35, picks: 80000, probability: 0.2},
	{name: "probabilisticN060k80d02", size: 60, picks: 80000, probability: 0.2},
	{name: "probabilisticN100k80d02", size: 100, picks: 80000, probability: 0.2},
	{name: "probabilisticN200k80d02", size: 200, picks: 80000, probability: 0.2},
	{name: "probabilisticN300k80d02", size: 300, picks: 80000, probability: 0.2},
	{name: "probabilisticN400k80d02", size: 400, picks: 80000, probability: 0.2},
	{name: "probabilisticN035k16d02", size: 35, picks: 160000, probability: 0.2},
	{name: "probabilisticN060k16d02", size: 60, picks: 160000, probability: 0.2},
	{name: "probabilisticN100k16d02", size: 100, picks: 160000, probability: 0.2},
	{name: "probabilisticN200k16d02", size: 200, picks: 160000, probability: 0.2},
	{name: "probabilisticN300k16d02", size: 300, picks: 160000, probability: 0.2},
	{name: "probabilisticN400k16d02", size: 400, picks: 160000, probability: 0.2},
	{name: "exactN035K08", size: 35, picks: 80000, maxLength: 8, exact: true},
	{name: "exactN035K13", size: 35, picks: 80000, maxLength: 13, exact: true},
	{name: "exactN060K08", size: 60, picks: 80000, maxLength: 8, exact: true},
	{name: "exactN100K08", size: 100, picks: 80000, maxLength: 8, exact: true},
	{name: "exactN100K10", size: 100, picks: 80000, maxLength: 10, exact: true},
	{name: "exactN200K10", size: 200, picks: 80000, maxLength: 10, exact: true},
	{name: "exactN300K09", size: 300, picks: 80000, maxLength: 9, exact: true},
	{name: "exactN400K09", size: 400, picks: 80000, maxLength: 9, exact: true},
	{name: "exactN035K208", size: 35, picks: 160000, maxLength: 8, exact: true},
	{name: "exactN035K213", size: 35, picks: 160000, maxLength: 13, exact: true},
	{name: "exactN060K208", size: 60, picks: 160000, maxLength: 8, exact: true},
	{name: "exactN100K208", size: 100, picks: 160000, maxLength: 8, exact: true},
	{name: "exactN100K210", size: 100, picks: 160000, maxLength: 10, exact: true},
	{name: "exactN200K210", size: 200, picks: 160000, maxLength: 10, exact: true},
	{name: "exactN300K209", size: 300, picks: 160000, maxLength: 9, exact: true},
	{name: "exactN400K209", size: 400, picks: 160000, maxLength: 9, exact: true},
}

var javaSubsetValueBenchmarkResult *SetEnumValue

func BenchmarkJavaSubsetValue(b *testing.B) {
	oldSeed, oldRNG, oldPoly := RandomEnumerableSeed(), ResetRandomEnumerableValues(), FP64IrredPoly()
	b.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRNG)
		FP64InitPoly(oldPoly)
	})
	SetRandomEnumerableSeed(15041980)
	ResetRandomEnumerableValues()
	FP64Init()
	subsets := make(map[int32]*SubsetValue, 6)
	for _, size := range []int32{35, 60, 100, 200, 300, 400} {
		subsets[size] = NewSubsetValue(NewIntervalValue(1, size)).Normalize().(*SubsetValue)
	}

	for _, benchmark := range javaSubsetValueBenchmarkCases {
		benchmark := benchmark
		b.Run(benchmark.name, func(b *testing.B) {
			SetRandomEnumerableSeed(15041980)
			ResetRandomEnumerableValues()
			FP64Init()
			subset := subsets[benchmark.size]
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result *SetEnumValue
				var err error
				if benchmark.exact {
					result, err = subset.GetRandomSetOfSubsetsUpTo(benchmark.picks, benchmark.maxLength)
				} else {
					result, err = subset.GetRandomSetOfSubsets(benchmark.picks, benchmark.probability)
				}
				if err != nil {
					b.Fatal(err)
				}
				javaSubsetValueBenchmarkResult = result
			}
		})
	}
}
