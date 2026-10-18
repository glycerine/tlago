package tlago

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original ElevatedSanyWarning.testSpec and beforeSetUp, with its absolute
// corpus path and inherited ModelCheckerTestCase settings and exit assertion.
func TestJavaElevatedSanyWarning(t *testing.T) {
	specPath, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "ElevatedSanyWarning", "W4802_Pre_Test.tla"))
	if err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&stdout, os.Stderr)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "ElevatedSanyWarning", specPath, true, true, true, 1,
		"-messagesAsErrors", "4802", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ElevatedSanyWarning.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCParsingFailed)) == 0 {
		t.Fatal("TLC_PARSING_FAILED absent")
	}
	if !strings.Contains(stdout.String(), "Warning treated as error") {
		t.Fatalf("promoted warning absent: %s", stdout.String())
	}
}
