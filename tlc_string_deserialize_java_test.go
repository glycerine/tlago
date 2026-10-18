package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// StringDeserializeTLCTest.test, including its ModelCheckerTestCase setup and
// successful exit assertion. The original .vos bytes are copied without using
// the Go Value API, so reading cannot benefit from pre-interning the data.
func TestJavaStringDeserialize(t *testing.T) {
	path, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "StringDeserialize", "StringDeserialize.vos"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("StringDeserializeTLCTest.vosFileName", path)
	// The original Java classpath supplies IOUtils from CommunityModules.
	modules, err := filepath.Abs(filepath.Join("test_vectors", "java-sany", "CommunityModules.jar"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLASSPATH", modules)
	result := runJavaTLCModelTest(t, "StringDeserialize")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status %d, want success; messages: %+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}
