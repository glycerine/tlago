package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"os"
	"path/filepath"
	"testing"
)

// Original AbsoluteSpecPathTest.test: the only CLI argument is the absolute
// source Test2 path. The workspace provides BASE_DIR's nonempty absolute root.
func TestJavaAbsoluteSpecPath(t *testing.T) {
	// Run the unchanged fixtures from an isolated absolute BASE_PATH so the
	// source default states directory is removed with this test's workspace.
	basePath := t.TempDir()
	for _, filename := range []string{"Test2.tla", "Test2.cfg"} {
		bytes, err := os.ReadFile(filepath.Join("tlc", "test_vectors", "models", "Test2", filename))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(basePath, filename), bytes, 0600); err != nil {
			t.Fatal(err)
		}
	}
	spec := filepath.Join(basePath, "Test2")
	r := runJavaTLCModelTestWithArguments(t, "Test2", spec, func(_, _ string) []string { return nil })
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "5")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "6", "5", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}
