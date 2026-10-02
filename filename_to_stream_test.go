// Java source: util/SimpleFilenameToStreamTest.java.
// Copyright (c) 2023, 2025, Oracle and/or its affiliates.
package tlago

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func javaFilenameResolver(t *testing.T, libraries []string) *tlc.SimpleFilenameToStream {
	t.Helper()
	standard, err := fs.Sub(embeddedJavaStandardModules, "test_vectors/java-sany/StandardModules")
	if err != nil {
		t.Fatal(err)
	}
	// This is the Go counterpart of the Java test runner's tool/community
	// classpath. Both resource roots are frozen, self-contained repository data.
	return tlc.NewSimpleFilenameToStream(libraries, tlc.FilenameResolverOptions{Classpath: []tlc.FilenameClasspathEntry{
		{Files: standard, Prefix: tlc.StandardModulesClasspath},
		{Path: "test_vectors/java-sany/CommunityModules.jar"},
	}})
}

func checkJavaStandardResolution(t *testing.T, resolver *tlc.SimpleFilenameToStream, path string, module bool) {
	t.Helper()
	file := resolver.Resolve(path, module)
	if file == nil || !file.Exists() {
		t.Fatalf("resolve(%q, %t) = %v does not exist", path, module, file)
	}
	if !resolver.IsStandardModule(path) {
		t.Fatalf("resolve(%q, %t) should be a standard module", path, module)
	}
}

func TestJavaResolveStandardModules(t *testing.T) {
	resolver := javaFilenameResolver(t, nil)
	for _, name := range []string{"Bags", "FiniteSets", "Integers", "Json", "Naturals", "Randomization", "Reals", "RealTime", "Sequences", "TLC", "TLCExt", "Toolbox"} {
		checkJavaStandardResolution(t, resolver, name, true)
		checkJavaStandardResolution(t, resolver, name+".tla", true)
		checkJavaStandardResolution(t, resolver, name+".tla", false)
	}
}

func TestJavaResolveCommunityModules(t *testing.T) {
	resolver := javaFilenameResolver(t, nil)
	for _, name := range []string{"Bitwise", "CSV", "IOUtils", "VectorClocks"} {
		checkJavaStandardResolution(t, resolver, name, true)
		checkJavaStandardResolution(t, resolver, name+".tla", true)
		checkJavaStandardResolution(t, resolver, name+".tla", false)
	}
}

func writeJavaResolverFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
}

func javaResolverUserDirectory(t *testing.T, dir string) {
	t.Helper()
	old := tlc.GetFilenameUserDirectory()
	tlc.SetFilenameUserDirectory(&dir)
	t.Cleanup(func() { tlc.SetFilenameUserDirectory(old) })
}

func TestJavaResolveByAbsolutePath(t *testing.T) {
	module := filepath.Join(t.TempDir(), "MyModule.tla")
	writeJavaResolverFile(t, module)
	if !javaFilenameResolver(t, nil).Resolve(module, true).Exists() {
		t.Fatal("resolver should find an absolute path outside library paths")
	}
}

func TestJavaResolveFromUserDir(t *testing.T) {
	d1, d2 := t.TempDir(), t.TempDir()
	javaResolverUserDirectory(t, d1)
	if javaFilenameResolver(t, nil).Resolve("MyModule", true).Exists() {
		t.Fatal("nonexistent module resolved")
	}
	writeJavaResolverFile(t, filepath.Join(d1, "MyModule.tla"))
	if !javaFilenameResolver(t, nil).Resolve("MyModule", true).Exists() {
		t.Fatal("empty module in user directory did not resolve")
	}
	if javaFilenameResolver(t, nil).IsStandardModule("MyModule") {
		t.Fatal("user module is marked as standard")
	}
	tlc.SetFilenameUserDirectory(&d2)
	if javaFilenameResolver(t, nil).Resolve("MyModule", true).Exists() {
		t.Fatal("module resolved from another user directory")
	}
	subdir := filepath.Join(d2, "subdir")
	if err := os.Mkdir(subdir, 0700); err != nil {
		t.Fatal(err)
	}
	writeJavaResolverFile(t, filepath.Join(subdir, "MyModule.tla"))
	if javaFilenameResolver(t, nil).Resolve("MyModule", true).Exists() {
		t.Fatal("search recursed into a subdirectory")
	}
	for _, name := range []string{"subdir/MyModule", "subdir/MyModule.tla"} {
		if !javaFilenameResolver(t, nil).Resolve(name, true).Exists() {
			t.Fatalf("module %q did not resolve", name)
		}
	}
	if javaFilenameResolver(t, nil).Resolve("subdir/MyModule", false).Exists() {
		t.Fatal("non-module resolution appended .tla")
	}
	if !javaFilenameResolver(t, nil).Resolve("subdir/MyModule.tla", false).Exists() {
		t.Fatal("explicit non-module path did not resolve")
	}
}

func TestJavaResolveWithCustomLibraryPath(t *testing.T) {
	d1, d2 := t.TempDir(), t.TempDir()
	resolver := javaFilenameResolver(t, []string{d1})
	if !resolver.Resolve("Integers", true).Exists() || !resolver.IsStandardModule("Integers") {
		t.Fatal("standard module missing with explicit library path")
	}
	writeJavaResolverFile(t, filepath.Join(d1, "MyModule.tla"))
	writeJavaResolverFile(t, filepath.Join(d2, "MyModule.tla"))
	if !resolver.Resolve("MyModule", true).Exists() {
		t.Fatal("explicit library module missing")
	}
	if resolver.IsStandardModule("MyModule") {
		t.Fatal("explicit filesystem library is marked as standard")
	}
	if !resolver.IsStandardModule("Integers") {
		t.Fatal("standard module missing after adding a custom module")
	}
	javaResolverUserDirectory(t, d2)
	if !strings.HasPrefix(javaFilenameResolver(t, []string{d1}).Resolve("MyModule", true).GetPath(), d2) {
		t.Fatal("user directory should precede the explicit library path")
	}
}

func TestJavaTLALibrarySystemProperty(t *testing.T) {
	d1, d2 := t.TempDir(), t.TempDir()
	resolver := javaFilenameResolver(t, nil)
	t.Setenv(tlc.TLALibraryProperty, d1)
	if !resolver.Resolve("Integers", true).Exists() || !resolver.IsStandardModule("Integers") {
		t.Fatal("standard module missing after setting the library property")
	}
	writeJavaResolverFile(t, filepath.Join(d1, "MyModule.tla"))
	resolver = javaFilenameResolver(t, nil)
	if !resolver.Resolve("MyModule", true).Exists() {
		t.Fatal("module in TLA-Library missing")
	}
	if resolver.IsStandardModule("MyModule") {
		t.Fatal("TLA-Library filesystem module marked as standard")
	}
	if javaFilenameResolver(t, []string{d2}).Resolve("MyModule", true).Exists() {
		t.Fatal("explicit library path did not replace TLA-Library")
	}
}

func TestJavaWindowsTLAFileCreation(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	} // The Java test also guards this case.
	parent := `X:\Develop\myspecs\DecentSpec\`
	file := tlc.NewTLAFileInDirectory(parent, parent+"Fromage.tla", nil)
	if count := strings.Count(file.GetAbsolutePath(), "X:"); count != 1 {
		t.Fatalf("drive letter count = %d, want 1", count)
	}
}

func TestJavaBizarreWorkingDirectorySearchBehavior(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("user.dir", dir)
	writeJavaResolverFile(t, filepath.Join(dir, "Test.tla"))
	writeJavaResolverFile(t, filepath.Join(dir, "Test.cfg"))
	subdir := filepath.Join(dir, "test")
	if err := os.Mkdir(subdir, 0700); err != nil {
		t.Fatal(err)
	}
	writeJavaResolverFile(t, filepath.Join(subdir, "Test.tla"))
	writeJavaResolverFile(t, filepath.Join(subdir, "Test.cfg"))
	if got, want := javaFilenameResolver(t, nil).Resolve(filepath.Join("test", "Test.tla"), true).GetAbsolutePath(), filepath.Join(subdir, "Test.tla"); got != want {
		t.Fatalf("named subdirectory resolved to %q, want %q", got, want)
	}
}
