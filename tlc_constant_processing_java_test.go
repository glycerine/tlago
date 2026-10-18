package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaGithub361(t *testing.T) {
	// The Java override uses two workers so single-worker optimizations cannot
	// hide expensive initialization of the Partitions constant.
	result := runJavaTLCModelTestWithWorkers(t, "Github361", true, true, 2,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github361.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitGenerated1, "1")
}
