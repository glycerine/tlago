package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestJavaJson(t *testing.T) {
	const childVariable = "TLAGO_JAVA_JSON_TEST_CHILD"
	if os.Getenv(childVariable) != "true" {
		fixture, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "JsonTests"))
		if err != nil {
			t.Fatal(err)
		}
		directory := t.TempDir()
		if err := os.CopyFS(filepath.Join(directory, "tlc", "test_vectors", "models", "JsonTests"), os.DirFS(fixture)); err != nil {
			t.Fatal(err)
		}
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		// Source test classloader isolation and a private cwd for target/json output.
		command := exec.Command(executable, "-test.run=^TestJavaJson$", "-test.v")
		command.Dir = directory
		command.Env = append(os.Environ(), childVariable+"=true")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("original JsonTest: %v\n%s", err, output)
		}
		return
	}
	result := runJavaTLCModelTest(t, "JsonTests", "-config", "JsonTests.tla")
	// ModelCheckerTestCase.tearDown checks its default successful exit too.
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "0")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL unexpectedly recorded: %v", records)
	}
}
