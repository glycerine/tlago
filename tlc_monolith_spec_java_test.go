package tlago

import (
	"bytes"
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Original MonolithSpecTest.testSpec and beforeSetUp's actual ToolIO streams.
func TestJavaMonolithSpec(t *testing.T) {
	var stdout, stderr bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&stdout, &stderr)
	defer restore()
	tlc.ToolIOReset()
	r := runJavaTLCModelTest(t, "MonolithSpec", "-config", "MonolithSpec.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "MonolithSpecTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "214", "54", "0")
	if strings.Contains(stderr.String(), "File does not exist") {
		t.Fatalf("unexpected missing-file warning: %s", stderr.String())
	}
	// Source test-model path denotes the model-fixture directory; retain its
	// exact suffix using this port's test_vectors/models/MonolithSpec directory.
	sep := regexp.QuoteMeta(string(filepath.Separator))
	for _, name := range []string{"EWD840", "Mod4711", "Mod4712"} {
		pattern := "Parsing file .*" + sep + name + `.tla \(.*` + sep + "models" + sep + "MonolithSpec" + sep + `MonolithSpec.tla\)`
		if !regexp.MustCompile(pattern).MatchString(stdout.String()) {
			t.Fatalf("parsing provenance missing for %s: %s", name, stdout.String())
		}
	}
	requireJavaTLCUncovered(t, r)
}
