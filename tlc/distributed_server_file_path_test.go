package tlc

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// No original distributed Java test directly exercises TLCServer.read's
// diagnostics. getAbsolutePath retains dot components, including those after
// symlinks; the pathname used for reading and reporting must describe one file.
func TestDistributedServerReadPreservesTraversalPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture requires Unix symlink traversal")
	}
	directory := t.TempDir()
	t.Chdir(directory)
	if err := os.MkdirAll("branch/inside", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("branch/inside", "link"); err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string]string{"Spec.tla": "collapsed", "branch/Spec.tla": "traversed"} {
		if err := os.WriteFile(name, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, prefix := range []string{"link/..", directory + "/link/.."} {
		t.Run(prefix, func(t *testing.T) {
			if got := string(readDistributedServerFile(prefix + "/Spec.tla")); got != "traversed" {
				t.Fatalf("read traversed file = %q", got)
			}
			absolute := prefix
			if !filepath.IsAbs(prefix) {
				absolute = directory + "/" + prefix
			}
			for _, missing := range []bool{false, true} {
				name, want := prefix, "Unsupported operation, file "+absolute+" is a directory"
				if missing {
					name += "/missing.tla"
					want = "Exception occured on reading file " + absolute + "/missing.tla"
				}
				var failure any
				func() {
					defer func() { failure = recover() }()
					readDistributedServerFile(name)
				}()
				err, ok := failure.(error)
				if !ok {
					t.Fatalf("read %q failure = %v", name, failure)
				}
				if missing {
					// Source read wraps the contextual failure once more in finally.
					err = errors.Unwrap(err)
					var openFailure *FileNotFoundException
					if err == nil || !errors.As(errors.Unwrap(err), &openFailure) {
						t.Fatalf("read failure lost its nested file-open cause: %v", failure)
					}
				}
				if err.Error() != want {
					t.Fatalf("read %q diagnostic = %q, want %q", name, err.Error(), want)
				}
			}
		})
	}
}
