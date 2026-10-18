//go:build tlago_disabled_distributed_tests

package tlago

import "testing"

// These diagnostics deliberately bypass Java's unconditional setup assumption.
// They retain every original assertion and source Ant configuration. Passing a
// diagnostic is not an enabled original-test pass. The final off-heap flusher
// lifecycle bug is deliberately fixed in Go; see JAVA_BUG_FOUND.md.
func TestDiagnosticJavaDieHardDistributed(t *testing.T) {
	checkJavaDieHardDistributed(t)
}

func TestDiagnosticJavaEWD840Distributed(t *testing.T) {
	checkJavaEWD840Distributed(t)
}

func TestDiagnosticJavaEWD840DistributedWithFPSet(t *testing.T) {
	checkJavaEWD840DistributedWithFPSet(t)
}

func TestDiagnosticJavaTSnapShotDistributed(t *testing.T) {
	checkJavaTSnapShotDistributed(t)
}
