// Java source: tlc2/util/FP64Test.java (Microsoft Research, MIT license).
package tlc

import "testing"

// Port of tlc2.util.FP64Test.testExtendLongInt, after both Java extension
// implementations are present. Java's test also uses 1,000 random long/int pairs.
func TestJavaFP64ExtendLongInt(t *testing.T) {
	FP64Init()
	random := NewJavaRandomDefault()
	for i := 0; i < 1000; i++ {
		fp, x := uint64(random.NextLong()), random.NextInt()
		if got, want := FP64ExtendInt(fp, x), FP64ExtendLoop(fp, x); got != want {
			t.Fatalf("Extend(%016x, %d) = %016x, loop = %016x", fp, x, got, want)
		}
	}
}
