//go:build linux

package tlc

import (
	"errors"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
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

func fileMoveCheckDirectoryAccess(src string) error {
	if err := syscall.Access(ioUtilsTXTSystemPath(src), 2); err != nil {
		return ioUtilsTXTPathError(src, err)
	}
	return nil
}
func fileMoveCopySpecial(dst string, source os.FileInfo) error {
	stat := source.Sys().(*syscall.Stat_t)
	if err := syscall.Mknod(dst, stat.Mode, int(stat.Rdev)); err != nil {
		return ioUtilsTXTPathError(dst, err)
	}
	return nil
}
func fileMoveCopyAttributes(src, dst string, source os.FileInfo) error {
	srcPath, dstPath := ioUtilsTXTSystemPath(src), ioUtilsTXTSystemPath(dst)
	stat := source.Sys().(*syscall.Stat_t)
	if source.Mode()&os.ModeSymlink != 0 {
		_ = os.Lchown(dstPath, int(stat.Uid), int(stat.Gid))
		return nil
	}
	// Failure to copy POSIX or extended attributes is nonfatal for Files.move;
	// the basic timestamp copy remains a checked failure.
	if err := os.Chown(dstPath, int(stat.Uid), int(stat.Gid)); err == nil {
		_ = os.Chmod(dstPath, source.Mode())
	}
	count, err := syscall.Listxattr(srcPath, nil)
	if err == nil && count > 0 {
		buffer := make([]byte, count)
		if n, err := syscall.Listxattr(srcPath, buffer); err == nil {
			for _, name := range strings.Split(string(buffer[:n]), "\x00") {
				if name == "" {
					continue
				}
				count, err := syscall.Getxattr(srcPath, name, nil)
				if err != nil {
					continue
				}
				value := make([]byte, count)
				if n, err := syscall.Getxattr(srcPath, name, value); err == nil {
					_ = syscall.Setxattr(dstPath, name, value[:n], 0)
				}
			}
		}
	}
	// The source Unix provider passes microseconds to futimes/utimes.
	atime := time.Unix(stat.Atim.Sec, stat.Atim.Nsec).Truncate(time.Microsecond)
	if err := os.Chtimes(dstPath, atime, source.ModTime().Truncate(time.Microsecond)); err != nil {
		return ioUtilsTXTPathError(dst, err)
	}
	return nil
}

// Linux ordinary replacement move, matching the source Unix provider.
func fileUtilReplaceFile(src string, dst string) error {
	var err error
	src, err = ioUtilsTXTPath(filenameNormalizeFile(src))
	if err != nil {
		return err
	}
	dst, err = ioUtilsTXTPath(filenameNormalizeFile(dst))
	if err != nil {
		return err
	}
	srcPath, dstPath := ioUtilsTXTSystemPath(src), ioUtilsTXTSystemPath(dst)
	source, err := os.Lstat(srcPath)
	if err != nil {
		return ioUtilsTXTPathError(src, err)
	}
	if source.IsDir() {
		if err := fileMoveCheckDirectoryAccess(src); err != nil {
			return err
		}
	}
	if target, err := os.Lstat(dstPath); err == nil {
		if os.SameFile(source, target) {
			return nil
		}
		if err := os.Remove(dstPath); err != nil {
			if target.IsDir() && (errors.Is(err, syscall.ENOTEMPTY) || os.IsExist(err)) {
				return NewDirectoryNotEmptyException(javaString(dst))
			}
			return ioUtilsTXTPathError(dst, err)
		}
	}
	if err := os.Rename(srcPath, dstPath); err != nil {
		if errors.Is(err, syscall.EXDEV) || errors.Is(err, syscall.EISDIR) {
			return fileMoveAcross(src, dst, source)
		}
		failure := ioUtilsTXTPathError(src, err)
		switch failure := failure.(type) {
		case *NoSuchFileException:
			return NewNoSuchFileExceptionWithReason(javaString(src), javaString(dst), failure.GetReason())
		case *AccessDeniedException:
			return NewAccessDeniedExceptionWithReason(javaString(src), javaString(dst), failure.GetReason())
		case *FileAlreadyExistsException:
			return NewFileAlreadyExistsExceptionWithReason(javaString(src), javaString(dst), failure.GetReason())
		case *FileSystemException:
			return NewFileSystemExceptionWithReason(javaString(src), javaString(dst), failure.GetReason())
		}
		return failure
	}
	return nil
}

// An ordinary Files.move may copy to another filesystem and then remove the
// source. A failed source removal removes the newly created destination too.
func fileMoveAcross(src, dst string, source os.FileInfo) error {
	srcPath, dstPath := ioUtilsTXTSystemPath(src), ioUtilsTXTSystemPath(dst)
	if source.IsDir() {
		file, err := os.Open(srcPath)
		if err != nil {
			return ioUtilsTXTPathError(src, err)
		}
		names, readErr := file.Readdirnames(1)
		closeErr := file.Close()
		if len(names) != 0 {
			return NewDirectoryNotEmptyException(javaString(src))
		}
		if readErr != nil && readErr != io.EOF {
			return ioUtilsTXTPathError(src, readErr)
		}
		if closeErr != nil {
			return ioUtilsTXTPathError(src, closeErr)
		}
		if err := os.Mkdir(dstPath, source.Mode().Perm()); err != nil {
			return ioUtilsTXTPathError(dst, err)
		}
	} else if source.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(srcPath)
		if err != nil {
			return ioUtilsTXTPathError(src, err)
		}
		if err := os.Symlink(target, dstPath); err != nil {
			return ioUtilsTXTPathError(dst, err)
		}
	} else if !source.Mode().IsRegular() {
		if err := fileMoveCopySpecial(dst, source); err != nil {
			return err
		}
	} else if err := copyFile(src, dst); err != nil {
		return err
	}
	if err := fileMoveCopyAttributes(src, dst, source); err != nil {
		_ = os.Remove(dstPath)
		return err
	}
	if err := os.Remove(srcPath); err != nil {
		_ = os.Remove(dstPath)
		if source.IsDir() && (errors.Is(err, syscall.ENOTEMPTY) || os.IsExist(err)) {
			return NewDirectoryNotEmptyException(javaString(src))
		}
		return ioUtilsTXTPathError(src, err)
	}
	return nil
}
