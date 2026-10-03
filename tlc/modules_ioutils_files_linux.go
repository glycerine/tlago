//go:build linux

// Linux default provider UnixPath and UnixChannelFactory boundaries.
// OpenJDK's GPLv2 with Classpath Exception notices are in tlc/licenses/.
package tlc

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func ioUtilsTXTPath(input string) (string, error) {
	if strings.IndexByte(input, 0) >= 0 {
		return "", NewInvalidPathException(input, "Nul character not allowed")
	}
	// Keep dot/parent components: filesystem resolution can traverse symlinks.
	normalized := javaUnixNormalizePath(input)
	// Linux native filename UTF-8 is represented here. Native charset discovery
	// for other process locales remains a separate source feature.
	encoded, err := javaCharsetEncode(normalized, "UTF-8", true)
	if err != nil {
		return "", NewInvalidPathException(normalized, "Malformed input or input contains unmappable characters")
	}
	return string(encoded), nil
}

func ioUtilsTXTSystemPath(path string) string {
	if path == "" {
		return "."
	}
	return path
}

func ioUtilsTXTDeleteAfterClose(string, ioUtilsFileOptions) {}

func ioUtilsTXTOpen(path string, options ioUtilsFileOptions) (*os.File, error) {
	createNew := options.flag&os.O_EXCL != 0
	// Preserve UnixChannelFactory's unchecked final-byte access too.
	if createNew && path == "" {
		return nil, NewArrayIndexOutOfBoundsException(-1, 0)
	}
	// UnixChannelFactory reports EEXIST for a final dot without opening it.
	if createNew && (path == "." || strings.HasSuffix(path, "/.")) {
		return nil, NewFileAlreadyExistsException(javaString(path))
	}
	flag := options.flag
	noFollow := options.deleteOnClose && !createNew
	if noFollow {
		flag |= syscall.O_NOFOLLOW
	}
	file, err := os.OpenFile(ioUtilsTXTSystemPath(path), flag, 0o666)
	if err != nil {
		if createNew && errors.Is(err, syscall.EISDIR) {
			return nil, NewFileAlreadyExistsException(javaString(path))
		}
		if noFollow && errors.Is(err, syscall.ELOOP) {
			return nil, NewIOException(ioUtilsTXTNativeMessage(syscall.ELOOP) + " (NOFOLLOW_LINKS specified)")
		}
		return nil, ioUtilsTXTPathError(path, err)
	}
	if options.deleteOnClose {
		// Unix provider attempts unlink immediately and ignores its error.
		_ = syscall.Unlink(ioUtilsTXTSystemPath(path))
	}
	return file, nil
}

func ioUtilsTXTPathError(path string, err error) error {
	switch {
	case errors.Is(err, syscall.EACCES):
		return NewAccessDeniedException(javaString(path))
	case errors.Is(err, syscall.ENOENT):
		return NewNoSuchFileException(javaString(path))
	case errors.Is(err, syscall.EEXIST):
		return NewFileAlreadyExistsException(javaString(path))
	}
	reason := ioUtilsTXTNativeError(err)
	if errors.Is(err, syscall.ELOOP) {
		reason += " or unable to access attributes of symbolic link"
	}
	return NewFileSystemExceptionWithReason(javaString(path), nil, javaString(reason))
}

func ioUtilsTXTChannelError(err error) error {
	return NewIOException(ioUtilsTXTNativeError(err))
}

func ioUtilsTXTNativeError(err error) string {
	var number syscall.Errno
	if errors.As(err, &number) {
		return ioUtilsTXTNativeMessage(number)
	}
	return err.Error()
}

// Linux strerror in the C locale used by the native Unix provider. Go's
// syscall table differs in capitalization and unassigned/new errno slots.
func ioUtilsTXTNativeMessage(number syscall.Errno) string {
	hardwarePoison := syscall.Errno(133)
	if strings.HasPrefix(runtime.GOARCH, "mips") {
		hardwarePoison = 168
	}
	switch number {
	case 0:
		return "Success"
	case hardwarePoison:
		return "Memory page has hardware error"
	}
	message := number.Error()
	if strings.HasPrefix(message, "errno ") {
		return "Unknown error " + strconv.FormatUint(uint64(number), 10)
	}
	if message != "" && message[0] >= 'a' && message[0] <= 'z' {
		message = string(message[0]-'a'+'A') + message[1:]
	}
	return message
}
