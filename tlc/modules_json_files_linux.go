//go:build linux

// java.io.File/UnixFileSystem and FileOutputStream's Linux boundaries.
// OpenJDK's GPLv2 with Classpath Exception notices are in tlc/licenses/.
package tlc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func jsonOpenFileWriter(input string) (*os.File, error) {
	path := javaUnixNormalizePath(input)
	if strings.IndexByte(path, 0) >= 0 {
		return nil, NewFileNotFoundException("Invalid file path")
	}
	// java.io's native pathname conversion replaces malformed UTF-16, unlike
	// the strict UnixPath encoding used by java.nio.file.Files.
	encoded, err := javaCharsetEncode(path, "UTF-8", false)
	if err != nil {
		return nil, err
	}
	native := string(encoded)
	if parent := javaIOFileParent(native); parent != "" {
		_ = javaIOMkdirs(parent) // File.mkdirs' boolean is ignored by Json.
	}
	file, err := os.OpenFile(native, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return nil, NewFileNotFoundException(path + " (" + ioUtilsTXTNativeError(err) + ")")
	}
	return file, nil
}

func javaUnixNormalizePath(input string) string {
	var path strings.Builder
	lastSlash := false
	for i := 0; i < len(input); i++ {
		slash := input[i] == '/'
		if !slash || !lastSlash {
			path.WriteByte(input[i])
		}
		lastSlash = slash
	}
	normalized := path.String()
	if len(normalized) > 1 && lastSlash {
		normalized = normalized[:len(normalized)-1]
	}
	return normalized
}

func javaIOFileParent(path string) string {
	index := strings.LastIndexByte(path, '/')
	if index < 0 || path == "/" {
		return ""
	}
	if index == 0 {
		return "/"
	}
	return path[:index]
}

func javaIOMkdirs(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return false
	}
	if err := os.Mkdir(path, 0o777); err == nil {
		return true
	}
	canonical, err := javaIOCanonicalPath(path)
	if err != nil {
		return false
	}
	parent := javaIOFileParent(canonical)
	if parent == "" {
		return false
	}
	if !javaIOMkdirs(parent) {
		if _, err := os.Stat(parent); err != nil {
			return false
		}
	}
	return os.Mkdir(canonical, 0o777) == nil
}

// Unix canonicalization resolves an existing prefix and removes lexical dot
// components from the unresolved suffix. This differs from creating every raw
// parent with MkdirAll, e.g. a missing component followed by "..".
func javaIOCanonicalPath(path string) (string, error) {
	if !strings.HasPrefix(path, "/") {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = cwd + "/" + path
	}
	prefix := path
	for {
		// realpath follows the native kernel's symlink limit. Go's lexical
		// resolver permits more links, so first retain the native ELOOP result.
		var resolved string
		_, err := os.Stat(prefix)
		if !errors.Is(err, syscall.ELOOP) {
			resolved, err = filepath.EvalSymlinks(prefix)
		}
		if err == nil {
			return filepath.Join(resolved, path[len(prefix):]), nil
		}
		// The final component can remain unresolved. A symlink loop or other
		// hard native failure in a parent cannot be discarded by peeling past it.
		if prefix != path && !errors.Is(err, syscall.ENOENT) && !errors.Is(err, syscall.ENOTDIR) && !errors.Is(err, syscall.EACCES) {
			if strings.Contains(err.Error(), "too many links") {
				return "", NewIOException(ioUtilsTXTNativeMessage(syscall.ELOOP))
			}
			return "", ioUtilsTXTChannelError(err)
		}
		parent := javaIOFileParent(prefix)
		if parent == "" || parent == prefix {
			return "", err
		}
		prefix = parent
	}
}
