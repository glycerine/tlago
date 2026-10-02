package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
