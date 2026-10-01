package tlc

import "testing"

func TestFPIntSetStatusLevelUsesJavaUnsignedShift(t *testing.T) {
	fpIntSetMu.Lock()
	oldLevel := fpIntSetLevel
	fpIntSetLevel = 1<<30 - 1
	fpIntSetMu.Unlock()
	t.Cleanup(func() {
		fpIntSetMu.Lock()
		fpIntSetLevel = oldLevel
		fpIntSetMu.Unlock()
	})

	statusBits := ^uint32(3)
	status := int32(statusBits)
	if got := FPIntSetLevelOf(status); got != 1<<30-1 {
		t.Fatalf("LevelOf(%d) = %d, want Java unsigned-shift level %d", status, got, 1<<30-1)
	}
	if !FPIntSetIsLeaf(status) {
		t.Fatalf("IsLeaf(%d) = false, want true at Java unsigned-shift level", status)
	}
}
