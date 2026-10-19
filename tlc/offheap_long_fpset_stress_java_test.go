//go:build tlc_fp_stress

package tlc

import "testing"

// Complete FPSetTest.testMaxFPSetSizeRnd inherited by OffHeapDiskFPSetLongTest.
// The original factory ignores its argument and uses ratio 1.0; the shared
// method preserves all 2,147,483,648 iterations and every original assertion.
// Run normally with -timeout=0, never with -race.
func TestJavaOffHeapDiskFPSetLong_testMaxFPSetSizeRnd(t *testing.T) {
	skipSlowTest(t)

	javaLongFPSetRandomFull(t, "tlc2.tool.fp.OffHeapDiskFPSet", func(ignored *FPSetConfiguration) FPSet {
		return NewOffHeapDiskFPSet(NewFPSetConfigurationWithRatioAndImplementation(1.0, "tlc2.tool.fp.OffHeapDiskFPSet"))
	})
}
