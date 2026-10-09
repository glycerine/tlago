//go:build !linux

package tlc

import (
	"os"
	"path/filepath"
)

// Resolve existing native ancestors before reducing missing path components.
func statePoolCanonicalPath(path string) (string, error) {
	absolute, err := metadataAbsolutePath(path)
	if err != nil {
		return "", err
	}
	prefix := absolute
	var suffix []string
	for {
		resolved, err := filepath.EvalSymlinks(prefix)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := metadataParentPath(prefix)
		if parent == prefix || parent == "" {
			return "", err
		}
		suffix = append(suffix, distributedJavaFileName(prefix))
		prefix = parent
	}
}
