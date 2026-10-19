package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Linux descriptor paths (including strace -yy output) are absolute and have
// symlinks resolved. Keep the original directory for application I/O, but use
// this spelling for native observations. Resolve the existing parent before
// appending a checkpoint filename: strace -P cannot canonicalize a file that
// the child has not created yet.
func canonicalNativeTestDirectory(t *testing.T, directory string) string {
	t.Helper()
	if !filepath.IsAbs(directory) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("get working directory for native test path %q: %v", directory, err)
		}
		// Include any symlinks in the working directory and resolve them
		// before reducing ".." in the supplied path.
		directory = cwd + string(os.PathSeparator) + directory
	}
	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatalf("resolve native test directory %q: %v", directory, err)
	}
	return resolved
}

// Call only on Linux, while directory still exists. Both the positive startup
// checks and the negative cleanup checks must compare the same path spelling.
func nativeTestDescriptorsUnder(t *testing.T, directory string) []string {
	t.Helper()
	directory = canonicalNativeTestDirectory(t, directory)
	prefix := strings.TrimSuffix(directory, string(os.PathSeparator)) + string(os.PathSeparator)
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, entry := range entries {
		path, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err == nil && strings.HasPrefix(path, prefix) {
			paths = append(paths, path)
		}
	}
	return paths
}
