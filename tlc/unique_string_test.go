package tlc

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestUniqueStringCommitCheckpointErrorUsesJavaText(t *testing.T) {
	dir := t.TempDir()
	err := CommitChkptUniqueStrings(dir)
	if err == nil {
		t.Fatalf("CommitChkptUniqueStrings without tmp file returned nil")
	}
	want := "InternTable.commitChkpt: cannot delete " + filepath.Join(dir, "vars.chkpt")
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("CommitChkptUniqueStrings error = %q, want containing %q", err.Error(), want)
	}
}
