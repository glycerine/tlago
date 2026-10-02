package tlc

import (
	"math"
	"testing"
)

// Port of tlc2/tool/distributed/TLCWorkerSmartProxyTest.java. Java's
// DummyTLCWorker returns a NextStateResult with the requested duration;
// measureNextStates is the Go proxy's decorator around that endpoint call.
func TestJavaTLCWorkerSmartProxyNetworkOverhead(t *testing.T) {
	// Preserve Java's integer division: MAX_ARRAY_SIZE is actually zero.
	const maxArraySize = math.MaxInt32 * (1 / 10)
	tests := []struct {
		name     string
		duration int64
		count    int
	}{
		{"MaxStateOne", math.MaxInt64, 1},
		{"MinStateOne", math.MinInt64, 1},
		{"ZeroStateOne", 0, 1},
		{"MaxStateZero", math.MaxInt64, 0},
		{"MinStateZero", math.MinInt64, 0},
		{"ZeroStateZero", 0, 0},
		{"MinStateMax", math.MinInt64, maxArraySize},
		{"MaxStateMa", math.MaxInt64, maxArraySize},
		{"ZeroStateMax", 0, maxArraySize},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proxy := NewDistributedWorkerSmartProxy(nil)
			result, err := proxy.measureNextStates(make([]*TLCStateMut, test.count), func() (*NextStateResult, error) {
				return NewNextStateResult(nil, nil, test.duration, -1), nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if result == nil {
				t.Fatal("getNextStates returned null")
			}
			if overhead := proxy.GetNetworkOverhead(); !(overhead > 0) {
				t.Fatalf("network overhead = %v, want > 0", overhead)
			}
		})
	}
}
