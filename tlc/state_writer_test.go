package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIDotStateWriterUsesJavaStrictPrefixWithoutStrictFiltering(t *testing.T) {
	dotFile := filepath.Join(t.TempDir(), "states.dot")
	_, dump, err := parseTLCDumpOption([]string{"-dump", "dot", dotFile}, 0)
	if err != nil {
		t.Fatalf("parseTLCDumpOption returned error: %v", err)
	}
	writer, err := dump.newStateWriter("")
	if err != nil {
		t.Fatalf("newStateWriter returned error: %v", err)
	}
	if writer.strict != nil {
		t.Fatalf("plain CLI -dump dot enabled duplicate-edge filtering; Java only does that for the strict sub-option")
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	data, err := os.ReadFile(dotFile)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.HasPrefix(string(data), "strict digraph DiskGraph {\n") {
		t.Fatalf("DOT header = %q, want Java CLI strict prefix", firstLine(string(data)))
	}
}

func TestDotWriterDerivedNamesUseJavaReplaceAll(t *testing.T) {
	name := filepath.Join("dump.dot.dir", "states.dot")
	if got, want := dotWriterJavaReplace(name, "_snapshot.dot"), filepath.Join("dump_snapshot.dot.dir", "states_snapshot.dot"); got != want {
		t.Fatalf("snapshot name = %q, want Java String.replace result %q", got, want)
	}
	if got, want := dotWriterJavaReplace(name, "_liveness.dot"), filepath.Join("dump_liveness.dot.dir", "states_liveness.dot"); got != want {
		t.Fatalf("liveness name = %q, want Java String.replace result %q", got, want)
	}
}

func firstLine(text string) string {
	if idx := strings.IndexByte(text, '\n'); idx >= 0 {
		return text[:idx+1]
	}
	return text
}
