//go:build !linux

package tlc

import (
	"os"
	"strings"
)

// Other default providers retain the existing implementation until their own
// source path, exception and delete rules have been ported.
func ioUtilsTXTPath(input string) (string, error) {
	if strings.IndexByte(input, 0) >= 0 {
		return "", NewInvalidPathException(input, "Nul character not allowed")
	}
	return input, nil
}

func ioUtilsTXTSystemPath(path string) string       { return path }
func ioUtilsTXTPathError(_ string, err error) error { return err }
func ioUtilsTXTChannelError(err error) error        { return err }

func ioUtilsTXTOpen(path string, options ioUtilsFileOptions) (*os.File, error) {
	return os.OpenFile(path, options.flag, 0o666)
}

func ioUtilsTXTDeleteAfterClose(path string, options ioUtilsFileOptions) {
	if options.deleteOnClose {
		_ = os.Remove(path)
	}
}
