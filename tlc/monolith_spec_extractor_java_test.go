package tlc

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const javaMonolithSpec = "D:\\software\\TLA+\\Specs\\Scratch\\Scratch.tla\n\n\n" +
	"---- MODULE Scratch ----\nEXTENDS TLC\nSpec == TRUE /\\ [][TRUE]_TRUE\n======\n\n" +
	"---- CONFIG Scratch ----\nSPECIFICATION Spec\n=====\n"

// Whole original util.MonolithSpecExtractorTest, all five methods.
func TestJavaMonolithSpecExtractor(t *testing.T) {
	for _, input := range []struct{ name, config, want string }{
		{"testExtractConfig", "Scratch", "SPECIFICATION Spec"},
		{"testConfigWithWindowsPathAsName", `d:\software\TLA+\Specs\Scratch\Scratch`, ""},
	} {
		t.Run(input.name, func(t *testing.T) {
			// ByteArrayInputStream/UTF-8 fixture lowers to the same source characters.
			stream := strings.NewReader(javaMonolithSpec)
			data, err := io.ReadAll(stream)
			if err != nil {
				t.Fatal(err)
			}
			config := ExtractMonolithConfigSource(string(data), input.config)
			got := strings.Join(monolithLines(config), "\n")
			if got != input.want {
				t.Fatalf("config=%q, want %q", got, input.want)
			}
		})
	}
	for _, input := range []struct {
		name, module string
		found        bool
	}{
		{"testExtractModule", "Scratch", true},
		{"testModuleWithWindowsPathAsName", `d:\software\TLA+\Specs\Scratch\Scratch`, false},
	} {
		t.Run(input.name, func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "MonolithSpecExtractorTest*.tla")
			if err != nil {
				t.Fatal(err)
			}
			path := file.Name()
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			defer os.Remove(path)
			if err := os.WriteFile(path, []byte(javaMonolithSpec), 0600); err != nil {
				t.Fatal(err)
			}
			stream, err := MonolithModule(path, input.module)
			if err != nil {
				t.Fatal(err)
			}
			if !input.found {
				if stream != nil {
					t.Fatal("expected null module")
				}
				return
			}
			if stream == nil {
				t.Fatal("module() should find the MODULE Scratch section")
			}
			// Keep the source delete-on-exit artifact until this test is finished.
			t.Cleanup(func() {
				if !stream.closed {
					_ = stream.Close()
				}
				_ = os.RemoveAll(filepath.Dir(stream.SourceFile()))
			})
			if got := stream.GetModuleName(); got != "Scratch" {
				t.Fatalf("module name=%q", got)
			}
			data, err := io.ReadAll(stream)
			if err != nil {
				t.Fatal(err)
			}
			module := strings.Join(monolithLines(string(data)), "\n")
			want := "---- MODULE Scratch ----\nEXTENDS TLC\nSpec == TRUE /\\ [][TRUE]_TRUE\n======"
			if module != want {
				t.Fatalf("module=%q, want %q", module, want)
			}
			if err := stream.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("testGetConfig", func(t *testing.T) {
		if got := MonolithGetConfig("Scratch.tla"); got != "Scratch.tla" {
			t.Fatalf("config=%q", got)
		}
		if got := MonolithGetConfig("Scratch"); got != "Scratch.cfg" {
			t.Fatalf("config=%q", got)
		}
	})
}
