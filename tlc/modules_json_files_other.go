//go:build !linux

package tlc

import (
	"os"
	"path/filepath"
	"strings"
)

// Other native File providers remain a separate source port; retain the existing
// host file path here while sharing the ordinary writer's codec/control flow.
func jsonOpenFileWriter(path string) (*os.File, error) {
	if strings.IndexByte(path, 0) >= 0 {
		return nil, NewFileNotFoundException("Invalid file path")
	}
	if parent := filepath.Dir(path); parent != "." && parent != "" {
		_ = os.MkdirAll(parent, 0o777)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return nil, NewFileNotFoundException(err.Error())
	}
	return file, nil
}
