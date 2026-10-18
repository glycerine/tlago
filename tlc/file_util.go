package tlc

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// filesCopyInputStream implements Files.copy(InputStream, Path,
// REPLACE_EXISTING). The caller retains ownership of input; Files owns only the
// output stream it creates after deleting the old target.
func filesCopyInputStream(input io.Reader, path string) (err error) {
	path, err = ioUtilsTXTPath(filenameNormalizeFile(path))
	if err != nil {
		return err
	}
	systemPath := ioUtilsTXTSystemPath(path)
	if _, statErr := os.Lstat(systemPath); statErr == nil {
		if removeErr := os.Remove(systemPath); removeErr != nil {
			return ioUtilsTXTPathError(path, removeErr)
		}
	} else if !os.IsNotExist(statErr) {
		return ioUtilsTXTPathError(path, statErr)
	}
	output, err := os.OpenFile(systemPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		return ioUtilsTXTPathError(path, err)
	}
	defer func() {
		failure := recover()
		closeErr := output.Close()
		if closeErr != nil {
			closeErr = ioUtilsTXTChannelError(closeErr)
		}
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			err = closeErr
		} else {
			javaSuppressCloseError(err, closeErr)
		}
	}()
	buffer := make([]byte, 16384)
	if _, err = io.CopyBuffer(output, input, buffer); err != nil {
		return ioUtilsTXTChannelError(err)
	}
	return nil
}

// openFileInputStream applies FileInputStream's constructor boundary before
// callers allocate or mutate their input streams. Go permits opening a directory
// for reading; Java rejects it during open, with FileNotFoundException.
func openFileInputStream(name string) (*os.File, error) {
	name = filenameNormalizeFile(name)
	if strings.IndexByte(name, 0) >= 0 {
		return nil, NewFileNotFoundException("Invalid file path")
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, bufferedRandomAccessFileOpenError(name, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, bufferedRandomAccessFileOpenError(name, err)
	}
	if info.IsDir() {
		_ = file.Close()
		return nil, bufferedRandomAccessFileOpenError(name, syscall.EISDIR)
	}
	return file, nil
}

// openFileOutputStream applies FileOutputStream(File)'s constructor boundary.
// It truncates an existing ordinary file and does not create parent directories.
func openFileOutputStream(name string) (*os.File, error) {
	return openFileOutputStreamMode(name, false)
}

// openFileOutputStreamAppend applies FileOutputStream(File, true)'s constructor
// boundary without changing the shared truncating constructor.
func openFileOutputStreamAppend(name string) (*os.File, error) {
	return openFileOutputStreamMode(name, true)
}

func openFileOutputStreamMode(name string, appendMode bool) (*os.File, error) {
	name = filenameNormalizeFile(name)
	if strings.IndexByte(name, 0) >= 0 {
		return nil, NewFileNotFoundException("Invalid file path")
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if appendMode {
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}
	file, err := os.OpenFile(name, flag, 0o666)
	if err != nil {
		return nil, bufferedRandomAccessFileOpenError(name, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, bufferedRandomAccessFileOpenError(name, err)
	}
	if info.IsDir() {
		_ = file.Close()
		return nil, bufferedRandomAccessFileOpenError(name, syscall.EISDIR)
	}
	return file, nil
}

// MakeMetaDir is FileUtil.makeMetaDir. A non-null checkpoint path is returned
// verbatim, including an empty string, without touching the filesystem.
func MakeMetaDir(date time.Time, specDir string, fromCheckpoint *string) string {
	if fromCheckpoint != nil {
		return *fromCheckpoint
	}
	path := tlcMetaDirPath(date, specDir, Globals.MetaDir)
	created, err := CreateExclusiveDirectoryWithApproximateName(path)
	if err != nil {
		panic(NewTLCRuntimeException(ECSystemMetadirCreationError, path))
	}
	return created
}

func tlcMetaDirPath(date time.Time, specDir, root string) string {
	if root == "" {
		root = filenamePathResolve(specDir, MetaRoot)
	}
	filenameCheckPath(root)
	return filenamePathResolve(root, date.In(time.Local).Format(tlcMetaDirDateLayout()))
}

// CreateExclusiveDirectory creates parents before atomically creating the
// requested directory, preserving the already-exists result for callers.
func CreateExclusiveDirectory(directory string) error {
	filenameCheckPath(directory)
	absolute, err := metadataAbsolutePath(directory)
	if err != nil {
		return err
	}
	if err = metadataCreateParents(metadataParentPath(absolute)); err != nil {
		return err
	}
	return os.Mkdir(absolute, 0o755)
}

// CreateExclusiveDirectoryWithApproximateName returns the original spelling
// on first creation. A collision returns an absolute temporary sibling, as
// Java does after converting the approximate Path toAbsolutePath.
func CreateExclusiveDirectoryWithApproximateName(approximatePath string) (string, error) {
	approximatePath = filenameNormalizeFile(approximatePath)
	if err := CreateExclusiveDirectory(approximatePath); err == nil {
		return approximatePath, nil
	} else if !os.IsExist(err) {
		return "", err
	}
	absolute, err := metadataAbsolutePath(approximatePath)
	if err != nil {
		return "", err
	}
	// Add the Go replacement marker at the end so a literal '*' in Java's
	// prefix remains literal rather than being consumed as the marker.
	return os.MkdirTemp(metadataParentPath(absolute), filepath.Base(absolute)+"*")
}

func metadataAbsolutePath(path string) (string, error) {
	path = filenameNormalizeFile(path)
	if filepath.IsAbs(path) {
		return path, nil
	}
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Path.toAbsolutePath does not normalize dot segments.
	return filenameFileJoin(directory, path), nil
}

func metadataParentPath(absolute string) string {
	index := strings.LastIndexByte(absolute, filepath.Separator)
	if index <= len(filepath.VolumeName(absolute)) {
		return absolute[:index+1]
	}
	return absolute[:index]
}

func metadataCreateParents(path string) error {
	// Files.createDirectories first tries the original Path. Its recursive
	// fallback relativizes the path, normalizing dot segments. Keep the final
	// exclusive create on the original path: missing components before ".."
	// can still make it fail even though the normalized parents were created.
	err := os.Mkdir(path, 0o755)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		if info, statErr := os.Stat(path); statErr == nil && info.IsDir() {
			return nil
		}
		return err
	}
	return os.MkdirAll(filepath.Clean(path), 0o755)
}

// filenameFileMkdirs follows java.io.File.mkdirs: try the original spelling,
// then create the canonical parent/path. Canonicalization resolves an existing
// ancestor's symlinks before reducing the unresolved dot components.
func filenameFileMkdirs(path string) bool {
	path = filenameNormalizeFile(path)
	if _, err := os.Stat(path); err == nil {
		return false
	}
	if err := os.Mkdir(path, 0o755); err == nil {
		return true
	}
	absolute, err := metadataAbsolutePath(path)
	if err != nil {
		return false
	}
	ancestor := absolute
	var suffix []string
	for {
		if canonical, err := filepath.EvalSymlinks(ancestor); err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				canonical = filepath.Join(canonical, suffix[i])
			}
			if err := os.MkdirAll(filepath.Dir(canonical), 0o755); err != nil {
				return false
			}
			return os.Mkdir(canonical, 0o755) == nil
		}
		parent := metadataParentPath(ancestor)
		if parent == ancestor || parent == "" {
			return false
		}
		suffix = append(suffix, distributedJavaFileName(ancestor))
		ancestor = parent
	}
}
